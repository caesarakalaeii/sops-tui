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
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textinput"
	"filippo.io/age"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
)

// RecipientFormModel is a full-screen modal overlay for entering an age public key.
// The user types a key, presses Enter to validate/confirm, or Esc to cancel.
// Validation uses age.ParseX25519Recipient per D-02 (T-05-05 mitigation).
type RecipientFormModel struct {
	keys      keys.RecipientFormKeyMap
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
		keys:   keys.DefaultRecipientFormKeyMap,
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
			rawInput := strings.TrimSpace(m.input.Value())
			recipient, err := age.ParseX25519Recipient(rawInput)
			if err != nil {
				m.errMsg = "Invalid age key: " + err.Error()
				return m, nil
			}
			// Re-serialize from the parsed recipient to get the canonical form.
			// This discards any trailing content that the parser may have ignored,
			// preventing argument injection into the sops subprocess (T-05-05).
			canonical := recipient.String()
			if canonical != rawInput {
				m.errMsg = "Invalid age key: unexpected trailing characters"
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

	// Prompt label (Phase 7.1 D-110: lifted to package var; same muted chain)
	prompt := OverlayMutedFooterStyle.Render("Age public key:")

	// Phase 7.1 D-113: inputWidth derives from post-WrapTitled inner area
	// (m.width - 4: NormalBorder=2 + Padding(0,1)=2) minus an 8-col label +
	// spacing budget = m.width - 12 (coincidentally identical to the
	// pre-strip math, which derived -12 = -(2 + 10) for the inner-border
	// envelope; post-strip it derives -12 = -(4 + 8) for the WrapTitled
	// envelope and the label budget).
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

	// Phase 7.1 D-112: View() returns inner content only; the outer
	// WrapTitled at AppModel.View() (model.go:1342) is the single border
	// source. Width/height are still tracked via SetSize for input math.
	inner := title + "\n\n" + prompt + " " + inputArea + errLine + "\n\n" + footer
	return inner
}

// Hints returns the 2-hint persistent menu set for RecipientFormModel per D-09.
// Derives from RecipientFormKeyMap.ShortHelp() per D-301 (total derivation).
func (m RecipientFormModel) Hints() []keys.MenuHint {
	return keys.HintsFromBindings(m.keys.ShortHelp())
}
