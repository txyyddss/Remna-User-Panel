package statistics

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/rollover"
)

func (s *Service) monthlyAverageUsageBPS(ctx context.Context, now time.Time) (int, error) {
	members, err := s.repository.ActiveMemberUsageForStatistics(ctx, now)
	if err != nil {
		return 0, fmt.Errorf("list active members for usage statistics: %w", err)
	}
	if len(members) == 0 {
		return 0, nil
	}

	var total float64
	count := 0
	for _, member := range members {
		purchase := member.Purchase
		start := purchase.ValidFrom.UTC()
		end := now
		if end.After(purchase.ValidUntil) {
			end = purchase.ValidUntil.UTC()
		}
		if !end.After(start) {
			continue
		}
		snapshot, snapshotErr := s.provider.UsageSnapshotForRollover(ctx, member.RemoteUserID, start, end)
		if errors.Is(snapshotErr, rollover.ErrRemoteUserMissing) {
			continue
		}
		if snapshotErr != nil {
			continue
		}
		projection := rollover.ProjectUsage(purchase, purchase.RolloverMinRemainingBPS, snapshot, now)
		if projection.Term == nil || projection.ActualUsedTrafficBytes == nil || projection.Term.AllocatedTrafficBytes <= 0 {
			continue
		}
		ratio := float64(*projection.ActualUsedTrafficBytes) * 10000 / float64(projection.Term.AllocatedTrafficBytes)
		if ratio < 0 || math.IsNaN(ratio) || math.IsInf(ratio, 0) {
			continue
		}
		total += ratio
		count++
	}
	if count == 0 {
		return 0, nil
	}
	return int(math.Round(total / float64(count))), nil
}
