# Phase 8: Header Info Panel + Crumb Chips - Pattern Map

**Mapped:** 2026-04-28
**Files analyzed:** 15 (5 new, 7 modified, 4 golden refresh)
**Analogs found:** 15 / 15 (every new/modified file has a strong in-repo analog)

---

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/ui/infopanel.go` | renderer (pure component) | request-response (data → string) | `internal/ui/logo.go` (pure renderer + package-var styles) + `internal/ui/metadata.go:80-103` (label/value JoinHorizontal pattern) | exact |
| `internal/ui/crumbs.go` | renderer (pure component) | request-response (data → string) | `~/git/k9s/internal/ui/crumbs.go:62-74` (k9s parity source — verbatim format) + `internal/ui/menu.go:109-112` (lipgloss.JoinHorizontal pattern) | exact (k9s parity required by project memory) |
| `internal/ui/agekey.go` | utility (parser/IO helper) | file-I/O (read once at startup) | `internal/ui/recipientform.go:22, 99-103` (existing `filippo.io/age` import + ParseX25519Recipient pattern) + `internal/validator/startup.go:48-62` (HOME resolution + keys.txt path construction) | role-match (same library, parallel parser) |
| `internal/git/status.go` (extend) | service / git backend | file-I/O (sync go-git read) | itself — `GetFileStatuses` / `GetFileHistory` / `GetLastCommitTime` (lines 65-115 / 126-163 / 211-235) — same `PlainOpenWithOptions` + `ErrRepositoryNotExists` pattern | exact (additive function in same file) |
| `internal/ui/styles.go` (extend) | config (style declarations) | static (package-var allocation) | itself — existing additive `var (...)` block, especially `MenuKeyStyle` / `MenuDescStyle` / `InfoPanelPlaceholderStyle` (lines 218-263) | exact |
| `internal/ui/statusbar.go` (modify) | sub-model (component) | event-driven (Bubble Tea Update/View) | itself pre-modification — the change is removal-of-code (drop `renderBreadcrumb` / center / pipes) + additive `Segments()` accessor | exact (self-modify) |
| `internal/ui/chrome.go` (modify) | composer (renderer) | request-response | itself — `RenderChrome` body (lines 106-134); the parameter-and-substitute pattern (replace `InfoPanelPlaceholderStyle.Render("")` with `RenderInfoPanel(info)`) | exact (self-modify, signature extension) |
| `internal/app/model.go` (modify) | controller (Bubble Tea root model) | event-driven (msg-based cache refresh) | itself — existing `m.files`/`m.gitRepoRoot`/`m.currentParsed` field-cache + handler-refresh pattern (lines 328-363, 586-614) | exact (self-modify, additive cache field) |
| `internal/ui/infopanel_test.go` | test (unit) | request-response | `internal/ui/logo_test.go` + `internal/ui/menu_test.go` | exact |
| `internal/ui/crumbs_test.go` | test (unit) | request-response | `internal/ui/menu_test.go` (allowlist-based ASCII checking + style RGB triplet assertion) | exact |
| `internal/ui/agekey_test.go` | test (unit, file-I/O) | file-I/O | `internal/git/status_test.go` (t.TempDir + write file + assert) | role-match |
| `internal/git/status_test.go` (extend) | test (unit) | file-I/O | itself — existing `TestGetFileStatuses_*` + `TestGetLastCommitTime_*` 3-subtest pattern (lines 76-127, 169-198) | exact (additive test in same file) |
| `internal/app/chrome_test.go` (extend) | test (grep-gate + integration) | static AST scan + Bubble Tea WindowSizeMsg | itself — Phase 7/7.1 test patterns (lines 47-127, 150-271) + existing file-scope arrays | exact (self-extend) |
| `internal/ui/submodel_view_no_newstyle_test.go` (extend) | test (grep-gate) | static AST scan | itself — `submodelFiles` allowlist (line 29-38) | exact (self-extend, list append) |
| `internal/app/testdata/resize_*.golden` (regen) | test fixture | golden file regeneration | existing Phase 7.1 goldens (40×12, 80×24, 120×40, 200×60) | exact |
| `internal/ui/statusbar_test.go` (modify) | test (unit) | request-response | itself pre-modification — delete tests for removed center/pipe sections, add tests for `Segments()` accessor and right-aligned shape | exact (self-modify) |

---

## Pattern Assignments

### `internal/ui/infopanel.go` (renderer, pure function)

**Plan:** Plan 1
**Analogs:** `internal/ui/logo.go` (package shape + pure-renderer discipline) + `internal/ui/metadata.go:80-103` (label/value JoinHorizontal pattern)

**Package header pattern** (copy from `logo.go:1-19`):

```go
// Package ui - header info-panel primitive (Phase 8).
//
// RenderInfoPanel returns the 5-row info panel rendered into the 38x6
// envelope reserved by Phase 7's InfoPanelPlaceholderStyle (D-16).
// Phase 8 D-201..D-204 lock the row schema: cfg / age / rcp / git / fil.
//
// Pure function of InfoPanelData — no I/O, no AppModel coupling.
// View() reads the cached InfoPanelData on AppModel; refresh happens at
// four event seams (D-213) elsewhere.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
    "fmt"
    "strconv"

    "charm.land/lipgloss/v2"
    "github.com/charmbracelet/x/ansi"
)
```

**Data struct pattern** (copy shape from existing structs, e.g. `EnvStatus` in `statusbar.go:37-46`):

```go
// InfoPanelData is the value passed to RenderInfoPanel — pre-computed,
// no I/O at render time (Pitfall 15).
type InfoPanelData struct {
    SopsYamlRelPath string  // "" → renders "-"
    AgeFingerprint  string  // "age1...xyz" → middle-truncated to ≤10 cells
    RecipientCount  int     // -1 → renders "-"; 0+ renders decimal
    GitBranch       string  // "" → renders "-"
    GitDetached     bool    // true → "HEAD@<7-char-hash>" format (branch holds short hash)
    GitDirty        bool    // true → trailing " *"
    FileCount       int     // -1 → renders "-"; 0+ renders decimal
}
```

**Render pattern** (use `JoinVertical` of pre-rendered rows; mirror `metadata.go:95` row composition):

```go
// RenderInfoPanel returns the 5-row info panel string. Caller wraps via
// InfoPanelPlaceholderStyle (Width=38, Height=6) at the chrome composer
// to enforce the envelope; this function returns un-padded content so
// the wrapper can size to its declared 38x6 box.
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

// infoPanelRow composes a single row by joining the muted Width(5) label
// (via InfoPanelLabelStyle) with the foreground value (via InfoPanelValueStyle).
// Mirrors metadata.go:95 — label.Render(labelText) + value.Render(valueText).
func infoPanelRow(label, value string) string {
    return InfoPanelLabelStyle.Render(label) + InfoPanelValueStyle.Render(value)
}
```

**Empty-marker helpers** (D-204 — single source of truth):

```go
const emptyMarker = "-" // D-204: ASCII hyphen-minus for missing/uncomputed.

func sopsYamlDisplay(path string) string {
    if path == "" {
        return emptyMarker
    }
    return middleTruncate(path, 32) // 38 - 5 label = 33; D-201 budget 32 + 1 padding cell
}

func ageDisplay(fingerprint string) string {
    if fingerprint == "" {
        return emptyMarker
    }
    return middleTruncate(fingerprint, 10) // D-203: ≤10 cells with visible "…"
}

func rcpDisplay(n int) string {
    if n < 0 {
        return emptyMarker
    }
    return strconv.Itoa(n)
}

func gitDisplay(branch string, detached, dirty bool) string {
    if branch == "" {
        return emptyMarker
    }
    if detached {
        return "HEAD@" + branch // branch holds 7-char hash per D-215
    }
    if dirty {
        return branch + " *"
    }
    return branch
}

func filDisplay(n int) string {
    if n < 0 {
        return emptyMarker
    }
    return strconv.Itoa(n)
}
```

**middleTruncate pattern** (RESEARCH.md verified — `ansi.Truncate` is tail-only; `ansi.TruncateLeft` is head-strip):

```go
// middleTruncate keeps the start and end of s, replacing the middle with
// "…" (U+2026), so the result is at most maxCells visible cells.
// Returns s unchanged if it already fits. ANSI- and grapheme-aware via
// charmbracelet/x/ansi v0.11.7.
//
// Algorithm:
//   - if width(s) <= maxCells → return s
//   - left half  = ansi.Truncate(s, leftBudget, "")  (keep first leftBudget cells)
//   - right half = ansi.TruncateLeft(s, totalCells-rightBudget, "")  (drop first totalCells-rightBudget cells)
//   - return left + "…" + right
func middleTruncate(s string, maxCells int) string {
    if lipgloss.Width(s) <= maxCells {
        return s
    }
    const ellipsis = "…"
    ellipsisW := lipgloss.Width(ellipsis)
    if maxCells <= ellipsisW {
        return ellipsis
    }
    available := maxCells - ellipsisW
    left := available / 2
    right := available - left
    leftPart := ansi.Truncate(s, left, "")
    rightPart := ansi.TruncateLeft(s, lipgloss.Width(s)-right, "")
    return leftPart + ellipsis + rightPart
}

var _ = fmt.Sprintf // reserved for Plan 1 author byte-layout discretion
```

---

### `internal/ui/crumbs.go` (renderer, pure function)

**Plan:** Plan 1
**Analogs:**
- `~/git/k9s/internal/ui/crumbs.go:62-74` — k9s parity source (project memory: hard product goal)
- `internal/ui/menu.go:109-112` — `lipgloss.JoinHorizontal` discipline + filter-then-render pattern

**k9s reference excerpt** (verbatim — D-205, D-207, D-208):

```go
// k9s crumbs.go:62-74 — verbatim source for Phase 8 chip format.
func (c *Crumbs) refresh(crumbs []string) {
    c.Clear()
    last, bgColor := len(crumbs)-1, c.styles.Frame().Crumb.BgColor
    for i, crumb := range crumbs {
        if i == last {
            bgColor = c.styles.Frame().Crumb.ActiveColor
        }
        _, _ = fmt.Fprintf(c, "[%s:%s:b] <%s> [-:%s:-] ",
            c.styles.Frame().Crumb.FgColor,
            bgColor, strings.ReplaceAll(strings.ToLower(crumb), " ", ""),
            c.styles.Body().BgColor)
    }
}
// k9s crumbs.go:32 — SetBorderPadding(0,0,1,1) → D-208 row pad.
```

**Sops-tui port pattern**:

```go
// Package ui - breadcrumb chip-pill row primitive (Phase 8).
//
// RenderCrumbs returns the chip-pill row above the titled body: each
// segment rendered as <segment> via CrumbChipStyle (inactive) or
// CrumbChipActiveStyle (last segment). Verbatim k9s parity per D-205,
// D-207, D-208 (project memory: hard product goal).
//
// Sops-tui deviation from k9s: bold weight on active chip is the
// redundant-encoding channel (Pitfall 9) — k9s relies on bg-only swap
// which fails on 16-color terminals. Plan reviewers must reject any
// drift back to bg-only.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
    "strings"

    "charm.land/lipgloss/v2"
)

// RenderCrumbs renders the segments slice as a row of <segment> chip pills.
// width is the terminal width; the row is padded 1 cell on each side
// (D-208) so the chip budget is width - 2.
func RenderCrumbs(segments []string, width int) string {
    if len(segments) == 0 {
        // Defensive: empty row at least 1 cell tall (lipgloss.Height("") == 1).
        return CrumbRowStyle.Width(width).Render("")
    }
    normalised := normaliseSegments(segments)
    fitted := truncateSegmentsToWidth(normalised, width-2) // -2 for row pad

    chips := make([]string, 0, len(fitted))
    last := len(fitted) - 1
    for i, seg := range fitted {
        text := "<" + seg + ">"
        switch {
        case seg == "…": // overflow ellipsis chip — D-216
            chips = append(chips, CrumbChipEllipsisStyle.Render(text))
        case i == last: // active chip — D-206
            chips = append(chips, CrumbChipActiveStyle.Render(text))
        default: // inactive chip — D-206
            chips = append(chips, CrumbChipStyle.Render(text))
        }
    }
    joined := strings.Join(chips, " ") // D-208: single-space separator
    return CrumbRowStyle.Width(width).Render(joined)
}

// normaliseSegments applies k9s crumbs.go:70-71 verbatim:
// strings.ReplaceAll(strings.ToLower(seg), " ", "") per D-207.
func normaliseSegments(segs []string) []string {
    out := make([]string, len(segs))
    for i, s := range segs {
        out[i] = strings.ReplaceAll(strings.ToLower(s), " ", "")
    }
    return out
}

// truncateSegmentsToWidth iteratively drops middle segments and inserts
// a "…" sentinel when the joined chip width + separators exceeds maxCells.
// Algorithm: drop middle-most segment first, replace with sentinel "…",
// re-measure, repeat until fits or only [first, "…", last] remains.
//
// The sentinel "…" is a magic value RenderCrumbs recognises and styles
// via CrumbChipEllipsisStyle (muted fg, no bg).
func truncateSegmentsToWidth(segs []string, maxCells int) []string {
    measure := func(s []string) int {
        total := 0
        for i, seg := range s {
            total += 2 + lipgloss.Width(seg) // "<" + seg + ">"
            if i > 0 {
                total++ // single-space separator
            }
        }
        return total
    }
    if measure(segs) <= maxCells {
        return segs
    }
    work := append([]string(nil), segs...)
    for measure(work) > maxCells && len(work) > 2 {
        // Drop the middle-most segment (favour preserving first + last).
        mid := len(work) / 2
        // If we're inserting the first sentinel, replace; otherwise just drop.
        if !containsEllipsisSentinel(work) {
            work = append(work[:mid], append([]string{"…"}, work[mid+1:]...)...)
        } else {
            // Already have a sentinel; drop the next middle non-sentinel segment.
            for i := mid; i >= 0 && i < len(work); {
                if work[i] != "…" {
                    work = append(work[:i], work[i+1:]...)
                    break
                }
                if i > 0 {
                    i--
                } else {
                    break
                }
            }
        }
    }
    return work
}

func containsEllipsisSentinel(segs []string) bool {
    for _, s := range segs {
        if s == "…" {
            return true
        }
    }
    return false
}
```

---

### `internal/ui/agekey.go` (utility, file-I/O parser)

**Plan:** Plan 1
**Analogs:**
- `internal/ui/recipientform.go:22, 99-103` — existing `filippo.io/age` import precedent + parse-and-validate pattern
- `internal/validator/startup.go:48-62` — HOME resolution + keys.txt path

**Imports pattern** (copy from `recipientform.go:17-25` — same module's `filippo.io/age` import path):

```go
package ui

import (
    "os"
    "path/filepath"

    "filippo.io/age"
)
```

**Path resolution pattern** (mirror `internal/validator/startup.go:48-62`):

```go
// AgeKeyFilePath returns the path to the user's SOPS age identities file.
// Honours $SOPS_AGE_KEY_FILE per SOPS CLI behaviour; falls back to
// ~/.config/sops/age/keys.txt (existing convention used by
// internal/validator/startup.go:59).
//
// D-214 + D-220 question 4: the path itself MUST NOT appear in any
// rendered chrome surface. This function is consumed only by
// parseAgeKeyFingerprint at startup.
func AgeKeyFilePath() (string, error) {
    if env := os.Getenv("SOPS_AGE_KEY_FILE"); env != "" {
        return env, nil
    }
    home, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(home, ".config", "sops", "age", "keys.txt"), nil
}
```

**Parser pattern** (mirror `recipientform.go:99-103` — Parse → type-assert → call .Recipient().String()):

```go
// parseAgeKeyFingerprint reads keyFilePath via age.ParseIdentities and
// returns the first identity's Recipient().String() (e.g. "age1abc...xyz").
// Returns "" on any failure (file missing, parse error, no identities,
// non-X25519 identity type) — render layer shows "-" per D-204.
//
// Critical (RESEARCH.md): age.Identity is an interface and does NOT have
// a Recipient() method. The concrete *age.X25519Identity does. We MUST
// type-assert before calling Recipient().String(); calling String()
// directly on the Identity would log the AGE-SECRET-KEY private key
// (Pitfall 11 leak).
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
    if x25519, ok := ids[0].(*age.X25519Identity); ok {
        return x25519.Recipient().String() // "age1..." Bech32-encoded PUBLIC key
    }
    return "" // Plugin / hybrid identities — Plan 1 author may extend later.
}
```

---

### `internal/git/status.go` (extend with `GetBranch`)

**Plan:** Plan 2
**Analog:** itself — `GetFileStatuses` (lines 65-115) is the canonical PlainOpenWithOptions + ErrRepositoryNotExists pattern.

**Shared opening pattern** (existing in `status.go:66-73, 127, 212`):

```go
// status.go:66-70 — canonical entry point.
repo, err := gogit.PlainOpenWithOptions(repoRoot, &gogit.PlainOpenOptions{DetectDotGit: true})
if err == gogit.ErrRepositoryNotExists {
    // D-12: non-git directory is not an error condition.
    return /* zero values */, gogit.ErrRepositoryNotExists
}
if err != nil {
    return /* zero values */, err
}
```

**New function pattern** (D-215 — append to `status.go` after `GetLastCommitTime` at line ~235):

```go
// GetBranch returns the current HEAD branch name plus a flag indicating
// whether HEAD is detached.
//
// On a normal branch: branch = "main" (or whatever Short() returns from
// "refs/heads/main"); detached = false.
//
// On detached HEAD: branch = first 7 chars of the commit hash; detached = true.
//
// On a non-git directory: returns ("", false, gogit.ErrRepositoryNotExists)
// matching the GetFileStatuses contract (D-12).
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
    return ref.Hash().String()[:7], true, nil
}
```

---

### `internal/ui/styles.go` (extend with 8 new style vars)

**Plan:** Plan 1
**Analog:** itself — existing `MenuKeyStyle` / `MenuDescStyle` / `InfoPanelPlaceholderStyle` block (lines 218-263).

**Existing precedent excerpt** (`styles.go:218-263`):

```go
// MenuKeyStyle renders the mnemonic column "[key]" labels in accent color (D-05).
MenuKeyStyle = lipgloss.NewStyle().Foreground(ColorAccent)

// MenuDescStyle renders the description column text in default foreground (D-05).
MenuDescStyle = lipgloss.NewStyle().Foreground(ColorFg)

// MetadataLabelStyle is the muted label column style in MetadataModel
// content lines (Phase 7.1 D-110 lift from metadata.go:83).
MetadataLabelStyle = lipgloss.NewStyle().Foreground(ColorMuted).Width(16)

// MetadataValueStyle is the foreground value column style in
// MetadataModel content lines (Phase 7.1 D-110 lift from metadata.go:84).
MetadataValueStyle = lipgloss.NewStyle().Foreground(ColorFg)

// InfoPanelPlaceholderStyle reserves the 6-row x 38-col top-left area
// of the chrome for Phase 8's header info panel (D-16, Pitfall 1 mitigation).
InfoPanelPlaceholderStyle = lipgloss.NewStyle().
    Width(38).
    Height(6)
```

**New vars to append** (D-206 + UI-SPEC §Phase 8 new style declarations):

```go
// Phase 8: Header info panel + crumb chip styles (D-201..D-208, D-216).
// All declared as package-level vars to satisfy TestViewNoNewStyle (BFS)
// and TestSubmodelViewsNoNewStyle (file-scope, scope extended to
// infopanel.go + crumbs.go in Plan 3).

// InfoPanelLabelStyle renders the muted 5-cell label column (D-201).
// Width(5) enforces 4-char-label + 1-trailing-space alignment without
// manual padding (e.g. "cfg:" + " " rendered at Width(5) → "cfg: ").
InfoPanelLabelStyle = lipgloss.NewStyle().Foreground(ColorMuted).Width(5)

// InfoPanelValueStyle renders the foreground value column (D-201, D-204).
// Width is NOT applied — values are pre-truncated via middleTruncate
// in infopanel.go before reaching this style.
InfoPanelValueStyle = lipgloss.NewStyle().Foreground(ColorFg)

// InfoPanelSepStyle is reserved for Phase 10 visual tweak (UI-SPEC
// §Color §Phase 8 new style declarations). Phase 8 does not render an
// explicit separator cell — the trailing space inside the label
// Width(5) provides the gap. Declared as no-op so the symbol exists
// for forward compat without API churn.
InfoPanelSepStyle = lipgloss.NewStyle()

// CrumbChipStyle renders inactive crumb chip pills (D-206).
// Two-channel encoding: surface bg + fg color contrast.
CrumbChipStyle = lipgloss.NewStyle().Background(ColorSurface).Foreground(ColorFg)

// CrumbChipActiveStyle renders the active (last) crumb chip pill (D-206).
// THREE-channel encoding: accent bg + inverted fg (bg color used as fg) +
// bold weight. Bold is the colorblind-safe redundancy channel (Pitfall 9).
// k9s deviation: k9s uses bg-only swap; sops-tui adds bold deliberately
// so the active-vs-inactive distinction survives 16-color downsampling.
CrumbChipActiveStyle = lipgloss.NewStyle().
    Background(ColorAccent).
    Foreground(ColorBg).
    Bold(true)

// CrumbChipSepStyle is reserved for forward compat (Phase 10). Phase 8
// renders the inter-chip separator as a plain " " literal — no styling.
CrumbChipSepStyle = lipgloss.NewStyle()

// CrumbChipEllipsisStyle renders the middle-truncation overflow chip
// "<…>" (D-216). Muted foreground + no bg fill so the chip reads as
// "data was here, dropped due to width" — distinct from both inactive
// (bg-filled) and active (bg+bold) chips.
CrumbChipEllipsisStyle = lipgloss.NewStyle().Foreground(ColorMuted)

// CrumbRowStyle is the row container for the joined chips (D-208).
// PaddingLeft(SpaceXS) + PaddingRight(SpaceXS) mirrors k9s
// crumbs.go:32 SetBorderPadding(0,0,1,1).
CrumbRowStyle = lipgloss.NewStyle().
    PaddingLeft(SpaceXS).
    PaddingRight(SpaceXS)
```

---

### `internal/ui/statusbar.go` (modify — shrink + Segments accessor)

**Plan:** Plan 2
**Analog:** itself pre-modification (`statusbar.go:50-58, 137-174, 179-195`).

**Field deletions** (D-209 — delete `itemCount`/`itemLabel`):

```go
// BEFORE — statusbar.go:50-58
type StatusBarModel struct {
    breadcrumb   string
    itemCount    int       // ← DELETE
    itemLabel    string    // ← DELETE
    env          EnvStatus
    flash        string
    flashGen     int
    clipboardHot bool
}

// AFTER — Phase 8
type StatusBarModel struct {
    breadcrumb   string
    env          EnvStatus
    flash        string
    flashGen     int
    clipboardHot bool
}
```

**Method body to drop** (delete the `itemLabel` defaulting in `NewStatusBarModel:62-68`); `SetItemCount` becomes either deleted or a no-op (D-209 Recommendation: delete; Plan 2 author confirms no test-only references).

**New accessor pattern** (D-210 — append to statusbar.go after `SetBreadcrumb`):

```go
// Segments returns the underlying breadcrumb segments split on " > ".
// Phase 8: read-path counterpart to SetBreadcrumb; fed to
// ui.RenderCrumbs by AppModel.View() (D-210).
//
// Returns nil if the breadcrumb is empty.
func (m StatusBarModel) Segments() []string {
    if m.breadcrumb == "" {
        return nil
    }
    return strings.Split(m.breadcrumb, " > ")
}
```

**View shrink pattern** (D-211 — replace lines 137-174):

```go
// BEFORE: 3-section layout with pipes — drop this entirely.
//   left := renderBreadcrumb(m.breadcrumb)
//   center := DimText.Render(fmt.Sprintf("%d %s", m.itemCount, m.itemLabel))
//   sep := lipgloss.NewStyle().Foreground(ColorMuted).Padding(0, SpaceSM).Render("|")  // ← lipgloss.NewStyle() in View(), grep-gate hazard
//   composed := lipgloss.JoinHorizontal(lipgloss.Top, left, sep, center, sep, right)

// AFTER — D-211: right-align env+clipboard only; flash unchanged (D-212).
func (m StatusBarModel) View(width int) string {
    if m.flash != "" {
        return StatusBarStyle.
            Width(width).
            Align(lipgloss.Center).
            Render(m.flash)
    }

    right := renderEnvIndicators(m.env)
    if m.clipboardHot {
        clipIndicator := ClipboardHotStyle.Render("[clip]")
        spacer := StatusBarStyle.Render(" ") // already package-var; safe in View()
        right = lipgloss.JoinHorizontal(lipgloss.Top, clipIndicator, spacer, right)
    }

    return StatusBarStyle.
        Width(width).
        Align(lipgloss.Right).
        Render(right)
}
```

**Function deletion** (D-211 — `renderBreadcrumb` lines 179-195 is now dead code; delete entirely).

**`renderEnvIndicators` lipgloss.NewStyle() leak** — `statusbar.go:206-235` calls `lipgloss.NewStyle()` per-frame. This is reachable from `StatusBarModel.View()` and so reachable from `AppModel.View()`; it's already a violation but is not currently caught because TestViewNoNewStyle's BFS is in `internal/app/` and `TestSubmodelViewsNoNewStyle` does not list `statusbar.go`. Plan 2 author **MAY** opportunistically lift these to package vars (`SopsOkStyle`, `SopsErrorStyle`, etc.) — but per CONTEXT.md "out of scope per Phase 7.1's deferral rule". **Recommendation: lift only the styles touched by the View() shrink itself; defer the rest.** Specifically the `sep := lipgloss.NewStyle()...` on the OLD line 152-155 must be deleted (already gone since center/pipe sections are deleted).

---

### `internal/ui/chrome.go` (modify — `RenderChrome` signature change)

**Plan:** Plan 3
**Analog:** itself — `RenderChrome` body lines 106-134.

**Signature change** (UI-SPEC §interaction-contract):

```go
// BEFORE — chrome.go:106
func RenderChrome(hints []keys.MenuHint, logoStatus LogoStatus, width int) string

// AFTER — Phase 8
func RenderChrome(hints []keys.MenuHint, logoStatus LogoStatus, info InfoPanelData, width int) string
```

**Body change** (full-tier path only — replace line 126):

```go
// BEFORE — chrome.go:126
info := InfoPanelPlaceholderStyle.Render("")

// AFTER — Phase 8: render the live info panel into the 38x6 envelope.
// InfoPanelPlaceholderStyle declares Width(38) Height(6) so the wrapper
// guarantees the slot dimensions even if RenderInfoPanel produces
// fewer than 6 rows.
infoSlot := InfoPanelPlaceholderStyle.Render(RenderInfoPanel(info))
menuWidth := width - infoPanelWidth - logoWidth
if menuWidth < 1 {
    menuWidth = 1
}
menu := RenderMenu(hints, menuWidth)
logo := RenderLogo(logoStatus, logoWidth)
return lipgloss.JoinHorizontal(lipgloss.Top, infoSlot, menu, logo)
```

**Mid-tier and narrow-tier paths IGNORE `info`** — Phase 7.1 D-116 unchanged; the `info` parameter is consumed only inside the `width >= infoPanelWidth+logoWidth+menuCols*minFullMenuCol` branch.

---

### `internal/app/model.go` (modify — cache field + 4 refresh paths + crumbsHeight flip + View update)

**Plan:** Plan 3
**Analog:** itself — existing field-cache + handler-refresh patterns.

**Struct field addition** (after line 265 — extend `AppModel` struct):

```go
// AFTER (line 266 area)
type AppModel struct {
    // ... existing fields unchanged ...
    currentParsed   parser.ParsedFile

    // Phase 8: cached info-panel data. Refreshed at four event seams
    // per D-213 (NewAppModel + FilesDiscoveredMsg + GitStatusMsg +
    // recipient/edit ops). View() reads this cache only — zero I/O
    // (Pitfall 15).
    infoPanel ui.InfoPanelData
}
```

**NewAppModel population** (extend line 272-286):

```go
// AFTER — append after m.status.SetBreadcrumb("files") at line 283
func NewAppModel(env ui.EnvStatus, sopsYamlPath string) AppModel {
    m := AppModel{
        // ... existing fields unchanged ...
    }
    m.status.SetBreadcrumb("files")

    // Phase 8 D-213: populate startup-known fields.
    // Sentinel values for not-yet-computed fields (rcp / git / fil) per D-204.
    m.infoPanel = ui.InfoPanelData{
        SopsYamlRelPath: deriveSopsYamlRelPath(sopsYamlPath),
        AgeFingerprint:  loadAgeFingerprint(),
        RecipientCount:  -1, // → "-" until first FilesParsedMsg / detail nav
        GitBranch:       "", // → "-" until first GitStatusMsg
        FileCount:       -1, // → "-" until first FilesDiscoveredMsg
    }
    return m
}

// deriveSopsYamlRelPath converts the absolute .sops.yaml path to a
// repo-relative form for chrome rendering (D-220 question 2: paths must
// never expose $HOME or absolute filesystem layout). Falls back to
// basename if filepath.Rel fails (different volume / other OS edge).
func deriveSopsYamlRelPath(absPath string) string {
    if absPath == "" {
        return ""
    }
    cwd, err := os.Getwd()
    if err != nil {
        return filepath.Base(absPath)
    }
    rel, err := filepath.Rel(cwd, absPath)
    if err != nil {
        return filepath.Base(absPath)
    }
    return rel
}

// loadAgeFingerprint reads the user's age identity once at startup and
// returns the public Recipient().String() for the chrome `age:` row.
// Returns "" on any failure — render layer shows "-" per D-204.
// D-220 question 1: only the public Recipient string is surfaced; the
// private AGE-SECRET-KEY-... is never read or rendered.
func loadAgeFingerprint() string {
    path, err := ui.AgeKeyFilePath()
    if err != nil {
        return ""
    }
    return ui.ParseAgeKeyFingerprint(path) // exported wrapper from internal/ui/agekey.go
}
```

(Note: Plan 1 declares `parseAgeKeyFingerprint` lowercase per UI-SPEC API summary — but Plan 3 needs cross-package access. Plan 1 author exports it as `ParseAgeKeyFingerprint` OR Plan 3 inlines the parsing; UI-SPEC §Phase 8 public API summary chose the lowercase form, so Plan 3 must inline OR Plan 1 author renames. Recommendation: rename to `ParseAgeKeyFingerprint` — exported is consistent with the rest of the `internal/ui` API.)

**FilesDiscoveredMsg refresh** (extend handler at line 328-363):

```go
// AFTER line 348 (after m.status.SetItemCount or, if SetItemCount removed in Plan 2, after items := ...).
// D-213 — refresh FileCount + SopsYamlRelPath in case files were re-discovered.
m.infoPanel.FileCount = len(msg.Files)
m.infoPanel.SopsYamlRelPath = deriveSopsYamlRelPath(m.sopsYamlPath)
```

**GitStatusMsg refresh** (extend handler at line 586-614):

```go
// AFTER line 591 (after m.gitRepoRoot = filepath.Dir(m.sopsYamlPath)).
// D-213 + D-215 — refresh git fields. GetBranch returns ("", false, ErrRepositoryNotExists)
// for non-git directories; we treat that as the "-" empty state.
if msg.GitAvailable {
    branch, detached, err := gitpkg.GetBranch(m.gitRepoRoot)
    if err != nil {
        m.infoPanel.GitBranch = ""
        m.infoPanel.GitDetached = false
    } else {
        m.infoPanel.GitBranch = branch
        m.infoPanel.GitDetached = detached
    }
} else {
    m.infoPanel.GitBranch = ""
    m.infoPanel.GitDetached = false
}
// Aggregate dirty: any file with non-clean GitStatus.
dirty := false
for _, f := range m.files {
    if f.GitStatus != "" && f.GitStatus != string(gitpkg.GitStatusClean) {
        dirty = true
        break
    }
}
m.infoPanel.GitDirty = dirty
```

**Recipient/edit refresh** (extend after every successful re-encrypt and recipient add/remove handler):

```go
// Pattern: after every existing `m.currentParsed = newParsed` line in
// the recipient-confirm + edit-finish + bulk-rekey paths, append:
m.infoPanel.RecipientCount = len(m.currentParsed.Metadata.AgeRecipients)

// Plan 3 author: grep for `m.currentParsed = ` and audit each call-site.
```

**View() composition update** (replace lines 1345-1356):

```go
// BEFORE — model.go:1345-1356
chrome := ui.RenderChrome(hints, ui.LogoInfo, m.width)
statusBar := m.status.View(m.width)
sections := []string{chrome}
if crumbsHeight(m) > 0 {
    sections = append(sections, "") // placeholder
}
sections = append(sections, wrapped, statusBar)

// AFTER — Phase 8
chrome := ui.RenderChrome(hints, ui.LogoInfo, m.infoPanel, m.width)
crumbs := ui.RenderCrumbs(m.status.Segments(), m.width)
statusBar := m.status.View(m.width)
sections := []string{chrome, crumbs, wrapped, statusBar}
```

**chromeHeight() update** (line 1532 — pass `m.infoPanel`):

```go
// BEFORE
chrome := ui.RenderChrome(m.menuHints(), ui.LogoInfo, m.width)
// AFTER
chrome := ui.RenderChrome(m.menuHints(), ui.LogoInfo, m.infoPanel, m.width)
```

**crumbsHeight() flip** (replace lines 1539-1542):

```go
// BEFORE
func crumbsHeight(m AppModel) int {
    _ = m
    return 0
}

// AFTER — D-216 + Claude's Discretion (recommendation: lipgloss.Height for consistency with chromeHeight).
// First-frame safety: returns 0 when m.width == 0 so bodyDims doesn't
// over-subtract before the first WindowSizeMsg arrives. Mirrors
// chromeHeight at model.go:1528-1530.
func crumbsHeight(m AppModel) int {
    if m.width == 0 {
        return 0
    }
    return lipgloss.Height(ui.RenderCrumbs(m.status.Segments(), m.width))
}
```

---

### `internal/ui/infopanel_test.go` (NEW)

**Plan:** Plan 1
**Analogs:** `logo_test.go` + `menu_test.go`.

**Test scaffolding pattern** (copy from `logo_test.go:1-15`):

```go
// Tests for the Phase 8 info-panel renderer: InfoPanelData struct,
// RenderInfoPanel composition, middleTruncate helper.
package ui_test

import (
    "strings"
    "testing"

    "github.com/charmbracelet/x/ansi"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/caesarakalaeii/sops-tui/internal/ui"
)
```

**Test names** (D-219 list):

- `TestRenderInfoPanel_AllRowsAligned` — assert each row starts with the muted-color label "cfg:"/"age:"/"rcp:"/"git:"/"fil:" at column 0; assert lipgloss-stripped output has 5 lines with consistent column-2 value-start.
- `TestRenderInfoPanel_EmptyMarkers` — D-204 — five sub-tests, one per field set to its sentinel; assert "-" rendered.
- `TestRenderInfoPanel_TruncatesAge` — D-203 — pass `AgeFingerprint = "age1abcdefghijklmnopqrstuvwxyz"` and assert `lipgloss.Width(stripped age row value) <= 10` AND output contains "…".
- `TestRenderInfoPanel_TruncatesPath` — D-203 — pass long `SopsYamlRelPath`, assert middle-truncation with "…" present.
- `TestMiddleTruncate_KeepsShortStrings` — return value unchanged for inputs already ≤ maxCells.
- `TestMiddleTruncate_PreservesEnds` — for "abcdefghij" with maxCells=7, assert result starts with "abc" and ends with "ij" with "…" between.

(Use `ansi.Strip` before length / contains assertions, mirroring `logo_test.go:46-50`.)

---

### `internal/ui/crumbs_test.go` (NEW)

**Plan:** Plan 1
**Analog:** `menu_test.go` (allowlist + RGB triplet pattern).

**RGB-triplet style assertion pattern** (copy from `menu_test.go:232-240`):

```go
// menu_test.go:236-239 — ColorAccent #89b4fa → rgb(137,180,250) → "137;180;250"
out := ui.RenderMenu(hints, 80)
assert.Contains(t, out, "137;180;250", "MenuKeyStyle (ColorAccent) must apply RGB triplet")
```

**Phase 8 application** (active chip = bg ColorAccent + fg ColorBg + bold):

```go
func TestRenderCrumbs_ActiveBoldBg(t *testing.T) {
    out := ui.RenderCrumbs([]string{"sops-tui", "files", "metadata"}, 80)
    // ColorAccent #89b4fa → "137;180;250"; ColorBg #1e1e2e → "30;30;46"
    assert.Contains(t, out, "137;180;250", "active chip must apply ColorAccent bg")
    assert.Contains(t, out, "30;30;46", "active chip must invert fg to ColorBg")
    // Bold attribute = SGR code 1; lipgloss emits it as ";1m" or "[1m"
    assert.Contains(t, out, "1m", "active chip must include bold SGR (Pitfall 9 redundancy channel)")
}
```

**Test names** (D-219 list):

- `TestRenderCrumbs_KnsExactPills` — assert each segment renders as `<seg>` (k9s parity); use `ansi.Strip` then `strings.Contains`.
- `TestRenderCrumbs_ActiveBoldBg` — see above; D-206 three-channel encoding.
- `TestRenderCrumbs_LowercaseStripSpaces` — D-207 — input `["Files", "prod.yaml [M]"]` → output contains `<files>` and `<prod.yaml[m]>`.
- `TestRenderCrumbs_MiddleEllipsis` — D-216 — pass 8 segments at width=40, assert output collapses to `<first> <…> <last>` and contains "…".
- `TestRenderCrumbs_EmptySafe` — empty segments slice does not panic.
- `TestTruncateSegmentsToWidth_DropsMiddle` — pure function test for the algorithm.

---

### `internal/ui/agekey_test.go` (NEW)

**Plan:** Plan 1
**Analog:** `internal/git/status_test.go:23-57` (t.TempDir + write + assert).

**Scaffolding pattern**:

```go
package ui_test

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/caesarakalaeii/sops-tui/internal/ui"
)

func TestParseAgeKeyFingerprint_FirstIdentity(t *testing.T) {
    dir := t.TempDir()
    keyPath := filepath.Join(dir, "keys.txt")
    // Sample age private key in canonical format. Generated via age-keygen
    // for the test (one-time fixture); the expected Recipient is the
    // matching public key. Use a deterministic, well-known test pair —
    // age-keygen -o /dev/stdout produces the format expected by ParseIdentities.
    err := os.WriteFile(keyPath, []byte(`# created: 2026-01-01T00:00:00Z
# public key: age1...
AGE-SECRET-KEY-1...`), 0o600)
    require.NoError(t, err)

    fingerprint := ui.ParseAgeKeyFingerprint(keyPath)
    assert.True(t, strings.HasPrefix(fingerprint, "age1"),
        "fingerprint must be the bech32-encoded public key, got: %q", fingerprint)
}

func TestParseAgeKeyFingerprint_MissingFile(t *testing.T) {
    fingerprint := ui.ParseAgeKeyFingerprint("/nonexistent/keys.txt")
    assert.Equal(t, "", fingerprint)
}

func TestParseAgeKeyFingerprint_MalformedFile(t *testing.T) {
    dir := t.TempDir()
    keyPath := filepath.Join(dir, "keys.txt")
    require.NoError(t, os.WriteFile(keyPath, []byte("not a valid age key"), 0o600))

    fingerprint := ui.ParseAgeKeyFingerprint(keyPath)
    assert.Equal(t, "", fingerprint, "malformed key must return empty (renders as '-')")
}
```

(Plan 1 author: the well-known test key pair must be a real age-keygen output. Generate once and commit as a fixture — DO NOT use a real production key. Or use an inline stub via age.GenerateX25519Identity() at test time and write to disk.)

---

### `internal/git/status_test.go` (extend)

**Plan:** Plan 2
**Analog:** itself — existing `TestGetFileStatuses` 3-subtest pattern (lines 76-127).

**Test scaffolding pattern** (copy from `status_test.go:60-72`):

```go
// status_test.go:76-127 — 3-subtest shape:
//   t.Run("non-git directory returns empty/zero/error", ...)
//   t.Run("normal case", ...)
//   t.Run("edge case", ...)
```

**New test** (D-219 + UI-SPEC §validation-hooks):

```go
// TestGetBranch verifies branch resolution and the detached-HEAD case.
// Mirrors TestGetFileStatuses subtest shape.
func TestGetBranch(t *testing.T) {
    t.Run("non-git directory returns ErrRepositoryNotExists", func(t *testing.T) {
        dir := t.TempDir()
        branch, detached, err := git.GetBranch(dir)
        assert.ErrorIs(t, err, gogit.ErrRepositoryNotExists)
        assert.Equal(t, "", branch)
        assert.False(t, detached)
    })

    t.Run("normal branch returns branch name", func(t *testing.T) {
        dir := t.TempDir()
        repo := initRepo(t, dir)
        commitFile(t, repo, dir, "first.yaml", "data", "initial commit")

        branch, detached, err := git.GetBranch(dir)
        require.NoError(t, err)
        assert.False(t, detached)
        // PlainInit creates the default "master" branch (go-git default).
        // Accept either master or main since CI may differ.
        assert.Contains(t, []string{"master", "main"}, branch)
    })

    t.Run("detached HEAD returns 7-char hash with detached=true", func(t *testing.T) {
        dir := t.TempDir()
        repo := initRepo(t, dir)
        commitFile(t, repo, dir, "first.yaml", "data", "initial commit")

        // Detach HEAD by checking out the commit hash directly.
        head, err := repo.Head()
        require.NoError(t, err)
        wt, err := repo.Worktree()
        require.NoError(t, err)
        err = wt.Checkout(&gogit.CheckoutOptions{Hash: head.Hash()})
        require.NoError(t, err)

        branch, detached, err := git.GetBranch(dir)
        require.NoError(t, err)
        assert.True(t, detached)
        assert.Len(t, branch, 7, "detached HEAD branch must be 7-char short hash")
    })
}
```

---

### `internal/app/chrome_test.go` (extend grep-gates + add 3 integration tests)

**Plan:** Plan 3
**Analog:** itself — existing test patterns (lines 47-127, 150-271).

**File-scope extension** (line 57-62 — extend `files` slice in `TestChromeASCIIOnly`):

```go
// BEFORE
files := []string{
    "internal/ui/chrome.go",
    "internal/ui/logo.go",
    "internal/ui/menu.go",
    "internal/ui/crumbs.go", // Phase 8; skipped if missing.
}

// AFTER — Phase 8 lands both files; remove skip-if-missing comment.
files := []string{
    "internal/ui/chrome.go",
    "internal/ui/logo.go",
    "internal/ui/menu.go",
    "internal/ui/crumbs.go",
    "internal/ui/infopanel.go",
}
```

**Allowlist note** (`chrome_test.go:53` — `…` already present):

```go
// VERIFIED 2026-04-28: '…' already in allowlist (D-203 / D-216).
'…': true,
// No allowlist change needed for Phase 8.
```

**TestChromeNormalBorderOnly extension** (line 105-109):

```go
// BEFORE
files := []string{
    "internal/ui/chrome.go",
    "internal/ui/logo.go",
    "internal/ui/menu.go",
}

// AFTER — Phase 8
files := []string{
    "internal/ui/chrome.go",
    "internal/ui/logo.go",
    "internal/ui/menu.go",
    "internal/ui/crumbs.go",
    "internal/ui/infopanel.go",
}
```

**3 new integration tests** (append at end of file):

```go
// TestRenderChrome_FullTierWithInfoPanel asserts the full-tier output
// contains all 5 info-panel labels at width=200 (D-219).
func TestRenderChrome_FullTierWithInfoPanel(t *testing.T) {
    info := ui.InfoPanelData{
        SopsYamlRelPath: "secrets/.sops.yaml",
        AgeFingerprint:  "age1abc...xyz",
        RecipientCount:  4,
        GitBranch:       "main",
        GitDirty:        true,
        FileCount:       12,
    }
    hints := []keys.MenuHint{{Mnemonic: "?", Description: "help", Visible: true}}
    out := ui.RenderChrome(hints, ui.LogoInfo, info, 200)
    stripped := ansi.Strip(out)
    for _, label := range []string{"cfg:", "age:", "rcp:", "git:", "fil:"} {
        assert.Containsf(t, stripped, label, "full-tier chrome must contain %q label", label)
    }
    assert.Contains(t, stripped, "secrets/.sops.yaml")
    assert.Contains(t, stripped, "main")
}

// TestCrumbsHeight_NonZero asserts crumbsHeight(m) > 0 after a
// WindowSizeMsg + breadcrumb set (D-219).
func TestCrumbsHeight_NonZero(t *testing.T) {
    m := newAppModelForTest(t) // existing helper if present, or inline NewAppModel
    m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40}).(AppModel), nil
    m.status.SetBreadcrumb("files", "prod.yaml")
    h := crumbsHeight(m)
    assert.Greater(t, h, 0, "crumbsHeight must be >0 once width is set + breadcrumb populated")
}

// TestInfoPanelCacheRefresh_OnFilesDiscovered asserts m.infoPanel.FileCount
// reflects len(msg.Files) after the FilesDiscoveredMsg handler runs (D-219).
func TestInfoPanelCacheRefresh_OnFilesDiscovered(t *testing.T) {
    m := newAppModelForTest(t)
    files := []sops.DiscoveredFile{
        {Name: "a.yaml", AbsPath: "/tmp/a.yaml"},
        {Name: "b.yaml", AbsPath: "/tmp/b.yaml"},
    }
    updated, _ := m.Update(FilesDiscoveredMsg{Files: files})
    am := updated.(AppModel)
    assert.Equal(t, 2, am.infoPanel.FileCount, "FileCount must reflect msg.Files length")
}
```

(Plan 3 author: `newAppModelForTest` and `keys.MenuHint` constructor are existing helpers — see `internal/app/model_test.go` if present, or use `NewAppModel(ui.EnvStatus{}, "")` directly.)

---

### `internal/ui/submodel_view_no_newstyle_test.go` (extend)

**Plan:** Plan 3
**Analog:** itself — `submodelFiles` allowlist (line 29-38).

**Append to `submodelFiles` slice**:

```go
// BEFORE — line 29-38
var submodelFiles = []string{
    "filelist.go",
    "detail.go",
    "help.go",
    "diff.go",
    "metadata.go",
    "health.go",
    "history.go",
    "recipientform.go",
}

// AFTER — Phase 8 D-219
var submodelFiles = []string{
    "filelist.go",
    "detail.go",
    "help.go",
    "diff.go",
    "metadata.go",
    "health.go",
    "history.go",
    "recipientform.go",
    "infopanel.go", // Phase 8 D-219
    "crumbs.go",    // Phase 8 D-219
}
```

---

### `internal/app/testdata/resize_*.golden` (regen all 4)

**Plan:** Plan 3
**Analog:** existing Phase 7.1 goldens.

**Regen workflow** (existing pattern — Plan 3 author runs goldie's `-update` flag):

```bash
# Run from repo root:
go test ./internal/app/ -run TestResize -update
# Verify diff visually before committing:
git diff internal/app/testdata/resize_*.golden
```

**Expected golden changes per RESEARCH.md finding 5**:

| File | Change | Reason |
|------|--------|--------|
| `resize_40x12.golden` | crumbs row appears above titled body | D-216 — crumbs render at all tiers including narrow |
| `resize_80x24.golden` | crumbs row + status bar shrunk to right-aligned | mid-tier, D-211 + D-216 |
| `resize_120x40.golden` | info panel populated + crumbs row + shrunk status bar | full-tier, D-201 + D-211 + D-216 |
| `resize_200x60.golden` | info panel populated + crumbs row + shrunk status bar | full-tier, all changes visible |

**40×12 crumbsHeight contract change:** previously `crumbsHeight=0` so body had `height-1(chrome)-1(status)=10` rows; post-Phase-8 body has `height-1(chrome)-1(crumbs)-1(status)=9` rows. UI-SPEC §"Per-tier visibility verification" lists this explicitly. Plan 3 author MUST verify the body-reachable contract still holds at 40×12 (the file list still renders inside the titled border).

---

### `internal/ui/statusbar_test.go` (modify)

**Plan:** Plan 2
**Analog:** itself.

**Tests to delete** (per D-209 + D-211 — center/pipe sections gone):

- `TestStatusBarItemCountInView` (line 31-37) — center section deleted
- `TestStatusBarSetItemCount` (line 129-135) — `SetItemCount` becomes no-op or deleted

**Test to modify** (`TestStatusBarBreadcrumbInView` at line 23-28 — breadcrumb no longer in View, only in Segments accessor):

```go
// BEFORE
func TestStatusBarBreadcrumbInView(t *testing.T) {
    m := ui.NewStatusBarModel(defaultEnv())
    m.SetBreadcrumb("files")
    view := m.View(80)
    assert.Contains(t, view, "files", ...)
}

// AFTER — Phase 8 D-210 — breadcrumb is read via Segments(), not View().
func TestStatusBarSegmentsAccessor(t *testing.T) {
    m := ui.NewStatusBarModel(defaultEnv())
    m.SetBreadcrumb("files", "prod.yaml")
    segs := m.Segments()
    assert.Equal(t, []string{"sops-tui", "files", "prod.yaml"}, segs,
        "Segments must round-trip through SetBreadcrumb (D-210)")
}
```

**New tests** (D-219 - `TestStatusBar_RightAlignOnly` + `TestStatusBar_SegmentsAccessor`):

```go
// TestStatusBarRightAlignOnly verifies post-Phase-8 status-bar shape:
// no breadcrumb in View(), no item count, no pipe separators (D-209, D-211).
func TestStatusBarRightAlignOnly(t *testing.T) {
    m := ui.NewStatusBarModel(defaultEnv())
    m.SetBreadcrumb("files", "prod.yaml")
    view := m.View(80)
    stripped := ansi.Strip(view)
    assert.NotContains(t, stripped, "prod.yaml",
        "Phase 8 D-210: breadcrumb must NOT appear in status bar View()")
    assert.NotContains(t, stripped, "|",
        "Phase 8 D-211: pipe separators must be deleted")
    assert.Contains(t, stripped, "sops:",
        "env indicators must still appear right-aligned")
}
```

(Plan 2 author: keep `TestStatusBarFlash*` tests unchanged — D-212 flash path is preserved.)

---

## Shared Patterns

### Pattern: `lipgloss.AdaptiveColor` ban

**Source:** `internal/ui/styles.go:5-6` + project memory + every package doc-comment.
**Apply to:** all new/modified files.

```go
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
```

Every Phase 8 file must include this stanza in its package-level doc comment (mirrors `logo.go:14`, `menu.go:24`, `chrome.go:33`, `recipientform.go:14`).

---

### Pattern: package-var styles (Pitfall 2 / D-22 / TestViewNoNewStyle)

**Source:** `internal/ui/styles.go` (the entire file).
**Apply to:** every new style added by Phase 8 — declare in `styles.go`, never inline in `View()` or any helper reachable from View().

```go
// Pattern: declare style once at package scope.
SomeStyle = lipgloss.NewStyle().Foreground(ColorMuted).Width(N)

// In View() / RenderX(): reference the package var directly.
return SomeStyle.Render(text)
```

Both walkers (`TestViewNoNewStyle` BFS in `internal/app/`, `TestSubmodelViewsNoNewStyle` file-scope in `internal/ui/`) enforce this. Plan 3 extends the latter's allowlist to include the new files.

---

### Pattern: explicit hex colors (no AdaptiveColor)

**Source:** `internal/ui/styles.go:11-54`.
**Apply to:** every new color reference in Phase 8.

```go
// All colors are explicit hex constants:
ColorAccentHex = "#89b4fa"
ColorAccent    = lipgloss.Color(ColorAccentHex)
```

Phase 8 introduces NO new hex values — every chip/info-panel style references the existing 8-color palette (`ColorBg`, `ColorSurface`, `ColorAccent`, `ColorMuted`, `ColorFg`).

---

### Pattern: ASCII-only chrome (UI-15 / Pitfall 6 / TestChromeASCIIOnly)

**Source:** `internal/app/chrome_test.go:47-88`.
**Apply to:** `infopanel.go` + `crumbs.go` (new) — both grep-gated by Plan 3.

The allowlist already includes `…` (U+2026) for middle-truncation. Plan 1 author MUST NOT introduce any other non-ASCII codepoint in either file.

---

### Pattern: `gogit.PlainOpenWithOptions` + `ErrRepositoryNotExists` for non-git directories

**Source:** `internal/git/status.go:51-54` (`IsGitRepo`), `:65-73` (`GetFileStatuses`), `:127-130` (`GetFileHistory`), `:212-215` (`GetLastCommitTime`).
**Apply to:** `GetBranch` (Plan 2) — the project's canonical entry point for go-git.

```go
repo, err := gogit.PlainOpenWithOptions(repoRoot, &gogit.PlainOpenOptions{DetectDotGit: true})
if err == gogit.ErrRepositoryNotExists {
    return /* zero values */, gogit.ErrRepositoryNotExists
}
if err != nil {
    return /* zero values */, err
}
```

`GetBranch` matches `GetFileStatuses` semantics — non-git is not an error condition; the caller treats `ErrRepositoryNotExists` as the "show `-`" state (D-204).

---

### Pattern: cached field on `AppModel` + handler refresh (Pitfall 15 / event-driven)

**Source:** `internal/app/model.go:237` (`m.files`), `:252` (`m.gitRepoRoot`), `:265` (`m.currentParsed`).
**Apply to:** `m.infoPanel` (Plan 3 D-213).

```go
// Pattern: declare field on AppModel (line ~265 area).
infoPanel ui.InfoPanelData

// Refresh in event handlers: FilesDiscoveredMsg (line 328), GitStatusMsg (line 586),
// recipient/edit ops (multiple sites).

// Read-only access in View() via m.infoPanel — zero I/O.
chrome := ui.RenderChrome(hints, ui.LogoInfo, m.infoPanel, m.width)
```

`View()` MUST NOT call `os.Stat`, `parser.ParseFile`, `git.*`, or any I/O. The cache is the single source of truth.

---

### Pattern: pure-function renderer with package-var styles (D-22 + Pitfall 2)

**Source:** `internal/ui/logo.go:52-63` (`RenderLogo`) + `internal/ui/menu.go:57-113` (`RenderMenu`).
**Apply to:** `RenderInfoPanel`, `RenderCrumbs`.

Each renderer:
1. Takes pre-computed value-typed input (no pointers, no model coupling).
2. Returns a `string`.
3. References package-var styles only (never `lipgloss.NewStyle()` in the body).
4. Is testable in isolation via unit tests.

---

### Pattern: D-220 Pitfall 11 security review (5-question gate per field)

**Source:** CONTEXT.md D-220 + UI-SPEC §accessibility.
**Apply to:** Plan 3's `08-03-SUMMARY.md` sign-off table.

Per info-panel field, answer:
1. Does this field derive from private key material? (`age:` → yes; truncate to ≤10 cells + `…`; `Recipient().String()` only, never `identity.String()`.)
2. Could this field expose absolute filesystem paths? (`cfg:` → must be `filepath.Rel(cwd, sopsYamlPath)`; `agekey.go` reads `$SOPS_AGE_KEY_FILE` but never renders the path itself.)
3. Does any keybinding copy this field to clipboard? (no — chrome is display-only; no copy bindings target chrome content.)
4. Does this field appear in stderr logs? (no — `internal/ui/errorbox.go` is the only stderr surface.)
5. Could a screenshot of this field, posted publicly, narrow an attacker's search space? (fingerprint truncation + relative path are the mitigations; document residual risk.)

Plan 3 author writes the sign-off table.

---

## No Analog Found

None. Every Phase 8 file has at least a role-match analog in the existing codebase.

| File | Analog Coverage |
|------|-----------------|
| `internal/ui/infopanel.go` | exact (logo.go shape + metadata.go label/value) |
| `internal/ui/crumbs.go` | exact (k9s parity source + menu.go JoinHorizontal) |
| `internal/ui/agekey.go` | role-match (recipientform.go age import + validator/startup.go path) |
| `internal/git/status.go` extension | exact (sibling functions in same file) |
| `internal/ui/styles.go` extension | exact (additive to existing var block) |
| `internal/ui/statusbar.go` modify | exact (self-modify, removal-of-code) |
| `internal/ui/chrome.go` modify | exact (self-modify, signature extension) |
| `internal/app/model.go` modify | exact (self-modify, additive cache field) |
| Test files | exact (sibling tests in same packages) |

---

## Metadata

**Analog search scope:** `internal/ui/`, `internal/app/`, `internal/git/`, `internal/parser/`, `internal/validator/`, `~/git/k9s/internal/ui/` (k9s parity reference).

**Files scanned:**
- `internal/ui/logo.go` (63 lines) — pure-renderer + LogoStatus enum precedent
- `internal/ui/menu.go` (127 lines) — JoinHorizontal manual-columns precedent
- `internal/ui/chrome.go` (260 lines) — RenderChrome composition + 3-tier dispatch
- `internal/ui/styles.go` (338 lines) — package-var style block
- `internal/ui/statusbar.go` (237 lines) — current 3-section View() to be shrunk
- `internal/ui/recipientform.go` (168 lines) — `filippo.io/age` import precedent + parse pattern
- `internal/ui/metadata.go` (130 lines, partial) — label/value buildContentLines pattern
- `internal/git/status.go` (238 lines) — gogit.PlainOpenWithOptions canonical entry point
- `internal/git/status_test.go` (217 lines) — t.TempDir + initRepo + commitFile test scaffolding
- `internal/app/model.go` (selected ranges 220-300, 328-365, 586-614, 1280-1545) — AppModel struct + handlers + View() composition
- `internal/app/chrome_test.go` (315 lines) — grep-gate + BFS walker patterns
- `internal/ui/submodel_view_no_newstyle_test.go` (126 lines) — file-scope walker
- `internal/ui/menu_test.go` (240 lines) — RGB-triplet style assertion + ASCII allowlist patterns
- `internal/ui/logo_test.go` (92 lines) — ansi.Strip + lipgloss.Width + RGB triplet patterns
- `internal/ui/statusbar_test.go` (141 lines) — current shape; tests to delete/modify
- `internal/validator/startup.go` (lines 40-100) — HOME resolution + keys.txt path construction
- `~/git/k9s/internal/ui/crumbs.go` (74 lines) — VERBATIM source for chip pill format (D-205, D-207, D-208)

**Verified line numbers (per RESEARCH.md, re-verified 2026-04-28):**
- `internal/app/model.go:1296` AppModel.View() — VERIFIED
- `internal/app/model.go:1345` first RenderChrome call-site — VERIFIED
- `internal/app/model.go:1414` currentFileBreadcrumb() — VERIFIED
- `internal/app/model.go:1437` bodyDims() — VERIFIED
- `internal/app/model.go:1528` chromeHeight() — VERIFIED (note: line 1528 holds the func declaration; the RenderChrome call is at line 1532)
- `internal/app/model.go:1532` second RenderChrome call-site — VERIFIED
- `internal/app/model.go:1539` crumbsHeight() stub returning 0 — VERIFIED
- `internal/ui/chrome.go:106` RenderChrome signature — VERIFIED
- `internal/ui/styles.go:257-263` InfoPanelPlaceholderStyle — VERIFIED
- `internal/ui/recipientform.go:22` filippo.io/age import — VERIFIED
- `internal/parser/yaml.go:31, 125` ParsedFile.AgeRecipients — VERIFIED via grep references in code; not directly read in this pattern map but unchanged from RESEARCH.md citation

**Pattern extraction date:** 2026-04-28
