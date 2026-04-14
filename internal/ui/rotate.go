// Package ui provides format detection and random value generation for secret rotation.
//
// rotate.go implements DetectFormat, GenerateValue, AllFormats, and FormatLabel
// which together power the X key rotation flow (EDT-03).
//
// Detection order: bcrypt BEFORE base64 to prevent false base64 matches (D-13).
// All random generation uses crypto/rand (T-03-13: never math/rand).
// bcrypt uses cost 12 per UI-SPEC (T-03-14).
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// SecretFormat represents the detected or selected format of a secret value.
type SecretFormat int

const (
	// FormatUnknown means the format could not be auto-detected.
	// The format selection menu is shown when rotation is requested.
	FormatUnknown SecretFormat = iota
	// FormatBase64 means the secret is base64-encoded (32 bytes → 44 chars).
	FormatBase64
	// FormatHex means the secret is a hex string (32 bytes → 64 chars).
	FormatHex
	// FormatUUID means the secret is a UUID v4.
	FormatUUID
	// FormatBcrypt means the secret is a bcrypt hash ($2a$ prefix, cost 12).
	FormatBcrypt
	// FormatAlphanumeric means the secret is an alphanumeric string (32 chars).
	FormatAlphanumeric
)

// Detection regexes. Order of application in DetectFormat matters:
// bcrypt must be checked before base64 to avoid false matches (D-13).
var (
	reBcrypt = regexp.MustCompile(`^\$2[ayb]\$\d{2}\$`)
	reUUID   = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89aAbB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
	reHex    = regexp.MustCompile(`^[0-9a-fA-F]{32,}$`)
	reBase64 = regexp.MustCompile(`^[A-Za-z0-9+/]{22,}={0,2}$`)
)

// DetectFormat attempts to identify the format of a secret value.
// Detection order: bcrypt → UUID → hex → base64 → unknown.
// bcrypt is checked first because bcrypt hashes can match the base64 regex (D-13).
func DetectFormat(value string) SecretFormat {
	switch {
	case reBcrypt.MatchString(value):
		return FormatBcrypt
	case reUUID.MatchString(value):
		return FormatUUID
	case reHex.MatchString(value):
		return FormatHex
	case reBase64.MatchString(value):
		return FormatBase64
	default:
		return FormatUnknown
	}
}

// FormatLabel returns a human-readable name for each format.
// Used in the format selection menu and rotation success flash messages.
func FormatLabel(f SecretFormat) string {
	switch f {
	case FormatBase64:
		return "base64 (32 bytes)"
	case FormatHex:
		return "hex (32 bytes)"
	case FormatUUID:
		return "UUID v4"
	case FormatBcrypt:
		return "bcrypt (cost 12)"
	case FormatAlphanumeric:
		return "alphanumeric (32 chars)"
	default:
		return "unknown"
	}
}

// GenerateValue generates a cryptographically random value in the specified format.
// All generation uses crypto/rand (T-03-13). bcrypt uses cost 12 (T-03-14).
// FormatUnknown falls back to URL-safe base64 (32 bytes).
func GenerateValue(format SecretFormat) (string, error) {
	switch format {
	case FormatBase64:
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return "", fmt.Errorf("generate base64: %w", err)
		}
		return base64.StdEncoding.EncodeToString(b), nil

	case FormatHex:
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return "", fmt.Errorf("generate hex: %w", err)
		}
		return hex.EncodeToString(b), nil

	case FormatUUID:
		return generateUUIDv4()

	case FormatBcrypt:
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return "", fmt.Errorf("generate bcrypt input: %w", err)
		}
		hash, err := bcrypt.GenerateFromPassword(b, 12)
		if err != nil {
			return "", fmt.Errorf("generate bcrypt: %w", err)
		}
		return string(hash), nil

	case FormatAlphanumeric:
		return generateAlphanumeric(32)

	default:
		// Fallback: URL-safe base64 (32 bytes)
		b := make([]byte, 24)
		if _, err := rand.Read(b); err != nil {
			return "", fmt.Errorf("generate fallback: %w", err)
		}
		return base64.URLEncoding.EncodeToString(b), nil
	}
}

// AllFormats returns the ordered list of selectable formats for the format selection menu.
// The menu always shows these 5 options when format is ambiguous (FormatUnknown).
func AllFormats() []SecretFormat {
	return []SecretFormat{
		FormatBase64,
		FormatHex,
		FormatUUID,
		FormatBcrypt,
		FormatAlphanumeric,
	}
}

// generateUUIDv4 generates a random UUID version 4.
// Sets the version nibble (byte 6 high bits = 0100) and variant bits (byte 8 high bits = 10xx).
func generateUUIDv4() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate UUID: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits (RFC 4122)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// generateAlphanumeric generates a random alphanumeric string of the given length.
// Uses crypto/rand with modular sampling over the charset.
func generateAlphanumeric(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate alphanumeric: %w", err)
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}
