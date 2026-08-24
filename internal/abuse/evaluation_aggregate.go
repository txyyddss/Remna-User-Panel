package abuse

import "time"

type observationKey struct {
	userID string
	at     time.Time
}

func aggregateEvents(events []LogEvent, policy Policy, rules []compiledRule, whitelist map[string]bool) (map[string]*evaluationBucket, map[observationKey]int) {
	buckets := map[string]*evaluationBucket{}
	observations := map[observationKey]int{}
	seen := map[string]bool{}
	for _, event := range events {
		if whitelist[event.UserID] {
			continue
		}
		uniqueKey := event.UserID + "\x00" + event.EventSecond.Format(time.RFC3339Nano) + "\x00" + event.Fingerprint
		if seen[uniqueKey] {
			continue
		}
		seen[uniqueKey] = true
		observations[observationKey{userID: event.UserID, at: event.EventSecond}]++
		if policy.GlobalEnabled && policy.GlobalLimit > 0 {
			addEventBucket(buckets, event, "global", policy.GlobalLimit, 1)
		}
		for _, rule := range rules {
			if rule.regex.MatchString(event.Domain) {
				addEventBucket(buckets, event, rule.Name, rule.QPSLimit, 1)
			}
		}
	}
	return buckets, observations
}

func addEventBucket(out map[string]*evaluationBucket, event LogEvent, reason string, limit, count int) {
	key := event.UserID + "\x00" + reason + "\x00" + event.EventSecond.Format(time.RFC3339Nano)
	item := out[key]
	if item == nil {
		item = &evaluationBucket{userID: event.UserID, reason: reason, at: event.EventSecond, limit: limit, nodes: map[string]bool{}}
		out[key] = item
	}
	item.qps += count
	item.limit = limit
	item.nodes[event.NodeUUID] = true
}

func addLegacyBuckets(out map[string]*evaluationBucket, samples []Sample, policy Policy, rules []compiledRule, whitelist map[string]bool) {
	limits := map[string]int{}
	if policy.GlobalEnabled && policy.GlobalLimit > 0 {
		limits["global"] = policy.GlobalLimit
	}
	for _, rule := range rules {
		limits[rule.Name] = rule.QPSLimit
	}
	for _, sample := range samples {
		if whitelist[sample.UserID] {
			continue
		}
		limit, enabled := limits[sample.ReasonName]
		if !enabled {
			continue
		}
		event := LogEvent{UserID: sample.UserID, NodeUUID: sample.NodeUUID, EventSecond: sample.BucketAt}
		addEventBucket(out, event, sample.ReasonName, limit, sample.Count)
	}
}

func buildRollups(observations map[observationKey]int) []QPSRollup {
	windows := map[time.Time]QPSRollup{}
	for key, qps := range observations {
		window := key.at.UTC().Truncate(30 * time.Minute)
		item := windows[window]
		item.WindowAt = window
		item.ObservationCount++
		item.Sum += qps
		if item.Min == 0 || qps < item.Min {
			item.Min = qps
		}
		if qps > item.Max {
			item.Max = qps
		}
		windows[window] = item
	}
	out := make([]QPSRollup, 0, len(windows))
	for _, item := range windows {
		out = append(out, item)
	}
	return out
}
