// Package ui - chrome composer, titled-border wrapper, and title overlay
// helper (Phase 7).
//
// RenderChrome composes the persistent 6-row header band via
// JoinHorizontal of an info-panel placeholder (reserved for Phase 8),
// the persistent keybinding menu, and the ASCII logo.
//
// RenderChrome falls back through three width tiers per
// 07.1-UI-SPEC.md Section B (Phase 7.1 D-116): at width >= 99 cols the
// full 3-slot (info-panel + menu + logo) layout renders; at
// 41 <= width < 99 the info-panel slot is dropped (mid tier); at
// width < 41 a single-line "press ? for help" stub renders (narrow
// tier). All three tiers preserve UI-15's NormalBorder-only +
// ASCII-only contract. (Threshold formulas live with their constants
// at minMenuCol / minFullMenuCol below.)
//
// WrapTitled wraps any sub-model body in a NormalBorder box with a
// title injected into the top border line via overlayTitle. The wrapper
// uses the TitledBorderStyle package var (BorderForeground: ColorMuted,
// Padding(0, 1)) per D-12 / D-13 - UI-15 grep-gate enforces NormalBorder
// exclusivity in Plan 3.
//
// overlayTitle is a community-standard string-splice pattern for
// lipgloss v2, which as of 2026-04-24 has no native border-title API.
// CONTEXT.md D-14 cited charmbracelet/soft-serve/pkg/ui/components/header
// as a reference; the Phase 7 research (07-RESEARCH.md section"Closed Research
// Gaps" #1) verified that soft-serve main HEAD does NOT contain the
// pattern. The pattern is community-standard across the bubbletea
// ecosystem (gh, glow, charm's own examples) - see 07-RESEARCH.md section1
// for the full reference-implementation gap analysis and pkg.go.dev
// verification of lipgloss v2's lack of native border-title API.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
)

// infoPanelWidth reserves the Phase 8 header info-panel column (D-16).
// Plan 3's golden refresh verifies the 38-col + 25-col + residual menu
// width math at 40 / 80 / 120 / 200 column terminals.
const infoPanelWidth = 38

// logoWidth is the fixed Phase 7 logo envelope per D-01 (Candidate A).
const logoWidth = 25

// minMenuCol is the per-column floor for menu cell width - below this,
// the menu would be too narrow to render mnemonic + description
// cleanly. Used by Phase 7.1 D-116 width-fallback as the narrow-tier
// boundary multiplier (per-column * menuCols=2).
const minMenuCol = 8

// minFullMenuCol is the per-column comfortable budget for the full
// 3-slot tier - below this per-column budget, the info-panel slot
// crowds the menu so much that descriptions truncate to ~4 cells. The
// mid-tier (menu+logo, info-panel dropped) gives much more breathing
// room at these middling widths and is the right rendering for
// width=80 (per 07.1-CONTEXT D-119 + 07.1-PLAN-05 success criteria).
//
// Phase 7.1 D-116 deviation note: the original CONTEXT formula used a
// single minMenuCol for both tier boundaries, which placed width=80 in
// full-tier (since 80 >= 38+25+8=71) - but the plan explicitly required
// 80x24 to render mid-tier (menu+logo only). The 18-cell comfortable
// floor reconciles the two: narrow at width<41, mid at 41<=width<99,
// full at width>=99. Verified at the four golden widths (40, 80, 120,
// 200) per 07.1-05-SUMMARY.md.
const minFullMenuCol = 18

// minTitledWidth / minTitledHeight clamp WrapTitled arguments so the
// underlying lipgloss border math never renders a sub-border box.
const (
	minTitledWidth  = 4
	minTitledHeight = 3
)

// RenderChrome composes the persistent header band, falling back through
// three width tiers per 07.1-UI-SPEC.md Section B (Phase 7.1 D-116):
//
//   - Full (width >= infoPanelWidth + logoWidth + 2*minFullMenuCol, ~99+):
//     [info-panel-placeholder 38x6 | menu residual x 6 | logo 25x6] -
//     the existing 3-slot layout.
//   - Mid (logoWidth + 2*minMenuCol <= width < ~99):
//     [menu | logo] - info-panel slot dropped so the menu has breathing
//     room at middling widths.
//   - Narrow (width < logoWidth + 2*minMenuCol, ~41):
//     single-line "press ? for help" stub via ChromeNarrowFallbackStyle.
//
// Threshold derivation: each tier requires per-column legibility.
// Narrow->Mid boundary: 2 columns * minMenuCol (8) + logoWidth (25) = 41.
// Mid->Full boundary: 2 columns * minFullMenuCol (18) + infoPanelWidth (38) +
// logoWidth (25) = 99. See minFullMenuCol doc for the comfort-budget
// rationale; the formula reconciles 07.1-CONTEXT D-116 with the plan's
// per-width golden expectations (40 narrow, 80 mid, 120/200 full).
//
// logoStatus is plumbed for Phase 10 severity coupling; Phase 7 callers
// pass LogoInfo unconditionally per D-02. The narrow-tier fallback was
// added in Phase 7.1 D-116 to close WR-03 / 07-VERIFICATION.md
// Anti-Pattern 4 (chrome overflows at narrow widths, body unreachable
// at 40x12).
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
		// menuWidth is guaranteed >= 2*minMenuCol because the narrow-tier
		// check above caught width < logoWidth + 2*minMenuCol.
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

// WrapTitled wraps body in a NormalBorder box of outer width x height
// with title injected into the top border line via overlayTitle.
// width and height are clamped to minTitledWidth / minTitledHeight so
// lipgloss never produces a degenerate box.
//
// In lipgloss v2, Style.Width(W).Height(H) on a bordered style sets
// the OUTER rendered dimensions (border included). The 6-row Phase 7
// chrome therefore passes the full body envelope (w, h) straight
// through to TitledBorderStyle.Width(w).Height(h); the horizontal
// Padding(0, 1) from TitledBorderStyle is absorbed into the inner
// content area, leaving (width - 4) usable cells per row for the body.
func WrapTitled(title, body string, width, height int) string {
	if width < minTitledWidth {
		width = minTitledWidth
	}
	if height < minTitledHeight {
		height = minTitledHeight
	}
	rendered := TitledBorderStyle.
		Width(width).
		Height(height).
		Render(body)
	return overlayTitle(rendered, " "+title+" ")
}

// overlayTitle injects title into the top border line of rendered at
// column position 2, preserving the top-left corner (col 0) and
// top-right corner (last col). Width of the first line is preserved
// exactly via replacement (not insertion). Titles wider than
// firstLineWidth - 4 are truncated with the ellipsis rune via
// ansi.Truncate.
//
// Behavior:
//   - Empty title -> returns rendered byte-identical.
//   - Single-line input (no '\n') -> returns rendered unchanged.
//   - firstLineWidth < 4 -> returns rendered unchanged (too narrow).
//
// This is a community-standard pattern for lipgloss v2, which has no
// native border-title API (verified 2026-04-24 against lipgloss v2.0.3
// pkg.go.dev). See package doc and 07-RESEARCH.md section1 for the full
// reference-implementation gap analysis and the soft-serve revision
// audit that closed CONTEXT.md D-14's source citation.
func overlayTitle(rendered, title string) string {
	if title == "" {
		return rendered
	}
	nl := strings.IndexByte(rendered, '\n')
	if nl < 0 {
		return rendered
	}
	firstLine := rendered[:nl]
	rest := rendered[nl:]

	firstLineWidth := lipgloss.Width(firstLine)
	if firstLineWidth < 4 {
		return rendered
	}

	maxTitleWidth := firstLineWidth - 4
	titleW := lipgloss.Width(title)
	if titleW > maxTitleWidth {
		title = ansi.Truncate(title, maxTitleWidth, "…")
		titleW = lipgloss.Width(title)
	}

	newFirstLine := spliceRenderedLine(firstLine, 2, 2+titleW, title)
	return newFirstLine + rest
}

// spliceRenderedLine replaces the visible-column range [startCol, endCol)
// of firstLine with replacement. firstLine is assumed to be the top
// border line of a NormalBorder box - pure ASCII-range box-drawing
// characters with no ANSI sequences embedded inside the border characters
// themselves. (TitledBorderStyle applies BorderForeground via a single SGR
// pair wrapping the whole line; the splice preserves that wrapper by
// passing through ANSI escape sequences while only counting visible
// columns against the replacement window.)
//
// Implementation: walk runes tracking visible column position; emit
// replacement once col reaches startCol; skip visible runes in
// [startCol, endCol); copy the rest unchanged. ANSI SGR sequences are
// written through verbatim and do not advance the column counter.
func spliceRenderedLine(firstLine string, startCol, endCol int, replacement string) string {
	var out strings.Builder
	col := 0
	inSGR := false
	var pendingSGR strings.Builder

	emitted := false
	for _, r := range firstLine {
		// Track ANSI SGR sequences: they don't advance column count.
		if r == 0x1b { // ESC
			inSGR = true
			pendingSGR.Reset()
			pendingSGR.WriteRune(r)
			continue
		}
		if inSGR {
			pendingSGR.WriteRune(r)
			if r == 'm' {
				// End of SGR sequence - emit it with the output.
				out.WriteString(pendingSGR.String())
				inSGR = false
			}
			continue
		}

		// Visible rune handling.
		if col == startCol && !emitted {
			out.WriteString(replacement)
			emitted = true
		}
		if col < startCol || col >= endCol {
			out.WriteRune(r)
		}
		col++
	}
	// If startCol was at the very end and we never emitted, tack
	// replacement on at the end (shouldn't happen for well-formed
	// border lines but defensive).
	if !emitted {
		out.WriteString(replacement)
	}
	return out.String()
}
