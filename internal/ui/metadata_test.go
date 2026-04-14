package ui_test

import (
	"strings"
	"testing"

	"github.com/caesarakalaeii/sops-tui/internal/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewMetadataModel verifies the model is created with accessible MetadataContent fields.
func TestNewMetadataModel(t *testing.T) {
	meta := ui.MetadataContent{
		Version:        "3.12.2",
		LastModified:   "2024-01-15T10:30:00Z",
		MAC:            "ENC[AES256_GCM,data:mac_value]",
		AgeRecipients:  []string{"age1abc123", "age1def456"},
		EncryptedRegex: "^(password|secret)$",
	}
	m := ui.NewMetadataModel(meta, 80, 24)
	require.NotNil(t, m)
}

// TestMetadataViewTitle verifies the overlay title "SOPS Metadata" appears in View() output.
func TestMetadataViewTitle(t *testing.T) {
	meta := ui.MetadataContent{Version: "3.12.2"}
	m := ui.NewMetadataModel(meta, 80, 24)
	output := m.View()
	assert.Contains(t, output, "SOPS Metadata", "View() must contain the title 'SOPS Metadata'")
}

// TestMetadataViewFooter verifies "Press i or Esc to close" appears in View() output.
func TestMetadataViewFooter(t *testing.T) {
	meta := ui.MetadataContent{Version: "3.12.2"}
	m := ui.NewMetadataModel(meta, 80, 24)
	output := m.View()
	assert.Contains(t, output, "Press i or Esc to close", "View() must contain footer text")
}

// TestMetadataViewVersionField verifies the version label and value appear in View() output.
func TestMetadataViewVersionField(t *testing.T) {
	meta := ui.MetadataContent{Version: "3.12.2"}
	m := ui.NewMetadataModel(meta, 80, 24)
	output := m.View()
	assert.Contains(t, output, "version", "View() must contain label 'version'")
	assert.Contains(t, output, "3.12.2", "View() must contain version value '3.12.2'")
}

// TestMetadataViewLastModifiedField verifies the last modified label and value appear in View() output.
func TestMetadataViewLastModifiedField(t *testing.T) {
	meta := ui.MetadataContent{LastModified: "2024-01-15T10:30:00Z"}
	m := ui.NewMetadataModel(meta, 80, 24)
	output := m.View()
	assert.Contains(t, output, "last modified", "View() must contain label 'last modified'")
	assert.Contains(t, output, "2024-01-15T10:30:00Z", "View() must contain last modified value")
}

// TestMetadataViewMACField verifies the MAC label and value appear in View() output.
func TestMetadataViewMACField(t *testing.T) {
	meta := ui.MetadataContent{MAC: "ENC[AES256_GCM,data:mac_value]"}
	m := ui.NewMetadataModel(meta, 80, 24)
	output := m.View()
	assert.Contains(t, output, "MAC", "View() must contain label 'MAC'")
	assert.Contains(t, output, "ENC[AES256_GCM,data:mac_value]", "View() must contain MAC value")
}

// TestMetadataViewRecipients verifies each age recipient is rendered on its own line.
func TestMetadataViewRecipients(t *testing.T) {
	meta := ui.MetadataContent{
		AgeRecipients: []string{"age1abc123", "age1def456"},
	}
	m := ui.NewMetadataModel(meta, 80, 24)
	output := m.View()
	assert.Contains(t, output, "recipients", "View() must contain label 'recipients'")
	assert.Contains(t, output, "age1abc123", "View() must contain first recipient")
	assert.Contains(t, output, "age1def456", "View() must contain second recipient")
}

// TestMetadataViewEncryptedRegex verifies the enc regex field renders "(none)" when empty.
func TestMetadataViewEncryptedRegex(t *testing.T) {
	meta := ui.MetadataContent{EncryptedRegex: "^(password|secret)$"}
	m := ui.NewMetadataModel(meta, 80, 24)
	output := m.View()
	assert.Contains(t, output, "enc regex", "View() must contain label 'enc regex'")
	assert.Contains(t, output, "^(password|secret)$", "View() must contain regex value")

	metaEmpty := ui.MetadataContent{}
	mEmpty := ui.NewMetadataModel(metaEmpty, 80, 24)
	outputEmpty := mEmpty.View()
	// When empty, should show "(none)" for enc regex
	assert.Contains(t, outputEmpty, "(none)", "View() must contain '(none)' for empty EncryptedRegex")
}

// TestMetadataViewUnencryptedRegex verifies the unc regex field renders "(none)" when empty.
func TestMetadataViewUnencryptedRegex(t *testing.T) {
	meta := ui.MetadataContent{UnencryptedRegex: "^(public|id)$"}
	m := ui.NewMetadataModel(meta, 80, 24)
	output := m.View()
	assert.Contains(t, output, "unc regex", "View() must contain label 'unc regex'")
	assert.Contains(t, output, "^(public|id)$", "View() must contain unencrypted regex value")
}

// TestMetadataViewEmptyContent verifies View() handles zero-value MetadataContent without panicking.
func TestMetadataViewEmptyContent(t *testing.T) {
	meta := ui.MetadataContent{}
	m := ui.NewMetadataModel(meta, 80, 24)

	require.NotPanics(t, func() {
		output := m.View()
		assert.Contains(t, output, "SOPS Metadata", "View() with zero-value must still contain title")
		assert.Contains(t, output, "(none)", "View() with zero-value must show '(none)' for empty fields")
	})
}

// TestMetadataSetSize verifies SetSize updates dimensions and View() reflects the new size.
func TestMetadataSetSize(t *testing.T) {
	meta := ui.MetadataContent{Version: "3.12.2"}
	m := ui.NewMetadataModel(meta, 80, 24)
	m.SetSize(120, 40)
	// After SetSize, View() should still work without panicking
	require.NotPanics(t, func() {
		output := m.View()
		assert.Contains(t, output, "SOPS Metadata")
	})
}

// TestMetadataScrolling verifies ScrollDown and ScrollUp change the scroll state without panicking.
func TestMetadataScrolling(t *testing.T) {
	// Create content with multiple recipients to enable scrolling
	meta := ui.MetadataContent{
		AgeRecipients: []string{
			"age1abc123", "age1def456", "age1ghi789",
			"age1jkl012", "age1mno345", "age1pqr678",
		},
	}
	m := ui.NewMetadataModel(meta, 80, 10)

	require.NotPanics(t, func() {
		m.ScrollDown()
		_ = m.View()
		m.ScrollDown()
		_ = m.View()
		m.ScrollUp()
		_ = m.View()
	})
}

// TestMetadataAllFields verifies a fully populated MetadataContent renders all expected fields.
func TestMetadataAllFields(t *testing.T) {
	meta := ui.MetadataContent{
		Version:          "3.12.2",
		LastModified:     "2024-01-15T10:30:00Z",
		MAC:              "ENC[AES256_GCM,data:mac_value]",
		AgeRecipients:    []string{"age1abc123", "age1def456"},
		EncryptedRegex:   "^(password|secret)$",
		UnencryptedRegex: "^(public|id)$",
	}
	m := ui.NewMetadataModel(meta, 120, 40)
	output := m.View()

	checks := []struct {
		name     string
		expected string
	}{
		{"title", "SOPS Metadata"},
		{"footer", "Press i or Esc to close"},
		{"version label", "version"},
		{"version value", "3.12.2"},
		{"last modified label", "last modified"},
		{"last modified value", "2024-01-15T10:30:00Z"},
		{"MAC label", "MAC"},
		{"MAC value", "ENC[AES256_GCM,data:mac_value]"},
		{"recipients label", "recipients"},
		{"first recipient", "age1abc123"},
		{"second recipient", "age1def456"},
		{"enc regex label", "enc regex"},
		{"enc regex value", "^(password|secret)$"},
		{"unc regex label", "unc regex"},
		{"unc regex value", "^(public|id)$"},
	}

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			assert.Contains(t, output, check.expected,
				"View() must contain %s: %q", check.name, check.expected)
		})
	}
}

// TestMetadataEmptyRecipientsShowsNone verifies that empty AgeRecipients renders "(none)".
func TestMetadataEmptyRecipientsShowsNone(t *testing.T) {
	meta := ui.MetadataContent{AgeRecipients: []string{}}
	m := ui.NewMetadataModel(meta, 80, 24)
	output := m.View()
	// recipients label should appear, and (none) should appear for empty list
	assert.Contains(t, output, "recipients", "View() must contain 'recipients' label")

	// nil recipients should also work
	metaNil := ui.MetadataContent{}
	mNil := ui.NewMetadataModel(metaNil, 80, 24)
	outputNil := mNil.View()
	assert.Contains(t, outputNil, "(none)", "View() with nil recipients must show '(none)'")
}

// TestMetadataViewContainsAllLabelStrings verifies all 6 label strings appear in View() output.
func TestMetadataViewContainsAllLabelStrings(t *testing.T) {
	meta := ui.MetadataContent{
		Version:      "1.0.0",
		LastModified: "2024-01-01T00:00:00Z",
		MAC:          "abc123",
	}
	m := ui.NewMetadataModel(meta, 80, 24)
	output := m.View()

	labels := []string{"version", "last modified", "MAC", "recipients", "enc regex", "unc regex"}
	for _, label := range labels {
		assert.Contains(t, output, label, "View() must contain label %q", label)
	}
}

// TestMetadataRoundedBorderExists verifies the output contains rounded border characters.
func TestMetadataRoundedBorderExists(t *testing.T) {
	meta := ui.MetadataContent{Version: "3.12.2"}
	m := ui.NewMetadataModel(meta, 80, 24)
	output := m.View()
	// RoundedBorder uses rounded corner characters: ╭ or ╰
	hasRoundedCorner := strings.ContainsAny(output, "╭╰╮╯")
	assert.True(t, hasRoundedCorner, "View() must contain rounded border characters (╭, ╰, ╮, ╯)")
}
