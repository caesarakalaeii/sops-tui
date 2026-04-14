// Package ui provides the view components for sops-tui.
// This file implements DetailModel, a Bubble Tea component that renders
// a collapsible YAML tree view for a single SOPS-encrypted file.
//
// Per D-06: YAML tree indentation preserves structure (2 cells per level).
// Per D-07: [+]/[-] indicators show node collapse state.
// Per NAV-03: j/k/g/G/ctrl-d/u navigation; Enter/l expands, h/left collapses.
// Per 01-UI-SPEC.md §YAML Tree Rendering: connectors ├─ └─ │ in muted color.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
)

// Temporary styles -- replaced by canonical styles from styles.go in Plan 02-03
var typeHintStyleTemp = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("#6c7086"))
var badgePlainTemp = lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af"))

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
}

// flatRow is an internal representation of a single visible row in the tree.
// It is recomputed whenever expand/collapse state changes.
type flatRow struct {
	// node is a pointer into the nodes slice for mutation support.
	node         *TreeNode
	depth        int
	isLast       bool   // last sibling at this level (renders └─ instead of ├─)
	parentIsLast []bool // for each ancestor, whether it was the last sibling (controls │ vs space)
}

// DetailModel is the Bubble Tea component for the YAML tree detail pane.
// It operates on a flat slice of visible rows computed from the tree structure.
type DetailModel struct {
	filename  string
	nodes     []TreeNode
	flatRows  []flatRow // recomputed from tree on each expand/collapse
	cursor    int       // index into flatRows
	scrollTop int       // first visible row index
	width     int
	height    int
	keys      keys.DetailKeyMap
}

// NewDetailModel creates a DetailModel for the given file.
// nodes are the top-level tree nodes. Initially all top-level nodes are shown;
// expanded state per node controls child visibility.
func NewDetailModel(filename string, nodes []TreeNode, width, height int) DetailModel {
	m := DetailModel{
		filename: filename,
		nodes:    nodes,
		cursor:   0,
		width:    width,
		height:   height,
		keys:     keys.DefaultDetailKeyMap,
	}
	m.flatRows = flattenNodes(m.nodes, 0, nil)
	return m
}

// flattenNodes produces a flat list of visible rows from the tree,
// walking recursively and only including children of expanded nodes.
// parentIsLast tracks whether each ancestor is the last sibling (for │ connector logic).
func flattenNodes(nodes []TreeNode, depth int, parentIsLast []bool) []flatRow {
	rows := make([]flatRow, 0, len(nodes))
	for i := range nodes {
		isLast := i == len(nodes)-1
		pil := append(append([]bool(nil), parentIsLast...), isLast) //nolint:gocritic
		rows = append(rows, flatRow{
			node:         &nodes[i],
			depth:        depth,
			isLast:       isLast,
			parentIsLast: parentIsLast,
		})
		if nodes[i].Expanded && len(nodes[i].Children) > 0 {
			rows = append(rows, flattenNodes(nodes[i].Children, depth+1, pil)...)
		}
	}
	return rows
}

// Update processes messages for the detail view.
// Navigation keys are handled via key.Matches against DetailKeyMap.
// Expand (enter/l/right) and Collapse (h/left) toggle the selected node.
func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
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
					m.flatRows = flattenNodes(m.nodes, 0, nil)
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.Collapse):
			if len(m.flatRows) > 0 {
				node := m.flatRows[m.cursor].node
				if len(node.Children) > 0 && node.Expanded {
					node.Expanded = false
					m.flatRows = flattenNodes(m.nodes, 0, nil)
					// Clamp cursor in case visible rows shrank
					if m.cursor >= len(m.flatRows) && len(m.flatRows) > 0 {
						m.cursor = len(m.flatRows) - 1
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
// Each row is rendered with:
//   - Indentation: 2 cells per depth level
//   - Tree connectors: ├─ (non-last) or └─ (last) in TreeConnector style (muted color)
//   - Parent continuation: │  or spaces based on ancestor isLast state
//   - Group nodes: [+] (collapsed) or [-] (expanded) in TreeIndicator style (accent color)
//   - Leaf nodes: key: *** with value in DimText
//   - Selected row: SelectedRow style applied across full width
func (m DetailModel) View() string {
	if len(m.nodes) == 0 {
		return DimText.Render("No keys found in this file")
	}
	if len(m.flatRows) == 0 {
		return DimText.Render("No keys found in this file")
	}

	var sb strings.Builder

	// Determine visible range
	start := m.scrollTop
	end := start + m.height
	if end > len(m.flatRows) {
		end = len(m.flatRows)
	}

	for idx := start; idx < end; idx++ {
		row := m.flatRows[idx]
		node := row.node

		line := renderRow(row, node, m.width)

		if idx == m.cursor {
			// Apply selected row style across full width
			line = SelectedRow.Width(m.width).Render(line)
		}

		if idx > start {
			sb.WriteByte('\n')
		}
		sb.WriteString(line)
	}

	return sb.String()
}

// renderRow builds the rendered string for a single tree row.
// Per UI-SPEC §YAML Tree Rendering:
//   - Connectors in TreeConnector style (muted)
//   - Indicators in TreeIndicator style (accent)
//   - Values in DimText style (faint)
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
		if node.Encrypted {
			// Encrypted value: show *** with type hint
			sb.WriteString(DimText.Render("***"))
			sb.WriteString(typeHintStyleTemp.Render(" (" + node.TypeHint + ")"))
		} else if node.IsPlain {
			// Plain (unencrypted) value in a SOPS file: show value with [plain] badge
			sb.WriteString(node.Value)
			sb.WriteString("  ")
			sb.WriteString(badgePlainTemp.Render("[plain]"))
		} else {
			// Default fallback: masked value (Phase 1 behavior)
			sb.WriteString(DimText.Render("***"))
		}
	}

	return sb.String()
}

// SetSize updates the component dimensions.
func (m *DetailModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.adjustScroll()
}

// SelectedIndex returns the current cursor position into the flat row list.
func (m DetailModel) SelectedIndex() int {
	return m.cursor
}
