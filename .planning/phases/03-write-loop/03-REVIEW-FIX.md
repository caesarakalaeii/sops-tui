---
phase: 03-write-loop
fixed_at: 2026-04-14T00:00:00Z
review_path: .planning/phases/03-write-loop/03-REVIEW.md
iteration: 1
findings_in_scope: 5
fixed: 5
skipped: 0
status: all_fixed
---

# Phase 3: Code Review Fix Report

**Fixed at:** 2026-04-14
**Source review:** .planning/phases/03-write-loop/03-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 5 (2 Critical, 3 Warning)
- Fixed: 5
- Skipped: 0

## Fixed Issues

### CR-01: Modular bias in `generateAlphanumeric` weakens generated secrets

**Files modified:** `internal/ui/rotate.go`
**Commit:** 3449cee
**Applied fix:** Replaced the naive `charset[b % 62]` loop with rejection sampling. Bytes in the biased tail (248–255) are discarded; accepted range 0–247 divides evenly by 62 (248 % 62 == 0), giving a uniform distribution. An oversampled buffer (`length*2`) is read per iteration to minimise the probability of needing a second `rand.Read` call. Comment documents the math.

---

### CR-02: Non-atomic destination write in `EncryptFile` can corrupt the original file

**Files modified:** `internal/sops/executor.go`
**Commit:** 5c5c1b7
**Applied fix:** Replaced `os.WriteFile(destPath, stdout.Bytes(), 0o600)` with a three-step atomic write: (1) create a sibling temp file in `filepath.Dir(destPath)` so both paths are on the same filesystem, (2) pipe `cmd.Stdout` directly to the temp file (eliminating the in-memory `bytes.Buffer` as well — addresses IN-04 concern), (3) `Sync()` then `os.Rename()` atomically over `destPath`. All error paths close and remove the temp file. Added `path/filepath` import.

---

### WR-01: `h` key bound to both `Collapse` and `Back` — Back binding shadowed

**Files modified:** `internal/keys/bindings.go`
**Commit:** b242013
**Applied fix:** Removed `"h"` from `DetailKeyMap.Back`'s key list, leaving only `"esc"`. `Collapse` retains `"h"` and `"left"` as before. The help text already read `"esc"` so the visible label is unchanged.

---

### WR-02: Dead-code double-handler for Esc in `stateDiff`

**Files modified:** `internal/app/model.go`
**Commit:** 9a97194
**Applied fix:** Removed the `if m.state == stateDiff` block from the Esc priority chain (lines 637–641). The `stateDiff` block above (around line 481) already handles Esc fully via `m.diff.Cancelled()` and returns early with `"Cancelled"` flash. Updated the surrounding comment to document that only `stateHelp` and `stateMetadata` are handled by Priority 2.

---

### WR-03: `DetectFormat` hex pattern matches pure-hex base64 strings

**Files modified:** `internal/ui/rotate.go`
**Commit:** 8212d28
**Applied fix:** Added `&& len(value)%2 == 0` to the hex case in `DetectFormat`. Valid hex strings encode whole bytes and therefore always have an even character count. This prevents misclassifying a base64 value whose characters all fall within `[0-9a-fA-F]` but whose length is odd (e.g., standard base64 output is 44 chars for 32 bytes — even, but the regex match alone was sufficient before; odd-length ambiguous values are now correctly passed to the base64 case). Extended the function doc comment to explain the ambiguity and note the format menu as the user's fallback.

---

_Fixed: 2026-04-14_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
