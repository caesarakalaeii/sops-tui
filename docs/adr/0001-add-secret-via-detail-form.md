# ADR 0001: Add a new secret via a detail-view form

- Status: Accepted
- Date: 2026-06-20

## Context

sops-tui could reveal, edit, rotate, and re-key existing secrets, but had no way
to create a *new* key/value in a file from the TUI — users had to drop to the
`sops` CLI or `$EDITOR`. We need an in-TUI "add a new secret" action.

The codebase is a Bubble Tea root model (`AppModel`) with a `sessionState` enum
routing to sub-models, plus a test-locked keybinding/menu system:

- The persistent menu has a hard 12-slot cap; surplus bindings stay discoverable
  in the `?` overlay via each keymap's `HiddenFromMenu()`.
- `DetailKeyMap.ShortHelp()` already filled all 12 visible slots (13 bindings,
  Blame hidden).
- Hint sets are derived totally from `keymap.ShortHelp()` (D-301), and a drift
  detector + golden menu snapshots lock the rendered output.

## Decision

Add the action as a new modal flow that reuses existing machinery rather than
inventing a new write path:

1. **Binding** — `n` ("add secret") on `DetailKeyMap`, available only in
   `stateDetail` when search is inactive. It is added to `ShortHelp()` (so it
   appears in `?`) and to `HiddenFromMenu()` (so the 12-slot persistent menu is
   unchanged — `n` joins Blame as discoverable-but-not-pinned).
2. **State + sub-model** — a new `stateAddSecretForm` and an
   `AddSecretFormModel` (two `textinput`s: key path + value, Tab to switch,
   Enter to confirm, Esc to cancel). It mirrors `RecipientFormModel`: inner
   content only (outer border from `AppModel.View()`), package-var styles only,
   client-side validation before any subprocess.
3. **Validation** — reject empty key paths, array-index notation (`sops set`
   cannot target it), and keys that already exist (steer the user to `e` to edit
   so an add never silently overwrites).
4. **Write path** — on confirm, route the new `{keyPath, "" → value}` through
   the existing diff-confirm overlay (`stateDiff`) and `sops.SetKey`. `sops set`
   creates the key when absent (verified end-to-end), so no new executor code is
   needed.
5. **Tree refresh** — because the new key is not yet in the detail tree, a
   successful `ReEncryptDoneMsg` for an add (tracked via `addedSecretKeyPath`)
   triggers a re-parse so the key surfaces, alongside the usual git-status
   refresh.

## Consequences

- No new subprocess surface: add and inline-edit share `sops.SetKey`.
- The persistent menu and its golden snapshots are unchanged (visible count
  stays 12); only the hint-count assertions for the detail keymap move 13 → 14.
- The drift detector keeps passing because `n` is consistently present in both
  `ShortHelp()` and `HiddenFromMenu()`.
- An e2e test (`internal/sops`, skips when `sops`/`age` are absent) locks the
  "SetKey creates a new key" assumption that this flow depends on.
- Future "add secret from the file list" (without opening the file) is possible
  but out of scope; it would need a target-file picker.
