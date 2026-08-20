package httpapi

import (
	"context"
	"strconv"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/botcommands"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

func (s *Server) telegramSubscriptionReply(ctx context.Context, user model.User, copy botcommands.Copy) string {
	dashboard, err := s.deps.Catalog.Dashboard(ctx, user)
	if err != nil {
		return ""
	}
	return botcommands.FormatSubscription(copy, dashboard, time.Now().UTC())
}

func (s *Server) telegramComboReply(ctx context.Context, user model.User, copy botcommands.Copy) string {
	dashboard, err := s.deps.Catalog.Dashboard(ctx, user)
	if err != nil || dashboard.ActivePurchase == nil {
		return botcommands.FormatNoSubscription(copy)
	}
	purchase := dashboard.ActivePurchase
	names := s.telegramSquadNames(ctx, purchase.SquadUUIDs)
	return botcommands.FormatCombo(copy, purchase, names, s.telegramRolloverState(ctx, user, *purchase))
}

func (s *Server) telegramSquadNames(ctx context.Context, squadUUIDs []string) []string {
	catalog, err := s.deps.Catalog.Catalog(ctx)
	if err != nil {
		return nil
	}
	names := make(map[string]string, len(catalog.Addons))
	for _, combo := range catalog.Combos {
		for _, squad := range combo.IncludedSquads {
			names[squad.RemnaSquadUUID] = squad.Name
		}
	}
	for _, squad := range catalog.Addons {
		names[squad.RemnaSquadUUID] = squad.Name
	}
	result := make([]string, 0, len(squadUUIDs))
	for _, uuid := range squadUUIDs {
		if name := names[uuid]; name != "" {
			result = append(result, name)
		} else {
			result = append(result, uuid)
		}
	}
	return result
}

func (s *Server) telegramRolloverState(ctx context.Context, user model.User, purchase model.Purchase) botcommands.RolloverSummary {
	if !purchase.AutoRenewEnabled {
		return botcommands.RolloverSummary{State: botcommands.RolloverDisabled}
	}
	projection, err := s.deps.Catalog.RolloverProjection(ctx, user, purchase.ID)
	if err != nil {
		return botcommands.RolloverSummary{State: botcommands.RolloverUnavailable}
	}
	if projection.WarningCode != nil {
		if *projection.WarningCode == "AUTO_RENEWAL_DISABLED" {
			return botcommands.RolloverSummary{State: botcommands.RolloverDisabled}
		}
		return botcommands.RolloverSummary{State: botcommands.RolloverUnavailable}
	}
	if projection.PredictedRollover == nil {
		return botcommands.RolloverSummary{State: botcommands.RolloverIneligible}
	}
	minor, err := strconv.ParseInt(projection.PredictedRollover.Minor, 10, 64)
	if err != nil {
		return botcommands.RolloverSummary{State: botcommands.RolloverUnavailable}
	}
	if minor > 0 {
		amount := *projection.PredictedRollover
		return botcommands.RolloverSummary{State: botcommands.RolloverPredicted, Amount: &amount}
	}
	return botcommands.RolloverSummary{State: botcommands.RolloverIneligible}
}

func (s *Server) telegramCheckInAverage(ctx context.Context) *int64 {
	snapshot, err := s.deps.Statistics.Snapshot(ctx)
	if err != nil || snapshot.DatabaseGeneratedAt.IsZero() {
		return nil
	}
	minor, err := strconv.ParseInt(snapshot.Database.AverageCheckInReward.Minor, 10, 64)
	if err != nil {
		s.deps.Logger.Warn("parse cached average check-in reward", "value", snapshot.Database.AverageCheckInReward.Minor, "error", err)
		return nil
	}
	return &minor
}
