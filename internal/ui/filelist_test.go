package ui_test

import (
	"strings"
	"testing"

	"github.com/caesarakalaeii/sops-tui/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFileListNewModelItemCount verifies that NewFileListModel with items returns a model
// where ItemCount() matches the input length.
func TestFileListNewModelItemCount(t *testing.T) {
	items := []ui.FileItem{
		{Name: "secrets/prod.yaml", Path: "/repo/secrets/prod.yaml"},
		{Name: "secrets/staging.yaml", Path: "/repo/secrets/staging.yaml"},
		{Name: "config/db.yaml", Path: "/repo/config/db.yaml"},
	}
	m := ui.NewFileListModel(items, 80, 24)
	require.Equal(t, 3, m.ItemCount(), "ItemCount must equal number of input items")
}

// TestFileListItemFilterValue verifies FileItem implements list.Item by checking
// FilterValue returns the filename.
func TestFileListItemFilterValue(t *testing.T) {
	item := ui.FileItem{Name: "secrets/prod.yaml", Path: "/repo/secrets/prod.yaml"}
	assert.Equal(t, "secrets/prod.yaml", item.FilterValue(), "FilterValue must return the filename (Name field)")
}

// TestFileListEmptyStateView verifies that FileListModel.View() with zero items
// renders the "No SOPS files found" empty state text per UI-SPEC copywriting contract.
func TestFileListEmptyStateView(t *testing.T) {
	m := ui.NewFileListModel([]ui.FileItem{}, 80, 24)
	view := m.View()
	assert.True(t, strings.Contains(view, "No SOPS files found"),
		"empty state must contain 'No SOPS files found', got: %q", view)
}

// TestFileListSetSize verifies that SetSize correctly updates internal list dimensions
// without panicking.
func TestFileListSetSize(t *testing.T) {
	items := []ui.FileItem{
		{Name: "a.yaml", Path: "/a.yaml"},
	}
	m := ui.NewFileListModel(items, 80, 24)
	// Should not panic
	m.SetSize(120, 40)
	assert.Equal(t, 1, m.ItemCount(), "ItemCount unchanged after SetSize")
}

// TestFileListSelectedItem verifies SelectedItem returns the currently selected FileItem
// after construction (first item is selected by default).
func TestFileListSelectedItem(t *testing.T) {
	items := []ui.FileItem{
		{Name: "first.yaml", Path: "/first.yaml"},
		{Name: "second.yaml", Path: "/second.yaml"},
	}
	m := ui.NewFileListModel(items, 80, 24)
	got, ok := m.SelectedItem()
	require.True(t, ok, "SelectedItem must return ok=true when items are present")
	assert.Equal(t, "first.yaml", got.Name, "first item should be selected after construction")
}

// TestFileListSelectedItemEmptyList verifies SelectedItem returns ok=false for empty list.
func TestFileListSelectedItemEmptyList(t *testing.T) {
	m := ui.NewFileListModel([]ui.FileItem{}, 80, 24)
	_, ok := m.SelectedItem()
	assert.False(t, ok, "SelectedItem must return ok=false when list is empty")
}

// TestFileListItemInterfaces verifies FileItem implements list.Item (compile-time check
// via interface assignment). If this compiles, the interface is satisfied.
func TestFileListItemInterfaces(t *testing.T) {
	item := ui.FileItem{Name: "test.yaml", Path: "/test.yaml"}
	// Compile-time check: FileItem must satisfy list.Item (FilterValue), list.DefaultItem (Title, Description)
	assert.Equal(t, "test.yaml", item.FilterValue())
	assert.Equal(t, "test.yaml", item.Title())
	assert.Equal(t, "/test.yaml", item.Description())
}
