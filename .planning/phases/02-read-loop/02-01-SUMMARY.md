---
phase: 02-read-loop
plan: 01
subsystem: backend-services
tags: [sops, yaml-parser, file-discovery, tdd]
dependency_graph:
  requires: []
  provides:
    - internal/sops/discoverer.go (SopsDiscoverer: Discover, DiscoveredFile, SopsConfig, CreationRule)
    - internal/parser/yaml.go (YamlParser: ParseFile, ParsedFile, SopsMetadata)
    - internal/ui/detail.go (TreeNode extended with Encrypted, TypeHint, IsPlain)
  affects:
    - internal/ui/detail.go (renderRow updated for type-hinted masked values and plain badges)
tech_stack:
  added:
    - github.com/goccy/go-yaml v1.19.2 (direct dependency)
  patterns:
    - TDD (RED-GREEN cycle per task)
    - Pure Go services with zero Bubbletea dependency
    - Type switch on yaml.MapSlice values (T-02-04)
    - regexp.Compile not MustCompile for graceful invalid regex (T-02-01)
    - filepath.Abs + strings.HasPrefix for directory traversal prevention (T-02-02)
    - 10MB file size guard before YAML parsing (T-02-03)
key_files:
  created:
    - internal/sops/discoverer.go
    - internal/sops/discoverer_test.go
    - internal/parser/yaml.go
    - internal/parser/yaml_test.go
  modified:
    - internal/ui/detail.go
    - go.mod
    - go.sum
decisions:
  - Used yaml.UseOrderedMap() in both discoverer (hasSOPSMarker) and parser (ParseFile) to preserve YAML key insertion order
  - Skipped .sops.yaml file itself during WalkDir to avoid it appearing as a discovered secret file
  - Temporary lipgloss styles (typeHintStyleTemp, badgePlainTemp) in detail.go; canonical styles from styles.go deferred to Plan 02-03 (Wave 2)
  - isPlainValue returns false when neither EncryptedRegex nor UnencryptedRegex is set (anomalous case)
metrics:
  duration: ~20 minutes
  completed: 2026-04-14
  tasks_completed: 2
  files_changed: 7
---

# Phase 2 Plan 1: SopsDiscoverer and YamlParser Summary

**One-liner:** Pure Go backend services for SOPS file discovery (path_regex + sops: marker) and encrypted YAML tree extraction with type hints, metadata, and plain value detection using goccy/go-yaml.

## What Was Built

### Task 1: SopsDiscoverer (`internal/sops/discoverer.go`)

SOPS file discovery service that parses `.sops.yaml` creation rules and walks the filesystem to find all managed files. Key behaviors:

- `Discover(sopsYamlPath string) ([]DiscoveredFile, error)` — entry point; walks directory tree from `.sops.yaml` parent
- `matchRule` — first-match-wins rule ordering using `regexp.Compile` (not MustCompile); skips invalid regex rules gracefully
- `hasSOPSMarker` — checks top-level YAML keys for `"sops:"` using `yaml.UseOrderedMap()`
- Directory traversal prevention: all returned `AbsPath` values validated to start with `sopsYamlDir` prefix
- Skips `.git`, `.svn`, `node_modules` and the `.sops.yaml` file itself during WalkDir
- `goccy/go-yaml v1.19.2` added as direct dependency

**Tests (10):** encrypted file, unencrypted file, catch-all rule, first-match-wins, creation rule fields, hasSOPSMarker true/false, matchRule relative path, invalid regex safety, directory traversal safety.

### Task 2: YamlParser (`internal/parser/yaml.go`) + TreeNode extension (`internal/ui/detail.go`)

**TreeNode struct extended** with three new fields:
- `Encrypted bool` — leaf holds a SOPS-encrypted ENC[...] value
- `TypeHint string` — type code from ENC envelope (e.g., "str", "int", "bool")
- `IsPlain bool` — leaf was left unencrypted by encrypted_regex/unencrypted_regex

**renderRow updated** to display:
- Encrypted leaf: `key: *** (str)` using temporary `typeHintStyleTemp`
- Plain leaf: `key: value  [plain]` using temporary `badgePlainTemp`
- Default fallback: `key: ***` (Phase 1 behavior preserved)

**ParseFile** reads YAML, extracts `SopsMetadata` from `sops:` block (hidden from tree per D-05), walks remaining keys recursively via type switch, applies `extractTypeHint` and `isPlainValue` logic.

**Tests (15):** key order, sops key hidden, metadata version/lastmodified/recipients/encryptedRegex, encrypted leaves for str/int/bool, plain value detection, nested nodes with depth, extractTypeHint with and without type:, non-string values no panic, unencrypted file plaintext display.

## Verification Results

```
ok  github.com/caesarakalaeii/sops-tui/internal/sops    (10 tests)
ok  github.com/caesarakalaeii/sops-tui/internal/parser  (15 tests)
ok  github.com/caesarakalaeii/sops-tui/internal/ui      (all existing tests pass)
ok  github.com/caesarakalaeii/sops-tui/internal/app
ok  github.com/caesarakalaeii/sops-tui/internal/validator
go vet ./... — no warnings
```

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `.sops.yaml` itself matched catch-all creation rule**
- **Found during:** Task 1 — TestDiscover_CatchAllRule had 3 results instead of 2
- **Issue:** WalkDir traversed the `.sops.yaml` file itself; with an empty PathRegex (catch-all), it would appear as a discovered secret file
- **Fix:** Added `filepath.Base(path) == ".sops.yaml"` skip guard in WalkDir callback before rule matching
- **Files modified:** `internal/sops/discoverer.go`
- **Commit:** 25d1cd9

No other deviations — plan executed as written.

## Threat Mitigations Applied

| Threat | Mitigation |
|--------|------------|
| T-02-01 | `regexp.Compile` used everywhere; invalid path_regex rules skipped with `continue` |
| T-02-02 | `filepath.Abs` + `strings.HasPrefix(absPath, sopsYamlDir+sep)` validates all discovered paths |
| T-02-03 | `len(data) > 10*1024*1024` guard in `ParseFile` returns error before parsing |
| T-02-04 | Full type switch in `buildNode`; no type assertions on YAML values |

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| Task 1 | 25d1cd9 | feat(02-01): implement SopsDiscoverer with file discovery from .sops.yaml |
| Task 2 | 77fae06 | feat(02-01): implement YamlParser and extend TreeNode for encrypted YAML display |

## Self-Check: PASSED
