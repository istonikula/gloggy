package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// V35 AMEND / B36: Validate rejects a whitespace-only Pattern (not just the
// empty string) — an all-spaces Pattern builds a filter that matches/excludes
// ~everything. A Pattern containing internal or literal surrounding spaces is
// still valid (only the all-whitespace case is rejected; the stored Pattern is
// not trimmed).
func TestValidate_WhitespaceOnlyPattern_Rejected_B36(t *testing.T) {
	fs := NewFilterSet()

	for _, p := range []string{"", " ", "   ", "\t", " \n\t "} {
		err := fs.Validate(Filter{Field: "msg", Pattern: p, Syntax: Literal})
		require.Errorf(t, err, "whitespace-only pattern %q must be rejected", p)
		assert.Equalf(t, "pattern required", err.Error(), "for pattern %q", p)
	}

	for _, p := range []string{"a", "a b", " x "} {
		assert.NoErrorf(t, fs.Validate(Filter{Field: "msg", Pattern: p, Syntax: Literal}),
			"pattern %q has non-whitespace content and must pass", p)
	}
}
