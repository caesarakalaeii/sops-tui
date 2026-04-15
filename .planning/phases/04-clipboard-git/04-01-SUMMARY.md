---
phase: 04-clipboard-git
plan: "01"
subsystem: clipboard
tags: [clipboard, signal-handling, styles, keybindings, statusbar]
dependency_graph:
  requires: []
  provides: [clipboard-copy, auto-clear, signal-cleanup, phase4-styles, phase4-keybindings]
  affects: [internal/app/model.go, internal/ui/statusbar.go, internal/ui/styles.go, internal/keys/bindings.go, cmd/sops-tui/main.go]
tech_stack:
  added: [github.com/atotto/clipboard v0.1.4 (direct)]
  patterns: [tea.Tick generation counter, signal.NotifyContext, TDD red-green]
key_files:
  created: [internal/app/model_clipboard_test.go]
  modified:
    - internal/ui/styles.go
    - internal/keys/bindings.go
    - internal/ui/statusbar.go
    - internal/app/model.go
    - internal/ui/detail.go
    - cmd/sops-tui/main.go
    - go.mod
decisions:
  - "[clip] indicator only visible in normal (non-flash) status bar mode — tests use IsClipboardHot() for state assertion and FlashClearMsg to verify indicator in view"
  - "SelectedNode() added to DetailModel (missing from codebase, required by copy handler)"
  - "IsClipboardHot() exported on AppModel for test access without inspecting ANSI-escaped view strings"
  - "ClipboardTimeout exported as var (not func) to allow test override if needed"
metrics:
  duration_minutes: 35
  completed: "2026-04-15T08:04:07Z"
  tasks_completed: 2
  files_modified: 7
---

# Phase 4 Plan 1: Clipboard Copy, Auto-Clear, and Signal-Safe Cleanup Summary

Clipboard copy with 30s auto-clear, generation-counter stale-clear prevention, SIGINT/SIGTERM signal handler, and all Phase 4 style/keybinding contracts added for use by Plans 02 and 03.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Phase 4 style vars, keybindings, status bar extensions | 0097a66 | styles.go, bindings.go, statusbar.go |
| 2 (RED) | Failing clipboard tests | b472f0b | model_clipboard_test.go, detail.go |
| 2 (GREEN) | Clipboard copy, auto-clear, signal-safe cleanup | 3bab68f | model.go, main.go, go.mod, model_clipboard_test.go |

## What Was Built

### Task 1: Phase 4 Foundation (styles, bindings, statusbar)

**styles.go** — 9 new style vars added after `FormatMenuStyle`:
- `BadgeModified`, `BadgeAdded`, `BadgeUntracked` — git badge styles for Plans 02/03
- `ClipboardHotStyle` — accent-colored `[clip]` indicator
- `GitNoRepoStyle` — muted "no git" indicator
- `HistoryHashStyle`, `HistoryDateStyle`, `HistoryAuthorStyle`, `HistorySubjectStyle` — history overlay styles

**bindings.go** — `Copy` (ctrl+y) and `Blame` (b) added to `DetailKeyMap` struct, `DefaultDetailKeyMap`, `ShortHelp()`, and `FullHelp()`.

**statusbar.go** — `GitAvailable bool` added to `EnvStatus`. `clipboardHot bool` field, `SetClipboardHot()`/`IsClipboardHot()` methods added to `StatusBarModel`. View renders `[clip]` before env indicators when hot. `renderEnvIndicators` renders "no git" when `GitAvailable` is false.

### Task 2: Clipboard Implementation (TDD)

**model.go** — Full clipboard implementation:
- `ClipboardClearMsg{Gen int}` — message type for auto-clear timer
- `clipboardGen int`, `clipboardHot bool` — fields on `AppModel`
- `clipboardTimeout()` — reads `SOPS_TUI_CLIPBOARD_TIMEOUT`, falls back to 30s
- `ClipboardTimeout` — exported var for test access
- `IsClipboardHot()` — exported accessor for test assertions
- `copyToClipboard(value string)` — guards `clipboard.Unsupported`, writes, increments gen, sets hot, flashes, schedules `ClipboardClearMsg` via `tea.Tick`
- `ClipboardClearMsg` handler in `Update()` — generation check prevents stale clears
- `ctrl+y` handler in `tea.KeyPressMsg` — only in `stateDetail`, checks `Revealed`, flashes "Reveal first with r" on masked leaf

**detail.go** — `SelectedNode() (TreeNode, bool)` added to `DetailModel` (was referenced in plan but missing from codebase).

**main.go** — Signal-safe cleanup:
- `defer clipboard.WriteAll("")` — clears on normal exit
- `signal.NotifyContext` goroutine for SIGINT/SIGTERM — calls `clipboard.WriteAll("")` then `os.Exit(0)` without interacting with bubbletea program

**go.mod** — `github.com/atotto/clipboard v0.1.4` promoted from indirect to direct.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing] SelectedNode() not present in DetailModel**
- **Found during:** Task 2 implementation
- **Issue:** Plan spec referenced `m.detail.SelectedNode()` but `DetailModel` only had `SelectedIndex()`. No method existed to get the current node by cursor.
- **Fix:** Added `SelectedNode() (TreeNode, bool)` to `DetailModel` in `detail.go`. Returns the `TreeNode` at `flatRows[cursor]` with bounds check.
- **Files modified:** `internal/ui/detail.go`
- **Commit:** b472f0b

**2. [Rule 1 - Bug] Test assertions needed to account for flash mode hiding [clip]**
- **Found during:** Task 2 TDD GREEN phase
- **Issue:** Tests checking for `[clip]` in `View().Content` failed because after `ctrl+y`, the status bar enters flash mode (showing "Copied (clears in 30s)"), which replaces all normal content including `[clip]`. The `[clip]` indicator is only visible when not in flash mode.
- **Fix:** Added `IsClipboardHot()` accessor on `AppModel` for direct state assertions. Tests for indicator visibility send `ui.FlashClearMsg{Gen: 1}` to exit flash mode first, then check view content.
- **Files modified:** `internal/app/model.go`, `internal/app/model_clipboard_test.go`
- **Commit:** 3bab68f

## Threat Mitigations Applied

| Threat | Mitigation Applied |
|--------|-------------------|
| T-04-01: Clipboard information disclosure | 30s auto-clear via `tea.Tick` + generation counter; `defer clipboard.WriteAll("")` + signal handler |
| T-04-03: DoS via invalid SOPS_TUI_CLIPBOARD_TIMEOUT | `strconv.Atoi` with `n > 0` check; falls back to 30s |
| T-04-04: clipboard.Unsupported | Guard in `copyToClipboard` before `WriteAll`; flashes user-visible message |
| T-04-05: Stale ClipboardClearMsg tampering | Generation counter — `msg.Gen == m.clipboardGen` check in handler |

T-04-02 (SIGKILL) accepted per threat model — cannot be intercepted.

## Known Stubs

None. All clipboard functionality is fully wired end-to-end.

## Threat Flags

None. No new network endpoints, auth paths, file access patterns, or schema changes introduced beyond what the threat model covers.

## Self-Check

Files exist:
- `internal/ui/styles.go` — modified (BadgeModified present)
- `internal/keys/bindings.go` — modified (Copy/Blame bindings present)
- `internal/ui/statusbar.go` — modified (clipboardHot, GitAvailable present)
- `internal/app/model.go` — modified (ClipboardClearMsg, clipboardTimeout present)
- `internal/ui/detail.go` — modified (SelectedNode present)
- `cmd/sops-tui/main.go` — modified (signal.NotifyContext present)
- `internal/app/model_clipboard_test.go` — created (9 tests)

Commits exist: 0097a66, b472f0b, 3bab68f

## Self-Check: PASSED
