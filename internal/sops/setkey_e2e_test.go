// Package sops e2e test — exercises the real `sops` binary end-to-end to prove
// that SetKey CREATES a previously-absent key (the core assumption behind the
// "add new secret" feature) rather than only updating existing ones.
//
// This is the one place the suite drives the real subprocess: it is fully
// hermetic (generates its own age key in a temp dir, never touches the user's
// keyring) and skips cleanly when sops/age-keygen are not installed, so CI
// without those binaries stays green. All other executor tests avoid the real
// binary by design (command-structure / JSON-encoding / error-path coverage).
package sops

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSetKey_CreatesNewKey_E2E encrypts a file, adds a brand-new nested key via
// SetKey, then decrypts and asserts both the new key and the pre-existing key
// are present. Skips when sops or age-keygen are unavailable.
func TestSetKey_CreatesNewKey_E2E(t *testing.T) {
	sopsBin, err := exec.LookPath("sops")
	if err != nil {
		t.Skip("sops not installed; skipping end-to-end SetKey test")
	}
	if _, err := exec.LookPath("age-keygen"); err != nil {
		t.Skip("age-keygen not installed; skipping end-to-end SetKey test")
	}

	dir := t.TempDir()

	// Generate a hermetic age key inside the temp dir.
	keyPath := filepath.Join(dir, "keys.txt")
	keygen := exec.Command("age-keygen", "-o", keyPath)
	var keygenErr bytes.Buffer
	keygen.Stderr = &keygenErr
	require.NoError(t, keygen.Run(), "age-keygen failed: %s", keygenErr.String())

	// Derive the public recipient from the freshly generated key.
	pub := exec.Command("age-keygen", "-y", keyPath)
	pubOut, err := pub.Output()
	require.NoError(t, err, "age-keygen -y failed")
	recipient := strings.TrimSpace(string(pubOut))
	require.NotEmpty(t, recipient)

	// Point sops at the hermetic key so SetKey/DecryptFile can decrypt.
	t.Setenv("SOPS_AGE_KEY_FILE", keyPath)

	secretPath := filepath.Join(dir, "secret.yaml")
	require.NoError(t, os.WriteFile(secretPath, []byte("existing:\n  key: hello\n"), 0o600))

	// Encrypt in place. Recipients are passed via --age so no .sops.yaml lookup
	// is needed; the recipient list is then embedded in the file's sops metadata,
	// which is what SetKey/DecryptFile rely on afterwards (cwd-independent).
	enc := exec.Command(sopsBin, "encrypt", "-i", "--age", recipient, secretPath)
	var encErr bytes.Buffer
	enc.Stderr = &encErr
	require.NoError(t, enc.Run(), "sops encrypt failed: %s", encErr.String())

	// Add a brand-new nested key that does not exist yet.
	require.NoError(t, SetKey(context.Background(), secretPath, "database.password", "s3cret-value"),
		"SetKey must create a new key")

	// Decrypt and confirm both keys are present.
	plain, err := DecryptFile(context.Background(), secretPath)
	require.NoError(t, err, "decrypt after SetKey")
	out := string(plain)
	require.Contains(t, out, "s3cret-value", "newly added secret must be present")
	require.Contains(t, out, "hello", "pre-existing value must be preserved")

	// Sanity: a quoted/special value also round-trips through json.Marshal + --value-stdin.
	require.NoError(t, SetKey(context.Background(), secretPath, "quoted", `he said "hi"`))
	plain2, err := DecryptFile(context.Background(), secretPath)
	require.NoError(t, err)
	require.Contains(t, string(plain2), `he said "hi"`, "special characters must survive the round-trip")
}
