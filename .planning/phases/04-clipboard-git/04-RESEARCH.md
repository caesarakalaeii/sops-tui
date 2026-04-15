# Phase 4: Clipboard & Git - Research

**Researched:** 2026-04-15
**Domain:** Clipboard management, OS signal handling, go-git integration, cross-file fuzzy search
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Clipboard Copy Flow**
- D-01: Copy-to-clipboard is triggered by `ctrl+y`. Chord avoids accidental copy of sensitive data.
- D-02: Copy only works on revealed leaf values — user must reveal with `r` first, then `ctrl+y`. `ctrl+y` on a masked value is a no-op (flash: "Reveal first with r").
- D-03: Copy is available in the detail view only — not from the file list.
- D-04: Copy feedback uses flash message: "Copied (clears in 30s)" displayed in status bar for 2-3 seconds. Follows existing flash pattern.

**Auto-Clear & Signal Safety**
- D-05: Auto-clear timeout defaults to 30 seconds. Configurable via `SOPS_TUI_CLIPBOARD_TIMEOUT` env var (integer seconds). Falls back to 30s if unset or invalid.
- D-06: If user copies again before timeout, timer resets — new content replaces old, countdown starts fresh.
- D-07: Clipboard cleanup on exit uses `os/signal` goroutine — `signal.NotifyContext` for SIGINT/SIGTERM. Also `defer clipboard.WriteAll("")` in main as backup.
- D-08: Subtle indicator appears in status bar while clipboard holds a secret. Disappears after auto-clear.

**Git Change Badges**
- D-09: Badges: `[M]` (modified/yellow), `[A]` (added/green), `[?]` (untracked/muted). Text format.
- D-10: Badges appear in both file list (next to filename) and detail view header/breadcrumb.
- D-11: Git status fetched on startup (during file discovery) and refreshed after any write operation. No polling.
- D-12: Non-git repos show a subtle "no git" indicator in status bar. Git badges simply don't appear.

**Git History View**
- D-13: Git history is a full-screen overlay triggered by `b` key. Same `sessionState`/`prevState` pattern as help/metadata/diff. Press `b` or Esc to close.
- D-14: History accessible from detail view only.
- D-15: Each entry shows: short hash, relative date, author name, commit subject. Scrollable with j/k.
- D-16: `go-git/go-git` v5 provides git backend — pure Go, no git binary dependency.

### Claude's Discretion
- Cross-file search implementation (GIT-03) — how to aggregate results across files, result presentation format
- Clipboard indicator icon/symbol choice for status bar
- Git history overlay layout and formatting details
- `go-git` repository caching strategy (open once, reuse handle)
- Auto-clear timer implementation (time.AfterFunc vs goroutine + ticker)
- Badge positioning relative to filename in file list delegate

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CLB-01 | User can copy decrypted value to clipboard | atotto/clipboard v0.1.4 `WriteAll()` API verified; already indirect dep, needs promotion to direct |
| CLB-02 | Clipboard auto-clears after configurable timeout (default 30s) | `tea.Tick` pattern from existing `statusbar.go` Flash() — same generation-counter approach applies |
| CLB-03 | Clipboard clears on process exit (including SIGINT/SIGTERM) | `os/signal.NotifyContext` + `defer clipboard.WriteAll("")` in main.go; no existing signal handling to conflict with |
| GIT-01 | User sees uncommitted change badges on files ([M], [A], [?]) | `go-git` v5.17.0 `Worktree.Status()` returns `Status` map keyed by relative path; `FileStatus.Worktree` is `StatusCode` byte |
| GIT-02 | User can view git blame/history per secret file | `repo.Log(&LogOptions{FileName: &relPath})` returns `CommitIter`; each `Commit` has Hash, Author.Name, Author.When, Message |
| GIT-03 | User can fuzzy search across all files and key names simultaneously | Extend existing `sahilm/fuzzy` + `SearchModel` pattern; requires parsing all files on-demand then caching key lists |
</phase_requirements>

---

## Summary

Phase 4 adds three independent feature areas that integrate cleanly with the existing codebase patterns. The clipboard work extends `statusbar.go` with a new persistent indicator and reuses the `tea.Tick` + generation-counter pattern already established for flash messages. The git badge work extends `DiscoveredFile` with a `GitStatus` field populated during discovery via `go-git` v5 `Worktree.Status()`. The git history overlay follows the exact same `sessionState`/`prevState` pattern as the existing metadata and diff overlays. Cross-file search (GIT-03) is the most novel feature — it requires aggregating key lists from all files, which can be done lazily by parsing on first `/` activation in the file list.

The key integration risk is that `go-git/go-git/v5` was added to go.mod as an indirect dependency during research (it was `go get`-installed to verify the API). The plan must promote it to a direct dependency. The clipboard tool (atotto/clipboard) is already in go.mod as indirect — it also needs promotion. Both promotions are a one-line `go.mod` edit.

**Primary recommendation:** Implement in four distinct plans: (1) clipboard copy + auto-clear, (2) signal cleanup + status-bar indicator, (3) git badges + history overlay, (4) cross-file search. This mirrors the existing phase decomposition pattern and minimizes merge conflicts.

---

## Project Constraints (from CLAUDE.md)

| Constraint | Directive |
|-----------|-----------|
| Never use `type any` | Use proper typing throughout |
| Never use `lipgloss.AdaptiveColor` | Use explicit hex colors (issue #1036 confirmed hang) |
| Stack locked | Go 1.26.2, bubbletea v2.0.4, lipgloss v2, bubbles v2 |
| bubbletea v2 API | `View()` returns `tea.View`, `tea.KeyPressMsg`, `msg.Code`/`msg.Text`/`msg.Mod` |
| Key space: `" "` → `"space"` in `msg.String()` | Verify chord encoding for `ctrl+y` |
| atotto/clipboard v0.1.4 | Already in go.mod as indirect — promote to direct |
| go-git/go-git v5.17.2 (CLAUDE.md) | v5.17.0 was available on registry; v5.17.2 not yet present — use v5.17.0 |
| Single-binary distribution | CGO_ENABLED=0 — atotto/clipboard uses subprocess (xclip/xsel/wl-copy), no CGO needed |

---

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| atotto/clipboard | v0.1.4 | Clipboard read/write | Already in go.mod (indirect); single API: `WriteAll(string)`, `ReadAll() string`; no CGo; uses wl-copy on Wayland, xclip/xsel on X11 |
| go-git/go-git/v5 | v5.17.0 | Git status and history | Pure Go, no `git` binary dep; `PlainOpen`, `Worktree.Status()`, `repo.Log()` verified in module cache |
| os/signal | stdlib | SIGINT/SIGTERM handling | `signal.NotifyContext` wraps context cancellation; standard Go pattern |
| sahilm/fuzzy | v0.1.1 | Cross-file fuzzy search | Already direct dep; `fuzzy.Find(pattern, sources)` returns ranked matches with `MatchedIndexes` |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| tea.Tick | bubbletea v2.0.4 | Clipboard auto-clear timer | Send `ClipboardClearMsg{Gen: n}` after timeout — same generation-counter pattern as `FlashClearMsg` |
| time | stdlib | Duration parsing, relative dates | `time.Since(commit.Author.When)` for relative history dates; `strconv.Atoi` for env var timeout |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| tea.Tick for clipboard timer | time.AfterFunc + goroutine | `tea.Tick` stays inside the bubbletea event loop — no synchronization needed. `time.AfterFunc` fires from a separate goroutine and requires `p.Send()` to inject the message. `tea.Tick` is simpler and matches the existing `Flash()` pattern exactly. |
| go-git Worktree.Status() | os/exec git subprocess | Subprocess returns unstructured text requiring parsing. go-git returns typed `Status` map. Also removes external `git` binary dependency. |
| Cached go-git repo handle | PlainOpen on every operation | PlainOpen is cheap (just opens .git/config) but holding a `*git.Repository` handle as an AppModel field is cleaner and avoids repeated stat calls |

**Installation (go.mod promotion only — packages already downloaded):**
```bash
go get github.com/go-git/go-git/v5@v5.17.0
go get github.com/atotto/clipboard@v0.1.4
```

**Version verification:** [VERIFIED: go module cache] Both packages confirmed in `/home/moersener/go/pkg/mod/`:
- `github.com/atotto/clipboard@v0.1.4` — already indirect in go.mod
- `github.com/go-git/go-git/v5@v5.17.0` — added during research session (go get ran successfully)

Note: CLAUDE.md lists v5.17.2 for go-git but v5.17.0 is the version available in the module cache after `go get`. Use v5.17.0 — the API is identical for the operations needed.

---

## Architecture Patterns

### Recommended Project Structure

New files to create:
```
internal/
├── git/
│   └── status.go            # GitRepo: PlainOpen, Status, Log — wraps go-git, returns typed results
internal/ui/
├── history.go               # HistoryModel: full-screen overlay (mirrors metadata.go)
├── history_test.go
internal/app/
└── model.go                 # Extended: stateHistory, GitStatusMsg, ClipboardClearedMsg, GitRepoMsg
```

Modified files:
```
internal/ui/
├── statusbar.go             # Add clipboardHot bool + "no git" indicator to EnvStatus/View
├── filelist.go              # FileItem: add GitStatus string field; Title() append badge
├── detail.go                # Header rendering: add GitStatus badge to breadcrumb
├── styles.go                # Add BadgeModified, BadgeAdded, BadgeUntracked style vars
internal/keys/
└── bindings.go              # DetailKeyMap: add Copy (ctrl+y) and Blame (b) bindings
internal/sops/
└── discoverer.go            # DiscoveredFile: add GitStatus string field
cmd/sops-tui/
└── main.go                  # Signal handler: os/signal.NotifyContext + defer clipboard.WriteAll("")
```

### Pattern 1: Clipboard Timer with Generation Counter

**What:** Use `tea.Tick` to schedule clipboard auto-clear, with a generation counter to prevent stale clears. Mirrors the existing `Flash()` / `FlashClearMsg` pattern in `statusbar.go`.

**When to use:** Any time-bounded state that can be reset (copy triggers new timer, old timer is orphaned but harmless due to generation check).

```go
// Source: internal/ui/statusbar.go (existing pattern, adapted)
type ClipboardClearMsg struct {
    Gen int
}

// In AppModel:
clipboardGen int
clipboardHot bool

func (m AppModel) copyToClipboard(value string) (AppModel, tea.Cmd) {
    if err := clipboard.WriteAll(value); err != nil {
        var statusCmd tea.Cmd
        m.status, statusCmd = m.status.Flash("Clipboard error: " + err.Error())
        return m, statusCmd
    }
    m.clipboardGen++
    gen := m.clipboardGen
    m.clipboardHot = true
    m.status.SetClipboardHot(true)
    var statusCmd tea.Cmd
    m.status, statusCmd = m.status.Flash("Copied (clears in 30s)")
    timeout := clipboardTimeout()
    clearCmd := tea.Tick(timeout, func(_ time.Time) tea.Msg {
        return ClipboardClearMsg{Gen: gen}
    })
    return m, tea.Batch(statusCmd, clearCmd)
}

func clipboardTimeout() time.Duration {
    if s := os.Getenv("SOPS_TUI_CLIPBOARD_TIMEOUT"); s != "" {
        if n, err := strconv.Atoi(s); err == nil && n > 0 {
            return time.Duration(n) * time.Second
        }
    }
    return 30 * time.Second
}
```

### Pattern 2: Signal-Safe Clipboard Cleanup in main.go

**What:** Register a SIGINT/SIGTERM handler that clears the clipboard before exit. Belt-and-suspenders with `defer clipboard.WriteAll("")`.

**When to use:** Any sensitive data held in process memory that must be cleaned up on exit.

```go
// Source: os/signal stdlib docs + D-07 decision
// In cmd/sops-tui/main.go:
import (
    "os/signal"
    "syscall"
    "github.com/atotto/clipboard"
)

func main() {
    // ... existing startup validation ...

    // Belt: defer clear on normal exit
    defer clipboard.WriteAll("")

    // Suspenders: signal handler for SIGINT/SIGTERM
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    go func() {
        <-ctx.Done()
        clipboard.WriteAll("")
        os.Exit(0)
    }()

    // ... p := tea.NewProgram(model); p.Run() ...
}
```

**Critical:** `signal.NotifyContext` requires `context` and `syscall` imports. The goroutine that waits on `ctx.Done()` must call `os.Exit(0)` — `p.Quit()` is not available outside the bubbletea run loop.

### Pattern 3: go-git Status and Badge Mapping

**What:** Open repo once with `PlainOpen`, call `Worktree.Status()` to get all file statuses, look up each discovered file by relative path.

**When to use:** Startup (file discovery) and after each write operation (re-encryption completes).

```go
// Source: [VERIFIED: go module cache] go-git v5.17.0 status.go + worktree_status.go
import (
    git "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// In internal/git/status.go:
type GitStatus string

const (
    GitStatusClean    GitStatus = ""
    GitStatusModified GitStatus = "M"
    GitStatusAdded    GitStatus = "A"
    GitStatusUntracked GitStatus = "?"
)

func GetFileStatuses(repoRoot string, relPaths []string) (map[string]GitStatus, error) {
    repo, err := git.PlainOpen(repoRoot)
    if err != nil {
        // Not a git repo — return empty map, not an error
        if err == git.ErrRepositoryNotExists {
            return map[string]GitStatus{}, nil
        }
        return nil, err
    }
    wt, err := repo.Worktree()
    if err != nil {
        return nil, err
    }
    status, err := wt.Status()
    if err != nil {
        return nil, err
    }
    result := make(map[string]GitStatus, len(relPaths))
    for _, relPath := range relPaths {
        fs := status.File(relPath)
        switch {
        case fs.Worktree == git.Modified || fs.Staging == git.Modified:
            result[relPath] = GitStatusModified
        case fs.Worktree == git.Added || fs.Staging == git.Added:
            result[relPath] = GitStatusAdded
        case fs.Worktree == git.Untracked:
            result[relPath] = GitStatusUntracked
        default:
            result[relPath] = GitStatusClean
        }
    }
    return result, nil
}
```

**Key API facts** [VERIFIED: go module cache]:
- `git.PlainOpen(path string) (*Repository, error)` — opens existing repo at path or parent dirs
- `git.ErrRepositoryNotExists` — sentinel error for non-git directories
- `Worktree.Status() (Status, error)` — returns `map[string]*FileStatus` keyed by slash-separated relative path
- `FileStatus.Staging StatusCode` and `FileStatus.Worktree StatusCode` — `byte` constants: `' '`=Unmodified, `'M'`=Modified, `'A'`=Added, `'?'`=Untracked, `'D'`=Deleted
- `Status.File(path string) *FileStatus` — safe getter, creates default entry if absent

### Pattern 4: Git History Log

**What:** Use `repo.Log(&LogOptions{FileName: &relPath})` to get per-file commit history. Iterate with `CommitIter.ForEach`.

**When to use:** When user presses `b` in detail view — fetched async via `tea.Cmd`, displayed in `HistoryModel`.

```go
// Source: [VERIFIED: go module cache] go-git v5.17.0 options.go + plumbing/object/commit.go
import (
    git "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing/object"
)

type CommitEntry struct {
    ShortHash string    // first 7 chars of hash hex
    RelDate   string    // e.g. "3 days ago"
    Author    string    // commit.Author.Name
    Subject   string    // first line of commit.Message
}

func GetFileHistory(repoRoot, relPath string, limit int) ([]CommitEntry, error) {
    repo, err := git.PlainOpen(repoRoot)
    if err != nil {
        return nil, err
    }
    iter, err := repo.Log(&git.LogOptions{FileName: &relPath})
    if err != nil {
        return nil, err
    }
    var entries []CommitEntry
    err = iter.ForEach(func(c *object.Commit) error {
        if limit > 0 && len(entries) >= limit {
            return storer.ErrStop
        }
        subject := c.Message
        if i := strings.IndexByte(subject, '\n'); i >= 0 {
            subject = subject[:i]
        }
        entries = append(entries, CommitEntry{
            ShortHash: c.Hash.String()[:7],
            RelDate:   relativeTime(c.Author.When),
            Author:    c.Author.Name,
            Subject:   subject,
        })
        return nil
    })
    return entries, err
}
```

**Commit fields** [VERIFIED: go module cache] `plumbing/object/commit.go` + `plumbing/object/object.go`:
- `c.Hash` → `plumbing.Hash` — call `.String()` then take first 7 chars for short hash
- `c.Author` → `object.Signature{Name string, Email string, When time.Time}`
- `c.Message` → `string` — first line is subject; full message may include body

### Pattern 5: HistoryModel Overlay

**What:** A new `HistoryModel` struct in `internal/ui/history.go` mirroring `MetadataModel` — bordered box, j/k scrolling, Esc closes.

**When to use:** `stateHistory` in `AppModel.sessionState` enum.

```go
// Source: Mirrors internal/ui/metadata.go (existing pattern)
type HistoryModel struct {
    entries []CommitEntry  // from internal/git package
    width   int
    height  int
    scroll  int
}

// HistoryRequestMsg sent by detail model when user presses b
type HistoryRequestMsg struct {
    FilePath string
    RelPath  string
}

// GitHistoryMsg carries async result
type GitHistoryMsg struct {
    Entries []git.CommitEntry
    Err     error
}
```

**sessionState enum addition:** Add `stateHistory` after `stateFormatMenu` in `model.go`.

### Pattern 6: Cross-File Search (GIT-03)

**What:** When `/` is pressed from the file list, activate a global search mode that matches across all file names AND all key paths from all files. Results show `filename > key.path` format.

**When to use:** Only triggered from `stateFileList` when the user presses `/` — this is a new mode distinct from the existing per-view search.

**Recommended approach:** Lazy key aggregation. On first global search activation, parse all discovered files to collect their key paths. Cache results in `AppModel`. Results presented as a flat list of `SearchResult{FileName, RelPath, KeyPath}` items in `FileListModel` or a new overlay list.

```go
// SearchResult aggregates file + key path for display
type SearchResult struct {
    FileName string  // "secrets/prod.yaml"
    KeyPath  string  // "database.password" (empty for file-level matches)
    RelPath  string  // absolute path for navigation
}

// SearchResultItem implements list.DefaultItem
func (r SearchResult) Title() string {
    if r.KeyPath == "" {
        return r.FileName
    }
    return r.FileName + " > " + r.KeyPath
}
func (r SearchResult) Description() string { return r.RelPath }
func (r SearchResult) FilterValue() string { return r.Title() }
```

**Implementation choices** (Claude's discretion):
- Parse key paths from `DiscoveredFile` YAML without decryption — only need key names, not values. Use existing `parser.ParseFile` which already reads unencrypted key structure.
- Cache in `AppModel.allKeyPaths []SearchResult` populated on first global search.
- Invalidate cache on `FilesDiscoveredMsg` (re-discovery).
- Use `sahilm/fuzzy.Find(pattern, titles)` over the `Title()` values — same pattern as existing file search.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Clipboard access | Custom xclip/wl-copy subprocess code | `atotto/clipboard.WriteAll()` | Already handles Wayland vs X11 detection at init() time; `Unsupported` flag available for graceful degradation |
| Timer with stale-clear prevention | goroutine + channel + mutex | `tea.Tick` + generation counter | Pattern already proven in Flash() — zero synchronization complexity inside the bubbletea event loop |
| Git status parsing | Parsing `git status --porcelain` output | `go-git Worktree.Status()` | Returns typed `Status` map; no subprocess, no text parsing, no PATH dependency |
| Git log parsing | Parsing `git log --format=...` output | `go-git repo.Log()` + `CommitIter` | Returns typed `*object.Commit` with `Hash`, `Author.Name`, `Author.When`, `Message` |
| Relative time formatting | String builder for "3 days ago" | Small `relativeTime(t time.Time) string` helper | No library needed — a 15-line switch on `time.Since(t)` buckets is sufficient and has no deps |

**Key insight:** The entire feature set uses exactly two external packages (atotto/clipboard + go-git) both of which are already in go.mod. The clipboard timer, history overlay, badge rendering, and search aggregation all use existing project patterns verbatim.

---

## Common Pitfalls

### Pitfall 1: Clipboard Unsupported Flag
**What goes wrong:** On machines without xclip, xsel, or wl-copy/wl-paste installed, `atotto/clipboard` sets `clipboard.Unsupported = true` at `init()` time. Calling `WriteAll()` returns an error.
**Why it happens:** The library detects clipboard tools in `init()` — if none found, `Unsupported = true`.
**How to avoid:** Check `clipboard.Unsupported` before offering the copy keybinding. If true, show flash "Clipboard not available (install xclip or wl-clipboard)" instead of attempting the write.
**Warning signs:** `WriteAll()` returns `errors.New("No clipboard utilities available...")` on Linux.

**Environment note:** The development machine has `wl-copy` available (`WAYLAND_DISPLAY=wayland-0`), so clipboard will work in development. The README must document the `xclip` or `wl-clipboard` requirement for X11/Wayland respectively.

### Pitfall 2: Stale ClipboardClearMsg After Re-copy
**What goes wrong:** User copies value A (timer fires after 30s). Before 30s, user copies value B. Timer for value A fires and clears clipboard — deleting value B.
**Why it happens:** Both timers are in flight.
**How to avoid:** Generation counter — `D-06` decision. Increment `clipboardGen` on each copy; only process `ClipboardClearMsg` if `msg.Gen == m.clipboardGen`. Identical to Flash generation pattern.

### Pitfall 3: go-git PlainOpen on Non-Git Directory
**What goes wrong:** `git.PlainOpen(path)` returns `git.ErrRepositoryNotExists` if the directory is not a git repo (or has no `.git` parent). Unhandled, this shows an error to the user.
**Why it happens:** `PlainOpen` searches upward for `.git` — returns sentinel error if not found.
**How to avoid:** Check `err == git.ErrRepositoryNotExists` and treat as "no git" state (D-12). The `EnvStatus` struct needs a `GitAvailable bool` field. Set it to `false` when `ErrRepositoryNotExists` is returned.

### Pitfall 4: go-git Worktree.Status() Performance on Large Repos
**What goes wrong:** `Worktree.Status()` on a large repo (thousands of files) is slower than `git status` because go-git is pure Go. In practice, for sops-tui's use case (a small set of secret files), this is unlikely to matter.
**Why it happens:** go-git reads the index and hashes all tracked files in Go vs native C in git.
**How to avoid:** Run git status asynchronously as a `tea.Cmd` (same pattern as file discovery). The `FilesDiscoveredMsg` handler already has a natural async hook. Add a `GitStatusMsg` that arrives after discovery with the badge data. Never block the UI thread.

### Pitfall 5: Signal Handler Race with bubbletea's Internal Signal Handling
**What goes wrong:** bubbletea v2 uses `signal.Notify` internally for `SIGWINCH` (terminal resize). Adding a second `signal.Notify` for SIGINT/SIGTERM from a goroutine can interfere.
**Why it happens:** Multiple callers of `signal.Notify` for the same signal cause the signal to be delivered to all registered channels.
**How to avoid:** Use `signal.NotifyContext` (wraps the context pattern cleanly) rather than raw `signal.Notify`. The goroutine waiting on `ctx.Done()` calls `clipboard.WriteAll("")` then `os.Exit(0)` — it does not try to interact with the bubbletea program. The `defer clipboard.WriteAll("")` in main covers clean shutdown paths.

### Pitfall 6: go-git Log FileName Filter Requires Slash Separators
**What goes wrong:** `LogOptions.FileName` uses forward-slash paths (git convention). On Windows, `filepath.Rel` returns backslash paths. The filter silently matches nothing.
**Why it happens:** go-git uses Unix path conventions internally.
**How to avoid:** Always convert relative paths with `filepath.ToSlash(relPath)` before passing to `LogOptions.FileName`. Already done in `discoverer.go` — use the same pattern.

### Pitfall 7: ctrl+y Key Code in bubbletea v2
**What goes wrong:** `ctrl+y` produces a specific `tea.KeyPressMsg` — verify the `msg.String()` value matches what `key.NewBinding(key.WithKeys("ctrl+y"))` expects.
**Why it happens:** bubbletea v2 changed key representation (`msg.Code` is a rune, `msg.Mod` replaces `msg.Alt`). Chord matching via `key.Matches()` should still work but must be verified.
**How to avoid:** In tests, send `tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl}` and verify the binding fires. The existing `ctrl+u`/`ctrl+d` bindings in `bindings.go` confirm this pattern already works.

---

## Code Examples

### Clipboard Write with Error Handling
```go
// Source: [VERIFIED: go module cache] github.com/atotto/clipboard@v0.1.4/clipboard.go
import "github.com/atotto/clipboard"

// Check availability before attempting write
if clipboard.Unsupported {
    // flash "Clipboard not available (install xclip or wl-clipboard)"
    return
}
if err := clipboard.WriteAll(value); err != nil {
    // flash error
}
// Clear:
clipboard.WriteAll("") // empty string clears clipboard
```

### go-git Status Lookup Pattern
```go
// Source: [VERIFIED: go module cache] go-git v5.17.0 status.go
import git "github.com/go-git/go-git/v5"

repo, err := git.PlainOpen(dir)
if err == git.ErrRepositoryNotExists {
    // Not a git repo — normal, not an error
}
wt, _ := repo.Worktree()
status, _ := wt.Status()
fs := status.File("secrets/prod.yaml")  // slash-separated relative path
// fs.Worktree: ' '=clean, 'M'=modified, 'A'=added, '?'=untracked, 'D'=deleted
// fs.Staging:  same codes for staged changes
```

### Relative Time Helper
```go
// Source: [ASSUMED] — standard pattern, no library needed
func relativeTime(t time.Time) string {
    d := time.Since(t)
    switch {
    case d < time.Minute:
        return "just now"
    case d < time.Hour:
        return fmt.Sprintf("%d minutes ago", int(d.Minutes()))
    case d < 24*time.Hour:
        return fmt.Sprintf("%d hours ago", int(d.Hours()))
    case d < 30*24*time.Hour:
        return fmt.Sprintf("%d days ago", int(d.Hours()/24))
    case d < 365*24*time.Hour:
        return fmt.Sprintf("%d months ago", int(d.Hours()/(24*30)))
    default:
        return fmt.Sprintf("%d years ago", int(d.Hours()/(24*365)))
    }
}
```

### HistoryModel View (mirrors MetadataModel)
```go
// Source: mirrors internal/ui/metadata.go (existing pattern)
func (m HistoryModel) View() string {
    title := HelpSectionHeader.Render("Git History")
    var lines []string
    hashStyle := lipgloss.NewStyle().Foreground(ColorAccent).Width(8)
    dateStyle := lipgloss.NewStyle().Foreground(ColorMuted).Width(14)
    authorStyle := lipgloss.NewStyle().Foreground(ColorFg).Width(16)
    for _, e := range m.entries {
        line := hashStyle.Render(e.ShortHash) +
                dateStyle.Render(e.RelDate) +
                authorStyle.Render(e.Author) +
                e.Subject
        lines = append(lines, line)
    }
    // apply scroll, render bordered box — same as MetadataModel.View()
}
```

### EnvStatus Extension for Git
```go
// Source: internal/ui/statusbar.go (existing struct, to be extended)
type EnvStatus struct {
    SopsAvailable     bool
    AgeAvailable      bool
    SopsYamlAvailable bool
    GitAvailable      bool  // NEW: false when not in a git repo
}
```

### FileItem Extension for Git Badge
```go
// Source: internal/ui/filelist.go (existing struct, to be extended)
type FileItem struct {
    Name        string
    Path        string
    IsEncrypted bool
    Rule        sops.CreationRule
    GitStatus   string  // NEW: "M", "A", "?", or "" for clean
}

func (i FileItem) Title() string {
    base := i.Name
    if !i.IsEncrypted {
        base += " " + BadgeUnencrypted.Render("[unencrypted]")
    }
    switch i.GitStatus {
    case "M":
        base += " " + BadgeModified.Render("[M]")
    case "A":
        base += " " + BadgeAdded.Render("[A]")
    case "?":
        base += " " + BadgeUntracked.Render("[?]")
    }
    return base
}
```

### DiscoveredFile Extension
```go
// Source: internal/sops/discoverer.go (existing struct, to be extended)
type DiscoveredFile struct {
    Name        string
    AbsPath     string
    IsEncrypted bool
    Rule        CreationRule
    GitStatus   string  // NEW: "M", "A", "?", or "" for clean/no-git
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `signal.Notify(ch, os.Interrupt)` | `signal.NotifyContext(ctx, syscall.SIGINT, ...)` | Go 1.16 | Context-based cancellation is idiomatic; goroutine waits on `<-ctx.Done()` |
| go-git v4 (`gopkg.in/src-d/go-git.v4`) | go-git v5 (`github.com/go-git/go-git/v5`) | 2019 | v4 is abandoned; v5 is the maintained fork by go-git org |

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `storer.ErrStop` can be used to halt `CommitIter.ForEach` iteration early | Architecture Patterns #4 | History log loads all commits instead of stopping at limit; memory waste but not a correctness bug |
| A2 | `relativeTime` helper is sufficient without a library | Don't Hand-Roll | Would need to add a dependency like `dustin/go-humanize`; low risk |
| A3 | go-git v5.17.0 and v5.17.2 APIs are identical for PlainOpen/Status/Log | Standard Stack | No behavior difference for the operations used; confirmed same minor version branch |

---

## Open Questions

1. **Cross-file search result navigation (GIT-03)**
   - What we know: Results must show `filename > keypath` items in a searchable list
   - What's unclear: When user selects a cross-file result, do we navigate to the file and highlight the key? This requires `DetailModel` to accept a "scroll-to-key" hint.
   - Recommendation: For v1, navigate to the file (open detail view). Key highlighting is a v2 enhancement. This matches the k9s pattern where search narrows the list but doesn't scroll the detail view.

2. **Git status async timing with file discovery**
   - What we know: D-11 says git status is fetched on startup during file discovery
   - What's unclear: Should git status be fetched in the same goroutine as `sops.Discover` or in a separate `tea.Cmd`?
   - Recommendation: Two-phase: `FilesDiscoveredMsg` fires first (fast, just file list), then a separate `GitStatusMsg` fires after. This keeps the file list responsive if git status is slow on a large repo. Both are `tea.Cmd` closures batched in `Init()`.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| git | go-git (PlainOpen walks up for .git) | n/a | n/a | go-git is pure Go, no git binary needed |
| wl-copy / wl-paste | atotto/clipboard (Wayland) | wl-copy: ✓ | wl-tools 2024 | xclip/xsel (X11); clipboard.Unsupported=true if neither |
| xclip | atotto/clipboard (X11 fallback) | ✗ | — | wl-copy (detected first via WAYLAND_DISPLAY) |
| xsel | atotto/clipboard (X11 fallback) | ✗ | — | wl-copy (detected first via WAYLAND_DISPLAY) |
| Go 1.26.2 | Build | ✓ | 1.26.2 | — |

**Dev machine is Wayland** (`WAYLAND_DISPLAY=wayland-0`), `wl-copy` available. Clipboard operations will work in development. README must document that users need either `xclip`, `xsel`, or `wl-clipboard` depending on their display server.

**Missing dependencies with no fallback:** None that block execution. `clipboard.Unsupported` flag provides graceful degradation.

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib testing + stretchr/testify v1.11.1 |
| Config file | none (go test ./...) |
| Quick run command | `go test ./internal/ui/... ./internal/app/... ./internal/git/... -run TestClipboard\|TestGit\|TestHistory -v` |
| Full suite command | `go test ./... -count=1` |

### Phase Requirements to Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CLB-01 | ctrl+y on revealed leaf copies value | unit (AppModel) | `go test ./internal/app/... -run TestCopyToClipboard -v` | No — Wave 0 |
| CLB-01 | ctrl+y on masked leaf is no-op, flashes "Reveal first with r" | unit (AppModel) | `go test ./internal/app/... -run TestCopyBlockedWhenMasked -v` | No — Wave 0 |
| CLB-02 | ClipboardClearMsg with matching gen clears hot state | unit (AppModel) | `go test ./internal/app/... -run TestClipboardAutoClear -v` | No — Wave 0 |
| CLB-02 | Stale ClipboardClearMsg (wrong gen) is ignored | unit (AppModel) | `go test ./internal/app/... -run TestClipboardStaleGen -v` | No — Wave 0 |
| CLB-03 | `defer clipboard.WriteAll("")` compiles and is reachable | compile check | `go build ./...` | n/a |
| GIT-01 | FileItem.Title() appends [M] badge when GitStatus="M" | unit (ui) | `go test ./internal/ui/... -run TestFileItemGitBadge -v` | No — Wave 0 |
| GIT-01 | StatusBarModel renders "no git" when GitAvailable=false | unit (ui) | `go test ./internal/ui/... -run TestStatusBarNoGit -v` | No — Wave 0 |
| GIT-02 | HistoryModel.View() renders commit entries in expected format | unit (ui) | `go test ./internal/ui/... -run TestHistoryModelView -v` | No — Wave 0 |
| GIT-02 | GetFileHistory returns ErrRepositoryNotExists for non-git dir | unit (git pkg) | `go test ./internal/git/... -run TestGetFileHistoryNoGit -v` | No — Wave 0 |
| GIT-03 | SearchResult.Title() renders "file > key.path" format | unit (ui/app) | `go test ./internal/... -run TestCrossFileSearchResult -v` | No — Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./... -count=1`
- **Per wave merge:** `go test ./... -count=1 -race`
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `internal/git/status_test.go` — covers GIT-01, GIT-02 git package tests
- [ ] `internal/ui/history_test.go` — covers GIT-02 HistoryModel rendering
- [ ] AppModel clipboard test helpers in `internal/app/model_test.go` — covers CLB-01, CLB-02
- [ ] FileItem badge tests in `internal/ui/filelist_test.go` — covers GIT-01 badge rendering

---

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | n/a |
| V3 Session Management | no | n/a |
| V4 Access Control | no | n/a |
| V5 Input Validation | yes | `SOPS_TUI_CLIPBOARD_TIMEOUT` env var parsed with `strconv.Atoi`; invalid values fall back to 30s |
| V6 Cryptography | no | Clipboard clears in memory; no crypto operations added |

### Known Threat Patterns

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Clipboard sniffing (other processes read clipboard before auto-clear) | Information Disclosure | 30s timeout (D-05) + clear on exit (D-07); document in README that clipboard access is inherently visible to X11/Wayland clients |
| Clipboard not cleared on crash/OOM kill | Information Disclosure | `defer` covers panic paths in normal Go flow; hard kills (SIGKILL, OOM killer) cannot be intercepted by any program — document as known limitation |
| Git history leaking commit messages to shoulder surfers | Information Disclosure | History overlay is full-screen and requires deliberate `b` keypress; values remain encrypted in git history |

---

## Sources

### Primary (HIGH confidence)
- [VERIFIED: go module cache] `github.com/atotto/clipboard@v0.1.4` — `clipboard_unix.go` read directly; confirmed `WriteAll`, `ReadAll`, `Unsupported` var, Wayland detection at init()
- [VERIFIED: go module cache] `github.com/go-git/go-git/v5@v5.17.0` — `status.go`, `options.go`, `plumbing/object/commit.go`, `plumbing/object/object.go` read directly
- [VERIFIED: codebase] `internal/ui/statusbar.go` — Flash() + generation counter pattern confirmed
- [VERIFIED: codebase] `internal/ui/metadata.go` — HistoryModel overlay template confirmed
- [VERIFIED: codebase] `internal/keys/bindings.go` — DetailKeyMap structure confirmed; ctrl+u/ctrl+d bindings confirm chord pattern works
- [VERIFIED: codebase] `internal/app/model.go` — sessionState enum, prevState pattern, async tea.Cmd pattern confirmed
- [VERIFIED: codebase] `internal/sops/discoverer.go` — DiscoveredFile struct confirmed; filepath.ToSlash usage confirmed
- [VERIFIED: go module cache] Go stdlib `os/signal` — `signal.NotifyContext` available since Go 1.16; confirmed present in Go 1.26.2

### Secondary (MEDIUM confidence)
- [CITED: github.com/go-git/go-git/v5 README] `git.ErrRepositoryNotExists` sentinel for non-git directories
- [VERIFIED: codebase] `go.mod` — atotto/clipboard v0.1.4 confirmed as indirect dep; go-git v5.17.0 added during research session

### Tertiary (LOW confidence)
- [ASSUMED] `storer.ErrStop` halts `CommitIter.ForEach` — standard go-git convention from examples; not verified by reading storer package source

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all packages verified in local module cache
- Architecture: HIGH — all patterns derived from existing codebase code, not assumptions
- Pitfalls: HIGH — clipboard/Unsupported and go-git/ErrRepositoryNotExists verified from source
- Cross-file search: MEDIUM — sahilm/fuzzy usage verified, aggregation strategy is reasoned design

**Research date:** 2026-04-15
**Valid until:** 2026-05-15 (stable libraries; bubbletea v2 API unlikely to change in 30 days)
