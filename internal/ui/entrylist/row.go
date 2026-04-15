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
		raw := string(entry.Raw)
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

	msg := entry.Msg
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

// padOrTrunc returns s padded with spaces or truncated to exactly n bytes.
func padOrTrunc(s string, n int) string {
	if len(s) >= n {
		return s[:n]
	}
	return s + strings.Repeat(" ", n-len(s))
}
