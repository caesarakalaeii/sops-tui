# Phase 6: Layout Groundwork — Pattern Map

**Mapped:** 2026-04-24
**Files analyzed:** 11 (9 created, 2 modified)
**Analogs found:** 9 / 11 (2 are greenfield testdata fixtures / new package with no direct analog)

---

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/app/model.go` *(modified)* | pure helper (layout arithmetic) | request-response (local compute) | `internal/app/model.go:1442-1446` (existing `statusBarHeight`) | exact (same file, same idiom) |
| `internal/app/model.go` *(migration)* | controller (WindowSize fan-out + on-demand sub-model creation) | event-driven (resize + message handlers) | `internal/app/model.go:313-329` (current `WindowSizeMsg` block) | exact (identical block is the rewrite target) |
| `go.mod` *(modified)* | config (dependency manifest) | n/a | current `go.mod` `require ( ... // indirect )` block | exact (one-line move) |
| `internal/app/layout_test.go` | test (unit + meta-test) | request-response (pure compute) | `internal/app/model_test.go` | exact (same package-test layout, testify style, `defaultEnv()` helper) |
| `internal/app/bench_test.go` | test (benchmark) | request-response | *none in repo* → closest: `internal/app/model_test.go` for `NewAppModel` + `WindowSizeMsg` seed | partial (analog covers the setup, `b.Loop()` idiom is new) |
| `internal/app/resize_test.go` | test (golden snapshot) | request-response | `internal/app/model_test.go::TestAppModelWindowSizePropagation` (line 108-114) | role-match (same setup; new: golden assertion) |
| `internal/app/testdata/resize_40x12.golden` | fixture (ANSI-stripped snapshot) | n/a (text fixture) | *none* (no testdata exists in repo today) | new (Go convention only) |
| `internal/app/testdata/resize_80x24.golden` | fixture | n/a | *none* | new |
| `internal/app/testdata/resize_120x40.golden` | fixture | n/a | *none* | new |
| `internal/app/testdata/resize_200x60.golden` | fixture | n/a | *none* | new |
| `internal/testutil/golden.go` | utility (test harness library) | file-I/O (env-gated read/write) | *none* (new package) → closest: `internal/ui/statusbar.go` package doc style | role-match (package conventions only) |
| `internal/testutil/golden_test.go` | test (self-tests for golden harness) | request-response | `internal/ui/statusbar_test.go` | role-match (same testify style, package_test naming) |

---

## Pattern Assignments

### `internal/app/model.go` — add `bodyDims`, `chromeHeight`, `crumbsHeight` helpers

**Analog:** `internal/app/model.go:1442-1446` (the existing `statusBarHeight` helper — same file, same naming, same call-site shape)

**Location pattern** (model.go:1442-1446 — verbatim):
```go
// statusBarHeight returns the rendered height of the status bar in terminal rows.
func statusBarHeight(m AppModel) int {
	statusBar := m.status.View(m.width)
	return lipgloss.Height(statusBar)
}
```

**Key conventions to mirror exactly:**
1. Free function (NOT a method) — `func name(m AppModel) return`. No pointer receiver. D-01 locks this.
2. Parameter is `m AppModel` (value, not pointer).
3. Single-line doc comment on the preceding line starting with the function name.
4. Placed near end of `model.go` (after all `AppModel` methods).
5. No package-level variables, no receiver methods — pure functions.

**New helpers to add immediately after line 1446** (per D-02, D-03, D-04 — verbatim from RESEARCH.md Pattern 1):

```go
// bodyDims returns the width and height available for the body region —
// the content area after subtracting the chrome (Phase 7), crumb row (Phase 8),
// and status bar. Clamped to >= 0 so bubbles/v2/list does not receive a negative
// height on terminals shorter than the chrome.
func bodyDims(m AppModel) (w, h int) {
	w = m.width
	h = m.height - statusBarHeight(m) - chromeHeight(m) - crumbsHeight(m)
	if h < 0 {
		h = 0
	}
	return w, h
}

// chromeHeight returns the rendered height of the header chrome in terminal rows.
// Phase 6: stub returning 0 (no chrome rendered yet).
// Phase 7: flipped to the real rendered height of the logo + menu + info panel.
func chromeHeight(m AppModel) int {
	_ = m
	return 0
}

// crumbsHeight returns the rendered height of the breadcrumb chip row.
// Phase 6: stub returning 0 (breadcrumb still lives in the status bar).
// Phase 8: flipped to the real rendered height of the chip pill row.
func crumbsHeight(m AppModel) int {
	_ = m
	return 0
}
```

The `_ = m` idiom is load-bearing: it keeps the signature `AppModel`-taking today so Phase 7 flips the body without touching a single call-site.

---

### `internal/app/model.go` — migrate 17 call-sites to `bodyDims`

**Analog:** `internal/app/model.go:313-329` (the current `WindowSizeMsg` fan-out — the canonical "before" pattern appears 15 times in this file)

**Before pattern** (verbatim from model.go:313-329):
```go
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		mainH := m.height - statusBarHeight(m)
		if mainH < 0 {
			mainH = 0
		}
		// Propagate to all children that need dimensions
		m.fileList.SetSize(m.width, mainH)
		m.detail.SetSize(m.width, mainH)
		m.help.SetSize(m.width, mainH)
		m.metadata.SetSize(m.width, mainH)
		m.diff.SetSize(m.width, mainH)
		m.history.SetSize(m.width, mainH)
		m.health.SetSize(m.width, mainH)
		m.recipientForm.SetSize(m.width, mainH)
		return m, nil
```

**After pattern** (the rewrite for Plan 2 — apply verbatim to all 15 sites at lines 316, 349, 377, 485, 502, 567, 631, 724, 761, 846, 924, 1005, 1089, 1110, 1250):
```go
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		w, h := bodyDims(m)
		// Propagate to all children that need dimensions
		m.fileList.SetSize(w, h)
		m.detail.SetSize(w, h)
		m.help.SetSize(w, h)
		m.metadata.SetSize(w, h)
		m.diff.SetSize(w, h)
		m.history.SetSize(w, h)
		m.health.SetSize(w, h)
		m.recipientForm.SetSize(w, h)
		return m, nil
```

**Template for on-demand creation sites** (the 14 sites where a sub-model is *constructed* rather than resized — e.g., line 349):

*Before (model.go:349-353):*
```go
		mainH := m.height - statusBarHeight(m)
		if mainH < 0 {
			mainH = 0
		}
		m.fileList = ui.NewFileListModel(items, m.width, mainH)
```

*After:*
```go
		w, h := bodyDims(m)
		m.fileList = ui.NewFileListModel(items, w, h)
```

Net removal per site: the `mainH :=`, the `if`, the `mainH = 0` (3 lines) plus `m.width, mainH` → `w, h` token substitution.

**Outlier 1 — `model.go:1329-1336` inside `View()`** (verbatim before):
```go
func (m AppModel) View() tea.View {
	// Render main content based on active state
	statusBar := m.status.View(m.width)
	statusBarH := lipgloss.Height(statusBar)
	mainH := m.height - statusBarH
	if mainH < 0 {
		mainH = 0
	}
```

*After (cosmetic rewrite so chrome/crumbs subtraction flows through `bodyDims`):*
```go
func (m AppModel) View() tea.View {
	// Render main content based on active state
	statusBar := m.status.View(m.width)
	_, mainH := bodyDims(m)
```

Note: the banned regex does NOT match line 1333 today (it uses `statusBarH` local, not `statusBarHeight(m)`). Migration is cosmetic-consistency per RESEARCH.md §"Outlier 1".

**Outlier 2 — `model.go:1799` pointer receiver** (verbatim before):
```go
	mainH := m.height - statusBarHeight(*m)
	if mainH < 0 {
		mainH = 0
	}
	m.diff = ui.NewDiffModel(fmt.Sprintf("Confirm Re-key: %s", file.Name), entries, m.width, mainH)
```

*After (dereference at call site — Option A per RESEARCH.md):*
```go
	w, h := bodyDims(*m)
	m.diff = ui.NewDiffModel(fmt.Sprintf("Confirm Re-key: %s", file.Name), entries, w, h)
```

**Outlier 3 — `model.go:1862` DEFERRED** (verbatim before):
```go
	boxHeight := m.height - 4
	if boxHeight < 1 {
		boxHeight = 1
	}
```

*After (insert TODO only, NO functional change):*
```go
	// TODO(phase-7): replace magic -4 with a named modal-frame constant or
	// bodyDims usage once modal chrome lands.
	boxHeight := m.height - 4
	if boxHeight < 1 {
		boxHeight = 1
	}
```

---

### `go.mod` — promote `charmbracelet/x/ansi` from indirect to direct

**Analog:** `go.mod:5-16` (existing direct `require` block) and `go.mod:26` (current indirect declaration)

**Current state** (go.mod:26 — indirect):
```
	github.com/charmbracelet/x/ansi v0.11.7 // indirect
```

**Target state:** Line moves into the top `require ( ... )` block alphabetically between `github.com/atotto/clipboard v0.1.4` (line 9) and `github.com/charmbracelet/x/term v0.2.2` (line 10). The `// indirect` comment is removed.

**Mechanism** (run from repo root):
```bash
go get github.com/charmbracelet/x/ansi@v0.11.7
go mod tidy
```

Expected diff: the line relocates; no `go.sum` churn (version already resolved). Verify with `git diff go.mod go.sum`.

---

### `internal/app/layout_test.go` (NEW — unit tests + grep-gate)

**Analog:** `internal/app/model_test.go` (same package, same external-test convention, same testify imports)

**Imports pattern** (copy verbatim from model_test.go:1-14):
```go
package app_test

import (
	"fmt"
	"os"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/caesarakalaeii/sops-tui/internal/app"
	"github.com/caesarakalaeii/sops-tui/internal/sops"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)
```

**Note:** The grep-gate test reads `.go` files from the repo. If the helpers under test (`bodyDims`, `chromeHeight`, `crumbsHeight`) are lowercase (per the existing `statusBarHeight` convention at model.go:1443), the unit tests must live in package `app` (not `app_test`) OR expose `BodyDimsForTest(m AppModel)` test-exports.

The existing file `model_test.go` uses `package app_test` and exercises behaviour through public API (`app.NewAppModel`, `app.FilesDiscoveredMsg`). For the helper unit tests, two options:

**Option A (recommended):** Put `layout_test.go` in `package app` (internal test), following the `ParsedFileForTest` precedent at `model.go:207-209` which is a public test helper that lives in production code. Internal-test syntax:

```go
package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
	"github.com/stretchr/testify/assert"
)

// defaultEnvInternal mirrors the external defaultEnv() helper in model_test.go:16-22.
// Duplicated deliberately because model_test.go is package app_test.
func defaultEnvInternal() ui.EnvStatus {
	return ui.EnvStatus{
		SopsAvailable:     true,
		AgeAvailable:      true,
		SopsYamlAvailable: true,
	}
}

func TestBodyDims(t *testing.T) {
	m := NewAppModel(defaultEnvInternal(), "")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(AppModel)
	w, h := bodyDims(m)
	assert.Equal(t, 80, w, "width must equal m.width")
	assert.Equal(t, 24-statusBarHeight(m), h, "height must be m.height - statusBarHeight - chromeHeight(0) - crumbsHeight(0)")
}

func TestBodyDimsClampsAtZero(t *testing.T) {
	m := NewAppModel(defaultEnvInternal(), "")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 1})
	m = updated.(AppModel)
	_, h := bodyDims(m)
	assert.GreaterOrEqual(t, h, 0, "bodyDims must clamp h to >= 0 when terminal is shorter than chrome")
}

func TestChromeHeightReturnsZero(t *testing.T) {
	m := NewAppModel(defaultEnvInternal(), "")
	assert.Equal(t, 0, chromeHeight(m), "Phase 6: chromeHeight stub returns 0")
}

func TestCrumbsHeightReturnsZero(t *testing.T) {
	m := NewAppModel(defaultEnvInternal(), "")
	assert.Equal(t, 0, crumbsHeight(m), "Phase 6: crumbsHeight stub returns 0")
}
```

**`TestBodyDimsMigration` structure** (see RESEARCH.md §Pattern 2 for the full sketch — the grep-gate test body). Critical detail from Pitfall 1: split the regex literal so the test source does NOT self-match:

```go
// Deliberate split — the test file must not contain the full banned
// pattern as one contiguous literal, or TestBodyDimsMigration would
// match itself.
banned := regexp.MustCompile(`m\.height\s*-\s*` + `statusBarHeight`)
```

---

### `internal/app/bench_test.go` (NEW — `BenchmarkAppView`)

**Analog:** `internal/app/model_test.go:108-114` (`TestAppModelWindowSizePropagation`) for the `NewAppModel` + `WindowSizeMsg` setup sequence. No existing benchmark in the repo — this is the first.

**Template** (verbatim from RESEARCH.md Pattern 4):
```go
package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

// BenchmarkAppView records the v1.0-baseline render cost so Phase 7's chrome
// skeleton has a concrete "before" number to compare its <= 50 us/op target
// against. No gating on absolute value in Phase 6.
func BenchmarkAppView(b *testing.B) {
	env := ui.EnvStatus{
		SopsAvailable:     true,
		AgeAvailable:      true,
		SopsYamlAvailable: true,
		GitAvailable:      true,
	}
	m := NewAppModel(env, "")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 60})
	m = updated.(AppModel)

	b.ReportAllocs()
	for b.Loop() {
		_ = m.View()
	}
}
```

**Key idioms:**
- `b.Loop()` (Go 1.24+, stable in 1.26) — NOT `for i := 0; i < b.N; i++`.
- `b.ReportAllocs()` before the loop.
- Fixed 200×60 (D-12 locks this — no sub-benchmarks).
- `package app` (internal) so `NewAppModel` is accessible directly (same rationale as `layout_test.go`).

Run with: `go test -bench=BenchmarkAppView -benchmem ./internal/app/...`

---

### `internal/app/resize_test.go` (NEW — 4 golden snapshot tests)

**Analog:** `internal/app/model_test.go:108-114` (`TestAppModelWindowSizePropagation`) — same `NewAppModel` + `WindowSizeMsg` setup, but replaces `assert.NotEmpty(t, v.Content)` with a golden-file comparison.

**Setup pattern** (verbatim from model_test.go:108-114):
```go
func TestAppModelWindowSizePropagation(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	m2 := send(t, m, tea.WindowSizeMsg{Width: 120, Height: 40})
	v := m2.View()
	assert.NotEmpty(t, v.Content, "View must not be empty after WindowSizeMsg")
}
```

**New test template** (per RESEARCH.md §Golden Fixture Design):
```go
package app_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/caesarakalaeii/sops-tui/internal/app"
	"github.com/caesarakalaeii/sops-tui/internal/testutil"
)

func TestResize_40x12(t *testing.T) {
	m := app.NewAppModel(defaultEnv(), "")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 12})
	m = updated.(app.AppModel)

	v := m.View()
	testutil.RequireGoldenStructure(t, "resize_40x12", v.Content)
	testutil.RequireGoldenColors(t, "resize_40x12", v.Content, nil) // Phase 6: no color asserts
}

// ... same shape for 80x24, 120x40, 200x60
```

**Conventions:**
- Reuse the existing `defaultEnv()` helper from `model_test.go:16-22` (same `package app_test`).
- Reuse the existing `send` helper pattern from `model_test.go:26-30` OR inline the update (above is inlined; pick one style and apply to all four tests).
- Golden file name: `resize_<WxH>` (no extension — the helper appends `.golden`).
- Empty `wantColors` slice is deliberate per D-09 — asserts the scaffolding is in place for Phase 7.

---

### `internal/app/testdata/resize_<WxH>.golden` (4 NEW fixture files)

**Analog:** None in repo. Go convention for `testdata/` directories (ignored by `go build`).

**Generation mechanism:** Run `GOLDEN_UPDATE=1 go test ./internal/app/... -run TestResize` once to create; visually inspect each file; commit.

**Content discipline** (per RESEARCH.md §Determinism audit + §Pitfall 3):
- No trailing whitespace per line (the `normalise()` helper in `golden.go` trims before write).
- LF line endings only (normalise forces `\n`).
- Contains plain text only (ANSI already stripped at write time).
- UTF-8 encoding.

**Optional but recommended:** Add a `.gitattributes` entry to prevent whitespace mangling on other checkouts:
```
*.golden text eol=lf -whitespace
```

---

### `internal/testutil/golden.go` (NEW — golden harness library)

**Analog:** None in repo (new package). Package conventions mirrored from `internal/ui/statusbar.go:1-17` (package doc comment style).

**Package doc comment pattern** (verbatim structure from `internal/ui/statusbar.go:1-17`):
```go
// Package testutil provides test harness helpers shared across sops-tui tests.
//
// RequireGoldenStructure performs ANSI-stripped structural comparison against
// testdata/<name>.golden fixtures. RequireGoldenColors asserts raw ANSI byte
// presence separately so structural goldens stay stable across lipgloss bumps.
//
// Per Phase 6 D-08: no external golden library (goldie, teatest.RequireEqualOutput)
// is used; x/ansi.Strip + string compare covers the need in ~30 LOC.
// Per Phase 6 D-10: regeneration is gated on GOLDEN_UPDATE=1 env var rather than
// a -update flag, to avoid reflexive fixture churn.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package testutil
```

Follow `statusbar.go:1-17` conventions exactly:
- Package comment starts with `// Package <name>`.
- Blank comment line between paragraphs (`//\n`).
- Closes with the "Never use type any" reminder from CLAUDE.md (this is repo-wide convention — present in both `model.go:15` and `statusbar.go:16`).

**Imports pattern** (following the grouping style in `model.go:18-41` — stdlib, blank line, third-party, blank line, internal):
```go
package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)
```

**Full implementation** (verbatim from RESEARCH.md Pattern 3 — keep as-is):
```go
// RequireGoldenStructure asserts that ansi.Strip(output) equals the content of
// testdata/<name>.golden. When GOLDEN_UPDATE=1 is set, the file is (re)written
// instead of compared — intentional friction per D-10 so developers have to
// opt in to regeneration.
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

// RequireGoldenColors asserts that each wantColors substring appears in the raw
// (non-stripped) output. Empty wantColors is a no-op — Phase 6 scaffolding.
func RequireGoldenColors(t *testing.T, name, output string, wantColors []string) {
	t.Helper()
	for _, c := range wantColors {
		if !strings.Contains(output, c) {
			t.Errorf("%s: expected color sequence %q not present in output", name, c)
		}
	}
}

// normalise strips trailing whitespace per line and normalises line endings.
// Prevents golden drift from editor auto-format rules (Pitfall 8 warning).
func normalise(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}
```

**Parameter naming:** `name string` (not `goldenName` or `fixture`) — matches Go stdlib `testing` conventions.

**Signature discipline (CLAUDE.md constraint):** No `any` types. `wantColors []string` is correctly typed. `*testing.T` + `string` + `[]string` + `error` throughout.

---

### `internal/testutil/golden_test.go` (NEW — harness self-tests)

**Analog:** `internal/ui/statusbar_test.go:1-20` (same `_test` package pattern, same testify imports, same `package_test` naming).

**Imports pattern** (mirror from `statusbar_test.go:1-12`):
```go
package testutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/caesarakalaeii/sops-tui/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)
```

**Test conventions to mirror** (from `model_test.go` + `statusbar_test.go`):
- Each test: `func TestX(t *testing.T) { ... }`.
- Doc comment above each test describing the invariant (`// TestX verifies that ...`).
- `require` for preconditions (file creation, expected no-error setup).
- `assert` for value checks (content match, function return).
- `t.TempDir()` for isolated fixture directories (avoids polluting `testdata/`).
- `t.Setenv("GOLDEN_UPDATE", "1")` for env-gated branch coverage (auto-reverts at test end).

**Minimum tests** (per RESEARCH.md §Validation Architecture):
- `TestRequireGoldenStructure_WritesOnUpdateEnv` — `GOLDEN_UPDATE=1` creates the file.
- `TestRequireGoldenStructure_ComparesWhenUnset` — default mode compares against existing file.
- `TestRequireGoldenStructure_ANSIStrip` — input containing ANSI escapes is stripped before write/compare.
- `TestRequireGoldenColors_Empty` — empty `wantColors` is a no-op (doesn't fail).
- `TestRequireGoldenColors_Missing` — missing color sequence fails via `t.Errorf` (test using sub-test + `testing.T.Failed()` assertion).
- `TestNormalise` — trailing whitespace trimmed, `\r\n` → `\n`.

---

## Shared Patterns

### Test Package Naming

**Source:** `internal/app/model_test.go:1`, `internal/ui/statusbar_test.go:1`

**Apply to:** `resize_test.go`, `golden_test.go` (external tests via public API) → use `package app_test` / `package testutil_test`.

**Apply inverse to:** `layout_test.go`, `bench_test.go` (need access to unexported `bodyDims`, `chromeHeight`, `crumbsHeight`, `NewAppModel` bypass patterns) → use `package app` (internal tests).

Internal-test precedent exists at `model.go:207-209` (`ParsedFileForTest` exposes internals for cross-package tests). The pattern for Phase 6 is:
- Helpers that need direct access to unexported identifiers → `package app` test file.
- End-to-end tests that exercise the public API → `package app_test`.

### Testify Usage

**Source:** `internal/app/model_test.go:12-13` and throughout

```go
import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)
```

**Apply to:** All new test files.

**Rules** (observed throughout `model_test.go`):
- `require.NoError(t, err)` / `require.NotNil(t, cmd)` — preconditions that halt the test on failure.
- `assert.Equal`, `assert.Contains`, `assert.NotEmpty` — value assertions that accumulate failures but continue.
- Always pass a descriptive message as the last argument with the actual value (`got: %q`).

Example from `model_test.go:37`:
```go
assert.NotEmpty(t, v.Content, "Initial View().Content must not be empty")
```

### `defaultEnv()` Helper

**Source:** `internal/app/model_test.go:16-22` AND `internal/ui/statusbar_test.go:14-20` (duplicated — one per test package)

```go
func defaultEnv() ui.EnvStatus {
	return ui.EnvStatus{
		SopsAvailable:     true,
		AgeAvailable:      true,
		SopsYamlAvailable: true,
	}
}
```

**Apply to:** All new tests that construct `AppModel`. Reuse the existing helper in `model_test.go` (for `app_test` package) or duplicate in `package app` internal-test files (for `layout_test.go`, `bench_test.go`).

Note: `GitAvailable` is NOT set in the current `defaultEnv()` — the benchmark explicitly adds it (RESEARCH.md Pattern 4) to exercise the full status bar rendering path at the 200×60 baseline.

### `tea.KeyPressMsg{Code: …}` API (Bubbletea v2 migration)

**Source:** `internal/app/model_test.go:45-53`

```go
m2 := send(t, m, tea.KeyPressMsg{Code: '?'})
m3 := send(t, m2, tea.KeyPressMsg{Code: tea.KeyEsc})
```

**Apply to:** Any test dispatching key events in `resize_test.go` (none currently planned — resize tests only fire `WindowSizeMsg`).

**NOT** `tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}` — that was v1.

### `send()` Helper

**Source:** `internal/app/model_test.go:26-30`

```go
func send(t *testing.T, m tea.Model, msg tea.Msg) tea.Model {
	t.Helper()
	updated, _ := m.Update(msg)
	return updated
}
```

**Apply to:** `resize_test.go` if the tests grow to multiple `Update` calls. For the current single-message-per-test shape, inline `updated, _ := m.Update(msg); m = updated.(app.AppModel)` is acceptable (see `model_test.go:62` for inline precedent).

### Package Doc Comment Format

**Source:** `internal/app/model.go:1-16`, `internal/ui/statusbar.go:1-17`

Conventions:
- Starts with `// Package <name> provides ...`.
- Multiple paragraphs separated by blank `//` lines.
- References to decision docs (`Per D-10:`, `Per RESEARCH.md Pattern X:`).
- Closes with `// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).` on any new file that does non-trivial work (repo-wide reminder).

**Apply to:** `internal/testutil/golden.go` (mandatory — new package).
**NOT applied to:** Test files (existing test files have no package doc comment; following that convention).

### License Header

**Source:** No file in the repo carries a per-file license header. The AGPL-3.0 license lives in the top-level `LICENSE` file.

**Apply to:** New files MUST NOT carry a per-file license header block. This matches the entire existing codebase.

### `_test.go` File Placement

**Source:** Observed across `internal/app/*_test.go` and `internal/ui/*_test.go`

**Rule:** Test files live in the SAME directory as the code they test. No central `tests/` directory. Fixture files go in a `testdata/` subdirectory of the package owning the test (Go convention; ignored by `go build`).

**Apply to:**
- `layout_test.go`, `bench_test.go`, `resize_test.go` → `internal/app/`
- `testdata/resize_*.golden` → `internal/app/testdata/`
- `golden_test.go` → `internal/testutil/`

---

## No Analog Found

| File | Role | Data Flow | Reason | Fallback |
|------|------|-----------|--------|----------|
| `internal/testutil/golden.go` | utility package | file-I/O | First utility package in the repo; no existing cross-package test helpers | Use `internal/ui/statusbar.go` package-doc style; use Go stdlib `testing` + `os` + `filepath` idioms; body is locked verbatim in RESEARCH.md Pattern 3 |
| `internal/app/bench_test.go` | benchmark file | request-response | Repo has zero benchmarks today | Use `testing.B.Loop()` (Go 1.26 idiom — RESEARCH.md §State of the Art); setup cribbed from `model_test.go:108-114` |
| `internal/app/testdata/resize_*.golden` | ANSI-stripped text fixture | n/a | Repo has zero `testdata/` directories or golden files today | Go convention only (`testdata/` dir ignored by build); generate with `GOLDEN_UPDATE=1` + visual inspection; optional `.gitattributes` entry `*.golden text eol=lf -whitespace` |

---

## Metadata

**Analog search scope:**
- `internal/app/` (model.go, model_test.go, model_reveal_test.go, model_clipboard_test.go)
- `internal/ui/` (statusbar.go, statusbar_test.go)
- `go.mod`
- Repo root (`.gitattributes`, `LICENSE` absence at file level, no CI directory)

**Files scanned:** ~40 Go source/test files, 1 `go.mod`, 1 `LICENSE`

**Key facts verified during mapping:**
1. `statusBarHeight` at `model.go:1442-1446` is the exact idiomatic precedent (free function, `AppModel` value receiver, single-line doc, uses `lipgloss.Height`).
2. `github.com/charmbracelet/x/ansi v0.11.7` is at `go.mod:26` marked `// indirect` — one-line move.
3. No existing file uses `package testutil` — new package.
4. No `testdata/` directories anywhere in the repo.
5. No existing `BenchmarkXxx` functions in the repo.
6. No per-file license headers anywhere — LICENSE is top-level only.
7. `defaultEnv()` is duplicated between `internal/app/model_test.go:16-22` and `internal/ui/statusbar_test.go:14-20` — the repo convention is to duplicate per-package rather than import, so Phase 6's internal-test files (`layout_test.go`, `bench_test.go`) follow the same pattern.
8. `ParsedFileForTest` at `model.go:207-209` is the precedent for exposing unexported internals to tests — the alternative is `package app` internal test files (which is what Phase 6 uses for helper unit tests).

**Pattern extraction date:** 2026-04-24
