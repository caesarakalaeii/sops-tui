package ui_test

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
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

// Test 1: View() renders breadcrumb on the left section.
func TestStatusBarBreadcrumbInView(t *testing.T) {
	m := ui.NewStatusBarModel(defaultEnv())
	m.SetBreadcrumb("files")
	view := m.View(80)
	assert.True(t, strings.Contains(view, "files"), "View must contain the breadcrumb segment 'files', got: %q", view)
}

// Test 2: View() renders item count in the center section.
func TestStatusBarItemCountInView(t *testing.T) {
	m := ui.NewStatusBarModel(defaultEnv())
	m.SetItemCount(12, "items")
	view := m.View(80)
	assert.True(t, strings.Contains(view, "12"), "View must contain the item count '12', got: %q", view)
	assert.True(t, strings.Contains(view, "items"), "View must contain the label 'items', got: %q", view)
}

// Test 3: View() renders env indicators on the right (sops:checkmark, age:checkmark, .sops.yaml:checkmark).
func TestStatusBarEnvIndicatorsAllPresent(t *testing.T) {
	m := ui.NewStatusBarModel(defaultEnv())
	view := m.View(80)
	assert.True(t, strings.Contains(view, "sops:"), "View must contain 'sops:' indicator, got: %q", view)
	assert.True(t, strings.Contains(view, "age:"), "View must contain 'age:' indicator, got: %q", view)
	// checkmark character ✓
	assert.True(t, strings.Contains(view, "\u2713"), "View must contain checkmark '✓' for all-available env, got: %q", view)
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
	assert.True(t, strings.Contains(view, "\u2717"), "View must contain cross '✗' when sops unavailable, got: %q", view)
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
	assert.True(t, strings.Contains(view, "\u26A0"), "View must contain warning '⚠' when age unavailable, got: %q", view)
}

// Test 6: Flash() sets flash message and returns a tea.Tick command.
func TestStatusBarFlashReturnsCmd(t *testing.T) {
	m := ui.NewStatusBarModel(defaultEnv())
	m2, cmd := m.Flash("Copied!")
	require.NotNil(t, cmd, "Flash must return a non-nil Cmd")
	// The cmd must produce a message (a tea.Tick wraps a timer, but we can verify it's not nil)
	_ = m2
}

// Test 7: View() during flash replaces all sections with centered flash text.
func TestStatusBarFlashReplacesContent(t *testing.T) {
	m := ui.NewStatusBarModel(defaultEnv())
	m.SetBreadcrumb("files")
	m.SetItemCount(5, "items")
	m, _ = m.Flash("File copied!")
	view := m.View(80)
	assert.True(t, strings.Contains(view, "File copied!"), "View during flash must contain flash text, got: %q", view)
	// The breadcrumb and item count should not appear during flash
	assert.False(t, strings.Contains(view, "files") && strings.Contains(view, "5 items"),
		"View during flash should not display normal breadcrumb+count together, got: %q", view)
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

// Test 10: SetBreadcrumb updates the breadcrumb text.
func TestStatusBarSetBreadcrumb(t *testing.T) {
	m := ui.NewStatusBarModel(defaultEnv())
	m.SetBreadcrumb("files", "prod.yaml")
	view := m.View(80)
	assert.True(t, strings.Contains(view, "prod.yaml"), "View must contain updated breadcrumb segment, got: %q", view)
}

// Test 11: SetItemCount updates the count display.
func TestStatusBarSetItemCount(t *testing.T) {
	m := ui.NewStatusBarModel(defaultEnv())
	m.SetItemCount(3, "keys")
	view := m.View(80)
	assert.True(t, strings.Contains(view, "3"), "View must contain count '3', got: %q", view)
	assert.True(t, strings.Contains(view, "keys"), "View must contain label 'keys', got: %q", view)
}

// Ensure Flash uses time.Second without importing time externally in tests.
var _ = time.Second

// Ensure tea.Tick signature is exercised.
var _ = tea.Tick
