package entrylist

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/istonikula/gloggy/internal/logsource"
)

// CursorLevel indicates which navigation level is active.
type CursorLevel int

const (
	LevelEvent  CursorLevel = iota // navigating between entries
	LevelSubRow                    // navigating within sub-rows of an entry
)

// CursorModel handles two-level cursor navigation within the entry list.
// At LevelEvent: j/k move between entries.
// At LevelSubRow: up/down move between sub-rows; h/←/Esc returns to LevelEvent.
type CursorModel struct {
	Level      CursorLevel
	EntryIndex int // index into the full entry list
	SubIndex   int // index within current entry's sub-rows
	subFields  []string
}

// NewCursorModel creates a cursor model that shows the given fields as sub-rows.
func NewCursorModel(subFields []string) CursorModel {
	return CursorModel{subFields: subFields}
}

// SubRows returns the displayable sub-rows for an entry.
// Each sub-row is (fieldName, valueString).
func SubRows(entry logsource.Entry, fields []string) [][2]string {
	var rows [][2]string
	if !entry.IsJSON {
		return rows
	}
	for _, field := range fields {
		var val string
		switch field {
		case "time":
			if !entry.Time.IsZero() {
				val = entry.Time.Format("2006-01-02T15:04:05Z07:00")
			}
		case "level":
			val = entry.Level
		case "msg":
			val = entry.Msg
		case "logger":
			val = entry.Logger
		case "thread":
			val = entry.Thread
		default:
			if raw, ok := entry.Extra[field]; ok {
				var v interface{}
				if err := json.Unmarshal(raw, &v); err == nil {
					val = fmt.Sprintf("%v", v)
				} else {
					val = string(raw)
				}
			}
		}
		if val != "" {
			rows = append(rows, [2]string{field, val})
		}
	}
	return rows
}

// HasSubRows reports whether the entry has any sub-rows to display.
func (m CursorModel) HasSubRows(entry logsource.Entry) bool {
	return len(SubRows(entry, m.subFields)) > 0
}

// EnterSubLevel enters sub-row navigation for the current entry.
func (m CursorModel) EnterSubLevel() CursorModel {
	m.Level = LevelSubRow
	m.SubIndex = 0
	return m
}

// ExitSubLevel returns to event-level navigation.
func (m CursorModel) ExitSubLevel() CursorModel {
	m.Level = LevelEvent
	m.SubIndex = 0
	return m
}

// Update handles navigation keys.
// entryCount is the total number of entries. currentEntry is the entry at EntryIndex.
func (m CursorModel) Update(msg tea.Msg, entryCount int, currentEntry logsource.Entry) (CursorModel, tea.Cmd) {
	if entryCount == 0 {
		return m, nil
	}

	subRows := SubRows(currentEntry, m.subFields)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.Level == LevelEvent {
			switch msg.String() {
			case "j", "down":
				if m.EntryIndex < entryCount-1 {
					m.EntryIndex++
				}
			case "k", "up":
				if m.EntryIndex > 0 {
					m.EntryIndex--
				}
			case "l", "right", "tab":
				if len(subRows) > 0 {
					m = m.EnterSubLevel()
				}
			}
		} else { // LevelSubRow
			switch msg.String() {
			case "j", "down":
				if m.SubIndex < len(subRows)-1 {
					m.SubIndex++
				}
			case "k", "up":
				if m.SubIndex > 0 {
					m.SubIndex--
				}
			case "h", "left", "esc":
				m = m.ExitSubLevel()
			}
		}
	}
	return m, nil
}

// RenderEntry renders a single entry as compact row + optional sub-rows.
// isSelected indicates whether this is the cursor entry.
// Returns the rendered string with appropriate indentation and cursor markers.
func (m CursorModel) RenderEntry(entry logsource.Entry, entryIdx int, rowStr string) string {
	var sb strings.Builder

	// Entry boundary marker for entries with sub-row fields.
	hasSubRows := m.HasSubRows(entry)
	isCurrent := entryIdx == m.EntryIndex

	cursor := "  "
	if isCurrent && m.Level == LevelEvent {
		cursor = "> "
	}
	sb.WriteString(cursor)
	sb.WriteString(rowStr)

	if hasSubRows && isCurrent && m.Level == LevelSubRow {
		rows := SubRows(entry, m.subFields)
		for i, row := range rows {
			sb.WriteByte('\n')
			subCursor := "    "
			if i == m.SubIndex {
				subCursor = "  > "
			}
			sb.WriteString(subCursor)
			sb.WriteString(row[0])
			sb.WriteString(": ")
			sb.WriteString(row[1])
		}
	} else if hasSubRows {
		// Visual boundary indicator: show a separator line.
		sb.WriteString(" ...")
	}

	return sb.String()
}
