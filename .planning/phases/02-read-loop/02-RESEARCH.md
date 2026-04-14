# Phase 2: Read Loop - Research

**Researched:** 2026-04-14
**Domain:** SOPS file format parsing, goccy/go-yaml tree walking, sahilm/fuzzy integration, Bubbletea v2 state machine extension
**Confidence:** HIGH

---

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**D-01:** Use both approaches combined — parse `.sops.yaml` `creation_rules[].path_regex` entries as primary discovery, then scan matched files for the `sops:` metadata key to confirm actual encryption status.

**D-02:** Files matching path_regex but lacking the `sops:` marker appear in the file list with a dim `[unencrypted]` badge. Users see the full picture of what SOPS would manage.

**D-03:** Drilling into an `[unencrypted]` file opens the detail view normally with a banner: "Not yet encrypted — matches .sops.yaml rules". Values are shown in plaintext since there's nothing to mask.

**D-04:** Encrypted leaf values display with type hints from the SOPS envelope: `*** (str)`, `*** (int)`, `*** (bool)`. Users know what kind of value is stored without seeing content.

**D-05:** The `sops:` metadata key is hidden from the YAML tree entirely. Its information is surfaced through the dedicated metadata panel (DEC-04) instead.

**D-06:** Non-encrypted values in SOPS files (via `encrypted_regex`/`unencrypted_regex`) display their actual plaintext value with a subtle `[plain]` badge, making it obvious which keys SOPS left unencrypted.

**D-07:** SOPS metadata (version, lastmodified, recipients, MAC status) is accessed via `i` keypress. Opens a full-screen overlay panel (same pattern as help `?` overlay from Phase 1).

**D-08:** The metadata panel is accessible from both the file list view (shows metadata for highlighted file) and the detail view (shows metadata for current file). Consistent `i` keybinding in both contexts.

**D-09:** Metadata panel renders as a full-screen overlay using the existing `sessionState` pattern: `stateMetadata` with `prevState` for return. Esc or `i` closes it.

**D-10:** Pressing `/` activates an inline filter input at the bottom of the current view. The list filters in real time as the user types. Esc clears the search and restores the full list. Enter selects the highlighted item.

**D-11:** Search is context-aware: in the file list, `/` filters file names. In the detail view, `/` filters key paths within the current file. Each view searches its own domain.

**D-12:** Fuzzy match highlighting uses accent color (`ColorAccent`) on matched characters, leveraging `sahilm/fuzzy` matched character positions. Consistent with the existing design system.

### Claude's Discretion

- `.sops.yaml` parsing edge cases (multiple creation rules, nested configs)
- File tree walking performance strategy (eager vs lazy loading)
- Metadata panel layout and formatting details
- Search input position (top vs bottom of view)
- Empty search results messaging
- Detail view key path flattening for search (e.g., `database.password` vs just `password`)

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope.
</user_constraints>

---

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| NAV-01 | User can browse all SOPS-encrypted files discovered via `.sops.yaml` path rules | D-01/D-02: parse `.sops.yaml` creation_rules[].path_regex + sops: marker scan; SopsDiscoverer service feeds FileListModel |
| NAV-02 | User can view key names from encrypted files without decrypting | D-04/D-05/D-06: goccy/go-yaml MapSlice walk extracts keys; ENC[...] values masked; sops: block hidden |
| NAV-04 | User can fuzzy search files and keys with `/` (k9s-style) | D-10/D-11/D-12: SearchModel as inline mode on existing states; sahilm/fuzzy.Find() with MatchedIndexes for highlight |
| DEC-03 | Secret values are masked by default, revealed on keypress | D-04: `*** (str)` display; type hint parsed from ENC[...,type:X] suffix; TreeNode.Encrypted bool + TypeHint field |
| DEC-04 | User can view SOPS metadata (version, lastmodified, recipients, MAC status) without decrypting | D-07/D-08/D-09: MetadataModel overlay mirrors HelpModel; stateMetadata + prevState; Metadata struct fields from sops: block |
</phase_requirements>

---

## Summary

Phase 2 builds three independent concerns: (1) file discovery — parsing `.sops.yaml` and walking the filesystem, (2) encrypted YAML parsing — extracting key structure and type hints without decryption, and (3) interactive filtering — wiring `sahilm/fuzzy` into both the file list and YAML tree. All three feed into the existing Phase 1 models (`FileListModel`, `DetailModel`) with minimal structural change to `AppModel`.

The SOPS encrypted file format is well-understood. Every encrypted leaf value is a string of the form `ENC[AES256_GCM,data:...,iv:...,tag:...,type:str]` where the `type:` suffix carries the original Go type hint. The `sops:` block is a top-level YAML key that holds metadata and must be excluded from the tree display (D-05). Files that match `.sops.yaml` `path_regex` but lack a `sops:` top-level key are treated as unencrypted (D-02).

`goccy/go-yaml` is mandated by CLAUDE.md but is not yet in `go.mod` — it must be added as a direct dependency in Wave 0. `sahilm/fuzzy` v0.1.1 is already present as an indirect dep (via bubbles). The `bubbles/v2/textinput` component handles the search bar input. Search is implemented as a mode flag on `FileListModel` and `DetailModel`, not a new `sessionState`.

**Primary recommendation:** Implement `SopsDiscoverer` (file-system service) and `YamlParser` (YAML tree builder) as pure Go packages in `internal/sops/` and `internal/parser/` respectively — no Bubbletea dependency. Wire discovery in `NewAppModel()` and YAML parsing in the `Enter/l` handler of `AppModel.Update()`.

---

## Project Constraints (from CLAUDE.md)

| Directive | Impact on Phase 2 |
|-----------|-------------------|
| Use `goccy/go-yaml` not `gopkg.in/yaml.v3` | Must add `github.com/goccy/go-yaml` as direct dep; `go.mod` currently has only `gopkg.in/yaml.v3` as indirect |
| Never use `type any` | All parser structs must use concrete types: `string`, `[]ageKey`, etc. |
| Never use `lipgloss.AdaptiveColor` | All new styles in `styles.go` use explicit hex via existing `Color*` constants |
| `View()` returns `tea.View` struct | `MetadataModel.View()` returns `string`; only `AppModel.View()` returns `tea.View` |
| Use `tea.KeyPressMsg`, `msg.Code`, `msg.Text`, `msg.Mod` | All new key handlers in `FileListModel`, `DetailModel`, `AppModel` use these |
| SOPS integration via subprocess | Phase 2 is read-only from YAML — no subprocess calls needed (metadata extraction is pure parsing) |
| `CGO_ENABLED=0` | `goccy/go-yaml` is pure Go — no CGo needed |

---

## Standard Stack

### Core (already in go.mod)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| charm.land/bubbletea/v2 | v2.0.4 | TUI framework | Locked in Phase 1 |
| charm.land/bubbles/v2 | v2.1.0 | textinput for search bar | Already present; v2.1.0 verified [VERIFIED: go.mod] |
| charm.land/lipgloss/v2 | v2.0.3 | Styling for new badges/styles | Already present |
| github.com/sahilm/fuzzy | v0.1.1 | Fuzzy matching with MatchedIndexes | Already in go.mod as indirect [VERIFIED: go.mod] |

### Must Add
| Library | Version | Purpose | Why |
|---------|---------|---------|-----|
| github.com/goccy/go-yaml | v1.19.2 | YAML parsing preserving order, MapSlice walk | Mandated by CLAUDE.md; yaml.v3 is only indirect [ASSUMED: version — CLAUDE.md states v1.19.2] |

### Installation
```bash
go get github.com/goccy/go-yaml@v1.19.2
```

**Version verification:** `sahilm/fuzzy` v0.1.1 confirmed in `go.mod` and `go.sum`. [VERIFIED: go.mod, go.sum]

---

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── sops/             # File discovery service (new)
│   ├── discoverer.go # SopsDiscoverer: parse .sops.yaml + walk filesystem
│   └── discoverer_test.go
├── parser/           # YAML tree builder (new)
│   ├── yaml.go       # ParseFile: returns []ui.TreeNode + SopsMetadata
│   └── yaml_test.go
├── ui/
│   ├── filelist.go   # FileListModel: add search mode + [unencrypted] badge
│   ├── detail.go     # DetailModel: real TreeNode data + search mode + type hints
│   ├── metadata.go   # MetadataModel: new overlay (mirrors help.go)
│   ├── search.go     # SearchModel: inline filter bar (single-row textinput)
│   ├── styles.go     # Add 5 new named styles
│   └── ...
├── app/
│   └── model.go      # Add stateMetadata; wire discovery + parser; add i/slash handlers
└── keys/
    └── bindings.go   # Add Search, Info keybindings to FileListKeyMap + DetailKeyMap
```

### Pattern 1: SopsDiscoverer — Two-Pass File Discovery (D-01)

**What:** Parse `.sops.yaml` to get `path_regex` rules. Walk the filesystem relative to the `.sops.yaml` directory. Test each candidate file against rules; then quick-scan for `sops:` presence.

**When to use:** Called once at startup in `NewAppModel()` or `Init()` as a `tea.Cmd`.

```go
// Source: derived from SOPS config.go rule matching logic [CITED: github.com/getsops/sops/blob/main/config/config.go]
// internal/sops/discoverer.go

type SopsConfig struct {
    CreationRules []CreationRule `yaml:"creation_rules"`
}

type CreationRule struct {
    PathRegex         string `yaml:"path_regex"`
    EncryptedRegex    string `yaml:"encrypted_regex"`
    UnencryptedRegex  string `yaml:"unencrypted_regex"`
    Age               string `yaml:"age"`
}

// Discover returns all files matching any creation_rule path_regex.
// Each result has IsEncrypted=true if the file contains a top-level "sops:" key.
func Discover(sopsYamlPath string) ([]DiscoveredFile, error)

type DiscoveredFile struct {
    Name        string // relative to sops.yaml dir
    AbsPath     string
    IsEncrypted bool   // false = unencrypted badge
    Rule        CreationRule // the matched rule (for encrypted_regex/unencrypted_regex)
}
```

**Rule matching (first-match-wins, consistent with SOPS source):**
1. Compile `path_regex` as Go `regexp.MustCompile`
2. Test against file path relative to `.sops.yaml` directory
3. Empty `path_regex` catches all files
4. Return first matching rule

**sops: marker detection:** Unmarshal only top-level keys as `map[string]interface{}` and check for `"sops"` key — do NOT fully parse the encrypted content.

### Pattern 2: YamlParser — Encrypted YAML Tree Extraction (D-04, D-05, D-06)

**What:** Parse an encrypted YAML file using `goccy/go-yaml` with `UseOrderedMap()`. Walk the `MapSlice` recursively to produce `[]ui.TreeNode`. Skip the `sops:` top-level key (D-05). Detect encrypted vs plain leaf values.

**Encrypted value detection:**
- Leaf value string starts with `ENC[` → encrypted
- Extract type hint: parse `type:X` from inside `ENC[...,type:X]`
- Type codes: `str` → `"(str)"`, `float` → `"(float)"`, `bool` → `"(bool)"`, `bytes` → `"(bytes)"`, `comment` → `"(comment)"`
- Non-ENC leaf → check if current file has `encrypted_regex`/`unencrypted_regex` to determine `[plain]` badge

```go
// Source: SOPS example.yaml structure [CITED: github.com/getsops/sops/blob/main/example.yaml]
// internal/parser/yaml.go

type ParsedFile struct {
    Nodes    []ui.TreeNode
    Metadata SopsMetadata  // extracted from sops: block
}

type SopsMetadata struct {
    Version       string
    LastModified  string
    MAC           string
    AgeRecipients []string // public keys
    EncryptedRegex    string
    UnencryptedRegex  string
}

func ParseFile(absPath string, rule sops.CreationRule) (ParsedFile, error)
```

**Leaf rendering in TreeNode:**
The existing `TreeNode` struct needs two new fields:
```go
type TreeNode struct {
    Key        string
    Value      string     // plaintext value for [plain] nodes; empty for encrypted
    Children   []TreeNode
    Expanded   bool
    Depth      int
    Encrypted  bool       // true = render as "*** (type)"
    TypeHint   string     // "str", "int", "bool", "float", "bytes" — empty if not encrypted
    IsPlain    bool       // true = SOPS left this value unencrypted; render [plain] badge
}
```

**ENC value regex for type extraction:**
```go
// ENC[AES256_GCM,data:...,iv:...,tag:...,type:str]
// Extract type hint without full regex — use strings.LastIndex(",type:") + suffix
func extractTypeHint(enc string) string {
    const prefix = ",type:"
    idx := strings.LastIndex(enc, prefix)
    if idx < 0 {
        return "str" // safe default
    }
    // value ends at ']'
    hint := enc[idx+len(prefix):]
    hint = strings.TrimSuffix(hint, "]")
    return hint
}
```

**goccy/go-yaml ordered map walk:**
```go
// Source: goccy/go-yaml API [CITED: pkg.go.dev/github.com/goccy/go-yaml]
var root yaml.MapSlice
if err := yaml.UnmarshalWithOptions(data, &root, yaml.UseOrderedMap()); err != nil {
    return ParsedFile{}, err
}
nodes := walkMapSlice(root, 0, rule)
```

### Pattern 3: SearchModel — Inline Filter Bar (D-10, D-11, D-12)

**What:** A mode flag (`searchActive bool`) added to `FileListModel` and `DetailModel`. When active, a `bubbles/v2/textinput.Model` renders as a single row at the bottom of the content area. The view shrinks by 1 row to accommodate it. Filter runs on every keystroke via `sahilm/fuzzy.Find()`.

**Implementation approach (per UI-SPEC D-09 note):** Search is NOT a new `sessionState`. It is a boolean mode on the active model. This avoids complex state machine transitions for what is a transient overlay.

```go
// Added to FileListModel and DetailModel
type FileListModel struct {
    list         list.Model
    keys         keys.FileListKeyMap
    width        int
    height       int
    searchActive bool
    searchInput  textinput.Model
    allItems     []FileItem  // full unfiltered list
    filteredItems []FileItem // current filtered view
}
```

**fuzzy.Find() usage:**
```go
// Source: sahilm/fuzzy API [CITED: pkg.go.dev/github.com/sahilm/fuzzy]
matches := fuzzy.Find(pattern, names) // names []string from allItems
// matches[i].MatchedIndexes = []int of char positions to highlight with SearchMatchStyle
```

**Highlight rendering:** Walk the display string character by character; apply `SearchMatchStyle` (accent foreground) to positions in `MatchedIndexes`, normal foreground to the rest. Join as a single styled string.

### Pattern 4: MetadataModel — Full-Screen Overlay (D-07, D-08, D-09)

**What:** Mirrors `HelpModel` exactly. A `stateMetadata` is added to the `sessionState` enum. `prevState` is used for return. `i` key opens it from `stateFileList` or `stateDetail`. `i` or `Esc` closes it.

```go
// internal/ui/metadata.go — mirrors help.go structure
type MetadataModel struct {
    meta   parser.SopsMetadata
    width  int
    height int
    scroll int  // for j/k scrolling
}

func NewMetadataModel(meta parser.SopsMetadata, width, height int) MetadataModel
func (m *MetadataModel) SetSize(width, height int)
func (m MetadataModel) View() string  // returns string, not tea.View
```

**Content layout (from UI-SPEC Metadata Overlay Content Contract):**
- Title: `SOPS Metadata` in bold
- Label column: 16 cells fixed width, `ColorMuted` foreground
- Value column: remaining width, `ColorFg` foreground
- Fields: version, last modified, MAC, recipients (one per line), enc regex, unc regex
- Footer: `Press i or Esc to close` in `ColorMuted`
- Border: `lipgloss.RoundedBorder()` in `ColorMuted`, background `ColorSurface`

### Anti-Patterns to Avoid

- **Parsing ENC values with full regex:** Use `strings.HasPrefix("ENC[")` for detection and `strings.LastIndex(",type:")` for extraction — regexp is unnecessary overhead for a fixed-format string.
- **Making search a sessionState:** Search is a mode, not a view. Making it a `sessionState` means duplicating all navigation logic. Mode flag is simpler.
- **Fully decoding YAML values during tree walk:** Only test if leaf value starts with `ENC[` — do not attempt to base64-decode or validate cryptographic material.
- **Re-implementing FileListModel filtering via bubbles/list built-in filter:** The built-in filter was explicitly disabled in Phase 1 (`l.SetFilteringEnabled(false)`) for a reason — sops-tui owns the search UX.
- **Walking the `.sops.yaml` dir eagerly for all files:** For large repos, walk lazily or use `filepath.WalkDir` with early return on match. Cache results — do not re-walk on each keystroke.
- **Passing `SopsMetadata` to MetadataModel before a file is selected:** MetadataModel must handle the zero-value case gracefully (show "No file selected").

---

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Fuzzy matching algorithm | Custom Levenshtein or contains() | `sahilm/fuzzy.Find()` | Already in go.mod (indirect); returns scored matches + MatchedIndexes for highlighting; Sublime Text-style ranking |
| Text input with cursor/backspace | Custom rune buffer | `bubbles/v2/textinput.Model` | Already in go.mod; handles cursor, backspace, reset, focus/blur; v2 has real cursor support |
| Ordered YAML unmarshalling | Custom YAML walker | `goccy/go-yaml` with `UseOrderedMap()` | MapSlice preserves YAML key order; required by CLAUDE.md |
| `.sops.yaml` path_regex matching | Custom glob | `regexp.MustCompile(rule.PathRegex)` | SOPS uses Go standard regexp; first-match-wins is identical to SOPS behavior |
| Filesystem walk | Custom os.ReadDir recursion | `filepath.WalkDir` | Standard library; handles symlinks, permission errors gracefully |

**Key insight:** The three non-UI services (discoverer, parser, search) are pure Go functions — no Bubbletea involved. This makes them trivially unit-testable and keeps the model layer thin.

---

## SOPS File Format Reference

### Encrypted File Structure (YAML)
```yaml
# User keys (all leaf values encrypted):
mykey: ENC[AES256_GCM,data:Tr7o=,iv:1=,aad:No=,tag:k=,type:str]
nested:
  password: ENC[AES256_GCM,data:...,type:str]
  port: ENC[AES256_GCM,data:...,type:int]
  enabled: ENC[AES256_GCM,data:...,type:bool]

# Metadata block (HIDDEN from tree per D-05):
sops:
  version: 3.12.2
  lastmodified: "2024-01-15T10:30:00Z"
  mac: ENC[AES256_GCM,data:...,type:str]
  age:
    - recipient: age1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
      enc: |
        -----BEGIN AGE ENCRYPTED FILE-----
        ...
        -----END AGE ENCRYPTED FILE-----
  encrypted_regex: "^(password|secret|key)$"  # optional
  unencrypted_regex: ""  # optional
```
[CITED: github.com/getsops/sops/blob/main/example.yaml]

### Type Hints in Encrypted Values
| SOPS type code | Display as | TreeNode.TypeHint |
|----------------|------------|-------------------|
| `str` | `(str)` | `"str"` |
| `int` | `(int)` | `"int"` |
| `float` | `(float)` | `"float"` |
| `bool` | `(bool)` | `"bool"` |
| `bytes` | `(bytes)` | `"bytes"` |
| `comment` | `(comment)` | `"comment"` |
[CITED: github.com/getsops/sops/blob/main/example.yaml — type annotations confirmed in ENC[...] suffix]

**Note:** SOPS uses `type:int` but the YAML spec stores all integers — verify by checking whether the `int` code appears in practice. CLAUDE.md only mentions str/int/bool. If `int` doesn't appear in real files (SOPS may use `float` for all numbers), the type hint should still display what the SOPS envelope says.

### .sops.yaml Structure
```yaml
creation_rules:
  - path_regex: "secrets/.*\\.yaml$"
    age: age1recipient1
    encrypted_regex: "^(password|secret|token)$"
  - path_regex: "config/.*\\.yaml$"
    age: age1recipient2
    unencrypted_regex: "^(host|port)$"
  - path_regex: ""     # catch-all (empty = matches everything)
    age: age1fallback
```

**Matching rules:**
1. Rules are tested in order — first match wins [CITED: github.com/getsops/sops/blob/main/config/config.go]
2. `path_regex` is matched against file path relative to the `.sops.yaml` directory
3. Empty `path_regex` matches all files
4. Multiple `.sops.yaml` files: SOPS searches upward from the file being operated on — `validator.FindSopsYaml()` already implements this walk

---

## Common Pitfalls

### Pitfall 1: MapSlice vs map[string]interface{} for YAML Walking
**What goes wrong:** Using `yaml.Unmarshal(data, &map[string]interface{}{})` loses key order and makes tree rendering non-deterministic.
**Why it happens:** Go maps have undefined iteration order.
**How to avoid:** Always use `yaml.UnmarshalWithOptions(data, &root, yaml.UseOrderedMap())` where `root` is `yaml.MapSlice`. Walk the `MapSlice` recursively.
**Warning signs:** Keys in detail view appear in a different order on each run.
[CITED: pkg.go.dev/github.com/goccy/go-yaml — UseOrderedMap option]

### Pitfall 2: sops: Block Must Be Excluded Before Tree Walk
**What goes wrong:** The `sops:` top-level key contains deeply nested encrypted metadata. If not excluded before tree walk, it appears as a subtree in the detail view, polluting the display with internal SOPS data.
**Why it happens:** YAML parsers treat `sops:` as a regular key.
**How to avoid:** At the top-level `MapSlice` iteration, skip items where `item.Key == "sops"` (per D-05). Extract the `sops:` value separately into `SopsMetadata` for the metadata overlay.
**Warning signs:** User sees `sops > age > recipient: ENC[...]` entries in the tree.

### Pitfall 3: Regex path_regex Relative to .sops.yaml Directory
**What goes wrong:** Matching `path_regex` against the absolute file path instead of the path relative to the `.sops.yaml` file's directory.
**Why it happens:** `validator.FindSopsYaml()` returns an absolute path; naive callers match against it directly.
**How to avoid:** Compute `relPath, _ := filepath.Rel(filepath.Dir(sopsYamlPath), filePath)` before calling `regexp.MatchString(relPath)`.
**Warning signs:** Files that should appear in the list are missing, or all files match when they shouldn't.
[CITED: github.com/getsops/sops/blob/main/config/config.go — path normalization before matching]

### Pitfall 4: SetSize Not Called on New Components
**What goes wrong:** `MetadataModel` renders at 0x0 (blank screen) on first open if `AppModel.WindowSizeMsg` handler doesn't call `m.metadata.SetSize(...)`.
**Why it happens:** Phase 1 `WindowSizeMsg` handler propagates to all known children — a new component must be added to this propagation.
**How to avoid:** In `AppModel.Update()`, the `tea.WindowSizeMsg` handler must call `SetSize` on `MetadataModel` alongside existing children.
**Warning signs:** Metadata overlay appears blank; no border visible.

### Pitfall 5: Search Mode Height Accounting
**What goes wrong:** When search bar is visible, the content area still uses the full height, causing the search bar to overlap the last content row.
**Why it happens:** `SetSize` passes `mainH` to models, but models don't know if 1 row is consumed by the search bar.
**How to avoid:** When search mode is active, pass `height-1` to the content rendering and render the search bar as the final row before the status bar.
**Warning signs:** Last item in list is hidden behind search bar; search bar overlaps status bar.

### Pitfall 6: textinput.Model Needs Focus() on Activation
**What goes wrong:** Typing characters after pressing `/` doesn't update the filter — characters go to the model's navigation handlers instead.
**Why it happens:** `textinput.Model` only captures key events when focused. Focus must be explicitly set via `m.searchInput.Focus()`.
**How to avoid:** When `/` is pressed, call `m.searchInput.Focus()` and set `m.searchActive = true` atomically. Return the resulting `tea.Cmd` from `Focus()`.
**Warning signs:** `/` activates search bar visually but typing does nothing.

### Pitfall 7: ENC Value Inside Non-String YAML Fields
**What goes wrong:** goccy/go-yaml may unmarshal `ENC[...,type:int]` as the integer `0` or a type error rather than the string `"ENC[...]"` during tree walk.
**Why it happens:** SOPS encrypted files are valid YAML where all encrypted values are YAML strings (quoted). But if the file is malformed or a non-string YAML type is encountered, type assertion `item.Value.(string)` panics.
**How to avoid:** Use a type switch instead of direct type assertion when checking leaf values. If `item.Value` is not a string, render it as `"*** (unknown)"` rather than panicking.
**Warning signs:** Panic on `item.Value.(string)` during tree walk of certain files.

---

## Code Examples

### SOPS Discoverer — Core Logic
```go
// Source: derived from SOPS config.go matching logic [CITED: github.com/getsops/sops/blob/main/config/config.go]
// internal/sops/discoverer.go

func matchRule(filePath, sopsYamlDir string, rules []CreationRule) (CreationRule, bool) {
    relPath, err := filepath.Rel(sopsYamlDir, filePath)
    if err != nil {
        return CreationRule{}, false
    }
    for _, rule := range rules {
        if rule.PathRegex == "" {
            return rule, true // catch-all
        }
        re, err := regexp.Compile(rule.PathRegex)
        if err != nil {
            continue
        }
        if re.MatchString(relPath) {
            return rule, true
        }
    }
    return CreationRule{}, false
}

// hasSOPSMarker does a fast top-level key scan — does NOT unmarshal full content
func hasSOPSMarker(absPath string) bool {
    data, err := os.ReadFile(absPath)
    if err != nil {
        return false
    }
    var top yaml.MapSlice
    if err := yaml.UnmarshalWithOptions(data, &top, yaml.UseOrderedMap()); err != nil {
        return false
    }
    for _, item := range top {
        if k, ok := item.Key.(string); ok && k == "sops" {
            return true
        }
    }
    return false
}
```

### YamlParser — Tree Walk with Type Hints
```go
// Source: goccy/go-yaml MapSlice API [CITED: pkg.go.dev/github.com/goccy/go-yaml]
// internal/parser/yaml.go

func walkMapSlice(ms yaml.MapSlice, depth int, rule sops.CreationRule) []ui.TreeNode {
    nodes := make([]ui.TreeNode, 0, len(ms))
    for _, item := range ms {
        key, ok := item.Key.(string)
        if !ok {
            continue
        }
        // Skip sops: metadata block (D-05)
        if depth == 0 && key == "sops" {
            continue
        }
        node := buildNode(key, item.Value, depth, rule)
        nodes = append(nodes, node)
    }
    return nodes
}

func buildNode(key string, value interface{}, depth int, rule sops.CreationRule) ui.TreeNode {
    switch v := value.(type) {
    case yaml.MapSlice:
        return ui.TreeNode{
            Key:      key,
            Children: walkMapSlice(v, depth+1, rule),
            Expanded: false,
            Depth:    depth,
        }
    case string:
        if strings.HasPrefix(v, "ENC[") {
            hint := extractTypeHint(v)
            return ui.TreeNode{
                Key:       key,
                Depth:     depth,
                Encrypted: true,
                TypeHint:  hint,
            }
        }
        // Plaintext leaf — check rule to determine if [plain] badge applies
        isPlain := isEncryptedFile && isPlainDueToRule(key, rule)
        return ui.TreeNode{
            Key:     key,
            Value:   v,
            Depth:   depth,
            IsPlain: isPlain,
        }
    default:
        // int, bool, float64 — not encrypted (SOPS encrypts everything to strings)
        return ui.TreeNode{
            Key:     key,
            Value:   fmt.Sprintf("%v", v),
            Depth:   depth,
            IsPlain: true,
        }
    }
}
```

### sahilm/fuzzy Integration
```go
// Source: sahilm/fuzzy API [CITED: pkg.go.dev/github.com/sahilm/fuzzy]
// internal/ui/filelist.go (search mode)

func (m FileListModel) applyFilter(pattern string) []FileItem {
    if pattern == "" {
        return m.allItems
    }
    names := make([]string, len(m.allItems))
    for i, item := range m.allItems {
        names[i] = item.Name
    }
    matches := fuzzy.Find(pattern, names)
    result := make([]FileItem, 0, len(matches))
    for _, match := range matches {
        result = append(result, m.allItems[match.Index])
    }
    return result
}

// Highlight rendering using MatchedIndexes
func highlightMatch(s string, matchedIdxs []int, matchStyle lipgloss.Style) string {
    idxSet := make(map[int]bool, len(matchedIdxs))
    for _, idx := range matchedIdxs {
        idxSet[idx] = true
    }
    var sb strings.Builder
    for i, r := range s {
        if idxSet[i] {
            sb.WriteString(matchStyle.Render(string(r)))
        } else {
            sb.WriteRune(r)
        }
    }
    return sb.String()
}
```

### bubbles/v2/textinput for Search Bar
```go
// Source: bubbles/v2/textinput API [CITED: pkg.go.dev/charm.land/bubbles/v2/textinput]
// internal/ui/search.go

type SearchModel struct {
    input  textinput.Model
    width  int
}

func NewSearchModel(width int) SearchModel {
    ti := textinput.New()
    ti.Placeholder = ""
    ti.CharLimit = 100
    // Style: matches UI-SPEC SearchInputStyle
    // Note: set Prompt and PromptStyle separately, not in New()
    return SearchModel{input: ti, width: width}
}

func (m SearchModel) Value() string { return m.input.Value() }

func (m SearchModel) View() string {
    // "/" prompt in accent, input in surface background
    prompt := lipgloss.NewStyle().Foreground(ColorAccent).Render("/")
    bar := lipgloss.NewStyle().
        Background(ColorSurface).
        Foreground(ColorFg).
        Width(m.width - 1). // -1 for "/" prompt
        Render(m.input.View())
    return prompt + bar
}
```

### New Styles for styles.go
```go
// Source: 02-UI-SPEC.md §New Named Styles [CITED: .planning/phases/02-read-loop/02-UI-SPEC.md]
var (
    BadgeUnencrypted = lipgloss.NewStyle().
        Bold(true).
        Foreground(ColorError)

    BadgePlain = lipgloss.NewStyle().
        Foreground(ColorWarning)

    TypeHintStyle = lipgloss.NewStyle().
        Faint(true).
        Foreground(ColorMuted)

    SearchInputStyle = lipgloss.NewStyle().
        Background(ColorSurface).
        Foreground(ColorFg)

    SearchMatchStyle = lipgloss.NewStyle().
        Foreground(ColorAccent)
)
```

### AppModel — Adding stateMetadata
```go
// internal/app/model.go additions

const (
    stateFileList  sessionState = iota
    stateDetail
    stateHelp
    stateMetadata  // NEW: full-screen SOPS metadata overlay
)

type AppModel struct {
    // existing fields ...
    metadata ui.MetadataModel  // NEW
}

// In Update(), KeyPressMsg handler:
if msg.String() == "i" {
    if m.state == stateFileList {
        if item, ok := m.fileList.SelectedItem(); ok {
            meta := loadMetadata(item.AbsPath) // parser.ParseFile returns SopsMetadata
            m.metadata = ui.NewMetadataModel(meta, m.width, m.height)
            m.prevState = m.state
            m.state = stateMetadata
        }
    } else if m.state == stateDetail {
        m.prevState = m.state
        m.state = stateMetadata
    }
    return m, nil
}
if msg.String() == "i" && m.state == stateMetadata {
    m.state = m.prevState
    return m, nil
}
```

---

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `gopkg.in/yaml.v3` Node API | `goccy/go-yaml` MapSlice with `UseOrderedMap()` | CLAUDE.md mandate (2026) | Better spec compliance, preserved key order, cleaner API |
| Plain `***` for all leaf values (Phase 1) | `*** (str)` / `*** (int)` with type hints | Phase 2 (D-04) | Users see value type without decryption |
| bubbles/list built-in filter | Custom fuzzy filter via `sahilm/fuzzy` | Phase 1 decision (filter disabled) | Full control over match rendering and highlight |

**Deprecated/outdated in this project:**
- `detail.go` leaf rendering: `DimText.Render("***")` — Phase 2 replaces with type-hinted display (update `renderRow`)
- `app/model.go` `Enter/l` handler: creates `DetailModel` with `[]ui.TreeNode{}` — Phase 2 replaces with real parsed data

---

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `goccy/go-yaml` v1.19.2 is the correct version to add | Standard Stack | Wrong version may not compile; run `go get github.com/goccy/go-yaml@latest` to verify actual latest |
| A2 | SOPS uses `type:int` (not `type:float`) for integer values in encrypted YAML | SOPS File Format Reference | Type hint shown as `(int)` in UI but actual files use `float` — visual mismatch, minor impact |
| A3 | `filepath.WalkDir` is adequate performance for typical repo sizes (<5000 files) | Architecture Patterns | Very large repos (10k+ files) may have noticeable startup delay; if so, add goroutine + progress indicator (but that is Phase 2+ work) |
| A4 | `goccy/go-yaml` unmarshals all SOPS encrypted leaf values as Go `string` type (not attempts to parse `ENC[...]` as a struct) | Code Examples | If goccy/go-yaml interprets `ENC[...]` as something other than a string, the `HasPrefix("ENC[")` check fails |

---

## Open Questions

1. **Does SOPS use `type:int` or `type:float` for integer YAML values?**
   - What we know: The example.yaml shows `type:str`, `type:float`, `type:bool`. The CLAUDE.md UI-SPEC mentions `*** (int)` in display.
   - What's unclear: Whether SOPS ever emits `type:int` vs always using `type:float` for numeric values.
   - Recommendation: Display whatever `type:X` appears in the ENC string verbatim. Show `(int)` only if the hint is literally `int`. This is safe regardless of SOPS behavior.

2. **Performance of eager file walk at startup for large repos**
   - What we know: `filepath.WalkDir` is synchronous; large repos with many YAML files may cause TUI startup delay.
   - What's unclear: Whether a spinner/loading state is needed or startup latency is acceptable.
   - Recommendation: Run discovery as a `tea.Cmd` (goroutine) and send results as a `FilesDiscoveredMsg`. File list shows spinner until results arrive. This is the idiomatic Bubbletea pattern and avoids blocking `Init()`.

3. **goccy/go-yaml behavior with SOPS MAC field**
   - What we know: The `sops.mac` field value is itself an `ENC[...]` string.
   - What's unclear: Whether to display the MAC as the full ENC string or as `[encrypted]` in the metadata overlay.
   - Recommendation: Display the full MAC hex string after SOPS decrypts it — but since Phase 2 has no decryption, display `[encrypted until decrypted]` for MAC in the metadata overlay. This is a discretionary display decision.

---

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go 1.26.2 | All packages | Yes | 1.26.2 | — |
| `sops` binary | Integration tests (Phase 3+) | Checked at runtime by validator | — | Phase 2 doesn't call sops subprocess |
| `goccy/go-yaml` | YamlParser | Not yet in go.mod | TBD | Cannot use yaml.v3 (CLAUDE.md mandate) |
| `sahilm/fuzzy` | SearchModel | Yes (indirect dep) | v0.1.1 | — |
| `bubbles/v2/textinput` | SearchModel | Yes (part of bubbles v2.1.0) | v2.1.0 | — |

**Missing dependencies with no fallback:**
- `goccy/go-yaml` — must be added via `go get github.com/goccy/go-yaml@v1.19.2` in Wave 0

[VERIFIED: go.mod, go.sum — sahilm/fuzzy v0.1.1 confirmed present; goccy/go-yaml confirmed absent]

---

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + `stretchr/testify` v1.11.1 |
| Config file | none (go test ./...) |
| Quick run command | `go test ./internal/sops/... ./internal/parser/... ./internal/ui/... -run . -count=1` |
| Full suite command | `go test ./... -count=1` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| NAV-01 | Discoverer finds all path_regex matched files | unit | `go test ./internal/sops/... -run TestDiscover` | No — Wave 0 |
| NAV-01 | Discoverer correctly identifies unencrypted files (no sops: marker) | unit | `go test ./internal/sops/... -run TestHasSOPSMarker` | No — Wave 0 |
| NAV-02 | Parser extracts TreeNodes without decryption | unit | `go test ./internal/parser/... -run TestParseFile` | No — Wave 0 |
| NAV-02 | Parser skips sops: block from tree (D-05) | unit | `go test ./internal/parser/... -run TestParseFile_SkipSopsBlock` | No — Wave 0 |
| NAV-04 | fuzzy.Find returns correct MatchedIndexes for highlighting | unit | `go test ./internal/ui/... -run TestHighlightMatch` | No — Wave 0 |
| DEC-03 | Encrypted leaf renders "*** (str)" with TypeHintStyle | unit | `go test ./internal/ui/... -run TestDetailView_TypeHints` | No — Wave 0 |
| DEC-03 | Non-encrypted leaf renders plaintext with [plain] badge | unit | `go test ./internal/ui/... -run TestDetailView_PlainBadge` | No — Wave 0 |
| DEC-04 | MetadataModel renders all SopsMetadata fields | unit | `go test ./internal/ui/... -run TestMetadataModel_View` | No — Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/sops/... ./internal/parser/... ./internal/ui/... -count=1`
- **Per wave merge:** `go test ./... -count=1`
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `internal/sops/discoverer_test.go` — covers NAV-01 (with temp dir + fake .sops.yaml fixtures)
- [ ] `internal/parser/yaml_test.go` — covers NAV-02, DEC-03 (with embedded test YAML strings)
- [ ] `internal/ui/metadata_test.go` — covers DEC-04
- [ ] Test fixture: minimal encrypted YAML string with ENC[...] values (inlined in test file, no temp files needed)

*(Existing test infrastructure in `internal/ui/` and `internal/validator/` covers Phase 1 requirements — no gaps there.)*

---

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | Phase 2 is read-only filesystem access |
| V3 Session Management | No | No sessions |
| V4 Access Control | No | SOPS key access is OS-level (file permissions) |
| V5 Input Validation | Yes | path_regex from .sops.yaml must be validated before `regexp.Compile`; catch malformed regex with error return |
| V6 Cryptography | No | Phase 2 never calls any crypto; ENC[] values are treated as opaque strings |

### Known Threat Patterns for SOPS TUI Read Phase

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Malformed path_regex in .sops.yaml causing `regexp.Compile` panic | Tampering | Use `regexp.Compile` (returns error) not `regexp.MustCompile`; skip malformed rules with log |
| Directory traversal via `.sops.yaml` path_regex matching files outside repo | Tampering | Validate that discovered absolute paths are under or equal to the sops.yaml directory |
| ENC[...] value injection in key names (YAML key injection) | Tampering | Key names from YAML parsing are display-only strings — never executed or evaluated |
| Large YAML files causing unbounded memory during tree walk | DoS | Consider file size limit (e.g., 10MB) before parsing; return error for oversized files |

---

## Sources

### Primary (HIGH confidence)
- `go.mod` / `go.sum` — dependency versions verified directly [VERIFIED: go.mod]
- `internal/ui/*.go`, `internal/app/model.go`, `internal/keys/bindings.go`, `internal/validator/startup.go` — Phase 1 codebase [VERIFIED: codebase read]
- `.planning/phases/02-read-loop/02-CONTEXT.md` — locked decisions [VERIFIED: file read]
- `.planning/phases/02-read-loop/02-UI-SPEC.md` — visual contracts [VERIFIED: file read]
- `pkg.go.dev/github.com/sahilm/fuzzy` — API: Find(), Match.MatchedIndexes [CITED: https://pkg.go.dev/github.com/sahilm/fuzzy]
- `pkg.go.dev/charm.land/bubbles/v2/textinput` — API: New(), Focus(), Value(), Update() [CITED: https://pkg.go.dev/charm.land/bubbles/v2@v2.1.0/textinput]

### Secondary (MEDIUM confidence)
- `github.com/getsops/sops/blob/main/example.yaml` — ENC[AES256_GCM,...,type:str] format confirmed [CITED: https://github.com/getsops/sops/blob/main/example.yaml]
- `github.com/getsops/sops/blob/main/stores/stores.go` — Metadata struct fields confirmed [CITED: https://raw.githubusercontent.com/getsops/sops/main/stores/stores.go]
- `github.com/getsops/sops/blob/main/config/config.go` — creationRule struct, path_regex first-match-wins logic [CITED: https://raw.githubusercontent.com/getsops/sops/main/config/config.go]
- `pkg.go.dev/github.com/goccy/go-yaml` — UseOrderedMap(), MapSlice, UnmarshalWithOptions [CITED: https://pkg.go.dev/github.com/goccy/go-yaml]

### Tertiary (LOW confidence)
- `type:int` vs `type:float` for integer YAML values in SOPS — could not confirm from example.yaml alone [LOW — see Assumptions A2]

---

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all deps verified in go.mod; only goccy version is ASSUMED
- Architecture: HIGH — patterns derived directly from Phase 1 codebase and locked decisions
- SOPS file format: HIGH (MEDIUM for type:int edge case) — confirmed from official example.yaml
- sahilm/fuzzy API: HIGH — confirmed from pkg.go.dev
- Pitfalls: HIGH — derived from direct code reading and API docs

**Research date:** 2026-04-14
**Valid until:** 2026-05-14 (stable ecosystem — Go stdlib, SOPS format is stable)
