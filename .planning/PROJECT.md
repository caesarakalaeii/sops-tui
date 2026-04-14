# sops-tui

## What This Is

A k9s-inspired terminal UI for managing SOPS-encrypted secrets. Works with any repository that has a `.sops.yaml` configuration. Browse, decrypt, edit, rotate, and audit secrets — all from a keyboard-driven TUI built with Go and Bubble Tea.

## Core Value

Developers can manage all their SOPS-encrypted secrets from a single terminal interface without remembering CLI flags or writing shell scripts.

## Requirements

### Validated

- [x] Vim-like keyboard navigation throughout — Validated in Phase 1: Foundation (hjkl, g/G, ctrl-d/u)
- [x] Startup error if `sops` binary missing — Validated in Phase 1: Foundation
- [x] Startup error if age key missing — Validated in Phase 1: Foundation (soft warning, TUI still launches)
- [x] Browse all SOPS-encrypted files discovered via `.sops.yaml` path rules — Validated in Phase 2: Read Loop
- [x] View encrypted keys at a glance without decrypting — Validated in Phase 2: Read Loop
- [x] Fuzzy search across all files and keys (k9s-style `/` search) — Validated in Phase 2: Read Loop

- [x] Decrypt and reveal secret values on demand (single key or full file) — Validated in Phase 3: Write Loop
- [x] Edit secret values with automatic re-encryption via `sops` subprocess — Validated in Phase 3: Write Loop
- [x] Rotate secrets to format-aware random values (base64, hex, UUID, bcrypt, etc.) — Validated in Phase 3: Write Loop
- [x] Diff view before re-encrypting to prevent accidental edits — Validated in Phase 3: Write Loop

### Active
- [ ] Recipient management — add/remove age keys across files, bulk re-key
- [ ] Secret health checks — detect weak secrets, duplicates across files, stale values
- [ ] Clipboard copy with auto-clear after configurable timeout
- [ ] Git integration — blame/history per secret, detect uncommitted changes
- [ ] Vim-like keyboard navigation throughout

### Out of Scope

- GUI / web interface — this is a terminal tool only
- Cloud KMS key management (AWS KMS, GCP KMS, Azure KV) — v1 focuses on age keys
- Kubernetes cluster interaction — this operates on files, not live cluster secrets
- SOPS as a Go library — v1 shells out to `sops` CLI for reliability and version compatibility
- Secret generation templates / policies — format-aware rotation is sufficient for v1

## Context

- **SOPS ecosystem gap**: SOPS has 21k+ GitHub stars but zero interactive terminal tooling. The entire UX today is `sops -d`, `sops -e`, `sops rotate`, and shell scripts.
- **Encryption backend**: v1 targets age encryption (modern, simple). GPG support is a natural v2 extension.
- **File format**: SOPS encrypts individual YAML/JSON values. The TUI needs to parse both the encrypted envelope and the decrypted content.
- **UX inspiration**: k9s (keyboard-driven, vim bindings, resource browser), lazygit (diff views, staging), gpg-tui (key management in terminal).
- **Closest existing tools**: Doppler TUI (SaaS-only, no SOPS), gpg-tui (GPG keys only, 1.7k stars), vaul7y (HashiCorp Vault only).
- **Target users**: DevOps engineers, SREs, and developers managing SOPS-encrypted secrets in Git repositories.

## Constraints

- **Stack**: Go + Bubble Tea (Charm ecosystem) — chosen for single-binary distribution and k9s ecosystem alignment
- **SOPS integration**: Subprocess calls to `sops` CLI — must handle `sops` not being installed gracefully
- **Encryption**: age keys via `~/.config/sops/age/keys.txt` — standard SOPS convention
- **License**: AGPL-3.0
- **Dependencies**: Requires `sops` binary installed and age key available for decryption operations

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go + Bubble Tea over tview | Modern Elm architecture, better composability, active Charm ecosystem | — Pending |
| Subprocess sops over native library | Simpler, always matches user's sops version, less maintenance burden | — Pending |
| Age-only for v1 | Simplifies key management UI; GPG/KMS are v2 extensions | — Pending |
| Generic (any SOPS repo) over bespoke | Maximizes FOSS value; users bring their own `.sops.yaml` | — Pending |
| Format-aware rotation | Different secret types need different generators (base64, hex, UUID, bcrypt) | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-04-14 after Phase 3 completion*
