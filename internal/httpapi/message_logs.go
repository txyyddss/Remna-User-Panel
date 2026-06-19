package httpapi

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type messageLogEntry struct {
	UserID            int64
	TargetUserID      int64
	TelegramUsername  string
	TelegramFirstName string
	EventType         string
	Content           string
	RawUpdatePreview  string
	IsAdminEvent      bool
	Payload           any
}

func recordMessageLog(ctx context.Context, pool *pgxpool.Pool, entry messageLogEntry) {
	if pool == nil || entry.EventType == "" {
		return
	}
	payload := entry.Payload
	if payload == nil {
		payload = map[string]any{}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		body = []byte(`{}`)
	}
	_, _ = pool.Exec(ctx, `
INSERT INTO message_logs (user_id, telegram_username, telegram_first_name, event_type, content, raw_update_preview, is_admin_event, target_user_id, payload)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		nullableInt64(entry.UserID),
		emptyStringToNil(entry.TelegramUsername),
		emptyStringToNil(entry.TelegramFirstName),
		entry.EventType,
		emptyStringToNil(entry.Content),
		emptyStringToNil(entry.RawUpdatePreview),
		entry.IsAdminEvent,
		nullableInt64(entry.TargetUserID),
		body,
	)
}

func adminMessageLogs(ctx context.Context, pool *pgxpool.Pool, page int, pageSize int, userID int64) ([]map[string]any, int64, error) {
	if page < 0 {
		page = 0
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 50
	}
	offset := page * pageSize
	where := ""
	args := []any{}
	if userID > 0 {
		where = "WHERE ml.user_id=$1 OR ml.target_user_id=$1"
		args = append(args, userID)
	}
	query := `
SELECT ml.log_id, COALESCE(ml.user_id,0), COALESCE(ml.target_user_id,0),
	COALESCE(ml.telegram_username,''), COALESCE(ml.telegram_first_name,''),
	ml.event_type, COALESCE(ml.content,''), ml.is_admin_event, ml.timestamp,
	COALESCE(author.username,''), COALESCE(author.email,''), COALESCE(author.first_name,''), COALESCE(author.last_name,''),
	COALESCE(target.username,''), COALESCE(target.email,''), COALESCE(target.first_name,''), COALESCE(target.last_name,'')
FROM message_logs ml
LEFT JOIN users author ON author.user_id = ml.user_id
LEFT JOIN users target ON target.user_id = ml.target_user_id
` + where + `
ORDER BY ml.timestamp DESC
LIMIT $` + strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)
	args = append(args, pageSize, offset)
	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	logs := []map[string]any{}
	for rows.Next() {
		var logID, authorID, targetID int64
		var telegramUsername, telegramFirstName, eventType, content string
		var adminEvent bool
		var timestamp time.Time
		var authorUsername, authorEmail, authorFirstName, authorLastName string
		var targetUsername, targetEmail, targetFirstName, targetLastName string
		if err := rows.Scan(
			&logID, &authorID, &targetID, &telegramUsername, &telegramFirstName, &eventType, &content, &adminEvent, &timestamp,
			&authorUsername, &authorEmail, &authorFirstName, &authorLastName,
			&targetUsername, &targetEmail, &targetFirstName, &targetLastName,
		); err != nil {
			return nil, 0, err
		}
		logs = append(logs, map[string]any{
			"log_id":              logID,
			"user_id":             zeroInt64ToNil(authorID),
			"user_label":          adminLogUserLabel(authorID, telegramUsername, telegramFirstName, authorUsername, authorFirstName, authorLastName, authorEmail),
			"telegram_username":   telegramUsername,
			"telegram_first_name": telegramFirstName,
			"email":               authorEmail,
			"event_type":          eventType,
			"content":             content,
			"is_admin_event":      adminEvent,
			"target_user_id":      zeroInt64ToNil(targetID),
			"target_user_label":   adminLogUserLabel(targetID, "", "", targetUsername, targetFirstName, targetLastName, targetEmail),
			"timestamp":           timestamp.Format(time.RFC3339),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	countArgs := args[:len(args)-2]
	countQuery := "SELECT COUNT(*) FROM message_logs ml " + where
	var total int64
	if err := pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func adminLogUserLabel(id int64, telegramUsername string, telegramFirstName string, username string, firstName string, lastName string, email string) string {
	name := stringsTrimJoin(firstName, lastName)
	if name == "" {
		name = stringsTrimJoin(telegramFirstName)
	}
	if name == "" {
		name = firstNonEmpty(username, telegramUsername, email)
	}
	if name == "" && id > 0 {
		name = strconv.FormatInt(id, 10)
	}
	return name
}

func stringsTrimJoin(values ...string) string {
	result := ""
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if result != "" {
			result += " "
		}
		result += value
	}
	return result
}

func nullableInt64(value int64) any {
	if value == 0 {
		return nil
	}
	return value
}

func zeroInt64ToNil(value int64) any {
	if value == 0 {
		return nil
	}
	return value
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
