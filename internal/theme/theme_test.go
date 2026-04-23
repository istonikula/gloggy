package theme

import (
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTheme_AllBuiltins(t *testing.T) {
	for _, name := range BuiltinNames() {
		t.Run(name, func(t *testing.T) {
			th := GetTheme(name)
			assert.Equal(t, name, th.Name)
			assert.NotEmpty(t, string(th.LevelError), "level colors not populated")
			assert.NotEmpty(t, string(th.LevelWarn), "level colors not populated")
			assert.NotEmpty(t, string(th.LevelInfo), "level colors not populated")
			assert.NotEmpty(t, string(th.LevelDebug), "level colors not populated")
			assert.NotEmpty(t, string(th.SyntaxKey), "syntax colors not populated")
			assert.NotEmpty(t, string(th.SyntaxString), "syntax colors not populated")
			assert.NotEmpty(t, string(th.Mark), "UI colors not populated")
			assert.NotEmpty(t, string(th.Dim), "UI colors not populated")
			assert.NotEmpty(t, string(th.SearchHighlight), "UI colors not populated")
			assert.NotEmpty(t, string(th.CursorHighlight), "visual-polish tokens not populated")
			assert.NotEmpty(t, string(th.HeaderBg), "visual-polish tokens not populated")
			assert.NotEmpty(t, string(th.FocusBorder), "visual-polish tokens not populated")
			assert.NotEmpty(t, string(th.DividerColor), "Tier 9 pane-state tokens (DividerColor) not populated")
			assert.NotEmpty(t, string(th.UnfocusedBg), "Tier 9 pane-state tokens (UnfocusedBg) not populated")
			assert.NotEqual(t, string(th.Dim), string(th.DividerColor),
				"DividerColor must be distinct from Dim; got %s", th.DividerColor)
			assert.NotEqual(t, string(th.FocusBorder), string(th.DividerColor),
				"DividerColor must be distinct from FocusBorder; got %s", th.DividerColor)
			assert.NotEqual(t, string(th.Dim), string(th.UnfocusedBg),
				"UnfocusedBg must be distinct from Dim; got %s", th.UnfocusedBg)
			assert.NotEqual(t, string(th.FocusBorder), string(th.UnfocusedBg),
				"UnfocusedBg must be distinct from FocusBorder; got %s", th.UnfocusedBg)
			// T-171: DragHandle populated and distinct from DividerColor + FocusBorder
			// (config R4 AC 9 + AC 10).
			assert.NotEmpty(t, string(th.DragHandle), "Tier 23 DragHandle token not populated")
			assert.NotEqual(t, string(th.DividerColor), string(th.DragHandle),
				"DragHandle must be distinct from DividerColor; got %s", th.DragHandle)
			assert.NotEqual(t, string(th.FocusBorder), string(th.DragHandle),
				"DragHandle must be distinct from FocusBorder; got %s", th.DragHandle)
			// T-176: BaseBg populated and != UnfocusedBg
			// (config R4 AC 12 + AC 15).
			assert.NotEmpty(t, string(th.BaseBg), "Tier 24 BaseBg token not populated")
			assert.NotEqual(t, string(th.UnfocusedBg), string(th.BaseBg),
				"BaseBg must be distinct from UnfocusedBg; got %s", th.BaseBg)
		})
	}
}

// T-177 (cavekit-config.md R4 AC 16): each theme constructor cites its
// canonical upstream source via a named exported constant, discoverable
// at test time. Each citation must contain an `https://` URL and a
// canonical-identifier keyword (variant name, flavor, or legacy-palette
// label). This is a code-review drift tripwire — if a contributor edits
// a palette hex but forgets to update the citation, the source/value
// pair is still locally correct; if they edit the citation but forget
// the palette, the test still holds. The real drift guard is that every
// hex now routes through a named palette field, so moving to a different
// variant becomes a visible field-rename, not a silent hex edit.
func TestTheme_CanonicalSourceCitations_Discoverable(t *testing.T) {
	cases := []struct {
		theme       string
		source      string
		keywordsAny []string
	}{
		{"tokyo-night", TokyoNightSource, []string{"night"}},
		{"catppuccin-mocha", CatppuccinMochaSource, []string{"mocha"}},
		{"material-dark", MaterialDarkSource, []string{"Astorino", "material"}},
	}
	for _, tc := range cases {
		t.Run(tc.theme, func(t *testing.T) {
			require.NotEmpty(t, tc.source, "citation constant empty")
			assert.Contains(t, tc.source, "https://", "citation missing https:// URL: %q", tc.source)
			matched := false
			for _, kw := range tc.keywordsAny {
				if strings.Contains(strings.ToLower(tc.source),
					strings.ToLower(kw)) {
					matched = true
					break
				}
			}
			assert.True(t, matched, "citation missing canonical keyword %v: %q",
				tc.keywordsAny, tc.source)
			assert.Equal(t, tc.source, Source(tc.theme),
				"Source(%q)", tc.theme)
		})
	}

	assert.Empty(t, Source("nonexistent"), "Source(unknown) should be empty")
}

// T-178 (cavekit-config.md R4 AC 17): catppuccin-mocha's FocusBorder
// is the upstream Lavender accent (#b4befe), not Blue (#89b4fa). All
// official catppuccin ports (neovim, lazygit, btop) use Lavender for
// active borders — matching the canonical identity.
func TestCatppuccinMocha_FocusBorder_IsLavender(t *testing.T) {
	th := GetTheme("catppuccin-mocha")
	const want = "#b4befe"
	assert.Equal(t, want, string(th.FocusBorder),
		"catppuccin-mocha FocusBorder should be upstream Lavender")
}

// T-176 (cavekit-config.md R4 AC 14): BaseBg values must be pairwise
// distinct across the three bundled themes so the "at a glance they are
// perceptibly different" property has objective grounding in the palette
// data, not just in rendered luminance.
func TestBaseBg_PairwiseDistinct_AllThemes(t *testing.T) {
	names := BuiltinNames()
	seen := make(map[string]string, len(names))
	for _, name := range names {
		th := GetTheme(name)
		bg := string(th.BaseBg)
		require.NotEmpty(t, bg, "%s: BaseBg empty", name)
		other, dup := seen[bg]
		assert.False(t, dup, "BaseBg collision: %s and %s both use %s", other, name, bg)
		seen[bg] = name
	}
}

// T-175 (cavekit-config.md R4 AC 11, cavekit-app-shell.md R15 AC 16):
// DragHandle must read as a mid-tone neutral — clearly brighter than
// DividerColor, dimmer than FocusBorder. The tui-mcp HUMAN sign-off
// harness was unavailable during this tier (posix_spawnp spawn failure at
// harness level — not a gloggy-side defect, analogous to F-124). We pin
// the objective luminance-ordering invariant that underlies the perceptual
// AC as a regression test: WCAG relative luminance must strictly increase
// DividerColor → DragHandle → FocusBorder, and both gaps must exceed a
// perceptual threshold (0.02 on the 0..1 WCAG scale, roughly equivalent
// to a single-step L* difference).
func TestDragHandle_LuminanceOrdering_AllThemes(t *testing.T) {
	const minGap = 0.02
	for _, name := range BuiltinNames() {
		t.Run(name, func(t *testing.T) {
			th := GetTheme(name)
			ld := wcagLuminance(t, string(th.DividerColor))
			ldh := wcagLuminance(t, string(th.DragHandle))
			lf := wcagLuminance(t, string(th.FocusBorder))
			assert.True(t, ld < ldh && ldh < lf,
				"luminance not ordered: Divider=%.4f DragHandle=%.4f Focus=%.4f", ld, ldh, lf)
			assert.GreaterOrEqual(t, ldh-ld, minGap,
				"Divider→DragHandle gap below perceptual threshold %.2f", minGap)
			assert.GreaterOrEqual(t, lf-ldh, minGap,
				"DragHandle→Focus gap below perceptual threshold %.2f", minGap)
		})
	}
}

// wcagLuminance returns WCAG 2.x relative luminance (0..1) for an sRGB
// hex color like "#5a6475". Matches the formula used in T-175's one-off
// verification — see loop-log.md Iteration 45.
func wcagLuminance(t *testing.T, hex string) float64 {
	t.Helper()
	s := strings.TrimPrefix(hex, "#")
	require.Len(t, s, 6, "bad hex color: %q", hex)
	ch := func(i int) float64 {
		v, err := strconv.ParseInt(s[i:i+2], 16, 32)
		require.NoError(t, err, "parse %q", s[i:i+2])
		f := float64(v) / 255.0
		if f <= 0.03928 {
			return f / 12.92
		}
		return math.Pow((f+0.055)/1.055, 2.4)
	}
	return 0.2126*ch(0) + 0.7152*ch(2) + 0.0722*ch(4)
}

func TestGetTheme_UnknownFallsBackToDefault(t *testing.T) {
	th := GetTheme("nonexistent")
	assert.Equal(t, DefaultThemeName, th.Name,
		"unknown theme should fallback to %s", DefaultThemeName)
}

func TestBuiltinNames(t *testing.T) {
	names := BuiltinNames()
	require.Len(t, names, 3, "want 3 built-in themes")
}

func TestDefaultThemeName(t *testing.T) {
	assert.Equal(t, "tokyo-night", DefaultThemeName)
}

// V30: NextName advances through BuiltinNames in declaration order and
// wraps at the end. An unknown name returns the first bundled theme, so
// a stale config value cannot strand the cycle.
func TestNextName_CyclesInDeclarationOrder(t *testing.T) {
	names := BuiltinNames()
	require.Len(t, names, 3, "expected 3 bundled themes")
	assert.Equal(t, names[1], NextName(names[0]))
	assert.Equal(t, names[2], NextName(names[1]))
	assert.Equal(t, names[0], NextName(names[2]), "last should wrap to first")
}

func TestNextName_UnknownReturnsFirst(t *testing.T) {
	names := BuiltinNames()
	assert.Equal(t, names[0], NextName("nonexistent"))
	assert.Equal(t, names[0], NextName(""))
}
