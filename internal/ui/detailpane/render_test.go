package detailpane

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
	"github.com/muesli/termenv"
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
	if !strings.Contains(result, "{\n") {
		t.Error("expected indented JSON with newlines")
	}
	if !strings.Contains(result, "  ") {
		t.Error("expected indentation spaces")
	}
}

// T-035: R2.2 — All fields present including extra
func TestRenderJSON_AllFieldsPresent(t *testing.T) {
	result := RenderJSON(jsonEntry(), theme.GetTheme("tokyo-night"), nil)
	for _, field := range []string{"time", "level", "msg", "count", "active", "data"} {
		if !strings.Contains(result, `"`+field+`"`) {
			t.Errorf("expected field %q in output", field)
		}
	}
}

// T-035: R2.3 — Key color in ANSI output
func TestRenderJSON_KeyColor(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	result := RenderJSON(jsonEntry(), th, nil)
	want := colorANSI(th.SyntaxKey)
	if !strings.Contains(result, want) {
		t.Errorf("expected SyntaxKey ANSI code %s in output", want)
	}
}

// T-035: R2.4 — String value color
func TestRenderJSON_StringColor(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	result := RenderJSON(jsonEntry(), th, nil)
	want := colorANSI(th.SyntaxString)
	if !strings.Contains(result, want) {
		t.Errorf("expected SyntaxString ANSI code %s in output", want)
	}
}

// T-035: R2.5 — Number value color
func TestRenderJSON_NumberColor(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	result := RenderJSON(jsonEntry(), th, nil)
	want := colorANSI(th.SyntaxNumber)
	if !strings.Contains(result, want) {
		t.Errorf("expected SyntaxNumber ANSI code %s in output", want)
	}
}

// T-035: R2.6 — Boolean value color
func TestRenderJSON_BoolColor(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	result := RenderJSON(jsonEntry(), th, nil)
	want := colorANSI(th.SyntaxBoolean)
	if !strings.Contains(result, want) {
		t.Errorf("expected SyntaxBoolean ANSI code %s in output", want)
	}
}

// T-035: R2.7 — Null value color
func TestRenderJSON_NullColor(t *testing.T) {
	th := theme.GetTheme("tokyo-night")
	result := RenderJSON(jsonEntry(), th, nil)
	want := colorANSI(th.SyntaxNull)
	if !strings.Contains(result, want) {
		t.Errorf("expected SyntaxNull ANSI code %s in output", want)
	}
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
	if out1 == out2 {
		t.Error("expected different output when switching themes")
	}
	if !strings.Contains(out2, colorANSI(th2.SyntaxKey)) {
		t.Errorf("catppuccin-mocha output missing its key ANSI code %s", colorANSI(th2.SyntaxKey))
	}
}

// T-035: hidden fields are omitted
func TestRenderJSON_HiddenFields(t *testing.T) {
	result := RenderJSON(jsonEntry(), theme.GetTheme("tokyo-night"), []string{"level", "count"})
	if strings.Contains(result, `"level"`) {
		t.Error("hidden field 'level' should not appear in output")
	}
	if strings.Contains(result, `"count"`) {
		t.Error("hidden field 'count' should not appear in output")
	}
	if !strings.Contains(result, `"msg"`) {
		t.Error("non-hidden field 'msg' should appear in output")
	}
}

// T-036: R3.1 — Non-JSON entry displays as plain raw text
func TestRenderRaw_PlainText(t *testing.T) {
	entry := logsource.Entry{
		IsJSON: false,
		Raw:    []byte("2024-01-01 ERROR could not connect to database"),
	}
	result := RenderRaw(entry)
	if result != "2024-01-01 ERROR could not connect to database" {
		t.Errorf("unexpected raw render: %q", result)
	}
}

// T-036: R3.2 — No JSON formatting applied to non-JSON entries
func TestRenderRaw_NoANSI(t *testing.T) {
	entry := logsource.Entry{
		IsJSON: false,
		Raw:    []byte("plain text {not json}"),
	}
	result := RenderRaw(entry)
	// Result must not contain ANSI escape sequences.
	if strings.Contains(result, "\x1b[") {
		t.Error("expected no ANSI escape codes in raw output")
	}
	if result != "plain text {not json}" {
		t.Errorf("unexpected output: %q", result)
	}
}
