package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/ui/appshell"
)

// V43/B32: a failed config.Save must surface a UI notice, never silently drop.
// A bad config dir makes atomicWrite fail; the theme-cycle `t` triggers the
// save path.
func TestModel_V43_SaveConfigError_Surfaces_B32(t *testing.T) {
	m := New("", false, "/nonexistent-dir-t44-xyz/config.toml", testCfg())
	m = resize(m, 120, 30)
	require.False(t, m.keyhints.HasNotice(), "precondition: no notice")

	m = key(m, "t") // theme cycle → saveConfig → write fails

	assert.True(t, m.keyhints.HasNotice(), "failed config save must surface a notice")
	assert.Truef(t, containsSubstring(m.keyhints.View(), "config save failed"),
		"notice must name the failure; got %q", m.keyhints.View())
}

// V43/B35: a ratio key (+/-/=/|) pressed with no detail pane open has no
// divider to move — signal it instead of swallowing the keypress.
func TestModel_V43_RatioKeyNoPane_Surfaces_B35(t *testing.T) {
	m := newModel()
	m = resize(m, 120, 30)
	m = m.SetEntries(makeEntries(3))
	require.False(t, m.pane.IsOpen(), "precondition: pane closed")
	require.Equal(t, appshell.FocusEntryList, m.focus, "precondition: list focused")

	m = key(m, "+")

	assert.True(t, m.keyhints.HasNotice(), "ratio key with no pane must signal, not swallow")
	assert.Truef(t, containsSubstring(m.keyhints.View(), "no detail pane"),
		"notice must explain why the key did nothing; got %q", m.keyhints.View())
}

// V43/B35 negative: a ratio key while a list-search is in input mode is a
// literal query char, NOT a resize attempt — no "no detail pane" notice.
func TestModel_V43_RatioKeyDuringListSearch_NoNotice(t *testing.T) {
	m := newModel()
	m = resize(m, 120, 30)
	m = m.SetEntries(makeEntries(3))
	m = key(m, "/") // open list search (input mode)
	require.True(t, m.list.HasActiveSearch(), "precondition: list search active")

	m = key(m, "=") // ratio key — must be consumed as a query char

	assert.False(t, m.keyhints.HasNotice(),
		"ratio key in list-search input mode must not emit the no-pane notice")
}
