package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/txyyddss/Remna-User-Panel/internal/activity"
	"github.com/txyyddss/Remna-User-Panel/internal/coupons"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
	"strings"
	"time"
)

func (s *Store) SaveLuckyDraw(ctx context.Context, input activity.LuckyDrawInput, now time.Time) (activity.LuckyDraw, error) {
	if err := input.Validate(); err != nil {
		return activity.LuckyDraw{}, err
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	now = now.UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return activity.LuckyDraw{}, err
	}
	defer func() { _ = tx.Rollback() }()
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if input.ID == "" {
		input.ID, err = ids.New()
		if err != nil {
			return activity.LuckyDraw{}, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO activity_lucky_draws(id,name,description,enabled,fee_minor,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`, input.ID, input.Name, input.Description, boolInt(input.Enabled), input.FeeMinor, stamp(now), stamp(now))
	} else {
		var result sql.Result
		result, err = tx.ExecContext(ctx, `UPDATE activity_lucky_draws SET name=?,description=?,enabled=?,fee_minor=?,updated_at=? WHERE id=?`, input.Name, input.Description, boolInt(input.Enabled), input.FeeMinor, stamp(now), input.ID)
		if err == nil {
			if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
				return activity.LuckyDraw{}, rowsErr
			} else if affected == 0 {
				return activity.LuckyDraw{}, ErrNotFound
			}
		}
	}
	if err != nil {
		return activity.LuckyDraw{}, fmt.Errorf("save lucky draw: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM activity_lucky_prizes WHERE draw_id=?`, input.ID); err != nil {
		return activity.LuckyDraw{}, fmt.Errorf("replace lucky-draw prizes: %w", err)
	}
	for position, prize := range input.Prizes {
		if prize.Reward.Kind == activity.RewardCouponGrant {
			coupon, loadErr := couponByID(ctx, tx, prize.Reward.CouponID)
			if loadErr != nil {
				return activity.LuckyDraw{}, loadErr
			}
			if loadErr = couponAvailable(coupon, now); loadErr != nil || (coupon.Kind != coupons.KindPurchaseRecurring && coupon.Kind != coupons.KindPurchaseOnce) {
				return activity.LuckyDraw{}, ErrConflict
			}
		}
		prizeID := prize.ID
		if prizeID == "" {
			prizeID, err = ids.New()
			if err != nil {
				return activity.LuckyDraw{}, err
			}
		}
		payload, marshalErr := json.Marshal(prize.Reward)
		if marshalErr != nil {
			return activity.LuckyDraw{}, marshalErr
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO activity_lucky_prizes(id,draw_id,name,position,weight,stock_remaining,reward_kind,reward_payload) VALUES(?,?,?,?,?,?,?,?)`,
			prizeID, input.ID, strings.TrimSpace(prize.Name), position, prize.Weight, nullableInt64(prize.StockRemaining), prize.Reward.Kind, string(payload))
		if err != nil {
			return activity.LuckyDraw{}, fmt.Errorf("save lucky-draw prize: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return activity.LuckyDraw{}, err
	}
	draw, err := luckyDrawByID(ctx, s.db, input.ID, false)
	if err != nil {
		return activity.LuckyDraw{}, err
	}
	draw.Prizes, err = luckyPrizes(ctx, s.db, input.ID, false)
	return draw, err
}

// ListLuckyDraws returns enabled member draws or all administrator draws.

func (s *Store) ListLuckyDraws(ctx context.Context, enabledOnly bool) ([]activity.LuckyDraw, error) {
	query := `SELECT id,name,description,enabled,fee_minor,created_at,updated_at FROM activity_lucky_draws`
	if enabledOnly {
		query += ` WHERE enabled=1`
	}
	query += ` ORDER BY name,id`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	draws := make([]activity.LuckyDraw, 0)
	for rows.Next() {
		draw, scanErr := scanLuckyDraw(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		draws = append(draws, draw)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for index := range draws {
		draws[index].Prizes, err = luckyPrizes(ctx, s.db, draws[index].ID, false)
		if err != nil {
			return nil, err
		}
	}
	return draws, nil
}

// PlayLuckyDraw atomically verifies worst-case coverage, charges, selects, and rewards.
