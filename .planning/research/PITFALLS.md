# Pitfalls Research — v1.1 k9s Visual Parity

**Domain:** Bubble Tea v2 TUI — adding persistent k9s-style visual shell to existing full-height app
**Researched:** 2026-04-23
**Confidence:** HIGH (verified against existing codebase, Bubble Tea v2 source, k9s source)

Scope note: v1.0 pitfalls (alt-screen scrollback, blocking I/O in Update, subprocess argv leakage, clipboard persistence, MAC invalidation, AdaptiveColor hang, WindowSizeMsg propagation, etc.) are archived in `.planning/research/v1.0-archive/PITFALLS.md` and explicitly **not re-catalogued** here. This file focuses only on pitfalls introduced by — or made newly dangerous by — the v1.1 chrome rework.

---

## Critical Pitfalls

### Pitfall 1: Chrome Height Subtraction Only At Top-Level; Child Widgets Still Assume Full Height

**What goes wrong:**
`AppModel.View()` subtracts `statusBarHeight(m)` from `m.height` today (model.go:1333). When the chrome (header + optional crumbs row + titled-border) is added, the top-level subtraction will be updated — but every child widget that cached its own `height` via `SetSize(width, mainH)` during `WindowSizeMsg` (model.go:313–328) now has a stale value one chrome-height too tall. Result: `FileListModel.list` has `height = screen - statusBar`, not `screen - statusBar - header - borders`. The embedded `bubbles/v2/list` paginates to `height/2` for PgDn (filelist.go:275, 282) and `list.SetSize(width, height)` at filelist.go:346. The list scrolls past rows that get painted *under* the header or status bar, and the selection indicator can land on rows that are visually clipped.

**Why it happens:**
Height is propagated through multiple SetSize calls from eight different sites in model.go (lines 321–328, 349, 353, 377, 385, 489, 502, 509, 567, 574, 631, 635, 724, 728, 761, 765). When chrome is added, only the top-level View arithmetic is obvious; the propagation sites are easy to miss because they live inside `case` blocks scattered across a 400-line Update method. teatest snapshots taken at 80×24 will look fine because the visible region clips neatly; the bug manifests at 120×40 when the hidden rows below the status bar start receiving content.

**Consequences:**
- Selection cursor can leave the visible area on large terminals
- PgDn jumps past the visible window (selected row scrolls off-screen into hidden region)
- Bordered title frame's bottom edge is clipped or overwritten by status bar
- Resize to shorter terminal crashes if any child uses a negative height from the delayed propagation

**How to avoid:**
1. Introduce a `chromeHeight(m AppModel) int` helper beside the existing `statusBarHeight(m AppModel) int` (model.go:1443). Single source of truth.
2. Introduce a `bodyHeight(m AppModel) int` helper that returns `m.height - chromeHeight(m) - statusBarHeight(m)`, clamped ≥ 0. Every `m.height - statusBarHeight(m)` expression in model.go becomes `bodyHeight(m)`.
3. Write a compile-time guard: grep-gate in CI that fails the build if any new `m.height - statusBar` arithmetic appears outside the `bodyHeight` helper.
4. Add a teatest case at 40×12, 80×24, and 200×60 that asserts the rendered frame's last-line characters are status-bar characters, not content, and the line above is body content (not header).

**Warning signs:**
- A `SetSize` call uses `m.height - statusBarHeight(m)` without passing through the new helper
- `lipgloss.Height(frame) != m.height` in any teatest snapshot
- Resize tests pass at 80×24 but fail at 80×12 (chrome + status bar exceeds body budget)

**Phase to address:** FIRST phase of v1.1 — **layout-arithmetic groundwork must land before any chrome rendering is merged.** If the helper goes in after the header, every SetSize site needs a second audit pass.

---

### Pitfall 2: Chrome Renders on Every View() Call; Logo / Menu Rebuild Cost Amortizes Into Input Latency

**What goes wrong:**
Bubble Tea v2's `View()` returns a `tea.View` struct and is called on **every** Update — including every keystroke and every Tick. The k9s logo is a 6-line ASCII block, the menu is a 6-row × N-column table, and the cluster-info-equivalent is a multi-row key:value pane. If these are built string-by-string inside View (looping over `LogoSmall`, hydrating menu hints from current state, formatting git status and age fingerprint), the render allocates dozens of KB per keystroke. The Cursed Renderer is ~30% faster than the previous renderer but still diffs the full frame. Under hold-down-j scrolling with a 240 fps keyboard, allocations pile up and GC latency spikes visible as input stutter.

**Why it happens:**
The tview-based k9s keeps `Logo`, `Menu`, `ClusterInfo` as persistent widget trees that only refresh on state change (see k9s `ui/logo.go:refreshLogo` called from `StylesChanged`). In Bubble Tea v2 immediate-mode, there is no persistent widget — the naive port writes a pure function `renderLogo(state LogoState) string` and calls it every frame. Looks identical to tview output at 1 fps, falls apart at 240 fps.

**Consequences:**
- Noticeable input lag on long hold-down-j scrolling
- Cursed Renderer's frame diff fails to find stable regions because the logo string is re-rendered with fresh `lipgloss.NewStyle()` allocations
- `go test -bench` for the model shows steady allocations per keystroke that grow with chrome complexity

**How to avoid:**
1. Cache the rendered logo/menu strings on the model when the inputs change, not on every View call. Pattern:
   ```go
   type AppModel struct {
       // ...
       headerCache   string // rendered header; invalidated on state change
       headerCacheKey headerCacheKey // {width, state, env, clipboardHot, flashGen}
   }
   func (m AppModel) header() string {
       key := m.headerCacheKey()
       if m.cachedHeaderKey == key { return m.headerCache }
       // rebuild
   }
   ```
   But because View is on a value receiver, cache mutation has to happen via Update returning a new model (not inside View). Simpler alternative: compute once in `WindowSizeMsg` handler and on every explicit state transition, store in model, read in View.
2. Build the logo/menu/cluster-info style tables once in `NewAppModel` and reuse, not per-render. lipgloss styles are values and safe to share.
3. Benchmark the View with `testing.B`: target ≤ 50 µs per View call at 200×60.
4. Never call `lipgloss.NewStyle()` inside View. All styles live in `internal/ui/styles.go` as package-level vars — this rule extends to the new menu/logo/cluster-info styles.

**Warning signs:**
- `go test -bench=BenchmarkView` shows > 100 µs per iteration, or > 20 KB/op allocs
- Visible stutter on held-down `j` in a 200+ row file list
- `pprof` shows `lipgloss.NewStyle` in the top 10 allocators during normal interaction

**Phase to address:** Chrome rendering phase. Add benchmark before merging the logo/menu.

---

### Pitfall 3: Menu Hints Derived From Wrong State During Modal Overlays

**What goes wrong:**
The k9s `Menu` widget is hydrated from `Component.Hints()` on `StackPushed` / `StackPopped` (menu.go:60–71). sops-tui does not have a component stack — modal state is a flat `sessionState` enum with 14 values (stateFileList through stateBulkReKeyConfirm). If the menu hint source uses `m.state`, the hints show `stateDiff` keybindings while the user is in `stateRecipientConfirm` because both reuse `m.diff.View()` (model.go:1360, 1364). Conversely, if the menu hint source keys off *m.prevState*, it shows file-list hints while the user is looking at the diff.

**Why it happens:**
Ten of the 14 states render via shared sub-models (`m.diff` for stateDiff/stateRecipientConfirm/stateBulkReKeyConfirm; `m.help` for stateHelp; metadata/history use their own models). The `recipientAction` field ("addrecipient" / "removerecipient" / "bulkrekey" / "healthcheck") disambiguates what the diff represents — this flag drives the correct hints, not `m.state` alone. Missing this distinction yields wrong hint text even though the diff content is correct.

**Consequences:**
- Menu suggests `[y]Confirm [n]Cancel [Esc]Back` when the user is in a nested form that expects `[Tab]Next [Enter]Submit [Esc]Cancel`
- User tries a key listed in the menu, nothing happens, trust in the chrome erodes
- Worst case: key *does* something — user follows a stale hint and triggers the wrong state transition

**How to avoid:**
1. Define menu hints as a pure function of `(sessionState, recipientAction, IsSearchActive)` — not just sessionState. Co-locate the hint set with each keymap in `internal/keys/`.
   ```go
   func MenuHintsFor(state sessionState, action string, searchActive bool) []MenuHint {
       switch state {
       case stateDiff:
           switch action {
           case "addrecipient": return RecipientConfirmHints
           case "bulkrekey": return BulkReKeyHints
           case "healthcheck": return HealthScanConfirmHints
           default: return DiffHints
           }
       // ...
       }
   }
   ```
2. Add a golden-file teatest for each `(state, action)` tuple that asserts the rendered menu matches the keymap's actual `key.Matches` bindings. The test walks each menu hint and verifies the same key triggers a handler in the corresponding Update branch — drift between menu and keymap becomes a compile-time-ish failure.
3. When search is active inside file list (model.go `IsSearchActive()` — filelist.go), the menu must show `[Esc]ExitSearch [Enter]Select` NOT the normal file-list hints. Add `search-active` as an axis.
4. Single-source-of-truth contract: **the keymap defines the menu**, never free-form strings in the menu. If a key binding changes, the menu updates mechanically.

**Warning signs:**
- Hint text as a string literal adjacent to the menu-render function instead of imported from a keymap
- Any `if m.state == stateX || m.state == stateY` branch in the hint builder — that's where `recipientAction` was forgotten
- Users in beta reporting "the menu said X but X does nothing"

**Phase to address:** Chrome state-integration phase, after the basic menu render works at a single state. The test harness from pitfall #10 (golden files) must exist first so the `(state, action)` matrix is enforceable.

---

### Pitfall 4: Skin Loading Fails Closed, Locks User Out of TUI

**What goes wrong:**
Adding user-configurable skins (YAML file with color hex values) introduces a new failure mode: the skin loader parses a bad hex string ("#gggggg", "notahex"), returns an error, and the TUI refuses to start. Worse, if the skin is loaded lazily inside View (bad pattern but tempting for hot-reload), a mid-session parse failure panics inside `lipgloss.Color()` which silently treats a malformed string as an empty color, producing invisible text on the matching background.

**Why it happens:**
k9s has `SkinsDirWatcher` (view/app.go:352) and hot-reloads skins mid-session — this is a feature. Porting that feature naively to Bubble Tea v2 puts skin-parse errors on the hot path. The lipgloss v2 color parser is lenient with malformed strings; it logs nothing and falls back to an empty color. On a dark background, `Foreground(lipgloss.Color(""))` is the same as `Foreground(ColorBg)` — text becomes black-on-black and invisible.

**Consequences:**
- User edits ~/.config/sops-tui/skin.yaml, hits save, TUI window becomes unreadable with no error shown
- Bad skin shipped in a dotfiles repo means the TUI is broken for everyone who clones it
- If skin errors are fatal at startup, a corrupt ~/.config file makes the TUI unlaunchable — user loses access to a working binary for a cosmetic issue

**How to avoid:**
1. **Fail-open at startup.** If skin parse fails, log a warning (via the existing stderr pattern for startup warnings), render the warning in the status bar flash on first frame, and fall back to the default palette defined in styles.go. Never panic, never refuse to launch.
2. **Validate hex strings eagerly, not at render time.** When loading a skin, run `validateHex(s string) error` for every color field. Return a clear error like `skin: field "header.bg" is not a valid hex color: "#gg1234"` with the field name. Surface via status flash and keep defaults.
3. **No hot-reload in v1.1.** Skins load once at startup. Defer SkinsDirWatcher to a later milestone — it adds a goroutine + fsnotify dependency + mid-session color-swap complexity that is not required to match k9s visual parity.
4. **Unit test with a corpus of bad inputs.** Fuzz the skin loader with empty strings, non-hex chars, 2-digit hex, 7-digit hex, RGB decimal, named colors. Confirm each yields a validation error and falls back to defaults.
5. **Explicit color-parse regression test.** Assert `lipgloss.Color("").Dark()` and `lipgloss.Color("").Foreground()` values in a test — if the lipgloss lenient-parse behavior changes in a future version, the test catches it before it reaches users.

**Warning signs:**
- Any `lipgloss.Color(userInput)` call without prior validation
- Skin fields typed as `string` in the config struct instead of a custom `type HexColor string` with a Validate() method
- Absence of fail-open test ("startup with corrupt skin.yaml → TUI launches with warning")

**Phase to address:** Skin/theming phase. Loader and validation both land in the same phase; hot-reload is out of scope for v1.1.

---

### Pitfall 5: Color-Profile Downsampling Makes Logo and Chips Invisible on 16-Color Terminals

**What goes wrong:**
The existing palette (styles.go:14–31) is 24-bit true color: `#89b4fa` accent, `#cdd6f4` foreground, `#313244` surface. lipgloss v2 auto-downsamples to the terminal's detected profile. On a 16-color terminal (common in SSH to old servers, `TERM=xterm`, tmux without `-2` or with broken terminfo), `#89b4fa` (light blue) and `#cdd6f4` (light grey) both map to ANSI `white` (15). Two colors → one post-downsample → **the chip's foreground and background become the same color**. A breadcrumb chip with `bg=accent fg=foreground` renders as a solid unreadable block. The logo, which relies on a single accent color against surface background, becomes either invisible (same color) or shown with scrambled nearest-neighbor mappings.

**Why it happens:**
sops-tui's v1.0 usage was mostly foreground-only (text in accent on default background), which downsamples safely — even if `accent` maps to white, white-on-black is readable. The chrome rework introduces *paired* colors (chip bg+fg, titled-border line+bg, selected-menu-item bg+fg) where both members of the pair must downsample to distinct ANSI colors. The palette was picked for visual harmony in 24-bit, not for downsample distance.

**Consequences:**
- User on CI runs teatest and sees a passing test because teatest forces `NoColor` profile (per CLAUDE.md "teatest uses lipgloss.NoColor profile in tests") — but same user in production SSH session sees broken chips
- Colorblind users configured with 16-color palettes see nothing where a breadcrumb should be
- Apple Terminal (stuck on xterm-256color by default, but many users override to xterm) renders chips unreadably

**How to avoid:**
1. **Terminal capability detection at startup.** lipgloss v2 exposes the detected color profile. If profile is `termenv.Ascii` or `termenv.ANSI` (4-bit), swap to a **16-color-safe fallback palette** defined in styles.go. Keep the 24-bit palette for `termenv.TrueColor` and `termenv.ANSI256`.
2. **Chip rendering must not rely on bg/fg contrast alone.** Add a brackets fallback:
   - 24-bit: `⬤ prod ⬤` with bg accent
   - 16-color: `< prod >` plain text with underline for the active chip
   Select the rendering path from the detected profile at startup, not per-frame.
3. **Test matrix.** Run teatest with forced profile `termenv.Ascii`, `termenv.ANSI`, `termenv.ANSI256`, `termenv.TrueColor`. Four golden files per view. Assert the ANSI and Ascii variants are *readable* (contains the chip text; no empty-width or zero-char output).
4. **Never use `adaptive`-style pairing that assumes distinct-after-downsample.** For each color pair in the palette, hand-verify the 16-color ANSI indices: `AccentForeground != AccentBackground` under the `ANSI` profile.
5. Document the supported TERM values in README. If user is on a bona fide monochrome terminal, chrome degrades to bracket/underline variants, not invisible chips.

**Warning signs:**
- Any teatest assertion that only runs under `lipgloss.NoColor` (teatest default) — chrome tests must explicitly enumerate profiles
- Chip/border styles that set both `Background()` and `Foreground()` from the 24-bit palette without a profile-aware dispatch
- User report "the menu is empty on my server but works on my laptop"

**Phase to address:** Chrome rendering phase (before titled borders), same phase that introduces colored chips. Profile detection must be wired before the palette is used for paired bg/fg.

---

### Pitfall 6: Unicode Width Miscalculation — Borders Break on macOS Terminal, Emoji Chips Overflow on tmux

**What goes wrong:**
Three specific width-calculation traps compound in the chrome:
- **Box-drawing borders** (`│`, `─`, `┌`, `┘`) are `runewidth == 1` on all sane terminals but `Height` and `Width` calls in lipgloss use `uniseg.StringWidth` which works — *unless* the border character is followed by a zero-width joiner or VS16 emoji variation selector, which can happen if the chrome title is a user-supplied string (a rotated recipient name, `.sops.yaml` path containing a non-ASCII folder).
- **Emoji in chips** (`🔓`, `⚠`, `✓` already used in statusbar.go:197–228): 🔓 is `runewidth == 2` on modern terminals but `1` on macOS Terminal.app and old xterm. A chip with `[🔓 prod]` budgeted at 12 cells overflows to 13 on macOS Terminal, pushing the status bar one character right of the edge; next frame, lipgloss truncates and the chip becomes `[🔓 pro]` mid-session.
- **Fullwidth chars** in branch names or file paths (CJK characters, emoji flags) break tmux pane widths that are odd-numbered: a 2-cell character at cell 1 of a 3-cell pane leaves 1 cell of garbage.

**Why it happens:**
`lipgloss.Width()` is accurate for most text but relies on terminfo-provided width tables. tmux terminfo frequently lags upstream emoji width tables by 2-3 years. macOS Terminal is notorious for treating most emoji as `1` cell wide, ignoring VS16. Developers test on Alacritty (gets it right) and ship bugs to macOS Terminal users.

**Consequences:**
- Border `│` at right edge appears one column past the frame on macOS Terminal, creating a visible "broken corner"
- Emoji chip widths drift between frames as content changes
- Tmux pane resize corrupts the chrome until re-render
- teatest golden files pass because they force a known width profile, users see broken frames

**How to avoid:**
1. **No emoji in the persistent chrome for v1.1.** Use ASCII letters + color instead: `[I]` info, `[W]` warn, `[E]` error (colored). Emoji stays in the status bar flash messages where a 1-column drift is cosmetic.
2. **Box-drawing characters:** restrict to the safe subset `│ ─ ┌ ┐ └ ┘ ├ ┤ ┬ ┴ ┼`. These are width-1 on every terminal in our support matrix (verified in the ecosystem research STACK.md).
3. **User-supplied strings (file names, recipient emails, paths) MUST pass through `runewidth.Truncate(s, n, "…")` before being embedded in the chrome.** Use `github.com/mattn/go-runewidth` which k9s relies on (see k9s menu.go:17, 188). Add it to go.mod.
4. **Defensive re-measure:** after composing the chrome string, assert `lipgloss.Width(chrome) <= m.width`. If false (emoji-width miscalculation snuck in), truncate with ellipsis. Never let chrome overflow — overflow is what wraps the layout.
5. **Snapshot tests on multiple TERM environments.** Golden file per `(TERM, profile)` combo: `xterm-256color`, `tmux-256color`, `screen-256color`, `alacritty`, `xterm-kitty`. Use teatest's `lipgloss.ForceColorProfile` + env var to simulate.

**Warning signs:**
- Any rune with code point > U+2B00 in the chrome templates
- Width calc that uses `len(s)` or `utf8.RuneCountInString(s)` instead of `runewidth.StringWidth(s)` / `lipgloss.Width(s)`
- Absence of a macOS Terminal test run in the release checklist

**Phase to address:** Chrome layout primitive phase (titled border rendering). Emoji-free requirement applies to every new chrome render site from day one.

---

### Pitfall 7: Focus Indicators Suggest Interactivity In a Single-Column App

**What goes wrong:**
k9s has multiple focusable panes (resource list, log view, shell). The "focus ring" — a differently-colored border around the active pane — is a core affordance. Porting this visual language to sops-tui creates a problem: sops-tui has one primary content pane at a time (file list OR detail OR diff; modal overlays are full-screen). If the title bar renders a focus ring (accent color) around the Files pane, the user infers there's another pane to Tab to. There isn't. Tab does nothing, or worse, toggles something unrelated. Users waste time looking for hidden navigation.

**Why it happens:**
Visual vocabulary carries implicit promises. A focus ring is a convention: "Tab moves focus." Rendering a focus ring without implementing focus-cycling breaks that convention.

**Consequences:**
- User presses Tab, nothing visible happens (if unbound) or an unexpected state transitions (Tab currently unbound in the keymap — confirmed by inspecting keys.go — but a future contributor may bind it to something else based on seeing the focus ring)
- Support question: "how do I switch to the other pane?"
- Inconsistent with the always-visible-help principle: the menu shows no Tab binding yet the visual says there should be one

**How to avoid:**
1. **No focus ring in v1.1.** The titled border uses a single consistent color (accent) whether or not content is in focus, because there is always exactly one pane.
2. **When modal overlays appear (diff, form, recipient list), dim the content below with `.Faint(true)` but don't change the border.** The overlay has its own title and sits on top. User's eye follows the prompt, not the focus ring.
3. **If multi-pane is added in a future milestone**, introduce focus rings at that point, alongside Tab cycling. Not before.
4. **Menu hint content is the focus indicator.** The currently-active state's hints are shown; that tells the user "these are the keys that work right now." No extra visual ring needed.

**Warning signs:**
- A `BorderColor` style parameter named `FocusedBorder` / `UnfocusedBorder` — that naming is the trap
- Any pattern `if m.focused { render with accent } else { render with muted }` around a single-pane view
- Tab key bound to "cycle panes" when there is only one pane

**Phase to address:** Chrome titled-border phase. Write the rule into the phase acceptance criteria: "titled borders are one color, no focus variants in v1.1."

---

### Pitfall 8: Teatest Golden Files Become Unstable Across Go / lipgloss Versions Due to ANSI Byte Drift

**What goes wrong:**
teatest golden-file tests compare raw byte output including ANSI escape sequences. lipgloss v2 changed the escape-sequence emission pattern between v2.0.0 and v2.0.4 (consolidated `ESC[38;2;R;G;Bm` runs into shorter `ESC[38;2;R;G;Bm;48;2;R;G;Bm` combined forms when adjacent). Go compiler changes can alter the order of style application in rare cases (unlikely but documented in charm.land issues). The effect: golden files that passed on a developer's laptop fail in CI three weeks later when CI bumps Go from 1.26.0 → 1.26.2, or when lipgloss releases v2.0.5 with a different ANSI emission. Every PR then regenerates goldens and the diffs become meaningless noise.

**Why it happens:**
Golden files encode presentation details (ANSI byte sequences) rather than semantic content (styled regions). The presentation layer has higher churn than the semantic layer.

**Consequences:**
- CI flakes on Go or lipgloss bumps; devs start running `-update` reflexively without checking what changed
- Real regressions hide inside cosmetic escape-sequence changes
- Golden file size grows (ANSI codes are verbose), PR reviews become hard

**How to avoid:**
1. **Golden files compare ANSI-stripped text** for structural assertions. Use `ansi.Strip(output)` (charmbracelet/x/ansi) before comparing. One golden file for "what the user reads," separate from any pixel-exact color assertion.
2. **Color assertions are separate, targeted tests.** Assert `lipgloss.Width(menu) == expectedCells` and `strings.Contains(output, ansi.Foreground(ColorAccent))` as independent checks. Don't bundle color + layout into one mega-golden.
3. **Pin lipgloss and bubbletea exactly** (`=v2.0.4`, not `^v2.0.4`) in go.mod for golden-heavy packages. Treat a lipgloss bump as intentional: regen goldens in that PR only.
4. **Golden file ownership rule:** each golden belongs to one test. Never share goldens across tests — that's how "update one golden, break five others" happens.
5. **Normalize line endings and trailing whitespace** in the goldie helper (use goldie's `WithDiffEngine` and a post-processor). Different Go versions have emitted slightly different whitespace around empty lines.
6. **Commit golden files to git with a `.golden` extension** that's excluded from gofmt / editor auto-format. A stray editor strip-trailing-whitespace rewrite will silently break CI.

**Warning signs:**
- Golden files > 50 lines — too much surface area
- PR description says "regenerated goldens, no code change" — check what actually drifted
- `diff -u` between goldens shows only escape-sequence reordering, no visible change

**Phase to address:** Testing harness phase, parallel to chrome rendering. Build the ANSI-stripping test helper and the "semantic + color" split **before** writing any chrome teatest, so the pattern is established.

---

### Pitfall 9: Color-Only Status Indicators Fail for Colorblind Users (Deuteranopia Mangles Red/Green Chips)

**What goes wrong:**
The v1.0 status bar already has this problem (`✓` green, `✗` red, `⚠` yellow) but it's a single indicator per env component. The v1.1 chrome multiplies the usage: the logo color signals info/warn/error (k9s pattern: `LogoColorInfo`, `LogoColorWarn`, `LogoColorError` in logo.go:73–91). If info is green, warn is yellow, error is red, users with deuteranopia (≈6% of men) see green and yellow as nearly identical. The logo never visibly changes from info → warn. A user hits a soft-failure state (age key missing, file unreadable) and sees no indication. The chrome adds *more* color-only signals (chip activity, menu highlight, cluster-info line color for git-dirty status). Each new signal compounds.

**Why it happens:**
Color alone is the fastest signal to reach for when designing in a graphical editor with 20/20 color perception. The k9s chrome uses color-only for logo state because tview text is text; adding symbols takes code. Porting verbatim ports the accessibility bug.

**Consequences:**
- Silent failures for colorblind users: "I don't see the error." They hit operational problems (missing age key, corrupted file) that the chrome should surface visibly
- Compliance gaps if sops-tui is adopted in orgs with accessibility requirements
- Issues filed against the project that are hard to reproduce without accessibility testing tools

**How to avoid:**
1. **Every color-coded state gets a shape-coded companion.** Logo state:
   - info: green + no prefix (`sops-tui` as-is)
   - warn: yellow + `⚠ ` (U+26A0, width-1 in terminal) prefix OR `[WARN]` text label
   - error: red + `✗ ` prefix OR `[ERROR]` text label
   Already partially done in styles.go (`ErrorLabel`, `WarnLabel`) — extend to every new chrome state.
2. **Chips use text content, not color, for primary meaning.** `<files>` vs `<detail>` vs `<diff>` is already the convention (crumbs.go:68). Keep that. Active chip gets inverted bg+fg (structural difference, works under any color profile including monochrome).
3. **Logo row 2 (status text) uses the text itself to communicate state, not the color.** k9s already does this: the status row shows "Info", "Warn", or "Error" text plus the colored background. Keep the text label; it survives colorblindness and 16-color downsample both.
4. **Test with a deuteranopia simulator.** Run teatest output through a colorblind-simulation filter (e.g., `colorblind` npm tool) and visually review. Add this to the phase's DoD checklist, not to CI (too heavy).
5. **Underline or bold the active breadcrumb** in addition to background color change. Redundant encoding is the mitigation.

**Warning signs:**
- A state indicator that has *only* a foreground color difference, no text or glyph differentiation
- The phrase "red means error" in any comment — if the comment has to explain it, colorblind users will be stuck
- No accessibility test in the phase's acceptance criteria

**Phase to address:** Chrome rendering phase (same phase as colored chips and logo state). Redundant-encoding rule goes into the phase's acceptance checklist.

---

## Moderate Pitfalls

### Pitfall 10: Alt-Screen + Chrome Interactions With VSCode Integrated Terminal and SSH Clients

**What goes wrong:**
Bubble Tea v2 enables the alt-screen via `v.AltScreen = true` (model.go:1374). The alt-screen swap issues ANSI `ESC[?1049h` (enter alt-screen) and `ESC[?1049l` (leave). Known compat issues:
- **VSCode integrated terminal** (xterm.js-based) historically has flicker on alt-screen enter, and under WSL2+VSCode the alt-screen can render with a 1-row offset against the prompt position — a chrome header sitting at row 0 gets partially clipped by the VSCode status bar overlay on first frame.
- **SSH over slow links:** synchronized output mode (Mode 2026) is what the Cursed Renderer uses to batch frames. Some older SSH mux clients (old Putty, Mosh versions before 1.4) don't forward Mode 2026 and the header appears "jittery" during resize as it repaints before the body.
- **tmux inside VSCode:** double-nesting alt-screen. tmux's own alt-screen tracking interacts with the TUI's alt-screen, occasionally leaving the chrome visible after TUI exit.

**Why it happens:**
Alt-screen is an 80s-era ANSI feature that every terminal implements slightly differently. Adding persistent chrome means the "first frame after alt-screen entry" matters visually; it didn't matter much in v1.0 when the status bar was one line and the file list naturally filled the rest.

**Consequences:**
- Clipped logo on first frame in VSCode integrated terminal → user sees broken ASCII art on startup
- Mid-resize flicker on SSH → momentarily double-rendered chrome
- Residual chrome after exit in tmux+VSCode → user sees the logo still painted in their shell prompt area

**How to avoid:**
1. **Paint a fill frame immediately on alt-screen enter.** Before rendering the chrome, write a single `lipgloss.NewStyle().Background(ColorBg).Width(w).Height(h).Render("")` frame. This clears any residual content from the terminal and establishes the background. One extra frame, cheap.
2. **Synchronized output (Mode 2026) defaults on.** Cursed Renderer emits `ESC[?2026h` around frame writes. If any env variable forces legacy mode (for debugging), log a warning so users know why they're seeing jitter.
3. **Explicit exit cleanup.** On `tea.Quit`, emit one blank frame before the alt-screen leaves, not just the status "Exiting…" text. Prevents the last-rendered chrome from being stuck in tmux's buffer.
4. **Matrix test:** Manually verify on at least:
   - Linux + Alacritty (baseline)
   - Linux + Ghostty
   - Linux + tmux inside Alacritty
   - macOS + Terminal.app
   - macOS + iTerm2
   - WSL2 + Windows Terminal
   - VSCode integrated terminal (xterm.js)
   - SSH from macOS into Linux with chrome enabled
5. **README troubleshooting section.** Document known limitation on VSCode integrated terminal with a workaround (open the tool in an external terminal, or disable the chrome via a flag).

**Warning signs:**
- Exit leaves residual rendering in the prompt area — alt-screen isn't being released cleanly
- First frame shows a partial logo
- User reports "flicker" on resize in a specific terminal

**Phase to address:** Polish phase, after chrome is visually correct in Alacritty. Terminal compat sweep is a standalone checklist item.

---

### Pitfall 11: Header Info-Panel Leaks Secret Metadata to Clipboard / Shell History / Telemetry

**What goes wrong:**
The header info-panel (analog to k9s `ClusterInfo`) displays: `.sops.yaml` path, age key fingerprint, recipient count, git repo state, file count. Individually these aren't secret values — but:
- **Age key fingerprint** is derived from a private key. k9s `ClusterMeta` shows only fingerprints which are fine *per the age spec*, but if a dev shortcuts the display to show the key file *path plus its first/last chars* as "identification," that's a disclosure: the path tells attackers where the key is, the chars narrow brute-force space.
- **File paths** containing environment names (`.sops/prod/db.yaml`, `.sops/prod/stripe.yaml`) reveal deployment structure. If the path appears in the header and the user screenshots the TUI for a bug report or a demo, the full path is shared.
- **tmux copy-mode** captures the whole screen including chrome. A user who runs `prefix + [` to scroll up and then accidentally enters copy-mode captures chrome into tmux's buffer, which some configurations sync to the system clipboard.
- **Shell history** (chrome sits above the prompt after TUI exit in terminals that don't use alt-screen correctly — see pitfall 10) — if the chrome survives exit, the next `history` command shows nothing, but `tmux capture-pane` does.
- **Bug reports:** users paste TUI screenshots into issues. If the chrome shows a fingerprint + file path, that pair goes into the public issue tracker.

**Why it happens:**
Header info-panels are designed to be informative. The instinct is "show more context." The security model says "show the minimum the user needs to identify the environment they're in, nothing more."

**Consequences:**
- Public issue trackers accumulate small disclosures (fingerprints + paths) which together build an attacker's map of the project's SOPS layout
- A screenshot in a Slack channel reveals recipient keys to everyone with channel access
- tmux clipboard sync exports fingerprint to system clipboard, into clipboard manager history, onto disk

**How to avoid:**
1. **Fingerprint display: show ≤ 10 chars, prefixed with `…` to make truncation visible.** k9s ClusterInfo shows truncated values already (cluster_info.go:83 uses `infoCell` with expansion; content like `age1xyz…` rather than the full key).
2. **File path display: show path relative to repo root, not absolute.** Already the pattern in v1.0 (filelist.go). Apply same rule to the header's `.sops.yaml` path display — `/.sops.yaml` not `/home/alice/work/secretrepo/.sops.yaml`.
3. **Never show environment variable values in the header**, even if they seem innocuous. `SOPS_AGE_KEY_FILE` path is on-disk location disclosure.
4. **Recipient count is a number, never a list.** k9s shows counts; the expanded list lives in a separate view. Follow that pattern — `4 recipients` in the header, full list in the recipient management screen only.
5. **Document the "what appears on screen" as part of the security-review PR checklist.** Every new header field goes through a short 5-question review: "Can an attacker with screenshot access use this? Does it appear in `ps`? Does tmux capture-pane export it? Does clipboard-manager history retain it?"
6. **No auto-copy bindings in the chrome.** The header is display-only. Do not add any "press c to copy fingerprint to clipboard" feature — the clipboard surface is already audited for secret values; fingerprints would join that risk surface.

**Warning signs:**
- Any new field in the header struct that derives from private key material (fingerprint, short ID)
- A header field that uses an absolute path from `os.Getenv()`
- Copy / export actions bound from the chrome region

**Phase to address:** Header info-panel design phase. Security review gate before the panel ships.

---

### Pitfall 12: lipgloss Border Characters Render Differently on Fonts Without Full Box-Drawing Coverage

**What goes wrong:**
Titled borders use box-drawing characters from Unicode range U+2500–U+257F. Not every terminal font ships the full range. Specifically:
- **Default macOS Terminal font (Menlo)**: covers basic (`│ ─ ┌ ┘ etc`) but renders double-lines (`║ ═ ╔ ╚`) with gaps; heavy/bold variants (`┃ ━ ┏ ┛`) fall back to thin
- **Windows Terminal with Cascadia Code**: full coverage
- **DejaVu Sans Mono**: full
- **Fonts without the range**: characters render as the missing-glyph box `□` — the title border becomes a row of boxes
If the chrome uses rounded borders (`╭ ╮ ╰ ╯`), those are less commonly supported than right-angle (`┌ ┐ └ ┘`).

**Why it happens:**
Developers pick "pretty" border styles (`lipgloss.RoundedBorder()`, `lipgloss.ThickBorder()`, `lipgloss.DoubleBorder()`) without verifying coverage. Works on the developer's font, fails on user's font.

**Consequences:**
- User on macOS with default font sees a chrome of `□` boxes where borders should be
- Bug report: "sops-tui is broken on my machine" — it's the font, not the tool
- Mixed-width glyphs (some ornate box-drawing chars are width-2 on certain renderers) push the chrome off-grid

**How to avoid:**
1. **Use only `lipgloss.NormalBorder()`** for v1.1. Single-line right-angle box drawing. Universal coverage.
2. **No rounded corners, no double lines, no heavy/bold borders.** Ever. They're cosmetic and fragile.
3. **Title position: `lipgloss.Border`'s title feature uses the top edge.** Confirm the rendered title fits within the border width; emoji in title will corrupt the top-edge alignment (see pitfall 6).
4. **Test matrix includes a "minimal font" environment** — a VM with only the default macOS terminal font — and assert borders render as expected box-drawing chars, not `□`.

**Warning signs:**
- Any use of `lipgloss.RoundedBorder()` / `lipgloss.ThickBorder()` / `lipgloss.DoubleBorder()` in the chrome
- User screenshot showing border positions as boxes

**Phase to address:** Chrome titled-border phase. Enforced via a lint/grep that rejects non-`NormalBorder` usage.

---

### Pitfall 13: Bubble Tea v2 `tea.KeyPressMsg` Narrower Than Expected — Key Bindings Fire on Release Instead of Press

**What goes wrong:**
Bubble Tea v2 introduced `tea.KeyMsg` as an interface; `tea.KeyPressMsg` and `tea.KeyReleaseMsg` implement it (per CLAUDE.md). Existing code handles `tea.KeyPressMsg`. When adding new chrome keybindings (toggle header `Ctrl-E` per k9s `toggleHeaderCmd` — app.go:636), it's tempting to handle `tea.KeyMsg` directly to catch both press and release. That fires the toggle twice per keypress: once on press, once on release. User toggles chrome and it comes right back.

**Why it happens:**
Migration from Bubble Tea v1 (where `tea.KeyMsg` was a struct, always a press event). Developers muscle-memory write `case tea.KeyMsg:` in the v2 code; the linter may not catch it because it's valid Go.

**Consequences:**
- Chrome toggle flickers — on then off immediately
- Any new binding (logo-state override, skin-switch, info-panel toggle) that handles `tea.KeyMsg` instead of `tea.KeyPressMsg` has the same bug

**How to avoid:**
1. **Lint rule / CI grep:** any `case tea.KeyMsg:` in the codebase is a regression. Must be `case tea.KeyPressMsg:` (the type assertion `msg.(tea.KeyPressMsg)` is also acceptable).
2. **Keep the v1.0 pattern** — every keymap handler already uses `tea.KeyPressMsg`. New chrome handlers follow that convention without deviation.
3. **Test:** simulate a key press + release via teatest and assert the bound action fires exactly once.

**Warning signs:**
- Any chrome handler triggered twice per input
- `case tea.KeyMsg:` matches in `git grep`

**Phase to address:** Chrome keybinding phase, when adding the first new binding (toggle chrome / cycle skin).

---

## Minor Pitfalls

### Pitfall 14: Breadcrumb Chip Wraps When Path Is Long; Chrome Overflows to Second Row

**What goes wrong:**
Existing breadcrumb renders `sops-tui > files > prod.yaml` as plain text (statusbar.go:179). Replacing with `<sops-tui> <files> <prod.yaml>` chips makes each segment wider. On narrow terminals (80 cols), deep paths (`sops-tui > recipients > remove > age1xy…`) can exceed the available width. If the chrome row has no wrap suppression, it wraps to a second row, breaking the fixed-height header budget.

**How to avoid:** Truncate middle segments with ellipsis when total width exceeds available space. Preserve first and last segments (most meaningful: root and current). Pattern: `<sops-tui> <…> <current>` when space is tight.

**Phase to address:** Breadcrumb chip phase.

---

### Pitfall 15: Info-Panel Values Fetched on Every Render Cause File System Stat Storms

**What goes wrong:**
`.sops.yaml` path, age key presence, git state are cheap individually but stat-ing them on every View call (60+ times/sec during scroll) produces unnecessary syscalls. On a network filesystem (NFS-mounted homedir), each stat is slow, and the info panel becomes the bottleneck.

**How to avoid:** Cache env state on the model; refresh only on startup, on `WindowSizeMsg`, and after operations that could change state (recipient add/remove, git commit detected). Pattern already established by `EnvStatus` (statusbar.go:36); extend to new header-info fields.

**Phase to address:** Header info-panel phase.

---

### Pitfall 16: Header Overlaps Search Input When `/` Is Pressed in File List

**What goes wrong:**
v1.0 search input appears as a row at the top of the file list (`filelist.go:346` shows `list.SetSize(width, height-1)` when search is active). With chrome at top, the search row either overlaps the chrome (if positioned at screen row 0) or the file list shrinks an extra row (if positioned below chrome). The visual effect is a search input sandwiched between the chrome and the titled border — confusing.

**How to avoid:** Search input renders *inside* the titled border frame, as its first row. The titled border's title can display "Files — searching: …" while the search box sits in row 1 of the bordered region. This keeps the chrome stable and signals search-mode via the title text.

**Phase to address:** Titled-border integration phase (when wrapping existing widgets).

---

### Pitfall 17: Re-rendering `View()` During SOPS Subprocess Leaves Chrome Stale

**What goes wrong:**
SOPS subprocess calls (decrypt, rotate, re-key) are async (v1.0 pattern). During the subprocess, the spinner runs in the status bar and the file list's "decrypting…" placeholder shows. The chrome (logo, info-panel) continues to render as it was before. If the subprocess fails with an error that changes env state (e.g., age key became unreadable mid-session because permissions changed), the header still shows "age: ✓" until the next state-changing message. The user sees an inconsistent UI briefly.

**How to avoid:** After any SOPS operation that can return an auth/access error, re-check env state and update the header. Pattern: the existing `checkEnvAsync` Cmd runs on startup; run it again on relevant error messages (sops exit codes 128/129 etc.). Cheap because env state is cached.

**Phase to address:** Chrome integration with async ops phase.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Hardcode chrome height as `const chromeRows = 7` | One less function call | Breaks when chrome becomes configurable (hide logo, hide info-panel); silent arithmetic bug | Never — use `chromeHeight(m)` helper |
| Build menu hints as string literals beside the render function | Faster first draft | Menu goes stale vs. keymap; every keybinding change needs two edits | Never — hints come from the keymap |
| Put skin parsing errors on the hot path with `.Render()` fallback | No startup blocker | Invisible text when user skin is bad; silent failure | Never — validate once at load |
| Use `lipgloss.NewStyle()` inside `View()` for "just this one" | Saves a style-var declaration | Per-frame allocations visible in benchmarks | Only in tests that benchmark the overhead itself |
| Share golden files across tests to avoid duplication | Fewer files | Cross-test coupling — updating one breaks five | Never — one test, one golden |
| Emoji in chrome chips because they look nicer | Visually richer | Width miscalculation on macOS Terminal; breaks on 16-color | Accepted in status-bar flash only (transient, non-structural) |

---

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Existing `WindowSizeMsg` handler | Forget to subtract `chromeHeight` from propagated dimensions | Use `bodyHeight(m)` helper; grep-gate prevents direct `m.height - statusBarHeight` |
| Existing `StatusBarModel.SetBreadcrumb` | Keep breadcrumb in the status bar after crumb-chip chrome lands | Migrate breadcrumb ownership to chrome; deprecate `SetBreadcrumb`; leave status bar breadcrumb-free |
| Existing `ClipboardHotStyle` indicator | Duplicate the `[clip]` indicator in chrome and status bar | Choose one location; remove from the other (recommend: keep in status bar, link to flash lifecycle) |
| Existing `state == stateHelp` rendering of full-screen overlay | Chrome still renders when help overlay is up | Help overlay is full-screen and opaque — skip chrome rendering when `m.state == stateHelp`. Simpler: help overlay takes over the full body region inside the chrome, letting chrome keep showing. Pick one, document it, enforce in snapshot tests. |
| Existing `recipientAction` sentinel | Menu hint builder uses `m.state` alone | Menu hint builder MUST consume `(m.state, m.recipientAction, m.fileList.IsSearchActive())` tuple |
| k9s `ClusterInfo` pattern | Port verbatim (10+ info rows) | sops-tui has ≤5 meaningful rows (.sops.yaml path, age fingerprint, recipient count, file count, git state). Don't pad. Fewer rows = more room for content. |
| k9s `SkinsDirWatcher` hot-reload | Port into v1.1 | Defer. Skins load once at startup in v1.1. |

---

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Rebuilding chrome strings every View | Input lag on held-down j/k | Cache chrome render; invalidate on state change only | Held-down key at > 20 Hz on terminals without Mode 2026 |
| `lipgloss.NewStyle()` inside View | GC pressure; profiler shows lipgloss.NewStyle in top allocators | Styles as package vars | > 100 Hz render loop |
| Info-panel fields fetched via syscall every frame | Slow first frame on NFS homedirs | Cache in model; refresh on events only | NFS or sshfs homedir; remote dev environments |
| Golden files include ANSI bytes | CI flakes on lipgloss bumps | ANSI-stripped goldens + separate color-presence assertions | Minor lipgloss version bumps |
| Emoji width recomputation | Subtle 1-column drifts on macOS Terminal | No emoji in persistent chrome | Always on macOS Terminal.app |

---

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Showing full age public key in info-panel | Fingerprint + path disclosure via screenshots | Truncate to ≤10 chars with visible ellipsis |
| Showing absolute `.sops.yaml` path with `$HOME` in it | Discloses username + directory structure | Show path relative to repo root |
| Showing `SOPS_AGE_KEY_FILE` env var value | Discloses key file location | Derive presence boolean only; never show the path |
| Adding `copy fingerprint to clipboard` keybinding in chrome | Expands clipboard surface area for auditing | Chrome is display-only; no copy actions |
| Emitting chrome content to stderr logs | Logs contain fingerprint/path | Never log chrome state; log only event types |
| tmux `capture-pane` exports chrome | Exfiltration via clipboard managers | Warn in README; minimize sensitive fields |
| Screenshot bug report includes header → public issue tracker | Metadata disclosure | Security review gate on every new header field; README note: "redact header before sharing screenshots" |

---

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Chrome visible even in full-screen overlay states (diff, help) | Confusing focus; user presses menu-hint keys that don't apply | Full-screen overlays hide or dim chrome; menu hints reflect overlay bindings only |
| Breadcrumb chips wrap on narrow terminal | Chrome overflows to 2nd row; body shrinks | Middle-segment ellipsis; fixed chrome height contract |
| Focus ring on single-pane views | User hunts for non-existent Tab navigation | No focus variants in v1.1 |
| Menu hints diverge from actual keybindings | User follows hint; nothing happens; loses trust | Keymap is single source; menu is derived, snapshot-tested |
| Color-only error states | Colorblind users see no difference between warn and error | Redundant shape/text encoding |
| Skin errors kill startup | User locked out of tool for cosmetic misconfig | Fail-open with status-bar warning |
| Logo visible after TUI exit | Residual paint in shell | Emit blank frame before alt-screen leave |

---

## "Looks Done But Isn't" Checklist

- [ ] **Chrome height subtraction:** All 16+ `SetSize` / `NewXxxModel` call sites in model.go pass `bodyHeight(m)`, not `m.height - statusBarHeight(m)`. Grep-gated.
- [ ] **Menu hint source:** Every `(state, recipientAction, searchActive)` tuple has a snapshot test that asserts the rendered menu matches the keymap for that state.
- [ ] **Skin fail-open:** Manually corrupt `~/.config/sops-tui/skin.yaml` (bad hex, missing field, malformed YAML) → TUI still launches with default palette and a visible warning.
- [ ] **Color profile fallback:** Running with `TERM=xterm` (16-color) produces readable chips — verify on `TERM=xterm-16color` in a VM.
- [ ] **Emoji-free chrome:** `rg '[\x{1F000}-\x{1FFFF}]' internal/ui/*.go` finds hits only in non-chrome files (or only in status-bar flash lines).
- [ ] **Box-drawing only:** `rg 'lipgloss\.RoundedBorder\(\)|lipgloss\.DoubleBorder\(\)|lipgloss\.ThickBorder\(\)'` in chrome files returns no hits.
- [ ] **No `tea.KeyMsg` handlers:** `rg 'case tea\.KeyMsg:' internal/` returns nothing. All handlers use `tea.KeyPressMsg`.
- [ ] **Chrome render benchmark:** `go test -bench=BenchmarkAppView -benchmem` shows ≤ 50 µs/op and ≤ 10 KB/op at 200×60.
- [ ] **Terminal compat sweep:** Manual verification on Alacritty, macOS Terminal, iTerm2, Windows Terminal, VSCode integrated terminal. Each with and without tmux. Screenshot record per combo.
- [ ] **Accessibility redundant encoding:** Every color-coded state (info/warn/error, active chip, selected row) has a text or shape differentiator too.
- [ ] **Info-panel PII review:** Every header field has been reviewed against the 5-question security checklist (screenshot, ps, tmux, clipboard, telemetry).
- [ ] **Goldens comparing ANSI-stripped text:** Primary snapshot assertions don't embed ANSI bytes. Color presence is asserted in separate targeted tests.
- [ ] **No `lipgloss.NewStyle()` in `View()`:** `rg -A3 'func.*View\(\)' internal/` — no NewStyle calls inside.
- [ ] **Header cached, not recomputed per frame:** Model has a `headerCache` field or equivalent; updated in Update, read in View.
- [ ] **Chrome hidden or adapted in all 14 session states:** Manual walkthrough: enter each state via keybinding, confirm chrome is correct (menu shows right hints, no overflow, no flicker).

---

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Chrome height miscalculation shipped to users | LOW | Patch release with `bodyHeight()` helper; existing tests catch it on next PR |
| Skin breaks user's TUI | LOW (fail-open) | Already recoverable by spec — TUI launches with warning, user deletes bad skin file |
| Menu hints go stale | MEDIUM | Regenerate snapshot tests; backfill hint-to-keymap derivation; release patch. Users may have memorized wrong keys during the window. |
| Color downsample makes chips invisible on 16-color | MEDIUM | Ship the monochrome fallback palette; set via env override in meantime |
| Fingerprint disclosed in info-panel | MEDIUM (reputation, no key compromise) | Patch release truncating fingerprint; ask affected users to rotate recipients if PII concerns |
| Golden files break CI on every bump | HIGH (death by 1000 cuts) | Refactor to ANSI-stripped goldens + separate color assertions; regen once cleanly |
| Alt-screen residue in user's tmux buffer | LOW | Patch the exit cleanup to emit blank frame before alt-screen leave |
| Emoji chip width drift | LOW | Replace with ASCII letters + color; release patch |

---

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| #1 Chrome height arithmetic | v1.1 Phase 1 — layout primitives | `bodyHeight(m)` helper exists; grep-gate in CI; resize test at 3 terminal sizes |
| #2 Render cost amortization | v1.1 Phase 2 — chrome rendering | `BenchmarkAppView` target ≤ 50 µs/op; no `lipgloss.NewStyle()` in View |
| #3 Menu hint state sync | v1.1 Phase 3 — menu integration | Snapshot test per `(state, action, search)` tuple; keymap→menu derivation |
| #4 Skin fail-open | v1.1 Phase 4 — theming | Corrupt-skin test; validation errors surface via status flash |
| #5 Color profile downsampling | v1.1 Phase 2 — chrome rendering | Multi-profile teatest matrix |
| #6 Unicode width | v1.1 Phase 1 — layout primitives | Emoji-free rule in chrome; runewidth.Truncate for user strings |
| #7 Focus indicator in single-pane | v1.1 Phase 2 — titled borders | No `FocusedBorder`/`UnfocusedBorder` pattern in styles |
| #8 Golden file stability | v1.1 Phase 0 — testing harness | ANSI-stripped comparison helper exists before chrome teatest |
| #9 Color-only accessibility | v1.1 Phase 2 — chrome rendering | Every state has redundant shape/text encoding; DoD checklist item |
| #10 Alt-screen compat | v1.1 Phase 5 — polish | Terminal matrix sweep recorded |
| #11 Secret metadata leakage | v1.1 Phase 3 — header info-panel | 5-question security review per field; truncation enforced |
| #12 Border character fonts | v1.1 Phase 2 — titled borders | Only `lipgloss.NormalBorder()`; grep-gated |
| #13 `tea.KeyMsg` vs KeyPressMsg | v1.1 Phase 3 — chrome keybindings | Grep-gate; single-press/release test |
| #14 Breadcrumb wrap | v1.1 Phase 3 — breadcrumb chips | Narrow-terminal (40-col) snapshot test |
| #15 Info-panel stat storms | v1.1 Phase 3 — header info-panel | Info-panel fields cached; invalidation tied to events |
| #16 Search input + chrome overlap | v1.1 Phase 2 — titled borders | Search-mode snapshot test shows input inside bordered region |
| #17 Chrome stale during async ops | v1.1 Phase 5 — async op integration | Env re-check after SOPS error messages |

---

## Sources

Project-internal:
- `.planning/research/v1.0-archive/PITFALLS.md` — v1.0 pitfalls (not re-catalogued here)
- `/home/moersener/git/sops-tui/internal/app/model.go` — View composition, statusBarHeight helper, WindowSizeMsg propagation sites
- `/home/moersener/git/sops-tui/internal/ui/styles.go` — color palette, existing style inventory
- `/home/moersener/git/sops-tui/internal/ui/statusbar.go` — breadcrumb, flash, env indicator patterns
- `/home/moersener/git/sops-tui/internal/ui/filelist.go` — SetSize propagation, list pagination dependency on height

k9s source (patterns to emulate / patterns to adapt):
- `/home/moersener/git/k9s/internal/view/app.go:273–331` — toggleHeader, buildHeader, 7-row header constant
- `/home/moersener/git/k9s/internal/view/cluster_info.go` — ClusterInfo layout, truncation, style listener pattern
- `/home/moersener/git/k9s/internal/ui/menu.go` — HydrateMenu, StackPushed/StackPopped hints lifecycle, maxRows constant, keyConv for mac alt/opt
- `/home/moersener/git/k9s/internal/ui/logo.go` — Logo state info/warn/error, refreshLogo, styles listener
- `/home/moersener/git/k9s/internal/ui/crumbs.go` — chip rendering with bg+fg from styles.Frame()

External references (confirmed current 2026-04):
- Bubble Tea v2 migration — https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md (tea.KeyPressMsg interface, View struct)
- lipgloss v2 color profiles — https://pkg.go.dev/charm.land/lipgloss/v2 (explicit profile detection API)
- runewidth — https://pkg.go.dev/github.com/mattn/go-runewidth (k9s dependency, used for Truncate)
- teatest — https://charm.land/blog/teatest/ (golden file + WaitFor patterns)
- Mode 2026 synchronized output — https://gist.github.com/christianparpart/d8a62cc1ab659194337d73e399004036
- VSCode integrated terminal alt-screen issues — https://github.com/microsoft/vscode/issues (known long-standing, multiple threads on xterm.js alt-screen behavior)
- Colorblind-safe palette guidance — https://www.nature.com/articles/nmeth.1618 (Wong 2011, cited in many a11y docs)

---
*Pitfalls research for: v1.1 k9s visual parity milestone*
*Researched: 2026-04-23*
