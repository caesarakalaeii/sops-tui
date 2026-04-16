---
phase: 05-power-features
fixed_at: 2026-04-16T00:00:00Z
review_path: .planning/phases/05-power-features/05-REVIEW.md
iteration: 1
findings_in_scope: 6
fixed: 6
skipped: 0
status: all_fixed
---

# Phase 5: Code Review Fix Report

**Fixed at:** 2026-04-16
**Source review:** .planning/phases/05-power-features/05-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 6
- Fixed: 6
- Skipped: 0

## Fixed Issues

### CR-01: Age public key passed unsanitized to sops subprocess — argument injection risk

**Files modified:** `internal/ui/recipientform.go`
**Commit:** af51e9a
**Applied fix:** Added `strings` import. In `Update()`, replaced the bare `age.ParseX25519Recipient(m.input.Value())` call with a trimmed-input flow: trim whitespace, parse, re-serialize via `recipient.String()`, and reject if the canonical form does not exactly match the trimmed input. This discards any trailing content the age parser may have silently ignored, preventing argument injection into the sops subprocess.

---

### WR-01: `showBulkReKeyConfirm` silently ignores `parser.ParseFile` error — empty recipient list

**Files modified:** `internal/app/model.go`
**Commit:** e5c8d7f
**Applied fix:** Replaced `parsed, _ := parser.ParseFile(...)` with a proper `parsed, err :=` check. On error, flash the error message, increment `bulkReKey.skipped`, call `advanceBulkReKey()` to skip to the next file, and return early — preventing sops rotate from running against a file whose recipients could not be read.

---

### WR-02: `ReEncryptDoneMsg` handler dispatches `GetFileStatuses` without checking `m.gitRepoRoot`

**Files modified:** `internal/app/model.go`
**Commit:** 6572de2
**Applied fix:** Added `&& m.gitRepoRoot != ""` to the guard condition so that `GetFileStatuses` is only dispatched when both `sopsYamlPath` and `gitRepoRoot` are non-empty, preventing a spurious goroutine from opening an unrelated git repo when the user is not in a git repository.

---

### WR-03: Health check stale-file detection flags files with `err != nil` from `GetLastCommitTime` as stale

**Files modified:** `internal/app/model.go`
**Commit:** a7b8e74
**Applied fix:** Restructured the staleness check so that `err == nil` wraps the entire non-error path (including the `!commitTime.IsZero()` check and the zero-time silent-skip comment). The `else` branch now appends to `result.Errors` with a descriptive `git error:` prefix instead of creating a false `StaleFile` entry with `DaysSince: -1`.

---

### WR-04: `HealthModel.ScrollDown` max-scroll calculation — stale scroll after `SetResults`

**Files modified:** `internal/ui/health.go`
**Commit:** 4876055
**Applied fix:** Added `m.scroll = 0` to `SetResults()` with an explanatory comment. This ensures that any scroll position accumulated while the health check was running (loading state) is reset when results arrive, preventing `allLines[m.scroll:]` from panicking if the new result set has fewer lines than the stale scroll position.

---

### WR-05: `stateRecipientList` key handler limits recipient count to 9 without user feedback

**Files modified:** `internal/app/model.go`
**Commit:** 13f0a88
**Applied fix:** Added a `const maxDisplay = 9` cap in `renderRecipientList()`. The `display` slice is truncated to `maxDisplay` when the full list is longer, and a muted note `(showing first 9 of N recipients)` is appended to the rendered lines. The prompt and footer are updated to show `1-displayCount` so they accurately reflect the reachable range.

---

_Fixed: 2026-04-16_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
