---
phase: 9
slug: keybinding-discoverability
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-30
---

# Phase 9 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing` + `stretchr/testify` v1.x |
| **Config file** | none — `go test ./...` directly |
| **Quick run command** | `go test ./internal/keys/... ./internal/ui/... ./internal/app/... -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~30 seconds (full suite, locally) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/keys/... ./internal/ui/... -count=1`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** ~30 seconds

---

## Per-Task Verification Map

> Filled in by the planner during plan generation. Each PLAN.md task contributes one row.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 09-01-* | 01 | 1+ | UI-09 | — | N/A | unit (existing per-sub-model `TestXxxHints`) | `go test ./internal/ui/... -count=1` | ✅ | ⬜ pending |
| 09-01-* | 01 | 1 | UI-09 | — | N/A | compile (`var _ help.KeyMap = XxxKeyMap{}`) | `go build ./internal/keys/...` | ❌ W0 | ⬜ pending |
| 09-01-* | 01 | 1 | UI-09 | — | N/A | hint-equality update | `go test ./internal/app/... -run TestMenuHints -count=1` | ✅ (update) | ⬜ pending |
| 09-02-* | 02 | 1+ | UI-10 | — | N/A | runtime equality drift detector | `go test ./internal/app/... -run TestMenuHints_Drift -count=1` | ❌ W0 | ⬜ pending |
| 09-02-* | 02 | 2 | UI-10 | — | N/A | 13-entry golden matrix | `go test ./internal/app/... -run TestMenuGolden -count=1` | ❌ W0 | ⬜ pending |
| 09-02-* | 02 | 2 | UI-11 | — | N/A | regression — `?` overlay tests stay green | `go test ./internal/ui/... -run TestHelp -count=1` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Nyquist Signals per Success Criterion

> Each SC needs ≥ 2 independent test signals (Nyquist sampling theorem applied to verification).

| SC | Signal 1 | Signal 2 |
|----|----------|----------|
| SC1 (UI-09: total derivation) | Per-sub-model `TestXxxHints` (existing, expected values updated to match keymap-derived output) | Compile-time `var _ help.KeyMap = XxxKeyMap{}` checks for all 11 new keymaps |
| SC2 (UI-10: modal re-hydration) | 13-entry golden matrix (`testdata/menu_*.golden`) — locks rendered menu per state | Drift detector — runtime equality `model.Hints() == HintsFromBindings(km.ShortHelp())` per sub-model |
| SC3 (UI-10: pure function of `(state, IsSearchActive)`) | Existing 13 `TestMenuHints_State*` tests in `internal/app/hints_test.go` | 13 new golden files locking rendered output per `(state, IsSearchActive)` tuple |
| SC4 (UI-11: `?` overlay retained as complete reference) | `TestHelpHints` and `TestHelpView*` continue to pass unchanged | Help model `FullHelp()` still returns the complete binding set (no regression in overlay behaviour) |
| SC5 (changing a `key.Binding` auto-updates the menu — no second edit) | Drift detector — any `Hints()` slice that lags behind `key.Binding.Help()` description fails CI | Golden matrix — any binding description change requires regenerating goldens or the matrix turns red |

---

## Wave 0 Requirements

> Tests/files that must exist before normal task waves begin (no production code changes Wave 0).

**Plan 1 — Wave 0 (test scaffolding for keymap extraction):**

- [ ] `internal/keys/bindings_test.go` — add compile-time `var _ help.KeyMap = XxxKeyMap{}` assertions for all 11 new keymaps (`HelpKeyMap`, `DiffKeyMap`, `HealthKeyMap`, `HistoryKeyMap`, `MetadataKeyMap`, `RecipientFormKeyMap`, `FileListSearchKeyMap`, `RecipientConfirmKeyMap`, `BulkReKeyConfirmKeyMap`, `RecipientListKeyMap`, `FormatMenuKeyMap`)
- [ ] Update 4 hint-equality assertions in `internal/app/hints_test.go` (lines 91, 99, 139, 147) — replace `keys.RecipientConfirmHints`/etc. references with `keys.HintsFromBindings(keys.DefaultRecipientConfirmKeyMap.ShortHelp())` after the package vars are deleted

**Plan 2 — Wave 0 (test scaffolding for drift detector + golden matrix):**

- [ ] `internal/app/menuhints_drift_test.go` — new file: `TestMenuHints_Drift` (runtime equality, per sub-model) + `TestMenuGolden_*` (13-entry matrix, one sub-test per `(state, IsSearchActive)` tuple)
- [ ] `internal/app/testdata/menu_*.golden` — 13 new golden files (generated initially via `GOLDEN_UPDATE=1 go test ./internal/app/... -run TestMenuGolden`)
  - 12 per-state goldens: `menu_stateFileList.golden`, `menu_stateDetail.golden`, `menu_stateMetadata.golden`, `menu_stateDiff.golden`, `menu_stateHelp.golden`, `menu_stateHistory.golden`, `menu_stateHealth.golden`, `menu_stateRecipientList.golden`, `menu_stateRecipientForm.golden`, `menu_stateFormatMenu.golden`, `menu_stateRecipientConfirm.golden`, `menu_stateBulkReKeyConfirm.golden`
  - 1 search-active override: `menu_stateFileList_search.golden`

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Visual confirmation: persistent menu still reads identically to Phase 8 baseline (no churn) | UI-09, UI-10 | Goldens prove byte-equivalence post-strip; this manual check confirms no human-visible regression in colour/spacing/font weight | Run `make run` (or `go run ./cmd/sops-tui`) on a repo with `.sops.yaml`; navigate through every state (File list, Detail, Help, Diff, Health, History, Metadata, RecipientForm, RecipientList, FormatMenu, RecipientConfirm, BulkReKeyConfirm, FileList /search). Confirm menu reads identical to pre-Phase-9 |
| Confirm `?` overlay still shows complete binding reference (UI-11) | UI-11 | Overlay structure isn't covered by golden matrix (matrix is menu-only per D-310) | Open `?` overlay from each major state; confirm full binding reference renders (no missing rows, no stale text) |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references (Plan 1: bindings_test.go + hints_test.go updates; Plan 2: drift detector + 13 goldens)
- [ ] No watch-mode flags (Go's `go test` doesn't expose one; `-count=1` used to disable cache)
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
