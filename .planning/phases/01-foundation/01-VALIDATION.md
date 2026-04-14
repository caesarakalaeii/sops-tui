---
phase: 1
slug: foundation
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-14
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
| 1-01-01 | 01 | 0 | — | — | N/A | setup | `go build ./...` | ❌ W0 | ⬜ pending |
| 1-02-01 | 02 | 1 | HLT-01 | — | sops missing → styled error + exit 1 | integration | `go test ./internal/preflight/...` | ❌ W0 | ⬜ pending |
| 1-02-02 | 02 | 1 | HLT-02 | — | age key missing → warning, TUI still launches | integration | `go test ./internal/preflight/...` | ❌ W0 | ⬜ pending |
| 1-03-01 | 03 | 2 | NAV-03 | — | hjkl/g/G/ctrl-d/u navigation works | unit | `go test ./internal/tui/...` | ❌ W0 | ⬜ pending |
| 1-04-01 | 04 | 2 | NAV-05 | — | ? toggles help overlay with contextual keys | unit | `go test ./internal/tui/...` | ❌ W0 | ⬜ pending |
| 1-05-01 | 05 | 2 | NAV-06 | — | status bar visible with location/counts/env | unit | `go test ./internal/tui/...` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `go.mod` — add bubbletea v2, lipgloss v2, bubbles v2, testify dependencies
- [ ] `internal/preflight/preflight_test.go` — stubs for HLT-01, HLT-02
- [ ] `internal/tui/model_test.go` — stubs for NAV-03, NAV-05, NAV-06

*Wave 0 establishes the test infrastructure that all subsequent waves depend on.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Lipgloss-styled stderr box renders correctly | HLT-01, HLT-02 | Visual rendering depends on terminal capabilities | Run `sops-tui` without sops installed, visually confirm styled box |
| Status bar flash message timing | NAV-06 | 2-3s timer requires real-time observation | Trigger a flash, confirm it disappears after ~2s |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
