// Package sops provides SOPS file discovery services for sops-tui.
package sops

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const encryptedFileFixture = `database:
  host: ENC[AES256_GCM,data:abc,iv:def,tag:ghi,type:str]
  port: ENC[AES256_GCM,data:xyz,iv:uvw,tag:rst,type:int]
sops:
  version: "3.12.2"
  lastmodified: "2024-01-15T10:30:00Z"
  mac: "ENC[AES256_GCM,data:mac,iv:iv,tag:tag,type:str]"
`

const unencryptedFileFixture = `database:
  host: localhost
  port: 5432
`

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	err := os.MkdirAll(filepath.Dir(path), 0755)
	require.NoError(t, err)
	err = os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return path
}

// TestDiscover_EncryptedFile verifies that a file matching path_regex with a sops: key
// is returned as DiscoveredFile{IsEncrypted: true}.
func TestDiscover_EncryptedFile(t *testing.T) {
	dir := t.TempDir()

	sopsYaml := `creation_rules:
  - path_regex: "secrets/.*\\.yaml$"
    age: age1xxxxxxxxxx
`
	writeTempFile(t, dir, ".sops.yaml", sopsYaml)
	writeTempFile(t, dir, "secrets/prod.yaml", encryptedFileFixture)

	files, err := Discover(filepath.Join(dir, ".sops.yaml"))
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "secrets/prod.yaml", files[0].Name)
	assert.True(t, files[0].IsEncrypted)
}

// TestDiscover_UnencryptedFile verifies that a file matching path_regex without a sops: key
// is returned as DiscoveredFile{IsEncrypted: false}.
func TestDiscover_UnencryptedFile(t *testing.T) {
	dir := t.TempDir()

	sopsYaml := `creation_rules:
  - path_regex: "secrets/.*\\.yaml$"
    age: age1xxxxxxxxxx
`
	writeTempFile(t, dir, ".sops.yaml", sopsYaml)
	writeTempFile(t, dir, "secrets/dev.yaml", unencryptedFileFixture)

	files, err := Discover(filepath.Join(dir, ".sops.yaml"))
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "secrets/dev.yaml", files[0].Name)
	assert.False(t, files[0].IsEncrypted)
}

// TestDiscover_CatchAllRule verifies that an empty path_regex matches all YAML files.
func TestDiscover_CatchAllRule(t *testing.T) {
	dir := t.TempDir()

	sopsYaml := `creation_rules:
  - age: age1xxxxxxxxxx
`
	writeTempFile(t, dir, ".sops.yaml", sopsYaml)
	writeTempFile(t, dir, "a.yaml", unencryptedFileFixture)
	writeTempFile(t, dir, "subdir/b.yaml", unencryptedFileFixture)

	files, err := Discover(filepath.Join(dir, ".sops.yaml"))
	require.NoError(t, err)
	assert.Len(t, files, 2)
}

// TestDiscover_FirstMatchWins verifies that first-match-wins rule ordering is applied.
func TestDiscover_FirstMatchWins(t *testing.T) {
	dir := t.TempDir()

	sopsYaml := `creation_rules:
  - path_regex: "secrets/.*\\.yaml$"
    encrypted_regex: "^(host|password)$"
    age: age1first
  - path_regex: "secrets/.*\\.yaml$"
    encrypted_regex: "^(all)$"
    age: age1second
`
	writeTempFile(t, dir, ".sops.yaml", sopsYaml)
	writeTempFile(t, dir, "secrets/prod.yaml", encryptedFileFixture)

	files, err := Discover(filepath.Join(dir, ".sops.yaml"))
	require.NoError(t, err)
	require.Len(t, files, 1)
	// Should use the FIRST matching rule
	assert.Equal(t, "^(host|password)$", files[0].Rule.EncryptedRegex)
	assert.Equal(t, "age1first", files[0].Rule.Age)
}

// TestDiscover_CreationRuleFields verifies that EncryptedRegex and UnencryptedRegex
// fields are populated from the .sops.yaml.
func TestDiscover_CreationRuleFields(t *testing.T) {
	dir := t.TempDir()

	sopsYaml := `creation_rules:
  - path_regex: "secrets/.*\\.yaml$"
    encrypted_regex: "^(password|secret)$"
    unencrypted_regex: "^(host|port)$"
    age: age1xxxxxxxxxx
`
	writeTempFile(t, dir, ".sops.yaml", sopsYaml)
	writeTempFile(t, dir, "secrets/prod.yaml", encryptedFileFixture)

	files, err := Discover(filepath.Join(dir, ".sops.yaml"))
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "^(password|secret)$", files[0].Rule.EncryptedRegex)
	assert.Equal(t, "^(host|port)$", files[0].Rule.UnencryptedRegex)
}

// TestHasSOPSMarker_True verifies hasSOPSMarker returns true for files with a sops: key.
func TestHasSOPSMarker_True(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "encrypted.yaml", encryptedFileFixture)
	assert.True(t, hasSOPSMarker(path))
}

// TestHasSOPSMarker_False verifies hasSOPSMarker returns false for files without a sops: key.
func TestHasSOPSMarker_False(t *testing.T) {
	dir := t.TempDir()
	path := writeTempFile(t, dir, "plain.yaml", unencryptedFileFixture)
	assert.False(t, hasSOPSMarker(path))
}

// TestMatchRule_RelativePath verifies that matchRule computes file path relative to
// .sops.yaml directory before regex matching (Pitfall 3).
func TestMatchRule_RelativePath(t *testing.T) {
	dir := t.TempDir()

	rules := []CreationRule{
		{PathRegex: "secrets/.*\\.yaml$"},
	}

	// absolute path to a file inside secrets/
	absFilePath := filepath.Join(dir, "secrets", "prod.yaml")
	rule, matched := matchRule(absFilePath, dir, rules)
	assert.True(t, matched)
	assert.Equal(t, "secrets/.*\\.yaml$", rule.PathRegex)

	// file outside secrets/ should NOT match
	absOtherPath := filepath.Join(dir, "other", "prod.yaml")
	_, matched2 := matchRule(absOtherPath, dir, rules)
	assert.False(t, matched2)
}

// TestDiscover_InvalidRegex verifies that an invalid path_regex does not cause a panic
// and that the rule is simply skipped.
func TestDiscover_InvalidRegex(t *testing.T) {
	dir := t.TempDir()

	sopsYaml := `creation_rules:
  - path_regex: "[invalid"
    age: age1bad
  - path_regex: "secrets/.*\\.yaml$"
    age: age1good
`
	writeTempFile(t, dir, ".sops.yaml", sopsYaml)
	writeTempFile(t, dir, "secrets/prod.yaml", unencryptedFileFixture)

	// Should not panic; invalid rule is skipped; valid rule still matches
	files, err := Discover(filepath.Join(dir, ".sops.yaml"))
	require.NoError(t, err)
	assert.Len(t, files, 1)
}

// TestDiscover_NoTraversal verifies that Discover does not traverse outside the
// .sops.yaml parent directory.
func TestDiscover_NoTraversal(t *testing.T) {
	dir := t.TempDir()

	// A catch-all rule
	sopsYaml := `creation_rules:
  - age: age1xxxxxxxxxx
`
	writeTempFile(t, dir, ".sops.yaml", sopsYaml)
	writeTempFile(t, dir, "safe.yaml", unencryptedFileFixture)

	files, err := Discover(filepath.Join(dir, ".sops.yaml"))
	require.NoError(t, err)

	// All returned files must have AbsPath starting with dir
	for _, f := range files {
		assert.True(t, filepath.HasPrefix(f.AbsPath, dir),
			"file %q should be under %q", f.AbsPath, dir)
	}
}
