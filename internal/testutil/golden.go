// Package testutil provides test harness helpers shared across sops-tui tests.
//
// RequireGoldenStructure performs ANSI-stripped structural comparison against
// testdata/<name>.golden fixtures. RequireGoldenColors asserts raw ANSI byte
// presence separately so structural goldens stay stable across lipgloss bumps
// (Pitfall 8).
//
// Per Phase 6 D-08: no external golden library (goldie, teatest.RequireEqualOutput)
// is used; x/ansi.Strip + string compare covers the need in ~30 LOC.
// Per Phase 6 D-10: regeneration is gated on GOLDEN_UPDATE=1 env var rather than
// a -update flag, to avoid reflexive fixture churn.
//
// Repo conventions: all helper signatures use concrete types (no empty
// interface type). Never use lipgloss.AdaptiveColor (issue #1036).
package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// RequireGoldenStructure asserts that ansi.Strip(output) equals the content
// of testdata/<name>.golden after trailing-whitespace + line-ending
// normalisation. When GOLDEN_UPDATE=1 is set, the file is (re)written
// instead of compared — intentional friction per D-10.
func RequireGoldenStructure(t *testing.T, name, output string) {
	t.Helper()
	stripped := normalise(ansi.Strip(output))
	path := filepath.Join("testdata", name+".golden")

	if os.Getenv("GOLDEN_UPDATE") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(path, []byte(stripped), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("GOLDEN_UPDATE=1: wrote %s", path)
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v  (hint: run with GOLDEN_UPDATE=1 to create)", path, err)
	}
	got := stripped
	if got != string(want) {
		t.Fatalf("golden mismatch for %s:\n--- want ---\n%s\n--- got ---\n%s\n\n"+
			"(hint: if this change is intentional, re-run with GOLDEN_UPDATE=1)",
			name, string(want), got)
	}
}

// RequireGoldenColors asserts that each wantColors substring appears in
// the raw (non-stripped) output. Empty wantColors is a no-op — Phase 6
// scaffolding for Phase 7 to populate when chrome colors land.
func RequireGoldenColors(t *testing.T, name, output string, wantColors []string) {
	t.Helper()
	for _, c := range missingColors(output, wantColors) {
		t.Errorf("%s: expected color sequence %q not present in output", name, c)
	}
}

// missingColors returns the subset of wantColors that do not appear as a
// substring of output. Factored out of RequireGoldenColors so self-tests
// can assert on the missing-list without triggering t.Errorf on a real
// *testing.T (subtests propagate failure and cannot be "caught").
func missingColors(output string, wantColors []string) []string {
	var missing []string
	for _, c := range wantColors {
		if !strings.Contains(output, c) {
			missing = append(missing, c)
		}
	}
	return missing
}

// MissingColorsForTest is a test-only re-export of missingColors for
// self-tests. Do NOT call in production code; use RequireGoldenColors.
func MissingColorsForTest(output string, wantColors []string) []string {
	return missingColors(output, wantColors)
}

// normalise strips trailing whitespace per line and normalises line endings
// to LF. Prevents golden drift from editor auto-format rules (Pitfall 8).
func normalise(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}

// NormaliseForTest is a test-only re-export of normalise for self-tests.
// Do NOT call in production code; use RequireGoldenStructure instead.
func NormaliseForTest(s string) string {
	return normalise(s)
}
