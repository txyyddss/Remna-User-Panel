package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/integrations/remnawave"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	jobpayload "github.com/txyyddss/Remna-User-Panel/internal/outbox"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/outbox"
)

func registerAbuseOutboxHandlers(worker *outbox.Worker, store *database.Store, remna remnaAdapter, telegram *queuedTelegram) error {
	if err := worker.Register("abuse_punishment", outbox.HandlerFunc(func(ctx context.Context, job model.OutboxJob) error {
		return handleAbusePunishment(ctx, job, store, remna)
	})); err != nil {
		return err
	}
	if err := worker.Register("abuse_restore", outbox.HandlerFunc(func(ctx context.Context, job model.OutboxJob) error {
		return handleAbuseRestore(ctx, job, store, remna)
	})); err != nil {
		return err
	}
	return worker.Register("abuse_notification", outbox.HandlerFunc(func(ctx context.Context, job model.OutboxJob) error {
		return handleAbuseNotification(ctx, job, store, telegram)
	}))
}
func handleAbusePunishment(ctx context.Context, job model.OutboxJob, store *database.Store, remna remnaAdapter) error {
	id, err := jobpayload.TargetID(job, "recordId")
	if err != nil {
		return err
	}
	item, err := store.AbuseJob(ctx, id)
	if err != nil {
		return err
	}
	switch string(item.Action) {
	case "warning", "none":
		return nil
	case "subscription_revoke":
		return remna.AbuseRevoke(ctx, item.RemoteUserID)
	case "temporary_ban":
		return remna.AbuseSetStatus(ctx, item.RemoteUserID, remnawave.UserStatusDisabled)
	case "ip_ban":
		return remna.AbuseIPBan(ctx, item.RemoteUserID, item.Nodes, item.AllNodes, item.DurationMinutes*60)
	default:
		return fmt.Errorf("unsupported abuse action %q", item.Action)
	}
}
func handleAbuseRestore(ctx context.Context, job model.OutboxJob, store *database.Store, remna remnaAdapter) error {
	userID, err := jobpayload.TargetID(job, "userId")
	if err != nil {
		return err
	}
	remoteID, err := store.AbuseRestoreRemoteID(ctx, userID)
	if err == database.ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	if err = remna.AbuseSetStatus(ctx, remoteID, remnawave.UserStatusActive); err != nil {
		return err
	}
	return store.CompleteAbuseRestore(ctx, userID, time.Now().UTC())
}
func handleAbuseNotification(ctx context.Context, job model.OutboxJob, store *database.Store, telegram *queuedTelegram) error {
	recordID, err := jobpayload.TargetID(job, "recordId")
	if err != nil {
		return err
	}
	rawID, err := jobpayload.TargetID(job, "telegramId")
	if err != nil {
		return err
	}
	var telegramID int64
	if _, err = fmt.Sscan(rawID, &telegramID); err != nil {
		return err
	}
	delivery, err := store.AbuseDelivery(ctx, recordID, telegramID)
	if err != nil {
		return err
	}
	if delivery.Delivered {
		return nil
	}
	message := abuseMessage(delivery.Reason, delivery.QPS, delivery.Limit, string(delivery.Action), delivery.ExpiresAt)
	if err = telegram.SendMarkdownV2Message(ctx, telegramID, 0, message); err != nil {
		return err
	}
	return store.MarkAbuseDelivery(ctx, recordID, telegramID, time.Now().UTC())
}
func abuseMessage(reason string, qps, limit int, action string, expires *time.Time) string {
	parts := []string{"⚠️ *Abuse detector*", "Reason: " + escapeMarkdown(reason), fmt.Sprintf("QPS: %d / %d", qps, limit), "Action: " + escapeMarkdown(action)}
	if expires != nil {
		parts = append(parts, "Expiry: "+escapeMarkdown(expires.UTC().Format(time.RFC3339)))
	}
	return strings.Join(parts, "\n")
}
func escapeMarkdown(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(", "\\(", ")", "\\)", "~", "\\~", "`", "\\`", ">", "\\>", "#", "\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|", "\\|", "{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!")
	return replacer.Replace(value)
}
