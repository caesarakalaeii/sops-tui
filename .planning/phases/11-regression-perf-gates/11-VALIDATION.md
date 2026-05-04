---
phase: 11
slug: regression-perf-gates
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-05-04
---

# Phase 11 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing` + `testing.Benchmark` (no `teatest`; project uses Update-loop pattern) |
| **Config file** | `go.mod` (Go 1.26.x) |
| **Quick run command** | `go test ./internal/app/ -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Bench-only** | `go test ./internal/app/ -bench=BenchmarkAppView -benchmem -count=3` |
| **Estimated runtime** | ~5s (per-package) / ~30s (full suite) / ~10s (bench × 3) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/app/ -count=1`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green; bench under 50,000 ns/op
- **Max feedback latency:** 30 seconds (full suite)

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 11-01-01 | 01 | 1 | UI-21 | — | N/A | unit | `go test ./internal/app/ -run TestChromeCache_HitRateAtSteadyState -count=1` | ❌ W0 | ⬜ pending |
| 11-01-02 | 01 | 1 | UI-21 | — | N/A | bench | `go test ./internal/app/ -run TestBenchmarkAppView_UnderBudget -count=1` | ✅ | ⬜ pending |
| 11-01-03 | 01 | 1 | UI-21 | — | N/A | unit | `go test ./internal/app/ -run TestViewNoNewStyle -count=1` | ✅ | ⬜ pending |
| 11-02-01 | 02 | 2 | UI-20 | — | N/A | unit | `go test ./internal/app/ -run TestRegression_ClipboardAutoClearWithChrome -count=1` | ❌ W0 | ⬜ pending |
| 11-02-02 | 02 | 2 | UI-20 | — | N/A | unit | `go test ./internal/app/ -run TestRegression_RecipientFormMenuHints -count=1` | ❌ W0 | ⬜ pending |
| 11-02-03 | 02 | 2 | UI-20 | — | N/A | unit | `go test ./internal/app/ -run TestRegression_HealthOverlayOnNarrowWidth -count=1` | ❌ W0 | ⬜ pending |
| 11-02-04 | 02 | 2 | UI-20 | — | N/A | manual | Manual sweep checklist (4 Linux combos × resize/enter/exit/q/ctrl-c) | N/A | ⬜ pending |
| 11-02-05 | 02 | 2 | UI-20 | — | N/A | unit | `go test ./internal/app/ -count=1` (full pkg green after `m.quitting` wiring) | ✅ | ⬜ pending |
| 11-02-06 | 02 | 2 | UI-20 | — | N/A | grep | `grep -c "Verified Terminals" README.md` returns 1 | ❌ W0 | ⬜ pending |
| 11-02-07 | 02 | 2 | UI-20 | — | N/A | grep | `test -f .github/ISSUE_TEMPLATE/terminal-bug.yml && grep -c "terminal name" .github/ISSUE_TEMPLATE/terminal-bug.yml` returns ≥1 | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

*Wave 0 = file does not exist yet; planner's task creates it.*

---

## Wave 0 Requirements

- [ ] `internal/app/regression_test.go` — covers UI-20 (3 chrome-interaction sanity teatests per D-507)
- [ ] `README.md` "Verified Terminals" H2 section — covers SC3 community-feedback channel (D-510)
- [ ] `.github/ISSUE_TEMPLATE/terminal-bug.yml` — covers SC3 community feedback (D-510)
- [ ] `.planning/phases/11-regression-perf-gates/screenshots/{alacritty,ghostty,tmux-nested,vscode-integrated}.png` — covers SC3 manual sweep evidence (D-511)

*All other test infrastructure exists. No framework install needed. teatest is intentionally NOT used; project uses Update-loop pattern.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Chrome renders correctly on 4 Linux terminals | UI-20 SC3 | Requires real terminal emulator; CI cannot replicate xterm.js / tmux nesting / Ghostty pipeline | Per-combo (alacritty, ghostty, tmux-nested-in-alacritty, vscode-integrated): launch sops-tui, resize 80×24 ↔ 200×60, observe no flicker, alt-screen enters cleanly, alt-screen exits cleanly, `q` returns shell with cursor at expected position, ctrl-c returns shell with clipboard cleared. Capture PNG screenshot at 200×60 in stateFileList. |
| Alt-screen exit leaves no chrome residue | UI-20 SC4 | `m.quitting` flag wiring is asserted via test, but visual verification on real terminals is the SC4 evidence | During SC3 sweep above: after `q` press, observe shell prompt area for chrome residue. Per-combo note: "no residue / present (filed as #XXXX)". |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify (manual sweep is contained to one task in Plan 02 wave 2)
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s (full suite)
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
