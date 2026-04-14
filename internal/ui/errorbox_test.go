// Package ui tests the styled stderr error box renderer.
package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/caesarakalaeii/sops-tui/internal/validator"
	"github.com/stretchr/testify/assert"
)

// TestErrorBoxSeverityError verifies that a SeverityError result produces "[ERROR]" label
// and includes the error message in the output.
func TestErrorBoxSeverityError(t *testing.T) {
	results := []validator.ValidationResult{
		{
			Severity: validator.SeverityError,
			Message:  "sops binary not found",
			Fix:      "Install sops: https://github.com/getsops/sops#install",
		},
	}

	var buf bytes.Buffer
	RenderErrorBox(results, true, &buf)

	output := buf.String()
	assert.Contains(t, output, "[ERROR]", "output should contain [ERROR] label")
	assert.Contains(t, output, "sops binary not found", "output should contain the error message")
}

// TestErrorBoxSeverityWarn verifies that a SeverityWarn result produces "[WARN]" label
// and includes the warning message in the output.
func TestErrorBoxSeverityWarn(t *testing.T) {
	results := []validator.ValidationResult{
		{
			Severity: validator.SeverityWarn,
			Message:  "Age key file not found",
			Fix:      "Create key: age-keygen -o ~/.config/sops/age/keys.txt",
		},
	}

	var buf bytes.Buffer
	RenderErrorBox(results, false, &buf)

	output := buf.String()
	assert.Contains(t, output, "[WARN]", "output should contain [WARN] label")
	assert.Contains(t, output, "Age key file not found", "output should contain the warning message")
}

// TestErrorBoxMixed verifies that mixed error+warn results both appear in output.
func TestErrorBoxMixed(t *testing.T) {
	results := []validator.ValidationResult{
		{
			Severity: validator.SeverityError,
			Message:  "sops binary not found",
			Fix:      "Install sops: https://github.com/getsops/sops#install",
		},
		{
			Severity: validator.SeverityWarn,
			Message:  "Age key file not found",
			Fix:      "Create key: age-keygen -o ~/.config/sops/age/keys.txt",
		},
	}

	var buf bytes.Buffer
	RenderErrorBox(results, true, &buf)

	output := buf.String()
	assert.Contains(t, output, "[ERROR]", "output should contain [ERROR] label")
	assert.Contains(t, output, "[WARN]", "output should contain [WARN] label")
	assert.Contains(t, output, "sops binary not found")
	assert.Contains(t, output, "Age key file not found")
}

// TestErrorBoxWarningsOnlyBorder verifies warnings-only case does not use error border styling.
// We check this indirectly: the header text differs between hard error and warnings-only.
func TestErrorBoxWarningsOnlyBorder(t *testing.T) {
	results := []validator.ValidationResult{
		{
			Severity: validator.SeverityWarn,
			Message:  "Age key file not found",
			Fix:      "Create key: age-keygen -o ~/.config/sops/age/keys.txt",
		},
	}

	var buf bytes.Buffer
	RenderErrorBox(results, false, &buf)

	output := buf.String()
	// When hasHardError=false the header must be "sops-tui: warnings", not "sops-tui: startup failed".
	assert.Contains(t, output, "sops-tui: warnings", "warnings-only header should say 'sops-tui: warnings'")
	assert.NotContains(t, output, "sops-tui: startup failed", "warnings-only header should not say 'startup failed'")
}

// TestErrorBoxWritesToWriter verifies RenderErrorBox writes to the provided io.Writer.
func TestErrorBoxWritesToWriter(t *testing.T) {
	results := []validator.ValidationResult{
		{
			Severity: validator.SeverityError,
			Message:  "sops binary not found",
			Fix:      "Install sops: https://github.com/getsops/sops#install",
		},
	}

	var buf bytes.Buffer
	RenderErrorBox(results, true, &buf)

	assert.True(t, buf.Len() > 0, "RenderErrorBox should write output to the provided writer")
}

// TestErrorBoxFixText verifies that fix/resolution text appears in the output.
func TestErrorBoxFixText(t *testing.T) {
	results := []validator.ValidationResult{
		{
			Severity: validator.SeverityError,
			Message:  "sops binary not found",
			Fix:      "Install sops: https://github.com/getsops/sops#install",
		},
		{
			Severity: validator.SeverityWarn,
			Message:  "Age key file not found",
			Fix:      "Create key: age-keygen -o ~/.config/sops/age/keys.txt",
		},
	}

	var buf bytes.Buffer
	RenderErrorBox(results, true, &buf)

	output := buf.String()
	assert.Contains(t, output, "https://github.com/getsops/sops#install", "fix text for sops should appear in output")
	assert.Contains(t, output, "age-keygen", "fix text for age key should appear in output")
}

// TestErrorBoxHardErrorHeader verifies hasHardError=true renders "sops-tui: startup failed" header.
func TestErrorBoxHardErrorHeader(t *testing.T) {
	results := []validator.ValidationResult{
		{
			Severity: validator.SeverityError,
			Message:  "sops binary not found",
			Fix:      "Install sops: https://github.com/getsops/sops#install",
		},
	}

	var buf bytes.Buffer
	RenderErrorBox(results, true, &buf)

	output := buf.String()
	assert.True(t,
		strings.Contains(output, "sops-tui: startup failed"),
		"hard error header should say 'sops-tui: startup failed'")
}
