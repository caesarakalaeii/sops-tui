// Package ui - persistent keybinding menu primitive (Phase 7).
//
// RenderMenu renders the persistent keybinding menu as a 2-col x 6-row
// grid via manual lipgloss.JoinHorizontal of two pre-rendered
// fixed-width columns (Phase 7.1 D-117 - replaces the Phase 7
// lipgloss/v2/table builder so cell wrapping never engages). Fixed
// layout at all widths per D-04 (narrow-terminal aesthetics deferred
// to Phase 10). Column-major fill per D-07: hints 0..5 land in
// column 0, hints 6..11 land in column 1. Only hints with Visible=true
// count toward the 12 slots per D-06; surplus remains discoverable in
// the ? full-screen overlay (UI-11).
//
// Each cell is composed as:
//
//	MenuKeyStyle.Render("[" + mnemonic + "]") + " " + MenuDescStyle.Render(ansi.Truncate(desc, descWidth, ""))
//
// so the mnemonic gets the accent foreground and the description gets
// the default foreground per D-05. ansi.Truncate clips the description
// to fit the column budget - guaranteeing single-line cells per D-117
// "cell wrapping never engages". (Earlier MaxWidth-based attempt caused
// wrapping in lipgloss/v2 because MaxWidth allows soft-wrap of long
// content; explicit pre-render truncation is the only reliable clip.)
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

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

	// Each cell is `[mnem] desc`. The mnemonic-with-brackets segment
	// length depends on the mnemonic itself (e.g. "j" → "[j]" = 3 cells;
	// "enter" → "[enter]" = 7 cells). We compute descWidth per-cell from
	// the actual mnemonic length so descriptions get all remaining
	// horizontal budget without bleeding into the next column.
	//
	// Per-cell formula:
	//
	//	descWidth = colWidth - lipgloss.Width("[" + mnem + "] ")
	//
	// Floor at 1 so ansi.Truncate always has at least 1 cell to emit.

	// D-07: column-major fill. hints[0..menuRows-1] -> col 0;
	// hints[menuRows..2*menuRows-1] -> col 1.
	leftRows := make([]string, menuRows)
	rightRows := make([]string, menuRows)
	for r := 0; r < menuRows; r++ {
		// Left column
		if r < len(visible) {
			leftRows[r] = renderMenuCell(visible[r], colWidth)
		} else {
			leftRows[r] = ""
		}
		// Right column
		idx := r + menuRows
		if idx < len(visible) {
			rightRows[r] = renderMenuCell(visible[idx], colWidth)
		} else {
			rightRows[r] = ""
		}
	}

	leftCol := MenuColumnStyle.Width(colWidth).Render(strings.Join(leftRows, "\n"))
	rightCol := MenuColumnStyle.Width(colWidth).Render(strings.Join(rightRows, "\n"))

	return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)
}

// renderMenuCell composes a single menu cell at the requested column
// width, truncating the description with ansi.Truncate so the cell
// never wraps to a second line (Phase 7.1 D-117 hard invariant).
func renderMenuCell(h keys.MenuHint, colWidth int) string {
	keyLabel := MenuKeyStyle.Render("[" + h.Mnemonic + "]")
	keyVisible := lipgloss.Width(keyLabel) + 1 // +1 for the space separator
	descWidth := colWidth - keyVisible
	if descWidth < 1 {
		descWidth = 1
	}
	desc := ansi.Truncate(h.Description, descWidth, "")
	return keyLabel + " " + MenuDescStyle.Render(desc)
}
