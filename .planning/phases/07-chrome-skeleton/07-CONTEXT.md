# Phase 7: Chrome Skeleton - Context

**Gathered:** 2026-04-24
**Status:** Ready for planning

<domain>
## Phase Boundary

First visible "looks like k9s" step. Add a persistent 6-row ASCII logo (top-right) and a persistent 2-column × 6-row keybinding menu (header), wrap every primary view in a titled bordered region via a central `WrapTitled` helper, and flip the `chromeHeight(m)` stub from Phase 6 to its real value. `AppModel.View()` recomposes to `[chrome][crumbs-placeholder][titled body][status bar]`.

**In scope (this phase):**
- `internal/keys/hints.go` — `MenuHint` struct, `Hinter` interface, `HintsFromBindings` converter
- `internal/ui/logo.go` — 6-row ASCII art + `RenderLogo`
- `internal/ui/menu.go` — `RenderMenu` built on `lipgloss/v2/table`
- `internal/ui/chrome.go` — chrome composer + `WrapTitled` + `overlayTitle` string-splice helper
- `Hints()` on all 9 interactive sub-models (FileList, Detail, Help, Diff, Metadata, Health, History, RecipientList, RecipientForm)
- AppModel dispatcher for hints keyed on `(state, recipientAction, IsSearchActive)`
- `chromeHeight` stub flipped to the real value; `AppModel.View()` rewritten to compose through chrome
- Grep-gate tests: chrome ASCII-only, `NormalBorder()`-only, no `lipgloss.NewStyle()` inside `View()`
- `BenchmarkAppView ≤ 50 µs/op @ 200×60` gate
- Migrate `model.go:1841` magic `m.height - 4` to `bodyDims`-based computation (the Phase 6 deferred TODO)

**Out of scope (deferred per ROADMAP):**
- Header info panel (Phase 8)
- Breadcrumb chip row (Phase 8) — `crumbsHeight` stays stubbed at 0
- Logo severity coupling to env/flash/health (Phase 10, UI-03) — default `ColorAccent` only
- Color-profile detection + 16-color fallback (Phase 10, UI-13) — single-palette only this phase
- k9s-tuned palette tune (Phase 10, UI-12)
- Per-(state, action) golden matrix with hint-vs-keymap drift tests (Phase 9, UI-10)
- Status bar shrink to env + clipboard (Phase 8)
- Alt-screen fill/blank frame on enter/exit (Phase 11)

</domain>

<decisions>
## Implementation Decisions

### Logo (UI-02)
- **D-01:** 6-row stacked "SOPS-TUI" design — 5-row block figlet of "SOPS" + "tui" subscript on row 6. Width ~26 cols. ASCII-only per UI-15 and Pitfall 6 (no emoji, no VS16 variation selectors, no ZWJ).
- **D-02:** Default color for Phase 7 = `ColorAccent`. Severity-coupled recoloring (info / warn / error) is deferred to Phase 10 per UI-03 — `RenderLogo(status LogoStatus, width int)` accepts the parameter but Phase 7 callers pass `LogoInfo` unconditionally.
- **D-03:** Anchored top-right of the header region per ROADMAP UI-02. Positioned by the chrome composer via `lipgloss.JoinHorizontal` with a flexible spacer between info-panel-placeholder, menu, and logo.

### Menu (UI-01)
- **D-04:** Fixed 2 columns × 6 rows = 12 hint slots at all widths. Same layout at 40×12 as at 200×60; narrow-terminal safe. Richer responsive layouts deferred to first bug report.
- **D-05:** Rendered via `lipgloss/v2/table` with `StyleFunc(row, col int) lipgloss.Style`. Mnemonic column uses `MenuKeyStyle` (ColorAccent); description column uses `MenuDescStyle` (ColorFg). Both styles declared as package vars in `internal/ui/styles.go` (Pitfall 2 — zero `NewStyle()` in `View()`).
- **D-06:** `MenuHint.Visible` (bool) controls inclusion in the persistent menu's 12 slots. Sub-models curate which bindings are "persistent-worthy" (≤12); everything remains discoverable in the `?` full-screen overlay per UI-11.
- **D-07:** Cell format is `[mnemonic] description` — mnemonic left-aligned in a fixed-width subcolumn of column 0; description fills column 1.

### Hints() interface (UI-01 / Pitfall 3 prep)
- **D-08:** `internal/keys/hints.go` — new file — defines:
  - `type MenuHint struct { Mnemonic, Description string; Visible bool }`
  - `type Hinter interface { Hints() []MenuHint }`
  - `func HintsFromBindings(bindings []key.Binding) []MenuHint` — converts each `key.Binding.Help()` (`{Key, Desc}`) to a `MenuHint{Mnemonic: Key, Description: Desc, Visible: true}`.
- **D-09:** All 9 interactive sub-models implement `Hints()` in Phase 7: FileList, Detail, Help, Diff, Metadata, Health, History, RecipientList, RecipientForm. The stateFormatMenu modal uses an inline hint set (no owning sub-model). SUMMARY's 4-then-5 split is rejected — closes Pitfall 3 risk in one pass and simplifies Phase 8 scope.
- **D-10:** AppModel hint dispatcher is a pure function of `(state sessionState, recipientAction string, IsSearchActive bool)` per Pitfall 3. Modal states (stateDiff, stateRecipientConfirm, stateBulkReKeyConfirm) each map to their correct hint set via `recipientAction`. stateFormatMenu wired inline.
- **D-11:** When `IsSearchActive()` returns true in stateFileList, the persistent menu shows `[Esc]Exit search / [Enter]Select` (search-scoped hints), not the default FileList hints — Pitfall 3 third axis.

### Titled border (UI-06)
- **D-12:** `WrapTitled(title, body string, width, height int) string` lives in `internal/ui/chrome.go`. Signature per ARCHITECTURE §Pattern 3.
- **D-13:** Uses `lipgloss.NormalBorder()` exclusively (UI-15). Border foreground = `ColorMuted`. No `FocusedBorder`/`UnfocusedBorder` (Pitfall 7).
- **D-14:** Title injection via `overlayTitle(rendered, " Title ") string` helper — string-splice on the first line at column position 2 (two cells in from the left corner). Pattern extracted from `github.com/charmbracelet/soft-serve/pkg/ui/components/header` per SUMMARY's Phase 7 research gap. Source revision cited in a comment block above the helper. Unit test `TestOverlayTitle_PreservesCornersAndWidth` verifies:
  - Top-left corner `╭` and top-right corner `╮` remain intact after overlay
  - Total first-line `lipgloss.Width()` equals the passed-in width
  - Title strings wider than `width - 4` are truncated with ellipsis
  - Empty title renders the border unchanged
- **D-15:** Title format by view — list views use bare `(N)`; Health uses `(N findings)` unit-ful suffix; Detail uses `Detail: <filename>` colon-separated subject:
  | State                  | Title                              | Source of count/subject                             |
  |------------------------|------------------------------------|-----------------------------------------------------|
  | stateFileList          | `Files (N)`                        | `m.fileList.ItemCount()`                            |
  | stateDetail            | `Detail: <filename>`               | `m.currentFile.Name` (no git badge in title)        |
  | stateMetadata          | `Metadata`                         | —                                                   |
  | stateDiff              | `Diff`                             | —                                                   |
  | stateHelp              | `Help`                             | —                                                   |
  | stateHistory           | `History (N)`                      | `m.history.CommitCount()`                           |
  | stateHealth            | `Health (N findings)`              | `m.health.FindingCount()` (unit-ful deliberately)   |
  | stateRecipientList     | `Recipients (N)`                   | `len(m.currentFile.Recipients)` or accessor         |
  | stateRecipientForm     | `RecipientForm`                    | —                                                   |
  | stateFormatMenu        | `Format`                           | — (overlay, not a primary view)                     |
  | stateRecipientConfirm  | `Diff`                             | shared diff body                                    |
  | stateBulkReKeyConfirm  | `Diff`                             | shared diff body                                    |

### chromeHeight + View composition
- **D-16:** `chromeHeight(m AppModel) int` at `internal/app/model.go:1415` flips from `return 0` to the real value. Phase 7 chrome = logo (6 rows) + menu (6 rows) aligned horizontally via `lipgloss.JoinHorizontal`, so `chromeHeight = lipgloss.Height(renderedChrome)` (expected 6). Info-panel area is rendered as a 6-row blank block of width — preserves column alignment for Phase 8 without inflating chromeHeight. `crumbsHeight(m)` stays stubbed at 0 (Phase 8 flips it).
- **D-17:** `AppModel.View()` recomposes to `[chrome][crumbs-placeholder][wrapped body][status bar]` via `lipgloss.JoinVertical(lipgloss.Left, …)`. Sub-model bodies render at `bodyDims(m).w - 2` × `bodyDims(m).h - 2` (accounting for titled border edges); `WrapTitled` then restores the width/height to the full `bodyDims(m)` envelope.
- **D-18:** Chrome rendering strategy: **pure every-frame composition** with all styles as package vars. Rationale: Pitfall 2 prescribes either caching OR zero-allocation pure composition; pure is simpler and avoids mutating state from a value-receiver `View()`. Phase 7's `BenchmarkAppView ≤ 50 µs/op @ 200×60` gate with the "no `lipgloss.NewStyle()` inside `View()`" grep-gate enforces the zero-alloc discipline. If a later palette pass (Phase 10) regresses the bench, caching can be bolted on without public API change.
- **D-19:** `internal/app/model.go:1841` migrates `m.height - 4` magic constant — the Phase 6 deferred TODO. New shape: `renderRecipientList` returns the inner body (pre-border) and `AppModel.View()` calls `WrapTitled("Recipients (N)", body, w, h)` with `w, h := bodyDims(m)`. The magic -4 disappears because the border math moves into `WrapTitled`.

### Grep-gate discipline (UI-15, Pitfall 2, Pitfall 6, Pitfall 7)
- **D-20:** New test `TestChromeASCIIOnly` in `internal/app` — scans `internal/ui/{chrome,logo,menu,crumbs}.go` for non-ASCII codepoints (runes > 0x7F), with an allowlist for box-drawing characters used by `lipgloss.NormalBorder()` (`─`, `│`, `╭`, `╮`, `╰`, `╯`). Fails if any other non-ASCII appears. Scope: chrome files only; status bar flash messages keep emoji per Pitfall 6 guidance.
- **D-21:** New test `TestChromeNormalBorderOnly` — scans chrome files for any of `RoundedBorder|ThickBorder|DoubleBorder|HiddenBorder|FocusedBorder|UnfocusedBorder`. Fails if any appear. Enforces UI-15 NormalBorder() rule and Pitfall 7 no-focus-indicator rule. `ARCHITECTURE.md` Pattern 3 pseudocode shows `RoundedBorder` — the implementation swaps to `NormalBorder` per UI-15 (this is an intentional departure from the architecture sketch).
- **D-22:** New test `TestViewNoNewStyle` — walks `internal/app/model.go` AST, locates the `View()` method body, and fails if any `lipgloss.NewStyle(` literal appears inside it (directly or in helper lambdas). Locks in Pitfall 2's "no per-frame style allocation" rule starting Phase 7 — Phase 11 UI-21 is an orthogonal benchmark gate.
- **D-23:** Phase 6's existing `TestBodyDimsMigration` grep-gate is untouched. Its banned regex (`m\.height\s*-\s*statusBarHeight`) remains in force; Phase 7 does not widen or narrow its scope.

### Benchmark gate (UI-21 preview)
- **D-24:** Extend Phase 6's `BenchmarkAppView` in `internal/app/bench_test.go` — add `TestBenchmarkAppView_UnderBudget` that runs the bench once and asserts `ns/op ≤ 50_000` (50µs) at 200×60 with full Phase 7 chrome rendered. Implemented as a standard Go test (same strategy as the Phase 6 grep-gate) so it runs in `go test ./...` with no new CI infrastructure. Phase 11 UI-21 is the formal sign-off; Phase 7 enforces it early so regressions surface at the source.

### Plan split (3-plan ROADMAP budget)
- **D-25:** Three plans, primitive-first:
  - **Plan 1 — Primitives + hints interface:** `internal/keys/hints.go` (MenuHint, Hinter, HintsFromBindings), `internal/ui/logo.go` (6-row art + RenderLogo with LogoStatus parameter used trivially in Phase 7), `internal/ui/menu.go` (RenderMenu built on lipgloss/v2/table with StyleFunc). New style-var additions to `internal/ui/styles.go` (MenuKeyStyle, MenuDescStyle, LogoStyle{Info,Warn,Error} — Info used; Warn/Error landed for Phase 10 pickup). Unit tests for each primitive. Zero AppModel changes in this plan.
  - **Plan 2 — Chrome composer + WrapTitled + overlayTitle research:** `internal/ui/chrome.go` — chrome composer, `WrapTitled`, `overlayTitle` helper with soft-serve reference comment. `TestOverlayTitle_PreservesCornersAndWidth` unit test covering corners, width preservation, overlong title truncation, empty title. Zero AppModel changes in this plan.
  - **Plan 3 — Integration + `Hints()` on 9 sub-models + gates:** Flip `chromeHeight` stub. Rewrite `AppModel.View()` composition. Add `Hints()` to all 9 sub-models. Add AppModel hint dispatcher (state, recipientAction, IsSearchActive) + stateFormatMenu inline. Migrate `renderRecipientList`'s magic `m.height - 4` (D-19). Add three grep-gate tests (D-20, D-21, D-22). Add bench-budget test (D-24). Add chrome goldens at 40×12 / 80×24 / 120×40 / 200×60 exercising Phase 6's `testutil/golden.go` split (structure + color presence). One atomic PR; 2-3 commits acceptable.
- **D-26:** Plan 3 is deliberately the largest — the integration is tightly coupled (chromeHeight flip, View composition, Hints dispatcher, all grep-gates, goldens). Splitting by sub-model family would require re-opening `AppModel.View()` multiple times and multiply golden-file churn.

### Folded Todos
- **T-01:** Phase 7 research pass on `overlayTitle` (from STATE.md Pending Todos) — folded into Plan 2 per D-14. Closure deliverable: `overlayTitle` helper in `internal/ui/chrome.go` with a source-revision comment citing `charmbracelet/soft-serve/pkg/ui/components/header`, and `TestOverlayTitle_PreservesCornersAndWidth` covering corners, width, truncation, and empty-title behaviour.

### Claude's Discretion
- Exact logo byte-art within the 6×~26 envelope (cleanest figlet variant that fits "SOPS" block + "tui" subscript).
- `overlayTitle` implementation details — `lipgloss.Width`-based measurement of first line, insertion at column 2, handling of title strings wider than `width − 4` (truncate with ellipsis).
- `RenderLogo(LogoStatus, width)` return shape — whether the logo + optional 1-row status message share the same function or split into two (Phase 10 decision either way).
- `RenderMenu` column-width math — how wide the mnemonic subcolumn is (likely max-mnemonic-width + 1 cell).
- Golden file naming — `chrome_stateFileList_80x24.golden` vs per-state directory layout.
- Exact new line number where `chromeHeight`'s real body lands after the stub flip (depends on the composer's final shape).
- Whether `stateFormatMenu` gets a titled border (it's an overlay, so possibly no) — Plan 3 author decides.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project decision docs
- `.planning/ROADMAP.md` §"Phase 7: Chrome Skeleton" — Goal, 5 success criteria, 3-plan budget, UI-01/UI-02/UI-06/UI-15 requirements, phase-boundary discipline
- `.planning/ROADMAP.md` §"Milestone v1.1 — Explicitly Out of Scope for v1.1" — reject list (polling goroutines, `:` command bar, responsive column hiding, skin loader, etc.); plan review must reject any resurfacing
- `.planning/REQUIREMENTS.md` §"Header Region" — UI-01 (persistent menu), UI-02 (6-row logo top-right), UI-03 (logo severity — deferred to Phase 10)
- `.planning/REQUIREMENTS.md` §"Content Framing" — UI-06 (titled bordered regions, item-count titles)
- `.planning/REQUIREMENTS.md` §"Theming & Accessibility" — UI-15 (ASCII-only chrome, NormalBorder-only, grep-gated)
- `.planning/PROJECT.md` — stack constraints (Go + Bubble Tea v2, no framework swap), v1.0 functional-core non-regression rule, k9s-visual-parity product goal

### Research — v1.1 milestone (MUST READ)
- `.planning/research/SUMMARY.md` §"Phase 7: Chrome skeleton — logo + menu + titled borders" — delivers list (logo, menu, chrome composer, WrapTitled, Hints on 4 sub-models — extended here to 9 per D-09), discipline rules, `lipgloss/v2/table` recommendation, bench gate ≤ 50 µs/op
- `.planning/research/SUMMARY.md` §"Research gaps" — Phase 7 `overlayTitle` reference implementation extraction gap (closed by D-14 + T-01)
- `.planning/research/ARCHITECTURE.md` §"File inventory" — signatures for `chrome.go`, `logo.go`, `menu.go`, `keys/hints.go`
- `.planning/research/ARCHITECTURE.md` §"Pattern 1: Immediate-Mode Chrome Composition" + §"Pattern 3: Uniform Titled Border via Central Helper" + §"Pattern 4: Data-Derived Logo State" — the three patterns that govern this phase
- `.planning/research/ARCHITECTURE.md` §"Phase Sequencing" — Phase 7 depends on Phase 6 layout groundwork (already landed)
- `.planning/research/PITFALLS.md` §"Pitfall 2: Chrome Renders on Every View()…" — zero `NewStyle()` in View; benchmark gate; cached-vs-pure analysis (D-18 picks pure)
- `.planning/research/PITFALLS.md` §"Pitfall 3: Menu Hints Derived From Wrong State…" — (state, recipientAction, IsSearchActive) triad (D-10, D-11)
- `.planning/research/PITFALLS.md` §"Pitfall 6: Unicode Width Miscalculation…" — no emoji in chrome; ASCII letters + color instead (D-20)
- `.planning/research/PITFALLS.md` §"Pitfall 7: Focus Indicators Suggest Interactivity…" — no `FocusedBorder`/`UnfocusedBorder` (D-21)
- `.planning/research/PITFALLS.md` §"Pitfall 9: Color-Only Status Indicators…" — chip text content primary, color secondary (Phase 10 territory but informs Phase 7's no-color-only rule in menu)
- `.planning/research/PITFALLS.md` §"Pitfall 12: lipgloss Border Characters Render Differently…" — font-coverage caveat on macOS Terminal; NormalBorder() is the safe choice
- `.planning/research/PITFALLS.md` §"Looks Done But Isn't" checklist — Phase 7 items (persistent menu renders every view, logo ASCII-only, titled border uniform, no NewStyle in View)
- `.planning/research/STACK.md` §"charm.land/lipgloss/v2" §"lipgloss/v2/table" — menu renderer dep; already pulled in transitively

### Upstream Phase 6 deliverables (reused here)
- `internal/app/model.go:1394` — `statusBarHeight(m)` helper (unchanged)
- `internal/app/model.go:1403` — `bodyDims(m)` helper (unchanged; sub-models size through this)
- `internal/app/model.go:1415` — `chromeHeight(m)` stub (FLIPPED in Plan 3)
- `internal/app/model.go:1423` — `crumbsHeight(m)` stub (UNCHANGED in Phase 7; Phase 8 flips)
- `internal/testutil/golden.go` — `RequireGoldenStructure` + `RequireGoldenColors` (exercised in Plan 3)
- `internal/app/bench_test.go` — `BenchmarkAppView` baseline (Plan 3 adds `TestBenchmarkAppView_UnderBudget` at ≤ 50 µs/op)
- `internal/app/resize_test.go` — resize goldens at 40×12 / 80×24 / 120×40 / 200×60 (Phase 7 updates these goldens as chrome lands)
- `.planning/phases/06-layout-groundwork/06-CONTEXT.md` — Phase 6 decisions (carried forward)

### Existing codebase patterns (Phase 7 builds on, doesn't change)
- `internal/ui/styles.go` — design system; new styles added (MenuKeyStyle, MenuDescStyle, LogoStyle{Info,Warn,Error}, TitledBorderStyle, TitleLabelStyle); explicit hex colors only (no AdaptiveColor, v1.0 discipline)
- `internal/keys/bindings.go` — existing `key.Binding` values + `ShortHelp()`/`FullHelp()` methods; new `Hinter` implementations consume via `HintsFromBindings`
- `internal/ui/filelist.go`, `detail.go`, `help.go`, `diff.go`, `metadata.go`, `health.go`, `history.go`, `recipientform.go` — each gains one `Hints() []keys.MenuHint` method in Plan 3; existing `View()` bodies unchanged

### Technology / external references
- `CLAUDE.md` §Technology Stack — `charm.land/bubbletea/v2` v2.0.4, `charm.land/lipgloss/v2` v2.x (includes `lipgloss/v2/table`), `charm.land/bubbles/v2` v2.x, `charmbracelet/x/ansi` (already direct dep from Phase 6)
- `CLAUDE.md` §Bubbletea v2 migration — `View()` returns `tea.View`; `tea.KeyPressMsg` interface
- `github.com/charmbracelet/soft-serve/pkg/ui/components/header` — reference implementation for the `overlayTitle` border-title pattern. Plan 2 extracts + cites source revision. https://github.com/charmbracelet/soft-serve/tree/main/pkg/ui/components/header
- `charm.land/lipgloss/v2` package docs — https://pkg.go.dev/charm.land/lipgloss/v2 — `NormalBorder()`, `JoinHorizontal`, `JoinVertical`, `Width`/`Height`
- `charm.land/lipgloss/v2/table` — https://pkg.go.dev/charm.land/lipgloss/v2/table — `StyleFunc(row, col int) Style` API

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/app/model.go:1403` — `bodyDims(m)` helper is the single-source-of-truth for body width/height (Phase 6). Sub-models size through this; Phase 7 does not touch call-sites.
- `internal/app/model.go:1415` — `chromeHeight(m)` stub — the entire Phase 6 → Phase 7 unlock hinges on flipping this one function body; all 17 `bodyDims(m)` call-sites automatically get the right height.
- `internal/keys/bindings.go` — every keymap has `ShortHelp() []key.Binding` and `FullHelp() [][]key.Binding` per bubbles/v2/help.KeyMap. `HintsFromBindings` reads `binding.Help()` (returns `{Key, Desc}`) — no keymap changes needed.
- `internal/ui/styles.go` — design-system layer with ~30 named styles. Phase 7 adds ~7 more (MenuKeyStyle, MenuDescStyle, LogoStyleInfo/Warn/Error, TitledBorderStyle, TitleLabelStyle) in the established `var ( … )` block pattern.
- `internal/testutil/golden.go` (Phase 6) — `RequireGoldenStructure` + `RequireGoldenColors` with `GOLDEN_UPDATE=1` regeneration friction. Phase 7 adds chrome goldens exercising this split.
- `internal/app/bench_test.go` (Phase 6) — `BenchmarkAppView` at 200×60 provides the v1.0 baseline number; Plan 3 adds `TestBenchmarkAppView_UnderBudget` (≤ 50 µs/op) gate.
- `charmbracelet/x/ansi` — already a direct dep (promoted in Phase 6); `ansi.Strip` already used by `testutil/golden.go`.

### Established Patterns
- **Pure functions for renderers** — `internal/ui/*.go` components already expose `View() string` as pure functions of sub-model state. Phase 7's `RenderLogo`, `RenderMenu`, `WrapTitled`, `overlayTitle` follow the same shape.
- **Styles as package vars** — v1.0 discipline that Phase 7 now enforces via `TestViewNoNewStyle`.
- **Explicit hex colors** — no `lipgloss.AdaptiveColor` (issue #1036); Phase 7 additions honour this.
- **Helper tests co-located** — new `chrome_test.go`, `logo_test.go`, `menu_test.go`, `hints_test.go` beside implementation, not under `testutil/`.
- **TODO-with-phase-tag** — `// TODO(phase-7): replace magic -4 with …` at `model.go:1841` — Phase 7 closes it.

### Integration Points
- `internal/app/model.go` — `AppModel.View()` (line 1284) is the single integration seam. Phase 7 rewrites its body to compose `[chrome][crumbs][titled body][statusbar]` via `lipgloss.JoinVertical`.
- `internal/app/model.go` — `chromeHeight(m)` body flips (single function). No other call-site edits.
- `internal/app/model.go` — new private `(m AppModel) menuHints() []keys.MenuHint` method owns the (state, recipientAction, IsSearchActive) dispatcher.
- `internal/app/model.go:1841` — `renderRecipientList`'s magic `m.height - 4` migrates to a `bodyDims`-based computation; the constant disappears into `WrapTitled`'s border math.
- `internal/keys/hints.go` — new file; dependency of `internal/app/model.go` and every sub-model that implements `Hints()`.
- `internal/ui/{logo,menu,chrome}.go` — new files; dependency of `internal/app/model.go`.
- `go.mod` — `charm.land/lipgloss/v2/table` may already be pulled in transitively via lipgloss v2 main; Plan 1 confirms and promotes to direct require if needed (zero version churn).
- `internal/app/resize_test.go` — Phase 6 goldens at four sizes; Phase 7 updates goldens as chrome lands (`GOLDEN_UPDATE=1` regen).

</code_context>

<specifics>
## Specific Ideas

- **Logo art direction is locked to "SOPS block + `tui` subscript"** — user chose this from a preview of four directions. Claude picks the cleanest figlet variant that fits the 6×~26 envelope; user reviews during `/gsd-verify-work`.
- **Menu is fixed 2 cols × 6 rows (12 slots)** — not responsive. Same layout at 40×12 as at 200×60. Richer responsive behaviour is explicitly out of scope for v1.1 per ROADMAP; deferred to first bug report.
- **Health title format is deliberately inconsistent with other list views** — `Health (3 findings)` (unit-ful) vs `Files (12)` / `History (47)` / `Recipients (4)` (bare). User picked the preview with this variance; it's intentional.
- **`Hints()` scope is Phase 7, not split across Phase 7+8** — contradicts SUMMARY.md's recommendation but closes the Pitfall 3 risk (wrong menu on non-wired states) in a single pass. Phase 8 scope reduces accordingly; Phase 9's discipline (golden-file matrix per tuple) is unchanged.
- **Chrome caching = pure every-frame composition** — not model-cached. Pitfall 2 framing: "cache OR zero-alloc"; we pick zero-alloc via three guardrails (styles-as-package-vars, grep-gated no-NewStyle-in-View, bench budget). Cache can be added in Phase 10 without API change if a palette pass regresses the bench.
- **`overlayTitle` research gap is closed in Plan 2** — the STATE.md pending todo "Phase 7 research pass recommended before planning" becomes a deliverable with a unit test. The soft-serve source revision is cited in a comment so Phase 11's compat sweep has a trail if the upstream pattern shifts.
- **`ARCHITECTURE.md` Pattern 3 pseudocode says `RoundedBorder`; Phase 7 uses `NormalBorder`** — intentional departure because UI-15 locks `NormalBorder` specifically. `TestChromeNormalBorderOnly` enforces this — any drift back to `Rounded` would fail CI.
- **`BenchmarkAppView ≤ 50 µs/op` enforced here, not only in Phase 11** — early enforcement catches regressions at their source; Phase 11 UI-21 is the formal sign-off.

</specifics>

<deferred>
## Deferred Ideas

### Phase 8 (already scoped)
- Header info panel (5 rows: `.sops.yaml` path / age fingerprint / recipient count / git branch+dirty / file count) — stubbed as blank 6-row placeholder in Phase 7 to preserve horizontal alignment.
- Breadcrumb chip row above body — `crumbsHeight` stays 0 in Phase 7.
- Status bar shrink to env + clipboard — breadcrumb lives in the status bar for now.

### Phase 9 (already scoped)
- Golden-file matrix per `(state, recipientAction, IsSearchActive)` tuple with hint-vs-keymap drift assertion — Phase 7 ships one minimal menu golden per state; Phase 9 adds the full matrix.
- `Hints() []MenuHint` formalised as a lint/test contract — Phase 7 ships the interface and per-sub-model implementations; Phase 9 adds the discipline layer.

### Phase 10 (already scoped)
- Logo severity coupling (`LogoStatus` parameter drives recoloring from env + flash + aggregate health) — UI-03.
- k9s-tuned palette shift toward hot-pink/purple — UI-12.
- Color-profile detection + 16-color fallback palette — UI-13. SUMMARY suggested Phase 7; ROADMAP wins.
- Redundant shape/text encoding for color-coded states (`[I]`/`[W]`/`[E]` prefixes, inverted active chip) — UI-14.
- Narrow-terminal rendering safety (40×12 through 200×60 without layout corruption) — UI-16.

### Phase 11
- Alt-screen fill-frame on enter, blank frame on exit — Pitfall 10.
- Terminal compat sweep (Alacritty, Ghostty, macOS Terminal, iTerm2, Windows Terminal, WSL2, VSCode, tmux-nested).
- Full "Looks Done But Isn't" sign-off.

### v2 (milestone-deferred per ROADMAP)
- User skin YAML loader (THM-01).
- Embedded builtin skins — dracula / gruvbox / monokai (THM-02).
- fsnotify-based skin hot-reload (THM-03).
- Live `:skin <name>` runtime switcher.
- Menu second-page overflow (Alt+M) — unnecessary while `MenuHint.Visible` suppresses cleanly.

### Not a phase concern
- Polling goroutines, `:` command bar, log forwarding, image scanner, port-forward manager, alias hot-reload, mouse interactions, animated/gradient logos — explicitly out of scope for v1.1 per ROADMAP. Reject in plan review if they resurface.

</deferred>

---

*Phase: 07-chrome-skeleton*
*Context gathered: 2026-04-24*
