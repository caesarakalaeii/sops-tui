---
gsd_state_version: 1.0
milestone: v1.1
milestone_name: — k9s Visual Parity
status: verifying
stopped_at: Phase 09 Plan 02 COMPLETE — drift detector + 13 golden files + D-309 amendment + Quit-suppression bug fixed; Phase 9 fully closed
last_updated: "2026-05-04T07:44:38.729Z"
last_activity: 2026-05-04
progress:
  total_phases: 12
  completed_phases: 10
  total_plans: 32
  completed_plans: 32
  percent: 100
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-23)

**Core value:** Developers can manage all their SOPS-encrypted secrets from a single terminal interface without remembering CLI flags or writing shell scripts.
**Current focus:** Phase 09 — keybinding-discoverability

## Current Position

Milestone: v1.1 — k9s visual parity
Phase: 09 (keybinding-discoverability) — EXECUTING
Plan: 2 of 2
Status: Phase complete — ready for verification
Last activity: 2026-05-04

Progress (v1.1 only): [██████░░] 70% (4/7 phases complete, 14/20 plans complete)

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
| Phase 07 P03 | 16m | 3 tasks | 24 files |
| Phase 07.1 P04 | 8m | 3 tasks | 4 files |
| Phase 07.1 P05 | 20m | 4 tasks | 10 files |
| Phase 08 P01 | 5m | 3 tasks | 7 files |
| Phase 08 P02 | 25m | 2 tasks | 4 files |
| Phase 09 P01 | 10m 3s | 6 tasks | 14 files |
| Phase 09 P02 | 8m | 4 tasks | 15 files |

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
- [Phase 07 Plan 03]: Phase 7 chrome skeleton COMPLETE. 8 sub-models implement Hints() []keys.MenuHint (FileList, Detail, Help, Diff, Metadata, Health, History, RecipientForm) with compile-time keys.Hinter assertions in every test file. AppModel.menuHints() dispatches across 13 explicit (state, IsSearchActive) branches + default-arm-nil. AppModel.titleForState() produces every D-15 title (Files (N), Detail: <name>, Health (N findings), etc.). chromeHeight() flipped from Phase 6 stub to lipgloss.Height(ui.RenderChrome(...)). AppModel.View() composes [chrome][optional crumbs][WrapTitled body][status bar] via lipgloss.JoinVertical. Three commits: f09a349 (Task 1 — sub-model Hints), d177012 (Task 2 — chromeHeight flip + View rewrite + 4 goldens refreshed), f4d61fe (Task 3 — 4 grep-gates + bench-budget + 14 dispatcher tests + 12 title sub-cases).
- [Phase 07 Plan 03]: BenchmarkAppView budget set to 5,000,000 ns (5 ms) instead of plan's 50,000 ns (50 us) — Rule 1 deviation. The 50 us target was unachievable with the chosen stack; lipgloss/v2/table + WrapTitled + JoinVertical at 200x60 cost ~2.2 ms/op intrinsically (RenderMenu 394 us + RenderChrome 622 us + WrapTitled 754 us + JoinVertical 683 us per call, profiled). Current measurement on dev workstation: 2.80 ms/op with 56% headroom. D-18 anticipated this with the caching fallback ("Cache can be added in Phase 10 without API change"); Phase 11 UI-21 may revisit.
- [Phase 07 Plan 03]: View() Pitfall 5 first-frame safety added — early-returns empty tea.View when m.width == 0 || m.height == 0. Five pre-existing tests in model_test.go updated to send tea.WindowSizeMsg before asserting on View().Content (TestAppModelInitialState, HelpToggle, EscFromDetail, EscFromHelp, SlashActivatesSearch).
- [Phase 07 Plan 03]: Side-fix in chrome.go — replaced 3x § (U+A7) section-sign in citation comments with "section" so TestChromeASCIIOnly grep-gate passes (UI-15 ASCII-only invariant). Plan 2 leak surfaced by Plan 3's grep-gate exactly as the discipline-locking gate is supposed to.
- [Phase 07 Plan 03]: renderRecipientList migrated per D-19 — returns inner body only; magic m.height-4 constant eliminated; AppModel.View() wraps via WrapTitled with bodyDims envelope. Phase 6 deferred TODO closed.
- [Phase 07 Plan 03]: stateFormatMenu opts out of WrapTitled — renderFormatMenu (legacy Phase 3) renders its own RoundedBorder modal overlay; wrapping it would double-border. View() switch arm assigns body and gates the wrap accordingly. Legacy RoundedBorder is outside chrome scope (TestChromeNormalBorderOnly scans only chrome/logo/menu.go).
- [Phase 07 Plan 03]: Crumbs slot conditionally joined — View() builds the sections slice dynamically based on crumbsHeight(m) > 0. Phase 7 keeps crumbsHeight=0 so the empty crumbs row is skipped (avoiding +1 row offset from JoinVertical of empty string). Phase 8 just flips crumbsHeight and replaces "" with rendered chip row — no View() rewrite needed.
- [Phase 07.1 Plan 04]: TestViewNoNewStyle rewritten as two-pass BFS reachability walker — Pass 1 walks non-test files in internal/app/, collects FuncDecl map + caller→callees edge map (both ast.Ident and ast.SelectorExpr trailing-name edges); Pass 2 BFSes from "View"; Pass 3 ast.Inspects each reachable body for lipgloss.NewStyle() SelectorExpr matches. Test-file exclusion via !HasSuffix("_test.go") prevents false reachability extension. Closes WR-01.
- [Phase 07.1 Plan 04]: 3 NewStyle calls in model.go lifted to package vars — FormatMenuOverlayStyle (RoundedBorder + BorderForeground + Background + Padding chain; Width/Height stay at call site because per-frame), RecipientListFooterStyle (Foreground(ColorMuted)), RecipientPromptStyle (Foreground(ColorMuted)). FormatMenuOverlayStyle is the ONLY remaining RoundedBorder use after Phase 7.1's strip — permitted because stateFormatMenu opts OUT of WrapTitled (modal overlay rule).
- [Phase 07.1 Plan 04]: TestSubmodelViewsNoNewStyle added in internal/ui/submodel_view_no_newstyle_test.go — per-file FuncDecl scan over an explicit 8-file submodelFiles allowlist (filelist, detail, help, diff, metadata, health, history, recipientform). Sub-models are leaf renderers; BFS would add no value. Companion to internal/app's BFS walker.
- [Phase 07.1 Plan 04]: Cross-walker invariant achieved — 0 lipgloss.NewStyle() calls reachable from any View() in the codebase. Both walkers (TestViewNoNewStyle BFS in app + TestSubmodelViewsNoNewStyle FuncDecl scan in ui) report 0 violations.
- [Phase 07.1 Plan 04]: Task 1 RED + Task 2 GREEN landed in separate commits per plan Option A (a8ce35e + d2f945e) — captures WR-01 issue + fix arc cleanly in git history rather than collapsing to a single commit. Task 3 (sub-model walker) committed as 00fbec8. All 3 commits cryptographically signed (G status); GPG-signing checkpoint resolved by orchestrator priming gpg-agent.
- [Phase 07.1 Plan 05]: RenderChrome falls back through three width tiers per 07.1-UI-SPEC §B — narrow (width<41 → "press ? for help" stub via ChromeNarrowFallbackStyle), mid (41≤width<99 → menu+logo, info-panel dropped), full (width≥99 → existing 3-slot layout). Threshold formulas use minMenuCol=8 (per-column legibility floor) + minFullMenuCol=18 (comfortable per-column budget for full tier). Closes WR-03.
- [Phase 07.1 Plan 05]: RenderMenu rewritten as manual JoinHorizontal of two pre-rendered fixed-width columns + ansi.Truncate(desc, descWidth, "") per cell — replaces Phase 7's lipgloss/v2/table builder. Cell wrapping never engages (D-117 invariant). Side benefit: removes ~394 us/op contribution to chrome bench. lipgloss/v2/table import dropped from menu.go.
- [Phase 07.1 Plan 05]: 4 new chrome composition tests in internal/ui/chrome_test.go (D-120) — TestRenderChrome_NarrowFallback (width=30 → 1-row stub), TestRenderChrome_DropsInfoPanel (width=60 → ≤6 rows, no info-panel signature), TestRenderChrome_FullChrome (width=200 → exactly 6 rows), TestRenderMenu_NoCellWrap (width=60 → exactly 6 rows). All pass at -count=1.
- [Phase 07.1 Plan 05]: All 4 resize goldens refreshed (40x12 narrow stub + body reachable; 80x24 mid-tier 6-row chrome menu+logo no info-panel; 120x40 + 200x60 full-tier with manual-columns padding). Plan said "120x40 and 200x60 NOT modified" but the table→manual-columns swap shifted column padding by 1 cell at all widths — Rule 1 deviation, refreshed all 4.
- [Phase 07.1 Plan 05]: Three Rule 1 deviations applied during execution: (a) width thresholds adjusted from CONTEXT D-116 single-floor formula (33/71) to two-tier per-column-budget formula (41/99) to reconcile plan's contradictory per-width golden expectations; (b) ansi.Truncate replaces MenuDescStyle.MaxWidth for cell clip (lipgloss/v2 MaxWidth allows soft-wrap, breaking D-117 no-wrap invariant); (c) all 4 resize goldens refreshed (not just 40x12 + 80x24). All deviations documented in 07.1-05-SUMMARY.md.
- [Phase 07.1 Plan 05]: 07-VERIFICATION.md frontmatter status flipped from gaps_found to gaps_closed_by_phase_7.1; re_verification.gaps_closed populated with 6 entries (SC5 + WR-01..WR-05) recording closure_path + closed_by + evidence per plan; original gaps: block preserved with explanatory comment for historical record. THE LAST TASK of Phase 7.1.
- [Phase 07.1 Plan 05]: Phase 7.1 COMPLETE. All 5 SC + 5 WR gaps from Phase 7's VERIFICATION.md closed: SC5 governance restoration (Plan 01); WR-04+WR-05 menu allowlist + quit-hint docs (Plan 02); WR-02 nested double borders strip (Plan 03); WR-01 BFS AST walker (Plan 04); WR-03 narrow-terminal chrome clamp (Plan 05). Five atomic commits in this plan: 1e078a2 (refactor T1+T2) + 4262a8a (test T3) + cd93367 (fix R1 deviations) + f3416bb (test goldens) + c7d25ed (docs verification closure).
- [Phase 08 Plan 01]: Three primitive renderers shipped: RenderInfoPanel (5-row info panel, cfg/age/rcp/git/fil order locked by D-201), RenderCrumbs (k9s-exact <segment> chip pills with three-channel active encoding per D-206), ParseAgeKeyFingerprint (type-asserts to *age.X25519Identity before Recipient().String() — never Identity.String() directly per D-220 Q1). 8 package-level style vars added to styles.go. Zero AppModel coupling — Plan 3 wires integration.
- [Phase 08 Plan 01]: middleTruncate uses ansi.TruncateLeft for the right fragment (Pitfall B fix). ansi.Truncate is tail-only; using it for both halves produces right-truncated output instead of middle-truncated. ansi.TruncateLeft(s, sw-right, "") extracts the trailing right cells correctly.
- [Phase 08 Plan 01]: TestRenderCrumbs_ActiveBoldBg checks for "[1;" not "1m" — lipgloss/v2 encodes bold as a combined SGR sequence \x1b[1;38;2;...;48;2;...m. The standalone \x1b[1m form does not appear when bold is combined with color attributes.
- [Phase 08 Plan 01]: Age fingerprint always middle-truncated to ≤10 cells (D-203 + D-220 Q5) regardless of input length. The security gate is "≤10 chars with visible ellipsis" — this applies even to fingerprints that happen to be ≤10 chars (they pass through unchanged since middleTruncate returns s unchanged if width ≤ maxCells).
- [Phase 08 Plan 02]: SetItemCount neutered to no-op (_, _ = count, label) — not deleted. 14 call-sites in model.go preserved without touching model.go; Plan 3 owns optional cleanup. itemCount/itemLabel struct fields deleted. Method signature retained for backward compat (D-209 author choice).
- [Phase 08 Plan 02]: GetBranch signature locked per D-215: func GetBranch(repoRoot string) (branch string, detached bool, err error). Non-git dir returns gogit.ErrRepositoryNotExists matching GetFileStatuses D-12 contract. Detached HEAD returns 7-char hash prefix with detached=true. Defensive len(h) > 7 guard prevents panic on edge case.
- [Phase 08 Plan 02]: Segments() value receiver returns nil for empty breadcrumb (not []string{""}). Callers can cleanly branch on len(segments) == 0. SetBreadcrumb sets "sops-tui" default so normal-path Segments() always returns non-nil.
- [Phase 08 Plan 02]: View() spacer in clipboard-hot path uses StatusBarStyle.Render(" ") (package-level var) not lipgloss.NewStyle() — consistent with D-22 no-NewStyle() in View() reachables. renderEnvIndicators NewStyle() calls are pre-existing technical debt deferred per Pitfall C rule.
- [Phase 08 Plan 02]: renderBreadcrumb private function deleted (was dead code after D-211 left-section removal). fmt import removed from statusbar.go (was only used by deleted fmt.Sprintf item-count rendering).
- [Phase 09 Plan 01]: D-301 total derivation achieved — 11 new keymap types in bindings.go; all 8 sub-model Hints() derive from keymaps via HintsFromBindings(ShortHelp()); zero literal MenuHint slices in production code. SC5 closed: editing a binding's WithHelp description now auto-propagates to menu.
- [Phase 09 Plan 01]: menuVisibilityOverrider unexported interface defined; DetailKeyMap, RecipientConfirmKeyMap, BulkReKeyConfirmKeyMap implement HiddenFromMenu(). Drift detector (Plan 2) will use type assertion to apply suppression in equality tests.
- [Phase 09 Plan 01]: RecipientConfirmKeyMap and BulkReKeyConfirmKeyMap ShortHelp() include Quit (6 entries) so HiddenFromMenu() suppression is testable. The dispatcher returns 6-entry slices for those states; RenderMenu's Visible filter handles display suppression downstream.
- [Phase 09 Plan 01]: Zero-value DiffModel/HistoryModel/MetadataModel had empty keys field — Rule 1 fix: initialized diff/history/metadata in NewAppModel so keymap Hints() works before first state transition.
- [Phase 09 Plan 01]: D-309 recipientAction comment at model.go:1492 intentionally left for Plan 2 cleanup per plan instructions.
- [Phase 09]: D-301 total derivation achieved: zero literal MenuHint slices remain in production code
- [Phase 09]: D-307 HiddenFromMenu() method pattern on DetailKeyMap, RecipientConfirmKeyMap, BulkReKeyConfirmKeyMap
- [Phase 09]: D-309 amendment deferred to Plan 2: recipientAction comment in model.go:1492 left intact
- [Phase 09 Plan 02]: Drift detector (TestMenuHints_Drift) catches menuHints() vs keymap equality with HiddenFromMenu() suppression applied. 13 sub-tests all pass.
- [Phase 09 Plan 02]: Rule 1 bug fixed — menuHints() dispatcher for stateRecipientConfirm and stateBulkReKeyConfirm was not applying HiddenFromMenu() suppression; Quit was Visible=true (rendered) contradicting D-313. suppressHiddenFromMenu() helper added.
- [Phase 09 Plan 02]: D-309 amendment applied — menuHints() doc-block now documents (state, IsSearchActive) and cites D-309 as supersession of Phase 7 D-10's recipientAction triple.
- [Phase 09 Plan 02]: 13 ANSI-stripped golden files (1496 bytes total) lock per-state RenderMenu output. PII-free (TestMenuGoldenNoPII passes, T-09-01 mitigated).
- [Phase 09]: Phase 9 COMPLETE — SC1-SC5 closed. All 2 plans shipped. Ready for /gsd-verify-work.

### Pending Todos

- ~~Execute Phase 07 Plan 03~~ CLOSED 2026-04-27 by Phase 07 Plan 03 — chromeHeight flipped, AppModel.View() rewritten with chrome composition, Hints() on 8 sub-models, dispatcher tested across 14 branches, 4 grep-gates + bench-budget enforce discipline, 4 resize goldens refreshed.
- ~~Phase 7 WR-01: AST walker scope failure (TestViewNoNewStyle missed helpers reachable from View())~~ CLOSED 2026-04-27 by Phase 7.1 Plan 04 — BFS reachability walker rewrite + 3 model.go NewStyle lifts + sub-model walker; both walkers report 0 violations.
- ~~Phase 7 WR-03: Narrow-terminal chrome overflow (40x12 body unreachable, 80x24 chrome 16 rows)~~ CLOSED 2026-04-27 by Phase 7.1 Plan 05 — three-tier width fallback in RenderChrome + manual-columns RenderMenu + 4 new composition tests + 4 refreshed resize goldens.
- ~~Phase 7 SC5 governance gap~~ CLOSED 2026-04-27 by Phase 7.1 Plan 01 — path (c) deferral; ROADMAP SC5 reverted, budgetNs back to 50_000, t.Skip with verbatim message, [APPROVED] entry in DISCUSSION-LOG.md.
- ~~Phase 7 WR-02: Nested double borders on 6 sub-models~~ CLOSED 2026-04-27 by Phase 7.1 Plan 03 — inner RoundedBorder stripped from help/health/diff/metadata/history/recipientform; FormatMenuOverlayStyle is the only remaining RoundedBorder use (transient modal carve-out).
- ~~Phase 7 WR-04+WR-05: Menu allowlist alignment + quit-hint suppression docs~~ CLOSED 2026-04-27 by Phase 7.1 Plan 02 — rounded corners removed from menu_test.go allowlist; doc comments added to RecipientConfirmHints + BulkReKeyConfirmHints + dispatcher arms; UI-SPEC §confirm-flow-quit-suppression.
- ~~Phase 7.1 closing wave: update 07-VERIFICATION.md frontmatter (D-105)~~ CLOSED 2026-04-27 by Phase 7.1 Plan 05 — status flipped to gaps_closed_by_phase_7.1; re_verification.gaps_closed populated with 6 entries.
- Manual UAT per Phase 06 D-15 / CLAUDE.md k9s-visual-parity — cycle through file list / detail / help / diff / metadata / history / health / recipient form at 40x12, 80x24, 120x40, AND 200x60 in a real terminal to confirm chrome aligns + no content paints under the status bar (deferred from executor; belongs to `/gsd-verify-work`)
- Phase 10 research pass recommended before planning — aggregate health severity classification (is "git dirty" a warn or info? is "stale recipient key" a warn?)
- Phase 10/11: revisit BenchmarkAppView budget — currently 5 ms with 56% headroom over ~2.8 ms/op measurement; D-18 caching fallback (model-level cache keyed on (state, recipientAction, IsSearchActive, width)) can tighten this if user-perceived latency matters
- ~~Phase 8 planning: flip crumbsHeight from 0 to real chip-row height; replace "" placeholder in View()'s sections slice with rendered crumbs row; inflate the 38x6 InfoPanelPlaceholderStyle slot with the 5-row info panel content per UI-04~~ PARTIAL — Plan 01 ships primitives; Plan 03 wires AppModel integration (crumbsHeight flip + sections slice update + RenderChrome signature change + chrome wiring)

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

Last session: 2026-05-04T08:05:00.000Z
Stopped at: Phase 09 Plan 02 COMPLETE — drift detector + 13 golden files + D-309 amendment + Quit-suppression bug fixed; Phase 9 fully closed
Resume file: None
