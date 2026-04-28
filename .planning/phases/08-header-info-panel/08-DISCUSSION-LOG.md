# Phase 8: Header Info Panel + Crumb Chips - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in 08-CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-28
**Phase:** 08-header-info-panel
**Areas discussed:** Info-panel schema, Chip pill design, Status-bar shrink, Data refresh + age key

---

## Info-Panel Schema

### Q1: Info-panel row label style — terse k9s-skin convention or verbose for readability?

| Option | Description | Selected |
|--------|-------------|----------|
| Terse 3-char | `cfg:` / `age:` / `rcp:` / `git:` / `fil:` — fits 38-col envelope cleanly, leaves max room for values. Matches k9s skin frame.menu.keyColor convention. Aligned via fixed 5-col label column. | ✓ (recommended) |
| Verbose full-word | `Config:` / `Age:` / `Recipients:` / `Git:` / `Files:` — more readable for new users, but eats 11 cols of label leaving only ~25 for values. May force aggressive value truncation. | |
| Sentence-case shortened | `Sops:` / `Age:` / `Keys:` / `Git:` / `Files:` — middle ground: 5-col aligned, capital-leading, distinct from k9s. | |

**User's choice:** Terse 3-char (recommended)
**Notes:** Matches k9s skin frame.menu.keyColor terseness convention; preserves max value cells in the 38-col envelope (5 label + 1 sep + 32 value). Captured as D-201.

### Q2: What does the `recipients` row count?

| Option | Description | Selected |
|--------|-------------|----------|
| Current file | `len(m.currentParsed.Metadata.AgeRecipients)` — value changes as user navigates. Mirrors k9s ClusterInfo context-current pattern. Shows `—` while in stateFileList. | ✓ (recommended) |
| Aggregate distinct | Count of unique age recipients across all .sops.yaml rules — stable across navigation; project-wide view. | |
| Sum of recipient slots | Total recipients across all files (with duplicates) — matches bulk-rekey ergonomics. | |

**User's choice:** Current file (recommended)
**Notes:** Captured as D-202. Mirrors k9s ClusterInfo's context-currency pattern; `-` placeholder when no current file (D-204).

### Q3: When the .sops.yaml path or age fingerprint exceeds the value column, how should it truncate?

| Option | Description | Selected |
|--------|-------------|----------|
| Right-trunc + ellipsis | `secrets/prod/very-deep/.../trunc…` — truncate from right, append `…`. Simplest; loses filename context. | |
| Middle-trunc + ellipsis | `secrets/.../prod.yaml` — preserve start (repo) and end (filename); ellipsis the middle. k9s does this. Best preserves identifiable content. | ✓ (recommended) |
| Left-trunc + ellipsis | `…/very/deep/prod.yaml` — preserve filename; drop directory prefix. | |

**User's choice:** Middle-trunc + ellipsis (recommended)
**Notes:** Captured as D-203. Preserves both repo-root + filename anchors. Requires extending `TestChromeASCIIOnly` allowlist for U+2026.

### Q4: Marker for an info-panel row whose value isn't computed yet?

| Option | Description | Selected |
|--------|-------------|----------|
| ASCII `-` | Single hyphen-minus. Already inside chrome ASCII allowlist. Reads as 'absence' without confusing as error state. | ✓ (recommended) |
| Word `none` | Lowercase `none` in muted color. More explicit; longer. | |
| Em-dash `—` | U+2014. Visually cleaner than ASCII hyphen. Requires allowlist extension. | |

**User's choice:** ASCII `-` (recommended)
**Notes:** Captured as D-204. No grep-gate change needed; precedent: v1.0 already uses `-` for absence in detail/metadata views.

---

## Chip Pill Design

### Q1: Crumb chip wrapper characters — which bracket style?

| Option | Description | Selected |
|--------|-------------|----------|
| `<segment>` k9s exact | Angle brackets, no leading/trailing whitespace inside. Verbatim k9s `crumbs.go:62-74`. | ✓ (recommended) |
| `< segment >` padded angles | Reads more pill-like; 2 extra cells per chip; overflows narrow widths sooner. | |
| `[segment]` square brackets | Reads as 'tag' rather than 'pill'. Diverges from k9s. | |

**User's choice:** `<segment>` k9s exact (recommended)
**Notes:** Captured as D-205. Project memory specifies k9s parity as a hard quality attribute; this is the canonical match.

### Q2: How should the active (last) crumb be visually distinct from history?

| Option | Description | Selected |
|--------|-------------|----------|
| Bg+fg accent + bold | Background = ColorAccent, foreground inverted, bold weight. Two-channel encoding survives 16-color downsampling per Pitfall 5/9. | ✓ (recommended) |
| Bg+fg accent only (k9s exact) | Background swap only. Works in TrueColor; fails on 16-color. | |
| Underline + bold, no bg swap | Works in monochrome; loses the 'pill' feel. | |

**User's choice:** Bg+fg accent + bold (recommended)
**Notes:** Captured as D-206. Bold is the redundant encoding channel that satisfies Pitfall 9 (color-only is banned); the only deliberate divergence from k9s exact in Phase 8.

### Q3: How should crumb segment text be normalised?

| Option | Description | Selected |
|--------|-------------|----------|
| Lowercase + strip-spaces | `strings.ToLower + ReplaceAll(" ","")` per k9s `crumbs.go:70-71`. `Secret Health` → `<secrethealth>`. | ✓ (recommended) |
| Lowercase + dash-join | `Secret Health` → `<secret-health>`. More readable; diverges from k9s. | |
| Preserve case + strip-spaces | `<SecretHealth>` PascalCase. No project segments have spaces, so moot. | |

**User's choice:** Lowercase + strip-spaces (recommended)
**Notes:** Captured as D-207. Centralised in `RenderCrumbs` so existing 16 SetBreadcrumb call-sites stay untouched; the git status badge `[M]/[A]/[?]` survives normalisation as `[m]/[a]/[?]`.

### Q4: Inter-chip separator and outer padding for the crumb row?

| Option | Description | Selected |
|--------|-------------|----------|
| Single space between, 1-cell row pad | `<files> <prod.yaml> <metadata>` — single space between; 1 cell row padding. k9s `crumbs.go:32` `SetBorderPadding(0,0,1,1)`. | ✓ (recommended) |
| No separator, no row pad | Chips touch each other and edges. Saves cells but cramped. | |
| Two spaces between, 2-cell row pad | More breathing room; overflows narrow widths sooner. | |

**User's choice:** Single space between, 1-cell row pad (recommended)
**Notes:** Captured as D-208. Width budget for overflow check is `m.width - 2` (row pad).

---

## Status-Bar Shrink

### Q1: The center section currently shows `12 items` / `47 commits` etc. Where does it go?

| Option | Description | Selected |
|--------|-------------|----------|
| Drop entirely | Titled-border title already shows the count per Phase 7 D-15. Drop the center section. | ✓ (recommended) |
| Keep in status bar center | Conflicts with the 'shrink to env + clipboard only' goal of UI-08. Two sources of same number. | |
| Move to crumb row right-end | Crumb row mixes navigation + count; harder to scan. Diverges from k9s. | |

**User's choice:** Drop entirely (recommended)
**Notes:** Captured as D-209. `SetItemCount` becomes a no-op or is removed; Plan 2 author confirms no test references and picks.

### Q2: After the breadcrumb leaves the status bar, where does the data live?

| Option | Description | Selected |
|--------|-------------|----------|
| Keep on StatusBarModel, expose `Segments() []string` | 16 existing SetBreadcrumb call-sites untouched. Add `Segments()` accessor. Status bar `View()` stops rendering breadcrumb. Minimal churn. | ✓ (recommended) |
| Move ownership to AppModel | New field on AppModel; rename `SetBreadcrumb`; migrate 16 call-sites. Big-bang refactor. | |
| Hybrid: AppModel proxies through StatusBar | Avoids big-bang refactor but two paths to same data. | |

**User's choice:** Keep on StatusBarModel, expose `Segments() []string` (recommended)
**Notes:** Captured as D-210. Minimum-migration choice; zero call-site churn.

### Q3: When only env + clipboard remain on the right, how is the rest of the status bar treated?

| Option | Description | Selected |
|--------|-------------|----------|
| Right-align, drop pipes, no fill | Status bar still spans full width with surface bg; only right cluster renders. Pipes deleted. | ✓ (recommended) |
| Center the right cluster | Diverges from UI-08's explicit 'right-aligned'. | |
| Keep `|` separator before clipboard | Slight readability win; keeps a piece of v1.0 separator pattern. | |

**User's choice:** Right-align, drop pipes, no fill (recommended)
**Notes:** Captured as D-211. Surface bg still spans full width for visual continuity with v1.0 status bar.

### Q4: Flash messages currently replace all 3 sections with full-width centered text. After shrink?

| Option | Description | Selected |
|--------|-------------|----------|
| Unchanged — full-width center on flash | When `m.flash != ""`, status bar still renders full-width centered flash. Crumb row stays visible above. | ✓ (recommended) |
| Flash overlays crumb row instead | Closer to user's eye when scanning content. More complex. | |
| Flash inline-right of env indicators | Avoids hiding env state. Visually crowded; flash messages can be 30+ chars. | |

**User's choice:** Unchanged — full-width center on flash (recommended)
**Notes:** Captured as D-212. Behavior identical to v1.0; generation-counter handling preserved.

---

## Data Refresh + Age Key

### Q1: Where does the cached `InfoPanelData` live and when does it refresh?

| Option | Description | Selected |
|--------|-------------|----------|
| AppModel field, refresh on events | New `infoPanel ui.InfoPanelData` field. Refreshed on FilesDiscoveredMsg / GitStatusMsg / recipient ops / edit success. Per Pitfall 15 (no per-frame stat). | ✓ (recommended) |
| Per-frame computation in View() | Simpler (no cache invalidation); violates Pitfall 15. Could regress 50µs budget. | |
| Lazy `infoPanelDirty bool` flag | More complex than event-driven; only buys deferred work in narrow paths. | |

**User's choice:** AppModel field, refresh on events (recommended)
**Notes:** Captured as D-213. View() must NOT call os.Stat / parser / git in chrome path — reads cache only.

### Q2: Where does the age fingerprint shown in the `age:` row come from?

| Option | Description | Selected |
|--------|-------------|----------|
| Parse user's keys.txt with filippo.io/age | `age.ParseIdentities` on `~/.config/sops/age/keys.txt`; first identity's `Recipient().String()`; truncate ≤10 chars + `…`. filippo.io/age v1.3.1 already in go.mod. Closest to UI-04 'age key fingerprint' wording. | ✓ (recommended) |
| Show first recipient from .sops.yaml | Mismatches UI-04 wording (would be a *file's* fingerprint, not the *user's*). | |
| Show count of identities only | Trivial; no PII risk; deviates from UI-04 spec. | |

**User's choice:** Parse user's keys.txt with filippo.io/age (recommended)
**Notes:** Captured as D-214. Matches UI-04 owner-coded interpretation; reuses existing dep + import precedent in `recipientform.go:22`.

### Q3: How does the `git:` row fetch branch + dirty marker?

| Option | Description | Selected |
|--------|-------------|----------|
| New `git.GetBranch(repoRoot)` helper | One additive function in `internal/git/status.go` using `repo.Head()`. Dirty derived from existing `m.files[*].GitStatus` aggregate. No new async; no per-frame stat. | ✓ (recommended) |
| Shell out to `git rev-parse --abbrev-ref HEAD` | Inconsistent with v1.0 go-git decision. New external-binary dep. Reject. | |
| Read .git/HEAD file directly | Bypasses go-git abstraction. Reject for consistency. | |

**User's choice:** New `git.GetBranch(repoRoot)` helper (recommended)
**Notes:** Captured as D-215. Returns `(branch string, detached bool, err error)`; called per GitStatusMsg cycle.

### Q4: When RenderChrome drops to mid-tier or narrow-tier (Phase 7.1 D-116), what happens to the crumb row?

| Option | Description | Selected |
|--------|-------------|----------|
| Crumbs render normally; ellipsis-middle on overflow | Crumb row independent of chrome tier; always renders if `crumbsHeight > 0`. Width-overflow drops middle segments and inserts `<…>`. | ✓ (recommended) |
| Hide crumb row in narrow tier | User loses navigation context at the worst moment. | |
| Crumbs always render; clip from the left on overflow | Loses project-root context (`<sops-tui>`). | |

**User's choice:** Crumbs render normally; ellipsis-middle on overflow (recommended)
**Notes:** Captured as D-216. Pitfall 14 mitigation; first + active chip always visible.

---

## Closing Decision

### Q: Ready for CONTEXT.md or explore more gray areas?

| Option | Description | Selected |
|--------|-------------|----------|
| Ready for context | All 4 selected gray areas have locked decisions. | ✓ (recommended) |
| Explore more gray areas | Grep-gate file-list extension, perf budget interaction, plan split, security review gate. | |

**User's choice:** Ready for context (recommended)
**Notes:** Phase 8 CONTEXT.md authored 2026-04-28 with 20 decisions across the 4 gray areas + plan split + validation + security review.

---

## Claude's Discretion

Items the user explicitly deferred to the planner / executor (see CONTEXT.md §Claude's Discretion):

- Exact byte-layout of `RenderInfoPanel` rows (`JoinHorizontal` vs `fmt.Sprintf` padding).
- `middleTruncate` algorithm specifics (split-point bias, ellipsis byte-position handling).
- `truncateSegmentsToWidth` algorithm specifics (binary search vs linear-scan).
- Whether the git status badge `[M]/[A]/[?]` survives `currentFileBreadcrumb()` chip rendering — keep recommended.
- `SetItemCount` deletion vs no-op — delete recommended once Plan 2 confirms no test refs.
- Age key parser file location (`internal/ui/agekey.go` vs inline in `model.go`) — separate file recommended for testability.
- Exact dirty-marker glyph (`*` recommended, no marker for clean).
- `crumbsHeight` return strategy (`lipgloss.Height(rendered)` recommended for consistency with `chromeHeight`).

## Deferred Ideas

- Multi-cluster / multi-context analogue (no analogue today; deferred indefinitely).
- "Press c to copy fingerprint" keybinding (Pitfall 11 explicitly bans).
- Polling goroutine that periodically re-stats keys.txt (rotation mid-session is rare).
- Mouse interactions on chips (keyboard-only by core value).
- All Phase 9, 10, 11, v2 items per CONTEXT.md §deferred.
