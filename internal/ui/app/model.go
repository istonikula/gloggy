// Package app is the top-level Bubble Tea model that wires all subsystems together.
package app

import (
	"context"
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

// searchNoPaneNotice is shown when `/` is pressed with the detail pane
// closed (T-116, app-shell R13). Auto-dismisses via noticeClearAfter.
const searchNoPaneNotice = "open an entry first (Enter) to search"

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
	// draggingDivider is true between a Press on the right-split
	// divider column and the subsequent Release. While set, Motion
	// events update width_ratio regardless of the cursor's current zone
	// (T-104).
	draggingDivider bool
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
		return logsource.TailFile(m.tailCtx, m.sourceName, 1)
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
		}
		l := m.layout.Layout()
		m.list, _ = m.list.Update(tea.WindowSizeMsg{Width: l.ListContentWidth(), Height: l.EntryListHeight()})
		// T-123 (F-013): vertical allocation is orientation-aware — in
		// right-split the pane gets the full main-area slot, not
		// height_ratio × terminalHeight.
		m.pane = m.pane.SetHeight(appshell.DetailPaneVerticalRows(l)).SetWidth(l.DetailContentWidth())
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
			m.list = m.list.AppendEntries(inner.Entries)
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
			m.entries = append(m.entries, inner.Entry)
			m.list = m.list.AppendEntries([]logsource.Entry{inner.Entry})
			m.header = m.header.WithCounts(len(m.entries), m.visibleCount())
			cmd = msg.Next()
		case logsource.TailStopMsg:
			// Watcher stopped; nothing to do.
		}
		return m, cmd

	// Entry selection from list.
	case entrylist.SelectionMsg:
		if m.pane.IsOpen() {
			m.pane = m.pane.Open(msg.Entry)
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
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit.
	if msg.String() == "q" && m.focus == appshell.FocusEntryList {
		if m.tailCancel != nil {
			m.tailCancel()
		}
		return m, tea.Quit
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
			m.focus = next
			m.keyhints = m.keyhints.WithFocus(m.focus)
		}
		return m, nil
	}

	switch m.focus {
	case appshell.FocusDetailPane:
		// T-098: ratio keymap (+/-/=/|). Active ratio depends on
		// orientation — height_ratio in below-mode, width_ratio in
		// right-mode. Clamps to [0.10, 0.80].
		if appshell.IsRatioKey(msg.String()) {
			if m.resize.Orientation() == appshell.OrientationRight {
				newR, _ := appshell.NextRatio(m.cfg.Config.DetailPane.WidthRatio, msg.String())
				m.cfg.Config.DetailPane.WidthRatio = newR
				m.layout = m.layout.SetWidthRatio(newR)
			} else {
				newR, _ := appshell.NextRatio(m.paneHeight.Ratio(), msg.String())
				m.paneHeight = m.paneHeight.SetRatio(newR)
				m.cfg.Config.DetailPane.HeightRatio = newR
				m.pane = m.pane.SetHeight(m.paneHeight.PaneHeight())
			}
			m = m.relayout()
			m.saveConfig() // T-099: persist ratio change immediately.
			return m, nil
		}
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
		// T-116 (app-shell R13 / F-001, F-011): cross-pane `/`.
		// - Pane open: transfer focus to the pane AND activate search in
		//   a single keystroke so the user does not have to Tab first.
		// - Pane closed: emit a transient notice and auto-dismiss it so
		//   `/` is never a silent no-op.
		if msg.String() == "/" {
			if m.pane.IsOpen() {
				m.focus = appshell.FocusDetailPane
				m.keyhints = m.keyhints.WithFocus(appshell.FocusDetailPane)
				lines := m.pane.ContentLines()
				m.paneSearch, _ = m.paneSearch.Update(msg, lines)
				return m, nil
			}
			m.keyhints = m.keyhints.WithNotice(searchNoPaneNotice)
			return m, noticeClearAfter(autoCloseNoticeDuration)
		}
		// Clipboard copy of marked entries.
		if msg.String() == "y" {
			marks := m.list.Marks()
			markedIDs := make(map[int]bool)
			for _, e := range m.entries {
				if marks.IsMarked(e.LineNumber) {
					markedIDs[e.LineNumber] = true
				}
			}
			appshell.CopyMarkedEntries(m.entries, markedIDs) //nolint:errcheck
			return m, nil
		}
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	router := appshell.NewMouseRouter(m.layout.Layout())
	zone := router.RouteMouseMsg(msg)

	// T-104: drag on the vertical divider in right-split resizes the
	// detail pane width. A Press at the divider column starts a drag
	// session; subsequent Motion events update width_ratio from the
	// cursor's x even when the cursor has moved outside the divider.
	// Release ends the session.
	if m.resize.Orientation() == appshell.OrientationRight && msg.Button == tea.MouseButtonLeft {
		if msg.Action == tea.MouseActionPress && zone == appshell.ZoneDivider {
			m.draggingDivider = true
		}
		if m.draggingDivider {
			if msg.Action == tea.MouseActionRelease {
				m.draggingDivider = false
				m.saveConfig() // T-099: persist final width_ratio on drag release.
				return m, nil
			}
			newR := appshell.RatioFromDragX(msg.X, m.resize.Width())
			m.cfg.Config.DetailPane.WidthRatio = newR
			m.layout = m.layout.SetWidthRatio(newR)
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

	header := m.header.WithCursorPos(m.list.CursorPosition()).View()

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

	status := m.keyhints.View()
	if m.loading.IsActive() {
		status = m.loading.View()
	}

	return m.layout.Render(header, list, paneView, status)
}

// SetEntries loads entries synchronously (used for stdin).
func (m Model) SetEntries(entries []logsource.Entry) Model {
	m.entries = entries
	m.cachedVisibleCount = len(entries)
	m.list = m.list.SetEntries(entries)
	m.header = m.header.WithCounts(len(entries), len(entries))
	return m
}

func (m Model) openPane(entry logsource.Entry) Model {
	m.pane = m.pane.Open(entry)
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
	m.list, _ = m.list.Update(tea.WindowSizeMsg{
		Width:  l.ListContentWidth(),
		Height: l.EntryListHeight(),
	})
	// T-123 (F-013, F-014): recompute pane vertical allocation on every
	// relayout (open, ratio change, orientation flip). In right-split that
	// means the full main-area slot; in below-mode it stays height_ratio.
	m.pane = m.pane.
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
