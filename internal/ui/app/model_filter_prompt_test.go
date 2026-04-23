package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/filter"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/ui/appshell"
	"github.com/istonikula/gloggy/internal/ui/detailpane"
)

// jsonEntryLvl builds an entry whose Raw bytes parse as a JSON object
// with the given level/msg. Sibling `jsonEntry(line, msg)` defaults the
// level to INFO; T28's tests need variable level so they filter-toggle
// visibility, so this helper exposes the extra parameter.
func jsonEntryLvl(line int, level, msg string) logsource.Entry {
	raw := `{"level":"` + level + `","msg":"` + msg + `"}`
	return logsource.Entry{
		LineNumber: line,
		IsJSON:     true,
		Level:      level,
		Msg:        msg,
		Raw:        []byte(raw),
	}
}

// seedModelWithPaneOpen builds a Model at 90x30 with N JSON entries and
// the detail pane opened on the first entry — the minimum surface for
// T28's click-to-filter integration path. 90 cols is below the default
// orientation_threshold_cols=100 so the pane sits BELOW the list and
// spans the full terminal width; a mouse click anywhere inside the pane
// vertical band lands in ZoneDetailPane regardless of X.
func seedModelWithPaneOpen(t *testing.T, entries []logsource.Entry) Model {
	t.Helper()
	m := newModel()
	m = resize(m, 90, 30)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	require.True(t, m.pane.IsOpen(), "precondition: pane open")
	require.Equalf(t, appshell.OrientationBelow, m.resize.Orientation(),
		"precondition: below-mode so X doesn't matter for pane zoning")
	return m
}

// runCmd exhausts a tea.Cmd chain by feeding each emitted message back
// through Update. Test helpers `send` / `key` discard commands so
// follow-on dispatches (FilterConfirmedMsg from a prompt Enter, etc.)
// never fire. Callers that need the full chain call sendCmd or keyCmd
// which drain exactly one cmd step.
func runCmd(m Model, cmd tea.Cmd) Model {
	if cmd == nil {
		return m
	}
	msg := cmd()
	if msg == nil {
		return m
	}
	return send(m, msg)
}

// sendCmd dispatches a message AND its resulting command's first
// emission back through Update — needed for any path where one Update
// call emits a message consumed by the next.
func sendCmd(m Model, msg tea.Msg) Model {
	updated, cmd := m.Update(msg)
	m = updated.(Model)
	return runCmd(m, cmd)
}

// keyCmd is the cmd-draining sibling of `key`.
func keyCmd(m Model, k string) Model {
	var msg tea.KeyMsg
	switch k {
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		msg = tea.KeyMsg{Type: tea.KeyEsc}
	case "tab":
		msg = tea.KeyMsg{Type: tea.KeyTab}
	default:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	}
	return sendCmd(m, msg)
}

// ---------- V33: FieldClickMsg opens prompt pre-filled ----------

// TestModel_FieldClick_OpensPromptWithPreFill (T28/V33) — sending a
// FieldClickMsg with a (field, value) pair must open the filter prompt
// with that pre-fill and leave the FilterSet untouched until Enter.
func TestModel_FieldClick_OpensPromptWithPreFill(t *testing.T) {
	m := seedModelWithPaneOpen(t, []logsource.Entry{jsonEntryLvl(1, "INFO", "hello")})
	require.False(t, m.filterPrompt.IsActive(), "precondition: prompt inactive")

	m = send(m, detailpane.FieldClickMsg{Field: "level", Value: "INFO"})

	assert.True(t, m.filterPrompt.IsActive(), "prompt should open on FieldClickMsg")
	assert.Equal(t, "level", m.filterPrompt.Field(), "field pre-fill")
	assert.Equal(t, "INFO", m.filterPrompt.Pattern(), "value pre-fill")
	assert.Empty(t, m.filterSet.GetAll(), "FilterSet must stay empty until Enter")
}

// ---------- V33: full integration — mouse click → Enter → FilterSet ----------

// findFieldLineY walks the pane's content lines and returns the absolute
// terminal Y for the first line matching `"<field>":`. The test uses this
// to emit a MouseMsg whose Y lands on a real field row after the pane
// has soft-wrapped its content under the active width.
func findFieldLineY(t *testing.T, m Model, field string) int {
	t.Helper()
	lines := m.pane.ContentLines()
	for i, line := range lines {
		if strings.Contains(line, `"`+field+`":`) {
			startY := m.layout.Layout().DetailPaneContentTopY()
			return startY + i - 0 // scroll offset is 0 for a freshly opened pane
		}
	}
	require.Failf(t, "field line not found", "no line with %q in %v", field, lines)
	return -1
}

// TestModel_FilterPrompt_MouseClick_Enter_AddsFilter_V33 drives the entire
// chain end-to-end: pane open → MouseMsg on a `"level": ...` row in the
// pane → app emits FieldClickMsg → prompt opens pre-filled → user Enter
// → FilterConfirmedMsg → refilter. This is the §V.33 integration coverage
// test — exactly what was missing when the flow shipped with passing
// isolated unit tests but no `case FieldClickMsg` in app.Update (B7).
func TestModel_FilterPrompt_MouseClick_Enter_AddsFilter_V33(t *testing.T) {
	entries := []logsource.Entry{
		jsonEntryLvl(1, "INFO", "hello"),
		jsonEntryLvl(2, "ERROR", "boom"),
		jsonEntryLvl(3, "INFO", "world"),
	}
	m := seedModelWithPaneOpen(t, entries)

	// Full total is 3; pre-click cachedVisibleCount mirrors SetEntries.
	require.Equalf(t, len(entries), m.cachedVisibleCount,
		"precondition: no filters → all entries visible")

	clickY := findFieldLineY(t, m, "level")
	// Mouse-click emits FieldClickMsg via a tea.Cmd — drain it so the
	// FieldClickMsg reaches Update and opens the prompt (V33 integration).
	m = sendCmd(m, tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
		X:      5,
		Y:      clickY,
	})

	require.Truef(t, m.filterPrompt.IsActive(),
		"V33: click on `\"level\":` row must open the filter prompt — B7 recurrence guard")
	assert.Equal(t, "level", m.filterPrompt.Field(), "prompt field pre-fill")
	assert.Equal(t, "INFO", m.filterPrompt.Pattern(), "prompt value pre-fill")

	// User confirms. Enter → prompt emits FilterConfirmedMsg; the existing
	// handler in Update calls m.refilter(). keyCmd drains that cmd so the
	// confirmation reaches app.Update within this test step.
	m = keyCmd(m, "enter")

	assert.Falsef(t, m.filterPrompt.IsActive(),
		"prompt should close on Enter")
	require.Lenf(t, m.filterSet.GetAll(), 1,
		"FilterSet must hold exactly the confirmed filter post-Enter")
	f := m.filterSet.GetAll()[0]
	assert.Equal(t, "level", f.Field, "filter field")
	assert.Equal(t, "INFO", f.Pattern, "filter pattern")
	assert.True(t, f.Enabled, "filter should be enabled on add")
}

// TestModel_FilterPrompt_Enter_RefilterApplied extends the V33 integration
// past the FilterSet assertion to the rendered outcome: the include-filter
// for `level:INFO` hides the ERROR row, so cachedVisibleCount drops from
// 3 to 2.
func TestModel_FilterPrompt_Enter_RefilterApplied(t *testing.T) {
	entries := []logsource.Entry{
		jsonEntryLvl(1, "INFO", "hello"),
		jsonEntryLvl(2, "ERROR", "boom"),
		jsonEntryLvl(3, "INFO", "world"),
	}
	m := seedModelWithPaneOpen(t, entries)

	m = send(m, detailpane.FieldClickMsg{Field: "level", Value: "INFO"})
	m = keyCmd(m, "enter")

	assert.Equalf(t, 2, m.cachedVisibleCount,
		"include-filter level:INFO should hide the 1 ERROR row (V33)")
}

// ---------- V14: Esc cancels with no mutation ----------

// TestModel_FilterPrompt_Esc_CancelsNoMutation (T28/V14) — Esc pressed
// while the prompt is active closes the prompt and leaves the FilterSet
// untouched; no FilterChangedMsg, no refilter.
func TestModel_FilterPrompt_Esc_CancelsNoMutation(t *testing.T) {
	m := seedModelWithPaneOpen(t, []logsource.Entry{jsonEntryLvl(1, "INFO", "hi")})
	m = send(m, detailpane.FieldClickMsg{Field: "level", Value: "INFO"})
	require.True(t, m.filterPrompt.IsActive(), "precondition: prompt active")

	m = key(m, "esc")

	assert.False(t, m.filterPrompt.IsActive(), "prompt closed on Esc")
	assert.Empty(t, m.filterSet.GetAll(), "Esc must leave FilterSet empty")
}

// ---------- V14: reserved globals are literal while prompt active ----------

// TestModel_FilterPrompt_Active_ReservedKeysAreLiteral (T28/V14) — while
// the prompt is active, reserved globals (`q`, `?`, `F`, `/`) are
// consumed by the prompt-intercept and MUST NOT trigger their usual
// actions (quit, help, filter-toggle, list-search). The prompt stays
// open. Same V14 invariant that gates help/themesel vs. pane-search.
func TestModel_FilterPrompt_Active_ReservedKeysAreLiteral(t *testing.T) {
	m := seedModelWithPaneOpen(t, []logsource.Entry{jsonEntryLvl(1, "INFO", "hi")})
	m.filterSet.Add(filter.Filter{Field: "level", Pattern: "INFO", Mode: filter.Include, Enabled: true})
	m = send(m, detailpane.FieldClickMsg{Field: "level", Value: "INFO"})
	require.True(t, m.filterPrompt.IsActive(), "precondition: prompt active")
	require.False(t, m.help.IsOpen(), "precondition: help closed")
	require.False(t, m.filterSet.IsGloballyDisabled(), "precondition: filters enabled")
	require.False(t, m.list.HasActiveSearch(), "precondition: no list search")

	for _, k := range []string{"q", "?", "F", "/"} {
		m = key(m, k)
		require.Truef(t, m.filterPrompt.IsActive(),
			"V14: key %q must NOT dismiss the filter prompt", k)
	}

	assert.False(t, m.help.IsOpen(), "? must not open help while prompt active")
	assert.Falsef(t, m.filterSet.IsGloballyDisabled(),
		"F must not toggle filters globally while prompt active")
	assert.False(t, m.list.HasActiveSearch(),
		"/ must not activate list search while prompt active")
}

// ---------- V14: pane-search input mode suppresses FieldClick emission ----------

// TestModel_FieldClick_GatedByPaneSearchInput (T28/V14) — while the
// detail pane's search is in input mode, a left-click on a pane field
// MUST NOT emit FieldClickMsg. Opening a prompt on top of an active
// query would preempt the user's typed characters — same V14 class as
// `?`/`T`/`t`/`F` being consumed as query chars during input mode.
func TestModel_FieldClick_GatedByPaneSearchInput(t *testing.T) {
	m := seedModelWithPaneOpen(t, []logsource.Entry{jsonEntryLvl(1, "INFO", "hi")})
	m = setFocus(m, appshell.FocusDetailPane)
	// Activate pane search into input mode with a `/` key.
	m = key(m, "/")
	require.Truef(t, m.paneSearch.IsActive(),
		"precondition: pane search active")
	require.Equalf(t, detailpane.SearchModeInput, m.paneSearch.Mode(),
		"precondition: pane search in input mode")

	clickY := findFieldLineY(t, m, "level")
	m = send(m, tea.MouseMsg{
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
		X:      5,
		Y:      clickY,
	})

	assert.Falsef(t, m.filterPrompt.IsActive(),
		"V14: FieldClick must be suppressed while pane search is in input mode")
}

// ---------- View: prompt replaces keyhints row ----------

// TestModel_FilterPrompt_View_ReplacesStatusRow (T28/V15) — while active
// the prompt's single-line View replaces the keyhints bar at the bottom
// so the user sees the pre-filled field/value + Tab/Enter/Esc hints.
func TestModel_FilterPrompt_View_ReplacesStatusRow(t *testing.T) {
	m := seedModelWithPaneOpen(t, []logsource.Entry{jsonEntryLvl(1, "INFO", "hi")})
	m = send(m, detailpane.FieldClickMsg{Field: "level", Value: "INFO"})
	require.True(t, m.filterPrompt.IsActive(), "precondition: prompt active")

	out := m.View()

	assert.Containsf(t, out, "Add filter: level:INFO",
		"View must render the prompt while active")
	assert.Containsf(t, out, "Enter=confirm",
		"View must surface the prompt hints")
}
