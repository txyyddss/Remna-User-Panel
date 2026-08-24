package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/abuse"
)

func (s *Store) MemberRecords(ctx context.Context, userID, cursor string, limit int) (abuse.RecordPage, error) {
	where, args := ` WHERE record.user_id=? AND record.deleted_at IS NULL`, []any{userID}
	if cursor != "" {
		where += ` AND record.created_at<?`
		args = append(args, cursor)
	}
	return s.abuseRecordPage(ctx, where, args, limit)
}
func (s *Store) AdminAbuseRecords(ctx context.Context, cursor string, limit int) (abuse.RecordPage, error) {
	where, args := ` WHERE record.deleted_at IS NULL`, []any{}
	if cursor != "" {
		where += ` AND record.created_at<?`
		args = append(args, cursor)
	}
	return s.abuseRecordPage(ctx, where, args, limit)
}
func (s *Store) abuseRecordPage(ctx context.Context, where string, args []any, limit int) (abuse.RecordPage, error) {
	if limit < 1 || limit > 100 {
		return abuse.RecordPage{}, abuse.ErrInvalid
	}
	args = append(args, limit+1)
	query := `SELECT record.id,record.created_at,record.measured_qps,record.qps_limit,record.selected_action,record.expires_at,COALESCE(GROUP_CONCAT(reason.name,', '),'global') FROM abuse_records record LEFT JOIN abuse_record_reasons reason ON reason.record_id=record.id` + where + ` GROUP BY record.id ORDER BY record.created_at DESC,record.id DESC LIMIT ?`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return abuse.RecordPage{}, err
	}
	defer rows.Close()
	page := abuse.RecordPage{Items: []abuse.Record{}}
	for rows.Next() {
		var item abuse.Record
		var created string
		var expires sql.NullString
		if err = rows.Scan(&item.ID, &created, &item.MeasuredQPS, &item.QPSLimit, &item.Action, &expires, &item.Reason); err != nil {
			return page, err
		}
		item.OccurredAt, err = parseStamp(created)
		if err != nil {
			return page, err
		}
		if expires.Valid {
			parsed, parseErr := parseStamp(expires.String)
			if parseErr != nil {
				return page, parseErr
			}
			item.ExpiresAt = &parsed
		}
		page.Items = append(page.Items, item)
	}
	if len(page.Items) > limit {
		page.NextCursor = page.Items[limit].OccurredAt.Format(time.RFC3339Nano)
		page.Items = page.Items[:limit]
	}
	return page, rows.Err()
}
