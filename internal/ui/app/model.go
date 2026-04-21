// Package app is the top-level Bubble Tea model that wires all subsystems together.
package app

import (
	"context"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/istonikula/gloggy/internal/config"
	"github.com/istonikula/gloggy/internal/filter"
	"github.com/istonikula/gloggy/internal/logsource"
	"github.com/istonikula/gloggy/internal/theme"
	"github.com/istonikula/gloggy/internal/ui/appshell"
	"github.com/istonikula/gloggy/internal/ui/detailpane"
	"github.com/istonikula/gloggy/internal/ui/entrylist"
	uifilter "github.com/istonikula/gloggy/internal/ui/filter"
)

// noticeClearMsg fires after the auto-close notice timeout to restore the
// key-hint bar (T-091).
type noticeClearMsg struct{}

// noticeClearAfter returns a tea.Cmd that fires noticeClearMsg after d.
func noticeClearAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return noticeClearMsg{} })
}

const autoCloseNoticeDuration = 3 * time.Second

const autoCloseNoticeText = "detail pane auto-closed — terminal too small"

// clipboardNoMarksNotice + clipboardNoticeDuration + clipboardErrorDuration
// back the visible feedback required by cavekit-app-shell R9 (T-138).
const (
	clipboardNoMarksNotice  = "no marked entries"
	clipboardNoticeDuration = 2 * time.Second
	clipboardErrorDuration  = 3 * time.Second
)

// formatClipboardCopiedNotice returns the status-bar text for a successful
// clipboard write — singular vs plural mirrors the kit AC phrasing.
func formatClipboardCopiedNotice(count int) string {
	if count == 1 {
		return "copied 1 entry"
	}
	return "copied " + strconv.Itoa(count) + " entries"
}

// Model is the root Bubble Tea model.
type Model struct {
	// config
	cfg        config.LoadResult
	th         theme.Theme
	configPath string

	// source
	sourceName string
	followMode bool
	tailCtx    context.Context
	tailCancel context.CancelFunc
	entries    []logsource.Entry

	// subsystems
	list        entrylist.ListModel
	pane        detailpane.PaneModel
	paneHeight  detailpane.HeightModel
	paneSearch  detailpane.SearchModel
	visibility  detailpane.VisibilityModel
	filterSet   *filter.FilterSet
	filterPanel uifilter.Model
	header      appshell.HeaderModel
	loading     appshell.LoadingModel
	keyhints    appshell.KeyHintBarModel
	help        appshell.HelpOverlayModel
	layout      appshell.LayoutModel
	resize      appshell.ResizeModel

	focus              appshell.FocusTarget
	cachedVisibleCount int
	// draggingDivider is true between a Press on the divider cell (vertical
	// in right-split, horizontal in below-mode) and the subsequent Release.
	// While set, Motion events update the active detail-pane ratio
	// (width_ratio or height_ratio per orientation) regardless of the
	// cursor's current zone. Focus is never modified by a drag (T-156,
	// cavekit-app-shell R15).
	draggingDivider bool
	// dragDirty tracks whether a Motion actually changed the ratio during
	// the current drag session (T-164, F-129). Release only persists to
	// disk when true — a bare Press+Release with no Motion leaves the
	// config untouched. Reset to false on Press.
	dragDirty bool
}

// New creates the root model from parsed CLI args and loaded config.
// configPath is the path to the config file (for writeback).
func New(sourceName string, followMode bool, configPath string, cfgResult config.LoadResult) Model {
	th := theme.GetTheme(cfgResult.Config.Theme)
	fs := filter.NewFilterSet()
	ctx, cancel := context.WithCancel(context.Background())

	m := Model{
		cfg:         cfgResult,
		th:          th,
		configPath:  configPath,
		sourceName:  sourceName,
		followMode:  followMode,
		tailCtx:     ctx,
		tailCancel:  cancel,
		filterSet:   fs,
		filterPanel: uifilter.New(fs),
		help:        appshell.NewHelpOverlayModel(),
		resize:      appshell.NewResizeModel(80, 24).WithConfig(cfgResult.Config),
		focus:       appshell.FocusEntryList,
	}

	// These get proper dimensions on the first WindowSizeMsg.
	m.list = entrylist.NewListModel(th, cfgResult.Config, 80, 22)
	m.pane = detailpane.NewPaneModel(th, 8)
	m.paneHeight = detailpane.NewHeightModel(cfgResult.Config.DetailPane.HeightRatio, 24)
	m.paneSearch = detailpane.NewSearchModel(th)
	m.visibility = detailpane.NewVisibilityModel(configPath, cfgResult)
	m.header = appshell.NewHeaderModel(th, 80).WithSource(sourceName).WithFollow(followMode)
	m.loading = appshell.NewLoadingModel()
	m.keyhints = appshell.NewKeyHintBarModel(th, 80)
	m.layout = appshell.NewLayoutModel(80, 24).WithTheme(th)

	// Activate loading indicator upfront when we know a file will be loaded.
	// Init() is a value receiver and cannot mutate the model, so we initialise
	// this here instead.
	if sourceName != "" && !followMode {
		m.loading = m.loading.Start()
	}

	return m
}

// Init kicks off initial loading.
func (m Model) Init() tea.Cmd {
	if m.sourceName == "" {
		// stdin — read synchronously before TUI starts (handled in main).
		return nil
	}
	if m.followMode {
		// cavekit-log-source.md R8 AC4: startLineNum=0 means "emit everything",
		// so follow mode opens with the existing file contents visible, then
		// streams appends as they arrive.
		return logsource.TailFile(m.tailCtx, m.sourceName, 0)
	}
	return logsource.LoadFile(m.sourceName)
}

// Update is the central message dispatcher.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Help overlay intercepts everything while open.
	var forward bool
	m.help, forward = m.help.Update(msg)
	if !forward {
		return m, nil
	}

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.resize, _ = m.resize.Update(msg)
		m.paneHeight, _ = m.paneHeight.Update(msg)
		w, h := m.resize.Width(), m.resize.Height()
		m.layout = m.layout.
			SetSize(w, h).
			SetDetailPane(m.pane.IsOpen(), m.paneHeight.PaneHeight()).
			SetOrientation(m.resize.Orientation()).
			SetWidthRatio(m.cfg.Config.DetailPane.WidthRatio)
		// T-091: auto-close pane when its content shrinks below the
		// minimum-viable threshold for the active orientation.
		var cmd tea.Cmd
		if appshell.ShouldAutoCloseDetail(m.layout.Layout()) {
			m.pane = m.pane.Close()
			m.focus = appshell.FocusEntryList
			m.layout = m.layout.SetDetailPane(false, m.paneHeight.PaneHeight())
			m.keyhints = m.keyhints.WithFocus(appshell.FocusEntryList).WithNotice(autoCloseNoticeText)
			cmd = noticeClearAfter(autoCloseNoticeDuration)
			// T-162 (F-125): belt-and-braces — terminate any active
			// divider-drag session here so subsequent Motion/Release
			// events in the stream cannot mutate ratios or write config.
			m.draggingDivider = false
			m.dragDirty = false
		}
		l := m.layout.Layout()
		m.list = m.list.WithScrolloff(m.cfg.Config.Scrolloff).WithContentTopY(l.ListContentTopY())
		m.list, _ = m.list.Update(tea.WindowSizeMsg{Width: l.ListContentWidth(), Height: l.EntryListHeight()})
		// T-123 (F-013): vertical allocation is orientation-aware — in
		// right-split the pane gets the full main-area slot, not
		// height_ratio × terminalHeight.
		// F-200 (R7 AC 9): WindowSizeMsg-driven orientation flips must
		// refresh pane-local orientation-dependent flags so the R10
		// drag-seam top-border paint tracks the new orientation — symmetric
		// with relayout() so neither path leaves belowMode stale.
		m.pane = m.pane.
			WithScrolloff(m.cfg.Config.Scrolloff).
			WithBelowMode(m.resize.Orientation() == appshell.OrientationBelow).
			SetHeight(appshell.DetailPaneVerticalRows(l)).
			SetWidth(l.DetailContentWidth())
		m.header = m.header.WithWidth(w)
		m.keyhints = m.keyhints.WithWidth(w).WithPaneOpen(m.pane.IsOpen())
		return m, cmd

	case noticeClearMsg:
		m.keyhints = m.keyhints.WithNotice("")
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.MouseMsg:
		return m.handleMouse(msg)

	// Background loading stream.
	case logsource.LoadFileStreamMsg:
		inner := msg.Unwrap()
		var cmd tea.Cmd
		switch inner := inner.(type) {
		case logsource.EntryBatchMsg:
			m.entries = append(m.entries, inner.Entries...)
			m = m.appendToList(inner.Entries)
			m.loading = m.loading.Update(len(m.entries))
			if len(m.filterSet.GetEnabled()) == 0 {
				m.cachedVisibleCount = len(m.entries)
			}
			m.header = m.header.WithCounts(len(m.entries), m.visibleCount())
			cmd = msg.Next()
		case logsource.LoadProgressMsg:
			m.loading = m.loading.Update(inner.Count)
			cmd = msg.Next()
		case logsource.LoadDoneMsg:
			m.loading = m.loading.Done()
		}
		return m, cmd

	case logsource.LoadDoneMsg:
		m.loading = m.loading.Done()
		return m, nil

	// Tail stream.
	case logsource.TailStreamMsg:
		inner := msg.Unwrap()
		var cmd tea.Cmd
		switch inner := inner.(type) {
		case logsource.TailMsg:
			m.entries = append(m.entries, inner.Entries...)
			m = m.appendToList(inner.Entries)
			m.header = m.header.WithCounts(len(m.entries), m.visibleCount())
			cmd = msg.Next()
		case logsource.TailStopMsg:
			// Watcher stopped; nothing to do.
		}
		return m, cmd

	// Entry selection from list.
	case entrylist.SelectionMsg:
		if m.pane.IsOpen() {
			// T-127 (F-020): live-preview re-render must honor the current
			// hiddenFields set — without this, cursor-driven re-renders
			// would show suppressed fields until the user re-openPanes.
			m.pane = m.pane.
				WithScrolloff(m.cfg.Config.Scrolloff).
				WithHiddenFields(m.visibility.HiddenFields()).
				Open(msg.Entry)
		}
		return m, nil

	// Double-click / Enter on entry → open detail pane.
	case entrylist.OpenDetailPaneMsg:
		m = m.openPane(msg.Entry)
		return m, nil

	// Detail pane closed.
	case detailpane.BlurredMsg:
		m.focus = appshell.FocusEntryList
		m.keyhints = m.keyhints.WithFocus(appshell.FocusEntryList)
		// T-117 (F-006): dismiss any lingering search state so the
		// next time the pane opens there is no stale query or match
		// set carried over from a previous entry.
		m.paneSearch = m.paneSearch.Dismiss()
		m = m.relayout()
		return m, nil

	// Filter confirmed from prompt.
	case uifilter.FilterConfirmedMsg:
		m = m.refilter()
		return m, nil

	// Filter panel changed.
	case uifilter.FilterChangedMsg:
		m = m.refilter()
		return m, nil

	// T-138 (cavekit-app-shell R9): surface clipboard result to the user.
	case appshell.ClipboardCopiedMsg:
		m.keyhints = m.keyhints.WithNotice(formatClipboardCopiedNotice(msg.Count))
		return m, noticeClearAfter(clipboardNoticeDuration)
	case appshell.ClipboardErrorMsg:
		m.keyhints = m.keyhints.WithNotice("clipboard error: " + msg.Err.Error())
		return m, noticeClearAfter(clipboardErrorDuration)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit. T-146 (cavekit-app-shell R14): exempted while the list
	// search is capturing input so `q` typed into the query string is not
	// swallowed as a quit. Navigate mode (post-Enter) still quits — users
	// can dismiss with Esc first if they want `q` to extend the query.
	if msg.String() == "q" && m.focus == appshell.FocusEntryList {
		if !(m.list.HasActiveSearch() && m.list.Search().InputMode()) {
			if m.tailCancel != nil {
				m.tailCancel()
			}
			return m, tea.Quit
		}
	}

	// T-096: Tab cycles focus between visible panes. Tab never closes a
	// pane and is inert while any overlay (filter panel or help) is open.
	// Help is handled earlier in Update (intercept); the filter panel is
	// overlay-like when focused.
	if msg.String() == "tab" {
		visible := []appshell.FocusTarget{appshell.FocusEntryList}
		if m.pane.IsOpen() {
			visible = append(visible, appshell.FocusDetailPane)
		}
		overlayOpen := m.focus == appshell.FocusFilterPanel
		next := appshell.NextFocus(m.focus, visible, overlayOpen)
		if next != m.focus {
			prevFocus := m.focus
			m.focus = next
			m.keyhints = m.keyhints.WithFocus(m.focus)
			// T-143 (cavekit-entry-list R13): list search state clears
			// when the list loses focus via Tab cycle.
			if prevFocus == appshell.FocusEntryList && m.list.HasActiveSearch() {
				m.list = m.list.DeactivateSearch()
			}
		}
		return m, nil
	}

	// T-155 (cavekit-app-shell R12, revised): focus-aware resize keymap
	// (+/-/=/|). Routes AFTER list-search handoff (handled below in the
	// FocusEntryList branch) but BEFORE any other focus-specific
	// dispatch, so a ratio key fires whichever pane is focused (detail
	// or list). Ratio is held on the detail pane: in below-orientation
	// the active storage is height_ratio, in right-orientation
	// width_ratio. When the detail pane is closed, all four keys are
	// silent no-ops — there is no divider to move.
	if appshell.IsRatioKey(msg.String()) && m.focus != appshell.FocusFilterPanel {
		if m.focus == appshell.FocusEntryList && m.list.HasActiveSearch() && m.list.Search().InputMode() {
			// List-search input mode consumes every rune — ratio keys
			// extend the query rather than resizing.
		} else if m.pane.IsOpen() {
			return m.handleRatioKey(msg.String())
		}
	}

	switch m.focus {
	case appshell.FocusDetailPane:
		// In-pane search. T-113/T-114: match against ContentLines
		// (soft-wrapped, unstyled) — splitting View() would include
		// border glyphs and syntax-highlight ANSI and mis-index matches.
		if m.paneSearch.IsActive() || msg.String() == "/" {
			// T-118 (F-008): in navigation mode, forward non-search keys
			// (j/k/scrolling/etc.) to the pane so the user can move the
			// viewport while search stays visibly open. The search-owned
			// keys below stay with paneSearch.
			if m.paneSearch.IsActive() && m.paneSearch.Mode() == detailpane.SearchModeNavigate {
				switch msg.String() {
				case "/", "n", "N", "esc", "enter", "backspace", "ctrl+h":
					// fall through to paneSearch.Update below
				default:
					var cmd tea.Cmd
					m.pane, cmd = m.pane.Update(msg)
					return m, cmd
				}
			}
			lines := m.pane.ContentLines()
			m.paneSearch, _ = m.paneSearch.Update(msg, lines)
			// T-115: after any search update that has a current match,
			// scroll the pane viewport so the match line is visible.
			if m.paneSearch.MatchCount() > 0 {
				m.pane = m.pane.ScrollToLine(m.paneSearch.CurrentMatchLine())
			}
			return m, nil
		}
		var cmd tea.Cmd
		m.pane, cmd = m.pane.Update(msg)
		return m, cmd

	case appshell.FocusFilterPanel:
		var cmd tea.Cmd
		m.filterPanel, cmd = m.filterPanel.Update(msg)
		// Esc or Enter from filter panel returns focus to list.
		if msg.String() == "esc" {
			m.focus = appshell.FocusEntryList
			m.keyhints = m.keyhints.WithFocus(appshell.FocusEntryList)
		}
		return m, cmd

	default: // FocusEntryList
		// T-143/T-144 (cavekit-entry-list R13): when list search is
		// active, keys that belong to the search handler must reach the
		// list before app-level single-key intercepts (/, f, y, enter,
		// esc) swallow them. Tab is handled earlier and clears search
		// via the focus-cycle branch, so it is not affected by this
		// route.
		//   Input mode: every key goes to list (typing builds the
		//   query, Enter commits to navigate mode, Backspace edits,
		//   Esc dismisses).
		//   Navigate mode: only Esc is redirected here so dismissal
		//   takes priority over `Esc closes pane` and `Esc clears
		//   transient`. `n`/`N` / `j` / `k` / `Enter` / etc. fall
		//   through — Enter on the current matched row opens the
		//   detail pane, navigation keys reach list.Update via the
		//   default fallthrough (handleSearchKey passes them on).
		if m.list.HasActiveSearch() && (m.list.Search().InputMode() || msg.String() == "esc") {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
		// T-097 priority 3: Esc on list clears transient state (wrap
		// indicator from level-jump / mark-nav). No-op otherwise.
		// Priorities 1 (help/filter overlays) and 2 (detail-pane close)
		// are handled earlier — the help overlay intercepts Esc at the
		// top of Update, the filter panel case closes itself, and the
		// detail-pane branch forwards Esc to pane.Update (T-041).
		if msg.String() == "esc" {
			// T-126 (F-024): Esc from list-focus with the pane open closes
			// the pane. Users do not need to Tab to the pane first just to
			// dismiss it — Esc is the single dismissal key everywhere.
			if m.pane.IsOpen() {
				m.pane = m.pane.Close()
				m.paneSearch = m.paneSearch.Dismiss()
				m = m.relayout()
				return m, nil
			}
			if m.list.HasTransient() {
				m.list = m.list.ClearTransient()
			}
			return m, nil
		}
		// Open filter panel.
		if msg.String() == "f" {
			// T-147 (cavekit-entry-list R13 AC 7): clear list-search state
			// when focus transfers to the filter panel so a stale query +
			// highlights do not linger after the user switches contexts.
			if m.list.HasActiveSearch() {
				m.list = m.list.DeactivateSearch()
			}
			m.focus = appshell.FocusFilterPanel
			m.keyhints = m.keyhints.WithFocus(appshell.FocusFilterPanel)
			return m, nil
		}
		// Enter opens detail pane.
		if msg.String() == "enter" {
			if entry, ok := m.list.SelectedEntry(); ok {
				m = m.openPane(entry)
				return m, nil
			}
		}
		// T-144 (cavekit-app-shell R13 revised, cavekit-entry-list R13):
		// focus-based `/` routing. When the list is focused, `/` opens
		// the list-scope free-text search (T-143) — regardless of
		// whether the detail pane is also open. The old behaviour
		// (transfer focus + activate pane search, or show "open entry
		// first" notice) is replaced because the list now has its own
		// search. Pane search is reached via Tab → pane then `/`, or
		// via clicking the pane then `/`.
		if msg.String() == "/" {
			m.list = m.list.ActivateSearch()
			return m, nil
		}
		// Clipboard copy of marked entries (cavekit-app-shell R9).
		// Every `y` press MUST produce visible feedback — success count,
		// error detail, or "no marked entries". Never a silent action.
		if msg.String() == "y" {
			marks := m.list.Marks()
			markedIDs := make(map[int]bool)
			for _, e := range m.entries {
				if marks.IsMarked(e.LineNumber) {
					markedIDs[e.LineNumber] = true
				}
			}
			if len(markedIDs) == 0 {
				m.keyhints = m.keyhints.WithNotice(clipboardNoMarksNotice)
				return m, noticeClearAfter(clipboardNoticeDuration)
			}
			return m, appshell.CopyMarkedEntriesCmd(m.entries, markedIDs)
		}
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	router := appshell.NewMouseRouter(m.layout.Layout())
	zone := router.RouteMouseMsg(msg)

	// T-156 (cavekit-app-shell R15): mouse-drag resize on the divider
	// works in BOTH orientations. Press on the divider cell starts a
	// drag; subsequent motion events update the active detail-pane
	// ratio live (width_ratio in right-split, height_ratio in
	// below-mode). Release saves config exactly once. Drag is
	// focus-neutral (returns before the click-focus transfer below) and
	// is a no-op when the detail pane is closed (no divider exists to
	// anchor the drag). T-104 was the right-split-only precursor.
	if msg.Button == tea.MouseButtonLeft {
		if msg.Action == tea.MouseActionPress && zone == appshell.ZoneDivider && m.pane.IsOpen() {
			m.draggingDivider = true
			m.dragDirty = false // T-164 (F-129): reset on new Press.
			// T-164 (F-129) + T-161 X-axis audit: Press does NOT apply
			// the inverse-math mapping. Earlier code fell through into
			// the Motion branch on Press, which rewrote the ratio even
			// on a bare click. Press now only seeds the drag-session
			// state; only explicit Motion moves the divider.
			return m, nil
		}
		if m.draggingDivider {
			// T-162 (F-125): if the pane was auto-closed mid-drag (e.g.
			// terminal resize via WindowSizeMsg), the session state is
			// stale. Swallow the event and end the session without
			// mutating the ratio or writing config.
			if !m.pane.IsOpen() {
				m.draggingDivider = false
				m.dragDirty = false
				return m, nil
			}
			if msg.Action == tea.MouseActionRelease {
				wasDirty := m.dragDirty
				m.draggingDivider = false
				m.dragDirty = false
				if wasDirty {
					m.saveConfig() // T-099: persist final ratio on drag release.
				}
				// T-164 (F-129): a bare Press+Release with no Motion
				// leaves the config file untouched.
				return m, nil
			}
			// Only Motion events move the divider. Ignore repeated
			// Press or other mouse actions that arrive during an
			// active session.
			if msg.Action != tea.MouseActionMotion {
				return m, nil
			}
			// T-165 (F-126): degenerate-terminal guard. When the active
			// dimension is ≤ 0 the inverse-math helpers can only return
			// the clamped default, which would shadow the user's
			// persisted ratio. Skip the ratio write entirely in that
			// case — the next real Motion (after the terminal regrows)
			// will apply a meaningful update.
			if m.resize.Orientation() == appshell.OrientationRight {
				termW := m.resize.Width()
				if termW <= 0 {
					return m, nil
				}
				newR := appshell.RatioFromDragX(msg.X, termW)
				if newR != m.cfg.Config.DetailPane.WidthRatio {
					m.dragDirty = true
				}
				m.cfg.Config.DetailPane.WidthRatio = newR
				m.layout = m.layout.SetWidthRatio(newR)
			} else {
				termH := m.resize.Height()
				if termH <= 0 {
					return m, nil
				}
				newR := appshell.RatioFromDragY(msg.Y, termH)
				if newR != m.cfg.Config.DetailPane.HeightRatio {
					m.dragDirty = true
				}
				m.paneHeight = m.paneHeight.SetRatio(newR)
				m.cfg.Config.DetailPane.HeightRatio = newR
				m.pane = m.pane.SetHeight(m.paneHeight.PaneHeight())
			}
			m = m.relayout()
			return m, nil
		}
	}

	// T-095: click-to-focus. After a click routes to a visible pane,
	// transfer focus there so subsequent key events target that pane.
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
		switch zone {
		case appshell.ZoneEntryList:
			if m.focus != appshell.FocusEntryList {
				m.focus = appshell.FocusEntryList
				m.keyhints = m.keyhints.WithFocus(appshell.FocusEntryList)
			}
		case appshell.ZoneDetailPane:
			if m.pane.IsOpen() && m.focus != appshell.FocusDetailPane {
				// T-143: focus leaving the list clears list search.
				if m.list.HasActiveSearch() {
					m.list = m.list.DeactivateSearch()
				}
				m.focus = appshell.FocusDetailPane
				m.keyhints = m.keyhints.WithFocus(appshell.FocusDetailPane)
			}
		}
	}

	switch zone {
	case appshell.ZoneEntryList:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	case appshell.ZoneDetailPane:
		var cmd tea.Cmd
		m.pane, cmd = m.pane.Update(msg)
		return m, cmd
	}
	return m, nil
}

// saveConfig writes the current config to disk (T-099). Errors are
// intentionally swallowed — a failed write should not interrupt the UI.
// The user will see their edits take effect live regardless; the next
// launch picks up whatever the last successful write left on disk.
func (m Model) saveConfig() {
	if m.configPath == "" {
		return
	}
	_ = config.Save(m.configPath, m.cfg)
}

// View composes the full screen.
func (m Model) View() string {
	if m.help.IsOpen() {
		return m.help.View()
	}

	// R14: FOLLOW badge lights when tail mode is active AND cursor is on the
	// last entry. Upward nav off the last row clears it; `G` restores it.
	header := m.header.
		WithCursorPos(m.list.CursorPosition()).
		WithFollow(m.followMode && m.list.IsAtTail()).
		View()

	// T-100/T-101: set per-pane focus + alone state before View() so each
	// pane applies the DESIGN.md §4 visual matrix.
	paneOpen := m.pane.IsOpen()
	m.list.Focused = (m.focus == appshell.FocusEntryList)
	m.list.Alone = !paneOpen
	list := m.list.View()

	paneView := ""
	if paneOpen {
		m.pane.Focused = (m.focus == appshell.FocusDetailPane)
		// T-114: attach the app's paneSearch so the pane renders the
		// prompt row, (cur/total) counter, and highlights.
		m.pane = m.pane.WithSearch(m.paneSearch)
		paneView = m.pane.View()
	}

	// T-143 (cavekit-entry-list R13): when list search is active, the
	// key-hint bar surfaces the prompt + query + (cur/total) match
	// counter — or "No matches" when the query yields zero hits.
	if m.list.HasActiveSearch() && !m.keyhints.HasNotice() {
		m.keyhints = m.keyhints.WithNotice(formatListSearchNotice(m.list.Search()))
	}

	status := m.keyhints.View()
	if m.loading.IsActive() {
		status = m.loading.View()
	}

	return m.layout.Render(header, list, paneView, status)
}

// formatListSearchNotice builds the status-bar text shown while list
// search is active (T-143, cavekit-entry-list R13). Input mode shows a
// trailing `_` cursor; navigate mode shows `(cur/total)`. An empty
// query in input mode just shows the prompt. A non-empty query with
// zero matches shows "No matches".
func formatListSearchNotice(s entrylist.SearchModel) string {
	if s.InputMode() {
		if s.Query() == "" {
			return "/: "
		}
		if s.NotFound() {
			return "/: " + s.Query() + "  (No matches)"
		}
		return "/: " + s.Query() + "_"
	}
	if s.MatchCount() == 0 {
		return "/: " + s.Query() + "  (No matches) — Esc to dismiss"
	}
	return "/: " + s.Query() + "  (" + strconv.Itoa(s.CurrentIndex()+1) + "/" + strconv.Itoa(s.MatchCount()) + ") n/N next/prev, Esc dismiss"
}

// SetEntries loads entries synchronously (used for stdin).
func (m Model) SetEntries(entries []logsource.Entry) Model {
	m.entries = entries
	m.cachedVisibleCount = len(entries)
	m.list = m.list.SetEntries(entries)
	m.header = m.header.WithCounts(len(entries), len(entries))
	return m
}

// appendToList forwards new entries to the list and, when the list's
// tail-follow snap moved the cursor (cavekit-entry-list R14), re-syncs
// the detail pane to the newly selected entry in the same frame. Without
// this sync, the cursor advances to the appended entry but the pane
// keeps rendering the previous selection until the user presses a key —
// silently breaking the live-preview invariant (cavekit-detail-pane R1).
// Pane re-sync runs only when the cursor actually moved, so appends that
// leave the cursor in place (user scrolled away from tail) do not
// clobber the user's pane selection.
func (m Model) appendToList(entries []logsource.Entry) Model {
	prevCursor := m.list.Cursor()
	m.list = m.list.AppendEntries(entries)
	if m.pane.IsOpen() && m.list.Cursor() != prevCursor {
		if entry, ok := m.list.SelectedEntry(); ok {
			m.pane = m.pane.
				WithScrolloff(m.cfg.Config.Scrolloff).
				WithHiddenFields(m.visibility.HiddenFields()).
				Open(entry)
		}
	}
	return m
}

func (m Model) openPane(entry logsource.Entry) Model {
	// T-127 (F-020): wire config-driven hidden fields into the pane BEFORE
	// Open, so JSON rendering skips suppressed keys. Previously the pane
	// hardcoded `nil` internally, so visibility config never reached it.
	// T-132 (F-026): seed scrolloff from config so cursor-tracking nav
	// keeps the configured context margin inside the pane.
	m.pane = m.pane.
		WithScrolloff(m.cfg.Config.Scrolloff).
		WithHiddenFields(m.visibility.HiddenFields()).
		Open(entry)
	// T-126 (F-017): opening the pane does NOT transfer keyboard focus —
	// focus stays on the entry list so `j`/`k` keep moving the cursor and
	// the pane acts as a live preview. Focus transfers only on explicit
	// user action: Tab (R11), mouse click on pane (R6), cross-pane `/` (R13).
	m.keyhints = m.keyhints.WithPaneOpen(true)
	// T-117 (F-006): a fresh entry must start with clean search state so
	// matches and the query do not leak across open/close cycles.
	m.paneSearch = m.paneSearch.Dismiss()
	m = m.relayout()
	return m
}

func (m Model) relayout() Model {
	m.keyhints = m.keyhints.WithPaneOpen(m.pane.IsOpen())
	m.layout = m.layout.
		SetDetailPane(m.pane.IsOpen(), m.paneHeight.PaneHeight()).
		SetOrientation(m.resize.Orientation()).
		SetWidthRatio(m.cfg.Config.DetailPane.WidthRatio)
	l := m.layout.Layout()
	m.list = m.list.WithScrolloff(m.cfg.Config.Scrolloff).WithContentTopY(l.ListContentTopY())
	m.list, _ = m.list.Update(tea.WindowSizeMsg{
		Width:  l.ListContentWidth(),
		Height: l.EntryListHeight(),
	})
	// T-123 (F-013, F-014): recompute pane vertical allocation on every
	// relayout (open, ratio change, orientation flip). In right-split that
	// means the full main-area slot; in below-mode it stays height_ratio.
	// T-173: below-mode flag drives drag-seam top-border rendering.
	m.pane = m.pane.
		WithScrolloff(m.cfg.Config.Scrolloff).
		WithBelowMode(m.resize.Orientation() == appshell.OrientationBelow).
		SetHeight(appshell.DetailPaneVerticalRows(l)).
		SetWidth(l.DetailContentWidth())
	return m
}

func (m Model) refilter() Model {
	indices := filter.Apply(m.filterSet, m.entries)
	m.cachedVisibleCount = len(indices)
	m.list = m.list.SetFilter(indices)
	m.header = m.header.WithCounts(len(m.entries), len(indices))
	return m
}

func (m Model) visibleCount() int {
	return m.cachedVisibleCount
}

// handleRatioKey applies a focus-aware resize key press (cavekit-app-shell
// R12 revised, T-155). Caller guarantees the detail pane is open, the key
// is a ratio key, and focus is not the filter panel. Returns the updated
// model + live-write command.
func (m Model) handleRatioKey(key string) (tea.Model, tea.Cmd) {
	listFocused := m.focus == appshell.FocusEntryList
	var current, newR float64
	if m.resize.Orientation() == appshell.OrientationRight {
		current = m.cfg.Config.DetailPane.WidthRatio
		newR, _ = appshell.NextDetailRatio(current, key, listFocused)
		m.cfg.Config.DetailPane.WidthRatio = newR
		m.layout = m.layout.SetWidthRatio(newR)
	} else {
		current = m.paneHeight.Ratio()
		newR, _ = appshell.NextDetailRatio(current, key, listFocused)
		m.paneHeight = m.paneHeight.SetRatio(newR)
		m.cfg.Config.DetailPane.HeightRatio = newR
		m.pane = m.pane.SetHeight(m.paneHeight.PaneHeight())
	}
	m = m.relayout()
	// T6 / B3: skip saveConfig at clamp-pin or preset no-op — unconditional
	// save advanced config mtime on every repeated boundary press.
	if newR != current {
		m.saveConfig() // T-099: persist ratio change immediately.
	}
	return m, nil
}
