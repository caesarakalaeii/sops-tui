package keys_test

import (
	"testing"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"github.com/stretchr/testify/assert"

	"github.com/caesarakalaeii/sops-tui/internal/keys"
)

// TestDetailKeyMap_AddSecretBinding verifies the AddSecret action is bound to 'n'.
func TestDetailKeyMap_AddSecretBinding(t *testing.T) {
	km := keys.DefaultDetailKeyMap
	assert.True(t, key.Matches(press("n"), km.AddSecret), "AddSecret must match 'n'")
	assert.Equal(t, "n", km.AddSecret.Help().Key, "AddSecret menu mnemonic must be 'n'")
	assert.Equal(t, "add secret", km.AddSecret.Help().Desc, "AddSecret description")
}

// TestDetailKeyMap_AddSecretHiddenFromMenu verifies AddSecret joins Blame as a
// menu-suppressed-but-discoverable binding (14-entry ShortHelp, 12-slot menu cap).
func TestDetailKeyMap_AddSecretHiddenFromMenu(t *testing.T) {
	km := keys.DefaultDetailKeyMap
	hidden := km.HiddenFromMenu()
	hiddenKeys := make(map[string]bool, len(hidden))
	for _, b := range hidden {
		hiddenKeys[b.Help().Key] = true
	}
	assert.True(t, hiddenKeys["b"], "Blame (b) must stay hidden from the menu")
	assert.True(t, hiddenKeys["n"], "AddSecret (n) must be hidden from the menu")
}

// TestAddSecretFormKeyMap_Bindings verifies the add-secret form bindings.
func TestAddSecretFormKeyMap_Bindings(t *testing.T) {
	km := keys.DefaultAddSecretFormKeyMap
	assert.True(t, key.Matches(press("tab"), km.NextField), "NextField must match 'tab'")
	assert.True(t, key.Matches(press("enter"), km.Confirm), "Confirm must match 'enter'")
	assert.True(t, key.Matches(press("esc"), km.Cancel), "Cancel must match 'esc'")
}

// TestAddSecretFormKeyMap_ImplementsHelpKeyMap verifies the keymap implements help.KeyMap.
func TestAddSecretFormKeyMap_ImplementsHelpKeyMap(t *testing.T) {
	var _ help.KeyMap = keys.AddSecretFormKeyMap{}
	short := keys.DefaultAddSecretFormKeyMap.ShortHelp()
	assert.Len(t, short, 3, "ShortHelp returns NextField, Confirm, Cancel")
	full := keys.DefaultAddSecretFormKeyMap.FullHelp()
	assert.NotEmpty(t, full, "FullHelp must return at least one group")
}
