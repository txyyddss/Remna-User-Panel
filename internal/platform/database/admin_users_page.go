package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

// AdminUserRecord is the list-safe user and aggregate balance projection.
type AdminUserRecord struct {
	User    model.User
	Balance model.Money
}

// ListAdminUsersPage returns a stable cursor page with its balances in one query.
func (s *Store) ListAdminUsersPage(ctx context.Context, cursor, search string, limit int) ([]AdminUserRecord, *string, error) {
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	search = strings.TrimSpace(search)
	if len(search) > 100 {
		return nil, nil, ErrInvalidCursor
	}
	query := `SELECT ` + userColumns + `,COALESCE(balances.txb_minor,0) FROM users
		LEFT JOIN balances ON balances.user_id=users.id WHERE 1=1`
	args := make([]any, 0, 8)
	if search != "" {
		pattern := "%" + escapeLike(search) + "%"
		query += ` AND (users.telegram_first_name LIKE ? ESCAPE '\' COLLATE NOCASE
			OR users.telegram_last_name LIKE ? ESCAPE '\' COLLATE NOCASE
			OR users.telegram_username LIKE ? ESCAPE '\' COLLATE NOCASE
			OR COALESCE(users.username,'') LIKE ? ESCAPE '\' COLLATE NOCASE
			OR CAST(users.telegram_id AS TEXT) LIKE ? ESCAPE '\'
			OR users.id LIKE ? ESCAPE '\' COLLATE NOCASE
			OR COALESCE(users.remna_user_id,'') LIKE ? ESCAPE '\' COLLATE NOCASE)`
		args = append(args, pattern, pattern, pattern, pattern, pattern, pattern, pattern)
	}
	if cursor != "" {
		decoded, err := decodeTimestampCursor(cursor, pageFilterFingerprint(search))
		if err != nil {
			return nil, nil, err
		}
		query += ` AND (users.created_at<? OR (users.created_at=? AND users.id<?))`
		args = append(args, decoded.Timestamp, decoded.Timestamp, decoded.ID)
	}
	query += ` ORDER BY users.created_at DESC,users.id DESC LIMIT ?`
	args = append(args, limit+1)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("list admin user page: %w", err)
	}
	defer func() { _ = rows.Close() }()
	items := make([]AdminUserRecord, 0, limit+1)
	for rows.Next() {
		var balanceMinor int64
		user, scanErr := scanUserWith(rows, &balanceMinor)
		if scanErr != nil {
			return nil, nil, scanErr
		}
		items = append(items, AdminUserRecord{User: user, Balance: model.TXBMoney(balanceMinor)})
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	if len(items) <= limit {
		return items, nil, nil
	}
	items = items[:limit]
	last := items[len(items)-1].User
	next, err := encodeTimestampCursor(last.CreatedAt, last.ID, pageFilterFingerprint(search))
	return items, &next, err
}

func escapeLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `%`, `\%`)
	return strings.ReplaceAll(value, `_`, `\_`)
}
