# Linux 4-Combo Manual Sweep — PENDING

**Status:** Awaiting developer manual sweep + PNG capture
**Created:** 2026-05-04 (Plan 11-02 Task 3)
**Binary:** `/tmp/sops-tui-v1.1-rc` (built from current HEAD)

## Why This Is Pending

Plan 11-02 Task 3 is a `human-action` checkpoint. It requires the developer to launch sops-tui in 4 real Linux terminals, capture PNG screenshots at 200×60 in stateFileList, and record per-combo observations. This cannot be automated by the executor agent.

Tasks 1, 2, and 4 of Plan 11-02 are complete and committed. The bench gate, regression sanity tests, README "Verified Terminals" matrix, and `.github/ISSUE_TEMPLATE/terminal-bug.yml` are all live. Only the manual sweep + 4 PNGs + per-combo observations remain.

## Required Outputs

Place 4 PNG screenshots at the following paths (drop the `CHECKPOINT-PENDING.md` placeholder file and remove this directory's `.gitkeep` if present):

- `.planning/phases/11-regression-perf-gates/screenshots/alacritty.png`
- `.planning/phases/11-regression-perf-gates/screenshots/ghostty.png`
- `.planning/phases/11-regression-perf-gates/screenshots/tmux-nested.png`
- `.planning/phases/11-regression-perf-gates/screenshots/vscode-integrated.png`

Each PNG should show the chrome at 200×60 in stateFileList (the most chrome-rich state — info panel + menu + ASCII logo + breadcrumb chips + status bar all visible).

## Per-Combo Manual Sweep Checklist

For each terminal combo, run the following checklist using `/tmp/sops-tui-v1.1-rc` (the binary built at plan execution time). Record PASS/FAIL/DEFERRED-as-v1.1.x for each item.

### Combo 1: Alacritty (baseline TrueColor)

```
[ ] alacritty:
    chrome at 200×60 (info panel left, menu middle, logo right): PASS / FAIL
    chrome at 80×24 (mid-tier per Phase 7.1 D-116, no info panel): PASS / FAIL
    resize 80↔200 no flicker: PASS / FAIL
    alt-screen enter clean (no residual shell content): PASS / FAIL
    alt-screen exit clean on q (no chrome residue, cursor at expected position): PASS / FAIL
    SIGINT (ctrl-c) clean (clipboard cleared if previously copied): PASS / FAIL / DEFERRED-as-v1.1.x
    Notes: ...
```

**Mechanism:** Plan 11-01 wired `m.quitting = true` on the Quit branch (model.go:1086-1089) → View() top branch returns blank `tea.View{Content:"", AltScreen:true}` (model.go:1497-1503) → Cursed Renderer's last frame leaves no chrome residue.

**Capture:** `gnome-screenshot -w -f .planning/phases/11-regression-perf-gates/screenshots/alacritty.png` (or your preferred screenshot tool) at the END of the checklist (before pressing q).

### Combo 2: Ghostty

Repeat the per-combo checklist above.

```
[ ] ghostty:
    chrome at 200×60: PASS / FAIL
    chrome at 80×24: PASS / FAIL
    resize 80↔200 no flicker: PASS / FAIL
    alt-screen enter clean: PASS / FAIL
    alt-screen exit clean (q): PASS / FAIL
    SIGINT (ctrl-c) clean: PASS / FAIL / DEFERRED-as-v1.1.x
    Notes: Different rasteriser vs Alacritty; chrome should look identical, font rendering may differ — that's fine.
```

**Capture:** `.planning/phases/11-regression-perf-gates/screenshots/ghostty.png`.

### Combo 3: tmux nested in Alacritty

```
[ ] tmux-nested:
    chrome at 200×60 inside tmux (outer terminal slightly larger): PASS / FAIL
    chrome at 80×24 inside tmux: PASS / FAIL
    resize outer Alacritty propagates clean to inner sops-tui: PASS / FAIL
    alt-screen enter clean inside tmux (double-alt-screen interaction): PASS / FAIL
    alt-screen exit clean (q returns to tmux prompt, not shell directly): PASS / FAIL
    SIGINT (ctrl-c) clean: PASS / FAIL / DEFERRED-as-v1.1.x
    exit tmux clean (Alacritty prompt clean): PASS / FAIL
    Notes: Phase 11 prep; Pitfall 10 §"double-alt-screen" surface.
```

**Capture:** `.planning/phases/11-regression-perf-gates/screenshots/tmux-nested.png`.

### Combo 4: VSCode integrated terminal

```
[ ] vscode-integrated:
    chrome at 200×60 in xterm.js: PASS / FAIL
    chrome at 80×24: PASS / FAIL
    no 1-row offset on alt-screen enter (Pitfall 10 §VSCode historical issue): PASS / FAIL
    resize behavior on terminal pane drag: PASS / FAIL
    alt-screen exit clean on q (xterm.js sometimes leaves residual): PASS / FAIL
    SIGINT (ctrl-c) clean: PASS / FAIL / DEFERRED-as-v1.1.x
    Notes: Per Plan 11-01 D-513, zero-state guard + Cursed Renderer suffice; expect no 1-row offset.
```

**Capture:** `.planning/phases/11-regression-perf-gates/screenshots/vscode-integrated.png`.

## Substitution Policy

If a combo is unavailable on this developer's workstation (e.g. Ghostty not installed):

1. Substitute with another Linux terminal (Konsole, GNOME Terminal, foot — picker's discretion).
2. Update README.md "Verified Terminals" table to match what was actually verified (rename row, keep cell schema).
3. Rename the screenshot file accordingly (e.g. `konsole.png` instead of `ghostty.png`).
4. Document the substitution in the resume signal so `/gsd-verify-work` can adjust 11-VERIFICATION.md SC3 evidence rows.

## Resume Signal

When the sweep is complete:

1. Delete this `CHECKPOINT-PENDING.md` file.
2. Place the 4 PNGs in this directory.
3. Commit with a message including the per-combo observation block (4 combos × ≥6 checklist items each).
4. Re-run `/gsd-execute-phase 11` (or signal the orchestrator) to advance Plan 11-02 Task 4 → SUMMARY.md → state updates.

## What's Already Done (Context for Verifier)

Tasks 1, 2, and 4 of Plan 11-02 are complete:

- **Task 1** (commit `667fe5b`): `internal/app/regression_test.go` with 3 chrome-interaction sanity tests (clipboard auto-clear with chrome, recipient form menu hints, health overlay on narrow width). All 3 pass.
- **Task 2** (commit `5dd0b44`): README.md "Verified Terminals" H2 + 8-row matrix + `.github/ISSUE_TEMPLATE/terminal-bug.yml` with 6 required fields.
- **Task 4** (final test pass): Will be committed via the SUMMARY.md commit. Full repo suite + bench gate green.

The 4 PNGs + observation block are the ONLY remaining artifacts blocking `/gsd-verify-work` from building 11-VERIFICATION.md SC3 evidence rows.

---
*Plan: 11-02 — Regression Sanity Tests + Linux Compat Sweep + README + Issue Template*
*Phase: 11-regression-perf-gates*
