// Package ui provides the help overlay component for sops-tui.
//
// HelpModel wraps charm.land/bubbles/v2/help.Model and renders a full-screen
// overlay with contextual keybindings based on the active view state.
//
// Per D-08: help is a full-screen overlay, not a footer strip.
// Per D-09: help content is contextual — file list vs detail keybindings differ.
// Per UI-SPEC Help Overlay: RoundedBorder, surface background, "Press ? or Esc to close" footer.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/lipgloss/v2"
	"github.com/caesarakalaeii/sops-tui/internal/keys"
)

// ViewState identifies which view is currently active, so HelpModel can
// render the correct contextual keybindings.
type ViewState int

const (
	// ViewFileList indicates the file browser is the active pane.
	ViewFileList ViewState = iota
	// ViewDetail indicates the YAML tree detail view is the active pane.
	ViewDetail
	// ViewMetadata indicates the metadata overlay is the active pane.
	// Falls through to ViewFileList for help rendering (metadata has minimal bindings).
	ViewMetadata
)

// HelpModel renders a full-screen keybinding overlay.
// It wraps bubbles/help.Model and adds the sops-tui styled border and footer.
type HelpModel struct {
	help   help.Model
	width  int
	height int
}

// NewHelpModel creates a HelpModel sized to the given dimensions.
// ShowAll is set to true per D-08 (full-screen means all bindings are visible).
func NewHelpModel(width, height int) HelpModel {
	h := help.New()
	h.ShowAll = true
	// Apply our design system styles to the help widget
	h.Styles.FullKey = HelpKeyStyle
	h.Styles.FullDesc = HelpDescStyle
	h.Styles.FullSeparator = HelpDescStyle
	return HelpModel{
		help:   h,
		width:  width,
		height: height,
	}
}

// SetSize updates the component dimensions.
func (m *HelpModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// View renders the full-screen help overlay for the given view state.
// The keybindings shown depend on fromState:
//   - ViewFileList: FileListKeyMap bindings
//   - ViewDetail: DetailKeyMap bindings
//
// Both always include global bindings (?, q) because they are embedded in
// the respective KeyMap types.
func (m HelpModel) View(fromState ViewState) string {
	var km help.KeyMap
	switch fromState {
	case ViewDetail:
		km = keys.DefaultDetailKeyMap
	default:
		km = keys.DefaultFileListKeyMap
	}

	content := m.help.View(km)

	// Footer per UI-SPEC Copywriting Contract
	footer := lipgloss.NewStyle().
		Foreground(ColorMuted).
		Render("Press ? or Esc to close")

	inner := content + "\n\n" + footer

	// Full-screen bordered box per UI-SPEC Help Overlay
	boxWidth := m.width - 2
	if boxWidth < 1 {
		boxWidth = 1
	}
	boxHeight := m.height - 2
	if boxHeight < 1 {
		boxHeight = 1
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorMuted).
		Background(ColorSurface).
		Padding(1, SpaceMD).
		Width(boxWidth).
		Height(boxHeight).
		Render(inner)
}

// Hints returns the 3-hint persistent menu set for HelpModel per D-09.
// The full help reference is the view itself (UI-11 retains the ?
// overlay as the complete reference); the persistent menu just shows
// how to close and quit.
func (m HelpModel) Hints() []keys.MenuHint {
	return []keys.MenuHint{
		{Mnemonic: "Esc", Description: "close help", Visible: true},
		{Mnemonic: "?", Description: "close help", Visible: true},
		{Mnemonic: "q", Description: "quit", Visible: true},
	}
}
