// Package ui — file list tree rendering.
//
// This file implements the directory-tree presentation for FileListModel.
// FileItem stays flat; the tree exists only as a presentation layer built
// fresh from []FileItem on every items-change. Collapse state lives on
// FileListModel keyed by directory path so it survives across rebuilds.
package ui

import (
	"path"
	"sort"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"

	"github.com/caesarakalaeii/sops-tui/internal/sops"
)

// dirItem is a directory node rendered as a row in the file list.
// Implements list.Item and list.DefaultItem so it can sit in bubbles/list
// alongside fileTreeItem.
type dirItem struct {
	// fullPath is the directory path relative to the repo root, e.g. "secrets/base".
	fullPath string
	// name is the last path segment, e.g. "base".
	name string
	// expanded is true when children are visible; false when collapsed.
	expanded bool
	// depth is the nesting level (0 == root child).
	depth int
	// isLast is true when this row is the last sibling at its level.
	isLast bool
	// parentIsLast records whether each ancestor was the last at its level.
	// Used to render │ vs blank for the continuation columns.
	parentIsLast []bool
	// dim is true when the row is rendered faded (e.g., during search with no
	// descendant matching).
	dim bool
}

// Title returns the rendered tree row for a directory.
func (d dirItem) Title() string {
	var sb strings.Builder
	for _, last := range d.parentIsLast {
		if last {
			sb.WriteString(TreeConnector.Render("   "))
		} else {
			sb.WriteString(TreeConnector.Render("│  "))
		}
	}
	if d.isLast {
		sb.WriteString(TreeConnector.Render("└─ "))
	} else {
		sb.WriteString(TreeConnector.Render("├─ "))
	}
	indicator := "▸"
	if d.expanded {
		indicator = "▾"
	}
	sb.WriteString(TreeIndicator.Render(indicator + " "))
	name := d.name + "/"
	if d.dim {
		sb.WriteString(DimText.Render(name))
	} else {
		sb.WriteString(name)
	}
	return sb.String()
}

// Description returns empty so the default delegate renders single-line rows
// (with ShowDescription=false there is no second line at all).
func (d dirItem) Description() string { return "" }

// FilterValue lets bubbles/list filter on the directory path. We disable the
// built-in filter, so this is mostly cosmetic, but keeping it stable lets
// future code reuse the field if the cross-file search ever extends to dirs.
func (d dirItem) FilterValue() string { return d.fullPath }

// fileTreeItem wraps FileItem with tree display state.
// FileItem.Title() rendering is intentionally NOT delegated — fileTreeItem
// renders the basename (last path segment) with the usual badges, since
// the directory ancestry is already expressed by the tree connectors.
type fileTreeItem struct {
	file         FileItem
	depth        int
	isLast       bool
	parentIsLast []bool
	dim          bool
}

// Title renders one file row: ancestor connectors + own connector + basename + badges.
func (f fileTreeItem) Title() string {
	var sb strings.Builder
	for _, last := range f.parentIsLast {
		if last {
			sb.WriteString(TreeConnector.Render("   "))
		} else {
			sb.WriteString(TreeConnector.Render("│  "))
		}
	}
	if f.isLast {
		sb.WriteString(TreeConnector.Render("└─ "))
	} else {
		sb.WriteString(TreeConnector.Render("├─ "))
	}
	base := path.Base(f.file.Name)
	if f.file.Selected {
		base = SelectionIndicatorStyle.Render("[+]") + " " + base
	}
	body := base + fileBadges(f.file)
	if f.dim {
		sb.WriteString(DimText.Render(body))
	} else {
		sb.WriteString(body)
	}
	return sb.String()
}

// Description: empty for single-line rendering.
func (f fileTreeItem) Description() string { return "" }

// FilterValue is the relative path so cross-file search semantics still match
// against the same string FileItem exposes.
func (f fileTreeItem) FilterValue() string { return f.file.Name }

// fileBadges returns the trailing badge string for a file (encryption + git).
// Selection prefix is rendered in fileTreeItem.Title() / FileItem.Title()
// because it goes BEFORE the name.
func fileBadges(i FileItem) string {
	var b strings.Builder
	if !i.IsEncrypted {
		b.WriteString(" ")
		b.WriteString(BadgeUnencrypted.Render("[unencrypted]"))
	}
	switch i.GitStatus {
	case "M":
		b.WriteString(" ")
		b.WriteString(BadgeModified.Render("[M]"))
	case "A":
		b.WriteString(" ")
		b.WriteString(BadgeAdded.Render("[A]"))
	case "?":
		b.WriteString(" ")
		b.WriteString(BadgeUntracked.Render("[?]"))
	}
	return b.String()
}

// treeNode is the in-memory tree built from []FileItem.
// It has children (subdirectories) keyed by segment name and a slice of
// files that live directly in this directory.
type treeNode struct {
	// name is the segment name, e.g. "base". Empty for the root.
	name string
	// fullPath is the relative path from the root, e.g. "secrets/base". Empty for the root.
	fullPath string
	// children maps a child segment name to its node.
	children map[string]*treeNode
	// files are the files in THIS directory (not its subdirs).
	files []FileItem
}

// buildTree groups []FileItem into a directory tree using the slash-separated
// FileItem.Name as the path. Files at the repo root attach to the root node's
// files slice with no children.
func buildTree(items []FileItem) *treeNode {
	root := &treeNode{children: map[string]*treeNode{}}
	for _, it := range items {
		segs := strings.Split(it.Name, "/")
		cursor := root
		// walk every segment except the last (the basename) — those are dirs
		for i := 0; i < len(segs)-1; i++ {
			seg := segs[i]
			child, ok := cursor.children[seg]
			if !ok {
				child = &treeNode{
					name:     seg,
					fullPath: strings.Join(segs[:i+1], "/"),
					children: map[string]*treeNode{},
				}
				cursor.children[seg] = child
			}
			cursor = child
		}
		cursor.files = append(cursor.files, it)
	}
	return root
}

// flatten walks the tree honoring `collapsed` and produces an ordered slice of
// list.Item rows ready to feed into bubbles/list.
//
// Sort order at each level: subdirectories first (alphabetical), then files
// (alphabetical by basename). dirs that appear in `collapsed` are emitted as
// rows but their descendants are skipped.
//
// `dimMatch` (optional, may be nil) when set is consulted per-file: if it
// returns false, the row is rendered with .dim=true. Dirs are dimmed when
// none of their descendant files match.
func flatten(t *treeNode, collapsed map[string]bool, dimMatch func(FileItem) bool) []list.Item {
	var out []list.Item
	flattenInto(&out, t, 0, nil, collapsed, dimMatch)
	return out
}

// flattenInto is the recursive worker.
func flattenInto(
	out *[]list.Item,
	node *treeNode,
	depth int,
	parentIsLast []bool,
	collapsed map[string]bool,
	dimMatch func(FileItem) bool,
) {
	// Sort children alphabetically.
	dirNames := make([]string, 0, len(node.children))
	for name := range node.children {
		dirNames = append(dirNames, name)
	}
	sort.Strings(dirNames)

	// Sort files alphabetically by basename. The slice is a copy so the
	// original FileItem ordering on FileListModel is preserved.
	files := make([]FileItem, len(node.files))
	copy(files, node.files)
	sort.Slice(files, func(i, j int) bool {
		return path.Base(files[i].Name) < path.Base(files[j].Name)
	})

	totalSiblings := len(dirNames) + len(files)
	emitted := 0

	// Subdirectories first. Linear chains (a dir whose only content is a
	// single subdir, with no files of its own) are visually compacted into
	// a single row showing the joined path — so e.g.
	//   apps/workloads/caesar-website/secrets/foo.yaml
	// renders as one dir row "apps/workloads/caesar-website/secrets/" with
	// the file beneath, instead of four nested rows.
	for _, dirName := range dirNames {
		child := node.children[dirName]
		terminal, mergedName := compactLinearChain(child)
		isLast := emitted == totalSiblings-1
		expanded := !collapsed[terminal.fullPath]
		dirRow := dirItem{
			fullPath:     terminal.fullPath,
			name:         mergedName,
			expanded:     expanded,
			depth:        depth,
			isLast:       isLast,
			parentIsLast: append([]bool{}, parentIsLast...),
		}
		if dimMatch != nil && !subtreeHasMatch(terminal, dimMatch) {
			dirRow.dim = true
		}
		*out = append(*out, dirRow)
		emitted++
		if expanded {
			childParents := append(append([]bool{}, parentIsLast...), isLast)
			flattenInto(out, terminal, depth+1, childParents, collapsed, dimMatch)
		}
	}

	// Files at this level.
	for _, f := range files {
		isLast := emitted == totalSiblings-1
		row := fileTreeItem{
			file:         f,
			depth:        depth,
			isLast:       isLast,
			parentIsLast: append([]bool{}, parentIsLast...),
		}
		if dimMatch != nil && !dimMatch(f) {
			row.dim = true
		}
		*out = append(*out, row)
		emitted++
	}
}

// compactLinearChain walks down a linear chain of single-child directories
// (each having exactly one subdir and no files of its own) and returns the
// terminal node plus the joined display name. A non-linear node returns
// itself unchanged.
//
// Example: a → b → c, where each only has the next one and c has files,
// returns (c, "a/b/c").
func compactLinearChain(node *treeNode) (*treeNode, string) {
	terminal := node
	merged := node.name
	for len(terminal.children) == 1 && len(terminal.files) == 0 {
		var onlyName string
		for k := range terminal.children {
			onlyName = k
		}
		terminal = terminal.children[onlyName]
		merged = merged + "/" + terminal.name
	}
	return terminal, merged
}

// subtreeHasMatch returns true if `match` returns true for any file in `node`
// or any of its descendants.
func subtreeHasMatch(node *treeNode, match func(FileItem) bool) bool {
	for _, f := range node.files {
		if match(f) {
			return true
		}
	}
	for _, child := range node.children {
		if subtreeHasMatch(child, match) {
			return true
		}
	}
	return false
}

// allDirPaths returns every directory fullPath in the tree (depth-first).
// Used when restoring expansion state and when implementing "expand all
// ancestors of a match" during search.
func allDirPaths(node *treeNode) []string {
	var out []string
	collectDirPaths(&out, node)
	return out
}

func collectDirPaths(out *[]string, node *treeNode) {
	for _, child := range node.children {
		*out = append(*out, child.fullPath)
		collectDirPaths(out, child)
	}
}

// ancestorPaths returns the chain of directory paths leading to `dir`,
// e.g. "a/b/c" → ["a", "a/b", "a/b/c"]. Empty input → empty slice.
func ancestorPaths(dir string) []string {
	if dir == "" {
		return nil
	}
	segs := strings.Split(dir, "/")
	out := make([]string, 0, len(segs))
	for i := range segs {
		out = append(out, strings.Join(segs[:i+1], "/"))
	}
	return out
}

// fileParentDir returns the directory portion of a FileItem.Name, or ""
// for files at the repo root.
func fileParentDir(name string) string {
	d := path.Dir(name)
	if d == "." {
		return ""
	}
	return d
}

// findFileTreeIndex locates the row index of a FileItem in `rows` by Path.
// Returns -1 if not present (e.g., its parent dir is collapsed).
func findFileTreeIndex(rows []list.Item, target FileItem) int {
	for i, r := range rows {
		if f, ok := r.(fileTreeItem); ok && f.file.Path == target.Path {
			return i
		}
	}
	return -1
}

// findFirstFileIndex returns the index of the first fileTreeItem in rows,
// or -1 if there are no files.
func findFirstFileIndex(rows []list.Item) int {
	for i, r := range rows {
		if _, ok := r.(fileTreeItem); ok {
			return i
		}
	}
	return -1
}

// collectDirFiles returns every FileItem under the given directory subtree.
// `dirPath` is the relative path of the directory ("" means root, returns
// only root-level files; nil descent rules apply).
func collectDirFiles(node *treeNode, dirPath string) []FileItem {
	target := findDirNode(node, dirPath)
	if target == nil {
		return nil
	}
	var out []FileItem
	collectAllFiles(&out, target)
	return out
}

func findDirNode(node *treeNode, dirPath string) *treeNode {
	if dirPath == "" {
		return node
	}
	segs := strings.Split(dirPath, "/")
	cursor := node
	for _, seg := range segs {
		child, ok := cursor.children[seg]
		if !ok {
			return nil
		}
		cursor = child
	}
	return cursor
}

func collectAllFiles(out *[]FileItem, node *treeNode) {
	*out = append(*out, node.files...)
	for _, child := range node.children {
		collectAllFiles(out, child)
	}
}

// _ unused-import guards for go vet — strict imports list. Not actually unused;
// listed here so future contributors don't drop them by mistake.
var (
	_ = lipgloss.NewStyle
	_ = sops.CreationRule{}
)
