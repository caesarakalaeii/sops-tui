package ui_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDiffModelViewSingleEntry verifies that View() renders old/new value lines,
// title, and the confirmation footer for a single-entry diff.
func TestDiffModelViewSingleEntry(t *testing.T) {
	entries := []ui.DiffEntry{
		{KeyPath: "database.password", OldValue: "old_secret", NewValue: "new_secret"},
	}
	m := ui.NewDiffModel("Changes: database.password", entries, 80, 24)
	view := m.View()
	stripped := stripAnsi(view)

	assert.True(t, strings.Contains(stripped, "Changes: database.password"),
		"view must contain title, got: %q", stripped)
	assert.True(t, strings.Contains(stripped, "old_secret"),
		"view must contain old value, got: %q", stripped)
	assert.True(t, strings.Contains(stripped, "new_secret"),
		"view must contain new value, got: %q", stripped)
	assert.True(t, strings.Contains(stripped, "[y]"),
		"view must contain [y] footer, got: %q", stripped)
	assert.True(t, strings.Contains(stripped, "[n/Esc]"),
		"view must contain [n/Esc] footer, got: %q", stripped)
	assert.True(t, strings.Contains(stripped, "confirm re-encrypt"),
		"view must contain 'confirm re-encrypt', got: %q", stripped)
}

// TestDiffModelViewMultiEntry verifies that multiple entries all render with key headers.
func TestDiffModelViewMultiEntry(t *testing.T) {
	entries := []ui.DiffEntry{
		{KeyPath: "database.password", OldValue: "old_pass", NewValue: "new_pass"},
		{KeyPath: "api.secret", OldValue: "old_secret", NewValue: "new_secret"},
	}
	m := ui.NewDiffModel("Changes: 2 keys modified", entries, 80, 24)
	view := m.View()
	stripped := stripAnsi(view)

	assert.True(t, strings.Contains(stripped, "Changes: 2 keys modified"),
		"view must contain multi-entry title, got: %q", stripped)
	assert.True(t, strings.Contains(stripped, "database.password"),
		"view must contain first entry key path, got: %q", stripped)
	assert.True(t, strings.Contains(stripped, "api.secret"),
		"view must contain second entry key path, got: %q", stripped)
	assert.True(t, strings.Contains(stripped, "old_pass"),
		"view must contain first old value, got: %q", stripped)
	assert.True(t, strings.Contains(stripped, "new_secret"),
		"view must contain second new value, got: %q", stripped)
}

// TestDiffModelScrollClamp verifies scroll is clamped at top (ScrollUp at 0) and bottom.
func TestDiffModelScrollClamp(t *testing.T) {
	entries := []ui.DiffEntry{
		{KeyPath: "key", OldValue: "old", NewValue: "new"},
	}
	m := ui.NewDiffModel("Changes: key", entries, 80, 24)

	// ScrollUp at 0 should stay at 0
	m.ScrollUp()
	// No panic and view still renders
	view := m.View()
	assert.NotEmpty(t, view, "view must not be empty after ScrollUp at 0")

	// ScrollDown many times — should not panic
	for i := 0; i < 100; i++ {
		m.ScrollDown()
	}
	view2 := m.View()
	assert.NotEmpty(t, view2, "view must not be empty after many ScrollDown calls")
}

// TestDiffModelConfirmY verifies that the y key sets Confirmed() to true.
func TestDiffModelConfirmY(t *testing.T) {
	entries := []ui.DiffEntry{
		{KeyPath: "key", OldValue: "old", NewValue: "new"},
	}
	m := ui.NewDiffModel("Changes: key", entries, 80, 24)
	require.False(t, m.Confirmed(), "must not be confirmed before y key")

	msg := tea.KeyPressMsg{Code: 'y'}
	m, _ = m.Update(msg)
	assert.True(t, m.Confirmed(), "must be confirmed after y key")
	assert.False(t, m.Cancelled(), "must not be cancelled after y key")
}

// TestDiffModelCancelN verifies that the n key sets Cancelled() to true.
func TestDiffModelCancelN(t *testing.T) {
	entries := []ui.DiffEntry{
		{KeyPath: "key", OldValue: "old", NewValue: "new"},
	}
	m := ui.NewDiffModel("Changes: key", entries, 80, 24)
	require.False(t, m.Cancelled(), "must not be cancelled before n key")

	msg := tea.KeyPressMsg{Code: 'n'}
	m, _ = m.Update(msg)
	assert.True(t, m.Cancelled(), "must be cancelled after n key")
	assert.False(t, m.Confirmed(), "must not be confirmed after n key")
}

// TestDiffModelCancelEsc verifies that the Esc key sets Cancelled() to true.
func TestDiffModelCancelEsc(t *testing.T) {
	entries := []ui.DiffEntry{
		{KeyPath: "key", OldValue: "old", NewValue: "new"},
	}
	m := ui.NewDiffModel("Changes: key", entries, 80, 24)
	require.False(t, m.Cancelled(), "must not be cancelled before esc")

	msg := tea.KeyPressMsg{Code: tea.KeyEsc}
	m, _ = m.Update(msg)
	assert.True(t, m.Cancelled(), "must be cancelled after esc key")
}

// TestDiffModelEntriesReturnsAll verifies Entries() returns the original entries.
func TestDiffModelEntriesReturnsAll(t *testing.T) {
	entries := []ui.DiffEntry{
		{KeyPath: "key1", OldValue: "old1", NewValue: "new1"},
		{KeyPath: "key2", OldValue: "old2", NewValue: "new2"},
	}
	m := ui.NewDiffModel("Changes: 2 keys modified", entries, 80, 24)
	got := m.Entries()
	require.Len(t, got, 2, "Entries() must return all entries")
	assert.Equal(t, "key1", got[0].KeyPath)
	assert.Equal(t, "key2", got[1].KeyPath)
}
