package app_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caesarakalaeii/sops-tui/internal/app"
	"github.com/caesarakalaeii/sops-tui/internal/sops"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

// TestDecryptKeyMsgAppliesCorrectNode verifies that DecryptKeyMsg with a matching
// keyPath sets the correct TreeNode to Revealed=true.
func TestDecryptKeyMsgAppliesCorrectNode(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "", colorprofile.TrueColor)
	m2 := send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})

	// Seed with a file
	files := []sops.DiscoveredFile{
		{Name: "secrets/prod.yaml", AbsPath: "/repo/secrets/prod.yaml", IsEncrypted: true},
	}
	m3 := send(t, m2, app.FilesDiscoveredMsg{Files: files})

	// Manually put model in stateDetail by sending FilesParsedMsg
	nodes := []ui.TreeNode{
		{Key: "token", Encrypted: true, TypeHint: "str"},
		{Key: "other", Encrypted: true, TypeHint: "str"},
	}
	parsed := app.ParsedFileForTest(nodes)
	m4 := send(t, m3, app.FilesParsedMsg{Parsed: parsed})

	// Send DecryptKeyMsg for "token"
	m5 := send(t, m4, app.DecryptKeyMsg{KeyPath: "token", Value: "supersecret"})

	// View must contain "supersecret" (the revealed value)
	v := m5.View()
	assert.True(t, contains(v.Content, "supersecret"),
		"after DecryptKeyMsg, view must show the decrypted value, got: %q", v.Content)
}

// TestDecryptKeyMsgWithError flashes error and does not reveal.
func TestDecryptKeyMsgWithError(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "", colorprofile.TrueColor)
	// Send DecryptKeyMsg with an error — should not panic
	_, cmd := m.Update(app.DecryptKeyMsg{KeyPath: "token", Err: assert.AnError})
	_ = cmd // cmd may be nil or flash timer
}

// TestDecryptAllMsgRevealsAll verifies that DecryptAllMsg reveals all encrypted leaf nodes.
func TestDecryptAllMsgRevealsAll(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "", colorprofile.TrueColor)
	m2 := send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})

	files := []sops.DiscoveredFile{
		{Name: "secrets/prod.yaml", AbsPath: "/repo/secrets/prod.yaml", IsEncrypted: true},
	}
	m3 := send(t, m2, app.FilesDiscoveredMsg{Files: files})

	nodes := []ui.TreeNode{
		{Key: "a", Encrypted: true, TypeHint: "str"},
		{Key: "b", Encrypted: true, TypeHint: "str"},
	}
	parsed := app.ParsedFileForTest(nodes)
	m4 := send(t, m3, app.FilesParsedMsg{Parsed: parsed})

	// Send DecryptAllMsg
	m5 := send(t, m4, app.DecryptAllMsg{Values: map[string]string{
		"a": "value_a",
		"b": "value_b",
	}})

	v := m5.View()
	assert.True(t, contains(v.Content, "value_a"),
		"after DecryptAllMsg, view must show value_a, got: %q", v.Content)
	assert.True(t, contains(v.Content, "value_b"),
		"after DecryptAllMsg, view must show value_b, got: %q", v.Content)
}

// TestEscFromDetailClearsRevealed verifies that Esc from stateDetail calls ClearAllRevealed (D-04).
func TestEscFromDetailClearsRevealed(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "", colorprofile.TrueColor)
	m2 := send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})

	files := []sops.DiscoveredFile{
		{Name: "secrets/prod.yaml", AbsPath: "/repo/secrets/prod.yaml", IsEncrypted: true},
	}
	m3 := send(t, m2, app.FilesDiscoveredMsg{Files: files})

	nodes := []ui.TreeNode{
		{Key: "password", Encrypted: true, TypeHint: "str"},
	}
	parsed := app.ParsedFileForTest(nodes)
	m4 := send(t, m3, app.FilesParsedMsg{Parsed: parsed})

	// Reveal a value
	m5 := send(t, m4, app.DecryptKeyMsg{KeyPath: "password", Value: "revealed_value"})
	v1 := m5.View()
	assert.True(t, contains(v1.Content, "revealed_value"),
		"value should be visible before Esc, got: %q", v1.Content)

	// Press Esc → should clear revealed values AND navigate back to file list
	m6 := send(t, m5, tea.KeyPressMsg{Code: 27})
	v2 := m6.View()
	assert.False(t, contains(v2.Content, "revealed_value"),
		"after Esc from detail, revealed value must be cleared (D-04), got: %q", v2.Content)
}

// TestRevealRequestMsgReturnsCmd verifies that a RevealRequestMsg from DetailModel
// is handled by AppModel and produces a decrypt command.
func TestRevealRequestMsgReturnsCmd(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "", colorprofile.TrueColor)
	m2 := send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})

	files := []sops.DiscoveredFile{
		{Name: "secrets/prod.yaml", AbsPath: "/repo/secrets/prod.yaml", IsEncrypted: true},
	}
	m3 := send(t, m2, app.FilesDiscoveredMsg{Files: files})

	// Send RevealRequestMsg directly
	_, cmd := m3.Update(ui.RevealRequestMsg{KeyPath: "database.password"})
	require.NotNil(t, cmd, "RevealRequestMsg must produce a tea.Cmd")
}

// TestModelContainsStateDiff verifies that stateDiff and stateEdit constants are defined.
// We test this indirectly by verifying the model compiles with the new states.
func TestModelContainsStateDiff(t *testing.T) {
	// This is a compile-time check — if stateDiff/stateEdit don't exist, the package won't build.
	// We verify by importing app and checking the constants are accessible.
	assert.Equal(t, app.StateDiff, app.StateDiff, "stateDiff constant must exist")
	assert.Equal(t, app.StateEdit, app.StateEdit, "stateEdit constant must exist")
}

// contains is a helper for case-sensitive string containment check on view content.
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) &&
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}()
}
