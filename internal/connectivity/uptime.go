package connectivity

import (
	"sort"
	"time"
)

// UnknownTimeline explicitly represents an unmeasured 24-hour window.
func UnknownTimeline(now time.Time) UptimeTimeline {
	now = now.UTC()
	from := now.Add(-24 * time.Hour)
	return UptimeTimeline{From: from, To: now, State: "unknown", Segments: []UptimeSegment{{From: from, To: now, State: "unknown"}}}
}

func appendSegment(segments []UptimeSegment, from, to time.Time, state string) []UptimeSegment {
	if !from.Before(to) {
		return segments
	}
	if len(segments) > 0 && segments[len(segments)-1].State == state && segments[len(segments)-1].To.Equal(from) {
		segments[len(segments)-1].To = to
		return segments
	}
	return append(segments, UptimeSegment{From: from, To: to, State: state})
}

// HostTimeline uses completion timestamps; a running check does not erase its predecessor.
func HostTimeline(attempts []Attempt, now time.Time, freshness time.Duration) UptimeTimeline {
	result := UnknownTimeline(now)
	items := make([]Attempt, 0, len(attempts))
	for _, attempt := range attempts {
		if attempt.FinishedAt != nil && !attempt.FinishedAt.After(now) {
			items = append(items, attempt)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].FinishedAt.Equal(*items[j].FinishedAt) {
			if items[i].StartedAt.Equal(items[j].StartedAt) {
				return items[i].ID < items[j].ID
			}
			return items[i].StartedAt.Before(items[j].StartedAt)
		}
		return items[i].FinishedAt.Before(*items[j].FinishedAt)
	})
	segments := []UptimeSegment{}
	cursor := result.From
	for index, attempt := range items {
		from := *attempt.FinishedAt
		if from.Before(result.From) {
			from = result.From
		}
		to := now.UTC()
		if index+1 < len(items) {
			to = *items[index+1].FinishedAt
		}
		if !from.Before(to) {
			continue
		}
		segments = appendSegment(segments, cursor, from, "unknown")
		state := observationState(attempt.Status)
		expires := attempt.FinishedAt.Add(freshness)
		if expires.Before(from) {
			expires = from
		}
		if state != "unknown" && expires.Before(to) {
			segments = appendSegment(segments, from, expires, state)
			segments = appendSegment(segments, expires, to, "unknown")
		} else {
			segments = appendSegment(segments, from, to, state)
		}
		cursor = to
	}
	segments = appendSegment(segments, cursor, result.To, "unknown")
	if len(segments) > 0 {
		result.Segments = segments
	}
	if len(items) > 0 {
		latest := items[len(items)-1]
		if latest.FinishedAt.Add(freshness).After(now) {
			result.State = observationState(latest.Status)
		}
	}
	return result
}
