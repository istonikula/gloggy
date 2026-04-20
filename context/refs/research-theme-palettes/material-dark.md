---
generated: "2026-04-19"
topic: "material-dark canonical palette + disambiguation"
sources:
  - https://material-theme.com/docs/reference/color-palette/ (403 — blocked)
  - https://marketplace.visualstudio.com/items?itemName=Equinusocio.vsc-material-theme
  - https://github.com/myambitions/vsc-community-material-theme
  - https://github.com/material-theme/vsc-material-theme (now vira-themes)
  - Material Theme Documentation color-palette page (indirect, via web search)
---

# Raw Findings: Material Dark — Canonical Palette + Disambiguation

## DISAMBIGUATION: Which Material Theme?

There are THREE distinct "material" themes in circulation:

1. **Material Theme by Mattia Astorino (equinusocio)** — The VSCode theme cited by gloggy's comment `// materialDark — https://material-theme.com`. Originally ~9 million installs. Now DEPRECATED as of 2024; official successor is Vira Theme. Community-maintained fork: `vsc-community-material-theme`.

2. **Google Material Design 3 dark scheme** — Google's design system color tokens (M3). Uses tonal palettes, dynamic color. Background is #1C1B1F (near-black warm grey). NOT what gloggy tracks.

3. **Material You / material-code** — Dynamic M3 port for VSCode by rakibdev. NOT what gloggy tracks.

**Conclusion (HIGH confidence)**: Gloggy tracks **Material Theme by Mattia Astorino / equinusocio** (the VSCode theme, not Google's M3). The hex values match exactly. The `#1a1a1a` background (gloggy HeaderBg) and `#0d0d0d` UnfocusedBg both track the deep-black aesthetic of Astorino's original. The syntax colors (#f07178, #c792ea, #89ddff, #c3e88d) are the original Material Theme Dark signature palette.

## Canonical Material Theme Dark Palette (Astorino original)

### Background/UI Colors
| Role                  | Hex     | Notes                                    |
|-----------------------|---------|------------------------------------------|
| Editor background     | #212121 | Near-black; IDENTITY color               |
| Contrast/deepest bg   | #1a1a1a | Used in gloggy as HeaderBg — close        |
| Second background     | #292929 | Slightly lighter panels                  |
| Selection background  | #404040 | Pure grey selection                      |
| Active tab/element    | #323232 | —                                        |
| Foreground (text)     | #eeffff | Near-white, cool tinted                  |
| Muted/secondary fg    | #b0bec5 | Blue-grey (Material Blue Grey 200)       |

### Syntax/Accent Colors
| Role                  | Hex     | Gloggy field     | Match?       |
|-----------------------|---------|------------------|--------------|
| Red (tags, errors)    | #f07178 | LevelError       | EXACT        |
| Orange (numbers)      | #f78c6c | SyntaxNumber, SearchHighlight | EXACT |
| Yellow (attributes)   | #ffcb6b | LevelWarn, Mark  | EXACT        |
| Green (strings)       | #c3e88d | SyntaxString     | EXACT        |
| Cyan/teal (operators) | #89ddff | SyntaxKey        | EXACT        |
| Blue (functions)      | #82aaff | LevelInfo, FocusBorder | EXACT   |
| Purple (keywords)     | #c792ea | SyntaxBoolean    | EXACT        |
| Comment grey          | #616161 | NOT mapped       | MISSING      |
| Error bright          | #ff5370 | NOT mapped       | MISSING      |
| Parameters (orange2)  | #f78c6c | SyntaxNumber     | (same token) |

## Gloggy Drift Analysis

| Gloggy Field    | Gloggy Hex | Canonical Hex | Notes                                              |
|-----------------|------------|---------------|----------------------------------------------------|
| LevelError      | #f07178    | #f07178       | EXACT                                              |
| LevelWarn       | #ffcb6b    | #ffcb6b       | EXACT                                              |
| LevelInfo       | #82aaff    | #82aaff       | EXACT                                              |
| LevelDebug      | #676e95    | (no canonical) | Gloggy invented a muted violet-grey for debug. Upstream comment color is #616161 (neutral grey) |
| SyntaxKey       | #89ddff    | #89ddff (cyan) | EXACT                                             |
| SyntaxString    | #c3e88d    | #c3e88d       | EXACT                                              |
| SyntaxNumber    | #f78c6c    | #f78c6c       | EXACT                                              |
| SyntaxBoolean   | #c792ea    | #c792ea       | EXACT                                              |
| SyntaxNull      | #676e95    | (no canonical) | Same as LevelDebug; Upstream null uses comment grey #616161 |
| Dim             | #4a4a6a    | (no canonical) | Gloggy invented a violet-tinted dim. Upstream has #474747 disabled or #616161 comments |
| CursorHighlight | #4a5568    | (no canonical) | #4a5568 is from Tailwind CSS slate-600; not from Material Theme upstream |
| HeaderBg        | #1a1a1a    | #1a1a1a (contrast) | Matches "Contrast" color — reasonable         |
| FocusBorder     | #82aaff    | #82aaff (blue) | EXACT (same as LevelInfo)                         |
| DividerColor    | #37474f    | #37474f (Blue Grey 800) | EXACT — a Material Blue Grey color       |
| UnfocusedBg     | #0d0d0d    | #0d0d0d (near-black) | Close; no exact canonical match              |

**Background drift**: Canonical editor bg is #212121 but gloggy has no bg token for this — it relies on terminal default. This means gloggy material-dark doesn't have the iconic #212121 background rendered anywhere.

## Missing Upstream Roles Not in Gloggy

| Missing role           | Canonical hex | TUI use?                                   |
|------------------------|---------------|--------------------------------------------|
| Error bright           | #ff5370       | POSSIBLE — alternative error highlight     |
| Comment grey           | #616161       | LOW — SyntaxNull/LevelDebug could use this instead of the violet #676e95 |
| Foreground (eeffff)    | not needed — terminal default              |

## Distinctive Character

- **Identity color**: Near-black #212121 background + #89ddff electric cyan for operators/keys
- **Temperature**: Neutral — between warm and cool; blue-grey surface colors
- **Saturation**: HIGHER than both tokyo-night and catppuccin — accents are vivid and saturated
- **Signature**: Electric cyan SyntaxKey (#89ddff) contrasts starkly with the deep dark bg; stronger visual punch than the teal/softer versions in the other two themes
- **Bg depth**: Deepest of the three. #0d0d0d UnfocusedBg is near-black; catppuccin-mocha is #11111b, tokyo-night is #16161e.

## Confidence

MEDIUM-HIGH — syntax colors confirmed exact from multiple sources. Background canonical is #212121 (not rendered by gloggy since it lacks a "base bg" token). Some UI colors (CursorHighlight, Dim, LevelDebug) appear to be gloggy-invented values not tracking any Material Theme upstream token.
