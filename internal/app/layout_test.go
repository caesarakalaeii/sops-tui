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
	"github.com/caesarakalaeii/sops-tui/internal/ui"
	"github.com/stretchr/testify/assert"
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
// remaining height after the status bar is subtracted, while the chrome
// and crumbs stubs return 0 (the Phase 6 invariant).
func TestBodyDims(t *testing.T) {
	m := NewAppModel(defaultEnvInternal(), "")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(AppModel)
	w, h := bodyDims(m)
	assert.Equal(t, 80, w, "width must equal m.width")
	assert.Equal(t, 24-statusBarHeight(m), h,
		"height must equal model height minus status bar height while chromeHeight and crumbsHeight stubs return 0")
}

// TestBodyDimsClampsAtZero verifies bodyDims clamps negative heights to zero.
func TestBodyDimsClampsAtZero(t *testing.T) {
	m := NewAppModel(defaultEnvInternal(), "")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 1})
	m = updated.(AppModel)
	_, h := bodyDims(m)
	assert.GreaterOrEqual(t, h, 0,
		"bodyDims must clamp h to >= 0 on short terminals (Pitfall 1)")
}

// TestChromeHeightReturnsZero — Phase 6 stub invariant. Phase 7 flips this.
func TestChromeHeightReturnsZero(t *testing.T) {
	m := NewAppModel(defaultEnvInternal(), "")
	assert.Equal(t, 0, chromeHeight(m),
		"Phase 6: chromeHeight is a stub returning 0 until Phase 7")
}

// TestCrumbsHeightReturnsZero — Phase 6 stub invariant. Phase 8 flips this.
func TestCrumbsHeightReturnsZero(t *testing.T) {
	m := NewAppModel(defaultEnvInternal(), "")
	assert.Equal(t, 0, crumbsHeight(m),
		"Phase 6: crumbsHeight is a stub returning 0 until Phase 8")
}

// plan2MigrationAllowlist enumerates the 17 known unmigrated call-sites as
// of Phase 6 Plan 1. Plan 2 migrates all 17 sites AND deletes this variable
// plus the corresponding filter branch below in one atomic commit.
// See .planning/phases/06-layout-groundwork/06-02-PLAN.md Task 1 Step F.
//
// Source: 06-RESEARCH.md "Call-Site Inventory" (15 SetSize sites + 2 outliers).
// Note: the outlier previously documented at line 1799 now lives at 1828
// post-Plan-1 because Task 1's helper block shifted everything in the tail
// of model.go by 29 lines. Line 1333 remains unchanged (it sits before the
// insertion point) and uses the `statusBarH` local — it does not match the
// banned regex today but is listed here for Plan 2's reference.
var plan2MigrationAllowlist = map[string][]int{
	"internal/app/model.go": {316, 349, 377, 485, 502, 567, 631, 724, 761, 846, 924, 1005, 1089, 1110, 1250, 1333, 1828},
}

// TestBodyDimsMigration enforces UI-18: the banned subtraction pattern
// (the expression subtracting the status-bar helper result from the model
// height) must not appear outside bodyDims or the Plan-2 allowlist.
//
// SELF-MATCH AVOIDANCE: The banned regex is assembled at runtime from THREE
// separate string constants below so this file does not contain the full
// contiguous pattern as one literal — in code, doc comments, or error text.
// Any paraphrased references in comments use wording like "the banned
// subtraction pattern" or "UI-18 banned expression".
//
// LIVE FROM PLAN 1: This test runs under default go test and passes because
// the 17 legitimate pre-migration occurrences are allowlisted. Plan 2's
// atomic migration commit deletes plan2MigrationAllowlist and this filter.
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
			// Plan-2 allowlist: the 17 enumerated pre-migration sites.
			// Deleted by Plan 2 atomically with the migration.
			if allowed, ok := plan2MigrationAllowlist[rel]; ok {
				hit := false
				for _, al := range allowed {
					if al == lineNo {
						hit = true
						break
					}
				}
				if hit {
					continue
				}
			}
			violations = append(violations, rel+":"+strconv.Itoa(lineNo)+"  "+strings.TrimSpace(line))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk failed: %v", err)
	}
	if len(violations) > 0 {
		t.Fatalf("UI-18 violation: banned subtraction pattern found outside bodyDims and outside the Plan-2 allowlist:\n  %s\n\n"+
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
