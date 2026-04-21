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
