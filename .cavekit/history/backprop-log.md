---
last_edited: "2026-04-19T11:29:08+03:00"
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
- **Fix commit (impl + kit):** `<pending>`
