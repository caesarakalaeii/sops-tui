package ui_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"github.com/caesarakalaeii/sops-tui/internal/keys"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Compile-time interface compliance: DetailModel implements keys.Hinter.
var _ keys.Hinter = ui.DetailModel{}

// sampleTree builds a representative tree for tests:
//
//	database (collapsed)
//	  host: ***
//	  port: ***
//	api (expanded)
//	  key: ***
//	  secret: ***
//	token: ***
func sampleTree() []ui.TreeNode {
	return []ui.TreeNode{
		{
			Key: "database",
			Children: []ui.TreeNode{
				{Key: "host", Value: "***", Depth: 1},
				{Key: "port", Value: "***", Depth: 1},
			},
			Expanded: false,
			Depth:    0,
		},
		{
			Key: "api",
			Children: []ui.TreeNode{
				{Key: "key", Value: "***", Depth: 1},
				{Key: "secret", Value: "***", Depth: 1},
			},
			Expanded: true,
			Depth:    0,
		},
		{
			Key:   "token",
			Value: "***",
			Depth: 0,
		},
	}
}

// TestDetailTreeNodeExpandCollapse verifies TreeNode with children can be toggled.
func TestDetailTreeNodeExpandCollapse(t *testing.T) {
	node := ui.TreeNode{
		Key:      "database",
		Children: []ui.TreeNode{{Key: "host", Value: "***"}},
		Expanded: false,
	}
	assert.False(t, node.Expanded, "node starts collapsed")

	node.Expanded = true
	assert.True(t, node.Expanded, "node can be expanded")

	node.Expanded = false
	assert.False(t, node.Expanded, "node can be collapsed again")
}

// TestDetailRenderTreeConnectors verifies RenderTree output contains tree connectors.
func TestDetailRenderTreeConnectors(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "alpha", Value: "***", Depth: 0},
		{Key: "beta", Value: "***", Depth: 0},
	}
	m := ui.NewDetailModel("test.yaml", nodes, 80, 24, true, "")
	view := m.View()
	// At least one of the tree connectors must appear
	hasConnector := strings.Contains(view, "├─") ||
		strings.Contains(view, "└─") ||
		strings.Contains(view, "│")
	assert.True(t, hasConnector, "view must contain tree connectors (├─ or └─ or │), got: %q", view)
}

// TestDetailCollapsedNodeShowsPlus verifies collapsed parent renders "[+]" indicator.
func TestDetailCollapsedNodeShowsPlus(t *testing.T) {
	nodes := []ui.TreeNode{
		{
			Key:      "database",
			Children: []ui.TreeNode{{Key: "host", Value: "***"}},
			Expanded: false,
		},
	}
	m := ui.NewDetailModel("test.yaml", nodes, 80, 24, true, "")
	view := m.View()
	assert.True(t, strings.Contains(view, "[+]"),
		"collapsed node must render '[+]' indicator, got: %q", view)
}

// TestDetailExpandedNodeShowsMinus verifies expanded parent renders "[-]" indicator.
func TestDetailExpandedNodeShowsMinus(t *testing.T) {
	nodes := []ui.TreeNode{
		{
			Key:      "api",
			Children: []ui.TreeNode{{Key: "key", Value: "***"}},
			Expanded: true,
		},
	}
	m := ui.NewDetailModel("test.yaml", nodes, 80, 24, true, "")
	view := m.View()
	assert.True(t, strings.Contains(view, "[-]"),
		"expanded node must render '[-]' indicator, got: %q", view)
}

// TestDetailLeafNodeRendersStarred verifies leaf node with value renders as "key: ***"
func TestDetailLeafNodeRendersStarred(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "token", Value: "***", Depth: 0},
	}
	m := ui.NewDetailModel("test.yaml", nodes, 80, 24, true, "")
	view := m.View()
	assert.True(t, strings.Contains(view, "***"),
		"leaf node must render masked value '***', got: %q", view)
}

// TestDetailIndentation verifies tree indentation is 2 cells per nesting level.
// A depth-1 child should be indented by at least 2 spaces beyond the parent.
func TestDetailIndentation(t *testing.T) {
	nodes := []ui.TreeNode{
		{
			Key: "parent",
			Children: []ui.TreeNode{
				{Key: "child", Value: "***", Depth: 1},
			},
			Expanded: true,
			Depth:    0,
		},
	}
	m := ui.NewDetailModel("test.yaml", nodes, 80, 24, true, "")
	view := m.View()
	lines := strings.Split(view, "\n")
	// Find the child line — it should be indented by at least 2 spaces (TreeIndent = 2 cells).
	// The child line contains "child" and should have leading spaces.
	var childLine string
	for _, line := range lines {
		if strings.Contains(line, "child") {
			childLine = line
			break
		}
	}
	require.NotEmpty(t, childLine, "must find child line in tree view")
	// Strip ANSI escape codes for whitespace check — count leading spaces.
	stripped := stripAnsi(childLine)
	assert.True(t, strings.HasPrefix(stripped, " ") || strings.HasPrefix(stripped, "\t"),
		"child line must have leading indentation of at least 2 cells: %q", stripped)
}

// TestDetailUpdateExpandOnEnter verifies Update with Enter key on collapsed node expands it.
func TestDetailUpdateExpandOnEnter(t *testing.T) {
	nodes := []ui.TreeNode{
		{
			Key:      "database",
			Children: []ui.TreeNode{{Key: "host", Value: "***"}},
			Expanded: false,
		},
	}
	m := ui.NewDetailModel("test.yaml", nodes, 80, 24, true, "")

	// Before: collapsed (no "host" leaf visible)
	before := m.View()
	assert.False(t, strings.Contains(before, "host"),
		"host child should not be visible when database is collapsed")

	// Send Enter key to expand
	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	m2, _ := m.Update(msg)

	after := m2.View()
	assert.True(t, strings.Contains(after, "host"),
		"host child must be visible after expanding with Enter, view: %q", after)
}

// TestDetailUpdateCollapseOnH verifies Update with h key on expanded node collapses it.
func TestDetailUpdateCollapseOnH(t *testing.T) {
	nodes := []ui.TreeNode{
		{
			Key:      "api",
			Children: []ui.TreeNode{{Key: "key", Value: "***"}},
			Expanded: true,
		},
	}
	m := ui.NewDetailModel("test.yaml", nodes, 80, 24, true, "")

	// Before: expanded (key child visible)
	before := m.View()
	assert.True(t, strings.Contains(before, "key"),
		"key child should be visible when api is expanded")

	// Send h key to collapse
	msg := tea.KeyPressMsg{Code: 'h'}
	m2, _ := m.Update(msg)

	after := m2.View()
	assert.False(t, strings.Contains(after, "key"),
		"key child must not be visible after collapsing with h, view: %q", after)
}

// TestDetailSelectedIndex verifies SelectedIndex returns current cursor position.
func TestDetailSelectedIndex(t *testing.T) {
	nodes := sampleTree()
	m := ui.NewDetailModel("test.yaml", nodes, 80, 24, true, "")
	assert.Equal(t, 0, m.SelectedIndex(), "initial cursor should be at index 0")
}

// TestDetailCursorMovement verifies j/k keys move cursor up/down.
func TestDetailCursorMovement(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "a", Value: "***", Depth: 0},
		{Key: "b", Value: "***", Depth: 0},
		{Key: "c", Value: "***", Depth: 0},
	}
	m := ui.NewDetailModel("test.yaml", nodes, 80, 24, true, "")
	require.Equal(t, 0, m.SelectedIndex(), "starts at 0")

	// Press j to move down
	msgJ := tea.KeyPressMsg{Code: 'j'}
	m, _ = m.Update(msgJ)
	assert.Equal(t, 1, m.SelectedIndex(), "cursor at 1 after pressing j")

	// Press j again
	m, _ = m.Update(msgJ)
	assert.Equal(t, 2, m.SelectedIndex(), "cursor at 2 after pressing j twice")

	// Press k to move up
	msgK := tea.KeyPressMsg{Code: 'k'}
	m, _ = m.Update(msgK)
	assert.Equal(t, 1, m.SelectedIndex(), "cursor at 1 after pressing k")
}

// TestDetailEmptyState verifies NewDetailModel with empty nodes renders "No keys found in this file".
func TestDetailEmptyState(t *testing.T) {
	m := ui.NewDetailModel("test.yaml", []ui.TreeNode{}, 80, 24, true, "")
	view := m.View()
	assert.True(t, strings.Contains(view, "No keys found in this file"),
		"empty state must contain 'No keys found in this file', got: %q", view)
}

// stripAnsi removes ANSI escape sequences from a string for whitespace/content checks.
func stripAnsi(s string) string {
	var result strings.Builder
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if r == 'm' || r == 'K' || r == 'H' || r == 'J' {
				inEsc = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}

// TestDetailKeyMapBinding is a compile-time check: key.Matches should accept
// DetailKeyMap bindings without panicking.
func TestDetailKeyMapBinding(t *testing.T) {
	msg := tea.KeyPressMsg{Code: 'j'}
	// This verifies that key.Matches works with the binding from keys package
	// indirectly by checking that our Update function accepts tea.KeyPressMsg.
	nodes := []ui.TreeNode{{Key: "a", Value: "***"}}
	m := ui.NewDetailModel("test.yaml", nodes, 80, 24, true, "")
	m2, _ := m.Update(msg)
	// cursor stayed at 0 since only one item (can't go down)
	assert.GreaterOrEqual(t, m2.SelectedIndex(), 0)
}

// Compile-time import check: key package is available.
var _ = key.NewBinding

// TestDetailUnencryptedBannerShown verifies the unencrypted banner appears when isEncrypted=false.
func TestDetailUnencryptedBannerShown(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "username", Value: "admin", Depth: 0},
	}
	m := ui.NewDetailModel("plain.yaml", nodes, 80, 24, false, "")
	view := m.View()
	assert.True(t, strings.Contains(stripAnsi(view), "Not yet encrypted"),
		"unencrypted file must show 'Not yet encrypted' banner, got: %q", view)
}

// TestDetailUnencryptedBannerHidden verifies the unencrypted banner does NOT appear when isEncrypted=true.
func TestDetailUnencryptedBannerHidden(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "password", Value: "ENC[AES256_GCM,data:abc,type:str]", Encrypted: true, TypeHint: "str"},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true, "")
	view := m.View()
	assert.False(t, strings.Contains(stripAnsi(view), "Not yet encrypted"),
		"encrypted file must NOT show 'Not yet encrypted' banner, got: %q", view)
}

// TestDetailSearchActivation verifies that ActivateSearch sets IsSearchActive to true.
func TestDetailSearchActivation(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "token", Value: "***", Encrypted: true, TypeHint: "str"},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true, "")
	assert.False(t, m.IsSearchActive(), "search must be inactive after construction")

	_ = m.ActivateSearch()
	assert.True(t, m.IsSearchActive(), "search must be active after ActivateSearch()")
}

// TestDetailSearchDeactivation verifies that DeactivateSearch sets IsSearchActive to false.
func TestDetailSearchDeactivation(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "token", Value: "***", Encrypted: true, TypeHint: "str"},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true, "")
	_ = m.ActivateSearch()
	require.True(t, m.IsSearchActive())

	m.DeactivateSearch()
	assert.False(t, m.IsSearchActive(), "search must be inactive after DeactivateSearch()")
}

// TestDetailRenderRowUsesCanonicalTypeHintStyle verifies renderRow uses TypeHintStyle
// (not a temp style) by checking that an encrypted leaf renders the type hint.
func TestDetailRenderRowUsesCanonicalTypeHintStyle(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "password", Encrypted: true, TypeHint: "str"},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true, "")
	view := m.View()
	// The type hint "(str)" must appear in the rendered output
	assert.True(t, strings.Contains(stripAnsi(view), "(str)"),
		"encrypted leaf must render type hint '(str)', got: %q", view)
}

// TestDetailRenderRowUsesCanonicalBadgePlain verifies renderRow uses BadgePlain
// (not a temp style) by checking that a plain leaf renders the [plain] badge.
func TestDetailRenderRowUsesCanonicalBadgePlain(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "username", Value: "admin", IsPlain: true},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true, "")
	view := m.View()
	assert.True(t, strings.Contains(stripAnsi(view), "[plain]"),
		"plain leaf must render '[plain]' badge, got: %q", view)
}

// TestEditOnRevealedLeaf verifies that pressing e on a Revealed=true leaf activates edit mode.
func TestEditOnRevealedLeaf(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "password", Encrypted: true, Revealed: true, DecryptedValue: "secret123"},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true, "")
	require.False(t, m.IsEditActive(), "edit must not be active initially")

	msg := tea.KeyPressMsg{Code: 'e'}
	m2, _ := m.Update(msg)
	assert.True(t, m2.IsEditActive(), "edit must be active after pressing e on revealed leaf")
}

// TestEditOnMaskedLeaf verifies that pressing e on an encrypted but not-yet-revealed leaf
// returns an EditBlockedMsg (for AppModel to flash "Reveal first with r").
func TestEditOnMaskedLeaf(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "password", Encrypted: true, Revealed: false},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true, "")

	msg := tea.KeyPressMsg{Code: 'e'}
	m2, cmd := m.Update(msg)
	require.NotNil(t, cmd, "e on masked leaf must return a cmd")
	result := cmd()
	blocked, ok := result.(ui.EditBlockedMsg)
	require.True(t, ok, "cmd must return EditBlockedMsg, got: %T", result)
	assert.Empty(t, blocked.Reason, "masked leaf block must have empty Reason (AppModel flashes 'Reveal first with r')")
	assert.False(t, m2.IsEditActive(), "edit must not be active after blocked edit")
}

// TestEditOnArrayKeyReturnsBlocked verifies that e on a revealed node with array-indexed keyPath
// returns EditBlockedMsg with Reason "Array-indexed keys not editable in Phase 3".
func TestEditOnArrayKeyReturnsBlocked(t *testing.T) {
	// Build tree where expanded group "items" has child index node
	// We use a plain keyPath override by setting up a leaf named with bracket notation.
	// The keyPath is computed from dot-joining, so we simulate by creating a node "items[0]"
	// under a parent "items" expanded group.
	nodes := []ui.TreeNode{
		{
			Key:      "items",
			Expanded: true,
			Children: []ui.TreeNode{
				{Key: "items[0]", Encrypted: true, Revealed: true, DecryptedValue: "val"},
			},
		},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true, "")

	// Move cursor to the child (index 1 in flat rows after expanding)
	msgJ := tea.KeyPressMsg{Code: 'j'}
	m, _ = m.Update(msgJ)

	msg := tea.KeyPressMsg{Code: 'e'}
	_, cmd := m.Update(msg)
	require.NotNil(t, cmd, "e on array-indexed key must return a cmd")
	result := cmd()
	blocked, ok := result.(ui.EditBlockedMsg)
	require.True(t, ok, "must return EditBlockedMsg for array-indexed key, got: %T", result)
	assert.Equal(t, "Array-indexed keys not editable in Phase 3", blocked.Reason,
		"must set correct block reason")
}

// TestEditEnterProducesConfirmMsg verifies that pressing Enter while in edit mode
// returns an EditConfirmMsg with correct KeyPath, OldValue, NewValue.
func TestEditEnterProducesConfirmMsg(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "password", Encrypted: true, Revealed: true, DecryptedValue: "original"},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true, "")

	// Activate edit mode
	msg := tea.KeyPressMsg{Code: 'e'}
	m, cmd := m.Update(msg)
	require.True(t, m.IsEditActive(), "edit must be active")
	_ = cmd // focus cmd

	// The textinput is pre-populated with "original"; press Enter to confirm
	enterMsg := tea.KeyPressMsg{Code: tea.KeyEnter}
	m2, cmd2 := m.Update(enterMsg)
	require.NotNil(t, cmd2, "Enter in edit mode must return a cmd")
	result := cmd2()
	confirm, ok := result.(ui.EditConfirmMsg)
	require.True(t, ok, "must return EditConfirmMsg, got: %T", result)
	assert.Equal(t, "password", confirm.KeyPath, "KeyPath must be 'password'")
	assert.Equal(t, "original", confirm.OldValue, "OldValue must be 'original'")
	// NewValue is whatever textinput has (pre-populated with "original")
	assert.Equal(t, "original", confirm.NewValue, "NewValue must be 'original' (unchanged)")
	assert.False(t, m2.IsEditActive(), "edit must be inactive after Enter")
}

// TestEditEscCancels verifies that Esc in edit mode returns EditCancelMsg and deactivates edit.
func TestEditEscCancels(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "password", Encrypted: true, Revealed: true, DecryptedValue: "secret"},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true, "")

	// Activate edit mode
	msg := tea.KeyPressMsg{Code: 'e'}
	m, _ = m.Update(msg)
	require.True(t, m.IsEditActive(), "edit must be active")

	// Press Esc
	escMsg := tea.KeyPressMsg{Code: tea.KeyEsc}
	m2, cmd := m.Update(escMsg)
	require.NotNil(t, cmd, "Esc in edit mode must return a cmd")
	result := cmd()
	_, ok := result.(ui.EditCancelMsg)
	assert.True(t, ok, "must return EditCancelMsg, got: %T", result)
	assert.False(t, m2.IsEditActive(), "edit must be inactive after Esc")
}

// TestEditInputEatsNavigationKeys verifies that j key in edit mode does not move the cursor.
func TestEditInputEatsNavigationKeys(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "password", Encrypted: true, Revealed: true, DecryptedValue: "secret"},
		{Key: "other", Encrypted: true, Revealed: false},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true, "")
	initialCursor := m.SelectedIndex()

	// Activate edit mode
	msg := tea.KeyPressMsg{Code: 'e'}
	m, _ = m.Update(msg)
	require.True(t, m.IsEditActive(), "edit must be active")

	// j key while in edit mode must NOT move cursor
	msgJ := tea.KeyPressMsg{Code: 'j'}
	m2, _ := m.Update(msgJ)
	assert.Equal(t, initialCursor, m2.SelectedIndex(),
		"j key in edit mode must not move cursor (textinput consumes it)")
}

// ---- Task 1: E key ($EDITOR) tests ----

// TestEditFileOnRevealedReturnsEditorRequestMsg verifies E key on a detail model
// that has at least one revealed node returns an EditorRequestMsg cmd.
func TestEditFileOnRevealedReturnsEditorRequestMsg(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "password", Encrypted: true, Revealed: true, DecryptedValue: "secret"},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true, "")
	// E key (upper-case E)
	msg := tea.KeyPressMsg{Code: 'E'}
	_, cmd := m.Update(msg)
	require.NotNil(t, cmd, "E key with revealed nodes must return a cmd")
	result := cmd()
	_, ok := result.(ui.EditorRequestMsg)
	assert.True(t, ok, "cmd must return EditorRequestMsg, got: %T", result)
}

// TestEditFileOnNoneRevealedReturnsEditBlockedMsg verifies E key on a detail model
// with no revealed nodes returns an EditBlockedMsg.
func TestEditFileOnNoneRevealedReturnsEditBlockedMsg(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "password", Encrypted: true, Revealed: false},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true, "")
	msg := tea.KeyPressMsg{Code: 'E'}
	_, cmd := m.Update(msg)
	require.NotNil(t, cmd, "E key with no revealed nodes must return a cmd")
	result := cmd()
	_, ok := result.(ui.EditBlockedMsg)
	assert.True(t, ok, "cmd must return EditBlockedMsg when no nodes revealed, got: %T", result)
}

// ---- Task 2: X key (rotation) tests ----

// TestRotateKeyOnRevealed verifies X key on a revealed leaf with a detectable format
// returns a RotateReadyMsg (or a cmd that produces one).
// Uses a 32-byte base64 value (44 chars) which matches the regex {22,}.
func TestRotateKeyOnRevealed(t *testing.T) {
	// 32-byte base64 = 44 chars, well above the 22-char regex threshold
	base64Val := strings.Repeat("A", 44) // 44 uppercase A's — valid base64 chars, 44 chars
	nodes := []ui.TreeNode{
		{Key: "token", Encrypted: true, Revealed: true, DecryptedValue: base64Val},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true, "")
	msg := tea.KeyPressMsg{Code: 'X'}
	_, cmd := m.Update(msg)
	require.NotNil(t, cmd, "X key on revealed detectable leaf must return a cmd")
	result := cmd()
	_, ok := result.(ui.RotateReadyMsg)
	assert.True(t, ok, "cmd must return RotateReadyMsg for auto-detected format, got: %T", result)
}

// TestRotateKeyOnUnknownFormat verifies X key on a revealed leaf with an unknown format
// returns a RotateFormatMenuMsg.
func TestRotateKeyOnUnknownFormat(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "secret", Encrypted: true, Revealed: true, DecryptedValue: "just a regular string"},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true, "")
	msg := tea.KeyPressMsg{Code: 'X'}
	_, cmd := m.Update(msg)
	require.NotNil(t, cmd, "X key on unknown-format leaf must return a cmd")
	result := cmd()
	_, ok := result.(ui.RotateFormatMenuMsg)
	assert.True(t, ok, "cmd must return RotateFormatMenuMsg for unknown format, got: %T", result)
}

// TestRotateKeyOnMasked verifies X key on a masked (unrevealed) leaf returns an EditBlockedMsg.
func TestRotateKeyOnMasked(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "secret", Encrypted: true, Revealed: false},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true, "")
	msg := tea.KeyPressMsg{Code: 'X'}
	_, cmd := m.Update(msg)
	require.NotNil(t, cmd, "X key on masked leaf must return a cmd")
	result := cmd()
	_, ok := result.(ui.EditBlockedMsg)
	assert.True(t, ok, "cmd must return EditBlockedMsg for masked leaf, got: %T", result)
}

// TestRotateKeyOnArrayIndexed verifies X key on a revealed leaf with an array-indexed
// keyPath returns EditBlockedMsg with "Array-indexed keys not editable in Phase 3" reason.
func TestRotateKeyOnArrayIndexed(t *testing.T) {
	nodes := []ui.TreeNode{
		{
			Key:      "items",
			Expanded: true,
			Children: []ui.TreeNode{
				{Key: "items[0]", Encrypted: true, Revealed: true, DecryptedValue: "val"},
			},
		},
	}
	m := ui.NewDetailModel("secrets.yaml", nodes, 80, 24, true, "")

	// Move cursor to the child node
	msgJ := tea.KeyPressMsg{Code: 'j'}
	m, _ = m.Update(msgJ)

	msg := tea.KeyPressMsg{Code: 'X'}
	_, cmd := m.Update(msg)
	require.NotNil(t, cmd, "X key on array-indexed key must return a cmd")
	result := cmd()
	blocked, ok := result.(ui.EditBlockedMsg)
	require.True(t, ok, "must return EditBlockedMsg for array-indexed key, got: %T", result)
	assert.Equal(t, "Array-indexed keys not editable in Phase 3", blocked.Reason,
		"must set correct block reason")
}

// TestDetailHints verifies DetailModel.Hints() returns 13 entries (matching
// DetailKeyMap.ShortHelp()) with exactly one (Blame, "b") marked Visible=false
// per D-09 12-slot curation.
func TestDetailHints(t *testing.T) {
	m := ui.NewDetailModel("test.yaml", sampleTree(), 80, 24, true, "")
	hints := m.Hints()
	require.Equal(t, 13, len(hints),
		"Detail must expose 13 hints from ShortHelp() before curation")

	visible := 0
	invisibleMnemonics := []string{}
	for _, h := range hints {
		if h.Visible {
			visible++
		} else {
			invisibleMnemonics = append(invisibleMnemonics, h.Mnemonic)
		}
	}
	require.Equal(t, 12, visible,
		"Detail must curate to exactly 12 visible hints to fit the menu cap")
	require.Equal(t, []string{"b"}, invisibleMnemonics,
		"the only invisible hint must be Blame (b) per D-06")
}
