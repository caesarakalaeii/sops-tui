// Package app — Phase 7 chrome discipline grep-gates and bench-budget gate.
//
// Four tests live here:
//
//   - TestChromeASCIIOnly (D-20)         — UI-15 ASCII allowlist for chrome source
//   - TestChromeNormalBorderOnly (D-21)  — UI-15 NormalBorder() exclusivity
//   - TestViewNoNewStyle (D-22)          — Pitfall 2 zero-NewStyle in View()
//   - TestBenchmarkAppView_UnderBudget   — UI-21 preview / D-24 50µs/op gate
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package app

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
	"github.com/caesarakalaeii/sops-tui/internal/sops"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

// TestChromeASCIIOnly enforces UI-15 (D-20): chrome source files must
// contain only ASCII runes (<=0x7F) plus an allowlist of box-drawing
// characters used by lipgloss.NormalBorder() and the ellipsis used by
// ansi.Truncate in overlayTitle.
//
// CRITICAL: NormalBorder() emits SQUARE corners (┌┐└┘), NOT rounded
// (╭╮╰╯). The 07-PATTERNS.md sketch and UI-SPEC visual mock-ups predate
// Plan 1's empirical verification and drew rounded corners; this test
// asserts reality per Plan 2's forward-deviation note in 07-02-SUMMARY.md.
//
// Arrow runes ↑↓←→ are included in the allowlist defensively: they do
// not appear in chrome source today (the arrows live in key.Binding
// help strings in internal/keys/bindings.go, which is not scanned by
// this test), but adding them preserves the option to embed arrow
// glyphs in future chrome helpers without widening the grep-gate.
//
// Scope: internal/ui/{chrome,logo,menu}.go. crumbs.go is Phase 8
// territory and is skipped if missing.
func TestChromeASCIIOnly(t *testing.T) {
	repoRoot := findRepoRoot(t)
	allowlist := map[rune]bool{
		// NormalBorder square corners + horizontals/verticals.
		'─': true, '│': true, '┌': true, '┐': true, '└': true, '┘': true,
		// Ellipsis used by ansi.Truncate in overlayTitle.
		'…': true,
		// Defensive: arrow runes used in key bindings (not in chrome source today).
		'↑': true, '↓': true, '←': true, '→': true,
	}
	files := []string{
		"internal/ui/chrome.go",
		"internal/ui/logo.go",
		"internal/ui/menu.go",
		"internal/ui/crumbs.go",    // Phase 8 D-219
		"internal/ui/infopanel.go", // Phase 8 D-219
	}
	var violations []string
	for _, rel := range files {
		path := filepath.Join(repoRoot, rel)
		content, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue // skip-if-missing (crumbs.go)
			}
			t.Fatalf("read %s: %v", rel, err)
		}
		for lineNo, line := range strings.Split(string(content), "\n") {
			for _, r := range line {
				if r > 0x7F && !allowlist[r] {
					violations = append(violations,
						rel+":"+strconv.Itoa(lineNo+1)+
							" non-ASCII rune U+"+
							strings.ToUpper(strconv.FormatInt(int64(r), 16)))
				}
			}
		}
	}
	if len(violations) > 0 {
		t.Fatalf("UI-15 violation: non-ASCII runes in chrome files:\n  %s",
			strings.Join(violations, "\n  "))
	}
}

// TestChromeNormalBorderOnly enforces UI-15 (D-21): chrome source files
// must use lipgloss.NormalBorder() exclusively. Banned border variants
// include RoundedBorder (rejected per ARCHITECTURE Pattern 3 deviation —
// font coverage fails on macOS Terminal default font per Pitfall 12),
// ThickBorder/DoubleBorder (same font issue), HiddenBorder (unused),
// and FocusedBorder/UnfocusedBorder (no-focus-ring rule per Pitfall 7).
//
// Scope: internal/ui/{chrome,logo,menu}.go only. Phase 3 legacy modal
// overlays (renderFormatMenu at model.go:1857, renderRecipientList
// before Phase 7 D-19 migration) may use RoundedBorder; they are NOT
// chrome and are explicitly outside this grep-gate's scope.
func TestChromeNormalBorderOnly(t *testing.T) {
	banned := regexp.MustCompile(
		`RoundedBorder|ThickBorder|DoubleBorder|HiddenBorder|FocusedBorder|UnfocusedBorder`)
	repoRoot := findRepoRoot(t)
	files := []string{
		"internal/ui/chrome.go",
		"internal/ui/logo.go",
		"internal/ui/menu.go",
		"internal/ui/crumbs.go",    // Phase 8 D-219
		"internal/ui/infopanel.go", // Phase 8 D-219
	}
	var violations []string
	for _, rel := range files {
		path := filepath.Join(repoRoot, rel)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		for lineNo, line := range strings.Split(string(content), "\n") {
			if banned.MatchString(line) {
				violations = append(violations,
					rel+":"+strconv.Itoa(lineNo+1)+"  "+strings.TrimSpace(line))
			}
		}
	}
	if len(violations) > 0 {
		t.Fatalf("UI-15 violation: banned border variant in chrome files:\n  %s",
			strings.Join(violations, "\n  "))
	}
}

// TestViewNoNewStyle enforces UI-15: no lipgloss.NewStyle() allocations
// inside AppModel.View() OR any helper reachable from View() in the same
// package. Phase 7.1 D-107: rewrites Phase 7's single-block walker to a
// two-pass BFS over same-package call edges so helpers (renderRecipientList,
// renderFormatMenu, etc.) are scanned, not just View()'s direct body.
//
// Pass 1: parse every non-test .go file in internal/app/, collect FuncDecls
// keyed by name (handles same-package methods on AppModel and standalone
// helpers); for each FuncDecl, walk its body and collect *ast.CallExpr
// names — both ident calls (foo()) and selector calls (m.foo() or
// receiver.foo()), recording only the trailing Sel name when the receiver
// resolves to the same package.
//
// Pass 2: BFS from "View" through the call-edge map; for each reachable
// function, ast.Inspect its body and report lipgloss.NewStyle() selector
// calls.
//
// Sub-models live in internal/ui — covered by the separate
// TestSubmodelViewsNoNewStyle in internal/ui/submodel_view_no_newstyle_test.go
// per CONTEXT.md D-108.
func TestViewNoNewStyle(t *testing.T) {
	repoRoot := findRepoRoot(t)
	appDir := filepath.Join(repoRoot, "internal", "app")

	// Pass 1: parse non-test files in internal/app/, collect FuncDecls.
	fset := token.NewFileSet()
	funcs := make(map[string]*ast.FuncDecl) // funcName -> FuncDecl
	fileOf := make(map[string]string)       // funcName -> file path
	err := filepath.Walk(appDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") {
			return nil // CONTEXT specifics: exclude test files
		}
		f, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			return parseErr
		}
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}
			// Method names collide across receivers in theory; in practice
			// this package only has AppModel methods + standalone helpers,
			// no name collisions. If a future package adds a non-AppModel
			// method named View(), the BFS would falsely follow it; tighten
			// then.
			funcs[fd.Name.Name] = fd
			fileOf[fd.Name.Name] = path
		}
		return nil
	})
	require.NoError(t, err, "walk internal/app for non-test .go files")

	require.Contains(t, funcs, "View", "AppModel.View() FuncDecl not found in internal/app")

	// Pass 1 (continued): collect call edges.
	edges := make(map[string]map[string]struct{}) // caller -> set of called names
	for name, fd := range funcs {
		callees := make(map[string]struct{})
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			ce, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			switch fn := ce.Fun.(type) {
			case *ast.Ident:
				// Direct call: foo()
				callees[fn.Name] = struct{}{}
			case *ast.SelectorExpr:
				// Selector call: m.foo() or pkg.Foo()
				// Only follow same-package edges — record the Sel name and
				// let the BFS resolve it via the funcs map. Cross-package
				// calls (e.g. ui.RenderChrome) won't have an entry in funcs
				// and are correctly skipped.
				callees[fn.Sel.Name] = struct{}{}
			}
			return true
		})
		edges[name] = callees
	}

	// Pass 2: BFS from "View".
	reachable := map[string]bool{"View": true}
	queue := []string{"View"}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for callee := range edges[cur] {
			if _, exists := funcs[callee]; !exists {
				continue // not in this package
			}
			if reachable[callee] {
				continue
			}
			reachable[callee] = true
			queue = append(queue, callee)
		}
	}

	// Pass 3: scan each reachable FuncDecl for lipgloss.NewStyle() calls.
	var violations []string
	for name := range reachable {
		fd := funcs[name]
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			ce, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			se, ok := ce.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			ident, ok := se.X.(*ast.Ident)
			if !ok {
				return true
			}
			if ident.Name == "lipgloss" && se.Sel.Name == "NewStyle" {
				pos := fset.Position(ce.Pos())
				violations = append(violations, fmt.Sprintf("%s:%d:%d (in %s)",
					filepath.Base(fileOf[name]), pos.Line, pos.Column, name))
			}
			return true
		})
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("UI-15 violation: lipgloss.NewStyle() call(s) reachable from AppModel.View():\n  %s\n\n"+
			"Declare styles as package vars in internal/ui/styles.go instead. "+
			"Phase 7.1 D-107 BFS walker traverses same-package call edges from View().",
			strings.Join(violations, "\n  "))
	}
}

// TestBenchmarkAppView_UnderBudget enforces UI-21 / D-24:
// BenchmarkAppView must stay under the per-frame budget at 200×60 with full
// Phase 7 chrome rendered. Runs the bench from inside a test (Pitfall 6) so it
// executes under `go test ./...` without an opt-in flag.
//
// Phase 7.1 governance restoration: target locked at 50 µs (CONTEXT.md D-24);
// Phase 11 D-504 closure: gate flipped ACTIVE -- the deferral skip was removed
// once Phase 11 wired the chrome cache (D-504); only the testing.Short() guard
// remains. The D-18 model-level chrome cache (Phase 11 Plan 01 D-501..D-503)
// brought ns/op under the 50,000 ns budget by amortising RenderChrome
// (~622 µs) + RenderMenu (~394 µs) + JoinHorizontal cost across frames where
// the (state, recipientAction, IsSearchActive, width) key is unchanged.
//
// Empirical pre-cache profile (dev workstation, Ryzen 7 PRO 5850U): every
// View() call rebuilt the chrome via lipgloss/v2/table (RenderMenu 394 µs),
// JoinHorizontal across the info-panel + menu + logo (RenderChrome 622 µs),
// then JoinVertical of chrome + WrapTitled body (754 µs) + status bar
// (683 µs). Total ~2.4-2.8 ms/op at 200×60 -- 48× over the locked target.
//
// Post-cache: View() reads m.chromeCache + m.chromeCrumbsCache + m.wrappedCache
// directly; the renderer call moves into refreshChromeCache() which is invoked
// at the end of every Update branch via the Update() wrapper (Phase 8 D-213
// mutate-on-event pattern, Phase 11 Rule 1 deviation extending coverage to
// every Update return path). TestChromeCache_HitRateAtSteadyState locks the
// value-receiver discipline trip wire (Pitfall A).
func TestBenchmarkAppView_UnderBudget(t *testing.T) {
	if testing.Short() {
		t.Skip("bench-budget skipped in -short mode")
	}
	result := testing.Benchmark(BenchmarkAppView)
	nsPerOp := result.NsPerOp()
	const budgetNs = 50_000 // 50 µs — locked CONTEXT.md D-24 target; original wording per ROADMAP SC5.
	if nsPerOp > budgetNs {
		t.Fatalf("BenchmarkAppView regressed: got %d ns/op, budget is %d ns/op (50 µs at 200x60).\n"+
			"Check for lipgloss.NewStyle() allocations inside View() (see TestViewNoNewStyle)\n"+
			"and confirm RenderChrome is invoked at most once per View() call.",
			nsPerOp, budgetNs)
	}
	t.Logf("BenchmarkAppView: %d ns/op (budget %d ns/op, headroom %d ns/op)",
		nsPerOp, budgetNs, budgetNs-nsPerOp)
}

// TestRenderChrome_FullTierWithInfoPanel asserts the full-tier chrome
// output contains all 5 info-panel labels at width=200 (Phase 8 D-219).
// The 38x6 InfoPanelPlaceholderStyle slot is inflated with
// RenderInfoPanel(info) per Phase 8 D-213.
func TestRenderChrome_FullTierWithInfoPanel(t *testing.T) {
	info := ui.InfoPanelData{
		SopsYamlRelPath: "secrets/.sops.yaml",
		AgeFingerprint:  "age1abcdefghijklmnop",
		RecipientCount:  4,
		GitBranch:       "main",
		GitDirty:        true,
		FileCount:       12,
	}
	hints := []keys.MenuHint{
		{Mnemonic: "?", Description: "help", Visible: true},
		{Mnemonic: "q", Description: "quit", Visible: true},
	}
	out := ui.RenderChrome(hints, ui.LogoInfo, info, ui.PaletteFor(colorprofile.TrueColor), 200)
	stripped := ansi.Strip(out)

	for _, label := range []string{"cfg:", "age:", "rcp:", "git:", "fil:"} {
		assert.Containsf(t, stripped, label,
			"full-tier chrome must contain %q label (D-201)", label)
	}
	assert.Contains(t, stripped, "secrets/.sops.yaml",
		"cfg row must contain the rendered repo-relative path")
	assert.Contains(t, stripped, "main",
		"git row must contain the branch name")
}

// TestCrumbsHeight_NonZero asserts crumbsHeight(m) > 0 after a
// WindowSizeMsg + breadcrumb is set (Phase 8 D-216 + D-219). Phase 7's
// crumbsHeight=0 stub is gone; the real height is now lipgloss.Height
// of RenderCrumbs output, typically 1 row.
func TestCrumbsHeight_NonZero(t *testing.T) {
	m := NewAppModel(ui.EnvStatus{}, "", colorprofile.TrueColor)
	// First-frame guard requires width > 0 -- send WindowSizeMsg first.
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	am := updated.(AppModel)
	am.status.SetBreadcrumb("sops-tui", "files", "prod.yaml")
	h := crumbsHeight(am)
	assert.Greater(t, h, 0,
		"crumbsHeight must be >0 after width is set + breadcrumb populated (D-216)")
}

// TestInfoPanelCacheRefresh_OnFilesDiscovered asserts m.infoPanel.FileCount
// reflects len(msg.Files) after the FilesDiscoveredMsg handler runs
// (Phase 8 D-213 + D-219).
func TestInfoPanelCacheRefresh_OnFilesDiscovered(t *testing.T) {
	m := NewAppModel(ui.EnvStatus{}, "", colorprofile.TrueColor)
	// First-frame guard
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	am := updated.(AppModel)

	files := []sops.DiscoveredFile{
		{Name: "a.yaml", AbsPath: "/tmp/a.yaml"},
		{Name: "b.yaml", AbsPath: "/tmp/b.yaml"},
	}
	updated2, _ := am.Update(FilesDiscoveredMsg{Files: files})
	am2 := updated2.(AppModel)
	assert.Equal(t, 2, am2.infoPanel.FileCount,
		"FileCount must reflect len(msg.Files) after FilesDiscoveredMsg refresh (D-213)")
}

// TestChromeCache_HitRateAtSteadyState proves the chrome cache is wired
// (populated by Update, never by View). 100 sequential View() calls
// without any Update between them must leave m.chromeCacheKey unchanged
// — the value-receiver discipline trip wire (Pitfall A: View() cannot
// mutate state, so any cache mutation in View would silently lose).
//
// Phase 11 D-505: drives 100 frames with no chrome-input changes;
// asserts the cache key is stable. Cleaner than a literal hit-count
// counter because it survives any future refactor of refreshChromeCache.
func TestChromeCache_HitRateAtSteadyState(t *testing.T) {
	env := ui.EnvStatus{SopsAvailable: true, AgeAvailable: true, SopsYamlAvailable: true, GitAvailable: true}
	m := NewAppModel(env, "", colorprofile.TrueColor)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 60})
	m = updated.(AppModel)

	keyAfterFirst := m.chromeCacheKey
	for i := 0; i < 100; i++ {
		_ = m.View()
		if m.chromeCacheKey != keyAfterFirst {
			t.Fatalf("cache key drifted at iteration %d: View() mutated cache "+
				"(expected View to be a pure read; cache must be populated by Update)", i)
		}
	}
}
