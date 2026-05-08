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
//
// NOTE: In tree mode, FileListModel wraps each FileItem in a fileTreeItem
// which renders the basename + tree connectors. FileItem.Title() is kept for
// backward-compat callers (existing tests + cross-file flat rendering paths).
func (i FileItem) Title() string {
	base := i.Name
	if i.Selected {
		base = SelectionIndicatorStyle.Render("[+]") + " " + base
	}
	return base + fileBadges(i)
}

// Description returns the absolute path. Used only by the legacy flat rendering
// path; in tree mode the wrapper returns "" so rows are single-line.
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
//
// Items are rendered as a directory tree built from FileItem.Name. Collapse
// state lives in `collapsed` keyed by directory full-path so it survives
// rebuilds (toggles, selection changes, search activate/deactivate).
type FileListModel struct {
	list            list.Model
	keys            keys.FileListKeyMap
	width           int
	height          int
	searchActive    bool
	search          SearchModel
	allItems        []FileItem      // full unfiltered list
	tree            *treeNode       // rebuilt from allItems whenever it changes
	collapsed       map[string]bool // dirs currently collapsed (keyed by fullPath)
	savedCollapsed  map[string]bool // snapshot taken on filter activation
	filterPattern   string          // current filter pattern (empty when not filtering)
	crossFileMode   bool            // true when searching across files+keys
	crossFileTitles []string        // the "filename > key.path" strings for fuzzy matching
}

// NewFileListModel creates a FileListModel with the given items and dimensions.
// Items are rendered as a directory tree (dirs expanded by default).
// Built-in help, status bar, and filter UI are disabled — sops-tui provides
// its own overlay (D-08) and status bar (D-10).
func NewFileListModel(items []FileItem, width, height int) FileListModel {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false

	l := list.New(nil, delegate, width, height)
	l.Title = "SOPS Files"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetShowFilter(false)
	l.SetFilteringEnabled(false)

	m := FileListModel{
		list:      l,
		keys:      keys.DefaultFileListKeyMap,
		width:     width,
		height:    height,
		allItems:  items,
		collapsed: map[string]bool{},
		search:    NewSearchModel(width),
	}
	m.tree = buildTree(m.allItems)
	m.refreshRows()
	return m
}

// refreshRows recomputes the visible rows from the tree honoring `collapsed`
// and the active filter, then pushes them into the embedded list. Tries to
// preserve the cursor on the same row when possible.
func (m *FileListModel) refreshRows() {
	if m.tree == nil {
		m.tree = buildTree(m.allItems)
	}

	var match func(FileItem) bool
	if m.filterPattern != "" {
		matches := ApplyFilter(m.filterPattern, fileItemNames(m.allItems))
		matchSet := make(map[string]struct{}, len(matches))
		for _, mt := range matches {
			matchSet[m.allItems[mt.Index].Path] = struct{}{}
		}
		match = func(f FileItem) bool {
			_, ok := matchSet[f.Path]
			return ok
		}
	}

	// Snapshot current cursor so we can try to restore it after the rebuild.
	var prevSel list.Item
	if it := m.list.SelectedItem(); it != nil {
		prevSel = it
	}

	rows := flatten(m.tree, m.collapsed, match)
	m.list.SetItems(rows)

	// Try to restore the cursor on the same logical row (file by Path,
	// directory by fullPath).
	switch prev := prevSel.(type) {
	case fileTreeItem:
		idx := findFileTreeIndex(rows, prev.file)
		if idx >= 0 {
			m.list.Select(idx)
			return
		}
	case dirItem:
		for i, r := range rows {
			if d, ok := r.(dirItem); ok && d.fullPath == prev.fullPath {
				m.list.Select(i)
				return
			}
		}
	}

	// Fallback: select the first file (or first row if there are no files).
	if first := findFirstFileIndex(rows); first >= 0 {
		m.list.Select(first)
	} else if len(rows) > 0 {
		m.list.Select(0)
	}
}

// fileItemNames returns just the Name field of each item — used as the corpus
// for fuzzy filtering. ApplyFilter takes []string.
func fileItemNames(items []FileItem) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.Name
	}
	return out
}

// ActivateSearch activates inline fuzzy search mode.
// Returns the Focus cmd from the textinput.
func (m *FileListModel) ActivateSearch() tea.Cmd {
	m.searchActive = true
	// Snapshot expansion state so Esc restores exactly what was visible
	// before the search.
	m.savedCollapsed = copyBoolMap(m.collapsed)
	m.filterPattern = ""
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
	wasCrossFile := m.crossFileMode
	m.searchActive = false
	m.crossFileMode = false
	m.crossFileTitles = nil
	m.search.Reset()
	m.filterPattern = ""

	// Restore expansion snapshot from before the filter was active.
	if !wasCrossFile && m.savedCollapsed != nil {
		m.collapsed = copyBoolMap(m.savedCollapsed)
	}
	m.savedCollapsed = nil
	m.refreshRows()
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
//
// Tree-specific intercepts:
//   - Enter on a directory toggles its expand/collapse state.
//   - Space on a directory recursively selects/deselects every file beneath it.
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
				m.filterPattern = ""
				m.savedCollapsed = nil
				m.refreshRows()
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

		// Normal file-list search: filter against allItems by Name, dim
		// non-matching files, auto-expand ancestors of matches.
		newPattern := m.search.Value()
		if newPattern != m.filterPattern {
			m.filterPattern = newPattern
			m.expandAncestorsOfMatches()
			m.refreshRows()
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
			// Space on a directory: recursively toggle Selected on every file
			// underneath. Otherwise toggle the highlighted file.
			if it := m.list.SelectedItem(); it != nil {
				if d, ok := it.(dirItem); ok {
					m.toggleDirSelection(d.fullPath)
				} else if f, ok := it.(fileTreeItem); ok {
					m.toggleFileSelection(f.file.Path)
				}
				m.refreshRowsKeepCursor()
			}
			return m, nil
		}
		// Enter on a directory toggles expand/collapse. We don't consume
		// Enter on a file — AppModel needs to see it to navigate to detail.
		if msg.String() == "enter" || msg.String() == "l" {
			if it := m.list.SelectedItem(); it != nil {
				if d, ok := it.(dirItem); ok {
					if m.collapsed[d.fullPath] {
						delete(m.collapsed, d.fullPath)
					} else {
						m.collapsed[d.fullPath] = true
					}
					m.refreshRowsKeepCursor()
					return m, nil
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// refreshRowsKeepCursor rebuilds rows but preserves the cursor on the same
// row by Path/fullPath identity. Used after toggle operations where the
// cursor should stay where it is.
func (m *FileListModel) refreshRowsKeepCursor() {
	m.refreshRows()
}

// toggleFileSelection flips Selected on the FileItem with the given Path.
func (m *FileListModel) toggleFileSelection(path string) {
	for idx := range m.allItems {
		if m.allItems[idx].Path == path {
			m.allItems[idx].Selected = !m.allItems[idx].Selected
			return
		}
	}
}

// toggleDirSelection toggles Selected on every file under the given directory
// path. The new state is the inverse of the FIRST file's current state — so
// pressing Space on a partially-selected dir collapses to "all selected" if
// the first file was unselected, "all unselected" if it was selected. Empty
// dirs are no-ops.
func (m *FileListModel) toggleDirSelection(dirPath string) {
	files := collectDirFiles(m.tree, dirPath)
	if len(files) == 0 {
		return
	}
	target := !files[0].Selected
	want := make(map[string]struct{}, len(files))
	for _, f := range files {
		want[f.Path] = struct{}{}
	}
	for idx := range m.allItems {
		if _, ok := want[m.allItems[idx].Path]; ok {
			m.allItems[idx].Selected = target
		}
	}
	// The tree caches FileItem snapshots from buildTree; rebuild so dim/
	// connector renderers see the new Selected state.
	m.tree = buildTree(m.allItems)
}

// expandAncestorsOfMatches uncollapses every ancestor of every matching file
// so matches are always visible. Called on filter pattern change.
func (m *FileListModel) expandAncestorsOfMatches() {
	if m.filterPattern == "" {
		return
	}
	matches := ApplyFilter(m.filterPattern, fileItemNames(m.allItems))
	for _, mt := range matches {
		dir := fileParentDir(m.allItems[mt.Index].Name)
		for _, anc := range ancestorPaths(dir) {
			delete(m.collapsed, anc)
		}
	}
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

// SelectedItem returns the currently highlighted FileItem and whether the
// cursor sits on a file (not a directory). Returns (FileItem{}, false) for
// directories, empty lists, or cross-file rows.
func (m FileListModel) SelectedItem() (FileItem, bool) {
	item := m.list.SelectedItem()
	if item == nil {
		return FileItem{}, false
	}
	if f, ok := item.(fileTreeItem); ok {
		return f.file, true
	}
	if fi, ok := item.(FileItem); ok {
		// Defensive fallback: if a caller plumbed a raw FileItem in (e.g.,
		// during cross-file flow restoration), still unwrap it.
		return fi, true
	}
	return FileItem{}, false
}

// SelectedFileItem returns the currently selected FileItem including Rule and IsEncrypted.
// Returns (FileItem{}, false) when the cursor sits on a directory or the list is empty.
func (m FileListModel) SelectedFileItem() (FileItem, bool) {
	return m.SelectedItem()
}

// ItemCount returns the total number of FILES in the list (directories excluded).
// Used by AppModel for the status-bar item count.
func (m FileListModel) ItemCount() int {
	return len(m.allItems)
}

// Hints returns the persistent-menu hint set for FileListModel per D-09.
// Derives all 12 hints from FileListKeyMap.ShortHelp() per D-304 and D-301
// (total derivation — no manual append). GoTop (g) and GoBottom (G) are
// now part of ShortHelp() at positions 10 and 11 respectively.
func (m FileListModel) Hints() []keys.MenuHint {
	return keys.HintsFromBindings(m.keys.ShortHelp())
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
	m.tree = buildTree(m.allItems)
	m.refreshRows()
}

// copyBoolMap clones a map[string]bool. Returns nil on a nil input.
func copyBoolMap(in map[string]bool) map[string]bool {
	if in == nil {
		return nil
	}
	out := make(map[string]bool, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
