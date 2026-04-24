---
phase: 6
slug: layout-groundwork
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-23
---

# Phase 6 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Derived from `06-RESEARCH.md` §"Validation Architecture".

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing` (Go 1.26.2) + `stretchr/testify` v1.11.1 |
| **Config file** | none (stdlib `go test`) |
| **Quick run command** | `go test ./internal/app/... ./internal/testutil/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~3–5 s quick; ~30 s full suite |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/app/... ./internal/testutil/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd-verify-work`:** Full suite green + `BenchmarkAppView` baseline recorded + manual smoke at 40×12 and 200×60 (D-15)
- **Max feedback latency:** 5 seconds (quick), 30 seconds (full)

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 06-01-01 | 01 | 1 | UI-17 | — | `bodyDims` returns `(width, clamped-height)` | unit | `go test ./internal/app/... -run 'TestBodyDims$' -v` | ❌ W0 | ⬜ pending |
| 06-01-02 | 01 | 1 | UI-17 | — | `bodyDims` clamps negative heights to 0 | unit | `go test ./internal/app/... -run TestBodyDimsClampsAtZero -v` | ❌ W0 | ⬜ pending |
| 06-01-03 | 01 | 1 | UI-17 | — | `chromeHeight(m)` returns 0 (Phase 6 stub) | unit | `go test ./internal/app/... -run TestChromeHeightReturnsZero -v` | ❌ W0 | ⬜ pending |
| 06-01-04 | 01 | 1 | UI-17 | — | `crumbsHeight(m)` returns 0 (Phase 6 stub) | unit | `go test ./internal/app/... -run TestCrumbsHeightReturnsZero -v` | ❌ W0 | ⬜ pending |
| 06-01-05 | 01 | 1 | UI-18 | — | Banned regex `m\.height\s*-\s*statusBarHeight` absent outside `bodyDims` body | integration | `go test ./internal/app/... -run TestBodyDimsMigration -v` | ❌ W0 | ⬜ pending |
| 06-01-06 | 01 | 1 | UI-19 | — | `RequireGoldenStructure` write-on-`GOLDEN_UPDATE=1`, diff-on-unset | unit | `go test ./internal/testutil/... -run TestRequireGoldenStructure -v` | ❌ W0 | ⬜ pending |
| 06-01-07 | 01 | 1 | UI-19 | — | ANSI strip round-trips correctly | unit | `go test ./internal/testutil/... -run TestRequireGoldenStructure_ANSIStrip -v` | ❌ W0 | ⬜ pending |
| 06-01-08 | 01 | 1 | UI-19 | — | `RequireGoldenColors` no-ops on empty `wantColors` | unit | `go test ./internal/testutil/... -run TestRequireGoldenColors_Empty -v` | ❌ W0 | ⬜ pending |
| 06-01-09 | 01 | 1 | UI-19 | — | `normalise` trims trailing whitespace + normalises line endings | unit | `go test ./internal/testutil/... -run TestNormalise -v` | ❌ W0 | ⬜ pending |
| 06-01-10 | 01 | 1 | *(baseline)* | — | `BenchmarkAppView` records v1.0 cost at 200×60 | benchmark | `go test -bench=BenchmarkAppView -benchmem ./internal/app/...` | ❌ W0 | ⬜ pending |
| 06-02-01 | 02 | 2 | UI-17, UI-18 | — | All 15 SetSize call-sites + 2 outliers migrated to `bodyDims` | integration | `go test ./internal/app/... -run TestBodyDimsMigration -v` (flips red→green) | ❌ W0 | ⬜ pending |
| 06-02-02 | 02 | 2 | UI-17 | — | Resize goldens at 40×12 match v1.0 structural output | golden | `go test ./internal/app/... -run TestResize/40x12 -v` | ❌ W0 | ⬜ pending |
| 06-02-03 | 02 | 2 | UI-17 | — | Resize goldens at 80×24 match v1.0 structural output | golden | `go test ./internal/app/... -run TestResize/80x24 -v` | ❌ W0 | ⬜ pending |
| 06-02-04 | 02 | 2 | UI-17 | — | Resize goldens at 120×40 match v1.0 structural output | golden | `go test ./internal/app/... -run TestResize/120x40 -v` | ❌ W0 | ⬜ pending |
| 06-02-05 | 02 | 2 | UI-17 | — | Resize goldens at 200×60 match v1.0 structural output | golden | `go test ./internal/app/... -run TestResize/200x60 -v` | ❌ W0 | ⬜ pending |
| 06-02-06 | 02 | 2 | *(regression)* | — | All pre-existing v1.0 tests continue to pass | integration | `go test ./...` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Plan 1 delivers Wave 0. Plan 2 depends on all Wave 0 items landing first:

- [ ] `internal/app/model.go` — `bodyDims`, `chromeHeight` (returns 0), `crumbsHeight` (returns 0) definitions
- [ ] `internal/app/layout_test.go` — helper unit tests + `TestBodyDimsMigration` grep-gate
- [ ] `internal/app/bench_test.go` — `BenchmarkAppView` baseline
- [ ] `internal/testutil/golden.go` — `RequireGoldenStructure`, `RequireGoldenColors`, `normalise`
- [ ] `internal/testutil/golden_test.go` — round-trip + env-var-gated regen tests
- [ ] `go.mod` — promote `github.com/charmbracelet/x/ansi` from indirect to direct

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Positioning at 40×12 matches v1.0 exactly | UI-17 / Success #4 | Pitfall 1 warns goldens at 80×24 can hide the bug; visual scan of small terminal required | Run `sops-tui`, resize to 40×12, cycle through file list / detail / help / diff / metadata / history / health / recipient form views; no content painted under status bar |
| Positioning at 200×60 matches v1.0 exactly | UI-17 / Success #4 | Large-terminal bug class surfaces here per Pitfall 1 | Same cycle as above at 200×60 |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5 s (quick), < 30 s (full)
- [ ] `nyquist_compliant: true` set in frontmatter (set when plans are verified)

**Approval:** pending
