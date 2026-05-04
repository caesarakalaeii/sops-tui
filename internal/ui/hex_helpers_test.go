// Phase 10 Plan 02 (D-417) test helpers — keep color-presence assertions
// derived from the named hex constants so future palette tunes only need
// to flip Color*Hex; tests that compare via this helper auto-update.
//
// The helper lives in package ui_test (same suffix as the hex-literal-using
// test files); other *_test.go files in the directory call it without an
// import statement.
package ui_test

import (
	"fmt"
	"strconv"
	"strings"
)

// hexToRGBTriplet converts a 6-digit hex color string ("#aabbcc") to the
// "r;g;b" triplet that lipgloss/v2 emits in 24-bit SGR sequences. Used by
// color-presence test assertions that need to verify a specific palette
// constant appears in rendered output, without hardcoding the literal RGB
// values (which would silently desync from D-415's palette flip).
//
// Phase 10 D-417: tests reference constants (ColorAccentHex etc.), not literals.
// This helper is the single mapping point — future palette tunes only need
// to flip the constant; tests that compare via this helper auto-update.
//
// Panics on malformed input (test-only helper; failures should be loud).
func hexToRGBTriplet(hex string) string {
	if !strings.HasPrefix(hex, "#") || len(hex) != 7 {
		panic(fmt.Sprintf("hexToRGBTriplet: malformed hex %q (expected '#RRGGBB')", hex))
	}
	r, err := strconv.ParseUint(hex[1:3], 16, 8)
	if err != nil {
		panic(fmt.Sprintf("hexToRGBTriplet: bad red component in %q: %v", hex, err))
	}
	g, err := strconv.ParseUint(hex[3:5], 16, 8)
	if err != nil {
		panic(fmt.Sprintf("hexToRGBTriplet: bad green component in %q: %v", hex, err))
	}
	b, err := strconv.ParseUint(hex[5:7], 16, 8)
	if err != nil {
		panic(fmt.Sprintf("hexToRGBTriplet: bad blue component in %q: %v", hex, err))
	}
	return fmt.Sprintf("%d;%d;%d", r, g, b)
}

// hexBgSGR returns the lipgloss/v2 24-bit background SGR substring
// "48;2;r;g;b" derived from a hex color. Convenience wrapper.
func hexBgSGR(hex string) string {
	return "48;2;" + hexToRGBTriplet(hex)
}

// hexFgSGR returns the lipgloss/v2 24-bit foreground SGR substring
// "38;2;r;g;b" derived from a hex color. Convenience wrapper.
func hexFgSGR(hex string) string {
	return "38;2;" + hexToRGBTriplet(hex)
}
