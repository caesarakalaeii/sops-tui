---
phase: 08
slug: header-info-panel
status: ready
nyquist_compliant: true
wave_0_complete: false
created: 2026-04-28
updated_by_planner: 2026-04-28
---

# Phase 08 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Filled by gsd-planner during plan creation; finalised after plan-checker passes.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) + sebdah/goldie + charmbracelet/x/exp/teatest + stretchr/testify |
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

> One row per task; every REQ-ID is covered; every Pitfall mitigation is testable; every task has an `<automated>` verify command (Nyquist Dimension 8 compliance).

| Task ID | Plan | Wave | Requirement | Pitfall Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-------------|-----------------|-----------|-------------------|-------------|--------|
| 08-01-T1 | 01 | 1 | UI-04, UI-07 | — | 8 new package-var styles (no AdaptiveColor); CrumbChipActiveStyle uses Bold(true) per D-206 (Pitfall 9 redundancy) | unit (compile + style assertion) | `go build ./... && go test ./internal/ui/ -count=1 -run 'TestStyles\|TestSubmodel\|TestNoAdaptive'` | ✅ Plan 1 owns | ⬜ pending |
| 08-01-T2 | 01 | 1 | UI-04, UI-05 | Pitfall A, Pitfall 11 | Type-assert to `*age.X25519Identity` before `.Recipient().String()`; never `Identity.String()` (D-220 question 1 leak prevention) | unit (file-I/O + library) | `go test ./internal/ui/ -count=1 -run 'TestParseAgeKeyFingerprint\|TestAgeKeyFilePath'` | ✅ Plan 1 owns | ⬜ pending |
| 08-01-T3 | 01 | 1 | UI-04, UI-05, UI-07 | Pitfall B, Pitfall G, Pitfall 14 | middleTruncate uses ansi.TruncateLeft for right half (Pitfall B); active chip has bold (Pitfall G); overflow row collapses to ellipsis chip (Pitfall 14) | unit (renderer + RGB triplet) | `go test ./internal/ui/ -count=1 -run 'TestRenderInfoPanel\|TestRenderCrumbs\|TestMiddleTruncate'` | ✅ Plan 1 owns | ⬜ pending |
| 08-02-T1 | 02 | 1 | UI-04 | — | git.GetBranch returns ErrRepositoryNotExists for non-git (D-12 contract); 7-char short hash for detached HEAD | unit (file-I/O + go-git) | `go test ./internal/git/ -count=1 -run TestGetBranch -v` | ✅ Plan 2 owns | ⬜ pending |
| 08-02-T2 | 02 | 1 | UI-08 | Pitfall C (deferred) | View() normal path right-aligned env+clipboard only; no breadcrumb LEFT, no item-count CENTER, no `" | "` pipes; flash path unchanged (D-212) | unit (renderer + Segments accessor) | `go test ./internal/ui/ -count=1 -run TestStatusBar -v` | ✅ Plan 2 owns | ⬜ pending |
| 08-03-T1 | 03 | 2 | UI-04, UI-05, UI-07, UI-08 | Pitfall D, Pitfall E, Pitfall 15 | Both RenderChrome call-sites updated atomically (Pitfall E); 4 refresh seams (D-213); crumbsHeight flip (D-216); zero I/O in View() (Pitfall 15) | integration (compile + cache wiring) | `go build ./... && go vet ./internal/... && grep -c 'RenderChrome(' internal/app/model.go && grep -c 'm.infoPanel' internal/app/model.go` | ✅ Plan 3 owns | ⬜ pending |
| 08-03-T2 | 03 | 2 | UI-04, UI-05, UI-07, UI-15 | — | Grep-gate file scope includes infopanel.go + crumbs.go; 3 new integration tests pass (TestRenderChrome_FullTierWithInfoPanel + TestCrumbsHeight_NonZero + TestInfoPanelCacheRefresh_OnFilesDiscovered) | integration (grep-gate + AST walker) | `go test ./internal/app/ -count=1 -run 'TestChromeASCIIOnly\|TestChromeNormalBorderOnly\|TestViewNoNewStyle\|TestRenderChrome_FullTierWithInfoPanel\|TestCrumbsHeight_NonZero\|TestInfoPanelCacheRefresh_OnFilesDiscovered'` | ✅ Plan 3 owns | ⬜ pending |
| 08-03-T3 | 03 | 2 | UI-04, UI-07, UI-08 | Pitfall F | All 4 resize goldens regenerated (40×12 WILL change per Pitfall F); deterministic via t.Setenv("SOPS_AGE_KEY_FILE", ...); review-safe (no host paths) | golden file (TUI snapshot) | `go test ./internal/app/ -count=1 -run TestResize -v && grep -q "cfg:" internal/app/testdata/resize_200x60.golden && ! grep -q "/home/" internal/app/testdata/resize_*.golden` | ✅ Plan 3 owns | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

> All test files are co-located with production files in this phase. Plan 1 + Plan 2 + Plan 3 each create their own test files alongside production code (TDD-friendly red-then-green within each task). No separate Wave 0 task is needed because each task's `<verify>` block runs its own test scaffolding.

- [x] `internal/ui/infopanel_test.go` — Plan 1 Task 3 creates with `TestRenderInfoPanel_*`, `TestMiddleTruncate_*`
- [x] `internal/ui/crumbs_test.go` — Plan 1 Task 3 creates with `TestRenderCrumbs_*`
- [x] `internal/ui/agekey_test.go` — Plan 1 Task 2 creates with `TestParseAgeKeyFingerprint_*`, `TestAgeKeyFilePath_*`
- [x] `internal/git/status_test.go` — Plan 2 Task 1 extends with `TestGetBranch` (3 subtests)
- [x] `internal/ui/statusbar_test.go` — Plan 2 Task 2 modifies (deletes `TestStatusBarSetItemCount`; adds `TestStatusBar_RightAlignOnly`, `TestStatusBar_SegmentsAccessor`, `TestStatusBar_FlashUnchanged`)
- [x] `internal/app/chrome_test.go` — Plan 3 Task 2 extends file-scope arrays + adds 3 integration tests
- [x] `internal/ui/submodel_view_no_newstyle_test.go` — Plan 3 Task 2 extends `submodelFiles` allowlist with `infopanel.go` + `crumbs.go`

*Frameworks (`go test`, `goldie`, `teatest`, `testify`) are already installed; no install step required. `filippo.io/age`, `go-git/v5`, `lipgloss/v2`, `ansi` are all already direct deps in go.mod (verified 2026-04-28).*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Visual chip pill rendering at 80×24 / 120×40 / 200×60 | UI-07 | Subjective — pill shape + bg fill + active accent legibility | Run `./sops-tui` in 3 terminal sizes; confirm `<files> <prod.yaml>` chip pills render with active highlight; confirm narrow-tier `<41` cols still shows 1-row stub + crumb + body + status |
| Active chip bold survives 16-color downsample | Pitfall 5, Pitfall 9, UI-07 | Requires actual `TERM=xterm` terminal | `TERM=xterm ./sops-tui` — confirm active chip still distinguishable via bold weight even when bg/fg colors degrade |
| Info-panel paths are repo-relative | UI-05, Pitfall 11, D-220 Q2 | Cross-environment check | Run from a repo where `pwd` is below `$HOME`; confirm `cfg:` row shows `secrets/prod.yaml`-style relative path, never `$HOME/...` or absolute |
| Age fingerprint truncation visible | UI-05, Pitfall 11, D-220 Q5 | Visual ellipsis confirmation | Confirm `age:` row shows `age1abc…xyz`-style middle-truncated fingerprint with U+2026 ellipsis; never the full ~62-char string |
| Status bar shows env+clipboard only (no breadcrumb LEFT, no item-count) | UI-08 | Visual confirmation of D-211 deletion | Confirm bottom status bar contains ONLY env indicators (`sops:✓ age:✓ .sops.yaml:✓`) optionally prefixed with `[clip]`, right-aligned; no `sops-tui > files` breadcrumb on the left, no `12 items` count in the middle, no ` | ` pipe separators |

---

## Grep-Gate Coverage Map

| Gate Test | Phase 7/7.1 file scope | Phase 8 additions (Plan 3 Task 2) |
|-----------|-----------------------|-----------------------------------|
| `TestChromeASCIIOnly` | chrome.go, logo.go, menu.go, crumbs.go (skip-if-missing pre-Phase-8) | + infopanel.go (full add); crumbs.go skip-if-missing comment can be dropped (file now exists). Allowlist: `…` (U+2026) ALREADY PRESENT at chrome_test.go:53 — no allowlist change needed for Phase 8 |
| `TestChromeNormalBorderOnly` | chrome.go, logo.go, menu.go | + infopanel.go + crumbs.go |
| `TestViewNoNewStyle` (BFS in internal/app/) | View() reachables via BFS over call edges | UNCHANGED — already covers all helpers reachable from View(); Plan 1 + Plan 3 use only package-var styles in any helper that View() reaches |
| `TestSubmodelViewsNoNewStyle` (file-scope walker in internal/ui/) | filelist.go, detail.go, help.go, diff.go, metadata.go, health.go, history.go, recipientform.go (8 files) | + infopanel.go + crumbs.go (Plan 3 Task 2 extends `submodelFiles` allowlist) |

---

## Golden File Regeneration

| Golden | Width × Height | Tier | Why It Changes |
|--------|----------------|------|----------------|
| `resize_80x24.golden` | 80×24 | mid (chrome 6 rows + menu, no info panel) | Crumb row inflates from `""` to chip pill row; status bar shrinks (loses breadcrumb LEFT, item-count CENTER, pipe separators) |
| `resize_120x40.golden` | 120×40 | full (chrome 6 rows + info panel + menu + logo) | Both info-panel content (5 "-" markers in test fixture) AND crumb row inflate; status bar shrinks |
| `resize_200x60.golden` | 200×60 | full | Same pattern as 120×40 |
| `resize_40x12.golden` | 40×12 | narrow (1-row stub) | Crumb row STILL inflates (D-216: independent of chrome tier); previously absent. Status bar also shrinks. Per RESEARCH.md finding 5 + Pitfall F: this golden WILL change despite the chrome remaining a 1-row stub. |

**Determinism gate:** Plan 3 Task 3 adds `t.Setenv("SOPS_AGE_KEY_FILE", filepath.Join(t.TempDir(), "nonexistent-keys.txt"))` to the resize test setup so `loadAgeFingerprint()` returns "" deterministically (no host fingerprint leakage). All 4 goldens render `cfg: -`, `age: -`, `rcp: -`, `git: -`, `fil: -` (where applicable per tier visibility) — review-safe.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies (every Plan 1/2/3 task has an `<automated>` block per Nyquist Dimension 8)
- [x] Sampling continuity: no 3 consecutive tasks without automated verify (every task verifies via go test or grep-based file assertion)
- [x] Wave 0 covers all MISSING references (test files are created in-task per the TDD discipline — no separate Wave 0 setup required)
- [x] No watch-mode flags (`go test -count=1` always; never `go test -count=0`)
- [x] Feedback latency < 30s (estimated ~25s for full suite per Test Infrastructure)
- [x] `nyquist_compliant: true` set in frontmatter (per-task table is complete; every REQ-ID + Pitfall has an automated verification command)

**Approval:** ready for execution
