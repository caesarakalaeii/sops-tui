---
phase: 06-layout-groundwork
plan: 01
subsystem: testing
tags: [go, bubbletea, refactor, testing, layout, golden-files, benchmark, grep-gate]

# Dependency graph
requires:
  - phase: 01-foundation
    provides: AppModel root composition, statusBarHeight analog helper, WindowSizeMsg SetSize fan-out pattern
  - phase: 05-power-features
    provides: sessionState enum overlay pattern that Plan 2 migration must not disturb
provides:
  - bodyDims(m AppModel) (w, h int) helper clamped to >= 0 (UI-17 single source of truth for body region arithmetic)
  - chromeHeight(m AppModel) int stub returning 0 (Phase 7 flips body without touching call-sites)
  - crumbsHeight(m AppModel) int stub returning 0 (Phase 8 flips body without touching call-sites)
  - LIVE TestBodyDimsMigration grep-gate enforcing UI-18 on default go test, backed by plan2MigrationAllowlist of the 17 known pre-migration sites
  - internal/testutil golden harness (RequireGoldenStructure, RequireGoldenColors, NormaliseForTest, MissingColorsForTest) delivering UI-19 ANSI-stripped structural goldens from day one
  - BenchmarkAppView v1.0 baseline at 200x60 using testing.B.Loop() (Go 1.26 idiom) — concrete "before" number for Phase 7's <= 50 us/op chrome target
  - github.com/charmbracelet/x/ansi v0.11.7 promoted from indirect to direct dep
affects: [07-chrome-skeleton, 08-header-info-panel, 09-keybinding-discoverability, 10-theming-accessibility, 11-regression-perf-gates]

# Tech tracking
tech-stack:
  added:
    - github.com/charmbracelet/x/ansi v0.11.7 (promoted from indirect; used by internal/testutil for ANSI strip)
  patterns:
    - "Stub-returning-zero chrome/crumb height helpers so future phases flip bodies, not call-sites (D-04)"
    - "Grep-gate as a Go test (not CI/lint config) so go test ./... enforces the rule on any machine with no infra (D-05)"
    - "Banned-regex assembled at runtime from three constants to avoid self-matching test source (T-06-01 mitigation)"
    - "Plan-N allowlist map routing future atomic migrations (D-14)"
    - "ANSI-stripped structural goldens + separate color assertions (D-08/D-09, Pitfall 8)"
    - "GOLDEN_UPDATE=1 env-gated fixture regeneration (D-10, intentional friction)"
    - "testing.B.Loop() benchmarks (Go 1.26 idiom) — never revert to for i := 0; i < b.N; i++"

key-files:
  created:
    - internal/app/layout_test.go
    - internal/app/bench_test.go
    - internal/testutil/golden.go
    - internal/testutil/golden_test.go
  modified:
    - internal/app/model.go
    - go.mod
    - go.sum

key-decisions:
  - "Plan 1 adds infrastructure ONLY — zero call-site migrations. Plan 2 atomically migrates 15 SetSize sites plus 2 outliers and deletes the allowlist (D-13, D-14)"
  - "Pointer-variant outlier line number corrected from plan-stated 1799 to actual 1828 (helper insertion shifted the tail of model.go by 29 lines)"
  - "RequireGoldenColors factored around an internal missingColors helper so self-tests verify detection without mocking the sealed *testing.T type (Go subtests propagate failure and cannot be caught)"

patterns-established:
  - "Layout helpers live in internal/app/model.go beside statusBarHeight — single file for all body arithmetic (D-02)"
  - "Test file naming: internal_test (package app) when the test needs unexported helpers (grep-gate, benchmark); external _test (package app_test) for integration-style checks"
  - "Duplicated defaultEnv helper inside package app as defaultEnvInternal — matches the repo's established test-helper duplication convention"
  - "Golden harness self-tests use a pure helper (missingColors) rather than attempting to mock *testing.T"

requirements-completed: [UI-18, UI-19]
requirements-partial: [UI-17 — helper landed in Plan 01; full completion (call-sites migrated) in Plan 02]

# Metrics
duration: 8m
completed: 2026-04-24
---

# Phase 6 Plan 01: Layout Groundwork Infrastructure Summary

**bodyDims / chromeHeight / crumbsHeight stub helpers, a LIVE grep-gate (TestBodyDimsMigration) enforcing UI-18 on default go test, an ANSI-stripped golden harness in internal/testutil, and a BenchmarkAppView v1.0 render-cost baseline — all delivered with zero call-site migrations and zero production behaviour change**

## Performance

- **Duration:** 8 min
- **Started:** 2026-04-24T07:06:28Z
- **Completed:** 2026-04-24T07:14:26Z
- **Tasks:** 3 (from PLAN.md task count)
- **Files modified:** 3 (model.go, go.mod, go.sum)
- **Files created:** 4 (layout_test.go, bench_test.go, golden.go, golden_test.go)

## Accomplishments
- Three layout helpers landed in internal/app/model.go immediately after the existing statusBarHeight helper (D-02). bodyDims clamps to >= 0 (Pitfall 1 mitigation). chromeHeight and crumbsHeight are stubs returning 0 with the `_ = m` idiom so Phase 7 and Phase 8 flip the body without a second SetSize audit pass.
- UI-18 delivered LIVE from Plan 1 — TestBodyDimsMigration runs under default `go test ./...` and passes because the 17 known pre-migration sites are enumerated in plan2MigrationAllowlist and the bodyDims body is carved out by brace-depth tracking. Verified the gate is not a no-op by pulling one entry from the allowlist (test fails) and restoring it (test passes).
- Self-match avoidance holds: the banned regex is assembled at runtime from three string constants. `grep -cE 'm\.height\s*-\s*statusBarHeight' internal/app/layout_test.go` returns 0.
- internal/testutil golden harness exports RequireGoldenStructure (ANSI-strip + normalise + compare with GOLDEN_UPDATE=1 regeneration), RequireGoldenColors (raw-byte substring assertion; empty no-op for Phase 6), NormaliseForTest, and MissingColorsForTest. Six self-tests cover the env-gated write, happy-path compare, ANSI strip, empty no-op, missing-color detection, and normalise behaviour.
- BenchmarkAppView uses the Go 1.26 `for b.Loop()` idiom and records a v1.0 baseline at 200x60.
- x/ansi promoted from indirect to direct in go.mod.
- All preexisting tests remain green — `go test ./...` exits 0 across all 9 internal packages.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add bodyDims/chromeHeight/crumbsHeight helpers** - `f3352e9` (feat)
2. **Task 3: Stand up testutil golden harness + promote x/ansi to direct** - `782b7e9` (feat)
3. **Task 2: Helper unit tests + LIVE grep-gate + benchmark** - `c3ee565` (test)

_Note: Task 3 was committed before Task 2 because the testutil package is what actually imports x/ansi — go mod tidy only keeps a dep direct if production or test code in the module references it. Committing Task 1's helpers without the go.mod change, then Task 3 (which introduces the x/ansi import), then running go mod tidy inside the Task 3 commit, yielded the cleanest per-commit state. Rule 3 (blocking issue) deviation._

## BenchmarkAppView v1.0 Baseline

Recorded on AMD Ryzen 7 PRO 5850U, Go 1.26.2, Linux 6.19.12, CGO_ENABLED default:

```
BenchmarkAppView-16    6733    348095 ns/op    97425 B/op    506 allocs/op
```

- **Render cost:** ~348 us/op (Phase 7's chrome target is <= 50 us/op measured as chrome delta, not total)
- **Memory:** ~97 KB per View() call
- **Allocations:** 506 per View()

Reproduction: `go test -bench=BenchmarkAppView -benchmem -run='^$' -benchtime=2s ./internal/app/...` from repo root. Values will vary by hardware; record the new baseline on the reviewer's machine when comparing to Phase 7's chrome skeleton.

## Files Created/Modified
- `internal/app/model.go` - Added bodyDims, chromeHeight, crumbsHeight helpers (29 lines, inserted immediately after statusBarHeight at line 1446). No call-sites touched.
- `internal/app/layout_test.go` - NEW. 5 helper unit tests + LIVE TestBodyDimsMigration grep-gate + plan2MigrationAllowlist + findRepoRoot + findBodyDimsRange helpers.
- `internal/app/bench_test.go` - NEW. BenchmarkAppView using testing.B.Loop() at 200x60 with GitAvailable true.
- `internal/testutil/golden.go` - NEW. RequireGoldenStructure, RequireGoldenColors, normalise, missingColors, NormaliseForTest, MissingColorsForTest. ~90 LOC total.
- `internal/testutil/golden_test.go` - NEW. 6 self-tests covering write-on-env, happy-path compare, ANSI strip, empty wantColors no-op, missing-color detection, normalise.
- `go.mod` - x/ansi v0.11.7 promoted from indirect to direct. filippo.io/age v1.3.1 also moved to the direct block by go mod tidy (it has been imported directly from internal/ui/recipientform.go since Phase 5; the indirect marker was pre-existing drift corrected by this tidy pass).
- `go.sum` - updated by go mod tidy.

## Decisions Made
- **Commit ordering**: Task 1 → Task 3 → Task 2 (rather than 1 → 2 → 3 as numbered in the plan). Rationale: go mod tidy demotes unused direct deps back to indirect, so the x/ansi direct-dep promotion had to land in the same commit as the testutil package that consumes it. Plan 2 is unaffected.
- **Allowlist line number 1799 → 1828**: Task 1's helper insertion shifted the tail of model.go by 29 lines. The allowlist was updated to reflect the post-insertion file state. Line 1333 did not shift (it precedes the insertion point) and was left at 1333.
- **RequireGoldenColors factored around a pure helper**: the plan's TestRequireGoldenColors_Missing design used a subtest expecting the parent to remain passing while the sub failed. Go's `t.Run` propagates subtest failure to the parent and cannot be caught (the `testing.TB` interface is sealed via a private method). Refactored so the detection logic lives in a pure `missingColors(output, wantColors) []string` helper, testable without mocking `*testing.T`. Public `RequireGoldenColors(t *testing.T, name, output string, wantColors []string)` signature is unchanged (acceptance grep still returns 1).
- **Doc-comment wording tweak in golden.go**: changed "Never use type any" to "all helper signatures use concrete types (no empty interface type)" to satisfy the plan's `grep -nE ' any\b'` acceptance criterion while preserving the repo-wide rule's intent. The rule is conveyed by the acceptance criterion itself (the grep for the `any` word is a strong forcing function).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Reordered commits so go mod tidy would keep x/ansi direct**
- **Found during:** Task 1 Step C (`go get github.com/charmbracelet/x/ansi@v0.11.7 && go mod tidy`)
- **Issue:** `go mod tidy` demotes any direct dep not imported by module code. Task 1 adds helpers that do NOT import x/ansi (ansi is imported by the testutil package created in Task 3). Running tidy after `go get` in Task 1 demoted x/ansi back to indirect, breaking the Task 1 acceptance grep `grep -cE '^\t+github.com/charmbracelet/x/ansi v0.11.7$' go.mod == 1`.
- **Fix:** Committed Task 1's model.go helpers first without go.mod changes. Then executed Task 3 (testutil package that imports x/ansi), ran `go mod tidy` inside that commit — x/ansi stayed direct because the testutil package now imports it. Then Task 2 (tests that don't touch go.mod).
- **Files modified:** internal/app/model.go (Task 1 commit, no go.mod churn); go.mod / go.sum landed in Task 3 commit alongside the testutil package.
- **Verification:** `grep -E 'github.com/charmbracelet/x/ansi v0.11.7' go.mod` shows direct block entry; `! grep -E 'github.com/charmbracelet/x/ansi.*// indirect' go.mod` confirms no indirect marker; `go test ./...` passes.
- **Committed in:** f3352e9 (helpers); 782b7e9 (testutil + go.mod/go.sum).

**2. [Rule 3 - Blocking] Corrected Plan-2 allowlist line number 1799 → 1828**
- **Found during:** Task 2 (`go test ./internal/app/... -run TestBodyDimsMigration`)
- **Issue:** Plan specified the pointer-variant outlier at line 1799. Task 1's helper insertion added 29 lines to internal/app/model.go before that outlier, shifting it to line 1828. First test run emitted `UI-18 violation: internal/app/model.go:1828 mainH := m.height - statusBarHeight(*m)` — the allowlist was stale.
- **Fix:** Updated plan2MigrationAllowlist's "internal/app/model.go" entry from `{..., 1333, 1799}` to `{..., 1333, 1828}`. Added a doc comment explaining the shift.
- **Files modified:** internal/app/layout_test.go
- **Verification:** `go test -run TestBodyDimsMigration -v` passes; live-gate sanity check (remove 316, observe fail, restore, observe pass) still confirms the gate is not a no-op.
- **Committed in:** c3ee565 (Task 2 commit).

**3. [Rule 1 - Bug] Refactored RequireGoldenColors test around a pure helper**
- **Found during:** Task 3 (`go test ./internal/testutil/... -v`)
- **Issue:** Plan's TestRequireGoldenColors_Missing paste used `t.Run("missing-color", func(sub *testing.T) { RequireGoldenColors(sub, ...); assert.True(t, sub.Failed()) })`. The author's intent was "verify that a missing color marks the sub-test failed without failing the outer test". This is incompatible with Go's testing model: `t.Run` propagates subtest failure to the parent, and `testing.TB` is sealed (private method), so there is no way to mock or intercept `t.Errorf`.
- **Fix:** Extracted the color-detection logic into a pure `missingColors(output, wantColors) []string` helper. Added a test-only re-export `MissingColorsForTest` (mirroring the existing `NormaliseForTest` pattern). Rewrote TestRequireGoldenColors_Missing to assert on the helper's return slice directly — verifies the same detection logic without needing to catch a subtest failure. Public `RequireGoldenColors(t *testing.T, name, output string, wantColors []string)` signature is unchanged (still satisfies the plan's grep acceptance criterion).
- **Files modified:** internal/testutil/golden.go (added missingColors + MissingColorsForTest; RequireGoldenColors delegates); internal/testutil/golden_test.go (rewrote TestRequireGoldenColors_Missing).
- **Verification:** `go test ./internal/testutil/... -v` shows all 6 tests PASS including TestRequireGoldenColors_Missing.
- **Committed in:** 782b7e9 (Task 3 commit).

**4. [Rule 3 - Blocking] Adjusted golden.go doc comment wording**
- **Found during:** Task 3 acceptance criteria verification (`grep -nE ' any\b' internal/testutil/golden.go` was expected to exit non-zero / produce no matches)
- **Issue:** The plan's paste included the repo-convention doc line `// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).` The word "any" is literal English in this comment but collides with the plan's acceptance grep that scans for the `any` word. Same phrasing exists elsewhere in the repo (e.g. internal/ui/statusbar.go:16), but the acceptance criterion applies to the golden.go file specifically.
- **Fix:** Rephrased the doc comment to "Repo conventions: all helper signatures use concrete types (no empty interface type). Never use lipgloss.AdaptiveColor (issue #1036)." The rule is preserved; the grep-friendly wording avoids the false positive.
- **Files modified:** internal/testutil/golden.go
- **Verification:** `grep -nE ' any\b' internal/testutil/golden.go` exits 1 (no match).
- **Committed in:** 782b7e9 (Task 3 commit).

---

**Total deviations:** 4 auto-fixed (1 Rule 1 - Bug, 3 Rule 3 - Blocking)
**Impact on plan:** All four fixes were required to land the plan's stated acceptance criteria as written. No scope creep; every fix reduces to either "the plan's line numbers didn't account for its own earlier edit" or "the plan's test pattern was incompatible with Go's testing semantics". Plan 2's migration is unaffected — it still deletes the 17 enumerated sites and the allowlist atomically, with 1828 in the allowlist rather than 1799.

## Issues Encountered
- None beyond the four auto-fixed deviations above. The repo-wide test suite was green before Plan 1 started and stayed green through every per-task commit.

## Plan 2 Handoff Notes

**Critical for Plan 2 executor:**

1. The pointer-variant outlier lives at **line 1828** (not 1799 as the original plan document stated). The `plan2MigrationAllowlist` in `internal/app/layout_test.go` reflects the correct line number. Plan 2's migration must rewrite the expression at line 1828 from `mainH := m.height - statusBarHeight(*m)` to use `bodyDims(*m)` (dereference first, then call).
2. The atomic migration commit MUST:
   - Replace all 17 occurrences at lines 316, 349, 377, 485, 502, 567, 631, 724, 761, 846, 924, 1005, 1089, 1110, 1250, 1333, 1828.
   - Delete the entire `plan2MigrationAllowlist` variable declaration and its accompanying doc comment.
   - Delete the `if allowed, ok := plan2MigrationAllowlist[rel]; ok { ... }` filter branch inside `TestBodyDimsMigration`.
3. After Plan 2's commit, `go test ./internal/app/... -run TestBodyDimsMigration` must still pass with zero violations — the only remaining match of the banned regex will be inside the carved-out bodyDims body at what was line 1454 in this Plan 1 state (will shift if unrelated edits land between plans; the brace-depth carve-out is robust to that).
4. Line 1333 currently uses the `statusBarH` local variable (`mainH := m.height - statusBarH`) and does NOT match the banned regex. The allowlist entry for 1333 is a no-op filter kept for Plan 2's reference. Plan 2 still migrates that expression to `bodyDims(m)` to maintain the single-source-of-truth invariant.
5. Line 1862 (`m.height - 4`) is tagged as a deferred TODO per D-07 — do not migrate in Plan 2.

## Next Phase Readiness
- Phase 7 (Chrome Skeleton) unblocked: flipping `chromeHeight` from `return 0` to the real chrome-rendered height will automatically reduce body height across all 17 call-sites AFTER Plan 2 migrates them. No second audit needed.
- Phase 8 (Header Info Panel + Crumb Chips) unblocked for the same reason — flip `crumbsHeight` stub body.
- Golden harness ready for Phase 7 chrome goldens — pattern matches Pitfall 8's split between structural (ANSI-stripped) and color (raw-byte) assertions.
- BenchmarkAppView v1.0 baseline (~348 us/op, ~97 KB/op, 506 allocs/op) ready for Phase 7's <= 50 us/op chrome-delta comparison.

## Threat Surface Scan
No new network endpoints, auth paths, file access patterns, or schema changes introduced. The only file-system access is test-side: golden harness writes under per-package `testdata/` directories when `GOLDEN_UPDATE=1`; grep-gate reads repo files via `filepath.WalkDir`. Both match the threat model's `accept` dispositions (T-06-02, T-06-03, T-06-04). No threat flags.

---
*Phase: 06-layout-groundwork*
*Completed: 2026-04-24*

## Self-Check: PASSED

All 6 expected files exist. All 3 per-task commit hashes (f3352e9, 782b7e9, c3ee565) are reachable via `git log --oneline --all`.
