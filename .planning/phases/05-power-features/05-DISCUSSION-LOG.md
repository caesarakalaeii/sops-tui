# Phase 5: Power Features — Discussion Log

**Date:** 2026-04-15
**Areas discussed:** 4

## Area 1: Recipient Management Flow

**Q: How should the add-recipient interaction work?**
Options: Modal form (Recommended) | Inline input | You decide
**Selected:** Modal form (Recommended)

**Q: How should remove-recipient work when a file has multiple recipients?**
Options: Select from list | Remove current (cursor-based) | You decide
**Selected:** Select from list

**Q: After adding/removing a recipient, how should confirmation work?**
Options: Diff-style confirmation (Recommended) | Simple yes/no prompt | You decide
**Selected:** Diff-style confirmation (Recommended)

**Q: Should the TUI validate age public keys before attempting re-key?**
Options: Validate format (Recommended) | Let sops validate | You decide
**Selected:** Validate format (Recommended)

## Area 2: Bulk Re-Key UX

**Q: How should users select multiple files for bulk re-key?**
Options: Toggle selection in file list (Recommended) | Select all matching a pattern | Both toggle + pattern
**Selected:** Toggle selection in file list (Recommended)

**Q: How should confirmation work for bulk operations across multiple files?**
Options: Per-file confirmation (Recommended) | Batch confirmation with summary | You decide
**Selected:** Per-file confirmation (Recommended)

**Q: How should progress be shown during a bulk re-key?**
Options: Progress counter with file name | Full progress overlay | You decide
**Selected:** Progress counter with file name

## Area 3: Health Check Criteria

**Q: What should count as a 'weak' secret?**
Options: Length + entropy (Recommended) | Format-aware checks | Both length + format | You decide
**Selected:** Both length + format

**Q: How should duplicate detection work across files?**
Options: Exact value match (Recommended) | Key name match only | Both value + key name
**Selected:** Exact value match (Recommended)

**Q: What makes a secret 'stale'?**
Options: Git last-modified age (Recommended) | SOPS lastmodified field | You decide
**Selected:** Git last-modified age (Recommended)

## Area 4: Health Check Triggering

**Q: How should users trigger a health check?**
Options: On-demand keybinding (Recommended) | Automatic on startup | Both on-demand + badges
**Selected:** On-demand keybinding (Recommended)

**Q: How should health check results be displayed?**
Options: Dedicated health overlay (Recommended) | Inline annotations | Both overlay + badges
**Selected:** Dedicated health overlay (Recommended)

---
*Discussion log for audit trail. Not consumed by downstream agents.*
