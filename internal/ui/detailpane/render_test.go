package detailpane

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"

	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
)

func init() {
	// Force TrueColor in tests so ANSI color codes are embedded in output.
	lipgloss.SetColorProfile(termenv.TrueColor)
}

// colorANSI renders a test string with the given color and extracts the
// foreground ANSI escape code (e.g. "38;2;115;218;202") from the output.
// This mirrors exactly what lipgloss/termenv will produce, avoiding any
// manual hex→int conversion that might differ by rounding.
func colorANSI(c lipgloss.Color) string {
	rendered := lipgloss.NewStyle().Foreground(c).Render("x")
	start := strings.Index(rendered, "\x1b[")
	if start == -1 {
		return string(c)
	}
	end := strings.Index(rendered[start:], "m")
	if end == -1 {
		return string(c)
	}
	return rendered[start+2 : start+end]
}

func jsonEntry() logsource.Entry {
	raw := []byte(`{"time":"2024-01-01T12:00:00Z","level":"INFO","msg":"hello world","count":42,"active":true,"data":null}`)
	return logsource.Entry{
		IsJSON:     true,
		Raw:        raw,
		LineNumber: 1,
		Time:       time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Level:      "INFO",
		Msg:        "hello world",
		Extra: map[string]json.RawMessage{
			"count":  json.RawMessage(`42`),
			"active": json.RawMessage(`true`),
			"data":   json.RawMessage(`null`),
		},
	}
}

// T-035: R2.1 — JSONL entry renders as indented JSON
func TestRenderJSON_IsIndented(t *testing.T) {
	result := RenderJSON(jsonEntry(), theme.GetTheme("tokyo-night"), nil)
	assert.Contains(t, result, "{\n", "expected indented JSON with newlines")
	assert.Contains(t, result, "  ", "expected indentation spaces")
}

// T-035: R2.2 — All fields present including extra
func TestRenderJSON_AllFieldsPresent(t *testing.T) {
	result := RenderJSON(jsonEntry(), theme.GetTheme("tokyo-night"), nil)
	for _, field := range []string{"time", "level", "msg", "count", "active", "data"} {
		assert.Containsf(t, result, `"`+field+`"`, "expected field %q in output", field)
	}
}

// T-035: R2.3 — Key color in ANSI output
func TestRenderJSON_KeyColor(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	result := RenderJSON(jsonEntry(), th, nil)
	want := colorANSI(th.SyntaxKey)
	assert.Containsf(t, result, want, "expected SyntaxKey ANSI code %s in output", want)
}

// T-035: R2.4 — String value color
func TestRenderJSON_StringColor(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	result := RenderJSON(jsonEntry(), th, nil)
	want := colorANSI(th.SyntaxString)
	assert.Containsf(t, result, want, "expected SyntaxString ANSI code %s in output", want)
}

// T-035: R2.5 — Number value color
func TestRenderJSON_NumberColor(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	result := RenderJSON(jsonEntry(), th, nil)
	want := colorANSI(th.SyntaxNumber)
	assert.Containsf(t, result, want, "expected SyntaxNumber ANSI code %s in output", want)
}

// T-035: R2.6 — Boolean value color
func TestRenderJSON_BoolColor(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	result := RenderJSON(jsonEntry(), th, nil)
	want := colorANSI(th.SyntaxBoolean)
	assert.Containsf(t, result, want, "expected SyntaxBoolean ANSI code %s in output", want)
}

// T-035: R2.7 — Null value color
func TestRenderJSON_NullColor(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	result := RenderJSON(jsonEntry(), th, nil)
	want := colorANSI(th.SyntaxNull)
	assert.Containsf(t, result, want, "expected SyntaxNull ANSI code %s in output", want)
}

// T-035: R2.8 — Theme switch changes ANSI codes
func TestRenderJSON_ThemeSwitch(t *testing.T) {
	entry := jsonEntry()
	th1 := theme.GetTheme("tokyo-night")
	th2 := theme.GetTheme("catppuccin-mocha")
	out1 := RenderJSON(entry, th1, nil)
	out2 := RenderJSON(entry, th2, nil)

	if string(th1.SyntaxKey) == string(th2.SyntaxKey) {
		t.Skip("themes have identical key color, skipping switch test")
	}
	assert.NotEqual(t, out1, out2, "expected different output when switching themes")
	want := colorANSI(th2.SyntaxKey)
	assert.Containsf(t, out2, want, "catppuccin-mocha output missing its key ANSI code %s", want)
}

// T-035: hidden fields are omitted
func TestRenderJSON_HiddenFields(t *testing.T) {
	result := RenderJSON(jsonEntry(), theme.GetTheme("tokyo-night"), []string{"level", "count"})
	assert.NotContains(t, result, `"level"`, "hidden field 'level' should not appear in output")
	assert.NotContains(t, result, `"count"`, "hidden field 'count' should not appear in output")
	assert.Contains(t, result, `"msg"`, "non-hidden field 'msg' should appear in output")
}

// T-036: R3.1 — Non-JSON entry displays as plain raw text
func TestRenderRaw_PlainText(t *testing.T) {
	entry := logsource.Entry{
		IsJSON: false,
		Raw:    []byte("2024-01-01 ERROR could not connect to database"),
	}
	result := RenderRaw(entry)
	assert.Equal(t, "2024-01-01 ERROR could not connect to database", result)
}

// T-036: R3.2 — No JSON formatting applied to non-JSON entries
func TestRenderRaw_NoANSI(t *testing.T) {
	entry := logsource.Entry{
		IsJSON: false,
		Raw:    []byte("plain text {not json}"),
	}
	result := RenderRaw(entry)
	// Result must not contain ANSI escape sequences.
	assert.NotContains(t, result, "\x1b[", "expected no ANSI escape codes in raw output")
	assert.Equal(t, "plain text {not json}", result)
}
