// Tests for the Phase 7 persistent keybinding menu: RenderMenu (D-04, D-05,
// D-06, D-07; UI-SPEC §"Menu" lines 417-449). The menu is a 2-col x 6-row
// grid built on charm.land/lipgloss/v2/table with column-major fill.
package ui_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

// visHint builds a visible MenuHint for terse test setup.
func visHint(mnem, desc string) keys.MenuHint {
	return keys.MenuHint{Mnemonic: mnem, Description: desc, Visible: true}
}

// invisHint builds a Visible=false MenuHint to verify D-06 filtering.
func invisHint(mnem, desc string) keys.MenuHint {
	return keys.MenuHint{Mnemonic: mnem, Description: desc, Visible: false}
}

// TestRenderMenu_ReturnsNonEmpty verifies a 6-hint render produces output
// containing each mnemonic in bracketed form per D-07 cell composition.
func TestRenderMenu_ReturnsNonEmpty(t *testing.T) {
	hints := []keys.MenuHint{
		visHint("j/↓", "move down"),
		visHint("k/↑", "move up"),
		visHint("/", "search"),
		visHint("?", "toggle help"),
		visHint("q", "quit"),
		visHint("enter", "open"),
	}
	out := ui.RenderMenu(hints, ui.PaletteFor(colorprofile.TrueColor), 80)
	stripped := ansi.Strip(out)
	require.NotEmpty(t, stripped)
	assert.Contains(t, stripped, "[j/↓]")
	assert.Contains(t, stripped, "[k/↑]")
	assert.Contains(t, stripped, "[/]")
	assert.Contains(t, stripped, "[?]")
	assert.Contains(t, stripped, "[q]")
	assert.Contains(t, stripped, "[enter]")
}

// TestRenderMenu_ColumnMajorFill verifies D-07: hints 0..5 land in column 0
// and hints 6..11 land in column 1, row by row left-to-right.
func TestRenderMenu_ColumnMajorFill(t *testing.T) {
	hints := []keys.MenuHint{
		visHint("a", "alpha"),
		visHint("b", "bravo"),
		visHint("c", "charlie"),
		visHint("d", "delta"),
		visHint("e", "echo"),
		visHint("f", "foxtrot"),
		visHint("g", "golf"),
		visHint("h", "hotel"),
		visHint("i", "india"),
		visHint("j", "juliett"),
		visHint("k", "kilo"),
		visHint("l", "lima"),
	}
	out := ui.RenderMenu(hints, ui.PaletteFor(colorprofile.TrueColor), 80)
	stripped := ansi.Strip(out)
	lines := strings.Split(stripped, "\n")
	require.GreaterOrEqual(t, len(lines), 6, "expected at least 6 content rows")

	// Column-major: rows[r][col] = hints[col*6 + r].
	// Row 0: alpha (col 0) then golf (col 1).
	// Row 1: bravo (col 0) then hotel (col 1).
	// ... etc.
	expectedPairs := [][2]string{
		{"alpha", "golf"},
		{"bravo", "hotel"},
		{"charlie", "india"},
		{"delta", "juliett"},
		{"echo", "kilo"},
		{"foxtrot", "lima"},
	}
	for i, pair := range expectedPairs {
		line := lines[i]
		col0Idx := strings.Index(line, pair[0])
		col1Idx := strings.Index(line, pair[1])
		require.GreaterOrEqualf(t, col0Idx, 0, "row %d missing col-0 desc %q in line: %q", i, pair[0], line)
		require.GreaterOrEqualf(t, col1Idx, 0, "row %d missing col-1 desc %q in line: %q", i, pair[1], line)
		require.Lessf(t, col0Idx, col1Idx,
			"row %d: col-0 description %q must precede col-1 description %q (column-major fill, D-07)",
			i, pair[0], pair[1])
	}
}

// TestRenderMenu_InvisibleHintsSkipped verifies D-06 — Visible=false hints
// are filtered out before column-major fill so visible hints occupy the
// first 9 slots without gaps.
func TestRenderMenu_InvisibleHintsSkipped(t *testing.T) {
	hints := []keys.MenuHint{
		visHint("a", "alpha"),
		visHint("b", "bravo"),
		visHint("c", "charlie"),
		visHint("d", "delta"),
		visHint("e", "echo"),
		invisHint("hidden1", "hidden-desc-one"),
		invisHint("hidden2", "hidden-desc-two"),
		invisHint("hidden3", "hidden-desc-three"),
		visHint("f", "foxtrot"),
		visHint("g", "golf"),
		visHint("h", "hotel"),
		visHint("i", "india"),
	}
	out := ui.RenderMenu(hints, ui.PaletteFor(colorprofile.TrueColor), 80)
	stripped := ansi.Strip(out)

	// All visible descriptions must appear.
	for _, desc := range []string{"alpha", "bravo", "charlie", "delta", "echo",
		"foxtrot", "golf", "hotel", "india"} {
		assert.Containsf(t, stripped, desc, "visible description %q missing from output", desc)
	}
	// All invisible descriptions must NOT appear.
	for _, desc := range []string{"hidden-desc-one", "hidden-desc-two", "hidden-desc-three"} {
		assert.NotContainsf(t, stripped, desc, "invisible description %q must not render", desc)
	}
}

// TestRenderMenu_CapsAt12Hints verifies D-04/D-06 — only the first 12
// visible hints render; surplus is discarded (full set remains in the ?
// help overlay per UI-11).
func TestRenderMenu_CapsAt12Hints(t *testing.T) {
	hints := make([]keys.MenuHint, 0, 20)
	for i := 0; i < 20; i++ {
		hints = append(hints, visHint(string(rune('a'+i)), "desc-"+string(rune('a'+i))))
	}
	out := ui.RenderMenu(hints, ui.PaletteFor(colorprofile.TrueColor), 80)
	stripped := ansi.Strip(out)
	// First 12 (a..l) must appear.
	for i := 0; i < 12; i++ {
		mnem := "[" + string(rune('a'+i)) + "]"
		assert.Containsf(t, stripped, mnem, "first 12 mnemonic %q must appear", mnem)
	}
	// Hints 12..19 (m..t) must NOT appear.
	for i := 12; i < 20; i++ {
		desc := "desc-" + string(rune('a'+i))
		assert.NotContainsf(t, stripped, desc, "surplus description %q must be truncated", desc)
	}
}

// TestRenderMenu_EmptyHints verifies safe handling of nil and empty input.
func TestRenderMenu_EmptyHints(t *testing.T) {
	p := ui.PaletteFor(colorprofile.TrueColor)
	t.Run("nil input no panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			_ = ui.RenderMenu(nil, p, 80)
		})
	})
	t.Run("empty input no panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			_ = ui.RenderMenu([]keys.MenuHint{}, p, 80)
		})
	})
}

// TestRenderMenu_NarrowTerminalSafe verifies D-04 narrow-terminal contract —
// at width 40 (the minimum supported) and width 10 (degenerate), the
// renderer must not panic. Output may be clipped.
func TestRenderMenu_NarrowTerminalSafe(t *testing.T) {
	hints := []keys.MenuHint{
		visHint("a", "alpha"),
		visHint("b", "bravo"),
		visHint("c", "charlie"),
		visHint("d", "delta"),
		visHint("e", "echo"),
		visHint("f", "foxtrot"),
		visHint("g", "golf"),
		visHint("h", "hotel"),
		visHint("i", "india"),
		visHint("j", "juliett"),
	}
	p := ui.PaletteFor(colorprofile.TrueColor)
	t.Run("width 40", func(t *testing.T) {
		assert.NotPanics(t, func() {
			_ = ui.RenderMenu(hints, p, 40)
		})
	})
	t.Run("width 10", func(t *testing.T) {
		assert.NotPanics(t, func() {
			_ = ui.RenderMenu(hints, p, 10)
		})
	})
}

// TestRenderMenu_ASCIIOnlyBody verifies UI-15 + Pitfall 6 — chrome content
// must be ASCII-only except for explicit allowlist runes (arrow keys from
// existing keymaps, plus the box-drawing allowlist if borders ever leak).
func TestRenderMenu_ASCIIOnlyBody(t *testing.T) {
	hints := []keys.MenuHint{
		visHint("j/↓", "move down"),
		visHint("k/↑", "move up"),
		visHint("h/←", "collapse"),
		visHint("l/→", "expand"),
		visHint("?", "toggle help"),
		visHint("q", "quit"),
	}
	out := ui.RenderMenu(hints, ui.PaletteFor(colorprofile.TrueColor), 80)
	stripped := ansi.Strip(out)

	// Allowlist mirrors internal/app/chrome_test.go:46-54 (TestChromeASCIIOnly,
	// the canonical chrome ASCII allowlist). NormalBorder() emits SQUARE
	// corners ┌┐└┘, NOT rounded ╭╮╰╯; per Phase 7 Plan 02 empirical finding.
	// Drop ╭╮╰╯ — they are never produced by lipgloss.NormalBorder() and
	// keeping them in the allowlist would silently accept the wrong corner
	// family if a future change re-enables a border on RenderMenu's table.
	allowed := map[rune]bool{
		'\n': true,
		// NormalBorder square corners + horizontals/verticals.
		'─': true, '│': true, '┌': true, '┐': true, '└': true, '┘': true,
		// Ellipsis used by ansi.Truncate in overlayTitle (defensive).
		'…': true,
		// Arrow runes used in key bindings (defensive; not in menu source today).
		'↑': true, '↓': true, '←': true, '→': true,
	}
	for _, r := range stripped {
		if r > 0x7F && !allowed[r] {
			t.Fatalf("non-ASCII rune U+%04X (%q) not in allowlist; got output: %q", r, r, stripped)
		}
	}
}

// TestRenderMenu_AccentAppliedToMnemonic verifies D-05 — MenuKeyStyle
// (foreground ColorAccent) is applied to the bracketed mnemonic.
// lipgloss emits TrueColor SGR sequences containing the raw r;g;b values.
// Phase 10 D-417: triplet derived from ColorAccentHex via hexToRGBTriplet.
func TestRenderMenu_AccentAppliedToMnemonic(t *testing.T) {
	hints := []keys.MenuHint{
		visHint("j", "move down"),
	}
	accentTriplet := hexToRGBTriplet(ui.ColorAccentHex)
	out := ui.RenderMenu(hints, ui.PaletteFor(colorprofile.TrueColor), 80)
	assert.Containsf(t, out, accentTriplet,
		"MenuKeyStyle (ColorAccent) must apply RGB triplet (%s) derived from ColorAccentHex %s",
		accentTriplet, ui.ColorAccentHex)
}
