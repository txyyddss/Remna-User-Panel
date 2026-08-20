package notifications

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	jobpayload "github.com/txyyddss/Remna-User-Panel/internal/outbox"
)

// Sender delivers one MarkdownV2 message through the queued Telegram adapter.
type Sender interface {
	SendMarkdownV2Message(context.Context, int64, int64, string) error
}

// Worker owns durable private-chat delivery.
type Worker struct {
	sender   Sender
	logger   *slog.Logger
	location *time.Location
}

// NewWorker creates a user-notification outbox handler.
func NewWorker(sender Sender, logger *slog.Logger, location *time.Location) *Worker {
	return &Worker{sender: sender, logger: logger, location: location}
}

// HandleOutbox validates, formats, and sends one immutable event.
func (w *Worker) HandleOutbox(ctx context.Context, job model.OutboxJob) error {
	payload, err := jobpayload.DecodeUserNotification(job)
	if err != nil {
		return err
	}
	message, err := Format(payload, w.location)
	if err != nil {
		return fmt.Errorf("format user notification %s: %w", payload.EventKey, err)
	}
	if err := w.sender.SendMarkdownV2Message(ctx, payload.ChatID, 0, message); err != nil {
		if w.logger != nil {
			w.logger.Warn("user notification delivery will retry", "event_key", payload.EventKey, "kind", payload.Kind,
				"user_id", payload.UserID, "attempt", job.Attempts, "error", err)
		}
		return fmt.Errorf("send user notification %s: %w", payload.EventKey, err)
	}
	if w.logger != nil {
		w.logger.Info("user notification delivered", "event_key", payload.EventKey, "kind", payload.Kind, "user_id", payload.UserID)
	}
	return nil
}
