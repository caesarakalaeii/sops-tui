---
phase: "03-write-loop"
plan: "01"
subsystem: "sops-executor, detail-reveal, keybindings, app-model"
tags: ["sops", "decrypt", "reveal", "keybindings", "tdd", "security"]
dependency_graph:
  requires: []
  provides:
    - "sops.DecryptKey — single-value decrypt subprocess"
    - "sops.DecryptFile — full-file decrypt subprocess"
    - "sops.SetKey — atomic single-key re-encrypt via --value-stdin"
    - "sops.EncryptFile — full-file encrypt for $EDITOR flow"
    - "sops.IsArrayIndexedKeyPath — array-index guard (Open Question 1)"
    - "TreeNode.Revealed / TreeNode.DecryptedValue fields"
    - "DetailModel.ClearAllRevealed / RevealNode / MaskNode / RevealAllNodes / AnyRevealed"
    - "RevealRequestMsg / RevealAllRequestMsg / DecryptKeyMsg / DecryptAllMsg / ReEncryptDoneMsg"
    - "DetailKeyMap.Reveal / RevealAll / Edit / EditFile / Rotate bindings"
    - "RevealedValueStyle / RevealedIconStyle / DiffAddedStyle / DiffRemovedStyle / ConfirmPromptStyle"
    - "stateDiff / stateEdit session states"
  affects:
    - "internal/app/model.go — Esc clears revealed values before file-list transition"
    - "internal/ui/detail.go — renderRow shows plaintext + lock-open for revealed nodes"
tech_stack:
  added:
    - "golang.org/x/crypto v0.50.0 — bcrypt rotation (Plan 03)"
  patterns:
    - "tea.Cmd for async sops subprocesses (RevealRequestMsg → DecryptKeyMsg)"
    - "Reveal by keyPath matching (not cursor index) — Pitfall 2 mitigation"
    - "ClearAllRevealed on Esc-to-file-list — T-03-02 memory-leak prevention"
    - "context.WithTimeout(ctx, SopsTimeout) — T-03-04 hung-process prevention"
    - "--value-stdin for SetKey — T-03-01 prevents secret in process listings"
    - "json.Marshal(newValue) for SetKey stdin — Pitfall 1 correct JSON encoding"
    - "strings.TrimRight(stdout, newline) for DecryptKey — Pitfall 5 trailing newline"
key_files:
  created:
    - "internal/sops/executor.go"
    - "internal/sops/executor_test.go"
    - "internal/keys/bindings_reveal_test.go"
    - "internal/ui/detail_reveal_test.go"
    - "internal/app/model_reveal_test.go"
  modified:
    - "internal/ui/styles.go — 9 new named styles"
    - "internal/ui/detail.go — TreeNode extension, reveal methods, renderRow, r/R handlers"
    - "internal/keys/bindings.go — 5 new bindings on DetailKeyMap"
    - "internal/app/model.go — new states, message types, decrypt handlers, Esc clear"
    - "go.mod / go.sum — golang.org/x/crypto v0.50.0 added"
decisions:
  - "RevealRequestMsg returned from DetailModel.Update (not AppModel) — detail stays ignorant of sops package, AppModel owns subprocess dispatch"
  - "parseDecryptedValues inline in model.go (not parser package) — avoids cross-package dependency for what is a one-off transformation"
  - "ParsedFileForTest exported helper — allows external test packages to put AppModel into stateDetail without real file I/O"
  - "StateDiff/StateEdit exported constants — allow _test packages to verify new session state values"
metrics:
  duration: "7 minutes"
  completed_date: "2026-04-14"
  tasks_completed: 2
  files_modified: 10
---

# Phase 3 Plan 1: SOPS Executor and On-Demand Decrypt Summary

**One-liner:** SOPS subprocess wrapper (DecryptKey/DecryptFile/SetKey/EncryptFile) with async r/R reveal-toggle in detail view, ClearAllRevealed on Esc, and all Phase 3 styles/keybindings registered.

## What Was Built

### Task 1: SOPS Executor Subprocess Wrapper

`internal/sops/executor.go` provides four exported functions wrapping the `sops` CLI:

- `DecryptKey(ctx, filePath, keyPath)` — runs `sops decrypt --extract '["key"]'`, strips trailing newline
- `DecryptFile(ctx, filePath)` — runs `sops decrypt`, returns raw decrypted YAML bytes
- `SetKey(ctx, filePath, keyPath, value)` — runs `sops set --value-stdin`, passes JSON-encoded value via stdin
- `EncryptFile(ctx, srcPath, destPath)` — runs `sops encrypt`, writes ciphertext to destPath

Supporting utilities:
- `dotPathToIndex` — converts `"database.password"` → `["database"]["password"]`
- `IsArrayIndexedKeyPath` — guards against array-indexed key paths that sops set cannot handle
- `SopsTimeout = 30 * time.Second` — applied by callers via `context.WithTimeout`

Security mitigations verified:
- T-03-01: `--value-stdin` exclusively; no secret in process listing
- T-03-04: `SopsTimeout` constant for hung-process prevention
- Pitfall 1: `json.Marshal(newValue)` for correct JSON encoding
- Pitfall 5: `strings.TrimRight(stdout, "\n")` strips sops trailing newline
- Open Question 2 documented in code comment above `SetKey`

### Task 2: Reveal Keybindings, TreeNode Extension, Styles, Async Wiring

**styles.go** — 9 new named styles added: `RevealedValueStyle`, `RevealedIconStyle`, `DiffAddedStyle`, `DiffRemovedStyle`, `DiffKeyStyle`, `DiffContextStyle`, `EditInputStyle`, `ConfirmPromptStyle`, `FormatMenuStyle`.

**bindings.go** — `DetailKeyMap` extended with: `Reveal` (r), `RevealAll` (R), `Edit` (e), `EditFile` (E), `Rotate` (X). All registered in `DefaultDetailKeyMap`. `ShortHelp`/`FullHelp` updated to include new actions group.

**detail.go** — `TreeNode` gains `Revealed bool` and `DecryptedValue string`. `DetailModel` gains:
- `ClearAllRevealed()` — recursively zeros all revealed state (D-04, T-03-02)
- `RevealNode(keyPath, value)` — matches by keyPath (not cursor), prevents Pitfall 2 race
- `MaskNode(keyPath)` / `MaskAllNodes()` — targeted and bulk mask
- `RevealAllNodes(map)` — bulk reveal from decrypted YAML value map
- `AnyRevealed()` — state query for R toggle logic

`renderRow` updated: revealed encrypted leaves show `RevealedValueStyle.Render(value) + "  " + RevealedIconStyle.Render("🔓")`.

`Update` handles `r` and `R`: returns `RevealRequestMsg` or `RevealAllRequestMsg` tea.Cmd, or masks inline.

**model.go** — New session states `stateDiff`/`stateEdit` (exported as `StateDiff`/`StateEdit` for tests). New message types: `DecryptKeyMsg`, `DecryptAllMsg`, `ReEncryptDoneMsg`, `RevealRequestMsg`, `RevealAllRequestMsg`, `ParsedFileForTest` helper. Handlers dispatch async `sops.DecryptKey`/`sops.DecryptFile` with `context.WithTimeout`. Esc Priority 3 now calls `m.detail.ClearAllRevealed()` before transitioning (D-04).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical Functionality] `parseDecryptedValues` placed inline in model.go**
- **Found during:** Task 2 implementation
- **Issue:** Plan specified "parser package or inline in model.go" — chose inline to avoid creating a cross-package dependency for a single-use function
- **Fix:** Added `parseDecryptedValues` + `collectLeafValues` to `internal/app/model.go`
- **Files modified:** `internal/app/model.go`
- **Commit:** 40baf68

**2. [Rule 2 - Missing Critical Functionality] `ParsedFileForTest` exported helper**
- **Found during:** Task 2 test writing
- **Issue:** `model_reveal_test.go` (package `app_test`) needs to drive `AppModel` into `stateDetail` by sending `FilesParsedMsg` with test nodes — requires building a `parser.ParsedFile` externally
- **Fix:** Exported `ParsedFileForTest(nodes []ui.TreeNode) parser.ParsedFile` in `model.go`
- **Files modified:** `internal/app/model.go`
- **Commit:** 40baf68

**3. [Rule 2 - Missing Critical Functionality] `StateDiff`/`StateEdit` exported constants**
- **Found during:** Task 2 test writing
- **Issue:** `TestModelContainsStateDiff` in `app_test` package cannot reference unexported `stateDiff`/`stateEdit`
- **Fix:** Added `const StateDiff = stateDiff; const StateEdit = stateEdit` in `model.go`
- **Files modified:** `internal/app/model.go`
- **Commit:** 40baf68

## Known Stubs

None. All wired components produce real behavior. The reveal flow is fully functional end-to-end (pending `sops` binary + age key available in the user's environment — this is a runtime prerequisite, not a stub).

## Threat Flags

All threat items from the plan's threat model were addressed:

| Flag | File | Mitigation Applied |
|------|------|--------------------|
| T-03-01 | `internal/sops/executor.go` | `--value-stdin` used exclusively in `SetKey`; `cmd.Stdin = bytes.NewReader(jsonVal)` |
| T-03-02 | `internal/ui/detail.go`, `internal/app/model.go` | `ClearAllRevealed()` called on every Esc-to-file-list transition; zeroes `DecryptedValue` |
| T-03-03 | `internal/sops/executor.go` | Key paths from YAML parser (already sanitized); accepted |
| T-03-04 | `internal/app/model.go` | `context.WithTimeout(context.Background(), sops.SopsTimeout)` on every subprocess |
| T-03-05 | Not yet applicable | `crypto/rand` usage comes in Plan 03 (rotation); no `math/rand` introduced here |

## Self-Check: PASSED

All files confirmed present on disk:
- `internal/sops/executor.go` — FOUND
- `internal/sops/executor_test.go` — FOUND
- `internal/ui/detail_reveal_test.go` — FOUND
- `internal/app/model_reveal_test.go` — FOUND
- `internal/keys/bindings_reveal_test.go` — FOUND
- `.planning/phases/03-write-loop/03-01-SUMMARY.md` — FOUND

All commits confirmed in git history:
- `cc7e54c` — test(03-01): RED phase executor tests
- `b02387a` — feat(03-01): SOPS executor implementation
- `7fa6fef` — test(03-01): RED phase reveal/mask tests
- `40baf68` — feat(03-01): reveal/mask/styles/keybindings/model wiring
