---
last_edited: "2026-04-19T11:19:41+03:00"
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
- **Fix commit:** _pending — to be filled after commit_
