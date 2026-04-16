// Package ui_test provides tests for the HistoryModel git history overlay.
package ui_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	gitpkg "github.com/caesarakalaeii/sops-tui/internal/git"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

// TestHistoryModel runs all HistoryModel rendering tests.
func TestHistoryModel(t *testing.T) {
	t.Run("renders entry lines with hash date author subject", func(t *testing.T) {
		entries := []gitpkg.CommitEntry{
			{ShortHash: "abc1234", RelDate: "3 days ago", Author: "Alice", Subject: "Add secrets file"},
			{ShortHash: "def5678", RelDate: "1 week ago", Author: "Bob", Subject: "Update credentials"},
		}
		m := ui.NewHistoryModel("secrets/prod.yaml", 120, 30)
		m.SetEntries(entries)
		view := m.View()

		assert.Contains(t, view, "abc1234", "should contain first short hash")
		assert.Contains(t, view, "3 days ago", "should contain relative date")
		assert.Contains(t, view, "Alice", "should contain author")
		assert.Contains(t, view, "Add secrets file", "should contain commit subject")
		assert.Contains(t, view, "def5678", "should contain second short hash")
		assert.Contains(t, view, "Bob", "should contain second author")
	})

	t.Run("renders empty state when no commits", func(t *testing.T) {
		m := ui.NewHistoryModel("secrets/prod.yaml", 120, 30)
		m.SetEntries([]gitpkg.CommitEntry{})
		view := m.View()

		assert.Contains(t, view, "No commits found", "empty state should show heading")
		assert.Contains(t, view, "This file has no git history in the current repository.", "empty state should show body")
	})

	t.Run("scroll down and up clamp within entry bounds", func(t *testing.T) {
		entries := []gitpkg.CommitEntry{
			{ShortHash: "aaa0001", RelDate: "1 day ago", Author: "Dev", Subject: "First"},
			{ShortHash: "aaa0002", RelDate: "2 days ago", Author: "Dev", Subject: "Second"},
			{ShortHash: "aaa0003", RelDate: "3 days ago", Author: "Dev", Subject: "Third"},
		}
		m := ui.NewHistoryModel("file.yaml", 120, 30)
		m.SetEntries(entries)

		// Scroll down multiple times — should not panic or go past len(entries)-1
		m.ScrollDown()
		m.ScrollDown()
		m.ScrollDown()
		m.ScrollDown() // extra — should clamp

		// View should still render without panic
		view := m.View()
		assert.NotEmpty(t, view)

		// Scroll back up past zero
		m.ScrollUp()
		m.ScrollUp()
		m.ScrollUp()
		m.ScrollUp() // extra — should clamp at 0

		view2 := m.View()
		assert.NotEmpty(t, view2)
	})

	t.Run("renders title with git log filename", func(t *testing.T) {
		m := ui.NewHistoryModel("secrets/prod.yaml", 120, 30)
		m.SetEntries([]gitpkg.CommitEntry{})
		view := m.View()

		assert.Contains(t, view, "git log -- secrets/prod.yaml", "title should include git log filename")
	})

	t.Run("renders footer with scroll and close hints", func(t *testing.T) {
		m := ui.NewHistoryModel("file.yaml", 120, 30)
		m.SetEntries([]gitpkg.CommitEntry{})
		view := m.View()

		assert.Contains(t, view, "j/k scroll  b or esc close", "footer should show scroll and close hints")
	})

	t.Run("renders loading state when loading is true", func(t *testing.T) {
		m := ui.NewHistoryModel("secrets/prod.yaml", 120, 30)
		// loading is true by default before SetEntries is called
		view := m.View()

		assert.Contains(t, view, "Loading history...", "loading state should show loading message")
		// Should NOT show the footer or entries while loading
		assert.False(t, strings.Contains(view, "No commits found"), "loading state should not show empty state")
	})
}
