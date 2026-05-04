package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

// BenchmarkAppView records the v1.0-baseline render cost so Phase 7's
// chrome skeleton has a concrete "before" number to compare its
// <= 50 us/op target against.
//
// Phase 11 D-504 closure: the gate at chrome_test.go:TestBenchmarkAppView_UnderBudget
// is now ACTIVE. The cache wired in Phase 11 Plan 01 (D-501..D-503) brought
// the per-frame cost from the empirical 2.4-2.8 ms baseline to under the
// 50,000 ns (50 µs) budget by amortising RenderChrome / RenderMenu /
// JoinHorizontal across frames where (state, recipientAction, IsSearchActive,
// width) is unchanged. Cache hit rate at steady state is asserted by
// TestChromeCache_HitRateAtSteadyState (chrome_test.go).
//
// Uses testing.B.Loop() (Go 1.26+ idiom) — do NOT revert to `for i := 0; i < b.N; i++`.
// Fixed at 200x60 (D-12) — do NOT parametrise into sub-benchmarks.
func BenchmarkAppView(b *testing.B) {
	env := ui.EnvStatus{
		SopsAvailable:     true,
		AgeAvailable:      true,
		SopsYamlAvailable: true,
		GitAvailable:      true,
	}
	m := NewAppModel(env, "", colorprofile.TrueColor)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 60})
	m = updated.(AppModel)

	b.ReportAllocs()
	for b.Loop() {
		_ = m.View()
	}
}
