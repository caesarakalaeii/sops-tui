---
phase: 05-power-features
reviewed: 2026-04-16T00:00:00Z
depth: standard
files_reviewed: 16
files_reviewed_list:
  - internal/app/model.go
  - internal/app/model_test.go
  - internal/git/status.go
  - internal/git/status_test.go
  - internal/health/checker.go
  - internal/health/checker_test.go
  - internal/keys/bindings.go
  - internal/sops/executor.go
  - internal/sops/executor_test.go
  - internal/ui/filelist.go
  - internal/ui/filelist_test.go
  - internal/ui/health.go
  - internal/ui/health_test.go
  - internal/ui/recipientform.go
  - internal/ui/recipientform_test.go
  - internal/ui/styles.go
findings:
  critical: 1
  warning: 5
  info: 4
  total: 10
status: issues_found
---

# Phase 5: Code Review Report

**Reviewed:** 2026-04-16
**Depth:** standard
**Files Reviewed:** 16
**Status:** issues_found

## Summary

This phase introduces health checks, recipient management (add/remove age recipients), and bulk re-key — all significant security-adjacent features. The code is generally well-structured with good use of context timeouts, atomic file writes, and separation of concerns. The Bubble Tea model state machine is coherent and follows established patterns.

One critical issue was found: the `pubkey` value entered by the user in `RecipientFormModel` is passed directly to the `sops rotate --add-age` subprocess without sanitization. While `age.ParseX25519Recipient` validates bech32 encoding before the call, a crafted `pubkey` with shell metacharacters could potentially interfere with the exec argument list if `sops` itself has a parsing vulnerability, though the real and immediate risk is argument injection — passing a string like `age1xxx --some-flag` would append extra flags to the sops command. Five warnings cover logic gaps and defensive-programming shortcomings. Four informational items cover dead code, naming, and minor style concerns.

---

## Critical Issues

### CR-01: Age public key passed unsanitized to sops subprocess — argument injection risk

**File:** `internal/sops/executor.go:162-170`

**Issue:** `AddRecipient` and `RemoveRecipient` pass `pubkey` directly as a command argument. `age.ParseX25519Recipient` validates that the string is a syntactically valid bech32-encoded age key, but it does not prevent a caller from supplying a string that starts with a valid age key prefix followed by whitespace and extra flag text (e.g. a value containing newlines or shell-style flags). The concern here is not shell injection (Go `exec.Command` does not use a shell), but that a sufficiently malformed pubkey that slips through `age.ParseX25519Recipient`'s validation could be consumed as a distinct command argument by `sops`. More concretely: the `CharLimit = 200` cap means a user can type a 62-character bech32 key followed by 138 characters of arbitrary text. The age validator accepts the key portion and ignores trailing content if the parser stops at key end. Confirm whether `age.ParseX25519Recipient` validates the entire string or only the key prefix.

Additionally, the pubkey is also used in `RemoveRecipient` for an already-stored recipient pulled from `m.currentParsed.Metadata.AgeRecipients` (file-sourced data), so the path for removal is safer — but the add path accepts direct user input.

**Fix:** After `age.ParseX25519Recipient` succeeds, ensure the entire trimmed input is exactly the canonical string representation of the parsed key. This eliminates any trailing garbage:

```go
// In RecipientFormModel.Update, replace the enter case:
case "enter":
    rawInput := strings.TrimSpace(m.input.Value())
    recipient, err := age.ParseX25519Recipient(rawInput)
    if err != nil {
        m.errMsg = "Invalid age key: " + err.Error()
        return m, nil
    }
    // Re-serialize from the parsed recipient to get the canonical form.
    // This discards any trailing content that the parser may have ignored.
    canonical := recipient.String()
    if canonical != rawInput {
        m.errMsg = "Invalid age key: unexpected trailing characters"
        return m, nil
    }
    m.confirmed = true
    return m, nil
```

And ensure `AppModel` uses `m.recipientForm.Value()` after this normalization (it already does, but `Value()` should now return only the pre-validated canonical string stored by the form, not raw input). Alternatively, store the canonical form in a separate field and expose it via a `CanonicalValue()` accessor.

---

## Warnings

### WR-01: `showBulkReKeyConfirm` silently ignores `parser.ParseFile` error — empty recipient list

**File:** `internal/app/model.go:1787`

**Issue:** `showBulkReKeyConfirm` calls `parser.ParseFile(file.AbsPath, file.Rule, true)` and discards the error with `parsed, _ := ...`. If parsing fails (e.g., the file is unreadable or malformed), `entries` will be empty and the diff overlay will show no recipient rows. The user sees a confirmation dialog with no information about what they are confirming. More critically, after confirming an empty dialog, `sops rotate -i` is still executed on the file — which will succeed or fail silently, but the user had no accurate view of the operation.

**Fix:** Surface the parse error instead of silently continuing:

```go
parsed, err := parser.ParseFile(file.AbsPath, file.Rule, true)
if err != nil {
    m.status, _ = m.status.Flash("Re-key: could not read recipients: " + err.Error())
    m.advanceBulkReKey() // skip this file
    return
}
```

---

### WR-02: `ReEncryptDoneMsg` handler dispatches `GetFileStatuses` without checking `m.gitRepoRoot`

**File:** `internal/app/model.go:552-563`

**Issue:** When `ReEncryptDoneMsg` succeeds, the handler dispatches an async `gitpkg.GetFileStatuses` call guarded only by `m.sopsYamlPath != ""`. However, `m.gitRepoRoot` may be empty even when `sopsYamlPath` is non-empty (the user is not in a git repo). In this case `GetFileStatuses` receives an empty string as `repoRoot`, which then calls `gogit.PlainOpenWithOptions("", ...)`. This opens the git repo relative to the current working directory — potentially a different repo than expected, or returning `ErrRepositoryNotExists`.

Functionally the result is benign (the `GitStatusMsg` handler sets `GitAvailable: false` on error), but it does fire an unnecessary goroutine and may return spurious git status data from an unrelated repo.

**Fix:** Add a check for `m.gitRepoRoot`:

```go
if msg.Err == nil && m.sopsYamlPath != "" && m.gitRepoRoot != "" {
    // dispatch git status refresh
}
```

---

### WR-03: Health check stale-file detection flags files with `err != nil` from `GetLastCommitTime` as stale

**File:** `internal/app/model.go:1732-1738`

**Issue:** The else-if branch at line 1732 appends a `StaleFile` entry with `DaysSince: -1` whenever `GetLastCommitTime` returns a non-nil error. However, `GetLastCommitTime` returns an error for any repo-level failure (not just "no commits for this file"). A transient error — such as a locked pack file or a malformed object database — would incorrectly flag every file in the repo as stale (with DaysSince=-1), producing a health report full of false positives that cannot be distinguished from genuinely uncommitted files.

The comment says "No git history — flag with DaysSince=-1", but `GetLastCommitTime` returns errors for conditions that are not "no history" (e.g., if the repo itself fails to open).

**Fix:** Distinguish between "no commits for this file" (zero timestamp + nil error from `GetLastCommitTime` after `storer.ErrStop`) and actual repo errors. The current `GetLastCommitTime` implementation already returns `(time.Time{}, nil)` when no commits exist (via `storer.ErrStop` handling). Therefore the `err != nil` branch should not be treated as "no git history" — it should be an error log entry, not a stale file:

```go
commitTime, err := gitpkg.GetLastCommitTime(gitRepoRoot, relPath)
if err == nil {
    if !commitTime.IsZero() {
        daysSince := int(time.Since(commitTime).Hours() / 24)
        if daysSince > stalenessThreshold {
            result.StaleFiles = append(result.StaleFiles, health.StaleFile{
                FilePath: f.Name, LastCommitTime: commitTime, DaysSince: daysSince,
            })
        }
        // Zero commitTime with nil error means the file has never been committed — skip silently.
    }
} else {
    result.Errors = append(result.Errors, f.Name+": git error: "+err.Error())
}
```

---

### WR-04: `HealthModel.ScrollDown` max-scroll calculation uses `len(lines) - 1` — off-by-one when no content

**File:** `internal/ui/health.go:62-68`

**Issue:** `ScrollDown` computes `maxScroll = len(lines) - 1`. When the results are non-empty but `buildContentLines()` returns exactly one line, `maxScroll = 0` and scrolling is disabled entirely — correct. But when the model is in loading state, `buildContentLines()` returns an empty slice (`len = 0`), so `maxScroll = -1`, then it is clamped to 0. The clamping is safe, so this does not currently cause a panic. However, if `View()` is called in loading state and `m.scroll` was somehow set to a non-zero value (e.g., by SetResults race), the `allLines[m.scroll:]` slice expression at line 158 would panic with an out-of-bounds index because `m.loading` is false at that point but the check branches are `loading → else-if empty → else content`.

The real risk is: the loading check in `View()` at line 145 correctly short-circuits before accessing `allLines`, so there is no current panic path. But `SetResults` does not reset `m.scroll`, meaning if a user scrolls down during loading (before results arrive), `m.scroll` retains a stale positive value. When results arrive and `SetResults` is called, `View()` may slice into `allLines` at an index beyond its length if the new results have fewer lines than the scroll position.

**Fix:** Reset `scroll` in `SetResults`:

```go
func (m *HealthModel) SetResults(results health.HealthCheckResult) {
    m.results = results
    m.loading = false
    m.scroll = 0 // reset scroll position when new results arrive
}
```

---

### WR-05: `stateRecipientList` key handler limits recipient count to 9 without user feedback

**File:** `internal/app/model.go:747-770`

**Issue:** The recipient removal overlay only handles keys `'1'` through `'9'` (indices 0–8). If a file has 10 or more age recipients, the 10th and subsequent recipients are rendered in `renderRecipientList()` but are unreachable by keyboard (the selection prompt even says `1-N` where N can be > 9). The user sees a recipient row they cannot select, which is misleading and a functional gap.

**Fix:** Either cap the displayed list to 9 recipients with a note ("showing first 9 of N") or implement multi-character number input. The simpler approach:

```go
// In renderRecipientList(), cap display:
maxDisplay := 9
display := m.recipientList
if len(display) > maxDisplay {
    display = display[:maxDisplay]
}
// ... render display with a note if truncated
```

And document this limitation in the UX spec. Alternatively, use `j`/`k` navigation + Enter for selection to avoid the 9-item cap.

---

## Info

### IN-01: `SelectedItem` and `SelectedFileItem` in `FileListModel` are identical

**File:** `internal/ui/filelist.go:352-372`

**Issue:** `SelectedItem()` and `SelectedFileItem()` have identical bodies. One of them appears to be dead code — both are exported but differ only in name. `SelectedFileItem` is used by `AppModel`; `SelectedItem` is used in tests. Consider removing the duplicate or adding a concrete distinction (e.g., `SelectedItem` returns `list.Item` interface, `SelectedFileItem` returns the concrete `FileItem`).

**Fix:** Remove `SelectedItem` and update the single test call site to use `SelectedFileItem`, or document the intentional distinction.

---

### IN-02: `contains` helper in `recipientform_test.go` reimplements `strings.Contains`

**File:** `internal/ui/recipientform_test.go:116-126`

**Issue:** The `contains` function is a manual reimplementation of `strings.Contains` using a loop. It is functionally equivalent but harder to read and serves no purpose — `strings.Contains` is available in the standard library and handles all the same cases more clearly.

**Fix:** Replace all `contains(s, sub)` calls with `strings.Contains(s, sub)` and remove the helper.

---

### IN-03: `stripAnsi` helper referenced in `filelist_test.go` is not defined in the reviewed file set

**File:** `internal/ui/filelist_test.go:87`

**Issue:** `stripAnsi(title)` is called in `TestFileItemTitleEncrypted` and `TestFileItemTitleUnencrypted`, but the function is not visible in `filelist_test.go`. It must be defined elsewhere in the `ui_test` package. This is not a bug but a scoping note: the function should be checked for correctness (ANSI stripping is needed to reliably test lipgloss output) and placed in a shared test helper file with a clear name.

**Fix:** Confirm `stripAnsi` is defined in a shared test helper within the `ui_test` package (e.g., `helpers_test.go`). No code change needed if it is; document the file location in comments if non-obvious.

---

### IN-04: `HealthSectionHeaderStyle` defined in `styles.go` but `HelpSectionHeader` used in `health.go`

**File:** `internal/ui/styles.go:201` and `internal/ui/health.go:87`

**Issue:** `styles.go` defines `HealthSectionHeaderStyle` (line 201) for section headers in the health overlay, but `health.go` uses `HelpSectionHeader` (the help overlay header style) for its section headers instead. This means health section titles render with the help overlay typography rather than the dedicated health style. This is a minor visual inconsistency; `HelpSectionHeader` is `Bold(true)` and `HealthSectionHeaderStyle` is `Bold(true).Foreground(ColorFg)`, so the only difference is the explicit `ColorFg` foreground. In practice the output may look the same in most terminals, but the dedicated style exists for a reason.

**Fix:** In `health.go`, replace `HelpSectionHeader.Render(...)` calls with `HealthSectionHeaderStyle.Render(...)` to use the intended style.

---

_Reviewed: 2026-04-16_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
