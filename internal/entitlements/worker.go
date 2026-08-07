// Package entitlements applies durable purchase state to Remnawave.
package entitlements

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

// Repository is the durable outbox and entitlement query surface.
type Repository interface {
	ClaimOutboxJob(context.Context, time.Time) (*model.OutboxJob, error)
	CompleteOutboxJob(context.Context, string, int, error, time.Time) error
	PurchaseByID(context.Context, string) (model.Purchase, error)
	UserForPurchase(context.Context, string) (model.User, error)
	UserByID(context.Context, string) (model.User, error)
	DesiredEntitlement(context.Context, string, time.Time) (*model.Purchase, error)
	ClaimPurchaseTrafficReset(context.Context, string, time.Time) (bool, error)
	MarkPurchaseSyncResult(context.Context, string, bool, time.Time) error
	ExpirePurchase(context.Context, string, time.Time) error
}

// RemnawaveClient atomically replaces the complete entitlement view upstream.
type RemnawaveClient interface {
	ApplyEntitlement(ctx context.Context, remoteUserID string, trafficLimitBytes int64, resetStrategy string, squadUUIDs []string, resetTraffic bool) error
	RemoveEntitlement(ctx context.Context, remoteUserID string) error
}

// Worker processes retryable external synchronization one job at a time.
type Worker struct {
	repository Repository
	remnawave  RemnawaveClient
	now        func() time.Time
}

// NewWorker creates an entitlement worker.
func NewWorker(repository Repository, remnawave RemnawaveClient) *Worker {
	return &Worker{repository: repository, remnawave: remnawave, now: time.Now}
}

// Drain processes up to limit ready jobs and returns the first infrastructure error.
func (w *Worker) Drain(ctx context.Context, limit int) error {
	if limit <= 0 {
		limit = 20
	}
	for range limit {
		if err := ctx.Err(); err != nil {
			return err
		}
		job, err := w.repository.ClaimOutboxJob(ctx, w.now().UTC())
		if err != nil {
			return err
		}
		if job == nil {
			return nil
		}
		jobErr := w.process(ctx, *job)
		if jobErr != nil && job.Attempts >= 10 && job.Kind == "remna_apply_entitlement" {
			_ = w.repository.MarkPurchaseSyncResult(ctx, job.AggregateID, false, w.now().UTC())
		}
		if err := w.repository.CompleteOutboxJob(ctx, job.ID, job.Attempts, jobErr, w.now().UTC()); err != nil {
			return err
		}
	}
	return nil
}

func (w *Worker) process(ctx context.Context, job model.OutboxJob) error {
	switch job.Kind {
	case "remna_apply_entitlement":
		purchase, err := w.repository.PurchaseByID(ctx, job.AggregateID)
		if err != nil {
			return err
		}
		if purchase.Status == "cancelled" || purchase.Status == "expired" {
			return nil
		}
		if !w.now().UTC().Before(purchase.ValidUntil) {
			return w.repository.ExpirePurchase(ctx, purchase.ID, w.now().UTC())
		}
		user, err := w.repository.UserForPurchase(ctx, purchase.ID)
		if err != nil {
			return err
		}
		if user.RemnaUserID == nil {
			return errors.New("user has no Remnawave identity")
		}
		resetTraffic, err := w.repository.ClaimPurchaseTrafficReset(ctx, purchase.ID, w.now().UTC())
		if err != nil {
			return err
		}
		if err := w.remnawave.ApplyEntitlement(ctx, *user.RemnaUserID, purchase.TrafficLimitBytes, purchase.ResetStrategy, purchase.SquadUUIDs, resetTraffic); err != nil {
			return fmt.Errorf("apply Remnawave entitlement: %w", err)
		}
		return w.repository.MarkPurchaseSyncResult(ctx, purchase.ID, true, w.now().UTC())
	case "remna_sync_user":
		user, err := w.repository.UserByID(ctx, job.AggregateID)
		if err != nil {
			return err
		}
		if user.RemnaUserID == nil {
			return nil
		}
		desired, err := w.repository.DesiredEntitlement(ctx, user.ID, w.now().UTC())
		if err != nil {
			return err
		}
		if desired == nil {
			return w.remnawave.RemoveEntitlement(ctx, *user.RemnaUserID)
		}
		return w.remnawave.ApplyEntitlement(ctx, *user.RemnaUserID, desired.TrafficLimitBytes, desired.ResetStrategy, desired.SquadUUIDs, false)
	default:
		return fmt.Errorf("unknown outbox job kind %q", job.Kind)
	}
}
