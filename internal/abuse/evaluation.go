package abuse

import (
	"context"
	"sort"
	"time"
)

type evaluationBucket struct {
	userID, reason string
	at             time.Time
	qps, limit     int
	nodes          map[string]bool
}

func (s *Service) evaluateClaim(ctx context.Context, claim EventClaim, policy Policy, rules []compiledRule, whitelist map[string]bool) (EvaluationResult, error) {
	buckets, observations := aggregateEvents(claim.Events, policy, rules, whitelist)
	addLegacyBuckets(buckets, claim.Legacy, policy, rules, whitelist)
	ordered := orderedBuckets(buckets)
	result := EvaluationResult{Rollups: buildRollups(observations)}
	states := map[string]DetectorState{}
	incidents := map[string]Incident{}
	for _, bucket := range ordered {
		stateKey := bucket.userID + "\x00" + bucket.reason
		state, ok := states[stateKey]
		if !ok {
			var err error
			state, err = s.repo.DetectorStateV2(ctx, bucket.userID, bucket.reason)
			if err != nil {
				return EvaluationResult{}, err
			}
		}
		if !state.LastSecond.IsZero() && !bucket.at.After(state.LastSecond) {
			continue
		}
		contiguous := !state.LastSecond.IsZero() && bucket.at.Sub(state.LastSecond) == time.Second
		if !contiguous {
			state.StreakSeconds = 0
			state.IncidentEmitted = false
		}
		if reachesLimit(bucket.qps, bucket.limit) {
			state.StreakSeconds++
			if !state.IncidentEmitted && state.StreakSeconds >= policy.StreakSeconds {
				mergeIncident(incidents, bucket)
				state.IncidentEmitted = true
			}
			if state.IncidentEmitted && state.StreakSeconds > policy.StreakSeconds {
				state.StreakSeconds = policy.StreakSeconds
			}
		} else {
			state.StreakSeconds = 0
			state.IncidentEmitted = false
		}
		state.UserID, state.ReasonName, state.LastSecond = bucket.userID, bucket.reason, bucket.at
		states[stateKey] = state
	}
	for _, state := range states {
		result.States = append(result.States, state)
	}
	for _, incident := range incidents {
		result.Incidents = append(result.Incidents, incident)
	}
	return result, nil
}

func mergeIncident(out map[string]Incident, bucket evaluationBucket) {
	key := bucket.userID + "\x00" + bucket.at.Format(time.RFC3339Nano)
	item := out[key]
	item.UserID, item.OccurredAt = bucket.userID, bucket.at
	if item.MeasuredQPS < bucket.qps {
		item.MeasuredQPS = bucket.qps
	}
	if item.QPSLimit == 0 || bucket.limit < item.QPSLimit {
		item.QPSLimit = bucket.limit
	}
	item.Reasons = unique(append(item.Reasons, bucket.reason))
	for node := range bucket.nodes {
		item.Nodes = append(item.Nodes, node)
	}
	item.Nodes = unique(item.Nodes)
	out[key] = item
}

func orderedBuckets(values map[string]*evaluationBucket) []evaluationBucket {
	out := make([]evaluationBucket, 0, len(values))
	for _, value := range values {
		out = append(out, *value)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].userID != out[j].userID {
			return out[i].userID < out[j].userID
		}
		if out[i].reason != out[j].reason {
			return out[i].reason < out[j].reason
		}
		return out[i].at.Before(out[j].at)
	})
	return out
}
