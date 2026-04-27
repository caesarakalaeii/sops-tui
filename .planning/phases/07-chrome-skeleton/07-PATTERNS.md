# Phase 7: Chrome Skeleton - Pattern Map

**Mapped:** 2026-04-24
**Files analyzed:** 22 (10 new, 12 modified/extended)
**Analogs found:** 22 / 22 (strong matches for every file; sops-tui already has 6+ phases of UI primitives, sub-models, and test harness patterns to copy from)

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/keys/hints.go` | keys / contract | transform (bindings → hints) | `internal/keys/bindings.go` | exact (struct + helper funcs in same package) |
| `internal/keys/hints_test.go` | test | transform | `internal/keys/bindings_test.go` | exact (same test package convention) |
| `internal/ui/logo.go` | ui renderer (pure) | request-response | `internal/ui/styles.go` + block-ASCII pattern from research §5 | role-match (styles file holds pkg-var art; logo is pure render-from-const) |
| `internal/ui/logo_test.go` | test | request-response | `internal/ui/filelist_test.go` | role-match (plain assertion-style unit tests on a pure renderer) |
| `internal/ui/menu.go` | ui renderer (pure) | transform (hints → grid) | `internal/ui/health.go` §buildContentLines + lipgloss/v2/table | role-match (build-and-render pattern w/o bubbletea state) |
| `internal/ui/menu_test.go` | test | transform | `internal/ui/diff_test.go` + `internal/ui/health_test.go` | role-match (t.Run subtests asserting string content) |
| `internal/ui/chrome.go` | ui renderer / composer | transform (inputs → frame) | `internal/ui/history.go` §View + bordered-box recipe | role-match (renderer composes border + content; WrapTitled/overlayTitle are pure string helpers) |
| `internal/ui/chrome_test.go` | test | transform | `internal/ui/diff_test.go` (stripAnsi helper) + `internal/app/layout_test.go` (grep-gate) | role-match |
| `internal/app/chrome_test.go` | test (grep-gate + bench) | file-scan + bench | `internal/app/layout_test.go` (TestBodyDimsMigration) + `internal/app/bench_test.go` | exact (same `findRepoRoot`, AST-walk, `testing.B.Loop()` pattern) |
| `internal/app/hints_test.go` | test | request-response | `internal/app/model_test.go` + `internal/app/layout_test.go` | exact (same `defaultEnvInternal()` + `tea.WindowSizeMsg`/state setup pattern) |
| `internal/ui/styles.go` (MODIFIED — 8 vars added) | design-system config | static declaration | existing `var ( … )` block at `styles.go:75-213` | exact (append to same block) |
| `internal/app/model.go` (MODIFIED — View/chromeHeight/menuHints/titleForState) | controller / root model | request-response | own file at `model.go:1284-1326` (View), `model.go:1415-1418` (chromeHeight stub), `model.go:1394-1397` (statusBarHeight), `model.go:1403-1410` (bodyDims) | exact (Phase 6 literally just touched these helpers; D-19 migration mirrors Phase 6 plan-2 pattern) |
| `internal/ui/filelist.go` (+Hints()) | sub-model | request-response | self — Hints is one-method addition next to existing `IsSearchActive()`, `ItemCount()` | exact |
| `internal/ui/detail.go` (+Hints()) | sub-model | request-response | self + `keys.DefaultDetailKeyMap.ShortHelp()` at `bindings.go:192-194` | exact |
| `internal/ui/help.go` (+Hints()) | sub-model | request-response | self — minimal method on `HelpModel` | exact |
| `internal/ui/diff.go` (+Hints()) | sub-model | request-response | self + inline binding list from `diff.go:96-114` | exact |
| `internal/ui/metadata.go` (+Hints()) | sub-model | request-response | self — minimal accessor method | exact |
| `internal/ui/health.go` (+Hints(), possibly +FindingCount()) | sub-model | request-response | self — pattern mirrors `FileListModel.ItemCount()` | exact |
| `internal/ui/history.go` (+Hints(), possibly +CommitCount()) | sub-model | request-response | self — pattern mirrors `FileListModel.ItemCount()` | exact |
| `internal/ui/recipientform.go` (+Hints()) | sub-model | request-response | self — minimal accessor method on form model | exact |
| `internal/ui/{filelist,detail,help,diff,metadata,health,history,recipientform}_test.go` (+TestHints) | tests | request-response | `internal/ui/filelist_test.go:74-79` (`TestFileListItemInterfaces` shape) | exact |
| `internal/app/testdata/resize_{40x12,80x24,120x40,200x60}.golden` (REFRESHED) | golden fixtures | file-I/O | existing same files + `internal/testutil/golden.go:30-56` | exact |

## Pattern Assignments

### `internal/keys/hints.go` (keys contract, transform)

**Analog:** `internal/keys/bindings.go`

**Imports pattern** (bindings.go lines 1-17):
```go
// Package keys defines keybinding contracts for all Phase 1 views in sops-tui.
// ...
// Per D-09: global keys (?, q) appear in every context via embedding.
package keys

import (
	"charm.land/bubbles/v2/key"
)
```
New `hints.go` uses the **same package + same import** (just `charm.land/bubbles/v2/key` — nothing else needed for `HintsFromBindings`).

**Core pattern: struct + interface + factory var pattern** (mirrors bindings.go lines 21-38 `GlobalKeyMap` + `DefaultGlobalKeyMap`):
```go
// GlobalKeyMap holds keys available in every view.
type GlobalKeyMap struct { Help, Quit key.Binding }

// DefaultGlobalKeyMap is the default instance of GlobalKeyMap.
var DefaultGlobalKeyMap = GlobalKeyMap{ ... }
```
Phase 7 mirrors this doc-comment style verbatim: godoc summary line on each exported symbol, short godoc on fields, package-var `Default*` instances for the five inline hint set literals (`FileListSearchHints`, `RecipientConfirmHints`, `BulkReKeyConfirmHints`, `RecipientListHints`, `FormatMenuHints`).

**Helper function pattern** (mirrors `ShortHelp()` at `bindings.go:74-76`):
```go
func (k FileListKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, ... k.Help, k.Quit}
}
```
`HintsFromBindings(bindings []key.Binding) []MenuHint` walks each binding, calls `b.Help()` which returns `key.Help{Key, Desc}`, and builds `MenuHint{Mnemonic: h.Key, Description: h.Desc, Visible: true}`. The pattern is a pure transform — identical shape to `ShortHelp()`, just operating on an input slice.

**Public API shape** (per UI-SPEC §Chrome rendering composition functions):
```go
type MenuHint struct {
	Mnemonic    string
	Description string
	Visible     bool
}

type Hinter interface {
	Hints() []MenuHint
}

func HintsFromBindings(bindings []key.Binding) []MenuHint { ... }

// Inline hint sets for states with no owning sub-model (D-09)
var FileListSearchHints = []MenuHint{ ... }
var RecipientConfirmHints = []MenuHint{ ... }
var BulkReKeyConfirmHints = []MenuHint{ ... }
var RecipientListHints = []MenuHint{ ... }
var FormatMenuHints = []MenuHint{ ... }
```
Concrete hint copy locked by UI-SPEC §Inline hint sets table (lines 241-245).

---

### `internal/keys/hints_test.go` (test, transform)

**Analog:** `internal/keys/bindings_test.go`

**Test package + import pattern** (bindings_test.go lines 1-11):
```go
package keys_test

import (
	"fmt"
	"testing"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"github.com/caesarakalaeii/sops-tui/internal/keys"
	"github.com/stretchr/testify/assert"
)
```
Use `package keys_test` (external test) — same as bindings_test.go.

**Test-function shape** (bindings_test.go lines 67-82):
```go
func TestFileListKeyMap_ImplementsHelpKeyMap(t *testing.T) {
	var _ help.KeyMap = keys.FileListKeyMap{}
	short := keys.DefaultFileListKeyMap.ShortHelp()
	assert.NotEmpty(t, short, "ShortHelp must return at least one binding")
	...
}
```
New test cases cover:
- `HintsFromBindings` round-trips `key.Binding` → `MenuHint` with Key/Desc fields intact
- `Visible` defaults to true for converted bindings
- Empty input → empty output (not nil panic)
- Inline hint-set vars (`FileListSearchHints`, etc.) match UI-SPEC copy strings verbatim

---

### `internal/ui/logo.go` (ui renderer, request-response)

**Analog:** `internal/ui/styles.go` (for the package-var ASCII-art declaration pattern) + `internal/ui/history.go` §View (for `Render()`-on-package-var-style pattern).

**Imports pattern** (styles.go lines 1-10 — this is the lightest-weight precedent):
```go
package ui

import (
	"charm.land/lipgloss/v2"
)
```
No bubbletea/v2 import needed — logo is a pure string renderer, no Msg/Update.

**Package-var ASCII art pattern** (styles.go lines 14-32 — same `var ( … )` idiom applies to string-slice constants; see also the `Logo_small.txt` sprite in k9s referenced by SUMMARY):
```go
const (
	ColorBgHex = "#1e1e2e"
	...
)
var (
	ColorBg = lipgloss.Color(ColorBgHex)
	...
)
```
Phase 7 logo declaration (from RESEARCH §5 Candidate A):
```go
// LogoSmall is the 6-row "SOPS" block + "tui" subscript per D-01.
// Width: 25 cols. ASCII-only (no emoji, no VS16, no ZWJ) per UI-15.
var LogoSmall = []string{
	`  ____   ___  ____  ____  `,
	` / ___| / _ \|  _ \/ ___| `,
	` \___ \| | | | |_) \___ \ `,
	`  ___) | |_| |  __/ ___) |`,
	` |____/ \___/|_|   |____/ `,
	`                      tui `,
}
```

**LogoStatus enum pattern** (mirrors `ViewState` at `help.go:21-31`):
```go
// ViewState identifies which view is currently active ...
type ViewState int
const (
	ViewFileList ViewState = iota
	ViewDetail
	ViewMetadata
)
```
Phase 7 uses the identical iota-enum pattern for `LogoStatus`:
```go
type LogoStatus int
const (
	LogoInfo LogoStatus = iota
	LogoWarn
	LogoError
)
```

**Render pattern** (mirrors history.go lines 133-141 — apply a package-var style and `.Render()`):
```go
return lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	...
	Render(inner)
```
Phase 7 equivalent — read pre-declared style var, don't construct; grep-gate-safe:
```go
func RenderLogo(status LogoStatus, width int) string {
	art := strings.Join(LogoSmall, "\n")
	switch status {
	case LogoWarn:  return LogoStyleWarn.Render(art)
	case LogoError: return LogoStyleError.Render(art)
	default:        return LogoStyleInfo.Render(art)
	}
}
```
`width` param is accepted but unused in Phase 7 body (locked 25 cols); plumbed for Phase 10 per D-02.

---

### `internal/ui/logo_test.go` (test, request-response)

**Analog:** `internal/ui/filelist_test.go` (simple assertion style on pure functions)

**Test shape** (filelist_test.go lines 32-38):
```go
func TestFileListEmptyStateView(t *testing.T) {
	m := ui.NewFileListModel([]ui.FileItem{}, 80, 24)
	view := m.View()
	assert.True(t, strings.Contains(view, "No SOPS files found"),
		"empty state must contain 'No SOPS files found', got: %q", view)
}
```
Phase 7 logo tests (minimal — no state, just string assertions):
- `TestRenderLogo_SixRows` — strips ANSI, splits by `\n`, asserts len == 6
- `TestRenderLogo_ASCIIOnly` — every rune in every row is ≤ 0x7F
- `TestRenderLogo_WidthAround26` — every row's `lipgloss.Width()` in [22, 26]
- `TestRenderLogo_AllStatusVariants` — LogoInfo/Warn/Error each render without panic

Use `package ui_test`, import `github.com/caesarakalaeii/sops-tui/internal/ui`, assert via `testify/assert`.

---

### `internal/ui/menu.go` (ui renderer, transform)

**Analog:** `internal/ui/health.go` §buildContentLines (lines 83-134) for the build-up-rows-then-render pattern.

**Core pattern** — RESEARCH §2 gives the complete ready-to-paste code:
```go
package ui

import (
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
)

func RenderMenu(hints []keys.MenuHint, width int) string {
	const maxRows = 6
	const cols = 2

	// D-06: filter to visible hints, cap at 12
	visible := make([]keys.MenuHint, 0, cols*maxRows)
	for _, h := range hints {
		if h.Visible {
			visible = append(visible, h)
			if len(visible) == cols*maxRows { break }
		}
	}

	// D-07 column-major fill: rows[r][c] = hint at (col c, row r)
	rows := make([][]string, maxRows)
	for r := 0; r < maxRows; r++ { rows[r] = make([]string, cols) }
	for i, h := range visible {
		col := i / maxRows
		row := i % maxRows
		if col >= cols { break }
		rows[row][col] = MenuKeyStyle.Render("["+h.Mnemonic+"]") + " " + MenuDescStyle.Render(h.Description)
	}

	t := table.New().
		BorderTop(false).BorderBottom(false).
		BorderLeft(false).BorderRight(false).
		BorderRow(false).BorderColumn(false).
		BorderHeader(false).
		StyleFunc(func(row, col int) lipgloss.Style { return MenuCellStyle }).
		Rows(rows...).
		Width(width)
	return t.Render()
}
```

**Narrow-terminal safety pattern** (mirrors `health.go:168-175`):
```go
boxWidth := m.width - 2
if boxWidth < 1 { boxWidth = 1 }
```
At width < 4, `table.Width(width)` auto-handles degenerate cases (verified by goldens at 40×12).

---

### `internal/ui/menu_test.go` (test, transform)

**Analog:** `internal/ui/diff_test.go` (stripAnsi helper + `t.Run` subtest pattern) + `internal/ui/health_test.go` §TestHealthModel (nested t.Run structure)

**Test harness pattern** (diff_test.go lines 1-11 + health_test.go lines 14-27):
```go
func TestHealthModel(t *testing.T) {
	t.Run("renders loading state before SetResults", func(t *testing.T) {
		m := ui.NewHealthModel(80, 24)
		view := m.View()
		assert.Contains(t, view, "Running health check", ...)
	})
	t.Run("renders no issues found for empty results", func(t *testing.T) { ... })
}
```
Phase 7 menu tests (mirror):
- `TestRenderMenu_ColumnMajorFill` — 12 hints → rows[0-5] column-major order (0..5 in col 0; 6..11 in col 1)
- `TestRenderMenu_InvisibleHintsSkipped` — `Visible=false` cells render empty, hints don't shift up
- `TestRenderMenu_NarrowTerminalSafe` — width=40 and width=10 both render without panic
- `TestRenderMenu_CapsAt12Hints` — pass 20 hints, only first 12 visible appear in output
- `TestRenderMenu_ASCIIOnly` — every rune ≤ 0x7F (already implicitly guarded by grep-gate, but unit-level gate here too)

---

### `internal/ui/chrome.go` (ui composer, transform)

**Analog:** `internal/ui/history.go` §View (lines 78-141) for bordered-box construction, and **RESEARCH §1** for the `overlayTitle` algorithm (community-standard, not present in soft-serve main HEAD).

**Imports pattern** (history.go lines 11-19):
```go
package ui

import (
	"strings"

	"charm.land/lipgloss/v2"

	gitpkg "github.com/caesarakalaeii/sops-tui/internal/git"
)
```
Phase 7 `chrome.go`:
```go
package ui

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/charmbracelet/x/ansi"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
)
```
`ansi.Truncate` is already a direct dep (Phase 6); `ansi.Strip` is already used in `testutil/golden.go:23`.

**WrapTitled pattern** (RESEARCH Pattern 3 + styles.go `TitledBorderStyle` package var — mirrors history.go lines 133-140 w/ RoundedBorder SWAPPED TO NormalBorder per D-13, D-21):
```go
func WrapTitled(title, body string, width, height int) string {
	if width < 4 { width = 4 }
	if height < 3 { height = 3 }
	rendered := TitledBorderStyle.
		Width(width - 2).
		Height(height - 2).
		Render(body)
	return overlayTitle(rendered, " "+title+" ")
}
```
`TitledBorderStyle` is declared as package var in `styles.go` (never `lipgloss.NewStyle()` inside this function — grep-gated).

**overlayTitle pattern** (RESEARCH §1 closed gap; community-standard lipgloss v2 recipe — no native API):
```go
// overlayTitle injects " Title " into the top border of a rendered box.
// Splices the rune slice of the first line at column 2, preserving the
// top-left corner (col 0) and top-right corner (last col) of the border.
//
// Reference: this is a community-standard pattern for lipgloss v2, which
// has no native border-title API (confirmed 2026-04-24 against lipgloss
// v2.0.3 docs). The soft-serve pkg/ui/components/header was cited as
// a reference in sops-tui's Phase 7 research (CONTEXT.md D-14); verified
// the pattern is not in soft-serve main @ ac135366 but is the documented
// approach across bubbletea-based TUIs (gh, glow, charm examples).
func overlayTitle(rendered, title string) string {
	if title == "" { return rendered }
	nl := strings.IndexByte(rendered, '\n')
	if nl < 0 { return rendered }
	firstLine := rendered[:nl]
	rest := rendered[nl:]
	firstLineWidth := lipgloss.Width(firstLine)
	if firstLineWidth < 4 { return rendered }
	maxTitleWidth := firstLineWidth - 4
	titleW := lipgloss.Width(title)
	if titleW > maxTitleWidth {
		title = ansi.Truncate(title, maxTitleWidth, "…")
		titleW = lipgloss.Width(title)
	}
	newFirstLine := spliceRenderedLine(firstLine, 2, 2+titleW, title)
	return newFirstLine + rest
}
```
`spliceRenderedLine` is a rune-slice walk replacing `[startCol, endCol)` with replacement. Implementation trivial (≤20 LOC) since NormalBorder top line is pure ASCII-range box-drawing chars.

**RenderChrome pattern** (pure `JoinHorizontal` — mirrors Pattern 1 composition + info-panel-placeholder per Pitfall 1):
```go
func RenderChrome(hints []keys.MenuHint, logoStatus LogoStatus, width int) string {
	const infoPanelWidth = 38
	const logoWidth = 25
	infoBlank := lipgloss.NewStyle().Width(infoPanelWidth).Height(6).Render("")
	// ^^ NOTE: lipgloss.NewStyle() OK inside internal/ui; grep-gate only scans
	// internal/app/model.go View(). Package vars preferred; if a reviewer flags
	// this, replace with a pre-declared InfoPanelPlaceholderStyle.
	menuWidth := width - infoPanelWidth - logoWidth
	if menuWidth < 1 { menuWidth = 1 }
	menu := RenderMenu(hints, menuWidth)
	logo := RenderLogo(logoStatus, logoWidth)
	return lipgloss.JoinHorizontal(lipgloss.Top, infoBlank, menu, logo)
}
```
Note: the above uses `lipgloss.NewStyle()` which is **allowed** here (not in AppModel.View); see Pitfall 1 mitigation — but the cleaner path is declaring `InfoPanelPlaceholderStyle` as a package var in styles.go alongside the 8 already-specified Phase 7 additions. **Plan 2 author decides** (see Claude's Discretion below).

---

### `internal/ui/chrome_test.go` (test, transform)

**Analog:** `internal/ui/diff_test.go` (stripAnsi helper) + RESEARCH §1 test contract (7 cases)

**TestOverlayTitle_PreservesCornersAndWidth pattern** (RESEARCH §1 test contract):
```go
func TestOverlayTitle_PreservesCornersAndWidth(t *testing.T) {
	// Build a plain NormalBorder box via WrapTitled with empty title, splice it manually.
	box := TitledBorderStyle.Width(20).Height(4).Render("body")

	t.Run("preserves top-left corner", func(t *testing.T) {
		got := overlayTitle(box, " Title ")
		require.True(t, strings.HasPrefix(got, "╭"))
	})
	t.Run("preserves top-right corner", func(t *testing.T) {
		got := overlayTitle(box, " Title ")
		firstLineEnd := strings.IndexByte(got, '\n')
		require.True(t, strings.HasSuffix(got[:firstLineEnd], "╮"))
	})
	t.Run("width unchanged", func(t *testing.T) {
		got := overlayTitle(box, " Title ")
		firstLineOrig := box[:strings.IndexByte(box, '\n')]
		firstLineGot := got[:strings.IndexByte(got, '\n')]
		require.Equal(t, lipgloss.Width(firstLineOrig), lipgloss.Width(firstLineGot))
	})
	t.Run("overlong title truncated with ellipsis", func(t *testing.T) {
		got := overlayTitle(box, " an extremely long title that exceeds the width ")
		require.Contains(t, got, "…")
	})
	t.Run("empty title returns unchanged", func(t *testing.T) {
		got := overlayTitle(box, "")
		require.Equal(t, box, got)
	})
}
```

**TestWrapTitled pattern** — assert corners + title appears + padding respected (straightforward pure-function test; model on the shape in `diff_test.go:15-35`).

---

### `internal/app/chrome_test.go` (test, grep-gate + bench)

**Analog:** `internal/app/layout_test.go` (TestBodyDimsMigration + findRepoRoot + findBodyDimsRange) + `internal/app/bench_test.go` (testing.B.Loop)

**Grep-gate pattern — TestChromeASCIIOnly** (mirrors layout_test.go lines 79-133 `TestBodyDimsMigration` structure):
```go
func TestChromeASCIIOnly(t *testing.T) {
	repoRoot := findRepoRoot(t)
	// Allowlist: box-drawing from NormalBorder + ellipsis
	allowlist := map[rune]bool{'─': true, '│': true, '╭': true, '╮': true, '╰': true, '╯': true, '…': true}
	files := []string{
		"internal/ui/chrome.go",
		"internal/ui/logo.go",
		"internal/ui/menu.go",
		"internal/ui/crumbs.go", // may not exist yet; skip-if-missing
	}
	var violations []string
	for _, rel := range files {
		content, err := os.ReadFile(filepath.Join(repoRoot, rel))
		if err != nil { continue } // crumbs.go lands Phase 8
		for lineNo, line := range strings.Split(string(content), "\n") {
			for _, r := range line {
				if r > 0x7F && !allowlist[r] {
					violations = append(violations,
						rel+":"+strconv.Itoa(lineNo+1)+" non-ASCII rune U+"+
							strconv.FormatInt(int64(r), 16))
				}
			}
		}
	}
	if len(violations) > 0 {
		t.Fatalf("UI-15 violation: non-ASCII runes in chrome files:\n  %s",
			strings.Join(violations, "\n  "))
	}
}
```

**Grep-gate pattern — TestChromeNormalBorderOnly** (regex-walk, mirrors `TestBodyDimsMigration` lines 82-126):
```go
func TestChromeNormalBorderOnly(t *testing.T) {
	banned := regexp.MustCompile(
		`RoundedBorder|ThickBorder|DoubleBorder|HiddenBorder|FocusedBorder|UnfocusedBorder`)
	files := []string{"internal/ui/chrome.go", "internal/ui/logo.go", "internal/ui/menu.go"}
	// ... same walk as TestBodyDimsMigration, fail if banned matches
}
```

**AST-walk pattern — TestViewNoNewStyle** (Pitfall 2 prescribes `ast.Inspect` over `ast.Walk`):
```go
func TestViewNoNewStyle(t *testing.T) {
	repoRoot := findRepoRoot(t)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filepath.Join(repoRoot, "internal/app/model.go"), nil, 0)
	require.NoError(t, err)
	var viewBody *ast.BlockStmt
	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Name.Name != "View" || fd.Recv == nil { continue }
		// Only methods on AppModel
		if !isAppModelReceiver(fd.Recv) { continue }
		viewBody = fd.Body
		break
	}
	require.NotNil(t, viewBody, "View() method on AppModel not found")
	var violations []token.Pos
	ast.Inspect(viewBody, func(n ast.Node) bool {
		ce, ok := n.(*ast.CallExpr)
		if !ok { return true }
		se, ok := ce.Fun.(*ast.SelectorExpr)
		if !ok { return true }
		ident, ok := se.X.(*ast.Ident)
		if !ok { return true }
		if ident.Name == "lipgloss" && se.Sel.Name == "NewStyle" {
			violations = append(violations, ce.Pos())
		}
		return true
	})
	if len(violations) > 0 {
		t.Fatalf("UI-15 violation: lipgloss.NewStyle() call(s) inside View():\n  %v", violations)
	}
}
```

**Bench-budget pattern — TestBenchmarkAppView_UnderBudget** (mirrors bench_test.go:16-31):
```go
func TestBenchmarkAppView_UnderBudget(t *testing.T) {
	result := testing.Benchmark(BenchmarkAppView)
	nsPerOp := result.NsPerOp()
	if nsPerOp > 50_000 {
		t.Fatalf("BenchmarkAppView regressed: got %d ns/op, want <= 50000 (50µs)", nsPerOp)
	}
}
```
`testing.Benchmark()` runs the benchmark once from inside a test — same mechanism Phase 6 used to scaffold the baseline.

**Shared helpers** — reuse `findRepoRoot(t)` at `layout_test.go:135-153`; it's already in the `app` package (internal, not `app_test`), so TestChromeASCIIOnly/TestChromeNormalBorderOnly live in the same internal `package app` to inherit it.

---

### `internal/app/hints_test.go` (test, request-response)

**Analog:** `internal/app/model_test.go` + `internal/app/layout_test.go` (same package-internal test pattern, same `defaultEnvInternal()` setup)

**Test setup pattern** (layout_test.go lines 19-38):
```go
func defaultEnvInternal() ui.EnvStatus { ... }

func TestBodyDims(t *testing.T) {
	m := NewAppModel(defaultEnvInternal(), "")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(AppModel)
	w, h := bodyDims(m)
	...
}
```
Phase 7 hints tests (same shape) — set state via the exported `State*` constants at `model.go:76-87`:
```go
func TestMenuHints_StateFileList(t *testing.T) {
	m := NewAppModel(defaultEnvInternal(), "")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(AppModel)
	// m.state defaults to stateFileList
	hints := m.menuHints()
	require.NotEmpty(t, hints)
	// Assert expected mnemonics from DefaultFileListKeyMap.ShortHelp()
}

func TestMenuHints_FileListSearchActive(t *testing.T) {
	// Activate search, assert hints switches to FileListSearchHints
}

func TestMenuHints_RecipientActionDispatch(t *testing.T) {
	// Set state=stateRecipientConfirm, recipientAction="add"
	// Assert hints == keys.RecipientConfirmHints
}
```

Dispatcher-tuple test matrix per D-10 / UI-SPEC §Hints dispatch tuple (12 states × up-to-4 recipientActions × 2 IsSearchActive — but practically only ~13 meaningful combinations per dispatch-table in UI-SPEC lines 301-315).

---

### `internal/ui/styles.go` MODIFIED — 8 new package vars (config, static declaration)

**Analog:** existing `var ( … )` block at `styles.go:75-213`

**Append-to-block pattern** (styles.go lines 75-80):
```go
var (
	// ErrorLabel renders [ERROR] prefix text in bold red for startup error boxes.
	ErrorLabel = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorError)

	// WarnLabel ...
)
```
Phase 7 adds 8 new vars inside the same `var ( … )` block (specific definitions locked by UI-SPEC §New style declarations, lines 166-176):
```go
// Phase 7: Menu styles (D-05)
MenuKeyStyle  = lipgloss.NewStyle().Foreground(ColorAccent)
MenuDescStyle = lipgloss.NewStyle().Foreground(ColorFg)
MenuCellStyle = lipgloss.NewStyle() // reserved for Phase 10 per-column tweaks

// Phase 7: Logo severity styles (D-02) — Info used in Phase 7; Warn/Error reserved for Phase 10 (UI-03)
LogoStyleInfo  = lipgloss.NewStyle().Foreground(ColorAccent)
LogoStyleWarn  = lipgloss.NewStyle().Foreground(ColorWarning)
LogoStyleError = lipgloss.NewStyle().Foreground(ColorError)

// Phase 7: Titled border (D-12, D-13)
TitledBorderStyle = lipgloss.NewStyle().
	Border(lipgloss.NormalBorder()).
	BorderForeground(ColorMuted).
	Padding(0, 1)

// Phase 7: Title label inside border overlay
TitleLabelStyle = lipgloss.NewStyle().Foreground(ColorMuted)
```
Comment-per-var style matches existing Phase 5 block comments (e.g. `HealthWeakStyle` at line 198, with `(HLT-03, D-12)` source refs).

---

### `internal/app/model.go` MODIFIED (controller / root model)

#### chromeHeight flip (D-16)

**Existing stub** (model.go:1412-1418):
```go
// chromeHeight returns the rendered height of the header chrome in terminal rows.
// Phase 6: stub returning 0 (no chrome rendered yet).
// Phase 7: flipped to the real rendered height of the logo + menu + info panel.
func chromeHeight(m AppModel) int {
	_ = m
	return 0
}
```

**Phase 7 body** (flip pattern — mirrors `statusBarHeight` at lines 1394-1397):
```go
func statusBarHeight(m AppModel) int {
	statusBar := m.status.View(m.width)
	return lipgloss.Height(statusBar)
}
```
New chromeHeight body (same shape — render-then-measure):
```go
func chromeHeight(m AppModel) int {
	if m.width == 0 { return 0 }
	chrome := ui.RenderChrome(m.menuHints(), ui.LogoInfo, m.width)
	return lipgloss.Height(chrome)
}
```

**Safety note** — Phase 6's `TestChromeHeightReturnsZero` at `layout_test.go:51-55` will need to be removed or inverted. D-23 says Phase 6's migration test stays; the `TestChromeHeightReturnsZero` stub test is a separate file and must be deleted in Plan 3 (this is a Plan-3 task, not a guardrail gap).

#### View() rewrite (D-17)

**Existing View** (model.go:1284-1327):
```go
func (m AppModel) View() tea.View {
	statusBar := m.status.View(m.width)
	_, mainH := bodyDims(m)

	var content string
	switch m.state {
	case stateFileList: content = m.fileList.View()
	// ...
	}
	body := lipgloss.NewStyle().Height(mainH).Render(content)
	full := lipgloss.JoinVertical(lipgloss.Left, body, statusBar)
	v := tea.NewView(full)
	v.AltScreen = true
	return v
}
```
**Phase 7 rewrite** (Pattern 1 shape from RESEARCH §Pattern 1):
```go
func (m AppModel) View() tea.View {
	if m.width == 0 || m.height == 0 {
		v := tea.NewView("")
		v.AltScreen = true
		return v
	}

	// 1. Derive chrome inputs (pure)
	hints := m.menuHints()
	title := m.titleForState()

	// 2. Render sub-model body at inner dims
	w, h := bodyDims(m)
	var body string
	switch m.state {
	case stateFileList:         body = m.fileList.View()
	case stateDetail:           body = m.detail.View()
	case stateHelp:             body = m.help.View(ui.ViewState(m.prevState))
	case stateMetadata:         body = m.metadata.View()
	case stateDiff:             body = m.diff.View()
	case stateFormatMenu:       body = renderFormatMenu(m.formatMenuCursor, w, h)
	case stateHistory:          body = m.history.View()
	case stateHealth:           body = m.health.View()
	case stateRecipientForm:    body = m.recipientForm.View()
	case stateRecipientConfirm: body = m.diff.View()
	case stateRecipientList:    body = m.renderRecipientList()
	case stateBulkReKeyConfirm: body = m.diff.View()
	}

	// 3. Wrap body in titled border
	wrapped := ui.WrapTitled(title, body, w, h)

	// 4. Compose final frame
	chrome := ui.RenderChrome(hints, ui.LogoInfo, m.width)
	crumbs := "" // crumbsHeight stays 0 in Phase 7 (D-16)
	statusBar := m.status.View(m.width)

	full := lipgloss.JoinVertical(lipgloss.Left, chrome, crumbs, wrapped, statusBar)

	v := tea.NewView(full)
	v.AltScreen = true
	return v
}
```
Discipline — no `lipgloss.NewStyle()` inside this body or any helper lambda (grep-gated by `TestViewNoNewStyle`).

#### menuHints() dispatcher (D-10)

New method on AppModel — the pure dispatcher. Dispatch table locked by UI-SPEC §Hints dispatch tuple (lines 301-315):
```go
func (m AppModel) menuHints() []keys.MenuHint {
	switch m.state {
	case stateFileList:
		if m.fileList.IsSearchActive() { return keys.FileListSearchHints }
		return m.fileList.Hints()
	case stateDetail:            return m.detail.Hints()
	case stateMetadata:          return m.metadata.Hints()
	case stateDiff:              return m.diff.Hints()
	case stateRecipientConfirm:  return keys.RecipientConfirmHints
	case stateBulkReKeyConfirm:  return keys.BulkReKeyConfirmHints
	case stateHelp:              return m.help.Hints()
	case stateHistory:           return m.history.Hints()
	case stateHealth:            return m.health.Hints()
	case stateRecipientForm:     return m.recipientForm.Hints()
	case stateRecipientList:     return keys.RecipientListHints
	case stateFormatMenu:        return keys.FormatMenuHints
	}
	return nil
}
```

#### titleForState() helper (D-15)

New method on AppModel — the per-state title map (UI-SPEC §Titled border titles lines 196-207):
```go
func (m AppModel) titleForState() string {
	switch m.state {
	case stateFileList: return fmt.Sprintf("Files (%d)", m.fileList.ItemCount())
	case stateDetail:   return "Detail: " + m.currentFile.Name
	case stateMetadata: return "Metadata"
	case stateDiff, stateRecipientConfirm, stateBulkReKeyConfirm: return "Diff"
	case stateHelp:     return "Help"
	case stateHistory:  return fmt.Sprintf("History (%d)", m.history.CommitCount())
	case stateHealth:   return fmt.Sprintf("Health (%d findings)", m.health.FindingCount())
	case stateRecipientList: return fmt.Sprintf("Recipients (%d)", len(m.recipientList))
	case stateRecipientForm: return "RecipientForm"
	case stateFormatMenu:    return "Format"
	}
	return ""
}
```
(Requires adding `CommitCount()` and `FindingCount()` accessors to HistoryModel/HealthModel — see below.)

#### renderRecipientList magic -4 migration (D-19)

**Existing line 1841 with TODO** (model.go:1839-1841):
```go
// TODO(phase-7): replace magic -4 with a named modal-frame constant or
// bodyDims usage once modal chrome lands.
boxHeight := m.height - 4
```
**Phase 7 migration** — the renderer stops computing its own frame. It returns the inner body (title + lines + prompt + footer joined with `\n`), and `AppModel.View()` wraps it via `WrapTitled("Recipients (N)", body, w, h)`:
```go
// Before: renderRecipientList renders its own bordered box
// After:  renderRecipientList returns the inner body; View() wraps it
func (m AppModel) renderRecipientList() string {
	// ... build `inner` as before (title + lines + prompt + footer)
	return inner
	// NO boxWidth/boxHeight math; NO lipgloss.NewStyle().Border().Render()
}
```
The magic `m.height - 4` disappears because `WrapTitled` does the border math from the full `bodyDims(m)` envelope. Phase 6 already migrated 17 call-sites this same way (see `D-19` CONTEXT ref and recent commit `057d7b9`).

---

### Sub-model `Hints()` methods (8 files × 1 method each)

**Universal shape** (minimal method — ≤3 lines per implementation):
```go
// internal/ui/filelist.go
func (m FileListModel) Hints() []keys.MenuHint {
	return keys.HintsFromBindings(m.keys.ShortHelp())
}
```
Each sub-model accesses its embedded keymap field (named `keys` on FileListModel/DetailModel, or hard-coded bindings for models with no keymap field like HelpModel/DiffModel/HealthModel/HistoryModel/MetadataModel/RecipientFormModel).

**Detail specifics per model:**

| Model | Keymap source | Curation strategy |
|-------|---------------|-------------------|
| FileListModel | `m.keys.ShortHelp()` (10 bindings) — add `g/G` to reach 12 | Return `HintsFromBindings(m.keys.ShortHelp())` then `append` g/G MenuHints explicitly. Plan 3 author picks whether to modify `ShortHelp()` or append inline. |
| DetailModel | `m.keys.ShortHelp()` (13 bindings — over cap) | Drop 1 from visible (RESEARCH §4 recommends dropping E edit-in-editor or b blame) via `Visible=false` filter |
| HelpModel | no keymap field — hard-code `GlobalKeyMap.Help + Quit + Esc` hints (3 slots) | Return literal `[]MenuHint{...}` slice |
| DiffModel | no keymap field — hard-code 6-hint literal (y/n/Esc/j/k/q) | Return literal `[]MenuHint{...}` slice |
| MetadataModel | no keymap field — hard-code 5-hint literal (j/k/i/Esc/q) | Return literal `[]MenuHint{...}` slice |
| HealthModel | no keymap field — hard-code 5-hint literal (j/k/H/Esc/q) | Return literal |
| HistoryModel | no keymap field — hard-code 5-hint literal (j/k/b/Esc/q) | Return literal |
| RecipientFormModel | no keymap field — hard-code 2-hint literal (Enter/Esc) | Return literal |

Concrete mnemonic/description strings locked by RESEARCH §4 tables (lines 533-695) and UI-SPEC §Copywriting rules (lines 216-226).

**File-level TestHints shape** (per `*_test.go`):
```go
func TestFileListHints(t *testing.T) {
	m := ui.NewFileListModel(nil, 80, 24)
	hints := m.Hints()
	require.NotEmpty(t, hints)
	// Assert at least mnemonic "j/↓" present (sanity; full matrix lives in app/hints_test.go)
	require.True(t, containsMnemonic(hints, "j/↓"))
}
```

Add `containsMnemonic` helper (or inline loop) per test file — it's 3 lines, not worth a shared util.

---

### `internal/app/testdata/resize_{40x12,80x24,120x40,200x60}.golden` (REFRESHED)

**Analog:** same files plus `internal/testutil/golden.go:30-56` (RequireGoldenStructure + GOLDEN_UPDATE=1 regen path)

**Regen pattern** — invoke once after Plan 3 integrates:
```bash
GOLDEN_UPDATE=1 go test ./internal/app -run TestResize
```
Then `git add` the 4 refreshed `.golden` files. Pattern already used in Phase 6 Plan 2 commit `e05165d` (see recent commits).

**Verification** — the test re-runs without `GOLDEN_UPDATE` in CI and asserts byte-equality (ANSI-stripped + trailing-whitespace-normalised per `normalise()` at `testutil/golden.go:90-97`).

**Color assertions** — existing `resize_test.go` passes `nil` to `RequireGoldenColors` (scaffolded in Phase 6 at D-08). Plan 3 can optionally tighten this to assert `ColorAccentHex` (menu mnemonics) and `ColorMutedHex` (titled border) appear in the raw output — **Plan 3 author's discretion** per CONTEXT "Claude's Discretion" bullet on golden naming.

---

## Shared Patterns

### Always-import pattern for lipgloss + bubbletea

**Source:** every `internal/ui/*.go` file
**Apply to:** all new ui package files (`logo.go`, `menu.go`, `chrome.go`)
```go
package ui

import (
	"strings"
	"charm.land/lipgloss/v2"
	// bubbletea only if Update/Msg needed — logo/menu/chrome are pure render, no bubbletea
)
```
`ansi` from `github.com/charmbracelet/x/ansi` is added for `chrome.go` only (truncate). `charm.land/lipgloss/v2/table` added for `menu.go` only.

### Package-var style discipline

**Source:** `internal/ui/styles.go` lines 75-213 (entire var block)
**Apply to:** every new render function across `logo.go`, `menu.go`, `chrome.go`
- **Never** call `lipgloss.NewStyle()` inline in a render function that will be called from `AppModel.View()` path.
- **Always** declare styles as package vars in `styles.go`.
- **Single exception:** Inline `lipgloss.NewStyle().Width(w).Height(h).Render("")` for 6-row blank info-panel placeholder in `RenderChrome` is **borderline** — safer to declare an `InfoPanelPlaceholderStyle` package var. See Pitfall 1.

### Test harness — `defaultEnvInternal()` and `defaultEnv()`

**Source:** `internal/app/layout_test.go:19-25` (internal) + `internal/app/resize_test.go:16` (external, uses helper `defaultEnv()` from model_test.go)
**Apply to:** `internal/app/chrome_test.go` (internal — reuse `defaultEnvInternal`), `internal/app/hints_test.go` (internal — reuse `defaultEnvInternal`)
```go
func defaultEnvInternal() ui.EnvStatus {
	return ui.EnvStatus{
		SopsAvailable:     true,
		AgeAvailable:      true,
		SopsYamlAvailable: true,
	}
}
```

### findRepoRoot + grep-walk pattern

**Source:** `internal/app/layout_test.go:135-153`
**Apply to:** `internal/app/chrome_test.go` (TestChromeASCIIOnly, TestChromeNormalBorderOnly)
Reuse — these tests live in the same `package app` (internal test package — not `app_test`) so the helper is already in scope. Do not duplicate.

### Doc-comment style

**Source:** every existing `internal/ui/*.go` and `internal/keys/*.go` file header
**Apply to:** all new files (`hints.go`, `logo.go`, `menu.go`, `chrome.go`)
```go
// Package X provides <one-line summary>.
//
// <2-3 sentence expanded description including the key struct/function>
//
// Per <spec-ID>: <behaviour rule>.
// Per <spec-ID>: <another behaviour rule>.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package X
```
Spec IDs for Phase 7 additions: D-01..D-26, UI-01, UI-02, UI-06, UI-15.

### Golden update gate — GOLDEN_UPDATE=1 env var

**Source:** `internal/testutil/golden.go:35-43`
**Apply to:** resize_*.golden refresh in Plan 3
```bash
GOLDEN_UPDATE=1 go test ./internal/app -run TestResize
```
Never add a `-update` flag; Phase 6 D-10 explicitly rejected it.

### tea.View construction + AltScreen

**Source:** `internal/app/model.go:1322-1326`
**Apply to:** rewritten AppModel.View() in Plan 3
```go
v := tea.NewView(full)
v.AltScreen = true
return v
```
Do not use `WithAltScreen()` program option — it was removed in bubbletea v2 (CLAUDE.md migration rule).

---

## No Analog Found

None. Every Phase 7 file has a close analog in the existing codebase — this is the expected shape for a phase that extends an already-mature UI foundation (post-Phase 6). The `overlayTitle` helper is the closest thing to a no-analog case, but RESEARCH §1 closed the gap with a concrete community-standard algorithm and 7-case test contract, so Plan 2 has a direct recipe to paste.

---

## Metadata

**Analog search scope:**
- `internal/ui/` (15 files — all 8 sub-models + styles + existing renderers)
- `internal/keys/` (3 files — bindings, bindings_test, bindings_reveal_test)
- `internal/app/` (8 files — model, resize_test, layout_test, bench_test, model_test, etc.)
- `internal/testutil/` (golden.go + test)

**Files scanned:** 34 Go files
**Pattern extraction date:** 2026-04-24
**Confidence:** HIGH — every listed pattern is extracted from a live file in the repo at a specific line range, and Phase 7 additions are direct mechanical extensions (append-to-var-block, one-method-per-sub-model, dispatcher-on-sessionState) rather than new architectural patterns.

**Key pattern insights for planner:**
- The **hardest new pattern** is `overlayTitle` — Plan 2 closes it with RESEARCH §1's algorithm. Every other pattern is a trivial copy/extend.
- The **tightest coupling** is Plan 3 (chromeHeight flip + View() rewrite + menuHints dispatcher + 8 sub-model Hints() + 3 grep-gates + bench gate + 4 golden refreshes + magic-const migration) — 13 discrete concerns in one atomic commit. D-26 explicitly locks this scope together because splitting would reopen View() multiple times.
- **No new runtime dependencies** (ROADMAP v1.1 invariant). `charm.land/lipgloss/v2/table` is transitive — `go mod tidy` after Plan 1 may promote it to direct require (one-line diff) or leave it transitive.
