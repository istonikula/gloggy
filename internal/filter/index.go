package filter

import "github.com/istonikula/gloggy/internal/logsource"

// Apply returns indices (into entries) of entries that pass the active
// include/exclude logic of fs. Original order is preserved.
//
// Rules:
//   - If any enabled include filters exist, an entry must match at least one.
//   - Any enabled exclude filter match removes the entry regardless.
//   - Disabled filters are ignored entirely.
func Apply(fs *FilterSet, entries []logsource.Entry) []int {
	var includes, excludes []Filter
	for _, f := range fs.GetAll() {
		if !f.Enabled {
			continue
		}
		switch f.Mode {
		case Include:
			includes = append(includes, f)
		case Exclude:
			excludes = append(excludes, f)
		}
	}

	var result []int
	for i, e := range entries {
		if matchAny(excludes, e) {
			continue
		}
		if len(includes) > 0 && !matchAny(includes, e) {
			continue
		}
		result = append(result, i)
	}
	return result
}

func matchAny(filters []Filter, e logsource.Entry) bool {
	for _, f := range filters {
		if ok, _ := Match(f, e); ok {
			return true
		}
	}
	return false
}

// FilteredIndex holds the indices of entries that currently pass the
// active filter set.
type FilteredIndex struct {
	Indices []int
}

// NewFilteredIndex computes the initial index.
func NewFilteredIndex(fs *FilterSet, entries []logsource.Entry) *FilteredIndex {
	return &FilteredIndex{Indices: Apply(fs, entries)}
}

// Recompute replaces Indices with a fresh Apply result.
func (fi *FilteredIndex) Recompute(fs *FilterSet, entries []logsource.Entry) {
	fi.Indices = Apply(fs, entries)
}
