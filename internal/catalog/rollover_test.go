package catalog

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	"github.com/txyyddss/Remna-User-Panel/internal/rollover"
)

func TestRolloverProjectionValidatesOwnerAndUsesQueuedSnapshot(t *testing.T) {
	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	remoteID := "remote-1"
	purchase := model.Purchase{ID: "purchase-1", UserID: "user-1", Status: "active",
		PriceTXBMinor: 10_000, ValidFrom: now.Add(-24 * time.Hour), ValidUntil: now.Add(24 * time.Hour),
		RolloverMaxTXBMinor: 1_000, Price: model.TXBMoney(10_000), RolloverMax: model.TXBMoney(1_000)}
	remote := &rolloverRemote{catalogRemnawave: &catalogRemnawave{}, snapshot: rollover.UsageSnapshot{LimitBytes: 1_000, Strategy: "NO_RESET"}}
	repository := &rolloverRepository{catalogRepository: &catalogRepository{}, purchase: purchase}
	service := NewService(repository, remote, time.Minute)
	service.now = func() time.Time { return now }

	projection, err := service.RolloverProjection(context.Background(), model.User{ID: "user-1", OnboardingState: "complete", RemnaUserID: &remoteID}, purchase.ID)
	if err != nil || projection.PurchaseID != purchase.ID || !remote.start.Equal(purchase.ValidFrom) || !remote.end.Equal(now) {
		t.Fatalf("RolloverProjection() = (%+v, %v), range %s to %s", projection, err, remote.start, remote.end)
	}
	if _, err := service.RolloverProjection(context.Background(), model.User{ID: "other", OnboardingState: "complete", RemnaUserID: &remoteID}, purchase.ID); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("owner check error = %v", err)
	}
	if _, err := service.RolloverProjection(context.Background(), model.User{ID: "user-1", OnboardingState: "complete", RemnaUserID: &remoteID}, "missing"); err == nil {
		t.Fatal("missing purchase unexpectedly succeeded")
	}
}

func TestRolloverProjectionRejectsInactiveAndMissingIdentity(t *testing.T) {
	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	purchase := model.Purchase{ID: "purchase-1", UserID: "user-1", Status: "expired", ValidFrom: now.Add(-time.Hour), ValidUntil: now.Add(time.Hour)}
	repository := &rolloverRepository{catalogRepository: &catalogRepository{}, purchase: purchase}
	service := NewService(repository, &rolloverRemote{catalogRemnawave: &catalogRemnawave{}}, time.Minute)
	service.now = func() time.Time { return now }
	if _, err := service.RolloverProjection(context.Background(), model.User{ID: "user-1", OnboardingState: "complete"}, purchase.ID); !errors.Is(err, ErrRolloverNotActive) {
		t.Fatalf("inactive error = %v", err)
	}
	purchase.Status = "active"
	repository.purchase = purchase
	if _, err := service.RolloverProjection(context.Background(), model.User{ID: "user-1", OnboardingState: "complete"}, purchase.ID); !errors.Is(err, ErrRolloverUnavailable) {
		t.Fatalf("missing identity error = %v", err)
	}
}

type rolloverRepository struct {
	*catalogRepository
	purchase model.Purchase
}

func (r *rolloverRepository) PurchaseByID(_ context.Context, id string) (model.Purchase, error) {
	if r.purchase.ID == "" || id != r.purchase.ID {
		return model.Purchase{}, database.ErrNotFound
	}
	return r.purchase, nil
}

type rolloverRemote struct {
	*catalogRemnawave
	snapshot rollover.UsageSnapshot
	err      error
	start    time.Time
	end      time.Time
}

func (r *rolloverRemote) UsageSnapshotForRollover(_ context.Context, _ string, start, end time.Time) (rollover.UsageSnapshot, error) {
	r.start, r.end = start, end
	return r.snapshot, r.err
}
