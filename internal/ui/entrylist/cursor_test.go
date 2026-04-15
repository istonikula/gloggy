package entrylist

import (
	"encoding/json"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/istonikula/gloggy/internal/logsource"
)

func jsonCursorEntry(level, logger, msg string) logsource.Entry {
	return logsource.Entry{
		IsJSON: true,
		Time:   time.Now(),
		Level:  level,
		Logger: logger,
		Msg:    msg,
		Extra: map[string]json.RawMessage{
			"thread": json.RawMessage(`"worker-1"`),
		},
		Raw: []byte(`{}`),
	}
}

func newCursor() CursorModel {
	return NewCursorModel([]string{"level", "logger", "thread"})
}

// T-030: R4.1 — j moves cursor to next entry
func TestCursor_JMovesDown(t *testing.T) {
	m := newCursor()
	entries := []logsource.Entry{jsonCursorEntry("INFO", "a", "1"), jsonCursorEntry("WARN", "b", "2")}
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}, len(entries), entries[0])
	if m2.EntryIndex != 1 {
		t.Errorf("j: EntryIndex = %d, want 1", m2.EntryIndex)
	}
}

// T-030: R4.2 — k moves cursor to previous entry
func TestCursor_KMovesUp(t *testing.T) {
	m := newCursor()
	m.EntryIndex = 1
	entries := []logsource.Entry{jsonCursorEntry("INFO", "a", "1"), jsonCursorEntry("WARN", "b", "2")}
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}, len(entries), entries[1])
	if m2.EntryIndex != 0 {
		t.Errorf("k: EntryIndex = %d, want 0", m2.EntryIndex)
	}
}

// T-030: R4.3 — j/k never land on sub-rows (sub-rows are shown via expand, not as cursor positions)
func TestCursor_JKStayAtEventLevel(t *testing.T) {
	m := newCursor()
	entries := []logsource.Entry{jsonCursorEntry("INFO", "a", "1"), jsonCursorEntry("WARN", "b", "2")}
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}, len(entries), entries[0])
	if m2.Level != LevelEvent {
		t.Error("j should stay at LevelEvent, not enter sub-row level")
	}
}

// T-030: R4.4 — l/Tab enters sub-row level
func TestCursor_LEntersSubRowLevel(t *testing.T) {
	m := newCursor()
	entry := jsonCursorEntry("INFO", "app", "hello")

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}, 1, entry)
	if m2.Level != LevelSubRow {
		t.Errorf("l: Level = %v, want LevelSubRow", m2.Level)
	}

	m3, _ := m.Update(tea.KeyMsg{Type: tea.KeyType(tea.KeyTab)}, 1, entry)
	if m3.Level != LevelSubRow {
		t.Errorf("Tab: Level = %v, want LevelSubRow", m3.Level)
	}
}

// T-030: R4.5 — sub-rows appear indented in render
func TestCursor_SubRowsRenderedIndented(t *testing.T) {
	m := newCursor()
	m = m.EnterSubLevel()
	entry := jsonCursorEntry("INFO", "app", "hello")
	rendered := m.RenderEntry(entry, 0, "row-text")
	if !containsSubRowIndent(rendered) {
		t.Errorf("expected indented sub-rows in render, got:\n%s", rendered)
	}
}

func containsSubRowIndent(s string) bool {
	lines := splitLines(s)
	for _, l := range lines[1:] {
		if len(l) > 2 && (l[:2] == "  " || l[:4] == "    ") {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	var lines []string
	cur := ""
	for _, c := range s {
		if c == '\n' {
			lines = append(lines, cur)
			cur = ""
		} else {
			cur += string(c)
		}
	}
	lines = append(lines, cur)
	return lines
}

// T-030: R4.6 — each sub-row shows field name and value
func TestCursor_SubRowsShowFieldAndValue(t *testing.T) {
	m := newCursor()
	m = m.EnterSubLevel()
	entry := jsonCursorEntry("INFO", "com.example.App", "test")
	rendered := m.RenderEntry(entry, 0, "row-text")
	if !contains(rendered, "logger") {
		t.Errorf("expected 'logger' in sub-rows, got:\n%s", rendered)
	}
	if !contains(rendered, "com.example.App") {
		t.Errorf("expected logger value in sub-rows, got:\n%s", rendered)
	}
}

// T-030: R4.7 — h/←/Esc exits sub-row level
func TestCursor_HExitsSubRowLevel(t *testing.T) {
	m := newCursor().EnterSubLevel()
	entry := jsonCursorEntry("INFO", "app", "msg")

	for _, key := range []string{"h", "esc"} {
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}, 1, entry)
		if m2.Level != LevelEvent {
			t.Errorf("%s: Level = %v, want LevelEvent", key, m2.Level)
		}
	}

	m3, _ := m.Update(tea.KeyMsg{Type: tea.KeyLeft}, 1, entry)
	if m3.Level != LevelEvent {
		t.Errorf("left arrow: Level = %v, want LevelEvent", m3.Level)
	}
}

// T-030: R4.8 — entries with sub-row fields show visual boundary at event level
func TestCursor_VisualBoundaryAtEventLevel(t *testing.T) {
	m := newCursor() // at LevelEvent
	entry := jsonCursorEntry("INFO", "app", "hello") // has logger + thread sub-row fields
	rendered := m.RenderEntry(entry, 0, "row-text")
	// The "..." suffix is our visual boundary marker.
	if !contains(rendered, "...") {
		t.Errorf("expected visual boundary indicator for entry with sub-row fields, got:\n%s", rendered)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
