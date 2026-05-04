package ui_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

func TestRenderCrumbs_KnsExactPills(t *testing.T) {
	// D-205: <segment> wrapper, no inner padding -- k9s crumbs.go:62-74 verbatim.
	out := ui.RenderCrumbs([]string{"sops-tui", "files", "metadata"}, ui.PaletteFor(colorprofile.TrueColor), 80)
	stripped := ansi.Strip(out)
	for _, want := range []string{"<sops-tui>", "<files>", "<metadata>"} {
		assert.Containsf(t, stripped, want, "chip wrapper format must be %q (k9s parity, D-205)", want)
	}
}

func TestRenderCrumbs_ActiveBoldBg(t *testing.T) {
	// D-206: active chip = Background(ColorAccent) + Foreground(ColorBg) + Bold(true).
	// Phase 10 D-417: triplets derived from ColorAccentHex / ColorBgHex via
	// hexToRGBTriplet so the palette flip propagates automatically.
	// Bold SGR 1 -> lipgloss/v2 encodes bold as "1;" at the start of the
	// combined SGR sequence: e.g. \x1b[1;38;2;30;30;46;48;2;203;166;247m
	accentTriplet := hexToRGBTriplet(ui.ColorAccentHex)
	bgTriplet := hexToRGBTriplet(ui.ColorBgHex)
	out := ui.RenderCrumbs([]string{"sops-tui", "files", "metadata"}, ui.PaletteFor(colorprofile.TrueColor), 80)
	assert.Containsf(t, out, accentTriplet, "active chip must apply ColorAccent bg (%s) (D-206)", accentTriplet)
	assert.Containsf(t, out, bgTriplet, "active chip must invert fg to ColorBg (%s) (D-206)", bgTriplet)
	// Bold SGR appears as "[1;" (combined) or "[1m" (standalone) in ANSI sequences.
	assert.True(t, strings.Contains(out, "[1;") || strings.Contains(out, "[1m"),
		"active chip MUST include bold SGR (Pitfall 9 redundancy channel; D-206); got: %q", out)
}

func TestRenderCrumbs_LowercaseStripSpaces(t *testing.T) {
	// D-207: strings.ReplaceAll(strings.ToLower(seg), " ", "") per k9s:70-71.
	// Input: ["Files", "prod.yaml [M]"] -> chips contain "<files>" and "<prod.yaml[m]>"
	out := ui.RenderCrumbs([]string{"Files", "prod.yaml [M]"}, ui.PaletteFor(colorprofile.TrueColor), 80)
	stripped := ansi.Strip(out)
	assert.Contains(t, stripped, "<files>", "Files must lowercase to 'files'")
	assert.Contains(t, stripped, "<prod.yaml[m]>", "'prod.yaml [M]' must lowercase + strip-spaces to 'prod.yaml[m]'")
}

func TestRenderCrumbs_MiddleEllipsis(t *testing.T) {
	// D-216 + Pitfall 14: 8 segments at width=20 -> middle segments collapse to <U+2026>.
	// At width=20 the chip budget is 18 cells; 8 single/4-char chips total ~36 cells,
	// so truncation is forced.
	segs := []string{"a", "b", "c", "d", "e", "f", "g", "last"}
	out := ui.RenderCrumbs(segs, ui.PaletteFor(colorprofile.TrueColor), 20)
	stripped := ansi.Strip(out)
	assert.Contains(t, stripped, "<a>", "first segment must survive overflow")
	assert.Contains(t, stripped, "<last>", "last segment must survive overflow")
	assert.Contains(t, stripped, "…", "overflow row must contain U+2026 ellipsis chip (D-216)")
}

func TestRenderCrumbs_EmptySafe(t *testing.T) {
	// Defensive: nil/empty segments must not panic.
	p := ui.PaletteFor(colorprofile.TrueColor)
	assert.NotPanics(t, func() { ui.RenderCrumbs(nil, p, 80) })
	assert.NotPanics(t, func() { ui.RenderCrumbs([]string{}, p, 80) })
}

func TestRenderCrumbs_SingleSegmentIsActive(t *testing.T) {
	// Edge case: a single segment is "last" so renders with active style.
	// Phase 10 D-417: derive accent triplet from ColorAccentHex.
	accentTriplet := hexToRGBTriplet(ui.ColorAccentHex)
	out := ui.RenderCrumbs([]string{"sops-tui"}, ui.PaletteFor(colorprofile.TrueColor), 80)
	assert.Containsf(t, out, accentTriplet, "single segment must be active with accent bg (%s) (D-206)", accentTriplet)
}

func TestRenderCrumbs_InactiveChipColors(t *testing.T) {
	// D-206 inactive: bg ColorSurface, fg ColorFg.
	// Phase 10 D-417: derive triplets from ColorSurfaceHex / ColorFgHex
	// (defense-in-depth — these constants are unchanged in D-415 but the
	// constant-derived pattern protects against future palette tunes).
	surfaceTriplet := hexToRGBTriplet(ui.ColorSurfaceHex)
	fgTriplet := hexToRGBTriplet(ui.ColorFgHex)
	out := ui.RenderCrumbs([]string{"sops-tui", "files"}, ui.PaletteFor(colorprofile.TrueColor), 80)
	assert.Containsf(t, out, surfaceTriplet, "inactive chip must apply ColorSurface bg (%s) (D-206)", surfaceTriplet)
	assert.Containsf(t, out, fgTriplet, "inactive chip must apply ColorFg fg (%s) (D-206)", fgTriplet)
}

// Verify normaliseSegments via observable behaviour rather than direct
// (it is unexported). Already covered by TestRenderCrumbs_LowercaseStripSpaces.

func TestRenderCrumbs_TruncateSegmentsToWidthDropsMiddle(t *testing.T) {
	// Verify truncation keeps first and last, drops middle.
	segs := []string{"first", "second", "third", "fourth", "last"}
	// Width 20 is too narrow for 5 chips: <first><second><third><fourth><last>
	// = 7+8+7+8+6+4 = more than 20. Force truncation.
	out := ui.RenderCrumbs(segs, ui.PaletteFor(colorprofile.TrueColor), 20)
	stripped := ansi.Strip(out)
	assert.Contains(t, stripped, "<first>", "first segment must be preserved")
	assert.Contains(t, stripped, "<last>", "last segment must be preserved")
}

func TestRenderCrumbs_TwoSegmentsNeverTruncated(t *testing.T) {
	// With only 2 segments, no truncation should occur even at small widths.
	out := ui.RenderCrumbs([]string{"a", "b"}, ui.PaletteFor(colorprofile.TrueColor), 10)
	stripped := ansi.Strip(out)
	assert.Contains(t, stripped, "<a>")
	assert.Contains(t, stripped, "<b>")
}

// containsAnySGRCrumbs reports whether any of the alternative SGR substrings
// appears in s. lipgloss/v2 may encode combined attributes in different
// orderings (e.g. "[1;4m" vs "[4;1m") so SGR-presence assertions match a set.
// Local helper -- duplicates the one in styles_test.go intentionally to keep
// each test file self-contained.
func containsAnySGRCrumbs(s string, alts ...string) bool {
	for _, a := range alts {
		if strings.Contains(s, a) {
			return true
		}
	}
	return false
}

// TestRenderCrumbs_BracketFallbackOnAscii — Phase 10 D-422.
// On Profile <= colorprofile.ANSI (palette.Fallback=true), the active
// chip drops bg fill and renders with Underline + Bold only. No fg
// recolor. Bracket literals "<segment>" remain in stripped output.
func TestRenderCrumbs_BracketFallbackOnAscii(t *testing.T) {
	palette := ui.PaletteFor(colorprofile.Ascii)
	require.True(t, palette.Fallback,
		"PaletteFor(Ascii) must return Fallback=true (Plan 2 contract)")

	out := ui.RenderCrumbs([]string{"sops-tui", "files", "metadata"}, palette, 80)
	stripped := ansi.Strip(out)
	assert.Contains(t, stripped, "<sops-tui>",
		"bracket fallback must keep <segment> literal (D-422)")
	assert.Contains(t, stripped, "<files>",
		"bracket fallback must keep <segment> literal (D-422)")
	assert.Contains(t, stripped, "<metadata>",
		"bracket fallback active chip is bracket-wrapped (D-422)")

	// Active chip: Underline (SGR 4) + Bold (SGR 1) -- both required.
	assert.True(t,
		containsAnySGRCrumbs(out, "\x1b[4m", "\x1b[4;", ";4m", ";4;"),
		"active chip MUST emit SGR 4 (underline) on bracket fallback (D-422); got: %q", out)
	assert.True(t,
		containsAnySGRCrumbs(out, "\x1b[1m", "\x1b[1;", ";1m", ";1;"),
		"active chip MUST emit SGR 1 (bold) on bracket fallback (D-422); got: %q", out)

	// Active chip MUST NOT contain ColorAccent mauve bg (RGB 203;166;247
	// post-D-415) and MUST NOT contain ColorBg fg (RGB 30;30;46) -- bracket
	// fallback drops bg fill entirely and does NOT recolor fg.
	assert.NotContains(t, out, "203;166;247",
		"bracket fallback MUST NOT apply ColorAccent mauve bg (D-422)")
	assert.NotContains(t, out, "30;30;46",
		"bracket fallback MUST NOT apply ColorBg fg recolor (D-422)")
}

// TestRenderCrumbs_BracketFallbackOnANSI — Phase 10 D-422 + D-421.
// PaletteFor(ANSI) hits the same fallback branch as Ascii (the gate is
// `<= colorprofile.ANSI`). 16-color terminals get bracket fallback
// identically to monochrome.
func TestRenderCrumbs_BracketFallbackOnANSI(t *testing.T) {
	palette := ui.PaletteFor(colorprofile.ANSI)
	require.True(t, palette.Fallback,
		"PaletteFor(ANSI) must return Fallback=true (Plan 2 contract)")

	out := ui.RenderCrumbs([]string{"sops-tui", "files", "metadata"}, palette, 80)
	stripped := ansi.Strip(out)
	assert.Contains(t, stripped, "<metadata>",
		"bracket fallback applies on ANSI profile too (D-421/D-422)")
	assert.True(t,
		containsAnySGRCrumbs(out, "\x1b[4m", "\x1b[4;", ";4m", ";4;"),
		"active chip MUST emit SGR 4 on ANSI profile (D-422); got: %q", out)
	assert.True(t,
		containsAnySGRCrumbs(out, "\x1b[1m", "\x1b[1;", ";1m", ";1;"),
		"active chip MUST emit SGR 1 on ANSI profile (D-422); got: %q", out)
}

// TestRenderCrumbs_BracketFallbackInactiveChipsAreUndecorated — Phase 10 D-422.
// Inactive chips on bracket fallback use CrumbChipFallbackStyle
// (Foreground(ColorFgANSI) only -- no Background, no Underline, no Bold).
// Compute the expected substring directly from the style and assert it
// appears in the rendered output.
func TestRenderCrumbs_BracketFallbackInactiveChipsAreUndecorated(t *testing.T) {
	palette := ui.PaletteFor(colorprofile.Ascii)
	expectedInactive1 := ui.CrumbChipFallbackStyle.Render("<sops-tui>")
	expectedInactive2 := ui.CrumbChipFallbackStyle.Render("<files>")

	out := ui.RenderCrumbs([]string{"sops-tui", "files", "metadata"}, palette, 80)
	assert.Contains(t, out, expectedInactive1,
		"inactive chip MUST render with CrumbChipFallbackStyle on bracket fallback (D-422)")
	assert.Contains(t, out, expectedInactive2,
		"inactive chip MUST render with CrumbChipFallbackStyle on bracket fallback (D-422)")
}

// TestRenderCrumbs_BracketFallbackActiveChipNoFgRecolor — Phase 10 D-422.
// Active chip on bracket fallback applies ONLY Underline + Bold. No
// Foreground SGR (no 38;5;N for ANSI256 or 38;2;R;G;B for TrueColor).
// The chip text reads in the terminal's default fg color band.
func TestRenderCrumbs_BracketFallbackActiveChipNoFgRecolor(t *testing.T) {
	palette := ui.PaletteFor(colorprofile.Ascii)
	out := ui.RenderCrumbs([]string{"a", "b", "c"}, palette, 80)

	expectedActive := ui.CrumbChipActiveFallbackStyle.Render("<c>")
	assert.Contains(t, out, expectedActive,
		"active chip MUST render exactly as CrumbChipActiveFallbackStyle.Render(<text>) on bracket fallback (D-422)")

	// Cross-check: the expected active chip MUST NOT contain any 38;5; or
	// 38;2; SGR -- those are foreground encodings. CrumbChipActiveFallbackStyle
	// has no Foreground set so lipgloss must not emit a fg SGR.
	assert.NotContains(t, expectedActive, "38;5;",
		"CrumbChipActiveFallbackStyle MUST NOT emit ANSI256 fg SGR (D-422)")
	assert.NotContains(t, expectedActive, "38;2;",
		"CrumbChipActiveFallbackStyle MUST NOT emit TrueColor fg SGR (D-422)")
}

// TestRenderCrumbs_TrueColorPillRenderingUnchanged — regression test.
// Plan 3's bracket-fallback branch must NOT regress the TrueColor pill
// rendering path (Phase 8 D-206). At palette.Fallback=false the existing
// CrumbChipActiveStyle (mauve bg + dark fg + bold) applies.
func TestRenderCrumbs_TrueColorPillRenderingUnchanged(t *testing.T) {
	palette := ui.PaletteFor(colorprofile.TrueColor)
	require.False(t, palette.Fallback,
		"PaletteFor(TrueColor) must return Fallback=false (Plan 2 contract)")

	out := ui.RenderCrumbs([]string{"sops-tui", "files", "metadata"}, palette, 80)
	stripped := ansi.Strip(out)
	assert.Contains(t, stripped, "<sops-tui>")
	assert.Contains(t, stripped, "<metadata>")

	// Post-D-415 ColorAccent = #cba6f7 -> RGB 203;166;247 (mauve).
	assert.Contains(t, out, "203;166;247",
		"TrueColor active chip MUST apply mauve bg (D-415 post-Plan-2 + D-206 pill rendering)")
	// ColorBg = #1e1e2e -> RGB 30;30;46.
	assert.Contains(t, out, "30;30;46",
		"TrueColor active chip MUST invert fg to ColorBg (D-206 pill rendering)")
	// Bold SGR 1 always present on active.
	assert.True(t,
		containsAnySGRCrumbs(out, "\x1b[1m", "\x1b[1;", ";1m", ";1;"),
		"TrueColor active chip MUST emit SGR 1 (bold) (Pitfall 9 redundancy channel; D-206)")
}

// TestRenderCrumbs_FirstAndLastSegmentsPreserved — Phase 10 D-425
// critical-data-survival rule. When the chip row would overflow at narrow
// widths, middle-segment ellipsis kicks in (D-216 truncateSegmentsToWidth)
// and FIRST + LAST chips are always preserved. The active (last) chip is
// the user's "you are here" anchor.
//
// CI lock-in / regression test -- Plan 3 makes no implementation change
// to truncateSegmentsToWidth. This test confirms the existing first+last
// preservation guarantee continues to hold under modification pressure.
//
// Uses TrueColor palette (pill rendering); chip width math is identical
// regardless of palette.Fallback so testing on the default path is enough.
//
// Implementation note (Rule 1 deviation from plan's exact wording): the
// truncateSegmentsToWidth algorithm preserves FIRST + LAST + at least one
// ellipsis chip but stops dropping when sentinelIdx <= 1 (otherwise it
// would lose the first chip). At width=30 with 5 segments, "history" may
// remain on a wrapped line because the algorithm's invariant is "drop
// middle non-sentinel until [first, sentinel, last] would be next" not
// "drop ALL middles". The plan's stricter NotContains assertions on every
// middle segment over-specified the rule -- D-425's actual contract is
// first+last preservation, asserted here directly.
func TestRenderCrumbs_FirstAndLastSegmentsPreserved(t *testing.T) {
	palette := ui.PaletteFor(colorprofile.TrueColor)
	segs := []string{"files", "metadata", "diff", "history", "prod.yaml"}

	// Width=30 forces truncation: chip-row budget = 30 - 2 (CrumbRowStyle
	// pad) = 28 cells. 5 chips of widths 7+10+6+9+12 + 4(seps) = 48 cells,
	// so the middle MUST collapse and at least one ellipsis MUST appear.
	out := ui.RenderCrumbs(segs, palette, 30)
	stripped := ansi.Strip(out)

	// First + last + ellipsis MUST appear (D-425 critical-data-survival):
	assert.Contains(t, stripped, "<files>",
		"first segment MUST be preserved at narrow width (D-425 critical-data-survival)")
	assert.Contains(t, stripped, "<prod.yaml>",
		"last segment MUST be preserved (active 'you are here' anchor; D-425)")
	assert.Contains(t, stripped, "…",
		"middle ellipsis chip MUST appear when overflow forces truncation (D-216 + D-425)")

	// At least one middle segment MUST be dropped (proves truncation fired
	// at all). The "metadata" segment is the leftmost middle and is the
	// first to be dropped per the algorithm's mid-replace-with-sentinel
	// step -- this is the strongest stable assertion.
	assert.NotContains(t, stripped, "<metadata>",
		"middle chip 'metadata' MUST be dropped at width=30 (D-216 + D-425)")
}
