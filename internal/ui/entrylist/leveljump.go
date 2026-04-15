package entrylist

import (
	"strings"

	"github.com/istonikula/gloggy/internal/logsource"
)

// WrapDirection indicates which direction wrapped during a level-jump.
type WrapDirection int

const (
	NoWrap    WrapDirection = iota
	WrapForward              // wrapped from end to beginning
	WrapBack                 // wrapped from beginning to end
)

// NextLevel returns the index of the next entry (after current) whose Level
// matches targetLevel. Searches the full entries slice (not just filtered).
// Returns (newIndex, direction). If no matches exist, returns (current, NoWrap).
func NextLevel(entries []logsource.Entry, current int, targetLevel string) (int, WrapDirection) {
	n := len(entries)
	if n == 0 {
		return current, NoWrap
	}
	target := strings.ToUpper(targetLevel)
	for i := 1; i <= n; i++ {
		idx := (current + i) % n
		if strings.ToUpper(entries[idx].Level) == target {
			if idx <= current {
				return idx, WrapForward
			}
			return idx, NoWrap
		}
	}
	return current, NoWrap
}

// PrevLevel returns the index of the previous entry (before current) whose Level
// matches targetLevel. Searches the full entries slice.
func PrevLevel(entries []logsource.Entry, current int, targetLevel string) (int, WrapDirection) {
	n := len(entries)
	if n == 0 {
		return current, NoWrap
	}
	target := strings.ToUpper(targetLevel)
	for i := 1; i <= n; i++ {
		idx := (current - i + n) % n
		if strings.ToUpper(entries[idx].Level) == target {
			if idx >= current {
				return idx, WrapBack
			}
			return idx, NoWrap
		}
	}
	return current, NoWrap
}
