package keys_test

import (
	"testing"

	"charm.land/bubbles/v2/key"
	"github.com/caesarakalaeii/sops-tui/internal/keys"
	"github.com/stretchr/testify/assert"
)

// TestDetailKeyMapHasReveal verifies that DetailKeyMap has a Reveal binding for "r".
func TestDetailKeyMapHasReveal(t *testing.T) {
	km := keys.DefaultDetailKeyMap
	assert.True(t, key.Matches(press("r"), km.Reveal), "Reveal must match 'r'")
}

// TestDetailKeyMapHasRevealAll verifies that DetailKeyMap has a RevealAll binding for "R".
func TestDetailKeyMapHasRevealAll(t *testing.T) {
	km := keys.DefaultDetailKeyMap
	assert.True(t, key.Matches(press("R"), km.RevealAll), "RevealAll must match 'R'")
}

// TestDetailKeyMapHasEdit verifies that DetailKeyMap has an Edit binding for "e".
func TestDetailKeyMapHasEdit(t *testing.T) {
	km := keys.DefaultDetailKeyMap
	assert.True(t, key.Matches(press("e"), km.Edit), "Edit must match 'e'")
}

// TestDetailKeyMapHasEditFile verifies that DetailKeyMap has an EditFile binding for "E".
func TestDetailKeyMapHasEditFile(t *testing.T) {
	km := keys.DefaultDetailKeyMap
	assert.True(t, key.Matches(press("E"), km.EditFile), "EditFile must match 'E'")
}

// TestDetailKeyMapHasRotate verifies that DetailKeyMap has a Rotate binding for "X".
func TestDetailKeyMapHasRotate(t *testing.T) {
	km := keys.DefaultDetailKeyMap
	assert.True(t, key.Matches(press("X"), km.Rotate), "Rotate must match 'X'")
}

// TestDetailKeyMapRevealInFullHelp verifies reveal/mask bindings appear in FullHelp.
func TestDetailKeyMapRevealInFullHelp(t *testing.T) {
	km := keys.DefaultDetailKeyMap
	full := km.FullHelp()
	// Find the actions group and verify reveal-related bindings are included
	found := false
	for _, group := range full {
		for _, b := range group {
			keys := b.Keys()
			for _, k := range keys {
				if k == "r" || k == "R" {
					found = true
				}
			}
		}
	}
	assert.True(t, found, "FullHelp must include reveal/revealAll bindings (r or R key)")
}
