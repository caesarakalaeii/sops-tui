# Technology Stack — Milestone v1.1 (k9s visual parity)

**Project:** sops-tui
**Milestone:** v1.1 — k9s visual shell (persistent menu, logo, info panel, titled content, breadcrumb chips, theme/skin support)
**Researched:** 2026-04-23
**Confidence:** HIGH

## TL;DR

**No new runtime dependencies are needed.** The existing stack — specifically `charm.land/lipgloss/v2` v2.0.3 and `charm.land/bubbles/v2` v2.1.0 — already ships every primitive required for the v1.1 visual shell. The skin/theme loader reuses `github.com/goccy/go-yaml` v1.19.2 already pulled in for SOPS file parsing.

| Feature | Library / primitive to use | New dep? |
|---------|---------------------------|----------|
| Persistent keybinding menu (top-right) | `charm.land/lipgloss/v2/table` (multi-column grid) | No |
| ASCII logo (top-right) | `lipgloss.NewStyle().Foreground(...).Render(strings.Join(logoLines, "\n"))` | No |
| Header info panel (top-left) | `lipgloss.JoinVertical` of labelled rows, fixed-width label column via `Style.Width()` | No |
| Titled bordered content regions | `lipgloss.Border` + `BorderTop(true)` + manual title embed, or `table.New().Border(...)` for data views | No |
| Breadcrumb chips | `lipgloss.NewStyle().Background(...).Padding(0,1)` per segment, `lipgloss.JoinHorizontal` | No |
| Theme / skin YAML loader | `goccy/go-yaml` v1.19.2 (already present) + in-house struct mapping to `lipgloss.Color` | No |
| Multi-layer layout composition | `lipgloss.JoinHorizontal` / `JoinVertical` (first choice) or `lipgloss.Canvas` + `Layer` (overlays only) | No |

**Bottom line:** This is a pure refactor milestone. We are NOT pulling in `tview`, `tcell`, or any k9s components. We port the *patterns* from k9s (6-line small logo array, multi-column menu layout, `<segment>` crumb pills) into idiomatic Lipgloss v2 code. Everything we need to replicate k9s visuals is already a first-class primitive in Lipgloss v2.

## Recommended Stack (no changes to go.mod)

### Existing — no version bumps required

| Technology | Current (go.mod) | Role in v1.1 | Status |
|------------|-------------------|--------------|--------|
| `charm.land/bubbletea/v2` | v2.0.4 | Program loop unchanged | Keep |
| `charm.land/lipgloss/v2` | v2.0.3 | Borders, colors, layout, `table` subpackage, `Canvas`/`Layer` for overlays | Keep |
| `charm.land/bubbles/v2` | v2.1.0 | `help`, `viewport`, `list`, `textinput` unchanged | Keep |
| `github.com/goccy/go-yaml` | v1.19.2 | Reused for parsing `~/.config/sops-tui/skins/*.yaml` | Keep |

### Explicitly NOT added

| Rejected library | Why rejected |
|------------------|--------------|
| `github.com/derailed/tview` / `tcell` | Framework swap ruled out by milestone scope ("stay on Bubble Tea v2"). tview is retained-mode; Bubble Tea is Elm architecture — mixing them means two runloops and breaks our teatest-based golden snapshot tests. |
| `charm.land/bubbles/v2/table` | Table *component* for interactive data browsing (scrollable, focusable rows). We're building a *static, non-interactive* 6-row menu grid — `lipgloss/v2/table` is the correct tool. The `bubbles` table would carry cursor state we don't want. |
| `charm.land/huh/v2` (for skin picker) | Already in our stack mentally but unused in v1.0. A skin picker is out of scope for v1.1 (user edits YAML file; no in-app picker). Leave huh out until a form-driven flow actually exists. |
| Any "focus manager" library (e.g. `mritd/bubbles` focus) | We only have one interactive pane at a time (menu + logo + info are static, breadcrumb is derived state). No focus rotation is needed — the existing state machine in `internal/app/model.go` already tracks which pane receives key events. |
| `mitchellh/colorstring` / `fatih/color` | Lipgloss is a strict superset. Mixing ANSI libraries breaks the Ultraviolet renderer that Bubble Tea v2 uses. |
| A dedicated skin-theme library (e.g. hypothetical `charmbracelet/skin`) | None exists in the Charm ecosystem as of April 2026 (verified via Context7). Roll our own trimmed schema. |

## Feature-to-Primitive Mapping

### 1. Persistent keybinding menu (top-right, multi-column)

**k9s reference:** `~/git/k9s/internal/ui/menu.go` — a `tview.Table` with `maxRows = 6`, auto-computing `colCount` from `len(hints)/6+1`, formatting each cell as `<mnemonic> description` with color-coded key + description.

**sops-tui replacement:** `charm.land/lipgloss/v2/table`.

```go
import "charm.land/lipgloss/v2/table"

// 6 rows is the k9s convention; columns fan out automatically.
const menuMaxRows = 6

func renderMenu(hints []Hint, width int) string {
    rows := make([][]string, menuMaxRows)
    // ... fill column-major (like k9s buildMenuTable) ...
    t := table.New().
        Border(lipgloss.HiddenBorder()).   // no border — menu is chromeless
        BorderTop(false).BorderBottom(false).
        StyleFunc(func(row, col int) lipgloss.Style {
            // even col = mnemonic (accent), odd col = description (fg)
            if col%2 == 0 { return ui.MenuKeyStyle }
            return ui.MenuDescStyle
        }).
        Rows(rows...).
        Width(width)
    return t.String()
}
```

**Why `lipgloss/v2/table` over hand-rolled `JoinHorizontal`:**
- Handles column width equalization automatically (k9s `maxKeys[col]` padding logic falls out for free).
- `StyleFunc(row, col)` matches k9s's mnemonic-vs-description color pattern exactly.
- Static, stateless render — no cursor, no scrolling, no `tea.Msg` handling — perfect for a chrome element.

**Why NOT `bubbles/v2/table`:** That one holds `Cursor`, `Focused`, selection state, and emits key messages. Complete overkill for a hint grid.

### 2. ASCII logo (top-right, 6 lines, state-colored)

**k9s reference:** `~/git/k9s/internal/ui/logo.go` holds `LogoSmall []string` (6 lines), renders into a `tview.TextView` with dynamic foreground color set via `Err()`/`Warn()`/`Info()` methods.

**sops-tui replacement:** A pure-function renderer + three state-keyed style variables.

```go
// internal/ui/logo.go (new file)
var LogoSmall = []string{
    `  ____                 _____ _   _ ___ `,
    ` / ___|  ___  _ __  __|_   _| | | |_ _|`,
    ` \___ \ / _ \| '_ \/ __|| | | | | || | `,
    `  ___) | (_) | |_) \__ \| | | |_| || | `,
    ` |____/ \___/| .__/|___/|_|  \___/|___|`,
    `             |_|                       `,
}

type LogoState int
const (LogoInfo LogoState = iota; LogoWarn; LogoErr)

func RenderLogo(state LogoState) string {
    color := ColorAccent
    switch state {
    case LogoWarn: color = ColorWarning
    case LogoErr:  color = ColorError
    }
    return lipgloss.NewStyle().Foreground(color).Bold(true).
        Render(strings.Join(LogoSmall, "\n"))
}
```

**No library needed.** This is 15 lines of code. The `sync.Mutex` in k9s's Logo is irrelevant — Bubble Tea's single-threaded update loop means we never need to guard logo state. State change is a `tea.Msg` → `Update` → new render.

**Logo width note:** The k9s small logo is 26 columns wide (see `header.AddItem(a.Logo(), 26, 1, false)` in `app.go:327`). Our sops-tui logo should stay in the 24-30 column band so the menu has room in normal terminal widths (≥120 cols).

### 3. Header info panel (top-left, ClusterInfo analog)

**k9s reference:** `~/git/k9s/internal/view/cluster_info.go` — a `tview.Table` of label/value pairs. `buildHeader` in `internal/view/app.go:305` allocates `clusterInfoWidth` columns for it.

**sops-tui replacement:** `lipgloss.JoinVertical` of rows, each row = `JoinHorizontal(labelStyle.Width(labelCol).Render("Key:"), valueStyle.Render(val))`.

```go
// Rows we display (per milestone spec):
//   SOPS config  /abs/path/.sops.yaml
//   Age key      AGE-SECRET-KEY-...A1B2 (fingerprint)
//   Recipients   3 active
//   Git          feature/v1-1 (clean) | (dirty) | (no repo)
//   Files        12 encrypted
```

**No library needed.** This is the bread-and-butter Lipgloss use case. The existing `ui.DimText` (for labels) and `ui.ColorFg` (for values) cover styling. Implementation is ~80 lines of Go.

**Sizing:** k9s allocates `clusterInfoWidth = 30` cells by default. We'll target ~38 cells (longer labels: ".sops.yaml" vs "Context:"), leaving `terminal_width - 38 - 26 (logo) - padding ≈ 50+ cells` for the menu. At an 80-col min terminal this is tight but workable; fall back to logo-hidden at <100 cols.

### 4. Titled bordered content regions

**k9s reference:** `tview.Box` has built-in title support (`SetTitle`, `SetTitleAlign`). In sops-tui this currently doesn't exist — the content area is an unbordered raw render.

**sops-tui replacement:** Lipgloss v2 borders + manual title overlay. Lipgloss's `Border()` style does not embed a title natively (confirmed: no `Title()` method exists on `Style`). The standard pattern is:

```go
func titledBox(title, content string, width, height int) string {
    body := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(ColorMuted).
        Padding(0, 1).
        Width(width - 2).
        Height(height - 2).
        Render(content)

    // Overlay the title onto the top border line.
    // The `» Title «` format is a k9s-style brand element.
    titleText := TitleStyle.Render(" " + title + " ")
    return overlayTitleOnBorder(body, titleText) // trivial string splice
}
```

**No library needed.** `overlayTitleOnBorder` is ~20 lines: split by newline, replace columns `[2:2+len(titleText)]` of the first line with the title, rejoin. This pattern is documented in the Lipgloss README's "advanced" section.

**Alternative that we should avoid:** Using `lipgloss.Canvas` + `Layer` z-stacking for *every* titled box. Canvas is fantastic for modal overlays (help popup, confirm dialog — actually a great candidate for upgrading the existing `help.go` overlay) but overkill for inline frame titles. Use cost-effective string splicing for titles.

### 5. Breadcrumb chips

**k9s reference:** `~/git/k9s/internal/ui/crumbs.go` — each segment rendered as ``[fg:bgColor:b] <segment> [-:bgBody:-]`` (tview color syntax). Active segment uses `ActiveColor` background; others use `BgColor` background. Trailing space between chips.

**sops-tui replacement:** Drop-in replacement for the current `renderBreadcrumb` in `internal/ui/statusbar.go:179`. Each chip is a single Lipgloss `Style.Background(...).Foreground(...).Padding(0,1)` render, joined by a single-space separator.

```go
// internal/ui/crumbs.go (new file)
var (
    ChipStyle = lipgloss.NewStyle().
        Foreground(ColorFg).
        Background(ColorSurface).
        Bold(true).
        Padding(0, 1)   // <space>text<space>
    ChipActiveStyle = ChipStyle.Background(ColorAccent).Foreground(ColorBg)
)

func RenderCrumbs(segments []string) string {
    chips := make([]string, len(segments))
    for i, seg := range segments {
        s := strings.ReplaceAll(strings.ToLower(seg), " ", "")
        if i == len(segments)-1 {
            chips[i] = ChipActiveStyle.Render("<" + s + ">")
        } else {
            chips[i] = ChipStyle.Render("<" + s + ">")
        }
    }
    return strings.Join(chips, " ")
}
```

**No library needed.** This replaces the current `" > "` separator logic with a pill-renderer. Integration point: `internal/ui/statusbar.go:179` `renderBreadcrumb` function body — gut it and call `RenderCrumbs` on the `segments` slice.

**Migration note:** The breadcrumb currently lives inside the 1-line status bar. In v1.1 this is promoted to its own row between header and content (k9s puts crumbs *below* the main body — see `view/app.go:289` `toggleCrumbs`; crumbs are toggled at flex index 2, after header (0) and body (1) but let's put them above-body for clarity matching the existing UI-SPEC position). Either location works; both are a single `JoinVertical` call site.

### 6. Theme / skin YAML loader

**k9s reference:** `~/git/k9s/internal/config/styles.go` (815 lines) + YAMLs at `~/git/k9s/skins/*.yaml` (dracula, nord, gruvbox, etc.). Full schema covers body/prompt/info/dialog/frame/views/charts/xray/yaml/logs — way more surface than sops-tui has.

**sops-tui replacement:** A *trimmed* skin schema, hand-written struct, parsed with the already-imported `goccy/go-yaml`.

```go
// internal/ui/skin.go (new file)
type Skin struct {
    Body   SkinBody   `yaml:"body"`
    Frame  SkinFrame  `yaml:"frame"`
    Info   SkinInfo   `yaml:"info"`
}
type SkinBody struct {
    FgColor  string `yaml:"fgColor"`
    BgColor  string `yaml:"bgColor"`
    LogoColor string `yaml:"logoColor"`
}
type SkinFrame struct {
    Border SkinBorder `yaml:"border"`
    Menu   SkinMenu   `yaml:"menu"`
    Crumbs SkinCrumbs `yaml:"crumbs"`
    Title  SkinTitle  `yaml:"title"`
    Status SkinStatus `yaml:"status"`
}
// ... ~10 nested structs total, maybe 80 lines ...

func LoadSkin(path string) (*Skin, error) {
    b, err := os.ReadFile(path)
    if err != nil { return nil, err }
    var s Skin
    if err := yaml.Unmarshal(b, &s); err != nil { return nil, err }
    return &s, nil
}

// ApplySkin rewrites the package-level style vars in internal/ui/styles.go.
// Called once at startup after reading ~/.config/sops-tui/skin.yaml (if present).
func ApplySkin(s *Skin) { ... }
```

**Why this is better than importing a theming library:**
- There is no maintained Go library that implements "YAML-to-Lipgloss-styles mapping." (Verified via Context7 + WebSearch on 2026-04-23.) Every project in the space rolls this ourselves.
- The k9s schema is over-engineered for our needs. k9s has 30+ distinct style slots; sops-tui has ~8 (body, border, title, menu, crumbs, status/error/success accents). Keeping our schema small means we can ship the skin loader in <150 LOC.
- Already-imported `goccy/go-yaml` handles all the parsing, including anchor/alias YAML refs (`&foreground` / `*foreground`) that k9s skins rely on. No new dependency.

**Skin schema compatibility note:** We should deliberately choose a subset of k9s's field names so power users can reuse their existing k9s skin YAML files (at minimum: `body.fgColor`, `body.bgColor`, `body.logoColor`, `frame.border.fgColor`, `frame.menu.fgColor`, `frame.menu.keyColor`, `frame.crumbs.fgColor`, `frame.crumbs.bgColor`, `frame.crumbs.activeColor`, `frame.title.fgColor`, `frame.title.bgColor`). Extra k9s keys are silently ignored by `goccy/go-yaml`. This is a differentiator worth ~30 lines of design discipline.

## Bonus: Candidate upgrade for `help.go`

The current `internal/ui/help.go` renders a full-screen bordered help box. With Lipgloss v2 now shipping `Canvas` + `Layer`, we could promote the help overlay to a *centred modal* that doesn't clobber the underlying content — a polished k9s-ish behaviour. This is **out of scope for v1.1** per the milestone boundary (pure visual shell rework, no behaviour change), but worth flagging for v1.2.

## Alternatives Considered (per feature)

| Feature | Recommended | Alternative | Why alternative rejected |
|---------|-------------|-------------|--------------------------|
| Menu grid | `lipgloss/v2/table` | Hand-rolled `JoinHorizontal(JoinVertical(...))` | Column-equalization logic rediscovers `lipgloss/v2/table`'s width algorithm. Using the library gets `StyleFunc(row,col)` for free and tests smaller. |
| Menu grid | `lipgloss/v2/table` | `bubbles/v2/table` | Interactive table with cursor/focus — we want a static chrome element, not a selectable widget. |
| Logo | Plain string render | ASCII-art library (e.g. `common-nighthawk/go-figure`) | Figlet-style dynamic text is the opposite of what we want — our logo is a fixed 6-line hand-designed banner. Library adds weight for zero benefit. |
| Titled borders | Manual title overlay on `Border()` | `lipgloss.Canvas` + `Layer` for every frame | Canvas/Layer is for z-stacked overlays (modals, dropdowns). Inline titles are 20 lines of string splicing. |
| Breadcrumb chips | Inline `Style.Background().Padding()` | `lipgloss/v2/list` subpackage | List renders vertically with bullets/indent. Our crumbs are horizontal chips — mismatched primitive. |
| Skin loader | `goccy/go-yaml` + hand struct | External theme library | None exists in Charm ecosystem. Rolling our own is ~150 LOC and keeps the schema tight. |
| Skin loader | `goccy/go-yaml` (existing) | `gopkg.in/yaml.v3` | Already ruled out project-wide per CLAUDE.md stack decisions — goccy passes more spec cases and yaml.v3's author recommends migration. |

## Installation

**No `go get` commands.** The go.mod file stays unchanged. All new code imports the packages already resolved:

```go
import (
    "charm.land/lipgloss/v2"
    "charm.land/lipgloss/v2/table"      // new to sops-tui in v1.1
    "github.com/goccy/go-yaml"           // already in use for SOPS parsing
)
```

Run `go mod tidy` after adding new files to confirm no transitive churn. Expected result: zero changes to go.sum.

## Integration Points with Existing Code

| Existing file | v1.1 change |
|---------------|------------|
| `internal/ui/styles.go` | Add ~8 new style vars: `MenuKeyStyle`, `MenuDescStyle`, `LogoInfoStyle`/`Warn`/`Err`, `ChipStyle`, `ChipActiveStyle`, `TitleStyle`, `InfoLabelStyle`. Keep hex constants as single source of truth so skin loader can override. |
| `internal/ui/statusbar.go:179` `renderBreadcrumb` | Replace body with call to new `RenderCrumbs` from `internal/ui/crumbs.go`. Remove the `" > "` + `BreadcrumbSep`/`BreadcrumbActive` logic. |
| `internal/app/model.go` `View()` around line 1329 | Compose header via new helper: `header := lipgloss.JoinHorizontal(lipgloss.Top, infoPanel, menu, logo)`. Body becomes `titledBox(title, body, w, bodyH)`. Crumbs row becomes `RenderCrumbs(segments)`. |
| `internal/ui/help.go` | Unchanged in v1.1 (still full-screen `?` overlay). Flagged for v1.2 upgrade. |
| `internal/ui/crumbs.go` (new) | Contains `RenderCrumbs` + chip styles. |
| `internal/ui/logo.go` (new) | Contains `LogoSmall` + `RenderLogo(state)`. |
| `internal/ui/menu.go` (new) | Contains `Hint` struct + `RenderMenu([]Hint, width)` using `lipgloss/v2/table`. |
| `internal/ui/info.go` (new) | Contains `InfoPanel` struct (fields: `SopsConfigPath`, `AgeFingerprint`, `RecipientCount`, `GitBranch`, `GitDirty`, `FileCount`) + `RenderInfo(p InfoPanel, width int)`. |
| `internal/ui/skin.go` (new) | Contains `Skin` struct + `LoadSkin` + `ApplySkin`. |
| `.planning/` milestone docs | UI-SPEC update describing new chrome layout. |

## Version Pin Notes

- **No version bumps**: v1.1 ships on the exact same go.mod as v1.0. The new `lipgloss/v2/table` subpackage is already available at v2.0.3 (verified: `ls ~/go/pkg/mod/charm.land/lipgloss/v2@v2.0.3/table/` shows the package).
- **CGO_ENABLED=0 still applies**: None of the new code paths need CGo. Lipgloss v2 is pure Go.
- **No security-sensitive deps touched**: `filippo.io/age` and `go-git/go-git/v5` are not involved in UI work.

## Sources

- Lipgloss v2 table subpackage confirmed local at `~/go/pkg/mod/charm.land/lipgloss/v2@v2.0.3/table/table.go`
- Lipgloss v2 Canvas + Layer confirmed local at `~/go/pkg/mod/charm.land/lipgloss/v2@v2.0.3/{canvas.go,layer.go}`
- Context7 `/charmbracelet/lipgloss` (benchmark score 89.87) — confirmed `table.New().StyleFunc(row, col)` API, `HeaderRow` const, rounded/markdown/hidden border variants, `Rows(...)` builder
- Context7 `/charmbracelet/bubbles` (benchmark score 80.92) — confirmed no skin/theme loader component exists in the v2 bubbles catalog (components listed: spinner, textinput, textarea, table, progress, paginator, viewport, list)
- k9s reference code read locally: `internal/ui/{menu.go, logo.go, splash.go, crumbs.go}`, `internal/view/app.go:305` `buildHeader`
- k9s skin YAML schema surveyed: `~/git/k9s/skins/dracula.yaml` (114 lines) and `~/git/k9s/internal/config/styles.go` (815 lines) — confirmed we want a trimmed-down subset (~10 of 30+ style slots)
- Existing stack decisions: `/home/moersener/git/sops-tui/CLAUDE.md` (STACK section); `go.mod` pins at `charm.land/lipgloss/v2 v2.0.3`, `charm.land/bubbles/v2 v2.1.0`, `goccy/go-yaml v1.19.2`
