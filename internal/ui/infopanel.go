// Package ui - header info-panel primitive (Phase 8).
//
// RenderInfoPanel returns the 5-row info panel rendered into the 38x6
// envelope reserved by Phase 7's InfoPanelPlaceholderStyle (D-16).
// Phase 8 D-201..D-204 lock the row schema: cfg / age / rcp / git / fil.
//
// Pure function of InfoPanelData -- no I/O, no AppModel coupling.
// View() reads the cached InfoPanelData on AppModel; refresh happens at
// four event seams (D-213) elsewhere.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"strconv"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// InfoPanelData is the value passed to RenderInfoPanel -- pre-computed,
// no I/O at render time (Pitfall 15). All string fields use "" as the
// empty/missing sentinel; integer fields use -1.
type InfoPanelData struct {
	SopsYamlRelPath string // "" -> renders "-"
	AgeFingerprint  string // "age1...xyz" -> middle-truncated to <=10 cells; "" -> "-"
	RecipientCount  int    // -1 -> renders "-"; 0+ renders decimal
	GitBranch       string // "" -> renders "-"
	GitDetached     bool   // true -> "HEAD@<7-char-hash>" (branch holds short hash per D-215)
	GitDirty        bool   // true (and not detached) -> trailing " *"
	FileCount       int    // -1 -> renders "-"; 0+ renders decimal
}

// RenderInfoPanel returns the 5-row info panel string. Caller wraps via
// InfoPanelPlaceholderStyle (Width=38, Height=6) at the chrome composer
// to enforce the envelope; this function returns un-padded content so
// the wrapper can size to its declared 38x6 box.
//
// Row order is LOCKED: cfg, age, rcp, git, fil (D-201). Any executor
// reordering is a contract violation.
func RenderInfoPanel(d InfoPanelData) string {
	rows := []string{
		infoPanelRow("cfg:", sopsYamlDisplay(d.SopsYamlRelPath)),
		infoPanelRow("age:", ageDisplay(d.AgeFingerprint)),
		infoPanelRow("rcp:", rcpDisplay(d.RecipientCount)),
		infoPanelRow("git:", gitDisplay(d.GitBranch, d.GitDetached, d.GitDirty)),
		infoPanelRow("fil:", filDisplay(d.FileCount)),
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// infoPanelRow composes a single row by joining the muted Width(5) label
// (via InfoPanelLabelStyle) with the foreground value (via InfoPanelValueStyle).
func infoPanelRow(label, value string) string {
	return InfoPanelLabelStyle.Render(label) + InfoPanelValueStyle.Render(value)
}

const emptyMarker = "-" // D-204: ASCII hyphen-minus for missing/uncomputed.

// sopsYamlDisplay applies D-203 middle-truncation at 32 cells.
func sopsYamlDisplay(path string) string {
	if path == "" {
		return emptyMarker
	}
	return middleTruncate(path, 32)
}

// ageDisplay applies D-203 + Pitfall 11: <=10 cells with U+2026 ellipsis.
// ALWAYS middle-truncates regardless of input width -- the security gate
// is "<=10 chars with visible ellipsis" (D-220 question 5).
func ageDisplay(fingerprint string) string {
	if fingerprint == "" {
		return emptyMarker
	}
	return middleTruncate(fingerprint, 10)
}

func rcpDisplay(n int) string {
	if n < 0 {
		return emptyMarker
	}
	return strconv.Itoa(n)
}

// gitDisplay formats the branch + dirty/detached marker per D-215 +
// Claude's Discretion recommendation:
//   - detached HEAD: "HEAD@<7-char-hash>" (branch holds the hash)
//   - dirty (and not detached): branch + " *" trailing
//   - clean: branch only (no marker)
//
// Defensive 32-cell middle-truncation for unusually long branch names.
func gitDisplay(branch string, detached, dirty bool) string {
	if branch == "" {
		return emptyMarker
	}
	var v string
	switch {
	case detached:
		v = "HEAD@" + branch // D-215: branch field holds 7-char hash
	case dirty:
		v = branch + " *"
	default:
		v = branch
	}
	return middleTruncate(v, 32)
}

func filDisplay(n int) string {
	if n < 0 {
		return emptyMarker
	}
	return strconv.Itoa(n)
}

// middleTruncate keeps the start and end of s, replacing the middle with
// "…" (U+2026 HORIZONTAL ELLIPSIS), so the result is at most maxCells
// visible cells. Returns s unchanged if it already fits. ANSI- and
// grapheme-aware via charmbracelet/x/ansi v0.11.7.
//
// Algorithm:
//   - if width(s) <= maxCells -> return s
//   - left half  = ansi.Truncate(s, leftBudget, "")  (keep first leftBudget cells)
//   - right half = ansi.TruncateLeft(s, totalCells-rightBudget, "")  (drop first totalCells-rightBudget cells)
//   - return left + ellipsis + right
//
// CRITICAL (08-RESEARCH.md Pitfall B): ansi.Truncate is tail-only; we
// MUST use ansi.TruncateLeft for the right fragment, otherwise we
// produce right-truncated output instead of middle-truncated.
func middleTruncate(s string, maxCells int) string {
	sw := lipgloss.Width(s)
	if sw <= maxCells {
		return s
	}
	const ellipsis = "…" // U+2026 HORIZONTAL ELLIPSIS -- in chrome ASCII allowlist
	ellipsisW := lipgloss.Width(ellipsis)
	if maxCells <= ellipsisW {
		return ellipsis
	}
	available := maxCells - ellipsisW
	left := available / 2
	right := available - left
	leftPart := ansi.Truncate(s, left, "")
	rightPart := ansi.TruncateLeft(s, sw-right, "")
	return leftPart + ellipsis + rightPart
}
