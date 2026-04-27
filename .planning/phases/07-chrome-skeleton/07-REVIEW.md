---
phase: 07-chrome-skeleton
reviewed: 2026-04-27T13:03:52Z
depth: deep
files_reviewed: 22
files_reviewed_list:
  - internal/keys/hints.go
  - internal/keys/hints_test.go
  - internal/ui/logo.go
  - internal/ui/logo_test.go
  - internal/ui/menu.go
  - internal/ui/menu_test.go
  - internal/ui/chrome.go
  - internal/ui/chrome_test.go
  - internal/ui/styles.go
  - internal/ui/filelist.go
  - internal/ui/detail.go
  - internal/ui/help.go
  - internal/ui/diff.go
  - internal/ui/metadata.go
  - internal/ui/health.go
  - internal/ui/history.go
  - internal/ui/recipientform.go
  - internal/app/model.go
  - internal/app/chrome_test.go
  - internal/app/hints_test.go
  - internal/app/layout_test.go
  - internal/app/testdata/resize_40x12.golden
findings:
  critical: 0
  warning: 5
  info: 7
  total: 12
status: issues_found
---

# Phase 7: Code Review Report

**Reviewed:** 2026-04-27T13:03:52Z
**Depth:** deep
**Files Reviewed:** 22
**Status:** issues_found

## Summary

Phase 7 chrome-skeleton is structurally well-designed: the keys/hints contract is
clean, the chrome composer has clear responsibility boundaries, the AST walker
catches the targeted regression class, and the new code is free of concurrency
issues, dead code, hardcoded secrets, or other security concerns. Test discipline
is strong (compile-time interface assertions, exact-copy hint locks, RGB triplet
checks for color regressions).

The review surfaces five WARNING-class items that warrant attention before Phase 8
tightens chrome contracts further:

1. The `TestViewNoNewStyle` AST walker does NOT scan named helpers reachable from
   `View()` (`renderRecipientList`, `renderFormatMenu`). Three concrete
   `lipgloss.NewStyle()` allocations live in those helpers today.
2. Six sub-model `View()` methods still wrap their content in a `RoundedBorder`
   box, then `WrapTitled` wraps that with `NormalBorder` — producing nested
   double borders and nested-box width math. The verifier already flagged this;
   I cross-cite the per-file lines and surface the correctness implication for
   `m.width` math (each sub-model also subtracts `m.width - 2` for its own
   border, double-counting the chrome's border budget).
3. The `RenderChrome` width clamp at line `chrome.go:65` allows the rendered
   chrome to exceed the terminal width at narrow widths (40-col case
   reproducible in `resize_40x12.golden`), and the menu `lipgloss/v2/table`
   cell-wraps text at sub-cell widths producing 16-row chrome at 80 cols
   (visible in `resize_80x24.golden`). Verifier note 3 documents this; I
   surface the user-visible breakage in the golden files.
4. `TestRenderMenu_ASCIIOnlyBody` (menu_test.go:207-217) allowlists
   rounded-corner glyphs `╭╮╰╯` that `NormalBorder` never emits — this
   inconsistency with the authoritative `TestChromeASCIIOnly` allowlist
   weakens the grep-gate guarantee.
5. The dispatcher inconsistency between `DiffModel.Hints()` (6 entries
   including `q quit`) and the inline `RecipientConfirmHints`/
   `BulkReKeyConfirmHints` (5 entries, no `q`) is silent — when the same
   diff body shows in two states, the menu loses the quit affordance
   without an explicit decision recorded against UI-SPEC.

INFO-class items cover dead-code stubs, minor style issues, and defensive
suggestions for future refactoring.

## Warnings

### WR-01: AST walker does not scan named helpers called from `View()`

**File:** `internal/app/chrome_test.go:137-184`, `internal/app/model.go:1935,1940,1981`
**Issue:** `TestViewNoNewStyle` declares it enforces "no `lipgloss.NewStyle()`
inside this body or any helper lambda" (per `model.go:1293-1294`), but the
walker's scope is the `*ast.BlockStmt` of `AppModel.View()` only. Named helpers
reachable from `View()` are not scanned. Three concrete violations exist today
in helpers reached every frame from the View dispatcher:

- `model.go:1935` — `lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(...)`
  inside `renderRecipientList` (called from View at `model.go:1331`).
- `model.go:1940` — `lipgloss.NewStyle().Foreground(ui.ColorMuted).Render(...)`
  inside `renderRecipientList` (same call site).
- `model.go:1981` — `lipgloss.NewStyle().Border(...).Render(inner)` inside
  `renderFormatMenu` (called from View at `model.go:1321`).

These allocate fresh styles per frame, contribute to the 2.4-2.8 ms/op
benchmark regression, and falsify the `model.go:1293-1294` doc comment's
"no `lipgloss.NewStyle()` inside this body or any helper" claim. The walker
also does not catch the 12 sub-model `lipgloss.NewStyle()` calls
(`help.go:82,98`, `health.go:128,178`, `recipientform.go:132,163`,
`metadata.go:83,84,85,158,174`, `diff.go:177`, `history.go:86,134`) — all
reached from View on every frame.

**Fix:**
1. Replace the three `lipgloss.NewStyle()` calls in `model.go` helpers with
   existing package vars (e.g., reuse `ui.GitNoRepoStyle` for the muted
   "(showing first %d of %d recipients)" line and the "Select recipient
   to remove" prompt; lift the format menu's RoundedBorder style to a
   package var named e.g. `ui.FormatMenuOverlayStyle`).
2. Extend the AST walker to follow `*ast.CallExpr` from `View()` into
   functions in the same package and recurse one or two levels (configurable),
   OR widen the scan to file-scope and allowlist only known cold-path
   `NewStyle()` calls (constructor `New*` functions). The first option is
   strictly more correct; the second is a faster patch.
   ```go
   // Sketch: walk all FuncDecls in model.go; if any are reachable from View,
   // also AST-Inspect their bodies. Build a small reachability set rooted at
   // View, then assert no NewStyle in the set.
   reachable := map[string]bool{"View": true}
   // Pre-pass: collect call edges (function name -> set of called names).
   // Then BFS from "View" and mark each helper, then run the same Inspect
   // over the FuncDecls in the reachable set.
   ```

---

### WR-02: Six sub-model `View()` methods produce nested double borders

**File:** `internal/ui/help.go:98-106`, `internal/ui/health.go:178-186`,
`internal/ui/recipientform.go:163-171`, `internal/ui/metadata.go:174-182`,
`internal/ui/diff.go:177-185`, `internal/ui/history.go:134-142`
**Issue:** Each of the six sub-models renders its own `RoundedBorder`
(`m.width - 2`, `m.height - 2`) box inside `View()`. AppModel.View() at
`model.go:1342` then wraps that body via `ui.WrapTitled(title, body, w, h)`
with `NormalBorder()`. The result is a nested double border on every overlay
state plus correctness drift in width math: the sub-model subtracts 2 for
its own border, then `WrapTitled` subtracts another 4 (border + Padding(0,1)),
so the inner content area is `m.width - 6`, not `m.width - 4` as the chrome
contract assumes. The verifier already flagged this as a known issue —
this finding cross-cites the exact file lines and surfaces the width-math
correctness implication.

The grep-gate `TestChromeNormalBorderOnly` only scans
`internal/ui/{chrome,logo,menu}.go` (`chrome_test.go:103-107`), so
`RoundedBorder` calls in sub-models are intentionally outside its scope.
That keeps the chrome guarantee crisp but leaves the nested-border
contradiction undocumented.

**Fix:**
Pick one of two options and document it:
1. **Strip sub-model borders.** Each sub-model's `View()` returns plain
   inner content; `WrapTitled` is the single border source. Pros: matches
   D-12/D-13 chrome contract; eliminates double-render. Cons: legacy
   visual signature changes; goldens regenerate.
   ```go
   // Example: ui/diff.go View()
   // Delete lines 168-184; just `return inner`.
   ```
2. **Skip `WrapTitled` for these states** (like `stateFormatMenu` already
   does at `model.go:1339-1340`). Pros: preserves legacy look. Cons:
   inconsistent chrome contract; titles disappear on these states; the
   chrome's "every primary view wears the same border" promise is broken.

Option 1 is the cleaner fix and aligns with the explicit "stateFormatMenu
opts out — it renders its own RoundedBorder overlay" comment at
`model.go:1336-1337` (i.e., the option-2 carve-out is currently the only
documented exception).

---

### WR-03: Chrome rendered width can exceed terminal width at narrow sizes

**File:** `internal/ui/chrome.go:60-71`, `internal/app/testdata/resize_40x12.golden`,
`internal/app/testdata/resize_80x24.golden`
**Issue:** The width composition in `RenderChrome` is:
```
infoPanelWidth (38) + menuWidth (clamped >=1) + logoWidth (25) = >=64
```
At width=40, the infoPanelWidth alone (`InfoPanelPlaceholderStyle.Width(38)`
at `styles.go:252-254`) consumes 38 cols; logo is 25 cols; menu clamps to
1 col. Total rendered chrome width is 64 cols — 24 cols wider than the
terminal. `resize_40x12.golden` shows the result: chrome wraps to 78+
visible rows in a 12-row terminal, completely obscuring the body.

At width=80, the `lipgloss/v2/table` cell-wraps at narrow column widths
producing the broken `[enter/l\n] open  toggle\n        help` layout
visible in `resize_80x24.golden:7-9`. This balloons chrome from 6 rows
to 16 rows; `chromeHeight(m)` correctly reports 16, so the body box does
shrink — but the table-cell wrapping itself is the user-visible artifact.

The verifier already flagged this as a known performance/UX issue; this
finding records the exact reproduction path in the goldens so the fix is
testable.

**Fix:** Two complementary clamps.
1. Clamp the chrome's outer width to `min(width, infoPanelWidth + logoWidth + minMenuCol)`
   — when the terminal cannot hold full chrome, drop the info-panel
   placeholder OR collapse to a single-line breadcrumb-style hint row
   (deferred to Phase 10 per UI-16, but a hard clamp now prevents the
   negative menuWidth → cell-wrap death spiral).
   ```go
   // chrome.go RenderChrome
   if width < infoPanelWidth + logoWidth + 8 {
       // Phase 7 fallback: drop the info panel; render menu+logo only
       menuWidth := width - logoWidth
       if menuWidth < 8 {
           // Even smaller — render a one-line "press ? for help" stub.
           return ChromeNarrowFallbackStyle.Render("press ? for help")
       }
       menu := RenderMenu(hints, menuWidth)
       logo := RenderLogo(logoStatus, logoWidth)
       return lipgloss.JoinHorizontal(lipgloss.Top, menu, logo)
   }
   ```
2. Inside `RenderMenu`, replace the `lipgloss/v2/table` builder with a
   manual JoinHorizontal of two pre-rendered columns at `lipgloss.NewStyle().Width(colW)`
   so cell wrapping never engages. Each cell becomes
   `MenuKeyStyle.Render("[k]") + " " + MenuDescStyle.MaxWidth(colW - keyW - 2).Render(desc)`.
   Side benefit: removes the 394 µs lipgloss/v2/table contribution to the
   bench, addressing the chrome_test.go:204-239 5 ms budget headroom.

---

### WR-04: `TestRenderMenu_ASCIIOnlyBody` allowlists glyphs `NormalBorder` never emits

**File:** `internal/ui/menu_test.go:207-217`
**Issue:** The runtime allowlist contains rounded corners `╭╮╰╯` plus
NormalBorder verticals/horizontals `─│`, but `NormalBorder()` produces
SQUARE corners `┌┐└┘`. The authoritative grep-gate
`TestChromeASCIIOnly` (chrome_test.go:46-54) correctly allowlists `┌┐└┘`
and explicitly comments at lines 32-35 that "NormalBorder() emits SQUARE
corners (┌┐└┘), NOT rounded (╭╮╰╯)".

The menu test contradicts that. Today this is harmless (RenderMenu's
table is borderless — `BorderTop(false)` etc. at `menu.go:73-79` — so no
corner glyphs reach the output), but if a future change re-enables a
border on the menu table, the menu test will silently accept the wrong
corner family.

**Fix:** Align the menu test's allowlist with `TestChromeASCIIOnly`:
```go
// menu_test.go
allowed := map[rune]bool{
    '\n': true,
    '─':  true, '│': true,
    '┌':  true, '┐': true, '└': true, '┘': true,  // square corners (NormalBorder)
    '↑':  true, '↓': true, '←': true, '→': true,
}
// Drop ╭╮╰╯ entirely — they are never produced by NormalBorder.
```
Also worth considering: extract the allowlist into `internal/ui/internal/testutil`
or similar so all chrome-adjacent ASCII tests share one source of truth.

---

### WR-05: Quit hint disappears in `stateRecipientConfirm` / `stateBulkReKeyConfirm`

**File:** `internal/keys/hints.go:68-74,76-84`, `internal/ui/diff.go:192-200`,
`internal/app/model.go:1466-1469`
**Issue:** When the user is on a Diff body (`stateDiff`), the persistent
menu shows 6 hints including `[q] quit` (per `DiffModel.Hints()` at
`diff.go:192-200`). When the same diff body is shown for recipient-action
confirmation (`stateRecipientConfirm`) or bulk re-key (`stateBulkReKeyConfirm`),
the dispatcher at `model.go:1466-1469` returns the inline `keys.RecipientConfirmHints`
or `keys.BulkReKeyConfirmHints` — both 5 entries with no `q` quit affordance.

The user perceives a body that looks identical, but the hint set silently
drops "quit". This may be intentional (forcing the user to resolve the
y/n decision before quitting), but the hints.go inline package-var doc
comments don't say so — they only say "Disambiguates the shared stateDiff
body via AppModel.state". The 07-PATTERNS or 07-VALIDATION docs may have
context, but if not, this is silent UI behaviour drift.

**Fix:** Decide and document.
- Option A (preserve quit): append `{Mnemonic: "q", Description: "quit", Visible: true}`
  to both `RecipientConfirmHints` and `BulkReKeyConfirmHints` and update
  `TestRecipientConfirmHints_ExactCopy` / `TestBulkReKeyConfirmHints_ExactCopy`
  to expect 6 entries.
- Option B (drop quit deliberately): add a doc comment to both inline hint
  sets and the dispatcher arms in `menuHints()` explaining "quit suppressed
  during confirm flows so the user resolves y/n first" with a UI-SPEC
  cross-reference.

Either option is acceptable; the WARNING is the silence.

## Info

### IN-01: `lipgloss.NewStyle()` allocations across 12 sub-model View() sites

**File:** `internal/ui/help.go:82,98`, `internal/ui/health.go:128,178`,
`internal/ui/recipientform.go:132,163`, `internal/ui/metadata.go:83,84,85,158,174`,
`internal/ui/diff.go:177`, `internal/ui/history.go:86,134`
**Issue:** Each of these 12 `lipgloss.NewStyle()` calls runs on every
frame's View(). The styles built are deterministic (no per-frame data) so
they could all be lifted to `internal/ui/styles.go` package vars, matching
the Phase 7 `MenuKeyStyle` / `MenuDescStyle` discipline. This is partial
explanation for the bench at 2.4-2.8 ms/op vs the original 50 µs target.
**Fix:** Lift each to a package var in `styles.go` (e.g., `OverlayMutedFooterStyle`,
`OverlayBorderBoxStyle`, `MetadataLabelStyle`, etc.) and have View() reference
the package var. Tracked separately from WR-01 because WR-01 is about the
walker's scope; this item is the broader allocation hygiene.

---

### IN-02: `BadgeUntracked` package var has misaligned formatting

**File:** `internal/ui/styles.go:181-183`
**Issue:** `BadgeUntracked` has an extra space before `=` that breaks the
visual column alignment with `BadgeModified` and `BadgeAdded` immediately
above. Pre-existing (not Phase 7), but worth a `gofmt`-style cleanup since
this file was modified in Phase 7.
**Fix:**
```go
BadgeModified  = lipgloss.NewStyle().Foreground(ColorWarning)
BadgeAdded     = lipgloss.NewStyle().Foreground(ColorSuccess)
BadgeUntracked = lipgloss.NewStyle().Foreground(ColorMuted)
```

---

### IN-03: `crumbsHeight` parameter discard via `_ = m`

**File:** `internal/app/model.go:1535-1538`
**Issue:** `crumbsHeight(m AppModel) int` discards `m` via `_ = m` and
returns 0. Since `m` is unused and the function is package-private, the
parameter could be dropped entirely until Phase 8 wires it up.
**Fix:** Either keep as-is (Phase 8 will use `m`), or change to
`func crumbsHeight() int { return 0 }` and update the two call sites in
`model.go` (1352 and 1439). The `_ = m` is harmless either way; this is a
style preference call. Recommend keeping as-is to minimise churn.

---

### IN-04: `RenderLogo` discards `width` parameter via `_ = width`

**File:** `internal/ui/logo.go:52-53`
**Issue:** Same pattern as IN-03 — `width` reserved for Phase 10. Doc
comment is clear about why. No action.
**Fix:** None needed; flagged for completeness so it isn't quietly
removed in a future cleanup pass that misses the Phase 10 plumbing
intent.

---

### IN-05: `DetailModel.Hints()` linear scan for "b" mnemonic is fragile

**File:** `internal/ui/detail.go:820-830`
**Issue:** The override walks `hints` looking for `Mnemonic == "b"` and
flips `Visible=false`. If `bindings.go` ever changes the Blame mnemonic
from `"b"` (e.g. to `"B"` to disambiguate from a future bulk operation),
this loop silently does nothing — the test
`TestMenuHints_StateDetail` only asserts `len(hints) == 13`, not which
hint is hidden.
**Fix:** Either (a) reach for `keys.DefaultDetailKeyMap.Blame.Help().Key`
to derive the expected mnemonic at runtime, or (b) add a test that
asserts the Blame hint specifically is `Visible=false`:
```go
// detail_test.go (new test)
func TestDetailModel_HintsBlameInvisible(t *testing.T) {
    m := NewDetailModel(...)
    hints := m.Hints()
    found := false
    for _, h := range hints {
        if h.Description == "blame line" || h.Mnemonic == "b" {
            assert.False(t, h.Visible, "Blame hint must be Visible=false (D-06)")
            found = true
        }
    }
    require.True(t, found, "Blame hint must exist in DetailModel.Hints()")
}
```

---

### IN-06: `MenuHint.Mnemonic` empty-string is silently rendered as `[]`

**File:** `internal/ui/menu.go:70`
**Issue:** `RenderMenu` composes `"["+h.Mnemonic+"]"` with no validation.
If `h.Mnemonic == ""`, the rendered cell is `[] description` — visually
broken but not an error. Today no producer emits empty mnemonics so this
is theoretical; flagging for future-proofing.
**Fix:** Either skip empty-mnemonic hints in the visible filter at
`menu.go:48-55`, or add a panic/log assertion in `keys.HintsFromBindings`
when `h.Key == ""`. Recommend skipping silently — defensive, no behavioural
surprise.

---

### IN-07: `spliceRenderedLine` doc says "ANSI sequences are written through
verbatim" but only matches SGR (`m`-terminated) sequences

**File:** `internal/ui/chrome.go:155-197`
**Issue:** The function tracks ESC-introduced sequences and ends them on
seeing `m`. Other ANSI CSI terminators (`H`, `J`, `K`, cursor-move codes,
etc.) and non-CSI escapes (OSC `]`, single-shift `c`) would be misclassified
— accumulated indefinitely or until the next `m`. Lipgloss only emits SGR
today, so this is theoretical, but the doc comment at line 152 ("pure
ASCII-range box-drawing characters with no ANSI sequences embedded inside
the border characters themselves") is overly broad: it should call out the
SGR-only assumption explicitly.
**Fix:** Tighten the doc comment to "this function assumes lipgloss-emitted
SGR sequences (`ESC [ ... m`) and is NOT a full ANSI parser." Optional:
in addition to `r == 'm'`, end the SGR mode on any byte in the CSI final
range `0x40..0x7E` to be safe. This costs nothing and prevents pathological
inputs from looping forever.

---

_Reviewed: 2026-04-27T13:03:52Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: deep_
