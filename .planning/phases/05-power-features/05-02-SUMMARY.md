---
phase: 05-power-features
plan: "02"
subsystem: ui
tags: [overlay, health-check, recipient-management, tdd, age-validation]
dependency_graph:
  requires:
    - internal/health/checker.go (health types: HealthCheckResult, WeakSecret, Duplicate, StaleFile, Location)
  provides:
    - internal/ui/health.go (HealthModel overlay — imported by plan 03 AppModel wiring)
    - internal/ui/recipientform.go (RecipientFormModel overlay — imported by plan 03 AppModel wiring)
  affects:
    - internal/ui/styles.go (Phase 5 styles added)
    - go.mod (filippo.io/age v1.3.1 added)
tech_stack:
  added:
    - filippo.io/age v1.3.1 (age.ParseX25519Recipient for recipient key validation)
  patterns:
    - HistoryModel analog for HealthModel (loading state, SetResults, ScrollDown/Up, full-screen box)
    - SearchModel analog for RecipientFormModel (textinput wrap, Focus, Update, View)
    - DiffModel overlay pattern for RecipientFormModel (bordered box, confirmed/cancelled flags)
key_files:
  created:
    - internal/ui/health.go
    - internal/ui/health_test.go
    - internal/ui/recipientform.go
    - internal/ui/recipientform_test.go
    - internal/health/checker.go (local copy for compilation — plan 01 owns canonical version)
  modified:
    - internal/ui/styles.go (Phase 5 health/validation/selection styles added)
    - go.mod (filippo.io/age v1.3.1)
    - go.sum
decisions:
  - "Created internal/health/checker.go locally in worktree to unblock compilation during wave 1 parallel execution; plan 01 owns the canonical version"
  - "Used bubbles/v2/textinput instead of huh/v2 for RecipientFormModel — single-field input does not benefit from multi-field form handling; SearchModel serves as proven template"
  - "ValidationErrorStyle added to styles.go (not in plan 01 styles list) since it is exclusively needed by RecipientFormModel"
metrics:
  duration_minutes: 3
  completed_date: "2026-04-16"
  tasks_completed: 2
  files_created: 5
  files_modified: 3
---

# Phase 5 Plan 02: UI Components — HealthModel and RecipientFormModel Summary

**One-liner:** HealthModel full-screen overlay with grouped findings and RecipientFormModel modal with filippo.io/age bech32 validation.

## What Was Built

### Task 1: HealthModel overlay

`internal/ui/health.go` — `HealthModel` renders a full-screen bordered overlay for health check results:

- **Loading state**: Shows "Running health check..." until `SetResults()` is called
- **Empty state**: Shows "No issues found" with "All secrets passed health checks." subtitle
- **Findings state**: Three grouped sections rendered with severity-tagged prefixes:
  - `[WEAK]` (warning color) — weak secrets with file path, key path, reason
  - `[DUPE]` (error color) — duplicate values with all location pairs joined by "AND"
  - `[STALE]` (muted color) — stale files with "N days ago" text
- **Errors footer**: "N file(s) skipped -- could not decrypt" when errors present
- **Navigation footer**: "j/k scroll H or Esc close"
- **Scroll**: `ScrollDown()`/`ScrollUp()` operate on `buildContentLines()`, clamped at bounds
- **No Update() method**: scrolling driven by parent AppModel, matching HistoryModel/MetadataModel pattern

### Task 2: RecipientFormModel modal

`internal/ui/recipientform.go` — `RecipientFormModel` is a full-screen modal overlay:

- **Single textinput** with `Placeholder = "age1..."`, `CharLimit = 200` (T-05-06 DoS mitigation)
- **Lifecycle**: `Activate()` resets state and returns focus cmd; `IsActive()` tracks active/done state
- **Validation on Enter**: `age.ParseX25519Recipient()` validates bech32 + 32-byte key (T-05-05)
- **Error display**: `ValidationErrorStyle.Render(errMsg)` shown inline below input on invalid key
- **Cancel on Esc**: sets `cancelled = true`
- **Confirm on valid Enter**: sets `confirmed = true`
- **View**: title + prompt + input area + optional error + footer with [enter]/[esc] hints

### Styles added to styles.go

- `HealthWeakStyle` — warning color for [WEAK] tag
- `HealthDupeStyle` — error color for [DUPE] tag
- `HealthStaleStyle` — muted color for [STALE] tag
- `HealthOkStyle` — success color for healthy indicator
- `ValidationErrorStyle` — error color for inline validation errors
- `SelectionIndicator` — accent color for bulk re-key selection indicator

## Test Coverage

| File | Tests | Results |
|------|-------|---------|
| health_test.go | 8 subtests in TestHealthModel | All PASS |
| recipientform_test.go | 12 subtests in TestRecipientFormModel | All PASS |

Test coverage includes: loading state, empty state, each finding type ([WEAK]/[DUPE]/[STALE]), errors footer, scroll clamping, Activate lifecycle, Esc cancel, Enter validation (empty + invalid key), confirmed/cancelled state transitions, IsActive flag.

## Commits

| Task | Commit | Description |
|------|--------|-------------|
| Task 1 | fe2dd5d | feat(05-02): HealthModel overlay with health checker types |
| Task 2 | f82326e | feat(05-02): RecipientFormModel modal with age key validation |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Created health package locally for parallel wave compilation**
- **Found during:** Task 1
- **Issue:** `internal/health/checker.go` (owned by plan 01) does not exist in this worktree during wave 1 parallel execution. `health.go` imports `internal/health`, which would fail to compile.
- **Fix:** Created `internal/health/checker.go` in this worktree with the full type definitions from PATTERNS.md. Plan 01 will create the canonical version; the orchestrator merges both worktrees after the wave.
- **Files modified:** `internal/health/checker.go`
- **Commit:** fe2dd5d (included in Task 1 commit)

**2. [Rule 2 - Missing functionality] Added ValidationErrorStyle to styles.go**
- **Found during:** Task 2
- **Issue:** `recipientform.go` uses `ValidationErrorStyle` for inline error rendering (required for T-05-05 mitigation), but this style was not in plan 01's styles list.
- **Fix:** Added `ValidationErrorStyle = lipgloss.NewStyle().Foreground(ColorError)` to styles.go alongside the other Phase 5 styles.
- **Files modified:** `internal/ui/styles.go`
- **Commit:** fe2dd5d (included in Task 1 styles commit)

## Known Stubs

None — all rendering paths are fully implemented. HealthModel and RecipientFormModel have no hardcoded placeholder values. Plan 03 will wire these components into AppModel with real data.

## Threat Surface Scan

No new network endpoints, auth paths, or file access patterns introduced. All threat model items addressed:

| Threat | Mitigation | Status |
|--------|-----------|--------|
| T-05-05: Tampering via recipient key input | `age.ParseX25519Recipient()` validates before any sops call | Implemented |
| T-05-06: DoS via long input | `CharLimit = 200` on textinput | Implemented |
| T-05-07: Info disclosure via health overlay | Health overlay shows only file paths and key names, not values | Accepted by design |

## Self-Check: PASSED

Files created/modified:
- `internal/ui/health.go` — FOUND
- `internal/ui/health_test.go` — FOUND
- `internal/ui/recipientform.go` — FOUND
- `internal/ui/recipientform_test.go` — FOUND
- `internal/health/checker.go` — FOUND
- `internal/ui/styles.go` — FOUND (modified)

Commits:
- fe2dd5d — FOUND
- f82326e — FOUND
