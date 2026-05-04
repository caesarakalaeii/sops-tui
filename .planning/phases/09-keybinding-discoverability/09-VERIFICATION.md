---
phase: 09-keybinding-discoverability
verified: 2026-05-04T07:54:00Z
status: passed
score: 5/5 must-haves verified
overrides_applied: 0
---

# Phase 9: Keybinding Discoverability — Verification Report

**Phase Goal:** Every interactive sub-model exposes its hotkeys through a single `Hints()` interface derived from its existing keymap, and the persistent menu re-hydrates per view — so the always-visible hints never diverge from the keys that actually work.
**Verified:** 2026-05-04T07:54:00Z
**Status:** passed
**Re-verification:** No — initial verification

---

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Every interactive sub-model exposes `Hints() []keys.MenuHint` derived from `key.Binding.ShortHelp()` — keymap is the single source of truth | VERIFIED | All 8 sub-models call `keys.HintsFromBindings(m.keys.ShortHelp())`. DetailModel additionally applies `HiddenFromMenu()` suppression inline. Zero literal `MenuHint{}` slices remain in production `internal/ui/*.go`. 13 compile-time `var _ help.KeyMap` assertions in `bindings_test.go` confirm interface satisfaction. |
| 2 | The persistent menu re-hydrates from the active sub-model's `Hints()` on every `View()` call; modal states show their modal keybindings | VERIFIED | `menuHints()` dispatcher at `model.go:1523` routes every `sessionState` to the correct `Hints()` or keymap-derived call. `stateRecipientConfirm` and `stateBulkReKeyConfirm` arms apply `suppressHiddenFromMenu()` before returning (D-313). `TestMenuGolden` 13-entry golden matrix locks this per state and passes. |
| 3 | Menu hint dispatch is a pure function of `(state, IsSearchActive)` (D-309 amends original SC3 to drop `recipientAction`) | VERIFIED | `menuHints()` signature reads only `m.state` and `m.fileList.IsSearchActive()`. D-309 amendment documents this in `model.go:1512-1522` and `hints_test.go:1-7`. `grep '(state, recipientAction, IsSearchActive)' model.go` returns nothing. `grep 'recipientAction' hints_test.go` returns only the D-309 doc-comment reference. 13-sub-test drift detector `TestMenuHints_Drift` is fully green covering all tuples. |
| 4 | The `?` full-screen help overlay is retained as the complete reference | VERIFIED | `HelpModel.View()` still renders `help.Model` full help content from `DefaultDetailKeyMap`/`DefaultFileListKeyMap`. No Phase 9 changes to overlay structure. All 6 `TestHelpView*` and `TestHelpHints` tests pass. |
| 5 | Changing a `key.Binding` value automatically updates the rendered menu — no second edit ever required | VERIFIED | Every `Hints()` derives from `m.keys.ShortHelp()` (zero literal slices). `TestMenuHints_Drift` enforces runtime equality `menuHints() == HintsFromBindings(km.ShortHelp())` for all 13 states. `TestMenuGolden` 13-entry matrix would fail on next run if the derived output diverges from the golden, requiring golden regeneration to surface in review. |

**Score:** 5/5 truths verified

---

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/keys/bindings.go` | 11 new keymap types + `menuVisibilityOverrider` | VERIFIED | 11 types present (HelpKeyMap, DiffKeyMap, HealthKeyMap, HistoryKeyMap, MetadataKeyMap, RecipientFormKeyMap, FileListSearchKeyMap, RecipientConfirmKeyMap, BulkReKeyConfirmKeyMap, RecipientListKeyMap, FormatMenuKeyMap). Interface `menuVisibilityOverrider` unexported at line 41. All `Default*KeyMap` instances present. |
| `internal/keys/hints.go` | Only `MenuHint`, `Hinter`, `HintsFromBindings` — 5 inline vars deleted | VERIFIED | File is 54 lines. Contains only the 3 named exports. No `FileListSearchHints`, `RecipientConfirmHints`, `BulkReKeyConfirmHints`, `RecipientListHints`, or `FormatMenuHints` package-level vars. |
| `internal/ui/help.go` | `Hints()` derives from `HelpKeyMap.ShortHelp()` | VERIFIED | Line 100: `return keys.HintsFromBindings(m.keys.ShortHelp())` |
| `internal/ui/diff.go` | `Hints()` derives from `DiffKeyMap.ShortHelp()` | VERIFIED | Line 180: `return keys.HintsFromBindings(m.keys.ShortHelp())` |
| `internal/ui/health.go` | `Hints()` derives from `HealthKeyMap.ShortHelp()` | VERIFIED | Line 185: `return keys.HintsFromBindings(m.keys.ShortHelp())` |
| `internal/ui/history.go` | `Hints()` derives from `HistoryKeyMap.ShortHelp()` | VERIFIED | Line 137: `return keys.HintsFromBindings(m.keys.ShortHelp())` |
| `internal/ui/metadata.go` | `Hints()` derives from `MetadataKeyMap.ShortHelp()` | VERIFIED | Line 171: `return keys.HintsFromBindings(m.keys.ShortHelp())` |
| `internal/ui/recipientform.go` | `Hints()` derives from `RecipientFormKeyMap.ShortHelp()` | VERIFIED | Line 167: `return keys.HintsFromBindings(m.keys.ShortHelp())` |
| `internal/ui/filelist.go` | `Hints()` one-liner (g/G moved into ShortHelp per D-304) | VERIFIED | Line 384: `return keys.HintsFromBindings(m.keys.ShortHelp())` |
| `internal/ui/detail.go` | `Hints()` applies `HiddenFromMenu()` suppression | VERIFIED | Lines 822-831: derives from `m.keys.ShortHelp()`, then loops `m.keys.HiddenFromMenu()` to set `Visible=false` on Blame. |
| `internal/app/menuhints_drift_test.go` | Drift detector with 13 sub-tests + 13 golden sub-tests + PII grep-gate | VERIFIED | File exists. `TestMenuHints_Drift` (13 sub-tests), `TestMenuGolden` (13 sub-tests), `TestMenuGoldenNoPII` (1 grep-gate) all present and green. Local `overrider` interface declared inside `expectedHintsWithSuppression` — `menuVisibilityOverrider` not referenced. |
| `internal/app/testdata/menu_*.golden` | 13 ANSI-stripped golden files | VERIFIED | Exactly 13 files: bulk_re_key_confirm, detail, diff, file_list, file_list_search, format_menu, health, help, history, metadata, recipient_confirm, recipient_form, recipient_list. Total size 1,496 bytes. No ANSI escapes, no PII. |
| `internal/app/model.go` (D-309 amendment) | menuHints() doc-block documents `(state, IsSearchActive)`; `(state, recipientAction, IsSearchActive)` form removed | VERIFIED | Lines 1512-1522 document D-309 amendment. Old tuple form absent. `recipientAction` field at line 262 and 6 Update() usages untouched. |
| `internal/app/hints_test.go` (D-309 amendment) | Package doc-comment drops `recipientAction-via-state` reference | VERIFIED | Line 3 only references D-309 doc-comment. `grep 'recipientAction' hints_test.go` returns line 3 only (the D-309 citation itself), no old tuple references. |

---

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| `DetailKeyMap.HiddenFromMenu()` | `DetailModel.Hints()` | Loop at `detail.go:823-830` | WIRED | Suppression applied before return. |
| `DefaultRecipientConfirmKeyMap.HiddenFromMenu()` | `menuHints()` dispatcher | `suppressHiddenFromMenu()` at `model.go:1541-1542` | WIRED | Quit Visible=false in confirm state. |
| `DefaultBulkReKeyConfirmKeyMap.HiddenFromMenu()` | `menuHints()` dispatcher | `suppressHiddenFromMenu()` at `model.go:1546-1547` | WIRED | Quit Visible=false in bulk re-key state. |
| Every sub-model `Hints()` | `AppModel.menuHints()` dispatcher | Switch arms `model.go:1527-1561` | WIRED | All 12 session states + search-active override covered. |
| `menuHints()` output | `RenderMenu()` in `View()` | Chrome rendering path | WIRED | `TestMenuGolden` exercises `RenderMenu(m.menuHints(), 80)` for all 13 states and goldens match. |

---

### Data-Flow Trace (Level 4)

Not applicable — this phase produces only metadata (binding descriptions are compile-time constants, not runtime data). The "data" is `key.Binding.Help()` values set via `key.WithHelp()` at package init time; no database, no fetch, no external I/O.

---

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| All keys/ui/app tests green | `go test ./internal/keys/... ./internal/ui/... ./internal/app/... -count=1` | `ok` for all 3 packages | PASS |
| Drift detector 13 sub-tests green | `go test ./internal/app/... -run TestMenuHints_Drift -count=1` | All 13 PASS | PASS |
| Golden matrix 13 sub-tests green | `go test ./internal/app/... -run TestMenuGolden -count=1` | All 13 PASS | PASS |
| PII grep-gate passes | `go test ./internal/app/... -run TestMenuGoldenNoPII -count=1` | PASS | PASS |
| Exactly 13 golden files present | `ls internal/app/testdata/menu_*.golden \| wc -l` | 13 | PASS |
| No literal MenuHint slices in production ui | `grep -rn 'MenuHint{' internal/ui/*.go` (excluding tests) | Only in chrome_test.go (test fixtures — acceptable) and menu_test.go (test helpers — acceptable) | PASS |
| 5 inline hint-set vars deleted | `grep -rn 'FileListSearchHints\|RecipientConfirmHints\|...' internal/` | Only in test function names and bindings.go doc-comments referencing historical names | PASS |
| No PII in golden files | `grep -rE '/Users/\|/home/\|BEGIN AGE...' internal/app/testdata/menu_*.golden` | No matches (exit 1) | PASS |
| `go vet ./...` clean | `go vet ./...` | No output | PASS |
| Full suite green | `go test ./... -count=1` | All 9 packages pass | PASS |
| `?` overlay tests regression-free | `go test ./internal/ui/... -run TestHelp -count=1` | All 6 TestHelp* PASS | PASS |

---

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| UI-09 | 09-01, 09-02 | Every interactive sub-model exposes `Hints() []keys.MenuHint` derived from `key.Binding.ShortHelp()` | SATISFIED | 8 sub-models confirmed. `HintsFromBindings` is the sole mechanism. 5 inline vars deleted from `hints.go`. 13 compile-time `var _ help.KeyMap` assertions. `TestMenuHints_Drift` runtime equality. |
| UI-10 | 09-01, 09-02 | Persistent menu re-hydrates from active sub-model's `Hints()` on every `View()` call; modal states show modal keybindings | SATISFIED | `menuHints()` dispatcher covers all 13 `(state, IsSearchActive)` tuples. `TestMenuGolden` 13-entry matrix locks rendered output per state. `suppressHiddenFromMenu()` applied to confirm states. |
| UI-11 | 09-01, 09-02 | `?` full-screen help overlay retained as the complete reference | SATISFIED | No changes to overlay structure. `HelpModel.View()` unchanged. 6 `TestHelpView*` + `TestHelpHints` all green. |

---

### Anti-Patterns Found

No blockers. Three advisory items noted from code review (WR-01..WR-03):

| File | Pattern | Severity | Impact |
|------|---------|----------|--------|
| `internal/ui/diff.go` (Update method) | WR-01: `kMsg.String()` literal comparisons in `Update()` instead of `key.Matches()` | Info | Phase 9 contract is keymap-as-source-of-truth for `Hints()` only. `Update()` key handling is outside Phase 9 scope (D-301 specifies `Hints()` derivation, not `Update()` routing). Candidate cleanup for Phase 10/11. Not a regression. |
| `internal/app/menuhints_drift_test.go` | WR-02: Drift detector is self-referential for the 11 stateless-state sub-models — `menuHints()` calls `HintsFromBindings(keys.DefaultXxxKeyMap.ShortHelp())` and the drift test calls the same. Correctness is locked by the 13 golden files, not the drift equality check alone. | Info | Known design choice per 09-RESEARCH.md. The goldens are the real correctness lock; drift catches dispatcher routing mistakes and visibility override omissions. No action required. |
| `internal/app/model.go` | WR-03: `stateEdit` is a dead `sessionState` constant — no `menuHints()` case for it. | Info | Pre-existing issue, not introduced by Phase 9. Phase 10/11 cleanup candidate. |

---

### Human Verification Required

None. All success criteria are verifiable programmatically. Visual appearance of the persistent menu is locked by the 13 ANSI-stripped golden files — palette changes (Phase 10) will not churn them because `RequireGoldenStructure` strips ANSI.

---

## Gaps Summary

No gaps. All 5 success criteria are met:

- SC1 (total derivation / UI-09): All 8 sub-models call `HintsFromBindings(m.keys.ShortHelp())`. Zero literal `MenuHint{}` slices in production `internal/ui/*.go`. 5 inline package vars deleted from `hints.go`. 13 compile-time assertions in `bindings_test.go`.
- SC2 (per-view re-hydration / UI-10): `menuHints()` dispatcher covers all 13 states. `TestMenuGolden` matrix green.
- SC3 (pure function / UI-10 + D-309): Dispatcher reads only `(m.state, m.fileList.IsSearchActive())`. D-309 documentation amendment present in both `model.go` and `hints_test.go`.
- SC4 (`?` overlay retained / UI-11): Overlay unchanged. All 6 help tests green.
- SC5 (no second edit): Single source of truth enforced by derivation chain + `TestMenuHints_Drift` runtime equality + `TestMenuGolden` golden lock.

---

_Verified: 2026-05-04T07:54:00Z_
_Verifier: Claude (gsd-verifier)_
