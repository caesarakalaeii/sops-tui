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

## Requirements

- Go 1.22+
- [sops](https://github.com/getsops/sops) installed and configured
- [age](https://github.com/FiloSottile/age) key available at `~/.config/sops/age/keys.txt`

## License

[AGPL-3.0](LICENSE)
