package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/abuse"
)

func (s *Store) NodeByDigest(ctx context.Context, digest string) (abuse.NodeCredential, error) {
	var item abuse.NodeCredential
	var reported sql.NullString
	var rotated string
	err := s.db.QueryRowContext(ctx, `SELECT node_uuid,node_name,last_report_at,rotated_at FROM abuse_node_credentials WHERE key_digest=?`, digest).Scan(&item.UUID, &item.Name, &reported, &rotated)
	if err == sql.ErrNoRows {
		return abuse.NodeCredential{}, ErrNotFound
	}
	if err != nil {
		return item, fmt.Errorf("load abuse node key: %w", err)
	}
	if reported.Valid {
		parsed, err := parseStamp(reported.String)
		if err != nil {
			return item, err
		}
		item.LastReportAt = &parsed
	}
	parsed, err := parseStamp(rotated)
	if err != nil {
		return item, err
	}
	item.RotatedAt = parsed
	return item, nil
}
func (s *Store) TouchNodeReport(ctx context.Context, nodeID string, now time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE abuse_node_credentials SET last_report_at=?,updated_at=? WHERE node_uuid=?`, stamp(now), stamp(now), nodeID)
	return err
}
func (s *Store) KnownUsers(ctx context.Context) (map[string]string, map[string]bool, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,remna_user_id FROM users WHERE remna_user_id IS NOT NULL AND TRIM(remna_user_id)<>''`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	users := map[string]string{}
	for rows.Next() {
		var id, remote string
		if err = rows.Scan(&id, &remote); err != nil {
			return nil, nil, err
		}
		users[remote] = id
	}
	if err = rows.Err(); err != nil {
		return nil, nil, err
	}
	whiteRows, err := s.db.QueryContext(ctx, `SELECT remna_user_id FROM abuse_whitelist`)
	if err != nil {
		return nil, nil, err
	}
	defer whiteRows.Close()
	whitelist := map[string]bool{}
	for whiteRows.Next() {
		var remote string
		if err = whiteRows.Scan(&remote); err != nil {
			return nil, nil, err
		}
		whitelist[remote] = true
	}
	return users, whitelist, whiteRows.Err()
}
func (s *Store) StoreSamples(ctx context.Context, nodeID string, fingerprints []string, samples []abuse.Sample, now time.Time) (abuse.ReportCounts, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return abuse.ReportCounts{}, err
	}
	defer func() { _ = tx.Rollback() }()
	counts := abuse.ReportCounts{}
	fresh := map[string]bool{}
	for _, fingerprint := range fingerprints {
		result, execErr := tx.ExecContext(ctx, `INSERT OR IGNORE INTO abuse_log_fingerprints(node_uuid,fingerprint,observed_at) VALUES(?,?,?)`, nodeID, fingerprint, stamp(now))
		if execErr != nil {
			return counts, execErr
		}
		changed, _ := result.RowsAffected()
		if changed == 0 {
			counts.Duplicate++
		} else {
			fresh[fingerprint] = true
			counts.Accepted++
		}
	}
	for _, sample := range samples {
		if !fresh[sample.Fingerprint] || sample.UserID == "" || sample.QPSLimit < 1 {
			continue
		}
		_, execErr := tx.ExecContext(ctx, `INSERT INTO abuse_qps_samples(user_id,node_uuid,bucket_at,reason_name,qps_limit,qps) VALUES(?,?,?,?,?,?) ON CONFLICT(user_id,node_uuid,bucket_at,reason_name) DO UPDATE SET qps=qps+excluded.qps,qps_limit=excluded.qps_limit`, sample.UserID, sample.NodeUUID, stamp(sample.BucketAt), sample.ReasonName, sample.QPSLimit, sample.Count)
		if execErr != nil {
			return counts, fmt.Errorf("store QPS sample: %w", execErr)
		}
	}
	if err = tx.Commit(); err != nil {
		return counts, err
	}
	return counts, nil
}
func (s *Store) ReadyBuckets(ctx context.Context, cutoff time.Time) ([]abuse.Sample, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT user_id,node_uuid,bucket_at,reason_name,qps_limit,qps FROM abuse_qps_samples WHERE bucket_at<=? ORDER BY user_id,reason_name,bucket_at,node_uuid`, stamp(cutoff))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []abuse.Sample{}
	for rows.Next() {
		var item abuse.Sample
		var at string
		if err = rows.Scan(&item.UserID, &item.NodeUUID, &at, &item.ReasonName, &item.QPSLimit, &item.Count); err != nil {
			return nil, err
		}
		item.BucketAt, err = parseStamp(at)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
func (s *Store) DetectorState(ctx context.Context, userID, reason string) (time.Time, int, error) {
	var raw string
	var streak int
	err := s.db.QueryRowContext(ctx, `SELECT last_bucket_at,streak_seconds FROM abuse_detector_state WHERE user_id=? AND reason_name=?`, userID, reason).Scan(&raw, &streak)
	if err == sql.ErrNoRows {
		return time.Time{}, 0, nil
	}
	if err != nil {
		return time.Time{}, 0, err
	}
	at, err := parseStamp(raw)
	return at, streak, err
}
func (s *Store) SaveDetectorState(ctx context.Context, userID, reason string, at time.Time, streak int) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO abuse_detector_state(user_id,reason_name,last_bucket_at,streak_seconds) VALUES(?,?,?,?) ON CONFLICT(user_id,reason_name) DO UPDATE SET last_bucket_at=excluded.last_bucket_at,streak_seconds=excluded.streak_seconds`, userID, reason, stamp(at), streak)
	return err
}
