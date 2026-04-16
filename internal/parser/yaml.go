// Package parser provides encrypted YAML tree extraction for sops-tui.
//
// ParseFile reads a SOPS-encrypted (or plain) YAML file and returns an ordered
// TreeNode tree suitable for display in DetailModel. The sops: metadata block is
// hidden from the tree and its contents are surfaced through SopsMetadata (D-05).
//
// Security:
//   - T-02-03: File size guard rejects files > 10MB before parsing.
//   - T-02-04: Type switch (not assertion) prevents panics on unexpected YAML types.
//
// Note: Never use type any. Use concrete types throughout.
package parser

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	yaml "github.com/goccy/go-yaml"

	"github.com/caesarakalaeii/sops-tui/internal/sops"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

// SopsMetadata holds metadata extracted from the sops: block of an encrypted file.
type SopsMetadata struct {
	Version          string
	LastModified     string
	MAC              string
	AgeRecipients    []string
	EncryptedRegex   string
	UnencryptedRegex string
}

// ParsedFile is the result of parsing a SOPS-encrypted YAML file.
type ParsedFile struct {
	Nodes    []ui.TreeNode
	Metadata SopsMetadata
}

// sopsBlock is a helper struct for unmarshalling the sops: metadata block.
type sopsBlock struct {
	Version          string     `yaml:"version"`
	LastModified     string     `yaml:"lastmodified"`
	MAC              string     `yaml:"mac"`
	Age              []ageEntry `yaml:"age"`
	EncryptedRegex   string     `yaml:"encrypted_regex"`
	UnencryptedRegex string     `yaml:"unencrypted_regex"`
}

// ageEntry represents a single age recipient entry in the sops: block.
type ageEntry struct {
	Recipient string `yaml:"recipient"`
}

// ParseFile reads a YAML file at absPath and returns the tree structure and SOPS metadata.
// The sops: top-level key is excluded from the tree (D-05) and extracted into Metadata.
// Encrypted leaf values (starting with "ENC[") get Encrypted=true and TypeHint from the type: suffix.
// The rule parameter determines which plain values get IsPlain=true based on encrypted_regex/unencrypted_regex.
// If isEncrypted is false (file has no sops: block), all values are shown as plaintext.
func ParseFile(absPath string, rule sops.CreationRule, isEncrypted bool) (ParsedFile, error) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return ParsedFile{}, fmt.Errorf("read %q: %w", absPath, err)
	}

	// T-02-03: File size guard — reject files > 10MB before parsing.
	if len(data) > 10*1024*1024 {
		return ParsedFile{}, fmt.Errorf("file too large (max 10MB)")
	}

	var root yaml.MapSlice
	if err := yaml.UnmarshalWithOptions(data, &root, yaml.UseOrderedMap()); err != nil {
		return ParsedFile{}, fmt.Errorf("parse %q: %w", absPath, err)
	}

	var meta SopsMetadata
	var nodes []ui.TreeNode

	for _, item := range root {
		key, ok := item.Key.(string)
		if !ok {
			continue
		}

		if key == "sops" {
			// Extract metadata from sops: block and skip it from the tree (D-05).
			meta = extractSopsMetadata(item.Value)
			continue
		}

		node := buildNode(key, item.Value, 0, rule, isEncrypted)
		nodes = append(nodes, node)
	}

	return ParsedFile{
		Nodes:    nodes,
		Metadata: meta,
	}, nil
}

// extractSopsMetadata parses the value of the sops: key into a SopsMetadata struct.
// Re-marshals the value to YAML then unmarshals into sopsBlock to handle nested types.
func extractSopsMetadata(value interface{}) SopsMetadata {
	raw, err := yaml.Marshal(value)
	if err != nil {
		return SopsMetadata{}
	}

	var block sopsBlock
	if err := yaml.Unmarshal(raw, &block); err != nil {
		return SopsMetadata{}
	}

	recipients := make([]string, 0, len(block.Age))
	for _, a := range block.Age {
		recipients = append(recipients, a.Recipient)
	}

	return SopsMetadata{
		Version:          block.Version,
		LastModified:     block.LastModified,
		MAC:              block.MAC,
		AgeRecipients:    recipients,
		EncryptedRegex:   block.EncryptedRegex,
		UnencryptedRegex: block.UnencryptedRegex,
	}
}

// buildNode constructs a ui.TreeNode for a single YAML key-value pair.
// Uses a type switch (not assertion) per T-02-04 to prevent panics on unexpected types.
func buildNode(key string, value interface{}, depth int, rule sops.CreationRule, isEncrypted bool) ui.TreeNode {
	switch v := value.(type) {
	case yaml.MapSlice:
		// Group node — recurse into children
		children := walkMapSlice(v, depth+1, rule, isEncrypted)
		return ui.TreeNode{
			Key:      key,
			Children: children,
			Depth:    depth,
			Expanded: false,
		}

	case string:
		if strings.HasPrefix(v, "ENC[") {
			// Encrypted leaf — extract type hint from ENC envelope
			return ui.TreeNode{
				Key:       key,
				Value:     v,
				Depth:     depth,
				Encrypted: true,
				TypeHint:  extractTypeHint(v),
			}
		}
		// Plain string value
		plain := isPlainValue(key, rule, isEncrypted)
		displayValue := ""
		if plain || !isEncrypted {
			displayValue = v
		}
		return ui.TreeNode{
			Key:     key,
			Value:   displayValue,
			Depth:   depth,
			IsPlain: plain,
		}

	case int:
		displayValue := fmt.Sprintf("%v", v)
		return ui.TreeNode{
			Key:     key,
			Value:   displayValue,
			Depth:   depth,
			IsPlain: isEncrypted,
		}

	case int64:
		displayValue := fmt.Sprintf("%v", v)
		return ui.TreeNode{
			Key:     key,
			Value:   displayValue,
			Depth:   depth,
			IsPlain: isEncrypted,
		}

	case float64:
		displayValue := fmt.Sprintf("%v", v)
		return ui.TreeNode{
			Key:     key,
			Value:   displayValue,
			Depth:   depth,
			IsPlain: isEncrypted,
		}

	case bool:
		displayValue := fmt.Sprintf("%v", v)
		return ui.TreeNode{
			Key:     key,
			Value:   displayValue,
			Depth:   depth,
			IsPlain: isEncrypted,
		}

	default:
		// Fallback for any other type
		return ui.TreeNode{
			Key:   key,
			Value: fmt.Sprintf("%v", v),
			Depth: depth,
		}
	}
}

// walkMapSlice recursively converts a yaml.MapSlice into []ui.TreeNode.
func walkMapSlice(ms yaml.MapSlice, depth int, rule sops.CreationRule, isEncrypted bool) []ui.TreeNode {
	nodes := make([]ui.TreeNode, 0, len(ms))
	for _, item := range ms {
		key, ok := item.Key.(string)
		if !ok {
			key = fmt.Sprintf("%v", item.Key)
		}
		nodes = append(nodes, buildNode(key, item.Value, depth, rule, isEncrypted))
	}
	return nodes
}

// extractTypeHint parses the type hint from a SOPS ENC string.
// SOPS ENC format: ENC[AES256_GCM,data:...,iv:...,tag:...,type:str]
// Returns "str" as default when no ",type:" suffix is found.
func extractTypeHint(enc string) string {
	idx := strings.LastIndex(enc, ",type:")
	if idx < 0 {
		return "str"
	}
	suffix := enc[idx+len(",type:"):]
	// Strip the closing "]"
	suffix = strings.TrimSuffix(suffix, "]")
	if suffix == "" {
		return "str"
	}
	return suffix
}

// isPlainValue determines whether a non-ENC string value in a SOPS file should
// be flagged as a plain (unencrypted) value with the [plain] badge.
//
// Logic:
//   - File NOT encrypted -> false (whole file is plaintext, no badge needed)
//   - File encrypted AND rule has EncryptedRegex: key does NOT match -> was left unencrypted -> true
//   - File encrypted AND rule has UnencryptedRegex: key DOES match -> explicitly left unencrypted -> true
//   - File encrypted AND rule has neither regex -> return false (anomalous, no context)
func isPlainValue(key string, rule sops.CreationRule, isEncrypted bool) bool {
	if !isEncrypted {
		return false
	}

	if rule.EncryptedRegex != "" {
		re, err := regexp.Compile(rule.EncryptedRegex)
		if err != nil {
			return false
		}
		// If key does NOT match encrypted_regex, it was left unencrypted
		return !re.MatchString(key)
	}

	if rule.UnencryptedRegex != "" {
		re, err := regexp.Compile(rule.UnencryptedRegex)
		if err != nil {
			return false
		}
		// If key DOES match unencrypted_regex, it was explicitly left unencrypted
		return re.MatchString(key)
	}

	// Neither regex set: can't determine, return false
	return false
}
