---
phase: 02-read-loop
plan: "03"
subsystem: integration
tags: [app-model, state-machine, file-discovery, search, metadata, wiring]
dependency_graph:
  requires:
    - 02-01 (SopsDiscoverer, YamlParser, TreeNode with Encrypted/TypeHint/IsPlain)
    - 02-02 (MetadataModel, SearchModel, Phase 2 named styles)
  provides:
    - internal/app/model.go (AppModel fully wired: discovery, parsing, metadata overlay, search)
    - internal/keys/bindings.go (Search and Info keybindings on both KeyMaps)
    - internal/ui/filelist.go (FileItem with IsEncrypted/Rule, search mode, unencrypted badge)
    - internal/ui/detail.go (search mode, unencrypted banner, canonical styles, Nodes() accessor)
    - internal/ui/help.go (ViewMetadata added to ViewState enum)
  affects:
    - cmd/sops-tui/main.go (NewAppModel now takes sopsYamlPath)
tech_stack:
  added: []
  patterns:
    - Async tea.Cmd pattern for file discovery (FilesDiscoveredMsg) and parsing (FilesParsedMsg)
    - prevState field for overlay dismiss (help, metadata share same pattern)
    - searchActive mode flag on child models (not a sessionState)
    - keyPath dot-joined string for fuzzy search in DetailModel flatRow
    - DeactivateSearch restores allItems/allFlatRows on dismiss
key_files:
  created: []
  modified:
    - internal/app/model.go
    - internal/app/model_test.go
    - internal/keys/bindings.go
    - internal/keys/bindings_test.go
    - internal/ui/filelist.go
    - internal/ui/filelist_test.go
    - internal/ui/detail.go
    - internal/ui/detail_test.go
    - internal/ui/help.go
    - cmd/sops-tui/main.go
decisions:
  - Metadata overlay opened synchronously (parser.ParseFile called inline on i keypress) to keep state machine simple; latency acceptable for read-only metadata display
  - DeactivateSearch restores full allItems/allFlatRows slice directly rather than re-running discovery/flatten — simpler and avoids re-allocation on every Esc
  - statusBarHeight() helper function computes height dynamically by rendering status bar — avoids hardcoded row count that breaks with multi-line status bars
  - countLeafNodes() helper counts only leaf nodes (no children) for the "N keys" status bar count
  - Esc priority chain: search deactivation > overlay close > navigate back — matches k9s behavior
metrics:
  duration: ~19 minutes
  completed: 2026-04-14
  tasks_completed: 2
  files_changed: 10
---

# Phase 2 Plan 03: Integration Wiring Summary

AppModel fully wired with async SOPS file discovery, YamlParser drill-in, MetadataModel overlay, and SearchModel inline fuzzy filter — connecting all Phase 2 components through the state machine.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | Keybindings, FileItem extension, and AppModel state machine wiring | 4a5f8bb | 10 files modified |
| 2 | Visual verification (auto-approved) | — | — |

## What Was Built

### Keybindings (`internal/keys/bindings.go`)

Two new keybindings added to both `FileListKeyMap` and `DetailKeyMap`:
- `Search` — key `/`, help text "search"
- `Info` — key `i`, help text "file info"

Both `ShortHelp()` and `FullHelp()` updated to include the new bindings in the actions group.

### FileItem Extension + Search Mode (`internal/ui/filelist.go`)

`FileItem` extended with:
- `IsEncrypted bool` — drives `[unencrypted]` badge in `Title()`
- `Rule sops.CreationRule` — passed through to parser on drill-in

`FileListModel` extended with:
- `searchActive bool`, `search SearchModel`, `allItems []FileItem`
- `ActivateSearch()`, `DeactivateSearch()`, `IsSearchActive()` methods
- `Update()` routes all messages to `SearchModel` when search active; applies `ApplyFilter` after each keystroke
- `View()` appends search bar as bottom row when active
- `SelectedFileItem()` method returns full `FileItem` with `Rule` and `IsEncrypted`
- `ItemCount()` now returns `len(m.allItems)` (unfiltered count, stable for status bar)

### DetailModel Search Mode + Unencrypted Banner (`internal/ui/detail.go`)

- Removed `typeHintStyleTemp` and `badgePlainTemp`; `renderRow` now uses canonical `TypeHintStyle` and `BadgePlain` from `styles.go`
- `flatRow` extended with `keyPath string` — dot-joined ancestor key path (e.g., `database.password`) for fuzzy search
- `flattenNodes` accepts `parentKeyPath string` parameter to build key paths recursively
- `DetailModel` extended with `searchActive`, `search SearchModel`, `allFlatRows []flatRow`, `isEncrypted bool`
- `ActivateSearch()`, `DeactivateSearch()`, `IsSearchActive()` methods (same pattern as FileListModel)
- `NewDetailModel` accepts `isEncrypted bool` parameter
- `View()` prepends `WarnLabel.Render("Not yet encrypted \u2014 matches .sops.yaml rules")` + blank line when `!isEncrypted && len(nodes) > 0`
- `Nodes()` accessor method added (used by `countLeafNodes` in AppModel)

### Help Overlay Extension (`internal/ui/help.go`)

`ViewMetadata` added to `ViewState` enum. Falls through to `ViewFileList` for help rendering (metadata overlay has minimal bindings).

### AppModel Wiring (`internal/app/model.go`)

Complete rewrite of AppModel to wire all Phase 2 components:

1. **State enum**: `stateMetadata` added alongside existing `stateFileList`, `stateDetail`, `stateHelp`
2. **Message types**: `FilesDiscoveredMsg{Files, Err}` and `FilesParsedMsg{Parsed, Err}`
3. **New fields**: `metadata MetadataModel`, `sopsYamlPath string`, `files []sops.DiscoveredFile`, `currentFile sops.DiscoveredFile`
4. **`NewAppModel(env, sopsYamlPath)`**: accepts sopsYamlPath for discovery
5. **`Init()`**: runs `tea.Batch` of window size request + async `sops.Discover` goroutine
6. **`FilesDiscoveredMsg` handler**: populates `FileListModel` with real `FileItem` data; updates status bar item count
7. **`FilesParsedMsg` handler**: creates `DetailModel` with parsed nodes and `isEncrypted`; transitions to `stateDetail`
8. **`i` key handler**: opens/closes metadata overlay; calls `parser.ParseFile` synchronously; sets `prevState`
9. **`/` key handler**: delegates to `fileList.ActivateSearch()` or `detail.ActivateSearch()`
10. **Esc priority chain**: search deactivation → overlay close → navigate back
11. **`Enter/l` handler**: now calls `parser.ParseFile` async via `tea.Cmd` instead of creating empty `DetailModel`
12. **`WindowSizeMsg`**: propagates `SetSize` to `metadata` model
13. **`View()`**: `stateMetadata` case renders `m.metadata.View()`
14. **Helper functions**: `statusBarHeight(m AppModel) int` and `countLeafNodes(nodes []TreeNode) int`

### main.go

`main.go` updated to call `validator.FindSopsYaml(opts.StartDir)` and pass the result to `app.NewAppModel(env, sopsYamlPath)`.

## Verification Results

```
go build ./...         — PASS (clean build, no errors)
go vet ./...           — PASS (no warnings)
go test ./... -count=1 — PASS (all tests green)
```

Test counts by package:
- `internal/app`: 13 tests (5 existing + 8 new)
- `internal/keys`: 7 tests (5 existing + 2 new)
- `internal/ui`: 79 tests (all existing + 12 new in detail_test + 5 new in filelist_test)
- `internal/parser`: 15 tests (unchanged)
- `internal/sops`: 10 tests (unchanged)
- `internal/validator`: 9 tests (unchanged)

## Deviations from Plan

None — plan executed exactly as written.

The only implementation note: `ItemCount()` on `FileListModel` was updated to return `len(m.allItems)` rather than `len(m.list.Items())` so the status bar item count remains stable during search filtering. This is the correct behavior (status bar shows total files, not filtered subset) and was implied by the plan's status bar update contracts.

## Known Stubs

None. All components are fully wired with real data:
- File list populated from `sops.Discover` results at startup
- YAML tree populated from `parser.ParseFile` results on drill-in
- Metadata overlay populated from `parser.ParseFile` results on `i` keypress
- Search filters real file names / key paths via `sahilm/fuzzy`

## Threat Mitigations Applied

| Threat | Mitigation |
|--------|------------|
| T-02-09 | Phase 2 read-only; no decryption; ENC values displayed as "*** (type)" only |
| T-02-10 | `sops.Discover` runs as async `tea.Cmd` goroutine; TUI remains responsive during discovery |
| T-02-11 | `parser.ParseFile` 10MB file size guard (from Plan 01) prevents memory exhaustion |

## Self-Check: PASSED
