---
phase: 05-power-features
plan: 03
subsystem: app-integration
tags: [integration, state-machine, bulk-rekey, health-check, recipient-management, file-selection]
dependency_graph:
  requires: [05-01, 05-02]
  provides: [end-to-end Phase 5 features wired into AppModel and FileListModel]
  affects: [internal/app/model.go, internal/ui/filelist.go]
tech_stack:
  added: []
  patterns:
    - "bulkReKeyState struct for sequential queue management"
    - "recipientAction sentinel string to distinguish healthcheck vs. recipient confirm in stateDiff"
    - "currentParsed parser.ParsedFile stored on FilesParsedMsg for recipient access"
    - "runHealthCheck package-level func dispatched as tea.Cmd goroutine"
    - "flattenYAML recursive map[string]interface{} traversal for health analysis"
key_files:
  created: []
  modified:
    - internal/ui/filelist.go
    - internal/ui/filelist_test.go
    - internal/app/model.go
    - internal/app/model_test.go
decisions:
  - "Health check confirmation reuses stateDiff with recipientAction='healthcheck' sentinel rather than a dedicated state — avoids adding a 6th new state for a simple y/n gate"
  - "Bulk re-key uses sops rotate -i (data key rotation) not recipient modification — per D-05/D-06 spec which says 'rotate data key on selected files'"
  - "statusBarHeight called as statusBarHeight(*m) in pointer receiver methods to match value receiver signature"
  - "Keys accessed via keys.DefaultFileListKeyMap/keys.DefaultDetailKeyMap directly (existing pattern in model.go) rather than adding exported Keys() methods to FileListModel/DetailModel"
metrics:
  duration: "~45 minutes"
  completed: "2026-04-16"
  tasks_completed: 2
  tasks_total: 2
  files_modified: 4
---

# Phase 5 Plan 03: AppModel Integration Summary

Integration plan that wires all Phase 5 features from Plans 01 and 02 into a working end-to-end experience. FileListModel gains file selection and bulk re-key support. AppModel gains 5 new session states, async health check pipeline, recipient add/remove flows, and complete Esc chain handling.

## Tasks Completed

### Task 1: FileItem.Selected toggle and SelectedItems helper (commit: 049147d)

Added `Selected bool` field to `FileItem` struct. Updated `Title()` to prepend `SelectionIndicatorStyle.Render("[+]")` when selected. Added `ToggleSelect` key intercept in `FileListModel.Update()` before `m.list.Update(msg)`. Added `SelectedItems()` and `ClearSelections()` helper methods.

Files modified:
- `internal/ui/filelist.go` — Selected field, Title() badge, ToggleSelect handler, SelectedItems(), ClearSelections()
- `internal/ui/filelist_test.go` — TestFileItemToggleSelected, TestSelectedItems, TestClearSelections

### Task 2: AppModel wiring — 5 new states, all Phase 5 features (commit: e7847dd)

Wired all Phase 5 UI features into AppModel:

**New session states:** `stateHealth`, `stateRecipientForm`, `stateRecipientConfirm`, `stateRecipientList`, `stateBulkReKeyConfirm` (with exported test constants).

**New message types:** `HealthCheckResultMsg`, `ReKeyDoneMsg`, `RecipientDoneMsg`.

**New AppModel fields:** `health`, `recipientForm`, `bulkReKey *bulkReKeyState`, `recipientAction`, `recipientPubkey`, `recipientList`, `currentParsed`.

**Key handlers added:**
- `H` key in stateFileList: confirmation gate via stateDiff with `recipientAction="healthcheck"` sentinel, then async `runHealthCheck`
- `K` key in stateFileList: builds queue from SelectedItems, enters stateBulkReKeyConfirm per file
- `a` key in stateDetail: opens stateRecipientForm (add-recipient modal)
- `d` key in stateDetail: opens stateRecipientList (numbered remove list)

**Helper functions:** `runHealthCheck`, `flattenYAML`, `showBulkReKeyConfirm`, `advanceBulkReKey`, `renderRecipientList`.

**Esc chain:** All 5 new states handled correctly. `stateBulkReKeyConfirm` Esc skips current file and advances queue.

**View switch:** All 5 new states route to correct render function.

Files modified:
- `internal/app/model.go` — all above additions (~690 lines added)
- `internal/app/model_test.go` — 7 new tests covering health check, recipient form, bulk re-key, and Esc flows

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] statusBarHeight called with pointer receiver**
- **Found during:** Task 2 compilation
- **Issue:** `showBulkReKeyConfirm` is a pointer receiver `*AppModel` method but `statusBarHeight` takes `AppModel` by value
- **Fix:** Changed `statusBarHeight(m)` to `statusBarHeight(*m)` in the pointer receiver method
- **Files modified:** `internal/app/model.go`
- **Commit:** e7847dd (inline fix, same commit)

**2. [Rule 1 - Adaptation] Keys accessed via DefaultKeyMap rather than Keys() method**
- **Found during:** Task 2 planning — plan referenced `m.fileList.Keys().BulkReKey` but no exported `Keys()` method exists on FileListModel
- **Fix:** Used `keys.DefaultFileListKeyMap.BulkReKey` and `keys.DefaultDetailKeyMap.AddRecipient` — the same pattern already used throughout model.go for other global key handlers
- **Files modified:** `internal/app/model.go`
- **Commit:** e7847dd (design choice, not a bug)

**3. [Rule 1 - Simplification] Health check confirmation reuses stateDiff sentinel**
- **Found during:** Task 2 implementation — plan suggested multiple approaches for the H key confirmation gate
- **Fix:** Used `m.recipientAction = "healthcheck"` sentinel in the existing `stateDiff` handler rather than adding a 6th new state or an additional bool flag. The sentinel is cleared before dispatching the health scan.
- **Files modified:** `internal/app/model.go`

## Known Stubs

None. All Phase 5 features are wired end-to-end. Actual sops subprocess calls will fail in test environments without a real sops binary and age key, but the state machine wiring, async dispatch patterns, and UI rendering are all complete.

## Threat Surface Scan

No new network endpoints or trust boundaries beyond those documented in the plan's threat model. The `flattenYAML` function traverses `map[string]interface{}` from goccy/go-yaml — bounded by file size (T-05-12). The `recipientList` number-key bounds-check (T-05-08) is implemented at line ~754 in model.go.

## Test Results

```
ok  github.com/caesarakalaeii/sops-tui/internal/app  0.191s
ok  github.com/caesarakalaeii/sops-tui/internal/ui   0.287s
ok  github.com/caesarakalaeii/sops-tui/...           all packages green
```

## Self-Check: PASSED
