---
phase: 07-chrome-skeleton
plan: 02
subsystem: ui
tags: [tui, lipgloss, chrome-skeleton, chrome-composer, wraptitled, overlaytitle, normalborder, ansi-truncate]

# Dependency graph
requires:
  - phase: 07-chrome-skeleton (Plan 01)
    provides: TitledBorderStyle + TitleLabelStyle + LogoStyleInfo/Warn/Error + MenuKeyStyle/MenuDescStyle/MenuCellStyle package vars; RenderLogo (LogoStatus, width); RenderMenu (hints, width); keys.MenuHint contract
  - phase: 06-layout-groundwork
    provides: ansi.Strip helper integration in testutil/golden, package-var styles discipline, no-AdaptiveColor invariant
  - phase: 01-foundation
    provides: internal/ui module structure, internal/keys/bindings.go shape that hints.go mirrors
provides:
  - ui.RenderChrome(hints, logoStatus, width) - 6-row JoinHorizontal of info-panel placeholder + menu + logo
  - ui.WrapTitled(title, body, width, height) - NormalBorder titled-region wrapper at exact outer dimensions
  - ui.overlayTitle(rendered, title) - unexported string-splice helper at column 2 of top border line
  - ui.spliceRenderedLine(line, startCol, endCol, replacement) - unexported ANSI-aware column splice helper
  - ui.InfoPanelPlaceholderStyle - new package var (Width 38 x Height 6) for Phase 8 forward-compat
  - 16 unit subtests across 3 test functions verifying all chrome-composer invariants
affects:
  - 07-03 (integration: AppModel.View() will call RenderChrome + WrapTitled; menuHints dispatcher feeds chrome; titleForState feeds WrapTitled)
  - 07-03 grep-gate authoring: TestChromeASCIIOnly allowlist must be {┌, ─, ┐, │, └, ┘, …} - NOT the rounded variants written into PATTERNS.md
  - 08-header-info-panel (will inflate the 38-col placeholder reserved by InfoPanelPlaceholderStyle)
  - 09-keybinding-discoverability (will exercise per-tuple goldens through this composer)
  - 10-theming-accessibility (logo severity coupling will flip RenderChrome's logoStatus arg)

# Tech tracking
tech-stack:
  added:
    - github.com/charmbracelet/x/ansi (ansi.Truncate) - already a direct dep since Phase 6; no go.mod churn
  patterns:
    - String-splice column overlay on a rendered border line as a community-standard lipgloss-v2 substitute for the missing native border-title API
    - ANSI-aware rune walk that skips SGR sequences so styled border lines (BorderForeground) are spliced without disturbing the SGR wrapper
    - Outer-dimension envelope contract for bordered styles: WrapTitled(w, h) passes (w, h) straight through, NOT (w-2, h-2) - lipgloss v2 Style.Width.Height already include the border in the outer dimension
    - InfoPanelPlaceholderStyle as package var (vs inline NewStyle) to satisfy the Plan 3 TestViewNoNewStyle grep-gate even though that gate scans only AppModel.View()

key-files:
  created:
    - internal/ui/chrome.go (197 LOC) - RenderChrome, WrapTitled, overlayTitle, spliceRenderedLine
    - internal/ui/chrome_test.go (204 LOC) - 16 subtests covering all three composer functions
  modified:
    - internal/ui/styles.go (+8 LOC) - InfoPanelPlaceholderStyle package var appended to the existing var block

key-decisions:
  - "WrapTitled passes outer (width, height) directly to TitledBorderStyle.Width.Height - lipgloss v2 contract is OUTER-includes-border, so the plan's Width(w-2).Height(h-2) recipe was a double-subtraction bug surfaced by TestWrapTitled outer-dimensions subtest. Production fix landed within the Plan 2 boundary; doc-comment now records the v2 contract explicitly."
  - "Test contract asserts ┌/┐ corner glyphs (NormalBorder reality) instead of the ╭/╮ rounded glyphs drawn in UI-SPEC visual sketches. lipgloss v2's NormalBorder() emits square corners ┌─┐│└┘ - empirically verified via standalone probe against charm.land/lipgloss/v2. The rounded sketches in UI-SPEC and PATTERNS predate Plan 1's empirical rendering test."
  - "spliceRenderedLine includes ANSI SGR pass-through despite NormalBorder top-line containing no embedded SGR sequences when BorderForeground is uniform. The defensive code is dead-but-safe today; protects against future palette changes that might wrap individual border chars in SGR pairs."
  - "Fixed const minTitledWidth=4 / minTitledHeight=3 clamp lives at the WrapTitled boundary, not inside overlayTitle - keeps the splice helper a pure transform on already-rendered input and lets the composer carry all envelope-defense logic."
  - "InfoPanelPlaceholderStyle declared rather than inline NewStyle inside RenderChrome - Pitfall 1 mitigation. Even though the Plan 3 grep-gate only scans AppModel.View(), the discipline is package-var-first so future migrations don't have to revisit this file."
  - "overlayTitle source-revision citation block kept verbatim from the research finding: notes that CONTEXT.md D-14 cited soft-serve as the reference, that 07-RESEARCH.md §1 verified soft-serve main HEAD does NOT contain the pattern, and points future readers to the research document for the full gap analysis. This citation block is the closure deliverable for the STATE.md 'Phase 7 research pass on overlayTitle' pending todo."

patterns-established:
  - "Pattern: NormalBorder corner glyphs are square (┌┐└┘) not rounded (╭╮╰╯) - any test or grep-gate allowlist that touches NormalBorder output must use ┌─┐│└┘ as the canonical glyph set"
  - "Pattern: Internal test package (package ui, not package ui_test) is the right call for any sub-model file that needs to verify unexported helpers - mirrors how internal/app/layout_test.go uses package app to access findRepoRoot etc."
  - "Pattern: lipgloss v2 Style.Width(W).Height(H) on a bordered style sets the OUTER rendered dimension - if you want N content rows you ask for Height(N+2) at the inner level OR you accept that the requested dim IS the outer. WrapTitled chose the latter for caller-friendliness."
  - "Pattern: ansi-aware column splice via rune walk + SGR pass-through state machine - 30 LOC, suitable for any future first-line modification (header timestamps, badge overlays, etc.)"

requirements-completed: [UI-06]

# Metrics
duration: 5min
completed: 2026-04-27
---

# Phase 7 Plan 2: Chrome Composer + WrapTitled + overlayTitle Summary

**Three composition primitives - RenderChrome (6-row JoinHorizontal band), WrapTitled (NormalBorder titled wrapper), and overlayTitle (string-splice with ANSI-aware column tracking) - shipped pure, fully unit-tested (16 subtests), and ready for Plan 3's AppModel.View() rewrite. Closes the STATE.md overlayTitle research gap with a tested deliverable.**

## Performance

- **Duration:** ~5 min
- **Started:** 2026-04-27T12:11:54Z
- **Completed:** 2026-04-27T12:16:47Z
- **Tasks:** 2
- **Files created:** 2
- **Files modified:** 1
- **LOC delta:** +409 (production: +197 chrome.go, +8 styles.go; tests: +204 chrome_test.go)
- **New tests:** 16 subtests across 3 test functions - all green on first run after the WrapTitled fix

## Accomplishments

- Shipped `internal/ui/chrome.go` with the three Phase 7 composition primitives - `RenderChrome`, `WrapTitled`, `overlayTitle` - plus the unexported `spliceRenderedLine` helper. The file imports `charm.land/lipgloss/v2`, `github.com/charmbracelet/x/ansi`, and `internal/keys` only - no `strings` dependency leaked beyond the already-needed `IndexByte`/`Builder` usage.
- Closed the STATE.md "Phase 7 research pass on overlayTitle" pending todo with a tested deliverable. The `overlayTitle` doc comment cites `07-RESEARCH.md §1` and the soft-serve revision audit verbatim - future readers see the full gap-analysis trail.
- Added `InfoPanelPlaceholderStyle = lipgloss.NewStyle().Width(38).Height(6)` to `internal/ui/styles.go` (Pitfall 1 mitigation per D-22 prep). `RenderChrome` consumes this package var instead of inlining `lipgloss.NewStyle()` - even though the Plan 3 `TestViewNoNewStyle` grep-gate only scans `AppModel.View()`, package-var-first is the established Phase 7 discipline.
- Verified empirically that `lipgloss.v2`'s `Style.Width(W).Height(H)` on a bordered style sets the OUTER dimensions - caught the plan recipe's `Width(width-2).Height(height-2)` double-subtraction bug. WrapTitled now passes `(w, h)` straight through; outer envelope is now exactly what callers request.
- Verified empirically that `lipgloss.NormalBorder()` emits SQUARE corner glyphs `┌┐└┘` - NOT the rounded `╭╮╰╯` drawn in UI-SPEC visual sketches and PATTERNS test snippets. Updated test contract accordingly; recorded forward-deviation note for Plan 3's `TestChromeASCIIOnly` grep-gate allowlist.
- Zero AppModel changes; existing 148-test suite (Plan 1 + earlier phases) remains byte-identical green at every commit. AppModel resize goldens unchanged.

## Task Commits

Each task was committed atomically. Both tasks are tagged `tdd="true"` in the plan; the production-first ordering is intentional because Task 2's tests verify the algorithm contracts that Task 1 establishes.

1. **Task 1: Create internal/ui/chrome.go + InfoPanelPlaceholderStyle in styles.go** - `6adf629` (feat)
   - chrome.go (initial draft with the plan's `Width(width-2).Height(height-2)` recipe)
   - styles.go (+8 LOC: InfoPanelPlaceholderStyle var)
   - All Task 1 grep-gates passed: zero `lipgloss.NewStyle(`, zero banned border types, soft-serve + 07-RESEARCH.md citations present, ansi.Truncate referenced, TitledBorderStyle referenced 4 times.

2. **Task 2: Create internal/ui/chrome_test.go (16 subtests) + fix WrapTitled height math** - `860a6df` (test)
   - chrome_test.go (16 subtests across 3 test functions)
   - chrome.go (Rule 1 bug fix: WrapTitled outer-dimension math, surfaced by `TestWrapTitled/outer_dimensions_match` failing on rendered-vs-requested height)

_Note on TDD ordering: Task 1 wrote production code first; Task 2's tests then ran against it as the GREEN gate. Two of the 16 subtests initially RED-failed and revealed real bugs (corner-glyph contract divergence + height-math double-subtraction); both were fixed atomically inside the Task 2 commit. This preserves the invariant that the production code on each commit boundary is correct and tested._

## Files Created/Modified

- `internal/ui/chrome.go` (new, 197 LOC) - Three exported composition primitives + one unexported splice helper:
  - `RenderChrome(hints []keys.MenuHint, logoStatus LogoStatus, width int) string` - 6-row JoinHorizontal of info-panel placeholder (38 cols) + menu (residual) + logo (25 cols)
  - `WrapTitled(title, body string, width, height int) string` - NormalBorder titled wrapper at exact outer (width, height) with title injected at column 2 of the top border line via overlayTitle
  - `overlayTitle(rendered, title string) string` (unexported) - string-splice helper with empty-title bypass, single-line bypass, narrow-line bypass, and ansi.Truncate truncation for overlong titles
  - `spliceRenderedLine(line string, startCol, endCol int, replacement string) string` (unexported) - rune walk that skips ANSI SGR sequences and replaces the visible-column range [startCol, endCol)
  - Constants: `infoPanelWidth = 38`, `logoWidth = 25`, `minTitledWidth = 4`, `minTitledHeight = 3`
  - Package doc cites `07-RESEARCH.md §1` for the overlayTitle community-standard pattern and the soft-serve gap analysis (closure deliverable for STATE.md's Phase 7 research pending todo)

- `internal/ui/chrome_test.go` (new, 204 LOC) - Internal test package (`package ui`) so unexported `overlayTitle` and `spliceRenderedLine` are reachable. Three test functions, 16 subtests:
  - `TestOverlayTitle_PreservesCornersAndWidth` (7 subtests): top-left ┌, top-right ┐, width unchanged, overlong-truncate-with-ellipsis, empty-title bypass, single-line bypass, narrow-line bypass
  - `TestWrapTitled` (5 subtests): title at col 2, dimension clamp, empty-title border-only first line, NormalBorder-only (no thick/double), outer dims match requested
  - `TestRenderChrome` (4 subtests): exactly 6 rows at widths 40/80/120/200, logo art presence, menu mnemonics, info-panel 38-col blank prefix
  - Helper: `firstLine(s)` - one-line substring up to first '\n', avoids repeating `strings.IndexByte` boilerplate
  - All subtests use `ansi.Strip` for structural assertions to avoid SGR escape sequence interference

- `internal/ui/styles.go` (+8 LOC inserted into existing var block) - `InfoPanelPlaceholderStyle` package var:
  ```go
  // InfoPanelPlaceholderStyle reserves the 6-row x 38-col top-left area
  // of the chrome for Phase 8's header info panel (D-16, Pitfall 1 mitigation).
  // Phase 7 renders the empty string into this style so lipgloss.Height
  // returns exactly 6 and JoinHorizontal alignment is preserved.
  InfoPanelPlaceholderStyle = lipgloss.NewStyle().Width(38).Height(6)
  ```

## Decisions Made

- **WrapTitled outer-dimension contract**: The plan recipe (`Width(width-2).Height(height-2)`) was a double-subtraction bug because lipgloss v2's `Style.Width(W).Height(H)` on a bordered style ALREADY sets the OUTER rendered dimensions (border included). Empirically verified via standalone probe at `charm.land/lipgloss/v2` v2.0.3: `NewStyle().Border(NormalBorder()).Padding(0, 1).Width(20).Height(5).Render("body")` produces a string with `lipgloss.Height` exactly 5 and first-line width exactly 20. Fix landed at the boundary: WrapTitled now passes `(w, h)` straight through.
- **NormalBorder corner-glyph reality vs UI-SPEC sketches**: UI-SPEC §"Titled border" visual sketch (07-UI-SPEC.md lines 453-465) and PATTERNS test snippets drew rounded `╭─ Files (12) ────╮` corners. Empirical render at lipgloss v2 NormalBorder produces SQUARE `┌─ Title ────┐`. The tests assert reality (`┌`, `┐`); UI-SPEC docs are visual reference only and not enforced. Plan 3 must update its `TestChromeASCIIOnly` allowlist accordingly (see "Forward Deviations for Plan 3" below).
- **Internal test package for chrome_test.go**: `package ui` (NOT `package ui_test`) — required to reach unexported `overlayTitle` and `spliceRenderedLine`. This breaks the `ui_test`-everywhere convention used by `filelist_test.go`, `diff_test.go`, etc. Justified because the alternative (exporting `OverlayTitle` and `SpliceRenderedLine`) would expand the package's public API surface unnecessarily. Mirrors the `internal/app/layout_test.go` pattern that uses `package app` for the same reason.
- **InfoPanelPlaceholderStyle as package var**: Even though Plan 3's `TestViewNoNewStyle` grep-gate scans only `internal/app/model.go`'s `View()` method body, the chrome composer keeps package-var-first discipline. Costs 8 LOC of styles.go; gains future-proof against any reviewer who decides the grep-gate scope should expand to `internal/ui/*.go`.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed WrapTitled outer-dimension double-subtraction**
- **Found during:** Task 2, when `TestWrapTitled/outer_dimensions_match_requested_width_and_height` failed asserting `len(lines)==5` got 3, and `lipgloss.Width(lines[0])==20` got the configured size minus 2
- **Issue:** The plan's recipe specified `TitledBorderStyle.Width(width - 2).Height(height - 2).Render(body)`. In lipgloss v2 a bordered style's `Style.Width(W).Height(H)` sets the OUTER rendered size (border included), so requesting outer 20×5 produced an actual 18×3 box - off by 2 in both axes.
- **Fix:** Removed the `-2` subtractions. WrapTitled now passes `(w, h)` directly to `TitledBorderStyle.Width(width).Height(height).Render(body)`. Updated the function doc comment to document the lipgloss v2 outer-dim contract for future readers.
- **Files modified:** `internal/ui/chrome.go`
- **Verification:** All 16 subtests green; `TestWrapTitled/outer_dimensions_match_requested_width_and_height` now asserts the box is exactly 20×5 as requested.
- **Committed in:** `860a6df` (Task 2 commit)

**2. [Rule 1 - Spec] Updated test corner-glyph assertions to NormalBorder reality**
- **Found during:** Task 2, when `TestOverlayTitle/preserves_top-left_corner` and `preserves_top-right_corner` failed - the actual first line was `┌─ Title ────┐` (SQUARE corners) but the test asserted `╭` and `╮` (ROUNDED).
- **Issue:** UI-SPEC visual sketches (07-UI-SPEC.md §"Titled border" lines 453-465) and PATTERNS test snippets (07-PATTERNS.md lines 421-450) drew rounded corners on what is supposed to be a `NormalBorder()` box. UI-SPEC and PATTERNS are documentation; the empirical reality is that `lipgloss.NormalBorder()` produces square corners `┌─┐│└┘` per upstream lipgloss v2.0.3.
- **Fix:** Updated test assertions to expect `┌`/`┐`. Added a comment block in chrome_test.go explaining the divergence so a future reader doesn't get confused. The `empty title renders only border chars and spaces` subtest's allowed-rune set was also updated to `{┌, ┐, ─, ' '}`.
- **Files modified:** `internal/ui/chrome_test.go`
- **Verification:** All 7 overlayTitle subtests + 5 WrapTitled subtests green.
- **Committed in:** `860a6df` (Task 2 commit)

---

**Total deviations:** 2 auto-fixed (1 production bug, 1 test/spec alignment)
**Impact on plan:** Both fixes were within Task 2 scope. Public API of `chrome.go` unchanged (`WrapTitled` signature is identical, only the internal `Width/Height` math fixed). The corner-glyph adjustment is a documentation-vs-implementation reconciliation; no breaking change to any caller because Plan 2 has no callers yet.

## Forward Deviations for Plan 3

Two plan-future-impact deviations Plan 3's executor must be aware of:

1. **TestChromeASCIIOnly allowlist must use SQUARE corner glyphs.** PATTERNS.md line 466 currently shows the allowlist as `'─': true, '│': true, '╭': true, '╮': true, '╰': true, '╯': true, '…': true`. The correct Phase 7 allowlist (matching what `lipgloss.NormalBorder()` actually emits) is `'─': true, '│': true, '┌': true, '┐': true, '└': true, '┘': true, '…': true`. Plan 3 author MUST update this allowlist before running the grep-gate or it will false-flag legitimate NormalBorder output.

2. **WrapTitled body-content envelope is `(w-4) × (h-2)`, not `(w-2) × (h-2)`.** Plan 3's `AppModel.View()` rewrite calls `WrapTitled(title, body, w, h)` with `w, h := bodyDims(m)`. Sub-models render their body to the inner content envelope. The horizontal Padding(0, 1) consumes 1 cell on each side of the content area inside the border, so:
   - Outer width `w` → border (2 cells) + padding (2 cells) + content area `(w-4)` cells per row
   - Outer height `h` → border (2 cells) + content area `(h-2)` rows
   Sub-models that resize their content area should use `bodyDims` then subtract `(4, 2)` for actual writable cells. This is consistent with Phase 6's bodyDims contract; just be aware the inner content width is 4 less than outer (not 2).

## Issues Encountered

None blocking. The two auto-fixes were resolved in-task within the Task 2 commit boundary.

## Verification Results

- `go build ./...` exits 0 (both commits)
- `go vet ./internal/ui` exits 0 (both commits)
- `go test ./internal/ui -run 'TestOverlayTitle_PreservesCornersAndWidth|TestWrapTitled|TestRenderChrome' -count=1 -v` passes all 16 subtests:
  - 7/7 in TestOverlayTitle_PreservesCornersAndWidth
  - 5/5 in TestWrapTitled
  - 4/4 in TestRenderChrome
- `go test ./... -count=1` exits 0 - no regression in any package, AppModel resize goldens (Phase 6) byte-identical green
- `go.mod` / `go.sum` unchanged (charm.land/lipgloss/v2/table and github.com/charmbracelet/x/ansi were already direct deps)
- All Task 1 acceptance criteria gates green:
  - ✓ `func WrapTitled` count: 1
  - ✓ `func overlayTitle` count: 1
  - ✓ `func RenderChrome` count: 1
  - ✓ `func spliceRenderedLine` count: 1
  - ✓ `lipgloss.NormalBorder()` references in chrome.go: 0 (inherited via TitledBorderStyle)
  - ✓ Banned border types (RoundedBorder|ThickBorder|DoubleBorder|HiddenBorder|FocusedBorder|UnfocusedBorder): 0
  - ✓ `lipgloss.NewStyle(` calls in chrome.go: 0
  - ✓ `soft-serve` mention in chrome.go: 3 (citation block intact)
  - ✓ `07-RESEARCH.md` mention in chrome.go: 3 (research-doc cite intact)
  - ✓ `ansi.Truncate` reference: 2
  - ✓ `InfoPanelPlaceholderStyle = ` in styles.go: 1
- All Task 2 acceptance criteria gates green:
  - ✓ chrome_test.go declares `package ui` (internal — required for unexported helper access)
  - ✓ TestOverlayTitle_PreservesCornersAndWidth: 1
  - ✓ TestWrapTitled: 1
  - ✓ TestRenderChrome: 1
  - ✓ `t.Run(` count: 16 (≥15 expected)
  - ✓ `ansi.Strip` references in tests: 10 (multiple, used throughout)
  - ✓ `charm.land/lipgloss/v2` import in tests: 1

## Threat Surface Scan

No new trust boundaries introduced. Per the plan's `<threat_model>`:
- T-7-02-01 (Information Disclosure / "Detail: <filename>" title): mitigated — filenames are repo-relative per Phase 6 invariant; titles flow through overlayTitle as already-sanitized strings; no path expansion happens here
- T-7-02-02 (DoS / massive WrapTitled dimensions): accepted — `lipgloss.Width/Height` are O(n) over input bytes; bubbletea v2 bounds terminal dimensions at the OS layer
- T-7-02-03 (Tampering / overlayTitle splice math off-by-one): mitigated — `TestOverlayTitle_PreservesCornersAndWidth` Test "width unchanged" verifies width preservation exactly; six other subtests cover edges (empty, single-line, narrow, overlong)
- T-7-02-04 (Information Disclosure / overlayTitle citation comment): accepted — the citation cites a public research document and a public soft-serve revision; no secret data disclosed

No `threat_flag` needed - no new network endpoints, auth paths, file access, or schema changes at trust boundaries.

## TDD Gate Compliance

Both tasks declared `tdd="true"`. Gate sequence:

- **Task 1 RED gate:** Production-code-first (the chrome.go file). Acceptance criteria treats the code's existence + grep-gates as the gate. RED phase was implicit: before commit `6adf629`, `internal/ui/chrome.go` did not exist; `go test` against any test file referencing `WrapTitled`/`overlayTitle`/`RenderChrome` would fail to compile. This is the build-error RED phase that Plan 1 also used (07-01-SUMMARY.md decision: "TDD red phase via build errors").
- **Task 1 GREEN gate:** `go build ./...` exits 0 after `6adf629` lands.
- **Task 2 RED gate:** chrome_test.go added; first run had 4 failing subtests (2 corner-glyph + 2 height-math) revealing real bugs. This is the canonical RED phase.
- **Task 2 GREEN gate:** WrapTitled fix + corner-glyph test alignment landed inside the same `860a6df` commit. All 16 subtests green.

Commit messages reflect the gate sequence: `feat(07-02): add chrome composer ...` then `test(07-02): add chrome_test.go (16 subtests) + fix WrapTitled height math`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

**Chrome composer ready. Plan 3 may compose `AppModel.View()` as `JoinVertical(RenderChrome(...), "", WrapTitled(title, body, w, h), statusBar)` and pass `bodyDims(m)` as the (w, h) envelope to `WrapTitled`.**

Plan 3's executor consumes:
- `ui.RenderChrome(hints []keys.MenuHint, logoStatus LogoStatus, width int) string` — call with `m.menuHints()`, `ui.LogoInfo`, `m.width`
- `ui.WrapTitled(title, body string, width, height int) string` — call with `m.titleForState()`, sub-model body output, `bodyDims(m)`'s width and height
- `ui.InfoPanelPlaceholderStyle` (if any code path needs to reserve the same 38×6 envelope outside RenderChrome — currently RenderChrome is the only consumer)
- The empirical contract that NormalBorder corners are SQUARE — Plan 3's `TestChromeASCIIOnly` allowlist must reflect this

Sub-model body content envelope: outer `(w, h)` from `bodyDims` minus 4 cells horizontal (2 border + 2 padding) and 2 cells vertical (2 border, no vertical padding) = inner content area `(w-4) × (h-2)`. Sub-models that already use `bodyDims` are unaffected at the API level — they just need to know that 4 fewer columns than `bodyDims.w` are writable inside the titled box.

No blockers.

## Self-Check: PASSED

Verified post-write:

- ✓ `internal/ui/chrome.go` exists (FOUND)
- ✓ `internal/ui/chrome_test.go` exists (FOUND)
- ✓ `internal/ui/styles.go` modified with InfoPanelPlaceholderStyle (FOUND)
- ✓ Commit `6adf629` exists in `git log` (Task 1)
- ✓ Commit `860a6df` exists in `git log` (Task 2)
- ✓ All 16 subtests pass (`go test ./internal/ui -run 'TestOverlayTitle_PreservesCornersAndWidth|TestWrapTitled|TestRenderChrome' -count=1`)
- ✓ Full project test suite passes (`go test ./... -count=1`)
- ✓ Source-revision citation present (3 grep matches for `soft-serve` + 3 for `07-RESEARCH.md` in chrome.go)

---

*Phase: 07-chrome-skeleton*
*Completed: 2026-04-27*
