// Package health_test tests the health check pure functions.
//
// Note: Never use type any. Follow the table-driven test pattern from executor_test.go.
package health_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caesarakalaeii/sops-tui/internal/health"
)

// TestShannonEntropy verifies ShannonEntropy returns correct values for known inputs.
func TestShannonEntropy(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		minVal  float64
		maxVal  float64
		exactly *float64
	}{
		{
			name:  "empty string returns 0.0",
			input: "",
			minVal: 0.0,
			maxVal: 0.0,
		},
		{
			name:   "all same character returns near 0.0",
			input:  "aaaaaaaaaa",
			minVal: 0.0,
			maxVal: 0.001,
		},
		{
			name:   "diverse characters returns >= 3.5",
			input:  "aBcDeFgH1234!@#$",
			minVal: 3.5,
			maxVal: 10.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := health.ShannonEntropy(tc.input)
			assert.GreaterOrEqual(t, result, tc.minVal, "entropy should be >= %f", tc.minVal)
			assert.LessOrEqual(t, result, tc.maxVal, "entropy should be <= %f", tc.maxVal)
		})
	}
}

// TestIsWeakSecret verifies IsWeakSecret correctly identifies weak secrets.
func TestIsWeakSecret(t *testing.T) {
	tests := []struct {
		name           string
		keyPath        string
		value          string
		expectWeak     bool
		expectReason   string
	}{
		{
			name:         "too short",
			keyPath:      "db.password",
			value:        "short",
			expectWeak:   true,
			expectReason: "too short",
		},
		{
			name:         "low entropy",
			keyPath:      "db.password",
			value:        "aaaaaaaaaaaaaaaa",
			expectWeak:   true,
			expectReason: "low entropy",
		},
		{
			name:         "strong generic key",
			keyPath:      "db.password",
			value:        "xK9$mP2wL5nR8qF3vZ7yA4",
			expectWeak:   false,
			expectReason: "",
		},
		{
			name:         "format mismatch on _token suffix",
			keyPath:      "api_token",
			value:        "not-base64-not-hex!!",
			expectWeak:   true,
			expectReason: "format mismatch",
		},
		{
			name:         "valid base64 with _token suffix",
			keyPath:      "api_token",
			value:        "dGhpcyBpcyBhIHZhbGlkIGJhc2U2NA==",
			expectWeak:   false,
			expectReason: "",
		},
		{
			name:         "valid UUID with _key suffix",
			keyPath:      "service_key",
			value:        "550e8400-e29b-41d4-a716-446655440000",
			expectWeak:   false,
			expectReason: "",
		},
		{
			name:         "valid hex with _secret suffix",
			keyPath:      "db_secret",
			value:        "a1b2c3d4e5f6a7b8a1b2c3d4e5f6a7b8",
			expectWeak:   false,
			expectReason: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			weak, reason := health.IsWeakSecret(tc.keyPath, tc.value)
			assert.Equal(t, tc.expectWeak, weak, "IsWeakSecret(%q, %q) weak mismatch", tc.keyPath, tc.value)
			assert.Equal(t, tc.expectReason, reason, "IsWeakSecret(%q, %q) reason mismatch", tc.keyPath, tc.value)
		})
	}
}

// TestFindDuplicates verifies FindDuplicates correctly identifies duplicate values.
func TestFindDuplicates(t *testing.T) {
	t.Run("two files sharing same value returns 1 Duplicate with 2 Locations", func(t *testing.T) {
		fileValues := map[string]map[string]string{
			"secrets/a.yaml": {
				"db.password": "shared-secret-value-here",
			},
			"secrets/b.yaml": {
				"api.token": "shared-secret-value-here",
			},
		}
		dups := health.FindDuplicates(fileValues)
		require.Len(t, dups, 1, "expected exactly 1 duplicate group")
		assert.Len(t, dups[0].Locations, 2, "expected 2 locations for the duplicate")
		assert.NotEmpty(t, dups[0].ValueHash, "ValueHash must not be empty")
	})

	t.Run("no shared values returns empty slice", func(t *testing.T) {
		fileValues := map[string]map[string]string{
			"secrets/a.yaml": {
				"db.password": "unique-value-one-here",
			},
			"secrets/b.yaml": {
				"api.token": "unique-value-two-here",
			},
		}
		dups := health.FindDuplicates(fileValues)
		assert.Empty(t, dups, "expected no duplicates")
	})
}

// TestHealthCheckResult verifies the IsEmpty method works correctly.
func TestHealthCheckResult(t *testing.T) {
	t.Run("empty result returns true for IsEmpty", func(t *testing.T) {
		r := health.HealthCheckResult{}
		assert.True(t, r.IsEmpty(), "empty HealthCheckResult should be empty")
	})

	t.Run("result with WeakSecrets returns false for IsEmpty", func(t *testing.T) {
		r := health.HealthCheckResult{
			WeakSecrets: []health.WeakSecret{
				{FilePath: "a.yaml", KeyPath: "db.password", Reason: "too short"},
			},
		}
		assert.False(t, r.IsEmpty(), "HealthCheckResult with WeakSecrets should not be empty")
	})

	t.Run("result with Duplicates returns false for IsEmpty", func(t *testing.T) {
		r := health.HealthCheckResult{
			Duplicates: []health.Duplicate{
				{ValueHash: "abc123"},
			},
		}
		assert.False(t, r.IsEmpty(), "HealthCheckResult with Duplicates should not be empty")
	})

	t.Run("result with Errors returns false for IsEmpty", func(t *testing.T) {
		r := health.HealthCheckResult{
			Errors: []string{"failed to decrypt a.yaml"},
		}
		assert.False(t, r.IsEmpty(), "HealthCheckResult with Errors should not be empty")
	})
}

// TestHealthCheckResult_HasErrLevelFindings verifies the Phase 10 D-401 predicate
// excludes StaleFiles from the Err-level signal so stale-only results stay below
// Warn (logo stays Info per D-402).
func TestHealthCheckResult_HasErrLevelFindings(t *testing.T) {
	t.Run("zero-value result returns false", func(t *testing.T) {
		r := health.HealthCheckResult{}
		assert.False(t, r.HasErrLevelFindings(), "zero-value HealthCheckResult must not raise to Err")
	})

	t.Run("StaleFiles only returns false (D-401 demotion)", func(t *testing.T) {
		r := health.HealthCheckResult{
			StaleFiles: []health.StaleFile{{FilePath: "old.yaml", DaysSince: 120}},
		}
		assert.False(t, r.HasErrLevelFindings(),
			"D-401: stale files alone must NOT raise the logo to Err")
	})

	t.Run("WeakSecrets present returns true", func(t *testing.T) {
		r := health.HealthCheckResult{
			WeakSecrets: []health.WeakSecret{{FilePath: "f", KeyPath: "k", Reason: "too short"}},
		}
		assert.True(t, r.HasErrLevelFindings(),
			"weak secrets must raise to Err per D-401")
	})

	t.Run("Duplicates present returns true", func(t *testing.T) {
		r := health.HealthCheckResult{
			Duplicates: []health.Duplicate{{ValueHash: "abc123"}},
		}
		assert.True(t, r.HasErrLevelFindings(),
			"duplicates must raise to Err per D-401")
	})

	t.Run("Errors present returns true", func(t *testing.T) {
		r := health.HealthCheckResult{
			Errors: []string{"failed to decrypt a.yaml"},
		}
		assert.True(t, r.HasErrLevelFindings(),
			"scan errors must raise to Err per D-401")
	})

	t.Run("StaleFiles plus WeakSecrets returns true (any non-stale finding raises)", func(t *testing.T) {
		r := health.HealthCheckResult{
			StaleFiles:  []health.StaleFile{{FilePath: "old.yaml"}},
			WeakSecrets: []health.WeakSecret{{FilePath: "f", KeyPath: "k", Reason: "short"}},
		}
		assert.True(t, r.HasErrLevelFindings(),
			"any non-stale finding must raise to Err even with stale files present")
	})
}
