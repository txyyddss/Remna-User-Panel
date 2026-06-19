package httpapi

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ProcessQueuedWebhookEvents closes the loop for webhook_events persisted by
// webhook handlers. The current Go port records and acknowledges events; domain
// handlers can be added here without changing the ingestion endpoints.
func ProcessQueuedWebhookEvents(ctx context.Context, pool *pgxpool.Pool, limit int) (int, error) {
	if pool == nil {
		return 0, nil
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := pool.Query(ctx, `
WITH picked AS (
	SELECT event_id
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
RETURNING w.event_id`, limit)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	processed := 0
	for rows.Next() {
		var eventID int64
		if err := rows.Scan(&eventID); err != nil {
			return processed, err
		}
		processed++
	}
	return processed, rows.Err()
}
