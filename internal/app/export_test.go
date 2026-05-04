// Package app exposes test-only shims for unexported AppModel methods.
//
// File suffix _test.go ensures these helpers compile only during go test runs;
// they never appear in production binaries. The shims let tests in package
// app_test reach unexported methods (resolveLogoState) and unexported field
// values (status, health) without polluting the public API surface.
package app

import "github.com/caesarakalaeii/sops-tui/internal/ui"

// ResolveLogoStateForTest is a test-only re-export of resolveLogoState.
// Phase 10 Plan 1 / D-403: pure function of state, walks Err checks first
// (D-404 precedence) and returns LogoError / LogoWarn / LogoInfo.
func ResolveLogoStateForTest(m AppModel) ui.LogoStatus {
	return m.resolveLogoState()
}

// StatusForTest exposes m.status for severity-classifier table tests.
// Returns a value copy; tests must round-trip mutated state via WithStatusForTest.
func (m AppModel) StatusForTest() ui.StatusBarModel {
	return m.status
}

// WithStatusForTest returns a copy of m with status replaced.
// Used by severity tests to drive flash severity through the model without
// touching unexported field references.
func (m AppModel) WithStatusForTest(sb ui.StatusBarModel) AppModel {
	m.status = sb
	return m
}

// HealthForTest exposes m.health for severity-classifier table tests.
// Returns a value copy; tests must round-trip mutated state via WithHealthForTest.
func (m AppModel) HealthForTest() ui.HealthModel {
	return m.health
}

// WithHealthForTest returns a copy of m with health replaced.
func (m AppModel) WithHealthForTest(h ui.HealthModel) AppModel {
	m.health = h
	return m
}
