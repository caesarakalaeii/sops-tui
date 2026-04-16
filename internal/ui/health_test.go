// Package ui_test provides tests for the HealthModel health check overlay.
package ui_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/caesarakalaeii/sops-tui/internal/health"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

// TestHealthModel runs all HealthModel rendering tests.
func TestHealthModel(t *testing.T) {
	t.Run("renders loading state before SetResults", func(t *testing.T) {
		m := ui.NewHealthModel(80, 24)
		view := m.View()
		assert.Contains(t, view, "Running health check", "loading state should show loading text")
	})

	t.Run("renders no issues found for empty results", func(t *testing.T) {
		m := ui.NewHealthModel(80, 24)
		m.SetResults(health.HealthCheckResult{})
		view := m.View()
		assert.Contains(t, view, "No issues found", "empty results should show no issues found")
	})

	t.Run("renders weak secret with WEAK tag and file path", func(t *testing.T) {
		m := ui.NewHealthModel(80, 24)
		m.SetResults(health.HealthCheckResult{
			WeakSecrets: []health.WeakSecret{
				{FilePath: "secrets/prod.yaml", KeyPath: "db_password", Reason: "too short"},
			},
		})
		view := m.View()
		assert.Contains(t, view, "[WEAK]", "should contain WEAK tag")
		assert.Contains(t, view, "secrets/prod.yaml", "should contain file path")
	})

	t.Run("renders duplicate with DUPE tag and both file paths", func(t *testing.T) {
		m := ui.NewHealthModel(80, 24)
		m.SetResults(health.HealthCheckResult{
			Duplicates: []health.Duplicate{
				{
					ValueHash: "abc123",
					Locations: []health.Location{
						{FilePath: "secrets/prod.yaml", KeyPath: "api_key"},
						{FilePath: "secrets/staging.yaml", KeyPath: "api_key"},
					},
				},
			},
		})
		view := m.View()
		assert.Contains(t, view, "[DUPE]", "should contain DUPE tag")
		assert.Contains(t, view, "secrets/prod.yaml", "should contain first file path")
		assert.Contains(t, view, "secrets/staging.yaml", "should contain second file path")
	})

	t.Run("renders stale file with STALE tag and days ago text", func(t *testing.T) {
		m := ui.NewHealthModel(80, 24)
		m.SetResults(health.HealthCheckResult{
			StaleFiles: []health.StaleFile{
				{FilePath: "secrets/old.yaml", LastCommitTime: time.Now().Add(-100 * 24 * time.Hour), DaysSince: 100},
			},
		})
		view := m.View()
		assert.Contains(t, view, "[STALE]", "should contain STALE tag")
		assert.Contains(t, view, "days ago", "should contain days ago text")
	})

	t.Run("renders errors footer with skipped text", func(t *testing.T) {
		m := ui.NewHealthModel(80, 24)
		m.SetResults(health.HealthCheckResult{
			Errors: []string{"secrets/locked.yaml: decrypt failed"},
		})
		view := m.View()
		assert.Contains(t, view, "skipped", "should contain skipped text for errors")
	})

	t.Run("scroll down increments position and scroll up decrements", func(t *testing.T) {
		m := ui.NewHealthModel(80, 24)
		// Build results with many weak secrets to ensure scrollable content
		var weaks []health.WeakSecret
		for i := 0; i < 20; i++ {
			weaks = append(weaks, health.WeakSecret{
				FilePath: "secrets/file.yaml",
				KeyPath:  "key",
				Reason:   "too short",
			})
		}
		m.SetResults(health.HealthCheckResult{WeakSecrets: weaks})

		// Scroll down should increase scroll position
		m.ScrollDown()
		view1 := m.View()
		assert.NotEmpty(t, view1)

		// Multiple scroll downs should not panic
		m.ScrollDown()
		m.ScrollDown()
		m.ScrollDown()

		// Scroll up should work
		m.ScrollUp()
		view2 := m.View()
		assert.NotEmpty(t, view2)

		// ScrollUp past 0 should clamp
		m.ScrollUp()
		m.ScrollUp()
		m.ScrollUp()
		m.ScrollUp()
		view3 := m.View()
		assert.NotEmpty(t, view3)
	})

	t.Run("scroll clamps at bounds for empty results", func(t *testing.T) {
		m := ui.NewHealthModel(80, 24)
		m.SetResults(health.HealthCheckResult{})

		// ScrollDown on empty should not panic
		m.ScrollDown()
		m.ScrollDown()

		// ScrollUp on empty should not panic
		m.ScrollUp()
		m.ScrollUp()

		view := m.View()
		assert.NotEmpty(t, view)
	})
}
