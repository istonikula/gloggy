# tui-mcp upstream wishlist

Capability gaps observed while driving gloggy through tui-mcp for §T human-verify rows. Use as PR-description scaffold when forking [tui-mcp](https://github.com/nvms/tui-mcp).

## Gap 1 — `send_mouse` lacks `motion` action

**Current API**
```jsonschema
"action": { "enum": ["press", "release", "scroll"] }
```

**Ask**
Add `"motion"` to the enum. When `action: "motion"` is sent, the tool emits an SGR mouse-motion sequence:

| variant | encoding | when |
|---------|----------|------|
| free motion (no button) | `\x1b[<35;X;YM` | mouse moved while no button held |
| drag motion (left held) | `\x1b[<32;X;YM` | mouse moved while left button held |
| drag motion (middle)    | `\x1b[<33;X;YM` | etc. |
| drag motion (right)     | `\x1b[<34;X;YM` | etc. |

X and Y are 1-indexed terminal cells (per xterm SGR mouse spec).

**Why we need it**
Drag-resize verification on a TUI requires `press → motion(x1,y1) → motion(x2,y2) → … → release`. `press|release` alone leaves the Motion-driven path (the only path that mutates state and persists config) untestable end-to-end through the pty. Forces fallback to in-process unit tests against bubbletea `tea.Program.Update()`, which skip the pty round-trip entirely.

**Concrete bit-ground**
gloggy SPEC.md §T.5 + invariants V17 (ratio clamp), V18 (drag is single-persist on Release), V19 (drag-seam scope). The current model-layer tests (`internal/ui/app/model_test.go::TestModel_T156_*`, `_T164_*`) cover the logic but not the pty wire format.

## Gap 2 — `send_text` does not decode escape sequences

**Current behavior**
`send_text("\x1b[<32;65;11M")` types 15 literal characters: `\`, `x`, `1`, `b`, `[`, `<`, …

**Ask**
Decode common escape escapes (`\x1b`, ``, `\e`, `\n`, `\r`, `\t`) so callers can inject raw control sequences without a separate `send_raw_bytes` tool. Alternatively, expose a new `send_raw_bytes` tool that takes a base64-encoded byte string and writes it verbatim to the pty.

**Why we need it**
A workaround for Gap 1 (Motion) would be to inject the raw SGR escape via `send_text` — but `\x1b` doesn't decode, so bubbletea never sees the ESC byte that anchors the parse. Either fix would unblock raw mouse / function-key / arbitrary-CSI testing.

**Concrete bit-ground**
Attempted during gloggy §T.5 verification 2026-04-21:
1. `send_mouse press (57,10)` — drag starts ✓
2. `send_text "\x1b[<32;65;11M"` — typed 15 literal chars, no Motion event ✗
3. `send_mouse release (57,10)` — drag ends, no save (correct per V18, but Motion never fired)

## Gap 3 — `XDG_CONFIG_HOME` not honored on darwin (out of scope for tui-mcp)

Not a tui-mcp gap, but related friction during isolated-test setup: on macOS, Go's `os.UserConfigDir()` returns `$HOME/Library/Application Support` unconditionally, ignoring `XDG_CONFIG_HOME`. Tests that need an isolated config dir on darwin must either override `$HOME` or back up + restore the real config file. Nothing for tui-mcp to fix; recorded here so future-me doesn't re-discover it during the next harness setup.

## Priorities

If forking and submitting a single PR: **Gap 1 first**. Motion is the load-bearing missing primitive; once it lands, drag-resize, hover, and any pointer-tracking flow becomes verifiable end-to-end. Gap 2 is nice-to-have (general-purpose escape injection) but Motion via `send_mouse` removes the need for it in the drag case.
