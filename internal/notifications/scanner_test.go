package notifications

import (
	"context"
	"math"
	"testing"
	"time"
)

type scanRepositoryStub struct {
	traffic          []TrafficUser
	automatic        []TrafficUser
	automaticHandled bool
}

func (r *scanRepositoryStub) ProcessAutomaticTrafficResetObservation(_ context.Context, _ string, used, limit int64,
	reset string, lastReset *time.Time, _ time.Time) (AutomaticResetResult, error) {
	r.automatic = append(r.automatic, TrafficUser{UsedBytes: used, LimitBytes: limit, ResetStrategy: reset, LastTrafficResetAt: lastReset})
	return AutomaticResetResult{Handled: r.automaticHandled, EventCreated: r.automaticHandled}, nil
}

func (r *scanRepositoryStub) EnqueueExpiryReminderNotifications(context.Context, time.Time) (int, error) {
	return 0, nil
}

func TestScannerUsesStrictAutomaticResetBoundary(t *testing.T) {
	t.Parallel()
	if aboveNinetyNinePercent(990, 1000) {
		t.Fatal("exactly 99% was considered above 99%")
	}
	if !aboveNinetyNinePercent(991, 1000) || !aboveNinetyNinePercent(math.MaxInt64, math.MaxInt64-1) {
		t.Fatal("aboveNinetyNinePercent rejected an eligible value")
	}
	repository := &scanRepositoryStub{automaticHandled: true}
	scanner := NewScanner(repository, trafficRemoteStub{users: []TrafficUser{
		{ID: 1, UsedBytes: 990, LimitBytes: 1000},
		{ID: 2, UsedBytes: 991, LimitBytes: 1000},
	}}, nil)
	if err := scanner.Scan(context.Background(), time.Now().UTC()); err != nil {
		t.Fatalf("Scan(): %v", err)
	}
	if len(repository.automatic) != 1 || len(repository.traffic) != 1 {
		t.Fatalf("automatic/generic calls = %d/%d, want 1/1", len(repository.automatic), len(repository.traffic))
	}
}

func (r *scanRepositoryStub) EnqueueTrafficThresholdNotification(_ context.Context, remoteID string, used, limit int64,
	reset string, lastReset *time.Time, _ time.Time) (bool, error) {
	r.traffic = append(r.traffic, TrafficUser{UsedBytes: used, LimitBytes: limit, ResetStrategy: reset, LastTrafficResetAt: lastReset})
	return remoteID != "", nil
}

type trafficRemoteStub struct{ users []TrafficUser }

func (r trafficRemoteStub) ListNotificationUsers(context.Context, string, int) ([]TrafficUser, *string, bool, error) {
	return r.users, nil, false, nil
}

func TestScannerUsesStrictOverflowSafeTrafficBoundary(t *testing.T) {
	t.Parallel()
	if aboveNinetyPercent(900, 1000) {
		t.Fatal("exactly 90% was considered above 90%")
	}
	if !aboveNinetyPercent(901, 1000) || !aboveNinetyPercent(math.MaxInt64, math.MaxInt64-1) {
		t.Fatal("aboveNinetyPercent rejected an eligible value")
	}
	repository := &scanRepositoryStub{}
	reset := time.Now().UTC()
	scanner := NewScanner(repository, trafficRemoteStub{users: []TrafficUser{
		{ID: 1, UsedBytes: 900, LimitBytes: 1000},
		{ID: 2, UsedBytes: 901, LimitBytes: 1000, ResetStrategy: "NO_RESET", LastTrafficResetAt: &reset},
	}}, nil)
	if err := scanner.Scan(context.Background(), reset); err != nil {
		t.Fatalf("Scan(): %v", err)
	}
	if len(repository.traffic) != 1 || repository.traffic[0].ResetStrategy != "NO_RESET" {
		t.Fatalf("traffic calls = %+v", repository.traffic)
	}
}
