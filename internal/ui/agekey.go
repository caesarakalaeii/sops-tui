// Package ui - SOPS age identity parser (Phase 8).
//
// ParseAgeKeyFingerprint reads a SOPS age identity file and returns
// the FIRST identity's PUBLIC Recipient string (Bech32 "age1..." form).
// Used by AppModel at startup to populate the chrome info panel's
// age: row (D-201, D-214). On any failure (file missing, parse error,
// no identities, non-X25519 identity type) returns "" -- the chrome
// renderer shows "-" per D-204.
//
// SECURITY (D-220 question 1): we MUST type-assert to *age.X25519Identity
// before calling Recipient().String(). Calling Identity.String()
// directly on the private key would log "AGE-SECRET-KEY-..." -- a
// direct Pitfall 11 leak. The interface assertion is mandatory.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"os"
	"path/filepath"

	"filippo.io/age"
)

// AgeKeyFilePath returns the path to the user's SOPS age identities file.
// Honours $SOPS_AGE_KEY_FILE per SOPS CLI behaviour; falls back to
// ~/.config/sops/age/keys.txt (existing convention used by
// internal/validator/startup.go).
//
// D-214 + D-220 question 4: the path itself MUST NOT appear in any
// rendered chrome surface. This function is consumed only by
// AppModel startup wiring; the path string is used to open the file
// and is then discarded.
func AgeKeyFilePath() (string, error) {
	if env := os.Getenv("SOPS_AGE_KEY_FILE"); env != "" {
		return env, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "sops", "age", "keys.txt"), nil
}

// ParseAgeKeyFingerprint reads keyFilePath via age.ParseIdentities and
// returns the first identity's Recipient().String() (e.g. "age1abc...xyz").
// Returns "" on any failure (file missing, parse error, no identities,
// non-X25519 identity type) -- render layer shows "-" per D-204.
//
// Critical (08-RESEARCH.md Pitfall A): age.Identity is an interface
// and does NOT have a Recipient() method. The concrete *age.X25519Identity
// does. We MUST type-assert before calling Recipient().String();
// calling String() directly on the Identity would log the
// AGE-SECRET-KEY private key (Pitfall 11 leak; D-220 question 1).
func ParseAgeKeyFingerprint(keyFilePath string) string {
	f, err := os.Open(keyFilePath)
	if err != nil {
		return ""
	}
	defer f.Close()

	ids, err := age.ParseIdentities(f)
	if err != nil || len(ids) == 0 {
		return ""
	}
	if x25519, ok := ids[0].(*age.X25519Identity); ok {
		return x25519.Recipient().String() // Bech32-encoded PUBLIC key.
	}
	return "" // Plugin / hybrid identities -- future-proof default arm.
}
