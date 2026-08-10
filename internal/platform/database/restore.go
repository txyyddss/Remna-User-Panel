package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

const (
	maximumRestoreTables  = 500
	maximumRestoreColumns = 500
)

// PrepareRestoreSnapshot applies this binary's pending migrations to a staged
// SQLite snapshot and verifies that its resulting application schema matches a
// freshly migrated database. It must only be used on an offline staged copy.
func PrepareRestoreSnapshot(ctx context.Context, path string) (returnedErr error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("restore snapshot path is empty")
	}
	absolute, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("resolve restore snapshot: %w", err)
	}
	dsn := "file:" + filepath.ToSlash(absolute) + "?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(DELETE)&_pragma=synchronous(FULL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("open staged restore snapshot: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	closed := false
	defer func() {
		if !closed {
			closeErr := db.Close()
			if returnedErr == nil && closeErr != nil {
				returnedErr = fmt.Errorf("close staged restore snapshot: %w", closeErr)
			}
		}
	}()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping staged restore snapshot: %w", err)
	}
	if err := migrate(ctx, db); err != nil {
		return fmt.Errorf("migrate staged restore snapshot: %w", err)
	}
	if err := validateRestoreDatabase(ctx, db); err != nil {
		return err
	}

	// Close before syncing so DELETE-journal commits and connection cleanup are
	// fully reflected in the staged database file.
	if err := db.Close(); err != nil {
		return fmt.Errorf("close staged restore snapshot: %w", err)
	}
	closed = true
	file, err := os.OpenFile(absolute, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open staged restore snapshot for sync: %w", err)
	}
	syncErr := file.Sync()
	closeErr := file.Close()
	if syncErr != nil {
		return fmt.Errorf("sync staged restore snapshot: %w", syncErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close synced restore snapshot: %w", closeErr)
	}
	return nil
}

func validateRestoreDatabase(ctx context.Context, candidate *sql.DB) error {
	var integrity string
	if err := candidate.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&integrity); err != nil {
		return fmt.Errorf("verify migrated restore integrity: %w", err)
	}
	if integrity != "ok" {
		return fmt.Errorf("migrated restore integrity check returned %q", integrity)
	}
	foreignRows, err := candidate.QueryContext(ctx, `PRAGMA foreign_key_check`)
	if err != nil {
		return fmt.Errorf("verify migrated restore foreign keys: %w", err)
	}
	if foreignRows.Next() {
		_ = foreignRows.Close()
		return errors.New("migrated restore contains foreign-key violations")
	}
	if err := foreignRows.Err(); err != nil {
		_ = foreignRows.Close()
		return fmt.Errorf("read migrated restore foreign keys: %w", err)
	}
	if err := foreignRows.Close(); err != nil {
		return fmt.Errorf("close migrated restore foreign-key check: %w", err)
	}

	expected, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return fmt.Errorf("open expected restore schema: %w", err)
	}
	expected.SetMaxOpenConns(1)
	defer func() { _ = expected.Close() }()
	if _, err := expected.ExecContext(ctx, `PRAGMA foreign_keys=ON`); err != nil {
		return fmt.Errorf("configure expected restore schema: %w", err)
	}
	if err := migrate(ctx, expected); err != nil {
		return fmt.Errorf("build expected restore schema: %w", err)
	}
	want, err := readRestoreSchemaShape(ctx, expected)
	if err != nil {
		return fmt.Errorf("read expected restore schema: %w", err)
	}
	got, err := readRestoreSchemaShape(ctx, candidate)
	if err != nil {
		return fmt.Errorf("read staged restore schema: %w", err)
	}
	if !reflect.DeepEqual(got, want) {
		return errors.New("migrated restore schema does not match this application build")
	}
	return nil
}
