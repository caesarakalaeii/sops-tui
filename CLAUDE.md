<!-- GSD:project-start source:PROJECT.md -->
## Project

**sops-tui**

A k9s-inspired terminal UI for managing SOPS-encrypted secrets. Works with any repository that has a `.sops.yaml` configuration. Browse, decrypt, edit, rotate, and audit secrets — all from a keyboard-driven TUI built with Go and Bubble Tea.

**Core Value:** Developers can manage all their SOPS-encrypted secrets from a single terminal interface without remembering CLI flags or writing shell scripts.

### Constraints

- **Stack**: Go + Bubble Tea (Charm ecosystem) — chosen for single-binary distribution and k9s ecosystem alignment
- **SOPS integration**: Subprocess calls to `sops` CLI — must handle `sops` not being installed gracefully
- **Encryption**: age keys via `~/.config/sops/age/keys.txt` — standard SOPS convention
- **License**: AGPL-3.0
- **Dependencies**: Requires `sops` binary installed and age key available for decryption operations
<!-- GSD:project-end -->

<!-- GSD:stack-start source:research/STACK.md -->
## Technology Stack

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
- `View()` now returns `tea.View` (struct), not `string`
- `tea.KeyMsg` is now an interface; use `tea.KeyPressMsg` for key presses
- `msg.Type` → `msg.Code` (rune); `msg.Runes` → `msg.Text` (string); `msg.Alt` → `msg.Mod`
- Space: `" "` → `"space"` in `msg.String()`
- Program options (`WithAltScreen()`) replaced by View struct fields (`view.AltScreen = true`)
### SOPS Integration
| Technology | Approach | Why |
|------------|----------|-----|
| os/exec + context | `exec.CommandContext(ctx, "sops", ...)` | Subprocess to the user's installed `sops` binary. Project scope explicitly rules out using sops as a Go library. This avoids version skew issues and ensures the TUI always behaves identically to the CLI. SOPS v3.12.2 is current stable. |
- Always use `exec.CommandContext` with a `context.WithTimeout` for decrypt/encrypt calls.
- Capture `stderr` separately for error surfacing to the user.
- Check for `sops` on `$PATH` at startup; surface a clear error if absent.
- For edit flows: pass content via stdin with `sops --input-type yaml --output-type yaml /dev/stdin` OR write to a temp file.
- Never buffer stdout/stderr with `Output()` for large files; use `Pipe()` + goroutine reads to avoid deadlock.
### Encryption / Key Parsing
| Technology | Version | Import Path | Why |
|------------|---------|-------------|-----|
| filippo.io/age | v1.3.1 | `filippo.io/age` | Official age Go library by the age spec author (Filippo Valsorda). v1.3.1 (Dec 2025) adds post-quantum hybrid keys. Use for parsing `~/.config/sops/age/keys.txt` to display key fingerprints, validate key availability before decrypt attempts, and show recipient public keys. Do NOT use for encryption — that stays with the `sops` subprocess. |
### YAML Parsing
| Technology | Version | Import Path | Why |
|------------|---------|-------------|-----|
| goccy/go-yaml | v1.19.2 | `github.com/goccy/go-yaml` | Preferred over `gopkg.in/yaml.v3` for three reasons: (1) passes ~60 more YAML test suite cases than yaml.v3, (2) the yaml.v3 author recommends migrating to goccy/go-yaml, (3) better marshal/unmarshal customization needed for handling SOPS encrypted value format. Published Jan 2026. |
### JSON Parsing
### Fuzzy Search
| Technology | Version | Import Path | Why |
|------------|---------|-------------|-----|
| sahilm/fuzzy | v0.1.1 | `github.com/sahilm/fuzzy` | Zero external dependencies, optimized for filename/symbol matching (exactly our use case: fuzzy-matching across secret key names). Returns matched character positions enabling highlight rendering with lipgloss. Sublime Text-style ranking. Last updated May 2023, but the algorithm is stable and the API won't change. |
### Clipboard
| Technology | Version | Import Path | Why |
|------------|---------|-------------|-----|
| atotto/clipboard | v0.1.4 | `github.com/atotto/clipboard` | Simple, dependency-free clipboard access via xclip/xsel on Linux. No CGo required. Last release 2021, but the API is complete and stable — clipboard access on Linux fundamentally hasn't changed. |
### Git Integration
| Technology | Version | Import Path | Why |
|------------|---------|-------------|-----|
| go-git/go-git | v5.17.2 | `github.com/go-git/go-git/v5` | Pure Go git implementation — no `git` binary dependency. v5.17.2 (March 2026) with security fixes. Use for: `git log --follow <file>` (history per secret file), `git blame` equivalent to show last-commit-per-line, detecting uncommitted changes in secret files. |
### Build and Release
| Technology | Version | Why |
|------------|---------|-----|
| goreleaser | v2.15.2 | Standard for Go binary releases. v2.15.2 (March 2026) is current. `goreleaser init` generates correct defaults for linux/darwin/amd64/arm64 matrices. Use for GitHub releases with checksums, SBOM generation, and Homebrew tap. |
### Testing
| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| stdlib testing | Go 1.26 | Unit + integration tests | Sufficient for pure function testing (YAML parsing, fuzzy search, key fingerprinting). No external dep. |
| stretchr/testify | v1.x | Assertions | `require` for precondition failures, `assert` for value checks. De-facto standard. Do not use the go-openapi v2 fork — it is unrelated and unmaintained by the original team. |
| charmbracelet/x/exp/teatest | latest | TUI golden file tests | Official Charm test harness for bubbletea programs. `WaitFor` polls the rendered output until a condition matches. `RequireEqualOutput` compares against golden files. Force `lipgloss.NoColor` profile in tests to avoid CI color-profile divergence. |
| sebdah/goldie | v2.x | Golden file management | Manages `.golden` test fixture files with `-update` flag support. Use alongside teatest for snapshot diffs. |
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
# Core framework
# Encryption / parsing
# UX utilities
# Git integration
# Testing
# go.mod minimum Go version
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
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

Conventions not yet established. Will populate as patterns emerge during development.
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

Architecture not yet mapped. Follow existing patterns found in the codebase.
<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->
## Project Skills

No project skills found. Add skills to any of: `.claude/skills/`, `.agents/skills/`, `.cursor/skills/`, or `.github/skills/` with a `SKILL.md` index file.
<!-- GSD:skills-end -->

<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
