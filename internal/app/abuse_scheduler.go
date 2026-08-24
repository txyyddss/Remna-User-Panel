package app

import (
	"context"
	"errors"
	"time"
)

const abuseProcessingTimeout = 5 * time.Minute

func (a *Application) runAbuseScheduler(ctx context.Context) {
	if a.abuse == nil {
		return
	}
	if err := a.abuse.RecoverProcessing(ctx); err != nil && !errors.Is(err, context.Canceled) {
		a.logger.Error("startup abuse claim recovery failed", "error", err)
	}
	a.processAbuseLogs(ctx, time.Now().UTC())
	timer := time.NewTimer(time.Until(nextAbuseProcessing(time.Now())))
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-timer.C:
			a.processAbuseLogs(ctx, now.UTC())
			timer.Reset(time.Until(nextAbuseProcessing(time.Now())))
		}
	}
}

func (a *Application) processAbuseLogs(ctx context.Context, now time.Time) {
	processCtx, cancel := context.WithTimeout(ctx, abuseProcessingTimeout)
	defer cancel()
	if err := a.abuse.Process(processCtx, now.UTC()); err != nil && !errors.Is(err, context.Canceled) {
		a.logger.Error("30-minute abuse processing failed", "error", err)
	}
}

func nextAbuseProcessing(now time.Time) time.Time {
	return now.UTC().Truncate(30 * time.Minute).Add(30 * time.Minute)
}
