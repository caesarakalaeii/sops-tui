// Package ui provides the add-recipient modal overlay for sops-tui.
//
// RecipientFormModel is a full-screen modal that wraps a single textinput for
// entering an age public key. It validates the key using filippo.io/age before
// confirming (D-02 requirement: client-side validation before any sops call).
//
// Rendering pattern mirrors DiffModel: rounded border, surface background.
// Input wrapping pattern mirrors SearchModel: textinput.Model with CharLimit.
//
// Per RCP-02: add-recipient modal requires valid age public key (bech32 + 32 bytes).
// Per T-05-05: age.ParseX25519Recipient validates encoding before sops subprocess.
// Per T-05-06: CharLimit=200 prevents unbounded input (DoS mitigation).
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"filippo.io/age"
)

// RecipientFormModel is a full-screen modal overlay for entering an age public key.
// The user types a key, presses Enter to validate/confirm, or Esc to cancel.
// Validation uses age.ParseX25519Recipient per D-02 (T-05-05 mitigation).
type RecipientFormModel struct {
	input     textinput.Model
	errMsg    string // validation error displayed below input
	width     int
	height    int
	confirmed bool
	cancelled bool
}

// NewRecipientFormModel creates a RecipientFormModel sized to the given dimensions.
// The textinput starts unfocused; call Activate() before showing the modal.
func NewRecipientFormModel(width, height int) RecipientFormModel {
	ti := textinput.New()
	ti.Placeholder = "age1..."
	ti.CharLimit = 200 // T-05-06: mitigate DoS via long input
	ti.Prompt = ""
	return RecipientFormModel{
		input:  ti,
		width:  width,
		height: height,
	}
}

// Activate resets the form state and focuses the input.
// Call this each time the modal is opened.
// Returns a tea.Cmd to focus the textinput (must be returned from parent's Update).
func (m *RecipientFormModel) Activate() tea.Cmd {
	m.confirmed = false
	m.cancelled = false
	m.errMsg = ""
	m.input.SetValue("")
	return m.input.Focus()
}

// IsActive returns true when the form has not been confirmed or cancelled.
func (m RecipientFormModel) IsActive() bool {
	return !m.confirmed && !m.cancelled
}

// Confirmed returns true after the user pressed Enter with a valid age public key.
func (m RecipientFormModel) Confirmed() bool {
	return m.confirmed
}

// Cancelled returns true after the user pressed Esc.
func (m RecipientFormModel) Cancelled() bool {
	return m.cancelled
}

// Value returns the current text in the input field.
func (m RecipientFormModel) Value() string {
	return m.input.Value()
}

// SetSize updates the component dimensions.
func (m *RecipientFormModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Update processes key events for the recipient form modal.
// enter → validate with age.ParseX25519Recipient; confirm if valid, show error if not.
// esc → cancel.
// All other keys → delegate to textinput for character input.
func (m RecipientFormModel) Update(msg tea.Msg) (RecipientFormModel, tea.Cmd) {
	if kMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch kMsg.String() {
		case "enter":
			_, err := age.ParseX25519Recipient(m.input.Value())
			if err != nil {
				m.errMsg = "Invalid age key: " + err.Error()
				return m, nil
			}
			m.confirmed = true
			return m, nil
		case "esc":
			m.cancelled = true
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View renders the full-screen add-recipient modal overlay.
// Per 05-UI-SPEC.md §Recipient Form Overlay.
func (m RecipientFormModel) View() string {
	// Title per UI-SPEC Copywriting Contract
	title := DiffKeyStyle.Render("Add Recipient")

	// Prompt label
	prompt := lipgloss.NewStyle().Foreground(ColorMuted).Render("Age public key:")

	// Input area width constrained to avoid overflow
	inputWidth := m.width - 12
	if inputWidth < 1 {
		inputWidth = 1
	}
	inputArea := EditInputStyle.Width(inputWidth).Render(m.input.View())

	// Inline validation error (if any)
	errLine := ""
	if m.errMsg != "" {
		errLine = "\n" + ValidationErrorStyle.Render(m.errMsg)
	}

	// Footer with key hints
	footer := ConfirmPromptStyle.Render("[enter]") + " confirm   " +
		ConfirmPromptStyle.Render("[esc]") + " cancel"

	inner := title + "\n\n" + prompt + " " + inputArea + errLine + "\n\n" + footer

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
