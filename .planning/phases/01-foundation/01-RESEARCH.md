# Phase 1: Foundation - Research

**Researched:** 2026-04-14
**Domain:** Go + Bubbletea v2 TUI skeleton, startup validation, vim navigation, contextual help, status bar
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** Use lipgloss-styled stderr for missing-dependency errors — render a colored, bordered box to stderr without initializing a TUI session. Exit with non-zero code. Scriptable and CI-friendly.
- **D-02:** Run all validation checks together and report all issues in a single error box. User fixes everything in one pass.
- **D-03:** `sops` binary missing is a hard error (exit). Age key missing is a soft warning — TUI launches but decryption will be unavailable. This allows browsing encrypted files without a key.
- **D-04:** Validate `.sops.yaml` at startup — check cwd and parent dirs. If missing, show styled warning but still launch (the skeleton has nothing to browse yet).
- **D-05:** Single-pane layout with drill-down navigation. File list takes the full terminal width. Selecting a file replaces the view with a detail view. Esc returns to the file list. Like k9s resource drill-down.
- **D-06:** Detail view renders content as a YAML-style tree with indentation preserving the original YAML structure. Nested keys displayed as a tree hierarchy.
- **D-07:** Deeply nested YAML nodes are collapsible — groups can be collapsed/expanded with Enter or arrow keys. [+]/[-] indicators show collapse state.
- **D-08:** Help is a full-screen overlay toggled with `?`. Replaces current content entirely. Press `?` or `Esc` to close. Like k9s help screen.
- **D-09:** Help content is contextual — shows only keybindings relevant to the current view (file list, detail, search). Global keys (`?`, `q`, `/`) appear in every context.
- **D-10:** Single-line status bar at the bottom of the terminal. Standard TUI convention (k9s, vim, lazygit).
- **D-11:** Status bar shows: left = current path/view breadcrumb, center = item count, right = environment status indicators (sops/age/.sops.yaml availability with checkmarks or warnings). Content adapts per view.
- **D-12:** Transient feedback (e.g., "Copied to clipboard") uses flash messages — temporarily replaces status bar content for 2-3 seconds, then restores normal content.

### Claude's Discretion

- Exact lipgloss color palette and border styles
- Internal component architecture and state management patterns
- Flash message duration (2-3 second range)
- YAML tree indentation width and collapse/expand key mappings (h/l or Enter)

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| HLT-01 | User sees startup error with instructions if `sops` binary is missing | exec.LookPath pattern; lipgloss styled stderr box; hard exit code 1 |
| HLT-02 | User sees startup error with instructions if age key file is missing | os.Stat on ~/.config/sops/age/keys.txt; age.ParseIdentities; soft warning — TUI still launches |
| NAV-03 | User can navigate with vim keybindings (hjkl, g/G, ctrl-d/u) | bubbles/v2 viewport has these bindings built in; list component supports j/k/g/G; key.Binding pattern with tea.KeyPressMsg |
| NAV-05 | User can view contextual help panel with `?` | bubbles/v2 help component; full-screen overlay via view state enum; KeyMap interface with ShortHelp/FullHelp |
| NAV-06 | User sees persistent status bar (file path, encryption status, operation feedback) | lipgloss JoinHorizontal pattern; tea.WindowSizeMsg for width; tea.Tick for flash message timer |
</phase_requirements>

---

## Summary

Phase 1 delivers the TUI skeleton: startup validation, a navigable file-list view, a placeholder YAML-tree detail view, a full-screen help overlay, and a persistent status bar. No decryption occurs. The phase establishes every architectural pattern that later phases extend.

The technology is fully locked in CLAUDE.md: `charm.land/bubbletea/v2`, `charm.land/lipgloss/v2`, `charm.land/bubbles/v2`. All three are stable production-quality libraries with well-documented v2 APIs. The bubbletea v2 upgrade guide and pkg.go.dev docs are the authoritative sources — training knowledge was verified against both.

The biggest implementation risks are: (1) the bubbletea v2 API breaks from v1 are pervasive and well-documented — any copy-paste from v1 examples will fail; (2) `lipgloss.AdaptiveColor` must never be used (confirmed hang issue #1036); (3) the startup error box must write to stderr via `lipgloss.Fprintln(os.Stderr, ...)` without starting a TUI session.

**Primary recommendation:** Implement in a single `main.go` entry point that runs validation, then constructs and runs a `tea.Program`. Use a `sessionState` enum in the root model to route messages and compose views across file-list, detail, and help states. Use `bubbles/v2/list` for file list, `bubbles/v2/viewport` for YAML tree scrolling, `bubbles/v2/help` for keybinding display, and `tea.Tick` for flash message timeout.

---

## Project Constraints (from CLAUDE.md)

| Directive | Category | Enforcement |
|-----------|----------|-------------|
| Use `charm.land/bubbletea/v2` import path (NOT `github.com/charmbracelet/bubbletea`) | Import path | Build fails if wrong path used |
| `View()` returns `tea.View` struct, not `string` | API contract | Compile error if wrong |
| Use `tea.KeyPressMsg` not `tea.KeyMsg` for key handling | API contract | Wrong type silently drops all key events |
| `msg.Code` (rune), `msg.Text` (string), `msg.Mod` for modifier | API contract | v1 field names cause compile errors |
| Space key: `"space"` not `" "` in `msg.String()` | Keybinding | Silent bug — space never matches |
| `view.AltScreen = true` (View field), not `tea.WithAltScreen()` | Program options | v1 option silently ignored |
| Never use `lipgloss.AdaptiveColor` | Color | Confirmed hang (issue #1036) |
| Use explicit hex colors throughout | Color | Required per CLAUDE.md and UI-SPEC |
| `goccy/go-yaml` not `gopkg.in/yaml.v3` | Dependency | Project constraint |
| `exec.CommandContext` with timeout for sops subprocess | Subprocess | Required; bare `exec.Command` lacks cancellation |
| `CGO_ENABLED=0` for all release builds | Build | Stated in CLAUDE.md §go.mod Notes |
| Never use `type any` | Go typing | CLAUDE.md global rule |

---

## Standard Stack

### Core

| Library | Version | Import Path | Purpose | Why Standard |
|---------|---------|-------------|---------|--------------|
| bubbletea | v2.0.4 | `charm.land/bubbletea/v2` | TUI event loop, Model/Update/View | Only Elm-architecture TUI for Go; v2 stable Apr 2026 |
| lipgloss | v2.x | `charm.land/lipgloss/v2` | Layout, color, borders, styling | Required by bubbletea v2; pure rendering (no I/O conflicts) |
| bubbles | v2.x | `charm.land/bubbles/v2` | list, viewport, help, spinner components | Official Charm component library for bubbletea v2 |

### Supporting (Phase 1)

| Library | Version | Import Path | Purpose | When to Use |
|---------|---------|-------------|---------|-------------|
| filippo.io/age | v1.3.1 | `filippo.io/age` | Parse keys.txt, detect key availability | Startup validation only in Phase 1 |
| os/exec | stdlib | `os/exec` | exec.LookPath for sops binary detection | Startup validation |
| os | stdlib | `os` | os.Stat for age keys.txt existence | Startup validation |
| stretchr/testify | v1.x | `github.com/stretchr/testify` | Test assertions (require/assert) | All test files |
| charmbracelet/x/exp/teatest | latest | `github.com/charmbracelet/x/exp/teatest` | TUI golden file tests | TUI integration tests |
| sebdah/goldie | v2.x | `github.com/sebdah/goldie/v2` | Golden file management | Alongside teatest |

### Not Used in Phase 1

| Library | Deferred To |
|---------|-------------|
| goccy/go-yaml | Phase 2 (file browsing needs YAML parsing) |
| sahilm/fuzzy | Phase 2 (fuzzy search) |
| atotto/clipboard | Phase 4 |
| go-git/go-git | Phase 4 |
| charm.land/huh/v2 | Phase 3+ (multi-field forms) |

### Installation

```bash
go get charm.land/bubbletea/v2@v2.0.4
go get charm.land/lipgloss/v2
go get charm.land/bubbles/v2
go get filippo.io/age@v1.3.1
go get github.com/stretchr/testify@v1.x
go get github.com/charmbracelet/x/exp/teatest@latest
go get github.com/sebdah/goldie/v2
```

**Version verification:** [VERIFIED: pkg.go.dev] — bubbletea v2.0.4 published April 13 2026; filippo.io/age v1.3.1 published Dec 2025; teatest last published Feb 16 2026.

---

## Architecture Patterns

### Recommended Project Structure

```
cmd/
└── sops-tui/
    └── main.go           # Entry point: validate, run program
internal/
├── app/
│   ├── model.go          # Root tea.Model: sessionState enum, child models
│   ├── update.go         # Update() dispatch by sessionState
│   └── view.go           # View() compose: route to child or overlay
├── validator/
│   └── startup.go        # RunChecks() → []ValidationResult
├── ui/
│   ├── styles.go         # All lipgloss styles (const colors, named styles)
│   ├── errorbox.go       # Stderr error/warning box renderer
│   ├── statusbar.go      # StatusBar model: flash timer, env indicators
│   ├── help.go           # HelpModel: wraps bubbles/help, contextual KeyMaps
│   ├── filelist.go       # FileListModel: wraps bubbles/list
│   └── detail.go         # DetailModel: YAML tree with viewport
└── keys/
    └── bindings.go       # key.Binding definitions, KeyMap structs per view
```

This structure keeps the TUI models flat enough for Phase 1 while creating the packages that Phases 2-5 extend without moving files.

### Pattern 1: Root Model with sessionState Enum

**What:** A single root `tea.Model` holds a `sessionState` (int type alias) that controls which child model receives updates and renders. No external router library.

**When to use:** Any time the same program hosts multiple full-screen "views" that swap out completely. Standard Bubbletea pattern.

```go
// Source: https://github.com/charmbracelet/bubbletea/blob/main/examples/composable-views/main.go (adapted to v2)
type sessionState int

const (
    stateFileList sessionState = iota
    stateDetail
    stateHelp
)

type AppModel struct {
    state    sessionState
    prevState sessionState  // for help overlay to know where to return
    width    int
    height   int
    fileList ui.FileListModel
    detail   ui.DetailModel
    help     ui.HelpModel
    status   ui.StatusBarModel
}

func (m AppModel) Init() tea.Cmd {
    return tea.RequestWindowSize
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        // propagate to children that need dimensions
        return m, nil
    case tea.KeyPressMsg:
        switch msg.String() {
        case "?":
            if m.state == stateHelp {
                m.state = m.prevState
            } else {
                m.prevState = m.state
                m.state = stateHelp
            }
            return m, nil
        case "q", "ctrl+c":
            return m, tea.Quit()
        case "esc":
            if m.state == stateHelp {
                m.state = m.prevState
            } else if m.state == stateDetail {
                m.state = stateFileList
            }
            return m, nil
        }
    }
    // route remaining messages to active child
    switch m.state {
    case stateFileList:
        var cmd tea.Cmd
        m.fileList, cmd = m.fileList.Update(msg)
        return m, cmd
    case stateDetail:
        var cmd tea.Cmd
        m.detail, cmd = m.detail.Update(msg)
        return m, cmd
    }
    return m, nil
}

func (m AppModel) View() tea.View {
    var content string
    switch m.state {
    case stateFileList:
        content = m.fileList.View()
    case stateDetail:
        content = m.detail.View()
    case stateHelp:
        content = m.help.View(m.prevState)
    }
    // stack status bar at bottom
    statusBar := m.status.View(m.width)
    mainHeight := m.height - lipgloss.Height(statusBar)
    body := lipgloss.NewStyle().Height(mainHeight).Render(content)
    full := lipgloss.JoinVertical(lipgloss.Left, body, statusBar)
    v := tea.NewView(full)
    v.AltScreen = true
    return v
}
```

[VERIFIED: charm.land/bubbletea/v2 upgrade guide + pkg.go.dev/charm.land/bubbletea/v2]

### Pattern 2: Startup Validation with Lipgloss Stderr Box

**What:** Before creating any `tea.Program`, run all validation checks. Collect results. If any hard errors or soft warnings, render a styled box directly to `os.Stderr` using `lipgloss.Fprintln`. Hard error (sops missing) → `os.Exit(1)`. Soft warnings → print box then proceed to TUI.

**When to use:** Always in `main()` before `tea.NewProgram()`. Keeps stderr output independent of TUI alternate screen.

```go
// Source: pkg.go.dev/charm.land/lipgloss/v2 (Fprintln writes to any io.Writer)
// internal/validator/startup.go

type Severity int
const (
    SeverityError Severity = iota
    SeverityWarn
)

type ValidationResult struct {
    Severity Severity
    Message  string
    Fix      string
}

func RunChecks() ([]ValidationResult, bool) {
    var results []ValidationResult
    hasHard := false

    // Check sops binary
    if _, err := exec.LookPath("sops"); err != nil {
        results = append(results, ValidationResult{
            Severity: SeverityError,
            Message:  "sops binary not found",
            Fix:      "Install sops: https://github.com/getsops/sops#install",
        })
        hasHard = true
    }

    // Check age key (soft warning)
    keyPath := filepath.Join(os.UserHomeDir(), ".config", "sops", "age", "keys.txt")
    if _, err := os.Stat(keyPath); err != nil {
        results = append(results, ValidationResult{
            Severity: SeverityWarn,
            Message:  "Age key file not found",
            Fix:      "Create key: age-keygen -o ~/.config/sops/age/keys.txt",
        })
    }

    // Check .sops.yaml walk up from cwd (soft warning)
    if !findSopsYaml() {
        results = append(results, ValidationResult{
            Severity: SeverityWarn,
            Message:  ".sops.yaml not found in current directory or parents",
            Fix:      "Run sops-tui in a repository with a .sops.yaml configuration",
        })
    }

    return results, hasHard
}
```

```go
// internal/ui/errorbox.go
func RenderErrorBox(results []validator.ValidationResult, w io.Writer) {
    // build content lines
    var lines []string
    for _, r := range results {
        label := styles.ErrorLabel.Render("[ERROR]")
        if r.Severity == validator.SeverityWarn {
            label = styles.WarnLabel.Render("[WARN] ")
        }
        lines = append(lines, fmt.Sprintf("%s %s", label, r.Message))
        lines = append(lines, fmt.Sprintf("       %s", r.Fix))
        lines = append(lines, "")
    }
    content := strings.Join(lines, "\n")

    // border color: error red if any hard errors, else warning yellow
    borderColor := lipgloss.Color("#f38ba8") // ColorError
    for _, r := range results {
        if r.Severity == validator.SeverityError {
            break
        }
        borderColor = lipgloss.Color("#f9e2af") // ColorWarning
    }

    box := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(borderColor).
        Padding(1, 2).
        Width(min(termWidth()-4, 72)).
        Render(content)

    lipgloss.Fprintln(w, "sops-tui: startup failed")
    lipgloss.Fprintln(w, box)
}
```

[VERIFIED: pkg.go.dev/charm.land/lipgloss/v2 — Fprintln, RoundedBorder, explicit colors confirmed]

### Pattern 3: Status Bar with Flash Timer

**What:** Status bar is a separate model holding current env state and optional flash message. Flash uses `tea.Tick` to schedule clear after 2 seconds.

**When to use:** Every phase. The flash pattern generalizes to all future "operation feedback" in Phase 3+.

```go
// internal/ui/statusbar.go
type flashClearMsg struct{}

type StatusBarModel struct {
    width       int
    breadcrumb  string
    itemCount   int
    envSops     bool
    envAge      bool
    envSopsYaml bool
    flash       string // empty = not flashing
}

func (m StatusBarModel) Flash(msg string) (StatusBarModel, tea.Cmd) {
    m.flash = msg
    return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
        return flashClearMsg{}
    })
}

func (m StatusBarModel) Update(msg tea.Msg) (StatusBarModel, tea.Cmd) {
    switch msg.(type) {
    case flashClearMsg:
        m.flash = ""
    }
    return m, nil
}

func (m StatusBarModel) View(width int) string {
    if m.flash != "" {
        return lipgloss.NewStyle().
            Width(width).
            Background(lipgloss.Color("#313244")). // ColorSurface
            Foreground(lipgloss.Color("#cdd6f4")). // ColorFg
            Align(lipgloss.Center).
            Render(m.flash)
    }
    // three-section layout with JoinHorizontal
    left := renderBreadcrumb(m.breadcrumb)
    center := renderCount(m.itemCount)
    right := renderEnvIndicators(m.envSops, m.envAge, m.envSopsYaml)
    // fill remaining width with surface color background
    return lipgloss.NewStyle().Width(width).Background(lipgloss.Color("#313244")).
        Render(lipgloss.JoinHorizontal(lipgloss.Top, left, center, right))
}
```

[VERIFIED: pkg.go.dev/charm.land/bubbletea/v2 — tea.Tick confirmed; VERIFIED: lipgloss v2 — JoinHorizontal confirmed]

### Pattern 4: Contextual Help Overlay via bubbles/help

**What:** `bubbles/v2/help` renders a help view from any type implementing `KeyMap` (ShortHelp + FullHelp). Each view (fileList, detail) has its own `KeyMap` struct. The root model passes the appropriate KeyMap to `help.Model.View()` based on `m.prevState`.

**When to use:** `?` keypress from any view. Help is full-screen, not a popup — it replaces content in View().

```go
// keys/bindings.go
import "charm.land/bubbles/v2/key"

type FileListKeyMap struct {
    Up, Down    key.Binding
    GoTop, GoBottom key.Binding
    HalfUp, HalfDown key.Binding
    Open        key.Binding
    Help        key.Binding
    Quit        key.Binding
}

var DefaultFileListKeyMap = FileListKeyMap{
    Up: key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("↑/k", "move up")),
    Down: key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("↓/j", "move down")),
    GoTop: key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "go to top")),
    GoBottom: key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "go to bottom")),
    HalfUp: key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "half page up")),
    HalfDown: key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "half page down")),
    Open: key.NewBinding(key.WithKeys("enter", "l"), key.WithHelp("enter/l", "open")),
    Help: key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
    Quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

func (k FileListKeyMap) ShortHelp() []key.Binding { return []key.Binding{k.Up, k.Down, k.Help, k.Quit} }
func (k FileListKeyMap) FullHelp() [][]key.Binding {
    return [][]key.Binding{
        {k.Up, k.Down, k.GoTop, k.GoBottom, k.HalfUp, k.HalfDown},
        {k.Open, k.Help, k.Quit},
    }
}
```

```go
// In HelpModel.View(state sessionState) string:
func (m HelpModel) View(fromState sessionState) string {
    var km help.KeyMap
    switch fromState {
    case stateFileList:
        km = keys.DefaultFileListKeyMap
    case stateDetail:
        km = keys.DefaultDetailKeyMap
    }
    m.help.ShowAll = true
    content := m.help.View(km)
    // wrap in full-screen RoundedBorder box
    return lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("#6c7086")). // ColorMuted
        Background(lipgloss.Color("#313244")).       // ColorSurface
        Padding(1, 4).
        Width(m.width - 2).
        Height(m.height - 2).
        Render(content)
}
```

[VERIFIED: pkg.go.dev/charm.land/bubbles/v2/help — KeyMap interface, Model.View(), ShowAll confirmed]

### Pattern 5: Layout Height Calculation (lipgloss.Height)

**What:** Never hard-code the status bar height. Use `lipgloss.Height(statusBar)` to measure it, then subtract from `tea.WindowSizeMsg.Height` to get the available height for the main content area.

**Why:** If status bar styling adds borders or padding later, the calculation stays correct without code changes.

```go
// In View():
statusBar := m.status.View(m.width)
statusBarH := lipgloss.Height(statusBar)
mainH := m.height - statusBarH
body := lipgloss.NewStyle().Height(mainH).Render(mainContent)
full := lipgloss.JoinVertical(lipgloss.Left, body, statusBar)
```

[CITED: https://leg100.github.io/en/posts/building-bubbletea-programs/ — "use lipgloss's height/width methods rather than hard-coded arithmetic"]

### Anti-Patterns to Avoid

- **Using `tea.KeyMsg` instead of `tea.KeyPressMsg`:** `tea.KeyMsg` is now an interface in v2. Switching on it will match both press and release; use `tea.KeyPressMsg` to handle only presses.
- **Using `msg.Type`, `msg.Runes`, `msg.Alt`:** These fields do not exist in bubbletea v2. Use `msg.Code`, `msg.Text`, `msg.Mod` respectively.
- **Using `tea.WithAltScreen()` in program options:** Removed in v2. Set `view.AltScreen = true` on the returned `tea.View`.
- **Using `p.Start()`:** Use `p.Run()` in v2.
- **Calling `View()` returning string:** v2 `View()` must return `tea.View`, created via `tea.NewView(content)`.
- **Using `lipgloss.AdaptiveColor`:** Confirmed hang in bubbletea integration (issue #1036). Use explicit hex colors.
- **Blocking in Update():** Any I/O (file reads, subprocess calls) must be wrapped in a `tea.Cmd` and dispatched to `Init()` or returned from `Update()`. The event loop is single-threaded.
- **Using `tea.Sequentially()`:** Renamed to `tea.Sequence()` in v2.
- **Using `tea.WindowSize()`:** Renamed to `tea.RequestWindowSize` in v2.

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Key binding display | Custom help renderer | `bubbles/v2/help` | Handles truncation, multi-column, disabled keys, ShowAll toggle |
| Scrollable file list | Custom list with pagination | `bubbles/v2/list` | Filtering, pagination, key navigation, delegate pattern |
| Scrollable text area | Manual ANSI string slicing | `bubbles/v2/viewport` | Correct hjkl/g/G/ctrl-d/u built in; horizontal scroll; GotoTop/GotoBottom |
| Key matching | `if msg.String() == "k"` everywhere | `key.Matches(msg, binding)` | Handles multi-key bindings, disabled state, help generation |
| Loading spinner | Manual frame cycling | `bubbles/v2/spinner` | Multiple built-in styles, correct Tick integration |
| Color downsampling | Manual ANSI stripping | `lipgloss.Fprintln(os.Stderr, ...)` | Automatic downsample to ASCII in non-TTY (CI/pipe) |

**Key insight:** The bubbles components handle edge cases (terminal resize, disabled bindings, scroll bounds, filter state) that are trivial to miss in custom code and extremely tedious to test.

---

## Common Pitfalls

### Pitfall 1: v1 API Cargo-Culted into v2

**What goes wrong:** Copy-pasting bubbletea v1 examples. `View()` returns `string`, `tea.KeyMsg` used as concrete type, `WithAltScreen()` in program options, `tea.Sequentially()` instead of `tea.Sequence()`.

**Why it happens:** v1 examples are everywhere; v2 is newer. The Go compiler catches most type errors, but `msg.String()` matching space as `" "` vs `"space"` is a silent runtime bug.

**How to avoid:** Use the upgrade guide as a checklist before any key-handling code. Run tests with space key explicitly.

**Warning signs:** `View()` compile error about string/View mismatch; key handlers never firing; space bar doing nothing.

### Pitfall 2: lipgloss.AdaptiveColor Hang

**What goes wrong:** `lipgloss.AdaptiveColor` causes bubbletea to hang indefinitely when `View()` is called in certain terminal configurations (issue #1036).

**Why it happens:** AdaptiveColor tries to query terminal background from stdin/stdout, which conflicts with bubbletea's I/O control.

**How to avoid:** Use only explicit hex colors: `lipgloss.Color("#89b4fa")`. The UI spec defines the full palette with constant names.

**Warning signs:** TUI starts, then freezes with no output. CPU usage low (blocked on I/O).

### Pitfall 3: Startup Validation Initializes TUI Before Checking

**What goes wrong:** Calling `tea.NewProgram(m).Run()` before checking for sops binary. If sops is missing, the TUI appears (alternate screen), then errors, then exits — leaving the terminal in a broken state.

**Why it happens:** Validation logic placed inside `Init()` rather than before `tea.NewProgram()`.

**How to avoid:** All validation (sops LookPath, age keys.txt stat, .sops.yaml walk) must happen in `main()` before `tea.NewProgram()`. Only proceed to `p.Run()` after validation returns.

**Warning signs:** Terminal remains in alternate screen buffer after crash; `reset` command needed to recover.

### Pitfall 4: Status Bar Height Calculated as Constant

**What goes wrong:** Using `m.height - 1` to reserve the status bar row. When styling adds padding or borders, content overflows or underflows by 1-2 rows.

**Why it happens:** Hard-coded offset doesn't account for lipgloss padding/margin in the status bar style.

**How to avoid:** Always `lipgloss.Height(m.status.View(m.width))` and subtract that dynamically in `View()`.

**Warning signs:** Content partially hidden behind status bar; status bar appearing one row too high.

### Pitfall 5: WindowSizeMsg Not Propagated to Children

**What goes wrong:** Root model updates `m.width`/`m.height` on `tea.WindowSizeMsg` but does not call `SetSize()` on child components (list, viewport). Children render at their initial dimensions and do not respond to terminal resize.

**Why it happens:** The root model catches `WindowSizeMsg` and often does not forward it explicitly to child models.

**How to avoid:** In the `WindowSizeMsg` case of root Update(), call `m.fileList.SetSize(msg.Width, mainH)`, `m.detail.SetSize(msg.Width, mainH)`, `m.status.SetWidth(msg.Width)`.

**Warning signs:** Resizing terminal causes list to render partially clipped or with blank lines.

### Pitfall 6: Flash Timer Leaks After Multiple Flashes

**What goes wrong:** User triggers two flash messages quickly. First `tea.Tick` fires and clears the second flash prematurely.

**Why it happens:** Each `tea.Tick` returns a `flashClearMsg` independently; there is no cancellation of the first tick when the second flash starts.

**How to avoid:** Include a flash generation counter in StatusBarModel. `flashClearMsg` carries the generation it was created for. Only clear if generations match.

```go
type flashClearMsg struct{ gen int }

func (m StatusBarModel) Flash(msg string) (StatusBarModel, tea.Cmd) {
    m.flashGen++
    gen := m.flashGen
    m.flash = msg
    return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
        return flashClearMsg{gen: gen}
    })
}
// In Update: only clear if msg.gen == m.flashGen
```

**Warning signs:** Flash messages disappear too early when multiple operations run in sequence.

---

## Code Examples

Verified patterns from official sources:

### Bubbletea v2 Minimal Program

```go
// Source: pkg.go.dev/charm.land/bubbletea/v2 (official example)
package main

import (
    "fmt"
    "os"
    tea "charm.land/bubbletea/v2"
)

type model struct{ count int }

func (m model) Init() tea.Cmd              { return nil }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if msg, ok := msg.(tea.KeyPressMsg); ok {
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit()
        case "j", "down":
            m.count++
        case "k", "up":
            m.count--
        }
    }
    return m, nil
}
func (m model) View() tea.View {
    v := tea.NewView(fmt.Sprintf("Count: %d\nPress q to quit.", m.count))
    v.AltScreen = true
    return v
}

func main() {
    p := tea.NewProgram(model{})
    if _, err := p.Run(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

### LookPath for sops Binary Check

```go
// Source: pkg.go.dev/os/exec
import (
    "errors"
    "os/exec"
)

func checkSops() error {
    _, err := exec.LookPath("sops")
    if errors.Is(err, exec.ErrNotFound) {
        return fmt.Errorf("sops binary not found in $PATH")
    }
    return err // nil if found, other error if PATH issue
}
```

### Age Key Existence Check

```go
// Source: pkg.go.dev/filippo.io/age — ParseIdentities
import (
    "filippo.io/age"
    "os"
    "path/filepath"
)

func checkAgeKey() (bool, error) {
    home, _ := os.UserHomeDir()
    keyPath := filepath.Join(home, ".config", "sops", "age", "keys.txt")
    f, err := os.Open(keyPath)
    if err != nil {
        return false, nil // soft: file missing is not fatal
    }
    defer f.Close()
    ids, err := age.ParseIdentities(f)
    if err != nil {
        return false, nil
    }
    return len(ids) > 0, nil
}
```

### .sops.yaml Discovery (Walk Up Directories)

```go
// Source: [ASSUMED] — standard "walk up" pattern; no specific library needed
func findSopsYaml() bool {
    dir, _ := os.Getwd()
    for {
        if _, err := os.Stat(filepath.Join(dir, ".sops.yaml")); err == nil {
            return true
        }
        parent := filepath.Dir(dir)
        if parent == dir {
            return false // reached filesystem root
        }
        dir = parent
    }
}
```

### Viewport Setup with vim Keybindings

```go
// Source: pkg.go.dev/charm.land/bubbles/v2/viewport
import "charm.land/bubbles/v2/viewport"

vp := viewport.New(viewport.WithWidth(width), viewport.WithHeight(height))
vp.SetContent(yamlTreeString)

// In Update:
case tea.KeyPressMsg:
    switch msg.String() {
    case "g":
        vp.GotoTop()
    case "G":
        vp.GotoBottom()
    // ctrl+d, ctrl+u, j, k handled by vp.Update(msg) automatically
    }
var cmd tea.Cmd
vp, cmd = vp.Update(msg)
```

### Lipgloss Stderr Error Box

```go
// Source: pkg.go.dev/charm.land/lipgloss/v2 — Fprintln, RoundedBorder
import (
    "charm.land/lipgloss/v2"
    "os"
)

box := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("#f38ba8")). // ColorError
    Padding(1, 2).
    Render("sops binary not found\nInstall: https://github.com/getsops/sops#install")

lipgloss.Fprintln(os.Stderr, box)
os.Exit(1)
```

### Teatest Integration Test

```go
// Source: github.com/charmbracelet/x/exp/teatest (last published Feb 16 2026)
import (
    "testing"
    "time"
    "github.com/charmbracelet/x/exp/teatest"
    "github.com/charmbracelet/lipgloss"
    "github.com/muesli/termenv"
)

func init() {
    lipgloss.SetColorProfile(termenv.Ascii) // strip colors for golden file stability
}

func TestHelpOverlay(t *testing.T) {
    m := app.NewModel()
    tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))
    
    tm.Send(tea.KeyPressMsg{Code: '?'}) // open help
    tm.WaitFor(
        func(bts []byte) bool { return strings.Contains(string(bts), "Navigation") },
        teatest.WithCheckInterval(time.Millisecond*100),
        teatest.WithDuration(time.Second*3),
    )
    teatest.RequireEqualOutput(t, tm.FinalOutput(t))
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `View() string` | `View() tea.View` struct | bubbletea v2.0.0 | Compile error if using v1 signature |
| `tea.KeyMsg` concrete struct | `tea.KeyPressMsg` concrete, `tea.KeyMsg` interface | bubbletea v2.0.0 | Silent bug if type-switching on `tea.KeyMsg` |
| `tea.WithAltScreen()` program option | `view.AltScreen = true` field | bubbletea v2.0.0 | Option silently ignored in v2 |
| `tea.Sequentially()` | `tea.Sequence()` | bubbletea v2.0.0 | Compile error if using old name |
| `tea.WindowSize()` | `tea.RequestWindowSize` | bubbletea v2.0.0 | Renamed to value constant |
| `p.Start()` | `p.Run()` | bubbletea v2.0.0 | Compile error |
| `lipgloss.AdaptiveColor` | Explicit `lipgloss.Color("#hex")` | lipgloss v2 + btea hang fix | AdaptiveColor causes hangs; removed from v2 new code |
| `github.com/charmbracelet/bubbletea` | `charm.land/bubbletea/v2` | v2.0.0 April 2026 | Wrong import path = module not found |

**Deprecated/outdated:**
- `github.com/charmbracelet/bubbletea/v2`: Old pre-release path — use `charm.land/bubbletea/v2`
- `lipgloss.AdaptiveColor`: Banned (hang). Use explicit colors or the `compat` package only for migrations.
- `p.StartReturningModel()`: Removed. `p.Run()` returns `(Model, error)` in v2.

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `.sops.yaml` walk-up uses `filepath.Dir` loop to root; no stdlib helper exists | Code Examples | Low risk — standard pattern; would need to switch to a library |
| A2 | `bubbles/v2/viewport` automatically handles `ctrl+d`/`ctrl+u` when `vp.Update(msg)` is called | Architecture Patterns (Pattern 1, Viewport) | Medium — if not auto-handled, must call `HalfPageDown()`/`HalfPageUp()` explicitly in Update |
| A3 | `charm.land/bubbles/v2/list` `DefaultKeyMap` includes `g`/`G` for GoToStart/GoToEnd | Architecture Patterns | Medium — if not included, must add custom binding |

**Claims A2 and A3:** Viewport and list key defaults were confirmed via pkg.go.dev docs but exact auto-routing behavior (whether `vp.Update(msg)` or `m.list.Update(msg)` handles the keys internally vs. needing explicit branch) should be confirmed against a live Go build. The docs show `GotoTop()` / `GotoBottom()` as methods, suggesting they may need explicit calls.

---

## Open Questions

1. **Does `bubbles/v2/list` include `g`/`G` in its default KeyMap?**
   - What we know: list.DefaultKeyMap includes GoToStart/GoToEnd bindings
   - What's unclear: Whether they are bound to `g`/`G` by default or require explicit configuration
   - Recommendation: Write a small probe program in Wave 0; if not default, add to custom KeyMap

2. **Does viewport's `Update(msg)` automatically consume `ctrl+d`/`ctrl+u`, or must Update explicitly call `HalfPageDown()`/`HalfPageUp()`?**
   - What we know: Viewport has `HalfPageDown()`, `HalfPageUp()`, `GotoTop()`, `GotoBottom()` methods; docs show default KeyMap includes ctrl+d/u
   - What's unclear: Whether `vp.Update(msg)` internally dispatches or requires external key matching
   - Recommendation: Pass all messages to `vp.Update(msg)` and verify in Wave 0 test; add explicit branches only if needed

3. **Terminal width for stderr error box before TUI starts**
   - What we know: `lipgloss.Width()` measures rendered strings; terminal width must come from `os.Stdout` or environment
   - What's unclear: Best way to query terminal width without bubbletea running (before `tea.WindowSizeMsg`)
   - Recommendation: Use `golang.org/x/term` `GetSize(int(os.Stdout.Fd()))` or `COLUMNS` env var with fallback to 80

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| sops binary | HLT-01 (tested), runtime ops | ✓ | 3.12.2 | — (hard error on missing) |
| age binary | HLT-02 (indirectly) | ✓ | 1.3.1 | — (soft warning on missing) |
| Go toolchain | Build | ✓ | 1.26.2 | — |
| charm.land modules | TUI | not yet fetched | v2.0.4 target | — (go get required) |

[VERIFIED: bash — sops 3.12.2, age 1.3.1, go 1.26.2-X:nodwarf5 confirmed on dev machine]

**No blocking missing dependencies.** `go get` commands will fetch charm.land modules from the network on first build.

---

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | stdlib `testing` + `stretchr/testify` v1.x |
| Config file | none — Go standard `go test ./...` |
| Quick run command | `go test ./internal/... -run TestUnit -v` |
| Full suite command | `go test ./... -timeout 30s` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| HLT-01 | sops missing → styled error box written to stderr, exit 1 | unit | `go test ./internal/validator/... -run TestSopsMissing -v` | Wave 0 |
| HLT-02 | age key missing → warning rendered, TUI still launches | unit | `go test ./internal/validator/... -run TestAgeKeyMissing -v` | Wave 0 |
| NAV-03 | hjkl/g/G/ctrl-d/ctrl-u navigate without errors | integration | `go test ./internal/ui/... -run TestNavigation -v` | Wave 0 |
| NAV-05 | `?` opens help overlay showing correct keybindings | integration (teatest) | `go test ./internal/app/... -run TestHelpOverlay -v` | Wave 0 |
| NAV-06 | Status bar visible on every screen, flash clears after 2s | integration (teatest) | `go test ./internal/ui/... -run TestStatusBar -v` | Wave 0 |

### Sampling Rate

- **Per task commit:** `go test ./internal/validator/... -v`
- **Per wave merge:** `go test ./... -timeout 30s`
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps

- [ ] `internal/validator/startup_test.go` — covers HLT-01, HLT-02
- [ ] `internal/ui/statusbar_test.go` — covers NAV-06
- [ ] `internal/app/app_test.go` — covers NAV-03, NAV-05 via teatest
- [ ] `internal/app/testdata/*.golden` — golden output files for teatest
- [ ] `lipgloss.SetColorProfile(termenv.Ascii)` in `TestMain` or package `init()` in each test file

---

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | — |
| V3 Session Management | no | — |
| V4 Access Control | no | — |
| V5 Input Validation | yes (key input) | tea.KeyPressMsg type switch — no raw string exec |
| V6 Cryptography | no (Phase 1 touches zero secret values) | — |

### Known Threat Patterns for this Phase

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Key path traversal (malicious CWD) | Tampering | `filepath.Dir` loop stops at root; `os.Stat` — no exec of any path |
| Age keys.txt readable by others | Info Disclosure | Detection only — display key availability, never log key content |
| sops binary replaced with malicious binary in PATH | Elevation | Not mitigated in Phase 1 — Phase 1 only runs `exec.LookPath`, no subprocess exec |
| Terminal escape injection via filenames | Tampering | lipgloss renders through ANSI encoding — no raw escape passthrough; but note for Phase 2 when filenames are displayed |

**Phase 1 security posture:** The phase makes zero subprocess calls to sops (only `LookPath` checks existence) and reads no secret values. The attack surface is minimal. The `AltScreen = true` pattern ensures clean terminal cleanup on exit.

---

## Sources

### Primary (HIGH confidence)

- `pkg.go.dev/charm.land/bubbletea/v2` — Model interface, tea.View struct, tea.KeyPressMsg, KeyMod, Program.Run(), tea.Tick, tea.Sequence, tea.RequestWindowSize [VERIFIED: WebFetch]
- `github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md` — Full v1→v2 breaking changes list [VERIFIED: WebFetch]
- `pkg.go.dev/charm.land/lipgloss/v2` — NewStyle, color types, border styles, Fprintln/Fprint, JoinHorizontal/JoinVertical [VERIFIED: WebFetch]
- `pkg.go.dev/charm.land/bubbles/v2/list` — Item interface, Model, KeyMap, DefaultDelegate [VERIFIED: WebFetch]
- `pkg.go.dev/charm.land/bubbles/v2/viewport` — New(), SetContent, GotoTop/Bottom, HalfPageDown/Up, KeyMap [VERIFIED: WebFetch]
- `pkg.go.dev/charm.land/bubbles/v2/help` — KeyMap interface, Model, View(), ShortHelpView, FullHelpView [VERIFIED: WebFetch]
- `pkg.go.dev/filippo.io/age` — ParseIdentities, X25519Identity, Recipient().String() [VERIFIED: WebFetch]
- `pkg.go.dev/os/exec` — LookPath, ErrNotFound, ErrDot [VERIFIED: WebFetch]
- `github.com/charmbracelet/x/exp/teatest` — WaitFor, RequireEqualOutput, NewTestModel (Feb 2026) [VERIFIED: WebSearch + charm.land/blog/teatest]
- `.planning/phases/01-foundation/01-UI-SPEC.md` — Color palette, spacing, keybindings, layout spec [VERIFIED: Read]
- `.planning/phases/01-foundation/01-CONTEXT.md` — All locked decisions D-01 through D-12 [VERIFIED: Read]

### Secondary (MEDIUM confidence)

- `github.com/charmbracelet/lipgloss/discussions/506` — lipgloss v2 AdaptiveColor removal, LightDark replacement [VERIFIED: WebFetch]
- `charm.land/blog/teatest/` — teatest WaitFor, golden files, lipgloss.SetColorProfile(termenv.Ascii) pattern [VERIFIED: WebFetch]
- `leg100.github.io/en/posts/building-bubbletea-programs/` — lipgloss.Height() for layout calculation, message ordering in tea.Batch [VERIFIED: WebFetch]
- `github.com/charmbracelet/bubbletea/discussions/1374` — Bubble Tea v2 overview, AltScreen as View field [VERIFIED: WebSearch cross-reference]

### Tertiary (LOW confidence)

- WebSearch results for overlay patterns (bubbletea-overlay, composable-views) — used only to confirm the sessionState enum pattern is idiomatic, not as code source

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all packages verified against pkg.go.dev with publish dates
- Architecture: HIGH — patterns verified against official examples and upgrade guide
- Pitfalls: HIGH — v2 API breaks documented in official upgrade guide; AdaptiveColor bug confirmed in issue tracker
- Testing: MEDIUM — teatest path verified (github.com/charmbracelet/x/exp/teatest), but exact v2 compatibility not confirmed against a live build

**Research date:** 2026-04-14
**Valid until:** 2026-05-14 (charm ecosystem moves fast; re-verify bubbletea/bubbles versions if planning is delayed beyond 30 days)
