// Package database opens SQLite, applies embedded migrations, and exposes the application store.
package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

const destructiveExpansionMigration = "009_expansion_and_cleanup.sql"

// Open creates the data directory, opens SQLite, and applies pending migrations.
func Open(ctx context.Context, path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	dsn := "file:" + filepath.ToSlash(path) + "?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(FULL)&_pragma=cache_size(-16384)&_pragma=mmap_size(134217728)&_pragma=temp_store(MEMORY)&_pragma=wal_autocheckpoint(1000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(30 * time.Minute)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	if err := backupBeforeExpansionMigration(ctx, db, path); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := migrate(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func backupBeforeExpansionMigration(ctx context.Context, db *sql.DB, databasePath string) error {
	var migrationTableCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'`).Scan(&migrationTableCount); err != nil {
		return fmt.Errorf("inspect migration table before expansion: %w", err)
	}
	if migrationTableCount == 0 {
		return nil
	}
	var alreadyApplied, appliedCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version=?`, destructiveExpansionMigration).Scan(&alreadyApplied); err != nil {
		return fmt.Errorf("inspect expansion migration state: %w", err)
	}
	if alreadyApplied > 0 {
		return nil
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&appliedCount); err != nil {
		return fmt.Errorf("count applied migrations: %w", err)
	}
	if appliedCount == 0 {
		return nil
	}
	backupPath := databasePath + ".pre-009-" + time.Now().UTC().Format("20060102T150405.000000000Z") + ".db"
	if _, err := db.ExecContext(ctx, `PRAGMA wal_checkpoint(PASSIVE)`); err != nil {
		return fmt.Errorf("checkpoint before expansion backup: %w", err)
	}
	if _, err := db.ExecContext(ctx, `VACUUM INTO ?`, backupPath); err != nil {
		return fmt.Errorf("create pre-expansion backup: %w", err)
	}
	backupDB, err := sql.Open("sqlite", "file:"+filepath.ToSlash(backupPath)+"?mode=ro&_pragma=query_only(1)")
	if err != nil {
		return fmt.Errorf("open pre-expansion backup: %w", err)
	}
	var integrity string
	verifyErr := backupDB.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&integrity)
	closeErr := backupDB.Close()
	if verifyErr != nil {
		return fmt.Errorf("verify pre-expansion backup: %w", verifyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close pre-expansion backup: %w", closeErr)
	}
	if integrity != "ok" {
		return fmt.Errorf("pre-expansion backup integrity result: %s", integrity)
	}
	return nil
}

// Checkpoint flushes committed WAL pages without changing the on-disk database
// as the authoritative source of truth.
func Checkpoint(ctx context.Context, db *sql.DB, truncate bool) error {
	mode := "PASSIVE"
	if truncate {
		mode = "TRUNCATE"
	}
	if _, err := db.ExecContext(ctx, `PRAGMA wal_checkpoint(`+mode+`)`); err != nil {
		return fmt.Errorf("checkpoint SQLite WAL: %w", err)
	}
	return nil
}

func migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}
	entries, err := fs.ReadDir(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		var applied int
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, entry.Name()).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %s: %w", entry.Name(), err)
		}
		if applied > 0 {
			continue
		}
		script, err := migrations.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", entry.Name(), err)
		}
		if _, err := tx.ExecContext(ctx, string(script)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`, entry.Name(), time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", entry.Name(), err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", entry.Name(), err)
		}
	}
	return nil
}

// MigrationVersions returns the ordered names of every migration embedded in
// this binary. Restore validation uses this allowlist to reject snapshots that
// were created by a newer or otherwise incompatible application build.
func MigrationVersions() ([]string, error) {
	entries, err := fs.ReadDir(migrations, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}
	versions := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		versions = append(versions, entry.Name())
	}
	sort.Strings(versions)
	return versions, nil
}
