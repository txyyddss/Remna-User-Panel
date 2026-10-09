package connectivity

import (
	"sort"
	"time"
)

// AggregateTimelines requires full coverage and evaluates simultaneous states.
func AggregateTimelines(timelines []UptimeTimeline, now time.Time) UptimeTimeline {
	result := UnknownTimeline(now)
	if len(timelines) == 0 {
		return result
	}
	boundaries := []time.Time{result.From, result.To}
	for _, timeline := range timelines {
		for _, segment := range timeline.Segments {
			if segment.From.After(result.From) && segment.From.Before(result.To) {
				boundaries = append(boundaries, segment.From)
			}
			if segment.To.After(result.From) && segment.To.Before(result.To) {
				boundaries = append(boundaries, segment.To)
			}
		}
	}
	sort.Slice(boundaries, func(i, j int) bool { return boundaries[i].Before(boundaries[j]) })
	positions := make([]int, len(timelines))
	segments := []UptimeSegment{}
	for index := 0; index+1 < len(boundaries); index++ {
		from, to := boundaries[index], boundaries[index+1]
		if !from.Before(to) {
			continue
		}
		working, down, unknown := 0, 0, false
		for child, timeline := range timelines {
			for positions[child] < len(timeline.Segments) && !timeline.Segments[positions[child]].To.After(from) {
				positions[child]++
			}
			if positions[child] >= len(timeline.Segments) {
				unknown = true
				continue
			}
			segment := timeline.Segments[positions[child]]
			if segment.From.After(from) {
				unknown = true
				continue
			}
			switch segment.State {
			case "operational":
				working++
			case "outage":
				down++
			case "partial":
			default:
				unknown = true
			}
		}
		state := "partial"
		if unknown {
			state = "unknown"
		} else if working == len(timelines) {
			state = "operational"
		} else if down == len(timelines) {
			state = "outage"
		}
		segments = appendSegment(segments, from, to, state)
	}
	result.Segments = segments
	current := make([]string, len(timelines))
	for index, timeline := range timelines {
		current[index] = timeline.State
	}
	result.State = aggregateStates(current)
	return result
}

func observationState(status string) string {
	if status == "connected" {
		return "operational"
	}
	if status == "failed" {
		return "outage"
	}
	return "unknown"
}

func aggregateStates(states []string) string {
	working, down := 0, 0
	if len(states) == 0 {
		return "unknown"
	}
	for _, state := range states {
		switch state {
		case "operational":
			working++
		case "outage":
			down++
		case "partial":
		default:
			return "unknown"
		}
	}
	if working == len(states) {
		return "operational"
	}
	if down == len(states) {
		return "outage"
	}
	return "partial"
}
