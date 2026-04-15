// Package app provides the root Bubble Tea model for sops-tui.
//
// AppModel owns the sessionState enum that routes all messages and view
// rendering to the correct child component. It composes FileListModel,
// DetailModel, HelpModel, StatusBarModel, and MetadataModel from the ui package.
//
// Navigation flow:
//   - stateFileList: default entry view; Enter/l drills into detail
//   - stateDetail: YAML tree view; Esc returns to file list
//   - stateHelp: full-screen overlay; ? or Esc returns to previous state
//   - stateMetadata: full-screen metadata overlay; i or Esc returns to previous state
//
// Per RESEARCH.md Pattern 1: Root Model with sessionState Enum.
// Per CLAUDE.md: View() returns tea.View, use tea.KeyPressMsg, v.AltScreen = true.
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	yaml "github.com/goccy/go-yaml"

	gitpkg "github.com/caesarakalaeii/sops-tui/internal/git"
	"github.com/caesarakalaeii/sops-tui/internal/keys"
	"github.com/caesarakalaeii/sops-tui/internal/parser"
	"github.com/caesarakalaeii/sops-tui/internal/sops"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

// sessionState controls which child model receives updates and renders.
type sessionState int

const (
	// stateFileList is the default view — the file browser pane.
	stateFileList sessionState = iota
	// stateDetail is the YAML tree detail view for a selected file.
	stateDetail
	// stateHelp is the full-screen contextual keybinding overlay.
	stateHelp
	// stateMetadata is the full-screen SOPS metadata overlay.
	stateMetadata
	// stateDiff is the full-screen diff overlay before re-encryption (D-09, D-10).
	stateDiff
	// stateEdit is the inline value edit mode (D-05).
	stateEdit
	// stateFormatMenu is the format selection menu for ambiguous rotation (EDT-03).
	stateFormatMenu
	// stateHistory is the full-screen git history overlay (D-13, D-14, D-15).
	stateHistory
)

// StateDiff, StateEdit, StateFormatMenu, StateHistory are exported for tests to verify the constants exist.
const (
	StateDiff       = stateDiff
	StateEdit       = stateEdit
	StateFormatMenu = stateFormatMenu
	StateHistory    = stateHistory
)

// FilesDiscoveredMsg carries the result of SOPS file discovery.
type FilesDiscoveredMsg struct {
	Files []sops.DiscoveredFile
	Err   error
}

// FilesParsedMsg carries the result of parsing a single YAML file.
type FilesParsedMsg struct {
	Parsed parser.ParsedFile
	Err    error
}

// DecryptKeyMsg carries the result of a sops.DecryptKey subprocess call.
// KeyPath is the dot-joined path of the decrypted key; Value is the plaintext.
// If Err is non-nil the decrypt failed and Value is empty.
type DecryptKeyMsg struct {
	KeyPath string
	Value   string
	Err     error
}

// DecryptAllMsg carries the result of a sops.DecryptFile subprocess call.
// Values maps dot-joined key paths to plaintext values for all encrypted leaves.
// If Err is non-nil the decrypt failed and Values is nil.
type DecryptAllMsg struct {
	Values map[string]string
	Err    error
}

// ReEncryptDoneMsg carries the result of a sops.SetKey or sops.EncryptFile call.
// Err is non-nil if the re-encryption failed.
type ReEncryptDoneMsg struct {
	Err error
}

// EditorReadyMsg is sent internally after DecryptFile succeeds for the $EDITOR flow.
// AppModel handles it by writing the decrypted content to a temp file and launching
// $EDITOR via tea.ExecProcess.
type EditorReadyMsg struct {
	DecryptedContent []byte
}

// EditorFinishedMsg is sent by the tea.ExecProcess callback after $EDITOR exits.
// TmpPath is the path to the temp file containing the (possibly edited) decrypted YAML.
// OriginalContent is the original decrypted YAML before editing (for diff comparison).
// Err is non-nil if the editor exited with a non-zero status.
type EditorFinishedMsg struct {
	TmpPath         string // path to edited temp file (must be cleaned up by handler)
	OriginalContent []byte // original decrypted YAML for YAML-aware diff comparison
	Err             error
}

// ClipboardClearMsg is sent by tea.Tick after the clipboard timeout expires.
// Gen is compared against clipboardGen to prevent stale clears (D-06, Pitfall 2).
type ClipboardClearMsg struct {
	Gen int
}

// GitStatusMsg carries the result of async git status fetch.
// Statuses maps relative file paths to their git status code.
// GitAvailable is false when the directory is not a git repo (D-12).
type GitStatusMsg struct {
	Statuses     map[string]gitpkg.GitStatus
	GitAvailable bool
	Err          error
}

// CrossFileSearchItem represents a searchable item combining file name and key path.
// Used by cross-file search (GIT-03) to enable / from file list to search across all files and keys.
type CrossFileSearchItem struct {
	FileName  string           // relative file path, e.g. "secrets/prod.yaml"
	KeyPath   string           // dot-joined key path, e.g. "database.password" (empty for file-level match)
	AbsPath   string           // absolute path for navigation
	Rule      sops.CreationRule
	IsEnc     bool
	GitStatus string
}

// Title returns the display string for search results.
// File-level items show just the filename. Key-level items show "filename > key.path".
func (c CrossFileSearchItem) Title() string {
	if c.KeyPath == "" {
		return c.FileName
	}
	return c.FileName + " > " + c.KeyPath
}

// HistoryRequestMsg is sent when user presses b in detail view.
type HistoryRequestMsg struct {
	FilePath string // absolute path
	RelPath  string // relative path (for go-git FileName filter)
}

// GitHistoryMsg carries async git history results.
type GitHistoryMsg struct {
	Entries []gitpkg.CommitEntry
	Err     error
}

// ParsedFileForTest is a test helper that builds a parser.ParsedFile from tree nodes.
// It is exported so external test packages (_test suffix) can drive AppModel into stateDetail.
func ParsedFileForTest(nodes []ui.TreeNode) parser.ParsedFile {
	return parser.ParsedFile{Nodes: nodes}
}

// AppModel is the root Bubble Tea model. It owns a sessionState enum and
// holds all child models as value fields. Methods return updated copies.
type AppModel struct {
	state        sessionState
	prevState    sessionState // restored when help/metadata overlay closes
	width        int
	height       int
	fileList     ui.FileListModel
	detail       ui.DetailModel
	help         ui.HelpModel
	status       ui.StatusBarModel
	metadata     ui.MetadataModel
	diff         ui.DiffModel          // stateDiff overlay for confirming writes
	editFilePath string                // file path for current edit operation (for SetKey)
	sopsYamlPath string                // path to .sops.yaml
	files        []sops.DiscoveredFile // cached discovery results
	currentFile  sops.DiscoveredFile   // file currently shown in detail
	// $EDITOR flow fields (EDT-01, T-03-12)
	editorEditedContent []byte // edited YAML from $EDITOR; nil after confirm/cancel
	// Format menu fields (EDT-03)
	formatMenuActive   bool
	formatMenuKeyPath  string
	formatMenuOldValue string
	formatMenuCursor   int
	// Rotation tracking for flash message
	rotateFormat ui.SecretFormat
	// Clipboard fields (D-05, D-06, D-07, D-08)
	clipboardGen int  // generation counter for clipboard auto-clear (D-06)
	clipboardHot bool // true when clipboard holds a secret
	// Git fields (D-09, D-11)
	gitRepoRoot string // root directory of the git repo (empty if not a git repo)
	// Git history overlay fields (D-13, D-14, D-15)
	history ui.HistoryModel
	// Cross-file search fields (GIT-03)
	crossFileItems     []CrossFileSearchItem // cached cross-file search items (lazily populated)
	crossFilePopulated bool                  // true after first cross-file search populates the cache
}

// NewAppModel constructs the initial AppModel.
// env provides startup validation results for the status bar indicators.
// sopsYamlPath is the path to the .sops.yaml file (may be empty if not found).
// The file list is initially empty — discovery runs asynchronously in Init.
func NewAppModel(env ui.EnvStatus, sopsYamlPath string) AppModel {
	m := AppModel{
		state:        stateFileList,
		fileList:     ui.NewFileListModel([]ui.FileItem{}, 0, 0),
		detail:       ui.NewDetailModel("", []ui.TreeNode{}, 0, 0, true, ""),
		help:         ui.NewHelpModel(0, 0),
		status:       ui.NewStatusBarModel(env),
		sopsYamlPath: sopsYamlPath,
	}
	m.status.SetBreadcrumb("files")
	m.status.SetItemCount(0, "items")
	return m
}

// Init requests the current terminal size and triggers async file discovery.
// WindowSizeMsgs are also delivered automatically when the program starts
// and on resize, but this ensures the initial size is captured immediately.
func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return tea.RequestWindowSize() },
		func() tea.Msg {
			if m.sopsYamlPath == "" {
				return FilesDiscoveredMsg{Err: nil}
			}
			files, err := sops.Discover(m.sopsYamlPath)
			return FilesDiscoveredMsg{Files: files, Err: err}
		},
	)
}

// Update dispatches incoming messages.
// Order of precedence:
//  1. WindowSizeMsg: propagate dimensions to all children
//  2. FilesDiscoveredMsg / FilesParsedMsg: async results
//  3. KeyPressMsg: global keys first (?, q, esc, i, /), then route to active child
//  4. All other messages: update status bar (for flash timer), then route to child
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		mainH := m.height - statusBarHeight(m)
		if mainH < 0 {
			mainH = 0
		}
		// Propagate to all children that need dimensions
		m.fileList.SetSize(m.width, mainH)
		m.detail.SetSize(m.width, mainH)
		m.help.SetSize(m.width, mainH)
		m.metadata.SetSize(m.width, mainH)
		m.diff.SetSize(m.width, mainH)
		m.history.SetSize(m.width, mainH)
		return m, nil

	case FilesDiscoveredMsg:
		if msg.Err != nil {
			m.status, _ = m.status.Flash("Error discovering files: " + msg.Err.Error())
			return m, nil
		}
		m.files = msg.Files
		// Invalidate cross-file search cache on new discovery (T-04-09)
		m.crossFilePopulated = false
		m.crossFileItems = nil
		items := make([]ui.FileItem, len(msg.Files))
		for i, f := range msg.Files {
			items[i] = ui.FileItem{
				Name:        f.Name,
				Path:        f.AbsPath,
				IsEncrypted: f.IsEncrypted,
				Rule:        f.Rule,
			}
		}
		mainH := m.height - statusBarHeight(m)
		if mainH < 0 {
			mainH = 0
		}
		m.fileList = ui.NewFileListModel(items, m.width, mainH)
		m.status.SetItemCount(len(items), "items")
		// Dispatch async git status fetch (D-11).
		relPaths := make([]string, len(msg.Files))
		for i, f := range msg.Files {
			relPaths[i] = f.Name
		}
		sopsDir := filepath.Dir(m.sopsYamlPath)
		gitCmd := func() tea.Msg {
			isGit := gitpkg.IsGitRepo(sopsDir)
			if !isGit {
				return GitStatusMsg{GitAvailable: false}
			}
			statuses, err := gitpkg.GetFileStatuses(sopsDir, relPaths)
			return GitStatusMsg{Statuses: statuses, GitAvailable: true, Err: err}
		}
		return m, gitCmd

	case FilesParsedMsg:
		if msg.Err != nil {
			m.status, _ = m.status.Flash("Error parsing file: " + msg.Err.Error())
			m.state = stateFileList
			return m, nil
		}
		mainH := m.height - statusBarHeight(m)
		if mainH < 0 {
			mainH = 0
		}
		m.detail = ui.NewDetailModel(
			m.currentFile.Name,
			msg.Parsed.Nodes,
			m.width,
			mainH,
			m.currentFile.IsEncrypted,
			m.currentFile.GitStatus,
		)
		m.state = stateDetail
		m.status.SetBreadcrumb("files", m.currentFileBreadcrumb())
		m.status.SetItemCount(countLeafNodes(msg.Parsed.Nodes), "keys")
		return m, nil

	case ui.RevealRequestMsg:
		// User pressed r on an encrypted leaf — dispatch async sops.DecryptKey (DEC-01).
		// T-03-04: context.WithTimeout prevents hung sops process.
		absPath := m.currentFile.AbsPath
		keyPath := msg.KeyPath
		return m, func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), sops.SopsTimeout)
			defer cancel()
			value, err := sops.DecryptKey(ctx, absPath, keyPath)
			return DecryptKeyMsg{KeyPath: keyPath, Value: value, Err: err}
		}

	case ui.RevealAllRequestMsg:
		// User pressed R with no values revealed — dispatch async sops.DecryptFile (DEC-02).
		absPath := m.currentFile.AbsPath
		return m, func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), sops.SopsTimeout)
			defer cancel()
			data, err := sops.DecryptFile(ctx, absPath)
			if err != nil {
				return DecryptAllMsg{Err: err}
			}
			values, err := parseDecryptedValues(data)
			return DecryptAllMsg{Values: values, Err: err}
		}

	case DecryptKeyMsg:
		if msg.Err != nil {
			m.status, _ = m.status.Flash("Decrypt error: " + msg.Err.Error())
			return m, nil
		}
		m.detail.RevealNode(msg.KeyPath, msg.Value)
		m.status, _ = m.status.Flash("Decrypted")
		return m, nil

	case DecryptAllMsg:
		if msg.Err != nil {
			m.status, _ = m.status.Flash("Decrypt error: " + msg.Err.Error())
			return m, nil
		}
		m.detail.RevealAllNodes(msg.Values)
		m.status, _ = m.status.Flash("All values decrypted")
		return m, nil

	case ui.EditorRequestMsg:
		// User pressed E on detail with revealed values — decrypt full file then launch $EDITOR.
		// T-03-04: context.WithTimeout prevents hung sops process.
		absPath := m.currentFile.AbsPath
		return m, func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), sops.SopsTimeout)
			defer cancel()
			decrypted, err := sops.DecryptFile(ctx, absPath)
			if err != nil {
				return EditorFinishedMsg{Err: fmt.Errorf("decrypt for editor: %w", err)}
			}
			return EditorReadyMsg{DecryptedContent: decrypted}
		}

	case EditorReadyMsg:
		// Decryption succeeded — write to temp file and launch $EDITOR via tea.ExecProcess.
		return m, launchEditor(msg.DecryptedContent)

	case EditorFinishedMsg:
		// $EDITOR exited — compare original vs edited YAML, show multi-key diff if changed.
		if msg.Err != nil {
			if msg.TmpPath != "" {
				os.Remove(msg.TmpPath) // Pitfall 3: clean up on all error paths
			}
			m.status, _ = m.status.Flash("Editor error: " + msg.Err.Error())
			return m, nil
		}
		// Read the (possibly edited) content from the temp file
		edited, err := os.ReadFile(msg.TmpPath)
		os.Remove(msg.TmpPath) // Pitfall 3: clean up immediately after read
		if err != nil {
			m.status, _ = m.status.Flash("Read error: " + err.Error())
			return m, nil
		}
		// YAML-aware comparison (Pitfall 6: byte comparison misses key reordering)
		diffs, err := compareDecryptedYAML(msg.OriginalContent, edited)
		if err != nil {
			m.status, _ = m.status.Flash("Diff error: " + err.Error())
			return m, nil
		}
		if len(diffs) == 0 {
			m.status, _ = m.status.Flash("No changes detected")
			return m, nil
		}
		// Show multi-key diff overlay
		title := fmt.Sprintf("Changes: %d keys modified", len(diffs))
		mainH := m.height - statusBarHeight(m)
		if mainH < 0 {
			mainH = 0
		}
		m.diff = ui.NewDiffModel(title, diffs, m.width, mainH)
		m.editFilePath = m.currentFile.AbsPath
		m.editorEditedContent = edited // store for EncryptFile on confirm (T-03-12: cleared after use)
		m.prevState = m.state
		m.state = stateDiff
		return m, nil

	case ui.EditConfirmMsg:
		// User confirmed edit in textinput — transition to stateDiff if values differ.
		if msg.OldValue == msg.NewValue {
			m.status, _ = m.status.Flash("No changes")
			return m, nil
		}
		mainH := m.height - statusBarHeight(m)
		if mainH < 0 {
			mainH = 0
		}
		m.diff = ui.NewDiffModel(
			"Changes: "+msg.KeyPath,
			[]ui.DiffEntry{{KeyPath: msg.KeyPath, OldValue: msg.OldValue, NewValue: msg.NewValue}},
			m.width, mainH,
		)
		m.editFilePath = m.currentFile.AbsPath
		m.prevState = m.state
		m.state = stateDiff
		return m, nil

	case ui.EditBlockedMsg:
		// User pressed e on an un-editable node — flash appropriate message.
		if msg.Reason != "" {
			m.status, _ = m.status.Flash(msg.Reason)
		} else {
			m.status, _ = m.status.Flash("Reveal first with r")
		}
		return m, nil

	case ui.EditCancelMsg:
		// User pressed Esc in edit mode — no-op (detail model already reset itself).
		return m, nil

	case ReEncryptDoneMsg:
		// Result of sops.SetKey or sops.EncryptFile subprocess call.
		if msg.Err != nil {
			m.status, _ = m.status.Flash("Re-encryption failed: " + msg.Err.Error())
		} else if m.rotateFormat != 0 {
			// Rotation completed — flash format-specific message
			m.status, _ = m.status.Flash("Rotated to " + ui.FormatLabel(m.rotateFormat))
			m.rotateFormat = 0
			// Update revealed node for single-entry rotation
			entries := m.diff.Entries()
			if len(entries) == 1 {
				m.detail.RevealNode(entries[0].KeyPath, entries[0].NewValue)
			}
		} else {
			m.status, _ = m.status.Flash("Re-encrypted")
			// Update the revealed node's DecryptedValue to the new value
			entries := m.diff.Entries()
			if len(entries) == 1 {
				m.detail.RevealNode(entries[0].KeyPath, entries[0].NewValue)
			}
		}
		m.state = stateDetail
		// After write operations, refresh git status to reflect any new uncommitted changes (D-11).
		if msg.Err == nil && m.sopsYamlPath != "" {
			relPaths := make([]string, len(m.files))
			for i, f := range m.files {
				relPaths[i] = f.Name
			}
			sopsDir := filepath.Dir(m.sopsYamlPath)
			return m, func() tea.Msg {
				statuses, err := gitpkg.GetFileStatuses(sopsDir, relPaths)
				return GitStatusMsg{Statuses: statuses, GitAvailable: true, Err: err}
			}
		}
		return m, nil

	case ui.RotateReadyMsg:
		// X key on revealed leaf with detectable format — show single-entry diff overlay.
		mainH := m.height - statusBarHeight(m)
		if mainH < 0 {
			mainH = 0
		}
		m.diff = ui.NewDiffModel(
			"Changes: "+msg.KeyPath,
			[]ui.DiffEntry{{KeyPath: msg.KeyPath, OldValue: msg.OldValue, NewValue: msg.NewValue}},
			m.width, mainH,
		)
		m.editFilePath = m.currentFile.AbsPath
		m.rotateFormat = msg.Format
		m.prevState = m.state
		m.state = stateDiff
		return m, nil

	case ui.RotateErrorMsg:
		m.status, _ = m.status.Flash("Rotation failed: " + msg.Err.Error())
		return m, nil

	case ui.RotateFormatMenuMsg:
		// X key on revealed leaf with unknown format — open format selection menu.
		m.formatMenuActive = true
		m.formatMenuKeyPath = msg.KeyPath
		m.formatMenuOldValue = msg.OldValue
		m.formatMenuCursor = 0
		m.prevState = m.state
		m.state = stateFormatMenu
		return m, nil

	case ClipboardClearMsg:
		if msg.Gen == m.clipboardGen {
			clipboard.WriteAll("") //nolint:errcheck
			m.clipboardHot = false
			m.status.SetClipboardHot(false)
		}
		return m, nil

	case GitStatusMsg:
		// Update git availability indicator in the status bar (D-12).
		env := m.status.Env()
		env.GitAvailable = msg.GitAvailable
		m.status.SetEnv(env)
		m.gitRepoRoot = filepath.Dir(m.sopsYamlPath)

		if msg.Err != nil || !msg.GitAvailable {
			return m, nil
		}
		// Propagate git statuses to cached files.
		for i := range m.files {
			if gs, ok := msg.Statuses[m.files[i].Name]; ok {
				m.files[i].GitStatus = string(gs)
			}
		}
		// Rebuild file list items with git badge data.
		items := make([]ui.FileItem, len(m.files))
		for i, f := range m.files {
			items[i] = ui.FileItem{
				Name:        f.Name,
				Path:        f.AbsPath,
				IsEncrypted: f.IsEncrypted,
				Rule:        f.Rule,
				GitStatus:   f.GitStatus,
			}
		}
		mainH := m.height - statusBarHeight(m)
		if mainH < 0 {
			mainH = 0
		}
		m.fileList = ui.NewFileListModel(items, m.width, mainH)
		m.status.SetItemCount(len(items), "items")
		return m, nil

	case GitHistoryMsg:
		if msg.Err != nil {
			m.status, _ = m.status.Flash("Git history error: " + msg.Err.Error())
			m.state = m.prevState
			return m, nil
		}
		m.history.SetEntries(msg.Entries)
		return m, nil

	case tea.KeyPressMsg:
		// stateDiff: route all keys to diff model, then check Confirmed/Cancelled.
		// This runs before global key handling so y/n/Esc are captured by the overlay.
		if m.state == stateDiff {
			m.diff, _ = m.diff.Update(msg)
			if m.diff.Confirmed() {
				entries := m.diff.Entries()
				filePath := m.editFilePath
				if len(entries) == 1 && m.editorEditedContent == nil {
					// Single key inline edit: use sops set
					entry := entries[0]
					return m, func() tea.Msg {
						ctx, cancel := context.WithTimeout(context.Background(), sops.SopsTimeout)
						defer cancel()
						err := sops.SetKey(ctx, filePath, entry.KeyPath, entry.NewValue)
						return ReEncryptDoneMsg{Err: err}
					}
				}
				// Multi-key ($EDITOR) or rotation with stored edited content: write temp, encrypt
				editedContent := m.editorEditedContent
				m.editorEditedContent = nil // T-03-12: clear reference immediately
				return m, func() tea.Msg {
					tmpFile, err := os.CreateTemp("", "sops-tui-enc-*.yaml")
					if err != nil {
						return ReEncryptDoneMsg{Err: err}
					}
					tmpPath := tmpFile.Name()
					defer os.Remove(tmpPath)
					if _, err := tmpFile.Write(editedContent); err != nil {
						tmpFile.Close()
						return ReEncryptDoneMsg{Err: err}
					}
					tmpFile.Close()
					ctx, cancel := context.WithTimeout(context.Background(), sops.SopsTimeout)
					defer cancel()
					err = sops.EncryptFile(ctx, tmpPath, filePath)
					return ReEncryptDoneMsg{Err: err}
				}
			}
			if m.diff.Cancelled() {
				m.editorEditedContent = nil // T-03-12: clear on cancel too
				m.state = m.prevState
				m.status, _ = m.status.Flash("Cancelled")
				return m, nil
			}
			return m, nil
		}

		// stateFormatMenu: handle format selection keys before global key routing.
		if m.state == stateFormatMenu {
			formats := ui.AllFormats()
			switch msg.String() {
			case "j", "down":
				if m.formatMenuCursor < len(formats)-1 {
					m.formatMenuCursor++
				}
			case "k", "up":
				if m.formatMenuCursor > 0 {
					m.formatMenuCursor--
				}
			case "enter":
				selected := formats[m.formatMenuCursor]
				m.formatMenuActive = false
				newVal, err := ui.GenerateValue(selected)
				if err != nil {
					m.state = m.prevState
					m.status, _ = m.status.Flash("Generation failed: " + err.Error())
					return m, nil
				}
				mainH := m.height - statusBarHeight(m)
				if mainH < 0 {
					mainH = 0
				}
				m.diff = ui.NewDiffModel(
					"Changes: "+m.formatMenuKeyPath,
					[]ui.DiffEntry{{KeyPath: m.formatMenuKeyPath, OldValue: m.formatMenuOldValue, NewValue: newVal}},
					m.width, mainH,
				)
				m.editFilePath = m.currentFile.AbsPath
				m.rotateFormat = selected
				m.state = stateDiff
				return m, nil
			case "esc":
				m.formatMenuActive = false
				m.state = m.prevState
				return m, nil
			}
			return m, nil
		}

		// Global key: toggle help overlay
		if key.Matches(msg, keys.DefaultGlobalKeyMap.Help) {
			if m.state == stateHelp {
				m.state = m.prevState
			} else {
				m.prevState = m.state
				m.state = stateHelp
			}
			return m, nil
		}

		// Global key: quit application
		if key.Matches(msg, keys.DefaultGlobalKeyMap.Quit) {
			return m, func() tea.Msg { return tea.Quit() }
		}

		// i key: toggle metadata overlay
		if key.Matches(msg, keys.DefaultFileListKeyMap.Info) || key.Matches(msg, keys.DefaultDetailKeyMap.Info) {
			if m.state == stateMetadata {
				m.state = m.prevState
				// Restore status bar based on prevState
				if m.prevState == stateFileList {
					m.status.SetBreadcrumb("files")
					m.status.SetItemCount(m.fileList.ItemCount(), "items")
				} else if m.prevState == stateDetail {
					m.status.SetBreadcrumb("files", m.currentFileBreadcrumb())
					m.status.SetItemCount(countLeafNodes(m.detail.Nodes()), "keys")
				}
				return m, nil
			}
			if (m.state == stateFileList && !m.fileList.IsSearchActive()) ||
				(m.state == stateDetail && !m.detail.IsSearchActive()) {
				var filePath string
				var rule sops.CreationRule
				var isEnc bool
				if m.state == stateFileList {
					if item, ok := m.fileList.SelectedFileItem(); ok {
						filePath = item.Path
						rule = item.Rule
						isEnc = item.IsEncrypted
					}
				} else {
					filePath = m.currentFile.AbsPath
					rule = m.currentFile.Rule
					isEnc = m.currentFile.IsEncrypted
				}
				if filePath != "" {
					parsed, err := parser.ParseFile(filePath, rule, isEnc)
					if err != nil {
						m.status, _ = m.status.Flash("Error reading metadata: " + err.Error())
						return m, nil
					}
					meta := ui.MetadataContent{
						Version:          parsed.Metadata.Version,
						LastModified:     parsed.Metadata.LastModified,
						MAC:              parsed.Metadata.MAC,
						AgeRecipients:    parsed.Metadata.AgeRecipients,
						EncryptedRegex:   parsed.Metadata.EncryptedRegex,
						UnencryptedRegex: parsed.Metadata.UnencryptedRegex,
					}
					mainH := m.height - statusBarHeight(m)
					if mainH < 0 {
						mainH = 0
					}
					m.metadata = ui.NewMetadataModel(meta, m.width, mainH)
					m.prevState = m.state
					m.state = stateMetadata
					m.status.SetBreadcrumb("files", m.currentFileBreadcrumb(), "metadata")
					m.status.SetItemCount(0, "")
				}
				return m, nil
			}
		}

		// / key: activate search (only when not already active)
		if msg.String() == "/" {
			if m.state == stateFileList && !m.fileList.IsSearchActive() {
				m.populateCrossFileItems()
				// Build search source from cross-file items
				titles := make([]string, len(m.crossFileItems))
				for i, item := range m.crossFileItems {
					titles[i] = item.Title()
				}
				cmd := m.fileList.ActivateCrossFileSearch(titles)
				m.status.SetBreadcrumb("files", "search")
				return m, cmd
			}
			if m.state == stateDetail && !m.detail.IsSearchActive() {
				cmd := m.detail.ActivateSearch()
				m.status.SetBreadcrumb("files", m.currentFileBreadcrumb(), "search")
				return m, cmd
			}
		}

		// Enter during cross-file search: navigate to matched file's detail view
		if m.state == stateFileList && m.fileList.IsSearchActive() && m.fileList.IsCrossFileMode() {
			if msg.String() == "enter" {
				idx := m.fileList.SelectedCrossFileIndex()
				if idx >= 0 && idx < len(m.crossFileItems) {
					item := m.crossFileItems[idx]
					m.fileList.DeactivateSearch()
					m.currentFile = sops.DiscoveredFile{
						Name:        item.FileName,
						AbsPath:     item.AbsPath,
						IsEncrypted: item.IsEnc,
						Rule:        item.Rule,
						GitStatus:   item.GitStatus,
					}
					return m, func() tea.Msg {
						parsed, err := parser.ParseFile(item.AbsPath, item.Rule, item.IsEnc)
						return FilesParsedMsg{Parsed: parsed, Err: err}
					}
				}
			}
		}

		// ctrl+y: copy revealed value to clipboard (D-01, D-02, D-03)
		if key.Matches(msg, keys.DefaultDetailKeyMap.Copy) {
			if m.state == stateDetail && !m.detail.IsSearchActive() {
				node, ok := m.detail.SelectedNode()
				if !ok {
					return m, nil
				}
				if !node.Revealed {
					m.status, _ = m.status.Flash("Reveal first with r")
					return m, nil
				}
				return m.copyToClipboard(node.DecryptedValue)
			}
		}

		// b key: toggle git history overlay (D-13, D-14)
		if key.Matches(msg, keys.DefaultDetailKeyMap.Blame) {
			if m.state == stateHistory {
				m.state = m.prevState
				m.status.SetBreadcrumb("files", m.currentFileBreadcrumb())
				m.status.SetItemCount(countLeafNodes(m.detail.Nodes()), "keys")
				return m, nil
			}
			if m.state == stateDetail && !m.detail.IsSearchActive() {
				if m.gitRepoRoot == "" {
					m.status, _ = m.status.Flash("No git repository found")
					return m, nil
				}
				mainH := m.height - statusBarHeight(m)
				if mainH < 0 {
					mainH = 0
				}
				m.history = ui.NewHistoryModel(m.currentFile.Name, m.width, mainH)
				m.prevState = m.state
				m.state = stateHistory
				m.status.SetBreadcrumb("files", m.currentFileBreadcrumb(), "history")
				// Async fetch git history
				relPath := m.currentFile.Name
				repoRoot := m.gitRepoRoot
				return m, func() tea.Msg {
					entries, err := gitpkg.GetFileHistory(repoRoot, relPath, 50)
					return GitHistoryMsg{Entries: entries, Err: err}
				}
			}
		}

		// Esc: priority chain
		if msg.String() == "esc" {
			// Priority 1: Close search if active
			if m.state == stateFileList && m.fileList.IsSearchActive() {
				m.fileList.DeactivateSearch()
				m.status.SetBreadcrumb("files")
				m.status.SetItemCount(m.fileList.ItemCount(), "items")
				return m, nil
			}
			if m.state == stateDetail && m.detail.IsSearchActive() {
				m.detail.DeactivateSearch()
				m.status.SetBreadcrumb("files", m.currentFileBreadcrumb())
				m.status.SetItemCount(countLeafNodes(m.detail.Nodes()), "keys")
				return m, nil
			}
			// Priority 2: Close overlays.
			// Note: stateDiff Esc is fully handled by the stateDiff block above (line ~481)
			// which returns early after setting Cancelled. This chain only handles
			// stateHelp and stateMetadata overlays.
			if m.state == stateHelp {
				m.state = m.prevState
				return m, nil
			}
			if m.state == stateMetadata {
				m.state = m.prevState
				return m, nil
			}
			if m.state == stateHistory {
				m.state = m.prevState
				m.status.SetBreadcrumb("files", m.currentFileBreadcrumb())
				m.status.SetItemCount(countLeafNodes(m.detail.Nodes()), "keys")
				return m, nil
			}
			// Priority 3: Navigate back from detail to file list
			// Per D-04 and T-03-02: clear all revealed values before leaving detail view.
			if m.state == stateDetail {
				m.detail.ClearAllRevealed()
				m.state = stateFileList
				m.status.SetBreadcrumb("files")
				m.status.SetItemCount(m.fileList.ItemCount(), "items")
				return m, nil
			}
		}

		// Enter/l in file list: parse file and drill into detail (async)
		if m.state == stateFileList && !m.fileList.IsSearchActive() {
			if msg.String() == "enter" || msg.String() == "l" {
				if item, ok := m.fileList.SelectedFileItem(); ok {
					m.currentFile = sops.DiscoveredFile{
						Name:        item.Name,
						AbsPath:     item.Path,
						IsEncrypted: item.IsEncrypted,
						Rule:        item.Rule,
						GitStatus:   item.GitStatus,
					}
					return m, func() tea.Msg {
						parsed, err := parser.ParseFile(item.Path, item.Rule, item.IsEncrypted)
						return FilesParsedMsg{Parsed: parsed, Err: err}
					}
				}
			}
		}
	}

	// Always update status bar so flash timer messages are processed
	var statusCmd tea.Cmd
	m.status, statusCmd = m.status.Update(msg)

	// Route remaining messages to active child model
	switch m.state {
	case stateFileList:
		var cmd tea.Cmd
		m.fileList, cmd = m.fileList.Update(msg)
		return m, tea.Batch(cmd, statusCmd)
	case stateDetail:
		var cmd tea.Cmd
		m.detail, cmd = m.detail.Update(msg)
		return m, tea.Batch(cmd, statusCmd)
	case stateHelp:
		// Help is display-only; no child update needed
		return m, statusCmd
	case stateMetadata:
		// Handle j/k scrolling in metadata overlay
		if kMsg, ok := msg.(tea.KeyPressMsg); ok {
			if key.Matches(kMsg, keys.DefaultDetailKeyMap.Down) {
				m.metadata.ScrollDown()
			} else if key.Matches(kMsg, keys.DefaultDetailKeyMap.Up) {
				m.metadata.ScrollUp()
			}
		}
		return m, statusCmd
	case stateHistory:
		// Handle j/k scrolling in history overlay
		if kMsg, ok := msg.(tea.KeyPressMsg); ok {
			if key.Matches(kMsg, keys.DefaultDetailKeyMap.Down) {
				m.history.ScrollDown()
			} else if key.Matches(kMsg, keys.DefaultDetailKeyMap.Up) {
				m.history.ScrollUp()
			}
		}
		return m, statusCmd
	}

	return m, statusCmd
}

// View composes the terminal frame: main content stacked above status bar.
// AltScreen is always true per the design decision for a full-screen TUI.
func (m AppModel) View() tea.View {
	// Render main content based on active state
	statusBar := m.status.View(m.width)
	statusBarH := lipgloss.Height(statusBar)
	mainH := m.height - statusBarH
	if mainH < 0 {
		mainH = 0
	}

	var content string
	switch m.state {
	case stateFileList:
		content = m.fileList.View()
	case stateDetail:
		content = m.detail.View()
	case stateHelp:
		// Contextual: use prevState so the overlay knows which keybindings to show
		content = m.help.View(ui.ViewState(m.prevState))
	case stateMetadata:
		content = m.metadata.View()
	case stateDiff:
		content = m.diff.View()
	case stateFormatMenu:
		content = renderFormatMenu(m.formatMenuCursor, m.width, mainH)
	case stateHistory:
		content = m.history.View()
	}

	// Pad or constrain main content to fill available height
	body := lipgloss.NewStyle().Height(mainH).Render(content)

	// Stack body above status bar
	full := lipgloss.JoinVertical(lipgloss.Left, body, statusBar)

	v := tea.NewView(full)
	v.AltScreen = true
	return v
}

// IsClipboardHot returns true when the clipboard currently holds a secret.
// Exported for use in tests to verify clipboard state without inspecting View output.
func (m AppModel) IsClipboardHot() bool {
	return m.clipboardHot
}

// clipboardTimeout returns the clipboard auto-clear duration.
// Default 30s, overridden by SOPS_TUI_CLIPBOARD_TIMEOUT env var (D-05).
func clipboardTimeout() time.Duration {
	if s := os.Getenv("SOPS_TUI_CLIPBOARD_TIMEOUT"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return 30 * time.Second
}

// ClipboardTimeout is the exported wrapper for clipboardTimeout.
// Used by tests to verify timeout configuration (D-05).
var ClipboardTimeout = clipboardTimeout

// copyToClipboard writes value to the system clipboard, increments the generation
// counter, sets clipboardHot=true, flashes "Copied (clears in 30s)", and schedules
// a ClipboardClearMsg via tea.Tick (D-01, D-04, D-06).
func (m AppModel) copyToClipboard(value string) (AppModel, tea.Cmd) {
	if clipboard.Unsupported {
		var statusCmd tea.Cmd
		m.status, statusCmd = m.status.Flash("Clipboard not available (install xclip or wl-clipboard)")
		return m, statusCmd
	}
	if err := clipboard.WriteAll(value); err != nil {
		var statusCmd tea.Cmd
		m.status, statusCmd = m.status.Flash("Clipboard error: " + err.Error())
		return m, statusCmd
	}
	m.clipboardGen++
	gen := m.clipboardGen
	m.clipboardHot = true
	m.status.SetClipboardHot(true)
	var statusCmd tea.Cmd
	m.status, statusCmd = m.status.Flash("Copied (clears in 30s)")
	timeout := clipboardTimeout()
	clearCmd := tea.Tick(timeout, func(_ time.Time) tea.Msg {
		return ClipboardClearMsg{Gen: gen}
	})
	return m, tea.Batch(statusCmd, clearCmd)
}

// currentFileBreadcrumb returns m.currentFile.Name with a git badge suffix
// appended when the file has uncommitted changes. Used in SetBreadcrumb calls
// so the detail view breadcrumb reflects git status (D-09, D-10).
func (m AppModel) currentFileBreadcrumb() string {
	name := m.currentFile.Name
	switch m.currentFile.GitStatus {
	case "M":
		name += " [M]"
	case "A":
		name += " [A]"
	case "?":
		name += " [?]"
	}
	return name
}

// statusBarHeight returns the rendered height of the status bar in terminal rows.
func statusBarHeight(m AppModel) int {
	statusBar := m.status.View(m.width)
	return lipgloss.Height(statusBar)
}

// populateCrossFileItems lazily populates the cross-file search item cache (GIT-03).
// It collects file-level items and key-path items for all discovered files.
// T-04-09: lazy population (only on first / press); cached after first use;
// invalidated only on new file discovery via FilesDiscoveredMsg.
func (m *AppModel) populateCrossFileItems() {
	if m.crossFilePopulated {
		return
	}
	var items []CrossFileSearchItem
	for _, f := range m.files {
		// Add file-level item
		items = append(items, CrossFileSearchItem{
			FileName:  f.Name,
			AbsPath:   f.AbsPath,
			Rule:      f.Rule,
			IsEnc:     f.IsEncrypted,
			GitStatus: f.GitStatus,
		})
		// Parse key paths without decryption
		parsed, err := parser.ParseFile(f.AbsPath, f.Rule, f.IsEncrypted)
		if err != nil {
			continue
		}
		keyPaths := collectKeyPaths(parsed.Nodes, "")
		for _, kp := range keyPaths {
			items = append(items, CrossFileSearchItem{
				FileName:  f.Name,
				KeyPath:   kp,
				AbsPath:   f.AbsPath,
				Rule:      f.Rule,
				IsEnc:     f.IsEncrypted,
				GitStatus: f.GitStatus,
			})
		}
	}
	m.crossFileItems = items
	m.crossFilePopulated = true
}

// collectKeyPaths recursively extracts dot-joined key paths from tree nodes.
func collectKeyPaths(nodes []ui.TreeNode, prefix string) []string {
	var paths []string
	for _, n := range nodes {
		path := prefix
		if path != "" {
			path += "."
		}
		path += n.Key
		paths = append(paths, path)
		if len(n.Children) > 0 {
			paths = append(paths, collectKeyPaths(n.Children, path)...)
		}
	}
	return paths
}

// countLeafNodes recursively counts all leaf nodes (nodes with no children).
func countLeafNodes(nodes []ui.TreeNode) int {
	count := 0
	for _, n := range nodes {
		if len(n.Children) == 0 {
			count++
		} else {
			count += countLeafNodes(n.Children)
		}
	}
	return count
}

// parseDecryptedValues parses a fully-decrypted YAML byte slice (output of sops decrypt)
// and returns a flat map of dot-joined key path → string value for all leaf nodes.
// Used by the RevealAllRequestMsg handler to populate DecryptAllMsg.Values.
func parseDecryptedValues(data []byte) (map[string]string, error) {
	var root yaml.MapSlice
	if err := yaml.UnmarshalWithOptions(data, &root, yaml.UseOrderedMap()); err != nil {
		return nil, err
	}
	result := make(map[string]string)
	collectLeafValues(root, "", result)
	return result, nil
}

// collectLeafValues walks a yaml.MapSlice recursively, building dot-joined key paths
// and collecting string representations of leaf values into the result map.
func collectLeafValues(slice yaml.MapSlice, parentPath string, result map[string]string) {
	for _, item := range slice {
		key, ok := item.Key.(string)
		if !ok {
			continue
		}
		// Skip the sops metadata block — it is not a secret value
		if parentPath == "" && key == "sops" {
			continue
		}
		path := key
		if parentPath != "" {
			path = parentPath + "." + key
		}
		switch v := item.Value.(type) {
		case yaml.MapSlice:
			collectLeafValues(v, path, result)
		case string:
			result[path] = v
		case int, int64, float64, bool:
			result[path] = fmt.Sprintf("%v", v)
		default:
			// Skip non-scalar, non-map values (arrays, nulls, etc.)
		}
	}
}

// launchEditor writes decryptedContent to a 0600 temp file and suspends the TUI
// via tea.ExecProcess to launch the user's $EDITOR (or $VISUAL, or "vi" fallback).
// The ExecProcess callback returns EditorFinishedMsg with the temp file path and
// original content for YAML-aware diff comparison.
// Per T-03-10: temp file created with Chmod(0600) to restrict read access.
// Per Pitfall 3: temp file path is stored in EditorFinishedMsg so the handler
// can clean it up on all paths (including error paths).
func launchEditor(decryptedContent []byte) tea.Cmd {
	tmpFile, err := os.CreateTemp("", "sops-tui-*.yaml")
	if err != nil {
		return func() tea.Msg { return EditorFinishedMsg{Err: err} }
	}
	tmpPath := tmpFile.Name()
	// T-03-10: restrict permissions to owner read/write only
	if err := tmpFile.Chmod(0600); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return func() tea.Msg { return EditorFinishedMsg{Err: err} }
	}
	if _, err := tmpFile.Write(decryptedContent); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return func() tea.Msg { return EditorFinishedMsg{Err: err} }
	}
	tmpFile.Close()

	// Determine editor: $EDITOR → $VISUAL → "vi"
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}

	cmd := exec.Command(editor, tmpPath)

	// Copy original content for diff comparison after editing
	originalCopy := make([]byte, len(decryptedContent))
	copy(originalCopy, decryptedContent)

	return tea.ExecProcess(cmd, func(execErr error) tea.Msg {
		return EditorFinishedMsg{
			TmpPath:         tmpPath,
			OriginalContent: originalCopy,
			Err:             execErr,
		}
	})
}

// compareDecryptedYAML compares two decrypted YAML byte slices and returns a slice
// of DiffEntry for keys whose values differ. Key ordering is ignored (Pitfall 6).
// Skips the "sops" metadata block. Returns an empty slice if content is identical.
// Exported as CompareDecryptedYAML for use in tests.
func compareDecryptedYAML(original, edited []byte) ([]ui.DiffEntry, error) {
	origMap, err := flattenYAMLToMap(original)
	if err != nil {
		return nil, fmt.Errorf("parse original: %w", err)
	}
	editMap, err := flattenYAMLToMap(edited)
	if err != nil {
		return nil, fmt.Errorf("parse edited: %w", err)
	}
	var diffs []ui.DiffEntry
	for k, origVal := range origMap {
		editVal, exists := editMap[k]
		if !exists {
			diffs = append(diffs, ui.DiffEntry{KeyPath: k, OldValue: origVal, NewValue: "(deleted)"})
		} else if origVal != editVal {
			diffs = append(diffs, ui.DiffEntry{KeyPath: k, OldValue: origVal, NewValue: editVal})
		}
	}
	for k, editVal := range editMap {
		if _, exists := origMap[k]; !exists {
			diffs = append(diffs, ui.DiffEntry{KeyPath: k, OldValue: "(new)", NewValue: editVal})
		}
	}
	// Sort by KeyPath for stable, deterministic ordering
	sort.Slice(diffs, func(i, j int) bool { return diffs[i].KeyPath < diffs[j].KeyPath })
	return diffs, nil
}

// CompareDecryptedYAML is the exported wrapper for compareDecryptedYAML.
// Used by tests to verify YAML-aware diff comparison.
func CompareDecryptedYAML(original, edited []byte) ([]ui.DiffEntry, error) {
	return compareDecryptedYAML(original, edited)
}

// flattenYAMLToMap parses a YAML byte slice into a flat map of dot-joined key path
// to string value. Uses ordered map parsing to preserve key order for walkMapSlice.
func flattenYAMLToMap(data []byte) (map[string]string, error) {
	var root yaml.MapSlice
	if err := yaml.UnmarshalWithOptions(data, &root, yaml.UseOrderedMap()); err != nil {
		return nil, err
	}
	result := make(map[string]string)
	walkMapSlice(root, "", result)
	return result, nil
}

// walkMapSlice recursively walks a yaml.MapSlice, collecting leaf values into out
// with dot-joined key paths. Skips the "sops" metadata block at the top level.
func walkMapSlice(ms yaml.MapSlice, prefix string, out map[string]string) {
	for _, item := range ms {
		key := fmt.Sprintf("%v", item.Key)
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		if prefix == "" && key == "sops" {
			continue // skip sops metadata block
		}
		switch v := item.Value.(type) {
		case yaml.MapSlice:
			walkMapSlice(v, path, out)
		default:
			out[path] = fmt.Sprintf("%v", v)
		}
	}
}

// renderFormatMenu renders the format selection menu overlay for ambiguous rotation.
// Per 03-UI-SPEC.md §Format Menu: rounded border, surface background, 5 format options.
func renderFormatMenu(cursor, width, height int) string {
	title := ui.DiffKeyStyle.Render("Select format for new secret:")
	var sb strings.Builder
	formats := ui.AllFormats()
	for i, f := range formats {
		label := ui.FormatLabel(f)
		if i == cursor {
			sb.WriteString(ui.SelectedRow.Width(width - 8).Render("> " + label))
		} else {
			sb.WriteString(ui.FormatMenuStyle.Render("  " + label))
		}
		sb.WriteByte('\n')
	}
	footer := ui.DimText.Render("j/k to select, Enter to confirm, Esc to cancel")
	inner := title + "\n\n" + sb.String() + "\n" + footer

	boxWidth := width - 2
	if boxWidth < 1 {
		boxWidth = 1
	}
	boxHeight := height - 2
	if boxHeight < 1 {
		boxHeight = 1
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorMuted).
		Background(ui.ColorSurface).
		Padding(1, ui.SpaceMD).
		Width(boxWidth).
		Height(boxHeight).
		Render(inner)
}
