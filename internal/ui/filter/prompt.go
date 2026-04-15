package filter

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/istonikula/gloggy/internal/filter"
)

// FilterConfirmedMsg is emitted when the user confirms a filter prompt.
// The filter has been added to the FilterSet and the index should be recomputed.
type FilterConfirmedMsg struct {
	FilterID int
	FilterSet *filter.FilterSet
}

// FilterCancelledMsg is emitted when the user cancels the filter prompt.
type FilterCancelledMsg struct{}

// PromptModel is a transient dialog that lets the user confirm an
// add-filter action with pre-filled field/pattern, choosing include or exclude.
//
// Navigation:
//   Tab / Shift+Tab  cycle mode (Include <-> Exclude)
//   Enter            confirm and emit FilterConfirmedMsg
//   Esc              cancel and emit FilterCancelledMsg
type PromptModel struct {
	active  bool
	field   string
	pattern string
	mode    filter.Mode  // Include or Exclude
	fs      *filter.FilterSet
}

// NewPromptModel creates a PromptModel that operates on the given FilterSet.
func NewPromptModel(fs *filter.FilterSet) PromptModel {
	return PromptModel{fs: fs}
}

// IsActive returns true when the prompt is showing.
func (m PromptModel) IsActive() bool { return m.active }

// Field returns the pre-filled field name.
func (m PromptModel) Field() string { return m.field }

// Pattern returns the pre-filled pattern value.
func (m PromptModel) Pattern() string { return m.pattern }

// Mode returns the current selected mode.
func (m PromptModel) Mode() filter.Mode { return m.mode }

// Open activates the prompt with pre-filled field and pattern, defaulting to Include.
func (m PromptModel) Open(field, pattern string) PromptModel {
	m.active = true
	m.field = field
	m.pattern = pattern
	m.mode = filter.Include
	return m
}

// Close deactivates the prompt without committing.
func (m PromptModel) Close() PromptModel {
	m.active = false
	return m
}

// Update handles key events when the prompt is active.
func (m PromptModel) Update(msg tea.Msg) (PromptModel, tea.Cmd) {
	if !m.active {
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m = m.Close()
			return m, func() tea.Msg { return FilterCancelledMsg{} }
		case "enter":
			id := m.fs.Add(filter.Filter{
				Field:   m.field,
				Pattern: m.pattern,
				Mode:    m.mode,
				Enabled: true,
			})
			m = m.Close()
			fs := m.fs
			return m, func() tea.Msg { return FilterConfirmedMsg{FilterID: id, FilterSet: fs} }
		case "tab", "shift+tab":
			if m.mode == filter.Include {
				m.mode = filter.Exclude
			} else {
				m.mode = filter.Include
			}
		}
	}
	return m, nil
}

// View renders the prompt as a single-line string.
func (m PromptModel) View() string {
	if !m.active {
		return ""
	}
	modeStr := "include"
	if m.mode == filter.Exclude {
		modeStr = "exclude"
	}
	return "Add filter: " + m.field + ":" + m.pattern + " [" + modeStr + "] (Tab=toggle mode, Enter=confirm, Esc=cancel)"
}
