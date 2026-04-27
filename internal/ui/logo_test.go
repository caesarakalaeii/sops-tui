// Tests for the Phase 7 logo primitive: LogoSmall byte-art, LogoStatus enum,
// and RenderLogo (D-01, D-02, D-03, UI-SPEC §"Visuals — Logo").
package ui_test

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

// TestLogoSmall_SixRows verifies the byte-art is exactly 6 rows per D-01.
func TestLogoSmall_SixRows(t *testing.T) {
	require.Len(t, ui.LogoSmall, 6, "LogoSmall must contain exactly 6 rows per D-01")
}

// TestLogoSmall_ASCIIOnly enforces UI-15 + Pitfall 6: no emoji, no VS16,
// no ZWJ. Every rune must be in the ASCII range (<= 0x7F).
func TestLogoSmall_ASCIIOnly(t *testing.T) {
	for i, row := range ui.LogoSmall {
		for _, r := range row {
			require.LessOrEqualf(t, r, rune(0x7F),
				"row %d contains non-ASCII rune U+%04X — violates UI-15 / Pitfall 6", i, r)
		}
	}
}

// TestLogoSmall_WidthInRange verifies every row's lipgloss-measured width
// is within the locked envelope ~25 cols (Candidate A target per Research §5).
func TestLogoSmall_WidthInRange(t *testing.T) {
	for i, row := range ui.LogoSmall {
		w := lipgloss.Width(row)
		assert.GreaterOrEqual(t, w, 22, "row %d width %d is below 22-col floor", i, w)
		assert.LessOrEqual(t, w, 26, "row %d width %d is above 26-col ceiling", i, w)
	}
}

// TestRenderLogo_ReturnsSixRows verifies that RenderLogo emits exactly 6
// content rows (5 inter-row newlines after ANSI-stripping).
func TestRenderLogo_ReturnsSixRows(t *testing.T) {
	rendered := ui.RenderLogo(ui.LogoInfo, 26)
	stripped := ansi.Strip(rendered)
	// 6 rows joined with 5 "\n" — assert exactly 5 newlines.
	assert.Equal(t, 5, strings.Count(stripped, "\n"),
		"expected 5 inter-row newlines (6 rows total), got: %q", stripped)
}

// TestRenderLogo_AllStatusVariants verifies all three severity styles
// render without panic and embed the expected RGB triplet for the chosen
// foreground color (lipgloss emits TrueColor SGR sequences containing the
// raw r;g;b values — stable across terminal output).
func TestRenderLogo_AllStatusVariants(t *testing.T) {
	// ColorAccent #89b4fa -> rgb(137, 180, 250) -> "137;180;250"
	infoRendered := ui.RenderLogo(ui.LogoInfo, 26)
	require.NotEmpty(t, infoRendered, "Info render must be non-empty")
	assert.Contains(t, infoRendered, "137;180;250",
		"LogoInfo must embed ColorAccent RGB triplet (#89b4fa)")

	// ColorWarning #f9e2af -> rgb(249, 226, 175) -> "249;226;175"
	warnRendered := ui.RenderLogo(ui.LogoWarn, 26)
	require.NotEmpty(t, warnRendered, "Warn render must be non-empty")
	assert.Contains(t, warnRendered, "249;226;175",
		"LogoWarn must embed ColorWarning RGB triplet (#f9e2af)")

	// ColorError #f38ba8 -> rgb(243, 139, 168) -> "243;139;168"
	errorRendered := ui.RenderLogo(ui.LogoError, 26)
	require.NotEmpty(t, errorRendered, "Error render must be non-empty")
	assert.Contains(t, errorRendered, "243;139;168",
		"LogoError must embed ColorError RGB triplet (#f38ba8)")
}

// TestRenderLogo_Width0NoPanic verifies degenerate width input does not
// panic — width is plumbed for Phase 10 width-responsive logic but ignored
// in Phase 7. The art is locked at ~25 cols per D-01.
func TestRenderLogo_Width0NoPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		_ = ui.RenderLogo(ui.LogoInfo, 0)
	})
}

// TestLogoStatus_Enum locks the iota ordering — Info=0, Warn=1, Error=2.
// Phase 10 (UI-03) consumes the same ordering for severity coupling.
func TestLogoStatus_Enum(t *testing.T) {
	assert.Equal(t, ui.LogoStatus(0), ui.LogoInfo)
	assert.Equal(t, ui.LogoStatus(1), ui.LogoWarn)
	assert.Equal(t, ui.LogoStatus(2), ui.LogoError)
}
