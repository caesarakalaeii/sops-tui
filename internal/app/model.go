// Package app provides the root Bubble Tea model for sops-tui.
//
// AppModel owns the sessionState enum that routes all messages and view
// rendering to the correct child component. It composes FileListModel,
// DetailModel, HelpModel, and StatusBarModel from the ui package.
//
// Navigation flow:
//   - stateFileList: default entry view; Enter/l drills into detail
//   - stateDetail: YAML tree view; Esc returns to file list
//   - stateHelp: full-screen overlay; ? or Esc returns to previous state
//
// Per RESEARCH.md Pattern 1: Root Model with sessionState Enum.
// Per CLAUDE.md: View() returns tea.View, use tea.KeyPressMsg, v.AltScreen = true.
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package app

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

// sessionState controls which child model receives updates and renders.
type sessionState int

const (
	// stateFileList is the default view — the file browser pane.
	stateFileList sessionState = iota
	// stateDetail is the YAML tree detail view for a selected file.
	stateDetail
	// stateHelp is the full-screen contextual keybinding overlay.
	stateHelp
)

// AppModel is the root Bubble Tea model. It owns a sessionState enum and
// holds all child models as value fields. Methods return updated copies.
type AppModel struct {
	state     sessionState
	prevState sessionState // restored when help overlay closes
	width     int
	height    int
	fileList  ui.FileListModel
	detail    ui.DetailModel
	help      ui.HelpModel
	status    ui.StatusBarModel
}

// NewAppModel constructs the initial AppModel.
// env provides startup validation results for the status bar indicators.
// The file list is initially empty — Phase 2 wires file discovery.
func NewAppModel(env ui.EnvStatus) AppModel {
	m := AppModel{
		state:    stateFileList,
		fileList: ui.NewFileListModel([]ui.FileItem{}, 0, 0),
		detail:   ui.NewDetailModel("", []ui.TreeNode{}, 0, 0),
		help:     ui.NewHelpModel(0, 0),
		status:   ui.NewStatusBarModel(env),
	}
	m.status.SetBreadcrumb("files")
	m.status.SetItemCount(0, "items")
	return m
}

// Init requests the current terminal size. WindowSizeMsgs are also delivered
// automatically when the program starts and on resize, but this ensures the
// initial size is captured immediately.
func (m AppModel) Init() tea.Cmd {
	return func() tea.Msg {
		return tea.RequestWindowSize()
	}
}

// Update dispatches incoming messages.
// Order of precedence:
//  1. WindowSizeMsg: propagate dimensions to all children
//  2. KeyPressMsg: global keys first (?, q, esc), then route to active child
//  3. All other messages: update status bar (for flash timer), then route to child
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Calculate main area height: full height minus status bar height
		statusBar := m.status.View(m.width)
		statusBarH := lipgloss.Height(statusBar)
		mainH := m.height - statusBarH
		if mainH < 0 {
			mainH = 0
		}
		// Propagate to all children that need dimensions (Pitfall 5)
		m.fileList.SetSize(m.width, mainH)
		m.detail.SetSize(m.width, mainH)
		m.help.SetSize(m.width, mainH)
		return m, nil

	case tea.KeyPressMsg:
		// Global key: toggle help overlay
		if key.Matches(msg, keys.DefaultGlobalKeyMap.Help) {
			if m.state == stateHelp {
				m.state = m.prevState
			} else {
				m.prevState = m.state
				m.state = stateHelp
			}
			return m, nil
		}
		// Global key: quit application
		if key.Matches(msg, keys.DefaultGlobalKeyMap.Quit) {
			return m, func() tea.Msg { return tea.Quit() }
		}
		// Esc: close help or return from detail to file list
		if msg.String() == "esc" {
			if m.state == stateHelp {
				m.state = m.prevState
				return m, nil
			} else if m.state == stateDetail {
				m.state = stateFileList
				m.status.SetBreadcrumb("files")
				m.status.SetItemCount(m.fileList.ItemCount(), "items")
				return m, nil
			}
		}
		// Enter/l in file list: drill into detail
		if m.state == stateFileList {
			if msg.String() == "enter" || msg.String() == "l" {
				if item, ok := m.fileList.SelectedItem(); ok {
					m.detail = ui.NewDetailModel(item.Name, []ui.TreeNode{}, m.width, m.height)
					m.state = stateDetail
					m.status.SetBreadcrumb("files", item.Name)
					m.status.SetItemCount(0, "keys")
					return m, nil
				}
			}
		}
	}

	// Always update status bar so flash timer messages are processed
	var statusCmd tea.Cmd
	m.status, statusCmd = m.status.Update(msg)

	// Route remaining messages to active child model
	switch m.state {
	case stateFileList:
		var cmd tea.Cmd
		m.fileList, cmd = m.fileList.Update(msg)
		return m, tea.Batch(cmd, statusCmd)
	case stateDetail:
		var cmd tea.Cmd
		m.detail, cmd = m.detail.Update(msg)
		return m, tea.Batch(cmd, statusCmd)
	case stateHelp:
		// Help is display-only; no child update needed
		return m, statusCmd
	}

	return m, statusCmd
}

// View composes the terminal frame: main content stacked above status bar.
// AltScreen is always true per the design decision for a full-screen TUI.
func (m AppModel) View() tea.View {
	// Render main content based on active state
	var content string
	switch m.state {
	case stateFileList:
		content = m.fileList.View()
	case stateDetail:
		content = m.detail.View()
	case stateHelp:
		// Contextual: use prevState so the overlay knows which keybindings to show
		content = m.help.View(ui.ViewState(m.prevState))
	}

	// Render status bar
	statusBar := m.status.View(m.width)

	// Calculate available height for main content
	statusBarH := lipgloss.Height(statusBar)
	mainH := m.height - statusBarH
	if mainH < 0 {
		mainH = 0
	}

	// Pad or constrain main content to fill available height
	body := lipgloss.NewStyle().Height(mainH).Render(content)

	// Stack body above status bar
	full := lipgloss.JoinVertical(lipgloss.Left, body, statusBar)

	v := tea.NewView(full)
	v.AltScreen = true
	return v
}
