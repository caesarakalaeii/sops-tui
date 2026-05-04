# Phase 11: Regression + Perf Gates - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in 11-CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-04
**Phase:** 11-regression-perf-gates
**Areas discussed:** Caching shape for 50µs target (SC2), v1.0 regression suite scope (SC1), Terminal compat sweep + alt-screen cleanup (SC3+SC4), "Looks Done But Isn't" sign-off mechanism (SC5)

---

## Caching shape for 50µs target (SC2)

### Q1 — What gets cached?

| Option | Description | Selected |
|--------|-------------|----------|
| Full chrome string (Recommended) | AppModel caches the rendered chrome and crumbs strings; invalidate when any chrome input changes. Body rendered fresh every frame because it's the small slice (~10%). Matches D-18 verbatim. Smallest code change; biggest immediate win because the whole 622µs + 394µs path is amortised to zero on identical-input frames (held-down j/k). | ✓ |
| Per-renderer caches | Each renderer (RenderMenu, RenderChrome, RenderCrumbs) owns its own cache via sync.Map or a model field. Better separation of concerns; harder to reason about overall hit rate. | |
| Allocation hygiene only (no cache) | Replace lipgloss/v2/table with hand-rolled column joiner; pre-build all styles once at construction. Pure runtime savings without invalidation surface. Risk: may not hit 50µs alone — the JoinVertical cost remains. | |
| Full string cache + allocation hygiene | Defense-in-depth. Highest ceiling; biggest plan. | |

**User's choice:** Full chrome string (Recommended)
**Notes:** D-18 verbatim path. Locked → D-501.

### Q2 — Cache key fields?

| Option | Description | Selected |
|--------|-------------|----------|
| D-18 minimum (Recommended) | (state, recipientAction, IsSearchActive, width). The 4 fields D-18 listed verbatim. Highest hit rate; smallest key. Risk: if palette ever mutates (it doesn't in v1.1) the cache returns a stale render. | ✓ |
| Full chrome inputs | (state, recipientAction, IsSearchActive, width, palette, logoStatus, infoPanelData, flashGen). Bulletproof correctness; harder to compute the key cheaply. | |
| Generation counter | Track `chromeGen uint64` on AppModel; bump in any Update branch that mutates a chrome input. Easiest invalidation; every mutation site needs the bump (audit cost). | |

**User's choice:** D-18 minimum (Recommended)
**Notes:** Locked → D-502. Palette is read-only after startup; logoStatus is flash-driven (state-by-flash mutation invalidates anyway); infoPanel is event-refreshed (Phase 8 D-213); flashGen drives status bar not chrome.

### Q3 — Cache wiring?

| Option | Description | Selected |
|--------|-------------|----------|
| Field on AppModel, recomputed in Update (Recommended) | AppModel gains `chromeCache string` + `chromeCacheKey chromeKey` fields. Update returns new model with refreshed cache when chrome-input mutating message lands. View() reads cache. Matches Phase 8 infoPanel cache pattern (Pitfall 15 prescription). | ✓ |
| Lazy compute in View | sync.Map on package-level var, keyed on chromeKey struct. Thread-safe, but introduces global state. Easier wiring (no Update changes) but puts cache outside model contract. | |
| Pointer receiver for chrome only | `(m *AppModel) ensureChromeCache()` called at top of View(). Mutates through pointer. Breaks the value-receiver pattern v1.0 + Phase 7.1 + Phase 10 enforced. Not recommended. | |

**User's choice:** Field on AppModel, recomputed in Update (Recommended)
**Notes:** Locked → D-503. Phase 8 infoPanel cache pattern verbatim.

### Q4 — Perf gate?

| Option | Description | Selected |
|--------|-------------|----------|
| Flip TestBenchmarkAppView_UnderBudget on (Recommended) | Remove `t.Skip` in chrome_test.go:311; test runs `testing.Benchmark(BenchmarkAppView)` and fails when nsPerOp > 50,000. Existing infrastructure; gate is exact match for ROADMAP SC5 wording. Single CI-enforceable line. | ✓ |
| Add micro-benches per renderer + flip the gate | BenchmarkRenderChrome / BenchmarkRenderMenu / BenchmarkRenderCrumbs plus existing BenchmarkAppView. Each gates per-call budget. More granular; helps pinpoint future regressions. More test surface to maintain. | |
| Cache-hit-rate test instead of wall-clock | TestChromeCacheHitRate that simulates 100 frames of held-down j and asserts cache reused 99/100 times. Implementation-tied; correctness-focused; survives Go version bumps. Doesn't directly enforce 50µs SC. | |
| All three | Flip gate + cache-hit-rate test + micro-benches. Maximum confidence; heaviest test footprint. | |

**User's choice:** Flip TestBenchmarkAppView_UnderBudget on (Recommended)
**Notes:** Locked → D-504. Plan also adds TestChromeCache_HitRateAtSteadyState per D-505 (proves the cache is wired) — derived during the wrap-up summary, not asked as a separate question.

### Q5 — Continue or move on?

**User's choice:** Next area

---

## v1.0 regression suite scope (SC1)

### Q1 — Test scope?

| Option | Description | Selected |
|--------|-------------|----------|
| Inventory + run existing suite (Recommended) | 9-capability inventory matrix mapping each v1.0 capability → existing tests → pass/fail status with chrome rendered. No new test code; gaps filed as follow-up issues. Cheapest plan; respects v1.0 archive ("don't churn what works"); SC reads "pass unchanged" so adding tests is technically out-of-scope. | ✓ |
| Add 9 E2E teatest flows per capability | New teatest scenarios driving full user journey per capability with chrome rendered. 9 new test files. Highest confidence; biggest plan footprint; arguably scope creep. | |
| Inventory + targeted E2E for chrome-prone risk areas | Inventory all 9, then add 2-3 new E2E flows for capabilities most likely to interact with chrome. Focused scope; addresses chrome-interaction risk specifically. | |
| Tagged TestRegression_* sweep over both | Inventory + new flows + `TestRegression_v10` build tag. Most ceremony; gives release engineering single command. Heaviest commitment. | |

**User's choice:** Inventory + run existing suite (Recommended)
**Notes:** Locked → D-506. "Pass unchanged" SC wording read literally.

### Q2 — Inventory artifact location?

| Option | Description | Selected |
|--------|-------------|----------|
| Inventory in 11-VERIFICATION.md (Recommended) | 9-row matrix in standard /gsd-verifier output. Lives where every other phase's pass/fail evidence lives; reviewers know where to look. Plan 2 owns building the matrix; verifier confirms. | ✓ |
| Inventory in 11-CONTEXT.md `<code_context>` plus VERIFICATION | Capability → test file mapping captured up-front in CONTEXT.md so planner has matrix ready; evidence still in VERIFICATION.md. Earlier visibility on gaps; CONTEXT becomes part-document part-spec. | |
| Standalone REGRESSION-MATRIX.md in the phase dir | Dedicated artifact + `make regression` target in Makefile. Permanent reference for future v1.x patch releases; biggest scope; introduces tooling. | |

**User's choice:** Inventory in 11-VERIFICATION.md (Recommended)
**Notes:** Locked → D-506.

### Q3 — Gap rule?

| Option | Description | Selected |
|--------|-------------|----------|
| File follow-up issue, ship anyway (Recommended) | SC1 sign-off treats "existing tests pass unchanged" as the bar. Gaps recorded as Phase 11 deferred ideas + GitHub issues for v1.1.1 patch work. Respects "no scope creep" rule. | ✓ |
| Backfill a thin smoke test before sign-off | Plan 2 adds *minimum* possible test (5-line teatest) for each gap found. Could land 0-3 small tests. Modest scope expansion; closes gaps now rather than deferring. | |
| Block SC1 sign-off until covered | Phase 11 cannot be VERIFIED until every capability has a teatest covering it with chrome rendered. Most rigorous; rebudgets Plan 2 by N tests. | |

**User's choice:** File follow-up issue, ship anyway (Recommended)
**Notes:** Locked → D-508.

### Q4 — Chrome-interaction sanity checks?

| Option | Description | Selected |
|--------|-------------|----------|
| Clipboard auto-clear race | v1.0 CLB-02 timeout uses tea.Tick; chrome's flash severity prefix renders [W]/[E] at flash time. Worth one teatest that copies, waits 2.5s (mocked), and verifies chrome's clipboardHot indicator + status bar both clear. | ✓ |
| Recipient form + menu hint dispatch | Phase 9 D-309 changes how RecipientForm hints reach menu. Worth a teatest confirming form's field-level hints render in the menu (not underlying file-list ones). | ✓ |
| Health overlay + crumb chip rendering | v1.0 HLT-03 health overlay coexists with crumb row; on narrow widths Phase 10 D-425 first-and-last preservation kicks in. Worth a teatest running health scan at 80x24 and 60x24 and asserting overlay text is reachable. | ✓ |
| No chrome-specific risk — inventory is sufficient | All v1.0 functional flows exercised by existing tests with chrome rendered; inventory will catch any actual regressions. Don't add anything beyond the matrix. | ✓ |

**User's choice:** Clipboard auto-clear race + Recipient form + menu hint dispatch + Health overlay + crumb chip rendering + No chrome-specific risk
**Notes:** Interpreted multi-select as "include the three sanity checks AND treat 'no broader new test surface' as the baseline." Locked → D-507 (3 chrome-prone sanity teatests, no broader new tests beyond those + inventory).

### Q5 — Continue or move on?

**User's choice:** Next area

---

## Terminal compat sweep + alt-screen cleanup (SC3 + SC4)

### Q1 — Sweep scope?

| Option | Description | Selected |
|--------|-------------|----------|
| Linux self-sweep + community for the rest (Recommended) | Verify Alacritty + Ghostty + tmux-nested + VSCode integrated personally; README "verified terminals" matrix + GitHub issue template for community-contributed reports. Phase 11 ships 4/8 verified; rest closes through user reports. | ✓ |
| Linux-only sweep + defer mac/Windows to v1.1.x | Verify the 4 Linux combos; macOS/Windows explicitly OOS for v1.1. SC3 wording softened from 8 to 4 in this phase. | |
| Spin up Mac/Windows VMs or borrow access | Genuinely sweep all 8 combos before sign-off. Highest fidelity; significant out-of-codebase effort. Could push the milestone. | |
| Drop the sweep entirely; rely on user reports | Skip SC3 manual verification; close via "covered by Phase 10 4-profile teatest matrix." Risk: terminal-specific quirks (tmux double-nesting, VSCode 1-row offset) aren't detected by golden tests. | |

**User's choice:** Linux self-sweep + community for the rest (Recommended)
**Notes:** Locked → D-509 + D-510. Pragmatic; doesn't block milestone on hardware not available.

### Q2 — Screenshot record location?

| Option | Description | Selected |
|--------|-------------|----------|
| `.planning/phases/11-.../screenshots/` (Recommended) | Phase artifact directory. Stays inside planning system; doesn't bloat public repo with binary blobs. ~50KB × 4 combos = ~200KB. Convention follows other phases that captured artifacts. | ✓ |
| `docs/terminal-matrix/` in repo root | Public-facing; users browsing the README can see "good" examples. Slightly larger commit; better discoverability. | |
| External — GitHub release page or wiki | Don't commit binary blobs; upload to GitHub release page when v1.1 tag goes out. Keeps repo lean; ties verification to release moment; harder to reproduce from fresh checkout. | |
| No screenshots — ASCII description in VERIFICATION.md | Verification report describes what was observed in prose. No binary artifacts; lowest fidelity; hardest to compare against community reports. | |

**User's choice:** `.planning/phases/11-.../screenshots/` (Recommended)
**Notes:** Locked → D-511.

### Q3 — Alt-screen exit wiring?

| Option | Description | Selected |
|--------|-------------|----------|
| tea.Quit returns blank View (Recommended) | When Quit handler fires, return tea.View{} with empty string. Cursed Renderer flushes blank frame, then ESC[?1049l swaps to user shell. Single integration point in model.go; works inside tea contract; no extra writes to stdout. View() needs `m.quitting` flag. | ✓ |
| main.go post-Run() ANSI reset | After p.Run() returns, write `\x1b[2J\x1b[H` to os.Stdout. Outside tea contract; simpler; works whether Quit is via key or signal. Risk: if alt-screen exit ANSI hasn't flushed, clear targets the alt-screen instead of user shell. | |
| Both — belt + suspenders | tea.View{} on Quit + ANSI reset in main.go. Maximum coverage; tiny code; matches existing `defer clipboard.WriteAll("")` pattern. | |
| Trust Bubble Tea v2 default | Cursed Renderer already swaps out of alt-screen via standard ESC[?1049l; verify by manual sweep; if combo leaves residue, fix as follow-up. Risk: Pitfall 10 specifically called this out for VSCode + tmux-nested. | |

**User's choice:** tea.Quit returns blank View (Recommended)
**Notes:** Locked → D-512. Single integration point; keeps cleanup in tea event loop.

### Q4 — Alt-screen enter wiring?

| Option | Description | Selected |
|--------|-------------|----------|
| Trust Cursed Renderer + zero-state guard (Recommended) | AppModel.View() already returns empty `tea.NewView("")` with AltScreen=true when m.width or m.height is zero. That IS the clear/fill frame. No new code. Verify on 4 Linux terminals; if VSCode integrated terminal still shows 1-row offset, address as deferred or v1.1.1. | ✓ |
| Add explicit ColorBg fill frame | Replace empty View with `lipgloss.NewStyle().Background(ColorBg).Width(w).Height(h).Render("")` — but NewStyle inside View() violates Phase 7 D-15 / TestViewNoNewStyle. Declare FillFrameStyle package var instead. Slightly more rendering work; matches Pitfall 10 §1 verbatim. | |
| Plan 2 verifies, then decides | During Linux compat sweep, observe whether any combo shows residual content from prior shell on first frame. If yes, add explicit fill frame. If no, ship without. Iterative; modest cost. | |

**User's choice:** Trust Cursed Renderer + zero-state guard (Recommended)
**Notes:** Locked → D-513. If VSCode/tmux shows 1-row offset, falls to deferred ideas with explicit FillFrameStyle pattern.

### Q5 — Continue or move on?

**User's choice:** Next area

---

## "Looks Done But Isn't" sign-off mechanism (SC5)

### Q1 — Sign-off table location?

| Option | Description | Selected |
|--------|-------------|----------|
| 11-VERIFICATION.md SC5 section (Recommended) | 15-row table in verifier output. Lives where every other phase verification lives. PITFALLS.md (research artifact) stays untouched. Plan 2 builds the table; verifier confirms each row's evidence. | ✓ |
| Standalone CHECKLIST.md in phase 11 dir | Dedicated artifact with [x] / [ ] / [N/A] markers + one-line evidence per row. More prominent than burying inside VERIFICATION.md. Adds artifact convention other phases haven't used. | |
| Update PITFALLS.md with [x] inline | Mutate research/PITFALLS.md lines 559-573: turn `- [ ]` into `- [x]` with phase pointers. Closes loop where checklist was authored. Risk: research artifacts intended as point-in-time references; mutating blurs snapshot semantics. | |
| Both VERIFICATION.md table + PITFALLS.md updates | Defense in depth; double the maintenance surface. | |

**User's choice:** 11-VERIFICATION.md SC5 section (Recommended)
**Notes:** Locked → D-514. PITFALLS.md stays untouched as snapshot.

### Q2 — N/A handling?

| Option | Description | Selected |
|--------|-------------|----------|
| Explicit `N/A` row with reason (Recommended) | Row 3 stays in table: `[N/A] Skin fail-open — deferred to v2 (THM-01..03)`. Future readers see deferral didn't fall through cracks; v2 milestone planning can grep `[N/A]`. Honest about milestone scope. | ✓ |
| Drop the row | Phase 11 SC5 table is 14 rows long. Cleaner-looking; future readers might wonder where row #3 went. | |
| Two-table split: "Verified" + "Deferred" | 13 verified rows in main table; smaller "Deferred to v2" sub-table lists skin-loader item with rationale. Maximum clarity; introduces sub-table convention. | |

**User's choice:** Explicit `N/A` row with reason (Recommended)
**Notes:** Locked → D-515.

### Q3 — Evidence capture for already-verified items?

| Option | Description | Selected |
|--------|-------------|----------|
| Re-run gates at sign-off + cite prior phase (Recommended) | Phase 11 verifier executes relevant test command per item and captures pass/fail in SC5 table. Cell value: `✓ PASS — evidence: chrome_test.go:182 + Phase 7.1 VERIFICATION.md SC2`. Costs ~2-3 minutes; catches drift between phases. | ✓ |
| Cite prior verification only | Cell value: `✓ Verified in Phase N (date) — see N-VERIFICATION.md SCM`. No re-run. Cheapest; trusts prior verifications; if regression slipped between phase N and Phase 11, this won't catch it. | |
| Re-run only the grep gates, cite test-suite items | Grep gates re-run because they're cheap; richer suites cite prior verification. Splits the difference. | |

**User's choice:** Re-run gates at sign-off + cite prior phase (Recommended)
**Notes:** Locked → D-516. Catches inter-phase drift cheaply.

### Q4 — Wrap up or more questions?

**User's choice:** Wrap up

---

## Final wrap-up

**User's choice:** I'm ready for context

---

## Claude's Discretion

Areas where the user did not need to lock the answer; downstream planner picks per the recommendations recorded in 11-CONTEXT.md `<decisions>` § "Claude's Discretion":

- chromeKey struct shape (4-field struct vs string concatenation) — recommendation: 4-field struct
- Single chromeCache vs split chromeCache + chromeCrumbsCache — recommendation: split
- Whether `quitting` lives on AppModel or as a Cmd-driven state machine — recommendation: AppModel field
- Plan 2 ordering (regression inventory vs Linux sweep) — recommendation: regression first
- README "Verified Terminals" matrix exact format — recommendation: H2 section + markdown table
- Issue template field shape — recommendation: GitHub Forms YAML
- Whether to fold `m.quitting` wiring into Plan 1 or Plan 2 — recommendation: Plan 2
- PNG vs asciinema for the 4 verification screenshots — recommendation: PNG
- TestChromeCache_HitRateAtSteadyState file location (chrome_test.go vs new cache_test.go) — recommendation: chrome_test.go

## Deferred Ideas

Captured in 11-CONTEXT.md `<deferred>` section. Summary:

- **v1.1.x patches:** per-renderer caches if hit rate proves low; allocation hygiene + manual menu columns if cache-only misses 50µs (Plan 1 fallback escalation, in-place); generation counter for cache keys; explicit FillFrameStyle on alt-screen enter; post-`p.Run()` ANSI reset; signal-handler blank frame; Pitfall 17 info-panel re-fetch on SOPS error; macOS / iTerm2 / Windows / WSL2 manual sweeps via community.
- **v2:** skin YAML loader (THM-01..03); multi-pane focus rings; `:` command bar; mouse interactions.
- **Possibly Plan 2, possibly v1.1.x:** README matrix exact format; issue template shape; m.quitting wiring placement.
- **Out of scope (scope creep):** E2E flows for all 9 v1.0 capabilities; tagged build tag; standalone REGRESSION-MATRIX.md / CHECKLIST.md; PITFALLS.md mutation; two-table SC5 split; VM-based macOS/Windows verification; new chrome features added to cache; ROADMAP SC2 wording change.
