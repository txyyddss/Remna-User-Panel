package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
)

// ClaimMaintenanceRun acquires one durable lease for a configured local date.
func (s *Store) ClaimMaintenanceRun(ctx context.Context, localDate, leaseOwner string, leaseUntil, now time.Time) (string, bool, error) {
	localDate, leaseOwner = strings.TrimSpace(localDate), strings.TrimSpace(leaseOwner)
	if _, err := time.Parse(time.DateOnly, localDate); err != nil || leaseOwner == "" || !leaseUntil.After(now) {
		return "", false, ErrConflict
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", false, fmt.Errorf("begin maintenance claim: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	runID, err := ids.New()
	if err != nil {
		return "", false, err
	}
	result, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO maintenance_runs(id,local_date,status,lease_owner,
		lease_expires_at,started_at,updated_at) VALUES(?,?,'running',?,?,?,?)`, runID, localDate,
		leaseOwner, stamp(leaseUntil), stamp(now), stamp(now))
	if err != nil {
		return "", false, fmt.Errorf("insert maintenance claim: %w", err)
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
		return "", false, rowsErr
	} else if affected == 0 {
		result, err = tx.ExecContext(ctx, `UPDATE maintenance_runs SET id=?,status='running',lease_owner=?,lease_expires_at=?,
			backup_run_id=NULL,cleanup_counts_json='{}',error_code='',started_at=?,updated_at=?,completed_at=NULL
			WHERE local_date=? AND status<>'succeeded' AND (status='failed' OR lease_expires_at<=?)`, runID,
			leaseOwner, stamp(leaseUntil), stamp(now), stamp(now), localDate, stamp(now))
		if err != nil {
			return "", false, fmt.Errorf("renew maintenance claim: %w", err)
		}
		if changed, changeErr := result.RowsAffected(); changeErr != nil {
			return "", false, changeErr
		} else if changed == 0 {
			if err := tx.QueryRowContext(ctx, `SELECT id FROM maintenance_runs WHERE local_date=?`, localDate).Scan(&runID); err != nil {
				return "", false, fmt.Errorf("load maintenance claim: %w", err)
			}
			return runID, false, nil
		}
	}
	if err := tx.Commit(); err != nil {
		return "", false, fmt.Errorf("commit maintenance claim: %w", err)
	}
	return runID, true, nil
}

// CompleteMaintenanceRun records sanitized outcome and aggregate deletion counts.
func (s *Store) CompleteMaintenanceRun(ctx context.Context, runID, backupRunID string, counts map[string]int64, runErr error, now time.Time) error {
	if counts == nil {
		counts = map[string]int64{}
	}
	for _, count := range counts {
		if count < 0 {
			return ErrConflict
		}
	}
	payload, err := json.Marshal(counts)
	if err != nil {
		return fmt.Errorf("encode maintenance counts: %w", err)
	}
	status, errorCode := "succeeded", ""
	if runErr != nil {
		status, errorCode = "failed", maintenanceErrorCode(runErr)
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	result, err := s.db.ExecContext(ctx, `UPDATE maintenance_runs SET status=?,backup_run_id=?,cleanup_counts_json=?,
		error_code=?,lease_expires_at=?,updated_at=?,completed_at=? WHERE id=? AND status='running'`, status,
		nullIfEmpty(backupRunID), string(payload), errorCode, stamp(now), stamp(now), stamp(now), runID)
	if err != nil {
		return fmt.Errorf("complete maintenance run: %w", err)
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
		return ErrConflict
	}
	return nil
}

func maintenanceErrorCode(err error) string {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return "MAINTENANCE_CANCELLED"
	}
	return "MAINTENANCE_FAILED"
}
