# Phase 4: Clipboard & Git - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-15
**Phase:** 04-clipboard-git
**Areas discussed:** Clipboard copy flow, Auto-clear & signal safety, Git change badges, Git history view

---

## Clipboard Copy Flow

| Option | Description | Selected |
|--------|-------------|----------|
| c (lowercase) | Matches vim yank convention loosely. Simple, single key. Available. | |
| y (lowercase) | Direct vim yank mapping. Familiar to vim users. Also available. | |
| ctrl+y | Chord avoids accidental copy. Slightly less discoverable but safer. | |

**User's choice:** ctrl+y
**Notes:** Security-sensitive action warrants a chord to prevent accidental copy.

| Option | Description | Selected |
|--------|-------------|----------|
| Revealed value only | Copy only works on a revealed leaf (r first). Consistent with Phase 3 pattern. | |
| Auto-decrypt and copy | Copy on any encrypted leaf auto-decrypts, copies, then discards. One fewer keystroke. | |
| Both with modifier | c copies revealed value; C auto-decrypts and copies. Consistent with r/R convention. | |

**User's choice:** Revealed value only
**Notes:** Consistent with Phase 3 pattern: must reveal before interacting.

| Option | Description | Selected |
|--------|-------------|----------|
| Flash message only | Status bar flashes "Copied (clears in 30s)". Consistent with existing flash pattern. | |
| Flash + persistent countdown | Flash initially, then persistent countdown in status bar. | |
| You decide | Claude picks. | |

**User's choice:** Flash message only
**Notes:** Minimal, consistent with existing D-12 flash pattern.

| Option | Description | Selected |
|--------|-------------|----------|
| Detail view only | Copy requires navigating into file and revealing. Consistent with other secret keys. | |
| Both views | File list: copies file path. Detail: copies revealed value. | |

**User's choice:** Detail view only
**Notes:** Consistent with all other secret-interaction keys being detail-view-only.

---

## Auto-Clear & Signal Safety

| Option | Description | Selected |
|--------|-------------|----------|
| Hardcoded 30s default | Simple. No config needed for v1. | |
| Environment variable | SOPS_TUI_CLIPBOARD_TIMEOUT. Familiar to CLI users. Falls back to 30s. | |
| Both env var and flag | --clipboard-timeout flag plus env var. Flag wins. Most flexible. | |

**User's choice:** Environment variable
**Notes:** SOPS_TUI_CLIPBOARD_TIMEOUT follows SOPS naming convention.

| Option | Description | Selected |
|--------|-------------|----------|
| Reset timer | New copy replaces old content and restarts countdown from zero. | |
| Cancel old, start new | Explicitly clear old first, write new, start fresh timer. | |
| You decide | Claude picks. | |

**User's choice:** Reset timer
**Notes:** Simple and predictable behavior.

| Option | Description | Selected |
|--------|-------------|----------|
| os/signal goroutine + defer | Register SIGINT/SIGTERM handler. Also defer clipboard clear in main. | |
| Bubbletea quit handler only | Intercept quit messages in event loop. May miss raw SIGINT. | |
| Both layers (Recommended) | os/signal goroutine + Bubble Tea quit handler. Belt and suspenders. | |

**User's choice:** os/signal goroutine + defer
**Notes:** Covers both graceful and forced exit paths.

| Option | Description | Selected |
|--------|-------------|----------|
| No indicator | Flash message is enough. Don't clutter status bar. | |
| Subtle status bar icon | Small indicator while clipboard holds a secret. Disappears after clear. | |

**User's choice:** Subtle status bar icon
**Notes:** Persistent awareness that sensitive data is exposed.

---

## Git Change Badges

| Option | Description | Selected |
|--------|-------------|----------|
| [M] [A] [?] text badges | Standard git status convention. Color-coded. | |
| Colored dot indicators | Small colored dots. Cleaner but less informative. | |
| Icon + letter | Unicode symbols with color. Combines visual weight with readability. | |

**User's choice:** [M] [A] [?] text badges
**Notes:** Matches standard git conventions developers already know.

| Option | Description | Selected |
|--------|-------------|----------|
| File list only | Badges next to filenames in browser. Detail view doesn't need them. | |
| File list + detail header | Badge in file list AND detail view breadcrumb/header. | |
| You decide | Claude picks. | |

**User's choice:** File list + detail header
**Notes:** Persistent reminder of git status while navigating and editing.

| Option | Description | Selected |
|--------|-------------|----------|
| On startup + after writes | Fetch on discovery, refresh after re-encryption. Event-driven. | |
| On every view transition | Re-check each time user returns to file list. More current but adds latency. | |
| On startup only | Single check at launch. Simplest but badges become stale. | |

**User's choice:** On startup + after writes
**Notes:** Event-driven refresh keeps badges accurate without polling overhead.

| Option | Description | Selected |
|--------|-------------|----------|
| No badges, no error | Silently skip git features. Badges don't appear. | |
| Dim status bar note | Subtle "no git" indicator alongside sops/age indicators. | |

**User's choice:** Dim status bar note
**Notes:** Consistent with existing env indicator pattern from Phase 1.

---

## Git History View

| Option | Description | Selected |
|--------|-------------|----------|
| Full-screen overlay | Same pattern as help/metadata overlays. Consistent convention. | |
| Inline panel below detail | Split detail view: top = tree, bottom = history. More contextual. | |
| Separate view via key | Dedicated full-screen view with commit list. Like lazygit log. | |

**User's choice:** Full-screen overlay
**Notes:** Consistent with established overlay convention.

| Option | Description | Selected |
|--------|-------------|----------|
| Hash + date + author + subject | Standard git log format. Compact one-line entries. | |
| Hash + date + author + full diff | Expandable entries with actual changes. Heavier. | |
| You decide | Claude picks detail level. | |

**User's choice:** Hash + date + author + subject
**Notes:** Standard compact format. Scrollable with j/k.

| Option | Description | Selected |
|--------|-------------|----------|
| b (blame/history) | Available. 'b' for blame is vim-fugitive convention. | |
| g (git) | Intuitive 'g for git'. Note: g is GoTop in file list. | |
| ctrl+g | Chord avoids collision. Clear mnemonic. Less discoverable. | |

**User's choice:** b (blame/history)
**Notes:** vim-fugitive convention familiar to vim users.

| Option | Description | Selected |
|--------|-------------|----------|
| Detail view only | Consistent with other file-specific actions. | |
| Both views | File list: history for highlighted file. Detail: history for current file. | |

**User's choice:** Detail view only
**Notes:** Consistent with other file-specific actions.

---

## Claude's Discretion

- Cross-file search implementation details (GIT-03)
- Clipboard indicator icon/symbol choice
- Git history overlay layout and formatting
- go-git repository caching strategy
- Auto-clear timer implementation approach
- Badge positioning relative to filename

## Deferred Ideas

None — discussion stayed within phase scope.
