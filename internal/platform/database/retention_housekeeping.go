package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func pruneHousekeepingTx(ctx context.Context, tx *sql.Tx, cutoff7Days, cutoff24Hours, now time.Time,
	counts map[string]int64) error {
	var err error
	if counts["user_notification_events"], err = deleteCount(ctx, tx, `DELETE FROM user_notification_events
		WHERE queued_at IS NOT NULL AND queued_at<?`, stamp(now.Add(-3*24*time.Hour))); err != nil {
		return fmt.Errorf("prune completed user notifications: %w", err)
	}
	if counts["sessions"], err = deleteCount(ctx, tx, `DELETE FROM sessions WHERE expires_at<=?
		OR user_id IN (SELECT id FROM users WHERE role='user' AND onboarding_state IN ('intro','membership') AND created_at<?)
		OR EXISTS (SELECT 1 FROM sessions newer WHERE newer.user_id=sessions.user_id
			AND (newer.created_at>sessions.created_at OR (newer.created_at=sessions.created_at AND newer.token_hash>sessions.token_hash)))`,
		stamp(now), stamp(cutoff24Hours)); err != nil {
		return fmt.Errorf("prune stale sessions: %w", err)
	}
	if err := pruneAbandonedUsersTx(ctx, tx, cutoff24Hours, counts); err != nil {
		return err
	}
	if counts["maintenance_runs"], err = deleteCount(ctx, tx, `DELETE FROM maintenance_runs WHERE started_at<?`,
		stamp(cutoff7Days)); err != nil {
		return fmt.Errorf("prune maintenance runs: %w", err)
	}
	return nil
}

func pruneAbandonedUsersTx(ctx context.Context, tx *sql.Tx, cutoff time.Time, counts map[string]int64) error {
	if _, err := tx.ExecContext(ctx, `CREATE TEMP TABLE maintenance_user_candidates(id TEXT PRIMARY KEY,telegram_id INTEGER NOT NULL)`); err != nil {
		return fmt.Errorf("create abandoned user set: %w", err)
	}
	_, err := tx.ExecContext(ctx, `INSERT INTO maintenance_user_candidates(id,telegram_id)
		SELECT user.id,user.telegram_id FROM users user
		WHERE user.role='user' AND user.onboarding_state IN ('intro','membership') AND user.created_at<?
		AND NOT EXISTS (SELECT 1 FROM ledger_entries WHERE user_id=user.id)
		AND NOT EXISTS (SELECT 1 FROM purchases WHERE user_id=user.id)
		AND NOT EXISTS (SELECT 1 FROM payment_orders WHERE user_id=user.id)
		AND NOT EXISTS (SELECT 1 FROM affiliate_settlements WHERE invited_user_id=user.id OR inviter_user_id=user.id)
		AND NOT EXISTS (SELECT 1 FROM affiliate_tier_awards WHERE inviter_user_id=user.id)`, stamp(cutoff))
	if err != nil {
		return fmt.Errorf("select abandoned users: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE users SET inviter_id=NULL
		WHERE inviter_id IN (SELECT telegram_id FROM maintenance_user_candidates)`); err != nil {
		return fmt.Errorf("detach abandoned inviters: %w", err)
	}
	if counts["abandoned_users"], err = deleteCount(ctx, tx, `DELETE FROM users
		WHERE id IN (SELECT id FROM maintenance_user_candidates)`); err != nil {
		return fmt.Errorf("prune abandoned users: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DROP TABLE maintenance_user_candidates`); err != nil {
		return fmt.Errorf("drop abandoned user set: %w", err)
	}
	return nil
}
