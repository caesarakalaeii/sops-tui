---
phase: 01-foundation
plan: "03"
subsystem: ui-components
tags: [tui, bubbles, filelist, detail, yaml-tree, vim-navigation, tdd]
dependency_graph:
  requires: ["01-01"]
  provides: ["FileListModel", "DetailModel", "TreeNode"]
  affects: ["01-04"]
tech_stack:
  added: []
  patterns:
    - "bubbles/list wrapping with custom vim keybindings (g/G/ctrl-d/u intercepted before list delegation)"
    - "flat row rendering from recursive tree flattening (flattenNodes)"
    - "pointer-based node mutation for expand/collapse state"
    - "ANSI-safe tree connector rendering via lipgloss styles"
key_files:
  created:
    - internal/ui/filelist.go
    - internal/ui/filelist_test.go
    - internal/ui/detail.go
    - internal/ui/detail_test.go
  modified:
    - go.mod
    - go.sum
decisions:
  - "Intercept g/G/ctrl-d/u before delegating to bubbles/list — list's default KeyMap doesn't cover vim half-page scroll"
  - "flattenNodes uses pointer slice indexing (&nodes[i]) to enable in-place Expanded mutation through flatRow.node"
  - "h key is both Collapse and Back in DetailKeyMap — Collapse handled first; Back handled by root model in Plan 04"
  - "DetailModel.View() renders all rows without bubbles/viewport — sufficient for Phase 1 placeholder data"
  - "go.mod/go.sum updated to pull in list transitive deps (sahilm/fuzzy, atotto/clipboard)"
metrics:
  duration: "~15 minutes"
  completed: "2026-04-14"
  tasks_completed: 2
  files_created: 4
  files_modified: 2
---

# Phase 1 Plan 03: File List and Detail View Components Summary

**One-liner:** FileListModel wrapping bubbles/list with vim navigation and DetailModel rendering a collapsible YAML tree with ├─/└─/│ connectors and [+]/[-] indicators.

## What Was Built

### Task 1: FileListModel

`internal/ui/filelist.go` implements the file browser pane:

- `FileItem` struct satisfying `list.Item` (FilterValue) and `list.DefaultItem` (Title, Description)
- `FileListModel` wraps `charm.land/bubbles/v2/list.Model` with `keys.FileListKeyMap`
- Navigation keys g/G/ctrl-d/u intercepted before bubbles/list delegation; j/k/up/down handled by list
- Built-in list chrome disabled: `SetShowHelp(false)`, `SetShowStatusBar(false)`, `SetShowFilter(false)`, `SetFilteringEnabled(false)`
- Empty state renders exact UI-SPEC copywriting: "No SOPS files found" with body text in DimText style
- `SelectedItem() (FileItem, bool)` typed accessor with ok-pattern
- `ItemCount() int` for parent model consumption
- 7 tests all pass

### Task 2: DetailModel

`internal/ui/detail.go` implements the YAML tree detail pane:

- `TreeNode` struct with Key, Value, Children []TreeNode, Expanded bool, Depth int
- `flatRow` internal type tracking depth, isLast, parentIsLast for connector rendering
- `flattenNodes` recursive function producing visible rows (only expanded children included)
- Vim navigation: j/k, g/G, ctrl-u/d with scroll clamping; `adjustScroll` keeps cursor in viewport
- Expand: Enter/l/right sets Expanded=true, rebuilds flatRows
- Collapse: h/left sets Expanded=false, rebuilds flatRows, clamps cursor
- Tree connectors in TreeConnector style (muted): ├─ (non-last), └─ (last), │ (ancestor continuation)
- Group indicators in TreeIndicator style (accent): [+] collapsed, [-] expanded
- Leaf values in DimText style (faint): key: ***
- Selected row highlighted with SelectedRow.Width(m.width) for full-width highlight
- Empty state: "No keys found in this file" per UI-SPEC copywriting contract
- 12 tests all pass

## Commits

| Hash | Message |
|------|---------|
| 1b6c3b8 | test(01-03): add failing tests for FileListModel |
| 74b7889 | feat(01-03): implement FileListModel wrapping bubbles/list |
| 8fe1cb6 | test(01-03): add failing tests for DetailModel collapsible YAML tree |
| 4202a4a | feat(01-03): implement DetailModel with collapsible YAML tree |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Missing go.sum entries for bubbles/list transitive deps**
- **Found during:** Task 1 GREEN phase
- **Issue:** `charm.land/bubbles/v2/list` imports `github.com/sahilm/fuzzy` and `atotto/clipboard` which were not in go.sum
- **Fix:** Ran `go get charm.land/bubbles/v2/list@v2.1.0 && go mod tidy` to populate go.sum
- **Files modified:** go.mod, go.sum
- **Commit:** 74b7889

**2. [Rule 1 - Bug] Test compile error: `key.Matches` is generic in bubbles v2**
- **Found during:** Task 2 RED phase
- **Issue:** `var _ = func() { _ = key.Matches }` failed with "cannot use generic function key.Matches without instantiation"
- **Fix:** Replaced compile-time check with `var _ = key.NewBinding` (non-generic function reference)
- **Files modified:** internal/ui/detail_test.go
- **Commit:** Included in 8fe1cb6 before task GREEN phase

## Known Stubs

None. Both components render placeholder data ("***" for all values, hardcoded test items) as explicitly intended by Plan 03 — real data wiring is deferred to Phase 2 (file discovery, YAML parsing). The "***" masking is the correct Phase 1 behavior per plan spec.

## Threat Flags

None. Plan 03 creates UI components with no I/O, no subprocess calls, no file reads per the threat model. Key names flow through lipgloss style functions (ANSI sanitization). No new network endpoints or trust boundaries introduced.

## Self-Check: PASSED

Files created:
- FOUND: internal/ui/filelist.go
- FOUND: internal/ui/filelist_test.go
- FOUND: internal/ui/detail.go
- FOUND: internal/ui/detail_test.go

Commits:
- FOUND: 1b6c3b8
- FOUND: 74b7889
- FOUND: 8fe1cb6
- FOUND: 4202a4a

Tests:
- `go test ./internal/ui/... -run TestFileList -v -count=1`: 7/7 PASS
- `go test ./internal/ui/... -run TestDetail -v -count=1`: 12/12 PASS
- `go test ./... -count=1`: all packages PASS
- `go vet ./internal/ui/...`: clean
