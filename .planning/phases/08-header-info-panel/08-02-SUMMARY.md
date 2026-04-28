---
phase: 08-header-info-panel
plan: 02
subsystem: ui, git
tags: [go-git, lipgloss, statusbar, breadcrumb, GetBranch, detached-HEAD]

# Dependency graph
requires:
  - phase: 04-clipboard-git
    provides: GetFileStatuses/GetFileHistory/GetLastCommitTime patterns in internal/git/status.go
  - phase: 07-chrome-skeleton
    provides: StatusBarModel with View(width), SetBreadcrumb, SetItemCount, renderBreadcrumb
provides:
  - git.GetBranch(repoRoot) (branch, detached, err) — D-215 locked signature
  - StatusBarModel.Segments() []string — D-210 read accessor for breadcrumb data
  - StatusBarModel.View() normal path shrunk to right-aligned env+clipboard only — D-211
  - SetItemCount body neutered to no-op — D-209 backward-compat for 14 model.go call-sites
  - renderBreadcrumb private function deleted (dead code after D-211)
  - itemCount/itemLabel struct fields deleted
affects:
  - 08-03 (Plan 3 wires GetBranch into GitStatusMsg + Segments() into RenderCrumbs call)
  - internal/app/model.go (14 SetItemCount call-sites are now no-ops; Plan 3 may clean them up)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "go-git PlainOpenWithOptions + ErrRepositoryNotExists guard pattern reused for GetBranch"
    - "ref.Name().IsBranch() / ref.Name().Short() for branch resolution; ref.Hash().String()[:7] for detached HEAD"
    - "Segments() as inverse of SetBreadcrumb's strings.Join — pure read accessor, value receiver"
    - "SetItemCount no-op pattern: body = _, _ = count, label; signature retained for caller compat"

key-files:
  created: []
  modified:
    - internal/git/status.go
    - internal/git/status_test.go
    - internal/ui/statusbar.go
    - internal/ui/statusbar_test.go

key-decisions:
  - "D-209 author choice: SetItemCount becomes no-op (_, _ = count, label), not deleted — 14 model.go call-sites preserved without touching model.go; Plan 3 owns optional cleanup"
  - "D-215: GetBranch returns (branch, detached, err); non-git dir returns gogit.ErrRepositoryNotExists matching GetFileStatuses D-12 contract"
  - "D-210: Segments() is a value receiver returning nil for empty breadcrumb (not []string{''})"
  - "D-211: View() normal path uses StatusBarStyle.Render(' ') as spacer (package var, not lipgloss.NewStyle()) — avoids new inline NewStyle() calls in View()"
  - "Removed fmt import (was only used by deleted fmt.Sprintf item-count rendering)"

patterns-established:
  - "GetBranch follows exact PlainOpenWithOptions + ErrRepositoryNotExists guard shape as all other functions in internal/git/status.go"
  - "Defensive 7-char hash slice: if len(h) > 7 { h = h[:7] } — avoids panic on zero-length edge case"
  - "3-subtest TestGetBranch mirrors TestGetFileStatuses shape: non-repo / normal-branch / detached-HEAD"

requirements-completed:
  - UI-04
  - UI-08

# Metrics
duration: ~25min
completed: 2026-04-28
---

# Phase 8 Plan 02: git.GetBranch + StatusBar Shrink Summary

**go-git GetBranch helper (D-215) + StatusBarModel shrunk to right-aligned env+clipboard only with Segments() read accessor (D-209/D-210/D-211)**

## Performance

- **Duration:** ~25 min
- **Started:** 2026-04-28T11:00:00Z
- **Completed:** 2026-04-28T11:13:04Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- `git.GetBranch(repoRoot)` shipped with D-215 locked signature — handles normal branch, detached HEAD (7-char hash), and non-git directory (ErrRepositoryNotExists)
- `StatusBarModel.View()` normal path collapsed from three sections (left breadcrumb + center item-count + right env/clipboard with pipes) to right-aligned env+clipboard only (D-211); flash path unchanged per D-212
- `Segments() []string` accessor added — reverses `SetBreadcrumb`'s `strings.Join` via `strings.Split`; returns nil for empty breadcrumb (D-210)
- `renderBreadcrumb` private function deleted (dead code after D-211 left-section removal)
- `itemCount`/`itemLabel` struct fields deleted; `SetItemCount` neutered to no-op `_, _ = count, label` (D-209 author choice — 14 model.go call-sites preserved; Plan 3 owns cleanup)
- Full repo test suite green; `internal/app` still compiles cleanly with the 14 no-op SetItemCount call-sites

## Task Commits

1. **Task 1: git.GetBranch + 3-subtest TestGetBranch** — `f60fd5e` (feat)
2. **Task 2: StatusBarModel shrink + Segments() + tests** — `c4efa0d` (feat)

## Files Created/Modified

- `internal/git/status.go` — GetBranch function appended after GetLastCommitTime
- `internal/git/status_test.go` — TestGetBranch with 3 subtests added
- `internal/ui/statusbar.go` — struct fields deleted, SetItemCount neutered, Segments() added, View() shrunk, renderBreadcrumb deleted, fmt import removed
- `internal/ui/statusbar_test.go` — 4 tests deleted (breadcrumb/item-count/SetBreadcrumb/SetItemCount), 4 tests added (RightAlignOnly/SegmentsAccessor/SegmentsEmpty/FlashUnchanged), 7 flash+env tests kept

## Decisions Made

**D-209 author choice: SetItemCount no-op vs delete**

Chose no-op (`_, _ = count, label`) rather than deletion. Rationale: 14 call-sites exist in `internal/app/model.go`. Deleting the method here would require Plan 2 to also touch `model.go`, creating file-ownership conflict with Plan 3 (which owns `model.go` per CONTEXT D-217). The no-op approach keeps Plan 2 self-contained. Plan 3 may delete the call-sites and the method body in its own commit; Plan 2 does not block that.

**View() spacer in clipboard-hot path**

Used `StatusBarStyle.Render(" ")` as the spacer between `[clip]` and the env indicators (package-level style var, not `lipgloss.NewStyle()`). This avoids introducing any new inline `NewStyle()` calls in `View()`, consistent with Phase 7 D-22 "no NewStyle() in View() reachables". The pre-existing `renderEnvIndicators` NewStyle() calls are pre-existing technical debt outside Phase 8 scope per Pitfall C deferral rule.

**Segments() nil vs empty slice**

`Segments()` returns `nil` (not `[]string{""}`) when breadcrumb is empty, so callers can cleanly branch on `len(segments) == 0`. `TestStatusBar_SegmentsEmpty` uses a zero-value `StatusBarModel` (breadcrumb == "") to exercise the nil path directly.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Plan test code used wrong EnvStatus field name**

- **Found during:** Task 2 (statusbar_test.go writing)
- **Issue:** Plan's `TestStatusBar_RightAlignOnly` sample code used `AgeKeyAvailable: true` but the actual `EnvStatus` struct field is `AgeAvailable`. Would have caused a compile error.
- **Fix:** Used `AgeAvailable: true` matching the actual struct definition in statusbar.go.
- **Files modified:** internal/ui/statusbar_test.go
- **Verification:** `go build ./...` succeeds; all tests pass
- **Committed in:** c4efa0d (Task 2 commit)

**2. [Rule 1 - Bug] Missing `strings` import in test file**

- **Found during:** Task 2 (first test run)
- **Issue:** Test file used `strings.Contains` but `strings` was not in the import block.
- **Fix:** Added `"strings"` to test imports.
- **Files modified:** internal/ui/statusbar_test.go
- **Verification:** `go test ./internal/ui/ -count=1 -run TestStatusBar` exits 0
- **Committed in:** c4efa0d (Task 2 commit)

**3. [Rule 1 - Bug] Unused variable `m` in TestStatusBar_SegmentsEmpty**

- **Found during:** Task 2 (first test run)
- **Issue:** The plan's TestStatusBar_SegmentsEmpty used `m := ui.NewStatusBarModel(...)` but then tested a `var zero ui.StatusBarModel` (zero-value model), leaving `m` declared but unused — compile error.
- **Fix:** Removed the unused `m` declaration; the test directly declares `var zero ui.StatusBarModel` to reach the nil-return path.
- **Files modified:** internal/ui/statusbar_test.go
- **Verification:** Compile and test pass
- **Committed in:** c4efa0d (Task 2 commit)

**4. [Rule 1 - Bug] Plan acceptance criteria grep for itemCount/itemLabel**

- **Found during:** Task 2 (acceptance criteria verification)
- **Issue:** A doc comment on the struct mentioned "itemCount and itemLabel fields removed" — this caused `grep -c "itemCount" statusbar.go` to return 1 instead of 0, failing the acceptance criterion.
- **Fix:** Rephrased comment to "count and label fields removed" (preserving meaning without the field names).
- **Files modified:** internal/ui/statusbar.go
- **Verification:** `grep -c "itemCount" internal/ui/statusbar.go` returns 0
- **Committed in:** c4efa0d (Task 2 commit)

---

**Total deviations:** 4 auto-fixed (all Rule 1 — plan sample code bugs, compile errors, grep criterion)
**Impact on plan:** All fixes were minor compile/test errors in the plan's sample code. No behavioral or architectural changes. Scope stayed exactly within Plan 2's boundary.

## Issues Encountered

None beyond the auto-fixed compile errors noted above.

## Test Changes (statusbar_test.go)

| Change | Test | Reason |
|--------|------|--------|
| Deleted | TestStatusBarBreadcrumbInView | View() no longer renders breadcrumb (D-211) |
| Deleted | TestStatusBarItemCountInView | View() no longer renders item count (D-209/D-211) |
| Deleted | TestStatusBarSetBreadcrumb | View() no longer shows breadcrumb for assertion |
| Deleted | TestStatusBarSetItemCount | SetItemCount is a no-op — no testable behavior |
| Added | TestStatusBar_RightAlignOnly | Verifies D-211: only right cluster, no pipes, full width |
| Added | TestStatusBar_SegmentsAccessor | Verifies D-210: Segments() round-trips SetBreadcrumb |
| Added | TestStatusBar_SegmentsEmpty | Verifies D-210: nil return for empty breadcrumb |
| Added | TestStatusBar_FlashUnchanged | Verifies D-212: flash path center-aligned full-width |
| Kept | 7 flash+env indicator tests | Unchanged behavior — all still pass |

## Forward Work for Plan 3

- **model.go SetItemCount call-sites:** 14 call-sites at `m.status.SetItemCount(...)` are now no-ops. Plan 3 may delete them and the method signature if it modifies model.go. Plan 2 does not block this.
- **Segments() wiring:** Plan 3 calls `ui.RenderCrumbs(m.status.Segments(), m.width)` to replace the `""` placeholder in `AppModel.View()`'s sections slice and to flip `crumbsHeight` from `return 0`.
- **GetBranch wiring:** Plan 3 extends the `GitStatusMsg` async cmd to call `git.GetBranch(m.gitRepoRoot)` and populate `m.infoPanel.GitBranch` / `m.infoPanel.GitDetached`.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

Plan 3 (Wave 2) can proceed immediately. Both primitive deliverables (`GetBranch` and `Segments()`) are ready for wiring into AppModel. The status-bar shrink is live — the bottom bar now shows only env+clipboard right-aligned.

---
*Phase: 08-header-info-panel*
*Completed: 2026-04-28*
