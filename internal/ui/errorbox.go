// Package ui provides the styled stderr error box renderer for startup validation output.
//
// RenderErrorBox writes a lipgloss-styled, bordered box to any io.Writer.
// In production the caller passes os.Stderr; in tests a bytes.Buffer is used.
// This satisfies D-01: styled stderr output before any TUI session is initialized.
package ui

import (
	"fmt"
	"io"
	"os"
	"strings"

	xterm "github.com/charmbracelet/x/term"

	"charm.land/lipgloss/v2"

	"github.com/caesarakalaeii/sops-tui/internal/validator"
)

const (
	// errorBoxMaxWidth is the maximum content width of the error box (per 01-UI-SPEC.md).
	errorBoxMaxWidth = 72
	// errorBoxFallbackWidth is used when terminal width detection fails.
	errorBoxFallbackWidth = 80
)

// termWidth returns the current terminal width detected from stderr, with a
// fallback to errorBoxFallbackWidth when detection is unavailable (e.g. in CI
// or when stderr is not a TTY).
func termWidth() int {
	w, _, err := xterm.GetSize(os.Stderr.Fd())
	if err != nil || w <= 0 {
		return errorBoxFallbackWidth
	}
	return w
}

// RenderErrorBox renders a lipgloss-styled error/warning box to w.
//
// Each ValidationResult is rendered with a label ([ERROR] or [WARN]) followed
// by the message on the same line, and the fix/resolution on the next line
// indented to align with the message text.
//
// The box border is ColorError when hasHardError is true, or ColorWarning for
// a warnings-only pass.  The header line is written above the box so that it
// appears even in narrow terminals.
//
// Per D-02 all results appear in a single box.
func RenderErrorBox(results []validator.ValidationResult, hasHardError bool, w io.Writer) {
	// Determine header text (D-03 hard vs soft distinction).
	header := "sops-tui: warnings"
	if hasHardError {
		header = "sops-tui: startup failed"
	}

	// Determine border color.
	borderColor := lipgloss.Color(ColorWarningHex)
	if hasHardError {
		borderColor = lipgloss.Color(ColorErrorHex)
	}

	// Build content lines — one entry per result.
	var contentLines []string
	for i, r := range results {
		var label string
		if r.Severity == validator.SeverityError {
			label = ErrorLabel.Render("[ERROR]")
		} else {
			// Trailing space aligns "[WARN] " (7 chars) with "[ERROR]" (7 chars).
			label = WarnLabel.Render("[WARN] ")
		}
		// Message line: "<label> <message>"
		contentLines = append(contentLines, fmt.Sprintf("%s %s", label, r.Message))
		// Fix line: 8 spaces (label width 7 + 1 separator) aligns fix under message.
		contentLines = append(contentLines, fmt.Sprintf("        %s", r.Fix))
		// Blank separator between results (but not after the last one).
		if i < len(results)-1 {
			contentLines = append(contentLines, "")
		}
	}
	content := strings.Join(contentLines, "\n")

	// Determine box width: min(terminalWidth-4, errorBoxMaxWidth).
	tw := termWidth()
	boxWidth := tw - 4
	if boxWidth > errorBoxMaxWidth {
		boxWidth = errorBoxMaxWidth
	}
	if boxWidth < 1 {
		boxWidth = errorBoxMaxWidth
	}

	// Render the styled box.
	// Padding(1, 2): 1 row vertical, 2 cells horizontal (sm per 01-UI-SPEC.md §Spacing Scale).
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Width(boxWidth).
		Render(content)

	lipgloss.Fprintln(w, header)
	lipgloss.Fprintln(w, box)
}
