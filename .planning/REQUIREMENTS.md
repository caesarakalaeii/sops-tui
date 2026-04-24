# Requirements: sops-tui

**Defined:** 2026-04-13
**Core Value:** Developers can manage all their SOPS-encrypted secrets from a single terminal interface without remembering CLI flags or writing shell scripts.

## v1 Requirements

### Navigation & Discovery

- [x] **NAV-01**: User can browse all SOPS-encrypted files discovered via `.sops.yaml` path rules
- [x] **NAV-02**: User can view key names from encrypted files without decrypting
- [ ] **NAV-03**: User can navigate with vim keybindings (hjkl, g/G, ctrl-d/u)
- [x] **NAV-04**: User can fuzzy search files and keys with `/` (k9s-style)
- [ ] **NAV-05**: User can view contextual help panel with `?`
- [ ] **NAV-06**: User sees persistent status bar (file path, encryption status, operation feedback)

### Decrypt & View

- [ ] **DEC-01**: User can decrypt and reveal individual secret values on demand
- [ ] **DEC-02**: User can decrypt and reveal all values in a file
- [x] **DEC-03**: Secret values are masked by default, revealed on keypress
- [x] **DEC-04**: User can view SOPS metadata (version, lastmodified, recipients, MAC status) without decrypting

### Edit & Rotate

- [ ] **EDT-01**: User can edit a secret value with automatic re-encryption
- [ ] **EDT-02**: User sees diff view before confirming re-encryption
- [ ] **EDT-03**: User can rotate a secret to a format-aware random value (base64, hex, UUID, bcrypt)
- [ ] **EDT-04**: User must confirm before any destructive write operation

### Clipboard

- [ ] **CLB-01**: User can copy decrypted value to clipboard
- [ ] **CLB-02**: Clipboard auto-clears after configurable timeout (default 30s)
- [ ] **CLB-03**: Clipboard clears on process exit (including SIGINT/SIGTERM)

### Recipients

- [ ] **RCP-01**: User can view age key recipients per file
- [ ] **RCP-02**: User can add/remove age key recipients
- [ ] **RCP-03**: User can bulk re-key multiple files

### Git Integration

- [ ] **GIT-01**: User sees uncommitted change badges on files ([M], [A], [?])
- [ ] **GIT-02**: User can view git blame/history per secret file
- [ ] **GIT-03**: User can fuzzy search across all files and key names

### Health & Audit

- [ ] **HLT-01**: User sees startup error with instructions if `sops` binary is missing
- [ ] **HLT-02**: User sees startup error with instructions if age key file is missing
- [ ] **HLT-03**: User can run secret health checks (weak secrets, duplicates, staleness)

## v1.1 Requirements — k9s Visual Parity

Goal: Reshape the UI so it looks and behaves like k9s — persistent keybinding menu, ASCII logo, header info panel, titled bordered content, colored breadcrumb chips, k9s-tuned palette. v1.0 functional loops must not regress.

### Header Region

- [ ] **UI-01**: User sees a persistent multi-column keybinding menu in the header on every view — no `?` press required to discover hotkeys
- [ ] **UI-02**: User sees a 6-row ASCII logo anchored to the top-right of the header, ~26 columns wide
- [ ] **UI-03**: Logo recolors to reflect aggregate app status (info / warn / error) derived from env checks, flash severity, and health aggregate
- [ ] **UI-04**: User sees a header info panel (top-left) with five rows: `.sops.yaml` relative path, age key fingerprint, recipient count, git branch + clean/dirty marker, file count
- [ ] **UI-05**: Info-panel fields are truncated and de-PII'd before render: age fingerprint ≤10 chars with ellipsis, paths are repo-relative, no copy bindings ever target chrome content

### Content Framing

- [ ] **UI-06**: Every primary view (Files, Detail, Metadata, Diff, Help, History, Health, Recipients, RecipientForm) is wrapped in a titled bordered region; title encodes the view name and when relevant an item count
- [ ] **UI-07**: Breadcrumb segments render as colored chip pills replacing the legacy ` > ` text separator; the active (last) segment uses the accent color
- [ ] **UI-08**: Bottom status bar shrinks to only right-aligned env indicators + clipboard state; the breadcrumb moves to a dedicated crumb row above the titled body

### Keybinding Discoverability

- [ ] **UI-09**: Every interactive sub-model exposes a `Hints() []keys.MenuHint` method derived from its existing `key.Binding.ShortHelp()` definitions — single source of truth is the keymap
- [ ] **UI-10**: The persistent menu re-hydrates from the active sub-model's `Hints()` on every `View()` call; modal states (diff, recipient confirm, bulk re-key) show their modal keybindings, not the underlying file-list ones
- [ ] **UI-11**: The `?` full-screen help overlay is retained as the complete reference; day-to-day hotkeys are discoverable without opening it

### Theming & Accessibility

- [ ] **UI-12**: Default palette is tuned to k9s conventions (accent shifts toward hot-pink/purple typical of k9s skins) while keeping the AdaptiveColor ban from v1.0
- [ ] **UI-13**: On 16-color terminals (`TERM=xterm` / Ascii profile) a safe fallback palette is applied so paired bg/fg chips and menu cells remain legible
- [ ] **UI-14**: Every color-coded state (info / warn / error, active vs inactive chip, env indicators, flash severity) uses redundant shape or text encoding (prefix like `[I]` / `[W]` / `[E]`, inverted bg+fg for active, underline for focus) so the UI remains usable for colorblind users
- [ ] **UI-15**: Persistent chrome content is ASCII-only; `lipgloss.NormalBorder()` is the only border style used in chrome (grep-gated in CI to prevent regressions to fancy borders or emoji)
- [ ] **UI-16**: The app survives rendering at 40×12 through 200×60 without layout corruption; narrow-terminal rendering may be ugly but must not truncate critical data or overflow the viewport

### Layout Safety (groundwork)

- [ ] **UI-17**: A `bodyDims(m) (w, h int)` helper is the single source of truth for body size arithmetic, subtracting chrome + crumbs + status-bar heights; all existing `m.height - statusBarHeight(m)` call-sites migrate to it before any chrome renders
- [x] **UI-18**: A CI grep-gate prevents reintroduction of the raw `m.height - statusBarHeight(m)` pattern outside the helper
- [x] **UI-19**: A teatest harness helper strips ANSI escape sequences for structural golden comparison and asserts color presence separately so goldens stay stable across lipgloss bumps

### Regression & Performance

- [ ] **UI-20**: All v1.0 functional flows (file discovery, reveal, edit, diff, rotate, clipboard, git, recipient management, health) keep passing their integration tests after the chrome lands; no v1.0 feature regresses
- [ ] **UI-21**: `BenchmarkAppView` stays ≤50 µs/op at 200×60 with chrome rendered; no `lipgloss.NewStyle()` calls appear inside `View()` (styles declared as package vars)

## v2 Requirements

### Encryption Backend Expansion

- **ENC-01**: Support GPG/PGP encrypted files alongside age
- **ENC-02**: Support cloud KMS (AWS KMS, GCP KMS, Azure Key Vault) as key sources
- **ENC-03**: Support HashiCorp Vault transit engine as SOPS backend

### Theming (deferred from v1.1)

- **THM-01**: Load user skin from `~/.config/sops-tui/skin.yaml` using a k9s-compatible YAML schema subset (so an existing k9s dracula/gruvbox skin drops in); fail-open with flash warning on invalid hex values
- **THM-02**: Ship 2–3 builtin skins (dracula, gruvbox-dark, monokai) embedded in the binary via `embed.FS`
- **THM-03**: Live skin reload on file change via fsnotify

### Advanced Editing

- **ADV-01**: Full-file editing via `$EDITOR` subprocess
- **ADV-02**: Secret template/policy definitions (enforce formats per key)
- **ADV-03**: Batch edit operations across multiple files

### Observability

- **OBS-01**: Audit log of all decrypt/edit/rotate operations within a session
- **OBS-02**: Export health check results to JSON/CSV

## Out of Scope

| Feature | Reason |
|---------|--------|
| Web UI / REST API | Security surface; tool's value is terminal-only with no network exposure |
| Kubernetes Secret sync | Different product category; ArgoCD/Flux handle this |
| Built-in text editor | Complex, duplicates `sops <file>` which opens `$EDITOR` |
| Live watch / auto-reload | Causes confusing in-memory vs on-disk state splits |
| Multi-user RBAC | SOPS has no RBAC; access control is via key distribution |
| GPG key management UI | gpg-tui (1.7k stars) already exists for this |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| HLT-01 | Phase 1 | Pending |
| HLT-02 | Phase 1 | Pending |
| NAV-03 | Phase 1 | Pending |
| NAV-05 | Phase 1 | Pending |
| NAV-06 | Phase 1 | Pending |
| NAV-01 | Phase 2 | Complete |
| NAV-02 | Phase 2 | Complete |
| NAV-04 | Phase 2 | Complete |
| DEC-03 | Phase 2 | Complete |
| DEC-04 | Phase 2 | Complete |
| DEC-01 | Phase 3 | Pending |
| DEC-02 | Phase 3 | Pending |
| EDT-01 | Phase 3 | Pending |
| EDT-02 | Phase 3 | Pending |
| EDT-03 | Phase 3 | Pending |
| EDT-04 | Phase 3 | Pending |
| CLB-01 | Phase 4 | Pending |
| CLB-02 | Phase 4 | Pending |
| CLB-03 | Phase 4 | Pending |
| GIT-01 | Phase 4 | Pending |
| GIT-02 | Phase 4 | Pending |
| GIT-03 | Phase 4 | Pending |
| RCP-01 | Phase 5 | Pending |
| RCP-02 | Phase 5 | Pending |
| RCP-03 | Phase 5 | Pending |
| HLT-03 | Phase 5 | Pending |
| UI-17 | Phase 6 | Pending |
| UI-18 | Phase 6 | Complete (Plan 01) |
| UI-19 | Phase 6 | Complete (Plan 01) |
| UI-01 | Phase 7 | Pending |
| UI-02 | Phase 7 | Pending |
| UI-06 | Phase 7 | Pending |
| UI-15 | Phase 7 | Pending |
| UI-04 | Phase 8 | Pending |
| UI-05 | Phase 8 | Pending |
| UI-07 | Phase 8 | Pending |
| UI-08 | Phase 8 | Pending |
| UI-09 | Phase 9 | Pending |
| UI-10 | Phase 9 | Pending |
| UI-11 | Phase 9 | Pending |
| UI-03 | Phase 10 | Pending |
| UI-12 | Phase 10 | Pending |
| UI-13 | Phase 10 | Pending |
| UI-14 | Phase 10 | Pending |
| UI-16 | Phase 10 | Pending |
| UI-20 | Phase 11 | Pending |
| UI-21 | Phase 11 | Pending |

**Coverage:**
- v1 requirements: 26 total, mapped: 26, unmapped: 0 ✓
- v1.1 requirements: 21 total (UI-01 through UI-21), mapped: 21, unmapped: 0 ✓

---
*Requirements defined: 2026-04-13*
*Last updated: 2026-04-23 — v1.1 traceability table populated; UI-01..UI-21 mapped across Phases 6-11*
