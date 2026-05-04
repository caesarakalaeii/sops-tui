package app_test

import (
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/caesarakalaeii/sops-tui/internal/app"
	"github.com/caesarakalaeii/sops-tui/internal/testutil"
)

// setDeterministicAgeEnv points SOPS_AGE_KEY_FILE at a guaranteed-missing
// path so loadAgeFingerprint returns "" and the age: row renders "-"
// deterministically across all environments (Phase 8 D-220 / golden
// review-safety). Without this, goldens would contain host-specific
// age key fingerprints.
func setDeterministicAgeEnv(t *testing.T) {
	t.Helper()
	t.Setenv("SOPS_AGE_KEY_FILE", filepath.Join(t.TempDir(), "nonexistent-keys.txt"))
}

// TestResize_40x12 verifies that the top-level App View renders structurally
// correctly at a narrow terminal size. Golden under testdata/resize_40x12.golden.
// Pitfall 1 warns that 80x24 goldens alone can hide chrome-height regressions;
// 40x12 exercises the clamp-at-zero path in bodyDims when chrome lands.
// Phase 8 Pitfall F: 40x12 WILL change because crumbs row is now visible at
// narrow tier per D-216 (crumbsHeight is independent of chrome tier).
func TestResize_40x12(t *testing.T) {
	setDeterministicAgeEnv(t)
	m := app.NewAppModel(defaultEnv(), "", colorprofile.TrueColor)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 12})
	m = updated.(app.AppModel)

	v := m.View()
	testutil.RequireGoldenStructure(t, "resize_40x12", v.Content)
	// Phase 6: no color assertions (scaffolding only; Phase 7 populates).
	testutil.RequireGoldenColors(t, "resize_40x12", v.Content, nil)
}

// TestResize_80x24 — standard terminal baseline.
func TestResize_80x24(t *testing.T) {
	setDeterministicAgeEnv(t)
	m := app.NewAppModel(defaultEnv(), "", colorprofile.TrueColor)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(app.AppModel)

	v := m.View()
	testutil.RequireGoldenStructure(t, "resize_80x24", v.Content)
	testutil.RequireGoldenColors(t, "resize_80x24", v.Content, nil)
}

// TestResize_120x40 — mid-range wide terminal.
func TestResize_120x40(t *testing.T) {
	setDeterministicAgeEnv(t)
	m := app.NewAppModel(defaultEnv(), "", colorprofile.TrueColor)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = updated.(app.AppModel)

	v := m.View()
	testutil.RequireGoldenStructure(t, "resize_120x40", v.Content)
	testutil.RequireGoldenColors(t, "resize_120x40", v.Content, nil)
}

// TestResize_200x60 — large terminal; matches BenchmarkAppView's dimensions.
func TestResize_200x60(t *testing.T) {
	setDeterministicAgeEnv(t)
	m := app.NewAppModel(defaultEnv(), "", colorprofile.TrueColor)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 60})
	m = updated.(app.AppModel)

	v := m.View()
	testutil.RequireGoldenStructure(t, "resize_200x60", v.Content)
	testutil.RequireGoldenColors(t, "resize_200x60", v.Content, nil)
}
