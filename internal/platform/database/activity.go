package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/activity"
	"github.com/txyyddss/Remna-User-Panel/internal/coupons"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
)

const activityGameSelect = `SELECT id,name,icon,description,enabled,win_chance_bps,minimum_stake_minor,maximum_stake_minor,
	return_multiplier_bps,created_at,updated_at FROM activity_games`

// SaveActivityGame creates or updates one betting game.
func (s *Store) SaveActivityGame(ctx context.Context, input activity.GameInput, now time.Time) (activity.Game, error) {
	if err := input.Validate(); err != nil {
		return activity.Game{}, err
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Icon = strings.TrimSpace(input.Icon)
	input.Description = strings.TrimSpace(input.Description)
	now = now.UTC()
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	var err error
	if input.ID == "" {
		input.ID, err = ids.New()
		if err != nil {
			return activity.Game{}, err
		}
		_, err = s.db.ExecContext(ctx, `INSERT INTO activity_games(id,name,icon,description,enabled,win_chance_bps,minimum_stake_minor,maximum_stake_minor,return_multiplier_bps,created_at,updated_at)
			VALUES(?,?,?,?,?,?,?,?,?,?,?)`, input.ID, input.Name, input.Icon, input.Description, boolInt(input.Enabled), input.WinChanceBPS, input.MinimumStakeMinor,
			input.MaximumStakeMinor, input.ReturnMultiplierBPS, stamp(now), stamp(now))
	} else {
		var result sql.Result
		result, err = s.db.ExecContext(ctx, `UPDATE activity_games SET name=?,icon=?,description=?,enabled=?,win_chance_bps=?,minimum_stake_minor=?,maximum_stake_minor=?,return_multiplier_bps=?,updated_at=? WHERE id=?`,
			input.Name, input.Icon, input.Description, boolInt(input.Enabled), input.WinChanceBPS, input.MinimumStakeMinor, input.MaximumStakeMinor, input.ReturnMultiplierBPS, stamp(now), input.ID)
		if err == nil {
			if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
				return activity.Game{}, rowsErr
			} else if affected == 0 {
				return activity.Game{}, ErrNotFound
			}
		}
	}
	if err != nil {
		return activity.Game{}, fmt.Errorf("save activity game: %w", err)
	}
	return activityGameByID(ctx, s.db, input.ID, false)
}

// ListActivityGames returns enabled member games or every administrator game.
func (s *Store) ListActivityGames(ctx context.Context, enabledOnly bool) ([]activity.Game, error) {
	query := activityGameSelect
	if enabledOnly {
		query += ` WHERE enabled=1`
	}
	query += ` ORDER BY name,id`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list activity games: %w", err)
	}
	defer func() { _ = rows.Close() }()
	result := make([]activity.Game, 0)
	for rows.Next() {
		game, scanErr := scanActivityGame(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, game)
	}
	return result, rows.Err()
}

// PlaceActivityBet atomically debits the stake before selecting and recording an outcome.
func (s *Store) PlaceActivityBet(ctx context.Context, userID, gameID string, stakeMinor int64, idempotencyKey string, rng activity.RandomSource, now time.Time) (activity.BetResult, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(gameID) == "" || stakeMinor <= 0 || strings.TrimSpace(idempotencyKey) == "" || len(idempotencyKey) > 128 || rng == nil {
		return activity.BetResult{}, activity.ErrInvalidInput
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	now = now.UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return activity.BetResult{}, fmt.Errorf("begin activity bet: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if existing, loadErr := activityBetByKeyTx(ctx, tx, userID, idempotencyKey); loadErr == nil {
		existing.Replayed = true
		return existing, nil
	} else if !errors.Is(loadErr, ErrNotFound) {
		return activity.BetResult{}, loadErr
	}
	game, err := activityGameByID(ctx, tx, gameID, true)
	if err != nil {
		return activity.BetResult{}, err
	}
	if stakeMinor < game.MinimumStakeMinor || stakeMinor > game.MaximumStakeMinor {
		return activity.BetResult{}, activity.ErrInvalidInput
	}
	betID, err := ids.New()
	if err != nil {
		return activity.BetResult{}, err
	}
	balance, err := changeBalanceTx(ctx, tx, userID, -stakeMinor, now)
	if err != nil {
		return activity.BetResult{}, err
	}
	roll, err := rng.Int63n(10_000)
	if err != nil {
		return activity.BetResult{}, fmt.Errorf("select activity bet outcome: %w", err)
	}
	won := roll < int64(game.WinChanceBPS)
	payout := int64(0)
	if won {
		payout, err = fixedMultiplyFloor(stakeMinor, game.ReturnMultiplierBPS, 10_000)
		if err != nil {
			return activity.BetResult{}, err
		}
		balance, err = changeBalanceTx(ctx, tx, userID, payout, now)
		if err != nil {
			return activity.BetResult{}, err
		}
	}
	snapshot, err := json.Marshal(game)
	if err != nil {
		return activity.BetResult{}, fmt.Errorf("encode activity game snapshot: %w", err)
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO activity_bets(id,user_id,game_id,stake_minor,won,payout_minor,balance_after_minor,configuration_snapshot,idempotency_key,created_at)
		VALUES(?,?,?,?,?,?,?,?,?,?)`, betID, userID, game.ID, stakeMinor, boolInt(won), payout, balance, string(snapshot), idempotencyKey, stamp(now))
	if err != nil {
		return activity.BetResult{}, fmt.Errorf("record activity bet: %w", err)
	}
	stakeBalance := balance
	if won {
		stakeBalance -= payout
	}
	if _, err := insertLedgerTx(ctx, tx, userID, -stakeMinor, stakeBalance, "activity_bet_stake", betID, game.Name, now); err != nil {
		return activity.BetResult{}, err
	}
	if won {
		if _, err := insertLedgerTx(ctx, tx, userID, payout, balance, "activity_bet_payout", betID, game.Name, now); err != nil {
			return activity.BetResult{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return activity.BetResult{}, fmt.Errorf("commit activity bet: %w", err)
	}
	return s.activityBetByID(ctx, betID)
}

// ClaimDailyActivity credits at most one reward per configured local calendar date.
func (s *Store) ClaimDailyActivity(ctx context.Context, userID, localDate, timezone string, rewardMinor int64, now time.Time) (activity.DailyCheckIn, error) {
	return s.claimDailyActivity(ctx, userID, localDate, timezone, rewardMinor, rewardMinor, nil, now)
}

// ClaimDailyActivityRange chooses the inclusive reward exactly once, after the
// idempotency row check and while holding the write transaction lock.
func (s *Store) ClaimDailyActivityRange(ctx context.Context, userID, localDate, timezone string, minimumMinor, maximumMinor int64, rng activity.RandomSource, now time.Time) (activity.DailyCheckIn, error) {
	return s.claimDailyActivity(ctx, userID, localDate, timezone, minimumMinor, maximumMinor, rng, now)
}

func (s *Store) claimDailyActivity(ctx context.Context, userID, localDate, timezone string, minimumMinor, maximumMinor int64, rng activity.RandomSource, now time.Time) (activity.DailyCheckIn, error) {
	location, err := time.LoadLocation(timezone)
	if err != nil || minimumMinor < 0 || maximumMinor < minimumMinor || maximumMinor-minimumMinor == math.MaxInt64 || strings.TrimSpace(userID) == "" {
		return activity.DailyCheckIn{}, activity.ErrInvalidInput
	}
	if _, err := time.Parse(time.DateOnly, localDate); err != nil || now.In(location).Format(time.DateOnly) != localDate {
		return activity.DailyCheckIn{}, activity.ErrInvalidInput
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	now = now.UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return activity.DailyCheckIn{}, err
	}
	defer func() { _ = tx.Rollback() }()
	if existing, loadErr := dailyCheckInByDateTx(ctx, tx, userID, localDate); loadErr == nil {
		existing.AlreadyClaimed = true
		return existing, nil
	} else if !errors.Is(loadErr, ErrNotFound) {
		return activity.DailyCheckIn{}, loadErr
	}
	rewardMinor := minimumMinor
	if maximumMinor > minimumMinor {
		if rng == nil {
			return activity.DailyCheckIn{}, activity.ErrInvalidInput
		}
		offset, randomErr := rng.Int63n(maximumMinor - minimumMinor + 1)
		if randomErr != nil {
			return activity.DailyCheckIn{}, randomErr
		}
		rewardMinor += offset
	}
	checkInID, err := ids.New()
	if err != nil {
		return activity.DailyCheckIn{}, err
	}
	balance, err := changeBalanceTx(ctx, tx, userID, rewardMinor, now)
	if err != nil {
		return activity.DailyCheckIn{}, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO activity_daily_checkins(id,user_id,local_date,timezone,reward_minor,balance_after_minor,created_at) VALUES(?,?,?,?,?,?,?)`,
		checkInID, userID, localDate, timezone, rewardMinor, balance, stamp(now))
	if err != nil {
		return activity.DailyCheckIn{}, fmt.Errorf("record daily activity: %w", err)
	}
	if _, err := insertLedgerTx(ctx, tx, userID, rewardMinor, balance, "activity_daily_checkin", checkInID, localDate, now); err != nil {
		return activity.DailyCheckIn{}, err
	}
	if err := tx.Commit(); err != nil {
		return activity.DailyCheckIn{}, err
	}
	return s.dailyCheckInByID(ctx, checkInID)
}

// SaveLuckyDraw atomically replaces a draw and its ordered prize pool.
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

// ListActivityHistory returns bounded newest-first member outcomes.
func (s *Store) ListActivityHistory(ctx context.Context, userID string, limit int) (activity.History, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	history := activity.History{
		Bets:     make([]activity.BetResult, 0),
		CheckIns: make([]activity.DailyCheckIn, 0),
		Draws:    make([]activity.DrawResult, 0),
	}
	betRows, err := s.db.QueryContext(ctx, activityBetSelect+` WHERE user_id=? ORDER BY created_at DESC,id DESC LIMIT ?`, userID, limit)
	if err != nil {
		return activity.History{}, err
	}
	for betRows.Next() {
		result, scanErr := scanActivityBet(betRows)
		if scanErr != nil {
			_ = betRows.Close()
			return activity.History{}, scanErr
		}
		history.Bets = append(history.Bets, result)
	}
	if err := betRows.Err(); err != nil {
		_ = betRows.Close()
		return activity.History{}, err
	}
	_ = betRows.Close()
	checkInRows, err := s.db.QueryContext(ctx, dailyCheckInSelect+` WHERE user_id=? ORDER BY created_at DESC,id DESC LIMIT ?`, userID, limit)
	if err != nil {
		return activity.History{}, err
	}
	for checkInRows.Next() {
		result, scanErr := scanDailyCheckIn(checkInRows)
		if scanErr != nil {
			_ = checkInRows.Close()
			return activity.History{}, scanErr
		}
		history.CheckIns = append(history.CheckIns, result)
	}
	if err := checkInRows.Err(); err != nil {
		_ = checkInRows.Close()
		return activity.History{}, err
	}
	_ = checkInRows.Close()
	drawRows, err := s.db.QueryContext(ctx, drawResultSelect+` WHERE user_id=? ORDER BY created_at DESC,id DESC LIMIT ?`, userID, limit)
	if err != nil {
		return activity.History{}, err
	}
	defer func() { _ = drawRows.Close() }()
	for drawRows.Next() {
		result, scanErr := scanDrawResult(drawRows)
		if scanErr != nil {
			return activity.History{}, scanErr
		}
		history.Draws = append(history.Draws, result)
	}
	if err := drawRows.Err(); err != nil {
		return activity.History{}, err
	}
	return history, nil
}

// GroupMessageRewardStatus returns one user's current local-day group-message progress.
func (s *Store) GroupMessageRewardStatus(ctx context.Context, userID, localDate string, threshold int, rewardMinor int64) (activity.GroupMessageRewardStatus, error) {
	status, err := groupMessageRewardStatusQuery(ctx, s.db, userID, localDate, threshold, rewardMinor)
	if err != nil {
		return activity.GroupMessageRewardStatus{}, fmt.Errorf("load group-message reward status: %w", err)
	}
	return status, nil
}

// RecordGroupMessage idempotently counts one subscribed user's group message and credits the daily reward.
func (s *Store) RecordGroupMessage(ctx context.Context, userID string, chatID, messageID int64, localDate, timezone string, threshold int, rewardMinor int64, now time.Time) (activity.GroupMessageRewardResult, error) {
	if strings.TrimSpace(userID) == "" || chatID == 0 || messageID <= 0 || threshold < 0 || rewardMinor < 0 {
		return activity.GroupMessageRewardResult{}, activity.ErrInvalidInput
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	now = now.UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return activity.GroupMessageRewardResult{}, fmt.Errorf("begin group-message reward: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if threshold == 0 || rewardMinor == 0 {
		return activity.GroupMessageRewardResult{Status: activity.GroupMessageRewardStatus{
			Enabled: false, LocalDate: localDate, Threshold: threshold, RewardMinor: rewardMinor,
		}}, nil
	}

	var existingCounted int
	lookupErr := tx.QueryRowContext(ctx, `SELECT counted FROM activity_group_message_events WHERE chat_id=? AND message_id=?`, chatID, messageID).Scan(&existingCounted)
	if lookupErr == nil {
		if err := tx.Rollback(); err != nil {
			return activity.GroupMessageRewardResult{}, fmt.Errorf("close group-message replay transaction: %w", err)
		}
		status, statusErr := s.GroupMessageRewardStatus(ctx, userID, localDate, threshold, rewardMinor)
		if statusErr != nil {
			return activity.GroupMessageRewardResult{}, statusErr
		}
		return activity.GroupMessageRewardResult{Status: status, Counted: existingCounted == 1, Replayed: true}, nil
	}
	if !errors.Is(lookupErr, sql.ErrNoRows) {
		return activity.GroupMessageRewardResult{}, fmt.Errorf("load group-message event: %w", lookupErr)
	}

	var subscribed int
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(
		SELECT 1 FROM purchases WHERE user_id=? AND status IN ('activating','active') AND valid_from<=? AND valid_until>?
	)`, userID, stamp(now), stamp(now)).Scan(&subscribed); err != nil {
		return activity.GroupMessageRewardResult{}, fmt.Errorf("check group-message subscription: %w", err)
	}
	counted := subscribed == 1
	if _, err := tx.ExecContext(ctx, `INSERT INTO activity_group_message_events(chat_id,message_id,user_id,local_date,counted,created_at) VALUES(?,?,?,?,?,?)`,
		chatID, messageID, userID, localDate, boolInt(counted), stamp(now)); err != nil {
		return activity.GroupMessageRewardResult{}, fmt.Errorf("record group-message event: %w", err)
	}
	if !counted {
		if err := tx.Commit(); err != nil {
			return activity.GroupMessageRewardResult{}, fmt.Errorf("commit ineligible group message: %w", err)
		}
		status, statusErr := s.GroupMessageRewardStatus(ctx, userID, localDate, threshold, rewardMinor)
		if statusErr != nil {
			return activity.GroupMessageRewardResult{}, statusErr
		}
		return activity.GroupMessageRewardResult{Status: status}, nil
	}

	var messageCount int
	var rewardedAt sql.NullString
	windowErr := tx.QueryRowContext(ctx, `SELECT message_count,rewarded_at FROM activity_group_message_windows WHERE user_id=? AND local_date=?`, userID, localDate).Scan(&messageCount, &rewardedAt)
	if errors.Is(windowErr, sql.ErrNoRows) {
		if _, err := tx.ExecContext(ctx, `INSERT INTO activity_group_message_windows(user_id,local_date,message_count,created_at,updated_at) VALUES(?,?,?,?,?)`,
			userID, localDate, 0, stamp(now), stamp(now)); err != nil {
			return activity.GroupMessageRewardResult{}, fmt.Errorf("create group-message window: %w", err)
		}
	} else if windowErr != nil {
		return activity.GroupMessageRewardResult{}, fmt.Errorf("load group-message window: %w", windowErr)
	}

	messageCount++
	if _, err := tx.ExecContext(ctx, `UPDATE activity_group_message_windows SET message_count=?,updated_at=? WHERE user_id=? AND local_date=?`,
		messageCount, stamp(now), userID, localDate); err != nil {
		return activity.GroupMessageRewardResult{}, fmt.Errorf("update group-message window: %w", err)
	}
	if messageCount >= threshold && !rewardedAt.Valid {
		rewardID, idErr := ids.New()
		if idErr != nil {
			return activity.GroupMessageRewardResult{}, idErr
		}
		balance, balanceErr := changeBalanceTx(ctx, tx, userID, rewardMinor, now)
		if balanceErr != nil {
			return activity.GroupMessageRewardResult{}, balanceErr
		}
		if _, err := tx.ExecContext(ctx, `UPDATE activity_group_message_windows SET rewarded_at=?,updated_at=? WHERE user_id=? AND local_date=? AND rewarded_at IS NULL`,
			stamp(now), stamp(now), userID, localDate); err != nil {
			return activity.GroupMessageRewardResult{}, fmt.Errorf("mark group-message reward: %w", err)
		}
		if _, err := insertLedgerTx(ctx, tx, userID, rewardMinor, balance, "activity_group_message_reward", rewardID, localDate, now); err != nil {
			return activity.GroupMessageRewardResult{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM activity_group_message_events WHERE created_at<?`, stamp(now.Add(-31*24*time.Hour))); err != nil {
		return activity.GroupMessageRewardResult{}, fmt.Errorf("prune group-message events: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return activity.GroupMessageRewardResult{}, fmt.Errorf("commit group-message reward: %w", err)
	}
	status, err := s.GroupMessageRewardStatus(ctx, userID, localDate, threshold, rewardMinor)
	if err != nil {
		return activity.GroupMessageRewardResult{}, err
	}
	return activity.GroupMessageRewardResult{Status: status, Counted: true, Replayed: false}, nil
}

func groupMessageRewardStatusQuery(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, userID, localDate string, threshold int, rewardMinor int64) (activity.GroupMessageRewardStatus, error) {
	status := activity.GroupMessageRewardStatus{Enabled: threshold > 0 && rewardMinor > 0, LocalDate: localDate, Threshold: threshold, RewardMinor: rewardMinor}
	var rewardedAt sql.NullString
	err := queryer.QueryRowContext(ctx, `SELECT message_count,rewarded_at FROM activity_group_message_windows WHERE user_id=? AND local_date=?`, userID, localDate).Scan(&status.MessageCount, &rewardedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return status, nil
	}
	if err != nil {
		return activity.GroupMessageRewardStatus{}, err
	}
	if rewardedAt.Valid {
		value, parseErr := parseStamp(rewardedAt.String)
		if parseErr != nil {
			return activity.GroupMessageRewardStatus{}, parseErr
		}
		status.Rewarded = true
		status.RewardedAt = &value
	}
	return status, nil
}

func activityGameByID(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, gameID string, enabledOnly bool) (activity.Game, error) {
	query := activityGameSelect + ` WHERE id=?`
	if enabledOnly {
		query += ` AND enabled=1`
	}
	return scanActivityGame(queryer.QueryRowContext(ctx, query, gameID))
}

func scanActivityGame(row rowScanner) (activity.Game, error) {
	var game activity.Game
	var enabled int
	var created, updated string
	if err := row.Scan(&game.ID, &game.Name, &game.Icon, &game.Description, &enabled, &game.WinChanceBPS, &game.MinimumStakeMinor, &game.MaximumStakeMinor,
		&game.ReturnMultiplierBPS, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return activity.Game{}, ErrNotFound
		}
		return activity.Game{}, err
	}
	game.Enabled = enabled == 1
	var err error
	if game.CreatedAt, err = parseStamp(created); err != nil {
		return activity.Game{}, err
	}
	game.UpdatedAt, err = parseStamp(updated)
	return game, err
}

const activityBetSelect = `SELECT id,user_id,game_id,stake_minor,won,payout_minor,balance_after_minor,configuration_snapshot,idempotency_key,created_at FROM activity_bets`

func (s *Store) activityBetByID(ctx context.Context, betID string) (activity.BetResult, error) {
	return scanActivityBet(s.db.QueryRowContext(ctx, activityBetSelect+` WHERE id=?`, betID))
}

func activityBetByKeyTx(ctx context.Context, tx *sql.Tx, userID, key string) (activity.BetResult, error) {
	return scanActivityBet(tx.QueryRowContext(ctx, activityBetSelect+` WHERE user_id=? AND idempotency_key=?`, userID, key))
}

func scanActivityBet(row rowScanner) (activity.BetResult, error) {
	var result activity.BetResult
	var won int
	var created string
	if err := row.Scan(&result.ID, &result.UserID, &result.GameID, &result.StakeMinor, &won, &result.PayoutMinor, &result.BalanceAfterMinor,
		&result.ConfigurationSnapshot, &result.IdempotencyKey, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return activity.BetResult{}, ErrNotFound
		}
		return activity.BetResult{}, err
	}
	result.Won = won == 1
	parsed, err := parseStamp(created)
	result.CreatedAt = parsed
	return result, err
}

const dailyCheckInSelect = `SELECT id,user_id,local_date,timezone,reward_minor,balance_after_minor,created_at FROM activity_daily_checkins`

func dailyCheckInByDateTx(ctx context.Context, tx *sql.Tx, userID, localDate string) (activity.DailyCheckIn, error) {
	return scanDailyCheckIn(tx.QueryRowContext(ctx, dailyCheckInSelect+` WHERE user_id=? AND local_date=?`, userID, localDate))
}

func (s *Store) dailyCheckInByID(ctx context.Context, checkInID string) (activity.DailyCheckIn, error) {
	return scanDailyCheckIn(s.db.QueryRowContext(ctx, dailyCheckInSelect+` WHERE id=?`, checkInID))
}

func scanDailyCheckIn(row rowScanner) (activity.DailyCheckIn, error) {
	var result activity.DailyCheckIn
	var created string
	if err := row.Scan(&result.ID, &result.UserID, &result.LocalDate, &result.Timezone, &result.RewardMinor, &result.BalanceAfterMinor, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return activity.DailyCheckIn{}, ErrNotFound
		}
		return activity.DailyCheckIn{}, err
	}
	parsed, err := parseStamp(created)
	result.CreatedAt = parsed
	return result, err
}

func luckyDrawByID(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, drawID string, enabledOnly bool) (activity.LuckyDraw, error) {
	query := `SELECT id,name,description,enabled,fee_minor,created_at,updated_at FROM activity_lucky_draws WHERE id=?`
	if enabledOnly {
		query += ` AND enabled=1`
	}
	return scanLuckyDraw(queryer.QueryRowContext(ctx, query, drawID))
}

func scanLuckyDraw(row rowScanner) (activity.LuckyDraw, error) {
	var draw activity.LuckyDraw
	var enabled int
	var created, updated string
	if err := row.Scan(&draw.ID, &draw.Name, &draw.Description, &enabled, &draw.FeeMinor, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return activity.LuckyDraw{}, ErrNotFound
		}
		return activity.LuckyDraw{}, err
	}
	draw.Enabled = enabled == 1
	var err error
	if draw.CreatedAt, err = parseStamp(created); err != nil {
		return activity.LuckyDraw{}, err
	}
	draw.UpdatedAt, err = parseStamp(updated)
	return draw, err
}

func luckyPrizes(ctx context.Context, queryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, drawID string, availableOnly bool) ([]activity.PrizeInput, error) {
	query := `SELECT id,name,weight,stock_remaining,reward_payload FROM activity_lucky_prizes WHERE draw_id=?`
	if availableOnly {
		query += ` AND (stock_remaining IS NULL OR stock_remaining>0)`
	}
	query += ` ORDER BY position`
	rows, err := queryer.QueryContext(ctx, query, drawID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := make([]activity.PrizeInput, 0)
	for rows.Next() {
		var prize activity.PrizeInput
		var stock sql.NullInt64
		var payload string
		if err := rows.Scan(&prize.ID, &prize.Name, &prize.Weight, &stock, &payload); err != nil {
			return nil, err
		}
		prize.StockRemaining = int64Pointer(stock)
		if err := json.Unmarshal([]byte(payload), &prize.Reward); err != nil {
			return nil, fmt.Errorf("decode lucky-draw reward: %w", err)
		}
		result = append(result, prize)
	}
	return result, rows.Err()
}

const drawResultSelect = `SELECT id,user_id,draw_id,prize_id,prize_name,fee_minor,reward_payload,balance_after_minor,configuration_snapshot,idempotency_key,created_at FROM activity_draw_results`

func drawResultByKeyTx(ctx context.Context, tx *sql.Tx, userID, key string) (activity.DrawResult, error) {
	return scanDrawResult(tx.QueryRowContext(ctx, drawResultSelect+` WHERE user_id=? AND idempotency_key=?`, userID, key))
}

func (s *Store) drawResultByID(ctx context.Context, resultID string) (activity.DrawResult, error) {
	return scanDrawResult(s.db.QueryRowContext(ctx, drawResultSelect+` WHERE id=?`, resultID))
}

func scanDrawResult(row rowScanner) (activity.DrawResult, error) {
	var result activity.DrawResult
	var payload, created string
	if err := row.Scan(&result.ID, &result.UserID, &result.DrawID, &result.PrizeID, &result.PrizeName, &result.FeeMinor, &payload,
		&result.BalanceAfterMinor, &result.ConfigurationSnapshot, &result.IdempotencyKey, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return activity.DrawResult{}, ErrNotFound
		}
		return activity.DrawResult{}, err
	}
	if err := json.Unmarshal([]byte(payload), &result.Reward); err != nil {
		return activity.DrawResult{}, err
	}
	parsed, err := parseStamp(created)
	result.CreatedAt = parsed
	return result, err
}

func applySubscriptionExtensionTx(ctx context.Context, tx *sql.Tx, userID string, days int, sourceType, sourceID string, now time.Time) error {
	if days < 1 || days > 3650 {
		return activity.ErrInvalidInput
	}
	var purchaseID, validUntilRaw string
	err := tx.QueryRowContext(ctx, `SELECT id,valid_until FROM purchases WHERE user_id=? AND status IN ('activating','active') AND valid_from<=? AND valid_until>?
		ORDER BY valid_from DESC LIMIT 1`, userID, stamp(now), stamp(now)).Scan(&purchaseID, &validUntilRaw)
	if errors.Is(err, sql.ErrNoRows) {
		creditID, idErr := ids.New()
		if idErr != nil {
			return idErr
		}
		_, insertErr := tx.ExecContext(ctx, `INSERT INTO activity_extension_credits(id,user_id,days,source_type,source_id,created_at) VALUES(?,?,?,?,?,?)`,
			creditID, userID, days, sourceType, sourceID, stamp(now))
		return insertErr
	}
	if err != nil {
		return err
	}
	validUntil, err := parseStamp(validUntilRaw)
	if err != nil {
		return err
	}
	shiftedUntil, err := addSubscriptionDays(validUntil, days)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE purchases SET valid_until=?,updated_at=? WHERE id=?`, stamp(shiftedUntil), stamp(now), purchaseID); err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id,valid_from,valid_until FROM purchases WHERE user_id=? AND status='queued' AND valid_from>=? ORDER BY valid_from`, userID, validUntilRaw)
	if err != nil {
		return err
	}
	type queuedTerm struct{ id, from, until string }
	queued := make([]queuedTerm, 0)
	for rows.Next() {
		var term queuedTerm
		if err := rows.Scan(&term.id, &term.from, &term.until); err != nil {
			_ = rows.Close()
			return err
		}
		queued = append(queued, term)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	_ = rows.Close()
	for _, term := range queued {
		from, parseErr := parseStamp(term.from)
		if parseErr != nil {
			return parseErr
		}
		until, parseErr := parseStamp(term.until)
		if parseErr != nil {
			return parseErr
		}
		shiftedFrom, shiftErr := addSubscriptionDays(from, days)
		if shiftErr != nil {
			return shiftErr
		}
		shiftedUntil, shiftErr := addSubscriptionDays(until, days)
		if shiftErr != nil {
			return shiftErr
		}
		if _, err := tx.ExecContext(ctx, `UPDATE purchases SET valid_from=?,valid_until=?,updated_at=? WHERE id=?`, stamp(shiftedFrom), stamp(shiftedUntil), stamp(now), term.id); err != nil {
			return err
		}
	}
	return nil
}

// consumePendingExtensionsTx marks stored extension credits used by the term
// that is actually becoming active. Queued purchases must not consume credits
// early because an older queued term remains the member's next activation.
func consumePendingExtensionsTx(ctx context.Context, tx *sql.Tx, userID, purchaseID string, now time.Time) (int, error) {
	rows, err := tx.QueryContext(ctx, `SELECT id,days FROM activity_extension_credits WHERE user_id=? AND consumed_at IS NULL ORDER BY created_at,id`, userID)
	if err != nil {
		return 0, err
	}
	idsToConsume := make([]string, 0)
	total := int64(0)
	for rows.Next() {
		var id string
		var days int
		if err := rows.Scan(&id, &days); err != nil {
			_ = rows.Close()
			return 0, err
		}
		if total > math.MaxInt64-int64(days) {
			_ = rows.Close()
			return 0, activity.ErrInvalidInput
		}
		total += int64(days)
		idsToConsume = append(idsToConsume, id)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return 0, err
	}
	_ = rows.Close()
	for _, creditID := range idsToConsume {
		if _, err := tx.ExecContext(ctx, `UPDATE activity_extension_credits SET consumed_at=?,consumed_by_purchase_id=? WHERE id=? AND consumed_at IS NULL`, stamp(now), purchaseID, creditID); err != nil {
			return 0, err
		}
	}
	if total > int64(^uint(0)>>1) {
		return 0, activity.ErrInvalidInput
	}
	return int(total), nil
}

// applyPendingExtensionsToActivationTx applies credits saved while no term was
// active to the exact next activating term, then delays every following queued
// term so subscription periods remain contiguous and non-overlapping.
func applyPendingExtensionsToActivationTx(ctx context.Context, tx *sql.Tx, purchaseID string, now time.Time) error {
	var userID, validUntilRaw string
	if err := tx.QueryRowContext(ctx, `SELECT user_id,valid_until FROM purchases WHERE id=? AND status='activating'`, purchaseID).Scan(&userID, &validUntilRaw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrConflict
		}
		return err
	}
	days, err := consumePendingExtensionsTx(ctx, tx, userID, purchaseID, now)
	if err != nil || days == 0 {
		return err
	}
	validUntil, err := parseStamp(validUntilRaw)
	if err != nil {
		return err
	}
	shiftedUntil, err := addSubscriptionDays(validUntil, days)
	if err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `UPDATE purchases SET valid_until=?,updated_at=? WHERE id=? AND status='activating'`, stamp(shiftedUntil), stamp(now), purchaseID)
	if err != nil {
		return err
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
		return rowsErr
	} else if affected != 1 {
		return ErrConflict
	}
	rows, err := tx.QueryContext(ctx, `SELECT id,valid_from,valid_until FROM purchases
		WHERE user_id=? AND id<>? AND status='queued' AND valid_from>=? ORDER BY valid_from,id`, userID, purchaseID, validUntilRaw)
	if err != nil {
		return err
	}
	type queuedTerm struct{ id, from, until string }
	queued := make([]queuedTerm, 0)
	for rows.Next() {
		var term queuedTerm
		if err := rows.Scan(&term.id, &term.from, &term.until); err != nil {
			_ = rows.Close()
			return err
		}
		queued = append(queued, term)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	_ = rows.Close()
	for _, term := range queued {
		from, parseErr := parseStamp(term.from)
		if parseErr != nil {
			return parseErr
		}
		until, parseErr := parseStamp(term.until)
		if parseErr != nil {
			return parseErr
		}
		shiftedFrom, shiftErr := addSubscriptionDays(from, days)
		if shiftErr != nil {
			return shiftErr
		}
		shiftedUntil, shiftErr := addSubscriptionDays(until, days)
		if shiftErr != nil {
			return shiftErr
		}
		result, updateErr := tx.ExecContext(ctx, `UPDATE purchases SET valid_from=?,valid_until=?,updated_at=? WHERE id=? AND status='queued'`, stamp(shiftedFrom), stamp(shiftedUntil), stamp(now), term.id)
		if updateErr != nil {
			return updateErr
		}
		if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
			return rowsErr
		} else if affected != 1 {
			return ErrConflict
		}
	}
	return nil
}

func addSubscriptionDays(value time.Time, days int) (time.Time, error) {
	shifted := value.AddDate(0, 0, days)
	if shifted.Year() < 1 || shifted.Year() > 9999 || !shifted.After(value) {
		return time.Time{}, activity.ErrInvalidInput
	}
	return shifted, nil
}

func fixedMultiplyFloor(left, right, divisor int64) (int64, error) {
	if left < 0 || right < 0 || divisor <= 0 {
		return 0, activity.ErrInvalidInput
	}
	value := new(big.Int).Mul(big.NewInt(left), big.NewInt(right))
	value.Quo(value, big.NewInt(divisor))
	if !value.IsInt64() {
		return 0, activity.ErrInvalidInput
	}
	return value.Int64(), nil
}
