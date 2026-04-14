# Research Summary

**Project:** sops-tui
**Synthesized:** 2026-04-13
**Confidence:** HIGH

## Executive Summary

sops-tui is a TUI wrapper for Mozilla SOPS that surfaces read, write, and key-management operations currently accessible only via CLI flags. The established approach is Bubble Tea v2's Elm architecture with SOPS invoked as a subprocess — never as a library. The full charm.land v2 ecosystem is production-stable as of April 2026.

The dominant risk profile is **security, not technical complexity**. Four constraints must be established before any secret-display code ships:
1. `WithAltScreen()` mandatory to prevent scrollback leakage
2. `Update()` must never block — all SOPS calls through `tea.Cmd`
3. Secret values must never appear in subprocess argv (temp files or stdin only)
4. Clipboard auto-clear must be synchronously wired to all exit paths including OS signals

## Stack Highlights

- **bubbletea v2.0.4** — `charm.land/*/v2` import paths (not `github.com/charmbracelet/*`)
- **goccy/go-yaml v1.19.2** — replaces yaml.v3 (author's own recommendation)
- **filippo.io/age v1.3.1** — key parsing only, encryption stays with SOPS subprocess
- **go-git/v5.17.2** — structured git data without requiring `git` binary
- **sahilm/fuzzy** — bubbles/list ships this by default
- **atotto/clipboard** — Wayland limitation (needs xclip/xsel), must document
- **goreleaser v2.15.2** — CGO_ENABLED=0 release pipeline

## Feature Priorities

**Table stakes:** File browser, key list (no decrypt), on-demand decrypt + reveal, edit with re-encrypt, vim navigation, fuzzy search, clipboard with auto-clear, help panel, status bar.

**Differentiators (novel — no existing tool has these):** Diff-before-re-encrypt, SOPS metadata display (zero decrypt cost), git change badges, format-aware rotation, recipient management, cross-file search, secret health checks.

**Anti-features:** GPG key management UI, cloud KMS UI, built-in text editor, Kubernetes sync, live watch mode.

## Architecture

AppModel root with static view registration. All views created at startup. SOPS subprocess calls isolated in `internal/sops/runner.go` as `tea.Cmd` functions. Layout: 30/70 horizontal split. Command palette as overlay.

Build order: SOPS subprocess wrapper → config discovery → AppModel skeleton → file browser → secret viewer → status bar → editor → diff view → recipients → command palette.

## Critical Pitfalls

1. Scrollback leakage — alt screen + never write to stdout directly
2. Blocking Update() — every I/O op must be a tea.Cmd
3. Secret in subprocess argv — ps aux is world-readable, use temp files
4. Clipboard persistence after kill — wire clear to signal handlers
5. .sops.yaml CWD resolution — detect git root, pass --config
6. MAC invalidation from merge conflicts — pre-validate files
7. YAML type coercion — use Node API for round-trip fidelity

## Roadmap Implications

**Phase 1 (Read Loop):** Zero subprocess risk. TUI skeleton, file browser, key names, metadata, navigation, search, git badges. Can ship without age credentials.

**Phase 2 (Write Loop):** Decrypt/reveal, clipboard, edit with diff, rotation. All security patterns established here.

**Phase 3 (Power Features):** Recipients, cross-file search, blame/history, health checks. Multi-file operations with highest risk.

## Gaps to Address

- SOPS stdin editing with `encrypted_regex` — behavior undocumented
- Wayland clipboard limitation — document in README
- `lipgloss.AdaptiveColor` hang (issue #1036) — use explicit colors
- go-git v6 migration — reassess at Phase 3

---
*Synthesized: 2026-04-13*
