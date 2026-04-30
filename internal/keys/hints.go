// Package keys additions for Phase 7 (chrome skeleton) and hardened in
// Phase 9 (keybinding discoverability).
//
// MenuHint is one row in the persistent keybinding menu. Hinter is the
// interface every interactive sub-model implements so the AppModel
// dispatcher can query the active view's hints on every View() call.
// HintsFromBindings converts bubbles/v2/key.Binding.Help() output directly
// into MenuHint entries — the keymap is the single source of truth for
// menu content (D-301 — total derivation; closes SC5).
//
// After Phase 9: all hint sets derive from keymap.ShortHelp() — there are
// no inline package-var hint slices in this file anymore. The 5 stateless
// states that lacked an owning sub-model now have keymap-backed types in
// bindings.go (FileListSearchKeyMap, RecipientConfirmKeyMap,
// BulkReKeyConfirmKeyMap, RecipientListKeyMap, FormatMenuKeyMap). The
// visibility-suppression contract (Visible=false) is formalized via the
// unexported menuVisibilityOverrider interface in bindings.go (D-307).
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
