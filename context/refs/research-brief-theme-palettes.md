---
generated: "2026-04-19"
sources_consulted: 14
confidence: HIGH (tokyo-night, catppuccin-mocha), MEDIUM-HIGH (material-dark)
---

# Research Brief: Theme Palette Fidelity

## Summary

Gloggy's three built-in themes are faithful on their syntax/accent tokens but converge visually because: (1) all three use the identical semantic mapping (red=error, blue=info, green=string, etc.), (2) each theme's identity colors are either unmapped or aliased to already-used tokens, and (3) there is no base background token — the terminal bg hides the most distinctive per-theme color entirely. The fix requires adding a `BaseBg` token, promoting each theme's identity accent to a distinct role (catppuccin's lavender to FocusBorder, material's electric cyan remains but bg needs surfacing), and breaking the three hard aliases (SearchHighlight == SyntaxNumber, FocusBorder == LevelInfo, SyntaxNull == LevelDebug).

---

## Key Findings: Tokyo Night

**Variant identification**: Gloggy is based on **tokyo-night** (not storm). Accents and UnfocusedBg (#16161e) match the primary night variant. HeaderBg (#1f2335) is the storm sidebar bg — minor drift, acceptable as a secondary surface.

**Distinctive character**: Identity color is the navy editor bg #1a1b26 (cool blue-indigo). Temperature: cold. Saturation: medium. Purple-biased keyword accent (#bb9af7).

**Canonical full palette (tokyo-night variant)**:

| Role                   | Hex     |
|------------------------|---------|
| bg0 (editor)           | #1a1b26 |
| bg-1 (terminal/darker) | #16161e |
| bg1 (sidebar/storm-level) | #1f2335 |
| fg0 (editor)           | #a9b1d6 |
| fg1 (terminal)         | #787c99 |
| red                    | #f7768e |
| orange                 | #ff9e64 |
| yellow                 | #e0af68 |
| green                  | #9ece6a |
| teal                   | #73daca |
| blue                   | #7aa2f7 |
| purple                 | #bb9af7 |
| cyan (operators)       | #89ddff |
| comment grey           | #565f89 |
| variables              | #c0caf5 |

**Gloggy drift**: LOW — all 17 mapped syntax/accent tokens are canonical exact matches. The only drift is HeaderBg (#1f2335 = storm sidebar bg, not the night sidebar bg #16161e). This is acceptable as a visual hierarchy choice.

**Missing identity**: The navy bg #1a1b26 is never rendered — terminal default is used instead.

---

## Key Findings: Catppuccin Mocha

**Distinctive character**: Identity color is **lavender #b4befe** — a warm purple-blue used as active border color in ALL official catppuccin ports. Temperature: neutral-warm. Saturation: pastel (softest of the three). Three-level bg hierarchy: base/mantle/crust.

**Canonical full palette (mocha flavor)**:

| Role      | Hex     | Role     | Hex     |
|-----------|---------|----------|---------|
| Base (bg) | #1e1e2e | Text     | #cdd6f4 |
| Mantle    | #181825 | Subtext1 | #bac2de |
| Crust     | #11111b | Overlay0 | #6c7086 |
| Surface0  | #313244 | Surface1 | #45475a |
| Surface2  | #585b70 | Overlay2 | #9399b2 |
| Red       | #f38ba8 | Mauve    | #cba6f7 |
| Peach     | #fab387 | Yellow   | #f9e2af |
| Green     | #a6e3a1 | Teal     | #94e2d5 |
| Blue      | #89b4fa | Lavender | #b4befe |
| Sky       | #89dceb | Rosewater| #f5e0dc |

**Gloggy drift**: NONE on syntax/accents — all 17 mapped tokens are canonical exact matches. CRITICAL MISS: FocusBorder uses Blue (#89b4fa) instead of Lavender (#b4befe). Official catppuccin ports (neovim, lazygit, btop) universally use Lavender for active borders. This single change would make mocha instantly distinguishable from tokyo-night (which also uses a blue FocusBorder).

**Catppuccin style guide mappings** (official):
- Keywords: Mauve — gloggy uses Mauve for SyntaxBoolean, correct
- Strings: Green — EXACT
- Numbers: Peach — EXACT
- Booleans/null: Red (upstream) — gloggy uses Mauve/Overlay0 — MISMATCH. Upstream says Red for booleans/atoms; gloggy assigns Mauve (keywords). For a log viewer this is low-stakes but noteworthy.
- Active borders: Lavender — MISSED in gloggy

---

## Key Findings: Material Dark

**Disambiguation**: Gloggy tracks **Material Theme by Mattia Astorino (equinusocio)**, NOT Google Material Design 3. The equinusocio theme is now deprecated (2024); the official successor is Vira Theme. The community-maintained fork (`vsc-community-material-theme`) preserves the original palette. All gloggy syntax hexes match the original Astorino palette exactly.

**Distinctive character**: Identity: near-black bg #212121 + electric cyan #89ddff for operators/keys. Temperature: neutral (between warm and cool). Saturation: HIGHEST of the three — accents are vivid, not pastel.

**Canonical full palette (Material Theme Dark by Astorino)**:

| Role                  | Hex     |
|-----------------------|---------|
| Editor bg (IDENTITY)  | #212121 |
| Contrast/deepest bg   | #1a1a1a |
| Second bg             | #292929 |
| Selection bg          | #404040 |
| Foreground            | #eeffff |
| Muted fg              | #b0bec5 |
| Red (errors/tags)     | #f07178 |
| Error bright          | #ff5370 |
| Orange (numbers)      | #f78c6c |
| Yellow (attributes)   | #ffcb6b |
| Green (strings)       | #c3e88d |
| Cyan/operators        | #89ddff |
| Blue (functions)      | #82aaff |
| Purple (keywords)     | #c792ea |
| Comment grey          | #616161 |
| Blue Grey 800 (border)| #37474f |

**Gloggy drift**:
- LevelDebug (#676e95) — INVENTED. Canonical comment grey is #616161 (neutral); gloggy's debug uses a violet-grey not from the palette.
- SyntaxNull (#676e95) — INVENTED. Same token as LevelDebug; no canonical basis.
- Dim (#4a4a6a) — INVENTED. Violet-tinted dim; upstream has #474747 (neutral grey).
- CursorHighlight (#4a5568) — INVENTED. Appears to be Tailwind CSS slate-600; no Material Theme basis.
- BaseBg #212121 — NEVER RENDERED. The identity color is absent from all renderers.

---

## Missing Roles (with renderer-use assessment)

| Role                | Tokyo-N hex | Catpp-M hex | Matl-D hex | Renderer use?        |
|---------------------|-------------|-------------|------------|----------------------|
| BaseBg              | #1a1b26     | #1e1e2e     | #212121    | HIGH — pane fill bg  |
| Fg0 (default text)  | #a9b1d6     | #cdd6f4     | #eeffff    | LOW — terminal default |
| SelectionBg         | #364a82     | #585b70     | #404040    | CursorHighlight already covers this |
| Identity accent     | —           | #b4befe     | —          | FocusBorder (mocha)  |
| Comment/invisible   | #565f89     | #9399b2     | #616161    | SyntaxNull/LevelDebug could use |
| Operator color      | #89ddff     | #89dceb     | #89ddff    | LOW — no operator syntax in log JSON |

**BaseBg is the highest-value missing role**: adding it would let each theme paint its characteristic background on panes, solving the "themes look the same" problem at the most fundamental level.

---

## Pitfalls

| Pitfall                                | Severity | Notes                                   |
|----------------------------------------|----------|-----------------------------------------|
| No BaseBg token — terminal bg dominates | HIGH     | Most visible: tokyo-night navy never appears |
| Convergent semantic assignments        | HIGH     | red=error, blue=info across all themes  |
| Aliased tokens collapse per-theme diversity | MEDIUM | SearchHighlight==SyntaxNumber; FocusBorder==LevelInfo |
| Catppuccin FocusBorder uses Blue not Lavender | MEDIUM | Lavender is catppuccin's identity accent; Blue makes it look like tokyo-night |
| Material LevelDebug/SyntaxNull/Dim invented | MEDIUM | No upstream basis; breaks palette fidelity |
| UnfocusedBg too close to terminal default | MEDIUM | Near-black values may be invisible on default dark terminals |
| Truecolor assumed; no 256-color fallback | LOW      | lipgloss handles hex directly; subtle hue differences collapse in 256-color |
| Material Theme now deprecated (2024)    | LOW      | Upstream is Vira Theme; community fork preserves original palette |

---

## Implications for Design

### Theme struct changes

1. **Add `BaseBg lipgloss.Color`** — the primary editor/pane background. Renderers paint this on full-pane Background(). This is the single highest-ROI change.
2. **Consider `AccentColor lipgloss.Color`** — each theme's identity accent (lavender for catppuccin, navy-tinted for tokyo-night, electric-cyan-already-present for material). Currently FocusBorder doubles as both focus signal and accent; a separate AccentColor breaks that coupling.
3. **No other new fields immediately needed** — all current 19 color fields are consumed. New roles must be justified by a renderer that will use them.

### Constructor changes

1. **Introduce intermediate palette constants** in each constructor:
   ```go
   // palette constants — map canonical names to hex
   var mocha = struct{ base, mantle, crust, lavender, blue string }{
       "#1e1e2e", "#181825", "#11111b", "#b4befe", "#89b4fa",
   }
   // then assign semantic roles by name:
   FocusBorder: lipgloss.Color(mocha.lavender),  // was: mocha.blue
   ```
   This makes drift visible at review time and makes canonical sources auditable.
2. **Fix catppuccin FocusBorder**: change from Blue (#89b4fa) to Lavender (#b4befe).
3. **Fix material-dark LevelDebug/SyntaxNull**: change from violet-grey (#676e95) to canonical comment grey (#616161) or a grey that tracks the palette.
4. **Fix material-dark Dim**: change from invented violet (#4a4a6a) to a closer Material palette grey.
5. **Fix material-dark CursorHighlight**: replace Tailwind import (#4a5568) with a Material palette color.

### cavekit-config R4 changes

1. Add `BaseBg` to the required token list in R4.
2. Add AC: `BaseBg must be distinct from UnfocusedBg` (prevent them collapsing to the same shade).
3. Add AC: each theme's `BaseBg` matches the canonical editor background for that theme.
4. Consider adding a human sign-off AC: "themes are perceptibly distinct at a glance" (complementary to the existing per-theme sign-off).

---

## Open Questions

1. **Should gloggy paint BaseBg on pane backgrounds?** Doing so would override the user's terminal background choice — a UX tradeoff. Could be opt-in via config.
2. **Which tokyo-night variant is canonical for gloggy?** Current choice is ambiguous: accent colors = night, HeaderBg = storm. Document the choice explicitly.
3. **Should catppuccin FocusBorder change to Lavender?** This is a visible change that requires a human sign-off (R4 AC 5). Worth doing as part of the theme-palettes kit.
4. **Material Theme is deprecated** — should gloggy's comment URL be updated to the community fork or Vira Theme successor?
5. **SyntaxNull vs LevelDebug aliasing** — intentional design choice or accidental? If the null/undefined visual role should be distinct from a "debug badge" role, they need different tokens.

---

## Sources

1. `/home/isto/Projects/Nikula/gloggy/internal/theme/theme.go` — current Theme struct and all three constructors
2. `/home/isto/Projects/Nikula/gloggy/internal/theme/theme_test.go` — invariant tests including T-175 luminance ordering
3. `/home/isto/Projects/Nikula/gloggy/context/kits/cavekit-config.md` R4 — theme requirement text
4. https://github.com/tokyo-night/tokyo-night-vscode-theme — canonical tokyo-night VSCode theme
5. https://github.com/tokyo-night/tokyo-night-vscode-theme/blob/master/README.md — variant comparison (night/storm/light bg values)
6. https://github.com/tokyo-night/tokyo-night-vscode-theme/blob/master/themes/tokyo-night-storm-color-theme.json — storm variant bg values
7. https://github.com/catppuccin/palette — catppuccin mocha palette.json (all 26 named colors)
8. https://github.com/catppuccin/catppuccin — catppuccin design philosophy
9. https://raw.githubusercontent.com/catppuccin/catppuccin/main/docs/style-guide.md — official semantic role mappings
10. https://marketplace.visualstudio.com/items?itemName=Equinusocio.vsc-material-theme — Material Theme (deprecated notice, confirming Astorino origin)
11. https://github.com/myambitions/vsc-community-material-theme — community-maintained fork with legacy palette
12. https://dandavison.github.io/delta/custom-themes.html — delta theme structure (feature-based)
13. https://jvns.ca/blog/2024/10/01/terminal-colours/ — terminal color pitfalls (Oct 2024)
14. https://github.com/chriskempson/base16/blob/master/styling.md — base16 role definitions (base00-base0F)

Raw findings:
- `context/refs/research-theme-palettes/codebase.md`
- `context/refs/research-theme-palettes/tokyo-night.md`
- `context/refs/research-theme-palettes/catppuccin-mocha.md`
- `context/refs/research-theme-palettes/material-dark.md`
- `context/refs/research-theme-palettes/tui-fidelity-practices.md`
