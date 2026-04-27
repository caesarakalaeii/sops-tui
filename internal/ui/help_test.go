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

// Test 15: View() returns inner content (Phase 7.1 D-112: no inner border).
// Pre-Phase-7.1 this test asserted RoundedBorder corner characters appeared
// in HelpModel.View(); the inner-border envelope was stripped per WR-02 so
// only the outer WrapTitled NormalBorder remains (rendered by AppModel.View
// at model.go:1342). HelpModel.View() now returns inner content only.
func TestHelpViewHasNonEmptyContent(t *testing.T) {
	m := ui.NewHelpModel(120, 40)
	view := m.View(ui.ViewFileList)
	// The view should be non-empty and contain footer + at least one keybinding line.
	assert.NotEmpty(t, view, "Help view must not be empty")
	assert.True(t, strings.Contains(view, "Press ? or Esc to close"),
		"Help view must contain the close-prompt footer, got: %q", view)
	// Phase 7.1 D-112: HelpModel.View() must NOT carry an inner border —
	// the outer WrapTitled in AppModel.View() owns chrome borders.
	assert.False(t, strings.ContainsAny(view, "╭╮╯╰"),
		"HelpModel.View() must NOT emit RoundedBorder corners (single-border framing per WR-02), got: %q", view)
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
