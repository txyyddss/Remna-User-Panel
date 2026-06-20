package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	messageLogRetentionHours     = 72
	paymentOrderRetentionHours   = 72
	webhookEventRetentionHours   = 72
	telegramOutboxRetentionHours = 72
	closedTicketRetentionHours   = 24
	pendingOrderExpireHours      = 1
)

// RunDataCleanup removes expired data from the database to save space.
// - message_logs older than 72h are deleted
// - payment_orders in terminal state (failed/expired/cancelled) older than 72h are deleted
// - payment_orders stuck in pending for over 1h are auto-cancelled
// - webhook_events that are processed older than 72h are deleted
// - closed support tickets older than 24h are removed from app_settings
func RunDataCleanup(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return nil
	}

	now := time.Now().UTC()

	// 0. Auto-cancel pending orders older than 1h
	pendingCutoff := now.Add(-time.Duration(pendingOrderExpireHours) * time.Hour)
	result, err := pool.Exec(ctx, `UPDATE payment_orders SET status='expired', updated_at=NOW()
WHERE status='pending' AND created_at < $1`, pendingCutoff)
	if err != nil {
		slog.Warn("data cleanup: failed to expire pending orders", "error", err)
	} else {
		if result.RowsAffected() > 0 {
			slog.Info("data cleanup: auto-expired pending payment orders", "count", result.RowsAffected())
		}
	}

	// 1. Delete old message logs
	logCutoff := now.Add(-time.Duration(messageLogRetentionHours) * time.Hour)
	result, err = pool.Exec(ctx, "DELETE FROM message_logs WHERE timestamp < $1", logCutoff)
	if err != nil {
		slog.Warn("data cleanup: failed to delete old message logs", "error", err)
	} else {
		slog.Debug("data cleanup: deleted old message logs", "count", result.RowsAffected())
	}

	// 2. Delete terminal payment orders older than 72h
	payCutoff := now.Add(-time.Duration(paymentOrderRetentionHours) * time.Hour)
	result, err = pool.Exec(ctx, `DELETE FROM payment_orders WHERE status IN ('failed','expired','cancelled') AND updated_at < $1`, payCutoff)
	if err != nil {
		slog.Warn("data cleanup: failed to delete old payment orders", "error", err)
	} else {
		slog.Debug("data cleanup: deleted old terminal payment orders", "count", result.RowsAffected())
	}

	// 3. Delete processed webhook events older than 72h
	webhookCutoff := now.Add(-time.Duration(webhookEventRetentionHours) * time.Hour)
	result, err = pool.Exec(ctx, "DELETE FROM webhook_events WHERE status = 'processed' AND processed_at < $1", webhookCutoff)
	if err != nil {
		slog.Warn("data cleanup: failed to delete old webhook events", "error", err)
	} else {
		slog.Debug("data cleanup: deleted old processed webhook events", "count", result.RowsAffected())
	}

	// 4. Delete old completed Telegram outbox messages.
	outboxCutoff := now.Add(-time.Duration(telegramOutboxRetentionHours) * time.Hour)
	result, err = pool.Exec(ctx, "DELETE FROM telegram_outbox WHERE status IN ('sent','failed') AND COALESCE(sent_at,created_at) < $1", outboxCutoff)
	if err != nil {
		slog.Warn("data cleanup: failed to delete old telegram outbox messages", "error", err)
	} else {
		slog.Debug("data cleanup: deleted old telegram outbox messages", "count", result.RowsAffected())
	}

	// 5. Remove closed support tickets older than 24h from app_settings.
	ticketCutoff := now.Add(-time.Duration(closedTicketRetentionHours) * time.Hour)
	if err := cleanupClosedTickets(ctx, pool, ticketCutoff); err != nil {
		slog.Warn("data cleanup: failed to cleanup closed support tickets", "error", err)
	}

	return nil
}

// cleanupClosedTickets removes closed support tickets older than the cutoff time
// from the SUPPORT_TICKETS JSON array in app_settings.
func cleanupClosedTickets(ctx context.Context, pool *pgxpool.Pool, cutoff time.Time) error {
	var raw json.RawMessage
	err := pool.QueryRow(ctx, "SELECT value FROM app_settings WHERE key='SUPPORT_TICKETS'").Scan(&raw)
	if err != nil {
		return nil //nolint:nilerr // Key doesn't exist — nothing to clean.
	}

	var tickets []map[string]any
	if err := json.Unmarshal(raw, &tickets); err != nil {
		return fmt.Errorf("unmarshal support tickets: %w", err)
	}

	originalCount := len(tickets)
	filtered := make([]map[string]any, 0, len(tickets))
	removedCount := 0

	for _, ticket := range tickets {
		status := stringValue(ticket, "status")
		if status == "closed" {
			updatedAt := stringValue(ticket, "updated_at")
			if updatedAt != "" {
				t, err := time.Parse(time.RFC3339, updatedAt)
				if err == nil && t.Before(cutoff) {
					removedCount++
					continue
				}
			}
			// If we can't parse the date but status is closed and the ticket
			// has no updated_at, use created_at as fallback
			createdAt := stringValue(ticket, "created_at")
			if updatedAt == "" && createdAt != "" {
				t, err := time.Parse(time.RFC3339, createdAt)
				if err == nil && t.Before(cutoff) {
					removedCount++
					continue
				}
			}
		}
		filtered = append(filtered, ticket)
	}

	if removedCount == 0 {
		return nil
	}

	// Write back the filtered list
	updated, err := json.Marshal(filtered)
	if err != nil {
		return fmt.Errorf("marshal filtered tickets: %w", err)
	}

	_, err = pool.Exec(ctx, "UPDATE app_settings SET value=$2, updated_at=NOW() WHERE key='SUPPORT_TICKETS'", updated)
	if err != nil {
		return fmt.Errorf("update support tickets: %w", err)
	}

	slog.Debug("data cleanup: removed closed support tickets", "removed", removedCount, "remaining", originalCount-removedCount)
	return nil
}
