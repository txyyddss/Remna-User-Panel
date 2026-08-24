package database

import (
	"context"
	"encoding/json"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/abuse"
)

func (s *Store) CreateIncident(ctx context.Context, userID string, bucket time.Time, qps, limit int, reasons, nodes []string, policy abuse.Policy, now time.Time) (bool, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()
	recordID, created, err := ensureAbuseRecordTx(ctx, tx, userID, bucket, qps, limit, policy, now)
	if err != nil {
		return false, err
	}
	if recordID == "" {
		return false, tx.Commit()
	}
	for _, reason := range reasons {
		if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO abuse_record_reasons(record_id,name) VALUES(?,?)`, recordID, reason); err != nil {
			return false, err
		}
	}
	for _, node := range nodes {
		if _, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO abuse_record_nodes(record_id,node_uuid) VALUES(?,?)`, recordID, node); err != nil {
			return false, err
		}
	}
	if !created {
		return false, tx.Commit()
	}
	payload, _ := json.Marshal(map[string]string{"recordId": recordID})
	if err = insertOutboxTx(ctx, tx, "abuse_punishment", string(payload), now, now); err != nil {
		return false, err
	}
	if err = queueAbuseNotificationsTx(ctx, tx, recordID, userID, now); err != nil {
		return false, err
	}
	return true, tx.Commit()
}
