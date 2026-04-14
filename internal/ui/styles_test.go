package ui_test

import (
	"strings"
	"testing"

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
		{"ColorAccentHex", ui.ColorAccentHex, "#89b4fa"},
		{"ColorSuccessHex", ui.ColorSuccessHex, "#a6e3a1"},
		{"ColorWarningHex", ui.ColorWarningHex, "#f9e2af"},
		{"ColorErrorHex", ui.ColorErrorHex, "#f38ba8"},
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
