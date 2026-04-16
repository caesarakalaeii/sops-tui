// Package keys defines keybinding contracts for all Phase 1 views in sops-tui.
//
// Three KeyMap types are defined:
//   - GlobalKeyMap: keys available in every view (help, quit)
//   - FileListKeyMap: keys for the file list / browser view
//   - DetailKeyMap: keys for the YAML tree detail view
//
// Both FileListKeyMap and DetailKeyMap embed GlobalKeyMap and implement
// help.KeyMap from charm.land/bubbles/v2/help, enabling contextual help overlays.
//
// Per NAV-03: keybindings implement hjkl, g/G, ctrl-d/u navigation.
// Per D-09: global keys (?, q) appear in every context via embedding.
package keys

import (
	"charm.land/bubbles/v2/key"
)

// GlobalKeyMap holds keys available in every view.
// It is embedded in FileListKeyMap and DetailKeyMap.
type GlobalKeyMap struct {
	// Help toggles the full-screen help overlay.
	Help key.Binding
	// Quit exits the application.
	Quit key.Binding
}

// DefaultGlobalKeyMap is the default instance of GlobalKeyMap.
var DefaultGlobalKeyMap = GlobalKeyMap{
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// FileListKeyMap holds keybindings for the file list view.
// It embeds GlobalKeyMap so global keys are available everywhere.
// Implements help.KeyMap via ShortHelp() and FullHelp().
type FileListKeyMap struct {
	GlobalKeyMap

	// Up moves the selection up.
	Up key.Binding
	// Down moves the selection down.
	Down key.Binding
	// GoTop jumps to the first item.
	GoTop key.Binding
	// GoBottom jumps to the last item.
	GoBottom key.Binding
	// HalfUp scrolls half a page up.
	HalfUp key.Binding
	// HalfDown scrolls half a page down.
	HalfDown key.Binding
	// Open opens the selected file (navigates to detail view).
	Open key.Binding
	// Search activates the inline fuzzy filter.
	Search key.Binding
	// Info opens the metadata overlay for the highlighted file.
	Info key.Binding
	// ToggleSelect toggles the selection state of the highlighted file (D-05).
	ToggleSelect key.Binding
	// BulkReKey triggers bulk re-key on all selected files (D-05).
	BulkReKey key.Binding
	// HealthCheck triggers the on-demand secret health check (D-11).
	HealthCheck key.Binding
}

// ShortHelp returns a concise set of bindings shown in the collapsed help footer.
// Implements help.KeyMap.
func (k FileListKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Open, k.Search, k.Info, k.ToggleSelect, k.BulkReKey, k.HealthCheck, k.Help, k.Quit}
}

// FullHelp returns grouped bindings for the expanded help overlay.
// Groups: navigation, actions, global. Implements help.KeyMap.
func (k FileListKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.GoTop, k.GoBottom, k.HalfUp, k.HalfDown},
		{k.Open, k.Search, k.Info, k.ToggleSelect, k.BulkReKey, k.HealthCheck},
		{k.Help, k.Quit},
	}
}

// DefaultFileListKeyMap is the default instance of FileListKeyMap with
// vim-style navigation per 01-UI-SPEC.md §Keybindings.
var DefaultFileListKeyMap = FileListKeyMap{
	GlobalKeyMap: DefaultGlobalKeyMap,
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/↑", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/↓", "move down"),
	),
	GoTop: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "go to top"),
	),
	GoBottom: key.NewBinding(
		key.WithKeys("G"),
		key.WithHelp("G", "go to bottom"),
	),
	HalfUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "half page up"),
	),
	HalfDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "half page down"),
	),
	Open: key.NewBinding(
		key.WithKeys("enter", "l"),
		key.WithHelp("enter/l", "open"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Info: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "file info"),
	),
	ToggleSelect: key.NewBinding(
		key.WithKeys("space"),
		key.WithHelp("space", "toggle select"),
	),
	BulkReKey: key.NewBinding(
		key.WithKeys("K"),
		key.WithHelp("K", "bulk re-key selected"),
	),
	HealthCheck: key.NewBinding(
		key.WithKeys("H"),
		key.WithHelp("H", "health check"),
	),
}

// DetailKeyMap holds keybindings for the YAML tree detail view.
// It embeds GlobalKeyMap so global keys are available everywhere.
// Implements help.KeyMap via ShortHelp() and FullHelp().
type DetailKeyMap struct {
	GlobalKeyMap

	// Up moves the selection up.
	Up key.Binding
	// Down moves the selection down.
	Down key.Binding
	// GoTop jumps to the first item.
	GoTop key.Binding
	// GoBottom jumps to the last item.
	GoBottom key.Binding
	// HalfUp scrolls half a page up.
	HalfUp key.Binding
	// HalfDown scrolls half a page down.
	HalfDown key.Binding
	// Expand expands a collapsed tree node.
	Expand key.Binding
	// Collapse collapses an expanded tree node.
	Collapse key.Binding
	// Back returns to the file list view.
	Back key.Binding
	// Search activates the inline fuzzy filter for key paths.
	Search key.Binding
	// Info opens the metadata overlay for the current file.
	Info key.Binding
	// Reveal decrypts and reveals the selected encrypted value inline (r = toggle reveal/mask).
	Reveal key.Binding
	// RevealAll decrypts and reveals all values in the current file (R = toggle reveal-all/mask-all).
	RevealAll key.Binding
	// Edit enters inline edit mode on the selected revealed value.
	Edit key.Binding
	// EditFile suspends the TUI and opens the decrypted file in $EDITOR.
	EditFile key.Binding
	// Rotate generates a format-aware random replacement value for the selected leaf.
	Rotate key.Binding
	// Copy copies the selected revealed value to clipboard.
	Copy key.Binding
	// Blame opens the git history overlay for the current file.
	Blame key.Binding
	// AddRecipient opens the add-recipient modal for the current file (RCP-02).
	AddRecipient key.Binding
	// RemoveRecipient opens the remove-recipient list for the current file (RCP-02).
	RemoveRecipient key.Binding
}

// ShortHelp returns a concise set of bindings shown in the collapsed help footer.
// Implements help.KeyMap.
func (k DetailKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Reveal, k.RevealAll, k.Edit, k.Back, k.Search, k.Help, k.Quit, k.Copy, k.Blame, k.AddRecipient, k.RemoveRecipient}
}

// FullHelp returns grouped bindings for the expanded help overlay.
// Groups: navigation, tree actions, secret actions, global. Implements help.KeyMap.
func (k DetailKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.GoTop, k.GoBottom, k.HalfUp, k.HalfDown},
		{k.Expand, k.Collapse, k.Back, k.Search, k.Info, k.Blame},
		{k.Reveal, k.RevealAll, k.Edit, k.EditFile, k.Rotate, k.Copy, k.AddRecipient, k.RemoveRecipient},
		{k.Help, k.Quit},
	}
}

// DefaultDetailKeyMap is the default instance of DetailKeyMap with
// vim-style navigation per 01-UI-SPEC.md §Keybindings.
var DefaultDetailKeyMap = DetailKeyMap{
	GlobalKeyMap: DefaultGlobalKeyMap,
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/↑", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/↓", "move down"),
	),
	GoTop: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "go to top"),
	),
	GoBottom: key.NewBinding(
		key.WithKeys("G"),
		key.WithHelp("G", "go to bottom"),
	),
	HalfUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "half page up"),
	),
	HalfDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "half page down"),
	),
	Expand: key.NewBinding(
		key.WithKeys("enter", "l", "right"),
		key.WithHelp("enter/l", "expand"),
	),
	Collapse: key.NewBinding(
		key.WithKeys("h", "left"),
		key.WithHelp("h/←", "collapse"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back to file list"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Info: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "file info"),
	),
	Reveal: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reveal/hide value"),
	),
	RevealAll: key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "reveal/hide all values"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit value"),
	),
	EditFile: key.NewBinding(
		key.WithKeys("E"),
		key.WithHelp("E", "edit in $EDITOR"),
	),
	Rotate: key.NewBinding(
		key.WithKeys("X"),
		key.WithHelp("X", "rotate secret"),
	),
	Copy: key.NewBinding(
		key.WithKeys("ctrl+y"),
		key.WithHelp("ctrl+y", "copy to clipboard"),
	),
	Blame: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "git history"),
	),
	AddRecipient: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add recipient"),
	),
	RemoveRecipient: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "remove recipient"),
	),
}
