package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

// CommitAutoRenewal atomically settles one calculated rollover and successor.
func (s *Store) CommitAutoRenewal(ctx context.Context, purchaseID string, now time.Time) (model.Purchase, error) {
	return s.commitAutoRenewal(ctx, purchaseID, nil, now)
}

// CommitAutoRenewalExcludingAddons excludes upstream-unavailable paid squads.
func (s *Store) CommitAutoRenewalExcludingAddons(ctx context.Context, purchaseID string, excludedAddonIDs []string, now time.Time) (model.Purchase, error) {
	return s.commitAutoRenewal(ctx, purchaseID, excludedAddonIDs, now)
}

func automaticRenewalSuccessorIDTx(ctx context.Context, tx *sql.Tx, sourceID string) (string, bool, error) {
	var successorID string
	err := tx.QueryRowContext(ctx, `SELECT id FROM purchases WHERE auto_renew_source_purchase_id=?`, sourceID).Scan(&successorID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("load automatic renewal successor: %w", err)
	}
	return successorID, true, nil
}

func rolloverStatusForCredit(credit int64) string {
	if credit > 0 {
		return "credited"
	}
	return "zero"
}
