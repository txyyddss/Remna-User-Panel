package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/txyyddss/Remna-User-Panel/internal/activity"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
	"math"
	"strings"
	"time"
)

func (s *Store) PlayLuckyDraw(ctx context.Context, userID, drawID, idempotencyKey string, rng activity.RandomSource, now time.Time) (activity.DrawResult, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(drawID) == "" || strings.TrimSpace(idempotencyKey) == "" || len(idempotencyKey) > 128 || rng == nil {
		return activity.DrawResult{}, activity.ErrInvalidInput
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	now = now.UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return activity.DrawResult{}, err
	}
	defer func() { _ = tx.Rollback() }()
	if existing, loadErr := drawResultByKeyTx(ctx, tx, userID, idempotencyKey); loadErr == nil {
		existing.Replayed = true
		return existing, nil
	} else if !errors.Is(loadErr, ErrNotFound) {
		return activity.DrawResult{}, loadErr
	}
	draw, err := luckyDrawByID(ctx, tx, drawID, true)
	if err != nil {
		return activity.DrawResult{}, err
	}
	draw.Prizes, err = luckyPrizes(ctx, tx, draw.ID, true)
	if err != nil {
		return activity.DrawResult{}, err
	}
	if len(draw.Prizes) == 0 {
		return activity.DrawResult{}, ErrConflict
	}
	maximumDeduction := draw.MaximumPrizeDeduction()
	if maximumDeduction > math.MaxInt64-draw.FeeMinor {
		return activity.DrawResult{}, activity.ErrInvalidInput
	}
	currentBalance, err := balanceTx(ctx, tx, userID)
	if err != nil {
		return activity.DrawResult{}, err
	}
	if currentBalance < draw.FeeMinor+maximumDeduction {
		return activity.DrawResult{}, ErrInsufficientBalance
	}
	resultID, err := ids.New()
	if err != nil {
		return activity.DrawResult{}, err
	}
	balance, err := changeBalanceTx(ctx, tx, userID, -draw.FeeMinor, now)
	if err != nil {
		return activity.DrawResult{}, err
	}
	var totalWeight int64
	for _, prize := range draw.Prizes {
		if totalWeight > math.MaxInt64-prize.Weight {
			return activity.DrawResult{}, activity.ErrInvalidInput
		}
		totalWeight += prize.Weight
	}
	roll, err := rng.Int63n(totalWeight)
	if err != nil {
		return activity.DrawResult{}, fmt.Errorf("select lucky-draw prize: %w", err)
	}
	selected := draw.Prizes[len(draw.Prizes)-1]
	for _, prize := range draw.Prizes {
		if roll < prize.Weight {
			selected = prize
			break
		}
		roll -= prize.Weight
	}
	if selected.StockRemaining != nil {
		result, updateErr := tx.ExecContext(ctx, `UPDATE activity_lucky_prizes SET stock_remaining=stock_remaining-1 WHERE id=? AND stock_remaining>0`, selected.ID)
		if updateErr != nil {
			return activity.DrawResult{}, updateErr
		}
		if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
			if rowsErr != nil {
				return activity.DrawResult{}, rowsErr
			}
			return activity.DrawResult{}, ErrConflict
		}
	}
	switch selected.Reward.Kind {
	case activity.RewardNone:
	case activity.RewardTXBDelta:
		balance, err = changeBalanceTx(ctx, tx, userID, selected.Reward.TXBDeltaMinor, now)
		if err != nil {
			return activity.DrawResult{}, err
		}
	case activity.RewardCouponGrant:
		coupon, loadErr := couponByID(ctx, tx, selected.Reward.CouponID)
		if loadErr != nil {
			return activity.DrawResult{}, loadErr
		}
		if err := couponAvailable(coupon, now); err != nil {
			return activity.DrawResult{}, err
		}
		if _, err := grantCouponTx(ctx, tx, userID, coupon, "activity_draw", resultID, now); err != nil {
			return activity.DrawResult{}, err
		}
	case activity.RewardSubscriptionExtension:
		if err := applySubscriptionExtensionTx(ctx, tx, userID, selected.Reward.ExtensionDays, "activity_draw", resultID, now); err != nil {
			return activity.DrawResult{}, err
		}
	default:
		return activity.DrawResult{}, activity.ErrInvalidInput
	}
	snapshot, err := json.Marshal(draw)
	if err != nil {
		return activity.DrawResult{}, err
	}
	rewardPayload, err := json.Marshal(selected.Reward)
	if err != nil {
		return activity.DrawResult{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO activity_draw_results(id,user_id,draw_id,prize_id,prize_name,fee_minor,reward_kind,reward_payload,balance_after_minor,configuration_snapshot,idempotency_key,created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, resultID, userID, draw.ID, selected.ID, selected.Name, draw.FeeMinor, selected.Reward.Kind, string(rewardPayload), balance, string(snapshot), idempotencyKey, stamp(now))
	if err != nil {
		return activity.DrawResult{}, fmt.Errorf("record lucky-draw result: %w", err)
	}
	if draw.FeeMinor != 0 {
		feeBalance := balance
		if selected.Reward.Kind == activity.RewardTXBDelta {
			feeBalance -= selected.Reward.TXBDeltaMinor
		}
		if _, err := insertLedgerTx(ctx, tx, userID, -draw.FeeMinor, feeBalance, "activity_draw_fee", resultID, draw.Name, now); err != nil {
			return activity.DrawResult{}, err
		}
	}
	if selected.Reward.Kind == activity.RewardTXBDelta {
		if _, err := insertLedgerTx(ctx, tx, userID, selected.Reward.TXBDeltaMinor, balance, "activity_draw_reward", resultID, selected.Name, now); err != nil {
			return activity.DrawResult{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return activity.DrawResult{}, err
	}
	return s.drawResultByID(ctx, resultID)
}

