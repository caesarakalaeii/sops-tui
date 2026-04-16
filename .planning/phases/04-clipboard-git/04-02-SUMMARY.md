---
phase: 04-clipboard-git
plan: "02"
subsystem: git-integration
tags: [git, badges, go-git, async, status-bar, file-list, detail-view]
dependency_graph:
  requires: [04-01]
  provides: [git-backend, git-badges, async-git-status, git-breadcrumb]
  affects:
    - internal/git/status.go
    - internal/git/status_test.go
    - internal/sops/discoverer.go
    - internal/ui/filelist.go
    - internal/ui/detail.go
    - internal/ui/statusbar.go
    - internal/app/model.go
    - go.mod
tech_stack:
  added: [github.com/go-git/go-git/v5 v5.17.0 (direct)]
  patterns:
    - go-git v5 Worktree.Status() with direct map access (not status.File())
    - async tea.Cmd git status fetch dispatched from FilesDiscoveredMsg
    - generation-counter pattern reused from flash/clipboard for git refresh
    - TDD red-green cycle for git backend
key_files:
  created:
    - internal/git/status.go
    - internal/git/status_test.go
  modified:
    - internal/sops/discoverer.go
    - internal/ui/filelist.go
    - internal/ui/detail.go
    - internal/ui/statusbar.go
    - internal/app/model.go
    - internal/ui/detail_test.go
    - internal/ui/detail_reveal_test.go
    - go.mod
decisions:
  - "Direct map access over status.File() for go-git: status.File() auto-creates entries with Untracked/Untracked for unknown paths, causing clean committed files to appear untracked. Direct map[path] lookup treats missing entries as GitStatusClean."
  - "gitStatus param added to NewDetailModel rather than computed in View(): detail model stores the value for the GitStatus() getter, breadcrumb generation happens in AppModel"
  - "Env()/SetEnv() accessor pair on StatusBarModel: cleaner than exposing the env field directly, consistent with existing SetClipboardHot/IsClipboardHot pattern"
  - "RelativeTime exported as var (not func) for test access without import cycles"
metrics:
  duration_minutes: 20
  completed: "2026-04-15T08:15:30Z"
  tasks_completed: 2
  files_modified: 8
---

# Phase 4 Plan 2: Git Backend and Change Badges Summary

go-git v5 backend with IsGitRepo/GetFileStatuses/GetFileHistory, async git status fetch on startup and after writes, [M]/[A]/[?] badges in file list and detail breadcrumb, and "no git" status bar indicator for non-git repos.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 (RED) | Failing git backend tests | 79c6864 | internal/git/status_test.go, go.mod |
| 1 (GREEN) | Git backend package implementation | 694f30b | internal/git/status.go, go.mod, go.sum |
| 2 | Git badge rendering and model wiring | aa855a5 | discoverer.go, filelist.go, detail.go, statusbar.go, model.go, test files |

## What Was Built

### Task 1: Git Backend Package (TDD)

`internal/git/status.go` provides:

- `IsGitRepo(dir)` — DetectDotGit-based check, safe on any directory
- `GetFileStatuses(repoRoot, relPaths)` — returns `map[string]GitStatus` keyed by relative path. Non-git dirs return empty map + nil error (D-12). Uses direct map lookup instead of `status.File()` to correctly identify clean committed files (see Deviations).
- `GetFileHistory(repoRoot, relPath, limit)` — per-file commit log with `storer.ErrStop` limit enforcement
- `relativeTime(t)` — human-readable durations: "just now", "3 minutes ago", "2 days ago", etc.
- `CommitEntry` struct with ShortHash (7 chars), RelDate, Author, Subject

`go-git/go-git/v5 v5.17.0` promoted from indirect to direct in go.mod.

9 tests covering all behaviors: non-git dirs, modified/untracked/clean files, history entry fields, limit enforcement, relative time formatting.

### Task 2: Git Badge Rendering and Model Wiring

**`internal/sops/discoverer.go`**: `GitStatus string` field added to `DiscoveredFile`. Populated by AppModel after async git status fetch.

**`internal/ui/filelist.go`**: `GitStatus string` field added to `FileItem`. `Title()` appends `[M]`/`[A]`/`[?]` badges using `BadgeModified`/`BadgeAdded`/`BadgeUntracked` styles from styles.go (added in Plan 01).

**`internal/ui/detail.go`**: `gitStatus string` field added to `DetailModel`. `NewDetailModel` signature extended with `gitStatus string` parameter. `GitStatus()` getter exported.

**`internal/ui/statusbar.go`**: `Env()` and `SetEnv()` accessors added for async git availability update from AppModel.

**`internal/app/model.go`**:
- `GitStatusMsg` type with `Statuses`, `GitAvailable`, `Err` fields
- `gitRepoRoot` field on `AppModel`
- `gitpkg` import alias for `internal/git`
- `path/filepath` import added
- `FilesDiscoveredMsg` handler dispatches async `gitCmd` that calls `gitpkg.IsGitRepo` then `gitpkg.GetFileStatuses`
- `GitStatusMsg` handler: updates `env.GitAvailable` in status bar, propagates statuses to `m.files`, rebuilds file list items with badge data
- `currentFileBreadcrumb()` helper appends `[M]`/`[A]`/`[?]` suffix to filename for all breadcrumb calls (5 sites updated)
- `ReEncryptDoneMsg` handler dispatches git status refresh on success (D-11)
- `Enter/l` handler carries `GitStatus` from `FileItem` to `currentFile`

**Test fixes**: All `NewDetailModel` call sites in `detail_test.go` (11 calls) and `detail_reveal_test.go` (11 calls) updated for the new signature.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] go-git status.File() auto-creates Untracked entries for clean files**

- **Found during:** Task 1 GREEN phase — `TestGetFileStatuses/clean_committed_file_returns_GitStatusClean` failed
- **Issue:** `go-git`'s `Status.File(path)` auto-creates a new `FileStatus{Staging: Untracked, Worktree: Untracked}` entry when the path is not in the status map. A committed, unchanged file has no entry in the map, so `File()` would return `Untracked` — incorrectly reporting a clean file as untracked.
- **Fix:** Replaced `status.File(slashPath)` with direct map lookup `fs, ok := status[slashPath]`. Missing entry → `GitStatusClean`. Present entry → switch on Worktree/Staging codes as before.
- **Files modified:** `internal/git/status.go`
- **Commit:** 694f30b

**2. [Rule 3 - Blocking] `commitFile` test helper rejected identical-content commits**

- **Found during:** Task 1 GREEN phase — `TestGetFileHistory/limit_parameter_is_respected` failed with "cannot create empty commit: clean working tree"
- **Issue:** The helper wrote the same content ("data") three times to the same file; go-git correctly rejects the 2nd and 3rd commits as empty.
- **Fix:** Used distinct content strings ("version1", "version2", "version3") for each iteration.
- **Files modified:** `internal/git/status_test.go`
- **Commit:** 694f30b

**3. [Rule 3 - Blocking] `NewDetailModel` signature change broke all existing test call sites**

- **Found during:** Task 2 verification — `go test ./...` reported 22 build failures across `detail_test.go` and `detail_reveal_test.go`
- **Issue:** Adding `gitStatus string` as 6th parameter to `NewDetailModel` broke all pre-existing test calls with 5 arguments.
- **Fix:** Updated all 22 call sites to pass `""` (empty string = clean/no-git) as the 6th argument.
- **Files modified:** `internal/ui/detail_test.go`, `internal/ui/detail_reveal_test.go`
- **Commit:** aa855a5

## Known Stubs

None — git status is fully wired from go-git through to UI badges and breadcrumbs.

## Threat Flags

No new security-relevant surface introduced beyond what the plan's threat model covers. Git status runs asynchronously (T-04-06 mitigation: never blocks UI thread). Non-git directories return empty map and nil error (T-04-08 mitigation: explicit ErrRepositoryNotExists check).

## Self-Check: PASSED

- internal/git/status.go: FOUND
- internal/git/status_test.go: FOUND
- 04-02-SUMMARY.md: FOUND
- commit 79c6864 (RED tests): FOUND
- commit 694f30b (git backend): FOUND
- commit aa855a5 (badge wiring): FOUND
- go build ./...: OK
- go test ./...: all 7 packages pass
