# Phase 10: Theming + Accessibility — Pattern Map

**Mapped:** 2026-05-04
**Files analyzed:** 11 modified + ~16 new test goldens (no new source files; planner may optionally factor `palette.go` per CONTEXT Claude's-discretion)
**Analogs found:** 11 / 11 (every modified file has a same-package, same-role precedent)

---

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/ui/styles.go` (3 hex flips, 8 ANSI vars, `Palette` struct + `Palette(profile)` accessor, 2 fallback chip styles) | design-system / package-vars | constants + accessor function | self (`internal/ui/styles.go:37-54` 24-bit `Color*` block + lines 75-391 named-style `var (...)` block) | exact (additive) |
| `internal/ui/statusbar.go` (`FlashSeverity` enum, `flashSeverity` field, `FlashInfo`/`FlashWarn`/`FlashErr` methods, `FlashSeverity()` accessor, severity-aware `View()` flash branch) | UI primitive (sub-model) | event-driven (Flash → tea.Tick → FlashClearMsg) | self (`internal/ui/statusbar.go:53-59`/`125-151`/`160-180`) | exact |
| `internal/ui/logo.go` (no signature change unless planner adds `width`-trimming) | pure-function renderer | switch-on-enum | self (`internal/ui/logo.go:48-63` `RenderLogo(LogoStatus, width) string`) | exact (canonical) |
| `internal/ui/chrome.go` (`RenderChrome` gains `palette Palette` parameter; profile-aware tier path) | pure-function renderer / composer | width-tier switch + render-time variant select | self (`internal/ui/chrome.go:106-137` 3-tier width fallback) | exact |
| `internal/ui/crumbs.go` (`RenderCrumbs` gains `palette Palette`; bracket-fallback chip rendering on `Profile <= ANSI`) | pure-function renderer | iter-over-segments + truncate-to-width | self (`internal/ui/crumbs.go:37-60`) | exact |
| `internal/ui/menu.go` (`RenderMenu` gains `palette Palette`; profile-aware key/desc style) | pure-function renderer | column-major fill | self (`internal/ui/menu.go:46-113`) | exact |
| `internal/ui/infopanel.go` (`RenderInfoPanel` gains `palette Palette`; profile-aware label/value style) | pure-function renderer | row-builder | self (`internal/ui/infopanel.go:34-50`) | exact |
| `cmd/sops-tui/main.go` (add `colorprofile.Detect()` call between Step 5 + Step 6; pass into `NewAppModel`) | process bootstrap | linear startup flow | self (`cmd/sops-tui/main.go:28-76`) | exact |
| `internal/app/model.go` (`profile colorprofile.Profile` field; `NewAppModel` signature change; `resolveLogoState() ui.LogoStatus` method; 41-callsite flash re-classification; `RenderChrome`/`RenderCrumbs` callsite signature cascade) | App orchestration | aggregator | self (`internal/app/model.go:222-305` `AppModel` struct + `NewAppModel`; 1395-1408 chrome composition; 1602-1621 chrome/crumbs height helpers) | exact |
| `internal/app/chrome_test.go` + `resize_test.go` (4-profile golden matrix; 60×24 + 100×30 widths) | golden-matrix tests | matrix-driver | `internal/app/menuhints_drift_test.go:171-208` `TestMenuGolden`'s 13-state run-table; `internal/app/resize_test.go` 4-width matrix | exact (composition: cross-product width × profile) |
| `internal/ui/{statusbar,crumbs}_test.go` extensions (severity bg-tint + `[W]`/`[E]` prefix + bracket-fallback chip) | unit tests | render-and-assert | self (`internal/ui/statusbar_test.go:60-98`; `internal/ui/crumbs_test.go:22-67`) | exact |

---

## Pattern Assignments

### Pattern 1 — `internal/ui/styles.go` palette additions, ANSI variants, `Palette` struct + accessor, fallback chip styles

**Analog:** `internal/ui/styles.go` itself (additive — every Phase since 1 has appended to the package-vars block)

**Imports + hex-constant block** (lines 8-32, NEW Phase 10 entries slot after `ColorFgHex`):
```go
package ui

import (
	"charm.land/lipgloss/v2"
)

// Color hex values per 01-UI-SPEC.md §Color. These constants hold the raw hex strings
// so they can be verified in tests and used to construct lipgloss colors.
const (
	// ColorBgHex is the terminal background fill color (60% usage).
	ColorBgHex = "#1e1e2e"
	// ColorSurfaceHex is used for status bar, help overlay, and error box backgrounds (30% usage).
	ColorSurfaceHex = "#313244"
	// ColorAccentHex is reserved for selection highlights, active breadcrumbs, tree indicators,
	// and help key labels (10% usage).
	ColorAccentHex = "#89b4fa"      // Phase 10 D-415: flip to "#cba6f7" (Catppuccin Mauve)
	// ColorWarningHex indicates soft warnings (age key missing, .sops.yaml missing).
	ColorWarningHex = "#f9e2af"     // Phase 10 D-415: flip to "#fab387" (Catppuccin Peach)
	// ColorErrorHex indicates hard errors (sops binary missing, error box borders).
	ColorErrorHex = "#f38ba8"       // Phase 10 D-415: flip to "#eba0ac" (Catppuccin Maroon)
)
```

**24-bit `Color*` var block** (lines 36-54) — Phase 10 leaves these unchanged; the hex-constant flip propagates automatically:
```go
var (
	ColorAccent  = lipgloss.Color(ColorAccentHex)
	ColorWarning = lipgloss.Color(ColorWarningHex)
	ColorError   = lipgloss.Color(ColorErrorHex)
	// ... bg, surface, success, muted, fg unchanged
)
```

**Pattern for the NEW `Color*ANSI` parallel block** — mirrors the existing `Color*` block exactly, single declaration per variant, doc-comment naming the index mnemonic. Use `lipgloss.ANSIColor(N)` not `lipgloss.Color(N)`:
```go
// Phase 10 (D-420): 16-color fallback variants for Profile <= colorprofile.ANSI.
// Each index hand-verified in Pitfall 5 §4 — every chrome bg/fg pair is distinct under 4-bit.
var (
	// ColorAccentANSI is bright magenta (13) — distinct from Warning yellow + Error red.
	ColorAccentANSI = lipgloss.ANSIColor(13)
	// ColorBgANSI is black (0).
	ColorBgANSI = lipgloss.ANSIColor(0)
	// ColorSurfaceANSI is bright black / dark grey (8).
	ColorSurfaceANSI = lipgloss.ANSIColor(8)
	// ColorFgANSI is bright white (15).
	ColorFgANSI = lipgloss.ANSIColor(15)
	// ColorMutedANSI is white / light grey (7).
	ColorMutedANSI = lipgloss.ANSIColor(7)
	// ColorSuccessANSI is bright green (10).
	ColorSuccessANSI = lipgloss.ANSIColor(10)
	// ColorWarningANSI is bright yellow (11).
	ColorWarningANSI = lipgloss.ANSIColor(11)
	// ColorErrorANSI is bright red (9).
	ColorErrorANSI = lipgloss.ANSIColor(9)
)
```

**Pattern for the NEW `Palette` struct + accessor** — recommended location is top of `styles.go` immediately after the imports + hex constants (CONTEXT.md "Claude's Discretion"); flat shape mirrors the `Color*` var naming:
```go
// Phase 10 (D-421): Palette is the resolved color set passed into chrome
// renderers. RenderChrome / RenderCrumbs / RenderMenu / RenderInfoPanel /
// RenderLogo accept a Palette parameter and consult its fields rather than
// importing colorprofile directly.
type Palette struct {
	Accent  lipgloss.Color
	Bg      lipgloss.Color
	Surface lipgloss.Color
	Fg      lipgloss.Color
	Muted   lipgloss.Color
	Success lipgloss.Color
	Warning lipgloss.Color
	Error   lipgloss.Color
}

// PaletteFor returns the resolved Palette for the requested profile.
// Profile <= colorprofile.ANSI returns the 16-color fallback variants;
// otherwise returns the 24-bit Catppuccin Mocha variants.
func PaletteFor(profile colorprofile.Profile) Palette {
	if profile <= colorprofile.ANSI {
		return Palette{
			Accent: ColorAccentANSI, Bg: ColorBgANSI, Surface: ColorSurfaceANSI,
			Fg: ColorFgANSI, Muted: ColorMutedANSI, Success: ColorSuccessANSI,
			Warning: ColorWarningANSI, Error: ColorErrorANSI,
		}
	}
	return Palette{
		Accent: ColorAccent, Bg: ColorBg, Surface: ColorSurface,
		Fg: ColorFg, Muted: ColorMuted, Success: ColorSuccess,
		Warning: ColorWarning, Error: ColorError,
	}
}
```

**Pattern for the NEW `CrumbChipFallbackStyle` + `CrumbChipActiveFallbackStyle`** — slot into the existing Phase 8 chip-style block (lines 361-390). Direct analog is `CrumbChipStyle` / `CrumbChipActiveStyle`:
```go
// Phase 8 D-206 — kept unchanged for Profile > ANSI:
CrumbChipStyle = lipgloss.NewStyle().Background(ColorSurface).Foreground(ColorFg)
CrumbChipActiveStyle = lipgloss.NewStyle().
	Background(ColorAccent).
	Foreground(ColorBg).
	Bold(true)

// Phase 10 D-422 — bracket-fallback variants for Profile <= ANSI:
// No bg fill — paired bg/fg collapse on 16-color terminals; underline + bold
// SGR codes (4 + 1) survive every downsample including monochrome.
CrumbChipFallbackStyle = lipgloss.NewStyle().Foreground(ColorFgANSI)
CrumbChipActiveFallbackStyle = lipgloss.NewStyle().
	Underline(true).
	Bold(true)
```

**What to copy:**
- The doc-comment naming convention (every var has a doc comment naming the design-system role + UI-SPEC reference)
- The `lipgloss.NewStyle().X(...).Y(...)` chain layout — one method per line, terminating period before the line break
- The "Phase N: <feature> styles (D-XYZ)" group-header comment introducing each block of new vars (Phase 7.1 D-110, Phase 8 D-201..D-208 each use this convention)
- The `Color*Hex = "#..."` constant naming (Hex suffix) + paired `Color* = lipgloss.Color(*Hex)` derived var

**Deviations the planner should know:**
- Phase 10 does NOT introduce per-component Palette grouping (e.g. `Palette.Chip.ActiveBg`); CONTEXT recommends flat to mirror existing var naming. Planner may choose grouped if it justifies the extra type surface, but flat aligns with the rest of the file.
- The `Palette` struct uses `lipgloss.Color` field type. Both `lipgloss.Color("...")` and `lipgloss.ANSIColor(N)` return concrete `lipgloss.Color`-compatible values — verify the actual return type when planning (`ANSIColor` returns `ansi.IndexedColor` which implements `color.Color`; CONTEXT.md uses `lipgloss.Color` field type which is the union — confirm in planning by reading `lipgloss/v2/color.go`).
- The `Palette(profile)` accessor name in CONTEXT D-421 collides with the `Palette` struct type. Recommend `PaletteFor(profile)` (verb form). Plan author picks; the convention "func that returns a struct named X" is generally `XFor(...)` or `NewX(...)` — `NewStatusBarModel`, `NewAppModel` use `New*`. `PaletteFor` is closer to `RenderX` in shape. Plan 2 author picks.
- `colorprofile` import: the accessor lives in `styles.go`, so this file gains a new direct import on `github.com/charmbracelet/colorprofile`. Verify the package is available — `go.mod` already pulls it in transitively per RESEARCH.md.

---

### Pattern 2 — `internal/ui/statusbar.go` `FlashSeverity` enum + typed methods + severity-aware `View()` flash branch

**Analog:** `internal/ui/logo.go:21-33` `LogoStatus` enum (canonical iota-enum-with-doc pattern) + `internal/ui/statusbar.go:125-151` existing `Flash()` + `Update()` (generation-counter semantics)

**Imports** (lines 19-25):
```go
import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)
```

**`LogoStatus` enum pattern** — verbatim shape for Phase 10's `FlashSeverity` enum (`internal/ui/logo.go:21-33`):
```go
// LogoStatus drives logo recoloring by aggregate app severity. Phase 7
// uses LogoInfo unconditionally; Phase 10 (UI-03) derives the value from
// env checks, flash severity, and aggregate health.
type LogoStatus int

const (
	// LogoInfo renders the logo in ColorAccent (default).
	LogoInfo LogoStatus = iota
	// LogoWarn renders the logo in ColorWarning (Phase 10 wiring).
	LogoWarn
	// LogoError renders the logo in ColorError (Phase 10 wiring).
	LogoError
)
```

**Pattern for the NEW `FlashSeverity` enum** — copy the LogoStatus shape verbatim, swap names + zero-value (Info = 0 per D-409):
```go
// FlashSeverity classifies the active flash message. Phase 10 D-411 +
// D-412 use the enum to pick severity-tinted bg + redundant prefix at
// View() render time. FlashSevInfo is the zero value so a freshly-
// constructed StatusBarModel with no flash fired is safe.
type FlashSeverity int

const (
	// FlashSevInfo is the default neutral severity. Renders unprefixed.
	FlashSevInfo FlashSeverity = iota
	// FlashSevWarn renders with [W] prefix + ColorWarning bg + ColorBg fg.
	FlashSevWarn
	// FlashSevErr renders with [E] prefix + ColorError bg + ColorBg fg.
	FlashSevErr
)
```

**Existing `Flash()` generation-counter semantics** (`internal/ui/statusbar.go:129-137`) — Phase 10's three new methods MUST preserve the `flashGen++` semantics (Pitfall 6 mitigation):
```go
// Flash sets a flash message and schedules a clear after 2 seconds.
// The generation counter is incremented so that only the matching FlashClearMsg
// will clear this flash — earlier ticks are silently ignored.
// Returns the updated model and the tea.Tick command.
func (m StatusBarModel) Flash(msg string) (StatusBarModel, tea.Cmd) {
	m.flashGen++
	gen := m.flashGen
	m.flash = msg
	cmd := tea.Tick(2*time.Second, func(_ time.Time) tea.Msg {
		return FlashClearMsg{Gen: gen}
	})
	return m, cmd
}
```

**Pattern for the NEW `FlashInfo` / `FlashWarn` / `FlashErr` methods** — each is a thin wrapper that sets `m.flashSeverity` then delegates the gen-counter mechanics. Recommended factoring (preserves DRY, single tick-cmd allocation):
```go
// flash is the internal helper shared by Flash / FlashInfo / FlashWarn /
// FlashErr. It increments flashGen, sets msg + severity, and returns the
// tea.Tick clear cmd. Pattern mirrors existing Flash() with one extra field.
func (m StatusBarModel) flash(msg string, sev FlashSeverity) (StatusBarModel, tea.Cmd) {
	m.flashGen++
	gen := m.flashGen
	m.flash = msg
	m.flashSeverity = sev
	cmd := tea.Tick(2*time.Second, func(_ time.Time) tea.Msg {
		return FlashClearMsg{Gen: gen}
	})
	return m, cmd
}

// FlashInfo sets a neutral flash message (no prefix, no bg tint).
func (m StatusBarModel) FlashInfo(msg string) (StatusBarModel, tea.Cmd) {
	return m.flash(msg, FlashSevInfo)
}

// FlashWarn sets a warn-severity flash (renders with [W] prefix + warning bg + dark fg).
func (m StatusBarModel) FlashWarn(msg string) (StatusBarModel, tea.Cmd) {
	return m.flash(msg, FlashSevWarn)
}

// FlashErr sets an error-severity flash (renders with [E] prefix + error bg + dark fg).
func (m StatusBarModel) FlashErr(msg string) (StatusBarModel, tea.Cmd) {
	return m.flash(msg, FlashSevErr)
}

// Flash is the legacy info-severity entrypoint preserved for backward compat
// with the 22 neutral call-sites that don't move during Phase 10's typed-API
// migration (D-407). Equivalent to FlashInfo.
func (m StatusBarModel) Flash(msg string) (StatusBarModel, tea.Cmd) {
	return m.flash(msg, FlashSevInfo)
}
```

**Existing `Update()` ack pattern** (`internal/ui/statusbar.go:142-151`) — Phase 10 must clear the severity field on the same ack:
```go
func (m StatusBarModel) Update(msg tea.Msg) (StatusBarModel, tea.Cmd) {
	if clearMsg, ok := msg.(FlashClearMsg); ok {
		// Only clear if the generation matches — stale ticks are ignored (Pitfall 6).
		if clearMsg.Gen == m.flashGen {
			m.flash = ""
			m.flashSeverity = FlashSevInfo  // Phase 10: drop back to baseline.
		}
		return m, nil
	}
	return m, nil
}
```

**Existing `View()` flash branch** (`internal/ui/statusbar.go:160-167`):
```go
func (m StatusBarModel) View(width int) string {
	if m.flash != "" {
		return StatusBarStyle.
			Width(width).
			Align(lipgloss.Center).
			Render(m.flash)
	}
	// ... rest unchanged
}
```

**Pattern for the EXTENDED severity-aware `View()` flash branch** — branch on `m.flashSeverity`, prepend prefix at render time (D-411), pick bg-tinted style on Warn/Err (D-412). Two new package vars in `styles.go` for the tinted styles:
```go
// In styles.go (new vars next to StatusBarStyle):
FlashWarnBarStyle = lipgloss.NewStyle().Background(ColorWarning).Foreground(ColorBg)
FlashErrBarStyle  = lipgloss.NewStyle().Background(ColorError).Foreground(ColorBg)

// In statusbar.go View():
if m.flash != "" {
	style := StatusBarStyle
	text := m.flash
	switch m.flashSeverity {
	case FlashSevWarn:
		style = FlashWarnBarStyle
		text = "[W] " + m.flash
	case FlashSevErr:
		style = FlashErrBarStyle
		text = "[E] " + m.flash
	}
	return style.Width(width).Align(lipgloss.Center).Render(text)
}
```

**Pattern for the NEW `FlashSeverity()` accessor** — mirror existing `Env()` and `IsClipboardHot()` accessors (`internal/ui/statusbar.go:107`/`121-123`):
```go
// Existing pattern:
func (m StatusBarModel) Env() EnvStatus { return m.env }

// Phase 10 NEW:
// FlashSeverity returns the severity of the active flash. When no flash
// is active (m.flash == ""), returns FlashSevInfo (zero value) so the
// classifier short-circuits to env-derived severity.
func (m StatusBarModel) FlashSeverity() FlashSeverity {
	if m.flash == "" {
		return FlashSevInfo
	}
	return m.flashSeverity
}
```

**What to copy:**
- iota-enum-with-doc-comment-per-const shape from `LogoStatus` (each const has its own line + doc comment)
- value-receiver methods returning updated copies (`StatusBarModel` is a value type per package doc)
- the `flashGen++` increment-then-capture-then-tick sequence (preserves Pitfall 6 stale-clear suppression across all four entrypoints)
- doc-comment cross-references to D-numbers (D-411, D-412 etc.)
- internal-helper pattern with lowercase name (`flash`) for shared mechanics — same shape as how `lipgloss/v2` itself often factors private helpers

**Deviations the planner should know:**
- The internal `flash` helper collides with the `flash string` field name. Pick `setFlash` or unexport via a different name to avoid shadowing. Plan 1 author picks.
- `FlashSeverity` is both a type name and an accessor method name. Go allows this (type vs method namespaces are separate), but readers may find `m.FlashSeverity()` returning a `FlashSeverity` slightly verbose. Acceptable per CONTEXT D-410; plan author may elect to rename either side.
- `renderEnvIndicators` (`statusbar.go:185-222`) currently uses inline `lipgloss.NewStyle()` calls — Phase 10 does NOT need to lift those because TestSubmodelViewsNoNewStyle's BFS scope check on `View()` does not currently traverse this helper (it's reached but the test was carved Phase 8-era). If Plan 1 introduces `lipgloss.NewStyle()` inside `View()` directly (e.g. building `FlashWarnBarStyle` inline), that WILL fail the BFS walker — keep all severity-tinted styles as package vars.

---

### Pattern 3 — `internal/ui/logo.go` `RenderLogo` (canonical pure-function-renderer-with-switch-on-enum)

**Analog:** `internal/ui/logo.go:48-63` itself — Phase 10 may add `palette Palette` parameter (planner discretion per CONTEXT.md), but the body shape is preserved.

**Existing canonical pattern** (`internal/ui/logo.go:48-63`):
```go
// RenderLogo returns the 6-row logo string rendered in the style
// corresponding to the requested severity. The width parameter is
// plumbed for Phase 10 width-responsive layouts; Phase 7 ignores it
// (the art is locked at ~25 cols per D-01).
func RenderLogo(status LogoStatus, width int) string {
	_ = width // reserved for Phase 10 width-responsive logic (D-02)
	art := strings.Join(LogoSmall, "\n")
	switch status {
	case LogoWarn:
		return LogoStyleWarn.Render(art)
	case LogoError:
		return LogoStyleError.Render(art)
	default:
		return LogoStyleInfo.Render(art)
	}
}
```

**Pattern for the OPTIONAL Phase 10 palette plumbing** (CONTEXT.md "Claude's Discretion" — planner picks):
```go
// Option A — keep current signature, no palette plumbing (logo color stays
// driven by LogoStyleInfo/Warn/Error which auto-inherit hex flips per D-415).
// Phase 10 does NOT need to thread palette here because:
//   1. The 24-bit hex flip propagates via the existing Color*Hex constants
//      to the existing LogoStyle* package vars
//   2. The 16-color fallback can be applied at lipgloss.Writer downsample
//      time (post-render) without per-renderer profile awareness
// This is the minimum-blast-radius choice.

// Option B — accept palette and re-pick Foreground per render call:
func RenderLogo(status LogoStatus, palette Palette, width int) string {
	_ = width
	art := strings.Join(LogoSmall, "\n")
	var fg lipgloss.Color
	switch status {
	case LogoWarn:
		fg = palette.Warning
	case LogoError:
		fg = palette.Error
	default:
		fg = palette.Accent
	}
	return lipgloss.NewStyle().Foreground(fg).Render(art)
}
// REJECTED — Option B introduces lipgloss.NewStyle() inside a renderer
// reachable from View(); will fail TestViewNoNewStyle's BFS walker.
```

**What to copy:**
- `func RenderX(status XStatus, width int) string` signature — the canonical pure-function-renderer shape
- `_ = unused` for plumbed-but-unused parameters (the `width` parameter precedent for Phase 10's optional `palette`)
- switch with default-fallthrough on the enum, NOT a map lookup (cheaper, more readable for 3-arm enums)

**Deviations the planner should know:**
- CONTEXT D-405 locks the 6-row logo art; do NOT introduce a 7th row even if `width` parameter triggers responsive trimming.
- If Plan 2 chooses Option A (no palette param), Phase 10's `RenderLogo` signature does NOT change — the cascade is one less file to update.
- Recommendation: Option A. The hex flip propagates automatically; the ANSI fallback flow handles the 16-color path via the writer downsample. Adding a 4th profile-aware variant block (`LogoStyleInfoANSI` etc.) is unnecessary if the writer downsamples accent → ANSI 13 cleanly, which it does per `colorprofile/writer.go`.

---

### Pattern 4 — `internal/ui/chrome.go` `RenderChrome` 3-tier width fallback (Phase 10 layers profile-aware variant selection)

**Analog:** `internal/ui/chrome.go:106-137` itself — Phase 7.1 D-116 established the 3-tier shape; Phase 10 layers `palette` selection inside.

**Existing 3-tier width-fallback pattern** (`internal/ui/chrome.go:106-137`):
```go
func RenderChrome(hints []keys.MenuHint, logoStatus LogoStatus, info InfoPanelData, width int) string {
	// Narrow tier: width below logoWidth + 2*minMenuCol (~41 cols).
	// Render a single-line muted stub; no border. Info parameter ignored.
	if width < logoWidth+menuCols*minMenuCol {
		return ChromeNarrowFallbackStyle.Render("press ? for help")
	}

	// Mid tier: width below infoPanelWidth + logoWidth + 2*minFullMenuCol
	// (~99 cols). Drop the info-panel slot; render menu+logo only with
	// generous menu width. Info parameter ignored per Phase 7.1 D-116.
	if width < infoPanelWidth+logoWidth+menuCols*minFullMenuCol {
		menuWidth := width - logoWidth
		menu := RenderMenu(hints, menuWidth)
		logo := RenderLogo(logoStatus, logoWidth)
		return lipgloss.JoinHorizontal(lipgloss.Top, menu, logo)
	}

	// Full tier: existing 3-slot layout. Phase 8 D-201..D-204: inflate the
	// 38x6 slot with the live info panel. RenderInfoPanel produces 5 rows
	// of label+value content; the InfoPanelPlaceholderStyle wrapper enforces
	// the 38x6 envelope per Phase 7 D-16 (unchanged in Phase 8).
	infoSlot := InfoPanelPlaceholderStyle.Render(RenderInfoPanel(info))
	menuWidth := width - infoPanelWidth - logoWidth
	if menuWidth < 1 {
		menuWidth = 1
	}
	menu := RenderMenu(hints, menuWidth)
	logo := RenderLogo(logoStatus, logoWidth)
	return lipgloss.JoinHorizontal(lipgloss.Top, infoSlot, menu, logo)
}
```

**Pattern for Phase 10's signature change** — palette parameter slots in after the existing positional args, threaded into each renderer call within each tier:
```go
// New signature — palette is the resolved color set; tiers stay unchanged.
// Each per-tier render call cascades the palette down.
func RenderChrome(
	hints []keys.MenuHint,
	logoStatus LogoStatus,
	info InfoPanelData,
	palette Palette,
	width int,
) string {
	if width < logoWidth+menuCols*minMenuCol {
		return ChromeNarrowFallbackStyle.Render("press ? for help")
	}
	if width < infoPanelWidth+logoWidth+menuCols*minFullMenuCol {
		menuWidth := width - logoWidth
		menu := RenderMenu(hints, palette, menuWidth)
		logo := RenderLogo(logoStatus, palette, logoWidth)
		return lipgloss.JoinHorizontal(lipgloss.Top, menu, logo)
	}
	infoSlot := InfoPanelPlaceholderStyle.Render(RenderInfoPanel(info, palette))
	menuWidth := width - infoPanelWidth - logoWidth
	if menuWidth < 1 {
		menuWidth = 1
	}
	menu := RenderMenu(hints, palette, menuWidth)
	logo := RenderLogo(logoStatus, palette, logoWidth)
	return lipgloss.JoinHorizontal(lipgloss.Top, infoSlot, menu, logo)
}
```

**What to copy:**
- The 3-tier `if width < ...` ladder shape — early-return on narrow tier, mid-tier branch, full-tier as fallthrough
- The named-constant width thresholds (`logoWidth`, `minMenuCol`, `minFullMenuCol`, `infoPanelWidth`) — DO NOT inline literal widths
- The `lipgloss.JoinHorizontal(lipgloss.Top, ...)` composition pattern across all tiers
- The doc comment with explicit threshold derivation (Narrow→Mid: `2 columns * minMenuCol (8) + logoWidth (25) = 41`; Mid→Full: `2 columns * minFullMenuCol (18) + infoPanelWidth (38) + logoWidth (25) = 99`)

**Deviations the planner should know:**
- Narrow-tier currently uses `ChromeNarrowFallbackStyle` (a package var with Foreground=ColorMuted). Phase 10's palette plumbing into narrow tier is OPTIONAL since the stub is 1 line + muted color. CONTEXT D-422 + Claude's discretion suggest narrow tier inherits the existing single-color path; the 4-profile matrix gets one representative ANSI golden for narrow per CONTEXT.md "Plan 3 author picks; recommendation: at narrow tier the chrome is a 1-row 'press ? for help' stub (Phase 7.1 D-116) so a single representative ANSI golden is sufficient."
- Profile-vs-Palette parameter type is per CONTEXT recommendation — `Palette` (avoids `colorprofile` import in every UI file). Plan 2 author confirms.
- Parameter order: planner picks `(hints, logoStatus, info, palette, width)` with palette before width (per the cascade order — palette is renderer-config, width is layout-config); BUT current signature has `info, width` — adding palette between can be `(hints, logoStatus, info, palette, width)` keeping width last. This is the natural append-mid pattern Phase 8 used when adding `info` parameter — match that.

---

### Pattern 5 — `internal/ui/crumbs.go` `RenderCrumbs` profile-aware bracket-fallback chip rendering

**Analog:** `internal/ui/crumbs.go:37-60` itself — Phase 8 established the chip-pill loop; Phase 10 layers profile branching inside.

**Existing pill-style chip-render pattern** (`internal/ui/crumbs.go:37-60`):
```go
func RenderCrumbs(segments []string, width int) string {
	if len(segments) == 0 {
		return CrumbRowStyle.Width(width).Render("")
	}
	normalised := normaliseSegments(segments)
	fitted := truncateSegmentsToWidth(normalised, width-2)

	chips := make([]string, 0, len(fitted))
	last := len(fitted) - 1
	for i, seg := range fitted {
		text := "<" + seg + ">"
		switch {
		case seg == ellipsisSentinel:
			chips = append(chips, CrumbChipEllipsisStyle.Render(text))
		case i == last:
			chips = append(chips, CrumbChipActiveStyle.Render(text))
		default:
			chips = append(chips, CrumbChipStyle.Render(text))
		}
	}
	joined := strings.Join(chips, " ")
	return CrumbRowStyle.Width(width).Render(joined)
}
```

**Pattern for Phase 10's profile-aware variant selection** — branch INSIDE the `switch` on `palette`'s implicit profile flag (or alongside `palette` carry a boolean `palette.Fallback`). Recommended approach: add a flag to `Palette` so renderers don't import `colorprofile`:
```go
// In styles.go — extend Palette:
type Palette struct {
	Accent  lipgloss.Color
	// ... existing fields
	// Fallback is true when the palette was resolved for Profile <= ANSI.
	// Renderers consult this to pick bracket-fallback variants instead of
	// pill-fill rendering (D-422).
	Fallback bool
}

// PaletteFor sets Fallback when profile <= ANSI:
func PaletteFor(profile colorprofile.Profile) Palette {
	if profile <= colorprofile.ANSI {
		return Palette{
			// ...
			Fallback: true,
		}
	}
	return Palette{
		// ...
		Fallback: false,
	}
}

// In crumbs.go — gate chip-style selection:
func RenderCrumbs(segments []string, palette Palette, width int) string {
	if len(segments) == 0 {
		return CrumbRowStyle.Width(width).Render("")
	}
	normalised := normaliseSegments(segments)
	fitted := truncateSegmentsToWidth(normalised, width-2)

	// Pick chip styles based on the resolved palette's profile tier.
	inactiveStyle := CrumbChipStyle
	activeStyle := CrumbChipActiveStyle
	if palette.Fallback {
		inactiveStyle = CrumbChipFallbackStyle
		activeStyle = CrumbChipActiveFallbackStyle
	}

	chips := make([]string, 0, len(fitted))
	last := len(fitted) - 1
	for i, seg := range fitted {
		text := "<" + seg + ">"
		switch {
		case seg == ellipsisSentinel:
			chips = append(chips, CrumbChipEllipsisStyle.Render(text))
		case i == last:
			chips = append(chips, activeStyle.Render(text))
		default:
			chips = append(chips, inactiveStyle.Render(text))
		}
	}
	joined := strings.Join(chips, " ")
	return CrumbRowStyle.Width(width).Render(joined)
}
```

**What to copy:**
- The `<segment>` literal wrapper (verbatim k9s parity per `crumbs.go` doc comment)
- The `truncateSegmentsToWidth(normalised, width-2)` `-2` row-padding budget — preserved
- The `last := len(fitted) - 1` + `i == last` active-chip detection
- Single-space `strings.Join(chips, " ")` separator (Phase 8 D-208)
- The `CrumbRowStyle.Width(width).Render(joined)` row-padding wrapper

**Deviations the planner should know:**
- The `Fallback` boolean on `Palette` is one of two valid encodings for "is this a 16-color path" — alternative is to expose `palette.Profile colorprofile.Profile` directly. Plan 2 author picks; Fallback flag is simpler and avoids the colorprofile import in renderer files.
- The `CrumbChipFallbackStyle` does NOT recolor the active chip's foreground — `Underline(true).Bold(true)` only. The fallback inactive chip uses `Foreground(ColorFgANSI)` plain. This is critical per Pitfall 5 §2 verbatim — recoloring the foreground inside the fallback variant defeats the "no paired bg/fg" survival rule.
- Bracket literals `<` `>` already pass `TestChromeASCIIOnly` (already in Phase 7 source); no allowlist update needed.
- Three NEW unit tests required per CONTEXT.md: bracket-fallback chip render assertion (active uses Underline+Bold SGR; inactive uses no bg fill); existing pill-fill tests stay green on TrueColor profile (Plan 3 splits).

---

### Pattern 6 — `internal/ui/menu.go` + `internal/ui/infopanel.go` profile-parameter cascade

**Analog:** `internal/ui/menu.go:46-127` + `internal/ui/infopanel.go:34-50`. Both already use package-var styles (`MenuKeyStyle`, `MenuDescStyle`, `InfoPanelLabelStyle`, `InfoPanelValueStyle`); Phase 10 may either (A) leave styles as-is and let auto-downsample at write time handle 16-color, or (B) gate style selection on `palette.Fallback`.

**Existing menu cell-render pattern** (`internal/ui/menu.go:118-127`):
```go
func renderMenuCell(h keys.MenuHint, colWidth int) string {
	keyLabel := MenuKeyStyle.Render("[" + h.Mnemonic + "]")
	keyVisible := lipgloss.Width(keyLabel) + 1
	descWidth := colWidth - keyVisible
	if descWidth < 1 {
		descWidth = 1
	}
	desc := ansi.Truncate(h.Description, descWidth, "")
	return keyLabel + " " + MenuDescStyle.Render(desc)
}
```

**Existing infopanel row-render pattern** (`internal/ui/infopanel.go:54-56`):
```go
func infoPanelRow(label, value string) string {
	return InfoPanelLabelStyle.Render(label) + InfoPanelValueStyle.Render(value)
}
```

**Pattern for Phase 10's profile-aware style selection** — IF the planner determines the auto-downsample is insufficient (Pitfall 5 §4 verification table fails for any chrome bg/fg pair), introduce parallel ANSI-variant styles and gate at render entry:
```go
// In styles.go — add ANSI variants ONLY for styles where Pitfall 5 §4 fails:
MenuKeyStyleANSI = lipgloss.NewStyle().Foreground(ColorAccentANSI)  // bright magenta
MenuDescStyleANSI = lipgloss.NewStyle().Foreground(ColorFgANSI)     // bright white

// In menu.go — gate at RenderMenu entry, not per-cell:
func RenderMenu(hints []keys.MenuHint, palette Palette, width int) string {
	keyStyle := MenuKeyStyle
	descStyle := MenuDescStyle
	if palette.Fallback {
		keyStyle = MenuKeyStyleANSI
		descStyle = MenuDescStyleANSI
	}
	// ... rest of body, threading keyStyle/descStyle into renderMenuCell
}
```

**What to copy:**
- Package-var style declarations (every style is a `var X = lipgloss.NewStyle()...` in styles.go, NEVER inline in the renderer)
- `ansi.Truncate` for cell-width clipping
- `lipgloss.JoinVertical(lipgloss.Left, rows...)` for rows; `lipgloss.JoinHorizontal(lipgloss.Top, ...)` for columns
- Single-pass for-range over hints/data with index-based column-major calculations

**Deviations the planner should know:**
- CONTEXT D-421 says all 5 renderers gain palette parameter. The IMPLEMENTATION can either thread palette through (and switch on `palette.Fallback` per renderer) OR rely on lipgloss writer-time downsample (the existing 24-bit styles auto-downsample to ANSI on `colorprofile.ANSI` writers). RESEARCH.md §1 says writer downsample IS effective for foreground-only chrome but NOT for paired bg+fg (chip fill). For Menu / InfoPanel where styles are foreground-only, the auto-downsample is sufficient — Plan 2 may elect to thread palette but not switch on it for these renderers. The bracket-fallback chip is the ONLY hard requirement for profile-aware branching (D-422). Plan 2 author picks; recommend: thread palette to all 5 renderers for symmetry, but only `crumbs.go` actually consults `palette.Fallback`.
- TestSubmodelViewsNoNewStyle (BFS walker) scans `internal/ui/*.go` and reports any `lipgloss.NewStyle()` reachable from a `View()` method. The Phase 10 ANSI variants live in `styles.go` (not reachable from View()) so they don't trip the walker.

---

### Pattern 7 — `internal/app/model.go` `AppModel.profile` field, `NewAppModel` signature change, `resolveLogoState` helper

**Analog:** `internal/app/model.go:222-305` `AppModel` struct + `NewAppModel` (existing field-on-construction pattern); 1572-1592 `titleForState()` (existing pure-function-of-state helper, perfect template for `resolveLogoState`)

**Existing `AppModel` field declaration pattern** (`internal/app/model.go:222-271`) — fields grouped by phase + tagged with phase comments:
```go
type AppModel struct {
	state        sessionState
	prevState    sessionState // restored when help/metadata overlay closes
	width        int
	height       int
	fileList     ui.FileListModel
	detail       ui.DetailModel
	help         ui.HelpModel
	status       ui.StatusBarModel
	// ... many fields ...

	// Phase 8 D-213: cached info-panel data. Refreshed at four event
	// seams (NewAppModel + FilesDiscoveredMsg + FilesParsedMsg +
	// GitStatusMsg). View() reads this cache only -- zero I/O at
	// render time (Pitfall 15).
	infoPanel ui.InfoPanelData
}
```

**Pattern for Phase 10's NEW `profile` field** — slot at the bottom of the struct alongside `infoPanel`, with a phase-tagged doc comment:
```go
type AppModel struct {
	// ... existing fields ...

	// Phase 8 D-213: cached info-panel data.
	infoPanel ui.InfoPanelData

	// Phase 10 D-419: terminal color profile detected once at startup
	// in cmd/sops-tui/main.go. Read-only after construction; NEVER
	// re-detected (Pitfall 15: zero per-frame I/O). Plumbed into
	// RenderChrome / RenderCrumbs / RenderMenu / RenderInfoPanel /
	// RenderLogo via ui.PaletteFor(m.profile).
	profile colorprofile.Profile
}
```

**Existing `NewAppModel` constructor pattern** (`internal/app/model.go:277-305`):
```go
func NewAppModel(env ui.EnvStatus, sopsYamlPath string) AppModel {
	m := AppModel{
		state:         stateFileList,
		fileList:      ui.NewFileListModel([]ui.FileItem{}, 0, 0),
		// ... lots of sub-model construction ...
		sopsYamlPath:  sopsYamlPath,
	}
	m.status.SetBreadcrumb("files")
	m.status.SetItemCount(0, "items")
	m.infoPanel = ui.InfoPanelData{
		SopsYamlRelPath: deriveSopsYamlRelPath(sopsYamlPath),
		AgeFingerprint:  loadAgeFingerprint(),
		RecipientCount:  -1,
		GitBranch:       "",
		FileCount:       -1,
	}
	return m
}
```

**Pattern for Phase 10's NEW `NewAppModel` signature** — append `profile` parameter at the end (matches Phase 8's append-end pattern when adding `sopsYamlPath`):
```go
func NewAppModel(env ui.EnvStatus, sopsYamlPath string, profile colorprofile.Profile) AppModel {
	m := AppModel{
		state:         stateFileList,
		// ... existing field initializers ...
		sopsYamlPath:  sopsYamlPath,
		profile:       profile,  // Phase 10 D-419
	}
	// ... rest unchanged
	return m
}
```

**Existing `titleForState` pure-function-of-state helper** (`internal/app/model.go:1569-1593`) — canonical template for `resolveLogoState`:
```go
func (m AppModel) titleForState() string {
	switch m.state {
	case stateFileList:
		return fmt.Sprintf("Files (%d)", m.fileList.ItemCount())
	case stateDetail:
		return "Detail: " + m.currentFile.Name
	// ... more cases ...
	}
	return ""
}
```

**Pattern for Phase 10's NEW `resolveLogoState` method** — pure function of state, walks Err checks first (D-404 precedence):
```go
// resolveLogoState computes the aggregate logo severity from env, flash
// severity, and health findings (D-401, D-402, D-403). Pure function of
// state — re-evaluated every View() frame; no caching, no sticky state.
// Severity precedence: Err > Warn > Info (D-404).
func (m AppModel) resolveLogoState() ui.LogoStatus {
	// Err checks (D-401)
	if m.status.FlashSeverity() == ui.FlashSevErr {
		return ui.LogoError
	}
	if !m.health.LastResult().IsEmpty() {
		// Health findings (Weak ∪ Duplicate ∪ Errors) raise to Err.
		// Stale files alone do NOT raise — see HealthCheckResult.IsEmpty.
		// Note: IsEmpty includes StaleFiles in its zero-check. Plan 1 author
		// chooses: either (a) add an IsEmptyExcludingStale() helper to
		// internal/health/checker.go, or (b) reach into the slices directly
		// here. Recommend (a) for centralised severity rules.
		hr := m.health.LastResult()
		if len(hr.WeakSecrets) > 0 || len(hr.Duplicates) > 0 || len(hr.Errors) > 0 {
			return ui.LogoError
		}
	}

	// Warn checks (D-402)
	if m.status.FlashSeverity() == ui.FlashSevWarn {
		return ui.LogoWarn
	}
	env := m.status.Env()
	if !env.AgeAvailable || !env.SopsYamlAvailable {
		return ui.LogoWarn
	}

	// Default
	return ui.LogoInfo
}
```

**Existing chrome composition callsite** (`internal/app/model.go:1395-1408`) — Phase 10 cascades signature changes through here (2 callsites total for `RenderChrome`, 2 for `RenderCrumbs`):
```go
// Phase 8 D-213 + D-216: chrome consumes m.infoPanel cache;
// crumbs row is unconditional (independent of chrome tier per
// D-216) -- the conditional guard from Phase 7 D-17 is removed.
chrome := ui.RenderChrome(hints, ui.LogoInfo, m.infoPanel, m.width)
crumbs := ui.RenderCrumbs(m.status.Segments(), m.width)
statusBar := m.status.View(m.width)
sections := []string{chrome, crumbs, wrapped, statusBar}
full := lipgloss.JoinVertical(lipgloss.Left, sections...)
```

**Pattern for Phase 10's UPDATED chrome composition callsite** — `LogoInfo` becomes `m.resolveLogoState()`; both calls accept the resolved palette:
```go
palette := ui.PaletteFor(m.profile)
chrome := ui.RenderChrome(hints, m.resolveLogoState(), m.infoPanel, palette, m.width)
crumbs := ui.RenderCrumbs(m.status.Segments(), palette, m.width)
statusBar := m.status.View(m.width)
sections := []string{chrome, crumbs, wrapped, statusBar}
full := lipgloss.JoinVertical(lipgloss.Left, sections...)
```

**Existing `chromeHeight`/`crumbsHeight` cascade** (`internal/app/model.go:1602-1621`):
```go
func chromeHeight(m AppModel) int {
	if m.width == 0 {
		return 0
	}
	chrome := ui.RenderChrome(m.menuHints(), ui.LogoInfo, m.infoPanel, m.width)
	return lipgloss.Height(chrome)
}

func crumbsHeight(m AppModel) int {
	if m.width == 0 {
		return 0
	}
	return lipgloss.Height(ui.RenderCrumbs(m.status.Segments(), m.width))
}
```

**Pattern for Phase 10's UPDATED `chromeHeight`/`crumbsHeight`** — same palette plumbing:
```go
func chromeHeight(m AppModel) int {
	if m.width == 0 {
		return 0
	}
	palette := ui.PaletteFor(m.profile)
	chrome := ui.RenderChrome(m.menuHints(), m.resolveLogoState(), m.infoPanel, palette, m.width)
	return lipgloss.Height(chrome)
}

func crumbsHeight(m AppModel) int {
	if m.width == 0 {
		return 0
	}
	palette := ui.PaletteFor(m.profile)
	return lipgloss.Height(ui.RenderCrumbs(m.status.Segments(), palette, m.width))
}
```

**41-callsite flash re-classification** (CONTEXT D-407) — full enumeration is in CONTEXT.md, but the per-callsite shape is identical: identify the callsite via `grep -n "m.status.Flash" internal/app/model.go`, classify by error/warn/info per D-401/D-402, replace the method name. Example transformations:
```go
// Before:
m.status, _ = m.status.Flash("Decrypt error: " + msg.Err.Error())  // line 440 - error path
m.status, _ = m.status.Flash("Decrypted")                          // line 444 - success path
m.status, _ = m.status.Flash("Reveal first with r")                // line 532 - validation block

// After (per CONTEXT D-407 classification):
m.status, _ = m.status.FlashErr("Decrypt error: " + msg.Err.Error())   // error path
m.status, _ = m.status.Flash("Decrypted")                              // success path → stays as Flash (alias for FlashInfo)
m.status, _ = m.status.FlashWarn("Reveal first with r")                // soft validation → Warn
```

**What to copy:**
- The "Phase N D-XYZ" comment tag on every new field (every Phase since 7 has done this — `Phase 7 D-22`, `Phase 8 D-213`, etc.)
- Field grouping: keep `profile` grouped with `infoPanel` at the bottom of the struct (those are the "Phase 8+/derived state" cluster)
- The `_ "import/path"` blank-import is NOT used here — `colorprofile` is a real value-type import in `internal/app/model.go`
- The pure-function-of-state helper shape: receiver method, no I/O, switch-on-state, default return
- Chrome composition callsite update: thread `palette` once via `ui.PaletteFor(m.profile)` then reuse — DO NOT call `PaletteFor` inside loops or in helpers reachable from View()

**Deviations the planner should know:**
- `m.health.LastResult()` does not yet exist — `HealthModel` has `SetResults` but no public read accessor. CONTEXT.md "Claude's Discretion" says: "Plan 1 author scouts; recommendation: add a small accessor on the existing sub-model rather than duplicating state." Plan 1 author will need to add `func (m HealthModel) LastResult() health.HealthCheckResult { return m.results }` to `internal/ui/health.go`.
- `HealthCheckResult` has `IsEmpty()` (lines 57-61) NOT `IsClean()` as CONTEXT references in §"Reusable Assets". Plan 1 author uses `IsEmpty()` — but per D-401 the severity classifier needs to EXCLUDE StaleFiles from the "raise to Err" check, so `IsEmpty()` (which includes StaleFiles) is the wrong predicate. Plan 1 author either (a) reaches into the slices directly with `len(hr.WeakSecrets) > 0 || len(hr.Duplicates) > 0 || len(hr.Errors) > 0`, or (b) adds a centralised `HasErrLevelFindings()` method on `HealthCheckResult`. Recommend (b) for documentation + future-proofing.
- The `Flash` callsite count: CONTEXT says "42 callsites" but the grep shows 41 matches in current model.go. The discrepancy is likely a count drift since CONTEXT was written; Plan 1 author re-counts and audits. The migration sweep is identical regardless of exact count.
- `colorprofile` is currently an indirect dependency (per RESEARCH.md). Phase 10 promotes it to a direct require in BOTH `cmd/sops-tui/main.go` AND `internal/app/model.go` (the `AppModel.profile` field's type). `go mod tidy` after Plan 2's edits resolves the directness.

---

### Pattern 8 — `cmd/sops-tui/main.go` startup-detected option threaded into `NewAppModel`

**Analog:** `cmd/sops-tui/main.go:62-71` — Step 5 builds `EnvStatus`; Step 6 constructs AppModel. Phase 10 inserts a "Step 5.5" between them.

**Existing startup pattern** (`cmd/sops-tui/main.go:60-71`):
```go
// Step 5: Build env status from validation results for the status bar
env := ui.EnvStatus{
	SopsAvailable:     !hasResultWithMessage(results, "sops binary not found"),
	AgeAvailable:      !hasResultWithMessage(results, "Age key file not found"),
	SopsYamlAvailable: !hasResultWithMessage(results, ".sops.yaml not found"),
}

// Step 6: Create and run the root TUI program (View().AltScreen = true in AppModel)
sopsYamlPath, _ := validator.FindSopsYaml(opts.StartDir)
model := app.NewAppModel(env, sopsYamlPath)
p := tea.NewProgram(model)
```

**Pattern for Phase 10's startup-time detection** — slot a "Step 5.5" with the same numbered-comment style:
```go
// Step 5: Build env status from validation results for the status bar
env := ui.EnvStatus{
	SopsAvailable:     !hasResultWithMessage(results, "sops binary not found"),
	AgeAvailable:      !hasResultWithMessage(results, "Age key file not found"),
	SopsYamlAvailable: !hasResultWithMessage(results, ".sops.yaml not found"),
}

// Step 5.5: Detect terminal color profile once at startup (Phase 10 D-419).
// Profile drives bracket-vs-pill chip rendering and ANSI palette fallback.
// SOPSTUI_FORCE_ASCII=1 is supported as an escape hatch for users on
// terminals that mis-self-report (recommended in CONTEXT.md Claude's Discretion).
profile := colorprofile.Detect(os.Stdout, os.Environ())
if os.Getenv("SOPSTUI_FORCE_ASCII") == "1" {
	profile = colorprofile.Ascii
}

// Step 6: Create and run the root TUI program (View().AltScreen = true in AppModel)
sopsYamlPath, _ := validator.FindSopsYaml(opts.StartDir)
model := app.NewAppModel(env, sopsYamlPath, profile)
p := tea.NewProgram(model)
```

**Existing import group pattern** (`cmd/sops-tui/main.go:12-26`) — alphabetical within stdlib + alphabetical within third-party + alphabetical within project blocks:
```go
import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"

	"github.com/caesarakalaeii/sops-tui/internal/app"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
	"github.com/caesarakalaeii/sops-tui/internal/validator"
)
```

**Pattern for Phase 10's NEW import** — slot `colorprofile` into the third-party block alphabetically:
```go
import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/colorprofile"

	"github.com/caesarakalaeii/sops-tui/internal/app"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
	"github.com/caesarakalaeii/sops-tui/internal/validator"
)
```

**What to copy:**
- The numbered `// Step N:` comment style with sub-step `// Step N.5:` for inserts
- One-shot detection — declared with `:=`, used immediately, no caching mechanism beyond the `model.profile` field
- The env-override-after-detect order: detect first, then check env override, then construct
- Three-block import grouping (stdlib, third-party, project) with blank-line separators

**Deviations the planner should know:**
- `colorprofile.Ascii` is the alias name for `colorprofile.ASCII` per RESEARCH.md (`profile.go:13-26` — `const Ascii = ASCII`). Plan 2 author may use either; CONTEXT.md uses `Ascii` (alias).
- The `tea.WithColorProfile(profile)` option is OPTIONAL per RESEARCH.md §1 — bubbletea v2 auto-detects internally. Plan 2 author MAY add it for symmetry but RESEARCH recommends NOT adding it (Option A). The current `tea.NewProgram(model)` line stays as-is.
- The `SOPSTUI_FORCE_ASCII` override is per CONTEXT.md Claude's Discretion — a 4-line addition. Recommended yes.

---

### Pattern 9 — Test pattern: 4-profile golden matrix (cross-product of width × profile)

**Analog:** `internal/app/menuhints_drift_test.go:166-208` `TestMenuGolden` (13-state run-table) + `internal/app/resize_test.go:28-74` (4-width matrix). Phase 10's matrix composes the two: width × profile cross-product.

**Existing 13-state run-table pattern** (`internal/app/menuhints_drift_test.go:166-208`):
```go
// TestMenuGolden locks the rendered persistent menu per (state, IsSearchActive)
// tuple. 13 sub-tests, one golden file per state. RequireGoldenStructure
// strips ANSI so Phase 10's palette pass does not churn (D-311).
//
// Generate goldens initially with: GOLDEN_UPDATE=1 go test ./internal/app/... -run TestMenuGolden
func TestMenuGolden(t *testing.T) {
	const width = 80

	run := func(name string, setup func(m *AppModel)) {
		t.Run(name, func(t *testing.T) {
			m := buildAppModel(t)
			if setup != nil {
				setup(&m)
			}
			hints := m.menuHints()
			rendered := ui.RenderMenu(hints, width)
			testutil.RequireGoldenStructure(t, "menu_"+name, rendered)
		})
	}

	run("file_list", nil) // default state
	run("file_list_search", func(m *AppModel) {
		updated, _ := m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
		*m = updated.(AppModel)
		// ... search activation
	})
	run("detail", func(m *AppModel) { m.state = stateDetail })
	// ... 10 more
}
```

**Existing 4-width golden matrix pattern** (`internal/app/resize_test.go:28-74`):
```go
func TestResize_40x12(t *testing.T) {
	setDeterministicAgeEnv(t)
	m := app.NewAppModel(defaultEnv(), "")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 12})
	m = updated.(app.AppModel)

	v := m.View()
	testutil.RequireGoldenStructure(t, "resize_40x12", v.Content)
	testutil.RequireGoldenColors(t, "resize_40x12", v.Content, nil)
}
// ... TestResize_80x24, _120x40, _200x60 with same shape
```

**Pattern for Phase 10's NEW 4-profile × representative-state matrix** — combine the two analogs into a nested run-table:
```go
// TestRenderChrome_4ProfileMatrix locks the rendered chrome per (profile, scene)
// tuple. 16 goldens (4 profiles × 4 representative scenes per CONTEXT D-423).
// RequireGoldenColors asserts severity-distinct color SGR bytes per profile.
//
// Generate goldens initially with: GOLDEN_UPDATE=1 go test ./internal/app/... -run TestRenderChrome_4ProfileMatrix
func TestRenderChrome_4ProfileMatrix(t *testing.T) {
	scenes := []struct {
		name string
		// scene-specific setup
		setup func(m *app.AppModel)
	}{
		{"chrome_full", nil},
		{"crumbs_active", func(m *app.AppModel) { m.SetTestCrumbs([]string{"sops-tui", "files", "prod.yaml"}) }},
		{"menu_populated", func(m *app.AppModel) { m.SetTestState(app.StateDetail) }},
		// ... 4th scene per CONTEXT.md
	}
	profiles := []struct {
		name    string
		profile colorprofile.Profile
	}{
		{"ascii", colorprofile.Ascii},
		{"ansi", colorprofile.ANSI},
		{"ansi256", colorprofile.ANSI256},
		{"truecolor", colorprofile.TrueColor},
	}
	for _, p := range profiles {
		for _, s := range scenes {
			t.Run(p.name+"_"+s.name, func(t *testing.T) {
				setDeterministicAgeEnv(t)
				m := app.NewAppModel(defaultEnv(), "", p.profile)  // Phase 10 signature
				updated, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 60})
				m = updated.(app.AppModel)
				if s.setup != nil {
					s.setup(&m)
				}
				v := m.View()
				name := s.name + "_" + p.name
				testutil.RequireGoldenStructure(t, name, v.Content)
				// Assert profile-specific color presence:
				testutil.RequireGoldenColors(t, name, v.Content,
					expectedColorsFor(p.profile))
			})
		}
	}
}

// expectedColorsFor returns the SGR-byte substrings that must appear in
// rendered output for the given profile. TrueColor expects "203;166;247" (Mauve);
// ANSI expects "\x1b[95m" (bright magenta). See Pitfall 5 §4 verification table.
func expectedColorsFor(profile colorprofile.Profile) []string {
	if profile <= colorprofile.ANSI {
		return []string{"\x1b[95m"} // bright magenta
	}
	return []string{"203;166;247"} // Catppuccin Mauve RGB
}
```

**Pattern for Phase 10's NEW 60×24 + 100×30 width goldens** (CONTEXT D-424) — copy `TestResize_40x12` shape verbatim, change widths:
```go
// TestResize_60x24 — Phase 10 D-424: mid-narrow tier between 40x12 and 80x24.
// Verifies UI-16 critical-data-survival at narrow → mid chrome boundary.
func TestResize_60x24(t *testing.T) {
	setDeterministicAgeEnv(t)
	m := app.NewAppModel(defaultEnv(), "", colorprofile.TrueColor)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m = updated.(app.AppModel)

	v := m.View()
	testutil.RequireGoldenStructure(t, "resize_60x24", v.Content)
	testutil.RequireGoldenColors(t, "resize_60x24", v.Content, nil)
}

// TestResize_100x30 — Phase 10 D-424: mid-tier with crumbs + chrome.
func TestResize_100x30(t *testing.T) {
	setDeterministicAgeEnv(t)
	m := app.NewAppModel(defaultEnv(), "", colorprofile.TrueColor)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = updated.(app.AppModel)

	v := m.View()
	testutil.RequireGoldenStructure(t, "resize_100x30", v.Content)
	testutil.RequireGoldenColors(t, "resize_100x30", v.Content, nil)
}
```

**Existing color-presence-by-named-constant pattern** (no current `internal/app/*_test.go` references hex literals; `internal/ui/styles_test.go:23-26` uses the `Color*Hex` constants directly). Phase 10 wires color-presence assertions referencing `ui.ColorAccentHex` etc. so the hex flip propagates without test churn (D-417):
```go
// Phase 10 NEW pattern for color-presence assertions in app/chrome_test.go:
testutil.RequireGoldenColors(t, name, v.Content, []string{
	rgbForHex(ui.ColorAccentHex),  // helper that converts "#cba6f7" → "203;166;247"
})

// Helper (testutil candidate):
func rgbForHex(hex string) string {
	// Strip leading '#', split into 3 pairs, parse as decimal triplet.
	r, _ := strconv.ParseUint(hex[1:3], 16, 8)
	g, _ := strconv.ParseUint(hex[3:5], 16, 8)
	b, _ := strconv.ParseUint(hex[5:7], 16, 8)
	return fmt.Sprintf("%d;%d;%d", r, g, b)
}
```

**What to copy:**
- The `t.Run("name", func(t *testing.T) { ... })` sub-test loop pattern
- `testutil.RequireGoldenStructure` for ANSI-stripped structural check + `testutil.RequireGoldenColors` for raw SGR-byte presence
- The `setDeterministicAgeEnv(t)` helper invocation at the top of each test (Phase 8 D-220 / golden review-safety)
- The `tea.WindowSizeMsg{Width: W, Height: H}` first-frame init pattern
- `GOLDEN_UPDATE=1 go test ...` regen idiom — friction-by-design per Phase 6 D-10

**Deviations the planner should know:**
- The 4-profile matrix forces profile via the `NewAppModel(profile)` constructor parameter — NOT via mutating `lipgloss.Writer.Profile` global (that would be a flake source per RESEARCH.md §4 Option B warning).
- Color-presence assertions: Phase 10 must not hardcode hex literals (`"#cba6f7"`) or RGB triplets (`"203;166;247"`) — wire through `rgbForHex(ui.ColorAccentHex)` so the hex flip cascades.
- `app.AppModel` does NOT currently expose `SetTestCrumbs`/`SetTestState` exported helpers. The `_test.go` setup callbacks in CONTEXT use sub-model accessors; for the 4-profile matrix the planner may need a small `_test_helpers.go` with `_ "package_only_test_API"` patterns if the tests live in `app_test` (external). The existing `ParsedFileForTest` (`model.go:209-211`) is the precedent for exported test helpers.
- ~16 new goldens cap per CONTEXT.md (4 profiles × 4 scenes minus overlap). Plan 3 author confirms exact list.
- The `expectedColorsFor` helper goes in either `internal/testutil/golden.go` (if reused across packages) or a local `*_test.go` helper. Plan 3 author picks; recommend testutil for reusability across `internal/ui/*_test.go` too.

---

## Shared Patterns

### Authentication / Authorization
N/A — sops-tui is a single-user CLI, no auth surface.

### Error Handling
**Source:** `internal/app/model.go` flash-on-error pattern (e.g. line 440 `m.status, _ = m.status.Flash("Decrypt error: " + msg.Err.Error())`)
**Apply to:** All Phase 10 error-classification call-sites
**Phase 10 transformation:** Re-classify each `Flash(...)` to `FlashErr(...)` / `FlashWarn(...)` / `FlashInfo(...)` per D-407 + D-401/D-402.

### Validation
**Source:** `internal/ui/statusbar.go:160-165` flash-empty-string-as-zero-state pattern (`if m.flash != "" { ... }`)
**Apply to:** Phase 10's `FlashSeverity()` accessor (returns `FlashSevInfo` when `m.flash == ""`)

### Logging
**Source:** None — sops-tui has no logger; status bar flash is the user-facing surface
**Apply to:** N/A

### Doc Comment Cross-References
**Source:** Every Phase 1-9 file uses `// Phase N D-XYZ:` tags on field declarations + `// per UI-SPEC §...` cross-references
**Apply to:** All Phase 10 NEW vars, fields, methods, and tests
**Example:** Phase 10's `flashSeverity FlashSeverity` field gains `// Phase 10 D-410: severity field updated by FlashInfo/Warn/Err methods; cleared on FlashClearMsg ack alongside m.flash.`

### Package-Var Style Discipline
**Source:** `internal/app/chrome_test.go:160-281` `TestViewNoNewStyle` BFS walker
**Apply to:** All Phase 10 NEW styles — `FlashWarnBarStyle`, `FlashErrBarStyle`, `MenuKeyStyleANSI` (if introduced), `CrumbChipFallbackStyle`, `CrumbChipActiveFallbackStyle`, `LogoStyleInfoANSI` etc.
**Verification:** Plan 1 + Plan 2 + Plan 3 all run `go test ./internal/app/... -run TestViewNoNewStyle` after each change.

### Grep-Gate Discipline
**Source:** `internal/app/chrome_test.go:54-138` `TestChromeASCIIOnly` + `TestChromeNormalBorderOnly`
**Apply to:** Phase 10's bracket-fallback chip uses `<` `>` (already in scan target — `internal/ui/crumbs.go`); no widening of allowlist needed. The `[W]` `[E]` flash prefix lives in `internal/ui/statusbar.go` which is NOT in the chrome ASCII-allowlist scope (per existing Phase 7 + 8 scoping: only `chrome.go`, `logo.go`, `menu.go`, `crumbs.go`, `infopanel.go`).

### Testing
**Source:** `internal/testutil/golden.go` `RequireGoldenStructure` (ANSI-stripped) + `RequireGoldenColors` (raw SGR bytes)
**Apply to:** All Phase 10 NEW goldens
**GOLDEN_UPDATE workflow:** `GOLDEN_UPDATE=1 go test ./...` (friction-by-design — D-416 atomic regen pass).

---

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| (none) | — | — | Every Phase 10 NEW symbol has a same-package, same-role, same-data-flow precedent. The closest "novel" pattern — the 4-profile teatest matrix — is a cross-product of two existing matrices (Phase 9's 13-state + Phase 7.1's 4-width) and reuses every primitive. |

---

## Metadata

**Analog search scope:**
- `cmd/sops-tui/` (1 file: main.go)
- `internal/app/` (10 .go files + tests)
- `internal/ui/` (33 .go files + tests)
- `internal/health/` (checker.go for HealthCheckResult shape)
- `internal/testutil/` (golden.go for assertion helpers)
- `internal/keys/` (bindings.go for keymap dispatcher)

**Files scanned:** ~50 source files (via grep + 12 fully-Read source files + 10 test files)

**Key constants verified:**
- `internal/ui/styles.go:21` `ColorAccentHex = "#89b4fa"` (24-bit current; Phase 10 → `"#cba6f7"`)
- `internal/ui/styles.go:25` `ColorWarningHex = "#f9e2af"` (Phase 10 → `"#fab387"`)
- `internal/ui/styles.go:27` `ColorErrorHex = "#f38ba8"` (Phase 10 → `"#eba0ac"`)
- `internal/ui/logo.go:24-33` `LogoStatus` iota enum (canonical — Phase 10's `FlashSeverity` mirror)
- `internal/health/checker.go:57-61` `HealthCheckResult.IsEmpty()` (NOT `IsClean()` — CONTEXT typo; planner uses IsEmpty + slice-length checks per D-401 stale-files exclusion)

**Pattern extraction date:** 2026-05-04

## PATTERN MAPPING COMPLETE
