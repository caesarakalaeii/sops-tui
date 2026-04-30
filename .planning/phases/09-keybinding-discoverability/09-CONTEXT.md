# Phase 9: Keybinding Discoverability - Context

**Gathered:** 2026-04-30
**Status:** Ready for planning

<domain>
## Phase Boundary

Close the SC5 contract: changing a `key.Binding` value automatically updates the persistent menu — no second "also update the menu text" edit is ever required. Phase 7 shipped the architecture (`Hinter`, `HintsFromBindings`, dispatcher, `Hints()` on 8 sub-models, one minimal golden per state). Phase 9 adds the **discipline + drift defense** layers on top of that architecture.

**In scope (this phase):**
- Refactor every `Hints()` implementation to derive from a typed `KeyMap.ShortHelp()`. Today only FileList + Detail derive; Help/Diff/Health/History/Metadata/RecipientForm return literal hint lists.
- Extract 11 new keymap types into `internal/keys/bindings.go`:
  - 6 sub-model keymaps: `HelpKeyMap`, `DiffKeyMap`, `HealthKeyMap`, `HistoryKeyMap`, `MetadataKeyMap`, `RecipientFormKeyMap`
  - 5 stateless-state keymaps (replace inline package vars in `keys/hints.go`): `FileListSearchKeyMap`, `RecipientConfirmKeyMap`, `BulkReKeyConfirmKeyMap`, `RecipientListKeyMap`, `FormatMenuKeyMap`
  - Each implements `help.KeyMap` (`ShortHelp` + `FullHelp`) so Phase 11 can also re-derive the `?` overlay if it ever wants to.
- Drift detector test in `internal/app/menuhints_drift_test.go` — runtime equality `Hints() == HintsFromBindings(km.ShortHelp())` per sub-model, modulo `Visible` toggles.
- 13-entry golden matrix at `internal/app/testdata/menu_<state>.golden` (12 states + `menu_stateFileList_search.golden`); `RequireGoldenStructure` only (ANSI-stripped) so Phase 10's palette pass doesn't churn them.
- Move FileList's `g`/`G` micro-keys into `FileListKeyMap.ShortHelp()` so derivation is total — no manual append in `FileList.Hints()`.
- Formalize the `Visible=false` override pattern: every keymap binding must appear in `Hints()` either `Visible=true` or explicitly `Visible=false`; drift detector enforces no-silent-disappearance.
- Drop inline package vars from `internal/keys/hints.go` (`FileListSearchHints`, `RecipientConfirmHints`, `BulkReKeyConfirmHints`, `RecipientListHints`, `FormatMenuHints`); their data moves into the new keymap types.
- Amend Phase 7 D-10: `menuHints()` dispatcher uses `(state, IsSearchActive)`, not `(state, recipientAction, IsSearchActive)`. The `recipientAction` axis was never wired (current code dispatches on state alone for confirm states); Phase 9 confirms this design simplification and removes the parameter from the documented contract.

**Out of scope (deferred per ROADMAP / explicit decisions):**
- New keybindings (`Phase 9 is discoverability, not surface-area expansion`).
- AST walker version of the drift detector (rejected in favour of runtime equality test — lighter weight; reconsider if runtime test misses cases).
- `?` full-screen help overlay refactor to also derive from `FullHelp()` — UI-11 retains the overlay as the complete reference; the persistent menu is what Phase 9 hardens. Help overlay derive-from-keymap would be a Phase 10/11 cleanup if it ever surfaces.
- Logo severity coupling (Phase 10, UI-03).
- 16-color fallback / palette tune (Phase 10, UI-12/UI-13).
- Narrow-terminal aesthetics matrix (Phase 10, UI-16).
- v1.0 functional regression sweep (Phase 11, UI-20).
- BenchmarkAppView budget tightening (Phase 11, UI-21 — current 5 ms with 56% headroom stays).
- Wiring `recipientAction` as a real dispatcher axis (rejected — current `(state, IsSearchActive)` is sufficient because confirm states already have separate sessionState values).

</domain>

<decisions>
## Implementation Decisions

### Hints() single-source-of-truth scope (UI-09, SC1, SC5)

- **D-301: Total derivation.** Every `Hints()` implementation derives from a `key.Binding`-backed keymap via `keys.HintsFromBindings(km.ShortHelp())`. Zero literal `MenuHint` structs survive in `internal/ui/*.go` or `internal/keys/hints.go` after Phase 9. The keymap is the single source of truth for menu content. Closes SC5: editing a binding's `Help()` value automatically propagates to the rendered menu — no second edit.
- **D-302: All keymaps centralized in `internal/keys/bindings.go`.** Same pattern as Phase 1's `FileListKeyMap` and `DetailKeyMap`. Each new keymap implements `bubbles/v2/help.KeyMap` (`ShortHelp() []key.Binding` + `FullHelp() [][]key.Binding`). 11 new types: 6 sub-model keymaps (Help/Diff/Health/History/Metadata/RecipientForm) + 5 stateless-state keymaps (FileListSearch/RecipientConfirm/BulkReKeyConfirm/RecipientList/FormatMenu).
- **D-303: `Visible=false` is a formal contract.** Every binding in a keymap must appear in `Hints()` either `Visible=true` (renders in the persistent menu) or explicitly `Visible=false` (suppressed from the persistent menu but discoverable in `?` overlay per UI-11). The drift detector test enforces this — a binding that silently disappears from the menu fails CI. Detail's `Blame` binding is the canonical example today; the contract makes the pattern reusable for any sub-model whose `ShortHelp()` exceeds the 12-slot menu cap.
- **D-304: FileList's `g`/`G` move into `FileListKeyMap.ShortHelp()`.** `Hints()` reduces to `keys.HintsFromBindings(m.keys.ShortHelp())` with no manual append. Side effect: `g`/`G` show up in the collapsed `?` help footer too — acceptable, they were always there in `FullHelp()`. This makes derivation total — no per-sub-model append idiom for the planner to remember.

### Drift defense layer (UI-09, UI-10, SC1, SC2, SC5)

- **D-305: Runtime equality drift detector.** New test asserts `model.Hints()` equals `keys.HintsFromBindings(km.ShortHelp())` for every sub-model with a keymap, applying the keymap's declared `Visible` overrides. Lighter than the BFS AST walker pattern from Phase 7.1 Plan 04 — no `go/ast` work; just construct each sub-model with its default keymap, call both methods, compare slices. Catches: literal hints that drift from bindings, forgotten `Visible=false` toggles, mnemonic typos.
- **D-306: Drift detector lives at `internal/app/menuhints_drift_test.go`.** Co-located with the `menuHints()` dispatcher so the test scans the dispatcher's switch arms from the same vantage point. 13-state coverage matches the dispatcher's branches. Avoids the `internal/keys` ↔ `internal/ui` import cycle that a `keys/drift_test.go` would create (sub-models import `keys`, not the reverse).
- **D-307: `Visible=false` semantics in the equality check.** `HintsFromBindings(km.ShortHelp())` returns all bindings with `Visible=true` by default (Phase 7 D-08 contract). The drift detector applies a per-keymap visibility map — keymaps may override visibility via a method like `func (k DetailKeyMap) HiddenFromMenu() []key.Binding { return []key.Binding{k.Blame} }` (final shape Plan 1's discretion). The test then compares `model.Hints()` to the visibility-adjusted `HintsFromBindings` output.

### Golden matrix coverage (UI-10, SC3)

- **D-308: 13-entry matrix on `(state, IsSearchActive)`.** One golden per `sessionState` value — 12 states (`stateFileList`, `stateDetail`, `stateMetadata`, `stateDiff`, `stateHelp`, `stateHistory`, `stateHealth`, `stateRecipientList`, `stateRecipientForm`, `stateFormatMenu`, `stateRecipientConfirm`, `stateBulkReKeyConfirm`) — plus one `menu_stateFileList_search.golden` for the search-active override. Dispatcher honestly only branches on `(state, IsSearchActive)` (Phase 7 D-10's `recipientAction` was never wired); the matrix reflects reality.
- **D-309: `recipientAction` parameter removed from D-10 design.** No code change to `menuHints()` (it never took `recipientAction`); this is a documentation amendment recorded in `09-01-SUMMARY.md` (or wherever Plan 2 lands the dispatcher cleanup). Phase 7 `07-CONTEXT.md` D-10 stays intact as historical record; Phase 9 supersedes via this entry. Closes a piece of architectural drift carried forward from Phase 7.
- **D-310: Each golden captures `RenderMenu` output only — not the full chrome strip.** `testdata/menu_state<Name>.golden` files contain the rendered persistent menu (2 cols × 6 rows = 12 hint slots) for that state. Excludes info panel, logo, body, status bar — those have their own goldens (Phase 7 chrome goldens, Phase 8 resize goldens). Smallest blast radius: Phase 10's palette pass doesn't churn these.
- **D-311: Structure-only goldens via `RequireGoldenStructure`.** ANSI-stripped per state. Color is exercised by Phase 7's `TestRenderChrome_*` color-presence assertions; Phase 9's matrix verifies content + layout. Stable across Phase 10's `MenuKeyStyle`/`MenuDescStyle` palette tune.

### Inline package-var hint sets (UI-09, UI-10)

- **D-312: All five inline hint sets convert to keymap structs.** `internal/keys/hints.go` after Phase 9: keeps `MenuHint`, `Hinter`, `HintsFromBindings`. Loses `FileListSearchHints`, `RecipientConfirmHints`, `BulkReKeyConfirmHints`, `RecipientListHints`, `FormatMenuHints`. New keymap types in `bindings.go`: `FileListSearchKeyMap`, `RecipientConfirmKeyMap`, `BulkReKeyConfirmKeyMap`, `RecipientListKeyMap`, `FormatMenuKeyMap` — each with `ShortHelp()` returning the bindings that today are inline literals. Default instances follow the `DefaultXxxKeyMap` naming convention.
- **D-313: Quit-suppression via `Visible=false` in confirm keymaps.** `RecipientConfirmKeyMap` and `BulkReKeyConfirmKeyMap` embed `GlobalKeyMap` (so the `Quit` binding still exists for any future state coupling) but their visibility map marks `Quit` as `Visible=false`. Honors the D-303 contract: every binding either visible or explicitly hidden. The doc-comment from `keys/hints.go:65-73` (the UI-SPEC §confirm-flow-quit-suppression citation) migrates to the new keymap definition's doc block. AppModel.Update()'s actual quit-suppression code path stays unchanged — this is purely about menu rendering.
- **D-314: Plan split — primitive-first.** Plan 1 lands the keymap extraction (the structural refactor); Plan 2 lands the discipline layer (drift detector + golden matrix) on top. Same shape as Phase 7's primitive-first split (D-25). Plan 2 doesn't kick in until Plan 1's keymaps exist — testing a contract that doesn't yet have data is busywork.

### Plan Split (2-plan ROADMAP budget)

- **D-315: Two plans, primitive-first matching Phase 7's split:**
  - **Plan 1 — Keymap extraction + Hints() derivation:**
    - 11 new keymap types in `internal/keys/bindings.go` (6 sub-model + 5 stateless-state) with `Default*KeyMap` instances; each implements `help.KeyMap` (`ShortHelp` + `FullHelp`) and declares its `Visible=false` overrides.
    - Refactor `internal/ui/{help,diff,health,history,metadata,recipientform}.go` `Hints()` to `keys.HintsFromBindings(m.keys.ShortHelp())` minus visibility overrides; each sub-model gains a `keys` field of the appropriate keymap type.
    - Move FileList's `g`/`G` into `FileListKeyMap.ShortHelp()`; drop the manual append from `FileList.Hints()`.
    - Delete the 5 inline hint-set vars from `internal/keys/hints.go`; update `menuHints()` switch arms to use `keys.HintsFromBindings(km.ShortHelp())` for the stateless states.
    - Existing per-sub-model hint tests stay green (the literal expected values now come from the keymap).
    - Zero changes to behaviour or visual output — pure refactor.
  - **Plan 2 — Drift detector + golden matrix + recipientAction cleanup:**
    - `internal/app/menuhints_drift_test.go` — runtime equality test per (state, IsSearchActive) tuple.
    - 13 menu goldens at `internal/app/testdata/menu_<state>.golden` (`RequireGoldenStructure`).
    - Documentation amendment recording the `recipientAction` parameter removal from D-10 design (referenced from `09-02-SUMMARY.md` per D-309).
    - Drift detector lockdown: Plan 2 is also the place where any test-only helpers (visibility map accessor on each keymap) get their final shape.
- **D-316: Plan 2 is the smaller plan.** Most weight is in Plan 1 (11 new keymaps + 6 sub-model refactors). Plan 2 is the test-infrastructure crown that would be unbuildable without Plan 1. Splitting differently would re-open the keymap definitions multiple times.

### Folded Todos

None — the prior STATE.md "Pending Todos" relevant to chrome (Phase 6 D-15 manual UAT, Phase 10 research) belong to other phases.

### Claude's Discretion

- Exact `key.Binding` values for the 6 sub-model keymaps (Help/Diff/Health/History/Metadata/RecipientForm) — Plan 1 author scouts each sub-model's `Update()` body, identifies every `key.Match(...)` clause, and synthesizes a binding. Default keymap field names follow the action verb (`Quit`, `Scroll`, `Confirm`, `Cancel`); `WithHelp` strings match each model's existing literal hint description exactly so no goldens churn unexpectedly.
- Whether the `Visible=false` override surface is a method on each KeyMap (`HiddenFromMenu()`) or a struct field on each binding via `key.WithHelp("…", "")` (empty desc convention) — Plan 1 author picks. Recommendation: a method, because the convention is per-state policy, not per-binding metadata; `key.Binding` should stay portable across states (e.g., `Quit` is universally bound but visibly suppressed only in confirm states).
- Whether to move the doc-comment from `keys/hints.go:65-73` (RecipientConfirmHints quit-suppression rationale) into one keymap-level block or split across `RecipientConfirmKeyMap` + `BulkReKeyConfirmKeyMap` — Plan 1 author picks. Recommendation: a shared package-level block at the top of `bindings.go` that both keymaps reference.
- Golden file naming convention — `menu_stateFileList.golden` vs `menu_filelist.golden` vs `menu_FileList.golden`. Plan 2 author picks; recommendation: lowercase + underscore (`menu_file_list.golden`, `menu_file_list_search.golden`) for filesystem-friendliness.
- Whether to delete the now-unused `recipientAction` parameter mentions from `internal/app/model.go` comments + chrome.go signature (if any) — Plan 2 author scouts and cleans up. Most likely a no-op since the parameter never made it into code.
- Test construction strategy for the drift detector — whether each sub-model is constructed via its real constructor (`NewHelpModel(...)`) or zero-value with manual keymap injection. Plan 2 author picks; recommendation: real constructor for fidelity to production code paths.
- BFS AST walker reuse — Phase 7.1 Plan 04's `TestViewNoNewStyle` walker is the existing pattern for reachability tests. Phase 9's drift detector is runtime equality (no AST walking required), so the walker isn't reused. If the runtime equality approach proves insufficient (e.g., a sub-model is hard to construct in tests), the AST walker is the documented fallback.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project decision docs
- `.planning/ROADMAP.md` §"Phase 9: Keybinding Discoverability" — Goal, 5 success criteria, 2-plan budget, UI-09/UI-10/UI-11 requirements
- `.planning/ROADMAP.md` §"Milestone v1.1 — Explicitly Out of Scope for v1.1" — reject list (polling goroutines, `:` command bar, skin loader, etc.)
- `.planning/REQUIREMENTS.md` §"Keybinding Discoverability" — UI-09 (Hints from ShortHelp), UI-10 (per-state re-hydration), UI-11 (`?` overlay retained as complete reference)
- `.planning/REQUIREMENTS.md` §"Theming & Accessibility" — UI-15 (ASCII-only chrome — applies to menu cells), UI-21 (no `lipgloss.NewStyle()` reachable from `View()` — applies to refactored sub-models)
- `.planning/PROJECT.md` — k9s visual parity is a hard product attribute (per user-memory `feedback_k9s_visual_parity.md`)

### Prior phase decisions (carried forward)
- `.planning/phases/07-chrome-skeleton/07-CONTEXT.md` §decisions D-08 (`MenuHint`/`Hinter`/`HintsFromBindings` interface), D-09 (8 sub-models implement `Hints()`), D-10 (dispatcher contract — Phase 9 D-309 amends this to drop `recipientAction`), D-11 (search-active override), D-22 (TestViewNoNewStyle BFS walker — pattern reference, not reused)
- `.planning/phases/07-chrome-skeleton/07.1-04-PLAN.md` — BFS AST reachability walker reference implementation (pattern only; Phase 9 uses runtime equality not AST walking)
- `.planning/phases/08-header-info-panel/08-CONTEXT.md` §domain — chrome integration baseline (info panel, crumbs, status-bar shrink — done before Phase 9 starts)
- `.planning/phases/06-layout-groundwork/06-CONTEXT.md` §decisions — `RequireGoldenStructure` / `RequireGoldenColors` split (D-311 inherits the structure-only path)

### Existing implementation (Phase 9 refactors / extends)
- `internal/keys/hints.go` — `MenuHint`, `Hinter`, `HintsFromBindings`; 5 inline hint-set vars (`FileListSearchHints`, `RecipientConfirmHints`, `BulkReKeyConfirmHints`, `RecipientListHints`, `FormatMenuHints`) — these vars are REMOVED in Plan 1 per D-312
- `internal/keys/bindings.go` — `GlobalKeyMap`, `FileListKeyMap`, `DetailKeyMap`; pattern for the 11 new keymaps (D-302); `FileListKeyMap.ShortHelp()` updated in Plan 1 to include `g`/`G` per D-304
- `internal/ui/filelist.go:383` — `FileListModel.Hints()` — Plan 1 simplifies the manual `g`/`G` append per D-304
- `internal/ui/detail.go:820` — `DetailModel.Hints()` — Plan 1 keeps the `Blame` `Visible=false` override but moves the visibility map into the keymap per D-307
- `internal/ui/help.go:96`, `diff.go:176`, `health.go:181`, `history.go:133`, `metadata.go:167`, `recipientform.go:163` — six `Hints()` implementations refactored to derive from new keymaps (D-301)
- `internal/app/model.go:1498` — `menuHints()` dispatcher — Plan 1 updates the 5 stateless-state arms to use new keymaps; Plan 2 documents the `recipientAction` parameter removal (D-309)
- `internal/testutil/golden.go` — `RequireGoldenStructure` (Plan 2 uses) + `GOLDEN_UPDATE=1` regeneration friction
- `internal/ui/menu.go` — `RenderMenu` — target output of the 13 menu goldens (D-310)
- `internal/app/view_no_newstyle_test.go` — Phase 7.1 BFS walker reference (D-305 chose runtime equality over BFS — note documented for future maintainers)

### Technology / external references
- `charm.land/bubbles/v2/key.Binding.Help() returns key.Binding.Help{Key, Desc}` — what `HintsFromBindings` consumes (Phase 7 D-08)
- `charm.land/bubbles/v2/help.KeyMap` interface — `ShortHelp() []key.Binding` + `FullHelp() [][]key.Binding`; every Phase 9 keymap implements this
- `charm.land/lipgloss/v2` — `RequireGoldenStructure` ANSI stripping via `charmbracelet/x/ansi`

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/keys/hints.go` `MenuHint` struct + `Hinter` interface + `HintsFromBindings(bindings []key.Binding) []MenuHint` — Phase 9 builds on these without changing any signature.
- `internal/keys/bindings.go` `FileListKeyMap` + `DetailKeyMap` — canonical shape for all 11 new keymap types (struct embedding `GlobalKeyMap`, individual `key.Binding` fields, `ShortHelp()` + `FullHelp()` methods, `Default*KeyMap` instance).
- `internal/app/model.go:1498` `menuHints()` dispatcher — switch arms on `m.state` + `m.fileList.IsSearchActive()`. Plan 1 updates the 5 stateless-state arms to call `keys.HintsFromBindings(km.ShortHelp())`.
- `internal/testutil/golden.go` `RequireGoldenStructure` — ANSI-stripped golden comparator with `GOLDEN_UPDATE=1` regen. Phase 9's 13 menu goldens use this.
- `internal/ui/menu.go` `RenderMenu(hints []keys.MenuHint, width int) string` — the renderer that produces each golden's body content.

### Established Patterns
- **Keymaps centralised in `internal/keys/bindings.go`** — every keymap there, embeds `GlobalKeyMap`, implements `help.KeyMap`. Plan 1 follows this pattern for all 11 new types.
- **`Default*KeyMap` instance pattern** — `var DefaultXxxKeyMap = XxxKeyMap{...}` declares the canonical default. Sub-models accept a `keys XxxKeyMap` field with constructors taking the default.
- **Doc-comment cited UI-SPEC reference** — Phase 7's `keys/hints.go:65-73` cites UI-SPEC §confirm-flow-quit-suppression. Plan 1 migrates this comment to the new keymap definition (D-313).
- **Runtime equality tests over AST walkers** — D-305 picks runtime equality. AST walkers (Phase 7.1 Plan 04) remain available for reachability questions but aren't needed here.
- **Goldens via `RequireGoldenStructure` (ANSI-stripped) by default** — Phase 6/7 precedent. Color-presence is a separate test family. D-311 picks structure-only.

### Integration Points
- `internal/keys/bindings.go` — 11 new keymap type definitions + 11 `Default*KeyMap` instances + 11 `ShortHelp()`/`FullHelp()` method pairs. The single biggest landing of Plan 1.
- `internal/keys/hints.go` — 5 inline hint-set vars deleted; file shrinks to ~60 lines (interface + helper only).
- `internal/ui/{help,diff,health,history,metadata,recipientform}.go` — each gains a `keys *KeyMap` field (or value); each `Hints()` method shrinks to one line. The struct's existing constructor signature changes: `NewHelpModel(...)` etc. accept the keymap (or default-construct it).
- `internal/ui/filelist.go` — `FileListModel.Hints()` simplifies; the `m.keys` field already exists.
- `internal/ui/detail.go` — `DetailModel.Hints()` simplifies; the `Visible=false` override moves to a `HiddenFromMenu()` method on `DetailKeyMap` (or equivalent).
- `internal/app/model.go menuHints()` — switch arms for stateless states (`stateRecipientConfirm`, `stateBulkReKeyConfirm`, `stateRecipientList`, `stateFormatMenu`, search-active override) call `keys.HintsFromBindings(km.ShortHelp())` against the new keymaps.
- `internal/app/menuhints_drift_test.go` — new file, Plan 2. Drift detector test + 13-entry golden matrix.
- `internal/app/testdata/menu_*.golden` — 13 new golden files, Plan 2.
- All existing `_test.go` files in `internal/ui/` — Plan 1 verifies the per-sub-model hint tests still pass with the keymap-derived output (literal expectations match because `WithHelp` strings preserve the existing description text).
- `go.mod` — no new dependencies. `charm.land/bubbles/v2` is already pulled in transitively via Phase 1.

</code_context>

<specifics>
## Specific Ideas

- **Total derivation, no carve-outs.** The user picked the strongest SC5 stance: every Hints() output traces back to a keymap. No "stateless-state inline literal is the source of truth" carve-out. This drives the 11-keymap extraction in Plan 1.
- **Visible=false IS the discipline mechanism, not a hack.** Detail's Blame visibility override formalizes as the canonical idiom. Future sub-models with >12 bindings use the same pattern — bind it, hide it, document it.
- **Drift detection happens at runtime, not via AST walking.** Phase 7.1 Plan 04's BFS walker remains the precedent for reachability tests; Phase 9's drift question is "do these two slices match?" — best answered by calling both methods and comparing, not by walking source.
- **Golden matrix is menu-only.** The cross product is `(state × IsSearchActive)` = 13 cases. Each golden captures the 2×6 menu output via `RenderMenu`, not the full chrome. Phase 10's palette pass won't churn structure-only goldens.
- **`recipientAction` was always vestigial.** Phase 7 D-10 specced it; the code never wired it. Phase 9 D-309 confirms the simplification — `(state, IsSearchActive)` is the real dispatcher signature, in code and contract.
- **Plan 2 is the smaller plan.** All architectural weight lives in Plan 1 (11 keymaps + 6 sub-model refactors); Plan 2 is the test crown. Splitting differently would re-open the keymap surface multiple times.

</specifics>

<deferred>
## Deferred Ideas

### Phase 10 (already scoped)
- Logo severity coupling — UI-03.
- k9s-tuned palette tune (accent → hot-pink/purple) — UI-12.
- 16-color fallback palette — UI-13.
- Redundant shape/text encoding for color-coded states — UI-14.
- Narrow-terminal aesthetics matrix — UI-16.

### Phase 11 (already scoped)
- v1.0 functional regression test pass — UI-20.
- BenchmarkAppView budget tightening (current 5 ms with 56% headroom; D-18 caching fallback if needed) — UI-21.
- Terminal compat sweep + alt-screen cleanup.

### Possibly Phase 10/11 cleanup, possibly v2
- `?` full-screen help overlay refactor to derive from `FullHelp()` — currently the overlay renders its own structure. If the keymap-as-source-of-truth contract should extend there too, it's a natural Phase 11 cleanup. Phase 9 keeps the overlay as-is per UI-11 ("retained as the complete reference"), but the Phase 9 keymap centralization makes the future refactor cheap.
- AST walker variant of the drift detector — if the runtime equality test ever proves insufficient (e.g., a sub-model is hard to construct in tests, or visibility logic gets too intricate), the BFS walker pattern from Phase 7.1 Plan 04 is the documented fallback. Tracking as deferred so it doesn't get re-discovered.

### Not a phase concern
- Mouse interactions, animated/gradient logos, polling goroutines, alias hot-reload, etc. — explicitly out of scope for v1.1 per ROADMAP. Reject in plan review if they resurface.

</deferred>

---

*Phase: 09-keybinding-discoverability*
*Context gathered: 2026-04-30*
