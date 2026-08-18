package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
	"github.com/txyyddss/Remna-User-Panel/internal/providerops"
)

// EditAdminEntitlement atomically replaces mutable fields, audits, and queues sync.
func (s *Store) EditAdminEntitlement(ctx context.Context, input AdminEntitlementEditInput, now time.Time) (model.Purchase, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Purchase{}, err
	}
	defer func() { _ = tx.Rollback() }()
	before, err := scanPurchase(tx.QueryRowContext(ctx, purchaseSelect+` WHERE purchases.id=? AND purchases.user_id=?`, input.PurchaseID, input.UserID))
	if err != nil {
		return model.Purchase{}, err
	}
	before.SquadUUIDs, err = purchaseSquadsFrom(ctx, tx, before.ID)
	if err != nil {
		return model.Purchase{}, err
	}
	operation, replayed, err := createProviderOperationTx(ctx, tx, providerops.CreateInput{
		ActorUserID: input.ActorUserID, OwnerUserID: input.UserID, Kind: providerops.KindAdminEntitlementEdit,
		IdempotencyKey: input.IdempotencyKey, RequestFingerprint: input.RequestFingerprint,
		Items: []providerops.ItemInput{{Key: input.UserID, TargetType: "user", TargetID: input.UserID}},
	}, now)
	if err != nil {
		return model.Purchase{}, err
	}
	if replayed {
		if err := tx.Commit(); err != nil {
			return model.Purchase{}, err
		}
		return s.PurchaseByID(ctx, input.PurchaseID)
	}
	if !before.UpdatedAt.Equal(input.ExpectedUpdatedAt) {
		return model.Purchase{}, ErrConflict
	}
	var comboExists int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM combos WHERE id=?`, input.ComboID).Scan(&comboExists); err != nil {
		return model.Purchase{}, err
	}
	if comboExists != 1 {
		return model.Purchase{}, ErrNotFound
	}
	squads, err := json.Marshal(input.SquadUUIDs)
	if err != nil {
		return model.Purchase{}, err
	}
	result, err := tx.ExecContext(ctx, `UPDATE purchases SET combo_id=?,valid_from=?,valid_until=?,status=?,
		entitlement_traffic_limit_bytes=?,entitlement_reset_strategy=?,entitlement_squad_uuids=?,updated_at=?
		WHERE id=? AND user_id=? AND updated_at=?`, input.ComboID, stamp(input.ValidFrom), stamp(input.ValidUntil), input.Status,
		input.TrafficLimitBytes, input.ResetStrategy, string(squads), stamp(now), input.PurchaseID, input.UserID, stamp(before.UpdatedAt))
	if err != nil {
		return model.Purchase{}, fmt.Errorf("edit entitlement: %w", err)
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
		if rowsErr != nil {
			return model.Purchase{}, rowsErr
		}
		return model.Purchase{}, ErrConflict
	}
	if err := reconcilePurchaseContinuityTx(ctx, tx, input.PurchaseID, input.Status, input.ValidFrom, now); err != nil {
		return model.Purchase{}, err
	}
	if err := insertEntitlementAudit(ctx, tx, input.ActorUserID, "entitlement.edit", input.PurchaseID,
		input.Reason, operation.Receipt.ID, entitlementSnapshot(before), map[string]any{"comboId": input.ComboID,
			"validFrom": input.ValidFrom, "validUntil": input.ValidUntil, "status": input.Status,
			"trafficLimitBytes": input.TrafficLimitBytes, "resetStrategy": input.ResetStrategy, "squadUuids": input.SquadUUIDs}, now); err != nil {
		return model.Purchase{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.Purchase{}, err
	}
	return s.PurchaseByID(ctx, input.PurchaseID)
}

func entitlementSnapshot(purchase model.Purchase) map[string]any {
	return map[string]any{"comboId": purchase.ComboID, "validFrom": purchase.ValidFrom, "validUntil": purchase.ValidUntil,
		"status": purchase.Status, "trafficLimitBytes": purchase.TrafficLimitBytes,
		"resetStrategy": purchase.ResetStrategy, "squadUuids": purchase.SquadUUIDs}
}

func insertEntitlementAudit(ctx context.Context, tx *sql.Tx, actorID, action, purchaseID, reason, operationID string, before, after any, now time.Time) error {
	auditID, err := ids.New()
	if err != nil {
		return err
	}
	detail, err := json.Marshal(map[string]any{"reason": reason, "operationId": operationID, "before": before, "after": after})
	if err != nil {
		return fmt.Errorf("encode entitlement audit: %w", err)
	}
	return insertAuditTx(ctx, tx, auditID, &actorID, action, "purchase", purchaseID, string(detail), now)
}
