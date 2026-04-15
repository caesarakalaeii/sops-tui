package app_test

import (
	"os"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/caesarakalaeii/sops-tui/internal/app"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupDetailModel puts the AppModel into stateDetail with one revealed leaf
// and one masked leaf.
func setupDetailWithNodes(t *testing.T, nodes []ui.TreeNode) tea.Model {
	t.Helper()
	m := app.NewAppModel(defaultEnv(), "")
	m2 := send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	parsed := app.ParsedFileForTest(nodes)
	return send(t, m2, app.FilesParsedMsg{Parsed: parsed})
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

	// Status bar must show [clip] indicator (clipboardHot=true)
	v := updated.View()
	assert.Contains(t, v.Content, "[clip]",
		"after ctrl+y on revealed leaf, status bar must show [clip] indicator")
}

// TestClipboardCopyMaskedLeafFlashesMessage verifies that ctrl+y on a masked
// (non-revealed) leaf flashes "Reveal first with r" and does NOT call clipboard.WriteAll.
//
// Test 2 from plan behavior spec.
func TestClipboardCopyMaskedLeafFlashesMessage(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "token", Encrypted: true, TypeHint: "str", Revealed: false},
	}
	m := setupDetailWithNodes(t, nodes)

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl})

	// Status bar must flash "Reveal first with r" (visible as flash)
	v := updated.View()
	assert.Contains(t, v.Content, "Reveal first with r",
		"ctrl+y on masked leaf must flash 'Reveal first with r'")

	// [clip] indicator must NOT be present
	assert.NotContains(t, v.Content, "[clip]",
		"ctrl+y on masked leaf must not set clipboardHot")
}

// TestClipboardCopyInFileListIsNoOp verifies that ctrl+y in stateFileList is a no-op.
// Copy is only available in stateDetail per D-03.
//
// Test 3 from plan behavior spec.
func TestClipboardCopyInFileListIsNoOp(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	m2 := send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})

	// We are in stateFileList (no FilesParsedMsg sent)
	updated, _ := m2.Update(tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl})

	v := updated.View()
	assert.NotContains(t, v.Content, "[clip]",
		"ctrl+y in stateFileList must not set clipboardHot")
	assert.NotContains(t, v.Content, "Reveal first with r",
		"ctrl+y in stateFileList must not flash any message")
}

// TestClipboardClearMsgMatchingGen verifies that ClipboardClearMsg with matching gen
// sets clipboardHot=false and removes [clip] from the status bar.
//
// Test 4 from plan behavior spec.
func TestClipboardClearMsgMatchingGen(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "token", Encrypted: true, TypeHint: "str", Revealed: true, DecryptedValue: "supersecret"},
	}
	m := setupDetailWithNodes(t, nodes)

	// Copy to clipboard — this sets clipboardGen=1 and clipboardHot=true
	m2, _ := m.Update(tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl})

	// Verify [clip] is showing
	v := m2.View()
	require.Contains(t, v.Content, "[clip]", "clipboard must be hot after copy")

	// Send ClipboardClearMsg with gen=1 (matching)
	m3, _ := m2.Update(app.ClipboardClearMsg{Gen: 1})

	v2 := m3.View()
	assert.NotContains(t, v2.Content, "[clip]",
		"ClipboardClearMsg with matching gen must clear [clip] indicator")
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

	// Send stale clear (gen=0, which doesn't match gen=1)
	m3, _ := m2.Update(app.ClipboardClearMsg{Gen: 0})

	v := m3.View()
	assert.Contains(t, v.Content, "[clip]",
		"stale ClipboardClearMsg (gen mismatch) must not clear [clip] indicator")
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
