// Package ui provides the git history overlay component for sops-tui.
//
// HistoryModel renders a full-screen overlay showing git commit history for a file.
// It mirrors MetadataModel's pattern: bordered box, surface background, j/k scroll.
//
// Per D-13: full-screen overlay triggered by b key, same prevState pattern.
// Per D-14: accessible from detail view only.
// Per D-15: each entry shows short hash, relative date, author, commit subject.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"strings"

	gitpkg "github.com/caesarakalaeii/sops-tui/internal/git"
	"github.com/caesarakalaeii/sops-tui/internal/keys"
)

// HistoryModel renders a full-screen overlay showing git commit history for a file.
// It mirrors MetadataModel's pattern: bordered box, surface background, j/k scroll.
type HistoryModel struct {
	filename string
	entries  []gitpkg.CommitEntry
	loading  bool
	width    int
	height   int
	scroll   int
}

// NewHistoryModel creates a HistoryModel sized to the given dimensions.
// The model starts in loading=true state; call SetEntries() to transition to content.
func NewHistoryModel(filename string, width, height int) HistoryModel {
	return HistoryModel{
		filename: filename,
		loading:  true,
		width:    width,
		height:   height,
		scroll:   0,
	}
}

// SetEntries populates the model with commit entries and clears the loading state.
func (m *HistoryModel) SetEntries(entries []gitpkg.CommitEntry) {
	m.entries = entries
	m.loading = false
}

// SetSize updates the component dimensions.
func (m *HistoryModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// ScrollDown scrolls the history content down by one line.
// Clamped to prevent scrolling past the last entry.
func (m *HistoryModel) ScrollDown() {
	maxScroll := len(m.entries) - 1
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scroll < maxScroll {
		m.scroll++
	}
}

// ScrollUp scrolls the history content up by one line.
// Clamped to 0 (top of content).
func (m *HistoryModel) ScrollUp() {
	if m.scroll > 0 {
		m.scroll--
	}
}

// View renders the full-screen git history overlay.
// Returns a bordered box with commit entries and scroll support.
func (m HistoryModel) View() string {
	// Title: "git log -- filename" per UI-SPEC Copywriting Contract
	title := HelpSectionHeader.Render("git log -- " + m.filename)

	var inner string

	// Footer per UI-SPEC Copywriting Contract — always shown unless loading
	footer := OverlayMutedFooterStyle.Render("j/k scroll  b or esc close")

	if m.loading {
		// Loading state: title + loading message (no footer)
		loadingText := DimText.Render("Loading history...")
		inner = title + "\n\n" + loadingText
	} else if len(m.entries) == 0 {
		// Empty state: title + no-history message + footer
		emptyHeading := HelpSectionHeader.Render("No commits found")
		emptyBody := DimText.Render("This file has no git history in the current repository.")
		inner = title + "\n\n" + emptyHeading + "\n" + emptyBody + "\n\n" + footer
	} else {
		// Build entry lines with fixed-width columns
		hashStyle := HistoryHashStyle.Width(9)      // 7 chars + 2 padding
		dateStyle := HistoryDateStyle.Width(16)     // "12 months ago" + padding
		authorStyle := HistoryAuthorStyle.Width(18) // author name + padding

		var lines []string
		for _, e := range m.entries {
			line := hashStyle.Render(e.ShortHash) +
				dateStyle.Render(e.RelDate) +
				authorStyle.Render(e.Author) +
				HistorySubjectStyle.Render(e.Subject)
			lines = append(lines, line)
		}

		// Apply scroll offset
		visibleLines := lines
		if m.scroll > 0 && m.scroll < len(lines) {
			visibleLines = lines[m.scroll:]
		}

		content := strings.Join(visibleLines, "\n")
		inner = title + "\n\n" + content + "\n\n" + footer
	}

	// Phase 7.1 D-112: View() returns inner content only; the outer
	// WrapTitled at AppModel.View() (model.go:1342) is the single border
	// source. Width/height are still tracked via SetSize for scroll math.
	return inner
}

// CommitCount returns the number of commit entries currently loaded.
// Consumed by AppModel.titleForState() to render "History (N)" per D-15.
func (m HistoryModel) CommitCount() int {
	return len(m.entries)
}

// Hints returns the 5-hint persistent menu set for HistoryModel per D-09.
func (m HistoryModel) Hints() []keys.MenuHint {
	return []keys.MenuHint{
		{Mnemonic: "j", Description: "scroll down", Visible: true},
		{Mnemonic: "k", Description: "scroll up", Visible: true},
		{Mnemonic: "b", Description: "close history", Visible: true},
		{Mnemonic: "Esc", Description: "close history", Visible: true},
		{Mnemonic: "q", Description: "quit", Visible: true},
	}
}
