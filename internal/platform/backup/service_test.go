package backup

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	_ "modernc.org/sqlite"
)

func TestRunCreatesVerifiedSQLiteBackup(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	databasePath := filepath.Join(t.TempDir(), "source.db")
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(databasePath))
	if err != nil {
		t.Fatalf("sql.Open(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.ExecContext(ctx, `CREATE TABLE sample(value TEXT NOT NULL); INSERT INTO sample(value) VALUES ('durable')`); err != nil {
		t.Fatalf("seed database: %v", err)
	}

	repository := &backupRepository{}
	service := NewService(db, repository, filepath.Join(t.TempDir(), "backups"), 7*24*time.Hour)
	fixedNow := time.Now().UTC()
	service.now = func() time.Time { return fixedNow }
	run, err := service.Run(ctx)
	if err != nil {
		t.Fatalf("Run(): %v", err)
	}
	if run.Status != "complete" || run.SizeBytes <= 0 || repository.completedErr != nil {
		t.Fatalf("backup run = %+v, completion error = %v", run, repository.completedErr)
	}

	backupDB, err := sql.Open("sqlite", "file:"+filepath.ToSlash(run.Path)+"?mode=ro")
	if err != nil {
		t.Fatalf("open published backup: %v", err)
	}
	defer backupDB.Close()
	var value string
	if err := backupDB.QueryRowContext(ctx, `SELECT value FROM sample`).Scan(&value); err != nil || value != "durable" {
		t.Fatalf("backup value = %q, error = %v", value, err)
	}
}

func TestRemoveExpiredCleansPublishedAndCrashTemporaryFiles(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	now := time.Now().UTC()
	old := now.Add(-8 * 24 * time.Hour)
	files := map[string]bool{
		"tx-carpool-old.db":      true,
		".tx-carpool-old.db.tmp": true,
		"tx-carpool-new.db":      false,
		".tx-carpool-new.db.tmp": false,
		"unrelated-old.db":       false,
	}
	for name, remove := range files {
		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, []byte("test"), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		if remove || name == "unrelated-old.db" {
			if err := os.Chtimes(path, old, old); err != nil {
				t.Fatalf("age %s: %v", name, err)
			}
		}
	}
	service := NewService(nil, nil, directory, 7*24*time.Hour)
	service.now = func() time.Time { return now }
	if err := service.RemoveExpired(); err != nil {
		t.Fatalf("RemoveExpired(): %v", err)
	}
	for name, removed := range files {
		_, err := os.Stat(filepath.Join(directory, name))
		if removed && !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("%s still exists", name)
		}
		if !removed && err != nil {
			t.Fatalf("%s was removed: %v", name, err)
		}
	}
}

type backupRepository struct {
	run          model.BackupRun
	completedErr error
}

func (r *backupRepository) StartBackupRun(_ context.Context, path string, now time.Time) (model.BackupRun, error) {
	r.run = model.BackupRun{ID: "backup-1", Path: path, Status: "running", CreatedAt: now}
	return r.run, nil
}

func (r *backupRepository) CompleteBackupRun(_ context.Context, _ string, size int64, backupErr error, now time.Time) error {
	r.completedErr = backupErr
	r.run.SizeBytes = size
	r.run.CompletedAt = &now
	return nil
}
