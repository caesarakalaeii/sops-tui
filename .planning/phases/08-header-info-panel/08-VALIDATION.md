---
phase: 08
slug: header-info-panel
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-28
---

# Phase 08 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Filled by gsd-planner during plan creation; finalised after plan-checker passes.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) + sebdah/goldie + charmbracelet/x/exp/teatest |
| **Config file** | none — uses go test discovery |
| **Quick run command** | `go test ./internal/ui/ ./internal/git/ -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~25 seconds (full suite incl. golden regen comparison) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./<package> -count=1 -run <task-test-name>`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd-verify-work`:** Full suite green + grep-gates pass
- **Max feedback latency:** ~25 seconds

---

## Per-Task Verification Map

> Filled by gsd-planner — one row per task, each REQ-ID covered, each Pitfall mitigation testable.

| Task ID | Plan | Wave | Requirement | Pitfall Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-------------|-----------------|-----------|-------------------|-------------|--------|
| TBD | 01 | 1 | UI-04 | P-11 | fingerprint truncated ≤10 cells | unit | `go test ./internal/ui/ -run TestRenderInfoPanel` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/ui/infopanel_test.go` — stubs for `TestRenderInfoPanel_*`, `TestMiddleTruncate_*`, `TestParseAgeKey_*`
- [ ] `internal/ui/crumbs_test.go` — stubs for `TestRenderCrumbs_*`, `TestTruncateSegmentsToWidth_*`
- [ ] `internal/git/status_test.go` — extend with `TestGetBranch_*` subtests (non-repo / branch / detached)
- [ ] `internal/app/chrome_test.go` — extend file-scope arrays for `TestChromeASCIIOnly` + `TestChromeNormalBorderOnly` to include `infopanel.go` + `crumbs.go`; allowlist `…` (U+2026)
- [ ] `internal/ui/submodel_view_no_newstyle_test.go` — extend allowlist for `infopanel.go` + `crumbs.go`

*Frameworks (`go test`, `goldie`, `teatest`) are already installed; no install step required.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Visual chip pill rendering at 80×24 / 120×40 / 200×60 | UI-07 | Subjective — pill shape + bg fill + active accent legibility | Run `./sops-tui` in 3 terminal sizes; confirm `<files> <prod.yaml>` chip pills render with active highlight; confirm narrow-tier `<41` cols still shows 1-row stub + crumb + body + status |
| Active chip bold survives 16-color downsample | Pitfall 5, Pitfall 9, UI-07 | Requires actual `TERM=xterm` terminal | `TERM=xterm ./sops-tui` — confirm active chip still distinguishable via bold weight even when bg/fg colors degrade |
| Info-panel paths are repo-relative | UI-05, Pitfall 11 | Cross-environment check | Run from a repo where `pwd` is below `$HOME`; confirm `cfg:` row shows `secrets/prod.yaml`-style relative path, never `$HOME/...` or absolute |
| Age fingerprint truncation visible | UI-05, Pitfall 11 | Visual ellipsis confirmation | Confirm `age:` row shows `age1abc…xyz`-style middle-truncated fingerprint with U+2026 ellipsis; never the full ~62-char string |

---

## Grep-Gate Coverage Map

| Gate Test | Phase 7/7.1 file scope | Phase 8 additions |
|-----------|-----------------------|-------------------|
| `TestChromeASCIIOnly` | chrome.go, logo.go, menu.go, statusbar.go | + infopanel.go + crumbs.go; allowlist `…` (U+2026) |
| `TestChromeNormalBorderOnly` | chrome.go, statusbar.go | + infopanel.go + crumbs.go |
| `TestViewNoNewStyle` / `TestSubmodelViewsNoNewStyle` | View() reachables | + infopanel.go + crumbs.go (no `lipgloss.NewStyle()` allowed at render time) |

---

## Golden File Regeneration

| Golden | Width × Height | Tier | Why It Changes |
|--------|----------------|------|---------------|
| `resize_80x24.golden` | 80×24 | mid (chrome stub + menu, no info panel) | Crumb row inflates from `""` to `<files>` chip |
| `resize_120x40.golden` | 120×40 | full (chrome + info panel + menu) | Both info-panel content AND crumb row inflate |
| `resize_200x60.golden` | 200×60 | full | Both info-panel + crumb row + new style vars |
| `resize_40x12.golden` | 40×12 | narrow (1-row stub) | Crumb row STILL inflates (D-216: independent of chrome tier); previously absent |

> Per RESEARCH.md finding 5: 40×12 golden WILL change; executor must regenerate.

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags (`go test -count=1` always; never `go test -count=0`)
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter (after planner finalises matrix)

**Approval:** pending
