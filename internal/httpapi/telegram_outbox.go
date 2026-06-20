package httpapi

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"remna-user-panel/internal/config"
)

// ProcessTelegramOutbox claims and sends a bounded batch of queued messages.
func ProcessTelegramOutbox(ctx context.Context, settings config.Settings, pool *pgxpool.Pool, limit int) (int, error) {
	if pool == nil {
		return 0, fmt.Errorf("database_not_configured")
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := pool.Query(ctx, `SELECT message_id FROM telegram_outbox
WHERE status='queued' AND next_attempt_at<=NOW() ORDER BY message_id LIMIT $1`, limit)
	if err != nil {
		return 0, err
	}
	ids := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	processed := 0
	for _, id := range ids {
		var chatID, userID int64
		var text, parseMode, batchID string
		var attempts int
		err := pool.QueryRow(ctx, `UPDATE telegram_outbox SET status='sending',attempts=attempts+1
WHERE message_id=$1 AND status='queued' RETURNING chat_id,COALESCE(user_id,0),text,parse_mode,batch_id,attempts`, id).Scan(&chatID, &userID, &text, &parseMode, &batchID, &attempts)
		if err != nil {
			continue
		}
		sendErr := sendTelegramMessage(ctx, settings, chatID, text, parseMode)
		if sendErr == nil {
			_, _ = pool.Exec(ctx, "UPDATE telegram_outbox SET status='sent',sent_at=NOW(),last_error=NULL WHERE message_id=$1", id)
			processed++
			continue
		}
		if attempts >= 5 {
			_, _ = pool.Exec(ctx, "UPDATE telegram_outbox SET status='failed',last_error=$2 WHERE message_id=$1", id, sendErr.Error())
			recordMessageLog(ctx, pool, messageLogEntry{TargetUserID: userID, EventType: "telegram_outbox_failed", Content: text, IsAdminEvent: true, Payload: map[string]any{"batch_id": batchID, "error": sendErr.Error()}})
		} else {
			delay := time.Duration(attempts*attempts) * time.Minute
			_, _ = pool.Exec(ctx, "UPDATE telegram_outbox SET status='queued',last_error=$2,next_attempt_at=$3 WHERE message_id=$1", id, sendErr.Error(), time.Now().Add(delay))
		}
	}
	return processed, nil
}
