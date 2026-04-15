// Package detailpane renders log entry detail views with syntax highlighting.
package detailpane

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
)

// RenderJSON renders a JSONL entry as indented, syntax-highlighted JSON text.
// Fields listed in hiddenFields are omitted. Uses the active theme for colors.
func RenderJSON(entry logsource.Entry, th theme.Theme, hiddenFields []string) string {
	hidden := make(map[string]bool, len(hiddenFields))
	for _, f := range hiddenFields {
		hidden[f] = true
	}

	// Unmarshal the raw JSON to get all original fields and values.
	var rawObj map[string]json.RawMessage
	if err := json.Unmarshal(entry.Raw, &rawObj); err != nil {
		// Fallback: show raw bytes as plain text.
		return string(entry.Raw)
	}

	// Render as ordered JSON: known fields first (time, level, msg, logger, thread),
	// then extra fields sorted alphabetically. Skip hidden fields.
	var sb strings.Builder
	sb.WriteString("{\n")

	orderedKeys := knownKeyOrder(rawObj, hidden)
	for i, key := range orderedKeys {
		rawVal := rawObj[key]
		var v interface{}
		if err := json.Unmarshal(rawVal, &v); err != nil {
			v = string(rawVal)
		}
		keyStyle := lipgloss.NewStyle().Foreground(th.SyntaxKey)
		escapedKey, _ := json.Marshal(key)
		sb.WriteString("  ")
		sb.WriteString(keyStyle.Render(string(escapedKey)))
		sb.WriteString(": ")
		renderValue(&sb, v, th, 1)
		if i < len(orderedKeys)-1 {
			sb.WriteByte(',')
		}
		sb.WriteByte('\n')
	}
	sb.WriteString("}\n")
	return sb.String()
}

// RenderRaw renders a non-JSON entry as plain text with no formatting.
func RenderRaw(entry logsource.Entry) string {
	return string(entry.Raw)
}

var preferredKeyOrder = []string{"time", "timestamp", "level", "severity", "msg", "message", "logger", "thread"}

func knownKeyOrder(obj map[string]json.RawMessage, hidden map[string]bool) []string {
	seen := make(map[string]bool)
	var keys []string

	for _, k := range preferredKeyOrder {
		if _, exists := obj[k]; exists && !hidden[k] {
			keys = append(keys, k)
			seen[k] = true
		}
	}

	// Remaining keys sorted alphabetically.
	var rest []string
	for k := range obj {
		if !seen[k] && !hidden[k] {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)
	return append(keys, rest...)
}

const indent = "  "

func renderValue(sb *strings.Builder, v interface{}, th theme.Theme, depth int) {
	switch val := v.(type) {
	case map[string]interface{}:
		renderObject(sb, val, th, depth)
	case []interface{}:
		renderArray(sb, val, th, depth)
	case string:
		s := lipgloss.NewStyle().Foreground(th.SyntaxString)
		b, _ := json.Marshal(val)
		sb.WriteString(s.Render(string(b)))
	case float64:
		s := lipgloss.NewStyle().Foreground(th.SyntaxNumber)
		sb.WriteString(s.Render(formatNumber(val)))
	case bool:
		s := lipgloss.NewStyle().Foreground(th.SyntaxBoolean)
		if val {
			sb.WriteString(s.Render("true"))
		} else {
			sb.WriteString(s.Render("false"))
		}
	case nil:
		s := lipgloss.NewStyle().Foreground(th.SyntaxNull)
		sb.WriteString(s.Render("null"))
	default:
		b, _ := json.Marshal(val)
		s := lipgloss.NewStyle().Foreground(th.SyntaxString)
		sb.WriteString(s.Render(string(b)))
	}
}

func renderObject(sb *strings.Builder, m map[string]interface{}, th theme.Theme, depth int) {
	if len(m) == 0 {
		sb.WriteString("{}")
		return
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sb.WriteString("{\n")
	prefix := strings.Repeat(indent, depth+1)
	keyStyle := lipgloss.NewStyle().Foreground(th.SyntaxKey)
	for i, k := range keys {
		escapedKey, _ := json.Marshal(k)
		sb.WriteString(prefix)
		sb.WriteString(keyStyle.Render(string(escapedKey)))
		sb.WriteString(": ")
		renderValue(sb, m[k], th, depth+1)
		if i < len(keys)-1 {
			sb.WriteByte(',')
		}
		sb.WriteByte('\n')
	}
	sb.WriteString(strings.Repeat(indent, depth))
	sb.WriteByte('}')
}

func renderArray(sb *strings.Builder, arr []interface{}, th theme.Theme, depth int) {
	if len(arr) == 0 {
		sb.WriteString("[]")
		return
	}
	sb.WriteString("[\n")
	prefix := strings.Repeat(indent, depth+1)
	for i, v := range arr {
		sb.WriteString(prefix)
		renderValue(sb, v, th, depth+1)
		if i < len(arr)-1 {
			sb.WriteByte(',')
		}
		sb.WriteByte('\n')
	}
	sb.WriteString(strings.Repeat(indent, depth))
	sb.WriteByte(']')
}

func formatNumber(f float64) string {
	if f == float64(int64(f)) {
		return fmt.Sprintf("%d", int64(f))
	}
	return fmt.Sprintf("%g", f)
}
