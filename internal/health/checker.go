// Package health provides secret health check analysis for sops-tui.
//
// ShannonEntropy, IsWeakSecret, FindDuplicates, and IsStale are pure functions
// with no Bubble Tea or UI dependencies. They operate on decrypted string maps.
//
// Per D-08: weak = length < 16 OR entropy < 3.5 bits/char, with format-aware suffix checks.
// Per D-09: duplicate = exact decrypted value match across files (compared by SHA-256 hash).
// Per D-10: stale = days since last commit > threshold (default 90, env SOPS_TUI_STALE_DAYS).
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package health

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

// WeakSecret identifies a key whose value is considered weak.
type WeakSecret struct {
	FilePath string
	KeyPath  string
	Reason   string // "too short", "low entropy", "format mismatch"
}

// Duplicate identifies two or more keys that share the same decrypted value.
type Duplicate struct {
	ValueHash string     // SHA-256 hex of the value (never the plaintext)
	Locations []Location // all file+keypath pairs sharing this value
}

// Location is a file+keypath pair.
type Location struct {
	FilePath string
	KeyPath  string
}

// StaleFile identifies a file whose last commit is older than the threshold.
type StaleFile struct {
	FilePath       string
	LastCommitTime time.Time
	DaysSince      int
}

// HealthCheckResult is the complete output of a health scan.
type HealthCheckResult struct {
	WeakSecrets []WeakSecret
	Duplicates  []Duplicate
	StaleFiles  []StaleFile
	Errors      []string // files that failed to decrypt
}

// IsEmpty returns true when no issues were found and no errors occurred.
func (r HealthCheckResult) IsEmpty() bool {
	return len(r.WeakSecrets) == 0 && len(r.Duplicates) == 0 &&
		len(r.StaleFiles) == 0 && len(r.Errors) == 0
}

// HasErrLevelFindings returns true when the result contains any Err-severity
// finding: WeakSecrets, Duplicates, or Errors. StaleFiles are deliberately
// excluded — Phase 10 D-401 demotes staleness BELOW Warn (visible per-file but
// does not raise the logo to Err).
//
// This is the predicate the AppModel.resolveLogoState() classifier consults.
// It is distinct from IsEmpty() which DOES include StaleFiles in its zero-check.
func (r HealthCheckResult) HasErrLevelFindings() bool {
	return len(r.WeakSecrets) > 0 ||
		len(r.Duplicates) > 0 ||
		len(r.Errors) > 0
}

// ShannonEntropy computes the Shannon entropy of s in bits per character.
// Returns 0.0 for empty strings.
func ShannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	freq := make(map[rune]int)
	for _, r := range s {
		freq[r]++
	}
	n := float64(len([]rune(s)))
	var h float64
	for _, count := range freq {
		p := float64(count) / n
		h -= p * math.Log2(p)
	}
	return h
}

// weakKeyNameSuffixes lists key name suffixes that indicate structured secret formats
// (per D-08). When a key name ends with one of these and the value does not match
// any known format, it is flagged as a "format mismatch".
var weakKeyNameSuffixes = []string{"_token", "_key", "_secret", "_password", "_pwd"}

// Format-aware detection regexes (per D-08).
// These are duplicated from internal/ui/rotate.go because the health package
// cannot import the ui package (ui imports health, which would create a circular dependency).
var (
	reBase64 = regexp.MustCompile(`^[A-Za-z0-9+/]{22,}={0,2}$`)
	reHex    = regexp.MustCompile(`^[0-9a-fA-F]{32,}$`)
	reUUID   = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89aAbB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
	reBcrypt = regexp.MustCompile(`^\$2[ayb]\$\d{2}\$`)
)

// hasKnownFormat returns true if the value matches any recognized secret format
// (base64, hex, UUID, or bcrypt). Used by IsWeakSecret for format-aware detection
// per D-08: when a key name ends in a weak suffix (_token, _key, _secret, etc.)
// and the value does not match any known format, it is flagged as weak.
//
// These regexes are duplicated from internal/ui/rotate.go because the health
// package cannot import the ui package (ui imports health, which would create
// a circular dependency).
func hasKnownFormat(value string) bool {
	return reBcrypt.MatchString(value) ||
		reUUID.MatchString(value) ||
		(reHex.MatchString(value) && len(value)%2 == 0) ||
		reBase64.MatchString(value) ||
		isValidBase64(value)
}

// isValidBase64 checks if the value successfully decodes as base64
// and produces at least 16 bytes of output (non-trivial payload).
func isValidBase64(value string) bool {
	decoded, err := base64.StdEncoding.DecodeString(value)
	return err == nil && len(decoded) >= 16
}

// IsWeakSecret returns true and a reason string if the value at keyPath is considered weak.
// Checks: length < 16, entropy < 3.5, and format mismatch for known-sensitive key name suffixes.
//
// Per D-08: for key names with sensitive suffixes (_token, _key, _secret, etc.), if the value
// matches a known format (base64, hex, UUID, bcrypt), it is accepted regardless of entropy —
// structured formats like UUIDs may have low character diversity by design. If the suffix matches
// but the format does not, it is flagged as "format mismatch".
func IsWeakSecret(keyPath, value string) (bool, string) {
	if len(value) < 16 {
		return true, "too short"
	}
	// Format-aware check per D-08: check known-suffix keys before entropy check.
	// A valid UUID/base64/hex/bcrypt value with a sensitive suffix is NOT weak even if
	// its entropy is low (e.g. a UUID with many zeros is still a valid UUID).
	lower := strings.ToLower(keyPath)
	for _, suffix := range weakKeyNameSuffixes {
		if strings.HasSuffix(lower, suffix) {
			if hasKnownFormat(value) {
				// Known-format value with a sensitive suffix is accepted.
				return false, ""
			}
			// The key name suggests a structured secret, but value doesn't match any known format.
			return true, "format mismatch"
		}
	}
	// For keys without sensitive suffixes, fall back to entropy check.
	if ShannonEntropy(value) < 3.5 {
		return true, "low entropy"
	}
	return false, ""
}

// FindDuplicates scans a map of file→(keyPath→plaintextValue) for exact duplicates.
// Returns Duplicate entries identified by value SHA-256 hash, never the plaintext.
// The input map is consumed and the plaintext strings are not retained after hashing.
func FindDuplicates(fileValues map[string]map[string]string) []Duplicate {
	// map[valueHash][]Location
	seen := make(map[string][]Location)
	for filePath, kvs := range fileValues {
		for keyPath, value := range kvs {
			h := sha256.Sum256([]byte(value))
			hash := fmt.Sprintf("%x", h)[:16] // truncated to 16 hex chars (8 bytes) for display
			seen[hash] = append(seen[hash], Location{FilePath: filePath, KeyPath: keyPath})
		}
	}
	var dups []Duplicate
	for hash, locs := range seen {
		if len(locs) > 1 {
			dups = append(dups, Duplicate{ValueHash: hash, Locations: locs})
		}
	}
	return dups
}
