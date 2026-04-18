package entrylist

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
	"github.com/istonikula/gloggy/internal/ui/appshell"
)

// SelectionMsg is emitted when the cursor moves to a new entry.
type SelectionMsg struct {
	Entry logsource.Entry
}

// WrapIndicator is emitted when a level-jump or mark navigation wraps around.
type WrapIndicator struct {
	Direction WrapDirection
}

// OpenDetailPaneMsg is emitted when the user double-clicks an entry to open the detail pane.
type OpenDetailPaneMsg struct {
	Entry logsource.Entry
}

// ListModel is the virtual-rendering entry list Bubble Tea model.
// It only renders visible rows plus a small buffer, regardless of total entry count.
type ListModel struct {
	entries     []logsource.Entry
	filtered    []int // filtered indices; nil means show all entries
	scroll      ScrollState
	th          theme.Theme
	cfg         config.Config
	width       int
	marks         *MarkSet
	wrapDir       WrapDirection // last wrap direction, reset on next navigation
	pinnedFullIdx int           // -1 = none; full-entries index of a level-jump match that is filtered out, pinned into the visible list with a "outside filter" indicator
	lastClickRow  int
	lastClickTime time.Time
	// search holds list-scope free-text search state (T-143, cavekit-
	// entry-list R13). Activated by app via ActivateSearch() when `/` is
	// pressed with the list focused.
	search SearchModel
	// Focused is set by the app shell before View() is called (T-100).
	// When true, the pane uses PaneStateFocused styling; otherwise
	// PaneStateUnfocused, unless Alone is set.
	Focused bool
	// Alone signals that this pane is the only visible pane (T-101). When
	// true, focused styling is applied regardless of Focused.
	Alone bool
}

// NewListModel creates a ListModel.
func NewListModel(th theme.Theme, cfg config.Config, width, height int) ListModel {
	return ListModel{
		th:            th,
		cfg:           cfg,
		width:         width,
		marks:         NewMarkSet(),
		scroll:        ScrollState{ViewportHeight: height, Scrolloff: cfg.Scrolloff},
		pinnedFullIdx: -1,
		search:        NewSearchModel(th),
	}
}

// WithScrolloff sets the cursor-margin rows applied after every cursor move
// and mouse-wheel tick (T-135, F-026). Wired from `cfg.Scrolloff` by the
// app at WindowSizeMsg and on config reload. Shared across list + detail
// pane (one top-level key — cavekit-config R5).
func (m ListModel) WithScrolloff(n int) ListModel {
	if n < 0 {
		n = 0
	}
	m.scroll.Scrolloff = n
	return m
}

// SetEntries replaces the entry list and resets scroll state.
func (m ListModel) SetEntries(entries []logsource.Entry) ListModel {
	m.entries = entries
	m.scroll.TotalEntries = len(entries)
	m.scroll.Cursor = 0
	m.scroll.Offset = 0
	return m
}

// AppendEntries adds entries (used during background loading). T-148
// (cavekit-entry-list R13 streaming AC): when a list search is active with
// a non-empty query and no filter is applied, the newly appended entries
// are scanned against the query and their visible-set indices appended to
// the search match set so streaming arrivals are matched in real time.
// When a filter is active, appended entries are not visible until the
// filter index is rebuilt — SetFilter deactivates search on its own path,
// so no extension is needed here.
func (m ListModel) AppendEntries(entries []logsource.Entry) ListModel {
	oldLen := len(m.entries)
	m.entries = append(m.entries, entries...)
	m.scroll.TotalEntries = len(m.entries)
	if m.filtered == nil && m.search.IsActive() && m.search.Query() != "" {
		m.search = m.search.ExtendMatches(entries, oldLen, m.width, m.cfg)
	}
	return m
}

// Cursor returns the current cursor index into entries.
func (m ListModel) Cursor() int { return m.scroll.Cursor }

// CursorPosition returns 1-based cursor position within visible set, or 0 when empty.
func (m ListModel) CursorPosition() int {
	vis := m.visibleEntries()
	if len(vis) == 0 {
		return 0
	}
	return m.scroll.Cursor + 1
}

// SelectedEntry returns the entry at the cursor, or zero value if empty.
func (m ListModel) SelectedEntry() (logsource.Entry, bool) {
	if len(m.entries) == 0 || m.scroll.Cursor >= len(m.entries) {
		return logsource.Entry{}, false
	}
	return m.entries[m.scroll.Cursor], true
}

// Search returns the current list search model (T-143, cavekit-entry-list
// R13). Callers (app, View) read query/matches/current state via this
// accessor.
func (m ListModel) Search() SearchModel { return m.search }

// ActivateSearch opens the list search input (T-143). The app calls this
// when `/` is pressed with the list focused and detail pane not focused
// (cavekit-app-shell R13 focus-based routing, T-144).
func (m ListModel) ActivateSearch() ListModel {
	m.search = m.search.Activate()
	return m
}

// DeactivateSearch clears the list search (T-143). Called on Esc, on Tab
// focus cycle leaving the list, and on filter change.
func (m ListModel) DeactivateSearch() ListModel {
	m.search = m.search.Deactivate()
	return m
}

// HasActiveSearch reports whether list search is currently open.
func (m ListModel) HasActiveSearch() bool { return m.search.IsActive() }

// SetFilter applies a filtered index (slice of entry indices that pass filters).
// On filter change, keeps selection if still passing, else moves to nearest.
// Any pinned out-of-filter level-jump match is cleared (the new filter set
// invalidates it). List search state is also cleared — the match set is
// no longer valid once the visible entries change (cavekit-entry-list
// R13 AC).
func (m ListModel) SetFilter(indices []int) ListModel {
	// Check if current cursor entry still passes.
	current := m.scroll.Cursor
	found := false
	nearest := 0
	for i, idx := range indices {
		if idx == current {
			found = true
			nearest = i
			break
		}
		if idx <= current {
			nearest = i
		}
	}
	m.filtered = indices
	m.pinnedFullIdx = -1
	m.search = m.search.Deactivate()
	if found {
		m.scroll.Cursor = current
	} else if len(indices) > 0 {
		m.scroll.Cursor = indices[nearest]
	}
	m.scroll.TotalEntries = len(indices)
	m.scroll.Offset = clampOffset(m.scroll.Offset, len(indices), m.scroll.ViewportHeight)
	return m
}

// ClearFilter removes the filter, showing all entries.
func (m ListModel) ClearFilter() ListModel {
	m.filtered = nil
	m.scroll.TotalEntries = len(m.entries)
	return m
}

// ClearTransient clears transient in-list state: the mark-nav / level-jump
// wrap indicator and any pinned out-of-filter level-jump match. Invoked when
// the list receives Esc with no higher-priority dismissal pending (T-097).
func (m ListModel) ClearTransient() ListModel {
	m.wrapDir = NoWrap
	m.pinnedFullIdx = -1
	return m
}

// HasTransient reports whether any transient state is set (wrap indicator
// active or an out-of-filter entry pinned for display).
func (m ListModel) HasTransient() bool {
	return m.wrapDir != NoWrap || m.pinnedFullIdx >= 0
}

// PinnedFullIdx returns the full-entries index of an out-of-filter level-jump
// match pinned for display, or -1 when none is pinned.
func (m ListModel) PinnedFullIdx() int { return m.pinnedFullIdx }

// visibleEntries returns the entries to display (filtered or all). When a
// pin is set and the pinned entry is not in the filter, it is spliced in at
// its sorted position so the user can see it with an outside-filter indicator.
func (m ListModel) visibleEntries() []logsource.Entry {
	out, _ := m.visibleEntriesAndPin()
	return out
}

// visibleEntriesAndPin returns the visible entries plus the visible-list
// position of the pinned out-of-filter entry, or -1 when no pin is showing.
func (m ListModel) visibleEntriesAndPin() ([]logsource.Entry, int) {
	if m.filtered == nil {
		return m.entries, -1
	}
	base := make([]logsource.Entry, len(m.filtered))
	for i, idx := range m.filtered {
		base[i] = m.entries[idx]
	}
	if m.pinnedFullIdx < 0 || m.pinnedFullIdx >= len(m.entries) {
		return base, -1
	}
	for _, idx := range m.filtered {
		if idx == m.pinnedFullIdx {
			// Pin already in the filtered set — no splice, no pin marker.
			return base, -1
		}
	}
	spliceAt := len(m.filtered)
	for i, idx := range m.filtered {
		if idx > m.pinnedFullIdx {
			spliceAt = i
			break
		}
	}
	out := make([]logsource.Entry, 0, len(m.filtered)+1)
	out = append(out, base[:spliceAt]...)
	out = append(out, m.entries[m.pinnedFullIdx])
	out = append(out, base[spliceAt:]...)
	return out, spliceAt
}

// Marks returns the mark set.
func (m ListModel) Marks() *MarkSet { return m.marks }

// rowForY converts a mouse Y coordinate (relative to the list top) to a
// visible-list cursor index, or -1 if out of bounds.
func (m ListModel) rowForY(y int) int {
	vis := m.visibleEntries()
	idx := m.scroll.Offset + y
	if idx < 0 || idx >= len(vis) {
		return -1
	}
	return idx
}

// Init satisfies tea.Model.
func (m ListModel) Init() tea.Cmd { return nil }

// Update handles keyboard navigation.
func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	// Always process WindowSizeMsg, even when entries are empty.
	// The initial resize arrives before async loading finishes.
	// T-100 wraps the list in a full lipgloss border (top/bottom + left/right),
	// so the inner viewport is 2 rows and 2 cols smaller than the outer
	// allocation handed in by the caller.
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		h := ws.Height - 2
		if h < 1 {
			h = 1
		}
		w := ws.Width - 2
		if w < 1 {
			w = 1
		}
		m.scroll.ViewportHeight = h
		m.width = w
		return m, nil
	}

	vis := m.visibleEntries()
	n := len(vis)
	if n == 0 {
		return m, nil
	}

	var cmd tea.Cmd
	prev := m.scroll.Cursor
	m.wrapDir = NoWrap

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// T-143 (cavekit-entry-list R13): when list search is active, all
		// typing / navigation keys route to search first. Esc dismisses
		// search, Enter commits input→navigate, Backspace edits the query,
		// runes extend the query (input mode), n/N cycle the cursor
		// through matches (navigate mode) honouring R12 scrolloff.
		if m.search.IsActive() {
			consumed := false
			m, consumed = m.handleSearchKey(msg, vis)
			if consumed {
				vis = m.visibleEntries()
				n = len(vis)
				if m.scroll.Cursor != prev && m.scroll.Cursor < len(vis) {
					entry := vis[m.scroll.Cursor]
					cmd = func() tea.Msg { return SelectionMsg{Entry: entry} }
				}
				return m, cmd
			}
		}
		switch msg.String() {
		case "j", "down":
			m.pinnedFullIdx = -1
			m.scroll.Cursor = clampCursor(m.scroll.Cursor+1, n)
		case "k", "up":
			m.pinnedFullIdx = -1
			m.scroll.Cursor = clampCursor(m.scroll.Cursor-1, n)
		case "g":
			m.pinnedFullIdx = -1
			m.scroll = GoTop(m.scroll)
		case "G":
			m.pinnedFullIdx = -1
			m.scroll = GoBottom(m.scroll)
		case "ctrl+d":
			m.pinnedFullIdx = -1
			m.scroll = HalfPageDown(m.scroll)
		case "ctrl+u":
			m.pinnedFullIdx = -1
			m.scroll = HalfPageUp(m.scroll)
		case "m":
			// Toggle mark on current visible entry.
			if m.scroll.Cursor < len(vis) {
				id := vis[m.scroll.Cursor].LineNumber
				m.marks.Toggle(id)
			}
		case "u":
			// Next mark with wrap. Mark nav drops any pin.
			m.pinnedFullIdx = -1
			vis = m.visibleEntries()
			n = len(vis)
			visIDs := make([]int, len(vis))
			for i, e := range vis {
				visIDs[i] = e.LineNumber
			}
			if next := m.marks.NextMark(m.scroll.Cursor, visIDs); next >= 0 {
				if next <= m.scroll.Cursor {
					m.wrapDir = WrapForward
				}
				m.scroll.Cursor = next
			}
		case "U":
			// Previous mark with wrap. Mark nav drops any pin.
			m.pinnedFullIdx = -1
			vis = m.visibleEntries()
			n = len(vis)
			visIDs := make([]int, len(vis))
			for i, e := range vis {
				visIDs[i] = e.LineNumber
			}
			if prevMark := m.marks.PrevMark(m.scroll.Cursor, visIDs); prevMark >= 0 {
				if prevMark >= m.scroll.Cursor {
					m.wrapDir = WrapBack
				}
				m.scroll.Cursor = prevMark
			}
		case "e":
			m = m.applyLevelJump(true, "ERROR")
		case "E":
			m = m.applyLevelJump(false, "ERROR")
		case "w":
			m = m.applyLevelJump(true, "WARN")
		case "W":
			m = m.applyLevelJump(false, "WARN")
		}
		// Pin or filter changes may have reshaped vis; recompute before
		// scroll housekeeping and selection emit.
		vis = m.visibleEntries()
		n = len(vis)
		m.scroll.Cursor = clampCursor(m.scroll.Cursor, n)
		// T-135 (F-026): keep cursor scrolloff rows from viewport edges
		// after every cursor-moving handler (j/k/g/G/Ctrl+d/Ctrl+u plus
		// level-jump and mark nav which also mutate Cursor above).
		m.scroll.TotalEntries = n
		m.scroll = followCursor(m.scroll)

		if m.scroll.Cursor != prev {
			if m.scroll.Cursor < len(vis) {
				entry := vis[m.scroll.Cursor]
				cmd = func() tea.Msg { return SelectionMsg{Entry: entry} }
			}
		}

	case tea.MouseMsg:
		vis := m.visibleEntries()
		n := len(vis)
		switch msg.Button {
		case tea.MouseButtonLeft:
			if msg.Action == tea.MouseActionPress {
				row := m.rowForY(msg.Y)
				if row >= 0 && row < n {
					m.scroll.Cursor = row
					m.scroll.Offset = ensureVisible(m.scroll.Cursor, m.scroll.Offset, m.scroll.ViewportHeight)
					if m.scroll.Cursor != prev {
						entry := vis[m.scroll.Cursor]
						cmd = func() tea.Msg { return SelectionMsg{Entry: entry} }
					}
				}
			} else if msg.Action == tea.MouseActionRelease {
				// Double-click detection: check if this is a second click on same row.
				// Bubble Tea exposes double-click as Action == tea.MouseActionMotion or
				// explicit double-click event. For simplicity, treat consecutive quick
				// presses as double-click — use DoubleClick action if available.
			}
		case tea.MouseButtonWheelDown:
			// T-135 (F-026): wheel scrolls offset first; cursor drags only
			// when it would enter the scrolloff margin.
			m.scroll.TotalEntries = n
			m.scroll = WheelDown(m.scroll)
			if m.scroll.Cursor != prev && m.scroll.Cursor < len(vis) {
				entry := vis[m.scroll.Cursor]
				cmd = func() tea.Msg { return SelectionMsg{Entry: entry} }
			}
		case tea.MouseButtonWheelUp:
			m.scroll.TotalEntries = n
			m.scroll = WheelUp(m.scroll)
			if m.scroll.Cursor != prev && m.scroll.Cursor < len(vis) {
				entry := vis[m.scroll.Cursor]
				cmd = func() tea.Msg { return SelectionMsg{Entry: entry} }
			}
		}
		// Timestamp-based double-click detection.
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			row := m.rowForY(msg.Y)
			now := time.Now()
			if row >= 0 && row == m.lastClickRow &&
				now.Sub(m.lastClickTime) < 500*time.Millisecond && n > 0 {
				// Double-click on same row within 500ms — open detail pane.
				entry := vis[m.scroll.Cursor]
				existingCmd := cmd
				cmd = func() tea.Msg {
					if existingCmd != nil {
						existingCmd() // drain selection msg
					}
					return OpenDetailPaneMsg{Entry: entry}
				}
				m.lastClickTime = time.Time{} // reset to prevent triple-click
			} else if row >= 0 {
				m.lastClickRow = row
				m.lastClickTime = now
			}
		}

	}

	return m, cmd
}

// entryIndexForVisible converts a visible-list cursor position to an index
// into the full m.entries slice. When a pin is spliced into the visible list
// the pin slot maps back to pinnedFullIdx and indices past it shift by one.
func (m ListModel) entryIndexForVisible(visIdx int) int {
	if m.filtered == nil {
		return visIdx
	}
	_, pinPos := m.visibleEntriesAndPin()
	if pinPos >= 0 {
		if visIdx == pinPos {
			return m.pinnedFullIdx
		}
		if visIdx > pinPos {
			fi := visIdx - 1
			if fi >= 0 && fi < len(m.filtered) {
				return m.filtered[fi]
			}
			return visIdx
		}
	}
	if visIdx >= 0 && visIdx < len(m.filtered) {
		return m.filtered[visIdx]
	}
	return visIdx
}

// visibleIndexForEntry converts a full-entries index to its visible index.
// When the entry is the pinned out-of-filter entry, returns the splice slot.
// When the entry is in the filter, returns its position (shifted by 1 if it
// sits after the pin splice). When neither, returns the current cursor.
func (m ListModel) visibleIndexForEntry(entryIdx int) int {
	if m.filtered == nil {
		return entryIdx
	}
	_, pinPos := m.visibleEntriesAndPin()
	if pinPos >= 0 && m.pinnedFullIdx == entryIdx {
		return pinPos
	}
	for vi, fi := range m.filtered {
		if fi == entryIdx {
			if pinPos >= 0 && vi >= pinPos {
				return vi + 1
			}
			return vi
		}
	}
	return m.scroll.Cursor
}

// indexInFilter reports whether a full-entries index appears in indices.
func indexInFilter(indices []int, idx int) bool {
	for _, fi := range indices {
		if fi == idx {
			return true
		}
	}
	return false
}

// applyLevelJump runs e/E/w/W navigation. It searches the full entry set,
// records wrap direction, and pins out-of-filter matches so the user can
// see them with an outside-filter indicator (R8 #6, #7-8).
func (m ListModel) applyLevelJump(forward bool, level string) ListModel {
	if len(m.entries) == 0 {
		return m
	}
	fullCur := m.entryIndexForVisible(m.scroll.Cursor)
	var newIdx int
	var dir WrapDirection
	if forward {
		newIdx, dir = NextLevel(m.entries, fullCur, level)
	} else {
		newIdx, dir = PrevLevel(m.entries, fullCur, level)
	}
	m.wrapDir = dir
	// Reset pin; re-pin if the match exists and is outside the filter.
	m.pinnedFullIdx = -1
	if newIdx == fullCur {
		// No match anywhere — leave cursor in place, no pin.
		return m
	}
	if m.filtered != nil && !indexInFilter(m.filtered, newIdx) {
		m.pinnedFullIdx = newIdx
	}
	m.scroll.Cursor = m.visibleIndexForEntry(newIdx)
	return m
}

// WrapDir returns the last wrap direction from a level-jump or mark nav.
func (m ListModel) WrapDir() WrapDirection { return m.wrapDir }

// handleSearchKey processes a KeyMsg while list search is active (T-143,
// cavekit-entry-list R13). Returns (updatedModel, consumed). When
// consumed == false, the caller should fall through to the normal
// navigation handlers — used for keys like `j`/`k`/`g`/`G` during navigate
// mode so the user can move the cursor while search stays open.
func (m ListModel) handleSearchKey(msg tea.KeyMsg, vis []logsource.Entry) (ListModel, bool) {
	key := msg.String()
	// Esc always dismisses, regardless of mode (R13 AC: "Esc dismisses
	// the search input, clears the SearchHighlight bg on all rows").
	if key == "esc" {
		m.search = m.search.Deactivate()
		return m, true
	}
	if m.search.InputMode() {
		switch key {
		case "enter":
			m.search = m.search.CommitInput()
			// Jump cursor to the first match if any so the next `n`/`N`
			// cycles from the actual match set, not the old cursor pos.
			if line := m.search.CurrentMatchLine(); line >= 0 && line < len(vis) {
				m.scroll.Cursor = line
				m.scroll.TotalEntries = len(vis)
				m.scroll = followCursor(m.scroll)
			}
			return m, true
		case "backspace", "ctrl+h":
			m.search = m.search.BackspaceRune(vis, m.width, m.cfg)
			return m, true
		default:
			// Append runes (covers a..z/A..Z/digits/space, and
			// multi-byte emoji/CJK via msg.Runes). Control keys with
			// no runes fall through as unconsumed.
			if len(msg.Runes) == 0 {
				return m, false
			}
			for _, r := range msg.Runes {
				m.search = m.search.AppendRune(r, vis, m.width, m.cfg)
			}
			return m, true
		}
	}
	// Navigate mode.
	switch key {
	case "n":
		m.search = m.search.Next()
		if line := m.search.CurrentMatchLine(); line >= 0 && line < len(vis) {
			m.scroll.Cursor = line
			m.scroll.TotalEntries = len(vis)
			m.scroll = followCursor(m.scroll)
		}
		return m, true
	case "N":
		m.search = m.search.Prev()
		if line := m.search.CurrentMatchLine(); line >= 0 && line < len(vis) {
			m.scroll.Cursor = line
			m.scroll.TotalEntries = len(vis)
			m.scroll = followCursor(m.scroll)
		}
		return m, true
	}
	// Any other key in navigate mode falls through to normal list
	// handling so the user can j/k/g/G while search stays visible.
	return m, false
}

// View renders exactly ViewportHeight rows — no more, no less.
// Rows are taken from offset..offset+ViewportHeight; shortfalls are padded with
// empty lines so the layout never overflows or underflows.
func (m ListModel) View() string {
	vis, pinPos := m.visibleEntriesAndPin()
	n := len(vis)
	vh := m.scroll.ViewportHeight
	if vh <= 0 {
		return ""
	}

	start := m.scroll.Offset
	end := start + vh
	if end > n {
		end = n
	}

	// T-143: precompute a match-set lookup for SearchHighlight rendering
	// on non-cursor rows. R13 AC: cursor-row bg takes visual priority
	// over SearchHighlight, so only non-cursor matches get the highlight.
	matchSet := map[int]bool{}
	if m.search.IsActive() {
		for _, idx := range m.search.MatchLines() {
			matchSet[idx] = true
		}
	}
	var sb strings.Builder
	rendered := 0
	for i := start; i < end; i++ {
		if rendered > 0 {
			sb.WriteByte('\n')
		}
		isCursor := i == m.scroll.Cursor
		// Single 2-cell prefix slot. Priority: pinned out-of-filter > wrap
		// indicator on cursor row > mark glyph. The wrap and pin glyphs are
		// transient (cleared on next nav or Esc); marks are persistent.
		prefix := ""
		switch {
		case pinPos >= 0 && i == pinPos:
			prefix = lipgloss.NewStyle().Foreground(m.th.LevelWarn).Render("⌀ ")
		case isCursor && m.wrapDir != NoWrap:
			prefix = lipgloss.NewStyle().Foreground(m.th.Mark).Render("↻ ")
		case m.marks.IsMarked(vis[i].LineNumber):
			prefix = lipgloss.NewStyle().Foreground(m.th.Mark).Render("* ")
		}
		switch {
		case isCursor:
			sb.WriteString(prefix + RenderCompactRowWithBg(vis[i], m.width, m.th, m.cfg, m.th.CursorHighlight))
		case matchSet[i]:
			// T-143: non-cursor match row → SearchHighlight bg.
			sb.WriteString(prefix + RenderCompactRowWithBg(vis[i], m.width, m.th, m.cfg, m.th.SearchHighlight))
		default:
			sb.WriteString(prefix + RenderCompactRow(vis[i], m.width, m.th, m.cfg))
		}
		rendered++
	}
	// Pad remaining lines so the list always occupies exactly ViewportHeight rows.
	for rendered < vh {
		if rendered > 0 {
			sb.WriteByte('\n')
		}
		rendered++
	}
	return m.applyPaneStyle(sb.String())
}

// applyPaneStyle wraps the rendered list in the DESIGN.md §4 pane style
// matrix (T-100/T-101). Focused or alone panes get FocusBorder borders;
// unfocused-but-visible panes get DividerColor borders + UnfocusedBg + Faint.
func (m ListModel) applyPaneStyle(content string) string {
	state := appshell.PaneStateUnfocused
	if m.Focused || m.Alone {
		state = appshell.PaneStateFocused
	}
	return appshell.PaneStyle(m.th, state).Render(content)
}

// RenderedRowCount returns how many entry rows were rendered in the last View() call.
// Used for benchmark validation.
func (m ListModel) RenderedRowCount() int {
	vis := m.visibleEntries()
	n := len(vis)
	if n == 0 {
		return 0
	}
	end := m.scroll.Offset + m.scroll.ViewportHeight
	if end > n {
		end = n
	}
	count := end - m.scroll.Offset
	if count < 0 {
		return 0
	}
	return count
}
