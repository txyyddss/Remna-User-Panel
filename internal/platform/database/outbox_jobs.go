package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"time"
)

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
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if jobErr == nil {
		if _, err := tx.ExecContext(ctx, `UPDATE outbox_jobs SET status='done',last_error='',updated_at=? WHERE id=?`, stamp(now), jobID); err != nil {
			return err
		}
		return tx.Commit()
	}
	status := "pending"
	if attempts >= 10 {
		status = "failed"
	}
	delay := time.Minute << min(attempts-1, 8)
	message := sanitizeError(jobErr)
	if _, err := tx.ExecContext(ctx, `UPDATE outbox_jobs SET status=?,last_error=?,available_at=?,updated_at=? WHERE id=?`, status, message, stamp(now.Add(delay)), stamp(now), jobID); err != nil {
		return err
	}
	if status == "failed" {
		if _, err := tx.ExecContext(ctx, `UPDATE questionnaire_imports SET status='failed',last_error=?,updated_at=?
			WHERE status IN ('queued','processing') AND id=(SELECT json_extract(payload,'$.importId') FROM outbox_jobs WHERE id=? AND kind='questionnaire_settlement')`, message, stamp(now), jobID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// RetryOutboxJob lets an administrator explicitly retry a failed operation.
func (s *Store) RetryOutboxJob(ctx context.Context, jobID string, now time.Time) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var kind string
	if err := tx.QueryRowContext(ctx, `SELECT kind FROM outbox_jobs WHERE id=? AND status='failed'`, jobID).Scan(&kind); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrConflict
		}
		return err
	}
	if kind == "questionnaire_settlement" {
		result, err := tx.ExecContext(ctx, `UPDATE questionnaire_imports SET status='queued',last_error='',updated_at=?
			WHERE status='failed' AND id=(SELECT json_extract(payload,'$.importId') FROM outbox_jobs WHERE id=?)`, stamp(now), jobID)
		if err != nil {
			return err
		}
		if affected, _ := result.RowsAffected(); affected != 1 {
			return ErrConflict
		}
	}
	result, err := tx.ExecContext(ctx, `UPDATE outbox_jobs SET status='pending',attempts=0,last_error='',available_at=?,updated_at=? WHERE id=? AND status='failed'`, stamp(now), stamp(now), jobID)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrConflict
		}
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrConflict
	}
	return tx.Commit()
}

// DeleteOutboxJob removes any job that is not currently leased by a worker.
func (s *Store) DeleteOutboxJob(ctx context.Context, jobID string) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var kind, status string
	if err := tx.QueryRowContext(ctx, `SELECT kind,status FROM outbox_jobs WHERE id=?`, jobID).Scan(&kind, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if status == "processing" {
		return ErrConflict
	}
	if kind == "questionnaire_settlement" && status != "done" {
		if err := abandonQuestionnaireJobTx(ctx, tx, jobID, time.Now().UTC()); err != nil {
			return err
		}
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM outbox_jobs WHERE id=? AND status<>'processing'`, jobID)
	if err != nil {
		return fmt.Errorf("delete outbox job: %w", err)
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
		return fmt.Errorf("inspect deleted outbox job: %w", rowsErr)
	} else if affected == 0 {
		return ErrConflict
	}
	return tx.Commit()
}

func abandonQuestionnaireJobTx(ctx context.Context, tx *sql.Tx, jobID string, now time.Time) error {
	var importID, questionnaireID, status string
	err := tx.QueryRowContext(ctx, `SELECT qi.id,qi.questionnaire_id,qi.status FROM questionnaire_imports qi
		JOIN outbox_jobs job ON job.id=? AND job.kind='questionnaire_settlement' AND qi.id=json_extract(job.payload,'$.importId')`, jobID).
		Scan(&importID, &questionnaireID, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if status == "processing" {
		return ErrConflict
	}
	if status == "settled" {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `UPDATE questionnaires SET status='closed',updated_at=? WHERE id=? AND status='settling'`, stamp(now), questionnaireID); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM questionnaire_imports WHERE id=? AND status<>'processing'`, importID)
	return err
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
