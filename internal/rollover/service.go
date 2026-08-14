// Package rollover settles unused-traffic credits before a term can reset.
package rollover

import (
	"context"
	"errors"
	"fmt"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/outbox"
	"time"
)

var ErrRemoteUserMissing = errors.New("Remnawave user is missing")

type Repository interface {
	RolloverByPurchase(context.Context, string) (model.PurchaseRollover, error)
	UserForPurchase(context.Context, string) (model.User, error)
	MarkRolloverProcessing(context.Context, string, time.Time) error
	FinalizeRollover(context.Context, string, int64, int64, string, time.Time) (model.PurchaseRollover, error)
}

type usageFinalizer interface {
	FinalizeRolloverUsage(context.Context, string, model.RolloverUsageSummary, string, time.Time) (model.PurchaseRollover, error)
}

type purchaseLoader interface {
	PurchaseByID(context.Context, string) (model.Purchase, error)
}

type Remote interface {
	QuiesceForRollover(context.Context, string) error
	TrafficForRollover(context.Context, string) (limitBytes, usedBytes int64, err error)
}

type DailyUsage struct {
	Date  time.Time
	Bytes int64
}

type UsageSnapshot struct {
	LimitBytes  int64
	Strategy    string
	LastResetAt *time.Time
	Daily       []DailyUsage
}

type usageSnapshotRemote interface {
	UsageSnapshotForRollover(context.Context, string, time.Time, time.Time) (UsageSnapshot, error)
}

type Service struct {
	repository Repository
	remote     Remote
	now        func() time.Time
}

func NewService(repository Repository, remote Remote) *Service {
	return &Service{repository: repository, remote: remote, now: time.Now}
}

func (s *Service) HandleOutbox(ctx context.Context, job model.OutboxJob) error {
	if job.Kind != "rollover_finalize" {
		return fmt.Errorf("unsupported rollover job %q", job.Kind)
	}
	purchaseID, err := outbox.TargetID(job, "purchaseId")
	if err != nil {
		return err
	}
	rollover, err := s.repository.RolloverByPurchase(ctx, purchaseID)
	if err != nil {
		return err
	}
	if rollover.Status == "credited" || rollover.Status == "zero" || rollover.Status == "exception" {
		return nil
	}
	user, err := s.repository.UserForPurchase(ctx, purchaseID)
	if err != nil {
		return err
	}
	if user.RemnaUserID == nil {
		if rollover.Status == "pending" {
			if err := s.repository.MarkRolloverProcessing(ctx, purchaseID, s.now().UTC()); err != nil {
				return err
			}
		}
		_, err := s.repository.FinalizeRollover(ctx, purchaseID, 0, 0, "local_identity_missing", s.now().UTC())
		return err
	}
	if rollover.Status == "pending" {
		if err := s.remote.QuiesceForRollover(ctx, *user.RemnaUserID); err != nil {
			if errors.Is(err, ErrRemoteUserMissing) {
				if markErr := s.repository.MarkRolloverProcessing(ctx, purchaseID, s.now().UTC()); markErr != nil {
					return markErr
				}
				_, finalizeErr := s.repository.FinalizeRollover(ctx, purchaseID, 0, 0, "remnawave_user_missing", s.now().UTC())
				return finalizeErr
			}
			return fmt.Errorf("quiesce rollover: %w", err)
		}
		if err := s.repository.MarkRolloverProcessing(ctx, purchaseID, s.now().UTC()); err != nil {
			return err
		}
	}
	if extended, ok := s.remote.(usageSnapshotRemote); ok {
		if loader, loadOK := s.repository.(purchaseLoader); loadOK {
			purchase, loadErr := loader.PurchaseByID(ctx, purchaseID)
			if loadErr != nil {
				return loadErr
			}
			snapshot, snapshotErr := extended.UsageSnapshotForRollover(ctx, *user.RemnaUserID, purchase.ValidFrom, purchase.ValidUntil)
			if errors.Is(snapshotErr, ErrRemoteUserMissing) {
				_, snapshotErr = s.repository.FinalizeRollover(ctx, purchaseID, 0, 0, "remnawave_user_missing", s.now().UTC())
				return snapshotErr
			}
			if snapshotErr != nil {
				return fmt.Errorf("fetch rollover traffic: %w", snapshotErr)
			}
			if finalizer, finalizerOK := s.repository.(usageFinalizer); finalizerOK {
				summary := CalculateUsage(purchase, rollover.MinimumRemainingBPS, snapshot)
				_, finalErr := finalizer.FinalizeRolloverUsage(ctx, purchaseID, summary, "", s.now().UTC())
				return finalErr
			}
		}
	}
	limit, used, err := s.remote.TrafficForRollover(ctx, *user.RemnaUserID)
	if errors.Is(err, ErrRemoteUserMissing) {
		_, err = s.repository.FinalizeRollover(ctx, purchaseID, 0, 0, "remnawave_user_missing", s.now().UTC())
		return err
	}
	if err != nil {
		return fmt.Errorf("fetch rollover traffic: %w", err)
	}
	_, err = s.repository.FinalizeRollover(ctx, purchaseID, limit, used, "", s.now().UTC())
	return err
}

const UsageAlgorithmVersion = "cadence-v2"

// CalculateUsage derives cadence allowances from daily upstream data without
// retaining the raw provider series.
