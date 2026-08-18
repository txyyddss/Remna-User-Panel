package entitlements

import (
	"context"
	"errors"
	"fmt"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	jobpayload "github.com/txyyddss/Remna-User-Panel/internal/outbox"
)

func (w *Worker) prepareContinuity(ctx context.Context, job model.OutboxJob) error {
	successorID, err := jobpayload.TargetID(job, "purchaseId")
	if err != nil {
		return err
	}
	current, err := w.repository.ContinuityEntitlement(ctx, successorID, w.now().UTC())
	if err != nil {
		return err
	}
	if current == nil {
		return nil
	}
	user, err := w.repository.UserForPurchase(ctx, current.ID)
	if err != nil {
		return err
	}
	if user.RemnaUserID == nil {
		return errors.New("user has no Remnawave identity")
	}
	if err := w.remnawave.ApplyEntitlement(ctx, *user.RemnaUserID, current.TrafficLimitBytes, current.ResetStrategy, current.SquadUUIDs, current.ValidUntil); err != nil {
		return fmt.Errorf("prepare Remnawave entitlement continuity: %w", err)
	}
	return nil
}
