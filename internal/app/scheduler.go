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
	maintenanceTimer := time.NewTimer(time.Until(a.nextMaintenance(time.Now())))
	statisticsTimer := time.NewTimer(time.Until(nextStatisticsRefresh(time.Now())))
	defer outboxTicker.Stop()
	defer transitionTicker.Stop()
	defer maintenanceTicker.Stop()
	defer starsTicker.Stop()
	defer maintenanceTimer.Stop()
	defer statisticsTimer.Stop()
	a.configureTelegram(ctx)
	startupNow := time.Now().UTC()
	a.refreshProductStatistics(ctx, startupNow)
	if err := a.store.RecoverOutbox(ctx, startupNow, startupNow); err != nil {
		a.logger.Error("startup outbox recovery failed", "error", err)
	}
	if err := a.store.EnqueueContinuityBacklog(ctx, startupNow); err != nil {
		a.logger.Error("startup entitlement continuity scan failed", "error", err)
	}
	if err := a.catalog.ProcessDueAutoRenewals(ctx, startupNow); err != nil {
		a.logger.Error("startup automatic renewal scan failed", "error", err)
	}
	if err := a.store.EnqueueDueEntitlementTransitions(ctx, startupNow); err != nil {
		a.logger.Error("startup entitlement transition scan failed", "error", err)
	}
	if err := a.outbox.Drain(ctx, 50); err != nil && !errors.Is(err, context.Canceled) {
		a.logger.Error("startup outbox drain failed", "error", err)
	}
	if err := a.store.EnqueueDueEntitlementTransitions(ctx, startupNow); err != nil {
		a.logger.Error("startup successor transition scan failed", "error", err)
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
			if err := a.catalog.ProcessDueAutoRenewals(ctx, now.UTC()); err != nil {
				a.logger.Error("automatic renewal scan failed", "error", err)
			}
			if err := a.store.EnqueueDueEntitlementTransitions(ctx, now.UTC()); err != nil {
				a.logger.Error("entitlement transition scan failed", "error", err)
			}
			if err := a.outbox.Drain(ctx, 50); err != nil && !errors.Is(err, context.Canceled) {
				a.logger.Error("rollover settlement drain failed", "error", err)
			}
			if err := a.store.EnqueueDueEntitlementTransitions(ctx, now.UTC()); err != nil {
				a.logger.Error("successor transition scan failed", "error", err)
			}
			if err := a.store.ExpireStalePaymentOrders(ctx, now.UTC()); err != nil {
				a.logger.Error("payment expiry scan failed", "error", err)
			}
		case <-maintenanceTicker.C:
			a.configureTelegram(ctx)
		case <-starsTicker.C:
			a.reconcileStars(ctx)
		case now := <-statisticsTimer.C:
			a.refreshProductStatistics(ctx, now.UTC())
			statisticsTimer.Reset(time.Until(nextStatisticsRefresh(time.Now())))
		case <-maintenanceTimer.C:
			maintenanceCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
			maintenanceNow := time.Now().UTC()
			if err := a.maintenance.Run(maintenanceCtx, maintenanceNow); err != nil && !errors.Is(err, context.Canceled) {
				a.logger.Error("daily maintenance failed", "error", err)
			} else if err := database.PruneExpansionBackups(a.config.DatabasePath, maintenanceNow); err != nil {
				a.logger.Error("migration snapshot retention failed", "error", err)
			}
			cancel()
			maintenanceTimer.Reset(time.Until(a.nextMaintenance(time.Now())))
		}
	}
}

func (a *Application) nextMaintenance(now time.Time) time.Time {
	local := now.In(a.config.Timezone)
	next := time.Date(local.Year(), local.Month(), local.Day(), a.config.BackupHour, 0, 0, 0, a.config.Timezone)
	if !next.After(local) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}
