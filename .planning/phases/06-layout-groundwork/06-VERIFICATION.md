---
phase: 06-layout-groundwork
verified: 2026-04-24T10:15:00Z
status: human_needed
score: 5/5 must-haves verified (automated)
overrides_applied: 0
human_verification:
  - test: "Manual smoke at 40x12 cycling through non-empty sub-model paths"
    expected: "No content is ever painted under the status bar row at any size, across file list / detail / help / diff / metadata / history / health / recipient-form views"
    why_human: "Pitfall 1 explicitly warns goldens at 80x24 can hide the bug; D-15 scopes UAT to a real terminal because automated resize goldens only exercise the empty-state (NewAppModel with empty sopsYamlPath produces 'No SOPS files found'). Non-empty sub-model paths (populated file list, detail overlay rendering a parsed file, diff overlay with real entries, etc.) require a real .sops.yaml + age key in scope."
  - test: "Manual smoke at 200x60 cycling through non-empty sub-model paths"
    expected: "Same invariant — no content under the status bar at large terminal size; all 8 views render identically to v1.0"
    why_human: "Same rationale — large-terminal bug class surfaces per Pitfall 1. Currently the resize_200x60.golden only captures the empty-state path."
---

# Phase 6: Layout Groundwork Verification Report

**Phase Goal:** Height arithmetic is centralised behind a single helper and a regression-proof test harness is in place, so that every downstream chrome phase can size its children correctly without re-auditing 15 SetSize call-sites.
**Verified:** 2026-04-24T10:15:00Z
**Status:** human_needed
**Re-verification:** No — initial verification.

## Goal Achievement

All five ROADMAP success criteria are verified in the codebase. The automated contract is fully satisfied; only the manual UAT per D-15 remains outstanding (explicitly deferred by the executor and scoped to this verification stage).

### Observable Truths

| # | Truth (Success Criterion)                                                                                                                             | Status     | Evidence                                                                                                                                                             |
| - | ----------------------------------------------------------------------------------------------------------------------------------------------------- | ---------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | SC-1: `bodyDims(m) (w, h int)` helper in `internal/app/model.go` subtracts chrome + crumbs + status-bar; every `m.height - statusBarHeight(m)` call-site is migrated | VERIFIED   | `internal/app/model.go:1403` defines `func bodyDims(m AppModel) (w, h int)` with body `h = m.height - statusBarHeight(m) - chromeHeight(m) - crumbsHeight(m)` clamped to `>= 0`. 17 call-sites use it (15 value, 1 pointer-deref, 1 `_, mainH` outlier). |
| 2 | SC-2: CI grep-gate fails the build if `m.height - statusBarHeight` arithmetic appears outside `bodyDims`                                              | VERIFIED   | `TestBodyDimsMigration` passes under default `go test ./...`. Regression-proof: injected a probe file `internal/app/gate_probe.go` with the banned pattern, test FAILED with `UI-18 violation: internal/app/gate_probe.go:9 mainH := m.height - statusBarHeight(m)`. Cleanup restored PASS. |
| 3 | SC-3: Test harness exposes ANSI-stripped golden comparison; color presence can be asserted separately                                                 | VERIFIED   | `internal/testutil/golden.go:30` exports `RequireGoldenStructure`; line 61 exports `RequireGoldenColors`. Lines 90/101 expose `normalise` / `NormaliseForTest`. All 6 self-tests PASS (`TestRequireGoldenStructure_WritesOnUpdateEnv`, `_ComparesWhenUnset`, `_ANSIStrip`, `TestRequireGoldenColors_Empty`, `_Missing`, `TestNormalise`). |
| 4 | SC-4: Resize behaviour at 40x12, 80x24, 120x40, 200x60 verified by teatest snapshots; no content painted under the status bar                         | VERIFIED (automated); partial for non-empty paths (human_needed) | 4 goldens under `internal/app/testdata/` (plain UTF-8, no ANSI, no CRLF, no trailing WS). 4 `TestResize_<WxH>` tests PASS. Manual inspection of `resize_80x24.golden` confirms empty-state content + status bar pinned to last row. **Caveat:** goldens only exercise the empty-state path (no `.sops.yaml` in test cwd); non-empty sub-model paths at 40x12 and 200x60 require manual UAT per D-15. |
| 5 | SC-5: Phase ships zero visible change to end users; pure refactor + test infrastructure                                                               | VERIFIED   | Commit `79ec449` diff is purely mechanical — every site swaps 4-line `mainH := m.height - statusBarHeight(m); if mainH < 0 { mainH = 0 }` block for 1-line `w, h := bodyDims(m)`. No arity, ordering, or semantic change. `chromeHeight`/`crumbsHeight` stub to 0, so arithmetic is identical to pre-migration. `go test ./...` passes across all 9 packages with zero regressions. |

**Score:** 5/5 truths verified (automated contract); 1 truth (SC-4) has a scoped human_needed extension for non-empty paths.

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `internal/app/model.go` | 3 helpers (bodyDims, chromeHeight, crumbsHeight); 17 migrated call-sites; TODO(phase-7) at `boxHeight := m.height - 4` | VERIFIED | `bodyDims` at line 1403 with exact D-03 body; `chromeHeight` at 1415 returns 0 with `_ = m`; `crumbsHeight` at 1423 returns 0 with `_ = m`. 17 call-sites confirmed via `grep bodyDims\(` (15 `bodyDims(m)` + 1 `bodyDims(*m)` at line 1779 + 1 `_, mainH := bodyDims(m)` at line 1287). TODO comment at line 1839-1840 with exact canonical wording. |
| `internal/app/layout_test.go` | 5 unit/gate tests; no allowlist; no skip; runtime-assembled regex | VERIFIED | `TestBodyDims` (line 30), `TestBodyDimsClampsAtZero` (line 41), `TestChromeHeightReturnsZero` (line 51), `TestCrumbsHeightReturnsZero` (line 58), `TestBodyDimsMigration` (line 79). `grep -c plan2MigrationAllowlist` = 0, `grep -c t\.Skip` = 0. Regex assembled at line 82-85 from `regexPrefix + regexMinus + regexHelper`. |
| `internal/app/bench_test.go` | `BenchmarkAppView` at 200x60 using `b.Loop()` | VERIFIED | `for b.Loop()` at line 28. Sanity-run: `BenchmarkAppView-16 1 381598 ns/op 103128 B/op 912 allocs/op` (1x sample; ~348us/op per Plan 01 2s baseline). |
| `internal/app/resize_test.go` | 4 `TestResize_<WxH>` tests in `package app_test` using shared `defaultEnv()` | VERIFIED | `TestResize_40x12` (15), `_80x24` (27), `_120x40` (38), `_200x60` (49). `package app_test` — no duplicated `defaultEnv()`. Each calls `RequireGoldenStructure` + `RequireGoldenColors(..., nil)`. All 4 PASS. |
| `internal/app/testdata/resize_{40x12,80x24,120x40,200x60}.golden` | 4 ANSI-stripped UTF-8 fixtures | VERIFIED | All 4 present (215-264 bytes each). Binary inspection: `esc=False cr=False trailing_ws=False` for all. Manual inspection of `resize_80x24.golden` shows 24 rows: "No SOPS files found" top, padding middle, status bar last row. |
| `internal/testutil/golden.go` | `RequireGoldenStructure`, `RequireGoldenColors`, `normalise`, `NormaliseForTest` | VERIFIED | Signatures at lines 30, 61, 90, 101. Also exports `missingColors` (72) and `MissingColorsForTest` (84) — SUMMARY documents this as a deviation-auto-fix to work around `testing.TB` being sealed. Uses `ansi.Strip(output)` (x/ansi direct dep). No `any` type. |
| `internal/testutil/golden_test.go` | 6 self-tests | VERIFIED | All 6 PASS: `TestRequireGoldenStructure_WritesOnUpdateEnv`, `_ComparesWhenUnset`, `_ANSIStrip`, `TestRequireGoldenColors_Empty`, `_Missing`, `TestNormalise`. |
| `go.mod` | `github.com/charmbracelet/x/ansi v0.11.7` as direct dep | VERIFIED | Line 11 in the top-level `require` block; no `// indirect` marker. (SUMMARY notes `filippo.io/age v1.3.1` also moved to direct by `go mod tidy` — pre-existing drift, independently correct.) |
| `.gitattributes` | `*.golden text eol=lf -whitespace` | VERIFIED | One-line file at repo root: `*.golden text eol=lf -whitespace`. |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | -- | --- | ------ | ------- |
| 17 call-sites in `model.go` (WindowSizeMsg + on-demand creation) | `bodyDims` helper | `w, h := bodyDims(m)` / `bodyDims(*m)` / `_, mainH := bodyDims(m)` | WIRED | 17 occurrences confirmed via `grep bodyDims\(` in `model.go` (excluding the helper definition). Each flows `w`/`h` into `SetSize(w, h)` or `NewFooModel(..., w, h)`. |
| `TestBodyDimsMigration` | All `*.go` files under repo root | `filepath.WalkDir` + runtime-assembled regex + `bodyDims` body carve-out | WIRED | Walker skips `vendor/`, `.git/`, all dot-prefixed dirs. Carve-out uses brace-depth tracking via `findBodyDimsRange`. Regex assembled at compile-time from three string constants (self-match avoidance). |
| `TestResize_<WxH>` | `testutil.RequireGoldenStructure` | `m.View().Content` -> `ansi.Strip` -> `normalise` -> compare `testdata/resize_<WxH>.golden` | WIRED | All 4 tests import `testutil` and pass. `RequireGoldenColors(..., nil)` is called alongside (Phase 6 no-op scaffolding for Phase 7). |
| `RequireGoldenStructure` | `github.com/charmbracelet/x/ansi` | `ansi.Strip(output)` before normalise + compare | WIRED | Line 32 of `golden.go`: `stripped := normalise(ansi.Strip(output))`. x/ansi is now a direct dep. |
| `BenchmarkAppView` | `AppModel.View()` | `for b.Loop() { _ = m.View() }` | WIRED | Bench runs cleanly (`BenchmarkAppView-16 1 381598 ns/op`). |

### Data-Flow Trace (Level 4)

Not applicable in the dynamic-data sense. Phase 6 is a pure refactor + test infrastructure phase — no new rendered state, no new data source. The four resize goldens capture the deterministic empty-state `View()` output (no `.sops.yaml` discovered, so NewAppModel produces a synthetic fixed state). Determinism confirmed: regenerating under `GOLDEN_UPDATE=1` produces byte-identical fixtures run-to-run because `defaultEnv()` is a fixed struct and no timestamps/cwd/PIDs leak into `View()`.

### Behavioural Spot-Checks

| Behaviour | Command | Result | Status |
| --------- | ------- | ------ | ------ |
| All targeted phase tests pass | `go test ./internal/app/... -run 'TestBodyDims$|TestBodyDimsClampsAtZero|TestChromeHeightReturnsZero|TestCrumbsHeightReturnsZero|TestBodyDimsMigration|^TestResize_' -v` | 9 PASS, 0 FAIL, 0 SKIP | PASS |
| testutil self-tests pass | `go test ./internal/testutil/... -v` | 6 PASS | PASS |
| Full suite regression-free | `go test ./...` | 9 packages OK, 0 fail | PASS |
| Benchmark runs | `go test -bench=BenchmarkAppView -benchmem -run='^$' -benchtime=1x ./internal/app/...` | `BenchmarkAppView-16 1 381598 ns/op 103128 B/op 912 allocs/op` | PASS |
| Build clean | `go build ./...` | exit 0, no output | PASS |
| Vet clean | `go vet ./...` | exit 0, no output | PASS |
| **Grep-gate regression probe** | Injected `gate_probe.go` with banned pattern; ran `go test -run TestBodyDimsMigration`; cleanup; re-ran | Probe: FAIL with `UI-18 violation: internal/app/gate_probe.go:9 mainH := m.height - statusBarHeight(m)` Cleanup: PASS | PASS (gate is regression-proof, not a no-op) |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ----------- | ----------- | ------ | -------- |
| UI-17 | 06-01-PLAN (helper), 06-02-PLAN (migration) | `bodyDims(m) (w, h int)` is the single source of truth for body-region arithmetic; all `m.height - statusBarHeight(m)` call-sites migrate to it | SATISFIED | Helper defined at `model.go:1403`; 17 call-sites migrated (15 value, 1 pointer, 1 View outlier); only remaining occurrence of the banned regex is inside `bodyDims` itself (confirmed by `grep` = 1). Resize tests at 4 sizes prove end-to-end. |
| UI-18 | 06-01-PLAN (gate + allowlist), 06-02-PLAN (allowlist deletion) | CI grep-gate prevents reintroduction of the banned `m.height - statusBarHeight(m)` pattern outside `bodyDims` | SATISFIED | `TestBodyDimsMigration` runs under default `go test ./...` (no skip, no env gate, no allowlist). Regression-proof: probe injected a violation, test FAILED immediately with named violation. Cleanup: test passes. Plan-2 allowlist deleted atomically with migration (commit `79ec449`): `grep -c plan2MigrationAllowlist internal/app/layout_test.go` = 0. |
| UI-19 | 06-01-PLAN (harness), 06-02-PLAN (exercised via resize tests) | teatest harness helper strips ANSI escape sequences for structural golden comparison; color presence asserted separately | SATISFIED | `internal/testutil/golden.go` exports `RequireGoldenStructure` (strips ANSI via `ansi.Strip`, normalises CRLF→LF, right-trims per-line whitespace, compares against `testdata/<name>.golden`, supports `GOLDEN_UPDATE=1` regen) and `RequireGoldenColors` (raw-byte substring match, empty `wantColors` no-op for Phase 6). 6 self-tests pass. 4 resize tests exercise both helpers end-to-end. |

No orphaned requirements for Phase 6 — REQUIREMENTS.md lists only UI-17, UI-18, UI-19 under "Layout Safety (groundwork)" and all three are covered by the plan frontmatters.

### Locked Decisions (D-01 through D-15) — Honour Check

| Decision | Expectation | Status |
| -------- | ----------- | ------ |
| D-01 | Signature `bodyDims(m AppModel) (w, h int)`, free function | HONOURED (model.go:1403) |
| D-02 | Placement immediately after `statusBarHeight` | HONOURED (statusBarHeight at 1394, bodyDims at 1403) |
| D-03 | Body `h = m.height - statusBarHeight(m) - chromeHeight(m) - crumbsHeight(m)` clamped >= 0 | HONOURED (lines 1405-1408) |
| D-04 | `chromeHeight` / `crumbsHeight` stubs return 0 | HONOURED (lines 1415-1418, 1423-1426, both use `_ = m`) |
| D-05 | Grep-gate as a Go test in `internal/app` | HONOURED (`TestBodyDimsMigration` in `internal/app/layout_test.go`) |
| D-06 | Banned regex `m\.height\s*-\s*statusBarHeight` across all `*.go` | HONOURED (regex assembled at runtime; WalkDir covers repo) |
| D-07 | TODO tag at `m.height - 4` site; magic number preserved | HONOURED (model.go:1839-1840 comment; line 1841 expression unchanged) |
| D-08 | Hand-rolled harness using `x/ansi.Strip`; direct dep | HONOURED (go.mod:11 direct; golden.go uses `ansi.Strip`) |
| D-09 | `RequireGoldenStructure` + `RequireGoldenColors` in `internal/testutil` | HONOURED (both exported with concrete types) |
| D-10 | `GOLDEN_UPDATE=1` env var for regeneration | HONOURED (golden.go line 35: `os.Getenv("GOLDEN_UPDATE") == "1"`) |
| D-11 | Per-package `testdata/` directories | HONOURED (`internal/app/testdata/resize_<WxH>.golden`) |
| D-12 | `BenchmarkAppView` using `b.Loop()` at 200x60 | HONOURED (bench_test.go:28) |
| D-13 | Exactly 2 plans | HONOURED (06-01-PLAN.md + 06-02-PLAN.md) |
| D-14 | Migration + allowlist deletion is ONE atomic commit | HONOURED (commit `79ec449`: 2 files changed, 50 insertions, 127 deletions; model.go + layout_test.go atomic) |
| D-15 | Manual UAT at 40x12 and 200x60 | PENDING — deferred to this verification; see Human Verification section |

### Anti-Patterns Found

Scanned files modified in this phase for stub/placeholder/TODO patterns. All findings are intentional or informational.

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| `internal/app/model.go` | 1839-1840 | `TODO(phase-7): replace magic -4 ...` | Info | **Intentional** — per D-07, the `m.height - 4` magic number is deliberately deferred to Phase 7. The TODO tag is a planned deferral, not a stub. |
| `internal/app/model.go` | 1416, 1424 | `chromeHeight` / `crumbsHeight` stubs return 0 with `_ = m` | Info | **Intentional** — per D-04, these are flip-point stubs for Phase 7/8. The `_ = m` idiom is load-bearing (silences unused-parameter warning while preserving the AppModel signature). |
| `internal/testutil/golden.go` | 74-80 | `RequireGoldenColors` is a no-op when `wantColors` is empty | Info | **Intentional** — per D-09, Phase 6 scaffolds colour assertions for Phase 7 to populate. Resize tests explicitly call `RequireGoldenColors(..., nil)`. |

None of these represent stubs that prevent goal achievement; all three are the phase's explicit deferral pattern.

### Code Review (06-REVIEW.md) — Carried Forward

The phase code review (`06-REVIEW.md`) found **0 critical, 0 warning, 4 info** — no blockers. Findings for the record (not gating this verification):

- **IN-01**: `RequireGoldenStructure` name parameter permits directory traversal on `GOLDEN_UPDATE=1` writes. Test-only helper, no untrusted input; reviewer classified as Info.
- **IN-02**: `View()` now renders the status bar twice per frame via `bodyDims(m)` callback (once for the local `statusBar`, once inside `statusBarHeight`). Minor hot-path regression captured in `BenchmarkAppView`; correctness-neutral because `StatusBarModel.View` is a pure function. Reviewer recommendation: defer optimisation to Phase 7's chrome refactor.
- **IN-03**: `TestBodyDimsMigration` walker does not explicitly skip `testdata/` directories. Today harmless (fixture files fail the `.go` suffix filter); defensive tweak for future contributors.
- **IN-04**: Cosmetic — the `name != "."` guard in the walker's dir-skip is dead for `filepath.WalkDir`'s callback semantics.

All four are Info-severity and do not block phase closure. They are candidates for opportunistic fixes in Phase 7 or a future housekeeping pass.

### Human Verification Required

Automated verification is complete; the following manual UAT is required per D-15 of `06-CONTEXT.md` and the "Pending Verification" section of `06-02-SUMMARY.md`:

#### 1. Manual smoke at 40x12 — non-empty sub-model paths

**Test:** Build `/tmp/sops-tui` with `go build -o /tmp/sops-tui ./...`. Open a terminal in a directory containing a real `.sops.yaml` and at least one SOPS-encrypted file with an age recipient for which the host has the private key. Resize to **40 cols x 12 rows**. Run `/tmp/sops-tui`. Cycle through: file list -> a file's detail view -> help (`?`) -> diff -> metadata (`i`) -> history (`h`) -> health -> recipient form (`a`).
**Expected:** No content is ever painted under the status bar row at any size. Status bar is pinned to the last row. All 8 views render identically to v1.0 (visual diff — compare against pre-Phase-6 screenshots if available).
**Why human:** Pitfall 1 explicitly warns that goldens at 80x24 can hide the bug class. The automated resize tests only cover the empty-state path (`NewAppModel(defaultEnv(), "")` yields "No SOPS files found"). Non-empty paths (populated file list, detail with parsed YAML/JSON nodes, diff with real entries, metadata panel, git history, health scan results, recipient form) require a real `.sops.yaml` + age key in scope — out of scope for the executor's automated harness.

#### 2. Manual smoke at 200x60 — non-empty sub-model paths

**Test:** Same sequence as test 1 at **200 cols x 60 rows**.
**Expected:** Same invariant — no content ever painted under the status bar row. All 8 views render identically to v1.0 at large terminal size.
**Why human:** Pitfall 1 specifically flags "the bug manifests at 120x40 when hidden rows below the status bar start receiving content" — a size-dependent regression class. Large-terminal coverage requires visual inspection.

### Gaps Summary

No gaps blocking the phase goal. All five ROADMAP success criteria are satisfied by the codebase and pass end-to-end:

- `bodyDims` exists with the exact D-01/D-03 signature and body, correctly clamped.
- All 17 call-sites migrated; the only remaining `m.height - statusBarHeight` occurrence in the repo is inside the helper itself (grep count = 1).
- The grep-gate is LIVE (no skip, no allowlist, no env gate) and regression-proof (verified by injecting a probe file that produced a named violation, then cleaning up and observing re-PASS).
- 4 resize goldens exist as plain UTF-8 (no ANSI, no CRLF, no trailing whitespace), and the 4 `TestResize_<WxH>` tests pass.
- Zero behaviour change — the commit diff is purely mechanical, stubs flip 0 to 0, benchmark runs cleanly.

The single outstanding verification is the explicit D-15 manual UAT at 40x12 and 200x60 cycling through the 8 non-empty views, which is routed to human testing per the "Human Verification Required" section above. This is not a gap — it is the D-15 acceptance plan.

---

_Verified: 2026-04-24T10:15:00Z_
_Verifier: Claude (gsd-verifier)_
