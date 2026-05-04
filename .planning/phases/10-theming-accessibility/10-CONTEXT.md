# Phase 10: Theming + Accessibility - Context

**Gathered:** 2026-05-04
**Status:** Ready for planning

<domain>
## Phase Boundary

Make the v1.1 chrome legible across 16-color terminals, communicative through severity coupling + redundant encoding, and survivable across the 40–200 column matrix — without introducing a user-facing skin loader (deferred to v2). Five requirements close: UI-03 (logo severity), UI-12 (k9s-tuned palette), UI-13 (16-color fallback), UI-14 (redundant shape/text encoding), UI-16 (narrow-terminal aesthetics).

**In scope (this phase):**
- `internal/ui/statusbar.go` — typed flash severity. New `FlashSeverity` enum (`FlashSevInfo` / `FlashSevWarn` / `FlashSevErr`). New methods `FlashInfo(msg)` / `FlashWarn(msg)` / `FlashErr(msg)`; existing `Flash(msg)` becomes a thin wrapper for `FlashInfo`. New accessor `FlashSeverity() FlashSeverity` so `resolveLogoState()` can read it. `View()` prepends `[W] ` or `[E] ` (no `[I]`) at render time; severity-tinted bg on Warn/Err (`ColorWarning` / `ColorError` bg + `ColorBg` fg); Info keeps current `StatusBarStyle` surface bg.
- `internal/app/model.go` — 42-callsite flash migration. Re-classify each `m.status.Flash(...)` per the severity rules: ~15 error paths → `FlashErr`, ~5 warn paths → `FlashWarn`, ~22 neutral paths stay on `Flash`/`FlashInfo`. New `resolveLogoState() ui.LogoStatus` method co-located with chrome composition (Pattern 4 from ARCHITECTURE.md). Wire output into both `RenderChrome` calls (`AppModel.View()` and the `chromeHeight` helper).
- `internal/ui/styles.go` — palette tune. `ColorAccentHex` `#89b4fa` → `#cba6f7` (Catppuccin Mauve). `ColorWarningHex` `#f9e2af` → `#fab387` (Catppuccin Peach). `ColorErrorHex` `#f38ba8` → `#eba0ac` (Catppuccin Maroon). All other palette constants unchanged.
- `internal/ui/styles.go` — 16-color fallback variants. New parallel block: `ColorAccentANSI`, `ColorBgANSI`, `ColorSurfaceANSI`, `ColorFgANSI`, `ColorMutedANSI`, `ColorSuccessANSI`, `ColorWarningANSI`, `ColorErrorANSI` — `lipgloss.ANSIColor(N)` values verified-distinct under 4-bit (per Pitfall 5 §4). Small `Palette(profile colorprofile.Profile) Palette` accessor returning the right set; `Palette` is a struct of named `lipgloss.Color` fields. The accessor is consulted at chrome-render entry, not per-style.
- `internal/ui/chrome.go`, `internal/ui/crumbs.go`, `internal/ui/menu.go`, `internal/ui/infopanel.go`, `internal/ui/logo.go` — accept a `colorprofile.Profile` parameter (or a pre-resolved `Palette` value) and select the right style variants. Renderer signatures change but all returns stay `string`.
- `internal/ui/crumbs.go` — bracket fallback chip rendering on Ascii/ANSI. Active chip drops bg fill, applies `Underline(true).Bold(true)` (no fg recolor). Inactive chip drops bg fill, plain fg. 24-bit/ANSI256 keeps the Phase 8 D-206 pill-fill rendering (bg accent + inverted fg + bold).
- `cmd/sops-tui/main.go` — call `colorprofile.Detect(os.Stdout, os.Environ())` once at startup; pass into `app.NewAppModel(env, sopsYamlPath, profile)`. AppModel stores a read-only `profile colorprofile.Profile` field; never re-detected.
- `internal/app/testdata/` + `internal/ui/testdata/` — 4-profile golden matrix per representative state (`Ascii` / `ANSI` / `ANSI256` / `TrueColor`). Force the profile via `lipgloss.Writer.Profile = ...` (or equivalent test seam) before render. ANSI-stripped structure goldens stay single-profile (palette-independent); color-presence assertions parameterise by `Color*Hex` constants so the named-constant tests survive the hex flip without per-test edits.
- `internal/app/testdata/` — narrow-terminal goldens at additional widths (60×24 + 100×30 in addition to the existing 40×12 / 80×24 / 120×40 / 200×60). Verifies UI-16 across the full range without ballooning the matrix.
- Single `GOLDEN_UPDATE=1 go test ./...` regen pass after palette change — atomic commit so the reviewer sees "color SGR bytes changed, no structural drift."
- Grep-gate scope continues unchanged: `TestChromeASCIIOnly` (the bracket fallback chips use ASCII `<` and `>`, already in allowlist), `TestChromeNormalBorderOnly`, `TestViewNoNewStyle` (the new `Palette` accessor, profile-resolved styles, and bracket-fallback styles all stay package vars).

**Out of scope (deferred per ROADMAP / explicit decisions):**
- User-facing skin YAML loader, builtin skins (dracula/gruvbox/monokai), live skin reload (`THM-01..THM-03`, all v2).
- Full `Skin` struct scaffolding for v2 forward-compat — research SUMMARY.md §"Phase 10" suggested it but the v1.1 ROADMAP merged it into v2; this phase ships the palette flat in `styles.go` only.
- `BenchmarkAppView ≤ 50 µs/op` re-target (Phase 11 SC2; D-18 caching fallback if the palette tune regresses the bench).
- v1.0 functional regression test sweep (Phase 11 SC1).
- Terminal compat sweep + alt-screen cleanup + "Looks Done But Isn't" sign-off (Phase 11).
- `:` command bar / number-key view switching / `[`/`]` history navigation (v1.2+).
- Stale-file findings raising the logo (rejected — too noisy; stale stays below Warn).
- Git-dirty raising the logo (rejected — most users always have at least one dirty file; logo would be near-permanently yellow).
- Logo status-text row / 7th row inflation (rejected — Phase 7 6-row art stays locked; severity is communicated via color + flash prefix, no in-logo status string).
- AppModel-cached severity field (rejected — `resolveLogoState()` is a pure function of `(env, flashSeverity, lastHealthResult)` and is computed per-frame; no caching needed because all inputs are already on the model).
- Sticky logo state after flash auto-clear (rejected — pure function of state; logo drops to baseline when flash clears, matches k9s).
- Per-frame profile detection (rejected — startup once is the standard).
- Polling goroutine for env re-check (forbidden by ROADMAP "Explicitly Out of Scope for v1.1").

</domain>

<decisions>
## Implementation Decisions

### Severity Classifier (UI-03, UI-14)

- **D-401: LogoError raised by ANY of: flash classified Err, non-empty `HealthCheckResult` (Weak ∪ Duplicate ∪ Errors), or persistent env failure.** "Persistent env failure" detection plumbs through existing event paths (`FilesDiscoveredMsg`, `GitStatusMsg`, recipient ops, edit success) — no polling, no per-frame `os.Stat`. The startup gate already eliminates `sops` missing before the TUI launches, so the in-TUI error path is principally driven by mid-session events. Stale files and git-dirty are deliberately excluded from error severity.
- **D-402: LogoWarn raised by ANY of: soft env (`!Env().AgeAvailable` or `!Env().SopsYamlAvailable`), or flash classified Warn.** Stale files and git-dirty stay BELOW Warn — they remain visible per-file (file-list badges, git: row in info panel, health overlay) but do not promote the logo. This avoids the "logo is permanently yellow" failure mode that would arise if any uncommitted change raised severity.
- **D-403: Logo state is a pure function of state, computed per-frame.** `resolveLogoState() ui.LogoStatus` reads `m.status.Env()` + `m.status.FlashSeverity()` + `m.lastHealthResult` and returns the resolved status. When a flash auto-clears (after 2s), severity drops to baseline naturally — no sticky state, no acknowledgment requirement, no separate `pendingErrorAck` field. Pattern 4 from ARCHITECTURE.md verbatim.
- **D-404: Severity precedence: Err > Warn > Info.** When multiple inputs apply (e.g., a Warn flash overlapping a clean env + non-empty health Err findings), the highest severity wins. Single-pass switch in `resolveLogoState()` walks Err checks first, then Warn checks, falls through to Info.
- **D-405: Logo art and rows stay locked at the Phase 7 6-row layout.** No in-logo status text row, no row 6 repurposing, no 7th row inflation. Severity is communicated via color (`LogoStyleInfo` / `LogoStyleWarn` / `LogoStyleError`, all already declared in styles.go) plus the redundant `[W]` / `[E]` flash prefix (D-411). `RenderLogo(LogoStatus, width)` signature unchanged from Phase 7.

### Flash Typed-API (UI-03)

- **D-406: Three new methods alongside existing `Flash`.** `FlashInfo(msg string) (StatusBarModel, tea.Cmd)`, `FlashWarn(msg string) (StatusBarModel, tea.Cmd)`, `FlashErr(msg string) (StatusBarModel, tea.Cmd)`. Existing `Flash(msg string)` becomes a thin wrapper for `FlashInfo` so backward-compat is preserved without a one-shot 42-callsite signature break. Migration cost is per-callsite re-classification, not signature update.
- **D-407: 42-callsite migration plan.** Re-classify each `m.status.Flash(...)` in `internal/app/model.go`: error paths (decrypt error, edit error, re-encryption failed, rotation failed, health scan failed, git history error, etc.) become `FlashErr`; soft-validation failures ("No changes detected", "No changes", "Reveal first with r", "Read error", "Diff error") become `FlashWarn`; success / informational paths ("Decrypted", "Re-encrypted", "Rotated to {format}") stay on `Flash`/`FlashInfo`. Plan author audits each callsite against the severity classifier in D-401 / D-402 and applies the matching method.
- **D-408: `FlashSeverity` enum lives in `internal/ui` (statusbar.go); `resolveLogoState()` lives in `internal/app` (model.go).** The enum and the three flash methods are local to the StatusBarModel surface (Pattern 4 in ARCHITECTURE.md). The classifier function aggregates inputs across packages (env + flash + health) so it's centralised in AppModel where those inputs already converge — no cross-package wiring beyond the `FlashSeverity()` accessor.
- **D-409: `FlashSevInfo` is the zero value.** `FlashSevInfo = 0`, `FlashSevWarn = 1`, `FlashSevErr = 2`. Existing zero-value StatusBarModel constructions stay safe — a freshly-constructed status bar with no flash fired yet returns `FlashSevInfo` from the accessor, which `resolveLogoState()` treats as the neutral baseline.
- **D-410: `FlashSeverity()` accessor returns `FlashSevInfo` when the flash is empty.** `m.flash == ""` means no active flash; the severity field is irrelevant; the accessor returns the zero value so the classifier short-circuits to env-derived severity. The severity field is updated when a typed flash method fires; cleared back to zero on `FlashClearMsg` ack (alongside `m.flash = ""`).

### Redundant Encoding (UI-14)

- **D-411: `[W]` / `[E]` flash prefix is added at render time, not at `Flash()` call time.** Inside `StatusBarModel.View()`'s flash branch: when `m.flashSeverity == FlashSevWarn`, prepend `"[W] "` to the rendered text; when `FlashSevErr`, prepend `"[E] "`. `FlashSevInfo` (the dominant case for the 22 neutral flashes) renders unprefixed — no `[I]` clutter on day-to-day status messages. Test fixtures for `m.flash` strings stay unchanged because the prefix is rendered, not stored.
- **D-412: Severity-tinted bg on Warn/Err flash bar.** `FlashSevWarn` flash row paints with `Background(ColorWarning).Foreground(ColorBg)` (peach bg + dark fg post-D-415). `FlashSevErr` flash row paints with `Background(ColorError).Foreground(ColorBg)` (maroon bg + dark fg post-D-415). `FlashSevInfo` keeps the existing `StatusBarStyle` surface bg + fg. Two-channel encoding (color + prefix) on Warn/Err; surface-only on Info.
- **D-413: Active crumb chip stays at 3-channel encoding on TrueColor/ANSI256 (Phase 8 D-206 unchanged).** bg=accent + fg=bg + bold. Phase 10 ADDS the bracket-fallback chip variant for Ascii/ANSI profiles (D-422) — the 24-bit path doesn't change.
- **D-414: Status-bar env indicators (`sops:✓` / `age:⚠` / `.sops.yaml:⚠`) stay as-is.** Phase 4 already pairs label-text + glyph + color; that's text-redundant + symbol-redundant + color-coded. UI-14 is satisfied at the env-indicator surface without change.

### Palette Tune (UI-12)

- **D-415: Three hex flips in `internal/ui/styles.go` constants.** `ColorAccentHex = "#89b4fa"` → `"#cba6f7"` (Catppuccin Mauve). `ColorWarningHex = "#f9e2af"` → `"#fab387"` (Catppuccin Peach). `ColorErrorHex = "#f38ba8"` → `"#eba0ac"` (Catppuccin Maroon). Bg / Surface / Success / Muted / Fg unchanged — minimum diff while addressing UI-12's "k9s hot-pink/purple" wording and giving the new mauve enough separation from Warning + Error so paired chip color contrast holds. The named `Color*` lipgloss vars derived from these constants pick up the new hex automatically.
- **D-416: Single GOLDEN_UPDATE pass after the palette change is committed atomically.** Land the three hex flips in one commit; run `GOLDEN_UPDATE=1 go test ./...`; commit refreshed `.golden` files in a single follow-up commit so the reviewer sees a clean "palette tune — only color SGR bytes changed, no structural drift" diff. Per-file regen as separate commits is rejected because it multiplies review surface and obscures the pure-color delta.
- **D-417: Color-presence test assertions reference named constants, not hex literals.** Tests using `RequireGoldenColors(t, name, output, []string{ui.ColorAccentHex})` continue to pass after the hex flip because the constant is the source of truth. Plan author audits the 16 test references identified in scout to confirm none use hex literals directly; any literal-using tests are migrated to constant references in the same plan.
- **D-418: No accent-related Phase 1-5 style is renamed.** Existing styles like `BreadcrumbActive = Foreground(ColorAccent)`, `TreeIndicator = Foreground(ColorAccent)`, `MenuKeyStyle = Foreground(ColorAccent)` automatically inherit the new mauve. No accent-specific style needs touching beyond the hex constant change. Same applies to `LogoStyleInfo = Foreground(ColorAccent)`.

### 16-Color Fallback (UI-13)

- **D-419: Profile detection happens once at startup in `cmd/sops-tui/main.go`.** Call `colorprofile.Detect(os.Stdout, os.Environ())` after the existing validation gate; pass the resolved `colorprofile.Profile` value into `app.NewAppModel(env, sopsYamlPath, profile)`. AppModel stores the profile on a read-only field set at construction. Pure-function path; never re-detected; never per-frame.
- **D-420: Parallel `var (...)` block of `Color*ANSI` named-ANSI fallback variants.** New entries in `styles.go` declared next to the 24-bit vars: `ColorAccentANSI = lipgloss.ANSIColor(13)` (bright magenta — distinct from Warning yellow + Error red under 4-bit), `ColorBgANSI = lipgloss.ANSIColor(0)` (black), `ColorSurfaceANSI = lipgloss.ANSIColor(8)` (bright black), `ColorFgANSI = lipgloss.ANSIColor(15)` (bright white), `ColorMutedANSI = lipgloss.ANSIColor(7)` (white), `ColorSuccessANSI = lipgloss.ANSIColor(10)` (bright green), `ColorWarningANSI = lipgloss.ANSIColor(11)` (bright yellow), `ColorErrorANSI = lipgloss.ANSIColor(9)` (bright red). Pitfall 5 §4 hand-verification: every paired bg/fg used in chrome (chip bg/fg, menu bg/fg, info-panel label/value, titled-border line/bg) maps to distinct ANSI indices.
- **D-421: `Palette(profile)` accessor returns a `Palette` struct of resolved colors.** Single function in `styles.go` that switches on profile (`<= colorprofile.ANSI` returns the ANSI variants, otherwise the 24-bit vars). Chrome renderers receive the profile (or pre-resolved Palette) at entry and apply it through their existing styles. Renderers that already accept `width` add a `profile colorprofile.Profile` (or `palette Palette`) parameter — `RenderChrome`, `RenderCrumbs`, `RenderMenu`, `RenderInfoPanel`, `RenderLogo`. Plan author picks Profile vs Palette for the parameter type — recommend Palette so renderers don't import `colorprofile`.
- **D-422: Bracket-fallback chip rendering on Ascii/ANSI profiles.** When `profile <= colorprofile.ANSI`, `RenderCrumbs` renders chips as plain `<segment>` with no bg fill; the active chip uses `Underline(true).Bold(true)` (no fg recolor — the underline + bold survive 16-color downsample as structural cues). When `profile > colorprofile.ANSI`, the Phase 8 D-206 pill-fill rendering applies unchanged. Two new style vars: `CrumbChipFallbackStyle` (no bg, fg = `ColorFgANSI` on the fallback path), `CrumbChipActiveFallbackStyle` (Underline + Bold, no bg). The `<` and `>` literals already pass `TestChromeASCIIOnly`.
- **D-423: 4-profile teatest matrix per representative state.** Force `lipgloss.Writer.Profile` (or use a teatest seam) to each of `Ascii` / `ANSI` / `ANSI256` / `TrueColor` and capture goldens for: `RenderChrome` at full-tier (200×60), `RenderCrumbs` with the active chip, `RenderMenu` with one populated state, the flash bar at each of the three severities. Total: ~16 new color-bearing goldens (4 profiles × 4 representative scenes). ANSI-stripped structure goldens stay single-profile.

### Narrow-Terminal Survival (UI-16)

- **D-424: Width matrix expands to 6 captured widths.** Existing: 40×12 (Phase 7.1), 80×24 (Phase 7.1 + Phase 8), 120×40 (Phase 8), 200×60 (Phase 8). Phase 10 adds: 60×24 (mid-narrow tier between Phase 7.1's narrow and Phase 8's mid-tier breakpoints) and 100×30 (mid-tier with crumbs). Captures the gap regions in the existing 3-tier chrome fallback (D-116) and proves the bracket-fallback chips render correctly when the active chip middle-truncates.
- **D-425: "Critical data must survive" rule means the active crumb + the currently-selected file + flash text are non-truncatable.** When the chip row would overflow at narrow widths, middle-segment ellipsis kicks in (Phase 8 D-216 — already shipped) but the FIRST and LAST chips are always preserved. The active (last) chip with its underline-or-pill fill marker is the user's "you are here" anchor — losing it would silently break navigation context. Plan author confirms the existing `truncateSegmentsToWidth` helper guarantees first+last preservation; if a regression is found, fix as part of Plan 3.

### Plan Split (3-plan ROADMAP budget)

- **D-426: Three plans, primitive-first matching Phase 7 + Phase 8 splits:**
  - **Plan 1 — Severity classifier + flash typed API + flash bg tint + redundant prefix:**
    - `internal/ui/statusbar.go`: `FlashSeverity` enum, `FlashSeverity()` accessor, `FlashInfo`/`FlashWarn`/`FlashErr` methods (Flash becomes wrapper); `View()` flash branch extended for severity-tinted bg + `[W]`/`[E]` prefix; updated tests.
    - `internal/app/model.go`: `resolveLogoState()` method; `RenderChrome` call sites (2 locations) pass the resolved `LogoStatus` instead of unconditional `LogoInfo`; 42-callsite flash re-classification per D-407.
    - Unit tests: severity classifier table (every (env, flashSev, healthResult) tuple → expected LogoStatus), flash-prefix render assertions, flash-bg-tint render assertions.
    - Zero palette change in this plan; all current goldens stay green.
  - **Plan 2 — Palette tune + profile detection + 16-color fallback:**
    - 3 hex constant flips in `internal/ui/styles.go` per D-415.
    - Parallel `Color*ANSI` block per D-420; Pitfall 5 §4 hand-verification table.
    - `Palette` struct + `Palette(profile colorprofile.Profile) Palette` accessor.
    - `cmd/sops-tui/main.go`: `colorprofile.Detect` call; `app.NewAppModel` signature change.
    - `internal/ui/{chrome,crumbs,menu,infopanel,logo}.go`: signature change to accept `profile`/`palette`; profile-aware style selection.
    - Atomic `GOLDEN_UPDATE=1 go test ./...` regen pass per D-416.
    - Hex-literal-using tests audited and migrated to named constants per D-417.
  - **Plan 3 — Bracket-fallback chips + 4-profile teatest matrix + narrow-terminal sweep:**
    - `internal/ui/crumbs.go`: profile-aware chip rendering per D-422; `CrumbChipFallbackStyle` + `CrumbChipActiveFallbackStyle` package vars.
    - 4-profile teatest matrix per D-423: ~16 new color-bearing goldens.
    - 60×24 + 100×30 narrow-terminal goldens per D-424.
    - Critical-data-survival regression test per D-425.
    - Update `08-VERIFICATION.md`-style verification artifacts for Phase 10 closure.
- **D-427: Plan 1 is the largest plan.** 42-callsite flash sweep + classifier + flash bar work + redundant prefix + bg tint all converge in one place. Splitting differently (e.g., classifier in Plan 1, flash methods in Plan 2) would re-open `model.go` multiple times and force tests to be written twice. Plan 2 ships alongside the GOLDEN_UPDATE wave so the palette flip + fallback variants land together; Plan 3 is the surface-area expansion (test matrix + narrow widths + UI-16 closure).

### Folded Todos

- **T-401: STATE.md "Phase 10 research pass: aggregate health severity classification (is 'git dirty' a warn or info? is 'stale recipient key' a warn?)"** — folded entirely. Resolved by D-401 (LogoError signals: flash Err, non-empty health finding, persistent env failure) and D-402 (LogoWarn signals: soft env, flash Warn). Stale files and git-dirty are explicitly demoted below Warn.
- **T-402: STATE.md "Color-profile downsampling on 16-color terminals — paired chip bg/fg collapses to single color on `TERM=xterm`; mitigation is Phase 10 profile detection + 16-color fallback palette"** — folded entirely. Resolved by D-419 (startup profile detection), D-420 (parallel ANSI variants), D-421 (Palette accessor), D-422 (bracket-fallback chips).

### Reviewed Todos (not folded)

- **STATE.md "Phase 10/11: revisit BenchmarkAppView budget — currently 5 ms with 56% headroom over ~2.8 ms/op measurement; D-18 caching fallback (model-level cache keyed on (state, recipientAction, IsSearchActive, width)) can tighten this if user-perceived latency matters"** — kept deferred to Phase 11 SC2 per Phase 7.1 Plan 01's path-c deferral and ROADMAP §"Phase 11". Phase 10's signature changes (RenderChrome/RenderCrumbs/etc gain a profile parameter) do not affect the bench budget materially.

### Claude's Discretion

- Exact `Palette` struct shape — whether it groups colors by surface area (`Palette.Chip.ActiveBg`) or stays flat (`Palette.Accent`, `Palette.Warning`, etc.). Plan 2 author picks; recommendation: flat to mirror the existing `Color*` var naming.
- Whether the renderer parameter is `profile colorprofile.Profile` or a pre-resolved `palette Palette`. Plan 2 author picks; recommendation: `Palette` because it removes the `colorprofile` import dependency from every UI file.
- ANSI color index choices (`ANSIColor(13)` for accent vs `ANSIColor(5)` etc.) — Plan 2 author runs the Pitfall 5 §4 hand-verification table and picks indices that satisfy "every chrome bg/fg pair is distinct under 4-bit." If a non-trivial swap is needed (e.g., bright-cyan-on-bright-blue collides), planner picks the alternative and documents in 10-02-SUMMARY.md.
- Whether `resolveLogoState()` returns `ui.LogoStatus` (a struct of `{Message, State}`) or just `ui.LogoStatus` (the enum value). Today only `LogoStatus` (the enum) is consumed by `RenderLogo`; the message field from ARCHITECTURE.md Pattern 4 is dropped because D-405 keeps the logo art locked. Plan 1 author picks; recommendation: return just the enum (drop the unused message struct).
- Whether `m.lastHealthResult` is a new field on AppModel or pulled from `m.health` (the existing HealthModel sub-model). Plan 1 author scouts; recommendation: add a small accessor on the existing sub-model rather than duplicating state.
- File location of the `Palette` accessor (top of `styles.go` vs new `internal/ui/palette.go`). Plan 2 author picks; recommendation: top of `styles.go` because it's tightly coupled to the palette declarations.
- Whether to expose a `SOPSTUI_FORCE_ASCII=1` env var override for users who want to force the fallback path. Plan 2 author picks; recommendation: yes, it's a 4-line addition (consult env in `cmd/sops-tui/main.go` before `colorprofile.Detect`) and answers "16-color goldens look great but my fancy terminal still mis-detects" support questions cheaply.
- Logo art width-responsiveness — `RenderLogo` already accepts a `width` parameter (currently ignored). Plan 2 author may trim trailing spaces or center the art at narrow widths; not required for UI-03 but a low-cost polish if it falls out of profile detection work.
- Whether the 4-profile teatest matrix forces profile via `lipgloss.Writer.Profile = ...` or via a teatest constructor option. Plan 3 author picks; recommendation: whichever the lipgloss/v2 + teatest current API supports cleanly.
- Whether the narrow-tier (`<41` cols) at the 4-profile matrix gets its own goldens or inherits from full-tier. Plan 3 author picks; recommendation: at narrow tier the chrome is a 1-row "press ? for help" stub (Phase 7.1 D-116) so a single representative ANSI golden is sufficient — the chrome has no chips at that width.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project decision docs

- `.planning/ROADMAP.md` §"Phase 10: Theming + Accessibility" — Goal, 5 success criteria, 3-plan budget, UI-03/UI-12/UI-13/UI-14/UI-16 requirements
- `.planning/ROADMAP.md` §"Milestone v1.1 — Explicitly Out of Scope for v1.1" — reject list (skin loader, polling goroutines, `:` command bar, mouse interactions, animated/gradient logos, runtime skin reload, etc.)
- `.planning/REQUIREMENTS.md` §"Theming & Accessibility" — UI-12 (k9s palette tune + AdaptiveColor ban), UI-13 (16-color fallback), UI-14 (redundant shape/text encoding), UI-15 (ASCII-only chrome — applies to fallback chip rendering), UI-16 (40×12–200×60 survival)
- `.planning/REQUIREMENTS.md` §"Header Region" — UI-03 (logo recolors to reflect aggregate severity)
- `.planning/PROJECT.md` — k9s visual parity is a hard product attribute (per user-memory `feedback_k9s_visual_parity.md`); v1.0 functional non-regression; AGPL-3.0 license; AdaptiveColor ban via #1036

### Prior phase decisions (carried forward, do not re-decide)

- `.planning/phases/07-chrome-skeleton/07-CONTEXT.md` D-02 (Logo default = ColorAccent unconditional in Phase 7; severity coupling deferred to Phase 10) — Phase 10 implements the deferred wiring
- `.planning/phases/07-chrome-skeleton/07-CONTEXT.md` D-13 (NormalBorder() + ColorMuted border foreground) — palette tune does NOT change border color
- `.planning/phases/07-chrome-skeleton/07-CONTEXT.md` D-18 (pure every-frame composition + caching deferred to Phase 11) — Phase 10's signature changes preserve the pure-function discipline; D-18 caching stays Phase 11
- `.planning/phases/07.1-chrome-gap-closure/07.1-CONTEXT.md` D-116 (3-tier chrome width fallback: narrow `<41` → "press ? for help" stub; mid `41 ≤ width < 99` → menu+logo; full `≥99` → info-panel + menu + logo) — Phase 10 layers profile-aware variants on top; tier breakpoints unchanged
- `.planning/phases/07.1-chrome-gap-closure/07.1-CONTEXT.md` (relevant decisions) — narrow-terminal body-reachable contract at 40×12 + 80×24
- `.planning/phases/08-header-info-panel/08-CONTEXT.md` D-206 (active crumb chip 3-channel encoding: bg+fg+bold) — Phase 10 keeps this on TrueColor/ANSI256, ADDS bracket variant on Ascii/ANSI
- `.planning/phases/08-header-info-panel/08-CONTEXT.md` D-216 (crumbs middle-segment ellipsis on overflow) — UI-16 critical-data-survival rule (D-425) builds on this; first+last chip preservation becomes a tested guarantee
- `.planning/phases/09-keybinding-discoverability/09-CONTEXT.md` D-310 (menu goldens are RenderMenu-only, ANSI-stripped, structure-only) — Phase 10's palette tune does not churn these; only the new color-bearing goldens at the 4-profile matrix add to the test surface

### Research — v1.1 milestone (MUST READ)

- `.planning/research/SUMMARY.md` §"Phase 9: Logo state + flash severity + palette tune" — original phase scope (now merged into ROADMAP Phase 10); FlashErr/FlashWarn/FlashInfo triad rationale
- `.planning/research/SUMMARY.md` §"Phase 10: k9s-tuned default palette shipped; skin YAML loader deferred to v1.2" — clarifies that v1.1 ships only the tuned default palette; Skin struct prep deferred to v2 per ROADMAP
- `.planning/research/SUMMARY.md` §"Aggregate health severity classification" — calls out the ROADMAP-flagged ambiguity that Phase 10 D-401/D-402 closes
- `.planning/research/PITFALLS.md` §"Pitfall 4: Skin Loading Fails Closed" — sidestepped entirely by deferring the user-facing loader to v2
- `.planning/research/PITFALLS.md` §"Pitfall 5: Color-Profile Downsampling on 16-Color Terminals" — Phase 10's primary mitigation source: profile detection + ANSI variants + bracket fallback for chips. §4 hand-verification table is the test rule for D-420
- `.planning/research/PITFALLS.md` §"Pitfall 6: Unicode Width Miscalculation" — fallback chip uses ASCII `<` `>` already in allowlist; emoji-free chrome rule continues to apply
- `.planning/research/PITFALLS.md` §"Pitfall 9: Color-Only Status Indicators Fail for Colorblind Users" — D-411 (flash prefix) + D-422 (bracket+underline active chip) are the redundant-encoding mitigations; Wong 2011 colorblind-safe palette guidance referenced
- `.planning/research/PITFALLS.md` §"Pitfall 15: Stat'ing the Filesystem on Every View Call" — `resolveLogoState()` is pure; persistent env failure detection plumbs through existing event paths only (D-401)
- `.planning/research/ARCHITECTURE.md` §"Pattern 4: Data-Derived Logo State (not Event-Driven)" — `resolveLogoState()` template; D-403 verbatim
- `.planning/research/ARCHITECTURE.md` §"Phase Sequencing" — Phase 10 follows Phase 8 (chrome data paths) + Phase 9 (menu drives from hints)
- `.planning/research/STACK.md` §"Bubble Tea v2 / lipgloss/v2" — `colorprofile.Detect` API and `lipgloss.Writer.Profile` rules

### Existing implementation (Phase 10 modifies / extends)

- `cmd/sops-tui/main.go` — startup gate; Phase 10 adds `colorprofile.Detect` after validation; `app.NewAppModel` signature change
- `internal/app/model.go` — `AppModel` struct + `View()` + `Update()` handlers + 42 `m.status.Flash(...)` call-sites + `RenderChrome` call sites (2 locations) + `crumbsHeight()` (calls `RenderCrumbs` — signature change cascade) + `chromeHeight()` (calls `RenderChrome` — signature change cascade)
- `internal/ui/statusbar.go` — `StatusBarModel`: add `flashSeverity` field, `FlashSeverity()` accessor, `FlashInfo`/`FlashWarn`/`FlashErr` methods; extend `View()` flash branch
- `internal/ui/styles.go` — 3 hex constant flips (D-415); add 8 `Color*ANSI` parallel vars (D-420); add `Palette` struct + `Palette()` accessor (D-421); add `CrumbChipFallbackStyle` + `CrumbChipActiveFallbackStyle` package vars (D-422); existing 30+ named styles unchanged
- `internal/ui/chrome.go` — `RenderChrome` signature gains profile/palette parameter; full/mid/narrow tier paths consult the profile to pick variant styles
- `internal/ui/crumbs.go` — `RenderCrumbs` signature gains profile/palette parameter; profile-aware chip rendering per D-422
- `internal/ui/menu.go` — `RenderMenu` signature gains profile/palette parameter; menu key + desc colors switch on profile
- `internal/ui/infopanel.go` — `RenderInfoPanel` signature gains profile/palette parameter; label color switches on profile
- `internal/ui/logo.go` — `RenderLogo(LogoStatus, width)` may also accept profile (planner discretion); `LogoStyleInfo`/`Warn`/`Error` already exist in styles.go (Phase 7)
- `internal/health/checker.go` — `HealthCheckResult` shape (`WeakSecrets`, `Duplicates`, `StaleFiles`, `Errors`) — `resolveLogoState()` reads this; `IsClean()` exists already
- `internal/app/chrome_test.go` — Phase 7 + 7.1 + 8 grep-gates (`TestChromeASCIIOnly`, `TestChromeNormalBorderOnly`, `TestViewNoNewStyle`); Phase 10 extends color-bearing goldens but keeps the grep-gate scope unchanged (no new files inside chrome that need ASCII-only enforcement)
- `internal/testutil/golden.go` — `RequireGoldenStructure` (Phase 6) + `RequireGoldenColors` (Phase 6); Phase 10's 4-profile matrix uses both
- `internal/keys/bindings.go` — keymap extraction unchanged; no new keybindings in Phase 10

### Technology / external references

- `charm.land/lipgloss/v2` package docs — https://pkg.go.dev/charm.land/lipgloss/v2 — `Writer.Profile`, `ANSIColor`, color downsampling rules
- `github.com/charmbracelet/colorprofile` v0.2.x — https://pkg.go.dev/github.com/charmbracelet/colorprofile — `Detect(output io.Writer, env []string) Profile`; profile values `NoTTY` / `Ascii` / `ANSI` / `ANSI256` / `TrueColor`; respects `NO_COLOR`, `CLICOLOR`, `CLICOLOR_FORCE`
- `CLAUDE.md` §"Core TUI Framework" — `lipgloss.AdaptiveColor` ban (#1036) continues to apply; Phase 10's fallback uses explicit `ANSIColor(N)` not adaptive
- `CLAUDE.md` §"Testing" — golden file stability via `charmbracelet/x/exp/teatest` + ANSI-stripped comparison helper; Phase 10's 4-profile matrix forces `lipgloss.Writer.Profile` in tests

### k9s visual parity references (project memory: hard quality attribute)

- `~/git/k9s/internal/ui/logo.go:60-90` — k9s `Err`/`Warn`/`Info` methods on `Logo`; severity-coupled refresh pattern. **D-401, D-402 source for the per-state classification model; sops-tui's pure-function approach (D-403) replaces k9s's event-driven update.**
- `~/git/k9s/internal/ui/flash.go` — k9s flash bar with severity-tinted bg. **D-412 source; sops-tui follows the same bg-tint convention.**
- `~/git/k9s/internal/config/styles.go` — k9s skin schema with `LogoColor`, `LogoColorInfo`, `LogoColorWarn`, `LogoColorError` keys. **Forward-compat reference for v2 `THM-01..03` skin loader; not implemented in Phase 10.**

### Code files in scope (Phase 10 modifies / adds)

- `cmd/sops-tui/main.go` — `colorprofile.Detect` call + `NewAppModel` signature change. Plan 2
- `internal/app/model.go` — `AppModel.profile` field; `resolveLogoState()`; 42-callsite flash re-classification; `RenderChrome` + `RenderCrumbs` call-site signature updates. Plans 1 + 2 (Plan 1 owns the severity + flash bits; Plan 2 owns the profile plumbing)
- `internal/ui/statusbar.go` — `FlashSeverity` enum, accessor, three flash methods, `View()` flash branch. Plan 1
- `internal/ui/styles.go` — 3 hex flips (D-415); 8 ANSI variants (D-420); `Palette` struct + accessor (D-421); 2 fallback chip styles (D-422). Plan 2
- `internal/ui/chrome.go` — `RenderChrome` signature change; profile-aware tier path selection. Plan 2
- `internal/ui/crumbs.go` — `RenderCrumbs` signature change; bracket-fallback chip rendering. Plans 2 + 3 (Plan 2 plumbs the parameter; Plan 3 wires the fallback rendering)
- `internal/ui/menu.go` — `RenderMenu` signature change; profile-aware menu key/desc style. Plan 2
- `internal/ui/infopanel.go` — `RenderInfoPanel` signature change; profile-aware label/value style. Plan 2
- `internal/ui/logo.go` — `RenderLogo` may gain profile parameter. Plan 2
- `internal/app/chrome_test.go` — color-bearing goldens at 4-profile × representative-state matrix; narrow-terminal goldens at 60×24 + 100×30. Plans 1 + 3
- `internal/ui/{statusbar,crumbs}_test.go` — flash-prefix + flash-bg-tint render assertions; bracket-fallback chip render assertions. Plans 1 + 3
- `internal/app/testdata/` — refreshed color goldens (palette tune wave) + new 4-profile matrix + 60×24/100×30 widths. Plans 2 + 3
- `.planning/phases/10-theming-accessibility/10-UI-SPEC.md` — to be authored by `/gsd-ui-phase 10` (recommended before Plan 1)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `internal/ui/logo.go` `LogoStyleInfo`/`LogoStyleWarn`/`LogoStyleError` package vars — already declared in Phase 7 (D-02); Phase 10 wires them via the typed `LogoStatus` parameter that `RenderLogo` already accepts.
- `internal/ui/styles.go` `Color*Hex` string constants — single source of truth for the palette; tests reference these constants. Phase 10's 3 hex flips (D-415) plus 8 new ANSI variants (D-420) join the existing additive var-block pattern.
- `internal/ui/statusbar.go` `Flash(msg) (StatusBarModel, tea.Cmd)` — generation-counter pattern (Pitfall 6) preserved unchanged; Phase 10 extends with three typed methods that all share the same `flashGen++` semantics.
- `internal/ui/statusbar.go` `FlashClearMsg{Gen int}` — clears flash on tick; Phase 10 also clears `flashSeverity` field on the same ack so severity drops back to baseline naturally.
- `internal/health/checker.go` `HealthCheckResult.IsClean()` — already returns `true` when all four slices are empty; Phase 10's `resolveLogoState()` consults this to classify health severity.
- `internal/ui/chrome.go` `RenderChrome` 3-tier width fallback (Phase 7.1 D-116) — Phase 10 layers profile-aware variant selection on top; tier breakpoints (`<41` / `41 ≤ width < 99` / `≥99`) unchanged.
- `internal/ui/crumbs.go` `truncateSegmentsToWidth` (Phase 8 D-216) — first+last preservation already implemented; D-425's critical-data-survival rule reuses this without modification.
- `internal/testutil/golden.go` `RequireGoldenColors(t, name, output, []string{wantColors...})` — color-presence assertion (Phase 6); Phase 10's 4-profile matrix uses both this and `RequireGoldenStructure`.
- `cmd/sops-tui/main.go` — clean entry point; Phase 10's `colorprofile.Detect` call slots in cleanly between Step 5 (env build) and Step 6 (NewProgram).

### Established Patterns

- **Pure functions for renderers + styles as package vars** — `internal/ui/*.go` exposes `RenderX(...) string`; `TestViewNoNewStyle` BFS walker enforces no `NewStyle()` reachable from `View()`. Phase 10's new fallback styles (`CrumbChipFallbackStyle`, `CrumbChipActiveFallbackStyle`) and the `Palette` struct fields stay package-level.
- **Explicit hex (`Color()`) for 24-bit + ANSI indices (`ANSIColor(N)`) for fallback** — no `lipgloss.AdaptiveColor` (#1036). Phase 10 doubles down on this with the ANSI fallback variants.
- **Helper tests co-located with implementations** — `crumbs_test.go` lives beside `crumbs.go`; same for `infopanel_test.go`, `menu_test.go`, etc. Phase 10's new tests follow the pattern.
- **3-callsite atomic refactor pattern** — Phase 7.1 + Phase 8 demonstrate that signature changes through chrome renderers (`RenderChrome` + `RenderCrumbs` + helpers) land cleanly in a single plan. Phase 10's profile-parameter cascade follows the same shape.
- **GOLDEN_UPDATE wave commits** — Phases 7 + 8 each had a "refresh goldens" commit at the end of their integration plan. Phase 10 has two such waves: one for the palette flip (Plan 2), one for the 4-profile matrix (Plan 3).
- **Async data refresh via tea.Cmd messages** — `FilesDiscoveredMsg`, `GitStatusMsg`, `HealthCheckResultMsg` are the existing event-driven seams. Phase 10's "persistent env failure" detection (D-401) plumbs through these existing seams; no new event types.
- **Empty/zero-state handling via terse markers** — Phase 8's `-` for missing values, Phase 5's `(none)` for empty recipient lists. Phase 10's `FlashSevInfo` zero-value semantics (D-409) follow the same low-friction defaults.

### Integration Points

- `cmd/sops-tui/main.go:65` — Step 5 builds `EnvStatus`; insert `profile := colorprofile.Detect(os.Stdout, os.Environ())` immediately after.
- `cmd/sops-tui/main.go:70` — `model := app.NewAppModel(env, sopsYamlPath)` becomes `model := app.NewAppModel(env, sopsYamlPath, profile)`.
- `internal/app/model.go` `NewAppModel` — accepts the new `profile` parameter; stores on `m.profile colorprofile.Profile`.
- `internal/app/model.go:1400` + `model.go:1606` — `ui.RenderChrome(hints, ui.LogoInfo, m.infoPanel, m.width)` becomes `ui.RenderChrome(hints, m.resolveLogoState(), m.infoPanel, m.width, m.palette())` (or equivalent) — both call sites updated.
- `internal/app/model.go` `crumbsHeight(m)` — currently calls `RenderCrumbs(m.status.Segments(), m.width)`; signature update flows here too.
- `internal/app/model.go` 42 `m.status.Flash(...)` call-sites — re-classified per D-407.
- `internal/ui/statusbar.go` `StatusBarModel` struct — gains `flashSeverity FlashSeverity` field.
- `internal/ui/statusbar.go:160` `View()` flash branch — extends with severity-aware bg + prefix per D-411 + D-412.
- `internal/ui/styles.go` Color* vars — 3 hex flips at constants; 8 new ANSI variants in a parallel block; `Palette` struct + accessor at top of file.
- `internal/ui/chrome.go` + `crumbs.go` + `menu.go` + `infopanel.go` + `logo.go` — signature change to accept `palette Palette` (or `profile colorprofile.Profile`); profile-aware style selection inside.
- `internal/app/chrome_test.go` — extends the color-bearing test suite with 4-profile × representative-state matrix; adds 60×24 + 100×30 narrow-terminal goldens.
- `internal/ui/crumbs_test.go` — adds bracket-fallback rendering assertions; existing pill-fill tests stay green on TrueColor profile.
- `internal/ui/statusbar_test.go` — adds severity-tinted bg + prefix assertions; existing flash tests stay green on `FlashSevInfo` baseline.
- `go.mod` — `github.com/charmbracelet/colorprofile` already pulled in transitively via lipgloss/v2; `cmd/sops-tui/main.go` adds it as a direct import.

</code_context>

<specifics>
## Specific Ideas

- **Severity classifier is the broadest read of error signals.** D-401 takes "any non-empty `HealthCheckResult` (Weak ∪ Duplicate ∪ Errors)" as the raise-to-error trigger. The user picked all four health options in the multi-select; rather than reading that as "pick one of two MX options," I treated it as "use the union — every health-output category counts as error." Stale files alone (HealthCheckResult.StaleFiles non-empty without any weak/duplicate/error) stay below Warn — that's the explicit demotion in D-402.
- **Logo art stays locked at 6 rows.** D-405 rejects both row-6 repurposing and 7th-row inflation. Phase 7's locked logo is preserved; severity is communicated through color (`LogoStyle*`) plus the redundant `[W]`/`[E]` flash prefix (D-411). This is the minimum-blast-radius choice — no chrome height math change, no goldens churn beyond the palette flip + profile matrix.
- **Three new Flash methods + Flash as alias.** D-406 picks the minimum-churn migration shape. Existing 22 neutral flashes don't move; only the 15 errors + 5 warns get explicit severity. The 42-callsite sweep is per-callsite re-classification, not a signature break — every existing callsite that says `m.status.Flash(...)` keeps compiling without modification, but the planner audits each one against D-401/D-402 and upgrades to `FlashErr`/`FlashWarn` where the severity rule applies.
- **Severity-tinted flash bar bg gives a stronger signal than prefix alone.** D-412 paints the row with `ColorWarning`/`ColorError` bg + `ColorBg` fg on Warn/Err. Two-channel encoding (color + prefix) on Warn/Err; surface-only on Info. The 16-color downsample paths still differentiate cleanly because `ColorWarningANSI` = bright yellow (11), `ColorErrorANSI` = bright red (9), `ColorSurfaceANSI` = bright black (8) — all three distinct in 4-bit.
- **Palette tune is three hex flips, not a full retune.** D-415 keeps Bg / Surface / Success / Muted / Fg unchanged. Accent shifts to Catppuccin Mauve (`#cba6f7`); Warning to Catppuccin Peach (`#fab387`); Error to Catppuccin Maroon (`#eba0ac`). The new mauve has visible separation from peach + maroon at 24-bit, and from bright-magenta + bright-yellow + bright-red under 4-bit downsample. Phase 1-5 styles that already say `Foreground(ColorAccent)` automatically inherit the new mauve.
- **Single GOLDEN_UPDATE pass after palette tune.** D-416 picks the atomic regen approach. Per-file regen as separate plan tasks would multiply the diff surface area and obscure the "only color changed" delta — the reviewer's job is much easier when the palette commit is tagged "color SGR bytes only, no structural drift."
- **Profile detection is one-shot at startup, never per-frame.** D-419 puts the `colorprofile.Detect` call in `cmd/sops-tui/main.go` between the existing validation and `tea.NewProgram`. AppModel stores the profile as a read-only field. Pitfall 15 spirit: zero per-frame I/O.
- **Bracket-fallback chip rendering is the Pitfall 5 §2 prescription verbatim.** D-422: on Ascii/ANSI, chips drop bg fill, render as `<segment>` plain, active chip uses `Underline + Bold` (no fg recolor — both attributes survive 16-color and monochrome). On TrueColor/ANSI256, the Phase 8 D-206 pill-fill rendering is unchanged.
- **4-profile teatest matrix is the Pitfall 5 §3 prescription verbatim.** D-423: Ascii + ANSI + ANSI256 + TrueColor each get color-bearing goldens for `RenderChrome` + `RenderCrumbs` (active chip) + `RenderMenu` + the flash bar at 3 severities. ~16 new goldens. ANSI-stripped structure goldens stay single-profile.
- **Plan 1 is the largest plan.** Severity classifier + 42-callsite flash sweep + flash bar redundancy work converge in one place. Splitting differently would re-open `model.go` multiple times. Plan 2 ships the palette flip + profile fork together so the GOLDEN_UPDATE wave is atomic. Plan 3 is the test-matrix expansion + UI-16 closure.

</specifics>

<deferred>
## Deferred Ideas

### Phase 11 (already scoped)

- v1.0 functional regression test sweep — UI-20 (file discovery, reveal, edit, diff, rotate, clipboard, git, recipient management, health all pass after chrome lands)
- `BenchmarkAppView` budget tightening to ≤50 µs/op — UI-21 (D-18 caching fallback if Phase 10's profile parameter cascade regresses the bench; current 5 ms with 56% headroom over ~2.8 ms/op)
- Terminal compat sweep + alt-screen cleanup — Alacritty / Ghostty / macOS Terminal / iTerm2 / Windows Terminal / VSCode integrated terminal / tmux-nested
- Full 15-item "Looks Done But Isn't" sign-off

### v2 (milestone-deferred per ROADMAP)

- User-facing skin YAML loader (`THM-01`) — `~/.config/sops-tui/skin.yaml` with k9s-compatible schema subset
- Builtin skins embedded via `embed.FS` — dracula / gruvbox-dark / monokai (`THM-02`)
- Live skin reload via fsnotify (`THM-03`)
- `Skin` struct scaffolding — research SUMMARY.md §"Phase 10" suggested it as forward-compat prep, but the ROADMAP merged it into v2 entirely
- `:skin <name>` runtime switcher — depends on `:` command bar (v1.2)

### Possibly Phase 11, possibly v2

- Logo width-responsive trimming / centering at narrow widths — `RenderLogo` already accepts `width`; not required for UI-03 but a low-cost polish if it falls out of profile detection work in Plan 2
- Health-finding severity at-rest classifier in `internal/health/checker.go` — Phase 10 reads `HealthCheckResult` via `IsClean()` + non-empty checks; a richer severity field on the result struct would be a Phase 11 cleanup if needed
- AppModel-cached severity field — D-403 picks pure per-frame computation; if Phase 11 bench work shows the classifier is hot, caching with invalidation on env/flash/health change is the documented fallback

### Out of scope this phase (would be scope creep)

- Stale files contributing to logo severity — explicit demotion in D-402; staleness stays per-file in the file list / health overlay
- Git-dirty contributing to logo severity — rejected as too noisy (most users always have at least one dirty file)
- Polling goroutine for env re-check (`os.Stat` on age key file every N seconds) — forbidden by ROADMAP "Explicitly Out of Scope for v1.1"
- Sticky logo state after flash auto-clear (`pendingErrorAck` field) — rejected; pure function of state per D-403
- Logo status text row inside the 6-row art — rejected; logo art stays locked per D-405
- Mouse interactions on chips — keyboard-only by core value
- "Press c to copy fingerprint" or any chrome-content copy binding — Pitfall 11 explicit ban (Phase 8 D-220 already enforces)
- Animated / pulsing / gradient logo — adds ticker goroutine, breaks teatest snapshots
- 5-profile matrix (adding `NoTTY`) — `colorprofile.NoTTY` only fires when the output isn't a terminal (e.g., piped to a file); the TUI can't render usefully there anyway, so coverage for this profile is not meaningful

### Reviewed Todos (not folded)

- "Phase 10/11: revisit BenchmarkAppView budget" (STATE.md) — kept deferred to Phase 11 SC2 per Phase 7.1 Plan 01 path-c deferral. Phase 10's signature changes (RenderChrome / RenderCrumbs / etc gain a profile parameter) do not materially affect the bench budget.
- "Manual UAT per Phase 06 D-15" (STATE.md) — terminal-resize verification, deferred to `/gsd-verify-work` for Phase 10 once the chrome lands.

</deferred>

---

*Phase: 10-theming-accessibility*
*Context gathered: 2026-05-04*
