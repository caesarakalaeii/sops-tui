package ui_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/caesarakalaeii/sops-tui/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper: build a model with three nested files for tree-shape assertions.
func newNestedTreeModel(t *testing.T) ui.FileListModel {
	t.Helper()
	items := []ui.FileItem{
		{Name: "secrets/base/alpha.enc.yaml", Path: "/r/secrets/base/alpha.enc.yaml", IsEncrypted: true},
		{Name: "secrets/base/beta.enc.yaml", Path: "/r/secrets/base/beta.enc.yaml", IsEncrypted: true},
		{Name: "apps/foo/keel.enc.yaml", Path: "/r/apps/foo/keel.enc.yaml", IsEncrypted: true},
	}
	return ui.NewFileListModel(items, 80, 24)
}

// TestTree_DropsAbsolutePathFromRows verifies that no row in the rendered
// view contains the file's absolute Path — that was the "useless total path"
// the user complained about.
func TestTree_DropsAbsolutePathFromRows(t *testing.T) {
	m := newNestedTreeModel(t)
	view := stripAnsi(m.View())
	assert.NotContains(t, view, "/r/secrets/base/alpha.enc.yaml",
		"absolute path must not appear in any rendered row")
	assert.NotContains(t, view, "/r/apps/foo/keel.enc.yaml",
		"absolute path must not appear in any rendered row")
}

// TestTree_RendersDirsAsBranches verifies the rendered view contains the
// directory names as standalone tree branches with the ▾ expanded indicator.
// Single-child chains are compacted: apps→foo (foo has 1 file, apps has no
// other children) renders as one row "▾ apps/foo/".
func TestTree_RendersDirsAsBranches(t *testing.T) {
	m := newNestedTreeModel(t)
	view := stripAnsi(m.View())
	assert.Contains(t, view, "▾ apps/foo/", "linear chain apps→foo must compact")
	assert.Contains(t, view, "▾ secrets/base/", "linear chain secrets→base must compact")
	// Sanity: no standalone single-segment intermediate rows.
	assert.NotRegexp(t, `(?m)^.*▾ apps/$`, "apps/ must not appear as its own row")
	assert.NotRegexp(t, `(?m)^.*▾ secrets/$`, "secrets/ must not appear as its own row")
}

// TestTree_DirsBeforeFilesAlphabetical verifies sort order at each level:
// dirs alphabetical first, then files alphabetical. (Single-level dirs that
// directly contain files are not subject to chain compaction.)
func TestTree_DirsBeforeFilesAlphabetical(t *testing.T) {
	items := []ui.FileItem{
		{Name: "zz-root.yaml", Path: "/r/zz-root.yaml", IsEncrypted: true},
		{Name: "aaa/x.yaml", Path: "/r/aaa/x.yaml", IsEncrypted: true},
		{Name: "secrets/y.yaml", Path: "/r/secrets/y.yaml", IsEncrypted: true},
		{Name: "aa-root.yaml", Path: "/r/aa-root.yaml", IsEncrypted: true},
	}
	m := ui.NewFileListModel(items, 100, 24)
	view := stripAnsi(m.View())
	aaaIdx := strings.Index(view, "▾ aaa/")
	secretsIdx := strings.Index(view, "▾ secrets/")
	aaRootIdx := strings.Index(view, "aa-root.yaml")
	zzRootIdx := strings.Index(view, "zz-root.yaml")
	require.True(t, aaaIdx >= 0 && secretsIdx >= 0 && aaRootIdx >= 0 && zzRootIdx >= 0,
		"all four entries must appear in view")
	assert.Less(t, aaaIdx, secretsIdx, "aaa/ must precede secrets/")
	assert.Less(t, secretsIdx, aaRootIdx, "all dirs must precede root files")
	assert.Less(t, aaRootIdx, zzRootIdx, "root files must be alphabetical")
}

// TestTree_CompactsDeepLinearChain verifies that a chain of single-child dirs
// 4+ levels deep collapses into a single row with the joined path.
func TestTree_CompactsDeepLinearChain(t *testing.T) {
	items := []ui.FileItem{
		{Name: "apps/workloads/caesar-website/secrets/keel.enc.yaml", Path: "/r/keel.enc.yaml", IsEncrypted: true},
	}
	m := ui.NewFileListModel(items, 100, 24)
	view := stripAnsi(m.View())
	assert.Contains(t, view, "▾ apps/workloads/caesar-website/secrets/",
		"4-deep linear chain must render as one merged row")
	assert.Contains(t, view, "keel.enc.yaml", "file under merged chain must be visible")
	// Intermediate dirs must NOT appear as their own rows.
	assert.NotContains(t, view, "▾ workloads/", "intermediate workloads/ must not have its own row")
	assert.NotContains(t, view, "▾ caesar-website/", "intermediate caesar-website/ must not have its own row")
}

// TestTree_DoesNotCompactWhenSiblingFiles verifies that a dir with both a
// subdir AND files of its own is NOT compacted with its child.
func TestTree_DoesNotCompactWhenSiblingFiles(t *testing.T) {
	items := []ui.FileItem{
		{Name: "apps/own.yaml", Path: "/r/apps/own.yaml", IsEncrypted: true},
		{Name: "apps/sub/inner.yaml", Path: "/r/apps/sub/inner.yaml", IsEncrypted: true},
	}
	m := ui.NewFileListModel(items, 100, 24)
	view := stripAnsi(m.View())
	assert.Contains(t, view, "▾ apps/", "apps/ must render as its own row when it has files")
	assert.Contains(t, view, "▾ sub/", "sub/ must render as its own row")
	assert.NotContains(t, view, "▾ apps/sub/", "compacted form must not appear when apps has files")
}

// TestTree_DoesNotCompactWhenBranching verifies that a dir with multiple
// subdirs is not compacted with any of them.
func TestTree_DoesNotCompactWhenBranching(t *testing.T) {
	items := []ui.FileItem{
		{Name: "apps/a/x.yaml", Path: "/r/apps/a/x.yaml", IsEncrypted: true},
		{Name: "apps/b/y.yaml", Path: "/r/apps/b/y.yaml", IsEncrypted: true},
	}
	m := ui.NewFileListModel(items, 100, 24)
	view := stripAnsi(m.View())
	assert.Contains(t, view, "▾ apps/", "branching dir must render as its own row")
	assert.Contains(t, view, "▾ a/", "child dir under branching parent must render separately")
	assert.Contains(t, view, "▾ b/", "child dir under branching parent must render separately")
}


// TestTree_EnterOnDirCollapses verifies that pressing Enter on a directory
// row toggles its collapse state, hiding its children.
func TestTree_EnterOnDirCollapses(t *testing.T) {
	m := newNestedTreeModel(t)
	// Cursor starts on the first FILE under apps/foo (alphabetical "apps" first).
	// Move up to land on the apps/foo dir, then up again to apps/. We can also
	// use g (GoTop).
	m, _ = m.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})
	view0 := stripAnsi(m.View())
	require.Contains(t, view0, "▾ apps/foo/", "apps/foo/ must start expanded (compacted chain)")
	require.Contains(t, view0, "keel.enc.yaml", "keel.enc.yaml visible while expanded")

	// First row in the rendered view is "apps/" (alphabetical first dir).
	// Selection defaults to the first FILE; refreshRows fallback selects
	// the first file row when items are present. To toggle apps/, move the
	// cursor up until it hits the dir row. Easiest: send GoTop and confirm
	// the cursor is at index 0 (which is the apps/ dir row).
	m, _ = m.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})

	// Send Enter — should toggle collapse on the merged apps/foo/ row.
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter, Text: ""})
	view1 := stripAnsi(m.View())
	assert.Contains(t, view1, "▸ apps/foo/", "apps/foo/ must show collapsed indicator after Enter")
	assert.NotContains(t, view1, "keel.enc.yaml",
		"children of apps/foo/ must be hidden when collapsed")
	// Other dirs unaffected.
	assert.Contains(t, view1, "▾ secrets/base/", "secrets/base/ must still be expanded")
}

// TestTree_FilterAutoExpandsAncestors verifies that activating search and
// typing a pattern uncollapses dirs that contain matches.
func TestTree_FilterAutoExpandsAncestors(t *testing.T) {
	m := newNestedTreeModel(t)

	// Pre-collapse everything via raw mutation through Enter on the first dir,
	// then via search just typing — easier to just collapse with Enter, then
	// start search.
	m, _ = m.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter, Text: ""}) // collapse apps/foo/
	view0 := stripAnsi(m.View())
	require.NotContains(t, view0, "keel.enc.yaml", "precondition: keel hidden")

	// Activate search and type "keel"
	_ = m.ActivateSearch()
	for _, r := range "keel" {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	view1 := stripAnsi(m.View())
	assert.Contains(t, view1, "keel.enc.yaml",
		"filter must auto-expand ancestors so 'keel' becomes visible")
}

// TestTree_FilterRestoresCollapseStateOnEsc verifies that on Esc the tree
// restores exactly the collapse state present when the search was activated.
func TestTree_FilterRestoresCollapseStateOnEsc(t *testing.T) {
	m := newNestedTreeModel(t)
	m, _ = m.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter, Text: ""}) // collapse apps/foo/

	require.Contains(t, stripAnsi(m.View()), "▸ apps/foo/",
		"precondition: apps/foo/ collapsed")

	_ = m.ActivateSearch()
	for _, r := range "keel" {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	require.Contains(t, stripAnsi(m.View()), "keel.enc.yaml",
		"precondition: filter expanded apps/ to show keel")

	// Esc deactivates search and should restore the original collapse state.
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc, Text: ""})
	view := stripAnsi(m.View())
	assert.Contains(t, view, "▸ apps/foo/",
		"after Esc, apps/foo/ must be collapsed again per restored snapshot")
	assert.NotContains(t, view, "keel.enc.yaml",
		"children must be hidden again after Esc")
}

// TestTree_DirSpaceRecursiveSelect verifies Space on a directory selects all
// files underneath it (and a second Space deselects them).
func TestTree_DirSpaceRecursiveSelect(t *testing.T) {
	m := newNestedTreeModel(t)
	require.Empty(t, m.SelectedItems(), "start with no selections")

	// Cursor to the apps/ dir row (first row).
	m, _ = m.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})
	// Press Space — should select keel.enc.yaml (the only file under apps/).
	m, _ = m.Update(tea.KeyPressMsg{Code: ' ', Text: " "})

	got := m.SelectedItems()
	require.Len(t, got, 1, "Space on apps/ must select its 1 file")
	assert.Equal(t, "apps/foo/keel.enc.yaml", got[0].Name)

	// Press Space again — should deselect.
	m, _ = m.Update(tea.KeyPressMsg{Code: ' ', Text: " "})
	assert.Empty(t, m.SelectedItems(),
		"second Space on apps/ must deselect the file")
}

// TestTree_RootLevelFilesUnderConnector verifies that files at the repo root
// (no directory) still render with a tree connector, just at depth 0.
func TestTree_RootLevelFilesUnderConnector(t *testing.T) {
	items := []ui.FileItem{
		{Name: "lone.enc.yaml", Path: "/r/lone.enc.yaml", IsEncrypted: true},
	}
	m := ui.NewFileListModel(items, 80, 24)
	view := stripAnsi(m.View())
	assert.Contains(t, view, "lone.enc.yaml")
	// Only one row → must be the "last" sibling so connector is └─.
	assert.Contains(t, view, "└─ lone.enc.yaml")
}

// TestTree_ItemCountIsFilesOnly verifies ItemCount returns the count of files,
// not the count of rendered rows (which includes directory rows).
func TestTree_ItemCountIsFilesOnly(t *testing.T) {
	m := newNestedTreeModel(t)
	assert.Equal(t, 3, m.ItemCount(),
		"ItemCount must return file count (3), not rendered rows (which include dirs)")
}

// TestTree_SelectedItemFalseOnDir verifies that when the cursor sits on a
// directory row, SelectedItem returns ok=false so AppModel doesn't try to
// open a "file" that doesn't exist.
func TestTree_SelectedItemFalseOnDir(t *testing.T) {
	m := newNestedTreeModel(t)
	// Cursor to the first dir row.
	m, _ = m.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})
	_, ok := m.SelectedItem()
	assert.False(t, ok, "cursor on dir → SelectedItem returns ok=false")
}
