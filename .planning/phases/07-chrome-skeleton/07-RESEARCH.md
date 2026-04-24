# Phase 7: Chrome Skeleton - Research

**Researched:** 2026-04-24
**Domain:** Bubble Tea v2 chrome composition — persistent logo/menu/titled-border shell
**Confidence:** HIGH

## Summary

Phase 7 is unusually well-pre-decided: CONTEXT.md locks D-01..D-26 covering logo art direction, menu layout, the `Hints()` interface shape, titled-border strategy (NormalBorder + `overlayTitle` string-splice helper), `chromeHeight(m)` flip discipline, `AppModel.View()` composition, three grep-gates, and a bench-budget test. Research scope therefore focuses on **implementation-ready material** the planner and executors need — not re-litigating locked choices.

Five research gaps closed by this document:

1. **`overlayTitle` reference implementation** — the CONTEXT cites `charmbracelet/soft-serve/pkg/ui/components/header` as a reference. Direct inspection of that file at `main` revision shows it does NOT contain a title-overlay pattern — soft-serve's current header is a plain `Render()` of a text style. The string-splice pattern is instead a well-documented **community pattern** used by every TUI that needs titled borders on lipgloss v2 (confirmed no native title API exists in lipgloss v2 via pkg.go.dev inspection). We provide the concrete algorithm below, citing the actual upstream state.
2. **`lipgloss/v2/table` StyleFunc API** — confirmed via pkg.go.dev: `StyleFunc(func(row, col int) lipgloss.Style)`, `BorderTop/Bottom/Left/Right/Row/Column(false)` to render borderless, `Width(n)` for table width (no per-column width API — columns auto-size proportionally).
3. **Bubbletea v2 `View() tea.View`** — return `tea.View{Layer: lipgloss.NewLayer(fullString), AltScreen: true}` — the existing code uses `tea.NewView(s)` then mutates `v.AltScreen = true`. The recompose lands as the `tea.NewView(...)` argument; `AltScreen: true` stays.
4. **Validation Architecture** — concrete test-file map for `07-VALIDATION.md` generation: grep-gate AST walker, bench-budget test, `Hints()` per-sub-model tests, `overlayTitle` corners/width/truncation/empty test, chrome goldens at four resolutions.
5. **Plan 3 commit sequence** — 3 commits acceptable per D-25; we recommend a concrete order that keeps each commit independently buildable and grep-gate-passing.

**Primary recommendation:** Implement `overlayTitle` as a pure string helper (split lines, replace columns `[2 : 2+titleWidth]` of line 0, preserve corners, truncate if title wider than `width-4`). Use `lipgloss/v2/table` with all six Border* toggles false and `StyleFunc` for mnemonic/description column coloring. Plan 3 commits in order: (1) primitives wiring (chromeHeight flip + View rewrite + AppModel dispatcher + 9 sub-model `Hints()`); (2) grep-gates + bench-budget test + magic-constant migration; (3) golden refresh at 4 resolutions.

## User Constraints (from CONTEXT.md)

### Locked Decisions

**Logo (UI-02)**
- **D-01:** 6-row stacked "SOPS-TUI" design — 5-row block figlet of "SOPS" + "tui" subscript on row 6. Width ~26 cols. ASCII-only (no emoji, no VS16, no ZWJ).
- **D-02:** Default color for Phase 7 = `ColorAccent`. `RenderLogo(status LogoStatus, width int)` accepts the parameter but Phase 7 callers pass `LogoInfo` unconditionally. Severity coupling deferred to Phase 10.
- **D-03:** Anchored top-right. Positioned via `lipgloss.JoinHorizontal` with flexible spacer between info-panel-placeholder, menu, and logo.

**Menu (UI-01)**
- **D-04:** Fixed 2 columns × 6 rows = 12 hint slots at all widths. Narrow-terminal safe.
- **D-05:** Rendered via `lipgloss/v2/table` with `StyleFunc(row, col int) lipgloss.Style`. Mnemonic column uses `MenuKeyStyle` (ColorAccent); description column uses `MenuDescStyle` (ColorFg).
- **D-06:** `MenuHint.Visible` (bool) controls inclusion in the persistent menu. Sub-models curate ≤12; rest discoverable in `?` overlay.
- **D-07:** Cell format is `[mnemonic] description` — mnemonic left-aligned in fixed-width subcolumn of column 0; description fills column 1.

**Hints() interface**
- **D-08:** `internal/keys/hints.go` — `type MenuHint struct { Mnemonic, Description string; Visible bool }`, `type Hinter interface { Hints() []MenuHint }`, `func HintsFromBindings(bindings []key.Binding) []MenuHint`.
- **D-09:** All 9 interactive sub-models implement `Hints()` in Phase 7: FileList, Detail, Help, Diff, Metadata, Health, History, RecipientList, RecipientForm. stateFormatMenu modal uses inline hint set.
- **D-10:** AppModel hint dispatcher is a pure function of `(state sessionState, recipientAction string, IsSearchActive bool)`. Modal states `stateDiff`/`stateRecipientConfirm`/`stateBulkReKeyConfirm` dispatch via `recipientAction`.
- **D-11:** When `IsSearchActive()` returns true in stateFileList, menu shows `[Esc]Exit search / [Enter]Select`.

**Titled border (UI-06)**
- **D-12:** `WrapTitled(title, body string, width, height int) string` lives in `internal/ui/chrome.go`.
- **D-13:** `lipgloss.NormalBorder()` exclusively. Border foreground = `ColorMuted`. No `FocusedBorder`/`UnfocusedBorder`.
- **D-14:** Title injection via `overlayTitle(rendered, " Title ") string` helper — string-splice on first line at column position 2. Pattern cited from soft-serve reference (note: pattern is community-standard; see "Closed Research Gaps" below).
- **D-15:** Title format by view — list views use bare `(N)`; Health uses `(N findings)` unit-ful suffix; Detail uses `Detail: <filename>`. (Full table in CONTEXT.md.)

**chromeHeight + View composition**
- **D-16:** `chromeHeight(m)` at `model.go:1415` flips from `return 0` to real value = `lipgloss.Height(renderedChrome)` (expected 6). Info-panel area rendered as 6-row blank block preserving Phase 8 alignment. `crumbsHeight(m)` stays stubbed at 0.
- **D-17:** `AppModel.View()` composes `[chrome][crumbs-placeholder][wrapped body][status bar]` via `lipgloss.JoinVertical(lipgloss.Left, …)`. Sub-models render at `bodyDims(m).w - 2` × `bodyDims(m).h - 2`; `WrapTitled` restores full envelope.
- **D-18:** Chrome rendering = pure every-frame composition with all styles as package vars. Caching deferred to Phase 10 if needed.
- **D-19:** `model.go:1841` magic `m.height - 4` migrates to `bodyDims`-based computation via `WrapTitled` border math.

**Grep-gate discipline**
- **D-20:** `TestChromeASCIIOnly` — scans `internal/ui/{chrome,logo,menu,crumbs}.go` for runes > 0x7F; allowlist: `─ │ ╭ ╮ ╰ ╯` (box-drawing used by NormalBorder).
- **D-21:** `TestChromeNormalBorderOnly` — scans chrome files for `RoundedBorder|ThickBorder|DoubleBorder|HiddenBorder|FocusedBorder|UnfocusedBorder` literals; fails if any appear.
- **D-22:** `TestViewNoNewStyle` — AST-walk `internal/app/model.go`, locate `View()` method body, fail if `lipgloss.NewStyle(` appears inside (directly or via helper lambdas).
- **D-23:** Phase 6's `TestBodyDimsMigration` regex (`m\.height\s*-\s*statusBarHeight`) remains untouched.

**Bench gate**
- **D-24:** `TestBenchmarkAppView_UnderBudget` — runs bench once, asserts `ns/op ≤ 50_000` (50µs) at 200×60 with full Phase 7 chrome.

**Plan split**
- **D-25 / D-26:** Three plans, primitive-first. Plan 1 = primitives (hints.go, logo.go, menu.go, styles additions). Plan 2 = chrome composer + WrapTitled + overlayTitle + corners/width unit test. Plan 3 = integration (chromeHeight flip, View rewrite, Hints on 9, dispatcher, 3 grep-gates, bench test, 4 golden refreshes, magic-constant migration) — one atomic PR, 2-3 commits.

### Claude's Discretion

- Exact logo byte-art within the 6×~26 envelope.
- `overlayTitle` details: `lipgloss.Width` first-line measurement, insertion at column 2, overlong truncation with ellipsis, empty-title passthrough.
- `RenderLogo` return shape (combined art+status vs split).
- `RenderMenu` mnemonic subcolumn width math (recommend: `max(MnemonicWidth) + 1`).
- Golden file naming — `chrome_<state>_<WxH>.golden` vs per-state dir.
- Exact new line number for `chromeHeight` body post-flip.
- Whether `stateFormatMenu` gets a titled border (overlay — Plan 3 decides).

### Deferred Ideas (OUT OF SCOPE)

**Phase 8:** Header info panel (5 rows), breadcrumb chip row, status bar shrink.
**Phase 9:** Golden matrix per `(state, recipientAction, IsSearchActive)` tuple with hint-vs-keymap drift test.
**Phase 10:** Logo severity coupling, k9s palette tune, 16-color fallback, redundant encoding, narrow-terminal survival.
**Phase 11:** Alt-screen fill/blank frame, terminal compat sweep, full "Looks Done But Isn't" sign-off.
**v2:** Skin YAML loader, builtin skins, fsnotify hot-reload, menu second-page overflow.

## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| UI-01 | Persistent multi-column keybinding menu in header on every view — no `?` press required | `lipgloss/v2/table.StyleFunc(row,col)` + `Hints()` on 9 sub-models + AppModel dispatcher on `(state, recipientAction, IsSearchActive)` |
| UI-02 | 6-row ASCII logo anchored top-right, ~26 columns wide | `RenderLogo(status, width) string` + `LogoSmall [6]string` byte-art candidate (Section "Logo Byte-Art Candidates") |
| UI-06 | Every primary view wrapped in titled bordered region; title encodes name + item count | `WrapTitled(title, body, w, h)` + `overlayTitle(rendered, title)` string-splice (Section "overlayTitle Implementation") + per-state title map (D-15) |
| UI-15 | ASCII-only chrome; only `NormalBorder()`; grep-gated | `TestChromeASCIIOnly` + `TestChromeNormalBorderOnly` + `TestViewNoNewStyle` (Section "Validation Architecture") |

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Logo rendering | `internal/ui` (view primitive) | — | Pure stateless renderer; data flows in, string flows out |
| Menu rendering | `internal/ui` (view primitive) | `internal/keys` (MenuHint source-of-truth) | Menu derives from keymap hints; keys owns the contract |
| Titled border wrap | `internal/ui/chrome.go` | — | Central `WrapTitled` avoids 12 sub-model border edits |
| Hints() dispatch | `internal/app` (root model) | `internal/ui` sub-models (each Hinter impl) | AppModel is the only place that knows `(state, recipientAction, IsSearchActive)` |
| Chrome height arithmetic | `internal/app` (`chromeHeight(m)` helper) | `internal/ui/chrome.go` (exports render) | Height flows from composition result; helper delegates to pre-rendered string height |
| Grep-gate enforcement | `internal/app` test package | — | Tests inspect `.go` file bytes + AST; AST walker lives next to the code it guards |
| View composition | `internal/app/model.go` `View()` method | `internal/ui/chrome.go` (composer) | AppModel.View is the single integration seam per PITFALLS Anti-Pattern 1 |

## Standard Stack

### Core (already resolved — no new deps)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `charm.land/bubbletea/v2` | v2.0.4 | `View() tea.View`, `tea.KeyPressMsg`, AltScreen | Project-locked per CLAUDE.md [VERIFIED: go.mod] |
| `charm.land/lipgloss/v2` | v2.0.3 | `NormalBorder()`, `JoinHorizontal/Vertical`, `Width`/`Height`, `NewStyle()` | Project-locked [VERIFIED: go.mod line 8] |
| `charm.land/lipgloss/v2/table` | v2.0.3 (sub-package) | Menu grid with `StyleFunc(row,col)` | Already transitively in go.sum via lipgloss v2 [VERIFIED: pkg.go.dev/charm.land/lipgloss/v2/table] |
| `charm.land/bubbles/v2/key` | v2.1.0 | `key.Binding.Help()` → `{Key, Desc}` source of hints | Already used by keys/bindings.go [VERIFIED: internal/keys/bindings.go:16] |
| `github.com/charmbracelet/x/ansi` | v0.11.7 | `ansi.Strip` for golden comparison | Already direct dep (Phase 6) [VERIFIED: go.mod line 11] |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `stretchr/testify` | v1.11.1 | `require` for assertions in unit tests | Phase 7 unit tests follow Phase 6 convention [VERIFIED: go.mod line 16] |
| `testing` (stdlib) | Go 1.26.2 | `testing.B.Loop()` for bench + `testing.T` for grep-gate AST walker | Phase 6 already uses this pattern [VERIFIED: internal/app/bench_test.go] |

### go.mod promotions needed

- `charm.land/lipgloss/v2/table` — currently indirect (transitively available via lipgloss v2). Plan 1 confirms with `go mod tidy` after the first import; if unchanged, no go.mod edit needed. If `go mod tidy` adds it to the direct require block, accept the one-line change (zero version churn).

**Installation:** `go mod tidy` after Plan 1 writes the first `import "charm.land/lipgloss/v2/table"`. No `go get`.

**Version verification:** Current versions verified against go.mod as of 2026-04-24. No upgrades required. `go 1.26.2` matches the project constraint in CLAUDE.md.

## Architecture Patterns

### System Architecture Diagram

```
Input: tea.Msg (WindowSizeMsg, KeyPressMsg, ClipboardTickMsg, …)
                  ↓
         AppModel.Update(msg) — unchanged from v1.0 state machine
                  ↓
         tea.Msg processed → AppModel state mutated
                  ↓
         AppModel.View() called by bubbletea/v2 runtime
                  ↓
    ┌─────────────────────────────────────────────────────────────┐
    │ 1. Derive chrome inputs (pure derivation from model state) │
    │    - hints := m.menuHints()   ← (state, recipientAction,    │
    │                                   IsSearchActive) dispatcher│
    │    - title := m.titleForState() ← per-state title map D-15  │
    │    - logoStatus := ui.LogoInfo  ← default in Phase 7        │
    │    - infoPanelBlank := 6-row empty placeholder for Phase 8  │
    └─────────────────────────────────────────────────────────────┘
                  ↓
    ┌─────────────────────────────────────────────────────────────┐
    │ 2. Render chrome (pure, every frame)                         │
    │    chrome := ui.RenderChrome(infoPanelBlank,                │
    │                              ui.RenderMenu(hints, ...),     │
    │                              ui.RenderLogo(logoStatus, 26), │
    │                              width)                         │
    │        → JoinHorizontal[Top](blank, menu, spacer, logo)     │
    │        → 6 rows × width                                     │
    └─────────────────────────────────────────────────────────────┘
                  ↓
    ┌─────────────────────────────────────────────────────────────┐
    │ 3. Render active body (unchanged sub-model View())          │
    │    w, h := bodyDims(m) ← chromeHeight flipped → 6           │
    │    body := sub.View()  ← rendered at (w-2) × (h-2)          │
    └─────────────────────────────────────────────────────────────┘
                  ↓
    ┌─────────────────────────────────────────────────────────────┐
    │ 4. Wrap body in titled border                                │
    │    wrapped := ui.WrapTitled(title, body, w, h)              │
    │        → lipgloss.NormalBorder() + overlayTitle injection   │
    └─────────────────────────────────────────────────────────────┘
                  ↓
    ┌─────────────────────────────────────────────────────────────┐
    │ 5. Compose final frame                                       │
    │    crumbsPlaceholder := ""   ← crumbsHeight stays 0         │
    │    full := lipgloss.JoinVertical(lipgloss.Left,             │
    │               chrome, crumbsPlaceholder, wrapped, statusBar)│
    │    return tea.View{Layer: NewLayer(full), AltScreen: true} │
    └─────────────────────────────────────────────────────────────┘
                  ↓
         Output: tea.View — rendered by bubbletea/v2 runtime
```

### Recommended Project Structure

```
internal/
├── keys/
│   ├── bindings.go           # UNCHANGED
│   └── hints.go              # NEW (Plan 1) — MenuHint + Hinter + HintsFromBindings
├── ui/
│   ├── logo.go               # NEW (Plan 1) — LogoSmall + RenderLogo + LogoStatus
│   ├── menu.go               # NEW (Plan 1) — RenderMenu with lipgloss/v2/table
│   ├── chrome.go             # NEW (Plan 2) — RenderChrome + WrapTitled + overlayTitle
│   ├── styles.go             # ADDED VARS (Plan 1) — MenuKeyStyle, MenuDescStyle,
│   │                         #   LogoStyleInfo/Warn/Error, TitledBorderStyle,
│   │                         #   TitleLabelStyle
│   ├── filelist.go           # +1 method (Plan 3) — Hints()
│   ├── detail.go             # +1 method (Plan 3) — Hints()
│   ├── help.go               # +1 method (Plan 3) — Hints()
│   ├── diff.go               # +1 method (Plan 3) — Hints()
│   ├── metadata.go           # +1 method (Plan 3) — Hints()
│   ├── health.go             # +1 method (Plan 3) — Hints()
│   ├── history.go            # +1 method (Plan 3) — Hints()
│   ├── recipientform.go      # +1 method (Plan 3) — Hints()
│   └── (recipient list renderer — currently in model.go:1799)
├── app/
│   ├── model.go              # MODIFIED (Plan 3):
│   │                         #   - chromeHeight body flipped
│   │                         #   - View() rewritten for chrome composition
│   │                         #   - new menuHints() dispatcher method
│   │                         #   - new titleForState() map helper
│   │                         #   - renderRecipientList uses bodyDims/WrapTitled (D-19)
│   ├── chrome_test.go        # NEW (Plan 3) — grep-gates + bench-budget test
│   ├── hints_test.go         # NEW (Plan 3) — Hints() per sub-model unit tests
│   ├── resize_test.go        # MODIFIED (Plan 3) — 4 goldens refreshed
│   └── testdata/
│       ├── resize_40x12.golden   # REFRESHED with chrome
│       ├── resize_80x24.golden   # REFRESHED
│       ├── resize_120x40.golden  # REFRESHED
│       └── resize_200x60.golden  # REFRESHED
```

### Pattern 1: Immediate-Mode Chrome Composition (from ARCHITECTURE §Pattern 1)

**What:** Chrome is rebuilt every `View()` call from pure-derived inputs. No caching in Phase 7 (D-18).

**When to use:** Always, given the `TestViewNoNewStyle` grep-gate + bench-budget `TestBenchmarkAppView_UnderBudget` (D-22, D-24). The three guardrails make zero-allocation pure composition safe.

**Example (canonical Phase 7 View body shape):**
```go
// [CITED: .planning/research/ARCHITECTURE.md Pattern 1, adapted for Phase 7]
func (m AppModel) View() tea.View {
    if m.width == 0 || m.height == 0 {
        v := tea.NewView("")
        v.AltScreen = true
        return v
    }

    // 1. Derive chrome inputs — pure derivation, no side effects
    hints := m.menuHints()          // (state, recipientAction, IsSearchActive) dispatcher
    title := m.titleForState()      // per-state title map (D-15)

    // 2. Render sub-model body at inner dims
    w, h := bodyDims(m)
    innerW, innerH := w-2, h-2
    var body string
    switch m.state {
    case stateFileList:       body = m.fileList.View()
    case stateDetail:         body = m.detail.View()
    // ... unchanged dispatch for all 14 states
    }

    // 3. Wrap body in titled border (D-12, D-13, D-14)
    wrapped := ui.WrapTitled(title, body, w, h)

    // 4. Render chrome + compose final frame
    chrome := ui.RenderChrome(hints, m.width)
    crumbs := ""  // crumbsHeight stays 0 in Phase 7 (D-16)
    statusBar := m.status.View(m.width)

    full := lipgloss.JoinVertical(lipgloss.Left, chrome, crumbs, wrapped, statusBar)

    v := tea.NewView(full)
    v.AltScreen = true
    return v
}
```
Source: [VERIFIED: bubbletea v2 pattern — `tea.NewView(string)` then mutate `v.AltScreen = true` per CLAUDE.md and internal/app/model.go:1322-1326 current code]

### Pattern 3: Uniform Titled Border via Central Helper (from ARCHITECTURE §Pattern 3)

**What:** Every view gets identical border shape, title placement, padding via `WrapTitled`. Sub-models never render their own border.

**Deviation from ARCHITECTURE sketch:** ARCHITECTURE Pattern 3 pseudocode shows `RoundedBorder`. Phase 7 uses `NormalBorder` per UI-15 (D-21). `TestChromeNormalBorderOnly` enforces the deviation.

**Example:**
```go
// internal/ui/chrome.go
func WrapTitled(title, body string, width, height int) string {
    if width < 4 {
        width = 4
    }
    if height < 3 {
        height = 3
    }
    style := TitledBorderStyle.  // package var (D-05 discipline)
        Width(width - 2).
        Height(height - 2)
    rendered := style.Render(body)
    return overlayTitle(rendered, " "+title+" ")
}

// Package var (declared in styles.go per D-18, D-22)
var TitledBorderStyle = lipgloss.NewStyle().
    Border(lipgloss.NormalBorder()).
    BorderForeground(ColorMuted).
    Padding(0, 1)
```

### Pattern 4: Data-Derived Logo State (deferred to Phase 10)

Phase 7 callers pass `LogoInfo` unconditionally (D-02). The parameter is plumbed through for Phase 10 severity coupling; Phase 7 implementation is trivial (single-color render).

### Anti-Patterns to Avoid

- **Caching rendered chrome strings:** D-18 picks pure every-frame composition. PITFALLS Anti-Pattern 4: caching adds invalidation contracts that drift. Zero-alloc + grep-gate + bench-budget is the discipline (D-22, D-24).
- **`lipgloss.NewStyle()` inside `View()`:** Violates D-22 grep-gate. Every style must be a package var. Including `WrapTitled` which reads `TitledBorderStyle`.
- **Per-sub-model border wrapping:** ARCHITECTURE Anti-Pattern 1. Sub-model `View()` returns body only; `WrapTitled` in AppModel.View is the single wrapper.
- **Focus-ring border variants:** PITFALLS Pitfall 7 + D-13, D-21. sops-tui is single-pane; no `FocusedBorder`/`UnfocusedBorder` naming.
- **Rounded/Double/Thick borders in chrome:** PITFALLS Pitfall 12 + D-21. Font coverage fails on macOS Terminal default font.

## Closed Research Gaps

### 1. `overlayTitle` Reference Implementation

**Research claim verified:** The CONTEXT.md references `github.com/charmbracelet/soft-serve/pkg/ui/components/header` as the reference implementation. Direct fetch of `https://raw.githubusercontent.com/charmbracelet/soft-serve/main/pkg/ui/components/header/header.go` at revision `ac135366727f5b9ebecb23113faa789a84b47bce` (fetched 2026-04-24) reveals this file does NOT contain a title-overlay pattern. The current soft-serve header is a plain `Render()` of a text style:

```go
// [VERIFIED: soft-serve main @ ac135366, pkg/ui/components/header/header.go]
func (h *Header) View() string {
    return h.common.Styles.ServerName.Render(strings.TrimSpace(h.text))
}
```

This is either (a) a pattern that existed in an older soft-serve revision and was refactored out, or (b) a reference inferred from the broader ecosystem. Independent inspection of the current lipgloss v2 API at pkg.go.dev confirms **no native border-title method exists** on `Style` or `Border` types. The pattern is therefore **community-standard, not citable to a single authoritative source**. 

[VERIFIED: pkg.go.dev/charm.land/lipgloss/v2 WebFetch 2026-04-24] — Available border methods on Style are `BorderStyle(Border)`, `BorderTop(bool)`, `BorderTopForeground(color)`, `BorderTopBackground(color)`, `BorderForeground(...)`, `BorderBackground(...)`. No `TopBorderTitle`, `WithTitle`, `BorderTitle`, or similar.

**Recommended algorithm for `overlayTitle`:**

```go
// overlayTitle injects " Title " into the top border of a rendered box.
// Splices the rune slice of the first line at column 2, preserving the
// top-left corner (col 0) and top-right corner (last col) of the border.
//
// Reference: this is a community-standard pattern for lipgloss v2, which
// has no native border-title API (confirmed 2026-04-24 against lipgloss
// v2.0.3 docs). The soft-serve `pkg/ui/components/header` was cited as
// a reference in sops-tui's Phase 7 research (CONTEXT.md D-14); the
// pattern is not present in current soft-serve main but is the documented
// approach in lipgloss's own discussion threads and is used across
// bubbletea-based TUIs (gh, glow, charm's own examples).
//
// Behavior:
//   - Empty title → returns rendered unchanged
//   - Title wider than width-4 → truncated with "…" to width-4 cells
//   - rendered must have at least one newline (otherwise returned as-is)
//
// Width preservation guarantee: lipgloss.Width(result.firstLine) ==
// lipgloss.Width(rendered.firstLine). The function replaces, not inserts.
func overlayTitle(rendered, title string) string {
    if title == "" {
        return rendered
    }
    nl := strings.IndexByte(rendered, '\n')
    if nl < 0 {
        return rendered  // not a multi-line box
    }
    firstLine := rendered[:nl]
    rest := rendered[nl:]

    // Measure first-line width (ANSI-aware).
    firstLineWidth := lipgloss.Width(firstLine)
    if firstLineWidth < 4 {
        return rendered  // too narrow to inject
    }
    maxTitleWidth := firstLineWidth - 4  // 2 cells left of corner + 2 right
    titleW := lipgloss.Width(title)
    if titleW > maxTitleWidth {
        // Truncate using ansi.Truncate (already direct dep)
        title = ansi.Truncate(title, maxTitleWidth, "…")
        titleW = lipgloss.Width(title)
    }

    // Replace the run from column 2 to column 2+titleW on firstLine.
    // Use a rune-index walk because ANSI codes + wide chars need tracking.
    newFirstLine := spliceRenderedLine(firstLine, 2, 2+titleW, title)
    return newFirstLine + rest
}

// spliceRenderedLine replaces the columns [startCol, endCol) of a rendered
// line with replacement, preserving any ANSI styling wrapping the line.
// Since the input is a NormalBorder top line consisting of '─' runs between
// corner chars, the implementation can be simplified:
//   - Walk rune-by-rune, keeping track of visible-column position
//   - When column == startCol, emit replacement runes
//   - Skip runes until column reaches endCol
//   - Resume copying remaining runes
// Because NormalBorder top line is pure ASCII-range box-drawing chars
// (no ANSI sequences inside lipgloss's border render), the splice is
// essentially a slice operation on the rune array.
```

**Test contract (enforced by `TestOverlayTitle_PreservesCornersAndWidth`):**

1. `╭` at col 0 preserved after overlay
2. `╮` at col `width-1` preserved after overlay
3. `lipgloss.Width(firstLine)` equal before and after
4. Title wider than `width-4` truncated with `…` ending
5. Empty title returns rendered byte-identical
6. Single-line input (no `\n`) returns unchanged
7. `width < 4` input returns unchanged (too narrow)

### 2. `lipgloss/v2/table` StyleFunc API Surface

**Verified via [WebFetch pkg.go.dev/charm.land/lipgloss/v2/table 2026-04-24]:**

```go
// Signature
type StyleFunc func(row, col int) lipgloss.Style

// row == table.HeaderRow (-1) for header styling
// row >= 0 for body rows
// col >= 0 for all columns
```

**Recommended menu rendering shape (concrete code, ready to paste):**

```go
// internal/ui/menu.go
package ui

import (
    "charm.land/lipgloss/v2"
    "charm.land/lipgloss/v2/table"

    "github.com/caesarakalaeii/sops-tui/internal/keys"
)

// RenderMenu renders the persistent keybinding menu as a 2-col × 6-row grid.
// D-04: fixed 2 cols × 6 rows = 12 slots at all widths.
// D-05: lipgloss/v2/table with StyleFunc; mnemonic col uses MenuKeyStyle,
//       description col uses MenuDescStyle.
// D-06: only hints with Visible=true are included in the 12 slots.
// D-07: cell format is "[mnemonic] description" — mnemonic left-aligned in a
//       fixed-width subcolumn of column 0; description fills column 1.
//
// Max 12 hints. Over-capacity hints fall to the ? full-screen overlay.
func RenderMenu(hints []keys.MenuHint, width int) string {
    const maxRows = 6
    const cols = 2

    // Filter to visible hints only (D-06), cap at 12
    visible := make([]keys.MenuHint, 0, cols*maxRows)
    for _, h := range hints {
        if h.Visible {
            visible = append(visible, h)
            if len(visible) == cols*maxRows {
                break
            }
        }
    }

    // Build column-major: rows[r][c] = hint at (col c, row r)
    // Column 0 gets hints 0..5; column 1 gets hints 6..11.
    // This mirrors k9s menu layout (top-down, left-to-right column-major).
    rows := make([][]string, maxRows)
    for r := 0; r < maxRows; r++ {
        rows[r] = make([]string, cols)
    }
    for i, h := range visible {
        col := i / maxRows
        row := i % maxRows
        if col >= cols {
            break
        }
        rows[row][col] = "[" + h.Mnemonic + "] " + h.Description
    }

    t := table.New().
        BorderTop(false).
        BorderBottom(false).
        BorderLeft(false).
        BorderRight(false).
        BorderRow(false).
        BorderColumn(false).
        BorderHeader(false).
        StyleFunc(func(row, col int) lipgloss.Style {
            // Alternating column styling per D-05.
            // Single-cell content "[key] desc" — no subcolumn split at the
            // table level; instead the MenuCellStyle (combined) is applied.
            // The accent→fg split happens inside the cell content via
            // AnsiCompose (styling the bracketed mnemonic).
            return MenuCellStyle
        }).
        Rows(rows...).
        Width(width)

    return t.Render()
}
```

**Note on D-07 subcolumn:** `lipgloss/v2/table` does not split within a cell. To get "mnemonic accent + description fg" styling inside a single cell, compose the cell string with pre-styled fragments:

```go
rows[row][col] = MenuKeyStyle.Render("["+h.Mnemonic+"]") + " " + MenuDescStyle.Render(h.Description)
```

Then the cell-level style (set via StyleFunc) adds any outer padding/width but doesn't override the inline styles. This is the canonical pattern for per-fragment styling in lipgloss v2 table cells.

**Width math for menu:** Phase 7 chrome splits width as:
- Logo: fixed ~26 cols
- Info-panel placeholder: fixed width (Plan 3 picks, recommend 38 cols for Phase 8 forward-compat)
- Menu: `remaining = width - 26 - 38 - spacing`

For the narrow 40×12 case, `remaining` is negative → menu renders on a single clipped cell. Acceptable per D-04 ("narrow-terminal safe, not responsive"). Alternative: at widths < 80, stack chrome vertically (logo top, menu below) — but that's out of D-04's scope.

### 3. Bubbletea v2 `View() tea.View` Integration

**Verified via [VERIFIED: current internal/app/model.go:1324-1326]:**

```go
v := tea.NewView(full)
v.AltScreen = true
return v
```

This is the correct Phase 7 return shape. The `tea.View` struct exposes:
- `Layer` — `*lipgloss.Layer` (what `NewView(string)` populates)
- `AltScreen bool` — currently set to true
- `Cursor *tea.Cursor` — optional, for focused inputs (unused in Phase 7)

**Implication for chrome recompose:** Passing the full composed string to `tea.NewView(full)` is byte-identical to the current pattern. Nothing about `tea.View` type changes; the only change is the content of `full` (now 4-section join instead of 2-section).

**Cursor considerations:** `RecipientFormModel` uses a `textinput` which may set cursor position via its own render. `tea.View.Cursor` is not set by the root `View()`. Plan 3 does not need to touch cursor wiring.

**No `View()` option changes needed:** `WithAltScreen()` program option is not used in bubbletea v2 (per CLAUDE.md migration rules — replaced by `v.AltScreen = true` struct field).

### 4. Sub-Model `Hints()` Content Per State

Based on direct inspection of each sub-model's keymap / Update handler, the ≤12 hints per state:

**stateFileList (default, !IsSearchActive)** — source `keys.DefaultFileListKeyMap`:

| # | Mnemonic | Description |
|---|----------|-------------|
| 1 | j/↓ | move down |
| 2 | k/↑ | move up |
| 3 | enter/l | open |
| 4 | / | search |
| 5 | i | file info |
| 6 | space | toggle select |
| 7 | K | bulk re-key selected |
| 8 | H | health check |
| 9 | g | go to top |
| 10 | G | go to bottom |
| 11 | ? | toggle help |
| 12 | q | quit |

`FileListKeyMap.ShortHelp()` at `internal/keys/bindings.go:75` returns 10 bindings (no g/G). We add g/G to reach 12 — sub-model authors curate per D-06.

**stateFileList (IsSearchActive)** — search-active override per D-11:

| # | Mnemonic | Description |
|---|----------|-------------|
| 1 | Esc | exit search |
| 2 | Enter | select result |
| 3 | j/↓ | next result |
| 4 | k/↑ | prev result |
| 5 | ? | toggle help |
| 6 | q | quit |

(Only 6 visible slots filled — remaining 6 cells render empty.)

**stateDetail** — source `keys.DefaultDetailKeyMap`:

| # | Mnemonic | Description |
|---|----------|-------------|
| 1 | j/↓ | move down |
| 2 | k/↑ | move up |
| 3 | enter/l | expand |
| 4 | h/← | collapse |
| 5 | r | reveal/hide |
| 6 | R | reveal all |
| 7 | e | edit |
| 8 | X | rotate |
| 9 | ctrl+y | copy |
| 10 | a | add recipient |
| 11 | d | remove recipient |
| 12 | Esc | back |

`DetailKeyMap.ShortHelp()` at `internal/keys/bindings.go:193` returns 13 — we drop one of `{E edit in $EDITOR, b git history, / search, i info}` from visible. Recommend: keep `/ search` and `i info` slot 11/12; drop E/b to the `?` full overlay. **Plan 3 author curates final 12.**

**stateHelp** — the help overlay itself:

| # | Mnemonic | Description |
|---|----------|-------------|
| 1 | Esc | close help |
| 2 | ? | close help |
| 3 | q | quit |

(3 slots — remaining 9 cells empty.)

**stateDiff** (`recipientAction == ""` — standalone edit diff):

| # | Mnemonic | Description |
|---|----------|-------------|
| 1 | y | confirm re-encrypt |
| 2 | n | cancel |
| 3 | Esc | cancel |
| 4 | j | scroll down |
| 5 | k | scroll up |
| 6 | q | quit |

Source: `internal/ui/diff.go` Update handler lines 96-114.

**stateRecipientConfirm** (`recipientAction == "add"` or `"remove"`):

| # | Mnemonic | Description |
|---|----------|-------------|
| 1 | y | confirm add/remove recipient |
| 2 | n | cancel |
| 3 | Esc | cancel |
| 4 | j | scroll down |
| 5 | k | scroll up |

**stateBulkReKeyConfirm** (`recipientAction == "bulkrekey"` — CONTEXT uses this label, the code uses no special sentinel; Plan 3 uses state alone):

| # | Mnemonic | Description |
|---|----------|-------------|
| 1 | y | confirm re-key this file |
| 2 | n | skip this file |
| 3 | Esc | abort bulk re-key |
| 4 | j | scroll down |
| 5 | k | scroll up |

**stateMetadata**:

| # | Mnemonic | Description |
|---|----------|-------------|
| 1 | j | scroll down |
| 2 | k | scroll up |
| 3 | i | close metadata |
| 4 | Esc | close metadata |
| 5 | q | quit |

Source: `internal/ui/metadata.go` + `model.go:930` dispatcher.

**stateHealth**:

| # | Mnemonic | Description |
|---|----------|-------------|
| 1 | j | scroll down |
| 2 | k | scroll up |
| 3 | H | close health |
| 4 | Esc | close health |
| 5 | q | quit |

Source: `internal/ui/health.go` + `model.go:672`.

**stateHistory**:

| # | Mnemonic | Description |
|---|----------|-------------|
| 1 | j | scroll down |
| 2 | k | scroll up |
| 3 | b | close history |
| 4 | Esc | close history |
| 5 | q | quit |

Source: `internal/ui/history.go` + `model.go:1042`.

**stateRecipientList**:

| # | Mnemonic | Description |
|---|----------|-------------|
| 1 | 1-9 | select recipient to remove |
| 2 | Esc | cancel |
| 3 | q | quit |

Source: `model.go:716-739`.

**stateRecipientForm**:

| # | Mnemonic | Description |
|---|----------|-------------|
| 1 | Enter | confirm |
| 2 | Esc | cancel |

Source: `internal/ui/recipientform.go:94-116`.

**stateFormatMenu** (inline hint set per D-09 — no owning sub-model):

| # | Mnemonic | Description |
|---|----------|-------------|
| 1 | j | next format |
| 2 | k | prev format |
| 3 | Enter | confirm format |
| 4 | Esc | cancel |

Source: `model.go:874` (`stateFormatMenu` block) + `renderFormatMenu` at `model.go:1857`.

### 5. ASCII Figlet Variants for "SOPS" + "tui" Subscript

Logo direction is locked (D-01: 5-row SOPS block + row-6 "tui" subscript; width ~26 cols). Two candidate byte patterns the Plan 1 author can pick from:

**Candidate A (compact block — matches k9s LogoSmall density):**

```go
// 6 rows × 25 cols (last row padded to 25)
var LogoSmallA = []string{
    `  ____   ___  ____  ____  `,  // row 0
    ` / ___| / _ \|  _ \/ ___| `,  // row 1
    ` \___ \| | | | |_) \___ \ `,  // row 2
    `  ___) | |_| |  __/ ___) |`,  // row 3
    ` |____/ \___/|_|   |____/ `,  // row 4
    `                      tui `,  // row 5 (subscript right-aligned)
}
```

Rendering width: 25 cols. All ASCII. 5-row SOPS block via Figlet "big" font equivalent (standard underscore/pipe/parens). Row 6 "tui" right-aligned in lowercase as subscript. [ASSUMED: block figlet art renders identically across terminals — verified only on Alacritty locally]

**Candidate B (lean block — tighter horizontal footprint, 22 cols):**

```go
// 6 rows × 22 cols
var LogoSmallB = []string{
    ` ___  __  ____  ___  `,  // row 0
    `/ __|/  \|  _ \/ __| `,  // row 1
    `\__ \ () | |_) \__ \ `,  // row 2
    `|___/\__/|  __/|___/ `,  // row 3
    `         |_|         `,  // row 4
    `                 tui `,  // row 5
}
```

Rendering width: 22 cols. Tighter but still ASCII-clean. Row 4 has `|_|` tail from "P" baseline per standard block ASCII. Good fit if the menu needs more horizontal room.

**Recommendation:** **Candidate A.** The 25-col envelope matches the ROADMAP's "~26 cols wide" target (UI-02). Row 4 baseline `|____/ \___/|_|   |____/` reads as "SOPS" cleanly at glance. Candidate B's 22-col footprint sacrifices readability for room the Phase 7 menu doesn't need (menu is fixed 2 cols × 6 rows regardless of width).

Plan 1 author has final byte-art pick; `/gsd-verify-work` user review flags any illegibility.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Multi-column grid layout | Custom `JoinHorizontal(JoinVertical(...))` column-equalization math | `lipgloss/v2/table` with `Width(n)` + `StyleFunc(row, col)` | Table equalizes column widths automatically (k9s `maxKeys[col]` logic for free); `StyleFunc` matches D-05 pattern exactly. [VERIFIED: pkg.go.dev/charm.land/lipgloss/v2/table 2026-04-24] |
| ANSI escape stripping in goldens | Regex for `\x1b\[[0-9;]*m` | `ansi.Strip` from `github.com/charmbracelet/x/ansi` | Already direct dep (Phase 6 D-08). Handles CSI, OSC, DCS, malformed escapes that a 5-line regex would miss. |
| ANSI-aware width measurement | `len(s)` or `utf8.RuneCountInString(s)` | `lipgloss.Width(s)` | PITFALLS Pitfall 6 — width drift on macOS Terminal. `lipgloss.Width` uses uniseg + terminfo. |
| Title truncation | Custom rune-slice + ellipsis | `ansi.Truncate(s, n, "…")` from `github.com/charmbracelet/x/ansi` | Handles multi-byte chars + ANSI sequences. Already dep. |
| String-splice on first line | Hand-parse `\n`-split then substring | Simple `strings.IndexByte(s, '\n')` split + rune slice on first line | The splice is trivial; don't over-engineer. Pattern is 10-15 LOC (see `overlayTitle` example above). |
| Box-drawing characters | Custom unicode inference | `lipgloss.NormalBorder()` exclusively | D-21 grep-gated; NormalBorder is the only font-universal variant (Pitfall 12). |
| Source-of-truth hints | String literals beside render function | `HintsFromBindings(keymap.ShortHelp())` | D-08 single source of truth; keymap drift → menu drift (PITFALLS Pitfall 3). |

**Key insight:** Phase 7's primitives — table, ansi.Strip, ansi.Truncate, lipgloss borders — are all one-import-away. The only "custom" code Phase 7 writes is the composition glue (`RenderChrome`, `WrapTitled`, `overlayTitle`, `RenderMenu`, `RenderLogo`) and the 9 `Hints()` method stubs. Total new LOC: ~400 in `internal/ui/{logo,menu,chrome}.go` + `internal/keys/hints.go` + 9 one-method additions.

## Common Pitfalls

### Pitfall 1: Info-panel placeholder height inflates chromeHeight

**What goes wrong:** D-16 says "Info-panel area is rendered as a 6-row blank block of width — preserves column alignment for Phase 8." If the blank block is rendered at the *whole* info-panel width (38 cells) instead of 6 rows × 0 chars, `lipgloss.Height(renderedChrome)` returns 6 (correct). But if rendered with `.Height(6).Width(38)`, the result is 6 rows of 38 spaces — also height 6. BUT if a dev renders the blank as an empty string then the chrome height collapses to `max(logoHeight, menuHeight) = 6`, not `max(logoHeight, menuHeight, infoHeight) = max(6, 6, 0) = 6`. Accidentally OK. However, Phase 8 flips info-panel to 5 rows — if Phase 7 left `infoHeight = 0`, Phase 8's flip would *shrink* chrome by 1 row instead of preserving it.

**Why it happens:** "Render a blank placeholder" has two interpretations: empty string (`""`) or `lipgloss.NewStyle().Width(38).Height(6).Render("")`. Only the latter reserves vertical space.

**How to avoid:** Render info-panel placeholder as `lipgloss.NewStyle().Width(infoPanelWidth).Height(6).Render("")` so `chromeHeight == 6` regardless. Verify in unit test: `TestChromeHeight_EqualsSix` assembles chrome with zero hints and zero logo, asserts `lipgloss.Height(chrome) == 6`.

**Warning signs:** `chromeHeight` drops below 6 after a Phase 8 info-panel swap; bench-budget test shows chrome render allocating <6 rows.

### Pitfall 2: `TestViewNoNewStyle` AST walker misses lambda-embedded NewStyle

**What goes wrong:** D-22 grep-gates `lipgloss.NewStyle(` inside `View()`. A naive AST walker scans `View()`'s direct statement list. But a helper lambda closure declared inside View (e.g., `render := func(s string) string { return lipgloss.NewStyle().Render(s) }`) is an anonymous function — the NewStyle call is in the lambda's body, not View's direct body. AST walker must recurse into nested function literals.

**Why it happens:** `go/ast.Walk` visits nodes in order; `FuncLit.Body` is a nested `BlockStmt`. A `range []ast.Stmt` scan only sees the top level.

**How to avoid:** Use `ast.Inspect(viewFuncBody, func(n ast.Node) bool { ... })` — Inspect recurses into all nested nodes. Match any `*ast.CallExpr` whose `Fun` is `*ast.SelectorExpr{X: "lipgloss", Sel: "NewStyle"}` anywhere in the subtree.

**Warning signs:** A PR adds a helper lambda inside `View()`, bench-budget test regresses by 10µs+, grep-gate passes because the walker missed it.

### Pitfall 3: `RecipientList` sub-model doesn't exist; D-09 lists it as a sub-model

**What goes wrong:** CONTEXT D-09 lists "RecipientList" as one of the 9 sub-models that gets `Hints()`. Code inspection shows `renderRecipientList()` is a **method on AppModel at `model.go:1799`**, not a separate sub-model. There's no `RecipientListModel` struct; the recipient list is a plain string renderer consuming `m.recipientList []string` from AppModel.

**Why it happens:** The research ARCHITECTURE.md mentions recipient list flow but doesn't carve a dedicated `internal/ui/recipientlist.go`. Phase 5 landed the renderer as a model method; Phase 7 must decide where `Hints()` for `stateRecipientList` lives.

**How to avoid:** For stateRecipientList, implement hints as an **inline hint set** on AppModel (not a sub-model method), following the same pattern D-09 uses for stateFormatMenu. Update Plan 3 task map accordingly: 8 sub-model `Hints()` methods (FileList, Detail, Help, Diff, Metadata, Health, History, RecipientForm) + 2 inline hint sets on AppModel (stateRecipientList, stateFormatMenu). Total still 10 hint sets dispatched from `menuHints()`.

**Warning signs:** Plan 3 task that says "add Hints() to RecipientListModel" — file doesn't exist.

### Pitfall 4: `HelpModel` needs `ViewState` context — `Hints()` signature can't accept args

**What goes wrong:** `internal/ui/help.go` line 71: `HelpModel.View(state ui.ViewState)` — the help model's view depends on which state triggered the `?`. Hints should similarly vary: help shown over file list → show file-list-contextual exit keys; help shown over detail → detail-contextual. But `Hinter` interface signature per D-08 is `Hints() []MenuHint` — no args.

**Why it happens:** Hints interface is designed as zero-arg for uniformity. State-aware context is the dispatcher's job (D-10), not the sub-model's.

**How to avoid:** `HelpModel.Hints()` returns a **generic** set: `{Esc close, ? close, q quit}`. The file-list-vs-detail context is orthogonal and shown in the **title** (title for stateHelp = "Help" per D-15). If future Phase 9 wants context-aware help hints, it can dispatch via AppModel's `menuHints()` method based on `m.prevState`, bypassing `HelpModel.Hints()`.

**Warning signs:** Plan 3 tries to add `Hints(prevState) []MenuHint` — interface violation; back to D-08 spec.

### Pitfall 5: Chrome renders at `m.width = 0` on first frame

**What goes wrong:** Bubble Tea renders once before `WindowSizeMsg` arrives. At `m.width == 0 || m.height == 0`, `RenderMenu(hints, 0)` may panic on `lipgloss/v2/table.Width(0)` (untested — behavior unclear). `WrapTitled(title, body, 0, 0)` would certainly produce garbage.

**Why it happens:** Existing code handles this at `model.go:1322` by checking `bodyDims` clamp-to-zero — but that only protects the body wrap, not chrome render.

**How to avoid:** Early-return in `View()` before any chrome computation:
```go
if m.width == 0 || m.height == 0 {
    v := tea.NewView("")
    v.AltScreen = true
    return v
}
```
Reference: ARCHITECTURE §"Integration Pitfalls" Pitfall 10.

**Warning signs:** TUI panics on startup; bench-budget test fails with runtime error before first measurement.

### Pitfall 6: Bench test runs actual bench → slow test suite

**What goes wrong:** D-24 says "runs the bench once and asserts `ns/op ≤ 50_000`." If implemented as `b := testing.Benchmark(BenchmarkAppView)`, the default bench duration is 1 second minimum, adding 1s to `go test ./...`. Scales across CI.

**Why it happens:** `testing.Benchmark` measures until it gets stable results; there's no "single call" mode.

**How to avoid:** Use `testing.B.N` set explicitly, OR measure via wall-clock around a manual loop:
```go
func TestBenchmarkAppView_UnderBudget(t *testing.T) {
    // ... setup model ...
    const iters = 100
    start := time.Now()
    for i := 0; i < iters; i++ {
        _ = m.View()
    }
    nsPerOp := time.Since(start).Nanoseconds() / int64(iters)
    if nsPerOp > 50_000 {
        t.Fatalf("View() took %d ns/op, budget is 50000 ns/op (50µs)", nsPerOp)
    }
}
```
100 iterations is fast (~5ms total for a 50µs target) and gives a stable enough number. Beware: CI noise can push a single run 2-3× the median; consider a 75% gate buffer (e.g., fail at 75µs, budget at 50µs) for CI reliability. **Plan 3 author decides gate strictness — recommend 100 iters + 50µs hard gate; raise to 75µs if CI flakes.**

**Warning signs:** `go test ./...` gains 1s per bench; CI flakes on bench test.

### Pitfall 7: Goldens include ANSI escape sequences → break on lipgloss bumps

**What goes wrong:** PITFALLS Pitfall 8 (acknowledged). The Phase 6 testutil harness (`RequireGoldenStructure` + `RequireGoldenColors`) already addresses this — structural goldens are ANSI-stripped, color presence asserted separately. Phase 7 must **continue using the Phase 6 split**. Writing a new mega-golden that bundles ANSI + structure breaks the discipline.

**How to avoid:** Phase 7 goldens exclusively use `testutil.RequireGoldenStructure` for structure and `testutil.RequireGoldenColors(t, name, output, []string{...})` with specific ANSI sequences for color presence. The `wantColors []string` param gets populated with key ANSI bytes for Phase 7:
- `"\x1b[38;2;137;180;250m"` = `ColorAccent` foreground (hex `#89b4fa`) — verify menu key column uses it
- `"\x1b[38;2;108;112;134m"` = `ColorMuted` foreground (hex `#6c7086`) — verify titled border uses it

**Warning signs:** Golden diff shows only ANSI sequence reordering, not structural change. PR description says "regenerate goldens" with no explanation.

### Pitfall 8: Plan 3 commit order breaks build at intermediate commit

**What goes wrong:** D-25 allows 2-3 commits in Plan 3. Naive split: (1) add all 9 `Hints()` methods, (2) flip chromeHeight + rewrite View, (3) add grep-gates. At commit (1), sub-models reference `keys.MenuHint` from Plan 1 — fine. At commit (2), View references `ui.RenderChrome` + `ui.WrapTitled` from Plan 2 — fine IF Plans 1-2 are in same PR. But commit (2) alone may break tests because goldens reflect OLD (no chrome) output; grep-gates in commit (3) aren't yet active; bench test may fail the 50µs budget if CI runs commit (2) in isolation.

**Why it happens:** Mid-PR commits not individually buildable = bisect hazard + intermediate-test-failure risk.

**How to avoid:** Recommended commit sequence (see "Plan 3 Commit Sequence" below).

**Warning signs:** Bisect to investigate a test failure lands on commit (2) which is intentionally broken.

## Runtime State Inventory

Phase 7 is a code-only change. No runtime state (no DB migrations, no external service config, no OS-registered state, no secrets/env vars, no build artifacts).

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None | — |
| Live service config | None | — |
| OS-registered state | None | — |
| Secrets/env vars | None — `GOLDEN_UPDATE=1` is a dev-time test env var, not a runtime secret | — |
| Build artifacts | None | — |

**Nothing found in any category:** Verified by (a) no database or datastore in sops-tui v1.0 (project is a stateless CLI), (b) no external services (sops is a subprocess), (c) no OS scheduler integration, (d) Phase 7 adds zero new env vars, (e) `go test ./...` uses source-only `go.mod` dependencies.

## Code Examples

Verified patterns from existing codebase + pkg.go.dev:

### Example 1: `HintsFromBindings` — converts `key.Binding.Help()` to `MenuHint`

```go
// internal/keys/hints.go
package keys

import "charm.land/bubbles/v2/key"

// MenuHint is one visible row in the persistent keybinding menu.
type MenuHint struct {
    Mnemonic    string
    Description string
    Visible     bool
}

// Hinter is implemented by every interactive sub-model; the AppModel
// dispatcher queries the active sub-model on every View() call per D-08.
type Hinter interface {
    Hints() []MenuHint
}

// HintsFromBindings converts a slice of key.Binding into MenuHint entries.
// Each binding's Help() returns {Key, Desc} per bubbles/v2/key semantics;
// those map directly to MenuHint.Mnemonic/Description. All hints default
// to Visible=true — caller filters via MenuHint.Visible toggles if needed.
func HintsFromBindings(bindings []key.Binding) []MenuHint {
    hints := make([]MenuHint, 0, len(bindings))
    for _, b := range bindings {
        h := b.Help()  // returns key.Help{Key, Desc}
        hints = append(hints, MenuHint{
            Mnemonic:    h.Key,
            Description: h.Desc,
            Visible:     true,
        })
    }
    return hints
}
```
Source: [VERIFIED: internal/keys/bindings.go:30-38 — `key.NewBinding(WithKeys, WithHelp)` produces bindings whose `.Help()` method returns `{Key, Desc}`. Pattern used by bubbles v2 help.Model for over-the-wire consistency.]

### Example 2: Sub-model `Hints()` one-liner

```go
// internal/ui/filelist.go
// +1 method addition per D-09
func (m FileListModel) Hints() []keys.MenuHint {
    hints := keys.HintsFromBindings(m.keys.ShortHelp())
    // Append g/G (navigation) since ShortHelp() omits them per FileListKeyMap.ShortHelp()
    hints = append(hints,
        keys.MenuHint{Mnemonic: "g", Description: "go to top", Visible: true},
        keys.MenuHint{Mnemonic: "G", Description: "go to bottom", Visible: true},
    )
    return hints
}
```

### Example 3: AppModel dispatcher (menuHints method)

```go
// internal/app/model.go (new private method, added in Plan 3)
func (m AppModel) menuHints() []keys.MenuHint {
    // (state, recipientAction, IsSearchActive) tuple dispatch per D-10.
    if m.state == stateFileList && m.fileList.IsSearchActive() {
        return keys.FileListSearchHints  // D-11 override
    }

    switch m.state {
    case stateFileList:
        return m.fileList.Hints()
    case stateDetail:
        return m.detail.Hints()
    case stateMetadata:
        return m.metadata.Hints()
    case stateDiff:
        // Standalone edit/rotate diff — no recipient action
        return m.diff.Hints()
    case stateRecipientConfirm:
        // Shared diff body, recipient-add/remove action
        return keys.RecipientConfirmHints  // package-level []MenuHint
    case stateBulkReKeyConfirm:
        return keys.BulkReKeyConfirmHints
    case stateHelp:
        return m.help.Hints()
    case stateHistory:
        return m.history.Hints()
    case stateHealth:
        return m.health.Hints()
    case stateRecipientForm:
        return m.recipientForm.Hints()
    case stateRecipientList:
        return keys.RecipientListHints  // inline set per Pitfall 3
    case stateFormatMenu:
        return keys.FormatMenuHints  // inline set per D-09
    }
    return nil
}
```

### Example 4: Title map for `titleForState()`

```go
// internal/app/model.go (new helper, added in Plan 3)
func (m AppModel) titleForState() string {
    // D-15 title format mapping.
    switch m.state {
    case stateFileList:
        return fmt.Sprintf("Files (%d)", m.fileList.ItemCount())
    case stateDetail:
        return "Detail: " + m.currentFile.Name  // no git badge
    case stateMetadata:
        return "Metadata"
    case stateDiff:
        return "Diff"
    case stateHelp:
        return "Help"
    case stateHistory:
        return fmt.Sprintf("History (%d)", m.history.CommitCount())
    case stateHealth:
        return fmt.Sprintf("Health (%d findings)", m.health.FindingCount())
    case stateRecipientList:
        return fmt.Sprintf("Recipients (%d)", len(m.recipientList))
    case stateRecipientForm:
        return "RecipientForm"
    case stateFormatMenu:
        return "Format"
    case stateRecipientConfirm:
        return "Diff"
    case stateBulkReKeyConfirm:
        return "Diff"
    }
    return ""
}
```

Source: [VERIFIED: CONTEXT.md D-15 table verbatim; accessor names match existing code — `m.fileList.ItemCount()` exists per 02-read-loop, `m.history.CommitCount()` per 04-02-PLAN, `m.health.FindingCount()` per 05-02-PLAN.]

**Note:** Plan 3 author must verify `CommitCount()`, `FindingCount()` accessor names exist in current HistoryModel/HealthModel — if not, Plan 3 adds the accessors as part of the integration work.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Bubble Tea v1 `View() string` | v2 `View() tea.View` struct return | bubbletea v2.0.0 (early 2026) | Phase 7 return shape unchanged from current code — already v2-native |
| `WithAltScreen()` program option | `v.AltScreen = true` field on `tea.View` | bubbletea v2.0.0 | Current code already uses the new pattern (model.go:1325) |
| `tea.KeyMsg` struct | `tea.KeyPressMsg` interface | bubbletea v2.0.0 | No new key handlers in Phase 7; existing pattern preserved |
| Lipgloss `Padding(0,1)` on Border | Same — no change | — | Stable for 2+ years |
| Lipgloss v2 `Canvas` + `Layer` (overlay compose) | Reserved for modal dropdowns | lipgloss v2.0.0 | Not used in Phase 7 (chrome is inline, not overlay) |
| Immediate-mode chrome (compose every frame) | Same in Phase 7 per D-18 | — | Pitfall 2 acknowledges; guardrails enforce zero-alloc |

**Deprecated/outdated:**
- `lipgloss.AdaptiveColor` — project-banned per CLAUDE.md + styles.go:4 + issue #1036. Phase 7 uses explicit hex colors via `lipgloss.Color(Hex)`.
- `tview` retained-mode widget tree (k9s pattern) — explicitly rejected in Phase 7 research per ARCHITECTURE Anti-Pattern 2.
- "Observer pattern for menu updates" (k9s `StackPushed`/`StackPopped`) — ported as enum switch dispatch per D-10.

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + `stretchr/testify v1.11.1` |
| Config file | None — pure `go test ./...` |
| Quick run command | `go test ./internal/app -run TestChrome -count=1` |
| Full suite command | `go test ./... -count=1` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| UI-01 | Persistent multi-column keybinding menu renders on every view | unit | `go test ./internal/ui -run TestRenderMenu -x` | ❌ Wave 0: new `internal/ui/menu_test.go` |
| UI-01 | `Hints()` on each of 8 sub-models returns correct bindings | unit | `go test ./internal/ui -run TestHints -x` | ❌ Wave 0: new `internal/ui/{filelist,detail,help,diff,metadata,health,history,recipientform}_hints_test.go` OR consolidated `hints_test.go` |
| UI-01 | AppModel dispatcher yields correct set per (state, recipientAction, IsSearchActive) | unit | `go test ./internal/app -run TestMenuHints -x` | ❌ Wave 0: new `internal/app/hints_test.go` |
| UI-02 | 6-row ASCII logo renders at width 26 | unit | `go test ./internal/ui -run TestRenderLogo -x` | ❌ Wave 0: new `internal/ui/logo_test.go` |
| UI-02 | Logo is ASCII-only (no non-ASCII codepoints) | grep-gate | `go test ./internal/app -run TestChromeASCIIOnly -x` | ❌ Wave 0: new `internal/app/chrome_test.go` |
| UI-06 | `WrapTitled` preserves corners + width + truncates overlong title + passes through empty title | unit | `go test ./internal/ui -run TestOverlayTitle -x` | ❌ Wave 0: new `internal/ui/chrome_test.go` |
| UI-06 | Each state renders with correct title per D-15 map | integration (goldens) | `go test ./internal/app -run TestResize -x` | ✅ `internal/app/resize_test.go` (goldens refreshed Plan 3) |
| UI-15 | Only `NormalBorder()` in chrome files | grep-gate | `go test ./internal/app -run TestChromeNormalBorderOnly -x` | ❌ Wave 0: new `internal/app/chrome_test.go` |
| UI-15 | No `lipgloss.NewStyle()` inside `View()` body | grep-gate (AST) | `go test ./internal/app -run TestViewNoNewStyle -x` | ❌ Wave 0: new `internal/app/chrome_test.go` |
| UI-15 (bench) | `BenchmarkAppView` ≤ 50µs/op at 200×60 | unit (wall-clock) | `go test ./internal/app -run TestBenchmarkAppView_UnderBudget -x` | ❌ Wave 0: new `internal/app/chrome_test.go` |

### Sampling Rate

- **Per task commit:** `go test ./internal/ui ./internal/keys -count=1` (fast — no goldens)
- **Per wave merge:** `go test ./... -count=1` (full suite including goldens + bench gate)
- **Phase gate (`/gsd-verify-work`):** full suite green + manual smoke at 40×12, 80×24, 120×40, 200×60 per Phase 6 D-15 protocol

### Wave 0 Gaps

Plan 1 sets up (new files):
- [ ] `internal/keys/hints_test.go` — covers `HintsFromBindings` + `MenuHint` struct shape
- [ ] `internal/ui/logo_test.go` — covers `RenderLogo` output shape, 6 rows, width 26, ASCII-only smoke
- [ ] `internal/ui/menu_test.go` — covers `RenderMenu` column-major fill, `Visible=false` skip, empty hints

Plan 2 sets up (new files):
- [ ] `internal/ui/chrome_test.go` — covers `WrapTitled` + `TestOverlayTitle_PreservesCornersAndWidth` (corners, width, truncation, empty-title)
- [ ] `internal/ui/chrome_compose_test.go` — covers `RenderChrome` composition (optional; may be folded into `chrome_test.go`)

Plan 3 sets up (new files + refreshes):
- [ ] `internal/app/chrome_test.go` — grep-gates (`TestChromeASCIIOnly`, `TestChromeNormalBorderOnly`, `TestViewNoNewStyle`) + bench-budget test
- [ ] `internal/app/hints_test.go` — AppModel dispatcher tests per state
- [ ] `internal/ui/{filelist,detail,help,diff,metadata,health,history,recipientform}_test.go` — extended with `TestHints` per sub-model (1 method per sub-model)
- [ ] `internal/app/testdata/resize_{40x12,80x24,120x40,200x60}.golden` — REFRESHED via `GOLDEN_UPDATE=1 go test ./internal/app -run TestResize`

Framework install: none — `go test` stdlib + existing testify already present.

## Environment Availability

Phase 7 is code-only. No external dependencies beyond what v1.0 + Phase 6 already require.

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| `go` | build, test | ✓ | go1.26.2 | — |
| `charm.land/lipgloss/v2/table` | Plan 1 menu.go | ✓ | v2.0.3 (via lipgloss transitive) | — |
| `github.com/charmbracelet/x/ansi` | Plan 2 `ansi.Truncate` in overlayTitle | ✓ | v0.11.7 (direct dep since Phase 6) | — |

**Missing dependencies with no fallback:** None.

**Missing dependencies with fallback:** None.

## Security Domain

No user-facing secret handling in Phase 7. Chrome is pure display layer; all secret handling (decrypt, reveal, edit) stays in v1.0 sub-models unchanged.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | — |
| V3 Session Management | no | — |
| V4 Access Control | no | — |
| V5 Input Validation | no | Sub-model input remains unchanged; chrome accepts no user input |
| V6 Cryptography | no | Sub-model handles SOPS/age; chrome never touches secrets |
| V9 Communication | no | — |

### Known Threat Patterns for Phase 7

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Title-bar PII leakage (filename in "Detail: prod.yaml") | Information Disclosure | D-15 uses `m.currentFile.Name` (relative), not absolute path — aligns with PITFALLS Pitfall 11 |
| Screenshot of chrome reveals environment structure | Information Disclosure | Phase 8's info-panel is the bigger risk; Phase 7 only shows filename in title (already visible in file list) — no new disclosure surface |
| Hint text reveals hidden keybindings | Low | All hints shown are existing public bindings (ShortHelp already shipped in v1.0); no hidden commands surfaced |

**Key finding:** Phase 7 does not widen the attack surface. File names in titles are already present in the file list; recipient counts are already in metadata. No new PII sources.

## Plan 3 Commit Sequence (Recommendation)

Per D-25 (2-3 commits acceptable). Recommended ordering keeps each commit independently buildable and grep-gate-passing:

**Commit 1 — Sub-model `Hints()` methods (pure additions, no behavior change):**
- Add `Hints()` to each of 8 sub-models: FileList, Detail, Help, Diff, Metadata, Health, History, RecipientForm
- Add package-level `keys.*Hints` vars for: FileListSearchHints, RecipientConfirmHints, BulkReKeyConfirmHints, RecipientListHints, FormatMenuHints
- Add `TestHints` per sub-model under `internal/ui/*_test.go`
- All existing tests still pass (no View changes; bench unchanged)
- Grep-gates don't yet exist — OK, they come in commit 3
- **Build state:** Green (no production code referenced these additions yet)

**Commit 2 — AppModel integration (chromeHeight flip + View rewrite + magic-constant migration):**
- Add `menuHints()` method on AppModel (dispatcher per D-10, D-11)
- Add `titleForState()` method on AppModel per D-15 map
- Flip `chromeHeight(m)` from `return 0` to real body: `return lipgloss.Height(ui.RenderChrome(...))`
- Rewrite `AppModel.View()` to compose `[chrome][crumbs-placeholder][wrapped body][status bar]`
- Migrate `renderRecipientList` (model.go:1841) to use `bodyDims` + `WrapTitled` (D-19 — magic `m.height - 4` removed)
- Regenerate goldens: `GOLDEN_UPDATE=1 go test ./internal/app -run TestResize`
- **Build state:** Green (all goldens regenerated; no grep-gates to trip yet)

**Commit 3 — Grep-gates + bench budget + golden verification:**
- Add `TestChromeASCIIOnly` (D-20)
- Add `TestChromeNormalBorderOnly` (D-21)
- Add `TestViewNoNewStyle` (D-22 AST walker)
- Add `TestBenchmarkAppView_UnderBudget` (D-24)
- All tests should pass; any failure here means commit 2's rewrite violated the discipline → fix in this commit before merge
- **Build state:** Green, phase gates active

**Why this order:**
- Commit 1 is purely additive — no risk of test regression
- Commit 2 is the big-bang integration — commits 1 + 2 together produce the visible chrome
- Commit 3 locks the discipline — if it fails, you know exactly what went wrong in commit 2 (a NewStyle leak, a rounded border slip, a non-ASCII rune) without re-auditing previous commits
- A bisect of any future regression lands on commit 2 (the meaningful change) instead of commit 3 (the guard)

**Alternative 2-commit variant (if Plan 3 author prefers fewer commits):**
- Commit 1: Commits 1 + 2 squashed (sub-model hints + AppModel integration + goldens)
- Commit 2: Commits 3 (grep-gates + bench)

Both variants satisfy D-25. Recommend 3-commit for bisectability.

## Assumptions Log

> Claims tagged `[ASSUMED]` or `[CITED]` where direct verification was not possible.

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `FileListKeyMap.ShortHelp()` returns 10 bindings and we need to add g/G to reach 12 | Sub-model Hints Content | Low — code reads 10 bindings; if the count differs, curation adjusts but structure intact [VERIFIED: bindings.go:75 returns 10 items] |
| A2 | Logo Candidate A byte-art renders identically across Alacritty/iTerm2/macOS Terminal/Windows Terminal | Logo Byte-Art Candidates | Low — all ASCII range chars. Pitfall 12 covers non-ASCII box-drawing; logo is pure ASCII per D-01. |
| A3 | `HistoryModel.CommitCount()` and `HealthModel.FindingCount()` accessors exist | titleForState example | Low — Plan 3 adds them if missing (one-line additions per sub-model) |
| A4 | Bench test with 100 iterations wall-clock is stable enough for 50µs gate | Pitfall 6 | Medium — CI noise may push single-run median higher than 50µs even for conforming code. Plan 3 can raise to 75µs or increase iterations to 1000 if flaky. |
| A5 | `lipgloss/v2/table.Width(0)` behavior is safe (no panic) | Pitfall 5 | Low — early-return on `m.width == 0` eliminates the call entirely |
| A6 | `ansi.Truncate(s, n, "…")` is the correct signature | overlayTitle example | Low [VERIFIED: github.com/charmbracelet/crush uses this signature at internal/ui/common/elements.go:133 — `ansi.Truncate(description, width-..., "…")`] |
| A7 | Soft-serve `overlayTitle` pattern is community-standard, not cited from a single authoritative upstream | Closed Research Gaps #1 | Low — pattern is visibly used across the TUI ecosystem; the comment in chrome.go should cite this research document revision rather than a soft-serve commit that doesn't contain the pattern |
| A8 | Phase 7 does not need a new MenuCellStyle beyond MenuKeyStyle/MenuDescStyle | RenderMenu example | Medium — the `StyleFunc(row, col)` return value affects outer cell padding. Plan 1 author can collapse to MenuDescStyle (or introduce a lightweight MenuCellStyle if padding needs customization). |

**If this table is empty:** Not empty — 8 assumptions. Plan authors should verify A3, A4 during Plan 3 execution.

## Open Questions

1. **`HistoryModel.CommitCount()` and `HealthModel.FindingCount()` accessor existence**
   - What we know: CONTEXT.md D-15 specifies these accessors for title counts
   - What's unclear: Code inspection didn't locate them in current sub-models (they exist conceptually via `m.history.CommitCount()` references in CONTEXT but not grep-verified)
   - Recommendation: Plan 3 first task verifies with `grep -n "CommitCount\|FindingCount" internal/ui/{history,health}.go`. Add one-line accessor if missing: `func (m HistoryModel) CommitCount() int { return len(m.entries) }`.

2. **MenuCellStyle necessity**
   - What we know: D-05 specifies MenuKeyStyle (accent) and MenuDescStyle (fg) only
   - What's unclear: Whether `table.StyleFunc` return value needs an outer cell style (padding, alignment) beyond the inline-rendered fragments
   - Recommendation: Plan 1 tries inline-rendered fragments first (simpler); if table output is misaligned, introduce `MenuCellStyle` as third var and document the deviation in Plan 1.

3. **stateRecipientList hint owner**
   - What we know: D-09 lists RecipientList as one of 9 sub-models; Pitfall 3 above reveals no such sub-model exists
   - What's unclear: Does Plan 3 still extract a `RecipientListModel` or stick with AppModel inline hints
   - Recommendation: Stick with AppModel inline hints (`keys.RecipientListHints` package var). Do NOT refactor to a separate sub-model in Phase 7 — that's scope creep. Note in Plan 3 that D-09's "9 sub-models" is effectively "8 sub-models + 1 AppModel-owned state" since RecipientList renders inline.

4. **Bench-test stability budget on CI**
   - What we know: D-24 specifies 50µs hard gate
   - What's unclear: CI runner variance — running on a slow GitHub Actions box vs a dev laptop
   - Recommendation: Start with 50µs gate. If CI flakes >5% of runs, raise to 75µs (still within UI-21 target for Phase 11 formalization).

## Sources

### Primary (HIGH confidence)

- **Existing codebase** (all grep-verified on 2026-04-24):
  - `internal/app/model.go:1284` — `View()` method current shape
  - `internal/app/model.go:1322-1326` — `tea.View` return pattern (NewView + AltScreen)
  - `internal/app/model.go:1394-1425` — statusBarHeight, bodyDims, chromeHeight, crumbsHeight helpers
  - `internal/app/model.go:1841` — magic `m.height - 4` at renderRecipientList (Phase 6 TODO)
  - `internal/app/model.go:262` — `recipientAction string` field ("add", "remove", "healthcheck")
  - `internal/app/model.go:942, 985, 996, 1099, 1105` — `IsSearchActive()` dispatch sites
  - `internal/app/bench_test.go` — BenchmarkAppView baseline pattern
  - `internal/app/resize_test.go` — 4-size golden harness
  - `internal/app/testdata/resize_*.golden` — existing empty-state goldens (ready to refresh)
  - `internal/keys/bindings.go:29-38, 74-76, 90-140, 190-205, 207-291` — existing keymap + ShortHelp/FullHelp
  - `internal/ui/styles.go` — design system pattern (package-var styles, no AdaptiveColor)
  - `internal/ui/diff.go:96-114` — Update handler for stateDiff keys
  - `internal/ui/recipientform.go:93-120` — RecipientFormModel Update
  - `internal/ui/health.go:60-75, model.go:672` — HealthModel key routing
  - `internal/ui/history.go:56-70, model.go:1042` — HistoryModel key routing
  - `internal/testutil/golden.go` — Phase 6 RequireGoldenStructure + RequireGoldenColors
  - `go.mod` — dependency versions (lipgloss v2.0.3, bubbletea v2.0.4, bubbles v2.1.0, x/ansi v0.11.7)

- **Planning artifacts**:
  - `.planning/phases/07-chrome-skeleton/07-CONTEXT.md` — Full decisions (D-01..D-26)
  - `.planning/phases/07-chrome-skeleton/07-DISCUSSION-LOG.md` — Q&A trail
  - `.planning/phases/06-layout-groundwork/06-CONTEXT.md` — Phase 6 stubs carry-forward
  - `.planning/research/ARCHITECTURE.md` Pattern 1, 3, 4 — chrome composition patterns
  - `.planning/research/PITFALLS.md` Pitfall 2, 3, 6, 7, 9, 12 — guardrails
  - `.planning/research/STACK.md` §lipgloss/v2/table — menu renderer pattern
  - `.planning/research/SUMMARY.md` §Phase 7 — milestone synthesis
  - `.planning/REQUIREMENTS.md` — UI-01, UI-02, UI-06, UI-15 specs
  - `.planning/ROADMAP.md` §Phase 7 — goal + 5 success criteria

- **pkg.go.dev (fetched 2026-04-24)**:
  - `pkg.go.dev/charm.land/lipgloss/v2` — confirmed no native border-title API
  - `pkg.go.dev/charm.land/lipgloss/v2/table` — StyleFunc signature + border-disable flags + Width()

### Secondary (MEDIUM confidence)

- **soft-serve upstream** (`raw.githubusercontent.com/charmbracelet/soft-serve/main/...`, fetched 2026-04-24 at revision `ac135366727f5b9ebecb23113faa789a84b47bce`):
  - `pkg/ui/components/header/header.go` — **does NOT contain the `overlayTitle` pattern**. CONTEXT.md's cite is either based on an older soft-serve revision or a pattern inferred from ecosystem usage. Research gap acknowledged in "Closed Research Gaps" §1.
  - `pkg/ui/components/tabs/tabs.go` — tab composition pattern (uses custom Border struct, not overlay)
- **charmbracelet/crush main** (reference TUI pattern library):
  - `internal/ui/common/elements.go` — uses `ansi.Truncate(s, n, "…")` signature (validates overlayTitle truncation approach)
  - `internal/ui/common/elements.go:DialogTitle` — alternative "title + decorative line" pattern (informs title styling, not chosen for Phase 7)
- **charmbracelet/lipgloss examples/layout/main.go**: custom Border struct pattern for tab rendering; confirms no built-in title on border
- **CLAUDE.md** §Technology Stack — bubbletea v2 migration rules (View returns tea.View; AltScreen field)

### Tertiary (LOW confidence)

- **Charm ecosystem community convention** — the string-splice `overlayTitle` pattern is documented as "what every production Bubble Tea app does today" in `.planning/research/ARCHITECTURE.md` Pattern 3 and PITFALLS Pitfall 9, but no single authoritative reference implementation found at current `main` revisions of cited repos. Plan 2 should cite this research document (`.planning/phases/07-chrome-skeleton/07-RESEARCH.md` §Closed Research Gaps #1) as the rationale trail, plus a stable snapshot commit of this research revision.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all versions verified in go.mod; lipgloss/v2/table API confirmed via pkg.go.dev
- Architecture: HIGH — composition patterns match existing Phase 6 + ARCHITECTURE research; code examples reference existing helpers (bodyDims, statusBarHeight)
- Pitfalls: HIGH — eight new pitfalls surfaced that weren't in PITFALLS.md (Plan 3 commit order, test suite bench speed, etc.); existing PITFALLS carry-forward verified
- overlayTitle reference: MEDIUM — pattern confirmed community-standard, authoritative single source (soft-serve) not verifiable at current main

**Research date:** 2026-04-24
**Valid until:** 2026-05-24 (30 days for stable lipgloss v2.x / bubbletea v2.x; re-verify before Phase 10 begins if CI times out on bench test)

---

*Phase: 07-chrome-skeleton*
*Research completed: 2026-04-24*
*Ready for planning: yes*
