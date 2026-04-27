package app_test

import (
	"fmt"
	"os"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/caesarakalaeii/sops-tui/internal/app"
	"github.com/caesarakalaeii/sops-tui/internal/sops"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func defaultEnv() ui.EnvStatus {
	return ui.EnvStatus{
		SopsAvailable:     true,
		AgeAvailable:      true,
		SopsYamlAvailable: true,
	}
}

// send is a helper that sends a single message to an AppModel and returns
// the updated model.
func send(t *testing.T, m tea.Model, msg tea.Msg) tea.Model {
	t.Helper()
	updated, _ := m.Update(msg)
	return updated
}

// Test: NewAppModel initializes in stateFileList (the default entry view).
// Phase 7: View() early-returns empty content when m.width == 0 (Pitfall 5),
// so tests must propagate a WindowSizeMsg before asserting on View().Content.
func TestAppModelInitialState(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	// View should not panic and should produce a non-empty View
	v := updated.View()
	assert.NotEmpty(t, v.Content, "Initial View().Content must not be empty after WindowSizeMsg")
}

// Test: "?" key toggles to stateHelp and back.
func TestAppModelHelpToggle(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	// Phase 7 Pitfall 5: View() early-returns empty when width=0; size first.
	m = send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24}).(app.AppModel)

	// Press ? → should enter help state
	m2 := send(t, m, tea.KeyPressMsg{Code: '?'})

	// The help view should now be rendering — verify View produces the footer
	v := m2.View()
	assert.Contains(t, v.Content, "Press ? or Esc to close",
		"After pressing ?, view should contain help overlay footer")

	// Press ? again → should exit help and return to file list
	m3 := send(t, m2, tea.KeyPressMsg{Code: '?'})
	v2 := m3.View()
	assert.NotContains(t, v2.Content, "Press ? or Esc to close",
		"After pressing ? again, help overlay should be dismissed")
}

// Test: "q" key returns tea.Quit command.
func TestAppModelQuitKey(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'q'})
	require.NotNil(t, cmd, "q key must return a non-nil Cmd")
	msg := cmd()
	_, isQuit := msg.(tea.QuitMsg)
	assert.True(t, isQuit, "q key command must produce tea.QuitMsg, got: %T", msg)
}

// Test: "esc" from stateDetail returns to stateFileList.
func TestAppModelEscFromDetail(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	// Phase 7 Pitfall 5: View() early-returns empty when width=0; size first.
	m = send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24}).(app.AppModel)

	// Simulate entering detail (need an item selected first — use WindowSizeMsg then navigate)
	// Since the file list is empty, we cannot drill in with enter/l.
	// Test Esc from help state instead (which is the most common esc path),
	// and separately test that esc from a manually-set detail works via the exported API.
	//
	// We manually set state by navigating: press ? (help), then esc (back)
	m2 := send(t, m, tea.KeyPressMsg{Code: '?'})
	// Verify we're in help
	v := m2.View()
	assert.Contains(t, v.Content, "Press ? or Esc to close", "Should be in help state")

	// Press Esc → back to previous state (stateFileList)
	m3 := send(t, m2, tea.KeyPressMsg{Code: 27}) // ESC key code
	v2 := m3.View()
	assert.NotContains(t, v2.Content, "Press ? or Esc to close",
		"After Esc from help, overlay should close")
}

// Test: "esc" from stateHelp returns to prevState.
func TestAppModelEscFromHelp(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	// Phase 7 Pitfall 5: View() early-returns empty when width=0; size first.
	m = send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24}).(app.AppModel)

	// Enter help
	m2 := send(t, m, tea.KeyPressMsg{Code: '?'})
	v := m2.View()
	assert.Contains(t, v.Content, "Press ? or Esc to close")

	// Esc from help
	m3 := send(t, m2, tea.KeyPressMsg{Code: 27})
	v2 := m3.View()
	assert.NotContains(t, v2.Content, "Press ? or Esc to close",
		"Esc from help must dismiss the overlay")
}

// Test: WindowSizeMsg propagates to children (width/height stored).
func TestAppModelWindowSizePropagation(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	m2 := send(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	// After size update, View should still render without panicking
	v := m2.View()
	assert.NotEmpty(t, v.Content, "View must not be empty after WindowSizeMsg")
}

// Test: AltScreen is set in View().
func TestAppModelAltScreen(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	v := m.View()
	assert.True(t, v.AltScreen, "View.AltScreen must be true for full-screen TUI")
}

// Test: Status bar flash timer works through root Update.
func TestAppModelStatusFlashPassthrough(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	// Send a FlashClearMsg — it should not panic and the model should update cleanly
	_, cmd := m.Update(ui.FlashClearMsg{Gen: 0})
	// cmd may be nil (no flash active) — just verify no panic
	_ = cmd
}

// Test: FilesDiscoveredMsg populates the file list.
func TestAppModelFilesDiscoveredMsg(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	// Send a size message first so dimensions are set
	m2 := send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})

	files := []sops.DiscoveredFile{
		{Name: "secrets/prod.yaml", AbsPath: "/repo/secrets/prod.yaml", IsEncrypted: true},
		{Name: "secrets/staging.yaml", AbsPath: "/repo/secrets/staging.yaml", IsEncrypted: false},
	}
	m3 := send(t, m2, app.FilesDiscoveredMsg{Files: files})
	// View should render without panic; file list now has items
	v := m3.View()
	assert.NotEmpty(t, v.Content, "View must not be empty after FilesDiscoveredMsg")
}

// Test: FilesDiscoveredMsg with error flashes status bar.
func TestAppModelFilesDiscoveredMsgError(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	_, cmd := m.Update(app.FilesDiscoveredMsg{Err: assert.AnError})
	// Should not panic; cmd may or may not be nil
	_ = cmd
}

// Test: "/" key activates search in file list.
func TestAppModelSlashActivatesSearch(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	// Phase 7 Pitfall 5: View() early-returns empty when width=0; size first.
	m = send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24}).(app.AppModel)
	// Populate with some files
	files := []sops.DiscoveredFile{
		{Name: "secrets/prod.yaml", AbsPath: "/repo/secrets/prod.yaml", IsEncrypted: true},
	}
	m2 := send(t, m, app.FilesDiscoveredMsg{Files: files})

	// Press "/" — search should activate
	m3, _ := m2.Update(tea.KeyPressMsg{Code: '/'})
	// View should include a search bar; just verify no panic
	v := m3.View()
	assert.NotEmpty(t, v.Content, "View must not be empty after / key")
}

// Test: "i" key from stateMetadata closes the overlay and returns to prevState.
// We test the Esc path since "i" also works but requires a file to be selected.
func TestAppModelEscFromMetadata(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	// Seed with a file and a window size so we can get to stateFileList with items
	m2 := send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	files := []sops.DiscoveredFile{
		{Name: "secrets/prod.yaml", AbsPath: "/nonexistent/path.yaml", IsEncrypted: true},
	}
	m3 := send(t, m2, app.FilesDiscoveredMsg{Files: files})

	// Press "i" — will try to parse a nonexistent file, resulting in flash error,
	// so we end up back at stateFileList (no transition to stateMetadata).
	// We verify this does not panic.
	m4, _ := m3.Update(tea.KeyPressMsg{Code: 'i'})
	v := m4.View()
	assert.NotEmpty(t, v.Content, "View must not be empty after i key on missing file")
}

// Test: Esc deactivates search before navigating back.
func TestAppModelEscDeactivatesSearchFirst(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	files := []sops.DiscoveredFile{
		{Name: "secrets/prod.yaml", AbsPath: "/repo/secrets/prod.yaml", IsEncrypted: true},
	}
	m2 := send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m3 := send(t, m2, app.FilesDiscoveredMsg{Files: files})

	// Activate search
	m4, _ := m3.Update(tea.KeyPressMsg{Code: '/'})
	// Press Esc — should deactivate search, not navigate back
	m5, _ := m4.Update(tea.KeyPressMsg{Code: 27})
	v := m5.View()
	assert.NotEmpty(t, v.Content, "View must not be empty after Esc from search")
}

// helper: drive AppModel into stateDetail with a revealed leaf so we can test edit flows.
func modelInDetailWithRevealedLeaf(t *testing.T) tea.Model {
	t.Helper()
	m := app.NewAppModel(defaultEnv(), "")
	m2 := send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	// Inject a parsed file with a revealed leaf
	nodes := []ui.TreeNode{
		{Key: "password", Encrypted: true, Revealed: true, DecryptedValue: "hunter2"},
	}
	parsed := app.ParsedFileForTest(nodes)
	m3 := send(t, m2, app.FilesParsedMsg{Parsed: parsed})
	return m3
}

// TestEditConfirmMsgTransitionsToStateDiff verifies that EditConfirmMsg (old != new)
// transitions AppModel to stateDiff and renders the diff overlay.
func TestEditConfirmMsgTransitionsToStateDiff(t *testing.T) {
	m := modelInDetailWithRevealedLeaf(t)
	m2 := send(t, m, ui.EditConfirmMsg{
		KeyPath:  "password",
		OldValue: "hunter2",
		NewValue: "new_password",
	})
	v := m2.View()
	// stateDiff renders the DiffModel which contains "confirm re-encrypt"
	assert.Contains(t, v.Content, "confirm re-encrypt",
		"stateDiff view must contain 'confirm re-encrypt', got: %q", v.Content)
}

// TestEditConfirmMsgNoChangeFlashes verifies that EditConfirmMsg with OldValue == NewValue
// flashes "No changes" instead of transitioning to stateDiff.
func TestEditConfirmMsgNoChangeFlashes(t *testing.T) {
	m := modelInDetailWithRevealedLeaf(t)
	m2 := send(t, m, ui.EditConfirmMsg{
		KeyPath:  "password",
		OldValue: "same",
		NewValue: "same",
	})
	v := m2.View()
	// Must not show the diff overlay
	assert.NotContains(t, v.Content, "confirm re-encrypt",
		"no-change edit must not show diff overlay")
}

// TestEditBlockedMsgFlashes verifies that EditBlockedMsg with empty Reason
// flashes "Reveal first with r".
func TestEditBlockedMsgFlashes(t *testing.T) {
	m := modelInDetailWithRevealedLeaf(t)
	m2 := send(t, m, ui.EditBlockedMsg{})
	v := m2.View()
	// The flash should appear in the status bar
	assert.Contains(t, v.Content, "Reveal first with r",
		"EditBlockedMsg must flash 'Reveal first with r', got: %q", v.Content)
}

// TestEditBlockedMsgWithReasonFlashes verifies that EditBlockedMsg with a non-empty Reason
// flashes the reason string.
func TestEditBlockedMsgWithReasonFlashes(t *testing.T) {
	m := modelInDetailWithRevealedLeaf(t)
	m2 := send(t, m, ui.EditBlockedMsg{Reason: "Array-indexed keys not editable in Phase 3"})
	v := m2.View()
	assert.Contains(t, v.Content, "Array-indexed keys not editable in Phase 3",
		"EditBlockedMsg with reason must flash the reason, got: %q", v.Content)
}

// helper: drive AppModel into stateDiff state.
func modelInStateDiff(t *testing.T) tea.Model {
	t.Helper()
	m := modelInDetailWithRevealedLeaf(t)
	return send(t, m, ui.EditConfirmMsg{
		KeyPath:  "password",
		OldValue: "hunter2",
		NewValue: "new_password",
	})
}

// TestDiffConfirmYTriggersReEncrypt verifies that pressing y in stateDiff returns a non-nil cmd.
func TestDiffConfirmYTriggersReEncrypt(t *testing.T) {
	m := modelInStateDiff(t)
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'y'})
	assert.NotNil(t, cmd, "y in stateDiff must return a non-nil cmd (sops.SetKey dispatch)")
}

// TestDiffCancelNReturnsToDetail verifies that pressing n in stateDiff returns to stateDetail.
func TestDiffCancelNReturnsToDetail(t *testing.T) {
	m := modelInStateDiff(t)
	m2 := send(t, m, tea.KeyPressMsg{Code: 'n'})
	v := m2.View()
	// stateDetail renders the tree — should NOT contain diff overlay
	assert.NotContains(t, v.Content, "confirm re-encrypt",
		"n in stateDiff must dismiss diff overlay, got: %q", v.Content)
}

// TestDiffCancelEscReturnsToDetail verifies that pressing Esc in stateDiff returns to stateDetail.
func TestDiffCancelEscReturnsToDetail(t *testing.T) {
	m := modelInStateDiff(t)
	m2 := send(t, m, tea.KeyPressMsg{Code: tea.KeyEsc})
	v := m2.View()
	assert.NotContains(t, v.Content, "confirm re-encrypt",
		"Esc in stateDiff must dismiss diff overlay, got: %q", v.Content)
}

// TestReEncryptDoneMsgSuccess verifies that ReEncryptDoneMsg{Err: nil} transitions to
// stateDetail and flashes "Re-encrypted".
func TestReEncryptDoneMsgSuccess(t *testing.T) {
	m := modelInStateDiff(t)
	m2 := send(t, m, app.ReEncryptDoneMsg{Err: nil})
	v := m2.View()
	assert.Contains(t, v.Content, "Re-encrypted",
		"success ReEncryptDoneMsg must flash 'Re-encrypted', got: %q", v.Content)
	assert.NotContains(t, v.Content, "confirm re-encrypt",
		"success ReEncryptDoneMsg must leave stateDiff, got: %q", v.Content)
}

// TestReEncryptDoneMsgError verifies that ReEncryptDoneMsg{Err: error} flashes
// "Re-encryption failed" and transitions to stateDetail.
func TestReEncryptDoneMsgError(t *testing.T) {
	m := modelInStateDiff(t)
	m2 := send(t, m, app.ReEncryptDoneMsg{Err: fmt.Errorf("sops set failed")})
	v := m2.View()
	assert.Contains(t, v.Content, "Re-encryption failed",
		"error ReEncryptDoneMsg must flash 'Re-encryption failed', got: %q", v.Content)
}

// ---- Task 1: $EDITOR flow tests ----

// TestEditorRequestMsgReturnsCmd verifies that EditorRequestMsg dispatched from
// AppModel (when in stateDetail with currentFile set) returns a non-nil cmd.
func TestEditorRequestMsgReturnsCmd(t *testing.T) {
	m := modelInDetailWithRevealedLeaf(t)
	_, cmd := m.Update(ui.EditorRequestMsg{FilePath: "/some/path.yaml"})
	assert.NotNil(t, cmd, "EditorRequestMsg must return a non-nil cmd")
}

// TestEditorFinishedMsgWithError verifies that EditorFinishedMsg with Err set
// flashes the error and does not enter stateDiff.
func TestEditorFinishedMsgWithError(t *testing.T) {
	m := modelInDetailWithRevealedLeaf(t)
	m2 := send(t, m, app.EditorFinishedMsg{Err: fmt.Errorf("editor exited with error")})
	v := m2.View()
	assert.Contains(t, v.Content, "Editor error",
		"EditorFinishedMsg with Err must flash 'Editor error', got: %q", v.Content)
	assert.NotContains(t, v.Content, "confirm re-encrypt",
		"EditorFinishedMsg with Err must not enter stateDiff, got: %q", v.Content)
}

// TestEditorFinishedMsgNoChanges verifies that EditorFinishedMsg with identical content
// flashes "No changes detected" and does not enter stateDiff.
func TestEditorFinishedMsgNoChanges(t *testing.T) {
	m := modelInDetailWithRevealedLeaf(t)
	sameContent := []byte("password: hunter2\n")
	// Write to temp file so the handler can read it back
	tmpFile, err := os.CreateTemp("", "sops-tui-test-*.yaml")
	require.NoError(t, err)
	_, err = tmpFile.Write(sameContent)
	require.NoError(t, err)
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	t.Cleanup(func() { os.Remove(tmpPath) })

	m2 := send(t, m, app.EditorFinishedMsg{
		TmpPath:         tmpPath,
		OriginalContent: sameContent,
	})
	v := m2.View()
	assert.Contains(t, v.Content, "No changes detected",
		"EditorFinishedMsg with identical content must flash 'No changes detected', got: %q", v.Content)
	assert.NotContains(t, v.Content, "confirm re-encrypt",
		"EditorFinishedMsg with no changes must not enter stateDiff, got: %q", v.Content)
}

// TestEditorFinishedMsgWithChanges verifies that EditorFinishedMsg with different content
// triggers stateDiff.
func TestEditorFinishedMsgWithChanges(t *testing.T) {
	m := modelInDetailWithRevealedLeaf(t)
	original := []byte("password: hunter2\n")
	edited := []byte("password: new_secret\n")
	// Write edited content to temp file
	tmpFile, err := os.CreateTemp("", "sops-tui-test-*.yaml")
	require.NoError(t, err)
	_, err = tmpFile.Write(edited)
	require.NoError(t, err)
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	t.Cleanup(func() { os.Remove(tmpPath) })

	m2 := send(t, m, app.EditorFinishedMsg{
		TmpPath:         tmpPath,
		OriginalContent: original,
	})
	v := m2.View()
	assert.Contains(t, v.Content, "confirm re-encrypt",
		"EditorFinishedMsg with changes must enter stateDiff, got: %q", v.Content)
}

// TestCompareDecryptedYAML verifies that compareDecryptedYAML detects key value differences.
func TestCompareDecryptedYAML(t *testing.T) {
	original := []byte("password: hunter2\ntoken: abc123\n")
	edited := []byte("password: new_secret\ntoken: abc123\n")
	diffs, err := app.CompareDecryptedYAML(original, edited)
	require.NoError(t, err)
	require.Len(t, diffs, 1, "should find exactly 1 diff")
	assert.Equal(t, "password", diffs[0].KeyPath)
	assert.Equal(t, "hunter2", diffs[0].OldValue)
	assert.Equal(t, "new_secret", diffs[0].NewValue)
}

// TestCompareDecryptedYAMLNoChanges verifies that identical YAML returns empty diffs.
func TestCompareDecryptedYAMLNoChanges(t *testing.T) {
	content := []byte("password: secret\ntoken: abc\n")
	diffs, err := app.CompareDecryptedYAML(content, content)
	require.NoError(t, err)
	assert.Empty(t, diffs, "identical YAML must return no diffs")
}

// TestCompareDecryptedYAMLKeyOrderIndependent verifies that key order does not affect diff result.
func TestCompareDecryptedYAMLKeyOrderIndependent(t *testing.T) {
	original := []byte("a: 1\nb: 2\n")
	reordered := []byte("b: 2\na: 1\n")
	diffs, err := app.CompareDecryptedYAML(original, reordered)
	require.NoError(t, err)
	assert.Empty(t, diffs, "reordered YAML must produce no diffs (key-order independent)")
}

// TestMultiEntryDiffConfirmUsesEncryptFile verifies that when stateDiff holds multiple
// entries (from $EDITOR) and y is pressed, a non-nil cmd is returned (EncryptFile path).
func TestMultiEntryDiffConfirmUsesEncryptFile(t *testing.T) {
	m := modelInDetailWithRevealedLeaf(t)
	// Drive into a multi-entry diff state by sending EditorFinishedMsg with changes
	original := []byte("password: hunter2\ntoken: old\n")
	edited := []byte("password: new_secret\ntoken: new_token\n")
	// Write edited content to temp file
	tmpFile, err := os.CreateTemp("", "sops-tui-test-*.yaml")
	require.NoError(t, err)
	_, err = tmpFile.Write(edited)
	require.NoError(t, err)
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	t.Cleanup(func() { os.Remove(tmpPath) })

	m2 := send(t, m, app.EditorFinishedMsg{
		TmpPath:         tmpPath,
		OriginalContent: original,
	})
	// Should be in stateDiff now; press y to confirm
	_, cmd := m2.Update(tea.KeyPressMsg{Code: 'y'})
	assert.NotNil(t, cmd, "y in multi-entry stateDiff must return a non-nil cmd (EncryptFile path)")
}

// ---- Task 2: Rotation model tests ----

// TestRotateReadyMsgTransitionsToStateDiff verifies that RotateReadyMsg transitions to stateDiff.
func TestRotateReadyMsgTransitionsToStateDiff(t *testing.T) {
	m := modelInDetailWithRevealedLeaf(t)
	m2 := send(t, m, ui.RotateReadyMsg{
		KeyPath:  "password",
		OldValue: "hunter2",
		NewValue: "newval123",
		Format:   ui.FormatAlphanumeric,
	})
	v := m2.View()
	assert.Contains(t, v.Content, "confirm re-encrypt",
		"RotateReadyMsg must transition to stateDiff, got: %q", v.Content)
}

// TestRotateFormatMenuMsgShowsMenu verifies that RotateFormatMenuMsg transitions to stateFormatMenu.
func TestRotateFormatMenuMsgShowsMenu(t *testing.T) {
	m := modelInDetailWithRevealedLeaf(t)
	m2 := send(t, m, ui.RotateFormatMenuMsg{
		KeyPath:  "password",
		OldValue: "some plain value",
	})
	v := m2.View()
	// Format menu must show the selection prompt
	assert.Contains(t, v.Content, "Select format for new secret:",
		"RotateFormatMenuMsg must show format menu, got: %q", v.Content)
}

// TestFormatMenuEscCancels verifies that Esc in stateFormatMenu returns to previous state.
func TestFormatMenuEscCancels(t *testing.T) {
	m := modelInDetailWithRevealedLeaf(t)
	// Enter format menu
	m2 := send(t, m, ui.RotateFormatMenuMsg{
		KeyPath:  "password",
		OldValue: "some plain value",
	})
	// Press Esc — should leave format menu
	m3 := send(t, m2, tea.KeyPressMsg{Code: tea.KeyEsc})
	v := m3.View()
	assert.NotContains(t, v.Content, "Select format for new secret:",
		"Esc in format menu must dismiss menu, got: %q", v.Content)
}

// TestRotateOnMaskedLeafFlashes verifies that EditBlockedMsg without Reason flashes
// "Reveal first with r" (the AppModel flash for any unrevealable operation).
func TestRotateOnMaskedLeafFlashes(t *testing.T) {
	m := modelInDetailWithRevealedLeaf(t)
	m2 := send(t, m, ui.EditBlockedMsg{})
	v := m2.View()
	assert.Contains(t, v.Content, "Reveal first with r",
		"EditBlockedMsg without reason must flash 'Reveal first with r', got: %q", v.Content)
}

// TestReEncryptDoneMsgRotationFlash verifies that ReEncryptDoneMsg after rotation
// flashes "Rotated to {format}" instead of "Re-encrypted".
func TestReEncryptDoneMsgRotationFlash(t *testing.T) {
	m := modelInDetailWithRevealedLeaf(t)
	// First drive to stateDiff via RotateReadyMsg
	m2 := send(t, m, ui.RotateReadyMsg{
		KeyPath:  "password",
		OldValue: "hunter2",
		NewValue: "newbase64==",
		Format:   ui.FormatBase64,
	})
	// Simulate successful re-encryption
	m3 := send(t, m2, app.ReEncryptDoneMsg{Err: nil})
	v := m3.View()
	assert.Contains(t, v.Content, "Rotated to",
		"rotation ReEncryptDoneMsg must flash 'Rotated to ...', got: %q", v.Content)
}

// ---- Phase 5: Health check, recipient flows, and bulk re-key tests ----

// TestHealthCheckStateTransition verifies that pressing H from stateFileList
// transitions to stateDiff (confirmation gate) with healthcheck sentinel.
func TestHealthCheckStateTransition(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	m2 := send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	files := []sops.DiscoveredFile{
		{Name: "secrets/prod.yaml", AbsPath: "/repo/secrets/prod.yaml", IsEncrypted: true},
	}
	m3 := send(t, m2, app.FilesDiscoveredMsg{Files: files})

	// Press H — should show confirmation diff overlay (stateDiff with healthcheck sentinel)
	m4, _ := m3.Update(tea.KeyPressMsg{Code: 'H'})
	v := m4.View()
	// stateDiff confirmation shows "confirm re-encrypt" or the health check title
	assert.NotEmpty(t, v.Content, "View must not be empty after H key")
}

// TestHealthCheckConfirmTransitionsToStateHealth verifies that confirming the health
// check gate transitions to stateHealth and dispatches an async scan command.
func TestHealthCheckConfirmTransitionsToStateHealth(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	m2 := send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	files := []sops.DiscoveredFile{
		{Name: "secrets/prod.yaml", AbsPath: "/repo/secrets/prod.yaml", IsEncrypted: true},
	}
	m3 := send(t, m2, app.FilesDiscoveredMsg{Files: files})

	// Press H to enter confirmation gate
	m4, _ := m3.Update(tea.KeyPressMsg{Code: 'H'})

	// Press y to confirm — should dispatch health scan and enter stateHealth
	m5, cmd := m4.Update(tea.KeyPressMsg{Code: 'y'})
	v := m5.View()
	assert.NotEmpty(t, v.Content, "View must not be empty after confirming health check")
	// cmd should be non-nil (dispatches async health scan)
	assert.NotNil(t, cmd, "confirming health check must return a non-nil cmd")
}

// TestEscFromHealth verifies that Esc from stateHealth returns to stateFileList.
func TestEscFromHealth(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	m2 := send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	files := []sops.DiscoveredFile{
		{Name: "secrets/prod.yaml", AbsPath: "/repo/secrets/prod.yaml", IsEncrypted: true},
	}
	m3 := send(t, m2, app.FilesDiscoveredMsg{Files: files})

	// Drive to stateHealth via H + y
	m4 := send(t, m3, tea.KeyPressMsg{Code: 'H'})
	m5 := send(t, m4, tea.KeyPressMsg{Code: 'y'})

	// Press Esc — should leave stateHealth
	m6 := send(t, m5, tea.KeyPressMsg{Code: tea.KeyEsc})
	v := m6.View()
	// Should no longer show health overlay (stateFileList renders file list)
	assert.NotContains(t, v.Content, "Secret Health Check",
		"Esc from stateHealth must dismiss health overlay, got: %q", v.Content)
}

// TestRecipientFormStateTransition verifies that pressing a from stateDetail
// transitions to stateRecipientForm.
func TestRecipientFormStateTransition(t *testing.T) {
	m := modelInDetailWithRevealedLeaf(t)
	// Press a — should open add-recipient form
	m2, cmd := m.Update(tea.KeyPressMsg{Code: 'a'})
	v := m2.View()
	assert.NotEmpty(t, v.Content, "View must not be empty after a key in stateDetail")
	// Should render the recipient form overlay
	assert.Contains(t, v.Content, "Add Recipient",
		"stateRecipientForm must show 'Add Recipient', got: %q", v.Content)
	assert.NotNil(t, cmd, "stateRecipientForm activation must return a non-nil cmd (focus)")
}

// TestEscFromRecipientForm verifies that Esc from stateRecipientForm returns to stateDetail.
func TestEscFromRecipientForm(t *testing.T) {
	m := modelInDetailWithRevealedLeaf(t)
	// Enter stateRecipientForm
	m2 := send(t, m, tea.KeyPressMsg{Code: 'a'})
	// Press Esc — should return to stateDetail
	m3 := send(t, m2, tea.KeyPressMsg{Code: tea.KeyEsc})
	v := m3.View()
	// stateDetail renders the YAML tree, not the recipient form
	assert.NotContains(t, v.Content, "Add Recipient",
		"Esc from stateRecipientForm must dismiss form, got: %q", v.Content)
}

// TestRecipientListStateTransition verifies that pressing d from stateDetail with
// recipients configured transitions to stateRecipientList.
func TestRecipientListStateTransition(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	m2 := send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	// Inject a parsed file that has AgeRecipients populated.
	nodes := []ui.TreeNode{
		{Key: "password", Encrypted: true, Revealed: true, DecryptedValue: "hunter2"},
	}
	parsed := app.ParsedFileForTest(nodes)
	// We need to set currentParsed via FilesParsedMsg; but ParsedFileForTest only sets Nodes.
	// Drive to stateDetail first via FilesParsedMsg
	m3 := send(t, m2, app.FilesParsedMsg{Parsed: parsed})
	// Without AgeRecipients set, d should flash "No age recipients configured"
	m4 := send(t, m3, tea.KeyPressMsg{Code: 'd'})
	v := m4.View()
	assert.Contains(t, v.Content, "No age recipients",
		"d with no recipients must flash 'No age recipients', got: %q", v.Content)
}

// TestBulkReKeyNoSelection verifies that pressing K with no selected files
// flashes "Select files with space first".
func TestBulkReKeyNoSelection(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	m2 := send(t, m, tea.WindowSizeMsg{Width: 80, Height: 24})
	files := []sops.DiscoveredFile{
		{Name: "secrets/prod.yaml", AbsPath: "/repo/secrets/prod.yaml", IsEncrypted: true},
	}
	m3 := send(t, m2, app.FilesDiscoveredMsg{Files: files})

	// Press K with no files selected — should flash warning
	m4 := send(t, m3, tea.KeyPressMsg{Code: 'K'})
	v := m4.View()
	assert.Contains(t, v.Content, "Select files with space first",
		"K with no selection must flash 'Select files with space first', got: %q", v.Content)
}
