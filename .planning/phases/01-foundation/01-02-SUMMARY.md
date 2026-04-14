---
phase: 01-foundation
plan: 02
subsystem: validator, ui
tags: [startup-validation, error-rendering, tdd, lipgloss, go]
dependency_graph:
  requires: ["01-01"]
  provides: ["startup-validation", "error-box-renderer"]
  affects: ["cmd/sops-tui/main.go (plan 04 wires RunChecks + RenderErrorBox)"]
tech_stack:
  added: []
  patterns:
    - "Options struct with func fields for dependency injection (testability without mocking frameworks)"
    - "TDD RED/GREEN cycle: tests written first, verified failing, then implementation"
    - "FindSopsYaml exported for reuse across phases (file discovery in Phase 2)"
    - "charmbracelet/x/term GetSize for terminal width detection with fallback"
key_files:
  created:
    - internal/validator/startup.go
    - internal/validator/startup_test.go
    - internal/ui/errorbox.go
    - internal/ui/errorbox_test.go
decisions:
  - "Options struct with SopsLookPath func field chosen over exec.LookPath wrapper to enable pure-Go test doubles without temp file setup"
  - "FindSopsYaml exported (not findSopsYaml) to enable Phase 2 file-discovery reuse without API change"
  - "charmbracelet/x/term used for terminal width (already in go.mod as indirect dep) instead of adding golang.org/x/term as direct dep"
metrics:
  duration: "~8 minutes"
  completed: "2026-04-14"
  tasks_completed: 2
  tasks_total: 2
  files_created: 4
  files_modified: 0
---

# Phase 01 Plan 02: Startup Validation and Error Box Summary

**One-liner:** TDD-implemented startup validator (sops/age-key/.sops.yaml checks) and lipgloss-styled stderr error box renderer with dependency-injection testability.

## What Was Built

### Task 1: Startup Validation (`internal/validator/startup.go`)

`RunChecks(opts Options) ([]ValidationResult, bool)` performs all three environment checks in a single pass (D-02):

1. **sops binary** — `exec.LookPath("sops")` via injected `SopsLookPath` func. Missing = `SeverityError`, `hasHardError=true` (D-03 hard error, HLT-01).
2. **age key file** — `os.Stat` on configurable `AgeKeyPath`. Missing = `SeverityWarn`, `hasHardError` unchanged (D-03 soft warning, HLT-02).
3. **.sops.yaml discovery** — `FindSopsYaml(startDir)` walks up via `filepath.Dir` loop. Missing = `SeverityWarn` (D-04 soft warning).

`FindSopsYaml` is exported for Phase 2 file-discovery reuse. Root termination guard (`parent == dir`) prevents infinite loop at filesystem root (T-01-04 mitigated).

### Task 2: Error Box Renderer (`internal/ui/errorbox.go`)

`RenderErrorBox(results []validator.ValidationResult, hasHardError bool, w io.Writer)` writes a lipgloss `RoundedBorder()` box to any `io.Writer`:

- Header: `"sops-tui: startup failed"` (hard error) or `"sops-tui: warnings"` (soft only)
- Border color: `ColorError` (#f38ba8) for hard errors, `ColorWarning` (#f9e2af) for warnings-only
- Labels: `ErrorLabel.Render("[ERROR]")` / `WarnLabel.Render("[WARN] ")` (styles from `styles.go`)
- Padding: sm (1 row vertical, 2 cells horizontal) per 01-UI-SPEC.md
- Width: `min(termWidth-4, 72)` with 80-col fallback when stderr is not a TTY

## Test Coverage

| File | Tests | Scenarios |
|------|-------|-----------|
| `startup_test.go` | 8 functions (11 sub-tests) | sops missing/found, age key missing, .sops.yaml missing/in-parent, hard/soft distinction, single-pass D-02, walk-up termination |
| `errorbox_test.go` | 7 functions | [ERROR] label, [WARN] label, mixed results, warnings-only header, writer output, fix text, hard-error header |

All 18 tests pass. Full suite (`go test ./...`) green.

## Deviations from Plan

### Auto-fixed Issues

None.

### Design Decisions Made During Implementation

**1. `Options` struct injected as parameter (not global var)**
- **Found during:** Task 1 design
- **Issue:** Plan suggested either `Options struct` or "functional options" — chose the simpler concrete struct since all three fields (SopsLookPath, AgeKeyPath, StartDir) are always set together. This is cleaner than variadic functional options for 3 fields.
- **Impact:** Zero — API is `RunChecks(opts Options)`, same as plan spec.

**2. `charmbracelet/x/term` used instead of `golang.org/x/term`**
- **Found during:** Task 2 terminal width detection
- **Reason:** `charmbracelet/x/term` was already in `go.mod` as an indirect dependency (from lipgloss). Adding `golang.org/x/term` as a direct dep would be redundant. Both provide identical `GetSize(fd)` semantics.
- **Files modified:** None (no new dependencies added to go.mod).

## Known Stubs

None. Both functions are fully implemented and wired to real system calls in production paths.

## Threat Surface Scan

No new network endpoints, auth paths, file writes, or schema changes introduced. All surface additions are read-only system calls (`exec.LookPath`, `os.Stat`) consistent with the threat model:

| Threat | Status |
|--------|--------|
| T-01-03: exec.LookPath only checks existence, never executes | Confirmed — Phase 1 only calls LookPath |
| T-01-04: filepath.Dir termination at root | Mitigated — `parent == dir` guard in FindSopsYaml |
| T-01-05: Error messages expose no secrets | Confirmed — only binary names, standard paths, public URLs |
| T-01-06: Age key path existence check only | Confirmed — os.Stat never reads content |

## Self-Check: PASSED

| Check | Result |
|-------|--------|
| internal/validator/startup.go | FOUND |
| internal/validator/startup_test.go | FOUND |
| internal/ui/errorbox.go | FOUND |
| internal/ui/errorbox_test.go | FOUND |
| Commit d4f6dbd (Task 1) | FOUND |
| Commit d771a66 (Task 2) | FOUND |
