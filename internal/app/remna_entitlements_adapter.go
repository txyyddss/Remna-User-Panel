package app

import (
	"context"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/entitlements"
	"github.com/txyyddss/Remna-User-Panel/internal/integrations/remnawave"
	"github.com/txyyddss/Remna-User-Panel/internal/rollover"
)

func (a remnaAdapter) ApplyEntitlement(ctx context.Context, remoteID string, trafficLimitBytes int64, resetStrategy string, squadUUIDs []string) error {
	userID, err := remnaUserID(remoteID)
	if err != nil {
		return err
	}
	status := remnawave.UserStatusActive
	strategy := remnawave.TrafficLimitStrategy(resetStrategy)
	expires := time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)
	squads := append([]string(nil), squadUUIDs...)
	return remnaExecute(ctx, a, func(callCtx context.Context, client remnaClient) error {
		_, callErr := client.UpdateUser(callCtx, remnawave.UpdateUserRequest{
			ID: userID, Status: &status, TrafficLimitBytes: &trafficLimitBytes,
			TrafficLimitStrategy: &strategy, ExpireAt: &expires,
			ActiveInternalSquads: &squads, ClearExternalSquad: true,
		})
		return callErr
	})
}

func (a remnaAdapter) ResetTraffic(ctx context.Context, remoteID string) error {
	userID, err := remnaUserID(remoteID)
	if err != nil {
		return err
	}
	return remnaExecute(ctx, a, func(callCtx context.Context, client remnaClient) error {
		_, callErr := client.ResetTraffic(callCtx, userID)
		return callErr
	})
}

func (a remnaAdapter) RemoveEntitlement(ctx context.Context, remoteID string) error {
	userID, err := remnaUserID(remoteID)
	if err != nil {
		return err
	}
	status := remnawave.UserStatusDisabled
	limit := int64(0)
	strategy := remnawave.TrafficNoReset
	expires := time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)
	squads := []string{}
	return remnaExecute(ctx, a, func(callCtx context.Context, client remnaClient) error {
		_, callErr := client.UpdateUser(callCtx, remnawave.UpdateUserRequest{
			ID: userID, Status: &status, TrafficLimitBytes: &limit,
			TrafficLimitStrategy: &strategy, ExpireAt: &expires,
			ActiveInternalSquads: &squads, ClearExternalSquad: true,
		})
		return callErr
	})
}

func (a remnaAdapter) QuiesceForRollover(ctx context.Context, remoteID string) error {
	userID, err := remnaUserID(remoteID)
	if err != nil {
		return err
	}
	status := remnawave.UserStatusDisabled
	err = remnaExecute(ctx, a, func(callCtx context.Context, client remnaClient) error {
		_, callErr := client.UpdateUser(callCtx, remnawave.UpdateUserRequest{ID: userID, Status: &status})
		return callErr
	})
	if remnawave.IsNotFound(err) {
		return rollover.ErrRemoteUserMissing
	}
	return err
}

func (a remnaAdapter) TrafficForRollover(ctx context.Context, remoteID string) (int64, int64, error) {
	userID, err := remnaUserID(remoteID)
	if err != nil {
		return 0, 0, err
	}
	user, err := remnaCall(ctx, a, func(callCtx context.Context, client remnaClient) (*remnawave.User, error) {
		return client.GetUserByID(callCtx, userID)
	})
	if remnawave.IsNotFound(err) {
		return 0, 0, rollover.ErrRemoteUserMissing
	}
	if err != nil {
		return 0, 0, err
	}
	return user.TrafficLimitBytes, user.UserTraffic.UsedTrafficBytes, nil
}

func (a remnaAdapter) UsageSnapshotForRollover(ctx context.Context, remoteID string, start, end time.Time) (rollover.UsageSnapshot, error) {
	userID, err := remnaUserID(remoteID)
	if err != nil {
		return rollover.UsageSnapshot{}, err
	}
	user, err := remnaCall(ctx, a, func(callCtx context.Context, client remnaClient) (*remnawave.User, error) {
		return client.GetUserByID(callCtx, userID)
	})
	if remnawave.IsNotFound(err) {
		return rollover.UsageSnapshot{}, rollover.ErrRemoteUserMissing
	}
	if err != nil {
		return rollover.UsageSnapshot{}, err
	}
	stats, err := remnaCall(ctx, a, func(callCtx context.Context, client remnaClient) (*remnawave.UserStats, error) {
		return client.GetUserStats(callCtx, userID, start.UTC(), end.Add(-time.Nanosecond).UTC(), 20)
	})
	if remnawave.IsNotFound(err) {
		return rollover.UsageSnapshot{}, rollover.ErrRemoteUserMissing
	}
	if err != nil {
		return rollover.UsageSnapshot{}, err
	}
	daily := make([]rollover.DailyUsage, 0, len(stats.Categories))
	for index, category := range stats.Categories {
		date, parseErr := time.Parse(time.DateOnly, category)
		if parseErr != nil {
			continue
		}
		if index < len(stats.SparklineData) {
			daily = append(daily, rollover.DailyUsage{Date: date, Bytes: stats.SparklineData[index]})
		}
	}
	currentUsed := user.UserTraffic.UsedTrafficBytes
	return rollover.UsageSnapshot{
		LimitBytes: user.TrafficLimitBytes, Strategy: string(user.TrafficLimitStrategy),
		LastResetAt: user.LastTrafficResetAt, CurrentUsedBytes: &currentUsed, Daily: daily,
	}, nil
}

var _ entitlements.RemnawaveClient = remnaAdapter{}
var _ rollover.Remote = remnaAdapter{}
