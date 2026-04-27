// Package ui - persistent keybinding menu primitive (Phase 7).
//
// RenderMenu renders the persistent keybinding menu as a 2-col x 6-row
// grid built on charm.land/lipgloss/v2/table. Fixed layout at all widths
// per D-04 (narrow-terminal safe; responsive layouts deferred to first
// bug report). Column-major fill per D-07: hints 0..5 land in column 0,
// hints 6..11 land in column 1. Only hints with Visible=true count
// toward the 12 slots per D-06; surplus remains discoverable in the ?
// full-screen overlay (UI-11).
//
// Each cell is composed as:
//
//	MenuKeyStyle.Render("[" + mnemonic + "]") + " " + MenuDescStyle.Render(desc)
//
// so the mnemonic gets the accent foreground and the description gets
// the default foreground per D-05. The table StyleFunc returns
// MenuCellStyle (currently a no-op; reserved for Phase 10 per-column
// tweaks) so the inline fragment styling wins.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

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
// skipped per D-06; surplus (>12) is discarded. At width < ~4 cells per
// column the table auto-clips; narrow-terminal aesthetics deferred to
// Phase 10 (UI-16).
func RenderMenu(hints []keys.MenuHint, width int) string {
	// D-06: filter to visible hints only, cap at menuSlots.
	visible := make([]keys.MenuHint, 0, menuSlots)
	for _, h := range hints {
		if h.Visible {
			visible = append(visible, h)
			if len(visible) == menuSlots {
				break
			}
		}
	}

	// D-07: column-major fill. hints[0..5] -> col 0 rows 0..5;
	// hints[6..11] -> col 1 rows 0..5. Empty cells render as blank.
	rows := make([][]string, menuRows)
	for r := 0; r < menuRows; r++ {
		rows[r] = make([]string, menuCols)
	}
	for i, h := range visible {
		col := i / menuRows
		row := i % menuRows
		if col >= menuCols {
			break
		}
		// Compose inline: accent mnemonic + fg description.
		rows[row][col] = MenuKeyStyle.Render("["+h.Mnemonic+"]") + " " + MenuDescStyle.Render(h.Description)
	}

	t := table.New().
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderRow(false).
		BorderColumn(false).
		BorderHeader(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			// MenuCellStyle is a no-op package var in Phase 7 (D-05);
			// reserved for Phase 10 per-column tweaks. Inline fragment
			// styling (MenuKeyStyle / MenuDescStyle) provides accent/fg.
			return MenuCellStyle
		}).
		Rows(rows...).
		Width(width)

	return t.Render()
}
