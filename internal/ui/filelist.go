// Package ui provides the view components for sops-tui.
// This file implements FileListModel, a Bubble Tea component that wraps
// charm.land/bubbles/v2/list for the file browser pane.
//
// Per NAV-03: j/k/g/G/ctrl-d/u vim-style navigation.
// Per D-05: file list is the entry point of the single-pane drill-down.
// Per 01-UI-SPEC.md §Copywriting Contract: empty state text is exact.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
	"github.com/caesarakalaeii/sops-tui/internal/sops"
)

// FileItem represents a SOPS-encrypted file entry in the file browser.
// It implements list.Item (FilterValue) and list.DefaultItem (Title, Description)
// from charm.land/bubbles/v2/list.
type FileItem struct {
	// Name is the filename relative to the repo root (e.g., "secrets/prod.yaml").
	Name string
	// Path is the absolute path for later use in decrypt/edit operations.
	Path string
	// IsEncrypted indicates whether the file contains a SOPS encryption marker.
	// Per D-02: false shows [unencrypted] badge.
	IsEncrypted bool
	// Rule is the matched .sops.yaml creation rule for this file (used by parser).
	Rule sops.CreationRule
}

// Title returns the display name used by the default list delegate.
// Appends [unencrypted] badge when IsEncrypted is false.
// Implements list.DefaultItem.
func (i FileItem) Title() string {
	if !i.IsEncrypted {
		return i.Name + " " + BadgeUnencrypted.Render("[unencrypted]")
	}
	return i.Name
}

// Description returns the absolute path shown as a secondary line.
// Implements list.DefaultItem.
func (i FileItem) Description() string { return i.Path }

// FilterValue is the value used when fuzzy-filtering the list.
// Implements list.Item.
func (i FileItem) FilterValue() string { return i.Name }

// FileListModel wraps charm.land/bubbles/v2/list.Model and provides
// vim-style navigation (j/k/g/G/ctrl-d/u) via the custom FileListKeyMap.
// Supports inline fuzzy search via SearchModel.
type FileListModel struct {
	list         list.Model
	keys         keys.FileListKeyMap
	width        int
	height       int
	searchActive bool
	search       SearchModel
	allItems     []FileItem // full unfiltered list
}

// NewFileListModel creates a FileListModel with the given items and dimensions.
// Items are displayed using the default bubbles delegate.
// Built-in help, status bar, and filter UI are disabled — sops-tui provides
// its own overlay (D-08) and status bar (D-10).
func NewFileListModel(items []FileItem, width, height int) FileListModel {
	listItems := make([]list.Item, len(items))
	for idx, item := range items {
		listItems[idx] = item
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(listItems, delegate, width, height)
	l.Title = "SOPS Files"

	// Disable built-in chrome — sops-tui owns these surfaces.
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetShowFilter(false)
	l.SetFilteringEnabled(false)

	return FileListModel{
		list:     l,
		keys:     keys.DefaultFileListKeyMap,
		width:    width,
		height:   height,
		allItems: items,
		search:   NewSearchModel(width),
	}
}

// ActivateSearch activates inline fuzzy search mode.
// Returns the Focus cmd from the textinput.
func (m *FileListModel) ActivateSearch() tea.Cmd {
	m.searchActive = true
	return m.search.SetActive(true)
}

// DeactivateSearch deactivates search mode and restores the full item list.
func (m *FileListModel) DeactivateSearch() {
	m.searchActive = false
	m.search.Reset()
	// Restore full item list
	listItems := make([]list.Item, len(m.allItems))
	for i, item := range m.allItems {
		listItems[i] = item
	}
	m.list.SetItems(listItems)
}

// IsSearchActive returns whether search mode is currently active.
func (m FileListModel) IsSearchActive() bool {
	return m.searchActive
}

// Update processes messages. Navigation keys (g/G/ctrl-d/u) are intercepted
// before delegating to the embedded list (which handles j/k/up/down).
// Per RESEARCH.md Open Questions Q1 RESOLVED: we do not rely on bubbles/list
// default KeyMap for g/G/ctrl-d/u — we handle them explicitly.
func (m FileListModel) Update(msg tea.Msg) (FileListModel, tea.Cmd) {
	if m.searchActive {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "esc":
				m.DeactivateSearch()
				return m, nil
			case "enter":
				m.searchActive = false
				m.search.Reset()
				return m, nil
			}
		}
		// Route all other messages to search model
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		// Apply filter to item list
		pattern := m.search.Value()
		if pattern == "" {
			listItems := make([]list.Item, len(m.allItems))
			for i, item := range m.allItems {
				listItems[i] = item
			}
			m.list.SetItems(listItems)
		} else {
			names := make([]string, len(m.allItems))
			for i, item := range m.allItems {
				names[i] = item.Name
			}
			matches := ApplyFilter(pattern, names)
			filtered := make([]list.Item, 0, len(matches))
			for _, match := range matches {
				filtered = append(filtered, m.allItems[match.Index])
			}
			m.list.SetItems(filtered)
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.GoTop):
			m.list.Select(0)
			return m, nil
		case key.Matches(msg, m.keys.GoBottom):
			last := len(m.list.Items()) - 1
			if last >= 0 {
				m.list.Select(last)
			}
			return m, nil
		case key.Matches(msg, m.keys.HalfUp):
			idx := m.list.Index() - m.height/2
			if idx < 0 {
				idx = 0
			}
			m.list.Select(idx)
			return m, nil
		case key.Matches(msg, m.keys.HalfDown):
			idx := m.list.Index() + m.height/2
			maxIdx := len(m.list.Items()) - 1
			if maxIdx < 0 {
				maxIdx = 0
			}
			if idx > maxIdx {
				idx = maxIdx
			}
			m.list.Select(idx)
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the file list. If there are no items, the empty state is shown
// per 01-UI-SPEC.md §Copywriting Contract.
// When search is active, the search bar is appended at the bottom.
func (m FileListModel) View() string {
	var content string
	if len(m.list.Items()) == 0 && !m.searchActive {
		lines := []string{
			"No SOPS files found",
			"",
			DimText.Render("No .sops.yaml discovered in this directory or parents."),
			DimText.Render("Run sops-tui in a repository with a .sops.yaml configuration."),
		}
		content = strings.Join(lines, "\n")
	} else {
		content = m.list.View()
	}

	if m.searchActive {
		return content + "\n" + m.search.View()
	}
	return content
}

// SetSize updates the component dimensions and propagates to the embedded list.
func (m *FileListModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.search.SetWidth(width)
	if m.searchActive {
		m.list.SetSize(width, height-1)
	} else {
		m.list.SetSize(width, height)
	}
}

// SelectedItem returns the currently selected FileItem and whether a selection exists.
// Returns (FileItem{}, false) if the list is empty or selection is invalid.
func (m FileListModel) SelectedItem() (FileItem, bool) {
	item := m.list.SelectedItem()
	if item == nil {
		return FileItem{}, false
	}
	fi, ok := item.(FileItem)
	return fi, ok
}

// SelectedFileItem returns the currently selected FileItem including Rule and IsEncrypted.
// Returns (FileItem{}, false) if the list is empty or selection is invalid.
func (m FileListModel) SelectedFileItem() (FileItem, bool) {
	item := m.list.SelectedItem()
	if item == nil {
		return FileItem{}, false
	}
	fi, ok := item.(FileItem)
	return fi, ok
}

// ItemCount returns the total number of items in the list.
func (m FileListModel) ItemCount() int {
	return len(m.allItems)
}
