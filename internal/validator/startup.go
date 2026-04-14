// Package validator provides startup environment checks for sops-tui.
//
// RunChecks validates that the sops binary is on PATH (hard error), that the
// age key file exists (soft warning), and that a .sops.yaml is discoverable in
// the directory hierarchy (soft warning).  All checks run in a single pass per
// D-02 so the user sees every issue at once.
package validator

import (
	"os"
	"os/exec"
	"path/filepath"
)

// Severity classifies a ValidationResult as a hard error or a soft warning.
type Severity int

const (
	// SeverityError is a hard error that prevents the TUI from launching (e.g. sops missing).
	SeverityError Severity = iota
	// SeverityWarn is a soft warning; the TUI still launches but functionality is degraded.
	SeverityWarn
)

// ValidationResult holds a single check outcome with a human-readable message and
// an actionable fix instruction.
type ValidationResult struct {
	Severity Severity
	Message  string
	Fix      string
}

// Options configures RunChecks for testability.  Production code passes
// DefaultOptions(); tests override individual fields.
type Options struct {
	// SopsLookPath overrides exec.LookPath for the sops binary lookup.
	// Defaults to exec.LookPath.
	SopsLookPath func(file string) (string, error)
	// AgeKeyPath is the path to the age keys.txt file.
	// Defaults to ~/.config/sops/age/keys.txt.
	AgeKeyPath string
	// StartDir is the directory to begin .sops.yaml discovery from.
	// Defaults to the current working directory.
	StartDir string
}

// DefaultOptions returns Options populated with production defaults.
func DefaultOptions() (Options, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Options{}, err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return Options{}, err
	}
	return Options{
		SopsLookPath: exec.LookPath,
		AgeKeyPath:   filepath.Join(home, ".config", "sops", "age", "keys.txt"),
		StartDir:     cwd,
	}, nil
}

// RunChecks performs all startup validation checks in a single pass (D-02).
//
// It returns the collected results and a boolean indicating whether any hard
// errors were found.  A non-empty result slice with hasHardError=false means
// only soft warnings were collected — the TUI should still launch.
func RunChecks(opts Options) ([]ValidationResult, bool) {
	var results []ValidationResult
	hasHardError := false

	// 1. sops binary check (HLT-01, D-03 hard error).
	if _, err := opts.SopsLookPath("sops"); err != nil {
		results = append(results, ValidationResult{
			Severity: SeverityError,
			Message:  "sops binary not found",
			Fix:      "Install sops: https://github.com/getsops/sops#install",
		})
		hasHardError = true
	}

	// 2. age key check (HLT-02, D-03 soft warning).
	if _, err := os.Stat(opts.AgeKeyPath); err != nil {
		results = append(results, ValidationResult{
			Severity: SeverityWarn,
			Message:  "Age key file not found",
			Fix:      "Create key: age-keygen -o ~/.config/sops/age/keys.txt",
		})
	}

	// 3. .sops.yaml discovery (D-04 soft warning).
	if _, found := FindSopsYaml(opts.StartDir); !found {
		results = append(results, ValidationResult{
			Severity: SeverityWarn,
			Message:  ".sops.yaml not found in current directory or parents",
			Fix:      "Run sops-tui in a repository with a .sops.yaml configuration",
		})
	}

	return results, hasHardError
}

// FindSopsYaml walks up the directory tree from startDir looking for a
// .sops.yaml file.  It returns the path and true on the first match, or
// ("", false) when it reaches the filesystem root without finding one.
//
// The function uses filepath.Dir termination (parent == dir) to avoid an
// infinite loop at the root — this satisfies threat T-01-04.
//
// FindSopsYaml is exported so that later phases (file discovery) can reuse it
// without re-implementing the walk logic.
func FindSopsYaml(startDir string) (string, bool) {
	dir := startDir
	for {
		candidate := filepath.Join(dir, ".sops.yaml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the filesystem root — stop.
			return "", false
		}
		dir = parent
	}
}
