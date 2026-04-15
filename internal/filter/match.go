package filter

import (
	"regexp"
	"strings"

	"github.com/istonikula/gloggy/internal/logsource"
)

// Match reports whether the filter's pattern matches the entry's field value.
// Literal patterns use substring matching; patterns containing regex metacharacters
// use RE2 matching. Returns an error if the pattern is invalid regex.
func Match(f Filter, entry logsource.Entry) (bool, error) {
	val, found := entryFieldValue(f.Field, entry)
	if !found {
		return false, nil
	}
	return matchPattern(f.Pattern, val)
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
		raw, ok := entry.Extra[field]
		if !ok {
			return "", false
		}
		s := string(raw)
		// Strip JSON string quotes for string values.
		if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
			s = s[1 : len(s)-1]
		}
		return s, true
	}
}

// matchPattern tests the pattern against value. Uses regex if the pattern
// contains metacharacters, otherwise uses substring match.
func matchPattern(pattern, value string) (bool, error) {
	if containsMetaChar(pattern) {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false, err
		}
		return re.MatchString(value), nil
	}
	return strings.Contains(value, pattern), nil
}

func containsMetaChar(s string) bool {
	return strings.ContainsAny(s, `.*+?^$|[](){}\ `)
}
