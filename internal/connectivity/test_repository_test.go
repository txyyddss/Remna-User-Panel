package connectivity

import (
	"context"
	"errors"
	"sync"
	"time"
)

type testRepository struct {
	mu       sync.Mutex
	items    []Attempt
	beginErr error
	prunes   int
	recovers int
}

func (r *testRepository) BeginConnectivityAttempt(_ context.Context, item Attempt) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.beginErr != nil {
		return r.beginErr
	}
	r.items = append(r.items, item)
	return nil
}

func (r *testRepository) FinishConnectivityAttempt(_ context.Context, id string, outcome Outcome, at time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for index := range r.items {
		if r.items[index].ID == id {
			r.items[index].Outcome, r.items[index].FinishedAt = outcome, &at
			return nil
		}
	}
	return errors.New("attempt not found")
}

func (r *testRepository) LatestConnectivityAttempts(_ context.Context, hash string, now time.Time) ([]Attempt, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	latest := make(map[string]Attempt)
	for _, item := range r.items {
		if item.ConfigHash == hash && item.HostUUID != "" && item.StartedAt.After(now.Add(-RetentionWindow)) {
			latest[item.HostUUID] = item
		}
	}
	items := make([]Attempt, 0, len(latest))
	for _, item := range latest {
		items = append(items, item)
	}
	return items, nil
}

func (r *testRepository) ConnectivityHistory(context.Context, string, string, int, time.Time) (HistoryPage, error) {
	return HistoryPage{Items: r.attempts()}, nil
}

func (r *testRepository) RecoverConnectivityAttempts(_ context.Context, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recovers++
	for index := range r.items {
		if r.items[index].Status == "running" {
			r.items[index].Outcome = Outcome{Status: "interrupted", ErrorCode: "CONNECTIVITY_INTERRUPTED"}
			r.items[index].FinishedAt = &now
		}
	}
	return nil
}

func (r *testRepository) PruneConnectivityAttempts(_ context.Context, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.prunes++
	items := make([]Attempt, 0, len(r.items))
	for _, item := range r.items {
		if item.StartedAt.After(now.Add(-RetentionWindow)) {
			items = append(items, item)
		}
	}
	r.items = items
	return nil
}

func (r *testRepository) attempts() []Attempt {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]Attempt(nil), r.items...)
}
