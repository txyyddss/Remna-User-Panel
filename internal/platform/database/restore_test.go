package database

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestPrepareRestoreSnapshotMigratesAnOlderPrefix(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "older.db")
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path))
	if err != nil {
		t.Fatal(err)
	}
	script, err := migrations.ReadFile("migrations/001_initial.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, string(script)); err != nil {
		_ = db.Close()
		t.Fatalf("apply old migration: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO schema_migrations(version,applied_at) VALUES('001_initial.sql','2026-08-08T00:00:00Z')`); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	if err := PrepareRestoreSnapshot(ctx, path); err != nil {
		t.Fatalf("PrepareRestoreSnapshot(): %v", err)
	}
	prepared, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = prepared.Close() }()
	versions, err := MigrationVersions()
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := prepared.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != len(versions) {
		t.Fatalf("migration count = %d, want %d", count, len(versions))
	}
}

func TestPrepareRestoreSnapshotRejectsFalselyMarkedSchema(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "false-markers.db")
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `CREATE TABLE schema_migrations(version TEXT PRIMARY KEY,applied_at TEXT NOT NULL)`); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	versions, err := MigrationVersions()
	if err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	for _, version := range versions {
		if _, err := db.ExecContext(ctx, `INSERT INTO schema_migrations(version,applied_at) VALUES(?,?)`, version, "2026-08-08T00:00:00Z"); err != nil {
			_ = db.Close()
			t.Fatal(err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	err = PrepareRestoreSnapshot(ctx, path)
	if err == nil || !strings.Contains(err.Error(), "schema does not match") {
		t.Fatalf("PrepareRestoreSnapshot() error = %v, want schema mismatch", err)
	}
}
