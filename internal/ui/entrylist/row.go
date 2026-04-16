package entrylist

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
)

const (
	timePlaceholder = "--:--:--"
	timeFormat      = "15:04:05"
	timeWidth       = 8 // HH:MM:SS
	levelWidth      = 5
)

// levelColor returns the theme color token for a log level string.
func levelColor(level string, th theme.Theme) lipgloss.Color {
	switch strings.ToUpper(level) {
	case "ERROR":
		return th.LevelError
	case "WARN", "WARNING":
		return th.LevelWarn
	case "INFO":
		return th.LevelInfo
	case "DEBUG", "TRACE":
		return th.LevelDebug
	default:
		return th.Dim
	}
}

// RenderCompactRow renders a single log entry as a compact row.
//
// JSONL entries: "HH:MM:SS LEVEL LOGGER MSG"
// Non-JSON entries: raw text in dim color.
//
// Output is truncated so visible characters fit within width.
func RenderCompactRow(entry logsource.Entry, width int, th theme.Theme, cfg config.Config) string {
	if !entry.IsJSON {
		style := lipgloss.NewStyle().Foreground(th.Dim)
		raw := flattenNewlines(string(entry.Raw))
		if width > 0 && len(raw) > width {
			raw = raw[:width]
		}
		return style.Render(raw)
	}

	// Time column
	timeStr := timePlaceholder
	if !entry.Time.IsZero() {
		timeStr = entry.Time.Format(timeFormat)
	}

	// Level column (padded/truncated to levelWidth, colored)
	lvl := padOrTrunc(strings.ToUpper(entry.Level), levelWidth)
	lvlStyle := lipgloss.NewStyle().Foreground(levelColor(entry.Level, th))
	levelStr := lvlStyle.Render(lvl)

	// Logger column
	loggerStr := AbbreviateLogger(entry.Logger, cfg.LoggerDepth)

	// Compute remaining width for message.
	// Visible prefix: timeWidth + 1(space) + levelWidth + 1(space) + len(loggerStr) + 1(space)
	visiblePrefixLen := timeWidth + 1 + levelWidth + 1 + len(loggerStr) + 1

	msg := flattenNewlines(entry.Msg)
	if width > 0 {
		remaining := width - visiblePrefixLen
		if remaining < 0 {
			remaining = 0
		}
		if len(msg) > remaining {
			msg = msg[:remaining]
		}
	}

	var b strings.Builder
	b.WriteString(timeStr)
	b.WriteByte(' ')
	b.WriteString(levelStr)
	b.WriteByte(' ')
	b.WriteString(loggerStr)
	b.WriteByte(' ')
	b.WriteString(msg)
	return b.String()
}

// RenderCompactRowWithBg renders a compact row with a background color applied
// to all segments, used for cursor highlighting.
func RenderCompactRowWithBg(entry logsource.Entry, width int, th theme.Theme, cfg config.Config, bg lipgloss.Color) string {
	bgStyle := lipgloss.NewStyle().Background(bg)

	if !entry.IsJSON {
		style := lipgloss.NewStyle().Foreground(th.Dim).Background(bg).Width(width)
		raw := flattenNewlines(string(entry.Raw))
		if width > 0 && len(raw) > width {
			raw = raw[:width]
		}
		return style.Render(raw)
	}

	// Time column
	timeStr := timePlaceholder
	if !entry.Time.IsZero() {
		timeStr = entry.Time.Format(timeFormat)
	}
	timeStyled := bgStyle.Render(timeStr)

	// Level column (padded/truncated to levelWidth, colored with bg)
	lvl := padOrTrunc(strings.ToUpper(entry.Level), levelWidth)
	lvlStyle := lipgloss.NewStyle().Foreground(levelColor(entry.Level, th)).Background(bg)
	levelStr := lvlStyle.Render(lvl)

	// Logger column
	loggerStr := bgStyle.Render(AbbreviateLogger(entry.Logger, cfg.LoggerDepth))

	// Compute remaining width for message.
	visiblePrefixLen := timeWidth + 1 + levelWidth + 1 + len(AbbreviateLogger(entry.Logger, cfg.LoggerDepth)) + 1

	msg := flattenNewlines(entry.Msg)
	if width > 0 {
		remaining := width - visiblePrefixLen
		if remaining < 0 {
			remaining = 0
		}
		if len(msg) > remaining {
			msg = msg[:remaining]
		}
	}

	// Pad to fill full width.
	totalVisible := visiblePrefixLen + len(msg)
	padding := ""
	if width > totalVisible {
		padding = strings.Repeat(" ", width-totalVisible)
	}

	msgStyled := bgStyle.Render(msg + padding)

	sp := bgStyle.Render(" ")

	var b strings.Builder
	b.WriteString(timeStyled)
	b.WriteString(sp)
	b.WriteString(levelStr)
	b.WriteString(sp)
	b.WriteString(loggerStr)
	b.WriteString(sp)
	b.WriteString(msgStyled)
	return b.String()
}

// flattenNewlines replaces newlines (and surrounding whitespace) with a single
// space so that each compact row is exactly one terminal line.
func flattenNewlines(s string) string {
	if !strings.ContainsAny(s, "\n\r") {
		return s
	}
	// Replace \r\n, \n, \r (possibly with surrounding tabs/spaces) with a single space.
	var b strings.Builder
	b.Grow(len(s))
	inWs := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\n' || c == '\r' || c == '\t' {
			if !inWs {
				b.WriteByte(' ')
				inWs = true
			}
			continue
		}
		if c == ' ' && inWs {
			continue
		}
		inWs = false
		b.WriteByte(c)
	}
	return b.String()
}

// padOrTrunc returns s padded with spaces or truncated to exactly n bytes.
func padOrTrunc(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + strings.Repeat(" ", n-len(s))
}
