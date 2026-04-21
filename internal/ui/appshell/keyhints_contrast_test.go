package appshell

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/istonikula/gloggy/internal/theme"
)

// sgrRe extracts the parameter portion of a CSI SGR escape: the digits/
// semicolons between `\x1b[` and `m`.
var sgrRe = regexp.MustCompile(`\x1b\[([0-9;]*)m`)

// sgrParams returns the set of SGR parameter strings that appear in s.
// Reset sequences (empty or `0`) are dropped — they carry no style
// information and would otherwise make the notice / hints outputs look
// identical.
func sgrParams(s string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, m := range sgrRe.FindAllStringSubmatch(s, -1) {
		p := m[1]
		if p == "" || p == "0" {
			continue
		}
		out[p] = struct{}{}
	}
	return out
}

// TestKeyHintBar_NoticeSgrDiffersFromHints_AllThemes locks in V15/V25
// class-(b) coverage: the SGR attributes KeyHintBarModel.View() emits
// for a transient notice MUST differ from the attributes emitted for the
// keyhints row, on every bundled theme. B1 regressed on tokyo-night
// because both branches shared `Foreground(m.th.Dim)` and the notice
// text, while present, blended into the keyhints row.
func TestKeyHintBar_NoticeSgrDiffersFromHints_AllThemes(t *testing.T) {
	themes := []string{"tokyo-night", "catppuccin-mocha", "material-dark"}
	const width = 80

	for _, name := range themes {
		t.Run(name, func(t *testing.T) {
			th := theme.GetTheme(name)
			base := NewKeyHintBarModel(th, width)

			hintsView := base.View()
			noticeView := base.WithNotice("no marked entries").View()

			hintsAttrs := sgrParams(hintsView)
			noticeAttrs := sgrParams(noticeView)

			require.NotEmptyf(t, noticeAttrs,
				"notice view emitted no SGR attributes — test env may be stripping styles; hintsView=%q noticeView=%q",
				hintsView, noticeView)

			assert.Falsef(t, equalStringSet(hintsAttrs, noticeAttrs),
				"notice SGR attrs == hints SGR attrs for theme %q — V25 class-(b) violation (notice would blend into keyhints row).\n  hints:  %v\n  notice: %v",
				name, sortedKeys(hintsAttrs), sortedKeys(noticeAttrs))

			// Sanity: the notice text itself must appear in the rendered
			// output. Without this, a degenerate stub View() that emitted
			// only SGR codes would pass the distinctness check.
			assert.Containsf(t, noticeView, "no marked entries",
				"notice text missing from notice view: %q", noticeView)
		})
	}
}

func equalStringSet(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

func sortedKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	// simple insertion sort — tiny sets
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j-1] > out[j]; j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}
