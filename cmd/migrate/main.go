package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"remna-user-panel/internal/config"
	"remna-user-panel/internal/db"
	"remna-user-panel/internal/logging"
)

func main() {
	settings, err := config.Load()
	if err != nil {
		slog.Error("failed to load settings", "error", err)
		os.Exit(1)
	}
	logging.Configure(settings.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// 自动创建目标数据库（如果不存在）
	if err := ensureDatabase(ctx, settings); err != nil {
		slog.Error("failed to ensure database exists", "error", err)
		os.Exit(1)
	}

	pool, err := db.Open(ctx, settings)
	if err != nil {
		slog.Error("failed to connect database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := db.RunMigrations(ctx, pool, settings); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}
	slog.Info("migration completed")
}

// ensureDatabase 连接到 postgres 默认数据库，检查目标数据库是否存在，
// 如果不存在则创建。
func ensureDatabase(ctx context.Context, settings config.Settings) error {
	dbName := extractDatabaseName(settings.DatabaseURL)
	if dbName == "" || strings.EqualFold(dbName, "postgres") {
		return nil
	}

	// 构建到 postgres 默认数据库的连接串
	defaultURL := replaceDatabaseName(settings.DatabaseURL, "postgres")

	poolConfig, err := pgxpool.ParseConfig(defaultURL)
	if err != nil {
		return fmt.Errorf("parse postgres default database url: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("connect to postgres default database: %w", err)
	}
	defer pool.Close()

	var exists bool
	err = pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname=$1)", dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check database existence: %w", err)
	}
	if exists {
		return nil
	}

	slog.Info("creating database", "database", dbName)
	// 数据库名需要用双引号包裹以支持特殊字符
	if _, err := pool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", quoteIdentifier(dbName))); err != nil {
		return fmt.Errorf("create database %s: %w", dbName, err)
	}
	return nil
}

// extractDatabaseName 从连接串中提取数据库名。
func extractDatabaseName(databaseURL string) string {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(u.Path, "/")
}

// replaceDatabaseName 替换连接串中的数据库名。
func replaceDatabaseName(databaseURL string, newDB string) string {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return databaseURL
	}
	u.Path = "/" + newDB
	return u.String()
}

// quoteIdentifier 用双引号包裹 PostgreSQL 标识符。
func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
