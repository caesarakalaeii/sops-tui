---
phase: 01-foundation
plan: "04"
subsystem: app-integration
tags: [tui, statusbar, help-overlay, root-model, main-entrypoint, startup-validation]
dependency_graph:
  requires: ["01-01", "01-02", "01-03"]
  provides: [runnable-binary, app-model, statusbar-model, help-model, main-entrypoint]
  affects: [all-future-phases]
tech_stack:
  added:
    - "internal/app package — root AppModel with sessionState routing"
    - "StatusBarModel with flash timer and generation counter"
    - "HelpModel wrapping bubbles/v2/help with contextual keymaps"
  patterns:
    - "sessionState enum for full-screen view routing (Pattern 1)"
    - "Generation counter for flash timer leak prevention (Pitfall 6)"
    - "tea.Tick for 2-second flash message timeout"
    - "tea.NewView + v.AltScreen=true per bubbletea v2 API"
    - "lipgloss.Height() for dynamic status bar height calculation (Pattern 5)"
key_files:
  created:
    - internal/ui/statusbar.go
    - internal/ui/statusbar_test.go
    - internal/ui/help.go
    - internal/ui/help_test.go
    - internal/app/model.go
    - internal/app/model_test.go
  modified:
    - cmd/sops-tui/main.go
decisions:
  - "StatusBarModel.Update returns (StatusBarModel, tea.Cmd) not (tea.Model, tea.Cmd) — it is a child component, not a root model; callers forward the cmd"
  - "tea.RequestWindowSize() and tea.Quit() both return Msg not Cmd — wrapped as func() tea.Msg{} in Init/Update"
  - "Help ViewState mirrors sessionState 1:1 (ViewFileList=0, ViewDetail=1) — casting between them is safe as long as stateHelp is never cast"
metrics:
  duration: "~20 minutes"
  completed: "2026-04-14"
  tasks_completed: 3
  files_changed: 7
requirements_satisfied: [NAV-05, NAV-06, HLT-01, HLT-02]
---

# Phase 01 Plan 04: TUI Integration and Entry Point — Summary

**One-liner:** Complete sops-tui binary with sessionState-routed AppModel, flash-timer StatusBarModel, contextual HelpModel, and validator-gated main.go entry point.

## What Was Built

### Task 1: StatusBarModel and HelpModel (TDD)

**StatusBarModel** (`internal/ui/statusbar.go`):
- Three-section layout: left=breadcrumb, center=item count, right=env indicators
- Flash message system: `Flash(msg) (StatusBarModel, tea.Cmd)` increments a generation counter and schedules a 2-second `tea.Tick`
- `FlashClearMsg{Gen int}` — only clears flash when `Gen == flashGen` (prevents stale timer clearing a newer flash, per RESEARCH Pitfall 6)
- Unicode env indicators: checkmark `\u2713` (success), cross `\u2717` (error), warning `\u26A0` (warning)
- `SetBreadcrumb(segments ...string)` — joins with " > ", always prefixes "sops-tui", last segment in accent color
- 11 tests, all passing

**HelpModel** (`internal/ui/help.go`):
- Wraps `charm.land/bubbles/v2/help.Model` with `ShowAll = true` (full-screen per D-08)
- `ViewState` enum (`ViewFileList`, `ViewDetail`) selects contextual keybindings
- Full-screen `RoundedBorder` overlay with surface background and muted border
- Footer: `"Press ? or Esc to close"` per UI-SPEC copywriting contract
- 5 tests, all passing

### Task 2: Root AppModel and main.go

**AppModel** (`internal/app/model.go`):
- `sessionState` enum: `stateFileList`, `stateDetail`, `stateHelp`
- `prevState` field: used by help overlay to return to the correct previous view
- `Init()` returns `func() tea.Msg { return tea.RequestWindowSize() }` — `RequestWindowSize` returns `Msg` not `Cmd` in v2, so it must be wrapped
- Global key dispatch in `Update`: `?` toggles help, `q`/`ctrl+c` quits (via wrapped `tea.Quit()`), `esc` closes help or returns from detail
- `WindowSizeMsg` propagates to all children with `lipgloss.Height()` for dynamic status bar height
- Status bar `Update` always runs (flash timer must fire regardless of active state)
- `View()` returns `tea.View` with `v.AltScreen = true`
- 8 tests, all passing

**main.go** (`cmd/sops-tui/main.go`):
- `validator.DefaultOptions()` → `validator.RunChecks(opts)` — accepts Options for testability
- `ui.RenderErrorBox(results, hasHardError, os.Stderr)` — styled stderr output before TUI
- `os.Exit(1)` on hard error (T-01-09 mitigation: no path to TUI with missing sops)
- `ui.EnvStatus` built from validation result messages
- `app.NewAppModel(env)` → `tea.NewProgram(model)` → `p.Run()`

### Task 3: Verification (Auto-approved)

Build and test verification passed:
- `go build -o /dev/null ./cmd/sops-tui/` — exits 0
- `go test ./... -count=1` — all packages pass (app, keys, ui, validator)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] StatusBarModel.Update signature changed from tea.Model to concrete type**
- **Found during:** Task 1 TDD GREEN phase
- **Issue:** Plan specified `Update(tea.Msg) (tea.Model, tea.Cmd)` but `StatusBarModel` does not implement `tea.Model` (missing `Init()`) — it is a child component, not a root model
- **Fix:** Changed to `Update(msg tea.Msg) (StatusBarModel, tea.Cmd)` — correct for child components
- **Files modified:** `internal/ui/statusbar.go`, `internal/ui/statusbar_test.go`
- **Commit:** e0a5ad0

**2. [Rule 1 - Bug] tea.RequestWindowSize and tea.Quit are Msg not Cmd in bubbletea v2**
- **Found during:** Task 2 implementation
- **Issue:** Plan's RESEARCH example shows `return tea.RequestWindowSize` as a Cmd, but in v2 both `RequestWindowSize()` and `Quit()` return `tea.Msg`, not `tea.Cmd`
- **Fix:** Wrapped as `func() tea.Msg { return tea.RequestWindowSize() }` and `func() tea.Msg { return tea.Quit() }` in `Init()` and `Update()` respectively
- **Files modified:** `internal/app/model.go`
- **Commit:** 168dc3e

**3. [Rule 1 - Bug] validator.RunChecks takes Options parameter**
- **Found during:** Task 2 main.go wiring
- **Issue:** Plan's interface stub shows `RunChecks()` with no arguments, but the actual implementation from Plan 02 requires `RunChecks(opts Options)` for testability
- **Fix:** `main.go` calls `validator.DefaultOptions()` first, then passes opts to `RunChecks(opts)`
- **Files modified:** `cmd/sops-tui/main.go`
- **Commit:** 168dc3e

## Known Stubs

- **File list is empty** (`[]ui.FileItem{}`): `NewAppModel` initializes with no files. Phase 2 wires SOPS file discovery. The empty state renders "No SOPS files found" per UI-SPEC copywriting contract — this is intentional for Phase 1.
- **Detail view nodes empty** (`[]ui.TreeNode{}`): When drilling into a file, `NewDetailModel` receives empty nodes. Phase 2/3 will populate with parsed YAML content. Renders "No keys found in this file" — intentional for Phase 1.

These stubs do not prevent Phase 1 goals (skeleton TUI with navigation, help, and status bar).

## Threat Surface Scan

No new network endpoints, auth paths, file access patterns, or schema changes were introduced beyond what the plan's `<threat_model>` covers. The `T-01-09` mitigation (hard exit before TUI on missing sops) is implemented as specified in `main.go`.

## Self-Check: PASSED

All files found. Both task commits verified. Binary builds. All 4 test packages pass.
