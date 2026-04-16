---
phase: 05-power-features
plan: 01
subsystem: health-check, sops-executor, git-integration, ui-design-system
tags: [health-check, shannon-entropy, sops-rotate, recipient-management, git-staleness, styles, keybindings]
dependency_graph:
  requires: []
  provides:
    - internal/health package with pure computation functions
    - AddRecipient/RemoveRecipient in sops/executor
    - GetLastCommitTime in git/status
    - Phase 5 styles in ui/styles.go
    - Phase 5 keybindings in keys/bindings.go
  affects:
    - internal/health (new package)
    - internal/sops/executor.go
    - internal/git/status.go
    - internal/ui/styles.go
    - internal/keys/bindings.go
tech_stack:
  added: []
  patterns:
    - TDD (RED/GREEN) for health check package
    - Pure computation package with no Bubble Tea dependencies
    - Format-aware weak secret detection per D-08 (entropy + suffix + regex)
    - SHA-256 hash-keyed duplicate detection (T-05-02: never store plaintext)
    - SopsRotateTimeout=60s for recipient rotation operations (T-05-03)
    - storer.ErrStop pattern for single-commit iteration in go-git
key_files:
  created:
    - internal/health/checker.go
    - internal/health/checker_test.go
  modified:
    - internal/sops/executor.go
    - internal/sops/executor_test.go
    - internal/git/status.go
    - internal/git/status_test.go
    - internal/ui/styles.go
    - internal/keys/bindings.go
decisions:
  - IsWeakSecret format-aware check precedes entropy check for sensitive-suffix keys — valid formats (UUID, base64, hex, bcrypt) bypass entropy even when entropy is low by design (e.g. a UUID with many zeros)
  - health package duplicates regex patterns from ui/rotate.go to avoid circular import (ui imports health)
  - FindDuplicates uses 16-char truncated SHA-256 hex for ValueHash (T-05-02)
metrics:
  duration: "~20 minutes"
  completed: "2026-04-16T09:30:48Z"
  tasks: 2
  files: 8
---

# Phase 5 Plan 01: Backend Foundation Summary

**One-liner:** Health check pure functions (Shannon entropy, weak/duplicate detection per D-08), sops recipient rotation wrappers, git staleness function, and 8 Phase 5 design tokens.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 (RED) | Health check tests | d2b23dd | internal/health/checker_test.go |
| 1 (GREEN) | Health check implementation | 70344f9 | internal/health/checker.go |
| 2 | Executor, git, styles, bindings | 6aa9fc8 | executor.go, status.go, styles.go, bindings.go + tests |

## What Was Built

### internal/health/checker.go (new package)

Pure computation package with no Bubble Tea or UI dependencies:

- `ShannonEntropy(s string) float64` — frequency map over runes, stdlib math.Log2
- `IsWeakSecret(keyPath, value string) (bool, string)` — three-stage check: length < 16, format-aware suffix check (D-08), then entropy < 3.5
- `FindDuplicates(fileValues map[string]map[string]string) []Duplicate` — SHA-256 hash dedup across file maps (T-05-02: never stores plaintext)
- `hasKnownFormat(value string) bool` — regex matching for bcrypt, UUID, hex, base64
- Types: `WeakSecret`, `Duplicate`, `Location`, `StaleFile`, `HealthCheckResult`, `HealthCheckResult.IsEmpty()`

### internal/sops/executor.go additions

- `const SopsRotateTimeout = 60 * time.Second` — longer timeout for rotate operations
- `func AddRecipient(ctx, filePath, pubkey)` — calls `sops rotate -i --add-age`
- `func RemoveRecipient(ctx, filePath, pubkey)` — calls `sops rotate -i --rm-age`

### internal/git/status.go addition

- `func GetLastCommitTime(repoRoot, relPath string) (time.Time, error)` — uses `repo.Log` with `storer.ErrStop` to stop after most recent commit

### internal/ui/styles.go additions (Phase 5)

8 new style variables: `HealthWeakStyle`, `HealthDupeStyle`, `HealthStaleStyle`, `HealthSectionHeaderStyle`, `HealthSkippedStyle`, `SelectionIndicatorStyle`, `ValidationErrorStyle`, `RecipientIndexStyle`

### internal/keys/bindings.go additions (Phase 5)

5 new keybindings:
- FileListKeyMap: `ToggleSelect` (space), `BulkReKey` (K), `HealthCheck` (H)
- DetailKeyMap: `AddRecipient` (a), `RemoveRecipient` (d)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] IsWeakSecret check order for format-aware detection**
- **Found during:** Task 1 GREEN phase — `TestIsWeakSecret/valid_UUID_with__key_suffix` failed
- **Issue:** The plan's pseudocode placed entropy check before format-aware suffix check. UUID `550e8400-e29b-41d4-a716-446655440000` has low character diversity (many zeros), so it scored < 3.5 entropy and was wrongly flagged as "low entropy" before reaching the format check.
- **Fix:** Moved the sensitive-suffix format check BEFORE the entropy check. Valid formats with sensitive suffixes now bypass entropy regardless (a UUID with zeros is still a valid UUID). Only keys WITHOUT sensitive suffixes fall through to the entropy check.
- **Files modified:** internal/health/checker.go
- **Commit:** 70344f9

## TDD Gate Compliance

- RED gate: `test(05-01)` commit d2b23dd — failing tests written first
- GREEN gate: `feat(05-01)` commit 70344f9 — implementation passes all tests

## Known Stubs

None — all functions are fully implemented with no placeholder data.

## Threat Flags

None — no new network endpoints, auth paths, or file access patterns introduced beyond those in the plan's threat model.

## Self-Check: PASSED

All created/modified files verified present:

- internal/health/checker.go: FOUND
- internal/health/checker_test.go: FOUND
- internal/sops/executor.go (modified): FOUND
- internal/sops/executor_test.go (modified): FOUND
- internal/git/status.go (modified): FOUND
- internal/git/status_test.go (modified): FOUND
- internal/ui/styles.go (modified): FOUND
- internal/keys/bindings.go (modified): FOUND

All commits verified present:
- d2b23dd (RED): FOUND
- 70344f9 (GREEN): FOUND
- 6aa9fc8 (Task 2): FOUND
