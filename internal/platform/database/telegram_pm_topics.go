package database

import (
	"context"
	"time"
)

// ClaimPMTopic serializes creation independently of how many outbox drains run.
func (s *Store) ClaimPMTopic(ctx context.Context, id, operationID string, now time.Time) (bool, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	result, err := s.db.ExecContext(ctx, `UPDATE telegram_pm_conversations SET topic_state='creating',topic_operation_id=?,updated_at=?
		WHERE id=? AND topic_state='new' AND topic_id IS NULL`, operationID, stamp(now), id)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return count == 1, err
}

func (s *Store) SavePMTopic(ctx context.Context, id, operationID string, topicID int64, now time.Time) error {
	if topicID <= 1 {
		return ErrConflict
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	result, err := s.db.ExecContext(ctx, `UPDATE telegram_pm_conversations SET topic_id=?,topic_state='ready',updated_at=?
		WHERE id=? AND topic_state='creating' AND topic_operation_id=?`, topicID, stamp(now), id, operationID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err == nil && count != 1 {
		return ErrConflict
	}
	return err
}

func (s *Store) ReleasePMTopic(ctx context.Context, id, operationID string, uncertain bool, now time.Time) error {
	state := "new"
	if uncertain {
		state = "pending_review"
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_, err := s.db.ExecContext(ctx, `UPDATE telegram_pm_conversations SET topic_state=?,updated_at=? WHERE id=? AND topic_state='creating' AND topic_operation_id=?`, state, stamp(now), id, operationID)
	return err
}

func (s *Store) SavePMProfile(ctx context.Context, id string, topicID, messageID int64, now time.Time) error {
	if messageID <= 0 {
		return ErrConflict
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	result, err := s.db.ExecContext(ctx, `UPDATE telegram_pm_conversations SET profile_message_id=?,profile_state='ready',updated_at=? WHERE id=? AND topic_id=? AND topic_state='ready'`, messageID, stamp(now), id, topicID)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err == nil && count != 1 {
		return ErrConflict
	}
	return err
}

func (s *Store) ClearPMProfile(ctx context.Context, id string, messageID int64) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_, err := s.db.ExecContext(ctx, `UPDATE telegram_pm_conversations SET profile_message_id=NULL,profile_state='missing' WHERE id=? AND profile_message_id=?`, id, messageID)
	return err
}

func (s *Store) PMProfileUncertain(ctx context.Context, id, operationID string) (bool, error) {
	var uncertain bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM provider_operation_items WHERE
		((target_type='pm_conversation' AND target_id=? AND item_key='profile') OR (target_type='pm_topic_recovery' AND target_id LIKE ?))
		AND operation_id<>? AND (status='processing' OR (status='pending_review' AND error_code='PM_PROFILE_UNCERTAIN')))`, id, id+":%", operationID).Scan(&uncertain)
	return uncertain, err
}

func (s *Store) MarkPMProfileUncertain(ctx context.Context, id string) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_, err := s.db.ExecContext(ctx, `UPDATE telegram_pm_conversations SET profile_state='pending_review' WHERE id=? AND profile_message_id IS NULL`, id)
	return err
}
