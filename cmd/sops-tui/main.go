// Package main is the entry point for sops-tui.
//
// Startup flow (per D-01, D-02, D-03, Pitfall 3):
//  1. Run all startup validation checks before initializing any TUI session.
//  2. If any issues exist, render a styled error/warning box to stderr.
//  3. Hard errors (sops missing) → print box and exit 1.
//  4. Soft warnings (age key missing, no .sops.yaml) → print box and launch TUI.
//  5. Build EnvStatus from validation results for the status bar indicators.
//  6. Create and run the Bubble Tea program with the root AppModel.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/caesarakalaeii/sops-tui/internal/app"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
	"github.com/caesarakalaeii/sops-tui/internal/validator"
)

func main() {
	// Step 1: Build validation options with production defaults (D-02)
	opts, err := validator.DefaultOptions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "sops-tui: failed to determine defaults: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Run all startup validation checks in a single pass (D-02)
	results, hasHardError := validator.RunChecks(opts)

	// Step 3: Render styled error/warning box to stderr if any issues exist (D-01)
	if len(results) > 0 {
		ui.RenderErrorBox(results, hasHardError, os.Stderr)
	}

	// Step 4: Hard error → exit before starting TUI (D-03, T-01-09)
	if hasHardError {
		os.Exit(1)
	}

	// Step 5: Build env status from validation results for the status bar
	env := ui.EnvStatus{
		SopsAvailable:     !hasResultWithMessage(results, "sops binary not found"),
		AgeAvailable:      !hasResultWithMessage(results, "Age key file not found"),
		SopsYamlAvailable: !hasResultWithMessage(results, ".sops.yaml not found"),
	}

	// Step 6: Create and run the root TUI program (View().AltScreen = true in AppModel)
	sopsYamlPath, _ := validator.FindSopsYaml(opts.StartDir)
	model := app.NewAppModel(env, sopsYamlPath)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running sops-tui: %v\n", err)
		os.Exit(1)
	}
}

// hasResultWithMessage returns true if any ValidationResult message contains substr.
// Used to derive boolean availability flags for the status bar from validation output.
func hasResultWithMessage(results []validator.ValidationResult, substr string) bool {
	for _, r := range results {
		if strings.Contains(r.Message, substr) {
			return true
		}
	}
	return false
}
