---
phase: 10
plan: 02
subsystem: theming-accessibility
tags: [theming, accessibility, palette, colorprofile, ansi-fallback, signature-cascade]
requires:
  - "Phase 10 Plan 01: FlashSeverity enum + FlashWarnBarStyle/FlashErrBarStyle named-var indirection (auto-pick-up the new Peach/Maroon hex)"
  - "Phase 8 D-206 pill-fill chip rendering on RenderCrumbs (preserved verbatim in Plan 2)"
  - "Phase 7 LogoStyle{Info,Warn,Error} package vars (auto-pick-up the new Mauve/Peach/Maroon)"
provides:
  - "internal/ui/styles.go: 3 hex flips (#cba6f7 Mauve / #fab387 Peach / #eba0ac Maroon) + 8 Color*ANSI lipgloss.ANSIColor fallback variants + Palette struct + PaletteFor accessor"
  - "internal/ui/chrome.go: RenderChrome(hints, logoStatus, info, palette, width) signature"
  - "internal/ui/crumbs.go: RenderCrumbs(segments, palette, width) signature with `_ = palette` forward-compat seam for Plan 3 bracket-fallback branch"
  - "internal/ui/menu.go: RenderMenu(hints, palette, width) signature with `_ = palette` seam"
  - "internal/ui/infopanel.go: RenderInfoPanel(d, palette) signature with `_ = palette` seam"
  - "internal/app/model.go: AppModel.profile + AppModel.palette read-only fields; NewAppModel takes profile colorprofile.Profile; all 4 renderer callsites forward m.palette"
  - "cmd/sops-tui/main.go: colorprofile.Detect(os.Stdout, os.Environ()) at startup + SOPSTUI_FORCE_ASCII env override + tea.WithColorProfile(profile) on tea.NewProgram"
  - "go.mod: github.com/charmbracelet/colorprofile v0.4.3 promoted from indirect to direct require"
  - "internal/ui/hex_helpers_test.go: hexToRGBTriplet / hexBgSGR / hexFgSGR test helpers (D-417 — defense against future palette tunes)"
affects:
  - "Plan 10-03 will consume palette.Fallback inside RenderCrumbs to switch chip rendering to bracket+underline+bold on Ascii/ANSI profiles (D-422). The `_ = palette` discard line is the cleanup target."
  - "Plan 10-03 will add 4-profile teatest matrix (Ascii/ANSI/ANSI256/TrueColor) verifying SGR downsample correctness (D-423)."
  - "Plan 10-03 may add 60×24 + 100×30 narrow-terminal goldens validating bracket-chip rendering."
  - "Phase 11 verification will UAT the Catppuccin Mauve/Peach/Maroon palette under user terminal capability."
tech-stack:
  added:
    - "github.com/charmbracelet/colorprofile (already pulled transitively via lipgloss/v2; promoted to direct require)"
  patterns:
    - "Field type image/color.Color (the standard library interface that lipgloss.Color and lipgloss.ANSIColor both satisfy) for the 8 Palette color fields — single struct holds either 24-bit or 16-color variant set without conversion"
    - "PaletteFor(profile) accessor as the single switch point — chrome renderers consult Palette fields, never colorprofile.Profile directly"
    - "Renderer signature cascade with `_ = palette` forward-compat seams (Plan 2 plumbs the parameter; Plan 3 wires the fallback rendering)"
    - "Construction-time palette computation: AppModel.palette = ui.PaletteFor(profile) computed once at NewAppModel; never re-detected (Pitfall 15 spirit)"
    - "Profile-detection-at-startup pattern: colorprofile.Detect once + tea.WithColorProfile(profile) so render-time variant choice and write-time downsample agree"
    - "hexToRGBTriplet test helper (D-417) — color-presence assertions derive RGB triplets from named Color*Hex constants instead of hardcoding literals"
    - "Helper-on-AppModel pattern via buildAppModel(t) and newCleanAppModel(t) helpers — single update point cascades into 13+ test functions"
    - "PaletteFor (verb form) accessor name to avoid collision with Palette struct type (per pattern-mapper deviation #4)"
key-files:
  created:
    - "internal/ui/hex_helpers_test.go (hexToRGBTriplet / hexBgSGR / hexFgSGR — defense against future palette tunes)"
  modified:
    - "internal/ui/styles.go (Task 1: 3 hex flips + 8 ANSI vars + Palette struct + PaletteFor accessor)"
    - "internal/ui/styles_test.go (Task 1: 9 new tests — Catppuccin lock + unchanged constants regression guard + ANSI indices + Palette struct fields + 5 PaletteFor profile cases)"
    - "internal/ui/chrome.go (Task 2: RenderChrome signature + body callsite forwarding to RenderMenu / RenderInfoPanel)"
    - "internal/ui/crumbs.go (Task 2: RenderCrumbs signature + _ = palette forward-compat seam)"
    - "internal/ui/menu.go (Task 2: RenderMenu signature + _ = palette seam)"
    - "internal/ui/infopanel.go (Task 2: RenderInfoPanel signature + _ = palette seam)"
    - "internal/ui/chrome_test.go (Task 2: 5 RenderChrome callsites + 1 RenderMenu callsite updated)"
    - "internal/ui/crumbs_test.go (Task 2: 8 RenderCrumbs callsites updated; Task 5: 3 hex-literal-using tests migrated to hexToRGBTriplet)"
    - "internal/ui/menu_test.go (Task 2: 10 RenderMenu callsites updated; Task 5: TestRenderMenu_AccentAppliedToMnemonic migrated)"
    - "internal/ui/infopanel_test.go (Task 2: 6 RenderInfoPanel callsites updated)"
    - "internal/ui/logo_test.go (Task 5: TestRenderLogo_AllStatusVariants migrated to hexToRGBTriplet)"
    - "internal/ui/statusbar_test.go (Task 5: 3 SGR-byte tests migrated to hexBgSGR/hexFgSGR — Plan 1 forward deviation #1 closed)"
    - "internal/app/model.go (Task 3: AppModel struct gains profile + palette fields; NewAppModel signature change; 4 renderer callsites forward m.palette)"
    - "internal/app/bench_test.go (Task 3: 1 NewAppModel callsite)"
    - "internal/app/chrome_test.go (Task 3: 2 NewAppModel callsites + 1 RenderChrome callsite)"
    - "internal/app/hints_test.go (Task 3: buildAppModel(t) helper update — cascades into 16 dispatcher tests + 13 menuhints_drift_test.go sub-tests)"
    - "internal/app/layout_test.go (Task 3: 3 NewAppModel callsites)"
    - "internal/app/menuhints_drift_test.go (Task 3: 1 RenderMenu callsite)"
    - "internal/app/model_clipboard_test.go (Task 3: 2 NewAppModel callsites)"
    - "internal/app/model_reveal_test.go (Task 3: 5 NewAppModel callsites)"
    - "internal/app/model_test.go (Task 3: 19 NewAppModel callsites)"
    - "internal/app/resize_test.go (Task 3: 4 NewAppModel callsites)"
    - "internal/app/severity_test.go (Task 3: newCleanAppModel(t) helper update — cascades into 12 classifier tests)"
    - "cmd/sops-tui/main.go (Task 4: colorprofile.Detect + SOPSTUI_FORCE_ASCII override + NewAppModel arg + tea.WithColorProfile)"
    - "go.mod (Task 4: colorprofile promoted from indirect to direct require via go mod tidy)"
decisions:
  - "PaletteFor (verb form) accessor name avoids collision with Palette struct type — per pattern-mapper deviation #4 (D-421 in CONTEXT.md said Palette(profile) which would shadow the type)"
  - "Field type image/color.Color (interface) for Palette's 8 color fields rather than lipgloss.Color (which is a func not a type in lipgloss/v2) — single struct holds either variant set without conversion"
  - "SOPSTUI_FORCE_ASCII recognises ANY non-empty string (not just '1') for forgiving UX — answers terminal mis-detection support questions cheaply"
  - "Goldens did NOT need refresh — testutil/golden.go uses ansi.Strip (structural-only); the palette flip changes only color SGR bytes and resize_test.go callsites pass nil for wantColors. Task 6 GOLDEN_UPDATE pass produced zero file modifications, confirming the design intent."
  - "tea.WithColorProfile(profile) on NewProgram aligns Bubble Tea's Cursed Renderer downsample with AppModel's variant selection — eliminates render-time vs write-time disagreement (RESEARCH.md §5)"
  - "_ = palette discard line in RenderCrumbs / RenderMenu / RenderInfoPanel bodies — forward-compat seam Plan 3 will remove when wiring the fallback rendering body"
metrics:
  duration_minutes: 25
  date_completed: "2026-05-04"
  tasks_total: 6
  tasks_completed: 6
  files_changed: 25
  files_created: 1
---

# Phase 10 Plan 02: Palette Tune + 16-Color Profile Fallback Infrastructure Summary

Landed the Catppuccin Mauve/Peach/Maroon palette tune toward k9s
hot-pink/purple register; declared 8 ANSI fallback variants, the Palette
struct, and the PaletteFor accessor; cascaded a `palette ui.Palette`
parameter through 4 chrome renderers; added profile detection at startup
with `tea.WithColorProfile` consistency; promoted colorprofile to a
direct require; migrated 5 hex-literal-using tests + 3 statusbar SGR
assertions to constant-derived triplets via the new `hexToRGBTriplet`
helper; and atomically verified zero golden file churn.

## Commits

| Hash      | Task | Type | Subject                                                                  |
| --------- | ---- | ---- | ------------------------------------------------------------------------ |
| `74dc324` | 1    | feat | flip palette + add ANSI variants + Palette/PaletteFor                    |
| `b6b6688` | 2    | feat | cascade Palette parameter through chrome renderers                       |
| `69931f9` | 3    | feat | wire profile + palette through AppModel                                  |
| `3db3e77` | 4    | feat | detect color profile + plumb tea.WithColorProfile                        |
| `9d5f92e` | 5    | test | migrate hex-literal SGR assertions to constant-derived                   |
| `00d02fe` | 6    | fix  | replace em-dash with hyphen-minus in chrome.go doc comment               |

## What Landed

### Task 1 — Palette tune + ANSI variants + Palette / PaletteFor (commit `74dc324`)

**Hex flips in `internal/ui/styles.go` (D-415):**

| Constant         | Before    | After (Plan 10-02) | Catppuccin Name |
| ---------------- | --------- | ------------------ | --------------- |
| `ColorAccentHex` | `#89b4fa` | `#cba6f7`          | Mauve           |
| `ColorWarningHex`| `#f9e2af` | `#fab387`          | Peach           |
| `ColorErrorHex`  | `#f38ba8` | `#eba0ac`          | Maroon          |

The 5 unchanged constants (`ColorBgHex` `#1e1e2e`, `ColorSurfaceHex`
`#313244`, `ColorSuccessHex` `#a6e3a1`, `ColorMutedHex` `#6c7086`,
`ColorFgHex` `#cdd6f4`) are byte-identical to v1.0. Phase 1 named-var
indirection means the 8 derived `Color*` lipgloss vars + every named
style (`BreadcrumbActive`, `MenuKeyStyle`, `LogoStyleInfo`,
`BadgeUnencrypted`, `FlashWarnBarStyle`, `FlashErrBarStyle`, etc.)
inherits the new hex automatically.

**8 `Color*ANSI` ANSI fallback variants (D-420):**

| Variable           | Index | ANSI Color Name      |
| ------------------ | ----- | -------------------- |
| `ColorAccentANSI`  | 13    | Bright Magenta       |
| `ColorBgANSI`      | 0     | Black                |
| `ColorSurfaceANSI` | 8     | Bright Black (grey)  |
| `ColorFgANSI`      | 15    | Bright White         |
| `ColorMutedANSI`   | 7     | White (light grey)   |
| `ColorSuccessANSI` | 10    | Bright Green         |
| `ColorWarningANSI` | 11    | Bright Yellow        |
| `ColorErrorANSI`   | 9     | Bright Red           |

Each index hand-verified in 10-RESEARCH.md §"ANSI 16-Color Verification
Table"; every chrome bg/fg pair is distinct under 4-bit downsampling.

**Palette struct + PaletteFor accessor (D-421):**

```go
type Palette struct {
    Accent, Bg, Surface, Fg, Muted, Success, Warning, Error color.Color
    Fallback bool
}

func PaletteFor(profile colorprofile.Profile) Palette {
    if profile <= colorprofile.ANSI {
        return Palette{...ANSI variants..., Fallback: true}
    }
    return Palette{...24-bit variants..., Fallback: false}
}
```

The `Fallback` bool is `true` on `Ascii` / `ANSI` / `NoTTY` profiles
(the gate is `<= colorprofile.ANSI`); `false` on `ANSI256` / `TrueColor`.
Plan 3's bracket-fallback chip rendering (D-422) reads the bool to
switch to `Underline+Bold` SGR codes that survive 16-color downsampling.

**9 new tests in `styles_test.go`:**

- `TestStyleColorHexValues_Catppuccin` — locks the three flipped constants
- `TestStyleColorHexValues_UnchangedConstants` — regression guard for the 5 unchanged constants
- `TestColorXANSI_IndicesPerD420` — locks all 8 ANSI indices via reflect comparison
- `TestPalette_StructFields` — enforces 8 color.Color fields + 1 bool field via reflect
- `TestPaletteFor_TrueColorReturns24BitVariants` — TrueColor profile gets 24-bit + Fallback=false
- `TestPaletteFor_ANSI256Returns24BitVariants` — ANSI256 gets 24-bit (gate is `<= ANSI`)
- `TestPaletteFor_ANSIReturnsANSIVariants` — ANSI gets the 16-color variants + Fallback=true
- `TestPaletteFor_AsciiReturnsANSIVariants` — Ascii (TTY-no-color) takes the same fallback branch
- `TestPaletteFor_NoTTYReturnsANSIVariants` — NoTTY (< ANSI) captured by the `<=` gate

### Task 2 — Renderer signature cascade (commit `b6b6688`)

| Renderer            | Old Signature                                                  | New Signature                                                                   |
| ------------------- | -------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| `RenderChrome`      | `(hints, logoStatus, info, width)`                             | `(hints, logoStatus, info, palette, width)`                                     |
| `RenderCrumbs`      | `(segments, width)`                                            | `(segments, palette, width)`                                                    |
| `RenderMenu`        | `(hints, width)`                                               | `(hints, palette, width)`                                                       |
| `RenderInfoPanel`   | `(d)`                                                          | `(d, palette)`                                                                  |
| `RenderLogo`        | `(status, width)` UNCHANGED                                    | UNCHANGED — `LogoStyle{Info,Warn,Error}` package vars auto-flip with the hex   |

`RenderChrome` body forwards `palette` into the 2 internal `RenderMenu`
callsites (mid + full tier) and the 1 `RenderInfoPanel` callsite (full
tier). The `_ = palette` discard lines in `RenderCrumbs`,
`RenderMenu`, and `RenderInfoPanel` bodies are forward-compat seams —
Plan 3 removes them when wiring the fallback rendering body.

**29 test callsites updated** (5 + 8 + 10 + 6) across `chrome_test.go`,
`crumbs_test.go`, `menu_test.go`, `infopanel_test.go` to pass
`ui.PaletteFor(colorprofile.TrueColor)`.

### Task 3 — AppModel + NewAppModel + 4 renderer callsites + 32 test callsites (commit `69931f9`)

**AppModel struct gains:**

```go
profile colorprofile.Profile
palette ui.Palette
```

Both read-only after construction. `palette` is computed once via
`ui.PaletteFor(profile)` in `NewAppModel`; never re-detected.

**`NewAppModel` signature change:**

```go
// Before
func NewAppModel(env ui.EnvStatus, sopsYamlPath string) AppModel

// After (Phase 10 D-419)
func NewAppModel(env ui.EnvStatus, sopsYamlPath string, profile colorprofile.Profile) AppModel
```

**4 renderer callsites in `internal/app/model.go` forward `m.palette`:**

- `View()` line 1402 — `ui.RenderChrome(hints, m.resolveLogoState(), m.infoPanel, m.palette, m.width)`
- `View()` line 1403 — `ui.RenderCrumbs(m.status.Segments(), m.palette, m.width)`
- `chromeHeight()` — `ui.RenderChrome(m.menuHints(), m.resolveLogoState(), m.infoPanel, m.palette, m.width)`
- `crumbsHeight()` — `lipgloss.Height(ui.RenderCrumbs(m.status.Segments(), m.palette, m.width))`

**Test callsite cascade — 38 callsites total** (close-counted by
`grep -rcE 'NewAppModel.*colorprofile\.TrueColor'`):

| File                                  | Callsites |
| ------------------------------------- | --------- |
| `internal/app/model_test.go`          | 19        |
| `internal/app/model_reveal_test.go`   | 5         |
| `internal/app/resize_test.go`         | 4         |
| `internal/app/layout_test.go`         | 3         |
| `internal/app/chrome_test.go`         | 2         |
| `internal/app/model_clipboard_test.go`| 2         |
| `internal/app/severity_test.go`       | 1 (helper)|
| `internal/app/hints_test.go`          | 1 (helper)|
| `internal/app/bench_test.go`          | 1         |

The `hints_test.go` `buildAppModel(t)` helper update cascades into 16
dispatcher tests + 13 `menuhints_drift_test.go` `TestMenuGolden`
sub-tests via the helper indirection. Similarly,
`severity_test.go`'s `newCleanAppModel(t)` helper update cascades
into the 12 severity classifier tests Plan 1 added.

`colorprofile.TrueColor` is the lipgloss/v2 default test profile, so
render-time SGR bytes in goldens stay byte-identical from a profile-
selection perspective; only the D-415 hex flip drives the diff.

### Task 4 — main.go: colorprofile.Detect + override + tea.WithColorProfile + go.mod direct require (commit `3db3e77`)

**`cmd/sops-tui/main.go` — Step 5.5 inserted:**

```go
profile := colorprofile.Detect(os.Stdout, os.Environ())
if os.Getenv("SOPSTUI_FORCE_ASCII") != "" {
    profile = colorprofile.Ascii
}
```

**Step 6 updated:**

```go
model := app.NewAppModel(env, sopsYamlPath, profile)
p := tea.NewProgram(model, tea.WithColorProfile(profile))
```

`colorprofile.Detect` consumes `NO_COLOR`, `CLICOLOR`, `CLICOLOR_FORCE`,
`COLORTERM`, `TERM` env vars and returns one of `NoTTY` / `Ascii` /
`ANSI` / `ANSI256` / `TrueColor`. Pure-function detection — never
re-checked, never per-frame.

`SOPSTUI_FORCE_ASCII` recognises ANY non-empty string (not just `1`)
for forgiving UX. Cheap (4 lines) and answers "16-color goldens look
great but my fancy terminal still mis-detects" support questions.

`tea.WithColorProfile(profile)` ensures Bubble Tea's Cursed Renderer
downsamples SGR bytes to the same profile AppModel uses for variant
selection — eliminates any chance of disagreement between render-time
variant choice and write-time downsample.

**`go.mod` change:** `// indirect` removed from the `colorprofile`
require line. `github.com/charmbracelet/colorprofile v0.4.3` is now a
direct require; version unchanged (already pulled in transitively via
lipgloss/v2). go.sum has zero churn — no new transitive dependencies.

### Task 5 — Hex-literal-using test migration via hexToRGBTriplet (commit `9d5f92e`)

**New helper at `internal/ui/hex_helpers_test.go` (`package ui_test`):**

```go
func hexToRGBTriplet(hex string) string  // "#aabbcc" -> "r;g;b"
func hexBgSGR(hex string) string         // "#aabbcc" -> "48;2;r;g;b"
func hexFgSGR(hex string) string         // "#aabbcc" -> "38;2;r;g;b"
```

**Migrations:**

| Test                                       | File                  | Migration                                                |
| ------------------------------------------ | --------------------- | -------------------------------------------------------- |
| `TestRenderLogo_AllStatusVariants`         | `logo_test.go`        | 3 triplets → `hexToRGBTriplet(ui.Color{Accent,Warning,Error}Hex)` |
| `TestRenderCrumbs_ActiveBoldBg`            | `crumbs_test.go`      | accent + bg triplets → `hexToRGBTriplet`                 |
| `TestRenderCrumbs_SingleSegmentIsActive`   | `crumbs_test.go`      | accent triplet → `hexToRGBTriplet(ui.ColorAccentHex)`    |
| `TestRenderCrumbs_InactiveChipColors`      | `crumbs_test.go`      | surface + fg triplets (defense-in-depth)                 |
| `TestRenderMenu_AccentAppliedToMnemonic`   | `menu_test.go`        | accent triplet → `hexToRGBTriplet(ui.ColorAccentHex)`    |
| `TestStatusBar_FlashWarnPaintsBgTint`      | `statusbar_test.go`   | warning bg + bg fg → `hexBgSGR` / `hexFgSGR`             |
| `TestStatusBar_FlashErrPaintsBgTint`       | `statusbar_test.go`   | error bg + bg fg → `hexBgSGR` / `hexFgSGR`               |
| `TestStatusBar_FlashInfoUsesSurfaceBg`     | `statusbar_test.go`   | warn/err NOT-contains + surface bg contains → SGR helpers |

The 3 statusbar tests close Plan 1's forward deviation #1: Plan 1's
13 SGR-byte tests hardcoded the pre-flip Warning/Error hex SGRs
(`48;2;249;226;175` / `48;2;243;139;168`). Plan 2 migrates the 3
that asserted those specific hexes; the other 10 statusbar tests
either assert structural content (no color SGRs) or use
`assert.NotContains` on text that doesn't depend on color.

**Acceptance grep gate:**

```bash
! grep -rE '"137;180;250"|"249;226;175"|"243;139;168"' internal/ cmd/
```

Returns ZERO matches.

### Task 6 — Atomic GOLDEN_UPDATE pass + UI-15 ASCII fix (commit `00d02fe`)

**Outcome:** `GOLDEN_UPDATE=1 go test ./... -count=1` produced
**ZERO golden file modifications**.

This was unexpected but correct: `internal/testutil/golden.go`'s
`RequireGoldenStructure` uses `ansi.Strip(output)` (structural-only,
not color bytes); `RequireGoldenColors` callsites in
`resize_test.go` all pass `nil` for `wantColors` (Phase 6
scaffolding; Phase 10 doesn't add color-byte assertions to
goldens). The D-415 palette flip changes only color SGR bytes — the
printable chrome characters are byte-identical, so goldens don't
need to be touched.

This confirms the design intent: the palette tune is purely
cosmetic. The Phase 10 Plan 3 work (bracket-fallback rendering)
*will* change golden bytes because it switches from filled chip
pills to bracket characters.

**UI-15 ASCII fix (Rule 1 auto-fix):** Task 2's RenderChrome
doc-comment introduced an em-dash (U+2014) that
`TestChromeASCIIOnly` flagged. Fix: replace with hyphen-minus.
The non-ASCII em-dashes in `internal/ui/styles.go` doc comments
remain — `styles.go` is NOT in the chrome ASCII allowlist scope
(only `chrome.go`, `logo.go`, `menu.go`, `crumbs.go`,
`infopanel.go` are scanned).

## 17 .golden Files Untouched

```bash
$ git diff --stat -- '*.golden'  # produces no output
$ find . -name "*.golden" | wc -l
17
```

All 17 fixtures (13 `menu_*.golden` + 4 `resize_*.golden`) are
byte-identical to their pre-Plan-10-02 state.

## Deviations from Plan

### Implementation Deviations (Auto-Fixed)

**1. [Rule 3 — Plan blocker] lipgloss.Color is a function, not a type in lipgloss/v2**
- **Found during:** Task 1 (writing the test file)
- **Issue:** PLAN.md interfaces section quoted `type Color string` for lipgloss/v2 but actually `lipgloss.Color(s string) color.Color` is a function returning the standard library `image/color.Color` interface.
- **Fix:** Use `color.Color` (the `image/color` interface) as the field type for the 8 Palette color fields. Both `lipgloss.Color(...)` and `lipgloss.ANSIColor(...)` are assignable to `color.Color`.
- **Files modified:** `internal/ui/styles.go` (Palette field types), `internal/ui/styles_test.go` (TestPalette_StructFields uses `reflect.Interface` kind comparison)
- **Commit:** `74dc324`

**2. [Rule 1 — Bug] Em-dash in chrome.go doc comment violates UI-15 ASCII allowlist**
- **Found during:** Task 6 (running full test suite)
- **Issue:** Task 2's RenderChrome doc-comment update added an em-dash (U+2014) which is NOT in the chrome ASCII allowlist. `TestChromeASCIIOnly` failed.
- **Fix:** Replace em-dash with hyphen-minus.
- **Files modified:** `internal/ui/chrome.go`
- **Commit:** `00d02fe`

### D-416 Atomic GOLDEN_UPDATE Pass — Unexpected Zero-Diff Outcome

**Plan expected:** "17 .golden files in `internal/app/testdata/` will refresh; the diff shows only 24-bit RGB SGR byte changes (`137;180;250` → `203;166;247` Mauve, etc.)."

**Actual:** Zero golden file modifications. The plan's mental model was that goldens contained color SGR bytes; they don't. `internal/testutil/golden.go` strips ANSI before comparison (Phase 6 D-08 design). The only color-bytes channel is `RequireGoldenColors`, and Phase 10 doesn't add any.

**No code changed for this** — the D-416 atomic pass *was* run; it just produced no diff. The behaviour is correct and desirable: the palette tune is purely cosmetic and verified by the surviving SGR-byte unit tests in `crumbs_test.go`, `logo_test.go`, `menu_test.go`, `statusbar_test.go` (all migrated to constant-derived triplets in Task 5).

**Effect on Plan 10-03:** Plan 3's bracket-fallback rendering *will*
change golden bytes — bracket chip rendering replaces filled chip
pills, which is a structural delta. Plan 3 should expect goldens
to refresh.

## Test Counts

| Suite                                    | Before | After | Delta  |
| ---------------------------------------- | ------ | ----- | ------ |
| `internal/ui` styles tests (Catppuccin + ANSI + Palette + PaletteFor) | 4 | 13 | +9 |
| `internal/ui` renderer tests (chrome/crumbs/menu/infopanel) | unchanged | unchanged | 0 |
| `internal/ui` statusbar SGR-byte tests | 3 | 3 | 0 (migrated) |
| `internal/ui` logo / crumbs / menu hex-literal tests | 5 | 5 | 0 (migrated) |
| `internal/app` AppModel callsite tests | unchanged | unchanged | 0 |
| `internal/app` grep gates (TestChrome*, TestView*) | green | green | 0 |
| Full suite                               | green  | green | 0      |

`go build ./... && go vet ./... && go test ./... -count=1` exits 0.

## Acceptance Verification

```bash
# Hex flips (D-415)
$ grep -c '#cba6f7' internal/ui/styles.go    # 1 (Mauve)
$ grep -c '#fab387' internal/ui/styles.go    # 1 (Peach)
$ grep -c '#eba0ac' internal/ui/styles.go    # 1 (Maroon)

# 8 ANSI variants (D-420)
$ grep -c "lipgloss.ANSIColor" internal/ui/styles.go  # 8 (locked)

# Palette + PaletteFor (D-421)
$ grep -c "type Palette struct" internal/ui/styles.go  # 1
$ grep -c "func PaletteFor" internal/ui/styles.go      # 1

# Profile detection (D-419)
$ grep -c "colorprofile.Detect" cmd/sops-tui/main.go    # 1
$ grep -c "SOPSTUI_FORCE_ASCII" cmd/sops-tui/main.go    # 2 (call + comment)
$ grep -c "tea.WithColorProfile" cmd/sops-tui/main.go   # 2 (call + comment)

# go.mod direct require
$ grep -E "github.com/charmbracelet/colorprofile v0.4.3$" go.mod   # match (no // indirect)

# AppModel wiring
$ grep -c "profile colorprofile.Profile" internal/app/model.go  # 2 (field + signature)
$ grep -c "palette ui.Palette" internal/app/model.go            # 1
$ grep -c "palette: ui.PaletteFor(profile)" internal/app/model.go  # 1

# Renderer signatures
$ grep -c "palette Palette" internal/ui/{chrome,crumbs,menu,infopanel}.go  # 1 each = 4 total

# Renderer callsites
$ grep -cE 'RenderChrome\(.+m\.palette' internal/app/model.go  # 2 (View + chromeHeight)
$ grep -cE 'RenderCrumbs\(.+m\.palette' internal/app/model.go  # 2 (View + crumbsHeight)

# Hex literal cleanup (D-417)
$ grep -rE '"137;180;250"|"249;226;175"|"243;139;168"' internal/ cmd/  # zero matches

# Grep gates
$ go test ./internal/app/ -run "TestChromeASCIIOnly|TestChromeNormalBorderOnly|TestViewNoNewStyle" -count=1  # PASS
$ go test ./internal/ui/  -run "TestSubmodelViewsNoNewStyle" -count=1  # PASS
```

All gates pass.

## Forward Deviations for Plan 10-03

1. **`palette.Fallback` is exposed but not yet consumed in `RenderCrumbs` body.** Plan 2's
   `_ = palette` discard line in `internal/ui/crumbs.go` is the cleanup target.
   Plan 3 wires the `if palette.Fallback { ...bracket+underline+bold chips... }`
   branch per D-422.

2. **The `_ = palette` discard lines in `RenderMenu` and `RenderInfoPanel` may
   stay** if Plan 3 chooses to keep the existing pill rendering for menu/info-panel
   on fallback profiles (they already use distinct fg colors which 16-color
   downsamples preserve). Plan 3 author decision.

3. **4-profile teatest matrix (D-423) does not exist yet.** Plan 1 + Plan 2's
   tests are TrueColor-profile-only (lipgloss/v2 test default). Plan 3 will
   add `Ascii` / `ANSI` / `ANSI256` / `TrueColor` matrix tests that verify the
   `FlashWarnBarStyle`, `FlashErrBarStyle`, and crumb chip SGR bytes downsample
   correctly. The `hexBgSGR` / `hexFgSGR` test helpers from Plan 2 are
   TrueColor-only; Plan 3's matrix needs a parallel mechanism that produces
   the expected bytes per profile (e.g., a `bgSGRForProfile(hex, profile)`
   helper that consults a downsample lookup table or runs a forced-profile
   render and compares the SGR substring).

4. **Goldens refresh expected in Plan 3.** Plan 2's atomic pass produced zero
   refreshed goldens (intended; structural-only goldens). Plan 3's
   bracket-fallback chip rendering changes the printable chrome characters
   on `Ascii`/`ANSI` profiles — the goldens for those scenarios *will* drift,
   and the plan should budget the GOLDEN_UPDATE pass at a different layer
   (e.g., per-profile golden keying like `resize_80x24_ascii.golden` if
   Plan 3 adds the 4-profile matrix to existing resize fixtures).

5. **`SOPSTUI_FORCE_ASCII=1` is the only env override today.** Plan 3 may
   want to add `SOPSTUI_FORCE_ANSI256=1` for testing the ANSI256 fallback
   path, or a more general `SOPSTUI_FORCE_PROFILE=<name>` switch. This is
   Claude's-Discretion territory; not in scope unless Plan 3 needs it
   for golden generation harnesses.

## Self-Check: PASSED

Created files:
- FOUND: /home/moersener/git/sops-tui/internal/ui/hex_helpers_test.go

Modified files (verified by `git log --pretty=format`):
- FOUND: internal/ui/styles.go (74dc324)
- FOUND: internal/ui/styles_test.go (74dc324)
- FOUND: internal/ui/chrome.go (b6b6688 + 00d02fe)
- FOUND: internal/ui/crumbs.go (b6b6688)
- FOUND: internal/ui/menu.go (b6b6688)
- FOUND: internal/ui/infopanel.go (b6b6688)
- FOUND: internal/ui/chrome_test.go (b6b6688)
- FOUND: internal/ui/crumbs_test.go (b6b6688 + 9d5f92e)
- FOUND: internal/ui/menu_test.go (b6b6688 + 9d5f92e)
- FOUND: internal/ui/infopanel_test.go (b6b6688)
- FOUND: internal/ui/logo_test.go (9d5f92e)
- FOUND: internal/ui/statusbar_test.go (9d5f92e)
- FOUND: internal/app/model.go (69931f9)
- FOUND: internal/app/bench_test.go, chrome_test.go, hints_test.go, layout_test.go, menuhints_drift_test.go, model_clipboard_test.go, model_reveal_test.go, model_test.go, resize_test.go, severity_test.go (69931f9)
- FOUND: cmd/sops-tui/main.go (3db3e77)
- FOUND: go.mod (3db3e77)

Commits:
- FOUND: 74dc324 (feat: flip palette + add ANSI variants + Palette/PaletteFor)
- FOUND: b6b6688 (feat: cascade Palette parameter through chrome renderers)
- FOUND: 69931f9 (feat: wire profile + palette through AppModel)
- FOUND: 3db3e77 (feat: detect color profile + plumb tea.WithColorProfile)
- FOUND: 9d5f92e (test: migrate hex-literal SGR assertions to constant-derived)
- FOUND: 00d02fe (fix: replace em-dash with hyphen-minus in chrome.go doc comment)

Acceptance grep checks:
- FOUND: 1 occurrence each of #cba6f7, #fab387, #eba0ac in styles.go (D-415)
- FOUND: 8 lipgloss.ANSIColor declarations in styles.go (D-420)
- FOUND: 1 type Palette struct + 1 func PaletteFor in styles.go (D-421)
- FOUND: 1 colorprofile.Detect, 2 SOPSTUI_FORCE_ASCII, 2 tea.WithColorProfile in main.go (D-419)
- FOUND: github.com/charmbracelet/colorprofile v0.4.3 (no // indirect) in go.mod
- FOUND: 0 occurrences of "137;180;250" / "249;226;175" / "243;139;168" hex literals (D-417)
- FOUND: 4 renderers with palette parameter; 4 model.go callsites forwarding m.palette
- FOUND: 38 NewAppModel callsites passing colorprofile.TrueColor across 9 test files

Goldens:
- VERIFIED: 17 .golden files unchanged (zero diff after GOLDEN_UPDATE pass)

Build / vet / test:
- VERIFIED: go build ./... exits 0
- VERIFIED: go vet ./... exits 0
- VERIFIED: go test ./... -count=1 exits 0 (full suite green)
- VERIFIED: go test ./internal/app/ -run "TestChromeASCIIOnly|TestChromeNormalBorderOnly|TestViewNoNewStyle" -count=1 all green
- VERIFIED: go test ./internal/ui/ -run "TestSubmodelViewsNoNewStyle" -count=1 green
