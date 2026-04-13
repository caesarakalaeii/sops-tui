# Domain Pitfalls

**Domain:** Go TUI + SOPS secret management tool
**Researched:** 2026-04-13

---

## Critical Pitfalls

Mistakes that cause rewrites, security incidents, or data loss.

---

### Pitfall 1: Decrypted Secret Values Leak Into Terminal Scrollback

**What goes wrong:** When a TUI renders decrypted values by printing them to the normal screen buffer (not the alternate screen), or when the app exits without clearing the alternate screen first, decrypted values can persist in the terminal's scrollback buffer. Terminal emulators like gnome-terminal, xfce4-terminal, and Terminator write scrollback to disk — even over SSH sessions. A user who views a secret once may have that value readable on disk indefinitely.

**Why it happens:** Developers assume `fmt.Println` cleared by `tea.ClearScreen` is gone. It is not. Scrollback survives process exit. The alternate screen buffer (`WithAltScreen()`) *does* prevent scrollback leakage — content rendered there disappears on exit — but only if the app never breaks out to the normal screen buffer mid-session.

**Consequences:** Decrypted secrets persist on disk. Anyone with read access to the terminal profile directory (or a co-user of the same workstation) can recover them.

**Prevention:**
- Always use `WithAltScreen()` program option (not `EnterAltScreen()` in Init, which is async and may race)
- Never write decrypted values to `os.Stdout` directly; route everything through Bubble Tea's renderer
- On graceful exit, render a blank frame before quitting so the last rendered state is empty

**Detection:** Run the TUI, reveal a secret, exit, then scroll up in the terminal. If you can see the value, the leak is present.

**Phase:** Foundation/core rendering setup. Get this right before building any secret display feature.

---

### Pitfall 2: Blocking I/O in Bubble Tea's Update() Function

**What goes wrong:** Calling `sops` subprocess, reading files, or any I/O directly inside the `Update()` function freezes the entire TUI. Because Bubble Tea's event loop is single-threaded, any blocking call inside `Update()` blocks keyboard input, rendering, and resize handling simultaneously. Users see a hung, unresponsive terminal.

**Why it happens:** The Elm architecture is unfamiliar. Developers coming from imperative code reach for `exec.Command(...).Output()` in-line where it is most convenient — inside the message handler.

**Consequences:** The TUI becomes unresponsive for the entire duration of any SOPS operation. For large files or slow key derivation, this can be multiple seconds.

**Prevention:**
- All SOPS subprocess calls must be wrapped in `tea.Cmd` functions that return a message
- Pattern: `return model, tea.Cmd(func() tea.Msg { result, err := runSops(...); return sopsResultMsg{result, err} })`
- Show a spinner component while the command runs; update it with `tea.Tick`
- The Bubble Tea source code notes that commands cannot be cancelled once started — design SOPS calls to be short-lived or implement cancellation at the OS process level via `exec.CommandContext`

**Detection:** Add a slow operation to `Update()` in development; confirm the TUI freezes. Run the race detector: `go test -race`.

**Phase:** Architecture decision before any SOPS integration. Establish the async command pattern as a project-wide convention in Phase 1.

---

### Pitfall 3: Subprocess Argument Exposure via Process List

**What goes wrong:** When passing secret values as command-line arguments to a subprocess (e.g., `sops --set '["key"] value'`), those arguments are visible in `/proc/<pid>/cmdline` and via `ps aux` to any user on the same system for the lifetime of the process.

**Why it happens:** The sops CLI accepts values inline as arguments for some operations. Developers use this because it is the documented interface without considering that argv is a global, world-readable structure on Linux.

**Consequences:** Any other user on the system, any monitoring agent, or any script polling `ps` can read the secret value.

**Prevention:**
- Never pass secret values as command-line arguments to subprocesses
- Use SOPS's stdin-based interfaces or file-based temp file approach (with `os.CreateTemp`, proper permissions, deferred deletion)
- For editing: use `sops --edit` with the value written to a temp file (mode 0600), not via argument
- If a temp file is used, zero the contents before deletion: write `\x00` * len over the file before `os.Remove`

**Detection:** During a secret-write operation, run `ps aux | grep sops` from another terminal and check if the secret appears in the argument list.

**Phase:** SOPS subprocess integration phase. Must be addressed before any write/edit operation is implemented.

---

### Pitfall 4: Clipboard Persistence After Auto-Clear Failure

**What goes wrong:** The clipboard auto-clear timer is implemented as a goroutine that sleeps and then calls the OS clipboard tool. If the TUI exits before the timer fires, the goroutine is killed and the secret stays in the clipboard indefinitely.

**Why it happens:** Goroutines are cheap to create and fire-and-forget is idiomatic. Developers implement "copy and clear after 30 seconds" as a goroutine without wiring it to the application's lifecycle.

**Consequences:** Secrets persist in the clipboard across application exit. Any application with clipboard access can read them. Clipboard managers (which often write history to disk) may persist the value permanently.

**Prevention:**
- Track all pending clipboard-clear goroutines in the model state
- On `tea.Quit`, synchronously clear the clipboard before exiting
- Register OS signal handlers (SIGINT, SIGTERM) that also clear the clipboard
- Use a `context.Context` passed to the clear goroutine so it responds to cancellation
- Warn the user if the OS clipboard tool (`xclip`, `xsel`, `pbcopy`, `wl-copy`) is not available rather than silently skipping the clear

**Detection:** Copy a secret, kill the process with Ctrl-C rather than the normal quit keybinding, then inspect clipboard contents.

**Phase:** Clipboard feature implementation. Auto-clear must be part of the initial clipboard implementation, not a follow-on.

---

### Pitfall 5: SOPS .sops.yaml Path Resolution Relative to CWD, Not File Location

**What goes wrong:** SOPS resolves `.sops.yaml` from the current working directory (CWD) of the *process*, not from the directory of the file being encrypted/decrypted. If the TUI is launched from a directory different from the repository root, SOPS may fail to find the config or match the wrong creation rule.

**Why it happens:** Documentation is ambiguous. Developers assume `.sops.yaml` lookup follows the file being operated on (like `.gitignore` traversal). It does not.

**Consequences:** Files that should decrypt fail silently or use the wrong key set. The TUI appears to work correctly when launched from the repo root during development but fails in production for users who launch from a different directory.

**Prevention:**
- Detect the git repository root on startup (via `git rev-parse --show-toplevel` or by walking up the directory tree looking for `.git/`)
- `os.Chdir()` to the repository root before spawning any `sops` subprocess, or pass `--config /path/to/.sops.yaml` explicitly
- Display the resolved `.sops.yaml` path prominently in the UI so users can verify the correct config is active
- Confirm `.sops.yaml` exists at startup and show a clear error if it does not

**Detection:** Launch the TUI from a subdirectory and attempt a decrypt. Confirm SOPS still resolves the correct config.

**Phase:** Initial file discovery and SOPS subprocess integration.

---

### Pitfall 6: SOPS MAC Invalidation After Manual File Edits or Merge Conflicts

**What goes wrong:** SOPS computes a Message Authentication Code (MAC) over all keys *and their ordering* in the encrypted file. Git merge conflicts, manual key additions/removals without going through SOPS, or tools that reformat YAML (sort keys, reorder fields) will break MAC validation. The file becomes unrecryptable with the error "cipher: message authentication failed" or similar.

**Why it happens:** The MAC check is invisible during normal use. It only manifests when the TUI tries to operate on a file that has been touched outside of SOPS, which is common in real repositories with active development.

**Consequences:** The encrypted file is permanently unrecoverable unless the user has the plaintext elsewhere. Attempting `sops -d` on such a file returns a MAC mismatch error and decryption fails.

**Prevention:**
- Before any operation on a file, run a lightweight validity check (attempt a dry-run decrypt and check for MAC errors)
- Detect and clearly label MAC-invalid files as "corrupted" in the file browser rather than showing them as normal encrypted files
- Expose a recovery path: for MAC errors, offer to run with `--ignore-mac` and show a warning that integrity is unverified
- Document that `.sops.yaml` `mac_only_encrypted: true` exists as a per-file option for repos where unencrypted keys change frequently

**Detection:** Manually reorder two keys in an encrypted YAML file and attempt to decrypt it.

**Phase:** File parsing and validation phase. Build the file health check before the edit/decrypt features.

---

## Moderate Pitfalls

---

### Pitfall 7: Treating Encrypted Values as Always Being Strings

**What goes wrong:** SOPS preserves the YAML type of leaf values — booleans, integers, floats, and null values are encrypted differently than strings. A naively written parser that reads all encrypted values as strings will produce wrong types on display (e.g., showing `true` as `"true"`) and will cause re-encryption failures when the value is written back with the wrong type.

**Prevention:**
- Use a YAML library that preserves node types (e.g., `gopkg.in/yaml.v3` with `yaml.Node` rather than `interface{}` unmarshalling)
- Round-trip test: decrypt → re-encrypt → compare byte-for-byte with original. Any type coercion will show up here.

**Phase:** YAML parsing foundation. Must use type-preserving YAML before building any edit feature.

---

### Pitfall 8: Goroutine Leaks from Long-Running SOPS Commands

**What goes wrong:** Bubble Tea launches `tea.Cmd` functions as goroutines. The framework itself notes that "it's not possible to cancel them so we'll have to leak the goroutine until Cmd returns." A slow SOPS call (e.g., key derivation hanging because the age key is missing) will hold a goroutine open indefinitely.

**Prevention:**
- Always use `exec.CommandContext` with a timeout (e.g., 30 seconds) for SOPS subprocess calls
- On timeout, send a `sopsTimeoutMsg` to Update and display a user-visible error
- Verify with `runtime.NumGoroutine()` in tests that goroutine count returns to baseline after operations

**Phase:** SOPS subprocess integration. Build with context and timeout from the start.

---

### Pitfall 9: age Key File Permission and Path Portability

**What goes wrong:** The age key file path differs by platform and by user configuration. On macOS the default is `~/Library/Application Support/sops/age/keys.txt`; on Linux it is `~/.config/sops/age/keys.txt`. Additionally, SOPS will fail silently or with a confusing error if the key file exists but has world-readable permissions (600 is required; 644 or 755 triggers errors in some versions).

**Prevention:**
- Read the `SOPS_AGE_KEY_FILE` environment variable first; fall back to XDG-aware platform defaults
- On startup, verify the key file exists and has mode 0600. If not, show a clear, actionable error (not a raw SOPS error message)
- Never log the path to the key file in debug output without confirming it does not contain key material inline

**Detection:** Test startup on macOS and Linux without the environment variable set. Test with a 0644 key file.

**Phase:** Startup/key detection phase, before any decrypt operation.

---

### Pitfall 10: Terminal Resize Rendering Artifacts and Unhandled WindowSizeMsg

**What goes wrong:** Bubble Tea sends `WindowSizeMsg` on startup and on every `SIGWINCH`. If the model does not propagate this message to all child components (list, viewport, table), components retain their initial dimensions, causing clipped content, misaligned borders, or rendering artifacts after resize.

**Prevention:**
- Write a `WindowSizeMsg` handler at the root model that propagates to every subcomponent
- Test resize by wrapping the running TUI in a tmux pane and resizing it
- Note: Windows does not emit SIGWINCH; the TUI must handle initial sizing correctly since no resize event will follow

**Detection:** Resize the terminal window while the TUI is running. Check for visual corruption.

**Phase:** TUI layout/component architecture. Handle resize in the first rendered component.

---

### Pitfall 11: Destructive Operations Without Confirmation (Re-encryption, Key Rotation, Recipient Removal)

**What goes wrong:** Key rotation and recipient removal permanently alter the encrypted file. If the subprocess call fails partway through (power loss, process kill, SOPS error), the file may be written in a partially re-encrypted state. Additionally, bulk re-key across many files has no undo.

**Prevention:**
- Require explicit confirmation dialogs for any operation that modifies an encrypted file
- Before any bulk operation, write a dry-run summary screen showing exactly what will change
- For rotation and re-key: create a `.bak` copy before the subprocess runs; delete it only on verified success
- Prefer `sops updatekeys` (atomic re-key) over manual key manipulation
- Detect uncommitted git changes before any bulk operation and warn the user

**Phase:** Edit/rotation features. Never implement a destructive operation without a confirmation step.

---

### Pitfall 12: .sops.yaml Only Supports `.sops.yaml`, Not `.sops.yml`

**What goes wrong:** SOPS's automatic config discovery is case-sensitive and extension-specific. Files named `.sops.yml` (common mistake), `sops.yaml`, or `.SOPS.yaml` are not auto-discovered. The TUI's file discovery will silently find no configuration.

**Prevention:**
- When scanning for the config file, explicitly check for `.sops.yaml` only (document this in the UI)
- If the config is not found, check for common misspellings (`.sops.yml`, `sops.yaml`) and warn the user by name

**Phase:** Initial config discovery feature.

---

## Minor Pitfalls

---

### Pitfall 13: Decrypted Values in Log Files or Error Messages

**What goes wrong:** Error handling that includes raw SOPS stdout/stderr in log messages can inadvertently capture decrypted values. SOPS error output sometimes includes the partially decrypted tree for debugging.

**Prevention:**
- Never log SOPS stdout; only log exit codes and stderr after scrubbing for potential value patterns
- Structured logging: log event types and file names, never values

**Phase:** Error handling implementation throughout.

---

### Pitfall 14: lipgloss AdaptiveColor Hanging View()

**What goes wrong:** A confirmed bug (June 2024, bubbletea issue #1036) causes `View()` to hang for seconds when `lipgloss.AdaptiveColor` is used and the terminal background detection call blocks. In terminals with slow or unavailable `OSC 10/11` response, this manifests as a frozen render.

**Prevention:**
- Prefer explicit `lipgloss.Color()` values over `lipgloss.AdaptiveColor()` in production code
- If adaptive colors are required, set a timeout on the background detection call
- Track bubbletea releases; the issue may be resolved in later versions

**Phase:** Theming/styling phase. Verify this issue's status against the version in use.

---

### Pitfall 15: SOPS-Encrypted Files That Are Only Partially Encrypted (encrypted_regex / encrypted_suffix)

**What goes wrong:** SOPS supports `encrypted_regex` and `encrypted_suffix` rules in `.sops.yaml` that cause only matching keys to be encrypted. The rest of the file is plaintext. A TUI that assumes all leaf values are encrypted will fail to distinguish between ciphertext and plaintext values, incorrectly marking plaintext values as needing decryption or attempting to re-encrypt already-plaintext fields.

**Prevention:**
- Parse the `sops.encrypted_regex` / `sops.encrypted_suffix` metadata from the file's own `sops` key before rendering
- Display per-value encryption state (encrypted / plaintext) in the key browser
- Do not attempt to decrypt values that are not marked with the `#sops:enc` comment

**Phase:** File parsing and display, before any value-level UI is built.

---

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| TUI foundation setup | Alt-screen not enabled; secrets visible in scrollback | Use `WithAltScreen()` before first render |
| SOPS subprocess integration | Blocking I/O in `Update()` | Establish `tea.Cmd` async pattern as the only pattern for SOPS calls |
| File discovery | CWD-relative `.sops.yaml` resolution | Detect git root on startup; `chdir` or pass `--config` explicitly |
| Value display | Type coercion on YAML round-trip | Use `yaml.v3` Node API; write round-trip tests |
| Clipboard feature | Secret remains in clipboard after exit | Synchronous clipboard clear on all exit paths including signals |
| Edit/write operations | Secret value in process argv | Use temp files (mode 0600) with deferred zeroing, not inline args |
| Key rotation / re-key | Partial write leaves file unrecoverable | `.bak` copy before operation; verify before deleting backup |
| Startup / key detection | Age key not found; opaque SOPS error surfaced raw | Explicit key file check with actionable error messages |
| Terminal resize | Subcomponents not updated on `WindowSizeMsg` | Root-level resize propagation to all components |
| MAC validation | Corrupted file shown as normal | Pre-operation validity check; label corrupted files visibly |

---

## Sources

- SOPS MAC validation issue — https://github.com/getsops/sops/issues/972 (MAC only over encrypted values discussion)
- SOPS MAC mismatch in FluxCD — https://shivering-isles.com/til/2024/01/flux-sops-ignored-mac
- SOPS path_regex CWD behavior — https://github.com/getsops/sops/issues/465
- SOPS .sops.yaml recursive lookup — https://github.com/getsops/sops/discussions/955
- SOPS key rotation GCP failure — https://github.com/getsops/sops/issues/1764
- Bubble Tea blocking I/O pattern — https://github.com/charmbracelet/bubbletea (official docs)
- Bubble Tea goroutine cancellation limitation — https://github.com/charmbracelet/bubbletea/blob/main/tea.go
- Bubble Tea alt-screen in Init() misuse — https://pkg.go.dev/github.com/charmbracelet/bubbletea
- Bubble Tea resize artifacts (bubble-table) — https://github.com/Evertras/bubble-table/issues/121
- Bubble Tea lipgloss AdaptiveColor hang — https://github.com/charmbracelet/bubbletea/issues/1036
- Terminal scrollback written to disk — https://randomdrake.com/2012/03/07/terminal-scrollbacks-write-data-to-local-disk-even-over-ssh/
- Go subprocess command injection — https://semgrep.dev/docs/cheat-sheets/go-command-injection
- Go path security in subprocess — https://go.dev/blog/path-security
- Go runtime/secret package (Go 1.26 experimental) — https://pkg.go.dev/runtime/secret
- Go memguard library — https://github.com/awnumar/memguard
- age key file platform path issue — https://github.com/getsops/sops/issues/983
- SOPS .sops.yaml file naming — https://getsops.io/docs/
