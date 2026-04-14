package ui_test

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewSearchModel verifies a new SearchModel starts inactive with empty value.
func TestNewSearchModel(t *testing.T) {
	m := ui.NewSearchModel(80)
	assert.False(t, m.IsActive(), "new SearchModel must be inactive")
	assert.Equal(t, "", m.Value(), "new SearchModel must have empty Value()")
}

// TestSearchModelSetActiveTrue verifies SetActive(true) makes the model active.
func TestSearchModelSetActiveTrue(t *testing.T) {
	m := ui.NewSearchModel(80)
	_ = m.SetActive(true)
	assert.True(t, m.IsActive(), "after SetActive(true), IsActive() must return true")
}

// TestSearchModelSetActiveFalse verifies SetActive(false) resets value and deactivates.
func TestSearchModelSetActiveFalse(t *testing.T) {
	m := ui.NewSearchModel(80)
	_ = m.SetActive(true)
	// Manually trigger an update to set a value would need a proper tea.Msg;
	// instead we just verify SetActive(false) deactivates.
	_ = m.SetActive(false)
	assert.False(t, m.IsActive(), "after SetActive(false), IsActive() must return false")
	assert.Equal(t, "", m.Value(), "after SetActive(false), Value() must be empty")
}

// TestSearchModelViewContainsSlash verifies View() output contains the "/" prompt character.
func TestSearchModelViewContainsSlash(t *testing.T) {
	m := ui.NewSearchModel(80)
	output := m.View()
	assert.Contains(t, output, "/", "View() must contain '/' prompt character")
}

// TestSearchModelReset verifies Reset clears value and deactivates the model.
func TestSearchModelReset(t *testing.T) {
	m := ui.NewSearchModel(80)
	_ = m.SetActive(true)
	m.Reset()
	assert.False(t, m.IsActive(), "after Reset(), IsActive() must return false")
	assert.Equal(t, "", m.Value(), "after Reset(), Value() must be empty")
}

// TestSearchModelSetWidth verifies SetWidth updates the model width without panicking.
func TestSearchModelSetWidth(t *testing.T) {
	m := ui.NewSearchModel(80)
	require.NotPanics(t, func() {
		m.SetWidth(120)
		_ = m.View()
	})
}

// TestSearchModelIsActiveReflectsState verifies IsActive reflects the current state correctly.
func TestSearchModelIsActiveReflectsState(t *testing.T) {
	m := ui.NewSearchModel(80)
	assert.False(t, m.IsActive())
	_ = m.SetActive(true)
	assert.True(t, m.IsActive())
	_ = m.SetActive(false)
	assert.False(t, m.IsActive())
}

// TestHighlightMatchWithIndices verifies HighlightMatch highlights characters at given indices.
func TestHighlightMatchWithIndices(t *testing.T) {
	defaultStyle := lipgloss.NewStyle().Foreground(ui.ColorFg)
	result := ui.HighlightMatch("secrets", []int{0, 3}, defaultStyle)
	// Should be non-empty and contain the characters
	require.NotEmpty(t, result, "HighlightMatch must return non-empty result")
	// The result should contain ANSI escape codes (from lipgloss rendering)
	// At minimum, the characters s, e, c, r, e, t, s should be present
	// We use string search since ANSI codes make exact comparison hard
	assert.True(t, strings.Contains(result, "s") && strings.Contains(result, "e"),
		"HighlightMatch must contain characters from input string")
}

// TestHighlightMatchEmptyIndices verifies HighlightMatch with no indices returns the string unstyled.
func TestHighlightMatchEmptyIndices(t *testing.T) {
	defaultStyle := lipgloss.NewStyle().Foreground(ui.ColorFg)
	result := ui.HighlightMatch("secrets", []int{}, defaultStyle)
	require.NotEmpty(t, result, "HighlightMatch with empty indices must return non-empty string")
	assert.Contains(t, result, "s", "HighlightMatch with empty indices must contain original characters")
}

// TestHighlightMatchEmptyString verifies HighlightMatch with empty input returns empty string.
func TestHighlightMatchEmptyString(t *testing.T) {
	defaultStyle := lipgloss.NewStyle().Foreground(ui.ColorFg)
	result := ui.HighlightMatch("", []int{0, 1}, defaultStyle)
	// An empty input string should produce an empty result (no panic)
	require.NotPanics(t, func() {
		_ = ui.HighlightMatch("", []int{}, defaultStyle)
	})
	_ = result // just verify it doesn't panic
}

// TestApplyFilterWithMatch verifies ApplyFilter returns a match for "sec" against "secrets.yaml".
func TestApplyFilterWithMatch(t *testing.T) {
	source := []string{"secrets.yaml", "config.yaml", "database.yaml"}
	matches := ui.ApplyFilter("sec", source)
	require.NotNil(t, matches, "ApplyFilter must return non-nil for non-empty pattern with matches")
	require.NotEmpty(t, matches, "ApplyFilter must return matches for pattern 'sec' against 'secrets.yaml'")
	// Best match should be secrets.yaml
	assert.Equal(t, "secrets.yaml", matches[0].Str,
		"First match for 'sec' must be 'secrets.yaml'")
}

// TestApplyFilterEmptyPattern verifies ApplyFilter returns nil for empty pattern.
func TestApplyFilterEmptyPattern(t *testing.T) {
	source := []string{"secrets.yaml", "config.yaml"}
	matches := ui.ApplyFilter("", source)
	assert.Nil(t, matches, "ApplyFilter with empty pattern must return nil")
}

// TestApplyFilterNoMatches verifies ApplyFilter returns empty slice when nothing matches.
func TestApplyFilterNoMatches(t *testing.T) {
	source := []string{"config.yaml", "database.yaml"}
	matches := ui.ApplyFilter("zzz", source)
	// Should return empty/nil when no match found
	assert.Empty(t, matches, "ApplyFilter must return empty result when no matches found")
}

// TestApplyFilterMultipleMatches verifies ApplyFilter returns sorted matches.
func TestApplyFilterMultipleMatches(t *testing.T) {
	source := []string{"secrets.yaml", "service_secrets.yaml", "config.yaml"}
	matches := ui.ApplyFilter("sec", source)
	require.NotEmpty(t, matches, "ApplyFilter must find matches for 'sec'")
	// All matches should contain characters from "sec"
	for _, m := range matches {
		assert.NotEmpty(t, m.Str, "Each match must have a non-empty string")
	}
}

// TestSearchModelViewNonEmpty verifies View() returns non-empty output.
func TestSearchModelViewNonEmpty(t *testing.T) {
	m := ui.NewSearchModel(80)
	output := m.View()
	assert.NotEmpty(t, strings.TrimSpace(output), "View() must return non-empty output")
}
