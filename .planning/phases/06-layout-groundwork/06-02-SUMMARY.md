---
phase: 06-layout-groundwork
plan: 02
subsystem: ui
tags: [go, bubbletea, migration, golden-tests, layout, refactor]

# Dependency graph
requires:
  - phase: 06-layout-groundwork
    plan: 01
    provides: bodyDims/chromeHeight/crumbsHeight helpers, LIVE TestBodyDimsMigration grep-gate with plan2MigrationAllowlist, internal/testutil golden harness (RequireGoldenStructure / RequireGoldenColors), x/ansi direct dep
provides:
  - 17-site migration of all banned `m.height - statusBarHeight(...)` subtractions in internal/app/model.go to bodyDims(m) / bodyDims(*m) — single source of truth for body-region arithmetic (UI-17 closed)
  - Deletion of plan2MigrationAllowlist variable and its filter branch from internal/app/layout_test.go — the grep-gate now enforces UI-18 with only the bodyDims body carve-out, so any future developer who copy-pastes the banned subtraction trips a red test (UI-18 closed definitively)
  - Four ANSI-stripped resize goldens under internal/app/testdata/ at 40x12, 80x24, 120x40, 200x60 proving UI-17 end-to-end across the k9s target terminal-size range (Pitfall 1 mitigated)
  - internal/app/resize_test.go with four TestResize_<WxH> tests reusing package app_test's existing defaultEnv() helper
  - .gitattributes pinning `*.golden text eol=lf -whitespace` so editors/CI cannot mutate fixtures
  - TODO(phase-7) tag above `boxHeight := m.height - 4` in renderRecipientList marking the deferred modal-chrome migration
affects: [07-chrome-skeleton, 08-header-info-panel, 09-keybinding-discoverability, 10-theming-accessibility, 11-regression-perf-gates]

# Tech tracking
tech-stack:
  added: []  # Plan 2 added no dependencies — migration only
  patterns:
    - "Atomic migration + allowlist-deletion commit (D-14) so the grep-gate never widens silently between migration and allowlist cleanup"
    - "ANSI-stripped resize goldens at four distinct sizes (40x12, 80x24, 120x40, 200x60) — 80x24 alone hides clamp-at-zero regressions (Pitfall 1)"
    - ".gitattributes `eol=lf -whitespace` for all .golden files — defends against editor auto-format drift"

key-files:
  created:
    - internal/app/resize_test.go
    - internal/app/testdata/resize_40x12.golden
    - internal/app/testdata/resize_80x24.golden
    - internal/app/testdata/resize_120x40.golden
    - internal/app/testdata/resize_200x60.golden
    - .gitattributes
  modified:
    - internal/app/model.go
    - internal/app/layout_test.go

key-decisions:
  - "Outlier 1 (View at former line 1333) collapses to `_, mainH := bodyDims(m)` — statusBar local preserved; the now-dead `statusBarH` local removed because no other View code referenced it"
  - "Outlier 2 (showBulkReKeyConfirm at former line 1828) uses `w, h := bodyDims(*m)` with value deref — matches the existing statusBarHeight(*m) call pattern for pointer-receiver context"
  - "TODO(phase-7) comment added above `boxHeight := m.height - 4` with the canonical wording (including the article 'a' before 'named modal-frame constant') — exact text is load-bearing for the acceptance grep"
  - "Allowlist deletion landed in the same atomic commit as the 17-site migration (D-14). Plan-2-allowlist references also removed from the TestBodyDimsMigration error message and docstring so `grep -c 'plan2MigrationAllowlist'` returns 0"
  - "Resize goldens generated via GOLDEN_UPDATE=1 and visually inspected: all four show the `No SOPS files found` empty-state (no .sops.yaml in the test cwd) with trailing padding rows and the status bar pinned to the bottom row — no content ever paints below the status bar at any size, confirming bodyDims clamp works end-to-end"

patterns-established:
  - "Resize goldens live under internal/app/testdata/ and use RequireGoldenStructure (structural) plus RequireGoldenColors(..., nil) (Phase 6 color no-op; Phase 7 populates)"
  - ".gitattributes is the correct home for per-file-type line-ending/whitespace pins — not .editorconfig (Go tooling does not read .editorconfig)"

requirements-completed: [UI-17, UI-18]

# Metrics
duration: 8min
completed: 2026-04-24
---

# Phase 6 Plan 02: Layout Groundwork Migration Summary

**17-site atomic migration to bodyDims, deletion of the Plan-2 migration allowlist, and four resize goldens at 40x12 / 80x24 / 120x40 / 200x60 proving the clamp holds across the k9s terminal-size range — Phase 6 closed with v1.0 behaviour parity**

## Performance

- **Duration:** ~8 min
- **Started:** 2026-04-24T09:19:00Z (approx; after Plan 01's metadata commit 698e635)
- **Completed:** 2026-04-24T09:27:15Z (e05165d timestamp)
- **Tasks:** 2 (Task 1 atomic migration + allowlist deletion; Task 2 resize tests + goldens + .gitattributes)
- **Files modified:** 2 (model.go, layout_test.go)
- **Files created:** 6 (resize_test.go, four .golden fixtures, .gitattributes)

## Accomplishments

- **All 17 banned-regex call-sites in internal/app/model.go migrated to bodyDims** — 15 `SetSize`/`NewFooModel` sites at the canonical lines (316, 349, 377, 485, 502, 567, 631, 724, 761, 846, 924, 1005, 1089, 1110, 1250) plus the View outlier (formerly 1333) plus the pointer-receiver outlier in showBulkReKeyConfirm (formerly 1828). Post-migration `grep -cE 'm\.height\s*-\s*statusBarHeight' internal/app/model.go` outputs **1** — the only remaining occurrence is inside bodyDims itself.
- **plan2MigrationAllowlist deleted atomically** with the migration. Post-commit `grep -c 'plan2MigrationAllowlist' internal/app/layout_test.go` outputs **0**. The variable declaration, the filter branch inside TestBodyDimsMigration, the mention inside the test docstring, and the Plan-2 reference in the error message were all removed in the same commit as the 17-site migration.
- **TestBodyDimsMigration passes under default `go test ./...`** with zero violations — it relies exclusively on the bodyDims body carve-out (brace-depth range lookup). Any future contributor who copy-pastes the banned `m.height - statusBarHeight(...)` subtraction trips a red test immediately.
- **TODO(phase-7) comment added above `boxHeight := m.height - 4`** in renderRecipientList with canonical wording (`// TODO(phase-7): replace magic -4 with a named modal-frame constant or / // bodyDims usage once modal chrome lands.`). The magic number itself is preserved per D-07 — Phase 7 flips it when modal chrome lands.
- **Four resize goldens** committed under `internal/app/testdata/` at 40x12, 80x24, 120x40, 200x60. Each is plain UTF-8, LF-terminated, trailing-whitespace-free, ANSI-free. Visually inspected: empty-state `No SOPS files found` renders on the top rows, padding fills the middle, status bar pins to the bottom row — **nothing paints below the status bar at any size**, confirming the bodyDims clamp works.
- **Four TestResize_<WxH> tests** in internal/app/resize_test.go (`package app_test`) — each calls `RequireGoldenStructure` plus `RequireGoldenColors(..., nil)`. Shares the existing `defaultEnv()` helper from model_test.go via the same package (no duplication; preconditions confirmed).
- **.gitattributes** added to repo root with `*.golden text eol=lf -whitespace` — defends fixtures against editor auto-format and CI whitespace checks.
- **Full test suite green** including TestBodyDimsMigration plus all four resize tests. BenchmarkAppView still runs cleanly (smoke ran at `-benchtime=1x`, ~527 us/op on this run; within the Plan 01 ~348 us/op baseline ballpark — numbers vary run-to-run at 1x).

## Task Commits

Each task committed atomically per D-14:

1. **Task 1: Atomic 17-site migration + plan2MigrationAllowlist deletion** — `79ec449` (refactor)
2. **Task 2: Resize goldens + resize_test.go + .gitattributes** — `e05165d` (test)

Both commits signed with gitmoji-style conventional commit messages referencing the phase/plan number and UI-17 / UI-18 requirements. Commit 79ec449 shows `2 files changed, 50 insertions(+), 127 deletions(-)` — a net reduction of 77 lines from model.go + layout_test.go as the duplicated clamp boilerplate and the 13-line allowlist block were collapsed.

## Files Created/Modified

- **internal/app/model.go** — 17 banned-regex sites migrated to bodyDims (15 via the canonical `w, h := bodyDims(m)` + `SetSize(w, h)` / `NewFooModel(..., w, h)` pattern; outlier 1 via `_, mainH := bodyDims(m)` preserving `statusBar`; outlier 2 via `w, h := bodyDims(*m)`); TODO(phase-7) comment added above `boxHeight := m.height - 4`. Net: -48 lines (clamp boilerplate collapsed).
- **internal/app/layout_test.go** — plan2MigrationAllowlist variable declaration deleted; filter branch inside TestBodyDimsMigration deleted; Plan-2 allowlist references removed from docstring and error-message text. Net: -29 lines.
- **internal/app/resize_test.go** — NEW. 57 LOC. Four TestResize_<WxH> tests in `package app_test` reusing the package's existing `defaultEnv()` helper from model_test.go.
- **internal/app/testdata/resize_40x12.golden** — NEW. 12 rendered rows (status bar wraps to 2 rows because 40 cols is narrower than the status bar text — existing v1.0 behaviour; body clamps correctly to 10).
- **internal/app/testdata/resize_80x24.golden** — NEW. 24 rows, status bar fits on one row.
- **internal/app/testdata/resize_120x40.golden** — NEW. 40 rows.
- **internal/app/testdata/resize_200x60.golden** — NEW. 60 rows.
- **.gitattributes** — NEW. One line: `*.golden text eol=lf -whitespace`.

## Decisions Made

- **Outlier 1 collapsed to a single bodyDims call.** The plan authorised keeping `statusBarH := lipgloss.Height(statusBar)` if downstream View code referenced it. Inspection confirmed no other code in View() reads `statusBarH` — the local was dead. Removed it along with the `mainH := ... clamp` block; the migration is now 3 lines removed (`statusBarH := ...`, `mainH := ...`, `if mainH < 0 { ...` plus the closing brace and clamp assignment) and 1 line added (`_, mainH := bodyDims(m)`).
- **Docstring + error message also purged of 'plan2MigrationAllowlist'.** The acceptance criterion is `grep -c 'plan2MigrationAllowlist' internal/app/layout_test.go == 0`; the prose references (docstring and error-message text) would have satisfied a code-only interpretation but not the literal grep. Removed both to satisfy the literal criterion and keep the file self-consistent.
- **Resize goldens generated in a repo context without .sops.yaml** (by design — NewAppModel(defaultEnv(), "") passes an empty sopsYamlPath). All four goldens show the "No SOPS files found" empty-state plus the status bar. This is the stable deterministic starting state — no host paths, no timestamps, no env leakage — and matches Plan 2's threat-model disposition `accept` for T-06-07 (information disclosure).
- **.gitattributes created at repo root with a single line** rather than a comment-laden multi-line file. The rule is self-explanatory and the acceptance criterion is coverage of `*.golden`; a one-liner minimises diff noise.

## Deviations from Plan

None requiring deviation-rule handling.

Two small adjustments documented above under Decisions Made are refinements within the plan's explicit authorisation envelope (the plan explicitly said "if anything later in View() still reads `statusBarH`, either keep the `statusBarH := ...` line OR replace with `lipgloss.Height(statusBar)` inline"; and the plan-stated acceptance criterion is a literal `grep -c 'plan2MigrationAllowlist' == 0` which forces removal from prose too).

## Issues Encountered

None. Plan 01 had already resolved the two structural hazards that would otherwise have tripped Plan 02 execution:

- the pointer-variant outlier line shift (1799 → 1828) was already encoded in plan2MigrationAllowlist and the prompt's phase-specific guardrails, and
- the test-file package / helper sharing was already normalised (model_test.go is `package app_test` with `defaultEnv()` at line 16).

`go build ./...`, `go vet ./...`, and `go test ./...` all passed at every intermediate state (pre-commit, after Task 1 commit, after Task 2 commit).

## Pending Verification (post-automation)

**Manual UAT per D-15 — NOT run by the executor (requires a real terminal).** The automated verification passes end-to-end; the manual smoke test below is deferred to `/gsd-verify-work` or an interactive session:

> Build `/tmp/sops-tui` with `go build -o /tmp/sops-tui ./...`. Open a terminal, resize to **40x12**, run `/tmp/sops-tui`, and cycle through: file list → a file's detail view → help (?) → diff → metadata (i) → history (h) → health → recipient form (a). Repeat at **200x60**. Confirm: no content is ever painted under the status bar row at any size.

The automated resize tests at 40x12 and 200x60 cover the AppModel.View() empty-state path structurally. What the manual smoke adds is coverage of the non-empty paths inside each sub-model (fileList populated with real items, detail overlay rendering a parsed file, diff overlay with real entries, etc.) which currently require a real .sops.yaml + age key in scope — out of scope for the executor, in scope for the verifier.

## Phase 6 Sign-Off

**UI-17 — bodyDims single source of truth for body-region arithmetic:** CLOSED.
- Helper landed in Plan 01 (commit f3352e9).
- All 15 SetSize/NewFooModel call-sites + 2 outliers migrated in Plan 02 (commit 79ec449).
- TODO(phase-7) tag preserves the intentional `m.height - 4` in renderRecipientList for Phase 7 to flip.

**UI-18 — grep-gate enforcing the single source of truth:** CLOSED.
- Live under default `go test ./...` since Plan 01 (commit c3ee565).
- Plan 02 deleted plan2MigrationAllowlist (commit 79ec449); the gate now relies exclusively on the bodyDims body carve-out. Verified by pulling the carve-out conceptually — any banned-regex match outside bodyDims now fails the build with a named violation.

**UI-19 — ANSI-stripped structural golden harness:** CLOSED.
- Harness delivered by Plan 01 (`internal/testutil/golden.go`).
- Plan 02 exercises it with four resize goldens at 40x12 / 80x24 / 120x40 / 200x60 (commit e05165d).

## Next Phase Readiness

- **Phase 7 (Chrome Skeleton)** unblocked: flipping `chromeHeight` from `return 0` to the real chrome-rendered height will automatically reduce body height across all 17 migrated call-sites. Zero second-audit pass needed. The TODO(phase-7) tag at `boxHeight := m.height - 4` in renderRecipientList marks the single remaining hand-wired `-constant` site that chrome must either replace or explicitly justify.
- **Phase 8 (Header Info Panel + Crumb Chips)** unblocked for the same reason — flip `crumbsHeight` stub body.
- **Golden harness** ready for Phase 7 chrome goldens. Four resize fixtures at the k9s target sizes give Phase 7 a direct diff target for chrome-added rows.
- **BenchmarkAppView v1.0 baseline** (~348 us/op per Plan 01's 2s sample; single-shot runs show ~527 us/op on this machine) remains the "before chrome" number for Phase 7's <= 50 us/op chrome-delta gate.

## Threat Surface Scan

No new network endpoints, auth paths, file access patterns, or schema changes introduced. The migration is pure refactor; goldens are static test fixtures. All changes remain inside the threat model's existing `accept`/`mitigate` dispositions (T-06-05 / T-06-06 / T-06-11 explicitly mitigated by this plan's acceptance greps). No threat flags.

---
*Phase: 06-layout-groundwork*
*Completed: 2026-04-24*

## Self-Check: PASSED

All 7 expected files exist (`internal/app/resize_test.go`, four resize `.golden` fixtures, `.gitattributes`, this SUMMARY). Both task commit hashes (`79ec449`, `e05165d`) are reachable via `git log --oneline --all`.
