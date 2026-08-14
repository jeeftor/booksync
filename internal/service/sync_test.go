package service

import "testing"

func TestDecideSyncDirectionKindleAhead(t *testing.T) {
	d := decideSyncDirection(80, 20)
	if d.Direction != "kindle_to_abs" {
		t.Fatalf("expected kindle_to_abs, got %s", d.Direction)
	}
	if d.ToPct != 80 {
		t.Errorf("expected ToPct 80, got %v", d.ToPct)
	}
}

func TestDecideSyncDirectionABSAhead(t *testing.T) {
	d := decideSyncDirection(20, 80)
	if d.Direction != "abs_ahead_no_kindle_write" {
		t.Fatalf("expected abs_ahead_no_kindle_write, got %s", d.Direction)
	}
	if d.ToPct != 80 {
		t.Errorf("expected ToPct 80, got %v", d.ToPct)
	}
	if d.Message == "" {
		t.Error("expected an explanatory message when abs is ahead but can't be written back")
	}
}

func TestDecideSyncDirectionEqual(t *testing.T) {
	d := decideSyncDirection(50, 50)
	if d.Direction != "noop" {
		t.Fatalf("expected noop for equal progress, got %s", d.Direction)
	}
	if d.ToPct != 50 {
		t.Errorf("expected ToPct 50, got %v", d.ToPct)
	}
}

func TestDecideSyncDirectionWithinEpsilonIsNoop(t *testing.T) {
	// progressEpsilon is 0.5; a 0.3-point difference should not trigger a sync
	// (Amazon/ABS progress values jitter by fractions of a percent).
	d := decideSyncDirection(50.3, 50.0)
	if d.Direction != "noop" {
		t.Fatalf("expected noop within epsilon, got %s", d.Direction)
	}
}

func TestDecideSyncDirectionJustOverEpsilon(t *testing.T) {
	d := decideSyncDirection(50.6, 50.0)
	if d.Direction != "kindle_to_abs" {
		t.Fatalf("expected kindle_to_abs just over epsilon, got %s", d.Direction)
	}
}

func TestDecideSyncDirectionBoundaryIsNotStrictlyGreater(t *testing.T) {
	// A difference exactly equal to progressEpsilon should not trigger a sync
	// (the comparison is strictly greater-than).
	d := decideSyncDirection(50.5, 50.0)
	if d.Direction != "noop" {
		t.Fatalf("expected noop exactly at epsilon boundary, got %s", d.Direction)
	}
}
