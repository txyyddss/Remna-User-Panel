package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/txyyddss/Remna-User-Panel/internal/activity"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
	"strings"
	"time"
)

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
