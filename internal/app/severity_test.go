// Package app_test verifies the Phase 10 severity classifier (resolveLogoState).
//
// Phase 10 D-401, D-402, D-403, D-404: pure-function-of-state classifier reads
// (env, flashSeverity, lastHealthResult) and returns LogoError / LogoWarn /
// LogoInfo with severity precedence Err > Warn > Info.
//
// The classifier is unexported so tests reach it through the test-only shim
// ResolveLogoStateForTest defined in export_test.go.
package app_test

import (
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/stretchr/testify/require"

	"github.com/caesarakalaeii/sops-tui/internal/app"
	"github.com/caesarakalaeii/sops-tui/internal/health"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

// newCleanAppModel builds an AppModel with all-good env (sops + age + .sops.yaml
// available) and no flash and no health findings — the LogoInfo baseline.
// Phase 10 D-419: passes colorprofile.TrueColor to match the lipgloss/v2 default
// test profile (zero render-time SGR delta vs. the pre-Plan-10 baseline).
func newCleanAppModel(t *testing.T) app.AppModel {
	t.Helper()
	env := ui.EnvStatus{
		SopsAvailable:     true,
		AgeAvailable:      true,
		SopsYamlAvailable: true,
	}
	return app.NewAppModel(env, "", colorprofile.TrueColor)
}

func TestResolveLogoState_DefaultIsInfo(t *testing.T) {
	m := newCleanAppModel(t)
	require.Equal(t, ui.LogoInfo, app.ResolveLogoStateForTest(m),
		"clean env + no flash + no health findings = LogoInfo baseline")
}

func TestResolveLogoState_FlashErrRaisesError(t *testing.T) {
	m := newCleanAppModel(t)
	sb, _ := m.StatusForTest().FlashErr("decrypt failed")
	m = m.WithStatusForTest(sb)
	require.Equal(t, ui.LogoError, app.ResolveLogoStateForTest(m),
		"flash Err must raise to LogoError per D-401")
}

func TestResolveLogoState_FlashWarnRaisesWarn(t *testing.T) {
	m := newCleanAppModel(t)
	sb, _ := m.StatusForTest().FlashWarn("no changes")
	m = m.WithStatusForTest(sb)
	require.Equal(t, ui.LogoWarn, app.ResolveLogoStateForTest(m),
		"flash Warn must raise to LogoWarn per D-402")
}

func TestResolveLogoState_FlashInfoStaysInfo(t *testing.T) {
	m := newCleanAppModel(t)
	sb, _ := m.StatusForTest().FlashInfo("decrypted")
	m = m.WithStatusForTest(sb)
	require.Equal(t, ui.LogoInfo, app.ResolveLogoStateForTest(m),
		"flash Info on clean env stays at LogoInfo baseline")
}

func TestResolveLogoState_HealthWeakSecretRaisesError(t *testing.T) {
	m := newCleanAppModel(t)
	h := m.HealthForTest()
	h.SetResults(health.HealthCheckResult{
		WeakSecrets: []health.WeakSecret{
			{FilePath: "f", KeyPath: "k", Reason: "too short"},
		},
	})
	m = m.WithHealthForTest(h)
	require.Equal(t, ui.LogoError, app.ResolveLogoStateForTest(m),
		"non-empty WeakSecrets must raise to LogoError per D-401")
}

func TestResolveLogoState_HealthDuplicateRaisesError(t *testing.T) {
	m := newCleanAppModel(t)
	h := m.HealthForTest()
	h.SetResults(health.HealthCheckResult{
		Duplicates: []health.Duplicate{{ValueHash: "abc123"}},
	})
	m = m.WithHealthForTest(h)
	require.Equal(t, ui.LogoError, app.ResolveLogoStateForTest(m),
		"non-empty Duplicates must raise to LogoError per D-401")
}

func TestResolveLogoState_HealthScanErrorRaisesError(t *testing.T) {
	m := newCleanAppModel(t)
	h := m.HealthForTest()
	h.SetResults(health.HealthCheckResult{
		Errors: []string{"locked.yaml: decrypt failed"},
	})
	m = m.WithHealthForTest(h)
	require.Equal(t, ui.LogoError, app.ResolveLogoStateForTest(m),
		"non-empty Errors must raise to LogoError per D-401")
}

func TestResolveLogoState_StaleFilesAloneStaysInfo(t *testing.T) {
	// D-401: StaleFiles alone do NOT raise to Err. With clean env + no flash
	// the logo must stay at the LogoInfo baseline. This proves the stale
	// demotion built into HasErrLevelFindings.
	m := newCleanAppModel(t)
	h := m.HealthForTest()
	h.SetResults(health.HealthCheckResult{
		StaleFiles: []health.StaleFile{{FilePath: "old.yaml", DaysSince: 120}},
	})
	m = m.WithHealthForTest(h)
	require.Equal(t, ui.LogoInfo, app.ResolveLogoStateForTest(m),
		"stale-files-only must NOT raise the logo per D-401 demotion")
}

func TestResolveLogoState_SoftEnvAgeMissingRaisesWarn(t *testing.T) {
	m := newCleanAppModel(t)
	sb := m.StatusForTest()
	sb.SetEnv(ui.EnvStatus{
		SopsAvailable:     true,
		AgeAvailable:      false,
		SopsYamlAvailable: true,
	})
	m = m.WithStatusForTest(sb)
	require.Equal(t, ui.LogoWarn, app.ResolveLogoStateForTest(m),
		"missing age key must raise to LogoWarn per D-402")
}

func TestResolveLogoState_SoftEnvSopsYamlMissingRaisesWarn(t *testing.T) {
	m := newCleanAppModel(t)
	sb := m.StatusForTest()
	sb.SetEnv(ui.EnvStatus{
		SopsAvailable:     true,
		AgeAvailable:      true,
		SopsYamlAvailable: false,
	})
	m = m.WithStatusForTest(sb)
	require.Equal(t, ui.LogoWarn, app.ResolveLogoStateForTest(m),
		"missing .sops.yaml must raise to LogoWarn per D-402")
}

func TestResolveLogoState_PrecedenceFlashErrOverSoftEnvWarn(t *testing.T) {
	// D-404: Err > Warn. Flash Err must win even when env is soft-warn.
	m := newCleanAppModel(t)
	sb, _ := m.StatusForTest().FlashErr("hard error")
	sb.SetEnv(ui.EnvStatus{
		SopsAvailable:     true,
		AgeAvailable:      false,
		SopsYamlAvailable: true,
	})
	m = m.WithStatusForTest(sb)
	require.Equal(t, ui.LogoError, app.ResolveLogoStateForTest(m),
		"flash Err must beat soft env Warn per D-404 precedence")
}

func TestResolveLogoState_PrecedenceHealthErrOverFlashWarn(t *testing.T) {
	// D-404: Err > Warn. Health Err findings must win even when the active
	// flash is only Warn-severity.
	m := newCleanAppModel(t)
	sb, _ := m.StatusForTest().FlashWarn("soft warning")
	m = m.WithStatusForTest(sb)
	h := m.HealthForTest()
	h.SetResults(health.HealthCheckResult{
		WeakSecrets: []health.WeakSecret{
			{FilePath: "f", KeyPath: "k", Reason: "too short"},
		},
	})
	m = m.WithHealthForTest(h)
	require.Equal(t, ui.LogoError, app.ResolveLogoStateForTest(m),
		"health Err must beat flash Warn per D-404 precedence")
}
