# Phase 11: Regression + Perf Gates - Pattern Map

**Mapped:** 2026-05-04
**Files analyzed:** 7 (5 code/text + 1 binary screenshots dir + 1 doc)
**Analogs found:** 6 / 7 (terminal-bug.yml has no in-repo analog — external schema only)

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/app/model.go` (chromeCache + chromeKey + refreshChromeCache + Update wiring + View branches + m.quitting) | model + reducer | event-driven (cache-on-event) | `internal/app/model.go` `infoPanel` cache (Phase 8 D-213) — same file, same package, identical mutate-on-event shape | exact (in-file precedent) |
| `internal/app/regression_test.go` (NEW; 3 chrome-interaction sanity teatests) | test (integration) | request-response (Update-loop) | `internal/app/model_clipboard_test.go` (Update-loop pattern + ClipboardTimeout test seam + ANSI-strip View().Content assertions) | exact |
| `internal/app/chrome_test.go` (remove t.Skip line 311; add `TestChromeCache_HitRateAtSteadyState`) | test (gate + bench harness) | request-response | `internal/app/chrome_test.go:284-326` `TestBenchmarkAppView_UnderBudget` (existing test in same file; Plan 1 deletes line 311 only) | exact (in-file edit) |
| `internal/app/bench_test.go` (doc-comment updates only — no body change) | test (bench fixture) | request-response | `internal/app/bench_test.go:12-17` (existing comment block; Plan 1 extends) | exact (in-file edit) |
| `README.md` (add "Verified Terminals" H2 between Status and Stack — planner discretion on placement) | docs | content-presentation | Existing H2 sections at `README.md:7,16,20,25,31` (Features / Status / Stack / Requirements / License) — single H2 + body shape | role-match |
| `.github/ISSUE_TEMPLATE/terminal-bug.yml` (NEW; GitHub Forms YAML) | config (GitHub meta) | n/a | **NONE** — `.github/` directory does not exist; no other YAML config in this project follows GitHub Forms schema. External reference only: https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/syntax-for-issue-forms | no analog |
| `.planning/phases/11-regression-perf-gates/screenshots/{alacritty,ghostty,tmux-nested,vscode-integrated}.png` (4 binary captures) | binary fixture | file-I/O (manual capture) | **NONE** — no PNG fixtures elsewhere in repo; all goldens are `.golden` text files | no analog |

---

## Pattern Assignments

### `internal/app/model.go` chromeCache field + helpers (model, event-driven cache)

**Analog:** `internal/app/model.go` `infoPanel` cache from Phase 8 D-213 (same file — in-place precedent).

#### A1. Field declaration on AppModel struct (lines 267-271):

```go
// Phase 8 D-213: cached info-panel data. Refreshed at four event
// seams (NewAppModel + FilesDiscoveredMsg + FilesParsedMsg +
// GitStatusMsg). View() reads this cache only -- zero I/O at
// render time (Pitfall 15).
infoPanel ui.InfoPanelData
```

**Plan 1 copies:** Identical comment + field shape. Adds four new fields adjacent (after the `infoPanel` block):

```go
// Phase 11 D-501..D-503: cached chrome strings. Refreshed at every
// Update branch that mutates a chromeKey field (state, recipientAction,
// fileList.IsSearchActive(), width). View() reads this cache only -- zero
// renderer call at hit-path (mirrors Phase 8 D-213 infoPanel discipline).
chromeCache       string
chromeCrumbsCache string
chromeCacheKey    chromeKey
// Phase 11 D-512: alt-screen exit blank frame. Set true on Quit branch
// before returning tea.Quit; View() top branches on this and returns
// blank tea.View with AltScreen=true so the Cursed Renderer's last
// frame leaves no chrome residue in the user's shell.
quitting bool
```

#### A2. NewAppModel cache seed (lines 313-319):

```go
m.infoPanel = ui.InfoPanelData{
    SopsYamlRelPath: deriveSopsYamlRelPath(sopsYamlPath),
    AgeFingerprint:  loadAgeFingerprint(),
    RecipientCount:  -1,
    GitBranch:       "",
    FileCount:       -1,
}
```

**Plan 1 copies:** No explicit cache seed needed — zero-value `chromeKey{}` has `width=0`, which the existing zero-state guard (line 1365) catches before any cache read. Plan 1 leverages the existing guard rather than a sentinel constant (per RESEARCH Pattern 5 recommendation).

#### A3. Update branch mutation pattern (lines 384-386, FilesDiscoveredMsg cache refresh):

```go
m.status.SetItemCount(len(items), "items")
// Phase 8 D-213: refresh cached FileCount + repo-relative path.
m.infoPanel.FileCount = len(msg.Files)
m.infoPanel.SopsYamlRelPath = deriveSopsYamlRelPath(m.sopsYamlPath)
```

**Plan 1 copies the contract** (mutate-on-event discipline) and implements via the helper:

```go
// computeChromeKey returns the current cache key derived from AppModel
// fields that affect chrome rendering. Pure function of state — zero
// allocations (4-field struct with no slices/maps).
func (m AppModel) computeChromeKey() chromeKey {
    return chromeKey{
        state:           m.state,
        recipientAction: m.recipientAction,
        searchActive:    m.fileList.IsSearchActive(),
        width:           m.width,
    }
}

// refreshChromeCache rebuilds chromeCache + chromeCrumbsCache when the
// key has changed since last refresh. Called at the end of every Update
// branch that mutates a key field. Pattern matches Phase 8 D-213
// infoPanel refresh discipline (Pitfall 15: never-on-render).
func (m AppModel) refreshChromeCache() AppModel {
    newKey := m.computeChromeKey()
    if newKey == m.chromeCacheKey {
        return m
    }
    m.chromeCacheKey = newKey
    m.chromeCache = ui.RenderChrome(m.menuHints(), m.resolveLogoState(), m.infoPanel, m.palette, m.width)
    m.chromeCrumbsCache = ui.RenderCrumbs(m.status.Segments(), m.palette, m.width)
    return m
}
```

#### A4. WindowSizeMsg handler (lines 348-361 — first mutation site to audit per Pitfall D):

```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    w, h := bodyDims(m)
    // Propagate to all children that need dimensions
    m.fileList.SetSize(w, h)
    m.detail.SetSize(w, h)
    m.help.SetSize(w, h)
    m.metadata.SetSize(w, h)
    m.diff.SetSize(w, h)
    m.history.SetSize(w, h)
    m.health.SetSize(w, h)
    m.recipientForm.SetSize(w, h)
    return m, nil
```

**Plan 1 modifies:** Append `m = m.refreshChromeCache()` before `return m, nil`. Width changes are the most reliable cache invalidator.

#### A5. State-transition mutation sites (~20 in model.go — partial list verified by `grep -n "m\.state = state" model.go`):

```
Line 406: m.state = stateFileList   (FilesParsedMsg error path)
Line 423: m.state = stateDetail     (FilesParsedMsg success)
Line 523: m.state = stateDiff       (EditConfirmMsg)
Line 540: m.state = stateDiff       (EditorFinishedMsg)
Line 577: m.state = stateDetail     (RotateReadyMsg)
Line 603: m.state = stateDiff       (RotateReadyMsg follow-on)
Line 617: m.state = stateFormatMenu (RotateFormatMenuMsg)
Line 698: m.state = stateFileList   (HealthCheckResultMsg error)
Line 773: m.state = stateRecipientConfirm
Line 807: m.state = stateRecipientConfirm
Line 823: m.state = stateDetail
Line 856: m.state = stateFileList
Line 890: m.state = stateHealth
Line 970: m.state = stateDiff
Line 986: m.state = stateHelp
Line 1043: m.state = stateMetadata
Line 1124: m.state = stateHistory
Line 1143: m.state = stateRecipientForm
Line 1159: m.state = stateRecipientList
Line 1226: m.state = stateFileList
Line 1282: m.state = stateDiff
Line 2038: m.state = stateBulkReKeyConfirm
Line 2049: m.state = stateFileList
```

**Plan 1 copies the audit-and-edit pattern** from Phase 8 D-213 (which audited `m.infoPanel.*` mutation sites): each branch ends with `m = m.refreshChromeCache(); return m, cmd`. Place AFTER all sub-model assignments to avoid Pitfall E.

#### A6. recipientAction mutation sites (4 verified — all in model.go):

```
Line 796:  m.recipientAction = "remove"
Line 886:  m.recipientAction = ""           (cleared after healthcheck dispatch)
Line 1144: m.recipientAction = "add"
Line 1283: m.recipientAction = "healthcheck"
```

**Plan 1 audits all 4** — including the clearing site at line 886 (Pitfall F). Each branch ends with `refreshChromeCache`.

#### A7. Single-bool flip pattern for `m.quitting` — analog at lines 620-626 (`ClipboardClearMsg` clears `clipboardHot`):

```go
case ClipboardClearMsg:
    if msg.Gen == m.clipboardGen {
        clipboard.WriteAll("") //nolint:errcheck
        m.clipboardHot = false
        m.status.SetClipboardHot(false)
    }
    return m, nil
```

**Plan 1 mirrors:** Single boolean flip on AppModel (`m.clipboardHot = false` → `m.quitting = true`), within an Update branch, returned by value. The existing Quit branch at line 993 is the analog mutation site:

```go
// Global key: quit application
if key.Matches(msg, keys.DefaultGlobalKeyMap.Quit) {
    return m, func() tea.Msg { return tea.Quit() }
}
```

becomes:

```go
// Global key: quit application
if key.Matches(msg, keys.DefaultGlobalKeyMap.Quit) {
    m.quitting = true
    return m, func() tea.Msg { return tea.Quit() }
}
```

**Note:** `m.quitting` is the closest pattern to a single existing AppModel bool flag. There is no `m.searching` or `m.fileList.quitting` field; the closest precedents on AppModel are `clipboardHot` (line 251), `crossFilePopulated` (line 258), and `formatMenuActive` (line 243) — all simple bool flips inside Update branches with one read site. The `flashGen` field cited in the prompt is on `ui.StatusBarModel`, not `AppModel`, so it's a less direct precedent than `clipboardHot`.

---

### `internal/app/model.go` View() — zero-state guard + quitting branch + cache read

**Analog:** Existing zero-state guard at `internal/app/model.go:1364-1369` (in same file).

#### V1. Zero-state guard pattern (lines 1364-1369):

```go
func (m AppModel) View() tea.View {
    if m.width == 0 || m.height == 0 {
        v := tea.NewView("")
        v.AltScreen = true
        return v
    }
```

**Plan 1 copies the shape verbatim** for the quitting branch slotted ABOVE the existing zero-state guard:

```go
func (m AppModel) View() tea.View {
    // Phase 11 D-512: alt-screen exit blank frame. m.quitting is set in
    // the Quit handler before returning tea.Quit; this final View() returns
    // blank with AltScreen=true so the Cursed Renderer's last frame leaves
    // no chrome residue in the user's shell prompt area.
    if m.quitting {
        v := tea.NewView("")
        v.AltScreen = true
        return v
    }

    // Existing zero-state guard (Phase 7 Pitfall 5 first-frame safety).
    if m.width == 0 || m.height == 0 {
        v := tea.NewView("")
        v.AltScreen = true
        return v
    }
```

#### V2. Existing chrome composition (lines 1413-1422):

```go
// Phase 8 D-213 + D-216: chrome consumes m.infoPanel cache;
// crumbs row is unconditional (independent of chrome tier per
// D-216) -- the conditional guard from Phase 7 D-17 is removed.
// Phase 10 D-403: logo severity is resolved per-frame from
// (env, flashSeverity, lastHealthResult) via resolveLogoState().
chrome := ui.RenderChrome(hints, m.resolveLogoState(), m.infoPanel, m.palette, m.width)
crumbs := ui.RenderCrumbs(m.status.Segments(), m.palette, m.width)
statusBar := m.status.View(m.width)
sections := []string{chrome, crumbs, wrapped, statusBar}
full := lipgloss.JoinVertical(lipgloss.Left, sections...)
```

**Plan 1 modifies — replaces the renderer calls with cache reads:**

```go
// Phase 11 D-503: chrome + crumbs are cached on AppModel by
// refreshChromeCache() called at the end of every Update branch
// that mutates a chromeKey field. View() reads only — never calls
// RenderChrome/RenderCrumbs directly (eliminates the ~2.4-2.8 ms
// hot path measured in Phase 7.1 chrome_test.go:296-298).
chrome := m.chromeCache
crumbs := m.chromeCrumbsCache
statusBar := m.status.View(m.width)
sections := []string{chrome, crumbs, wrapped, statusBar}
full := lipgloss.JoinVertical(lipgloss.Left, sections...)
```

---

### `internal/app/regression_test.go` (NEW; 3 chrome-interaction sanity teatests)

**Primary analog:** `internal/app/model_clipboard_test.go` (Update-loop pattern, ClipboardTimeout test seam, ANSI-strip assertions). External test package `app_test`.

#### R1. Imports + helpers analog (`model_clipboard_test.go:1-32`):

```go
package app_test

import (
    "os"
    "testing"
    "time"

    tea "charm.land/bubbletea/v2"
    "github.com/charmbracelet/colorprofile"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/caesarakalaeii/sops-tui/internal/app"
    "github.com/caesarakalaeii/sops-tui/internal/ui"
)

// setupDetailWithNodes puts the AppModel into stateDetail with the given nodes.
func setupDetailWithNodes(t *testing.T, nodes []ui.TreeNode) tea.Model {
    t.Helper()
    m := app.NewAppModel(defaultEnv(), "", colorprofile.TrueColor)
    m2 := send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
    parsed := app.ParsedFileForTest(nodes)
    return send(t, m2, app.FilesParsedMsg{Parsed: parsed})
}

// asAppModel asserts the tea.Model is an AppModel and returns it.
func asAppModel(t *testing.T, m tea.Model) app.AppModel {
    t.Helper()
    am, ok := m.(app.AppModel)
    require.True(t, ok, "expected AppModel, got %T", m)
    return am
}
```

**Plan 2 copies:** Identical package + import block. The `defaultEnv()`, `send(t, m, msg)`, `setupDetailWithNodes(t, nodes)`, `asAppModel(t, m)`, `app.ParsedFileForTest(nodes)` helpers are reusable from `model_test.go:18-32` and `model_clipboard_test.go:18-32`. Plan 2 needs **NO new helpers** — every primitive already exists.

#### R2. Test 1 (TestRegression_ClipboardAutoClearWithChrome) — analog `model_clipboard_test.go:60-83`:

```go
// TestClipboardIndicatorVisibleAfterFlashClears verifies that [clip] appears in the
// status bar in normal (non-flash) mode after the flash clears.
func TestClipboardIndicatorVisibleAfterFlashClears(t *testing.T) {
    nodes := []ui.TreeNode{
        {Key: "token", Encrypted: true, TypeHint: "str", Revealed: true, DecryptedValue: "supersecret"},
    }
    m := setupDetailWithNodes(t, nodes)

    // Copy — this activates flash AND sets clipboardHot
    updated, cmd := m.Update(tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl})
    require.NotNil(t, cmd)

    // Drain all commands to find the FlashClearMsg gen
    am := asAppModel(t, updated)
    require.True(t, am.IsClipboardHot(), "clipboardHot must be true after copy")

    // Clear flash by sending FlashClearMsg (gen=1, matching the first flash)
    updated2, _ := updated.Update(ui.FlashClearMsg{Gen: 1})
    v := updated2.View()

    // Now in normal mode — [clip] should be visible
    assert.Contains(t, v.Content, "[clip]",
        "after flash clears, [clip] must appear in status bar while clipboardHot")
}
```

**Plan 2 extends this shape:** Drives `ctrl+y` → `ui.FlashClearMsg{Gen: 1}` → `app.ClipboardClearMsg{Gen: 1}` (per `model_clipboard_test.go:128-155` `TestClipboardClearMsgMatchingGen` analog at lines 138-149) and asserts the clipboard indicator clears. The chrome assertion (no `[W]`/`[E]` prefix) uses `ansi.Strip(v.Content)` per the project's standard ANSI-stripped structural assertion convention (see `internal/testutil/golden.go:30 RequireGoldenStructure`).

The `app.ClipboardClearMsg` test seam pattern (lines 144-149):

```go
// Send ClipboardClearMsg with gen=1 (matching)
m3, _ := m2.Update(app.ClipboardClearMsg{Gen: 1})

am3 := asAppModel(t, m3)
assert.False(t, am3.IsClipboardHot(),
    "ClipboardClearMsg with matching gen must clear clipboardHot")
```

**Plan 2 copies verbatim** for the regression test's clipboard-clear stage.

#### R3. Test 2 (TestRegression_RecipientFormMenuHints) — analog `model_test.go:603-615`:

```go
// TestRecipientFormStateTransition verifies that pressing a from stateDetail
// transitions to stateRecipientForm.
func TestRecipientFormStateTransition(t *testing.T) {
    m := modelInDetailWithRevealedLeaf(t)
    // Press a — should open add-recipient form
    m2, cmd := m.Update(tea.KeyPressMsg{Code: 'a'})
    v := m2.View()
    assert.NotEmpty(t, v.Content, "View must not be empty after a key in stateDetail")
    // Should render the recipient form overlay
    assert.Contains(t, v.Content, "Add Recipient",
        "stateRecipientForm must show 'Add Recipient', got: %q", v.Content)
    assert.NotNil(t, cmd, "stateRecipientForm activation must return a non-nil cmd (focus)")
}
```

**Plan 2 copies the entry path verbatim** (KeyPressMsg{Code: 'a'} from stateDetail with revealed leaf — see `modelInDetailWithRevealedLeaf` helper at `model_test.go:222-233`). Then ANSI-strips View().Content and asserts presence of `Tab/Enter/Esc` form-hints + absence of `[j]/[k]` FileList mnemonics. The hints come from `ui.RecipientFormModel.Hints()` at `internal/ui/recipientform.go:166-168` which calls `keys.HintsFromBindings(m.keys.ShortHelp())` against `DefaultRecipientFormKeyMap` (`internal/keys/bindings.go:624-633`) — confirming the menu emits `Enter / Esc` mnemonics in stateRecipientForm.

**Test 2 mnemonic-presence assertion analog** — `menuhints_drift_test.go:142-148` (in-package test):

```go
t.Run("stateRecipientForm", func(t *testing.T) {
    m := buildAppModel(t)
    m.state = stateRecipientForm
    require.Equal(t,
        expectedHintsWithSuppression(keys.DefaultRecipientFormKeyMap),
        m.menuHints())
})
```

**Plan 2 lives in `app_test` (external) package** — it can't directly poke `m.state = stateRecipientForm`. Instead it drives via `KeyPressMsg{Code: 'a'}` from a pre-revealed stateDetail (modelInDetailWithRevealedLeaf pattern at `model_test.go:222-233`).

#### R4. Test 3 (TestRegression_HealthOverlayOnNarrowWidth) — analogs `model_test.go:563-580` (health entry) + `resize_test.go:67-91` (narrow-width pattern):

Health entry path analog (`model_test.go:563-580`):

```go
// TestHealthCheckConfirmTransitionsToStateHealth verifies that confirming the health
// check gate transitions to stateHealth and dispatches an async scan command.
func TestHealthCheckConfirmTransitionsToStateHealth(t *testing.T) {
    m := app.NewAppModel(defaultEnv(), "", colorprofile.TrueColor)
    m2 := send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
    files := []sops.DiscoveredFile{
        {Name: "secrets/prod.yaml", AbsPath: "/repo/secrets/prod.yaml", IsEncrypted: true},
    }
    m3 := send(t, m2, app.FilesDiscoveredMsg{Files: files})

    // Press H to enter confirmation gate
    m4, _ := m3.Update(tea.KeyPressMsg{Code: 'H'})

    // Press y to confirm — should dispatch health scan and enter stateHealth
    m5, cmd := m4.Update(tea.KeyPressMsg{Code: 'y'})
    v := m5.View()
    assert.NotEmpty(t, v.Content, "View must not be empty after confirming health check")
    // cmd should be non-nil (dispatches async health scan)
    assert.NotNil(t, cmd, "confirming health check must return a non-nil cmd")
}
```

Narrow-width WindowSizeMsg pattern analog (`resize_test.go:78-91`):

```go
// TestResize_60x24 — mid-narrow tier (Phase 10 D-424). 60 lands in the
// mid-tier per Phase 7.1 D-116 (41 <= width < 99) so the chrome shows
// menu+logo without the info panel. [...]
func TestResize_60x24(t *testing.T) {
    setDeterministicAgeEnv(t)
    m := app.NewAppModel(defaultEnv(), "", colorprofile.TrueColor)
    updated, _ := m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
    m = updated.(app.AppModel)
```

**Plan 2 composes both analogs:** for each width in `{60×24, 80×24}`, send `WindowSizeMsg` → `FilesDiscoveredMsg` → `KeyPressMsg{Code: 'H'}` → `KeyPressMsg{Code: 'y'}` → `app.HealthCheckResultMsg{Result: empty}` → ANSI-strip View → assert overlay text (`Secret Health Check` per `internal/ui/health.go:141`, `All secrets passed health checks.` per line 155, or `No issues found` per line 154 for empty-result path) AND assert `<health>` chip present in crumb row (set by `m.status.SetBreadcrumb("files", "health")` at `model.go:891`).

**Note:** The breadcrumb segment is rendered as `<health>` (chip-pill) by `ui.RenderCrumbs` — assertion uses `assert.Contains(stripped, "health")` rather than the literal `<health>` because the rendered chip wraps the segment in chip-pill styling (verified via `internal/ui/crumbs.go:46 RenderCrumbs`).

---

### `internal/app/chrome_test.go` line 311 deletion + new `TestChromeCache_HitRateAtSteadyState`

**Analog:** Existing `TestBenchmarkAppView_UnderBudget` at `chrome_test.go:284-326` (same file, same package — Plan 1 deletes one line and adds one new test).

#### C1. Existing `t.Skip` at line 310-311 (Plan 1 DELETES line 311):

```go
func TestBenchmarkAppView_UnderBudget(t *testing.T) {
    t.Skip("deferred to Phase 11 SC2 — D-18 caching fallback; original 50 µs target preserved per Phase 7.1 governance restoration")
    if testing.Short() {
        t.Skip("bench-budget skipped in -short mode")
    }
    result := testing.Benchmark(BenchmarkAppView)
    nsPerOp := result.NsPerOp()
    const budgetNs = 50_000 // 50 µs — locked CONTEXT.md D-24 target; original wording per ROADMAP SC5.
    if nsPerOp > budgetNs {
        t.Fatalf("BenchmarkAppView regressed: got %d ns/op, budget is %d ns/op (50 µs at 200x60).\n"+
            "Check for lipgloss.NewStyle() allocations inside View() (see TestViewNoNewStyle)\n"+
            "and confirm RenderChrome is invoked at most once per View() call.",
            nsPerOp, budgetNs)
    }
    t.Logf("BenchmarkAppView: %d ns/op (budget %d ns/op, headroom %d ns/op)",
        nsPerOp, budgetNs, budgetNs-nsPerOp)
}
```

**Plan 1 deletes line 311 only** (the `t.Skip(...)` line). Lines 312-326 stay byte-identical. The `testing.Short()` skip stays as the only remaining skip path.

#### C2. New `TestChromeCache_HitRateAtSteadyState` — analog is the in-package construction pattern from `bench_test.go:18-32`:

```go
func BenchmarkAppView(b *testing.B) {
    env := ui.EnvStatus{
        SopsAvailable:     true,
        AgeAvailable:      true,
        SopsYamlAvailable: true,
        GitAvailable:      true,
    }
    m := NewAppModel(env, "", colorprofile.TrueColor)
    updated, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 60})
    m = updated.(AppModel)

    b.ReportAllocs()
    for b.Loop() {
        _ = m.View()
    }
}
```

**Plan 1 copies the env + NewAppModel + WindowSizeMsg + cast pattern** but loops 100× over `_ = m.View()` and asserts `m.chromeCacheKey` stays equal across all iterations (per RESEARCH Pattern 6). The test lives in `chrome_test.go` in the **internal `package app`** (NOT `app_test`) so it can read the unexported `m.chromeCacheKey` field directly — confirmed by `chrome_test.go:11` declaring `package app`.

**Test skeleton (per RESEARCH Pattern 6):**

```go
// TestChromeCache_HitRateAtSteadyState proves the chrome cache is wired
// (populated by Update, never by View). 100 sequential View() calls
// without any Update between them must leave m.chromeCacheKey unchanged
// — the value-receiver discipline trip wire (Pitfall A: View() cannot
// mutate state, so any cache mutation in View would silently lose).
func TestChromeCache_HitRateAtSteadyState(t *testing.T) {
    env := ui.EnvStatus{SopsAvailable: true, AgeAvailable: true, SopsYamlAvailable: true, GitAvailable: true}
    m := NewAppModel(env, "", colorprofile.TrueColor)
    updated, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 60})
    m = updated.(AppModel)

    keyAfterFirst := m.chromeCacheKey
    for i := 0; i < 100; i++ {
        _ = m.View()
        if m.chromeCacheKey != keyAfterFirst {
            t.Fatalf("cache key drifted at iteration %d: View() mutated cache "+
                "(expected View to be a pure read; cache must be populated by Update)", i)
        }
    }
}
```

---

### `internal/app/bench_test.go` doc-comment update (no body change)

**Analog:** Existing comment block at `bench_test.go:12-17`:

```go
// BenchmarkAppView records the v1.0-baseline render cost so Phase 7's
// chrome skeleton has a concrete "before" number to compare its
// <= 50 us/op target against. No gating on absolute value in Phase 6 (D-12).
//
// Uses testing.B.Loop() (Go 1.26+ idiom) — do NOT revert to `for i := 0; i < b.N; i++`.
// Fixed at 200x60 (D-12) — do NOT parametrise into sub-benchmarks.
```

**Plan 1 extends:** Adds a Phase 11 comment block above or below documenting the empirical 2.4-2.8 ms baseline (from `chrome_test.go:294-298`), the post-cache 50 µs target (50,000 ns budgetNs at `chrome_test.go:317`), and the cache hit rate dependency (TestChromeCache_HitRateAtSteadyState gate). No body change — `BenchmarkAppView` stays byte-identical at lines 18-33.

---

### `README.md` "Verified Terminals" H2 section (Plan 2)

**Analog:** Existing H2 sections in `README.md`:

```
Line 1:  # sops-tui                  (H1 title)
Line 7:  ## Features                 (H2 — first body section)
Line 16: ## Status
Line 20: ## Stack
Line 25: ## Requirements
Line 31: ## License
```

Existing H2 body shape (lines 7-15 — Features H2 with bulleted list):

```markdown
## Features

- **Browse** — Tree view of all SOPS-encrypted files in your repo
- **Inspect** — View encrypted keys at a glance without decrypting
- **Decrypt & View** — Reveal secret values on demand
- **Edit** — Modify values with automatic re-encryption
- **Rotate** — Generate new random values for secrets
- **k9s-style UX** — Keyboard-driven, vim-like navigation, command palette
```

**Plan 2 adds a new H2 between existing sections** — placement options:
- **Recommended (CONTEXT.md D-510 + RESEARCH.md SC3 row):** Between `## Installation` and `## Usage`. Note that the current README has neither — it has `## Features` / `## Status` / `## Stack` / `## Requirements` / `## License`. Plan 2 author picks the best fit; reasonable choices: between `## Status` (line 16) and `## Stack` (line 20), or between `## Stack` (line 20) and `## Requirements` (line 25).
- **Format (per CONTEXT.md "Claude's Discretion"):** H2 + markdown table with columns Terminal / Version Tested / OS / Status / Notes. 4 verified rows (alacritty, ghostty, tmux-nested, vscode-integrated) + 4 community-contributed rows (macOS Terminal, iTerm2, Windows Terminal, WSL2) with link to issue template.

**No code analog to copy — Plan 2 author writes a new markdown block following the existing H2-bulleted-list aesthetic.**

---

### `.github/ISSUE_TEMPLATE/terminal-bug.yml` (Plan 2)

**No analog in this project.**

- `.github/` directory does not exist — Plan 2 must create both `.github/` and `.github/ISSUE_TEMPLATE/` subdirectories.
- No existing GitHub Forms YAML in this repo.
- External reference (the only schema source): https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/syntax-for-issue-forms

**Required field shape per CONTEXT.md D-510:** terminal name, version, OS, screenshot, expected behaviour, observed behaviour, reproduction steps. Plan 2 author writes the YAML directly against the GitHub schema.

**Schema sketch (Plan 2 reference, not from in-repo analog):**

```yaml
name: Terminal Bug Report
description: Report a chrome rendering issue on a specific terminal emulator
labels: ["terminal-bug", "needs-triage"]
body:
  - type: input
    id: terminal-name
    attributes:
      label: Terminal name
      placeholder: e.g. iTerm2, Alacritty, Windows Terminal
    validations:
      required: true
  - type: input
    id: terminal-version
    attributes:
      label: Terminal version
    validations:
      required: true
  - type: dropdown
    id: os
    attributes:
      label: Operating system
      options:
        - Linux
        - macOS
        - Windows
        - Other
    validations:
      required: true
  - type: textarea
    id: expected
    attributes:
      label: Expected behaviour
    validations:
      required: true
  - type: textarea
    id: observed
    attributes:
      label: Observed behaviour
    validations:
      required: true
  - type: textarea
    id: repro
    attributes:
      label: Reproduction steps
    validations:
      required: true
  - type: textarea
    id: screenshot
    attributes:
      label: Screenshot
      description: Drag a PNG or paste a screenshot here
    validations:
      required: false
```

---

### `.planning/phases/11-regression-perf-gates/screenshots/*.png` (4 binary captures)

**No analog in this project.**

- No existing PNG fixtures in the repo (verified via `find . -name "*.png" -not -path "./.git/*"` returning empty).
- All test goldens are `.golden` text files under `internal/*/testdata/`.
- These are **manual capture artifacts**, not generated. Plan 2 author captures via terminal screenshot tooling at 200×60 in stateFileList per CONTEXT.md D-511.

---

## Shared Patterns

### Cache-on-Event Discipline (Pitfall 15 prescription)

**Source:** `internal/app/model.go:267-271 + 384-386 + 411-414 + 638-650 + 669` (Phase 8 D-213 infoPanel cache).

**Apply to:** `chromeCache` + `chromeCrumbsCache` per Plan 1.

**Rule:** Update mutates, View reads. Never the inverse. The trip wire is `TestChromeCache_HitRateAtSteadyState` — if a planner accidentally puts cache mutation inside `View()`, the value-receiver semantics silently lose the assignment (Pitfall A).

```go
// CORRECT — mutate inside Update, read inside View
case SomeMsg:
    m.someField = newVal
    m = m.refreshChromeCache()  // ← Phase 8 D-213 discipline
    return m, cmd
```

### Value-Receiver `m = m.helper()` Pattern

**Source:** `internal/app/model.go:1453 copyToClipboard` returns `(AppModel, tea.Cmd)`:

```go
func (m AppModel) copyToClipboard(value string) (AppModel, tea.Cmd) {
    if clipboard.Unsupported {
        var statusCmd tea.Cmd
        m.status, statusCmd = m.status.FlashWarn("...")
        return m, statusCmd
    }
    // ... mutations ...
    m.clipboardGen++
    m.clipboardHot = true
    // ... etc ...
    return m, tea.Batch(statusCmd, clearCmd)
}
```

**Apply to:** `refreshChromeCache()` per Plan 1 — value-receiver method returning the modified `AppModel`. Update branches assign `m = m.refreshChromeCache()`.

### Update-Loop Test Pattern (NOT teatest)

**Source:** `internal/app/model_test.go:18-32` + `model_clipboard_test.go:18-32` helpers. teatest framework intentionally NOT used (verified in RESEARCH §"Test Patterns" — `charmbracelet/x/exp/teatest` absent from go.mod).

```go
package app_test  // external test package — uses exported API + test seams

func defaultEnv() ui.EnvStatus { ... }
func send(t *testing.T, m tea.Model, msg tea.Msg) tea.Model { ... }
func setupDetailWithNodes(t *testing.T, nodes []ui.TreeNode) tea.Model { ... }
func asAppModel(t *testing.T, m tea.Model) app.AppModel { ... }
```

**Apply to:** All 3 regression sanity teatests in Plan 2's `regression_test.go`. The "teatest" naming in CONTEXT.md D-507 is a misnomer — these are integration tests using the project's existing Update-loop harness.

### ANSI-Strip Assertion Pattern

**Source:** `internal/testutil/golden.go:30 RequireGoldenStructure` + `profile_matrix_test.go:135-137` (in-package use):

```go
import "github.com/charmbracelet/x/ansi"

stripped := ansi.Strip(rendered)
assert.Contains(t, stripped, "cfg:", "info panel cfg label visible on Ascii")
```

**Apply to:** All 3 regression sanity teatests for menu-mnemonic / overlay-content / chip-text assertions per Plan 2.

### Single-Bool Flip on AppModel

**Source:** `internal/app/model.go:620-626 ClipboardClearMsg` handler — single bool field flip + return-by-value:

```go
case ClipboardClearMsg:
    if msg.Gen == m.clipboardGen {
        clipboard.WriteAll("") //nolint:errcheck
        m.clipboardHot = false
        m.status.SetClipboardHot(false)
    }
    return m, nil
```

**Apply to:** `m.quitting = true` in the Quit branch per Plan 1 (or Plan 2 — discretion). Single-bool, single-mutation-site, single-read-site (View top branch).

### chromeKey struct as Map-Key-Compatible Comparable

**Source:** No exact in-project precedent — chromeKey is novel.

**Verification of Go semantics (per RESEARCH Pattern 1):**
- 4-field struct `{state sessionState; recipientAction string; searchActive bool; width int}` is comparable via `==` because all fields are comparable types (no slices, maps, or func fields).
- Hash directly with zero allocation. Equality compare is constant-time over the 4 fields.
- `sessionState` is `int` (verified at `model.go:45 type sessionState int`); `string` is comparable; `bool` is comparable; `int` is comparable. ✓

The struct must be declared near the AppModel struct (or in a new `internal/app/chromecache.go` per planner discretion):

```go
// chromeKey is the cache invalidation key for chromeCache + chromeCrumbsCache.
// 4-field minimum per CONTEXT.md D-501 / D-502 — broader keys (palette,
// logoStatus, infoPanelData, flashGen) are explicitly rejected.
// Comparable via == (no slices/maps); usable as map key if a future phase
// needs per-state caches.
type chromeKey struct {
    state           sessionState
    recipientAction string
    searchActive    bool
    width           int
}
```

---

## No Analog Found

Files with no close match in the codebase (planner uses external schema or external file conventions):

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `.github/ISSUE_TEMPLATE/terminal-bug.yml` | config (GitHub meta) | n/a | `.github/` directory does not exist; no existing GitHub Forms YAML in this project. Plan 2 must create both directories. External schema: https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/syntax-for-issue-forms |
| `.planning/phases/11-regression-perf-gates/screenshots/{alacritty,ghostty,tmux-nested,vscode-integrated}.png` | binary fixture | file-I/O | No PNG fixtures in repo; all goldens are `.golden` text files. Manual capture per CONTEXT.md D-511. |

---

## Metadata

**Analog search scope:**
- `internal/app/` (model.go + 13 test files)
- `internal/ui/` (recipientform.go + health.go + crumbs.go + chrome.go)
- `internal/keys/` (bindings.go for RecipientFormKeyMap)
- `internal/testutil/` (golden.go for ANSI-strip helpers)
- `cmd/sops-tui/` (main.go for entry-point reference; no changes per CONTEXT.md)
- `README.md` (existing H2 structure)
- `.github/` (verified absent)

**Files scanned:** 18

**Pattern extraction date:** 2026-05-04

---

*Phase: 11-regression-perf-gates*
*Pattern map ready for planner consumption.*
