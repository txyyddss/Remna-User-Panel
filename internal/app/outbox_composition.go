package app

import (
	"context"
	"fmt"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/affiliates"
	"github.com/txyyddss/Remna-User-Panel/internal/billing"
	"github.com/txyyddss/Remna-User-Panel/internal/connections"
	"github.com/txyyddss/Remna-User-Panel/internal/emby"
	"github.com/txyyddss/Remna-User-Panel/internal/entitlements"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/notifications"
	jobpayload "github.com/txyyddss/Remna-User-Panel/internal/outbox"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/outbox"
	"github.com/txyyddss/Remna-User-Panel/internal/providerops"
	"github.com/txyyddss/Remna-User-Panel/internal/rollover"
)

func registerCoreOutboxHandlers(worker *outbox.Worker, store *database.Store, entitlementsWorker *entitlements.Worker,
	rolloverWorker *rollover.Service, embyService *emby.Service, paymentAnnouncements *billing.PaymentAnnouncementWorker,
	affiliateNotifications *affiliates.NotificationWorker, userNotifications *notifications.Worker,
	scanWorker *connections.Worker, expiryWorker *connections.ExpiryWorker, dispatcher *providerops.Dispatcher) error {
	if err := worker.Register(jobpayload.PaymentSuccessAnnouncementKind, paymentAnnouncements); err != nil {
		return fmt.Errorf("register payment announcement outbox handler: %w", err)
	}
	for _, kind := range []string{jobpayload.AffiliateSuccessKind, jobpayload.AffiliateTierUpgradeKind} {
		if err := worker.Register(kind, affiliateNotifications); err != nil {
			return fmt.Errorf("register %s outbox handler: %w", kind, err)
		}
	}
	if err := worker.Register(jobpayload.UserNotificationKind, userNotifications); err != nil {
		return fmt.Errorf("register user notification outbox handler: %w", err)
	}
	for _, kind := range []string{"remna_apply_entitlement", "remna_sync_user", jobpayload.ContinuityKind} {
		if err := worker.Register(kind, entitlementsWorker); err != nil {
			return fmt.Errorf("register %s outbox handler: %w", kind, err)
		}
	}
	if err := worker.Register("rollover_finalize", outbox.HandlerFunc(rolloverWorker.HandleOutbox)); err != nil {
		return err
	}
	if err := worker.Register(connections.ScanRequestOutboxKind, scanWorker); err != nil {
		return err
	}
	if err := worker.Register(connections.BlockExpiryOutboxKind, expiryWorker); err != nil {
		return err
	}
	if err := registerProviderDispatcher(worker, dispatcher); err != nil {
		return err
	}
	if err := worker.Register(emby.ProvisionOutboxKind, outbox.HandlerFunc(func(ctx context.Context, job model.OutboxJob) error {
		accountID, err := jobpayload.TargetID(job, "accountId")
		if err != nil {
			return err
		}
		return embyService.HandleProvisionJob(ctx, accountID)
	})); err != nil {
		return err
	}
	return worker.Register("questionnaire_settlement", outbox.HandlerFunc(func(ctx context.Context, job model.OutboxJob) error {
		importID, err := jobpayload.TargetID(job, "importId")
		if err != nil {
			return err
		}
		_, err = store.SettleQuestionnaireImport(ctx, importID, time.Now().UTC())
		return err
	}))
}
