# Phase 10: Theming + Accessibility - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-04
**Phase:** 10-theming-accessibility
**Areas discussed:** Severity classifier rules, Flash typed-API + [I]/[W]/[E] prefix, k9s palette tune target, 16-color fallback strategy

---

## Severity Classifier Rules

### Q1: Which signals kick the logo to LogoError?

| Option | Description | Selected |
|--------|-------------|----------|
| Recent flash classified Err | In-window flash with severity Err (re-encrypt failed, decrypt error, edit error, rotation failed, health scan crash). Stays red until flash clears (2s) or user navigates. | ✓ |
| Health scan: weak/duplicate findings | When the most recent HealthCheckResult has WeakSecrets or Duplicates > 0. Stale files don't count. | ✓ |
| Persistent env failure | If something the startup gate didn't catch fails later (e.g., age key file deleted mid-session). Detection requires re-stat. | ✓ |
| Health scan: errors only | ONLY HealthCheckResult.Errors (scanner crashed) raises to error; weak/duplicate stay Warn. | ✓ |

**User's choice:** All four — interpreted as "broadest read: any non-empty health output (weak, duplicate, OR errors) raises to LogoError; flash Err raises; persistent env failure raises."
**Notes:** D-401 captures this as the union of three categories. Mid-session env failure detection plumbs through existing event paths (no polling, no per-frame stat) per Pitfall 15.

---

### Q2: Which signals sit at LogoWarn?

| Option | Description | Selected |
|--------|-------------|----------|
| Soft env: age/.sops.yaml missing | When EnvStatus.AgeAvailable=false or SopsYamlAvailable=false. | ✓ |
| Recent flash classified Warn | Flash messages like "No changes detected", "No changes" (for confirms), "Reveal first with r". Promotes to Warn for ~2s. | ✓ |
| Health scan: stale files only | When HealthCheckResult.StaleFiles is non-empty but no weak/duplicate findings. | |
| Git: any uncommitted change | When ANY file in the project has a git badge ([M]/[A]/[?]). Risk: most users always have at least one dirty file. | |

**User's choice:** Soft env + flash classified Warn.
**Notes:** Stale files and git-dirty explicitly demoted below Warn per D-402.

---

### Q3: When a flash Err clears (after 2s), what should the logo do?

| Option | Description | Selected |
|--------|-------------|----------|
| Drop to baseline (env + health) | Logo immediately recomputes from underlying signals. Pure function of state, no sticky severity. | ✓ |
| Sticky for one navigation cycle | Logo stays Err until the user moves focus. Adds a `pendingErrorAck` state field. | |
| Sticky until manual ack | Logo stays Err until user dismisses. Feels too modal. | |

**User's choice:** Drop to baseline.
**Notes:** D-403 confirms `resolveLogoState()` is a pure function of state per Pattern 4 in ARCHITECTURE.md.

---

### Q4: Should the logo communicate severity beyond color?

| Option | Description | Selected |
|--------|-------------|----------|
| No status message — color only | Keep the current 6-row art; severity conveyed through color + redundant `[I]/[W]/[E]` flash prefix. | ✓ |
| Repurpose row 6 as status text | Row 6 becomes a short severity label. | |
| Adjacent 1-row status line below logo | Logo stays 6 rows; add a 7th. Inflates chromeHeight; breaks goldens. | |

**User's choice:** Color only.
**Notes:** D-405 keeps Phase 7's locked logo art intact. Severity surface lives in flash bar instead.

---

## Flash Typed-API + [I]/[W]/[E] Prefix

### Q1: How should the typed-flash API land?

| Option | Description | Selected |
|--------|-------------|----------|
| Three new methods + keep Flash as Info alias | Add FlashInfo / FlashWarn / FlashErr; existing `Flash(msg)` becomes a thin wrapper for `FlashInfo(msg)`. Migration: re-classify the 42 call-sites. | ✓ |
| Single Flash(severity, msg) signature | Replace `Flash(msg)` with `Flash(severity, msg)`. Cleanest API; breaks every callsite at once. | |
| Builder pattern | `Flash(msg).Err()` / `.Warn()` / `.Info()` chain. More ceremony for marginal gain. | |

**User's choice:** Three new methods + Flash as Info alias.
**Notes:** D-406 picks minimum-churn migration. Most existing callsites stay untouched; only error/warn callsites upgrade.

---

### Q2: Should the rendered flash text gain a `[I]`/`[W]`/`[E]` prefix?

| Option | Description | Selected |
|--------|-------------|----------|
| Yes — prefix added inside StatusBarModel.View() (Warn/Err only) | When `flashSeverity` is non-zero, prepend `[W] ` or `[E] `. FlashInfo stays unprefixed. | ✓ |
| Prefix all three (incl. [I]) | Always prefix. More uniform but adds clutter on the 22 neutral flashes. | |
| Color only — no prefix | Lean on the bg/fg color of the flash. Pitfall 9 risk. | |

**User's choice:** Prefix Warn/Err only (no [I] clutter).
**Notes:** D-411 captures: prefix at render time, not at Flash() call. Test fixtures stay clean.

---

### Q3: Should the flash bar background change with severity?

| Option | Description | Selected |
|--------|-------------|----------|
| Yes — bg shifts on Warn/Err | FlashWarn paints with ColorWarning bg + ColorBg fg; FlashErr with ColorError bg + ColorBg fg; Info keeps surface bg. Two-channel encoding on Warn/Err. | ✓ |
| No — keep surface bg always | Severity via prefix only. Weaker peripheral signal. | |
| Foreground tint only | Bg stays surface; fg shifts. Pitfall 9 risk under deuteranopia + 16-color combo. | |

**User's choice:** Severity-tinted bg.
**Notes:** D-412 confirms two-channel (bg + prefix) on Warn/Err; surface-only on Info.

---

### Q4: Where should `FlashSeverity` and `resolveLogoState()` live?

| Option | Description | Selected |
|--------|-------------|----------|
| FlashSeverity in `internal/ui`, classifier in `internal/app` | Enum + flash methods local to StatusBarModel; classifier in AppModel where env+flash+health converge. | ✓ |
| Both in `internal/ui` | Classifier moves to ui as a pure function. ui learns about HealthCheckResult. | |
| Both in a new `internal/severity` package | New package for one enum + one function. Over-engineered. | |

**User's choice:** Split — enum/methods in ui, classifier in app.
**Notes:** D-408 follows Pattern 4 in ARCHITECTURE.md. No cross-package wiring beyond `FlashSeverity()` accessor.

---

## k9s Palette Tune Target

### Q1: ColorAccent direction?

| Option | Description | Selected |
|--------|-------------|----------|
| Mauve `#cba6f7` (Catppuccin Mauve) | Catppuccin's purple. Stays in family with existing palette; distinct from warning-yellow + error-red. | ✓ |
| Pink `#f5c2e7` (Catppuccin Pink) | Hot-pink, closest to k9s default skin. Risk: pink + yellow contrast degrades on 16-color. | |
| Keep sky-blue `#89b4fa` | No accent change. Doesn't satisfy UI-12 wording. | |
| Custom hex | User-typed. | |

**User's choice:** Mauve `#cba6f7`.
**Notes:** D-415's first hex flip. Phase 1-5 styles auto-inherit through `Foreground(ColorAccent)`.

---

### Q2: How aggressive is the palette retune?

| Option | Description | Selected |
|--------|-------------|----------|
| Accent only — keep everything else | Only ColorAccent changes. Lowest risk. | |
| Accent + minor warning/error tune | Accent → mauve. Warning → Catppuccin Peach `#fab387`. Error → Catppuccin Maroon `#eba0ac`. Adds separation from new mauve. | ✓ |
| Full palette retune | Touch every named color. Highest risk; breaks color-presence goldens widely. | |

**User's choice:** Accent + warning/error tune.
**Notes:** D-415's three hex flips. Provides separation between mauve / peach / maroon under both 24-bit and 4-bit downsample.

---

### Q3: How should the goldens regen wave land?

| Option | Description | Selected |
|--------|-------------|----------|
| Single GOLDEN_UPDATE pass after palette change | Land hex flip; run `GOLDEN_UPDATE=1 go test ./...`; commit refreshed goldens in a single follow-up commit. Atomic. | ✓ |
| Per-file regen as plan tasks | Each plan task regenerates that sub-model's goldens. Multiplies diff surface. | |
| Defer color goldens to Phase 11 | Goldens diverge from rendered reality. Reviewer confusion. | |

**User's choice:** Single GOLDEN_UPDATE pass.
**Notes:** D-416. Reviewer sees clean "color SGR bytes only, no structural drift" diff.

---

## 16-Color Fallback Strategy

### Q1: When and how should the color profile be detected?

| Option | Description | Selected |
|--------|-------------|----------|
| Once at startup in main.go | `colorprofile.Detect(os.Stdout, os.Environ())` in `cmd/sops-tui/main.go`; pass into NewAppModel. Read-only field. | ✓ |
| Lazy on first View() | Detect inside View() once, cache. Pulls syscall path into per-frame method. | |
| Per-frame | Re-detect every render. Wasteful, against Pitfall 15. | |

**User's choice:** Startup once.
**Notes:** D-419. Pure-function path; never re-detected.

---

### Q2: How should the 16-color fallback palette ship?

| Option | Description | Selected |
|--------|-------------|----------|
| Parallel `var(...)` block + Palette() accessor | Add `Color*ANSI` vars (named-ANSI lipgloss values). Small `Palette(profile)` accessor returns the right set. | ✓ |
| Build-time variant styles, runtime swap | Two complete style sets per element; renderer functions accept profile and pick. Tighter but bigger API surface. | |
| Trust lipgloss auto-downsample | Don't ship a fallback. Pitfall 5 exact failure: chip bg/fg both → white. Reject. | |

**User's choice:** Parallel `var()` block + `Palette` accessor.
**Notes:** D-420 + D-421. ANSI variants verified-distinct under 4-bit per Pitfall 5 §4.

---

### Q3: What should chips look like on Ascii/ANSI profiles?

| Option | Description | Selected |
|--------|-------------|----------|
| Bracket + underline active | On Ascii/ANSI: `<segment>` no bg fill; active = underline + bold. Two-channel (text shape + bold) survives downsample. | ✓ |
| Drop bg, keep bold + invert fg | Bg-less but still fg-color-coded. Risk: deuteranopia + 16-color combo loses active marker. | |
| Keep current rendering (no fallback) | Rely on bold alone. Risk: invisible chip if both bg and fg downsample to white. | |

**User's choice:** Bracket + underline active.
**Notes:** D-422 implements Pitfall 5 §2 verbatim. `<` and `>` already in `TestChromeASCIIOnly` allowlist.

---

### Q4: How many profile variants in the golden matrix?

| Option | Description | Selected |
|--------|-------------|----------|
| Two: TrueColor + ANSI (4-bit) | Cover the two ends of the spectrum. Manageable golden count. | |
| Four: Ascii + ANSI + ANSI256 + TrueColor | Full Pitfall 5 prescription. 4x golden count; heavy maintenance. | ✓ |
| Three: Ascii + ANSI + TrueColor | Skip ANSI256 (auto-downsamples cleanly). | |
| One: TrueColor + a single 16-color smoke test | Lightest weight; relies on implementation correctness. | |

**User's choice:** All four profiles.
**Notes:** D-423 ships the full Pitfall 5 §3 prescription. ~16 new color-bearing goldens at the matrix entry points.

---

## Claude's Discretion

- Exact `Palette` struct shape (grouped by surface area vs flat).
- Whether the renderer parameter is `colorprofile.Profile` or a pre-resolved `Palette` struct.
- ANSI color index selection (the §4 Pitfall-5 hand-verification table).
- Whether `resolveLogoState()` returns the full struct or just the enum value.
- Whether `m.lastHealthResult` is a new field on AppModel or accessed via the existing HealthModel sub-model.
- File location of the `Palette` accessor (top of `styles.go` vs new `palette.go`).
- Whether to expose a `SOPSTUI_FORCE_ASCII=1` env var override.
- Logo art width-responsiveness (already accepts `width` parameter).
- Whether the 4-profile teatest matrix forces profile via `lipgloss.Writer.Profile = ...` or a teatest constructor option.
- Whether the narrow-tier (`<41` cols) at the 4-profile matrix gets its own goldens or inherits from full-tier.

## Deferred Ideas

- Stale files contributing to logo severity (rejected — too noisy).
- Git-dirty contributing to logo severity (rejected — most users always have a dirty file).
- Sticky logo state after flash auto-clear (rejected — pure function of state).
- Logo status text row inside the 6-row art (rejected — Phase 7 art locked).
- AppModel-cached severity field (rejected — pure per-frame, all inputs already on model).
- Per-frame profile detection (rejected — startup once is the standard).
- Polling goroutine for env re-check (forbidden by ROADMAP).
- Skin loader, builtin skins, live reload (all v2: THM-01..03).
- `Skin` struct scaffolding for v2 forward-compat (deferred to v2 entirely per ROADMAP).
- BenchmarkAppView budget tightening (Phase 11 SC2).
- v1.0 functional regression sweep (Phase 11 SC1).
- Terminal compat sweep + alt-screen cleanup (Phase 11).
- Health-finding severity at-rest classifier in `internal/health/checker.go` (possibly Phase 11).
