# Phase 6: Layout Groundwork - Context

**Gathered:** 2026-04-23
**Status:** Ready for planning

<domain>
## Phase Boundary

Centralise body-size arithmetic behind a single `bodyDims(m) (w, h int)` helper and stand up an ANSI-stripped teatest harness. All 15 `m.height - statusBarHeight(m)` call-sites in `internal/app/model.go` (plus the outliers at lines 1333 and 1799) migrate to the new helper before any chrome rendering is attempted. A grep-gate test prevents regression. Zero visible change to end users — this is pure refactor + test infrastructure that unlocks Phase 7's chrome skeleton without forcing a second SetSize audit pass (Pitfall 1).

**In scope:** layout helper, grep-gate, ANSI-stripped golden harness, resize-safety snapshots, v1.0 BenchmarkAppView baseline.
**Out of scope:** any chrome rendering (logo, menu, info panel, titled borders), `lipgloss/v2/table` adoption, `Hints()` interface (Phase 7+).

</domain>

<decisions>
## Implementation Decisions

### bodyDims helper shape
- **D-01:** Signature is `bodyDims(m AppModel) (w, h int)` — free function, returns both dimensions symmetrically even though only height differs today. Future-proofs for Phase 7 titled-border width padding without churning call-sites.
- **D-02:** Lives in `internal/app/model.go` beside the existing `statusBarHeight(m AppModel) int` helper (currently at model.go:1443). Single file for all layout arithmetic.
- **D-03:** Body implementation: `w = m.width; h = max(0, m.height - statusBarHeight(m) - chromeHeight(m) - crumbsHeight(m))`. Clamping to `>= 0` prevents `bubbles/v2/list` panic/blank-render on shorter-than-chrome terminals (Pitfall 1 warning sign).
- **D-04:** `chromeHeight(m AppModel) int` and `crumbsHeight(m AppModel) int` are introduced as stubs in Phase 6, each returning `0`. Phase 7 flips `chromeHeight`'s body to the real value; Phase 8 flips `crumbsHeight`. No call-site change needed when those land — that is the whole point of the stubs (avoids the second SetSize audit pass that Pitfall 1 warns about).

### Grep-gate
- **D-05:** Implemented as a self-contained Go test (`TestBodyDimsMigration` in `internal/app`). Reads all `*.go` files via `os.ReadFile`/`filepath.Walk`, fails if the banned regex appears outside the `bodyDims` function body. Runs in `go test ./...` with zero new CI infrastructure. No GitHub Actions workflow bootstrapped in Phase 6 — CI hosting is a separate question for a later milestone.
- **D-06:** Banned regex: `m\.height\s*-\s*statusBarHeight`. Applies to the entire repo (all `*.go` files). Only permitted inside the `bodyDims` function body — the test parses `model.go` and carves out the helper's line range before scanning.
- **D-07:** Outlier handling — `model.go:1333` (inside `View()`, currently uses `statusBarH` local) and `model.go:1799` (helper path) are migrated to `bodyDims` in Plan 2 alongside the 15 SetSize sites. `model.go:1862` uses `m.height - 4` (magic constant, separate concern) — tag with `// TODO(phase-7): replace magic -4 with named modal-frame constant or bodyDims usage` and defer.

### Golden file harness
- **D-08:** Hand-rolled ANSI strip + string compare. `charmbracelet/x/ansi` (currently indirect dep via lipgloss) is promoted to a direct dep in `go.mod`. ~30 LOC total. No external golden library (goldie, teatest.RequireEqualOutput) added.
- **D-09:** New package `internal/testutil` with `golden.go`:
  - `RequireGoldenStructure(t *testing.T, name string, output string)` — ANSI-strips `output`, compares against `testdata/<name>.golden` file, fails test on mismatch.
  - `RequireGoldenColors(t *testing.T, name string, output string, wantColors []string)` — asserts each `wantColors` byte sequence appears in the raw (non-stripped) `output`. Empty `wantColors` in Phase 6; populated in Phase 7.
  Rationale: Pitfall 8 requires the split between structural and color assertions from day one so Phase 7 inherits the pattern.
- **D-10:** Update mechanism: `GOLDEN_UPDATE=1` env var. Helper writes the file when set; diff-fails when unset. Chosen over `-update` flag for zero TestMain scaffolding. Intentional friction per Pitfall 8 ("devs start running -update reflexively").
- **D-11:** Test fixtures live in per-package `testdata/` directories (Go convention). Phase 6 fixtures: `internal/app/testdata/resize_40x12.golden`, `resize_80x24.golden`, `resize_120x40.golden`, `resize_200x60.golden`.

### Performance baseline
- **D-12:** `BenchmarkAppView` added in `internal/app` in Plan 1. Records v1.0 render cost at 200×60 today, before any chrome lands. Phase 7 reviewers get a concrete "before" number to compare the chrome's ≤ 50 µs/op target against. ~20 LOC — no gating on absolute value in Phase 6.

### Plan split
- **D-13:** Two plans per ROADMAP.md:
  - **Plan 1 — Layout helper + test infrastructure:** `bodyDims`/`chromeHeight`/`crumbsHeight` helpers with clamp logic, unit tests for helpers, `TestBodyDimsMigration` grep-gate test, `internal/testutil/golden.go` with `RequireGoldenStructure` and `RequireGoldenColors`, `BenchmarkAppView` baseline. No production-code behaviour changes in this plan.
  - **Plan 2 — Migration + resize goldens:** Single atomic commit migrating all 15 SetSize propagation sites (316, 349, 377, 485, 502, 567, 631, 724, 761, 846, 924, 1005, 1089, 1110, 1250) plus outliers 1333 and 1799 to `bodyDims`. Write resize goldens at 40×12, 80×24, 120×40, 200×60 asserting status bar occupies the last N rows and body occupies the rest. TODO comment at line 1862.
- **D-14:** Migration is one commit, not per-site or per-state. Rewrite is mechanical and identical 17 times; reviewer confirms the pattern once.
- **D-15:** Plan 1 UAT: `go test ./...` passes (grep-gate included), benchmark runs. Plan 2 UAT: 4-size goldens pass AND manual `sops-tui` smoke at 40×12 and 200×60 showing the file list / detail / help / diff / metadata / history / health / recipient form views identical to v1.0. Manual smoke is required because Pitfall 1 notes "teatest snapshots at 80×24 look fine; the bug manifests at 120×40 when hidden rows below the status bar start receiving content" — goldens alone can miss positioning bugs.

### Claude's Discretion
- Exact regex compilation / line-range carve-out implementation for `TestBodyDimsMigration`.
- File walk strategy (single `filepath.WalkDir` vs `go/ast` parse).
- `internal/testutil/golden.go` error message format and diff output style.
- Whether `RequireGoldenStructure` normalises trailing whitespace/line endings (Pitfall 8 recommends; implementation is Claude's call).
- `BenchmarkAppView` table size (200×60 fixed vs parametrised).
- Exact goldens file naming convention beyond `resize_<WxH>.golden`.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Decision Docs
- `.planning/ROADMAP.md` §"Phase 6: Layout Groundwork" — Goal, 5 success criteria, 2-plan budget, UI-17/UI-18/UI-19 requirements, scope boundary (no chrome rendering in this phase)
- `.planning/REQUIREMENTS.md` §"Layout Safety (groundwork)" — UI-17 (`bodyDims` as single source of truth), UI-18 (CI grep-gate), UI-19 (ANSI-stripped teatest harness)
- `.planning/PROJECT.md` — Project constraints (Go + Bubble Tea, subprocess sops, v1.0 functional core must not regress)

### Research — v1.1 milestone (MUST READ)
- `.planning/research/PITFALLS.md` §"Pitfall 1: Chrome Height Subtraction…" — lists all 15 call-sites by line number (316, 321–328, 349, 353, 377, 385, 489, 502, 509, 567, 574, 631, 635, 724, 728, 761, 765), prescribes `bodyHeight(m)` helper + grep-gate + multi-size resize test
- `.planning/research/PITFALLS.md` §"Pitfall 8: Teatest Golden Files…" — prescribes ANSI-stripped structural goldens + separate color assertions; rejects bundled ANSI-bytes-and-all goldens
- `.planning/research/PITFALLS.md` §"Pitfall-to-Phase Mapping" — confirms Pitfall 1 + Pitfall 8 are Phase 6's responsibility
- `.planning/research/PITFALLS.md` §"Looks Done But Isn't" checklist — Phase 6 items: "Chrome height subtraction: All 16+ SetSize call sites pass bodyHeight(m)", "Goldens comparing ANSI-stripped text"
- `.planning/research/SUMMARY.md` §"Phase 6: Layout-arithmetic groundwork" — delivers list (helpers, 15-site migration, grep-gate, ANSI helper, styles.go stubs, keys/hints.go struct), rationale ("Pitfall 1 is highest-severity regression risk; must land before any chrome rendering to avoid two audit passes on 15 SetSize sites")
- `.planning/research/ARCHITECTURE.md` §"Phase Sequencing" — chrome sub-models depend on layout arithmetic being centralised first
- `.planning/research/STACK.md` §"charm.land/lipgloss/v2", §"charmbracelet/x/ansi" — dep inventory; `x/ansi` already indirect

### Prior Phase Context (preferences to carry forward)
- `.planning/phases/01-foundation/01-CONTEXT.md` — Status bar pattern, design system conventions, AppModel root composition
- `.planning/phases/05-power-features/05-CONTEXT.md` — sessionState enum overlay pattern, async msg propagation, SetSize discipline already established

### Technology Stack
- `CLAUDE.md` §Technology Stack — `charm.land/bubbletea/v2` v2.0.4, `lipgloss/v2` v2.0.3, `bubbles/v2` v2.1.0, `charmbracelet/x/exp/teatest` available but not yet imported
- `CLAUDE.md` §Bubbletea v2 migration — `View()` returns `tea.View`, `tea.KeyPressMsg` interface vs v1 struct
- `go.mod` — current direct/indirect dep list; `x/ansi` is indirect via lipgloss

### External
- `charmbracelet/x/ansi` `Strip(s string) string` — https://pkg.go.dev/github.com/charmbracelet/x/ansi — ANSI escape removal helper
- `charm.land/bubbles/v2/list` — pagination depends on height passed via `SetSize`; documented on pkg.go.dev

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/app/model.go:1443` — `statusBarHeight(m AppModel) int` — existing helper, direct analogue for `bodyDims` naming and location.
- `internal/app/model.go:313-328` — `WindowSizeMsg` handler with 8-line `SetSize` fan-out to fileList/detail/help/metadata/diff/history/health/recipientForm — the first migration target.
- `internal/ui/statusbar.go` — `statusBarHeight` consumer; breadcrumb rendering (still owned by status bar in Phase 6; moves to crumbs row in Phase 8).
- `charmbracelet/x/ansi` — already pulled in transitively as `github.com/charmbracelet/x/ansi v0.11.7`. `ansi.Strip(s)` is the only function needed; promote to direct dep with no new version work.
- `testing` package — `testing.B` for `BenchmarkAppView`; `t.Helper()` + `t.Errorf` for golden diff output.

### Established Patterns
- **Helper functions at end of model.go** — `statusBarHeight` demonstrates the convention. `bodyDims`, `chromeHeight`, `crumbsHeight` follow immediately below.
- **`SetSize(width, height)` contract on sub-models** — every sub-model has this method; the migration never changes call sites' arity, only the value computation.
- **`WindowSizeMsg` is the sole propagation trigger** — Phase 6 preserves this. No new resize pathways; `bodyDims` is called in the same places the old expression was.
- **Tests live beside implementation** — `model_test.go`, `filelist_test.go`, etc. Phase 6 adds `layout_test.go` or appends to `model_test.go` (Claude's discretion).
- **No existing CI** — `.github/` absent. Phase 6's grep-gate is a Go test specifically so it runs locally with `go test` and will run in any future CI without infra changes.

### Integration Points
- All 15 `SetSize` call-sites receive `mainH := m.height - statusBarHeight(m); sub.SetSize(m.width, mainH)` → become `w, h := bodyDims(m); sub.SetSize(w, h)`.
- `model.go:1333` (View body) uses `mainH := m.height - statusBarH` where `statusBarH` is a local — rewrite to use `bodyDims(m)`.
- `model.go:1799` (helper pathway) — same pattern, migrate.
- `model.go:1862` — `boxHeight := m.height - 4`. NOT migrated; tagged with `// TODO(phase-7): …` comment.
- `go.mod` — `github.com/charmbracelet/x/ansi` moves from indirect to direct `require` block.
- `internal/testutil/` — new package. Importable by `internal/app/*_test.go` and future `internal/ui/*_test.go`.

</code_context>

<specifics>
## Specific Ideas

- Pitfall 1 is explicitly the #1 regression risk for v1.1. Phase 6's entire existence is about preventing the "content painted under the header" bug class before chrome lands — not about shipping any user-visible feature.
- Stubs-returning-zero pattern for `chromeHeight`/`crumbsHeight` is the key structural decision: Phase 7 and Phase 8 land chrome by flipping two stub bodies, not by auditing 17 call-sites a second time.
- `charmbracelet/x/ansi.Strip` rather than a hand-written regex because it already covers CSI, OSC, DCS, and malformed-escape recovery that a 5-line regex would miss. The direct-dep promotion is a one-line go.mod change.
- Manual smoke test at 40×12 and 200×60 is explicit in Plan 2's UAT because Pitfall 1 warns goldens at 80×24 can hide the bug. Not optional.
- `BenchmarkAppView` is a cheap investment in Phase 6 (<20 LOC) that gives Phase 7 a real baseline number to compare its ≤ 50 µs/op target against — otherwise the target is a guess.
- `GOLDEN_UPDATE=1` (env var) over `-update` (flag) because the env var has just enough friction that devs don't reflexively regenerate goldens without reading the diff.

</specifics>

<deferred>
## Deferred Ideas

- **model.go:1862 `m.height - 4` magic number** — separate concern from statusBarHeight arithmetic. Tag with `// TODO(phase-7): …` in Phase 6; address when modal-frame chrome lands (Phase 7 or 8). Not a Phase 6 regression — v1.0 already shipped this.
- **GitHub Actions CI bootstrap** — `.github/workflows/` absent. Phase 6's grep-gate is a Go test so it runs without CI; a future phase can add the CI workflow and it will pick up the test automatically. Not Phase 6 scope.
- **Mode 2026 synchronized output / alt-screen fill-frame** — Pitfall 10 concern, belongs to Phase 11 terminal compat sweep.
- **`Hints() []MenuHint` interface + `HintsFromBindings` helper** — research SUMMARY's Phase 6 deliverable list mentions this, but ROADMAP.md scopes it to Phase 9 (Keybinding Discoverability). Roadmap wins; deferred to Phase 9.
- **`styles.go` stubs for chrome styles** — research SUMMARY suggests adding empty style vars in Phase 6. ROADMAP.md scopes them to Phase 7. Roadmap wins; deferred to Phase 7.
- **golangci-lint adoption** — considered as grep-gate host; rejected for Phase 6 in favour of a Go test. Can be introduced later without invalidating the test.
- **Responsive narrow-terminal column hiding** — Phase 10 concern.

</deferred>

---

*Phase: 06-layout-groundwork*
*Context gathered: 2026-04-23*
