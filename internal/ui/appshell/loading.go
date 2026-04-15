package appshell

import "fmt"

// LoadingModel tracks background load state and renders a loading indicator.
// When loading is active, the View() returns a progress string.
// When done, View() returns an empty string.
type LoadingModel struct {
	active bool
	count  int // entries loaded so far
}

// NewLoadingModel creates a LoadingModel in the not-loading state.
func NewLoadingModel() LoadingModel { return LoadingModel{} }

// Start activates the loading indicator.
func (m LoadingModel) Start() LoadingModel {
	m.active = true
	m.count = 0
	return m
}

// Update sets the current loaded count.
func (m LoadingModel) Update(count int) LoadingModel {
	m.count = count
	return m
}

// Done deactivates the indicator.
func (m LoadingModel) Done() LoadingModel {
	m.active = false
	return m
}

// IsActive returns true while loading is in progress.
func (m LoadingModel) IsActive() bool { return m.active }

// View returns a loading status string, or empty string when not loading.
func (m LoadingModel) View() string {
	if !m.active {
		return ""
	}
	return fmt.Sprintf("Loading... (%d entries)", m.count)
}
