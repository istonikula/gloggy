---
generated: "2026-04-19"
topic: "TUI palette fidelity practices — how other tools handle this"
sources:
  - https://dandavison.github.io/delta/custom-themes.html
  - https://jvns.ca/blog/2024/10/01/terminal-colours/
  - https://nvim-mini.org/mini.nvim/doc/mini-base16.html
  - https://github.com/chriskempson/base16/blob/master/styling.md
  - https://raw.githubusercontent.com/catppuccin/catppuccin/main/docs/style-guide.md
---

# Raw Findings: TUI Palette Fidelity Practices

## How Other TUIs Structure Themes

### delta (git diff pager)

- A "theme" in delta is a named feature (group of settings), NOT a struct.
- Separates concerns: `syntax-theme` (bat/syntect palette for code highlighting) is decoupled from background colors and UI decorations.
- This allows using, e.g., `syntax-theme = "Tokyo Night"` while independently specifying `minus-style`, `plus-style`, `hunk-header-style` colors.
- Key insight: **delta separates base palette from semantic role assignment**. Users can pick a palette and remap roles independently.

### bat

- Uses TextMate/syntect themes for syntax; separate `--theme` from `--color` for UI.
- Tokyo Night, Catppuccin are first-class supported themes.
- No explicit "base role" concept — it delegates to the upstream theme's full color definition.

### lazygit

- Theme config uses explicit named roles: `activeBorderColor`, `inactiveBorderColor`, `selectedLineBgColor`, `cherryPickedCommitBgColor`, etc.
- Does NOT inherit from a named palette — user specifies each role as hex array.
- This is the "full role expansion" approach: TUI authors define all semantic roles explicitly.

### btop

- Ships named themes as files with a fixed set of variables: `main_bg`, `main_fg`, `title`, `hi_fg`, `selected_bg`, `selected_fg`, `inactive_fg`, `proc_misc`, `cpu_box`, `net_box`, etc.
- Catppuccin and Tokyo Night are available as community theme files.
- Each theme file maps palette colors to ~20 semantic roles — very similar scope to gloggy.

### Neovim colorschemes (e.g. tokyonight.nvim, catppuccin.nvim)

- Best-in-class ports define a `palette` table (named hex values) separately from `highlights` (semantic role assignments).
- Two-layer architecture: `palette.bg = "#1a1b26"`, then `highlights.Normal = { bg = palette.bg }`.
- This makes "swap the palette" trivial while keeping semantic intent.

### Base16 Framework

16 named slots: base00 (bg) through base0F (deprecated). Key semantic assignments:
- base00 = default bg
- base01 = lighter bg (status bar, line numbers)
- base02 = selection bg
- base03 = comments/invisibles/line highlight
- base04 = dark fg (status bars)
- base05 = default fg
- base08 = variables, errors, diff-deleted
- base09 = integers, booleans, constants
- base0A = classes, search text bg
- base0B = strings, diff-inserted
- base0C = support, regex, escape chars
- base0D = functions, headings
- base0E = keywords, storage
- base0F = deprecated

Key observation: base16 has **3 background levels** (base00/01/02) and **2 foreground levels** (base04/05) explicitly in the framework. Gloggy currently has 2 explicit bg levels (HeaderBg/UnfocusedBg) but no general-purpose base bg token.

## Pitfalls When Porting Editor Themes to TUIs

### 1. Missing base background token
Editor themes paint the bg explicitly. TUIs inherit terminal bg by default. This means gloggy never actually paints the #1a1b26 tokyo-night navy or #212121 material dark bg — users see their terminal's configured bg instead. For faithful reproduction, gloggy would need a `BaseBg` token used as full-pane background.

**Impact**: HIGH — this is the most noticeable fidelity gap. The entire identity of tokyo-night comes from its navy bg.

### 2. Terminal bg collision with UnfocusedBg
When UnfocusedBg is set to a near-terminal-default color (e.g. #11111b for catppuccin on a dark terminal), the unfocused pane may become invisible. The token needs to be distinctly tinted relative to whatever the terminal default is — hard to know without knowing the terminal.

### 3. Truecolor vs 256-color fallback
lipgloss passes colors as hex; if the terminal doesn't support truecolor (COLORTERM=truecolor), the terminal approximates to 256-color. Themes that rely on subtle hue differences (e.g. catppuccin's teal #94e2d5 vs sky #89dceb) will collapse to the same 256-color index. Gloggy currently has no 256-color fallback path.

### 4. Convergent semantic assignments
All three themes use: red=error, blue=info, yellow=warning, green=string, purple=keyword. This is the primary reason they look similar. To distinguish them, the non-semantic roles must diverge: backgrounds, dim colors, focus colors, surface tiers.

### 5. Role aliasing reduces distinctiveness
SyntaxNull == LevelDebug (same hex) in all themes; SearchHighlight == SyntaxNumber; FocusBorder == LevelInfo. Using different tokens for these roles would unlock per-theme visual diversity without changing semantic intent.

### 6. Ignoring identity colors
Each theme has 1-2 "identity" colors that make it instantly recognizable. Gloggy currently does not expose these:
- Tokyo-night: #1a1b26 navy bg (not rendered)
- Catppuccin-mocha: #b4befe lavender (not used — FocusBorder uses Blue instead)
- Material-dark: #212121 near-black bg + #89ddff electric cyan (bg not rendered, cyan IS used as SyntaxKey)

### 7. Two-layer vs one-layer architecture
Gloggy's constructors hardcode hex literals directly into semantic roles. There is no intermediate "palette" layer. This means to add a new role or fix a drift, the developer must know the canonical palette hex from memory. The two-layer pattern (palette constants → role assignments) is more maintainable and self-documenting.

## Recommendation from Practice

The most robust pattern (used by catppuccin.nvim, tokyonight.nvim, btop theme files):
1. Define a `palette` type with canonical named colors (bg0, bg1, bg2, fg0, red, green, yellow, blue, purple, cyan, orange, lavender/identity).
2. In the theme constructor, assign `Theme.SyntaxKey = palette.teal` (by name, not hex).
3. This makes drift immediately visible during review and makes port fidelity checkable by inspection.
