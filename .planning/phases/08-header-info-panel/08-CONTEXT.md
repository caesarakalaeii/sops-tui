# Phase 8: Header Info Panel + Crumb Chips - Context

**Gathered:** 2026-04-28
**Status:** Ready for planning

<domain>
## Phase Boundary

Inflate the 38×6 info-panel slot reserved by Phase 7's chrome with five live data rows; flip `crumbsHeight` from 0 to a real chip-pill row above the titled body; shrink the status bar to right-aligned env + clipboard only. Pure data-binding + presentation — no new state machine, no new keybindings, no functional regressions.

**In scope (this phase):**
- `internal/ui/infopanel.go` — new file. `InfoPanelData` struct + `RenderInfoPanel(d InfoPanelData) string` pure renderer. Five fixed rows (`cfg:` / `age:` / `rcp:` / `git:` / `fil:`).
- `internal/ui/crumbs.go` — new file. `RenderCrumbs(segments []string, width int) string` pure renderer. `<segment>` k9s-exact pill format; lowercase + strip-spaces normalisation; bg+fg accent + bold for active chip; middle-segment ellipsis on width overflow per Pitfall 14.
- `internal/git/status.go` — extend with `GetBranch(repoRoot string) (branch string, detached bool, err error)` helper using `repo.Head()`. No new package; pure additive function.
- `internal/ui/styles.go` — new package vars: `InfoPanelLabelStyle`, `InfoPanelValueStyle`, `InfoPanelSepStyle` (the `:`), `CrumbChipStyle`, `CrumbChipActiveStyle`, `CrumbChipSepStyle`, `CrumbChipEllipsisStyle`. Explicit hex per AdaptiveColor ban.
- `internal/ui/statusbar.go` — add `Segments() []string` accessor returning the underlying breadcrumb segments. Modify `View()` to render env + clipboard right-aligned only; drop left + center sections + pipe separators. `SetItemCount` becomes a no-op (or removed); titled-border title is the canonical count display per Phase 7 D-15.
- `internal/ui/chrome.go` — wire info-panel content. `RenderChrome` gains an `info InfoPanelData` parameter; full-tier renders `RenderInfoPanel(info)` into the 38-col left slot. Mid-tier and narrow-tier still drop the slot (Phase 7.1 D-116 unchanged).
- `internal/app/model.go` — add `infoPanel ui.InfoPanelData` cached field on `AppModel`; populate at startup (.sops.yaml relative path + age fingerprint from `~/.config/sops/age/keys.txt`); refresh on `FilesDiscoveredMsg` (file count), `GitStatusMsg` (git branch + clean/dirty), and successful recipient/edit operations (recipient count). `crumbsHeight()` flipped from `return 0` to real height of `RenderCrumbs(m.status.Segments(), m.width)`. `View()` replaces the empty `""` crumbs placeholder with `RenderCrumbs(...)` output. Pass `m.infoPanel` to `ui.RenderChrome` call site.
- Grep-gate maintenance: `TestChromeASCIIOnly`, `TestChromeNormalBorderOnly`, `TestViewNoNewStyle` / `TestSubmodelViewsNoNewStyle` extend their file scope to include the new `internal/ui/infopanel.go` and `internal/ui/crumbs.go`. Allowlist for `TestChromeASCIIOnly` extends to include U+2026 (`…`) for middle-truncation per D-203.
- Resize goldens at 80×24 / 120×40 / 200×60 refreshed (40×12 stays narrow-tier — chrome stub + body, no info panel + no crumbs aren't visible at that size; verify and document).
- Security review per Pitfall 11: every new info-panel field passes the 5-question gate (truncated to ≤10 chars where derived from key material; repo-relative paths only; no copy bindings; no env-var paths; not logged).

**Out of scope (deferred per ROADMAP / explicit decisions):**
- Logo severity coupling (`UI-03`, Phase 10).
- 16-color palette fallback for chip bg (`UI-13`, Phase 10) — Phase 8 ships TrueColor + ANSI256 + ANSI; ASCII profile chip degrades to bracket-only via lipgloss downsample.
- Comprehensive narrow-terminal aesthetics matrix (`UI-16`, Phase 10) — Phase 8 only verifies non-corruption at 80×24 + 120×40 + 200×60; 40×12 stays at the Phase 7.1 narrow-tier stub (info panel + crumbs both intentionally absent at that size).
- `[` / `]` view history navigation (out of scope for v1.1; deferred to v1.2 per ROADMAP).
- Skin loader / k9s-compatible YAML schema (`THM-01..THM-03`, deferred to v2).
- D-18 chrome caching (Phase 11 SC2).
- Status-bar `_ = m` cleanup, IN-02..IN-07 review opportunism — out of scope per Phase 7.1's deferral rule.
- Per-`(state, recipientAction, IsSearchActive)` golden-file matrix — Phase 9 owns this.

</domain>

<decisions>
## Implementation Decisions

### Info-Panel Schema (UI-04, UI-05)

- **D-201: Terse 3-char labels.** Five fixed rows in fixed order — `cfg:` (`.sops.yaml` repo-relative path), `age:` (user's age fingerprint, ≤10 cells + `…`), `rcp:` (current file's recipient count, integer), `git:` (branch name + `*` dirty marker or `(clean)`), `fil:` (project file count, integer). Labels render via `InfoPanelLabelStyle` (`Foreground(ColorMuted).Width(5)` so `cfg:` / `age:` / `rcp:` / `git:` / `fil:` align cleanly). Convention deliberately matches k9s skin frame.menu.keyColor terseness; leaves max cells of the 38-col envelope for values (38 − 5 label − 1 separator-space = 32 cells per value row before truncation).
- **D-202: Recipient count = current file.** `m.currentParsed.Metadata.AgeRecipients` length. Value updates as the user navigates between files via `Update()` paths that already call `parser.ParseFile(currentFile.AbsPath)`. In `stateFileList` (no current file) the row shows `-` per D-204. Aggregate / project-wide recipient count rejected because it requires walking all rules at startup and is less useful for the bulk-rekey ergonomics surface (Phase 5 already gives per-file counts in the recipient list view).
- **D-203: Middle-truncation with `…`.** When a value's rendered cell width exceeds 32, truncate via middle-ellipsis preserving start (repo-root or `age1`-prefix) and end (filename or last 4 chars of fingerprint): e.g., `secrets/.../prod.yaml`, `age1abc…xyz`. Implementation uses a small `middleTruncate(s string, max int) string` helper in `internal/ui/infopanel.go` that splits the string in half, drops middle bytes, and inserts U+2026. `TestChromeASCIIOnly` allowlist extends to include U+2026 (precedent: Phase 7 already extended for box-drawing).
- **D-204: ASCII `-` empty marker.** A single ASCII hyphen-minus rendered with `InfoPanelValueStyle` (no special muting beyond the value-column foreground) marks "value not yet computed" or "not applicable for current state". Concretely: `cfg: -` if `m.sopsYamlPath == ""`, `age: -` if keys.txt missing or unparseable, `rcp: -` if no current file or parse failed, `git: -` if not a git repo, `fil: -` only during the brief window before the first `FilesDiscoveredMsg` arrives. ASCII-only; no grep-gate change.

### Crumb Chip Pill Design (UI-07)

- **D-205: `<segment>` k9s-exact wrapper.** Each chip renders as `<segment>` with no leading/trailing whitespace inside the brackets. Verbatim match to k9s `internal/ui/crumbs.go:62-74` per the project memory's "k9s visual parity is a hard quality attribute" requirement. The bg fill (D-206) provides the pill shape; the angle brackets carry the chip semantic without any extra padding cells.
- **D-206: Bg+fg accent + bold for active.** Active (last) chip: `CrumbChipActiveStyle = lipgloss.NewStyle().Background(ColorAccent).Foreground(ColorBg).Bold(true)`. Inactive chips: `CrumbChipStyle = lipgloss.NewStyle().Background(ColorSurface).Foreground(ColorFg)`. Two-channel encoding (bg color swap + bold weight) survives 16-color downsampling per Pitfall 5 / Pitfall 9 redundant-encoding rule. Inverted bg+fg + bold is the colorblind-safe replacement for k9s's bg-only swap; the bold weight is the redundant channel.
- **D-207: Lowercase + strip-spaces.** Segment normalisation: `strings.ToLower(s)` then `strings.ReplaceAll(s, " ", "")` per k9s `crumbs.go:70-71`. Centralised in `RenderCrumbs` so existing `m.status.SetBreadcrumb("files", m.currentFileBreadcrumb(), "metadata")` call-sites stay untouched. Note: `currentFileBreadcrumb` already includes the git status badge suffix `[M]`/`[A]`/`[?]` (e.g., `prod.yaml [M]`); the strip-spaces step collapses this to `<prod.yaml[m]>`. Acceptable — the badge survives in the chip; the title `Detail: prod.yaml` continues to render the un-normalised filename. Future cleanup (drop badge from breadcrumb since git state is now in `git:` row) deferred.
- **D-208: Single space between, 1-cell row pad.** `RenderCrumbs` joins chips with `" "` and applies `PaddingLeft(1).PaddingRight(1)` to the row container. Mirrors k9s `crumbs.go:32` `SetBorderPadding(0,0,1,1)`. Total inter-chip cells: each chip = `<` + segment + `>` (2 + len(seg)) cells; between chips +1 space; row +2 padding cells. Width budget for overflow check is `m.width - 2` (row pad).

### Status-Bar Shrink (UI-08)

- **D-209: Drop item count entirely.** `SetItemCount` becomes a no-op (kept as a method for backward compatibility with existing v1.0 call-sites; can be removed in a follow-up cleanup commit if no risk). The titled-border title is the canonical count display per Phase 7 D-15: `Files (12)` / `History (47)` / `Health (3 findings)` / `Recipients (4)`. `StatusBarModel.itemCount` and `itemLabel` fields can be deleted; if removed, all existing `m.status.SetItemCount(...)` call-sites in `internal/app/model.go` must be deleted in the same commit. Recommendation: delete; planner verifies no test references the removed surface.
- **D-210: Breadcrumb data stays on `StatusBarModel`; new `Segments() []string` accessor.** Returns `strings.Split(m.breadcrumb, " > ")` (or splits a private `[]string` field if `SetBreadcrumb` is refactored to store segments instead of joined string). All 16 existing `m.status.SetBreadcrumb(...)` call-sites in `internal/app/model.go` stay untouched. `StatusBarModel.View()` stops calling `renderBreadcrumb` (the function can be deleted; was only used internally). `AppModel.View()` calls `ui.RenderCrumbs(m.status.Segments(), m.width)` and inserts the rendered string into the sections slice in place of the `""` Phase 7 placeholder.
- **D-211: Right-align env + clipboard, drop pipes.** `StatusBarModel.View(width)`: when no flash is active, render only the right cluster (`renderEnvIndicators(m.env)` plus optional `[clip]` prefix) with `lipgloss.NewStyle().Width(width).Align(lipgloss.Right).Background(ColorSurface).Foreground(ColorFg).Render(rightCluster)`. The left/center sections + pipe separators (`sep := …Render("|")` lines) are deleted. Surface bg still spans full width (visual continuity with v1.0 status bar).
- **D-212: Flash unchanged.** `m.flash != ""` path keeps the v1.0 behaviour: `StatusBarStyle.Width(width).Align(Center).Render(m.flash)`. Crumb row above remains visible; flash continues to occupy the 1-row status bar. No changes to flash routing or generation-counter handling.

### Data Refresh + Age Key (UI-05, Pitfall 11, Pitfall 15)

- **D-213: Cached `infoPanel ui.InfoPanelData` on `AppModel`.** Refreshed on three event paths:
  - `FilesDiscoveredMsg` handler: `m.infoPanel.SopsYamlRelPath` + `m.infoPanel.FileCount` recomputed from `len(m.files)`.
  - `GitStatusMsg` handler: `m.infoPanel.GitBranch`, `m.infoPanel.GitDetached`, `m.infoPanel.GitDirty` recomputed via the new `git.GetBranch` (D-215) plus aggregating `m.files[*].GitStatus != ""` for dirty.
  - On every successful recipient add/remove + every successful edit re-encrypt: `m.infoPanel.RecipientCount` recomputed from `len(m.currentParsed.Metadata.AgeRecipients)` after the post-write re-parse already in v1.0.
  - At startup (in `NewAppModel` or first `Init` cmd): `m.infoPanel.AgeFingerprint` populated from `~/.config/sops/age/keys.txt` parse (D-214); `m.infoPanel.SopsYamlRelPath` derived from `sopsYamlPath` argument relative to `cwd`.
  Per Pitfall 15: `View()` must NOT call `os.Stat`, `parser.ParseFile`, `git.*`, or any other I/O — it reads the cached struct only. The cache is the single source of truth for chrome's info-panel render.
- **D-214: Age fingerprint via `filippo.io/age`.** New `internal/ui/agekey.go` (or inline in `internal/app/model.go`) reads `~/.config/sops/age/keys.txt` (or `$SOPS_AGE_KEY_FILE` if set) using `age.ParseIdentities`; takes the first identity's `Recipient().String()` (e.g., `age1abc...xyz`); stores the full string. Render-time (in `RenderInfoPanel`) middle-truncates to ≤10 cells per Pitfall 11 ("≤10 chars with visible ellipsis"). Filippo.io/age v1.3.1 is already in `go.mod` and used by `internal/ui/recipientform.go` — zero new deps. On parse failure (file missing, malformed, no identities): fingerprint = `""`; render shows `-` per D-204. Multiple identities: first only — file count + recipient count are the multi-key surfaces; the user's fingerprint shown is the project's default identity.
- **D-215: New `git.GetBranch` helper.** Add to `internal/git/status.go`:
  ```go
  func GetBranch(repoRoot string) (branch string, detached bool, err error)
  ```
  Implementation: `repo.Head()` → `ref.Name().IsBranch()` → `ref.Name().Short()` (returns e.g. `main`); if `IsBranch()` is false the HEAD is detached (commit hash returned), set `detached=true` and `branch = ref.Hash().String()[:7]`. Non-git directory returns `("", false, gogit.ErrRepositoryNotExists)` consistent with `GetFileStatuses`. Called once per `GitStatusMsg` cycle (existing async path), not per-frame. Render format: `git: main *` (dirty) / `git: main` (clean) / `git: HEAD@abc1234` (detached, marker after `@`). Test contract mirrors `TestGetFileStatuses_*` pattern — non-repo, normal branch, detached HEAD subtests.
- **D-216: Crumbs render at every chrome tier; middle-segment ellipsis on overflow.** `crumbsHeight(m)` returns `lipgloss.Height(RenderCrumbs(m.status.Segments(), m.width))` — typically 1 row (chips don't wrap). Independence from chrome tier means: at narrow-tier (chrome = 1-row "press ? for help" stub) the crumbs row is still visible above the body — body region remains reachable via Phase 7.1's narrow-tier height math (chrome 1 + crumbs 1 + status 1 = 3 rows; body = height - 3). Width-overflow handling in `RenderCrumbs`: compute joined chip widths cumulatively; if total exceeds `width - 2` (row pad), drop middle segments and replace with a single `<…>` chip preserving the first and last segments. Implementation budget: a 30-line `truncateSegmentsToWidth` helper covering the iterative drop + ellipsis insertion.

### Plan Split (3-plan ROADMAP budget)

- **D-217: Three plans, primitive-first matching Phase 7's split:**
  - **Plan 1 — InfoPanel + Crumbs primitives + age key parser:** `internal/ui/infopanel.go` (`InfoPanelData` struct + `RenderInfoPanel` + `middleTruncate`), `internal/ui/crumbs.go` (`RenderCrumbs` + `truncateSegmentsToWidth`), age-key parser (small helper or inline), 7 new style vars in `internal/ui/styles.go`. Unit tests for each primitive: `TestRenderInfoPanel_AllRowsAligned`, `TestRenderInfoPanel_EmptyMarkers`, `TestRenderInfoPanel_TruncatesAge`, `TestRenderInfoPanel_TruncatesPath`, `TestRenderCrumbs_KnsExactPills`, `TestRenderCrumbs_ActiveBoldBg`, `TestRenderCrumbs_LowercaseStripSpaces`, `TestRenderCrumbs_MiddleEllipsis`, `TestParseAgeKey_FirstIdentity`, `TestParseAgeKey_MissingFile`. Zero AppModel changes in this plan.
  - **Plan 2 — `git.GetBranch` helper + status-bar shrink:** Extend `internal/git/status.go` with `GetBranch` + 3-subtest unit test. Strip `renderBreadcrumb` + center section + pipe separators from `internal/ui/statusbar.go`; add `Segments()` accessor; update `StatusBarStyle.Render` to right-align env+clipboard. Update statusbar_test.go for the new render shape (1-2 tests deleted, 1-2 new tests added). Zero AppModel changes.
  - **Plan 3 — Integration: AppModel cache, chrome wiring, crumbsHeight flip, goldens:** Add `infoPanel` field to `AppModel`; populate at startup (age key + .sops.yaml); refresh on `FilesDiscoveredMsg`/`GitStatusMsg`/recipient ops/edit success. Modify `ui.RenderChrome` signature to accept `info InfoPanelData` and render it in the full-tier slot. Flip `crumbsHeight` from `return 0` to real value; update `View()` to feed `RenderCrumbs(m.status.Segments(), m.width)` into the sections slice (replace the Phase 7 `""` placeholder). Extend grep-gate file scope: `TestChromeASCIIOnly` + `TestChromeNormalBorderOnly` + `TestViewNoNewStyle` / `TestSubmodelViewsNoNewStyle` add `internal/ui/{infopanel,crumbs}.go`. Allowlist extension for U+2026. Refresh resize goldens at 80×24 / 120×40 / 200×60. One atomic PR; 2-3 commits acceptable.
- **D-218: Plan 3 is deliberately the largest** — the integration is tightly coupled (`infoPanel` field + 3 refresh paths + `RenderChrome` signature change + `crumbsHeight` flip + grep-gate scope + goldens). Splitting by refresh path would re-open `View()` and `Update()` multiple times.

### Validation Strategy

- **D-219: No new validation architecture.** Phase 7 patterns apply unchanged: `go build ./...`, `go test ./... -count=1`, the four grep-gate tests with extended file scope, the resize-golden tests. Add three Phase-8-specific tests in `internal/app/chrome_test.go`: `TestRenderChrome_FullTierWithInfoPanel` (asserts info panel content present in full-tier output at width=200), `TestCrumbsHeight_NonZero` (asserts `crumbsHeight(m) > 0` after first WindowSizeMsg + breadcrumb set), `TestInfoPanelCacheRefresh_OnFilesDiscovered` (asserts `m.infoPanel.FileCount` reflects new file count after handler runs).

### Security Review (Pitfall 11)

- **D-220: 5-question security review per new info-panel field, signed off in 08-03-SUMMARY.md.** The five questions per field:
  1. Does this field derive from private key material? (age fingerprint: yes → truncate ≤10 chars + `…` + show only Recipient string, never the path-to-keyfile)
  2. Could this field expose absolute filesystem paths? (cfg row: must be `filepath.Rel(cwd, sopsYamlPath)`; never `$HOME` or absolute)
  3. Does any keybinding copy this field to clipboard? (no — chrome is display-only; no copy bindings target chrome content)
  4. Does this field appear in stderr logs? (no — `internal/ui/errorbox.go` is the only stderr surface; chrome content is not logged)
  5. Could a screenshot of this field, posted publicly, narrow an attacker's search space? (fingerprint truncation + relative path are the mitigations; document residual risk)
  Plan 3 author writes the sign-off table into `08-03-SUMMARY.md`.

### Folded Todos

- **T-201: STATE.md "Phase 8 planning" todo** — folded entirely. Per STATE.md: "Phase 8 planning: flip crumbsHeight from 0 to real chip-row height; replace `""` placeholder in View()'s sections slice with rendered crumbs row; inflate the 38x6 InfoPanelPlaceholderStyle slot with the 5-row info panel content per UI-04". Closed by D-213 (info-panel inflation), D-216 (crumbsHeight flip + RenderCrumbs feed), and Plan 3 integration (sections slice rewrite).
- **T-202: STATE.md "Info-panel PII leakage" concern** — folded into D-220 (5-question security review per field).

### Claude's Discretion

- Exact byte-layout of `RenderInfoPanel` rows — whether each row is a `JoinHorizontal(label, valueStyle.Width(32).Render(value))` or a plain `fmt.Sprintf("%-5s%s", label, value)` with manual padding. Plan 1 author picks based on `lipgloss.Width` interactions with empty-marker rendering.
- `middleTruncate` algorithm specifics — split point (exact half, or favouring the right side for filename preservation); ellipsis byte-position handling for multi-byte runes (none expected in ASCII paths/fingerprints; defensive code optional).
- `truncateSegmentsToWidth` algorithm specifics — drop strategy (binary search to find max segments that fit, or linear-scan-and-drop). Plan 1 author picks; performance is not critical (≤8 segments in practice).
- Whether `currentFileBreadcrumb()`'s git badge suffix (`[M]`/`[A]`/`[?]`) survives the `<filename[m]>` chip rendering or gets stripped in `Segments()` — Plan 1 author picks. Recommendation: keep (D-207 trade-off accepted) since the redundancy with `git: main *` is mild and the badge anchors per-file state.
- Whether `SetItemCount` and `itemCount`/`itemLabel` fields are deleted in Plan 2 vs left as a no-op — Plan 2 author confirms no test references and picks. Recommendation: delete for cleanliness.
- Whether the age key parser lives in a new `internal/ui/agekey.go` or inline in `model.go` — Plan 1 author picks. Recommendation: separate file for testability + reuse symmetry with `recipientform.go`'s `filippo.io/age` import.
- Exact dirty-marker glyph for `git:` row — D-215 sketches `*` (clean) / `(clean)` (clean) / `HEAD@hash` (detached). Plan 1 author picks the final glyph. Recommendation: trailing `*` for dirty, no marker for clean (terse, matches k9s minimal-noise philosophy).
- Whether `crumbsHeight` returns `lipgloss.Height(rendered)` (consistent with `chromeHeight`) or hard-codes `1` (the chip row never wraps to more than 1 row by D-216 design). Plan 3 author picks; recommendation: `lipgloss.Height` for consistency.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project decision docs

- `.planning/ROADMAP.md` §"Phase 8: Header Info Panel + Crumb Chips" — Goal, 5 success criteria, 3-plan budget, UI-04/UI-05/UI-07/UI-08 requirements
- `.planning/ROADMAP.md` §"Milestone v1.1 — Explicitly Out of Scope for v1.1" — reject list for plan review (skin loader, polling goroutines, `:` command bar, mouse interactions, etc.)
- `.planning/REQUIREMENTS.md` §"Header Region" — UI-04 (5-row info panel: `.sops.yaml` path / age fingerprint / recipient count / git branch+marker / file count), UI-05 (PII rules: ≤10-char fingerprint, repo-relative paths, no copy bindings on chrome)
- `.planning/REQUIREMENTS.md` §"Content Framing" — UI-07 (chip pills, last segment accent), UI-08 (status bar shrink to right-aligned env + clipboard, breadcrumb above titled body)
- `.planning/REQUIREMENTS.md` §"Theming & Accessibility" — UI-15 (ASCII-only chrome, NormalBorder()-only, grep-gated; the new infopanel.go and crumbs.go must comply)
- `.planning/PROJECT.md` — k9s visual parity is a hard product goal; v1.0 functional non-regression; no framework swap

### Phase 7 + 7.1 decisions that stay authoritative

- `.planning/phases/07-chrome-skeleton/07-CONTEXT.md` D-16 (info-panel slot reserved at 38×6) + D-17 (View() composition + conditional crumbs slot) + D-22 (no `lipgloss.NewStyle()` in View() reachables) — Phase 8 inflates the slot and flips the conditional; does NOT amend these decisions
- `.planning/phases/07-chrome-skeleton/07-UI-SPEC.md` §"Phase 7 fixed dimensions" — Logo 25 cols, Menu dynamic, Info-panel placeholder 38 cols, status bar 1 row. Phase 8 honours the same dimensions
- `.planning/phases/07-chrome-skeleton/07-UI-SPEC.md` §"Phase 7 Accent reserved-for list" — `ColorAccent` reservations carry forward; Phase 8 adds chip-active bg as the next reserved use
- `.planning/phases/07.1-chrome-gap-closure/07.1-CONTEXT.md` D-116 (3-tier chrome width fallback: narrow `<41` → "press ? for help" stub; mid `41 ≤ width < 99` → menu+logo; full `≥99` → info-panel + menu + logo) — Phase 8's info-panel render only applies in the full tier
- `.planning/phases/07.1-chrome-gap-closure/07.1-UI-SPEC.md` §"Section B narrow-terminal fallbacks" — body-reachable contract at 40×12 + 80×24; Phase 8 must preserve this

### Research — v1.1 milestone (MUST READ)

- `.planning/research/SUMMARY.md` §"Phase 8: Info panel + breadcrumb chips + crumb row" — files to add, pitfall callouts, gate criteria
- `.planning/research/SUMMARY.md` §"Pitfall 11: Info-Panel Field Leaks PII via Screenshots" — 5-question security review per field, fingerprint truncation, no copy bindings on chrome
- `.planning/research/SUMMARY.md` §"Pitfall 14: Breadcrumb Chip Wraps When Path Is Long" — middle-segment ellipsis mitigation
- `.planning/research/SUMMARY.md` §"Pitfall 15: Stat'ing the Filesystem on Every View Call" — cache info-panel data; refresh on event messages, never per-frame
- `.planning/research/ARCHITECTURE.md` §"File inventory" — signatures for `infopanel.go` (`type InfoPanelData struct {...}`, `RenderInfoPanel`), `crumbs.go` (`RenderCrumbs`)
- `.planning/research/ARCHITECTURE.md` §"Phase Sequencing" — Phase 8 depends on Phase 7's chrome skeleton + Phase 7.1's narrow-tier fallback
- `.planning/research/PITFALLS.md` §"Pitfall 5: Color-Profile Downsampling on 16-Color Terminals" — paired chip bg/fg degrades; redundant encoding (bold weight) is the Phase 8 mitigation
- `.planning/research/PITFALLS.md` §"Pitfall 9: Color-Only Status Indicators" — bg+fg + bold for active chip; never color-only
- `.planning/research/PITFALLS.md` §"Pitfall 11: Header Info-Panel PII Leakage" — fingerprint truncation, relative paths, no copy bindings, security review gate
- `.planning/research/PITFALLS.md` §"Pitfall 14: Breadcrumb Chip Wraps" — middle-segment ellipsis on overflow
- `.planning/research/PITFALLS.md` §"Pitfall 15: Per-Frame Stat'ing" — cache + event refresh
- `.planning/research/PITFALLS.md` §"Looks Done But Isn't" checklist — Phase 8 items (info panel reflects live env+git+file-count; crumb chips render with active highlight)
- `.planning/research/STACK.md` §"filippo.io/age v1.3.1" — public-key parsing for the `age:` row
- `.planning/research/STACK.md` §"go-git v5.17.x" — `repo.Head()` for branch resolution

### Upstream Phase 6 / 7 / 7.1 deliverables (reused here)

- `internal/app/model.go:1296` — `AppModel.View()` (Phase 7 D-17 composition; Phase 8 modifies the sections slice and the `RenderChrome` call)
- `internal/app/model.go:1437` — `bodyDims(m)` helper (Phase 6 unchanged; Phase 8's `crumbsHeight()` flip flows through it automatically)
- `internal/app/model.go:1528` — `chromeHeight(m)` (Phase 7 unchanged in Phase 8 path; the `RenderChrome` signature change is internal — `chromeHeight` continues to call `RenderChrome` with the same args + the new info-panel arg from `m.infoPanel`)
- `internal/app/model.go:1539` — `crumbsHeight(m)` stub (Phase 8 FLIPS — the only signature change here is body, not API)
- `internal/ui/chrome.go:106` — `RenderChrome` signature change: adds `info InfoPanelData` parameter. Mid-tier and narrow-tier paths ignore it (existing behaviour preserved)
- `internal/ui/statusbar.go` — `StatusBarModel`: drops left+center sections; adds `Segments()` accessor
- `internal/ui/styles.go:257-263` — `InfoPanelPlaceholderStyle.Width(38).Height(6)` (Phase 7 reservation; Phase 8 keeps the dimensions; render content fits in (38, 6))
- `internal/git/status.go` — `GetFileStatuses` + `GetFileHistory` + `GetLastCommitTime` (Phase 4 + 5 deliverables; Phase 8 adds `GetBranch`)
- `internal/parser/yaml.go:31` — `ParsedFile.AgeRecipients []string` (Phase 2 deliverable; Phase 8 reads `len(...)` for the `rcp:` row)
- `internal/ui/recipientform.go:22` — `filippo.io/age` import precedent (Phase 5 used it for validation; Phase 8 uses it for parsing the user's identity file)
- `internal/app/chrome_test.go` — Phase 7 + 7.1 grep-gates (`TestChromeASCIIOnly`, `TestChromeNormalBorderOnly`, `TestViewNoNewStyle`, `TestBenchmarkAppView_UnderBudget` skipped); Phase 8 extends file scope

### k9s visual parity references (project memory: hard quality attribute)

- `~/git/k9s/internal/ui/crumbs.go:62-74` — chip pill format `<segment>` + bg active swap + lowercase + strip-spaces normalisation. **D-205, D-206, D-207, D-208 source.**
- `~/git/k9s/internal/ui/crumbs.go:32` — `SetBorderPadding(0,0,1,1)` row pad. **D-208 source.**
- `~/git/k9s/internal/view/cluster_info.go:67-89` — 2-column label:value table layout for cluster info (analogue to Phase 8 info panel). **D-201 row layout source.**
- `~/git/k9s/internal/view/cluster_info.go:113-144` — `ClusterInfoChanged` event-driven refresh. **D-213 caching pattern source.**

### Code files in scope (Phase 8 modifies / adds)

- `internal/ui/infopanel.go` — NEW. `InfoPanelData` struct + `RenderInfoPanel` + `middleTruncate`. Plan 1
- `internal/ui/crumbs.go` — NEW. `RenderCrumbs` + `truncateSegmentsToWidth` + segment normalisation. Plan 1
- `internal/git/status.go` — extend with `GetBranch`. Plan 2
- `internal/ui/statusbar.go` — shrink + `Segments()` accessor. Plan 2
- `internal/ui/styles.go` — 7 new style vars. Plan 1
- `internal/ui/chrome.go` — `RenderChrome` signature change (adds `InfoPanelData` parameter). Plan 3
- `internal/app/model.go` — `infoPanel` field + 4 refresh paths + `crumbsHeight` flip + `View()` sections slice update + `RenderChrome` call-site update. Plan 3
- `internal/app/chrome_test.go` — extend grep-gate file scope; allowlist `…`; add Phase-8 integration tests. Plan 3
- `internal/ui/submodel_view_no_newstyle_test.go` — add `infopanel.go` + `crumbs.go` to allowlist. Plan 3
- `internal/app/testdata/resize_*.golden` — refresh 80×24, 120×40, 200×60. Plan 3
- `.planning/phases/08-header-info-panel/08-UI-SPEC.md` — to be authored by `/gsd-ui-phase 8` (recommended before Plan 1)

### Technology / external references

- `CLAUDE.md` §"Encryption / Key Parsing" — `filippo.io/age` v1.3.1 usage rule: parse keys, validate, do NOT encrypt (encryption stays with the `sops` subprocess). Phase 8's `age:` row only reads + truncates the public Recipient string
- `CLAUDE.md` §"Git Integration" — `go-git/go-git/v5` v5.17.x used for read-only operations; `repo.Head()` is the standard branch-resolution call
- `charm.land/lipgloss/v2` package docs — https://pkg.go.dev/charm.land/lipgloss/v2 — `Background`, `Foreground`, `Bold`, `Width`, `Height`, `JoinHorizontal`, `JoinVertical`
- `github.com/charmbracelet/x/ansi` — `ansi.Truncate(s, width, "…")` — already direct dep; Phase 8's `middleTruncate` may delegate to this for the right-half truncation step

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `internal/ui/recipientform.go:22` — already imports `filippo.io/age`; Phase 8 reuses the import path. Validation pattern (`age.ParseRecipients`) is the analogue for Phase 8's `age.ParseIdentities` (one parses recipients/public, the other parses identities/private — both return slices we take `[0]` of).
- `internal/git/status.go` — established pattern for go-git operations: `gogit.PlainOpenWithOptions(repoRoot, &gogit.PlainOpenOptions{DetectDotGit: true})` is the project's canonical entry point; non-git case returns empty/nil per D-12. Phase 8's `GetBranch` follows the exact same shape.
- `internal/ui/styles.go` — the established style-as-package-var location. Phase 8 adds 7 vars in the existing additive pattern (no struct, no skin loader; flat `var (...)` block).
- `internal/ui/chrome.go:106` — `RenderChrome` is the single chrome composition site. Phase 8's signature change (add `info InfoPanelData`) is local; only `chromeHeight` and `View()` call-sites update.
- `internal/app/model.go:1414` — `currentFileBreadcrumb()` helper appends `[M]/[A]/[?]` git status badge to filename. Phase 8's chip render absorbs this via lowercase+strip-spaces (D-207); badge survives in chip text. No call-site changes needed.
- `internal/app/model.go:283-284` + 16 call-sites — `m.status.SetBreadcrumb(...)` is the existing breadcrumb-data write path; Phase 8 keeps this untouched (D-210). The only addition is `m.status.Segments()` for the read path.
- `internal/app/model.go:1351-1356` — `View()` sections slice already supports a conditional crumbs slot. Phase 8 just replaces the `""` placeholder with `RenderCrumbs(...)` and the conditional with an unconditional append.
- `internal/parser/yaml.go:31, 125` — `ParsedFile.AgeRecipients []string` is the existing recipient-list source; Phase 8 reads `len(m.currentParsed.Metadata.AgeRecipients)` for the `rcp:` row. Already populated in v1.0 paths.
- `internal/ui/styles.go` named colors — `ColorAccent` (chip active bg), `ColorBg` (chip active fg), `ColorSurface` (inactive chip bg), `ColorFg` (inactive chip fg), `ColorMuted` (label text + ellipsis chip + dirty marker). All explicit hex per AdaptiveColor ban.

### Established Patterns

- **Pure functions for renderers** — `internal/ui/*.go` components expose `View() string` or `RenderX(...) string` as pure functions of input data. `RenderInfoPanel(InfoPanelData) string` and `RenderCrumbs([]string, int) string` follow this shape.
- **Styles as package vars** — Phase 7 `TestViewNoNewStyle` enforces this; Phase 8's new styles join the existing `var (...)` block in `internal/ui/styles.go`.
- **Explicit hex colors** — no `lipgloss.AdaptiveColor` (issue #1036); Phase 8 additions honour this.
- **Helper tests co-located** — `infopanel_test.go` + `crumbs_test.go` live beside their implementations (mirrors `chrome_test.go`, `logo_test.go`, `menu_test.go`).
- **Async data refresh via tea.Cmd messages** — `FilesDiscoveredMsg` and `GitStatusMsg` are the existing event-driven refresh seams; Phase 8 hooks `infoPanel` cache into both handlers (D-213).
- **Empty/zero-state handling** — Phase 1+2 established `(none)`/`-`/empty-string conventions for missing data; Phase 8 picks `-` (D-204) consistent with v1.0 styling discipline.

### Integration Points

- `internal/app/model.go` — `AppModel` struct (line 224) gains one new field: `infoPanel ui.InfoPanelData`. `NewAppModel` (line 272) populates `.AgeFingerprint` + `.SopsYamlRelPath` at construction. `Update()` handlers for `FilesDiscoveredMsg` (line 328+), `GitStatusMsg` (line 586+) refresh respective fields. Recipient add/remove + edit success paths (multiple) refresh `.RecipientCount`.
- `internal/app/model.go:1345` — `RenderChrome` call-site updates to `ui.RenderChrome(hints, ui.LogoInfo, m.infoPanel, m.width)`. `chromeHeight()` (line 1532) updates the same way.
- `internal/app/model.go:1539` — `crumbsHeight(m)` flips from `return 0` to `return lipgloss.Height(ui.RenderCrumbs(m.status.Segments(), m.width))` (or hard-coded `1` per D-216 discretion note).
- `internal/app/model.go:1352-1354` — `View()` sections-slice conditional becomes unconditional: drop the `if crumbsHeight(m) > 0` guard (or keep as defensive) and replace `""` with `ui.RenderCrumbs(m.status.Segments(), m.width)`.
- `internal/ui/chrome.go` — `RenderChrome` signature gains `info InfoPanelData`; full-tier path replaces `InfoPanelPlaceholderStyle.Render("")` with `RenderInfoPanel(info)` (already constrained to 38×6 via the style's Width/Height; if `RenderInfoPanel` produces an unsized string, wrap it via `InfoPanelPlaceholderStyle.Render(rendered)`).
- `internal/ui/statusbar.go` — `View()` body simplifies; left + center + pipe rendering deleted; right cluster + flash branch retained. New `Segments() []string` accessor.
- `internal/git/status.go` — additive `GetBranch` function; existing functions unchanged.
- `internal/app/chrome_test.go` — file-scope arrays for `TestChromeASCIIOnly` and `TestChromeNormalBorderOnly` extend with `infopanel.go` + `crumbs.go`. ASCII allowlist gains U+2026.

</code_context>

<specifics>
## Specific Ideas

- **k9s parity is the source of truth for chip pills.** D-205, D-206, D-207, D-208 are calibrated against `~/git/k9s/internal/ui/crumbs.go:32, 62-74` — wrapper format `<segment>`, lowercase + strip-spaces normalisation, bg+fg active swap, single-space separator + 1-cell row pad. The only deviation from k9s is the addition of bold weight on the active chip (Pitfall 9 redundant-encoding requirement) — k9s relies on bg-only swap which fails on 16-color terminals. Plan reviewers must reject any drift back to bg-only.
- **Info-panel row label terseness.** D-201 picks 3-char-aligned `cfg/age/rcp/git/fil` over verbose alternatives. The 38-col envelope is tight: 5 cells label + 1 cell separator + 32 cells value column = 38 total. Verbose labels (`Config:` / `Recipients:`) would force the value column to ≤25 cells, which middle-truncates `secrets/.../prod.yaml` filenames to ~10 visible cells. The terse convention preserves the value cells where signal lives.
- **Cached info-panel struct, not per-frame computation.** D-213 puts a `infoPanel ui.InfoPanelData` field on `AppModel`. Pitfall 15 in PITFALLS.md is explicit: stat'ing the filesystem in `View()` (called per Update / per Tick) is the ban. The 4-event refresh seam (startup + FilesDiscoveredMsg + GitStatusMsg + recipient/edit ops) covers every state transition that can change a panel value. `View()` reads the cached struct — zero I/O.
- **Filippo.io/age for the `age:` row, not from .sops.yaml recipients.** D-214 distinguishes "user's age fingerprint" (the identity at `~/.config/sops/age/keys.txt`) from "file's recipient" (an entry in the .sops.yaml rule). UI-04 wording is "age key fingerprint" — singular, owner-coded. The first-recipient-of-current-file alternative would mismatch the UI-04 spec and break when the user holds a key not on the file's recipient list.
- **Status bar shrinks to right-aligned env+clipboard only — no center, no pipes.** D-209 + D-211 collapse three sections to one. The titled-border title (`Files (12)`) is the canonical count display per Phase 7 D-15 — duplicating it in a status-bar center section adds maintenance with zero new signal.
- **Breadcrumb data ownership stays on `StatusBarModel`.** D-210 is a deliberate "minimum migration" choice. The 16 existing `m.status.SetBreadcrumb(...)` call-sites are correct as-is; the only gap is read access (`Segments()`). Moving ownership to `AppModel` would multiply the diff footprint without buying separation of concerns.
- **Crumb row is independent of chrome tier — middle-ellipsis is the overflow mitigation.** D-216 keeps the crumb row visible at every chrome tier. The Phase 7.1 narrow-tier (`<41` cols) drops the chrome to a 1-row stub but keeps the crumb row on top of the body region. At narrow widths the chips themselves middle-ellipsis (`<files> <…> <prod.yaml>`) so the body region stays reachable.
- **Plan 3 is deliberately the largest** — same shape as Phase 7's split. Splitting the integration by refresh path or by call-site would multiply golden-file churn and re-open `View()` repeatedly.
- **Security review is a gate, not a checkbox** — D-220 lists 5 questions per field; sign-off goes in `08-03-SUMMARY.md`. This is the Pitfall 11 mitigation made into a planner-visible artifact.

</specifics>

<deferred>
## Deferred Ideas

### Phase 9 (already scoped)
- Per-`(state, recipientAction, IsSearchActive)` golden-file matrix with hint-vs-keymap drift assertion — Phase 8 ships single-state info-panel golden refresh; Phase 9 adds the full matrix.
- `Hints() []MenuHint` formalised as a lint/test contract — already shipped in Phase 7 D-09; Phase 9 adds the discipline layer.

### Phase 10 (already scoped)
- Logo severity coupling (`UI-03`).
- 16-color fallback palette (`UI-13`); chip bg currently relies on TrueColor downsample. Phase 10 adds explicit fallback so `<segment>` chips remain legible on `TERM=xterm`.
- Redundant shape/text encoding (`UI-14`); Phase 8 ships bold weight on active chip — Phase 10 may extend with underline / `[I]/[W]/[E]` prefix conventions.
- Comprehensive narrow-terminal aesthetics (`UI-16`); Phase 8 only verifies non-corruption at 80/120/200; broader matrix is Phase 10.

### Phase 11
- `BenchmarkAppView ≤ 50 µs/op` re-target via D-18 caching fallback (`UI-21`); Phase 8's per-frame info-panel READ from the cache is bench-friendly; Phase 11 may further cache the chrome composition itself.
- Alt-screen fill/blank frame on enter/exit.
- Terminal compat sweep + signed-off "Looks Done But Isn't" checklist.

### v2 (milestone-deferred per ROADMAP)
- Skin loader (`THM-01..THM-03`); Phase 8's chip + info-panel styles are flat package vars, not skin-driven.
- `[` / `]` view history for navigable breadcrumb (cmd-history pattern from k9s `app.go:730-758`).
- Multi-cluster / multi-context analogue for sops-tui — there's no analogue today; deferred indefinitely.

### Out of scope this phase (would be scope creep)
- "Press c to copy fingerprint" keybinding on chrome — Pitfall 11 explicitly bans copy bindings on chrome content.
- Polling goroutine that periodically re-stats `~/.config/sops/age/keys.txt` — `keys.txt` rotation mid-session is rare; restart is acceptable.
- Live skin reload on file change — fsnotify deferred to v2.
- Mouse interactions on chips — keyboard-only by core value.
- Read-only global mode indicator — per-file git badges already cover the concept.

### Reviewed Todos (not folded)
- "Phase 10 research pass recommended before planning" (STATE.md) — not in Phase 8 scope; kept as Phase 10 todo.
- "Phase 10/11: revisit BenchmarkAppView budget" (STATE.md) — not in Phase 8 scope.
- "Manual UAT per Phase 06 D-15" (STATE.md) — terminal-resize verification, deferred to `/gsd-verify-work` for Phase 8 once the chrome lands.

</deferred>

---

*Phase: 08-header-info-panel*
*Context gathered: 2026-04-28*
