---
phase: 7
slug: chrome-skeleton
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-24
---

# Phase 7 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution. Mirrors RESEARCH.md §Validation Architecture.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing` + `stretchr/testify v1.11.1` (already direct dep) |
| **Config file** | none — pure `go test ./...` |
| **Quick run command** | `go test ./internal/ui ./internal/keys -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~12 seconds full suite, ~2 seconds quick |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/ui ./internal/keys -count=1` (fast — no goldens, no bench)
- **After every plan wave:** Run `go test ./... -count=1` (full suite including goldens + bench-budget)
- **Before `/gsd-verify-work`:** Full suite must be green; manual smoke at 40×12, 80×24, 120×40, 200×60 per Phase 6 D-15 protocol
- **Max feedback latency:** 12 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 7-01-01 | 01 | 1 | UI-01 | — | N/A | unit | `go test ./internal/keys -run TestHintsFromBindings -count=1` | ❌ W0 | ⬜ pending |
| 7-01-02 | 01 | 1 | UI-02 | — | N/A | unit | `go test ./internal/ui -run TestRenderLogo -count=1` | ❌ W0 | ⬜ pending |
| 7-01-03 | 01 | 1 | UI-01 | — | N/A | unit | `go test ./internal/ui -run TestRenderMenu -count=1` | ❌ W0 | ⬜ pending |
| 7-01-04 | 01 | 1 | UI-01/UI-02 | — | N/A | unit | `go test ./internal/ui -run "TestMenuKeyStyle|TestLogoStyle" -count=1` | ❌ W0 | ⬜ pending |
| 7-02-01 | 02 | 1 | UI-06 | — | N/A | unit | `go test ./internal/ui -run TestOverlayTitle_PreservesCornersAndWidth -count=1` | ❌ W0 | ⬜ pending |
| 7-02-02 | 02 | 1 | UI-06 | — | N/A | unit | `go test ./internal/ui -run TestWrapTitled -count=1` | ❌ W0 | ⬜ pending |
| 7-02-03 | 02 | 1 | UI-15 | — | N/A | unit | `go test ./internal/ui -run TestRenderChrome -count=1` | ❌ W0 | ⬜ pending |
| 7-03-01 | 03 | 2 | UI-01 | — | N/A | unit | `go test ./internal/ui -run TestHints -count=1` | ❌ W0 | ⬜ pending |
| 7-03-02 | 03 | 2 | UI-01 | — | N/A | unit | `go test ./internal/app -run TestMenuHints -count=1` | ❌ W0 | ⬜ pending |
| 7-03-03 | 03 | 2 | UI-06 | T-7-03 (PII) | filename in title is repo-relative `m.currentFile.Name`, never absolute path | integration (golden) | `go test ./internal/app -run TestResize -count=1` | ✅ refresh | ⬜ pending |
| 7-03-04 | 03 | 2 | UI-15 | — | N/A | grep-gate | `go test ./internal/app -run TestChromeASCIIOnly -count=1` | ❌ W0 | ⬜ pending |
| 7-03-05 | 03 | 2 | UI-15 | — | N/A | grep-gate | `go test ./internal/app -run TestChromeNormalBorderOnly -count=1` | ❌ W0 | ⬜ pending |
| 7-03-06 | 03 | 2 | UI-15 | — | N/A | grep-gate (AST) | `go test ./internal/app -run TestViewNoNewStyle -count=1` | ❌ W0 | ⬜ pending |
| 7-03-07 | 03 | 2 | UI-15 (perf) | — | N/A | bench-budget | `go test ./internal/app -run TestBenchmarkAppView_UnderBudget -count=1` | ❌ W0 | ⬜ pending |
| 7-03-08 | 03 | 2 | UI-06 | — | N/A | regression | `go test ./internal/app -run TestRender -count=1` | ✅ existing | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Plan 1 introduces (test files):
- [ ] `internal/keys/hints_test.go` — covers `MenuHint` shape + `HintsFromBindings`
- [ ] `internal/ui/logo_test.go` — covers `RenderLogo` (6 rows, ~26-col width, ASCII-only)
- [ ] `internal/ui/menu_test.go` — covers `RenderMenu` (column-major fill, `Visible=false` skip, empty-hint fallback, narrow-terminal safety)

Plan 2 introduces:
- [ ] `internal/ui/chrome_test.go` — covers `WrapTitled` + `TestOverlayTitle_PreservesCornersAndWidth` (corners, width, truncation, empty-title)

Plan 3 introduces:
- [ ] `internal/app/chrome_test.go` — `TestChromeASCIIOnly` + `TestChromeNormalBorderOnly` + `TestViewNoNewStyle` (AST walker via `go/ast`, `go/parser`) + `TestBenchmarkAppView_UnderBudget`
- [ ] `internal/app/hints_test.go` — AppModel `menuHints()` dispatcher tests across (state, recipientAction, IsSearchActive)
- [ ] `internal/ui/{filelist,detail,help,diff,metadata,health,history,recipientform}_test.go` — extended each with one `TestHints` per sub-model
- [ ] `internal/app/testdata/resize_{40x12,80x24,120x40,200x60}.golden` — REFRESHED via `GOLDEN_UPDATE=1 go test ./internal/app -run TestResize`

Framework install: none — `go test` stdlib + existing testify already present.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| k9s-visual-parity smoke at 40×12, 80×24, 120×40, 200×60 | UI-01, UI-02, UI-06 | Subjective "looks like k9s" judgment per CLAUDE.md user feedback | Resize terminal to each size; visually confirm chrome renders without overflow, logo top-right, menu 2×6 visible, titled border on body |
| Logo byte-art aesthetic acceptance | UI-02 | Claude picks figlet variant; user reviews during `/gsd-verify-work` | Run TUI; confirm "SOPS" block + "tui" subscript reads cleanly within ~26 cols |
| Menu hint relevance per state | UI-01 | "Are these the right 12 keys?" is judgment, not test | Cycle through every state (FileList, Detail, Help, Diff, Metadata, Health, History, RecipientList, RecipientForm, FormatMenu, search-active FileList); confirm menu reflects keys that actually work |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter (after planner authors plans)

**Approval:** pending
