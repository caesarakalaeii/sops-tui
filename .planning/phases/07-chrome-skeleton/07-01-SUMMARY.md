---
phase: 07-chrome-skeleton
plan: 01
subsystem: ui
tags: [tui, bubbletea, lipgloss, lipgloss-table, chrome-skeleton, keys, ui-primitives, ascii-art, menu, logo]

# Dependency graph
requires:
  - phase: 06-layout-groundwork
    provides: bodyDims/chromeHeight stub layout helpers, package-var styles discipline, RequireGoldenStructure/RequireGoldenColors test harness
  - phase: 01-foundation
    provides: keys.GlobalKeyMap/FileListKeyMap/DetailKeyMap with key.Binding.Help() metadata, internal/ui/styles.go palette and var-block conventions
provides:
  - keys.MenuHint struct, keys.Hinter interface, keys.HintsFromBindings converter
  - five inline hint-set vars (FileListSearchHints, RecipientConfirmHints, BulkReKeyConfirmHints, RecipientListHints, FormatMenuHints) with verbatim UI-SPEC copy
  - ui.LogoStatus enum (Info=0/Warn=1/Error=2), ui.LogoSmall byte-art (6 rows, ASCII-only), ui.RenderLogo
  - ui.RenderMenu (lipgloss/v2/table 2x6 column-major grid with StyleFunc)
  - eight new package-level style vars (MenuKeyStyle, MenuDescStyle, MenuCellStyle, LogoStyleInfo/Warn/Error, TitledBorderStyle, TitleLabelStyle)
affects:
  - 07-02 (chrome composer + WrapTitled + overlayTitle)
  - 07-03 (integration: chromeHeight flip, AppModel.View() rewrite, Hints() on 9 sub-models, dispatcher, grep-gates, bench gate, golden refresh)
  - 08-header-info-panel (will reuse TitledBorderStyle, TitleLabelStyle, info-panel placeholder envelope)
  - 09-keybinding-discoverability (will exercise per-tuple golden matrix on top of these primitives)
  - 10-theming-accessibility (UI-03 logo severity coupling consumes LogoStyleWarn/Error already declared here)

# Tech tracking
tech-stack:
  added:
    - charm.land/lipgloss/v2/table (already transitive via lipgloss v2.0.3 — no go.mod churn)
  patterns:
    - MenuHint contract (Mnemonic/Description/Visible) decoupled from key.Binding for reuse across inline hint sets
    - HintsFromBindings as a pure transform — keymap remains single source of truth, hints are projection
    - Inline hint-set package vars for states with no owning sub-model (Pitfall 3 closure)
    - 6-row ASCII logo as []string package var with iota-enum severity selector
    - lipgloss/v2/table.New().Border*(false) pattern with StyleFunc returning a no-op package var (zero NewStyle in render path)
    - Column-major 2x6 grid fill (i/menuRows = col, i%menuRows = row)
    - Cell composition via package-var style fragments (MenuKeyStyle.Render + MenuDescStyle.Render) instead of NewStyle inside render

key-files:
  created:
    - internal/keys/hints.go
    - internal/keys/hints_test.go
    - internal/ui/logo.go
    - internal/ui/logo_test.go
    - internal/ui/menu.go
    - internal/ui/menu_test.go
  modified:
    - internal/ui/styles.go (8 new vars appended to existing var block)

key-decisions:
  - "Logo art Candidate A locked from Research §5 — 5-row SOPS block figlet plus tui subscript, ASCII-only, ~25 cols"
  - "RenderLogo accepts width param but ignores it in Phase 7 — plumbed for Phase 10 width-responsive layouts (D-02)"
  - "MenuCellStyle declared as no-op lipgloss.NewStyle() package var — reserved for Phase 10 per-column tweaks; Phase 7 inline fragment styling owns accent/fg coloring"
  - "RGB-triplet assertions ('137;180;250') used in tests instead of raw ANSI escapes — stable across lipgloss output and tolerant of profile downsampling"
  - "lipgloss/v2/table package not promoted to direct require in go.mod — already covered by lipgloss/v2 v2.0.3; go mod tidy was a no-op"
  - "TitledBorderStyle uses lipgloss.NormalBorder() not RoundedBorder per UI-15 — intentional departure from ARCHITECTURE.md Pattern 3 sketch (locked by D-13)"

patterns-established:
  - "Pattern: hint converter as pure transform — keys.HintsFromBindings([]key.Binding) → []MenuHint preserves keymap as single source of truth"
  - "Pattern: inline hint-set package vars for stateless modes (search-active, confirm overlays, format menu) where no sub-model owns the keys"
  - "Pattern: ASCII-only chrome with iota-enum severity selector — RenderLogo(LogoInfo|LogoWarn|LogoError, width) plumbing-ready for Phase 10"
  - "Pattern: lipgloss/v2/table render with all Border*(false) toggles + StyleFunc returning no-op package var, fragment styling inside cells"
  - "Pattern: column-major fill index math (col := i / rows; row := i % rows) for fixed-grid menus"
  - "Pattern: RGB-triplet substring assertions in unit tests (color presence without ANSI-escape brittleness)"

requirements-completed: [UI-01, UI-02]

# Metrics
duration: 6min
completed: 2026-04-27
---

# Phase 7 Plan 1: Primitives + Hints Interface Summary

**MenuHint contract, 6-row ASCII SOPS logo, and 2x6 lipgloss/v2/table keybinding menu primitives — three pure renderers plus eight chrome style vars, zero AppModel changes, every commit independently buildable.**

## Performance

- **Duration:** ~6 min
- **Started:** 2026-04-27T12:00:32Z
- **Completed:** 2026-04-27T12:06:27Z
- **Tasks:** 3
- **Files created:** 6
- **Files modified:** 1
- **LOC delta:** +967 (production: +257; tests: +463; styles: +34 inserted)
- **New tests:** 24 (9 keys + 7 logo + 8 menu) — all passing

## Accomplishments

- Shipped `internal/keys/hints.go` with the `MenuHint` / `Hinter` / `HintsFromBindings` triad plus the five UI-SPEC-locked inline hint sets ready for Plan 3's dispatcher.
- Shipped `internal/ui/logo.go` with the locked 6-row Candidate A SOPS+tui art and the three-severity `RenderLogo` selector (Phase 10 wiring already in place).
- Shipped `internal/ui/menu.go` with `RenderMenu` built on `charm.land/lipgloss/v2/table`, all seven `Border*(false)` toggles disabled, column-major fill, visibility filter, and 12-slot cap.
- Extended `internal/ui/styles.go` with eight new package-level style vars (`MenuKeyStyle`, `MenuDescStyle`, `MenuCellStyle`, `LogoStyleInfo/Warn/Error`, `TitledBorderStyle`, `TitleLabelStyle`) — every Phase 7 chrome primitive now resolves through pre-allocated styles.
- Verified zero `lipgloss.NewStyle()` calls inside any render function body in `logo.go`/`menu.go` (Plan 3's `TestViewNoNewStyle` grep-gate will land green on first run).
- Verified zero non-ASCII codepoints anywhere in the three new chrome files (Plan 3's `TestChromeASCIIOnly` will land green; only allowlist-bound runes are present).

## Task Commits

Each task was committed atomically (TDD red→green sequence collapsed into a single feat commit per task since red-phase test failures were build errors, not runtime assertions):

1. **Task 1: Create internal/keys/hints.go with MenuHint, Hinter, HintsFromBindings, and 5 inline hint-set vars** — `f031ffd` (feat)
2. **Task 2: Add 8 Phase 7 style vars to internal/ui/styles.go and create internal/ui/logo.go with RenderLogo** — `0c2f41a` (feat)
3. **Task 3: Create internal/ui/menu.go with RenderMenu (lipgloss/v2/table) and 8 unit tests** — `8f6f0a2` (feat)

_Note: TDD pattern was preserved — each task wrote tests first, observed compile-time RED, then added the production code to land GREEN. The build-error RED phase verified test correctness before implementation existed._

## Files Created/Modified

- `internal/keys/hints.go` (new, 103 LOC) — `MenuHint` struct, `Hinter` interface, `HintsFromBindings` helper, five inline hint-set package vars (`FileListSearchHints`, `RecipientConfirmHints`, `BulkReKeyConfirmHints`, `RecipientListHints`, `FormatMenuHints`) with verbatim UI-SPEC copy
- `internal/keys/hints_test.go` (new, 141 LOC) — 9 unit tests: round-trip with synthetic bindings, nil/empty input safety, real `DefaultFileListKeyMap.ShortHelp()` round-trip, five exact-copy verifications on each inline hint set, `Hinter` interface compile assertion
- `internal/ui/logo.go` (new, 63 LOC) — `LogoStatus` iota enum, `LogoSmall` 6-row Candidate A byte-art (ASCII-only, ~25 cols), `RenderLogo(status, width) string` switching on severity to pick `LogoStyleInfo/Warn/Error`
- `internal/ui/logo_test.go` (new, 92 LOC) — 7 unit tests: six rows exact, ASCII-only rune guard, width 22-26 envelope per row, six-row newline count after ANSI strip, all-severity RGB-triplet check, width=0 no-panic, iota ordering lock
- `internal/ui/menu.go` (new, 91 LOC) — `RenderMenu(hints []keys.MenuHint, width int) string` with column-major fill, visibility filter, 12-slot cap, and `lipgloss/v2/table.New()` configured with all 7 borders off and `StyleFunc` returning `MenuCellStyle`
- `internal/ui/menu_test.go` (new, 230 LOC) — 8 unit tests: non-empty render with bracket-form mnemonic check, column-major fill verification (12-hint pairing matrix), invisible-skip filter, cap-at-12 truncation, nil/empty no-panic, narrow-terminal safety at 40 and 10 cols, ASCII-only body with arrow allowlist, accent RGB triplet on rendered mnemonic
- `internal/ui/styles.go` (modified, +34 LOC inserted into existing `var ( … )` block) — eight new package vars defined verbatim per UI-SPEC §"Phase 7 new style declarations"

## Decisions Made

- **Logo Candidate A locked verbatim** — Used the Research §5 recommendation byte-art exactly. No re-examination of figlet variants; user pre-committed to "SOPS block + tui subscript" direction in DISCUSSION-LOG.md.
- **No `go mod tidy` promotion** — `charm.land/lipgloss/v2/table` resolved cleanly through the existing `lipgloss/v2 v2.0.3` direct require. `go mod tidy` ran as a no-op; `go.mod`/`go.sum` unchanged. Plan 1 acceptance criteria explicitly allowed either outcome.
- **Test color assertions via RGB triplets** — Used substring matches on `"137;180;250"` (ColorAccent), `"249;226;175"` (ColorWarning), `"243;139;168"` (ColorError) instead of raw ANSI escape sequences. Triplets are emitted directly by lipgloss's TrueColor SGR generator and are profile-stable for the configured TrueColor terminal.
- **TDD red phase via build errors** — RED was achieved when the test file referenced symbols not yet defined in the production file. This is a stronger gate than runtime assertion failure because it proves the symbols don't exist at all (not just that they behave incorrectly). Sequence: write test → `go test` fails to compile → write production → `go test` passes.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Removed unused lipgloss import from logo.go**
- **Found during:** Task 2 (logo.go GREEN-phase compile)
- **Issue:** Initial logo.go imported `charm.land/lipgloss/v2` but the implementation uses `LogoStyle*.Render(art)` package vars from styles.go directly — the lipgloss package itself isn't referenced in this file. `goimports`-style unused-import compile error.
- **Fix:** Reduced the import block to just `"strings"`. Styles like `LogoStyleInfo` are package-level vars in the same `ui` package, so no lipgloss import is needed in logo.go.
- **Files modified:** `internal/ui/logo.go`
- **Verification:** `go build ./...` exits 0; logo tests pass.
- **Committed in:** `0c2f41a` (Task 2 commit)

**2. [Rule 2 - Missing Critical (defensive)] Replaced em-dashes with ASCII hyphens in logo.go doc comments**
- **Found during:** Task 2 acceptance-criteria verification
- **Issue:** The plan's stated acceptance check `LC_ALL=C grep -P '[^\x00-\x7F]' internal/ui/logo.go` returned 2 matches. Both were em-dashes (U+2014) in doc-comment prose. The literal acceptance criterion was "no matches"; the matches were not in the `LogoSmall` block (which was already pure ASCII). However, Plan 3 will introduce `TestChromeASCIIOnly` that grep-gates `internal/ui/logo.go` against non-ASCII runes outside an allowlist of `─│╭╮╰╯…` — em-dashes would fail that test.
- **Fix:** Replaced both em-dashes (U+2014) in the package doc comment with ASCII hyphen-minus (U+002D). Body of the file (LogoSmall byte-art) was already 100% ASCII.
- **Files modified:** `internal/ui/logo.go`
- **Verification:** `LC_ALL=C grep -cP '[^\x00-\x7F]' internal/ui/logo.go` returns 0; tests still pass.
- **Committed in:** `0c2f41a` (Task 2 commit)

---

**Total deviations:** 2 auto-fixed (1 unused-import bug, 1 defensive-ASCII fix)
**Impact on plan:** Both deviations were trivial cleanups within Task 2 scope; neither changed any public API or test expectation. Plan acceptance criteria all met. Zero scope creep.

## Issues Encountered

None — all three tasks landed straight through TDD red→green without iteration. Existing test suite (148 tests across 9 packages) remained byte-identical green at every commit.

## Verification Results

- `go build ./...` exits 0 (all commits)
- `go vet ./...` exits 0 (all commits)
- `go test ./... -count=1` exits 0 (all commits) — including `internal/app` resize goldens (Phase 6) which remain byte-identical because AppModel is unchanged
- `go mod verify` exits 0; `go.mod` and `go.sum` unchanged from pre-plan state
- 24 new unit tests added; existing test count unchanged
- Acceptance criteria from `<success_criteria>` block all met:
  - ✓ All 3 tasks executed and committed individually
  - ✓ All 6 new files exist (`hints.go`, `hints_test.go`, `logo.go`, `logo_test.go`, `menu.go`, `menu_test.go`)
  - ✓ `internal/ui/styles.go` contains all 8 new style vars
  - ✓ `go test ./internal/keys ./internal/ui -count=1` exits 0

## Threat Surface Scan

No new trust boundaries introduced. Per the plan's `<threat_model>`:
- T-7-01-01 (Information Disclosure / hint text): accepted — all hint copy is committed source code, no secrets
- T-7-01-02 (Tampering / LogoSmall byte-art): accepted — compile-time constant; Plan 3 grep-gate (`TestChromeASCIIOnly`) will catch smuggled glyphs in CI
- T-7-01-03 (DoS / RenderMenu degenerate widths): mitigated by `TestRenderMenu_NarrowTerminalSafe` (passes at width 40 and 10; lipgloss/v2/table handles width=0 gracefully via the empty-input/no-panic test)

No `threat_flag` needed — no new network endpoints, auth paths, file access, or schema changes at trust boundaries.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

Plan 2 (chrome composer + `WrapTitled` + `overlayTitle` research) may now consume:
- `MenuKeyStyle`, `MenuDescStyle`, `MenuCellStyle` from `internal/ui/styles.go`
- `LogoStyleInfo/Warn/Error` from `internal/ui/styles.go`
- `TitledBorderStyle`, `TitleLabelStyle` from `internal/ui/styles.go`
- `RenderLogo(status, width)` and the `LogoStatus` enum from `internal/ui/logo.go`
- `RenderMenu(hints, width)` from `internal/ui/menu.go`
- `keys.MenuHint`, `keys.Hinter`, `keys.HintsFromBindings`, and all five `keys.*Hints` inline vars from `internal/keys/hints.go`

Wave 1 sequential ordering preserved: Plan 2 will extend the same `var ( … )` block in `internal/ui/styles.go` with `InfoPanelPlaceholderStyle` (per Plan 2 PATTERNS guidance and the 07-02-PLAN modification declaration). No file conflicts pending.

No blockers. Plan 3 prerequisites all in place — `TestChromeASCIIOnly`, `TestChromeNormalBorderOnly`, and `TestViewNoNewStyle` will land green on first run because:
- `logo.go`/`menu.go` are 100% ASCII-only outside the box-drawing/arrow allowlist
- Neither file references `RoundedBorder|ThickBorder|DoubleBorder|HiddenBorder|FocusedBorder|UnfocusedBorder`
- Neither render function calls `lipgloss.NewStyle()` inside its body

## Self-Check: PASSED

Verified post-write:

- ✓ `internal/keys/hints.go` exists (FOUND)
- ✓ `internal/keys/hints_test.go` exists (FOUND)
- ✓ `internal/ui/logo.go` exists (FOUND)
- ✓ `internal/ui/logo_test.go` exists (FOUND)
- ✓ `internal/ui/menu.go` exists (FOUND)
- ✓ `internal/ui/menu_test.go` exists (FOUND)
- ✓ `internal/ui/styles.go` modified with 8 new vars (FOUND)
- ✓ Commit `f031ffd` exists in `git log` (Task 1)
- ✓ Commit `0c2f41a` exists in `git log` (Task 2)
- ✓ Commit `8f6f0a2` exists in `git log` (Task 3)

---

*Phase: 07-chrome-skeleton*
*Completed: 2026-04-27*
