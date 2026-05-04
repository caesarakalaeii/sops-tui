# Phase 11: Regression + Perf Gates - Research

**Researched:** 2026-05-04
**Domain:** Bubble Tea v2 / lipgloss/v2 chrome cache wiring + regression sanity teatests + Linux compat sweep + alt-screen exit + 15-row sign-off
**Status:** Ready for planning
**Confidence:** HIGH (CONTEXT.md is exhaustive — this research only fills tactical gaps)

## Approach Summary

Phase 11 closes UI-20 + UI-21 via two plans whose decisions are LOCKED in CONTEXT.md (D-501..D-518) — this research deliberately adds NO new decisions, only verifies the technical patterns and surfaces the landmines.

Plan 1 wires a four-field struct cache key per D-501 + D-502, populates it in `Update` per D-503 (Phase 8 D-213 mutate-on-event pattern), and flips the `t.Skip` on chrome_test.go:311 per D-504 — note that the t.Skip is **the FIRST line** of the test body and must be removed entirely (not converted to a conditional skip), and the existing `testing.Short()` skip on line 312 stays as the only remaining skip path. The cache invalidates implicitly: View() compares `m.computeChromeKey() == m.chromeCacheKey` and falls through to the renderer if they differ, but **the cache itself is mutated only inside `Update` via `refreshChromeCache()`** — Pitfall noted in [Pitfalls](#pitfalls) below because View()'s value receiver cannot mutate state without violating Pitfall 2's "no NewStyle in View" sister rule (no state mutation in View).

Plan 2 adds three Update-loop sanity tests in `internal/app/regression_test.go` per D-507, wires `m.quitting = true` on the Quit branch at model.go:993 per D-512 (single Quit site — not a sweep), captures four Linux screenshots per D-509, and produces the README "Verified Terminals" matrix + GitHub issue template per D-510. Critical clarification: this project does **NOT** use the `charmbracelet/x/exp/teatest` framework — `teatest` is not in go.mod. The "3 sanity teatests" in CONTEXT.md D-507 use the project's existing Update-loop pattern (`m.Update(msg)` + assertions on `View().Content` — see `internal/app/model_clipboard_test.go` for the canonical shape). [Test Patterns](#test-patterns) section below contains the fixture skeleton.

## Validation Architecture

> Every must-have ID maps to a verification method below. Plan 1 owns SC2 evidence; Plan 2 owns SC1 + SC3 + SC4 + SC5 evidence.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + `testing.Benchmark` (no teatest) |
| Quick run command | `go test ./internal/app/ -run 'TestRegression\|TestChromeCache\|TestBenchmarkAppView_UnderBudget' -count=1` |
| Full suite command | `go test ./... -count=1` |
| Bench-only | `go test ./internal/app/ -bench=BenchmarkAppView -benchmem -count=3` |
| Phase gate | Full suite green before `/gsd-verify-work`; bench under 50,000 ns/op |

### SC1 (UI-20) — v1.0 Functional Non-Regression

| Evidence | Verification | Where |
|----------|------------|-------|
| 9-capability inventory matrix | Hand-curated table built at `/gsd-verify-work` time | 11-VERIFICATION.md SC1 (D-506) |
| All v1.0 tests pass post-chrome | `go test ./... -count=1` exit code = 0 | Verifier captures pass count |
| TestRegression_ClipboardAutoClearWithChrome | Drives `ctrl+y` + `ui.FlashClearMsg{Gen: 1}` + `ClipboardClearMsg`; asserts `m.IsClipboardHot() == false`, `[clip]` absent in `View().Content`, no `[W]`/`[E]` prefix | `internal/app/regression_test.go` (NEW) |
| TestRegression_RecipientFormMenuHints | Drives into `stateRecipientForm`; asserts `ansi.Strip(View().Content)` contains `Tab`, `Enter`, `Esc`; asserts NO substring match on `[j]`, `[k]`, `[q]` mnemonics on the menu rows | `internal/app/regression_test.go` (NEW) |
| TestRegression_HealthOverlayOnNarrowWidth | Two sub-tests: `WindowSizeMsg{Width: 80, Height: 24}` and `{Width: 60, Height: 24}`; drive into `stateHealth`; assert `<health>` chip present in stripped output (D-425 contract) | `internal/app/regression_test.go` (NEW) |

### SC2 (UI-21) — Render Performance + Zero NewStyle in View

| Evidence | Verification | Where |
|----------|------------|-------|
| Bench under 50 µs/op | `t.Skip` removed from `chrome_test.go:311`; `TestBenchmarkAppView_UnderBudget` runs `testing.Benchmark(BenchmarkAppView)` and asserts `result.NsPerOp() <= 50_000` | `internal/app/chrome_test.go:310-326` (D-504) |
| TestChromeCache_HitRateAtSteadyState | Drives 100 sequential `View()` calls with NO `Update` between them; asserts cache rebuilt at most 1× across 100 frames (`hitCount >= 99`) | `internal/app/chrome_test.go` (NEW per D-505) |
| Zero NewStyle in View reachables | Existing `TestViewNoNewStyle` BFS walker passes (Phase 7.1 walker re-runs at verifier time per D-516) | `internal/app/chrome_test.go:161-282` (existing) |
| pprof shows RenderChrome NOT in hot path | Run `go test -bench=BenchmarkAppView -cpuprofile=cpu.out`, then `go tool pprof -top cpu.out`; capture top-10 list; verifier paste-includes in 11-VERIFICATION.md SC2 | Verifier output (D-506) |

### SC3 (chrome on Linux terminal matrix) — Manual Sweep

| Evidence | Verification | Where |
|----------|------------|-------|
| 4× PNG screenshots @ 200×60 in stateFileList | Manual capture by developer | `.planning/phases/11-regression-perf-gates/screenshots/{alacritty,ghostty,tmux-nested,vscode-integrated}.png` (D-511) |
| Per-combo checklist | Resize 80×24 ↔ 200×60 (no flicker), alt-screen enter cleanly, alt-screen exit cleanly, q press returns shell with cursor at expected position, ctrl-c returns shell with clipboard cleared (D-509) | 11-VERIFICATION.md SC3 |
| README "Verified Terminals" matrix | Markdown table: 4 Linux combos verified-by-author; 4 community combos with issue-template link (D-510) | `README.md` H2 section between Installation and Usage |
| GitHub issue template | YAML at `.github/ISSUE_TEMPLATE/terminal-bug.yml` with required fields: terminal name, version, OS, screenshot, expected, observed, repro (D-510) | `.github/ISSUE_TEMPLATE/terminal-bug.yml` (NEW) |

### SC4 (alt-screen cleanup) — `m.quitting` Flag

| Evidence | Verification | Where |
|----------|------------|-------|
| `m.quitting bool` field on AppModel | Added per D-512; zero-value `false` | `internal/app/model.go:225` struct |
| Quit branch sets flag | `model.go:993` becomes `m.quitting = true; return m, func() tea.Msg { return tea.Quit() }` | Single-site change |
| View top branch returns blank | `if m.quitting { v := tea.NewView(""); v.AltScreen = true; return v }` slotted ABOVE the existing zero-state guard at model.go:1365 | Single-site change |
| Manual SC3 sweep observation | "On q press the user's shell prompt area shows no chrome residue" recorded per-combo (D-509 checklist row) | 11-VERIFICATION.md SC4 |

### SC5 (15-row "Looks Done But Isn't" sign-off) — Re-run Gates + Cite Prior Phases

| Evidence | Verification | Where |
|----------|------------|-------|
| 15-row table in `11-VERIFICATION.md` SC5 | One row per PITFALLS.md line 559-573 bullet; columns: # / item / status / evidence / prior phase | D-514 |
| `[N/A]` row for skin fail-open | Item #3 marked `N/A` with v2 deferral evidence text (D-515) | Row 3 of table |
| Re-run gate sweep | Verifier executes `go test ./... -count=1 -run 'TestChromeASCIIOnly\|TestChromeNormalBorderOnly\|TestViewNoNewStyle\|TestSubmodelViewsNoNewStyle\|TestMenuHints_Drift\|TestRenderCrumbs_FirstAndLastSegmentsPreserved\|TestResize_'` and pastes pass/fail count | D-516 |
| Cited prior phase evidence | Each `Done` row includes `Phase N VERIFICATION.md SC{X}` reference | D-516 |

### Sampling Rate

- **Per task commit:** `go test ./internal/app/ -count=1` (~5 seconds; covers chrome_test, regression_test, severity_test, profile_matrix, resize)
- **Per wave merge:** `go test ./... -count=1` (~30 seconds; full suite including ui sub-models, sops, parser, git, health, keys)
- **Per plan close:** Add `go test ./internal/app/ -bench=BenchmarkAppView -benchmem -count=3` and capture pre/post numbers in PLAN-SUMMARY.md
- **Phase gate:** Full suite green; bench gate passes; manual sweep complete (SC3 screenshots committed)

### Wave 0 Gaps

- [ ] `internal/app/regression_test.go` — covers UI-20 (3 chrome-interaction sanity teatests per D-507)

*(All other test infrastructure exists. No framework install needed. teatest is intentionally NOT used; project uses Update-loop pattern.)*

## Technical Patterns

### Pattern 1: Cache Key as Struct (D-501)

`[VERIFIED: codebase grep — Phase 8 D-213 infoPanel cache uses identical mutate-on-event shape]`

Go map keys with structs hash directly without allocations. The 4-field key in D-501 compiles to a single 24-byte struct (assuming `sessionState` is `int`/`uint8`, `recipientAction` is `string` 16 bytes, `searchActive` is `bool` 1 byte, `width` is `int` 8 bytes; with padding ~24-32 bytes). Compared via `==` operator at zero allocation cost.

```go
// internal/app/model.go (slot near AppModel struct)
type chromeKey struct {
    state           sessionState
    recipientAction string
    searchActive    bool
    width           int
}

// AppModel struct gains four new fields (per D-501 + D-512)
type AppModel struct {
    // ... existing fields ...
    chromeCache       string    // Phase 11 D-501
    chromeCrumbsCache string    // Phase 11 D-501 (recommended split per discretion note)
    chromeCacheKey    chromeKey // Phase 11 D-501
    quitting          bool      // Phase 11 D-512
}

// Helper computes the current key without allocating
func (m AppModel) computeChromeKey() chromeKey {
    return chromeKey{
        state:           m.state,
        recipientAction: m.recipientAction,
        searchActive:    m.fileList.IsSearchActive(),
        width:           m.width,
    }
}
```

**Cite:** Lock CONTEXT.md D-501 verbatim — 4-field struct, not `fmt.Sprintf` concatenation key. Sprintf would allocate per Update, defeating the cache's purpose.

### Pattern 2: Mutate-on-Event Helper (D-503)

`[VERIFIED: codebase — Phase 8 model.go FilesDiscoveredMsg / FilesParsedMsg / GitStatusMsg already follow this pattern for infoPanel]`

```go
// refreshChromeCache rebuilds chromeCache + chromeCrumbsCache if the key
// has changed since last refresh. Called from every Update branch that
// mutates a key field. Pattern matches Phase 8 D-213 infoPanel cache.
func (m AppModel) refreshChromeCache() AppModel {
    newKey := m.computeChromeKey()
    if newKey == m.chromeCacheKey {
        return m // no-op fast path: key unchanged
    }
    m.chromeCacheKey = newKey
    hints := m.menuHints() // ← already pure function of state
    m.chromeCache = ui.RenderChrome(hints, m.resolveLogoState(), m.infoPanel, m.palette, m.width)
    m.chromeCrumbsCache = ui.RenderCrumbs(m.status.Segments(), m.palette, m.width)
    return m
}
```

**Mutation sites Plan 1 audits (~25 sites, NOT 42 — recipient action sweep is a separate axis):**

- `model.go:325-328` — `WindowSizeMsg` (width changes)
- `model.go` — every assignment `m.state = state*` (search returns ~20 sites)
- `model.go:796`, `model.go:822`, `model.go:885`, `model.go:1144`, `model.go:1283` — `m.recipientAction = ...` (5 sites confirmed via grep)
- `model.go:1052-1075` — search-toggle paths via `m.fileList.IsSearchActive()` propagation (the FileList Update is the canonical seam; `refreshChromeCache` after the FileList Update covers this without per-site calls)

**Pattern (per CONTEXT.md):** `m = m.refreshChromeCache(); return m, cmd`

### Pattern 3: View Reads Cache Only (D-503)

```go
func (m AppModel) View() tea.View {
    // Phase 11 D-512: alt-screen exit blank frame.
    // m.quitting is set in the Quit handler BEFORE returning tea.Quit;
    // the next (and final) View() call returns blank with AltScreen=true,
    // so Cursed Renderer's last frame before alt-screen leave is empty.
    if m.quitting {
        v := tea.NewView("")
        v.AltScreen = true
        return v
    }

    // Existing zero-state guard (Phase 7 Pitfall 5 first-frame safety)
    if m.width == 0 || m.height == 0 {
        v := tea.NewView("")
        v.AltScreen = true
        return v
    }

    // ... body switch (unchanged) ...

    // Phase 11 D-503: read cache populated by Update.
    // chromeCacheKey is guaranteed fresh because every Update branch
    // that mutates a key field calls refreshChromeCache() before return.
    chrome := m.chromeCache
    crumbs := m.chromeCrumbsCache
    statusBar := m.status.View(m.width)
    sections := []string{chrome, crumbs, wrapped, statusBar}
    full := lipgloss.JoinVertical(lipgloss.Left, sections...)

    v := tea.NewView(full)
    v.AltScreen = true
    return v
}
```

**Critical contract:** View() never calls `RenderChrome` / `RenderCrumbs` directly. Both are already cached. Status bar still renders per-frame because flash text changes mid-tick (FlashClearMsg arrives after 2s) and `m.status.View(width)` is cheap.

### Pattern 4: tea.View AltScreen Field (D-512 + D-513)

`[CITED: charm.land/bubbletea/v2 docs + CLAUDE.md "View struct fields"]`

In Bubble Tea v2, `WithAltScreen()` program option is removed. Alt-screen control moved to per-frame `tea.View.AltScreen bool`. The Cursed Renderer enters/exits the alt-screen based on the value of this field on the current frame. To exit cleanly:

1. The frame where `AltScreen=true` and content is empty paints a blank screen.
2. The next frame where `AltScreen=false` (or program exits) emits `ESC[?1049l` and returns control to the shell.

**With `tea.Quit`:** The `tea.Quit` command shuts down the program. Order matters:
- `m.quitting = true` MUST be set **before** returning `tea.Quit`. The View() call between Update and program-shutdown is the one that paints the blank frame.
- Returning `tea.Quit` first then setting `m.quitting` in a follow-up message is too late — the program has already begun shutdown.

**Existing Quit site (single):** `model.go:993` returns `m, func() tea.Msg { return tea.Quit() }`. Phase 11 changes to:

```go
if key.Matches(msg, keys.DefaultGlobalKeyMap.Quit) {
    m.quitting = true
    return m, func() tea.Msg { return tea.Quit() }
}
```

`[ASSUMED]` There is exactly ONE `tea.Quit()` site in `model.go` (verified via grep at line 993). Future code that adds a second Quit site (e.g. modal-confirm-then-quit) MUST also set `m.quitting = true` first. Document in a comment near the field declaration.

### Pattern 5: Initial Cache Key Sentinel

The first `View()` call must NOT hit the cache (the cache is empty). Two approaches:

1. **Recommended (CONTEXT.md "Integration Points" line 252):** Initialize `m.chromeCacheKey` to `chromeKey{state: stateUnknown}` in `NewAppModel`, where `stateUnknown` is a sentinel value distinct from `stateFileList` (the real first state). The first `Update` branch (typically `WindowSizeMsg`) calls `refreshChromeCache`, which sees the mismatch and populates the cache.

2. **Alternative:** Rely on `m.width == 0` zero-state guard at View top — View returns blank before reading cache; Update populates cache on first WindowSizeMsg before any non-zero-width View runs. This works because zero-value `chromeKey{}` has `width=0`, and the zero-state guard returns first.

**Recommendation:** Approach 2 (no sentinel) — leverages existing zero-state guard, no new sessionState constant. CONTEXT.md "Integration Points" mentions sentinel but the existing guard makes it unnecessary.

### Pattern 6: Bench Measurement (D-504)

`[CITED: pkg.go.dev/testing#Benchmark + bench_test.go:30 b.Loop()]`

`testing.Benchmark(fn func(*testing.B)) BenchmarkResult` runs the bench function and returns a `BenchmarkResult` whose `NsPerOp() int64` is the average ns per iteration. The existing `chrome_test.go:315` calls `testing.Benchmark(BenchmarkAppView)` and asserts `result.NsPerOp() > 50_000` is a regression.

**Bench test body (`bench_test.go:18-33`) is unchanged:** Constructs AppModel, sends one WindowSizeMsg{200, 60}, then loops `_ = m.View()` via `b.Loop()`. Plan 1 adds doc comments documenting the empirical baseline + post-cache target + cache hit rate dependency, but the body stays byte-identical.

**Cache hit-rate test mechanism (D-505):**

```go
// Drives 100 sequential View() calls without any Update between them
// (simulates held-down j frame burst — but without the Update we'd
// expect a 100% cache hit rate). Asserts the cache key is stable
// (matched 99/100 frames). Survives Go version bumps because it
// measures hit rate, not wall-clock.
func TestChromeCache_HitRateAtSteadyState(t *testing.T) {
    env := ui.EnvStatus{SopsAvailable: true, AgeAvailable: true, SopsYamlAvailable: true, GitAvailable: true}
    m := NewAppModel(env, "", colorprofile.TrueColor)
    updated, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 60})
    m = updated.(AppModel)

    // After the WindowSizeMsg, m.chromeCacheKey is populated.
    keyAfterFirst := m.chromeCacheKey

    // Drive 100 View() calls. View is value-receiver — the returned
    // tea.View is consumed but m is unchanged. Capture the cache key
    // each iteration; it must NOT change because no Update ran.
    for i := 0; i < 100; i++ {
        _ = m.View()
        if m.chromeCacheKey != keyAfterFirst {
            t.Fatalf("cache key drifted at iteration %d: View() mutated cache "+
                "(expected View to be a pure read; cache must be populated by Update)", i)
        }
    }
}
```

`[VERIFIED: code shape]` This is the proper measurement: the assertion is "cache key stays equal across 100 frames", proving the cache is *populated by Update* (not by View). Plan 1 author adds a `chromeRebuildCount` debug counter (test-only, gated behind `_test.go`) IF the planner wants to literally count rebuilds — but the simpler "key stability" assertion is sufficient and survives any future refactor of `refreshChromeCache`.

## Pitfalls

> Specific landmines for Phase 11 — read carefully before planning.

### Pitfall A: View() Cannot Mutate the Cache (Value Receiver Discipline)

**What goes wrong:** A natural-feeling implementation puts the cache check inside `View()`:

```go
// WRONG — value receiver; mutation is silently lost
func (m AppModel) View() tea.View {
    if m.computeChromeKey() != m.chromeCacheKey {
        m.chromeCacheKey = m.computeChromeKey() // ← lost; m is a value
        m.chromeCache = ui.RenderChrome(...)     // ← lost
    }
    return tea.NewView(m.chromeCache + ...)
}
```

The mutation has no effect. Next View() call sees the same stale key and re-renders again. The cache is "wired" (compiles, tests look like they pass) but achieves zero hit rate — bench stays at 2.4-2.8 ms.

**How to avoid:**

1. **Cache mutation flows through `Update` returning a new model** (per D-503). View is pure read-only.
2. The `refreshChromeCache()` helper is a value-receiver method that returns the modified `AppModel`. Update branches assign: `m = m.refreshChromeCache()`.
3. `TestChromeCache_HitRateAtSteadyState` (D-505) is the trip wire — if View accidentally tries to mutate, the cache key stays at the initial value and the test passes anyway (because Update populated it once); but if the planner forgets to call `refreshChromeCache` in an Update branch that should invalidate, the next View hits a stale cache. Plan 1's mutation-site audit (~25 sites) is the fix; Plan 2's 3 sanity teatests are the integration-level proof.

**Warning sign:** Any line inside `View()` that assigns to `m.cache*` or `m.*Key` fields. Run `git grep 'm\.chromeCache\|m\.chromeCacheKey' internal/app/model.go` and confirm every write is inside Update / inside `refreshChromeCache`, never inside View.

### Pitfall B: Cache Key Includes a Field That Shouldn't

**What goes wrong:** Adding `palette`, `logoStatus`, `infoPanelData`, or `flashGen` to the cache key (per CONTEXT.md "Out of scope" reject list) lowers the hit rate without catching real bugs:

- **Palette:** Set once at startup. Never changes mid-session. Adding to the key is a no-op for hit rate but adds a struct field needlessly.
- **logoStatus:** Pure function of state. The flash that *causes* the logo color change has its own state-mutation seam (FlashErr triggers severity flip via `resolveLogoState`); the user notices through the flash bar, not the logo. A flash-only frame should NOT invalidate the cache.
- **infoPanelData:** Refreshed on its own seam (Phase 8 D-213). It changes 4× per session typically. Adding to key triggers chrome rebuild on every git status change — wasted work for zero benefit.
- **flashGen:** Changes every flash. A cache invalidated on flashGen would rebuild the chrome 4× per `Flash()` call (set + clear + tick) — defeats the cache.

**How to avoid:** Lock the cache key to the 4 fields per D-502 verbatim. Treat any planner request to "also include X" as a re-litigation of D-502 — re-route through `/gsd-discuss-phase` if the planner believes a 5th field is needed.

### Pitfall C: `m.quitting` Set AFTER Returning `tea.Quit`

**What goes wrong:**

```go
// WRONG — tea.Quit fires first; the next-frame View() never runs
if key.Matches(msg, keys.DefaultGlobalKeyMap.Quit) {
    cmd := func() tea.Msg { return tea.Quit() }
    m.quitting = true
    return m, cmd
}
```

Actually this works — the Cmd is returned but tea processes the new model BEFORE running the Cmd. So `m.quitting=true` is observed before `tea.Quit` fires. But the more readable form is:

```go
// CORRECT — explicit ordering
if key.Matches(msg, keys.DefaultGlobalKeyMap.Quit) {
    m.quitting = true
    return m, func() tea.Msg { return tea.Quit() }
}
```

**Real gotcha:** A planner who reads "set m.quitting before returning tea.Quit" and writes:

```go
// WRONG — tea.Quit returned synchronously; bubbletea may shut down
// before the View call that would have painted the blank frame
if key.Matches(msg, keys.DefaultGlobalKeyMap.Quit) {
    m.quitting = true
    return m, tea.Quit  // ← tea.Quit is a function, not a value
}
```

`tea.Quit` in bubbletea/v2 is a `func() tea.Msg`. Passing it directly works (it's a Cmd literal). But the existing code at model.go:993 wraps it in another anonymous closure: `func() tea.Msg { return tea.Quit() }`. Don't refactor that during Phase 11 — preserve the existing pattern. Just add `m.quitting = true` above the `return` line.

### Pitfall D: First-Frame Cache Miss After WindowSizeMsg

**What goes wrong:** If `refreshChromeCache` is NOT called in the WindowSizeMsg handler, the first non-zero-width View() runs with `m.chromeCache == ""`. The renderer call inside View (which Plan 1 deletes) used to handle this implicitly. After Plan 1, View() returns an empty chrome string in the joined output — visible glitch on first frame.

**How to avoid:** The WindowSizeMsg handler at model.go:325-328 is the FIRST mutation site to audit. Confirm it ends with `m = m.refreshChromeCache(); return m, cmds`. The width-change is the most reliable cache invalidator (every other key field is post-startup-only).

**Trip wire:** `TestRegression_HealthOverlayOnNarrowWidth` (D-507) sends `WindowSizeMsg{60, 24}` then immediately drives into stateHealth and asserts the chrome row appears. If the first frame's chrome is empty, this test catches it.

### Pitfall E: `IsSearchActive()` Propagation Through FileList Update

**What goes wrong:** The search-active flag is owned by `FileListModel` (`m.fileList.IsSearchActive()`), not by `AppModel`. When the user presses `/` to enter search mode, the FileList Update returns a new FileListModel with `searchActive=true`. AppModel's Update propagates: `m.fileList = newFileList`. But the cache key is computed from `m.fileList.IsSearchActive()` AT THE MOMENT `refreshChromeCache` runs.

If the FileList Update returns BEFORE `refreshChromeCache` is called, the new IsSearchActive value is read correctly. Order of operations in the Update method:

1. Receive KeyPressMsg(`/`)
2. Forward to FileList Update → get new FileList with `searchActive=true`
3. Assign `m.fileList = newFileList`
4. **Call `m = m.refreshChromeCache()` ← reads `m.fileList.IsSearchActive() == true`, key changes, cache rebuilds**
5. Return `m, cmd`

**Pitfall:** If the planner places `refreshChromeCache` BEFORE step 3 (before assigning the new fileList), the cache is rebuilt with stale `searchActive=false`. The next View shows old menu hints (j/k/q instead of Esc/Enter).

**How to avoid:** Place `refreshChromeCache` calls at the END of each Update branch (just before `return`), AFTER all sub-model assignments are done. The CONTEXT.md "Integration Points" pattern `m = m.refreshChromeCache(); return m, cmd` enforces this implicitly.

### Pitfall F: `recipientAction` Field Cleared After Confirm

**What goes wrong:** `m.recipientAction` is set when entering a confirm flow (model.go:796, 822, 885, 1144, 1283) AND **cleared** when the confirm flow ends (model.go:886 sets `m.recipientAction = ""`). The clearing is itself a key-field mutation that requires `refreshChromeCache`.

**How to avoid:** Audit BOTH the assignment sites AND the clearing sites (model.go:886 and any other `m.recipientAction = ""` lines). A simple grep: `git grep -n 'recipientAction = ' internal/app/model.go` returns the full list. Each must be followed by `refreshChromeCache`.

### Pitfall G: Quit Path NOT Going Through `keys.DefaultGlobalKeyMap.Quit`

**What goes wrong:** SIGINT/SIGTERM via the OS-level signal handler (cmd/sops-tui/main.go:56-60 per CONTEXT.md `<deferred>`) bypasses the tea event loop. Setting `m.quitting=true` does nothing because no further View() will be called — the signal handler calls `os.Exit(0)`.

**Per CONTEXT.md:** The signal path leaves cleanup to the renderer's deferred shutdown hook (which emits the alt-screen leave sequence even on signal path). Phase 11 does NOT modify main.go for this. If a user reports SIGINT-leaves-residue on a tested combo, it's a v1.1.x patch (CONTEXT.md `<deferred>` line 285).

**Plan 2 verification:** During the SC3 manual sweep, ctrl-c in each of the 4 Linux terminals; per-combo note: "SIGINT returns to shell with clipboard cleared; chrome residue: none / present (filed as #XXXX)".

## Test Patterns

### teatest is NOT used — Project Uses Update-Loop Pattern

`[VERIFIED: go.mod grep returns no teatest entry; charmbracelet/x/exp/golden is present (used by internal/testutil/golden.go) but exp/teatest is not]`

The "3 sanity teatests" in CONTEXT.md D-507 are misnamed for clarity — they're **integration tests** using the project's existing test harness. The pattern (from `internal/app/model_clipboard_test.go`):

```go
package app_test  // external test package — uses exported API + test seams

import (
    "testing"
    tea "charm.land/bubbletea/v2"
    "github.com/charmbracelet/colorprofile"
    "github.com/charmbracelet/x/ansi"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/caesarakalaeii/sops-tui/internal/app"
    "github.com/caesarakalaeii/sops-tui/internal/ui"
)

func TestRegression_RecipientFormMenuHints(t *testing.T) {
    m := app.NewAppModel(defaultEnv(), "", colorprofile.TrueColor)
    m = send(t, m, tea.WindowSizeMsg{Width: 120, Height: 40}).(app.AppModel)
    // ... drive into stateRecipientForm via existing key-press sequence ...
    m = send(t, m, tea.KeyPressMsg{Code: 'r'}).(app.AppModel) // hypothetical entry key
    // ... etc until m.State() == stateRecipientForm

    v := m.View()
    stripped := ansi.Strip(v.Content)

    // Form-level hints present
    assert.Contains(t, stripped, "Tab", "menu must include Tab hint in stateRecipientForm")
    assert.Contains(t, stripped, "Enter", "menu must include Enter hint")
    assert.Contains(t, stripped, "Esc", "menu must include Esc hint")

    // FileList hints absent — search by mnemonic to avoid false positives on body content
    assert.NotContains(t, stripped, "[j]", "stateRecipientForm menu must not include FileList [j] hint")
    assert.NotContains(t, stripped, "[k]", "stateRecipientForm menu must not include FileList [k] hint")
}
```

**Existing helpers Plan 2 uses:**

- `defaultEnv()` — `internal/app/model_test.go:18` — returns ui.EnvStatus with sops/age/sopsYaml=true
- `send(t, m, msg)` — `internal/app/model_test.go:28` — calls `m.Update(msg)` and returns the new model
- `app.ParsedFileForTest(nodes)` — exported test seam for injecting parsed YAML
- `app.ClipboardTimeout` — `model.go:1448` — `var ClipboardTimeout = clipboardTimeout` (test seam: tests override `app.ClipboardTimeout = func() time.Duration { return 10 * time.Millisecond }`)
- `m.IsClipboardHot()` — `model.go:1431` — exported accessor
- `ansi.Strip(s)` — `github.com/charmbracelet/x/ansi` — strips SGR / cursor sequences for structural assertion

### Test 1: TestRegression_ClipboardAutoClearWithChrome (D-507)

**Goal:** After copying a value, advancing the clipboard timeout, the indicator clears AND no `[W]`/`[E]` flash leaked.

**Mechanism:**

1. Override the clipboard timeout via the test seam: `t.Cleanup(func() { app.ClipboardTimeout = origTimeout })`; `app.ClipboardTimeout = func() time.Duration { return 10 * time.Millisecond }`.
2. Drive into stateDetail with a revealed leaf (use `setupDetailWithNodes` pattern from `model_clipboard_test.go:18-24`).
3. Send `tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl}` — copies value, sets `clipboardHot=true`, fires Flash + clipboard-clear tick.
4. Send `ui.FlashClearMsg{Gen: 1}` — clears flash. (Existing pattern from `model_clipboard_test.go:77`.)
5. Send `app.ClipboardClearMsg{Gen: 1}` — clears clipboardHot. (Pattern: `m.Update(app.ClipboardClearMsg{Gen: 1})`.)
6. Assert: `m.IsClipboardHot() == false`, `View().Content` does NOT contain `[clip]`, does NOT contain `[W]` / `[E]` prefix on the status bar (regex on stripped output for `^\[W\]` / `^\[E\]` at the start of any status-bar row).

**Why it catches a regression:** If the typed flash API (Phase 10 D-411) leaks `[W]`/`[E]` prefixes into a non-warn/non-err flash (e.g. a refactor that wires `FlashErr` through the success path), this test catches it. If the clipboard timeout doesn't reach the chrome's `[clip]` indicator, this test catches it.

### Test 2: TestRegression_RecipientFormMenuHints (D-507)

**Goal:** In stateRecipientForm, the menu shows form-level hints (Tab/Enter/Esc), NOT file-list hints (j/k/q).

**Mechanism:**

1. Construct AppModel; send WindowSizeMsg; drive into stateFileList with at least one file (use `ParsedFileForTest` + `FilesDiscoveredMsg`).
2. Drive into stateRecipientForm via the existing key sequence — likely `r` followed by `a` (add recipient) or similar. **Plan 2 author audits the actual key path** by reading `internal/keys/bindings.go` `RecipientFormKeyMap` and the model.go Update branch for the entry trigger.
3. Capture `m.View().Content`, ANSI-strip via `ansi.Strip`.
4. Assert presence of form-level hint mnemonics: `Tab`, `Enter`, `Esc`.
5. Assert ABSENCE of FileList hint mnemonics on menu rows: search for `[j]` and `[k]` in the menu rows specifically (the body may legitimately contain `j` or `k` characters in form labels).

**Mnemonic scoping caveat:** The menu renders cells as `[mnemonic] description`. The `[j]` substring in stripped output is unique to the menu (the description column isn't bracketed). So `assert.NotContains(stripped, "[j]")` is safe.

**Why it catches a regression:** Phase 9's menuHints dispatcher (`m.menuHints()` at model.go:1538+) dispatches by `(state, IsSearchActive)`. A future change to the dispatcher that misroutes stateRecipientForm to FileListHints would leak `[j]/[k]/[q]` mnemonics into the form's menu — this test catches it.

### Test 3: TestRegression_HealthOverlayOnNarrowWidth (D-507)

**Goal:** Health overlay reachable at narrow widths (80×24 + 60×24); the active `<health>` crumb survives D-425's first/last-preservation rule.

**Mechanism (per width):**

1. Construct AppModel; send `WindowSizeMsg{Width: 80, Height: 24}` (or `{60, 24}`).
2. Drive into stateHealth — likely via `Z` key (health scan keybinding) followed by HealthCheckResultMsg with empty result.
3. Capture `m.View().Content`, ANSI-strip.
4. Assert presence of either: `Weak secrets:`, `Duplicates:`, `(none found)`, or `Health` header (whichever the empty-result HealthModel renders).
5. Assert presence of `<health>` chip in the crumb row (active segment per D-425).

**Why it catches a regression:** Phase 10 D-425's `truncateSegmentsToWidth` first/last-preservation rule is the protection against narrow-width chip drop. A future regression that breaks the algorithm (e.g. an off-by-one in segment-width math) would drop `<health>` from the visible crumb row at 60×24 — this test catches it.

### ANSI-Strip Helper

`[VERIFIED: go.sum has charmbracelet/x/ansi]`

The project already uses `github.com/charmbracelet/x/ansi` in `internal/testutil/golden.go` for golden-file structural comparison. Plan 2's regression tests reuse it directly:

```go
import "github.com/charmbracelet/x/ansi"

stripped := ansi.Strip(m.View().Content)
```

`ansi.Strip` removes all CSI/OSC/SGR sequences but preserves printable text + box-drawing runes + newlines. Perfect for "does the menu contain X" assertions.

## Open Questions

> CONTEXT.md is exhaustive. Most discretionary points are flagged in `<decisions>` "Claude's Discretion". The questions below are the small remaining ones.

1. **`stateUnknown` sentinel constant — needed?**
   - What we know: CONTEXT.md "Integration Points" line 252 mentions a sentinel. Pattern 5 above shows the existing zero-state guard (`m.width == 0`) makes it unnecessary.
   - What's unclear: Whether Plan 1 author prefers explicit sentinel for clarity vs reliance on existing guard.
   - Recommendation: Skip the sentinel — fewer moving parts. The first non-zero-width View() runs after the WindowSizeMsg handler has called `refreshChromeCache`, so the cache is populated by the time View needs it.

2. **Single `chromeCache` string vs. split `chromeCache + chromeCrumbsCache` (D-501 discretion)?**
   - What we know: D-501 says "single concat or two separate cache fields — planner discretion, recommendation: two fields so that crumbs can be invalidated separately if a future phase needs it".
   - What's unclear: Nothing — discretion is explicit.
   - Recommendation: Two fields per CONTEXT.md recommendation. Tiny memory cost; cleaner semantics; future-proofs Phase 12+ in case crumbs gain their own invalidation seam.

3. **`m.quitting` flag in Plan 1 vs Plan 2 (D-518 discretion)?**
   - What we know: D-518 recommends Plan 2 to keep SC2-only Plan 1 split clean. CONTEXT.md notes the planner may move it to Plan 1 if the same Update branches are already touched.
   - What's unclear: Whether Plan 1's mutation-site audit hits model.go:993 (the Quit branch).
   - Recommendation: Plan 1 ALREADY touches the Quit branch because the Quit branch is one of the ~25 mutation sites that needs `refreshChromeCache` (state doesn't change, but the planner audits every branch that returns a new model). Folding `m.quitting = true` into the same edit is one extra line. **Recommend Plan 1** — but defer to planner if they want to keep SC2-only purity.

4. **Cache hit-rate test: 99/100 vs 100/100 expected hits?**
   - What we know: D-505 says "asserts `m.chromeCache` is reused at least 99/100 frames".
   - What's unclear: Why 99 not 100? With no Update between Views, the cache must hit 100% — first View populates cache via Update before the loop, then 100 reads with no mutation = 100 hits.
   - Recommendation: Pattern 6's test asserts cache key STABILITY (key never changes during the 100-frame loop) which is equivalent to 100/100 hits. The 99/100 wording in D-505 leaves slack for measurement noise — adopt the stronger 100/100 stability assertion (cleaner test). If the planner prefers the 99/100 wording for forward-compat slack, count rebuilds via a test-only counter and assert `count <= 1`.

## RESEARCH COMPLETE

**Phase:** 11 - Regression + Perf Gates
**Confidence:** HIGH (CONTEXT.md exhaustive; this research adds tactical specifics not re-derivation)

### Key Findings

- **Cache shape verified:** 4-field struct key per D-501 hashes at zero allocation; mutate-on-event helper follows Phase 8 D-213 infoPanel pattern verbatim.
- **`m.quitting` flag site count = 1:** Single Quit branch at model.go:993; no signal-handler wiring needed (CONTEXT.md `<deferred>` line 285 confirms signal path closes via renderer's deferred shutdown).
- **teatest framework NOT used by this project:** `charmbracelet/x/exp/teatest` is absent from go.mod; the 3 "sanity teatests" in D-507 use the existing Update-loop pattern (`internal/app/model_clipboard_test.go:18-90` is the canonical shape).
- **Cache hit-rate test mechanism:** Drive 100 sequential View() calls without intervening Update; assert `m.chromeCacheKey` stays equal across all 100 iterations (proves cache populated by Update, not by View — value-receiver discipline trip wire).
- **Critical pitfall: View() cannot mutate cache.** Value receiver means any in-View assignment is silently lost. Cache mutation MUST flow through `refreshChromeCache()` called from Update branches.

### File Created

`/home/moersener/git/sops-tui/.planning/phases/11-regression-perf-gates/11-RESEARCH.md`

### Confidence Assessment

| Area | Level | Reason |
|------|-------|--------|
| Cache wiring patterns | HIGH | Phase 8 D-213 infoPanel cache is a 1:1 precedent; verified in source |
| Bench measurement | HIGH | `testing.Benchmark` + `result.NsPerOp()` is stdlib; existing chrome_test.go:315 already calls it |
| `tea.View.AltScreen` semantics | HIGH | CLAUDE.md migration notes explicit; existing model.go:1366-1368 zero-state guard is the same pattern |
| Test patterns | HIGH | Codebase uses Update-loop tests; teatest absence confirmed via go.mod grep |
| 4-combo Linux sweep procedure | MEDIUM | D-509 checklist is per-combo manual; success rate depends on developer's terminal availability — verifier judgment call |

### Ready for Planning

Plan 1 (SC2 closure) and Plan 2 (SC1 + SC3 + SC4 + SC5 closure) can now be drafted. CONTEXT.md provides the 18 locked decisions; this research provides the tactical patterns and pitfall trip wires the planner uses to size tasks and verification.
