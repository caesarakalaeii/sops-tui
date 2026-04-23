# Project Research Summary

**Project:** sops-tui
**Domain:** Terminal UI chrome rework — persistent k9s-style visual shell on existing Bubble Tea v2 app
**Milestone:** v1.1 — k9s visual parity
**Researched:** 2026-04-23
**Confidence:** HIGH

## Executive Summary

v1.1 is a **pure UI-shell refactor** on top of a stable v1.0 functional core. The existing `AppModel` is a healthy Elm-architecture root with a one-enum state machine and well-scoped sub-models — the right move is to **add a chrome layer**, not restructure anything. Every piece of the target look (persistent keybinding menu, ASCII logo with status coupling, header info panel, titled bordered content regions, breadcrumb chips, skin support) can be built with primitives already in the existing stack: `charm.land/lipgloss/v2` v2.0.3 (including the `table` subpackage and `Canvas`/`Layer`), `charm.land/bubbles/v2` v2.1.0, and `goccy/go-yaml` for the skin loader. **No new runtime dependencies are needed**; go.mod and go.sum should be unchanged.

The critical architectural decision is to render the chrome via a new `internal/ui/chrome.go` submodel plus a `WrapTitled` helper, driven by a lightweight `Hints() []keys.MenuHint` interface on each sub-model (pulled directly from existing `key.Binding.ShortHelp()`). This is **immediate-mode composition on every `View()` call** — k9s's observer pattern (`StackPushed`/`StackPopped`, style listeners) does not port to Bubble Tea and must be rejected. The menu, logo, info panel, and crumbs are pure functions of `(state, recipientAction, searchActive, env, flash, width)`; AppModel builds value structs in `View()` and passes them down.

The single highest-risk technical issue is **layout arithmetic regression**. There are ~15 call-sites in `internal/app/model.go` that compute body size as `m.height - statusBarHeight(m)`. Every one of them must migrate to a new `m.bodyDims()` helper that also subtracts `chromeHeight` and `crumbsHeight` **before** any chrome render work lands. If the helper goes in after the chrome, each SetSize site needs a second audit pass and the bug manifests only at large terminal sizes where hidden rows below the status bar start receiving content. Beyond that, the main risks are color-profile downsampling on 16-color terminals (paired bg/fg chips collapse to one color), Unicode/emoji width drift on macOS Terminal, and skin-loader failures locking users out — all addressable by well-defined mitigations (terminal profile detection, ASCII-only chrome content, fail-open skin loading).

## Key Findings

### Recommended Stack

**No changes to go.mod.** Every v1.1 feature maps to a primitive already resolved in the dependency tree. The skin loader reuses the `goccy/go-yaml` import already present for SOPS file parsing. See `STACK.md` for the full feature-to-primitive mapping.

**Core technologies:**
- `charm.land/lipgloss/v2/table` — multi-column menu grid (hint key + description cells, `StyleFunc(row,col)` matches k9s mnemonic/desc coloring for free); static chrome render with no cursor/focus state.
- `charm.land/lipgloss/v2` borders + `JoinVertical`/`JoinHorizontal` — header composition (info-panel left, menu center, logo right), titled content region (RoundedBorder with title overlaid on top edge via string splicing — lipgloss v2 has no native border-title API).
- `charm.land/lipgloss/v2/canvas` + `Layer` — reserved for modal overlays (help popup); **not** used for inline frame titles (overkill).
- `github.com/goccy/go-yaml` v1.19.2 — parses skin YAML with anchor/alias support (k9s-compatible schema subset).
- `github.com/mattn/go-runewidth` — transitively available via lipgloss; used for title truncation and user-supplied string width calculation.

**Explicitly NOT added:** `tview`/`tcell` (framework swap ruled out), `bubbles/v2/table` (interactive table, wrong primitive for static menu), a dedicated theming library (none exists in Charm ecosystem as of April 2026; rolling our own trimmed schema is ~150 LOC).

### Expected Features

**Must have (table stakes — all required to "look like k9s"):**
- **Header region** with three columns: info-panel (top-left), menu (center, flex), logo (top-right, ~26 cols); fixed height driven by 6-line logo.
- **Info panel** — 5 label:value rows: `.sops.yaml` path (relative to repo root), age key fingerprint (≤10 chars truncated with ellipsis), recipient count, git branch + clean/dirty, file count. Muted labels + bold accent values.
- **ASCII logo with status coupling** — 6-row hand-designed art, recolored (info/warn/error) from aggregate env + flash severity; 1-line status message below the art.
- **Persistent keybinding menu** — 6-row multi-column grid hydrated from the active sub-model's `Hints()`, re-rendered every frame (no observer pattern needed in immediate-mode).
- **Titled bordered content regions** — every primary view wrapped via `WrapTitled(title, body, w, h)`; title encodes `Name (count)` and optionally filter state; single accent color, **no focus-ring variants** (single-pane app).
- **Breadcrumb chips** — replace `" > "` separator with `<segment>` pills (background + padding); last segment uses active color.
- **Palette tuned to k9s conventions** — existing Catppuccin Mocha is close; tune accent to match k9s's typical hot-pink/purple highlight.
- **`Esc` audit on all prompts** — confirm search/prompt cancel without bubbling `tea.Quit`.

**Should have (differentiators, likely in this milestone):**
- **Logo severity bound to aggregate health** — reuses existing `internal/health/` data; low cost, high UX payoff.
- **k9s-compatible skin YAML schema subset** — deliberately reuse k9s field names (`body.fgColor`, `frame.border.fgColor`, `frame.menu.keyColor`, `frame.crumbs.activeColor`, etc.) so users can drop an existing k9s dracula/gruvbox skin in; extra keys silently ignored.

**Defer (v1.2+):**
- **Skin YAML user-facing loader** — recommend shipping v1.1 with a k9s-tuned **default palette only**; the YAML loader + 2-3 builtin skins (dracula/gruvbox) is L-complexity and better as a standalone v1.2 feature.
- `:` command bar + tab-completion + number-key view switching + `[`/`]` history (v1.2).
- Live skin reload via fsnotify (v1.3).
- Splash screen with big ASCII logo (nostalgic; no user value — v1.3).
- Responsive narrow-terminal header collapse (trigger: first bug report).

**Explicitly NOT v1.1 (anti-features from `FEATURES.md` Category 9):** `clusterUpdater` polling goroutine (no external data source to poll), exponential backoff + `BailOut`, benchmarking UI, log forwarding/streaming, image vulnerability scanner, port-forward manager, alias file hot-reload, dialog severity levels, context switcher, read-only global mode, namespace selector, GitHub update-check phone-home. Several are security-relevant (phone-home is a trust violation for a secrets tool; destructive commands via `:` bar break the "diff-before-re-encrypt" safety net).

### Architecture Approach

Add a **chrome submodel** above the existing sub-models without touching the state machine. AppModel.View() becomes the central composer: build value structs (`InfoPanelData`, `LogoStatus`, `[]MenuHint`, crumb segments) from model state, pass them to the chrome renderer, wrap the active sub-model's body in `WrapTitled`, and stack everything with `lipgloss.JoinVertical`. Sub-models each gain one method — `Hints() []keys.MenuHint` — delegating to a new `keys.HintsFromBindings()` helper that converts existing `key.Binding.ShortHelp()` output. Dispatch to the active sub-model's hints is done by `switch m.state` in AppModel (plus `recipientAction` and `IsSearchActive` axes for the shared diff/search cases).

**Major components (all new files):**
1. `internal/ui/chrome.go` — `ChromeModel` owns header composition (info + menu + logo) + exports `WrapTitled` helper and `Height()` accessor for AppModel's size arithmetic.
2. `internal/ui/infopanel.go` — pure renderer for the `InfoPanelData` struct (5 label:value rows mirroring k9s `ClusterInfo`).
3. `internal/ui/logo.go` — `LogoSmall []string` (6 ASCII rows, ~26 cols wide) + `RenderLogo(status, width)` with three state-keyed styles.
4. `internal/ui/menu.go` — `RenderMenu([]MenuHint, width, maxRows=6)` built on `lipgloss/v2/table`.
5. `internal/ui/crumbs.go` — `RenderCrumbs([]string, width)` with chip pill styling; last segment active.
6. `internal/keys/hints.go` — `MenuHint{Mnemonic, Description, Visible}` struct + `Hinter` interface + `HintsFromBindings([]key.Binding) []MenuHint` converter.

**Key integration points:**
- `internal/app/model.go` `View()` (around line 1329) — full rewrite of the composition block.
- `internal/app/model.go` `WindowSizeMsg` handler (around line 313) — subtract `chromeHeight` + `crumbsHeight` when propagating.
- `internal/ui/statusbar.go:179` `renderBreadcrumb` — gut the body; breadcrumb data (stay on `StatusBarModel`) now exposed via `Segments() []string` accessor for the new crumbs row above body. Status bar shrinks to right-aligned env + clipboard only.
- `internal/ui/styles.go` — add ~10 new styles (`MenuKeyStyle`, `MenuDescStyle`, `LogoStyle{Info,Warn,Error}`, `TitledBorder`, `TitleLabelStyle`, `InfoPanelLabelStyle`, `InfoPanelValueStyle`, `CrumbChipStyle`, `CrumbChipActiveStyle`, `CrumbChipSep`). Explicit hex per the AdaptiveColor ban (issue #1036).

**Files NOT changed:** existing sub-model `View()` bodies, `internal/sops/*`, `internal/parser/*`, `internal/git/*`, `internal/health/*`, the state machine enum, all existing `key.Binding` values. Full inventory in `ARCHITECTURE.md`.

### Critical Pitfalls

From `PITFALLS.md` — the top risks unique to v1.1 (v1.0 pitfalls are archived and not re-catalogued).

1. **Chrome height arithmetic regression (Pitfall 1)** — **highest priority; landing order matters.** 15 call-sites in `model.go` (lines 316, 349, 377, 484, 502, 567, 631, 724, 761, 846, 924, 1005, 1089, 1110, 1250) compute `m.height - statusBarHeight(m)` and pass it to `SetSize` on the embedded `bubbles/v2/list`, diff, health, metadata, etc. Adding chrome without migrating all 15 sites to a new `m.bodyDims()` helper produces invisible "content painted under header" bugs visible only at large terminal sizes. **Must land first, grep-gated in CI** (`m.height - statusBar` outside the helper fails the build). Resize tests at 40×12, 80×24, 200×60 required.
2. **Menu hints diverging from actual keybindings (Pitfall 3)** — `m.state` alone is insufficient: `stateDiff`, `stateRecipientConfirm`, and `stateBulkReKeyConfirm` all render the same diff sub-model; `recipientAction` disambiguates. Hint source must be a pure function of `(state, recipientAction, IsSearchActive)`. The keymap is the single source of truth; the menu is derived. Golden-file test per tuple prevents drift.
3. **Color-profile downsampling on 16-color terminals (Pitfall 5)** — paired chip bg/fg collapses to one color on `TERM=xterm`; chip becomes unreadable solid block. Detect profile at startup via lipgloss v2 API; swap to a 16-color-safe fallback palette; for chips, degrade to `< segment >` brackets + underline for active. Multi-profile teatest matrix (Ascii, ANSI, ANSI256, TrueColor).
4. **Skin loading fails closed (Pitfall 4)** — a bad hex string in `~/.config/sops-tui/skin.yaml` must **not** refuse launch or silently produce invisible text (lipgloss `Color("")` is lenient, resolves to empty color). Validate hex eagerly at load time (`validateHex` per field), fail **open** with status-bar flash warning, fall back to default palette. No hot-reload in v1.1. (Mitigated by deferring the user-facing loader to v1.2 — see Implications below.)
5. **Unicode width drift on macOS Terminal (Pitfall 6) + font coverage (Pitfall 12)** — emoji width differs across terminals; fancy borders (rounded, double, thick) render as `□` on fonts without full U+2500–U+257F coverage. Rules: **no emoji in persistent chrome** (ASCII + color only; `[I]`/`[W]`/`[E]` labels); **`lipgloss.NormalBorder()` only** (grep-gated); user-supplied strings pass through `runewidth.Truncate` before embed; defensive post-compose `lipgloss.Width(chrome) <= m.width` assertion.

**Also serious:** color-only accessibility (redundant shape/text encoding on every state — Pitfall 9); render cost in `View()` under held-j scrolling (no `lipgloss.NewStyle()` inside View; styles as package vars — Pitfall 2); teatest golden files including raw ANSI break on lipgloss bumps (compare ANSI-stripped text + separate color-presence assertions — Pitfall 8); focus-ring visual promising Tab navigation that doesn't exist (Pitfall 7 — ban `FocusedBorder`/`UnfocusedBorder` naming); info-panel PII leakage via screenshots/tmux capture (Pitfall 11 — truncate fingerprints, relative paths only, no auto-copy in chrome, 5-question security review per new field).

## Implications for Roadmap

Suggested phase structure — **6 phases total**, grouped around the architectural dependency DAG (`ARCHITECTURE.md` Phase Sequencing). Phase 6 is a non-skippable prerequisite; Phase 7 is the chrome skeleton; each subsequent phase adds one chrome column or behaviour and can ship an incremental visible change.

### Phase 6: Layout-arithmetic groundwork + testing harness
**Rationale:** Pitfall 1 is the highest-severity regression risk and **must land before any chrome rendering** to avoid two audit passes on 15 SetSize call-sites. Pitfall 8 (golden-file stability) also calls for ANSI-stripped comparison helpers to exist before any chrome teatest is written. This is a pure-refactor phase that ships no visible change — but it unlocks everything else.
**Delivers:**
- `chromeHeight(m)`, `crumbsHeight(m)`, `bodyDims(m) (w, h int)` helpers in `internal/app/model.go`.
- All 15 `m.height - statusBarHeight(m)` call-sites migrated to `m.bodyDims()`. Grep-gate added to CI.
- ANSI-stripped golden-file helper + color-presence assertion split in the test harness.
- `keys/hints.go` with `MenuHint` struct + `HintsFromBindings` converter (no deps on anything else).
- `styles.go` additions for all new style vars (empty bodies or default values OK; populated in later phases).
**Addresses:** Pitfall 1, Pitfall 8. Architectural foundation for Phases 7–10.
**Avoids:** Retroactively fixing height arithmetic after chrome merges.

### Phase 7: Chrome skeleton — logo + menu + titled borders
**Rationale:** The three components that together make the app "look like k9s" at first glance: logo + menu in the header, titled border around body. Info panel is stubbed (placeholder string) — deferred to Phase 8 to keep this phase focused on layout primitives. Titled borders are independent per the architecture DAG and can land here safely.
**Delivers:**
- `internal/ui/logo.go` (6-row ASCII + `RenderLogo(state)` with three state-keyed styles).
- `internal/ui/menu.go` using `lipgloss/v2/table` with `StyleFunc(row,col)` for mnemonic vs description coloring.
- `internal/ui/chrome.go` (minus info-panel) with `WrapTitled(title, body, w, h)` helper including the `overlayTitle` string-splice workaround (lipgloss v2 has no native border-title API — Pitfall 9).
- AppModel.View() rewritten to compose `[chrome][titled body][statusbar]`.
- `Hints()` methods on FileList, Detail, Help, Diff sub-models.
- **Discipline rules enforced:** no emoji in chrome, `NormalBorder()` only, no `FocusedBorder`/`UnfocusedBorder`, no `lipgloss.NewStyle()` inside View, color-profile detection on startup with 16-color fallback palette.
**Uses:** `lipgloss/v2/table`, `lipgloss/v2` borders + `JoinVertical`/`JoinHorizontal`.
**Avoids:** Pitfalls 2, 5, 6, 7, 9, 12.
**Gate:** teatest golden at stateFileList (ANSI-stripped) matches expected layout; benchmark `BenchmarkAppView` ≤ 50 µs/op at 200×60.

### Phase 8: Info panel + breadcrumb chips + crumb row
**Rationale:** With the header skeleton in place, add the data-bound info panel and replace the status-bar breadcrumb with chip pills above the body. Replacing the breadcrumb now (rather than earlier) means the crumbs row can be measured into `crumbsHeight` before the chrome skeleton depends on it. Remaining sub-models get `Hints()`.
**Delivers:**
- `internal/ui/infopanel.go` with `InfoPanelData` struct (`.sops.yaml` relative path, age fingerprint ≤10 chars with ellipsis, recipient count, git branch+dirty, file count).
- `internal/ui/crumbs.go` — chip pill renderer; last segment active; middle-segment ellipsis when narrow (Pitfall 14).
- `StatusBarModel.Segments() []string` accessor; `renderBreadcrumb()` body removed from status bar; status bar shrinks to right-aligned env + clipboard.
- `Hints()` on Health, History, RecipientForm, Metadata sub-models.
- Info-panel fields cached on model; refreshed on events only (Pitfall 15 — avoid stat storms on NFS homedirs).
- **Security-review gate** on each info-panel field: truncated fingerprint, relative path, no env-var paths, no copy bindings in chrome (Pitfall 11).
**Uses:** existing validator/keys/git/health data; `goccy/go-yaml` NOT needed in this phase.
**Addresses:** v1.1 Must-Have features (info panel, breadcrumb chips).
**Avoids:** Pitfalls 11, 14, 15.
**Gate:** info panel reflects live env+git+file-count; narrow-terminal (40-col) snapshot shows ellipsized crumb; fingerprint security review signed off.

### Phase 9: Logo state + flash severity + palette tune
**Rationale:** Wire the logo color to aggregate health (env failures, flash severity). Currently flash messages are plain strings — add `FlashErr`/`FlashWarn`/`FlashInfo` triad on `StatusBarModel`; update ~10 flash call-sites in model.go with explicit severity. Then tune the default palette to k9s conventions. This phase is the "it behaves like k9s too" polish.
**Delivers:**
- `StatusBarModel.FlashSeverity()` accessor + three-severity `Flash*` methods.
- AppModel `resolveLogoState()` derives `LogoStatus` from env + flash + aggregate health severity.
- `styles.go` palette tune: accent shifts to k9s hot-pink/purple; retain Catppuccin base tones.
- **Redundant encoding for every color-coded state** (Pitfall 9): info/warn/error logos get text prefix or glyph; active chip gets inverted bg+fg (structural); `ErrorLabel`/`WarnLabel` already in styles.go — extend pattern.
- Chrome stale-during-async-op defense: re-run env check after SOPS auth/access errors (Pitfall 17).
**Uses:** existing `internal/health/` severity aggregation.
**Addresses:** Must-Have "Logo wired to health aggregate" (Category 3 differentiator in FEATURES.md — flagged for v1.1 inclusion).
**Avoids:** Pitfalls 9, 17.
**Gate:** error flash turns logo red with `[E]` prefix; warn turns yellow with `[W]`; colorblind-simulator review on DoD checklist; env re-check runs after SOPS error.

### Phase 10: k9s-tuned default palette shipped; skin YAML loader deferred to v1.2
**Rationale:** `FEATURES.md` calls out skin support as an L-complexity feature with known failure modes (Pitfall 4); shipping the YAML loader in v1.1 widens scope and adds a security/UX risk surface that isn't required for the milestone. The milestone goal is "visual parity," which is satisfied by a single well-tuned default palette. Ship v1.1 with the palette tuned in Phase 9 and a `Skin` struct shape that's forward-compatible with k9s-schema-subset YAML in v1.2.
**Delivers:**
- `internal/ui/skin.go` scaffolding: `Skin` struct shape matching k9s-compatible schema (`body.fgColor`, `body.bgColor`, `body.logoColor*`, `frame.border.fgColor`, `frame.menu.fgColor`, `frame.menu.keyColor`, `frame.crumbs.fgColor`, `frame.crumbs.bgColor`, `frame.crumbs.activeColor`, `frame.title.fgColor`, `frame.title.bgColor`). No loader, no file I/O — just the struct + an `ApplySkin(s *Skin)` function that currently only the default skin hydrates.
- Default skin values moved from hex literals in `styles.go` into a `DefaultSkin` struct; style vars reference skin fields via `ApplySkin`. Enables trivial v1.2 YAML loader addition (parse + validate + apply).
- README note: skins documented as "coming in v1.2"; default palette described.
**Uses:** `goccy/go-yaml` NOT needed yet — reserved for v1.2.
**Addresses:** v1.1 Must-Have "palette tuned to k9s conventions" (already done in Phase 9; this phase structurally prepares for v1.2 user-facing skin files).
**Avoids:** Pitfall 4 (skin fail-open) — no user-facing failure path in v1.1 since there's no YAML input yet.

### Phase 11: Terminal compatibility sweep + release polish
**Rationale:** Final phase before milestone close. Matrix-test on the terminal emulators that matter (Pitfall 10), regenerate goldens, verify benchmarks, run the "Looks Done But Isn't" checklist from `PITFALLS.md`.
**Delivers:**
- Manual verification on Alacritty, Ghostty, macOS Terminal.app, iTerm2, Windows Terminal, WSL2, VSCode integrated terminal, tmux-nested. Screenshot record per combo.
- Alt-screen cleanup: paint fill frame on enter, blank frame on exit (Pitfall 10).
- README troubleshooting section for VSCode integrated terminal known issues.
- Re-verify `BenchmarkAppView` ≤ 50 µs/op at 200×60; no `lipgloss.NewStyle()` in `View()`; header cached not recomputed per frame (Pitfall 2).
- Run the full 15-item "Looks Done But Isn't" checklist from PITFALLS.md; sign off each.

### Phase Ordering Rationale

- **Phase 6 first** is non-negotiable. Height arithmetic is the highest-risk regression; grep-gate + helper must exist before chrome rendering lands, or the same 15 SetSize sites need a second audit pass (Pitfall 1). ANSI-stripped test helper lives here too so no chrome golden is ever written against raw ANSI bytes (Pitfall 8).
- **Phases 7 → 8 → 9** follow the architectural DAG: chrome skeleton (logo + menu + borders) → data binding (info panel + crumbs + per-view `Hints()`) → behaviour polish (severity routing + palette tune). Each phase produces a visible improvement and ships a coherent slice of "looks like k9s."
- **Phase 10 is structural prep, not user-facing shipping** — lays the skin struct shape so v1.2's YAML loader is a one-phase addition. Deferring the loader dodges Pitfall 4's fail-closed failure mode entirely for v1.1.
- **Phase 11 is the release gate**, not a feature phase. Terminal compat, benchmark, checklist sign-off.
- Anti-features from `FEATURES.md` Category 9 (polling goroutines, log forwarding, image scanning, context switcher, phone-home update check, destructive `:` commands, etc.) are **never** in-scope — listed explicitly in roadmap preamble so they don't re-emerge during phase planning.

### Research Flags

**Phases likely needing deeper research during planning (`/gsd-research-phase`):**
- **Phase 7 (Chrome skeleton):** The `overlayTitle` string-splice pattern for titled borders is non-trivial and currently best documented via reading `github.com/charmbracelet/soft-serve/pkg/ui/components/header`. Worth a research pass to confirm the reference implementation and pre-vet an API shape before the phase starts.
- **Phase 9 (Logo state + flash severity):** The aggregate health severity model spans `internal/health/`, `internal/validator/`, and the flash pipeline. A short research pass before the phase to enumerate severity sources and define the classification function (e.g., is "git dirty" a warn or info?) avoids back-and-forth during implementation.

**Phases with standard patterns (skip `/gsd-research-phase`):**
- **Phase 6 (Layout arithmetic):** pure mechanical refactor; no research needed. Grep-gate CI pattern is well-established.
- **Phase 8 (Info panel + crumbs):** all primitives already spelled out in `ARCHITECTURE.md` and `STACK.md`. Data sources already exist on `AppModel`.
- **Phase 10 (Skin struct prep):** struct shape is prescribed by the k9s schema subset in `FEATURES.md` / `STACK.md`.
- **Phase 11 (Compat sweep):** checklist-driven, no new patterns.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Every feature maps to a primitive locally verified in `~/go/pkg/mod/charm.land/lipgloss/v2@v2.0.3/`. Context7 confirmed `table.New().StyleFunc(row, col)` API and that no skin/theme loader exists in the Charm catalog (rolling our own is forced, not a shortcut). No version bumps required. |
| Features | HIGH | Grounded in k9s source at `~/git/k9s/internal/{ui,view}/` with line-level citations across 35 shipped skins. Anti-features explicitly enumerated (Category 9) so they're rejected mechanically during roadmap review. |
| Architecture | HIGH | Pattern-tested against existing `AppModel` sub-model ownership model; all 6 new files have defined interfaces and integration points; WindowSizeMsg propagation paths enumerated by line number. The one non-obvious bit — `overlayTitle` lipgloss v2 workaround — is flagged for Phase 7 research. |
| Pitfalls | HIGH | 17 pitfalls enumerated with specific file:line references, warning signs, and phase ownership. v1.0 pitfalls archived and not re-catalogued. Each pitfall has a verification mechanism in the "Looks Done But Isn't" checklist. |

**Overall confidence:** HIGH

### Gaps to Address

- **`overlayTitle` reference implementation** — `PITFALLS.md` Pitfall 9 and `ARCHITECTURE.md` Pattern 3 both point at `charmbracelet/soft-serve` as the canonical pattern but neither dereferences the actual code. **Mitigation:** Phase 7 research pass should extract the pattern into our repo with a comment citing the source revision, then unit-test `TestOverlayTitle_PreservesCornersAndWidth`.
- **Aggregate health severity classification** — Phase 9 requires a decision: does "git dirty" elevate the logo to warn, or stay info? Does "stale recipient key" (health-check finding) elevate to warn? **Mitigation:** Phase 9 research pass; classify each source from `internal/health/` and `internal/validator/` into info/warn/error explicitly; document in a short decision table.
- **Flash severity call-site audit** — 10 flash call-sites in `model.go` need conversion from plain-string to severity-typed calls. Count is approximate per `ARCHITECTURE.md` Pattern 4; actual count should be confirmed via grep in Phase 9. **Mitigation:** grep `\.Flash\(` in `internal/app/model.go` during Phase 9 planning, enumerate each site, decide severity per call.
- **Terminal compat matrix target list** — Phase 11 will verify on a specific set of terminals; `PITFALLS.md` lists 8 combos as the expected surface. **Mitigation:** confirm list with user during Phase 11 kickoff — especially whether VSCode integrated terminal is a must-support or best-effort target (known xterm.js alt-screen issues).
- **Narrow-terminal fallback ordering** — at widths <60, the chrome should degrade gracefully. Stack vertically? Hide logo? Hide info panel? `ARCHITECTURE.md` Pitfall 5 suggests vertical stack as fallback; `FEATURES.md` Category 1 differentiator describes responsive column thresholds (hide logo <100, hide info <70). Current recommendation: defer actual implementation to v1.2 (per FEATURES.md Future Consideration list); Phase 11 verifies chrome *survives* narrow widths without layout corruption even if ugly. **Mitigation:** document the "ugly but not broken" contract in Phase 11 acceptance criteria.

## Sources

### Primary (HIGH confidence)
- `charm.land/lipgloss/v2` v2.0.3 — local module inspection at `~/go/pkg/mod/charm.land/lipgloss/v2@v2.0.3/{table/table.go, canvas.go, layer.go}` confirming `table` subpackage and `Canvas`/`Layer` APIs.
- `charm.land/bubbles/v2` v2.1.0 — Context7 `/charmbracelet/bubbles` (benchmark 80.92) confirmed v2 component catalog; no skin loader exists.
- `charm.land/bubbletea/v2` v2.0.4 — UPGRADE_GUIDE_V2.md verifies `tea.KeyPressMsg` interface vs v1 struct (Pitfall 13).
- `github.com/goccy/go-yaml` v1.19.2 — already in go.mod, anchor/alias support verified for k9s-schema-subset parsing.
- k9s reference implementation at `~/git/k9s/`:
  - `internal/ui/{menu.go, logo.go, splash.go, crumbs.go, table.go, table_helper.go, action.go, prompt.go}` — visual-shell components, all cited by line number across research files.
  - `internal/view/{app.go:273-331, cluster_info.go, live_view.go}` — header composition, ClusterInfo layout, full-screen toggle patterns.
  - `internal/model/menu_hint.go, types.go:56` — `MenuHint` struct + `Hinter` interface reference.
  - `skins/{dracula.yaml, gruvbox-dark.yaml, monokai.yaml}` — skin schema verified across three distinct styles.
- sops-tui existing codebase:
  - `internal/app/model.go:313-329, 1329, 1443` — WindowSizeMsg propagation, View composition entry point, `statusBarHeight` helper.
  - `internal/ui/{styles.go, statusbar.go:179, help.go, filelist.go:275-346}` — current palette (AdaptiveColor ban noted), breadcrumb render site, help overlay pattern, list SetSize.
  - `internal/keys/bindings.go:72-85, 190-205` — existing `ShortHelp()` and `FullHelp()` methods feeding hint extraction.
  - `go.mod` — pinned versions for all above packages.
- `CLAUDE.md` — project stack conventions, `lipgloss.AdaptiveColor` ban (issue #1036), Bubble Tea v2 migration rules.
- `.planning/PROJECT.md` — v1.1 milestone goals and scope boundary; validated v1.0 features that must not regress.

### Secondary (MEDIUM confidence)
- Bubble Tea v2 Cursed Renderer + Mode 2026 synchronized output documentation — https://gist.github.com/christianparpart/d8a62cc1ab659194337d73e399004036 (mentioned in STACK.md context).
- teatest `WaitFor` + `RequireEqualOutput` patterns — https://charm.land/blog/teatest/ (Pitfall 8).
- `github.com/mattn/go-runewidth` — transitive dependency of lipgloss; k9s uses it directly (k9s `menu.go:17, 188`) confirming width-safe truncation pattern.
- `github.com/charmbracelet/soft-serve/pkg/ui/components/header` — cited as reference implementation for `overlayTitle` on-border-title pattern (ARCHITECTURE Pattern 3, Pitfall 9). **Not yet locally verified** — Phase 7 gap to close.

### Tertiary (LOW confidence)
- VSCode integrated terminal alt-screen issues — known long-standing xterm.js behaviour, referenced in Pitfall 10. No single canonical fix; mitigation is workaround (fill-frame on enter, blank-frame on exit) + README troubleshooting note. Verify during Phase 11 manual sweep.
- Colorblind-safe palette guidance — Wong 2011 / Nature Methods. Recommendation (redundant shape/text encoding) is well-established in a11y literature but specific palette distances for our accent/surface/error triad are inferred, not measured.

---
*Research completed: 2026-04-23*
*Ready for roadmap: yes*
