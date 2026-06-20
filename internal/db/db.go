// Package db owns PostgreSQL connectivity and schema initialization.
package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"remna-user-panel/internal/config"
)

const dbInitAdvisoryLockID int64 = 817512404897421337

// Open creates a PostgreSQL connection pool.
func Open(ctx context.Context, settings config.Settings) (*pgxpool.Pool, error) {
	if settings.DatabaseURL == "" {
		return nil, fmt.Errorf("database url is empty")
	}
	poolConfig, err := pgxpool.ParseConfig(settings.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

// RunMigrations applies the Go migration chain.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, settings config.Settings) (commitErr error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration transaction: %w", err)
	}
	defer func() {
		if commitErr != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				slog.Warn("migration rollback failed", "error", rollbackErr)
			}
		}
	}()

	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", dbInitAdvisoryLockID); err != nil {
		return fmt.Errorf("acquire migration advisory lock: %w", err)
	}
	for _, migration := range coreMigrations(settings) {
		if migration.ID != "core.0000_schema_migrations" {
			var exists bool
			if err := tx.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE id=$1)", migration.ID).Scan(&exists); err != nil {
				return fmt.Errorf("check migration %s: %w", migration.ID, err)
			}
			if exists {
				continue
			}
		}
		if err := migration.Up(ctx, tx); err != nil {
			return fmt.Errorf("apply migration %s: %w", migration.ID, err)
		}
		if migration.ID != "core.0000_schema_migrations" {
			if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (id) VALUES ($1) ON CONFLICT (id) DO NOTHING", migration.ID); err != nil {
				return fmt.Errorf("record migration %s: %w", migration.ID, err)
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migrations: %w", err)
	}
	return nil
}
