PRAGMA foreign_keys = ON;

CREATE TABLE coupon_definitions (
  id TEXT PRIMARY KEY,
  code TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('purchase_recurring','purchase_once','balance_add','balance_multiply')),
  discount_mode TEXT NOT NULL DEFAULT '' CHECK (discount_mode IN ('','fixed','percent')),
  value_minor_or_bps INTEGER NOT NULL CHECK (value_minor_or_bps > 0),
  percent_cap_minor INTEGER CHECK (percent_cap_minor IS NULL OR percent_cap_minor > 0),
  eligible_combo_ids TEXT NOT NULL DEFAULT '[]',
  eligible_squad_ids TEXT NOT NULL DEFAULT '[]',
  expires_at TEXT,
  global_use_limit INTEGER CHECK (global_use_limit IS NULL OR global_use_limit > 0),
  per_user_use_limit INTEGER CHECK (per_user_use_limit IS NULL OR per_user_use_limit > 0),
  active INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX coupon_active_expiry_idx ON coupon_definitions(active, expires_at);

CREATE TABLE coupon_grants (
  id TEXT PRIMARY KEY,
  coupon_id TEXT NOT NULL REFERENCES coupon_definitions(id),
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  source_type TEXT NOT NULL,
  source_id TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','consumed','expired')),
  use_count INTEGER NOT NULL DEFAULT 0 CHECK (use_count >= 0),
  created_at TEXT NOT NULL,
  consumed_at TEXT,
  UNIQUE(user_id, coupon_id, source_type, source_id)
);
CREATE INDEX coupon_grants_user_status_idx ON coupon_grants(user_id, status, created_at DESC);

CREATE TABLE coupon_redemptions (
  id TEXT PRIMARY KEY,
  coupon_id TEXT NOT NULL REFERENCES coupon_definitions(id),
  grant_id TEXT REFERENCES coupon_grants(id),
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  balance_delta_minor INTEGER NOT NULL DEFAULT 0,
  balance_after_minor INTEGER NOT NULL,
  idempotency_key TEXT NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(user_id, idempotency_key)
);
CREATE INDEX coupon_redemptions_coupon_user_idx ON coupon_redemptions(coupon_id, user_id, created_at);

CREATE TABLE coupon_uses (
  id TEXT PRIMARY KEY,
  coupon_id TEXT NOT NULL REFERENCES coupon_definitions(id),
  grant_id TEXT NOT NULL REFERENCES coupon_grants(id),
  user_id TEXT NOT NULL REFERENCES users(id),
  purchase_id TEXT NOT NULL,
  gross_price_minor INTEGER NOT NULL CHECK (gross_price_minor >= 0),
  discount_minor INTEGER NOT NULL CHECK (discount_minor >= 0),
  net_price_minor INTEGER NOT NULL CHECK (net_price_minor >= 0),
  created_at TEXT NOT NULL,
  UNIQUE(purchase_id),
  UNIQUE(grant_id, purchase_id)
);
CREATE INDEX coupon_uses_coupon_user_idx ON coupon_uses(coupon_id, user_id, created_at);

ALTER TABLE purchases ADD COLUMN coupon_grant_id TEXT REFERENCES coupon_grants(id);
ALTER TABLE purchases ADD COLUMN gross_price_txb_minor INTEGER CHECK (gross_price_txb_minor IS NULL OR gross_price_txb_minor >= 0);
ALTER TABLE purchases ADD COLUMN coupon_discount_txb_minor INTEGER NOT NULL DEFAULT 0 CHECK (coupon_discount_txb_minor >= 0);

CREATE TABLE activity_games (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  icon TEXT NOT NULL DEFAULT '',
  enabled INTEGER NOT NULL DEFAULT 0,
  win_chance_bps INTEGER NOT NULL CHECK (win_chance_bps BETWEEN 0 AND 10000),
  minimum_stake_minor INTEGER NOT NULL CHECK (minimum_stake_minor > 0),
  maximum_stake_minor INTEGER NOT NULL CHECK (maximum_stake_minor >= minimum_stake_minor),
  return_multiplier_bps INTEGER NOT NULL CHECK (return_multiplier_bps BETWEEN 10000 AND 1000000),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE activity_bets (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  game_id TEXT NOT NULL REFERENCES activity_games(id),
  stake_minor INTEGER NOT NULL CHECK (stake_minor > 0),
  won INTEGER NOT NULL,
  payout_minor INTEGER NOT NULL CHECK (payout_minor >= 0),
  balance_after_minor INTEGER NOT NULL CHECK (balance_after_minor >= 0),
  configuration_snapshot TEXT NOT NULL,
  idempotency_key TEXT NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(user_id, idempotency_key)
);
CREATE INDEX activity_bets_user_created_idx ON activity_bets(user_id, created_at DESC);

CREATE TABLE activity_daily_checkins (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  local_date TEXT NOT NULL,
  timezone TEXT NOT NULL,
  reward_minor INTEGER NOT NULL CHECK (reward_minor >= 0),
  balance_after_minor INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(user_id, local_date)
);
CREATE INDEX activity_checkins_user_created_idx ON activity_daily_checkins(user_id, created_at DESC);

CREATE TABLE activity_lucky_draws (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  enabled INTEGER NOT NULL DEFAULT 0,
  fee_minor INTEGER NOT NULL CHECK (fee_minor >= 0),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE activity_lucky_prizes (
  id TEXT PRIMARY KEY,
  draw_id TEXT NOT NULL REFERENCES activity_lucky_draws(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  position INTEGER NOT NULL CHECK (position >= 0),
  weight INTEGER NOT NULL CHECK (weight > 0),
  stock_remaining INTEGER CHECK (stock_remaining IS NULL OR stock_remaining >= 0),
  reward_kind TEXT NOT NULL CHECK (reward_kind IN ('none','txb_delta','coupon_grant','subscription_extension')),
  reward_payload TEXT NOT NULL,
  UNIQUE(draw_id, position)
);
CREATE INDEX activity_prizes_draw_idx ON activity_lucky_prizes(draw_id, position);

CREATE TABLE activity_draw_results (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  draw_id TEXT NOT NULL REFERENCES activity_lucky_draws(id),
  prize_id TEXT NOT NULL,
  prize_name TEXT NOT NULL,
  fee_minor INTEGER NOT NULL CHECK (fee_minor >= 0),
  reward_kind TEXT NOT NULL,
  reward_payload TEXT NOT NULL,
  balance_after_minor INTEGER NOT NULL CHECK (balance_after_minor >= 0),
  configuration_snapshot TEXT NOT NULL,
  idempotency_key TEXT NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(user_id, idempotency_key)
);
CREATE INDEX activity_draw_results_user_created_idx ON activity_draw_results(user_id, created_at DESC);

CREATE TABLE activity_extension_credits (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  days INTEGER NOT NULL CHECK (days BETWEEN 1 AND 3650),
  source_type TEXT NOT NULL,
  source_id TEXT NOT NULL,
  created_at TEXT NOT NULL,
  consumed_at TEXT,
  consumed_by_purchase_id TEXT REFERENCES purchases(id),
  UNIQUE(source_type, source_id)
);
CREATE INDEX activity_extension_user_idx ON activity_extension_credits(user_id, consumed_at, created_at);

CREATE TABLE questionnaires (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  form_url TEXT NOT NULL,
  reward_txb_minor INTEGER NOT NULL CHECK (reward_txb_minor >= 0),
  status TEXT NOT NULL CHECK (status IN ('draft','active','closed','settling','settled')),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE UNIQUE INDEX questionnaires_single_active_idx ON questionnaires((1)) WHERE status='active';
CREATE INDEX questionnaires_created_idx ON questionnaires(created_at DESC);

CREATE TABLE questionnaire_participants (
  id TEXT PRIMARY KEY,
  questionnaire_id TEXT NOT NULL REFERENCES questionnaires(id) ON DELETE CASCADE,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  validation_code TEXT NOT NULL UNIQUE,
  awarded_at TEXT,
  created_at TEXT NOT NULL,
  UNIQUE(questionnaire_id, user_id)
);
CREATE INDEX questionnaire_participants_lookup_idx ON questionnaire_participants(questionnaire_id, validation_code);

CREATE TABLE questionnaire_imports (
  id TEXT PRIMARY KEY,
  questionnaire_id TEXT NOT NULL REFERENCES questionnaires(id) ON DELETE CASCADE,
  status TEXT NOT NULL CHECK (status IN ('preview','queued','processing','settled','failed')),
  raw_csv BLOB NOT NULL,
  delimiter TEXT NOT NULL CHECK (length(delimiter) = 1),
  headers_json TEXT NOT NULL,
  sample_rows_json TEXT NOT NULL,
  data_row_count INTEGER NOT NULL CHECK (data_row_count >= 0),
  malformed_row_count INTEGER NOT NULL CHECK (malformed_row_count >= 0),
  code_column TEXT,
  analysis_json TEXT,
  report_json TEXT,
  idempotency_key TEXT NOT NULL,
  last_error TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(questionnaire_id, idempotency_key)
);
CREATE INDEX questionnaire_imports_status_idx ON questionnaire_imports(status, created_at);
