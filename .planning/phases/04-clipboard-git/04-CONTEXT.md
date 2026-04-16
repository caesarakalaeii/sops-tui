# Phase 4: Clipboard & Git - Context

**Gathered:** 2026-04-15
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 4 delivers clipboard copy with auto-clear and signal-safe cleanup, git change badges on the file browser, and a git history overlay for secret files. Users can safely copy decrypted values to the clipboard with guaranteed cleanup on timeout or exit, see uncommitted change indicators alongside their files, and inspect git history per file. Cross-file search (GIT-03) extends the existing fuzzy search to work across all files and key names simultaneously.

</domain>

<decisions>
## Implementation Decisions

### Clipboard Copy Flow
- **D-01:** Copy-to-clipboard is triggered by `ctrl+y` — chord avoids accidental copy of sensitive data. Single keypress would be too easy to hit by mistake.
- **D-02:** Copy only works on revealed leaf values — user must reveal with `r` first, then `ctrl+y` to copy. Consistent with Phase 3 pattern where all secret-interaction keys require reveal first. `ctrl+y` on a masked value is a no-op (flash: "Reveal first with r").
- **D-03:** Copy is available in the detail view only — not from the file list. Consistent with all other secret-interaction keys (r, R, e, E, X) being detail-view-only.
- **D-04:** Copy feedback uses flash message: "Copied (clears in 30s)" displayed in the status bar for 2-3 seconds. Follows existing flash pattern from Phase 1 (D-12).

### Auto-Clear & Signal Safety
- **D-05:** Auto-clear timeout defaults to 30 seconds. Configurable via `SOPS_TUI_CLIPBOARD_TIMEOUT` environment variable (integer seconds). Falls back to 30s if unset or invalid.
- **D-06:** If the user copies again before the previous timeout expires, the timer resets — new content replaces old, countdown starts fresh from zero.
- **D-07:** Clipboard cleanup on exit uses `os/signal` goroutine — register SIGINT and SIGTERM handlers via `os/signal.NotifyContext`. Handler clears clipboard before exit. Also `defer clipboard.WriteAll("")` in main as backup. Covers graceful shutdown and forced exit.
- **D-08:** A subtle indicator appears in the status bar while the clipboard holds a secret (e.g., a small icon or dot). Disappears after auto-clear completes. Provides persistent awareness that sensitive data is exposed.

### Git Change Badges
- **D-09:** Git change badges use text format: `[M]` (modified), `[A]` (added), `[?]` (untracked). Color-coded: M = warning/yellow (`ColorWarning`), A = success/green (`ColorSuccess`), ? = muted (`ColorMuted`). Matches standard git status conventions.
- **D-10:** Badges appear in both the file list (next to filename) and the detail view header/breadcrumb. Persistent reminder of git status while navigating.
- **D-11:** Git status is fetched on startup (during file discovery) and refreshed after any write operation (edit, rotate, $EDITOR re-encryption). No polling — event-driven refresh keeps badges accurate without overhead.
- **D-12:** Non-git repos show a subtle "no git" indicator in the status bar (alongside the existing sops/age/.sops.yaml indicators). Git badges simply don't appear. No error, no disruption.

### Git History View
- **D-13:** Git history is a full-screen overlay triggered by `b` key (blame/history mnemonic). Same `sessionState`/`prevState` pattern as help, metadata, and diff overlays. Press `b` or Esc to close.
- **D-14:** History is accessible from the detail view only — consistent with other file-specific actions.
- **D-15:** Each history entry shows: short hash, relative date, author name, commit subject. Standard compact one-line format. Scrollable with j/k navigation.
- **D-16:** `go-git/go-git` v5 provides the git backend — pure Go, no git binary dependency. Used for both git status (badges) and commit history. Already planned in CLAUDE.md technology stack.

### Claude's Discretion
- Cross-file search implementation (GIT-03) — how to aggregate results across files, result presentation format
- Clipboard indicator icon/symbol choice for status bar
- Git history overlay layout and formatting details
- `go-git` repository caching strategy (open once, reuse handle)
- Auto-clear timer implementation (time.AfterFunc vs goroutine + ticker)
- Badge positioning relative to filename in file list delegate

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Documentation
- `.planning/PROJECT.md` — Project vision, constraints, key decisions (subprocess sops, age-only v1)
- `.planning/REQUIREMENTS.md` — Full v1 requirements with traceability (CLB-01, CLB-02, CLB-03, GIT-01, GIT-02, GIT-03 are Phase 4)
- `.planning/ROADMAP.md` — Phase structure, success criteria, dependencies

### Prior Phase Context
- `.planning/phases/01-foundation/01-CONTEXT.md` — D-05 single-pane drill-down, D-10/D-11/D-12 status bar and flash messages
- `.planning/phases/02-read-loop/02-CONTEXT.md` — D-10/D-11/D-12 fuzzy search inline filter
- `.planning/phases/03-write-loop/03-CONTEXT.md` — D-01/D-02 reveal pattern (must reveal before interacting), overlay patterns

### Technology Stack
- `CLAUDE.md` §Technology Stack — atotto/clipboard v0.1.4, go-git/go-git v5.17.2, sahilm/fuzzy v0.1.1
- `CLAUDE.md` §Bubbletea v2 migration — `View()` returns `tea.View`, `tea.KeyPressMsg`, `msg.Code`/`msg.Text`/`msg.Mod` changes

### External References
- atotto/clipboard API: `https://github.com/atotto/clipboard` — `clipboard.WriteAll()`, `clipboard.ReadAll()`
- go-git v5 API: `https://pkg.go.dev/github.com/go-git/go-git/v5` — `PlainOpen`, `Worktree.Status`, `Log`
- Go os/signal: `https://pkg.go.dev/os/signal` — `signal.NotifyContext` for SIGINT/SIGTERM handling

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/ui/statusbar.go` — `StatusBarModel` with flash messages, env indicators. Phase 4 adds clipboard indicator and git indicator.
- `internal/ui/filelist.go` — `FileItem` struct with `Title()` method that appends badges. Phase 4 adds git badge rendering in `Title()`.
- `internal/ui/metadata.go` — `MetadataModel` full-screen overlay. Template for the git history overlay.
- `internal/ui/search.go` — `SearchModel` inline fuzzy search. Phase 4 extends this for cross-file search.
- `internal/keys/bindings.go` — `DetailKeyMap` with Phase 3 keys. Phase 4 adds `Copy` (ctrl+y) and `Blame` (b) bindings.
- `internal/ui/styles.go` — Full design system with `ColorWarning`, `ColorSuccess`, `ColorMuted` for badge colors.
- `internal/app/model.go` — `sessionState` enum, async message patterns (`FilesDiscoveredMsg`, etc.), flash message integration.
- `internal/sops/discoverer.go` — `DiscoveredFile` struct. Phase 4 extends with `GitStatus` field for badge data.
- `atotto/clipboard` — Already an indirect dependency in go.mod.

### Established Patterns
- **sessionState enum** — `stateHistory` follows the `stateHelp`/`stateMetadata`/`stateDiff` pattern with `prevState`.
- **Async msg pattern** — `FilesDiscoveredMsg`, `DecryptKeyMsg` pattern. Git status and clipboard operations follow the same async pattern.
- **Flash messages** — `m.status.Flash("message")` for transient feedback. Copy confirmation uses this.
- **Key routing** — Global keys first in `AppModel.Update()`, then routed to active child. New keys route through detail state.
- **FileItem badges** — `Title()` already appends `[unencrypted]` badge. Git badges follow the same approach.

### Integration Points
- `FileItem` struct — needs `GitStatus string` field (or similar) for badge data
- `FileItem.Title()` — needs to append git badge after filename
- `DetailModel` header rendering — needs git badge in breadcrumb
- `AppModel.Update()` — needs new key routes for `ctrl+y` (copy) and `b` (blame) in `stateDetail`
- `sessionState` enum — needs `stateHistory` for the git history overlay
- `keys.DetailKeyMap` — needs `Copy` and `Blame` bindings
- `StatusBarModel` — needs clipboard indicator state and "no git" env indicator
- `go.mod` — needs `github.com/go-git/go-git/v5` as direct dependency

</code_context>

<specifics>
## Specific Ideas

- `ctrl+y` chord for copy mirrors the security-conscious approach — copying a secret should be deliberate, not accidental
- Clipboard indicator in status bar provides "hot secret" awareness without cluttering the UI
- `[M]`/`[A]`/`[?]` badges match standard git conventions that developers already know
- Git history overlay follows the existing overlay convention (help, metadata, diff) — no new UI paradigm
- `b` for blame is a vim-fugitive convention familiar to vim users
- Signal-safe clipboard cleanup is security-critical — belt and suspenders approach with os/signal + defer
- SOPS_TUI_CLIPBOARD_TIMEOUT env var follows the `SOPS_*` naming convention for discoverability

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 04-clipboard-git*
*Context gathered: 2026-04-15*
