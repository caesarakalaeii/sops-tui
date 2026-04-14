# Phase 3: Write Loop - Research

**Researched:** 2026-04-14
**Domain:** SOPS subprocess integration, inline editing TUI, diff overlay, secret rotation
**Confidence:** HIGH

## Summary

Phase 3 adds the write path to sops-tui: decrypt/reveal individual or all values, edit them inline or in `$EDITOR`, rotate to format-aware random values, and gate every destructive write behind a diff/confirmation overlay. The codebase already has a strong foundation — `sessionState` enum with overlay routing, `prevState` for overlay dismissal, `bubbles/textinput` used in `SearchModel`, and async `tea.Cmd` message patterns. Phase 3 follows every one of these patterns rather than inventing new ones.

The SOPS subprocess layer requires three distinct interactions: `sops decrypt --extract` for single-key reveal, `sops decrypt` (full pipe) for reveal-all, `sops set --value-stdin` for atomic single-key re-encryption, and `sops decrypt`/`sops encrypt` temp-file round-trip for the `$EDITOR` flow. The `--value-stdin` flag is confirmed present in sops 3.12.2 and avoids leaking secret values in process listings. All four operations must be non-blocking (`tea.Cmd` pattern).

The `$EDITOR` subprocess integration uses `tea.ExecProcess` — confirmed available in bubbletea v2.0.4 — which suspends the TUI, hands control to the editor, then resumes and delivers an `EditorFinishedMsg`. bcrypt rotation requires adding `golang.org/x/crypto v0.50.0` as a direct dependency (not currently in `go.mod`); all other random generation (base64, hex, UUID v4) uses `crypto/rand` from stdlib.

**Primary recommendation:** Build `internal/sops/executor.go` as the SOPS subprocess wrapper, `internal/ui/diff.go` as the diff overlay (modelled on `metadata.go`), and extend `DetailModel` + `AppModel` following the established state-machine patterns. Add two new `sessionState` values: `stateDiff` and `stateEdit`.

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**D-01:** Single value reveal uses `r` key — `r` on an encrypted leaf decrypts and reveals inline. `r` on a revealed leaf toggles it back to masked.

**D-02:** Reveal all uses `R` (Shift+R) — single `sops -d` call; all leaves update. `R` again re-masks all.

**D-03:** Revealed values display as plaintext with a lock-unlock icon suffix. `*** (type)` hint disappears; actual value shown in normal text.

**D-04:** Navigating back to file list (Esc from detail) auto-masks all revealed values. No decrypted content persists across view transitions.

**D-05:** Inline edit uses `e` key. On a revealed leaf → value cell becomes editable textinput. Enter confirms, Esc cancels. `e` on masked value = no-op (or flash "Reveal first with r").

**D-06:** Full-file edit uses `E` (Shift+E). Suspends TUI, opens decrypted file in `$EDITOR`. After save+quit TUI resumes and detects changes. Value must be revealed first.

**D-07:** Re-encryption for inline edits uses `sops set <file> '["key"]["path"]' '"new_value"'` — atomic single-key update, no temp files.

**D-08:** Re-encryption for `$EDITOR` flow uses standard `sops` encrypt on modified temp file, then replaces original.

**D-09:** Before re-encryption, a full-screen diff overlay appears. Shows old → new with color coding (red = removed, green = added). Same overlay pattern as help/metadata.

**D-10:** The diff overlay IS the confirmation gate. `y` = confirm + trigger re-encryption, `n`/Esc = cancel without effect.

**D-11:** For `$EDITOR` flow, after editor exits TUI compares old vs new. All changed keys in scrollable diff overlay. `y`/`n` confirms all at once.

**D-12:** Secret rotation uses `X` (Shift+X). On a revealed leaf, auto-detects format, generates random value, shows in diff overlay for confirmation before re-encryption.

**D-13:** Format auto-detection: base64 → base64, hex → hex, UUID pattern → UUID v4, bcrypt → bcrypt. Ambiguous → format selection menu.

**D-14:** After format selection, new value appears in standard diff overlay with y/n confirmation.

### Claude's Discretion

- Exact byte lengths for generated values (reasonable defaults: 32 bytes for base64/hex, standard for UUID/bcrypt)
- bcrypt cost factor (standard: 10-12)
- Inline text input styling and cursor behavior
- `$EDITOR` temp file location and naming convention
- Error handling for sops subprocess failures (flash messages with error details)
- Format detection heuristics (regex patterns for base64, hex, UUID, bcrypt)
- Scrollable diff overlay navigation keys (j/k for scrolling within diff)

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| DEC-01 | User can decrypt and reveal individual secret values on demand | `sops decrypt --extract '["key"]'` subprocess; `TreeNode.Revealed`+`TreeNode.DecryptedValue` fields; `r` key binding |
| DEC-02 | User can decrypt and reveal all values in a file | `sops decrypt` full-file subprocess; iterate nodes setting `Revealed=true`; `R` key binding |
| EDT-01 | User can edit a secret value with automatic re-encryption | `bubbles/textinput` inline edit; `sops set --value-stdin` for re-encryption; `stateEdit` session state |
| EDT-02 | User sees diff view before confirming re-encryption | `DiffModel` full-screen overlay; `stateDiff` session state; old/new value storage |
| EDT-03 | User can rotate a secret to a format-aware random value (base64, hex, UUID, bcrypt) | `crypto/rand` + `encoding/base64`/`encoding/hex` for base64/hex/UUID; `golang.org/x/crypto/bcrypt` for bcrypt rotation; format detection regexes |
| EDT-04 | User must confirm before any destructive write operation | Diff overlay `y`/`n` gate is the universal confirmation gate for all write paths |
</phase_requirements>

---

## Standard Stack

### Core (already in go.mod)
| Library | Version | Purpose | Why |
|---------|---------|---------|-----|
| charm.land/bubbletea/v2 | v2.0.4 | Root model, state routing, ExecProcess | Already in use; `tea.ExecProcess` confirmed for $EDITOR suspension [VERIFIED: /home/caesar/go/pkg/mod/charm.land/bubbletea/v2@v2.0.4/exec.go] |
| charm.land/bubbles/v2 | v2.1.0 | textinput for inline edit | Already in use for SearchModel; same pattern applies to inline value editor [VERIFIED: codebase grep] |
| charm.land/lipgloss/v2 | v2.0.3 | Diff color coding, overlay borders | ColorSuccess (green) and ColorError (red) already defined in styles.go [VERIFIED: internal/ui/styles.go] |
| golang.org/x/crypto | **v0.50.0** | `bcrypt` package for rotation | Not currently in go.mod — must add as direct dep. v0.50.0 is latest as of 2026-04-09. [VERIFIED: proxy.golang.org query] |

### Supporting (stdlib — no new deps)
| Package | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `crypto/rand` | stdlib | Random bytes for base64/hex/UUID generation | All random secret generation except bcrypt |
| `encoding/base64` | stdlib | Base64 encoding of random bytes | Base64 rotation format |
| `encoding/hex` | stdlib | Hex encoding of random bytes | Hex rotation format |
| `fmt.Sprintf` | stdlib | UUID v4 formatting from 16 random bytes | UUID rotation format |
| `os` | stdlib | `os.CreateTemp` for $EDITOR temp file, `os.Remove` cleanup | $EDITOR flow temp file lifecycle |
| `strings` | stdlib | Dot-path to sops index expression conversion | `"database.password"` → `'["database"]["password"]'` |
| `regexp` | stdlib | Format detection heuristics | Identifying base64/hex/UUID/bcrypt patterns |

### New Dependency
```bash
go get golang.org/x/crypto@v0.50.0
```

**Version verification:** [VERIFIED: curl https://proxy.golang.org/golang.org/x/crypto/@latest] — v0.50.0 published 2026-04-09.

## Architecture Patterns

### Recommended Project Structure (additions only)

```
internal/
├── sops/
│   ├── discoverer.go       # existing
│   └── executor.go         # NEW: sops subprocess wrapper (decrypt, set, encrypt)
├── ui/
│   ├── detail.go           # EXTEND: Revealed/DecryptedValue on TreeNode; r/R/e/E/X keybindings
│   ├── diff.go             # NEW: DiffModel full-screen overlay
│   ├── rotate.go           # NEW: format detection + random value generation
│   └── styles.go           # EXTEND: RevealedValue style, DiffAdded/DiffRemoved styles
├── app/
│   └── model.go            # EXTEND: stateDiff, stateEdit enum values; new message types
└── keys/
    └── bindings.go         # EXTEND: DetailKeyMap with Reveal, RevealAll, Edit, EditFile, Rotate
```

### Pattern 1: SOPS Executor (internal/sops/executor.go)

**What:** A wrapper around `exec.CommandContext` calls to the sops binary. Returns results via return values, never panics, always captures stderr for error surfacing.

**When to use:** Any time the TUI needs to call sops for decryption or re-encryption. Never inline exec calls in model code.

```go
// Source: CLAUDE.md §SOPS Integration + exec.go pattern from internal/sops/discoverer.go
package sops

import (
    "bytes"
    "context"
    "fmt"
    "os/exec"
    "strings"
    "time"
)

// DecryptKey decrypts a single key from a SOPS-encrypted file using --extract.
// keyPath is dot-notation: "database.password" -> ["database"]["password"]
// Returns the decrypted plaintext string value.
func DecryptKey(ctx context.Context, filePath, keyPath string) (string, error) {
    index := dotPathToIndex(keyPath)
    cmd := exec.CommandContext(ctx, "sops", "decrypt", "--extract", index, filePath)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    if err := cmd.Run(); err != nil {
        return "", fmt.Errorf("sops decrypt --extract %s: %w\nstderr: %s", keyPath, err, stderr.String())
    }
    return strings.TrimRight(stdout.String(), "\n"), nil
}

// DecryptFile decrypts an entire SOPS file and returns the plaintext YAML bytes.
func DecryptFile(ctx context.Context, filePath string) ([]byte, error) {
    cmd := exec.CommandContext(ctx, "sops", "decrypt", filePath)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("sops decrypt %s: %w\nstderr: %s", filePath, err, stderr.String())
    }
    return stdout.Bytes(), nil
}

// SetKey atomically re-encrypts a single key using sops set --value-stdin.
// Passes value via stdin to avoid leaking secrets in process listings.
func SetKey(ctx context.Context, filePath, keyPath, newValue string) error {
    index := dotPathToIndex(keyPath)
    cmd := exec.CommandContext(ctx, "sops", "set", "--value-stdin", filePath, index)
    cmd.Stdin = strings.NewReader(`"` + strings.ReplaceAll(newValue, `"`, `\"`) + `"`)
    var stderr bytes.Buffer
    cmd.Stderr = &stderr
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("sops set %s: %w\nstderr: %s", keyPath, err, stderr.String())
    }
    return nil
}

// dotPathToIndex converts "database.password" -> '["database"]["password"]'
func dotPathToIndex(dotPath string) string {
    parts := strings.Split(dotPath, ".")
    var sb strings.Builder
    for _, p := range parts {
        sb.WriteString(`["`)
        sb.WriteString(p)
        sb.WriteString(`"]`)
    }
    return sb.String()
}
```

[VERIFIED: sops set --help confirms --value-stdin flag; executor.go pattern mirrors discoverer.go]

### Pattern 2: sessionState Extension (app/model.go)

**What:** Two new enum values and corresponding message types following the exact pattern established for `stateHelp` and `stateMetadata`.

```go
// Source: internal/app/model.go existing pattern
const (
    stateFileList sessionState = iota
    stateDetail
    stateHelp
    stateMetadata
    stateDiff   // NEW: diff/confirmation overlay
    stateEdit   // NEW: inline value editing active in detail pane
)

// DecryptKeyMsg carries result of single-key decryption.
type DecryptKeyMsg struct {
    KeyPath string
    Value   string
    Err     error
}

// DecryptAllMsg carries result of full-file decryption.
type DecryptAllMsg struct {
    Values map[string]string // keyPath -> plaintext value
    Err    error
}

// ReEncryptDoneMsg signals re-encryption completed (success or failure).
type ReEncryptDoneMsg struct {
    Err error
}

// EditorFinishedMsg is the ExecCallback return value after $EDITOR exits.
type EditorFinishedMsg struct {
    Err error
}
```

[VERIFIED: mirrors existing FilesDiscoveredMsg / FilesParsedMsg pattern in model.go]

### Pattern 3: TreeNode Extension (ui/detail.go)

**What:** Two new fields on `TreeNode` for reveal state. The `renderRow` function branches on `node.Revealed`.

```go
// Source: internal/ui/detail.go TreeNode struct
type TreeNode struct {
    // ... existing fields unchanged ...
    
    // Revealed indicates this leaf has been decrypted and DecryptedValue is populated.
    // Only true after user presses r (single reveal) or R (reveal all).
    Revealed       bool
    // DecryptedValue holds the plaintext value when Revealed is true.
    // Must be cleared when navigating back to file list (D-04).
    DecryptedValue string
}
```

Render branch in `renderRow`:

```go
if node.Encrypted && node.Revealed {
    // Revealed: show plaintext + unlock icon (D-03)
    sb.WriteString(node.DecryptedValue)
    sb.WriteString("  ")
    sb.WriteString(RevealedBadge.Render("\U0001F513")) // 🔓
} else if node.Encrypted {
    // Masked: show *** (type)
    sb.WriteString(DimText.Render("***"))
    sb.WriteString(TypeHintStyle.Render(" (" + node.TypeHint + ")"))
}
```

[VERIFIED: renderRow function in internal/ui/detail.go]

### Pattern 4: DiffModel Overlay (ui/diff.go)

**What:** Full-screen overlay modelled on `MetadataModel`. Contains old/new value pairs, scroll support, `y`/`n` confirmation. Planner creates this as a new file.

```go
// Source: internal/ui/metadata.go MetadataModel (template pattern)
type DiffEntry struct {
    KeyPath  string
    OldValue string
    NewValue string
}

type DiffModel struct {
    entries  []DiffEntry
    title    string
    width    int
    height   int
    scroll   int
}
```

Styling uses existing design tokens:
- Removed lines: `lipgloss.NewStyle().Foreground(ColorError)` (red)
- Added lines: `lipgloss.NewStyle().Foreground(ColorSuccess)` (green)
- Unchanged context: `DimText`
- Footer: `"Press y to confirm, n/Esc to cancel"`

[VERIFIED: ColorSuccess=#a6e3a1 (green), ColorError=#f38ba8 (red) in internal/ui/styles.go]

### Pattern 5: $EDITOR Flow with tea.ExecProcess

**What:** Suspends TUI, opens decrypted temp file in `$EDITOR`, resumes on exit, diffs, shows overlay.

```go
// Source: /home/caesar/go/pkg/mod/charm.land/bubbletea/v2@v2.0.4/exec.go
func launchEditor(decryptedContent []byte, filePath string) tea.Cmd {
    tmpFile, err := os.CreateTemp("", "sops-tui-*.yaml")
    if err != nil {
        return func() tea.Msg { return EditorFinishedMsg{Err: err} }
    }
    tmpPath := tmpFile.Name()
    if _, err := tmpFile.Write(decryptedContent); err != nil {
        os.Remove(tmpPath)
        return func() tea.Msg { return EditorFinishedMsg{Err: err} }
    }
    tmpFile.Close()

    editor := os.Getenv("EDITOR")
    if editor == "" {
        editor = os.Getenv("VISUAL")
    }
    if editor == "" {
        editor = "vi" // last resort fallback
    }

    cmd := exec.Command(editor, tmpPath)
    return tea.ExecProcess(cmd, func(err error) tea.Msg {
        return EditorFinishedMsg{TmpPath: tmpPath, Err: err}
    })
}
```

[VERIFIED: tea.ExecProcess signature in exec.go; ExecCallback = func(error) Msg]

### Pattern 6: Format Detection and Random Value Generation (ui/rotate.go)

**What:** Regexes detect the current value's format; `crypto/rand` generates new values; `golang.org/x/crypto/bcrypt` handles bcrypt.

```go
// Source: CONTEXT.md D-13 + stdlib crypto/rand
import (
    "crypto/rand"
    "encoding/base64"
    "encoding/hex"
    "regexp"
    "golang.org/x/crypto/bcrypt"
)

var (
    reBase64 = regexp.MustCompile(`^[A-Za-z0-9+/]{22,}={0,2}$`)
    reHex    = regexp.MustCompile(`^[0-9a-fA-F]{32,}$`)
    reUUID   = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89aAbB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
    reBcrypt = regexp.MustCompile(`^\$2[ayb]\$\d{2}\$`)
)

type SecretFormat int

const (
    FormatUnknown    SecretFormat = iota
    FormatBase64
    FormatHex
    FormatUUID
    FormatBcrypt
    FormatAlphanumeric
)

func DetectFormat(value string) SecretFormat {
    switch {
    case reBcrypt.MatchString(value): return FormatBcrypt
    case reUUID.MatchString(value):   return FormatUUID
    case reHex.MatchString(value):    return FormatHex
    case reBase64.MatchString(value): return FormatBase64
    default:                          return FormatUnknown
    }
}

func GenerateValue(format SecretFormat) (string, error) {
    switch format {
    case FormatBase64:
        b := make([]byte, 32)
        if _, err := rand.Read(b); err != nil { return "", err }
        return base64.StdEncoding.EncodeToString(b), nil
    case FormatHex:
        b := make([]byte, 32)
        if _, err := rand.Read(b); err != nil { return "", err }
        return hex.EncodeToString(b), nil
    case FormatUUID:
        return generateUUIDv4()
    case FormatBcrypt:
        b := make([]byte, 32)
        if _, err := rand.Read(b); err != nil { return "", err }
        hash, err := bcrypt.GenerateFromPassword(b, bcrypt.DefaultCost) // cost=10
        if err != nil { return "", err }
        return string(hash), nil
    default:
        b := make([]byte, 24)
        if _, err := rand.Read(b); err != nil { return "", err }
        return base64.URLEncoding.EncodeToString(b), nil
    }
}

func generateUUIDv4() (string, error) {
    b := make([]byte, 16)
    if _, err := rand.Read(b); err != nil { return "", err }
    b[6] = (b[6] & 0x0f) | 0x40 // version 4
    b[8] = (b[8] & 0x3f) | 0x80 // variant bits
    return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
        b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
```

[VERIFIED: crypto/rand stdlib; golang.org/x/crypto/bcrypt.GenerateFromPassword confirmed in module cache at v0.50.0]

### Anti-Patterns to Avoid

- **Blocking model Update:** Never call `sops` synchronously inside `Update()`. Always return a `tea.Cmd` that runs in a goroutine and sends a message back.
- **Value in process listing:** Never pass secret values as command-line arguments to `sops set`. Use `--value-stdin` exclusively.
- **Leaking decrypted state:** When `m.state` transitions from `stateDetail` to `stateFileList` via Esc, iterate all `TreeNode` and set `Revealed=false`, `DecryptedValue=""` (D-04).
- **String-concatenating JSON for sops set:** `dotPathToIndex` must handle keys with special characters. The index format uses JSON string encoding: key containing `"` must be escaped.
- **Temp file not cleaned up:** The `EditorFinishedMsg` handler must `defer os.Remove(tmpPath)` after reading the result, even on error paths.
- **$EDITOR empty string:** Always fall through: `$EDITOR` → `$VISUAL` → `"vi"`. Never exec an empty string.
- **Edit on masked leaf:** `e` on a non-revealed leaf must be a no-op or flash "Reveal first with r" (D-05). Do not enter `stateEdit` unless `node.Revealed == true`.
- **Parallel async decryptions:** If user presses `r` rapidly on multiple keys, in-flight `DecryptKeyMsg` responses must match against `keyPath` to apply to the correct node — don't rely on cursor position at response time.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Bcrypt hash generation | Custom bcrypt implementation | `golang.org/x/crypto/bcrypt` | Bcrypt is security-sensitive; cost factor tuning and salt generation are non-trivial |
| UUID v4 | uuid library dependency | `crypto/rand` + bit manipulation (16 bytes) | UUID v4 is only 16 random bytes with 6 bits fixed; no library needed |
| TUI suspension for editor | Custom terminal restore/capture | `tea.ExecProcess` | Handles terminal mode save/restore, stdin/stdout plumbing automatically [VERIFIED: exec.go] |
| Diff algorithm | Myers diff | Two-value comparison (old vs new per key) | For single-key inline edit: trivially old/new. For $EDITOR flow: compare parsed key sets — no LCS algorithm needed for key-value diffs |
| Format detection false positive: bcrypt looks like base64 | Complex single-regex | Ordered detection: bcrypt first, then others | bcrypt always starts with `$2...`; checking it first prevents false base64 detection |

**Key insight:** The entire write safety model (D-09 through D-11) is one overlay with two fields (old, new) per key. This is not a Myers diff problem — it's a structured key-value before/after comparison.

## Common Pitfalls

### Pitfall 1: sops set --value-stdin requires JSON-encoded value
**What goes wrong:** Passing `newValue` directly via stdin results in `sops set` treating it as a raw YAML scalar, which may fail or re-encrypt with wrong type annotation.
**Why it happens:** The sops `set` command's `value` argument is JSON-encoded. A string must be `"quoted"`. An integer would be `42`. The value on stdin must be valid JSON.
**How to avoid:** Always wrap string values: `strings.NewReader(`"`" + jsonEscape(newValue) + `"`")`. Use `encoding/json.Marshal` for the value to get correct JSON encoding:
```go
import "encoding/json"
jsonVal, _ := json.Marshal(newValue) // produces "\"hello world\""
cmd.Stdin = bytes.NewReader(jsonVal)
```
**Warning signs:** `sops set` exits with error "unexpected end of JSON input" or re-encrypted value has wrong type.

### Pitfall 2: DecryptKey races with cursor movement
**What goes wrong:** User presses `r` on key A, cursor moves to key B, `DecryptKeyMsg{KeyPath: "A", Value: "..."}` arrives and is applied to whatever node is currently under the cursor.
**Why it happens:** Async response is applied positionally rather than by identity.
**How to avoid:** `DecryptKeyMsg` carries `KeyPath string`. The handler must walk `m.nodes` to find the node with matching `keyPath`, not use cursor index.
**Warning signs:** Wrong node shows decrypted value; nodes appear to swap.

### Pitfall 3: Temp file not cleaned up after $EDITOR flow
**What goes wrong:** Decrypted plaintext left on disk in `/tmp/sops-tui-*.yaml` after TUI exits.
**Why it happens:** Error paths before `defer os.Remove` are reached.
**How to avoid:** Set up `os.Remove(tmpPath)` immediately after creating the temp file (deferred in the `EditorFinishedMsg` handler), even if subsequent steps fail. The temp file should exist for the shortest possible duration.
**Warning signs:** Plaintext files accumulating in `$TMPDIR`.

### Pitfall 4: stateEdit conflicts with normal key routing
**What goes wrong:** While in `stateEdit` (inline textinput active), `j`/`k` navigation keys are consumed by the textinput and the cursor moves instead of accepting input.
**Why it happens:** Key routing sends all keys to textinput, but single-character keys look like navigation.
**How to avoid:** When `stateEdit` is active in `DetailModel`, route ALL key events to the textinput first. Only `Enter` (confirm) and `Esc` (cancel) have special meaning. The textinput eats everything else. This is the same approach as `searchActive` in the existing `DetailModel.Update`.
**Warning signs:** `j` inserts a "j" character instead of navigating; or cursor moves when it should be editing.

### Pitfall 5: sops decrypt --extract output includes trailing newline
**What goes wrong:** Decrypted value has `\n` appended, which when re-encrypted stores `"value\n"` instead of `"value"`.
**Why it happens:** `sops decrypt --extract` outputs value followed by newline (standard Unix behavior).
**How to avoid:** Always `strings.TrimRight(output, "\n")` before storing in `DecryptedValue`.
**Warning signs:** Re-encrypted value differs from original by one trailing newline; hash comparisons fail.

### Pitfall 6: $EDITOR diff compares YAML bytes, not key-value pairs
**What goes wrong:** YAML serialization differences (key ordering, quoting style) produce false diff entries when SOPS re-serializes the file.
**Why it happens:** `sops decrypt` may output YAML in different key order or quoting than the input.
**How to avoid:** After editor exit, parse both the original decrypted bytes and the new file bytes with `goccy/go-yaml`, then compare key-value maps. Do not byte-compare YAML.
**Warning signs:** Diff overlay shows changes for keys the user didn't touch; spurious re-encryptions.

### Pitfall 7: RevealAll then navigate-back leaves decrypted values in memory
**What goes wrong:** Decrypted values in `TreeNode.DecryptedValue` persist in `m.detail` even after navigating back to file list.
**Why it happens:** `AppModel` doesn't call a `ClearRevealed()` method on `m.detail` during Esc-to-file-list transition.
**Why it matters:** While Go GC will eventually free them, the data is visible to crash dumps and memory profilers until GC runs.
**How to avoid:** Add `(m *DetailModel) ClearAllRevealed()` method. Call it in `AppModel.Update` when transitioning `stateDetail → stateFileList` via Esc. This satisfies D-04.
**Warning signs:** Memory inspection shows plaintext values after navigation.

## Code Examples

### sops set with JSON-encoded value via stdin

```go
// Source: verified against sops 3.12.2 --help and --value-stdin behavior
import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "os/exec"
)

func SetKey(ctx context.Context, filePath, keyPath, newValue string) error {
    index := dotPathToIndex(keyPath)
    jsonVal, err := json.Marshal(newValue)
    if err != nil {
        return fmt.Errorf("json marshal value: %w", err)
    }
    cmd := exec.CommandContext(ctx, "sops", "set", "--value-stdin", filePath, index)
    cmd.Stdin = bytes.NewReader(jsonVal)
    var stderr bytes.Buffer
    cmd.Stderr = &stderr
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("sops set %s %s: %w\n%s", filePath, keyPath, err, stderr.String())
    }
    return nil
}
```

### Async decrypt single key (tea.Cmd pattern)

```go
// Source: mirrors FilesDiscoveredMsg pattern in internal/app/model.go
func decryptKeyCmd(ctx context.Context, filePath, keyPath string) tea.Cmd {
    return func() tea.Msg {
        value, err := sops.DecryptKey(ctx, filePath, keyPath)
        return DecryptKeyMsg{KeyPath: keyPath, Value: value, Err: err}
    }
}
```

### Auto-mask on navigation (D-04)

```go
// Source: existing Esc handler in internal/app/model.go (Priority 3 branch)
// Add before state transition:
if m.state == stateDetail {
    m.detail.ClearAllRevealed() // new method on DetailModel
    m.state = stateFileList
    m.status.SetBreadcrumb("files")
    m.status.SetItemCount(m.fileList.ItemCount(), "items")
    return m, nil
}
```

### textinput inline edit (stateEdit)

```go
// Source: internal/ui/search.go SearchModel pattern (already used)
// stateEdit is a mode on DetailModel, not a full sessionState overlay.
// Only the value cell under cursor becomes the textinput; rest of tree still renders.
// DetailModel gains:
type DetailModel struct {
    // ... existing fields ...
    editActive bool
    editInput  textinput.Model
    editKeyPath string  // dot path of the node being edited
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `sops edit <file>` opens editor directly | `tea.ExecProcess` wraps exec.Cmd, suspends TUI | bubbletea v2.0.4 (2026-04-13) | Clean TUI suspension/resume without manual terminal state management |
| `sops set <file> index "value"` (value in process listing) | `sops set --value-stdin <file> index` | sops 3.7+ | Secret value no longer visible in `ps aux` output |
| `math/rand` for random generation | `crypto/rand` | Go 1.20 (math/rand/v2 seeded by default) | `crypto/rand` is required for security-sensitive values regardless |

**Deprecated/outdated:**
- `sops set` with value as CLI argument: Still works but leaks value in process listing — use `--value-stdin` instead [VERIFIED: sops 3.12.2 --help]
- Manual terminal restore before exec: Replaced by `tea.ExecProcess` which handles restore internally [VERIFIED: exec.go source]

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `sops decrypt --extract '["key"]'` strips the ENC wrapper and outputs plaintext | Standard Stack | If extract output differs from expected format, DecryptKey returns wrong data type; tests will catch this |
| A2 | `sops set --value-stdin` accepts JSON-encoded string value on stdin | Code Examples | If value format differs, re-encryption fails; mitigated by `encoding/json.Marshal` which is always valid JSON |
| A3 | Format detection order (bcrypt first) correctly prevents bcrypt being misidentified as base64 | Architecture Patterns | Wrong format detection means wrong rotation type; deterministic tests will verify all four patterns |
| A4 | $EDITOR fallback chain `EDITOR → VISUAL → vi` works on target systems | Architecture Patterns | If vi is absent and both env vars unset, ExecProcess returns non-zero exit; caught by EditorFinishedMsg.Err |

**Most claims are VERIFIED:** sops binary is at `/usr/sbin/sops` (v3.12.2), bubbletea v2 source is in module cache, all existing code patterns are directly observed.

## Open Questions (RESOLVED)

1. **sops set behavior with array keys** (RESOLVED -- guarded in plans)
   - What we know: sops set index syntax supports `["key"][0]` for arrays
   - What's unclear: Our `dotPathToIndex` uses dot notation which doesn't encode array indices
   - Disposition: Phase 3 only targets string-keyed YAML maps (the common case). Array-indexed keys are blocked from edit/rotate with an explicit guard: `e`/`X` key handlers in detail.go detect keyPaths containing array index notation (e.g. `[0]`, `[1]`) and return `EditBlockedMsg` with message "Array-indexed keys not editable in Phase 3". Guard implemented in Plan 02 Task 1 (e key) and Plan 03 Task 2 (X key). Test case added in Plan 01 Task 1 (`TestDotPathToIndexArrayNotation`) and Plan 02 Task 1 (`TestEditOnArrayKeyReturnsBlocked`).

2. **SOPS stdin editing with encrypted_regex** (RESOLVED -- accepted with code comment)
   - What we know: STATE.md explicitly flags this as a known concern: "SOPS stdin editing with `encrypted_regex` -- behavior undocumented, may surface during Phase 3 implementation"
   - What's unclear: If a file uses `encrypted_regex`, does `sops set` re-apply the regex filter when re-encrypting? This could mean unencrypted-regex fields get re-encrypted unintentionally.
   - Disposition: Accepted as known limitation. `sops set` targets a specific key by path and preserves the rest of the file as-is, so the regex filter is not re-evaluated on other fields. A code comment in executor.go documents this assumption. A unit test in executor_test.go (`TestSetKeyEncryptedRegexComment`) validates that SetKey constructs correct arguments regardless of `encrypted_regex` configuration (sops CLI handles re-encryption scope internally). If edge cases surface during testing, they will be addressed as bugs. See Plan 01 Task 1.

3. **Context cancellation timeout for sops subprocess** (RESOLVED -- implemented in Plan 01)
   - What we know: CLAUDE.md mandates `exec.CommandContext` with `context.WithTimeout`
   - What's unclear: Appropriate timeout value for decrypt operations (depends on key size, network, age key derivation)
   - Disposition: 30 seconds constant `SopsTimeout` in `executor.go`. Age key derivation is fast; 30s is ample for any reasonable file size. Implemented and tested in Plan 01 Task 1.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| sops binary | DEC-01, DEC-02, EDT-01, EDT-03, EDT-04 | Yes | 3.12.2 | Already validated at startup by validator package |
| age key (~/.config/sops/age/keys.txt) | Actual decryption operations | [ASSUMED: present per prior phase tests] | — | Startup validator warns if missing |
| $EDITOR / $VISUAL / vi | EDT-01 ($EDITOR flow) | Not set in current shell; `vi` likely available | — | Fallback chain: EDITOR → VISUAL → vi |
| golang.org/x/crypto | EDT-03 (bcrypt rotation) | Not in go.mod — must add | v0.50.0 | No fallback — bcrypt rotation blocked without it |

**Missing dependencies with no fallback:**
- `golang.org/x/crypto` — must run `go get golang.org/x/crypto@v0.50.0` in Wave 0.

**Missing dependencies with fallback:**
- `$EDITOR` — fallback chain ensures editor always launches (though `vi` may not be user's preference).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib testing + testify v1.11.1 |
| Config file | none (go test ./...) |
| Quick run command | `go test ./internal/sops/... ./internal/ui/... -run TestPhase3 -v` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| DEC-01 | `sops.DecryptKey` converts dot path and calls sops | unit | `go test ./internal/sops/ -run TestDecryptKey -v` | Wave 0 |
| DEC-01 | `DetailModel.Update` with `r` key sets node `Revealed=true` | unit | `go test ./internal/ui/ -run TestDetailRevealKey -v` | Wave 0 |
| DEC-01 | `DecryptKeyMsg` applies to correct node by keyPath, not cursor | unit | `go test ./internal/ui/ -run TestDecryptKeyMsgRouting -v` | Wave 0 |
| DEC-02 | `sops.DecryptFile` returns plaintext YAML bytes | unit | `go test ./internal/sops/ -run TestDecryptFile -v` | Wave 0 |
| DEC-02 | `R` key triggers `DecryptAllMsg` and all leaves revealed | unit | `go test ./internal/ui/ -run TestDetailRevealAll -v` | Wave 0 |
| DEC-01/02 | Esc from detail clears all revealed values (D-04) | unit | `go test ./internal/app/ -run TestEscClearsRevealed -v` | Wave 0 |
| EDT-01 | `e` on masked leaf flashes "Reveal first" | unit | `go test ./internal/ui/ -run TestEditMaskedFlash -v` | Wave 0 |
| EDT-01 | `e` on revealed leaf enters stateEdit | unit | `go test ./internal/ui/ -run TestEditRevealedEntersEdit -v` | Wave 0 |
| EDT-01 | Enter in edit mode produces `SetKeyCmd` | unit | `go test ./internal/ui/ -run TestEditConfirmProducesCmd -v` | Wave 0 |
| EDT-01 | Esc in edit mode cancels without mutation | unit | `go test ./internal/ui/ -run TestEditCancelNoMutation -v` | Wave 0 |
| EDT-01 | `sops.SetKey` builds correct index and pipes JSON value | unit | `go test ./internal/sops/ -run TestSetKey -v` | Wave 0 |
| EDT-02 | `DiffModel.View()` shows old (red) and new (green) values | unit | `go test ./internal/ui/ -run TestDiffModelView -v` | Wave 0 |
| EDT-02 | `y` key in stateDiff triggers re-encrypt | unit | `go test ./internal/app/ -run TestDiffConfirm -v` | Wave 0 |
| EDT-02 | `n`/Esc in stateDiff cancels without calling sops | unit | `go test ./internal/app/ -run TestDiffCancel -v` | Wave 0 |
| EDT-03 | `DetectFormat` correctly identifies base64/hex/UUID/bcrypt | unit | `go test ./internal/ui/ -run TestDetectFormat -v` | Wave 0 |
| EDT-03 | `GenerateValue` returns correct format for each type | unit | `go test ./internal/ui/ -run TestGenerateValue -v` | Wave 0 |
| EDT-04 | All write paths (inline edit, $EDITOR, rotation) route through diff overlay | integration | `go test ./internal/app/ -run TestAllWritesRouteThroughDiff -v` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/sops/... ./internal/ui/... -run TestPhase3 -v`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `internal/sops/executor_test.go` — covers DEC-01, DEC-02, EDT-01 sops subprocess unit tests
- [ ] `internal/ui/diff_test.go` — covers EDT-02 DiffModel rendering and y/n key handling
- [ ] `internal/ui/rotate_test.go` — covers EDT-03 format detection and value generation
- [ ] `internal/app/model_test.go` (additions) — covers EDT-04 universal diff gate and stateDiff/stateEdit transitions
- [ ] `go get golang.org/x/crypto@v0.50.0` — required for bcrypt tests to compile

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | n/a |
| V3 Session Management | no | n/a |
| V4 Access Control | no | n/a |
| V5 Input Validation | yes | Key path validated: only alphanumeric + dot/underscore/hyphen accepted in dotPathToIndex |
| V6 Cryptography | yes | `crypto/rand` for all random generation; `golang.org/x/crypto/bcrypt` with DefaultCost (10) |

### Known Threat Patterns for This Stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Secret value in process listing | Information Disclosure | `sops set --value-stdin` exclusively — value never passed as CLI arg [VERIFIED: sops --help] |
| Plaintext temp file not cleaned | Information Disclosure | `defer os.Remove(tmpPath)` immediately after creation; use `os.CreateTemp` with mode 0600 |
| Command injection via key path | Tampering | `dotPathToIndex` builds index from parsed YAML keys (already sanitized by YAML parser), not user free-text |
| Decrypted values in memory after navigation | Information Disclosure | `ClearAllRevealed()` on Esc-to-file-list transition (D-04); no clipboard or logging of decrypted values |
| Math/rand for secret generation | Tampering | Use `crypto/rand` exclusively — never `math/rand` |
| Bcrypt cost too low | Tampering | Use `bcrypt.DefaultCost` (10); acceptable for generated secrets that will be stored encrypted |

## Sources

### Primary (HIGH confidence)
- `/home/caesar/go/pkg/mod/charm.land/bubbletea/v2@v2.0.4/exec.go` — `tea.ExecProcess` signature and behavior verified
- `/home/caesar/git/sops-tui/internal/app/model.go` — sessionState enum, prevState pattern, async Cmd pattern verified
- `/home/caesar/git/sops-tui/internal/ui/detail.go` — TreeNode struct, renderRow, flatRow.keyPath verified
- `/home/caesar/git/sops-tui/internal/ui/metadata.go` — Overlay pattern template for DiffModel
- `/home/caesar/git/sops-tui/internal/ui/styles.go` — ColorSuccess, ColorError values verified
- `sops 3.12.2 --help` (local binary) — `--value-stdin`, `--extract`, `set` subcommand syntax verified
- `https://proxy.golang.org/golang.org/x/crypto/@latest` — v0.50.0, 2026-04-09 [VERIFIED: curl]

### Secondary (MEDIUM confidence)
- `/home/caesar/go/pkg/mod/charm.land/bubbles/v2@v2.1.0/textinput/textinput.go` — Focus/Blur/SetValue API verified for inline edit
- `/home/caesar/go/pkg/mod/golang.org/x/crypto@v0.50.0/bcrypt/bcrypt.go` — bcrypt.GenerateFromPassword, DefaultCost confirmed in module cache

### Tertiary (LOW confidence)
- [ACCEPTED] sops set with `encrypted_regex` files preserves existing unencrypted keys — accepted as known limitation with code comment and test (see Open Questions #2 RESOLVED disposition)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all versions verified against local module cache and proxy
- Architecture: HIGH — all patterns derived from existing codebase (direct code reading)
- Pitfalls: HIGH — derived from code reading + known sops behavior verified against local binary
- Security: HIGH — crypto/rand and golang.org/x/crypto are stdlib/standard; sops --value-stdin confirmed

**Research date:** 2026-04-14
**Valid until:** 2026-05-14 (sops and bubbletea are stable; golang.org/x/crypto may update but bcrypt API is frozen)
