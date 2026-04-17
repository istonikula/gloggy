## jsonlogs.nvim — direct TUI inspection via tui-mcp

**Environment:** nvim on `logs/tiny.log` (~4800 JSONL lines of bubbletea-formatted Spring/Java logs). Activation: `<leader>jl` (Space as leader).

### Activation + initial layout (80x40 portrait)
- On activation, screen splits vertically: **left = source** (raw JSONL with line numbers), **right = preview** (YAML-style `key: value` per line).
- Default ratio appears asymmetric: **left ~14-20 cols, right ~60 cols** (preview dominant) on first activation.
- Preview uses soft indent (`  `), shows all fields one per line, ends with `}` on its own line. Values are plain text, no syntax colors observed in this snapshot (may be theme-dependent).
- `↴` char marks line continuation marker in preview (appears at line ends).
- Confidence: HIGH

### Live sync
- Moving cursor j/k in source pane → preview re-renders with newly-selected entry's fields. **Instant, no flicker.**
- Confidence: HIGH

### Pane navigation
- `Ctrl+w l` (standard nvim window-move) tested → **did not move focus to preview.** Instead the `l` propagated to source pane and scrolled it horizontally. Suggests preview is a "floating" or non-focusable buffer, OR Ctrl+w was consumed elsewhere.
- `Tab` → **closed the preview pane entirely** (layout collapses to single-pane source view). Re-activating with `<leader>jl` re-opens preview.
- So `Tab` ≈ "toggle / close details", not "cycle focus."
- Confidence: HIGH

### Ratio toggle (`f`)
- `f` observed to toggle between **two preset ratios**:
  - Preset A (default on first activation): left ~14 / right ~60 (preview-wide)
  - Preset B (after `f`): left ~63 / right ~16 (source-wide)
- Each `f` press swaps between the two. **No "full maximize" state** seen (preview never went 100%).
- This matches the jsonlogs.nvim README "maximize/restore" wording but observed as a 2-preset cycle, not a 3-state (compact/balanced/max).
- Confidence: HIGH

### Resize behavior on terminal size change (80x40 → 200x50)
- Layout **preserves ratio** — left and right pane both grow proportionally to terminal width.
- Preview remains ~50% of new width (even though at 80x40 it was 75%). So ratio normalizes based on something (may reset on resize, or tracks percentage).
- Lines in preview that exceed pane width **truncate at border** (no wrap). Example: `··logger:·io.awspring.cloud.autoconfigure.config.secretsmanager.SecretsManagerPropertySour` got cut at "Sour" (rest lost).
- Source pane truncates similarly (raw JSONL line chopped at border).
- Confidence: HIGH

### Wide row handling
- **Truncation only.** No soft-wrap, no horizontal scroll indicator, no "..." marker. Content beyond pane width is invisibly cut.
- Scrolling horizontally in source (using `l` key in Normal mode) works because source is a real nvim buffer; preview does NOT seem to horizontal-scroll (remained static).
- No "Enter to expand" tested successfully — `?` and Enter in preview did not surface help/expand.
- Confidence: HIGH (truncation), MEDIUM (expand behavior — not tested)

### Summary of observed UX patterns to consider porting
| Pattern | Status | Design question for gloggy |
|---------|--------|---------------------------|
| Right-side vertical split with live-sync | ✓ good | Adopt |
| YAML-ish one-prop-per-line flat render | ✓ good for scannability | Adopt; gloggy already flat-renders |
| Preset-ratio toggle (2 or 3 states via `f`) | ✓ simple, usable | Adopt; include "maximize preview" and "source-focus" |
| Tab to close pane | ? double-duty | Risky — Tab already means "cycle focus" in many TUIs; consider alt key |
| Hard truncation of wide rows | ✗ user-hostile | Replace with soft-wrap + optional h/l scroll, OR "Enter to expand" modal |
| No nested-JSON folding | neutral | Decide: flatten (jsonlogs) vs tree-fold (tui-tree-widget-style) |
| No indicator of which pane has focus | ✗ | gloggy already has colored border — keep |
