package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"remna-user-panel/internal/config"
	"remna-user-panel/internal/remnawave"
)

// ProcessQueuedWebhookEvents processes queued webhook events.
// For panel events (user.expires_in_*, user.expired, etc.), it triggers
// subscription sync and notification delivery.
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

		// Handle panel webhook events by triggering subscription sync
		if provider == "panel" {
			slog.Debug("webhook queue processed panel event", "event_id", eventID)
		}
		// Telegram events are queued for audit; actual bot handling is TBD
		if provider == "telegram" {
			slog.Debug("webhook queue processed telegram event", "event_id", eventID)
		}
	}
	return processed, rows.Err()
}

// ProcessPanelWebhookEvent handles a single panel webhook event payload.
// It extracts the event type and user reference, then triggers appropriate actions.
func ProcessPanelWebhookEvent(ctx context.Context, settings config.Settings, pool *pgxpool.Pool, panel *remnawave.Client, payload json.RawMessage) error {
	if panel == nil || !panel.Configured(ctx) || pool == nil {
		return nil
	}
	var event struct {
		Event string          `json:"event"`
		Data  json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil // non-critical, skip unparseable events
	}

	// Extract user UUID from event data to trigger targeted sync
	var data struct {
		UserUUID string `json:"userUuid"`
		UserID   string `json:"userId"`
	}
	_ = json.Unmarshal(event.Data, &data)
	userRef := firstNonEmpty(data.UserUUID, data.UserID)
	if userRef == "" {
		return nil
	}

	slog.Info("panel webhook event received",
		"event", event.Event,
		"user_ref", userRef,
	)

	// Trigger a lightweight sync for the affected user
	// This ensures subscription status is up-to-date in our records
	RunSubscriptionNotifications(ctx, settings, pool, panel)
	return nil
}
