package keys_test

import (
	"fmt"
	"testing"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"github.com/caesarakalaeii/sops-tui/internal/keys"
	"github.com/stretchr/testify/assert"
)

// keyStringer wraps a key string so it satisfies fmt.Stringer for use with key.Matches.
type keyStringer struct {
	s string
}

func (k keyStringer) String() string { return k.s }

// press returns a keyStringer for the given key string.
func press(s string) fmt.Stringer { return keyStringer{s} }

// TestFileListKeyMap_Navigation verifies navigation bindings on DefaultFileListKeyMap.
func TestFileListKeyMap_Navigation(t *testing.T) {
	km := keys.DefaultFileListKeyMap

	assert.True(t, key.Matches(press("k"), km.Up), "Up must match 'k'")
	assert.True(t, key.Matches(press("up"), km.Up), "Up must match 'up'")

	assert.True(t, key.Matches(press("j"), km.Down), "Down must match 'j'")
	assert.True(t, key.Matches(press("down"), km.Down), "Down must match 'down'")

	assert.True(t, key.Matches(press("g"), km.GoTop), "GoTop must match 'g'")
	assert.True(t, key.Matches(press("G"), km.GoBottom), "GoBottom must match 'G'")

	assert.True(t, key.Matches(press("ctrl+u"), km.HalfUp), "HalfUp must match 'ctrl+u'")
	assert.True(t, key.Matches(press("ctrl+d"), km.HalfDown), "HalfDown must match 'ctrl+d'")

	assert.True(t, key.Matches(press("enter"), km.Open), "Open must match 'enter'")
	assert.True(t, key.Matches(press("l"), km.Open), "Open must match 'l'")
}

// TestDetailKeyMap_Navigation verifies navigation and drill-down bindings on DefaultDetailKeyMap.
func TestDetailKeyMap_Navigation(t *testing.T) {
	km := keys.DefaultDetailKeyMap

	assert.True(t, key.Matches(press("esc"), km.Back), "Back must match 'esc'")

	assert.True(t, key.Matches(press("enter"), km.Expand), "Expand must match 'enter'")
	assert.True(t, key.Matches(press("l"), km.Expand), "Expand must match 'l'")

	assert.True(t, key.Matches(press("h"), km.Collapse), "Collapse must match 'h'")
	assert.True(t, key.Matches(press("left"), km.Collapse), "Collapse must match 'left'")
}

// TestGlobalKeyMap_Bindings verifies global keybindings on DefaultGlobalKeyMap.
func TestGlobalKeyMap_Bindings(t *testing.T) {
	km := keys.DefaultGlobalKeyMap

	assert.True(t, key.Matches(press("?"), km.Help), "Help must match '?'")

	assert.True(t, key.Matches(press("q"), km.Quit), "Quit must match 'q'")
	assert.True(t, key.Matches(press("ctrl+c"), km.Quit), "Quit must match 'ctrl+c'")
}

// TestFileListKeyMap_ImplementsHelpKeyMap verifies FileListKeyMap implements help.KeyMap.
func TestFileListKeyMap_ImplementsHelpKeyMap(t *testing.T) {
	var _ help.KeyMap = keys.FileListKeyMap{}
	short := keys.DefaultFileListKeyMap.ShortHelp()
	assert.NotEmpty(t, short, "ShortHelp must return at least one binding")
	full := keys.DefaultFileListKeyMap.FullHelp()
	assert.NotEmpty(t, full, "FullHelp must return at least one group")
}

// TestDetailKeyMap_ImplementsHelpKeyMap verifies DetailKeyMap implements help.KeyMap.
func TestDetailKeyMap_ImplementsHelpKeyMap(t *testing.T) {
	var _ help.KeyMap = keys.DetailKeyMap{}
	short := keys.DefaultDetailKeyMap.ShortHelp()
	assert.NotEmpty(t, short, "ShortHelp must return at least one binding")
	full := keys.DefaultDetailKeyMap.FullHelp()
	assert.NotEmpty(t, full, "FullHelp must return at least one group")
}

// TestFileListKeyMap_SearchAndInfoBindings verifies Search and Info bindings on DefaultFileListKeyMap.
func TestFileListKeyMap_SearchAndInfoBindings(t *testing.T) {
	km := keys.DefaultFileListKeyMap

	assert.True(t, key.Matches(press("/"), km.Search), "Search must match '/'")
	assert.True(t, key.Matches(press("i"), km.Info), "Info must match 'i'")
}

// TestDetailKeyMap_SearchAndInfoBindings verifies Search and Info bindings on DefaultDetailKeyMap.
func TestDetailKeyMap_SearchAndInfoBindings(t *testing.T) {
	km := keys.DefaultDetailKeyMap

	assert.True(t, key.Matches(press("/"), km.Search), "Search must match '/'")
	assert.True(t, key.Matches(press("i"), km.Info), "Info must match 'i'")
}

// TestHelpKeyMap_ImplementsHelpKeyMap verifies HelpKeyMap implements help.KeyMap.
func TestHelpKeyMap_ImplementsHelpKeyMap(t *testing.T) {
	var _ help.KeyMap = keys.HelpKeyMap{}
	short := keys.DefaultHelpKeyMap.ShortHelp()
	assert.NotEmpty(t, short, "ShortHelp must return at least one binding")
	full := keys.DefaultHelpKeyMap.FullHelp()
	assert.NotEmpty(t, full, "FullHelp must return at least one group")
}

// TestDiffKeyMap_ImplementsHelpKeyMap verifies DiffKeyMap implements help.KeyMap.
func TestDiffKeyMap_ImplementsHelpKeyMap(t *testing.T) {
	var _ help.KeyMap = keys.DiffKeyMap{}
	short := keys.DefaultDiffKeyMap.ShortHelp()
	assert.NotEmpty(t, short, "ShortHelp must return at least one binding")
	full := keys.DefaultDiffKeyMap.FullHelp()
	assert.NotEmpty(t, full, "FullHelp must return at least one group")
}

// TestHealthKeyMap_ImplementsHelpKeyMap verifies HealthKeyMap implements help.KeyMap.
func TestHealthKeyMap_ImplementsHelpKeyMap(t *testing.T) {
	var _ help.KeyMap = keys.HealthKeyMap{}
	short := keys.DefaultHealthKeyMap.ShortHelp()
	assert.NotEmpty(t, short, "ShortHelp must return at least one binding")
	full := keys.DefaultHealthKeyMap.FullHelp()
	assert.NotEmpty(t, full, "FullHelp must return at least one group")
}

// TestHistoryKeyMap_ImplementsHelpKeyMap verifies HistoryKeyMap implements help.KeyMap.
func TestHistoryKeyMap_ImplementsHelpKeyMap(t *testing.T) {
	var _ help.KeyMap = keys.HistoryKeyMap{}
	short := keys.DefaultHistoryKeyMap.ShortHelp()
	assert.NotEmpty(t, short, "ShortHelp must return at least one binding")
	full := keys.DefaultHistoryKeyMap.FullHelp()
	assert.NotEmpty(t, full, "FullHelp must return at least one group")
}

// TestMetadataKeyMap_ImplementsHelpKeyMap verifies MetadataKeyMap implements help.KeyMap.
func TestMetadataKeyMap_ImplementsHelpKeyMap(t *testing.T) {
	var _ help.KeyMap = keys.MetadataKeyMap{}
	short := keys.DefaultMetadataKeyMap.ShortHelp()
	assert.NotEmpty(t, short, "ShortHelp must return at least one binding")
	full := keys.DefaultMetadataKeyMap.FullHelp()
	assert.NotEmpty(t, full, "FullHelp must return at least one group")
}

// TestRecipientFormKeyMap_ImplementsHelpKeyMap verifies RecipientFormKeyMap implements help.KeyMap.
func TestRecipientFormKeyMap_ImplementsHelpKeyMap(t *testing.T) {
	var _ help.KeyMap = keys.RecipientFormKeyMap{}
	short := keys.DefaultRecipientFormKeyMap.ShortHelp()
	assert.NotEmpty(t, short, "ShortHelp must return at least one binding")
	full := keys.DefaultRecipientFormKeyMap.FullHelp()
	assert.NotEmpty(t, full, "FullHelp must return at least one group")
}

// TestFileListSearchKeyMap_ImplementsHelpKeyMap verifies FileListSearchKeyMap implements help.KeyMap.
func TestFileListSearchKeyMap_ImplementsHelpKeyMap(t *testing.T) {
	var _ help.KeyMap = keys.FileListSearchKeyMap{}
	short := keys.DefaultFileListSearchKeyMap.ShortHelp()
	assert.NotEmpty(t, short, "ShortHelp must return at least one binding")
	full := keys.DefaultFileListSearchKeyMap.FullHelp()
	assert.NotEmpty(t, full, "FullHelp must return at least one group")
}

// TestRecipientConfirmKeyMap_ImplementsHelpKeyMap verifies RecipientConfirmKeyMap implements help.KeyMap.
func TestRecipientConfirmKeyMap_ImplementsHelpKeyMap(t *testing.T) {
	var _ help.KeyMap = keys.RecipientConfirmKeyMap{}
	short := keys.DefaultRecipientConfirmKeyMap.ShortHelp()
	assert.NotEmpty(t, short, "ShortHelp must return at least one binding")
	full := keys.DefaultRecipientConfirmKeyMap.FullHelp()
	assert.NotEmpty(t, full, "FullHelp must return at least one group")
}

// TestBulkReKeyConfirmKeyMap_ImplementsHelpKeyMap verifies BulkReKeyConfirmKeyMap implements help.KeyMap.
func TestBulkReKeyConfirmKeyMap_ImplementsHelpKeyMap(t *testing.T) {
	var _ help.KeyMap = keys.BulkReKeyConfirmKeyMap{}
	short := keys.DefaultBulkReKeyConfirmKeyMap.ShortHelp()
	assert.NotEmpty(t, short, "ShortHelp must return at least one binding")
	full := keys.DefaultBulkReKeyConfirmKeyMap.FullHelp()
	assert.NotEmpty(t, full, "FullHelp must return at least one group")
}

// TestRecipientListKeyMap_ImplementsHelpKeyMap verifies RecipientListKeyMap implements help.KeyMap.
func TestRecipientListKeyMap_ImplementsHelpKeyMap(t *testing.T) {
	var _ help.KeyMap = keys.RecipientListKeyMap{}
	short := keys.DefaultRecipientListKeyMap.ShortHelp()
	assert.NotEmpty(t, short, "ShortHelp must return at least one binding")
	full := keys.DefaultRecipientListKeyMap.FullHelp()
	assert.NotEmpty(t, full, "FullHelp must return at least one group")
}

// TestFormatMenuKeyMap_ImplementsHelpKeyMap verifies FormatMenuKeyMap implements help.KeyMap.
func TestFormatMenuKeyMap_ImplementsHelpKeyMap(t *testing.T) {
	var _ help.KeyMap = keys.FormatMenuKeyMap{}
	short := keys.DefaultFormatMenuKeyMap.ShortHelp()
	assert.NotEmpty(t, short, "ShortHelp must return at least one binding")
	full := keys.DefaultFormatMenuKeyMap.FullHelp()
	assert.NotEmpty(t, full, "FullHelp must return at least one group")
}
