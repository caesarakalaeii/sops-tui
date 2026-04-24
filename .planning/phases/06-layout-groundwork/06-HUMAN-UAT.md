---
status: partial
phase: 06-layout-groundwork
source: [06-VERIFICATION.md]
started: 2026-04-24T09:45:00Z
updated: 2026-04-24T09:45:00Z
---

## Current Test

[awaiting human testing]

## Tests

### 1. Manual smoke at 40x12 — cycle all v1.0 primary views

**Context:** Per D-15 in `06-CONTEXT.md`, the automated resize goldens only exercise the
empty-state `"No SOPS files found"` path (test cwd has no `.sops.yaml`). Pitfall 1 in
`.planning/research/PITFALLS.md` explicitly warns that teatest snapshots at 80×24 can
hide the "content painted under the status bar" bug class — a real-terminal smoke at
extreme sizes is required before trusting Phase 6 as complete.

**Setup:**
- Check out current HEAD (commit `057d7b9` or later) in a repo that has a real `.sops.yaml`
  plus an age key in `~/.config/sops/age/keys.txt`.
- Launch: `go run ./cmd/sops-tui`
- Resize the terminal to 40 columns × 12 rows.

**Steps (cycle through each view):**
1. File list — default view; Page Up/Down; arrow keys
2. Detail (Enter on a file) — scroll through parsed nodes
3. Help overlay (`?`) — dismiss with `?` or `esc`
4. Diff view (`d` on a file) — scroll if content is long
5. Metadata view (`m`) — scroll
6. History view (`h`) — scroll
7. Health view (`H` or equivalent)
8. Recipient form (via the rotate/add-recipient keybinding)

**Expected:** Every view renders with the status bar pinned to the bottom row. No content
is painted under the status bar. No visible text is clipped by the bottom of the terminal.
Breadcrumbs / titles fit within the 40-column width (wrapping or truncation is acceptable
as long as the status bar is intact).

**Result:** [pending]

---

### 2. Manual smoke at 200x60 — cycle all v1.0 primary views

**Context:** Pitfall 1 warns the bug class "manifests at 120×40 when hidden rows below
the status bar start receiving content". Testing at 200×60 catches any failure where the
body height arithmetic paints overflow into the terminal's bottom rows.

**Setup:**
- Same repo state as Test #1.
- Launch: `go run ./cmd/sops-tui`
- Resize the terminal to 200 columns × 60 rows.

**Steps:** Same view cycle as Test #1 (file list → detail → help → diff → metadata →
history → health → recipient form).

**Expected:** Every view renders identically to how it rendered pre-Phase 6 (commit
`776a7b9` is a good before-snapshot if side-by-side comparison is needed). The status bar
occupies exactly the last row. All body content sits above the status bar with no overlap.
Nothing is painted in rows below the status bar. Mouse/terminal resize events (try
shrinking to 120×40 then back to 200×60) preserve layout correctness across the transition.

**Result:** [pending]

## Summary

total: 2
passed: 0
issues: 0
pending: 2
skipped: 0
blocked: 0

## Gaps
