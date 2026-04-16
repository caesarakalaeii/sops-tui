// Package parser provides tests for the SOPS encrypted YAML tree parser.
package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caesarakalaeii/sops-tui/internal/sops"
)

// encryptedYAML is a well-formed SOPS-encrypted YAML fixture with:
// - nested map "database" with encrypted fields
// - sops: metadata block to hide from tree
// - encrypted_regex matching host/port/enabled keys
const encryptedYAML = `database:
  host: ENC[AES256_GCM,data:abc,iv:def,tag:ghi,type:str]
  port: ENC[AES256_GCM,data:xyz,iv:uvw,tag:rst,type:int]
  enabled: ENC[AES256_GCM,data:mno,iv:pqr,tag:stu,type:bool]
sops:
  version: "3.12.2"
  lastmodified: "2024-01-15T10:30:00Z"
  mac: "ENC[AES256_GCM,data:mac,iv:iv,tag:tag,type:str]"
  age:
    - recipient: age1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
  encrypted_regex: "^(host|port|enabled)$"
`

// partialEncryptedYAML has an encrypted_regex that leaves some keys unencrypted
const partialEncryptedYAML = `app:
  name: my-app
  password: ENC[AES256_GCM,data:abc,iv:def,tag:ghi,type:str]
sops:
  version: "3.12.2"
  lastmodified: "2024-01-15T10:30:00Z"
  mac: "ENC[AES256_GCM,data:mac,iv:iv,tag:tag,type:str]"
  encrypted_regex: "^(password)$"
`

// unencryptedYAML has no sops: block
const unencryptedYAML = `database:
  host: localhost
  port: 5432
`

// mixedTypesYAML has non-string YAML values (int, bool) in the top level
const mixedTypesYAML = `count: 42
enabled: true
name: my-service
sops:
  version: "3.12.2"
  lastmodified: "2024-01-15T10:30:00Z"
  mac: "ENC[AES256_GCM,data:mac,iv:iv,tag:tag,type:str]"
`

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return path
}

// TestParseFile_KeyOrder verifies that ParseFile returns TreeNodes with correct key
// names preserving YAML key order.
func TestParseFile_KeyOrder(t *testing.T) {
	path := writeTempYAML(t, encryptedYAML)
	rule := sops.CreationRule{EncryptedRegex: "^(host|port|enabled)$"}

	result, err := ParseFile(path, rule, true)
	require.NoError(t, err)

	// Should have 1 top-level node: database (sops is hidden)
	require.Len(t, result.Nodes, 1)
	assert.Equal(t, "database", result.Nodes[0].Key)
	assert.Len(t, result.Nodes[0].Children, 3)
	// Keys must be in order: host, port, enabled
	assert.Equal(t, "host", result.Nodes[0].Children[0].Key)
	assert.Equal(t, "port", result.Nodes[0].Children[1].Key)
	assert.Equal(t, "enabled", result.Nodes[0].Children[2].Key)
}

// TestParseFile_HidesSopsKey verifies that sops: top-level key is excluded from tree.
func TestParseFile_HidesSopsKey(t *testing.T) {
	path := writeTempYAML(t, encryptedYAML)
	rule := sops.CreationRule{EncryptedRegex: "^(host|port|enabled)$"}

	result, err := ParseFile(path, rule, true)
	require.NoError(t, err)

	// No node should have Key == "sops"
	for _, node := range result.Nodes {
		assert.NotEqual(t, "sops", node.Key, "sops key must be hidden from tree")
	}
}

// TestParseFile_MetadataVersion verifies SopsMetadata.Version extracted from sops: block.
func TestParseFile_MetadataVersion(t *testing.T) {
	path := writeTempYAML(t, encryptedYAML)
	rule := sops.CreationRule{EncryptedRegex: "^(host|port|enabled)$"}

	result, err := ParseFile(path, rule, true)
	require.NoError(t, err)
	assert.Equal(t, "3.12.2", result.Metadata.Version)
}

// TestParseFile_MetadataLastModified verifies SopsMetadata.LastModified extracted.
func TestParseFile_MetadataLastModified(t *testing.T) {
	path := writeTempYAML(t, encryptedYAML)
	rule := sops.CreationRule{EncryptedRegex: "^(host|port|enabled)$"}

	result, err := ParseFile(path, rule, true)
	require.NoError(t, err)
	assert.Equal(t, "2024-01-15T10:30:00Z", result.Metadata.LastModified)
}

// TestParseFile_MetadataAgeRecipients verifies SopsMetadata.AgeRecipients extracted.
func TestParseFile_MetadataAgeRecipients(t *testing.T) {
	path := writeTempYAML(t, encryptedYAML)
	rule := sops.CreationRule{EncryptedRegex: "^(host|port|enabled)$"}

	result, err := ParseFile(path, rule, true)
	require.NoError(t, err)
	require.Len(t, result.Metadata.AgeRecipients, 1)
	assert.Equal(t, "age1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx", result.Metadata.AgeRecipients[0])
}

// TestParseFile_MetadataEncryptedRegex verifies SopsMetadata.EncryptedRegex extracted.
func TestParseFile_MetadataEncryptedRegex(t *testing.T) {
	path := writeTempYAML(t, encryptedYAML)
	rule := sops.CreationRule{EncryptedRegex: "^(host|port|enabled)$"}

	result, err := ParseFile(path, rule, true)
	require.NoError(t, err)
	assert.Equal(t, "^(host|port|enabled)$", result.Metadata.EncryptedRegex)
}

// TestParseFile_EncryptedLeaf_Str verifies ENC[...,type:str] produces Encrypted=true, TypeHint="str".
func TestParseFile_EncryptedLeaf_Str(t *testing.T) {
	path := writeTempYAML(t, encryptedYAML)
	rule := sops.CreationRule{EncryptedRegex: "^(host|port|enabled)$"}

	result, err := ParseFile(path, rule, true)
	require.NoError(t, err)

	hostNode := result.Nodes[0].Children[0] // database.host
	assert.True(t, hostNode.Encrypted)
	assert.Equal(t, "str", hostNode.TypeHint)
}

// TestParseFile_EncryptedLeaf_Int verifies ENC[...,type:int] produces TypeHint="int".
func TestParseFile_EncryptedLeaf_Int(t *testing.T) {
	path := writeTempYAML(t, encryptedYAML)
	rule := sops.CreationRule{EncryptedRegex: "^(host|port|enabled)$"}

	result, err := ParseFile(path, rule, true)
	require.NoError(t, err)

	portNode := result.Nodes[0].Children[1] // database.port
	assert.True(t, portNode.Encrypted)
	assert.Equal(t, "int", portNode.TypeHint)
}

// TestParseFile_EncryptedLeaf_Bool verifies ENC[...,type:bool] produces TypeHint="bool".
func TestParseFile_EncryptedLeaf_Bool(t *testing.T) {
	path := writeTempYAML(t, encryptedYAML)
	rule := sops.CreationRule{EncryptedRegex: "^(host|port|enabled)$"}

	result, err := ParseFile(path, rule, true)
	require.NoError(t, err)

	enabledNode := result.Nodes[0].Children[2] // database.enabled
	assert.True(t, enabledNode.Encrypted)
	assert.Equal(t, "bool", enabledNode.TypeHint)
}

// TestParseFile_PlainValue verifies that a non-ENC string in a file with encrypted_regex
// gets IsPlain=true when key doesn't match the encrypted_regex.
func TestParseFile_PlainValue(t *testing.T) {
	path := writeTempYAML(t, partialEncryptedYAML)
	rule := sops.CreationRule{EncryptedRegex: "^(password)$"}

	result, err := ParseFile(path, rule, true)
	require.NoError(t, err)

	// Find the "app" node
	require.Len(t, result.Nodes, 1)
	appNode := result.Nodes[0]
	require.Len(t, appNode.Children, 2)

	// "name" is not in encrypted_regex so it's plain
	nameNode := appNode.Children[0] // app.name
	assert.Equal(t, "name", nameNode.Key)
	assert.True(t, nameNode.IsPlain)
	assert.Equal(t, "my-app", nameNode.Value)
}

// TestParseFile_NestedNodes verifies that nested YAML maps produce TreeNode with Children.
func TestParseFile_NestedNodes(t *testing.T) {
	path := writeTempYAML(t, encryptedYAML)
	rule := sops.CreationRule{EncryptedRegex: "^(host|port|enabled)$"}

	result, err := ParseFile(path, rule, true)
	require.NoError(t, err)

	// database is a group node (has children, no value)
	dbNode := result.Nodes[0]
	assert.Equal(t, "database", dbNode.Key)
	assert.Equal(t, 0, dbNode.Depth)
	assert.Len(t, dbNode.Children, 3)

	// Children should have depth=1
	for _, child := range dbNode.Children {
		assert.Equal(t, 1, child.Depth)
	}
}

// TestExtractTypeHint_Str verifies extractTypeHint returns "str" from type:str suffix.
func TestExtractTypeHint_Str(t *testing.T) {
	result := extractTypeHint("ENC[AES256_GCM,data:abc,iv:def,tag:ghi,type:str]")
	assert.Equal(t, "str", result)
}

// TestExtractTypeHint_NoType verifies extractTypeHint defaults to "str" when no type: suffix.
func TestExtractTypeHint_NoType(t *testing.T) {
	result := extractTypeHint("ENC[AES256_GCM,data:abc,iv:def,tag:ghi]")
	assert.Equal(t, "str", result)
}

// TestParseFile_NonStringValues verifies that int/bool YAML values don't cause panics.
func TestParseFile_NonStringValues(t *testing.T) {
	path := writeTempYAML(t, mixedTypesYAML)
	rule := sops.CreationRule{} // no encrypted_regex

	result, err := ParseFile(path, rule, true)
	require.NoError(t, err)

	// Should have count, enabled, name (sops hidden)
	assert.Len(t, result.Nodes, 3)
}

// TestParseFile_UnencryptedFile verifies that a file with isEncrypted=false shows all values as plaintext.
func TestParseFile_UnencryptedFile(t *testing.T) {
	path := writeTempYAML(t, unencryptedYAML)
	rule := sops.CreationRule{}

	result, err := ParseFile(path, rule, false)
	require.NoError(t, err)

	dbNode := result.Nodes[0]
	for _, child := range dbNode.Children {
		assert.False(t, child.Encrypted)
		assert.False(t, child.IsPlain)
		assert.NotEmpty(t, child.Value)
	}
}
