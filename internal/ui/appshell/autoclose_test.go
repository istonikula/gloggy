package appshell

import "testing"

// T-091: pane auto-closes when detail width drops below 30 in right-split.
func TestShouldAutoCloseDetail_RightSplit_BelowMinWidth(t *testing.T) {
	// At termWidth=60 with widthRatio=0.30: usable=55, detailW = 55-int(55*0.7)=55-38=17 → < 30 → close.
	l := NewLayout(60, 24, true, 0)
	l.Orientation = OrientationRight
	l.WidthRatio = 0.30
	if !ShouldAutoCloseDetail(l) {
		t.Errorf("expected auto-close at termWidth=60 (detailW=%d), got false", l.DetailContentWidth())
	}
}

// T-091: pane stays open when right-split detail width is >= 30.
func TestShouldAutoCloseDetail_RightSplit_AboveMinWidth(t *testing.T) {
	// At termWidth=140 with widthRatio=0.30: usable=135, detailW = 135-int(135*0.7)=135-94=41 → ≥ 30 → keep open.
	l := NewLayout(140, 24, true, 0)
	l.Orientation = OrientationRight
	l.WidthRatio = 0.30
	if ShouldAutoCloseDetail(l) {
		t.Errorf("expected keep-open at termWidth=140 (detailW=%d), got true", l.DetailContentWidth())
	}
}

// T-091: pane auto-closes when below-mode height drops below 3 rows.
func TestShouldAutoCloseDetail_BelowMode_BelowMinHeight(t *testing.T) {
	l := NewLayout(80, 24, true, 2)
	if !ShouldAutoCloseDetail(l) {
		t.Errorf("expected auto-close at paneHeight=2, got false")
	}
}

// T-091: below-mode at exactly 3 rows stays open.
func TestShouldAutoCloseDetail_BelowMode_AtMinHeight(t *testing.T) {
	l := NewLayout(80, 24, true, 3)
	if ShouldAutoCloseDetail(l) {
		t.Errorf("expected keep-open at paneHeight=3, got true")
	}
}

// T-091: closed pane never reports auto-close.
func TestShouldAutoCloseDetail_ClosedPane(t *testing.T) {
	l := NewLayout(60, 24, false, 0)
	l.Orientation = OrientationRight
	if ShouldAutoCloseDetail(l) {
		t.Errorf("closed pane must not report auto-close")
	}
}
