// Package app — Phase 10 Plan 3: 4-profile teatest matrix (D-423).
//
// These tests force each of Ascii / ANSI / ANSI256 / TrueColor and capture
// color-bearing goldens for the four representative chrome scenes:
// full-tier chrome at 200x60, the crumbs row with an active chip, the menu
// grid with one populated state, and the flash bar at each of three
// severities (Info / Warn / Err). Total: 24 sub-tests producing 24 goldens.
//
// CRITICAL: These tests MUST NOT call t.Parallel() at any level (top-level
// OR sub-test). The tests mutate lipgloss.Writer.Profile globally and
// concurrent mutation would corrupt cross-test state. Save+restore via defer
// guarantees isolation across the 4 profile sub-tests within a single Go
// test invocation, but `go test -p N` (parallel package execution) and
// t.Parallel() must not stack additional concurrency on top.
//
// Other tests in this package can still call t.Parallel() — the constraint
// is local to this file.
//
// Implementation note (Rule 1 deviation from plan's approach): mutating
// lipgloss.Writer.Profile alone does NOT change what Style.Render() returns
// — Render produces full-fidelity SGR in-memory, and only the
// colorprofile.Writer downsamples on Write. To capture per-profile SGR
// goldens, the test pipes the rendered output through a fresh
// colorprofile.Writer with the target Profile (downsampleForProfile helper).
// We also save+restore lipgloss.Writer.Profile per sub-test to mirror the
// production seam (cmd/sops-tui/main.go passes tea.WithColorProfile so the
// global is set at startup); this keeps the test isolation contract honest
// even though the Render path is captured via the explicit downsampler.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package app_test

import (
	"bytes"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
	"github.com/caesarakalaeii/sops-tui/internal/testutil"
	"github.com/caesarakalaeii/sops-tui/internal/ui"
)

// profileCases is the canonical 4-profile matrix. Order: Ascii (lowest)
// -> TrueColor (highest). Sub-test names match the suffix of the
// .golden file.
var profileCases = []struct {
	name    string
	profile colorprofile.Profile
}{
	{"ascii", colorprofile.Ascii},
	{"ansi", colorprofile.ANSI},
	{"ansi256", colorprofile.ANSI256},
	{"truecolor", colorprofile.TrueColor},
}

// withProfile saves lipgloss.Writer.Profile, sets it to p, and returns a
// restore func to be deferred. Mirrors the production seam (main.go sets
// the global once at startup) so the test contract is "global is set
// during the body of one sub-test, restored before the next".
func withProfile(t *testing.T, p colorprofile.Profile) func() {
	t.Helper()
	saved := lipgloss.Writer.Profile
	lipgloss.Writer.Profile = p
	return func() {
		lipgloss.Writer.Profile = saved
	}
}

// downsampleForProfile pipes the given full-fidelity rendered string
// through a colorprofile.Writer with profile p and returns the
// downsampled output. Required because Style.Render() always emits the
// full-fidelity SGR (24-bit) and only the Writer transforms colors for
// the active profile. This is the in-memory equivalent of the production
// stdout path (where lipgloss.Writer wraps os.Stdout with the detected
// profile).
func downsampleForProfile(p colorprofile.Profile, rendered string) string {
	var buf bytes.Buffer
	w := colorprofile.Writer{Forward: &buf, Profile: p}
	if _, err := w.WriteString(rendered); err != nil {
		// Should be unreachable; bytes.Buffer never errors.
		panic(err)
	}
	return buf.String()
}

// fixtureInfoPanel is the deterministic InfoPanelData used by the chrome
// matrix sub-tests so cross-profile goldens differ ONLY in SGR bytes.
func fixtureInfoPanel() ui.InfoPanelData {
	return ui.InfoPanelData{
		SopsYamlRelPath: "secrets/.sops.yaml",
		AgeFingerprint:  "age1abcdefghijklmnop",
		RecipientCount:  4,
		GitBranch:       "main",
		GitDirty:        true,
		FileCount:       12,
	}
}

// fixtureHints is the deterministic MenuHint slice used by chrome + menu
// matrix sub-tests.
func fixtureHints() []keys.MenuHint {
	return []keys.MenuHint{
		{Mnemonic: "?", Description: "help", Visible: true},
		{Mnemonic: "q", Description: "quit", Visible: true},
		{Mnemonic: "/", Description: "search", Visible: true},
		{Mnemonic: "r", Description: "reveal", Visible: true},
	}
}

// TestRenderChrome_FourProfiles captures full-tier chrome at 200x60 across
// the 4-profile matrix. The chrome composes info panel + menu + logo via
// ui.RenderChrome — the captured output exercises every sub-renderer that
// consults the palette parameter (Plan 2's signature cascade).
func TestRenderChrome_FourProfiles(t *testing.T) {
	for _, c := range profileCases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			defer withProfile(t, c.profile)()
			palette := ui.PaletteFor(c.profile)
			rendered := ui.RenderChrome(fixtureHints(), ui.LogoInfo, fixtureInfoPanel(), palette, 200)
			out := downsampleForProfile(c.profile, rendered)

			// Per-profile expected SGR substrings. Goldens carry the
			// ANSI-stripped structural form; RequireGoldenColors checks
			// raw bytes for the per-profile color encoding.
			var wantColors []string
			switch c.profile {
			case colorprofile.Ascii:
				// Ascii TTY: NO color SGR for the chrome surfaces.
				stripped := ansi.Strip(out)
				assert.Contains(t, stripped, "cfg:", "info panel cfg label visible on Ascii")
				assert.Contains(t, stripped, "secrets/.sops.yaml", "info panel cfg value")
				wantColors = nil
			case colorprofile.ANSI:
				// ANSI 4-bit: bare ANSI escape codes for foreground colors.
				// ColorAccent #cba6f7 post-downsample to ANSI -> nearest
				// color in 16-cube is bright blue 94 (verified empirically;
				// the downsample algorithm picks 94 over 95 because mauve
				// is closer to bright blue than bright magenta).
				wantColors = []string{"\x1b[94m"}
			case colorprofile.ANSI256:
				// ANSI256: 8-bit indexed via 38;5;N.
				wantColors = []string{"38;5;"}
			case colorprofile.TrueColor:
				// TrueColor: 24-bit RGB. Mauve = #cba6f7 = 203;166;247.
				wantColors = []string{"203;166;247"}
			}
			testutil.RequireGoldenColors(t, "profile_chrome_full_"+c.name, out, wantColors)
			testutil.RequireGoldenStructure(t, "profile_chrome_full_"+c.name, out)
		})
	}
}

// TestRenderCrumbs_FourProfiles captures the crumbs row with an active
// chip across the 4-profile matrix. Demonstrates the bracket-fallback vs
// pill rendering distinction at the SGR byte level.
func TestRenderCrumbs_FourProfiles(t *testing.T) {
	segs := []string{"sops-tui", "files", "prod.yaml"}
	for _, c := range profileCases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			defer withProfile(t, c.profile)()
			palette := ui.PaletteFor(c.profile)
			rendered := ui.RenderCrumbs(segs, palette, 80)
			out := downsampleForProfile(c.profile, rendered)

			var wantColors []string
			switch c.profile {
			case colorprofile.Ascii:
				// Bracket fallback: SGR 4 (underline) + SGR 1 (bold) on
				// active chip. No color SGR (Ascii strips fg/bg).
				wantColors = nil
			case colorprofile.ANSI:
				// Bracket fallback: same Underline+Bold; inactive chip uses
				// CrumbChipFallbackStyle Foreground(ColorFgANSI=ANSIColor(15))
				// which renders as "\x1b[97m" on ANSI profile.
				wantColors = []string{"\x1b[97m"}
			case colorprofile.ANSI256:
				// Pill rendering at 8-bit: 38;5;N for fg + 48;5;N for bg
				// + bold on active chip.
				wantColors = []string{"38;5;", "48;5;"}
			case colorprofile.TrueColor:
				// Pill rendering at 24-bit: full RGB on bg + fg + bold.
				wantColors = []string{"203;166;247", "30;30;46"}
			}
			testutil.RequireGoldenColors(t, "profile_crumbs_active_"+c.name, out, wantColors)
			testutil.RequireGoldenStructure(t, "profile_crumbs_active_"+c.name, out)
		})
	}
}

// TestRenderMenu_FourProfiles captures the menu grid with one populated
// state across the 4-profile matrix. The mnemonic columns use
// MenuKeyStyle (accent fg) which downsamples per-profile.
func TestRenderMenu_FourProfiles(t *testing.T) {
	hints := fixtureHints()
	for _, c := range profileCases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			defer withProfile(t, c.profile)()
			palette := ui.PaletteFor(c.profile)
			rendered := ui.RenderMenu(hints, palette, 200)
			out := downsampleForProfile(c.profile, rendered)

			var wantColors []string
			switch c.profile {
			case colorprofile.Ascii:
				wantColors = nil
			case colorprofile.ANSI:
				// MenuKeyStyle Foreground(ColorAccent=#cba6f7); ColorAccent
				// post-downsample to ANSI -> bright blue 94 (mauve is closer
				// to bright blue than bright magenta in the 16-cube algorithm).
				wantColors = []string{"\x1b[94m"}
			case colorprofile.ANSI256:
				wantColors = []string{"38;5;"}
			case colorprofile.TrueColor:
				wantColors = []string{"203;166;247"}
			}
			testutil.RequireGoldenColors(t, "profile_menu_"+c.name, out, wantColors)
			testutil.RequireGoldenStructure(t, "profile_menu_"+c.name, out)
		})
	}
}

// TestFlashBar_FourProfiles_Info captures the flash bar at FlashSevInfo
// across the 4-profile matrix. Info is unprefixed (no [I]) and uses the
// default StatusBarStyle surface bg per D-411 + D-412.
func TestFlashBar_FourProfiles_Info(t *testing.T) {
	for _, c := range profileCases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			defer withProfile(t, c.profile)()
			sb := ui.NewStatusBarModel(ui.EnvStatus{
				SopsAvailable:     true,
				AgeAvailable:      true,
				SopsYamlAvailable: true,
			})
			sb, _ = sb.FlashInfo("Decrypted")
			rendered := sb.View(80)
			out := downsampleForProfile(c.profile, rendered)
			stripped := ansi.Strip(out)

			// Info: no [W] / [E] / [I] prefix.
			assert.NotContains(t, stripped, "[W] ", "Info severity must be unprefixed (D-411)")
			assert.NotContains(t, stripped, "[E] ", "Info severity must be unprefixed (D-411)")
			assert.NotContains(t, stripped, "[I] ", "Info severity must be unprefixed (D-411)")
			require.Contains(t, stripped, "Decrypted", "flash text must render")

			var wantColors []string
			switch c.profile {
			case colorprofile.Ascii:
				wantColors = nil
			case colorprofile.ANSI:
				// StatusBarStyle Background(ColorSurface=#313244); post-
				// downsample to ANSI -> standard black bg 40 (the surface
				// gray is closer to black than to bright black under 16-cube
				// nearest-color). Lipgloss may emit standalone "\x1b[40m" or
				// combined "\x1b[94;40m" so match the literal "40m" suffix.
				wantColors = []string{"\x1b[40m"}
			case colorprofile.ANSI256:
				wantColors = []string{"48;5;"}
			case colorprofile.TrueColor:
				// ColorSurface = #313244 = RGB 49;50;68.
				wantColors = []string{"49;50;68"}
			}
			testutil.RequireGoldenColors(t, "profile_flash_info_"+c.name, out, wantColors)
			testutil.RequireGoldenStructure(t, "profile_flash_info_"+c.name, out)
		})
	}
}

// TestFlashBar_FourProfiles_Warn captures the flash bar at FlashSevWarn
// across the 4-profile matrix. Warn is prefixed with [W] (D-411) and
// uses FlashWarnBarStyle with ColorWarning bg + ColorBg fg (D-412 / Plan 1).
func TestFlashBar_FourProfiles_Warn(t *testing.T) {
	for _, c := range profileCases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			defer withProfile(t, c.profile)()
			sb := ui.NewStatusBarModel(ui.EnvStatus{
				SopsAvailable:     true,
				AgeAvailable:      true,
				SopsYamlAvailable: true,
			})
			sb, _ = sb.FlashWarn("No changes detected")
			rendered := sb.View(80)
			out := downsampleForProfile(c.profile, rendered)
			stripped := ansi.Strip(out)

			require.Contains(t, stripped, "[W] ", "Warn severity MUST emit [W] prefix (D-411)")
			require.Contains(t, stripped, "No changes detected", "flash text must render after prefix")

			var wantColors []string
			switch c.profile {
			case colorprofile.Ascii:
				wantColors = nil
			case colorprofile.ANSI:
				// ColorWarning post-D-415 = #fab387; post-downsample to ANSI
				// -> bright red bg 101 (peach is closer to bright red than
				// bright yellow in the 16-cube algorithm).
				wantColors = []string{"\x1b[101m"}
			case colorprofile.ANSI256:
				wantColors = []string{"48;5;"}
			case colorprofile.TrueColor:
				// ColorWarning = #fab387 = RGB 250;179;135 (peach).
				wantColors = []string{"250;179;135"}
			}
			testutil.RequireGoldenColors(t, "profile_flash_warn_"+c.name, out, wantColors)
			testutil.RequireGoldenStructure(t, "profile_flash_warn_"+c.name, out)
		})
	}
}

// TestFlashBar_FourProfiles_Err captures the flash bar at FlashSevErr
// across the 4-profile matrix. Err is prefixed with [E] (D-411) and uses
// FlashErrBarStyle with ColorError bg + ColorBg fg (D-412 / Plan 1).
func TestFlashBar_FourProfiles_Err(t *testing.T) {
	for _, c := range profileCases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			defer withProfile(t, c.profile)()
			sb := ui.NewStatusBarModel(ui.EnvStatus{
				SopsAvailable:     true,
				AgeAvailable:      true,
				SopsYamlAvailable: true,
			})
			sb, _ = sb.FlashErr("Re-encryption failed")
			rendered := sb.View(80)
			out := downsampleForProfile(c.profile, rendered)
			stripped := ansi.Strip(out)

			require.Contains(t, stripped, "[E] ", "Err severity MUST emit [E] prefix (D-411)")
			require.Contains(t, stripped, "Re-encryption failed", "flash text must render after prefix")

			var wantColors []string
			switch c.profile {
			case colorprofile.Ascii:
				wantColors = nil
			case colorprofile.ANSI:
				// ColorError post-D-415 = #eba0ac; post-downsample to ANSI
				// -> bright red bg 101 (maroon is closer to bright red).
				wantColors = []string{"\x1b[101m"}
			case colorprofile.ANSI256:
				wantColors = []string{"48;5;"}
			case colorprofile.TrueColor:
				// ColorError = #eba0ac = RGB 235;160;172 (maroon).
				wantColors = []string{"235;160;172"}
			}
			testutil.RequireGoldenColors(t, "profile_flash_err_"+c.name, out, wantColors)
			testutil.RequireGoldenStructure(t, "profile_flash_err_"+c.name, out)
		})
	}
}
