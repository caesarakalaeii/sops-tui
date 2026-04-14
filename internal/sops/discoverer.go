// Package sops provides SOPS file discovery services for sops-tui.
//
// Discoverer parses .sops.yaml creation_rules and walks the filesystem to find
// all files matching path_regex entries. Each matched file is checked for a
// top-level "sops:" YAML key to determine whether it is already encrypted.
//
// Security: Uses regexp.Compile (not MustCompile) so malformed regex rules are
// skipped gracefully (T-02-01). All discovered file paths are validated to be
// within the .sops.yaml parent directory (T-02-02).
package sops

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	yaml "github.com/goccy/go-yaml"
)

// SopsConfig represents the top-level .sops.yaml structure.
type SopsConfig struct {
	CreationRules []CreationRule `yaml:"creation_rules"`
}

// CreationRule represents a single creation_rule entry in .sops.yaml.
type CreationRule struct {
	PathRegex        string `yaml:"path_regex"`
	EncryptedRegex   string `yaml:"encrypted_regex"`
	UnencryptedRegex string `yaml:"unencrypted_regex"`
	Age              string `yaml:"age"`
}

// DiscoveredFile represents a file matched by .sops.yaml rules.
type DiscoveredFile struct {
	Name        string       // relative path from .sops.yaml directory
	AbsPath     string       // absolute path on disk
	IsEncrypted bool         // true if file contains top-level "sops:" key
	Rule        CreationRule // the first matching creation rule
}

// Discover parses the .sops.yaml at sopsYamlPath, walks the directory tree
// rooted at its parent, and returns all files matching any creation_rule path_regex.
// Each file is checked for a top-level "sops:" YAML key to determine IsEncrypted.
// Uses first-match-wins rule ordering, consistent with SOPS behavior.
func Discover(sopsYamlPath string) ([]DiscoveredFile, error) {
	data, err := os.ReadFile(sopsYamlPath)
	if err != nil {
		return nil, err
	}

	var cfg SopsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	absSOPS, err := filepath.Abs(sopsYamlPath)
	if err != nil {
		return nil, err
	}
	sopsYamlDir := filepath.Dir(absSOPS)

	var discovered []DiscoveredFile

	err = filepath.WalkDir(sopsYamlDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil // skip unreadable entries
		}

		// Skip hidden/vendor directories
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == ".svn" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only regular files
		if !d.Type().IsRegular() {
			return nil
		}

		// Skip the .sops.yaml config file itself
		if filepath.Base(path) == ".sops.yaml" {
			return nil
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil // skip on abs resolution failure
		}

		// Security: ensure file is within the sopsYamlDir (T-02-02)
		if !strings.HasPrefix(absPath, sopsYamlDir+string(filepath.Separator)) &&
			absPath != sopsYamlDir {
			return nil
		}

		rule, matched := matchRule(absPath, sopsYamlDir, cfg.CreationRules)
		if !matched {
			return nil
		}

		relPath, err := filepath.Rel(sopsYamlDir, absPath)
		if err != nil {
			return nil
		}

		discovered = append(discovered, DiscoveredFile{
			Name:        filepath.ToSlash(relPath),
			AbsPath:     absPath,
			IsEncrypted: hasSOPSMarker(absPath),
			Rule:        rule,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return discovered, nil
}

// matchRule returns the first CreationRule whose PathRegex matches the relative
// path of filePath from sopsYamlDir. An empty PathRegex matches all files
// (catch-all rule). Returns the matching rule and true, or zero value and false
// if no rule matches.
//
// Uses regexp.Compile (NOT MustCompile) — invalid regex rules are silently
// skipped, satisfying T-02-01.
func matchRule(filePath, sopsYamlDir string, rules []CreationRule) (CreationRule, bool) {
	relPath, err := filepath.Rel(sopsYamlDir, filePath)
	if err != nil {
		return CreationRule{}, false
	}
	relPath = filepath.ToSlash(relPath)

	for _, rule := range rules {
		// Empty PathRegex is a catch-all
		if rule.PathRegex == "" {
			return rule, true
		}

		re, err := regexp.Compile(rule.PathRegex)
		if err != nil {
			// Invalid regex: skip this rule (T-02-01)
			continue
		}

		if re.MatchString(relPath) {
			return rule, true
		}
	}

	return CreationRule{}, false
}

// hasSOPSMarker reads the YAML file at absPath and reports whether it contains
// a top-level "sops:" key. Uses yaml.UseOrderedMap() to preserve insertion
// order and avoid re-marshaling overhead.
func hasSOPSMarker(absPath string) bool {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return false
	}

	var root yaml.MapSlice
	if err := yaml.UnmarshalWithOptions(data, &root, yaml.UseOrderedMap()); err != nil {
		return false
	}

	for _, item := range root {
		key, ok := item.Key.(string)
		if ok && key == "sops" {
			return true
		}
	}

	return false
}
