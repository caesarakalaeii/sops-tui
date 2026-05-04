package ui_test

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func defaultEnv() ui.EnvStatus {
	return ui.EnvStatus{
		SopsAvailable:     true,
		AgeAvailable:      true,
		SopsYamlAvailable: true,
	}
}

// Test 3: View() renders env indicators on the right (sops:checkmark, age:checkmark, .sops.yaml:checkmark).
func TestStatusBarEnvIndicatorsAllPresent(t *testing.T) {
	m := ui.NewStatusBarModel(defaultEnv())
	view := m.View(80)
	assert.True(t, strings.Contains(view, "sops:"), "View must contain 'sops:' indicator, got: %q", view)
	assert.True(t, strings.Contains(view, "age:"), "View must contain 'age:' indicator, got: %q", view)
	// checkmark character ✓
	assert.True(t, strings.Contains(view, "✓"), "View must contain checkmark '✓' for all-available env, got: %q", view)
}

// Test 4: View() with envSops=false renders sops:cross in error color.
func TestStatusBarSopsUnavailable(t *testing.T) {
	env := ui.EnvStatus{
		SopsAvailable:     false,
		AgeAvailable:      true,
		SopsYamlAvailable: true,
	}
	m := ui.NewStatusBarModel(env)
	view := m.View(80)
	// cross character ✗
	assert.True(t, strings.Contains(view, "✗"), "View must contain cross '✗' when sops unavailable, got: %q", view)
}

// Test 5: View() with envAge=false renders age:warning symbol in warning color.
func TestStatusBarAgeUnavailable(t *testing.T) {
	env := ui.EnvStatus{
		SopsAvailable:     true,
		AgeAvailable:      false,
		SopsYamlAvailable: true,
	}
	m := ui.NewStatusBarModel(env)
	view := m.View(80)
	// warning character ⚠
	assert.True(t, strings.Contains(view, "⚠"), "View must contain warning '⚠' when age unavailable, got: %q", view)
}

// Test 6: Flash() sets flash message and returns a tea.Tick command.
func TestStatusBarFlashReturnsCmd(t *testing.T) {
	m := ui.NewStatusBarModel(defaultEnv())
	_, cmd := m.Flash("Copied!")
	require.NotNil(t, cmd, "Flash must return a non-nil Cmd")
}

// Test 7: View() during flash replaces content with centered flash text.
func TestStatusBarFlashReplacesContent(t *testing.T) {
	m := ui.NewStatusBarModel(defaultEnv())
	m.SetBreadcrumb("files")
	m.SetItemCount(5, "items") // no-op per D-209; must not affect flash path
	m, _ = m.Flash("File copied!")
	view := m.View(80)
	assert.True(t, strings.Contains(view, "File copied!"), "View during flash must contain flash text, got: %q", view)
}

// Test 8: flashClearMsg with matching generation clears the flash.
func TestStatusBarFlashClearMatchingGen(t *testing.T) {
	m := ui.NewStatusBarModel(defaultEnv())
	m, _ = m.Flash("Test flash")
	// Simulate the tick firing with the correct generation
	clearMsg := ui.FlashClearMsg{Gen: 1}
	m2, _ := m.Update(clearMsg)
	view := m2.View(80)
	assert.False(t, strings.Contains(view, "Test flash"), "Flash must be cleared after matching FlashClearMsg, view: %q", view)
}

// Test 9: flashClearMsg with stale generation does NOT clear the flash (per Pitfall 6).
func TestStatusBarFlashClearStaleGen(t *testing.T) {
	m := ui.NewStatusBarModel(defaultEnv())
	m, _ = m.Flash("First flash")
	m, _ = m.Flash("Second flash") // gen is now 2
	// Stale clear for gen=1 should be ignored
	clearMsg := ui.FlashClearMsg{Gen: 1}
	m2, _ := m.Update(clearMsg)
	view := m2.View(80)
	assert.True(t, strings.Contains(view, "Second flash"), "Stale FlashClearMsg must NOT clear current flash, view: %q", view)
}

// TestStatusBar_RightAlignOnly verifies Phase 8 D-211: the normal
// (non-flash) path renders ONLY the right cluster (env indicators +
// optional clipboard) right-aligned on full-width surface bg. No
// breadcrumb, no item-count, no pipe separators.
func TestStatusBar_RightAlignOnly(t *testing.T) {
	m := ui.NewStatusBarModel(ui.EnvStatus{
		SopsAvailable:     true,
		AgeAvailable:      true,
		SopsYamlAvailable: true,
	})
	m.SetBreadcrumb("sops-tui", "files", "prod.yaml")
	m.SetItemCount(12, "items") // no-op per D-209; must not affect output

	out := m.View(80)
	stripped := ansi.Strip(out)

	// Right cluster present
	assert.Contains(t, stripped, "sops",
		"right cluster must render env indicators (sops:checkmark etc.)")

	// Deleted sections absent
	assert.NotContains(t, stripped, " > ",
		"breadcrumb LEFT section is deleted in Phase 8 D-211 (moved to RenderCrumbs)")
	assert.NotContains(t, stripped, "12 items",
		"item-count CENTER section is deleted in Phase 8 D-209/D-211 (moved to titled-border title)")
	assert.NotContains(t, stripped, " | ",
		"pipe separators between sections are deleted in Phase 8 D-211")

	// Width spans full width
	assert.Equal(t, 80, lipgloss.Width(stripped),
		"status bar surface bg must span full width via StatusBarStyle.Width(width)")
}

// TestStatusBar_SegmentsAccessor verifies Phase 8 D-210: Segments()
// returns the breadcrumb split on " > " (the same separator
// SetBreadcrumb joins with).
func TestStatusBar_SegmentsAccessor(t *testing.T) {
	m := ui.NewStatusBarModel(ui.EnvStatus{})
	m.SetBreadcrumb("files", "prod.yaml")
	got := m.Segments()
	assert.Equal(t, []string{"sops-tui", "files", "prod.yaml"}, got,
		"Segments must reverse SetBreadcrumb's strings.Join via strings.Split")
}

// TestStatusBar_SegmentsEmpty verifies that an empty breadcrumb
// returns nil (not []string{""}). This lets RenderCrumbs(nil, width)
// take its empty-row path cleanly.
// Uses a zero-value StatusBarModel (breadcrumb == "") to exercise the nil path.
func TestStatusBar_SegmentsEmpty(t *testing.T) {
	var zero ui.StatusBarModel
	assert.Nil(t, zero.Segments(),
		"empty breadcrumb must return nil, not []string{''}")
}

// TestStatusBar_FlashUnchanged verifies Phase 8 D-212: flash path is
// unchanged from v1.0 — center-aligned full-width flash text on
// surface bg.
func TestStatusBar_FlashUnchanged(t *testing.T) {
	m := ui.NewStatusBarModel(ui.EnvStatus{})
	m, _ = m.Flash("test message")
	out := m.View(80)
	stripped := ansi.Strip(out)
	assert.Contains(t, stripped, "test message",
		"flash path renders the message text (D-212 unchanged)")
	assert.Equal(t, 80, lipgloss.Width(stripped),
		"flash path spans full width (D-212 unchanged)")
}

// Ensure Flash uses time.Second without importing time externally in tests.
var _ = time.Second

// Ensure tea.Tick signature is exercised.
var _ = tea.Tick

// Phase 10 Plan 1: typed flash severity tests (D-406 .. D-412).

// TestStatusBar_FlashSeverityZeroValue verifies D-409: a freshly-constructed
// StatusBarModel with no flash fired returns FlashSevInfo from the accessor.
func TestStatusBar_FlashSeverityZeroValue(t *testing.T) {
	m := ui.NewStatusBarModel(ui.EnvStatus{})
	require.Equal(t, ui.FlashSevInfo, m.FlashSeverity(),
		"zero-value StatusBarModel must return FlashSevInfo (D-409)")
}

// TestStatusBar_FlashWarnSetsSeverity verifies D-406: FlashWarn sets the
// severity field accessible via FlashSeverity().
func TestStatusBar_FlashWarnSetsSeverity(t *testing.T) {
	m := ui.NewStatusBarModel(ui.EnvStatus{})
	m, _ = m.FlashWarn("test msg")
	require.Equal(t, ui.FlashSevWarn, m.FlashSeverity())
}

// TestStatusBar_FlashErrSetsSeverity verifies D-406: FlashErr sets the
// severity field accessible via FlashSeverity().
func TestStatusBar_FlashErrSetsSeverity(t *testing.T) {
	m := ui.NewStatusBarModel(ui.EnvStatus{})
	m, _ = m.FlashErr("oops")
	require.Equal(t, ui.FlashSevErr, m.FlashSeverity())
}

// TestStatusBar_FlashInfoExplicitlySetsInfo verifies D-406: FlashInfo
// explicitly sets the severity to Info (same as the zero value).
func TestStatusBar_FlashInfoExplicitlySetsInfo(t *testing.T) {
	m := ui.NewStatusBarModel(ui.EnvStatus{})
	m, _ = m.FlashInfo("hi")
	require.Equal(t, ui.FlashSevInfo, m.FlashSeverity())
}

// TestStatusBar_LegacyFlashAliasesInfo verifies D-406: the legacy
// Flash(msg) method remains a thin alias for FlashInfo so backward-
// compat is preserved for the unmoved neutral callsites.
func TestStatusBar_LegacyFlashAliasesInfo(t *testing.T) {
	m := ui.NewStatusBarModel(ui.EnvStatus{})
	m, _ = m.Flash("legacy")
	require.Equal(t, ui.FlashSevInfo, m.FlashSeverity(),
		"Flash must remain a thin alias for FlashInfo")
}

// TestStatusBar_FlashWarnRendersBracketWPrefix verifies D-411: Warn flash
// renders with "[W] " prefix at render time (prefix added at View(), not
// stored on m.flash).
func TestStatusBar_FlashWarnRendersBracketWPrefix(t *testing.T) {
	m := ui.NewStatusBarModel(ui.EnvStatus{})
	m, _ = m.FlashWarn("disk low")
	rendered := m.View(40)
	stripped := ansi.Strip(rendered)
	assert.Contains(t, stripped, "[W] disk low",
		"Warn flash must render with [W] prefix per D-411")
}

// TestStatusBar_FlashErrRendersBracketEPrefix verifies D-411: Err flash
// renders with "[E] " prefix at render time.
func TestStatusBar_FlashErrRendersBracketEPrefix(t *testing.T) {
	m := ui.NewStatusBarModel(ui.EnvStatus{})
	m, _ = m.FlashErr("decrypt failed")
	rendered := m.View(40)
	stripped := ansi.Strip(rendered)
	assert.Contains(t, stripped, "[E] decrypt failed",
		"Err flash must render with [E] prefix per D-411")
}

// TestStatusBar_FlashInfoNoPrefix verifies D-411: Info flash renders
// raw (no [I] prefix). Avoids day-to-day status message clutter.
func TestStatusBar_FlashInfoNoPrefix(t *testing.T) {
	m := ui.NewStatusBarModel(ui.EnvStatus{})
	m, _ = m.FlashInfo("decrypted")
	rendered := m.View(40)
	stripped := ansi.Strip(rendered)
	assert.Contains(t, stripped, "decrypted")
	assert.NotContains(t, stripped, "[I]",
		"Info flash must NOT have prefix per D-411")
	assert.NotContains(t, stripped, "[W]",
		"Info flash must NOT have W prefix")
	assert.NotContains(t, stripped, "[E]",
		"Info flash must NOT have E prefix")
}

// TestStatusBar_FlashWarnPaintsBgTint verifies D-412: Warn paints with
// ColorWarning bg + ColorBg fg.
// Phase 10 D-417: SGR substrings derived from ColorWarningHex / ColorBgHex
// via hexBgSGR / hexFgSGR — the palette flip from #f9e2af to #fab387
// (Catppuccin Peach) auto-propagates without test edits.
func TestStatusBar_FlashWarnPaintsBgTint(t *testing.T) {
	m := ui.NewStatusBarModel(ui.EnvStatus{})
	m, _ = m.FlashWarn("warn")
	rendered := m.View(40)
	warnBgSGR := hexBgSGR(ui.ColorWarningHex)
	bgFgSGR := hexFgSGR(ui.ColorBgHex)
	assert.Containsf(t, rendered, warnBgSGR,
		"Warn must paint warning bg (%s derived from ColorWarningHex %s) per D-412",
		warnBgSGR, ui.ColorWarningHex)
	assert.Containsf(t, rendered, bgFgSGR,
		"Warn must paint dark foreground (ColorBg %s) per D-412", bgFgSGR)
}

// TestStatusBar_FlashErrPaintsBgTint verifies D-412: Err paints with
// ColorError bg + ColorBg fg.
// Phase 10 D-417: SGR substrings derived from ColorErrorHex / ColorBgHex
// — the palette flip from #f38ba8 to #eba0ac (Catppuccin Maroon)
// auto-propagates without test edits.
func TestStatusBar_FlashErrPaintsBgTint(t *testing.T) {
	m := ui.NewStatusBarModel(ui.EnvStatus{})
	m, _ = m.FlashErr("err")
	rendered := m.View(40)
	errBgSGR := hexBgSGR(ui.ColorErrorHex)
	bgFgSGR := hexFgSGR(ui.ColorBgHex)
	assert.Containsf(t, rendered, errBgSGR,
		"Err must paint error bg (%s derived from ColorErrorHex %s) per D-412",
		errBgSGR, ui.ColorErrorHex)
	assert.Containsf(t, rendered, bgFgSGR,
		"Err must paint dark foreground (ColorBg %s) per D-412", bgFgSGR)
}

// TestStatusBar_FlashInfoUsesSurfaceBg verifies D-412: Info renders on
// the existing StatusBarStyle surface bg (no peach/maroon SGR).
// Phase 10 D-417: SGR substrings derived from constants so the negative
// assertions auto-track the palette tune.
func TestStatusBar_FlashInfoUsesSurfaceBg(t *testing.T) {
	m := ui.NewStatusBarModel(ui.EnvStatus{})
	m, _ = m.FlashInfo("decrypted")
	rendered := m.View(40)
	warnBgSGR := hexBgSGR(ui.ColorWarningHex)
	errBgSGR := hexBgSGR(ui.ColorErrorHex)
	surfaceBgSGR := hexBgSGR(ui.ColorSurfaceHex)
	assert.NotContainsf(t, rendered, warnBgSGR,
		"Info flash must NOT use warning bg (%s) per D-412", warnBgSGR)
	assert.NotContainsf(t, rendered, errBgSGR,
		"Info flash must NOT use error bg (%s) per D-412", errBgSGR)
	assert.Containsf(t, rendered, surfaceBgSGR,
		"Info flash must use surface bg (%s) via StatusBarStyle", surfaceBgSGR)
}

// TestStatusBar_FlashClearMsgClearsSeverity verifies D-410: FlashClearMsg
// ack clears severity to baseline (FlashSevInfo) on the same tick that
// clears m.flash.
func TestStatusBar_FlashClearMsgClearsSeverity(t *testing.T) {
	m := ui.NewStatusBarModel(ui.EnvStatus{})
	m, _ = m.FlashErr("oops")
	require.Equal(t, ui.FlashSevErr, m.FlashSeverity())
	// First flash → flashGen incremented to 1; matching ack.
	m, _ = m.Update(ui.FlashClearMsg{Gen: 1})
	assert.Equal(t, ui.FlashSevInfo, m.FlashSeverity(),
		"FlashClearMsg ack must clear severity to baseline (D-410)")
}

// TestStatusBar_StaleFlashClearMsgPreservesSeverity verifies Pitfall 6:
// stale FlashClearMsg with mismatched Gen does NOT clear severity.
func TestStatusBar_StaleFlashClearMsgPreservesSeverity(t *testing.T) {
	m := ui.NewStatusBarModel(ui.EnvStatus{})
	m, _ = m.FlashErr("first")  // flashGen → 1
	m, _ = m.FlashWarn("second") // flashGen → 2; severity now Warn
	// Stale tick from gen=1 must NOT clear.
	m, _ = m.Update(ui.FlashClearMsg{Gen: 1})
	assert.Equal(t, ui.FlashSevWarn, m.FlashSeverity(),
		"Stale FlashClearMsg must NOT alter severity (Pitfall 6)")
}
