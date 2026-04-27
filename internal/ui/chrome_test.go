package ui

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
)

// firstLine returns the substring of s up to (but excluding) the first
// '\n', or s itself if no newline is present. Helper for tests that
// inspect the top border line of a rendered NormalBorder box.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// TestOverlayTitle_PreservesCornersAndWidth covers the 7-point contract
// from 07-RESEARCH.md §"Closed Research Gaps" #1 - the community-standard
// title-overlay pattern for lipgloss v2 (which has no native border-title
// API). Each subtest exercises one corner of the contract; together they
// close the STATE.md "Phase 7 research pass on overlayTitle" pending todo
// with a tested deliverable.
func TestOverlayTitle_PreservesCornersAndWidth(t *testing.T) {
	box := TitledBorderStyle.Width(18).Height(4).Render("body")

	// Note on corner glyphs: lipgloss v2's NormalBorder() produces square
	// corners ┌┐└┘ (NOT rounded ╭╮╰╯). UI-SPEC visual sketches and PATTERNS
	// snippets that drew rounded corners predate Plan 1's empirical Phase 7
	// rendering verification. The Phase 7 grep-gate (TestChromeASCIIOnly,
	// owned by Plan 3) must allowlist ┌─┐│└┘… - the rounded glyphs are
	// never emitted by NormalBorder. Recorded as deviation in 07-02-SUMMARY.md.
	t.Run("preserves top-left corner", func(t *testing.T) {
		got := overlayTitle(box, " Title ")
		stripped := ansi.Strip(firstLine(got))
		require.True(t, strings.HasPrefix(stripped, "┌"),
			"first line must start with ┌ (NormalBorder top-left), got %q", stripped)
	})

	t.Run("preserves top-right corner", func(t *testing.T) {
		got := overlayTitle(box, " Title ")
		stripped := ansi.Strip(firstLine(got))
		require.True(t, strings.HasSuffix(stripped, "┐"),
			"first line must end with ┐ (NormalBorder top-right), got %q", stripped)
	})

	t.Run("width unchanged", func(t *testing.T) {
		got := overlayTitle(box, " Title ")
		require.Equal(t,
			lipgloss.Width(firstLine(box)),
			lipgloss.Width(firstLine(got)),
			"overlay must preserve first-line width exactly")
	})

	t.Run("overlong title truncated with ellipsis", func(t *testing.T) {
		narrow := TitledBorderStyle.Width(14).Height(4).Render("body")
		got := overlayTitle(narrow, " an extremely long title that exceeds width ")
		stripped := ansi.Strip(firstLine(got))
		require.Contains(t, stripped, "…",
			"overlong title must be truncated with the ellipsis rune, got %q", stripped)
		require.Equal(t,
			lipgloss.Width(firstLine(narrow)),
			lipgloss.Width(firstLine(got)),
			"truncation must still preserve the first-line width exactly")
	})

	t.Run("empty title returns unchanged", func(t *testing.T) {
		got := overlayTitle(box, "")
		require.Equal(t, box, got,
			"empty title must return input byte-identical")
	})

	t.Run("single-line input returns unchanged", func(t *testing.T) {
		input := "no-newline-here"
		got := overlayTitle(input, " Title ")
		require.Equal(t, input, got,
			"single-line input has no border to modify; must return unchanged")
	})

	t.Run("width lt 4 returns unchanged", func(t *testing.T) {
		// Synthetic: a first line with visible width exactly 3.
		// overlayTitle must bail out (firstLineWidth < 4).
		input := "╭─╮\nmid\n╰─╯"
		got := overlayTitle(input, " Title ")
		require.Equal(t, input, got,
			"width < 4 too narrow for title injection; must return unchanged")
	})
}

// TestWrapTitled verifies WrapTitled composes a NormalBorder box at the
// requested envelope with the title injected via overlayTitle.
func TestWrapTitled(t *testing.T) {
	t.Run("wraps body with title at top border column 2", func(t *testing.T) {
		got := WrapTitled("Files (12)", "contents", 30, 5)
		stripped := ansi.Strip(firstLine(got))
		wantTitle := " Files (12) "
		require.Contains(t, stripped, wantTitle,
			"first border line must contain %q after overlay, got %q", wantTitle, stripped)
	})

	t.Run("clamps too-small dimensions", func(t *testing.T) {
		got := WrapTitled("", "", 0, 0)
		require.NotEmpty(t, got, "WrapTitled must never return empty even with 0,0 dims")
		require.GreaterOrEqual(t, lipgloss.Height(got), 3,
			"clamp must produce at least 3 rows")
		require.GreaterOrEqual(t, lipgloss.Width(firstLine(got)), 4,
			"clamp must produce at least 4 cols on first line")
	})

	t.Run("empty title renders only border chars and spaces on first line", func(t *testing.T) {
		// WrapTitled always passes " "+title+" " to overlayTitle. Empty
		// title => "  " (two spaces) injected at column 2. The first line
		// should still consist exclusively of NormalBorder chars and
		// spaces - no alphanumerics, no ANSI-visible artifacts.
		got := WrapTitled("", "body", 20, 5)
		stripped := ansi.Strip(firstLine(got))
		for _, r := range stripped {
			ok := r == '┌' || r == '┐' || r == '─' || r == ' '
			require.True(t, ok,
				"empty-title first line has unexpected rune %q in %q", string(r), stripped)
		}
	})

	t.Run("uses NormalBorder not thick double or rounded-alt variants", func(t *testing.T) {
		got := WrapTitled("X", "body", 20, 5)
		stripped := ansi.Strip(got)
		require.Contains(t, stripped, "─", "NormalBorder horizontal char must appear")
		require.NotContains(t, stripped, "━", "ThickBorder horizontal must not appear")
		require.NotContains(t, stripped, "┏", "ThickBorder corner must not appear")
		require.NotContains(t, stripped, "╔", "DoubleBorder corner must not appear")
	})

	t.Run("outer dimensions match requested width and height", func(t *testing.T) {
		got := WrapTitled("X", "body", 20, 5)
		stripped := ansi.Strip(got)
		lines := strings.Split(strings.TrimRight(stripped, "\n"), "\n")
		require.Equal(t, 5, len(lines), "height must be 5 rows")
		require.Equal(t, 20, lipgloss.Width(lines[0]),
			"width of first line must be 20, got %d (line=%q)",
			lipgloss.Width(lines[0]), lines[0])
	})
}

// TestRenderChrome verifies the 6-row persistent chrome band composes
// correctly at multiple widths with and without hints.
func TestRenderChrome(t *testing.T) {
	t.Run("returns exactly 6 rows at any width", func(t *testing.T) {
		for _, w := range []int{40, 80, 120, 200} {
			chrome := RenderChrome(nil, LogoInfo, w)
			require.Equal(t, 6, lipgloss.Height(chrome),
				"chrome height must be 6 at width %d, got %d",
				w, lipgloss.Height(chrome))
		}
	})

	t.Run("contains logo art at right-anchored position", func(t *testing.T) {
		chrome := RenderChrome(nil, LogoInfo, 200)
		stripped := ansi.Strip(chrome)
		// Logo Candidate A signature markers (logo.go LogoSmall): the row
		// 4 baseline contains "|____/" twice and row 5 contains "tui".
		require.Contains(t, stripped, "|____/", "logo baseline |____/ must appear")
		require.Contains(t, stripped, "tui", "logo subscript 'tui' must appear")
	})

	t.Run("contains menu mnemonics when hints given", func(t *testing.T) {
		hints := []keys.MenuHint{
			{Mnemonic: "j", Description: "move down", Visible: true},
			{Mnemonic: "k", Description: "move up", Visible: true},
			{Mnemonic: "?", Description: "help", Visible: true},
			{Mnemonic: "q", Description: "quit", Visible: true},
		}
		chrome := RenderChrome(hints, LogoInfo, 200)
		stripped := ansi.Strip(chrome)
		for _, mnem := range []string{"[j]", "[k]", "[?]", "[q]"} {
			require.Contains(t, stripped, mnem,
				"menu must render mnemonic %s in chrome output", mnem)
		}
	})

	t.Run("reserves info panel width via placeholder", func(t *testing.T) {
		chrome := RenderChrome(nil, LogoInfo, 200)
		stripped := ansi.Strip(chrome)
		lines := strings.Split(strings.TrimRight(stripped, "\n"), "\n")
		require.GreaterOrEqual(t, len(lines), 6,
			"chrome must render at least 6 rows")
		for i, line := range lines[:6] {
			if len(line) < 38 {
				continue // short-clipped row OK (defensive)
			}
			// Leftmost 38 cols must be whitespace (info-panel placeholder
			// reserves the column for Phase 8 without painting content).
			prefix := line[:38]
			require.Equal(t, "", strings.TrimSpace(prefix),
				"row %d leftmost 38 cols must be blank (info-panel placeholder), got %q",
				i, prefix)
		}
	})
}

// fullHintsFixture returns 12 visible MenuHints for chrome composition tests
// (Phase 7.1 D-120). Mirrors a saturated FileList hint set so RenderMenu
// fills both columns.
func fullHintsFixture() []keys.MenuHint {
	h := make([]keys.MenuHint, 12)
	for i := 0; i < 12; i++ {
		h[i] = keys.MenuHint{
			Mnemonic:    string(rune('a' + i)),
			Description: fmt.Sprintf("action %d", i),
			Visible:     true,
		}
	}
	return h
}

// TestRenderChrome_NarrowFallback asserts the narrow-tier (width < 33)
// fallback per Phase 7.1 D-116: single-line "press ? for help" stub.
// Closes WR-03 / 07-VERIFICATION.md Anti-Pattern 4 (chrome overflows at
// narrow widths, body unreachable at 40x12).
func TestRenderChrome_NarrowFallback(t *testing.T) {
	out := RenderChrome(fullHintsFixture(), LogoInfo, 30)

	// Single line — narrow tier produces a one-row stub.
	require.Equal(t, 1, lipgloss.Height(out),
		"narrow-tier RenderChrome must be exactly 1 row tall")

	// Stub text present (after ANSI stripping for color).
	plain := ansi.Strip(out)
	require.Contains(t, plain, "press ? for help",
		"narrow-tier output must contain the muted help stub")

	// ASCII-only invariant per UI-15 — every rune <= 0x7F.
	for _, r := range plain {
		require.LessOrEqualf(t, r, rune(0x7F),
			"narrow-tier output contains non-ASCII rune %q at byte: %s", r, plain)
	}
}

// TestRenderChrome_DropsInfoPanel asserts the mid-tier (33 <= width < 71)
// fallback per Phase 7.1 D-116: menu+logo only, info-panel slot dropped.
// At mid-tier the leftmost columns of every row must contain menu
// content (first column of the menu), NOT the 38-col blank info-panel
// placeholder that the full-tier reserves.
func TestRenderChrome_DropsInfoPanel(t *testing.T) {
	out := RenderChrome(fullHintsFixture(), LogoInfo, 60)

	// 6-row chrome (max(menuRows, logo rows) = 6).
	require.LessOrEqual(t, lipgloss.Height(out), 6,
		"mid-tier RenderChrome must fit in <= 6 rows")

	plain := ansi.Strip(out)
	firstLine := strings.SplitN(plain, "\n", 2)[0]
	require.LessOrEqual(t, lipgloss.Width(firstLine), 60,
		"mid-tier first line must not exceed requested width")

	// Mid-tier drops the info-panel placeholder; the leftmost cols of at
	// least one row should contain non-space content (the menu starts at
	// col 0). If the info-panel was still rendered, the leftmost 38 cols
	// would be all-blank on every row.
	var sawNonBlankLeft bool
	for _, line := range strings.Split(plain, "\n") {
		leadCheck := line
		if len(leadCheck) > 8 {
			leadCheck = leadCheck[:8]
		}
		if strings.TrimSpace(leadCheck) != "" {
			sawNonBlankLeft = true
			break
		}
	}
	require.True(t, sawNonBlankLeft,
		"mid-tier first 8 cols on at least one row must contain non-space content (info-panel dropped)")
}

// TestRenderChrome_FullChrome asserts the full-tier (width >= 71)
// 3-slot layout per Phase 7.1 D-116: info-panel + menu + logo, exactly
// 6 rows tall. This is the existing Phase 7 behaviour preserved by the
// 7.1 width-fallback addition.
func TestRenderChrome_FullChrome(t *testing.T) {
	out := RenderChrome(fullHintsFixture(), LogoInfo, 200)

	// Full chrome is exactly 6 rows per D-16 (Phase 7).
	require.Equal(t, 6, lipgloss.Height(out),
		"full-tier RenderChrome must be exactly 6 rows")

	// First line should extend toward the requested 200-col width.
	plain := ansi.Strip(out)
	firstLine := strings.SplitN(plain, "\n", 2)[0]
	require.GreaterOrEqual(t, lipgloss.Width(firstLine), 100,
		"full-tier first line must extend toward the requested 200-col width")
}

// TestRenderMenu_NoCellWrap asserts the manual-columns implementation
// per Phase 7.1 D-117 never wraps cells vertically. Compare to the
// Phase 7 lipgloss/v2/table builder which would produce 16+ rows at
// width=60 due to per-cell text wrapping (the root cause of WR-03 at
// 80x24).
func TestRenderMenu_NoCellWrap(t *testing.T) {
	out := RenderMenu(fullHintsFixture(), 60)
	require.Equal(t, 6, lipgloss.Height(out),
		"manual-columns RenderMenu must always produce exactly 6 rows; "+
			"vertical cell wrapping is forbidden by D-117")
}
