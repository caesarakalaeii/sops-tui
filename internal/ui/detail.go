// Package ui provides the view components for sops-tui.
// This file implements DetailModel, a Bubble Tea component that renders
// a collapsible YAML tree view for a single SOPS-encrypted file.
//
// Per D-06: YAML tree indentation preserves structure (2 cells per level).
// Per D-07: [+]/[-] indicators show node collapse state.
// Per NAV-03: j/k/g/G/ctrl-d/u navigation; Enter/l expands, h/left collapses.
// Per 01-UI-SPEC.md §YAML Tree Rendering: connectors ├─ └─ │ in muted color.
// Per D-01..D-04: r/R reveal/mask with ClearAllRevealed on Esc (T-03-02).
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
	sops "github.com/caesarakalaeii/sops-tui/internal/sops"
)

// RevealRequestMsg is sent by DetailModel.Update when the user presses r on an
// encrypted, non-revealed leaf. AppModel handles it by dispatching a sops.DecryptKey
// subprocess and returning DecryptKeyMsg.
type RevealRequestMsg struct {
	KeyPath string // dot-joined key path, e.g. "database.password"
}

// RevealAllRequestMsg is sent by DetailModel.Update when the user presses R and
// no values are currently revealed. AppModel handles it by dispatching sops.DecryptFile.
type RevealAllRequestMsg struct {
	FilePath string // absolute path to the SOPS-encrypted file
}

// EditConfirmMsg is sent by DetailModel.Update when the user presses Enter in edit mode.
// AppModel handles it by transitioning to stateDiff with a single-entry diff.
type EditConfirmMsg struct {
	KeyPath  string // dot-joined key path of the edited node
	OldValue string // the original decrypted value before editing
	NewValue string // the new value typed by the user
}

// EditCancelMsg is sent by DetailModel.Update when the user presses Esc in edit mode.
// AppModel handles it as a no-op (no state change required).
type EditCancelMsg struct{}

// EditBlockedMsg is sent by DetailModel.Update when the user presses e on a node
// that cannot be edited. Reason is empty for "masked leaf" (AppModel flashes
// "Reveal first with r"). Reason is non-empty for specific blocks (e.g. array-indexed keys).
type EditBlockedMsg struct {
	Reason string
}

// EditorRequestMsg is sent by DetailModel.Update when the user presses E (EditFile key)
// and at least one node is revealed. AppModel handles it by decrypting the file and
// launching $EDITOR via tea.ExecProcess.
type EditorRequestMsg struct {
	FilePath string // informational — AppModel uses currentFile.AbsPath
}

// RotateFormatMenuMsg is sent by DetailModel.Update when the user presses X on a revealed
// leaf whose format cannot be auto-detected (FormatUnknown). AppModel opens the format
// selection menu.
type RotateFormatMenuMsg struct {
	KeyPath  string
	OldValue string
}

// RotateReadyMsg is sent by DetailModel.Update when the user presses X on a revealed
// leaf with a detectable format. AppModel transitions directly to stateDiff.
type RotateReadyMsg struct {
	KeyPath  string
	OldValue string
	NewValue string
	Format   SecretFormat
}

// RotateErrorMsg is sent by DetailModel.Update when value generation for rotation fails.
type RotateErrorMsg struct {
	Err error
}

// TreeNode represents a single node in the YAML key tree.
// Group nodes (those with Children) can be collapsed or expanded.
// Leaf nodes (no Children) display a masked value ("***" in Phase 1).
type TreeNode struct {
	// Key is the YAML key name displayed in the tree.
	Key string
	// Value is the leaf value. Empty for group nodes. "***" for leaf nodes in Phase 1.
	Value string
	// Children are the child nodes for group nodes.
	Children []TreeNode
	// Expanded controls whether children are visible. Only meaningful for group nodes.
	Expanded bool
	// Depth is the nesting level (0 = top-level, 1 = first child level, etc.).
	Depth int
	// Encrypted indicates this leaf holds a SOPS-encrypted value (displays as "*** (type)").
	Encrypted bool
	// TypeHint is the SOPS type code extracted from ENC[...,type:X] (e.g., "str", "int", "bool").
	// Empty string if not an encrypted leaf.
	TypeHint string
	// IsPlain indicates a leaf value that SOPS left unencrypted (via encrypted_regex/unencrypted_regex).
	// These display their plaintext value with a [plain] badge.
	IsPlain bool
	// Revealed is true when the user has decrypted this leaf and the plaintext is visible.
	// Per D-01: toggled by r key; cleared by ClearAllRevealed on Esc (D-04, T-03-02).
	Revealed bool
	// DecryptedValue holds the plaintext secret value when Revealed=true.
	// It is zeroed by ClearAllRevealed to prevent memory retention (T-03-02).
	DecryptedValue string
}

// flatRow is an internal representation of a single visible row in the tree.
// It is recomputed whenever expand/collapse state changes.
type flatRow struct {
	// node is a pointer into the nodes slice for mutation support.
	node         *TreeNode
	depth        int
	isLast       bool   // last sibling at this level (renders └─ instead of ├─)
	parentIsLast []bool // for each ancestor, whether it was the last sibling (controls │ vs space)
	keyPath      string // e.g., "database.password" -- for fuzzy search
}

// DetailModel is the Bubble Tea component for the YAML tree detail pane.
// It operates on a flat slice of visible rows computed from the tree structure.
type DetailModel struct {
	filename     string
	nodes        []TreeNode
	flatRows     []flatRow // recomputed from tree on each expand/collapse
	cursor       int       // index into flatRows
	scrollTop    int       // first visible row index
	width        int
	height       int
	keys         keys.DetailKeyMap
	searchActive bool
	search       SearchModel
	allFlatRows  []flatRow // full unfiltered flat rows
	isEncrypted  bool      // per D-03: false shows unencrypted banner
	// Inline edit mode fields (D-05, stateEdit)
	editActive  bool            // true while the inline textinput is visible
	editInput   textinput.Model // the textinput component for editing
	editKeyPath string          // dot-joined key path of the node being edited
	editOldVal  string          // original decrypted value before edit (for diff)
}

// NewDetailModel creates a DetailModel for the given file.
// nodes are the top-level tree nodes. Initially all top-level nodes are shown;
// expanded state per node controls child visibility.
func NewDetailModel(filename string, nodes []TreeNode, width, height int, isEncrypted bool) DetailModel {
	m := DetailModel{
		filename:    filename,
		nodes:       nodes,
		cursor:      0,
		width:       width,
		height:      height,
		keys:        keys.DefaultDetailKeyMap,
		isEncrypted: isEncrypted,
	}
	m.flatRows = flattenNodes(m.nodes, 0, nil, "")
	m.allFlatRows = m.flatRows
	m.search = NewSearchModel(width)
	return m
}

// flattenNodes produces a flat list of visible rows from the tree,
// walking recursively and only including children of expanded nodes.
// parentIsLast tracks whether each ancestor is the last sibling (for │ connector logic).
// parentKeyPath is used to compute dot-joined key paths for fuzzy search.
func flattenNodes(nodes []TreeNode, depth int, parentIsLast []bool, parentKeyPath string) []flatRow {
	rows := make([]flatRow, 0, len(nodes))
	for i := range nodes {
		isLast := i == len(nodes)-1
		pil := append(append([]bool(nil), parentIsLast...), isLast) //nolint:gocritic

		// Compute the dot-joined key path for this node
		path := parentKeyPath
		if path != "" {
			path += "."
		}
		path += nodes[i].Key

		rows = append(rows, flatRow{
			node:         &nodes[i],
			depth:        depth,
			isLast:       isLast,
			parentIsLast: parentIsLast,
			keyPath:      path,
		})
		if nodes[i].Expanded && len(nodes[i].Children) > 0 {
			rows = append(rows, flattenNodes(nodes[i].Children, depth+1, pil, path)...)
		}
	}
	return rows
}

// ActivateSearch activates inline fuzzy search mode.
// Returns the Focus cmd from the textinput.
func (m *DetailModel) ActivateSearch() tea.Cmd {
	m.searchActive = true
	return m.search.SetActive(true)
}

// DeactivateSearch deactivates search mode and restores the full flat row list.
func (m *DetailModel) DeactivateSearch() {
	m.searchActive = false
	m.search.Reset()
	m.flatRows = m.allFlatRows
}

// IsSearchActive returns whether search mode is currently active.
func (m DetailModel) IsSearchActive() bool {
	return m.searchActive
}

// IsEditActive returns whether the inline edit mode is currently active.
func (m DetailModel) IsEditActive() bool {
	return m.editActive
}

// Nodes returns the top-level tree nodes (used by AppModel for leaf count).
func (m DetailModel) Nodes() []TreeNode {
	return m.nodes
}

// Update processes messages for the detail view.
// Navigation keys are handled via key.Matches against DetailKeyMap.
// Expand (enter/l/right) and Collapse (h/left) toggle the selected node.
func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	// editActive branch: route all input to the textinput (Pitfall 4: eat j/k nav keys).
	// Enter confirms, Esc cancels; all other keys go to the textinput.
	if m.editActive {
		if kMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch kMsg.String() {
			case "enter":
				newVal := m.editInput.Value()
				m.editActive = false
				m.editInput.Blur()
				oldVal := m.editOldVal
				keyPath := m.editKeyPath
				// Clear sensitive fields after capture (T-03-08)
				m.editOldVal = ""
				m.editKeyPath = ""
				return m, func() tea.Msg {
					return EditConfirmMsg{
						KeyPath:  keyPath,
						OldValue: oldVal,
						NewValue: newVal,
					}
				}
			case "esc":
				m.editActive = false
				m.editInput.Blur()
				m.editOldVal = ""
				m.editKeyPath = ""
				return m, func() tea.Msg { return EditCancelMsg{} }
			}
		}
		// Route all other messages to textinput (eats j/k and other nav keys)
		var cmd tea.Cmd
		m.editInput, cmd = m.editInput.Update(msg)
		return m, cmd
	}

	if m.searchActive {
		switch msg := msg.(type) {
		case tea.KeyPressMsg:
			switch msg.String() {
			case "esc":
				m.DeactivateSearch()
				return m, nil
			case "enter":
				m.searchActive = false
				m.search.Reset()
				return m, nil
			}
		}
		// Route all other messages to search model
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		// Apply filter to flatRows by keyPath
		pattern := m.search.Value()
		if pattern == "" {
			m.flatRows = m.allFlatRows
		} else {
			keyPaths := make([]string, len(m.allFlatRows))
			for i, row := range m.allFlatRows {
				keyPaths[i] = row.keyPath
			}
			matches := ApplyFilter(pattern, keyPaths)
			filtered := make([]flatRow, 0, len(matches))
			for _, match := range matches {
				filtered = append(filtered, m.allFlatRows[match.Index])
			}
			m.flatRows = filtered
		}
		// Clamp cursor
		if m.cursor >= len(m.flatRows) && len(m.flatRows) > 0 {
			m.cursor = len(m.flatRows) - 1
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.GoTop):
			m.cursor = 0
			m.adjustScroll()
			return m, nil

		case key.Matches(msg, m.keys.GoBottom):
			if len(m.flatRows) > 0 {
				m.cursor = len(m.flatRows) - 1
			}
			m.adjustScroll()
			return m, nil

		case key.Matches(msg, m.keys.HalfUp):
			m.cursor -= m.height / 2
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.adjustScroll()
			return m, nil

		case key.Matches(msg, m.keys.HalfDown):
			m.cursor += m.height / 2
			maxIdx := len(m.flatRows) - 1
			if maxIdx < 0 {
				maxIdx = 0
			}
			if m.cursor > maxIdx {
				m.cursor = maxIdx
			}
			m.adjustScroll()
			return m, nil

		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
			m.adjustScroll()
			return m, nil

		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.flatRows)-1 {
				m.cursor++
			}
			m.adjustScroll()
			return m, nil

		case key.Matches(msg, m.keys.Expand):
			if len(m.flatRows) > 0 {
				node := m.flatRows[m.cursor].node
				if len(node.Children) > 0 && !node.Expanded {
					node.Expanded = true
					m.flatRows = flattenNodes(m.nodes, 0, nil, "")
					m.allFlatRows = m.flatRows
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.Collapse):
			if len(m.flatRows) > 0 {
				node := m.flatRows[m.cursor].node
				if len(node.Children) > 0 && node.Expanded {
					node.Expanded = false
					m.flatRows = flattenNodes(m.nodes, 0, nil, "")
					m.allFlatRows = m.flatRows
					// Clamp cursor in case visible rows shrank
					if m.cursor >= len(m.flatRows) && len(m.flatRows) > 0 {
						m.cursor = len(m.flatRows) - 1
					}
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.Reveal):
			// r key: toggle reveal/mask for the selected encrypted leaf (D-01).
			if len(m.flatRows) > 0 {
				row := m.flatRows[m.cursor]
				node := row.node
				if node.Encrypted {
					if node.Revealed {
						// Already revealed → mask it
						m.MaskNode(row.keyPath)
						return m, nil
					}
					// Not yet revealed → request decrypt from AppModel
					keyPath := row.keyPath
					return m, func() tea.Msg {
						return RevealRequestMsg{KeyPath: keyPath}
					}
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.RevealAll):
			// R key: toggle reveal-all / mask-all (D-02).
			if m.AnyRevealed() {
				// Some/all revealed → mask all
				m.MaskAllNodes()
				return m, nil
			}
			// None revealed → request full-file decrypt from AppModel
			filename := m.filename
			return m, func() tea.Msg {
				return RevealAllRequestMsg{FilePath: filename}
			}

		case key.Matches(msg, m.keys.Edit):
			// e key: enter inline edit mode on a revealed encrypted leaf (D-05).
			if len(m.flatRows) > 0 {
				row := m.flatRows[m.cursor]
				node := row.node
				// Guard: block edit on array-indexed keys (Open Question 1 resolution)
				if sops.IsArrayIndexedKeyPath(row.keyPath) {
					return m, func() tea.Msg {
						return EditBlockedMsg{Reason: "Array-indexed keys not editable in Phase 3"}
					}
				}
				if node.Encrypted && node.Revealed {
					m.editActive = true
					m.editKeyPath = row.keyPath
					m.editOldVal = node.DecryptedValue
					ti := textinput.New()
					ti.CharLimit = 1000 // T-03-09: prevent memory exhaustion
					ti.Prompt = ""
					ti.SetValue(node.DecryptedValue)
					m.editInput = ti
					cmd := m.editInput.Focus()
					return m, cmd
				} else if node.Encrypted && !node.Revealed {
					// Masked leaf — return blocked msg for AppModel to flash "Reveal first with r"
					return m, func() tea.Msg {
						return EditBlockedMsg{}
					}
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.EditFile):
			// E key: open decrypted file in $EDITOR (EDT-01).
			if m.AnyRevealed() {
				filename := m.filename
				return m, func() tea.Msg {
					return EditorRequestMsg{FilePath: filename}
				}
			}
			return m, func() tea.Msg { return EditBlockedMsg{} }

		case key.Matches(msg, m.keys.Rotate):
			// X key: format-aware secret rotation (EDT-03).
			if len(m.flatRows) > 0 {
				row := m.flatRows[m.cursor]
				node := row.node
				// Guard: block rotation on array-indexed keys (Open Question 1 resolution)
				if sops.IsArrayIndexedKeyPath(row.keyPath) {
					return m, func() tea.Msg {
						return EditBlockedMsg{Reason: "Array-indexed keys not editable in Phase 3"}
					}
				}
				if node.Encrypted && node.Revealed {
					detected := DetectFormat(node.DecryptedValue)
					if detected == FormatUnknown {
						// Ambiguous: trigger format selection menu
						keyPath := row.keyPath
						oldValue := node.DecryptedValue
						return m, func() tea.Msg {
							return RotateFormatMenuMsg{KeyPath: keyPath, OldValue: oldValue}
						}
					}
					// Auto-detected: generate new value and go to diff
					keyPath := row.keyPath
					oldValue := node.DecryptedValue
					format := detected
					return m, func() tea.Msg {
						newVal, err := GenerateValue(format)
						if err != nil {
							return RotateErrorMsg{Err: err}
						}
						return RotateReadyMsg{
							KeyPath:  keyPath,
							OldValue: oldValue,
							NewValue: newVal,
							Format:   format,
						}
					}
				} else if node.Encrypted && !node.Revealed {
					// Masked leaf — return blocked msg for AppModel to flash "Reveal first with r"
					return m, func() tea.Msg {
						return EditBlockedMsg{}
					}
				}
			}
			return m, nil
		}
	}
	return m, nil
}

// adjustScroll ensures the cursor is within the visible viewport.
func (m *DetailModel) adjustScroll() {
	if m.cursor < m.scrollTop {
		m.scrollTop = m.cursor
	}
	if m.height > 0 && m.cursor >= m.scrollTop+m.height {
		m.scrollTop = m.cursor - m.height + 1
	}
}

// View renders the YAML tree.
// If there are no nodes, the empty state is shown per UI-SPEC copywriting contract.
// If isEncrypted is false, an unencrypted banner is prepended before the tree.
// When search is active, the search bar is appended at the bottom.
func (m DetailModel) View() string {
	if len(m.nodes) == 0 {
		return DimText.Render("No keys found in this file")
	}
	if len(m.flatRows) == 0 && !m.searchActive {
		return DimText.Render("No keys found in this file")
	}

	var sb strings.Builder

	// Per D-03 and UI-SPEC Unencrypted File Banner Contract:
	// Prepend banner when file is not yet encrypted.
	if !m.isEncrypted && len(m.nodes) > 0 {
		sb.WriteString(WarnLabel.Render("Not yet encrypted \u2014 matches .sops.yaml rules"))
		sb.WriteByte('\n')
		sb.WriteByte('\n')
	}

	// Determine visible range
	start := m.scrollTop
	end := start + m.height
	if end > len(m.flatRows) {
		end = len(m.flatRows)
	}

	for idx := start; idx < end; idx++ {
		row := m.flatRows[idx]
		node := row.node

		var line string
		if idx == m.cursor && m.editActive {
			// Render key portion normally, replace value with textinput (D-05)
			keyPart := renderRowKeyOnly(row, node)
			inputWidth := m.width - lipgloss.Width(keyPart) - 4
			if inputWidth < 1 {
				inputWidth = 1
			}
			inputView := EditInputStyle.Width(inputWidth).Render(m.editInput.View())
			line = keyPart + inputView
		} else {
			line = renderRow(row, node, m.width)
			if idx == m.cursor {
				// Apply selected row style across full width
				line = SelectedRow.Width(m.width).Render(line)
			}
		}

		if idx > start {
			sb.WriteByte('\n')
		}
		sb.WriteString(line)
	}

	content := sb.String()

	if m.searchActive {
		return content + "\n" + m.search.View()
	}
	return content
}

// renderRow builds the rendered string for a single tree row.
// Per UI-SPEC §YAML Tree Rendering:
//   - Connectors in TreeConnector style (muted)
//   - Indicators in TreeIndicator style (accent)
//   - Values in DimText style (faint)
//   - Type hints in TypeHintStyle (canonical from styles.go)
//   - Plain badges in BadgePlain (canonical from styles.go)
func renderRow(row flatRow, node *TreeNode, _ int) string {
	var sb strings.Builder

	// Parent continuation lines — one column per ancestor.
	// If ancestor was NOT last, draw │; otherwise draw spaces.
	for _, ancestorIsLast := range row.parentIsLast {
		if ancestorIsLast {
			sb.WriteString(TreeConnector.Render("   "))
		} else {
			sb.WriteString(TreeConnector.Render("│  "))
		}
	}

	// Immediate connector for this node.
	if row.isLast {
		sb.WriteString(TreeConnector.Render("└─ "))
	} else {
		sb.WriteString(TreeConnector.Render("├─ "))
	}

	// Node content
	if len(node.Children) > 0 {
		// Group node
		if node.Expanded {
			sb.WriteString(TreeIndicator.Render("[-]"))
		} else {
			sb.WriteString(TreeIndicator.Render("[+]"))
		}
		sb.WriteString(" ")
		sb.WriteString(node.Key)
	} else {
		// Leaf node rendering depends on encryption state
		sb.WriteString(node.Key)
		sb.WriteString(": ")
		if node.Encrypted && node.Revealed {
			// Revealed encrypted value: show plaintext + lock-open icon (D-03, 03-UI-SPEC.md)
			sb.WriteString(RevealedValueStyle.Render(node.DecryptedValue))
			sb.WriteString("  ")
			sb.WriteString(RevealedIconStyle.Render("\U0001F513"))
		} else if node.Encrypted {
			// Encrypted value: show *** with type hint using canonical TypeHintStyle
			sb.WriteString(DimText.Render("***"))
			sb.WriteString(TypeHintStyle.Render(" (" + node.TypeHint + ")"))
		} else if node.IsPlain {
			// Plain (unencrypted) value in a SOPS file: show value with [plain] badge
			// using canonical BadgePlain style
			sb.WriteString(node.Value)
			sb.WriteString("  ")
			sb.WriteString(BadgePlain.Render("[plain]"))
		} else {
			// Default fallback: masked value (Phase 1 behavior)
			sb.WriteString(DimText.Render("***"))
		}
	}

	return sb.String()
}

// renderRowKeyOnly builds the key portion of a tree row (connectors + key name + ": ")
// but omits the value. Used in edit mode to render the key column frozen while the
// value column is replaced by the inline textinput.
func renderRowKeyOnly(row flatRow, node *TreeNode) string {
	var sb strings.Builder

	// Parent continuation lines
	for _, ancestorIsLast := range row.parentIsLast {
		if ancestorIsLast {
			sb.WriteString(TreeConnector.Render("   "))
		} else {
			sb.WriteString(TreeConnector.Render("│  "))
		}
	}

	// Immediate connector for this node
	if row.isLast {
		sb.WriteString(TreeConnector.Render("└─ "))
	} else {
		sb.WriteString(TreeConnector.Render("├─ "))
	}

	// Key name + ": " separator (leaf nodes only — edit only activates on leaves)
	sb.WriteString(node.Key)
	sb.WriteString(": ")

	return sb.String()
}

// SetSize updates the component dimensions.
func (m *DetailModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.search.SetWidth(width)
	m.adjustScroll()
}

// SelectedIndex returns the current cursor position into the flat row list.
func (m DetailModel) SelectedIndex() int {
	return m.cursor
}

// ClearAllRevealed walks all nodes recursively and sets Revealed=false, DecryptedValue=""
// on every node. Called on Esc-to-file-list transition (D-04, T-03-02: prevent memory leak).
func (m *DetailModel) ClearAllRevealed() {
	clearRevealedNodes(m.nodes)
	m.flatRows = flattenNodes(m.nodes, 0, nil, "")
	m.allFlatRows = m.flatRows
}

// clearRevealedNodes recursively clears Revealed/DecryptedValue on all nodes.
func clearRevealedNodes(nodes []TreeNode) {
	for i := range nodes {
		nodes[i].Revealed = false
		nodes[i].DecryptedValue = ""
		if len(nodes[i].Children) > 0 {
			clearRevealedNodes(nodes[i].Children)
		}
	}
}

// RevealNode finds the node matching dotKeyPath and sets Revealed=true with the given value.
// Matching by keyPath (not cursor index) avoids Pitfall 2 (cursor race condition).
// Re-flattens after mutation.
func (m *DetailModel) RevealNode(dotKeyPath, value string) {
	setRevealedByPath(m.nodes, dotKeyPath, "", value, true)
	m.flatRows = flattenNodes(m.nodes, 0, nil, "")
	m.allFlatRows = m.flatRows
}

// MaskNode finds the node matching dotKeyPath and sets Revealed=false, DecryptedValue="".
// Re-flattens after mutation.
func (m *DetailModel) MaskNode(dotKeyPath string) {
	setRevealedByPath(m.nodes, dotKeyPath, "", "", false)
	m.flatRows = flattenNodes(m.nodes, 0, nil, "")
	m.allFlatRows = m.flatRows
}

// MaskAllNodes masks all revealed nodes. Alias for ClearAllRevealed for consistency.
func (m *DetailModel) MaskAllNodes() {
	m.ClearAllRevealed()
}

// RevealAllNodes walks all leaf nodes and sets Revealed=true / DecryptedValue from the
// provided map (keyPath → plaintext value). Re-flattens after mutation.
func (m *DetailModel) RevealAllNodes(values map[string]string) {
	revealAllByPath(m.nodes, "", values)
	m.flatRows = flattenNodes(m.nodes, 0, nil, "")
	m.allFlatRows = m.flatRows
}

// AnyRevealed returns true if any leaf node has Revealed=true.
func (m DetailModel) AnyRevealed() bool {
	return anyRevealedNodes(m.nodes)
}

// anyRevealedNodes recursively checks if any node has Revealed=true.
func anyRevealedNodes(nodes []TreeNode) bool {
	for i := range nodes {
		if nodes[i].Revealed {
			return true
		}
		if len(nodes[i].Children) > 0 && anyRevealedNodes(nodes[i].Children) {
			return true
		}
	}
	return false
}

// setRevealedByPath walks nodes recursively, computing each node's dot-joined keyPath.
// When a match is found, sets Revealed=reveal and DecryptedValue=value.
func setRevealedByPath(nodes []TreeNode, targetPath, parentPath, value string, reveal bool) {
	for i := range nodes {
		path := parentPath
		if path != "" {
			path += "."
		}
		path += nodes[i].Key
		if path == targetPath {
			nodes[i].Revealed = reveal
			if reveal {
				nodes[i].DecryptedValue = value
			} else {
				nodes[i].DecryptedValue = ""
			}
			return
		}
		if len(nodes[i].Children) > 0 {
			setRevealedByPath(nodes[i].Children, targetPath, path, value, reveal)
		}
	}
}

// revealAllByPath walks all leaf nodes and reveals those whose keyPath is in the values map.
func revealAllByPath(nodes []TreeNode, parentPath string, values map[string]string) {
	for i := range nodes {
		path := parentPath
		if path != "" {
			path += "."
		}
		path += nodes[i].Key
		if len(nodes[i].Children) == 0 {
			// Leaf node — reveal if in map
			if val, ok := values[path]; ok {
				nodes[i].Revealed = true
				nodes[i].DecryptedValue = val
			}
		} else {
			revealAllByPath(nodes[i].Children, path, values)
		}
	}
}
