---
phase: 3
slug: write-loop
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-14
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) + testify assertions |
| **Config file** | none — existing test infrastructure |
| **Quick run command** | `go test ./internal/sops/... ./internal/ui/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/sops/... ./internal/ui/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 03-01-01 | 01 | 1 | DEC-01 | T-03-01 / — | Single-key decrypt via sops subprocess | unit | `go test ./internal/sops/...` | ❌ W0 | ⬜ pending |
| 03-01-02 | 01 | 1 | DEC-02 | T-03-02 / — | Full-file decrypt via sops subprocess | unit | `go test ./internal/sops/...` | ❌ W0 | ⬜ pending |
| 03-02-01 | 02 | 2 | EDT-01 | — | Inline edit text input activation | unit | `go test ./internal/ui/...` | ❌ W0 | ⬜ pending |
| 03-02-02 | 02 | 2 | EDT-02 | — | $EDITOR subprocess suspension | integration | `go test ./internal/sops/...` | ❌ W0 | ⬜ pending |
| 03-02-03 | 02 | 2 | EDT-03 | — | Diff overlay before re-encryption | unit | `go test ./internal/ui/...` | ❌ W0 | ⬜ pending |
| 03-03-01 | 03 | 2 | EDT-04 | T-03-03 / — | Format-aware rotation generation | unit | `go test ./internal/sops/...` | ❌ W0 | ⬜ pending |
| 03-03-02 | 03 | 2 | EDT-04 | — | Confirmation gate blocks without y | unit | `go test ./internal/ui/...` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/sops/executor_test.go` — stubs for DEC-01, DEC-02, EDT-02 subprocess tests
- [ ] `internal/ui/diff_test.go` — stubs for diff overlay rendering and confirmation
- [ ] `internal/sops/rotation_test.go` — stubs for format detection and random generation
- [ ] `golang.org/x/crypto` dependency — required for bcrypt rotation

*Existing test infrastructure (go test, testify) covers framework needs.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| $EDITOR opens correctly | EDT-02 | Requires interactive terminal + real editor binary | Run `sops-tui`, navigate to file, press `R` then `E`, verify editor opens with decrypted content |
| Visual diff coloring | EDT-03 | Requires visual inspection of lipgloss rendered output | Edit a value, verify red/green diff coloring in overlay |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
