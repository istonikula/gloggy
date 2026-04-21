package integration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
	"github.com/istonikula/gloggy/internal/ui/detailpane"
)

// T-059: config/R6, detail-pane/R5 — hide field → config updated → restart → field still hidden.
func TestConfigWriteBack_HiddenFieldPersists(t *testing.T) {
	// Create a temp config file.
	f, err := os.CreateTemp("", "gloggy-config-*.toml")
	require.NoError(t, err)
	cfgPath := f.Name()
	f.Close()
	defer os.Remove(cfgPath)

	// Load config from path (will use defaults since file is empty).
	lr := config.Load(cfgPath)

	// Create a VisibilityModel and toggle "level" field to hidden.
	vm := detailpane.NewVisibilityModel(cfgPath, lr)
	vm2, err := vm.ToggleField("level")
	require.NoError(t, err, "ToggleField")

	// Verify the field is in the hidden list.
	hidden := vm2.HiddenFields()
	assert.Containsf(t, hidden, "level", "expected 'level' in hidden fields after toggle, got: %v", hidden)

	// "Restart": load config again from the same file.
	lr2 := config.Load(cfgPath)

	// Verify the field is still hidden after reload.
	vm3 := detailpane.NewVisibilityModel(cfgPath, lr2)
	hidden2 := vm3.HiddenFields()
	assert.Containsf(t, hidden2, "level", "expected 'level' to remain hidden after config reload, got: %v", hidden2)

	// Verify the field is omitted from the detail pane render.
	entry := logsource.Entry{
		IsJSON:     true,
		LineNumber: 1,
		Level:      "ERROR",
		Msg:        "test",
		Raw:        []byte(`{"level":"ERROR","msg":"test"}`),
	}
	th := theme.GetTheme("tokyo-night")
	rendered := detailpane.RenderJSON(entry, th, hidden2)
	assert.NotContainsf(t, rendered, `"level"`, "hidden field 'level' should not appear in rendered output:\n%s", rendered)
}
