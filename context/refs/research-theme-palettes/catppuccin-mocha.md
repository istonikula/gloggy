---
generated: "2026-04-19"
topic: "catppuccin-mocha canonical palette"
sources:
  - https://github.com/catppuccin/catppuccin
  - https://github.com/catppuccin/palette
  - https://raw.githubusercontent.com/catppuccin/catppuccin/main/docs/style-guide.md
---

# Raw Findings: Catppuccin Mocha Canonical Palette

## Upstream Repository

- URL: https://github.com/catppuccin/catppuccin
- Palette JSON: https://github.com/catppuccin/palette
- Four flavors: Latte (light), Frappe, Macchiato, **Mocha** (darkest)
- Community-driven pastel theme; consistent semantic mappings across 300+ ports

## Canonical Mocha Palette (all 26 named colors)

### Accent Colors
| Name       | Hex     | Semantic role per style guide          |
|------------|---------|----------------------------------------|
| Rosewater  | #f5e0dc | Cursor                                 |
| Flamingo   | #f2cdcd | —                                      |
| Pink       | #f5c2e7 | —                                      |
| Mauve      | #cba6f7 | Keywords                               |
| Red        | #f38ba8 | Errors; booleans/atoms/null-like       |
| Maroon     | #eba0ac | Parameters                             |
| Peach      | #fab387 | Numbers/Constants                      |
| Yellow     | #f9e2af | Classes/Types; Warnings                |
| Green      | #a6e3a1 | Strings; Success; Diff inserted        |
| Teal       | #94e2d5 | Info (secondary)                       |
| Sky        | #89dceb | Operators                              |
| Sapphire   | #74c7ec | —                                      |
| Blue       | #89b4fa | Functions/Methods; Links               |
| Lavender   | #b4befe | IDENTITY COLOR — active borders, visited links |

### Text / Surface Colors
| Name       | Hex     | Semantic role per style guide          |
|------------|---------|----------------------------------------|
| Text       | #cdd6f4 | Default foreground                     |
| Subtext1   | #bac2de | Secondary foreground                   |
| Subtext0   | #a6adc8 | Tertiary foreground                    |
| Overlay2   | #9399b2 | Comments; Selection fg                 |
| Overlay1   | #7f849c | —                                      |
| Overlay0   | #6c7086 | Inactive borders                       |
| Surface2   | #585b70 | Selection bg (20-30% opacity guide)    |
| Surface1   | #45475a | Status bar bg; lighter surface         |
| Surface0   | #313244 | Popup bg; elevated surface             |
| Base       | #1e1e2e | IDENTITY — primary editor bg           |
| Mantle     | #181825 | Secondary bg (header/sidebar)          |
| Crust      | #11111b | Deepest bg (terminal bg, borders)      |

## Gloggy Field Mapping vs Canonical

| Gloggy Field    | Gloggy Hex | Canonical Name | Canonical Hex | Match?      |
|-----------------|------------|----------------|---------------|-------------|
| LevelError      | #f38ba8    | Red            | #f38ba8       | EXACT       |
| LevelWarn       | #f9e2af    | Yellow         | #f9e2af       | EXACT       |
| LevelInfo       | #89b4fa    | Blue           | #89b4fa       | EXACT       |
| LevelDebug      | #6c7086    | Overlay0       | #6c7086       | EXACT       |
| SyntaxKey       | #94e2d5    | Teal           | #94e2d5       | EXACT       |
| SyntaxString    | #a6e3a1    | Green          | #a6e3a1       | EXACT       |
| SyntaxNumber    | #fab387    | Peach          | #fab387       | EXACT       |
| SyntaxBoolean   | #cba6f7    | Mauve          | #cba6f7       | EXACT       |
| SyntaxNull      | #6c7086    | Overlay0       | #6c7086       | EXACT (same as LevelDebug) |
| Mark            | #f9e2af    | Yellow         | #f9e2af       | EXACT       |
| Dim             | #45475a    | Surface1       | #45475a       | EXACT       |
| SearchHighlight | #fab387    | Peach          | #fab387       | EXACT (same as SyntaxNumber) |
| CursorHighlight | #585b70    | Surface2       | #585b70       | EXACT       |
| HeaderBg        | #181825    | Mantle         | #181825       | EXACT       |
| FocusBorder     | #89b4fa    | Blue           | #89b4fa       | EXACT (same as LevelInfo) |
| DividerColor    | #313244    | Surface0       | #313244       | EXACT       |
| UnfocusedBg     | #11111b    | Crust          | #11111b       | EXACT       |

**Result**: Catppuccin-mocha is the most faithful of the three — ALL 17 mapped colors match canonical palette names exactly.

## Missing Canonical Colors Not in Gloggy

| Missing role         | Canonical color | Hex     | TUI use?                   |
|----------------------|-----------------|---------|----------------------------|
| Lavender             | Lavender        | #b4befe | YES — identity color, would distinguish mocha from tokyo-night visually. Could replace FocusBorder to differentiate from LevelInfo |
| Sky (operators)      | Sky             | #89dceb | LOW — no operator syntax in log viewer |
| Rosewater (cursor)   | Rosewater       | #f5e0dc | POSSIBLE — alternative CursorHighlight or separate cursor-fg |
| Maroon (parameters)  | Maroon          | #eba0ac | LOW — no parameters in log JSON |

## Distinctive Character

- **Identity color**: Lavender #b4befe — warm purple-blue, used for active borders in official ports. Gloggy uses Blue for FocusBorder — MISSED OPPORTUNITY. Switching FocusBorder to #b4befe would make mocha instantly recognizable vs tokyo-night (which also uses a blue FocusBorder).
- **Temperature**: Neutral-warm — blue-grey base with subtle purple undertones, warmer than tokyo-night
- **Saturation**: Pastel — all accents are notably softer/more muted than material-dark
- **Signature**: Lavender as border/accent color; the "crust/mantle/base" three-level bg hierarchy

## Confidence

HIGH — all palette values verified from upstream palette.json via WebFetch.
