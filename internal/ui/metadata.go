// Package ui provides the metadata overlay component for sops-tui.
//
// MetadataModel renders a full-screen overlay showing SOPS file metadata.
// It mirrors HelpModel's pattern: bordered box, surface background, j/k scroll.
//
// Per D-07: metadata panel is a full-screen overlay (same as help ? overlay).
// Per D-08: accessible from both file list view and detail view.
// Per D-09: uses prevState/stateMetadata pattern; i or Esc closes it.
// Per UI-SPEC Overlay Layout Contract and Metadata Overlay Content Contract.
//
// Note: Never use type any. Never use lipgloss.AdaptiveColor (issue #1036).
package ui

import (
	"strings"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
)

// MetadataContent holds SOPS metadata for display in the overlay.
// This is a display-oriented struct that mirrors parser.SopsMetadata.
// Plan 03 wiring converts parser.SopsMetadata -> MetadataContent.
type MetadataContent struct {
	Version          string
	LastModified     string
	MAC              string
	AgeRecipients    []string
	EncryptedRegex   string
	UnencryptedRegex string
}

// MetadataModel renders a full-screen overlay showing SOPS file metadata.
// It mirrors HelpModel's pattern: bordered box, surface background, j/k scroll.
type MetadataModel struct {
	meta   MetadataContent
	width  int
	height int
	scroll int
}

// NewMetadataModel creates a MetadataModel sized to the given dimensions.
func NewMetadataModel(meta MetadataContent, width, height int) MetadataModel {
	return MetadataModel{
		meta:   meta,
		width:  width,
		height: height,
		scroll: 0,
	}
}

// SetSize updates the component dimensions.
func (m *MetadataModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// ScrollDown scrolls the metadata content down by one line.
// Clamped to prevent scrolling past the last content line.
func (m *MetadataModel) ScrollDown() {
	lines := m.buildContentLines()
	maxScroll := len(lines) - 1
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scroll < maxScroll {
		m.scroll++
	}
}

// ScrollUp scrolls the metadata content up by one line.
// Clamped to 0 (top of content).
func (m *MetadataModel) ScrollUp() {
	if m.scroll > 0 {
		m.scroll--
	}
}

// buildContentLines constructs all scrollable content lines for the metadata overlay.
// Returns each line as a pre-rendered string.
func (m MetadataModel) buildContentLines() []string {
	// Phase 7.1 D-110: package vars in styles.go, byte-identical chains.
	labelStyle := MetadataLabelStyle
	valueStyle := MetadataValueStyle
	noneStyle := MetadataNoneStyle

	none := noneStyle.Render("(none)")

	var lines []string

	// version
	versionVal := m.meta.Version
	if versionVal == "" {
		versionVal = ""
	}
	lines = append(lines, labelStyle.Render("version")+valueStyle.Render(versionVal))

	// last modified
	lastModVal := m.meta.LastModified
	lines = append(lines, labelStyle.Render("last modified")+valueStyle.Render(lastModVal))

	// MAC
	macVal := m.meta.MAC
	lines = append(lines, labelStyle.Render("MAC")+valueStyle.Render(macVal))

	// recipients
	if len(m.meta.AgeRecipients) == 0 {
		lines = append(lines, labelStyle.Render("recipients")+none)
	} else {
		for i, r := range m.meta.AgeRecipients {
			if i == 0 {
				lines = append(lines, labelStyle.Render("recipients")+valueStyle.Render(r))
			} else {
				// Subsequent recipients: indent by 16 cells to align under value column
				indent := strings.Repeat(" ", 16)
				lines = append(lines, indent+valueStyle.Render(r))
			}
		}
	}

	// enc regex
	encVal := m.meta.EncryptedRegex
	if encVal == "" {
		lines = append(lines, labelStyle.Render("enc regex")+none)
	} else {
		lines = append(lines, labelStyle.Render("enc regex")+valueStyle.Render(encVal))
	}

	// unc regex
	uncVal := m.meta.UnencryptedRegex
	if uncVal == "" {
		lines = append(lines, labelStyle.Render("unc regex")+none)
	} else {
		lines = append(lines, labelStyle.Render("unc regex")+valueStyle.Render(uncVal))
	}

	return lines
}

// View renders the full-screen metadata overlay.
// Returns a bordered box with all SOPS metadata fields and scroll support.
func (m MetadataModel) View() string {
	// Title: bold "SOPS Metadata" per UI-SPEC Copywriting Contract
	title := HelpSectionHeader.Render("SOPS Metadata")

	// Build all content lines
	allLines := m.buildContentLines()

	// Apply scroll offset
	visibleLines := allLines
	if m.scroll > 0 && m.scroll < len(allLines) {
		visibleLines = allLines[m.scroll:]
	}

	content := strings.Join(visibleLines, "\n")

	// Footer per UI-SPEC Copywriting Contract
	footer := OverlayMutedFooterStyle.Render("Press i or Esc to close")

	// Phase 7.1 D-112: View() returns inner content only; the outer
	// WrapTitled at AppModel.View() (model.go:1342) is the single border
	// source. Width/height are still tracked via SetSize for scroll math.
	inner := title + "\n\n" + content + "\n\n" + footer
	return inner
}

// Hints returns the 5-hint persistent menu set for MetadataModel per D-09.
func (m MetadataModel) Hints() []keys.MenuHint {
	return []keys.MenuHint{
		{Mnemonic: "j", Description: "scroll down", Visible: true},
		{Mnemonic: "k", Description: "scroll up", Visible: true},
		{Mnemonic: "i", Description: "close metadata", Visible: true},
		{Mnemonic: "Esc", Description: "close metadata", Visible: true},
		{Mnemonic: "q", Description: "quit", Visible: true},
	}
}
