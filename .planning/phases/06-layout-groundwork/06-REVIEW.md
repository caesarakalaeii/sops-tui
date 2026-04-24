---
phase: 06-layout-groundwork
reviewed: 2026-04-23T00:00:00Z
depth: standard
files_reviewed: 9
files_reviewed_list:
  - internal/app/model.go
  - internal/app/layout_test.go
  - internal/app/bench_test.go
  - internal/app/resize_test.go
  - internal/testutil/golden.go
  - internal/testutil/golden_test.go
  - internal/app/testdata/resize_40x12.golden
  - internal/app/testdata/resize_80x24.golden
  - internal/app/testdata/resize_120x40.golden
  - internal/app/testdata/resize_200x60.golden
  - .gitattributes
  - go.mod
findings:
  critical: 0
  warning: 0
  info: 4
  total: 4
status: has_findings
---

# Phase 6: Code Review Report

**Reviewed:** 2026-04-23
**Depth:** standard
**Files Reviewed:** 9 source + 4 golden fixtures + 2 build-config
**Status:** has_findings (4 x info; no blockers, no highs)

## Summary

Phase 6 is a mechanical refactor-plus-test-infra landing: introduces three helpers
(`bodyDims`, `chromeHeight`, `crumbsHeight`) around model.go:1393-1426, migrates 17
call-sites from the duplicated `m.height - statusBarHeight(m)` + clamp boilerplate
to single-line `bodyDims(m)` / `bodyDims(*m)` invocations, and adds a grep-gate
unit test (`TestBodyDimsMigration`) plus a minimal ANSI-stripping golden harness
(`internal/testutil`). `go test ./...` passes, `go vet ./...` is clean, and the
LIVE grep-gate demonstrably fails when the banned pattern is reintroduced (I
exercised it mentally against the assembled-from-three-strings regex construction).

**Threat model (T-06-01 — T-06-11):** No user-facing attack surface expanded. The
only new execution paths are a test helper that writes under `testdata/` (see
IN-01 below) and a benchmark wrapper. Zero network, zero subprocess, zero
filesystem writes from production code.

**Zero-behaviour-change invariant:** Verified via `git show 79ec449` diff. Each
of the 17 sites is a mechanical `mainH := m.height - statusBarHeight(m); if
mainH < 0 { mainH = 0 }` → `w, h := bodyDims(m)` rewrite. No arity, ordering, or
semantic change. Values previously derived per-call-site from the same expression
are now derived from the same expression in a helper — identical result.

**Helpers verified correct:**
- `bodyDims` (model.go:1403) correctly returns `(m.width, max(0, m.height -
  statusBarHeight - chromeHeight - crumbsHeight))`. Since Phase 6 stubs
  `chromeHeight` and `crumbsHeight` to return 0, behaviour is identical to the
  pre-migration `m.height - statusBarHeight(m)`.
- Clamp at zero: covered by `TestBodyDimsClampsAtZero` at height=1 (where
  `statusBarHeight` returns >=1 and subtraction would underflow without clamp).
- Integer underflow concern: not possible — both operands are native `int`,
  the clamp catches the negative case, and there is no `uint` conversion anywhere
  in the chain.
- `m.height == 0` edge case: `h` becomes `0 - statusBarHeight(m)` (negative) and
  is clamped to 0. `bodyDims` returns `(0, 0)` on a zero-size window. Safe.

**Outlier migrations verified:**
- `View()` outlier (model.go:1284-1327): the old local `statusBar := m.status.
  View(m.width); statusBarH := lipgloss.Height(statusBar); mainH := m.height -
  statusBarH` was refactored to `statusBar := m.status.View(m.width); _, mainH
  := bodyDims(m)`. The `statusBar` local is preserved for `lipgloss.JoinVertical`;
  no dangling `statusBarH` reference (grep confirms zero hits).
- Pointer-variant outlier (model.go:1779): `w, h := bodyDims(*m)` — deref is
  correct; `bodyDims` takes `AppModel` by value.
- Magic-`-4` preserved at model.go:1841 with TODO(phase-7) tag at line 1839-1840
  as specified. Text matches plan exactly.

**Golden harness verified:**
- `RequireGoldenStructure` strips ANSI, normalises line endings to LF, and
  right-trims whitespace per line. Self-tests cover the three main paths
  (update, compare, ANSI strip) and normalise.
- `.gitattributes` pins `*.golden text eol=lf -whitespace` — correct CRLF
  defence for Windows checkouts.
- Goldens at 40x12, 80x24, 120x40, 200x60 contain no cwd, timestamps, PIDs,
  or env vars. They render only the fixed "No SOPS files found" placeholder
  plus the status bar with `defaultEnv()` values (all bools fixed → no drift).

## Critical Issues

None.

## High

None.

## Medium

None.

## Low

None.

## Info

### IN-01: `RequireGoldenStructure` name parameter permits directory traversal on GOLDEN_UPDATE=1 writes

**File:** `internal/testutil/golden.go:30-44`
**Issue:** The `name` parameter is passed directly into `filepath.Join("testdata",
name+".golden")`. When `GOLDEN_UPDATE=1`, a malicious or buggy caller using
`name = "../../../etc/passwd"` would cause `os.WriteFile` to write the golden to
`../../etc/passwd.golden` outside the `testdata/` dir. Verified empirically:
`filepath.Join("testdata", "../../../etc/passwd"+".golden")` yields
`../../etc/passwd.golden` and `filepath.Clean` does not sanitise the escape.

This is a **test-only helper invoked under developer control** — no untrusted
input reaches it. Severity is Info (not Warning), because:
1. The caller is always a test file authored by a developer.
2. `GOLDEN_UPDATE=1` is an opt-in dev-mode env var.
3. No production code path reaches `RequireGoldenStructure`.

**Fix (defensive, optional):**
```go
func RequireGoldenStructure(t *testing.T, name, output string) {
    t.Helper()
    if strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") {
        t.Fatalf("golden name %q must be a flat identifier (no path separators or '..')", name)
    }
    // ... existing body
}
```

### IN-02: `View()` now renders the status bar twice per frame via `bodyDims(m)` callback

**File:** `internal/app/model.go:1284-1287`
**Issue:** Pre-migration `View()` computed `statusBarH` from the already-rendered
`statusBar` local via `lipgloss.Height(statusBar)` — one render per frame.
Post-migration, `View()` still renders `statusBar := m.status.View(m.width)` at
line 1286, and then the very next line calls `bodyDims(m)`, which internally
calls `statusBarHeight(m)` which calls `m.status.View(m.width)` AGAIN and
computes `lipgloss.Height(...)`. That doubles the status-bar render cost per
`View()` invocation — a minor hot-path regression captured in `BenchmarkAppView`.

Not a correctness bug: `StatusBarModel.View` is a pure value-receiver pure
function (verified in `internal/ui/statusbar.go:137`), so the two renders
produce identical output and identical heights. The commit message explicitly
notes the `statusBar` local is "preserved for rendering" — the author is aware.

**Fix (optional, Phase 7):** the natural place to land this optimisation is
when Phase 7 refactors View() for chrome. One option:

```go
statusBar := m.status.View(m.width)
statusBarH := lipgloss.Height(statusBar)
mainH := m.height - statusBarH - chromeHeight(m) - crumbsHeight(m)
if mainH < 0 { mainH = 0 }
```

But that would reintroduce an inline subtraction that the grep-gate forbids
(regex `m\.height\s*-\s*statusBarHeight` would not match since the local is
`statusBarH`, but the whole point of UI-17 is to flow through the helper).
Alternative: introduce `func bodyDimsWithStatusBar(m AppModel, statusBarH int)
(w, h int)` if the double-render shows up in the baseline benchmark that
this phase added.

**Recommendation:** Leave as-is for Phase 6; benchmark first, then decide in
Phase 7 if the double-render materially affects `<= 50µs/op` target.

### IN-03: `TestBodyDimsMigration` walker does not explicitly skip `testdata/` directories

**File:** `internal/app/layout_test.go:91-124`
**Issue:** The walker skips `vendor`, `.git`, and dotfile-prefixed dirs (e.g.
`.planning`, `.github`). It does NOT skip `testdata/` directories. Today this
does not matter — the only `testdata/` (at `internal/app/testdata/`) contains
only `.golden` files, which fail the `strings.HasSuffix(path, ".go")` filter
on line 102. But a future contributor adding a `*.go` test fixture under
`testdata/` could trip the banned regex, causing a test failure from what
is effectively a fixture, not live code.

Similarly, the carve-out for bodyDims uses `strings.HasSuffix(path,
"internal/app/model.go")` — so any vendored or embedded copy of model.go
(e.g. in a test fixture directory) would NOT have its bodyDims range carved
out, causing false-positive violations. Again: not a problem today.

**Fix (defensive):**
```go
if d.IsDir() {
    name := d.Name()
    if name == "vendor" || name == "testdata" || name == ".git" || (strings.HasPrefix(name, ".") && name != ".") {
        return filepath.SkipDir
    }
    return nil
}
```

### IN-04: `(strings.HasPrefix(name, ".") && name != ".")` redundancy

**File:** `internal/app/layout_test.go:97`
**Issue:** `filepath.WalkDir` invokes the callback with `d.Name()` returning the
directory's base name. For the root directory passed to `WalkDir(repoRoot, ...)`
the name is `filepath.Base(repoRoot)` — e.g. `"sops-tui"` — not `"."`. So the
guard `name != "."` is dead — `d.Name()` never returns `"."` for the repo root
or any child. Removing `&& name != "."` would make intent clearer without
behavioural change.

**Fix (cosmetic):**
```go
if name == "vendor" || name == ".git" || strings.HasPrefix(name, ".") {
    return filepath.SkipDir
}
```

---

## Focus-Area Checklist (from review request)

| # | Focus area | Result |
|---|------------|--------|
| 1 | `bodyDims` clamp correctness | Pass. No off-by-one, no underflow, `m.height==0` returns (0,0). |
| 2 | `TestBodyDimsMigration` robustness | Pass with three minor defensive tweaks (IN-03, IN-04). Self-match avoided (regex assembled from 3 string fragments across layout_test.go:82-85). Walker skips `.planning/`, `vendor/`, `.git/`, all dotfile dirs. Does NOT scan `.golden` files (extension filter). The carve-out correctly uses brace-depth tracking and handles the helper range via 1-indexed inclusive `[start, end]`. Manually verifying: adding a line `_ = m.height - statusBarHeight(m)` to `internal/app/model.go` outside the helper body would fail the test. |
| 3 | Outlier migration soundness | Pass. View() outlier preserves `statusBar` local, no dangling `statusBarH`. Pointer outlier uses `bodyDims(*m)`. TODO tag at 1839-1840 matches plan; `m.height - 4` at 1841 unchanged. |
| 4 | Golden harness determinism | Pass structural. IN-01 flags path-traversal via `name` (test-only severity). Normalisation handles CRLF→LF and trailing whitespace. |
| 5 | Resize test fixture leakage | Pass. Goldens at 40x12/80x24/120x40/200x60 contain only the "No SOPS files found" placeholder + status-bar. No cwd/env/time/PID leakage. `.gitattributes` pins LF line endings. `defaultEnv()` returns a fixed struct (GitAvailable: false → "no git" rendered). Note: resize_test.go uses direct `m.Update(msg)` not the `send()` helper from model_test.go — functionally equivalent, not a finding. |
| 6 | Threat-model coverage (T-06-01..T-06-11) | Pass. Zero production attack-surface change. New `internal/testutil/golden.go` is a test helper (no prod imports). `github.com/charmbracelet/x/ansi` v0.11.7 promoted to direct dep is used only for ANSI-stripping in tests. |
| 7 | Idiomatic Go | Pass. `for b.Loop()` used in `BenchmarkAppView` (bench_test.go:28). No `any` types. Error handling in the harness uses `t.Fatalf` consistently. `t.Helper()` called in all test helpers. `package app_test` used for the external-facing resize_test.go + golden_test.go; `package app` (internal test access) used for bench_test.go and layout_test.go as appropriate (they need `bodyDims`, `chromeHeight`, `crumbsHeight` which are unexported). CGO disabled. |
| 8 | Zero-behaviour-change invariant | Pass. Each migration is mechanical — swap a 4-line block for a 1-line helper call; pass-through semantics identical since Phase 6 stubs `chromeHeight`/`crumbsHeight` to 0. |

---

_Reviewed: 2026-04-23_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
