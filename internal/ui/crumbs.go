// Package ui - breadcrumb chip-pill row primitive (Phase 8).
//
// RenderCrumbs returns the chip-pill row above the titled body: each
// segment rendered as <segment> via CrumbChipStyle (inactive) or
// CrumbChipActiveStyle (last segment). Verbatim k9s parity per D-205,
// D-207, D-208 (project memory: hard product goal).
//
// Sops-tui deviation from k9s: bold weight on active chip is the
// redundant-encoding channel (Pitfall 9) -- k9s relies on bg-only swap
// which fails on 16-color terminals. Plan reviewers must reject any
// drift back to bg-only.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// ellipsisSentinel is the magic string truncateSegmentsToWidth inserts
// when middle segments are dropped due to width overflow (D-216).
// RenderCrumbs recognises this sentinel and styles via CrumbChipEllipsisStyle.
const ellipsisSentinel = "…" // U+2026 HORIZONTAL ELLIPSIS

// RenderCrumbs renders the segments slice as a row of <segment> chip pills.
// width is the terminal width; the row is padded 1 cell on each side
// (D-208) so the chip budget is width - 2.
//
// Active chip (last segment) uses CrumbChipActiveStyle (accent bg +
// body fg + bold). Inactive chips use CrumbChipStyle (surface bg + fg).
// Overflow ellipsis chips use CrumbChipEllipsisStyle (muted fg, no bg).
//
// Segments are normalised per D-207: strings.ToLower then strip spaces.
// This matches k9s crumbs.go:70-71 verbatim.
func RenderCrumbs(segments []string, width int) string {
	if len(segments) == 0 {
		// Defensive: empty row at least 1 cell tall (lipgloss.Height("") == 1).
		return CrumbRowStyle.Width(width).Render("")
	}
	normalised := normaliseSegments(segments)
	fitted := truncateSegmentsToWidth(normalised, width-2) // -2 for row pad

	chips := make([]string, 0, len(fitted))
	last := len(fitted) - 1
	for i, seg := range fitted {
		text := "<" + seg + ">"
		switch {
		case seg == ellipsisSentinel:
			chips = append(chips, CrumbChipEllipsisStyle.Render(text))
		case i == last:
			chips = append(chips, CrumbChipActiveStyle.Render(text))
		default:
			chips = append(chips, CrumbChipStyle.Render(text))
		}
	}
	joined := strings.Join(chips, " ") // D-208: single-space separator
	return CrumbRowStyle.Width(width).Render(joined)
}

// normaliseSegments applies D-207 verbatim from k9s crumbs.go:70-71:
//
//	strings.ReplaceAll(strings.ToLower(seg), " ", "")
//
// Centralised so existing m.status.SetBreadcrumb(...) call-sites stay
// untouched (D-210).
func normaliseSegments(segs []string) []string {
	out := make([]string, len(segs))
	for i, s := range segs {
		out[i] = strings.ReplaceAll(strings.ToLower(s), " ", "")
	}
	return out
}

// truncateSegmentsToWidth iteratively drops middle segments and inserts
// a U+2026 sentinel when the joined chip width + separators exceeds
// maxCells. Algorithm: drop middle-most segment first, replace with
// sentinel, re-measure, repeat until fits or only [first, sentinel, last]
// remains.
//
// The sentinel is a magic value RenderCrumbs recognises and styles via
// CrumbChipEllipsisStyle (muted fg, no bg).
func truncateSegmentsToWidth(segs []string, maxCells int) []string {
	if measureChipRow(segs) <= maxCells {
		return segs
	}
	work := append([]string(nil), segs...)
	// Insert one sentinel at midpoint on first overflow.
	if !containsEllipsisSentinel(work) && len(work) > 2 {
		mid := len(work) / 2
		work = append(work[:mid], append([]string{ellipsisSentinel}, work[mid+1:]...)...)
	}
	// Continue dropping non-sentinel middle segments until we fit
	// or only [first, sentinel, last] remains.
	for measureChipRow(work) > maxCells && len(work) > 3 {
		// Drop the segment immediately before the sentinel, preserving first + last.
		sentinelIdx := -1
		for i, s := range work {
			if s == ellipsisSentinel {
				sentinelIdx = i
				break
			}
		}
		if sentinelIdx <= 1 {
			// No room to drop without losing first or sentinel itself; stop.
			break
		}
		work = append(work[:sentinelIdx-1], work[sentinelIdx:]...)
	}
	return work
}

// measureChipRow returns the rendered cell width of a chip row:
//   - each chip = "<" + segment + ">" = 2 + lipgloss.Width(segment) cells
//   - between chips: +1 cell separator
func measureChipRow(segs []string) int {
	total := 0
	for i, seg := range segs {
		total += 2 + lipgloss.Width(seg)
		if i > 0 {
			total++ // single-space separator
		}
	}
	return total
}

func containsEllipsisSentinel(segs []string) bool {
	for _, s := range segs {
		if s == ellipsisSentinel {
			return true
		}
	}
	return false
}
