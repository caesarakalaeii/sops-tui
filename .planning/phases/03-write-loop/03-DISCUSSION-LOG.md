# Phase 3: Write Loop - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-14
**Phase:** 03-write-loop
**Areas discussed:** Reveal interaction, Edit flow, Diff & confirmation, Rotation formats

---

## Reveal Interaction

### Single value reveal (DEC-01)

| Option | Description | Selected |
|--------|-------------|----------|
| Enter/l on leaf | Same key that expands groups — context-aware: group node = expand, leaf node = decrypt | |
| Dedicated 'r' key | A new keybinding (r = reveal) on any leaf. Enter/l stays reserved for expand. More explicit. | ✓ |
| Spacebar toggle | Spacebar reveals/hides the selected value. Common in password managers. | |

**User's choice:** Dedicated 'r' key
**Notes:** User preferred explicit separation between expand and reveal actions.

### Reveal all values (DEC-02)

| Option | Description | Selected |
|--------|-------------|----------|
| Shift+R reveals all | Capital R decrypts all values with one sops -d call. Shift signals "bigger" action. | ✓ |
| Menu/command palette | A command palette with 'Decrypt all' option. More discoverable but heavier. | |
| Auto on file open | Prompt to decrypt when entering detail view. Proactive but may annoy browsers. | |

**User's choice:** Shift+R reveals all
**Notes:** None

### Hide behavior

| Option | Description | Selected |
|--------|-------------|----------|
| Same key toggles | r/R on revealed values hides them again. Esc from detail hides all. | ✓ |
| Explicit hide key | Separate key (H or Shift+R again) to re-mask. Revealed state persists. | |
| Auto-hide on navigation | Values auto-hide when cursor moves away. Maximum security. | |

**User's choice:** Same key toggles
**Notes:** None

### Visual state for revealed values

| Option | Description | Selected |
|--------|-------------|----------|
| Plaintext + unlock icon | Decrypted value in normal text with 🔓 suffix. Type hint disappears. | ✓ |
| Highlighted background | Distinct background color band on revealed values. | |
| Color change only | Accent color instead of dim. Subtle, minimal. | |

**User's choice:** Plaintext + unlock icon
**Notes:** None

---

## Edit Flow

### Single value editing (EDT-01)

| Option | Description | Selected |
|--------|-------------|----------|
| Inline text input | 'e' on leaf → editable text input. Enter confirms, Esc cancels. | ✓ |
| Modal form (huh v2) | Centered modal form with key name as label, value prefilled. | |
| External $EDITOR | Suspends TUI, opens full file in $EDITOR. | ✓ (via E) |

**User's choice:** Inline text input with `e`, plus `$EDITOR` with `E` (Shift+E)
**Notes:** User wanted both options: lowercase `e` for inline single-key editing, uppercase `E` for full-file $EDITOR.

### Edit requires reveal

| Option | Description | Selected |
|--------|-------------|----------|
| Yes, must reveal first | User must press 'r' to decrypt/reveal before 'e' works. Two deliberate actions. | ✓ |
| No, auto-reveal on edit | Pressing 'e' auto-decrypts and opens editor. One step. | |
| Configurable | Default require reveal, config flag to skip. | |

**User's choice:** Yes, must reveal first
**Notes:** None

### Re-encryption mechanism

| Option | Description | Selected |
|--------|-------------|----------|
| sops set command | `sops set <file> '["key"]' '"value"'` — atomic single-key update. | ✓ |
| Decrypt-modify-encrypt cycle | Full file decrypt → modify → re-encrypt. | |
| sops stdin pipe | Pipe modified YAML through sops. | |

**User's choice:** sops set command
**Notes:** None

---

## Diff & Confirmation

### Diff view style (EDT-02)

| Option | Description | Selected |
|--------|-------------|----------|
| Full-screen overlay | Same pattern as help/metadata. Old → new with red/green color coding. | ✓ |
| Inline confirmation banner | Compact banner below edited row. Less disruptive. | |
| Side-by-side split | Split terminal left=before, right=after. Breaks single-pane pattern. | |

**User's choice:** Full-screen overlay
**Notes:** None

### Confirmation gate (EDT-04)

| Option | Description | Selected |
|--------|-------------|----------|
| y/n prompt in diff overlay | Diff overlay IS the confirmation gate. y to confirm, n/Esc to cancel. | ✓ |
| Separate confirmation dialog | Second modal after diff. Two steps. | |
| Type-to-confirm | Must type key name to confirm. Maximum friction. | |

**User's choice:** y/n prompt in diff overlay
**Notes:** None

### $EDITOR diff handling

| Option | Description | Selected |
|--------|-------------|----------|
| Multi-key diff overlay after exit | Compare old vs new, show all changed keys in scrollable diff. y/n for all. | ✓ |
| No diff for $EDITOR | Trust user, re-encrypt immediately. Flash message with change count. | |
| Per-key confirmation | Walk through each changed key one by one. | |

**User's choice:** Multi-key diff overlay after exit
**Notes:** None

---

## Rotation Formats

### Format selection (EDT-03)

| Option | Description | Selected |
|--------|-------------|----------|
| Menu after keypress | Format selection menu appears after rotation key. | |
| Auto-detect from current value | Analyze current value to guess format. Menu fallback if ambiguous. | ✓ |
| Default with override | Default to base64, user can edit before confirming. | |

**User's choice:** Auto-detect from current value
**Notes:** None

### Rotation keybinding

| Option | Description | Selected |
|--------|-------------|----------|
| Shift+X | X = rotate/regenerate. Shift signals destructive action. | ✓ |
| Ctrl+R | R = rotate. Ctrl distinguishes from 'r' (reveal). | |
| Two-key sequence: r then x | Two deliberate keypresses. Vim-like sequence. | |

**User's choice:** Shift+X
**Notes:** None

### Ambiguous detection fallback

| Option | Description | Selected |
|--------|-------------|----------|
| Fall back to format menu | Show format selection menu when detection is ambiguous. | ✓ |
| Default to base64 | Always fall back to base64 (32 bytes). No menu. | |
| Inline prompt | Show inline text prompt with format codes. | |

**User's choice:** Fall back to format menu
**Notes:** None

---

## Claude's Discretion

- Exact byte lengths for generated rotation values
- bcrypt cost factor
- Inline text input styling
- $EDITOR temp file handling
- Error handling for sops failures
- Format detection heuristics
- Diff overlay scrolling navigation

## Deferred Ideas

None — discussion stayed within phase scope
