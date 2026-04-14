# Technology Stack

**Project:** sops-tui
**Researched:** 2026-04-13

## Recommended Stack

### Runtime / Language

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go | 1.26.x | Language | Current stable release (1.26.2 as of 2026-04-07). Required by charm.land v2 libraries. Single-binary distribution aligns with project constraints. |

### Core TUI Framework

| Technology | Version | Import Path | Why |
|------------|---------|-------------|-----|
| bubbletea | v2.0.4 | `charm.land/bubbletea/v2` | The only serious option for Elm-architecture TUI in Go. v2 is stable, production-ready, and includes the new Cursed Renderer (ncurses-based, ~30% faster rendering, Mode 2026 synchronized output). Confirmed production-ready in 2026. |
| lipgloss | v2.x | `charm.land/lipgloss/v2` | Required by bubbletea v2. Provides all layout, color, and border styling. v2 aligns color downsampling with explicit terminal capability detection — important for CI/SSH scenarios. |
| bubbles | v2.x | `charm.land/bubbles/v2` | Pre-built TUI components: list (file browser pane), textinput (inline editing), viewport (scrollable diff), spinner (decrypt progress), help (keybinding footer). Avoids re-implementing standard widgets. |
| huh | v2.x | `charm.land/huh/v2` | Form-based input for multi-field flows (add recipient, generate secret value). Built on bubbletea v2. Published Feb 2026. Use for modal dialogs requiring multiple fields. |

**Critical import path note:** All Charm v2 libraries moved from `github.com/charmbracelet/*` to `charm.land/*/v2`. The old `github.com/charmbracelet/bubbletea` module remains at v1-compatible releases but will not receive v2 features. New projects must use the `charm.land` vanity domain.

**v2 breaking change summary (relevant to this project):**
- `View()` now returns `tea.View` (struct), not `string`
- `tea.KeyMsg` is now an interface; use `tea.KeyPressMsg` for key presses
- `msg.Type` → `msg.Code` (rune); `msg.Runes` → `msg.Text` (string); `msg.Alt` → `msg.Mod`
- Space: `" "` → `"space"` in `msg.String()`
- Program options (`WithAltScreen()`) replaced by View struct fields (`view.AltScreen = true`)

### SOPS Integration

| Technology | Approach | Why |
|------------|----------|-----|
| os/exec + context | `exec.CommandContext(ctx, "sops", ...)` | Subprocess to the user's installed `sops` binary. Project scope explicitly rules out using sops as a Go library. This avoids version skew issues and ensures the TUI always behaves identically to the CLI. SOPS v3.12.2 is current stable. |

Subprocess patterns to adopt:
- Always use `exec.CommandContext` with a `context.WithTimeout` for decrypt/encrypt calls.
- Capture `stderr` separately for error surfacing to the user.
- Check for `sops` on `$PATH` at startup; surface a clear error if absent.
- For edit flows: pass content via stdin with `sops --input-type yaml --output-type yaml /dev/stdin` OR write to a temp file.
- Never buffer stdout/stderr with `Output()` for large files; use `Pipe()` + goroutine reads to avoid deadlock.

### Encryption / Key Parsing

| Technology | Version | Import Path | Why |
|------------|---------|-------------|-----|
| filippo.io/age | v1.3.1 | `filippo.io/age` | Official age Go library by the age spec author (Filippo Valsorda). v1.3.1 (Dec 2025) adds post-quantum hybrid keys. Use for parsing `~/.config/sops/age/keys.txt` to display key fingerprints, validate key availability before decrypt attempts, and show recipient public keys. Do NOT use for encryption — that stays with the `sops` subprocess. |

The age library is the correct tool for parsing key files to build the recipient management UI. Use `age.ParseX25519Identity` to read private keys and derive public key fingerprints for display.

### YAML Parsing

| Technology | Version | Import Path | Why |
|------------|---------|-------------|-----|
| goccy/go-yaml | v1.19.2 | `github.com/goccy/go-yaml` | Preferred over `gopkg.in/yaml.v3` for three reasons: (1) passes ~60 more YAML test suite cases than yaml.v3, (2) the yaml.v3 author recommends migrating to goccy/go-yaml, (3) better marshal/unmarshal customization needed for handling SOPS encrypted value format. Published Jan 2026. |

**Why not gopkg.in/yaml.v3:** As of April 2025, the yaml.v3 upstream author himself recommends goccy/go-yaml as a healthier alternative. yaml.v3 is effectively in maintenance-only mode.

Use goccy/go-yaml for: parsing `.sops.yaml` config files, parsing YAML secret files to enumerate keys, and round-trip parsing of SOPS-encrypted YAML envelopes to display key/value structure.

### JSON Parsing

Use the Go standard library `encoding/json`. The project does not have performance-critical JSON parsing paths (SOPS JSON files are small). goccy/go-json has reported runtime panics in 2025 benchmarks — not worth the risk for marginal gains.

### Fuzzy Search

| Technology | Version | Import Path | Why |
|------------|---------|-------------|-----|
| sahilm/fuzzy | v0.1.1 | `github.com/sahilm/fuzzy` | Zero external dependencies, optimized for filename/symbol matching (exactly our use case: fuzzy-matching across secret key names). Returns matched character positions enabling highlight rendering with lipgloss. Sublime Text-style ranking. Last updated May 2023, but the algorithm is stable and the API won't change. |

**Integration pattern:** Maintain an in-memory index of `(file, key)` tuples. On `/` keypress, run `fuzzy.Find(query, entries)` in the Update loop and re-render the filtered list. Positions from the result drive lipgloss styled character highlighting.

### Clipboard

| Technology | Version | Import Path | Why |
|------------|---------|-------------|-----|
| atotto/clipboard | v0.1.4 | `github.com/atotto/clipboard` | Simple, dependency-free clipboard access via xclip/xsel on Linux. No CGo required. Last release 2021, but the API is complete and stable — clipboard access on Linux fundamentally hasn't changed. |

**Wayland caveat:** atotto/clipboard calls `xclip` or `xsel`, which work under XWayland but not native Wayland sessions without `wl-clipboard` installed. This is an acceptable limitation for v1: document it in the README, check at startup, and surface a warning if neither `xclip`/`xsel`/`wl-copy` is available.

**Alternative considered:** `golang.design/x/clipboard` (v0.7.1, June 2025) — rejected because it requires CGo on Linux and `libx11-dev`, adding a build complexity and system dependency for the same result.

**Auto-clear pattern:** Use `time.AfterFunc(timeout, func() { clipboard.WriteAll("") })` immediately after writing a secret to clipboard.

### Git Integration

| Technology | Version | Import Path | Why |
|------------|---------|-------------|-----|
| go-git/go-git | v5.17.2 | `github.com/go-git/go-git/v5` | Pure Go git implementation — no `git` binary dependency. v5.17.2 (March 2026) with security fixes. Use for: `git log --follow <file>` (history per secret file), `git blame` equivalent to show last-commit-per-line, detecting uncommitted changes in secret files. |

**Why not v6:** v6.0.0-alpha.1 exists (April 2025) but is alpha. v5 is actively maintained with security patches through early 2026. Migrate at v6 stable.

**Why not subprocess git:** The TUI needs structured data (commit hashes, authors, timestamps) not formatted text. go-git returns typed structs making it easier to build the blame/history UI.

### Build and Release

| Technology | Version | Why |
|------------|---------|-----|
| goreleaser | v2.15.2 | Standard for Go binary releases. v2.15.2 (March 2026) is current. `goreleaser init` generates correct defaults for linux/darwin/amd64/arm64 matrices. Use for GitHub releases with checksums, SBOM generation, and Homebrew tap. |

Minimal `.goreleaser.yaml` targets: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`. CGo must be disabled (`CGO_ENABLED=0`) for the clipboard approach to work cleanly in CI.

### Testing

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| stdlib testing | Go 1.26 | Unit + integration tests | Sufficient for pure function testing (YAML parsing, fuzzy search, key fingerprinting). No external dep. |
| stretchr/testify | v1.x | Assertions | `require` for precondition failures, `assert` for value checks. De-facto standard. Do not use the go-openapi v2 fork — it is unrelated and unmaintained by the original team. |
| charmbracelet/x/exp/teatest | latest | TUI golden file tests | Official Charm test harness for bubbletea programs. `WaitFor` polls the rendered output until a condition matches. `RequireEqualOutput` compares against golden files. Force `lipgloss.NoColor` profile in tests to avoid CI color-profile divergence. |
| sebdah/goldie | v2.x | Golden file management | Manages `.golden` test fixture files with `-update` flag support. Use alongside teatest for snapshot diffs. |

**Testing strategy for this project:**
1. Unit tests: All business logic (YAML parsing, key parsing, fuzzy match scoring) — standard `testing` + `testify/require`.
2. TUI snapshot tests: Each major view (file list, secret detail, diff pane) — `teatest.WaitFor` + `RequireEqualOutput` golden files.
3. Integration tests: `sops` subprocess calls — use a test `sops` binary wrapper that validates inputs and returns fixture data. Do not require a real age key in CI.

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| TUI framework | charm.land/bubbletea/v2 | tview + tcell | tview is what k9s uses but is a retained-mode widget system, not Elm architecture. Harder to test, less composable. bubbletea's TEA model enables deterministic snapshot testing. |
| TUI framework | charm.land/bubbletea/v2 | termdash | Less active ecosystem, smaller community, fewer pre-built components. |
| YAML | goccy/go-yaml | gopkg.in/yaml.v3 | yaml.v3 author recommends goccy/go-yaml; fewer YAML spec compliance passes; effectively unmaintained. |
| Clipboard | atotto/clipboard | golang.design/x/clipboard | CGo dependency, libx11-dev system requirement, more complex build pipeline. |
| Git | go-git v5 | os/exec git subprocess | Subprocess returns unstructured text; go-git returns typed structs needed for blame/history UI. |
| JSON | encoding/json | goccy/go-json | Runtime panics reported in 2025 benchmarks. Not needed — SOPS JSON files are small. |
| Forms | charm.land/huh/v2 | custom bubbletea forms | huh handles focus, validation, accessibility. No reason to rebuild. |
| Release | goreleaser v2 | ko, ko-build | ko is container-focused. This is a binary tool. goreleaser produces checksummed archives for homebrew tap. |

## Installation

```bash
# Core framework
go get charm.land/bubbletea/v2
go get charm.land/lipgloss/v2
go get charm.land/bubbles/v2
go get charm.land/huh/v2

# Encryption / parsing
go get filippo.io/age
go get github.com/goccy/go-yaml

# UX utilities
go get github.com/sahilm/fuzzy
go get github.com/atotto/clipboard

# Git integration
go get github.com/go-git/go-git/v5

# Testing
go get github.com/stretchr/testify
go get github.com/charmbracelet/x/exp/teatest@latest
go get github.com/sebdah/goldie/v2
```

```bash
# go.mod minimum Go version
go mod edit -go=1.26
```

## go.mod Notes

- `CGO_ENABLED=0` for all release builds (atotto/clipboard uses xclip/xsel subprocess, no CGo needed)
- Set `GOFLAGS=-trimpath` for reproducible builds
- Pin `filippo.io/age` and `go-git` explicitly — both have security-sensitive histories

## Sources

- Bubbletea v2 stable: https://github.com/charmbracelet/bubbletea/releases (v2.0.4, April 13, 2026 confirmed)
- Bubbletea v2 migration guide: https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md
- charm.land import path: https://pkg.go.dev/charm.land/bubbletea/v2
- huh v2: https://pkg.go.dev/github.com/charmbracelet/huh/v2 (published Oct 2025, updated Feb 2026)
- filippo.io/age v1.3.1: https://pkg.go.dev/filippo.io/age (Dec 2025)
- goccy/go-yaml v1.19.2: https://pkg.go.dev/github.com/goccy/go-yaml (Jan 2026)
- goccy/go-yaml preferred over yaml.v3: https://github.com/cli/cli/issues/10784
- go-git v5.17.2: https://github.com/go-git/go-git/releases (March 2026)
- sahilm/fuzzy v0.1.1: https://pkg.go.dev/github.com/sahilm/fuzzy
- atotto/clipboard v0.1.4: https://github.com/atotto/clipboard
- golang.design/x/clipboard (rejected): https://pkg.go.dev/golang.design/x/clipboard
- teatest: https://charm.land/blog/teatest/
- goreleaser v2.15.2: https://goreleaser.com/blog/goreleaser-v2.15/ (March 2026)
- SOPS v3.12.2: https://github.com/getsops/sops/releases (March 2026)
- Go 1.26.2 current: https://go.dev/doc/devel/release (April 2026)
