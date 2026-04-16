# Phase 3: Write Loop - Context

**Gathered:** 2026-04-14
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 3 delivers on-demand decryption, inline value editing, external `$EDITOR` support, format-aware secret rotation, and a diff/confirmation safety gate before any re-encryption occurs. Users can decrypt, reveal, edit, and rotate secrets — with every destructive write requiring explicit confirmation via a diff overlay. No clipboard, git, or recipient management functionality — those belong in later phases.

</domain>

<decisions>
## Implementation Decisions

### Reveal Interaction
- **D-01:** Single value reveal uses dedicated `r` key — pressing `r` on an encrypted leaf decrypts and reveals that value inline. Enter/l remains reserved for expand/collapse on group nodes. `r` on a revealed leaf toggles it back to masked.
- **D-02:** Reveal all values uses `R` (Shift+R) — decrypts all values in the current file with a single `sops -d` call. All leaves update to show decrypted values. `R` again re-masks all.
- **D-03:** Revealed values display as plaintext with a 🔓 icon suffix. The `*** (type)` hint disappears and the actual value is shown in normal text. Clear visual distinction between masked and revealed states.
- **D-04:** Navigating back to the file list (Esc from detail) automatically re-masks all revealed values. No decrypted content persists across view transitions.

### Edit Flow
- **D-05:** Inline single-key editing uses `e` key. Press `e` on a revealed leaf → the value cell becomes an editable text input (bubbles/textinput). Enter confirms the edit, Esc cancels. Value must be revealed first — `e` on a masked value is a no-op (or shows flash: "Reveal first with r").
- **D-06:** Full-file editing uses `E` (Shift+E) key. Suspends TUI, opens the full decrypted file in `$EDITOR`. After save+quit, TUI resumes and detects changes. Value must be revealed first (at least one value, or use `R` for all).
- **D-07:** Re-encryption for inline edits uses `sops set <file> '["key"]["path"]' '"new_value"'` — atomic single-key update, no temp files, SOPS handles re-encryption. Available since sops 3.7+.
- **D-08:** Re-encryption for `$EDITOR` flow uses standard `sops` encrypt on the modified temp file, then replaces the original.

### Diff & Confirmation
- **D-09:** Before any re-encryption, a full-screen diff overlay appears (same pattern as help/metadata overlays via `sessionState`/`prevState`). Shows old value → new value with color coding (red = removed, green = added). For inline edits: single key diff. For `$EDITOR` flow: multi-key scrollable diff showing all changed keys.
- **D-10:** The diff overlay IS the confirmation gate (EDT-04). User presses `y` to confirm and trigger re-encryption, or `n`/Esc to cancel without effect. One screen serves both review and confirmation.
- **D-11:** For `$EDITOR` flow, after the editor exits the TUI compares old vs new decrypted content. All changed keys appear in a scrollable diff overlay. User confirms all changes at once with `y`/`n`.

### Rotation
- **D-12:** Secret rotation uses `X` (Shift+X) key. On a revealed leaf, `X` auto-detects the current value's format and generates a new random value in the same format. The generated value appears in the diff overlay for confirmation before re-encryption.
- **D-13:** Format auto-detection analyzes the current decrypted value: base64-encoded → base64, hex string → hex, UUID pattern → UUID v4, bcrypt hash → bcrypt. When detection is ambiguous, falls back to a format selection menu (base64, hex, UUID, alphanumeric, bcrypt).
- **D-14:** After format selection (auto or manual), the new value is generated and shown in the standard diff overlay with y/n confirmation. Same safety gate as editing.

### Claude's Discretion
- Exact byte lengths for generated values (reasonable defaults: 32 bytes for base64/hex, standard for UUID/bcrypt)
- bcrypt cost factor (standard: 10-12)
- Inline text input styling and cursor behavior
- `$EDITOR` temp file location and naming convention
- Error handling for sops subprocess failures (flash messages with error details)
- Format detection heuristics (regex patterns for base64, hex, UUID, bcrypt)
- Scrollable diff overlay navigation keys (j/k for scrolling within diff)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Documentation
- `.planning/PROJECT.md` — Project vision, constraints, key decisions (subprocess sops, age-only v1)
- `.planning/REQUIREMENTS.md` — Full v1 requirements with traceability (DEC-01, DEC-02, EDT-01, EDT-02, EDT-03, EDT-04 are Phase 3)
- `.planning/ROADMAP.md` — Phase structure, success criteria, dependencies

### Prior Phase Context
- `.planning/phases/01-foundation/01-CONTEXT.md` — D-05 single-pane drill-down, D-08/D-09 overlay pattern, D-12 flash messages
- `.planning/phases/02-read-loop/02-CONTEXT.md` — D-04 type hints on masked values, D-07/D-08/D-09 metadata overlay, D-10/D-11/D-12 fuzzy search

### Technology Stack
- `CLAUDE.md` §Technology Stack — Bubbletea v2, lipgloss v2, bubbles v2, huh v2 import paths and API notes
- `CLAUDE.md` §Bubbletea v2 migration — `View()` returns `tea.View`, `tea.KeyPressMsg`, `msg.Code`/`msg.Text`/`msg.Mod` changes
- `CLAUDE.md` §SOPS Integration — `exec.CommandContext`, capture stderr, `sops set` for single-key updates

### External References
- SOPS `set` command: `https://github.com/getsops/sops` — `sops set <file> '["key"]' '"value"'` for atomic single-key re-encryption
- SOPS decrypt/encrypt: `sops -d` for full-file decrypt, `sops -e` for full-file encrypt
- Bubbletea v2 program suspension: `tea.ExecProcess` for `$EDITOR` subprocess integration

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/app/model.go` — `sessionState` enum with `prevState` overlay pattern. New `stateDiff` and `stateEdit` states follow this pattern.
- `internal/ui/detail.go` — `DetailModel` with `TreeNode`, `flatRow`, cursor navigation, search. Phase 3 adds reveal/edit state to `TreeNode` and new keybindings.
- `internal/ui/metadata.go` — `MetadataModel` full-screen overlay. Template for the diff confirmation overlay.
- `internal/ui/styles.go` — Full design system with `ColorSuccess` (green for added), `ColorError` (red for removed), `ColorAccent`, `DimText`. All needed for diff coloring.
- `internal/parser/yaml.go` — `ParseFile()` returns `ParsedFile{Nodes, Metadata}`. Phase 3 needs a parallel `DecryptFile()` that returns plaintext values.
- `internal/keys/bindings.go` — `DetailKeyMap` with vim navigation. Phase 3 adds `r`, `R`, `e`, `E`, `X` bindings.
- `internal/ui/search.go` — `SearchModel` inline text input. Can inform the inline edit input approach.
- `internal/sops/discoverer.go` — SOPS file discovery. Phase 3 adds `sops` subprocess wrapper for decrypt/encrypt/set operations.
- `internal/validator/startup.go` — Validates sops binary availability. Phase 3 depends on sops being present.

### Established Patterns
- **sessionState enum** — `stateFileList`, `stateDetail`, `stateHelp`, `stateMetadata`. New `stateDiff` (diff overlay), `stateEdit` (inline editing mode) follow this pattern.
- **prevState for overlays** — Help and metadata use `prevState` to return. Diff overlay will use the same approach.
- **Key routing** — Global keys handled first in `AppModel.Update()`, then routed to active child. Phase 3 keys (`r`, `R`, `e`, `E`, `X`) route through detail state.
- **Async operations** — `FilesDiscoveredMsg`, `FilesParsedMsg` pattern with `tea.Cmd` returning messages. Decrypt and re-encrypt operations should follow the same async msg pattern.
- **Flash messages** — `m.status.Flash("message")` for transient feedback. Use for "Decrypted", "Re-encrypted", "Rotation complete", error messages.

### Integration Points
- `AppModel.Update()` tea.KeyPressMsg handler — needs new key routes for `r`, `R`, `e`, `E`, `X` when in `stateDetail`
- `TreeNode` struct — needs `Revealed bool` and `DecryptedValue string` fields for revealed state
- `renderRow()` in `detail.go` — needs branch for `node.Revealed` to show plaintext + 🔓 instead of `*** (type)`
- New `internal/sops/executor.go` (or similar) — subprocess wrapper for `sops -d`, `sops set`, `sops -e`
- New `internal/ui/diff.go` — `DiffModel` full-screen overlay for before/after comparison with y/n confirmation

</code_context>

<specifics>
## Specific Ideas

- Reveal toggle with `r`/`R` mirrors the lowercase/shift convention already established (lowercase = single item, shift = all items)
- Diff overlay follows the lazygit-inspired pattern: review changes before committing them
- `sops set` for inline edits is the cleanest path — atomic, no temp files, single key targeted
- `$EDITOR` flow should feel identical to running `sops <file>` from the command line — familiar to existing SOPS users
- Format auto-detection makes rotation feel "smart" — the tool understands what kind of secret you have
- All destructive operations funnel through the same diff overlay → y/n gate. Consistent safety model.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 03-write-loop*
*Context gathered: 2026-04-14*
