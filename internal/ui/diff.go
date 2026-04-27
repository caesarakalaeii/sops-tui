// Package ui provides the diff confirmation overlay for sops-tui.
//
// DiffModel is a full-screen overlay that shows old→new value diffs before
// any write operation (re-encryption). It is the universal safety gate for
// all write operations per D-09, D-10, EDT-04.
//
// Layout mirrors MetadataModel: rounded border, surface background, scrollable content.
//
// Per 03-UI-SPEC.md §Overlay Layout: DiffModel, §Interaction Contract: stateDiff.
// Per CONTEXT.md D-09, D-10: y confirms re-encryption; n/Esc cancels.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
)

// DiffEntry holds a single key's old and new value for display in the diff overlay.
type DiffEntry struct {
	KeyPath  string
	OldValue string
	NewValue string
}

// DiffModel is the full-screen diff confirmation overlay.
// It shows old and new values for one or more keys before re-encryption.
// The user confirms with y or cancels with n/Esc.
type DiffModel struct {
	entries   []DiffEntry
	title     string
	width     int
	height    int
	scroll    int
	confirmed bool
	cancelled bool
}

// NewDiffModel creates a DiffModel with the given title, entries, and dimensions.
// title should be "Changes: {key.path}" for a single entry or "Changes: N keys modified"
// for multiple entries.
func NewDiffModel(title string, entries []DiffEntry, width, height int) DiffModel {
	return DiffModel{
		title:   title,
		entries: entries,
		width:   width,
		height:  height,
	}
}

// SetSize updates the component dimensions.
func (m *DiffModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// ScrollDown scrolls the diff content down by one line (clamped to content length).
func (m *DiffModel) ScrollDown() {
	lines := m.buildContentLines()
	maxScroll := len(lines) - 1
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scroll < maxScroll {
		m.scroll++
	}
}

// ScrollUp scrolls the diff content up by one line (clamped to 0).
func (m *DiffModel) ScrollUp() {
	if m.scroll > 0 {
		m.scroll--
	}
}

// Confirmed returns true after the user pressed y to confirm re-encryption.
func (m DiffModel) Confirmed() bool {
	return m.confirmed
}

// Cancelled returns true after the user pressed n or Esc to cancel.
func (m DiffModel) Cancelled() bool {
	return m.cancelled
}

// Entries returns the diff entries held by this model.
func (m DiffModel) Entries() []DiffEntry {
	return m.entries
}

// Update processes key events for the diff overlay.
// y → confirmed; n/Esc → cancelled; j → scroll down; k → scroll up.
func (m DiffModel) Update(msg tea.Msg) (DiffModel, tea.Cmd) {
	if kMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch kMsg.String() {
		case "y":
			m.confirmed = true
			return m, nil
		case "n":
			m.cancelled = true
			return m, nil
		case "esc":
			m.cancelled = true
			return m, nil
		case "j":
			m.ScrollDown()
			return m, nil
		case "k":
			m.ScrollUp()
			return m, nil
		}
	}
	return m, nil
}

// buildContentLines constructs all scrollable diff content lines.
// For single-entry diffs: title header, blank line, removed line, added line.
// For multi-entry diffs: each entry gets its own key path header with removed/added lines.
func (m DiffModel) buildContentLines() []string {
	var lines []string

	if len(m.entries) == 1 {
		entry := m.entries[0]
		lines = append(lines, DiffRemovedStyle.Render("- "+entry.OldValue))
		lines = append(lines, DiffAddedStyle.Render("+ "+entry.NewValue))
	} else {
		for i, entry := range m.entries {
			if i > 0 {
				lines = append(lines, "") // blank separator
			}
			lines = append(lines, DiffKeyStyle.Render(entry.KeyPath))
			lines = append(lines, DiffRemovedStyle.Render("- "+entry.OldValue))
			lines = append(lines, DiffAddedStyle.Render("+ "+entry.NewValue))
		}
	}

	return lines
}

// View renders the full-screen diff overlay.
// Per 03-UI-SPEC.md §Overlay Layout: DiffModel.
func (m DiffModel) View() string {
	// Title: bold key path per DiffKeyStyle
	title := DiffKeyStyle.Render(m.title)

	// Build scrollable content
	allLines := m.buildContentLines()
	visibleLines := allLines
	if m.scroll > 0 && m.scroll < len(allLines) {
		visibleLines = allLines[m.scroll:]
	}
	content := strings.Join(visibleLines, "\n")

	// Footer with y/n confirmation per 03-UI-SPEC.md Copywriting Contract
	footer := ConfirmPromptStyle.Render("[y]") +
		" confirm re-encrypt   " +
		ConfirmPromptStyle.Render("[n/Esc]") +
		" cancel"

	inner := title + "\n\n" + content + "\n\n" + footer

	// Full-screen bordered box per UI-SPEC Overlay Layout Contract
	boxWidth := m.width - 2
	if boxWidth < 1 {
		boxWidth = 1
	}
	boxHeight := m.height - 2
	if boxHeight < 1 {
		boxHeight = 1
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorMuted).
		Background(ColorSurface).
		Padding(1, SpaceMD).
		Width(boxWidth).
		Height(boxHeight).
		Render(inner)
}

// Hints returns the 6-hint persistent menu set for DiffModel per D-09.
// Covers the confirm/cancel/scroll axes. The modal states
// stateRecipientConfirm and stateBulkReKeyConfirm use inline hint sets
// on AppModel (keys.RecipientConfirmHints / BulkReKeyConfirmHints) since
// they share the diff body but change the y/n semantics.
func (m DiffModel) Hints() []keys.MenuHint {
	return []keys.MenuHint{
		{Mnemonic: "y", Description: "confirm re-encrypt", Visible: true},
		{Mnemonic: "n", Description: "cancel", Visible: true},
		{Mnemonic: "Esc", Description: "cancel", Visible: true},
		{Mnemonic: "j", Description: "scroll down", Visible: true},
		{Mnemonic: "k", Description: "scroll up", Visible: true},
		{Mnemonic: "q", Description: "quit", Visible: true},
	}
}
