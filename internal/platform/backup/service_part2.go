package backup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

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

