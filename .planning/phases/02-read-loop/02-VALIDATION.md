---
phase: 2
slug: read-loop
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-04-14
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) + testify assertions |
| **Config file** | none — go test discovers *_test.go automatically |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test -v -count=1 ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test -v -count=1 ./...`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| TBD | TBD | TBD | NAV-01 | — | N/A | unit+integration | `go test ./internal/sops/...` | ❌ W0 | ⬜ pending |
| TBD | TBD | TBD | NAV-02 | — | N/A | unit | `go test ./internal/ui/...` | ✅ | ⬜ pending |
| TBD | TBD | TBD | NAV-04 | — | N/A | unit | `go test ./internal/ui/...` | ✅ | ⬜ pending |
| TBD | TBD | TBD | DEC-03 | — | N/A | unit | `go test ./internal/ui/...` | ✅ | ⬜ pending |
| TBD | TBD | TBD | DEC-04 | — | N/A | unit | `go test ./internal/ui/...` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/sops/discovery_test.go` — stubs for NAV-01 (file discovery from .sops.yaml)
- [ ] `internal/sops/parser_test.go` — stubs for NAV-02, DEC-03 (key extraction, value masking)
- [ ] `internal/ui/metadata_test.go` — stubs for DEC-04 (metadata overlay)
- [ ] `internal/ui/search_test.go` — stubs for NAV-04 (fuzzy search)

*If none: "Existing infrastructure covers all phase requirements."*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Fuzzy match highlighting renders correctly | NAV-04 | Visual rendering depends on terminal capabilities | Run sops-tui, press /, type partial filename, verify accent-colored matched chars |
| Status bar breadcrumb updates on navigation | NAV-06 | Visual state transitions in full TUI | Navigate file list → detail → back, verify breadcrumb text changes |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
