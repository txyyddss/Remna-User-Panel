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
		return copy.NoSubscription
	}
	purchase := dashboard.ActivePurchase
	names := s.telegramSquadNames(ctx, purchase.SquadUUIDs)
	return botcommands.FormatCombo(copy, purchase, names, s.telegramRolloverState(ctx, user, *purchase, copy))
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

func (s *Server) telegramRolloverState(ctx context.Context, user model.User, purchase model.Purchase, copy botcommands.Copy) string {
	if !purchase.AutoRenewEnabled {
		return copy.RolloverCannot
	}
	projection, err := s.deps.Catalog.RolloverProjection(ctx, user, purchase.ID)
	if err != nil || projection.WarningCode != nil || projection.PredictedRollover == nil {
		return copy.RolloverUnavailable
	}
	minor, err := strconv.ParseInt(projection.PredictedRollover.Minor, 10, 64)
	if err != nil {
		return copy.RolloverUnavailable
	}
	if minor > 0 {
		return copy.RolloverWill
	}
	return copy.RolloverWillNot
}
