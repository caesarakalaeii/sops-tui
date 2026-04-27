# Phase 7: Chrome Skeleton - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-24
**Phase:** 07-chrome-skeleton
**Areas discussed:** Logo ASCII art, Menu layout & density, Titled-border title format, Hints() scope in Phase 7

---

## Logo ASCII art

| Option | Description | Selected |
|--------|-------------|----------|
| k9s-style block SOPS | Just `SOPS` in 5-row block figlet — tight, iconic, matches k9s LogoSmall most directly | |
| Block SOPS-TUI stacked | 5-row block `SOPS` + `tui` subscript on row 6 — full name readable | ✓ |
| Figlet small `sops-tui` | Compact kebab-case signature fits 6 rows naturally; more "tool-ish" than branded | |
| Claude drafts 3 options | Defer pixel-level call to post-discuss, pick during `/gsd-verify-work` | |

**User's choice:** Block SOPS-TUI stacked (with `tui` subscript on row 6)
**Notes:** Default color = `ColorAccent` (Claude's Discretion); severity coupling deferred to Phase 10 per UI-03.

---

## Menu layout & density

### Column strategy

| Option | Description | Selected |
|--------|-------------|----------|
| Fixed 2 cols | 2 cols × 6 rows = 12 hints max; same layout at all widths | ✓ |
| Fixed 3 cols | 3 cols × 6 rows = 18 hints max; wastes space, can clip narrow | |
| Flex 2→3 at 100 cols | Responsive (2 narrow, 3 wide) — research recommendation | |
| Column count from hint count | Auto-flow, most elastic but hardest to golden-file | |

### Renderer

| Option | Description | Selected |
|--------|-------------|----------|
| lipgloss/v2/table + StyleFunc | StyleFunc(row,col) for mnemonic/description column styling | ✓ |
| Hand-rolled JoinHorizontal | Build columns via JoinHorizontal on pre-rendered cells | |
| Hand-rolled JoinVertical of rows | Row-first mental model, diverges from k9s | |

### Over-capacity policy (>12 bindings)

| Option | Description | Selected |
|--------|-------------|----------|
| `MenuHint.Visible` flag | Sub-model decides which 12 matter; rest stay in `?` overlay | ✓ |
| Priority-ordered, top 12 | Ordering discipline in keymap; take first 12 | |
| Truncate with ellipsis | `… ? for all` suffix; lowest author burden | |
| Overflow to second menu page | Alt+M flips; adds state + keybinding | |

**User's choices:** Fixed 2 cols; lipgloss/v2/table + StyleFunc; MenuHint.Visible flag
**Notes:** Mnemonic column styled ColorAccent; description column styled ColorFg. All 9 sub-models curate their 12.

---

## Titled-border title format

### Format style

| Option | Description | Selected |
|--------|-------------|----------|
| `Files (12)` | Paren count suffix — k9s-standard list-view pattern | ✓ |
| `Files · 12 items` | Dot + noun — readable but wider | |
| `[Files] 12` | Bracketed view name + bare count — closest to k9s resource headers | |
| `Files` only | Minimal; counts go elsewhere (status bar / info panel) | |

### Scope & subject handling

| Option | Description | Selected |
|--------|-------------|----------|
| Counts on list views + subject on contextual views | `Files (12)`, `Detail: prod.yaml`, others name-only | ✓ |
| Counts everywhere countable; no subject | `Files (12)`, `Detail`, others name-only — simpler generator | |
| Subjects on contextual views; no counts | Everything gets subject, no counts — stable titles | |
| `Files (12)` only — minimal scope | Only FileList gets a count | |

**User's choices:** `Files (N)` format; list views with counts + contextual views with subject + static views name-only
**Notes:** Health uses unit-ful `(N findings)` deliberately — variance picked verbatim from the preview.

---

## Hints() scope in Phase 7

| Option | Description | Selected |
|--------|-------------|----------|
| Stub Hints() on all 9 now | Phase 7 adds Hints() to all 9 sub-models; menu is always real | ✓ |
| SUMMARY plan: 4 now, rest in Phase 8 | FileList/Detail/Help/Diff only; fallback default hints otherwise | |
| 4 now + explicit fallback sentinel | Marked-generic row "(view hints not yet wired — press ? for all)" | |
| 4 now, menu hidden on un-wired views | Menu collapses when sub-model lacks Hints() | |

**User's choice:** Stub Hints() on all 9 now (FileList, Detail, Help, Diff, Metadata, Health, History, RecipientList, RecipientForm)
**Notes:** Contradicts SUMMARY's 4-then-5 split. Closes Pitfall 3 (wrong menu on un-wired states) in one pass. Phase 8 scope reduces — loses the Hints() backfill line item. Phase 9's discipline layer (golden matrix per tuple) unchanged.

---

## Post-area follow-up

**Question:** "We've covered the 4 selected areas. Chrome caching and plan split were going to Claude's Discretion. Anything you want to pull into the discussion?"

| Option | Selected |
|--------|----------|
| Ready for CONTEXT.md | ✓ |
| Discuss chrome caching | |
| Discuss plan split | |
| Different gray area entirely | |

**User's choice:** Ready for CONTEXT.md
**Implication:** Chrome caching → pure every-frame with styles-as-package-vars (Pitfall 2 zero-alloc path). Plan split → Claude decides within the 3-plan budget, captured as D-25/D-26 in CONTEXT.md.

## Claude's Discretion

- Exact logo byte-art within the 6×~26 envelope
- `overlayTitle` implementation details (first-line measurement, column-2 insertion, title-wider-than-border truncation)
- `RenderLogo` return shape (combined art+status vs split)
- `RenderMenu` mnemonic subcolumn width math
- Golden file naming convention
- Exact line number where `chromeHeight` lands after stub flip
- Whether `stateFormatMenu` gets a titled border (overlay modal)
- Chrome caching strategy (Pitfall 2): pure every-frame via styles-as-package-vars; cache can be added Phase 10+ without API change
- Plan split across 3-plan budget: primitive-first (Plan 1: hints+logo+menu; Plan 2: chrome+WrapTitled+overlayTitle; Plan 3: integration+Hints on 9+grep-gates+bench gate+goldens)

## Folded Todos

- Phase 7 research pass on `overlayTitle` (from STATE.md Pending Todos) — folded into Plan 2 as a unit-tested deliverable citing `charmbracelet/soft-serve/pkg/ui/components/header`

## Deferred Ideas

None — all architectural details for Phase 7 addressed within scope. Items deferred to later phases (8, 9, 10, 11, v2) are already enumerated in ROADMAP.md and restated in CONTEXT.md `<deferred>` section.

---

## [APPROVED] 2026-04-27 — Phase 7.1 SC5 governance restoration

**Decision:** Path (c) — defer perf to Phase 11 SC2; revert all unauthorized amendments.

**Accepted by:** moersener
**Accepted at:** 2026-04-27

**What was reverted:**
1. `internal/app/chrome_test.go` line 230 — `const budgetNs` reverted from `5_000_000` (5 ms) back to `50_000` (50 µs locked D-24 value).
2. `TestBenchmarkAppView_UnderBudget` now t.Skips with the original 50 µs constant preserved in code form.
3. `.planning/ROADMAP.md` line 192 — SC5 wording reverted to `BenchmarkAppView stays ≤ 50 µs/op at 200×60 with no lipgloss.NewStyle() inside View()` (the locked CONTEXT.md D-24 target).

**What stays:** The `[optional crumbs]` ROADMAP wording is a legitimate amendment (crumbs are conditionally omitted when crumbsHeight=0; unrelated to the perf gap). It remains.

**Phase 11 ownership:** Phase 11 SC2 retains the verbatim original target `BenchmarkAppView stays ≤ 50 µs/op at 200×60 with the full chrome rendered`. D-18 caching fallback (model-level cache keyed on `(state, recipientAction, IsSearchActive, width)`) is the anticipated resolution path; Phase 11 may choose alternative paths (allocation hygiene, lipgloss internal optimisations) provided the 50 µs target is met.

**Cross-references:**
- `.planning/phases/07.1-chrome-gap-closure/07.1-CONTEXT.md` D-101..D-106
- `.planning/phases/07-chrome-skeleton/07-VERIFICATION.md` Gap 1 (BLOCKER)
- `.planning/phases/07-chrome-skeleton/07-RESEARCH.md` §"Open Questions Q4 RESOLVED" (which forbade pre-emptively raising the budget — the original Rule 1 deviation contradicted this)
- `.planning/phases/07-chrome-skeleton/07-03-SUMMARY.md` "Rule 1 deviation" block (historical record retained with a closure pointer added in this same Phase 7.1 plan)

This is the audit-trail entry the original Rule 1 deviation should have produced.
