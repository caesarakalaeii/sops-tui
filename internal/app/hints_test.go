// Package app — Phase 7 menuHints() dispatcher matrix and titleForState() table tests.
//
// Verifies every branch of the (state, recipientAction-via-state, IsSearchActive)
// dispatcher per UI-SPEC §"Hints dispatch tuple" and every D-15 title string.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
)

// buildAppModel constructs a sized AppModel for dispatcher tests —
// mirrors the defaultEnvInternal + WindowSizeMsg pattern from layout_test.go.
func buildAppModel(t *testing.T) AppModel {
	t.Helper()
	m := NewAppModel(defaultEnvInternal(), "")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return updated.(AppModel)
}

// TestMenuHints_StateFileList_NoSearch — stateFileList default arm.
func TestMenuHints_StateFileList_NoSearch(t *testing.T) {
	m := buildAppModel(t)
	// state defaults to stateFileList; IsSearchActive defaults to false.
	hints := m.menuHints()
	require.Equal(t, 12, len(hints), "FileList.Hints() returns 12 entries (10 ShortHelp + g/G append)")
	require.Equal(t, "k/↑", hints[0].Mnemonic, "first hint is Up per DefaultFileListKeyMap.ShortHelp order")
}

// TestMenuHints_StateFileList_SearchActive — D-11 override.
//
// Two-tier strategy per Plan 3 Task 3 fallback notes:
//  1. Primary path — push '/' KeyPressMsg through Update; if FileListModel
//     activates inline search the override applies.
//  2. Fallback — if step 1 doesn't activate search after a single dispatch,
//     reach into m.fileList directly via the pointer-receiver
//     ActivateSearch() round-trip pattern.
func TestMenuHints_StateFileList_SearchActive(t *testing.T) {
	m := buildAppModel(t)
	// Need at least one file for the / key to take effect in the file list view.
	updated, _ := m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	m = updated.(AppModel)

	if !m.fileList.IsSearchActive() {
		// Fallback: direct setter via pointer round-trip.
		fl := m.fileList
		_ = (&fl).ActivateSearch()
		m.fileList = fl
	}
	require.True(t, m.fileList.IsSearchActive(),
		"precondition: search must be active before asserting hint override")

	hints := m.menuHints()
	require.Equal(t, keys.FileListSearchHints, hints,
		"search-active override must return keys.FileListSearchHints verbatim (D-11)")
}

// TestMenuHints_StateDetail — delegation to DetailModel.Hints().
func TestMenuHints_StateDetail(t *testing.T) {
	m := buildAppModel(t)
	m.state = stateDetail
	hints := m.menuHints()
	require.Equal(t, 13, len(hints), "Detail.Hints() returns 13 entries (one Visible=false)")
}

// TestMenuHints_StateMetadata — delegation.
func TestMenuHints_StateMetadata(t *testing.T) {
	m := buildAppModel(t)
	m.state = stateMetadata
	hints := m.menuHints()
	require.Equal(t, 5, len(hints))
}

// TestMenuHints_StateDiff — delegation.
func TestMenuHints_StateDiff(t *testing.T) {
	m := buildAppModel(t)
	m.state = stateDiff
	hints := m.menuHints()
	require.Equal(t, 6, len(hints))
	require.Equal(t, "y", hints[0].Mnemonic)
	require.Equal(t, "confirm re-encrypt", hints[0].Description)
}

// TestMenuHints_StateRecipientConfirm — inline package-var dispatch.
func TestMenuHints_StateRecipientConfirm(t *testing.T) {
	m := buildAppModel(t)
	m.state = stateRecipientConfirm
	hints := m.menuHints()
	require.Equal(t, keys.RecipientConfirmHints, hints)
}

// TestMenuHints_StateBulkReKeyConfirm — inline package-var dispatch.
func TestMenuHints_StateBulkReKeyConfirm(t *testing.T) {
	m := buildAppModel(t)
	m.state = stateBulkReKeyConfirm
	hints := m.menuHints()
	require.Equal(t, keys.BulkReKeyConfirmHints, hints)
}

// TestMenuHints_StateHelp — delegation.
func TestMenuHints_StateHelp(t *testing.T) {
	m := buildAppModel(t)
	m.state = stateHelp
	hints := m.menuHints()
	require.Equal(t, 3, len(hints))
}

// TestMenuHints_StateHistory — delegation.
func TestMenuHints_StateHistory(t *testing.T) {
	m := buildAppModel(t)
	m.state = stateHistory
	hints := m.menuHints()
	require.Equal(t, 5, len(hints))
}

// TestMenuHints_StateHealth — delegation.
func TestMenuHints_StateHealth(t *testing.T) {
	m := buildAppModel(t)
	m.state = stateHealth
	hints := m.menuHints()
	require.Equal(t, 5, len(hints))
}

// TestMenuHints_StateRecipientForm — delegation.
func TestMenuHints_StateRecipientForm(t *testing.T) {
	m := buildAppModel(t)
	m.state = stateRecipientForm
	hints := m.menuHints()
	require.Equal(t, 2, len(hints))
}

// TestMenuHints_StateRecipientList — Pitfall 3 inline.
func TestMenuHints_StateRecipientList(t *testing.T) {
	m := buildAppModel(t)
	m.state = stateRecipientList
	hints := m.menuHints()
	require.Equal(t, keys.RecipientListHints, hints)
}

// TestMenuHints_StateFormatMenu — D-09 inline.
func TestMenuHints_StateFormatMenu(t *testing.T) {
	m := buildAppModel(t)
	m.state = stateFormatMenu
	hints := m.menuHints()
	require.Equal(t, keys.FormatMenuHints, hints)
}

// TestMenuHints_DefaultArm_ReturnsNil — defensive: an unknown state yields nil.
// No state value in the enum should trigger this, but the default arm is there
// as a safety fallback per D-10. Use a made-up state value cast from int.
func TestMenuHints_DefaultArm_ReturnsNil(t *testing.T) {
	m := buildAppModel(t)
	// sessionState is a typed int; cast a high value that is not any defined constant.
	m.state = sessionState(9999)
	hints := m.menuHints()
	require.Nil(t, hints, "unknown state must return nil (D-10 default arm)")
}

// TestTitleForState_AllStates table-drives the D-15 title map.
func TestTitleForState_AllStates(t *testing.T) {
	cases := []struct {
		name     string
		setup    func(m *AppModel)
		expected string
	}{
		{"fileList", func(m *AppModel) { m.state = stateFileList }, "Files (0)"},
		{"detail", func(m *AppModel) {
			m.state = stateDetail
			// m.currentFile.Name defaults to "" — title becomes "Detail: "
		}, "Detail: "},
		{"metadata", func(m *AppModel) { m.state = stateMetadata }, "Metadata"},
		{"diff", func(m *AppModel) { m.state = stateDiff }, "Diff"},
		{"recipientConfirm", func(m *AppModel) { m.state = stateRecipientConfirm }, "Diff"},
		{"bulkReKeyConfirm", func(m *AppModel) { m.state = stateBulkReKeyConfirm }, "Diff"},
		{"help", func(m *AppModel) { m.state = stateHelp }, "Help"},
		{"history", func(m *AppModel) { m.state = stateHistory }, "History (0)"},
		{"health", func(m *AppModel) { m.state = stateHealth }, "Health (0 findings)"},
		{"recipientList", func(m *AppModel) { m.state = stateRecipientList }, "Recipients (0)"},
		{"recipientForm", func(m *AppModel) { m.state = stateRecipientForm }, "RecipientForm"},
		{"formatMenu", func(m *AppModel) { m.state = stateFormatMenu }, "Format"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := buildAppModel(t)
			tc.setup(&m)
			got := m.titleForState()
			require.Equal(t, tc.expected, got,
				"title for %s must match D-15 map", tc.name)
		})
	}
}
