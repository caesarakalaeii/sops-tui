# Phase 1: Foundation - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-14
**Phase:** 01-foundation
**Areas discussed:** Startup validation UX, Skeleton layout, Help panel behavior, Status bar design

---

## Startup validation UX

| Option | Description | Selected |
|--------|-------------|----------|
| Plain stderr + exit | Print styled but non-TUI error to stderr, exit with non-zero code | |
| Styled TUI error page | Launch full TUI, show styled error screen with lipgloss borders | |
| Lipgloss-styled stderr | Use lipgloss to render colored bordered box to stderr, no TUI session | ✓ |

**User's choice:** Lipgloss-styled stderr
**Notes:** Best of both worlds — visually polished but no TUI session needed, scriptable, clean exit

| Option | Description | Selected |
|--------|-------------|----------|
| Report all at once | Check sops AND age key, show all issues in single error box | ✓ |
| Fail on first | Stop at first missing dependency | |

**User's choice:** Report all at once
**Notes:** User fixes everything in one pass

| Option | Description | Selected |
|--------|-------------|----------|
| Soft warning | Warn about missing age key but start TUI anyway | ✓ |
| Strict exit | No age key = no start | |
| You decide | Claude's discretion | |

**User's choice:** Soft warning
**Notes:** Allows browsing encrypted files without a key; decryption fails gracefully when attempted

| Option | Description | Selected |
|--------|-------------|----------|
| Validate at startup | Check for .sops.yaml, warn if missing, still launch | ✓ |
| Defer to Phase 2 | Let Phase 2 handle when file browser is built | |
| You decide | Claude's discretion | |

**User's choice:** Validate at startup
**Notes:** Establishes config discovery pattern early

---

## Skeleton layout

| Option | Description | Selected |
|--------|-------------|----------|
| Dual pane | Left: file list, Right: detail. Bottom: status bar | |
| Single pane + detail view | Full-width file list, selecting replaces with detail view | ✓ |
| k9s-style stacked | Header, full-width list, bottom split for detail | |

**User's choice:** Single pane + detail view
**Notes:** Like drilling into a resource in k9s — Enter to open, Esc to go back

| Option | Description | Selected |
|--------|-------------|----------|
| Table layout | Key names left, values right in clean table | |
| YAML-style tree | Render actual YAML structure with indentation | ✓ |
| You decide | Claude's discretion | |

**User's choice:** YAML-style tree
**Notes:** Preserves original YAML structure, more familiar to YAML users

| Option | Description | Selected |
|--------|-------------|----------|
| Collapsible nodes | Nested groups collapse/expand with Enter or arrows, [+]/[-] indicators | ✓ |
| Always expanded | Full tree always, rely on scroll + search | |
| You decide | Claude's discretion | |

**User's choice:** Collapsible nodes
**Notes:** Keeps deep files navigable

---

## Help panel behavior

| Option | Description | Selected |
|--------|-------------|----------|
| Full-screen overlay | ? toggles full-screen help replacing current content | ✓ |
| Bottom drawer | Help slides up from bottom, ~40% height, current view stays visible | |
| Inline footer hint | Always-visible one-line hint, ? expands to full help | |

**User's choice:** Full-screen overlay
**Notes:** Like k9s help or vim :help — clean, plenty of room for keybinding groups

| Option | Description | Selected |
|--------|-------------|----------|
| Contextual | Shows only keybindings for current view, global keys in every context | ✓ |
| Global master list | One comprehensive list of all keybindings | |
| You decide | Claude's discretion | |

**User's choice:** Contextual
**Notes:** Less overwhelming, always accurate for the current view

---

## Status bar design

| Option | Description | Selected |
|--------|-------------|----------|
| Bottom | Single line at bottom of terminal | ✓ |
| Top header + bottom | App name at top, status at bottom — two chrome lines | |
| You decide | Claude's discretion | |

**User's choice:** Bottom
**Notes:** Standard TUI convention (k9s, vim, lazygit)

| Option | Description | Selected |
|--------|-------------|----------|
| Location + counts + env status | Left: path/breadcrumb, Center: item count, Right: sops/age/config indicators | ✓ |
| Minimal — path only | Just current file path or view name | |
| k9s-style with mode | Mode indicator (NORMAL, SEARCH, EDIT) plus path and status | |

**User's choice:** Location + counts + env status
**Notes:** Adapts per view, shows environment health at a glance

| Option | Description | Selected |
|--------|-------------|----------|
| Flash message | Temporarily replace status bar content for 2-3 seconds, then restore | ✓ |
| Dedicated toast area | Keep status bar intact, show floating toast in content area | |
| You decide | Claude's discretion | |

**User's choice:** Flash message
**Notes:** Simple, attention-grabbing, no extra UI chrome

---

## Claude's Discretion

- Exact lipgloss color palette and border styles
- Internal component architecture and state management patterns
- Flash message duration (2-3 second range)
- YAML tree indentation width and collapse/expand key mappings

## Deferred Ideas

None — discussion stayed within phase scope
