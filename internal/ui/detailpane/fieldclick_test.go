package detailpane

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// T-045: R8.1 — click on a field line emits FieldClickMsg with field and value pre-filled.
func TestFieldAtLine_StringValue(t *testing.T) {
	field, value := fieldAtLine(`  "level": "INFO"`)
	if field != "level" {
		t.Errorf("field: got %q, want %q", field, "level")
	}
	if value != "INFO" {
		t.Errorf("value: got %q, want %q", value, "INFO")
	}
}

func TestFieldAtLine_NumberValue(t *testing.T) {
	field, value := fieldAtLine(`  "count": 42`)
	if field != "count" {
		t.Errorf("field: got %q, want %q", field, "count")
	}
	if value != "42" {
		t.Errorf("value: got %q, want %q", value, "42")
	}
}

func TestFieldAtLine_TrailingComma(t *testing.T) {
	field, value := fieldAtLine(`  "msg": "hello",`)
	if field != "msg" {
		t.Errorf("field: got %q, want %q", field, "msg")
	}
	if value != "hello" {
		t.Errorf("value: got %q, want %q", value, "hello")
	}
}

func TestFieldAtLine_NonFieldLine(t *testing.T) {
	field, value := fieldAtLine("{")
	if field != "" || value != "" {
		t.Errorf("non-field line should return empty, got %q %q", field, value)
	}
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
	if cmd == nil {
		t.Fatal("expected FieldClickMsg cmd")
	}
	result := cmd()
	fc, ok := result.(FieldClickMsg)
	if !ok {
		t.Fatalf("expected FieldClickMsg, got %T", result)
	}
	if fc.Field != "level" {
		t.Errorf("field: got %q, want %q", fc.Field, "level")
	}
	if fc.Value != "ERROR" {
		t.Errorf("value: got %q, want %q", fc.Value, "ERROR")
	}
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
	if cmd != nil {
		t.Error("non-field line click should return nil cmd")
	}
}
