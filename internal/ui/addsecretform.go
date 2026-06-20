// Package ui provides the add-secret modal overlay for sops-tui.
//
// AddSecretFormModel is a full-screen modal that wraps two textinputs — one for
// the dot-joined key path and one for the plaintext value — used to insert a new
// secret into the currently-open file. On confirm the parent (AppModel) routes the
// key/value through the existing diff-confirm → `sops set` flow, which creates the
// key when it does not yet exist.
//
// Rendering pattern mirrors RecipientFormModel: inner content only (the outer
// WrapTitled border is supplied by AppModel.View()), package-var styles only.
// Input wrapping mirrors RecipientFormModel/SearchModel: textinput.Model with CharLimit.
//
// Validation happens client-side before any sops call:
//   - empty key path is rejected,
//   - array-index notation is rejected (sops set cannot target it),
//   - keys that already exist are rejected (use `e` to edit instead) so a new-secret
//     flow never silently overwrites an existing value.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textinput"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
	"github.com/caesarakalaeii/sops-tui/internal/sops"
)

// AddSecretFormModel is a full-screen modal overlay for entering a new secret.
// The user types a key path and a value, switches fields with Tab, presses
// Enter to validate/confirm, or Esc to cancel.
type AddSecretFormModel struct {
	keys         keys.AddSecretFormKeyMap
	keyInput     textinput.Model
	valueInput   textinput.Model
	focus        int             // 0 = key path input, 1 = value input
	existingKeys map[string]bool // dot-joined paths already present in the file
	errMsg       string          // validation error displayed below the inputs
	width        int
	height       int
	confirmed    bool
	cancelled    bool
}

// NewAddSecretFormModel creates an AddSecretFormModel sized to the given dimensions.
// existingKeys is the set of dot-joined key paths already present in the file; it is
// used to reject duplicates so the add flow never overwrites an existing value.
// The inputs start unfocused; call Activate() before showing the modal.
func NewAddSecretFormModel(width, height int, existingKeys []string) AddSecretFormModel {
	keyInput := textinput.New()
	keyInput.Placeholder = "database.password"
	keyInput.CharLimit = 256
	keyInput.Prompt = ""

	valueInput := textinput.New()
	valueInput.Placeholder = "value"
	valueInput.CharLimit = 4096
	valueInput.Prompt = ""

	set := make(map[string]bool, len(existingKeys))
	for _, k := range existingKeys {
		set[k] = true
	}

	return AddSecretFormModel{
		keys:         keys.DefaultAddSecretFormKeyMap,
		keyInput:     keyInput,
		valueInput:   valueInput,
		existingKeys: set,
		width:        width,
		height:       height,
	}
}

// Activate resets the form state and focuses the key-path input.
// Call this each time the modal is opened. Returns a tea.Cmd to focus the
// input (must be returned from the parent's Update).
func (m *AddSecretFormModel) Activate() tea.Cmd {
	m.confirmed = false
	m.cancelled = false
	m.errMsg = ""
	m.focus = 0
	m.keyInput.SetValue("")
	m.valueInput.SetValue("")
	m.valueInput.Blur()
	return m.keyInput.Focus()
}

// IsActive returns true when the form has not been confirmed or cancelled.
func (m AddSecretFormModel) IsActive() bool {
	return !m.confirmed && !m.cancelled
}

// Confirmed returns true after the user pressed Enter with a valid key/value.
func (m AddSecretFormModel) Confirmed() bool {
	return m.confirmed
}

// Cancelled returns true after the user pressed Esc.
func (m AddSecretFormModel) Cancelled() bool {
	return m.cancelled
}

// KeyPath returns the trimmed dot-joined key path entered by the user.
func (m AddSecretFormModel) KeyPath() string {
	return strings.TrimSpace(m.keyInput.Value())
}

// Value returns the plaintext value entered by the user.
func (m AddSecretFormModel) Value() string {
	return m.valueInput.Value()
}

// SetSize updates the component dimensions.
func (m *AddSecretFormModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Update processes key events for the add-secret modal.
// tab/shift+tab → switch focus between the two inputs.
// enter → validate the key path; confirm if valid, show an error otherwise.
// esc → cancel.
// All other keys → delegate to the focused textinput for character input.
func (m AddSecretFormModel) Update(msg tea.Msg) (AddSecretFormModel, tea.Cmd) {
	if kMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch kMsg.String() {
		case "esc":
			m.cancelled = true
			return m, nil
		case "tab", "shift+tab":
			m.focus = 1 - m.focus
			return m, m.syncFocus()
		case "enter":
			keyPath := strings.TrimSpace(m.keyInput.Value())
			if keyPath == "" {
				m.errMsg = "Key path required"
				return m, nil
			}
			if sops.IsArrayIndexedKeyPath(keyPath) {
				m.errMsg = "Array-index paths are not supported"
				return m, nil
			}
			if m.existingKeys[keyPath] {
				m.errMsg = "Key already exists — use e to edit"
				return m, nil
			}
			m.confirmed = true
			return m, nil
		}
	}
	var cmd tea.Cmd
	if m.focus == 0 {
		m.keyInput, cmd = m.keyInput.Update(msg)
	} else {
		m.valueInput, cmd = m.valueInput.Update(msg)
	}
	return m, cmd
}

// syncFocus blurs the unfocused input and focuses the active one,
// returning the focus tea.Cmd for the newly active textinput.
func (m *AddSecretFormModel) syncFocus() tea.Cmd {
	if m.focus == 0 {
		m.valueInput.Blur()
		return m.keyInput.Focus()
	}
	m.keyInput.Blur()
	return m.valueInput.Focus()
}

// View renders the inner content of the add-secret modal overlay.
// The outer WrapTitled border is supplied by AppModel.View() (single border source).
func (m AddSecretFormModel) View() string {
	title := DiffKeyStyle.Render("Add Secret")

	keyLabel := OverlayMutedFooterStyle.Render("Key path:")
	valLabel := OverlayMutedFooterStyle.Render("Value:   ")

	// inputWidth mirrors RecipientFormModel.View()'s m.width-12 budget: the
	// post-WrapTitled inner area (m.width-4) minus an 8-col label + spacing.
	inputWidth := m.width - 12
	if inputWidth < 1 {
		inputWidth = 1
	}
	keyArea := EditInputStyle.Width(inputWidth).Render(m.keyInput.View())
	valArea := EditInputStyle.Width(inputWidth).Render(m.valueInput.View())

	errLine := ""
	if m.errMsg != "" {
		errLine = "\n" + ValidationErrorStyle.Render(m.errMsg)
	}

	footer := ConfirmPromptStyle.Render("[tab]") + " switch field   " +
		ConfirmPromptStyle.Render("[enter]") + " confirm   " +
		ConfirmPromptStyle.Render("[esc]") + " cancel"

	inner := title + "\n\n" +
		keyLabel + " " + keyArea + "\n" +
		valLabel + " " + valArea + errLine + "\n\n" + footer
	return inner
}

// Hints returns the 3-hint persistent menu set for AddSecretFormModel per D-09.
// Derives from AddSecretFormKeyMap.ShortHelp() per D-301 (total derivation).
func (m AddSecretFormModel) Hints() []keys.MenuHint {
	return keys.HintsFromBindings(m.keys.ShortHelp())
}
