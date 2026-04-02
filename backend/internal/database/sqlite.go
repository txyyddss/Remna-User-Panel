package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// Init opens the SQLite database and runs migrations
func Init(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create db dir: %w", err)
	}

	var err error
	db, err = sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(1) // SQLite is single-writer
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Hour)

	if err := migrate(); err != nil {
		return fmt.Errorf("migration: %w", err)
	}

	slog.Info("database: initialized successfully")
	return nil
}

// DB returns the database connection
func DB() *sql.DB {
	return db
}

// Close closes the database connection
func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

func migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			telegram_id INTEGER UNIQUE NOT NULL,
			telegram_name TEXT NOT NULL DEFAULT '',
			remnawave_uuid TEXT DEFAULT '',
			jellyfin_user_id TEXT DEFAULT '',
			credit REAL NOT NULL DEFAULT 0,
			is_admin INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id)`,

		`CREATE TABLE IF NOT EXISTS combos (
			uuid TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			squad_uuid TEXT NOT NULL,
			traffic_gb INTEGER NOT NULL DEFAULT 0,
			strategy TEXT NOT NULL DEFAULT 'MONTH',
			cycle TEXT NOT NULL DEFAULT 'monthly',
			price_rmb REAL NOT NULL DEFAULT 0,
			reset_price REAL NOT NULL DEFAULT 0,
			active INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS subscriptions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id),
			combo_uuid TEXT NOT NULL REFERENCES combos(uuid),
			remnawave_uuid TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_user ON subscriptions(user_id)`,

		`CREATE TABLE IF NOT EXISTS jellyfin_accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER UNIQUE NOT NULL REFERENCES users(id),
			jellyfin_user_id TEXT NOT NULL,
			username TEXT NOT NULL,
			parental_rating INTEGER NOT NULL DEFAULT 0,
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_jellyfin_user ON jellyfin_accounts(user_id)`,

		`CREATE TABLE IF NOT EXISTS orders (
			uuid TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id),
			order_type TEXT NOT NULL,
			amount REAL NOT NULL,
			txb_discount REAL NOT NULL DEFAULT 0,
			final_amount REAL NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			service_status TEXT NOT NULL DEFAULT 'pending',
			payment_method TEXT NOT NULL DEFAULT '',
			payment_type TEXT NOT NULL DEFAULT '',
			upstream_id TEXT DEFAULT '',
			metadata TEXT DEFAULT '{}',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_user ON orders(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_upstream ON orders(upstream_id)`,

		`CREATE TABLE IF NOT EXISTS order_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_uuid TEXT NOT NULL REFERENCES orders(uuid) ON DELETE CASCADE,
			actor_user_id INTEGER REFERENCES users(id),
			event_type TEXT NOT NULL,
			message TEXT NOT NULL DEFAULT '',
			payload TEXT NOT NULL DEFAULT '{}',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_order_events_order ON order_events(order_uuid)`,
		`CREATE INDEX IF NOT EXISTS idx_order_events_time ON order_events(created_at)`,

		`CREATE TABLE IF NOT EXISTS credit_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id),
			amount REAL NOT NULL,
			balance REAL NOT NULL,
			reason TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_credit_logs_user ON credit_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_credit_logs_time ON credit_logs(created_at)`,

		`CREATE TABLE IF NOT EXISTS group_messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			telegram_msg_id INTEGER NOT NULL,
			telegram_name TEXT NOT NULL DEFAULT '',
			text TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS api_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			token_hash TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL DEFAULT '',
			permissions TEXT NOT NULL DEFAULT '[]',
			created_by INTEGER NOT NULL DEFAULT 0,
			last_used_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS signup_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id),
			date TEXT NOT NULL,
			value REAL NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, date)
		)`,

		`CREATE TABLE IF NOT EXISTS ip_change_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id),
			old_ip TEXT DEFAULT '',
			new_ip TEXT DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_ip_change_user ON ip_change_logs(user_id)`,

		`CREATE TABLE IF NOT EXISTS bet_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id),
			bet_amount REAL NOT NULL,
			result REAL NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_bet_logs_user ON bet_logs(user_id)`,
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			excerpt := m
			if len(excerpt) > 60 {
				excerpt = excerpt[:60] + "..."
			}
			return fmt.Errorf("exec migration: %w (sql: %s)", err, excerpt)
		}
	}

	// Safe column alterations
	if !columnExists("orders", "admin_note") {
		if _, err := db.Exec(`ALTER TABLE orders ADD COLUMN admin_note TEXT NOT NULL DEFAULT ''`); err != nil {
			return fmt.Errorf("add admin_note column: %w", err)
		}
	}
	if !columnExists("orders", "paid_at") {
		if _, err := db.Exec(`ALTER TABLE orders ADD COLUMN paid_at DATETIME`); err != nil {
			return fmt.Errorf("add paid_at column: %w", err)
		}
	}
	if !columnExists("orders", "service_status") {
		if _, err := db.Exec(`ALTER TABLE orders ADD COLUMN service_status TEXT NOT NULL DEFAULT 'pending'`); err != nil {
			return fmt.Errorf("add service_status column: %w", err)
		}
	}

	return nil
}

func columnExists(tableName, columnName string) bool {
	rows, err := db.Query("PRAGMA table_info(" + tableName + ")")
	if err != nil {
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dfltValue interface{}
		var pk int

		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err == nil {
			if name == columnName {
				return true
			}
		}
	}
	if err := rows.Err(); err != nil {
		slog.Error("database: column check iteration error", "table", tableName, "error", err)
	}
	return false
}

// Backup creates a backup of the database
func Backup(destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("database: create backup dir: %w", err)
	}
	if _, err := db.Exec("VACUUM INTO ?", destPath); err != nil {
		return fmt.Errorf("database: vacuum into %s: %w", destPath, err)
	}
	return nil
}
