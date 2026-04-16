---
phase: 05
slug: power-features
status: verified
threats_open: 0
asvs_level: 1
created: 2026-04-16
---

# Phase 05 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| TUI input -> age library validation | User-provided age public key string crosses from textinput to age.ParseX25519Recipient | Untrusted string (public, non-secret) |
| TUI input -> sops subprocess (AddRecipient/RemoveRecipient) | Validated age public key crosses from TUI to CLI args | Age public key (non-secret by design) |
| Decrypted values -> health checker | Plaintext secret values held in memory during health analysis | Secret plaintext (transient, never persisted) |
| User key selection (1-9) -> recipient removal | Number key maps to recipient index in stateRecipientList | Integer index into in-memory slice |
| Decrypted YAML -> flattenYAML -> health analysis | Untrusted YAML structure traversed recursively | Decrypted secret values (transient) |
| Health check plaintext map -> HealthCheckResultMsg | Plaintext values cleared after FindDuplicates; result contains only hashes and metadata | SHA-256 hashes (non-reversible), file paths, key paths |

---

## Threat Register

| Threat ID | Category | Component | Disposition | Mitigation | Status |
|-----------|----------|-----------|-------------|------------|--------|
| T-05-01 | Tampering | sops/executor.go AddRecipient | mitigate | age.ParseX25519Recipient validates bech32 + 32-byte key in RecipientFormModel before any sops call | closed |
| T-05-02 | Information Disclosure | health/checker.go FindDuplicates | mitigate | Duplicate.ValueHash stores 16-char truncated SHA-256 hex, never plaintext; caller clears input map | closed |
| T-05-03 | Denial of Service | sops/executor.go AddRecipient/RemoveRecipient | mitigate | SopsRotateTimeout = 60s; exec.CommandContext cancels subprocess on timeout | closed |
| T-05-04 | Information Disclosure | sops CLI args | accept | Age public keys are not secret — safe to pass as CLI args; private keys never appear in args | closed |
| T-05-05 | Tampering | recipientform.go input | mitigate | age.ParseX25519Recipient validates bech32 + 32-byte encoding; canonical re-serialization check rejects trailing characters | closed |
| T-05-06 | Denial of Service | recipientform.go textinput | mitigate | CharLimit = 200 on textinput; age.ParseX25519Recipient is O(n) on fixed-bound input | closed |
| T-05-07 | Information Disclosure | health.go View | accept | HealthModel displays file paths and key paths only — no plaintext secret values rendered in overlay | closed |
| T-05-08 | Tampering | model.go stateRecipientList | mitigate | Bounds-check `idx < len(m.recipientList)` before accessing recipient slice on number key press | closed |
| T-05-09 | Information Disclosure | model.go runHealthCheck | mitigate | fileValues map cleared via delete loop after FindDuplicates; HealthCheckResultMsg contains no plaintext values | closed |
| T-05-10 | Denial of Service | model.go runHealthCheck | mitigate | Each file decrypt uses context.WithTimeout(ctx, sops.SopsTimeout); total scan bounded by file count x timeout | closed |
| T-05-11 | Information Disclosure | model.go flattenYAML | accept | Decrypted values exist in Go heap during health scan; zeroed by GC after map clear; acceptable for a local TUI tool | closed |
| T-05-12 | Denial of Service | model.go flattenYAML | mitigate | Recursive YAML traversal bounded by file size (sops files are small); no external input controls recursion depth | closed |

*Status: open · closed*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-05-01 | T-05-04 | Age public keys are non-secret by design (they are the public half of an X25519 keypair). Passing them as CLI args to sops subprocess poses no confidentiality risk. Private keys are never passed as CLI arguments. | gsd-security-auditor | 2026-04-16 |
| AR-05-02 | T-05-07 | HealthModel overlay renders only file paths and key paths (non-secret filesystem metadata). Decrypted values are never included in health overlay rendering. All SOPS metadata is already visible in the filesystem to any user who can read the repo. | gsd-security-auditor | 2026-04-16 |
| AR-05-03 | T-05-11 | Decrypted secret values necessarily reside in Go heap during the health check scan (required to compute entropy and hashes). The map is explicitly cleared after analysis. Residual heap memory is reclaimed by GC. This is an inherent tradeoff for a local TUI tool that decrypts secrets for analysis; it is the same risk profile as the existing reveal and edit flows. | gsd-security-auditor | 2026-04-16 |

*Accepted risks do not resurface in future audit runs.*

---

## Verification Evidence

### T-05-01 (mitigate — CLOSED)
`internal/ui/recipientform.go:98` — `age.ParseX25519Recipient(rawInput)` called on Enter before setting `m.confirmed = true`. AddRecipient is only called after RecipientFormModel.Confirmed() is true.

### T-05-02 (mitigate — CLOSED)
`internal/health/checker.go:160-161` — `h := sha256.Sum256([]byte(value)); hash := fmt.Sprintf("%x", h)[:16]`. Duplicate struct field is `ValueHash string` — no plaintext field exists.

### T-05-03 (mitigate — CLOSED)
`internal/sops/executor.go:148` — `const SopsRotateTimeout = 60 * time.Second`. `internal/app/model.go:786,794,817` — `context.WithTimeout(context.Background(), sops.SopsRotateTimeout)` applied before each AddRecipient/RemoveRecipient call.

### T-05-04 (accept — CLOSED)
Documented in Accepted Risks Log as AR-05-01.

### T-05-05 (mitigate — CLOSED)
`internal/ui/recipientform.go:98-109` — `age.ParseX25519Recipient(rawInput)` validates bech32 + key bytes; additionally `recipient.String() != rawInput` check rejects trailing content (stronger than declared mitigation — prevents argument injection).

### T-05-06 (mitigate — CLOSED)
`internal/ui/recipientform.go:43` — `ti.CharLimit = 200`.

### T-05-07 (accept — CLOSED)
Documented in Accepted Risks Log as AR-05-02.

### T-05-08 (mitigate — CLOSED)
`internal/app/model.go:749` — `if idx < len(m.recipientList)` guards array access before `pubkey := m.recipientList[idx]`.

### T-05-09 (mitigate — CLOSED)
`internal/app/model.go:1749` — `delete(fileValues, k)` loop clears map after FindDuplicates. `HealthCheckResultMsg` struct contains `health.HealthCheckResult` which holds only WeakSecret/Duplicate/StaleFile structs — no plaintext value fields.

### T-05-10 (mitigate — CLOSED)
`internal/app/model.go:401,411,444,867,888,1697` — `context.WithTimeout(context.Background(), sops.SopsTimeout)` applied per-file in runHealthCheck.

### T-05-11 (accept — CLOSED)
Documented in Accepted Risks Log as AR-05-03.

### T-05-12 (mitigate — CLOSED)
`internal/app/model.go` — `flattenYAML` recursion is bounded by input map size from goccy/go-yaml YAML unmarshalling. SOPS-encrypted files are bounded by practical limits (users encrypt small secrets files). No user-controlled input affects recursion depth.

---

## Unregistered Threat Flags

None. SUMMARY.md `## Threat Flags` for all four plans reported no new attack surfaces beyond those in the declared threat model.

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-04-16 | 12 | 12 | 0 | gsd-security-auditor (claude-sonnet-4-6) |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-04-16
