---
status: partial
phase: 11-regression-perf-gates
source: [11-VERIFICATION.md, 11-02-PLAN.md, screenshots/CHECKPOINT-PENDING.md]
started: 2026-05-04T15:55:00.000Z
updated: 2026-05-04T15:55:00.000Z
---

## Current Test

[awaiting human testing — Linux 4-combo manual sweep + per-combo SC4 observation]

## Tests

### 1. Alacritty — chrome render + alt-screen exit cleanup

expected:
- Launch `/tmp/sops-tui-v1.1-rc` (or `go build -o /tmp/sops-tui-v1.1-rc ./cmd/sops-tui`) in Alacritty
- Resize 80×24 ↔ 200×60 — chrome adapts cleanly, no flicker
- Alt-screen enters cleanly (no residual content from prior shell)
- `q` press → shell prompt area shows no chrome residue, cursor at expected position
- `ctrl-c` → shell returns with clipboard cleared
- Capture PNG at 200×60 in stateFileList → commit as `.planning/phases/11-regression-perf-gates/screenshots/alacritty.png`

result: [pending]

### 2. Ghostty — chrome render + alt-screen exit cleanup

expected:
- Same as #1, in Ghostty
- Capture PNG at 200×60 in stateFileList → commit as `.planning/phases/11-regression-perf-gates/screenshots/ghostty.png`

result: [pending]

### 3. tmux nested in Alacritty — double-alt-screen interaction (Pitfall 10)

expected:
- Launch `tmux` inside Alacritty, then `/tmp/sops-tui-v1.1-rc` inside tmux
- Verify chrome renders correctly (double alt-screen interaction works)
- Resize tmux pane between 80×24 and 200×60 — no flicker
- `q` press inside sops-tui returns to tmux prompt with no chrome residue in tmux scrollback
- `ctrl-c` returns to tmux prompt with clipboard cleared
- Capture PNG at 200×60 in stateFileList → commit as `.planning/phases/11-regression-perf-gates/screenshots/tmux-nested.png`

result: [pending]

### 4. VSCode integrated terminal — xterm.js historical 1-row-offset path

expected:
- Launch `/tmp/sops-tui-v1.1-rc` in VSCode integrated terminal (xterm.js renderer)
- Verify chrome enters cleanly without 1-row-offset (Pitfall 10 §1 historical issue)
- Resize 80×24 ↔ 200×60 — chrome adapts cleanly
- `q` press returns to VSCode terminal prompt with no chrome residue
- `ctrl-c` returns with clipboard cleared
- Capture PNG at 200×60 in stateFileList → commit as `.planning/phases/11-regression-perf-gates/screenshots/vscode-integrated.png`

result: [pending]

## Summary

total: 4
passed: 0
issues: 0
pending: 4
skipped: 0
blocked: 0

## Gaps

(none yet — pending sweep)

## Resume Path

After all 4 PNGs are committed and per-combo observations recorded:

1. Open `.planning/phases/11-regression-perf-gates/11-VERIFICATION.md`
2. Update SC3 evidence rows from `⏳ Pending capture` to `✓ Verified` with per-combo notes
3. Update SC4 evidence (visual observation) from `⏳ HUMAN` to `✓ PASSED` per-combo
4. Update frontmatter: `status: human_needed` → `status: passed`; `must_haves_verified: 9/10` → `10/10`
5. Mark UI-21 `[x]` in `.planning/REQUIREMENTS.md` with Phase 11 evidence pointer
6. Run `/gsd-verify-work 11` to formally close, OR amend this UAT in-place via `/gsd-audit-uat`

If any combo surfaces an issue (1-row offset, chrome residue, flicker on resize): file as a v1.1.x bug per `.github/ISSUE_TEMPLATE/terminal-bug.yml`; record reference in this UAT's Gaps section; do NOT block v1.1 release on it (per CONTEXT.md `<deferred>` v1.1.x patch path).
