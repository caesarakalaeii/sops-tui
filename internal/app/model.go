// Package app provides the root Bubble Tea model for sops-tui.
//
// AppModel owns the sessionState enum that routes all messages and view
// rendering to the correct child component. It composes FileListModel,
// DetailModel, HelpModel, StatusBarModel, and MetadataModel from the ui package.
//
// Navigation flow:
//   - stateFileList: default entry view; Enter/l drills into detail
//   - stateDetail: YAML tree view; Esc returns to file list
//   - stateHelp: full-screen overlay; ? or Esc returns to previous state
//   - stateMetadata: full-screen metadata overlay; i or Esc returns to previous state
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
	"github.com/caesarakalaeii/sops-tui/internal/parser"
	"github.com/caesarakalaeii/sops-tui/internal/sops"
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
	// stateMetadata is the full-screen SOPS metadata overlay.
	stateMetadata
)

// FilesDiscoveredMsg carries the result of SOPS file discovery.
type FilesDiscoveredMsg struct {
	Files []sops.DiscoveredFile
	Err   error
}

// FilesParsedMsg carries the result of parsing a single YAML file.
type FilesParsedMsg struct {
	Parsed parser.ParsedFile
	Err    error
}

// AppModel is the root Bubble Tea model. It owns a sessionState enum and
// holds all child models as value fields. Methods return updated copies.
type AppModel struct {
	state        sessionState
	prevState    sessionState // restored when help/metadata overlay closes
	width        int
	height       int
	fileList     ui.FileListModel
	detail       ui.DetailModel
	help         ui.HelpModel
	status       ui.StatusBarModel
	metadata     ui.MetadataModel
	sopsYamlPath string             // path to .sops.yaml
	files        []sops.DiscoveredFile // cached discovery results
	currentFile  sops.DiscoveredFile   // file currently shown in detail
}

// NewAppModel constructs the initial AppModel.
// env provides startup validation results for the status bar indicators.
// sopsYamlPath is the path to the .sops.yaml file (may be empty if not found).
// The file list is initially empty — discovery runs asynchronously in Init.
func NewAppModel(env ui.EnvStatus, sopsYamlPath string) AppModel {
	m := AppModel{
		state:        stateFileList,
		fileList:     ui.NewFileListModel([]ui.FileItem{}, 0, 0),
		detail:       ui.NewDetailModel("", []ui.TreeNode{}, 0, 0, true),
		help:         ui.NewHelpModel(0, 0),
		status:       ui.NewStatusBarModel(env),
		sopsYamlPath: sopsYamlPath,
	}
	m.status.SetBreadcrumb("files")
	m.status.SetItemCount(0, "items")
	return m
}

// Init requests the current terminal size and triggers async file discovery.
// WindowSizeMsgs are also delivered automatically when the program starts
// and on resize, but this ensures the initial size is captured immediately.
func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return tea.RequestWindowSize() },
		func() tea.Msg {
			if m.sopsYamlPath == "" {
				return FilesDiscoveredMsg{Err: nil}
			}
			files, err := sops.Discover(m.sopsYamlPath)
			return FilesDiscoveredMsg{Files: files, Err: err}
		},
	)
}

// Update dispatches incoming messages.
// Order of precedence:
//  1. WindowSizeMsg: propagate dimensions to all children
//  2. FilesDiscoveredMsg / FilesParsedMsg: async results
//  3. KeyPressMsg: global keys first (?, q, esc, i, /), then route to active child
//  4. All other messages: update status bar (for flash timer), then route to child
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		mainH := m.height - statusBarHeight(m)
		if mainH < 0 {
			mainH = 0
		}
		// Propagate to all children that need dimensions
		m.fileList.SetSize(m.width, mainH)
		m.detail.SetSize(m.width, mainH)
		m.help.SetSize(m.width, mainH)
		m.metadata.SetSize(m.width, mainH)
		return m, nil

	case FilesDiscoveredMsg:
		if msg.Err != nil {
			m.status, _ = m.status.Flash("Error discovering files: " + msg.Err.Error())
			return m, nil
		}
		m.files = msg.Files
		items := make([]ui.FileItem, len(msg.Files))
		for i, f := range msg.Files {
			items[i] = ui.FileItem{
				Name:        f.Name,
				Path:        f.AbsPath,
				IsEncrypted: f.IsEncrypted,
				Rule:        f.Rule,
			}
		}
		mainH := m.height - statusBarHeight(m)
		if mainH < 0 {
			mainH = 0
		}
		m.fileList = ui.NewFileListModel(items, m.width, mainH)
		m.status.SetItemCount(len(items), "items")
		return m, nil

	case FilesParsedMsg:
		if msg.Err != nil {
			m.status, _ = m.status.Flash("Error parsing file: " + msg.Err.Error())
			m.state = stateFileList
			return m, nil
		}
		mainH := m.height - statusBarHeight(m)
		if mainH < 0 {
			mainH = 0
		}
		m.detail = ui.NewDetailModel(
			m.currentFile.Name,
			msg.Parsed.Nodes,
			m.width,
			mainH,
			m.currentFile.IsEncrypted,
		)
		m.state = stateDetail
		m.status.SetBreadcrumb("files", m.currentFile.Name)
		m.status.SetItemCount(countLeafNodes(msg.Parsed.Nodes), "keys")
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

		// i key: toggle metadata overlay
		if key.Matches(msg, keys.DefaultFileListKeyMap.Info) || key.Matches(msg, keys.DefaultDetailKeyMap.Info) {
			if m.state == stateMetadata {
				m.state = m.prevState
				// Restore status bar based on prevState
				if m.prevState == stateFileList {
					m.status.SetBreadcrumb("files")
					m.status.SetItemCount(m.fileList.ItemCount(), "items")
				} else if m.prevState == stateDetail {
					m.status.SetBreadcrumb("files", m.currentFile.Name)
					m.status.SetItemCount(countLeafNodes(m.detail.Nodes()), "keys")
				}
				return m, nil
			}
			if (m.state == stateFileList && !m.fileList.IsSearchActive()) ||
				(m.state == stateDetail && !m.detail.IsSearchActive()) {
				var filePath string
				var rule sops.CreationRule
				var isEnc bool
				if m.state == stateFileList {
					if item, ok := m.fileList.SelectedFileItem(); ok {
						filePath = item.Path
						rule = item.Rule
						isEnc = item.IsEncrypted
					}
				} else {
					filePath = m.currentFile.AbsPath
					rule = m.currentFile.Rule
					isEnc = m.currentFile.IsEncrypted
				}
				if filePath != "" {
					parsed, err := parser.ParseFile(filePath, rule, isEnc)
					if err != nil {
						m.status, _ = m.status.Flash("Error reading metadata: " + err.Error())
						return m, nil
					}
					meta := ui.MetadataContent{
						Version:          parsed.Metadata.Version,
						LastModified:     parsed.Metadata.LastModified,
						MAC:              parsed.Metadata.MAC,
						AgeRecipients:    parsed.Metadata.AgeRecipients,
						EncryptedRegex:   parsed.Metadata.EncryptedRegex,
						UnencryptedRegex: parsed.Metadata.UnencryptedRegex,
					}
					mainH := m.height - statusBarHeight(m)
					if mainH < 0 {
						mainH = 0
					}
					m.metadata = ui.NewMetadataModel(meta, m.width, mainH)
					m.prevState = m.state
					m.state = stateMetadata
					m.status.SetBreadcrumb("files", m.currentFile.Name, "metadata")
					m.status.SetItemCount(0, "")
				}
				return m, nil
			}
		}

		// / key: activate search (only when not already active)
		if msg.String() == "/" {
			if m.state == stateFileList && !m.fileList.IsSearchActive() {
				cmd := m.fileList.ActivateSearch()
				m.status.SetBreadcrumb("files", "search")
				return m, cmd
			}
			if m.state == stateDetail && !m.detail.IsSearchActive() {
				cmd := m.detail.ActivateSearch()
				m.status.SetBreadcrumb("files", m.currentFile.Name, "search")
				return m, cmd
			}
		}

		// Esc: priority chain
		if msg.String() == "esc" {
			// Priority 1: Close search if active
			if m.state == stateFileList && m.fileList.IsSearchActive() {
				m.fileList.DeactivateSearch()
				m.status.SetBreadcrumb("files")
				m.status.SetItemCount(m.fileList.ItemCount(), "items")
				return m, nil
			}
			if m.state == stateDetail && m.detail.IsSearchActive() {
				m.detail.DeactivateSearch()
				m.status.SetBreadcrumb("files", m.currentFile.Name)
				m.status.SetItemCount(countLeafNodes(m.detail.Nodes()), "keys")
				return m, nil
			}
			// Priority 2: Close overlays
			if m.state == stateHelp {
				m.state = m.prevState
				return m, nil
			}
			if m.state == stateMetadata {
				m.state = m.prevState
				return m, nil
			}
			// Priority 3: Navigate back from detail to file list
			if m.state == stateDetail {
				m.state = stateFileList
				m.status.SetBreadcrumb("files")
				m.status.SetItemCount(m.fileList.ItemCount(), "items")
				return m, nil
			}
		}

		// Enter/l in file list: parse file and drill into detail (async)
		if m.state == stateFileList && !m.fileList.IsSearchActive() {
			if msg.String() == "enter" || msg.String() == "l" {
				if item, ok := m.fileList.SelectedFileItem(); ok {
					m.currentFile = sops.DiscoveredFile{
						Name:        item.Name,
						AbsPath:     item.Path,
						IsEncrypted: item.IsEncrypted,
						Rule:        item.Rule,
					}
					return m, func() tea.Msg {
						parsed, err := parser.ParseFile(item.Path, item.Rule, item.IsEncrypted)
						return FilesParsedMsg{Parsed: parsed, Err: err}
					}
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
	case stateMetadata:
		// Handle j/k scrolling in metadata overlay
		if kMsg, ok := msg.(tea.KeyPressMsg); ok {
			if key.Matches(kMsg, keys.DefaultDetailKeyMap.Down) {
				m.metadata.ScrollDown()
			} else if key.Matches(kMsg, keys.DefaultDetailKeyMap.Up) {
				m.metadata.ScrollUp()
			}
		}
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
	case stateMetadata:
		content = m.metadata.View()
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

// statusBarHeight returns the rendered height of the status bar in terminal rows.
func statusBarHeight(m AppModel) int {
	statusBar := m.status.View(m.width)
	return lipgloss.Height(statusBar)
}

// countLeafNodes recursively counts all leaf nodes (nodes with no children).
func countLeafNodes(nodes []ui.TreeNode) int {
	count := 0
	for _, n := range nodes {
		if len(n.Children) == 0 {
			count++
		} else {
			count += countLeafNodes(n.Children)
		}
	}
	return count
}
