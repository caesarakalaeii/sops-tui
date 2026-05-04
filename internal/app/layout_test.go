package app

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/stretchr/testify/assert"

	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

// defaultEnvInternal mirrors model_test.go:16-22 defaultEnv() for use
// inside package app. Duplication matches the repo convention.
func defaultEnvInternal() ui.EnvStatus {
	return ui.EnvStatus{
		SopsAvailable:     true,
		AgeAvailable:      true,
		SopsYamlAvailable: true,
	}
}

// TestBodyDims verifies that bodyDims returns the window width and the
// remaining height after the status bar, the live chrome (Phase 7), and
// the crumbs stub (Phase 8 still 0) are subtracted.
func TestBodyDims(t *testing.T) {
	m := NewAppModel(defaultEnvInternal(), "", colorprofile.TrueColor)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(AppModel)
	w, h := bodyDims(m)
	assert.Equal(t, 80, w, "width must equal m.width")
	expected := 24 - statusBarHeight(m) - chromeHeight(m) - crumbsHeight(m)
	assert.Equal(t, expected, h,
		"bodyDims subtracts chrome + crumbs + status bar from model height (chromeHeight is live in Phase 7)")
}

// TestBodyDimsClampsAtZero verifies bodyDims clamps negative heights to zero.
func TestBodyDimsClampsAtZero(t *testing.T) {
	m := NewAppModel(defaultEnvInternal(), "", colorprofile.TrueColor)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 1})
	m = updated.(AppModel)
	_, h := bodyDims(m)
	assert.GreaterOrEqual(t, h, 0,
		"bodyDims must clamp h to >= 0 on short terminals (Pitfall 1)")
}

// TestCrumbsHeightReturnsZero — Phase 6 stub invariant. Phase 8 flips this.
func TestCrumbsHeightReturnsZero(t *testing.T) {
	m := NewAppModel(defaultEnvInternal(), "", colorprofile.TrueColor)
	assert.Equal(t, 0, crumbsHeight(m),
		"Phase 6: crumbsHeight is a stub returning 0 until Phase 8")
}

// TestBodyDimsMigration enforces UI-18: the banned subtraction pattern
// (the expression subtracting the status-bar helper result from the model
// height) must not appear outside bodyDims.
//
// SELF-MATCH AVOIDANCE: The banned regex is assembled at runtime from THREE
// separate string constants below so this file does not contain the full
// contiguous pattern as one literal — in code, doc comments, or error text.
// Any paraphrased references in comments use wording like "the banned
// subtraction pattern" or "UI-18 banned expression".
//
// LIVE UNDER DEFAULT go test: Plan 2 migrated all 17 pre-migration sites to
// bodyDims(m) and deleted the temporary Plan-2 migration allowlist plus its
// filter branch atomically with that migration. The only legitimate
// occurrence of the banned pattern now lives inside bodyDims itself, carved
// out via the brace-depth range lookup below.
func TestBodyDimsMigration(t *testing.T) {
	// Assembled from three pieces: regex prefix, operator, helper name.
	// Never appears as one contiguous literal anywhere in this file.
	regexPrefix := `m\.height\s*`
	regexMinus := `-\s*`
	regexHelper := `statusBarHeight`
	banned := regexp.MustCompile(regexPrefix + regexMinus + regexHelper)

	repoRoot := findRepoRoot(t)
	helperStart, helperEnd := findBodyDimsRange(t, filepath.Join(repoRoot, "internal/app/model.go"))

	var violations []string
	err := filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == ".git" || (strings.HasPrefix(name, ".") && name != ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(repoRoot, path)
		for i, line := range strings.Split(string(content), "\n") {
			if !banned.MatchString(line) {
				continue
			}
			lineNo := i + 1
			// Carve out the bodyDims function body — the one legitimate
			// home of the subtraction expression.
			if strings.HasSuffix(path, "internal/app/model.go") &&
				lineNo >= helperStart && lineNo <= helperEnd {
				continue
			}
			violations = append(violations, rel+":"+strconv.Itoa(lineNo)+"  "+strings.TrimSpace(line))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
	if len(violations) > 0 {
		t.Fatalf("UI-18 violation: banned subtraction pattern found outside bodyDims:\n  %s\n\n"+
			"Use bodyDims(m) instead; see UI-17 and UI-18.",
			strings.Join(violations, "\n  "))
	}
}

// findRepoRoot walks up from this file's dir until go.mod is found.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	_, here, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller(0) failed")
	}
	dir := filepath.Dir(here)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found walking up from %s", filepath.Dir(here))
		}
		dir = parent
	}
}

// findBodyDimsRange locates the 1-indexed [start, end] line range of
// `func bodyDims(...)` in modelPath via brace-depth tracking.
// Brittle-by-design: relies on the stable D-01 signature.
func findBodyDimsRange(t *testing.T, modelPath string) (start, end int) {
	t.Helper()
	b, err := os.ReadFile(modelPath)
	if err != nil {
		t.Fatalf("read %s: %v", modelPath, err)
	}
	lines := strings.Split(string(b), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "func bodyDims(") {
			start = i + 1
			break
		}
	}
	if start == 0 {
		t.Fatalf("func bodyDims not found in %s", modelPath)
	}
	depth := 0
	for i := start - 1; i < len(lines); i++ {
		depth += strings.Count(lines[i], "{") - strings.Count(lines[i], "}")
		if depth == 0 && i >= start {
			return start, i + 1
		}
	}
	t.Fatalf("unterminated bodyDims body in %s", modelPath)
	return 0, 0
}
