package filter

import (
	"regexp"
	"strings"
	"sync"
)

var globCache sync.Map // map[string]*regexp.Regexp

// globToRegexp compiles a glob pattern to a RE2 regexp (cached).
// Glob rules: * = any sequence of characters; ? = exactly one character.
// All other RE2 metacharacters are escaped so they match literally.
// No anchoring — matches anywhere in the value (substring-feeling preserved).
func globToRegexp(pattern string) (*regexp.Regexp, error) {
	if v, ok := globCache.Load(pattern); ok {
		return v.(*regexp.Regexp), nil
	}
	re, err := compileGlob(pattern)
	if err != nil {
		return nil, err
	}
	globCache.Store(pattern, re)
	return re, nil
}

// compileGlob translates a glob pattern to a RE2 string and compiles it.
// Translation: escape RE2 metachars except * and ?, then substitute
// * → .* and ? → .
func compileGlob(pattern string) (*regexp.Regexp, error) {
	var sb strings.Builder
	for i := 0; i < len(pattern); i++ {
		ch := pattern[i]
		switch ch {
		case '*':
			sb.WriteString(".*")
		case '?':
			sb.WriteByte('.')
		default:
			if isGlobEscapedMeta(ch) {
				sb.WriteByte('\\')
			}
			sb.WriteByte(ch)
		}
	}
	return regexp.Compile(sb.String())
}

// isGlobEscapedMeta reports whether ch is a RE2 metacharacter that must be
// escaped when appearing as a literal in a glob pattern.
// * and ? are excluded because they are handled as glob wildcards above.
func isGlobEscapedMeta(ch byte) bool {
	return strings.IndexByte(`\.+^${}[]()|`, ch) >= 0
}
