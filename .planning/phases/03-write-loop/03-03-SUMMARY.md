---
phase: 03-write-loop
plan: "03"
subsystem: write-loop
tags: [editor-flow, rotation, diff, tdd, security]
dependency_graph:
  requires: ["03-02"]
  provides: ["editor-flow", "format-rotation", "format-menu", "multi-key-diff"]
  affects: ["internal/app/model.go", "internal/ui/detail.go", "internal/ui/rotate.go"]
tech_stack:
  added: ["tea.ExecProcess", "os/exec", "crypto/rand", "golang.org/x/crypto/bcrypt", "sort"]
  patterns: ["TDD RED-GREEN", "ExecProcess subprocess suspension", "YAML-aware diff", "format menu state machine"]
key_files:
  created:
    - internal/ui/rotate.go
    - internal/ui/rotate_test.go
  modified:
    - internal/app/model.go
    - internal/app/model_test.go
    - internal/ui/detail.go
    - internal/ui/detail_test.go
decisions:
  - "EditorFinishedMsg carries TmpPath + OriginalContent for deferred diff comparison (not inline content)"
  - "YAML-aware compareDecryptedYAML ignores key ordering (Pitfall 6 mitigation)"
  - "stateFormatMenu added as new sessionState rather than modal overlay to avoid focus complexity"
  - "Test base64 values use 22+ char threshold matching DetectFormat regex (not arbitrary short strings)"
metrics:
  duration_seconds: 449
  completed_date: "2026-04-14"
  tasks_completed: 2
  files_changed: 6
---

# Phase 3 Plan 03: $EDITOR Flow, Multi-Key Diff, and Format-Aware Rotation — Summary

**One-liner:** $EDITOR full-file editing via tea.ExecProcess with YAML-aware diff, plus format-aware rotation (base64/hex/UUID/bcrypt/alphanumeric) all funneling through the same stateDiff confirmation gate.

## What Was Built

### Task 1: $EDITOR flow with tea.ExecProcess and multi-key diff

**`internal/ui/detail.go`**
- Added `EditorRequestMsg`, `RotateFormatMenuMsg`, `RotateReadyMsg`, `RotateErrorMsg` message types
- Added `E` key handling in `DetailModel.Update()`: returns `EditorRequestMsg` when any node is revealed, `EditBlockedMsg` otherwise
- Added `X` key handling for rotation: auto-detects format with `DetectFormat`, returns `RotateReadyMsg` (known format) or `RotateFormatMenuMsg` (unknown), guards array-indexed keys

**`internal/app/model.go`**
- Added `EditorReadyMsg`, `EditorFinishedMsg` message types
- Added `editorEditedContent []byte`, `formatMenuActive/KeyPath/OldValue/Cursor`, `rotateFormat` fields to `AppModel`
- Added `stateFormatMenu` to `sessionState` enum
- Implemented `launchEditor(decryptedContent []byte) tea.Cmd`: writes 0600 temp file, detects `$EDITOR/$VISUAL/vi`, suspends TUI via `tea.ExecProcess`
- Implemented `compareDecryptedYAML(original, edited []byte) ([]ui.DiffEntry, error)`: YAML-aware key-order-independent diff via `flattenYAMLToMap`/`walkMapSlice`
- `EditorFinishedMsg` handler: reads temp file, calls `compareDecryptedYAML`, shows multi-key diff ("Changes: N keys modified") or flashes "No changes detected"
- Updated `stateDiff` confirm handler: single-entry without `editorEditedContent` uses `sops.SetKey`; multi-entry or with `editorEditedContent` writes to second temp file and calls `sops.EncryptFile`
- `editorEditedContent` set to nil immediately after use (T-03-12 mitigation)

### Task 2: Format-aware secret rotation with X key

**`internal/ui/rotate.go`** (new file)
- `SecretFormat` enum: `FormatUnknown`, `FormatBase64`, `FormatHex`, `FormatUUID`, `FormatBcrypt`, `FormatAlphanumeric`
- `DetectFormat(value string) SecretFormat`: bcrypt-first detection order (D-13: prevents false base64 match)
- `GenerateValue(format SecretFormat) (string, error)`: all generation uses `crypto/rand` (T-03-13); bcrypt uses cost 12 (T-03-14)
- `AllFormats() []SecretFormat`: 5 selectable formats for format menu
- `FormatLabel(f SecretFormat) string`: human-readable format names for flash messages

**`internal/app/model.go`** (rotation wiring)
- `RotateReadyMsg` handler: transitions directly to `stateDiff` with single-entry diff
- `RotateFormatMenuMsg` handler: transitions to `stateFormatMenu`
- `stateFormatMenu` key routing: j/k navigate, Enter generates value and moves to `stateDiff`, Esc cancels
- `renderFormatMenu()`: rounded-border overlay with 5 format options
- `ReEncryptDoneMsg` handler updated: flashes "Rotated to {format}" when `rotateFormat != 0`

## Threat Mitigations Applied

| Threat | Mitigation | Status |
|--------|------------|--------|
| T-03-10: temp file disclosure | `Chmod(0600)` on creation; `os.Remove` immediately after read in handler | Applied |
| T-03-11: $EDITOR subprocess | tea.ExecProcess suspends TUI cleanly; editor is user-trusted | Accepted |
| T-03-12: editorEditedContent in memory | Field set to nil after confirm/cancel | Applied |
| T-03-13: crypto/rand for rotation | All `GenerateValue` paths use `crypto/rand`, never `math/rand` | Applied |
| T-03-14: bcrypt DoS | Cost 12 (~250ms) is single-shot, acceptable in interactive TUI | Accepted |
| T-03-15: format detection on decrypted values | Values already revealed in memory — no new disclosure | Accepted |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed base64 detection test values too short**
- **Found during:** Task 2 RED phase execution
- **Issue:** Test used `"dGVzdCBiYXNlNjQ="` (16 chars) but DetectFormat regex requires 22+ chars. Test expected `FormatBase64` but got `FormatUnknown`.
- **Fix:** Updated test to use 24-char base64 (18 decoded bytes) which correctly exceeds the 22-char threshold. Also updated `TestRotateKeyOnRevealed` to use a 44-char value (`strings.Repeat("A", 44)`).
- **Files modified:** `internal/ui/rotate_test.go`, `internal/ui/detail_test.go`

**2. [Rule 1 - Bug] Fixed duplicate View() variable declarations**
- **Found during:** Task 1 implementation
- **Issue:** After inserting the `stateFormatMenu` case into `View()`, `statusBar` and `mainH` were declared twice in the same function scope.
- **Fix:** Removed the duplicate declarations in the lower half of `View()`, reusing the variables computed at the top of the function.
- **Files modified:** `internal/app/model.go`

**3. [Rule 3 - Blocking] Added `EditorReadyMsg` as intermediate step**
- **Found during:** Task 1 implementation
- **Issue:** The plan's `EditorRequestMsg` handler needed to dispatch both a decrypt subprocess AND then `launchEditor`. The handler must return one `tea.Cmd`, so an intermediate `EditorReadyMsg` was needed to bridge the two async steps.
- **Fix:** Added `EditorReadyMsg` as internal message; `EditorRequestMsg` → decrypt → `EditorReadyMsg` → `launchEditor` → `EditorFinishedMsg`.
- **Files modified:** `internal/app/model.go`

## Known Stubs

None. All rotation and editor flows are wired end-to-end. The `sops.EncryptFile` call is real but will fail without a real SOPS-encrypted file and age key (expected — test coverage uses unit tests, not integration tests).

## Threat Flags

None. No new network endpoints, auth paths, or schema changes introduced. The temp file pattern was pre-planned in the threat model.

## Self-Check: PASSED

All files found. All commits verified.

| Item | Status |
|------|--------|
| internal/ui/rotate.go | FOUND |
| internal/ui/rotate_test.go | FOUND |
| internal/app/model.go | FOUND |
| internal/ui/detail.go | FOUND |
| 03-03-SUMMARY.md | FOUND |
| feat commit b93cb50 | FOUND |
| test commit bfabb40 | FOUND |
