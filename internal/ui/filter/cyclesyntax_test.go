package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/istonikula/gloggy/internal/filter"
)

// B42: the Syntax cycle helpers derive their modulo from filter.SyntaxCount,
// not a hardcoded `% 3`, so adding a Syntax value cannot silently break
// wraparound. Verify full forward + backward cycles over every value.
func TestCycleSyntax_WrapsOverAllValues_B42(t *testing.T) {
	// Forward: Literal → Glob → Regex → Literal.
	s := filter.Literal
	seen := []filter.Syntax{s}
	for i := 0; i < filter.SyntaxCount; i++ {
		s = cycleSyntaxNext(s)
		seen = append(seen, s)
	}
	assert.Equal(t,
		[]filter.Syntax{filter.Literal, filter.Glob, filter.Regex, filter.Literal},
		seen, "forward cycle must visit each value then wrap")

	// Backward is the exact inverse of forward for every value.
	for v := 0; v < filter.SyntaxCount; v++ {
		sv := filter.Syntax(v)
		assert.Equalf(t, sv, cycleSyntaxPrev(cycleSyntaxNext(sv)),
			"prev∘next must be identity for %v", sv)
		assert.Equalf(t, sv, cycleSyntaxNext(cycleSyntaxPrev(sv)),
			"next∘prev must be identity for %v", sv)
	}

	// Prev wraps Literal → Regex (the last value), not a negative index.
	assert.Equal(t, filter.Regex, cycleSyntaxPrev(filter.Literal),
		"prev from the first value wraps to the last")
}
