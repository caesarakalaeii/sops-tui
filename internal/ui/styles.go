// Package ui defines the design system for sops-tui: color palette, spacing tokens,
// and named lipgloss styles used throughout all views.
//
// Critical: Do NOT use lipgloss.AdaptiveColor — confirmed hang (issue #1036).
// All colors are explicit hex values per 01-UI-SPEC.md §Color.
package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

// Color hex values per 01-UI-SPEC.md §Color. These constants hold the raw hex strings
// so they can be verified in tests and used to construct lipgloss colors.
//
// Phase 10 D-415 flipped three accents toward Catppuccin Mocha's hot-pink/purple
// register (mauve / peach / maroon) for closer k9s visual parity. The other 5
// constants (Bg / Surface / Success / Muted / Fg) are byte-identical to v1.0.
const (
	// ColorBgHex is the terminal background fill color (60% usage).
	ColorBgHex = "#1e1e2e"
	// ColorSurfaceHex is used for status bar, help overlay, and error box backgrounds (30% usage).
	ColorSurfaceHex = "#313244"
	// ColorAccentHex is reserved for selection highlights, active breadcrumbs, tree indicators,
	// and help key labels (10% usage). Catppuccin Mauve (Phase 10 D-415).
	ColorAccentHex = "#cba6f7"
	// ColorSuccessHex indicates healthy status (sops/age/.sops.yaml available).
	ColorSuccessHex = "#a6e3a1"
	// ColorWarningHex indicates soft warnings (age key missing, .sops.yaml missing).
	// Catppuccin Peach (Phase 10 D-415).
	ColorWarningHex = "#fab387"
	// ColorErrorHex indicates hard errors (sops binary missing, error box borders).
	// Catppuccin Maroon (Phase 10 D-415).
	ColorErrorHex = "#eba0ac"
	// ColorMutedHex is used for dim text, separators, and borders.
	ColorMutedHex = "#6c7086"
	// ColorFgHex is the default foreground text color on dark background.
	ColorFgHex = "#cdd6f4"
)

// Color palette — lipgloss color values built from explicit hex constants.
// Never use lipgloss.AdaptiveColor (issue #1036 confirmed hang).
var (
	// ColorBg is the terminal background fill color (60% usage).
	ColorBg = lipgloss.Color(ColorBgHex)
	// ColorSurface is used for status bar, help overlay, and error box backgrounds (30% usage).
	ColorSurface = lipgloss.Color(ColorSurfaceHex)
	// ColorAccent is reserved for selection highlights, active breadcrumbs, tree indicators,
	// and help key labels (10% usage).
	ColorAccent = lipgloss.Color(ColorAccentHex)
	// ColorSuccess indicates healthy status (sops/age/.sops.yaml available).
	ColorSuccess = lipgloss.Color(ColorSuccessHex)
	// ColorWarning indicates soft warnings (age key missing, .sops.yaml missing).
	ColorWarning = lipgloss.Color(ColorWarningHex)
	// ColorError indicates hard errors (sops binary missing, error box borders).
	ColorError = lipgloss.Color(ColorErrorHex)
	// ColorMuted is used for dim text, separators, and borders.
	ColorMuted = lipgloss.Color(ColorMutedHex)
	// ColorFg is the default foreground text color on dark background.
	ColorFg = lipgloss.Color(ColorFgHex)
)

// Phase 10 (D-420): 16-color fallback variants for Profile <= colorprofile.ANSI.
// Each index hand-verified in 10-RESEARCH.md §"ANSI 16-Color Verification Table"
// — every chrome bg/fg pair is distinct under 4-bit downsampling.
//
// PaletteFor(profile) selects between these ANSI variants and the 24-bit
// Catppuccin Mocha variants above based on the active color profile.
var (
	// ColorAccentANSI is bright magenta (13) — distinct from Warning yellow (11)
	// + Error red (9) under 4-bit. Replaces 24-bit Mauve when palette.Fallback.
	ColorAccentANSI = lipgloss.ANSIColor(13)
	// ColorBgANSI is black (0) — terminal default background.
	ColorBgANSI = lipgloss.ANSIColor(0)
	// ColorSurfaceANSI is bright black / dark grey (8) — surface overlay / status bar bg.
	ColorSurfaceANSI = lipgloss.ANSIColor(8)
	// ColorFgANSI is bright white (15) — primary foreground text.
	ColorFgANSI = lipgloss.ANSIColor(15)
	// ColorMutedANSI is white / light grey (7) — secondary text + borders.
	ColorMutedANSI = lipgloss.ANSIColor(7)
	// ColorSuccessANSI is bright green (10) — sops/age/.sops.yaml available indicators.
	ColorSuccessANSI = lipgloss.ANSIColor(10)
	// ColorWarningANSI is bright yellow (11) — soft env warnings + Warn flash bg.
	ColorWarningANSI = lipgloss.ANSIColor(11)
	// ColorErrorANSI is bright red (9) — hard env errors + Err flash bg.
	ColorErrorANSI = lipgloss.ANSIColor(9)
)

// Palette is the resolved color set passed into chrome renderers. RenderChrome
// / RenderCrumbs / RenderMenu / RenderInfoPanel accept a Palette parameter and
// consult its fields rather than importing colorprofile directly. The Fallback
// bool gates Plan 3's bracket-chip rendering on Ascii/ANSI profiles (D-422).
//
// Phase 10 D-421. Field type is image/color.Color (the standard library
// interface that both lipgloss.Color() string-returning function values and
// lipgloss.ANSIColor() typed values satisfy) so a single struct can hold
// either the 24-bit or the 16-color variant set without conversion.
type Palette struct {
	// Accent — selection highlights, active breadcrumb chip bg, mnemonic labels.
	Accent color.Color
	// Bg — terminal background (60% usage, dark base).
	Bg color.Color
	// Surface — status bar, help overlay, error box, inactive chip bg (30% usage).
	Surface color.Color
	// Fg — default foreground text on dark background.
	Fg color.Color
	// Muted — dim text, separators, borders, ellipsis chip fg.
	Muted color.Color
	// Success — healthy env indicators (sops/age/.sops.yaml available).
	Success color.Color
	// Warning — soft env warnings + Warn-severity flash bg.
	Warning color.Color
	// Error — hard env errors + Err-severity flash bg + error box border.
	Error color.Color
	// Fallback is true when the palette is the 16-color ANSI variant set
	// (profile <= colorprofile.ANSI). Plan 3's RenderCrumbs gates bracket-
	// chip rendering on this bool — paired bg/fg collapses on 16-color
	// terminals so the active chip drops bg fill and uses Underline+Bold
	// SGR codes (which survive every downsample) instead.
	Fallback bool
}

// PaletteFor returns the resolved Palette for the requested color profile.
// Profile <= colorprofile.ANSI returns the 16-color fallback variants
// (Fallback=true); otherwise returns the 24-bit Catppuccin Mocha variants
// (Fallback=false). Per D-421 the accessor is the single switch point for
// profile-aware variant selection — chrome renderers consult the Palette
// fields, never colorprofile.Profile directly.
//
// NOTE: D-421 in CONTEXT.md said "Palette(profile)" but that name collides
// with the Palette struct type. The accessor is renamed to PaletteFor
// (verb form) per pattern-mapper deviation #4.
func PaletteFor(profile colorprofile.Profile) Palette {
	if profile <= colorprofile.ANSI {
		return Palette{
			Accent:   ColorAccentANSI,
			Bg:       ColorBgANSI,
			Surface:  ColorSurfaceANSI,
			Fg:       ColorFgANSI,
			Muted:    ColorMutedANSI,
			Success:  ColorSuccessANSI,
			Warning:  ColorWarningANSI,
			Error:    ColorErrorANSI,
			Fallback: true,
		}
	}
	return Palette{
		Accent:   ColorAccent,
		Bg:       ColorBg,
		Surface:  ColorSurface,
		Fg:       ColorFg,
		Muted:    ColorMuted,
		Success:  ColorSuccess,
		Warning:  ColorWarning,
		Error:    ColorError,
		Fallback: false,
	}
}

// Spacing tokens per 01-UI-SPEC.md §Spacing Scale.
// Values are in terminal cells (columns for horizontal, rows for vertical).
const (
	// SpaceXS is 1 cell — icon gap, inline separator padding.
	SpaceXS = 1
	// SpaceSM is 2 cells — compact horizontal padding (status bar sections).
	SpaceSM = 2
	// SpaceMD is 4 cells — default element horizontal padding (pane content indent).
	SpaceMD = 4
	// SpaceLG is 6 cells — section separation, border + content gap.
	SpaceLG = 6
	// SpaceXL is 8 cells — major layout margin.
	SpaceXL = 8
	// TreeIndent is 2 cells per nesting level for YAML tree rendering.
	TreeIndent = 2
)

// Named styles per 01-UI-SPEC.md §Typography and component specs.
// Styles are defined as package-level variables for reuse across views.
var (
	// ErrorLabel renders [ERROR] prefix text in bold red for startup error boxes.
	ErrorLabel = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorError)

	// WarnLabel renders [WARN] prefix text in bold yellow for startup warning boxes.
	WarnLabel = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWarning)

	// StatusBarStyle is the persistent bottom status bar: surface background, fg text.
	StatusBarStyle = lipgloss.NewStyle().
			Background(ColorSurface).
			Foreground(ColorFg)

	// Phase 10 D-412: severity-tinted flash bar styles. Used by
	// StatusBarModel.View() when m.flashSeverity is Warn or Err. The
	// background is the matching warning/error color; foreground is dark
	// (ColorBg) for high contrast. Plan 1 uses the existing ColorWarning
	// and ColorError values — the hex constants are flipped in Plan 2.
	FlashWarnBarStyle = lipgloss.NewStyle().Background(ColorWarning).Foreground(ColorBg)
	FlashErrBarStyle  = lipgloss.NewStyle().Background(ColorError).Foreground(ColorBg)

	// BreadcrumbActive highlights the active breadcrumb segment in accent color.
	BreadcrumbActive = lipgloss.NewStyle().
				Foreground(ColorAccent)

	// BreadcrumbSep renders the `>` separator between breadcrumb segments in muted color.
	BreadcrumbSep = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// TreeConnector renders box-drawing characters (├─, └─, │) in muted color.
	TreeConnector = lipgloss.NewStyle().
			Foreground(ColorMuted)

	// TreeIndicator renders collapse/expand indicators ([+], [-]) in accent color.
	TreeIndicator = lipgloss.NewStyle().
			Foreground(ColorAccent)

	// DimText applies faint/dim attribute for secondary info: item counts, masked values (***).
	DimText = lipgloss.NewStyle().
		Faint(true)

	// SelectedRow highlights the currently selected list or tree item.
	SelectedRow = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Background(ColorSurface)

	// HelpKeyStyle is the fixed-width key label column in help overlay (12 cells, accent color).
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Width(12)

	// HelpDescStyle is the description column in help overlay (normal body weight, fg color).
	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorFg)

	// HelpSectionHeader renders section names (Navigation, Global) in bold in help overlay.
	HelpSectionHeader = lipgloss.NewStyle().
				Bold(true)

	// BadgeUnencrypted renders the [unencrypted] badge on file list items (per D-02).
	BadgeUnencrypted = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorError)

	// BadgePlain renders the [plain] badge on unencrypted leaf values (per D-06).
	BadgePlain = lipgloss.NewStyle().
			Foreground(ColorWarning)

	// TypeHintStyle renders type hints like (str), (int), (bool) after masked values (per D-04).
	TypeHintStyle = lipgloss.NewStyle().
			Faint(true).
			Foreground(ColorMuted)

	// SearchInputStyle is the background style for the inline search bar (per D-10).
	SearchInputStyle = lipgloss.NewStyle().
				Background(ColorSurface).
				Foreground(ColorFg)

	// SearchMatchStyle highlights matched characters in fuzzy search results (per D-12).
	SearchMatchStyle = lipgloss.NewStyle().
				Foreground(ColorAccent)

	// RevealedValueStyle renders revealed secret values (normal weight, no faint) per 03-UI-SPEC.md.
	// Replaces DimText for rows where Revealed=true.
	RevealedValueStyle = lipgloss.NewStyle().Foreground(ColorFg)

	// RevealedIconStyle renders the 🔓 suffix after a revealed value per 03-UI-SPEC.md.
	RevealedIconStyle = lipgloss.NewStyle().Foreground(ColorSuccess)

	// DiffAddedStyle renders diff "+" lines (new value) in the DiffModel per 03-UI-SPEC.md.
	DiffAddedStyle = lipgloss.NewStyle().Foreground(ColorSuccess)

	// DiffRemovedStyle renders diff "−" lines (old value) in the DiffModel per 03-UI-SPEC.md.
	DiffRemovedStyle = lipgloss.NewStyle().Foreground(ColorError)

	// DiffKeyStyle renders the key path label in the diff header row per 03-UI-SPEC.md.
	DiffKeyStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorFg)

	// DiffContextStyle renders unchanged context lines in multi-key diffs per 03-UI-SPEC.md.
	DiffContextStyle = lipgloss.NewStyle().Foreground(ColorMuted)

	// EditInputStyle is the background style for the inline textinput during stateEdit per 03-UI-SPEC.md.
	EditInputStyle = lipgloss.NewStyle().Background(ColorSurface).Foreground(ColorFg)

	// ConfirmPromptStyle renders the [y/n] prompt label in the diff overlay footer per 03-UI-SPEC.md.
	ConfirmPromptStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorAccent)

	// FormatMenuStyle renders format selection menu items (base64/hex/UUID/bcrypt) per 03-UI-SPEC.md.
	FormatMenuStyle = lipgloss.NewStyle().Background(ColorSurface).Foreground(ColorFg)

	// Phase 4: Git badge styles (D-09)
	BadgeModified = lipgloss.NewStyle().Foreground(ColorWarning)
	BadgeAdded    = lipgloss.NewStyle().Foreground(ColorSuccess)
	BadgeUntracked = lipgloss.NewStyle().Foreground(ColorMuted)

	// Phase 4: Clipboard indicator style (D-08)
	ClipboardHotStyle = lipgloss.NewStyle().Foreground(ColorAccent)

	// Phase 4: Git "no repo" indicator style (D-12)
	GitNoRepoStyle = lipgloss.NewStyle().Foreground(ColorMuted)

	// Phase 4: Git history overlay styles (D-15)
	HistoryHashStyle    = lipgloss.NewStyle().Foreground(ColorAccent)
	HistoryDateStyle    = lipgloss.NewStyle().Foreground(ColorMuted)
	HistoryAuthorStyle  = lipgloss.NewStyle().Foreground(ColorFg)
	HistorySubjectStyle = lipgloss.NewStyle().Foreground(ColorFg)

	// Phase 5: Health check severity styles (HLT-03, D-12)
	HealthWeakStyle          = lipgloss.NewStyle().Foreground(ColorError)
	HealthDupeStyle          = lipgloss.NewStyle().Foreground(ColorError)
	HealthStaleStyle         = lipgloss.NewStyle().Foreground(ColorWarning)
	HealthSectionHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorFg)
	HealthSkippedStyle       = lipgloss.NewStyle().Foreground(ColorMuted)
	HealthOkStyle            = lipgloss.NewStyle().Foreground(ColorSuccess)

	// Phase 5: Selection indicator for bulk re-key (D-05)
	SelectionIndicatorStyle = lipgloss.NewStyle().Foreground(ColorAccent)

	// Phase 5: Recipient form validation error (D-02)
	ValidationErrorStyle = lipgloss.NewStyle().Foreground(ColorError)

	// Phase 5: Recipient numbered list index (D-03)
	RecipientIndexStyle = lipgloss.NewStyle().Foreground(ColorMuted)

	// Phase 7: Chrome primitives (D-05, D-12, D-13, D-22)
	// All declared as package-level vars to satisfy the "no lipgloss.NewStyle()
	// inside View()" discipline enforced by TestViewNoNewStyle (Plan 3).

	// MenuKeyStyle renders the mnemonic column "[key]" labels in accent color (D-05).
	MenuKeyStyle = lipgloss.NewStyle().Foreground(ColorAccent)

	// MenuDescStyle renders the description column text in default foreground (D-05).
	MenuDescStyle = lipgloss.NewStyle().Foreground(ColorFg)

	// MenuColumnStyle is the wrapper for each pre-rendered fixed-width
	// column in RenderMenu (Phase 7.1 D-117 — replaces the Phase 7
	// lipgloss/v2/table builder). Width is applied at the call site
	// because it varies per call (width/menuCols). The var has no
	// per-frame data so it does not trip TestSubmodelViewsNoNewStyle in
	// internal/ui (which scans sub-models, not styles.go).
	//
	// Phase 7's per-cell StyleFunc return value was deleted alongside
	// the table builder removal — package vars should match what's
	// actually rendered. Phase 10's per-column tweaks can introduce a
	// new var if needed.
	MenuColumnStyle = lipgloss.NewStyle()

	// LogoStyleInfo is the Phase 7 unconditional logo color (D-02).
	// Phase 10 (UI-03) flips between Info/Warn/Error based on aggregate severity.
	LogoStyleInfo = lipgloss.NewStyle().Foreground(ColorAccent)

	// LogoStyleWarn is declared in Phase 7 for Phase 10 severity coupling (UI-03).
	LogoStyleWarn = lipgloss.NewStyle().Foreground(ColorWarning)

	// LogoStyleError is declared in Phase 7 for Phase 10 severity coupling (UI-03).
	LogoStyleError = lipgloss.NewStyle().Foreground(ColorError)

	// TitledBorderStyle is the uniform titled-border wrapper for every primary view (D-12, D-13).
	// NormalBorder() (not RoundedBorder) per UI-15 — TestChromeNormalBorderOnly enforces this.
	TitledBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				BorderForeground(ColorMuted).
				Padding(0, 1)

	// TitleLabelStyle renders the title text inside the border top-line overlay (muted).
	TitleLabelStyle = lipgloss.NewStyle().Foreground(ColorMuted)

	// InfoPanelPlaceholderStyle reserves the 6-row x 38-col top-left area
	// of the chrome for Phase 8's header info panel (D-16, Pitfall 1 mitigation).
	// Phase 7 renders the empty string into this style so lipgloss.Height
	// returns exactly 6 and JoinHorizontal alignment is preserved.
	InfoPanelPlaceholderStyle = lipgloss.NewStyle().
					Width(38).
					Height(6)

	// Phase 7.1 D-110: package vars lifted out of overlay sub-models
	// (help.go, health.go, history.go, metadata.go, recipientform.go) so
	// the sub-models contain zero `lipgloss.NewStyle()` calls. This
	// satisfies Plan 04's TestSubmodelViewsNoNewStyle AST walker and the
	// project-wide "all styles are package vars" discipline.

	// OverlayMutedFooterStyle is the muted footer text style shared by 5
	// overlay sub-models (help, health, history, metadata, recipientform)
	// for "Press X or Esc to close" footers and similar muted prompts
	// (Phase 7.1 D-110 lift; was an inline lipgloss.NewStyle() in each
	// sub-model's View() before the AST walker BFS scope expansion).
	// Same chain as GitNoRepoStyle but a separate alias so Phase 8+ can
	// diverge them without touching call sites.
	OverlayMutedFooterStyle = lipgloss.NewStyle().Foreground(ColorMuted)

	// MetadataLabelStyle is the muted label column style in MetadataModel
	// content lines (Phase 7.1 D-110 lift from metadata.go:83).
	MetadataLabelStyle = lipgloss.NewStyle().Foreground(ColorMuted).Width(16)

	// MetadataValueStyle is the foreground value column style in
	// MetadataModel content lines (Phase 7.1 D-110 lift from metadata.go:84).
	MetadataValueStyle = lipgloss.NewStyle().Foreground(ColorFg)

	// MetadataNoneStyle is the muted "(none)" placeholder style in
	// MetadataModel content lines (Phase 7.1 D-110 lift from metadata.go:85;
	// could alias GitNoRepoStyle but kept as a separate name to document
	// intent for Phase 8+).
	MetadataNoneStyle = lipgloss.NewStyle().Foreground(ColorMuted)

	// Phase 7.1 D-109: package vars lifted out of internal/app/model.go
	// helpers reachable from AppModel.View() (renderRecipientList,
	// renderFormatMenu) so the BFS reachability walker
	// TestViewNoNewStyle reports zero violations across the View() call
	// graph. See internal/app/chrome_test.go for the walker.

	// FormatMenuOverlayStyle is the modal RoundedBorder envelope for the
	// stateFormatMenu overlay (Phase 7.1 D-109 lift from
	// internal/app/model.go renderFormatMenu helper). Width and Height
	// are applied at the call site because they vary per-frame; only the
	// deterministic chain (Border + BorderForeground + Background +
	// Padding) is the var.
	//
	// Note: this is the ONLY remaining RoundedBorder use after Phase 7.1's
	// strip of the 6 sub-model nested borders. RoundedBorder is permitted
	// here because stateFormatMenu is a transient overlay modal that
	// explicitly opts OUT of WrapTitled at model.go (the rule "WrapTitled
	// is the single border source" applies to primary views, not transient
	// overlay modals).
	FormatMenuOverlayStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorMuted).
				Background(ColorSurface).
				Padding(1, SpaceMD)

	// RecipientListFooterStyle is the muted "(showing first N of M)"
	// footer in renderRecipientList (Phase 7.1 D-109 lift from model.go
	// renderRecipientList). Same chain as OverlayMutedFooterStyle but
	// kept as a separate name so Phase 8+ can diverge them without
	// touching call sites.
	RecipientListFooterStyle = lipgloss.NewStyle().Foreground(ColorMuted)

	// RecipientPromptStyle is the muted "Select recipient to remove"
	// prompt in renderRecipientList (Phase 7.1 D-109 lift from model.go
	// renderRecipientList).
	RecipientPromptStyle = lipgloss.NewStyle().Foreground(ColorMuted)

	// ChromeNarrowFallbackStyle renders the narrow-tier "press ? for help"
	// stub in RenderChrome (Phase 7.1 D-116). Foreground = ColorMuted; no
	// border (the stub is a single line, not a box). At terminal widths
	// below logoWidth + minMenuCol (~33 cols), the chrome cannot fit the
	// 6-row menu+logo pair, so this single-line fallback keeps the body
	// region reachable.
	ChromeNarrowFallbackStyle = lipgloss.NewStyle().Foreground(ColorMuted)

	// Phase 8: Header info panel + crumb chip styles (D-201..D-208, D-216).
	// All declared as package-level vars to satisfy TestViewNoNewStyle (BFS)
	// and TestSubmodelViewsNoNewStyle (file-scope, scope extended to
	// infopanel.go + crumbs.go in Plan 3).

	// InfoPanelLabelStyle renders the muted 5-cell label column (D-201).
	// Width(5) enforces 4-char-label + 1-trailing-space alignment without
	// manual padding (e.g. "cfg:" + " " rendered at Width(5) -> "cfg: ").
	InfoPanelLabelStyle = lipgloss.NewStyle().Foreground(ColorMuted).Width(5)

	// InfoPanelValueStyle renders the foreground value column (D-201, D-204).
	// Width is NOT applied -- values are pre-truncated via middleTruncate
	// in infopanel.go before reaching this style.
	InfoPanelValueStyle = lipgloss.NewStyle().Foreground(ColorFg)

	// InfoPanelSepStyle is reserved for Phase 10 visual tweak (UI-SPEC
	// section Color section "Phase 8 new style declarations"). Phase 8 does
	// not render an explicit separator cell -- the trailing space inside
	// the label Width(5) provides the gap. Declared as no-op so the symbol
	// exists for forward compat without API churn.
	InfoPanelSepStyle = lipgloss.NewStyle()

	// CrumbChipStyle renders inactive crumb chip pills (D-206).
	// Two-channel encoding: surface bg + fg color contrast.
	CrumbChipStyle = lipgloss.NewStyle().Background(ColorSurface).Foreground(ColorFg)

	// CrumbChipActiveStyle renders the active (last) crumb chip pill (D-206).
	// THREE-channel encoding: accent bg + inverted fg (bg color used as fg) +
	// bold weight. Bold is the colorblind-safe redundancy channel (Pitfall 9).
	// k9s deviation: k9s uses bg-only swap; sops-tui adds bold deliberately
	// so the active-vs-inactive distinction survives 16-color downsampling.
	CrumbChipActiveStyle = lipgloss.NewStyle().
				Background(ColorAccent).
				Foreground(ColorBg).
				Bold(true)

	// CrumbChipSepStyle is reserved for forward compat (Phase 10). Phase 8
	// renders the inter-chip separator as a plain " " literal -- no styling.
	CrumbChipSepStyle = lipgloss.NewStyle()

	// CrumbChipEllipsisStyle renders the middle-truncation overflow chip
	// "<...>" (D-216). Muted foreground + no bg fill so the chip reads as
	// "data was here, dropped due to width" -- distinct from both inactive
	// (bg-filled) and active (bg+bold) chips.
	CrumbChipEllipsisStyle = lipgloss.NewStyle().Foreground(ColorMuted)

	// CrumbRowStyle is the row container for the joined chips (D-208).
	// PaddingLeft(SpaceXS) + PaddingRight(SpaceXS) mirrors k9s
	// crumbs.go:32 SetBorderPadding(0,0,1,1).
	CrumbRowStyle = lipgloss.NewStyle().
			PaddingLeft(SpaceXS).
			PaddingRight(SpaceXS)
)
