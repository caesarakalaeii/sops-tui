// Package ui provides the status bar component for sops-tui.
//
// StatusBarModel renders a three-section bottom bar:
//   - Left: breadcrumb navigation path (e.g. "sops-tui > files > prod.yaml")
//   - Center: item count (e.g. "12 items" or "3 keys")
//   - Right: environment indicators (sops, age, .sops.yaml availability)
//
// Flash messages temporarily replace all three sections for 2 seconds.
// A generation counter prevents stale ticks from clearing a newer flash
// (per RESEARCH.md Pitfall 6).
//
// Per D-10: status bar is always visible at the bottom of every screen.
// Per D-11: three-section layout with env indicators.
// Per D-12: flash messages for user feedback.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// FlashClearMsg is sent by tea.Tick after 2 seconds to clear the flash message.
// The Gen field is compared against the current flashGen to prevent stale clears
// (per RESEARCH.md Pitfall 6 — generation counter).
type FlashClearMsg struct {
	Gen int
}

// EnvStatus holds the availability of the three environment components
// shown as indicators on the right side of the status bar.
type EnvStatus struct {
	// SopsAvailable is true when the sops binary is found on PATH.
	SopsAvailable bool
	// AgeAvailable is true when ~/.config/sops/age/keys.txt exists.
	AgeAvailable bool
	// SopsYamlAvailable is true when a .sops.yaml is found in the current tree.
	SopsYamlAvailable bool
	// GitAvailable is true when current directory is inside a git repo.
	GitAvailable bool
}

// StatusBarModel holds the state for the persistent bottom status bar.
// It is a value type — methods that modify state return a new StatusBarModel.
//
// Phase 8 D-209/D-211: count and label fields removed; titled-border
// title is the canonical count display (Phase 7 D-15). SetItemCount is kept
// as a no-op for backward compat with 14 model.go call-sites.
type StatusBarModel struct {
	breadcrumb   string
	env          EnvStatus
	flash        string
	flashGen     int
	clipboardHot bool // true when clipboard holds a secret
}

// NewStatusBarModel creates a StatusBarModel with sensible defaults.
// The initial breadcrumb is "sops-tui" and no flash is active.
func NewStatusBarModel(env EnvStatus) StatusBarModel {
	return StatusBarModel{
		breadcrumb: "sops-tui",
		env:        env,
	}
}

// SetBreadcrumb updates the breadcrumb path.
// Segments are joined with " > " separators.
// The fixed prefix "sops-tui" is always prepended.
func (m *StatusBarModel) SetBreadcrumb(segments ...string) {
	parts := make([]string, 0, len(segments)+1)
	parts = append(parts, "sops-tui")
	parts = append(parts, segments...)
	m.breadcrumb = strings.Join(parts, " > ")
}

// Segments returns the underlying breadcrumb segments split on " > ".
// Phase 8 D-210: read-path counterpart to SetBreadcrumb; consumed by
// AppModel.View() via ui.RenderCrumbs(m.status.Segments(), m.width)
// to render the chip-pill row above the titled body.
//
// SetBreadcrumb joins parts with " > "; Segments reverses the join.
// Returns nil for an empty breadcrumb so the caller can distinguish
// "no segments" from "single empty segment".
func (m StatusBarModel) Segments() []string {
	if m.breadcrumb == "" {
		return nil
	}
	return strings.Split(m.breadcrumb, " > ")
}

// SetItemCount is retained as a no-op for backward compatibility with
// 14 existing call-sites in internal/app/model.go. Phase 8 D-209
// moves the canonical count display to the titled-border title (Phase 7
// D-15: "Files (12)" / "History (47)" / etc.); the status-bar item
// count center section is deleted (D-211). Plan 3 may delete the
// method body + the 14 call-sites in its own commit; Plan 2 does
// not block that.
func (m *StatusBarModel) SetItemCount(count int, label string) {
	_, _ = count, label
}

// Env returns the current environment status indicators.
func (m StatusBarModel) Env() EnvStatus { return m.env }

// SetEnv replaces the environment status indicators used by the right-hand
// section of the status bar. Called when git availability is determined
// asynchronously after startup.
func (m *StatusBarModel) SetEnv(env EnvStatus) { m.env = env }

// SetClipboardHot sets whether the clipboard currently holds a secret.
// When true, a [clip] indicator is rendered in the status bar right section.
func (m *StatusBarModel) SetClipboardHot(hot bool) {
	m.clipboardHot = hot
}

// IsClipboardHot returns true when the clipboard currently holds a secret.
func (m StatusBarModel) IsClipboardHot() bool {
	return m.clipboardHot
}

// Flash sets a flash message and schedules a clear after 2 seconds.
// The generation counter is incremented so that only the matching FlashClearMsg
// will clear this flash — earlier ticks are silently ignored.
// Returns the updated model and the tea.Tick command.
func (m StatusBarModel) Flash(msg string) (StatusBarModel, tea.Cmd) {
	m.flashGen++
	gen := m.flashGen
	m.flash = msg
	cmd := tea.Tick(2*time.Second, func(_ time.Time) tea.Msg {
		return FlashClearMsg{Gen: gen}
	})
	return m, cmd
}

// Update handles incoming messages. Only FlashClearMsg is handled; all other
// messages are passed through without state change.
// Returns the updated model and an optional Cmd.
func (m StatusBarModel) Update(msg tea.Msg) (StatusBarModel, tea.Cmd) {
	if clearMsg, ok := msg.(FlashClearMsg); ok {
		// Only clear if the generation matches — stale ticks are ignored (Pitfall 6).
		if clearMsg.Gen == m.flashGen {
			m.flash = ""
		}
		return m, nil
	}
	return m, nil
}

// View renders the status bar at the given width. After Phase 8 D-211
// the normal path renders ONLY the right cluster (env indicators +
// optional clipboard hot indicator) right-aligned on the surface bg.
// The breadcrumb LEFT section moved to ui.RenderCrumbs (chip-pill row
// above the titled body); the item-count CENTER section moved to the
// titled-border title (Phase 7 D-15). Pipe separators (" | ") are
// gone. Flash branch is unchanged per D-212.
func (m StatusBarModel) View(width int) string {
	if m.flash != "" {
		return StatusBarStyle.
			Width(width).
			Align(lipgloss.Center).
			Render(m.flash)
	}

	right := renderEnvIndicators(m.env)
	if m.clipboardHot {
		clipIndicator := ClipboardHotStyle.Render("[clip]")
		// StatusBarStyle is already a package var (no NewStyle() in View()).
		spacer := StatusBarStyle.Render(" ")
		right = lipgloss.JoinHorizontal(lipgloss.Top, clipIndicator, spacer, right)
	}

	return StatusBarStyle.
		Width(width).
		Align(lipgloss.Right).
		Render(right)
}

// renderEnvIndicators builds the right-hand env indicator string.
// Icons: checkmark (✓ \u2713) for available, cross (✗ \u2717) for missing sops,
// warning (⚠ \u26A0) for missing age or .sops.yaml.
func renderEnvIndicators(env EnvStatus) string {
	var indicators []string

	// sops indicator
	if env.SopsAvailable {
		indicators = append(indicators,
			lipgloss.NewStyle().Foreground(ColorSuccess).Render("sops:\u2713"))
	} else {
		indicators = append(indicators,
			lipgloss.NewStyle().Foreground(ColorError).Render("sops:\u2717"))
	}

	// age indicator
	if env.AgeAvailable {
		indicators = append(indicators,
			lipgloss.NewStyle().Foreground(ColorSuccess).Render("age:\u2713"))
	} else {
		indicators = append(indicators,
			lipgloss.NewStyle().Foreground(ColorWarning).Render("age:\u26A0"))
	}

	// .sops.yaml indicator
	if env.SopsYamlAvailable {
		indicators = append(indicators,
			lipgloss.NewStyle().Foreground(ColorSuccess).Render(".sops.yaml:\u2713"))
	} else {
		indicators = append(indicators,
			lipgloss.NewStyle().Foreground(ColorWarning).Render(".sops.yaml:\u26A0"))
	}

	// git availability indicator
	if !env.GitAvailable {
		indicators = append(indicators, GitNoRepoStyle.Render("no git"))
	}

	sep := lipgloss.NewStyle().Foreground(ColorMuted).Render(" ")
	return strings.Join(indicators, sep)
}
