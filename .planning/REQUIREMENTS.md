# Requirements: sops-tui

**Defined:** 2026-04-13
**Core Value:** Developers can manage all their SOPS-encrypted secrets from a single terminal interface without remembering CLI flags or writing shell scripts.

## v1 Requirements

### Navigation & Discovery

- [ ] **NAV-01**: User can browse all SOPS-encrypted files discovered via `.sops.yaml` path rules
- [ ] **NAV-02**: User can view key names from encrypted files without decrypting
- [ ] **NAV-03**: User can navigate with vim keybindings (hjkl, g/G, ctrl-d/u)
- [ ] **NAV-04**: User can fuzzy search files and keys with `/` (k9s-style)
- [ ] **NAV-05**: User can view contextual help panel with `?`
- [ ] **NAV-06**: User sees persistent status bar (file path, encryption status, operation feedback)

### Decrypt & View

- [ ] **DEC-01**: User can decrypt and reveal individual secret values on demand
- [ ] **DEC-02**: User can decrypt and reveal all values in a file
- [ ] **DEC-03**: Secret values are masked by default, revealed on keypress
- [ ] **DEC-04**: User can view SOPS metadata (version, lastmodified, recipients, MAC status) without decrypting

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

## v2 Requirements

### Encryption Backend Expansion

- **ENC-01**: Support GPG/PGP encrypted files alongside age
- **ENC-02**: Support cloud KMS (AWS KMS, GCP KMS, Azure Key Vault) as key sources
- **ENC-03**: Support HashiCorp Vault transit engine as SOPS backend

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
| NAV-01 | Pending | Pending |
| NAV-02 | Pending | Pending |
| NAV-03 | Pending | Pending |
| NAV-04 | Pending | Pending |
| NAV-05 | Pending | Pending |
| NAV-06 | Pending | Pending |
| DEC-01 | Pending | Pending |
| DEC-02 | Pending | Pending |
| DEC-03 | Pending | Pending |
| DEC-04 | Pending | Pending |
| EDT-01 | Pending | Pending |
| EDT-02 | Pending | Pending |
| EDT-03 | Pending | Pending |
| EDT-04 | Pending | Pending |
| CLB-01 | Pending | Pending |
| CLB-02 | Pending | Pending |
| CLB-03 | Pending | Pending |
| RCP-01 | Pending | Pending |
| RCP-02 | Pending | Pending |
| RCP-03 | Pending | Pending |
| GIT-01 | Pending | Pending |
| GIT-02 | Pending | Pending |
| GIT-03 | Pending | Pending |
| HLT-01 | Pending | Pending |
| HLT-02 | Pending | Pending |
| HLT-03 | Pending | Pending |

**Coverage:**
- v1 requirements: 26 total
- Mapped to phases: 0
- Unmapped: 26 ⚠️

---
*Requirements defined: 2026-04-13*
*Last updated: 2026-04-13 after initial definition*
