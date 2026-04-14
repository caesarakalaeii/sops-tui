// Package ui provides the inline fuzzy search filter bar for sops-tui.
//
// SearchModel is a single-row inline filter input activated by the "/" key.
// It wraps bubbles/textinput and integrates with sahilm/fuzzy for matching.
//
// Per D-10: pressing "/" activates inline filter at bottom of current view.
// Per D-11: search is context-aware (file names vs key paths).
// Per D-12: fuzzy match highlighting uses ColorAccent on matched characters.
// Per T-02-07: textinput.CharLimit set to 100 to mitigate DoS via long input.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/sahilm/fuzzy"
)

// SearchModel is an inline filter bar that renders as a single terminal row.
// It wraps bubbles/textinput and integrates with sahilm/fuzzy for matching.
// It is NOT a sessionState -- it is a mode flag on the parent model.
type SearchModel struct {
	input  textinput.Model
	active bool
	width  int
}

// NewSearchModel creates a search bar sized to the given width.
func NewSearchModel(width int) SearchModel {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.CharLimit = 100 // T-02-07: mitigate DoS via long input
	ti.Prompt = ""     // we render our own "/" prompt
	return SearchModel{
		input: ti,
		width: width,
	}
}

// SetActive activates or deactivates the search bar.
// When activated, the textinput receives focus (Pitfall 6: must call Focus()).
// When deactivated, the textinput is blurred and its value is reset.
func (m *SearchModel) SetActive(active bool) tea.Cmd {
	if active {
		m.active = true
		cmd := m.input.Focus()
		return cmd
	}
	m.active = false
	m.input.Blur()
	m.input.SetValue("")
	return nil
}

// IsActive returns whether the search bar is currently active.
func (m SearchModel) IsActive() bool {
	return m.active
}

// Value returns the current filter text.
func (m SearchModel) Value() string {
	return m.input.Value()
}

// Reset clears the filter text and deactivates the search bar.
func (m *SearchModel) Reset() {
	m.active = false
	m.input.Blur()
	m.input.SetValue("")
}

// SetWidth updates the search bar width.
func (m *SearchModel) SetWidth(width int) {
	m.width = width
}

// Update processes key events for the textinput.
// Only call this when IsActive() is true.
func (m SearchModel) Update(msg tea.Msg) (SearchModel, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View renders the search bar as a single row: "/" prompt in accent + filter text on surface background.
func (m SearchModel) View() string {
	// "/" prompt in accent color per UI-SPEC Search Bar Rendering Contract
	prompt := lipgloss.NewStyle().Foreground(ColorAccent).Render("/")
	space := " "

	// Calculate input area width: full width minus "/" prompt (1) and space (1)
	inputWidth := m.width - 2
	if inputWidth < 1 {
		inputWidth = 1
	}

	inputArea := SearchInputStyle.Width(inputWidth).Render(m.input.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, prompt, space, inputArea)
}

// HighlightMatch renders a string with matched character positions highlighted
// using SearchMatchStyle. Non-matched characters use defaultStyle.
// matchedIdxs comes from fuzzy.Match.MatchedIndexes.
func HighlightMatch(s string, matchedIdxs []int, defaultStyle lipgloss.Style) string {
	if len(s) == 0 {
		return ""
	}

	// Build a set from matchedIdxs for O(1) lookup
	idxSet := make(map[int]bool, len(matchedIdxs))
	for _, idx := range matchedIdxs {
		idxSet[idx] = true
	}

	// Walk runes of s: highlight matched characters, use defaultStyle for rest
	var sb strings.Builder
	for i, r := range s {
		ch := string(r)
		if idxSet[i] {
			sb.WriteString(SearchMatchStyle.Render(ch))
		} else {
			sb.WriteString(defaultStyle.Render(ch))
		}
	}
	return sb.String()
}

// ApplyFilter runs sahilm/fuzzy.Find against the given pattern and source strings.
// Returns fuzzy.Matches sorted by score (best first).
// If pattern is empty, returns nil (caller should show all items).
func ApplyFilter(pattern string, source []string) fuzzy.Matches {
	if pattern == "" {
		return nil
	}
	return fuzzy.Find(pattern, source)
}
