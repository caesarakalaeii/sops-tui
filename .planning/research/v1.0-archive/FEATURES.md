# Feature Landscape: SOPS TUI

**Domain:** TUI-based secret/key management tool (file-level encryption, git-native workflow)
**Researched:** 2026-04-13
**Reference tools:** k9s, lazygit, gpg-tui, vaul7y, KeePassXC, Bitwarden Vault Health, Doppler

---

## Table Stakes

Features users expect from any secret management TUI. Missing these and users return to the raw `sops` CLI.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| **File browser / discovery** | Users need to see all SOPS-encrypted files without running `find`. Entry point of the entire tool. | Low | Walk directory tree; use `.sops.yaml` creation_rules path globs to scope discovery. Show encrypted/unencrypted status badge. |
| **Key list view (without decryption)** | Show key names from the encrypted YAML/JSON envelope. Knowing *what* keys exist is useful even without revealing values. | Low | SOPS files are partially plaintext — key names are never encrypted. Zero decryption cost. |
| **On-demand decrypt and reveal** | The entire point of the tool: see a value without running `sops -d`. Must support per-key reveal and whole-file decrypt. | Medium | Shell out to `sops -d`. Cache decrypted state in-memory for session; clear on exit. Show values masked by default (`***`), reveal on keypress. |
| **Edit with auto re-encrypt** | SOPS' own `sops <file>` does open-in-editor. The TUI must replicate this: edit value in-place, then re-encrypt on save. | Medium | Shell out to `sops` for re-encryption. Diff view before confirming (see differentiators). Handle `sops` not installed gracefully. |
| **Vim-like navigation** | k9s, lazygit, gpg-tui all use hjkl. Any keyboard-centric TUI targeting DevOps engineers must support this. Non-negotiable for adoption. | Low | `j/k` for up/down in lists, `g/G` for top/bottom, `ctrl-d/ctrl-u` for page scroll. |
| **`/` fuzzy search** | k9s makes this pattern ubiquitous. Users expect to type `/` and filter any list. Charmbracelet `bubbles/list` ships fuzzy search (sahilm/fuzzy) by default. | Low | Apply to file list AND key list. Filter updates in real time. `Esc` to clear. |
| **Clipboard copy** | Every secret management tool (KeePassXC, Bitwarden, 1Password, gpg-tui) ships this. Copying a value to clipboard is faster than reading and retyping it. | Low | Bubble Tea v2 provides `tea.SetClipboard`. Must trigger auto-clear. |
| **Clipboard auto-clear** | KeePassXC defaults to 10s, industry standard is 30s. Leaving credentials in clipboard is a recognized attack surface. This is *expected*, not optional. | Low | Configurable timeout (default 30s). Show countdown in status bar. Clear on exit as fallback. |
| **Contextual help panel** | k9s and lazygit both show available keybindings for the current context. Users cannot be expected to memorize all bindings. | Low | `?` toggles help overlay. Show only bindings relevant to current focused pane. |
| **SOPS binary not found error** | Tool is useless without `sops` installed. Must fail gracefully with clear instructions rather than a panic or cryptic error. | Low | Check for `sops` on startup. Show install instructions if missing. Same for age key not found. |
| **Status bar** | All professional TUIs (k9s, lazygit, gpg-tui) show current context, key mode, and operation feedback in a persistent status bar. | Low | Show: current file path, encryption status, pending operation, last action result. |

---

## Differentiators

Features that distinguish this tool from simply running `sops` on the command line. These are the value-add that justify building a TUI at all.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| **Diff view before re-encrypt** | Lazygit's killer feature applied to secrets: see exactly what changed before saving. Prevents accidental overwrites. No other SOPS tooling offers this. | Medium | Generate before/after diff in memory. Show side-by-side or unified diff pane. Require explicit confirm (`y`) before re-encrypting. |
| **Secret rotation with format awareness** | `sops rotate` rotates the data key but generates no new secret *values*. A TUI that generates context-appropriate values (base64, hex, UUID, bcrypt hash) is a genuine workflow accelerator for DevOps rotation tasks. | Medium-High | Detect format heuristically from current value (base64 → generate base64, UUID pattern → generate UUID, etc.). Let user override. Never auto-apply without confirmation. |
| **Recipient management** | SOPS `updatekeys` and `--add-pgp`/`--rm-pgp` flags exist but require knowing CLI syntax. A TUI that lists current recipients per file and allows add/remove with `.sops.yaml`-driven suggestions is a significant UX win for team workflows. | High | Show age public keys per file from SOPS metadata. Diff recipient lists across files. Bulk re-key multiple files. This is the hardest feature to implement correctly. |
| **Cross-file fuzzy search** | The current SOPS workflow requires decrypting each file separately to find which file holds a given key. A global search across all keys (even just key names, without decrypting) is a gap no existing tool fills. | Medium | Index key names from all discovered files. Optionally extend to decrypted values (opt-in, with explicit warning). Show file + key path in results. |
| **Git integration: uncommitted change detection** | DevOps engineers frequently forget to `git add` after editing secrets. Show a visual indicator on files with uncommitted changes. This is table stakes in lazygit but novel for secret management tools. | Low | Run `git status --porcelain` on file discovery. Badge files with [M], [A], [?] indicators. Refresh on file change. |
| **Git integration: blame / history per key** | Knowing *when* a secret was last changed and *by whom* is critical for security audits. `git log -p -- file` exists but is tedious. Surfacing last-modified commit per file in the TUI saves manual investigation. | Medium | Shell out to `git log --follow -1 -- <file>`. Show last commit hash, author, date per file row. Full log in a popup pane on demand. |
| **Secret health checks** | KeePassXC has "Database Reports"; Bitwarden has "Vault Health Reports"; 1Password has "Watchtower". Users expect a health overview. For SOPS specifically: detect weak/short values, detect identical values across files (same secret shared insecurely), detect stale values via last-modified date. | High | Weakness: entropy check on decrypted values. Duplicates: hash comparison across files (requires decrypting all). Stale: `git log` date vs configurable age threshold. Opt-in only — requires full decrypt. |
| **SOPS metadata view** | SOPS embeds `lastmodified`, `version`, key fingerprints, and MAC in every file. Surfacing this in the TUI gives users an audit trail without opening raw files. This is unique to SOPS and not available in any comparable tool. | Low | Parse SOPS metadata from encrypted file (no decryption needed). Show in detail pane: version, last_modified, key fingerprints, MAC status. |

---

## Anti-Features

Features to deliberately NOT build in v1. These are scope traps that would delay shipping without proportionate user value.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| **GPG / PGP key management** | gpg-tui already exists and does this well. GPG key management is a distinct domain from SOPS secret management. Conflating them adds complexity with no SOPS-specific value. | Defer to v2 as an extension. Show a clear message: "GPG support planned for v2." |
| **Cloud KMS key management** (AWS KMS, GCP KMS, Azure KV UI) | Requires cloud SDK authentication, region config, IAM policy awareness. This is a product in itself. SOPS with cloud KMS is a minority use case vs. age keys. | Document that cloud KMS-encrypted files can still be decrypted if the user has valid ambient credentials — SOPS handles this transparently. |
| **Built-in text editor** | Embedding a full text editor (like micro or a vim emulation layer) is complex, hard to test, and duplicates `sops <file>` which already opens `$EDITOR`. The "edit" flow should open `$EDITOR` via subprocess, not re-implement it. | Shell out to `$EDITOR` for full-file edits. Use in-place value editing (single field) for the TUI edit flow. |
| **Secret template / policy engine** | Defining "this secret should always be a 256-bit base64 value" as a policy adds a configuration system that requires documentation, validation, and ongoing maintenance. | Format-aware rotation handles the generation side. Policy enforcement is a v3+ concern. |
| **Web UI or REST API** | Scope creep and security surface. The tool's value is precisely that it is a terminal-only, local tool with no network exposure. | Reject explicitly. File an issue if requested to track future demand. |
| **Multi-user / RBAC** | SOPS itself has no RBAC — access control is via key distribution. Reimplementing RBAC on top of SOPS files in a TUI would be misleading about the actual security model. | Document the SOPS security model clearly. Recipient management (adding/removing age keys) is the correct mechanism. |
| **Live secret reload / watch mode** | Polling or `inotify`-watching encrypted files adds complexity and can cause confusing state (file changed on disk while TUI has decrypted state). | Refresh on explicit user action (`r` to reload). Show [modified on disk] indicator if file mtime changes. |
| **Kubernetes Secret sync** | Applying secrets to live Kubernetes clusters blurs the tool's identity from "file-level secret manager" to "Kubernetes secrets management tool." This is a different product category. | Out of scope explicitly. The right tool for this is Flux/ArgoCD with SOPS, not this TUI. |

---

## Feature Dependencies

```
File discovery
  └── Key list view                  (needs discovered file list)
        ├── On-demand decrypt         (needs key list to select)
        │     ├── Clipboard copy      (needs decrypted value)
        │     ├── Edit + re-encrypt   (needs decrypted value)
        │     │     └── Diff view     (needs before/after decrypted state)
        │     └── Secret rotation     (needs decrypted value + format detection)
        ├── SOPS metadata view        (no decryption required, from file structure)
        └── Cross-file fuzzy search   (key names: no decrypt; values: needs decrypt)

Recipient management
  ├── File discovery                  (needs file list to operate on)
  └── SOPS metadata view              (recipient fingerprints come from metadata)

Secret health checks
  ├── On-demand decrypt               (weakness + duplicate checks require values)
  └── Git integration                 (stale check requires git history)

Git integration
  ├── File discovery                  (operates on discovered file paths)
  └── Uncommitted change detection    (simpler — git status only, no decrypt needed)

Clipboard copy
  └── Clipboard auto-clear            (hard dependency — copy without auto-clear is a security regression)
```

---

## MVP Recommendation

The v1 MVP must establish the TUI as genuinely useful before tackling complex features. Ship the full navigation+read loop first, then the write loop.

**Phase 1 — Read loop (no decryption):**
1. File browser with discovery via `.sops.yaml`
2. Key list view (key names only, no decrypt)
3. SOPS metadata view (version, last_modified, recipients)
4. Vim navigation + `/` fuzzy search
5. Git uncommitted change indicators
6. Contextual help, status bar, error handling

**Phase 2 — Write loop (decrypt + edit):**
7. On-demand decrypt and reveal (per-key and full-file)
8. Clipboard copy with auto-clear (30s default)
9. Edit with diff view + re-encrypt
10. Secret rotation with format detection

**Phase 3 — Power features:**
11. Recipient management (add/remove age keys, bulk re-key)
12. Cross-file fuzzy search (including decrypted values, opt-in)
13. Git blame/history per file
14. Secret health checks (weakness, duplicates, staleness)

**Defer to v2:**
- GPG/PGP support
- Cloud KMS UI
- Secret policy engine

**Rationale for ordering:**

Phase 1 is zero-risk: no subprocess calls to `sops`, no decryption, no potential for data corruption. It demonstrates the core navigation UX and lets early users try the tool without credentials configured.

Phase 2 introduces subprocess risk (sops not installed, decryption failure, re-encryption errors) and security concerns (in-memory secret handling). Deferring until Phase 1 UX is solid means the error handling is layered on a stable foundation.

Phase 3 features all require either full-file decryption of many files (health checks), complex git subprocess orchestration (blame), or multi-file write operations (recipient management). These are the features most likely to uncover edge cases — batch them after the single-file read/write loop is proven.

---

## Sources

- [gpg-tui features (README)](https://github.com/orhun/gpg-tui/blob/master/README.md) — HIGH confidence, official source
- [k9s TUI design patterns](https://k9scli.io/) — HIGH confidence, official source
- [lazygit design philosophy](https://jesseduffield.com/Lazygit-5-Years-On/) — HIGH confidence, author's own retrospective
- [Charmbracelet Bubbles list component (fuzzy search)](https://pkg.go.dev/github.com/charmbracelet/bubbles/list) — HIGH confidence, official Go docs
- [SOPS updatekeys / recipient management](https://github.com/getsops/sops) — HIGH confidence, official source
- [KeePassXC health reports](https://keepassxc.org/docs/KeePassXC_GettingStarted) — HIGH confidence, official docs
- [Bitwarden vault health reports](https://bitwarden.com/help/reports/) — HIGH confidence, official docs
- [Clipboard auto-clear: KeePassXC issue](https://github.com/keepassxreboot/keepassxc/issues/4126) — MEDIUM confidence (community discussion confirming 30s default expectation)
- [SOPS metadata structure](https://blog.gitguardian.com/a-comprehensive-guide-to-sops/) — MEDIUM confidence, verified against SOPS docs
- [vaul7y TUI for Vault](https://github.com/dkyanakiev/vaul7y) — LOW confidence (sparse documentation, early-stage project)
