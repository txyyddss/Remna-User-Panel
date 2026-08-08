// Package rollover settles unused-traffic credits before a term can reset.
package rollover

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

// ErrRemoteUserMissing is an authoritative upstream 404, not a transient error.
var ErrRemoteUserMissing = errors.New("Remnawave user is missing")

type Repository interface {
	RolloverByPurchase(context.Context, string) (model.PurchaseRollover, error)
	UserForPurchase(context.Context, string) (model.User, error)
	MarkRolloverProcessing(context.Context, string, time.Time) error
	FinalizeRollover(context.Context, string, int64, int64, string, time.Time) (model.PurchaseRollover, error)
}

type Remote interface {
	QuiesceForRollover(context.Context, string) error
	TrafficForRollover(context.Context, string) (limitBytes, usedBytes int64, err error)
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
	rollover, err := s.repository.RolloverByPurchase(ctx, job.AggregateID)
	if err != nil {
		return err
	}
	if rollover.Status == "credited" || rollover.Status == "zero" || rollover.Status == "exception" {
		return nil
	}
	user, err := s.repository.UserForPurchase(ctx, job.AggregateID)
	if err != nil {
		return err
	}
	if user.RemnaUserID == nil {
		if rollover.Status == "pending" {
			if err := s.repository.MarkRolloverProcessing(ctx, job.AggregateID, s.now().UTC()); err != nil {
				return err
			}
		}
		_, err := s.repository.FinalizeRollover(ctx, job.AggregateID, 0, 0, "local_identity_missing", s.now().UTC())
		return err
	}
	if rollover.Status == "pending" {
		if err := s.remote.QuiesceForRollover(ctx, *user.RemnaUserID); err != nil {
			if errors.Is(err, ErrRemoteUserMissing) {
				if markErr := s.repository.MarkRolloverProcessing(ctx, job.AggregateID, s.now().UTC()); markErr != nil {
					return markErr
				}
				_, finalizeErr := s.repository.FinalizeRollover(ctx, job.AggregateID, 0, 0, "remnawave_user_missing", s.now().UTC())
				return finalizeErr
			}
			return fmt.Errorf("quiesce rollover: %w", err)
		}
		if err := s.repository.MarkRolloverProcessing(ctx, job.AggregateID, s.now().UTC()); err != nil {
			return err
		}
	}
	limit, used, err := s.remote.TrafficForRollover(ctx, *user.RemnaUserID)
	if errors.Is(err, ErrRemoteUserMissing) {
		_, err = s.repository.FinalizeRollover(ctx, job.AggregateID, 0, 0, "remnawave_user_missing", s.now().UTC())
		return err
	}
	if err != nil {
		return fmt.Errorf("fetch rollover traffic: %w", err)
	}
	_, err = s.repository.FinalizeRollover(ctx, job.AggregateID, limit, used, "", s.now().UTC())
	return err
}
