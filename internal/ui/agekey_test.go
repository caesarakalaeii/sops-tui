package ui_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"filippo.io/age"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

func TestParseAgeKeyFingerprint_FirstIdentity(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	expectedRecipient := identity.Recipient().String()
	require.True(t, strings.HasPrefix(expectedRecipient, "age1"))

	dir := t.TempDir()
	keyPath := filepath.Join(dir, "keys.txt")
	// Identity.String() returns the canonical AGE-SECRET-KEY-... form
	// that age.ParseIdentities expects. We use it ONLY for fixture
	// generation in tests -- production code (ParseAgeKeyFingerprint)
	// never calls Identity.String() per D-220 question 1.
	contents := "# created: 2026-01-01T00:00:00Z\n" +
		"# public key: " + expectedRecipient + "\n" +
		identity.String() + "\n"
	require.NoError(t, os.WriteFile(keyPath, []byte(contents), 0o600))

	got := ui.ParseAgeKeyFingerprint(keyPath)
	assert.Equal(t, expectedRecipient, got,
		"ParseAgeKeyFingerprint must return Recipient().String() (the public bech32 'age1' key), not the private AGE-SECRET-KEY")
	assert.True(t, strings.HasPrefix(got, "age1"),
		"fingerprint must be the bech32-encoded public key, got: %q", got)
	assert.NotContains(t, got, "AGE-SECRET-KEY",
		"fingerprint MUST NOT leak the private key (D-220 question 1)")
}

func TestParseAgeKeyFingerprint_MissingFile(t *testing.T) {
	got := ui.ParseAgeKeyFingerprint(filepath.Join(t.TempDir(), "does-not-exist.txt"))
	assert.Equal(t, "", got, "missing file must return empty string (renders as '-' per D-204)")
}

func TestParseAgeKeyFingerprint_MalformedFile(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "keys.txt")
	require.NoError(t, os.WriteFile(keyPath, []byte("not a valid age key file"), 0o600))

	got := ui.ParseAgeKeyFingerprint(keyPath)
	assert.Equal(t, "", got, "malformed file must return empty (renders as '-' per D-204)")
}

func TestAgeKeyFilePath_EnvOverride(t *testing.T) {
	t.Setenv("SOPS_AGE_KEY_FILE", "/custom/path/keys.txt")
	path, err := ui.AgeKeyFilePath()
	require.NoError(t, err)
	assert.Equal(t, "/custom/path/keys.txt", path,
		"AgeKeyFilePath must honour $SOPS_AGE_KEY_FILE per D-214")
}

func TestAgeKeyFilePath_HomeDefault(t *testing.T) {
	t.Setenv("SOPS_AGE_KEY_FILE", "")
	path, err := ui.AgeKeyFilePath()
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(path, filepath.Join(".config", "sops", "age", "keys.txt")),
		"default path must end with .config/sops/age/keys.txt; got %q", path)
}
