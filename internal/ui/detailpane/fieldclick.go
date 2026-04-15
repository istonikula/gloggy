package detailpane

import (
	"encoding/json"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// FieldClickMsg is emitted when the user clicks on a field value in the detail pane.
// Field and Value are pre-filled for use by a filter prompt.
type FieldClickMsg struct {
	Field string
	Value string
}

// fieldAtLine parses a rendered JSON detail pane line to extract the field name and
// raw string value. Lines are expected to be in the form:
//
//	  "key": <value>
//
// Returns ("", "") if the line does not match this pattern.
func fieldAtLine(line string) (field, value string) {
	line = strings.TrimSpace(line)
	// Must start with a quoted key.
	if len(line) == 0 || line[0] != '"' {
		return "", ""
	}
	// Find closing quote of key.
	end := strings.Index(line[1:], `"`)
	if end < 0 {
		return "", ""
	}
	field = line[1 : end+1]
	// After key: `: ` then value.
	rest := strings.TrimSpace(line[end+2:])
	rest = strings.TrimPrefix(rest, ":")
	rest = strings.TrimSpace(rest)
	// Strip trailing comma.
	rest = strings.TrimSuffix(rest, ",")
	// If value is a JSON string, unquote it.
	if len(rest) >= 2 && rest[0] == '"' {
		var s string
		if err := json.Unmarshal([]byte(rest), &s); err == nil {
			return field, s
		}
	}
	// Return raw value (number, bool, null, nested object/array preview).
	return field, rest
}

// handleMouseClick checks whether msg is a left-click on a content line and, if so,
// returns a FieldClickMsg command for the clicked field.
// lineOffset is the scroll offset so we can map Y → absolute line index.
// lines are the raw content lines (before ANSI).
func handleMouseClick(msg tea.MouseMsg, lineOffset int, lines []string) tea.Cmd {
	if msg.Button != tea.MouseButtonLeft || msg.Action != tea.MouseActionPress {
		return nil
	}
	absLine := lineOffset + msg.Y
	if absLine < 0 || absLine >= len(lines) {
		return nil
	}
	field, value := fieldAtLine(lines[absLine])
	if field == "" {
		return nil
	}
	return func() tea.Msg { return FieldClickMsg{Field: field, Value: value} }
}
