---
phase: 09-keybinding-discoverability
reviewed: 2026-05-04T00:00:00Z
depth: standard
files_reviewed: 27
files_reviewed_list:
  - internal/app/hints_test.go
  - internal/app/menuhints_drift_test.go
  - internal/app/model.go
  - internal/app/testdata/menu_bulk_re_key_confirm.golden
  - internal/app/testdata/menu_detail.golden
  - internal/app/testdata/menu_diff.golden
  - internal/app/testdata/menu_file_list.golden
  - internal/app/testdata/menu_file_list_search.golden
  - internal/app/testdata/menu_format_menu.golden
  - internal/app/testdata/menu_health.golden
  - internal/app/testdata/menu_help.golden
  - internal/app/testdata/menu_history.golden
  - internal/app/testdata/menu_metadata.golden
  - internal/app/testdata/menu_recipient_confirm.golden
  - internal/app/testdata/menu_recipient_form.golden
  - internal/app/testdata/menu_recipient_list.golden
  - internal/keys/bindings.go
  - internal/keys/bindings_test.go
  - internal/keys/hints.go
  - internal/keys/hints_test.go
  - internal/ui/detail.go
  - internal/ui/diff.go
  - internal/ui/filelist.go
  - internal/ui/health.go
  - internal/ui/help.go
  - internal/ui/history.go
  - internal/ui/metadata.go
  - internal/ui/recipientform.go
findings:
  critical: 0
  warning: 3
  info: 4
  total: 7
status: issues_found
---

# Phase 9: Code Review Report

**Reviewed:** 2026-05-04
**Depth:** standard
**Files Reviewed:** 27
**Status:** issues_found

## Summary

Phase 9 is a pure keybinding-discoverability refactor: inline hint-set literals replaced
by typed keymap types in `internal/keys/bindings.go`, sub-model `Hints()` methods derived
via `HintsFromBindings(km.ShortHelp())`, a runtime drift detector in
`menuhints_drift_test.go`, and 13 ANSI-stripped golden fixtures. No new user-visible
keybindings are introduced.

The overall structure is sound. The `suppressHiddenFromMenu` helper and the
`menuVisibilityOverrider` interface are correctly wired; the golden fixtures contain no
PII and match expected layout. Three warnings merit fixing before the phase ships:

1. `DiffModel.Update` hardcodes key strings that are now out of sync with `DiffKeyMap`
   (the refactor added a keymap but did not migrate the Update switch).
2. The drift detector in `TestMenuHints_Drift` is structurally self-referential for 11 of
   13 sub-tests — it compares `menuHints()` against a derivation using the exact same
   keymap variable the production code uses, so a consistent-but-wrong change to the
   keymap would not be caught. The golden tests are the real correctness lock, but the
   drift test's stated purpose ("catch drift between keymap and dispatcher") is not
   fulfilled for sub-model states.
3. `stateEdit` is a declared `sessionState` constant that is never assigned to
   `m.state`; it falls through `menuHints()` to the nil default arm, producing a silent
   blank menu if it were ever reached.

---

## Warnings

### WR-01: `DiffModel.Update` uses hardcoded key strings not tied to `DiffKeyMap`

**File:** `internal/ui/diff.go:101-118`

**Issue:** `DiffModel.Update` dispatches on `kMsg.String()` with the literal strings
`"y"`, `"n"`, `"esc"`, `"j"`, `"k"`. Phase 9 introduced `DiffKeyMap` and
`DefaultDiffKeyMap` with authoritative bindings, but never migrated `Update` to use
`key.Matches`. This means:

- A future change to `DefaultDiffKeyMap.Confirm` (e.g., mapping `"enter"` instead of
  `"y"`) silently has no effect on actual Update behaviour.
- The "keymap is the single source of truth" invariant (D-301) is violated specifically
  for `DiffModel`.
- `HealthModel` and `HistoryModel` do not have `Update` methods (parent drives scroll),
  so only `DiffModel` and `RecipientFormModel` are affected. `RecipientFormModel.Update`
  already uses hardcoded strings too but its keymap has only 2 bindings and the risk is
  lower. `DiffModel` is the critical path for all write confirmations.

**Fix:**
```go
// diff.go Update() — replace the kMsg.String() switch:
case tea.KeyPressMsg:
    switch {
    case key.Matches(msg, m.keys.Confirm):
        m.confirmed = true
    case key.Matches(msg, m.keys.Cancel):
        m.cancelled = true
    case key.Matches(msg, m.keys.Close):
        m.cancelled = true
    case key.Matches(msg, m.keys.ScrollDown):
        m.ScrollDown()
    case key.Matches(msg, m.keys.ScrollUp):
        m.ScrollUp()
    }
    return m, nil
```
Add `"charm.land/bubbles/v2/key"` to the import block.

---

### WR-02: Drift detector is self-referential for sub-model states — does not catch keymap-only mutations

**File:** `internal/app/menuhints_drift_test.go:77-163`

**Issue:** For 11 of 13 sub-tests (all states except `stateRecipientConfirm` and
`stateBulkReKeyConfirm`), `TestMenuHints_Drift` asserts:

```go
require.Equal(t,
    expectedHintsWithSuppression(keys.DefaultDetailKeyMap),
    m.menuHints())  // which calls m.detail.Hints() which calls keys.HintsFromBindings(m.keys.ShortHelp())
```

`m.detail.keys` is set to `keys.DefaultDetailKeyMap` in `NewDetailModel`. Both sides of
the assertion trace back to `keys.DefaultDetailKeyMap.ShortHelp()` — the same package
variable. A change that mutates `DefaultDetailKeyMap.ShortHelp()` ordering and also
updates `m.detail.keys` consistently would still pass this test even if the menu output
changed visibly.

The golden tests (`TestMenuGolden`) do lock the rendered string content and are the real
correctness contract. The drift detector is therefore redundant for sub-model states
unless it is redesigned to compare against an independent reference (e.g., a hard-coded
expected `[]MenuHint` slice). As-written, it only reliably catches the case where
`menuHints()` dispatches to the wrong state arm entirely, which is already covered by
`TestMenuHints_State*` count assertions in `hints_test.go`.

The two confirm-state sub-tests (`stateRecipientConfirm`, `stateBulkReKeyConfirm`) are
exempt from this criticism because `expectedHintsWithSuppression` applies suppression
logic via the `overrider` interface check, while `menuHints()` calls
`suppressHiddenFromMenu` — a separate code path that could diverge, so the comparison
is meaningful there.

**Fix:** Either:
(a) Document in the test that the golden matrix (`TestMenuGolden`) is the correctness
    lock and the drift test only validates dispatch routing (acceptable scope reduction),
    or
(b) For each sub-model state, hard-code the expected `[]keys.MenuHint` slice in the test
    so that a keymap mutation without a golden update is caught.

Option (a) is the minimum viable fix:
```go
// t.Run("stateDetail", ...) — add comment:
// NOTE: This test only validates that menuHints() dispatches to the correct sub-model.
// String-level correctness is locked by TestMenuGolden/detail. A keymap mutation
// that is consistently reflected in both m.detail.keys and DefaultDetailKeyMap would
// pass this test — update the golden to catch those.
```

---

### WR-03: `stateEdit` is a dead `sessionState` constant that silently falls through `menuHints()` to nil

**File:** `internal/app/model.go:57-58, 1523-1561`

**Issue:** `stateEdit` is declared at line 58 in the `sessionState` iota but is never
assigned to `m.state` anywhere in `model.go` (confirmed by searching all
`m.state = state*` assignments). The inline edit mode uses `m.detail.editActive` (a
`bool` on `DetailModel`) instead.

Because `stateEdit` has no case in the `menuHints()` switch, if it were ever activated
(e.g., by a future branch or test that sets `m.state = stateEdit`), the function returns
`nil`. This is then passed to `ui.RenderMenu(nil, width)` which iterates a nil slice
safely but produces a blank menu — silent UI regression with no error signal.

The exported alias `StateEdit = stateEdit` (line 78) increases the risk by making the
constant reachable from test code.

**Fix:** Remove `stateEdit` from the iota and remove the `StateEdit` export, or add an
explicit case to `menuHints()` that documents its intent:
```go
case stateEdit:
    // Inline edit is a DetailModel sub-state (editActive bool), not a top-level state.
    // menuHints delegates to detail so the menu shows Detail bindings during editing.
    return m.detail.Hints()
```
The second option is preferable if `stateEdit` is kept for historical compatibility.

---

## Info

### IN-01: `TestMenuHints_StateFileList_NoSearch` count assertion uses a magic number

**File:** `internal/app/hints_test.go:34`

**Issue:** The assertion `require.Equal(t, 12, len(hints), ...)` hard-codes `12` rather
than deriving it from `len(keys.DefaultFileListKeyMap.ShortHelp())`. If the keymap
gains a binding (even in a future phase), this test gives a less actionable error than a
diff of the actual vs expected hints.

**Fix:**
```go
expected := keys.HintsFromBindings(keys.DefaultFileListKeyMap.ShortHelp())
require.Equal(t, expected, hints, "FileList.Hints() must match DefaultFileListKeyMap-derived hints")
```
This removes the magic number and makes the failure message show which hint changed.

---

### IN-02: `TestMenuHints_StateDetail` count assertion uses a magic number with misleading comment

**File:** `internal/app/hints_test.go:71`

**Issue:** `require.Equal(t, 13, len(hints), "Detail.Hints() returns 13 entries (one Visible=false)")`.
The count `13` is the raw `ShortHelp()` length. After `HiddenFromMenu()` suppression,
the slice still has 13 elements but one has `Visible=false` — the count assertion does
not verify the suppression happened. A bug where suppression is skipped entirely would
still pass this test.

**Fix:** Assert on content, not count:
```go
hints := m.menuHints()
visible := 0
for _, h := range hints {
    if h.Visible { visible++ }
}
require.Equal(t, 12, visible, "Detail menu must show exactly 12 visible hints (Blame suppressed)")
require.Equal(t, 13, len(hints), "Detail Hints slice must have 13 entries total")
// Optionally: assert that Blame is the suppressed one
blameIdx := -1
for i, h := range hints {
    if h.Mnemonic == "b" { blameIdx = i; break }
}
require.True(t, blameIdx >= 0, "Blame hint must be present")
require.False(t, hints[blameIdx].Visible, "Blame must be Visible=false")
```

---

### IN-03: `TestMenuGoldenNoPII` PII regex does not cover Windows-style absolute paths

**File:** `internal/app/menuhints_drift_test.go:218`

**Issue:** The `forbidden` regex covers `/Users/` and `/home/` (Unix absolute paths) but
not `C:\` or `%APPDATA%` (Windows-style paths). The project currently targets Linux and
macOS (per `CLAUDE.md`), so this is not an active risk. However, the comment "absolute
home paths" implies the intent is to cover all home-directory patterns.

**Fix:** This is a forward compatibility note; no immediate action required unless the
project adds Windows CI. If desired, add `|[A-Z]:\\\\Users\\\\` to the pattern.

---

### IN-04: Golden fixtures for confirm states (`menu_recipient_confirm.golden`, `menu_bulk_re_key_confirm.golden`) do not assert the suppressed `q` binding

**File:** `internal/app/testdata/menu_recipient_confirm.golden:1-5`,
`internal/app/testdata/menu_bulk_re_key_confirm.golden:1-5`

**Issue:** The goldens lock only the 5 *visible* rows rendered by `RenderMenu`. They do
not (and cannot) verify that `q` is present-but-suppressed (`Visible=false`) in the
hints slice. The correctness of the suppression logic is tested in
`TestMenuHints_StateRecipientConfirm` (hints_test.go:95-103), but if `RenderMenu`'s
`Visible=false` filtering were removed, the golden would fail (because `q` would appear
as a 6th row) — the golden does provide a regression signal for that case. The concern
is narrower: if the suppression were moved from `menuHints()` to happen *before*
`HintsFromBindings` (i.e., `Quit` is removed from `ShortHelp()` entirely rather than
suppressed), the golden would still pass but the `?` full-screen overlay would lose `q`
from its binding list.

This is a design observation, not an actionable bug in Phase 9 scope. The current
implementation is correct. Noting it for Phase 10 when the `?` overlay refactor lands.

---

_Reviewed: 2026-05-04_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
