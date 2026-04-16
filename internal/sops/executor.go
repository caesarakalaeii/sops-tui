// Package sops provides SOPS subprocess execution services for sops-tui.
//
// This file implements the executor: thin wrappers around `sops` CLI calls for
// decrypt/encrypt operations. All functions use exec.CommandContext with a
// caller-provided context so callers can apply timeouts (use SopsTimeout).
//
// Security:
//   - T-03-01: SetKey uses --value-stdin exclusively; the secret never appears in process listings.
//   - T-03-04: All subprocess calls use exec.CommandContext — callers apply context.WithTimeout.
//   - T-03-03: Key paths come from the YAML parser (already sanitized); sops CLI handles its own validation.
//
// Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package sops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// SopsTimeout is the default maximum duration for any sops subprocess call.
// Callers should apply context.WithTimeout(ctx, SopsTimeout) before passing context.
const SopsTimeout = 30 * time.Second

// arrayIndexRegex matches key paths containing array-index notation like [0] or [42].
var arrayIndexRegex = regexp.MustCompile(`\[\d+\]`)

// dotPathToIndex converts a dot-separated key path to SOPS bracket-index notation.
// For example: "database.password" → `["database"]["password"]`
// Single keys are wrapped in a single bracket pair: "simple" → `["simple"]`
// This is the format required by `sops decrypt --extract` and `sops set`.
func dotPathToIndex(dotPath string) string {
	parts := strings.Split(dotPath, ".")
	var sb strings.Builder
	for _, part := range parts {
		sb.WriteString(`["`)
		sb.WriteString(part)
		sb.WriteString(`"]`)
	}
	return sb.String()
}

// IsArrayIndexedKeyPath reports whether keyPath contains array-index notation
// (e.g., "items[0].name" or "list[42]"). Such paths are not supported by
// `sops set` and should be blocked in the UI (Open Question 1 resolution).
func IsArrayIndexedKeyPath(keyPath string) bool {
	return arrayIndexRegex.MatchString(keyPath)
}

// DecryptKey decrypts and returns the plaintext value of a single key from a
// SOPS-encrypted file. keyPath is a dot-separated path (e.g., "database.password").
//
// Runs: sops decrypt --extract '["database"]["password"]' /path/to/file
//
// The returned string has trailing newlines stripped (Pitfall 5 mitigation).
// ctx should have a timeout applied: context.WithTimeout(ctx, SopsTimeout).
func DecryptKey(ctx context.Context, filePath, keyPath string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	index := dotPathToIndex(keyPath)
	cmd := exec.CommandContext(ctx, "sops", "decrypt", "--extract", index, filePath)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("sops decrypt --extract: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	// Pitfall 5: sops always appends a trailing newline to extracted values.
	return strings.TrimRight(stdout.String(), "\n"), nil
}

// DecryptFile decrypts an entire SOPS-encrypted file and returns the raw YAML bytes.
//
// Runs: sops decrypt /path/to/file
//
// ctx should have a timeout applied: context.WithTimeout(ctx, SopsTimeout).
func DecryptFile(ctx context.Context, filePath string) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	cmd := exec.CommandContext(ctx, "sops", "decrypt", filePath)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("sops decrypt: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return stdout.Bytes(), nil
}

// SetKey sets a single key in a SOPS-encrypted file to a new value, atomically
// re-encrypting only that key.
//
// Runs: sops set --value-stdin /path/to/file '["key"]["path"]'
// The new value is JSON-encoded and passed via stdin (Pitfall 1: json.Marshal wrapping;
// T-03-01: --value-stdin prevents secret from appearing in process listings).
//
// NOTE: When the file uses encrypted_regex in .sops.yaml, sops set targets only the
// specified key path and does not re-evaluate the regex filter on other fields.
// This is accepted as the correct behavior per Open Question 2 (RESOLVED). If edge
// cases surface, address as bugs.
//
// ctx should have a timeout applied: context.WithTimeout(ctx, SopsTimeout).
func SetKey(ctx context.Context, filePath, keyPath, newValue string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	index := dotPathToIndex(keyPath)

	// Pitfall 1: JSON-encode the value so sops receives a properly quoted JSON string.
	// json.Marshal("hello") → `"hello"`, json.Marshal(`say "hi"`) → `"say \"hi\""`
	jsonVal, err := json.Marshal(newValue)
	if err != nil {
		return fmt.Errorf("sops set: json encoding failed: %w", err)
	}

	cmd := exec.CommandContext(ctx, "sops", "set", "--value-stdin", filePath, index)
	cmd.Stdin = bytes.NewReader(jsonVal)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sops set: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return nil
}

// SopsRotateTimeout is the maximum duration for sops rotate operations.
// Longer than SopsTimeout because rotate decrypts all key-encryption records
// then re-encrypts all of them — the operation scales with the number of recipients
// and the size of the file (RESEARCH.md Open Question 1).
const SopsRotateTimeout = 60 * time.Second

// AddRecipient adds an age public key as a recipient to a SOPS-encrypted file,
// rotating the key-encryption records so the new recipient can decrypt.
//
// Runs: sops rotate -i --add-age <pubkey> <filePath>
//
// ctx should have a timeout applied: context.WithTimeout(ctx, SopsRotateTimeout).
// Age public key validation is delegated to the caller (T-05-01: Plan 02 validates
// with age.ParseX25519Recipient before calling AddRecipient).
func AddRecipient(ctx context.Context, filePath, pubkey string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	cmd := exec.CommandContext(ctx, "sops", "rotate", "-i", "--add-age", pubkey, filePath)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sops rotate --add-age: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return nil
}

// RemoveRecipient removes an age public key from a SOPS-encrypted file's recipient list,
// rotating the key-encryption records so the removed recipient can no longer decrypt.
//
// Runs: sops rotate -i --rm-age <pubkey> <filePath>
//
// ctx should have a timeout applied: context.WithTimeout(ctx, SopsRotateTimeout).
func RemoveRecipient(ctx context.Context, filePath, pubkey string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	cmd := exec.CommandContext(ctx, "sops", "rotate", "-i", "--rm-age", pubkey, filePath)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sops rotate --rm-age: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	return nil
}

// EncryptFile encrypts srcPath using sops and writes the ciphertext to destPath.
// This is used in the $EDITOR flow (Plan 03) where a temp file is edited in plaintext
// then re-encrypted to replace the original.
//
// Runs: sops encrypt srcPath → writes stdout atomically to destPath
//
// The write is atomic: sops output is piped directly into a sibling temp file in
// filepath.Dir(destPath), then renamed over destPath. This prevents destPath from
// being left truncated or partially written if the process is interrupted mid-write
// (disk full, power loss, SIGKILL). os.Rename on POSIX is atomic when src and dst
// are on the same filesystem, which is guaranteed by the sibling temp placement.
//
// ctx should have a timeout applied: context.WithTimeout(ctx, SopsTimeout).
func EncryptFile(ctx context.Context, srcPath, destPath string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// Create a sibling temp file so os.Rename is guaranteed to be on the same device.
	dir := filepath.Dir(destPath)
	tmp, err := os.CreateTemp(dir, ".sops-tui-enc-*.tmp")
	if err != nil {
		return fmt.Errorf("sops encrypt: create temp: %w", err)
	}
	tmpPath := tmp.Name()

	// Set restrictive permissions before any data is written.
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("sops encrypt: chmod temp: %w", err)
	}

	cmd := exec.CommandContext(ctx, "sops", "encrypt", srcPath)
	cmd.Stdout = tmp // pipe directly; no in-memory buffer needed

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("sops encrypt: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	// Flush OS buffers to disk before the rename so the destination is never stale.
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("sops encrypt: sync temp: %w", err)
	}
	tmp.Close()

	// Atomic rename: on POSIX this replaces destPath in a single syscall.
	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("sops encrypt: rename: %w", err)
	}

	return nil
}
