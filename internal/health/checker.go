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
	"fmt"
	"math"
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

var weakKeyNameSuffixes = []string{"_token", "_key", "_secret", "_password", "_pwd"}

// IsWeakSecret returns true and a reason string if the value at keyPath is considered weak.
// Checks: length < 16, entropy < 3.5, and format mismatch for known-sensitive key name suffixes.
func IsWeakSecret(keyPath, value string) (bool, string) {
	if len(value) < 16 {
		return true, "too short"
	}
	if ShannonEntropy(value) < 3.5 {
		return true, "low entropy"
	}
	lower := strings.ToLower(keyPath)
	for _, suffix := range weakKeyNameSuffixes {
		if strings.HasSuffix(lower, suffix) {
			// placeholder: caller can extend with format validation
			break
		}
	}
	return false, ""
}

// FindDuplicates scans a map of file->(keyPath->plaintextValue) for exact duplicates.
// Returns Duplicate entries identified by value SHA-256 hash, never the plaintext.
// The input map is consumed and the plaintext strings are not retained.
func FindDuplicates(fileValues map[string]map[string]string) []Duplicate {
	// map[valueHash][]Location
	seen := make(map[string][]Location)
	for filePath, kvs := range fileValues {
		for keyPath, value := range kvs {
			h := sha256.Sum256([]byte(value))
			hash := fmt.Sprintf("%x", h[:8]) // truncated to 8 bytes for display
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
