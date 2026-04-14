# Phase 1: Foundation - Context

**Gathered:** 2026-04-14
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 1 delivers a working TUI skeleton that starts, validates its environment (sops binary, age key, .sops.yaml), and provides vim-style navigation with contextual help and a persistent status bar — without ever touching a secret value. This establishes the application shell that all subsequent phases fill with content.

</domain>

<decisions>
## Implementation Decisions

### Startup Validation
- **D-01:** Use lipgloss-styled stderr for missing-dependency errors — render a colored, bordered box to stderr without initializing a TUI session. Exit with non-zero code. Scriptable and CI-friendly.
- **D-02:** Run all validation checks together and report all issues in a single error box. User fixes everything in one pass.
- **D-03:** `sops` binary missing is a hard error (exit). Age key missing is a soft warning — TUI launches but decryption will be unavailable. This allows browsing encrypted files without a key.
- **D-04:** Validate `.sops.yaml` at startup — check cwd and parent dirs. If missing, show styled warning but still launch (the skeleton has nothing to browse yet).

### Skeleton Layout
- **D-05:** Single-pane layout with drill-down navigation. File list takes the full terminal width. Selecting a file replaces the view with a detail view. Esc returns to the file list. Like k9s resource drill-down.
- **D-06:** Detail view renders content as a YAML-style tree with indentation preserving the original YAML structure. Nested keys displayed as a tree hierarchy.
- **D-07:** Deeply nested YAML nodes are collapsible — groups can be collapsed/expanded with Enter or arrow keys. [+]/[-] indicators show collapse state.

### Help Panel
- **D-08:** Help is a full-screen overlay toggled with `?`. Replaces current content entirely. Press `?` or `Esc` to close. Like k9s help screen.
- **D-09:** Help content is contextual — shows only keybindings relevant to the current view (file list, detail, search). Global keys (`?`, `q`, `/`) appear in every context.

### Status Bar
- **D-10:** Single-line status bar at the bottom of the terminal. Standard TUI convention (k9s, vim, lazygit).
- **D-11:** Status bar shows: left = current path/view breadcrumb, center = item count, right = environment status indicators (sops/age/.sops.yaml availability with checkmarks or warnings). Content adapts per view.
- **D-12:** Transient feedback (e.g., "Copied to clipboard") uses flash messages — temporarily replaces status bar content for 2-3 seconds, then restores normal content. No extra UI chrome needed.

### Claude's Discretion
- Exact lipgloss color palette and border styles
- Internal component architecture and state management patterns
- Flash message duration (2-3 second range)
- YAML tree indentation width and collapse/expand key mappings (h/l or Enter)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Documentation
- `.planning/PROJECT.md` — Project vision, constraints, key decisions
- `.planning/REQUIREMENTS.md` — Full v1 requirements with traceability (HLT-01, HLT-02, NAV-03, NAV-05, NAV-06 are Phase 1)
- `.planning/ROADMAP.md` — Phase structure and success criteria

### Technology Stack
- `CLAUDE.md` §Technology Stack — Bubbletea v2, lipgloss v2, bubbles v2, huh v2 import paths and migration notes
- `CLAUDE.md` §Bubbletea v2 migration — `View()` returns `tea.View` struct, `tea.KeyPressMsg`, `msg.Code`/`msg.Text`/`msg.Mod` changes

### External References
- Bubbletea v2 upgrade guide: `https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md`
- SOPS v3.12.2 documentation: `https://github.com/getsops/sops`
- age key format: `https://github.com/FiloSottile/age`

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- None — greenfield project. Only `go.mod` (module `github.com/caesarakalaeii/sops-tui`, Go 1.26.2) exists.

### Established Patterns
- None yet — Phase 1 establishes all foundational patterns (Bubble Tea model structure, component composition, error handling).

### Integration Points
- `go.mod` already declares the module path — all new packages build from here
- `.gitignore` exists and should be extended for any build artifacts

</code_context>

<specifics>
## Specific Ideas

- Error presentation inspired by lipgloss's ability to render styled boxes to stderr (no TUI session needed)
- Navigation model inspired by k9s: single-pane drill-down, not side-by-side split
- Help overlay inspired by k9s `?` screen
- YAML tree with collapsible nodes inspired by file managers and IDE tree views
- Status bar follows vim/k9s convention: bottom line, contextual information

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 01-foundation*
*Context gathered: 2026-04-14*
