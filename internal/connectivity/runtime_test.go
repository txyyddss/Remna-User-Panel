package connectivity

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestSettingChangeCancelsObsoleteSuccessfulResult(t *testing.T) {
	settings, source, repository, cfg := configuredDependencies(t)
	started := make(chan struct{}, 1)
	service := startTestService(t, settings, source, testProbe(func(ctx context.Context, _ Target, _ Config) Outcome {
		started <- struct{}{}
		<-ctx.Done()
		return Outcome{Status: "connected"}
	}), repository)
	if _, err := service.Start(context.Background(), "manual"); err != nil {
		t.Fatal(err)
	}
	waitSignal(t, started)
	cfg.ProbeURL = "https://example.com/generate_204"
	settings.set(t, cfg)
	service.Invalidate()
	waitFor(t, func() bool { items := repository.attempts(); return len(items) == 1 && items[0].FinishedAt != nil })
	if item := repository.attempts()[0]; item.Status != "interrupted" || item.ErrorCode != "CONNECTIVITY_INTERRUPTED" {
		t.Fatalf("obsolete configuration published success: %+v", item)
	}
	snapshot, err := service.Snapshot(context.Background())
	if err != nil || len(snapshot.Hosts) != 1 || snapshot.Hosts[0].Latest != nil {
		t.Fatalf("old result leaked into replacement configuration: %+v %v", snapshot, err)
	}
}

func TestScheduleWaitsFromCompletionAndManualResetsInterval(t *testing.T) {
	settings, source, repository, cfg := configuredDependencies(t)
	cfg.ScheduledEnabled = true
	settings.set(t, cfg)
	base := time.Now().UTC()
	var seconds atomic.Int64
	now := func() time.Time { return base.Add(time.Duration(seconds.Load()) * time.Second) }
	service := startTestServiceClock(t, settings, source, testProbe(func(context.Context, Target, Config) Outcome {
		return Outcome{Status: "connected"}
	}), repository, now)
	waitCompletedRun(t, service, 1, repository)
	seconds.Store(299)
	service.observeConfiguration(context.Background())
	if len(repository.attempts()) != 1 {
		t.Fatal("scheduled batch ran before completion plus interval")
	}
	if _, err := service.Start(context.Background(), "manual"); err != nil {
		t.Fatal(err)
	}
	waitCompletedRun(t, service, 2, repository)
	seconds.Store(300)
	service.observeConfiguration(context.Background())
	if len(repository.attempts()) != 2 {
		t.Fatal("manual completion did not reset scheduled interval")
	}
	seconds.Store(599)
	service.observeConfiguration(context.Background())
	waitCompletedRun(t, service, 3, repository)
	if items := repository.attempts(); items[0].Trigger != "scheduled" || items[1].Trigger != "manual" || items[2].Trigger != "scheduled" {
		t.Fatalf("unexpected trigger sequence: %+v", items)
	}
}

func TestStartupRecoveryAndRetentionRunWhileDisabled(t *testing.T) {
	settings, source, repository, cfg := configuredDependencies(t)
	base := time.Now().UTC()
	hash := ConfigHash(cfg)
	repository.items = []Attempt{
		{ID: "expired", ConfigHash: hash, StartedAt: base.Add(-25 * time.Hour), Outcome: Outcome{Status: "connected"}},
		{ID: "unfinished", ConfigHash: hash, StartedAt: base.Add(-time.Hour), Outcome: Outcome{Status: "running"}},
	}
	service := startTestService(t, settings, source, testProbe(func(context.Context, Target, Config) Outcome {
		t.Error("disabled scheduler executed a probe")
		return Outcome{Status: "error"}
	}), repository)
	items := repository.attempts()
	if len(items) != 1 || items[0].ID != "unfinished" || items[0].Status != "interrupted" || items[0].FinishedAt == nil ||
		!items[0].StartedAt.Equal(base.Add(-time.Hour)) {
		t.Fatalf("recovery extended retention or lost interrupted state: %+v", items)
	}
	service.cleanup(context.Background())
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if repository.prunes != 2 || repository.recovers != 1 {
		t.Fatalf("disabled cleanup not maintained: prunes=%d recovers=%d", repository.prunes, repository.recovers)
	}
}

func waitCompletedRun(t *testing.T, service *Service, count int, repository *testRepository) {
	t.Helper()
	waitFor(t, func() bool {
		items := repository.attempts()
		if len(items) != count || items[count-1].FinishedAt == nil {
			return false
		}
		service.mu.RLock()
		defer service.mu.RUnlock()
		return service.active == nil && service.pending == nil
	})
}
