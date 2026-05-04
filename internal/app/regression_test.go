package app_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caesarakalaeii/sops-tui/internal/app"
	"github.com/caesarakalaeii/sops-tui/internal/sops"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

// Phase 11 D-507 chrome-interaction sanity tests. These are NOT teatest tests
// (the project uses Update-loop pattern; teatest is absent from go.mod). Each
// test targets ONE chrome-interaction risk specific to the v1.1 rework:
//
//   - TestRegression_ClipboardAutoClearWithChrome catches a typed flash API
//     (Phase 10 D-411) regression where [W]/[E] prefixes leak into a
//     non-warn/non-err path or where the clipboard timeout doesn't reach the
//     chrome's [clip] indicator.
//   - TestRegression_RecipientFormMenuHints catches a Phase 9 menuHints
//     dispatcher regression for the nested recipient form state — the menu
//     must show form-level hints (Enter/Esc), NOT file-list ones (j/k/q).
//   - TestRegression_HealthOverlayOnNarrowWidth catches a Phase 10 D-425
//     narrow-terminal first/last-segment regression intersecting with the
//     health overlay path.
//
// These three tests do NOT extend coverage of the 9 v1.0 capabilities (see
// CONTEXT.md D-508 — coverage gaps file follow-up issues, ship anyway). They
// target the chrome-interaction surface specifically, where the v1.1 rework
// most plausibly introduced a regression.

// TestRegression_ClipboardAutoClearWithChrome verifies that copying a value,
// clearing the flash, and clearing the clipboard via ClipboardClearMsg leaves
// the chrome [clip] indicator gone AND no [W]/[E] flash prefix leaked into
// the next-frame status bar (Phase 10 D-411 typed flash API regression
// surface).
func TestRegression_ClipboardAutoClearWithChrome(t *testing.T) {
	nodes := []ui.TreeNode{
		{Key: "token", Encrypted: true, TypeHint: "str", Revealed: true, DecryptedValue: "supersecret"},
	}
	m := setupDetailWithNodes(t, nodes)

	// Copy with ctrl+y — sets clipboardHot=true, flashes "Copied", schedules
	// ClipboardClearMsg via tea.Tick (D-06).
	updated, cmd := m.Update(tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl})
	require.NotNil(t, cmd, "ctrl+y on revealed leaf must return a non-nil Cmd")

	am := asAppModel(t, updated)
	require.True(t, am.IsClipboardHot(), "after ctrl+y, clipboardHot must be true")

	// Clear flash with FlashClearMsg{Gen: 1} (matching the first flash).
	updated2, _ := updated.Update(ui.FlashClearMsg{Gen: 1})

	// Clear clipboard with ClipboardClearMsg{Gen: 1} (matching gen).
	updated3, _ := updated2.Update(app.ClipboardClearMsg{Gen: 1})
	am3 := asAppModel(t, updated3)
	assert.False(t, am3.IsClipboardHot(),
		"ClipboardClearMsg with matching gen must clear clipboardHot")

	// Now read the rendered chrome.
	v := updated3.View()
	stripped := ansi.Strip(v.Content)

	// Chrome must NOT show [clip] indicator after clipboard cleared.
	assert.NotContains(t, stripped, "[clip]",
		"after ClipboardClearMsg, [clip] indicator must be gone from chrome")

	// Phase 10 D-411 typed flash API regression surface: a non-warn / non-err
	// flash path must NOT leak [W] or [E] prefix substrings. The Copied flash
	// is FlashInfo (no prefix); FlashWarn would prefix [W]; FlashErr [E].
	// After FlashClearMsg+ClipboardClearMsg, no flash is active — neither
	// prefix should appear anywhere in the stripped chrome.
	assert.NotContains(t, stripped, "[W]",
		"no [W] severity prefix should appear after Copied flash + ClipboardClearMsg")
	assert.NotContains(t, stripped, "[E]",
		"no [E] severity prefix should appear after Copied flash + ClipboardClearMsg")
}

// TestRegression_RecipientFormMenuHints verifies that the persistent menu
// shows form-level hints (Enter/Esc) when AppModel is in stateRecipientForm,
// NOT file-list hints (j/k/q). Catches a Phase 9 menuHints dispatcher
// regression for nested form states.
func TestRegression_RecipientFormMenuHints(t *testing.T) {
	// Drive into stateDetail with a revealed leaf (existing helper).
	m := modelInDetailWithRevealedLeaf(t)

	// Press 'a' — opens add-recipient form (model.go:1239-1248). The handler
	// returns a non-nil cmd (recipientForm.Activate() focus cmd).
	m2, cmd := m.Update(tea.KeyPressMsg{Code: 'a'})
	require.NotNil(t, cmd, "a-key in stateDetail must return a non-nil cmd (form focus)")

	v := m2.View()
	stripped := ansi.Strip(v.Content)

	// Form-level hint mnemonics derived from DefaultRecipientFormKeyMap
	// (internal/keys/bindings.go:624-633): WithHelp("Enter", "confirm") and
	// WithHelp("Esc", "cancel"). Menu cells are rendered as "[Enter] confirm"
	// and "[Esc] cancel" — the bracketed mnemonics are unique to the menu.
	assert.Contains(t, stripped, "[Enter]",
		"stateRecipientForm menu must show [Enter] mnemonic from DefaultRecipientFormKeyMap.Confirm")
	assert.Contains(t, stripped, "[Esc]",
		"stateRecipientForm menu must show [Esc] mnemonic from DefaultRecipientFormKeyMap.Cancel")

	// File-list hint mnemonics MUST NOT appear in menu rows. The menu renders
	// cells as "[mnemonic] description"; the [j] / [k] / [q] substrings are
	// unique to the menu (descriptions don't bracket their text). Search for
	// the bracketed forms specifically to avoid false positives on body text
	// (e.g. the form prompt "Enter age public key:" naturally contains 'k').
	assert.NotContains(t, stripped, "[j]",
		"stateRecipientForm menu must not include FileList [j] hint")
	assert.NotContains(t, stripped, "[k]",
		"stateRecipientForm menu must not include FileList [k] hint")
}

// TestRegression_HealthOverlayOnNarrowWidth verifies that health overlay
// content is reachable at narrow widths (80×24 + 60×24) AND the active
// "health" crumb segment survives the Phase 10 D-425 first/last-preservation
// rule even after segment-width truncation. Two sub-tests, one per width.
func TestRegression_HealthOverlayOnNarrowWidth(t *testing.T) {
	for _, tc := range []struct {
		name   string
		width  int
		height int
	}{
		{"80x24", 80, 24},
		{"60x24", 60, 24},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := app.NewAppModel(defaultEnv(), "", colorprofile.TrueColor)
			m2 := send(t, m, tea.WindowSizeMsg{Width: tc.width, Height: tc.height})

			// Provide at least one file so H key passes the empty-files guard
			// (model.go:1382-1385 returns "No files to scan" flash on empty).
			files := []sops.DiscoveredFile{
				{Name: "secrets/prod.yaml", AbsPath: "/repo/secrets/prod.yaml", IsEncrypted: true},
			}
			m3 := send(t, m2, app.FilesDiscoveredMsg{Files: files})

			// Press H to enter the confirmation gate (DiffModel as a confirm
			// gate before decrypting all files — model.go:1379-1402).
			m4, _ := m3.Update(tea.KeyPressMsg{Code: 'H'})

			// Press y to confirm — dispatches async health scan and enters
			// stateHealth (model.go ~973 sets crumb to "files", "health").
			m5, _ := m4.Update(tea.KeyPressMsg{Code: 'y'})

			// Inject empty-result HealthCheckResultMsg so the overlay paints
			// the "No issues found" empty state (internal/ui/health.go:152-156).
			m6 := send(t, m5, app.HealthCheckResultMsg{})

			v := m6.View()
			stripped := ansi.Strip(v.Content)

			// Health overlay content must be reachable. Empty-result path
			// renders "Secret Health Check" title + "No issues found" /
			// "All secrets passed health checks." (internal/ui/health.go:141,
			// 154-155). At narrow widths, WrapTitled may truncate but at
			// least one of these strings must remain visible.
			anyHealthString :=
				strings.Contains(stripped, "Secret Health Check") ||
					strings.Contains(stripped, "No issues found") ||
					strings.Contains(stripped, "All secrets passed")
			assert.True(t, anyHealthString,
				"health overlay must render at width=%d: expected one of "+
					"\"Secret Health Check\" / \"No issues found\" / \"All secrets passed\", "+
					"got stripped content (first 400 chars): %q",
				tc.width, truncateForLog(stripped, 400))

			// Phase 10 D-425 first/last-preservation: the active "health"
			// crumb is the last segment after stateHealth + breadcrumb
			// "files", "health" (model.go:973). Even at width=60 with mid-
			// segment ellipsis, the last segment must remain visible.
			assert.Contains(t, stripped, "health",
				"active health crumb must remain visible at width=%d "+
					"(D-425 first/last preservation contract)", tc.width)
		})
	}
}

// truncateForLog returns at most maxLen bytes from s with an ellipsis suffix
// on truncation. Used in failure messages so we don't dump 4 KB of stripped
// chrome on every assertion fail.
func truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
