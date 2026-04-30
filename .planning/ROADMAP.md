# Roadmap: sops-tui

## Overview

sops-tui ships in two milestones. **Milestone v1.0 (Phases 1-5)** delivered the functional core: file discovery, read/write loops, clipboard, git integration, recipient management, and secret health checks. **Milestone v1.1 (Phases 6-11)** reshapes the UI shell to match k9s visual conventions — persistent keybinding menu, ASCII logo with status coupling, header info panel, titled bordered content, colored breadcrumb chips, and a k9s-tuned palette with 16-color fallback and colorblind-safe redundant encoding. v1.1 is a pure UI-shell refactor on top of the stable v1.0 functional core; no v1.0 functionality regresses.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

**Milestone v1.0 — Functional Core**

- [x] **Phase 1: Foundation** - TUI skeleton, SOPS subprocess wrapper, config discovery, security groundwork, startup validation
- [x] **Phase 2: Read Loop** - File browser, key names without decrypt, SOPS metadata, fuzzy search, full navigation (completed 2026-04-14)
- [x] **Phase 3: Write Loop** - On-demand decrypt, reveal, edit with diff, format-aware rotation, re-encryption
- [x] **Phase 4: Clipboard & Git** - Clipboard with auto-clear and signal safety, git change badges, blame/history, cross-file search
- [x] **Phase 5: Power Features** - Recipient management, bulk re-key, secret health checks

**Milestone v1.1 — k9s Visual Parity**

- [x] **Phase 6: Layout Groundwork** - `bodyDims()` helper, migrate 15 SetSize call-sites, CI grep-gate, ANSI-stripped teatest harness
- [x] **Phase 7: Chrome Skeleton** - ASCII logo, persistent keybinding menu, titled bordered content regions
- [x] **Phase 7.1: Chrome Gap Closure (INSERTED)** - Restore SC5 governance (revert unauthorized ROADMAP/test-gate amendments; defer perf to Phase 11), strip nested sub-model RoundedBorder boxes, clamp narrow-terminal chrome overflow, extend View() AST walker into helpers, align menu test allowlist
- [ ] **Phase 8: Header Info Panel + Crumb Chips** - Top-left info panel, colored breadcrumb chips above body, shrunk status bar
- [ ] **Phase 9: Keybinding Discoverability** - `Hints() []MenuHint` interface, per-view menu hydration, `?` overlay retained
- [ ] **Phase 10: Theming + Accessibility** - Logo severity coupling, k9s-tuned palette, 16-color fallback, redundant encoding, narrow-terminal survival
- [ ] **Phase 11: Regression + Perf Gates** - v1.0 integration tests pass, `BenchmarkAppView` ≤ 50 µs/op, terminal compat sweep

## Phase Details

### Phase 1: Foundation
**Goal**: The application starts, validates its environment, and provides a navigable skeleton — without ever touching a secret value
**Depends on**: Nothing (first phase)
**Requirements**: HLT-01, HLT-02, NAV-03, NAV-05, NAV-06
**Success Criteria** (what must be TRUE):
  1. Running `sops-tui` with no `sops` binary shows a clear error with installation instructions and exits cleanly
  2. Running `sops-tui` with no age key file shows a clear error with setup instructions and exits cleanly
  3. User can navigate any view using hjkl, g/G, and ctrl-d/u without errors
  4. Pressing `?` opens a contextual help panel listing all keybindings
  5. A persistent status bar is visible on every screen showing file path and operation feedback
**Plans:** 4 plans

Plans:
- [x] 01-01-PLAN.md — Go module setup, design system colors/styles, keybinding contracts
- [x] 01-02-PLAN.md — Startup validation (sops/age/.sops.yaml) and styled stderr error box
- [x] 01-03-PLAN.md — File list and YAML tree detail view components
- [x] 01-04-PLAN.md — Help overlay, status bar, root model wiring, main.go entry point

**UI hint**: yes

### Phase 2: Read Loop
**Goal**: Users can browse all SOPS-encrypted files and inspect their contents without any decryption occurring
**Depends on**: Phase 1
**Requirements**: NAV-01, NAV-02, NAV-04, DEC-03, DEC-04
**Success Criteria** (what must be TRUE):
  1. User can open `sops-tui` in a repository and see all SOPS-encrypted files discovered via `.sops.yaml` path rules
  2. Selecting a file shows all key names with values masked by default — no decryption has occurred
  3. User can view SOPS metadata (version, lastmodified, recipients, MAC status) for any file without decrypting it
  4. Pressing `/` opens a fuzzy search that filters across file names and key names in real time
**Plans:** 3/3 plans complete

Plans:
- [x] 02-01-PLAN.md — SopsDiscoverer + YamlParser backend services, goccy/go-yaml dependency, TreeNode extension
- [x] 02-02-PLAN.md — MetadataModel overlay, SearchModel with fuzzy matching, 5 new named styles
- [x] 02-03-PLAN.md — AppModel wiring: discovery, parsing, metadata overlay, search, keybindings, state machine

**UI hint**: yes

### Phase 3: Write Loop
**Goal**: Users can decrypt, reveal, edit, and rotate secrets with a safety gate before any write is committed
**Depends on**: Phase 2
**Requirements**: DEC-01, DEC-02, EDT-01, EDT-02, EDT-03, EDT-04
**Success Criteria** (what must be TRUE):
  1. User can reveal a single secret value on demand; it appears in the view and can be hidden again
  2. User can decrypt and reveal all values in a file at once
  3. User can edit a secret value; before re-encryption a diff view is shown requiring explicit confirmation
  4. User can rotate a secret to a format-aware random value (base64, hex, UUID, bcrypt) with confirmation
  5. Any destructive write operation presents a confirmation prompt that can be cancelled without effect
**Plans:** 3 plans

Plans:
- [x] 03-01-PLAN.md — SOPS executor subprocess wrapper, on-demand decrypt/reveal, keybinding and style extensions
- [x] 03-02-PLAN.md — Diff confirmation overlay, inline single-key editing with textinput, stateDiff wiring
- [x] 03-03-PLAN.md — $EDITOR full-file editing flow, format-aware secret rotation with X key

**UI hint**: yes

### Phase 4: Clipboard & Git
**Goal**: Users can safely copy secrets to clipboard with guaranteed cleanup, and see git state alongside their secrets
**Depends on**: Phase 3
**Requirements**: CLB-01, CLB-02, CLB-03, GIT-01, GIT-02, GIT-03
**Success Criteria** (what must be TRUE):
  1. User can copy a decrypted value to clipboard; the clipboard clears automatically after the configured timeout (default 30s)
  2. Clipboard is cleared synchronously when `sops-tui` exits via any path including SIGINT and SIGTERM
  3. Files with uncommitted git changes display a badge ([M], [A], [?]) in the file browser
  4. User can view git blame and commit history for any secret file from within the TUI
  5. User can fuzzy search across all files and key names simultaneously with `/`
**Plans:** 3 plans

Plans:
- [x] 04-01-PLAN.md — Clipboard copy with auto-clear, signal safety, Phase 4 styles and keybindings
- [x] 04-02-PLAN.md — Git backend (go-git v5), change badges in file list and detail breadcrumb
- [x] 04-03-PLAN.md — Git history overlay, cross-file fuzzy search across all files and key paths

**UI hint**: yes

### Phase 5: Power Features
**Goal**: Users can manage age recipients across files and audit secret health — the highest-risk multi-file operations
**Depends on**: Phase 4
**Requirements**: RCP-01, RCP-02, RCP-03, HLT-03
**Success Criteria** (what must be TRUE):
  1. User can view the list of age key recipients configured for any file
  2. User can add or remove an age key recipient on a file with confirmation before re-key
  3. User can bulk re-key multiple files to a new recipient set with per-file confirmation
  4. User can run a health check that reports weak secrets, duplicate values across files, and stale (unchanged) values
**Plans:** 4 plans

Plans:
- [x] 05-01-PLAN.md — Health check package, sops executor recipient functions, git staleness, styles and keybindings
- [x] 05-02-PLAN.md — HealthModel overlay, RecipientFormModel with age key validation
- [x] 05-03-PLAN.md — AppModel wiring: file selection, bulk re-key, recipient flows, health pipeline
- [x] 05-04-PLAN.md — Integration test suite review and human verification of all flows

**UI hint**: yes

---

## Milestone v1.1 — k9s Visual Parity

**Goal:** Reshape the sops-tui UI shell so it looks and behaves like k9s — persistent keybinding menu, ASCII logo, header info panel, titled bordered content, colored breadcrumb chips — while keeping every v1.0 functional loop intact. Pure chrome rework; no framework swap; no behavioural changes to the read / write / clipboard / git / health / recipients loops.

**Scope boundary:** Six phases (6 through 11). Phase 6 (layout groundwork + testing harness) is non-skippable and **must land before any chrome rendering** per Pitfall 1 in research/PITFALLS.md — adding chrome without migrating the 15 existing `m.height - statusBarHeight(m)` call-sites first produces invisible "content painted under header" bugs visible only at large terminal sizes. All v1.0 functional tests (v1.0 Phases 1-5) must continue to pass through every v1.1 phase.

### Explicitly Out of Scope for v1.1

The following k9s-inspired features are **not** v1.1 scope and must not re-emerge during phase planning. Most are enumerated as anti-features in `research/FEATURES.md` Category 9 and in `research/SUMMARY.md` under "Explicitly NOT v1.1":

- **Polling goroutines** (`clusterUpdater` / env re-check every 15s) — no external data source that mutates outside the TUI
- **`:` command bar** with destructive actions — breaks the "diff-before-re-encrypt" safety net
- **Log forwarding / streaming** — secrets don't produce logs
- **Image vulnerability scanner** — we aren't a package scanner
- **Port-forward manager** — K8s-specific, no analogue
- **Alias file hot-reload** — fsnotify complexity for no v1.1 user value
- **Dialog severity popups** — existing `internal/ui/errorbox.go` is sufficient
- **Context switcher** — one `.sops.yaml` per directory, already handled at startup
- **Read-only global mode indicator** — per-file git badges already cover the concept
- **Namespace selector** — no analogue in SOPS
- **GitHub update phone-home** — trust violation for a secrets tool
- **Splash screen / big logo on startup** — nostalgic; zero user value
- **Mouse interactions** — project is keyboard-only by core value
- **`:` command tab-completion / number-key view switching / `[`/`]` history** — v1.2+
- **Animated logo / pulsing / gradients** — adds ticker goroutine, breaks teatest snapshots
- **User skin YAML loader** (THM-01) — deferred to v2; v1.1 ships a single well-tuned default palette only
- **Builtin skins dracula/gruvbox/monokai embedded in binary** (THM-02) — deferred to v2
- **Live skin reload via fsnotify** (THM-03) — deferred to v2
- **Live `:skin <name>` runtime switcher** — depends on `:` command bar (v1.2)
- **Responsive narrow-terminal column hiding** — Phase 10 covers "ugly but survives"; richer responsive behaviour deferred to first bug report

If any of these resurface during `/gsd-plan-phase`, reject in plan review.

### Phase 6: Layout Groundwork
**Goal**: Height arithmetic is centralised behind a single helper and a regression-proof test harness is in place, so that every downstream chrome phase can size its children correctly without re-auditing 15 SetSize call-sites
**Depends on**: Phase 5 (v1.0 functional core complete)
**Requirements**: UI-17, UI-18, UI-19
**Success Criteria** (what must be TRUE):
  1. A single `bodyDims(m) (w, h int)` helper in `internal/app/model.go` returns the content region after subtracting chrome + crumbs + status-bar heights; every existing `m.height - statusBarHeight(m)` call-site is migrated to it
  2. A CI grep-gate fails the build if `m.height - statusBarHeight` arithmetic appears anywhere outside the `bodyDims` helper
  3. Test harness exposes an ANSI-stripped golden-file comparison helper; color presence can be asserted separately from structural layout
  4. Resize behaviour at 40×12, 80×24, 120×40, and 200×60 is verified by teatest snapshots — no content is painted under the status bar at any size
  5. Phase ships zero visible change to end users; it is pure refactor + test infrastructure
**Plans:** 2/2 plans complete

Plans:
- [x] 06-01-PLAN.md — bodyDims/chromeHeight/crumbsHeight helpers, LIVE TestBodyDimsMigration grep-gate, internal/testutil golden harness, BenchmarkAppView v1.0 baseline
- [x] 06-02-PLAN.md — 17-site atomic migration to bodyDims, plan2MigrationAllowlist deletion, four resize goldens (40x12/80x24/120x40/200x60), .gitattributes

**UI hint**: yes

### Phase 7: Chrome Skeleton
**Goal**: Users see a persistent ASCII logo and multi-column keybinding menu in the header, and every primary view is wrapped in a titled bordered region — the first visible "it looks like k9s" step
**Depends on**: Phase 6
**Requirements**: UI-01, UI-02, UI-06, UI-15
**Success Criteria** (what must be TRUE):
  1. A persistent multi-column keybinding menu is rendered on every view; no `?` press is required to discover the core hotkeys
  2. A 6-row ASCII logo (~26 columns wide) is anchored to the top-right of the header on every view
  3. Every primary view (Files, Detail, Metadata, Diff, Help, History, Health, Recipients, RecipientForm) renders inside a titled bordered region whose title encodes the view name and, where relevant, an item count
  4. Only `lipgloss.NormalBorder()` appears in chrome rendering code; persistent chrome is ASCII-only (emoji-free); a CI grep-gate prevents regressions
  5. `AppModel.View()` composes `[header][optional crumbs][titled body][status bar]` AND `BenchmarkAppView` stays ≤ 50 µs/op at 200×60 with no `lipgloss.NewStyle()` inside `View()`
**Plans:** 3 plans

Plans:
- [x] 07-01-PLAN.md — Primitives: MenuHint/Hinter/HintsFromBindings + 5 inline hint-set vars, 6-row ASCII logo (Candidate A), RenderMenu via lipgloss/v2/table, 8 new style vars in styles.go
- [x] 07-02-PLAN.md — Chrome composer: RenderChrome + WrapTitled + overlayTitle string-splice (closes STATE.md overlayTitle research gap), corners/width/truncation/empty-title unit tests
- [x] 07-03-PLAN.md — Integration: flipped chromeHeight, rewrote View(), Hints() on 8 sub-models + CommitCount/FindingCount accessors, menuHints dispatcher, titleForState, migrated magic -4, 4 grep-gates (ASCII-only, NormalBorder-only, no-NewStyle-in-View AST walker, bench-budget at 5 ms — see 07-03-SUMMARY for the 50 µs deviation rationale), 4 goldens refreshed

**UI hint**: yes

### Phase 7.1: Chrome Gap Closure (INSERTED)
**Goal**: Close the Phase 7 verification gaps — restore the locked SC5 governance contract (revert the unauthorized 100×-loosened test-gate and the unauthorized SC5 wording amendment, defer perf to Phase 11 SC2 with the original 50 µs target preserved), strip the nested `RoundedBorder` boxes from the 6 sub-model `View()` methods so `WrapTitled` is the single border source, clamp narrow-terminal chrome so the body remains reachable at 40×12 and 80×24, extend the `TestViewNoNewStyle` AST walker so it reaches helpers and sub-models reachable from `View()`, and align the menu test allowlist with the chrome contract
**Depends on**: Phase 7
**Inserted because**: Phase 7 verification (`07-VERIFICATION.md`, status `gaps_found` 4/5) and code review (`07-REVIEW.md`, 5 WARNINGs / 7 INFO) surfaced one BLOCKER (SC5 budget breached + unauthorized ROADMAP/test-gate amendments) plus quality concerns that the verifier flagged as candidates for either a Phase 7.1 hotfix or Phase 8 cleanup. Closing them in 7.1 before Phase 8 starts keeps the chrome contract honest and prevents the Phase 8 info-panel work from inheriting double-border + width-math drift.
**Requirements**: UI-15 (chrome-contract integrity — currently SATISFIED but degraded by amendment), UI-16 (narrow-terminal layout — partial coverage; full survival stays in Phase 10), UI-21 (deferred to Phase 11 SC2 with original 50 µs target restored)
**Success Criteria** (what must be TRUE):
  1. ROADMAP SC5 wording is restored to the locked `BenchmarkAppView ≤ 50 µs/op at 200×60` text; `internal/app/chrome_test.go` has `budgetNs = 50_000` with `TestBenchmarkAppView_UnderBudget` calling `t.Skip("deferred to Phase 11 SC2 — D-18 caching fallback")`; `.planning/phases/07-chrome-skeleton/07-DISCUSSION-LOG.md` carries an `[APPROVED] 2026-04-27` entry recording the deferral decision and the Phase 11 owner
  2. Helpers reachable from `AppModel.View()` are scanned by `TestViewNoNewStyle` (BFS over same-package call edges rooted at `View`); `lipgloss.NewStyle()` calls in `renderRecipientList` (model.go:1935, 1940), `renderFormatMenu` (model.go:1981), and the 12 sub-model `View()` call-sites (`help.go:82,98`, `health.go:128,178`, `recipientform.go:132,163`, `metadata.go:83-85,158,174`, `diff.go:177`, `history.go:86,134`) are gone — replaced by package vars declared in `internal/ui/styles.go`
  3. `HelpModel.View()`, `MetadataModel.View()`, `DiffModel.View()`, `HealthModel.View()`, `HistoryModel.View()`, `RecipientFormModel.View()` return inner content only — no inner `RoundedBorder`/`Background`/`Padding` envelope; `WrapTitled` is the sole border source for these 6 states; `m.width` arithmetic is corrected so inner content area is `m.width − 4` (chrome border + Padding(0,1)) not `m.width − 6`; refreshed goldens at 120×40 and 200×60 show single-border framing
  4. At terminal widths below `infoPanelWidth + logoWidth + minMenuCol`, `RenderChrome` drops the info-panel slot and renders menu+logo only (or a one-line `press ? for help` fallback below `logoWidth + 8`); `RenderMenu` uses manual `JoinHorizontal` of two pre-rendered columns instead of `lipgloss/v2/table` so cell-wrapping never engages; `internal/app/testdata/resize_40x12.golden` and `resize_80x24.golden` show ≤ 6 chrome rows with the body region intact and reachable
  5. `internal/ui/menu_test.go` allowlist contains only the runes `\n ─ │ ┌ ┐ └ ┘ ↑ ↓ ← →` (no `╭╮╰╯`); the quit-hint visibility decision for `stateRecipientConfirm` and `stateBulkReKeyConfirm` is recorded as inline doc comments on `keys.RecipientConfirmHints`, `keys.BulkReKeyConfirmHints`, and the dispatcher arms in `AppModel.menuHints()` with a UI-SPEC cross-reference; `chrome_test.go` `TestChromeASCIIOnly` allowlist is the single source of truth (extracted into a shared test helper if option B from WR-04 is taken)
**Plans:** 5 plans

Plans:
- [x] 07.1-01-PLAN.md — Restore SC5 governance (revert ROADMAP+test gate, [APPROVED] DISCUSSION-LOG entry, 07-03-SUMMARY closure pointer)
- [x] 07.1-02-PLAN.md — Align menu_test.go allowlist with chrome_test.go canonical + document quit-hint suppression in hints.go + model.go + 07-UI-SPEC.md
- [x] 07.1-03-PLAN.md — Strip nested RoundedBorder envelope from 6 sub-models, lift surviving NewStyle calls, refresh 120x40+200x60 goldens
- [x] 07.1-04-PLAN.md — Rewrite TestViewNoNewStyle as BFS reachability walker, lift 3 model.go NewStyle calls, add sub-model AST walker in internal/ui
- [x] 07.1-05-PLAN.md — Narrow-terminal chrome clamp (RenderChrome 3-tier fallback + RenderMenu manual columns), 4 new chrome composition tests, refresh 40x12+80x24 goldens, update 07-VERIFICATION.md frontmatter

**UI hint**: yes

### Phase 8: Header Info Panel + Crumb Chips
**Goal**: Users see the context they're working in (`.sops.yaml` path, age key fingerprint, recipient count, git state, file count) in a top-left info panel, and the breadcrumb moves to a dedicated row of colored chip pills above the body
**Depends on**: Phase 7
**Requirements**: UI-04, UI-05, UI-07, UI-08
**Success Criteria** (what must be TRUE):
  1. A 5-row info panel renders in the top-left of the header showing: `.sops.yaml` relative path, age key fingerprint, recipient count, git branch + clean/dirty marker, file count
  2. Info-panel values are de-PII'd before render: age fingerprint ≤10 chars with visible ellipsis, all paths are repo-relative (never `$HOME`-absolute), no copy bindings ever target chrome content
  3. Breadcrumb segments render as colored chip pills (`<files>`, `<prod.yaml>`); the active (last) segment uses the accent color; the legacy ` > ` text separator is gone
  4. The breadcrumb row sits between the header and the titled body; the status bar at the bottom contains only right-aligned env indicators + clipboard state
  5. Info-panel fields are cached on `AppModel` and refreshed on events (file discovery, recipient change, git status change), not stat'd on every frame
**Plans:** 3 plans

Plans:
- [x] 08-01-PLAN.md — Primitives: InfoPanelData + RenderInfoPanel + middleTruncate, RenderCrumbs + truncateSegmentsToWidth, ParseAgeKeyFingerprint + AgeKeyFilePath, 8 new style vars in styles.go (Wave 1)
- [x] 08-02-PLAN.md — git.GetBranch helper (3-subtest coverage) + status-bar shrink (right-aligned env+clipboard only, drop renderBreadcrumb, neuter SetItemCount, add Segments() accessor) (Wave 1)
- [ ] 08-03-PLAN.md — Integration: AppModel.infoPanel cache + 4 refresh seams + RenderChrome signature change + crumbsHeight flip + grep-gate scope extension + 4 resize goldens regen + D-220 5-question security review sign-off (Wave 2)
**UI hint**: yes

### Phase 9: Keybinding Discoverability
**Goal**: Every interactive sub-model exposes its hotkeys through a single `Hints()` interface derived from its existing keymap, and the persistent menu re-hydrates per view — so the always-visible hints never diverge from the keys that actually work
**Depends on**: Phase 7 (menu exists), Phase 8 (remaining sub-models have their chrome integration done)
**Requirements**: UI-09, UI-10, UI-11
**Success Criteria** (what must be TRUE):
  1. Every interactive sub-model (FileList, Detail, Help, Diff, Health, History, RecipientForm, Metadata) exposes a `Hints() []keys.MenuHint` method derived from its existing `key.Binding.ShortHelp()` — the keymap is the single source of truth
  2. The persistent menu re-hydrates from the active sub-model's `Hints()` on every `View()` call; modal states (diff, recipient confirm, bulk re-key) show their modal keybindings, not the underlying file-list ones
  3. Menu hint dispatch is a pure function of `(state, recipientAction, IsSearchActive)` — golden-file teatest covers every `(state, action)` tuple that shares a sub-model
  4. The `?` full-screen help overlay is retained as the complete reference; day-to-day hotkeys remain discoverable in the persistent menu without opening it
  5. Changing a `key.Binding` value automatically updates the rendered menu — no second "also update the menu text" edit is ever required
**Plans:** 2 plans
**UI hint**: yes

### Phase 10: Theming + Accessibility
**Goal**: The logo communicates aggregate app severity, the default palette matches k9s conventions, and the UI survives 16-color terminals, colorblindness, and widths from 40 to 200 columns
**Depends on**: Phase 8 (chrome data paths exist), Phase 9 (menu driven by hints)
**Requirements**: UI-03, UI-12, UI-13, UI-14, UI-16
**Success Criteria** (what must be TRUE):
  1. The logo recolours to reflect aggregate app status (info / warn / error) derived from env checks, flash severity, and aggregate health; flash call-sites in `model.go` upgrade to typed `FlashInfo`/`FlashWarn`/`FlashErr` severity
  2. Default palette is tuned to k9s conventions (accent shifted toward hot-pink/purple) while keeping the explicit-color / AdaptiveColor-ban rule from v1.0
  3. On 16-color terminals (`TERM=xterm` / Ascii profile) a safe fallback palette is applied so paired bg/fg chips and menu cells remain legible; teatest runs across Ascii / ANSI / ANSI256 / TrueColor profiles
  4. Every color-coded state (info / warn / error, active vs inactive chip, env indicators, flash severity) uses redundant shape or text encoding (`[I]` / `[W]` / `[E]` prefixes, inverted bg+fg for active, underline for focus) so the UI remains usable for colorblind users
  5. Rendering at 40×12 through 200×60 never corrupts the layout; narrow-terminal rendering may be visually dense but must not truncate critical data or overflow the viewport (middle-segment crumb ellipsis kicks in when needed)
**Plans:** 3 plans
**UI hint**: yes

### Phase 11: Regression + Perf Gates
**Goal**: v1.1 is release-ready — no v1.0 feature has regressed, render performance meets the target, and the chrome survives the terminal emulators that matter
**Depends on**: Phase 10
**Requirements**: UI-20, UI-21
**Success Criteria** (what must be TRUE):
  1. All v1.0 functional integration tests (file discovery, reveal, edit, diff, rotate, clipboard, git, recipient management, health) pass unchanged after chrome lands — no v1.0 feature regresses
  2. `BenchmarkAppView` stays ≤ 50 µs/op at 200×60 with the full chrome rendered; `pprof` shows no `lipgloss.NewStyle()` calls inside `View()` — all styles are declared as package vars
  3. Chrome renders correctly on the supported terminal matrix (Alacritty, Ghostty, macOS Terminal, iTerm2, Windows Terminal, WSL2, VSCode integrated terminal, tmux-nested); screenshot record captured per combo
  4. Alt-screen cleanup is explicit: paint fill frame on enter, blank frame on exit; no residual chrome survives in the user's shell prompt area after TUI exit
  5. The full 15-item "Looks Done But Isn't" checklist from `research/PITFALLS.md` is signed off
**Plans:** 2 plans
**UI hint**: yes

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> 3 -> 4 -> 5 -> 6 -> 7 -> 7.1 -> 8 -> 9 -> 10 -> 11

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation | 4/4 | Complete | - |
| 2. Read Loop | 3/3 | Complete | 2026-04-14 |
| 3. Write Loop | 3/3 | Complete | - |
| 4. Clipboard & Git | 3/3 | Complete | - |
| 5. Power Features | 4/4 | Complete | - |
| 6. Layout Groundwork | 2/2 | Complete | 2026-04-24 |
| 7. Chrome Skeleton | 3/3 | Complete | 2026-04-27 |
| 7.1. Chrome Gap Closure (INSERTED) | 4/5 | Executing | - |
| 8. Header Info Panel + Crumb Chips | 0/3 | Planned | - |
| 9. Keybinding Discoverability | 0/2 | Planned | - |
| 10. Theming + Accessibility | 0/3 | Planned | - |
| 11. Regression + Perf Gates | 0/2 | Planned | - |
