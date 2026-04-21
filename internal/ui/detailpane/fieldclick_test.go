package detailpane

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T-045: R8.1 — click on a field line emits FieldClickMsg with field and value pre-filled.
func TestFieldAtLine_StringValue(t *testing.T) {
	field, value := fieldAtLine(`  "level": "INFO"`)
	assert.Equal(t, "level", field, "field")
	assert.Equal(t, "INFO", value, "value")
}

func TestFieldAtLine_NumberValue(t *testing.T) {
	field, value := fieldAtLine(`  "count": 42`)
	assert.Equal(t, "count", field, "field")
	assert.Equal(t, "42", value, "value")
}

func TestFieldAtLine_TrailingComma(t *testing.T) {
	field, value := fieldAtLine(`  "msg": "hello",`)
	assert.Equal(t, "msg", field, "field")
	assert.Equal(t, "hello", value, "value")
}

func TestFieldAtLine_NonFieldLine(t *testing.T) {
	field, value := fieldAtLine("{")
	assert.Emptyf(t, field, "non-field line field: %q", field)
	assert.Emptyf(t, value, "non-field line value: %q", value)
}

// T-045: R8.1 — handleMouseClick on a matching line emits FieldClickMsg.
func TestHandleMouseClick_EmitsFieldClickMsg(t *testing.T) {
	lines := []string{
		"{",
		`  "level": "ERROR"`,
		`  "msg": "fail"`,
		"}",
	}
	msg := tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
		X:      10,
		Y:      1, // line index 1 = "level": "ERROR"
	}
	cmd := handleMouseClick(msg, 0, lines)
	require.NotNil(t, cmd, "expected FieldClickMsg cmd")
	result := cmd()
	fc, ok := result.(FieldClickMsg)
	require.Truef(t, ok, "expected FieldClickMsg, got %T", result)
	assert.Equal(t, "level", fc.Field, "field")
	assert.Equal(t, "ERROR", fc.Value, "value")
}

// T-045: R8.2 — prompt allows choosing include/exclude (tested in prompt_test.go).
// Here we verify FieldClickMsg carries the right data for the prompt.
func TestHandleMouseClick_NonFieldLine_NoCmd(t *testing.T) {
	lines := []string{"{", "}"}
	msg := tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
		Y:      0,
	}
	cmd := handleMouseClick(msg, 0, lines)
	assert.Nil(t, cmd, "non-field line click should return nil cmd")
}
