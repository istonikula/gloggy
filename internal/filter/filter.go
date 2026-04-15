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
	filters []Filter
	ids     []int
	nextID  int
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
