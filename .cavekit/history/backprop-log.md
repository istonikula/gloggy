---
last_edited: "2026-04-20T20:19:44+03:00"
---

# Backpropagation Log

Append-only log of `/ck:revise --trace` cycles. Each entry traces a single
failure back to a kit R-ID, classifies the gap, records the regression
test, and links the fix commit. Audit trail for the iteration loop.

---

## #1 — F-132: T-165 tests bypass the guard they claim to exercise (2026-04-19)

- **Failure source:** `/ck:review` Pass 2 (Tier 20 branch review)
- **Failure description:** `TestModel_T165_Drag_Zero{Width,Height}_PreservesRatio` passes for the wrong reason. After the synthetic 0-dim `WindowSizeMsg` auto-closes the pane, the test re-sets `m.draggingDivider = true` but cannot re-open the pane. On the next Motion, `model.go:524` (`if !m.pane.IsOpen()`) short-circuits before the `termW/termH<=0` guard at `model.go:554-556`/`:565-567` is ever reached. Deleting the caller-guard left both tests green.
- **Classification:** `incomplete_criterion`
- **Kit:** `cavekit-app-shell.md` → R15 (degenerate-dim AC, prior text on lines ~202-203)
- **Spec change:** R15 degenerate-dim AC text extended to mandate that the regression test drive the guard with `pane.IsOpen()==true` AND `termDim==0` simultaneously, and that removing the caller-guard must make the test fail.
- **Regression tests:**
  - `internal/ui/app/model_test.go::TestModel_F132_DegenerateDim_Right_GuardFiresWith_PaneOpen`
  - `internal/ui/app/model_test.go::TestModel_F132_DegenerateDim_Below_GuardFiresWith_PaneOpen`
- **Verification:** Tests pass with guards intact (`termW<=0` / `termH<=0` returns at lines 554-556 / 565-567). Tests fail when guards removed: `ratio shadowed from 0.550 to 0.300` (right) and `0.450 to 0.300` (below).
- **Code change:** None — the guards already work correctly. Only the test was fraudulent. Old T-165 tests deleted as superseded.
- **Files touched:**
  - `internal/ui/app/model_test.go` (deleted T-165 tests, added F-132 tests)
  - `context/kits/cavekit-app-shell.md` (R15 degenerate-dim AC text + changelog entry)
- **Pattern category:** test-fidelity / validation-via-wrong-path
- **Fix commit:** `97c1b9b`

---

## #2 — F-133: X-axis inverse-math missing pin test + broken formula (2026-04-19)

- **Failure source:** `/ck:review` Pass 2 (Tier 20 branch review)
- **Failure description:** Cavekit `R15` AC at `cavekit-app-shell.md:199` mandates that BOTH `RatioFromDragY` AND `RatioFromDragX`, when inverted against forward ratio→size math, MUST yield the current ratio when Press lands on the current divider row/col. Only the Y-axis had a regression test (`ratiokeys_test.go::TestRatioFromDragY_PressAtCurrentDividerY_KeepsRatio`). The X-axis math (`detail = termWidth - x - 2` at `ratiokeys.go:124-144`) was off by 3 cells against the renderer-truth divider X established by T-160 — at termWidth=100, ratio=0.55, Press-at-current-X returned 0.589 (drift 0.039, exceeding the RatioStep/2=0.025 tolerance the Y-axis test uses). The author of T-161 had explicitly punted on this in code comments ("X-axis analogue of F-123 is present... left unchanged because the T-104 tests encode the current semantics"). The T-104 mid pin (`x=50, termWidth=100 → 48/95`) encoded the broken formula.
- **Classification:** `incomplete_criterion`
- **Kit:** `cavekit-app-shell.md` → R15 (inverse-math AC, line 199)
- **Spec change:** R15 inverse-math AC text extended to mandate parallel regression tests for BOTH axes and that the X-axis canonical Press column MUST be sourced from `Layout.ListContentWidth()` (renderer-truth per T-160) rather than from the inverse formula itself (which would tautologically agree).
- **Regression tests:**
  - `internal/ui/appshell/ratiokeys_test.go::TestRatioFromDragX_PressAtCurrentDividerX_KeepsRatio` (sweeps presets {0.30, 0.50, 0.55} × termWidth ∈ {80, 100})
- **Verification:** Test fails before fix (5/5 cases, drift up to 0.039 on the buggy `termWidth - x - 2` formula). Test passes after fix (`detail := usable - x` at `ratiokeys.go`). Full suite: 564 passed (was 563).
- **Code change:** `RatioFromDragX` rewritten as the exact inverse of `Layout.DetailContentWidth = usable - ListContentWidth`. T-161 audit caveat block stripped. T-104 `TestRatioFromDragX_Mid` pin updated 48/95 → 45/95 to reflect the corrected formula.
- **Files touched:**
  - `internal/ui/appshell/ratiokeys_test.go` (new pin test + T-104 pin update)
  - `internal/ui/appshell/ratiokeys.go` (formula fix + comment rewrite)
  - `context/kits/cavekit-app-shell.md` (R15 inverse-math AC text + changelog entry)
- **Pattern category:** test-fidelity / parallel-axis-coverage
- **Fix commit (test):** `68d2548`
- **Fix commit (impl + kit):** `fd5a26d`

---

## #3 — F-134: stripAnsi CSI terminator set incomplete (2026-04-19)

- **Failure source:** `/ck:review` Pass 2 (Tier 20 branch review)
- **Failure description:** `stripAnsi` at `internal/ui/appshell/mouse_test.go:228-245` exited escape mode only on terminators `{m, K, H, A, B, C, D, J}`. ECMA-48 defines the CSI final byte as the full range `0x40..0x7e`. The hardcoded subset omitted `f` (HVP), `G` (CHA), `n` (DSR), `c` (DA), `s/u` (save/restore cursor), `h/l` (set/reset mode), `~` (function-key terminator), and others. `stripAnsi` backs `locateGlyphCol`, which the R15 line-198 renderer-truth divider-col assertion depends on. Today lipgloss only emits SGR (`m`) so the bug was latent — but any future styling change emitting non-SGR CSI sequences would silently leak escape bytes into `[]rune(plain)` and corrupt `glyphCol`, giving false-positive divider-col assertions. Same pattern shape as F-132 / F-133: the test agreed with what it tested by accident, not by an enforceable invariant.
- **Classification:** `incomplete_criterion`
- **Kit:** `cavekit-app-shell.md` → R15 (renderer-truth AC, line 198)
- **Spec change:** R15 line-198 AC text extended to mandate that any ANSI-stripping helper handle the full ECMA-48 CSI two-step form (`ESC [ <params/intermediates> <final-byte 0x40..0x7e>`).
- **Regression tests:**
  - `internal/ui/appshell/mouse_test.go::TestStripAnsi_HandlesFullCSIFinalByteRange` (9 sub-cases: HVP, CHA, DECTCEM show/hide, function-key tilde, DSR, DA, save/restore cursor)
- **Verification:** Test fails 9/9 sub-cases against the hardcoded subset (escape sequences leak through, output is empty instead of `"X"`). First fix attempt (`r >= 0x40 && r <= 0x7e` directly in the original two-state machine) regressed save_cursor_s and restore_cursor_u — `[` (0x5b) was being mistaken for a final byte. Second fix introduces a three-state machine (`stPlain` / `stPostEsc` / `stCsiBody`) that consumes the `[` introducer as introducer, then scans to the actual final byte. All 9 sub-cases pass after the state-machine rewrite. Full suite: 574 passed (was 564).
- **Code change:** `stripAnsi` rewritten as a three-state machine. Original two-state machine could not distinguish `[` (CSI introducer, also in 0x40..0x7e) from a final byte without explicit state tracking.
- **Files touched:**
  - `internal/ui/appshell/mouse_test.go` (state-machine rewrite + new regression test, both in same file since `stripAnsi` is test-only)
  - `context/kits/cavekit-app-shell.md` (R15 line-198 AC text + changelog entry)
- **Pattern category:** test-fidelity / latent-fragility-in-test-helper
- **Fix commit (test):** `024d429`
- **Fix commit (impl + kit):** `91df657`

---

## #4 — F-200: WindowSizeMsg handler omits WithBelowMode on auto-orientation flip (2026-04-19)

- **Failure source:** `/ck:check` Tier 23 Pass 2 (inspector peer review)
- **Failure description:** `m.pane.WithBelowMode(...)` was called in `relayout()` at `internal/ui/app/model.go:735` but omitted from the inline pane-wiring chain in the `WindowSizeMsg` handler at `model.go:192-195`. The two paths did similar work — `relayout()` runs on open/close/focus/ratio/drag; the `WindowSizeMsg` branch runs on every terminal resize — but diverged on which pane-local flags they refreshed. With `detail_pane.position = "auto"` and the pane open, a terminal resize that crossed `orientation_threshold_cols` flipped `m.resize.Orientation()` but left `m.pane.belowMode` stale. Right→below: the pane's top border (which IS the R10 drag seam in below-mode) rendered in the pane's focus-state color instead of `DragHandle`. Below→right: the pane kept `belowMode=true`, so the right-mode top border (which should be a regular pane border, not a seam) painted as a spurious `DragHandle` row. `TestModel_OrientationFlip_VerticalSizeTracks` at `model_test.go:1154` exercised the exact flow but did not assert on seam SGR, so the bug shipped green.
- **Classification:** `missing_criterion`
- **Kit:** `cavekit-app-shell.md` → R7 (Terminal Resize Handling)
- **Spec change:** R7 description extended with a final sentence mandating that the resize handler and `relayout()` refresh the same pane-local orientation-dependent flags. New AC 9 appended: "When `detail_pane.position` is 'auto' and the detail pane is open, a WindowSizeMsg that crosses `orientation_threshold_cols` must refresh every pane-local rendering flag that depends on orientation (at minimum the detail pane's below-mode flag that drives the R10 drag-seam top-border paint). Post-flip, the rendered drag-seam SGR at the correct seam location (right: `│` glyph column mid-Y; below: the detail pane's top-border row) must match the NEW orientation's seam contract per R10 AC 10, not the pre-flip state. The regression test must exercise both flip directions (right→below and below→right) with the pane open throughout and assert the rendered SGR at each step."
- **Regression tests:**
  - `internal/ui/app/model_f200_test.go::TestModel_F200_WindowSizeMsg_AutoFlip_RefreshesBelowMode` (below→right→below round-trip with rendered-SGR assertion at each step, using standalone `colorANSIF200` helper + TrueColor profile init)
- **Verification:** Test fails before fix — "after below→right WindowSizeMsg flip, pane.belowMode is stale — pane top border still carries DragHandle SGR `\x1b[38;2;89;100;117m` (tokyo-night DragHandle)". Test passes after fix. Full suite 597/597 (+1 new test) across 11 packages.
- **Code change:** One-line addition to the `WindowSizeMsg` pane chain: `.WithBelowMode(m.resize.Orientation() == appshell.OrientationBelow)` at `internal/ui/app/model.go:194`, mirroring the existing `relayout()` call at `:735`. Kept inline (rather than centralizing via `m = m.relayout()`) because `WindowSizeMsg` does additional work the generic relayout does not — auto-close side-effects + cmd propagation.
- **Files touched:**
  - `internal/ui/app/model_f200_test.go` (new file; regression test + TrueColor init + local colorANSIF200 helper)
  - `context/kits/cavekit-app-shell.md` (R7 description + new AC 9 + changelog entry)
  - `internal/ui/app/model.go` (one-line WithBelowMode addition in WindowSizeMsg handler)
- **Pattern category:** integration / dual-wiring-drift
- **Fix commit (test + kit):** `e1e2f2b`
- **Fix commit (impl):** `40554c3`

---

## #5 — F-201: kit language calls drag seam "shared row" but render is detail-top only (2026-04-19)

- **Failure source:** `/ck:check` Tier 23 Pass 2 (inspector peer review)
- **Failure description:** `cavekit-app-shell.md` R10 AC 10 (line 133), R15 description (line 190), and R15 ACs at lines 200/206/207 described the below-mode drag seam as "the horizontal border row between list and detail pane" or "shared border row". `DESIGN.md` mirrored the drift at §2 token table (line 75), §2 three-new-tokens paragraph (lines 91-94), §4.5 drag-handle seam bullet (lines 202-206), §5.x mouse-drag press-cell (lines 438-441), and §6 pane border matrix (line 539). In reality the render is two adjacent rows — the list pane's own bottom border (rendered by its `PaneStyle` in list-focus color) and the detail pane's own top border (rendered by `PaneStyle` + `WithDragSeamTop` at `internal/ui/appshell/panestyle.go:57-59`). Only the detail pane's top is overridden to `DragHandle`. `MouseRouter.Zone` and the paint both target the detail-pane top only. The inspector flagged the mismatch during /ck:check: the code is correct but the spec language misdescribes the physical composition, so a future reader could mistakenly "fix" the paint to cover both rows and break the R10 focus-indicator + R6 mouse-hit-zone contracts.
- **Classification:** `wrong_criterion`
- **Kit:** `cavekit-app-shell.md` → R10 AC 10; R15 description + ACs at lines 200/206/207
- **Spec change:** All 5 kit locations rewritten from "shared/horizontal border row" → "detail pane's top border row" with explicit note that the list's bottom border is an adjacent, separate row in the list's focus-state color, NOT co-painted. R15 AC 11 (line 206) extended to mandate a pinning test asserting the list pane's bottom row does NOT carry `DragHandle` SGR. Companion edits in DESIGN.md §2 token table + §2 three-new-tokens paragraph + §4.5 drag-handle seam bullet + §5.x mouse-drag press-cell + §6 pane border matrix row.
- **Regression tests:**
  - `internal/ui/appshell/panestyle_test.go::TestPaneStyle_DragSeamOnlyOverridesDetailTop_NotListBottom` — 3 themes × asserts (a) detail pane top row contains `DragHandle` SGR via `WithDragSeamTop` AND (b) list pane bottom row rendered via `PaneStyle` does NOT contain `DragHandle` SGR. Pins the contract implicit in the current paint; a future "fix" making the seam a genuine shared row would fail this test and force re-evaluation against R10/R15.
- **Verification:** Pinning test passes against current code (all 3 themes, 4/4 sub-tests). Full suite 601 passed in 11 packages (was 597; pinning test adds 4 sub-tests under the table-driven loop).
- **Code change:** None — the code at `internal/ui/appshell/panestyle.go:57-59` (`WithDragSeamTop`) was already correct. This is a pure spec-language trace closing the gap between what the spec said and what the render physically did.
- **Files touched:**
  - `internal/ui/appshell/panestyle_test.go` (new pinning test)
  - `context/kits/cavekit-app-shell.md` (R10 AC 10 + R15 description + R15 ACs at 200/206/207 + changelog entry)
  - `DESIGN.md` (§2 table + §2 paragraph + §4.5 + §5.x + §6 table)
  - `context/designs/design-changelog.md` (new entry for §2/§4.5/§6 language pass)
- **Pattern category:** spec-reality-drift / language-precision
- **Fix commit (test):** `5b334f3`
- **Fix commit (kit + DESIGN):** `11c4e6a`

---

## #6 — Tail/follow mode silent after first Write event; initial content never emitted (2026-04-20)

- **Failure source:** user report via `/ck:revise --trace` ("tail/follow mode doesn't work — initial entries appear to render but subsequent appends are invisible")
- **Failure description:** `gloggy -f <file>` on a non-empty file renders nothing on startup; the first filesystem Write event dumps the entire existing content (minus line 1) in one burst, which looked to the user like a normal initial load; every subsequent append produces zero new entries. Two superimposed bugs: (a) `internal/logsource/tail.go` used a single long-lived `bufio.Scanner`; once `scanner.Scan()` returns false on EOF after the first drain, the Scanner is terminal and never emits again, so the 2nd/Nth Write events are swallowed. (b) `internal/ui/app/model.go:146` called `TailFile(ctx, sourceName, 1)` in follow mode, which skipped line 1 permanently AND deferred emission of lines 2..N until the first Write arrived. Unit + integration tests existed (`tail_test.go::TestTailFile_DetectsNewLines`, `tests/integration/tail_test.go::TestTailMode_NewEntriesAppear`) but both used a single append batch with `startLineNum == initialLines`, masking both bugs.
- **Classification:** `incomplete_criterion` (R8 AC1 — only required emission after one batch, not continuously) + `missing_criterion` (R8 AC4 — no coverage of initial content emission; R8 AC5 — no UI-level end-to-end assertion)
- **Kit:** `cavekit-log-source.md` → R8 (Tail Mode)
- **Spec change:** R8 description rewritten from "newly appended lines emitted" to "combined initial-emit + live-append mode, emission persists for the entire session". AC1 tightened to require emission across 1st, 2nd, and Nth Write events (explicit anti-EOF-deaf clause). AC4 added: "initial content emitted starting from line 1 before any subsequent append events". AC5 added: "entries reach `app.Model.entries` via Init/Update path, not just the logsource channel — at least one test drives the model directly across multiple appends".
- **Regression tests:**
  - `internal/logsource/tail_test.go::TestTailFile_MultipleAppendBatches` — 5-line file + 2 separate append batches, both drained and line numbers monotonic. Fails on the Scanner impl (2nd batch never arrives).
  - `internal/logsource/tail_test.go::TestTailFile_EmitsInitialContent` — 5-line file + `startLineNum=0`, no appends, assert all 5 emitted. Fails on the Scanner impl (nothing emitted until Write).
  - `internal/ui/app/tail_e2e_test.go::TestTailE2E_EntryListReceivesMultipleAppends` — drives `app.Model` via `Init`/`Update` across initial + 2 append batches; asserts `m.entries` grows correctly. Fails on the Scanner impl (hangs at 2nd batch).
- **Verification:** All three tests fail before fix (timeouts / panic after 30s). All three pass after fix. Full suite 624/624. Manual verification via `tui-mcp`: launched `gloggy -f` on a 3-line seed file, confirmed 3 initial entries render immediately, then appended three separate batches (2 + 2 + 1 lines) and verified each batch appeared in the entrylist without restart. Final header shows "0/8 entries" across the session.
- **Code change:**
  - `internal/logsource/tail.go` rewritten: persistent `*os.File` across Write events; fresh `bufio.Reader` created per drain (sidesteps `bufio.Reader`'s sticky `io.EOF`); `pending` buffer carries trailing bytes without a newline across Write events so logger half-flushes don't emit truncated entries; initial drain runs before `watcher.Add` to avoid a race on Writes that land between `os.Open` and `Add`.
  - `internal/ui/app/model.go:146` changed `TailFile(ctx, sourceName, 1)` → `TailFile(ctx, sourceName, 0)` so follow mode emits existing content on startup.
  - `internal/ui/app/tail_e2e_test.go` pump harness rewritten as a single-goroutine cmd-chain driver — the naive two-goroutine harness abandoned in-flight blocking `tea.Cmd` calls and lost the channel msg they eventually produced.
- **Files touched:**
  - `internal/logsource/tail.go` (TailFile rewrite)
  - `internal/ui/app/model.go` (Init startLineNum flip)
  - `internal/logsource/tail_test.go` (two new unit regression tests)
  - `internal/ui/app/tail_e2e_test.go` (new e2e regression test + harness)
  - `context/kits/cavekit-log-source.md` (R8 description + AC1 tightening + AC4/AC5 additions + changelog)
  - `context/impl/impl-logsource.md` (T-backprop-R8 row + revision log table)
  - `.cavekit/history/backprop-log.md` (this entry)
- **Pattern category:** integration (logsource lifecycle vs. UI lifecycle; Scanner EOF semantics)
- **Fix commit (tests):** `30f743b`
- **Fix commit (impl + harness):** `f98c116`
