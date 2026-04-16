---
phase: 04-clipboard-git
reviewed: 2026-04-15T00:00:00Z
depth: standard
files_reviewed: 15
files_reviewed_list:
  - cmd/sops-tui/main.go
  - go.mod
  - internal/app/model.go
  - internal/git/status.go
  - internal/git/status_test.go
  - internal/keys/bindings.go
  - internal/sops/discoverer.go
  - internal/ui/detail.go
  - internal/ui/detail_reveal_test.go
  - internal/ui/detail_test.go
  - internal/ui/filelist.go
  - internal/ui/history.go
  - internal/ui/history_test.go
  - internal/ui/statusbar.go
  - internal/ui/styles.go
findings:
  critical: 0
  warning: 5
  info: 4
  total: 9
status: issues_found
---

# Phase 4: Code Review Report

**Reviewed:** 2026-04-15
**Depth:** standard
**Files Reviewed:** 15
**Status:** issues_found

## Summary

Phase 4 adds clipboard copy (ctrl+y), git status badges on the file list, and a full-screen git history overlay. The overall quality is high: the generation-counter pattern is used correctly for both clipboard and flash messages, the clipboard clear-on-exit belt-and-suspenders approach in `main.go` is solid, and the go-git integration in `status.go` correctly avoids the auto-insertion bug in `status.File()`. No critical (security or data-loss) issues were found.

Five warnings were identified: a goroutine leak when the signal handler fires, a lost `tea.Cmd` in the `ReEncryptDoneMsg` branch when the git-status refresh runs, a scroll upper-bound bug in `HistoryModel.ScrollDown`, a missing `tea.Quit` call in the goroutine-based signal handler, and an environment-status string-matching fragility in `main.go`. Four info items are also noted.

---

## Warnings

### WR-01: Goroutine in signal handler calls `os.Exit(0)` without draining Bubble Tea

**File:** `cmd/sops-tui/main.go:55-59`
**Issue:** The goroutine started to handle SIGINT/SIGTERM calls `clipboard.WriteAll("")` and then `os.Exit(0)` directly. This races with the Bubble Tea main loop: it bypasses `tea.Quit`, leaving the terminal in raw mode (no cleanup from the Bubble Tea renderer) and skipping any `defer` statements in the program.  
The `defer clipboard.WriteAll("")` on line 50 and the `defer stop()` on line 54 both run only when `main()` returns normally; the goroutine's `os.Exit(0)` skips them. The clipboard write in the goroutine is therefore the only thing that fires, but the terminal state is not restored.

**Fix:** Send a `tea.Quit` message through the program instead of calling `os.Exit` directly, or send to a quit channel that main listens to after `p.Run()` returns:

```go
go func() {
    <-ctx.Done()
    clipboard.WriteAll("") //nolint:errcheck
    p.Quit()               // signals Bubble Tea to quit cleanly; main() returns and defers run
}()
```

### WR-02: Lost `tea.Cmd` from git-status refresh in `ReEncryptDoneMsg` handler

**File:** `internal/app/model.go:494-505`
**Issue:** When re-encryption succeeds and `m.sopsYamlPath != ""`, the handler builds a `gitCmd` and returns it:

```go
return m, func() tea.Msg {
    statuses, err := gitpkg.GetFileStatuses(sopsDir, relPaths)
    return GitStatusMsg{Statuses: statuses, GitAvailable: true, Err: err}
}
```

However, just before the `if msg.Err == nil` block (lines 471-491), the code calls `m.status.Flash(...)` which returns `(StatusBarModel, tea.Cmd)`, but the `tea.Cmd` (the flash-clear tick) is discarded with `_`:

```go
m.status, _ = m.status.Flash("Re-encrypted")
```

This is consistent with every other Flash call in the file — they all discard the tick cmd. The flash will therefore never be automatically cleared: the `FlashClearMsg` tick is never scheduled. This is a pre-existing pattern but the `ReEncryptDoneMsg` branch compounds it by then returning the git-status cmd, meaning the discarded flash tick is never recovered.

**Fix:** Return both commands via `tea.Batch`:

```go
m.status, flashCmd = m.status.Flash("Re-encrypted")
// ...
return m, tea.Batch(flashCmd, gitCmd)
```

Note: this Flash-cmd discard pattern appears throughout the entire file and is a systemic issue that would prevent all flash messages from auto-clearing. All `m.status, _ = m.status.Flash(...)` calls should capture and return the tick cmd.

### WR-03: `HistoryModel.ScrollDown` uses `len(entries) - 1` as upper bound — scrolling stops one row too early

**File:** `internal/ui/history.go:58-65`
**Issue:** `ScrollDown` clamps to `maxScroll = len(m.entries) - 1`. The `View` renders `lines[m.scroll:]`, so when `scroll == len(entries) - 1` (last entry), only the final entry is visible. With a terminal height that can show more than one row, the user can never scroll the first entry out of view — the last entry is reachable only when there is just one entry. For typical history lists (10-50 entries), a user can scroll to show only the last entry in the first position, which is fine. However, if the intent is to show all entries and allow scrolling the list until the last entry reaches the top, the bound should be `len(entries) - visibleHeight`. The current bound is at least safe (no panic), but may surprise users expecting to scroll further.

More importantly, in `View` the scroll guard is:
```go
if m.scroll > 0 && m.scroll < len(lines) {
    visibleLines = lines[m.scroll:]
}
```
When `m.scroll == len(lines) - 1` (last line) this slice is valid but shows only one line. When `m.scroll == len(lines)` (which can't happen due to the clamp) it would panic. The clamp prevents the panic, but the upper-bound condition in View should be `m.scroll <= len(lines)` (i.e., `<=` not `<`) to be defensive. If `SetEntries` is called after `ScrollDown` with a shorter slice, the stale `m.scroll` could temporarily equal `len(lines)` and the `m.scroll < len(lines)` check would skip slicing correctly, showing the full (new shorter) list from position 0 silently — a subtle reset.

**Fix:** After `SetEntries`, reset scroll to 0 to avoid stale scroll positions, and add a defensive clamp in `View`:

```go
func (m *HistoryModel) SetEntries(entries []gitpkg.CommitEntry) {
    m.entries = entries
    m.loading = false
    m.scroll = 0  // reset scroll when content changes
}
```

### WR-04: `hasResultWithMessage` in `main.go` uses substring match against display strings — fragile coupling

**File:** `cmd/sops-tui/main.go:63-66`
**Issue:** The `EnvStatus` is built by matching against human-readable validator message strings:
```go
SopsAvailable:     !hasResultWithMessage(results, "sops binary not found"),
AgeAvailable:      !hasResultWithMessage(results, "Age key file not found"),
SopsYamlAvailable: !hasResultWithMessage(results, ".sops.yaml not found"),
```
If any validator message wording changes (e.g., a typo fix, an i18n update), `EnvStatus` will silently compute the wrong values. The validator package presumably has structured result types; this coupling should be via a typed field (e.g., `result.CheckName` or `result.Code`), not a substring of the display message.

**Fix:** Add a `Code` or `Check` field to `validator.ValidationResult` and match against that. For example:
```go
SopsAvailable: !hasResultWithCode(results, validator.CheckSopsBinary),
```
This is a correctness risk: a user who has `sops` but gets a validator message that happens to contain "sops binary not found" as a substring of a longer message would incorrectly see the indicator as unavailable.

### WR-05: `HistoryModel.View` — scroll upper-bound check uses strict `<` instead of `<=`, but `m.scroll` can equal `len(lines)` if `SetEntries` is called with fewer entries after scrolling

**File:** `internal/ui/history.go:113-117`
**Issue:** (Related to WR-03.) The view scroll guard is:
```go
if m.scroll > 0 && m.scroll < len(lines) {
    visibleLines = lines[m.scroll:]
}
```
If `SetEntries` is called with a new (shorter) list while `m.scroll` is at a high value, and `m.scroll >= len(lines)` for the new list, the slice expression `lines[m.scroll:]` is skipped (because `m.scroll < len(lines)` is false), and `visibleLines = lines` is used — showing from position 0. This is a silent reset rather than a panic, so it is safe in practice. However, `ScrollDown` clamps to `len(m.entries) - 1` (line 62), not `len(lines) - 1`. Since `lines` is rebuilt from `m.entries` each render, these lengths are equal, so in practice no divergence occurs. The concern is a potential maintenance trap: if lines are built differently in the future (e.g., header lines added), the clamp in ScrollDown and the guard in View may diverge.

**Fix:** Add a defensive clamp at the top of `View` before slicing:
```go
if m.scroll >= len(lines) {
    m.scroll = len(lines) - 1
    if m.scroll < 0 {
        m.scroll = 0
    }
}
```

---

## Info

### IN-01: `SelectedItem` and `SelectedFileItem` in `filelist.go` are duplicates

**File:** `internal/ui/filelist.go:331-349`
**Issue:** `SelectedItem` (line 331) and `SelectedFileItem` (line 341) have identical implementations. Both type-assert `list.SelectedItem()` to `FileItem` and return the same result. This duplication will silently diverge if one is updated without the other.

**Fix:** Remove `SelectedItem` and rename all callers to use `SelectedFileItem`, or have `SelectedItem` delegate to `SelectedFileItem`.

### IN-02: Flash tick command is discarded throughout `model.go`

**File:** `internal/app/model.go` (multiple locations — e.g., lines 276, 315, 365, 369, 374, 378, etc.)
**Issue:** Every `m.status.Flash(...)` call in `AppModel.Update` discards the returned `tea.Cmd`:
```go
m.status, _ = m.status.Flash("...")
```
The `Flash` method returns a `tea.Tick` command that fires `FlashClearMsg` after 2 seconds. Discarding it means flash messages never auto-clear via the timer. The `StatusBarModel.Update` handles `FlashClearMsg` correctly, but the message is never sent because the tick command is dropped.

This is a pervasive pattern affecting all flash messages in the app. Either the design relies on some other mechanism to clear flashes, or this is a bug that has survived because flash messages are visible for the session lifetime without bothering anyone in tests.

**Fix:** Capture and return flash cmds:
```go
var flashCmd tea.Cmd
m.status, flashCmd = m.status.Flash("Decrypted")
return m, flashCmd
```
Where multiple commands exist, use `tea.Batch`.

### IN-03: `go-git` version pinned at `v5.17.0` rather than `v5.17.2` (security fix)

**File:** `go.mod:12`
**Issue:** CLAUDE.md recommends `go-git v5.17.2` (March 2026, with security fixes). The module is currently pinned to `v5.17.0`. Given CLAUDE.md's note about security-sensitive histories, staying on the latest patch version is advised.

**Fix:**
```
github.com/go-git/go-git/v5 v5.17.2
```

### IN-04: `stripAnsi` in `detail_test.go` does not handle all ANSI escape sequences

**File:** `internal/ui/detail_test.go:246-263`
**Issue:** The `stripAnsi` helper only strips sequences ending in `m`, `K`, `H`, or `J`. Lipgloss v2 may emit other CSI sequences (e.g., cursor movement `A`-`F`, `S`, `T`, `n`, `r`). If lipgloss emits such sequences, ANSI escape codes will bleed into the stripped string, causing false assertion failures in CI or on certain terminals. This doesn't affect correctness of the source code, but can make tests flaky.

**Fix:** Use a proper regex-based stripper or the `github.com/acarl005/stripansi` package, or expand the terminator set:
```go
if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
    inEsc = false
}
```

---

_Reviewed: 2026-04-15_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
