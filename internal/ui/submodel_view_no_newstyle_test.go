// Sub-model AST walker for the 8 internal/ui sub-models with View() methods.
//
// Phase 7.1 D-108: separate from internal/app's TestViewNoNewStyle BFS walker
// (chrome_test.go) because (a) different package, (b) two narrow tests are
// easier to debug than one wide one, (c) sub-model View() methods are leaf
// renderers — no BFS needed; just walk every FuncDecl in each file for
// completeness.
//
// After Plan 03 (D-110, D-112) lifted the 14 sub-model lipgloss.NewStyle()
// calls to package vars in styles.go, this walker reports zero violations.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// submodelFiles is the canonical list of sub-model View() owners +
// Phase 8 chrome primitives (infopanel.go + crumbs.go). The grep-gate
// ensures NO lipgloss.NewStyle() calls appear in any FuncDecl of these
// files, so package-var styles are the single source of truth.
var submodelFiles = []string{
	"filelist.go",
	"detail.go",
	"help.go",
	"diff.go",
	"metadata.go",
	"health.go",
	"history.go",
	"recipientform.go",
	"infopanel.go", // Phase 8 D-219
	"crumbs.go",    // Phase 8 D-219
}

// findRepoRootUI mirrors findRepoRoot from internal/app/layout_test.go.
// Walks upward from cwd looking for go.mod. Lives in the ui package so
// the sub-model walker test does not need to depend on internal/app.
func findRepoRootUI(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found walking up from cwd")
		}
		dir = parent
	}
}

// TestSubmodelViewsNoNewStyle enforces UI-15 / Phase 7.1 D-108: no
// lipgloss.NewStyle() calls in any FuncDecl of the 8 sub-model files
// (filelist, detail, help, diff, metadata, health, history, recipientform).
//
// After Plan 03 lifted the inner-border RoundedBorder chains and the
// muted footer/label/prompt styles to package vars in styles.go, this
// walker reports zero violations.
//
// Sub-models are leaf renderers (no nested helpers in same file beyond
// trivial buildContentLines), so a per-file FuncDecl scan is sufficient
// — no BFS needed. Test files are excluded by the explicit submodelFiles
// allowlist (only production .go files are listed).
//
// Companion to internal/app's TestViewNoNewStyle (Phase 7.1 D-107 BFS
// walker over the View() call graph in the app package). Together the
// two walkers cover every reachable lipgloss.NewStyle() from any View()
// in the codebase.
func TestSubmodelViewsNoNewStyle(t *testing.T) {
	repoRoot := findRepoRootUI(t)
	uiDir := filepath.Join(repoRoot, "internal", "ui")

	fset := token.NewFileSet()
	var violations []string

	for _, fname := range submodelFiles {
		path := filepath.Join(uiDir, fname)
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", fname, err)
		}
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}
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
						fname, pos.Line, pos.Column, fd.Name.Name))
				}
				return true
			})
		}
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("UI-15 / Phase 7.1 D-108 violation: lipgloss.NewStyle() call(s) in sub-model files:\n  %s\n\n"+
			"Declare styles as package vars in internal/ui/styles.go instead. "+
			"Companion to internal/app's TestViewNoNewStyle (Phase 7.1 D-107).",
			strings.Join(violations, "\n  "))
	}
}
