package connectivity

import (
	"testing"
	"time"
)

func uptimeObservation(id, status string, finished time.Time) Attempt {
	return Attempt{ID: id, StartedAt: finished.Add(-time.Second), FinishedAt: &finished, Outcome: Outcome{Status: status}}
}

func TestHostTimelineCompletionExpiryAndGaps(t *testing.T) {
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name, status, want string
		age                time.Duration
	}{
		{"connected", "connected", "operational", time.Minute},
		{"failed", "failed", "outage", time.Minute},
		{"unsupported", "unsupported", "unknown", time.Minute},
		{"interrupted", "interrupted", "unknown", time.Minute},
		{"checker error", "error", "unknown", time.Minute},
		{"exact expiry", "connected", "unknown", 5 * time.Minute},
		{"completion now", "failed", "outage", 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			at := now.Add(-test.age)
			result := HostTimeline([]Attempt{uptimeObservation("a", test.status, at), {Outcome: Outcome{Status: "running"}}}, now, 5*time.Minute)
			if result.State != test.want {
				t.Fatalf("state=%s want=%s", result.State, test.want)
			}
			if !result.Segments[0].From.Equal(now.Add(-24*time.Hour)) || result.Segments[0].State != "unknown" {
				t.Fatal("leading gap was fabricated")
			}
			if !result.Segments[len(result.Segments)-1].To.Equal(now) {
				t.Fatal("window incomplete")
			}
			for _, segment := range result.Segments {
				if !segment.From.Before(segment.To) {
					t.Fatal("empty/reversed interval")
				}
			}
		})
	}
}

func TestAggregateTimelineUsesSimultaneousAvailability(t *testing.T) {
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	start := now.Add(-10 * time.Minute)
	change := now.Add(-5 * time.Minute)
	left := HostTimeline([]Attempt{uptimeObservation("a", "failed", start), uptimeObservation("b", "connected", change)}, now, time.Hour)
	right := HostTimeline([]Attempt{uptimeObservation("c", "connected", start), uptimeObservation("d", "failed", change)}, now, time.Hour)
	squad := AggregateTimelines([]UptimeTimeline{left, right}, now)
	if squad.State != "partial" {
		t.Fatal("nonconcurrent failures claimed an outage")
	}
	for _, segment := range squad.Segments {
		if segment.State == "outage" {
			t.Fatal("false simultaneous outage")
		}
	}
	for _, test := range []struct {
		states []string
		want   string
	}{
		{[]string{"operational", "operational"}, "operational"},
		{[]string{"outage", "outage"}, "outage"},
		{[]string{"partial", "outage"}, "partial"},
		{[]string{"unknown", "outage"}, "unknown"},
		{[]string{}, "unknown"},
	} {
		if got := aggregateStates(test.states); got != test.want {
			t.Fatalf("%v => %s want %s", test.states, got, test.want)
		}
	}
}

func TestHostTimelineNeutralizesExpiredPreWindowObservation(t *testing.T) {
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	result := HostTimeline([]Attempt{uptimeObservation("a", "connected", now.Add(-25*time.Hour))}, now, time.Minute)
	if len(result.Segments) != 1 || result.State != "unknown" || !result.Segments[0].From.Equal(now.Add(-24*time.Hour)) {
		t.Fatal("expired observation escaped window")
	}
}
