package localcal

import (
	"testing"
	"time"

	"github.com/AbdeltwabMF/gomeet/configs"
)

func TestOccurrencesWithinInterval(t *testing.T) {
	t.Parallel()

	loc := time.UTC
	event := &configs.Event{
		Summary: "Standup",
		URL:     "https://meet.example.com/standup",
		Start:   configs.Start{Time: "09:30", Days: []string{"Monday"}},
	}

	monday := time.Date(2024, time.March, 4, 0, 0, 0, 0, loc)
	tuesday := time.Date(2024, time.March, 5, 0, 0, 0, 0, loc)

	got := occurrences(event, monday, tuesday, loc)
	if len(got) != 1 {
		t.Fatalf("occurrences() = %v, want one occurrence", got)
	}
	if want := time.Date(2024, time.March, 4, 9, 30, 0, 0, loc); !got[0].Equal(want) {
		t.Fatalf("occurrences() = %v, want %v", got[0], want)
	}
}

func TestOccurrencesExcludesAlreadyHandledWindow(t *testing.T) {
	t.Parallel()

	loc := time.UTC
	event := &configs.Event{
		Summary: "Standup",
		URL:     "https://meet.example.com/standup",
		Start:   configs.Start{Time: "09:30", Days: []string{"Monday"}},
	}

	monday := time.Date(2024, time.March, 4, 0, 0, 0, 0, loc)
	start := time.Date(2024, time.March, 4, 10, 0, 0, 0, loc)

	if got := occurrences(event, start, monday.AddDate(0, 0, 1), loc); len(got) != 0 {
		t.Fatalf("occurrences() = %v, want none", got)
	}
}

func TestOccurrencesIgnoresUnrelatedWeekday(t *testing.T) {
	t.Parallel()

	loc := time.UTC
	event := &configs.Event{
		Summary: "Standup",
		URL:     "https://meet.example.com/standup",
		Start:   configs.Start{Time: "09:30", Days: []string{"Monday"}},
	}

	tuesday := time.Date(2024, time.March, 5, 0, 0, 0, 0, loc)
	if got := occurrences(event, tuesday, tuesday.AddDate(0, 0, 1), loc); len(got) != 0 {
		t.Fatalf("occurrences() = %v, want none", got)
	}
}

func TestOccurrencesFindsEveryDayInWindow(t *testing.T) {
	t.Parallel()

	loc := time.UTC
	event := &configs.Event{
		Summary: "Daily",
		URL:     "https://meet.example.com/daily",
		Start:   configs.Start{Time: "09:30", Days: []string{"Monday", "Tuesday", "Wednesday"}},
	}

	monday := time.Date(2024, time.March, 4, 0, 0, 0, 0, loc)
	thursday := time.Date(2024, time.March, 7, 0, 0, 0, 0, loc)

	got := occurrences(event, monday, thursday, loc)
	if len(got) != 3 {
		t.Fatalf("occurrences() returned %d occurrences, want 3: %v", len(got), got)
	}
}

func TestParseClock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value     string
		wantHour  int
		wantMin   int
		wantValid bool
	}{
		{value: "09:30", wantHour: 9, wantMin: 30, wantValid: true},
		{value: "23:59", wantHour: 23, wantMin: 59, wantValid: true},
		{value: "9:30", wantValid: false},
		{value: "24:00", wantValid: false},
		{value: "12:60", wantValid: false},
		{value: "ab:cd", wantValid: false},
		{value: "", wantValid: false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()

			hour, minute, ok := parseClock(tt.value)
			if ok != tt.wantValid {
				t.Fatalf("parseClock(%q) valid = %v, want %v", tt.value, ok, tt.wantValid)
			}
			if ok && (hour != tt.wantHour || minute != tt.wantMin) {
				t.Fatalf("parseClock(%q) = %d:%d, want %d:%d", tt.value, hour, minute, tt.wantHour, tt.wantMin)
			}
		})
	}
}

func TestLocalKeyIsStablePerOccurrence(t *testing.T) {
	t.Parallel()

	event := &configs.Event{
		Summary: "Standup",
		URL:     "https://meet.example.com/standup",
		Start:   configs.Start{Time: "09:30", Days: []string{"Monday"}},
	}

	first := time.Date(2024, time.March, 4, 9, 30, 0, 0, time.UTC)
	second := first.AddDate(0, 0, 7)

	if localKey(event, first) == localKey(event, second) {
		t.Fatal("keys for different occurrences must differ")
	}
	if localKey(event, first) != localKey(event, first) {
		t.Fatal("keys for the same occurrence must match")
	}
}
