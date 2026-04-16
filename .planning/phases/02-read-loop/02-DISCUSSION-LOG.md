# Phase 2: Read Loop - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-14
**Phase:** 02-read-loop
**Areas discussed:** File discovery, Key extraction, Metadata display, Fuzzy search UX

---

## File Discovery

| Option | Description | Selected |
|--------|-------------|----------|
| Parse creation rules | Read .sops.yaml creation_rules[].path_regex, walk repo, match files. Most accurate. | |
| Scan for SOPS markers | Walk all YAML/JSON, check for sops: key. Finds files regardless of config but slower. | |
| Both approaches combined | Use path_regex as primary, scan matched files for sops: marker to confirm encryption. | ✓ |

**User's choice:** Both approaches combined
**Notes:** Dual confirmation — regex matching for discovery scope, marker scanning for encryption status.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Show with badge | Regex-matched-but-unencrypted files shown with dim [unencrypted] badge. | ✓ |
| Hide them | Only show actually encrypted files. Cleaner but hides useful info. | |
| Separate section | Encrypted first, divider, unencrypted below. Clear but adds complexity. | |

**User's choice:** Show with badge
**Notes:** Users see the full picture of what SOPS would manage, useful for spotting files that should be encrypted.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Show keys with note | Open detail view with banner, values visible since nothing to mask. | ✓ |
| Block with message | Don't open detail view, show "not encrypted" message. | |
| You decide | Claude picks. | |

**User's choice:** Show keys with note
**Notes:** Banner reads "Not yet encrypted — matches .sops.yaml rules". Values shown in plaintext.

---

## Key Extraction

| Option | Description | Selected |
|--------|-------------|----------|
| Masked with *** | All encrypted values show ***. Clean, consistent. | |
| Type hints | Show *** (str), *** (int), *** (bool) from SOPS envelope. | ✓ |
| Truncated ciphertext | Show first chars of ENC[AES256_GCM,...]. Debug-oriented. | |

**User's choice:** Type hints
**Notes:** SOPS preserves original types — surfacing them gives users useful context without revealing content.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Hidden from tree | Filter out sops: key, surface via metadata panel. | ✓ |
| Collapsed at bottom | Show sops: as collapsed group node. Transparent but noisy. | |
| You decide | Claude picks. | |

**User's choice:** Hidden from tree
**Notes:** sops: is internal plumbing — dedicated metadata panel (DEC-04) is the right surface.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Show plaintext with marker | Actual value with [plain] badge. | ✓ |
| Show plaintext, no marker | Actual value, no badge. Difference from *** is implicit. | |
| Mask everything equally | *** for all values. Consistent but hides useful info. | |

**User's choice:** Show plaintext with marker
**Notes:** Highlights which keys SOPS left unencrypted via encrypted_regex/unencrypted_regex rules.

---

## Metadata Display

| Option | Description | Selected |
|--------|-------------|----------|
| Info panel on keypress | Press 'i' for full metadata panel. Esc closes. | ✓ |
| Inline under filename | Compact summary below filename in file list. Always visible. | |
| Header in detail view | Metadata header above key tree in detail view. | |

**User's choice:** Info panel on keypress
**Notes:** Shows version, lastmodified, recipient list, MAC status. Clean separation from main views.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Both views | 'i' works in file list and detail view. | ✓ |
| Detail view only | Only after drilling into a file. | |
| You decide | Claude picks. | |

**User's choice:** Both views
**Notes:** Consistent keybinding wherever you are.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Full-screen overlay | Same pattern as help screen. stateMetadata with prevState. | ✓ |
| Centered popup box | Bordered box centered over dimmed content. New UI pattern. | |

**User's choice:** Full-screen overlay
**Notes:** Reuses existing overlay pattern from Phase 1 help screen. No new paradigm.

---

## Fuzzy Search UX

| Option | Description | Selected |
|--------|-------------|----------|
| Inline filter | / activates text input, list filters in real time. k9s-style. | ✓ |
| Overlay search | / opens centered search box. More modal. | |
| Command palette style | VS Code-style full-width palette at top. Extensible. | |

**User's choice:** Inline filter
**Notes:** Esc clears search, Enter selects highlighted item. Matched chars highlighted in accent.

---

| Option | Description | Selected |
|--------|-------------|----------|
| File names only in file list | / filters file names in file list, key paths in detail view. Context-aware. | ✓ |
| Global cross-file search | / searches across ALL files AND keys simultaneously. | |
| File names everywhere | / only searches file names in any view. | |

**User's choice:** File names only in file list (context-aware)
**Notes:** Each view searches its own domain — file names in file list, key paths in detail view.

---

| Option | Description | Selected |
|--------|-------------|----------|
| Accent color on matched chars | ColorAccent blue on matched characters via sahilm/fuzzy positions. | ✓ |
| Bold matched chars | Bold rendering. Works on monochrome but less distinct. | |
| You decide | Claude picks. | |

**User's choice:** Accent color on matched chars
**Notes:** Consistent with existing design system. sahilm/fuzzy provides matched character positions.

---

## Claude's Discretion

- .sops.yaml parsing edge cases (multiple creation rules, nested configs)
- File tree walking performance strategy
- Metadata panel layout and formatting details
- Search input position (top vs bottom)
- Empty search results messaging
- Detail view key path flattening for search

## Deferred Ideas

None — discussion stayed within phase scope
