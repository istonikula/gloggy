package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// V40: every public mutation is safe when called alone on a freshly
// constructed FilterSet — no method assumes ToggleAll initialized the
// internal maps first (B24: nil savedEnabled write).
func TestNewFilterSet_ShotgunMutations_NoPanic_V40(t *testing.T) {
	sample := Filter{Field: "level", Pattern: "error", Mode: Include, Enabled: true}

	mutations := map[string]func(fs *FilterSet){
		"Add":       func(fs *FilterSet) { fs.Add(sample) },
		"Remove":    func(fs *FilterSet) { fs.Remove(0) },
		"Enable":    func(fs *FilterSet) { fs.Enable(0) },
		"Disable":   func(fs *FilterSet) { fs.Disable(0) },
		"ToggleAll": func(fs *FilterSet) { fs.ToggleAll() },
		"Update":    func(fs *FilterSet) { fs.Update(0, sample) },
	}

	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			fs := NewFilterSet()
			assert.NotPanics(t, func() { mutate(fs) },
				"%s on a fresh FilterSet must not panic", name)
		})
	}
}

// V40/B24: the Add-while-globally-disabled path writes savedEnabled without a
// nil-map panic even when ToggleAll ran first on an initially-empty set.
func TestFilterSet_AddDuringGlobalDisable_Safe_V40(t *testing.T) {
	fs := NewFilterSet()
	fs.ToggleAll() // globallyDisabled = true on an empty set
	require.True(t, fs.IsGloballyDisabled())

	var id int
	assert.NotPanics(t, func() {
		id = fs.Add(Filter{Field: "msg", Pattern: "boom", Enabled: true})
	})

	// The added filter is disabled while globally muted; restoring brings back
	// its add-time Enabled value.
	all := fs.GetAll()
	require.Len(t, all, 1)
	assert.False(t, all[0].Enabled, "new filter disabled under global mute")

	fs.ToggleAll() // restore
	_ = id
	all = fs.GetAll()
	assert.True(t, all[0].Enabled, "restored to add-time Enabled=true")
}
