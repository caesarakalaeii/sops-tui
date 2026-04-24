# Phase 6: Layout Groundwork — Research

**Researched:** 2026-04-23
**Domain:** Go + Bubble Tea v2 refactor; helper introduction, grep-gated CI test, ANSI-stripped golden harness, benchmark baseline.
**Confidence:** HIGH

## Summary

Phase 6 is a pure-refactor / test-infrastructure phase. Every design question has been locked in `06-CONTEXT.md` (decisions D-01 through D-15). This research document answers the 9 planner questions raised in the phase brief: exact API shapes, current line numbers verified against `internal/app/model.go`, call-site fan-out targets, outlier diffs, library availability, and the Nyquist validation mapping.

Key verified facts from this session:

1. `github.com/charmbracelet/x/ansi v0.11.7` is **already resolved** as an indirect dependency in `go.mod:26`; `ansi.Strip(s string) string` lives at `~/go/pkg/mod/github.com/charmbracelet/x/ansi@v0.11.7/width.go:10`. Promoting to direct dep is a one-line move with no version work.
2. All **15 SetSize-adjacent call-sites** listed in CONTEXT.md D-13 were re-grepped against the current `model.go` (1911 lines) and are correct: lines **316, 349, 377, 485, 502, 567, 631, 724, 761, 846, 924, 1005, 1089, 1110, 1250**.
3. The **outlier at 1333** is inside `View()` (not `mainH := m.height - statusBarHeight(m)` — it already uses a `statusBarH := lipgloss.Height(statusBar)` local). The banned regex does NOT appear on line 1333 today; the migration there is a cosmetic rewrite to use `bodyDims()` so future chrome/crumbs subtraction flows through a single helper.
4. The **outlier at 1799** uses `m.height - statusBarHeight(*m)` (pointer deref — this is inside a `*AppModel` receiver method). The grep in step 3 confirms it matches the banned regex.
5. The **non-migrated line 1862** is `boxHeight := m.height - 4` (recipient-list modal); get-a-TODO-tag only, no migration.
6. `lipgloss v2` **removed `SetColorProfile`/`DefaultRenderer`** (per `UPGRADE_GUIDE_V2.md:216–242`). Color downsampling happens at the output layer. Tests calling `m.View()` get `view.Content` with full-fidelity ANSI; strip with `ansi.Strip` for structural compare. **No `lipgloss.NoColor` profile knob exists in v2** — CLAUDE.md still references one (legacy v1 note); for v2 we rely on `ansi.Strip` to produce a profile-independent byte stream.
7. `testing.B.Loop()` is available in Go 1.26.2 (`/usr/lib/go/src/testing/benchmark.go:502`) and is the recommended modern benchmark idiom.
8. No `internal/testutil` package exists today. No existing tests depend on golden files. No CI infrastructure exists (`.github/` absent). Zero conflicts for the new additions.
9. The empty-state `FileListModel.View()` (`filelist.go:320–338`) renders 4 deterministic lines with no time/env/cwd leakage — safe for goldens without fixture mocking.

**Primary recommendation:** The two-plan split (D-13) stands unchanged. Plan 1 is the new infrastructure (helpers, grep-gate test, testutil package, benchmark) with zero behaviour change. Plan 2 is the 17-site migration plus the four resize goldens plus the TODO comment — one atomic commit because the grep-gate test fails until every site is converted.

---

## User Constraints (from CONTEXT.md)

### Locked Decisions (D-01 through D-15)

**D-01** — `bodyDims(m AppModel) (w, h int)`: free function, returns both dimensions symmetrically.

**D-02** — Lives in `internal/app/model.go` beside `statusBarHeight` (currently at line 1443).

**D-03** — Body: `w = m.width; h = max(0, m.height - statusBarHeight(m) - chromeHeight(m) - crumbsHeight(m))`. Clamp to `>= 0`.

**D-04** — `chromeHeight(m AppModel) int` and `crumbsHeight(m AppModel) int` are stubs returning `0`. Phase 7 flips `chromeHeight`; Phase 8 flips `crumbsHeight`.

**D-05** — Grep-gate is a Go test `TestBodyDimsMigration` in `internal/app`. Zero new CI infrastructure.

**D-06** — Banned regex: `m\.height\s*-\s*statusBarHeight`. Applies to all `*.go` files. Only permitted inside the `bodyDims` function body.

**D-07** — Outliers: `model.go:1333` and `model.go:1799` are migrated to `bodyDims` in Plan 2. `model.go:1862` (`m.height - 4`) gets a `// TODO(phase-7): …` comment and is deferred.

**D-08** — Hand-rolled ANSI strip + string compare. Promote `charmbracelet/x/ansi` from indirect to direct dep.

**D-09** — New package `internal/testutil` with `golden.go`:
- `RequireGoldenStructure(t, name, output)` — ANSI-strip then compare against `testdata/<name>.golden`.
- `RequireGoldenColors(t, name, output, wantColors)` — raw-byte substring check. Empty `wantColors` in Phase 6.

**D-10** — `GOLDEN_UPDATE=1` env var for regeneration. Intentional friction.

**D-11** — Fixtures in per-package `testdata/`. Phase 6 files: `internal/app/testdata/resize_40x12.golden`, `resize_80x24.golden`, `resize_120x40.golden`, `resize_200x60.golden`.

**D-12** — `BenchmarkAppView` in `internal/app` at 200×60. ~20 LOC. No gating on absolute value.

**D-13** — Two plans:
- **Plan 1:** helpers + test infra + benchmark. No production-behaviour change.
- **Plan 2:** atomic migration of all 15 SetSize sites + 2 outliers (17 total) + resize goldens + TODO tag on 1862.

**D-14** — Migration is one commit, not per-site.

**D-15** — Plan 1 UAT: `go test ./...` green. Plan 2 UAT: 4-size goldens pass AND manual smoke at 40×12 and 200×60 confirming parity with v1.0 across every view.

### Claude's Discretion

- Exact regex compilation / line-range carve-out implementation for `TestBodyDimsMigration`.
- File walk strategy (`filepath.WalkDir` vs `go/ast` parse).
- `internal/testutil/golden.go` error message format and diff output style.
- Whether `RequireGoldenStructure` normalises trailing whitespace/line endings (Pitfall 8 recommends; implementation is Claude's call).
- `BenchmarkAppView` table size (200×60 fixed vs parametrised).
- Exact goldens file naming convention beyond `resize_<WxH>.golden`.

### Deferred Ideas (OUT OF SCOPE for Phase 6)

- **`model.go:1862` `m.height - 4` magic number** — TODO-tag only, defer to Phase 7/8 with the modal-frame chrome.
- **GitHub Actions CI bootstrap** — the grep-gate is a Go test so no CI is needed now; a future phase adds `.github/workflows/`.
- **Mode 2026 synchronized output / alt-screen fill-frame** — Pitfall 10, Phase 11.
- **`Hints() []MenuHint` interface + `HintsFromBindings` helper** — Phase 9.
- **`styles.go` stubs for chrome styles** — Phase 7.
- **golangci-lint adoption** — considered as grep-gate host; rejected for Phase 6.
- **Responsive narrow-terminal column hiding** — Phase 10.

---

## Phase Requirements

| ID | Description (from REQUIREMENTS.md) | Research Support |
|----|------------------------------------|------------------|
| **UI-17** | A `bodyDims(m) (w, h int)` helper is the single source of truth for body size arithmetic, subtracting chrome + crumbs + status-bar heights; all existing `m.height - statusBarHeight(m)` call-sites migrate to it before any chrome renders. | §Architecture Patterns §1, §Call-Site Inventory (17 concrete targets), §Stub Helper Pattern |
| **UI-18** | A CI grep-gate prevents reintroduction of the raw `m.height - statusBarHeight(m)` pattern outside the helper. | §Grep-Gate Implementation (with verified regex, walk strategy, carve-out algorithm) |
| **UI-19** | A teatest harness helper strips ANSI escape sequences for structural golden comparison and asserts color presence separately so goldens stay stable across lipgloss bumps. | §Golden-File Harness (x/ansi.Strip verified, helper signatures, normalization rules, resize-fixture design) |

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Body-dim arithmetic helper | `internal/app` (model.go) | — | Helper must be beside `statusBarHeight` (D-02); owns the AppModel receiver. |
| `chromeHeight`/`crumbsHeight` stubs | `internal/app` (model.go) | — | Co-located with the other layout helpers; Phase 7/8 flip them without moving the file. |
| Grep-gate test | `internal/app` (model_test.go or new file) | — | Same package so it can traverse module-relative file paths; no cross-package imports needed. |
| Golden harness (`RequireGoldenStructure`, `RequireGoldenColors`) | `internal/testutil` (new package) | — | Reusable by all `_test.go` files in the module; placed under `internal/` so it's not an exported public API. |
| Resize-fixture goldens | `internal/app/testdata/` | — | Per-package `testdata/` is the Go convention and is automatically ignored by `go build`. |
| `BenchmarkAppView` | `internal/app` (model_test.go or new `bench_test.go`) | — | Exercises the full `AppModel.View()` path; needs privileged access to `NewAppModel`. |

All work is local to `internal/app` and one new `internal/testutil` package. No other packages change.

---

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/charmbracelet/x/ansi` | v0.11.7 (pinned via lipgloss) | `ansi.Strip(string) string` for ANSI sequence removal in goldens | Already in `go.mod:26` as indirect. Handles CSI, OSC, DCS, Utf8State, malformed escapes — covers what a 5-line regex misses. [VERIFIED: go.mod; VERIFIED: ~/go/pkg/mod/github.com/charmbracelet/x/ansi@v0.11.7/width.go:10] |
| `charm.land/bubbletea/v2` | v2.0.4 | `tea.Model`, `tea.WindowSizeMsg`, `tea.View` struct (Content string) | Already pinned. `tea.View.Content` holds the raw styled string on line `~/go/pkg/mod/charm.land/bubbletea/v2@v2.0.4/tea.go:96`. [VERIFIED] |
| `charm.land/lipgloss/v2` | v2.0.3 | `lipgloss.Height(s)` for dimension measurement | Already pinned. The existing `statusBarHeight` helper at `model.go:1443` uses it. [VERIFIED] |
| `testing` (stdlib) | Go 1.26.2 | Unit tests, benchmarks, `t.Helper()`, `b.Loop()` | `testing.B.Loop()` confirmed available at `/usr/lib/go/src/testing/benchmark.go:502`. [VERIFIED] |
| `github.com/stretchr/testify` | v1.11.1 | `require`/`assert` | Already pinned; existing convention (`model_test.go:12–13`). [VERIFIED: go.mod:14] |

### Supporting (stdlib only — nothing new)

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `path/filepath` | stdlib | `filepath.WalkDir` for repo traversal in grep-gate | Walking `*.go` files across packages |
| `regexp` | stdlib | Compile banned regex once | Grep-gate matching |
| `os` | stdlib | `os.ReadFile`, `os.Getenv("GOLDEN_UPDATE")`, `os.WriteFile` for golden regeneration | Fixture read/write |
| `strings` | stdlib | Line-ending normalisation, trailing whitespace trim | Golden normalisation (per D-09 discretion) |
| `go/ast` / `go/parser` | stdlib | **Rejected for Phase 6** — overkill; see §Grep-Gate Implementation | — |

### Alternatives Considered

| Instead of | Could Use | Tradeoff / Why Not |
|------------|-----------|-------------------|
| Hand-rolled golden harness | `charmbracelet/x/exp/teatest` | teatest runs a full program (`tea.NewProgram`) and polls for output via `WaitFor`. Overkill for Phase 6 which only needs synchronous `m.View()` snapshot capture. D-08 locks this out. |
| Hand-rolled golden harness | `sebdah/goldie/v2` | Adds a new direct dep, uses `-update` flag that needs `TestMain` scaffolding (D-10 chose env var specifically to avoid this). |
| Regex walk for grep-gate | `go/ast` FuncDecl scan | ast is more robust but 3× the code. String scan with `func bodyDims` start + matching-brace end covers all realistic helper shapes. If Phase 7 refactors the helper into methods, revisit. |
| `filepath.WalkDir` | `go list ./... | xargs go doc` | Shell pipeline; not portable within `go test`. |
| `ansi.Strip` | Hand-rolled regex like `\x1b\[[0-9;]*m` | Misses OSC (hyperlinks, clipboard sets), DCS, malformed-escape recovery. `ansi.Strip` is 60 LOC of a full state machine — no reason to reimplement. |

**Installation (D-08):**

```bash
# Promote x/ansi from indirect to direct dep:
go get github.com/charmbracelet/x/ansi@v0.11.7
# Then clean up:
go mod tidy
```

Expected `go.mod` diff: one line moves from the `require ( ... // indirect )` block into the top `require ( ... )` block. No `go.sum` churn — version already resolved.

**Version verification (performed 2026-04-23):**

```bash
ls ~/go/pkg/mod/github.com/charmbracelet/x/ansi@v0.11.7/width.go   # exists
# go.mod line 26: github.com/charmbracelet/x/ansi v0.11.7 // indirect
```

[VERIFIED: local module cache]

---

## Architecture Patterns

### System Architecture Diagram

```
                      ┌─────────────────────────────┐
                      │  AppModel.Update(msg)       │
                      │   case tea.WindowSizeMsg:   │
                      │     m.width = msg.Width     │
                      │     m.height = msg.Height   │
                      └──────────────┬──────────────┘
                                     │ calls
                                     ▼
                      ┌─────────────────────────────┐
                      │  bodyDims(m) → (w, h)       │◀── CENTRAL HELPER (new)
                      │   w = m.width               │
                      │   h = max(0,                │
                      │       m.height              │
                      │       - statusBarHeight(m)  │
                      │       - chromeHeight(m)     │◀── stub returns 0 (Phase 6)
                      │       - crumbsHeight(m)     │◀── stub returns 0 (Phase 6)
                      │       )                     │
                      └──────────────┬──────────────┘
                                     │ fan-out to
        ┌─────────────┬──────────────┼──────────────┬──────────────┬─────────────┐
        ▼             ▼              ▼              ▼              ▼             ▼
  fileList.SetSize  detail      help   metadata  diff  history  health  recipientForm
      (8 targets at the WindowSizeMsg handler; other 9 sites create a specific sub-model)


On-demand creation sites (each computes mainH independently today):
  L349  fileList       ← FilesDiscoveredMsg
  L377  detail         ← FilesParsedMsg
  L485  diff           ← editor exit (multi-key diff)
  L502  diff           ← EditConfirmMsg (single-key diff)
  L567  diff           ← RotateReadyMsg
  L631  fileList       ← GitStatusMsg rebuild
  L724  diff           ← add-recipient confirm
  L761  diff           ← remove-recipient confirm
  L846  health         ← health-check sentinel after diff confirm
  L924  diff           ← format-menu selection
  L1005 metadata       ← i key
  L1089 history        ← git-history key
  L1110 recipientForm  ← a key (add recipient)
  L1250 diff           ← health-scan confirm prompt


Outliers:
  L1333 AppModel.View() body (uses local `statusBarH`)
  L1799 showBulkReKeyConfirm (pointer receiver — uses statusBarHeight(*m))
  L1862 renderRecipientList (m.height - 4 — DEFERRED, TODO tag only)


TESTS (new):
  internal/app/
    layout_test.go or model_test.go:
      TestBodyDimsMigration  (walks *.go, fails if banned regex appears outside bodyDims)
      TestBodyDimsClampsAtZero  (m.height=1, statusBarHeight=2 → (w, 0))
      TestChromeHeightReturnsZero  (Phase 6 sanity check)
      TestCrumbsHeightReturnsZero  (Phase 6 sanity check)
      TestResize_40x12      ─┐
      TestResize_80x24      ─┤  → call RequireGoldenStructure with
      TestResize_120x40     ─┤    testdata/resize_<WxH>.golden
      TestResize_200x60     ─┘
      BenchmarkAppView      (b.Loop() at 200×60)

  internal/testutil/
    golden.go:
      RequireGoldenStructure(t, name, output)
      RequireGoldenColors(t, name, output, wantColors)
    golden_test.go:
      round-trip sanity tests
```

### Recommended Project Structure

```
internal/
├── app/
│   ├── model.go                       # bodyDims + chromeHeight + crumbsHeight helpers added (Plan 1)
│   │                                   # 17 call-sites migrated (Plan 2)
│   ├── model_test.go                  # existing
│   ├── layout_test.go                 # NEW (Plan 1): TestBodyDims*, TestBodyDimsMigration
│   ├── bench_test.go                  # NEW (Plan 1): BenchmarkAppView
│   ├── resize_test.go                 # NEW (Plan 2): TestResize_<WxH> × 4
│   └── testdata/                      # NEW (Plan 2): resize_<WxH>.golden × 4
└── testutil/                          # NEW package (Plan 1)
    ├── golden.go                      # RequireGoldenStructure, RequireGoldenColors
    └── golden_test.go                 # Self-tests for the harness
```

Plan 1 ships `internal/testutil/`, `layout_test.go`, `bench_test.go`, and the three helpers (`bodyDims`, `chromeHeight`, `crumbsHeight`) in `model.go` — **no changes to existing production behaviour**. Plan 2 ships the 17-site edit, the 4 resize goldens, `resize_test.go`, and the TODO comment.

---

### Pattern 1: Central Body-Dim Helper with Stubs for Forward Compatibility

**What:** One function owns all body-dim arithmetic. Chrome/crumbs contributions start at zero and flip to real values in later phases without touching a single call-site.

**When to use:** Always in Phase 6. Every existing `m.height - statusBarHeight(m)` expression becomes `w, h := bodyDims(m)`.

**Example (verbatim from D-03 locked decision):**

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

Note the `_ = m` idiom: the signature takes `AppModel` today even though it's unused, so when Phase 7 flips `chromeHeight` to read from `m.chrome`, the signature does not change and no call-site moves.

### Pattern 2: Grep-Gate as a Go Test (D-05)

**What:** A self-contained `TestBodyDimsMigration` walks all `*.go` files via `filepath.WalkDir`, reads each, and scans for the banned regex. The helper's own body is exempt.

**When to use:** Always, as the sole enforcement mechanism for UI-18.

**Why Go test over golangci-lint / shell script / GitHub Actions grep:**
- Runs under `go test ./...` — same command everyone already runs locally and in any future CI.
- Failure message can be rich (`"banned pattern 'm.height - statusBarHeight' found at internal/foo/bar.go:42 — use bodyDims(m) instead"`).
- No external tool install, no CI config, no shell portability concerns.
- Future-proof: a Phase 7 contributor who adds chrome code gets a failing test if they copy-paste the old expression.

**Example:**

```go
// Source: Phase 6 D-05, D-06 — repo-internal design
package app

import (
    "os"
    "path/filepath"
    "regexp"
    "strings"
    "testing"
)

// TestBodyDimsMigration enforces UI-18: the banned expression
// `m.height - statusBarHeight(...)` must not appear outside the bodyDims
// helper body. Without this test, Phase 7 contributors could reintroduce
// the pattern without noticing.
func TestBodyDimsMigration(t *testing.T) {
    banned := regexp.MustCompile(`m\.height\s*-\s*statusBarHeight`)

    // Find the repo root by walking up from this test file until go.mod is found.
    repoRoot := findRepoRoot(t)

    // Phase 6: the only legitimate use is inside bodyDims in model.go.
    // Compute its line range once.
    helperStart, helperEnd := findBodyDimsRange(t, filepath.Join(repoRoot, "internal/app/model.go"))

    var violations []string
    err := filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if d.IsDir() {
            // Skip vendor, build artefacts, hidden dirs.
            name := d.Name()
            if name == "vendor" || name == ".git" || strings.HasPrefix(name, ".") {
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
        for lineNo, line := range strings.Split(string(content), "\n") {
            if !banned.MatchString(line) {
                continue
            }
            // Carve out the bodyDims body in model.go.
            if strings.HasSuffix(path, "internal/app/model.go") &&
                lineNo+1 >= helperStart && lineNo+1 <= helperEnd {
                continue
            }
            violations = append(violations, relPath(path, repoRoot)+
                ":"+itoa(lineNo+1)+"  "+strings.TrimSpace(line))
        }
        return nil
    })
    if err != nil {
        t.Fatalf("walk failed: %v", err)
    }
    if len(violations) > 0 {
        t.Fatalf("banned pattern `m.height - statusBarHeight` found outside bodyDims:\n  %s\n\n"+
            "Use bodyDims(m) instead; see UI-17/UI-18.",
            strings.Join(violations, "\n  "))
    }
}
```

**Key design points:**

1. **Carve-out strategy — string scan, not `go/ast`.** The helper has a stable `func bodyDims(m AppModel) (w, h int) {` signature (D-01). `findBodyDimsRange` does a simple linear scan for `^func bodyDims(` and then tracks brace depth to find the matching `}`. ~30 LOC. If Phase 7 ever changes the signature (unlikely per D-01 which future-proofs the shape), this becomes brittle — acceptable tradeoff for avoiding the `go/ast` infrastructure.
2. **Repo-root discovery.** Walk upward from `runtime.Caller(0)` looking for `go.mod`. Fail fast if not found.
3. **Skip directories.** `vendor/`, `.git/`, any hidden dir. No `testdata/` skip needed because testdata files are `.golden`, not `.go`.
4. **Self-test guard.** The test itself contains the banned regex literal (as a Go string). That literal is inside the `regexp.MustCompile(...)` argument on a string constant — the carve-out scans raw source lines, so the string `` `m\.height\s*-\s*statusBarHeight` `` DOES match the regex when applied to the test source file line-by-line. **Mitigation:** add an additional carve-out that exempts the test file itself by path suffix, OR encode the regex so it does not self-match (e.g., split the literal: `` `m\.height\s*-\s*` + `statusBarHeight` ``). The split-literal approach is cleaner — recommend that. [ASSUMED: the reviewer will verify this edge case during plan-check]

**`findBodyDimsRange` sketch:**

```go
func findBodyDimsRange(t *testing.T, modelPath string) (start, end int) {
    t.Helper()
    b, err := os.ReadFile(modelPath)
    if err != nil {
        t.Fatalf("read %s: %v", modelPath, err)
    }
    lines := strings.Split(string(b), "\n")
    start = 0
    for i, line := range lines {
        if strings.HasPrefix(line, "func bodyDims(") {
            start = i + 1 // 1-indexed
            break
        }
    }
    if start == 0 {
        t.Fatalf("func bodyDims not found in %s", modelPath)
    }
    // Brace tracking from the opening { of the signature line.
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
```

### Pattern 3: ANSI-Stripped Golden Comparison with Env-Gated Regen

**What:** Tests capture `m.View().Content`, pass it through `ansi.Strip`, compare against a `.golden` file. When `GOLDEN_UPDATE=1` is set, the helper writes the file instead of comparing (D-10).

**When to use:** All Phase 6 resize tests. Phase 7+ will add more goldens for chrome rendering.

**Why structural-only goldens (vs. raw ANSI):** Pitfall 8 — lipgloss emits slightly different ANSI sequences across versions (combined SGR runs, order of style application). Structural comparison survives lipgloss bumps; color presence is asserted separately via `RequireGoldenColors` (Phase 6 leaves the color assertions empty because no new chrome has landed).

**Example:**

```go
// Source: Phase 6 D-08, D-09, D-10 — repo-internal design
package testutil

import (
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/charmbracelet/x/ansi"
)

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

**Gotchas for ANSI-stripped comparison (per D-09 Claude's-discretion on normalisation):**

1. **Trailing whitespace.** `lipgloss.NewStyle().Width(w).Render(...)` pads with spaces. After `ansi.Strip`, those spaces remain. Editor auto-format rules ("strip trailing whitespace on save") will silently rewrite the golden. **Mitigation:** `normalise()` trims trailing whitespace per line before writing AND before comparing. Tradeoff: tests cannot distinguish `"foo"` from `"foo   "` — acceptable because the point of structural comparison is content, not cell-exact padding.
2. **Line endings.** `\r\n` (Windows checkout with `autocrlf=true`) vs `\n` (Unix). `normalise()` forces `\n` before comparing.
3. **Terminal-width clipping.** `lipgloss.NewStyle().Height(h).Render(content)` right-pads AND vertically pads with blank lines. After strip+normalise, those blank lines are bare `""` entries — matches across machines.
4. **Unicode.** `ansi.Strip` preserves UTF-8 correctly (its state machine handles `Utf8State` at `width.go:21`). Goldens committed as UTF-8 compare byte-for-byte.
5. **Control chars (non-ANSI).** `\t` literal tabs pass through `ansi.Strip`. Today no code path emits tabs into `View()` output, but if Phase 7 ever renders content with tabs, normalisation would need to expand them. Out of scope for Phase 6.
6. **Deterministic width/height.** `AppModel.View()` uses `m.width`/`m.height` only (no terminal query). Feeding `tea.WindowSizeMsg{Width: 40, Height: 12}` produces a fully reproducible output.

### Pattern 4: Benchmark with `testing.B.Loop()` (Go 1.26 idiom)

**What:** Measure `AppModel.View()` cost at 200×60 with the full Phase 5 feature set, producing a v1.0 baseline number for Phase 7 reviewers to compare the chrome's ≤ 50 µs/op target against.

**When to use:** Once, in Phase 6. Kept as a regression sentinel thereafter.

**Example:**

```go
// Source: Go 1.26 benchmark idiom, testing.B.Loop (benchmark.go:502)
package app

import (
    "testing"

    tea "charm.land/bubbletea/v2"
    "github.com/caesarakalaeii/sops-tui/internal/ui"
)

// BenchmarkAppView records the v1.0-baseline render cost so Phase 7's chrome
// skeleton has a concrete "before" number to compare its ≤ 50 µs/op target
// against. No gating on absolute value in Phase 6.
func BenchmarkAppView(b *testing.B) {
    env := ui.EnvStatus{
        SopsAvailable:     true,
        AgeAvailable:      true,
        SopsYamlAvailable: true,
        GitAvailable:      true,
    }
    m := NewAppModel(env, "")
    // Seed a realistic terminal size.
    updated, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 60})
    m = updated.(AppModel)

    b.ReportAllocs()
    for b.Loop() {
        _ = m.View()
    }
}
```

**Sample output (format only, not a guarantee of the actual number):**

```
BenchmarkAppView-12    123456    8234 ns/op    1456 B/op    23 allocs/op
```

Run with `go test -bench=BenchmarkAppView -benchmem ./internal/app/...`. The number is recorded in Phase 6's UAT notes; Phase 11 re-runs and compares.

### Anti-Patterns to Avoid

- **Hand-rolled ANSI regex.** `\x1b\[[0-9;]*m` catches CSI SGR only — misses OSC 8 (hyperlinks), DCS, control-char passthrough. Use `ansi.Strip`.
- **`go/ast` for the grep-gate.** Adds `go/parser` + `go/token` machinery for what is fundamentally a lexical check. The helper's signature is stable (D-01); string scan is sufficient.
- **Teatest `WaitFor` + `RequireEqualOutput`.** D-08 locks out external teatest infrastructure. `m.View().Content` is directly consumable because Phase 6 tests are synchronous (no async message wait needed).
- **`-update` flag instead of `GOLDEN_UPDATE` env var.** D-10 locked the env var specifically because `-update` needs `TestMain` + flag registration; env var is zero scaffolding.
- **Parametrised `BenchmarkAppView` (sub-benchmarks at multiple sizes).** D-12 locked the fixed 200×60 target. Adding sub-benchmarks now invents complexity Phase 11 would have to reason about.
- **Caching `m.View()` output or `statusBarHeight(m)` value.** Phase 6 is zero-behaviour-change; caching is a Phase 7 concern (Pitfall 2).

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| ANSI sequence removal | Regex-based stripper | `github.com/charmbracelet/x/ansi.Strip` | Full CSI/OSC/DCS/Utf8State state machine; handles malformed escapes; already resolved in go.mod. |
| Terminal height measurement | `len(strings.Split(s, "\n"))` | `lipgloss.Height(s)` | ANSI-aware (ignores escape bytes that contain `\n` inside OSC payloads). Existing `statusBarHeight` convention. |
| Repo walking in tests | Hand-rolled recursion | `filepath.WalkDir` (stdlib) | Correct directory skipping, symlink handling, error propagation. |
| Benchmark loop | `for i := 0; i < b.N; i++` | `for b.Loop()` | Go 1.26+ idiom; auto-handles timer reset and cleanup tail. |

**Key insight:** Phase 6 introduces zero new third-party code. Every primitive is either stdlib or already pinned in `go.mod`. The only `go.mod` change is promoting `x/ansi` from indirect to direct.

---

## Call-Site Inventory (verified against current model.go)

All line numbers verified on 2026-04-23 against `/home/moersener/git/sops-tui/internal/app/model.go` (1911 lines).

### 15 SetSize-adjacent sites matching `m\.height\s*-\s*statusBarHeight(m)`

| Line | Context (handler) | Sub-model created/sized | Migration |
|------|-------------------|-------------------------|-----------|
| **316** | `tea.WindowSizeMsg` | fan-out to 8 sub-models: fileList, detail, help, metadata, diff, history, health, recipientForm (lines 321–328) | `w, h := bodyDims(m)` then `m.fileList.SetSize(w, h)`, ...×8 |
| **349** | `FilesDiscoveredMsg` (after discovery) | `ui.NewFileListModel(items, m.width, mainH)` at L353 | `w, h := bodyDims(m); ui.NewFileListModel(items, w, h)` |
| **377** | `FilesParsedMsg` (after parse) | `ui.NewDetailModel(..., m.width, mainH, ...)` at L382–389 | `w, h := bodyDims(m); ui.NewDetailModel(..., w, h, ...)` |
| **485** | editor exit, multi-key diff | `ui.NewDiffModel(title, diffs, m.width, mainH)` at L489 | `w, h := bodyDims(m); ui.NewDiffModel(title, diffs, w, h)` |
| **502** | `ui.EditConfirmMsg` (inline single-key diff) | `ui.NewDiffModel(..., m.width, mainH)` at L506–510 | `w, h := bodyDims(m); ui.NewDiffModel(..., w, h)` |
| **567** | `ui.RotateReadyMsg` (rotate diff) | `ui.NewDiffModel(..., m.width, mainH)` at L571–574 | `w, h := bodyDims(m); ui.NewDiffModel(..., w, h)` |
| **631** | `GitStatusMsg` (file list rebuild with git badges) | `ui.NewFileListModel(items, m.width, mainH)` at L635 | `w, h := bodyDims(m); ui.NewFileListModel(items, w, h)` |
| **724** | add-recipient confirmation | `ui.NewDiffModel("Confirm: Add Recipient", entries, m.width, mainH)` at L728 | `w, h := bodyDims(m); ui.NewDiffModel(..., w, h)` |
| **761** | remove-recipient confirmation | `ui.NewDiffModel("Confirm: Remove Recipient", entries, m.width, mainH)` at L765 | `w, h := bodyDims(m); ui.NewDiffModel(..., w, h)` |
| **846** | health-check sentinel after diff confirm | `ui.NewHealthModel(m.width, mainH)` at L850 | `w, h := bodyDims(m); ui.NewHealthModel(w, h)` |
| **924** | format-menu selection (enter) | `ui.NewDiffModel(..., m.width, mainH)` at L928–932 | `w, h := bodyDims(m); ui.NewDiffModel(..., w, h)` |
| **1005** | `i` key → metadata | `ui.NewMetadataModel(meta, m.width, mainH)` at L1009 | `w, h := bodyDims(m); ui.NewMetadataModel(meta, w, h)` |
| **1089** | git-history key | `ui.NewHistoryModel(m.currentFile.Name, m.width, mainH)` at L1093 | `w, h := bodyDims(m); ui.NewHistoryModel(..., w, h)` |
| **1110** | `a` key → add-recipient form | `ui.NewRecipientFormModel(m.width, mainH)` at L1114 | `w, h := bodyDims(m); ui.NewRecipientFormModel(w, h)` |
| **1250** | health-scan confirm prompt | `ui.NewDiffModel(..., m.width, mainH)` at L1254–1256 | `w, h := bodyDims(m); ui.NewDiffModel(..., w, h)` |

**Before pattern** (each site today):

```go
mainH := m.height - statusBarHeight(m)
if mainH < 0 {
    mainH = 0
}
m.someSub = ui.NewFooModel(..., m.width, mainH)
```

**After pattern** (each site in Plan 2):

```go
w, h := bodyDims(m)
m.someSub = ui.NewFooModel(..., w, h)
```

Net removal: 3 lines per site (the `mainH :=`, the `if mainH < 0`, the `mainH = 0`), plus minor token substitution on the `NewFooModel` call. **~45 line reduction repo-wide.**

### Outlier 1: `model.go:1333` — inside `View()` body

Current code (lines 1329–1336):

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

Note: `statusBarH := lipgloss.Height(statusBar)` is a local computed once, then `mainH := m.height - statusBarH` — the banned regex (`m.height - statusBarHeight`) does NOT match this exact text (it uses `statusBarH`, not `statusBarHeight(m)`). **Migration is cosmetic-consistency not regex-compliance:** rewrite to flow through `bodyDims` so Phase 7's chrome/crumbs subtraction is automatically picked up.

**After (Plan 2):**

```go
func (m AppModel) View() tea.View {
    statusBar := m.status.View(m.width)
    _, mainH := bodyDims(m)
```

Tradeoff: `statusBar` is still computed directly (it's rendered below at L1371). `statusBarH` local is no longer needed because `bodyDims` internally calls `statusBarHeight(m)` which re-renders — a minor redundancy (two renders of the same status bar per frame) that is acceptable because Phase 6 is zero-behaviour-change and `statusBarHeight` is O(1) string composition. Phase 11's performance pass can collapse the redundancy if `BenchmarkAppView` flags it.

### Outlier 2: `model.go:1799` — `showBulkReKeyConfirm`

Current code (line 1799):

```go
func (m *AppModel) showBulkReKeyConfirm(file sops.DiscoveredFile) {
    // ... lines 1790-1798 build entries ...
    mainH := m.height - statusBarHeight(*m)
    if mainH < 0 {
        mainH = 0
    }
    m.diff = ui.NewDiffModel(fmt.Sprintf("Confirm Re-key: %s", file.Name), entries, m.width, mainH)
```

The difference here is the **pointer receiver**: `statusBarHeight(*m)` dereferences. The banned regex `m\.height\s*-\s*statusBarHeight` DOES match the raw line (the `(*m)` argument is irrelevant to regex matching).

**After (Plan 2):** Add a pointer-friendly overload OR dereference at the call site.

Option A (recommended — minimal change): dereference.

```go
w, h := bodyDims(*m)
m.diff = ui.NewDiffModel(fmt.Sprintf("Confirm Re-key: %s", file.Name), entries, w, h)
```

Option B (if pointer-receiver sites multiply in Phase 7+): add a method receiver form `func (m *AppModel) bodyDims() (w, h int) { return bodyDims(*m) }`. **Reject for Phase 6** — YAGNI; this is the only pointer site today.

### Outlier 3 (DEFERRED): `model.go:1862` — `renderRecipientList`

Current code:

```go
boxHeight := m.height - 4
if boxHeight < 1 {
    boxHeight = 1
}
return lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    // ...
    Height(boxHeight).
    Render(inner)
```

The `- 4` is a modal-frame allowance (borders + padding), not a status-bar concern. It predates the rename to `bodyDims`. The banned regex does not match this line.

**Action in Plan 2:** Insert a comment immediately above line 1862:

```go
// TODO(phase-7): replace magic -4 with a named modal-frame constant or
// bodyDims usage once modal chrome lands.
boxHeight := m.height - 4
```

No functional change. Tagged for Phase 7 or 8 triage.

---

## Golden Fixture Design

### Resize test structure

Each resize test follows the same shape:

```go
func TestResize_80x24(t *testing.T) {
    m := NewAppModel(defaultEnv(), "")
    updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
    m = updated.(AppModel)

    v := m.View()
    testutil.RequireGoldenStructure(t, "resize_80x24", v.Content)
    testutil.RequireGoldenColors(t, "resize_80x24", v.Content, nil) // Phase 6: no color asserts
}
```

`defaultEnv()` follows the existing `model_test.go:16–22` pattern — all-true env flags so the bottom-right status bar is deterministic across machines.

### Determinism audit

**Does `AppModel.View()` leak any machine-dependent state?**

Grep results across the rendering path:

| Potential leak source | Found in View path? | Result |
|-----------------------|---------------------|--------|
| `time.Now()` | `model.go:1387, 1586, 1588, 1690` — all outside View (clipboard timer, $EDITOR resolution, staleness env) | No leak |
| `os.Getenv("EDITOR")` / `VISUAL` | `model.go:1586, 1588` — only invoked on edit, not during View | No leak |
| `os.Getwd()` | Not found anywhere | No leak |
| `filelist.go` empty state | Deterministic 4-line string (`filelist.go:320–338`) | No leak |
| `statusbar.go` env indicators | Driven by `EnvStatus` struct fields (all-true in tests) | No leak |
| `statusbar.go` breadcrumb | `"sops-tui"` + segments; initial state is just `"sops-tui"` | No leak |

**Conclusion:** `NewAppModel(defaultEnv(), "")` followed by `WindowSizeMsg{W, H}` → `View()` produces a fully reproducible `Content` string. No test fixture files need to be loaded from disk.

### Fixture content preview (predictions)

At **80×24** on a fresh `AppModel` the view is:

```
No SOPS files found

No .sops.yaml discovered in this directory or parents.
Run sops-tui in a repository with a .sops.yaml configuration.
[blank rows padding to height 23]
sops-tui                  |   0 items   |   sops:✓  age:✓  .sops.yaml:✓  no git
```

After `ansi.Strip` + `normalise` (trim trailing whitespace per line, one `\n` per row), the golden is 24 lines of text. Height should be `height - statusBarHeight(m)` for the body (23) plus 1 status-bar line = 24 total. [ASSUMED: the exact blank-line padding depends on `lipgloss.NewStyle().Height(mainH).Render(content)` behaviour at `model.go:1368`; generate via `GOLDEN_UPDATE=1` in Plan 2 and visually inspect before committing]

At **40×12** the body is narrow enough that the "No SOPS files found" text remains un-truncated but the status-bar sections may collapse. At **200×60** there is generous padding. At **120×40** is the "most likely to show bugs" size per Pitfall 1.

**Recommendation for Plan 2:** after writing `resize_test.go`, run `GOLDEN_UPDATE=1 go test ./internal/app/... -run TestResize` once, visually inspect the four `.golden` files, commit. Subsequent runs without `GOLDEN_UPDATE` must pass.

---

## Common Pitfalls

### Pitfall 1: The grep-gate test's own source file contains the banned regex literal

**What goes wrong:** The test file declares `regexp.MustCompile(`m\.height\s*-\s*statusBarHeight`)`. When the test walks `*.go` files, it reads its own source and matches the literal.

**Why it happens:** The carve-out only exempts `internal/app/model.go:<bodyDims range>` — nothing exempts the test file.

**How to avoid:** Split the regex literal across string concatenation so the test source does NOT contain the full banned sequence as a single substring:

```go
// Deliberate split — the test file must not contain the full banned
// pattern as one contiguous literal, or TestBodyDimsMigration would
// match itself.
banned := regexp.MustCompile(`m\.height\s*-\s*` + `statusBarHeight`)
```

**Warning signs:** `TestBodyDimsMigration` fails reporting itself as a violation.

### Pitfall 2: Pointer vs value receiver when calling bodyDims

**What goes wrong:** `model.go:1799` uses `m.height - statusBarHeight(*m)` inside a pointer-receiver method. A naive rewrite to `bodyDims(m)` compiles-fails because `bodyDims` takes `AppModel` (value), not `*AppModel`.

**How to avoid:** Dereference at the call site: `w, h := bodyDims(*m)`. Do not add a pointer-receiver overload (YAGNI — only one site).

**Warning signs:** Build breaks on Plan 2 with `cannot use m (*AppModel) as AppModel in argument to bodyDims`.

### Pitfall 3: Trailing whitespace drift in goldens

**What goes wrong:** `lipgloss.Width().Render()` right-pads rows with spaces. Editor "strip trailing whitespace on save" rules silently rewrite the committed golden, breaking CI.

**How to avoid:**
1. `normalise()` trims trailing whitespace from every line before both writing AND comparing.
2. Add `*.golden` to `.gitattributes` as `text -whitespace` or equivalent so git's whitespace cleanup does not touch fixtures. [ASSUMED: project uses default git attributes; verify during Plan 2 execution]
3. Document the rule in a short `internal/app/testdata/README.md` (optional — Claude's discretion per D-11).

**Warning signs:** Goldens pass on laptop, fail in CI, or a diff appears "empty" (whitespace-only) in `git diff`.

### Pitfall 4: `lipgloss.Height` double-render cost in bodyDims

**What goes wrong:** Every call to `bodyDims` calls `statusBarHeight(m)`, which calls `m.status.View(m.width)` then `lipgloss.Height(...)`. The status bar is rendered once for measurement, then again in `View()` for output. Per render: two status-bar renders.

**How to avoid:** Accept it for Phase 6 (zero-behaviour-change rule). Phase 11's `BenchmarkAppView` will flag it if it materially affects the ≤ 50 µs/op target. If needed later, introduce a `headerCache` / `statusBarCache` field on AppModel (Pitfall 2 of the milestone research) — explicitly out of Phase 6 scope.

**Warning signs:** `BenchmarkAppView` shows allocations scaling with `View()` count proportional to status-bar complexity.

### Pitfall 5: `lipgloss.NoColor` profile does not exist in lipgloss v2

**What goes wrong:** CLAUDE.md mentions "Force `lipgloss.NoColor` profile in tests to avoid CI color-profile divergence" — that was the v1 API. In v2, the global `Renderer` and `SetColorProfile` were removed (`UPGRADE_GUIDE_V2.md:216–242`). `Style.Render()` always emits full-fidelity ANSI; downsampling happens at the output writer.

**How to avoid:** The test harness does not need (and cannot use) a profile switch. `ansi.Strip` produces profile-independent structural output regardless of what ANSI sequences `lipgloss` emitted. Color presence assertions (Phase 7+) will match against the raw ANSI emitted by `lipgloss.Color("#…").Foreground().Render(...)` — no downsampling in the pipeline.

**Warning signs:** A test that does `lipgloss.SetColorProfile(...)` fails to compile (symbol removed in v2).

### Pitfall 6: Carve-out line range shifts after any model.go edit

**What goes wrong:** `findBodyDimsRange` returns concrete line numbers. If a contributor adds code above the helper, all subsequent line numbers shift. The test still works (it recomputes the range on every run) — but comments in PRs referencing "line 1443" go stale.

**How to avoid:** Document in the test's doc comment that line numbers are computed dynamically; review comments should reference function names, not line numbers.

**Warning signs:** None at runtime — this is a docs/review-ergonomics concern, not a test correctness concern.

### Pitfall 7: `.sops.yaml` content in goldens at 200×60

**What goes wrong:** `NewAppModel(defaultEnv(), "")` passes empty `sopsYamlPath`. The FileListModel renders its empty state (no files) — fully deterministic.

**How to avoid:** Do NOT seed a fake `.sops.yaml` path or fake files for Phase 6 goldens. The empty-state render is the phase's regression target. File-list-with-items goldens are a Phase 2-era test concern, not Phase 6.

**Warning signs:** Plan 2 tempted to pass fixture files into `NewAppModel` for "more realistic" goldens — push back, that's Phase 7+ scope.

---

## Code Examples

All recommended code appears in `## Architecture Patterns` §1–§4 above. Summary of sources:

| Function / pattern | Source | Verified |
|--------------------|--------|----------|
| `bodyDims` / `chromeHeight` / `crumbsHeight` bodies | CONTEXT.md D-01 through D-04 | LOCKED |
| `ansi.Strip` usage | `~/go/pkg/mod/github.com/charmbracelet/x/ansi@v0.11.7/width.go:10` | [VERIFIED: local module cache] |
| `testing.B.Loop()` | `/usr/lib/go/src/testing/benchmark.go:502` | [VERIFIED: local stdlib] |
| `tea.View.Content` | `~/go/pkg/mod/charm.land/bubbletea/v2@v2.0.4/tea.go:96` | [VERIFIED: local module cache] |
| `filepath.WalkDir` | stdlib, current Go docs | [CITED: pkg.go.dev/path/filepath#WalkDir] |
| `GOLDEN_UPDATE=1` env var pattern | CONTEXT.md D-10 | LOCKED |

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `for i := 0; i < b.N; i++` in benchmarks | `for b.Loop()` | Go 1.24 (stable in 1.26) | Automatic timer reset; no manual `b.ResetTimer()` needed. Simpler code. [VERIFIED: benchmark.go:502] |
| `lipgloss.DefaultRenderer().SetColorProfile(...)` | `colorprofile.Detect(os.Stdout, os.Environ())` at output layer | lipgloss v2.0.0 (Mar 2025) | Tests do NOT set a profile; use `ansi.Strip` for structural comparison instead. [VERIFIED: UPGRADE_GUIDE_V2.md:216–242] |
| `teatest.RequireEqualOutput(t, got)` for golden files | Hand-rolled `RequireGoldenStructure` with `ansi.Strip` | Phase 6 decision (D-08) | Avoids teatest dep for synchronous snapshot tests; structural-only comparison insulates from lipgloss ANSI emission churn. |
| `-update` flag with `TestMain` | `GOLDEN_UPDATE=1` env var | Phase 6 decision (D-10) | Zero scaffolding; intentional friction per Pitfall 8. |

**Deprecated / outdated:**

- **`lipgloss.SetColorProfile`** — removed in v2. Any doc suggesting "force NoColor profile in tests" is v1 advice and does not apply.
- **`b.ResetTimer()` manual reset** — superseded by `b.Loop()`'s auto-reset semantics.

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Splitting the banned-regex literal in the test file avoids self-match | Grep-Gate Pitfall 1 | Test self-reports as a violation; mitigation is the split-literal trick. Easy to catch during plan-check. |
| A2 | `.gitattributes` / git whitespace handling leaves `.golden` fixtures untouched | Pitfall 3 | Windows checkouts might normalise line endings; adding `*.golden binary` or `text -whitespace` in `.gitattributes` mitigates. Verify during Plan 2. |
| A3 | 200×60 empty-state golden size / exact padding lines (24 rows at 80×24, etc.) | Golden Fixture Design | Predicted content based on reading `View()` + `filelist.go` + `statusbar.go`; exact padding depends on runtime `lipgloss.NewStyle().Height(...)` behaviour. Generate with `GOLDEN_UPDATE=1`, inspect, commit. |
| A4 | Introducing `bodyDims(m)` does not ripple into `_test.go` files that construct `AppModel` manually | Call-Site Inventory | Grep for `m.height - statusBarHeight` returned only the 17 sites in `model.go` — no test file currently replicates the expression. [VERIFIED: Grep run in this session returned 16 matches, all in model.go] |
| A5 | `go mod tidy` after promoting `x/ansi` does not churn `go.sum` | Installation | Version v0.11.7 already resolved as transitive of lipgloss — just the `// indirect` comment is removed. Low risk. |

---

## Open Questions (RESOLVED)

All five questions defer to Claude's discretion per CONTEXT.md §"Claude's Discretion" / D-09 / D-12. Concrete dispositions below.

1. **Should `.golden` files get a `.gitattributes` entry to prevent whitespace mangling?**
   - RESOLVED: Yes — Plan 2 adds `.gitattributes` at repo root with `*.golden text eol=lf -whitespace`. One-line change; insulates against Windows CRLF rewriting and editor auto-format.

2. **Where does `BenchmarkAppView` live — `bench_test.go`, `model_test.go`, or `layout_test.go`?**
   - RESOLVED: New file `internal/app/bench_test.go`. Keeps benchmarks discoverable separately from unit tests. (Claude's discretion per D-12.)

3. **Does Plan 2's atomic migration include a commit-message trailer listing all 17 migrated sites?**
   - RESOLVED: Commit body references `PLAN 2` and points to the call-site inventory in `06-RESEARCH.md` §"Call-Site Inventory" rather than enumerating inline. D-14 "one commit" is preserved.

4. **Does the grep-gate test need to run on `go generate` output or only committed files?**
   - RESOLVED: Committed files only. No `go:generate` directives exist in the repo today. If added later, the offending directory can be added to the walk skip list at that time.

5. **Should `RequireGoldenStructure` emit a unified diff (`diff -u` style) on mismatch, or a raw side-by-side?**
   - RESOLVED: Labelled sections (`--- want --- / --- got ---`). No external diff library — avoids adding `diffmatchpatch` dep (D-09).

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | all Phase 6 work | ✓ | 1.26.2-X | — |
| `github.com/charmbracelet/x/ansi` | golden harness (D-08) | ✓ (indirect) | v0.11.7 | — |
| `charm.land/bubbletea/v2` | benchmark, resize tests | ✓ | v2.0.4 | — |
| `charm.land/lipgloss/v2` | statusBarHeight, View() | ✓ | v2.0.3 | — |
| `github.com/stretchr/testify` | existing test style | ✓ | v1.11.1 | — |
| `sops` binary | NOT used by Phase 6 tests | — | — | — |
| age key | NOT used by Phase 6 tests | — | — | — |
| git | NOT used by Phase 6 tests | — | — | — |

**Missing dependencies with no fallback:** none.

**Missing dependencies with fallback:** none.

All Phase 6 tests run entirely in-process against `AppModel` with no subprocess, no disk I/O beyond `testdata/` fixtures, no network.

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` (Go 1.26.2) + `stretchr/testify` v1.11.1 |
| Config file | none (stdlib) |
| Quick run command | `go test ./internal/app/... ./internal/testutil/...` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| **UI-17** | `bodyDims(m)` returns `(m.width, m.height - statusBarHeight(m) - chromeHeight(m) - crumbsHeight(m))` clamped ≥ 0 | unit | `go test ./internal/app/... -run 'TestBodyDims$' -v` | ❌ Plan 1 creates `layout_test.go::TestBodyDims` |
| **UI-17** | `bodyDims(m)` clamps negative values to 0 | unit | `go test ./internal/app/... -run TestBodyDimsClampsAtZero -v` | ❌ Plan 1 creates |
| **UI-17** | `chromeHeight(m)` returns 0 in Phase 6 | unit | `go test ./internal/app/... -run TestChromeHeightReturnsZero -v` | ❌ Plan 1 creates |
| **UI-17** | `crumbsHeight(m)` returns 0 in Phase 6 | unit | `go test ./internal/app/... -run TestCrumbsHeightReturnsZero -v` | ❌ Plan 1 creates |
| **UI-17** | All 15 SetSize sites + 2 outliers route through `bodyDims` (observational — no visual regression at 4 terminal sizes) | integration / golden | `go test ./internal/app/... -run TestResize -v` | ❌ Plan 2 creates `resize_test.go` + 4 `.golden` files |
| **UI-18** | Banned regex does not appear outside `bodyDims` body | integration | `go test ./internal/app/... -run TestBodyDimsMigration -v` | ❌ Plan 1 creates the test; PASSES only after Plan 2 migrates all sites |
| **UI-18** | Regression test: contributor-introduced violation fails | manual / test review | Plan 1 plan-check reviews the test by temporarily adding a `m.height - statusBarHeight(m)` line to a random file and confirming the test fails, then reverts | ❌ Plan 1 plan-check step |
| **UI-19** | `RequireGoldenStructure` writes on `GOLDEN_UPDATE=1`, compares on unset | unit | `go test ./internal/testutil/... -run TestRequireGoldenStructure -v` | ❌ Plan 1 creates `golden_test.go` |
| **UI-19** | `RequireGoldenStructure` round-trips through `ansi.Strip` correctly | unit | `go test ./internal/testutil/... -run TestRequireGoldenStructure_ANSIStrip -v` | ❌ Plan 1 creates |
| **UI-19** | `RequireGoldenColors` passes empty `wantColors` (Phase 6 no-op) | unit | `go test ./internal/testutil/... -run TestRequireGoldenColors_Empty -v` | ❌ Plan 1 creates |
| **UI-19** | `normalise` trims trailing whitespace and normalises line endings | unit | `go test ./internal/testutil/... -run TestNormalise -v` | ❌ Plan 1 creates |
| *(baseline)* | `BenchmarkAppView` records v1.0 cost at 200×60 | benchmark | `go test -bench=BenchmarkAppView -benchmem ./internal/app/...` | ❌ Plan 1 creates `bench_test.go` |
| *(regression)* | All v1.0 existing tests continue to pass | integration | `go test ./...` | ✓ Existing 25 test files |

### Sampling Rate

- **Per task commit (Plan 1):** `go test ./internal/app/... ./internal/testutil/...` (≤ 3 seconds)
- **Per task commit (Plan 2):** `go test ./internal/app/...` (includes grep-gate + 4 resize tests; ≤ 5 seconds)
- **Per wave merge:** `go test ./...` (full suite, ≤ 30 seconds)
- **Phase gate:** Full suite green + `BenchmarkAppView` ran once (number recorded, no threshold) + manual smoke at 40×12 and 200×60 per D-15.

### Wave 0 Gaps

Plan 1 must land before Plan 2 can run. Explicit Wave 0 items:

- [ ] `internal/testutil/golden.go` — helpers `RequireGoldenStructure`, `RequireGoldenColors`, `normalise`
- [ ] `internal/testutil/golden_test.go` — round-trip + env-var-gated regen tests
- [ ] `internal/app/layout_test.go` — helper unit tests + `TestBodyDimsMigration` grep-gate
- [ ] `internal/app/bench_test.go` — `BenchmarkAppView`
- [ ] `internal/app/model.go` — helper definitions (`bodyDims`, `chromeHeight`, `crumbsHeight`)
- [ ] `go.mod` — promote `github.com/charmbracelet/x/ansi` from indirect to direct

No Wave 0 gaps for Plan 2 beyond what Plan 1 delivers. Plan 2's resize tests import `internal/testutil` which Plan 1 creates.

---

## Project Constraints (from CLAUDE.md)

Extracted from `./CLAUDE.md` for planner compliance checks:

1. **Never use the type `any`; use the proper typing** — Phase 6 code uses concrete types (`AppModel`, `*testing.T`, `[]string`, `[]byte`). No `any` required.
2. **Go + Bubble Tea (Charm ecosystem)** — all Phase 6 work stays within the existing stack.
3. **License: AGPL-3.0** — new files should carry a license header per existing convention (check `internal/ui/statusbar.go:1` style for the pattern).
4. **No `lipgloss.AdaptiveColor` (issue #1036)** — Phase 6 writes no color code. N/A.
5. **`charm.land/bubbletea/v2` v2.0.4** — `View()` returns `tea.View` struct, `tea.KeyPressMsg` is the key event type. `WindowSizeMsg{Width, Height}` is the resize event. Benchmarks + resize tests use this API. [VERIFIED]
6. **testify + `require`/`assert`** — new tests follow the existing `model_test.go` style (import `require` for preconditions, `assert` for value checks).
7. **GSD Workflow Enforcement** — Phase 6 work happens through `/gsd-execute-phase`; no direct edits.
8. **Tests live beside implementation** — new test files go in `internal/app/` (not a central `tests/` dir) and `internal/testutil/`.

---

## Sources

### Primary (HIGH confidence — local verification)

- `/home/moersener/git/sops-tui/internal/app/model.go` (1911 lines) — all 17 call-site line numbers, outlier exact text, helper location `model.go:1443` [VERIFIED: Grep + Read in this session]
- `/home/moersener/git/sops-tui/go.mod` (56 lines) — confirmed `github.com/charmbracelet/x/ansi v0.11.7 // indirect` at line 26 [VERIFIED]
- `/home/moersener/git/sops-tui/internal/ui/statusbar.go` (237 lines) — verified `View()` pathway, `SetBreadcrumb` fan-in pattern, empty-state determinism [VERIFIED]
- `/home/moersener/git/sops-tui/internal/ui/filelist.go:320–338` — empty-state render is 4 deterministic lines [VERIFIED]
- `/home/moersener/git/sops-tui/internal/app/model_test.go` — existing test style (`defaultEnv()`, `send()` helper, `tea.KeyPressMsg{Code: …}` API) [VERIFIED]
- `/home/moersener/go/pkg/mod/github.com/charmbracelet/x/ansi@v0.11.7/width.go:10` — `func Strip(s string) string` signature + state machine implementation [VERIFIED: Read]
- `/home/moersener/go/pkg/mod/charm.land/bubbletea/v2@v2.0.4/tea.go:76–126` — `tea.NewView` + `View.Content` field [VERIFIED: Read]
- `/home/moersener/go/pkg/mod/charm.land/lipgloss/v2@v2.0.3/UPGRADE_GUIDE_V2.md:216–276` — renderer removal, profile detection at output layer [VERIFIED: Read]
- `/usr/lib/go/src/testing/benchmark.go:502` — `func (b *B) Loop() bool` [VERIFIED: Grep + Bash `go doc`]
- `/home/moersener/git/sops-tui/.planning/phases/06-layout-groundwork/06-CONTEXT.md` — decisions D-01 through D-15 (source of truth for this phase) [VERIFIED: Read]

### Secondary (MEDIUM confidence — CONTEXT.md + PITFALLS.md attribution)

- `.planning/research/PITFALLS.md` §"Pitfall 1" — highest-severity regression risk, all 15 call-site line numbers matched this research session's grep results.
- `.planning/research/PITFALLS.md` §"Pitfall 8" — ANSI-stripped goldens + separate color assertions pattern.
- `.planning/research/SUMMARY.md` §"Phase 6" — delivers list aligns with CONTEXT.md.
- `.planning/research/ARCHITECTURE.md` §"Integration Pitfall 6" — confirms same 15-site list from independent grep.

### Tertiary (LOW confidence — training knowledge)

- **None.** Every claim in this document is backed by either a local file read or a CONTEXT.md locked decision. The Assumptions Log (§above) enumerates the few predictions (golden file exact contents, git attribute interactions) that need empirical verification during Plan 2 execution.

---

## Metadata

**Confidence breakdown:**

- Call-site inventory: HIGH — grep run in this session returned exactly the 16 lines CONTEXT.md enumerates, at exactly the line numbers CONTEXT.md gives.
- `x/ansi` API shape: HIGH — local module read, function signature unambiguous.
- lipgloss v2 color profile behaviour: HIGH — UPGRADE_GUIDE_V2.md explicit about `SetColorProfile` removal.
- Golden fixture exact content: MEDIUM — predictions for row counts and blank-line padding are reasoned from code reading, not generated. Plan 2 must regenerate and visually inspect.
- Grep-gate self-match mitigation: MEDIUM — split-literal trick is well-known but must be applied deliberately; plan-check step should verify.
- `BenchmarkAppView` absolute numbers: N/A — no predicted number; recording the baseline IS the deliverable.

**Research date:** 2026-04-23
**Valid until:** 2026-06-23 (60 days — stable stack, no fast-moving deps in Phase 6 scope)
