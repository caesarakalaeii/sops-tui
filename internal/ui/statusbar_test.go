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
