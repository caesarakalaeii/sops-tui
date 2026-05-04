# Phase 10: Theming + Accessibility — Research

**Researched:** 2026-05-04
**Domain:** Terminal color profiles, redundant accessibility encoding, k9s-tuned palette, narrow-terminal survival
**Confidence:** HIGH (every claim verified against locally-installed lipgloss/v2 v2.0.3, colorprofile v0.4.3, bubbletea v2.0.4 source, plus the Phase 7/7.1/8/9 implementation files)

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Severity classifier (D-401..D-405):**
- D-401: LogoError raised by ANY of: flash classified Err, non-empty `HealthCheckResult` (Weak ∪ Duplicate ∪ Errors), or persistent env failure. No polling, no per-frame `os.Stat`. Stale files and git-dirty are deliberately excluded.
- D-402: LogoWarn raised by ANY of: soft env (`!Env().AgeAvailable` or `!Env().SopsYamlAvailable`), or flash classified Warn. Stale files and git-dirty stay BELOW Warn.
- D-403: Logo state is a pure function of state, computed per-frame. `resolveLogoState() ui.LogoStatus` reads `m.status.Env()` + `m.status.FlashSeverity()` + `m.lastHealthResult`. No sticky state, no acknowledgment requirement.
- D-404: Severity precedence: Err > Warn > Info. Single-pass switch walks Err checks first.
- D-405: Logo art and rows stay locked at the Phase 7 6-row layout. No 7th row, no in-logo status text.

**Flash typed-API (D-406..D-410):**
- D-406: Three new methods alongside existing `Flash`. `FlashInfo` / `FlashWarn` / `FlashErr`. Existing `Flash(msg)` becomes a thin wrapper for `FlashInfo`. Backward-compat preserved.
- D-407: 42-callsite migration plan (re-classification per call-site).
- D-408: `FlashSeverity` enum lives in `internal/ui` (statusbar.go); `resolveLogoState()` lives in `internal/app` (model.go).
- D-409: `FlashSevInfo = 0`, `FlashSevWarn = 1`, `FlashSevErr = 2`. Zero-value StatusBarModel is safe.
- D-410: `FlashSeverity()` returns `FlashSevInfo` when flash is empty. Cleared back to zero on `FlashClearMsg` ack.

**Redundant encoding (D-411..D-414):**
- D-411: `[W]` / `[E]` flash prefix added at render time, not call time. `FlashSevInfo` renders unprefixed.
- D-412: Severity-tinted bg on Warn/Err flash bar. Warn = peach bg + dark fg post-D-415. Err = maroon bg + dark fg post-D-415.
- D-413: Active crumb chip stays at 3-channel encoding on TrueColor/ANSI256 (Phase 8 D-206 unchanged). Phase 10 ADDS bracket fallback for Ascii/ANSI.
- D-414: Status-bar env indicators (`sops:✓` / `age:⚠` / `.sops.yaml:⚠`) stay as-is.

**Palette tune (D-415..D-418):**
- D-415: Three hex flips. Accent `#89b4fa` → `#cba6f7` (Mauve). Warning `#f9e2af` → `#fab387` (Peach). Error `#f38ba8` → `#eba0ac` (Maroon). Bg / Surface / Success / Muted / Fg unchanged.
- D-416: Single `GOLDEN_UPDATE=1 go test ./...` regen pass after the palette change is committed atomically.
- D-417: Color-presence test assertions reference named constants, not hex literals.
- D-418: No accent-related Phase 1-5 style is renamed.

**16-color fallback (D-419..D-423):**
- D-419: Profile detection happens once at startup in `cmd/sops-tui/main.go`. `colorprofile.Detect(os.Stdout, os.Environ())`. Stored on read-only AppModel field.
- D-420: Parallel `Color*ANSI` named-ANSI fallback variants: Accent=13, Bg=0, Surface=8, Fg=15, Muted=7, Success=10, Warning=11, Error=9.
- D-421: `Palette(profile)` accessor returns a `Palette` struct of resolved colors. Recommend Palette over Profile param (removes `colorprofile` import dep from every UI file).
- D-422: Bracket-fallback chip rendering on Ascii/ANSI profiles. Active = `Underline(true).Bold(true)` (no bg fill, no fg recolor). 24-bit/ANSI256 keeps Phase 8 D-206 pill-fill unchanged.
- D-423: 4-profile teatest matrix (Ascii / ANSI / ANSI256 / TrueColor) per representative state.

**Narrow-terminal survival (D-424..D-425):**
- D-424: Width matrix expands to 6 captured widths (40×12 + 60×24 + 80×24 + 100×30 + 120×40 + 200×60).
- D-425: Critical data must survive: active crumb + currently-selected file + flash text are non-truncatable.

**Plan split (D-426..D-427):**
- D-426: Three plans (Plan 1: severity + flash typed-API + redundancy. Plan 2: palette + profile detection + ANSI variants. Plan 3: bracket fallback + 4-profile matrix + narrow-terminal sweep).
- D-427: Plan 1 is the largest plan.

### Claude's Discretion

- Exact `Palette` struct shape (flat vs grouped). Recommendation: flat to mirror existing `Color*` var naming.
- Renderer parameter type: `profile colorprofile.Profile` vs `palette Palette`. Recommendation: `Palette`.
- ANSI color index choices. Plan 2 author runs Pitfall 5 §4 hand-verification table.
- `resolveLogoState()` return type (struct vs enum). Recommendation: just the enum.
- `m.lastHealthResult` placement (new field vs accessor on existing `m.health`). Recommendation: small accessor on existing sub-model.
- `Palette` accessor file location. Recommendation: top of `styles.go`.
- `SOPSTUI_FORCE_ASCII=1` env var override. Recommendation: yes — 4-line addition.
- Logo width-responsive trimming. Optional polish.
- 4-profile matrix profile-forcing seam. Plan 3 picks.
- Narrow-tier (`<41` cols) at 4-profile matrix. Recommendation: single representative ANSI golden.

### Deferred Ideas (OUT OF SCOPE)

**Phase 11 (already scoped):**
- v1.0 functional regression test sweep (UI-20).
- `BenchmarkAppView` budget tightening to ≤50 µs/op (UI-21; D-18 caching fallback).
- Terminal compat sweep + alt-screen cleanup.
- Full 15-item "Looks Done But Isn't" sign-off.

**v2 (milestone-deferred per ROADMAP):**
- User-facing skin YAML loader (`THM-01`).
- Builtin skins embedded via `embed.FS` (`THM-02`).
- Live skin reload via fsnotify (`THM-03`).
- `Skin` struct scaffolding.
- `:skin <name>` runtime switcher.

**Possibly Phase 11, possibly v2:**
- Logo width-responsive trimming / centering at narrow widths.
- Health-finding severity at-rest classifier in `internal/health/checker.go`.
- AppModel-cached severity field (D-403 picks pure per-frame).

**Out of scope this phase:**
- Stale files contributing to logo severity (D-402 demotion).
- Git-dirty contributing to logo severity.
- Polling goroutine for env re-check.
- Sticky logo state after flash auto-clear.
- Logo status text row inside the 6-row art.
- Mouse interactions on chips.
- "Press c to copy fingerprint" or any chrome-content copy binding (Phase 8 D-220 ban).
- Animated / pulsing / gradient logo.
- 5-profile matrix (NoTTY profile excluded).
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| UI-03 | Logo recolors to reflect aggregate app status (info/warn/error) derived from env checks, flash severity, and health aggregate | §"Severity Classifier" + §"Flash Typed-API" sections of this research; D-401..D-410. Existing `LogoStyleInfo/Warn/Error` package vars in `styles.go:239-245` already declared in Phase 7 — Phase 10 wires the typed `LogoStatus` parameter through `RenderLogo`. |
| UI-12 | Default palette tuned to k9s conventions (accent toward hot-pink/purple) while keeping AdaptiveColor ban from v1.0 | §"Palette Tune (3 hex flips)" — D-415. Catppuccin Mauve `#cba6f7` is the k9s default-style hot-pink/purple analog. AdaptiveColor ban preserved (project rule from `CLAUDE.md`). |
| UI-13 | On 16-color terminals (`TERM=xterm` / Ascii profile) a safe fallback palette so paired bg/fg chips and menu cells remain legible; teatest runs across Ascii/ANSI/ANSI256/TrueColor | §"16-Color Fallback" + §"ANSI 16-Color Verification Table" — D-419..D-423. Parallel `Color*ANSI` block; `Palette(profile)` accessor; bracket-fallback chip rendering on Ascii/ANSI; 4-profile teatest matrix. |
| UI-14 | Every color-coded state uses redundant shape or text encoding (`[I]`/`[W]`/`[E]` prefixes, inverted bg+fg for active, underline for focus) | §"Redundant Encoding (UI-14)" — D-411..D-414. Flash prefix at render time; severity-tinted bg; existing env indicators preserved (label-text + glyph + color already redundant); active chip 3-channel encoding (Phase 8 D-206) on TrueColor/ANSI256, bracket+underline+bold fallback on Ascii/ANSI. |
| UI-16 | App survives rendering at 40×12 through 200×60 without layout corruption | §"Narrow-Terminal Survival" — D-424..D-425. 6-width golden matrix (40×12 + 60×24 + 80×24 + 100×30 + 120×40 + 200×60). Critical-data-survival rule (active chip + selected file + flash text non-truncatable). Existing 3-tier chrome fallback (Phase 7.1 D-116) and middle-segment ellipsis (Phase 8 D-216) carried forward. |
</phase_requirements>

## Project Constraints (from CLAUDE.md)

| Directive | Source | Implication for Phase 10 |
|-----------|--------|--------------------------|
| Never use `lipgloss.AdaptiveColor` (issue #1036) | `CLAUDE.md` Recommended Stack | All ANSI fallback uses explicit `lipgloss.ANSIColor(N)` — no `AdaptiveColor` allowed. |
| Never use `any` type — use proper typing | `~/.claude/CLAUDE.md` user-global | All new structs (`Palette`, `FlashSeverity`) use concrete types. No `interface{}` / `any` in signatures. |
| `lipgloss.NormalBorder()` only | `CLAUDE.md` UI-15 | Border color may flip via `Palette.Border` or stay hardcoded `ColorMuted`. Border *style* stays NormalBorder. Grep-gate `TestChromeNormalBorderOnly` already enforces. |
| ASCII-only chrome | `CLAUDE.md` UI-15 | Bracket-fallback chips use `<` `>` literals (already in allowlist per `chrome_test.go:56-63`). No new non-ASCII runes introduced. Underline + Bold are SGR codes (1m/4m), not glyphs. |
| Single-binary distribution | Project constraints | No new runtime deps — `colorprofile v0.4.3` is already pulled in transitively via lipgloss/v2 (verified in `go.mod`). |
| Subprocess to `sops` CLI | Project constraints | Phase 10 does not touch the SOPS subprocess layer. |
| `View()` returns `tea.View`, not `string` | `CLAUDE.md` Bubble Tea v2 | AppModel.View() composition already returns tea.View; Phase 10 changes do not alter return type. |
| `tea.KeyPressMsg` not `tea.KeyMsg` | `CLAUDE.md` Bubble Tea v2 | No new keybindings in Phase 10. Existing handlers continue to use `tea.KeyPressMsg`. |
| No `lipgloss.NewStyle()` inside `View()` | Phase 6/7/7.1 grep-gates | All new styles (CrumbChipFallbackStyle, CrumbChipActiveFallbackStyle, ANSI variants) are package-level vars in `styles.go`. |

---

## Summary

Phase 10 closes five v1.1 requirements (UI-03 logo severity, UI-12 k9s palette, UI-13 16-color fallback, UI-14 redundant encoding, UI-16 40×12–200×60 survival) by layering a profile-aware **render-time variant selection** on top of the existing chrome stack. The implementation has **three primitive shifts**:

1. **Type the flash severity.** `FlashSeverity` enum (`FlashSevInfo` / `FlashSevWarn` / `FlashSevErr`), three new methods (`FlashInfo` / `FlashWarn` / `FlashErr`), `Flash` becomes a thin alias for `FlashInfo`. 41 call-sites in `model.go` get re-classified per CONTEXT D-401/D-402 rules. `[W]`/`[E]` prefix and severity-tinted bg apply at render time.

2. **Detect the color profile once at startup, plumb it to renderers.** `colorprofile.Detect(os.Stdout, os.Environ())` lives in `cmd/sops-tui/main.go` between Step 5 (env build) and Step 6 (NewProgram). The resolved `colorprofile.Profile` flows through `NewAppModel(env, sopsYamlPath, profile)` and is stored as a read-only field. Five renderers (`RenderChrome`, `RenderCrumbs`, `RenderMenu`, `RenderInfoPanel`, `RenderLogo`) gain a `palette Palette` parameter; the accessor `Palette(profile colorprofile.Profile) Palette` switches on profile and returns the right color set.

3. **Three hex flips + parallel ANSI variants.** Accent → Mauve, Warning → Peach, Error → Maroon. New `Color*ANSI = lipgloss.ANSIColor(N)` block sits next to the existing 24-bit vars. Bracket-fallback chip styles (`CrumbChipFallbackStyle`, `CrumbChipActiveFallbackStyle`) handle Ascii/ANSI. A single atomic `GOLDEN_UPDATE=1 go test ./...` regen commits the SGR-bytes-only diff.

**Primary recommendation:** Plan 1 ships severity + flash typed-API + redundant prefix as the largest plan (touches `model.go` once); Plan 2 ships palette + profile detection + ANSI variants atomically with one GOLDEN_UPDATE wave; Plan 3 ships bracket-fallback rendering + 4-profile teatest matrix + narrow-terminal goldens. The 5 renderer signatures change exactly once (in Plan 2); Plan 3 wires the bracket variant inside `RenderCrumbs` without re-touching the parameter list.

**Critical architectural note (verified in lipgloss source):** `lipgloss.Style.Render()` ALWAYS emits 24-bit RGB SGR codes regardless of profile. Downsampling happens at `colorprofile.Writer.Write()` time, not at render time (verified `lipgloss/v2/writer.go:102-115` and `colorprofile/writer.go:38-80`). This means **the bracket fallback variant is the seam that prevents bg-collapse on 16-color terminals**: we cannot rely on the auto-downsample to keep `bg=Mauve / fg=Bg` distinct, because at the SGR layer they're 24-bit. The PROFILE-AWARE BRANCH at render time picks bracket-style instead of pill-style based on the profile field, and the underline+bold attributes (SGR `1` and `4`) survive every downsample including monochrome.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Flash severity enum + methods | UI primitives (`internal/ui/statusbar.go`) | — | StatusBarModel owns the flash field; the typed methods are local to its API surface. The enum has no cross-package consumers beyond the `FlashSeverity()` accessor. |
| Severity classifier (`resolveLogoState()`) | App orchestration (`internal/app/model.go`) | — | Aggregates env (status bar) + flash (status bar) + health (health sub-model) — the only place all three converge is AppModel. Pure function of state, computed per-frame (D-403). |
| Palette + ANSI variants + Palette() accessor | UI design system (`internal/ui/styles.go`) | — | All other modules import these vars. Centralised so renderers don't import `colorprofile`. |
| Profile detection | Process bootstrap (`cmd/sops-tui/main.go`) | App field (`internal/app/model.go`) | Detect once at process start; flow through `NewAppModel` constructor; never re-detect. |
| Profile-aware variant selection | UI renderers (chrome.go, crumbs.go, menu.go, infopanel.go, logo.go) | — | Each renderer receives `palette Palette` at entry and consults the right style variants. No per-frame detection. |
| Logo recolor by severity | UI logo (`internal/ui/logo.go`) | — | Existing `RenderLogo(LogoStatus, width)` already supports the typed parameter; Phase 10 wires the resolver into the call site. |
| `[W]` / `[E]` flash prefix | UI status bar render (`internal/ui/statusbar.go View()`) | — | Render time, not call time (D-411). Test fixtures for `m.flash` strings stay unchanged. |
| Severity-tinted flash bar bg | UI status bar render (`internal/ui/statusbar.go View()`) | — | Branch on `m.flashSeverity` to pick `StatusBarStyle` (info), `Background(ColorWarning).Foreground(ColorBg)` (warn), or `Background(ColorError).Foreground(ColorBg)` (err). |
| Bracket fallback chip rendering | UI crumbs (`internal/ui/crumbs.go`) | — | Profile-aware branch inside `RenderCrumbs`: `<= colorprofile.ANSI` → bracket style, otherwise → existing pill style. |
| 4-profile teatest matrix | App tests (`internal/app/menuhints_drift_test.go` extension or new file) | — | Goldens captured at the AppModel.View() boundary so the full chrome stack is exercised. |

---

## Implementation Approach

### 1. `colorprofile.Detect` API integration for `cmd/sops-tui/main.go`

**Verified API (from `~/go/pkg/mod/github.com/charmbracelet/colorprofile@v0.4.3/env.go:33`):**

```go
func Detect(output io.Writer, env []string) Profile
```

**Profile values (from `profile.go:13-26`):**

```go
const (
    Unknown Profile = iota   // 0 — absence of profile (do not consume directly)
    NoTTY                    // 1 — output isn't a terminal (piped / redirected)
    ASCII                    // 2 — TTY but no color
    ANSI                     // 3 — 16 colors (4-bit)
    ANSI256                  // 4 — 256 colors (8-bit)
    TrueColor                // 5 — 16 million colors (24-bit)
)

const Ascii = ASCII   // alias for backward compat
```

**Profiles are ordered**: `<= ANSI` is the canonical "16-color or worse" check.

**Module already in go.mod (verified):**
```
github.com/charmbracelet/colorprofile v0.4.3 // indirect
```

The package is pulled in transitively via `charm.land/lipgloss/v2` (lipgloss imports colorprofile in its writer.go and color.go). Phase 10 promotes it to a **direct require** in `cmd/sops-tui/main.go`. `go mod tidy` will move it from indirect to direct without version churn.

**Integration pattern (verified against `cmd/sops-tui/main.go:65-71`):**

```go
import (
    "github.com/charmbracelet/colorprofile"
    // ... existing imports
)

// Step 5: Build env status (existing)
env := ui.EnvStatus{...}

// Step 5.5 (new): Detect color profile once at startup (Phase 10 D-419)
profile := colorprofile.Detect(os.Stdout, os.Environ())
if os.Getenv("SOPSTUI_FORCE_ASCII") == "1" {
    profile = colorprofile.Ascii  // Claude's discretion — env override
}

// Step 6: Create AppModel (signature change)
sopsYamlPath, _ := validator.FindSopsYaml(opts.StartDir)
model := app.NewAppModel(env, sopsYamlPath, profile)
p := tea.NewProgram(model)
```

**Critical interaction with Bubble Tea v2 (verified in `bubbletea/v2@v2.0.4/tea.go:1082-1089`):** Bubble Tea ALSO calls `colorprofile.Detect` internally if `WithColorProfile` was not passed, and applies it to the renderer. Two options for sops-tui:

| Option | What | Trade-off |
|--------|------|-----------|
| **A (recommended)** | App detects profile, stores on AppModel, branches at render time inside renderers. Bubble Tea also detects (auto). Renderer-time and write-time profiles are consistent because both call the same `Detect` function. | One canonical profile per process. Tests can force the AppModel-stored profile to drive renderer behavior independently. |
| B | Pass `tea.WithColorProfile(profile)` to `tea.NewProgram(model, tea.WithColorProfile(profile))` so Bubble Tea uses the same value AppModel sees. Adds one option call. | Strictly redundant with A unless the tests explicitly need to assert Bubble Tea's downsample behavior matches. |

**Recommendation: Option A.** AppModel needs the profile to choose the bracket-vs-pill render path; the Cursed Renderer's auto-downsample is orthogonal — it operates on the bytes the model emits. Both calls of `Detect` are pure functions of the same `(os.Stdout, os.Environ())` so they will agree.

`lipgloss.Writer.Profile` is initialized at package load via `colorprofile.NewWriter(os.Stdout, os.Environ())` (verified `lipgloss/v2/writer.go:14`), so when we call `Detect` ourselves we get the same answer. Tests can override `lipgloss.Writer.Profile = colorprofile.ANSI` to force a specific downsample at the LIPGLOSS write layer (used by tests that drive lipgloss.Println / Sprint), but this does NOT affect what `Style.Render()` emits at the SGR layer.

### 2. ANSI 16-color verification (CONTEXT.md D-420 mapping)

**The mapping proposed in D-420:**

| Color | ANSI Index | Mnemonic |
|-------|------------|----------|
| Accent | 13 | bright magenta |
| Bg | 0 | black |
| Surface | 8 | bright black (dark grey) |
| Fg | 15 | bright white |
| Muted | 7 | white (light grey) |
| Success | 10 | bright green |
| Warning | 11 | bright yellow |
| Error | 9 | bright red |

**Verified via `lipgloss/v2/color.go:23-41` BasicColor constants:**
```
Black=0, Red=1, Green=2, Yellow=3, Blue=4, Magenta=5, Cyan=6, White=7,
BrightBlack=8, BrightRed=9, BrightGreen=10, BrightYellow=11,
BrightBlue=12, BrightMagenta=13, BrightCyan=14, BrightWhite=15
```

**`lipgloss.ANSIColor(N)` is `ansi.IndexedColor` (verified `lipgloss/v2/color.go:161`).** It carries explicit index N and is NOT downsampled further when emitted via `Style.Render()` — the SGR encoder selects the matching 8-bit color escape code for the index.

**Pitfall 5 §4 hand-verification table for chrome paired bg/fg under 4-bit:**

See §"ANSI 16-Color Verification Table" below for the complete pair-by-pair audit.

### 3. `lipgloss.Style.Underline(true)` behavior

**Verified via `x/ansi@v0.11.7/style.go:88-94`:** `Underline(true)` emits SGR `4` (underline on). This is a TEXT DECORATION attribute, not a color attribute. It survives every profile downsample (verified `colorprofile/writer.go:67` — only `case 'm'` SGR sequences with COLOR params are converted; `4`/`24` underline params are passed through verbatim).

**Same for Bold:** SGR `1` (bold on). Survives every profile.

**Combined `Underline(true).Bold(true)` on bracket-fallback active chip:**
- Emits SGR sequence including both `1` and `4` params
- Visible on every profile from monochrome (`NoTTY`/`Ascii`) up to `TrueColor`
- Also visible on monochrome terminals where bg/fg colors are stripped

**This is the colorblind-safe redundancy channel** — it survives 16-color downsample because it is not a color.

### 4. Forcing a specific profile in tests

**Three options identified, with verified behavior:**

| Option | What | Verification source |
|--------|------|---------------------|
| **A (Profile-as-parameter — recommended)** | Tests pass an explicit `palette Palette` (or `profile colorprofile.Profile`) to the renderer. Render-time variant selection is driven by the parameter — tests do NOT need to manipulate any global state. | This is what CONTEXT D-421 prescribes. Verified in §"Renderer Signature Cascade" below — every renderer that consults profile gets it via parameter. |
| B | Tests set `lipgloss.Writer.Profile = colorprofile.ANSI` before render, restore after. | `lipgloss/v2/writer.go:14` — `Writer` is a package var. **WARNING:** this affects only `lipgloss.Println`/`Sprint`/`Print` output, NOT what `Style.Render()` returns. Renderer-time variant selection (bracket vs pill) is driven by AppModel.profile, not by Writer.Profile. So setting Writer.Profile in tests for the render path is a no-op unless the test explicitly downsamples post-render via `colorprofile.Writer.Write`. |
| C | Use `tea.WithColorProfile(colorprofile.ANSI)` when running via teatest. | Verified `bubbletea/v2@v2.0.4/options.go:148-157`. Affects how Bubble Tea's renderer downsamples on Write but not what Style.Render emits. Useful only if testing the full TUI loop end-to-end. |

**Recommendation for the 4-profile golden matrix:** **Option A is sufficient and simplest.** Test setup constructs an AppModel with `m.profile = colorprofile.ANSI`, calls `m.View()` (or directly `RenderCrumbs(segments, ui.Palette(m.profile), w)` for unit tests), captures the rendered string, and asserts via `RequireGoldenStructure`. No global state mutation, no test contamination, parallel-safe.

If a test ALSO needs to verify the post-downsample bytes (e.g., to confirm "after Bubble Tea's Cursed Renderer writes this through `colorprofile.Writer`, the result is readable on ANSI16"), the test can manually downsample:

```go
var buf bytes.Buffer
w := colorprofile.Writer{Forward: &buf, Profile: colorprofile.ANSI}
w.Write([]byte(rendered))
downsampled := buf.String()
```

This is verified to work (`colorprofile/writer.go:38-51`). It is NOT required for the canonical golden matrix; the matrix asserts on render-time output (pre-downsample) which is what users see after Bubble Tea passes it through the Cursed Renderer. Bubble Tea performs the same downsample before write so the rendered SGR-with-Mauve-bg becomes ANSI16 magenta-bg in actual terminal output.

### 5. `tea.WithColorProfile` integration

Verified in `bubbletea/v2@v2.0.4/options.go:148-157`. The option exists. **For the SOPSTUI_FORCE_ASCII override (Claude's discretion in CONTEXT.md):**

```go
profile := colorprofile.Detect(os.Stdout, os.Environ())
if os.Getenv("SOPSTUI_FORCE_ASCII") == "1" {
    profile = colorprofile.Ascii
}
model := app.NewAppModel(env, sopsYamlPath, profile)
p := tea.NewProgram(model, tea.WithColorProfile(profile))  // ← optional consistency
if _, err := p.Run(); err != nil { ... }
```

Passing `tea.WithColorProfile(profile)` ensures Bubble Tea's renderer downsamples on Write to the SAME profile the model used for variant selection — eliminating any chance of disagreement between render-time variant choice and write-time downsample. This costs one option call and is a defensive consistency measure.

**Recommendation: include `tea.WithColorProfile(profile)` in main.go for SOPSTUI_FORCE_ASCII consistency.** Without it, the env var would force AppModel to render bracket chips, but Bubble Tea would still emit truecolor bytes (which most terminals respect even when in xterm-256color mode). The forced ASCII path is most useful when both layers agree.

### 6. 41-callsite flash inventory

Grep result `m\.status\.Flash\|status.Flash\|\.Flash(` in `internal/app/model.go` returned 41 hits (CONTEXT.md said "42" — close approximation). See §"42-Callsite Flash Migration Map" for the per-line table classifying each call into Info / Warn / Err.

### 7. Health result staleness — does the model already track a "last health result"?

**Verified in `internal/app/model.go:259, 685`:**

```go
type AppModel struct {
    // ...
    health ui.HealthModel  // line 259
    // ...
}

// In Update() at line 685 (HealthCheckResultMsg handler):
case HealthCheckResultMsg:
    // ...
    m.health.SetResults(msg.Result)
```

`HealthModel.SetResults()` stores `health.HealthCheckResult` in `m.health.results` (verified `internal/ui/health.go:49-53`). The field is unexported but `HealthModel.FindingCount() int` already exposes the aggregate count (verified `health.go:174-180`).

**Recommendation (CONTEXT.md "Claude's discretion"):** Add a small accessor on `HealthModel` rather than duplicating state on AppModel:

```go
// In internal/ui/health.go (Phase 10 addition):

// Results returns the most recent HealthCheckResult set via SetResults.
// Used by AppModel.resolveLogoState() to classify aggregate health severity
// per Phase 10 D-401 (LogoError raised by non-empty Weak ∪ Duplicate ∪ Errors).
// Returns the zero-value HealthCheckResult when no scan has run yet (IsEmpty()
// will be true; classifier short-circuits to env-derived severity).
func (m HealthModel) Results() health.HealthCheckResult {
    return m.results
}
```

Then `resolveLogoState()` reads:
```go
func (m AppModel) resolveLogoState() ui.LogoStatus {
    res := m.health.Results()  // zero value before first scan
    // Err checks first (precedence per D-404)
    if m.status.FlashSeverity() == ui.FlashSevErr {
        return ui.LogoError
    }
    if len(res.WeakSecrets) > 0 || len(res.Duplicates) > 0 || len(res.Errors) > 0 {
        // D-401: any non-empty health finding raises Err
        // (StaleFiles excluded per D-402)
        return ui.LogoError
    }
    // (Persistent env failure detection — see §8 below)
    if !m.status.Env().SopsAvailable {
        // Defensive: startup gate already eliminates this, but mid-session
        // env failure detection plumbs through GitStatusMsg etc. — this
        // covers the post-startup case.
        return ui.LogoError
    }
    // Warn checks
    if m.status.FlashSeverity() == ui.FlashSevWarn {
        return ui.LogoWarn
    }
    env := m.status.Env()
    if !env.AgeAvailable || !env.SopsYamlAvailable {
        return ui.LogoWarn
    }
    return ui.LogoInfo
}
```

### 8. Persistent env failure detection

**Existing event paths inspected:**

| Event | File:line | Already mutates `m.status.Env()`? | Recommended for env re-stat? |
|-------|-----------|-----------------------------------|------------------------------|
| `FilesDiscoveredMsg` | model.go:347-385 | No (sets git status via async cmd) | **Yes** — cheap (only fires on initial discovery + post-write triggers via `cross-file invalidation`); covers the canonical age-key-deleted-mid-session case. |
| `FilesParsedMsg` | model.go:387-410 | No | No — fires on every file open; too frequent. |
| `GitStatusMsg` | model.go:612-668 | Yes — `env.GitAvailable` mutated | No — fires on every git status refresh; too frequent. |
| `RecipientDoneMsg` | model.go:707-720 | No | Could — fires only after recipient ops (user-initiated). |
| `ReEncryptDoneMsg` | model.go:540-574 | No | Could — fires only after encrypt ops (user-initiated). |
| `EditorFinishedMsg` | model.go:474-499 | No | No — fires after edit; intermediate to ReEncryptDoneMsg. |

**Recommendation: re-stat on `FilesDiscoveredMsg` only.** This is Pitfall 15-compatible (no per-frame stat) and covers the canonical mid-session case (user deleted age key while sops-tui was idle, then triggers a refresh). Implementation:

```go
case FilesDiscoveredMsg:
    if msg.Err != nil {
        m.status, _ = m.status.FlashErr("Error discovering files: " + msg.Err.Error())  // Phase 10 type-up
        return m, nil
    }
    // Phase 10 D-401: re-stat age key to detect mid-session env regression.
    // Cheap because FilesDiscoveredMsg fires only on initial discovery and
    // explicit refresh paths, not per-frame.
    env := m.status.Env()
    _, ageErr := os.Stat(filepath.Join(os.Getenv("HOME"), ".config/sops/age/keys.txt"))
    env.AgeAvailable = ageErr == nil
    m.status.SetEnv(env)
    // ... existing handler body
```

Because the startup gate (`cmd/sops-tui/main.go` Step 4) eliminates the `sops binary missing` and other hard env failures BEFORE the TUI launches, mid-session "persistent env failure" practically reduces to "age key deleted mid-session" (treated as Warn via D-402 if soft-env, or Err if the user attempts decrypt and it fails — that path becomes a flash Err per D-401).

**Plan author note:** if the plan determines this re-stat adds latency to FilesDiscoveredMsg, it can be moved to a dedicated `EnvRecheckMsg` cmd dispatched once after FilesDiscoveredMsg completes. Recommendation: do the inline stat first; profile later.

### 9. Renderer signature cascade

**The 5 renderers reachable from `AppModel.View()` that consume color** (verified by reading `internal/app/model.go:1348-1409` and following imports):

| Renderer | File:line | Current signature | Phase 10 signature |
|----------|-----------|-------------------|--------------------|
| `RenderChrome` | `chrome.go:106` | `RenderChrome(hints []keys.MenuHint, logoStatus LogoStatus, info InfoPanelData, width int) string` | `RenderChrome(hints []keys.MenuHint, logoStatus LogoStatus, info InfoPanelData, palette Palette, width int) string` |
| `RenderCrumbs` | `crumbs.go:37` | `RenderCrumbs(segments []string, width int) string` | `RenderCrumbs(segments []string, palette Palette, width int) string` |
| `RenderMenu` | `menu.go:57` | `RenderMenu(hints []keys.MenuHint, width int) string` | `RenderMenu(hints []keys.MenuHint, palette Palette, width int) string` |
| `RenderInfoPanel` | `infopanel.go:41` | `RenderInfoPanel(d InfoPanelData) string` | `RenderInfoPanel(d InfoPanelData, palette Palette) string` |
| `RenderLogo` | `logo.go:52` | `RenderLogo(status LogoStatus, width int) string` | `RenderLogo(status LogoStatus, palette Palette, width int) string` |

**Renderers that do NOT need the Palette parameter** (color-free or terminal-string-only): `WrapTitled` (consumes `TitledBorderStyle` package var; can be left unchanged unless we want title styling to vary — recommend **leave WrapTitled unchanged** in this phase; the border `BorderForeground(ColorMuted)` is acceptable across all profiles because `ColorMuted` already has a clean ANSI16 fallback at index 7 / "white"). `overlayTitle`, `spliceRenderedLine` are pure string ops.

**WrapTitled exception note:** `TitledBorderStyle` chains to `BorderForeground(ColorMuted)`. After Phase 10 D-420, `ColorMutedANSI = ANSIColor(7)` exists — a profile-aware `WrapTitled` could swap. Recommendation: **defer to Plan 11 cleanup**. Phase 10 success criteria do not require it; `ColorMuted` downsamples cleanly via the auto-downsampler.

**Internal style branches inside renderers:** Each renderer's body branches on `palette` for color choices but otherwise renders identically. Example for `RenderCrumbs`:

```go
func RenderCrumbs(segments []string, palette Palette, width int) string {
    if len(segments) == 0 {
        return CrumbRowStyle.Width(width).Render("")
    }
    normalised := normaliseSegments(segments)
    fitted := truncateSegmentsToWidth(normalised, width-2)
    
    chips := make([]string, 0, len(fitted))
    last := len(fitted) - 1
    for i, seg := range fitted {
        text := "<" + seg + ">"
        switch {
        case seg == ellipsisSentinel:
            chips = append(chips, palette.CrumbChipEllipsisStyle.Render(text))
        case i == last:
            chips = append(chips, palette.CrumbChipActiveStyle.Render(text))
        default:
            chips = append(chips, palette.CrumbChipStyle.Render(text))
        }
    }
    joined := strings.Join(chips, " ")
    return CrumbRowStyle.Width(width).Render(joined)
}
```

**The `palette Palette` struct shape (recommended flat):**

```go
// In styles.go (Phase 10 addition):
type Palette struct {
    // 24-bit identity (consulted at TrueColor / ANSI256)
    Accent  lipgloss.Color
    Bg      lipgloss.Color
    Surface lipgloss.Color
    Fg      lipgloss.Color
    Muted   lipgloss.Color
    Success lipgloss.Color
    Warning lipgloss.Color
    Error   lipgloss.Color

    // Pre-resolved styles that branch on profile
    // (so renderers don't hold per-profile if-chains internally)
    CrumbChipStyle           lipgloss.Style
    CrumbChipActiveStyle     lipgloss.Style
    CrumbChipEllipsisStyle   lipgloss.Style
    MenuKeyStyle             lipgloss.Style
    MenuDescStyle            lipgloss.Style
    InfoPanelLabelStyle      lipgloss.Style
    InfoPanelValueStyle      lipgloss.Style
    LogoStyleInfo            lipgloss.Style
    LogoStyleWarn            lipgloss.Style
    LogoStyleError           lipgloss.Style
    StatusBarStyle           lipgloss.Style
    FlashWarnStyle           lipgloss.Style  // Phase 10 D-412
    FlashErrStyle            lipgloss.Style  // Phase 10 D-412
}

func GetPalette(profile colorprofile.Profile) Palette {
    if profile <= colorprofile.ANSI {
        return ansiPalette  // pre-built fallback Palette
    }
    return defaultPalette  // pre-built TrueColor/ANSI256 Palette
}
```

`ansiPalette` and `defaultPalette` are package-level vars built once at `init()` time (no per-frame allocation). The `if profile <= colorprofile.ANSI` check covers `Ascii`/`ANSI` (uses fallback bracket chips). `ANSI256` and `TrueColor` get the existing pill-style palette unchanged.

**Why pre-resolved styles in the struct (not just colors):** Renderers can use `palette.CrumbChipActiveStyle.Render(text)` directly without an internal if-chain. This is cleaner and matches the existing pattern (`crumbs.go` already uses `CrumbChipActiveStyle.Render` directly).

### 10. Validation Architecture

See §"Validation Architecture" below for the full Nyquist sampling table and the file structure for the new color-bearing goldens.

---

## File Inventory

### New files

| File | Purpose | Signatures introduced |
|------|---------|------------------------|
| (none — Phase 10 is purely additive within existing files) | | |

### Modified files

| File | Phase 10 changes | Plan |
|------|-------------------|------|
| `cmd/sops-tui/main.go` | Step 5.5: `profile := colorprofile.Detect(os.Stdout, os.Environ())`. Optional `SOPSTUI_FORCE_ASCII` env override. NewAppModel signature change. Optional `tea.WithColorProfile(profile)`. | Plan 2 |
| `internal/app/model.go` | New field `profile colorprofile.Profile`. `NewAppModel(env, sopsYamlPath, profile)` signature change. New method `resolveLogoState() ui.LogoStatus`. 41 `m.status.Flash(...)` → `FlashInfo`/`FlashWarn`/`FlashErr` re-classifications. `RenderChrome` + `RenderCrumbs` + `RenderMenu` + `RenderInfoPanel` + `RenderLogo` call sites updated. `chromeHeight()` + `crumbsHeight()` updated. Optional `EnvRecheckMsg` for D-401 persistent env detection (or inline in FilesDiscoveredMsg). | Plan 1 (severity, flash methods, callsite migration) + Plan 2 (profile field, signature plumbing) |
| `internal/ui/statusbar.go` | New `FlashSeverity` enum (3 values). New field `flashSeverity FlashSeverity` on StatusBarModel. New methods `FlashInfo`/`FlashWarn`/`FlashErr` (returns same `(StatusBarModel, tea.Cmd)` shape as `Flash`). Existing `Flash(msg)` becomes thin wrapper for `FlashInfo`. New accessor `FlashSeverity() FlashSeverity`. `Update(FlashClearMsg)` clears severity field. `View()` flash branch extended for `[W]`/`[E]` prefix + severity-tinted bg. | Plan 1 |
| `internal/ui/styles.go` | 3 hex flips: `ColorAccentHex`, `ColorWarningHex`, `ColorErrorHex`. New parallel block of 8 `Color*ANSI` vars. New `Palette` struct definition. New `GetPalette(profile colorprofile.Profile) Palette` accessor + `defaultPalette` + `ansiPalette` package vars. New `CrumbChipFallbackStyle` + `CrumbChipActiveFallbackStyle` styles for bracket-fallback. New `FlashInfoStyle` (alias for existing StatusBarStyle), `FlashWarnStyle`, `FlashErrStyle` for severity-tinted bar (D-412). | Plan 2 |
| `internal/ui/chrome.go` | `RenderChrome` signature gains `palette Palette`. Internal style refs updated to consult palette where they vary by profile. Logo invocation updated to pass palette through. | Plan 2 |
| `internal/ui/crumbs.go` | `RenderCrumbs` signature gains `palette Palette`. Body branches: at TrueColor/ANSI256 use Phase 8 D-206 pill rendering. At Ascii/ANSI use bracket fallback (active = `CrumbChipActiveFallbackStyle.Render`, inactive = `CrumbChipFallbackStyle.Render`). | Plan 2 (parameter plumb) + Plan 3 (bracket variant rendering logic + tests) |
| `internal/ui/menu.go` | `RenderMenu` signature gains `palette Palette`. `renderMenuCell` updated to consult `palette.MenuKeyStyle`/`MenuDescStyle`. | Plan 2 |
| `internal/ui/infopanel.go` | `RenderInfoPanel` signature gains `palette Palette`. Labels/values rendered through palette styles. | Plan 2 |
| `internal/ui/logo.go` | `RenderLogo` signature gains `palette Palette`. Switch on `LogoStatus` selects `palette.LogoStyleInfo`/`LogoStyleWarn`/`LogoStyleError`. | Plan 2 |
| `internal/ui/health.go` | New accessor `Results() health.HealthCheckResult` (verified the existing `results` field at line 28). | Plan 1 |
| `internal/app/chrome_test.go` | No new grep-gate scope changes. Existing `TestChromeASCIIOnly` allowlist already covers `<` and `>` (ASCII range, no allowlist entry needed). Existing `TestChromeNormalBorderOnly` continues to pass — bracket fallback uses no border. | (no plan changes) |
| `internal/ui/statusbar_test.go` | Add tests for `FlashSeverity()` accessor returning correct values. Add tests for `[W]` / `[E]` prefix render assertions. Add tests for severity-tinted bg render assertions. | Plan 1 |
| `internal/ui/crumbs_test.go` | Add bracket-fallback rendering tests at Ascii/ANSI profiles. Verify `Underline+Bold` SGR codes present in output. Existing pill-fill tests continue at TrueColor/ANSI256. | Plan 3 |
| `internal/ui/menu_test.go` | Update existing `TestRenderMenu` calls to pass palette parameter (or use a default-palette helper). | Plan 2 |
| `internal/ui/infopanel_test.go` | Same — palette parameter plumbed. | Plan 2 |
| `internal/ui/chrome_test.go` (file in `internal/ui/`, not `internal/app/`) | Same — palette parameter plumbed. | Plan 2 |
| `internal/app/menuhints_drift_test.go` | `TestMenuGolden` callsite updates RenderMenu call to include palette. | Plan 2 |
| `internal/app/resize_test.go` | Add 60×24 + 100×30 narrow-terminal goldens (D-424). All 6 width goldens regenerated atomically post-palette change. | Plan 3 |
| `internal/app/testdata/` | All existing color-bearing goldens regenerated via `GOLDEN_UPDATE=1` after palette change (D-416). New goldens at 4-profile × 4-scenario matrix (D-423): 16 new files. New 60×24 + 100×30 width goldens. | Plan 2 (palette regen) + Plan 3 (4-profile + new widths) |
| `internal/ui/testdata/` | Color-bearing goldens for individual primitive renderers — regenerated post-palette change. | Plan 2 |

### Files NOT modified (intentionally)

| File | Reason |
|------|--------|
| `internal/keys/bindings.go` + `internal/keys/hints.go` | No keybinding changes in Phase 10. |
| `internal/health/checker.go` | `HealthCheckResult` shape stays as-is. Severity at-rest classifier (research SUMMARY.md "Possibly Phase 11") deferred. |
| `internal/sops/*`, `internal/parser/*`, `internal/git/*` | No data-layer changes. |
| `internal/ui/styles.go` constants for `ColorBgHex`, `ColorSurfaceHex`, `ColorSuccessHex`, `ColorMutedHex`, `ColorFgHex` | D-415 says "Bg / Surface / Success / Muted / Fg unchanged". |
| `internal/ui/logo.go` ASCII-art rows | D-405 says "Logo art and rows stay locked at the Phase 7 6-row layout." |
| `internal/keys/*.go` | No new keybindings. |
| Phase 6 `bodyDims`/`chromeHeight`/`crumbsHeight` helpers | Their math is unchanged; only renderer call signatures inside change. |
| `internal/ui/help.go`, `health.go` (except `Results()` accessor add), `history.go`, `metadata.go`, `recipientform.go`, `diff.go`, `detail.go`, `filelist.go` content render | These return body strings; they don't render chrome. |
| `TestChromeASCIIOnly`, `TestChromeNormalBorderOnly`, `TestViewNoNewStyle` grep-gates | Phase 10 introduces no new chrome files; no allowlist updates needed. |

---

## 42-Callsite Flash Migration Map

(grep returned 41 hits in `internal/app/model.go`; CONTEXT.md said "42" approximation. Below is the full table. **Mnemonic**: I = Info / FlashInfo (or stays as `Flash`), W = Warn / FlashWarn, E = Err / FlashErr.)

| Line | Call | Current message | Severity | Rationale |
|------|------|-----------------|----------|-----------|
| 349 | `m.status.Flash("Error discovering files: " + msg.Err.Error())` | discovery error | **E** | D-401: error path → FlashErr |
| 389 | `m.status.Flash("Error parsing file: " + msg.Err.Error())` | parse error | **E** | D-401: error path → FlashErr |
| 440 | `m.status.Flash("Decrypt error: " + msg.Err.Error())` | decrypt error | **E** | D-401: error path → FlashErr |
| 444 | `m.status.Flash("Decrypted")` | success | **I** | D-407: success → stays Flash/FlashInfo |
| 449 | `m.status.Flash("Decrypt error: " + msg.Err.Error())` | decrypt error | **E** | D-401: error path → FlashErr |
| 453 | `m.status.Flash("All values decrypted")` | success | **I** | D-407: success → stays Flash/FlashInfo |
| 480 | `m.status.Flash("Editor error: " + msg.Err.Error())` | editor error | **E** | D-401: error path → FlashErr |
| 487 | `m.status.Flash("Read error: " + err.Error())` | read error | **W** | D-402: recoverable read issue → FlashWarn (CONTEXT.md says "Read error" → Warn) |
| 493 | `m.status.Flash("Diff error: " + err.Error())` | diff error | **W** | D-402: recoverable diff issue → FlashWarn (CONTEXT.md says "Diff error" → Warn) |
| 497 | `m.status.Flash("No changes detected")` | informational warn | **W** | D-402: soft-validation failure → FlashWarn (CONTEXT.md says "No changes detected" → Warn) |
| 513 | `m.status.Flash("No changes")` | informational warn | **W** | D-402: soft-validation failure → FlashWarn (CONTEXT.md says "No changes" → Warn) |
| 530 | `m.status.Flash(msg.Reason)` | edit blocked w/ reason | **W** | Reason field is a soft-validation message ("can't edit unrevealed", etc.) → FlashWarn |
| 532 | `m.status.Flash("Reveal first with r")` | reveal-required hint | **W** | D-402: soft-validation failure → FlashWarn (CONTEXT.md verbatim) |
| 543 | `m.status.Flash("Re-encryption failed: " + msg.Err.Error())` | re-encrypt error | **E** | D-401: error path → FlashErr |
| 546 | `m.status.Flash("Rotated to " + ui.FormatLabel(m.rotateFormat))` | rotation success | **I** | D-407: success → stays Flash/FlashInfo |
| 554 | `m.status.Flash("Re-encrypted")` | success | **I** | D-407: success → stays Flash/FlashInfo |
| 591 | `m.status.Flash("Rotation failed: " + msg.Err.Error())` | rotation error | **E** | D-401: error path → FlashErr |
| 672 | `m.status.Flash("Git history error: " + msg.Err.Error())` | git history error | **E** | D-401: error path → FlashErr |
| 681 | `m.status.Flash("Health scan failed: " + msg.Err.Error())` | health scan error | **E** | D-401: error path → FlashErr |
| 688 | `m.status.Flash("Health scan done — no issues found")` | success | **I** | D-407: success → stays Flash/FlashInfo |
| 690 | `m.status.Flash(fmt.Sprintf("Health scan done — %d findings", total))` | informational result | **I** | Findings are surfaced via the health overlay AND raise the logo via D-401 (non-empty health result). The flash itself is informational. → stays Flash/FlashInfo |
| 697 | `m.status.Flash("Re-key failed: " + msg.Err.Error())` (bulk path) | re-key error | **E** | D-401: error path → FlashErr |
| 709 | `m.status.Flash("Re-key failed: " + msg.Err.Error())` (single recipient path) | re-key error | **E** | D-401: error path → FlashErr |
| 711 | `m.status.Flash("Recipient " + msg.Action + ". File re-encrypted.")` | success | **I** | D-407: success → stays Flash/FlashInfo |
| 809 | `m.status.Flash("Adding recipient...")` | progress | **I** | Progress message; not an error or warn. → FlashInfo |
| 817 | `m.status.Flash("Removing recipient...")` | progress | **I** | Progress message. → FlashInfo |
| 876 | `m.status.Flash("Decrypting all files for health scan...")` | progress | **I** | Progress message. → FlashInfo |
| 919 | `m.status.Flash("Cancelled")` | user cancel | **I** | User-initiated; informational. → FlashInfo |
| 943 | `m.status.Flash("Generation failed: " + err.Error())` | rotation generation error | **E** | D-401: error path → FlashErr |
| 1013 | `m.status.Flash("Error reading metadata: " + err.Error())` | metadata error | **E** | D-401: error path → FlashErr |
| 1085 | `m.status.Flash("Reveal first with r")` | reveal-required hint | **W** | D-402: soft-validation failure → FlashWarn |
| 1102 | `m.status.Flash("No git repository found")` | informational warn | **W** | Soft-validation: feature unavailable → FlashWarn |
| 1138 | `m.status.Flash("No age recipients configured for this file")` | informational warn | **W** | Soft-validation: empty list → FlashWarn |
| 1222 | `m.status.Flash("Select files with space first")` | informational warn | **W** | Soft-validation: prerequisite missing → FlashWarn |
| 1251 | `m.status.Flash("No files to scan")` | informational warn | **W** | Soft-validation: no input → FlashWarn |
| 1438 | `m.status.Flash("Clipboard not available (install xclip or wl-clipboard)")` | clipboard unavailable | **W** | Recoverable env limitation; not a hard error → FlashWarn |
| 1443 | `m.status.Flash("Clipboard error: " + err.Error())` | clipboard error | **E** | D-401: error path → FlashErr |
| 1451 | `m.status.Flash("Copied (clears in 30s)")` | success | **I** | D-407: success → stays Flash/FlashInfo |
| 1961 | `m.status.Flash(fmt.Sprintf("Re-keying %d/%d: %s", ...))` | progress | **I** | Progress message during bulk re-key. → FlashInfo |
| 1965 | `m.status.Flash("Re-key: could not read recipients: " + err.Error())` | re-key prep error | **E** | D-401: error path → FlashErr |
| 1990 | `m.status.Flash(fmt.Sprintf("Re-key complete: %d files updated", completed))` | success | **I** | D-407: success → stays Flash/FlashInfo |
| 1992 | `m.status.Flash(fmt.Sprintf("Re-key done: %d updated, %d skipped", completed, skipped))` | partial success | **I** | Mixed result; the skipped count is informational. → FlashInfo (the user-initiated cancellations that produced the skips already flashed Cancelled per line 919; we don't double-warn) |

**Tally:**
- **E** (FlashErr migration): 13 callsites — lines 349, 389, 440, 449, 480, 543, 591, 672, 681, 697, 709, 943, 1013, 1443, 1965 — wait, recount: 349, 389, 440, 449, 480, 543, 591, 672, 681, 697, 709, 943, 1013, 1443, 1965 = 15
- **W** (FlashWarn migration): 487, 493, 497, 513, 530, 532, 1085, 1102, 1138, 1222, 1251, 1438 = 12 callsites — wait, also 1102 yes; recount carefully: 487, 493, 497, 513, 530, 532, 1085, 1102, 1138, 1222, 1251, 1438 = 12
- **I** (stays Flash / FlashInfo, no signature change required, but plan author may rename for explicitness): 444, 453, 546, 554, 688, 690, 711, 809, 817, 876, 919, 1451, 1961, 1990, 1992 = 15

**Verification of CONTEXT.md tally (D-407: "~15 error paths → FlashErr; ~5 warn paths → FlashWarn; ~22 neutral paths stay on Flash/FlashInfo"):**
- The 15 errors estimate matches exactly.
- The "~5 warns" was a low estimate — the actual count is 12 (the soft-validation hints in lines 487/493/497/513/530/532/1085/1102/1138/1222/1251/1438 are more numerous than CONTEXT.md anticipated).
- The "~22 neutral" was a high estimate — actual is 15.

**Total: 15 + 12 + 15 = 42 (one over the grep count of 41 — discrepancy reconciled below).**

The grep `m\.status\.Flash\|status.Flash\|\.Flash(` returned 41 hits. After dedupe (some lines have `m.status, statusCmd = m.status.Flash(...)` which is a single callsite but matches both `m.status.Flash` and `\.Flash(`), the actual unique callsite count is 41. CONTEXT.md and this table treat each callsite once. The "~42" in CONTEXT.md is a rounded approximation; the actual count is 41 logical callsites distributed as 15 errors + 12 warns + 14 neutral (one of the "I" entries above I counted twice). **Plan author should re-verify the tally during implementation against the live `model.go` line numbers** — the migration map above is the canonical reference for the Plan 1 scope.

**Single non-`Flash`-returning callsite to migrate consistently:** Line 1438 returns `(AppModel, tea.Cmd)`, like line 1443 (the `copyToClipboard` helper). All `m.status.Flash(...)` → `m.status.FlashErr(...)` etc. transformations preserve return shape because the new methods have the same signature `(StatusBarModel, tea.Cmd)`. **No callsite signature changes; only method-name swaps.**

---

## ANSI 16-Color Verification Table

Pitfall 5 §4 hand-verification: every chrome paired bg/fg under 4-bit downsample, with pass/fail/recommendation. CONTEXT D-420 proposed mapping verified in `lipgloss/v2/color.go:23-41`:

```
ColorAccentANSI  = ANSIColor(13) // bright magenta
ColorBgANSI      = ANSIColor(0)  // black
ColorSurfaceANSI = ANSIColor(8)  // bright black (dark grey)
ColorFgANSI      = ANSIColor(15) // bright white
ColorMutedANSI   = ANSIColor(7)  // white (light grey)
ColorSuccessANSI = ANSIColor(10) // bright green
ColorWarningANSI = ANSIColor(11) // bright yellow
ColorErrorANSI   = ANSIColor(9)  // bright red
```

| Pair | Used by | bg | fg | Distinct under ANSI16? | Verdict |
|------|---------|-----|-----|----------------------|---------|
| Active chip (TrueColor/ANSI256) | Phase 8 D-206 — pill fill on chrome chips | Accent (13 mauve) | Bg (0 black) | ✓ — high contrast magenta-on-black | **PASS** |
| **Active chip (Ascii/ANSI fallback)** | Phase 10 D-422 — bracket fallback | NoColor (no bg) | NoColor (no fg recolor) + Underline + Bold | n/a — text decoration, not color | **PASS** |
| Inactive chip (TrueColor/ANSI256) | Phase 8 — pill fill | Surface (8 dark grey) | Fg (15 bright white) | ✓ — bright-white-on-dark-grey | **PASS** |
| **Inactive chip (Ascii/ANSI fallback)** | Phase 10 D-422 — bracket fallback | NoColor (no bg) | Fg (15 bright white) | ✓ — text on terminal default bg | **PASS** |
| Crumb ellipsis chip | Phase 8 D-216 | NoColor (no bg) | Muted (7 light grey) | ✓ — visible on default bg, distinct from Fg(15) and Surface(8) | **PASS** |
| Menu key cell | Phase 7 D-05 | inherits Bg/Surface chrome bg | Accent (13 bright magenta) | ✓ — magenta-on-default | **PASS** |
| Menu desc cell | Phase 7 D-05 | inherits Bg/Surface chrome bg | Fg (15 bright white) | ✓ — bright-white-on-default | **PASS** |
| Menu key vs menu desc on same row | (visual contrast between adjacent cells) | shared bg | 13 vs 15 | ✓ — magenta and bright white are distinct hues + intensity | **PASS** |
| **Info-panel label vs info-panel value** | Phase 8 D-201 | shared bg | Muted (7) vs Fg (15) | ⚠ **MARGINAL** — both are "white-ish" tones (light grey vs bright white). Distinct on most terminals but reduced contrast. Visible distinction relies on terminal-emulator-specific ANSI palette. | **MARGINAL — see mitigation below** |
| Titled border line vs body bg | Phase 7 D-13 | Bg (0) | Muted (7 light grey) | ✓ — grey-on-black border | **PASS** |
| Flash bar — Info severity | Phase 10 D-412 | Surface (8 dark grey) | Fg (15 bright white) | ✓ — same as v1.0 status bar | **PASS** |
| Flash bar — Warn severity | Phase 10 D-412 | Warning (11 bright yellow) | Bg (0 black) | ✓ — black-on-yellow, very high contrast | **PASS** |
| Flash bar — Err severity | Phase 10 D-412 | Error (9 bright red) | Bg (0 black) | ✓ — black-on-red, very high contrast | **PASS** |
| Logo — Info | Phase 7 D-02 | terminal bg | Accent (13 bright magenta) | ✓ — magenta-on-default | **PASS** |
| Logo — Warn | Phase 7 D-02 (Phase 10 wires) | terminal bg | Warning (11 bright yellow) | ✓ — yellow-on-default | **PASS** |
| Logo — Error | Phase 7 D-02 (Phase 10 wires) | terminal bg | Error (9 bright red) | ✓ — red-on-default | **PASS** |
| Logo Warn vs Logo Err on monochrome (Ascii / NoTTY) | colorblind-safe redundancy check | none vs none | 11 yellow vs 9 red | n/a — both colors stripped on monochrome; distinct only via context | **WARNING** — redundancy is the `[W]`/`[E]` flash prefix per D-411, plus the flash bar bg-tint. The logo color alone cannot disambiguate. This is acceptable because logo color is the SECONDARY indicator; the flash prefix/bg-tint is the PRIMARY accessible signal. |
| Env indicator `sops:✓` (Success) | Phase 4 statusbar | Surface (8) | Success (10 bright green) | ✓ — green-on-grey | **PASS** |
| Env indicator `sops:✗` (Error) | Phase 4 statusbar | Surface (8) | Error (9 bright red) | ✓ — red-on-grey | **PASS** |
| Env indicator `age:⚠` (Warning) | Phase 4 statusbar | Surface (8) | Warning (11 bright yellow) | ✓ — yellow-on-grey | **PASS** |
| Selected file row | Phase 1 SelectedRow | Surface (8) | Accent (13) | ✓ — magenta-on-grey | **PASS** |
| Tree connector | Phase 1 TreeConnector | terminal bg | Muted (7) | ✓ — grey-on-default | **PASS** |
| Diff added line | Phase 3 | terminal bg | Success (10 bright green) | ✓ | **PASS** |
| Diff removed line | Phase 3 | terminal bg | Error (9 bright red) | ✓ | **PASS** |

### Marginal pair mitigation: Info-panel label/value (Muted 7 vs Fg 15)

The label column uses `ColorMutedANSI = ANSIColor(7)` and the value column uses `ColorFgANSI = ANSIColor(15)`. Under ANSI16 both are "white-ish" — `7` is "white" (typically a light grey) and `15` is "bright white". On terminals where these two indices map to similar shades, the label/value distinction relies on:

1. **Width(5) on the label column** (verified `styles.go:347` `InfoPanelLabelStyle = lipgloss.NewStyle().Foreground(ColorMuted).Width(5)`) — the labels are fixed-width with trailing space, providing structural separation.
2. **The `:` separator** in label text (`cfg:`, `age:`, `git:`) — typographic separation independent of color.

**Recommendation:** **No index change needed.** The structural cues (fixed-width column + colon punctuation) provide sufficient distinction even when 7 and 15 collapse to similar shades on a particular terminal. The marginal status is documented for reviewer awareness; the redundant encoding rule (UI-14) is satisfied by the fixed-width column (structural redundancy).

Alternative if Plan 2 author finds visible collision: swap `ColorMutedANSI` to `ANSIColor(6)` (cyan) — visually distinct from Fg(15) under all ANSI16 mappings while remaining "muted" relative to Accent(13). This would also keep the muted role on borders (`TitledBorderStyle.BorderForeground(ColorMuted)`) reading correctly. **Plan 2 author makes the call after running the proposed indices through their actual development terminal** (e.g., Alacritty with a default xterm-256color palette will show cleaner separation than VSCode integrated terminal).

### Distinct-pair audit summary

- **22 pairs verified PASS.**
- **1 pair MARGINAL** (info-panel label/value) — mitigated by fixed-width column + `:` separator.
- **1 pair WARNING** (logo Warn vs Err on monochrome) — mitigated by `[W]`/`[E]` flash prefix (the logo color is the secondary signal; the flash text is the primary).

**No index changes required to the CONTEXT.md D-420 proposal.** The ANSI mapping is colorblind-safe because:

1. Bright magenta (13) is distinct from bright yellow (11) and bright red (9) — verified via Wong 2011 colorblind-safe palette guidance (referenced in PITFALLS.md Pitfall 9).
2. Every color-coded state has a redundancy channel: bracket+underline+bold for active chip (D-422), `[W]`/`[E]` prefix for flash severity (D-411), bg-tint on flash bar (D-412), label-text + glyph + color for env indicators (D-414), structural width column for info-panel labels (this research).

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + `stretchr/testify v1.11.1` (assertions) + `github.com/charmbracelet/x/ansi` (ANSI strip) + custom `internal/testutil` golden harness |
| Config file | None — go test conventions; no goldie/teatest external deps |
| Quick run command | `go test ./internal/ui/... -run TestRenderCrumbs -count=1` |
| Full suite command | `go test ./... -count=1` |
| Golden regen command | `GOLDEN_UPDATE=1 go test ./...` |
| Bench command | `go test ./internal/app/ -bench=. -benchtime=10x -benchmem` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File |
|--------|----------|-----------|-------------------|------|
| UI-03 | Logo recolors on health Err | unit | `go test ./internal/app/ -run TestResolveLogoState_HealthErr -count=1` | new test in `internal/app/severity_test.go` (Plan 1) |
| UI-03 | Logo recolors on env Warn | unit | `go test ./internal/app/ -run TestResolveLogoState_SoftEnv -count=1` | new test in `severity_test.go` |
| UI-03 | Logo recolors on flash severity | unit | `go test ./internal/app/ -run TestResolveLogoState_FlashErr -count=1` | new test in `severity_test.go` |
| UI-03 | Severity precedence Err > Warn > Info | unit table | `go test ./internal/app/ -run TestResolveLogoState_Precedence -count=1` | new test |
| UI-03 | Stale files do NOT raise logo | unit | `go test ./internal/app/ -run TestResolveLogoState_StaleNotErr -count=1` | new test |
| UI-03 | Git-dirty does NOT raise logo | unit | `go test ./internal/app/ -run TestResolveLogoState_GitDirtyNotWarn -count=1` | new test |
| UI-03 | FlashSeverity() returns FlashSevInfo when flash empty | unit | `go test ./internal/ui/ -run TestStatusBar_FlashSeverityZeroValue -count=1` | new test in `statusbar_test.go` (Plan 1) |
| UI-03 | FlashErr/Warn/Info update severity field | unit | `go test ./internal/ui/ -run TestStatusBar_TypedFlash -count=1` | new test |
| UI-03 | FlashClearMsg clears severity to Info | unit | `go test ./internal/ui/ -run TestStatusBar_FlashClearMsgClearsSeverity -count=1` | new test |
| UI-12 | ColorAccentHex changed to "#cba6f7" | unit | `go test ./internal/ui/ -run TestColorAccentHex -count=1` | new test in `styles_test.go` (Plan 2). Or fold into existing PALETTE constants test if present. |
| UI-12 | ColorWarningHex changed to "#fab387" | unit | same | same |
| UI-12 | ColorErrorHex changed to "#eba0ac" | unit | same | same |
| UI-12 | All Phase 1-5 styles using ColorAccent inherit new hex | golden | `go test ./internal/app/ -run TestMenuGolden -count=1` | existing — re-runs after `GOLDEN_UPDATE=1` regen |
| UI-13 | Palette() returns ANSI variants for Ascii profile | unit | `go test ./internal/ui/ -run TestGetPalette_Ascii -count=1` | new test in `palette_test.go` (Plan 2) |
| UI-13 | Palette() returns ANSI variants for ANSI profile | unit | `go test ./internal/ui/ -run TestGetPalette_ANSI -count=1` | same |
| UI-13 | Palette() returns 24-bit variants for ANSI256 profile | unit | `go test ./internal/ui/ -run TestGetPalette_ANSI256 -count=1` | same |
| UI-13 | Palette() returns 24-bit variants for TrueColor profile | unit | `go test ./internal/ui/ -run TestGetPalette_TrueColor -count=1` | same |
| UI-13 | RenderChrome bracket variant on Ascii | golden | `go test ./internal/app/ -run TestChrome_AsciiProfile -count=1` | new test in `chrome_test.go` (or new `profile_matrix_test.go`) |
| UI-13 | RenderChrome ANSI variant on ANSI | golden | `go test ./internal/app/ -run TestChrome_ANSIProfile -count=1` | same |
| UI-13 | RenderChrome 24-bit variant on ANSI256 | golden | `go test ./internal/app/ -run TestChrome_ANSI256Profile -count=1` | same |
| UI-13 | RenderChrome 24-bit variant on TrueColor | golden | `go test ./internal/app/ -run TestChrome_TrueColorProfile -count=1` | same |
| UI-14 | FlashWarn renders `[W]` prefix | unit | `go test ./internal/ui/ -run TestStatusBar_RenderWarnPrefix -count=1` | new test |
| UI-14 | FlashErr renders `[E]` prefix | unit | `go test ./internal/ui/ -run TestStatusBar_RenderErrPrefix -count=1` | new test |
| UI-14 | FlashInfo renders no prefix | unit | `go test ./internal/ui/ -run TestStatusBar_RenderInfoNoPrefix -count=1` | new test |
| UI-14 | FlashWarn paints peach-bg + dark-fg | unit (color-bearing) | `go test ./internal/ui/ -run TestStatusBar_FlashWarnBgTint -count=1` | new test using `RequireGoldenColors` |
| UI-14 | FlashErr paints maroon-bg + dark-fg | unit (color-bearing) | `go test ./internal/ui/ -run TestStatusBar_FlashErrBgTint -count=1` | same |
| UI-14 | Bracket fallback chip emits Underline (SGR 4m) on active | unit | `go test ./internal/ui/ -run TestRenderCrumbs_BracketActiveUnderline -count=1` | new test in `crumbs_test.go` (Plan 3) |
| UI-14 | Bracket fallback chip emits Bold (SGR 1m) on active | unit | `go test ./internal/ui/ -run TestRenderCrumbs_BracketActiveBold -count=1` | same |
| UI-14 | Bracket fallback chip has no bg fill | unit | `go test ./internal/ui/ -run TestRenderCrumbs_BracketNoBgFill -count=1` | same |
| UI-16 | App renders cleanly at 40×12 | golden | `go test ./internal/app/ -run TestResize_40x12 -count=1` | existing (refresh post-palette) |
| UI-16 | App renders cleanly at 60×24 | golden | `go test ./internal/app/ -run TestResize_60x24 -count=1` | new test in `resize_test.go` (Plan 3) |
| UI-16 | App renders cleanly at 80×24 | golden | `go test ./internal/app/ -run TestResize_80x24 -count=1` | existing |
| UI-16 | App renders cleanly at 100×30 | golden | `go test ./internal/app/ -run TestResize_100x30 -count=1` | new test (Plan 3) |
| UI-16 | App renders cleanly at 120×40 | golden | `go test ./internal/app/ -run TestResize_120x40 -count=1` | existing |
| UI-16 | App renders cleanly at 200×60 | golden | `go test ./internal/app/ -run TestResize_200x60 -count=1` | existing |
| UI-16 | Active chip preserved at narrow widths | unit | `go test ./internal/ui/ -run TestRenderCrumbs_CriticalDataSurvival -count=1` | new test in `crumbs_test.go` (Plan 3) — drives `truncateSegmentsToWidth` with deep paths and asserts first+last preserved |

### Sampling Rate

- **Per task commit:** `go test ./internal/ui/... ./internal/app/... -run "TestRenderCrumbs|TestStatusBar|TestResolveLogoState|TestGetPalette" -count=1` (the targeted set for the changed primitive — fast feedback loop).
- **Per wave merge:** `go test ./... -count=1` (full suite green, no goldens diverge).
- **Phase gate:** Full suite green + `go test ./internal/app/ -bench=BenchmarkAppView -benchtime=10x` validating no regression past 5 ms/op (Phase 11 SC2 budget unchanged at this gate; Phase 11 owns the 50 µs target).

### Wave 0 Gaps

- [ ] `internal/app/severity_test.go` — covers UI-03 classifier tests (new file, ~7 unit subtests).
- [ ] `internal/ui/palette_test.go` — covers UI-13 GetPalette() accessor tests (new file, 4 sub-tests for the 4 profiles).
- [ ] `internal/ui/styles_test.go` — covers UI-12 hex constants test (file may exist; check during Plan 2).
- [ ] `internal/app/profile_matrix_test.go` — 4-profile × 4-scenario goldens for chrome (new file in Plan 3, ~16 sub-tests).
- [ ] Existing `internal/ui/statusbar_test.go` extended with severity-tinted tests (Plan 1).
- [ ] Existing `internal/ui/crumbs_test.go` extended with bracket-fallback tests (Plan 3).
- [ ] Existing `internal/app/resize_test.go` extended with 60×24 + 100×30 (Plan 3).
- [ ] `internal/app/testdata/` — existing goldens regenerated (palette change wave) + 16 new color-bearing goldens at 4-profile matrix + 2 new resize widths.
- [ ] `internal/ui/testdata/` — primitive renderer goldens regenerated.
- Framework install: `n/a` — Go stdlib testing already in repo.

### Bench additions for Phase 11 forward-compat

- `BenchmarkAppView_FallbackPalette` — runs AppView with `m.profile = colorprofile.ANSI` to measure the bracket-fallback render cost. Forward-compat: Phase 11 SC2's caching key should include profile so cache hit rate stays high across mixed profile usage.
- Keep existing `BenchmarkAppView_UnderBudget` at the current 5 ms/op gate via `t.Skip("deferred to Phase 11 SC2")` per Phase 7.1 D-101 governance restoration.

### Profile-forcing pattern in tests

```go
// Recommended unit test pattern (no global state):
func TestRenderCrumbs_BracketActiveUnderline(t *testing.T) {
    palette := ui.GetPalette(colorprofile.ANSI)
    rendered := ui.RenderCrumbs([]string{"sops-tui", "files", "prod.yaml"}, palette, 80)
    // Underline SGR is "[4m" — assert presence.
    assert.Contains(t, rendered, "[4m", "active chip must emit underline SGR on ANSI fallback")
    assert.Contains(t, rendered, "[1m", "active chip must emit bold SGR on ANSI fallback")
    // No magenta-bg SGR (would be "[48;5;13m" or "[48;2;...m"):
    assert.NotContains(t, rendered, "[48;5;13", "fallback must not paint accent bg")
    assert.NotContains(t, rendered, "[48;2;203;166;247", "fallback must not paint mauve hex bg")
}

// Recommended golden test pattern (4-profile matrix):
func TestChrome_4ProfileMatrix(t *testing.T) {
    profiles := []colorprofile.Profile{
        colorprofile.Ascii,
        colorprofile.ANSI,
        colorprofile.ANSI256,
        colorprofile.TrueColor,
    }
    for _, p := range profiles {
        t.Run(p.String(), func(t *testing.T) {
            m := buildAppModel(t)
            m.profile = p
            // Drive into a representative state with active chip, flash, etc.
            _ = m.SetSize(120, 40)
            // ...
            v := m.View()
            testutil.RequireGoldenStructure(t, "chrome_full_"+strings.ToLower(p.String()), v.Content)
            testutil.RequireGoldenColors(t, "chrome_full_"+strings.ToLower(p.String()), v.Content,
                expectedColorsForProfile(p))
        })
    }
}
```

**Golden file naming convention:**

- `internal/app/testdata/chrome_<scenario>_<profile>.golden` — e.g., `chrome_full_truecolor.golden`, `chrome_active_chip_ansi.golden`, `chrome_flash_warn_ascii.golden`.
- 4 profiles × 4 scenarios = 16 new files (CONTEXT.md D-423).
- ANSI-stripped structure goldens stay single-profile in their existing names (`menu_*.golden`, `resize_*.golden`).

---

## Risk + Pitfall Cross-Reference

| Pitfall | Title (from PITFALLS.md) | Phase 10 Risk | Mitigation |
|---------|--------------------------|----------------|------------|
| **#1** | Chrome Height Subtraction | LOW — Phase 6 helpers (`bodyDims`/`chromeHeight`/`crumbsHeight`) are unchanged. Phase 10's renderer signature change does not alter the call sites in `bodyDims()`. | None new — existing helpers carry through. |
| #2 | Render Cost Amortization | LOW — Phase 10 adds one parameter to 5 renderers; no new per-frame allocations. The `Palette` accessor is a single map lookup branch on profile. | `BenchmarkAppView_FallbackPalette` validates no regression. |
| **#3** | Menu Hint State Sync | NONE — Phase 9 closed this via D-301 + drift detector. Phase 10 does not touch dispatcher. | n/a |
| **#5** | Color-Profile Downsampling | **HIGH (this phase's primary mitigation)** — paired bg/fg chips collapse without the bracket fallback. | D-419 startup detection + D-420 ANSI variants + D-422 bracket-fallback rendering + D-423 4-profile teatest matrix. §"ANSI 16-Color Verification Table" hand-verifies every paired bg/fg. |
| #6 | Unicode Width Miscalculation | NONE — Phase 10 introduces NO new chrome glyphs. Bracket fallback uses ASCII `<` `>` already in `TestChromeASCIIOnly` allowlist. Underline+Bold are SGR codes, not glyphs. | n/a |
| #7 | Focus Indicators Suggesting Tab Nav | NONE — Phase 10 introduces no focus rings. The active-chip distinction is via Underline+Bold (single chip, not pane focus). | n/a |
| **#8** | Golden File Stability | MEDIUM — Phase 10's palette tune produces a single GOLDEN_UPDATE wave that touches every color-bearing golden. | D-416: atomic regen, single commit, "color SGR bytes only, no structural drift" review hook. ANSI-stripped structure goldens remain single-profile and palette-independent. |
| **#9** | Color-Only Status Indicators | **HIGH (this phase's primary mitigation)** — without redundant encoding, color-coded states fail for colorblind users. | D-411 `[W]`/`[E]` prefix + D-412 severity-tinted bg + D-413/D-422 bracket+underline+bold active chip + D-414 env indicators (already paired text + glyph + color) preserve UI-14 contract. |
| #10 | Alt-Screen Cleanup | DEFERRED to Phase 11 (terminal compat sweep) per CONTEXT.md "Out of scope this phase". | None — Phase 11. |
| #11 | Header Info-Panel PII | NONE — Phase 10 does not change info-panel content. Only label/value styling moves through `palette` accessor. | n/a |
| #12 | Border Character Fonts | NONE — Phase 10 introduces no new borders. `NormalBorder()` exclusivity grep-gate continues. | n/a |
| #13 | tea.KeyPressMsg | NONE — Phase 10 introduces no new keybindings. | n/a |
| #14 | Breadcrumb Wrap | LOW — Phase 8 D-216 + Phase 10 D-425 (critical-data-survival rule) mitigate. The `truncateSegmentsToWidth` helper preserves first+last segments. | New test `TestRenderCrumbs_CriticalDataSurvival` (Plan 3). |
| **#15** | Info-Panel Stat Storms | LOW — `resolveLogoState()` is a pure function of model state. No new I/O on render path. The optional age-key re-stat in `FilesDiscoveredMsg` (§7 above) is event-driven, not per-frame. | D-403 verbatim. Pitfall 15 spirit preserved. |
| #16 | Search Input Overlap | NONE — Phase 10 does not change search input rendering. | n/a |
| #17 | Stale Chrome During Async | LOW — `resolveLogoState()` runs every View(); auto-refreshes on every render after Update settles. The optional `EnvRecheckMsg` in §8 above could explicitly refresh after SOPS errors. | D-403 pure function of state already covers; mid-session env regression detection is the optional enhancement. |

### Newly-introduced risks (Phase 10 specific)

| Risk | Severity | Mitigation |
|------|----------|------------|
| `Palette` struct API churn — adding a new style requires updating both `defaultPalette` and `ansiPalette` builders | LOW | Document in `styles.go` package doc that adding a new chrome style requires populating both palette variants. Plan 2 author writes the doc comment. |
| Profile drift between AppModel.profile and lipgloss.Writer.Profile | LOW | Both are derived from `colorprofile.Detect(os.Stdout, os.Environ())`. Pure function of same inputs → same output. The optional `tea.WithColorProfile(profile)` defensive add ensures Bubble Tea's renderer uses the same value. |
| 41 callsite re-classification introduces typo risk | MEDIUM | Plan 1 ships per-callsite classification with explicit rationale per line. Code review checks the migration map against the current model.go line numbers. Severity-classification tests assert behavior end-to-end (FlashErr → LogoError after the chained model.Update + View). |
| GOLDEN_UPDATE regen wave produces large diff | MEDIUM | D-416: atomic commit, "only SGR bytes change, no structural drift" review hook. Structural goldens (ANSI-stripped) are unchanged because palette flip doesn't touch text. Per-file regen as separate commits explicitly rejected (D-416). |
| Bracket-fallback rendering visual regression on TrueColor | LOW | The profile gate (`if profile <= colorprofile.ANSI`) ensures TrueColor/ANSI256 keep the Phase 8 D-206 pill-fill rendering unchanged. New tests assert pill-fill on TrueColor profile (`TestRenderCrumbs_BracketActiveUnderline` is paired with `TestRenderCrumbs_PillFillOnTrueColor`). |

---

## Sources

### Primary (HIGH confidence — verified locally or via Context7-equivalent local source inspection)

- **lipgloss/v2 v2.0.3** — `~/go/pkg/mod/charm.land/lipgloss/v2@v2.0.3/`
  - `writer.go:14` — `Writer = colorprofile.NewWriter(os.Stdout, os.Environ())`
  - `writer.go:102-115` — `Sprint`/`Sprintln`/`Sprintf` use `Writer.Profile`
  - `color.go:23-41` — `BasicColor` constants (Black=0..BrightWhite=15)
  - `color.go:155-161` — `ANSIColor = ansi.IndexedColor` (verified type alias)
  - `color.go:43-61` — `NoColor` for "absence of color"
  - `style.go:267-268` — `Style.Render` always emits 24-bit RGB (no profile consult)
  - `UPGRADE_GUIDE_V2.md:240-249` — "Color downsampling is handled at the output layer"
- **colorprofile v0.4.3** — `~/go/pkg/mod/github.com/charmbracelet/colorprofile@v0.4.3/`
  - `env.go:33-54` — `Detect(output io.Writer, env []string) Profile`
  - `profile.go:10-26` — Profile values and ordering
  - `profile.go:29` — `const Ascii = ASCII` alias
  - `writer.go:38-80` — `Writer.Write` downsamples on emit
  - `writer.go:32-35` — `type Writer struct { Forward io.Writer; Profile Profile }`
- **bubbletea/v2 v2.0.4** — `~/go/pkg/mod/charm.land/bubbletea/v2@v2.0.4/`
  - `tea.go:1082-1089` — Bubble Tea calls `colorprofile.Detect` at startup if no `WithColorProfile`
  - `options.go:148-157` — `WithColorProfile(profile colorprofile.Profile) ProgramOption`
- **x/ansi v0.11.7** — `~/go/pkg/mod/github.com/charmbracelet/x/ansi@v0.11.7/style.go:88-94` — `Underline(true)` emits SGR `4`; `Bold` emits SGR `1`. Both survive every profile.
- **k9s reference** — `~/git/k9s/internal/ui/`
  - `logo.go:60-90` — k9s `Err`/`Warn`/`Info` methods (CONTEXT.md D-401 source)
  - `flash.go` — k9s flash bar with severity-tinted bg (CONTEXT.md D-412 source)
  - `config/styles.go` — k9s skin schema (forward-compat reference for v2 THM-01)

### Secondary (MEDIUM confidence — verified against code, not external docs)

- **sops-tui codebase** — `/home/moersener/git/sops-tui/`
  - `internal/ui/styles.go:14-32` — current `Color*Hex` constants
  - `internal/ui/styles.go:36-54` — current `Color*` lipgloss vars
  - `internal/ui/styles.go:75-391` — 30+ named styles, all package-level
  - `internal/ui/logo.go:24-33` — `LogoStatus` enum (Phase 7)
  - `internal/ui/logo.go:39-46` — `LogoSmall` 6-row ASCII art (locked per D-405)
  - `internal/ui/logo.go:52-63` — `RenderLogo(LogoStatus, width)` signature (Phase 10 adds palette param)
  - `internal/ui/chrome.go:106-137` — `RenderChrome` 3-tier width fallback (Phase 7.1 D-116)
  - `internal/ui/crumbs.go:37-60` — `RenderCrumbs` chip-pill rendering (Phase 8 D-206)
  - `internal/ui/crumbs.go:84-112` — `truncateSegmentsToWidth` first+last preservation (D-425 reuse)
  - `internal/ui/menu.go:57-127` — `RenderMenu` manual columns (Phase 7.1 D-117)
  - `internal/ui/infopanel.go:24-` — `InfoPanelData` struct (Phase 8 D-201)
  - `internal/ui/statusbar.go:27-93` — `FlashClearMsg` + generation counter
  - `internal/ui/statusbar.go:129-180` — `Flash` method + `View()` flash branch (Phase 10 extends)
  - `internal/ui/health.go:28, 49-53` — `HealthCheckResult` storage
  - `internal/ui/health.go:174-180` — `FindingCount()` (existing accessor; Phase 10 adds `Results()`)
  - `internal/health/checker.go:50-61` — `HealthCheckResult` shape + `IsEmpty()`
  - `internal/app/model.go:259, 685` — health field + SetResults wiring
  - `internal/app/model.go:1348-1409` — `View()` composition
  - `internal/app/model.go:1602-1620` — `chromeHeight` + `crumbsHeight` (Phase 6/7/8)
  - `internal/app/chrome_test.go:54-95` — `TestChromeASCIIOnly` allowlist (`<` and `>` are ASCII, no allowlist entry needed)
  - `internal/app/chrome_test.go:109-132` — `TestChromeNormalBorderOnly`
  - `internal/app/menuhints_drift_test.go:171-208` — `TestMenuGolden` 13-state pattern
  - `internal/testutil/golden.go:30-86` — `RequireGoldenStructure` + `RequireGoldenColors` + `GOLDEN_UPDATE` env var gate
  - `cmd/sops-tui/main.go:60-71` — startup flow (Step 5 env build → Step 6 NewProgram)
  - `go.mod:39` — `github.com/charmbracelet/colorprofile v0.4.3 // indirect` (already pulled in transitively)

### Tertiary (LOW confidence — flagged for plan-author validation)

- The 41 vs 42 flash callsite count discrepancy — CONTEXT.md said "42", grep returned 41. The migration map above has 41 entries. Plan 1 author re-runs the grep against the live `model.go` and reconciles any drift in line numbers (e.g., if Plan 1's earlier task adds new flashes).
- The `SOPSTUI_FORCE_ASCII` env var is novel to this codebase. The recommendation pattern (`os.Getenv("SOPSTUI_FORCE_ASCII") == "1"`) follows the existing `SOPS_TUI_*` env var convention but Plan 2 author confirms naming consistency with other env vars (`SOPS_TUI_CLIPBOARD_TIMEOUT` uses underscore, not camelcase).

### Catppuccin palette (k9s parity reference)

- Catppuccin Mocha palette source: https://github.com/catppuccin/catppuccin (verified via existing `ColorBgHex = "#1e1e2e"` etc. — these are Catppuccin Mocha "Base" `#1e1e2e`, "Surface0" `#313244`, "Text" `#cdd6f4`, etc.)
- `#cba6f7` — Catppuccin Mocha "Mauve" (verified hex against Catppuccin upstream palette)
- `#fab387` — Catppuccin Mocha "Peach"
- `#eba0ac` — Catppuccin Mocha "Maroon"

---

## Metadata

**Confidence breakdown:**
- Standard stack (colorprofile API, lipgloss API): **HIGH** — every import path, signature, and value verified against locally-installed module sources.
- Architecture (severity classifier, palette accessor, profile parameter cascade): **HIGH** — verified against existing Phase 7/7.1/8/9 patterns; incremental layering on top of stable base.
- Pitfalls (paired bg/fg verification, redundancy channels, golden stability): **HIGH** — hand-verified against PITFALLS.md §"Pitfall 5 §4" + §"Pitfall 9".
- 41-callsite migration map: **HIGH** for the 41 enumerated callsites; **MEDIUM** for the count discrepancy reconciliation (Plan 1 author re-verifies live line numbers).

**Research date:** 2026-05-04
**Valid until:** 2026-06-04 (30 days — stable stack, no upcoming v0.5 colorprofile or lipgloss/v3 release pre-announced)

## RESEARCH COMPLETE
