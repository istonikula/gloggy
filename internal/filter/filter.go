// Package filter provides log entry filtering with include/exclude rules.
package filter

import "fmt"

// Mode represents whether a filter includes or excludes matching entries.
type Mode int

const (
	Include Mode = iota
	Exclude
)

func (m Mode) String() string {
	switch m {
	case Include:
		return "include"
	case Exclude:
		return "exclude"
	default:
		return fmt.Sprintf("Mode(%d)", int(m))
	}
}

// Filter represents a single filter rule.
type Filter struct {
	Field   string
	Pattern string
	Mode    Mode
	Enabled bool
}

// FilterSet holds multiple filters with add/remove/toggle operations.
type FilterSet struct {
	filters          []Filter
	ids              []int
	nextID           int
	globallyDisabled bool
	savedEnabled     []bool // saved Enabled per filter index at time of global disable
}

// NewFilterSet creates an empty FilterSet.
func NewFilterSet() *FilterSet {
	return &FilterSet{}
}

// Add appends a filter and returns its ID.
func (fs *FilterSet) Add(f Filter) int {
	id := fs.nextID
	fs.nextID++
	fs.filters = append(fs.filters, f)
	fs.ids = append(fs.ids, id)
	return id
}

// Remove deletes a filter by ID. Returns false if not found.
func (fs *FilterSet) Remove(id int) bool {
	for i, fid := range fs.ids {
		if fid == id {
			fs.filters = append(fs.filters[:i], fs.filters[i+1:]...)
			fs.ids = append(fs.ids[:i], fs.ids[i+1:]...)
			return true
		}
	}
	return false
}

// Enable sets a filter's Enabled flag to true.
func (fs *FilterSet) Enable(id int) bool {
	for i, fid := range fs.ids {
		if fid == id {
			fs.filters[i].Enabled = true
			return true
		}
	}
	return false
}

// Disable sets a filter's Enabled flag to false.
func (fs *FilterSet) Disable(id int) bool {
	for i, fid := range fs.ids {
		if fid == id {
			fs.filters[i].Enabled = false
			return true
		}
	}
	return false
}

// GetAll returns a copy of all filters including disabled ones.
func (fs *FilterSet) GetAll() []Filter {
	out := make([]Filter, len(fs.filters))
	copy(out, fs.filters)
	return out
}

// GetIDs returns a copy of all filter IDs in their current order.
// The i-th ID corresponds to the i-th filter returned by GetAll().
func (fs *FilterSet) GetIDs() []int {
	out := make([]int, len(fs.ids))
	copy(out, fs.ids)
	return out
}

// GetEnabled returns only enabled filters.
func (fs *FilterSet) GetEnabled() []Filter {
	var out []Filter
	for _, f := range fs.filters {
		if f.Enabled {
			out = append(out, f)
		}
	}
	return out
}

// ToggleAll disables all filters globally on the first call, then re-enables
// only the ones that were individually enabled before on the second call.
// Filters that were individually disabled before the first call remain disabled
// after the second call.
func (fs *FilterSet) ToggleAll() {
	if !fs.globallyDisabled {
		// Save enabled state and disable all.
		fs.savedEnabled = make([]bool, len(fs.filters))
		for i := range fs.filters {
			fs.savedEnabled[i] = fs.filters[i].Enabled
			fs.filters[i].Enabled = false
		}
		fs.globallyDisabled = true
	} else {
		// Restore saved enabled states.
		for i := range fs.filters {
			if i < len(fs.savedEnabled) {
				fs.filters[i].Enabled = fs.savedEnabled[i]
			}
		}
		fs.savedEnabled = nil
		fs.globallyDisabled = false
	}
}
