# Phase 11: Regression + Perf Gates - Context

**Gathered:** 2026-05-04
**Status:** Ready for planning

<domain>
## Phase Boundary

Lock down v1.1 release-readiness. Drive `BenchmarkAppView` from the empirical ~2.4-2.8 ms/op back under the 50 µs target (48× lift) by wiring the D-18 chrome cache, prove no v1.0 functional flow has regressed under the chrome rework, sweep the chrome on the Linux terminals available to this developer, wire alt-screen exit cleanup, and roll up the "Looks Done But Isn't" 15-item checklist into a single SC5 sign-off table. Two requirements close: UI-20 (no v1.0 regress) and UI-21 (50 µs/op + zero `lipgloss.NewStyle()` reachable from `View()`).

**In scope (this phase):**

- `internal/app/model.go` — `AppModel` gains `chromeCache string`, `chromeCacheKey chromeKey` (struct of `state sessionState; recipientAction string; searchActive bool; width int`), and `quitting bool` fields. New unexported helper `(m AppModel) refreshChromeCache() AppModel` that recomputes `chromeCache` + `chromeCrumbsCache` (or a single concatenated string per planner discretion) from the cache key inputs. `Update` invokes `refreshChromeCache` on every branch that mutates a chrome-input field — `WindowSizeMsg`, every state transition (~20 sites), search toggle (`fileList.IsSearchActive` flip), `recipientAction` field assignments. `View()` reads the cache and skips the RenderChrome / RenderCrumbs path when `chromeCacheKey == m.chromeCacheKey`. Quit handling sets `m.quitting = true` and returns `tea.Quit`; the next (and final) `View()` call returns `tea.NewView("")` with `AltScreen=true` so the Cursed Renderer's last frame is blank before the alt-screen leaves.
- `internal/app/chrome_test.go` — Remove the `t.Skip(...)` on line 311 of `TestBenchmarkAppView_UnderBudget`. The 50,000 ns budget already lives at line 317; the test starts gating once the cache lands. Add `TestChromeCache_HitRateAtSteadyState` (drives 100 frames with no chrome-input changes, asserts `chromeCache` is reused 99/100 frames — proves the cache is wired, not just present) per Plan 1.
- `internal/app/bench_test.go` — Add inline comment documenting the empirical 2.4-2.8 ms baseline + post-cache target 50 µs/op + cache hit rate the perf depends on. No new benches.
- `internal/app/regression_test.go` (new) — Three chrome-interaction sanity teatests under one file:
  1. `TestRegression_ClipboardAutoClearWithChrome` — copy a value, advance simulated time by clipboard timeout (mocked via `app.ClipboardTimeout` test seam), verify `IsClipboardHot()` flips false + chrome's clipboard indicator clears + no [W]/[E] flash leaked from the typed flash API
  2. `TestRegression_RecipientFormMenuHints` — drive into `stateRecipientForm`, verify the rendered menu shows form-level hints (Tab/Enter/Esc) not file-list ones (j/k/Enter/q)
  3. `TestRegression_HealthOverlayOnNarrowWidth` — run a health scan at 80×24 and 60×24 (two narrow widths from Phase 10 D-424 matrix), assert health overlay text is reachable + crumb chips don't truncate the active "Health" segment
- `cmd/sops-tui/main.go` — No changes. Trust Cursed Renderer + zero-state guard for alt-screen enter (Pitfall 10 §1 prescription deferred to v1.1.x bug reports if any combo fails the manual sweep). The signal handler at lines 56-60 already exits cleanly via `os.Exit(0)`; no blank frame needed because the signal path bypasses the tea event loop.
- `.planning/phases/11-regression-perf-gates/11-VERIFICATION.md` — Built by `/gsd-verify-work` at phase close. Contains:
  - **SC1 9-row inventory matrix:** capability → covering test files (file:line) → pass/fail with chrome rendered → last-run timestamp. Columns: capability / unit tests / integration tests / regression sanity check / status.
  - **SC2 evidence:** `TestBenchmarkAppView_UnderBudget` output (ns/op, headroom over 50,000 ns); `pprof` output snippet showing `RenderChrome` + `RenderMenu` no longer in the hot path; `TestChromeCache_HitRateAtSteadyState` pass.
  - **SC3 evidence:** 4 Linux terminal screenshots in `screenshots/` subdirectory (alacritty.png, ghostty.png, tmux-nested.png, vscode-integrated.png); per-combo notes on resize behaviour, alt-screen enter/exit cleanliness, any visual quirks.
  - **SC4 evidence:** Manual observation that on `q` press the user's shell prompt area shows no chrome residue (per-terminal verified during the SC3 sweep). `m.quitting` flag wiring referenced (model.go:LINE).
  - **SC5 15-row table:** item / status (Done / Phase 11 / N/A) / evidence pointer (file:line + prior phase VERIFICATION.md SC reference). Re-run the existing gates (TestChromeASCIIOnly / TestChromeNormalBorderOnly / TestViewNoNewStyle / TestSubmodelViewsNoNewStyle / TestMenuHints_Drift / TestRenderCrumbs_FirstAndLastSegmentsPreserved / 4-profile matrix tests / etc) and capture pass/fail in the table.
- `.planning/phases/11-regression-perf-gates/screenshots/` (new) — Four PNG captures of the chrome rendering on the four Linux verification combos. Filenames match the verifier convention.
- `README.md` — Add a "Verified Terminals" section (likely a new H2 between "Installation" and "Usage" — exact placement is planner discretion). 4 Linux combos verified in v1.1; matrix table for community-contributed reports for the remaining 4 (macOS Terminal, iTerm2, Windows Terminal, WSL2). Link to a GitHub issue template (`.github/ISSUE_TEMPLATE/terminal-bug.yml`) for terminal-specific bug reports — template asks for terminal name + version + screenshot. Plan 2 owns the README + issue template additions.
- `.planning/STATE.md` — Phase 11 close updates per `/gsd-verify-work`.
- `.planning/REQUIREMENTS.md` — UI-20 + UI-21 marked Complete with Phase 11 evidence pointers per `/gsd-verify-work`.

**Out of scope (deferred per ROADMAP / explicit decisions):**

- Per-renderer caches (`RenderMenu` / `RenderChrome` / `RenderCrumbs` each with their own `sync.Map`) — rejected; full-string cache is simpler and the empirical profile already shows ~85% of the cost is in the chrome path, so a single string cache amortises the same work with one invalidation surface.
- Allocation hygiene + replacing `lipgloss/v2/table` with a hand-rolled column joiner for `RenderMenu` — rejected as primary mechanism (the cache amortises this anyway). Documented as the fallback in Plan 1 if the cache alone misses the 50 µs target on dev hardware: at that point planner switches Plan 1 from "cache only" to "cache + manual menu columns" and re-runs the bench.
- Generation counter (`chromeGen uint64` on AppModel, bumped in every chrome-input-mutating Update branch) — rejected; D-18 minimum cache key (state, recipientAction, IsSearchActive, width) is computable directly from existing AppModel fields without per-mutation bookkeeping. Avoids the 42-callsite-style audit cost.
- Cache key including `palette` / `logoStatus` / `infoPanelData` / `flashGen` — rejected; palette is set once at startup (read-only field), logoStatus is a pure function of inputs the user already sees through a flash, infoPanelData is refreshed on its own seam (Phase 8 D-213) and currently doesn't change during a held-down j/k scroll. The minimum key catches every realistic mid-session mutation.
- E2E teatest flows for all 9 v1.0 capabilities — rejected as scope creep against SC1 wording ("pass unchanged" not "extend coverage"). Only 3 chrome-prone sanity teatests added (clipboard race, recipient form, health overlay).
- Tagged `TestRegression_*` / `+build regression` build tag + `make regression` Makefile target — rejected; over-ceremony for a one-shot milestone closure. Planner places sanity teatests in `internal/app/regression_test.go` per existing convention; verifier runs `go test ./...` like every other phase.
- Standalone `REGRESSION-MATRIX.md` artifact — rejected; the 9-row inventory in 11-VERIFICATION.md SC1 is sufficient.
- Standalone `CHECKLIST.md` artifact — rejected; the 15-row sign-off table in 11-VERIFICATION.md SC5 is sufficient.
- Mutating `.planning/research/PITFALLS.md` with `[x]` inline next to the 15 checklist items — rejected; research artifacts stay point-in-time snapshots of the milestone-start understanding. The closure record lives in 11-VERIFICATION.md.
- `lipgloss.NewStyle().Background(ColorBg).Width(w).Height(h).Render("")` explicit fill frame on alt-screen enter (Pitfall 10 §1 verbatim) — rejected; the existing zero-state guard (`m.width == 0 || m.height == 0` → `tea.NewView("").AltScreen=true` at model.go:1364-1369) plus Bubble Tea v2's Cursed Renderer alt-screen enter handling already paint a clean first frame on the 4 Linux combos this phase verifies. If a community report for VSCode integrated terminal or tmux-nested surfaces a 1-row-offset / residual-content issue, it lands as a v1.1.x patch with the explicit `FillFrameStyle` package var.
- Post-`p.Run()` ANSI reset (`fmt.Print("\x1b[2J\x1b[H")` in `main.go`) for alt-screen exit — rejected; `tea.View{}` blank frame on `m.quitting = true` keeps the cleanup contract inside the tea event loop where every other AppModel concern lives. main.go stays a thin entry point.
- Signal-handler blank frame — the SIGINT/SIGTERM handler at main.go:56-60 already calls `os.Exit(0)` after clipboard cleanup; the alt-screen ANSI exit sequence is emitted by the renderer's deferred shutdown hook before exit. No additional wiring; if a signal-path leak surfaces as a bug report, lands as a v1.1.x patch.
- Manual sweep of macOS Terminal / iTerm2 / Windows Terminal / WSL2 — deferred; these closes via README "Verified Terminals" matrix + community issue template. v1.1 ships with 4/8 combos verified; remaining 4 close as users adopt the tool. ROADMAP SC3 wording ("Chrome renders correctly on the supported terminal matrix") is satisfied by the Linux combos + the deferred-with-template policy; verifier interprets the SC against actual user reach.
- VM-based macOS / Windows verification before sign-off — rejected; significant out-of-codebase effort (license keys, Mac hardware) that would push the milestone for marginal coverage gain over community reports.
- Re-running the Phase 6 / 7 / 7.1 / 8 / 9 / 10 grep gates *only* (without re-running test-suite items like `TestMenuHints_Drift` or the 4-profile matrix) — rejected; the marginal cost of `go test ./...` at verifier time is small enough that re-running everything catches inter-phase drift cheaply.
- Two-table SC5 split (`Verified` + `Deferred to v2`) — rejected; one 15-row table with explicit `[N/A]` rows is more honest about the milestone's scope.
- New chrome features added to the cache (e.g. dynamic logo width-responsive trimming, info-panel cache-busting on git-status mid-scroll) — out of scope; Phase 11 wires the cache for the chrome that exists today.
- Skin YAML loader, builtin skins, fsnotify hot-reload (`THM-01..03`) — v2 milestone per ROADMAP "Out of Scope for v1.1".
- `:` command bar / number-key view switching / `[`/`]` history navigation — v1.2+ per ROADMAP.
- Mouse interactions / multi-pane focus rings — v2; project core value is keyboard-only.
- Header info-panel re-fetch on SOPS error (Pitfall 17 — chrome stale during async ops) — possibly v1.1.x if user reports a stale-fingerprint after key-rotation flash; not gated by Phase 11 SC.
- ROADMAP SC2 wording change ("≤ 50 µs/op at 200×60" → looser budget) — explicit Phase 7.1 governance lock; original target preserved.
- New runtime dependencies — none needed; testing.B + testing.Benchmark already in stdlib; pprof analysis is human-driven.

</domain>

<decisions>
## Implementation Decisions

### Chrome Cache Shape (UI-21 / SC2)

- **D-501: Full chrome string cache on AppModel.** New fields `chromeCache string` and `chromeCacheKey chromeKey` on `AppModel`, where `chromeKey struct { state sessionState; recipientAction string; searchActive bool; width int }`. The cache string holds the result of `RenderChrome(...) + RenderCrumbs(...)` joined as the planner sees fit (single concat or two separate cache fields — planner discretion, recommendation: two fields so that `crumbs` can be invalidated separately if a future phase needs it). View() reads `m.chromeCache` and `m.chromeCrumbsCache` directly; skips the RenderChrome / RenderCrumbs JoinHorizontal cost on cache hit.
- **D-502: Cache key = D-18 minimum: (state, recipientAction, IsSearchActive, width).** No palette, logoStatus, infoPanelData, or flashGen in the key. Rationale: palette is read-only after startup (set once in NewAppModel), logoStatus is a pure derivation the user notices through the flash anyway (so the flash-driven invalidation by-state suffices), infoPanelData is event-refreshed (Phase 8 D-213 cache pattern guarantees it doesn't churn during a hot scroll), flashGen drives the status bar not the chrome. The minimum key catches every realistic mid-session mutation; broader keys would just lower the hit rate without fixing real bugs.
- **D-503: Cache populated in `Update`, never in `View`.** New unexported `(m AppModel) refreshChromeCache() AppModel` helper computes the key and rebuilds the cache strings if the key has changed since last refresh. Every Update branch that flips a key field calls this helper before returning the new model. Mutation sites (planner audits during Plan 1):
  - `WindowSizeMsg` handler (~model.go:313-328) — width changes
  - State-transition assignments (`m.state = ...`) — search returns ~20 sites
  - Search toggle inside `fileList.Update` propagation — `searchActive` derived from `m.fileList.IsSearchActive()`
  - `recipientAction` field assignments (`m.recipientAction = "addrecipient"` etc.) — search returns ~6 sites
  Pattern matches Phase 8 `infoPanel` cache (Pitfall 15 prescription): mutate-on-event, never-on-render. View() stays pure-function-of-state.
- **D-504: SC2 gate = flip `TestBenchmarkAppView_UnderBudget` on.** Plan 1 deletes the `t.Skip("deferred to Phase 11 SC2 — D-18 caching fallback...")` line at chrome_test.go:311. The existing test runs `testing.Benchmark(BenchmarkAppView)` and fails if `nsPerOp > 50_000`. The 200×60 fixture in `BenchmarkAppView` (bench_test.go:18) is preserved unchanged. No looser budget is accepted; if the cache alone misses 50 µs, planner switches to "cache + allocation hygiene" (drop `lipgloss/v2/table` for menu, hand-roll columns per Phase 7.1 D-116 narrow-tier precedent) within Plan 1 and re-runs.
- **D-505: Add `TestChromeCache_HitRateAtSteadyState`.** Drives 100 `View()` calls on an AppModel with no Update in between (simulates held-down j frame burst); asserts `m.chromeCache` is reused at least 99/100 frames. This proves the cache is *wired* (not just declared) — catches the failure mode where the cache field exists but Update never refreshes it. Deliberately measures hit rate, not wall-clock; survives Go version bumps.

### v1.0 Regression Suite (UI-20 / SC1)

- **D-506: 9-capability inventory matrix in 11-VERIFICATION.md SC1 section.** Built at verifier time. Rows: file discovery / reveal / edit / diff / rotate / clipboard / git / recipient management / health. Columns: capability label / unit test files (sops/discoverer_test.go, parser/yaml_test.go, sops/executor_test.go, git/status_test.go, health/checker_test.go) / integration test files (model_clipboard_test.go, model_reveal_test.go, model_test.go, ui/health_test.go, etc.) / regression sanity check (D-507) / status (PASS / FAIL / GAP). Evidence: `go test ./...` output run with chrome rendered (every NewAppModel call now includes the chrome path). No new test code beyond the three sanity checks below.
- **D-507: Three chrome-interaction sanity teatests in `internal/app/regression_test.go`.** Each is one-screen, one-flow, one-assertion:
  - **TestRegression_ClipboardAutoClearWithChrome** — copy a decrypted value, advance time via the `app.ClipboardTimeout` test seam, verify `m.IsClipboardHot()` flips false, the status bar's `[clip]` indicator clears, and no stray `[W]` or `[E]` prefix appears in the next-frame flash text. Catches: typed flash API regression interacting with the clipboard timeout.
  - **TestRegression_RecipientFormMenuHints** — drive into `stateRecipientForm`, ANSI-strip the rendered View, assert the menu row contains the form-level hint mnemonics (Tab/Enter/Esc) and does NOT contain file-list hints (j/k/q). Catches: Phase 9 menuHints dispatcher regression for nested form states.
  - **TestRegression_HealthOverlayOnNarrowWidth** — run a health scan at 80×24 and 60×24, assert the overlay's primary content (e.g. "Weak secrets:" or "(none found)") appears in the ANSI-stripped View, and the crumb row's last segment (`<health>`) is preserved per D-425. Catches: Phase 10 narrow-terminal first/last-segment regression intersecting with the health overlay.
- **D-508: Coverage gaps file follow-up GitHub issues; ship anyway.** SC1 reads "All v1.0 functional integration tests ... pass unchanged after chrome lands — no v1.0 feature regresses." If the inventory finds a v1.0 capability with thin coverage (e.g. format-aware rotation has only `internal/ui/rotate_test.go` unit-level coverage and no AppModel integration test), Phase 11 verifier files a GitHub issue tagged `regression-coverage-gap` for v1.1.x patch work and proceeds. The SC doesn't require expanding coverage; it requires no regressions in what already exists. This respects the "no scope creep" rule and keeps Plan 2 small.

### Terminal Compat Sweep + Alt-Screen Cleanup (SC3 + SC4)

- **D-509: Linux self-sweep on 4 combos.** Alacritty (baseline TrueColor), Ghostty (baseline TrueColor with different rendering pipeline), tmux-nested-in-Alacritty (covers the double-alt-screen interaction Pitfall 10 calls out), VSCode integrated terminal (xterm.js — covers the historical 1-row-offset issue). Per-combo verification checklist: chrome renders correctly at 80×24 + 200×60, no flicker on resize between those widths, alt-screen enters cleanly (no residual content from prior shell), alt-screen exits cleanly (no chrome residue in user shell prompt area), `q` press returns to shell with cursor at expected position, ctrl-c (SIGINT) returns to shell with clipboard cleared.
- **D-510: macOS Terminal / iTerm2 / Windows Terminal / WSL2 closed via community reports.** README gains a "Verified Terminals" section with a table: 4 Linux combos marked verified-by-author (with v1.1 release tag); 4 remaining combos marked "community-contributed reports welcome" with a link to a new `.github/ISSUE_TEMPLATE/terminal-bug.yml` template. Template fields: terminal name, version, OS, screenshot, expected vs actual behaviour, reproduction steps. The verifier interprets ROADMAP SC3 ("supported terminal matrix") against developer reach: 8/8 verification would push the milestone for marginal coverage gain over community reports; 4/8 verified + a clear path for the rest is the right tradeoff.
- **D-511: Screenshots in `.planning/phases/11-regression-perf-gates/screenshots/`.** PNG captures from the 4 Linux combos at 200×60 in stateFileList (the most chrome-rich state). Files: `alacritty.png`, `ghostty.png`, `tmux-nested.png`, `vscode-integrated.png`. Verifier cites these inline in 11-VERIFICATION.md SC3 evidence rows. ~50KB each × 4 = ~200KB total — negligible repo bloat. Stays inside the planning system; doesn't bloat the public README/docs/ tree.
- **D-512: Alt-screen exit blank frame via `m.quitting` flag.** `AppModel` gains a `quitting bool` field (zero-value false). Every Update branch that returns `tea.Quit` (today: `tea.KeyPressMsg` matching `keys.DefaultGlobalKeyMap.Quit`) first sets `m.quitting = true` before returning `m, tea.Quit`. View() top of function: `if m.quitting { v := tea.NewView(""); v.AltScreen = true; return v }`. Cursed Renderer's last frame before alt-screen leave is therefore blank; no chrome residue in user's shell prompt area or tmux scrollback. Single integration point; works whether Quit is via key or future signal-driven path.
- **D-513: Alt-screen enter trusted to Cursed Renderer + zero-state guard.** No new code on the enter side. Existing model.go:1364-1369 already returns `tea.NewView("")` with `AltScreen=true` when `m.width == 0 || m.height == 0` (before the first WindowSizeMsg arrives) — that's the de facto fill frame. If any of the 4 verification combos shows an enter-side artifact (Pitfall 10 §1 1-row-offset on VSCode), Plan 2 closes it via the explicit `FillFrameStyle` package var pattern; otherwise ship. The "Looks Done But Isn't" item #15 ("Chrome hidden or adapted in all 14 session states") is satisfied by the per-state goldens already shipped — Phase 11 doesn't re-audit it, just re-runs the gates.

### "Looks Done But Isn't" 15-Row Sign-Off (SC5)

- **D-514: 15-row sign-off table in 11-VERIFICATION.md SC5 section.** One row per checklist item from `.planning/research/PITFALLS.md` lines 559-573. Columns: # / item description (verbatim from PITFALLS.md) / status (✓ Done / ⏳ Phase 11 / N/A) / evidence (file:line + prior phase VERIFICATION.md SC reference, or N/A reason). Built by `/gsd-verify-work`. Self-contained — readers don't need to flip back to PITFALLS.md to understand the closure.
- **D-515: Explicit `[N/A]` row for skin-fail-open with v2 deferral reason.** Item #3 ("Skin fail-open: corrupt skin.yaml → TUI launches with default palette + warning") stays in the table with status `N/A` and evidence `Skin loader deferred to v2 (THM-01..03 per ROADMAP). Pitfall 4 sidestepped, not solved.` Future v2 milestone planning can grep for `[N/A]` to find what was punted.
- **D-516: Re-run gates at verifier time + cite prior phase evidence.** Each `Done` row's evidence cell takes the form `✓ PASS — {test_name} (file:line) + Phase N VERIFICATION.md SCM`. Verifier runs the relevant `go test -run TestX ./...` for each cited test (~30 seconds total for the gate sweep) and captures pass/fail. Catches drift between phase N's verification and Phase 11's verification (e.g. a later phase accidentally landing a `lipgloss.NewStyle()` reachable from View). For the 2 SC1+SC2 rows that close in Phase 11 itself, evidence points to the in-phase tests + bench output.

### Plan Split (2-plan ROADMAP budget)

- **D-517: Plan 1 = chrome cache wiring + bench gate flip + cache hit-rate test (SC2 closure).** All risk concentrates here:
  - `internal/app/model.go`: add `chromeCache`, `chromeCacheKey`, `chromeCrumbsCache`, `quitting` fields; `chromeKey` struct; `refreshChromeCache()` helper; `Update` branch instrumentation (~25 mutation sites audited); View() cache-read + zero-state + quitting branches; `m.quitting = true` on Quit handler.
  - `internal/app/chrome_test.go`: remove `t.Skip` at line 311; add `TestChromeCache_HitRateAtSteadyState`.
  - `internal/app/bench_test.go`: doc comment updates with empirical baseline + cache hit rate target.
  - Plan 1 SUMMARY: pre/post bench numbers, pprof snippet, cache hit rate measurement on hot scroll.
  - Fallback escalation: if cache-only doesn't hit 50 µs on dev hardware, Plan 1 expands in-place to "cache + manual menu columns" (drop `lipgloss/v2/table` for `RenderMenu`, hand-roll JoinHorizontal of 2 pre-rendered columns per Phase 7.1 D-116 narrow-tier precedent). Single plan; either approach lands atomically.
- **D-518: Plan 2 = regression inventory + 3 chrome-interaction sanity teatests + Linux compat sweep + alt-screen exit + 15-row sign-off prep + README + issue template (SC1 + SC3 + SC4 + SC5 closure).**
  - `internal/app/regression_test.go`: 3 sanity teatests per D-507.
  - `internal/app/model.go`: `m.quitting` flag wiring on Quit handlers (D-512). Note: this work is small enough that planner may decide to fold it into Plan 1 if Plan 1 is already touching the relevant Update branches; recommendation is to keep it in Plan 2 to maintain the SC2-only / non-SC2 split, but planner discretion.
  - 4-combo Linux verification: capture screenshots, build verification matrix.
  - `README.md`: "Verified Terminals" section with the 4-Linux + 4-community matrix.
  - `.github/ISSUE_TEMPLATE/terminal-bug.yml`: issue template per D-510.
  - 11-VERIFICATION.md draft: 9-row regression inventory matrix (D-506); SC3+SC4 evidence (screenshots + observations); 15-row "Looks Done But Isn't" sign-off (D-514); SC5 evidence (re-run gates + cite prior phases).
  - Plan 2 SUMMARY: per-combo observations, any quirks worth filing as v1.1.x issues, gate re-run pass/fail counts.

### Claude's Discretion

- **chromeKey struct shape** — Plan 1 author picks: 4-field struct as described in D-501, or a `string` concatenation key (`fmt.Sprintf("%d|%s|%v|%d", state, recipientAction, searchActive, width)`). Recommendation: 4-field struct because Go map keys with structs hash directly and avoid the `Sprintf` allocation per Update.
- **Single chromeCache vs split chromeCache + chromeCrumbsCache** — Plan 1 author picks. Recommendation: split, because `RenderCrumbs` and `RenderChrome` have different invalidation pressure (segments change with state but not with width if the segments themselves are stable). Splitting also lets the JoinVertical of (chrome, crumbs, body, statusbar) stay one cheap operation per frame.
- **Whether `quitting` lives on AppModel or as a Cmd-driven state machine** — Plan 1/2 author picks. Recommendation: AppModel field because it's a single bool with one mutation site (Quit handler) and one read site (View top); a Cmd-driven approach would over-engineer.
- **Plan 2 ordering** — author picks: regression inventory before or after Linux sweep? Recommendation: regression first (cheaper, catches gaps before the time investment of the manual sweep).
- **README "Verified Terminals" matrix exact format** — Plan 2 author picks. Recommendation: H2 section between "Installation" and "Usage"; markdown table with columns for Terminal / Version Tested / OS / Status / Notes; footer link to issue template.
- **Issue template field shape** — Plan 2 author picks. Recommendation: GitHub Forms YAML (`.github/ISSUE_TEMPLATE/terminal-bug.yml`) over markdown template; required fields: terminal name, version, OS, screenshot, expected behaviour, observed behaviour, reproduction steps.
- **Whether to fold `m.quitting` wiring into Plan 1 or Plan 2** — see D-518 note. Recommendation: Plan 2, but planner may move it if Plan 1 audit hits the same Update branches anyway.
- **Whether the 4 screenshots are committed as PNG or asciinema recordings** — Plan 2 author picks. Recommendation: PNG screenshots (smaller, simpler diff, no playback infrastructure); if asciinema gives better fidelity for tmux-nested resize behaviour, planner may use that instead.
- **Whether `TestChromeCache_HitRateAtSteadyState` lives in chrome_test.go or a new cache_test.go** — Plan 1 author picks. Recommendation: chrome_test.go (co-located with the bench it gates).

</decisions>

<specifics>
## Specific Ideas

- **D-18 verbatim drives D-501-D-503.** Phase 7 CONTEXT D-18 documented "If a later palette pass regresses the bench, caching can be bolted on without public API change. Pattern: model-level cache keyed on (state, recipientAction, IsSearchActive, width)." Phase 11 picks up that exact prescription with no narrative drift. The empirical 2.4-2.8 ms baseline is the input; D-18's documented fallback is the output.
- **"Pass unchanged" wording in SC1 read literally.** No new test surface beyond the 3 chrome-prone sanity checks. The user explicitly recommended "Inventory + run existing suite" — read this as a deliberate constraint to keep Plan 2 small + respect the v1.0 archive (don't churn what works). The 3 sanity checks are not new coverage areas — they're targeted assertions on chrome-interaction risks specific to the v1.1 rework.
- **Linux self-sweep treats SC3 as best-effort within developer reach.** No VM setup, no borrowed Mac access, no Windows license. Community contribution mechanism (README matrix + issue template) is the close-the-loop pattern. Verifier interprets ROADMAP SC3 against developer reach when no team Mac/Windows hardware exists. Explicitly NOT a milestone blocker.
- **Alt-screen exit via `tea.View{}` on `m.quitting` flag keeps cleanup inside the tea Update loop.** Alternative was a post-Run() ANSI reset in main.go (`fmt.Print("\\x1b[2J\\x1b[H")`), which would inject ANSI sequences outside the renderer's awareness. Routing through the tea contract preserves the "main.go is a thin entry point" pattern.
- **Cache key minimum (4 fields) over broader keys.** D-502 picks the minimum because broader keys (palette, logoStatus, infoPanelData, flashGen) don't catch real bugs — palette is set once, logoStatus is a flash-driven-derivation the user notices through the flash anyway, infoPanelData is event-refreshed (Phase 8 D-213), flashGen drives the status bar not the chrome. Lower-overhead, higher hit-rate.
- **15-row sign-off table is the canonical closure record; PITFALLS.md stays a snapshot.** Research artifacts (PITFALLS.md, SUMMARY.md, ARCHITECTURE.md) capture the milestone-start understanding. Phase 11 doesn't mutate them. The closure lives in 11-VERIFICATION.md SC5, which is the standard verifier output. Keeps research/ as a clean reference for v2 / v1.1.x retrospectives.
- **Re-run gates at verifier time + cite prior phase evidence.** Catches inter-phase drift cheaply (~30 seconds of `go test -run TestX ./...`). The 12 already-verified rows aren't trusted on prior verifications alone — they're re-validated against current HEAD, with prior verifications cited as the historical context. If a later phase accidentally landed a `lipgloss.NewStyle()` reachable from View, this catches it before sign-off.
- **Plan 1 owns SC2; Plan 2 owns the rest.** The 2-plan ROADMAP budget aligns naturally with the SC distribution: SC2 (perf cache) is the biggest technical risk and code change; SC1+SC3+SC4+SC5 (regression sanity + manual sweep + alt-screen exit + sign-off) are smaller integration pieces. Plan 1 is the bigger plan; Plan 2 is the closer.
- **Fallback escalation in Plan 1.** If cache-only doesn't hit 50 µs on dev hardware (Ryzen 7 PRO 5850U baseline from Phase 7.1 chrome_test.go comment), Plan 1 expands in-place to "cache + manual menu columns" (drop `lipgloss/v2/table` for `RenderMenu`, hand-roll JoinHorizontal of 2 pre-rendered columns per Phase 7.1 D-116 narrow-tier precedent). Single plan; either approach lands atomically. Documented up-front so the Plan 1 author knows what to do if first iteration misses budget.
- **Three sanity teatests target chrome-interaction risk, not capability coverage.** Clipboard race tests the typed flash API + clipboard timeout interaction; recipient form tests the menuHints dispatcher; health overlay tests narrow-width chip preservation. These are the three places where the chrome rework most plausibly introduced an interaction bug; non-chrome capabilities (file discovery, edit, diff, rotate, git history) are exercised by their existing tests.

</specifics>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project decision docs

- `.planning/ROADMAP.md` §"Phase 11: Regression + Perf Gates" — Goal, 5 Success Criteria (SC1-SC5), 2-plan budget, UI-20/UI-21 requirements
- `.planning/ROADMAP.md` §"Milestone v1.1 — Explicitly Out of Scope for v1.1" — reject list inherited from prior phases (no new dependencies, no skin loader, no `:` command bar, no mouse, etc.)
- `.planning/REQUIREMENTS.md` §"Regression & Performance" — UI-20 (no v1.0 regress) + UI-21 (≤50 µs/op + zero NewStyle in View)
- `.planning/PROJECT.md` — k9s visual parity is hard product attribute; v1.0 functional non-regression is non-negotiable; AGPL-3.0 license

### Prior phase decisions (carried forward, do not re-decide)

- `.planning/phases/07-chrome-skeleton/07-CONTEXT.md` D-18 — chrome cache fallback documented; "model-level cache keyed on (state, recipientAction, IsSearchActive, width)"; D-501-D-503 implement this verbatim
- `.planning/phases/07-chrome-skeleton/07-CONTEXT.md` D-24 — 50 µs target locked; D-504 enforces this gate
- `.planning/phases/07.1-chrome-gap-closure/07.1-CONTEXT.md` — SC1 governance restoration of the original 50 µs ROADMAP wording; `TestBenchmarkAppView_UnderBudget` `t.Skip` deferral pointer to Phase 11; budgetNs = 50_000 preserved in code
- `.planning/phases/07.1-chrome-gap-closure/07.1-CONTEXT.md` D-116 — 3-tier chrome width fallback (narrow `<41` → "press ? for help" stub; mid `41 ≤ width < 99` → menu+logo; full `≥99` → info-panel + menu + logo); fallback escalation in D-517 borrows the manual-column-join pattern from this tier
- `.planning/phases/08-header-info-panel/08-CONTEXT.md` D-213 — `infoPanel` cache field on AppModel + 4 refresh seams (Pitfall 15 prescription); D-503 follows the same mutate-on-event pattern verbatim
- `.planning/phases/08-header-info-panel/08-CONTEXT.md` D-216 — crumbs middle-segment ellipsis on overflow; D-507's narrow-width health overlay teatest verifies first/last preservation
- `.planning/phases/08-header-info-panel/08-CONTEXT.md` D-220 — 5-question security review for chrome content; informs why no chrome-content copy bindings are added in Phase 11
- `.planning/phases/09-keybinding-discoverability/09-CONTEXT.md` D-309 — single-source-of-truth menuHints dispatcher; D-507's recipient form sanity teatest verifies this for nested form states
- `.planning/phases/10-theming-accessibility/10-CONTEXT.md` D-403 — `resolveLogoState()` is a pure function of state per frame; D-502 confirms logoStatus is NOT in the cache key (the flash-driven-by-state-mutation invalidation suffices)
- `.planning/phases/10-theming-accessibility/10-CONTEXT.md` D-411-D-412 — typed flash API + severity-tinted bg; D-507's clipboard race teatest verifies the typed API + clipboard timeout interaction

### Research — v1.1 milestone (MUST READ)

- `.planning/research/PITFALLS.md` §"Pitfall 2: Chrome Renders on Every View() Call; Logo / Menu Rebuild Cost Amortizes Into Input Latency" — primary cache mitigation source; D-501-D-503 trace back here. Key prescription: "cache the rendered logo/menu strings on the model when the inputs change, not on every View call"
- `.planning/research/PITFALLS.md` §"Pitfall 10: Alt-Screen + Chrome Interactions With VSCode Integrated Terminal and SSH Clients" — Phase 11 SC3 + SC4 source; the 8-combo matrix originates here. Linux self-sweep selects 4 of those combos; community closes the rest
- `.planning/research/PITFALLS.md` §"Looks Done But Isn't Checklist" (lines 557-573) — the 15-item checklist that D-514 rolls up. Each row in the SC5 sign-off table corresponds to one bullet here. Item #3 (skin fail-open) is the N/A row per D-515
- `.planning/research/PITFALLS.md` §"Recovery Strategies" — informs the fallback escalation in D-517 (cache + allocation hygiene if cache alone misses budget)
- `.planning/research/PITFALLS.md` §"Pitfall-to-Phase Mapping" — confirms Phase 11 owns Pitfall 2 (cache), Pitfall 10 (alt-screen + compat), and the 15-item closure
- `.planning/research/SUMMARY.md` §"Phase 11" — original phase scope before ROADMAP refinement
- `.planning/research/ARCHITECTURE.md` §"Pattern 3: Cache-on-Event for Chrome Inputs" (if the doc has this pattern; otherwise the Phase 8 D-213 pattern in CONTEXT.md is the reference) — informs D-503 mutate-on-event discipline
- `.planning/research/STACK.md` §"Bubble Tea v2 / lipgloss/v2" — `tea.View` shape, `AltScreen=true`, `tea.Quit` semantics

### Existing implementation (Phase 11 modifies / extends)

- `internal/app/model.go` — `AppModel` struct (Phase 11 adds `chromeCache string`, `chromeCacheKey chromeKey`, `chromeCrumbsCache string`, `quitting bool`); `Update` (Phase 11 adds `refreshChromeCache()` calls + `m.quitting = true` on Quit branches); `View` (Phase 11 adds cache-read fast path + quitting branch). `chromeKey` struct declared in same file or new `internal/app/chromecache.go` per planner discretion
- `internal/app/chrome_test.go:284-326` — `TestBenchmarkAppView_UnderBudget` definition; Phase 11 removes `t.Skip` at line 311
- `internal/app/bench_test.go:18-33` — `BenchmarkAppView` at 200×60; Phase 11 adds doc comments documenting empirical baseline + post-cache target + cache hit rate dependency. No fixture changes
- `internal/app/severity_test.go` — Phase 10 truth-table tests; SC5 re-runs as part of gate-flip
- `internal/app/profile_matrix_test.go` — Phase 10 4-profile teatest matrix; SC5 re-runs as part of gate-flip
- `internal/app/resize_test.go` — Phase 10 6-width tests; SC5 re-runs as part of gate-flip
- `internal/app/menuhints_drift_test.go` — Phase 9 menuHints drift detector; SC5 re-runs as part of gate-flip
- `internal/app/model_clipboard_test.go` — Phase 4 clipboard tests; cited in SC1 inventory matrix
- `internal/app/model_reveal_test.go` — Phase 3 reveal tests; cited in SC1 inventory matrix
- `internal/app/model_test.go` — Phase 1+ AppModel construction tests; `TestAppModelAltScreen` confirms AltScreen=true; cited in SC4 evidence
- `internal/app/layout_test.go` + `internal/app/hints_test.go` + `internal/app/export_test.go` — Phase 6/7/8/9 layout, hints, export-API tests; cited in SC1 inventory
- `internal/app/regression_test.go` (NEW) — 3 chrome-interaction sanity teatests per D-507
- `internal/sops/discoverer_test.go` + `internal/sops/executor_test.go` — Phase 1+2 backend tests for file discovery + sops subprocess; cited in SC1 inventory matrix (capability: file discovery, reveal, edit, rotate)
- `internal/parser/yaml_test.go` — Phase 2 YAML parsing tests; cited in SC1 inventory
- `internal/git/status_test.go` — Phase 4 git tests; cited in SC1 inventory (capability: git)
- `internal/health/checker_test.go` — Phase 5 health check tests; cited in SC1 inventory (capability: health)
- `internal/keys/bindings_test.go` + `internal/keys/hints_test.go` + `internal/keys/bindings_reveal_test.go` — Phase 1/3/9 keymap tests; cited in SC1 inventory
- `internal/ui/{statusbar,filelist,detail,detail_reveal,help,health,history,recipientform,metadata,diff,chrome,crumbs,menu,errorbox,rotate,styles}_test.go` — Phase-by-phase UI sub-model tests; cited in SC1 inventory across all 9 capabilities
- `internal/ui/submodel_view_no_newstyle_test.go` — Phase 7.1 NewStyle BFS walker for sub-models; SC5 re-runs as part of gate-flip
- `cmd/sops-tui/main.go` — entry point; Phase 11 makes NO changes (alt-screen exit via tea.View{} on m.quitting; enter trusted to Cursed Renderer + zero-state guard)
- `README.md` — Phase 11 adds "Verified Terminals" H2 section (planner discretion on exact placement) — ~15 lines
- `.github/ISSUE_TEMPLATE/terminal-bug.yml` (NEW) — GitHub Forms YAML issue template per D-510

### Technology / external references

- `charm.land/bubbletea/v2` package docs — https://pkg.go.dev/charm.land/bubbletea/v2 — `tea.View`, `tea.Quit`, `View.AltScreen`, `tea.NewProgram`, `tea.WithColorProfile`
- `charm.land/lipgloss/v2` package docs — https://pkg.go.dev/charm.land/lipgloss/v2 — `JoinHorizontal`, `JoinVertical`, `lipgloss.Width`, `Style.Render`
- Go testing.B + testing.Benchmark — https://pkg.go.dev/testing#B — `b.Loop()` (Go 1.26+), `b.ReportAllocs()`, `testing.Benchmark`
- pprof — https://pkg.go.dev/runtime/pprof — `runtime/pprof` for the Plan 1 SUMMARY profile snippet (allocations no longer in `RenderChrome` / `RenderMenu` after cache wires)
- Mode 2026 synchronized output — https://gist.github.com/christianparpart/d8a62cc1ab659194337d73e399004036 — context for Cursed Renderer alt-screen behaviour referenced in Pitfall 10
- VSCode integrated terminal alt-screen issues — https://github.com/microsoft/vscode/issues — historical context for the 1-row-offset issue cited in D-513
- `CLAUDE.md` §"Core TUI Framework" — `tea.KeyPressMsg` interface, `View.AltScreen` field rules
- `CLAUDE.md` §"Testing" — golden file stability via `charmbracelet/x/exp/teatest` + ANSI-stripped comparison helper; D-507 sanity teatests follow this pattern

### k9s visual parity references (project memory: hard quality attribute)

- `~/git/k9s/internal/ui/menu.go:30-90` — k9s menu hydration pattern; informs D-503 cache-on-event vs cache-on-render discipline (k9s is retained-mode but the cache-on-state-change principle ports to the immediate-mode tea contract)
- `~/git/k9s/internal/view/app.go:540-580` — k9s alt-screen leave / shell prompt return; informs D-512 blank-frame-on-exit pattern (k9s relies on tcell's Suspend/Resume; Bubble Tea v2's tea.View{} is the equivalent)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `internal/app/model.go` `m.infoPanel` cache field + 4 refresh seams (Phase 8 D-213) — D-503 follows this pattern verbatim. The infoPanel cache invalidates on `FilesDiscoveredMsg`, `GitStatusMsg`, recipient ops, edit success; chromeCache adds invalidation seams at the 4 D-502 key fields (state transitions, search toggle, width, recipientAction).
- `internal/app/model.go:1364-1369` zero-state guard — early-returns `tea.NewView("").AltScreen=true` when `m.width == 0 || m.height == 0`. D-513 reuses this as the alt-screen enter fill frame; D-512 extends it with a `m.quitting` branch for the alt-screen exit blank frame.
- `internal/app/chrome_test.go:284-326` `TestBenchmarkAppView_UnderBudget` — bench gate already in code form; Plan 1 just removes the `t.Skip` line. The 50,000 ns budget at line 317 stays.
- `internal/app/bench_test.go:18-33` `BenchmarkAppView` — 200×60 fixture per Phase 6 D-12; Plan 1 doesn't change the fixture, only adds doc comments.
- `internal/testutil/golden.go` `RequireGoldenStructure` + `RequireGoldenColors` — Phase 6 + 10 golden helpers; D-507 sanity teatests use ANSI-stripped output (Pattern: ANSI-stripped structure + targeted color assertions).
- `internal/app/severity_test.go` truth-table tests + `internal/app/profile_matrix_test.go` 4-profile matrix + `internal/app/resize_test.go` 6-width matrix + `internal/app/menuhints_drift_test.go` drift detector — all live tests that the SC5 verifier re-runs.
- `internal/keys/bindings.go` `keys.DefaultGlobalKeyMap.Quit` — single Quit key binding; Plan 1's `m.quitting = true` flag attaches to the Update branch matching this binding.
- `internal/app/model.go` `AppModel.IsClipboardHot()` accessor — Phase 4 export for clipboard testing; D-507's clipboard race teatest reads through this.
- `internal/app/model.go` `app.ClipboardTimeout` test seam — Phase 4 var that overrides the timeout; D-507's clipboard race teatest sets this short to drive the race.
- `internal/keys/hints.go` `keys.RecipientFormHints` — Phase 9 D-309 form-level hints; D-507's recipient form teatest asserts these mnemonics appear in the rendered menu.

### Established Patterns

- **Pure functions for renderers + styles as package vars** — `internal/ui/*.go` exposes `RenderX(...) string`; `TestViewNoNewStyle` BFS walker enforces no `NewStyle()` reachable from View. D-501 cache stores rendered strings; the renderers stay pure.
- **Cache-on-event, never-on-render** — Phase 8 `infoPanel` cache pattern (Pitfall 15 prescription); D-503 follows verbatim. Update mutates, View reads.
- **value-receiver discipline** — every AppModel method is value-receiver; cache mutation flows through `Update` returning a new model. D-501-D-503 preserve this; the `quitting` flag follows the same shape.
- **Test scaffolding co-located with implementation** — `chrome_test.go` next to `chrome.go` etc. D-505 puts the cache hit-rate test in `chrome_test.go`; D-507 puts the regression sanity teatests in a new `regression_test.go` per "regression suite is its own concern" convention.
- **GOLDEN_UPDATE wave commits** — Phases 7/8/10 each had a goldens-refresh commit at end of integration plan. Phase 11 has no palette / structural changes, so no GOLDEN_UPDATE wave. The cache pure-function discipline guarantees zero structural drift.
- **Plan SUMMARY documents pre/post measurements** — Phase 7 D-15, Phase 8 D-213, Phase 10 D-415 SUMMARYs all captured before/after numbers. Plan 1 SUMMARY captures the 2.4-2.8 ms → <50 µs transition + cache hit rate at steady-state.

### Integration Points

- `internal/app/model.go` `AppModel` struct (around line 270-300) — Plan 1 adds `chromeCache`, `chromeCacheKey`, `chromeCrumbsCache`, `quitting` fields. Constructor `NewAppModel` initializes `chromeCacheKey` to a zero-value sentinel (e.g. `chromeKey{state: stateUnknown}`) so the first View call always misses the cache and populates it.
- `internal/app/model.go` `Update` mutation sites — Plan 1 audits these for `refreshChromeCache()` insertion:
  - `WindowSizeMsg` handler (around model.go:313-328)
  - State transition assignments (`m.state = stateXxx`) — search returns ~20 sites
  - Search toggle (FileList Update propagation flips `m.fileList.IsSearchActive()`)
  - Recipient action assignments (`m.recipientAction = "..."`) — search returns ~6 sites
  - Quit branch (`return m, tea.Quit`) — Plan 1 also adds `m.quitting = true` here
  Pattern: `m = m.refreshChromeCache(); return m, cmd`
- `internal/app/model.go` `View()` (around line 1364-1427) — Plan 1 adds:
  - Top: `if m.quitting { v := tea.NewView(""); v.AltScreen = true; return v }`
  - After the existing zero-state guard: `if m.chromeCacheKey == m.computeChromeKey() { /* read cache, JoinVertical with body+statusBar, return */ }`
  - Else: existing path runs (RenderChrome + RenderCrumbs + JoinVertical), but the result is also stored back via `refreshChromeCache` in the next Update — wait, no: View is value-receiver, can't mutate. Plan 1 needs to ensure `refreshChromeCache` is called in Update *before* View runs; the View path then either uses the cache or, on first frame after a key-changing Update, the cache is already fresh.
- `internal/app/chrome_test.go:308-326` `TestBenchmarkAppView_UnderBudget` — Plan 1 deletes line 311 `t.Skip(...)`. Test executes `testing.Benchmark(BenchmarkAppView)` and gates on `nsPerOp > 50_000`.
- `internal/app/regression_test.go` (NEW) — Plan 2 creates this file with 3 sanity teatests per D-507. Uses existing `teatest` + `internal/testutil/golden.go` helpers.
- `cmd/sops-tui/main.go` — NO changes. The signal handler at lines 56-60 (clipboard cleanup + os.Exit(0)) is preserved; the alt-screen ANSI exit sequence is emitted by the renderer's deferred shutdown hook before the process exits, even on signal path.
- `README.md` — Plan 2 adds "Verified Terminals" H2 section (planner discretion on exact placement). Markdown table format.
- `.github/ISSUE_TEMPLATE/terminal-bug.yml` (NEW) — Plan 2 creates this file. GitHub Forms YAML schema; required fields per D-510.
- `.planning/phases/11-regression-perf-gates/screenshots/{alacritty,ghostty,tmux-nested,vscode-integrated}.png` (NEW) — Plan 2 captures these from manual sweep.
- `.planning/phases/11-regression-perf-gates/11-VERIFICATION.md` (NEW) — Built by `/gsd-verify-work`. Contains 9-row SC1 inventory + SC2/SC3/SC4 evidence sections + 15-row SC5 sign-off table.
- `go.mod` — NO changes. All caching machinery uses stdlib + existing deps.

</code_context>

<deferred>
## Deferred Ideas

### v1.1.x patches (post-v1.1 release, if user reports surface)

- **Per-renderer caches** — if profiling after v1.1 ships shows the full-string cache has a low hit rate at certain workloads, splitting into per-renderer caches with finer-grained invalidation may help. Currently rejected because the empirical baseline + held-down-j workload is the dominant case.
- **Allocation hygiene + manual menu columns** — if Plan 1's cache-only approach misses 50 µs on dev hardware, the fallback documented in D-517 + D-518 escalates within the same plan. If the cache alone hits, this stays as a documented future option.
- **Generation counter for cache keys** — if the 4-field minimum key proves too narrow (e.g. a future phase adds a chrome input the planner forgot to add to the key), a `chromeGen uint64` bump-on-mutation pattern is the documented fallback.
- **Explicit `FillFrameStyle` package var for alt-screen enter** — if any of the 4 Linux verification combos (especially VSCode integrated terminal) shows residual content on first frame, Plan 2 closes it via a declared `lipgloss.NewStyle().Background(ColorBg).Width(w).Height(h).Render("")` package var (not in View() — TestViewNoNewStyle BFS walker would catch it). Currently rejected because zero-state guard + Cursed Renderer suffice on tested combos.
- **Post-`p.Run()` ANSI reset in main.go** — if `tea.View{}` on quit doesn't fully clean up on some terminals (rare but possible on tmux-nested), Plan 2 adds `fmt.Print("\x1b[2J\x1b[H")` after `p.Run()` returns. Currently rejected because tea.View{} keeps the cleanup contract inside the renderer.
- **Signal-handler blank frame** — if SIGINT/SIGTERM exit leaves chrome residue on any tested combo, the signal handler at main.go:56-60 gains a blank-frame ANSI write before `os.Exit(0)`. Currently rejected because the renderer's deferred shutdown emits the alt-screen leave sequence even on signal path.
- **Header info-panel re-fetch on SOPS error** (Pitfall 17 — chrome stale during async ops) — currently the infoPanel cache is refreshed on event seams (Phase 8 D-213); if a user reports a stale fingerprint after a key-rotation flash, this becomes a v1.1.x patch adding `checkEnvAsync` re-trigger after sops-error messages.
- **macOS Terminal / iTerm2 / Windows Terminal / WSL2 manual sweep** — closes via community contributions (D-510 README + issue template). If any combo accumulates bug reports, treat as a v1.1.x patch with explicit fixes for the reported symptoms.

### v2 (milestone-deferred per ROADMAP)

- **User-facing skin YAML loader** (`THM-01`) — `~/.config/sops-tui/skin.yaml` with k9s-compatible schema subset; closes Pitfall 4 ("skin fail-open") which Phase 11 SC5 marks `[N/A]` per D-515
- **Builtin skins embedded via `embed.FS`** — dracula / gruvbox-dark / monokai (`THM-02`)
- **Live skin reload via fsnotify** (`THM-03`) — adds goroutine + fsnotify dependency; explicitly OOS for v1.1
- **Multi-pane focus rings** — Pitfall 7 prevention; v1.1 keeps single-pane discipline
- **`:` command bar** — v1.2 per ROADMAP
- **Number-key view switching / `[`/`]` history navigation** — v1.2+
- **Mouse interactions** — v2; project core value is keyboard-only
- **Splash screen / big logo on startup** — explicitly OOS per ROADMAP
- **Image vulnerability scanner / port-forward manager / cluster context switcher** — k9s-specific anti-features per research/SUMMARY.md

### Possibly Phase 11 Plan 2, possibly v1.1.x

- **README "Verified Terminals" matrix exact format** — Plan 2 author picks; recommendation: H2 section with markdown table + footer link to issue template
- **Issue template field shape** — Plan 2 author picks; recommendation: GitHub Forms YAML
- **Whether `m.quitting` wiring lands in Plan 1 or Plan 2** — D-518 note; recommendation Plan 2 to keep the SC2-only Plan 1 split clean, but planner may move to Plan 1 if Update audit hits the same branches anyway

### Out of scope this phase (would be scope creep)

- E2E teatest flows for all 9 v1.0 capabilities (vs the 3 chrome-prone sanity checks) — SC1 says "pass unchanged" not "extend coverage"
- Tagged `TestRegression_*` build tag + Makefile target — over-ceremony for one milestone closure
- Standalone `REGRESSION-MATRIX.md` artifact — VERIFICATION.md is sufficient
- Standalone `CHECKLIST.md` artifact — VERIFICATION.md SC5 section is sufficient
- Mutating `.planning/research/PITFALLS.md` with `[x]` inline — research artifacts stay snapshots; closure lives in 11-VERIFICATION.md
- Two-table SC5 split (`Verified` + `Deferred to v2`) — one 15-row table with explicit `[N/A]` rows is more honest
- VM-based macOS / Windows verification before sign-off — out-of-codebase effort that pushes the milestone for marginal coverage gain
- New chrome features added to the cache (dynamic logo width-responsive trimming, info-panel cache-busting on git-status mid-scroll) — Phase 11 wires the cache for the chrome that exists today
- ROADMAP SC2 wording change (loosen the 50 µs target) — Phase 7.1 governance lock; original target preserved

### Reviewed Todos (not folded)

- "Phase 10/11: revisit BenchmarkAppView budget — currently 5 ms with 56% headroom over ~2.8 ms/op measurement; D-18 caching fallback (model-level cache keyed on (state, recipientAction, IsSearchActive, width)) can tighten this if user-perceived latency matters" (STATE.md) — folded entirely into D-501-D-504. Phase 11 wires the D-18 cache and flips the 50,000 ns gate. The note about "5 ms with 56% headroom" referenced the temporary Phase 7 looser test gate that Phase 7.1 already restored to 50 µs/op via t.Skip; Phase 11 closes the loop by removing the t.Skip.
- "Manual UAT per Phase 06 D-15" (STATE.md) — terminal-resize verification, addressed by D-509 Linux self-sweep on the 4 combos. The manual UAT has been satisfied across Phases 6-10 via the resize golden matrix; Phase 11 adds the wall-clock + visual sweep against actual terminal emulators.

</deferred>

---

*Phase: 11-regression-perf-gates*
*Context gathered: 2026-05-04*
