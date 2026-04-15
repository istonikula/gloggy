package entrylist

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

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
	if !strings.Contains(result, "14:30:59") {
		t.Errorf("expected HH:MM:SS, got %q", result)
	}
}

// T-022: R1.2 — level value present
func TestRenderCompactRow_LevelPresent(t *testing.T) {
	e := jsonEntry("WARN", "", "msg", time.Now())
	result := RenderCompactRow(e, 80, tokyoNight(), defaultCfg())
	if !strings.Contains(result, "WARN") {
		t.Errorf("expected WARN in output, got %q", result)
	}
}

// T-022: R1.3 — logger abbreviated to configured depth
func TestRenderCompactRow_LoggerAbbreviated(t *testing.T) {
	cfg := defaultCfg()
	cfg.LoggerDepth = 2
	e := jsonEntry("INFO", "com.example.service.Handler", "msg", time.Now())
	result := RenderCompactRow(e, 120, tokyoNight(), cfg)
	abbr := AbbreviateLogger("com.example.service.Handler", 2)
	if !strings.Contains(result, abbr) {
		t.Errorf("expected %q in output, got %q", abbr, result)
	}
}

// T-022: R1.4 — message truncated to fit width
func TestRenderCompactRow_MsgTruncated(t *testing.T) {
	e := jsonEntry("INFO", "app", "this is a very long message that should be truncated at some point", time.Now())
	result := RenderCompactRow(e, 30, tokyoNight(), defaultCfg())
	if strings.Contains(result, "truncated") {
		t.Errorf("message should have been cut before 'truncated', got %q", result)
	}
}

// T-022: R1.5 — non-JSON shows raw text
func TestRenderCompactRow_NonJSONRawText(t *testing.T) {
	e := logsource.Entry{IsJSON: false, Raw: []byte("some plain text log line")}
	result := RenderCompactRow(e, 80, tokyoNight(), defaultCfg())
	if !strings.Contains(result, "some plain text log line") {
		t.Errorf("expected raw text in output, got %q", result)
	}
}

// T-022: R1.7 — zero time shows placeholder
func TestRenderCompactRow_ZeroTimePlaceholder(t *testing.T) {
	e := jsonEntry("INFO", "app", "no time", time.Time{})
	result := RenderCompactRow(e, 80, tokyoNight(), defaultCfg())
	if !strings.Contains(result, "--:--:--") {
		t.Errorf("expected time placeholder, got %q", result)
	}
}

// T-023: R3.1 — ERROR uses LevelError color
func TestRenderCompactRow_ErrorColor(t *testing.T) {
	th := tokyoNight()
	e := jsonEntry("ERROR", "app", "fail", time.Now())
	result := RenderCompactRow(e, 80, th, defaultCfg())
	want := colorANSI(th.LevelError)
	if want == "" {
		t.Fatal("could not derive LevelError ANSI code")
	}
	if !strings.Contains(result, want) {
		t.Errorf("expected LevelError ANSI in output; got %q", result)
	}
}

// T-023: R3.2 — WARN uses LevelWarn color
func TestRenderCompactRow_WarnColor(t *testing.T) {
	th := tokyoNight()
	e := jsonEntry("WARN", "app", "warn", time.Now())
	result := RenderCompactRow(e, 80, th, defaultCfg())
	want := colorANSI(th.LevelWarn)
	if !strings.Contains(result, want) {
		t.Errorf("expected LevelWarn ANSI in output; got %q", result)
	}
}

// T-023: R3.3 — INFO uses LevelInfo color
func TestRenderCompactRow_InfoColor(t *testing.T) {
	th := tokyoNight()
	e := jsonEntry("INFO", "app", "info", time.Now())
	result := RenderCompactRow(e, 80, th, defaultCfg())
	want := colorANSI(th.LevelInfo)
	if !strings.Contains(result, want) {
		t.Errorf("expected LevelInfo ANSI in output; got %q", result)
	}
}

// T-023: R3.4 — DEBUG uses LevelDebug color
func TestRenderCompactRow_DebugColor(t *testing.T) {
	th := tokyoNight()
	e := jsonEntry("DEBUG", "app", "debug", time.Now())
	result := RenderCompactRow(e, 80, th, defaultCfg())
	want := colorANSI(th.LevelDebug)
	if !strings.Contains(result, want) {
		t.Errorf("expected LevelDebug ANSI in output; got %q", result)
	}
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

	if out1 == out2 {
		t.Error("switching theme should produce different output")
	}
	if !strings.Contains(out2, colorANSI(th2.LevelError)) {
		t.Errorf("catppuccin-mocha output missing its error color ANSI")
	}
}
