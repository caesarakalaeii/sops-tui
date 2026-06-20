// Package app — Phase 9 drift detector + golden matrix.
//
// Per D-305: runtime equality between model.Hints() and the keymap's
// HintsFromBindings(km.ShortHelp()) (modulo HiddenFromMenu() suppression).
// Per D-308: 13-entry golden matrix locks rendered RenderMenu output per
// (state, IsSearchActive) tuple. Per T-09-01: PII grep-gate.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package app

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"github.com/charmbracelet/colorprofile"
	"github.com/stretchr/testify/require"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
	"github.com/caesarakalaeii/sops-tui/internal/testutil"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

// expectedHintsWithSuppression derives the expected MenuHint slice from a keymap,
// applying any HiddenFromMenu() override discovered via concrete interface assertion.
// The local interface declaration mirrors the unexported menuVisibilityOverrider in
// internal/keys/bindings.go — Go satisfies the interface structurally so no export is needed.
func expectedHintsWithSuppression(km help.KeyMap) []keys.MenuHint {
	expected := keys.HintsFromBindings(km.ShortHelp())
	type overrider interface {
		HiddenFromMenu() []key.Binding
	}
	if ov, ok := km.(overrider); ok {
		for _, hidden := range ov.HiddenFromMenu() {
			for i := range expected {
				if expected[i].Mnemonic == hidden.Help().Key {
					expected[i].Visible = false
					break
				}
			}
		}
	}
	return expected
}

// TestMenuHints_Drift asserts runtime equality between every dispatcher
// branch's output and the keymap-derived expectation. 13 sub-tests cover
// every branch of menuHints() per D-308.
func TestMenuHints_Drift(t *testing.T) {
	t.Run("stateFileList", func(t *testing.T) {
		m := buildAppModel(t)
		// m.state defaults to stateFileList; IsSearchActive defaults false.
		require.Equal(t,
			expectedHintsWithSuppression(keys.DefaultFileListKeyMap),
			m.menuHints())
	})

	t.Run("stateFileList_search", func(t *testing.T) {
		m := buildAppModel(t)
		// Activate search via the same two-tier strategy as TestMenuHints_StateFileList_SearchActive.
		updated, _ := m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
		m = updated.(AppModel)
		if !m.fileList.IsSearchActive() {
			fl := m.fileList
			_ = (&fl).ActivateSearch()
			m.fileList = fl
		}
		require.True(t, m.fileList.IsSearchActive())
		require.Equal(t,
			expectedHintsWithSuppression(keys.DefaultFileListSearchKeyMap),
			m.menuHints())
	})

	t.Run("stateDetail", func(t *testing.T) {
		m := buildAppModel(t)
		m.state = stateDetail
		require.Equal(t,
			expectedHintsWithSuppression(keys.DefaultDetailKeyMap),
			m.menuHints())
	})

	t.Run("stateMetadata", func(t *testing.T) {
		m := buildAppModel(t)
		m.state = stateMetadata
		require.Equal(t,
			expectedHintsWithSuppression(keys.DefaultMetadataKeyMap),
			m.menuHints())
	})

	t.Run("stateDiff", func(t *testing.T) {
		m := buildAppModel(t)
		m.state = stateDiff
		require.Equal(t,
			expectedHintsWithSuppression(keys.DefaultDiffKeyMap),
			m.menuHints())
	})

	t.Run("stateRecipientConfirm", func(t *testing.T) {
		m := buildAppModel(t)
		m.state = stateRecipientConfirm
		require.Equal(t,
			expectedHintsWithSuppression(keys.DefaultRecipientConfirmKeyMap),
			m.menuHints())
	})

	t.Run("stateBulkReKeyConfirm", func(t *testing.T) {
		m := buildAppModel(t)
		m.state = stateBulkReKeyConfirm
		require.Equal(t,
			expectedHintsWithSuppression(keys.DefaultBulkReKeyConfirmKeyMap),
			m.menuHints())
	})

	t.Run("stateHelp", func(t *testing.T) {
		m := buildAppModel(t)
		m.state = stateHelp
		require.Equal(t,
			expectedHintsWithSuppression(keys.DefaultHelpKeyMap),
			m.menuHints())
	})

	t.Run("stateHistory", func(t *testing.T) {
		m := buildAppModel(t)
		m.state = stateHistory
		require.Equal(t,
			expectedHintsWithSuppression(keys.DefaultHistoryKeyMap),
			m.menuHints())
	})

	t.Run("stateHealth", func(t *testing.T) {
		m := buildAppModel(t)
		m.state = stateHealth
		require.Equal(t,
			expectedHintsWithSuppression(keys.DefaultHealthKeyMap),
			m.menuHints())
	})

	t.Run("stateRecipientForm", func(t *testing.T) {
		m := buildAppModel(t)
		m.state = stateRecipientForm
		require.Equal(t,
			expectedHintsWithSuppression(keys.DefaultRecipientFormKeyMap),
			m.menuHints())
	})

	t.Run("stateRecipientList", func(t *testing.T) {
		m := buildAppModel(t)
		m.state = stateRecipientList
		require.Equal(t,
			expectedHintsWithSuppression(keys.DefaultRecipientListKeyMap),
			m.menuHints())
	})

	t.Run("stateFormatMenu", func(t *testing.T) {
		m := buildAppModel(t)
		m.state = stateFormatMenu
		require.Equal(t,
			expectedHintsWithSuppression(keys.DefaultFormatMenuKeyMap),
			m.menuHints())
	})

	t.Run("stateAddSecretForm", func(t *testing.T) {
		m := buildAppModel(t)
		m.state = stateAddSecretForm
		require.Equal(t,
			expectedHintsWithSuppression(keys.DefaultAddSecretFormKeyMap),
			m.menuHints())
	})
}

// TestMenuGolden locks the rendered persistent menu per (state, IsSearchActive)
// tuple. 13 sub-tests, one golden file per state. RequireGoldenStructure
// strips ANSI so Phase 10's palette pass does not churn (D-311).
//
// Generate goldens initially with: GOLDEN_UPDATE=1 go test ./internal/app/... -run TestMenuGolden
func TestMenuGolden(t *testing.T) {
	const width = 80

	run := func(name string, setup func(m *AppModel)) {
		t.Run(name, func(t *testing.T) {
			m := buildAppModel(t)
			if setup != nil {
				setup(&m)
			}
			hints := m.menuHints()
			rendered := ui.RenderMenu(hints, ui.PaletteFor(colorprofile.TrueColor), width)
			testutil.RequireGoldenStructure(t, "menu_"+name, rendered)
		})
	}

	run("file_list", nil) // default state
	run("file_list_search", func(m *AppModel) {
		updated, _ := m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
		*m = updated.(AppModel)
		if !m.fileList.IsSearchActive() {
			fl := m.fileList
			_ = (&fl).ActivateSearch()
			m.fileList = fl
		}
		require.True(t, m.fileList.IsSearchActive())
	})
	run("detail", func(m *AppModel) { m.state = stateDetail })
	run("metadata", func(m *AppModel) { m.state = stateMetadata })
	run("diff", func(m *AppModel) { m.state = stateDiff })
	run("recipient_confirm", func(m *AppModel) { m.state = stateRecipientConfirm })
	run("bulk_re_key_confirm", func(m *AppModel) { m.state = stateBulkReKeyConfirm })
	run("help", func(m *AppModel) { m.state = stateHelp })
	run("history", func(m *AppModel) { m.state = stateHistory })
	run("health", func(m *AppModel) { m.state = stateHealth })
	run("recipient_form", func(m *AppModel) { m.state = stateRecipientForm })
	run("recipient_list", func(m *AppModel) { m.state = stateRecipientList })
	run("format_menu", func(m *AppModel) { m.state = stateFormatMenu })
	run("add_secret_form", func(m *AppModel) { m.state = stateAddSecretForm })
}

// TestMenuGoldenNoPII enforces threat T-09-01: no golden file may capture
// PII (host paths, age private key markers, .sops.yaml strings, ssh-rsa keys).
// RenderMenu input is binding metadata only — by construction these markers
// cannot reach the golden, but if a future regression captures v.Content
// instead of RenderMenu output, this test fails.
func TestMenuGoldenNoPII(t *testing.T) {
	// PII patterns: absolute home paths, age private key block, sops config name,
	// ssh-rsa keys, AGE-SECRET-KEY prefix.
	forbidden := regexp.MustCompile(`(?m)/Users/|/home/|BEGIN AGE|\.sops\.yaml|ssh-rsa |AGE-SECRET-KEY`)

	matches, err := filepath.Glob(filepath.Join("testdata", "menu_*.golden"))
	require.NoError(t, err)
	require.NotEmpty(t, matches, "expected at least 13 menu_*.golden files; run GOLDEN_UPDATE=1 first")

	for _, path := range matches {
		content, err := os.ReadFile(path)
		require.NoError(t, err, "reading %s", path)
		require.False(t, forbidden.Match(content),
			"T-09-01: %s contains PII markers (host paths, key blocks, .sops.yaml) — must capture RenderMenu output only, not full chrome", path)
	}
}
