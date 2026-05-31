// Package filter provides log entry filtering with include/exclude rules.
package filter

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Syntax controls how a filter's Pattern is interpreted.
type Syntax int

const (
	Literal Syntax = iota // substring match (strings.Contains)
	Glob                  // * = any chars, ? = one char; translated to RE2 internally
	Regex                 // raw RE2 pattern
)

func (s Syntax) String() string {
	switch s {
	case Literal:
		return "literal"
	case Glob:
		return "glob"
	case Regex:
		return "regex"
	default:
		return fmt.Sprintf("Syntax(%d)", int(s))
	}
}

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
	Syntax  Syntax // zero-value = Literal (substring match)
	Mode    Mode
	Enabled bool
}

// FilterSet holds multiple filters with add/remove/toggle operations.
type FilterSet struct {
	filters          []Filter
	ids              []int
	nextID           int
	globallyDisabled bool
	savedEnabled     map[int]bool // saved Enabled per filter ID at time of global disable
}

// NewFilterSet creates an empty FilterSet with all internal maps initialized,
// so every public mutation is safe from a fresh ctor regardless of call order
// (V40 — no implicit "ToggleAll-ran-first" precondition).
func NewFilterSet() *FilterSet {
	return &FilterSet{savedEnabled: make(map[int]bool)}
}

// Add appends a filter and returns its ID.
func (fs *FilterSet) Add(f Filter) int {
	id := fs.nextID
	fs.nextID++
	// If globally disabled, save the new filter's state and disable it.
	if fs.globallyDisabled {
		if fs.savedEnabled == nil { // V40 defense: never write a nil map
			fs.savedEnabled = make(map[int]bool)
		}
		fs.savedEnabled[id] = f.Enabled
		f.Enabled = false
	}
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
			// Clean up saved state if tracking.
			delete(fs.savedEnabled, id)
			return true
		}
	}
	return false
}

// Enable sets a filter's Enabled flag to true.
//
// V32 MANUAL-TOGGLE EXIT: if invoked while globallyDisabled=true, the
// global-toggle state machine is exited — globallyDisabled is cleared
// and savedEnabled is dropped. Without this, a subsequent `F` press
// would hit the restore branch and silently clobber this manual
// toggle while re-enabling filters the user thought they had cleared
// (B11).
func (fs *FilterSet) Enable(id int) bool {
	for i, fid := range fs.ids {
		if fid == id {
			fs.filters[i].Enabled = true
			fs.exitGloballyDisabled()
			return true
		}
	}
	return false
}

// Disable sets a filter's Enabled flag to false. V32 MANUAL-TOGGLE
// EXIT: see Enable.
func (fs *FilterSet) Disable(id int) bool {
	for i, fid := range fs.ids {
		if fid == id {
			fs.filters[i].Enabled = false
			fs.exitGloballyDisabled()
			return true
		}
	}
	return false
}

// exitGloballyDisabled clears the global-toggle state machine so a
// subsequent ToggleAll call starts a fresh 1st-press cycle from the
// current per-filter Enabled values.
func (fs *FilterSet) exitGloballyDisabled() {
	if fs.globallyDisabled {
		fs.globallyDisabled = false
		fs.savedEnabled = nil
	}
}

// Validate checks that f is a valid filter before Add/Edit commit:
// blank Pattern is rejected; Regex syntax with a bad pattern is rejected.
// Glob translation is deterministic and never fails (no validation needed).
// Returns a descriptive error to surface to the user via V15-pattern notice.
//
// V35 AMEND / B36: reject whitespace-only Pattern (TrimSpace), not just the
// empty string — a Pattern of spaces builds a filter that matches/excludes
// ~everything, indistinguishable from a working filter. The stored Pattern is
// NOT trimmed (leading/trailing spaces can be a deliberate literal); only the
// all-whitespace case is rejected.
func (fs *FilterSet) Validate(f Filter) error {
	if strings.TrimSpace(f.Pattern) == "" {
		return errors.New("pattern required")
	}
	if f.Syntax == Regex {
		if _, err := regexp.Compile(f.Pattern); err != nil {
			return fmt.Errorf("bad regex: %w", err)
		}
	}
	return nil
}

// Update replaces the filter with the given id in-place, preserving slice
// position and Enabled state (edit never flips enable/disable — Space does).
// Returns false if id is not found.
func (fs *FilterSet) Update(id int, f Filter) bool {
	for i, fid := range fs.ids {
		if fid == id {
			enabled := fs.filters[i].Enabled
			fs.filters[i] = f
			fs.filters[i].Enabled = enabled
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

// IsGloballyDisabled reports whether ToggleAll has muted all filters.
func (fs *FilterSet) IsGloballyDisabled() bool { return fs.globallyDisabled }

// ToggleAll disables all filters globally on the first call, then re-enables
// only the ones that were individually enabled before on the second call.
// Filters that were individually disabled before the first call remain disabled
// after the second call.
func (fs *FilterSet) ToggleAll() {
	if !fs.globallyDisabled {
		// Save enabled state by ID and disable all.
		fs.savedEnabled = make(map[int]bool, len(fs.filters))
		for i, id := range fs.ids {
			fs.savedEnabled[id] = fs.filters[i].Enabled
			fs.filters[i].Enabled = false
		}
		fs.globallyDisabled = true
	} else {
		// Restore saved enabled states by ID.
		for i, id := range fs.ids {
			if saved, ok := fs.savedEnabled[id]; ok {
				fs.filters[i].Enabled = saved
			} else {
				// Filter added while disabled without saved state; default to enabled.
				fs.filters[i].Enabled = true
			}
		}
		fs.savedEnabled = nil
		fs.globallyDisabled = false
	}
}
