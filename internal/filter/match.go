package filter

import (
	"encoding/json"
	"regexp"
	"strings"
	"sync"

	"github.com/istonikula/gloggy/internal/logsource"
)

// regexCache caches compiled regexes to avoid recompilation on every Match call.
var regexCache sync.Map // map[string]*regexp.Regexp

func cachedRegexp(pattern string) (*regexp.Regexp, error) {
	if v, ok := regexCache.Load(pattern); ok {
		return v.(*regexp.Regexp), nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	regexCache.Store(pattern, re)
	return re, nil
}

// Match reports whether the filter matches the entry.
//
// When Field is empty (""), the match target is the raw JSON bytes of the
// entry (whole-line filter). When Field is non-empty, the named field value
// is resolved via entryFieldValue; a missing field returns (false, nil).
//
// The Pattern is evaluated according to f.Syntax:
//   - Literal (default): strings.Contains substring match.
//   - Glob: * = any chars, ? = one char; remaining chars literal (RE2 internally).
//   - Regex: RE2 pattern (pre-validated at Add/Edit time via FilterSet.Validate).
func Match(f Filter, entry logsource.Entry) (bool, error) {
	var val string
	if f.Field == "" {
		val = string(entry.Raw)
	} else {
		var found bool
		val, found = entryFieldValue(f.Field, entry)
		if !found {
			return false, nil
		}
	}
	return matchBySyntax(f.Syntax, f.Pattern, val)
}

func matchBySyntax(syntax Syntax, pattern, value string) (bool, error) {
	switch syntax {
	case Glob:
		re, err := globToRegexp(pattern)
		if err != nil {
			return false, err
		}
		return re.MatchString(value), nil
	case Regex:
		re, err := cachedRegexp(pattern)
		if err != nil {
			return false, err
		}
		return re.MatchString(value), nil
	default: // Literal
		return strings.Contains(value, pattern), nil
	}
}

// entryFieldValue returns the string value of the named field from an Entry.
// Returns ("", false) when the field does not exist.
func entryFieldValue(field string, entry logsource.Entry) (string, bool) {
	switch strings.ToLower(field) {
	case "msg":
		return entry.Msg, true
	case "level":
		return entry.Level, true
	case "logger":
		return entry.Logger, true
	case "thread":
		return entry.Thread, true
	default:
		// Guard nil Extra map to avoid panic on map lookup.
		if entry.Extra == nil {
			return "", false
		}
		raw, ok := entry.Extra[field]
		if !ok {
			return "", false
		}
		s := string(raw)
		// Properly unquote JSON strings to handle escape sequences.
		if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
			var unquoted string
			if err := json.Unmarshal(raw, &unquoted); err == nil {
				s = unquoted
			} else {
				s = s[1 : len(s)-1]
			}
		}
		return s, true
	}
}
