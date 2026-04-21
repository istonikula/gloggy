package entrylist

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
)

func init() {
	lipgloss.SetColorProfile(termenv.TrueColor)
}

// colorANSI derives the exact ANSI escape prefix that lipgloss produces for a
// given color, by rendering a probe through the same lipgloss path.
func colorANSI(c lipgloss.Color) string {
	rendered := lipgloss.NewStyle().Foreground(c).Render("x")
	end := strings.Index(rendered, "x")
	if end <= 0 {
		return ""
	}
	return rendered[:end]
}

func defaultCfg() config.Config { return config.DefaultConfig() }
func tokyoNight() theme.Theme   { return theme.GetTheme("tokyo-night") }

func jsonEntry(level, logger, msg string, t time.Time) logsource.Entry {
	return logsource.Entry{
		IsJSON: true,
		Time:   t,
		Level:  level,
		Logger: logger,
		Msg:    msg,
		Raw:    []byte(`{}`),
	}
}

// T-022: R1.1 — HH:MM:SS time format
func TestRenderCompactRow_TimeFormatted(t *testing.T) {
	e := jsonEntry("INFO", "app", "hello", time.Date(2024, 3, 15, 14, 30, 59, 0, time.UTC))
	result := RenderCompactRow(e, 80, tokyoNight(), defaultCfg())
	assert.Contains(t, result, "14:30:59", "expected HH:MM:SS")
}

// T-022: R1.2 — level value present
func TestRenderCompactRow_LevelPresent(t *testing.T) {
	e := jsonEntry("WARN", "", "msg", time.Now())
	result := RenderCompactRow(e, 80, tokyoNight(), defaultCfg())
	assert.Contains(t, result, "WARN")
}

// T-022: R1.3 — logger abbreviated to configured depth
func TestRenderCompactRow_LoggerAbbreviated(t *testing.T) {
	cfg := defaultCfg()
	cfg.LoggerDepth = 2
	e := jsonEntry("INFO", "com.example.service.Handler", "msg", time.Now())
	result := RenderCompactRow(e, 120, tokyoNight(), cfg)
	abbr := AbbreviateLogger("com.example.service.Handler", 2)
	assert.Contains(t, result, abbr)
}

// T-022: R1.4 — message truncated to fit width
func TestRenderCompactRow_MsgTruncated(t *testing.T) {
	e := jsonEntry("INFO", "app", "this is a very long message that should be truncated at some point", time.Now())
	result := RenderCompactRow(e, 30, tokyoNight(), defaultCfg())
	assert.NotContains(t, result, "truncated", "message should have been cut before 'truncated'")
}

// T-022: R1.5 — non-JSON shows raw text
func TestRenderCompactRow_NonJSONRawText(t *testing.T) {
	e := logsource.Entry{IsJSON: false, Raw: []byte("some plain text log line")}
	result := RenderCompactRow(e, 80, tokyoNight(), defaultCfg())
	assert.Contains(t, result, "some plain text log line")
}

// T-022: R1.7 — zero time shows placeholder
func TestRenderCompactRow_ZeroTimePlaceholder(t *testing.T) {
	e := jsonEntry("INFO", "app", "no time", time.Time{})
	result := RenderCompactRow(e, 80, tokyoNight(), defaultCfg())
	assert.Contains(t, result, "--:--:--", "expected time placeholder")
}

// T-023: R3.1 — ERROR uses LevelError color
func TestRenderCompactRow_ErrorColor(t *testing.T) {
	th := tokyoNight()
	e := jsonEntry("ERROR", "app", "fail", time.Now())
	result := RenderCompactRow(e, 80, th, defaultCfg())
	want := colorANSI(th.LevelError)
	require.NotEmpty(t, want, "could not derive LevelError ANSI code")
	assert.Contains(t, result, want, "expected LevelError ANSI in output")
}

// T-023: R3.2 — WARN uses LevelWarn color
func TestRenderCompactRow_WarnColor(t *testing.T) {
	th := tokyoNight()
	e := jsonEntry("WARN", "app", "warn", time.Now())
	result := RenderCompactRow(e, 80, th, defaultCfg())
	want := colorANSI(th.LevelWarn)
	assert.Contains(t, result, want, "expected LevelWarn ANSI in output")
}

// T-023: R3.3 — INFO uses LevelInfo color
func TestRenderCompactRow_InfoColor(t *testing.T) {
	th := tokyoNight()
	e := jsonEntry("INFO", "app", "info", time.Now())
	result := RenderCompactRow(e, 80, th, defaultCfg())
	want := colorANSI(th.LevelInfo)
	assert.Contains(t, result, want, "expected LevelInfo ANSI in output")
}

// T-023: R3.4 — DEBUG uses LevelDebug color
func TestRenderCompactRow_DebugColor(t *testing.T) {
	th := tokyoNight()
	e := jsonEntry("DEBUG", "app", "debug", time.Now())
	result := RenderCompactRow(e, 80, th, defaultCfg())
	want := colorANSI(th.LevelDebug)
	assert.Contains(t, result, want, "expected LevelDebug ANSI in output")
}

// T-023: R3.5 — switching theme changes ANSI codes
func TestRenderCompactRow_ThemeSwitch(t *testing.T) {
	e := jsonEntry("ERROR", "app", "fail", time.Now())
	cfg := defaultCfg()

	th1 := theme.GetTheme("tokyo-night")
	th2 := theme.GetTheme("catppuccin-mocha")

	if string(th1.LevelError) == string(th2.LevelError) {
		t.Skip("themes have same error color, skip switch test")
	}

	out1 := RenderCompactRow(e, 80, th1, cfg)
	out2 := RenderCompactRow(e, 80, th2, cfg)

	assert.NotEqual(t, out1, out2, "switching theme should produce different output")
	assert.Contains(t, out2, colorANSI(th2.LevelError),
		"catppuccin-mocha output missing its error color ANSI")
}

// V4: messages with embedded newlines must render as exactly one terminal line.
func TestFlattenNewlines(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"no newlines", "no newlines"},
		{"line1\nline2", "line1 line2"},
		{"line1\r\nline2", "line1 line2"},
		{"tabs\t\there", "tabs\t\there"},
		{"mixed\n\t\tindent", "mixed indent"},
		{"trailing\n", "trailing "},
		{"\nleading", " leading"},
		{"multi\n\n\nnewlines", "multi newlines"},
	}
	for _, tt := range tests {
		got := flattenNewlines(tt.in)
		assert.Equalf(t, tt.want, got, "flattenNewlines(%q)", tt.in)
	}
}

// V4: an entry with embedded newlines in its message renders as one line.
func TestRenderCompactRow_FlattenNewlines(t *testing.T) {
	entry := logsource.Entry{
		IsJSON: true,
		Time:   time.Now(),
		Level:  "INFO",
		Logger: "test",
		Msg:    "line1\n\tline2\n\tline3",
		Raw:    []byte(`{}`),
	}
	row := RenderCompactRow(entry, 120, theme.GetTheme("tokyo-night"), defaultCfg())
	assert.NotContainsf(t, row, "\n", "compact row contains newline: %q", row)
}
