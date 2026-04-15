// Package theme provides color theme definitions for gloggy's TUI.
package theme

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Theme defines all color tokens used across the gloggy TUI.
type Theme struct {
	Name string

	// Level badge colors
	LevelError lipgloss.Color
	LevelWarn  lipgloss.Color
	LevelInfo  lipgloss.Color
	LevelDebug lipgloss.Color

	// Syntax highlighting colors
	SyntaxKey     lipgloss.Color
	SyntaxString  lipgloss.Color
	SyntaxNumber  lipgloss.Color
	SyntaxBoolean lipgloss.Color
	SyntaxNull    lipgloss.Color

	// UI element colors
	Mark            lipgloss.Color
	Dim             lipgloss.Color
	SearchHighlight lipgloss.Color
}

// DefaultThemeName is the theme used when none is specified or an unknown name is given.
const DefaultThemeName = "tokyo-night"

var builtinThemes = map[string]Theme{
	"tokyo-night":      tokyoNight(),
	"catppuccin-mocha": catppuccinMocha(),
	"material-dark":    materialDark(),
}

// BuiltinNames returns the names of all built-in themes.
func BuiltinNames() []string {
	return []string{"tokyo-night", "catppuccin-mocha", "material-dark"}
}

// GetTheme returns the named theme. Unknown names fall back to the default with a warning.
func GetTheme(name string) Theme {
	if t, ok := builtinThemes[name]; ok {
		return t
	}
	fmt.Fprintf(os.Stderr, "gloggy: unknown theme %q, falling back to %s\n", name, DefaultThemeName)
	return builtinThemes[DefaultThemeName]
}

// tokyoNight — https://github.com/enkia/tokyo-night-vscode-theme
func tokyoNight() Theme {
	return Theme{
		Name:            "tokyo-night",
		LevelError:      lipgloss.Color("#f7768e"),
		LevelWarn:       lipgloss.Color("#e0af68"),
		LevelInfo:       lipgloss.Color("#7aa2f7"),
		LevelDebug:      lipgloss.Color("#565f89"),
		SyntaxKey:       lipgloss.Color("#73daca"),
		SyntaxString:    lipgloss.Color("#9ece6a"),
		SyntaxNumber:    lipgloss.Color("#ff9e64"),
		SyntaxBoolean:   lipgloss.Color("#bb9af7"),
		SyntaxNull:      lipgloss.Color("#565f89"),
		Mark:            lipgloss.Color("#e0af68"),
		Dim:             lipgloss.Color("#414868"),
		SearchHighlight: lipgloss.Color("#ff9e64"),
	}
}

// catppuccinMocha — https://github.com/catppuccin/catppuccin
func catppuccinMocha() Theme {
	return Theme{
		Name:            "catppuccin-mocha",
		LevelError:      lipgloss.Color("#f38ba8"),
		LevelWarn:       lipgloss.Color("#f9e2af"),
		LevelInfo:       lipgloss.Color("#89b4fa"),
		LevelDebug:      lipgloss.Color("#6c7086"),
		SyntaxKey:       lipgloss.Color("#94e2d5"),
		SyntaxString:    lipgloss.Color("#a6e3a1"),
		SyntaxNumber:    lipgloss.Color("#fab387"),
		SyntaxBoolean:   lipgloss.Color("#cba6f7"),
		SyntaxNull:      lipgloss.Color("#6c7086"),
		Mark:            lipgloss.Color("#f9e2af"),
		Dim:             lipgloss.Color("#45475a"),
		SearchHighlight: lipgloss.Color("#fab387"),
	}
}

// materialDark — https://material-theme.com
func materialDark() Theme {
	return Theme{
		Name:            "material-dark",
		LevelError:      lipgloss.Color("#f07178"),
		LevelWarn:       lipgloss.Color("#ffcb6b"),
		LevelInfo:       lipgloss.Color("#82aaff"),
		LevelDebug:      lipgloss.Color("#676e95"),
		SyntaxKey:       lipgloss.Color("#89ddff"),
		SyntaxString:    lipgloss.Color("#c3e88d"),
		SyntaxNumber:    lipgloss.Color("#f78c6c"),
		SyntaxBoolean:   lipgloss.Color("#c792ea"),
		SyntaxNull:      lipgloss.Color("#676e95"),
		Mark:            lipgloss.Color("#ffcb6b"),
		Dim:             lipgloss.Color("#4a4a6a"),
		SearchHighlight: lipgloss.Color("#f78c6c"),
	}
}
