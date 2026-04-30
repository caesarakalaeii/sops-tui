// Tests for the Phase 7 hints contract: MenuHint struct, Hinter interface,
// HintsFromBindings converter, and the five inline hint-set package vars
// (D-08, D-09, D-11, UI-SPEC §"Inline hint sets").
package keys_test

import (
	"testing"

	"charm.land/bubbles/v2/key"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
)

// stubHinter is a minimal type used only to assert at compile-time that the
// Hinter interface has the expected single-method shape (Hints() []MenuHint).
type stubHinter struct{}

func (stubHinter) Hints() []keys.MenuHint { return nil }

// Compile-time assertion: stubHinter satisfies keys.Hinter.
var _ keys.Hinter = stubHinter{}

// TestHintsFromBindings_RoundTrips verifies that HintsFromBindings copies
// each binding's Help() {Key, Desc} into a MenuHint with Visible=true.
func TestHintsFromBindings_RoundTrips(t *testing.T) {
	bindings := []key.Binding{
		key.NewBinding(key.WithKeys("j"), key.WithHelp("j", "move down")),
		key.NewBinding(key.WithKeys("k"), key.WithHelp("k", "move up")),
	}

	got := keys.HintsFromBindings(bindings)

	require.Len(t, got, len(bindings), "result length must equal input length")
	assert.Equal(t, "j", got[0].Mnemonic)
	assert.Equal(t, "move down", got[0].Description)
	assert.True(t, got[0].Visible, "default visibility must be true")
	assert.Equal(t, "k", got[1].Mnemonic)
	assert.Equal(t, "move up", got[1].Description)
	assert.True(t, got[1].Visible)
}

// TestHintsFromBindings_EmptyInput verifies safe handling of nil and empty
// inputs — must return a zero-length slice and never panic.
func TestHintsFromBindings_EmptyInput(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		got := keys.HintsFromBindings(nil)
		assert.Len(t, got, 0, "nil input must return zero-length slice")
	})
	t.Run("empty input", func(t *testing.T) {
		got := keys.HintsFromBindings([]key.Binding{})
		assert.Len(t, got, 0, "empty input must return zero-length slice")
	})
}

// TestHintsFromBindings_RealFileListKeyMap feeds the real
// DefaultFileListKeyMap.ShortHelp() and asserts every binding round-trips.
// After D-304 (Phase 9): ShortHelp returns 12 bindings (GoTop + GoBottom appended).
func TestHintsFromBindings_RealFileListKeyMap(t *testing.T) {
	bindings := keys.DefaultFileListKeyMap.ShortHelp()
	require.Len(t, bindings, 12, "DefaultFileListKeyMap.ShortHelp must return 12 bindings")

	hints := keys.HintsFromBindings(bindings)
	require.Len(t, hints, 12, "expected 12 hints from 12 bindings")
	for i, b := range bindings {
		help := b.Help()
		assert.Equal(t, help.Key, hints[i].Mnemonic, "Mnemonic must equal binding.Help().Key at index %d", i)
		assert.Equal(t, help.Desc, hints[i].Description, "Description must equal binding.Help().Desc at index %d", i)
		assert.True(t, hints[i].Visible, "Visible must default to true at index %d", i)
	}
}

// TestFileListSearchHints_ExactCopy locks the verbatim description strings
// for the search-active override (D-11). After Phase 9 D-312, the authoritative
// source is DefaultFileListSearchKeyMap.ShortHelp() in bindings.go.
func TestFileListSearchHints_ExactCopy(t *testing.T) {
	expected := []keys.MenuHint{
		{Mnemonic: "Esc", Description: "exit search", Visible: true},
		{Mnemonic: "Enter", Description: "select result", Visible: true},
		{Mnemonic: "j/↓", Description: "next result", Visible: true},
		{Mnemonic: "k/↑", Description: "prev result", Visible: true},
		{Mnemonic: "?", Description: "toggle help", Visible: true},
		{Mnemonic: "q", Description: "quit", Visible: true},
	}
	assert.Equal(t, expected, keys.HintsFromBindings(keys.DefaultFileListSearchKeyMap.ShortHelp()))
}

// TestRecipientConfirmHints_ExactCopy locks the verbatim description strings
// for the y/n confirm overlay over the shared diff body. After Phase 9 D-312,
// the authoritative source is DefaultRecipientConfirmKeyMap.ShortHelp() in bindings.go.
// Note: Quit is the 6th element (from GlobalKeyMap embedding) but is suppressed
// from the persistent menu via RecipientConfirmKeyMap.HiddenFromMenu() (D-313).
func TestRecipientConfirmHints_ExactCopy(t *testing.T) {
	expected := []keys.MenuHint{
		{Mnemonic: "y", Description: "confirm add/remove recipient", Visible: true},
		{Mnemonic: "n", Description: "cancel", Visible: true},
		{Mnemonic: "Esc", Description: "cancel", Visible: true},
		{Mnemonic: "j", Description: "scroll down", Visible: true},
		{Mnemonic: "k", Description: "scroll up", Visible: true},
		{Mnemonic: "q", Description: "quit", Visible: true},
	}
	assert.Equal(t, expected, keys.HintsFromBindings(keys.DefaultRecipientConfirmKeyMap.ShortHelp()))
}

// TestBulkReKeyConfirmHints_ExactCopy locks the verbatim description strings
// for the per-file bulk re-key confirmation overlay. After Phase 9 D-312,
// the authoritative source is DefaultBulkReKeyConfirmKeyMap.ShortHelp() in bindings.go.
// Note: Quit is the 6th element (from GlobalKeyMap embedding) but is suppressed
// from the persistent menu via BulkReKeyConfirmKeyMap.HiddenFromMenu() (D-313).
func TestBulkReKeyConfirmHints_ExactCopy(t *testing.T) {
	expected := []keys.MenuHint{
		{Mnemonic: "y", Description: "confirm re-key this file", Visible: true},
		{Mnemonic: "n", Description: "skip this file", Visible: true},
		{Mnemonic: "Esc", Description: "abort bulk re-key", Visible: true},
		{Mnemonic: "j", Description: "scroll down", Visible: true},
		{Mnemonic: "k", Description: "scroll up", Visible: true},
		{Mnemonic: "q", Description: "quit", Visible: true},
	}
	assert.Equal(t, expected, keys.HintsFromBindings(keys.DefaultBulkReKeyConfirmKeyMap.ShortHelp()))
}

// TestRecipientListHints_ExactCopy locks the verbatim description strings
// for the recipient list view. After Phase 9 D-312, the authoritative source
// is DefaultRecipientListKeyMap.ShortHelp() in bindings.go.
func TestRecipientListHints_ExactCopy(t *testing.T) {
	expected := []keys.MenuHint{
		{Mnemonic: "1-9", Description: "select recipient to remove", Visible: true},
		{Mnemonic: "Esc", Description: "cancel", Visible: true},
		{Mnemonic: "q", Description: "quit", Visible: true},
	}
	assert.Equal(t, expected, keys.HintsFromBindings(keys.DefaultRecipientListKeyMap.ShortHelp()))
}

// TestFormatMenuHints_ExactCopy locks the verbatim description strings
// for the inline format menu modal. After Phase 9 D-312, the authoritative
// source is DefaultFormatMenuKeyMap.ShortHelp() in bindings.go.
func TestFormatMenuHints_ExactCopy(t *testing.T) {
	expected := []keys.MenuHint{
		{Mnemonic: "j", Description: "next format", Visible: true},
		{Mnemonic: "k", Description: "prev format", Visible: true},
		{Mnemonic: "Enter", Description: "confirm format", Visible: true},
		{Mnemonic: "Esc", Description: "cancel", Visible: true},
	}
	assert.Equal(t, expected, keys.HintsFromBindings(keys.DefaultFormatMenuKeyMap.ShortHelp()))
}

// TestHinterInterface_Compiles asserts a type with Hints() []MenuHint can be
// assigned to keys.Hinter — the contract for sub-model dispatch (D-08).
func TestHinterInterface_Compiles(t *testing.T) {
	var h keys.Hinter = stubHinter{}
	assert.Nil(t, h.Hints(), "stub Hints() returns nil — interface satisfied at compile time")
}
