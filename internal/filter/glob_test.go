package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// V35: glob translation table.

func TestGlobToRegexp_Star_MatchesAny(t *testing.T) {
	re, err := globToRegexp("foo*bar")
	require.NoError(t, err)
	assert.True(t, re.MatchString("fooXXXbar"))
	assert.True(t, re.MatchString("foobar"))
	assert.False(t, re.MatchString("foXbar"))
}

func TestGlobToRegexp_Question_MatchesOne(t *testing.T) {
	re, err := globToRegexp("ERR?R")
	require.NoError(t, err)
	assert.True(t, re.MatchString("ERROR"), "ERR?R should match ERROR (O satisfies ?)")
	assert.True(t, re.MatchString("ERRAR"), "ERR?R should match ERRAR (A satisfies ?)")
	assert.False(t, re.MatchString("ERRR"), "ERR?R needs exactly one char after ERR before R")
}

func TestGlobToRegexp_Dot_EscapedLiteral(t *testing.T) {
	re, err := globToRegexp("com.example")
	require.NoError(t, err)
	assert.True(t, re.MatchString("com.example.service"))
	assert.False(t, re.MatchString("comXexample"), "dot must match literal dot only")
}

func TestGlobToRegexp_ParenBracketPipe_EscapedLiteral(t *testing.T) {
	re, err := globToRegexp("a(b|c)")
	require.NoError(t, err)
	assert.True(t, re.MatchString("a(b|c)"))
	assert.False(t, re.MatchString("ab"), "parens and pipe must be literal, not alternation")
}

func TestGlobToRegexp_Backslash_EscapedLiteral(t *testing.T) {
	re, err := globToRegexp(`C:\Users`)
	require.NoError(t, err)
	assert.True(t, re.MatchString(`C:\Users\test`))
}

func TestGlobToRegexp_EmptyPattern_MatchesEverything(t *testing.T) {
	re, err := globToRegexp("")
	require.NoError(t, err)
	assert.True(t, re.MatchString("anything"))
	assert.True(t, re.MatchString(""))
}

func TestGlobToRegexp_DoubleStar_MatchesAny(t *testing.T) {
	re, err := globToRegexp("**")
	require.NoError(t, err)
	assert.True(t, re.MatchString("any string at all"))
}

func TestGlobToRegexp_SubstringBehavior(t *testing.T) {
	// No anchoring — matches anywhere in value.
	re, err := globToRegexp("ERROR")
	require.NoError(t, err)
	assert.True(t, re.MatchString("level:ERROR:details"))
}

func TestGlobToRegexp_Cached(t *testing.T) {
	// Two calls with same pattern return same pointer (cache hit).
	re1, err1 := globToRegexp("foo*")
	re2, err2 := globToRegexp("foo*")
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Same(t, re1, re2, "cached entry must return same *regexp.Regexp pointer")
}
