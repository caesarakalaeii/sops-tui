# Phase 9: Keybinding Discoverability - Pattern Map

**Mapped:** 2026-04-30
**Files analyzed:** 13 (3 created, 10 modified)
**Analogs found:** 13 / 13

---

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/keys/bindings.go` (modified) | config/keymap | transform | self — existing `FileListKeyMap`/`DetailKeyMap` blocks | exact |
| `internal/keys/hints.go` (modified — delete 5 vars) | config/hints | transform | self — remaining `MenuHint`/`Hinter`/`HintsFromBindings` | exact |
| `internal/ui/filelist.go` (modified) | component | request-response | self — `FileListModel.Hints()` at line 383 | exact |
| `internal/ui/detail.go` (modified) | component | request-response | self — `DetailModel.Hints()` at line 820 | exact |
| `internal/ui/help.go` (modified) | component | request-response | `internal/ui/health.go:181` — same literal-return Hints() shape | exact |
| `internal/ui/diff.go` (modified) | component | request-response | `internal/ui/health.go:181` — same literal-return Hints() shape | exact |
| `internal/ui/health.go` (modified) | component | request-response | `internal/ui/history.go:133` — structural twin | exact |
| `internal/ui/history.go` (modified) | component | request-response | `internal/ui/health.go:181` — structural twin | exact |
| `internal/ui/metadata.go` (modified) | component | request-response | `internal/ui/history.go:133` — structural twin | exact |
| `internal/ui/recipientform.go` (modified) | component | request-response | `internal/ui/help.go:96` — 2-entry literal return | exact |
| `internal/app/model.go` (modified — `menuHints()`) | controller | request-response | self — existing switch arms at line 1498 | exact |
| `internal/app/hints_test.go` (modified — 4 assertions) | test | transform | self — existing assertions at lines 91, 99, 139, 147 | exact |
| `internal/keys/bindings_test.go` (modified — add 11 compile checks) | test | transform | self — existing `var _ help.KeyMap = FileListKeyMap{}` at line 68 | exact |
| `internal/app/menuhints_drift_test.go` (created) | test | transform | `internal/app/hints_test.go` — same `buildAppModel` + `menuHints()` driver | role-match |
| `internal/app/testdata/menu_*.golden` (13 created) | test fixture | transform | `internal/app/testdata/resize_80x24.golden` — `RequireGoldenStructure` family | role-match |

---

## Pattern Assignments

### `internal/keys/bindings.go` — 11 new keymap types

**Analog:** `internal/keys/bindings.go` — existing `FileListKeyMap` (lines 43–140) and `DetailKeyMap` (lines 145–291)

**Struct + embedding pattern** (bindings.go lines 43–70):
```go
// FileListKeyMap holds keybindings for the file list view.
// It embeds GlobalKeyMap so global keys are available everywhere.
// Implements help.KeyMap via ShortHelp() and FullHelp().
type FileListKeyMap struct {
    GlobalKeyMap

    Up           key.Binding
    Down         key.Binding
    GoTop        key.Binding
    GoBottom     key.Binding
    // ... more bindings
}
```

**ShortHelp + FullHelp methods** (bindings.go lines 72–86):
```go
func (k FileListKeyMap) ShortHelp() []key.Binding {
    return []key.Binding{k.Up, k.Down, k.Open, k.Search, k.Info, k.ToggleSelect, k.BulkReKey, k.HealthCheck, k.Help, k.Quit}
}

func (k FileListKeyMap) FullHelp() [][]key.Binding {
    return [][]key.Binding{
        {k.Up, k.Down, k.GoTop, k.GoBottom, k.HalfUp, k.HalfDown},
        {k.Open, k.Search, k.Info, k.ToggleSelect, k.BulkReKey, k.HealthCheck},
        {k.Help, k.Quit},
    }
}
```

**Default instance pattern** (bindings.go lines 88–140):
```go
var DefaultFileListKeyMap = FileListKeyMap{
    GlobalKeyMap: DefaultGlobalKeyMap,
    Up: key.NewBinding(
        key.WithKeys("k", "up"),
        key.WithHelp("k/↑", "move up"),
    ),
    // ... remaining bindings
}
```

**Deltas for the 11 new types:**
- 6 sub-model keymaps (`HelpKeyMap`, `DiffKeyMap`, `HealthKeyMap`, `HistoryKeyMap`, `MetadataKeyMap`, `RecipientFormKeyMap`): same struct + ShortHelp/FullHelp + Default*KeyMap shape. No `GlobalKeyMap` embedding needed for sub-models whose quit is explicit (e.g., `DiffKeyMap` includes its own `Quit key.Binding`). Data-only models (Health, History, Metadata) can embed `GlobalKeyMap` or not — Plan 1 author's call.
- 5 stateless-state keymaps (`FileListSearchKeyMap`, `RecipientConfirmKeyMap`, `BulkReKeyConfirmKeyMap`, `RecipientListKeyMap`, `FormatMenuKeyMap`): same shape. `RecipientConfirmKeyMap` and `BulkReKeyConfirmKeyMap` embed `GlobalKeyMap` and add a `HiddenFromMenu()` method returning `[]key.Binding{k.Quit}` (D-313).
- `DetailKeyMap` gains a `HiddenFromMenu()` method (D-307): `func (k DetailKeyMap) HiddenFromMenu() []key.Binding { return []key.Binding{k.Blame} }`.
- Package-level interface in `bindings.go`: `type menuVisibilityOverrider interface { HiddenFromMenu() []key.Binding }` — unexported, accessible from `internal/app` drift test via same-package access in `internal/keys`.
- `FileListKeyMap.ShortHelp()` gains `k.GoTop` and `k.GoBottom` appended after `k.Quit` (D-304, to preserve `hints[10]=="g"` and `hints[11]=="G"` index assertions in `TestFileListHints`).

**Cross-file impact:** Changing `FileListKeyMap.ShortHelp()` to return 12 bindings directly enables `FileListModel.Hints()` to drop the manual append. The existing `TestHintsFromBindings_RealFileListKeyMap` in `internal/keys/hints_test.go` line 61 asserts `len == 10` and will need updating to 12.

---

### `internal/keys/hints.go` — delete 5 inline package vars

**Analog:** `internal/keys/hints.go` — the 5 package vars at lines 56–115 (these are the deleted items; the file shrinks to ~50 lines retaining `MenuHint`, `Hinter`, `HintsFromBindings`)

**What stays** (lines 1–51):
```go
type MenuHint struct {
    Mnemonic    string
    Description string
    Visible     bool
}

type Hinter interface {
    Hints() []MenuHint
}

func HintsFromBindings(bindings []key.Binding) []MenuHint {
    hints := make([]MenuHint, 0, len(bindings))
    for _, b := range bindings {
        h := b.Help()
        hints = append(hints, MenuHint{
            Mnemonic:    h.Key,
            Description: h.Desc,
            Visible:     true,
        })
    }
    return hints
}
```

**What is deleted:** `FileListSearchHints`, `RecipientConfirmHints`, `BulkReKeyConfirmHints`, `RecipientListHints`, `FormatMenuHints` — their data moves to `WithHelp()` strings on the 5 new stateless-state keymap types in `bindings.go`. The quit-suppression doc-comment at lines 65–73 migrates to a shared block on `RecipientConfirmKeyMap` and `BulkReKeyConfirmKeyMap` in `bindings.go`.

**Cross-file impact:** Deleting these vars causes compile failures in `internal/app/model.go` (5 arms), `internal/app/hints_test.go` (4 assertions), and `internal/keys/hints_test.go` (5 exact-copy tests). All must be updated in the same commit.

---

### `internal/ui/filelist.go` — simplify `Hints()`

**Analog:** `internal/ui/filelist.go:383–390` (self — current implementation becoming simpler)

**Current** (lines 383–390):
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

**After D-304:** One line: `return keys.HintsFromBindings(m.keys.ShortHelp())`. The `g`/`G` entries are now in `FileListKeyMap.ShortHelp()` at positions 10 and 11 (after `Quit`), so the test assertions on `hints[10].Mnemonic == "g"` and `hints[11].Mnemonic == "G"` stay green.

---

### `internal/ui/detail.go` — move `Visible=false` override to `HiddenFromMenu()`

**Analog:** `internal/ui/detail.go:820–830` (self — current post-hoc loop replaced by cleaner pattern)

**Current** (lines 820–830):
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

**After D-307:** The inline mnemonic-search loop is replaced by applying `HiddenFromMenu()` from the keymap. The new `Hints()` pattern:
```go
func (m DetailModel) Hints() []keys.MenuHint {
    hints := keys.HintsFromBindings(m.keys.ShortHelp())
    for _, hidden := range m.keys.HiddenFromMenu() {
        for i := range hints {
            if hints[i].Mnemonic == hidden.Help().Key {
                hints[i].Visible = false
                break
            }
        }
    }
    return hints
}
```
Or equivalently, the suppression can apply the index directly since the binding position in `ShortHelp()` is known — Plan 1 author's choice. The mnemonic-based loop is safer against reordering.

---

### `internal/ui/help.go` — refactor `Hints()` to derive from `HelpKeyMap`

**Analog:** `internal/ui/health.go:181–189` — same literal-return shape being replaced

**Current** (help.go lines 96–102):
```go
func (m HelpModel) Hints() []keys.MenuHint {
    return []keys.MenuHint{
        {Mnemonic: "Esc", Description: "close help", Visible: true},
        {Mnemonic: "?",   Description: "close help", Visible: true},
        {Mnemonic: "q",   Description: "quit",        Visible: true},
    }
}
```

**Constructor** (help.go line 42): `func NewHelpModel(width, height int) HelpModel`

**After Plan 1 (D-301 pattern):**
- Add `keys HelpKeyMap` field to `HelpModel` struct.
- Initialize to `DefaultHelpKeyMap` in `NewHelpModel()` — no constructor signature change.
- Replace `Hints()` body: `return keys.HintsFromBindings(m.keys.ShortHelp())`.
- The `WithHelp` strings in `DefaultHelpKeyMap` must exactly match the deleted literal strings: `"Esc"/"close help"`, `"?"/"close help"`, `"q"/"quit"`.

**Pattern to copy from `health.go` constructor** (health.go lines 36–50):
```go
func NewHealthModel(width, height int) HealthModel {
    return HealthModel{
        loading: true,
        width:   width,
        height:  height,
    }
}
```
Add `keys: DefaultHelpKeyMap` to the struct literal — zero call-site impact.

---

### `internal/ui/diff.go` — refactor `Hints()` to derive from `DiffKeyMap`

**Analog:** `internal/ui/diff.go:176–185` (self — literal return replaced)

**Current** (lines 176–185):
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

**Constructor** (diff.go line 46): `func NewDiffModel(title string, entries []DiffEntry, width, height int) DiffModel`

**After Plan 1:** Add `keys DiffKeyMap` field; initialize to `DefaultDiffKeyMap`; replace `Hints()` body with `return keys.HintsFromBindings(m.keys.ShortHelp())`. `WithHelp` strings must match the 6 literal descriptions verbatim.

---

### `internal/ui/health.go` — refactor `Hints()` to derive from `HealthKeyMap`

**Analog:** `internal/ui/history.go:133–141` — structural twin (same 5-binding pattern)

**Current** (health.go lines 181–189):
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

**Constructor** (health.go line 36): `func NewHealthModel(width, height int) HealthModel`

**After Plan 1:** Add `keys HealthKeyMap` field; initialize to `DefaultHealthKeyMap`; `Hints()` becomes one line. `WithHelp` strings must match verbatim: `"j"/"scroll down"`, `"k"/"scroll up"`, `"H"/"close health"`, `"Esc"/"close health"`, `"q"/"quit"`.

---

### `internal/ui/history.go` — refactor `Hints()` to derive from `HistoryKeyMap`

**Analog:** `internal/ui/health.go:181–189` — structural twin

**Current** (history.go lines 133–141):
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

**Constructor** (history.go line 33): `func NewHistoryModel(filename string, width, height int) HistoryModel`

**After Plan 1:** Same pattern. `WithHelp` strings: `"j"/"scroll down"`, `"k"/"scroll up"`, `"b"/"close history"`, `"Esc"/"close history"`, `"q"/"quit"`.

---

### `internal/ui/metadata.go` — refactor `Hints()` to derive from `MetadataKeyMap`

**Analog:** `internal/ui/history.go:133–141` — structural twin

**Current** (metadata.go lines 167–175):
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

**Constructor** (metadata.go line 42): `func NewMetadataModel(meta MetadataContent, width, height int) MetadataModel`

**After Plan 1:** Same pattern. `WithHelp` strings: `"j"/"scroll down"`, `"k"/"scroll up"`, `"i"/"close metadata"`, `"Esc"/"close metadata"`, `"q"/"quit"`.

---

### `internal/ui/recipientform.go` — refactor `Hints()` to derive from `RecipientFormKeyMap`

**Analog:** `internal/ui/help.go:96–102` — same 2-entry literal return shape

**Current** (recipientform.go lines 163–168):
```go
func (m RecipientFormModel) Hints() []keys.MenuHint {
    return []keys.MenuHint{
        {Mnemonic: "Enter", Description: "confirm", Visible: true},
        {Mnemonic: "Esc",   Description: "cancel",  Visible: true},
    }
}
```

**Constructor** (recipientform.go line 41): `func NewRecipientFormModel(width, height int) RecipientFormModel`

**After Plan 1:** Add `keys RecipientFormKeyMap` field; initialize to `DefaultRecipientFormKeyMap`. `WithHelp` strings: `"Enter"/"confirm"`, `"Esc"/"cancel"`. Note: `key.WithKeys("enter")` (lowercase) produces `key.WithHelp("Enter", "confirm")` (capitalized display) — this is existing convention in the codebase (see DetailKeyMap.Back: `key.WithKeys("esc")` + `key.WithHelp("esc", ...)`).

---

### `internal/app/model.go` — update 5 `menuHints()` stateless arms

**Analog:** `internal/app/model.go:1498–1533` — the switch arms being updated

**Current arms** (lines 1498–1531):
```go
func (m AppModel) menuHints() []keys.MenuHint {
    if m.state == stateFileList && m.fileList.IsSearchActive() {
        return keys.FileListSearchHints           // D-11 override
    }
    switch m.state {
    // ...
    case stateRecipientConfirm:
        return keys.RecipientConfirmHints
    case stateBulkReKeyConfirm:
        return keys.BulkReKeyConfirmHints
    // ...
    case stateRecipientList:
        return keys.RecipientListHints
    case stateFormatMenu:
        return keys.FormatMenuHints
    }
    return nil
}
```

**After Plan 1 — 5 arms become:**
```go
    if m.state == stateFileList && m.fileList.IsSearchActive() {
        return keys.HintsFromBindings(keys.DefaultFileListSearchKeyMap.ShortHelp())
    }
    // ...
    case stateRecipientConfirm:
        return keys.HintsFromBindings(keys.DefaultRecipientConfirmKeyMap.ShortHelp())
    case stateBulkReKeyConfirm:
        return keys.HintsFromBindings(keys.DefaultBulkReKeyConfirmKeyMap.ShortHelp())
    // ...
    case stateRecipientList:
        return keys.HintsFromBindings(keys.DefaultRecipientListKeyMap.ShortHelp())
    case stateFormatMenu:
        return keys.HintsFromBindings(keys.DefaultFormatMenuKeyMap.ShortHelp())
```

**Plan 2 (D-309):** Change comment at line 1492 from `(state, recipientAction, IsSearchActive)` to `(state, IsSearchActive)`.

---

### `internal/app/hints_test.go` — update 4 equality assertions

**Analog:** `internal/app/hints_test.go:91–152` — the 4 existing assertions that reference deleted package vars

**Current assertions** (lines 91–152):
```go
// line 95
require.Equal(t, keys.RecipientConfirmHints, hints)
// line 103
require.Equal(t, keys.BulkReKeyConfirmHints, hints)
// line 143
require.Equal(t, keys.RecipientListHints, hints)
// line 151
require.Equal(t, keys.FormatMenuHints, hints)
```

**After Plan 1 — replace with keymap-derived form:**
```go
require.Equal(t, keys.HintsFromBindings(keys.DefaultRecipientConfirmKeyMap.ShortHelp()), hints)
require.Equal(t, keys.HintsFromBindings(keys.DefaultBulkReKeyConfirmKeyMap.ShortHelp()), hints)
require.Equal(t, keys.HintsFromBindings(keys.DefaultRecipientListKeyMap.ShortHelp()), hints)
require.Equal(t, keys.HintsFromBindings(keys.DefaultFormatMenuKeyMap.ShortHelp()), hints)
```

**Note:** The search-active test at line 60 (`require.Equal(t, keys.FileListSearchHints, hints)`) also references a deleted var and must be updated to `keys.HintsFromBindings(keys.DefaultFileListSearchKeyMap.ShortHelp())`.

---

### `internal/keys/bindings_test.go` — add 11 compile-time `help.KeyMap` checks

**Analog:** `internal/keys/bindings_test.go:67–82` — existing `var _ help.KeyMap = FileListKeyMap{}` pattern

**Current pattern** (lines 66–82):
```go
// TestFileListKeyMap_ImplementsHelpKeyMap verifies FileListKeyMap implements help.KeyMap.
func TestFileListKeyMap_ImplementsHelpKeyMap(t *testing.T) {
    var _ help.KeyMap = keys.FileListKeyMap{}
    short := keys.DefaultFileListKeyMap.ShortHelp()
    assert.NotEmpty(t, short, "ShortHelp must return at least one binding")
    full := keys.DefaultFileListKeyMap.FullHelp()
    assert.NotEmpty(t, full, "FullHelp must return at least one group")
}
```

**After Plan 1 — add 11 analogous functions**, one per new keymap type:
```go
func TestHelpKeyMap_ImplementsHelpKeyMap(t *testing.T) {
    var _ help.KeyMap = keys.HelpKeyMap{}
    assert.NotEmpty(t, keys.DefaultHelpKeyMap.ShortHelp())
    assert.NotEmpty(t, keys.DefaultHelpKeyMap.FullHelp())
}
// ... 10 more identical blocks for Diff, Health, History, Metadata, RecipientForm,
//     FileListSearch, RecipientConfirm, BulkReKeyConfirm, RecipientList, FormatMenu
```

The file is `package keys_test` — it imports `charm.land/bubbles/v2/help` (already imported at line 7).

---

### `internal/app/menuhints_drift_test.go` (new — Plan 2)

**Analog:** `internal/app/hints_test.go` — same `buildAppModel` driver + `menuHints()` call pattern; same `package app` internal test file (chrome_test.go is also `package app`)

**Construction pattern** (from hints_test.go lines 20–25):
```go
func buildAppModel(t *testing.T) AppModel {
    t.Helper()
    m := NewAppModel(defaultEnvInternal(), "")
    updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
    return updated.(AppModel)
}
```

**Sub-model construction pattern** (real constructors, as recommended by D-307):
```go
// For data-only models:
health := ui.NewHealthModel(80, 24)
hints := health.Hints()
km := keys.DefaultHealthKeyMap
expected := keys.HintsFromBindings(km.ShortHelp())
require.Equal(t, expected, hints)

// For models with HiddenFromMenu():
detail := ui.NewDetailModel(/* ... */)
km := keys.DefaultDetailKeyMap
expected := keys.HintsFromBindings(km.ShortHelp())
// Apply HiddenFromMenu() suppression:
for _, hidden := range km.HiddenFromMenu() {
    for i := range expected {
        if expected[i].Mnemonic == hidden.Help().Key {
            expected[i].Visible = false
            break
        }
    }
}
require.Equal(t, expected, detail.Hints())
```

**Stateless state pattern** (from hints_test.go lines 91–96):
```go
m := buildAppModel(t)
m.state = stateRecipientConfirm
hints := m.menuHints()
require.Equal(t, keys.HintsFromBindings(keys.DefaultRecipientConfirmKeyMap.ShortHelp()), hints)
```

**Note:** This file is `package app` (internal, same-package access to unexported `sessionState` constants and `menuHints()`). The chrome_test.go `TestViewNoNewStyle` test is also `package app` (not `package app_test`) — same pattern applies here.

---

### `internal/app/testdata/menu_*.golden` (13 new — Plan 2)

**Analog:** `internal/app/testdata/resize_80x24.golden` — same `RequireGoldenStructure` family; same `testdata/` directory

**Golden generation driver pattern** (from resize_test.go lines 40–50):
```go
func TestResize_80x24(t *testing.T) {
    setDeterministicAgeEnv(t)
    m := app.NewAppModel(defaultEnv(), "")
    updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
    m = updated.(app.AppModel)
    v := m.View()
    testutil.RequireGoldenStructure(t, "resize_80x24", v.Content)
}
```

**Menu golden driver pattern** (Plan 2 equivalent):
```go
func TestMenuGolden_StateFileList(t *testing.T) {
    m := buildAppModel(t)
    // m.state already defaults to stateFileList
    hints := m.menuHints()
    rendered := ui.RenderMenu(hints, 80)
    testutil.RequireGoldenStructure(t, "menu_file_list", rendered)
}
```

**13 golden file names** (lowercase + underscore per CONTEXT.md §Claude's Discretion):
1. `menu_file_list.golden`
2. `menu_detail.golden`
3. `menu_metadata.golden`
4. `menu_diff.golden`
5. `menu_help.golden`
6. `menu_history.golden`
7. `menu_health.golden`
8. `menu_recipient_list.golden`
9. `menu_recipient_form.golden`
10. `menu_format_menu.golden`
11. `menu_recipient_confirm.golden`
12. `menu_bulk_re_key_confirm.golden`
13. `menu_file_list_search.golden`

**Generation command:** `GOLDEN_UPDATE=1 go test ./internal/app/... -run TestMenuGolden`

**Key difference from resize goldens:** Menu goldens capture `ui.RenderMenu(hints, 80)` only — not `v.Content` (full chrome). This means `setDeterministicAgeEnv(t)` is not needed (age fingerprint is not in the menu).

---

## Shared Patterns

### `keys.HintsFromBindings(km.ShortHelp())` — total derivation
**Source:** `internal/keys/hints.go:40–51` + `internal/keys/bindings.go` any ShortHelp() method
**Apply to:** All 6 refactored sub-model `Hints()` implementations + all 5 stateless dispatcher arms + search-active override

```go
// HintsFromBindings converts a slice of key.Binding into MenuHint entries.
func HintsFromBindings(bindings []key.Binding) []MenuHint {
    hints := make([]MenuHint, 0, len(bindings))
    for _, b := range bindings {
        h := b.Help()
        hints = append(hints, MenuHint{
            Mnemonic:    h.Key,
            Description: h.Desc,
            Visible:     true,
        })
    }
    return hints
}
```

### `Default*KeyMap` zero-call-site constructor injection
**Source:** `internal/keys/bindings.go:90–140` — `DefaultFileListKeyMap` instance
**Apply to:** All 6 new sub-model constructors adding a `keys XxxKeyMap` field

```go
// Pattern: add keys field, initialize in constructor, no signature change.
func NewHealthModel(width, height int) HealthModel {
    return HealthModel{
        keys:    DefaultHealthKeyMap,  // new field
        loading: true,
        width:   width,
        height:  height,
    }
}
```

### `var _ help.KeyMap = XxxKeyMap{}` compile-time contract
**Source:** `internal/keys/bindings_test.go:67–82` — existing FileList/Detail checks
**Apply to:** `internal/keys/bindings_test.go` — 11 new checks, one per new keymap type

```go
func TestHelpKeyMap_ImplementsHelpKeyMap(t *testing.T) {
    var _ help.KeyMap = keys.HelpKeyMap{}
    assert.NotEmpty(t, keys.DefaultHelpKeyMap.ShortHelp())
    assert.NotEmpty(t, keys.DefaultHelpKeyMap.FullHelp())
}
```

### `buildAppModel` + state-set + `menuHints()` dispatcher test
**Source:** `internal/app/hints_test.go:20–25` — `buildAppModel` helper
**Apply to:** `internal/app/menuhints_drift_test.go` — all 13 stateless-state drift sub-tests

```go
func buildAppModel(t *testing.T) AppModel {
    t.Helper()
    m := NewAppModel(defaultEnvInternal(), "")
    updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
    return updated.(AppModel)
}
// Usage:
m := buildAppModel(t)
m.state = stateHealth
hints := m.menuHints()
require.Equal(t, keys.HintsFromBindings(keys.DefaultHealthKeyMap.ShortHelp()), hints)
```

### `RequireGoldenStructure` golden test
**Source:** `internal/testutil/golden.go:30–56` + `internal/app/resize_test.go:40–50`
**Apply to:** `internal/app/menuhints_drift_test.go` `TestMenuGolden_*` sub-tests

```go
// Generation: GOLDEN_UPDATE=1 go test ./internal/app/... -run TestMenuGolden
// Comparison: testutil.RequireGoldenStructure(t, "menu_file_list", rendered)
func RequireGoldenStructure(t *testing.T, name, output string)
// name -> testdata/<name>.golden; ANSI-stripped; LF-normalised
```

---

## No Analog Found

All files have close analogs. No entries in this section.

---

## Metadata

**Analog search scope:** `internal/keys/`, `internal/ui/`, `internal/app/`, `internal/testutil/`
**Files scanned:** 14 source files + 8 test files
**Pattern extraction date:** 2026-04-30

### Description-string lock table (critical for test stability)

Plan 1 must use these exact `WithHelp` description strings or existing `TestXxxHints` tests will fail:

| Model | Binding | `WithHelp` key | `WithHelp` desc |
|---|---|---|---|
| `HelpKeyMap` | Close | `"Esc"` | `"close help"` |
| `HelpKeyMap` | ToggleHelp | `"?"` | `"close help"` |
| `HelpKeyMap` | Quit | `"q"` | `"quit"` |
| `DiffKeyMap` | Confirm | `"y"` | `"confirm re-encrypt"` |
| `DiffKeyMap` | Cancel | `"n"` | `"cancel"` |
| `DiffKeyMap` | Close | `"Esc"` | `"cancel"` |
| `DiffKeyMap` | ScrollDown | `"j"` | `"scroll down"` |
| `DiffKeyMap` | ScrollUp | `"k"` | `"scroll up"` |
| `DiffKeyMap` | Quit | `"q"` | `"quit"` |
| `HealthKeyMap` | ScrollDown | `"j"` | `"scroll down"` |
| `HealthKeyMap` | ScrollUp | `"k"` | `"scroll up"` |
| `HealthKeyMap` | Close | `"H"` | `"close health"` |
| `HealthKeyMap` | CloseAlt | `"Esc"` | `"close health"` |
| `HealthKeyMap` | Quit | `"q"` | `"quit"` |
| `HistoryKeyMap` | ScrollDown | `"j"` | `"scroll down"` |
| `HistoryKeyMap` | ScrollUp | `"k"` | `"scroll up"` |
| `HistoryKeyMap` | Close | `"b"` | `"close history"` |
| `HistoryKeyMap` | CloseAlt | `"Esc"` | `"close history"` |
| `HistoryKeyMap` | Quit | `"q"` | `"quit"` |
| `MetadataKeyMap` | ScrollDown | `"j"` | `"scroll down"` |
| `MetadataKeyMap` | ScrollUp | `"k"` | `"scroll up"` |
| `MetadataKeyMap` | Close | `"i"` | `"close metadata"` |
| `MetadataKeyMap` | CloseAlt | `"Esc"` | `"close metadata"` |
| `MetadataKeyMap` | Quit | `"q"` | `"quit"` |
| `RecipientFormKeyMap` | Confirm | `"Enter"` | `"confirm"` |
| `RecipientFormKeyMap` | Cancel | `"Esc"` | `"cancel"` |
| `FileListSearchKeyMap` | ExitSearch | `"Esc"` | `"exit search"` |
| `FileListSearchKeyMap` | Select | `"Enter"` | `"select result"` |
| `FileListSearchKeyMap` | NextResult | `"j/↓"` | `"next result"` |
| `FileListSearchKeyMap` | PrevResult | `"k/↑"` | `"prev result"` |
| `FileListSearchKeyMap` | ToggleHelp | `"?"` | `"toggle help"` |
| `FileListSearchKeyMap` | Quit | `"q"` | `"quit"` |
| `RecipientConfirmKeyMap` | Confirm | `"y"` | `"confirm add/remove recipient"` |
| `RecipientConfirmKeyMap` | Cancel | `"n"` | `"cancel"` |
| `RecipientConfirmKeyMap` | Abort | `"Esc"` | `"cancel"` |
| `RecipientConfirmKeyMap` | ScrollDown | `"j"` | `"scroll down"` |
| `RecipientConfirmKeyMap` | ScrollUp | `"k"` | `"scroll up"` |
| `BulkReKeyConfirmKeyMap` | Confirm | `"y"` | `"confirm re-key this file"` |
| `BulkReKeyConfirmKeyMap` | Cancel | `"n"` | `"skip this file"` |
| `BulkReKeyConfirmKeyMap` | Abort | `"Esc"` | `"abort bulk re-key"` |
| `BulkReKeyConfirmKeyMap` | ScrollDown | `"j"` | `"scroll down"` |
| `BulkReKeyConfirmKeyMap` | ScrollUp | `"k"` | `"scroll up"` |
| `RecipientListKeyMap` | Select | `"1-9"` | `"select recipient to remove"` |
| `RecipientListKeyMap` | Cancel | `"Esc"` | `"cancel"` |
| `RecipientListKeyMap` | Quit | `"q"` | `"quit"` |
| `FormatMenuKeyMap` | Next | `"j"` | `"next format"` |
| `FormatMenuKeyMap` | Prev | `"k"` | `"prev format"` |
| `FormatMenuKeyMap` | Confirm | `"Enter"` | `"confirm format"` |
| `FormatMenuKeyMap` | Cancel | `"Esc"` | `"cancel"` |

### Test files that must be updated in the same commit as package-var deletion

| Test file | Line(s) | Change |
|---|---|---|
| `internal/app/hints_test.go` | 60, 95, 103, 143, 151 | Replace `keys.XxxHints` with `keys.HintsFromBindings(keys.DefaultXxxKeyMap.ShortHelp())` |
| `internal/keys/hints_test.go` | 76–134 | 5 `TestXxxHints_ExactCopy` tests — delete or rewrite to assert against keymap-derived output |
| `internal/keys/hints_test.go` | 61 | `TestHintsFromBindings_RealFileListKeyMap` asserts `len == 10` — update to 12 after D-304 |
