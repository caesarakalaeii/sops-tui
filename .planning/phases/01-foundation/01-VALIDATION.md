---
phase: 1
slug: foundation
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-04-14
updated: 2026-04-14
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) + stretchr/testify |
| **Config file** | none — Wave 0 installs |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test -v -count=1 ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test -v -count=1 ./...`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 1-01-01 | 01 | 1 | — | — | N/A | setup | `go build ./...` | N/A | pending |
| 1-01-02 | 01 | 1 | NAV-03 | — | Color constants match spec | unit | `go test ./internal/ui/... -run TestStyle -v` | W0 | pending |
| 1-01-03 | 01 | 1 | NAV-03 | — | Keybindings match spec | unit | `go test ./internal/keys/... -v` | W0 | pending |
| 1-02-01 | 02 | 2 | HLT-01 | T-01-03 | sops missing -> styled error + exit 1 | integration | `go test ./internal/validator/... -v` | W0 | pending |
| 1-02-02 | 02 | 2 | HLT-02 | T-01-06 | age key missing -> warning, TUI still launches | integration | `go test ./internal/validator/... -v` | W0 | pending |
| 1-03-01 | 03 | 2 | NAV-03 | — | FileListModel wraps list with explicit g/G/ctrl-d/u | unit | `go test ./internal/ui/... -run TestFileList -v` | W0 | pending |
| 1-03-02 | 03 | 2 | NAV-03 | T-01-07 | DetailModel collapsible YAML tree | unit | `go test ./internal/ui/... -run TestDetail -v` | W0 | pending |
| 1-04-01 | 04 | 3 | NAV-05, NAV-06 | — | Status bar + help overlay | unit | `go test ./internal/ui/... -run "TestStatusBar\|TestHelp" -v` | W0 | pending |
| 1-04-02 | 04 | 3 | NAV-03, NAV-05, NAV-06, HLT-01, HLT-02 | T-01-09 | Root model wiring + main.go | integration | `go test ./internal/app/... -v && go build -o /dev/null ./cmd/sops-tui/` | W0 | pending |
| 1-04-03 | 04 | 3 | — | — | Human visual verification | manual | N/A | N/A | pending |

*Status: pending / green / red / flaky*

---

## Wave 0 Requirements

- [ ] `go.mod` — add bubbletea v2, lipgloss v2, bubbles v2, testify, golang.org/x/term dependencies
- [ ] `internal/ui/styles_test.go` — stubs for color/style assertions
- [ ] `internal/keys/bindings_test.go` — stubs for keybinding match assertions
- [ ] `internal/validator/startup_test.go` — stubs for HLT-01, HLT-02
- [ ] `internal/ui/filelist_test.go` — stubs for FileListModel navigation
- [ ] `internal/ui/detail_test.go` — stubs for DetailModel tree rendering
- [ ] `internal/ui/statusbar_test.go` — stubs for NAV-06
- [ ] `internal/ui/help_test.go` — stubs for NAV-05
- [ ] `internal/app/model_test.go` — stubs for root model wiring

*Wave 0 establishes the test infrastructure that all subsequent waves depend on.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Lipgloss-styled stderr box renders correctly | HLT-01, HLT-02 | Visual rendering depends on terminal capabilities | Run `sops-tui` without sops installed, visually confirm styled box |
| Status bar flash message timing | NAV-06 | 2-3s timer requires real-time observation | Trigger a flash, confirm it disappears after ~2s |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 10s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
