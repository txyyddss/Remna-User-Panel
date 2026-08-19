package statistics

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/rollover"
)

func (s *Service) monthlyAverageUsageBPS(ctx context.Context, now time.Time) (int, error) {
	purchases, err := s.repository.ActiveMemberPurchasesForStatistics(ctx, now)
	if err != nil {
		return 0, fmt.Errorf("list active members for usage statistics: %w", err)
	}
	if len(purchases) == 0 {
		return 0, nil
	}

	windowFloor := now.AddDate(0, 0, -30)
	var total float64
	count := 0
	for _, purchase := range purchases {
		user, loadErr := s.repository.UserForPurchase(ctx, purchase.ID)
		if loadErr != nil {
			return 0, fmt.Errorf("load usage member for purchase %s: %w", purchase.ID, loadErr)
		}
		if user.Role != "user" {
			continue
		}
		if user.RemnaUserID == nil || strings.TrimSpace(*user.RemnaUserID) == "" {
			continue
		}
		start := windowFloor
		if purchase.ValidFrom.After(start) {
			start = purchase.ValidFrom.UTC()
		}
		if !now.After(start) {
			continue
		}
		snapshot, snapshotErr := s.provider.UsageSnapshotForRollover(ctx, *user.RemnaUserID, start, now)
		if errors.Is(snapshotErr, rollover.ErrRemoteUserMissing) {
			continue
		}
		if snapshotErr != nil {
			continue
		}
		window := purchase
		window.ValidFrom, window.ValidUntil = start, now
		projection := rollover.ProjectUsage(window, purchase.RolloverMinRemainingBPS, snapshot, now)
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
