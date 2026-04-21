package detailpane

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
)

func testVisEntry() logsource.Entry {
	return logsource.Entry{
		IsJSON: true,
		Time:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Level:  "INFO",
		Msg:    "hello",
		Logger: "app",
		Raw:    []byte(`{"time":"2024-01-01T12:00:00Z","level":"INFO","msg":"hello","logger":"app"}`),
	}
}

// T-038: R5.1 — hidden field not in render output
func TestVisibilityModel_HiddenFieldOmitted(t *testing.T) {
	lr := config.LoadResult{Config: config.DefaultConfig()}
	lr.Config.HiddenFields = []string{"level"}
	m := NewVisibilityModel("", lr)
	result := m.RenderEntry(testVisEntry(), theme.GetTheme("tokyo-night"))
	assert.NotContains(t, result, `"level"`, "hidden field 'level' should not appear in output")
	assert.Contains(t, result, `"msg"`, "visible field 'msg' should appear in output")
}

// T-038: R5.2 — toggle causes re-render without the hidden field
func TestVisibilityModel_ToggleHides(t *testing.T) {
	lr := config.LoadResult{Config: config.DefaultConfig()}
	m := NewVisibilityModel("", lr)

	// Initially msg is visible.
	result1 := m.RenderEntry(testVisEntry(), theme.GetTheme("tokyo-night"))
	assert.Contains(t, result1, `"msg"`, "msg should be visible initially")

	// Hide it.
	m2, err := m.ToggleField("msg")
	require.NoError(t, err, "ToggleField")
	result2 := m2.RenderEntry(testVisEntry(), theme.GetTheme("tokyo-night"))
	assert.NotContains(t, result2, `"msg"`, "msg should be hidden after toggle")

	// Toggle again — should show.
	m3, _ := m2.ToggleField("msg")
	result3 := m3.RenderEntry(testVisEntry(), theme.GetTheme("tokyo-night"))
	assert.Contains(t, result3, `"msg"`, "msg should be visible after second toggle")
}

// T-038: R5.3 + R5.4 — visibility change written to config file; persists after reload
func TestVisibilityModel_PersistedToConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	lr := config.Load(path) // creates default (file doesn't exist yet)

	m := NewVisibilityModel(path, lr)
	m2, err := m.ToggleField("level")
	require.NoError(t, err, "ToggleField")

	// Verify config file was written.
	_, err = os.Stat(path)
	require.NoError(t, err, "config file not created")

	// Reload config and verify field is still hidden.
	lr2 := config.Load(path)
	assert.Containsf(t, lr2.Config.HiddenFields, "level",
		"hidden field 'level' not persisted after reload; HiddenFields=%v", lr2.Config.HiddenFields)

	// Verify the new VisibilityModel from reloaded config omits the field.
	m3 := NewVisibilityModel(path, lr2)
	result := m3.RenderEntry(testVisEntry(), theme.GetTheme("tokyo-night"))
	assert.NotContains(t, result, `"level"`, "after restart, hidden field 'level' should still be omitted")

	_ = m2
}
