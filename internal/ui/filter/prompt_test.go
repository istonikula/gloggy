package filter

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/filter"
)

func defaultPrompt() PromptModel {
	return NewPromptModel(filter.NewFilterSet())
}

// ---------- open / close lifecycle ----------

// T-044 R4.1: pane-click pre-fill sets field and pattern; cursor on Pattern row.
func TestPromptModel_OpenFromPaneClick_PreFilled(t *testing.T) {
	m := defaultPrompt().OpenFromPaneClick("level", "ERROR")
	assert.Equal(t, "level", m.Field())
	assert.Equal(t, "ERROR", m.Pattern())
	assert.True(t, m.IsActive(), "prompt should be active after OpenFromPaneClick")
	assert.Equal(t, filter.Literal, m.Syntax(), "default Syntax = Literal")
	assert.Equal(t, filter.Include, m.Mode(), "default Mode = Include")
}

func TestPromptModel_OpenBlank_AllDefaults(t *testing.T) {
	m := defaultPrompt().OpenBlank()
	assert.True(t, m.IsActive())
	assert.Equal(t, "", m.Field(), "blank Field")
	assert.Equal(t, "", m.Pattern(), "blank Pattern")
	assert.Equal(t, filter.Literal, m.Syntax())
	assert.Equal(t, filter.Include, m.Mode())
}

func TestPromptModel_OpenEdit_PreFilled(t *testing.T) {
	f := filter.Filter{Field: "msg", Pattern: "boom", Syntax: filter.Glob, Mode: filter.Exclude}
	id := 42
	m := defaultPrompt().OpenEdit(f, id)
	assert.True(t, m.IsActive())
	assert.Equal(t, "msg", m.Field())
	assert.Equal(t, "boom", m.Pattern())
	assert.Equal(t, filter.Glob, m.Syntax())
	assert.Equal(t, filter.Exclude, m.Mode())
	require.NotNil(t, m.editID)
	assert.Equal(t, 42, *m.editID)
}

// Esc cancels and emits FilterCancelledMsg.
func TestPromptModel_Esc_Cancels(t *testing.T) {
	m := defaultPrompt().OpenFromPaneClick("level", "ERROR")
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.False(t, m2.IsActive(), "prompt should be closed after Esc")
	require.NotNil(t, cmd, "expected FilterCancelledMsg cmd")
	_, ok := cmd().(FilterCancelledMsg)
	assert.True(t, ok, "expected FilterCancelledMsg")
}

// When closed, Update is a no-op.
func TestPromptModel_Closed_Noop(t *testing.T) {
	m := defaultPrompt()
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.False(t, m2.IsActive(), "should remain closed")
	assert.Nil(t, cmd, "expected nil cmd when closed")
}

// ---------- Tab row cycling ----------

// V35: Tab cycles focus across all 4 rows; Shift+Tab reverses.
func TestPromptModel_Tab_CyclesRows(t *testing.T) {
	m := defaultPrompt().OpenFromPaneClick("level", "INFO") // starts on rowPattern (1)

	// Tab from Pattern → Syntax → Mode → Field → Pattern (wrap)
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, rowSyntax, m2.focusRow, "Tab: Pattern → Syntax")
	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, rowMode, m3.focusRow, "Tab: Syntax → Mode")
	m4, _ := m3.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, rowField, m4.focusRow, "Tab: Mode → Field")
	m5, _ := m4.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, rowPattern, m5.focusRow, "Tab: Field → Pattern (wrap)")
}

func TestPromptModel_ShiftTab_ReversesCycles(t *testing.T) {
	m := defaultPrompt().OpenBlank() // starts on rowPattern (1)

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	assert.Equal(t, rowField, m2.focusRow, "Shift+Tab: Pattern → Field")
}

// V35: > prefix moves with focused row.
func TestPromptModel_View_CursorMovesWithTab(t *testing.T) {
	m := defaultPrompt().OpenBlank() // focusRow = rowPattern
	viewBefore := m.View()

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab}) // → Syntax
	viewAfter := m2.View()

	// Both views contain "> " but at different row positions.
	assert.Contains(t, viewBefore, "> ", "focused row should have > prefix")
	assert.Contains(t, viewAfter, "> ", "focused row should have > prefix after Tab")
	// The views are different (cursor moved).
	assert.NotEqual(t, viewBefore, viewAfter, "View() should change after Tab")
}

// ---------- enum cycling ----------

// V35: ←/→ cycle Syntax on Syntax row.
func TestPromptModel_RightArrow_CyclesSyntax(t *testing.T) {
	m := defaultPrompt().OpenBlank()
	// Tab to Syntax row.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	require.Equal(t, rowSyntax, m.focusRow)

	// Literal → Glob → Regex → Literal (wrap).
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, filter.Glob, m2.Syntax(), "→ once: Literal → Glob")
	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, filter.Regex, m3.Syntax(), "→ twice: Glob → Regex")
	m4, _ := m3.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, filter.Literal, m4.Syntax(), "→ three: Regex → Literal (wrap)")
}

func TestPromptModel_LeftArrow_CyclesSyntaxReverse(t *testing.T) {
	m := defaultPrompt().OpenBlank()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab}) // Syntax row
	require.Equal(t, rowSyntax, m.focusRow)

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	assert.Equal(t, filter.Regex, m2.Syntax(), "← from Literal wraps to Regex")
}

// V35: ←/→ toggle Mode on Mode row.
func TestPromptModel_RightArrow_TogglesMode(t *testing.T) {
	m := defaultPrompt().OpenBlank()
	// Tab to Mode row (rowPattern=1 → rowSyntax=2 → rowMode=3 = 2 tabs).
	for i := 0; i < 2; i++ {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	}
	require.Equal(t, rowMode, m.focusRow)

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, filter.Exclude, m2.Mode(), "→ toggles Include → Exclude")
	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, filter.Include, m3.Mode(), "→ toggles Exclude → Include")
}

// V35: [active] bracket appears on every enum row regardless of focus.
func TestPromptModel_View_BracketOnUnfocusedEnumRow(t *testing.T) {
	m := defaultPrompt().OpenBlank() // focusRow = rowPattern; Syntax=Literal, Mode=Include
	view := m.View()
	// Both enum rows show their active value bracketed.
	assert.Contains(t, view, "[literal]", "Syntax row always shows active value bracketed")
	assert.Contains(t, view, "[include]", "Mode row always shows active value bracketed")
}

// V35: ◂ ▸ cycle hint only on focused enum row.
func TestPromptModel_View_CycleHintOnlyWhenFocused(t *testing.T) {
	m := defaultPrompt().OpenBlank() // focusRow = rowPattern
	view := m.View()
	assert.NotContains(t, view, "◂", "no cycle hint when enum row not focused")

	// Tab to Syntax row.
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	view2 := m2.View()
	assert.Contains(t, view2, "◂", "cycle hint shown when Syntax row focused")
}

// V35: footer adapts per focused row type.
func TestPromptModel_View_FooterAdaptsForEnumRows(t *testing.T) {
	m := defaultPrompt().OpenBlank() // focusRow = rowPattern (text row)
	view := m.View()
	assert.NotContains(t, view, "←/→", "no ←/→ hint when text row focused")

	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab}) // Syntax row
	view2 := m2.View()
	assert.Contains(t, view2, "←/→", "←/→ hint shown when enum row focused")
}

// ---------- text input ----------

// Printable chars appended to focused text row.
func TestPromptModel_Typing_AppendsToFocusedTextRow(t *testing.T) {
	m := defaultPrompt().OpenBlank() // focusRow = rowPattern
	for _, ch := range "INFO" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}
	assert.Equal(t, "INFO", m.Pattern())
	assert.Equal(t, "", m.Field(), "typing on Pattern must not affect Field")
}

func TestPromptModel_Backspace_RemovesLastChar(t *testing.T) {
	m := defaultPrompt().OpenBlank()
	for _, ch := range "INFO" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	assert.Equal(t, "INF", m.Pattern(), "Backspace removes last char")
}

// ---------- Enter — Add path ----------

// T-044 R4.3: Enter confirms new filter → FilterConfirmedMsg{IsNew: true}.
func TestPromptModel_Enter_AddsFilter(t *testing.T) {
	fs := filter.NewFilterSet()
	m := NewPromptModel(fs).OpenFromPaneClick("level", "ERROR")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	msg := cmd()
	confirmed, ok := msg.(FilterConfirmedMsg)
	require.True(t, ok, "expected FilterConfirmedMsg, got %T", msg)
	assert.True(t, confirmed.IsNew, "pane-click path = Add = IsNew true")
	filters := fs.GetAll()
	require.Len(t, filters, 1)
	assert.Equal(t, "level", filters[0].Field)
	assert.Equal(t, "ERROR", filters[0].Pattern)
	assert.True(t, filters[0].Enabled)
}

// ---------- Enter — Edit path ----------

func TestPromptModel_Enter_EditsFilter(t *testing.T) {
	fs := filter.NewFilterSet()
	id := fs.Add(filter.Filter{Field: "msg", Pattern: "old", Syntax: filter.Literal, Mode: filter.Include, Enabled: true})

	f := fs.GetAll()[0]
	m := NewPromptModel(fs).OpenEdit(f, id)
	// Change the pattern.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	for _, ch := range "new" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	msg := cmd()
	confirmed, ok := msg.(FilterConfirmedMsg)
	require.True(t, ok, "expected FilterConfirmedMsg")
	assert.False(t, confirmed.IsNew, "Edit path = IsNew false")
	assert.Equal(t, "new", fs.GetAll()[0].Pattern, "filter pattern updated in-place")
}

// ---------- Enter — validation failure ----------

// V35: empty pattern → FilterRejectedMsg, prompt stays open.
func TestPromptModel_Enter_EmptyPattern_Rejected(t *testing.T) {
	m := defaultPrompt().OpenBlank() // pattern = ""
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	msg := cmd()
	rejected, ok := msg.(FilterRejectedMsg)
	require.True(t, ok, "expected FilterRejectedMsg, got %T", msg)
	assert.Contains(t, rejected.Reason, "pattern required")
	assert.True(t, m2.IsActive(), "prompt must stay open after rejection")
}

// V35: bad regex → FilterRejectedMsg, prompt stays open.
func TestPromptModel_Enter_BadRegex_Rejected(t *testing.T) {
	m := defaultPrompt().OpenBlank()
	// Type pattern, switch to Regex, enter.
	for _, ch := range "[invalid" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})  // → Syntax
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight}) // → Glob
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight}) // → Regex

	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	msg := cmd()
	rejected, ok := msg.(FilterRejectedMsg)
	require.True(t, ok, "expected FilterRejectedMsg for bad regex")
	assert.Contains(t, rejected.Reason, "bad regex")
	assert.True(t, m2.IsActive(), "prompt must stay open after rejection")
}

// ---------- V14: reserved globals become literal chars while prompt active ----------

func TestPromptModel_ReservedGlobals_BecomeQueryChars(t *testing.T) {
	m := defaultPrompt().OpenBlank() // focusRow = Pattern
	for _, key := range []string{"q", "?", "T", "F"} {
		m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
		assert.Nil(t, cmd, "reserved key '%s' must not emit a cmd (no quit/overlay)", key)
		assert.True(t, m2.IsActive(), "reserved key '%s' must not close prompt", key)
	}
	// All 4 chars appended to pattern.
	for _, ch := range "q?TF" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}
	assert.Equal(t, "q?TF", m.Pattern(), "reserved keys append as literal chars")
}

// ---------- V28: no \t in View() ----------

func TestPromptModel_View_NoTabs(t *testing.T) {
	m := defaultPrompt().OpenFromPaneClick("level", "INFO")
	assert.NotContains(t, m.View(), "\t", "View() must not contain tab characters (V28)")
}

// ---------- OpenFromPaneClick regression: same behavior as old Open ----------

func TestPromptModel_OpenFromPaneClick_Regression(t *testing.T) {
	fs := filter.NewFilterSet()
	m := NewPromptModel(fs).OpenFromPaneClick("level", "ERROR")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	msg := cmd()
	confirmed, ok := msg.(FilterConfirmedMsg)
	require.True(t, ok)
	assert.NotNil(t, confirmed.FilterSet)
	filters := fs.GetAll()
	require.Len(t, filters, 1)
	assert.Equal(t, "level", filters[0].Field)
	assert.Equal(t, filter.Include, filters[0].Mode)
	assert.True(t, filters[0].Enabled)
}

// ---------- whole-line filter via OpenBlank (Field stays empty) ----------

func TestPromptModel_OpenBlank_EmptyFieldOnCommit_WholeLine(t *testing.T) {
	fs := filter.NewFilterSet()
	m := NewPromptModel(fs).OpenBlank()
	for _, ch := range "DrawState" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	msg := cmd()
	_, ok := msg.(FilterConfirmedMsg)
	require.True(t, ok)
	f := fs.GetAll()[0]
	assert.Equal(t, "", f.Field, "Field must be empty for whole-line filter")
	assert.Equal(t, "DrawState", f.Pattern)
}

// ---------- title changes for Edit ----------

func TestPromptModel_View_TitleShowsEditForOpenEdit(t *testing.T) {
	f := filter.Filter{Field: "msg", Pattern: "foo", Syntax: filter.Literal, Mode: filter.Include}
	m := defaultPrompt().OpenEdit(f, 1)
	view := m.View()
	assert.True(t, strings.HasPrefix(view, "Edit filter"), "title should be 'Edit filter'")
}

func TestPromptModel_View_TitleShowsAddForOpenBlank(t *testing.T) {
	m := defaultPrompt().OpenBlank()
	view := m.View()
	assert.True(t, strings.HasPrefix(view, "Add filter"), "title should be 'Add filter'")
}
