---
phase: "01"
plan: "01"
subsystem: foundation
tags: [go-module, design-system, keybindings, tui, lipgloss, bubbles]
dependency_graph:
  requires: []
  provides:
    - go.mod with all Phase 1 charm.land v2 dependencies
    - internal/ui/styles.go — design token file (colors, spacing, named styles)
    - internal/keys/bindings.go — keybinding contracts (FileListKeyMap, DetailKeyMap, GlobalKeyMap)
  affects:
    - All subsequent Phase 1 plans (02, 03, 04) import from styles.go and bindings.go
tech_stack:
  added:
    - charm.land/bubbletea/v2 v2.0.4
    - charm.land/lipgloss/v2 v2.0.3
    - charm.land/bubbles/v2 v2.1.0
    - filippo.io/age v1.3.1
    - golang.org/x/term v0.42.0
    - github.com/stretchr/testify v1.11.1
  patterns:
    - lipgloss.Color() function (v2 API) returns color.Color interface — not a string type
    - Hex string constants (*Hex) exported alongside color.Color vars for testability
    - KeyMap structs embed GlobalKeyMap to propagate global keybindings
    - help.KeyMap interface implemented via ShortHelp()/FullHelp() methods
key_files:
  created:
    - go.mod — module with all Phase 1 dependencies
    - go.sum — locked dependency checksums
    - cmd/sops-tui/main.go — placeholder entry point (wired in Plan 04)
    - internal/ui/styles.go — 8 color constants, 6 spacing tokens, 12 named styles
    - internal/ui/styles_test.go — hex value assertions and render non-empty checks
    - internal/keys/bindings.go — GlobalKeyMap, FileListKeyMap, DetailKeyMap
    - internal/keys/bindings_test.go — 14 binding behavior tests via key.Matches
  modified:
    - .gitignore — anchor /sops-tui binary pattern; add tea_debug.log exclusion
decisions:
  - "Export hex color values as *Hex string constants alongside color.Color vars — lipgloss v2 Color() returns color.Color interface (not string), so direct string conversion is impossible; hex constants enable exact value testing"
  - "Back and Collapse both bind 'h' key in DetailKeyMap — per 01-UI-SPEC.md spec: h collapses expanded nodes AND serves as the back gesture from top-level; both bindings are active and tested"
metrics:
  duration: "7 minutes"
  completed: "2026-04-14"
  tasks: 3
  files_created: 7
  files_modified: 1
---

# Phase 01 Plan 01: Go Module + Design System + Keybindings Summary

**One-liner:** Go module with charm.land v2 deps, Catppuccin Mocha color palette via explicit hex constants, and vim-style keybinding contracts for file list and detail views.

## What Was Built

### Task 1: Go Module Initialization (commit e72c6c6)

Initialized the Go module with all Phase 1 dependencies. Created the project directory skeleton. Fixed `.gitignore` to anchor the binary exclusion pattern so `cmd/sops-tui/` is not accidentally ignored.

Key package versions resolved:
- `charm.land/bubbletea/v2 v2.0.4` — TEA event loop
- `charm.land/lipgloss/v2 v2.0.3` — styling (note: latest at time of install; plan specified v2.x)
- `charm.land/bubbles/v2 v2.1.0` — TUI components including key and help packages
- `filippo.io/age v1.3.1` — age key parsing (kept as indirect; no source imports yet)
- `golang.org/x/term v0.42.0` — terminal width detection (kept as indirect; no source imports yet)

### Task 2: Design System (commit 2470c75)

Created `internal/ui/styles.go` with the full Catppuccin Mocha palette as explicit hex constants per 01-UI-SPEC.md. No `lipgloss.AdaptiveColor` usage — confirmed avoided.

**Design discovery:** In lipgloss v2, `lipgloss.Color()` is a function returning `color.Color` interface, not a `type Color string` as in v1. This means `string(lipgloss.Color("#hex"))` does not compile. Solution: export parallel `*Hex` string constants (`ColorBgHex`, etc.) so tests can assert exact hex values, while the `color.Color` vars remain available for style composition.

### Task 3: Keybinding Contracts (commit aad3132)

Created `internal/keys/bindings.go` with three KeyMap structs. Both `FileListKeyMap` and `DetailKeyMap` embed `GlobalKeyMap` (Rule D-09: global keys in every context) and implement `help.KeyMap` interface via `ShortHelp()` and `FullHelp()` methods.

All 14 behavior tests pass using `key.Matches()` with a local `keyStringer` test helper that satisfies `fmt.Stringer`.

## Verification Results

```
go build ./...                    PASS
go test ./internal/ui/... -v      PASS (4 test functions, 26 sub-tests)
go test ./internal/keys/... -v    PASS (5 test functions)
AdaptiveColor usage               NONE (only in comments)
```

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] lipgloss v2 Color type incompatible with string conversion**
- **Found during:** Task 2 (TDD RED phase)
- **Issue:** `string(ui.ColorBg)` does not compile — lipgloss v2 changed `type Color string` to `func Color(string) color.Color`. The plan's test behavior described testing via string conversion.
- **Fix:** Exported parallel `*Hex` string constants (`ColorBgHex = "#1e1e2e"`, etc.) as the canonical testable values. The `color.Color` vars remain for style use. Tests assert on `*Hex` constants instead of converting the `color.Color` interface.
- **Files modified:** `internal/ui/styles.go` (added Hex constants), `internal/ui/styles_test.go` (test against Hex constants)
- **Commit:** 2470c75

**2. [Rule 3 - Blocking] `.gitignore` pattern `sops-tui` blocked `cmd/sops-tui/` directory**
- **Found during:** Task 1 commit staging
- **Issue:** `git add cmd/sops-tui/main.go` failed — the `sops-tui` pattern in `.gitignore` matched the `cmd/sops-tui/` directory, not just the binary.
- **Fix:** Changed `sops-tui` to `/sops-tui` (root-anchored) so only the top-level binary is excluded.
- **Files modified:** `.gitignore`
- **Commit:** e72c6c6

**3. [Rule 2 - Auto-add] `tea_debug.log` generated by lipgloss test runs**
- **Found during:** Task 2 post-test
- **Issue:** Running lipgloss tests created a `tea_debug.log` file in `internal/ui/` that would pollute the repo if committed.
- **Fix:** Added `tea_debug.log` to `.gitignore`.
- **Files modified:** `.gitignore`
- **Commit:** 2470c75

## Known Stubs

| File | Line | Content | Reason |
|------|------|---------|--------|
| `cmd/sops-tui/main.go` | 4-5 | Empty `main()` with comment | Intentional — wired in Plan 04 per plan spec |

## Self-Check: PASSED

**Files exist:**
- FOUND: go.mod
- FOUND: go.sum
- FOUND: cmd/sops-tui/main.go
- FOUND: internal/ui/styles.go
- FOUND: internal/ui/styles_test.go
- FOUND: internal/keys/bindings.go
- FOUND: internal/keys/bindings_test.go
- FOUND: internal/app/ (directory)
- FOUND: internal/validator/ (directory)

**Commits exist:**
- FOUND: e72c6c6 (chore: Go module initialization)
- FOUND: 2470c75 (feat: design system styles)
- FOUND: aad3132 (feat: keybinding contracts)
