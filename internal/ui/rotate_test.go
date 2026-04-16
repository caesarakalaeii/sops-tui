package ui_test

import (
	"encoding/base64"
	"encoding/hex"
	"regexp"
	"strings"
	"testing"

	"github.com/caesarakalaeii/sops-tui/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// uuidv4Regex matches UUID v4 format with correct version nibble and variant bits.
var uuidv4Regex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// TestDetectFormatBase64 verifies that standard base64-encoded strings are detected as FormatBase64.
// The regex requires 22+ base64 chars (equivalent to at least 16 decoded bytes).
func TestDetectFormatBase64(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		// 24-char base64 = 18 decoded bytes, matches regex {22,}
		{"24-char base64", base64.StdEncoding.EncodeToString(make([]byte, 18))},
		{"longer base64", "SGVsbG8gV29ybGQgdGhpcyBpcyBhIGxvbmdlciB0ZXN0IHN0cmluZw=="},
		{"32-byte base64", base64.StdEncoding.EncodeToString(make([]byte, 32))},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ui.DetectFormat(tc.value)
			assert.Equal(t, ui.FormatBase64, got,
				"expected FormatBase64 for %q, got %v", tc.value, got)
		})
	}
}

// TestDetectFormatHex verifies that hex strings of 32+ chars are detected as FormatHex.
func TestDetectFormatHex(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"32 hex chars", "48656c6c6f20576f726c64212121212121"},
		{"64 hex chars", strings.Repeat("a1", 32)},
		{"uppercase hex", strings.ToUpper(strings.Repeat("ab", 16))},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ui.DetectFormat(tc.value)
			assert.Equal(t, ui.FormatHex, got,
				"expected FormatHex for %q, got %v", tc.value, got)
		})
	}
}

// TestDetectFormatUUID verifies that UUID v4 strings are detected as FormatUUID.
func TestDetectFormatUUID(t *testing.T) {
	value := "550e8400-e29b-41d4-a716-446655440000"
	got := ui.DetectFormat(value)
	assert.Equal(t, ui.FormatUUID, got,
		"expected FormatUUID for %q, got %v", value, got)
}

// TestDetectFormatBcrypt verifies that bcrypt hashes are detected as FormatBcrypt.
func TestDetectFormatBcrypt(t *testing.T) {
	value := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	got := ui.DetectFormat(value)
	assert.Equal(t, ui.FormatBcrypt, got,
		"expected FormatBcrypt for %q, got %v", value, got)
}

// TestDetectFormatUnknown verifies that regular strings return FormatUnknown.
func TestDetectFormatUnknown(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"plain text", "just a regular string"},
		{"short string", "abc"},
		{"empty", ""},
		{"url", "https://example.com/path"},
		{"email", "user@example.com"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ui.DetectFormat(tc.value)
			assert.Equal(t, ui.FormatUnknown, got,
				"expected FormatUnknown for %q, got %v", tc.value, got)
		})
	}
}

// TestDetectFormatBcryptBeforeBase64 verifies that a bcrypt hash (which could match the base64
// regex due to its character set) is correctly identified as FormatBcrypt, not FormatBase64.
// This tests D-13: detection order — bcrypt must be checked before base64.
func TestDetectFormatBcryptBeforeBase64(t *testing.T) {
	// This bcrypt hash contains characters valid in base64 — must be detected as bcrypt.
	bcryptHash := "$2a$12$R9h/cIPz0gi.URNNX3kh2OPST9/PgBkqquzi.Ss7KIUgO2t0jWMUW"
	got := ui.DetectFormat(bcryptHash)
	assert.Equal(t, ui.FormatBcrypt, got,
		"bcrypt hash must be detected as FormatBcrypt (not FormatBase64), got %v", got)
}

// TestGenerateValueBase64 verifies that GenerateValue(FormatBase64) returns a valid
// standard base64 string of approximately 44 characters (32 bytes encoded).
func TestGenerateValueBase64(t *testing.T) {
	val, err := ui.GenerateValue(ui.FormatBase64)
	require.NoError(t, err)
	require.NotEmpty(t, val)

	// Must be valid base64
	decoded, decodeErr := base64.StdEncoding.DecodeString(val)
	require.NoError(t, decodeErr, "GenerateValue(FormatBase64) must produce valid base64, got: %q", val)
	assert.Equal(t, 32, len(decoded), "decoded length must be 32 bytes")
	// Encoded: 32 bytes → ceil(32/3)*4 = 44 chars
	assert.Equal(t, 44, len(val), "base64 of 32 bytes must be 44 chars, got: %q", val)
}

// TestGenerateValueHex verifies that GenerateValue(FormatHex) returns a valid hex string of 64 chars.
func TestGenerateValueHex(t *testing.T) {
	val, err := ui.GenerateValue(ui.FormatHex)
	require.NoError(t, err)
	require.NotEmpty(t, val)

	// Must be valid hex
	decoded, decodeErr := hex.DecodeString(val)
	require.NoError(t, decodeErr, "GenerateValue(FormatHex) must produce valid hex, got: %q", val)
	assert.Equal(t, 32, len(decoded), "decoded length must be 32 bytes")
	assert.Equal(t, 64, len(val), "hex of 32 bytes must be 64 chars, got: %q", val)
}

// TestGenerateValueUUID verifies that GenerateValue(FormatUUID) returns a valid UUID v4.
func TestGenerateValueUUID(t *testing.T) {
	val, err := ui.GenerateValue(ui.FormatUUID)
	require.NoError(t, err)
	require.NotEmpty(t, val)

	assert.True(t, uuidv4Regex.MatchString(val),
		"GenerateValue(FormatUUID) must produce valid UUID v4, got: %q", val)
}

// TestGenerateValueBcrypt verifies that GenerateValue(FormatBcrypt) returns a bcrypt hash
// starting with "$2a$".
func TestGenerateValueBcrypt(t *testing.T) {
	val, err := ui.GenerateValue(ui.FormatBcrypt)
	require.NoError(t, err)
	require.NotEmpty(t, val)

	assert.True(t, strings.HasPrefix(val, "$2a$"),
		"GenerateValue(FormatBcrypt) must start with '$2a$', got: %q", val)
}

// TestGenerateValueAlphanumeric verifies that GenerateValue(FormatAlphanumeric) returns
// a 32-character alphanumeric string.
func TestGenerateValueAlphanumeric(t *testing.T) {
	val, err := ui.GenerateValue(ui.FormatAlphanumeric)
	require.NoError(t, err)
	require.NotEmpty(t, val)

	assert.Equal(t, 32, len(val), "alphanumeric must be 32 chars, got: %q", val)
	for _, r := range val {
		assert.True(t, (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'),
			"alphanumeric must contain only [a-zA-Z0-9], got char %q in %q", r, val)
	}
}

// TestGenerateValueUnknown verifies that GenerateValue(FormatUnknown) returns a non-empty fallback.
func TestGenerateValueUnknown(t *testing.T) {
	val, err := ui.GenerateValue(ui.FormatUnknown)
	require.NoError(t, err)
	assert.NotEmpty(t, val, "GenerateValue(FormatUnknown) must return non-empty fallback")
}

// TestFormatLabel verifies that each format returns a non-empty human-readable label.
func TestFormatLabel(t *testing.T) {
	formats := []ui.SecretFormat{
		ui.FormatBase64,
		ui.FormatHex,
		ui.FormatUUID,
		ui.FormatBcrypt,
		ui.FormatAlphanumeric,
		ui.FormatUnknown,
	}
	for _, f := range formats {
		label := ui.FormatLabel(f)
		assert.NotEmpty(t, label, "FormatLabel must return non-empty string for format %v", f)
	}
}

// TestAllFormats verifies that AllFormats returns exactly 5 formats.
func TestAllFormats(t *testing.T) {
	formats := ui.AllFormats()
	assert.Len(t, formats, 5, "AllFormats must return exactly 5 formats")
}

// TestDetectFormatGenerateRoundTrip verifies that a generated value for each format
// is correctly re-detected by DetectFormat (except bcrypt which can't round-trip easily).
func TestDetectFormatGenerateRoundTrip(t *testing.T) {
	tests := []struct {
		format   ui.SecretFormat
		expected ui.SecretFormat
	}{
		{ui.FormatBase64, ui.FormatBase64},
		{ui.FormatHex, ui.FormatHex},
		{ui.FormatUUID, ui.FormatUUID},
	}
	for _, tc := range tests {
		t.Run(ui.FormatLabel(tc.format), func(t *testing.T) {
			val, err := ui.GenerateValue(tc.format)
			require.NoError(t, err)
			detected := ui.DetectFormat(val)
			assert.Equal(t, tc.expected, detected,
				"generated %s value %q should re-detect as same format", ui.FormatLabel(tc.format), val)
		})
	}
}
