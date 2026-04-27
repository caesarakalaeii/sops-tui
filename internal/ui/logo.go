// Package ui - logo primitive (Phase 7).
//
// RenderLogo returns the 6-row ASCII logo rendered in the requested
// severity color. Phase 7 callers pass LogoInfo unconditionally (D-02);
// Phase 10 (UI-03) flips between Info/Warn/Error from aggregate severity.
//
// The art is ASCII-only per UI-15 and Pitfall 6 - no emoji, no VS16
// variation selectors, no ZWJ. TestChromeASCIIOnly (Plan 3) grep-gates
// this file against regressions.
//
// Byte-art: Candidate A from Phase 7 research (5-row SOPS block figlet +
// row-6 "tui" subscript per D-01). Width: 25 cols.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"strings"
)

// LogoStatus drives logo recoloring by aggregate app severity. Phase 7
// uses LogoInfo unconditionally; Phase 10 (UI-03) derives the value from
// env checks, flash severity, and aggregate health.
type LogoStatus int

const (
	// LogoInfo renders the logo in ColorAccent (default).
	LogoInfo LogoStatus = iota
	// LogoWarn renders the logo in ColorWarning (Phase 10 wiring).
	LogoWarn
	// LogoError renders the logo in ColorError (Phase 10 wiring).
	LogoError
)

// LogoSmall is the 6-row "SOPS" block + "tui" subscript per D-01.
// Width: 25 cols (trailing space pads to 25 on most rows; row 4 baseline
// has the P-tail and closes to exactly 25 cols). ASCII-only (all runes
// <= 0x7F) per UI-15 and Pitfall 6.
var LogoSmall = []string{
	`  ____   ___  ____  ____  `,
	` / ___| / _ \|  _ \/ ___| `,
	` \___ \| | | | |_) \___ \ `,
	`  ___) | |_| |  __/ ___) |`,
	` |____/ \___/|_|   |____/ `,
	`                      tui `,
}

// RenderLogo returns the 6-row logo string rendered in the style
// corresponding to the requested severity. The width parameter is
// plumbed for Phase 10 width-responsive layouts; Phase 7 ignores it
// (the art is locked at ~25 cols per D-01).
func RenderLogo(status LogoStatus, width int) string {
	_ = width // reserved for Phase 10 width-responsive logic (D-02)
	art := strings.Join(LogoSmall, "\n")
	switch status {
	case LogoWarn:
		return LogoStyleWarn.Render(art)
	case LogoError:
		return LogoStyleError.Render(art)
	default:
		return LogoStyleInfo.Render(art)
	}
}
