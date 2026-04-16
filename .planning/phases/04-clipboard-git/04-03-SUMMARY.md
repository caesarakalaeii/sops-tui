---
phase: 04-clipboard-git
plan: "03"
subsystem: ui/git-history
tags: [git, history, search, overlay, tui]
dependency_graph:
  requires: [04-02]
  provides: [git-history-overlay, cross-file-search]
  affects: [internal/app/model.go, internal/ui/filelist.go]
tech_stack:
  added: []
  patterns:
    - stateHistory added to sessionState enum following prevState pattern
    - CrossFileSearchItem as domain type with Title() display method
    - Lazy cache invalidation on FilesDiscoveredMsg for cross-file items
    - Async tea.Cmd for git history fetch (T-04-10 mitigation)
key_files:
  created:
    - internal/ui/history.go
    - internal/ui/history_test.go
  modified:
    - internal/app/model.go
    - internal/ui/filelist.go
decisions:
  - Footer shown on empty state too (not just entries state) — consistent UX
  - Cross-file Enter key intercepted in model.go before FileListModel.Update to allow navigation
  - populateCrossFileItems() as AppModel method (not standalone function) to access m.files
metrics:
  duration: "6 minutes"
  completed: "2026-04-15"
  tasks_completed: 2
  files_changed: 4
---

# Phase 4 Plan 3: Git History Overlay and Cross-File Search Summary

Git history overlay (HistoryModel) and cross-file fuzzy search completing Phase 4 requirements GIT-02 and GIT-03.

## What Was Built

### Task 1: Git History Overlay (TDD)

**internal/ui/history.go** — HistoryModel full-screen overlay component:
- `NewHistoryModel(filename, width, height)` starts in `loading=true` state
- `SetEntries([]git.CommitEntry)` transitions from loading to content/empty
- `ScrollDown()` / `ScrollUp()` clamp within entry bounds
- `View()` renders: title `git log -- filename`, entry rows with fixed-width columns (hash 9, date 16, author 18, subject free), footer `j/k scroll  b or esc close`
- Loading state shows `Loading history...` without footer
- Empty state shows `No commits found` + body text + footer

**internal/ui/history_test.go** — 6 TDD tests covering all rendering scenarios, all passing.

**internal/app/model.go** — stateHistory wiring:
- `stateHistory` added to sessionState enum after `stateFormatMenu`
- `StateHistory` exported constant for tests
- `HistoryRequestMsg` and `GitHistoryMsg` message types
- `AppModel.history ui.HistoryModel` field
- `b` key: opens overlay from `stateDetail` (not during search), flashes error if no git repo, triggers async `gitpkg.GetFileHistory(repoRoot, relPath, 50)` returning `GitHistoryMsg`
- `b` key while `stateHistory`: closes overlay, restores breadcrumb
- `GitHistoryMsg` handler: on error flash + revert state; on success call `m.history.SetEntries()`
- Esc chain: `stateHistory` closes overlay after `stateMetadata`
- View renders `stateHistory` case via `m.history.View()`
- `j/k` scroll in `stateHistory` routing block
- `WindowSizeMsg` propagates `m.history.SetSize()`

### Task 2: Cross-File Search (GIT-03)

**internal/app/model.go** — Cross-file search infrastructure:
- `CrossFileSearchItem` struct with `FileName`, `KeyPath`, `AbsPath`, `Rule`, `IsEnc`, `GitStatus`
- `Title()` method: returns `filename > key.path` or just `filename` for file-level items
- `AppModel.crossFileItems []CrossFileSearchItem` and `crossFilePopulated bool` fields
- `populateCrossFileItems()`: lazy population, parses all files via `parser.ParseFile` for key paths, caches result
- `collectKeyPaths()`: recursive helper building dot-joined paths from `[]ui.TreeNode`
- Cache invalidated (`crossFilePopulated = false`, `crossFileItems = nil`) on `FilesDiscoveredMsg`
- `/` from `stateFileList`: calls `populateCrossFileItems()`, builds titles slice, calls `m.fileList.ActivateCrossFileSearch(titles)`
- Enter during cross-file search: reads `SelectedCrossFileIndex()`, constructs `sops.DiscoveredFile`, dispatches `parser.ParseFile` async → `FilesParsedMsg`

**internal/ui/filelist.go** — Cross-file search support:
- `CrossFileListItem` struct implementing `list.Item` and `list.DefaultItem` with `OrigIndex` for back-reference
- `FileListModel.crossFileMode bool` and `crossFileTitles []string` fields
- `ActivateCrossFileSearch(titles []string) tea.Cmd`: sets mode, stores titles, activates search input
- `DeactivateSearch()` updated to reset `crossFileMode` and `crossFileTitles`
- `IsCrossFileMode() bool` accessor
- `SelectedCrossFileIndex() int`: type-asserts `list.SelectedItem()` to `CrossFileListItem`
- `Update()` cross-file branch: filters `crossFileTitles` (not `allItems`), populates list with `CrossFileListItem` entries; Enter in cross-file mode returns early for model.go to handle

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Footer missing from empty state**
- **Found during:** Task 1 TDD GREEN phase
- **Issue:** Plan specified footer only in entry-rendering branch; test `renders_footer_with_scroll_and_close_hints` sets empty entries and expects footer present
- **Fix:** Moved footer construction before the loading/empty/entries branch split; added footer to empty state path; loading state deliberately omits footer
- **Files modified:** internal/ui/history.go
- **Commit:** f779c2c

None other — plan executed as written.

## Known Stubs

None.

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| threat_flag: denial-of-service | internal/app/model.go | populateCrossFileItems() calls parser.ParseFile for every discovered file synchronously on first / press; mitigated by lazy cache (T-04-09) but large repos with many files could cause noticeable lag |

Note: T-04-09 is in the plan's threat model as `mitigate` with lazy population + caching. The implementation follows the prescribed mitigation. Flagged here for verifier awareness.

## Self-Check: PASSED

- internal/ui/history.go: FOUND
- internal/ui/history_test.go: FOUND
- 47cfba0 (test: failing tests): FOUND
- f779c2c (feat: HistoryModel + stateHistory): FOUND
- 4ad8c7f (feat: cross-file search): FOUND
