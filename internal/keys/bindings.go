// Package keys defines keybinding contracts for all views in sops-tui.
//
// KeyMap types defined:
//   - GlobalKeyMap: keys available in every view (help, quit)
//   - FileListKeyMap: keys for the file list / browser view
//   - DetailKeyMap: keys for the YAML tree detail view
//   - HelpKeyMap: keys for the full-screen ? help overlay
//   - DiffKeyMap: keys for the diff confirmation overlay
//   - HealthKeyMap: keys for the health check overlay
//   - HistoryKeyMap: keys for the git history overlay
//   - MetadataKeyMap: keys for the SOPS metadata overlay
//   - RecipientFormKeyMap: keys for the add-recipient modal
//   - FileListSearchKeyMap: keys for the file list search-active override
//   - RecipientConfirmKeyMap: keys for stateRecipientConfirm (embeds GlobalKeyMap)
//   - BulkReKeyConfirmKeyMap: keys for stateBulkReKeyConfirm (embeds GlobalKeyMap)
//   - RecipientListKeyMap: keys for stateRecipientList
//   - FormatMenuKeyMap: keys for stateFormatMenu
//
// All keymaps implement help.KeyMap from charm.land/bubbles/v2/help.
//
// Per NAV-03: keybindings implement hjkl, g/G, ctrl-d/u navigation.
// Per D-09: global keys (?, q) appear in every context via embedding.
// Per D-301 (Phase 9): every Hints() derives from keymap.ShortHelp() —
// the keymap is the single source of truth for menu content (closes SC5).
package keys

import (
	"charm.land/bubbles/v2/key"
)

// menuVisibilityOverrider is implemented by keymaps that suppress some
// bindings from the persistent menu via Visible=false while keeping the
// bindings discoverable in the ? full-screen overlay (D-303, D-307).
//
// Implementers: DetailKeyMap (suppresses Blame), RecipientConfirmKeyMap
// and BulkReKeyConfirmKeyMap (suppress Quit per UI-SPEC
// §confirm-flow-quit-suppression).
//
// Sub-model Hints() implementations and the Plan 2 drift detector apply
// this suppression after deriving hints via HintsFromBindings(km.ShortHelp()).
type menuVisibilityOverrider interface {
	HiddenFromMenu() []key.Binding
}

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
// GoTop (g) and GoBottom (G) are appended after Quit per D-304 so that
// FileListModel.Hints() can reduce to a one-liner with no manual append.
// hints[10].Mnemonic == "g", hints[11].Mnemonic == "G" (TestFileListHints asserts this).
// Implements help.KeyMap.
func (k FileListKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Open, k.Search, k.Info, k.ToggleSelect, k.BulkReKey, k.HealthCheck, k.Help, k.Quit, k.GoTop, k.GoBottom}
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
	// AddSecret opens the add-secret modal to insert a new key/value into the current file.
	AddSecret key.Binding
}

// ShortHelp returns a concise set of bindings shown in the collapsed help footer.
// Implements help.KeyMap.
func (k DetailKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Reveal, k.RevealAll, k.Edit, k.Back, k.Search, k.Help, k.Quit, k.Copy, k.Blame, k.AddRecipient, k.RemoveRecipient, k.AddSecret}
}

// FullHelp returns grouped bindings for the expanded help overlay.
// Groups: navigation, tree actions, secret actions, global. Implements help.KeyMap.
func (k DetailKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.GoTop, k.GoBottom, k.HalfUp, k.HalfDown},
		{k.Expand, k.Collapse, k.Back, k.Search, k.Info, k.Blame},
		{k.Reveal, k.RevealAll, k.Edit, k.EditFile, k.Rotate, k.Copy, k.AddRecipient, k.RemoveRecipient, k.AddSecret},
		{k.Help, k.Quit},
	}
}

// HiddenFromMenu reports bindings that participate in the ? full-screen overlay
// (FullHelp) but are suppressed from the persistent menu. Blame is the
// canonical example — the 14-binding ShortHelp exceeds the 12-slot menu cap,
// so the least-frequently-used bindings are hidden (D-303, D-307). AddSecret
// (n) joins Blame (b) as hidden: both stay discoverable via the ? overlay.
func (k DetailKeyMap) HiddenFromMenu() []key.Binding {
	return []key.Binding{k.Blame, k.AddSecret}
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
	AddSecret: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "add secret"),
	),
}

// HelpKeyMap holds keybindings for the full-screen ? help overlay.
// Implements help.KeyMap via ShortHelp() and FullHelp().
// No GlobalKeyMap embedding — Help has its own Quit binding for explicit ShortHelp ordering.
type HelpKeyMap struct {
	// Close closes the help overlay (esc).
	Close key.Binding
	// ToggleHelp toggles the help overlay (?).
	ToggleHelp key.Binding
	// Quit exits the application (q, ctrl+c).
	Quit key.Binding
}

// ShortHelp returns the bindings rendered in the persistent menu for the help state.
// Implements help.KeyMap.
func (k HelpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Close, k.ToggleHelp, k.Quit}
}

// FullHelp returns grouped bindings for the expanded help overlay.
// Implements help.KeyMap. Single group — Help has only 3 bindings.
func (k HelpKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Close, k.ToggleHelp, k.Quit},
	}
}

// DefaultHelpKeyMap is the default instance with description strings matching
// the literal hint values returned before Phase 9 (description-string lock — TestHelpHints stays green).
var DefaultHelpKeyMap = HelpKeyMap{
	Close: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("Esc", "close help"),
	),
	ToggleHelp: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "close help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// DiffKeyMap holds keybindings for the diff confirmation overlay.
// Implements help.KeyMap via ShortHelp() and FullHelp().
type DiffKeyMap struct {
	// Confirm accepts the diff and confirms re-encryption (y).
	Confirm key.Binding
	// Cancel cancels without re-encrypting (n).
	Cancel key.Binding
	// Close closes the diff overlay (esc).
	Close key.Binding
	// ScrollDown scrolls the diff content down (j, down).
	ScrollDown key.Binding
	// ScrollUp scrolls the diff content up (k, up).
	ScrollUp key.Binding
	// Quit exits the application (q, ctrl+c).
	Quit key.Binding
}

// ShortHelp returns the bindings rendered in the persistent menu for the diff state.
// Implements help.KeyMap.
func (k DiffKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Confirm, k.Cancel, k.Close, k.ScrollDown, k.ScrollUp, k.Quit}
}

// FullHelp returns grouped bindings for the expanded help overlay.
// Implements help.KeyMap. Two groups: confirm/cancel, scroll/quit.
func (k DiffKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Confirm, k.Cancel, k.Close},
		{k.ScrollDown, k.ScrollUp, k.Quit},
	}
}

// DefaultDiffKeyMap is the default instance with description strings matching
// the literal hint values returned before Phase 9 (description-string lock — TestDiffHints stays green).
var DefaultDiffKeyMap = DiffKeyMap{
	Confirm: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "confirm re-encrypt"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "cancel"),
	),
	Close: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("Esc", "cancel"),
	),
	ScrollDown: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j", "scroll down"),
	),
	ScrollUp: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k", "scroll up"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// HealthKeyMap holds keybindings for the health check overlay.
// Implements help.KeyMap via ShortHelp() and FullHelp().
type HealthKeyMap struct {
	// ScrollDown scrolls the health content down (j, down).
	ScrollDown key.Binding
	// ScrollUp scrolls the health content up (k, up).
	ScrollUp key.Binding
	// Close closes the health overlay (H).
	Close key.Binding
	// CloseAlt closes the health overlay via Esc.
	CloseAlt key.Binding
	// Quit exits the application (q, ctrl+c).
	Quit key.Binding
}

// ShortHelp returns the bindings rendered in the persistent menu for the health state.
// Implements help.KeyMap.
func (k HealthKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ScrollDown, k.ScrollUp, k.Close, k.CloseAlt, k.Quit}
}

// FullHelp returns grouped bindings for the expanded help overlay.
// Implements help.KeyMap. Two groups: scroll, close/quit.
func (k HealthKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ScrollDown, k.ScrollUp},
		{k.Close, k.CloseAlt, k.Quit},
	}
}

// DefaultHealthKeyMap is the default instance with description strings matching
// the literal hint values returned before Phase 9 (description-string lock — TestHealthHints stays green).
var DefaultHealthKeyMap = HealthKeyMap{
	ScrollDown: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j", "scroll down"),
	),
	ScrollUp: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k", "scroll up"),
	),
	Close: key.NewBinding(
		key.WithKeys("H"),
		key.WithHelp("H", "close health"),
	),
	CloseAlt: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("Esc", "close health"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// HistoryKeyMap holds keybindings for the git history overlay.
// Implements help.KeyMap via ShortHelp() and FullHelp().
type HistoryKeyMap struct {
	// ScrollDown scrolls the history content down (j, down).
	ScrollDown key.Binding
	// ScrollUp scrolls the history content up (k, up).
	ScrollUp key.Binding
	// Close closes the history overlay (b).
	Close key.Binding
	// CloseAlt closes the history overlay via Esc.
	CloseAlt key.Binding
	// Quit exits the application (q, ctrl+c).
	Quit key.Binding
}

// ShortHelp returns the bindings rendered in the persistent menu for the history state.
// Implements help.KeyMap.
func (k HistoryKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ScrollDown, k.ScrollUp, k.Close, k.CloseAlt, k.Quit}
}

// FullHelp returns grouped bindings for the expanded help overlay.
// Implements help.KeyMap. Two groups: scroll, close/quit.
func (k HistoryKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ScrollDown, k.ScrollUp},
		{k.Close, k.CloseAlt, k.Quit},
	}
}

// DefaultHistoryKeyMap is the default instance with description strings matching
// the literal hint values returned before Phase 9 (description-string lock — TestHistoryHints stays green).
var DefaultHistoryKeyMap = HistoryKeyMap{
	ScrollDown: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j", "scroll down"),
	),
	ScrollUp: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k", "scroll up"),
	),
	Close: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "close history"),
	),
	CloseAlt: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("Esc", "close history"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// MetadataKeyMap holds keybindings for the SOPS metadata overlay.
// Implements help.KeyMap via ShortHelp() and FullHelp().
type MetadataKeyMap struct {
	// ScrollDown scrolls the metadata content down (j, down).
	ScrollDown key.Binding
	// ScrollUp scrolls the metadata content up (k, up).
	ScrollUp key.Binding
	// Close closes the metadata overlay (i).
	Close key.Binding
	// CloseAlt closes the metadata overlay via Esc.
	CloseAlt key.Binding
	// Quit exits the application (q, ctrl+c).
	Quit key.Binding
}

// ShortHelp returns the bindings rendered in the persistent menu for the metadata state.
// Implements help.KeyMap.
func (k MetadataKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ScrollDown, k.ScrollUp, k.Close, k.CloseAlt, k.Quit}
}

// FullHelp returns grouped bindings for the expanded help overlay.
// Implements help.KeyMap. Two groups: scroll, close/quit.
func (k MetadataKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ScrollDown, k.ScrollUp},
		{k.Close, k.CloseAlt, k.Quit},
	}
}

// DefaultMetadataKeyMap is the default instance with description strings matching
// the literal hint values returned before Phase 9 (description-string lock — TestMetadataHints stays green).
var DefaultMetadataKeyMap = MetadataKeyMap{
	ScrollDown: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j", "scroll down"),
	),
	ScrollUp: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k", "scroll up"),
	),
	Close: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "close metadata"),
	),
	CloseAlt: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("Esc", "close metadata"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// RecipientFormKeyMap holds keybindings for the add-recipient modal overlay.
// Implements help.KeyMap via ShortHelp() and FullHelp().
type RecipientFormKeyMap struct {
	// Confirm accepts the typed age public key (enter).
	Confirm key.Binding
	// Cancel cancels the add-recipient flow (esc).
	Cancel key.Binding
}

// ShortHelp returns the bindings rendered in the persistent menu for the recipient form state.
// Implements help.KeyMap.
func (k RecipientFormKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Confirm, k.Cancel}
}

// FullHelp returns grouped bindings for the expanded help overlay.
// Implements help.KeyMap. Single group — RecipientForm has only 2 bindings.
func (k RecipientFormKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Confirm, k.Cancel},
	}
}

// DefaultRecipientFormKeyMap is the default instance with description strings matching
// the literal hint values returned before Phase 9 (description-string lock — TestRecipientFormHints stays green).
var DefaultRecipientFormKeyMap = RecipientFormKeyMap{
	Confirm: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("Enter", "confirm"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("Esc", "cancel"),
	),
}

// AddSecretFormKeyMap holds keybindings for the add-secret modal overlay —
// a two-field form (key path + value) for inserting a new secret into the
// current file. Implements help.KeyMap via ShortHelp() and FullHelp().
type AddSecretFormKeyMap struct {
	// NextField moves focus between the key-path and value inputs (tab).
	NextField key.Binding
	// Confirm validates and submits the new key/value (enter).
	Confirm key.Binding
	// Cancel cancels the add-secret flow (esc).
	Cancel key.Binding
}

// ShortHelp returns the bindings rendered in the persistent menu for the add-secret form state.
// Implements help.KeyMap.
func (k AddSecretFormKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.NextField, k.Confirm, k.Cancel}
}

// FullHelp returns grouped bindings for the expanded help overlay.
// Implements help.KeyMap. Single group — AddSecretForm has only 3 bindings.
func (k AddSecretFormKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.NextField, k.Confirm, k.Cancel},
	}
}

// DefaultAddSecretFormKeyMap is the default instance of AddSecretFormKeyMap.
var DefaultAddSecretFormKeyMap = AddSecretFormKeyMap{
	NextField: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("Tab", "next field"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("Enter", "confirm"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("Esc", "cancel"),
	),
}

// FileListSearchKeyMap holds keybindings for the file list search-active override (D-11).
// Applied when FileListModel has an active inline search; replaces default FileList hints
// so the menu reflects keys that actually work in search mode.
// Implements help.KeyMap via ShortHelp() and FullHelp().
type FileListSearchKeyMap struct {
	// ExitSearch exits the search mode (esc).
	ExitSearch key.Binding
	// Select selects the highlighted search result (enter).
	Select key.Binding
	// NextResult moves to the next search result (j, down).
	NextResult key.Binding
	// PrevResult moves to the previous search result (k, up).
	PrevResult key.Binding
	// ToggleHelp toggles the help overlay (?).
	ToggleHelp key.Binding
	// Quit exits the application (q, ctrl+c).
	Quit key.Binding
}

// ShortHelp returns the bindings rendered in the persistent menu for the search-active state.
// Implements help.KeyMap.
func (k FileListSearchKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ExitSearch, k.Select, k.NextResult, k.PrevResult, k.ToggleHelp, k.Quit}
}

// FullHelp returns grouped bindings for the expanded help overlay.
// Implements help.KeyMap. Single group.
func (k FileListSearchKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ExitSearch, k.Select, k.NextResult, k.PrevResult, k.ToggleHelp, k.Quit},
	}
}

// DefaultFileListSearchKeyMap is the default instance with description strings matching
// the literal FileListSearchHints values from Phase 7 (description-string lock).
var DefaultFileListSearchKeyMap = FileListSearchKeyMap{
	ExitSearch: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("Esc", "exit search"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("Enter", "select result"),
	),
	NextResult: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j/↓", "next result"),
	),
	PrevResult: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k/↑", "prev result"),
	),
	ToggleHelp: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// RecipientConfirmKeyMap and BulkReKeyConfirmKeyMap suppress the Quit
// binding from the persistent menu (Visible=false) while keeping it
// bound and functional in AppModel.Update().
//
// Rationale (UI-SPEC §confirm-flow-quit-suppression): during a y/n
// confirmation gating a destructive operation (re-encryption, recipient
// change, bulk re-key per file), advertising the quit hint encourages
// bailing out without resolving the prompt. Quit is still bound (q,
// ctrl+c via embedded GlobalKeyMap) so users who memorize it still
// exit, but the menu nudges them toward y/n decisions.
//
// Compare DiffKeyMap.ShortHelp() (6 entries including q) — DiffModel
// shows quit because the diff body is non-destructive on its own;
// confirmation arms re-purpose the diff body for destructive flows
// and suppress quit accordingly (Phase 7 D-10 dispatcher disambiguation,
// Phase 9 D-313 formalization via HiddenFromMenu()).

// RecipientConfirmKeyMap holds keybindings for stateRecipientConfirm —
// the y/n confirmation over a shared diff body.
// Embeds GlobalKeyMap; Quit is suppressed from the persistent menu via HiddenFromMenu().
// Implements help.KeyMap via ShortHelp() and FullHelp().
type RecipientConfirmKeyMap struct {
	GlobalKeyMap
	// Confirm accepts the recipient change (y).
	Confirm key.Binding
	// Cancel rejects without changes (n).
	Cancel key.Binding
	// Abort cancels via Esc.
	Abort key.Binding
	// ScrollDown scrolls the diff content down (j, down).
	ScrollDown key.Binding
	// ScrollUp scrolls the diff content up (k, up).
	ScrollUp key.Binding
}

// ShortHelp returns the bindings rendered in the persistent menu for stateRecipientConfirm.
// Quit from GlobalKeyMap is included so HiddenFromMenu() can suppress it.
// Implements help.KeyMap.
func (k RecipientConfirmKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Confirm, k.Cancel, k.Abort, k.ScrollDown, k.ScrollUp, k.Quit}
}

// FullHelp returns grouped bindings for the expanded help overlay.
// Implements help.KeyMap. Two groups: confirm/cancel, scroll/global.
func (k RecipientConfirmKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Confirm, k.Cancel, k.Abort},
		{k.ScrollDown, k.ScrollUp, k.Help, k.Quit},
	}
}

// HiddenFromMenu suppresses Quit from the persistent menu per D-313 and
// UI-SPEC §confirm-flow-quit-suppression. The Quit binding remains functional
// in AppModel.Update(); only the menu rendering is suppressed.
func (k RecipientConfirmKeyMap) HiddenFromMenu() []key.Binding {
	return []key.Binding{k.Quit}
}

// DefaultRecipientConfirmKeyMap is the default instance with description strings matching
// the literal RecipientConfirmHints values from Phase 7 (description-string lock).
var DefaultRecipientConfirmKeyMap = RecipientConfirmKeyMap{
	GlobalKeyMap: DefaultGlobalKeyMap,
	Confirm: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "confirm add/remove recipient"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "cancel"),
	),
	Abort: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("Esc", "cancel"),
	),
	ScrollDown: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j", "scroll down"),
	),
	ScrollUp: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k", "scroll up"),
	),
}

// BulkReKeyConfirmKeyMap holds keybindings for stateBulkReKeyConfirm —
// the y/n/Esc per-file confirmation during a bulk re-key run.
// Embeds GlobalKeyMap; Quit is suppressed from the persistent menu via HiddenFromMenu().
// Implements help.KeyMap via ShortHelp() and FullHelp().
type BulkReKeyConfirmKeyMap struct {
	GlobalKeyMap
	// Confirm accepts re-keying this file (y).
	Confirm key.Binding
	// Cancel skips this file (n).
	Cancel key.Binding
	// Abort aborts the entire bulk re-key (esc).
	Abort key.Binding
	// ScrollDown scrolls the diff content down (j, down).
	ScrollDown key.Binding
	// ScrollUp scrolls the diff content up (k, up).
	ScrollUp key.Binding
}

// ShortHelp returns the bindings rendered in the persistent menu for stateBulkReKeyConfirm.
// Quit from GlobalKeyMap is included so HiddenFromMenu() can suppress it.
// Implements help.KeyMap.
func (k BulkReKeyConfirmKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Confirm, k.Cancel, k.Abort, k.ScrollDown, k.ScrollUp, k.Quit}
}

// FullHelp returns grouped bindings for the expanded help overlay.
// Implements help.KeyMap. Two groups: confirm/cancel, scroll/global.
func (k BulkReKeyConfirmKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Confirm, k.Cancel, k.Abort},
		{k.ScrollDown, k.ScrollUp, k.Help, k.Quit},
	}
}

// HiddenFromMenu suppresses Quit from the persistent menu per D-313 and
// UI-SPEC §confirm-flow-quit-suppression. The Quit binding remains functional
// in AppModel.Update(); only the menu rendering is suppressed.
func (k BulkReKeyConfirmKeyMap) HiddenFromMenu() []key.Binding {
	return []key.Binding{k.Quit}
}

// DefaultBulkReKeyConfirmKeyMap is the default instance with description strings matching
// the literal BulkReKeyConfirmHints values from Phase 7 (description-string lock).
var DefaultBulkReKeyConfirmKeyMap = BulkReKeyConfirmKeyMap{
	GlobalKeyMap: DefaultGlobalKeyMap,
	Confirm: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "confirm re-key this file"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "skip this file"),
	),
	Abort: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("Esc", "abort bulk re-key"),
	),
	ScrollDown: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j", "scroll down"),
	),
	ScrollUp: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k", "scroll up"),
	),
}

// RecipientListKeyMap holds keybindings for stateRecipientList — the inline
// recipient removal list rendered by AppModel.
// Implements help.KeyMap via ShortHelp() and FullHelp().
type RecipientListKeyMap struct {
	// Select selects a recipient by digit key (1-9).
	Select key.Binding
	// Cancel cancels without removing a recipient (esc).
	Cancel key.Binding
	// Quit exits the application (q, ctrl+c).
	Quit key.Binding
}

// ShortHelp returns the bindings rendered in the persistent menu for stateRecipientList.
// Implements help.KeyMap.
func (k RecipientListKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Select, k.Cancel, k.Quit}
}

// FullHelp returns grouped bindings for the expanded help overlay.
// Implements help.KeyMap. Single group.
func (k RecipientListKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Select, k.Cancel, k.Quit},
	}
}

// DefaultRecipientListKeyMap is the default instance with description strings matching
// the literal RecipientListHints values from Phase 7 (description-string lock).
// The "1-9" mnemonic is a display convention; the actual matchable keys are individual digits.
var DefaultRecipientListKeyMap = RecipientListKeyMap{
	Select: key.NewBinding(
		key.WithKeys("1", "2", "3", "4", "5", "6", "7", "8", "9"),
		key.WithHelp("1-9", "select recipient to remove"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("Esc", "cancel"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// FormatMenuKeyMap holds keybindings for stateFormatMenu — the inline format
// selection overlay driven by renderFormatMenu (no owning sub-model per D-09).
// No Quit in ShortHelp per OQ-3 — the format menu is a transient modal overlay;
// global quit via AppModel still works but is not advertised here.
// Implements help.KeyMap via ShortHelp() and FullHelp().
type FormatMenuKeyMap struct {
	// Next selects the next format (j, down).
	Next key.Binding
	// Prev selects the previous format (k, up).
	Prev key.Binding
	// Confirm accepts the selected format (enter).
	Confirm key.Binding
	// Cancel dismisses the format menu (esc).
	Cancel key.Binding
}

// ShortHelp returns the bindings rendered in the persistent menu for stateFormatMenu.
// No Quit per OQ-3 (format menu is a transient modal; quit not advertised).
// Implements help.KeyMap.
func (k FormatMenuKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Next, k.Prev, k.Confirm, k.Cancel}
}

// FullHelp returns grouped bindings for the expanded help overlay.
// Implements help.KeyMap. Single group of 4.
func (k FormatMenuKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Next, k.Prev, k.Confirm, k.Cancel},
	}
}

// DefaultFormatMenuKeyMap is the default instance with description strings matching
// the literal FormatMenuHints values from Phase 7 (description-string lock).
var DefaultFormatMenuKeyMap = FormatMenuKeyMap{
	Next: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("j", "next format"),
	),
	Prev: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("k", "prev format"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("Enter", "confirm format"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("Esc", "cancel"),
	),
}
