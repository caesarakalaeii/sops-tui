# Feature Research — Milestone v1.1 (k9s Visual Parity)

**Domain:** k9s-style visual shell for a secrets TUI
**Researched:** 2026-04-23
**Confidence:** HIGH (grounded in k9s source at `~/git/k9s`, cross-checked against 35 shipped skins and existing `internal/ui/styles.go`)

## Scope

This file catalogues what a k9s-style visual shell actually does, so v1.1 can reshape sops-tui's chrome without inventing k9s behavior from memory. The functional core (read / write / clipboard / git / health / recipients) is frozen; every item below is purely UI-shell behaviour.

Source references below cite the k9s tree under `~/git/k9s/`:

- `internal/ui/menu.go` — persistent keybinding menu
- `internal/ui/logo.go` + `internal/ui/splash.go` — ASCII logo, colour states
- `internal/view/cluster_info.go` — header info panel
- `internal/ui/crumbs.go` — breadcrumb chips
- `internal/ui/table_helper.go` + `internal/ui/table.go` (lines 664–720) — bordered content regions with title + counter + filter
- `internal/ui/prompt.go` — `:` command bar / `/` filter prompt
- `internal/view/app.go` — top-region layout (lines 166–183, 273–331), toggles, and status→logo routing (lines 575–612)
- `internal/ui/action.go` + `internal/model/menu_hint.go` — how views expose their hotkeys to the menu
- `skins/*.yaml` — 35 shipped skins (dracula, gruvbox-dark, monokai, etc.)

---

## Feature Landscape

Each category below is independent and can ship in its own phase. Every row notes **Complexity (S/M/L)** and **Depends on** (existing sops-tui code the item must integrate with).

**Complexity key:**
- **S** — a few hours: new lipgloss style, new small component, pure render pass.
- **M** — half-day to a day: new component with state, needs to plug into `Model.Update`, modifies layout arithmetic.
- **L** — multi-day: touches persistence, config file schema, or requires refactoring existing views.

### Category 1 — Header Region (top strip)

k9s allocates a 7-row horizontal flex at the top: `[ClusterInfo 50-col][Menu flex][Logo 26-col]` (`internal/view/app.go:305-331`). `buildHeader` sizes the cluster-info column to `len(clusterName)+15` but never shrinks below 50. The whole header is togglable with `Ctrl-E`.

#### Table Stakes

| Feature | Why Expected | Complexity | Depends on |
|---|---|---|---|
| Three-column top strip (info left, menu center, logo right) | This is literally what makes a TUI "look like k9s". A bottom-status-bar-only UI does not. Fixed-height rows (6 for logo, 1 for message, 7 total per `internal/view/app.go:305` + `internal/ui/logo.go:33-35`). | M | New container in `internal/ui/` composing existing views; `app.Model.View()` (`internal/app/model.go`) currently returns a single `tea.View` — needs a header region above current content. |
| Header honours `lipgloss.Height` so content region shrinks correctly | k9s hard-codes row heights in `tview.Flex.AddItem(view, 7, 1, false)`. Bubble Tea requires us to subtract header height from the content viewport manually or content clips. | S | `internal/ui/filelist.go`, `internal/ui/detail.go`, `internal/ui/health.go` — all currently assume full window height; they receive `WindowSizeMsg` from `app.Model`. |
| Toggle to hide header (power users want more rows) | k9s binds `Ctrl-E` for "ToggleHeader" (`internal/view/app.go:256, 635-646`). Users running in split panes need this. | S | Keybinding registration in `internal/app/model.go`; the header container should accept a `Hidden bool`. |

#### Differentiators

| Feature | Value Proposition | Complexity | Depends on |
|---|---|---|---|
| Header height adapts to narrow terminals (hide logo below ~100 cols, hide info below ~70 cols) | k9s is brittle on narrow terminals — it just clips. We can do better because Bubble Tea's `WindowSizeMsg` is already routed everywhere. | M | `app.Model` already receives `tea.WindowSizeMsg`; add width thresholds before composing header children. |
| Responsive column widths (info column grows with longest value, not fixed 50) | k9s's `clusterInfoWidth = 50` + `clusterInfoPad = 15` heuristic (`internal/view/app.go:40-41, 313-322`) is a legacy hack. We can measure content with `lipgloss.Width` at each render. | S | New sizing helper in `internal/ui/` header package. |

#### Anti-Features

| Feature | Why Requested | Why Problematic | Alternative |
|---|---|---|---|
| Make header collapsible with a mouse click | Feels modern | Breaks keyboard-only paradigm that is the project's core value. Adds mouse-event plumbing complexity to Bubble Tea v2. | Keyboard toggle only (`Ctrl-E`); document it on the help overlay. |
| Animate header slide-in on startup | "Looks polished" | Adds time-based state, invites flicker in CI/SSH terminals, and conflicts with `lipgloss.NoColor` test harness. `internal/view/app.go:550-561` shows k9s uses a 1-second splash delay then instant swap — no animation. | Splash page (ASCII logo only) that disappears on first keystroke; instant header render after. |

---

### Category 2 — Info Panel (top-left, `ClusterInfo` analogue)

k9s's `ClusterInfo` is a 2-column `tview.Table`: left column section labels with trailing colon ("Context:", "Cluster:", "K9s Rev:", etc.), right column values styled with `Info.SectionColor` bold (`internal/view/cluster_info.go:67-97, 162-172`). It refreshes every 15s from a `clusterUpdater` goroutine (`internal/view/app.go:39, 361-392`).

#### Table Stakes

| Feature | Why Expected | Complexity | Depends on |
|---|---|---|---|
| Left-aligned `Label:` / `Value` table with ~5-8 rows | This is exactly `ClusterInfo.layout()` (`internal/view/cluster_info.go:67-72`). For sops-tui the rows are: `.sops.yaml`, `Repo`, `Files`, `Recipients`, `Age key`, `Sops`. | M | `internal/validator/` already computes `.sops.yaml` path, repo root, and sops/age presence at startup. `internal/keys/` parses the age keyring. `internal/app/model.go` holds the file count. No data plumbing needed — all values already exist in `Model`. |
| Bold-but-muted section labels, bright values | `sectionCell.SetAlign(AlignLeft)` with `Info.FgColor` for labels; `infoCell` with bold `SectionColor`. Reversing this (bright labels, dim values) is wrong per every shipped k9s skin. | S | `internal/ui/styles.go` — add `InfoLabelStyle` + `InfoValueStyle` (muted fg + bold accent fg respectively). |
| Values that can warn when unhealthy | `ClusterInfoChanged` swaps value cells to `[orangered::b]` when N/A (`cluster_info.go:104-110, 132-137`). sops-tui equivalents: `age key: missing` in warning colour, `sops: v3.12.2 ✓` in success colour. | S | `internal/validator/` already returns typed results (has/missing) — plumb a severity per row. |
| N/A placeholder when a value can't be computed | `render.NAValue` (`cluster_info.go:71, 94`). Prevents visual gaps that break alignment. | S | Trivial — render constant `"n/a"` in muted colour. |

#### Differentiators

| Feature | Value Proposition | Complexity | Depends on |
|---|---|---|---|
| Age key fingerprint (last 8 chars) in label | k9s shows short revs ("v0.32.0 ⚡️v0.32.1" in `cluster_info.go:125-129`). Showing `age1…xk8h` is unique to a secrets tool and immediately reassures the user their key is loaded. | S | `internal/keys/` already parses public keys; just truncate and display. |
| Update-available indicator for sops binary | k9s does this with `curr.K9sLatest` + `⚡️` suffix. For us: compare installed sops version with latest GitHub release (cached for 24h). | L | New module under `internal/validator/`; needs HTTP + cache file. Candidate for **v1.2**, not v1.1 — flagged in roadmap. |
| "Dirty files" counter when repo has uncommitted SOPS files | Mirrors k9s `CPU %` / `MEM %` as a living health indicator. | S | `internal/git/` already detects uncommitted changes for badge rendering. |

#### Anti-Features

| Feature | Why Requested | Why Problematic | Alternative |
|---|---|---|---|
| Live "last-decrypted-at" clock in info panel | k9s has a ticking CPU% that feels alive | Our app is not cluster-coupled — there is no ambient process to display. A ticking clock burns CPU and adds a goroutine for decoration. | Static values refreshed only when underlying state changes (`WindowSizeMsg`, file edit, recipient change). |
| Info panel polling goroutine like k9s's `clusterUpdater` (15s tick, exp backoff) | "But k9s does it" | k9s polls because the K8s API is authoritative and mutates outside k9s. Our data source (filesystem + age keyring) does not change without user action inside the TUI. A poller here is pure waste. | Event-driven refresh via Bubble Tea messages only. |
| Replace sops-tui's startup error box with info-panel warnings | "Cleaner" | Missing `sops` is fatal — we must not launch into a half-broken TUI that hides the fact. Existing startup validator (`internal/validator/`) is correct behavior. | Keep startup gate for hard errors; use info panel for soft state. |

---

### Category 3 — ASCII Logo + Status Coupling (top-right)

`Logo` is a vertical flex: 6-row ASCII art above a 1-row status message (`internal/ui/logo.go:32-39`). Its key behavior is **status propagation** — calling `a.Status(FlashInfo, msg)` from anywhere routes via `setLogo` (`internal/view/app.go:600-611`) into `logo.Info/Warn/Err`, which recolours **both** the art and the status row with `LogoColorInfo/Warn/Error` (`internal/ui/logo.go:73-91`).

#### Table Stakes

| Feature | Why Expected | Complexity | Depends on |
|---|---|---|---|
| 6-row, ~24-col monospace ASCII logo | Fixed to 6 rows so row-count math is stable (`AddItem(l.logo, 6, 1, false)`). The sops-tui logo must also be 6 rows to share header height with `ClusterInfo`. | S | Create `internal/ui/logo.go`; store art as `[]string` the way `LogoSmall` does (`internal/ui/splash.go:15-22`). |
| Logo colour changes on app status (info=accent, warn=yellow, error=red) | `logo.update(msg, color)` recolours both panels atomically under a mutex (`internal/ui/logo.go:88-91, 93-112`). Colours come from `styles.Body().LogoColor*` fields — see `skins/monokai.yaml:26-30` (explicit `logoColorInfo`, `logoColorWarn`, `logoColorError`). | M | `internal/ui/styles.go` already defines `ColorSuccess/ColorWarning/ColorError`. Need a `Logo.SetStatus(level, msg)` call reachable from `app.Model` actions. |
| Status row below logo shows 1-line message + has coloured background at the level colour | `refreshStatus` sets `l.status.SetBackgroundColor(c.Color())` — it's a full-width colour band (`internal/ui/logo.go:93-101`). This is the single most visually distinctive k9s feature. | S | New lipgloss style `LogoStatusBand` per severity. |
| Reset/clear on idle | `ClearStatus` is called when things return to normal (`internal/view/app.go:591-598`). Without this the logo stays red forever after one error. | S | Route to `Logo.Reset()` on successful action completion. |

#### Differentiators

| Feature | Value Proposition | Complexity | Depends on |
|---|---|---|---|
| Logo colour driven by aggregate health (any missing recipient across files, any weak secret detected) | Static colour = boring. Live severity colour is immediately useful at-a-glance for a secrets audit tool. | M | `internal/health/` already produces severity levels. Compute max severity and feed to `Logo.SetStatus`. |
| Splash screen with big ASCII logo on startup (pre-main UI) | k9s has it (`internal/ui/splash.go:25-32`, big logo variant). Pleasant detail when paired with a short delay to hide the first-paint layout thrash. | S | New `internal/ui/splash.go`. Pair with `splashDelay = 1 * time.Second` like k9s (`internal/view/app.go:38`). |

#### Anti-Features

| Feature | Why Requested | Why Problematic | Alternative |
|---|---|---|---|
| Animated logo (rotating, pulsing, gradient) | "Fun" | Requires ticker goroutine, steals CPU, breaks golden-file TUI tests (`teatest` snapshots become non-deterministic). k9s never animates its logo. | Static art; colour changes only on state transitions. |
| Benchmarking indicator in status row (k9s `IsBenchmarking` at `internal/ui/logo.go:62-65`) | "k9s has it" | We have no benchmark concept for secrets. Copying it is cargo-culting. | Skip entirely. The `Logo.IsBenchmarking()` method does not translate. |
| Per-view logo (different logo on Files / Health / etc.) | "Distinctive" | k9s uses one logo. Changing visual identity mid-navigation is disorienting. The *status colour* should change, not the art. | One logo; status row indicates current mode. |

---

### Category 4 — Persistent Keybinding Menu (center of header)

The menu is a `tview.Table` laid out in columns of up to 6 rows (`internal/ui/menu.go:22-24, 117-149`). Number-keyed hints (e.g. namespace favourites) go in the first column; letter/symbol hints fan into subsequent columns. **Crucially it re-hydrates automatically** on view stack changes via `StackPushed/StackPopped/StackTop` (`menu.go:59-76`) — each view implements `Hints() model.MenuHints` (`internal/view/types.go:56`, example at `internal/view/log.go:199-201`), and the active view's hints replace the menu content on push.

Each hint renders as ` <key> description ` with the key in bold primary colour and description in frame foreground (`menu.go:199-215`). On macOS `alt` gets rewritten to `opt` (`menu.go:175-184`).

#### Table Stakes

| Feature | Why Expected | Complexity | Depends on |
|---|---|---|---|
| Always-visible grid of hotkeys with key + description | This is what replaces "help is hidden behind `?`". Missing = app does not look like k9s. Max 6 rows, column width = longest mnemonic + 2 (`menu.go:122-140, 208-215`). | M | `internal/ui/help.go` has a `KeyBinding` slice per view — reuse as the hint source. `app.Model.View()` needs to compose the menu into the header. |
| Menu re-hydrates when active view changes | Without this the menu shows stale keys (file-view hints while looking at health view). k9s wires it via `Content.AddListener(a.Menu())` (`internal/view/app.go:103`). | M | Bubble Tea doesn't have a stack listener pattern natively. We can emit a `ViewChangedMsg` from `app.Model.Update` when `state` changes, and the menu component listens. |
| View-local hint provider interface | Each view owns its hints (`Hints() model.MenuHints`). Copy the pattern to force each `ui/*.go` view to declare its hotkeys in one place. | M | Add `Hints() []HelpEntry` to a common interface in `internal/ui/`; implement on `FileList`, `Detail`, `Health`, `Recipients`, `History`. Right now hints are scattered in `help.go`. |
| Number keys rendered distinctly (bold, primary colour) | `formatNSMenu` vs `formatPlainMenu` (`menu.go:199-215`). k9s uses `<1>..<9>` as "favourites shortcut" cues. For us, number keys can jump between main views (files/health/recipients) — see Category 7. | S | New lipgloss style `MenuNumKeyStyle`. |
| Chord keys like `Alt-R`, `Ctrl-E` rendered readably | `ToMnemonic` wraps keys in `< >` and lowercases (`menu.go:191-197`). "Ctrl+E" as plain text reads like an instruction; "\<ctrl-e\>" reads like a key. | S | Helper function mirroring `ToMnemonic`. |

#### Differentiators

| Feature | Value Proposition | Complexity | Depends on |
|---|---|---|---|
| Contextual hints hide/show based on state (edit mode only shows Save/Cancel) | k9s has `Visible bool` on each `MenuHint` (`internal/model/menu_hint.go:14, 18-20`) and hides invisible ones from the menu — visible-but-sorted-last, blank ones drop. This makes the menu feel "live" rather than "fixed list". | M | Existing modal states in `app.Model` (`stateEdit`, `stateRotate`, `stateDiff`) already track mode — surface `Hints()` per state. |
| "More keys" indicator when hints overflow the visible rows | k9s's `maxRows = 6` will just clip. We can render a `+N more → ?` cue pointing at the full-screen help overlay. | S | Pure render logic. |
| Hints include data badges (e.g. `<e> edit (readonly)` in dim when file is read-only) | Makes the menu doubly useful as status surface. | S | `HelpEntry.Suffix string` field. |

#### Anti-Features

| Feature | Why Requested | Why Problematic | Alternative |
|---|---|---|---|
| Mouse-clickable menu rows | "Discoverability" | Project's core value is keyboard-only. Mouse routing in Bubble Tea v2 is possible but adds state complexity (`tea.MouseMsg` vs `tea.KeyMsg`). | Keyboard only; `?` overlay for discoverability. |
| Search-within-menu (type to filter hints) | "Nice on big menus" | Competes with the `/` fuzzy-search hotkey that is already a core feature. Two different meanings of `/` is a UX bug. | If hints overflow, surface the `?` full-help overlay (already exists). |
| Auto-scroll menu contents | "When there are many hints" | k9s explicitly never scrolls — it adds more columns instead (`menu.go:136-140`: `if row >= maxRows { row, col = 0, col+1 }`). Scrolling a menu hides state. | Multi-column layout; overflow indicator. |

---

### Category 5 — Content Framing (titled borders around Files / Keys / Diff / etc.)

k9s wraps every browser/table in a border with a title string that embeds the **resource name + optional namespace + row count** (`internal/ui/table_helper.go:26-30`):

```
TitleFmt   = " [fg:bg:b]%s[fg:bg:-][[count:bg:b]%s[fg:bg:-]][fg:bg:-] "
NSTitleFmt = " [fg:bg:b]%s([hilite:bg:b]%s[fg:bg:-])[fg:bg:-][[count:bg:b]%s[fg:bg:-]][fg:bg:-] "
```

This renders as ` Pods(default)[23] ` — three colour slots: title foreground, highlight (namespace), counter. Filter state appends `<[filter:bg:r]/nginx[fg:bg:-]>` (`table_helper.go:24`).

The title is refreshed via `Table.UpdateTitle()` after every data mutation (`ui/table.go:664-719`).

#### Table Stakes

| Feature | Why Expected | Complexity | Depends on |
|---|---|---|---|
| Every primary view has a titled border | Users visually anchor on the title. Missing = the UI feels like a pile of panels. | M | lipgloss supports bordered blocks via `Border()` + `BorderTitle()` equivalent. We'll need a `TitledPanel` wrapper in `internal/ui/` that takes any `View()` producer and frames it. |
| Title embeds a count (e.g. `Files [37]`) | `NSTitleFmt` with `render.AsThousands(rc)` (`table.go:702`). Counts are information-dense — users learn to read them before scanning rows. | S | `internal/ui/filelist.go` already knows file count. Same for `health.go`, `recipientform.go`. |
| Title updates reactively (adding a recipient -> count bumps) | `UpdateTitle` runs after model mutations. Without this counts go stale. | S | Call the title builder on every `View()` — lipgloss render is cheap. |
| Filter/search state shown in title | The `<[filter:bg:r]/query[fg:bg:-]>` suffix (`table_helper.go:24`) — survives view switches so users see "yes I'm filtered" in the title, not hidden in a status bar. | S | Existing fuzzy search state (`internal/ui/search.go`) already exposes the query string. |
| Focus indicator (border colour changes for focused pane) | `styles.Frame().Border.FocusColor` (e.g. `dracula.yaml:48`: `focusColor: *current_line`). Required once we have >1 pane on screen. | S | lipgloss `BorderForeground` per-pane based on `app.Model.focus` enum. |

#### Differentiators

| Feature | Value Proposition | Complexity | Depends on |
|---|---|---|---|
| Titles encode semantic mode (`Diff [key path]`, `History [filename]@HEAD`) | More informative than generic "Diff". Uses same `TitleFmt` with an `Extras` field. | S | Parameterize the panel wrapper's title composer. |
| Empty-state copy inside the bordered region ("No weak secrets detected ✓") | Rich empty states are what separate k9s from generic TUIs. The border stays, the content is a centered message. | S | Pure render pass. |
| Title colour slots exposed per skin (`title.fgColor`, `title.highlightColor`, `title.counterColor`, `title.filterColor`) | Mirrors the skin schema exactly (every k9s skin under `skins/*.yaml` has these — e.g. `dracula.yaml:70-75`, `monokai.yaml:91-96`). Lets skin authors re-use k9s colour conventions. | S | Requires Category 8 — skin schema — before it's useful. |

#### Anti-Features

| Feature | Why Requested | Why Problematic | Alternative |
|---|---|---|---|
| Resizable pane borders (drag to resize) | "Modern" | Not keyboard-reachable and not present in k9s. Adds mouse and layout state. | Fixed breakpoints; add a keybind to toggle a pane fullscreen (like k9s `LiveView.toggleFullScreen` at `internal/view/live_view.go:284-303`). |
| Rounded Unicode borders on ASCII-only terminals | Aesthetic | Breaks in PuTTY, low-quality SSH terminals, and CI. k9s relies on `tview.Borders` defaults. | Use lipgloss `RoundedBorder()` where supported; fall back to `NormalBorder()`. Add a skin option (`frame.border.style: normal|rounded|thick`). |
| Per-pane scrollbars rendered in border | "Helpful" | k9s does not do this; it relies on the title counter + row highlight. Scrollbars in terminal are visually noisy and rarely accurate. | Row count + selected row index in title suffix: `Files [12/37]`. |

---

### Category 6 — Breadcrumb Chips (bottom navigation trail)

k9s renders crumbs as pills: ` <segment> ` with background colour from `frame.crumbs.bgColor`, and the **last** crumb (active) uses `activeColor` (`internal/ui/crumbs.go:62-74`):

```go
fmt.Fprintf(c, "[%s:%s:b] <%s> [-:%s:-] ", fgColor, bgColor, segment, bodyBg)
```

Crumbs auto-update on view stack push/pop (`crumbs.go:47-57`). The stack is a `model.Stack` — last-N navigation.

#### Table Stakes

| Feature | Why Expected | Complexity | Depends on |
|---|---|---|---|
| Each crumb has a coloured background (pill/chip), not plain text | This is the literal difference between "crumbs" and "nav path". The background colour separates segments at a glance. `internal/ui/crumbs.go:64-73`. | S | `internal/ui/styles.go` currently has `BreadcrumbActive` + `BreadcrumbSep` only — add `BreadcrumbChip` (surface background, fg accent). |
| Active crumb visually distinct from history | `activeColor` vs `bgColor` (`crumbs.go:65-67`). Without this the breadcrumb reads as static text. | S | `BreadcrumbActive` exists already; confirm it uses **background** colour not just foreground. |
| Crumbs lowercase + strip spaces (matches k9s convention) | `strings.ReplaceAll(strings.ToLower(crumb), " ", "")` (`crumbs.go:70-71`). "Secret Health" → `<secrethealth>`. Keeps chip width tight. | S | Pure string op. |
| 1-cell left/right padding around the chip strip | `SetBorderPadding(0, 0, 1, 1)` (`crumbs.go:32`). Otherwise chips hug the terminal edge. | S | lipgloss `PaddingLeft(1)`. |

#### Differentiators

| Feature | Value Proposition | Complexity | Depends on |
|---|---|---|---|
| Breadcrumb is keyboard-reachable (press `[` to go back, `]` forward — like k9s `cmdHistory`) | Real navigation, not just a label. k9s binds `[`/`]` at `internal/view/app.go:259-260, 730-758`. | M | Need a view history stack in `app.Model` (currently there's only the state enum, not a stack). |
| Crumb for filter state (`files > filter:nginx > detail:app.yaml`) | Exposes current filter as a visible crumb — unambiguous vs a title suffix. | S | Surface filter state as its own crumb segment. |
| Numeric chips for "N files matching" as ephemeral crumbs | Rich information in the trail without an extra panel. | S | Render helper. |

#### Anti-Features

| Feature | Why Requested | Why Problematic | Alternative |
|---|---|---|---|
| Mouse-clickable breadcrumbs | "Intuitive" | Not keyboard-first. k9s explicitly does not wire click handlers — `Crumbs` is a `tview.TextView` with `SetDynamicColors(true)` only (`crumbs.go:30-37`). | Keyboard `[`/`]` history. |
| Full path as breadcrumb (`sops-tui > files > secrets/prod/app.yaml > key: db.password`) | "More context" | Wraps and clips on narrow terminals. k9s keeps crumbs to one-word segments. | Short semantic labels; full path lives in the panel title. |
| Crumbs that persist across sessions | "Resume where I was" | Adds filesystem persistence and confuses users who expect a clean launch. k9s does not persist crumbs. | Fresh stack on every launch; `cmdHistory` is in-memory only (`internal/view/app.go:66`). |

---

### Category 7 — Keyboard Interaction Patterns (command bar, number keys, `:` prompt)

k9s has two distinct prompt modes driven by `model.BufferKind` (`internal/ui/prompt.go:284-298`):

- **`:` command buffer** — icon `🐶`, prefix `>`, border colour `Prompt.Border.CommandColor`. Activated by `:`, used for `pod`, `ns prod`, `ctx prod-cluster`, etc.
- **`/` filter buffer** — icon `🐩`, prefix `/`, border colour `Prompt.Border.DefaultColor`. Activated by `/`, filters the current view.

Both support tab-completion via a `Suggester` interface (`prompt.go:28-40, 184-190`).

Other patterns:
- Number keys `1`..`9` for favourites/view switches (`menu.go:199-206` treats digit mnemonics specially).
- `Esc` deactivates prompt (`prompt.go:164-166`).
- `Enter` commits (`prompt.go:168-170`).
- `Ctrl-W`/`Ctrl-U` wipe input (`prompt.go:172-173`).
- `Tab` accepts current suggestion (`prompt.go:185-189`).
- `Ctrl-A` aliases, `[`/`]` history back/forward, `-` toggles last view (`internal/view/app.go:259-262`).

#### Table Stakes

| Feature | Why Expected | Complexity | Depends on |
|---|---|---|---|
| `/` activates fuzzy search prompt that is currently behind the same key | v1.0 already binds `/` — confirm the *visual* is a bordered input with an icon/prefix, not a naked textinput. `prompt.go:268-282` shows border-on-activate. | S | `internal/ui/search.go` exists; swap render to a bordered textinput. |
| `Esc` cancels prompt without committing | Fundamental text-input muscle memory. `prompt.go:164-166` literally maps `KeyEscape` to `ClearText+SetActive(false)`. | S | Check current `search.go` — must short-circuit `Esc` before `tea.Quit` bubbles up. |
| `Enter` commits + exits prompt mode | `prompt.go:168-170`. | S | Likely already present. |
| Number keys are unused (reserved, not accidentally hijacked by textinput) | Users expect `1`, `2`, `3` etc. to jump views, matching k9s favourites. Even if we don't bind them in v1.1, we must not grab them for text input. | S | Audit existing key handlers. |
| `?` opens full help overlay | Already implemented in v1.0 (`internal/ui/help.go`). Keep as the canonical complete reference. | (no change) | `internal/ui/help.go`. |

#### Differentiators

| Feature | Value Proposition | Complexity | Depends on |
|---|---|---|---|
| `:` command bar for jumping between views (`:files`, `:health`, `:recipients`) | This is **the** k9s UX moment. Combined with alias completion it's the "power user discovered the keyboard" hook. `prompt.go:293-294` uses `🐶` + `>`. | L | Needs new `internal/ui/prompt.go`; new dispatch table in `app.Model.Update` mapping command strings to state transitions; suggester backed by a static alias list initially. |
| Tab-completion on `:` commands | Core part of k9s UX (`prompt.go:185-189`). Without it `:` is just a poor substitute for keybinds. | M | Suggester interface + current-suggestion rendering (`prompt.go:240-242`). |
| Number-key view switching (`1`=files, `2`=health, `3`=recipients) | Fastest path between views; matches k9s favourites pattern. | S | New top-level keybindings in `app.Model.Update`. |
| `[` / `]` history back/forward | Mirrors k9s `previousCommand`/`nextCommand` (`app.go:730-758`). Pairs with breadcrumbs from Category 6. | M | Requires view history stack. |
| `-` toggles last view (cd - style) | `lastCommand` (`app.go:760-773`). Tiny muscle-memory win. | S | Trivial if history stack exists. |
| Command history buffer (up-arrow recalls last `:` command) | `prompt.go:175-183` — up/down arrows traverse suggestions. k9s uses a `model.History` with `MaxHistory` cap (`app.go:66-67`). | M | New in-memory ring buffer. |

#### Anti-Features

| Feature | Why Requested | Why Problematic | Alternative |
|---|---|---|---|
| Copy k9s's emoji icons (🐶, 🐩) for prompt prefixes | "Fun, matches k9s" | Dog/poodle branding is k9s-specific (it's a Kubernetes pun). For a secrets tool it's off-brand and noisy. `prompt.go:286-287` has a `noIcons` flag exactly for this reason. | Use plain `:` and `/` prefixes with no icon, matching `noIcons=true` path. |
| Global `:delete` / destructive commands via `:` bar | k9s has these for pods | Secrets are **not** recoverable after a bad `sops -d` redirect or key removal. Every destructive action should require a confirm diff (already done in v1.0). | Scope `:` to navigation-only: `:files`, `:help`, `:q`, `:back`. Destructive actions stay keybind + confirm. |
| Vi-style composable commands (`dd`, `yy`, `gg` — more than single-key) | "Vim muscle memory" | Single-key bindings already cover our surface; composable commands demand a full parser and modal state machine. Way out of scope. | Stick to single-key vim-flavoured bindings (h/j/k/l, g/G, `/`). |
| Aliases persisted to disk | k9s has an alias config | Adds a config-file-editing escape hatch that can't be validated at load time. Secrets tooling doesn't benefit. | Static aliases shipped in the binary. |

---

### Category 8 — Theming / Skin Support

k9s skins are YAML files under `$XDG_CONFIG_HOME/k9s/skins/` (user dir) or `skins/` (builtin, 35 shipped). Schema is flat under the `k9s:` root, grouped into: `body`, `prompt`, `info`, `help`, `dialog`, `frame.{border,menu,crumbs,status,title}`, `views.{charts,table,xray,yaml,logs}` (verified across `dracula.yaml`, `gruvbox-dark.yaml`, `monokai.yaml`).

YAML anchors are the standard pattern: define a palette at the top with `&name`, reference with `*name`. Every value is a hex color or `default` (terminal default) or the string `transparent`.

Styles are **live-reloadable**: `SkinsDirWatcher` (`app.go:349-354`) triggers `ReloadStyles` without restart when `K9s.UI.Reactive` is true. Every visual component registers itself as a listener: `styles.AddListener(c)` (in `NewMenu`, `NewLogo`, `NewCrumbs`, ...). `StylesChanged(s)` method on each component re-applies colours.

#### Table Stakes

| Feature | Why Expected | Complexity | Depends on |
|---|---|---|---|
| Default theme tuned to k9s visual conventions | Users coming from k9s expect pink/purple/cyan accents, dark bg. Current sops-tui palette (`internal/ui/styles.go:14-32`) is Catppuccin Mocha — close enough, tune `ColorAccent` (currently `#89b4fa`) to match k9s's typical "hot pink" highlight colour. | S | Pure palette tune. |
| Skin colour slots match k9s groupings (body, prompt, info, frame.{menu,crumbs,title,border,status}) | Anything else will confuse users porting k9s skins. | M | Refactor `internal/ui/styles.go` from flat constants into a nested `Styles` struct with k9s-compatible groupings. |
| Logo colour states named (`logoColorInfo`, `logoColorWarn`, `logoColorError`) | k9s skins like `monokai.yaml:26-30` expose these explicitly. Hard-coding them hides customization. | S | Add three colour slots under `body`. |

#### Differentiators

| Feature | Value Proposition | Complexity | Depends on |
|---|---|---|---|
| YAML skin file support with identical schema to k9s | Instant import path: users drop their k9s skin into sops-tui and it works. **Massive** discoverability win. | L | New `internal/ui/config/` package: YAML decoder (`goccy/go-yaml` already in `STACK.md`), `Styles` struct matching k9s grouping, load from `$XDG_CONFIG_HOME/sops-tui/skins/<name>.yaml`. |
| Ship 2-3 builtin skins (default, dracula, gruvbox) | No-effort customization for users who don't want to write YAML. k9s ships 35. | S | Copy skin YAMLs verbatim, subset to our actually-used colour slots. |
| Skin selection via `:skin <name>` command | Interactive reload without editing config. Pairs with Category 7 command bar. | M | Needs command bar first. |
| Live skin reload when file changes | k9s's killer demo: edit a skin, save, terminal updates instantly (`SkinsDirWatcher` at `app.go:351-354`). | L | `fsnotify` watcher on skin dir; route a `SkinChangedMsg` into `app.Model`. Real high-polish feature — strong candidate for **v1.2**. |

#### Anti-Features

| Feature | Why Requested | Why Problematic | Alternative |
|---|---|---|---|
| Adaptive colour based on terminal light/dark detection | "Accessible" | **Confirmed hang in lipgloss** — see `internal/ui/styles.go:5` comment: `Do NOT use lipgloss.AdaptiveColor — confirmed hang (issue #1036).` k9s also avoids this by shipping separate light+dark skin files (`gruvbox-light.yaml` vs `gruvbox-dark.yaml`). | Ship separate light/dark skins; user picks one. |
| In-TUI skin editor | "Easy customization" | k9s doesn't have one — the YAML file is the editor. Building one is multi-week work for marginal value. | Hot-reload on file save (differentiator above). |
| Per-view skins (different theme for Health vs Files) | "Expressive" | Violates UX consistency; makes the app look schizophrenic. k9s applies one skin globally. | One skin per session. |
| Inline colour overrides via env vars (`SOPS_TUI_COLOR_BG=...`) | "Dockerfile-friendly" | 50+ env vars to cover the schema. Bad substitution for a skin file. | `SOPS_TUI_SKIN=dracula` env var pointing at a skin name. |

---

### Category 9 — Anti-Features from k9s That Do NOT Translate

This is the "don't copy k9s reflexively" section. Items below are things k9s has that would be actively wrong for a secrets tool. Enumerated so they are rejected explicitly during roadmap planning rather than accidentally spec'd later.

| k9s Feature | Why k9s Has It | Why NOT for sops-tui |
|---|---|---|
| `clusterUpdater` goroutine polling every 15s (`app.go:361-392`) | K8s API is authoritative and external; resources mutate outside k9s. | Our data source is the filesystem + age keyring. Nothing mutates without user action. A poller is pure waste; it burns battery on laptops with no upside. |
| Exponential backoff on connection failure with eventual `BailOut` (`app.go:372, 416-424`) | K8s clusters become unreachable mid-session. | Files don't "go unreachable". Missing file = user-visible error, not a resumable failure. No backoff needed. |
| Benchmarking UI (`IsBenchmarking`, `internal/view/benchmark.go`) | k9s can hit pod endpoints with load. | Decrypting in a loop for benchmarking would leak plaintexts to history and offer zero value. Explicitly out of scope per `PROJECT.md`. |
| Log forwarding / streaming (`internal/view/log.go`) | Pods produce logs. | Secrets don't produce logs. Our existing `tea_debug.log` is for our own debugging, not a feature. |
| Image vulnerability scanner (`internal/vul/`, `initImgScanner` at `app.go:157-164`) | k9s scans container images. | We're not a package scanner. |
| Port-forward management | K8s-specific concept. | No equivalent. |
| Alias file loading + hot reload (`command.Reset` at `app.go:430-438`) | Many alias-able resource kinds (pods, svc, deploy, ds, sts, ...). | We have ~4 primary views. Static in-binary dispatch is fine. |
| Cow (error dialog) with "drain this node" severity levels | K8s operations are destructive at scale and need multi-button dialogs. | v1.0 already has a simpler error box (`internal/ui/errorbox.go`); do not escalate to k9s's `internal/ui/dialog/` complexity. |
| Context switcher (`switchContext` at `app.go:459-522`) | Multiple K8s clusters. | One `.sops.yaml` per directory. Already handled at startup. |
| ReadOnly indicator (`ui.ROIndicator` at `cluster_info.go:119`) | k9s can run in read-only mode globally. | We already have per-file read-only semantics (uncommitted-status badges from git integration). Don't add a global flag. |
| Namespace selector as first-class header element | Every k8s resource is namespaced. | `.sops.yaml` defines one set of creation rules globally. No analogue. |
| Splash page revision banner with update-check against GitHub | k9s polls GitHub for latest release. | Phoning home from a secrets tool is a trust problem. The user installed sops-tui from a package manager; they know what version they have. |

---

## Feature Dependencies

```
Category 1 (Header Region)
    ├── enables── Category 2 (Info Panel)            [info panel lives IN the header]
    ├── enables── Category 3 (Logo)                  [logo is header's right column]
    └── enables── Category 4 (Menu)                  [menu is header's center column]

Category 4 (Menu)
    └── requires── HelpEntry refactor in internal/ui/help.go   [so each view owns its hints]

Category 5 (Titled Borders)
    └── independent                                  [can ship before or after header]

Category 6 (Breadcrumb Chips)
    ├── enhances── existing bottom status bar
    └── enhanced by── view history stack (Category 7)

Category 7 (Command Bar + Number Keys)
    ├── enables── `:skin <name>` reload in Category 8
    ├── enables── `[` / `]` in Category 6
    └── requires── state transition audit in app.Model

Category 8 (Skin Schema)
    ├── requires── Styles struct refactor (flat constants → nested groups)
    ├── enables── Category 5's skinnable title slots
    └── conflicts with── hardcoded hex in internal/ui/styles.go
```

### Dependency Notes

- **Category 1 blocks 2, 3, 4:** Header layout is the carrier for info panel, logo, and menu. Ship the header skeleton first (even with placeholder children) before filling each column.
- **Category 4 requires HelpEntry refactor:** Current `internal/ui/help.go` holds a flat list of all bindings. To hydrate the menu per view, each view must own its hints (see `internal/view/log.go:199-201` pattern).
- **Category 5 is independent:** Titled borders can ship before the header region is built. This makes it a good Phase 1 win — visible progress, low risk.
- **Category 7 unlocks Category 8 polish:** `:skin <name>` command is a differentiator, not table stakes. Ship skins (Category 8) as files first, then wire the command-bar trigger.
- **Category 8 conflicts with current `internal/ui/styles.go`:** That file is flat `var Color* = lipgloss.Color(...)`. Skin support requires a nested `Styles` struct with a loader. Doing one without the other makes customization impossible. Handle as one big refactor commit.

---

## MVP Definition (v1.1 scope)

### Launch With (v1.1)

The visual-parity milestone. Everything below has to ship together or the app doesn't look like k9s.

- [ ] **Header region** (Category 1 table stakes) — essential layout change; every other category depends on it
- [ ] **Info panel** (Category 2 table stakes) — `.sops.yaml`, Repo, Files, Recipients, Age key, Sops version in 2-column table
- [ ] **ASCII logo with status coupling** (Category 3 table stakes) — 6-row art, status band below, colour changes on info/warn/error
- [ ] **Persistent keybinding menu** (Category 4 table stakes) — 6-row grid, re-hydrates per view, `?` retained as full overlay
- [ ] **Titled bordered content regions** (Category 5 table stakes) — title + count + filter suffix + focus colour
- [ ] **Breadcrumb chips** (Category 6 table stakes) — pill backgrounds, active highlight
- [ ] **`Esc` cancels all prompts correctly** (Category 7 table stakes) — audit existing `search.go` before new prompt work
- [ ] **Palette tuned to k9s conventions** (Category 8 table stakes) — currently Catppuccin, adjust accent to match k9s pink/purple norm
- [ ] **Logo status wired to health aggregate** (Category 3 diff) — worth including: reuses existing `internal/health/` data, low risk, huge UX payoff

### Add After Validation (v1.2)

Features that elevate the shell but aren't required to "look like k9s".

- [ ] **`:` command bar + number-key view switching** (Category 7 diff) — trigger: users report "how do I jump to health faster?"
- [ ] **`[` / `]` view history + breadcrumb is navigable** (Category 6 + 7 diff) — trigger: after command bar ships, history becomes cheap
- [ ] **Skin YAML file support with 2-3 builtin skins** (Category 8 diff) — trigger: after the Styles struct refactor is stable, layer skin loading on top
- [ ] **Contextual hint visibility per state** (Category 4 diff) — trigger: first UX feedback that menu feels generic

### Future Consideration (v1.3+)

Polish that's nice but not k9s-table-stakes.

- [ ] **Live skin reload via fsnotify** (Category 8 diff) — trigger: after skin support is in, add the watcher
- [ ] **Splash screen with big ASCII logo + 1s delay** (Category 3 diff) — nostalgic k9s touch; zero user-value
- [ ] **Responsive column widths / narrow-terminal handling** (Category 1 diff) — trigger: first narrow-terminal bug report
- [ ] **`:skin <name>` runtime switcher** (Category 8 diff) — trigger: after command bar + skin loading both land

---

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---|---|---|---|
| Header region skeleton | HIGH | MEDIUM (touches every view's size assumptions) | P1 |
| Info panel | HIGH | MEDIUM (data plumbing trivial, layout work non-trivial) | P1 |
| ASCII logo + status coupling | HIGH | MEDIUM (new component, but small) | P1 |
| Persistent keybinding menu | HIGH | MEDIUM (requires hint refactor) | P1 |
| Titled content borders | HIGH | MEDIUM (wrap every view) | P1 |
| Breadcrumb chips | MEDIUM | LOW (existing breadcrumb, restyle) | P1 |
| `Esc` audit on prompts | HIGH (blocks user trust) | LOW | P1 |
| Palette tune | MEDIUM | LOW | P1 |
| Logo → health severity aggregate | HIGH | LOW (data already exists) | P1 |
| `:` command bar | MEDIUM | HIGH | P2 |
| Number-key view switching | MEDIUM | LOW | P2 |
| `[` / `]` history | LOW | MEDIUM | P2 |
| Styles struct refactor | MEDIUM (enables skins) | HIGH | P2 |
| Skin YAML loader | MEDIUM | HIGH | P2 |
| Builtin skins (dracula, gruvbox) | MEDIUM | LOW (once loader exists) | P2 |
| Tab-completion on `:` | MEDIUM | MEDIUM | P2 |
| Splash screen | LOW | LOW | P3 |
| Live skin reload (fsnotify) | LOW | MEDIUM | P3 |
| Responsive header | LOW (until reported) | MEDIUM | P3 |
| `:skin <name>` runtime switch | LOW | LOW | P3 |

**Priority key:**
- **P1** — must ship in v1.1 for the milestone to be meaningful
- **P2** — should ship in v1.2; material UX upgrade
- **P3** — nice to have; defer

---

## Competitor Feature Analysis

k9s is **the** reference. No other product is a meaningful competitor for the visual shell — lazygit has a different visual grammar (no persistent hint menu, no info panel), gpg-tui is single-screen, vaul7y is sparse by design. Comparing to anything other than k9s produces noise.

| Visual Element | k9s (reference) | lazygit | Our v1.1 Plan |
|---|---|---|---|
| Persistent hotkey menu | Multi-column grid, rehydrates per view | Bottom-row dense text line | Copy k9s: multi-column grid, per-view hydration |
| Header info panel | 2-col cluster info table | None (branch name in title) | Copy k9s structure; sops-specific rows |
| Logo w/ status coupling | 6-row ASCII + coloured status band | None | Copy k9s; sops-tui ASCII art |
| Titled bordered regions | Every pane, with count + filter | Every pane, no filter suffix | Copy k9s (count + filter + focus colour) |
| Breadcrumb chips | `<segment>` pills | Plain text status line | Copy k9s exactly |
| `:` command bar | Animal icon + `>` prefix + border | No `:`; single-key commands | Copy k9s minus animal icon; `:` + `>` only |
| Skin YAML | Full schema, 35 shipped, hot reload | Static theme config | Subset of k9s schema; ship 2-3 skins; hot reload deferred |

---

## Sources

- `~/git/k9s/internal/ui/menu.go` — `HydrateMenu`, `formatNSMenu`, `formatPlainMenu`
- `~/git/k9s/internal/ui/logo.go` — `Info`/`Warn`/`Err` status routing, logo refresh
- `~/git/k9s/internal/ui/splash.go` — `LogoSmall` (6 rows), `LogoBig` (splash)
- `~/git/k9s/internal/ui/crumbs.go` — chip rendering, `StackPushed`/`StackPopped`
- `~/git/k9s/internal/ui/prompt.go` — `:` and `/` buffers, `Suggester`, icon flag
- `~/git/k9s/internal/ui/table_helper.go` — `TitleFmt`, `NSTitleFmt`, `SearchFmt`, `SkinTitle`
- `~/git/k9s/internal/ui/table.go` — `UpdateTitle`, `styleTitle`
- `~/git/k9s/internal/ui/action.go` — `KeyAction`, `KeyActions.Hints()`
- `~/git/k9s/internal/view/cluster_info.go` — info-panel layout, 2-col table, `ClusterInfoChanged`, `warnCell`
- `~/git/k9s/internal/view/app.go` — header composition (lines 166-183, 273-331), status→logo routing (600-611), `Ctrl-E`/`Ctrl-G` toggles (635-659), command history (730-773), cluster polling (361-392)
- `~/git/k9s/internal/model/menu_hint.go` — `MenuHint{Mnemonic,Description,Visible}`, digit-first sort
- `~/git/k9s/skins/dracula.yaml`, `skins/gruvbox-dark.yaml`, `skins/monokai.yaml` — skin schema verified across three styles
- `~/git/k9s/internal/view/{log,live_view,details}.go` — `Hints() model.MenuHints` implementations (per-view hint ownership pattern)
- `~/git/sops-tui/internal/ui/styles.go` — current palette (Catppuccin Mocha); notes `lipgloss.AdaptiveColor` hang (issue #1036)
- `~/git/sops-tui/internal/ui/help.go` — current full-screen help overlay to be retained
- `~/git/sops-tui/.planning/PROJECT.md` — validated v1.0 features that must not regress; Out-of-Scope list (informs anti-features in Category 9)

---
*Feature research for: k9s-style visual shell layered on existing sops-tui v1.0*
*Researched: 2026-04-23*
