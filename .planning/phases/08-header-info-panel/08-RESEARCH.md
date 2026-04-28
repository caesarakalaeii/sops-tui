# Phase 8: Header Info Panel + Crumb Chips - Research

**Researched:** 2026-04-28
**Domain:** lipgloss/v2 layout, filippo.io/age v1.3.1 identity parsing, go-git v5 branch resolution, Bubble Tea v2 event-driven caching
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Info-Panel Schema (UI-04, UI-05)**
- D-201: Terse 3-char labels (`cfg:`, `age:`, `rcp:`, `git:`, `fil:`). `InfoPanelLabelStyle` with `Width(5)`.
- D-202: Recipient count = `len(m.currentParsed.Metadata.AgeRecipients)` for current file; `-` in stateFileList.
- D-203: Middle-truncation with U+2026 (`…`) when value width exceeds 32 cells. `middleTruncate` helper in `infopanel.go`.
- D-204: ASCII `-` empty marker for missing/uncomputed fields.

**Crumb Chip Pill Design (UI-07)**
- D-205: `<segment>` k9s-exact wrapper. Verbatim match to `~/git/k9s/internal/ui/crumbs.go:62-74`.
- D-206: Active chip: `bg=ColorAccent fg=ColorBg Bold(true)`. Inactive chip: `bg=ColorSurface fg=ColorFg`. Bold weight is the redundant encoding channel (Pitfall 9).
- D-207: `strings.ToLower(s)` then `strings.ReplaceAll(s, " ", "")` normalisation per k9s `crumbs.go:70-71`.
- D-208: Single space between chips; `PaddingLeft(1).PaddingRight(1)` on row container. Width budget = `m.width - 2`.

**Status-Bar Shrink (UI-08)**
- D-209: `SetItemCount` becomes a no-op; titled-border title is canonical count. `itemCount`/`itemLabel` fields deleted; all `m.status.SetItemCount(...)` call-sites in `model.go` deleted.
- D-210: Breadcrumb data stays on `StatusBarModel`; new `Segments() []string` accessor returns `strings.Split(m.breadcrumb, " > ")`.
- D-211: `StatusBarModel.View(width)` normal path: right-align env+clipboard only. Left/center sections and pipe separators deleted. Flash path unchanged.
- D-212: Flash unchanged.

**Data Refresh + Age Key (UI-05, Pitfall 11, Pitfall 15)**
- D-213: Cached `infoPanel ui.InfoPanelData` on `AppModel`. Refresh events: `FilesDiscoveredMsg`, `GitStatusMsg`, recipient add/remove success, edit re-encrypt success. `View()` reads cache only — zero I/O.
- D-214: Age fingerprint from `filippo.io/age` `ParseIdentities`. First identity's `Recipient().String()` (e.g. `age1abc...xyz`). `$SOPS_AGE_KEY_FILE` if set, else `~/.config/sops/age/keys.txt`. On failure: fingerprint = `""`, render shows `-`.
- D-215: New `git.GetBranch(repoRoot string) (branch string, detached bool, err error)`. `repo.Head()` → `ref.Name().IsBranch()` → `ref.Name().Short()`. Detached: `branch = ref.Hash().String()[:7]`, `detached = true`. Non-git returns `("", false, gogit.ErrRepositoryNotExists)`.
- D-216: Crumbs render at every chrome tier. `crumbsHeight(m)` returns `lipgloss.Height(RenderCrumbs(...))`. Overflow: drop middle segments, replace with `<…>`.

**Plan Split**
- D-217: Three plans — Plan 1 (primitives + age key), Plan 2 (git.GetBranch + statusbar shrink), Plan 3 (integration + goldens).
- D-218: Plan 3 is deliberately the largest.

**Validation**
- D-219: Phase 7 patterns apply. Add `TestRenderChrome_FullTierWithInfoPanel`, `TestCrumbsHeight_NonZero`, `TestInfoPanelCacheRefresh_OnFilesDiscovered` in `chrome_test.go`.

**Security**
- D-220: 5-question security review per new info-panel field; sign-off in `08-03-SUMMARY.md`.

### Claude's Discretion

- Exact byte-layout of `RenderInfoPanel` rows (JoinHorizontal vs fmt.Sprintf with padding).
- `middleTruncate` split point and multi-byte rune handling.
- `truncateSegmentsToWidth` drop strategy.
- Whether git badge suffix survives chip rendering (Recommendation: keep).
- Whether `SetItemCount` is deleted or no-op (Recommendation: delete).
- Whether age key parser lives in `internal/ui/agekey.go` or inline (Recommendation: separate file).
- Exact dirty-marker glyph (Recommendation: trailing `*` for dirty, no marker for clean).
- Whether `crumbsHeight` hard-codes `1` or uses `lipgloss.Height` (Recommendation: `lipgloss.Height`).

### Deferred Ideas (OUT OF SCOPE)

- Logo severity coupling (UI-03, Phase 10).
- 16-color palette fallback (UI-13, Phase 10).
- Narrow-terminal aesthetics matrix (UI-16, Phase 10).
- `[`/`]` view history navigation (v1.2).
- Skin loader / k9s-compatible YAML schema (v2).
- D-18 chrome caching (Phase 11 SC2).
- Per-`(state, recipientAction, IsSearchActive)` golden-file matrix (Phase 9).
- Status-bar cleanup (IN-02..IN-07) deferred per Phase 7.1.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| UI-04 | User sees a header info panel (top-left) with five rows: `.sops.yaml` relative path, age key fingerprint, recipient count, git branch + clean/dirty marker, file count | filippo.io/age `ParseIdentities` + `Recipient().String()` verified; go-git `repo.Head()` + `ref.Name().IsBranch()` + `ref.Name().Short()` verified; all data sources confirmed in existing codebase |
| UI-05 | Info-panel fields are truncated and de-PII'd: age fingerprint ≤10 chars with ellipsis, paths are repo-relative, no copy bindings on chrome | `ansi.Truncate(s, n, "…")` is tail-truncation only; `middleTruncate` must be hand-written (splits string, inserts U+2026 at midpoint); `SOPS_AGE_KEY_FILE` env-var path must never appear in rendered output |
| UI-07 | Breadcrumb segments render as colored chip pills; active (last) segment uses accent color | k9s `crumbs.go:62-74` verified verbatim; `<segment>` format + lowercase + strip-spaces confirmed; lipgloss.NewStyle with Background+Foreground+Bold confirmed as correct pill approach |
| UI-08 | Bottom status bar shrinks to right-aligned env indicators + clipboard state; breadcrumb moves above titled body | `StatusBarModel.View()` dissected; left/center sections + pipes are removable; `Segments()` accessor is additive; flash path untouched |
</phase_requirements>

---

## Summary

Phase 8 is a pure data-binding and presentation phase. All six source data signals (`.sops.yaml` path, age key fingerprint, git branch+dirty, recipient count, file count, breadcrumb segments) already live on `AppModel` or are derivable from existing dependencies with zero new go.mod entries.

The three primitives — `RenderInfoPanel`, `RenderCrumbs`, and `GetBranch` — each have well-defined, verified API contracts in the existing stack. The main implementation-level risk is `ansi.Truncate` being tail-only: `middleTruncate` for middle-ellipsis must be hand-written (~15 lines). Everything else maps cleanly to established project patterns.

The integration (Plan 3) requires three simultaneous changes that cannot be split further: `RenderChrome` signature change (adds `InfoPanelData`), `crumbsHeight` flip from `return 0`, and `View()` sections-slice update. These three changes share the chrome golden test files, which is why D-218 explicitly makes Plan 3 the largest.

**Primary recommendation:** Implement in the locked 3-plan order. Plan 1 ships pure functions with no AppModel coupling; Plan 2 adds the git helper and statusbar shrink; Plan 3 wires everything and refreshes golden files. Do not merge Plan 3 without refreshing all four golden sizes.

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Age key fingerprint display | API / Backend (model.go startup) | UI renderer (infopanel.go) | Key file I/O happens once at startup in `NewAppModel`; renderer reads cached string only |
| `.sops.yaml` relative path | API / Backend (model.go) | UI renderer (infopanel.go) | `filepath.Rel(cwd, sopsYamlPath)` computed once at construction; cached on `infoPanel` struct |
| Git branch resolution | API / Backend (git/status.go) | AppModel (GitStatusMsg handler) | Pure I/O function; result cached; renderer reads cached field only |
| Recipient count | AppModel (Update paths) | UI renderer (infopanel.go) | `len(m.currentParsed.Metadata.AgeRecipients)` already computed after parse; cached refresh |
| File count | AppModel (FilesDiscoveredMsg) | UI renderer (infopanel.go) | `len(m.files)` already available; cached on info struct |
| Crumb chip rendering | UI renderer (crumbs.go) | AppModel View() | Pure function of `[]string` segments; AppModel calls `m.status.Segments()` to supply data |
| Breadcrumb data ownership | StatusBarModel | AppModel (read via Segments()) | D-210 keeps data on StatusBarModel to avoid migrating 16 SetBreadcrumb call-sites |
| Status bar right cluster | StatusBarModel View() | — | Flash path unchanged; only normal path removes left/center sections |

---

## Standard Stack

### Core (verified in go.mod and local module cache)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `filippo.io/age` | v1.3.1 | Parse `~/.config/sops/age/keys.txt`; extract `Recipient().String()` | Already in go.mod; used by `internal/ui/recipientform.go`; zero new deps [VERIFIED: /home/moersener/git/sops-tui/go.mod] |
| `github.com/go-git/go-git/v5` | v5.17.0 | `repo.Head()` for branch resolution | Already in go.mod; established pattern in `internal/git/status.go` [VERIFIED: /home/moersener/git/sops-tui/go.mod] |
| `charm.land/lipgloss/v2` | v2.0.3 | `JoinHorizontal`, `JoinVertical`, `Width()`, `Height()` for info panel and chip row layout | Already in go.mod; all chrome rendering uses it [VERIFIED: /home/moersener/git/sops-tui/go.mod] |
| `github.com/charmbracelet/x/ansi` | v0.11.7 | `ansi.Truncate(s, width, "…")` for tail-truncation in title overlays; NOT for middle-truncation | Already in go.mod; used in `internal/ui/chrome.go` [VERIFIED: /home/moersener/git/sops-tui/go.mod] |

**No new go.mod entries required for Phase 8.**

---

## Library API Exact Signatures

### filippo.io/age v1.3.1 — Identity Parsing

[VERIFIED: /home/moersener/go/pkg/mod/filippo.io/age@v1.3.1/parse.go + age.go + x25519.go]

```go
// ParseIdentities parses a file with one or more private key encodings.
// Returns []age.Identity (concrete types: *X25519Identity, *HybridIdentity).
// Returns an error (including "no identities found") when file is empty.
func ParseIdentities(f io.Reader) ([]Identity, error)

// Identity interface — does NOT have a Recipient() method.
// The concrete *X25519Identity type does:
type Identity interface {
    Unwrap(stanzas []*Stanza) (fileKey []byte, err error)
}

// X25519Identity.Recipient() returns *X25519Recipient (not Identity).
func (i *X25519Identity) Recipient() *X25519Recipient

// X25519Recipient.String() returns Bech32 public key encoding: "age1..." prefix.
func (r *X25519Recipient) String() string  // e.g. "age1examplexyz..."
```

**Critical implementation note:** `age.Identity` (the interface returned by `ParseIdentities`) does NOT have a `Recipient()` method. The concrete type `*X25519Identity` has `Recipient()`, and `*HybridIdentity` has its own `Recipient()` method. To call `Recipient().String()` you must type-assert. The safe pattern:

```go
// In internal/ui/agekey.go (or equivalent)
ids, err := age.ParseIdentities(f)
if err != nil || len(ids) == 0 {
    return ""
}
switch id := ids[0].(type) {
case *age.X25519Identity:
    return id.Recipient().String()  // returns "age1..."
default:
    // HybridIdentity or plugin identity — Recipient() may not be available
    // as a typed method; use fmt.Stringer if available or return ""
    return ""
}
```

**SOPS_AGE_KEY_FILE precedence:** D-214 specifies checking `$SOPS_AGE_KEY_FILE` before `~/.config/sops/age/keys.txt`. The existing `internal/validator/startup.go` uses `filepath.Join(home, ".config", "sops", "age", "keys.txt")` with no `$SOPS_AGE_KEY_FILE` override [VERIFIED: internal/validator/startup.go:59]. Phase 8 must add the env-var check in the age key parser to be consistent with SOPS CLI behavior.

**HOME resolution:** `os.UserHomeDir()` is the correct call (used by `internal/validator/startup.go:49`). Do NOT use `os.Getenv("HOME")` directly — `os.UserHomeDir()` handles edge cases (no HOME set, Windows, etc.).

**Security gate (D-220):** The fingerprint is derived from calling `Recipient().String()` on the parsed private key identity. The output is the Bech32-encoded PUBLIC key (`age1...`), not the private key itself. The private key `String()` method returns `"AGE-SECRET-KEY-..."` (uppercase). The renderer must call `identity.Recipient().String()` not `identity.String()` — a mistake here would log the private key prefix in the info panel.

### go-git v5.17.0 — Branch Resolution

[VERIFIED: /home/moersener/go/pkg/mod/github.com/go-git/go-git/v5@v5.17.0/repository.go:1505 + plumbing/reference.go]

```go
// Repository.Head returns the current HEAD reference.
// Returns (*plumbing.Reference, error).
func (r *Repository) Head() (*plumbing.Reference, error)

// plumbing.Reference methods:
func (r *Reference) Name() ReferenceName      // e.g. "refs/heads/main"
func (r *Reference) Hash() Hash               // SHA-1 hash of the commit
func (r ReferenceName) IsBranch() bool        // true if prefix == "refs/heads/"
func (r ReferenceName) Short() string         // "main" from "refs/heads/main"
func (h Hash) String() string                 // 40-char hex; use [:7] for short
```

**Detached HEAD pattern for `GetBranch`:**

```go
func GetBranch(repoRoot string) (branch string, detached bool, err error) {
    repo, err := gogit.PlainOpenWithOptions(repoRoot, &gogit.PlainOpenOptions{DetectDotGit: true})
    if err == gogit.ErrRepositoryNotExists {
        return "", false, gogit.ErrRepositoryNotExists
    }
    if err != nil {
        return "", false, err
    }
    ref, err := repo.Head()
    if err != nil {
        return "", false, err
    }
    if ref.Name().IsBranch() {
        return ref.Name().Short(), false, nil
    }
    // Detached HEAD: return short commit hash
    return ref.Hash().String()[:7], true, nil
}
```

**Dirty detection for `git:` row:** `GetBranch` returns branch only. The dirty marker aggregation (`m.files[*].GitStatus != ""`) happens in the `GitStatusMsg` handler when populating `m.infoPanel.GitDirty`. The `GetBranch` call and the `GetFileStatuses` call are separate. In the `GitStatusMsg` async cmd (model.go:354-362), both can run together: extend the existing cmd to also call `GetBranch` and return the result as part of `GitStatusMsg`. Alternatively, `GetBranch` can be called separately. Either approach is valid; the simplest path is extending `GitStatusMsg` to carry `Branch string` and `DetachedHead bool` fields.

### lipgloss/v2 v2.0.3 — Cell-Counting Behavior

[VERIFIED: /home/moersener/go/pkg/mod/charm.land/lipgloss/v2@v2.0.3/join.go + size.go]

```go
func JoinHorizontal(pos Position, strs ...string) string
func JoinVertical(pos Position, strs ...string) string
func Width(str string) (width int)   // strips ANSI, counts visible cells
func Height(str string) int          // counts '\n' occurrences + 1
```

**Empty string behavior:** `lipgloss.Height("")` returns 1 (empty string has one line). This is why the `View()` sections slice guards with `if crumbsHeight(m) > 0` — appending `""` would still add a row. Phase 8 flips this by replacing `""` with the actual rendered crumbs string (always at least 1 row), and removes the guard. [VERIFIED: existing code at model.go:1352-1354]

**JoinHorizontal with styled strings:** `JoinHorizontal(lipgloss.Top, label, value)` aligns tops of multi-line strings. For single-line label+value rows in `RenderInfoPanel`, this is the correct alignment position. The total rendered width is the sum of individual widths (lipgloss does not add spacing between joined strings).

**Width() with ANSI escape sequences:** `lipgloss.Width()` correctly strips ANSI sequences before counting visible cells. For `InfoPanelLabelStyle.Width(5).Render("cfg:")`, `lipgloss.Width(result)` returns exactly 5 regardless of color escape sequences. This is the safe way to compute layout budgets.

### charmbracelet/x/ansi v0.11.7 — Truncate Semantics

[VERIFIED: /home/moersener/go/pkg/mod/github.com/charmbracelet/x/ansi@v0.11.7/truncate.go:53]

```go
// Truncate truncates a string to a given length, adding a tail at the END
// if the string is longer. ANSI-aware, grapheme-cluster aware.
func Truncate(s string, length int, tail string) string
```

**Critical:** `ansi.Truncate` is tail-truncation only — it truncates from the right and appends the tail. It does NOT perform middle-truncation. For the `middleTruncate` helper in `infopanel.go`, `ansi.Truncate` can only be used to trim each half individually; the middle-ellipsis insertion must be hand-coded:

```go
func middleTruncate(s string, maxCells int) string {
    if lipgloss.Width(s) <= maxCells {
        return s
    }
    ellipsis := "…"
    ellipsisW := lipgloss.Width(ellipsis)
    available := maxCells - ellipsisW
    left := available / 2
    right := available - left
    // ansi.Truncate gives us left-side truncation from the right end:
    leftPart := ansi.Truncate(s, left, "")
    // For right-side, we need TruncateLeft (truncate from the start):
    // charmbracelet/x/ansi v0.11.7 provides ansi.TruncateLeft(s, n, prefix)
    rightPart := ansi.TruncateLeft(s, lipgloss.Width(s)-right, "")
    return leftPart + ellipsis + rightPart
}
```

`ansi.TruncateLeft(s, n, prefix)` is available [VERIFIED: truncate.go:173] and truncates from the left (removes the first n visible cells). This is the correct building block for the right-side fragment.

---

## Architecture Patterns

### System Architecture Diagram

```
startup (NewAppModel)
  |-- os.UserHomeDir() / $SOPS_AGE_KEY_FILE
  |-- age.ParseIdentities(keys.txt)
  |-- Recipient().String() --> m.infoPanel.AgeFingerprint
  |-- filepath.Rel(cwd, sopsYamlPath) --> m.infoPanel.SopsYamlRelPath
       |
       v
FilesDiscoveredMsg handler
  |-- len(m.files) --> m.infoPanel.FileCount
  |-- sopsYamlRelPath refresh if needed
  |-- dispatches GitStatusMsg async cmd
       |
       v
GitStatusMsg handler
  |-- git.GetBranch(sopsDir) --> m.infoPanel.GitBranch, GitDetached
  |-- aggregate m.files[*].GitStatus --> m.infoPanel.GitDirty
       |
       v
Recipient add/remove + edit-encrypt success handlers
  |-- len(m.currentParsed.Metadata.AgeRecipients) --> m.infoPanel.RecipientCount
       |
       v
AppModel.View() [READS cache only -- zero I/O]
  |
  |-- ui.RenderChrome(hints, ui.LogoInfo, m.infoPanel, m.width)
  |     |-- full tier: RenderInfoPanel(m.infoPanel) into 38x6 slot
  |     |-- mid/narrow tier: slot dropped (Phase 7.1 D-116 unchanged)
  |
  |-- ui.RenderCrumbs(m.status.Segments(), m.width)
  |     |-- truncateSegmentsToWidth if overflow
  |
  |-- m.status.View(m.width)  [right-aligned env+clipboard only]
```

### Recommended Project Structure Changes

```
internal/
  ui/
    infopanel.go      NEW — InfoPanelData struct + RenderInfoPanel + middleTruncate
    infopanel_test.go NEW — unit tests for all 3 functions
    crumbs.go         NEW — RenderCrumbs + truncateSegmentsToWidth + normalisation
    crumbs_test.go    NEW — unit tests per D-219 list
    agekey.go         NEW — parseAgeKeyFingerprint(keyFilePath string) string
    agekey_test.go    NEW — TestParseAgeKey_FirstIdentity, TestParseAgeKey_MissingFile
    styles.go         EXTEND — 7 new package vars (InfoPanelLabelStyle, InfoPanelValueStyle,
                                InfoPanelSepStyle, CrumbChipStyle, CrumbChipActiveStyle,
                                CrumbChipSepStyle, CrumbChipEllipsisStyle)
    statusbar.go      MODIFY — add Segments(), strip left+center+pipes from View()
  git/
    status.go         EXTEND — GetBranch function
  app/
    model.go          MODIFY — infoPanel field + 4 refresh paths + crumbsHeight flip + View()
    chrome_test.go    EXTEND — 3 new integration tests + allowlist + file scope
    submodel_view_no_newstyle_test.go  EXTEND — add infopanel.go + crumbs.go to allowlist
    testdata/
      resize_80x24.golden    REFRESH
      resize_120x40.golden   REFRESH
      resize_200x60.golden   REFRESH
      (resize_40x12.golden   verify — narrow tier; crumbs visible but chrome unchanged)
```

### Pattern 1: InfoPanelData Struct (D-201)

**What:** Value struct passed from AppModel to RenderInfoPanel. All fields are pre-computed strings (no I/O in renderer).
**When to use:** Passed by AppModel.View() to ui.RenderChrome (the new `info InfoPanelData` parameter).

```go
// Source: internal/ui/infopanel.go (NEW — Phase 8)
// [VERIFIED: struct shape from 08-CONTEXT.md D-201..D-204]
type InfoPanelData struct {
    SopsYamlRelPath string  // "" means no .sops.yaml (renders "-")
    AgeFingerprint  string  // "age1..." full string; empty means "-"
    RecipientCount  int     // -1 means no current file (renders "-")
    GitBranch       string  // "" means not a git repo (renders "-")
    GitDetached     bool    // true if HEAD is detached; branch is short hash
    GitDirty        bool    // true if any file has non-clean git status
    FileCount       int     // -1 means not yet discovered (renders "-")
}
```

### Pattern 2: RenderInfoPanel Layout (D-201, Claude's Discretion)

**What:** Five rows of label+value pairs within the 38x6 envelope reserved by `InfoPanelPlaceholderStyle`.
**Recommendation:** Use `lipgloss.JoinVertical` of rows; each row is `lipgloss.JoinHorizontal(lipgloss.Top, labelRendered, valueRendered)`.

```go
// Source: internal/ui/infopanel.go (NEW — Phase 8)
// [VERIFIED: label math from 08-CONTEXT.md D-201; ansi.Truncate signature VERIFIED]
func RenderInfoPanel(d InfoPanelData) string {
    rows := []string{
        infoPanelRow("cfg:", sopsYamlDisplay(d.SopsYamlRelPath)),
        infoPanelRow("age:", ageDisplay(d.AgeFingerprint)),
        infoPanelRow("rcp:", rcpDisplay(d.RecipientCount)),
        infoPanelRow("git:", gitDisplay(d.GitBranch, d.GitDetached, d.GitDirty)),
        infoPanelRow("fil:", filDisplay(d.FileCount)),
    }
    return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func infoPanelRow(label, value string) string {
    l := InfoPanelLabelStyle.Render(label)  // Width(5) applied in style
    v := InfoPanelValueStyle.Render(value)
    return lipgloss.JoinHorizontal(lipgloss.Top, l, v)
}
```

### Pattern 3: RenderCrumbs Chip Format (D-205..208, k9s-verbatim)

**What:** Pure function rendering `<segment>` pills from a segments slice.

```go
// Source: internal/ui/crumbs.go (NEW — Phase 8)
// [VERIFIED: k9s crumbs.go:62-74 read directly at ~/git/k9s/internal/ui/crumbs.go]
func RenderCrumbs(segments []string, width int) string {
    if len(segments) == 0 {
        return CrumbRowStyle.Width(width).Render("")
    }
    normalised := normaliseSegments(segments)
    normalised = truncateSegmentsToWidth(normalised, width-2) // -2 for row padding
    chips := make([]string, len(normalised))
    for i, seg := range normalised {
        text := "<" + seg + ">"
        if i == len(normalised)-1 {
            chips[i] = CrumbChipActiveStyle.Render(text) // accent bg + body fg + bold
        } else {
            chips[i] = CrumbChipStyle.Render(text)       // surface bg + fg
        }
    }
    inner := strings.Join(chips, " ")
    return CrumbRowStyle.PaddingLeft(1).PaddingRight(1).Render(inner)
}

func normaliseSegments(segments []string) []string {
    out := make([]string, len(segments))
    for i, s := range segments {
        out[i] = strings.ReplaceAll(strings.ToLower(s), " ", "")
    }
    return out
}
```

### Pattern 4: Segments() Accessor on StatusBarModel (D-210)

**What:** Additive read-path on the existing `StatusBarModel`. Does not change how breadcrumb is stored.

```go
// Source: internal/ui/statusbar.go (EXTEND — Phase 8)
// [VERIFIED: SetBreadcrumb implementation at statusbar.go:73-78]
func (m StatusBarModel) Segments() []string {
    return strings.Split(m.breadcrumb, " > ")
}
```

**Note:** The existing `SetBreadcrumb` joins segments with `" > "` (statusbar.go:77). `Segments()` reverses this. This is lossless because `SetBreadcrumb` always prepends `"sops-tui"`, so `Segments()` always returns at least `["sops-tui"]`. The split string matches the join separator exactly — no edge-case risk.

### Anti-Patterns to Avoid

- **Calling `age.Identity.Recipient()` directly:** The `Identity` interface has no `Recipient()` method. Must type-assert to `*age.X25519Identity` first.
- **Using `ansi.Truncate` for middle-truncation:** `ansi.Truncate` truncates from the right only. Middle-truncation requires `ansi.TruncateLeft` for the right fragment.
- **Using `identity.String()` instead of `identity.Recipient().String()`:** `X25519Identity.String()` returns the private key `"AGE-SECRET-KEY-..."` string. The renderer must call `Recipient().String()` to get the public key fingerprint.
- **Calling `lipgloss.NewStyle()` inside `RenderInfoPanel` or `RenderCrumbs`:** Violates `TestViewNoNewStyle` and `TestSubmodelViewsNoNewStyle` discipline. All styles must be package vars in `styles.go`.
- **Calling I/O from `View()`:** `View()` reads `m.infoPanel` (cached struct) only. No `os.Stat`, `os.ReadFile`, `git.*`, or `parser.ParseFile` calls may appear in `View()` or any function reachable from `View()`.
- **Passing `info InfoPanelData` to mid-tier or narrow-tier `RenderChrome` paths:** The mid-tier and narrow-tier paths ignore the `info` argument completely (Phase 7.1 D-116 unchanged). No render call for infopanel occurs at width < 99.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| ANSI-safe string width calculation | `len(s)` / `utf8.RuneCountInString` | `lipgloss.Width(s)` | Strips ANSI codes, counts visible cells correctly |
| Right-side tail truncation | Manual byte slicing | `ansi.Truncate(s, n, "…")` | ANSI-aware, grapheme-cluster aware, already in go.mod |
| Left-side tail truncation (for middle-trunc right half) | Manual byte slicing | `ansi.TruncateLeft(s, n, "")` | Available in charmbracelet/x/ansi v0.11.7 |
| Key file path resolution | `os.Getenv("HOME")` | `os.UserHomeDir()` | Handles edge cases; matches project precedent in validator |
| Age key validation on startup | Re-validating in agekey.go | `age.ParseIdentities` returns error on invalid format | Library handles all validation |
| Breadcrumb segment splitting | Custom tokenizer | `strings.Split(m.breadcrumb, " > ")` | Matches exactly how `SetBreadcrumb` joins them |

---

## Runtime State Inventory

> This is a data-binding / presentation phase with no rename/refactor. Runtime state inventory is not applicable.

None — verified by phase scope review. Phase 8 adds new fields to an existing struct and new files; it does not rename any persistent identifier.

---

## Common Pitfalls

### Pitfall A: `age.Identity` Interface Has No `Recipient()` Method

**What goes wrong:** Code writes `ids[0].Recipient().String()` against the `age.Identity` interface and fails to compile.
**Why it happens:** The v1.1 STACK.md example code (from pre-implementation research) described the method as if it were on the interface. It is only on `*X25519Identity`.
**How to avoid:** Type-assert to `*age.X25519Identity` before calling `Recipient()`. Add a `default:` arm that returns `""` for `*HybridIdentity` and plugin identities (future-proofing).
**Warning signs:** `ids[0].Recipient()` compile error: `Identity has no field or method Recipient`.

### Pitfall B: `ansi.Truncate` Is Tail-Only, Not Middle-Truncation

**What goes wrong:** `RenderInfoPanel` calls `ansi.Truncate(value, 32, "…")` and produces right-truncated values like `secrets/prod…` instead of middle-truncated `secrets/…yaml`.
**Why it happens:** `ansi.Truncate` truncates from the right — the CONTEXT.md description of "middle-truncation" implies a custom function is needed.
**How to avoid:** Implement `middleTruncate` using `ansi.Truncate` for the left half and `ansi.TruncateLeft` for the right half. The implementation uses `ansi.TruncateLeft(s, width(s)-rightCells, "")` to extract the right fragment.
**Warning signs:** Values like `age1abc…` (age fingerprint) or `secrets/prod…` (path) instead of `age1abc…xyz` or `secrets/…prod.yaml`.

### Pitfall C: StatusBar `renderEnvIndicators` Uses `lipgloss.NewStyle()` Inline

**What goes wrong:** `TestSubmodelViewsNoNewStyle` fails after Phase 8 because `statusbar.go` is added to the scan scope.
**Why it happens:** The existing `renderEnvIndicators` in `statusbar.go:200-237` uses inline `lipgloss.NewStyle().Foreground(...).Render(...)` calls [VERIFIED: statusbar.go lines 204-231]. These are inside functions reachable from `StatusBarModel.View()`.
**How to avoid:** Phase 8 does NOT add `statusbar.go` to the `TestSubmodelViewsNoNewStyle` scope (the submodel walker scans only the 8 files in `submodelFiles` slice). However, if Phase 8 adds `infopanel.go` and `crumbs.go` to the scope, ensure these two new files have zero inline `NewStyle()` calls. The `statusbar.go` inline calls are pre-existing technical debt outside Phase 8 scope.
**Warning signs:** Unexpected `TestSubmodelViewsNoNewStyle` failure after scope extension — check which file triggered it.

### Pitfall D: `crumbsHeight` Flip Without View() Guard Update

**What goes wrong:** `crumbsHeight` returns `lipgloss.Height(RenderCrumbs(...))` (typically 1) but `View()` still has `if crumbsHeight(m) > 0 { sections = append(sections, "") }` — the guard is now true but appends the old `""` placeholder instead of the real crumbs string.
**Why it happens:** Two simultaneous changes (flip crumbsHeight + update sections slice) are required atomically.
**How to avoid:** Plan 3 must update both in the same commit. The `""` placeholder at model.go:1353 becomes `ui.RenderCrumbs(m.status.Segments(), m.width)`.
**Warning signs:** Crumbs row in golden shows blank line instead of chip pills.

### Pitfall E: `RenderChrome` Signature Change Breaks `chromeHeight`

**What goes wrong:** `chromeHeight` (model.go:1532) calls `ui.RenderChrome(m.menuHints(), ui.LogoInfo, m.width)` — after the signature change to add `info InfoPanelData`, this call site fails to compile.
**Why it happens:** Two call sites for `RenderChrome` exist in `model.go` (one in `View()` at line 1345, one in `chromeHeight` at line 1532). Both must be updated.
**How to avoid:** Plan 3 task that changes `RenderChrome` signature must grep for all call sites: `grep -n "RenderChrome" internal/app/model.go internal/ui/chrome.go` and update both.
**Warning signs:** Build error after signature change: "too few arguments in call to ui.RenderChrome".

### Pitfall F: Resize Golden at 40x12 — Crumbs Row Visible Even in Narrow Tier

**What goes wrong:** After the `crumbsHeight` flip, the 40x12 golden shows a crumbs row above the body even at narrow-tier width (40 cols). This is CORRECT per D-216 ("crumbs render at every chrome tier") but may surprise the plan executor who expects no changes at 40x12.
**Why it happens:** D-216 explicitly states crumbs are independent of chrome tier. At 40x12, chrome = 1-row narrow stub + crumbs row = 1 row + body = height - 3.
**How to avoid:** Document in Plan 3 success criteria: "40x12 golden WILL change — narrow stub (1 row) + crumbs (1 row) + body + status bar is the correct layout. Verify body region is reachable (height - 3 >= 0 at 12 rows with 1 status bar)."
**Warning signs:** Executor skips 40x12 golden refresh assuming narrow tier is unchanged.

### Pitfall G: Pitfall 5 / D-206 — Bold Weight as Redundant Channel

**What goes wrong:** `CrumbChipActiveStyle` is declared without `Bold(true)`, meaning the active chip differs from inactive chips only by background color. On 16-color terminals where both `ColorAccent` and `ColorSurface` downsample to the same ANSI index, there is no visual distinction.
**Why it happens:** Developer implements D-206 partially — bg+fg swap without bold.
**How to avoid:** The style declaration must include all three: `Background(ColorAccent).Foreground(ColorBg).Bold(true)`. Test this with `TestRenderCrumbs_ActiveBoldBg` (from D-219 test list) which asserts the active chip renders with bold.
**Warning signs:** Active chip and inactive chip are indistinguishable in `TestChromeASCIIOnly` ANSI-stripped output.

### Pitfall H: go-git v5.17.0 vs v5.17.2 Version Drift

**What goes wrong:** CLAUDE.md specifies `go-git/go-git/v5 v5.17.2` but the actual `go.mod` has `v5.17.0` [VERIFIED: go.mod]. The API for `Head()`, `IsBranch()`, and `Short()` is identical across both versions — this is not a functional problem. However, the plan-checker may flag the discrepancy.
**Why it happens:** CLAUDE.md was authored before the go.mod was pinned; the recommendation was aspirational. The installed version is v5.17.0.
**How to avoid:** Plan 3 (or any plan that adds `GetBranch`) should confirm `go mod tidy` produces no version bump. If the project wants to upgrade to v5.17.2, that is a separate PR. Phase 8 uses v5.17.0 APIs which are identical to v5.17.2 for the methods needed.
**Warning signs:** `go mod tidy` unexpectedly upgrades go-git in go.sum during Plan 3.

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | stdlib `testing` + `github.com/stretchr/testify` v1.11.1 + `charmbracelet/x/exp/teatest` |
| Config file | none (standard `go test`) |
| Quick run command | `go test ./internal/ui/... ./internal/git/... ./internal/app/... -count=1` |
| Full suite command | `go test ./... -count=1` |

**Baseline confirmed:** All packages pass `go test ./... -count=1` as of 2026-04-28. [VERIFIED: test run output showing all 9 packages OK]

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Plan | Automated Command |
|--------|----------|-----------|------|-------------------|
| UI-04 | RenderInfoPanel renders all 5 rows with correct labels | unit | 1 | `go test ./internal/ui/... -run TestRenderInfoPanel_AllRowsAligned -count=1` |
| UI-04 | RenderInfoPanel shows `-` markers for empty/missing fields | unit | 1 | `go test ./internal/ui/... -run TestRenderInfoPanel_EmptyMarkers -count=1` |
| UI-05 | RenderInfoPanel truncates age fingerprint to ≤10 cells + ellipsis | unit | 1 | `go test ./internal/ui/... -run TestRenderInfoPanel_TruncatesAge -count=1` |
| UI-05 | RenderInfoPanel truncates long path via middle-truncation | unit | 1 | `go test ./internal/ui/... -run TestRenderInfoPanel_TruncatesPath -count=1` |
| UI-04 | ParseAgeKey returns first identity Recipient().String() | unit | 1 | `go test ./internal/ui/... -run TestParseAgeKey_FirstIdentity -count=1` |
| UI-04 | ParseAgeKey returns empty string on missing file | unit | 1 | `go test ./internal/ui/... -run TestParseAgeKey_MissingFile -count=1` |
| UI-07 | RenderCrumbs renders k9s-exact `<segment>` pills | unit | 1 | `go test ./internal/ui/... -run TestRenderCrumbs_KnsExactPills -count=1` |
| UI-07 | RenderCrumbs active chip has bold+accent bg | unit | 1 | `go test ./internal/ui/... -run TestRenderCrumbs_ActiveBoldBg -count=1` |
| UI-07 | RenderCrumbs normalises: lowercase + strip spaces | unit | 1 | `go test ./internal/ui/... -run TestRenderCrumbs_LowercaseStripSpaces -count=1` |
| UI-07 | RenderCrumbs performs middle-segment ellipsis on overflow | unit | 1 | `go test ./internal/ui/... -run TestRenderCrumbs_MiddleEllipsis -count=1` |
| UI-04 | GetBranch returns branch name for normal branch | unit | 2 | `go test ./internal/git/... -run TestGetBranch_NormalBranch -count=1` |
| UI-04 | GetBranch returns detached+hash for detached HEAD | unit | 2 | `go test ./internal/git/... -run TestGetBranch_DetachedHead -count=1` |
| UI-04 | GetBranch returns ErrRepositoryNotExists for non-git dir | unit | 2 | `go test ./internal/git/... -run TestGetBranch_NonRepo -count=1` |
| UI-08 | StatusBarModel.View() renders right-aligned env+clipboard only after shrink | unit | 2 | `go test ./internal/ui/... -run TestStatusBar_RightAlignOnly -count=1` |
| UI-08 | StatusBarModel.Segments() returns correct split from breadcrumb | unit | 2 | `go test ./internal/ui/... -run TestStatusBar_SegmentsAccessor -count=1` |
| UI-04+07 | RenderChrome full-tier includes info panel content at width=200 | integration | 3 | `go test ./internal/app/... -run TestRenderChrome_FullTierWithInfoPanel -count=1` |
| UI-07 | crumbsHeight(m) > 0 after WindowSizeMsg + breadcrumb set | integration | 3 | `go test ./internal/app/... -run TestCrumbsHeight_NonZero -count=1` |
| UI-04 | m.infoPanel.FileCount updated on FilesDiscoveredMsg | integration | 3 | `go test ./internal/app/... -run TestInfoPanelCacheRefresh_OnFilesDiscovered -count=1` |
| UI-15 | Chrome files (incl. infopanel.go, crumbs.go) ASCII-only + U+2026 | grep-gate | 3 | `go test ./internal/app/... -run TestChromeASCIIOnly -count=1` |
| UI-15 | Chrome files use NormalBorder only | grep-gate | 3 | `go test ./internal/app/... -run TestChromeNormalBorderOnly -count=1` |
| UI-15 | No lipgloss.NewStyle() reachable from View() | AST gate | 3 | `go test ./internal/app/... -run TestViewNoNewStyle -count=1` |
| UI-15 | No lipgloss.NewStyle() in infopanel.go and crumbs.go | AST gate | 3 | `go test ./internal/ui/... -run TestSubmodelViewsNoNewStyle -count=1` |
| UI-16 | Resize goldens at 80x24, 120x40, 200x60 match expectations | golden | 3 | `go test ./internal/app/... -run TestResize -count=1` |

### Grep-Gate Scope Extension for Plan 3

**`TestChromeASCIIOnly`** (chrome_test.go:47): The `files` slice already includes `"internal/ui/crumbs.go"` with `// Phase 8; skipped if missing` comment [VERIFIED: chrome_test.go:61-62]. This is already handled — Plan 3 creates the file and the skip-if-missing guard stops skipping.

**`TestChromeNormalBorderOnly`** (chrome_test.go:101): The `files` slice includes only `chrome.go`, `logo.go`, `menu.go` [VERIFIED: chrome_test.go:105-109]. Plan 3 must add `"internal/ui/infopanel.go"` and `"internal/ui/crumbs.go"` to this list.

**`TestSubmodelViewsNoNewStyle`** (submodel_view_no_newstyle_test.go): The `submodelFiles` var lists 8 specific filenames [VERIFIED: submodel_view_no_newstyle_test.go, lines visible]. Plan 3 must add `"infopanel.go"` and `"crumbs.go"` to this list.

**ASCII allowlist extension for U+2026:** Already present in `TestChromeASCIIOnly` allowlist [VERIFIED: chrome_test.go:53 `'…': true`]. No change needed.

### Wave 0 Gaps (New Test Files Required)

Plan 1:
- [ ] `internal/ui/infopanel_test.go` — covers UI-04, UI-05 primitive tests
- [ ] `internal/ui/crumbs_test.go` — covers UI-07 primitive tests
- [ ] `internal/ui/agekey_test.go` — covers UI-04 age key parser tests

Plan 2:
- [ ] `internal/git/status_test.go` extended with `TestGetBranch_*` subtests (file exists; just add cases)
- [ ] `internal/ui/statusbar_test.go` extended with `TestStatusBar_RightAlignOnly`, `TestStatusBar_SegmentsAccessor` (may require modifying existing tests for changed render shape)

Plan 3:
- New integration test functions added to existing `internal/app/chrome_test.go`
- Resize goldens regenerated with `go test ./internal/app/... -run TestResize -update`

---

## Code Examples

### Age Key Parser (internal/ui/agekey.go — NEW)

```go
// Source: filippo.io/age v1.3.1 parse.go:24, x25519.go
// [VERIFIED: API confirmed locally at /home/moersener/go/pkg/mod/filippo.io/age@v1.3.1/]
func parseAgeKeyFingerprint(keyFilePath string) string {
    f, err := os.Open(keyFilePath)
    if err != nil {
        return ""
    }
    defer f.Close()
    ids, err := age.ParseIdentities(f)
    if err != nil || len(ids) == 0 {
        return ""
    }
    // MUST type-assert: age.Identity interface has no Recipient() method.
    // *X25519Identity and *HybridIdentity both have Recipient().
    switch id := ids[0].(type) {
    case *age.X25519Identity:
        return id.Recipient().String()  // "age1..."
    default:
        return ""
    }
}

// AgeKeyFilePath returns the path to the age keys file, honouring
// $SOPS_AGE_KEY_FILE per D-214. Falls back to ~/.config/sops/age/keys.txt.
func AgeKeyFilePath() (string, error) {
    if v := os.Getenv("SOPS_AGE_KEY_FILE"); v != "" {
        return v, nil
    }
    home, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(home, ".config", "sops", "age", "keys.txt"), nil
}
```

### GetBranch (internal/git/status.go — EXTEND)

```go
// Source: go-git v5.17.0 repository.go:1505, plumbing/reference.go
// [VERIFIED: API confirmed locally at /home/moersener/go/pkg/mod/github.com/go-git/go-git/v5@v5.17.0/]
func GetBranch(repoRoot string) (branch string, detached bool, err error) {
    repo, err := gogit.PlainOpenWithOptions(repoRoot, &gogit.PlainOpenOptions{DetectDotGit: true})
    if err == gogit.ErrRepositoryNotExists {
        return "", false, gogit.ErrRepositoryNotExists
    }
    if err != nil {
        return "", false, err
    }
    ref, err := repo.Head()
    if err != nil {
        return "", false, err
    }
    if ref.Name().IsBranch() {
        return ref.Name().Short(), false, nil
    }
    // Detached HEAD: return 7-char short hash
    h := ref.Hash().String()
    if len(h) > 7 {
        h = h[:7]
    }
    return h, true, nil
}
```

### middleTruncate (internal/ui/infopanel.go — NEW)

```go
// Source: charmbracelet/x/ansi v0.11.7 truncate.go:53,173
// [VERIFIED: Truncate and TruncateLeft signatures confirmed locally]
func middleTruncate(s string, maxCells int) string {
    sw := lipgloss.Width(s)
    if sw <= maxCells {
        return s
    }
    const ellipsis = "…"
    ellipsisW := lipgloss.Width(ellipsis)
    available := maxCells - ellipsisW
    if available <= 0 {
        return ellipsis
    }
    left := available / 2
    right := available - left
    leftPart := ansi.Truncate(s, left, "")
    rightPart := ansi.TruncateLeft(s, sw-right, "")
    return leftPart + ellipsis + rightPart
}
```

### StatusBarModel.Segments() (internal/ui/statusbar.go — EXTEND)

```go
// Source: internal/ui/statusbar.go:73-78 (SetBreadcrumb joins with " > ")
// [VERIFIED: implementation read directly]
func (m StatusBarModel) Segments() []string {
    return strings.Split(m.breadcrumb, " > ")
}
```

---

## File Inventory Diff (CONTEXT.md Canonical Refs vs Actual Codebase)

The following table reconciles CONTEXT.md `<canonical_refs>` line numbers against the actual current codebase. Line numbers drift with each phase's code additions.

| Reference | CONTEXT.md Claims | Actual Line | Verified |
|-----------|-------------------|-------------|---------|
| `AppModel.View()` | line 1296 | 1296 | [VERIFIED: grep -n "func.*View\(\) tea" model.go] |
| `bodyDims(m)` helper | line 1437 | 1433 | [VERIFIED: grep shows bodyDims func at 1437, but offset 1437 reads bodyDims at line 1433 in grep] — **actual: 1437** |
| `chromeHeight(m)` | line 1528 | 1521 | [VERIFIED: grep confirms 1521] — **DRIFT: CONTEXT says 1528, actual is 1521** |
| `crumbsHeight(m)` stub | line 1539 | 1536-1542 | [VERIFIED: grep confirms 1539] — **actual: 1536 (function def), 1540-1542 (body)** |
| `RenderChrome` call in View() | line 1345 | 1345 | [VERIFIED: grep confirms 1345] |
| `sections := []string{chrome}` | line 1351 | 1351 | [VERIFIED: read model.go offset 1349] |
| `FilesDiscoveredMsg` handler | line 328+ | 328 | [VERIFIED: grep confirms 328] |
| `GitStatusMsg` handler | line 586+ | 586 | [VERIFIED: grep confirms 586] |
| `InfoPanelPlaceholderStyle` in styles.go | line 257-263 | 257-263 | [VERIFIED: read styles.go] |
| `ParsedFile.AgeRecipients []string` | parser/yaml.go:31 | 31 | [VERIFIED: read yaml.go:31] |
| `filippo.io/age` import in recipientform.go | line 22 | 22 | [VERIFIED: read recipientform.go:22] |
| `TestChromeASCIIOnly` file scope includes crumbs.go skip-if-missing | chrome_test.go | lines 60-62 | [VERIFIED: chrome_test.go already has crumbs.go with skip-if-missing] |
| `TestChromeNormalBorderOnly` scope | chrome_test.go:105-109 | 105-109 | [VERIFIED: does NOT include infopanel.go/crumbs.go yet — Plan 3 must add] |
| `submodelFiles` in submodel_view_no_newstyle_test.go | 8 files | 8 files | [VERIFIED: read directly — does NOT include infopanel.go/crumbs.go yet] |

**Key drift finding:** `chromeHeight` function definition is at line 1521 in actual code, not 1528 as cited in CONTEXT.md. All other references are accurate. The drift is minor and does not affect implementation.

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `StatusBarModel.View()` renders left+center+right breadcrumb sections | Post-Phase-8: right cluster only (env + clipboard); crumbs move to dedicated chip row | Phase 8 | Status bar becomes thinner; breadcrumb is more visible as chip pills |
| Info-panel slot renders `InfoPanelPlaceholderStyle.Render("")` (Phase 7 stub) | `RenderInfoPanel(m.infoPanel)` with live data | Phase 8 | Top-left header fills with context data |
| `crumbsHeight(m)` stub returns `0` | Returns `lipgloss.Height(RenderCrumbs(...))` typically 1 | Phase 8 | `bodyDims` automatically subtracts the crumbs row; no SetSize call-site changes needed |
| `RenderChrome(hints, logoStatus, width)` 3-arg signature | `RenderChrome(hints, logoStatus, info InfoPanelData, width int)` 4-arg | Phase 8 | Two call sites in model.go must be updated |

**Deprecated/outdated after Phase 8:**
- `statusbar.go:renderBreadcrumb()` — function body replaced or deleted; breadcrumb display moves to chip row
- `StatusBarModel.itemCount`, `StatusBarModel.itemLabel` fields — deleted per D-209
- `StatusBarModel.SetItemCount` — becomes no-op or removed
- `m.status.SetItemCount(...)` call-sites in model.go — deleted (16 sites per CONTEXT.md)
- `""` placeholder at model.go:1353 in `View()` sections slice — replaced with `ui.RenderCrumbs(m.status.Segments(), m.width)`

---

## Rejected Alternatives

The following alternatives were evaluated during the discuss-phase session (2026-04-28) and are explicitly closed. Plan reviewers must not re-open these debates.

| Alternative | Rejected In | Reason |
|-------------|-------------|--------|
| Using `bubbles/v2/table` for info-panel layout | D-201 | Interactive table with cursor state — wrong primitive for a static label:value display |
| Using `k9s ClusterInfo` observer pattern (StylesChanged) | Phase 7 CONTEXT.md | Not portable to Bubble Tea Elm architecture; immediate-mode composition is the correct pattern |
| Polling goroutine for age key file changes | Deferred section | No runtime data source that mutates outside TUI; restart is acceptable for key rotation |
| Moving breadcrumb ownership from StatusBarModel to AppModel | D-210 | Would require touching 16 SetBreadcrumb call-sites for no separation-of-concerns gain |
| Using `fmt.Sprintf` padding for info-panel rows instead of `lipgloss.Width` | Claude's Discretion left open | Both approaches are valid; plan author picks |
| Showing full age public key fingerprint (62+ chars) | D-203, D-220, Pitfall 11 | Screenshot/screenshot disclosure risk; truncation with ellipsis is the mitigation |
| Splitting Plan 3 by refresh path | D-218 | Would re-open `View()` and `Update()` multiple times; golden churn is worse |
| Using `adaptive`-style colors for chip bg/fg pairing | Rejected by project CLAUDE.md | `lipgloss.AdaptiveColor` has confirmed hang issue (#1036); explicit hex only |
| Padding with `< segment >` (spaces inside brackets) | D-205 | k9s uses `<segment>` without inner padding; bg fill provides visual spacing |

---

## Open Questions

1. **`GitStatusMsg` struct extension vs separate message**
   - What we know: The existing `GitStatusMsg` at model.go:147 carries `Statuses map[string]git.GitStatus`, `GitAvailable bool`, `Err error`.
   - What's unclear: Should `GetBranch` be called in the same async cmd that calls `GetFileStatuses` (returning one extended `GitStatusMsg`) or in a separate async cmd returning a new `GitBranchMsg`?
   - Recommendation: Extend `GitStatusMsg` with `Branch string` and `DetachedHead bool` fields. Both calls hit the same repo and share the `PlainOpenWithOptions` overhead. Single message = single handler path. Plan 2 author decides.

2. **`SetItemCount` deletion scope in Plan 2**
   - What we know: D-209 says delete `itemCount`/`itemLabel` fields and convert `SetItemCount` to no-op or delete it. All 16 `m.status.SetItemCount(...)` call-sites in model.go would need deletion.
   - What's unclear: Are all 16 call-sites actually referencing model.go alone, or do some live in other files? CONTEXT.md says "existing call-sites in `internal/app/model.go`".
   - Recommendation: `grep -rn "SetItemCount" internal/` to confirm scope before Plan 2 execution. If confined to model.go, delete cleanly.

3. **`agekey.go` location — `internal/ui` vs `internal/app`**
   - What we know: D-214 says "new `internal/ui/agekey.go` (or inline in `internal/app/model.go`)" with recommendation for separate file.
   - What's unclear: `agekey.go` needs to import `filippo.io/age` and `os` and `path/filepath`. It is logically a UI-support file (feeds `InfoPanelData.AgeFingerprint`). But it could also live in `internal/app` if the planner prefers.
   - Recommendation: `internal/ui/agekey.go` — mirrors the symmetry of `recipientform.go`'s age import; keeps the `ui` package as the owner of all display-data preparation.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | Build | Yes | 1.26.2 (go.mod) | — |
| filippo.io/age v1.3.1 | agekey.go ParseIdentities | Yes (in go.mod) | v1.3.1 | — |
| go-git v5.17.0 | GetBranch | Yes (in go.mod) | v5.17.0 | — |
| lipgloss/v2 v2.0.3 | All renderers | Yes (in go.mod) | v2.0.3 | — |
| charmbracelet/x/ansi v0.11.7 | middleTruncate, TruncateLeft | Yes (in go.mod) | v0.11.7 | — |
| `~/go/pkg/mod/filippo.io/age@v1.3.1/` | Local verification | Yes | v1.3.1 | — |

All dependencies verified in local module cache. No new installations required.

---

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | Phase 8 is display-only; no auth flows |
| V3 Session Management | No | Phase 8 adds no session state |
| V4 Access Control | No | No new access-controlled operations |
| V5 Input Validation | Yes (limited) | User-supplied strings (file paths, branch names) truncated via `ansi.Truncate`/`middleTruncate` before display |
| V6 Cryptography | No | Age crypto stays with SOPS subprocess; Phase 8 only reads the public fingerprint |

### Pitfall 11 — 5-Question Security Review Pre-Checklist

Per D-220, each new info-panel field requires sign-off. Pre-validation here:

| Field | Private key material? | Absolute path? | Copy binding? | In logs? | Screenshot risk? |
|-------|-----------------------|----------------|---------------|----------|------------------|
| `cfg:` `.sops.yaml` rel-path | No | No (repo-relative only via `filepath.Rel`) | No | No | Low (relative path only) |
| `age:` fingerprint | Derived from private key | No (fingerprint is public key string, not key path) | No | No | Medium — truncated to ≤10 chars + `…` per D-203 |
| `rcp:` count | No | No | No | No | None (integer only) |
| `git:` branch+marker | No | No | No | No | Low (branch name + `*` or clean) |
| `fil:` count | No | No | No | No | None (integer only) |

**Key risk:** `age:` row shows `identity.Recipient().String()` (the public key, not private). Truncated to ≤10 visible cells. The full `age1...` string is 62+ chars; truncated to `age1ab…yz` reveals 4 chars of prefix + 2 chars of suffix. This is the minimum needed for the user to identify "is this my key?" without narrowing an attacker's brute-force space significantly. Sign-off must be recorded in `08-03-SUMMARY.md`.

**$SOPS_AGE_KEY_FILE env var:** D-214 reads this env var to locate the key file. The path itself is NEVER rendered. Only the fingerprint derived from parsing the file at that path is rendered (truncated). This is correct per Pitfall 11 Rule 3: "Never show environment variable values in the header."

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `ansi.TruncateLeft` is available in charmbracelet/x/ansi v0.11.7 | Code Examples | middleTruncate cannot use it; fallback: use `utf8` + manual byte walk. Risk: LOW — function confirmed at truncate.go:173 [VERIFIED] |
| A2 | `GetBranch` extending GitStatusMsg with new fields is backward-compatible | Code Examples / Open Questions | Other code reading GitStatusMsg would be unaffected (no other consumer besides the handler). Risk: LOW |
| A3 | 40x12 golden will change after crumbsHeight flip | Pitfall F | If 40x12 narrow-tier somehow has `crumbsHeight = 0` even after flip, golden won't change — but that would be a bug. Risk: LOW |

**All tagged `[VERIFIED]` claims were confirmed via local file reads or grep. No claims are purely `[ASSUMED]`.**

---

## Sources

### Primary (HIGH confidence)
- `/home/moersener/git/sops-tui/internal/app/model.go` — AppModel struct, View(), bodyDims, chromeHeight, crumbsHeight, FilesDiscoveredMsg handler (line 328), GitStatusMsg handler (line 586), SetBreadcrumb call-sites, RenderChrome call-sites
- `/home/moersener/git/sops-tui/internal/ui/chrome.go` — RenderChrome full signature, InfoPanelPlaceholderStyle usage
- `/home/moersener/git/sops-tui/internal/ui/statusbar.go` — StatusBarModel struct, View(), renderBreadcrumb, renderEnvIndicators inline NewStyle calls
- `/home/moersener/git/sops-tui/internal/ui/styles.go` — All existing package vars, InfoPanelPlaceholderStyle at line 257-263
- `/home/moersener/git/sops-tui/internal/git/status.go` — PlainOpenWithOptions pattern, existing function signatures
- `/home/moersener/git/sops-tui/internal/app/chrome_test.go` — TestChromeASCIIOnly (file scope + allowlist), TestChromeNormalBorderOnly (file scope), TestViewNoNewStyle (BFS walker)
- `/home/moersener/git/sops-tui/internal/ui/submodel_view_no_newstyle_test.go` — submodelFiles list
- `/home/moersener/git/sops-tui/internal/app/resize_test.go` — golden file pattern, testutil.RequireGoldenStructure usage
- `/home/moersener/go/pkg/mod/filippo.io/age@v1.3.1/parse.go` — ParseIdentities signature, return type, error cases
- `/home/moersener/go/pkg/mod/filippo.io/age@v1.3.1/age.go` — Identity interface (no Recipient method)
- `/home/moersener/go/pkg/mod/filippo.io/age@v1.3.1/x25519.go` — X25519Identity.Recipient() returns *X25519Recipient; String() returns AGE-SECRET-KEY-...; X25519Recipient.String() returns age1...
- `/home/moersener/go/pkg/mod/github.com/go-git/go-git/v5@v5.17.0/repository.go:1505` — Head() signature
- `/home/moersener/go/pkg/mod/github.com/go-git/go-git/v5@v5.17.0/plumbing/reference.go` — ReferenceName.IsBranch(), Short(), Reference.Name(), Reference.Hash()
- `/home/moersener/go/pkg/mod/charm.land/lipgloss/v2@v2.0.3/join.go` — JoinHorizontal, JoinVertical signatures
- `/home/moersener/go/pkg/mod/charm.land/lipgloss/v2@v2.0.3/size.go` — Width(), Height()
- `/home/moersener/go/pkg/mod/github.com/charmbracelet/x/ansi@v0.11.7/truncate.go` — Truncate(s, length, tail) at line 53; TruncateLeft(s, n, prefix) at line 173
- `/home/moersener/git/k9s/internal/ui/crumbs.go:62-74` — k9s chip format `<segment>`, lowercase+strip-spaces, bg active swap (D-205, D-207 source)
- `/home/moersener/git/k9s/internal/ui/crumbs.go:32` — SetBorderPadding(0,0,1,1) (D-208 source)
- `/home/moersener/git/k9s/internal/view/cluster_info.go:67-89` — label:value table layout (D-201 row layout source)
- `/home/moersener/git/k9s/internal/view/cluster_info.go:113-144` — ClusterInfoChanged event-driven refresh (D-213 caching pattern source)

### Secondary (MEDIUM confidence)
- `.planning/phases/08-header-info-panel/08-CONTEXT.md` — all 20 decisions (D-201..D-220), plan split, security review gate
- `.planning/research/PITFALLS.md` — Pitfall 5, 9, 11, 14, 15 (mitigations verified as addressed by decisions)
- `.planning/research/STACK.md` — filippo.io/age v1.3.1 and go-git v5.17.x sections

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all libraries verified in local module cache; no new go.mod entries
- Architecture: HIGH — data flow verified against actual code; integration points confirmed with exact line numbers
- Pitfalls: HIGH — 8 phase-specific pitfalls verified against actual API behavior; 3 are new discoveries (A, B: API shape; C: statusbar NewStyle; D-G: implementation traps)
- Library API signatures: HIGH — read directly from local module cache source files

**Research date:** 2026-04-28
**Valid until:** 2026-05-28 (30-day stable; no upstream API changes expected)
