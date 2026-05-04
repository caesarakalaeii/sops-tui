package ui_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

func TestRenderInfoPanel_AllRowsAligned(t *testing.T) {
	d := ui.InfoPanelData{
		SopsYamlRelPath: "secrets/.sops.yaml",
		AgeFingerprint:  "age1abcdefghijklmnop",
		RecipientCount:  4,
		GitBranch:       "main",
		GitDirty:        true,
		FileCount:       12,
	}
	out := ui.RenderInfoPanel(d, ui.PaletteFor(colorprofile.TrueColor))
	stripped := ansi.Strip(out)
	lines := strings.Split(stripped, "\n")
	// 5 rows in fixed order per D-201
	if len(lines) < 5 {
		t.Fatalf("expected at least 5 rows, got %d: %q", len(lines), stripped)
	}
	assert.True(t, strings.HasPrefix(lines[0], "cfg:"), "row 0 must start with 'cfg:'; got %q", lines[0])
	assert.True(t, strings.HasPrefix(lines[1], "age:"), "row 1 must start with 'age:'; got %q", lines[1])
	assert.True(t, strings.HasPrefix(lines[2], "rcp:"), "row 2 must start with 'rcp:'; got %q", lines[2])
	assert.True(t, strings.HasPrefix(lines[3], "git:"), "row 3 must start with 'git:'; got %q", lines[3])
	assert.True(t, strings.HasPrefix(lines[4], "fil:"), "row 4 must start with 'fil:'; got %q", lines[4])
}

func TestRenderInfoPanel_EmptyMarkers(t *testing.T) {
	// All-empty struct -> all rows render the "-" empty marker (D-204).
	d := ui.InfoPanelData{
		SopsYamlRelPath: "",
		AgeFingerprint:  "",
		RecipientCount:  -1,
		GitBranch:       "",
		FileCount:       -1,
	}
	stripped := ansi.Strip(ui.RenderInfoPanel(d, ui.PaletteFor(colorprofile.TrueColor)))
	// Each row must contain ": -" (label colon, space, hyphen-minus).
	for _, label := range []string{"cfg:", "age:", "rcp:", "git:", "fil:"} {
		line := findLine(t, stripped, label)
		assert.Contains(t, line, "-", "row %q must contain '-' empty marker", label)
	}
}

func TestRenderInfoPanel_TruncatesAge(t *testing.T) {
	// 30-char fingerprint -> must middle-truncate to <=10 cells with U+2026 (D-203).
	d := ui.InfoPanelData{AgeFingerprint: "age1abcdefghijklmnopqrstuvwxyz"}
	stripped := ansi.Strip(ui.RenderInfoPanel(d, ui.PaletteFor(colorprofile.TrueColor)))
	ageLine := findLine(t, stripped, "age:")
	// Strip "age:" + spaces from front.
	value := strings.TrimSpace(strings.TrimPrefix(ageLine, "age:"))
	assert.LessOrEqual(t, len([]rune(value)), 10, "age value must be <=10 cells (D-203, Pitfall 11); got %q (%d runes)", value, len([]rune(value)))
	assert.Contains(t, value, "…", "age value must contain U+2026 ellipsis when truncated")
}

func TestRenderInfoPanel_TruncatesPath(t *testing.T) {
	// 80-char path -> middle-truncates to 32 cells with U+2026 (D-203).
	d := ui.InfoPanelData{SopsYamlRelPath: strings.Repeat("a/", 40) + "prod.yaml"}
	stripped := ansi.Strip(ui.RenderInfoPanel(d, ui.PaletteFor(colorprofile.TrueColor)))
	cfgLine := findLine(t, stripped, "cfg:")
	value := strings.TrimSpace(strings.TrimPrefix(cfgLine, "cfg:"))
	assert.LessOrEqual(t, len([]rune(value)), 32, "cfg value must be <=32 cells (D-203); got %q", value)
	assert.Contains(t, value, "…", "cfg value must contain U+2026 ellipsis when truncated")
}

func TestRenderInfoPanel_GitDirtyMarker(t *testing.T) {
	d := ui.InfoPanelData{GitBranch: "main", GitDirty: true}
	stripped := ansi.Strip(ui.RenderInfoPanel(d, ui.PaletteFor(colorprofile.TrueColor)))
	gitLine := findLine(t, stripped, "git:")
	assert.Contains(t, gitLine, "main *", "dirty branch must trail with ' *' per D-215 (Claude's Discretion recommendation)")
}

func TestRenderInfoPanel_GitDetachedHead(t *testing.T) {
	d := ui.InfoPanelData{GitBranch: "abc1234", GitDetached: true, GitDirty: true}
	stripped := ansi.Strip(ui.RenderInfoPanel(d, ui.PaletteFor(colorprofile.TrueColor)))
	gitLine := findLine(t, stripped, "git:")
	assert.Contains(t, gitLine, "HEAD@abc1234", "detached HEAD must render as 'HEAD@<hash>' per D-215")
	assert.NotContains(t, gitLine, " *", "detached HEAD must NOT show dirty marker (mid-rebase too transient per D-215)")
}

// findLine returns the line in stripped that starts with prefix.
func findLine(t *testing.T, stripped, prefix string) string {
	t.Helper()
	for _, line := range strings.Split(stripped, "\n") {
		if strings.HasPrefix(line, prefix) {
			return line
		}
	}
	t.Fatalf("no line starting with %q in:\n%s", prefix, stripped)
	return ""
}
