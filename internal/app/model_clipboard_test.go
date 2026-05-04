package app_test

import (
	"os"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caesarakalaeii/sops-tui/internal/app"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

// setupDetailWithNodes puts the AppModel into stateDetail with the given nodes.
func setupDetailWithNodes(t *testing.T, nodes []ui.TreeNode) tea.Model {
	t.Helper()
	m := app.NewAppModel(defaultEnv(), "", colorprofile.TrueColor)
	m2 := send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	parsed := app.ParsedFileForTest(nodes)
	return send(t, m2, app.FilesParsedMsg{Parsed: parsed})
}

// asAppModel asserts the tea.Model is an AppModel and returns it.
func asAppModel(t *testing.T, m tea.Model) app.AppModel {
	t.Helper()
	am, ok := m.(app.AppModel)
	require.True(t, ok, "expected AppModel, got %T", m)
	return am
}

// TestClipboardCopyRevealedLeaf verifies that ctrl+y on a revealed leaf
// sets clipboardHot=true, flashes "Copied (clears in 30s)", and returns a tea.Tick cmd.
//
// Test 1 from plan behavior spec.
func TestClipboardCopyRevealedLeaf(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "token", Encrypted: true, TypeHint: "str", Revealed: true, DecryptedValue: "supersecret"},
	}
	m := setupDetailWithNodes(t, nodes)

	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl})

	// cmd must be non-nil (flash timer + clipboard clear tick)
	require.NotNil(t, cmd, "ctrl+y on revealed leaf must return a non-nil Cmd")

	// clipboardHot must be true
	am := asAppModel(t, updated)
	assert.True(t, am.IsClipboardHot(),
		"after ctrl+y on revealed leaf, clipboardHot must be true")

	// Flash message must be "Copied (clears in 30s)"
	v := updated.View()
	assert.Contains(t, v.Content, "Copied (clears in 30s)",
		"after ctrl+y on revealed leaf, flash must say 'Copied (clears in 30s)'")
}

// TestClipboardIndicatorVisibleAfterFlashClears verifies that [clip] appears in the
// status bar in normal (non-flash) mode after the flash clears.
func TestClipboardIndicatorVisibleAfterFlashClears(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "token", Encrypted: true, TypeHint: "str", Revealed: true, DecryptedValue: "supersecret"},
	}
	m := setupDetailWithNodes(t, nodes)

	// Copy — this activates flash AND sets clipboardHot
	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl})
	require.NotNil(t, cmd)

	// Drain all commands to find the FlashClearMsg gen
	am := asAppModel(t, updated)
	require.True(t, am.IsClipboardHot(), "clipboardHot must be true after copy")

	// Clear flash by sending FlashClearMsg (gen=1, matching the first flash)
	updated2, _ := updated.Update(ui.FlashClearMsg{Gen: 1})
	v := updated2.View()

	// Now in normal mode — [clip] should be visible
	assert.Contains(t, v.Content, "[clip]",
		"after flash clears, [clip] must appear in status bar while clipboardHot")
}

// TestClipboardCopyMaskedLeafFlashesMessage verifies that ctrl+y on a masked
// (non-revealed) leaf flashes "Reveal first with r" and does NOT set clipboardHot.
//
// Test 2 from plan behavior spec.
func TestClipboardCopyMaskedLeafFlashesMessage(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "token", Encrypted: true, TypeHint: "str", Revealed: false},
	}
	m := setupDetailWithNodes(t, nodes)

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl})

	// Status bar must flash "Reveal first with r"
	v := updated.View()
	assert.Contains(t, v.Content, "Reveal first with r",
		"ctrl+y on masked leaf must flash 'Reveal first with r'")

	// clipboardHot must NOT be set
	am := asAppModel(t, updated)
	assert.False(t, am.IsClipboardHot(),
		"ctrl+y on masked leaf must not set clipboardHot")
}

// TestClipboardCopyInFileListIsNoOp verifies that ctrl+y in stateFileList is a no-op.
// Copy is only available in stateDetail per D-03.
//
// Test 3 from plan behavior spec.
func TestClipboardCopyInFileListIsNoOp(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "", colorprofile.TrueColor)
	m2 := send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})

	// We are in stateFileList (no FilesParsedMsg sent)
	updated, _ := m2.Update(tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl})

	am := asAppModel(t, updated)
	assert.False(t, am.IsClipboardHot(),
		"ctrl+y in stateFileList must not set clipboardHot")

	v := updated.View()
	assert.NotContains(t, v.Content, "Reveal first with r",
		"ctrl+y in stateFileList must not flash any message")
}

// TestClipboardClearMsgMatchingGen verifies that ClipboardClearMsg with matching gen
// sets clipboardHot=false.
//
// Test 4 from plan behavior spec.
func TestClipboardClearMsgMatchingGen(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "token", Encrypted: true, TypeHint: "str", Revealed: true, DecryptedValue: "supersecret"},
	}
	m := setupDetailWithNodes(t, nodes)

	// Copy to clipboard — this sets clipboardGen=1 and clipboardHot=true
	m2, _ := m.Update(tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl})

	am2 := asAppModel(t, m2)
	require.True(t, am2.IsClipboardHot(), "clipboard must be hot after copy")

	// Send ClipboardClearMsg with gen=1 (matching)
	m3, _ := m2.Update(app.ClipboardClearMsg{Gen: 1})

	am3 := asAppModel(t, m3)
	assert.False(t, am3.IsClipboardHot(),
		"ClipboardClearMsg with matching gen must clear clipboardHot")

	// Also verify [clip] is gone from the view (no flash active at this point)
	v := m3.View()
	assert.NotContains(t, v.Content, "[clip]",
		"ClipboardClearMsg with matching gen must remove [clip] from status bar")
}

// TestClipboardClearMsgStaleGenIgnored verifies that a stale ClipboardClearMsg
// (gen != clipboardGen) does NOT clear clipboardHot.
//
// Test 5 from plan behavior spec.
func TestClipboardClearMsgStaleGenIgnored(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "token", Encrypted: true, TypeHint: "str", Revealed: true, DecryptedValue: "supersecret"},
	}
	m := setupDetailWithNodes(t, nodes)

	// Copy once — clipboardGen becomes 1
	m2, _ := m.Update(tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl})

	am2 := asAppModel(t, m2)
	require.True(t, am2.IsClipboardHot(), "clipboard must be hot after copy")

	// Send stale clear (gen=0, which doesn't match gen=1)
	m3, _ := m2.Update(app.ClipboardClearMsg{Gen: 0})

	am3 := asAppModel(t, m3)
	assert.True(t, am3.IsClipboardHot(),
		"stale ClipboardClearMsg (gen mismatch) must not clear clipboardHot")
}

// TestClipboardTimeoutDefault verifies clipboardTimeout() returns 30s when
// SOPS_TUI_CLIPBOARD_TIMEOUT is unset.
//
// Test 6 from plan behavior spec.
func TestClipboardTimeoutDefault(t *testing.T) {
	os.Unsetenv("SOPS_TUI_CLIPBOARD_TIMEOUT")
	d := app.ClipboardTimeout()
	assert.Equal(t, 30*time.Second, d,
		"default clipboard timeout must be 30s when env var is unset")
}

// TestClipboardTimeoutEnvVar verifies clipboardTimeout() returns 10s when
// SOPS_TUI_CLIPBOARD_TIMEOUT="10".
//
// Test 7 from plan behavior spec.
func TestClipboardTimeoutEnvVar(t *testing.T) {
	t.Setenv("SOPS_TUI_CLIPBOARD_TIMEOUT", "10")
	d := app.ClipboardTimeout()
	assert.Equal(t, 10*time.Second, d,
		"clipboard timeout must be 10s when SOPS_TUI_CLIPBOARD_TIMEOUT=10")
}

// TestClipboardTimeoutInvalidFallsBackTo30s verifies clipboardTimeout() returns 30s
// when SOPS_TUI_CLIPBOARD_TIMEOUT is set to an invalid value.
//
// Test 8 from plan behavior spec.
func TestClipboardTimeoutInvalidFallsBackTo30s(t *testing.T) {
	t.Setenv("SOPS_TUI_CLIPBOARD_TIMEOUT", "invalid")
	d := app.ClipboardTimeout()
	assert.Equal(t, 30*time.Second, d,
		"clipboard timeout must fall back to 30s when SOPS_TUI_CLIPBOARD_TIMEOUT is invalid")
}
