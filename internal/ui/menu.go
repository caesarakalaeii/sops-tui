// Package ui - persistent keybinding menu primitive (Phase 7).
//
// RenderMenu renders the persistent keybinding menu as a 2-col x 6-row
// grid via manual lipgloss.JoinHorizontal of two pre-rendered
// fixed-width columns (Phase 7.1 D-117 — replaces the Phase 7
// lipgloss/v2/table builder so cell wrapping never engages). Fixed
// layout at all widths per D-04 (narrow-terminal aesthetics deferred
// to Phase 10). Column-major fill per D-07: hints 0..5 land in
// column 0, hints 6..11 land in column 1. Only hints with Visible=true
// count toward the 12 slots per D-06; surplus remains discoverable in
// the ? full-screen overlay (UI-11).
//
// Each cell is composed as:
//
//	MenuKeyStyle.Render("[" + mnemonic + "]") + " " + MenuDescStyle.MaxWidth(descWidth).Render(desc)
//
// so the mnemonic gets the accent foreground and the description gets
// the default foreground per D-05. MaxWidth on the description ensures
// the cell clips rather than wraps when the column width is tight.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
)

// menuRows is the fixed row count for the persistent menu (D-04).
const menuRows = 6

// menuCols is the fixed column count for the persistent menu (D-04).
const menuCols = 2

// menuSlots is the total capacity of the persistent menu - surplus
// hints remain discoverable in the ? full-screen help overlay (UI-11).
const menuSlots = menuRows * menuCols // 12

// RenderMenu renders the persistent keybinding menu as a 2-col x 6-row
// grid at the requested outer width. Hints with Visible=false are
// skipped per D-06; surplus (>12) is silently dropped per D-118.
//
// Phase 7.1 D-117: replaces the Phase 7 lipgloss/v2/table builder with
// a manual JoinHorizontal of two pre-rendered fixed-width columns.
// Cell wrapping never engages because each column is sized explicitly
// and each cell uses MenuDescStyle.MaxWidth(...) to clip rather than
// wrap. Side benefit: removes the ~394 us lipgloss/v2/table contribution
// to BenchmarkAppView (CONTEXT D-117 / WR-03 path forward for UI-21
// post-cache).
func RenderMenu(hints []keys.MenuHint, width int) string {
	// D-06 + D-118: filter to visible hints only, cap at menuSlots (12).
	visible := make([]keys.MenuHint, 0, menuSlots)
	for _, h := range hints {
		if h.Visible {
			visible = append(visible, h)
			if len(visible) == menuSlots {
				break
			}
		}
	}

	// Compute per-column width. Floor at minMenuCol (8) so cells stay
	// legible; the chrome composer's mid-tier fallback already gates
	// RenderMenu being called below this threshold.
	colWidth := width / menuCols
	if colWidth < minMenuCol {
		colWidth = minMenuCol
	}

	// Mnemonic + brackets + space budget: "[X] " = 4 cells minimum;
	// give the description the remaining space inside colWidth via
	// MaxWidth (clip, do not wrap).
	const keyBudget = 4 // "[" + mnemonic-1 + "]" + space
	descWidth := colWidth - keyBudget
	if descWidth < 1 {
		descWidth = 1
	}

	// D-07: column-major fill. hints[0..menuRows-1] -> col 0;
	// hints[menuRows..2*menuRows-1] -> col 1.
	leftRows := make([]string, menuRows)
	rightRows := make([]string, menuRows)
	for r := 0; r < menuRows; r++ {
		// Left column
		if r < len(visible) {
			leftRows[r] = MenuKeyStyle.Render("["+visible[r].Mnemonic+"]") + " " +
				MenuDescStyle.MaxWidth(descWidth).Render(visible[r].Description)
		} else {
			leftRows[r] = ""
		}
		// Right column
		idx := r + menuRows
		if idx < len(visible) {
			rightRows[r] = MenuKeyStyle.Render("["+visible[idx].Mnemonic+"]") + " " +
				MenuDescStyle.MaxWidth(descWidth).Render(visible[idx].Description)
		} else {
			rightRows[r] = ""
		}
	}

	leftCol := MenuColumnStyle.Width(colWidth).Render(strings.Join(leftRows, "\n"))
	rightCol := MenuColumnStyle.Width(colWidth).Render(strings.Join(rightRows, "\n"))

	return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)
}
