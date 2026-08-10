package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/activity"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

type statisticEvent struct {
	userID     string
	input      int64
	output     int64
	discount   int64
	addon      int64
	won        bool
	outcomeSet bool
	created    time.Time
}

// ComboStatistics reports server-charged purchase facts while squad membership
// remains live through the referenced combo record.
func (s *Store) ComboStatistics(ctx context.Context, comboID string, from, to time.Time, bucket string, location *time.Location) (model.AdminStatistics, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT p.user_id,p.charged_txb_minor,p.coupon_discount_txb_minor,
		COALESCE((SELECT SUM(charged_txb_minor) FROM purchase_addons WHERE purchase_id=p.id),0),p.created_at
		FROM purchases p WHERE p.combo_id=? AND p.status NOT IN ('cancelled','failed') AND p.created_at>=? AND p.created_at<? ORDER BY p.created_at`,
		comboID, stamp(from), stamp(to))
	if err != nil {
		return model.AdminStatistics{}, err
	}
	events, err := scanStatisticEvents(rows, func(rows *sql.Rows) (statisticEvent, error) {
		var event statisticEvent
		var created string
		err := rows.Scan(&event.userID, &event.input, &event.discount, &event.addon, &created)
		if err == nil {
			event.created, err = parseStamp(created)
		}
		return event, err
	})
	if err != nil {
		return model.AdminStatistics{}, err
	}
	distribution, err := s.catalogSquadDistribution(ctx, comboID, "", from, to)
	if err != nil {
		return model.AdminStatistics{}, err
	}
	return aggregateStatistics(comboID, from, to, bucket, location, events, distribution), nil
}

// SquadStatistics reports included-versus-addon selection usage for one live
// Remnawave internal squad UUID.
func (s *Store) SquadStatistics(ctx context.Context, squadUUID string, from, to time.Time, bucket string, location *time.Location) (model.AdminStatistics, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT p.user_id,p.charged_txb_minor,p.coupon_discount_txb_minor,
		COALESCE((SELECT SUM(charged_txb_minor) FROM purchase_addons WHERE purchase_id=p.id AND remna_squad_uuid=?),0),p.created_at
		FROM purchases p JOIN combos c ON c.id=p.combo_id
		WHERE p.status NOT IN ('cancelled','failed') AND p.created_at>=? AND p.created_at<? AND
		(EXISTS (SELECT 1 FROM json_each(c.included_squad_uuids) WHERE value=?) OR EXISTS (SELECT 1 FROM purchase_addons WHERE purchase_id=p.id AND remna_squad_uuid=?))
		ORDER BY p.created_at`, squadUUID, stamp(from), stamp(to), squadUUID, squadUUID)
	if err != nil {
		return model.AdminStatistics{}, err
	}
	events, err := scanStatisticEvents(rows, func(rows *sql.Rows) (statisticEvent, error) {
		var event statisticEvent
		var created string
		err := rows.Scan(&event.userID, &event.input, &event.discount, &event.addon, &created)
		if err == nil {
			event.created, err = parseStamp(created)
		}
		return event, err
	})
	if err != nil {
		return model.AdminStatistics{}, err
	}
	distribution, err := s.catalogSquadDistribution(ctx, "", squadUUID, from, to)
	if err != nil {
		return model.AdminStatistics{}, err
	}
	return aggregateStatistics(squadUUID, from, to, bucket, location, events, distribution), nil
}

func (s *Store) catalogSquadDistribution(ctx context.Context, comboID, squadUUID string, from, to time.Time) ([]model.StatisticSlice, error) {
	query := `SELECT kind,uuid,COUNT(*) FROM (
		SELECT 'included' AS kind,j.value AS uuid FROM purchases p JOIN combos c ON c.id=p.combo_id, json_each(c.included_squad_uuids) j
		WHERE p.status NOT IN ('cancelled','failed') AND p.created_at>=? AND p.created_at<?
		UNION ALL
		SELECT 'addon',a.remna_squad_uuid FROM purchases p JOIN purchase_addons a ON a.purchase_id=p.id
		WHERE p.status NOT IN ('cancelled','failed') AND p.created_at>=? AND p.created_at<?) usage WHERE 1=1`
	args := []any{stamp(from), stamp(to), stamp(from), stamp(to)}
	if comboID != "" {
		query = `SELECT kind,uuid,COUNT(*) FROM (
			SELECT 'included' AS kind,j.value AS uuid FROM purchases p JOIN combos c ON c.id=p.combo_id, json_each(c.included_squad_uuids) j
			WHERE p.combo_id=? AND p.status NOT IN ('cancelled','failed') AND p.created_at>=? AND p.created_at<?
			UNION ALL SELECT 'addon',a.remna_squad_uuid FROM purchases p JOIN purchase_addons a ON a.purchase_id=p.id
			WHERE p.combo_id=? AND p.status NOT IN ('cancelled','failed') AND p.created_at>=? AND p.created_at<?) usage WHERE 1=1`
		args = []any{comboID, stamp(from), stamp(to), comboID, stamp(from), stamp(to)}
	}
	if squadUUID != "" {
		query += ` AND uuid=?`
		args = append(args, squadUUID)
	}
	query += ` GROUP BY kind,uuid ORDER BY kind,uuid`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	result := make([]model.StatisticSlice, 0)
	for rows.Next() {
		var kind, uuid string
		var count int64
		if err := rows.Scan(&kind, &uuid, &count); err != nil {
			return nil, err
		}
		result = append(result, model.StatisticSlice{ID: kind + ":" + uuid, Label: kind, Count: count})
	}
	return result, rows.Err()
}

func (s *Store) ActivityGameStatistics(ctx context.Context, gameID string, from, to time.Time, bucket string, location *time.Location) (model.AdminStatistics, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT user_id,stake_minor,payout_minor,won,created_at FROM activity_bets WHERE game_id=? AND created_at>=? AND created_at<? ORDER BY created_at`, gameID, stamp(from), stamp(to))
	if err != nil {
		return model.AdminStatistics{}, err
	}
	events, err := scanStatisticEvents(rows, func(rows *sql.Rows) (statisticEvent, error) {
		var event statisticEvent
		var won int
		var created string
		err := rows.Scan(&event.userID, &event.input, &event.output, &won, &created)
		event.won = won == 1
		event.outcomeSet = true
		if err == nil {
			event.created, err = parseStamp(created)
		}
		return event, err
	})
	if err != nil {
		return model.AdminStatistics{}, err
	}
	stats := aggregateStatistics(gameID, from, to, bucket, location, events, nil)
	stats.Distribution = []model.StatisticSlice{{ID: "win", Label: "win", Count: stats.Wins}, {ID: "loss", Label: "loss", Count: stats.Losses}}
	return stats, nil
}

func (s *Store) LuckyDrawStatistics(ctx context.Context, drawID string, from, to time.Time, bucket string, location *time.Location) (model.AdminStatistics, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT user_id,fee_minor,reward_kind,reward_payload,created_at FROM activity_draw_results WHERE draw_id=? AND created_at>=? AND created_at<? ORDER BY created_at`, drawID, stamp(from), stamp(to))
	if err != nil {
		return model.AdminStatistics{}, err
	}
	events, err := scanStatisticEvents(rows, func(rows *sql.Rows) (statisticEvent, error) {
		var event statisticEvent
		var kind activity.RewardKind
		var payload, created string
		err := rows.Scan(&event.userID, &event.input, &kind, &payload, &created)
		if err == nil && kind == activity.RewardTXBDelta {
			var reward activity.Reward
			err = json.Unmarshal([]byte(payload), &reward)
			event.output = reward.TXBDeltaMinor
		}
		if err == nil {
			event.created, err = parseStamp(created)
		}
		return event, err
	})
	if err != nil {
		return model.AdminStatistics{}, err
	}
	distributionRows, err := s.db.QueryContext(ctx, `SELECT prize_id,prize_name,COUNT(*) FROM activity_draw_results WHERE draw_id=? AND created_at>=? AND created_at<? GROUP BY prize_id,prize_name ORDER BY COUNT(*) DESC,prize_name`, drawID, stamp(from), stamp(to))
	if err != nil {
		return model.AdminStatistics{}, err
	}
	defer func() { _ = distributionRows.Close() }()
	distribution := make([]model.StatisticSlice, 0)
	for distributionRows.Next() {
		var item model.StatisticSlice
		if err := distributionRows.Scan(&item.ID, &item.Label, &item.Count); err != nil {
			return model.AdminStatistics{}, err
		}
		distribution = append(distribution, item)
	}
	if err := distributionRows.Err(); err != nil {
		return model.AdminStatistics{}, err
	}
	return aggregateStatistics(drawID, from, to, bucket, location, events, distribution), nil
}

func scanStatisticEvents(rows *sql.Rows, scan func(*sql.Rows) (statisticEvent, error)) ([]statisticEvent, error) {
	defer func() { _ = rows.Close() }()
	result := make([]statisticEvent, 0)
	for rows.Next() {
		event, err := scan(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, event)
	}
	return result, rows.Err()
}

func aggregateStatistics(resourceID string, from, to time.Time, bucket string, location *time.Location, events []statisticEvent, distribution []model.StatisticSlice) model.AdminStatistics {
	if location == nil {
		location = time.UTC
	}
	stats := model.AdminStatistics{ResourceID: resourceID, TimeZone: location.String(), From: from.In(location).Format(time.DateOnly), To: to.Add(-time.Nanosecond).In(location).Format(time.DateOnly), Bucket: bucket, Series: []model.StatisticPoint{}, Distribution: distribution}
	users := make(map[string]struct{})
	type bucketAggregate struct {
		point model.StatisticPoint
		users map[string]struct{}
	}
	buckets := make(map[string]*bucketAggregate)
	for _, event := range events {
		stats.Count++
		users[event.userID] = struct{}{}
		stats.InputMinor += event.input
		stats.OutputMinor += event.output
		stats.DiscountMinor += event.discount
		stats.AddonMinor += event.addon
		if event.outcomeSet {
			if event.won {
				stats.Wins++
			} else {
				stats.Losses++
			}
		}
		key := statisticBucket(event.created.In(location), bucket)
		aggregate := buckets[key]
		if aggregate == nil {
			aggregate = &bucketAggregate{point: model.StatisticPoint{PeriodStart: key}, users: make(map[string]struct{})}
			buckets[key] = aggregate
		}
		aggregate.point.Count++
		aggregate.point.InputMinor += event.input
		aggregate.point.OutputMinor += event.output
		aggregate.users[event.userID] = struct{}{}
	}
	stats.UniqueUsers = int64(len(users))
	stats.NetMinor = stats.InputMinor - stats.OutputMinor
	keys := make([]string, 0, len(buckets))
	for key := range buckets {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		point := buckets[key].point
		point.UniqueUsers = int64(len(buckets[key].users))
		point.NetMinor = point.InputMinor - point.OutputMinor
		stats.Series = append(stats.Series, point)
	}
	return stats
}

func statisticBucket(value time.Time, bucket string) string {
	if bucket != "weekly" {
		return value.Format(time.DateOnly)
	}
	weekday := (int(value.Weekday()) + 6) % 7
	return time.Date(value.Year(), value.Month(), value.Day()-weekday, 0, 0, 0, 0, value.Location()).Format(time.DateOnly)
}
