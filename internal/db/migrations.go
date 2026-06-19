package db

import (
	"context"

	"github.com/jackc/pgx/v5"

	"remna-user-panel/internal/config"
)

// Migration is one idempotent database schema change.
type Migration struct {
	ID string
	Up func(context.Context, pgx.Tx) error
}

func coreMigrations(settings config.Settings) []Migration {
	defaultLanguage := settings.DefaultLanguage
	return []Migration{
		{
			ID: "core.0000_schema_migrations",
			Up: func(ctx context.Context, tx pgx.Tx) error {
				_, err := tx.Exec(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
	id VARCHAR(255) PRIMARY KEY,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`)
				return err
			},
		},
		{
			ID: "core.go.0001_minimum_runtime_tables",
			Up: func(ctx context.Context, tx pgx.Tx) error {
				_, err := tx.Exec(ctx, `
CREATE TABLE IF NOT EXISTS users (
	user_id BIGINT PRIMARY KEY,
	username VARCHAR NULL,
	email VARCHAR NULL UNIQUE,
	email_verified_at TIMESTAMPTZ NULL,
	password_hash VARCHAR NULL,
	password_set_at TIMESTAMPTZ NULL,
	telegram_id BIGINT NULL UNIQUE,
	telegram_photo_url TEXT NULL,
	telegram_notifications_status VARCHAR(32) NOT NULL DEFAULT 'unknown',
	telegram_notifications_checked_at TIMESTAMPTZ NULL,
	telegram_notifications_enabled_at TIMESTAMPTZ NULL,
	telegram_notifications_blocked_at TIMESTAMPTZ NULL,
	first_name VARCHAR NULL,
	last_name VARCHAR NULL,
	language_code VARCHAR NOT NULL DEFAULT '`+defaultLanguage+`',
	registration_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	is_banned BOOLEAN NOT NULL DEFAULT FALSE,
	panel_user_uuid VARCHAR NULL UNIQUE,
	referral_code VARCHAR(64) NULL UNIQUE,
	referred_by_id BIGINT NULL REFERENCES users(user_id),
	lifetime_used_traffic_bytes BIGINT NULL,
	lifetime_used_traffic_synced_at TIMESTAMPTZ NULL,
	trial_eligibility_reset_at TIMESTAMPTZ NULL,
	referral_welcome_bonus_claimed_at TIMESTAMPTZ NULL,
	channel_subscription_verified BOOLEAN NULL,
	channel_subscription_checked_at TIMESTAMPTZ NULL,
	channel_subscription_verified_for BIGINT NULL
);
CREATE INDEX IF NOT EXISTS ix_users_username ON users(username);
CREATE INDEX IF NOT EXISTS ix_users_email ON users(email);
CREATE INDEX IF NOT EXISTS ix_users_telegram_id ON users(telegram_id);
CREATE INDEX IF NOT EXISTS ix_users_panel_user_uuid ON users(panel_user_uuid);
CREATE TABLE IF NOT EXISTS webhook_events (
	event_id BIGSERIAL PRIMARY KEY,
	provider VARCHAR(64) NOT NULL,
	payload JSONB NOT NULL DEFAULT '{}'::jsonb,
	status VARCHAR(32) NOT NULL DEFAULT 'queued',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	processed_at TIMESTAMPTZ NULL
);
CREATE INDEX IF NOT EXISTS ix_webhook_events_provider_status ON webhook_events(provider, status);
CREATE TABLE IF NOT EXISTS app_settings (
	key VARCHAR(255) PRIMARY KEY,
	value JSONB NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`)
				return err
			},
		},
		{
			ID: "core.go.0002_payment_orders",
			Up: func(ctx context.Context, tx pgx.Tx) error {
				_, err := tx.Exec(ctx, `
CREATE TABLE IF NOT EXISTS payment_orders (
	payment_id BIGSERIAL PRIMARY KEY,
	order_id VARCHAR(64) NOT NULL UNIQUE,
	user_id BIGINT NOT NULL REFERENCES users(user_id),
	provider VARCHAR(64) NOT NULL,
	method VARCHAR(64) NOT NULL,
	payment_type VARCHAR(64) NOT NULL DEFAULT '',
	amount NUMERIC(12,2) NOT NULL,
	currency VARCHAR(12) NOT NULL,
	status VARCHAR(32) NOT NULL DEFAULT 'pending',
	description TEXT NULL,
	tariff_key VARCHAR(128) NULL,
	sale_mode VARCHAR(64) NULL,
	months INT NULL,
	traffic_gb NUMERIC(12,2) NULL,
	device_count INT NULL,
	provider_payment_id VARCHAR(255) NULL,
	payment_url TEXT NULL,
	qr_content TEXT NULL,
	display_amount VARCHAR(64) NULL,
	display_currency VARCHAR(32) NULL,
	payment_address TEXT NULL,
	network VARCHAR(64) NULL,
	url_scheme TEXT NULL,
	expires_at TIMESTAMPTZ NULL,
	raw_webhook JSONB NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	paid_at TIMESTAMPTZ NULL
);
CREATE INDEX IF NOT EXISTS ix_payment_orders_user_created ON payment_orders(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS ix_payment_orders_provider_status ON payment_orders(provider, status);
CREATE INDEX IF NOT EXISTS ix_payment_orders_provider_payment_id ON payment_orders(provider_payment_id)`)
				return err
			},
		},
	}
}
