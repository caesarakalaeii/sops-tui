---
phase: 03-write-loop
reviewed: 2026-04-14T00:00:00Z
depth: standard
files_reviewed: 15
files_reviewed_list:
  - internal/sops/executor.go
  - internal/sops/executor_test.go
  - internal/ui/detail.go
  - internal/ui/detail_test.go
  - internal/ui/detail_reveal_test.go
  - internal/ui/diff.go
  - internal/ui/diff_test.go
  - internal/ui/rotate.go
  - internal/ui/rotate_test.go
  - internal/ui/styles.go
  - internal/keys/bindings.go
  - internal/keys/bindings_reveal_test.go
  - internal/app/model.go
  - internal/app/model_test.go
  - internal/app/model_reveal_test.go
findings:
  critical: 2
  warning: 3
  info: 4
  total: 9
status: issues_found
---

# Phase 3: Code Review Report

**Reviewed:** 2026-04-14
**Depth:** standard
**Files Reviewed:** 15
**Status:** issues_found

## Summary

This phase implements the write loop: inline value editing (`e`), `$EDITOR` full-file editing (`E`), secret rotation (`X`), and a diff/confirm overlay before every re-encryption. The architecture is sound — `sops set --value-stdin` prevents secrets appearing in process listings, temp files use `0600` permissions, `context.WithTimeout` is applied to every subprocess call, and the format detection order (bcrypt before base64) is correctly documented and tested.

Two security issues were found: a modular bias in the alphanumeric random generator, and a non-atomic file write in `EncryptFile` that can corrupt the destination file on partial write. Three warnings cover a keybinding conflict, a dead-code double-handler, and a detection ordering nuance in `DetectFormat`. Four info items note minor code quality issues.

---

## Critical Issues

### CR-01: Modular bias in `generateAlphanumeric` weakens generated secrets

**File:** `internal/ui/rotate.go:167`
**Issue:** The alphanumeric charset has 62 characters. Random bytes from `crypto/rand` are in `[0, 255]`. `255 % 62 = 7`, so characters at index `0–7` of the charset (`a`–`h`) are generated with probability `5/256` instead of `4/256` — a ~25% relative over-representation. For a 32-character secret this produces measurable bias and violates the `T-03-13` requirement for cryptographically uniform randomness.

**Fix:** Use rejection sampling to discard bytes that fall in the biased tail:
```go
func generateAlphanumeric(length int) (string, error) {
    const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    const charsetLen = byte(len(charset)) // 62
    result := make([]byte, length)
    buf := make([]byte, length*2) // oversample to reduce retry probability
    generated := 0
    for generated < length {
        if _, err := rand.Read(buf); err != nil {
            return "", fmt.Errorf("generate alphanumeric: %w", err)
        }
        for _, b := range buf {
            // Reject bytes in the biased tail (255 - 255%62 = 248, so accept 0..247)
            if b < 248 {
                result[generated] = charset[int(b)%len(charset)]
                generated++
                if generated == length {
                    break
                }
            }
        }
    }
    return string(result), nil
}
```

Alternatively, use `math/big.Int` with `crypto/rand.Int`:
```go
max := big.NewInt(int64(len(charset)))
for i := range result {
    n, err := rand.Int(rand.Reader, max)
    if err != nil { ... }
    result[i] = charset[n.Int64()]
}
```

---

### CR-02: Non-atomic destination write in `EncryptFile` can corrupt the original file

**File:** `internal/sops/executor.go:164-168`
**Issue:** `os.WriteFile(destPath, stdout.Bytes(), 0o600)` writes directly to the destination path, which is typically the original encrypted file. If the write is interrupted (process killed, disk full, power loss), `destPath` is left truncated or partially written — destroying the user's only copy of the encrypted file. The encrypted backup is gone and the edit is lost.

```go
// Current (dangerous):
if err := os.WriteFile(destPath, stdout.Bytes(), 0o600); err != nil {
    return fmt.Errorf("sops encrypt: write dest: %w", err)
}
```

**Fix:** Write to a sibling temp file, then atomically rename over the destination:
```go
// Safe atomic write:
dir := filepath.Dir(destPath)
tmp, err := os.CreateTemp(dir, ".sops-tui-enc-*.tmp")
if err != nil {
    return fmt.Errorf("sops encrypt: create temp: %w", err)
}
tmpPath := tmp.Name()
if err := tmp.Chmod(0o600); err != nil {
    tmp.Close(); os.Remove(tmpPath)
    return fmt.Errorf("sops encrypt: chmod temp: %w", err)
}
if _, err := tmp.Write(stdout.Bytes()); err != nil {
    tmp.Close(); os.Remove(tmpPath)
    return fmt.Errorf("sops encrypt: write temp: %w", err)
}
if err := tmp.Sync(); err != nil { // flush to disk before rename
    tmp.Close(); os.Remove(tmpPath)
    return fmt.Errorf("sops encrypt: sync temp: %w", err)
}
tmp.Close()
if err := os.Rename(tmpPath, destPath); err != nil {
    os.Remove(tmpPath)
    return fmt.Errorf("sops encrypt: rename: %w", err)
}
```

Note: `os.Rename` is atomic on POSIX filesystems when src and dst are on the same device, which is guaranteed by using `filepath.Dir(destPath)` for the temp file.

---

## Warnings

### WR-01: `h` key is bound to both `Collapse` and `Back` — the Back binding will shadow Collapse on expanded nodes

**File:** `internal/keys/bindings.go:214-220`
**Issue:** `DetailKeyMap.Collapse` has keys `["h", "left"]` and `DetailKeyMap.Back` also has keys `["esc", "h"]`. In `detail.go`, `key.Matches(msg, m.keys.Collapse)` is evaluated before `key.Matches(msg, m.keys.Back)` — so pressing `h` collapses an expanded node rather than navigating back. However in `AppModel.Update` the Esc-chain handler at `model.go:621` intercepts `msg.String() == "esc"` directly (bypassing the detail key routing), so the `Back` binding's `h` key is the only way `h` reaches the detail model — where it hits `Collapse` first. The `Back` binding's `h` key is therefore silently swallowed by `Collapse` and never actually navigates back; users must press `Esc` to go back. This is a misleading help text entry: `FullHelp()` shows `Back` bound to `esc/h` but `h` does not actually navigate back.

**Fix:** Remove `"h"` from the `Back` binding — leave only `"esc"`:
```go
Back: key.NewBinding(
    key.WithKeys("esc"),
    key.WithHelp("esc", "back to file list"),
),
```

---

### WR-02: Dead-code double-handler for Esc in `stateDiff` can mask future logic errors

**File:** `internal/app/model.go:637-641`
**Issue:** The `stateDiff` key-routing block (lines 443–487) handles Esc by checking `m.diff.Cancelled()` after routing the key to `DiffModel.Update`. The function returns early (`return m, nil` at line 486) when cancelled, so execution never reaches the second Esc handler at line 637–641 for the `stateDiff` case. The second handler is dead code. While harmless today, its presence suggests the two cancel paths are not in sync, and a future refactor that removes the early return would silently produce a double state transition (flash "Cancelled" twice, state changed twice).

**Fix:** Remove the `stateDiff` case from the Esc priority chain since it is fully handled by the stateDiff block above:
```go
// In the Esc priority chain at line ~636, remove:
if m.state == stateDiff {
    m.state = m.prevState
    m.status, _ = m.status.Flash("Cancelled")
    return m, nil
}
```

---

### WR-03: `DetectFormat` hex pattern matches pure-hex base64 strings — detection may misclassify rotated base64 secrets

**File:** `internal/ui/rotate.go:47`
**Issue:** `reHex = regexp.MustCompile('^[0-9a-fA-F]{32,}$')` matches any string of 32+ hex characters. A base64-encoded value whose random bytes happen to decode to only `[0-9a-fA-F]` characters (probability ~`(16/64)^22 ≈ 0` for long values, but non-negligible for exactly 32-char values) would be detected as `FormatHex` and rotated as hex instead of base64. More concretely: a standard base64 value using only characters `[0-9A-Fa-f]` (valid in both hex and base64 alphabets) will be classified as hex since hex is checked first. After rotation the key's value format changes silently from base64 to hex.

**Fix:** Add a length guard to the hex pattern that requires an even number of chars (valid hex always encodes full bytes), and document the ambiguity clearly:
```go
// Hex strings encode full bytes, so length must be even and 32+ chars.
reHex = regexp.MustCompile(`^[0-9a-fA-F]{32,}$`)
// In DetectFormat, add a length-parity check before accepting as hex:
case reHex.MatchString(value) && len(value)%2 == 0:
    return FormatHex
```
For the base64/hex ambiguity (same characters, both valid), consider adding an explicit length discriminant: 64-char pure-hex strings are ambiguous with 48-byte base64 — add an explicit comment that the format menu is the user's fallback when auto-detection is wrong.

---

## Info

### IN-01: Incomplete ANSI stripper in tests may produce false negatives

**File:** `internal/ui/detail_test.go:246-263`
**Issue:** `stripAnsi` handles only terminal code terminators `m`, `K`, `H`, `J`. Lipgloss v2 may emit additional CSI sequences (e.g., cursor movement sequences ending in `A`, `B`, `C`, `D`; erase sequences ending in `X`). Unrecognized terminators leave `inEsc = true` permanently, stripping all subsequent non-escape characters from the string. This would cause `assert.True(t, strings.Contains(stripped, ...))` to fail incorrectly in CI environments that emit richer escape codes.

**Fix:** Either widen the terminator set to all ASCII letters (CSI sequences always end with a letter), or use the `golang.org/x/text` or a well-tested ANSI stripping library:
```go
func stripAnsi(s string) string {
    var result strings.Builder
    inEsc := false
    for _, r := range s {
        if r == '\x1b' {
            inEsc = true
            continue
        }
        if inEsc {
            // All CSI sequences end with a letter (A-Z, a-z)
            if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
                inEsc = false
            }
            continue
        }
        result.WriteRune(r)
    }
    return result.String()
}
```

---

### IN-02: Nil-context guards in `executor.go` are dead paths in production

**File:** `internal/sops/executor.go:64-66`, `88-90`, `118-120`, `151-153`
**Issue:** Every exported function guards `if ctx == nil { ctx = context.Background() }`. Per the Go standard library convention, passing a nil context is a programming error and `context.Background()` should be passed explicitly. The nil guard exists only to support the test helper pattern `DecryptKey(nil, ...)` (annotated `//nolint:staticcheck`). This makes the nil guard effectively test scaffolding leaked into production code.

**Fix:** Remove nil-context guards from production functions. Update tests to pass `context.Background()` explicitly:
```go
// In tests:
_, err := DecryptKey(context.Background(), "/nonexistent/path.yaml", "key")
```

---

### IN-03: `collectLeafValues` and `walkMapSlice` are near-duplicate functions

**File:** `internal/app/model.go:785-810`, `914-931`
**Issue:** Both `collectLeafValues` (used in `parseDecryptedValues`) and `walkMapSlice` (used in `flattenYAMLToMap`) implement the same recursive yaml.MapSlice traversal with dot-joined key path construction, both skip the `sops` top-level key, and both collect leaf values into a `map[string]string`. They differ only in how non-string scalars are handled: `collectLeafValues` has an explicit `case int, int64, float64, bool:` branch, while `walkMapSlice` uses `fmt.Sprintf("%v", v)` for all non-map values.

**Fix:** Consolidate into one function. The `fmt.Sprintf("%v", v)` approach of `walkMapSlice` is already more general:
```go
// Replace both with a single walkYAMLToMap and use it everywhere.
func walkYAMLToMap(ms yaml.MapSlice, prefix string, out map[string]string) {
    for _, item := range ms {
        key := fmt.Sprintf("%v", item.Key)
        if prefix == "" && key == "sops" {
            continue
        }
        path := key
        if prefix != "" {
            path = prefix + "." + key
        }
        switch v := item.Value.(type) {
        case yaml.MapSlice:
            walkYAMLToMap(v, path, out)
        default:
            out[path] = fmt.Sprintf("%v", v)
        }
    }
}
```

---

### IN-04: `EncryptFile` buffers stdout in-memory for potentially large files

**File:** `internal/sops/executor.go:155-169`
**Issue:** `cmd.Stdout = &stdout` buffers the entire encrypted file output in a `bytes.Buffer` before writing it to disk. For large secrets files this doubles peak memory consumption. The CLAUDE.md project notes warn: "Never buffer stdout/stderr with `Output()` for large files; use `Pipe()` + goroutine reads to avoid deadlock." The current implementation uses `bytes.Buffer` (equivalent to `Output()` buffering) for stdout.

**Fix:** Pipe stdout directly to the temp file (see CR-02 fix which already writes through a temp file; combining both fixes eliminates the in-memory buffer):
```go
cmd := exec.CommandContext(ctx, "sops", "encrypt", srcPath)
cmd.Stdout = tmp // pipe directly to the temp file writer
cmd.Stderr = &stderr
```
This also eliminates the second allocation at `stdout.Bytes()`.

---

_Reviewed: 2026-04-14_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
