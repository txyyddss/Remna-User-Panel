package database

import (
	"context"
	"strings"
	"time"
)

// RecordGroupMessageFact deduplicates every non-command message observed in
// the configured Telegram group without requiring a local user identity.
func (s *Store) RecordGroupMessageFact(ctx context.Context, chatID, messageID int64, localDate string, now time.Time) error {
	localDate = strings.TrimSpace(localDate)
	if chatID == 0 || messageID <= 0 || localDate == "" {
		return ErrConflict
	}
	if _, err := time.Parse(time.DateOnly, localDate); err != nil {
		return ErrConflict
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	_, err := s.db.ExecContext(ctx, `INSERT INTO activity_group_message_raw_events(chat_id,message_id,local_date,created_at)
		VALUES(?,?,?,?) ON CONFLICT(chat_id,message_id) DO NOTHING`, chatID, messageID, localDate, stamp(now.UTC()))
	return err
}
