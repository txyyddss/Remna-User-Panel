package app

import (
	"context"
	"errors"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

func (a *Application) runScheduler(ctx context.Context) {
	outboxTicker := time.NewTicker(5 * time.Second)
	transitionTicker := time.NewTicker(30 * time.Second)
	maintenanceTicker := time.NewTicker(10 * time.Minute)
	starsTicker := time.NewTicker(5 * time.Minute)
	backupTimer := time.NewTimer(time.Until(a.nextBackup(time.Now())))
	defer outboxTicker.Stop()
	defer transitionTicker.Stop()
	defer maintenanceTicker.Stop()
	defer starsTicker.Stop()
	defer backupTimer.Stop()
	a.configureTelegram(ctx)
	startupNow := time.Now().UTC()
	if err := a.store.RecoverOutbox(ctx, startupNow, startupNow); err != nil {
		a.logger.Error("startup outbox recovery failed", "error", err)
	}
	if err := a.store.EnqueueDueEntitlementTransitions(ctx, startupNow); err != nil {
		a.logger.Error("startup entitlement transition scan failed", "error", err)
	}
	if err := a.outbox.Drain(ctx, 50); err != nil && !errors.Is(err, context.Canceled) {
		a.logger.Error("startup outbox drain failed", "error", err)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-outboxTicker.C:
			if err := a.outbox.Drain(ctx, 20); err != nil && !errors.Is(err, context.Canceled) {
				a.logger.Error("outbox drain failed", "error", err)
			}
		case now := <-transitionTicker.C:
			if err := a.store.RecoverOutbox(ctx, now.UTC().Add(-2*time.Minute), now.UTC()); err != nil {
				a.logger.Error("outbox lease recovery failed", "error", err)
			}
			if err := a.store.EnqueueDueEntitlementTransitions(ctx, now.UTC()); err != nil {
				a.logger.Error("entitlement transition scan failed", "error", err)
			}
			if err := a.store.ExpireStalePaymentOrders(ctx, now.UTC()); err != nil {
				a.logger.Error("payment expiry scan failed", "error", err)
			}
		case <-maintenanceTicker.C:
			a.configureTelegram(ctx)
			maintenanceNow := time.Now().UTC()
			if err := a.store.DeleteExpiredSessions(ctx, maintenanceNow); err != nil {
				a.logger.Error("session cleanup failed", "error", err)
			}
			if err := a.store.PruneTransientRecords(ctx, maintenanceNow); err != nil {
				a.logger.Error("transient database retention failed", "error", err)
			}
			if err := a.backups.RemoveExpired(); err != nil {
				a.logger.Error("backup retention failed", "error", err)
			}
			if err := database.PruneExpansionBackups(a.config.DatabasePath, maintenanceNow); err != nil {
				a.logger.Error("migration snapshot retention failed", "error", err)
			}
		case <-starsTicker.C:
			a.reconcileStars(ctx)
		case <-backupTimer.C:
			backupCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
			if _, err := a.backups.Run(backupCtx); err != nil && !errors.Is(err, context.Canceled) {
				a.logger.Error("scheduled backup failed", "error", err)
			}
			cancel()
			backupTimer.Reset(time.Until(a.nextBackup(time.Now())))
		}
	}
}

func (a *Application) nextBackup(now time.Time) time.Time {
	local := now.In(a.config.Timezone)
	next := time.Date(local.Year(), local.Month(), local.Day(), a.config.BackupHour, 0, 0, 0, a.config.Timezone)
	if !next.After(local) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}
