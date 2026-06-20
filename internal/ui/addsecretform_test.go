// Package ui_test provides tests for the AddSecretFormModel new-secret input modal.
package ui_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

// Compile-time interface compliance: AddSecretFormModel implements keys.Hinter.
var _ keys.Hinter = ui.AddSecretFormModel{}

// typeRunes feeds each rune of s to the model's Update as a KeyPressMsg so the
// focused textinput accumulates the value (mirrors real keystroke delivery).
func typeRunes(m ui.AddSecretFormModel, s string) ui.AddSecretFormModel {
	for _, r := range s {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

func TestAddSecretFormModel(t *testing.T) {
	t.Run("view contains Add Secret title and field labels", func(t *testing.T) {
		m := ui.NewAddSecretFormModel(80, 24, nil)
		view := m.View()
		assert.Contains(t, view, "Add Secret", "view should contain modal title")
		assert.Contains(t, view, "Key path:", "view should contain key-path label")
		assert.Contains(t, view, "Value:", "view should contain value label")
	})

	t.Run("Activate resets state and focuses key input", func(t *testing.T) {
		m := ui.NewAddSecretFormModel(80, 24, nil)
		cmd := m.Activate()
		assert.True(t, m.IsActive(), "model should be active after Activate")
		assert.Equal(t, "", m.KeyPath(), "key path should be empty after Activate")
		assert.Equal(t, "", m.Value(), "value should be empty after Activate")
		assert.NotNil(t, cmd, "Activate returns a focus cmd")
	})

	t.Run("Esc cancels", func(t *testing.T) {
		m := ui.NewAddSecretFormModel(80, 24, nil)
		m.Activate()
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
		assert.True(t, m.Cancelled(), "Esc should set Cancelled")
		assert.False(t, m.Confirmed(), "Esc should not set Confirmed")
		assert.False(t, m.IsActive(), "form is inactive after Esc")
	})

	t.Run("Enter with empty key path shows validation error", func(t *testing.T) {
		m := ui.NewAddSecretFormModel(80, 24, nil)
		m.Activate()
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		assert.False(t, m.Confirmed(), "empty key path must not confirm")
		assert.Contains(t, m.View(), "Key path required", "view shows validation error")
	})

	t.Run("Enter with array-index key path is rejected", func(t *testing.T) {
		m := ui.NewAddSecretFormModel(80, 24, nil)
		m.Activate()
		m = typeRunes(m, "items[0].name")
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		assert.False(t, m.Confirmed(), "array-index path must not confirm")
		assert.Contains(t, m.View(), "Array-index", "view shows array-index error")
	})

	t.Run("Enter with existing key is rejected", func(t *testing.T) {
		m := ui.NewAddSecretFormModel(80, 24, []string{"database.password"})
		m.Activate()
		m = typeRunes(m, "database.password")
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		assert.False(t, m.Confirmed(), "existing key must not confirm")
		assert.Contains(t, m.View(), "already exists", "view shows duplicate-key error")
	})

	t.Run("Tab switches focus so value goes to the value field", func(t *testing.T) {
		m := ui.NewAddSecretFormModel(80, 24, nil)
		m.Activate()
		m = typeRunes(m, "api.token")
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		m = typeRunes(m, "s3cret")
		assert.Equal(t, "api.token", m.KeyPath(), "key path captured before Tab")
		assert.Equal(t, "s3cret", m.Value(), "value captured after Tab")
	})

	t.Run("Enter with valid key/value confirms", func(t *testing.T) {
		m := ui.NewAddSecretFormModel(80, 24, nil)
		m.Activate()
		m = typeRunes(m, "api.token")
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
		m = typeRunes(m, "s3cret")
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		assert.True(t, m.Confirmed(), "valid key/value must confirm")
		assert.Equal(t, "api.token", m.KeyPath())
		assert.Equal(t, "s3cret", m.Value())
	})

	t.Run("KeyPath trims surrounding whitespace", func(t *testing.T) {
		m := ui.NewAddSecretFormModel(80, 24, nil)
		m.Activate()
		m = typeRunes(m, "  spaced.key  ")
		assert.Equal(t, "spaced.key", m.KeyPath(), "KeyPath must be trimmed")
	})
}

// TestAddSecretFormHints verifies AddSecretFormModel.Hints() returns the 3-hint
// persistent menu set: Tab/Enter/Esc.
func TestAddSecretFormHints(t *testing.T) {
	m := ui.NewAddSecretFormModel(80, 24, nil)
	hints := m.Hints()
	require.Equal(t, 3, len(hints), "AddSecretForm must expose 3 hints")
	assert.Equal(t, "Tab", hints[0].Mnemonic)
	assert.Equal(t, "Enter", hints[1].Mnemonic)
	assert.Equal(t, "Esc", hints[2].Mnemonic)
	for i, h := range hints {
		assert.True(t, h.Visible, "hint %d must default Visible=true", i)
	}
}
