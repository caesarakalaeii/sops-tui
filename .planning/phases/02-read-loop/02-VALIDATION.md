---
phase: 2
slug: read-loop
status: draft
nyquist_compliant: true
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
| Plan 01 Task 1: SopsDiscoverer | 02-01 | 1 | NAV-01 | T-02-01, T-02-02 | Path traversal guard, safe regex compile | unit+integration | `go test ./internal/sops/...` | No (Wave 0) | pending |
| Plan 01 Task 2: YamlParser | 02-01 | 1 | NAV-02, DEC-03 | T-02-03, T-02-04 | File size guard, type switch (no assertion) | unit | `go test ./internal/parser/... ./internal/ui/...` | No (Wave 0) | pending |
| Plan 02 Task 1: MetadataModel | 02-02 | 1 | DEC-04 | — | N/A | unit | `go test ./internal/ui/... -run TestMetadata` | No (Wave 0) | pending |
| Plan 02 Task 2: SearchModel + styles | 02-02 | 1 | NAV-04 | T-02-07 | CharLimit 100 on textinput | unit | `go test ./internal/ui/...` | No (Wave 0) | pending |
| Plan 03 Task 1: AppModel wiring | 02-03 | 2 | NAV-01, NAV-02, NAV-04, DEC-03, DEC-04 | T-02-10, T-02-11 | Async discovery, file size guard | unit+integration | `go test ./... -count=1` | Yes (partial) | pending |

*Status: pending / green / red / flaky*

---

## Wave 0 Requirements

- [ ] `internal/sops/discoverer_test.go` — stubs for NAV-01 (file discovery from .sops.yaml)
- [ ] `internal/parser/yaml_test.go` — stubs for NAV-02, DEC-03 (key extraction, value masking)
- [ ] `internal/ui/metadata_test.go` — stubs for DEC-04 (metadata overlay)
- [ ] `internal/ui/search_test.go` — stubs for NAV-04 (fuzzy search)

*If none: "Existing infrastructure covers all phase requirements."*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Fuzzy match highlighting renders correctly | NAV-04 | Visual rendering depends on terminal capabilities | Run sops-tui, press /, type partial filename, verify accent-colored matched chars |
| Status bar breadcrumb updates on navigation | NAV-06 | Visual state transitions in full TUI | Navigate file list -> detail -> back, verify breadcrumb text changes |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 5s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
