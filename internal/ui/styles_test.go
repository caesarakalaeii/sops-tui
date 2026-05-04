package ui_test

import (
	"image/color"
	"reflect"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"github.com/caesarakalaeii/sops-tui/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStyleColorHexValues verifies each exported color hex constant has the correct value
// per 01-UI-SPEC.md §Color. The *Hex constants are the canonical source of truth;
// the color.Color vars are derived from them via lipgloss.Color().
func TestStyleColorHexValues(t *testing.T) {
	tests := []struct {
		name    string
		got     string
		wantHex string
	}{
		{"ColorBgHex", ui.ColorBgHex, "#1e1e2e"},
		{"ColorSurfaceHex", ui.ColorSurfaceHex, "#313244"},
		{"ColorAccentHex", ui.ColorAccentHex, "#cba6f7"},   // Catppuccin Mauve (D-415)
		{"ColorSuccessHex", ui.ColorSuccessHex, "#a6e3a1"},
		{"ColorWarningHex", ui.ColorWarningHex, "#fab387"}, // Catppuccin Peach (D-415)
		{"ColorErrorHex", ui.ColorErrorHex, "#eba0ac"},     // Catppuccin Maroon (D-415)
		{"ColorMutedHex", ui.ColorMutedHex, "#6c7086"},
		{"ColorFgHex", ui.ColorFgHex, "#cdd6f4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantHex, tt.got, "hex constant %s must equal %s", tt.name, tt.wantHex)
		})
	}
}

// TestStyleHexConstantsStartWithHash verifies all hex color constants start with '#',
// which is the structural proof that AdaptiveColor is not used — AdaptiveColor does not
// produce plain hex strings.
func TestStyleHexConstantsStartWithHash(t *testing.T) {
	hexColors := []struct {
		name  string
		value string
	}{
		{"ColorBgHex", ui.ColorBgHex},
		{"ColorSurfaceHex", ui.ColorSurfaceHex},
		{"ColorAccentHex", ui.ColorAccentHex},
		{"ColorSuccessHex", ui.ColorSuccessHex},
		{"ColorWarningHex", ui.ColorWarningHex},
		{"ColorErrorHex", ui.ColorErrorHex},
		{"ColorMutedHex", ui.ColorMutedHex},
		{"ColorFgHex", ui.ColorFgHex},
	}
	for _, c := range hexColors {
		t.Run(c.name, func(t *testing.T) {
			assert.True(t, strings.HasPrefix(c.value, "#"),
				"hex constant %s must be a plain hex value starting with '#': got %q", c.name, c.value)
		})
	}
}

// TestStyleSpacingConstants verifies spacing token constants are defined with correct values
// per 01-UI-SPEC.md §Spacing Scale.
func TestStyleSpacingConstants(t *testing.T) {
	assert.Equal(t, 1, ui.SpaceXS, "SpaceXS must be 1")
	assert.Equal(t, 2, ui.SpaceSM, "SpaceSM must be 2")
	assert.Equal(t, 4, ui.SpaceMD, "SpaceMD must be 4")
	assert.Equal(t, 6, ui.SpaceLG, "SpaceLG must be 6")
	assert.Equal(t, 8, ui.SpaceXL, "SpaceXL must be 8")
	assert.Equal(t, 2, ui.TreeIndent, "TreeIndent must be 2")
}

// TestStyleRenderNonEmpty verifies that named style variables produce non-empty rendered output.
func TestStyleRenderNonEmpty(t *testing.T) {
	styles := []struct {
		name     string
		rendered string
	}{
		{"ErrorLabel", ui.ErrorLabel.Render("[ERROR]")},
		{"WarnLabel", ui.WarnLabel.Render("[WARN]")},
		{"StatusBarStyle", ui.StatusBarStyle.Render("status")},
		{"BreadcrumbActive", ui.BreadcrumbActive.Render("active")},
		{"BreadcrumbSep", ui.BreadcrumbSep.Render(">")},
		{"TreeConnector", ui.TreeConnector.Render("├─")},
		{"TreeIndicator", ui.TreeIndicator.Render("[+]")},
		{"DimText", ui.DimText.Render("***")},
		{"SelectedRow", ui.SelectedRow.Render("selected")},
		{"HelpKeyStyle", ui.HelpKeyStyle.Render("?")},
		{"HelpDescStyle", ui.HelpDescStyle.Render("description")},
		{"HelpSectionHeader", ui.HelpSectionHeader.Render("Navigation")},
		// Phase 2 styles
		{"BadgeUnencrypted", ui.BadgeUnencrypted.Render("[unencrypted]")},
		{"BadgePlain", ui.BadgePlain.Render("[plain]")},
		{"TypeHintStyle", ui.TypeHintStyle.Render("(str)")},
		{"SearchInputStyle", ui.SearchInputStyle.Render("filter text")},
		{"SearchMatchStyle", ui.SearchMatchStyle.Render("match")},
	}
	for _, s := range styles {
		t.Run(s.name, func(t *testing.T) {
			assert.NotEmpty(t, strings.TrimSpace(s.rendered),
				"style %s must render non-empty string", s.name)
		})
	}
}

// TestPhase2StyleColorValues verifies the 5 new Phase 2 named styles have correct color attributes
// per 02-UI-SPEC.md §New Named Styles.
func TestPhase2StyleColorValues(t *testing.T) {
	t.Run("BadgeUnencrypted_bold_and_error_color", func(t *testing.T) {
		// BadgeUnencrypted must be Bold and use ColorError (#f38ba8)
		rendered := ui.BadgeUnencrypted.Render("[unencrypted]")
		require.NotEmpty(t, strings.TrimSpace(rendered), "BadgeUnencrypted must render non-empty")
		// Verify it renders with some ANSI styling (not plain text)
		assert.NotEqual(t, "[unencrypted]", rendered,
			"BadgeUnencrypted must apply ANSI styling (bold + color)")
	})

	t.Run("BadgePlain_warning_color", func(t *testing.T) {
		rendered := ui.BadgePlain.Render("[plain]")
		require.NotEmpty(t, strings.TrimSpace(rendered), "BadgePlain must render non-empty")
		assert.NotEqual(t, "[plain]", rendered,
			"BadgePlain must apply ANSI styling (color)")
	})

	t.Run("TypeHintStyle_faint_and_muted_color", func(t *testing.T) {
		rendered := ui.TypeHintStyle.Render("(str)")
		require.NotEmpty(t, strings.TrimSpace(rendered), "TypeHintStyle must render non-empty")
		assert.NotEqual(t, "(str)", rendered,
			"TypeHintStyle must apply ANSI styling (faint + color)")
	})

	t.Run("SearchInputStyle_surface_background", func(t *testing.T) {
		rendered := ui.SearchInputStyle.Render("filter text")
		require.NotEmpty(t, strings.TrimSpace(rendered), "SearchInputStyle must render non-empty")
	})

	t.Run("SearchMatchStyle_accent_color", func(t *testing.T) {
		rendered := ui.SearchMatchStyle.Render("match")
		require.NotEmpty(t, strings.TrimSpace(rendered), "SearchMatchStyle must render non-empty")
		assert.NotEqual(t, "match", rendered,
			"SearchMatchStyle must apply ANSI styling (accent color)")
	})
}

// Phase 10 Plan 2 (D-415): explicit Catppuccin lock for the three flipped
// hex constants. Future palette tunes flag this test as the canonical guard.
func TestStyleColorHexValues_Catppuccin(t *testing.T) {
	assert.Equal(t, "#cba6f7", ui.ColorAccentHex, "ColorAccentHex must equal Catppuccin Mauve per D-415")
	assert.Equal(t, "#fab387", ui.ColorWarningHex, "ColorWarningHex must equal Catppuccin Peach per D-415")
	assert.Equal(t, "#eba0ac", ui.ColorErrorHex, "ColorErrorHex must equal Catppuccin Maroon per D-415")
}

// TestStyleColorHexValues_UnchangedConstants is the regression guard:
// Bg / Surface / Success / Muted / Fg are EXPLICITLY unchanged in D-415.
func TestStyleColorHexValues_UnchangedConstants(t *testing.T) {
	assert.Equal(t, "#1e1e2e", ui.ColorBgHex, "ColorBgHex MUST NOT change in D-415")
	assert.Equal(t, "#313244", ui.ColorSurfaceHex, "ColorSurfaceHex MUST NOT change in D-415")
	assert.Equal(t, "#a6e3a1", ui.ColorSuccessHex, "ColorSuccessHex MUST NOT change in D-415")
	assert.Equal(t, "#6c7086", ui.ColorMutedHex, "ColorMutedHex MUST NOT change in D-415")
	assert.Equal(t, "#cdd6f4", ui.ColorFgHex, "ColorFgHex MUST NOT change in D-415")
}

// TestColorXANSI_IndicesPerD420 locks the 8 ANSI fallback indices.
// Hand-verification table: 10-RESEARCH.md §"ANSI 16-Color Verification".
func TestColorXANSI_IndicesPerD420(t *testing.T) {
	cases := []struct {
		name string
		got  color.Color
		want color.Color
	}{
		{"ColorAccentANSI", ui.ColorAccentANSI, lipgloss.ANSIColor(13)},
		{"ColorBgANSI", ui.ColorBgANSI, lipgloss.ANSIColor(0)},
		{"ColorSurfaceANSI", ui.ColorSurfaceANSI, lipgloss.ANSIColor(8)},
		{"ColorFgANSI", ui.ColorFgANSI, lipgloss.ANSIColor(15)},
		{"ColorMutedANSI", ui.ColorMutedANSI, lipgloss.ANSIColor(7)},
		{"ColorSuccessANSI", ui.ColorSuccessANSI, lipgloss.ANSIColor(10)},
		{"ColorWarningANSI", ui.ColorWarningANSI, lipgloss.ANSIColor(11)},
		{"ColorErrorANSI", ui.ColorErrorANSI, lipgloss.ANSIColor(9)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.got, "ANSI fallback %s must equal lipgloss.ANSIColor(N) per D-420", tc.name)
		})
	}
}

// TestPalette_StructFields locks the Palette shape — 8 color fields
// (color.Color from image/color, the interface lipgloss.Color() and
// lipgloss.ANSIColor() both satisfy) + 1 bool field. Plan 3 reads
// palette.Fallback to gate bracket chip rendering; this test is the
// contract guard.
func TestPalette_StructFields(t *testing.T) {
	pt := reflect.TypeOf(ui.Palette{})
	require.Equal(t, 9, pt.NumField(), "Palette must have exactly 9 fields")

	expectedColorFields := map[string]bool{
		"Accent":  true,
		"Bg":      true,
		"Surface": true,
		"Fg":      true,
		"Muted":   true,
		"Success": true,
		"Warning": true,
		"Error":   true,
	}
	const expectedBoolField = "Fallback"
	seen := map[string]bool{}
	for i := 0; i < pt.NumField(); i++ {
		f := pt.Field(i)
		switch {
		case expectedColorFields[f.Name]:
			// color.Color is an interface; reflect Kind reports "interface".
			assert.Equalf(t, reflect.Interface, f.Type.Kind(),
				"field %s must be color.Color (interface), got kind %s", f.Name, f.Type.Kind())
			assert.Containsf(t, f.Type.String(), "Color", "field %s must be a Color-named type", f.Name)
		case f.Name == expectedBoolField:
			assert.Equal(t, "bool", f.Type.Kind().String(), "field Fallback must be bool")
		default:
			t.Fatalf("unexpected Palette field %q", f.Name)
		}
		seen[f.Name] = true
	}
	for name := range expectedColorFields {
		assert.Truef(t, seen[name], "Palette must have field %q", name)
	}
	assert.Truef(t, seen[expectedBoolField], "Palette must have field %q", expectedBoolField)
}

// TestPaletteFor_TrueColorReturns24BitVariants asserts PaletteFor(TrueColor)
// returns the 24-bit Catppuccin Mocha variants with Fallback=false.
func TestPaletteFor_TrueColorReturns24BitVariants(t *testing.T) {
	p := ui.PaletteFor(colorprofile.TrueColor)
	assert.Equal(t, ui.ColorAccent, p.Accent, "TrueColor accent must be 24-bit Mauve")
	assert.Equal(t, ui.ColorWarning, p.Warning, "TrueColor warning must be 24-bit Peach")
	assert.Equal(t, ui.ColorError, p.Error, "TrueColor error must be 24-bit Maroon")
	assert.Equal(t, ui.ColorBg, p.Bg)
	assert.Equal(t, ui.ColorSurface, p.Surface)
	assert.Equal(t, ui.ColorFg, p.Fg)
	assert.Equal(t, ui.ColorMuted, p.Muted)
	assert.Equal(t, ui.ColorSuccess, p.Success)
	assert.False(t, p.Fallback, "TrueColor profile must NOT be flagged as fallback (Plan 3 gate)")
}

// TestPaletteFor_ANSI256Returns24BitVariants asserts ANSI256 also returns
// 24-bit (the gate is profile <= ANSI; ANSI256 > ANSI so the 24-bit branch is taken).
func TestPaletteFor_ANSI256Returns24BitVariants(t *testing.T) {
	p := ui.PaletteFor(colorprofile.ANSI256)
	assert.Equal(t, ui.ColorAccent, p.Accent)
	assert.False(t, p.Fallback, "ANSI256 profile must NOT be flagged as fallback")
}

// TestPaletteFor_ANSIReturnsANSIVariants asserts profile=ANSI takes the fallback branch.
func TestPaletteFor_ANSIReturnsANSIVariants(t *testing.T) {
	p := ui.PaletteFor(colorprofile.ANSI)
	assert.Equal(t, ui.ColorAccentANSI, p.Accent, "ANSI accent must be ANSIColor(13) per D-420")
	assert.Equal(t, ui.ColorWarningANSI, p.Warning, "ANSI warning must be ANSIColor(11) per D-420")
	assert.Equal(t, ui.ColorErrorANSI, p.Error, "ANSI error must be ANSIColor(9) per D-420")
	assert.Equal(t, ui.ColorBgANSI, p.Bg)
	assert.Equal(t, ui.ColorSurfaceANSI, p.Surface)
	assert.Equal(t, ui.ColorFgANSI, p.Fg)
	assert.Equal(t, ui.ColorMutedANSI, p.Muted)
	assert.Equal(t, ui.ColorSuccessANSI, p.Success)
	assert.True(t, p.Fallback, "ANSI profile MUST be flagged as fallback (Plan 3 gates bracket rendering on this)")
}

// TestPaletteFor_AsciiReturnsANSIVariants — Ascii (TTY-no-color) takes the same fallback branch as ANSI.
func TestPaletteFor_AsciiReturnsANSIVariants(t *testing.T) {
	p := ui.PaletteFor(colorprofile.Ascii)
	assert.Equal(t, ui.ColorAccentANSI, p.Accent)
	assert.True(t, p.Fallback, "Ascii profile MUST be flagged as fallback")
}

// TestPaletteFor_NoTTYReturnsANSIVariants — NoTTY < ANSI, so the <= gate captures it.
func TestPaletteFor_NoTTYReturnsANSIVariants(t *testing.T) {
	p := ui.PaletteFor(colorprofile.NoTTY)
	assert.True(t, p.Fallback, "NoTTY profile MUST be flagged as fallback (gate is <= ANSI)")
}

// containsAnySGR reports whether any of the alternative SGR substrings
// appears in s. lipgloss/v2 may encode combined attributes in different
// orderings (e.g. "[1;4m" vs "[4;1m") so SGR-presence assertions match a set.
func containsAnySGR(s string, alts ...string) bool {
	for _, a := range alts {
		if strings.Contains(s, a) {
			return true
		}
	}
	return false
}

// Phase 10 Plan 3 (D-422): bracket-fallback chip styles for Profile <=
// colorprofile.ANSI. CrumbChipFallbackStyle is the inactive chip variant
// (Foreground(ColorFgANSI) only, no bg fill). CrumbChipActiveFallbackStyle
// is the active chip variant (Underline + Bold only, no bg fill, no fg
// recolor — the decoration channel survives every profile downsample
// including monochrome per 10-RESEARCH.md §lipgloss.Style.Underline).

// TestCrumbChipFallbackStyle_NoBg asserts the inactive bracket-fallback
// chip has NO background set — the design rule is "bracket fallback drops
// bg fill entirely" per D-422.
func TestCrumbChipFallbackStyle_NoBg(t *testing.T) {
	zero := lipgloss.NewStyle().GetBackground()
	got := ui.CrumbChipFallbackStyle.GetBackground()
	assert.Equal(t, zero, got,
		"CrumbChipFallbackStyle MUST have no Background set (D-422 bracket fallback drops bg fill)")
}

// TestCrumbChipFallbackStyle_FgIsColorFgANSI asserts the inactive
// bracket-fallback chip uses ColorFgANSI (lipgloss.ANSIColor(15) — bright
// white) so the chip text reads in the terminal's default fg color band
// on monochrome.
func TestCrumbChipFallbackStyle_FgIsColorFgANSI(t *testing.T) {
	got := ui.CrumbChipFallbackStyle.GetForeground()
	assert.Equal(t, ui.ColorFgANSI, got,
		"CrumbChipFallbackStyle MUST use Foreground(ColorFgANSI) per D-422")
}

// TestCrumbChipActiveFallbackStyle_NoBg asserts the active bracket-fallback
// chip has NO background set — recoloring bg under 4-bit downsample risks
// chip-vs-terminal collision (D-422).
func TestCrumbChipActiveFallbackStyle_NoBg(t *testing.T) {
	zero := lipgloss.NewStyle().GetBackground()
	got := ui.CrumbChipActiveFallbackStyle.GetBackground()
	assert.Equal(t, zero, got,
		"CrumbChipActiveFallbackStyle MUST have no Background set (D-422)")
}

// TestCrumbChipActiveFallbackStyle_NoFg asserts the active bracket-fallback
// chip has NO foreground set — recoloring fg under 4-bit downsample risks
// bg/fg collision against the terminal-default bg the user controls. The
// active-vs-inactive distinction is carried by Underline + Bold instead.
func TestCrumbChipActiveFallbackStyle_NoFg(t *testing.T) {
	zero := lipgloss.NewStyle().GetForeground()
	got := ui.CrumbChipActiveFallbackStyle.GetForeground()
	assert.Equal(t, zero, got,
		"CrumbChipActiveFallbackStyle MUST have no Foreground set (D-422 — decoration carries the cue)")
}

// TestCrumbChipActiveFallbackStyle_HasUnderline asserts the active
// bracket-fallback chip has Underline(true). Underline emits SGR 4 which
// survives every profile downsample including monochrome.
func TestCrumbChipActiveFallbackStyle_HasUnderline(t *testing.T) {
	assert.True(t, ui.CrumbChipActiveFallbackStyle.GetUnderline(),
		"CrumbChipActiveFallbackStyle MUST have Underline(true) per D-422")
}

// TestCrumbChipActiveFallbackStyle_HasBold asserts the active
// bracket-fallback chip has Bold(true). Bold emits SGR 1, redundant
// channel for colorblind safety per Pitfall 9.
func TestCrumbChipActiveFallbackStyle_HasBold(t *testing.T) {
	assert.True(t, ui.CrumbChipActiveFallbackStyle.GetBold(),
		"CrumbChipActiveFallbackStyle MUST have Bold(true) per D-422")
}

// TestCrumbChipFallbackStyle_RendersBracketContent asserts the rendered
// inactive chip retains the literal "<files>" content (visible after
// ansi.Strip).
func TestCrumbChipFallbackStyle_RendersBracketContent(t *testing.T) {
	rendered := ui.CrumbChipFallbackStyle.Render("<files>")
	stripped := ansi.Strip(rendered)
	assert.Contains(t, stripped, "<files>",
		"CrumbChipFallbackStyle.Render must preserve <segment> literal (D-422 + D-205)")
}

// TestCrumbChipActiveFallbackStyle_EmitsUnderlineSGR asserts the rendered
// active chip's raw output contains the SGR 4 (underline) byte sequence
// in either standalone "[4m" or combined ";4;" / ";4m" / "[4;" form.
// lipgloss/v2 may emit duplicate SGR params (e.g. "\x1b[1;4;4m") and may
// vary attribute ordering; the matcher set covers all observed shapes.
func TestCrumbChipActiveFallbackStyle_EmitsUnderlineSGR(t *testing.T) {
	rendered := ui.CrumbChipActiveFallbackStyle.Render("<prod.yaml>")
	assert.True(t,
		containsAnySGR(rendered, "\x1b[4m", "\x1b[4;", ";4m", ";4;"),
		"CrumbChipActiveFallbackStyle.Render MUST emit SGR 4 (underline) per D-422; got: %q", rendered)
}

// TestCrumbChipActiveFallbackStyle_EmitsBoldSGR asserts the rendered
// active chip's raw output contains the SGR 1 (bold) byte sequence in
// either standalone "[1m" or combined "[1;" / ";1m" / ";1;" form.
func TestCrumbChipActiveFallbackStyle_EmitsBoldSGR(t *testing.T) {
	rendered := ui.CrumbChipActiveFallbackStyle.Render("<prod.yaml>")
	assert.True(t,
		containsAnySGR(rendered, "\x1b[1m", "\x1b[1;", ";1m", ";1;"),
		"CrumbChipActiveFallbackStyle.Render MUST emit SGR 1 (bold) per D-422; got: %q", rendered)
}
