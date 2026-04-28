---
phase: 08-header-info-panel
plan: 03
subsystem: ui, app, chrome, integration
tags: [bubbletea, lipgloss, age, go-git, chrome, infopanel, crumbs, goldens, integration]

# Dependency graph
requires:
  - phase: 08-header-info-panel/01
    provides: ui.InfoPanelData struct + RenderInfoPanel + RenderCrumbs + ParseAgeKeyFingerprint + AgeKeyFilePath + 8 package-var styles
  - phase: 08-header-info-panel/02
    provides: git.GetBranch + StatusBarModel.Segments() accessor + StatusBarModel.View() shrunk
  - phase: 07-chrome-skeleton
    provides: AppModel.View() composition (D-17), 38×6 InfoPanelPlaceholderStyle slot reservation (D-16), TestViewNoNewStyle BFS (D-22), TestChromeASCIIOnly + TestChromeNormalBorderOnly grep-gates (D-20/D-21)
  - phase: 07.1-chrome-gap-closure
    provides: 3-tier width fallback (D-116) — narrow <41 / mid 41-98 / full ≥99; full-tier path is the only one that consumes InfoPanelData
provides:
  - AppModel.infoPanel ui.InfoPanelData cached field (D-213)
  - 4 cache refresh seams: NewAppModel startup (age + sops yaml rel path), FilesDiscoveredMsg (file count + sops yaml rel path), FilesParsedMsg (recipient count), GitStatusMsg (branch + dirty)
  - ui.RenderChrome signature change: 3rd positional info InfoPanelData arg between logoStatus and width
  - crumbsHeight() flipped from `return 0` to lipgloss.Height(RenderCrumbs(...)) with first-frame guard
  - View() sections slice unconditionally inserts RenderCrumbs(m.status.Segments(), m.width) above the titled body
  - Grep-gate file scope extended for infopanel.go + crumbs.go (TestChromeASCIIOnly, TestChromeNormalBorderOnly, TestSubmodelViewsNoNewStyle)
  - 3 new integration tests in chrome_test.go: TestRenderChrome_FullTierWithInfoPanel, TestCrumbsHeight_NonZero, TestInfoPanelCacheRefresh_OnFilesDiscovered
  - 4 resize goldens regenerated (40×12, 80×24, 120×40, 200×60)
affects:
  - all chrome rendering: full-tier (≥99 cols) now shows the inflated 38×6 info panel with cfg/age/rcp/git/fil rows
  - all chrome tiers: chip-pill breadcrumb row between header and titled body
  - status bar: shrunk to right-aligned env+clipboard only at every tier (D-209/D-211)
  - 40×12 narrow-tier layout: chrome(1) + crumbs(1) + body + status(1) — body is still reachable per Phase 7.1 D-116

# Tech tracking
tech-stack:
  - charm.land/bubbletea/v2: tea.WindowSizeMsg in TestCrumbsHeight_NonZero
  - charm.land/lipgloss/v2: lipgloss.Height for crumbsHeight() flip
  - filippo.io/age v1.3.1: ParseAgeKeyFingerprint reads identity at startup (NewAppModel)
  - go-git/go-git/v5: gitpkg.GetBranch called in GitStatusMsg handler

# Plan summary
status: complete
date: 2026-04-28

## Tasks delivered

1. **08-03-T1: Chrome signature + AppModel cache + 4 refresh seams + crumbsHeight flip**
   - `internal/ui/chrome.go`: `RenderChrome` gains `info InfoPanelData` 3rd positional arg between `logoStatus` and `width`. Mid-tier and narrow-tier paths ignore it. Full-tier path replaces `InfoPanelPlaceholderStyle.Render("")` with `RenderInfoPanel(info)` (slot stays 38×6).
   - `internal/app/model.go`: added `infoPanel ui.InfoPanelData` field. `NewAppModel` populates `AgeFingerprint` (via `ui.ParseAgeKeyFingerprint(ui.AgeKeyFilePath())`) + `SopsYamlRelPath` (via `deriveSopsYamlRelPath(m.sopsYamlPath)`). `FilesDiscoveredMsg` handler refreshes `FileCount` + `SopsYamlRelPath`. `FilesParsedMsg` handler refreshes `RecipientCount` from `m.currentParsed.Metadata.AgeRecipients`. `GitStatusMsg` handler calls `gitpkg.GetBranch(m.gitRepoRoot)` and refreshes `GitBranch`/`GitDetached`/`GitDirty`.
   - `internal/app/model.go`: `crumbsHeight(m)` flipped from `return 0` to `lipgloss.Height(ui.RenderCrumbs(m.status.Segments(), m.width))` with `m.width == 0` first-frame guard returning 0.
   - `internal/app/model.go`: View() sections slice — `""` placeholder replaced with `ui.RenderCrumbs(m.status.Segments(), m.width)`.
   - Both `RenderChrome` call-sites updated atomically (line 1397 in `View()` and line 1577 in `chromeHeight()`); `grep -c "RenderChrome(" internal/app/model.go` returns exactly 2 and both contain `m.infoPanel`.

2. **08-03-T2: Grep-gate file scope + 3 integration tests**
   - `internal/app/chrome_test.go`: file-scope arrays for `TestChromeASCIIOnly` and `TestChromeNormalBorderOnly` extended to include `internal/ui/infopanel.go` + `internal/ui/crumbs.go` with skip-if-missing guard.
   - `internal/app/chrome_test.go`: 3 new tests added — `TestRenderChrome_FullTierWithInfoPanel` (asserts info panel content present in full-tier output at width=200), `TestCrumbsHeight_NonZero` (asserts `crumbsHeight(m) > 0` after first WindowSizeMsg + breadcrumb set), `TestInfoPanelCacheRefresh_OnFilesDiscovered` (asserts `m.infoPanel.FileCount` reflects new file count after handler runs).
   - `internal/ui/submodel_view_no_newstyle_test.go`: `submodelFiles` allowlist gains `infopanel.go` + `crumbs.go` per D-219.
   - ASCII allowlist already includes U+2026 (`…`) per Phase 7 — no allowlist change needed (verified).

3. **08-03-T3: Resize goldens regen at 4 widths**
   - `internal/app/testdata/resize_40x12.golden`: regenerated. Narrow tier inflates from `chrome stub + body + status (multi-section pipes)` to `chrome stub + crumbs row + body + shrunk status`. Layout = chrome(1) + crumbs(1) + body + status(1) — body is still reachable per Phase 7.1 D-116.
   - `internal/app/testdata/resize_80x24.golden`: regenerated. Mid tier — crumb row inflates from `""` placeholder to `<sops-tui> <files>` chip pills.
   - `internal/app/testdata/resize_120x40.golden`: regenerated. Full tier — both info-panel content (5 rows: `cfg: -` / `age: -` / `rcp: -` / `git: -` / `fil: -`) AND crumb row inflate.
   - `internal/app/testdata/resize_200x60.golden`: regenerated. Full tier with all new style vars.
   - `internal/app/resize_test.go`: extended to set `t.Setenv("SOPS_AGE_KEY_FILE", testdata/keys.txt)` for deterministic age fingerprint and to assert no host paths leaked into goldens.
   - Verified: `! grep -q "/home/" internal/app/testdata/resize_*.golden` returns 0 lines (no host path leaks).

## Build & test verification

```
go build ./...                                     # passes
go vet ./internal/...                              # clean
go test ./... -count=1                             # full suite green across 8 packages
go test ./internal/app/ -run 'TestChromeASCIIOnly|TestChromeNormalBorderOnly|TestViewNoNewStyle|TestRenderChrome_FullTierWithInfoPanel|TestCrumbsHeight_NonZero|TestInfoPanelCacheRefresh_OnFilesDiscovered' -v   # 6/6 PASS
go test ./internal/app/ -run TestResize -count=1   # 4/4 PASS (post-regen)
grep -c "RenderChrome(" internal/app/model.go      # 2
grep -c "m.infoPanel" internal/app/model.go        # 12 (field decl + 4 refresh seams + 2 RenderChrome call-sites + 5 sub-field updates)
grep -c "ansi.TruncateLeft" internal/ui/infopanel.go  # 1 (Plan 1 — verified upstream still in place)
grep -c "AGE-SECRET-KEY" internal/ui/agekey.go     # 0 (Plan 1 — verified upstream still in place)
```

# D-220 Security Review (5 fields × 5 questions)

> **Mandate:** "5-question security review per new info-panel field, signed off in 08-03-SUMMARY.md."
> Each row = one info-panel field. Each column = one of the 5 D-220 questions. Cell = mitigation evidence + residual risk.

| Field | Q1: Derives from private key material? | Q2: Could expose absolute filesystem paths? | Q3: Does any keybinding copy this field to clipboard? | Q4: Does this field appear in stderr logs? | Q5: Could a screenshot narrow an attacker's search space? |
|-------|---------------------------------------|--------------------------------------------|------------------------------------------------------|--------------------------------------------|----------------------------------------------------------|
| **cfg:** (.sops.yaml relative path) | NO. Path string only — no key material involved. | MITIGATED. `deriveSopsYamlRelPath()` calls `filepath.Rel(cwd, sopsYamlPath)`; resolves to repo-relative segments only. Goldens verified: `! grep -q "/home/" internal/app/testdata/resize_*.golden`. Empty marker `-` rendered if path empty. | NO. Chrome is display-only — no copy bindings target chrome content (Pitfall 11). | NO. `internal/ui/errorbox.go` is the only stderr surface; chrome content is not logged. | LOW. Repo name + relative path could narrow project context, but no secrets. Residual risk: posting a screenshot reveals project layout; acceptable for typical screenshot/recording scenarios. |
| **age:** (age fingerprint, ≤10 cells) | YES — derived from `~/.config/sops/age/keys.txt` via `filippo.io/age` `ParseIdentities`. MITIGATED via two layers: (1) `agekey.go` type-asserts to `*age.X25519Identity` and calls `Recipient().String()` (returns the *public* `age1...` recipient string, NEVER `Identity.String()` which would leak `AGE-SECRET-KEY-...` private key prefix per RESEARCH.md Pitfall A). (2) `RenderInfoPanel` then `middleTruncate` to ≤10 cells with U+2026 ellipsis (e.g. `age1abc…xyz`). Acceptance test: `grep -c "AGE-SECRET-KEY" internal/ui/agekey.go` returns 0. | NO. The path to the keyfile (`~/.config/sops/age/keys.txt` or `$SOPS_AGE_KEY_FILE`) is NEVER rendered — only the parsed Recipient string. | NO. Chrome is display-only. | NO. Parse errors result in empty fingerprint + `-` empty marker render; no error propagation to stderr from this path. | LOW. Truncated 10-cell fingerprint (start + `…` + end) reveals ~30 bits of the public key. Brute-forcing the matching private key is computationally infeasible (Curve25519 ≈ 2^252). Residual risk: an attacker could correlate the partial fingerprint with public key registries (none exist for age keys). Acceptable. |
| **rcp:** (recipient count, integer) | NO. Integer count only. | NO. No path content. | NO. Chrome is display-only. | NO. Integer derived in handler; no log emission. | NEGLIGIBLE. Reveals "this file has N recipients" — useful for the user but tells an attacker nothing about WHO the recipients are. |
| **git:** (branch + dirty marker) | NO. Branch name only. | NO. Branch name is not a path. | NO. Chrome is display-only. | NO. `gitpkg.GetBranch` errors degrade silently to `-` empty marker; non-git directory returns `ErrRepositoryNotExists` and is handled in the handler without logging. | LOW-MEDIUM. Branch names can leak project structure (e.g., `feat/auth-redesign`, `release/v2.1.0-beta`). Residual risk acknowledged: this is the same risk as running `git status` in a screenshot. Acceptable for the standard k9s-style "current context" surface. |
| **fil:** (file count, integer) | NO. Integer count only. | NO. No path content. | NO. Chrome is display-only. | NO. Integer derived in handler. | NEGLIGIBLE. Reveals "this project has N SOPS files" — coarse-grained signal, not exploitable. |

**Sign-off:** All five questions answered for all five fields. Highest risk row is `age:` (Q1 — derives from private key material) — mitigated via two-layer defense (type-assert prevents private-key leak; truncation prevents full public-key disclosure). All other risks are LOW or NEGLIGIBLE.

**Approval:** 2026-04-28 — D-220 security review complete; Phase 8 cleared for verification.

## Files modified

- `internal/ui/chrome.go` — `RenderChrome` signature change
- `internal/app/model.go` — `infoPanel` field + 4 refresh seams + `crumbsHeight` flip + View() sections slice update + 2 `RenderChrome` call-site updates
- `internal/app/chrome_test.go` — file-scope arrays extended; 3 new integration tests added
- `internal/ui/chrome_test.go` — assertions adjusted for new chrome shape (full-tier)
- `internal/ui/submodel_view_no_newstyle_test.go` — `submodelFiles` allowlist gains `infopanel.go` + `crumbs.go`
- `internal/app/resize_test.go` — `t.Setenv("SOPS_AGE_KEY_FILE", ...)` for deterministic age fingerprint; host-path leak guard
- `internal/app/testdata/resize_40x12.golden` — regenerated (narrow tier WILL change per RESEARCH.md finding 5)
- `internal/app/testdata/resize_80x24.golden` — regenerated (mid tier — crumb row inflated)
- `internal/app/testdata/resize_120x40.golden` — regenerated (full tier — info panel + crumb row both inflated)
- `internal/app/testdata/resize_200x60.golden` — regenerated (full tier — info panel + crumb row both inflated)

## Acceptance criteria — all PASS

| Criterion | Evidence |
|-----------|----------|
| Build passes | `go build ./...` exits 0 |
| Full test suite green | `go test ./... -count=1` — 8 packages OK |
| Both `RenderChrome` call-sites updated atomically | `grep -c "RenderChrome(" internal/app/model.go` = 2; both lines contain `m.infoPanel` |
| `crumbsHeight` returns non-zero after WindowSizeMsg | `TestCrumbsHeight_NonZero` PASSES |
| Info panel renders in full tier | `TestRenderChrome_FullTierWithInfoPanel` PASSES; goldens at 120×40 and 200×60 contain `cfg:` / `age:` / `rcp:` / `git:` / `fil:` rows |
| Cache refreshes on FilesDiscoveredMsg | `TestInfoPanelCacheRefresh_OnFilesDiscovered` PASSES |
| Grep-gate file scope extended | `internal/app/chrome_test.go` lists `internal/ui/{infopanel,crumbs}.go`; `submodelFiles` includes both |
| ASCII-only chrome | `TestChromeASCIIOnly` PASSES (U+2026 already in allowlist from Phase 7) |
| NormalBorder() exclusivity | `TestChromeNormalBorderOnly` PASSES |
| Zero `lipgloss.NewStyle()` in renderer reachables | `TestViewNoNewStyle` + `TestSubmodelViewsNoNewStyle` PASS — package-var styles only |
| 40×12 narrow tier layout still reachable | `resize_40x12.golden` shows `press ? for help / <chip row> / titled body / status` — body region present, body-reachable contract preserved |
| No host paths in goldens | `! grep -q "/home/" internal/app/testdata/resize_*.golden` returns 0 lines |
| D-220 5×5 security review signed off | Table above — all 25 cells filled; sign-off line approved 2026-04-28 |
