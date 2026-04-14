# Phase 2: Read Loop - Context

**Gathered:** 2026-04-14
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 2 delivers a fully functional read-only file browser: parsing `.sops.yaml` to discover SOPS-managed files, reading encrypted YAML/JSON to extract key names and types without decrypting, displaying SOPS metadata on demand, and providing fuzzy search across file names and key paths. No decryption occurs — all secret values remain masked.

</domain>

<decisions>
## Implementation Decisions

### File Discovery
- **D-01:** Use both approaches combined — parse `.sops.yaml` `creation_rules[].path_regex` entries as primary discovery, then scan matched files for the `sops:` metadata key to confirm actual encryption status.
- **D-02:** Files matching path_regex but lacking the `sops:` marker appear in the file list with a dim `[unencrypted]` badge. Users see the full picture of what SOPS would manage.
- **D-03:** Drilling into an `[unencrypted]` file opens the detail view normally with a banner: "Not yet encrypted — matches .sops.yaml rules". Values are shown in plaintext since there's nothing to mask.

### Key Extraction
- **D-04:** Encrypted leaf values display with type hints from the SOPS envelope: `*** (str)`, `*** (int)`, `*** (bool)`. Users know what kind of value is stored without seeing content.
- **D-05:** The `sops:` metadata key is hidden from the YAML tree entirely. Its information is surfaced through the dedicated metadata panel (DEC-04) instead.
- **D-06:** Non-encrypted values in SOPS files (via `encrypted_regex`/`unencrypted_regex`) display their actual plaintext value with a subtle `[plain]` badge, making it obvious which keys SOPS left unencrypted.

### Metadata Display
- **D-07:** SOPS metadata (version, lastmodified, recipients, MAC status) is accessed via `i` keypress. Opens a full-screen overlay panel (same pattern as help `?` overlay from Phase 1).
- **D-08:** The metadata panel is accessible from both the file list view (shows metadata for highlighted file) and the detail view (shows metadata for current file). Consistent `i` keybinding in both contexts.
- **D-09:** Metadata panel renders as a full-screen overlay using the existing `sessionState` pattern: `stateMetadata` with `prevState` for return. Esc or `i` closes it.

### Fuzzy Search
- **D-10:** Pressing `/` activates an inline filter input at the bottom of the current view. The list filters in real time as the user types. Esc clears the search and restores the full list. Enter selects the highlighted item.
- **D-11:** Search is context-aware: in the file list, `/` filters file names. In the detail view, `/` filters key paths within the current file. Each view searches its own domain.
- **D-12:** Fuzzy match highlighting uses accent color (`ColorAccent`) on matched characters, leveraging `sahilm/fuzzy` matched character positions. Consistent with the existing design system.

### Claude's Discretion
- `.sops.yaml` parsing edge cases (multiple creation rules, nested configs)
- File tree walking performance strategy (eager vs lazy loading)
- Metadata panel layout and formatting details
- Search input position (top vs bottom of view)
- Empty search results messaging
- Detail view key path flattening for search (e.g., `database.password` vs just `password`)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Documentation
- `.planning/PROJECT.md` — Project vision, constraints, key decisions
- `.planning/REQUIREMENTS.md` — Full v1 requirements with traceability (NAV-01, NAV-02, NAV-04, DEC-03, DEC-04 are Phase 2)
- `.planning/ROADMAP.md` — Phase structure and success criteria
- `.planning/phases/01-foundation/01-CONTEXT.md` — Phase 1 decisions (D-05 single-pane drill-down, D-06/D-07 collapsible YAML tree, D-08/D-09 help overlay pattern)

### Technology Stack
- `CLAUDE.md` §Technology Stack — Bubbletea v2, lipgloss v2, bubbles v2, goccy/go-yaml, sahilm/fuzzy import paths and API notes
- `CLAUDE.md` §Bubbletea v2 migration — `View()` returns `tea.View` struct, `tea.KeyPressMsg`, `msg.Code`/`msg.Text`/`msg.Mod` changes

### Existing Code (Phase 1)
- `internal/validator/startup.go` — `FindSopsYaml()` walks directory tree for `.sops.yaml` — reusable for Phase 2 discovery
- `internal/ui/filelist.go` — `FileListModel` with `FileItem{Name, Path}`, `bubbles/list` wrapper, vim navigation
- `internal/ui/detail.go` — `DetailModel` with `TreeNode{Key, Value, Children, Expanded, Depth}`, collapsible tree rendering
- `internal/ui/styles.go` — Full design system: `ColorAccent`, `DimText`, `TreeConnector`, `TreeIndicator`, `SelectedRow`
- `internal/ui/help.go` — Help overlay pattern (full-screen, `stateHelp`/`prevState`) — template for metadata overlay
- `internal/ui/statusbar.go` — `StatusBarModel` with breadcrumb, item count, env indicators, flash messages
- `internal/app/model.go` — Root `AppModel` with `sessionState` enum routing, `WindowSizeMsg` propagation
- `internal/keys/bindings.go` — Keybinding definitions and key maps

### External References
- SOPS `.sops.yaml` format: `https://github.com/getsops/sops` — creation_rules, path_regex, encrypted_regex/unencrypted_regex
- goccy/go-yaml API: `https://pkg.go.dev/github.com/goccy/go-yaml`
- sahilm/fuzzy API: `https://pkg.go.dev/github.com/sahilm/fuzzy` — matched character positions for highlight rendering

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `validator.FindSopsYaml()` — walks directory tree to find `.sops.yaml`, returns path. Reuse for discovery entry point.
- `ui.FileListModel` — wraps `bubbles/list`, already has `FileItem{Name, Path}`, vim navigation, empty state. Needs real items injected.
- `ui.DetailModel` — collapsible `TreeNode` tree with cursor, scroll, expand/collapse. Needs real YAML data instead of placeholders.
- `ui.StatusBarModel` — breadcrumb, item count, env indicators, flash messages. Already wired in `AppModel`.
- `ui.HelpModel` — full-screen overlay pattern. Template for the new metadata overlay.
- Design system (`styles.go`) — `ColorAccent`, `DimText`, `TreeConnector`, `TreeIndicator`, `SelectedRow` all ready.

### Established Patterns
- **sessionState enum** — `stateFileList`, `stateDetail`, `stateHelp` in `AppModel`. New `stateMetadata` and `stateSearch` follow this pattern.
- **prevState for overlays** — Help uses `prevState` to return to previous view. Metadata overlay will use the same approach.
- **SetSize propagation** — `WindowSizeMsg` propagates dimensions to all children. New components must implement `SetSize`.
- **Key routing** — Global keys (?, q, esc) handled first in `AppModel.Update()`, then routed to active child.
- **Value types** — All models are value types with pointer receiver `SetSize` methods. Copy-on-update pattern.

### Integration Points
- `AppModel.NewAppModel()` — currently creates `FileListModel` with empty items. Phase 2 wires file discovery here.
- `AppModel.Update()` `Enter/l` handler — currently creates `DetailModel` with empty nodes. Phase 2 parses the actual file.
- `sessionState` enum — needs `stateMetadata` (and possibly `stateSearch` or search as a mode within existing states).
- `keys.go` — needs new keybindings for `/` (search), `i` (info/metadata).

</code_context>

<specifics>
## Specific Ideas

- File discovery combines `.sops.yaml` regex matching with `sops:` marker scanning — dual confirmation approach
- Unencrypted files shown with `[unencrypted]` badge rather than hidden — user sees full SOPS management picture
- Type hints on masked values (`*** (str)`, `*** (int)`) — more informative than plain `***`
- Plaintext values in SOPS files get `[plain]` badge — highlights `encrypted_regex`/`unencrypted_regex` behavior
- Metadata overlay reuses the help overlay pattern (full-screen, `sessionState`, `prevState`) — no new UI paradigm
- Fuzzy search is inline filter (k9s-style), not overlay — stays integrated with the current view
- Search is context-aware: file names in file list, key paths in detail view

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 02-read-loop*
*Context gathered: 2026-04-14*
