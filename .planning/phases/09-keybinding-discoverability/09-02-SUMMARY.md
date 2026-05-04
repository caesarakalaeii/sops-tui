---
phase: "09"
plan: "02"
subsystem: app/keys
tags: [drift-detector, golden-matrix, documentation, d-309]
dependency_graph:
  requires: [09-01]
  provides: [menu-drift-detection, menu-golden-matrix, d-309-amendment]
  affects: [internal/app, internal/app/testdata]
tech_stack:
  added: []
  patterns: [RequireGoldenStructure-golden-testing, suppressHiddenFromMenu-visibility-helper]
key_files:
  created:
    - internal/app/menuhints_drift_test.go
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
  modified:
    - internal/app/model.go
    - internal/app/hints_test.go
decisions:
  - "suppressHiddenFromMenu helper added to model.go to apply visibility suppression in dispatcher arms (Rule 1 bug fix)"
  - "D-309 amendment documents (state, IsSearchActive) as the actual dispatcher signature"
metrics:
  duration: "~8 minutes"
  completed: "2026-05-04"
  tasks: 4
  files: 15
---

# Phase 9 Plan 02: Drift Detector + Golden Matrix + D-309 Documentation Amendment

**Plan:** 09-02
**Phase:** 09 — keybinding-discoverability
**Status:** COMPLETE
**Date completed:** 2026-05-04

## What Shipped

### Wave 0 — Drift Detector Skeleton

- Created `internal/app/menuhints_drift_test.go` with three test functions:
  - `TestMenuHints_Drift` — 13 sub-tests asserting `model.menuHints() == expectedHintsWithSuppression(keymap)` for every dispatcher branch (D-305, D-306, D-307).
  - `TestMenuGolden` — 13 sub-tests asserting per-state RenderMenu output matches `internal/app/testdata/menu_<name>.golden` (D-308, D-310, D-311).
  - `TestMenuGoldenNoPII` — T-09-01 grep-gate against host paths, age private key markers, .sops.yaml strings, ssh-rsa fragments.
- Local `expectedHintsWithSuppression` helper uses a structurally-satisfying `overrider` interface — avoids exporting `menuVisibilityOverrider` from `internal/keys`.
- Also added `suppressHiddenFromMenu()` helper to `internal/app/model.go` (see Rule 1 deviation below).

### Wave 1 — Golden Matrix

- Generated 13 golden files via `GOLDEN_UPDATE=1 go test ./internal/app/... -run TestMenuGolden`.
  - `menu_file_list.golden` — 12 visible hints (k/up, j/down, enter/l, /, i, space, K, H, ?, q, g, G)
  - `menu_file_list_search.golden` — 6 hints (Esc, Enter, j/down, k/up, ?, q)
  - `menu_detail.golden` — 12 visible hints (Blame suppressed via HiddenFromMenu per D-307)
  - `menu_metadata.golden` — 5 hints (j, k, i, Esc, q)
  - `menu_diff.golden` — 6 hints (y, n, Esc, j, k, q)
  - `menu_recipient_confirm.golden` — 5 visible hints (Quit suppressed per D-313)
  - `menu_bulk_re_key_confirm.golden` — 5 visible hints (Quit suppressed per D-313)
  - `menu_help.golden` — 3 hints (Esc, ?, q)
  - `menu_history.golden` — 5 hints (j, k, b, Esc, q)
  - `menu_health.golden` — 5 hints (j, k, H, Esc, q)
  - `menu_recipient_form.golden` — 2 hints (Enter, Esc)
  - `menu_recipient_list.golden` — 3 hints (1-9, Esc, q)
  - `menu_format_menu.golden` — 4 hints (j, k, Enter, Esc) — no Quit per OQ-3
- Width: 80 (matches buildAppModel WindowSize and the existing resize_80x24 anchor).
- All goldens ANSI-stripped via `testutil.RequireGoldenStructure` — Phase 10 palette pass will not churn (D-311).
- Total golden size: 1,496 bytes.

### Wave 1 — D-309 Documentation Amendment

- **Updated** `internal/app/model.go` — menuHints() doc-block now documents `(state, IsSearchActive)` signature and cites D-309 as the supersession of Phase 7 D-10.
- **Updated** `internal/app/hints_test.go` — package doc-comment updated to reference Phase 9 D-309 hardening and drop `recipientAction-via-state` from the dispatch tuple description.
- **Untouched:** `recipientAction` field declaration at `model.go:262` and its 6 Update() business-logic usages. The field stays — it is not a dispatcher axis but is used for diff-confirm flow disambiguation in Update().

## D-309 Amendment Record

> **Decision (Phase 9 D-309):** The `menuHints()` dispatcher dispatches on `(state, IsSearchActive)`, not the `(state, recipientAction, IsSearchActive)` triple specced in Phase 7 D-10. The `recipientAction` parameter was never wired into menuHints() — confirm flows use separate `sessionState` values (`stateRecipientConfirm`, `stateBulkReKeyConfirm`) so the dispatcher needs only `(state, IsSearchActive)`. The `recipientAction` field on `AppModel` continues to be used in Update() business logic, not in dispatch.
>
> **Effect:** Phase 7 `07-CONTEXT.md` D-10 stays intact as historical record. Phase 9 supersedes it via this entry. The corresponding code-side change is comment-only in model.go and hints_test.go (no signature changes; the dispatcher always took `(state, IsSearchActive)` in practice).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed Quit-suppression gap in menuHints() dispatcher**
- **Found during:** Task 1 — drift detector caught the mismatch
- **Issue:** `RecipientConfirmKeyMap` and `BulkReKeyConfirmKeyMap` both define `HiddenFromMenu()` returning `Quit` per D-313. However, the `menuHints()` dispatcher for `stateRecipientConfirm` and `stateBulkReKeyConfirm` called `keys.HintsFromBindings(km.ShortHelp())` directly, which returns all entries with `Visible=true`. This meant Quit was being rendered in the persistent menu — contradicting the D-313 intent (UI-SPEC §confirm-flow-quit-suppression).
- **Fix:** Added `suppressHiddenFromMenu(hidden []key.Binding, hints []keys.MenuHint) []keys.MenuHint` helper to `model.go`. Applied it in the `stateRecipientConfirm` and `stateBulkReKeyConfirm` dispatcher arms so `Quit` is `Visible=false` before RenderMenu renders the menu.
- **Side effect:** Updated `TestMenuHints_StateRecipientConfirm` and `TestMenuHints_StateBulkReKeyConfirm` in `hints_test.go` to assert the correct suppressed behavior (Quit is Visible=false) rather than the old wrong all-visible behavior.
- **Files modified:** `internal/app/model.go`, `internal/app/hints_test.go`
- **Commit:** c1f62a2

## Test Counts

- Tests added by Plan 2: 13 drift sub-tests + 13 golden sub-tests + 1 PII grep-gate = 27 new sub-tests.
- Total `internal/app` test count after Plan 2: 128 RUN lines (including parent test entries).
- Full suite green at every commit.

## Threat Mitigation

- **T-09-01** (Information Disclosure via test fixtures) — mitigated by `TestMenuGoldenNoPII`: any future regression that captures `v.Content` (full chrome with rendered info-panel paths and age fingerprints) instead of `RenderMenu(hints, 80)` (binding metadata only) will fail CI. Verified: `grep -rE '/Users/|/home/|BEGIN AGE|\.sops\.yaml|ssh-rsa |AGE-SECRET-KEY' internal/app/testdata/menu_*.golden` returns nothing.

## Forward Notes for Phase 10

- **D-307 visibility-override interface export:** Plan 1 left `menuVisibilityOverrider` unexported. Plan 2's drift detector re-declared the same interface shape locally (Go's structural typing satisfied it). If Phase 10 sub-models need to opt into the visibility-override contract from `internal/ui`, exporting the interface as `MenuVisibilityOverrider` from `internal/keys` is a 1-line change with zero call-site impact.
- **`?` overlay derivation:** Phase 9 retained the `?` full-screen overlay as the complete reference per UI-11 (no derive-from-keymap refactor). Phase 10 / Phase 11 may revisit if the overlay's manual structure becomes a maintenance burden — every keymap now has `FullHelp()` ready to drive a derived overlay.
- **Menu golden palette stability:** Phase 10's palette tune may need to verify these 13 goldens stay green (`RequireGoldenStructure` strips ANSI, so palette changes do NOT churn the structural goldens — by design, D-311).

## SC Closure

- **SC1 — Total derivation (UI-09):** CLOSED. Plan 1 + Plan 2's drift detector. Per-sub-model `TestXxxHints` + 13 compile-time `var _ help.KeyMap` checks + runtime equality via `TestMenuHints_Drift`.
- **SC2 — Modal re-hydration (UI-10):** CLOSED. Plan 2's 13-entry golden matrix + drift detector lock the menu per state.
- **SC3 — Pure function of (state, IsSearchActive) (UI-10):** CLOSED. Plan 2's drift detector covers all 13 tuples; D-309 documentation amendment closes the recipientAction artifact.
- **SC4 — `?` overlay retained (UI-11):** CLOSED. No Phase 9 refactor; existing `TestHelpHints` and `TestHelpView*` continue to pass.
- **SC5 — No second edit (UI-09):** CLOSED. Plan 1's keymap-as-source-of-truth + Plan 2's drift detector + golden matrix triple-locks this contract.

## Commits

- c1f62a2: test(09-02): drift detector skeleton — 13 sub-tests for menuHints() equality
- 96deaf0: feat(09-02): generate 13 ANSI-stripped menu golden files (D-308, D-310, D-311)
- b4bd3ce: docs(09-02): D-309 documentation amendment — dispatcher contract is (state, IsSearchActive)

## Self-Check

- [x] All 4 plan tasks executed in order
- [x] Each task committed atomically with conventional-commit message
- [x] `go test ./...` green at end of every committed task
- [x] `go vet ./...` clean
- [x] 13 golden files at `internal/app/testdata/menu_*.golden`
- [x] Goldens use `RequireGoldenStructure` (ANSI-stripped)
- [x] `TestMenuGoldenNoPII` passes (T-09-01 mitigated)
- [x] D-309 amendment in both `model.go` and `hints_test.go`
- [x] Production `recipientAction` code untouched (7 references remain)
- [x] `?` overlay regression-free (TestHelpHints and TestHelpView* green)
