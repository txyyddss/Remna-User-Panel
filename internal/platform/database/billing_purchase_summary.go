package database

import (
	"context"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

// ActiveAndQueuedPurchases returns the current home-screen entitlement summary.
func (s *Store) ActiveAndQueuedPurchases(ctx context.Context, userID string, now time.Time) (*model.Purchase, *model.Purchase, error) {
	purchases, err := s.ListPurchases(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	var active, queued *model.Purchase
	for index := range purchases {
		purchase := purchases[index]
		if (purchase.Status == "active" || purchase.Status == "activating") && !purchase.ValidUntil.Before(now) && active == nil {
			copy := purchase
			active = &copy
		}
		if purchase.Status == "queued" && queued == nil {
			copy := purchase
			queued = &copy
		}
	}
	return active, queued, nil
}
