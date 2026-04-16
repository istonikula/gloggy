package theme

import "testing"

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
		})
	}
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
