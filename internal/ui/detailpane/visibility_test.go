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

// T-127 (F-020): hidden fields set via WithHiddenFields reach the JSON
// renderer through Open — the suppressed key must not appear in rawContent.
func TestPaneModel_Open_HonorsHiddenFields(t *testing.T) {
	entry := logsource.Entry{
		IsJSON:     true,
		LineNumber: 1,
		Time:       time.Now(),
		Level:      "INFO",
		Msg:        "hello",
		Raw:        []byte(`{"level":"INFO","msg":"hello","secret":"hunter2"}`),
	}
	pane := defaultPane(20).WithHiddenFields([]string{"secret"}).Open(entry)
	assert.NotContainsf(t, pane.rawContent, "secret", "raw content should not include suppressed field `secret`, got: %q", pane.rawContent)
	assert.NotContainsf(t, pane.rawContent, "hunter2", "raw content should not include suppressed value, got: %q", pane.rawContent)
	assert.Containsf(t, pane.rawContent, "hello", "raw content should still include non-suppressed fields, got: %q", pane.rawContent)
}

// T-127 (F-020): Rerender with an updated hiddenFields set re-renders
// the current entry without the newly suppressed field.
func TestPaneModel_Rerender_RemovesNewlyHiddenField(t *testing.T) {
	entry := logsource.Entry{
		IsJSON:     true,
		LineNumber: 1,
		Time:       time.Now(),
		Level:      "INFO",
		Msg:        "hello",
		Raw:        []byte(`{"level":"INFO","msg":"hello","secret":"hunter2"}`),
	}
	pane := defaultPane(20).Open(entry)
	require.Containsf(t, pane.rawContent, "secret", "precondition: raw content should include `secret` before hide, got: %q", pane.rawContent)
	pane = pane.WithHiddenFields([]string{"secret"}).Rerender()
	assert.NotContainsf(t, pane.rawContent, "secret", "raw content should not include suppressed field after Rerender, got: %q", pane.rawContent)
}

// T-127 (F-020): Rerender preserves scroll offset so toggling a field's
// visibility does not jump the viewport back to the top.
func TestPaneModel_Rerender_PreservesScrollOffset(t *testing.T) {
	// Build an entry with enough fields to guarantee scrolling room.
	raw := `{"a":"1","b":"2","c":"3","d":"4","e":"5","f":"6","g":"7","h":"8","i":"9","j":"10","k":"11","l":"12"}`
	entry := logsource.Entry{IsJSON: true, Time: time.Now(), Level: "INFO", Msg: "x", Raw: []byte(raw)}
	pane := defaultPane(6).SetWidth(40).Open(entry)
	// Scroll down a few lines so we can detect offset preservation.
	pane.scroll.offset = 3
	pane.scroll = pane.scroll.Clamp()
	offBefore := pane.scroll.offset
	if offBefore == 0 {
		t.Skipf("content is too short to scroll; test not applicable")
	}

	pane = pane.WithHiddenFields([]string{"a"}).Rerender()
	offAfter := pane.scroll.offset
	assert.NotEqualf(t, 0, offAfter, "Rerender jumped to top; expected to preserve offset ~%d, got 0", offBefore)
}

// T-127: Rerender on a closed pane is a safe no-op.
func TestPaneModel_Rerender_ClosedPaneNoOp(t *testing.T) {
	pane := defaultPane(20)
	out := pane.Rerender()
	assert.False(t, out.IsOpen(), "Rerender on closed pane must not open it")
}
