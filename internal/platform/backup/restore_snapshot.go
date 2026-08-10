package backup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"strings"
)

func copyAndSync(source *os.File, destination string) (hash string, returnedErr error) {
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("rewind source snapshot: %w", err)
	}
	file, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", fmt.Errorf("create staged snapshot: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); returnedErr == nil && closeErr != nil {
			returnedErr = fmt.Errorf("close staged snapshot: %w", closeErr)
		}
		if returnedErr != nil {
			_ = os.Remove(destination)
		}
	}()
	digest := sha256.New()
	if _, err := io.Copy(io.MultiWriter(file, digest), source); err != nil {
		return "", fmt.Errorf("copy snapshot: %w", err)
	}
	if err := file.Sync(); err != nil {
		return "", fmt.Errorf("sync staged snapshot: %w", err)
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func verifySnapshot(ctx context.Context, path string, supportedMigrations []string) error {
	return verifySnapshotMigrations(ctx, path, supportedMigrations, false)
}

func verifyCurrentSnapshot(ctx context.Context, path string, supportedMigrations []string) error {
	return verifySnapshotMigrations(ctx, path, supportedMigrations, true)
}

func verifySnapshotMigrations(ctx context.Context, path string, supportedMigrations []string, requireCurrent bool) error {
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?mode=ro&_pragma=query_only(1)&_pragma=foreign_keys(1)")
	if err != nil {
		return fmt.Errorf("open snapshot: %w", err)
	}
	defer func() { _ = db.Close() }()
	var integrity string
	if err := db.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&integrity); err != nil {
		return fmt.Errorf("run integrity check: %w", err)
	}
	if integrity != "ok" {
		return fmt.Errorf("integrity check returned %q", integrity)
	}
	foreignRows, err := db.QueryContext(ctx, `PRAGMA foreign_key_check`)
	if err != nil {
		return fmt.Errorf("run foreign-key check: %w", err)
	}
	if foreignRows.Next() {
		_ = foreignRows.Close()
		return errors.New("foreign-key check found violations")
	}
	if err := foreignRows.Err(); err != nil {
		_ = foreignRows.Close()
		return fmt.Errorf("read foreign-key check: %w", err)
	}
	if err := foreignRows.Close(); err != nil {
		return fmt.Errorf("close foreign-key check: %w", err)
	}
	for index, version := range supportedMigrations {
		if strings.TrimSpace(version) == "" {
			return errors.New("supported migration list contains an empty version")
		}
		if index > 0 && supportedMigrations[index-1] >= version {
			return errors.New("supported migration list is not strictly ordered")
		}
	}
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return fmt.Errorf("read snapshot migrations: %w", err)
	}
	defer rows.Close()
	seen := 0
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("scan snapshot migration: %w", err)
		}
		if seen >= len(supportedMigrations) || supportedMigrations[seen] != version {
			return fmt.Errorf("snapshot migrations are not an ordered prefix at %q", version)
		}
		seen++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate snapshot migrations: %w", err)
	}
	if seen == 0 {
		return errors.New("snapshot does not contain application migrations")
	}
	if requireCurrent && seen != len(supportedMigrations) {
		return fmt.Errorf("snapshot has %d of %d required migrations", seen, len(supportedMigrations))
	}
	return nil
}
