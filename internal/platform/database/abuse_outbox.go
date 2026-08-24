package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/abuse"
)

type AbuseJob struct {
	RecordID, UserID, RemoteUserID, Reason string
	Action                                 abuse.Action
	DurationMinutes                        int
	AllNodes                               bool
	Nodes                                  []string
	QPS, Limit                             int
	ExpiresAt                              *time.Time
}
type AbuseDelivery struct {
	AbuseJob
	TelegramID int64
	Delivered  bool
}

func (s *Store) AbuseJob(ctx context.Context, recordID string) (AbuseJob, error) {
	var job AbuseJob
	var expires sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT record.id,record.user_id,user.remna_user_id,COALESCE(GROUP_CONCAT(reason.name,', '),'global'),record.selected_action,record.measured_qps,record.qps_limit,record.expires_at FROM abuse_records record JOIN users user ON user.id=record.user_id LEFT JOIN abuse_record_reasons reason ON reason.record_id=record.id WHERE record.id=? GROUP BY record.id`, recordID).Scan(&job.RecordID, &job.UserID, &job.RemoteUserID, &job.Reason, &job.Action, &job.QPS, &job.Limit, &expires)
	if err == sql.ErrNoRows {
		return job, ErrNotFound
	}
	if err != nil {
		return job, err
	}
	if expires.Valid {
		parsed, parseErr := parseStamp(expires.String)
		if parseErr != nil {
			return job, parseErr
		}
		job.ExpiresAt = &parsed
	}
	for _, rule := range mustPunishments(s, ctx) {
		if rule.Action == job.Action {
			job.DurationMinutes = rule.DurationMinutes
			job.AllNodes = rule.AllNodes
		}
	}
	rows, err := s.db.QueryContext(ctx, `SELECT node_uuid FROM abuse_record_nodes WHERE record_id=?`, recordID)
	if err != nil {
		return job, err
	}
	defer rows.Close()
	for rows.Next() {
		var node string
		if err = rows.Scan(&node); err != nil {
			return job, err
		}
		job.Nodes = append(job.Nodes, node)
	}
	return job, rows.Err()
}
func (s *Store) AbuseDelivery(ctx context.Context, recordID string, telegramID int64) (AbuseDelivery, error) {
	var item AbuseDelivery
	job, err := s.AbuseJob(ctx, recordID)
	if err != nil {
		return item, err
	}
	item.AbuseJob = job
	item.TelegramID = telegramID
	var delivered sql.NullString
	err = s.db.QueryRowContext(ctx, `SELECT delivered_at FROM abuse_notification_deliveries WHERE record_id=? AND recipient_telegram_id=? AND kind='incident'`, recordID, telegramID).Scan(&delivered)
	if err != nil {
		return item, err
	}
	item.Delivered = delivered.Valid
	return item, nil
}
func (s *Store) MarkAbuseDelivery(ctx context.Context, recordID string, telegramID int64, now time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE abuse_notification_deliveries SET delivered_at=? WHERE record_id=? AND recipient_telegram_id=? AND kind='incident' AND delivered_at IS NULL`, stamp(now), recordID, telegramID)
	return err
}
func (s *Store) AbuseRestoreRemoteID(ctx context.Context, userID string) (string, error) {
	var remote string
	err := s.db.QueryRowContext(ctx, `SELECT user.remna_user_id FROM abuse_temp_bans ban JOIN users user ON user.id=ban.user_id WHERE ban.user_id=? AND ban.restore_queued_at IS NOT NULL AND ban.restored_at IS NULL`, userID).Scan(&remote)
	if err == sql.ErrNoRows {
		return "", ErrNotFound
	}
	return remote, err
}
func (s *Store) CompleteAbuseRestore(ctx context.Context, userID string, now time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE abuse_temp_bans SET restored_at=? WHERE user_id=? AND restored_at IS NULL`, stamp(now), userID)
	return err
}
func (s *Store) PruneAbuseRecordsTx(ctx context.Context, tx *sql.Tx, now time.Time, counts map[string]int64) error {
	var err error
	if counts["abuse_qps_samples"], err = deleteCount(ctx, tx, `DELETE FROM abuse_qps_samples WHERE bucket_at<?`, stamp(now.Add(-24*time.Hour))); err != nil {
		return err
	}
	if counts["abuse_fingerprints"], err = deleteCount(ctx, tx, `DELETE FROM abuse_log_fingerprints WHERE observed_at<?`, stamp(now.Add(-30*24*time.Hour))); err != nil {
		return err
	}
	if counts["abuse_records"], err = deleteCount(ctx, tx, `DELETE FROM abuse_records WHERE created_at<?`, stamp(now.Add(-30*24*time.Hour))); err != nil {
		return err
	}
	return nil
}

var _ = fmt.Sprintf
