package ui_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

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
	// ColorAccent #89b4fa -> RGB 137;180;250
	// ColorBg     #1e1e2e -> RGB 30;30;46
	// Bold SGR 1 -> lipgloss/v2 encodes bold as "1;" at the start of the
	// combined SGR sequence: e.g. \x1b[1;38;2;30;30;46;48;2;137;180;250m
	out := ui.RenderCrumbs([]string{"sops-tui", "files", "metadata"}, ui.PaletteFor(colorprofile.TrueColor), 80)
	assert.Contains(t, out, "137;180;250", "active chip must apply ColorAccent bg (D-206)")
	assert.Contains(t, out, "30;30;46", "active chip must invert fg to ColorBg (D-206)")
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
	out := ui.RenderCrumbs([]string{"sops-tui"}, ui.PaletteFor(colorprofile.TrueColor), 80)
	assert.Contains(t, out, "137;180;250", "single segment must be active (D-206)")
}

func TestRenderCrumbs_InactiveChipColors(t *testing.T) {
	// D-206 inactive: bg ColorSurface #313244 -> 49;50;68; fg ColorFg #cdd6f4 -> 205;214;244.
	out := ui.RenderCrumbs([]string{"sops-tui", "files"}, ui.PaletteFor(colorprofile.TrueColor), 80)
	assert.Contains(t, out, "49;50;68", "inactive chip must apply ColorSurface bg (D-206)")
	assert.Contains(t, out, "205;214;244", "inactive chip must apply ColorFg fg (D-206)")
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
