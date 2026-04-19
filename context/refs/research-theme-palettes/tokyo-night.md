---
generated: "2026-04-19"
topic: "tokyo-night canonical palette"
sources:
  - https://github.com/tokyo-night/tokyo-night-vscode-theme
  - https://github.com/tokyo-night/tokyo-night-vscode-theme/blob/master/README.md
  - https://github.com/tokyo-night/tokyo-night-vscode-theme/blob/master/themes/tokyo-night-storm-color-theme.json
---

# Raw Findings: Tokyo Night Canonical Palette

## Upstream Repository

- Original author: enkia (repo now at tokyo-night/tokyo-night-vscode-theme after migration)
- URL: https://github.com/tokyo-night/tokyo-night-vscode-theme
- Three variants: **tokyo-night** (darkest), **tokyo-night-storm** (slightly lighter), **tokyo-night-light** (inverted)

## Variant Identification

| Variant          | Editor Background | Terminal/Sidebar Bg |
|------------------|-------------------|---------------------|
| tokyo-night      | #1a1b26           | #16161e             |
| tokyo-night-storm| #24283b           | #1f2335             |
| tokyo-night-light| #e6e7ed           | —                   |

**Gloggy's variant**: UnfocusedBg=#16161e matches terminal bg of **tokyo-night** (not storm). HeaderBg=#1f2335 matches storm's terminal/sidebar bg — this is a DRIFT. In canonical tokyo-night, #1f2335 is the storm sidebar bg, not the primary tokyo-night sidebar bg. PRIMARY identification: gloggy's tokyo-night is based on **tokyo-night** (not storm) for syntax and accents, but uses a storm-range color for HeaderBg.

## Canonical Base Colors (tokyo-night variant)

| Role                        | Hex     | Notes                              |
|-----------------------------|---------|-------------------------------------|
| Editor bg (bg0)             | #1a1b26 | Signature navy — identity color    |
| Terminal/sidebar bg (bg-1)  | #16161e | Darker variant bg                  |
| Storm sidebar (bg1)         | #1f2335 | Used as gloggy HeaderBg — drift    |
| Storm editor bg             | #24283b | Not used by gloggy                 |
| Editor foreground           | #a9b1d6 | Cool blue-tinted white             |
| Terminal foreground         | #787c99 | Muted blue-grey                    |

## Canonical Syntax/Accent Colors (shared across night+storm)

| Role                        | Hex     | Gloggy field   | Match? |
|-----------------------------|---------|----------------|--------|
| Red (tags, errors)          | #f7768e | LevelError     | EXACT  |
| Orange (numbers)            | #ff9e64 | SyntaxNumber, SearchHighlight | EXACT |
| Yellow (constants, params)  | #e0af68 | LevelWarn, Mark | EXACT |
| Green (strings)             | #9ece6a | SyntaxString   | EXACT  |
| Teal/cyan (keys, operators) | #73daca | SyntaxKey      | EXACT  |
| Blue (functions, info)      | #7aa2f7 | LevelInfo, FocusBorder | EXACT |
| Purple (keywords)           | #bb9af7 | SyntaxBoolean  | EXACT  |
| Comment grey                | #565f89 | LevelDebug, SyntaxNull | EXACT |
| Operators (bright cyan)     | #89ddff | not mapped     | MISSING |
| Variables                   | #c0caf5 | not mapped     | MISSING |

## Distinctive Character

- **Identity color**: Navy #1a1b26 — very cool, blue-indigo bg with purple undertones
- **Temperature**: Cold (dominant blue/purple family)
- **Saturation**: Medium — accents are vivid but bg is deeply desaturated
- **Signature**: The navy editor bg is unmistakable; no other popular theme shares it
- **Purple accent**: #bb9af7 as keywords is more purple-biased than catppuccin's mauve or material's violet

## Missing Upstream Roles (not in gloggy Theme struct)

| Upstream role           | Canonical hex | TUI use?                            |
|-------------------------|---------------|-------------------------------------|
| Selection bg            | #364a82       | YES — row selection (gloggy uses this as CursorHighlight) |
| Comment/invisible       | #565f89       | YES — already used as LevelDebug/SyntaxNull |
| Indent guide / whitespace | #3b4261     | YES — could map to DividerColor (close match already) |
| Line highlight          | #1f2335       | MAYBE — currently used as HeaderBg  |
| Git added               | #9ece6a       | LOW — no git integration in gloggy  |
| Git modified            | #e0af68       | LOW — no git integration in gloggy  |
| Git deleted             | #f7768e       | LOW — no git integration in gloggy  |
| Operators               | #89ddff       | POSSIBLE — SyntaxKey currently maps to teal; bright cyan for operators would add hierarchy |

## Confidence

HIGH — hex values verified from upstream JSON theme files.
