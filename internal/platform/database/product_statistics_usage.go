package database

import (
	"context"
	"fmt"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

// ActiveMemberPurchasesForStatistics returns one current combo per non-admin
// member. Provider usage remains live and is never persisted by this query.
func (s *Store) ActiveMemberPurchasesForStatistics(ctx context.Context, now time.Time) ([]model.Purchase, error) {
	rows, err := s.db.QueryContext(ctx, purchaseSelect+`
		JOIN users member ON member.id=purchases.user_id
		WHERE member.role='user' AND purchases.status IN ('active','activating')
		AND purchases.valid_from<=? AND purchases.valid_until>?
		ORDER BY purchases.user_id,purchases.valid_from DESC,purchases.created_at DESC`, stamp(now), stamp(now))
	if err != nil {
		return nil, fmt.Errorf("list active member purchases for statistics: %w", err)
	}
	defer func() { _ = rows.Close() }()

	seen := make(map[string]struct{})
	result := make([]model.Purchase, 0)
	for rows.Next() {
		purchase, scanErr := scanPurchase(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		if _, exists := seen[purchase.UserID]; exists {
			continue
		}
		seen[purchase.UserID] = struct{}{}
		result = append(result, purchase)
	}
	return result, rows.Err()
}
