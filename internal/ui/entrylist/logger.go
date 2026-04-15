package entrylist

import "strings"

// AbbreviateLogger shortens a dot-separated logger name by keeping the last
// `depth` segments at full length and abbreviating all earlier segments to
// their first character. If the name has depth or fewer segments, it is
// returned unchanged.
//
// Example with depth=2:
//
//	"org.springframework.data.repository.RepositoryDelegate"
//	→ "o.s.d.repository.RepositoryDelegate"
//
// Example with depth=1:
//
//	"com.example.server.AppServerKt"
//	→ "c.e.s.AppServerKt"
func AbbreviateLogger(name string, depth int) string {
	if depth < 1 {
		depth = 1
	}
	parts := strings.Split(name, ".")
	if len(parts) <= depth {
		return name
	}
	cutoff := len(parts) - depth
	result := make([]string, len(parts))
	for i, p := range parts {
		if i < cutoff && len(p) > 0 {
			result[i] = p[:1]
		} else {
			result[i] = p
		}
	}
	return strings.Join(result, ".")
}
