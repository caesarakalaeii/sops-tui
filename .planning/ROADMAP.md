# Roadmap: sops-tui

## Overview

sops-tui is built in five phases that progressively unlock capability while managing security risk. Phase 1 establishes the skeleton and security groundwork with zero exposure to secret values. Phase 2 adds the read-only file browser. Phase 3 unlocks decryption, editing, and rotation (the security-critical write loop). Phase 4 adds clipboard handling with signal safety and git integration. Phase 5 delivers recipient management and health checks — the highest-risk multi-file operations. Every phase ships something independently useful.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Foundation** - TUI skeleton, SOPS subprocess wrapper, config discovery, security groundwork, startup validation
- [ ] **Phase 2: Read Loop** - File browser, key names without decrypt, SOPS metadata, fuzzy search, full navigation
- [ ] **Phase 3: Write Loop** - On-demand decrypt, reveal, edit with diff, format-aware rotation, re-encryption
- [ ] **Phase 4: Clipboard & Git** - Clipboard with auto-clear and signal safety, git change badges, blame/history, cross-file search
- [ ] **Phase 5: Power Features** - Recipient management, bulk re-key, secret health checks

## Phase Details

### Phase 1: Foundation
**Goal**: The application starts, validates its environment, and provides a navigable skeleton — without ever touching a secret value
**Depends on**: Nothing (first phase)
**Requirements**: HLT-01, HLT-02, NAV-03, NAV-05, NAV-06
**Success Criteria** (what must be TRUE):
  1. Running `sops-tui` with no `sops` binary shows a clear error with installation instructions and exits cleanly
  2. Running `sops-tui` with no age key file shows a clear error with setup instructions and exits cleanly
  3. User can navigate any view using hjkl, g/G, and ctrl-d/u without errors
  4. Pressing `?` opens a contextual help panel listing all keybindings
  5. A persistent status bar is visible on every screen showing file path and operation feedback
**Plans:** 4 plans

Plans:
- [x] 01-01-PLAN.md — Go module setup, design system colors/styles, keybinding contracts
- [x] 01-02-PLAN.md — Startup validation (sops/age/.sops.yaml) and styled stderr error box
- [x] 01-03-PLAN.md — File list and YAML tree detail view components
- [x] 01-04-PLAN.md — Help overlay, status bar, root model wiring, main.go entry point

**UI hint**: yes

### Phase 2: Read Loop
**Goal**: Users can browse all SOPS-encrypted files and inspect their contents without any decryption occurring
**Depends on**: Phase 1
**Requirements**: NAV-01, NAV-02, NAV-04, DEC-03, DEC-04
**Success Criteria** (what must be TRUE):
  1. User can open `sops-tui` in a repository and see all SOPS-encrypted files discovered via `.sops.yaml` path rules
  2. Selecting a file shows all key names with values masked by default — no decryption has occurred
  3. User can view SOPS metadata (version, lastmodified, recipients, MAC status) for any file without decrypting it
  4. Pressing `/` opens a fuzzy search that filters across file names and key names in real time
**Plans:** 2/3 plans executed

Plans:
- [x] 02-01-PLAN.md — SopsDiscoverer + YamlParser backend services, goccy/go-yaml dependency, TreeNode extension
- [x] 02-02-PLAN.md — MetadataModel overlay, SearchModel with fuzzy matching, 5 new named styles
- [ ] 02-03-PLAN.md — AppModel wiring: discovery, parsing, metadata overlay, search, keybindings, state machine

**UI hint**: yes

### Phase 3: Write Loop
**Goal**: Users can decrypt, reveal, edit, and rotate secrets with a safety gate before any write is committed
**Depends on**: Phase 2
**Requirements**: DEC-01, DEC-02, EDT-01, EDT-02, EDT-03, EDT-04
**Success Criteria** (what must be TRUE):
  1. User can reveal a single secret value on demand; it appears in the view and can be hidden again
  2. User can decrypt and reveal all values in a file at once
  3. User can edit a secret value; before re-encryption a diff view is shown requiring explicit confirmation
  4. User can rotate a secret to a format-aware random value (base64, hex, UUID, bcrypt) with confirmation
  5. Any destructive write operation presents a confirmation prompt that can be cancelled without effect
**Plans**: TBD
**UI hint**: yes

### Phase 4: Clipboard & Git
**Goal**: Users can safely copy secrets to clipboard with guaranteed cleanup, and see git state alongside their secrets
**Depends on**: Phase 3
**Requirements**: CLB-01, CLB-02, CLB-03, GIT-01, GIT-02, GIT-03
**Success Criteria** (what must be TRUE):
  1. User can copy a decrypted value to clipboard; the clipboard clears automatically after the configured timeout (default 30s)
  2. Clipboard is cleared synchronously when `sops-tui` exits via any path including SIGINT and SIGTERM
  3. Files with uncommitted git changes display a badge ([M], [A], [?]) in the file browser
  4. User can view git blame and commit history for any secret file from within the TUI
  5. User can fuzzy search across all files and key names simultaneously with `/`
**Plans**: TBD
**UI hint**: yes

### Phase 5: Power Features
**Goal**: Users can manage age recipients across files and audit secret health — the highest-risk multi-file operations
**Depends on**: Phase 4
**Requirements**: RCP-01, RCP-02, RCP-03, HLT-03
**Success Criteria** (what must be TRUE):
  1. User can view the list of age key recipients configured for any file
  2. User can add or remove an age key recipient on a file with confirmation before re-key
  3. User can bulk re-key multiple files to a new recipient set with per-file confirmation
  4. User can run a health check that reports weak secrets, duplicate values across files, and stale (unchanged) values
**Plans**: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> 3 -> 4 -> 5

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation | 4/4 | Complete | - |
| 2. Read Loop | 2/3 | In Progress|  |
| 3. Write Loop | 0/? | Not started | - |
| 4. Clipboard & Git | 0/? | Not started | - |
| 5. Power Features | 0/? | Not started | - |
