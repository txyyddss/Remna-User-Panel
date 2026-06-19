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
	base_amount NUMERIC(12,2) NULL,
	base_currency VARCHAR(12) NULL,
	display_cny_amount NUMERIC(12,2) NULL,
	fx_rate NUMERIC(18,8) NULL,
	fx_source VARCHAR(64) NULL,
	fx_updated_at TIMESTAMPTZ NULL,
	plan_hash VARCHAR(128) NULL,
	plan_snapshot JSONB NULL,
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
	provisioned_at TIMESTAMPTZ NULL,
	provision_error TEXT NULL,
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
		{
			ID: "core.go.0003_payment_order_plan_snapshot",
			Up: func(ctx context.Context, tx pgx.Tx) error {
				_, err := tx.Exec(ctx, `
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS base_amount NUMERIC(12,2) NULL;
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS base_currency VARCHAR(12) NULL;
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS display_cny_amount NUMERIC(12,2) NULL;
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS fx_rate NUMERIC(18,8) NULL;
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS fx_source VARCHAR(64) NULL;
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS fx_updated_at TIMESTAMPTZ NULL;
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS plan_hash VARCHAR(128) NULL;
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS plan_snapshot JSONB NULL;
CREATE INDEX IF NOT EXISTS ix_payment_orders_plan_hash ON payment_orders(plan_hash)`)
				return err
			},
		},
		{
			ID: "core.go.0004_payment_order_provisioning",
			Up: func(ctx context.Context, tx pgx.Tx) error {
				_, err := tx.Exec(ctx, `
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS provisioned_at TIMESTAMPTZ NULL;
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS provision_error TEXT NULL;
CREATE INDEX IF NOT EXISTS ix_payment_orders_paid_unprovisioned ON payment_orders(status, provisioned_at)`)
				return err
			},
		},
		{
			ID: "core.go.0005_message_logs",
			Up: func(ctx context.Context, tx pgx.Tx) error {
				_, err := tx.Exec(ctx, `
CREATE TABLE IF NOT EXISTS message_logs (
	log_id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NULL REFERENCES users(user_id) ON DELETE SET NULL,
	telegram_username VARCHAR NULL,
	telegram_first_name VARCHAR NULL,
	event_type VARCHAR(128) NOT NULL,
	content TEXT NULL,
	raw_update_preview TEXT NULL,
	is_admin_event BOOLEAN NOT NULL DEFAULT FALSE,
	target_user_id BIGINT NULL REFERENCES users(user_id) ON DELETE SET NULL,
	payload JSONB NOT NULL DEFAULT '{}'::jsonb,
	timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS ix_message_logs_timestamp ON message_logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS ix_message_logs_user_id ON message_logs(user_id);
CREATE INDEX IF NOT EXISTS ix_message_logs_target_user_id ON message_logs(target_user_id);
CREATE INDEX IF NOT EXISTS ix_message_logs_event_type ON message_logs(event_type)`)
				return err
			},
		},
		{
			ID: "core.go.0006_promos_telemetry_stars",
			Up: func(ctx context.Context, tx pgx.Tx) error {
				_, err := tx.Exec(ctx, `
CREATE TABLE IF NOT EXISTS promo_activations (
	activation_id BIGSERIAL PRIMARY KEY,
	promo_id VARCHAR(64) NOT NULL,
	promo_code VARCHAR(64) NOT NULL,
	user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
	status VARCHAR(16) NOT NULL CHECK (status IN ('processing','applied','failed')),
	bonus_days INT NOT NULL DEFAULT 0,
	error_code VARCHAR(128) NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	applied_at TIMESTAMPTZ NULL,
	UNIQUE (promo_id, user_id)
);
CREATE INDEX IF NOT EXISTS ix_promo_activations_promo_status ON promo_activations(promo_id, status);
INSERT INTO promo_activations(promo_id,promo_code,user_id,status,bonus_days,created_at,updated_at,applied_at)
SELECT COALESCE(NULLIF(promo->>'id',''),UPPER(promo->>'code')),UPPER(COALESCE(promo->>'code','')),
 (activation->>'user_id')::BIGINT,'applied',CASE WHEN COALESCE(promo->>'bonus_days','') ~ '^[0-9]+$' THEN (promo->>'bonus_days')::INT ELSE 0 END,
 COALESCE(NULLIF(activation->>'activated_at','')::TIMESTAMPTZ,NOW()),NOW(),COALESCE(NULLIF(activation->>'activated_at','')::TIMESTAMPTZ,NOW())
FROM app_settings setting
CROSS JOIN LATERAL jsonb_array_elements(CASE WHEN jsonb_typeof(setting.value)='array' THEN setting.value ELSE '[]'::jsonb END) promo
CROSS JOIN LATERAL jsonb_array_elements(CASE WHEN jsonb_typeof(promo->'activations')='array' THEN promo->'activations' ELSE '[]'::jsonb END) activation
WHERE setting.key='ADMIN_PROMOS' AND COALESCE(activation->>'user_id','') ~ '^[0-9]+$'
 AND EXISTS(SELECT 1 FROM users WHERE user_id=(activation->>'user_id')::BIGINT)
ON CONFLICT(promo_id,user_id) DO NOTHING;

CREATE TABLE IF NOT EXISTS visitor_telemetry (
	visitor_hash VARCHAR(64) PRIMARY KEY,
	full_fingerprint_hash VARCHAR(64) NOT NULL,
	canvas_hash VARCHAR(64) NULL,
	webgl_hash VARCHAR(64) NULL,
	fonts_hash VARCHAR(64) NULL,
	audio_hash VARCHAR(64) NULL,
	network_hash VARCHAR(64) NULL,
	browser_hash VARCHAR(64) NULL,
	platform_hash VARCHAR(64) NULL,
	timezone_hash VARCHAR(64) NULL,
	screen_hash VARCHAR(64) NULL,
	hardware_hash VARCHAR(64) NULL,
	language_hash VARCHAR(64) NULL,
	user_id BIGINT NULL REFERENCES users(user_id) ON DELETE SET NULL,
	first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS ix_visitor_telemetry_fingerprint ON visitor_telemetry(full_fingerprint_hash);
CREATE INDEX IF NOT EXISTS ix_visitor_telemetry_last_seen ON visitor_telemetry(last_seen_at);
CREATE TABLE IF NOT EXISTS visitor_user_links (
	visitor_hash VARCHAR(64) NOT NULL,
	user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
	last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	PRIMARY KEY(visitor_hash,user_id)
);
CREATE INDEX IF NOT EXISTS ix_visitor_user_links_user ON visitor_user_links(user_id,last_seen_at);

CREATE TABLE IF NOT EXISTS referral_welcome_claims (
	claim_id BIGSERIAL PRIMARY KEY,
	user_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
	referrer_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
	visitor_hash VARCHAR(64) NULL,
	fingerprint_hash VARCHAR(64) NULL,
	status VARCHAR(16) NOT NULL CHECK (status IN ('processing','applied','rejected','failed')),
	risk_score INT NOT NULL DEFAULT 0,
	rule_code VARCHAR(64) NULL,
	bonus_days INT NOT NULL DEFAULT 0,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	applied_at TIMESTAMPTZ NULL,
	UNIQUE (user_id)
);
CREATE INDEX IF NOT EXISTS ix_referral_claims_visitor ON referral_welcome_claims(visitor_hash, status);
CREATE INDEX IF NOT EXISTS ix_referral_claims_fingerprint ON referral_welcome_claims(fingerprint_hash, status);

CREATE TABLE IF NOT EXISTS invite_visits (
	visit_id BIGSERIAL PRIMARY KEY,
	code VARCHAR(128) NOT NULL,
	kind VARCHAR(16) NOT NULL CHECK (kind IN ('referral','campaign')),
	visitor_hash VARCHAR(64) NOT NULL,
	fingerprint_hash VARCHAR(64) NOT NULL,
	first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	visit_count INT NOT NULL DEFAULT 1,
	registered_user_id BIGINT NULL REFERENCES users(user_id) ON DELETE SET NULL,
	converted_at TIMESTAMPTZ NULL
);
CREATE INDEX IF NOT EXISTS ix_invite_visits_code_seen ON invite_visits(kind, code, last_seen_at);
CREATE INDEX IF NOT EXISTS ix_invite_visits_visitor ON invite_visits(visitor_hash, last_seen_at);

CREATE TABLE IF NOT EXISTS installation_heartbeats (
	heartbeat_date DATE PRIMARY KEY,
	version VARCHAR(128) NOT NULL,
	provenance VARCHAR(16) NOT NULL,
	os VARCHAR(32) NOT NULL,
	locale VARCHAR(16) NOT NULL,
	user_count_range VARCHAR(32) NOT NULL,
	last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS telegram_payment_charge_id VARCHAR(255) NULL;
ALTER TABLE payment_orders ADD COLUMN IF NOT EXISTS telegram_invoice_payload VARCHAR(255) NULL;
CREATE UNIQUE INDEX IF NOT EXISTS ux_payment_orders_telegram_charge ON payment_orders(telegram_payment_charge_id) WHERE telegram_payment_charge_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS ux_payment_orders_telegram_payload ON payment_orders(telegram_invoice_payload) WHERE telegram_invoice_payload IS NOT NULL;
`)
				return err
			},
		},
	}
}
