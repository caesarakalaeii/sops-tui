# Phase 6: Layout Groundwork - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in 06-CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-23
**Phase:** 06-layout-groundwork
**Areas discussed:** bodyDims forward-compat, CI grep-gate hosting, Golden file framework, Plan split shape

---

## bodyDims forward-compat

| Option | Description | Selected |
|--------|-------------|----------|
| Stubs returning 0 | bodyDims subtracts statusBarHeight + chromeHeight + crumbsHeight; new stubs return 0 in Phase 6; Phase 7 flips stub bodies without SetSize audit | ✓ |
| Match current behaviour exactly | Only subtract statusBarHeight today; chromeHeight/crumbsHeight added in Phase 7 — reintroduces Pitfall 1 audit risk | |
| chromeHeight stub only | Add chromeHeight stub now; defer crumbsHeight to Phase 8 | |

**User's choice:** Stubs returning 0
**Notes:** Matches Pitfall 1's prescription. Two stubs = no second audit pass when chrome lands in Phase 7 and crumbs in Phase 8.

| Option | Description | Selected |
|--------|-------------|----------|
| internal/app/model.go beside statusBarHeight | Adjacent to existing helper at model.go:1443 | ✓ |
| New file internal/app/layout.go | Dedicated layout.go with all four helpers | |
| New package internal/layout | Overkill — import cycle risk | |

**User's choice:** internal/app/model.go beside statusBarHeight

| Option | Description | Selected |
|--------|-------------|----------|
| bodyDims(m) returning (w, h int) | Symmetric dims; matches research SUMMARY + ROADMAP SC-1 wording | ✓ |
| bodyHeight(m) int only | Height-only; width stays m.width | |
| Method (m AppModel) bodyDims() | Method receiver | |

**User's choice:** bodyDims(m) returning (w, h int)

| Option | Description | Selected |
|--------|-------------|----------|
| Clamp to >= 0 | max(0, ...) prevents negative-height crashes per Pitfall 1 | ✓ |
| Return raw negative value | No clamp; trust callers | |
| Clamp and flash a warning | Clamp + "terminal too small" status flash | |

**User's choice:** Clamp to >= 0
**Notes:** Pitfall 1 explicitly warns that bubbles/v2/list panics on negative height.

---

## CI grep-gate hosting

| Option | Description | Selected |
|--------|-------------|----------|
| Self-contained Go test | TestBodyDimsMigration reads model.go, fails on banned pattern. Zero CI infra. | ✓ |
| GitHub Actions workflow | Bootstraps .github/workflows/ for the whole project | |
| Makefile + pre-commit hook | Local-only enforcement via make check | |
| golangci-lint custom rule | Custom forbidigo/gocritic rule | |

**User's choice:** Self-contained Go test
**Notes:** No CI exists yet. Keeping the gate as a Go test means it works today and will work in any future CI without infra changes.

| Option | Description | Selected |
|--------|-------------|----------|
| Whole repo except bodyDims body | All *.go; only permitted inside bodyDims | ✓ |
| internal/app/ only | Narrow scope to migration site | |
| internal/ recursively | Middle ground | |

**User's choice:** Whole repo except bodyDims body

| Option | Description | Selected |
|--------|-------------|----------|
| m\.height\s*-\s*statusBarHeight | Matches the exact banned expression from Pitfall 1 | ✓ |
| statusBarHeight( outside bodyDims | Too broad — breaks legitimate status bar render call | |
| Ban all three (chrome/crumbs too) | Pre-emptive ban on chrome/crumbs patterns | |

**User's choice:** m\.height\s*-\s*statusBarHeight

| Option | Description | Selected |
|--------|-------------|----------|
| Migrate 1333 + 1799; leave 1862 with TODO | Migrate the two statusBarHeight outliers; defer the -4 magic number | ✓ |
| Migrate all three to bodyDims | One clean sweep including 1862's -4 | |
| Migrate only the 15 SetSize sites | Narrow: Pitfall 1 names 15 sites only | |

**User's choice:** Migrate 1333 + 1799; leave 1862 with TODO
**Notes:** Deliberate scope control — 1862's `-4` is a modal frame constant, a different concern from statusBarHeight arithmetic. Deferred to a later phase with a TODO comment.

---

## Golden file framework

| Option | Description | Selected |
|--------|-------------|----------|
| Hand-rolled ANSI strip + string compare | charmbracelet/x/ansi.Strip promoted from indirect to direct; ~30 LOC helper | ✓ |
| sebdah/goldie v2 | External dep, last release 2023, -update flag builtin | |
| teatest.RequireEqualOutput | Charm ecosystem, overkill for bare-string snapshots | |

**User's choice:** Hand-rolled ANSI strip + string compare

| Option | Description | Selected |
|--------|-------------|----------|
| internal/testutil/golden.go + testdata/ per package | New package, importable by any test | ✓ |
| Helper as unexported funcs in internal/ui/ui_test.go | No new package | |
| Separate internal/app/testharness package | Phase-6-specific package | |

**User's choice:** internal/testutil/golden.go + testdata/ per package

| Option | Description | Selected |
|--------|-------------|----------|
| GOLDEN_UPDATE=1 env var | os.Getenv check in helper; intentional friction | ✓ |
| -update flag via flag.Bool | TestMain registration required | |
| Both (env var as secondary) | Two knobs | |

**User's choice:** GOLDEN_UPDATE=1 env var
**Notes:** Pitfall 8 warns against reflexive regenerates — env var friction is the guardrail.

| Option | Description | Selected |
|--------|-------------|----------|
| Split helper now, empty color assertions | RequireGoldenStructure + RequireGoldenColors both exist in Phase 6; Phase 7 inherits the pattern | ✓ |
| Structure helper only; color in Phase 7 | YAGNI — no colors to assert yet | |
| Single helper with ANSI bytes | Rejected by Pitfall 8 | |

**User's choice:** Split helper now, empty color assertions

---

## Plan split shape

| Option | Description | Selected |
|--------|-------------|----------|
| Plan 1: harness + helper; Plan 2: migration + goldens | Infra first, mechanical refactor second | ✓ |
| Plan 1: helper + migration; Plan 2: harness + goldens | Atomic refactor first, test infra after | |
| Plan 1: 8 sites; Plan 2: 7 sites + goldens + gate | Split migration itself across plans | |

**User's choice:** Plan 1: harness + helper; Plan 2: migration + goldens

| Option | Description | Selected |
|--------|-------------|----------|
| Single commit for all 15 sites | One atomic mechanical rewrite | ✓ |
| Per-sessionState chunks | Multiple commits grouped by state | |
| One commit per call-site | 15 commits, maximum bisect granularity | |

**User's choice:** Single commit for all 15 sites

| Option | Description | Selected |
|--------|-------------|----------|
| Plan 1 UAT: tests + grep-gate green; Plan 2 UAT: goldens + manual smoke | Per-plan verification with extremes smoke-tested | ✓ |
| Combined UAT at phase end only | One UAT for both plans | |
| Automated-only (no manual smoke) | Trust the goldens | |

**User's choice:** Plan 1 UAT: tests + grep-gate green; Plan 2 UAT: 4-size goldens pass + manual sops-tui smoke

| Option | Description | Selected |
|--------|-------------|----------|
| Yes, stub benchmark in Phase 6 | v1.0 baseline for Phase 7 / 11 gates | ✓ |
| No, defer to Phase 7 | Phase 7 adds benchmark + chrome + gate together | |
| Baseline in a Phase 6.1 | Decimal phase for perf tooling | |

**User's choice:** Yes, stub benchmark in Phase 6

---

## Claude's Discretion

- Exact regex compilation / line-range carve-out implementation for TestBodyDimsMigration
- File walk strategy (single filepath.WalkDir vs go/ast parse)
- internal/testutil/golden.go error message format and diff output style
- Whether RequireGoldenStructure normalises trailing whitespace / line endings
- BenchmarkAppView table size (200×60 fixed vs parametrised)
- Exact goldens file naming convention beyond `resize_<WxH>.golden`

## Deferred Ideas

- model.go:1862 magic number `-4` — TODO tag, address in later phase
- GitHub Actions CI bootstrap — no .github/ needed for Phase 6
- Mode 2026 / alt-screen fill-frame — Phase 11
- Hints() interface + HintsFromBindings helper — Phase 9 (not Phase 6 per ROADMAP)
- styles.go chrome-style stubs — Phase 7 (not Phase 6 per ROADMAP)
- golangci-lint adoption — later milestone
- Responsive narrow-terminal column hiding — Phase 10
