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
	// GitStatus is the git worktree status code: "M", "A", "?", or "" for clean/no-git (D-09).
	GitStatus string
	// Selected indicates whether this file is selected for bulk operations (D-05).
	Selected bool
}

// Title returns the display name used by the default list delegate.
// Prepends [+] indicator when Selected is true (D-05).
// Appends [unencrypted] badge when IsEncrypted is false.
// Appends git badge [M]/[A]/[?] when the file has uncommitted changes (D-09).
// Implements list.DefaultItem.
func (i FileItem) Title() string {
	base := i.Name
	if i.Selected {
		base = SelectionIndicatorStyle.Render("[+]") + " " + base
	}
	if !i.IsEncrypted {
		base += " " + BadgeUnencrypted.Render("[unencrypted]")
	}
	switch i.GitStatus {
	case "M":
		base += " " + BadgeModified.Render("[M]")
	case "A":
		base += " " + BadgeAdded.Render("[A]")
	case "?":
		base += " " + BadgeUntracked.Render("[?]")
	}
	return base
}

// Description returns the absolute path shown as a secondary line.
// Implements list.DefaultItem.
func (i FileItem) Description() string { return i.Path }

// FilterValue is the value used when fuzzy-filtering the list.
// Implements list.Item.
func (i FileItem) FilterValue() string { return i.Name }

// CrossFileListItem wraps a cross-file search result for display in the list.
// It implements list.Item and list.DefaultItem from charm.land/bubbles/v2/list.
type CrossFileListItem struct {
	DisplayTitle string // "filename > key.path" or just "filename"
	Desc         string // absolute path or empty
	OrigIndex    int    // index into AppModel.crossFileItems
}

// Title returns the display title. Implements list.DefaultItem.
func (c CrossFileListItem) Title() string { return c.DisplayTitle }

// Description returns the description line. Implements list.DefaultItem.
func (c CrossFileListItem) Description() string { return c.Desc }

// FilterValue returns the value used for filtering. Implements list.Item.
func (c CrossFileListItem) FilterValue() string { return c.DisplayTitle }

// FileListModel wraps charm.land/bubbles/v2/list.Model and provides
// vim-style navigation (j/k/g/G/ctrl-d/u) via the custom FileListKeyMap.
// Supports inline fuzzy search via SearchModel.
type FileListModel struct {
	list            list.Model
	keys            keys.FileListKeyMap
	width           int
	height          int
	searchActive    bool
	search          SearchModel
	allItems        []FileItem // full unfiltered list
	crossFileMode   bool       // true when searching across files+keys
	crossFileTitles []string   // the "filename > key.path" strings for fuzzy matching
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

// ActivateCrossFileSearch activates cross-file search mode with the given title strings.
// titles contains "filename > key.path" (or just "filename" for file-level items).
// Returns the Focus cmd from the textinput.
func (m *FileListModel) ActivateCrossFileSearch(titles []string) tea.Cmd {
	m.searchActive = true
	m.crossFileMode = true
	m.crossFileTitles = titles
	return m.search.SetActive(true)
}

// DeactivateSearch deactivates search mode and restores the full item list.
func (m *FileListModel) DeactivateSearch() {
	m.searchActive = false
	m.crossFileMode = false
	m.crossFileTitles = nil
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

// IsCrossFileMode returns whether cross-file search mode is currently active.
func (m FileListModel) IsCrossFileMode() bool { return m.crossFileMode }

// SelectedCrossFileIndex returns the index into the cross-file items slice
// for the currently selected search result. Returns -1 if no selection.
func (m FileListModel) SelectedCrossFileIndex() int {
	item := m.list.SelectedItem()
	if item == nil {
		return -1
	}
	if cfi, ok := item.(CrossFileListItem); ok {
		return cfi.OrigIndex
	}
	return -1
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
				if m.crossFileMode {
					// In cross-file mode, Enter is handled by model.go for navigation.
					// Return without deactivating so model.go can read SelectedCrossFileIndex.
					return m, nil
				}
				m.searchActive = false
				m.search.Reset()
				return m, nil
			}
		}
		// Route all other messages to search model
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)

		if m.crossFileMode {
			// Cross-file search: filter against crossFileTitles
			pattern := m.search.Value()
			if pattern == "" {
				listItems := make([]list.Item, len(m.crossFileTitles))
				for i, t := range m.crossFileTitles {
					listItems[i] = CrossFileListItem{DisplayTitle: t, Desc: "", OrigIndex: i}
				}
				m.list.SetItems(listItems)
			} else {
				matches := ApplyFilter(pattern, m.crossFileTitles)
				filtered := make([]list.Item, 0, len(matches))
				for _, match := range matches {
					filtered = append(filtered, CrossFileListItem{
						DisplayTitle: m.crossFileTitles[match.Index],
						Desc:         "",
						OrigIndex:    match.Index,
					})
				}
				m.list.SetItems(filtered)
			}
			return m, cmd
		}

		// Normal file-list search: filter against allItems by name
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
		case key.Matches(msg, m.keys.ToggleSelect):
			// Toggle selection state on the highlighted item (D-05).
			if item, ok := m.SelectedItem(); ok {
				for idx := range m.allItems {
					if m.allItems[idx].Path == item.Path {
						m.allItems[idx].Selected = !m.allItems[idx].Selected
						break
					}
				}
				// Rebuild list items so Title() reflects the new Selected state.
				listItems := make([]list.Item, len(m.allItems))
				for idx, it := range m.allItems {
					listItems[idx] = it
				}
				m.list.SetItems(listItems)
			}
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

// Hints returns the persistent-menu hint set for FileListModel per D-09.
// Derives the first 10 hints from the keymap (single source of truth per
// D-08) and appends g/G explicitly since FileListKeyMap.ShortHelp()
// intentionally omits them (they are navigation micro-keys).
func (m FileListModel) Hints() []keys.MenuHint {
	hints := keys.HintsFromBindings(m.keys.ShortHelp())
	hints = append(hints,
		keys.MenuHint{Mnemonic: "g", Description: "go to top", Visible: true},
		keys.MenuHint{Mnemonic: "G", Description: "go to bottom", Visible: true},
	)
	return hints
}

// SelectedItems returns all FileItems with Selected == true.
// Used by AppModel for bulk re-key (D-05).
func (m FileListModel) SelectedItems() []FileItem {
	var selected []FileItem
	for _, item := range m.allItems {
		if item.Selected {
			selected = append(selected, item)
		}
	}
	return selected
}

// ClearSelections resets Selected to false on all items and rebuilds the list.
// Called after bulk re-key completes (D-05).
func (m *FileListModel) ClearSelections() {
	for idx := range m.allItems {
		m.allItems[idx].Selected = false
	}
	listItems := make([]list.Item, len(m.allItems))
	for idx, it := range m.allItems {
		listItems[idx] = it
	}
	m.list.SetItems(listItems)
}
