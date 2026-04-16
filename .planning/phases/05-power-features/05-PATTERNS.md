# Phase 5: Power Features - Pattern Map

**Mapped:** 2026-04-16
**Files analyzed:** 10 new/modified files
**Analogs found:** 10 / 10

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/ui/health.go` | component/overlay | request-response | `internal/ui/history.go` | exact (full-screen overlay, loading state, scrollable content) |
| `internal/ui/health_test.go` | test | — | `internal/git/status_test.go` | exact (table-driven, t.TempDir, testify) |
| `internal/ui/recipientform.go` | component/overlay | request-response | `internal/ui/search.go` | exact (single textinput, focus/blur, CharLimit, Update/View) |
| `internal/ui/recipientform_test.go` | test | — | `internal/ui/filelist_test.go` | role-match |
| `internal/sops/executor.go` | service | request-response | `internal/sops/executor.go` (existing) | extend-in-place |
| `internal/git/status.go` | service | request-response | `internal/git/status.go` (existing, GetFileHistory) | extend-in-place |
| `internal/health/checker.go` | utility | transform | `internal/ui/rotate.go` (pure computation, no deps) | role-match |
| `internal/health/checker_test.go` | test | — | `internal/sops/executor_test.go` | exact (table-driven, require/assert) |
| `internal/app/model.go` | app model | event-driven | `internal/app/model.go` (existing) | extend-in-place |
| `internal/keys/bindings.go` | config | — | `internal/keys/bindings.go` (existing) | extend-in-place |

---

## Pattern Assignments

### `internal/ui/health.go` (component, request-response)

**Analog:** `internal/ui/history.go`

This is the primary template. `HealthModel` mirrors `HistoryModel` exactly — loading state, `SetEntries`, `ScrollDown`/`ScrollUp`, full-screen bordered box. The difference is three grouped sections instead of a flat list.

**Imports pattern** (history.go lines 1–19):
```go
// Package ui provides the health check overlay component for sops-tui.
// ...
package ui

import (
    "strings"

    "charm.land/lipgloss/v2"
)
// Note: no tea import needed — HealthModel has no Update(); scrolling is driven
// by the parent AppModel exactly as HistoryModel and MetadataModel are scrolled.
```

**Struct pattern** (history.go lines 22–31):
```go
type HealthModel struct {
    results HealthCheckResult // replaces []CommitEntry
    loading bool
    width   int
    height  int
    scroll  int
}
```

**Constructor + SetEntries pattern** (history.go lines 34–48):
```go
func NewHealthModel(width, height int) HealthModel {
    return HealthModel{
        loading: true,
        width:   width,
        height:  height,
    }
}

func (m *HealthModel) SetResults(results HealthCheckResult) {
    m.results = results
    m.loading = false
}
```

**Scroll pattern** (history.go lines 56–66):
```go
func (m *HealthModel) ScrollDown() {
    lines := m.buildContentLines()
    maxScroll := len(lines) - 1
    if maxScroll < 0 { maxScroll = 0 }
    if m.scroll < maxScroll { m.scroll++ }
}
func (m *HealthModel) ScrollUp() {
    if m.scroll > 0 { m.scroll-- }
}
```

**Full-screen bordered box pattern** (history.go lines 133–141):
```go
return lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(ColorMuted).
    Background(ColorSurface).
    Padding(1, SpaceMD).
    Width(boxWidth).
    Height(boxHeight).
    Render(inner)
```

**Loading + empty state pattern** (history.go lines 88–97):
```go
if m.loading {
    loadingText := DimText.Render("Running health check...")
    inner = title + "\n\n" + loadingText
} else if m.results.IsEmpty() {
    emptyHeading := HelpSectionHeader.Render("No issues found")
    emptyBody := DimText.Render("All secrets passed health checks.")
    inner = title + "\n\n" + emptyHeading + "\n" + emptyBody + "\n\n" + footer
} else {
    // grouped sections: weak / duplicate / stale
}
```

**Section header style** — use `HelpSectionHeader` (bold) for "Weak Secrets", "Duplicates", "Stale Files" group headers, matching metadata.go line 92 and history.go line 80.

---

### `internal/ui/health_test.go` (test)

**Analog:** `internal/git/status_test.go` + `internal/sops/executor_test.go`

**Test file header pattern** (status_test.go lines 1–19):
```go
// Package ui tests — use _test suffix for black-box testing of HealthModel.
package ui_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/caesarakalaeii/sops-tui/internal/ui"
)
```

**Loading state test pattern** — verify `NewHealthModel().View()` contains loading text before `SetResults` is called. Mirrors history_test.go pattern.

**Table-driven test pattern** (executor_test.go lines 17–55):
```go
func TestHealthModel(t *testing.T) {
    tests := []struct {
        name     string
        results  ui.HealthCheckResult
        wantText string
    }{
        { name: "loading state", results: ui.HealthCheckResult{}, wantText: "Running health check" },
        // ...
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            m := ui.NewHealthModel(80, 24)
            if !tc.results.IsEmpty() {
                m.SetResults(tc.results)
            }
            view := m.View()
            assert.Contains(t, view, tc.wantText)
        })
    }
}
```

---

### `internal/ui/recipientform.go` (component, request-response)

**Analog:** `internal/ui/search.go`

`RecipientFormModel` is a modal overlay wrapping a single `textinput.Model`, identical in structure to `SearchModel` but rendered as a full-screen bordered box (like `DiffModel`) rather than an inline bar.

**Imports pattern** (search.go lines 1–21):
```go
package ui

import (
    tea "charm.land/bubbletea/v2"
    "charm.land/bubbles/v2/textinput"
    "charm.land/lipgloss/v2"
)
```

**Struct and constructor pattern** (search.go lines 23–42):
```go
type RecipientFormModel struct {
    input     textinput.Model
    errMsg    string // validation error to display below the input
    width     int
    height    int
    confirmed bool
    cancelled bool
}

func NewRecipientFormModel(width, height int) RecipientFormModel {
    ti := textinput.New()
    ti.Placeholder = "age1..."
    ti.CharLimit = 100  // T-02-07: mitigate DoS via long input
    ti.Prompt = ""
    cmd := ti.Focus()
    _ = cmd // Focus called eagerly at construction; caller must pass cmd via Init
    return RecipientFormModel{input: ti, width: width, height: height}
}
```

**SetActive / Focus pattern** (search.go lines 45–57):
```go
func (m *RecipientFormModel) Activate() tea.Cmd {
    m.confirmed = false
    m.cancelled = false
    m.errMsg = ""
    m.input.SetValue("")
    return m.input.Focus()
}
```

**Update with Enter validation** (search.go lines 83–87 + diff.go lines 96–117):
```go
func (m RecipientFormModel) Update(msg tea.Msg) (RecipientFormModel, tea.Cmd) {
    if kMsg, ok := msg.(tea.KeyPressMsg); ok {
        switch kMsg.String() {
        case "enter":
            if err := validateAgePublicKey(m.input.Value()); err != nil {
                m.errMsg = "Invalid age key: " + err.Error()
                return m, nil
            }
            m.confirmed = true
            return m, nil
        case "esc":
            m.cancelled = true
            return m, nil
        }
    }
    var cmd tea.Cmd
    m.input, cmd = m.input.Update(msg)
    return m, cmd
}
```

**View as full-screen bordered overlay** (diff.go lines 145–183):
```go
func (m RecipientFormModel) View() string {
    title := DiffKeyStyle.Render("Add Recipient")
    prompt := lipgloss.NewStyle().Foreground(ColorMuted).Render("Age public key:")
    inputArea := EditInputStyle.Width(m.width - 12).Render(m.input.View())
    errLine := ""
    if m.errMsg != "" {
        errLine = "\n" + lipgloss.NewStyle().Foreground(ColorError).Render(m.errMsg)
    }
    footer := ConfirmPromptStyle.Render("[enter]") + " confirm   " +
        ConfirmPromptStyle.Render("[esc]") + " cancel"
    inner := title + "\n\n" + prompt + " " + inputArea + errLine + "\n\n" + footer

    boxWidth := m.width - 2
    if boxWidth < 1 { boxWidth = 1 }
    boxHeight := m.height - 2
    if boxHeight < 1 { boxHeight = 1 }

    return lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(ColorMuted).
        Background(ColorSurface).
        Padding(1, SpaceMD).
        Width(boxWidth).
        Height(boxHeight).
        Render(inner)
}
```

---

### `internal/sops/executor.go` — AddRecipient / RemoveRecipient additions

**Analog:** `internal/sops/executor.go` (existing `SetKey` and `DecryptKey` functions)

**Function signature pattern** (executor.go lines 64–81):
```go
// DecryptKey decrypts and returns the plaintext value of a single key...
// ctx should have a timeout applied: context.WithTimeout(ctx, SopsTimeout).
func DecryptKey(ctx context.Context, filePath, keyPath string) (string, error) {
    if ctx == nil {
        ctx = context.Background()
    }
    cmd := exec.CommandContext(ctx, "sops", "decrypt", "--extract", index, filePath)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    if err := cmd.Run(); err != nil {
        return "", fmt.Errorf("sops decrypt --extract: %w: %s", err, strings.TrimSpace(stderr.String()))
    }
    return strings.TrimRight(stdout.String(), "\n"), nil
}
```

**New functions to add** (same package, same error wrapping pattern):
```go
// SopsRotateTimeout is the timeout for sops rotate operations.
// Longer than SopsTimeout because rotate decrypts all key-encryption records
// then re-encrypts all of them (Open Question 1 recommendation from RESEARCH.md).
const SopsRotateTimeout = 60 * time.Second

// AddRecipient adds an age public key recipient to a SOPS-encrypted file and
// re-encrypts the data key for the new recipient set.
//
// Runs: sops rotate -i --add-age <pubkey> <filePath>
//
// ctx should have a timeout applied: context.WithTimeout(ctx, SopsRotateTimeout).
func AddRecipient(ctx context.Context, filePath, pubkey string) error {
    if ctx == nil {
        ctx = context.Background()
    }
    cmd := exec.CommandContext(ctx, "sops", "rotate", "-i", "--add-age", pubkey, filePath)
    var stderr bytes.Buffer
    cmd.Stderr = &stderr
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("sops rotate --add-age: %w: %s", err, strings.TrimSpace(stderr.String()))
    }
    return nil
}

// RemoveRecipient removes an age public key recipient from a SOPS-encrypted file and
// re-encrypts the data key for the remaining recipient set.
//
// Runs: sops rotate -i --rm-age <pubkey> <filePath>
//
// ctx should have a timeout applied: context.WithTimeout(ctx, SopsRotateTimeout).
func RemoveRecipient(ctx context.Context, filePath, pubkey string) error {
    if ctx == nil {
        ctx = context.Background()
    }
    cmd := exec.CommandContext(ctx, "sops", "rotate", "-i", "--rm-age", pubkey, filePath)
    var stderr bytes.Buffer
    cmd.Stderr = &stderr
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("sops rotate --rm-age: %w: %s", err, strings.TrimSpace(stderr.String()))
    }
    return nil
}
```

---

### `internal/git/status.go` — GetLastCommitTime addition

**Analog:** `internal/git/status.go` (existing `GetFileHistory`)

**GetFileHistory pattern to copy** (status.go lines 126–163):
```go
func GetFileHistory(repoRoot, relPath string, limit int) ([]CommitEntry, error) {
    repo, err := gogit.PlainOpenWithOptions(repoRoot, &gogit.PlainOpenOptions{DetectDotGit: true})
    if err != nil {
        return nil, err
    }
    slashPath := filepath.ToSlash(relPath)
    iter, err := repo.Log(&gogit.LogOptions{FileName: &slashPath})
    if err != nil {
        return nil, err
    }
    var entries []CommitEntry
    err = iter.ForEach(func(c *object.Commit) error {
        if limit > 0 && len(entries) >= limit {
            return storer.ErrStop
        }
        // ...
        return nil
    })
    if err != nil {
        return nil, err
    }
    return entries, nil
}
```

**New function to add** (same file, same pattern — stop after first commit):
```go
// GetLastCommitTime returns the timestamp of the most recent commit for the
// file at relPath within the repository at repoRoot.
//
// relPath must be slash-separated and relative to the repository root (Pitfall 6).
// Returns time.Time{} (zero value) when the file has no commits.
// Returns error when repoRoot is not a git repository.
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
    // storer.ErrStop is expected — it is not a real error (per Assumption A3)
    if err != nil && err != storer.ErrStop {
        return time.Time{}, err
    }
    return commitTime, nil
}
```

---

### `internal/health/checker.go` (utility, transform)

**Analog:** `internal/ui/rotate.go` (pure computation, no Bubble Tea dependencies, stdlib only)

This is a new package. The pattern is a pure-computation file with no external dependencies beyond `math` and `strings` from stdlib.

**Package header pattern** (rotate.go lines 1–12):
```go
// Package health provides secret health check analysis for sops-tui.
//
// ShannonEntropy, IsWeakSecret, FindDuplicates, and IsStale are pure functions
// with no Bubble Tea or UI dependencies. They operate on decrypted string maps.
//
// Per D-08: weak = length < 16 OR entropy < 3.5 bits/char, with format-aware suffix checks.
// Per D-09: duplicate = exact decrypted value match across files (compared by SHA-256 hash).
// Per D-10: stale = days since last commit > threshold (default 90, env SOPS_TUI_STALE_DAYS).
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package health

import (
    "crypto/sha256"
    "fmt"
    "math"
    "strings"
    "time"
)
```

**Result types to define** (mirror RESEARCH.md Pattern 6):
```go
// WeakSecret identifies a key whose value is considered weak.
type WeakSecret struct {
    FilePath string
    KeyPath  string
    Reason   string // "too short", "low entropy", "format mismatch"
}

// Duplicate identifies two or more keys that share the same decrypted value.
type Duplicate struct {
    ValueHash string     // SHA-256 hex of the value (never the plaintext)
    Locations []Location // all file+keypath pairs sharing this value
}

// Location is a file+keypath pair.
type Location struct {
    FilePath string
    KeyPath  string
}

// StaleFile identifies a file whose last commit is older than the threshold.
type StaleFile struct {
    FilePath       string
    LastCommitTime time.Time
    DaysSince      int
}

// HealthCheckResult is the complete output of a health scan.
type HealthCheckResult struct {
    WeakSecrets []WeakSecret
    Duplicates  []Duplicate
    StaleFiles  []StaleFile
    Errors      []string // files that failed to decrypt
}

// IsEmpty returns true when no issues were found and no errors occurred.
func (r HealthCheckResult) IsEmpty() bool {
    return len(r.WeakSecrets) == 0 && len(r.Duplicates) == 0 &&
        len(r.StaleFiles) == 0 && len(r.Errors) == 0
}
```

**ShannonEntropy implementation** (RESEARCH.md Pattern 3, stdlib math only):
```go
// ShannonEntropy computes the Shannon entropy of s in bits per character.
// Returns 0.0 for empty strings.
func ShannonEntropy(s string) float64 {
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
```

**IsWeakSecret pattern** (RESEARCH.md Code Examples):
```go
var weakKeyNameSuffixes = []string{"_token", "_key", "_secret", "_password", "_pwd"}

// IsWeakSecret returns true and a reason string if the value at keyPath is considered weak.
// Checks: length < 16, entropy < 3.5, and format mismatch for known-sensitive key name suffixes.
func IsWeakSecret(keyPath, value string) (bool, string) {
    if len(value) < 16 {
        return true, "too short"
    }
    if ShannonEntropy(value) < 3.5 {
        return true, "low entropy"
    }
    lower := strings.ToLower(keyPath)
    for _, suffix := range weakKeyNameSuffixes {
        if strings.HasSuffix(lower, suffix) {
            // placeholder: caller can extend with format validation
            break
        }
    }
    return false, ""
}
```

**FindDuplicates using SHA-256** (RESEARCH.md Pitfall 2 — never store plaintext):
```go
// FindDuplicates scans a map of file→(keyPath→plaintextValue) for exact duplicates.
// Returns Duplicate entries identified by value SHA-256 hash, never the plaintext.
// The input map is consumed and the plaintext strings are not retained.
func FindDuplicates(fileValues map[string]map[string]string) []Duplicate {
    // map[valueHash][]Location
    seen := make(map[string][]Location)
    for filePath, kvs := range fileValues {
        for keyPath, value := range kvs {
            h := sha256.Sum256([]byte(value))
            hash := fmt.Sprintf("%x", h[:8]) // truncated to 8 bytes for display
            seen[hash] = append(seen[hash], Location{FilePath: filePath, KeyPath: keyPath})
        }
    }
    var dups []Duplicate
    for hash, locs := range seen {
        if len(locs) > 1 {
            dups = append(dups, Duplicate{ValueHash: hash, Locations: locs})
        }
    }
    return dups
}
```

---

### `internal/health/checker_test.go` (test)

**Analog:** `internal/sops/executor_test.go`

**Package + import pattern** (executor_test.go lines 1–12):
```go
// Package health tests.
package health_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/caesarakalaeii/sops-tui/internal/health"
)
```

**Table-driven test pattern** (executor_test.go lines 17–55):
```go
func TestShannonEntropy(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        wantMin  float64 // entropy must be >= wantMin
        wantMax  float64 // entropy must be <= wantMax
    }{
        { name: "empty string", input: "", wantMin: 0, wantMax: 0 },
        { name: "single character repeated", input: "aaaaaaaaaa", wantMin: 0, wantMax: 0.01 },
        { name: "random-looking 32-char string", input: "aBcDeFgH1234!@#$aBcDeFgH1234!@#$", wantMin: 3.5, wantMax: 5.0 },
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            got := health.ShannonEntropy(tc.input)
            assert.GreaterOrEqual(t, got, tc.wantMin)
            assert.LessOrEqual(t, got, tc.wantMax)
        })
    }
}
```

**Error path test pattern** (executor_test.go lines 173–178):
```go
func TestFindDuplicatesNoDuplicates(t *testing.T) {
    input := map[string]map[string]string{
        "a.yaml": {"key1": "value1"},
        "b.yaml": {"key1": "value2"},
    }
    dups := health.FindDuplicates(input)
    require.Empty(t, dups)
}
```

---

### `internal/app/model.go` — Phase 5 extensions

**Analog:** `internal/app/model.go` (existing — extend in place)

**New sessionState constants pattern** (model.go lines 44–62):
```go
// Add after stateHistory:
// stateHealth is the full-screen health check results overlay (HLT-03).
stateHealth
// stateRecipientForm is the modal overlay for entering a new age public key (RCP-02).
stateRecipientForm
// stateRecipientConfirm is the diff overlay showing recipient add/remove before re-encrypt (RCP-02).
stateRecipientConfirm
```

**New exported constants for tests** (model.go lines 64–70):
```go
const (
    StateHealth           = stateHealth
    StateRecipientForm    = stateRecipientForm
    StateRecipientConfirm = stateRecipientConfirm
)
```

**New message types** (model.go lines 72–170, follow existing pattern):
```go
// HealthCheckResultMsg carries completed health scan results (HLT-03).
type HealthCheckResultMsg struct {
    Result health.HealthCheckResult
    Err    error
}

// ReKeyDoneMsg carries the result of a sops rotate operation for one file (RCP-02, RCP-03).
type ReKeyDoneMsg struct {
    FilePath string
    Err      error
}
```

**Async cmd pattern** (model.go lines 340–362, RevealAllRequestMsg handler):
```go
// Health check async dispatch — follows the same pattern as RevealAllRequestMsg
return m, func() tea.Msg {
    result, err := runHealthCheck(ctx, m.files, m.gitRepoRoot, stalenessThreshold)
    return HealthCheckResultMsg{Result: result, Err: err}
}
```

**prevState pattern for new overlays** (model.go lines 428–436, stateDiff block):
```go
// stateHealth entry — mirrors stateHistory entry pattern (lines 817-840)
m.health = ui.NewHealthModel(m.width, mainH)
m.prevState = m.state
m.state = stateHealth
m.status.SetBreadcrumb("files", "health")
```

**Esc chain extension** (model.go lines 843–883 — add after stateHistory case):
```go
if m.state == stateHealth {
    m.state = m.prevState
    m.status.SetBreadcrumb("files")
    m.status.SetItemCount(m.fileList.ItemCount(), "items")
    return m, nil
}
if m.state == stateRecipientForm {
    m.state = m.prevState
    return m, nil
}
if m.state == stateRecipientConfirm {
    m.state = m.prevState
    return m, nil
}
```

**View switch extension** (model.go lines 960–976):
```go
case stateHealth:
    content = m.health.View()
case stateRecipientForm:
    content = m.recipientForm.View()
case stateRecipientConfirm:
    content = m.diff.View() // RecipientDiffOverlay reuses DiffModel
```

**New AppModel fields** (model.go lines 179–213):
```go
// Phase 5 fields
health         ui.HealthModel
recipientForm  ui.RecipientFormModel
bulkReKey      *bulkReKeyState  // nil when not in bulk re-key mode
```

**bulkReKeyState struct** (RESEARCH.md Pattern 5):
```go
// bulkReKeyState tracks progress of a sequential bulk re-key operation (RCP-03).
type bulkReKeyState struct {
    queue       []sops.DiscoveredFile // remaining files to process
    completed   int
    total       int
}
```

---

### `internal/keys/bindings.go` — Phase 5 additions

**Analog:** `internal/keys/bindings.go` (existing — extend in place)

**New FileListKeyMap fields** (bindings.go lines 43–64, after existing Info field):
```go
// ToggleSelect toggles the selection state of the highlighted file (D-05).
ToggleSelect key.Binding
// BulkReKey triggers bulk re-key on all selected files (D-05).
BulkReKey key.Binding
// HealthCheck triggers the on-demand secret health check (D-11).
HealthCheck key.Binding
```

**New DetailKeyMap fields** (bindings.go lines 127–166, after existing Blame field):
```go
// AddRecipient opens the add-recipient modal for the current file (RCP-02).
AddRecipient key.Binding
// RemoveRecipient opens the remove-recipient list for the current file (RCP-02).
RemoveRecipient key.Binding
```

**key.NewBinding pattern** (bindings.go lines 86–122):
```go
ToggleSelect: key.NewBinding(
    key.WithKeys("space"),
    key.WithHelp("space", "toggle select"),
),
BulkReKey: key.NewBinding(
    key.WithKeys("K"),
    key.WithHelp("K", "bulk re-key selected"),
),
HealthCheck: key.NewBinding(
    key.WithKeys("H"),
    key.WithHelp("H", "health check"),
),
// Detail bindings:
AddRecipient: key.NewBinding(
    key.WithKeys("a"),
    key.WithHelp("a", "add recipient"),
),
RemoveRecipient: key.NewBinding(
    key.WithKeys("d"),
    key.WithHelp("d", "remove recipient"),
),
```

**ShortHelp / FullHelp update** (bindings.go lines 68–80):
```go
// FileListKeyMap.ShortHelp — add ToggleSelect, BulkReKey, HealthCheck
func (k FileListKeyMap) ShortHelp() []key.Binding {
    return []key.Binding{k.Up, k.Down, k.Open, k.Search, k.Info, k.ToggleSelect, k.BulkReKey, k.HealthCheck, k.Help, k.Quit}
}
// DetailKeyMap.FullHelp — add AddRecipient, RemoveRecipient to actions group
```

---

### `internal/ui/filelist.go` — FileItem.Selected addition

**Analog:** `internal/ui/filelist.go` (existing — extend in place)

**FileItem struct extension** (filelist.go lines 26–38):
```go
// Add Selected bool field after existing GitStatus field:
// Selected indicates whether this file is selected for bulk operations (D-05).
Selected bool
```

**Title() method extension** (filelist.go lines 44–58):
```go
func (i FileItem) Title() string {
    base := i.Name
    if i.Selected {
        // Selection indicator rendered before badges (Claude's Discretion: checkbox style)
        base = lipgloss.NewStyle().Foreground(ColorAccent).Render("[+]") + " " + base
    }
    if !i.IsEncrypted {
        base += " " + BadgeUnencrypted.Render("[unencrypted]")
    }
    // ... existing git badge switch follows unchanged
}
```

**Update() intercept for ToggleSelect** (filelist.go lines 256–288 — intercept before m.list.Update()):
```go
// In the non-search KeyPressMsg block, before delegating to list:
case key.Matches(msg, m.keys.ToggleSelect):
    if item, ok := m.SelectedItem(); ok {
        // Toggle Selected on the item in allItems
        for i := range m.allItems {
            if m.allItems[i].Path == item.Path {
                m.allItems[i].Selected = !m.allItems[i].Selected
            }
        }
        // Rebuild list items to reflect new Selected state
        listItems := make([]list.Item, len(m.allItems))
        for i, it := range m.allItems {
            listItems[i] = it
        }
        m.list.SetItems(listItems)
    }
    return m, nil
```

**SelectedItems() helper** (new method, mirrors SelectedItem()):
```go
// SelectedItems returns all FileItems with Selected == true.
func (m FileListModel) SelectedItems() []FileItem {
    var selected []FileItem
    for _, item := range m.allItems {
        if item.Selected {
            selected = append(selected, item)
        }
    }
    return selected
}
```

---

## Shared Patterns

### Full-Screen Overlay Box
**Source:** `internal/ui/history.go` lines 133–141 (canonical — also metadata.go 172–180, diff.go 175–183)
**Apply to:** `HealthModel.View()`, `RecipientFormModel.View()`
```go
return lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(ColorMuted).
    Background(ColorSurface).
    Padding(1, SpaceMD).
    Width(m.width - 2).
    Height(m.height - 2).
    Render(inner)
```

### prevState Overlay Pattern
**Source:** `internal/app/model.go` lines 428–436 (stateDiff entry) and lines 862–883 (Esc chain)
**Apply to:** stateHealth, stateRecipientForm, stateRecipientConfirm in AppModel.Update()
```go
// Entering overlay:
m.prevState = m.state
m.state = stateNewOverlay
m.status.SetBreadcrumb(...)

// Esc chain (model.go ~line 860):
if m.state == stateNewOverlay {
    m.state = m.prevState
    // restore status bar to prevState values
    return m, nil
}
```

### Async Cmd Pattern
**Source:** `internal/app/model.go` lines 340–362 (RevealAllRequestMsg), lines 497–505 (git status refresh)
**Apply to:** HealthCheckRequestMsg handler, AddRecipient/RemoveRecipient handlers
```go
return m, func() tea.Msg {
    ctx, cancel := context.WithTimeout(context.Background(), sops.SopsRotateTimeout)
    defer cancel()
    err := sops.AddRecipient(ctx, filePath, pubkey)
    return ReKeyDoneMsg{FilePath: filePath, Err: err}
}
```

### Flash Progress Pattern
**Source:** `internal/app/model.go` line 369 (`m.status, _ = m.status.Flash(...)`)
**Apply to:** Bulk re-key progress updates (D-07)
```go
// In bulk re-key progression:
m.status, _ = m.status.Flash(fmt.Sprintf("Re-keying %d/%d: %s",
    m.bulkReKey.completed+1, m.bulkReKey.total,
    filepath.Base(m.bulkReKey.queue[0].Name)))
```

### DiffModel Reuse for Recipient Confirmation
**Source:** `internal/ui/diff.go` + `internal/app/model.go` lines 431–436 (NewDiffModel usage)
**Apply to:** stateRecipientConfirm (D-04 — recipient add/remove confirmation)
```go
// Build DiffEntries showing recipient changes (old = current recipients, new = updated recipients)
entries := []ui.DiffEntry{
    {KeyPath: "recipients", OldValue: oldRecipientList, NewValue: newRecipientList},
}
m.diff = ui.NewDiffModel("Recipient Change: "+filePath, entries, m.width, mainH)
m.prevState = m.state
m.state = stateRecipientConfirm  // same stateDiff block handles y/n confirmation
```

### Error Handling — sops Subprocess
**Source:** `internal/sops/executor.go` lines 75–77
**Apply to:** AddRecipient, RemoveRecipient
```go
if err := cmd.Run(); err != nil {
    return fmt.Errorf("sops rotate --add-age: %w: %s", err, strings.TrimSpace(stderr.String()))
}
```

### Test Helpers — initRepo / commitFile
**Source:** `internal/git/status_test.go` lines 23–57
**Apply to:** `internal/git/status_test.go` additions (TestGetLastCommitTime)
```go
func TestGetLastCommitTime(t *testing.T) {
    t.Run("returns commit timestamp for committed file", func(t *testing.T) {
        dir := t.TempDir()
        repo := initRepo(t, dir)
        commitFile(t, repo, dir, "secrets.yaml", "v1", "initial commit")
        ts, err := git.GetLastCommitTime(dir, "secrets.yaml")
        require.NoError(t, err)
        assert.False(t, ts.IsZero(), "commit time must not be zero")
    })
    t.Run("non-git directory returns error", func(t *testing.T) {
        dir := t.TempDir()
        _, err := git.GetLastCommitTime(dir, "secrets.yaml")
        assert.Error(t, err)
    })
}
```

### styles.go — New Styles (add at bottom of Phase 4 section)
**Source:** `internal/ui/styles.go` lines 180–196 (Phase 4 styles block pattern)
**Apply to:** `internal/ui/styles.go`
```go
// Phase 5: Health check severity styles (HLT-03)
HealthWeakStyle    = lipgloss.NewStyle().Foreground(ColorWarning)
HealthDupStyle     = lipgloss.NewStyle().Foreground(ColorError)
HealthStaleStyle   = lipgloss.NewStyle().Foreground(ColorMuted)
HealthOkStyle      = lipgloss.NewStyle().Foreground(ColorSuccess)
// Phase 5: Selection indicator for bulk re-key (D-05)
SelectionIndicator = lipgloss.NewStyle().Foreground(ColorAccent)
```

---

## No Analog Found

All Phase 5 files have close analogs in the codebase. No files lack a pattern match.

---

## Key Pitfalls to Carry Forward

1. **`space` and `K` and `H` must be intercepted before `m.list.Update(msg)` in `FileListModel.Update()`** — see RESEARCH.md Pitfall 3 and 4. The intercept block at filelist.go lines 256–288 is the exact insertion point.

2. **`stateRecipientConfirm` can reuse the existing `stateDiff` block in `AppModel.Update()`** — `DiffModel.Confirmed()` / `DiffModel.Cancelled()` drives the confirmation; the only difference is the success handler calls `sops.AddRecipient` / `sops.RemoveRecipient` instead of `sops.SetKey`.

3. **`storer.ErrStop` in `GetLastCommitTime` is expected** — do not treat it as a real error (see RESEARCH.md Assumption A3). The pattern from `GetFileHistory` (status.go line 144) already demonstrates this: `return storer.ErrStop` terminates iteration cleanly.

4. **Duplicate detection must clear the plaintext map after `HealthCheckResultMsg` is built** — `FindDuplicates` in `health/checker.go` takes the map and returns only hashes; the caller in `AppModel` must nil-out the map reference after calling it (RESEARCH.md Pitfall 2).

5. **`filippo.io/age` needs `go get`** — it is not in go.mod. The Wave 0 plan task should include `go get filippo.io/age@v1.3.1`.

---

## Metadata

**Analog search scope:** `internal/ui/`, `internal/app/`, `internal/sops/`, `internal/git/`, `internal/keys/`
**Files scanned:** 28 source files + 14 test files
**Pattern extraction date:** 2026-04-16
