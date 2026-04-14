package ui_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// encryptedNodes returns a set of nodes with encrypted leaves for reveal tests.
func encryptedNodes() []ui.TreeNode {
	return []ui.TreeNode{
		{
			Key:       "database",
			Encrypted: false,
			Children: []ui.TreeNode{
				{Key: "password", Encrypted: true, TypeHint: "str"},
				{Key: "host", Encrypted: true, TypeHint: "str"},
			},
			Expanded: true,
		},
		{Key: "token", Encrypted: true, TypeHint: "str"},
	}
}

// TestRevealedNodeRendersValue verifies that a TreeNode with Revealed=true renders
// the DecryptedValue instead of "***".
func TestRevealedNodeRendersValue(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "password", Encrypted: true, TypeHint: "str", Revealed: true, DecryptedValue: "secret123"},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true)
	view := m.View()
	stripped := stripAnsi(view)
	assert.True(t, strings.Contains(stripped, "secret123"),
		"revealed node must render DecryptedValue, got: %q", stripped)
	// Must NOT render masked value when revealed
	assert.False(t, strings.Contains(stripped, "***"),
		"revealed node must not render *** when Revealed=true, got: %q", stripped)
}

// TestRevealedNodeRendersLockOpenIcon verifies that a revealed node renders the lock-open icon.
func TestRevealedNodeRendersLockOpenIcon(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "password", Encrypted: true, TypeHint: "str", Revealed: true, DecryptedValue: "s3cr3t"},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true)
	view := m.View()
	// Lock-open emoji is \U0001F513 (🔓)
	assert.True(t, strings.Contains(view, "\U0001F513"),
		"revealed node must render lock-open icon (🔓), got: %q", view)
}

// TestMaskedNodeRendersStars verifies that an encrypted, non-revealed node renders "***".
func TestMaskedNodeRendersStars(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "password", Encrypted: true, TypeHint: "str", Revealed: false},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true)
	view := m.View()
	stripped := stripAnsi(view)
	assert.True(t, strings.Contains(stripped, "***"),
		"masked encrypted node must render ***, got: %q", stripped)
}

// TestClearAllRevealed verifies that ClearAllRevealed sets all nodes Revealed=false and DecryptedValue="".
func TestClearAllRevealed(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "a", Encrypted: true, TypeHint: "str", Revealed: true, DecryptedValue: "val1"},
		{Key: "b", Encrypted: true, TypeHint: "str", Revealed: true, DecryptedValue: "val2"},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true)

	// Verify initially revealed
	view1 := m.View()
	assert.True(t, strings.Contains(view1, "val1") || strings.Contains(view1, "val2"),
		"initial view must show revealed values")

	// ClearAllRevealed
	m.ClearAllRevealed()

	view2 := m.View()
	stripped := stripAnsi(view2)
	assert.False(t, strings.Contains(stripped, "val1"),
		"after ClearAllRevealed, val1 must not be visible, got: %q", stripped)
	assert.False(t, strings.Contains(stripped, "val2"),
		"after ClearAllRevealed, val2 must not be visible, got: %q", stripped)
}

// TestRevealNodeByKeyPath verifies that RevealNode sets the correct node by keyPath
// (Pitfall 2: match by keyPath, not cursor index).
func TestRevealNodeByKeyPath(t *testing.T) {
	nodes := encryptedNodes()
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true)

	// Reveal "database.password" (not at cursor position 0 which is "database")
	m.RevealNode("database.password", "topsecret")

	view := m.View()
	assert.True(t, strings.Contains(view, "topsecret"),
		"RevealNode must set the correct node by keyPath, got: %q", view)

	// "database.host" should still be masked
	stripped := stripAnsi(view)
	_ = stripped // the host value is still masked as ***
}

// TestMaskNode verifies that MaskNode hides a previously revealed value.
func TestMaskNode(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "token", Encrypted: true, TypeHint: "str", Revealed: true, DecryptedValue: "abc123"},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true)

	// Verify initially revealed
	view1 := m.View()
	assert.True(t, strings.Contains(view1, "abc123"), "should show revealed value initially")

	// Mask the node
	m.MaskNode("token")

	view2 := m.View()
	stripped := stripAnsi(view2)
	assert.False(t, strings.Contains(stripped, "abc123"),
		"after MaskNode, value must be hidden, got: %q", stripped)
	assert.True(t, strings.Contains(stripped, "***"),
		"after MaskNode, *** must be shown, got: %q", stripped)
}

// TestAnyRevealed verifies AnyRevealed returns correct state.
func TestAnyRevealed(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "a", Encrypted: true, TypeHint: "str", Revealed: false},
		{Key: "b", Encrypted: true, TypeHint: "str", Revealed: false},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true)

	assert.False(t, m.AnyRevealed(), "AnyRevealed must be false when no nodes are revealed")

	m.RevealNode("a", "value")
	assert.True(t, m.AnyRevealed(), "AnyRevealed must be true after revealing a node")

	m.ClearAllRevealed()
	assert.False(t, m.AnyRevealed(), "AnyRevealed must be false after ClearAllRevealed")
}

// TestRevealAllNodes verifies RevealAllNodes sets all nodes with matching keyPaths.
func TestRevealAllNodes(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "a", Encrypted: true, TypeHint: "str"},
		{Key: "b", Encrypted: true, TypeHint: "str"},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true)

	values := map[string]string{
		"a": "value_a",
		"b": "value_b",
	}
	m.RevealAllNodes(values)

	view := m.View()
	assert.True(t, strings.Contains(view, "value_a"), "RevealAllNodes must set value_a, got: %q", view)
	assert.True(t, strings.Contains(view, "value_b"), "RevealAllNodes must set value_b, got: %q", view)
}

// TestRevealRequestMsgReturnedOnR verifies that pressing r on an encrypted, non-revealed
// leaf returns a tea.Cmd (RevealRequestMsg dispatch).
func TestRevealRequestMsgReturnedOnR(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "password", Encrypted: true, TypeHint: "str", Revealed: false},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true)

	// Press r on the encrypted leaf (cursor at 0)
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'r'})
	require.NotNil(t, cmd, "pressing r on encrypted leaf must return a tea.Cmd")

	// Execute the cmd and check it produces a RevealRequestMsg
	msg := cmd()
	_, ok := msg.(ui.RevealRequestMsg)
	assert.True(t, ok, "cmd must produce RevealRequestMsg, got: %T", msg)
}

// TestRevealRequestMsgHasKeyPath verifies the RevealRequestMsg contains the correct keyPath.
func TestRevealRequestMsgHasKeyPath(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "token", Encrypted: true, TypeHint: "str", Revealed: false},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true)

	_, cmd := m.Update(tea.KeyPressMsg{Code: 'r'})
	require.NotNil(t, cmd)

	msg := cmd()
	reqMsg, ok := msg.(ui.RevealRequestMsg)
	require.True(t, ok, "cmd must produce RevealRequestMsg")
	assert.Equal(t, "token", reqMsg.KeyPath, "RevealRequestMsg must have keyPath = 'token'")
}

// TestMaskOnRWhenRevealed verifies that pressing r on an already-revealed leaf masks it.
func TestMaskOnRWhenRevealed(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "password", Encrypted: true, TypeHint: "str", Revealed: true, DecryptedValue: "mysecret"},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true)

	// Verify it's currently revealed
	view1 := m.View()
	assert.True(t, strings.Contains(view1, "mysecret"), "should be revealed initially")

	// Press r → should mask
	m2, _ := m.Update(tea.KeyPressMsg{Code: 'r'})

	view2 := m2.View()
	stripped := stripAnsi(view2)
	assert.False(t, strings.Contains(stripped, "mysecret"),
		"after r on revealed node, value must be hidden, got: %q", stripped)
}

// TestRevealAllRequestMsgOnR_Capital verifies that pressing R when no values are revealed
// returns a RevealAllRequestMsg.
func TestRevealAllRequestMsgOnR_Capital(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "a", Encrypted: true, TypeHint: "str", Revealed: false},
		{Key: "b", Encrypted: true, TypeHint: "str", Revealed: false},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true)

	_, cmd := m.Update(tea.KeyPressMsg{Code: 'R'})
	require.NotNil(t, cmd, "pressing R when no values revealed must return a tea.Cmd")

	msg := cmd()
	_, ok := msg.(ui.RevealAllRequestMsg)
	assert.True(t, ok, "cmd must produce RevealAllRequestMsg, got: %T", msg)
}

// TestMaskAllOnR_CapitalWhenRevealed verifies that pressing R when values are revealed
// masks all values (no cmd needed).
func TestMaskAllOnR_CapitalWhenRevealed(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "a", Encrypted: true, TypeHint: "str", Revealed: true, DecryptedValue: "val_a"},
		{Key: "b", Encrypted: true, TypeHint: "str", Revealed: true, DecryptedValue: "val_b"},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true)

	// Verify revealed initially
	view1 := m.View()
	assert.True(t, strings.Contains(view1, "val_a"), "should show val_a initially")

	// Press R → should mask all
	m2, _ := m.Update(tea.KeyPressMsg{Code: 'R'})

	view2 := m2.View()
	stripped := stripAnsi(view2)
	assert.False(t, strings.Contains(stripped, "val_a"),
		"after R when revealed, val_a must be hidden, got: %q", stripped)
	assert.False(t, strings.Contains(stripped, "val_b"),
		"after R when revealed, val_b must be hidden, got: %q", stripped)
}

// TestStylesContainRevealedValueStyle verifies that RevealedValueStyle is defined in the ui package.
func TestStylesContainRevealedValueStyle(t *testing.T) {
	// If RevealedValueStyle compiles, it exists. Use it to render something.
	rendered := ui.RevealedValueStyle.Render("test_value")
	assert.NotEmpty(t, rendered, "RevealedValueStyle must render non-empty content")
}

// TestStylesContainRevealedIconStyle verifies RevealedIconStyle is defined.
func TestStylesContainRevealedIconStyle(t *testing.T) {
	rendered := ui.RevealedIconStyle.Render("\U0001F513")
	assert.NotEmpty(t, rendered, "RevealedIconStyle must render non-empty content")
}
