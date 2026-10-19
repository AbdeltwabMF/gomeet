package schedule

import (
	"testing"
	"time"
)

func TestOccurrence(t *testing.T) {
	t.Parallel()

	base := time.Date(2024, time.March, 5, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		start   time.Time
		after   time.Time
		through time.Time
		want    bool
	}{
		{name: "inside interval", start: base, after: base.Add(-time.Minute), through: base.Add(time.Minute), want: true},
		{name: "at lower bound is excluded", start: base, after: base, through: base.Add(time.Minute), want: false},
		{name: "at upper bound is included", start: base, after: base.Add(-time.Minute), through: base, want: true},
		{name: "after upper bound", start: base.Add(time.Minute), after: base.Add(-time.Minute), through: base, want: false},
		{name: "zero after matches everything up to through", start: base, after: time.Time{}, through: base.Add(time.Minute), want: true},
		{name: "zero after still respects through", start: base.Add(time.Minute), after: time.Time{}, through: base, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := Occurrence(tt.start, tt.after, tt.through); got != tt.want {
				t.Fatalf("Occurrence() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeduper(t *testing.T) {
	t.Parallel()

	d := NewDeduper(4)
	now := time.Now()

	if !d.First("a", now, now) {
		t.Fatal("first occurrence of a should be new")
	}
	if d.First("a", now, now) {
		t.Fatal("repeated occurrence of a should be suppressed")
	}
	if !d.First("b", now, now) {
		t.Fatal("first occurrence of b should be new")
	}
}

func TestDeduperPrunesOldEntries(t *testing.T) {
	t.Parallel()

	d := NewDeduper(2)
	old := time.Now().Add(-48 * time.Hour)
	now := time.Now()

	d.First("old", old, old)
	d.First("new", now, now)
	d.First("third", now, now)

	if !d.First("old", old, now) {
		t.Fatal("pruned entry should be forgotten")
	}
}
