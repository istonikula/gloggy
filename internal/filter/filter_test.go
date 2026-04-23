package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterSetAddGetAll(t *testing.T) {
	fs := NewFilterSet()
	id1 := fs.Add(Filter{Field: "level", Pattern: "error", Mode: Include, Enabled: true})
	id2 := fs.Add(Filter{Field: "any", Pattern: "timeout", Mode: Exclude, Enabled: false})
	require.NotEqual(t, id1, id2, "IDs should be unique")
	assert.Len(t, fs.GetAll(), 2)
}

func TestFilterSetRemove(t *testing.T) {
	fs := NewFilterSet()
	id := fs.Add(Filter{Field: "msg", Pattern: "test", Mode: Include, Enabled: true})
	assert.True(t, fs.Remove(id), "Remove should return true")
	assert.False(t, fs.Remove(id), "Remove again should return false")
	assert.Empty(t, fs.GetAll(), "GetAll should be empty")
}

func TestFilterSetEnableDisable(t *testing.T) {
	fs := NewFilterSet()
	id := fs.Add(Filter{Field: "level", Pattern: "info", Mode: Include, Enabled: true})
	assert.True(t, fs.Disable(id), "Disable should return true")
	assert.False(t, fs.GetAll()[0].Enabled, "should be disabled")
	// Retained even when disabled
	assert.Len(t, fs.GetAll(), 1, "disabled filter should still be in GetAll")
	assert.True(t, fs.Enable(id), "Enable should return true")
	assert.True(t, fs.GetAll()[0].Enabled, "should be re-enabled")
}

func TestFilterSetGetEnabled(t *testing.T) {
	fs := NewFilterSet()
	fs.Add(Filter{Field: "a", Enabled: true})
	id2 := fs.Add(Filter{Field: "b", Enabled: true})
	fs.Add(Filter{Field: "c", Enabled: false})

	require.Len(t, fs.GetEnabled(), 2)
	fs.Disable(id2)
	require.Len(t, fs.GetEnabled(), 1)
}

func TestModeTypedEnum(t *testing.T) {
	f := Filter{Mode: Exclude}
	assert.Equal(t, Exclude, f.Mode)
	assert.Equal(t, "include", Include.String())
	assert.Equal(t, "exclude", Exclude.String())
}

// ---------- V32 MANUAL-TOGGLE EXIT (B11) ----------

// TestFilterSet_Enable_DuringGloballyDisabled_ExitsStateMachine —
// while globallyDisabled=true, Enable(id) must clear globallyDisabled
// + savedEnabled so the next ToggleAll call starts a fresh 1st-press
// cycle from the post-toggle Enabled values. Without this, ToggleAll
// would hit the restore branch and silently clobber the user's
// manual toggle (B11).
func TestFilterSet_Enable_DuringGloballyDisabled_ExitsStateMachine_V32(t *testing.T) {
	fs := NewFilterSet()
	id1 := fs.Add(Filter{Field: "level", Pattern: "INFO", Mode: Include, Enabled: true})
	id2 := fs.Add(Filter{Field: "thread", Pattern: "bg", Mode: Include, Enabled: true})

	fs.ToggleAll()
	require.True(t, fs.IsGloballyDisabled(), "precondition: globallyDisabled after first ToggleAll")
	require.False(t, fs.GetAll()[0].Enabled, "precondition: filter 0 disabled")
	require.False(t, fs.GetAll()[1].Enabled, "precondition: filter 1 disabled")

	assert.True(t, fs.Enable(id1), "Enable returns true")

	assert.Truef(t, fs.GetAll()[0].Enabled,
		"filter 0 should be enabled after Enable(id1)")
	assert.Falsef(t, fs.IsGloballyDisabled(),
		"manual Enable during globallyDisabled must exit the state machine")
	assert.Nilf(t, fs.savedEnabled,
		"savedEnabled must be dropped on exit; snapshot is stale post-manual-toggle")

	// Fresh 1st-press cycle: ToggleAll saves current state ({true,false})
	// then disables all. Previous saved={true,true} must NOT be reused.
	fs.ToggleAll()
	require.True(t, fs.IsGloballyDisabled(), "after ToggleAll: globally-disabled again")
	require.Falsef(t, fs.GetAll()[0].Enabled, "filter 0 disabled by fresh 1st-press")
	require.Falsef(t, fs.GetAll()[1].Enabled, "filter 1 disabled by fresh 1st-press")

	// Restore — must match the post-manual-toggle snapshot ({true,false}),
	// not the pre-manual-toggle snapshot ({true,true}).
	fs.ToggleAll()
	assert.Truef(t, fs.GetAll()[0].Enabled,
		"restore must reflect post-manual-toggle state for filter 0")
	assert.Falsef(t, fs.GetAll()[1].Enabled,
		"restore must reflect post-manual-toggle state for filter 1 — B11 repro: this was true before fix")
	_ = id2
}

// TestFilterSet_Disable_DuringGloballyDisabled_ExitsStateMachine —
// symmetric to Enable. Manual Disable while globallyDisabled=true
// also exits the state machine.
func TestFilterSet_Disable_DuringGloballyDisabled_ExitsStateMachine_V32(t *testing.T) {
	fs := NewFilterSet()
	id1 := fs.Add(Filter{Field: "level", Pattern: "INFO", Mode: Include, Enabled: true})
	_ = fs.Add(Filter{Field: "thread", Pattern: "bg", Mode: Include, Enabled: true})

	fs.ToggleAll()
	require.True(t, fs.IsGloballyDisabled(), "precondition: globally-disabled")

	fs.Enable(id1)                                           // first exits state machine
	require.False(t, fs.IsGloballyDisabled(), "precondition: state machine exited")

	fs.ToggleAll()                                           // save {true,false} + disable all
	require.True(t, fs.IsGloballyDisabled(), "precondition: re-entered global-disabled")

	assert.True(t, fs.Disable(id1), "Disable returns true")
	assert.Falsef(t, fs.IsGloballyDisabled(),
		"manual Disable during globally-disabled must exit the state machine")
	assert.Nilf(t, fs.savedEnabled, "savedEnabled must be dropped on exit")
}

// TestFilterSet_EnableDisable_OutsideGloballyDisabled_NoStateChange —
// Enable/Disable in the normal (non-globally-disabled) state must not
// touch globallyDisabled or savedEnabled. Guards the guarded-clause
// from regressing to "always clear" which would waste the allocation
// and obscure intent.
func TestFilterSet_EnableDisable_OutsideGloballyDisabled_NoStateChange(t *testing.T) {
	fs := NewFilterSet()
	id := fs.Add(Filter{Field: "level", Pattern: "INFO", Mode: Include, Enabled: true})

	require.False(t, fs.IsGloballyDisabled(), "precondition: not globally-disabled")
	require.Nil(t, fs.savedEnabled, "precondition: no snapshot")

	fs.Disable(id)
	assert.False(t, fs.IsGloballyDisabled(), "Disable outside must not set globallyDisabled")
	assert.Nil(t, fs.savedEnabled, "Disable outside must not create savedEnabled")

	fs.Enable(id)
	assert.False(t, fs.IsGloballyDisabled(), "Enable outside must not set globallyDisabled")
	assert.Nil(t, fs.savedEnabled, "Enable outside must not create savedEnabled")
}

// TestFilterSet_Add_DuringGloballyDisabled_UnchangedByT31 — V32's
// "add-while-disabled" behavior (new filter saves its Add-time
// Enabled to savedEnabled + disables it) MUST NOT be affected by
// T31's Enable/Disable changes. Guards that the exit-rule fix does
// not leak into unrelated paths.
func TestFilterSet_Add_DuringGloballyDisabled_UnchangedByT31_V32(t *testing.T) {
	fs := NewFilterSet()
	_ = fs.Add(Filter{Field: "level", Pattern: "INFO", Mode: Include, Enabled: true})

	fs.ToggleAll()
	require.True(t, fs.IsGloballyDisabled(), "precondition: globally-disabled")

	newID := fs.Add(Filter{Field: "msg", Pattern: "boom", Mode: Include, Enabled: true})

	assert.Truef(t, fs.IsGloballyDisabled(),
		"Add during globallyDisabled must NOT exit the state machine (only Enable/Disable do)")
	// Find the added filter and verify it was force-disabled on Add.
	for i, fid := range fs.ids {
		if fid == newID {
			assert.Falsef(t, fs.filters[i].Enabled,
				"Add during globallyDisabled must force Enabled=false on the new filter")
			break
		}
	}
}
