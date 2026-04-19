---
last_edited: "2026-04-19T11:35:46+03:00"
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
- **Fix commit (impl + kit):** `<pending>`
