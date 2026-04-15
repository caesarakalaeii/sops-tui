# Phase 5: Power Features - Context

**Gathered:** 2026-04-15
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 5 delivers recipient management (view, add, remove age recipients per file), bulk re-key across multiple files with per-file confirmation, and a secret health check system that reports weak secrets, duplicate values across files, and stale (unchanged) secrets. These are the highest-risk multi-file operations in the TUI.

</domain>

<decisions>
## Implementation Decisions

### Recipient Management — Add Flow
- **D-01:** Add-recipient uses a modal form overlay (huh/v2) — text input for the age public key string. Consistent with the overlay pattern used throughout (help, metadata, diff, history).
- **D-02:** Age public key is validated client-side before submitting to sops — check `age1...` prefix and key length using `filippo.io/age` library (already in go.mod). Prevents typos before the expensive sops re-encrypt operation.

### Recipient Management — Remove Flow
- **D-03:** Remove-recipient shows a numbered list of current recipients (from `parser.SopsMetadata.AgeRecipients`), user selects which to remove by number. Clear and explicit.

### Recipient Management — Confirmation
- **D-04:** After adding or removing a recipient, show a diff-style confirmation overlay listing which recipients will be added/removed. User must explicitly confirm before sops re-encrypts the file. Reuses the Phase 3 diff confirmation pattern.

### Bulk Re-Key
- **D-05:** File selection uses toggle in the file list — Space key toggles selection on individual files. Selected files get a visual indicator (e.g., checkbox or highlight). A dedicated key (e.g., `K`) triggers bulk re-key on all selected files.
- **D-06:** Bulk re-key uses per-file confirmation — each file shows its recipient diff individually, user confirms each one. Safest approach for the "highest-risk multi-file operation".
- **D-07:** Progress displays in the status bar as "Re-keying 3/12: secrets/api.yaml" — updates as each file completes. Uses existing flash/status bar pattern.

### Health Check — Criteria
- **D-08:** Weak secret detection uses both length + format checks. Baseline: flag values under 16 chars or with low Shannon entropy. Format-aware: validate against expected format when key name hints at type (e.g., `_token`, `_key`, `_secret` suffix → check base64 validity, UUID format, etc.).
- **D-09:** Duplicate detection uses exact decrypted value matching across all files. Requires decrypting all files (user must confirm the decrypt-all operation before the scan begins).
- **D-10:** Staleness detection uses git last-modified age via `go-git` (Phase 4 backend). Flag files not modified in N days — configurable, default 90 days.

### Health Check — Triggering & Display
- **D-11:** Health check is triggered on-demand via a dedicated keybinding (e.g., `H`) from the file list. Not automatic on startup.
- **D-12:** Results display in a dedicated full-screen health overlay — findings grouped by category (weak/duplicate/stale) with severity indicators. Scrollable, follows the metadata overlay pattern.

### Claude's Discretion
- Key binding choices for bulk re-key trigger and health check trigger (e.g., `K` vs `B` for bulk, `H` for health)
- Health overlay layout and formatting details
- Shannon entropy threshold for "weak" classification
- How to handle health check when some files can't be decrypted (missing key)
- Toggle selection visual indicator style
- Whether duplicate scan shows which files share the value or just flags them
- Staleness threshold configuration mechanism (env var, flag, or hardcoded default)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Documentation
- `.planning/PROJECT.md` — Project vision, constraints, key decisions (subprocess sops, age-only v1)
- `.planning/REQUIREMENTS.md` — Full v1 requirements with traceability (RCP-01, RCP-02, RCP-03, HLT-03 are Phase 5)
- `.planning/ROADMAP.md` — Phase structure, success criteria, dependencies

### Prior Phase Context
- `.planning/phases/01-foundation/01-CONTEXT.md` — D-05 single-pane drill-down, D-10/D-11/D-12 status bar and flash messages
- `.planning/phases/03-write-loop/03-CONTEXT.md` — D-01/D-02 reveal pattern, diff confirmation overlay, sops subprocess patterns
- `.planning/phases/04-clipboard-git/04-CONTEXT.md` — D-09/D-10 badge patterns, D-13/D-14 overlay patterns, go-git backend usage

### Technology Stack
- `CLAUDE.md` §Technology Stack — filippo.io/age v1.3.1 (key validation), go-git/go-git v5 (staleness), huh/v2 (modal forms)
- `CLAUDE.md` §Bubbletea v2 migration — `View()` returns `tea.View`, `tea.KeyPressMsg`, `msg.Code`/`msg.Text`/`msg.Mod` changes

### External References
- filippo.io/age API: `https://pkg.go.dev/filippo.io/age` — `age.ParseX25519Recipient()` for key validation
- huh v2 form API: `https://pkg.go.dev/github.com/charmbracelet/huh/v2` — form-based input for add-recipient
- SOPS recipient management: `sops updatekeys` and `sops -r` for re-keying operations

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/parser/yaml.go` — `SopsMetadata.AgeRecipients` already parsed from YAML sops block
- `internal/ui/metadata.go` — `MetadataModel` displays recipients. Template for health overlay.
- `internal/sops/executor.go` — SOPS subprocess wrapper with context timeout. Used for re-key operations.
- `internal/ui/styles.go` — Full design system with Phase 4 styles. Phase 5 adds health check styles.
- `internal/app/model.go` — `sessionState` enum, async message patterns, confirmation overlay wiring.
- `internal/git/status.go` — `GetFileHistory()` provides commit timestamps for staleness detection.
- `internal/ui/filelist.go` — `FileItem` with badges and toggle selection target.
- `internal/keys/bindings.go` — Key binding registry. Phase 5 adds bulk re-key and health check bindings.
- `filippo.io/age` — Already in go.mod. `age.ParseX25519Recipient()` validates age public keys.

### Established Patterns
- **sessionState enum** — `stateHealth` and `stateRecipientForm` follow existing overlay pattern.
- **Async msg pattern** — Health check results and re-key completion follow `FilesDiscoveredMsg`/`DecryptKeyMsg` pattern.
- **Flash messages** — Progress counter updates use `m.status.Flash("Re-keying 3/12: file.yaml")`.
- **Confirmation overlay** — Phase 3 diff confirmation pattern is the template for recipient change confirmation.
- **Sops subprocess** — `exec.CommandContext` with timeout for all sops operations including `sops updatekeys`.

### Integration Points
- `FileItem` struct — needs `Selected bool` field for toggle selection
- `FileItem.Title()` — needs selection indicator rendering
- `sessionState` enum — needs `stateHealth`, `stateRecipientForm`, `stateRecipientConfirm`
- `keys.FileListKeyMap` — needs `ToggleSelect` and `BulkReKey` and `HealthCheck` bindings
- `keys.DetailKeyMap` — needs `AddRecipient` and `RemoveRecipient` bindings
- `StatusBarModel` — needs progress counter for bulk operations

</code_context>

<specifics>
## Specific Ideas

- Per-file confirmation for bulk re-key prioritizes safety over speed — this is security-sensitive tooling
- `filippo.io/age` for client-side validation catches typos before the expensive sops re-encrypt round-trip
- Shannon entropy + format-aware checks give two layers of weak secret detection without being overly aggressive
- Exact value matching for duplicates requires decrypt-all, so user explicitly confirms the security trade-off
- Git last-modified from go-git (Phase 4) is already available — staleness detection is low-effort
- Health overlay grouped by category (weak/duplicate/stale) gives a dashboard-style view familiar from security audit tools

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 05-power-features*
*Context gathered: 2026-04-15*
