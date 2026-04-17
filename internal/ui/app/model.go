// Package app is the top-level Bubble Tea model that wires all subsystems together.
package app

import (
	"context"
	"strings"

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
		resize:      appshell.NewResizeModel(80, 24),
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
	m.layout = appshell.NewLayoutModel(80, 24)

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
		l := appshell.ApplyToLayout(m.resize, m.paneHeight.Ratio(), m.pane.IsOpen())
		m.layout = m.layout.SetSize(w, h).SetDetailPane(m.pane.IsOpen(), m.paneHeight.PaneHeight())
		m.list, _ = m.list.Update(tea.WindowSizeMsg{Width: w, Height: l.EntryListHeight()})
		m.pane = m.pane.SetHeight(m.paneHeight.PaneHeight())
		m.header = m.header.WithWidth(w)
		m.keyhints = m.keyhints.WithWidth(w)
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
		// +/- resize pane.
		if msg.String() == "+" || msg.String() == "-" {
			m.paneHeight, _ = m.paneHeight.Update(msg)
			m.pane = m.pane.SetHeight(m.paneHeight.PaneHeight())
			m = m.relayout()
			return m, nil
		}
		// In-pane search.
		if m.paneSearch.IsActive() || msg.String() == "/" {
			// Get content lines for match computation.
			lines := strings.Split(m.pane.View(), "\n")
			m.paneSearch, _ = m.paneSearch.Update(msg, lines)
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

// View composes the full screen.
func (m Model) View() string {
	if m.help.IsOpen() {
		return m.help.View()
	}

	header := m.header.WithCursorPos(m.list.CursorPosition()).View()
	list := m.list.View()

	paneView := ""
	if m.pane.IsOpen() {
		// T-083: Set focus state before rendering (detail pane shows its own border).
		m.pane.Focused = (m.focus == appshell.FocusDetailPane)
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
	m.focus = appshell.FocusDetailPane
	m.keyhints = m.keyhints.WithFocus(appshell.FocusDetailPane)
	m = m.relayout()
	return m
}

func (m Model) relayout() Model {
	l := appshell.ApplyToLayout(m.resize, m.paneHeight.Ratio(), m.pane.IsOpen())
	m.layout = m.layout.SetDetailPane(m.pane.IsOpen(), m.paneHeight.PaneHeight())
	m.list, _ = m.list.Update(tea.WindowSizeMsg{
		Width:  m.resize.Width(),
		Height: l.EntryListHeight(),
	})
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
