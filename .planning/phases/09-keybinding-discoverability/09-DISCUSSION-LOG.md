# Phase 9: Keybinding Discoverability - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-30
**Phase:** 09-keybinding-discoverability
**Areas discussed:** Hints() single-source-of-truth scope, Drift defense layer, Golden matrix coverage, Inline package-var hint sets

---

## Hints() single-source-of-truth scope

### Q1: Workflow when a binding's mnemonic/description changes

| Option | Description | Selected |
|--------|-------------|----------|
| Edit keymap; menu auto-derives (Recommended) | Strongest k9s parity. Every `Hints()` derives from `key.Binding.ShortHelp()`. Editing the keymap is the only change needed. Requires extracting keymaps for the 6 sub-models that today use literal hint lists. | ✓ |
| Edit keymap + literal; CI catches desync | Pragmatic. Keep literals where they exist today (6 sub-models) but add a drift detector that fails if literal mnemonic/description doesn't match underlying `key.Binding`. No keymap extraction. | |
| Mixed: derive where keymap exists, lint elsewhere | FileList + Detail keep their derive-from-ShortHelp pattern. The 6 literal-Hints sub-models stay as-is unless they have a corresponding keymap; a separate test covers each literal against actual `key.Match` clauses in `Update()`. | |

**User's choice:** Edit keymap; menu auto-derives.
**Notes:** Highest-fidelity SC5 satisfaction. Drives the 11-keymap extraction in Plan 1.

---

### Q2: Where new sub-model keymaps live

| Option | Description | Selected |
|--------|-------------|----------|
| All in `internal/keys/bindings.go` (Recommended) | Same pattern as `FileListKeyMap` and `DetailKeyMap`. Single canonical keybinding contract file. Implements `bubbles/v2/help.KeyMap` (`ShortHelp` + `FullHelp`) so the `?` overlay can reuse `FullHelp()` later. | ✓ |
| Co-located with sub-model | Reduces import cycles, keeps sub-models self-contained. Closer to `bubbles/v2` component idiom but spreads keymap surface across 8 files. | |
| Hybrid: navigation-heavy in bindings.go, modals co-located | Diff/Health/History/Metadata in bindings.go (multiple bindings); RecipientForm + Help (1–2 bindings each) co-located. | |

**User's choice:** All in `internal/keys/bindings.go`.
**Notes:** Consistent with Phase 1's centralization pattern.

---

### Q3: `Visible=false` formal contract

| Option | Description | Selected |
|--------|-------------|----------|
| Yes — explicit `Visible` toggle, test enforces (Recommended) | Every binding in a keymap must appear in `Hints()` either `Visible=true` OR explicitly `Visible=false`. Test asserts no binding silently disappears from the menu. Detail's Blame is the canonical example. `Visible=false` bindings remain discoverable in `?` overlay (UI-11). | ✓ |
| No — sub-model author's discretion | Current state. Authors choose what makes the persistent menu cut. No test enforcement. Acceptable risk: a forgotten binding never reaches the menu but is still in `?` help. | |
| Hybrid — contract requires every ShortHelp binding visible; FullHelp-only bindings opt-in | If a binding is in `ShortHelp()` but not in `Hints()`, test fails. If a binding is in `FullHelp()` only (g/G micro-keys), no rule. Lighter contract. | |

**User's choice:** Yes — explicit `Visible` toggle, test enforces.
**Notes:** Formalizes Detail's existing pattern; future sub-models with >12 bindings use the same idiom.

---

## Drift defense layer

### Q1: What the drift detector actually verifies

| Option | Description | Selected |
|--------|-------------|----------|
| `Hints() == HintsFromBindings(ShortHelp())` per sub-model (Recommended) | For every sub-model with a keymap, assert `Hints()` is byte-equal to `keys.HintsFromBindings(m.keys.ShortHelp())` modulo `Visible` toggles. Catches: literal hints that drift, forgotten `Visible=false`. Doesn't need AST walking — runtime slice comparison. | ✓ |
| Every visible mnemonic resolves to a real `key.Match` clause in `Update()` | AST walker finds every `Hints()` Mnemonic that's `Visible=true`, then verifies a `key.Match(...)` clause for that binding exists in the same file's `Update()`. Catches: rendered menu shows a key that doesn't actually do anything. Heavier (BFS AST walk). | |
| Both — layered defense | Runtime equality catches literal drift; AST walker catches reachability gaps. Two tests, complementary failure modes. | |

**User's choice:** `Hints() == HintsFromBindings(ShortHelp())` — runtime equality.
**Notes:** Lighter than Phase 7.1 Plan 04's BFS walker. AST walker remains the documented fallback if the runtime test ever proves insufficient.

---

### Q2: Drift detector test home

| Option | Description | Selected |
|--------|-------------|----------|
| `internal/app/menuhints_drift_test.go` (Recommended) | Co-located with the dispatcher in `internal/app`. Single test file scans across all sub-models from the AppModel side — same vantage point as `menuHints()`. | ✓ |
| Per sub-model (`internal/ui/<view>_test.go`) | Each sub-model owns its own drift test. Closer to component-test idiom. 8 separate tests, slight duplication. | |
| `internal/keys/drift_test.go` | Test in keys package, imports from `internal/ui`. Inverts dependency direction; likely creates an import cycle. | |

**User's choice:** `internal/app/menuhints_drift_test.go`.
**Notes:** Avoids the `internal/keys` ↔ `internal/ui` import cycle.

---

### Q3: FileList's `g`/`G` manual append rule

| Option | Description | Selected |
|--------|-------------|----------|
| Move `g`/`G` into `ShortHelp()` (Recommended) | Total derivation — `Hints() == HintsFromBindings(ShortHelp())` with no manual append. Affects `help.KeyMap` implementation (`g`/`G` show in collapsed help footer too). | ✓ |
| Keep append; drift test allows post-derivation additions | Drift test compares Hints() output to `HintsFromBindings(ShortHelp())` PLUS an explicit allowlist per sub-model. Preserves "micro-keys for navigation" idiom. | |
| Drop `g`/`G` from `Hints()` entirely; keep in `?` overlay only | `g`/`G` become FullHelp-only. Persistent menu shows j/k arrow nav; `g`/`G` are vim power-user knowledge. | |

**User's choice:** Move `g`/`G` into `ShortHelp()`.
**Notes:** Total derivation; no per-sub-model append idiom for the planner to remember.

---

## Golden matrix coverage

### Q1: Matrix scope

| Option | Description | Selected |
|--------|-------------|----------|
| 13 goldens — `(state × IsSearchActive)` only (Recommended) | One golden per `sessionState` + one for stateFileList-with-search-active. Matches what the dispatcher actually does today. `recipientAction` stays unused; remove the parameter from D-10 design. | ✓ |
| 13 goldens + drop unused `recipientAction` parameter | Same 13 goldens, but actively delete the `recipientAction` from D-10. Update dispatcher signature, canonical refs, amend Phase 7 D-10. | |
| Wire `recipientAction` now — ~16 goldens | Honor D-10 fully: `(state, recipientAction, IsSearchActive)`. recipientAction has 2-3 distinct values. Larger matrix. | |

**User's choice:** 13 goldens — `(state × IsSearchActive)` only.
**Notes:** D-309 records the `recipientAction` parameter removal as a doc amendment.

---

### Q2: What each menu golden captures

| Option | Description | Selected |
|--------|-------------|----------|
| `RenderMenu` output only (Recommended) | Each golden is the rendered persistent menu for that (state, search) tuple — no chrome, no body. Small, focused. Lives at `testdata/menu_<state>.golden`. | ✓ |
| Full chrome strip (info panel + menu + logo) | Each golden is the full top-of-screen chrome for that state. Bigger goldens; tests more interaction. Higher refresh cost when Phase 10's palette pass lands. | |
| Full View() output at 200x60 | Each golden is the entire screen render at 200x60 for that state. Maximum coverage, maximum brittleness. 13 more goldens beyond Phase 7's 4. | |

**User's choice:** `RenderMenu` output only.
**Notes:** Smallest blast radius for Phase 10's palette pass.

---

### Q3: Structure vs color split

| Option | Description | Selected |
|--------|-------------|----------|
| Structure-only via `RequireGoldenStructure` (Recommended) | ANSI-stripped goldens. Color is exercised by Phase 7's existing chrome tests. Stable across Phase 10's palette pass. | ✓ |
| Both structure AND color (full Phase 6 split) | Each state gets two goldens: structure + color. Phase 10 palette pass forces color-golden refresh; structure stays stable. | |
| Color-only | Hint text already verified by drift detector test. Goldens only verify rendering layer applied colors correctly. Fragile vs palette changes. | |

**User's choice:** Structure-only via `RequireGoldenStructure`.
**Notes:** Inherits Phase 6 ANSI-strip pattern; doesn't churn under Phase 10 palette tune.

---

## Inline package-var hint sets

### Q1: Treatment of the 5 inline hint sets in `keys/hints.go`

| Option | Description | Selected |
|--------|-------------|----------|
| Each becomes a typed KeyMap struct in bindings.go (Recommended) | Create FileListSearchKeyMap, RecipientConfirmKeyMap, BulkReKeyConfirmKeyMap, RecipientListKeyMap, FormatMenuKeyMap. Total SoT — zero literal MenuHint structs anywhere. Aligns with Area 1's "all keymaps in bindings.go" decision. | ✓ |
| Keep inline; document as canonical "stateless-state" pattern | These 5 vars stay literal. Doc comment explains the carve-out. Drift detector skips them. | |
| Hybrid: confirm/list/menu states get keymaps; search-active stays inline | RecipientConfirm/BulkReKeyConfirm/RecipientList/FormatMenu — yes. FileListSearch — stay inline because search overlays FileList's own keymap. | |

**User's choice:** Each becomes a typed KeyMap struct in `bindings.go`.
**Notes:** Total derivation — zero literal MenuHint after Phase 9.

---

### Q2: Quit suppression in confirm states

| Option | Description | Selected |
|--------|-------------|----------|
| Quit binding present + `Visible=false` in confirm keymaps (Recommended) | RecipientConfirmKeyMap and BulkReKeyConfirmKeyMap include `Quit` via embedded `GlobalKeyMap`, but their visibility map marks Quit `Visible=false`. Honors D-303 explicit-toggle contract. | ✓ |
| Don't embed `GlobalKeyMap` in confirm keymaps | Confirm keymaps stand alone with only y/n/Esc/j/k bindings. Cleanest absence-by-construction. Loses "every state inherits Global" pattern. | |
| Keep `GlobalKeyMap` embed; quit handler in `Update()` ignores key in confirm states | Quit stays a real binding; dispatcher's quit handler returns nil cmd in confirm states. Hints() includes Quit `Visible=false`. | |

**User's choice:** Quit binding present + `Visible=false` in confirm keymaps.
**Notes:** AppModel.Update()'s actual quit-suppression code path stays unchanged — this is purely about menu rendering. Doc comment from `keys/hints.go:65-73` migrates to the new keymap doc block.

---

### Q3: Plan split

| Option | Description | Selected |
|--------|-------------|----------|
| Plan 1 keymap extraction; Plan 2 drift detector + golden matrix (Recommended) | Plan 1: 11 new keymaps + 6 sub-model `Hints()` refactors + g/G into ShortHelp + delete inline vars. Plan 2: drift detector test, 13-entry golden matrix, recipientAction parameter removal documentation. | ✓ |
| Plan 1 drift detector first (red), Plan 2 keymap extraction (green) | TDD-flavored. Plan 1 adds drift detector in red state + golden scaffolding. Plan 2 extracts keymaps; drift detector goes green. Slower start. | |
| Plan 1 active sub-models + matrix; Plan 2 stateless-state keymaps + cleanup | Plan 1 goes farthest visible-impact direction first (Diff/Health/History/Metadata/Help/RecipientForm + drift detector + 13-entry matrix). Plan 2 handles stateless-state keymaps + recipientAction removal. | |

**User's choice:** Plan 1 keymap extraction; Plan 2 drift detector + golden matrix.
**Notes:** Primitive-first matches Phase 7's split (D-25). Plan 2 doesn't kick in until Plan 1's keymaps exist.

---

## Claude's Discretion

- Exact `key.Binding` values for the 6 sub-model keymaps (Help/Diff/Health/History/Metadata/RecipientForm) — Plan 1 author scouts each `Update()` body for `key.Match(...)` clauses.
- `Visible=false` override surface — method on KeyMap (`HiddenFromMenu()`) vs binding-level convention.
- Doc-comment placement for confirm-flow-quit-suppression rationale — package-level vs per-keymap.
- Golden file naming (`menu_state<Name>.golden` vs `menu_<state>.golden`).
- Test construction strategy — real constructor vs zero-value.

## Deferred Ideas

- `?` full-screen help overlay refactor to derive from `FullHelp()` — natural Phase 10/11 cleanup.
- AST walker drift detector — fallback if runtime equality proves insufficient.
- Phase 7 D-10 amendment recording the `recipientAction` simplification — handled in Plan 2.
