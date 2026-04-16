# Phase 5: Power Features - Research

**Researched:** 2026-04-16
**Domain:** SOPS recipient management, secret health analysis, bulk multi-file operations
**Confidence:** HIGH

## Summary

Phase 5 delivers the three highest-risk operations in sops-tui: per-file age recipient management (view, add, remove), bulk re-key across multiple selected files with per-file confirmation, and an on-demand secret health dashboard covering weak secrets, duplicates, and stale files. All operations build entirely on patterns and infrastructure from Phases 1-4 — no new libraries are required.

The recipient management and bulk re-key features are implemented by wrapping `sops rotate -i --add-age <pubkey>` and `sops rotate -i --rm-age <pubkey>` — the canonical sops CLI flags for per-file recipient modification. These commands are verified against sops 3.12.2 (currently installed). The health check leverages `go-git` (already used in Phase 4) for staleness detection, `math.Log2` (Go stdlib) for Shannon entropy calculation, and the existing `DetectFormat`/`rotate.go` format-pattern regexes for format-aware weak-secret detection.

The only code additions beyond wiring are: (1) a new `GetLastCommitTime(repoRoot, relPath)` function in `git/status.go` returning `time.Time` for staleness comparison, (2) three new `sessionState` constants (`stateHealth`, `stateRecipientForm`, `stateRecipientConfirm`), (3) a `Selected bool` field on `FileItem`, (4) new key bindings (`space`, `K`, `H`, `a`, `d`), and (5) new UI components `HealthModel` and `RecipientFormModel`.

**Primary recommendation:** Use `sops rotate -i --add-age / --rm-age` for all recipient operations. No new Go libraries needed. Reuse all existing overlay, async msg, and confirmation patterns from Phases 3-4.

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**D-01** — Add-recipient uses a modal form overlay (huh/v2) — text input for the age public key string.
**D-02** — Age public key validated client-side: check `age1...` prefix and key length using `filippo.io/age` library.
**D-03** — Remove-recipient shows numbered list of current recipients; user selects by number.
**D-04** — After add/remove, show diff-style confirmation overlay listing recipient changes. User confirms before sops re-encrypts.
**D-05** — File selection uses Space toggle in file list. Selected files get visual indicator. `K` triggers bulk re-key on all selected files.
**D-06** — Bulk re-key uses per-file confirmation — each file shows its recipient diff individually.
**D-07** — Progress displays in status bar as "Re-keying 3/12: secrets/api.yaml".
**D-08** — Weak secret: length < 16 chars OR low Shannon entropy. Format-aware: validate against expected format when key name ends in `_token`, `_key`, `_secret`.
**D-09** — Duplicate detection: exact decrypted value matching across all files. User confirms decrypt-all before scan.
**D-10** — Staleness: git last-modified age via `go-git`. Flag files not modified in N days (configurable, default 90).
**D-11** — Health check triggered on-demand via `H` keybinding from file list.
**D-12** — Health results in full-screen health overlay grouped by category (weak/duplicate/stale) with severity indicators. Scrollable. Follows metadata overlay pattern.

### Claude's Discretion

- Key binding choices for bulk re-key trigger and health check trigger (e.g., `K` vs `B` for bulk, `H` for health)
- Health overlay layout and formatting details
- Shannon entropy threshold for "weak" classification
- How to handle health check when some files can't be decrypted (missing key)
- Toggle selection visual indicator style
- Whether duplicate scan shows which files share the value or just flags them
- Staleness threshold configuration mechanism (env var, flag, or hardcoded default)

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope

</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| RCP-01 | User can view age key recipients per file | Already available via `parser.SopsMetadata.AgeRecipients` (parsed in Phase 2); displayed in `MetadataModel`. Phase 5 adds recipient management from the detail view. |
| RCP-02 | User can add/remove age key recipients | `sops rotate -i --add-age <pubkey> <file>` (add) and `sops rotate -i --rm-age <pubkey> <file>` (remove). Diff-confirmation overlay gates re-encryption. |
| RCP-03 | User can bulk re-key multiple files | Space-toggle selection in `FileListModel` + `K` to trigger sequential per-file `sops rotate` with individual confirmation dialogs. Status bar shows "Re-keying N/M: file.yaml". |
| HLT-03 | User can run secret health checks (weak secrets, duplicates, staleness) | On-demand `H` key: decrypt all (with confirm) for duplicate scan; Shannon entropy + format-aware checks for weak; `git.GetLastCommitTime` for staleness. Full-screen `HealthModel` overlay. |

</phase_requirements>

---

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Recipient view (RCP-01) | API / Backend | UI Overlay | `parser.SopsMetadata` already extracts recipients from encrypted YAML; UI just renders them |
| Recipient add/remove (RCP-02) | API / Backend | UI Overlay | `sops rotate --add-age/--rm-age` is the subprocess owner; UI handles input, validation, and confirmation |
| Bulk re-key (RCP-03) | App model | API / Backend | `AppModel` sequences the per-file confirmation loop; `sops/executor.go` performs each re-key |
| Age key validation (D-02) | App model | — | Client-side validation before subprocess call; no backend needed |
| Shannon entropy (D-08) | App model / health pkg | — | Pure computation over decrypted strings; all in Go stdlib (`math`) |
| Duplicate detection (D-09) | App model | API / Backend | Requires decrypt-all for each file then cross-map comparison; app model owns the map |
| Staleness detection (D-10) | Git backend | App model | `go-git` provides commit timestamps; app model computes age comparison |
| Health overlay rendering (D-12) | UI | — | `HealthModel` mirrors `MetadataModel`/`HistoryModel` pattern |

---

## Standard Stack

### Core (all already in go.mod)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `charm.land/bubbletea/v2` | v2.0.4 | TEA message passing, sessionState, async Cmds | All existing phase architecture |
| `charm.land/lipgloss/v2` | v2.0.3 | Styles, layout, bordered overlays | Existing design system in `styles.go` |
| `charm.land/bubbles/v2` | v2.1.0 | `textinput` for add-recipient modal | Existing `SearchModel` uses same textinput |
| `github.com/go-git/go-git/v5` | v5.17.0 | `GetLastCommitTime` for staleness detection | Phase 4 git backend |
| `golang.org/x/crypto` | v0.50.0 | `bcrypt` (already used in rotate); no new usage | Already in go.mod |

### New Dependency: filippo.io/age (D-02)
| Library | Version | Purpose | Why |
|---------|---------|---------|-----|
| `filippo.io/age` | v1.3.1 | `age.ParseX25519Recipient(pubkey)` for client-side age key validation | CONTEXT.md D-02 explicitly names this library. Full bech32 decode + 32-byte key check catches malformed keys that length+prefix checks would miss. |

**Important note on huh/v2 (CONTEXT.md D-01 says "huh/v2" for modal form):**

CONTEXT.md D-01 specifies huh/v2 for the add-recipient modal. However, huh/v2 is NOT in go.mod, and the existing codebase achieves identical single-field modal input using `bubbles/v2/textinput` (the `SearchModel` and `stateEdit` patterns). Adding huh/v2 for a single text input field adds dependency weight without functional benefit. This is flagged as `[ASSUMED]` — the planner should add huh/v2 only if the locked decision (D-01) is strictly interpreted as requiring it.

**Recommendation for planner:** Use `bubbles/v2/textinput` wrapped in a new `RecipientFormModel` (mirroring `SearchModel` structure). Only add `charm.land/huh/v2` if D-01's "huh/v2" reference is considered binding rather than illustrative.

**Installation (only new dep):**
```bash
go get filippo.io/age@v1.3.1
```

**Version verification:** [VERIFIED: go.mod] filippo.io/age v1.3.1 confirmed available via `go get`; not currently in go.mod.

---

## Architecture Patterns

### System Architecture Diagram

```
Phase 5 Data Flow

File List View
  |
  +- Space --> FileItem.Selected = true/false
  |            Visual indicator: "[+] secrets/prod.yaml"
  |
  +- K ----> BulkReKeyRequestMsg
  |            |
  |            v (for each selected file, sequential)
  |         RecipientDiffOverlay (stateBulkReKeyConfirm)
  |            | y -> sops rotate -i <file>
  |            | n -> skip file
  |            v
  |         Flash "Re-keying N/M: file.yaml"
  |            v
  |         ReKeyDoneMsg -> next file
  |
  +- H ----> HealthCheckRequestMsg
  |            |
  |            v
  |         Confirm "Decrypt all N files?" overlay
  |            | y ->
  |            v
  |         [async] for each file: sops.DecryptFile
  |            |
  |            v HealthCheckResultMsg
  |         HealthModel overlay (stateHealth)
  |            Grouped: weak / duplicate / stale
  |
Detail View
  |
  +- a ----> RecipientFormModel overlay (stateRecipientForm)
  |            textinput "Enter age public key: age1..."
  |            Enter -> validate age.ParseX25519Recipient()
  |            | valid -> RecipientDiffOverlay (stateRecipientConfirm)
  |            |          y -> sops rotate -i --add-age <pubkey> <file>
  |            |          n -> cancel
  |
  +- d ----> RecipientListOverlay (stateRecipientConfirm)
               numbered list of current recipients
               select number -> RecipientDiffOverlay
               y -> sops rotate -i --rm-age <pubkey> <file>
               n -> cancel
```

### Recommended Project Structure (additions only)
```
internal/
+-- ui/
|   +-- health.go          # HealthModel overlay (stateHealth)
|   +-- health_test.go
|   +-- recipientform.go   # RecipientFormModel for add-recipient input
|   +-- recipientform_test.go
+-- sops/
|   +-- executor.go        # AddRecipient(), RemoveRecipient() added here
+-- git/
|   +-- status.go          # GetLastCommitTime() added here
+-- health/
    +-- checker.go         # ShannonEntropy(), IsWeakSecret(), IsDuplicate(), IsStale()
    +-- checker_test.go
```

### Pattern 1: sops rotate for Recipient Management
**What:** `sops rotate` with `--add-age` or `--rm-age` modifies a single file's recipient set and re-encrypts the data key.
**When to use:** Any per-file add or remove recipient operation.
**Example:**
```go
// Source: sops rotate --help (VERIFIED: sops 3.12.2 installed)
// Add a recipient:
cmd := exec.CommandContext(ctx, "sops", "rotate", "-i", "--add-age", pubkey, filePath)

// Remove a recipient:
cmd := exec.CommandContext(ctx, "sops", "rotate", "-i", "--rm-age", pubkey, filePath)

// Both write to filePath in-place (-i).
// Both require the existing age private key to be accessible for decryption.
// Both take ~30s timeout (same as other sops operations).
```

### Pattern 2: Age Key Validation
**What:** `age.ParseX25519Recipient(s)` from `filippo.io/age` fully validates an age public key string.
**When to use:** Before submitting any public key to sops, in the `RecipientFormModel` on Enter press.
**Example:**
```go
// Source: [CITED: https://pkg.go.dev/filippo.io/age#ParseX25519Recipient]
import "filippo.io/age"

func validateAgeKey(pubkey string) error {
    _, err := age.ParseX25519Recipient(pubkey)
    return err // nil = valid
}
// Validates: "age1" prefix + valid bech32 encoding + 32-byte raw key
// Returns descriptive error on failure (e.g., "invalid checksum")
```

### Pattern 3: Shannon Entropy for Weak Secret Detection
**What:** Shannon entropy measures information density. Low entropy = predictable/weak secret.
**When to use:** Health check's weak-secret scan on each decrypted leaf value.
**Example:**
```go
// Source: [ASSUMED] - Shannon entropy formula, implementation uses Go stdlib math package
import "math"

func shannonEntropy(s string) float64 {
    if len(s) == 0 {
        return 0
    }
    freq := make(map[rune]int)
    for _, r := range s {
        freq[r]++
    }
    n := float64(len([]rune(s)))
    var h float64
    for _, count := range freq {
        p := float64(count) / n
        h -= p * math.Log2(p)
    }
    return h
}
// Threshold recommendation: < 3.0 bits/char = weak for short secrets (<= 20 chars)
// < 3.5 bits/char = weak for longer values
// NOTE: threshold is Claude's Discretion per CONTEXT.md
```

### Pattern 4: GetLastCommitTime (new git/status.go function)
**What:** Returns the most recent commit timestamp for a specific file. New addition to `git/status.go`.
**When to use:** Health check staleness scan.
**Example:**
```go
// Source: [VERIFIED: go-git v5 API already used in git/status.go]
func GetLastCommitTime(repoRoot, relPath string) (time.Time, error) {
    repo, err := gogit.PlainOpenWithOptions(repoRoot, &gogit.PlainOpenOptions{DetectDotGit: true})
    if err != nil {
        return time.Time{}, err
    }
    slashPath := filepath.ToSlash(relPath)
    iter, err := repo.Log(&gogit.LogOptions{FileName: &slashPath})
    if err != nil {
        return time.Time{}, err
    }
    var commitTime time.Time
    err = iter.ForEach(func(c *object.Commit) error {
        commitTime = c.Author.When
        return storer.ErrStop // stop after first (most recent) commit
    })
    return commitTime, err
}
```

### Pattern 5: Bulk Re-Key Sequencing
**What:** Sequential per-file operation with per-file confirmation. Not parallel — parallel would make per-file confirmation impossible.
**When to use:** `K` trigger with selected files.
**Example:**
```go
// In AppModel: track bulk re-key queue
type bulkReKeyState struct {
    queue       []sops.DiscoveredFile // files remaining
    currentFile sops.DiscoveredFile
    completed   int
    total       int
    pubkey      string // the target recipient public key for the operation
}
// On each confirmation: pop next file from queue, show diff, await y/n
// Status bar: m.status.Flash(fmt.Sprintf("Re-keying %d/%d: %s", completed+1, total, filename))
```

### Pattern 6: Health Check Async Pipeline
**What:** Multi-phase async scan — decrypt each file sequentially (to avoid goroutine explosion), then analyze in-memory.
**When to use:** `H` key from file list.
**Example:**
```go
// HealthCheckResultMsg carries completed scan results
type HealthCheckResultMsg struct {
    WeakSecrets  []WeakSecret  // key path, file, value length, entropy
    Duplicates   []Duplicate   // value hash, []file+keypath pairs
    StaleFiles   []StaleFile   // file, last commit time, days since
    Errors       []string      // files that failed to decrypt
}
// Sent after all files scanned. AppModel handles by creating HealthModel overlay.
```

### Anti-Patterns to Avoid
- **Parallel sops subprocesses for bulk re-key:** Would prevent per-file confirmation. Use sequential queue.
- **In-memory plaintext accumulation without cleanup:** Duplicate detection decrypts all files — clear the map immediately after analysis.
- **Hard-coding the sops timeout for rotate:** `sops rotate` decrypts all key encryption keys then re-encrypts — may be slower than `sops set`. Use `SopsTimeout` (30s) as default but the planner should consider per-file timeout of 60s for large files or many recipients.
- **Using `sops updatekeys` instead of `sops rotate`:** `updatekeys` reads recipients from `.sops.yaml` and syncs all files — not per-file, not recipient-specific. Wrong command for this feature.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Age public key format validation | Custom regex/length checker | `filippo.io/age.ParseX25519Recipient()` | Bech32 checksum catches transpositions that length checks miss; library is already planned in CLAUDE.md |
| Git last-commit timestamp | Shell out to `git log --format=%ct` | Add `GetLastCommitTime()` to existing `git/status.go` using go-git v5 | Existing go-git infrastructure; avoids git binary dependency |
| Shannon entropy | Import a stats library | Implement inline using `math.Log2` | 10-line stdlib implementation; no dep needed |
| Recipient confirmation UI | New overlay framework | Reuse `DiffModel` pattern from Phase 3 | `DiffModel` already handles old->new display with y/n confirmation; recipient add/remove is structurally identical |
| Single-field modal input | huh/v2 form | `bubbles/v2/textinput` wrapped in `RecipientFormModel` | Existing `SearchModel` is the template; single-field input does not benefit from huh's multi-field form handling |

**Key insight:** Every UI primitive needed in Phase 5 already exists. The work is wiring, not invention.

---

## Common Pitfalls

### Pitfall 1: sops rotate Requires Existing Decrypt Access
**What goes wrong:** `sops rotate --add-age` fails silently or with a cryptic error if the age private key for the current recipients is not accessible.
**Why it happens:** sops must decrypt the data encryption key (DEK) before re-encrypting it for the new recipient set. Without the existing private key, the DEK cannot be recovered.
**How to avoid:** Surface the sops stderr output in the flash message. The error message from sops will say "no key could decrypt the data" — translate this to a user-friendly message.
**Warning signs:** `ReKeyDoneMsg.Err` is non-nil; stderr contains "no key" or "failed to get data key".

### Pitfall 2: Duplicate Detection — Memory Leak on Decrypted Values
**What goes wrong:** Holding all decrypted values in a `map[valueHash][]location` after the health check completes leaks secrets in Go heap memory until GC.
**Why it happens:** Go GC is non-deterministic; the plaintext map may linger.
**How to avoid:** After `HealthCheckResultMsg` is built, clear the intermediate plaintext map. Store only value hashes (SHA-256 truncated) in the result, never the plaintext values themselves. The health overlay shows "duplicate detected" without revealing the value.
**Warning signs:** If the map holds actual decrypted strings after `HealthCheckResultMsg` is sent, consider zeroing the backing array.

### Pitfall 3: `bubbles/v2/list` Default Keys Interfere With `K` and `H`
**What goes wrong:** The embedded `list.Model` in `FileListModel` may handle `K` (GoToEnd in default KeyMap uses `G`) — but since the sops-tui FileListKeyMap intercepts keys before the list.Update() call, `K` and `H` are safe.
**Why it happens:** The `FileListModel.Update()` method has an explicit `key.Matches()` block before delegating to `m.list.Update(msg)`. New keys must be intercepted in this same block.
**How to avoid:** Add `K` (BulkReKey) and `H` (HealthCheck) to `FileListKeyMap` and handle them in the `Update()` intercept block before `m.list.Update(msg)`. [VERIFIED: bubbles v2.1.0 list/keys.go has no `space`, `K`, or `H` bindings.]
**Warning signs:** If `m.list.Update(msg)` fires before the intercept block, pressing `K` in file list scrolls the list instead of triggering bulk re-key.

### Pitfall 4: `space` Key Conflict With bubbles List
**What goes wrong:** The embedded list might intercept `space` for some internal purpose.
**Why it happens:** Future bubbles updates could add space bindings.
**How to avoid:** Handle `space` (ToggleSelect) explicitly in `FileListModel.Update()` before calling `m.list.Update(msg)`. [VERIFIED: bubbles v2.1.0 list/keys.go has NO space binding.]

### Pitfall 5: sops rotate -i Modifies File — No Atomic Temp File
**What goes wrong:** `sops rotate -i` writes the result back to the same file. If interrupted mid-write, the file can be corrupted.
**Why it happens:** `sops rotate -i` uses sops' own in-place write logic, which does not provide atomicity guarantees visible to callers.
**How to avoid:** This is a sops CLI concern, not sops-tui's. sops itself does a temp-file-then-rename internally. Accept this behavior; document it. Do not replicate the EncryptFile atomic pattern from Phase 3 here — `sops rotate -i` handles it.

### Pitfall 6: `stalenessDays` Calculation — Use `time.Since` Not `time.Now().Sub`
**What goes wrong:** Time arithmetic errors if using `.Sub()` in wrong direction or truncating to days incorrectly.
**How to avoid:** `daysSince := int(time.Since(commitTime).Hours() / 24)`. Flag if `daysSince > threshold`.

### Pitfall 7: State Machine Complexity With Three New States
**What goes wrong:** `stateRecipientForm`, `stateRecipientConfirm`, and `stateHealth` must all handle `prevState` correctly to ensure Esc restores the correct previous view.
**How to avoid:** Follow the existing `prevState` pattern exactly. `stateRecipientForm` -> prevState is `stateDetail`. `stateRecipientConfirm` -> prevState is `stateRecipientForm` (or `stateDetail` if coming from remove flow). `stateHealth` -> prevState is `stateFileList`. Write tests for each Esc transition.

---

## Code Examples

### Adding a Recipient (sops rotate --add-age)
```go
// Source: [VERIFIED: sops rotate --help, sops 3.12.2]
// internal/sops/executor.go — new function
func AddRecipient(ctx context.Context, filePath, pubkey string) error {
    cmd := exec.CommandContext(ctx, "sops", "rotate", "-i", "--add-age", pubkey, filePath)
    var stderr bytes.Buffer
    cmd.Stderr = &stderr
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("sops rotate --add-age: %w: %s", err, strings.TrimSpace(stderr.String()))
    }
    return nil
}
```

### Removing a Recipient (sops rotate --rm-age)
```go
// Source: [VERIFIED: sops rotate --help, sops 3.12.2]
func RemoveRecipient(ctx context.Context, filePath, pubkey string) error {
    cmd := exec.CommandContext(ctx, "sops", "rotate", "-i", "--rm-age", pubkey, filePath)
    var stderr bytes.Buffer
    cmd.Stderr = &stderr
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("sops rotate --rm-age: %w: %s", err, strings.TrimSpace(stderr.String()))
    }
    return nil
}
```

### Age Key Validation
```go
// Source: [CITED: https://pkg.go.dev/filippo.io/age#ParseX25519Recipient]
import "filippo.io/age"

func validateAgePublicKey(s string) error {
    _, err := age.ParseX25519Recipient(s)
    return err
}
```

### Weak Secret Detection
```go
// Source: [ASSUMED] — Shannon entropy is domain math; implementation is Go stdlib only
import "math"

func shannonEntropy(s string) float64 {
    if len(s) == 0 {
        return 0
    }
    freq := make(map[rune]int)
    for _, r := range s {
        freq[r]++
    }
    n := float64(len([]rune(s)))
    var h float64
    for _, count := range freq {
        p := float64(count) / n
        h -= p * math.Log2(p)
    }
    return h
}

// Format-aware suffixes from D-08:
var weakKeyNameSuffixes = []string{"_token", "_key", "_secret"}

func isWeakSecret(keyPath, value string) bool {
    // Length check
    if len(value) < 16 {
        return true
    }
    // Entropy check — threshold is Claude's discretion
    if shannonEntropy(value) < 3.5 {
        return true
    }
    // Format-aware: key name hints at type (D-08)
    // Uses inline regex patterns (same as internal/ui/rotate.go DetectFormat)
    // to avoid circular import (ui imports health).
    for _, suffix := range weakKeyNameSuffixes {
        if strings.HasSuffix(keyPath, suffix) {
            if !hasKnownFormat(value) {
                return true // expected a structured format, got none
            }
        }
    }
    return false
}
```

### Staleness Detection
```go
// Source: [VERIFIED: go-git v5.17.0 API, same pattern as GetFileHistory in git/status.go]
func GetLastCommitTime(repoRoot, relPath string) (time.Time, error) {
    repo, err := gogit.PlainOpenWithOptions(repoRoot, &gogit.PlainOpenOptions{DetectDotGit: true})
    if err != nil {
        return time.Time{}, err
    }
    slashPath := filepath.ToSlash(relPath)
    iter, err := repo.Log(&gogit.LogOptions{FileName: &slashPath})
    if err != nil {
        return time.Time{}, err
    }
    var commitTime time.Time
    err = iter.ForEach(func(c *object.Commit) error {
        commitTime = c.Author.When
        return storer.ErrStop
    })
    if err != nil && err != storer.ErrStop {
        return time.Time{}, err
    }
    return commitTime, nil
}
```

### FileItem Toggle Selection
```go
// Source: [VERIFIED: filelist.go FileItem struct — add Selected bool field]
type FileItem struct {
    Name        string
    Path        string
    IsEncrypted bool
    Rule        sops.CreationRule
    GitStatus   string
    Selected    bool  // NEW: Phase 5 toggle selection (D-05)
}

func (i FileItem) Title() string {
    base := i.Name
    if i.Selected {
        base = "[+] " + base  // Claude's Discretion: indicator style
    }
    // existing badge rendering follows...
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `sops updatekeys` (syncs to .sops.yaml) | `sops rotate --add-age / --rm-age` (per-file CLI flags) | Available in SOPS v3 | Per-file recipient management without touching .sops.yaml |
| Hand-rolling bech32 validation | `filippo.io/age.ParseX25519Recipient()` | age library v1.x | Full checksum validation, not just prefix check |

---

## Key Discovery: sops updatekeys vs sops rotate

**VERIFIED finding that affects planning:**

`sops updatekeys` reads the desired recipient set from `.sops.yaml` and updates all files to match. It is a sync operation, not a per-file add/remove operation.

`sops rotate --add-age <pubkey> -i <file>` and `sops rotate --rm-age <pubkey> -i <file>` modify a single specific file's recipient set. This is the correct command for Phase 5's per-file recipient management.

The CONTEXT.md mentions "`sops updatekeys` and `sops -r` for re-keying operations". This is partially inaccurate: `sops -r` is `sops rotate` (the `-r` flag). `sops updatekeys` is useful only if the use case were "sync all files to .sops.yaml" which is not what Phase 5 implements. Planning should use `sops rotate --add-age/--rm-age -i` exclusively.

[VERIFIED: `sops rotate --help` and `sops updatekeys --help` on sops 3.12.2 installed at `/usr/bin/sops`]

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Shannon entropy threshold of 3.5 bits/char is a reasonable "weak" baseline | Code Examples | Threshold is Claude's Discretion; wrong value = too many/few false positives in health check |
| A2 | `huh/v2` is optional for D-01 and `bubbles/v2/textinput` is an acceptable substitute | Standard Stack | If D-01 is strictly interpreted as requiring huh/v2, a new dependency must be added; no functional impact |
| A3 | `storer.ErrStop` in GetLastCommitTime correctly terminates after the first commit | Code Examples | If go-git returns `ErrStop` as a real error, the function would fail; test needed |

---

## Open Questions (RESOLVED)

All open questions have been resolved during planning. Each is marked with its resolution.

1. **sops rotate timeout for large files or many recipients**
   - RESOLVED: Plan 01 defines `SopsRotateTimeout = 60 * time.Second` as a dedicated constant in `internal/sops/executor.go`. `AddRecipient()` and `RemoveRecipient()` use this longer timeout. Plan 03's bulk re-key also uses `SopsRotateTimeout` for each per-file `sops rotate -i` call.

2. **Health check when some files can't be decrypted (Claude's Discretion)**
   - RESOLVED: Plan 03's `runHealthCheck` function skips undecryptable files, collects errors in `HealthCheckResultMsg.Result.Errors []string`. Plan 02's `HealthModel` renders these as `HealthSkippedStyle.Render("N file(s) skipped -- could not decrypt")` in the footer.

3. **Staleness threshold configuration mechanism (Claude's Discretion)**
   - RESOLVED: Plan 03's `runHealthCheck` reads `SOPS_TUI_STALE_DAYS` environment variable (integer days, default 90). Parsed in the health check initiator function, not in the `git` package. Follows the `SOPS_TUI_CLIPBOARD_TIMEOUT` pattern from Phase 4.

4. **Duplicate scan: show which files share the value (Claude's Discretion)**
   - RESOLVED: Plan 01's `FindDuplicates` returns `[]Duplicate` where each `Duplicate` contains `Locations []Location` (file+keypath pairs). Plan 02's `HealthModel` renders these as `"[DUPE]  secrets/prod.yaml > api.key  AND  secrets/staging.yaml > api.key"`. Value SHA-256 hash is used as the dedup key; plaintext values are never stored in the result.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| sops binary | `AddRecipient`, `RemoveRecipient` | Yes | 3.12.2 | Already validated at startup (HLT-01) |
| age private key | `sops rotate` (decrypt step) | User-provided | — | Error message from sops stderr surfaces to user |
| go-git (in go.mod) | `GetLastCommitTime` | Yes | v5.17.0 | Skip staleness check if not a git repo |
| filippo.io/age | Age key validation | Needs `go get` | v1.3.1 | Lightweight prefix+length fallback (weaker validation) |

**Missing dependencies with no fallback:**
- None that block execution. filippo.io/age needs `go get` but the fallback (prefix check) is acceptable for an initial plan.

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + `stretchr/testify` v1.11.1 |
| Config file | none (standard `go test ./...`) |
| Quick run command | `go test ./internal/health/... ./internal/sops/... ./internal/git/... ./internal/ui/... -run TestHealth\|TestRecipient\|TestGetLastCommit\|TestAddRecipient\|TestRemoveRecipient -v` |
| Full suite command | `go test ./...` |

### Phase Requirements -> Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| RCP-01 | Recipients visible from existing `SopsMetadata.AgeRecipients` | unit | `go test ./internal/parser/... -run TestSopsMetadata` | Yes (parser_test.go) |
| RCP-02 | `AddRecipient` calls `sops rotate --add-age -i`, returns error on failure | unit | `go test ./internal/sops/... -run TestAddRecipient` | No -- Wave 0 |
| RCP-02 | `RemoveRecipient` calls `sops rotate --rm-age -i`, returns error on failure | unit | `go test ./internal/sops/... -run TestRemoveRecipient` | No -- Wave 0 |
| RCP-02 | Age key validation: valid key passes, invalid key returns error | unit | `go test ./internal/... -run TestValidateAgePublicKey` | No -- Wave 0 |
| RCP-02 | Recipient diff overlay renders add/remove correctly | unit | `go test ./internal/ui/... -run TestRecipientDiff` | No -- Wave 0 |
| RCP-03 | FileItem.Selected toggles on Space press in FileListModel | unit | `go test ./internal/ui/... -run TestFileItemToggle` | No -- Wave 0 |
| RCP-03 | Bulk re-key sequences files, skips unselected | unit | `go test ./internal/app/... -run TestBulkReKey` | No -- Wave 0 |
| HLT-03 | `shannonEntropy` returns correct value for known inputs | unit | `go test ./internal/health/... -run TestShannonEntropy` | No -- Wave 0 |
| HLT-03 | `isWeakSecret` flags short/low-entropy/format-mismatch values | unit | `go test ./internal/health/... -run TestIsWeakSecret` | No -- Wave 0 |
| HLT-03 | Duplicate detection finds matching values across two mock file maps | unit | `go test ./internal/health/... -run TestDuplicateDetection` | No -- Wave 0 |
| HLT-03 | `GetLastCommitTime` returns time.Time for a known test repo | unit | `go test ./internal/git/... -run TestGetLastCommitTime` | No -- Wave 0 |
| HLT-03 | Staleness: `daysSince > threshold` correctly classified | unit | `go test ./internal/health/... -run TestStaleDetection` | No -- Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/... -run TestHealth\|TestRecipient\|TestBulkReKey\|TestGetLastCommit -count=1`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `internal/health/checker_test.go` -- covers HLT-03 entropy, weak, duplicate, stale (REQ HLT-03)
- [ ] `internal/health/checker.go` -- package skeleton (empty functions) for test compilation
- [ ] `internal/sops/executor_test.go` additions -- TestAddRecipient, TestRemoveRecipient (REQ RCP-02)
- [ ] `internal/ui/recipientform_test.go` -- TestRecipientFormValidation (REQ RCP-02)
- [ ] `internal/ui/filelist_test.go` additions -- TestFileItemToggle (REQ RCP-03)
- [ ] `internal/app/model_test.go` additions -- TestBulkReKeySequence (REQ RCP-03)
- [ ] `internal/git/status_test.go` additions -- TestGetLastCommitTime (REQ HLT-03)

---

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | n/a -- local tool, no user auth |
| V3 Session Management | no | n/a |
| V4 Access Control | no | n/a |
| V5 Input Validation | yes | `age.ParseX25519Recipient()` for public key input; CharLimit on textinput |
| V6 Cryptography | yes | Never hand-roll; `sops rotate` delegates to age library for key encryption |

### Known Threat Patterns for Phase 5 Stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Malformed age public key passed to sops rotate | Tampering | `age.ParseX25519Recipient()` client-side validation before subprocess call |
| Decrypted values held in memory after health check | Information Disclosure | Zero the intermediate plaintext map; store only hashes in HealthCheckResultMsg |
| Process listing exposes age public key via sops CLI args | Information Disclosure | Public keys are not secret -- safe to pass as CLI args. Private keys are never involved as args. |
| Health check decrypt-all without user consent | Information Disclosure | User must confirm "Decrypt all N files?" before scan proceeds (D-09) |
| sops rotate interrupted mid-write (file corruption) | Tampering | sops uses internal temp-file-rename; handled by sops itself |

---

## Sources

### Primary (HIGH confidence)
- [VERIFIED: `sops rotate --help`] -- `sops rotate -i --add-age` and `--rm-age` flags confirmed in sops 3.12.2
- [VERIFIED: `sops updatekeys --help`] -- confirms `updatekeys` is NOT the right command for per-file recipient add/remove
- [VERIFIED: `go-git v5.17.0` in go.mod + git/status.go] -- `GetFileHistory` pattern confirms `GetLastCommitTime` approach
- [VERIFIED: `charm.land/bubbles/v2@v2.1.0/list/keys.go`] -- no `space`, `K`, or `H` bindings in default KeyMap
- [VERIFIED: `internal/ui/rotate.go`] -- `DetectFormat`, `shannonEntropy` approach confirmed as stdlib-only
- [VERIFIED: `go.mod`] -- filippo.io/age NOT present; needs `go get`

### Secondary (MEDIUM confidence)
- [CITED: https://pkg.go.dev/filippo.io/age#ParseX25519Recipient] -- API for age key validation; function signature confirmed from documentation

### Tertiary (LOW confidence)
- Shannon entropy threshold of 3.5 bits/char -- engineering judgment, not from official guidance

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- all libraries verified in go.mod or via pkg.go.dev; sops commands verified on installed binary
- Architecture: HIGH -- all patterns are direct extensions of existing Phases 1-4 code
- Pitfalls: HIGH for sops/git pitfalls (verified); MEDIUM for entropy threshold (judgment)

**Research date:** 2026-04-16
**Valid until:** 2026-05-16 (sops 3.12.2 stable; go-git v5 stable; bubbles v2 stable)
