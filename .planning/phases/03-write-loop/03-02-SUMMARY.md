---
phase: 03-write-loop
plan: "02"
subsystem: ui-write-loop
tags: [diff-overlay, inline-edit, state-machine, sops-set, tdd]
dependency_graph:
  requires: ["03-01"]
  provides: ["diff-overlay", "inline-edit", "stateDiff-wiring"]
  affects: ["internal/ui/diff.go", "internal/ui/detail.go", "internal/app/model.go"]
tech_stack:
  added: []
  patterns:
    - "DiffModel overlay mirroring MetadataModel pattern (bordered box, scroll, View)"
    - "editActive inline textinput embedded in tree row — eats navigation keys"
    - "stateDiff key routing before global keys — y/n captured by overlay"
    - "T-03-08: editOldVal cleared at both Enter and Esc paths"
key_files:
  created:
    - internal/ui/diff.go
    - internal/ui/diff_test.go
  modified:
    - internal/ui/detail.go
    - internal/ui/detail_test.go
    - internal/app/model.go
    - internal/app/model_test.go
decisions:
  - "editActive branch in DetailModel.Update runs before searchActive branch — ensures textinput eats all keys including j/k (Pitfall 4)"
  - "stateDiff key routing placed before global key handler in AppModel so y/n/Esc are captured before ? and q"
  - "editOldVal and editKeyPath cleared after capture in Enter/Esc paths (T-03-08 mitigation)"
  - "Array-indexed key path blocking returns EditBlockedMsg with non-empty Reason; AppModel flashes the reason string"
  - "Esc in stateDiff handled in both stateDiff routing block and the Esc priority chain (belt and suspenders)"
metrics:
  duration_minutes: 18
  completed_date: "2026-04-14T21:12:50Z"
  tasks_completed: 2
  files_changed: 6
requirements: [EDT-01, EDT-02, EDT-04]
---

# Phase 3 Plan 02: Diff Confirmation Overlay and Inline Single-Key Editing — Summary

Diff overlay and inline edit flow delivering EDT-01 (edit a secret value with re-encryption), EDT-02 (diff view before confirming), and EDT-04 (confirmation gate for all destructive writes).

## What Was Built

### Task 1: DiffModel overlay and inline edit on DetailModel (aa05ff1)

**`internal/ui/diff.go`** — new file:
- `DiffEntry` struct: `KeyPath`, `OldValue`, `NewValue`
- `DiffModel` struct: full-screen overlay with title, entries, scroll, confirmed/cancelled flags
- `NewDiffModel(title, entries, width, height)` constructor
- `Update()` handles y → confirmed, n/Esc → cancelled, j/k → scroll
- `View()` renders bordered box (RoundedBorder, ColorSurface BG, ColorMuted border) with title in DiffKeyStyle, removed lines in DiffRemovedStyle (`- old`), added lines in DiffAddedStyle (`+ new`), footer with `[y] confirm re-encrypt   [n/Esc] cancel` in ConfirmPromptStyle
- Multi-entry layout: each entry gets a key path header with blank line separator

**`internal/ui/detail.go`** — extended:
- Added `EditConfirmMsg`, `EditCancelMsg`, `EditBlockedMsg` message types
- Added `editActive`, `editInput`, `editKeyPath`, `editOldVal` fields to `DetailModel`
- `IsEditActive() bool` method
- `e` key handler: on revealed encrypted leaf → activate edit; on masked leaf → EditBlockedMsg{}; on array-indexed key → EditBlockedMsg{Reason: "Array-indexed keys not editable in Phase 3"}
- `editActive` branch in `Update()` intercepts all messages including j/k before search and navigation
- Enter in editActive → `EditConfirmMsg` with old/new values; Esc → `EditCancelMsg`
- `renderRowKeyOnly()` helper renders connector + key name + ": " for inline edit row
- View renders textinput inline at cursor row when `editActive`

**Tests:** 6 diff tests + 6 edit tests — all pass.

### Task 2: AppModel wiring for stateDiff, stateEdit, SetKey re-encryption (9d91d03)

**`internal/app/model.go`** — extended:
- `diff ui.DiffModel` and `editFilePath string` fields on `AppModel`
- `WindowSizeMsg` propagates to `m.diff.SetSize()`
- `case ui.EditConfirmMsg`: if old == new → flash "No changes"; else build DiffModel with single entry, set `editFilePath`, transition to `stateDiff`
- `case ui.EditBlockedMsg`: flash Reason or "Reveal first with r"
- `case ui.EditCancelMsg`: no-op
- `case ReEncryptDoneMsg`: flash "Re-encrypted" + update revealed node value on success; flash "Re-encryption failed: ..." on error; always transition to `stateDetail`
- stateDiff key routing before global keys: routes to `m.diff.Update()`; on Confirmed() → dispatch `sops.SetKey` as `tea.Cmd`; on Cancelled() → return to prevState + flash "Cancelled"
- stateDiff added to Esc priority chain (before stateHelp)
- `case stateDiff:` added to View switch

**Tests:** 9 new model tests — all pass.

## Verification

```
go build ./...                          ✓ no errors
go test ./internal/ui/... -v            ✓ all pass
go test ./internal/app/... -v           ✓ all pass
go test ./...                           ✓ all pass (6 packages)
```

## Deviations from Plan

None — plan executed exactly as written.

## Security Notes (Threat Model Coverage)

| Threat | Mitigation | Verified |
|--------|-----------|---------|
| T-03-06 Tampering (EditConfirmMsg.NewValue → sops stdin) | json.Marshal in sops.SetKey (Plan 01) | From Plan 01 |
| T-03-08 Info Disclosure (editOldVal in DetailModel) | editOldVal cleared at Enter and Esc paths | ✓ lines 216, 228 detail.go |
| T-03-09 DoS (textinput CharLimit) | textinput.CharLimit = 1000 | ✓ detail.go line 398 |
| T-03-07 Repudiation | Accepted (no audit log in v1) | — |

## Self-Check: PASSED
