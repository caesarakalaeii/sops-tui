---
gsd_state_version: 1.0
milestone: v1.0
milestone_name: — Functional Core
status: executing
stopped_at: Phase 7 Plan 02 complete (chrome composer + WrapTitled + overlayTitle research gap closed)
last_updated: "2026-04-27T12:16:47Z"
last_activity: 2026-04-27 -- Phase 07 Plan 02 complete (RenderChrome + WrapTitled + overlayTitle, 16 unit subtests, STATE overlayTitle research todo closed)
progress:
  total_phases: 7
  completed_phases: 6
  total_plans: 22
  completed_plans: 21
  percent: 95
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-23)

**Core value:** Developers can manage all their SOPS-encrypted secrets from a single terminal interface without remembering CLI flags or writing shell scripts.
**Current focus:** Phase 07 — chrome-skeleton

## Current Position

Milestone: v1.1 — k9s visual parity
Phase: 07 (chrome-skeleton) — EXECUTING
Plan: 3 of 3 (Plans 01 + 02 complete)
Status: Executing Phase 07
Last activity: 2026-04-27 -- Phase 07 Plan 02 complete (RenderChrome + WrapTitled + overlayTitle, 16 unit subtests, STATE overlayTitle research todo closed)

Progress (v1.1 only): [████░░░░░░] 27% (1/6 phases complete, 4/15 plans complete)

## Performance Metrics

**Velocity:**

- Total plans completed (all milestones): 17 (v1.0)
- Average duration: -
- Total execution time: 0 hours

**By Phase:**

| Phase | Milestone | Plans | Total | Avg/Plan |
|-------|-----------|-------|-------|----------|
| 01 Foundation | v1.0 | 4 | - | - |
| 02 Read Loop | v1.0 | 3 | - | - |
| 03 Write Loop | v1.0 | 3 | - | - |
| 04 Clipboard & Git | v1.0 | 3 | - | - |
| 05 Power Features | v1.0 | 4 | - | - |
| 06 Layout Groundwork | v1.1 | 2 (planned) | - | - |
| 07 Chrome Skeleton | v1.1 | 3 (planned) | - | - |
| 08 Header Info Panel + Crumb Chips | v1.1 | 3 (planned) | - | - |
| 09 Keybinding Discoverability | v1.1 | 2 (planned) | - | - |
| 10 Theming + Accessibility | v1.1 | 3 (planned) | - | - |
| 11 Regression + Perf Gates | v1.1 | 2 (planned) | - | - |

**Recent Trend:**

- Last 5 plans: v1.0 Phase 5 (recipient management, health checks) — all completed
- Trend: transitioning from functional-core milestone to visual-parity milestone

*Updated after each plan completion*
| Phase 06 P01 | 8m | 3 tasks | 7 files |
| Phase 06 P02 | 8m | 2 tasks | 8 files |
| Phase 07 P01 | 6m | 3 tasks | 7 files |
| Phase 07 P02 | 5m | 2 tasks | 3 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.

**Recent decisions affecting v1.1 work:**

- Roadmap v1.1: Phase 6 (layout arithmetic groundwork + testing harness) is non-skippable and must land before any chrome rendering — prevents 15 SetSize call-sites needing a second audit pass after chrome merges (Pitfall 1)
- Roadmap v1.1: No new runtime dependencies — every chrome primitive maps to existing `charm.land/lipgloss/v2`, `charm.land/bubbles/v2`, or `goccy/go-yaml`; go.mod stays unchanged
- Roadmap v1.1: User-facing skin YAML loader deferred to v2 (THM-01/02/03) — v1.1 ships a single k9s-tuned default palette; skin loader and fsnotify hot-reload add fail-closed failure modes (Pitfall 4) not worth the milestone risk
- Roadmap v1.1: Immediate-mode chrome composition in `AppModel.View()` — rejected k9s's retained-mode observer pattern (`StackPushed`/`StackPopped`, style listeners); hints dispatched via enum switch on `(state, recipientAction, IsSearchActive)`
- Roadmap v1.1: 6 phases derived from the SUMMARY.md "Implications for Roadmap" section; Phase 10 absorbs what SUMMARY called Phases 9+10 (logo severity + palette + structural skin prep) because the user-facing skin loader is v2; a dedicated Phase 11 handles regression/perf gates and terminal compat sweep

**Carried forward from v1.0:**

- Roadmap: Security groundwork (WithAltScreen, tea.Cmd pattern) locked to Phase 1 — must ship before any secret-display code
- Roadmap: Clipboard signal handler wired to Phase 4 alongside git (both need OS-level signal awareness)
- Roadmap: Recipient management deferred to Phase 5 — highest-risk multi-file operation
- [Phase 02-read-loop]: MetadataContent defined as display-only struct independent of parser.SopsMetadata to avoid cross-plan build dependency during Wave 1 parallel execution
- [Phase 02-read-loop]: sahilm/fuzzy promoted to direct dependency in go.mod as SearchModel imports it directly
- [Phase 02-read-loop]: Metadata overlay opened synchronously (parser.ParseFile inline on i keypress) to keep state machine simple
- [Phase 02-read-loop]: Esc priority chain: search deactivation > overlay close > navigate back (matches k9s behavior)
- [Phase 02-read-loop]: ItemCount() returns len(allItems) not len(list.Items()) so status bar count is stable during search
- [Phase ?]: Phase 06 Plan 01: landed layout helpers + LIVE UI-18 grep-gate + testutil harness + v1.0 BenchmarkAppView baseline, zero call-site migrations (Plan 2 migrates atomically)
- [Phase ?]: Phase 06 Plan 01: pointer-variant outlier line number 1799 corrected to 1828 in plan2MigrationAllowlist after Task 1's 29-line helper insertion shifted model.go tail — Plan 2 must use 1828
- [Phase ?]: Phase 06 Plan 01: RequireGoldenColors factored around internal missingColors helper because Go subtests cannot catch propagated failure and testing.TB is sealed — public signature preserved, testability restored
- [Phase 06 Plan 02]: 17-site migration to bodyDims + plan2MigrationAllowlist deletion landed as one atomic commit (79ec449) per D-14; resize goldens + resize_test.go + .gitattributes a second commit (e05165d). Full suite green.
- [Phase 06 Plan 02]: UI-17 (bodyDims single source of truth) and UI-18 (grep-gate without allowlist crutch) CLOSED. UI-19 (golden harness) closed by Plan 01 and exercised by Plan 02's four TestResize_<WxH> tests.
- [Phase 06 Plan 02]: View() outlier collapsed — the plan authorised keeping `statusBarH := lipgloss.Height(statusBar)` if downstream code referenced it; inspection confirmed no other View code reads `statusBarH`, so the local was removed entirely (3 lines out, 1 in).
- [Phase 06 Plan 02]: Resize goldens show `No SOPS files found` empty-state at all four sizes (NewAppModel with "" sopsYamlPath) — no host paths, no timestamps, no env leakage. Deterministic and review-safe.
- [Phase 06 Plan 02]: Manual UAT per D-15 (cycle through all sub-views at 40x12 and 200x60) deferred to `/gsd-verify-work` — beyond executor scope (requires real terminal). Automated resize coverage is the empty-state path at four sizes.
- [Phase 07 Plan 01]: Three primitives shipped additively with zero AppModel changes — internal/keys/hints.go (MenuHint/Hinter/HintsFromBindings + 5 inline hint-set vars), internal/ui/logo.go (LogoStatus enum + 6-row Candidate A art + RenderLogo), internal/ui/menu.go (RenderMenu via lipgloss/v2/table column-major fill). Eight new package-level style vars appended to internal/ui/styles.go (MenuKeyStyle, MenuDescStyle, MenuCellStyle, LogoStyleInfo/Warn/Error, TitledBorderStyle, TitleLabelStyle).
- [Phase 07 Plan 01]: lipgloss/v2/table package was NOT promoted to direct require — already covered by lipgloss/v2 v2.0.3; go mod tidy was a no-op. go.mod/go.sum unchanged.
- [Phase 07 Plan 01]: 24 new tests landed (9 keys + 7 logo + 8 menu) — all green; existing 148-test suite remains byte-identical green at every commit (no AppModel changes means resize goldens unaffected).
- [Phase 07 Plan 01]: Two minor auto-fixes during Task 2 — removed unused lipgloss import from logo.go (Rule 1 build error) and replaced em-dashes (U+2014) in logo.go doc comments with ASCII hyphens (Rule 2 defensive against Plan 3's future TestChromeASCIIOnly grep-gate). Both fixes within Task 2 scope, committed in 0c2f41a.
- [Phase 07 Plan 01]: TDD red-phase achieved via build errors (test file referencing undefined symbols) rather than runtime assertion failure — stronger gate because it proves the symbols don't exist at all. Three commits: f031ffd (Task 1 hints), 0c2f41a (Task 2 logo + styles), 8f6f0a2 (Task 3 menu).
- [Phase 07 Plan 02]: Chrome composer landed — internal/ui/chrome.go (197 LOC) with RenderChrome (6-row JoinHorizontal of info-panel + menu + logo), WrapTitled (NormalBorder titled wrapper), overlayTitle (community-standard string-splice helper closing the STATE.md research-pass todo), spliceRenderedLine (ANSI-aware column splice). 16 unit subtests in chrome_test.go (package ui, internal — accesses unexported helpers). Two atomic commits: 6adf629 (Task 1 production), 860a6df (Task 2 tests + WrapTitled height-math fix).
- [Phase 07 Plan 02]: STATE.md "Phase 7 research pass on overlayTitle" pending todo CLOSED with tested deliverable. The overlayTitle source-revision citation block records that CONTEXT.md D-14 cited charmbracelet/soft-serve as a reference, that 07-RESEARCH.md §1 verified soft-serve main HEAD does NOT contain the pattern (it is community-standard across the bubbletea ecosystem), and that lipgloss v2 has no native border-title API per pkg.go.dev verification.
- [Phase 07 Plan 02]: Two empirical findings landed as auto-fixes during Task 2 — (a) lipgloss v2 Style.Width(W).Height(H) on a bordered style sets OUTER dimensions (not INNER), so the plan's Width(width-2).Height(height-2) recipe was a double-subtraction bug; WrapTitled now passes (w, h) straight through. (b) lipgloss.NormalBorder() emits SQUARE corners ┌┐└┘ not rounded ╭╮╰╯ as drawn in UI-SPEC sketches; tests assert reality. Both fixes within Task 2 commit boundary.
- [Phase 07 Plan 02]: Forward-deviation note for Plan 3 — TestChromeASCIIOnly allowlist must be {┌, ─, ┐, │, └, ┘, …} not {╭, ─, ╮, │, ╰, ╯, …} as written in 07-PATTERNS.md line 466. Empirical contract from NormalBorder() rendering is square corners; PATTERNS.md predates Plan 1's empirical verification. Recorded in 07-02-SUMMARY.md "Forward Deviations for Plan 3".
- [Phase 07 Plan 02]: WrapTitled inner content area is (w-4)x(h-2) not (w-2)x(h-2) — 2 cells horizontal padding (D-13 TitledBorderStyle Padding(0,1) on each side) PLUS 2 cells border. Plan 3 sub-models that derive their writable cell count from bodyDims must subtract 4 horizontal, 2 vertical when rendering inside the WrapTitled envelope. Documented in 07-02-SUMMARY.md "Forward Deviations".

### Pending Todos

- Execute Phase 07 Plan 03 (integration: chromeHeight flip, AppModel.View() rewrite, Hints() on 9 sub-models, dispatcher, grep-gates, bench gate, golden refresh) — Plans 01 + 02 primitives + composer ready as building blocks; remember TestChromeASCIIOnly allowlist must use SQUARE corners ┌─┐│└┘…
- Manual UAT per Phase 06 D-15 — cycle through file list / detail / help / diff / metadata / history / health / recipient form at 40x12 AND 200x60 in a real terminal to confirm no content paints under the status bar at any non-empty sub-view state (deferred from executor; belongs to `/gsd-verify-work`)
- Phase 10 research pass recommended before planning — aggregate health severity classification (is "git dirty" a warn or info? is "stale recipient key" a warn?)

### Blockers/Concerns

**v1.1-specific:**

- Chrome height arithmetic regression risk — Pitfall 1 from PITFALLS.md; mitigation is Phase 6 landing first with grep-gated CI
- Color-profile downsampling on 16-color terminals — paired chip bg/fg collapses to single color on `TERM=xterm`; mitigation is Phase 10 profile detection + 16-color fallback palette
- ~~`overlayTitle` for titled borders~~ CLOSED 2026-04-27 by Phase 07 Plan 02 — community-standard string-splice helper landed in internal/ui/chrome.go with 7-subtest contract; soft-serve main HEAD verified to NOT contain the pattern (07-RESEARCH.md §1 closure)
- Info-panel PII leakage — age fingerprints and file paths appear on screenshots / tmux capture; mitigation is Phase 8's 5-question security review per new field

**Carried forward from v1.0:**

- SOPS stdin editing with `encrypted_regex` — behavior undocumented, may surface during Phase 3 implementation
- Wayland clipboard limitation (needs xclip/xsel) — must document in README before Phase 4 ships
- `lipgloss.AdaptiveColor` hang (issue #1036) — use explicit colors throughout, not adaptive colors
- go-git v6 migration — reassess at Phase 4/5 boundary

## Session Continuity

Last session: 2026-04-27T12:16:47Z
Stopped at: Phase 7 Plan 02 complete (chrome composer + WrapTitled + overlayTitle research gap closed)
Resume file: .planning/phases/07-chrome-skeleton/07-03-PLAN.md
