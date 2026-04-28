---
phase: 08-header-info-panel
plan: 01
subsystem: ui
tags: [lipgloss, bubbletea, age, ansi, tui, crumbs, infopanel]

requires:
  - phase: 07-chrome-skeleton
    provides: InfoPanelPlaceholderStyle (38x6 slot), styles.go palette, package-var discipline
  - phase: 07.1-chrome-gap-closure
    provides: TestSubmodelViewsNoNewStyle scope, TestViewNoNewStyle BFS walker, TestChromeASCIIOnly allowlist

provides:
  - InfoPanelData struct + RenderInfoPanel pure renderer (infopanel.go)
  - middleTruncate helper using ansi.Truncate + ansi.TruncateLeft (Pitfall B fix)
  - RenderCrumbs pure renderer with k9s-exact chip pills (crumbs.go)
  - truncateSegmentsToWidth middle-segment ellipsis overflow (D-216)
  - ParseAgeKeyFingerprint + AgeKeyFilePath helpers (agekey.go)
  - 8 new package-level style vars in styles.go (D-201..D-208)
  - 20 unit tests across 3 test files

affects:
  - 08-02 (git.GetBranch + statusbar shrink — reads InfoPanelData shape)
  - 08-03 (AppModel integration — wires all 3 primitives via chrome + crumbsHeight)

tech-stack:
  added: []
  patterns:
    - "Pure renderer pattern: RenderInfoPanel(InfoPanelData) string — zero I/O, zero AppModel coupling (Pitfall 15)"
    - "middleTruncate uses ansi.Truncate (left half) + ansi.TruncateLeft (right half) for genuine middle-ellipsis (Pitfall B)"
    - "type-assert *age.X25519Identity before Recipient().String() — never Identity.String() directly (Pitfall 11 / D-220 Q1)"
    - "k9s-verbatim chip pills: <segment> + normaliseSegments (lowercase+strip-spaces) + active=Bold+bg+fg (D-205..D-208)"
    - "All styles declared as package-level vars in styles.go — zero lipgloss.NewStyle() in renderer files (TestSubmodelViewsNoNewStyle compliance)"

key-files:
  created:
    - internal/ui/infopanel.go
    - internal/ui/crumbs.go
    - internal/ui/agekey.go
    - internal/ui/infopanel_test.go
    - internal/ui/crumbs_test.go
    - internal/ui/agekey_test.go
  modified:
    - internal/ui/styles.go

key-decisions:
  - "Bold(true) on active chip is the redundant encoding channel for 16-color terminal compat (D-206 + Pitfall 9) — k9s uses bg-only swap which fails on 16-color; sops-tui adds bold weight deliberately"
  - "middleTruncate uses ansi.TruncateLeft for right fragment (Pitfall B fix) — ansi.Truncate is tail-only and cannot produce middle-truncation alone"
  - "type-assert to *age.X25519Identity before calling Recipient().String() (D-220 Q1) — calling Identity.String() directly leaks AGE-SECRET-KEY private key prefix"
  - "$SOPS_AGE_KEY_FILE checked before ~/.config/sops/age/keys.txt fallback (D-214) — consistent with SOPS CLI behaviour"
  - "TestRenderCrumbs_ActiveBoldBg checks for [1; prefix in ANSI sequence — lipgloss/v2 encodes bold as combined SGR 1;38;2;... not standalone 1m"

patterns-established:
  - "Renderer files contain zero lipgloss.NewStyle() calls — all styles are package vars in styles.go (forward-compat with Plan 3 TestSubmodelViewsNoNewStyle scope extension)"
  - "ellipsisSentinel constant (U+2026) as magic value in truncateSegmentsToWidth — RenderCrumbs detects sentinel and applies CrumbChipEllipsisStyle"
  - "InfoPanelData.RecipientCount and FileCount use -1 as the 'not yet computed' sentinel (not 0) — 0 is a valid count"

requirements-completed:
  - UI-04
  - UI-05
  - UI-07

duration: 5min
completed: 2026-04-28
---

# Phase 8 Plan 01: Header Info Panel + Crumb Chips — Primitives Summary

**Pure renderers RenderInfoPanel (5-row info panel) and RenderCrumbs (k9s-exact chip pills) + age key fingerprint parser, all with zero AppModel coupling and zero lipgloss.NewStyle() calls in renderer files**

## Performance

- **Duration:** 5 min
- **Started:** 2026-04-28T11:04:17Z
- **Completed:** 2026-04-28T11:12:49Z
- **Tasks:** 3
- **Files modified/created:** 7 (3 production, 3 test, 1 modified)

## Accomplishments

- Shipped `RenderInfoPanel(InfoPanelData) string` — 5-row panel in locked cfg/age/rcp/git/fil order with ASCII `-` empty markers, middle-truncation (Pitfall B fix using `ansi.TruncateLeft` for right fragment), and U+2026 ellipsis (D-201..D-204)
- Shipped `RenderCrumbs([]string, int) string` — k9s-exact `<segment>` chip pills with three-channel active encoding (accent bg + body fg + Bold per D-206/Pitfall 9), D-207 normalisation, D-216 middle-segment overflow
- Shipped `ParseAgeKeyFingerprint` + `AgeKeyFilePath` — type-asserts to `*age.X25519Identity` before `Recipient().String()`, never calls `Identity.String()` directly (D-220 Q1 private key leak prevention)
- Added 8 package-level style vars to `styles.go` (D-201..D-208), all using existing colour constants, zero `AdaptiveColor`
- 20 unit tests covering all D-219 named cases; full test suite remains green

## Task Commits

1. **Task 1: Add 8 new style vars to internal/ui/styles.go** — `256cdf6` (feat)
2. **Task 2: Implement internal/ui/agekey.go + agekey_test.go** — `9e0cab3` (feat)
3. **Task 3: Implement infopanel.go + crumbs.go + 2 test files** — `52fca6a` (feat)

## Files Created/Modified

- `internal/ui/styles.go` — extended with 8 Phase 8 package-level style vars (InfoPanelLabelStyle, InfoPanelValueStyle, InfoPanelSepStyle, CrumbChipStyle, CrumbChipActiveStyle, CrumbChipSepStyle, CrumbChipEllipsisStyle, CrumbRowStyle)
- `internal/ui/agekey.go` — AgeKeyFilePath + ParseAgeKeyFingerprint; type-asserts to `*age.X25519Identity`
- `internal/ui/agekey_test.go` — 5 tests: FirstIdentity, MissingFile, MalformedFile, EnvOverride, HomeDefault
- `internal/ui/infopanel.go` — InfoPanelData struct, RenderInfoPanel, middleTruncate (ansi.Truncate + ansi.TruncateLeft)
- `internal/ui/infopanel_test.go` — 6 tests: AllRowsAligned, EmptyMarkers, TruncatesAge, TruncatesPath, GitDirtyMarker, GitDetachedHead
- `internal/ui/crumbs.go` — RenderCrumbs, normaliseSegments, truncateSegmentsToWidth, measureChipRow, containsEllipsisSentinel
- `internal/ui/crumbs_test.go` — 9 tests: KnsExactPills, ActiveBoldBg, LowercaseStripSpaces, MiddleEllipsis, EmptySafe, SingleSegmentIsActive, InactiveChipColors, TruncateDropsMiddle, TwoSegmentsNeverTruncated

## Decisions Made

- `TestRenderCrumbs_ActiveBoldBg` checks for `[1;` in ANSI output (not `1m`) because lipgloss/v2 encodes bold as a combined SGR sequence `\x1b[1;38;2;...;48;2;...m`, not as standalone `\x1b[1m`.
- `middleTruncate` in infopanel.go uses `ansi.TruncateLeft(s, sw-right, "")` for the right fragment (Pitfall B). `ansi.Truncate` is tail-only; without `TruncateLeft` the result is right-truncated, not middle-truncated.
- age fingerprint is ALWAYS middle-truncated to ≤10 cells regardless of input length (D-220 Q5 security gate), not only when longer than 10 cells.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed `TestRenderCrumbs_MiddleEllipsis` width parameter**
- **Found during:** Task 3 (crumbs_test.go initial run)
- **Issue:** Test used width=40 for 8 single/short-char segments totalling ~36 rendered cells — all segments fit without overflow, so no ellipsis was produced
- **Fix:** Changed width from 40 to 20 (chip budget 18 cells forces truncation for 8 segments)
- **Files modified:** internal/ui/crumbs_test.go
- **Verification:** TestRenderCrumbs_MiddleEllipsis PASS
- **Committed in:** 52fca6a (Task 3 commit)

**2. [Rule 1 - Bug] Fixed `TestRenderCrumbs_ActiveBoldBg` bold SGR assertion**
- **Found during:** Task 3 (crumbs_test.go initial run)
- **Issue:** Test asserted `assert.Contains(t, out, "1m")` — lipgloss/v2 encodes bold as `1;` combined with other SGR params in one escape sequence, not as standalone `1m`
- **Fix:** Changed to `strings.Contains(out, "[1;") || strings.Contains(out, "[1m")` to match both combined and standalone bold encoding
- **Files modified:** internal/ui/crumbs_test.go
- **Verification:** TestRenderCrumbs_ActiveBoldBg PASS
- **Committed in:** 52fca6a (Task 3 commit)

**3. [Rule 3 - Blocking] Fixed unused `"strings"` import in crumbs_test.go**
- **Found during:** Task 3 (first build after writing crumbs_test.go)
- **Issue:** Test file had `"strings"` in imports but no direct `strings.` call (plan template included `strings` but the test functions didn't use it yet)
- **Fix:** Removed `"strings"` import, re-added when `strings.Contains` was needed for bold assertion fix
- **Files modified:** internal/ui/crumbs_test.go
- **Verification:** `go build ./...` exits 0
- **Committed in:** 52fca6a (Task 3 commit)

---

**Total deviations:** 3 auto-fixed (2 test correctness bugs, 1 blocking import error)
**Impact on plan:** All three are test-only fixes. Production files match plan specification exactly. No scope creep.

## Issues Encountered

None beyond the test-level deviations documented above.

## Known Stubs

None — all three production files are complete implementations with no placeholder values or TODO markers. Plan 1 delivers pure primitives with no AppModel wiring; the stubs are intentionally in Plan 3 (AppModel integration + chrome wiring).

## Threat Flags

None — all three new files are pure renderer/parser utilities with no network endpoints, no auth paths, no file-write operations, and no schema changes. `agekey.go` reads from `~/.config/sops/age/keys.txt` (read-only, user-controlled path) and is called only at startup, not from `View()`.

## Next Phase Readiness

Plan 2 (`08-02-PLAN.md`) can proceed immediately:
- `InfoPanelData` struct shape is locked — Plan 2's `git.GetBranch` will populate `GitBranch`/`GitDetached` fields
- `StatusBarModel.Segments()` accessor needed by `RenderCrumbs` is Plan 2's responsibility
- All Plan 2 changes are additive to different files (`internal/git/status.go`, `internal/ui/statusbar.go`)
- Zero conflicts with Plan 1 deliverables

Plan 3 (`08-03-PLAN.md`) AppModel integration is unblocked pending Plan 2 completion.

---
*Phase: 08-header-info-panel*
*Completed: 2026-04-28*
