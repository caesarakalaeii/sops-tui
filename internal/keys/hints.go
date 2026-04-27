// Package keys additions for Phase 7 (chrome skeleton).
//
// MenuHint is one row in the persistent keybinding menu. Hinter is the
// interface every interactive sub-model implements so the AppModel
// dispatcher can query the active view's hints on every View() call.
// HintsFromBindings converts bubbles/v2/key.Binding.Help() output directly
// into MenuHint entries — the keymap is the single source of truth for
// menu content (Pitfall 3 mitigation).
//
// Inline hint-set package vars cover states with no owning sub-model:
// FileListSearchHints (D-11 search-active override), RecipientConfirmHints
// and BulkReKeyConfirmHints (shared diff body, disambiguated by
// recipientAction), RecipientListHints (no sub-model — renderer lives on
// AppModel per Pitfall 3), and FormatMenuHints (modal overlay per D-09).
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package keys

import "charm.land/bubbles/v2/key"

// MenuHint is one visible row in the persistent keybinding menu.
// Visible=false suppresses the hint from the persistent menu's 12 slots
// while keeping the key discoverable in the ? full-screen overlay (D-06).
type MenuHint struct {
	Mnemonic    string
	Description string
	Visible     bool
}

// Hinter is implemented by every interactive sub-model. The AppModel
// dispatcher queries the active sub-model on every View() call per D-08.
type Hinter interface {
	Hints() []MenuHint
}

// HintsFromBindings converts a slice of key.Binding into MenuHint entries.
// Each binding's Help() returns {Key, Desc} per bubbles/v2/key semantics;
// those map directly to MenuHint.Mnemonic/Description. All hints default
// to Visible=true — caller filters via MenuHint.Visible toggles if needed.
func HintsFromBindings(bindings []key.Binding) []MenuHint {
	hints := make([]MenuHint, 0, len(bindings))
	for _, b := range bindings {
		h := b.Help()
		hints = append(hints, MenuHint{
			Mnemonic:    h.Key,
			Description: h.Desc,
			Visible:     true,
		})
	}
	return hints
}

// FileListSearchHints is the D-11 override applied when FileListModel
// has an active inline search. Replaces the default FileList hints while
// search is active so the menu reflects keys that actually work.
var FileListSearchHints = []MenuHint{
	{Mnemonic: "Esc", Description: "exit search", Visible: true},
	{Mnemonic: "Enter", Description: "select result", Visible: true},
	{Mnemonic: "j/↓", Description: "next result", Visible: true},
	{Mnemonic: "k/↑", Description: "prev result", Visible: true},
	{Mnemonic: "?", Description: "toggle help", Visible: true},
	{Mnemonic: "q", Description: "quit", Visible: true},
}

// RecipientConfirmHints is the hint set for stateRecipientConfirm —
// the y/n confirmation over a shared diff body. Disambiguates the shared
// stateDiff body via AppModel.state (recipientAction unused here per D-10).
var RecipientConfirmHints = []MenuHint{
	{Mnemonic: "y", Description: "confirm add/remove recipient", Visible: true},
	{Mnemonic: "n", Description: "cancel", Visible: true},
	{Mnemonic: "Esc", Description: "cancel", Visible: true},
	{Mnemonic: "j", Description: "scroll down", Visible: true},
	{Mnemonic: "k", Description: "scroll up", Visible: true},
}

// BulkReKeyConfirmHints is the hint set for stateBulkReKeyConfirm —
// the y/n/Esc per-file confirmation during a bulk re-key run.
var BulkReKeyConfirmHints = []MenuHint{
	{Mnemonic: "y", Description: "confirm re-key this file", Visible: true},
	{Mnemonic: "n", Description: "skip this file", Visible: true},
	{Mnemonic: "Esc", Description: "abort bulk re-key", Visible: true},
	{Mnemonic: "j", Description: "scroll down", Visible: true},
	{Mnemonic: "k", Description: "scroll up", Visible: true},
}

// RecipientListHints is the hint set for stateRecipientList — which is
// rendered inline by AppModel (renderRecipientList at model.go:1799)
// rather than an owning sub-model per Pitfall 3.
var RecipientListHints = []MenuHint{
	{Mnemonic: "1-9", Description: "select recipient to remove", Visible: true},
	{Mnemonic: "Esc", Description: "cancel", Visible: true},
	{Mnemonic: "q", Description: "quit", Visible: true},
}

// FormatMenuHints is the hint set for stateFormatMenu — the inline
// overlay modal driven by renderFormatMenu at model.go:1857 (no
// owning sub-model per D-09).
var FormatMenuHints = []MenuHint{
	{Mnemonic: "j", Description: "next format", Visible: true},
	{Mnemonic: "k", Description: "prev format", Visible: true},
	{Mnemonic: "Enter", Description: "confirm format", Visible: true},
	{Mnemonic: "Esc", Description: "cancel", Visible: true},
}
