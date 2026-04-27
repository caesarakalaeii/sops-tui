package ui

import (
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
