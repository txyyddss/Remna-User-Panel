package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/integrations/remnawave"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/upstreamqueue"
)

type rolloverStatsClient struct {
	remnaClient
	start time.Time
	end   time.Time
}

func (c *rolloverStatsClient) GetUserByID(context.Context, int64) (*remnawave.User, error) {
	return &remnawave.User{}, nil
}

func (c *rolloverStatsClient) GetUserStats(_ context.Context, _ int64, start, end time.Time, _ int) (*remnawave.UserStats, error) {
	if start.After(end) {
		return nil, errors.New("stats range is invalid")
	}
	c.start, c.end = start, end
	return &remnawave.UserStats{}, nil
}

func TestRolloverUsageSnapshotKeepsInclusiveFinalDate(t *testing.T) {
	queue, err := upstreamqueue.New(upstreamqueue.Config{Name: "remna-rollover-stats-test", Capacity: 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := queue.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = queue.Shutdown(context.Background()) }()

	client := &rolloverStatsClient{}
	adapter := remnaAdapter{queue: queue, clientFactory: func(context.Context) (remnaClient, error) { return client, nil }}
	moment := time.Date(2026, 8, 19, 10, 30, 0, 0, time.FixedZone("CST", 8*60*60))
	if _, err := adapter.UsageSnapshotForRollover(context.Background(), "7", moment, moment); err != nil {
		t.Fatalf("UsageSnapshotForRollover(): %v", err)
	}
	if !client.start.Equal(moment.UTC()) || !client.end.Equal(moment.UTC()) {
		t.Fatalf("stats range = %s to %s, want %s to %s", client.start, client.end, moment.UTC(), moment.UTC())
	}
}
