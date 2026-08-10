// Package backup creates verified online SQLite backups with bounded retention.
package backup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
)

// Repository records backup attempts in the primary database.
type Repository interface {
	StartBackupRun(context.Context, string, time.Time) (model.BackupRun, error)
	CompleteBackupRun(context.Context, string, int64, error, time.Time) error
}

// Delete removes one backup file and its completed metadata while no restore is
// staging, ready, or restarting. The path is resolved against the configured
// backup directory and symlinks are rejected before any mutation.
func (s *Service) Delete(ctx context.Context, backupID, actorID string) error {
	s.restoreMu.Lock()
	defer s.restoreMu.Unlock()
	s.mu.Lock()
	defer s.mu.Unlock()
	select {
	case <-s.restart:
		return ErrRestoreConflict
	default:
	}
	var activeRestores int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM restore_jobs WHERE status IN ('staging','ready','applying')`).Scan(&activeRestores); err != nil {
		return fmt.Errorf("check active restores: %w", err)
	}
	if activeRestores != 0 {
		return ErrRestoreConflict
	}
	var path, status string
	if err := s.db.QueryRowContext(ctx, `SELECT path,status FROM backup_runs WHERE id=?`, backupID).Scan(&path, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrBackupNotFound
		}
		return err
	}
	if status == "running" {
		return ErrRestoreConflict
	}
	resolved, err := s.resolveBackupFile(path, status == "failed")
	if err != nil {
		return err
	}
	staged := resolved + ".deleting-" + backupID
	fileExists := true
	if err := os.Rename(resolved, staged); err != nil {
		if errors.Is(err, os.ErrNotExist) && status == "failed" {
			fileExists = false
		} else {
			return fmt.Errorf("stage backup deletion: %w", err)
		}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		if fileExists {
			_ = os.Rename(staged, resolved)
		}
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM restore_jobs WHERE backup_run_id=? AND status NOT IN ('staging','ready','applying')`, backupID); err != nil {
		if fileExists {
			_ = os.Rename(staged, resolved)
		}
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM backup_runs WHERE id=?`, backupID); err != nil {
		if fileExists {
			_ = os.Rename(staged, resolved)
		}
		return err
	}
	auditID, err := ids.New()
	if err != nil {
		if fileExists {
			_ = os.Rename(staged, resolved)
		}
		return err
	}
	now := s.now().UTC().Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `INSERT INTO audit_events(id,actor_user_id,action,target_type,target_id,detail,created_at) VALUES(?,?,?,?,?,?,?)`,
		auditID, actorID, "backup.delete", "backup", backupID, `{"path":"redacted"}`, now); err != nil {
		if fileExists {
			_ = os.Rename(staged, resolved)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		if fileExists {
			_ = os.Rename(staged, resolved)
		}
		return err
	}
	if fileExists {
		if err := os.Remove(staged); err != nil {
			return fmt.Errorf("remove deleted backup file: %w", err)
		}
	}
	return nil
}

func (s *Service) resolveBackupFile(path string, allowMissing bool) (string, error) {
	directory, err := filepath.Abs(s.directory)
	if err != nil {
		return "", err
	}
	target, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(directory, target)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", errors.New("backup path escapes the configured directory")
	}
	info, err := os.Lstat(target)
	if err != nil {
		if allowMissing && errors.Is(err, os.ErrNotExist) {
			return target, nil
		}
		return "", err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("backup path is not a regular file")
	}
	return target, nil
}

// Service owns the SQLite online-backup and retention lifecycle.
type Service struct {
	db          *sql.DB
	repository  Repository
	directory   string
	retention   time.Duration
	now         func() time.Time
	mu          sync.Mutex
	restoreMu   sync.Mutex
	restart     chan struct{}
	restartOnce sync.Once
}

// NewService creates a backup service.
func NewService(db *sql.DB, repository Repository, directory string, retention time.Duration) *Service {
	return &Service{
		db: db, repository: repository, directory: filepath.Clean(directory), retention: retention,
		now: time.Now, restart: make(chan struct{}),
	}
}

// RestartRequested is closed after a restore marker has been durably staged.
// The application should return HTTP 202 before reacting to the signal, then
// gracefully stop so the next process startup can apply the restore pre-open.
func (s *Service) RestartRequested() <-chan struct{} { return s.restart }

// RequestRestart notifies the application that it can gracefully stop. HTTP
// handlers should call this only after writing the accepted restore response.
func (s *Service) RequestRestart() {
	s.restartOnce.Do(func() { close(s.restart) })
}

// Run creates a temporary VACUUM INTO snapshot, verifies it, and atomically publishes it.
func (s *Service) Run(ctx context.Context) (model.BackupRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now().UTC()
	if err := os.MkdirAll(s.directory, 0o750); err != nil {
		return model.BackupRun{}, fmt.Errorf("create backup directory: %w", err)
	}
	name := "tx-carpool-" + now.Format("20060102T150405.000000000Z") + ".db"
	finalPath := filepath.Join(s.directory, name)
	temporaryPath := filepath.Join(s.directory, "."+name+".tmp")
	run, err := s.repository.StartBackupRun(ctx, finalPath, now)
	if err != nil {
		return model.BackupRun{}, err
	}
	backupErr := s.createAndVerify(ctx, temporaryPath, finalPath)
	var size int64
	if backupErr == nil {
		if info, statErr := os.Stat(finalPath); statErr != nil {
			backupErr = statErr
		} else {
			size = info.Size()
		}
	}
	if completeErr := s.repository.CompleteBackupRun(ctx, run.ID, size, backupErr, s.now().UTC()); completeErr != nil && backupErr == nil {
		backupErr = completeErr
	}
	if backupErr != nil {
		_ = os.Remove(temporaryPath)
		run.Status = "failed"
		run.Error = backupErr.Error()
		return run, backupErr
	}
	run.Path, run.SizeBytes, run.Status = finalPath, size, "complete"
	completed := s.now().UTC()
	run.CompletedAt = &completed
	if err := s.RemoveExpired(); err != nil {
		return run, fmt.Errorf("backup complete but retention failed: %w", err)
	}
	return run, nil
}

func (s *Service) createAndVerify(ctx context.Context, temporaryPath, finalPath string) error {
	if _, err := os.Stat(temporaryPath); err == nil {
		if removeErr := os.Remove(temporaryPath); removeErr != nil {
			return fmt.Errorf("remove stale temporary backup: %w", removeErr)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `PRAGMA wal_checkpoint(PASSIVE)`); err != nil {
		return fmt.Errorf("checkpoint before SQLite backup: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `VACUUM INTO ?`, temporaryPath); err != nil {
		return fmt.Errorf("create online SQLite backup: %w", err)
	}
	backupDB, err := sql.Open("sqlite", "file:"+filepath.ToSlash(temporaryPath)+"?mode=ro&_pragma=query_only(1)")
	if err != nil {
		return fmt.Errorf("open backup for verification: %w", err)
	}
	var integrity string
	verificationErr := backupDB.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&integrity)
	closeErr := backupDB.Close()
	if verificationErr != nil {
		return fmt.Errorf("verify SQLite backup: %w", verificationErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close verified SQLite backup: %w", closeErr)
	}
	if integrity != "ok" {
		return fmt.Errorf("SQLite backup integrity result: %s", integrity)
	}
	if err := os.Rename(temporaryPath, finalPath); err != nil {
		return fmt.Errorf("publish backup: %w", err)
	}
	return nil
}

// RemoveExpired deletes only regular TX Carpool backup files older than retention.
func (s *Service) RemoveExpired() error {
	entries, err := os.ReadDir(s.directory)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	cutoff := s.now().UTC().Add(-s.retention)
	for _, entry := range entries {
		name := entry.Name()
		published := strings.HasPrefix(name, "tx-carpool-") && strings.HasSuffix(name, ".db")
		crashTemporary := strings.HasPrefix(name, ".tx-carpool-") && strings.HasSuffix(name, ".db.tmp")
		if entry.IsDir() || (!published && !crashTemporary) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			continue
		}
		if info.ModTime().UTC().Before(cutoff) {
			if err := os.Remove(filepath.Join(s.directory, entry.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}
