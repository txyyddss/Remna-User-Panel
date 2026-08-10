package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
)

// DeleteActivityGame removes feature rows and their ledger evidence without
// reversing any already-settled balance effects.
func (s *Store) DeleteActivityGame(ctx context.Context, actorID, gameID string, now time.Time) error {
	return s.deleteFeature(ctx, actorID, "activity_game.delete", "activity_game", gameID, func(tx *sql.Tx) error {
		if err := requireRowTx(ctx, tx, `SELECT 1 FROM activity_games WHERE id=?`, gameID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM ledger_entries WHERE kind IN ('activity_bet_stake','activity_bet_payout') AND reference_id IN (SELECT id FROM activity_bets WHERE game_id=?)`, gameID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM activity_bets WHERE game_id=?`, gameID); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `DELETE FROM activity_games WHERE id=?`, gameID)
		return err
	})
}

// DeleteLuckyDraw preserves coupon grants and extension credits but removes the
// draw configuration, outcomes, prizes, snapshots, and feature ledger rows.
func (s *Store) DeleteLuckyDraw(ctx context.Context, actorID, drawID string, now time.Time) error {
	return s.deleteFeature(ctx, actorID, "lucky_draw.delete", "lucky_draw", drawID, func(tx *sql.Tx) error {
		if err := requireRowTx(ctx, tx, `SELECT 1 FROM activity_lucky_draws WHERE id=?`, drawID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM ledger_entries WHERE kind IN ('activity_draw_fee','activity_draw_reward') AND reference_id IN (SELECT id FROM activity_draw_results WHERE draw_id=?)`, drawID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM activity_draw_results WHERE draw_id=?`, drawID); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `DELETE FROM activity_lucky_draws WHERE id=?`, drawID)
		return err
	})
}

// DeleteQuestionnaire rejects processing imports/jobs, removes non-processing
// settlement jobs, imports and participants, and leaves awarded balances intact.
func (s *Store) DeleteQuestionnaire(ctx context.Context, actorID, questionnaireID string, now time.Time) error {
	return s.deleteFeature(ctx, actorID, "questionnaire.delete", "questionnaire", questionnaireID, func(tx *sql.Tx) error {
		if err := requireRowTx(ctx, tx, `SELECT 1 FROM questionnaires WHERE id=?`, questionnaireID); err != nil {
			return err
		}
		var processing int
		if err := tx.QueryRowContext(ctx, `SELECT
			(SELECT COUNT(*) FROM questionnaire_imports WHERE questionnaire_id=? AND status='processing') +
			(SELECT COUNT(*) FROM outbox_jobs WHERE kind='questionnaire_settlement' AND status='processing' AND json_extract(payload,'$.importId') IN (SELECT id FROM questionnaire_imports WHERE questionnaire_id=?))`, questionnaireID, questionnaireID).Scan(&processing); err != nil {
			return err
		}
		if processing > 0 {
			return ErrConflict
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM outbox_jobs WHERE kind='questionnaire_settlement' AND status<>'processing' AND json_extract(payload,'$.importId') IN (SELECT id FROM questionnaire_imports WHERE questionnaire_id=?)`, questionnaireID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM ledger_entries WHERE kind='questionnaire_reward' AND reference_id IN (SELECT id FROM questionnaire_participants WHERE questionnaire_id=?)`, questionnaireID); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `DELETE FROM questionnaires WHERE id=?`, questionnaireID)
		return err
	})
}

func (s *Store) deleteFeature(ctx context.Context, actorID, action, targetType, targetID string, mutate func(*sql.Tx) error) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := mutate(tx); err != nil {
		return err
	}
	auditID, err := ids.New()
	if err != nil {
		return err
	}
	if err := insertAuditTx(ctx, tx, auditID, &actorID, action, targetType, targetID, `{}`, now.UTC()); err != nil {
		return fmt.Errorf("append deletion audit: %w", err)
	}
	return tx.Commit()
}

func requireRowTx(ctx context.Context, tx *sql.Tx, query string, args ...any) error {
	var exists int
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
