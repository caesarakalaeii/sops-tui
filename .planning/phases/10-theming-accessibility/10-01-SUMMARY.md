---
phase: 10
plan: 01
subsystem: theming-accessibility
tags: [theming, accessibility, severity, flash, redundant-encoding, ui-03, ui-14]
requires:
  - "Phase 7: Logo art locked at 6 rows (D-02 deferred severity wiring); LogoStatus enum + LogoStyle{Info,Warn,Error} package vars exist"
  - "Phase 8: StatusBarModel post-D-211 minimal shape (env indicators + flash branch); StatusBarStyle as the canonical surface bg"
  - "Phase 9: model.go composition shape; menuhints / hintscallsite topology (no behavior dependency, pattern only)"
provides:
  - "internal/ui/statusbar.go: FlashSeverity enum (FlashSevInfo=0/FlashSevWarn=1/FlashSevErr=2)"
  - "internal/ui/statusbar.go: FlashSeverity() accessor"
  - "internal/ui/statusbar.go: FlashInfo / FlashWarn / FlashErr typed methods (Flash preserved as thin alias)"
  - "internal/ui/statusbar.go: severity-aware View() flash branch with [W]/[E] prefix + bg tint"
  - "internal/ui/styles.go: FlashWarnBarStyle + FlashErrBarStyle package vars"
  - "internal/health/checker.go: HasErrLevelFindings() bool predicate (excludes StaleFiles per D-401)"
  - "internal/ui/health.go: LastResult() health.HealthCheckResult accessor"
  - "internal/app/model.go: resolveLogoState() ui.LogoStatus pure-function-of-state classifier"
  - "internal/app/model.go: both ui.RenderChrome callsites consume m.resolveLogoState()"
  - "internal/app/export_test.go: ResolveLogoStateForTest + StatusForTest + WithStatusForTest + HealthForTest + WithHealthForTest test-only shims"
affects:
  - "Plan 10-02 will read FlashWarnBarStyle / FlashErrBarStyle and pick up the new ColorWarning/ColorError hex values automatically (named-var indirection)"
  - "Plan 10-03 will add bracket-fallback chip rendering (D-422); Plan 1 leaves crumb chip rendering untouched per scope boundary"
  - "Phase 11 verification will UAT the severity coupling (logo recolor on flash Err / health Err / soft env)"
tech-stack:
  added: []
  patterns:
    - "Pure function of state for severity (PATTERNS.md Pattern 7) — re-evaluated every View() frame; no caching, no sticky state"
    - "Generation-counter Pitfall 6 protection extended to severity field (cleared on matching FlashClearMsg ack)"
    - "Render-time prefix string concatenation for redundant text encoding (m.flash storage stays clean)"
    - "Test-only export shims via _test.go file suffix (production binary unaffected)"
    - "TDD RED → GREEN cycle on each unit (HasErrLevelFindings, LastResult, resolveLogoState all tested first)"
key-files:
  created:
    - "internal/app/severity_test.go (12 classifier truth-table tests)"
    - "internal/app/export_test.go (test-only shims for unexported AppModel methods)"
  modified:
    - "internal/ui/statusbar.go (Task 1 — FlashSeverity enum + typed methods + severity-aware View)"
    - "internal/ui/styles.go (Task 1 — FlashWarnBarStyle + FlashErrBarStyle package vars)"
    - "internal/ui/statusbar_test.go (Task 1 — 13 new severity tests, all green)"
    - "internal/health/checker.go (Task 2 — HasErrLevelFindings predicate)"
    - "internal/health/checker_test.go (Task 2 — 6 new HasErrLevelFindings tests)"
    - "internal/ui/health.go (Task 2 — LastResult accessor on HealthModel)"
    - "internal/ui/health_test.go (Task 2 — 2 new LastResult tests)"
    - "internal/app/model.go (Task 2 — resolveLogoState + both RenderChrome callsites; Task 3 — 27 callsite re-classifications)"
decisions:
  - "Followed PATTERNS.md Pattern 7 verbatim for resolveLogoState (D-401/D-402/D-403/D-404)"
  - "Used HasErrLevelFindings predicate instead of inlining the slice checks in resolveLogoState (recommendation b from PATTERNS.md ‘Deviations the planner should know’)"
  - "test-only shims placed in NEW file internal/app/export_test.go (no prior shim file existed; followed the recommended pattern)"
  - "Did NOT add the optional EnvRecheckMsg path for persistent env failure detection (PLAN.md Step 2.3 NOTE marks this as Phase 11 if needed)"
metrics:
  duration_minutes: 35
  date_completed: "2026-05-04"
  tasks_total: 3
  tasks_completed: 3
  files_changed: 9
  files_created: 2
---

# Phase 10 Plan 01: Severity Classifier + Flash Typed-API + Redundant Prefix Summary

Wired the Phase 10 severity-classification primitive layer end-to-end without
touching the palette or any goldens. The `FlashSeverity` enum, typed
`FlashInfo`/`FlashWarn`/`FlashErr` methods, severity-tinted bg + `[W]`/`[E]`
prefix, `resolveLogoState()` classifier, and 42-callsite flash re-classification
all landed in three atomic commits with zero `.golden` churn.

## Commits

| Hash      | Task | Type | Subject                                                              |
| --------- | ---- | ---- | -------------------------------------------------------------------- |
| `a3fe389` | 1    | feat | FlashSeverity enum + typed flash methods + bg-tinted View            |
| `3f9cf98` | 2    | feat | severity classifier resolveLogoState + supporting accessors          |
| `8381a11` | 3    | feat | classify 27 flash callsites by severity (15 Err + 12 Warn)           |

## What Landed

### Task 1 — FlashSeverity primitive (commit `a3fe389`)

- `internal/ui/statusbar.go`: `FlashSeverity` enum (`FlashSevInfo=0` /
  `FlashSevWarn=1` / `FlashSevErr=2`); `flashSeverity` field on `StatusBarModel`;
  `setFlash` private helper; `FlashInfo` / `FlashWarn` / `FlashErr` typed methods;
  `Flash` preserved as thin alias for `FlashInfo` (zero-callsite-break migration);
  `FlashSeverity()` accessor (returns `FlashSevInfo` when `m.flash == ""`);
  `Update(FlashClearMsg)` extended to clear severity to baseline on matching ack;
  `View()` flash branch with severity-aware `style` + render-time `[W]`/`[E]`
  prefix.
- `internal/ui/styles.go`: `FlashWarnBarStyle` (`Background(ColorWarning).Foreground(ColorBg).Padding(0, 1)`);
  `FlashErrBarStyle` (`Background(ColorError).Foreground(ColorBg).Padding(0, 1)`).
- `internal/ui/statusbar_test.go`: 13 new tests covering zero-value, every typed
  method, render-time prefix, bg-tint SGR presence, FlashClearMsg ack, stale-tick
  protection (Pitfall 6).

### Task 2 — Severity classifier + supporting accessors (commit `3f9cf98`)

- `internal/health/checker.go`: `HasErrLevelFindings() bool` — `len(WeakSecrets) +
  len(Duplicates) + len(Errors) > 0`. **StaleFiles deliberately excluded** per
  D-401 so stale-only health stays below Warn (logo stays at env baseline).
  Distinct from `IsEmpty()` which DOES include stale files.
- `internal/ui/health.go`: `LastResult() health.HealthCheckResult` value-receiver
  accessor (read path safe as value receiver since `HealthCheckResult` is a struct
  of slice headers; underlying arrays not copied).
- `internal/app/model.go`: `resolveLogoState() ui.LogoStatus` — single-pass
  switch walking Err checks first then Warn checks, falls through to LogoInfo
  baseline. Inputs: `m.status.FlashSeverity()`, `m.health.LastResult().HasErrLevelFindings()`,
  `m.status.Env()`. Pure function of state — re-evaluated every `View()` frame.
- `internal/app/model.go`: both `ui.RenderChrome` callsites swapped from
  `ui.LogoInfo` (unconditional) to `m.resolveLogoState()` (severity-derived).
- `internal/app/export_test.go`: 5 test-only shims (`ResolveLogoStateForTest`,
  `StatusForTest`, `WithStatusForTest`, `HealthForTest`, `WithHealthForTest`).
- `internal/app/severity_test.go`: 12 truth-table tests (default → Info; flash
  Err → Error; flash Warn → Warn; flash Info on clean → Info; health Weak/Dup/Err
  → Error; stale-only → Info (D-401 demotion proof); soft env age missing → Warn;
  soft env .sops.yaml missing → Warn; flash Err over soft env Warn → Error
  (D-404 precedence); health Err over flash Warn → Error (D-404 precedence)).
- `internal/health/checker_test.go`: 6 new tests for `HasErrLevelFindings`.
- `internal/ui/health_test.go`: 2 new tests for `LastResult`.

### Task 3 — 42-callsite flash sweep (commit `8381a11`)

Re-classified each `m.status.Flash(...)` callsite per the D-401/D-402 severity
rules. Migration tally matches the canonical map exactly: **15 Err + 12 Warn +
15 Info (Flash alias) = 42**.

## 42-Callsite Migration Record

Live grep at execution start returned exactly 42 callsites — same as the
canonical map (no drift). Line numbers shifted slightly during execution because
Task 2's `resolveLogoState` body added ~30 lines around line 1593.

### FlashErr — 15 callsites (error paths, raise LogoError per D-401)

| Line (post-Task-2) | Message                                                | Notes                |
| ------------------ | ------------------------------------------------------ | -------------------- |
| 349                | `"Error discovering files: " + msg.Err.Error()`        | FilesDiscoveredMsg   |
| 389                | `"Error parsing file: " + msg.Err.Error()`             | FilesParsedMsg       |
| 440                | `"Decrypt error: " + msg.Err.Error()`                  | First occurrence     |
| 449                | `"Decrypt error: " + msg.Err.Error()`                  | Second occurrence    |
| 480                | `"Editor error: " + msg.Err.Error()`                   | $EDITOR flow         |
| 543                | `"Re-encryption failed: " + msg.Err.Error()`           | Edit confirm         |
| 591                | `"Rotation failed: " + msg.Err.Error()`                | EDT-03               |
| 672                | `"Git history error: " + msg.Err.Error()`              | History overlay      |
| 681                | `"Health scan failed: " + msg.Err.Error()`             | HLT-03               |
| 697                | `"Re-key failed: " + msg.Err.Error()`                  | First occurrence     |
| 709                | `"Re-key failed: " + msg.Err.Error()`                  | Second occurrence    |
| 943                | `"Generation failed: " + err.Error()`                  | Recipient form       |
| 1013               | `"Error reading metadata: " + err.Error()`             | Metadata overlay     |
| 1445               | `"Clipboard error: " + err.Error()`                    | (statusCmd path)     |
| 2010               | `"Re-key: could not read recipients: " + err.Error()`  | Bulk re-key prelude  |

### FlashWarn — 12 callsites (soft validation, raise LogoWarn per D-402)

| Line (post-Task-2) | Message                                                  | Notes               |
| ------------------ | -------------------------------------------------------- | ------------------- |
| 487                | `"Read error: " + err.Error()`                           | Recoverable read    |
| 493                | `"Diff error: " + err.Error()`                           | Recoverable diff    |
| 497                | `"No changes detected"`                                  | Edit no-op          |
| 513                | `"No changes"`                                           | Edit no-op variant  |
| 530                | `msg.Reason`                                             | Edit blocked reason |
| 532                | `"Reveal first with r"`                                  | First occurrence    |
| 1085               | `"Reveal first with r"`                                  | Second occurrence   |
| 1102               | `"No git repository found"`                              | History blocked     |
| 1138               | `"No age recipients configured for this file"`           | Recipient flow      |
| 1222               | `"Select files with space first"`                        | Bulk re-key prelude |
| 1251               | `"No files to scan"`                                     | Health prelude      |
| 1440               | `"Clipboard not available (install xclip or wl-clipboard)"` | (statusCmd path)  |

### Flash (Info alias) — 15 callsites (neutral/success/progress, stay LogoInfo)

| Line (post-Task-2) | Message                                                  | Notes              |
| ------------------ | -------------------------------------------------------- | ------------------ |
| 444                | `"Decrypted"`                                            | Reveal success     |
| 453                | `"All values decrypted"`                                 | Reveal-all success |
| 546                | `"Rotated to " + ui.FormatLabel(m.rotateFormat)`         | Rotate success     |
| 554                | `"Re-encrypted"`                                         | Edit confirm OK    |
| 688                | `"Health scan done — no issues found"`                   | Health success     |
| 690                | `fmt.Sprintf("Health scan done — %d findings", total)`   | Health summary     |
| 711                | `"Recipient " + msg.Action + ". File re-encrypted."`     | Recipient success  |
| 809                | `"Adding recipient..."`                                  | Progress           |
| 817                | `"Removing recipient..."`                                | Progress           |
| 876                | `"Decrypting all files for health scan..."`              | Progress           |
| 919                | `"Cancelled"`                                            | Neutral            |
| 1453               | `"Copied (clears in 30s)"`                               | Clipboard success  |
| 2006               | `fmt.Sprintf("Re-keying %d/%d: %s", ...)`                | Per-file progress  |
| 2035               | `fmt.Sprintf("Re-key complete: %d files updated", ...)`  | Bulk success       |
| 2037               | `fmt.Sprintf("Re-key done: %d updated, %d skipped", ...)` | Bulk partial      |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 — Plan blocker] Test plan struct field name mismatch**
- **Found during:** Task 2 (writing severity_test.go)
- **Issue:** PLAN.md `<behavior>` block referenced fictional types (`DuplicateGroup`,
  `ScanError`) and field names (`File`, `Key`) that do not exist in
  `internal/health/checker.go`. Actual types are `Duplicate` (with `ValueHash`,
  `Locations`), `Errors []string` (not `[]ScanError`), `WeakSecret{FilePath, KeyPath, Reason}`,
  `StaleFile{FilePath, LastCommitTime, DaysSince}`.
- **Fix:** Used the actual types in severity_test.go. The plan's intent
  (verify Weak/Duplicate/Errors raise to Err; Stale alone does NOT) is
  preserved exactly.
- **Files modified:** internal/app/severity_test.go
- **Commit:** 3f9cf98

**2. [Rule 3 — Plan blocker] Test plan module path mismatch**
- **Found during:** Task 2 (writing severity_test.go and export_test.go)
- **Issue:** PLAN.md test-plan code referenced `github.com/sipgate/sops-tui/internal/...`
  but `go.mod` declares `module github.com/caesarakalaeii/sops-tui`.
- **Fix:** Used the actual module path throughout the new test files.
- **Files modified:** internal/app/severity_test.go, internal/app/export_test.go
- **Commit:** 3f9cf98

### Tightening on Plan grep gate

**3. [Documentation] Acceptance grep `grep -c "ui\.LogoInfo"` reports 2 not 0**
- **Found during:** Task 2 verification
- **Issue:** PLAN.md acceptance criterion #7 says
  `grep -c "ui\.LogoInfo" /home/moersener/git/sops-tui/internal/app/model.go`
  must report 0. After Task 2 the count is 2: one is the doc-comment
  `// ui.LogoInfo unconditionally.` inside the resolveLogoState body; the other
  is the `return ui.LogoInfo` baseline-default of resolveLogoState (the canonical
  pattern from PATTERNS.md line 781).
- **Resolution:** The semantic intent (no `RenderChrome` callsite passes
  unconditional `ui.LogoInfo`) is satisfied:
  `grep -c "RenderChrome.*ui\.LogoInfo" internal/app/model.go` reports 0. The
  classifier MUST return `ui.LogoInfo` as the baseline value — that's how the
  enum works. Updated grep gate would be `grep -c "RenderChrome.*ui\.LogoInfo"
  internal/app/model.go` reports 0; live `grep -c "m\.resolveLogoState()"`
  reports 2 (View() composition + chromeHeight()).
- **No code changed for this** — documentation drift only.

## Goldens Confirmation

- `git status --short internal/app/testdata/ internal/ui/testdata/` after all 3
  commits: **clean** (no `.golden` modifications).
- `git diff --stat HEAD~3..HEAD -- '*.golden'`: **no output** (zero golden bytes
  changed).
- The classifier returns `LogoInfo` for the AppModel state used by all existing
  goldens (clean env, no flash, no health findings) — same severity the goldens
  were captured against, so chrome ANSI bytes are byte-identical.

## Test Counts

| Suite                                      | Before | After | Delta  |
| ------------------------------------------ | ------ | ----- | ------ |
| `internal/health` HasErrLevelFindings      | 0      | 6     | +6     |
| `internal/ui` HealthModel_LastResult       | 0      | 2     | +2     |
| `internal/ui` StatusBar_Flash* (Task 1)    | 0      | 13    | +13    |
| `internal/app` ResolveLogoState            | 0      | 12    | +12    |
| All grep-gates (TestChrome*, TestView*)    | green  | green | 0      |
| Full suite                                 | green  | green | 0      |

`go build ./... && go vet ./... && go test ./... -count=1` exits 0.

## Forward Deviations for Plan 10-02

1. **`FlashWarnBarStyle` and `FlashErrBarStyle` are wired through named-var
   indirection** (`Background(ColorWarning)`, `Background(ColorError)`). Plan
   2's three hex flips
   (`ColorWarningHex` `#f9e2af`→`#fab387` Peach; `ColorErrorHex` `#f38ba8`→`#eba0ac`
   Maroon) will propagate into the flash bar bg automatically — no edit to
   `styles.go` flash-bar var declarations required. Plan 2's `GOLDEN_UPDATE=1`
   pass will refresh any goldens that capture a Warn/Err flash bar (none exist
   today; the 13 new statusbar tests in Plan 1 assert SGR presence by hex
   substring `48;2;249;226;175` for the **pre-flip** Warning bg and `48;2;243;139;168`
   for the **pre-flip** Error bg — Plan 2 must update those hex assertions to
   the new Peach/Maroon SGR substrings).

2. **`resolveLogoState` reads `m.status.Env()`** which is set once at startup and
   updated only via `SetEnv` (no event currently re-checks env mid-session). Per
   D-401 the spec calls out "persistent env failure" as an Err raise — but the
   PLAN.md NOTE in Step 2.3 explicitly defers the `EnvRecheckMsg` path to Phase
   11 if behavior shows the gap. Plan 2 does not need to add this; Plan 2 only
   adds the `profile colorprofile.Profile` field and threading.

3. **`NewAppModel(env, sopsYamlPath)` signature unchanged** in Plan 1. Plan 2
   will add the `profile` parameter; the test-only shim
   `newCleanAppModel(t)` in `internal/app/severity_test.go` calls
   `app.NewAppModel(env, "")` so Plan 2 must update that test helper to pass a
   profile (recommendation: `colorprofile.TrueColor` to match the existing test
   default).

4. **No new `internal/ui` files**. Plan 1 left `internal/ui/{chrome,crumbs,menu,
   infopanel,logo,palette}.go` untouched. Plan 2's profile-detection +
   palette-accessor work has clean ground to land on.

## Forward Deviations for Plan 10-03

1. **Bracket-fallback chip rendering (D-422) is NOT yet wired.** Plan 1 left
   `internal/ui/crumbs.go` untouched. Plan 3 reads the `Palette` value (added
   in Plan 2) and gates chip rendering on it.

2. **The 4-profile teatest matrix (D-423) does not exist yet.** Plan 1's tests
   are TrueColor-profile-only (the lipgloss/v2 test default). Plan 3's
   `Ascii`/`ANSI`/`ANSI256`/`TrueColor` matrix will need to verify the
   FlashWarnBarStyle and FlashErrBarStyle SGR bytes downsample correctly — the
   Plan-1 test assertions on hex substrings work only at TrueColor.

3. **The 13 statusbar tests added in Plan 1 hardcode TrueColor SGR hex
   substrings** (`48;2;249;226;175` etc). Plan 3's matrix must either (a)
   parameterise these by profile, or (b) keep them as TrueColor-only and add
   parallel ANSI/Ascii assertions. Recommendation (a): one switch on the active
   profile, like the existing `RequireGoldenColors` helper.

## Self-Check: PASSED

Created files:
- FOUND: /home/moersener/git/sops-tui/internal/app/severity_test.go
- FOUND: /home/moersener/git/sops-tui/internal/app/export_test.go

Modified files (verified by `git log --pretty=format`):
- FOUND: internal/ui/statusbar.go (a3fe389)
- FOUND: internal/ui/styles.go (a3fe389)
- FOUND: internal/ui/statusbar_test.go (a3fe389)
- FOUND: internal/health/checker.go (3f9cf98)
- FOUND: internal/health/checker_test.go (3f9cf98)
- FOUND: internal/ui/health.go (3f9cf98)
- FOUND: internal/ui/health_test.go (3f9cf98)
- FOUND: internal/app/model.go (3f9cf98 + 8381a11)

Commits:
- FOUND: a3fe389 (feat: FlashSeverity enum + typed flash methods + bg-tinted View)
- FOUND: 3f9cf98 (feat: severity classifier resolveLogoState + supporting accessors)
- FOUND: 8381a11 (feat: classify 27 flash callsites by severity)

Acceptance grep checks:
- FOUND: HasErrLevelFindings (1 occurrence in checker.go)
- FOUND: LastResult() (1 occurrence in health.go)
- FOUND: resolveLogoState (1 occurrence in model.go)
- FOUND: m.resolveLogoState() (2 occurrences in model.go — both RenderChrome callsites)
- FOUND: 0 occurrences of `RenderChrome.*ui\.LogoInfo` (both callsites swapped)
- FOUND: 15 FlashErr callsites
- FOUND: 12 FlashWarn callsites
- FOUND: 15 legacy Flash() callsites (the Info-aliased neutral path)
- FOUND: 42 total m.status.Flash* callsites (matches canonical map)

Goldens:
- VERIFIED: git status -- '*.golden' shows no modifications

Build / vet / test:
- VERIFIED: go build ./... exits 0
- VERIFIED: go vet ./... exits 0
- VERIFIED: go test ./... -count=1 exits 0 (full suite green)
- VERIFIED: go test ./internal/app/ -run "TestChromeASCIIOnly|TestChromeNormalBorderOnly|TestViewNoNewStyle|TestSubmodelViewsNoNewStyle" -count=1 all green
