---
phase: 9
plan: 1
subsystem: keys + ui
tags: [refactor, keymaps, hints, discoverability]
dependency_graph:
  requires:
    - Phase 07 Plan 01 (HintsFromBindings, Hinter interface, MenuHint type)
    - Phase 07 Plan 03 (Hints() on all 8 sub-models, menuHints() dispatcher)
    - internal/keys/bindings.go (FileListKeyMap, DetailKeyMap patterns)
  provides:
    - 11 new keymap types in internal/keys/bindings.go
    - menuVisibilityOverrider interface
    - Keymap-derived Hints() on all 8 sub-models (single source of truth)
    - Simplified menuHints() dispatcher (5 stateless arms use Default*KeyMap)
  affects:
    - internal/keys/bindings.go (11 new types + FileListKeyMap.ShortHelp extension)
    - internal/keys/hints.go (shrunken to ~54 lines, 5 package vars deleted)
    - internal/ui/{help,diff,health,history,metadata,recipientform}.go (Hints() refactored)
    - internal/ui/filelist.go (Hints() simplified)
    - internal/ui/detail.go (Hints() uses HiddenFromMenu())
    - internal/app/model.go (dispatcher arms + NewAppModel initializations)
    - internal/keys/{bindings_test.go,hints_test.go} (11 new assertions + 5 updated)
    - internal/app/hints_test.go (5 equality assertions updated)
tech_stack:
  added: []
  patterns:
    - menuVisibilityOverrider unexported interface for per-keymap visibility suppression (D-307)
    - Default*KeyMap initialization in NewAppModel for all lazy sub-models
key_files:
  created: []
  modified:
    - internal/keys/bindings.go
    - internal/keys/hints.go
    - internal/keys/bindings_test.go
    - internal/keys/hints_test.go
    - internal/ui/help.go
    - internal/ui/diff.go
    - internal/ui/health.go
    - internal/ui/history.go
    - internal/ui/metadata.go
    - internal/ui/recipientform.go
    - internal/ui/filelist.go
    - internal/ui/detail.go
    - internal/app/model.go
    - internal/app/hints_test.go
decisions:
  - D-301 total derivation achieved: zero literal MenuHint slices remain in production code
  - D-304 GoTop/GoBottom moved into FileListKeyMap.ShortHelp() at positions [10],[11]
  - D-307 HiddenFromMenu() method pattern on DetailKeyMap, RecipientConfirmKeyMap, BulkReKeyConfirmKeyMap
  - D-309 amendment deferred to Plan 2 (recipientAction comment in model.go:1492 left intact)
  - Rule 1 fix: diff/history/metadata initialized in NewAppModel to prevent zero-value keymap Hints()
  - RecipientConfirmKeyMap/BulkReKeyConfirmKeyMap ShortHelp() includes Quit (6 hints), HiddenFromMenu() suppresses it for drift detector; dispatcher returns full 6-hint slice (menu rendering applies Visible filter)
metrics:
  duration: "10m 3s"
  completed_date: "2026-04-30"
  tasks: 6
  files: 14
---

# Phase 9 Plan 1: Keymap Extraction and Hints() Derivation Summary

**One-liner:** Extracted 11 `help.KeyMap`-implementing keymap types into `internal/keys/bindings.go`, refactored all 8 sub-model `Hints()` implementations to one-line keymap derivations, deleted 5 inline hint-set package vars from `hints.go`, and migrated the `menuHints()` dispatcher to use keymap-backed types — closing SC5 (single source of truth for menu content).

## What Was Built

### 11 New Keymap Types (internal/keys/bindings.go)

**6 sub-model keymaps:**
- `HelpKeyMap` (3 bindings: Close, ToggleHelp, Quit)
- `DiffKeyMap` (6 bindings: Confirm, Cancel, Close, ScrollDown, ScrollUp, Quit)
- `HealthKeyMap` (5 bindings: ScrollDown, ScrollUp, Close, CloseAlt, Quit)
- `HistoryKeyMap` (5 bindings: ScrollDown, ScrollUp, Close, CloseAlt, Quit)
- `MetadataKeyMap` (5 bindings: ScrollDown, ScrollUp, Close, CloseAlt, Quit)
- `RecipientFormKeyMap` (2 bindings: Confirm, Cancel)

**5 stateless-state keymaps:**
- `FileListSearchKeyMap` (6 bindings: ExitSearch, Select, NextResult, PrevResult, ToggleHelp, Quit)
- `RecipientConfirmKeyMap` (embeds GlobalKeyMap; Confirm, Cancel, Abort, ScrollDown, ScrollUp + HiddenFromMenu returns Quit)
- `BulkReKeyConfirmKeyMap` (embeds GlobalKeyMap; same shape, different descriptions + HiddenFromMenu returns Quit)
- `RecipientListKeyMap` (3 bindings: Select with "1-9" mnemonic, Cancel, Quit)
- `FormatMenuKeyMap` (4 bindings: Next, Prev, Confirm, Cancel — no Quit per OQ-3)

### menuVisibilityOverrider Interface

Unexported interface `menuVisibilityOverrider` defined in `bindings.go`. Implementers: `DetailKeyMap`, `RecipientConfirmKeyMap`, `BulkReKeyConfirmKeyMap`. Plan 2's drift detector will type-assert to apply suppression.

### FileListKeyMap.ShortHelp() Extension (D-304)

Extended to return 12 bindings: GoTop (g) at index 10, GoBottom (G) at index 11. `FileListModel.Hints()` is now a one-liner.

### hints.go Shrunk to ~54 Lines

Retains only `MenuHint`, `Hinter`, `HintsFromBindings`. Updated package doc to reflect Phase 9 D-301 total derivation.

### Dispatcher Migration (model.go menuHints())

5 stateless-state arms + search-active override now call `keys.HintsFromBindings(keys.DefaultXxxKeyMap.ShortHelp())`.

## Commits

| Hash | Type | Description |
|------|------|-------------|
| 7336206 | test | Wave 0 RED: 11 compile-time help.KeyMap assertions |
| 6215ca5 | feat | 6 sub-model keymap types + menuVisibilityOverrider + FileListKeyMap.ShortHelp |
| 2db02bc | feat | 5 stateless-state keymap types + quit-suppression doc-block |
| 09b8972 | refactor | 6 sub-model Hints() derive from keymaps via HintsFromBindings |
| 403335e | refactor | FileListModel.Hints() and DetailModel.Hints() simplified |
| b18cae4 | feat | Delete 5 inline vars + migrate dispatcher + update tests |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Zero-value DiffModel/HistoryModel/MetadataModel in NewAppModel returned empty Hints()**

- **Found during:** Task 6 — `TestMenuHints_StateDiff` failed with empty mnemonic string
- **Issue:** `AppModel.diff`, `AppModel.history`, and `AppModel.metadata` were not initialized in `NewAppModel`, leaving them as zero-value structs. The new `keys` field in each sub-model defaulted to a zero `DiffKeyMap`/etc. with empty bindings, so `Hints()` returned empty MenuHint slices.
- **Fix:** Added `diff: ui.NewDiffModel("", nil, 0, 0)`, `history: ui.NewHistoryModel("", 0, 0)`, `metadata: ui.NewMetadataModel(ui.MetadataContent{}, 0, 0)` to `NewAppModel` so the `keys` fields are initialized to their `Default*KeyMap` instances.
- **Files modified:** `internal/app/model.go`
- **Commit:** b18cae4

**2. [Rule 2 - Missing critical functionality] RecipientConfirmKeyMap and BulkReKeyConfirmKeyMap ShortHelp() include Quit**

- **Found during:** Task 6 design
- **Issue:** The old `RecipientConfirmHints` had 5 hints (no Quit). The new `RecipientConfirmKeyMap` embeds `GlobalKeyMap` (which includes Quit) and `ShortHelp()` explicitly includes `k.Quit` at position 5 so `HiddenFromMenu()` can suppress it in the drift detector (Plan 2). This means the dispatcher now returns 6 hints for these states vs the previous 5.
- **Behavior:** The `TestMenuHints_StateRecipientConfirm` assertion was updated to compare against `HintsFromBindings(DefaultRecipientConfirmKeyMap.ShortHelp())` (both sides derive from the same source), so the test passes. The menu renderer's `RenderMenu` checks `hint.Visible` — future Plan 2 drift detector will apply `HiddenFromMenu()` suppression to compare against the visible subset.
- **Notes:** This is intentional design per D-313: Quit is in `ShortHelp()` to be visible to the drift detector; its visibility is governed by `HiddenFromMenu()`.

## D-309 Amendment Note

Per CONTEXT.md D-309 and the plan instructions, the comment at `internal/app/model.go:1492` ("(state, recipientAction, IsSearchActive)") is intentionally left unchanged. Plan 2 owns this documentation cleanup.

## Threat Flags

None — Plan 1 is a pure structural refactor of compile-time-constant keybinding metadata. No new network endpoints, auth paths, file access patterns, or schema changes introduced.

## Self-Check: PASSED

Verified files exist:
- internal/keys/bindings.go: FOUND
- internal/keys/hints.go: FOUND (54 lines)
- internal/keys/bindings_test.go: FOUND
- internal/app/hints_test.go: FOUND
- internal/ui/{help,diff,health,history,metadata,recipientform,filelist,detail}.go: ALL FOUND
- internal/app/model.go: FOUND

Verified commits:
- 7336206 (test): FOUND
- 6215ca5 (feat): FOUND
- 2db02bc (feat): FOUND
- 09b8972 (refactor): FOUND
- 403335e (refactor): FOUND
- b18cae4 (feat): FOUND

Final suite: `go build ./... && go test ./... -count=1 && go vet ./...` — ALL CLEAN
