// Package ui defines the design system for sops-tui: color palette, spacing tokens,
// and named lipgloss styles used throughout all views.
//
// Critical: Do NOT use lipgloss.AdaptiveColor — confirmed hang (issue #1036).
// All colors are explicit hex values per 01-UI-SPEC.md §Color.
package ui

import (
	"charm.land/lipgloss/v2"
)

// Color hex values per 01-UI-SPEC.md §Color. These constants hold the raw hex strings
// so they can be verified in tests and used to construct lipgloss colors.
const (
	// ColorBgHex is the terminal background fill color (60% usage).
	ColorBgHex = "#1e1e2e"
	// ColorSurfaceHex is used for status bar, help overlay, and error box backgrounds (30% usage).
	ColorSurfaceHex = "#313244"
	// ColorAccentHex is reserved for selection highlights, active breadcrumbs, tree indicators,
	// and help key labels (10% usage).
	ColorAccentHex = "#89b4fa"
	// ColorSuccessHex indicates healthy status (sops/age/.sops.yaml available).
	ColorSuccessHex = "#a6e3a1"
	// ColorWarningHex indicates soft warnings (age key missing, .sops.yaml missing).
	ColorWarningHex = "#f9e2af"
	// ColorErrorHex indicates hard errors (sops binary missing, error box borders).
	ColorErrorHex = "#f38ba8"
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

	// Phase 5: Health check severity styles (HLT-03)
	HealthWeakStyle  = lipgloss.NewStyle().Foreground(ColorWarning)
	HealthDupeStyle  = lipgloss.NewStyle().Foreground(ColorError)
	HealthStaleStyle = lipgloss.NewStyle().Foreground(ColorMuted)
	HealthOkStyle    = lipgloss.NewStyle().Foreground(ColorSuccess)

	// Phase 5: Validation error style for RecipientFormModel inline errors (T-05-05)
	ValidationErrorStyle = lipgloss.NewStyle().Foreground(ColorError)

	// Phase 5: Selection indicator for bulk re-key (D-05)
	SelectionIndicator = lipgloss.NewStyle().Foreground(ColorAccent)
)
