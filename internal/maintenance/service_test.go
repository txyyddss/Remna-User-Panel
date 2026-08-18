package maintenance

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

func TestRunBacksUpBeforeCompactingOncePerLocalDate(t *testing.T) {
	location := time.FixedZone("CST", 8*60*60)
	repository := &maintenanceRepository{runID: "run-1", acquired: true}
	backup := &maintenanceBackup{run: model.BackupRun{ID: "backup-1", Status: "complete"}}
	service := NewService(repository, backup, location)
	at := time.Date(2026, 8, 18, 17, 0, 0, 0, time.UTC)
	if err := service.Run(context.Background(), at); err != nil {
		t.Fatalf("Run(): %v", err)
	}
	if len(repository.calls) != 1 || repository.calls[0] != "compact" || backup.calls != 1 {
		t.Fatalf("order calls=%v backup=%d", repository.calls, backup.calls)
	}
	if repository.date != "2026-08-19" {
		t.Fatalf("local date=%q", repository.date)
	}
	if repository.backupID != "backup-1" || repository.counts["payments"] != 3 {
		t.Fatalf("completion backup=%q counts=%v", repository.backupID, repository.counts)
	}
}

func TestRunSkipsCompactionWhenBackupFails(t *testing.T) {
	repository := &maintenanceRepository{runID: "run-1", acquired: true}
	backup := &maintenanceBackup{err: errors.New("disk full")}
	service := NewService(repository, backup, time.UTC)
	if err := service.Run(context.Background(), time.Date(2026, 8, 18, 2, 0, 0, 0, time.UTC)); err == nil {
		t.Fatal("Run() unexpectedly succeeded")
	}
	if len(repository.calls) != 0 || repository.completedErr == nil {
		t.Fatalf("calls=%v completedErr=%v", repository.calls, repository.completedErr)
	}
}

func TestRunReturnsWithoutWorkWhenDateLeaseIsHeld(t *testing.T) {
	repository := &maintenanceRepository{acquired: false}
	service := NewService(repository, &maintenanceBackup{}, time.UTC)
	if err := service.Run(context.Background(), time.Now()); err != nil {
		t.Fatalf("Run(): %v", err)
	}
	if repository.claims != 1 {
		t.Fatalf("claims=%d", repository.claims)
	}
}

type maintenanceRepository struct {
	runID         string
	acquired      bool
	claims        int
	date          string
	backupID      string
	counts        map[string]int64
	completedErr  error
	calls         []string
	compactErr    error
	completionErr error
}

func (r *maintenanceRepository) ClaimMaintenanceRun(_ context.Context, date, _ string, _, _ time.Time) (string, bool, error) {
	r.claims++
	r.date = date
	return r.runID, r.acquired, nil
}
func (r *maintenanceRepository) CompactAndPrune(context.Context, time.Time, time.Time, time.Time) (map[string]int64, error) {
	r.calls = append(r.calls, "compact")
	if r.counts == nil {
		r.counts = map[string]int64{"payments": 3}
	}
	return r.counts, r.compactErr
}
func (r *maintenanceRepository) CompleteMaintenanceRun(_ context.Context, _, backupID string, counts map[string]int64, runErr error, _ time.Time) error {
	r.backupID, r.counts, r.completedErr = backupID, counts, runErr
	return r.completionErr
}

type maintenanceBackup struct {
	run   model.BackupRun
	err   error
	calls int
}

func (b *maintenanceBackup) Run(context.Context) (model.BackupRun, error) {
	b.calls++
	return b.run, b.err
}
