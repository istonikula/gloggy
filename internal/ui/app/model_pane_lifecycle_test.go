package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/ui/appshell"
)

// ---------- T-091: detail pane auto-close on minimum underflow ----------

// TestModel_AutoClose_RightSplit_BelowMinWidth verifies the detail pane is
// auto-closed when a resize shrinks its content below MinDetailWidth while
// still in right-split orientation. Threshold default = 100 cols, so width
// 100 stays in right-split; detailW = 29 < 30 → auto-close.
func TestModel_AutoClose_RightSplit_BelowMinWidth(t *testing.T) {
	m := newModel()
	m = resize(m, 200, 24)
	entries := makeEntries(5)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	require.True(t, m.pane.IsOpen(), "precondition: pane should be open after openPane")
	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"precondition: orientation should be right at termWidth=200, got %v", m.resize.Orientation())

	// Shrink while keeping right-split: detailW drops below MinDetailWidth.
	m = resize(m, 100, 24)
	require.Equalf(t, appshell.OrientationRight, m.resize.Orientation(),
		"precondition: orientation should stay right at termWidth=100, got %v", m.resize.Orientation())

	assert.Falsef(t, m.pane.IsOpen(),
		"expected auto-close at termWidth=100 (detailW=%d), pane still open", m.layout.Layout().DetailContentWidth())
	assert.True(t, m.keyhints.HasNotice(), "expected status notice after auto-close")
	assert.Equalf(t, appshell.FocusEntryList, m.focus,
		"focus after auto-close: got %v, want FocusEntryList", m.focus)
}

// TestModel_AutoClose_NoticeClearedOnMsg verifies the noticeClearMsg resets
// the key-hint bar back to hints-mode.
func TestModel_AutoClose_NoticeClearedOnMsg(t *testing.T) {
	m := newModel()
	m = resize(m, 200, 24)
	entries := makeEntries(5)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])
	m = resize(m, 100, 24)
	require.True(t, m.keyhints.HasNotice(), "precondition: notice should be set after auto-close")

	m = send(m, noticeClearMsg{})
	assert.False(t, m.keyhints.HasNotice(), "notice should be cleared after noticeClearMsg")
}

// TestModel_AutoClose_AboveMin_KeepsPaneOpen verifies a resize that stays
// above the minimum threshold does not close the pane.
func TestModel_AutoClose_AboveMin_KeepsPaneOpen(t *testing.T) {
	m := newModel()
	m = resize(m, 200, 24)
	entries := makeEntries(5)
	m = m.SetEntries(entries)
	m = m.openPane(entries[0])

	m = resize(m, 180, 24)

	assert.Truef(t, m.pane.IsOpen(),
		"pane should remain open at termWidth=180 (detailW=%d)", m.layout.Layout().DetailContentWidth())
	assert.False(t, m.keyhints.HasNotice(), "notice should not be set when pane stays open")
}

// ---------- R14 (cavekit-entry-list): pane re-sync on tail-follow snap ----------

// jsonEntry builds a minimal JSON entry whose Raw unmarshals cleanly so
// the detail pane's RenderJSON path (not the raw fallback) runs. The
// `msg` field is echoed into the rendered JSON so tests can assert the
// pane content changed to match a new entry.
func jsonEntry(line int, msg string) logsource.Entry {
	return logsource.Entry{
		LineNumber: line,
		IsJSON:     true,
		Level:      "INFO",
		Msg:        msg,
		Raw:        []byte(`{"level":"INFO","msg":"` + msg + `"}`),
	}
}

// cavekit-entry-list R14 (AC: selection signal on tail-follow snap):
// when a tail append snaps the cursor to the new last entry AND the
// detail pane is open, the pane must re-render with that new entry in
// the same frame — no keypress required. Before this AC was added, the
// cursor advanced but the pane kept showing the previous entry.
func TestModel_TailFollow_TailMsg_PaneResyncsOnAppend(t *testing.T) {
	m := newModel()
	m = resize(m, 120, 30)

	initial := []logsource.Entry{
		jsonEntry(1, "alpha-msg"),
		jsonEntry(2, "beta-msg"),
		jsonEntry(3, "gamma-msg"),
	}
	m = m.SetEntries(initial)

	m = key(m, "G")
	require.Equalf(t, 2, m.list.Cursor(), "precondition: cursor at last entry, got %d", m.list.Cursor())
	m = key(m, "enter")
	require.True(t, m.pane.IsOpen(), "precondition: pane open after Enter")
	require.Truef(t, containsSubstring(m.pane.View(), "gamma-msg"),
		"precondition: pane shows gamma-msg; got: %q", m.pane.View())

	appended := jsonEntry(4, "delta-unique")
	m = send(m, logsource.NewTailStreamMsgForTest(logsource.TailMsg{Entries: []logsource.Entry{appended}}))

	assert.Equalf(t, 3, m.list.Cursor(),
		"cursor should snap to new last entry (3), got %d", m.list.Cursor())
	assert.True(t, m.pane.IsOpen(), "pane should still be open after append")
	post := m.pane.View()
	assert.Truef(t, containsSubstring(post, "delta-unique"),
		"pane must re-render with appended entry (delta-unique); got: %q", post)
	assert.Falsef(t, containsSubstring(post, "gamma-msg"),
		"pane must not keep rendering previous entry (gamma-msg); got: %q", post)
}

// cavekit-entry-list R14 symmetry: same re-sync invariant for
// background-load batches (EntryBatchMsg). Tests the same code path as
// TailMsg but via the LoadFileStreamMsg wrapper used during non-follow
// file load.
func TestModel_TailFollow_BatchMsg_PaneResyncsOnAppend(t *testing.T) {
	m := newModel()
	m = resize(m, 120, 30)

	initial := []logsource.Entry{
		jsonEntry(1, "alpha-msg"),
		jsonEntry(2, "beta-msg"),
	}
	m = m.SetEntries(initial)

	m = key(m, "G")
	m = key(m, "enter")
	require.Truef(t, m.pane.IsOpen() && m.list.Cursor() == 1,
		"precondition: cursor=%d pane=%v", m.list.Cursor(), m.pane.IsOpen())

	batch := logsource.EntryBatchMsg{Entries: []logsource.Entry{
		jsonEntry(3, "charlie-msg"),
		jsonEntry(4, "delta-unique"),
	}}
	m = send(m, logsource.NewLoadFileStreamMsgForTest(batch))

	assert.Equalf(t, 3, m.list.Cursor(),
		"cursor should snap to new last entry (3), got %d", m.list.Cursor())
	post := m.pane.View()
	assert.Truef(t, containsSubstring(post, "delta-unique"),
		"pane must re-render with last appended entry (delta-unique); got: %q", post)
	assert.Falsef(t, containsSubstring(post, "beta-msg"),
		"pane must not keep rendering pre-append entry (beta-msg); got: %q", post)
}

// cavekit-entry-list R14: signal must fire only when cursor actually
// moved. With the cursor NOT at the last entry (follow disengaged), an
// append must NOT re-render the pane — otherwise every tail event
// silently clobbers the user's pane selection while they are reading a
// historical entry.
func TestModel_TailFollow_NotAtTail_PaneNotResynced(t *testing.T) {
	m := newModel()
	m = resize(m, 120, 30)

	initial := []logsource.Entry{
		jsonEntry(1, "alpha-msg"),
		jsonEntry(2, "beta-msg"),
		jsonEntry(3, "gamma-msg"),
	}
	m = m.SetEntries(initial)

	// Cursor starts at 0 — NOT at tail. Open pane on alpha.
	m = key(m, "enter")
	require.Truef(t, m.pane.IsOpen() && containsSubstring(m.pane.View(), "alpha-msg"),
		"precondition: pane open on alpha; got: %q", m.pane.View())

	appended := jsonEntry(4, "delta-unique")
	m = send(m, logsource.NewTailStreamMsgForTest(logsource.TailMsg{Entries: []logsource.Entry{appended}}))

	assert.Equalf(t, 0, m.list.Cursor(),
		"cursor must NOT move when pre-append cursor < tail, got %d", m.list.Cursor())
	post := m.pane.View()
	assert.Truef(t, containsSubstring(post, "alpha-msg"),
		"pane must still show alpha-msg (user's non-tail selection preserved); got: %q", post)
	assert.Falsef(t, containsSubstring(post, "delta-unique"),
		"pane must NOT render appended entry (delta-unique) when follow disengaged; got: %q", post)
}

// ---------- T-127 (F-020): hidden-fields wiring ----------

// T-127 (F-020): openPane wires `visibility.HiddenFields()` into the pane
// so config-driven field hiding reaches the JSON renderer.
func TestModel_OpenPane_WiresHiddenFieldsFromVisibility(t *testing.T) {
	cfg := config.LoadResult{Config: config.DefaultConfig()}
	cfg.Config.HiddenFields = []string{"secret"}
	m := New("", false, "", cfg)
	m = resize(m, 80, 24)
	entry := logsource.Entry{
		LineNumber: 1,
		IsJSON:     true,
		Level:      "INFO",
		Msg:        "hi",
		Raw:        []byte(`{"level":"INFO","msg":"hi","secret":"shh"}`),
	}
	m = m.SetEntries([]logsource.Entry{entry})
	m = m.openPane(entry)

	view := m.pane.View()
	assert.Falsef(t, containsSubstring(view, "secret"),
		"detail pane view should not include suppressed field `secret`, got: %q", view)
	assert.Falsef(t, containsSubstring(view, "shh"),
		"detail pane view should not include suppressed value, got: %q", view)
}

// containsSubstring is a small helper to keep imports tight.
func containsSubstring(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
