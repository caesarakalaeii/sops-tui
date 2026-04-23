# Architecture Patterns

**Domain:** Go TUI — SOPS secret management tool
**Researched:** 2026-04-13
**Confidence:** HIGH (Bubble Tea patterns verified via official docs and production codebases; SOPS structure verified via getsops/sops source)

---

## Recommended Architecture

### Model

The root application model (`AppModel`) is the single source of truth. It holds:
- Active view identifier (enum)
- All child view models (file browser, secret viewer, editor, diff, recipient manager)
- A command palette model (overlay)
- Global state: current file path, current key path, decrypted cache, status/error message, terminal dimensions

The root model delegates rendering and input to whichever child is active. This follows the production pattern used by Charmbracelet's own `crush` tool (appModel as central message router).

### View Stack vs. Static Registration

Use **static registration** (all views created at startup, root model holds all of them) rather than a dynamic stack. Rationale:

- SOPS-TUI has a fixed, known set of views — not unbounded like a plugin system.
- Static registration avoids re-initialization cost when switching back to a view.
- Simplifies `WindowSizeMsg` broadcast — root model forwards it to all child models unconditionally.
- The model-stack (bubblon) pattern is better suited when views are dynamically created at runtime in unknown quantity.

For the command palette (overlay), use the **foreground/background composite** approach: the overlay model wraps the current active view as background and renders the palette on top via `lipgloss` positioning.

---

## Go Project Layout

```
sops-tui/
├── cmd/
│   └── sops-tui/
│       └── main.go          # Cobra root command, program bootstrap only
├── internal/
│   ├── app/
│   │   └── app.go           # AppModel: root tea.Model, view routing
│   ├── ui/
│   │   ├── filebrowser/     # View: file list from .sops.yaml discovery
│   │   ├── secretviewer/    # View: key tree for one file, masked/revealed values
│   │   ├── editor/          # View: edit a single secret value (textarea)
│   │   ├── diffview/        # View: before/after diff before re-encrypt
│   │   ├── recipients/      # View: add/remove age recipients
│   │   ├── palette/         # Overlay: command palette (fuzzy match commands)
│   │   └── statusbar/       # Component: bottom status bar (path, key, mode)
│   ├── sops/
│   │   ├── runner.go        # exec.Command wrapper: decrypt, encrypt, rotate, rekey
│   │   ├── config.go        # .sops.yaml parser: path_regex rules, age key refs
│   │   ├── parser.go        # YAML/JSON parser: extract key tree, sops metadata
│   │   └── age.go           # age keys.txt reader (~/.config/sops/age/keys.txt)
│   ├── keys/
│   │   └── bindings.go      # Key binding definitions (vim + k9s-style)
│   └── styles/
│       └── theme.go         # lipgloss styles and colour palette
├── go.mod
└── go.sum
```

**Rules:**
- `cmd/` contains only `main.go` — no business logic.
- `internal/` is the entire application. Nothing is exported from `internal/`.
- No `pkg/` directory at v1 — there is nothing to share externally yet. Add if a public SOPS parsing library becomes a goal.

---

## Component Boundaries

| Component | Owns | Communicates With | Does NOT Know About |
|-----------|------|-------------------|---------------------|
| `AppModel` | View routing, global state, terminal size | All child models (delegates Update/View), `sops.Runner` (issues Cmds) | SOPS subprocess details |
| `filebrowser` | File list, cursor position, search filter | AppModel (sends `FileSelectedMsg`) | sops subprocess |
| `secretviewer` | Key tree display, masked/revealed state per key | AppModel (sends `KeySelectedMsg`, `EditRequestMsg`) | How decryption happens |
| `editor` | `bubbles/textarea`, staged new value | AppModel (sends `SaveRequestMsg`, `CancelMsg`) | Encryption |
| `diffview` | Renders before/after diff string | AppModel (sends `ConfirmEncryptMsg`, `CancelMsg`) | How diff is computed |
| `recipients` | Age key list per file | AppModel (sends `RekeyRequestMsg`) | Subprocess details |
| `palette` | Command list, fuzzy filter, `bubbles/textinput` | AppModel (sends `CommandSelectedMsg`) | Any domain logic |
| `statusbar` | Read-only display component | Receives props from AppModel.View() | Nothing — it's pure View |
| `sops.Runner` | `exec.Command` lifecycle | Returns `tea.Cmd` funcs; sends result Msgs to AppModel | Tea framework internals |
| `sops.Config` | `.sops.yaml` parse result | Called by AppModel at startup | TUI |
| `sops.Parser` | Key tree extraction from raw YAML bytes | Called by `secretviewer` after decrypt | Subprocess |

---

## Data Flow

### Application Start

```
main.go
  └─► cobra.Command.Execute()
        └─► sops.Config.Load()        // parse .sops.yaml, find all matching files
              └─► AppModel{files}
                    └─► tea.NewProgram(AppModel).Run()
                          └─► AppModel.Init()
                                └─► Cmd: sops.CheckInstalled()   // verify sops binary
```

### File Browse → Decrypt → View

```
User: ↓/↑ in filebrowser
  └─► AppModel.Update(KeyMsg)
        └─► filebrowser.Update(KeyMsg) → filebrowser model updated

User: Enter on file
  └─► filebrowser sends FileSelectedMsg{path}
        └─► AppModel.Update(FileSelectedMsg)
              ├─► AppModel.activeView = secretviewer
              ├─► secretviewer.SetFile(path)
              └─► Cmd: sops.Runner.DecryptKeys(path)   // tea.Cmd → goroutine

                   [goroutine: exec.Command("sops", "--decrypt", path)]
                          └─► DecryptResultMsg{keyTree, raw, err}
                                └─► AppModel.Update(DecryptResultMsg)
                                      └─► secretviewer.SetTree(keyTree)
                                            └─► secretviewer.View() renders tree
```

### Edit → Diff → Re-encrypt

```
User: e on selected key in secretviewer
  └─► secretviewer sends EditRequestMsg{file, keyPath, currentValue}
        └─► AppModel.Update(EditRequestMsg)
              └─► AppModel.activeView = editor
                    └─► editor.SetContext(file, keyPath, currentValue)

User: writes new value, presses Ctrl+S
  └─► editor sends SaveRequestMsg{newValue}
        └─► AppModel.Update(SaveRequestMsg)
              ├─► AppModel.activeView = diffview
              ├─► Cmd: sops.Runner.PreviewEdit(file, keyPath, newValue)
              └─►      // runs sops decrypt, injects new value, diffs output

                   DiffReadyMsg{before, after}
                     └─► diffview.SetDiff(before, after)

User: confirms diff
  └─► diffview sends ConfirmEncryptMsg
        └─► AppModel.Update(ConfirmEncryptMsg)
              └─► Cmd: sops.Runner.ApplyEdit(file, keyPath, newValue)
                          └─► EncryptDoneMsg{err}
                                └─► AppModel returns to secretviewer, shows status
```

### Command Palette

```
User: presses : or Ctrl+P
  └─► AppModel.Update(KeyMsg)
        └─► AppModel.paletteOpen = true
              └─► palette.Focus()

                   View(): lipgloss overlay composite
                     background = currentActiveView.View()
                     foreground = palette.View()
                     result = overlay(background, foreground)

User: selects command
  └─► palette sends CommandSelectedMsg{command}
        └─► AppModel dispatches command action
              └─► AppModel.paletteOpen = false
```

---

## Bubble Tea Model Composition Pattern

Each child model is a standalone `tea.Model`:

```go
// Each view implements this interface
type View interface {
    tea.Model
    // Init() tea.Cmd
    // Update(tea.Msg) (tea.Model, tea.Cmd)
    // View() string
}
```

Root model delegates input and rendering:

```go
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    case tea.WindowSizeMsg:
        // Broadcast to ALL child models — every view needs dimensions
        var cmds []tea.Cmd
        m.fileBrowser, cmd = m.fileBrowser.Update(msg); cmds = append(cmds, cmd)
        m.secretViewer, cmd = m.secretViewer.Update(msg); cmds = append(cmds, cmd)
        // ... all views
        return m, tea.Batch(cmds...)

    case FileSelectedMsg:
        // Domain message — AppModel handles routing
        m.activeView = ViewSecretViewer
        m.secretViewer.SetFile(msg.Path)
        return m, sops.DecryptKeysCmd(msg.Path)

    case DecryptResultMsg:
        // Route result to the right child
        m.secretViewer, cmd = m.secretViewer.Update(msg)
        return m, cmd

    default:
        // Forward all other messages to the active view
        if m.paletteOpen {
            m.palette, cmd = m.palette.Update(msg)
        } else {
            m.activeView.model, cmd = m.activeView.model.Update(msg)
        }
        return m, cmd
    }
}
```

Root View composites all layers:

```go
func (m AppModel) View() string {
    content := m.activeModel().View()
    withStatus := lipgloss.JoinVertical(lipgloss.Left, content, m.statusBar.View())
    if m.paletteOpen {
        return overlay.Render(withStatus, m.palette.View())
    }
    return withStatus
}
```

---

## Async Operations: Handling Slow sops Subprocess Calls

`sops --decrypt` on a large file with a remote age key can take 1–3 seconds. The Bubble Tea command pattern handles this correctly — **never block Update()**.

### Pattern: Cmd returns a domain message

```go
// internal/sops/runner.go
func DecryptKeysCmd(path string) tea.Cmd {
    return func() tea.Msg {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        cmd := exec.CommandContext(ctx, "sops", "--decrypt", path)
        out, err := cmd.Output()
        if err != nil {
            return DecryptErrMsg{Path: path, Err: err}
        }
        tree, parseErr := parser.ParseKeyTree(out)
        if parseErr != nil {
            return DecryptErrMsg{Path: path, Err: parseErr}
        }
        return DecryptResultMsg{Path: path, Tree: tree, Raw: out}
    }
}
```

### Loading State

AppModel tracks per-operation loading state. The active view's `View()` renders a spinner (from `bubbles/spinner`) while the Cmd is in-flight:

```go
type AppModel struct {
    loading     bool
    loadingMsg  string
    spinner     spinner.Model
    // ...
}
```

On `FileSelectedMsg`: set `loading=true`, return `tea.Batch(DecryptKeysCmd(path), spinner.Tick)`.
On `DecryptResultMsg` or `DecryptErrMsg`: set `loading=false`.

### Cancellation

If the user navigates away before decrypt completes, the in-flight Cmd cannot be cancelled via Bubble Tea's model (commands are fire-and-forget). Accept this: the result message will arrive and AppModel will discard it if the file context has changed (check `msg.Path == m.currentFile`).

For Go 1.20+, use `cmd.WaitDelay` to ensure process group cleanup on timeout, not just the top-level sops process.

---

## SOPS File Discovery

`.sops.yaml` uses `path_regex` (not glob) for matching rules. The discovery flow:

```
sops.Config.Load():
  1. Walk upward from cwd to find .sops.yaml
  2. Parse YAML: extract creation_rules[].path_regex fields
  3. Walk the repository tree (filepath.WalkDir from .sops.yaml location)
  4. For each file: test each path_regex in order, first match wins
  5. Check file contains a "sops:" top-level key (cheap heuristic: grep bytes)
  6. Return []DiscoveredFile{Path, RuleName, AgeRecipients}
```

The "does this file contain sops metadata" check should read only the first 4KB to look for the `sops:` key — avoid decrypting at discovery time.

### SOPS Metadata in YAML

A SOPS-encrypted YAML file has its values replaced with ENC[AES256_GCM,...] ciphertext. The `sops:` top-level key contains:

```yaml
sops:
  age:
    - recipient: age1...
      enc: |
        -----BEGIN AGE ENCRYPTED FILE-----
        ...
  lastmodified: "2024-01-01T00:00:00Z"
  mac: ENC[AES256_GCM,...]
  version: 3.8.1
```

The TUI can display keys and their encrypted-value prefix without decrypting. Full values require `sops --decrypt`.

---

## Layout: Lipgloss Composition

The terminal layout uses `lipgloss.JoinHorizontal` and `lipgloss.JoinVertical` — no tview or tcell grid primitives.

```
┌─────────────────────────────────────────────────┐
│  file browser (30%)  │  secret viewer (70%)      │
│  secrets/            │  > database               │
│  > app.yaml          │    password  ENC[...]      │
│    db.yaml           │    host      db.example    │
│                      │  > api                    │
│                      │    key       ENC[...]      │
├─────────────────────────────────────────────────┤
│  app.yaml > database > password    [ENCRYPTED]  │
└─────────────────────────────────────────────────┘
                  status bar
```

The file browser and secret viewer use `bubbles/list` and a custom tree renderer respectively. On `WindowSizeMsg`, widths are recalculated as percentage fractions.

---

## Key Bindings (k9s-style)

Key bindings live in `internal/keys/bindings.go` using `bubbles/key`:

```go
type GlobalKeyMap struct {
    Quit       key.Binding  // q, Ctrl+C
    Help       key.Binding  // ?
    Palette    key.Binding  // :, Ctrl+P
    Search     key.Binding  // /
    Back       key.Binding  // Esc
}

type SecretViewerKeyMap struct {
    Reveal     key.Binding  // Space, Enter
    Edit       key.Binding  // e
    Copy       key.Binding  // y (yank, vim-style)
    Rotate     key.Binding  // r
    Recipients key.Binding  // R
}
```

Global bindings are checked in AppModel.Update() before delegating to the active view.

---

## Suggested Build Order (Phase Dependencies)

The following order respects hard dependencies between components:

1. **SOPS subprocess wrapper** (`internal/sops/runner.go`)
   - All views depend on decrypt/encrypt results as messages.
   - Build and test in isolation with mock responses before any TUI.

2. **SOPS config discovery** (`internal/sops/config.go`, `parser.go`)
   - File browser needs the discovered file list.
   - Can be built and unit-tested independently.

3. **Root AppModel skeleton** (`internal/app/app.go`)
   - Establish view enum, message types, routing structure.
   - Stub all child models to return placeholder View() strings.

4. **File browser view** (`internal/ui/filebrowser/`)
   - First interactive view. Uses `bubbles/list`.
   - Unblocks the full startup→browse flow.

5. **Secret viewer view** (`internal/ui/secretviewer/`)
   - Depends on SOPS parser output (key tree).
   - Integrates the first real decrypt Cmd flow.

6. **Status bar** (`internal/ui/statusbar/`)
   - Stateless component, trivial to add once views exist.

7. **Editor view** (`internal/ui/editor/`)
   - Depends on secret viewer (edit request originates there).
   - Uses `bubbles/textarea`.

8. **Diff view** (`internal/ui/diffview/`)
   - Depends on editor (precedes re-encrypt confirmation).
   - Pure renderer — no subprocess calls itself.

9. **Recipient manager** (`internal/ui/recipients/`)
   - Depends on config (age key list) and runner (rekey Cmd).
   - Largely independent of viewer/editor flow.

10. **Command palette overlay** (`internal/ui/palette/`)
    - Depends on all views being defined (palette lists commands from all contexts).
    - Uses `bubbles/textinput` for fuzzy filter.

---

## Anti-Patterns to Avoid

### Synchronous sops calls in Update()
**What goes wrong:** `Update()` blocks waiting for subprocess. Terminal freezes.
**Instead:** Always wrap `exec.Command` in a `tea.Cmd` returning a message.

### Global decrypted cache without eviction
**What goes wrong:** Decrypted secrets accumulate in memory for the process lifetime.
**Instead:** Store decrypted values in a `map[string]DecryptedFile` with explicit eviction when navigating away. Zero the map entry on view change.

### Passing full AppModel into child models
**What goes wrong:** Circular dependency, child models can mutate parent state directly.
**Instead:** Child models receive only the data they need via typed messages. They emit typed request messages upward; AppModel acts on them.

### One mega model with all state
**What goes wrong:** `Update()` becomes a 500-line switch. Impossible to test views in isolation.
**Instead:** Each view is its own `tea.Model` with its own `Update()`. Root model's Update only handles routing and global keys.

### Using goroutines directly in Update()
**What goes wrong:** Race conditions on model state. Bubble Tea's event loop is single-threaded by design.
**Instead:** All concurrency goes through `tea.Cmd`. Never call `go func()` inside Update().

---

## Scalability Considerations

| Concern | At 10 files | At 100 files | At 1000 files |
|---------|-------------|--------------|---------------|
| File discovery | Eager scan, trivial | Eager scan, ~50ms | Add lazy pagination or virtual scroll in filebrowser |
| Decrypted cache | In-memory map, trivial | In-memory, watch RSS | LRU eviction, configurable limit |
| Key tree display | Full tree rendered | Full tree rendered | Virtualised list (bubbles/list handles this) |
| sops subprocess | One at a time, fine | Concurrent per-file batch | Rate-limit concurrent subprocesses (semaphore) |

---

## Sources

- [Managing nested models with Bubble Tea — Roman Parykin](https://donderom.com/posts/managing-nested-models-with-bubble-tea/)
- [Tips for building Bubble Tea programs — leg100](https://leg100.github.io/en/posts/building-bubbletea-programs/)
- [Commands in Bubble Tea — Charm official blog](https://charm.land/blog/commands-in-bubbletea/)
- [Concurrency and Goroutines — Bubble Tea DeepWiki](https://deepwiki.com/charmbracelet/bubbletea/5.1-concurrency-and-goroutines)
- [TUI Architecture: appModel — charmbracelet/crush DeepWiki](https://deepwiki.com/charmbracelet/crush/5.1-tui-architecture-and-appmodel)
- [Multi-model message routing discussion — charmbracelet/bubbletea #751](https://github.com/charmbracelet/bubbletea/discussions/751)
- [The Bubbletea State Machine pattern — Zack Proser](https://zackproser.com/blog/bubbletea-state-machine)
- [getsops/sops stores package — pkg.go.dev](https://pkg.go.dev/github.com/getsops/sops/v3/stores)
- [bubbletea-overlay package — pkg.go.dev](https://pkg.go.dev/github.com/quickphosphat/bubbletea-overlay)
- [charmbracelet/bubbles — GitHub](https://github.com/charmbracelet/bubbles)
- [Advanced command execution in Go with os/exec — kowalczyk](https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html)
- [Standard Go Project Layout — golang-standards](https://github.com/golang-standards/project-layout)
