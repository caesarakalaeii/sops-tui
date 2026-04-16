---
phase: 02-read-loop
reviewed: 2026-04-14T00:00:00Z
depth: standard
files_reviewed: 11
files_reviewed_list:
  - cmd/sops-tui/main.go
  - internal/app/model.go
  - internal/keys/bindings.go
  - internal/parser/yaml.go
  - internal/sops/discoverer.go
  - internal/ui/detail.go
  - internal/ui/filelist.go
  - internal/ui/help.go
  - internal/ui/metadata.go
  - internal/ui/search.go
  - internal/ui/styles.go
findings:
  critical: 0
  warning: 2
  info: 4
  total: 6
status: issues_found
---

# Phase 02: Code Review Report

**Reviewed:** 2026-04-14
**Depth:** standard
**Files Reviewed:** 11
**Status:** issues_found

## Summary

All 11 source files were reviewed. The code is generally well-structured and follows the Bubble Tea v2 patterns correctly — `tea.KeyPressMsg`, `tea.View` return, `v.AltScreen = true`, `goccy/go-yaml`, explicit hex colors (no `AdaptiveColor`). No critical bugs, security vulnerabilities, or crashes were found.

Two warnings were identified: a logic bug that produces a stale breadcrumb (empty filename segment) when metadata is opened from the file list before visiting the detail view, and two `interface{}` parameter signatures that violate the CLAUDE.md "never use type any" rule. Four info-level items cover dead code and a fragile cross-package string coupling.

## Warnings

### WR-01: Stale `m.currentFile` used for breadcrumb in metadata overlay from file list state

**File:** `internal/app/model.go:242`

**Issue:** `m.currentFile` is only populated when the user drills into the detail view (line 301-309). If the user presses `i` from the file list before ever opening a file in detail, `m.currentFile.Name` is an empty string, producing a breadcrumb like `"sops-tui > files >  > metadata"` (note the blank segment). The selected file path and rule are correctly read from `SelectedFileItem()` at line 211, but the breadcrumb at line 242 ignores that value.

**Fix:** Capture the item name from `SelectedFileItem()` at line 211 and use it for the breadcrumb:

```go
// In the stateFileList branch (around line 210-215):
if m.state == stateFileList {
    if item, ok := m.fileList.SelectedFileItem(); ok {
        filePath = item.Path
        rule = item.Rule
        isEnc = item.IsEncrypted
        // Capture name for breadcrumb
        selectedName = item.Name  // declare var selectedName string above
    }
} else {
    filePath = m.currentFile.AbsPath
    rule = m.currentFile.Rule
    isEnc = m.currentFile.IsEncrypted
    selectedName = m.currentFile.Name
}
// ...
m.status.SetBreadcrumb("files", selectedName, "metadata")
```

---

### WR-02: `interface{}` parameter type violates "never use type any" mandate

**File:** `internal/parser/yaml.go:105,133`

**Issue:** Both `extractSopsMetadata(value interface{})` and `buildNode(key string, value interface{}, ...)` use `interface{}` as a parameter type. CLAUDE.md explicitly mandates "never use type any" (which is an alias for `interface{}`). The package comment on line 11 even repeats this rule. While the `goccy/go-yaml` library defines `MapItem.Value` as `interface{}`, the internal functions should not propagate this type into their own signatures when alternatives exist.

**Fix:** Pass the full `yaml.MapItem` to `buildNode` instead of extracting the value first, eliminating the naked `interface{}` in the internal API. For `extractSopsMetadata`, accept `yaml.MapItem` (the enclosing item) or keep it as a private helper with a comment explaining the constraint:

```go
// extractSopsMetadata receives the raw yaml.MapItem.Value from goccy/go-yaml.
// The interface{} type here is forced by the library's MapItem definition
// (yaml.MapItem.Value is interface{}). This is the only permitted use per
// CLAUDE.md — do not introduce additional interface{} params.
func extractSopsMetadata(value interface{}) SopsMetadata { // nolint:ireturn
```

Or refactor to pass the slice directly so the `interface{}` is confined to the switch inside `buildNode`:

```go
// buildNode receives a yaml.MapItem so callers don't hold interface{} variables.
func buildNode(item yaml.MapItem, depth int, rule sops.CreationRule, isEncrypted bool) ui.TreeNode {
    key, _ := item.Key.(string)
    switch v := item.Value.(type) {
    // ... same cases
    }
}
```

This keeps the type assertion contained to the single type-switch and removes the `interface{}` from the exported function surface.

---

## Info

### IN-01: `SelectedItem()` is dead production code — exact duplicate of `SelectedFileItem()`

**File:** `internal/ui/filelist.go:240-249`

**Issue:** `SelectedItem()` (lines 240-249) and `SelectedFileItem()` (lines 251-260) have byte-for-byte identical implementations. Only `SelectedFileItem()` is called from production code (`model.go`). `SelectedItem()` appears only in tests (`filelist_test.go`). The duplicate method creates maintenance risk: a future fix to one may be missed in the other.

**Fix:** Delete `SelectedItem()` and update the two test call sites in `filelist_test.go` to use `SelectedFileItem()`:

```go
// filelist_test.go line 60
got, ok := m.SelectedFileItem()

// filelist_test.go line 68
_, ok := m.SelectedFileItem()
```

---

### IN-02: No-op if-block in `buildContentLines` — dead code

**File:** `internal/ui/metadata.go:91-95`

**Issue:** The version field block reads:

```go
versionVal := m.meta.Version
if versionVal == "" {
    versionVal = ""  // no-op: assigning the same zero value
}
lines = append(lines, labelStyle.Render("version")+valueStyle.Render(versionVal))
```

The `if versionVal == ""` branch assigns `""` to a variable already holding `""`. This is a no-op and likely an incomplete copy-paste from the other fields that render `none` when empty.

**Fix:** Either render `(none)` consistently with the other fields, or remove the dead branch:

```go
versionVal := m.meta.Version
if versionVal == "" {
    lines = append(lines, labelStyle.Render("version")+noneStyle.Render("(none)"))
} else {
    lines = append(lines, labelStyle.Render("version")+valueStyle.Render(versionVal))
}
```

---

### IN-03: Fragile cross-package string coupling for env indicator flags

**File:** `cmd/sops-tui/main.go:47-49`

**Issue:** `hasResultWithMessage` performs substring matching against string literals hardcoded in `main.go` to derive the `EnvStatus` boolean flags:

```go
AgeAvailable:      !hasResultWithMessage(results, "Age key file not found"),
SopsYamlAvailable: !hasResultWithMessage(results, ".sops.yaml not found"),
```

The actual messages in `validator/startup.go` are `"Age key file not found"` and `".sops.yaml not found in current directory or parents"`. If either message string changes in the validator, the corresponding `EnvStatus` field silently flips to the wrong value with no compile-time error.

**Fix:** Export typed availability flags directly from the validator package rather than re-deriving them from message strings:

```go
// In validator/startup.go:
type CheckResult struct {
    Results      []ValidationResult
    HasHardError bool
    SopsFound    bool
    AgeFound     bool
    SopsYamlFound bool
}

// RunChecks returns CheckResult instead of ([]ValidationResult, bool).
```

Or, at minimum, export string constants the substring check can reference:

```go
// In validator/startup.go:
const MsgSopsMissing   = "sops binary not found"
const MsgAgeMissing    = "Age key file not found"
const MsgSopsYamlMissing = ".sops.yaml not found in current directory or parents"
```

---

### IN-04: `.sops.yaml` skip check uses `filepath.Base` — skips nested config files

**File:** `internal/sops/discoverer.go:86`

**Issue:** The walk skips files where `filepath.Base(path) == ".sops.yaml"`. This correctly skips the root `.sops.yaml` but would also skip any `.sops.yaml` found in subdirectories (e.g., `nested/.sops.yaml`). If a repository has nested SOPS configs that are themselves tracked as secret files (uncommon but possible), they would be silently excluded from discovery without any indication.

**Fix:** Compare the absolute path against the specific config file, not just the base name:

```go
// Skip only the specific .sops.yaml config file being processed, not all files named .sops.yaml
if absPath == absSOPS {
    return nil
}
```

Note: `absSOPS` is computed at line 58-62 and is in scope at the walk closure.

---

_Reviewed: 2026-04-14_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
