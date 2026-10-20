package googlecal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func TestFetchFollowsPagination(t *testing.T) {
	t.Parallel()

	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++

		page := map[string]any{}
		switch r.URL.Query().Get("pageToken") {
		case "":
			page["items"] = []map[string]any{{"id": "one", "summary": "First"}}
			page["nextPageToken"] = "second"
		default:
			page["items"] = []map[string]any{{"id": "two", "summary": "Second"}}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(page)
	}))
	defer server.Close()

	srv, err := calendar.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	events, err := Fetch(context.Background(), srv, time.Now())
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("Fetch() returned %d events, want 2", len(events))
	}
	if requests != 2 {
		t.Fatalf("Fetch() made %d requests, want 2", requests)
	}
}

func TestFetchUsesLocalDayBoundary(t *testing.T) {
	t.Parallel()

	var timeMin, timeMax string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeMin = r.URL.Query().Get("timeMin")
		timeMax = r.URL.Query().Get("timeMax")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"items": []any{}})
	}))
	defer server.Close()

	srv, err := calendar.NewService(context.Background(), option.WithEndpoint(server.URL), option.WithoutAuthentication())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	loc := time.FixedZone("UTC+3", 3*60*60)
	now := time.Date(2024, time.March, 5, 22, 30, 0, 0, loc)

	if _, err := Fetch(context.Background(), srv, now); err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	wantMax := time.Date(2024, time.March, 6, 0, 0, 0, 0, loc).Format(time.RFC3339)
	if timeMax != wantMax {
		t.Fatalf("timeMax = %q, want %q", timeMax, wantMax)
	}
	if timeMin != now.Add(-startupGrace).Format(time.RFC3339) {
		t.Fatalf("timeMin = %q, want %q", timeMin, now.Add(-startupGrace).Format(time.RFC3339))
	}
}

func TestEventStart(t *testing.T) {
	t.Parallel()

	if _, ok := eventStart(nil); ok {
		t.Fatal("nil event must not match")
	}
	if _, ok := eventStart(&calendar.Event{Id: "x"}); ok {
		t.Fatal("event without start must not match")
	}
	if _, ok := eventStart(&calendar.Event{Start: &calendar.EventDateTime{Date: "2024-03-05"}}); ok {
		t.Fatal("all-day event must not match")
	}

	e := &calendar.Event{Start: &calendar.EventDateTime{DateTime: "2024-03-05T09:30:00+03:00"}}
	start, ok := eventStart(e)
	if !ok {
		t.Fatal("timed event must match")
	}
	if want := time.Date(2024, time.March, 5, 9, 30, 0, 0, time.FixedZone("UTC+3", 3*60*60)); !start.Equal(want) {
		t.Fatalf("eventStart() = %v, want %v", start, want)
	}
}

func TestDue(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, time.March, 5, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name  string
		start time.Time
		want  bool
	}{
		{name: "future occurrence", start: now.Add(time.Minute), want: false},
		{name: "starting now", start: now, want: true},
		{name: "recently started", start: now.Add(-time.Minute), want: true},
		{name: "within the grace period", start: now.Add(-startupGrace + time.Second), want: true},
		{name: "outside the grace period", start: now.Add(-startupGrace - time.Minute), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := due(tt.start, now); got != tt.want {
				t.Fatalf("due() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeEventsKeepsLateArrivingEvents(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, time.March, 5, 10, 0, 0, 0, time.UTC)
	known := make(map[string]*calendar.Event)

	// An event that has just started and only shows up in the second fetch.
	late := &calendar.Event{
		Id:       "late",
		Summary:  "Late",
		Location: "https://meet.example.com/late",
		Start:    &calendar.EventDateTime{DateTime: now.Add(-30 * time.Second).Format(time.RFC3339)},
	}

	mergeEvents(known, []*calendar.Event{late}, now)
	if len(known) != 1 {
		t.Fatalf("known contains %d events, want 1", len(known))
	}

	for key, e := range known {
		start, ok := eventStart(e)
		if !ok {
			t.Fatal("event must have a start time")
		}
		if !due(start, now) {
			t.Fatalf("event %q must still be due", key)
		}
	}
}

func TestMergeEventsDropsOldEvents(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, time.March, 5, 10, 0, 0, 0, time.UTC)
	known := make(map[string]*calendar.Event)

	old := &calendar.Event{
		Id:       "old",
		Summary:  "Old",
		Location: "https://meet.example.com/old",
		Start:    &calendar.EventDateTime{DateTime: now.Add(-2 * time.Hour).Format(time.RFC3339)},
	}

	mergeEvents(known, []*calendar.Event{old}, now)
	if len(known) != 0 {
		t.Fatalf("known contains %d events, want 0", len(known))
	}
}

func TestMergeEventsDropsRemovedEvents(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, time.March, 5, 10, 0, 0, 0, time.UTC)
	event := &calendar.Event{
		Id:    "canceled",
		Start: &calendar.EventDateTime{DateTime: now.Add(time.Minute).Format(time.RFC3339)},
	}
	known := map[string]*calendar.Event{eventKey(event, now.Add(time.Minute)): event}

	mergeEvents(known, nil, now)
	if len(known) != 0 {
		t.Fatalf("known contains %d events after removal, want 0", len(known))
	}
}

func TestEventKeyDistinguishesOccurrences(t *testing.T) {
	t.Parallel()

	e := &calendar.Event{Id: "id", Location: "https://meet.example.com/room"}
	first := time.Date(2024, time.March, 5, 9, 30, 0, 0, time.UTC)
	second := first.AddDate(0, 0, 7)

	if eventKey(e, first) == eventKey(e, second) {
		t.Fatal("keys for different occurrences must differ")
	}
	if eventKey(e, first) != eventKey(e, first) {
		t.Fatal("keys for the same occurrence must match")
	}
}
