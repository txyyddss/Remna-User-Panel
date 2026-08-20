package app

import (
	"context"
	"errors"
	"time"
)

const notificationScanTimeout = 4 * time.Minute

func (a *Application) runNotificationScheduler(ctx context.Context) {
	a.scanUserNotifications(ctx, time.Now().UTC())
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			a.scanUserNotifications(ctx, now.UTC())
		}
	}
}

func (a *Application) scanUserNotifications(ctx context.Context, now time.Time) {
	if a.notifications == nil {
		return
	}
	scanCtx, cancel := context.WithTimeout(ctx, notificationScanTimeout)
	defer cancel()
	if err := a.notifications.Scan(scanCtx, now); err != nil && !errors.Is(err, context.Canceled) {
		a.logger.Error("user notification scan failed", "error", err)
	}
}
