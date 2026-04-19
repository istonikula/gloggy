package theme

import (
	"math"
	"strconv"
	"strings"
	"testing"
)

func TestGetTheme_AllBuiltins(t *testing.T) {
	for _, name := range BuiltinNames() {
		t.Run(name, func(t *testing.T) {
			th := GetTheme(name)
			if th.Name != name {
				t.Errorf("Name = %q, want %q", th.Name, name)
			}
			if string(th.LevelError) == "" || string(th.LevelWarn) == "" ||
				string(th.LevelInfo) == "" || string(th.LevelDebug) == "" {
				t.Error("level colors not populated")
			}
			if string(th.SyntaxKey) == "" || string(th.SyntaxString) == "" {
				t.Error("syntax colors not populated")
			}
			if string(th.Mark) == "" || string(th.Dim) == "" || string(th.SearchHighlight) == "" {
				t.Error("UI colors not populated")
			}
			if string(th.CursorHighlight) == "" || string(th.HeaderBg) == "" || string(th.FocusBorder) == "" {
				t.Error("visual-polish tokens not populated")
			}
			if string(th.DividerColor) == "" || string(th.UnfocusedBg) == "" {
				t.Error("Tier 9 pane-state tokens (DividerColor, UnfocusedBg) not populated")
			}
			if string(th.DividerColor) == string(th.Dim) || string(th.DividerColor) == string(th.FocusBorder) {
				t.Errorf("DividerColor must be distinct from Dim and FocusBorder; got %s", th.DividerColor)
			}
			if string(th.UnfocusedBg) == string(th.Dim) || string(th.UnfocusedBg) == string(th.FocusBorder) {
				t.Errorf("UnfocusedBg must be distinct from Dim and FocusBorder; got %s", th.UnfocusedBg)
			}
			// T-171: DragHandle populated and distinct from DividerColor + FocusBorder
			// (config R4 AC 9 + AC 10).
			if string(th.DragHandle) == "" {
				t.Error("Tier 23 DragHandle token not populated")
			}
			if string(th.DragHandle) == string(th.DividerColor) {
				t.Errorf("DragHandle must be distinct from DividerColor; got %s", th.DragHandle)
			}
			if string(th.DragHandle) == string(th.FocusBorder) {
				t.Errorf("DragHandle must be distinct from FocusBorder; got %s", th.DragHandle)
			}
		})
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
			if !(ld < ldh && ldh < lf) {
				t.Errorf("luminance not ordered: Divider=%.4f DragHandle=%.4f Focus=%.4f", ld, ldh, lf)
			}
			if gap := ldh - ld; gap < minGap {
				t.Errorf("Divider→DragHandle gap %.4f below perceptual threshold %.2f", gap, minGap)
			}
			if gap := lf - ldh; gap < minGap {
				t.Errorf("DragHandle→Focus gap %.4f below perceptual threshold %.2f", gap, minGap)
			}
		})
	}
}

// wcagLuminance returns WCAG 2.x relative luminance (0..1) for an sRGB
// hex color like "#5a6475". Matches the formula used in T-175's one-off
// verification — see loop-log.md Iteration 45.
func wcagLuminance(t *testing.T, hex string) float64 {
	t.Helper()
	s := strings.TrimPrefix(hex, "#")
	if len(s) != 6 {
		t.Fatalf("bad hex color: %q", hex)
	}
	ch := func(i int) float64 {
		v, err := strconv.ParseInt(s[i:i+2], 16, 32)
		if err != nil {
			t.Fatalf("parse %q: %v", s[i:i+2], err)
		}
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
	if th.Name != DefaultThemeName {
		t.Errorf("unknown theme should fallback to %s, got %s", DefaultThemeName, th.Name)
	}
}

func TestBuiltinNames(t *testing.T) {
	names := BuiltinNames()
	if len(names) != 3 {
		t.Fatalf("want 3 built-in themes, got %d", len(names))
	}
}

func TestDefaultThemeName(t *testing.T) {
	if DefaultThemeName != "tokyo-night" {
		t.Errorf("DefaultThemeName = %q, want tokyo-night", DefaultThemeName)
	}
}
