---
generated: "2026-04-19"
topic: "theme-palettes — codebase findings"
---

# Raw Findings: Gloggy Codebase — Theme Implementation

## Source Files

- `internal/theme/theme.go` — Theme struct + 3 constructors
- `internal/theme/theme_test.go` — invariants
- `context/kits/cavekit-config.md` — R4 requirement text

## Theme Struct: 20 Fields

```
Name string
LevelError, LevelWarn, LevelInfo, LevelDebug      (4 — level badges)
SyntaxKey, SyntaxString, SyntaxNumber,
SyntaxBoolean, SyntaxNull                          (5 — syntax)
Mark, Dim, SearchHighlight                         (3 — UI)
CursorHighlight, HeaderBg, FocusBorder             (3 — visual polish, Tier 8)
DividerColor, UnfocusedBg                          (2 — pane state, Tier 9)
DragHandle                                         (1 — drag seam, Tier 23)
```

Total: 19 color fields + Name = 20 struct fields.

## Renderer Consumer Map

Every field is consumed by at least one renderer (no dead fields):

| Field           | Consumer(s)                                              |
|-----------------|----------------------------------------------------------|
| LevelError      | entrylist/row.go:24 — badge foreground                  |
| LevelWarn       | entrylist/row.go:26, list.go:705 — badge + empty prefix  |
| LevelInfo       | entrylist/row.go:28                                      |
| LevelDebug      | entrylist/row.go:30                                      |
| SyntaxKey       | detailpane/render.go:42,130                              |
| SyntaxString    | detailpane/render.go:95,113                              |
| SyntaxNumber    | detailpane/render.go:99                                  |
| SyntaxBoolean   | detailpane/render.go:102                                 |
| SyntaxNull      | detailpane/render.go:109                                 |
| Mark            | entrylist/list.go:707,709 — mark/reload prefixes         |
| Dim             | detailpane/model.go:316,410; entrylist/row.go:44,98;    |
|                 | appshell/keyhints.go:80                                  |
| SearchHighlight | entrylist/list.go:716; detailpane/search.go:155          |
| CursorHighlight | entrylist/list.go:713; detailpane/model.go:385           |
| HeaderBg        | appshell/header.go:149                                   |
| FocusBorder     | appshell/panestyle.go:41; appshell/keyhints.go:125      |
| DividerColor    | appshell/panestyle.go:32; appshell/layout.go:280         |
| UnfocusedBg     | appshell/panestyle.go:35                                 |
| DragHandle      | appshell/divider.go:25; appshell/panestyle.go:58         |

## cavekit-config.md R4 Summary

R4 defines: level badges x4, syntax x5, mark, dim, search highlight, cursor highlight, header bg, focus border, divider color, unfocused bg, drag handle. DragHandle AC enforces non-empty, != DividerColor, != FocusBorder, mid-tone neutral (human sign-off). Test T-175 pins WCAG luminance ordering: DividerColor < DragHandle < FocusBorder with gaps > 0.02.

## Current Hex Values

### tokyo-night
| Field           | Hex     |
|-----------------|---------|
| LevelError      | #f7768e |
| LevelWarn       | #e0af68 |
| LevelInfo       | #7aa2f7 |
| LevelDebug      | #565f89 |
| SyntaxKey       | #73daca |
| SyntaxString    | #9ece6a |
| SyntaxNumber    | #ff9e64 |
| SyntaxBoolean   | #bb9af7 |
| SyntaxNull      | #565f89 |
| Mark            | #e0af68 |
| Dim             | #414868 |
| SearchHighlight | #ff9e64 |
| CursorHighlight | #364a82 |
| HeaderBg        | #1f2335 |
| FocusBorder     | #7aa2f7 |
| DividerColor    | #3b4261 |
| UnfocusedBg     | #16161e |
| DragHandle      | #5a6475 |

### catppuccin-mocha
| Field           | Hex     |
|-----------------|---------|
| LevelError      | #f38ba8 |
| LevelWarn       | #f9e2af |
| LevelInfo       | #89b4fa |
| LevelDebug      | #6c7086 |
| SyntaxKey       | #94e2d5 |
| SyntaxString    | #a6e3a1 |
| SyntaxNumber    | #fab387 |
| SyntaxBoolean   | #cba6f7 |
| SyntaxNull      | #6c7086 |
| Mark            | #f9e2af |
| Dim             | #45475a |
| SearchHighlight | #fab387 |
| CursorHighlight | #585b70 |
| HeaderBg        | #181825 |
| FocusBorder     | #89b4fa |
| DividerColor    | #313244 |
| UnfocusedBg     | #11111b |
| DragHandle      | #6e7388 |

### material-dark
| Field           | Hex     |
|-----------------|---------|
| LevelError      | #f07178 |
| LevelWarn       | #ffcb6b |
| LevelInfo       | #82aaff |
| LevelDebug      | #676e95 |
| SyntaxKey       | #89ddff |
| SyntaxString    | #c3e88d |
| SyntaxNumber    | #f78c6c |
| SyntaxBoolean   | #c792ea |
| SyntaxNull      | #676e95 |
| Mark            | #ffcb6b |
| Dim             | #4a4a6a |
| SearchHighlight | #f78c6c |
| CursorHighlight | #4a5568 |
| HeaderBg        | #1a1a1a |
| FocusBorder     | #82aaff |
| DividerColor    | #37474f |
| UnfocusedBg     | #0d0d0d |
| DragHandle      | #65737e |

## Observations

1. SyntaxNull == LevelDebug in all three themes (same hex). Not wrong, but limits future differentiation.
2. Mark == LevelWarn in all three themes (same hex). Deliberate (marks are warning-weight) but creates coupling.
3. SearchHighlight == SyntaxNumber in all three themes. Both use the orange/peach token. A distinct hue for search is possible.
4. FocusBorder == LevelInfo in all three themes. Both use the blue token.
5. No surface/bg tokens beyond HeaderBg and UnfocusedBg. No SelectionBg separate from CursorHighlight, no LineHighlight.
6. material-dark: Dim is #4a4a6a (violet-grey tint); tokyo-night Dim is #414868 (cool blue-grey); catppuccin Dim is #45475a (warmer grey). Small but visually present distinction.
