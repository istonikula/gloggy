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

	// Visual-polish tokens (Tier 8)
	CursorHighlight lipgloss.Color
	HeaderBg        lipgloss.Color
	FocusBorder     lipgloss.Color

	// Pane state tokens (Tier 9 — details-pane redesign).
	// DividerColor: quiet neutral for the right-split divider column.
	// Reads closer to Dim than to FocusBorder; never recolors on focus change.
	// UnfocusedBg: subtle background tint for unfocused-but-visible panes;
	// distinct from Dim (which is a foreground intensity).
	DividerColor lipgloss.Color
	UnfocusedBg  lipgloss.Color

	// DragHandle (Tier 23): mid-tone neutral painted on the pane-resize
	// drag seam (right-mode divider glyph; below-mode shared border row).
	// Brighter than DividerColor, dimmer than FocusBorder; focus-neutral.
	DragHandle lipgloss.Color

	// BaseBg (Tier 24): primary pane background. Painted on all rendered
	// pane surfaces wherever UnfocusedBg does not apply, so no pane falls
	// through to the terminal default bg (see config R4 ACs 12-15).
	BaseBg lipgloss.Color
}

// DefaultThemeName is the theme used when none is specified or an unknown name is given.
const DefaultThemeName = "tokyo-night"

// Canonical-source citations for each bundled theme (config R4 AC 16).
// Each constant names the upstream repo URL and the palette variant so
// drift is discoverable at test time and at code review. If the underlying
// hex values drift from the upstream palette, update both the palette
// struct below AND (if the variant changes) the citation constant.
const (
	TokyoNightSource      = "https://github.com/enkia/tokyo-night-vscode-theme (night variant)"
	CatppuccinMochaSource = "https://github.com/catppuccin/catppuccin (mocha flavor)"
	MaterialDarkSource    = "https://github.com/myambitions/vsc-community-material-theme (Astorino legacy palette)"
)

var builtinThemes = map[string]Theme{
	"tokyo-night":      tokyoNight(),
	"catppuccin-mocha": catppuccinMocha(),
	"material-dark":    materialDark(),
}

// BuiltinNames returns the names of all built-in themes.
func BuiltinNames() []string {
	return []string{"tokyo-night", "catppuccin-mocha", "material-dark"}
}

// Source returns the canonical-source citation for a bundled theme name.
// Returns an empty string for unknown theme names.
func Source(name string) string {
	switch name {
	case "tokyo-night":
		return TokyoNightSource
	case "catppuccin-mocha":
		return CatppuccinMochaSource
	case "material-dark":
		return MaterialDarkSource
	}
	return ""
}

// GetTheme returns the named theme. Unknown names fall back to the default with a warning.
func GetTheme(name string) Theme {
	if t, ok := builtinThemes[name]; ok {
		return t
	}
	fmt.Fprintf(os.Stderr, "gloggy: unknown theme %q, falling back to %s\n", name, DefaultThemeName)
	return builtinThemes[DefaultThemeName]
}

// tokyoNightPalette holds the canonical hex values for the tokyo-night
// night variant. Field names follow the upstream vocabulary where it
// exists; gloggy-specific picks (T-171 drag-seam) are flagged.
// See TokyoNightSource.
var tokyoNightPalette = struct {
	bgEditor, bgDarker, bgSidebarStorm string // bg0 (identity), bg1, storm-variant sidebar bg
	bgDivider, bgDim, bgCursor         string // dark3, panel-grey, selection highlight
	red, orange, yellow, green         string
	teal, blue, purple                 string
	commentGrey                        string // #565f89 — LevelDebug + SyntaxNull
	dragHandleMid                      string // T-171 mid-tone neutral; not in upstream palette
}{
	bgEditor: "#1a1b26", bgDarker: "#16161e", bgSidebarStorm: "#1f2335",
	bgDivider: "#3b4261", bgDim: "#414868", bgCursor: "#364a82",
	red: "#f7768e", orange: "#ff9e64", yellow: "#e0af68", green: "#9ece6a",
	teal: "#73daca", blue: "#7aa2f7", purple: "#bb9af7",
	commentGrey:   "#565f89",
	dragHandleMid: "#5a6475",
}

// tokyoNight — see TokyoNightSource.
func tokyoNight() Theme {
	p := tokyoNightPalette
	return Theme{
		Name:            "tokyo-night",
		LevelError:      lipgloss.Color(p.red),
		LevelWarn:       lipgloss.Color(p.yellow),
		LevelInfo:       lipgloss.Color(p.blue),
		LevelDebug:      lipgloss.Color(p.commentGrey),
		SyntaxKey:       lipgloss.Color(p.teal),
		SyntaxString:    lipgloss.Color(p.green),
		SyntaxNumber:    lipgloss.Color(p.orange),
		SyntaxBoolean:   lipgloss.Color(p.purple),
		SyntaxNull:      lipgloss.Color(p.commentGrey),
		Mark:            lipgloss.Color(p.yellow),
		Dim:             lipgloss.Color(p.bgDim),
		SearchHighlight: lipgloss.Color(p.orange),
		CursorHighlight: lipgloss.Color(p.bgCursor),
		HeaderBg:        lipgloss.Color(p.bgSidebarStorm),
		FocusBorder:     lipgloss.Color(p.blue),
		DividerColor:    lipgloss.Color(p.bgDivider),
		UnfocusedBg:     lipgloss.Color(p.bgDarker),
		DragHandle:      lipgloss.Color(p.dragHandleMid),
		BaseBg:          lipgloss.Color(p.bgEditor),
	}
}

// catppuccinMochaPalette holds the canonical hex values for catppuccin
// mocha. Field names match the upstream style-guide vocabulary (Base,
// Mantle, Crust, Surface0/1/2, Overlay0, plus the accents). FocusBorder
// uses Lavender per the upstream ports (neovim, lazygit, btop) — not
// Blue (T-178, config R4 AC 17). See CatppuccinMochaSource.
var catppuccinMochaPalette = struct {
	base, mantle, crust              string // bg hierarchy
	surface0, surface1, surface2     string // elevations
	overlay0                         string // dim text
	red, peach, yellow, green        string
	teal, blue, mauve, lavender      string
	dragHandleMid                    string // T-171 mid-tone neutral; not in upstream palette
}{
	base: "#1e1e2e", mantle: "#181825", crust: "#11111b",
	surface0: "#313244", surface1: "#45475a", surface2: "#585b70",
	overlay0: "#6c7086",
	red:      "#f38ba8", peach: "#fab387", yellow: "#f9e2af", green: "#a6e3a1",
	teal:     "#94e2d5", blue: "#89b4fa", mauve: "#cba6f7", lavender: "#b4befe",
	dragHandleMid: "#6e7388",
}

// catppuccinMocha — see CatppuccinMochaSource.
func catppuccinMocha() Theme {
	p := catppuccinMochaPalette
	return Theme{
		Name:            "catppuccin-mocha",
		LevelError:      lipgloss.Color(p.red),
		LevelWarn:       lipgloss.Color(p.yellow),
		LevelInfo:       lipgloss.Color(p.blue),
		LevelDebug:      lipgloss.Color(p.overlay0),
		SyntaxKey:       lipgloss.Color(p.teal),
		SyntaxString:    lipgloss.Color(p.green),
		SyntaxNumber:    lipgloss.Color(p.peach),
		SyntaxBoolean:   lipgloss.Color(p.mauve),
		SyntaxNull:      lipgloss.Color(p.overlay0),
		Mark:            lipgloss.Color(p.yellow),
		Dim:             lipgloss.Color(p.surface1),
		SearchHighlight: lipgloss.Color(p.peach),
		CursorHighlight: lipgloss.Color(p.surface2),
		HeaderBg:        lipgloss.Color(p.mantle),
		FocusBorder:     lipgloss.Color(p.lavender),
		DividerColor:    lipgloss.Color(p.surface0),
		UnfocusedBg:     lipgloss.Color(p.crust),
		DragHandle:      lipgloss.Color(p.dragHandleMid),
		BaseBg:          lipgloss.Color(p.base),
	}
}

// materialDarkPalette holds hex values for the Astorino legacy material
// dark palette. Fields marked `gloggy-specific` were synthesized for this
// project (not from Astorino); see research-brief-theme-palettes.md for
// drift notes on mutedPurple / dimGreyBlue / cursorGreyBlue / unfocusedBg.
// See MaterialDarkSource.
var materialDarkPalette = struct {
	bgEditor, bgContrast, bgUnfocused string // identity + chrome; bgUnfocused is gloggy-specific
	bgDivider, bgCursor, bgDim        string // Astorino non-editor border; cursor/dim are gloggy-specific
	red, orange, yellow, green        string
	cyan, blue, purple                string
	mutedPurple                       string // LevelDebug + SyntaxNull — gloggy-specific
	dragHandleMid                     string // T-171 mid-tone neutral; not in upstream palette
}{
	bgEditor: "#212121", bgContrast: "#1a1a1a", bgUnfocused: "#0d0d0d",
	bgDivider: "#37474f", bgCursor: "#4a5568", bgDim: "#4a4a6a",
	red:  "#f07178", orange: "#f78c6c", yellow: "#ffcb6b", green: "#c3e88d",
	cyan: "#89ddff", blue: "#82aaff", purple: "#c792ea",
	mutedPurple:   "#676e95",
	dragHandleMid: "#65737e",
}

// materialDark — see MaterialDarkSource.
func materialDark() Theme {
	p := materialDarkPalette
	return Theme{
		Name:            "material-dark",
		LevelError:      lipgloss.Color(p.red),
		LevelWarn:       lipgloss.Color(p.yellow),
		LevelInfo:       lipgloss.Color(p.blue),
		LevelDebug:      lipgloss.Color(p.mutedPurple),
		SyntaxKey:       lipgloss.Color(p.cyan),
		SyntaxString:    lipgloss.Color(p.green),
		SyntaxNumber:    lipgloss.Color(p.orange),
		SyntaxBoolean:   lipgloss.Color(p.purple),
		SyntaxNull:      lipgloss.Color(p.mutedPurple),
		Mark:            lipgloss.Color(p.yellow),
		Dim:             lipgloss.Color(p.bgDim),
		SearchHighlight: lipgloss.Color(p.orange),
		CursorHighlight: lipgloss.Color(p.bgCursor),
		HeaderBg:        lipgloss.Color(p.bgContrast),
		FocusBorder:     lipgloss.Color(p.blue),
		DividerColor:    lipgloss.Color(p.bgDivider),
		UnfocusedBg:     lipgloss.Color(p.bgUnfocused),
		DragHandle:      lipgloss.Color(p.dragHandleMid),
		BaseBg:          lipgloss.Color(p.bgEditor),
	}
}
