// Package ui provides the health check overlay component for sops-tui.
//
// HealthModel renders a full-screen overlay showing secret health check results
// grouped by category: weak secrets, duplicates, and stale files.
//
// It mirrors HistoryModel's pattern: bordered box, loading state, j/k scroll.
// Scrolling is driven by the parent AppModel (no Update() method) per D-11.
//
// Per HLT-03: health check overlay shows grouped findings with severity indicators.
// Per 05-UI-SPEC.md: loading state, empty state, three section types, errors footer.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/caesarakalaeii/sops-tui/internal/health"
)

// HealthModel renders a full-screen overlay showing secret health check results.
// It mirrors HistoryModel's pattern: bordered box, surface background, j/k scroll.
// The parent AppModel drives scrolling via ScrollDown()/ScrollUp() (no Update()).
type HealthModel struct {
	results health.HealthCheckResult
	loading bool
	width   int
	height  int
	scroll  int
}

// NewHealthModel creates a HealthModel sized to the given dimensions.
// The model starts in loading=true state; call SetResults() to transition to content.
func NewHealthModel(width, height int) HealthModel {
	return HealthModel{
		loading: true,
		width:   width,
		height:  height,
		scroll:  0,
	}
}

// SetResults populates the model with health check results and clears the loading state.
// Resets scroll to 0 to prevent out-of-bounds slice access if the user scrolled during loading.
func (m *HealthModel) SetResults(results health.HealthCheckResult) {
	m.results = results
	m.loading = false
	m.scroll = 0
}

// SetSize updates the component dimensions.
func (m *HealthModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// ScrollDown scrolls the health content down by one line.
// Clamped to prevent scrolling past the last content line.
func (m *HealthModel) ScrollDown() {
	lines := m.buildContentLines()
	maxScroll := len(lines) - 1
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scroll < maxScroll {
		m.scroll++
	}
}

// ScrollUp scrolls the health content up by one line.
// Clamped to 0 (top of content).
func (m *HealthModel) ScrollUp() {
	if m.scroll > 0 {
		m.scroll--
	}
}

// buildContentLines constructs all scrollable content lines for the health overlay.
// Returns each line as a pre-rendered string.
func (m HealthModel) buildContentLines() []string {
	var lines []string

	// Weak secrets section
	if len(m.results.WeakSecrets) > 0 {
		lines = append(lines, HelpSectionHeader.Render("Weak Secrets"))
		for _, ws := range m.results.WeakSecrets {
			line := HealthWeakStyle.Render("[WEAK]") +
				"  " + ws.FilePath + " > " + ws.KeyPath +
				"  (" + ws.Reason + ")"
			lines = append(lines, line)
		}
		lines = append(lines, "")
	}

	// Duplicates section
	if len(m.results.Duplicates) > 0 {
		lines = append(lines, HelpSectionHeader.Render("Duplicates"))
		for _, dup := range m.results.Duplicates {
			var locationParts []string
			for _, loc := range dup.Locations {
				locationParts = append(locationParts, loc.FilePath+" > "+loc.KeyPath)
			}
			line := HealthDupeStyle.Render("[DUPE]") +
				"  " + strings.Join(locationParts, "  AND  ")
			lines = append(lines, line)
		}
		lines = append(lines, "")
	}

	// Stale files section
	if len(m.results.StaleFiles) > 0 {
		lines = append(lines, HelpSectionHeader.Render("Stale Files"))
		for _, sf := range m.results.StaleFiles {
			line := HealthStaleStyle.Render("[STALE]") +
				"  " + sf.FilePath +
				"  (last changed " + fmt.Sprintf("%d", sf.DaysSince) + " days ago)"
			lines = append(lines, line)
		}
		lines = append(lines, "")
	}

	// Errors footer line
	if len(m.results.Errors) > 0 {
		errLine := lipgloss.NewStyle().Foreground(ColorMuted).Render(
			fmt.Sprintf("%d file(s) skipped -- could not decrypt", len(m.results.Errors)),
		)
		lines = append(lines, errLine)
	}

	return lines
}

// View renders the full-screen health check overlay.
// Returns a bordered box with health check results and scroll support.
func (m HealthModel) View() string {
	// Title per UI-SPEC Copywriting Contract
	title := HelpSectionHeader.Render("Secret Health Check")

	// Footer per UI-SPEC Copywriting Contract
	footer := DimText.Render("j/k scroll  H or Esc close")

	var inner string

	if m.loading {
		// Loading state: title + loading message (no footer)
		loadingText := DimText.Render("Running health check...")
		inner = title + "\n\n" + loadingText
	} else if m.results.IsEmpty() {
		// Empty state: no issues, no errors
		emptyHeading := HelpSectionHeader.Render("No issues found")
		emptyBody := DimText.Render("All secrets passed health checks.")
		inner = title + "\n\n" + emptyHeading + "\n" + emptyBody + "\n\n" + footer
	} else {
		// Build and scroll content lines
		allLines := m.buildContentLines()
		visibleLines := allLines
		if m.scroll > 0 && m.scroll < len(allLines) {
			visibleLines = allLines[m.scroll:]
		}
		content := strings.Join(visibleLines, "\n")
		inner = title + "\n\n" + content + "\n\n" + footer
	}

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
