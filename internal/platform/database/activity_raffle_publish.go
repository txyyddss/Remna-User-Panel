package database

import (
	"context"
	"encoding/json"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/activity"
)

// PublishRaffle queues the group announcement; entries remain closed until its ID is saved.
func (s *Store) PublishRaffle(ctx context.Context, id string, groupID int64, now time.Time) error {
	if groupID == 0 {
		return activity.ErrInvalidInput
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	draw, err := luckyDrawByID(ctx, tx, id, false)
	if err != nil {
		return err
	}
	draw.Prizes, err = luckyPrizes(ctx, tx, id, false)
	if err != nil {
		return err
	}
	if draw.Kind != "raffle" || draw.Status != "draft" {
		return ErrConflict
	}
	if err = draw.LuckyDrawInput.Validate(); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE activity_lucky_draws SET status='publishing',group_chat_id=?,updated_at=? WHERE id=? AND status='draft'`, groupID, stamp(now), id); err != nil {
		if isUniqueConstraint(err) {
			return ErrConflict
		}
		return err
	}
	payload, _ := json.Marshal(map[string]any{"drawId": id})
	if err = insertOutboxTx(ctx, tx, "draw_telegram_publish", string(payload), now, now); err != nil {
		return err
	}
	return tx.Commit()
}

// ConfirmRafflePublished opens only the exact publishing draft.
func (s *Store) ConfirmRafflePublished(ctx context.Context, id string, messageID int64, now time.Time) error {
	if messageID <= 0 {
		return activity.ErrInvalidInput
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	result, err := s.db.ExecContext(ctx, `UPDATE activity_lucky_draws SET status='open',announcement_message_id=?,updated_at=?
  WHERE id=? AND status='publishing' AND announcement_message_id IS NULL`, messageID, stamp(now), id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return ErrConflict
	}
	return nil
}
