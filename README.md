# sops-tui

A k9s-inspired terminal UI for managing [SOPS](https://github.com/getsops/sops)-encrypted secrets.

Browse, decrypt, edit, and rotate your SOPS secrets — all from the terminal.

## Features

- **Browse** — Tree view of all SOPS-encrypted files in your repo
- **Inspect** — View encrypted keys at a glance without decrypting
- **Decrypt & View** — Reveal secret values on demand
- **Edit** — Modify values with automatic re-encryption
- **Rotate** — Generate new random values for secrets
- **k9s-style UX** — Keyboard-driven, vim-like navigation, command palette

## Status

Early development. Not yet functional.

## Stack

- **Go** + [Bubble Tea](https://github.com/charmbracelet/bubbletea) (Elm architecture TUI framework)
- **SOPS** + **age** encryption backend

## Verified Terminals

The chrome (k9s-style header, info panel, breadcrumb chips) was visually verified during the v1.1 release on the following Linux terminal emulators. Other terminals are expected to work — community-contributed reports welcome.

| Terminal | Version Tested | OS | Status | Notes |
|----------|----------------|----|--------|-------|
| Alacritty | latest | Linux | Verified | Reference TrueColor combo. |
| Ghostty | latest | Linux | Verified | Different rendering pipeline; chrome behaves identically. |
| tmux (nested in Alacritty) | latest | Linux | Verified | Double alt-screen interaction confirmed clean. |
| VSCode integrated terminal | latest | Linux | Verified | xterm.js; no 1-row offset on chrome enter/exit. |
| macOS Terminal | — | macOS | Community-contributed reports welcome | [Open a terminal-bug issue](.github/ISSUE_TEMPLATE/terminal-bug.yml) |
| iTerm2 | — | macOS | Community-contributed reports welcome | [Open a terminal-bug issue](.github/ISSUE_TEMPLATE/terminal-bug.yml) |
| Windows Terminal | — | Windows | Community-contributed reports welcome | [Open a terminal-bug issue](.github/ISSUE_TEMPLATE/terminal-bug.yml) |
| WSL2 (any terminal) | — | Linux/Windows | Community-contributed reports welcome | [Open a terminal-bug issue](.github/ISSUE_TEMPLATE/terminal-bug.yml) |

If you hit a chrome rendering issue on a terminal not yet verified, file a terminal bug using the [terminal-bug template](.github/ISSUE_TEMPLATE/terminal-bug.yml) — include terminal name, version, OS, expected behaviour, observed behaviour, and a screenshot.

## Requirements

- Go 1.22+
- [sops](https://github.com/getsops/sops) installed and configured
- [age](https://github.com/FiloSottile/age) key available at `~/.config/sops/age/keys.txt`

## License

[AGPL-3.0](LICENSE)
