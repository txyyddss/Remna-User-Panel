package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/affiliates"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
)

// AcceptAffiliateReferral freezes a valid inviter on a private /start stub.
func (s *Store) AcceptAffiliateReferral(ctx context.Context, inviteeTelegramID, inviterTelegramID int64, now time.Time) (string, bool, error) {
	if inviteeTelegramID <= 0 || inviterTelegramID <= 0 || inviteeTelegramID == inviterTelegramID {
		return "", false, affiliates.ErrInvalidInput
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", false, err
	}
	defer func() { _ = tx.Rollback() }()
	var display string
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(NULLIF(username,''),NULLIF(telegram_username,''),NULLIF(telegram_first_name,''),'')
		FROM users WHERE telegram_id=? AND new_user=0`, inviterTelegramID).Scan(&display)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	stubID, err := ids.New()
	if err != nil {
		return "", false, err
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO users(id,telegram_id,new_user,inviter_id,notification_locale,created_at,updated_at)
		VALUES(?,?,1,?,'en',?,?) ON CONFLICT(telegram_id) DO UPDATE SET inviter_id=excluded.inviter_id,updated_at=excluded.updated_at
		WHERE users.new_user=1 AND users.inviter_id IS NULL`, stubID, inviteeTelegramID, inviterTelegramID, stamp(now), stamp(now))
	if err != nil {
		return "", false, fmt.Errorf("freeze affiliate inviter: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return "", false, err
	}
	if err := tx.Commit(); err != nil {
		return "", false, err
	}
	return display, affected == 1, nil
}
