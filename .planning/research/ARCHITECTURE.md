# Architecture Research — v1.1 k9s Visual Shell

**Domain:** Bubble Tea v2 immediate-mode TUI adding a k9s-style visual chrome (header logo + info-panel + persistent-keybinding-menu + titled borders + crumb chips) on top of an existing, stable state machine.
**Researched:** 2026-04-23
**Confidence:** HIGH (all recommendations grounded in the existing repo + k9s reference + verified Bubble Tea v2 / lipgloss v2 idioms from project CLAUDE.md)

## Executive Summary

The v1.0 `AppModel` is a healthy Elm-architecture root with a clean one-enum state machine and sub-models that already own their own `View() string` rendering. v1.1 should **add a chrome layer**, not refactor the state machine.

Key architectural decisions:

1. **New `internal/ui/chrome.go` submodel** owns the header (info-panel + menu + logo) and exposes `View(width int) string` plus a small set of setters. AppModel.View() composes `[chrome][body][statusbar]`.
2. **Introduce a `Hinter` interface** in `internal/keys`: `Hints() []MenuHint`. Each stateful sub-model (FileList, Detail, Health, History, RecipientForm) implements it. AppModel asks the *active* sub-model for hints on every `View()`; the menu is re-rendered on every frame (immediate-mode — no observer pattern needed).
3. **`KeyMap → []MenuHint` helper** in `internal/keys/hints.go`: walks `key.Binding.Help()` and produces `MenuHint{Mnemonic, Description, Visible}` entries. Sub-models delegate to this helper.
4. **Titled borders applied by a `WrapTitled(title, content, width, height)` helper** in `internal/ui/chrome.go`. Each state renders its body at reduced dimensions, and AppModel.View() wraps the resulting string with a RoundedBorder + title. Sub-models do NOT wrap themselves — chrome is uniform and centrally enforced.
5. **Info-panel data lives on AppModel** (it already owns all the source data: `sopsYamlPath`, `gitRepoRoot`, `files`, `env`). A new `internal/ui/infopanel.go` submodel is a pure renderer that receives a struct; AppModel builds the struct once per `View()`.
6. **Logo color is derived from `EnvStatus` + last flash classification** — pure derivation at render time, no separate state field. This avoids cross-model invalidation plumbing.
7. **Crumb chips** are rendered by a new function `renderCrumbChips([]string) string` in `internal/ui/crumbs.go`. `StatusBarModel.breadcrumb` is the existing data source; add `Segments() []string` accessor. The existing `SetBreadcrumb` call-sites (~15 of them) require zero changes.
8. **`StatusBarModel` is augmented, not replaced**: it keeps env indicators + flash + clipboard indicator. The breadcrumb text moves from status-bar-left to a dedicated chip row just under the header. The status bar becomes shorter (right-aligned env + clipboard only).

## Integration with Existing Code

### Files Augmented (not replaced)

| File | Change | Rationale |
|------|--------|-----------|
| `internal/app/model.go` (View method at line 1329) | Rewrite body composition: chrome.View(width) + crumbs + titled-wrapped content + status.View(width). Subtract chrome height from mainH calculation. | AppModel owns composition per existing pattern. |
| `internal/app/model.go` (WindowSizeMsg handler line 313) | Subtract chromeHeight in addition to statusBarHeight when propagating size to children. | Children must render into the remaining body region. |
| `internal/ui/statusbar.go` | Add `Segments() []string` accessor. Remove `renderBreadcrumb()` from `View()` output (breadcrumb now rendered above). Center/left sections collapse; right section (env + clipboard) remains. | Data source unchanged; presentation moves to chrome layer. |
| `internal/ui/styles.go` | Add: `LogoStyleInfo`, `LogoStyleWarn`, `LogoStyleError`, `TitledBorder`, `TitleLabelStyle`, `MenuKeyStyle`, `MenuDescStyle`, `InfoPanelLabelStyle`, `InfoPanelValueStyle`, `CrumbChipStyle`, `CrumbChipActiveStyle`, `CrumbChipSep`. | Design-system additions — all explicit hex per the AdaptiveColor ban (issue #1036). |
| `internal/keys/bindings.go` | Add `MenuHint` struct (or reuse a shared one in a new `internal/keys/hints.go`). Add `ShortHelp()`-equivalent walker returning `[]MenuHint`. FileListKeyMap and DetailKeyMap already have `ShortHelp()` returning `[]key.Binding` — add a `Hints()` wrapper that converts. | Keybinding data already exists; just needs re-shape for the menu renderer. |
| `internal/ui/filelist.go` | Implement `Hints() []keys.MenuHint { return keys.HintsFromBindings(keys.DefaultFileListKeyMap.ShortHelp()) }`. | Tiny addition, no behavioral change. |
| `internal/ui/detail.go`, `health.go`, `history.go`, `recipientform.go`, `metadata.go`, `diff.go` | Same: implement `Hints()`. Overlays (diff, help) can return a minimal set like `[y/n/Esc]`. | Each view authors its own hint set. |
| `internal/ui/help.go` | Stays as the complete-reference overlay. Secondary role. No signature changes. Rename nothing — `?` still toggles. | v1.1 keeps `?` as "full reference" — k9s uses `?` the same way. |

### Files Added

| File | Purpose | Interface |
|------|---------|-----------|
| `internal/ui/chrome.go` | Owns the full header region. Compose: `[info-panel][spacer][menu][spacer][logo]`. Returns a multi-line string sized to `width`. Also exports `WrapTitled(title, body string, width, height int) string`. | `NewChromeModel(width int) ChromeModel` / `(c *ChromeModel) SetSize(w int)` / `(c ChromeModel) View(info InfoPanelData, hints []keys.MenuHint, logo LogoState, w int) string` / `(c ChromeModel) Height() int` |
| `internal/ui/infopanel.go` | Pure renderer for the top-left context panel: `.sops.yaml`, age fingerprint, recipient count, git state, file count. Mirrors k9s ClusterInfo. | `type InfoPanelData struct { SopsYamlPath, AgeFingerprint, GitBranch string; RecipientCount, FileCount int; GitAvailable, HasUncommittedChanges bool }` + `func RenderInfoPanel(d InfoPanelData, width int) string` |
| `internal/ui/logo.go` | 6-line ASCII sops-tui logo + 1-line status line underneath. Pure renderer. | `type LogoState int` (LogoInfo=0, LogoWarn, LogoError) + `type LogoStatus struct { Message string; State LogoState }` + `func RenderLogo(status LogoStatus, width int) string`. Exports `LogoSmall []string` (6 ASCII-art rows, width 26). |
| `internal/ui/menu.go` | Renders `[]keys.MenuHint` as a 2- or 3-column key-table. Models k9s `Menu` but immediate-mode. | `func RenderMenu(hints []keys.MenuHint, width int, maxRows int) string` (default `maxRows = 6`). |
| `internal/ui/crumbs.go` | Renders `[]string` as styled chips `< segment >`. The last segment uses `CrumbChipActiveStyle`. Replaces the status-bar's text breadcrumb. | `func RenderCrumbs(segments []string, width int) string` |
| `internal/keys/hints.go` | `MenuHint` struct + `HintsFromBindings(bindings []key.Binding) []keys.MenuHint` converter that reads `binding.Help()` from bubbles/key. Optional `Hinter` interface declaration. | `type MenuHint struct { Mnemonic, Description string; Visible bool }` + `type Hinter interface { Hints() []MenuHint }` + helper `HintsFromBindings([]key.Binding) []MenuHint` |

### Files NOT Changed

- `internal/app/model.go` state machine (enum + Update switch): zero changes needed.
- `internal/keys/bindings.go` existing `key.Binding` values: zero changes needed.
- `internal/ui/filelist.go`, `detail.go`, etc. content rendering: zero changes — they return the same body string.
- `internal/sops/*`, `internal/parser/*`, `internal/git/*`, `internal/health/*`: untouched.

## System Overview

```
┌────────────────────────────────────────────────────────────────────────────┐
│                      Chrome Layer (new — internal/ui/chrome.go)             │
│   ┌───────────────────────┐  ┌──────────────────────┐  ┌────────────────┐  │
│   │  InfoPanel (top-left) │  │  Menu (center, flex) │  │ Logo (26 cols) │  │
│   │  infopanel.go         │  │  menu.go             │  │ logo.go        │  │
│   └───────────────────────┘  └──────────────────────┘  └────────────────┘  │
│                                                                              │
│   ┌──────────────────────────────────────────────────────────────────────┐  │
│   │  Crumbs row (crumbs.go)   < sops-tui >  < files >  < prod.yaml >     │  │
│   └──────────────────────────────────────────────────────────────────────┘  │
├────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐  │
│  │  Titled Border (WrapTitled helper, new — in chrome.go)               │  │
│  │  ╭─ Files ────────────────────────────────────────────────────────╮   │  │
│  │  │                                                                 │   │  │
│  │  │    Current active state's body (unchanged — existing sub-model)│   │  │
│  │  │    fileList.View() | detail.View() | diff.View() | help.View() │   │  │
│  │  │                                                                 │   │  │
│  │  ╰────────────────────────────────────────────────────────────────╯   │  │
│  └──────────────────────────────────────────────────────────────────────┘  │
│                                                                              │
├────────────────────────────────────────────────────────────────────────────┤
│   Status bar (existing, shrunk) — env + clipboard indicator only            │
└────────────────────────────────────────────────────────────────────────────┘

Data flow (per frame, synchronous):

   AppModel.View() ──────────┬──> Chrome.View(infoData, hints, logoState, w)
                             │                  │
                             │                  └──> InfoPanel / Menu / Logo render
                             │
                             ├──> RenderCrumbs(m.status.Segments(), w)
                             │
                             ├──> activeSubModel.View()  ──> body string
                             │                  │
                             │                  └──> WrapTitled(title, body, w, h)
                             │
                             └──> m.status.View(w)  (status bar)

     All four are joined with lipgloss.JoinVertical(Left, ...).
```

## Component Responsibilities

| Component | Responsibility | File |
|-----------|---------------|------|
| `ChromeModel` | Compose and measure the header region (info + menu + logo); expose `Height()` so AppModel can size children correctly. | `internal/ui/chrome.go` (new) |
| `RenderInfoPanel` | Pure function: render context data as 5 label-value rows (`.sops.yaml:`, `age:`, `recipients:`, `git:`, `files:`). | `internal/ui/infopanel.go` (new) |
| `RenderMenu` | Pure function: render `[]MenuHint` as 2- or 3-column table (columns derived from width). Uppercase mnemonic styled accent, description styled fg. | `internal/ui/menu.go` (new) |
| `RenderLogo` | Pure function: emit the 6-line ASCII logo colored by `LogoState` plus a 1-line status message. | `internal/ui/logo.go` (new) |
| `RenderCrumbs` | Pure function: render each segment as `< segment >` chip, last one active color. | `internal/ui/crumbs.go` (new) |
| `WrapTitled` | Pure function: wrap any body string in a RoundedBorder with a title baked into the top border. | `internal/ui/chrome.go` (new) |
| `keys.MenuHint` / `keys.Hinter` / `keys.HintsFromBindings` | Data contract + converter between existing `key.Binding.Help()` and the new menu renderer. | `internal/keys/hints.go` (new) |
| Existing sub-models (FileList, Detail, …) | Implement `Hints() []keys.MenuHint` (one-line delegate to `HintsFromBindings(SomeKeyMap.ShortHelp())`). Render unchanged body. | existing files, 1-method addition |
| `AppModel.View()` | Build `InfoPanelData`, resolve `LogoState` from env+flash, look up active sub-model's `Hints()`, call Chrome/Crumbs/body/StatusBar, join vertically. | `internal/app/model.go` (rewrite one method) |

## Architectural Patterns

### Pattern 1: Chrome as an Immediate-Mode Frame (not a retained widget tree)

**What:** k9s uses tview's retained-mode Flex: primitives are added once, mutated on state changes, re-drawn by the framework. Bubble Tea v2 is immediate-mode: `View()` returns a fresh string every frame. The chrome is composed every render from inputs — no push/pop events, no listeners.

**When to use:** Always, in Bubble Tea. Do not try to port k9s's observer pattern (`StackPushed`, `StackPopped`, `StylesChanged`). The sessionState enum + re-render-on-every-msg is the Bubble Tea equivalent and it already works.

**Trade-offs:**
- Pro: No invalidation bugs, no stale widget state, trivial testability (golden-file the string).
- Con: Every frame re-computes hints, info-panel strings, and the logo. This is fine (Bubble Tea re-renders only on message; the cost is a few string allocations).

**Example (the full App.View()):**
```go
func (m AppModel) View() tea.View {
    // 1. Derive chrome inputs from model state (pure derivation, no side effects)
    infoData := m.buildInfoPanelData()
    hints := m.activeHints()             // asks the sub-model matching m.state
    logoState := m.resolveLogoState()    // info/warn/error from env + flash
    crumbSegs := m.status.Segments()
    title := m.titleForState()           // "Files" / "Keys — prod.yaml" / "Health" / ...

    // 2. Measure and size children
    statusBar := m.status.View(m.width)
    chrome := m.chrome.View(infoData, hints, logoState, m.width)
    crumbs := ui.RenderCrumbs(crumbSegs, m.width)

    chromeH := lipgloss.Height(chrome)
    crumbsH := lipgloss.Height(crumbs)
    statusH := lipgloss.Height(statusBar)
    bodyH := m.height - chromeH - crumbsH - statusH
    if bodyH < 1 { bodyH = 1 }

    // 3. Active body — existing sub-models, unchanged
    body := m.renderActiveBody(m.width-2, bodyH-2) // -2 accounts for border

    // 4. Wrap body in titled border, then stack
    wrapped := ui.WrapTitled(title, body, m.width, bodyH)

    full := lipgloss.JoinVertical(lipgloss.Left, chrome, crumbs, wrapped, statusBar)
    v := tea.NewView(full)
    v.AltScreen = true
    return v
}
```

### Pattern 2: Hinter Interface via Method Dispatch on sessionState

**What:** k9s uses a `Hinter` interface on each Component; a menu observer calls `top.Hints()` on stack push/pop. In Bubble Tea we don't have a live stack — we have an enum. Dispatch by enum value in AppModel, not by polymorphism.

**When to use:** Any time you need "the active view supplies data X to the chrome". This is the simplest, most testable approach in a value-type Elm model.

**Trade-offs:**
- Pro: One switch statement, fully explicit, no interface coercion across value types. Testable by state.
- Con: Adding a new state requires editing one switch — but that is true for every other state-routing case in this codebase, so it's consistent.

**Example:**
```go
// internal/app/model.go
func (m AppModel) activeHints() []keys.MenuHint {
    switch m.state {
    case stateFileList:
        return m.fileList.Hints()
    case stateDetail:
        return m.detail.Hints()
    case stateHealth:
        return m.health.Hints()
    case stateHistory:
        return m.history.Hints()
    case stateRecipientForm:
        return m.recipientForm.Hints()
    case stateDiff, stateRecipientConfirm, stateBulkReKeyConfirm:
        return keys.DiffHints // package-level []MenuHint — [y]confirm [n]cancel [Esc]cancel
    case stateHelp:
        return keys.HelpOverlayHints // [?]close [Esc]close
    case stateMetadata, stateFormatMenu, stateRecipientList:
        return keys.NavigationHints // minimal j/k/Esc set
    default:
        return nil
    }
}
```

### Pattern 3: Uniform Titled Border via Central Helper (not Per-View Wrapping)

**What:** Every view gets the same border look, same title placement, same padding. Bake this once in `WrapTitled`. Sub-models never render their own border.

**When to use:** Always, for v1.1. Any sub-model that renders its own border (none currently do — verified by reading styles.go and help.go) would produce inconsistent chrome.

**Trade-offs:**
- Pro: Consistency enforced by architecture, not by reviewer vigilance. One place to change border color, title placement, padding.
- Con: Sub-models must render bodies that fit exactly `width-2` × `height-2` — careful sizing is required. This is covered by Pitfall 2 below.

**Example:**
```go
// internal/ui/chrome.go
func WrapTitled(title, body string, width, height int) string {
    if width < 4  { width = 4 }
    if height < 3 { height = 3 }
    border := lipgloss.RoundedBorder()
    // title rendered inline with top border — lipgloss v2 supports this via
    // BorderTop(true) + custom TopTitle via string manipulation, or via a
    // manually-composed border. The simplest approach: render a Border, then
    // overwrite the top-left region of the first line with " Title ".
    style := lipgloss.NewStyle().
        Border(border).
        BorderForeground(ColorMuted).
        Padding(0, 1).
        Width(width - 2).
        Height(height - 2)
    rendered := style.Render(body)
    return overlayTitle(rendered, title) // injects " Title " at position (0,2)
}
```

*Note on the title injection:* lipgloss v2 does not yet expose a first-class "border title" API (confirmed in lipgloss v2.x docs). The practical approach is string-level overlay on the first line — this is what every production Bubble Tea app does today (huh, glow, soft-serve all use this pattern). A helper `overlayTitle(rendered, " Files (12) ")` replaces the appropriate slice of the first line.

### Pattern 4: Data-Derived Logo State (not Event-Driven)

**What:** Logo color (info / warn / error) is computed from `EnvStatus` + the current flash category at render time. No separate "logo state" field. No cross-model setter calls.

**When to use:** Any time a UI indicator is a pure function of model state that already exists.

**Trade-offs:**
- Pro: Impossible to get out of sync. No event wiring.
- Con: Flash messages are currently plain strings — to classify them we need a small addition to `StatusBarModel.Flash(msg, severity)`. Add an overload or a `FlashErr(msg)` / `FlashWarn(msg)` / `FlashInfo(msg)` triad. Roadmap impact: ~10 call-sites in `model.go` that flash errors get upgraded to `FlashErr`.

**Example:**
```go
func (m AppModel) resolveLogoState() ui.LogoStatus {
    // Error: any hard env failure
    if !m.status.Env().SopsAvailable {
        return ui.LogoStatus{Message: "sops not found", State: ui.LogoError}
    }
    // Error: last flash was classified as error
    if m.status.FlashSeverity() == ui.SeverityError {
        return ui.LogoStatus{Message: m.status.Flash(), State: ui.LogoError}
    }
    // Warn: soft env issues
    env := m.status.Env()
    if !env.AgeAvailable || !env.SopsYamlAvailable {
        return ui.LogoStatus{Message: "check keys/config", State: ui.LogoWarn}
    }
    return ui.LogoStatus{Message: "ready", State: ui.LogoInfo}
}
```

## Data Flow

### Per-frame render flow

```
tea.Msg delivered
      ↓
AppModel.Update(msg) ── mutates state (enum, flash, files, etc.) ──
      ↓
AppModel.View() called by Bubble Tea
      ↓
 ┌────────────────────────────────────────────────────────────┐
 │ buildInfoPanelData()   ← reads m.sopsYamlPath, m.files,    │
 │                           m.gitRepoRoot, m.currentParsed   │
 │ activeHints()          ← enum-switch to sub-model.Hints()  │
 │ resolveLogoState()     ← reads m.status.Env(),             │
 │                           m.status.FlashSeverity()          │
 │ titleForState()        ← enum-switch to static string      │
 └────────────────────────────────────────────────────────────┘
      ↓
 chrome.View(infoData, hints, logoState, width)  ─┐
 RenderCrumbs(m.status.Segments(), width)        ─┼──> lipgloss.JoinVertical
 WrapTitled(title, body, width, bodyH)           ─┤
 m.status.View(width)                            ─┘
      ↓
 tea.View{ Body: joined, AltScreen: true }
```

### Window size flow (the one subtle part)

```
tea.WindowSizeMsg{Width=W, Height=H}
      ↓
AppModel.Update:
  m.width = W; m.height = H
  chromeH := m.chrome.Height()          ← Chrome fixes its own height (7: 6-line logo)
  crumbsH := 1                           ← Crumbs always 1 line
  statusH := statusBarHeight(m)          ← existing helper, returns 1
  mainH := H - chromeH - crumbsH - statusH
  body inner height = mainH - 2         ← account for titled border top+bottom

  Propagate to ALL children:
    m.fileList.SetSize(W-2, mainH-2)
    m.detail.SetSize(W-2, mainH-2)
    m.help.SetSize(W-2, mainH-2)
    ... (same for all sub-models)
```

The single change to `WindowSizeMsg` is: introduce a `chromeH` constant or a `m.chrome.Height()` accessor, and subtract it. Every other SetSize call already exists.

## Recommended Build Order (Phase Sequencing)

Dependencies between new components form a DAG:

```
Phase A (foundation)
  └─ internal/keys/hints.go          (no deps)
  └─ internal/ui/styles.go additions (no deps)

Phase B (leaf renderers — all independent, can be built in parallel)
  ├─ internal/ui/logo.go             (needs styles)
  ├─ internal/ui/menu.go             (needs keys/hints.go + styles)
  ├─ internal/ui/infopanel.go        (needs styles)
  └─ internal/ui/crumbs.go           (needs styles)

Phase C (composition)
  └─ internal/ui/chrome.go           (uses logo + menu + infopanel; provides WrapTitled)

Phase D (integration)
  └─ internal/app/model.go View() rewrite    (uses chrome + crumbs + status)
  └─ internal/app/model.go WindowSize update (height accounting)
  └─ Hints() methods on each sub-model       (trivial, delegates to keys.HintsFromBindings)

Phase E (polish)
  └─ Flash severity classification in statusbar.go + call-site updates in model.go
  └─ Logo status binding in resolveLogoState
  └─ Theme/skin pass — palette tune
```

### Concrete roadmap phase suggestion (3 phases)

**Phase 6 — Chrome skeleton** (foundational: logo + menu + chrome frame, no info-panel yet)
- Build: `keys/hints.go`, `ui/logo.go`, `ui/menu.go`, `ui/chrome.go` (minus info-panel), `WrapTitled`.
- Integrate: AppModel.View() wraps body with titled border, renders logo + menu header. Info-panel area is a placeholder string.
- Add `Hints()` to FileList, Detail, Help, Diff.
- Gate: golden-file teatest snapshot of stateFileList with new chrome matches expected layout.

**Phase 7 — Info panel + crumb chips** (data binding + secondary chrome)
- Build: `ui/infopanel.go`, `ui/crumbs.go`. Replace status-bar breadcrumb with chip row.
- Integrate: AppModel builds `InfoPanelData` per-frame. Status bar shrinks (env + clipboard only).
- Add `Hints()` to Health, History, RecipientForm, Metadata.
- Gate: infopanel data reflects live env+git+file-count; crumb chips render with active segment highlight.

**Phase 8 — Logo state + theming**
- Build: flash severity (`FlashErr` / `FlashWarn` / `FlashInfo` on StatusBarModel). Update all ~10 flash call-sites with explicit severity. `resolveLogoState` derives logo color.
- Theme: pass through `styles.go` to tune palette to k9s defaults (swap accent color to match k9s dracula if desired).
- Gate: error flash turns logo red; warn flash turns it yellow; no flash keeps it info-colored. Screenshots match UI-SPEC 04 (to be authored in Phase 8).

This ordering means Phase 6 can ship a visible "it looks like k9s now" change even though info-panel is stubbed. Phase 7 makes it functionally equivalent to k9s. Phase 8 is the theming polish.

## Replaced vs Augmented

| Element | Verdict | Detail |
|---------|---------|--------|
| Root `AppModel` state enum | AUGMENTED | No new states added. |
| `AppModel.View()` method body | REWRITTEN | Compose chrome + crumbs + titled body + status bar. |
| `StatusBarModel` breadcrumb rendering | MOVED (not replaced) | Data stays on StatusBarModel. Rendering moves to `crumbs.go`. `renderBreadcrumb()` in statusbar.go is deleted. |
| `StatusBarModel` three-section layout | REPLACED | New status bar is right-aligned env + clipboard only. Left/center sections gone (breadcrumb → chips above; item-count → moves to title bar e.g. "Files (12)"). |
| Item count display | MOVED | From status-bar-center to title-bar (e.g. `WrapTitled("Files (12)", body, …)`). |
| `HelpModel` | KEPT AS-IS | Still the full-reference overlay behind `?`. Secondary role as advertised in PROJECT.md. |
| `keys/bindings.go` | AUGMENTED | Add `MenuHint`, `Hinter`, `HintsFromBindings`. No existing binding changed. |
| `styles.go` | AUGMENTED | Add ~10 new styles. No existing style changed. |
| Sub-model `View()` methods | UNCHANGED | They still return a body string. They do NOT add borders. |
| Sub-model sizes (`SetSize`) | ADJUSTED CALL-SITES | AppModel passes `mainH - 2` instead of `mainH` (room for titled border). One-line change in WindowSizeMsg handler. |

## Integration Pitfalls

### Pitfall 1: lipgloss width arithmetic — border padding eats columns

**What goes wrong:** `lipgloss.NewStyle().Border(Rounded).Width(80).Render(body)` renders a box **82 cells wide** — the border adds 2 columns beyond the set Width. Same for Height. If AppModel passes `m.width` to WrapTitled, the chrome overruns the terminal by 2, causing wrap/truncation bugs visible as "the right edge of the header cuts off".

**Prevention:** WrapTitled must be called with the **outer** width, and internally call `style.Width(width-2).Height(height-2)`. Document this in the helper's doc comment. Also: when sub-models render with `SetSize(W-2, H-2)`, they fit inside the inner region correctly.

**Detection:** Screen snapshot tests at width 80, 100, 120, 160 — visually verify the right border is present and the status bar still aligns.

### Pitfall 2: lipgloss.Height vs string line count — ANSI-aware counting required

**What goes wrong:** `len(strings.Split(s, "\n"))` lies when the string contains ANSI escapes with embedded newlines (rare but possible in border decorations). Use `lipgloss.Height(s)` — it's ANSI-aware.

**Prevention:** Always use `lipgloss.Height(s)` and `lipgloss.Width(s)` to measure rendered output. The existing `statusBarHeight(m)` helper in `model.go:1443` already follows this rule — reuse the pattern for `chromeHeight(m)`.

**Detection:** The code already does this correctly for the status bar. Just don't regress.

### Pitfall 3: Flicker from rebuilding the menu every frame

**What goes wrong:** The menu hint slice is rebuilt on every `View()` call. In principle this is free (`[]MenuHint{...}` literal, ~10 entries), but if `RenderMenu` re-runs `fmt.Sprintf` for each hint on every key press, slow terminals (tmux over SSH) may show flicker.

**Prevention:** The Cursed Renderer (Bubble Tea v2 default) synchronizes output per-frame using Mode 2026. This eliminates flicker at the terminal protocol layer. Do NOT try to memoize menu rendering inside the model — it adds cache-invalidation bugs for near-zero win. Trust the renderer.

**Detection:** Smoke test with `TERM=xterm-256color` and `TERM=screen-256color`; if flicker appears, check `view.AltScreen = true` is still set (it is, per model.go:1374).

### Pitfall 4: Title overflow — long titles push the border

**What goes wrong:** `WrapTitled("Keys — very-long-file-name-secrets-prod.yaml (37 keys)", body, 80, h)` produces a title longer than 80 cells, which breaks the top border.

**Prevention:** Truncate title to `width - 4` with an ellipsis. Mirror k9s's `ui.Truncate()` behavior (uses `runewidth.Truncate`). Add a `truncate(s string, max int) string` helper in chrome.go using `runewidth` (already transitively available via lipgloss).

**Detection:** Render test with narrow widths (40, 60) and long file names.

### Pitfall 5: Menu column count at narrow widths

**What goes wrong:** At `width < 60`, a 2-column menu with 10 hints may overflow. At `width > 120`, a 2-column layout wastes space.

**Prevention:** Compute `colCount := max(1, (width - infoPanelWidth - logoWidth) / minColWidth)` where `minColWidth = longestMnemonic + longestDescription + 4`. This mirrors k9s's column-fit logic. Fall back to 1 column if the header goes narrower than can fit all three sections horizontally — stack vertically in that case (info-panel on top, menu below, logo below).

**Detection:** Render test at widths 40, 60, 80, 120, 200.

### Pitfall 6: Sub-model height accounting forgotten in one place

**What goes wrong:** AppModel has ~15 call-sites where it does `mainH := m.height - statusBarHeight(m)` to size a diff/metadata/health overlay on-demand (e.g. model.go:924 inside the format menu handler). If chrome height is only subtracted in `WindowSizeMsg`, on-demand overlays will render oversized.

**Prevention:** Introduce a single helper `func (m AppModel) bodyDims() (w, h int)` that returns `(m.width-2, m.height - chromeH - crumbsH - statusH - 2)`. Replace **all 15 call-sites** with `w, h := m.bodyDims()` as part of the Phase 6 integration. Grep for `m.height - statusBarHeight` to find them:
- model.go:316, 349, 377, 484, 502, 567, 631, 724, 761, 846, 924, 1005, 1089, 1110, 1250.

**Detection:** After refactor, every one of those lines should be `w, h := m.bodyDims()`. A quick grep confirms no stragglers.

### Pitfall 7: Info-panel data freshness — async updates must re-render

**What goes wrong:** Info-panel shows "git: no repo" until `GitStatusMsg` arrives. Fine — that's the same latency as the existing status-bar indicator. But file-count and recipient-count can lag if AppModel doesn't re-render after `FilesDiscoveredMsg`. Bubble Tea v2 re-renders on every processed message by default, so this works — **unless** a handler returns `(m, nil)` early without calling `statusCmd` or equivalent. Verify: every handler that mutates `m.files` or `m.currentParsed` still lets Bubble Tea do a render (i.e. doesn't use `tea.Suspend` or similar).

**Prevention:** No code change — document the invariant in chrome.go: "InfoPanelData must be read fresh from AppModel fields on every View() call; never cache on Chrome." The current codebase does this naturally.

**Detection:** After new file discovery, file count in info-panel updates without user action.

### Pitfall 8: `help.Model` inside HelpModel + menu duplication

**What goes wrong:** The existing `HelpModel` uses `charm.land/bubbles/v2/help.Model` to render the `?` overlay. The new menu also renders keybindings. There's a risk of confusing the user: are those redundant?

**Prevention:** Keep the distinction:
- **Menu** (always visible, top-right): the 4-8 most useful keys for the current view. Short help only.
- **Help overlay** (`?`): the complete reference with descriptions, grouped by section.

Implementation: `ShortHelp() []key.Binding` feeds the menu. `FullHelp() [][]key.Binding` feeds the overlay. Both already exist on FileListKeyMap and DetailKeyMap. Zero new data plumbing. Reassure users by noting in the overlay footer: "Press ? to close — the top bar shows the most common keys."

**Detection:** On fileList view, the menu shows `↑/k move up`, `↓/j move down`, `l open`, `/ search`, `i info`, `space select`, `K re-key`, `H health`, `? help`, `q quit` — subset of FullHelp. The `?` overlay shows all groups.

### Pitfall 9: lipgloss v2 Border title — no native API

**What goes wrong:** Developers reach for `lipgloss.Border{Top: "─ Title ─"}` thinking lipgloss has a built-in border-title feature. It does not (as of lipgloss v2.x — verified via `github.com/charmbracelet/lipgloss` changelog). The widely-used workaround is string-level overlay: render the border, then find the first `─` run on line 0 and overwrite it with ` Title `.

**Prevention:** Implement `overlayTitle(rendered, title string) string` that:
1. Splits `rendered` into lines.
2. Replaces the run starting at column 2 on line 0 with `" " + title + " "`.
3. Preserves the left corner character and right-edge characters.

Well-known reference implementation: see `github.com/charmbracelet/soft-serve/pkg/ui/components/header`. Copy the pattern, don't invent.

**Detection:** Unit test `TestOverlayTitle_PreservesCornersAndWidth`.

### Pitfall 10: AltScreen flicker on first render before WindowSizeMsg

**What goes wrong:** Bubble Tea renders once before the first `WindowSizeMsg` arrives. At that point `m.width == 0` and `m.height == 0`, so `WrapTitled("Files", body, 0, 0)` would produce nonsense or panic.

**Prevention:** Early-return in `View()` when `m.width == 0`:
```go
if m.width == 0 || m.height == 0 {
    return tea.View{AltScreen: true}  // empty first frame
}
```
The existing codebase already relies on `statusBarHeight` handling zero correctly; extend this guard for the new chrome.

**Detection:** Run the TUI; verify no panic on startup.

## Anti-Patterns

### Anti-Pattern 1: Pushing chrome into each sub-model

**What people do:** Add a `RenderChrome()` method to each sub-model so it "owns its layout".
**Why it's wrong:** Duplicates composition 12 times (once per state). Any style change means 12 edits. The entire point of the immediate-mode root-composer model is central composition.
**Do this instead:** AppModel.View() composes; sub-models return only their body. This is what the current codebase already does for statusbar, and v1.1 extends that pattern.

### Anti-Pattern 2: Observer pattern for menu updates

**What people do:** Port k9s's `StackPushed`/`StackPopped` literally — register the menu as an observer on state changes.
**Why it's wrong:** Bubble Tea is immediate-mode. Observers leak memory, double-render, and fight the framework. There is no event bus; there is only `Update → View`.
**Do this instead:** Compute hints in `AppModel.View()` via enum dispatch (Pattern 2). The "update" is automatic because `View()` runs every frame.

### Anti-Pattern 3: Global singleton ChromeModel

**What people do:** Package-level `var Chrome = ui.NewChromeModel()` for "convenience".
**Why it's wrong:** Chrome holds width (and potentially more state later). Package globals break parallel tests and teatest snapshots.
**Do this instead:** Chrome is a value field on AppModel (`chrome ui.ChromeModel`), initialized in `NewAppModel`. Parallel with how `status`, `help`, `metadata` are fields today.

### Anti-Pattern 4: Caching rendered chrome strings

**What people do:** "The chrome changes rarely — cache the rendered string, invalidate on state change."
**Why it's wrong:** Adds an invalidation contract that must stay in sync with every input (env, flash, file count, hints, state, width). One missed invalidation = stale UI. Bubble Tea re-renders are already batched at the message level; caching saves microseconds and costs hours of debugging.
**Do this instead:** Trust the Cursed Renderer. Render fresh every frame.

### Anti-Pattern 5: Using lipgloss.AdaptiveColor for chrome palette

**What people do:** "Adaptive colors make the chrome work in light terminals too."
**Why it's wrong:** Confirmed hang — lipgloss v2 issue #1036, already documented as a hard ban in `internal/ui/styles.go` comment and in project CLAUDE.md.
**Do this instead:** Explicit `lipgloss.Color(ColorXxxHex)`. If light-terminal support is ever needed, ship a separate `styles_light.go` and select at startup by reading `$COLORFGBG` — but that is out of scope for v1.1.

## Integration Points

### External dependencies (unchanged)

| Component | Integration | Notes |
|-----------|------------|-------|
| `charm.land/bubbletea/v2` | Bubble Tea v2 `tea.View` return from `View()`, immediate-mode loop | Already wired; no new usage. |
| `charm.land/lipgloss/v2` | All styling + measurement (`lipgloss.Height`, `JoinVertical`, `RoundedBorder`) | No new dependencies. Pattern 1, 3 rely on these. |
| `charm.land/bubbles/v2/key` | `key.Binding.Help()` for mnemonic + description extraction | Source of truth for `HintsFromBindings`. |
| `github.com/mattn/go-runewidth` | Title truncation | Transitively available through lipgloss. |

### Internal boundaries

| Boundary | Communication | Notes |
|----------|--------------|-------|
| `AppModel` ↔ `ChromeModel` | AppModel builds value structs (`InfoPanelData`, `LogoStatus`, `[]MenuHint`) and passes to `chrome.View`. No back-channel. | Chrome is stateless except for `width`. |
| `AppModel` ↔ sub-models | Sub-models gain ONE method: `Hints() []keys.MenuHint`. Everything else unchanged. | Minimal interface expansion. |
| `StatusBarModel` ↔ `ChromeModel` | Zero direct interaction. AppModel bridges via `Segments()` and `FlashSeverity()` accessors. | Keeps status bar self-contained. |
| `keys` package ↔ `ui` package | `ui` imports `keys.MenuHint` + `keys.Hinter`. No circular dep (keys already doesn't import ui). | Clean layering preserved. |

## Scaling Considerations

Not really applicable — this is a local TUI. A few bounded concerns:

| Scale axis | Concern | Mitigation |
|-----------|---------|------------|
| File count in info-panel | 1000s of SOPS files in one repo | `RenderInfoPanel` displays `len(m.files)` — that's O(1). No rendering scaling issue. |
| Hint count per view | 15+ keybindings | Menu's `maxRows=6` handles up to 18 hints in 3 columns. If a view ever exceeds that, mark less-common ones `Visible: false` (mirrors k9s). |
| Terminal width | 40 cells (narrow SSH) to 300+ cells (ultrawide) | Pitfall 5: compute column count dynamically, fall back to vertical stack at widths below 60. |

## Sources

- k9s reference implementation (as of this repo's `~/git/k9s` checkout):
  - `internal/view/app.go:305-331` — buildHeader composition
  - `internal/ui/menu.go` — menu rendering + hint hydration
  - `internal/ui/logo.go` + `splash.go` — logo + status composition, color-state update
  - `internal/ui/crumbs.go` — breadcrumb chip rendering
  - `internal/ui/app.go` — root view wiring
  - `internal/model/menu_hint.go` — MenuHint + MenuHints data types
  - `internal/model/types.go:68-75` — Hinter interface definition
- sops-tui v1.0 codebase:
  - `internal/app/model.go:1329` — current View() composition (to be rewritten)
  - `internal/app/model.go:313-329` — WindowSizeMsg propagation pattern (to be updated)
  - `internal/ui/statusbar.go` — three-section layout (breadcrumb to be extracted)
  - `internal/ui/help.go` — existing `help.Model` usage (kept as-is for `?` overlay)
  - `internal/ui/styles.go` — color-palette bindings, AdaptiveColor ban
  - `internal/keys/bindings.go:72-85, 190-205` — existing `ShortHelp()` and `FullHelp()` methods that feed hints
- Project context:
  - `CLAUDE.md` — Bubble Tea v2 idioms (`View() returns tea.View`, `tea.KeyPressMsg`, AdaptiveColor ban issue #1036)
  - `.planning/PROJECT.md` — v1.1 milestone goals (header, menu, logo, info-panel, titled borders, crumb chips, theme)
- Bubble Tea v2 / lipgloss v2:
  - `charm.land/bubbletea/v2` v2.0.4 — Cursed Renderer, Mode 2026 synchronized output
  - `charm.land/lipgloss/v2` — `Height`, `Width`, `JoinVertical`, `RoundedBorder`; no native border-title API (workaround: string overlay, pattern 9)

---
*Architecture research for: sops-tui v1.1 k9s visual parity*
*Researched: 2026-04-23*
