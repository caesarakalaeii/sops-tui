package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

// BenchmarkAppView records the v1.0-baseline render cost so Phase 7's
// chrome skeleton has a concrete "before" number to compare its
// <= 50 us/op target against. No gating on absolute value in Phase 6 (D-12).
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
	m := NewAppModel(env, "")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 60})
	m = updated.(AppModel)

	b.ReportAllocs()
	for b.Loop() {
		_ = m.View()
	}
}
