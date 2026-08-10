package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
)

func insertOutboxTx(ctx context.Context, tx *sql.Tx, kind, payload string, availableAt, now time.Time) error {
	kind = strings.TrimSpace(kind)
	var typedPayload map[string]json.RawMessage
	if kind == "" || json.Unmarshal([]byte(payload), &typedPayload) != nil || len(typedPayload) == 0 {
		return errors.New("outbox kind and typed payload are required")
	}
	canonicalPayload, err := json.Marshal(typedPayload)
	if err != nil {
		return fmt.Errorf("canonicalize outbox payload: %w", err)
	}
	id, err := ids.New()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT OR IGNORE INTO outbox_jobs(id,kind,payload,status,attempts,available_at,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?)`, id, kind, string(canonicalPayload), "pending", 0, stamp(availableAt), stamp(now), stamp(now))
	if err != nil {
		return fmt.Errorf("enqueue outbox job: %w", err)
	}
	return nil
}

// EnqueueOutbox appends a durable job.
func (s *Store) EnqueueOutbox(ctx context.Context, kind, payload string, availableAt time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := insertOutboxTx(ctx, tx, kind, payload, availableAt, time.Now().UTC()); err != nil {
		return err
	}
	return tx.Commit()
}

// EnqueueDueEntitlementTransitions turns time-bound state changes into durable work.
func (s *Store) EnqueueDueEntitlementTransitions(ctx context.Context, now time.Time) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	type expiredItem struct{ purchaseID, userID string }
	rows, err := tx.QueryContext(ctx, `SELECT id,user_id FROM purchases WHERE status IN ('active','activating') AND valid_until<=?`, stamp(now))
	if err != nil {
		return err
	}
	expired := make([]expiredItem, 0)
	for rows.Next() {
		var item expiredItem
		if err := rows.Scan(&item.purchaseID, &item.userID); err != nil {
			_ = rows.Close()
			return err
		}
		expired = append(expired, item)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	_ = rows.Close()
	for _, item := range expired {
		result, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO purchase_rollovers(purchase_id,status,traffic_limit_bytes,minimum_remaining_bps,maximum_txb_minor,net_paid_txb_minor,created_at,updated_at)
			SELECT p.id,'pending',c.traffic_limit_bytes,c.rollover_min_remaining_bps,c.rollover_max_txb_minor,p.charged_txb_minor,?,?
			FROM purchases p JOIN combos c ON c.id=p.combo_id WHERE p.id=?`, stamp(now), stamp(now), item.purchaseID)
		if err != nil {
			return err
		}
		if affected, _ := result.RowsAffected(); affected == 1 {
			if err := insertOutboxTx(ctx, tx, "rollover_finalize", `{"purchaseId":"`+item.purchaseID+`"}`, now, now); err != nil {
				return err
			}
		}
	}
	// Queued terms that were never activated have no upstream traffic to settle.
	if _, err := tx.ExecContext(ctx, `UPDATE purchases SET status='expired',updated_at=? WHERE status IN ('queued','failed') AND valid_until<=?`, stamp(now), stamp(now)); err != nil {
		return err
	}
	rows, err = tx.QueryContext(ctx, `SELECT id FROM purchases WHERE status='queued' AND valid_from<=? AND valid_until>?`, stamp(now), stamp(now))
	if err != nil {
		return err
	}
	queued := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return err
		}
		queued = append(queued, id)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	_ = rows.Close()
	for _, purchaseID := range queued {
		result, err := tx.ExecContext(ctx, `UPDATE purchases SET status='activating',updated_at=? WHERE id=? AND status='queued'
			AND NOT EXISTS (SELECT 1 FROM purchases prior WHERE prior.user_id=purchases.user_id AND prior.status IN ('active','activating'))`, stamp(now), purchaseID)
		if err != nil {
			return err
		}
		if affected, _ := result.RowsAffected(); affected != 1 {
			continue
		}
		if err := applyPendingExtensionsToActivationTx(ctx, tx, purchaseID, now); err != nil {
			return err
		}
		if err := insertOutboxTx(ctx, tx, "remna_apply_entitlement", `{"purchaseId":"`+purchaseID+`"}`, now, now); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ExpirePurchase closes a delayed activation and requests a user-level entitlement reconciliation.
func (s *Store) ExpirePurchase(ctx context.Context, purchaseID string, now time.Time) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var userID, status string
	if err := tx.QueryRowContext(ctx, `SELECT user_id,status FROM purchases WHERE id=?`, purchaseID).Scan(&userID, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if status == "expired" || status == "cancelled" {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `UPDATE purchases SET status='expired',updated_at=? WHERE id=?`, stamp(now), purchaseID); err != nil {
		return err
	}
	if err := insertOutboxTx(ctx, tx, "remna_sync_user", `{"userId":"`+userID+`"}`, now, now); err != nil {
		return err
	}
	return tx.Commit()
}

// ClaimOutboxJob leases the oldest due job. A single process scheduler invokes it serially.
func (s *Store) ClaimOutboxJob(ctx context.Context, now time.Time) (*model.OutboxJob, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	job, err := scanOutbox(tx.QueryRowContext(ctx, outboxSelect+` WHERE status='pending' AND available_at<=? ORDER BY available_at,created_at LIMIT 1`, stamp(now)))
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	result, err := tx.ExecContext(ctx, `UPDATE outbox_jobs SET status='processing',attempts=attempts+1,updated_at=? WHERE id=? AND status='pending'`, stamp(now), job.ID)
	if err != nil {
		return nil, err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return nil, nil
	}
	job.Status = "processing"
	job.Attempts++
	job.UpdatedAt = now
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &job, nil
}

// CompleteOutboxJob marks success or schedules bounded exponential retry.
func (s *Store) CompleteOutboxJob(ctx context.Context, jobID string, attempts int, jobErr error, now time.Time) error {
	if jobErr == nil {
		_, err := s.db.ExecContext(ctx, `UPDATE outbox_jobs SET status='done',last_error='',updated_at=? WHERE id=?`, stamp(now), jobID)
		return err
	}
	status := "pending"
	if attempts >= 10 {
		status = "failed"
	}
	delay := time.Minute << min(attempts-1, 8)
	_, err := s.db.ExecContext(ctx, `UPDATE outbox_jobs SET status=?,last_error=?,available_at=?,updated_at=? WHERE id=?`, status, sanitizeError(jobErr), stamp(now.Add(delay)), stamp(now), jobID)
	return err
}

// RetryOutboxJob lets an administrator explicitly retry a failed operation.
func (s *Store) RetryOutboxJob(ctx context.Context, jobID string, now time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE outbox_jobs SET status='pending',attempts=0,last_error='',available_at=?,updated_at=? WHERE id=? AND status='failed'`, stamp(now), stamp(now), jobID)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrConflict
		}
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrConflict
	}
	return nil
}

// DeleteOutboxJob removes any job that is not currently leased by a worker.
func (s *Store) DeleteOutboxJob(ctx context.Context, jobID string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM outbox_jobs WHERE id=? AND status<>'processing'`, jobID)
	if err != nil {
		return fmt.Errorf("delete outbox job: %w", err)
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
		return fmt.Errorf("inspect deleted outbox job: %w", rowsErr)
	} else if affected == 0 {
		var status string
		if loadErr := s.db.QueryRowContext(ctx, `SELECT status FROM outbox_jobs WHERE id=?`, jobID).Scan(&status); errors.Is(loadErr, sql.ErrNoRows) {
			return ErrNotFound
		} else if loadErr != nil {
			return loadErr
		}
		return ErrConflict
	}
	return nil
}

// RecoverOutbox returns leases abandoned by a crashed process to the pending queue.
func (s *Store) RecoverOutbox(ctx context.Context, staleBefore, now time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE outbox_jobs SET status='pending',available_at=?,last_error='worker interrupted',updated_at=? WHERE status='processing' AND updated_at<?`, stamp(now), stamp(now), stamp(staleBefore))
	if err != nil {
		return fmt.Errorf("recover outbox leases: %w", err)
	}
	return nil
}

// ListOutboxJobs returns recent synchronization state.
func (s *Store) ListOutboxJobs(ctx context.Context, limit int) ([]model.OutboxJob, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, outboxSelect+` ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	jobs := make([]model.OutboxJob, 0)
	for rows.Next() {
		job, err := scanOutbox(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

const outboxSelect = `SELECT id,kind,payload,status,attempts,available_at,last_error,created_at,updated_at FROM outbox_jobs`

func scanOutbox(row rowScanner) (model.OutboxJob, error) {
	var job model.OutboxJob
	var available, created, updated string
	if err := row.Scan(&job.ID, &job.Kind, &job.Payload, &job.Status, &job.Attempts, &available, &job.LastError, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.OutboxJob{}, ErrNotFound
		}
		return model.OutboxJob{}, err
	}
	var err error
	if job.AvailableAt, err = parseStamp(available); err != nil {
		return model.OutboxJob{}, err
	}
	if job.CreatedAt, err = parseStamp(created); err != nil {
		return model.OutboxJob{}, err
	}
	job.UpdatedAt, err = parseStamp(updated)
	return job, err
}

func sanitizeError(err error) string {
	value := err.Error()
	if len(value) > 500 {
		return value[:500]
	}
	return value
}

// MarkPurchaseSyncResult advances an activating purchase after upstream work.
func (s *Store) MarkPurchaseSyncResult(ctx context.Context, purchaseID string, success bool, now time.Time) error {
	status := "failed"
	predicate := "status='activating'"
	if success {
		status = "active"
		predicate = "status IN ('activating','failed')"
	}
	_, err := s.db.ExecContext(ctx, `UPDATE purchases SET status=?,updated_at=? WHERE id=? AND `+predicate, status, stamp(now), purchaseID)
	return err
}

// PurchaseTrafficResetPhase returns the durable phase of a new-term reset.
func (s *Store) PurchaseTrafficResetPhase(ctx context.Context, purchaseID string) (string, error) {
	var phase string
	err := s.db.QueryRowContext(ctx, `SELECT traffic_reset_phase FROM purchases WHERE id=?`, purchaseID).Scan(&phase)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("load purchase traffic reset phase: %w", err)
	}
	return phase, nil
}

// AdvancePurchaseTrafficReset atomically records a completed external phase.
// Repeating the same transition is accepted so a DB response loss is harmless.
func (s *Store) AdvancePurchaseTrafficReset(ctx context.Context, purchaseID, from, to string, now time.Time) error {
	valid := (from == "pending" && to == "quiesced") || (from == "quiesced" && to == "reset")
	if !valid {
		return ErrConflict
	}
	result, err := s.db.ExecContext(ctx, `UPDATE purchases SET traffic_reset_phase=?,updated_at=?
		WHERE id=? AND status IN ('activating','failed') AND traffic_reset_phase=?`, to, stamp(now), purchaseID, from)
	if err != nil {
		return fmt.Errorf("advance purchase traffic reset: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect purchase traffic reset transition: %w", err)
	}
	if affected == 1 {
		return nil
	}
	phase, loadErr := s.PurchaseTrafficResetPhase(ctx, purchaseID)
	if loadErr != nil {
		return loadErr
	}
	if phase == to {
		return nil
	}
	return ErrConflict
}

// UserForPurchase returns the local user owning a purchase.
func (s *Store) UserForPurchase(ctx context.Context, purchaseID string) (model.User, error) {
	return scanUser(s.db.QueryRowContext(ctx, userSelect+` JOIN purchases ON purchases.user_id=users.id WHERE purchases.id=?`, purchaseID))
}

// DesiredEntitlement returns the currently effective purchase, if any.
func (s *Store) DesiredEntitlement(ctx context.Context, userID string, now time.Time) (*model.Purchase, error) {
	purchase, err := scanPurchase(s.db.QueryRowContext(ctx, purchaseSelect+` WHERE purchases.user_id=? AND purchases.status IN ('activating','active') AND purchases.valid_from<=? AND purchases.valid_until>? ORDER BY purchases.valid_from DESC LIMIT 1`, userID, stamp(now), stamp(now)))
	if errors.Is(err, ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	purchase.SquadUUIDs, err = s.purchaseSquads(ctx, purchase.ID)
	return &purchase, err
}

// AppendAudit writes an immutable privileged action record.
func (s *Store) AppendAudit(ctx context.Context, actorUserID *string, action, targetType, targetID, detail string, now time.Time) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	id, err := ids.New()
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin audit retention: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := insertAuditTx(ctx, tx, id, actorUserID, action, targetType, targetID, detail, now); err != nil {
		return fmt.Errorf("append audit event: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM audit_events WHERE id IN (
		SELECT id FROM audit_events ORDER BY created_at DESC,id DESC LIMIT -1 OFFSET 200
	)`); err != nil {
		return fmt.Errorf("prune audit events: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit audit retention: %w", err)
	}
	return nil
}

func insertAuditTx(ctx context.Context, tx *sql.Tx, id string, actorUserID *string, action, targetType, targetID, detail string, now time.Time) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO audit_events(id,actor_user_id,action,target_type,target_id,detail,created_at) VALUES(?,?,?,?,?,?,?)`, id, actorUserID, action, targetType, targetID, detail, stamp(now))
	return err
}

// ListAuditEvents returns the newest administrative actions.
func (s *Store) ListAuditEvents(ctx context.Context, limit int) ([]model.AuditEvent, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,actor_user_id,action,target_type,target_id,detail,created_at FROM audit_events ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	events := make([]model.AuditEvent, 0)
	for rows.Next() {
		var event model.AuditEvent
		var actor sql.NullString
		var created string
		if err := rows.Scan(&event.ID, &actor, &event.Action, &event.TargetType, &event.TargetID, &event.Detail, &created); err != nil {
			return nil, err
		}
		event.ActorUserID = nullableString(actor)
		event.CreatedAt, err = parseStamp(created)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

// ListUsers returns accounts without leaking subscription bearer URLs.
func (s *Store) ListUsers(ctx context.Context, limit int) ([]model.User, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, userSelect+` ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := make([]model.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		user.RemnaSubscriptionURL = nil
		users = append(users, user)
	}
	return users, rows.Err()
}

// StartBackupRun records a backup attempt.
func (s *Store) StartBackupRun(ctx context.Context, path string, now time.Time) (model.BackupRun, error) {
	id, err := ids.New()
	if err != nil {
		return model.BackupRun{}, err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO backup_runs(id,path,status,created_at) VALUES(?,?,?,?)`, id, path, "running", stamp(now))
	if err != nil {
		return model.BackupRun{}, err
	}
	return model.BackupRun{ID: id, Path: path, Status: "running", CreatedAt: now}, nil
}

// CompleteBackupRun records verification outcome.
func (s *Store) CompleteBackupRun(ctx context.Context, id string, size int64, backupErr error, now time.Time) error {
	status, message := "complete", ""
	if backupErr != nil {
		status, message = "failed", sanitizeError(backupErr)
	}
	_, err := s.db.ExecContext(ctx, `UPDATE backup_runs SET size_bytes=?,status=?,error=?,completed_at=? WHERE id=?`, size, status, message, stamp(now), id)
	return err
}

// ListBackupRuns returns backup history.
func (s *Store) ListBackupRuns(ctx context.Context, limit int) ([]model.BackupRun, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,path,size_bytes,status,error,created_at,completed_at FROM backup_runs ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	runs := make([]model.BackupRun, 0)
	for rows.Next() {
		var run model.BackupRun
		var created string
		var completed sql.NullString
		if err := rows.Scan(&run.ID, &run.Path, &run.SizeBytes, &run.Status, &run.Error, &created, &completed); err != nil {
			return nil, err
		}
		run.CreatedAt, err = parseStamp(created)
		if err != nil {
			return nil, err
		}
		if completed.Valid {
			value, parseErr := parseStamp(completed.String)
			if parseErr != nil {
				return nil, parseErr
			}
			run.CompletedAt = &value
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}
