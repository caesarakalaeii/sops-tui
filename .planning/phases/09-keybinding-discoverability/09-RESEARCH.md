# Phase 9: Keybinding Discoverability - Research

**Researched:** 2026-04-30
**Domain:** Go TUI keymap refactor — bubbles/v2 `help.KeyMap` interface, runtime drift detection, golden-file discipline
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- D-301: Total derivation. Every `Hints()` derives from `keys.HintsFromBindings(km.ShortHelp())`. Zero literal `MenuHint` structs survive in `internal/ui/*.go` or `internal/keys/hints.go`.
- D-302: All keymaps centralized in `internal/keys/bindings.go`. Each implements `help.KeyMap` (`ShortHelp() []key.Binding` + `FullHelp() [][]key.Binding`). 11 new types.
- D-303: `Visible=false` is a formal contract. Every binding appears in `Hints()` either visible or explicitly hidden. Drift detector enforces no-silent-disappearance.
- D-304: FileList `g`/`G` move into `FileListKeyMap.ShortHelp()`. `Hints()` reduces to `keys.HintsFromBindings(m.keys.ShortHelp())` with no manual append.
- D-305: Runtime equality drift detector only. No AST walker. BFS walker is the documented fallback if runtime equality proves insufficient.
- D-306: Drift detector lives at `internal/app/menuhints_drift_test.go`.
- D-307: `Visible=false` semantics via a method on each keymap (e.g., `HiddenFromMenu() []key.Binding`). Final shape is Plan 1 author's discretion.
- D-308: 13-entry golden matrix on `(state, IsSearchActive)`. One golden per `sessionState` + one for search-active override.
- D-309: `recipientAction` parameter removed from D-10 design. No code change to `menuHints()`. Documentation amendment only.
- D-310: Each golden captures `RenderMenu` output only, not full chrome strip.
- D-311: Structure-only goldens via `RequireGoldenStructure`. ANSI-stripped. Stable across Phase 10 palette pass.
- D-312: All five inline hint-set vars convert to keymap structs. `hints.go` retains only `MenuHint`, `Hinter`, `HintsFromBindings`.
- D-313: Quit-suppression via `Visible=false` in confirm keymaps. `RecipientConfirmKeyMap` and `BulkReKeyConfirmKeyMap` embed `GlobalKeyMap` but mark `Quit` as `Visible=false`.
- D-314: Plan split primitive-first. Plan 1 = keymap extraction + Hints() derivation. Plan 2 = drift detector + golden matrix + D-309 documentation.
- D-315: Two plans. D-316: Plan 2 is the smaller plan.

### Claude's Discretion

- Exact `key.Binding` values for 6 sub-model keymaps: Plan 1 author scouts each `Update()` body for `key.Match` / `kMsg.String()` clauses.
- Whether `Visible=false` override surface is a method (`HiddenFromMenu()`) or a struct field convention. Recommendation: a method, because visibility is per-state policy not per-binding metadata.
- Whether quit-suppression doc-comment from `keys/hints.go:65-73` goes in one shared block or split across both confirm keymaps. Recommendation: shared package-level block at the top of `bindings.go`.
- Golden file naming convention. Recommendation: lowercase + underscore (`menu_file_list.golden`, `menu_file_list_search.golden`).
- Whether to delete `recipientAction` mentions from `model.go` comments + any chrome.go signature. Plan 2 scouts and cleans up.
- Drift detector construction strategy: real constructors vs zero-value with manual injection. Recommendation: real constructors for production code fidelity.
- BFS AST walker reuse: not needed, documented fallback only.

### Deferred Ideas (OUT OF SCOPE)

- New keybindings of any kind.
- AST walker version of drift detector (documented fallback only).
- `?` overlay refactor to derive from `FullHelp()`.
- Logo severity, palette, 16-color fallback, narrow-terminal aesthetics (Phase 10).
- v1.0 regression sweep, BenchmarkAppView budget tightening (Phase 11).
- Wiring `recipientAction` as a real dispatcher axis (rejected; current `(state, IsSearchActive)` is correct).

</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| UI-09 | Every interactive sub-model exposes `Hints() []keys.MenuHint` derived from `key.Binding.ShortHelp()` — keymap is single source of truth | D-301/D-302: 11 new keymap types + D-304 g/G move + D-312 inline var deletion fully address this |
| UI-10 | Persistent menu re-hydrates from active sub-model's `Hints()` on every `View()` call; modal states show modal keybindings | D-308 + D-305/D-306: 13-entry golden matrix + drift detector prove per-state correctness |
| UI-11 | `?` full-screen help overlay retained as complete reference | No code change required; overlay already works. Phase 9 only adds `FullHelp()` on new keymaps so Phase 10/11 can optionally derive from them later |

</phase_requirements>

---

## Summary

Phase 9 is a no-new-feature structural refactor plus a discipline-enforcement layer on the chrome architecture shipped in Phase 7. The codebase already has the scaffolding: `HintsFromBindings`, the `Hinter` interface, `FileListKeyMap`, `DetailKeyMap`, and working `Hints()` on all 8 sub-models. The problem is that 6 of those 8 `Hints()` implementations return literal `[]keys.MenuHint` slices rather than deriving from a keymap — so editing a binding description requires two edits (binding + hint slice). The 5 stateless-state hint sets in `keys/hints.go` are also literal slices with no backing keymap at all.

Plan 1 closes this by extracting 11 new keymap types into `internal/keys/bindings.go`, moving the literal hint data into `WithHelp()` strings on those keymaps, refactoring the 6 sub-model `Hints()` methods to one-liners via `HintsFromBindings(m.keys.ShortHelp())`, and deleting the 5 inline package vars. Plan 2 crowns the refactor with a runtime equality drift detector (`menuhints_drift_test.go`) plus a 13-entry golden matrix that locks down what the persistent menu renders in every app state.

The key technical insight: all 4 models that currently lack a keymap (Health, History, Metadata, RecipientForm) handle scrolling through AppModel-driven `ScrollDown()`/`ScrollUp()` calls with no `Update()` method of their own. Their keymaps will be data-only (no `key.Match` logic needed in the sub-model) and their `Hints()` will simply call `HintsFromBindings(m.keys.ShortHelp())`. Diff and RecipientForm do have `Update()` methods, so their new keymaps must encode the exact key strings those methods already match against (e.g., `"y"`, `"n"`, `"esc"`, `"enter"`).

**Primary recommendation:** Follow the plan split and file locations exactly as specified in CONTEXT.md. Do not deviate from `Default*KeyMap` naming, `help.KeyMap` interface implementation, or the `RequireGoldenStructure` approach established in Phase 6. The only true discretion areas are the `HiddenFromMenu()` method shape and the golden file naming convention.

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Keymap definitions (11 new types) | `internal/keys/bindings.go` | -- | All keymaps centralized per D-302; same tier as FileListKeyMap/DetailKeyMap |
| `Hints()` derivation on sub-models | `internal/ui/*.go` (6 files) | -- | Sub-models own their render contract; `Hints()` lives on the model |
| `menuHints()` dispatcher (stateless states) | `internal/app/model.go` | -- | AppModel owns the state machine; stateless states have no owning sub-model |
| Inline hint-set vars (deletion) | `internal/keys/hints.go` | -- | File shrinks to ~60 lines; vars replaced by keymap-backed instances |
| Runtime drift detector | `internal/app/menuhints_drift_test.go` | -- | Co-located with dispatcher; imports both `keys` and `ui` packages |
| Golden matrix | `internal/app/testdata/menu_*.golden` | -- | 13 files; captured via `RenderMenu` in test, not full chrome |

---

## Research Question Answers

### 1. What does `help.KeyMap` look like in bubbles/v2?

[VERIFIED: Context7 /charmbracelet/bubbles, confirmed against existing codebase usage]

```go
// charm.land/bubbles/v2/help
type KeyMap interface {
    ShortHelp() []key.Binding
    FullHelp()  [][]key.Binding
}
```

`key.Binding.Help()` returns `key.Help{Key string, Desc string}`. `HintsFromBindings` in `internal/keys/hints.go` already consumes this via `b.Help()`. No new import is needed.

Import paths confirmed already in use in `internal/keys/bindings.go`:
- `charm.land/bubbles/v2/key` — for `key.Binding`, `key.NewBinding`, `key.WithKeys`, `key.WithHelp`
- The `help.KeyMap` interface is satisfied by implementing `ShortHelp() []key.Binding` and `FullHelp() [][]key.Binding` — no import of `help` package needed in `bindings.go` itself (the interface is structural).

### 2. How does k9s handle keymap-as-source-of-truth?

[VERIFIED: ~/git/k9s/internal/model/menu_hint.go, ~/git/k9s/internal/ui/menu.go]

k9s `model.MenuHint` is structurally identical to sops-tui's `keys.MenuHint`:
```go
// k9s: internal/model/menu_hint.go
type MenuHint struct {
    Mnemonic    string
    Description string
    Visible     bool
}
```

k9s uses a retained-mode observer pattern — `Menu.StackPushed(c model.Component)` calls `m.HydrateMenu(c.Hints())`. sops-tui's CONTEXT.md D-25 (Phase 7) explicitly rejected this retained-mode pattern in favour of immediate-mode dispatch per enum switch on `(state, IsSearchActive)`. The dispatch architecture is already shipped; Phase 9 hardens the *data* side (what `Hints()` returns), not the dispatch mechanism.

The k9s precedent validates the `Visible bool` field approach for suppression — k9s uses the same field to hide numeric digit keys from the menu while keeping them functional. sops-tui's `Visible=false` on `Blame` is directly analogous.

### 3. What is `RequireGoldenStructure` and how does it work?

[VERIFIED: internal/testutil/golden.go — read directly]

```go
// internal/testutil/golden.go
func RequireGoldenStructure(t *testing.T, name, output string)
```

- Strips ANSI escape sequences via `github.com/charmbracelet/x/ansi.Strip(output)`
- Normalises trailing whitespace per line, LF endings
- Compares against `testdata/<name>.golden`
- Regenerates when `GOLDEN_UPDATE=1` env var is set

The golden files live in the same package's `testdata/` subdirectory. Phase 9's 13 golden files go in `internal/app/testdata/` (same location as the 4 existing resize goldens). The `name` argument maps to `testdata/<name>.golden` so naming convention directly determines the filename.

---

## Codebase Analysis: Current Shape of Each Sub-model

### FileListModel.Hints() — `internal/ui/filelist.go:383-390`

```go
func (m FileListModel) Hints() []keys.MenuHint {
    hints := keys.HintsFromBindings(m.keys.ShortHelp())
    hints = append(hints,
        keys.MenuHint{Mnemonic: "g", Description: "go to top", Visible: true},
        keys.MenuHint{Mnemonic: "G", Description: "go to bottom", Visible: true},
    )
    return hints
}
```

**Current count:** 12 (10 from ShortHelp + 2 appended).
**Phase 9 change (D-304):** Move `GoTop`/`GoBottom` into `FileListKeyMap.ShortHelp()`. `Hints()` becomes `return keys.HintsFromBindings(m.keys.ShortHelp())`. Count stays 12.
**Test that must keep passing:** `TestFileListHints` at filelist_test.go:179 — asserts `len(hints) == 12`, `hints[10].Mnemonic == "g"`, `hints[10].Description == "go to top"`, `hints[11].Mnemonic == "G"`, `hints[11].Description == "go to bottom"`. These stay green because `WithHelp` strings match exactly.

**Existing ShortHelp (10 bindings):** Up, Down, Open, Search, Info, ToggleSelect, BulkReKey, HealthCheck, Help, Quit. After D-304: GoTop and GoBottom are added to ShortHelp, making it 12 bindings.

**Note:** GoTop and GoBottom are currently only in `FullHelp()` at `bindings.go:82-86`. Plan 1 moves them into `ShortHelp()` as well (or adds them to both). This changes the `FullHelp()` group structure slightly — currently navigation group is `{Up, Down, GoTop, GoBottom, HalfUp, HalfDown}`. After D-304, GoTop/GoBottom appear in both ShortHelp and FullHelp — the overlap is acceptable per bubbles/v2 help rendering (ShortHelp and FullHelp are separate render paths).

### DetailModel.Hints() — `internal/ui/detail.go:820-830`

```go
func (m DetailModel) Hints() []keys.MenuHint {
    hints := keys.HintsFromBindings(m.keys.ShortHelp())
    for i := range hints {
        if hints[i].Mnemonic == "b" {
            hints[i].Visible = false
            break
        }
    }
    return hints
}
```

**Current count:** 13 (12 visible + 1 Visible=false Blame).
**Phase 9 change (D-307):** The post-hoc loop mutating `hints[i].Visible = false` moves to a `HiddenFromMenu()` method on `DetailKeyMap`. The drift detector uses this method to apply the same suppression when comparing. `Hints()` signature unchanged; implementation becomes cleaner.
**Test that must keep passing:** `TestDetailHints` at detail_test.go:604 — asserts `len(hints) == 13`, `visible == 12`, `invisibleMnemonics == []string{"b"}`. These stay green if `HiddenFromMenu()` returns `[]key.Binding{k.Blame}` and `Hints()` applies it.

**Existing ShortHelp bindings (13):** Up, Down, Reveal, RevealAll, Edit, Back, Search, Help, Quit, Copy, Blame, AddRecipient, RemoveRecipient.

### HelpModel.Hints() — `internal/ui/help.go:96-102`

```go
func (m HelpModel) Hints() []keys.MenuHint {
    return []keys.MenuHint{
        {Mnemonic: "Esc", Description: "close help", Visible: true},
        {Mnemonic: "?",   Description: "close help", Visible: true},
        {Mnemonic: "q",   Description: "quit",        Visible: true},
    }
}
```

**Current count:** 3 (literal slice, no backing keymap).
**Phase 9 change:** Introduce `HelpKeyMap` with 3 bindings. `Hints()` becomes `return keys.HintsFromBindings(m.keys.ShortHelp())`.
**Test that must keep passing:** `TestHelpHints` in help_test.go — asserts `len(hints) == 3`, `hints[0].Mnemonic == "Esc"`, `hints[0].Description == "close help"`. These stay green if `WithHelp` strings match exactly.
**HelpModel constructor:** `NewHelpModel(width, height int) HelpModel` — takes no keymap. After Plan 1, it either accepts an optional keymap or uses `DefaultHelpKeyMap` directly.

**Note:** `HelpModel.Update()` is not defined in `help.go` — it has no key handling of its own; AppModel drives `Esc`/`?`/`q` through the global key handlers. So `HelpKeyMap` bindings are data-only — no `key.Match` in the sub-model. The keymap just needs to encode the 3 mnemonics.

### DiffModel.Hints() — `internal/ui/diff.go:176-185`

```go
func (m DiffModel) Hints() []keys.MenuHint {
    return []keys.MenuHint{
        {Mnemonic: "y",   Description: "confirm re-encrypt", Visible: true},
        {Mnemonic: "n",   Description: "cancel",             Visible: true},
        {Mnemonic: "Esc", Description: "cancel",             Visible: true},
        {Mnemonic: "j",   Description: "scroll down",        Visible: true},
        {Mnemonic: "k",   Description: "scroll up",          Visible: true},
        {Mnemonic: "q",   Description: "quit",               Visible: true},
    }
}
```

**Current count:** 6 (literal slice).
**`DiffModel.Update()` keys (`diff.go:98-116`):** `"y"`, `"n"`, `"esc"`, `"j"`, `"k"` (5 keys). Quit (`"q"`) is handled globally by AppModel, not in DiffModel.Update(). So `DiffKeyMap` bindings: Confirm (`y`), Cancel (`n`), Close (`esc`), ScrollDown (`j`), ScrollUp (`k`), Quit (`q`).
**Test that must keep passing:** `TestDiffHints` in diff_test.go:140 — asserts `len == 6`, exact mnemonics and descriptions in order.
**Constructor:** `NewDiffModel(title string, entries []DiffEntry, width, height int) DiffModel` — no keymap param currently.

### HealthModel.Hints() — `internal/ui/health.go:181-189`

```go
func (m HealthModel) Hints() []keys.MenuHint {
    return []keys.MenuHint{
        {Mnemonic: "j",   Description: "scroll down",  Visible: true},
        {Mnemonic: "k",   Description: "scroll up",    Visible: true},
        {Mnemonic: "H",   Description: "close health", Visible: true},
        {Mnemonic: "Esc", Description: "close health", Visible: true},
        {Mnemonic: "q",   Description: "quit",         Visible: true},
    }
}
```

**Current count:** 5 (literal slice).
**Key handling:** HealthModel has no `Update()`. AppModel drives `j`/`k` via `ScrollDown()`/`ScrollUp()`, `H` and `Esc` via its own key dispatch, `q` via global handler. `HealthKeyMap` is data-only.
**Test that must keep passing:** `TestHealthHints` in health_test.go:141 — asserts `len == 5`, exact mnemonics.
**Constructor:** `NewHealthModel(width, height int) HealthModel` — no keymap param.

### HistoryModel.Hints() — `internal/ui/history.go:133-141`

```go
func (m HistoryModel) Hints() []keys.MenuHint {
    return []keys.MenuHint{
        {Mnemonic: "j",   Description: "scroll down",   Visible: true},
        {Mnemonic: "k",   Description: "scroll up",     Visible: true},
        {Mnemonic: "b",   Description: "close history", Visible: true},
        {Mnemonic: "Esc", Description: "close history", Visible: true},
        {Mnemonic: "q",   Description: "quit",          Visible: true},
    }
}
```

**Current count:** 5 (literal slice).
**Key handling:** HistoryModel has no `Update()`. AppModel drives all keys. `HistoryKeyMap` is data-only.
**Test that must keep passing:** `TestHistoryHints` in history_test.go:104 — asserts `len == 5`, exact mnemonics.
**Constructor:** `NewHistoryModel(filename string, width, height int) HistoryModel` — no keymap param.

### MetadataModel.Hints() — `internal/ui/metadata.go:167-175`

```go
func (m MetadataModel) Hints() []keys.MenuHint {
    return []keys.MenuHint{
        {Mnemonic: "j",   Description: "scroll down",    Visible: true},
        {Mnemonic: "k",   Description: "scroll up",      Visible: true},
        {Mnemonic: "i",   Description: "close metadata", Visible: true},
        {Mnemonic: "Esc", Description: "close metadata", Visible: true},
        {Mnemonic: "q",   Description: "quit",           Visible: true},
    }
}
```

**Current count:** 5 (literal slice).
**Key handling:** MetadataModel has no `Update()`. `MetadataKeyMap` is data-only.
**Test that must keep passing:** `TestMetadataHints` in metadata_test.go:246 — asserts `len == 5`, exact descriptions.
**Constructor:** `NewMetadataModel(meta MetadataContent, width, height int) MetadataModel` — no keymap param.

### RecipientFormModel.Hints() — `internal/ui/recipientform.go:163-168`

```go
func (m RecipientFormModel) Hints() []keys.MenuHint {
    return []keys.MenuHint{
        {Mnemonic: "Enter", Description: "confirm", Visible: true},
        {Mnemonic: "Esc",   Description: "cancel",  Visible: true},
    }
}
```

**Current count:** 2 (literal slice).
**`RecipientFormModel.Update()` keys (`recipientform.go:94-122`):** `"enter"` (validate + confirm) and `"esc"` (cancel). `RecipientFormKeyMap`: Confirm (`Enter`), Cancel (`Esc`).
**Note:** The `WithHelp` key string for Enter should be `"Enter"` (capital) to match the existing mnemonic in the literal slice. The actual `key.WithKeys("enter")` value is lowercase — but `WithHelp("Enter", "confirm")` produces the right display string. Precedent: existing `DetailKeyMap.Back` uses `key.WithKeys("esc")` + `key.WithHelp("esc", "back to file list")`. Confirm: use `key.WithHelp("Enter", "confirm")`.
**Test that must keep passing:** `TestRecipientFormHints` in recipientform_test.go:121 — asserts `hints[0].Mnemonic == "Enter"`, `hints[0].Description == "confirm"`.
**Constructor:** `NewRecipientFormModel(width, height int) RecipientFormModel` — no keymap param.

---

## Five Inline Package Vars — Current Content

Located in `internal/keys/hints.go`. All deleted in Plan 1; data moves to keymap `WithHelp()` strings.

### FileListSearchHints (hints.go:56-63)
```go
var FileListSearchHints = []MenuHint{
    {Mnemonic: "Esc",   Description: "exit search",    Visible: true},
    {Mnemonic: "Enter", Description: "select result",  Visible: true},
    {Mnemonic: "j/↓",  Description: "next result",    Visible: true},
    {Mnemonic: "k/↑",  Description: "prev result",    Visible: true},
    {Mnemonic: "?",     Description: "toggle help",    Visible: true},
    {Mnemonic: "q",     Description: "quit",           Visible: true},
}
```
**New type:** `FileListSearchKeyMap`. Note the `j/↓` and `k/↑` combined mnemonics — these require `key.WithHelp("j/↓", "next result")` style. The dispatcher uses this as the search-active override; after Plan 1 the arm becomes `return keys.HintsFromBindings(keys.DefaultFileListSearchKeyMap.ShortHelp())`.

### RecipientConfirmHints (hints.go:74-80)
```go
var RecipientConfirmHints = []MenuHint{
    {Mnemonic: "y",   Description: "confirm add/remove recipient", Visible: true},
    {Mnemonic: "n",   Description: "cancel",                      Visible: true},
    {Mnemonic: "Esc", Description: "cancel",                      Visible: true},
    {Mnemonic: "j",   Description: "scroll down",                 Visible: true},
    {Mnemonic: "k",   Description: "scroll up",                   Visible: true},
}
```
**New type:** `RecipientConfirmKeyMap`. Embeds `GlobalKeyMap` with `Quit` marked `Visible=false` (D-313). The quit-suppression doc-comment (lines 65-73) migrates to the new keymap.

### BulkReKeyConfirmHints (hints.go:90-96)
```go
var BulkReKeyConfirmHints = []MenuHint{
    {Mnemonic: "y",   Description: "confirm re-key this file", Visible: true},
    {Mnemonic: "n",   Description: "skip this file",           Visible: true},
    {Mnemonic: "Esc", Description: "abort bulk re-key",        Visible: true},
    {Mnemonic: "j",   Description: "scroll down",              Visible: true},
    {Mnemonic: "k",   Description: "scroll up",                Visible: true},
}
```
**New type:** `BulkReKeyConfirmKeyMap`. Same quit-suppression pattern as RecipientConfirm. Note different descriptions from RecipientConfirm: `"confirm re-key this file"`, `"skip this file"`, `"abort bulk re-key"`.

### RecipientListHints (hints.go:101-105)
```go
var RecipientListHints = []MenuHint{
    {Mnemonic: "1-9", Description: "select recipient to remove", Visible: true},
    {Mnemonic: "Esc", Description: "cancel",                     Visible: true},
    {Mnemonic: "q",   Description: "quit",                       Visible: true},
}
```
**New type:** `RecipientListKeyMap`. Note the `"1-9"` multi-key mnemonic — this is a display convention, not a real `key.WithKeys` value. Use `key.WithHelp("1-9", "select recipient to remove")` with `key.WithKeys("1", "2", "3", "4", "5", "6", "7", "8", "9")`. This is the only binding that uses a numeric range in the display. Acceptable because bubbles/v2 key rendering uses `WithHelp` for display only.

### FormatMenuHints (hints.go:110-115)
```go
var FormatMenuHints = []MenuHint{
    {Mnemonic: "j",     Description: "next format",    Visible: true},
    {Mnemonic: "k",     Description: "prev format",    Visible: true},
    {Mnemonic: "Enter", Description: "confirm format", Visible: true},
    {Mnemonic: "Esc",   Description: "cancel",         Visible: true},
}
```
**New type:** `FormatMenuKeyMap`. 4 bindings. No quit visible (format menu is a modal overlay per D-09; quit is available globally but suppressed from this menu contextually — Plan 1 author decides whether to explicitly add Quit as Visible=false or leave it absent from ShortHelp entirely).

---

## menuHints() Dispatcher — Current Arms (`model.go:1498-1533`)

```go
func (m AppModel) menuHints() []keys.MenuHint {
    if m.state == stateFileList && m.fileList.IsSearchActive() {
        return keys.FileListSearchHints           // D-11 override
    }
    switch m.state {
    case stateFileList:       return m.fileList.Hints()
    case stateDetail:         return m.detail.Hints()
    case stateMetadata:       return m.metadata.Hints()
    case stateDiff:           return m.diff.Hints()
    case stateRecipientConfirm:   return keys.RecipientConfirmHints
    case stateBulkReKeyConfirm:   return keys.BulkReKeyConfirmHints
    case stateHelp:           return m.help.Hints()
    case stateHistory:        return m.history.Hints()
    case stateHealth:         return m.health.Hints()
    case stateRecipientForm:  return m.recipientForm.Hints()
    case stateRecipientList:  return keys.RecipientListHints
    case stateFormatMenu:     return keys.FormatMenuHints
    }
    return nil
}
```

**After Plan 1, the 5 stateless arms become:**
- `stateRecipientConfirm` → `return keys.HintsFromBindings(keys.DefaultRecipientConfirmKeyMap.ShortHelp())`
- `stateBulkReKeyConfirm` → `return keys.HintsFromBindings(keys.DefaultBulkReKeyConfirmKeyMap.ShortHelp())`
- `stateRecipientList` → `return keys.HintsFromBindings(keys.DefaultRecipientListKeyMap.ShortHelp())`
- `stateFormatMenu` → `return keys.HintsFromBindings(keys.DefaultFormatMenuKeyMap.ShortHelp())`
- Search override → `return keys.HintsFromBindings(keys.DefaultFileListSearchKeyMap.ShortHelp())`

**Comment on line 1492:** "recipientAction, IsSearchActive" — Plan 2 D-309 cleanup: change to "(state, IsSearchActive)".

---

## recipientAction — Confirmed Not a Dispatch Axis

[VERIFIED: grepping model.go]

`recipientAction` is a string field (`"add"`, `"remove"`, `"healthcheck"` sentinel) on `AppModel`. It is used in `Update()` business logic (line 803: `action := m.recipientAction`) but **never read in `menuHints()`**. The comment on line 1492 ("recipientAction, IsSearchActive") describes the *intended* dispatch signature from Phase 7 D-10, but the code never implemented it. The D-309 amendment is purely a documentation correction — no code path changes needed. Leftover references:
- `model.go:262` — struct field definition (keep; still used in Update logic)
- `model.go:1492` — comment in `menuHints()` (Plan 2: update comment text to drop `recipientAction`)
- `hints_test.go:3` — test file header comment (Plan 2: update comment)

---

## Pattern Map — 11 New Keymaps vs Existing Analogs

All 11 new keymaps follow the pattern of `FileListKeyMap` / `DetailKeyMap` in `bindings.go`.

| New Keymap | Closest Analog | Key Deltas |
|------------|---------------|------------|
| `HelpKeyMap` | `GlobalKeyMap` | 3 bindings total: Close (Esc), ToggleHelp (?), Quit (q). No navigation. No embedding needed (or can embed GlobalKeyMap and add Close). |
| `DiffKeyMap` | `DetailKeyMap` (has Back/Esc) | 6 bindings: Confirm (y), Cancel (n), Close (Esc), ScrollDown (j), ScrollUp (k), Quit (q). Embeds nothing (quit is explicit, not via embedding). |
| `HealthKeyMap` | `HistoryKeyMap` (same pattern) | 5 bindings: ScrollDown (j), ScrollUp (k), Close (H), CloseAlt (Esc), Quit (q). |
| `HistoryKeyMap` | `HealthKeyMap` (same pattern) | 5 bindings: ScrollDown (j), ScrollUp (k), Close (b), CloseAlt (Esc), Quit (q). |
| `MetadataKeyMap` | `HistoryKeyMap` | 5 bindings: ScrollDown (j), ScrollUp (k), Close (i), CloseAlt (Esc), Quit (q). |
| `RecipientFormKeyMap` | `DetailKeyMap` (similar Enter/Esc pattern) | 2 bindings: Confirm (Enter), Cancel (Esc). No navigation. |
| `FileListSearchKeyMap` | `FileListKeyMap` (same sub-context) | 6 bindings: ExitSearch (Esc), Select (Enter), NextResult (j/↓), PrevResult (k/↑), ToggleHelp (?), Quit (q). |
| `RecipientConfirmKeyMap` | `BulkReKeyConfirmKeyMap` (near-twin) | 5 bindings: Confirm (y), Cancel (n), Abort (Esc), ScrollDown (j), ScrollUp (k). Embeds GlobalKeyMap; Quit Visible=false. |
| `BulkReKeyConfirmKeyMap` | `RecipientConfirmKeyMap` (near-twin) | Same structure, different descriptions. |
| `RecipientListKeyMap` | `GlobalKeyMap` (minimal) | 3 bindings: Select (1-9 range mnemonic), Cancel (Esc), Quit (q). |
| `FormatMenuKeyMap` | `RecipientFormKeyMap` (small modal) | 4 bindings: Next (j), Prev (k), Confirm (Enter), Cancel (Esc). |

**Binding field naming convention** (from existing DetailKeyMap/FileListKeyMap):
- Action verbs matching the binding's function: `Confirm`, `Cancel`, `Close`, `CloseAlt`, `ScrollDown`, `ScrollUp`, `Select`, `Next`, `Prev`, `ExitSearch`, `NextResult`, `PrevResult`.

---

## Standard Stack

No new dependencies. All tooling already in `go.mod`.

| Library | Version | Import Path | Role in Phase 9 |
|---------|---------|-------------|-----------------|
| charm.land/bubbles/v2 | v2.x | `charm.land/bubbles/v2/key` | `key.Binding`, `key.NewBinding`, `key.WithKeys`, `key.WithHelp` |
| charm.land/bubbles/v2 | v2.x | `charm.land/bubbles/v2/help` | `help.KeyMap` interface (structural — no import needed in bindings.go) |
| stretchr/testify | v1.x | `github.com/stretchr/testify/require` | Drift detector assertions |
| charmbracelet/x/ansi | latest | `github.com/charmbracelet/x/ansi` | Already used by testutil.RequireGoldenStructure |

[VERIFIED: go.mod and existing bindings.go/hints.go imports confirmed these are present]

---

## Risks and Landmines

### 1. Sub-model Constructor Signature Changes

4 of the 6 sub-models being refactored have no keymap param in their constructor: `NewHealthModel`, `NewHistoryModel`, `NewMetadataModel`, `NewRecipientFormModel`. DiffModel and HelpModel also lack a keymap param.

**Risk:** Adding a `keys XxxKeyMap` parameter to each constructor requires updating every call-site in `model.go` (6 sites total — one per `New*Model` call). This is low-risk because all call-sites are in `internal/app/model.go` and the compiler will catch missed updates.

**Recommendation:** Use the same pattern as `DefaultFileListKeyMap` — add a `keys XxxKeyMap` field to the struct, default-initialize it to `DefaultXxxKeyMap` inside `New*Model()` without changing the constructor signature. This is zero call-site impact and consistent with how `FileListModel` and `DetailModel` already hold their keymap as a field.

```go
// Pattern: zero call-site impact
func NewHealthModel(width, height int) HealthModel {
    return HealthModel{
        keys:    DefaultHealthKeyMap,  // new field, default-initialized here
        loading: true,
        width:   width,
        height:  height,
    }
}
```

### 2. Tests That Assert Literal Expected Hint Slices

The following existing tests assert exact mnemonic + description strings:

| Test | File | Asserts |
|------|------|---------|
| `TestFileListHints` | filelist_test.go:179 | `hints[10].Description == "go to top"`, `hints[11].Description == "go to bottom"` |
| `TestDetailHints` | detail_test.go:604 | invisible mnemonic is `"b"`, visible count is 12 |
| `TestHelpHints` | help_test.go (TestHelpHints) | `hints[0].Mnemonic == "Esc"`, `hints[0].Description == "close help"` |
| `TestDiffHints` | diff_test.go:140 | `hints[0].Description == "confirm re-encrypt"`, exact order |
| `TestHealthHints` | health_test.go:141 | `hints[2].Description == "close health"` |
| `TestHistoryHints` | history_test.go:104 | `hints[2].Description == "close history"` |
| `TestMetadataHints` | metadata_test.go:246 | `hints[2].Description == "close metadata"` |
| `TestRecipientFormHints` | recipientform_test.go:121 | `hints[0].Description == "confirm"` |
| `TestMenuHints_StateDetail` | hints_test.go:65 | `len == 13` |
| `TestMenuHints_StateRecipientConfirm` | hints_test.go:91 | equality with `keys.RecipientConfirmHints` |
| `TestMenuHints_StateBulkReKeyConfirm` | hints_test.go:99 | equality with `keys.BulkReKeyConfirmHints` |
| `TestMenuHints_StateRecipientList` | hints_test.go:139 | equality with `keys.RecipientListHints` |
| `TestMenuHints_StateFormatMenu` | hints_test.go:147 | equality with `keys.FormatMenuHints` |

**Critical:** The last 4 (`StateRecipientConfirm`, `StateBulkReKeyConfirm`, `StateRecipientList`, `StateFormatMenu`) assert equality with the **package-var objects that are being deleted**. After Plan 1, these assertions must change from `require.Equal(t, keys.RecipientConfirmHints, hints)` to a new form.

**Options:**
- Assert against `keys.HintsFromBindings(keys.DefaultRecipientConfirmKeyMap.ShortHelp())`
- Assert element-by-element on specific mnemonics and lengths

The first option is cleaner and self-documenting. Plan 1 must update these 4 test assertions when it deletes the package vars.

**All other tests stay green** as long as `WithHelp` strings match the exact description strings currently in the literal slices. The description strings are explicitly documented above for each model.

### 3. Visibility Override Design Choice (D-307)

Two viable patterns:

**Option A: Method on keymap struct**
```go
func (k DetailKeyMap) HiddenFromMenu() []key.Binding {
    return []key.Binding{k.Blame}
}
```
Pro: Per-state policy, not per-binding metadata. `key.Binding` stays portable. Explicit about intent. Con: Drift detector needs a type assertion or interface to call it.

**Option B: Empty `WithHelp` string convention**
```go
Blame: key.NewBinding(
    key.WithKeys("b"),
    key.WithHelp("b", ""),  // empty desc = hidden from menu
)
```
Pro: No extra interface needed. Con: Breaks the `?` overlay (FullHelp also uses this binding) — empty desc renders as no description in the help overlay.

**Recommendation (aligns with D-307 text):** Option A — the method approach. The drift detector can require an optional interface:
```go
type MenuVisibilityOverrider interface {
    HiddenFromMenu() []key.Binding
}
```
The test checks if the keymap implements this interface and applies the suppression if so. This is idiomatic Go.

### 4. Drift Detector Construction Strategy

**Option A: Real constructors** — `NewHelpModel(80, 24)`, `NewDiffModel("test", []DiffEntry{}, 80, 24)`, etc.

**Option B: Zero-value with manual keymap injection** — `HelpModel{}.Hints()` (but `keys` field would be zero-value, returning empty ShortHelp).

**Option A is strongly preferred** (aligns with D-307 recommendation, CONTEXT.md §Claude's Discretion). Real constructors set `keys = DefaultXxxKeyMap`, which is what production code uses. Zero-value would require setting the field manually, which is indirection with no benefit. The only constructor that is complex is `NewRecipientFormModel` (needs `textinput.New()`) — still callable in tests without side effects since it doesn't call SOPS.

### 5. Golden File Naming Convention

CONTEXT.md §Claude's Discretion recommends `menu_file_list.golden` (lowercase + underscore). Cross-referencing with the existing 4 resize goldens: `resize_40x12.golden`, `resize_80x24.golden` — they use lowercase. The state names in code are `stateFileList`, `stateDetail`, etc. — camelCase.

**Recommendation (from CONTEXT.md):** `menu_file_list.golden`, `menu_detail.golden`, `menu_metadata.golden`, `menu_diff.golden`, `menu_help.golden`, `menu_history.golden`, `menu_health.golden`, `menu_recipient_list.golden`, `menu_recipient_form.golden`, `menu_format_menu.golden`, `menu_recipient_confirm.golden`, `menu_bulk_re_key_confirm.golden`, `menu_file_list_search.golden`.

That is 13 files, matching D-308 exactly.

### 6. `recipientAction` Cleanup Scope

[VERIFIED: grepping model.go and hints_test.go]

`recipientAction` is **not** a dead field — it is actively used in `Update()` business logic (lines 777, 803, 866, 867, 1125, 1264). The field stays. Only the documentation references need updating:
- `model.go:1492` comment: remove "recipientAction" from the tuple description
- `hints_test.go:3` header comment: same removal
- No code changes required; no signatures to update

### 7. Bubbles/v2 Import Paths

[VERIFIED: existing bindings.go and hints.go]

```go
import (
    "charm.land/bubbles/v2/key"
    // help.KeyMap interface is satisfied structurally — no import needed in bindings.go
    // if the interface check is in test files only
)
```

The `help.KeyMap` interface check (`var _ help.KeyMap = XxxKeyMap{}`) requires importing `charm.land/bubbles/v2/help` in the test or in bindings.go. Existing pattern: `internal/ui/filelist_test.go:14` uses `var _ keys.Hinter = ui.FileListModel{}` — a compile-time check. The analog for keymaps would be in `keys/bindings_test.go`:
```go
import "charm.land/bubbles/v2/help"
var _ help.KeyMap = HelpKeyMap{}
```
This keeps `help` import out of production code (bindings.go imports only `key`).

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| ANSI stripping in golden comparison | custom ANSI parser | `testutil.RequireGoldenStructure` | Already in codebase; handles normalisation, GOLDEN_UPDATE=1 |
| Key binding display string generation | custom format | `key.WithHelp(mnem, desc)` + `b.Help()` | bubbles/v2 contract; `HintsFromBindings` already consumes it |
| Per-state test model construction | raw struct literals | real `New*Model(...)` constructors | Ensures default keymap is set; matches production code path |
| Visibility filtering in render | custom filter in View() | `MenuHint.Visible` field checked in `RenderMenu` | Already implemented in `menu.go:59-66` |

---

## Common Pitfalls

### Pitfall 1: Description String Mismatch Breaks Existing Tests

**What goes wrong:** Plan 1 author uses a slightly different description string in `WithHelp()` than the literal slice had (e.g., `"close"` instead of `"close help"`), causing existing `TestHealthHints` etc. to fail.

**Why it happens:** The literal strings in the 6 sub-model `Hints()` methods are the ground truth for the existing tests. Changing even one character changes the test expectation.

**How to avoid:** Use the exact description strings documented in the "Codebase Analysis" section above. Every description string is captured verbatim. Cross-check with the test assertions before committing.

**Warning signs:** `TestHealthHints` / `TestHistoryHints` / `TestMetadataHints` / `TestRecipientFormHints` / `TestHelpHints` / `TestDiffHints` fail after Plan 1 is applied.

### Pitfall 2: Forgetting to Update hints_test.go Equality Assertions

**What goes wrong:** Plan 1 deletes `keys.RecipientConfirmHints` but leaves `require.Equal(t, keys.RecipientConfirmHints, hints)` in hints_test.go — compilation fails.

**Why it happens:** Deletion of the package var must be coordinated with updating the 4 test assertions that reference it.

**How to avoid:** The 4 tests at hints_test.go:91, 99, 139, 147 must change in the same commit that deletes the vars. New assertion form: `require.Equal(t, keys.HintsFromBindings(keys.DefaultRecipientConfirmKeyMap.ShortHelp()), hints)`.

### Pitfall 3: GoTop/GoBottom in FullHelp() After D-304

**What goes wrong:** After moving GoTop/GoBottom into `FileListKeyMap.ShortHelp()`, they appear twice in `FullHelp()` (once in the navigation group already, once because ShortHelp is sometimes rendered separately).

**Why it happens:** The navigation group in `FullHelp()` already contains GoTop/GoBottom (`bindings.go:82-86`). Adding them to ShortHelp does not cause a FullHelp regression because ShortHelp and FullHelp are separate render paths.

**How to avoid:** No action needed — bubbles/v2 help rendering uses ShortHelp and FullHelp independently. The `TestFileListHints` assertion at filelist_test.go:190-193 will still pass since it checks `hints[10].Mnemonic == "g"` which comes from the 11th binding in ShortHelp (index 10, 0-based), now added via the keymap.

**Warning signs:** Test assertions on exact indices in `TestFileListHints` fail. Check the ordering in `FileListKeyMap.ShortHelp()` after adding GoTop/GoBottom.

### Pitfall 4: Drift Detector Import Cycle Risk

**What goes wrong:** Placing the drift test in `internal/keys/` would create an import cycle: `internal/keys` → `internal/ui` (to call `NewHealthModel().Hints()`) → `internal/keys` (sub-models import `keys`).

**Why it happens:** `keys` is a leaf package; sub-models import it.

**How to avoid (D-306):** Drift detector lives in `internal/app/menuhints_drift_test.go`. The `app` package already imports both `internal/keys` and `internal/ui` — no new import edges needed.

### Pitfall 5: `RenderMenu` Width for Golden Capture

**What goes wrong:** Plan 2 calls `RenderMenu(hints, 0)` or an arbitrary width, producing different column widths than the actual app renders at.

**Why it happens:** `RenderMenu` uses `colWidth := width / menuCols`, floored at `minMenuCol (8)`. With width=0, colWidth clamps to 8. The golden shows truncated descriptions.

**How to avoid:** Use a consistent fixed width for all 13 goldens, e.g., `RenderMenu(hints, 80)` or the width used in `buildAppModel` (80×24 model). Document the width in the test. Use `GOLDEN_UPDATE=1` to generate, then inspect for readability.

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + `stretchr/testify` v1.x |
| Config | None — `go test ./...` directly |
| Quick run | `go test ./internal/keys/... ./internal/ui/... ./internal/app/... -run TestMenuHints -v` |
| Full suite | `go test ./...` |

### Phase Requirements to Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| UI-09 | Every Hints() derives from keymap | unit (per sub-model) | `go test ./internal/ui/... -run 'Hints'` | All 8 Hints tests exist; update expectations |
| UI-09 | Compile-time `help.KeyMap` check on all 11 keymaps | compile | `go build ./internal/keys/...` | Wave 0 gap: add to bindings_test.go |
| UI-10 | Per-state menu re-hydration | golden matrix (13) | `go test ./internal/app/... -run TestMenuGolden` | Wave 0 gap: Plan 2 creates test + goldens |
| UI-10 | Drift detector: Hints() == HintsFromBindings(km.ShortHelp()) | runtime equality | `go test ./internal/app/... -run TestMenuHints_Drift` | Wave 0 gap: Plan 2 creates |
| UI-11 | ? overlay retained (no regression) | existing tests | `go test ./internal/ui/... -run TestHelp` | TestHelpHints + view tests exist |

### Nyquist Signals per Requirement

| SC | Signal 1 | Signal 2 |
|----|----------|----------|
| SC1 (UI-09: total derivation) | Per-sub-model `TestXxxHints` (existing, updated) | Compile-time `var _ help.KeyMap = XxxKeyMap{}` checks |
| SC2 (UI-10: modal re-hydration) | 13-entry golden matrix (Plan 2) | Drift detector runtime equality (Plan 2) |
| SC3 (UI-10: pure function of state, IsSearchActive) | `TestMenuHints_State*` tests (existing, 13 tests) | 13 golden files locking rendered output per state |
| SC4 (UI-11: ? overlay retained) | `TestHelpHints` (existing) | `TestHelpView*` tests (existing, 4 tests) |
| SC5 (no second edit) | Drift detector — changing any binding.Help() description auto-fails if Hints() returns stale literal | Golden matrix — changing binding description changes golden, fails if not regenerated |

### Sampling Rate
- **Per task commit:** `go test ./internal/keys/... ./internal/ui/... -count=1`
- **Per wave merge:** `go test ./... -count=1`
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps (Plan 1)
- [ ] `internal/keys/bindings_test.go` — compile-time `var _ help.KeyMap = XxxKeyMap{}` assertions for all 11 new keymaps
- [ ] Update 4 hint equality assertions in `internal/app/hints_test.go` (lines 91, 99, 139, 147) when package vars are deleted

### Wave 0 Gaps (Plan 2)
- [ ] `internal/app/menuhints_drift_test.go` — new file with drift detector + 13 golden sub-tests
- [ ] `internal/app/testdata/menu_*.golden` — 13 new golden files (generated via `GOLDEN_UPDATE=1`)

---

## Security Domain

Phase 9 is a pure refactor of keybinding metadata — no new network calls, no new input vectors, no new secret access. The only security-relevant check:

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V5 Input Validation | No | Key bindings are not user input |

No new threat patterns introduced. Existing SOPS subprocess handling unchanged.

---

## Environment Availability

Step 2.6: SKIPPED (no external dependencies — this phase modifies Go source files only; no CLI tools, external services, or databases involved).

---

## Implementation Skeleton

### Plan 1: Keymap Extraction + Hints() Derivation

**Task 1 — Add GoTop/GoBottom to FileListKeyMap.ShortHelp() (`internal/keys/bindings.go`)**

- Modify `FileListKeyMap.ShortHelp()` to return 12 bindings (add GoTop, GoBottom after Quit, or before Help — use position that produces same menu slot ordering as current `hints[10]=="g"`, `hints[11]=="G"`).
- Run `TestFileListHints` — assert still green.

**Task 2 — Simplify FileListModel.Hints() (`internal/ui/filelist.go`)**

- Replace the current 4-line implementation (HintsFromBindings + append) with `return keys.HintsFromBindings(m.keys.ShortHelp())`.
- Run `TestFileListHints` — assert still green (g/G now come from keymap, same strings).

**Task 3 — Add 6 sub-model keymap types + HiddenFromMenu() contract (`internal/keys/bindings.go`)**

Files: `bindings.go` only.
New types: `HelpKeyMap`, `DiffKeyMap`, `HealthKeyMap`, `HistoryKeyMap`, `MetadataKeyMap`, `RecipientFormKeyMap`.
Each: struct with named `key.Binding` fields, `ShortHelp()`, `FullHelp()`, `Default*KeyMap` instance.
`DetailKeyMap.HiddenFromMenu()` method: return `[]key.Binding{k.Blame}`.
Compile-time checks in `bindings_test.go`: `var _ help.KeyMap = HelpKeyMap{}` etc.

**Task 4 — Refactor 6 sub-model Hints() to derive from keymaps (`internal/ui/{help,diff,health,history,metadata,recipientform}.go`)**

For each: add `keys XxxKeyMap` field to struct; initialize to `DefaultXxxKeyMap` in constructor; replace literal Hints() return with `keys.HintsFromBindings(m.keys.ShortHelp())`.
For DetailModel: update `Hints()` to apply `HiddenFromMenu()` via the method rather than the inline mnemonic search loop.
Run all 6 hint tests — all green.

**Task 5 — Add 5 stateless-state keymap types (`internal/keys/bindings.go`)**

New types: `FileListSearchKeyMap`, `RecipientConfirmKeyMap`, `BulkReKeyConfirmKeyMap`, `RecipientListKeyMap`, `FormatMenuKeyMap`.
`RecipientConfirmKeyMap` and `BulkReKeyConfirmKeyMap`: embed `GlobalKeyMap`; declare `HiddenFromMenu() []key.Binding { return []key.Binding{k.Quit} }`.
`Default*KeyMap` instances for each.
Migrate quit-suppression doc-comment from `hints.go:65-73` to a shared block in `bindings.go`.

**Task 6 — Delete 5 inline package vars + update dispatcher arms (`internal/keys/hints.go` + `internal/app/model.go`)**

`hints.go`: delete `FileListSearchHints`, `RecipientConfirmHints`, `BulkReKeyConfirmHints`, `RecipientListHints`, `FormatMenuHints`.
`model.go:menuHints()`: update 5 stateless arms to call `keys.HintsFromBindings(keys.Default*KeyMap.ShortHelp())`.
Update `hints_test.go:91,99,139,147`: change `require.Equal(t, keys.XxxHints, hints)` to `require.Equal(t, keys.HintsFromBindings(keys.DefaultXxxKeyMap.ShortHelp()), hints)`.
Run `TestMenuHints_State*` tests — all green.
Run `go test ./...` — full green.

### Plan 2: Drift Detector + Golden Matrix + D-309 Documentation

**Task 1 — Create drift detector test (`internal/app/menuhints_drift_test.go`)**

New file. For each of the 8 sub-models with a keymap: construct via real constructor, call `Hints()`, call `keys.HintsFromBindings(km.ShortHelp())` applying `HiddenFromMenu()` if the keymap implements `MenuVisibilityOverrider`, assert equality.
For stateless states (5 keymaps): call `menuHints()` with `m.state = stateXxx`, compare to `keys.HintsFromBindings(keys.DefaultXxxKeyMap.ShortHelp())`.
Verify the search-active override arm separately.
Total: ~13 sub-tests covering every dispatcher branch.

**Task 2 — Create 13-entry golden matrix (`internal/app/menuhints_drift_test.go` + `testdata/menu_*.golden`)**

Add a `TestMenuGolden` test function. For each of the 13 `(state, IsSearchActive)` tuples: construct `AppModel` via `buildAppModel`, set state, call `menuHints()`, call `RenderMenu(hints, 80)`, call `testutil.RequireGoldenStructure(t, "menu_<name>", rendered)`.
Generate all 13 golden files with `GOLDEN_UPDATE=1 go test ./internal/app/... -run TestMenuGolden`.
Inspect generated goldens — verify content matches expected binding descriptions.

**Task 3 — D-309 documentation amendment**

Update `model.go:1492` comment: change "(state, recipientAction, IsSearchActive)" to "(state, IsSearchActive)".
Update `hints_test.go:3` header comment: remove `recipientAction` reference.
Write `09-02-SUMMARY.md` noting the amendment.

---

## Open Questions for the Planner

### OQ-1: ShortHelp() ordering for FileListKeyMap after adding GoTop/GoBottom

**What we know:** Current `TestFileListHints` asserts `hints[10].Mnemonic == "g"` and `hints[11].Mnemonic == "G"`. After D-304, GoTop/GoBottom must be the 11th and 12th items in `ShortHelp()` (0-indexed: positions 10 and 11). Current ShortHelp has 10 items: `{Up, Down, Open, Search, Info, ToggleSelect, BulkReKey, HealthCheck, Help, Quit}`.

**What's unclear:** Where to insert GoTop/GoBottom — before Help/Quit, after Quit, or as a new tail? To preserve test assertion on `hints[10].Mnemonic == "g"`, GoTop must be at position 10. Current position 10 is Help (index 8, 0-based) and Quit (index 9). Adding GoTop at index 10 and GoBottom at index 11 means appending after Quit.

**Recommendation:** Append GoTop and GoBottom after Quit in `ShortHelp()`. This preserves all existing index assertions in `TestFileListHints` exactly.

### OQ-2: HiddenFromMenu() interface name and location

**What we know:** D-307 says "final shape is Plan 1's discretion". CONTEXT.md recommends a method on each keymap struct.

**What's unclear:** Whether the drift detector uses a named interface or type-asserts to concrete types.

**Recommendation:** Define a package-internal interface in `internal/keys/bindings.go`:
```go
// menuVisibilityOverrider is implemented by keymaps that suppress some
// bindings from the persistent menu via Visible=false.
type menuVisibilityOverrider interface {
    HiddenFromMenu() []key.Binding
}
```
Only `DetailKeyMap`, `RecipientConfirmKeyMap`, and `BulkReKeyConfirmKeyMap` implement it. The drift detector type-asserts and applies suppression if the interface is present. Unexported interface — tests in the same package can access it directly.

### OQ-3: FormatMenuKeyMap — should Quit be explicitly Visible=false or absent from ShortHelp?

**What we know:** Current `FormatMenuHints` has 4 entries with no quit. The dispatcher switch arm returns only these 4 hints. No explicit quit suppression comment exists.

**What's unclear:** Whether FormatMenu should suppress quit (like confirm states) or simply omit it from ShortHelp.

**Recommendation:** Omit Quit from `FormatMenuKeyMap.ShortHelp()` entirely. FormatMenu is a transient modal (RoundedBorder overlay); global quit via AppModel still works, but showing it in the menu is confusing in a 4-item selection modal. This differs from confirm states (where quit is deliberately suppressed as a safety measure and documented as such). For FormatMenu, "not advertising it" suffices.

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | All 4 "AppModel-driven" sub-models (Health, History, Metadata, RecipientForm's Enter/Esc only) do not have their own `Update()` with key handling that would require `key.Match` in the keymap | Codebase Analysis | If any has hidden key dispatch, the keymap would need additional bindings; but grep confirms no `Update()` on Health, History, Metadata |
| A2 | `DiffModel.Update()` handles quit via the `"q"` string — but AppModel also has a global quit handler. The `q` binding in `DiffKeyMap` is for menu display only; actual dispatch happens in AppModel | Codebase Analysis | No impact on correctness; just documentation clarity |

**Note:** Both assumed claims were cross-verified via direct code inspection. Confidence level: HIGH.

---

## Sources

### Primary (HIGH confidence)
- `internal/keys/hints.go` — verified line-by-line current content of all 5 package vars and interfaces
- `internal/keys/bindings.go` — verified FileListKeyMap, DetailKeyMap patterns in full
- `internal/ui/{filelist,detail,help,diff,health,history,metadata,recipientform}.go` — verified all 8 Hints() implementations at their documented line numbers
- `internal/app/model.go:1491-1533` — verified `menuHints()` dispatcher content
- `internal/app/hints_test.go` — verified all 13 test function assertions
- `internal/testutil/golden.go` — verified RequireGoldenStructure signature and GOLDEN_UPDATE=1 behavior
- `internal/ui/menu.go` — verified RenderMenu width behavior
- `~/git/k9s/internal/model/menu_hint.go`, `~/git/k9s/internal/ui/menu.go` — verified k9s MenuHint struct parity
- Context7 /charmbracelet/bubbles — verified `help.KeyMap` interface shape

### Secondary (MEDIUM confidence)
- `.planning/phases/09-keybinding-discoverability/09-CONTEXT.md` — all 16 decisions
- `.planning/REQUIREMENTS.md` — UI-09, UI-10, UI-11 requirement text
- `.planning/ROADMAP.md` Phase 9 section — 5 success criteria

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all imports verified in codebase; no new dependencies
- Architecture: HIGH — all code paths read directly; no speculation
- Pitfalls: HIGH — derived from actual test assertions and grep-verified call patterns

**Research date:** 2026-04-30
**Valid until:** 2026-06-30 (stable — no fast-moving dependencies; bubbles/v2 API is settled)
