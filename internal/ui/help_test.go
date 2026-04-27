package ui_test

import (
	"strings"
	"testing"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compile-time interface compliance: HelpModel implements keys.Hinter.
var _ keys.Hinter = ui.HelpModel{}

// Test 12: View() with ViewFileList context includes file list keybindings.
func TestHelpViewFileListKeybindings(t *testing.T) {
	m := ui.NewHelpModel(120, 40)
	view := m.View(ui.ViewFileList)
	// The file list keymap includes "open" binding with "enter/l" key
	assert.True(t, strings.Contains(view, "open") || strings.Contains(view, "enter") || strings.Contains(view, "move up"),
		"ViewFileList help must include file list keybindings, got: %q", view)
}

// Test 13: View() with ViewDetail context includes detail keybindings.
func TestHelpViewDetailKeybindings(t *testing.T) {
	m := ui.NewHelpModel(120, 40)
	view := m.View(ui.ViewDetail)
	// The detail keymap includes "expand" and "collapse" bindings
	assert.True(t, strings.Contains(view, "expand") || strings.Contains(view, "collapse") || strings.Contains(view, "back"),
		"ViewDetail help must include detail keybindings, got: %q", view)
}

// Test 14: View() always includes global keybindings (?, q).
func TestHelpViewGlobalKeybindings(t *testing.T) {
	m := ui.NewHelpModel(120, 40)
	viewFileList := m.View(ui.ViewFileList)
	viewDetail := m.View(ui.ViewDetail)
	for _, view := range []string{viewFileList, viewDetail} {
		assert.True(t, strings.Contains(view, "quit") || strings.Contains(view, "toggle help"),
			"Help must always include global keybindings (?, q), got: %q", view)
	}
}

// Test 15: View() renders with RoundedBorder and surface background.
func TestHelpViewHasBorderOrSurface(t *testing.T) {
	m := ui.NewHelpModel(120, 40)
	view := m.View(ui.ViewFileList)
	// The view should be non-empty and have some structure
	assert.NotEmpty(t, view, "Help view must not be empty")
	// Rounded border produces corner characters like ╭ or ╰
	hasBorder := strings.ContainsAny(view, "╭╮╯╰│─")
	assert.True(t, hasBorder, "Help view must render with rounded border characters, got: %q", view)
}

// Test 16: View() footer contains "Press ? or Esc to close".
func TestHelpViewFooterText(t *testing.T) {
	m := ui.NewHelpModel(120, 40)
	view := m.View(ui.ViewFileList)
	assert.True(t, strings.Contains(view, "Press ? or Esc to close"),
		"Help view must contain footer 'Press ? or Esc to close', got: %q", view)
}

// TestHelpHints verifies HelpModel.Hints() returns the 3-hint persistent
// menu set per D-09: Esc, ?, q.
func TestHelpHints(t *testing.T) {
	m := ui.NewHelpModel(80, 24)
	hints := m.Hints()
	require.Equal(t, 3, len(hints), "Help must expose 3 hints")
	assert.Equal(t, "Esc", hints[0].Mnemonic)
	assert.Equal(t, "close help", hints[0].Description)
	assert.Equal(t, "?", hints[1].Mnemonic)
	assert.Equal(t, "q", hints[2].Mnemonic)
	for i, h := range hints {
		assert.True(t, h.Visible, "hint %d must default Visible=true", i)
	}
}
