package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ProcessQueuedWebhookEvents processes queued webhook events.
// Telegram and payment callbacks are retained for asynchronous audit work.
func ProcessQueuedWebhookEvents(ctx context.Context, pool *pgxpool.Pool, limit int) (int, error) {
	if pool == nil {
		return 0, nil
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := pool.Query(ctx, `
WITH picked AS (
	SELECT event_id, provider, payload
	FROM webhook_events
	WHERE status='queued'
	ORDER BY created_at ASC
	LIMIT $1
	FOR UPDATE SKIP LOCKED
)
UPDATE webhook_events w
SET status='processed', processed_at=NOW()
FROM picked
WHERE w.event_id=picked.event_id
RETURNING w.event_id, picked.provider, picked.payload`, limit)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	processed := 0
	for rows.Next() {
		var eventID int64
		var provider string
		var payload json.RawMessage
		if err := rows.Scan(&eventID, &provider, &payload); err != nil {
			return processed, err
		}
		processed++

		// Telegram delivery is handled synchronously; the queue is its audit record.
		if provider == "telegram" {
			slog.Debug("webhook queue processed telegram event", "event_id", eventID)
		}
	}
	return processed, rows.Err()
}
