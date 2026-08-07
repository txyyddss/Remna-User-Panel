PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS schema_migrations (
  version TEXT PRIMARY KEY,
  applied_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  telegram_id INTEGER NOT NULL UNIQUE,
  telegram_first_name TEXT NOT NULL DEFAULT '',
  telegram_last_name TEXT NOT NULL DEFAULT '',
  telegram_username TEXT NOT NULL DEFAULT '',
  username TEXT UNIQUE,
  role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
  onboarding_state TEXT NOT NULL DEFAULT 'intro' CHECK (onboarding_state IN ('intro','membership','username','agreement','complete')),
  group_joined INTEGER NOT NULL DEFAULT 0,
  channel_joined INTEGER NOT NULL DEFAULT 0,
  policy_accepted_at TEXT,
  remna_user_id TEXT UNIQUE,
  remna_subscription_url TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
  token_hash BLOB PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at TEXT NOT NULL,
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS sessions_user_id_idx ON sessions(user_id);

CREATE TABLE IF NOT EXISTS join_invites (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  chat_kind TEXT NOT NULL CHECK (chat_kind IN ('group','channel')),
  chat_id INTEGER NOT NULL,
  invite_link TEXT NOT NULL,
  expires_at TEXT NOT NULL,
  approved_at TEXT,
  revoked_at TEXT,
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS join_invites_link_idx ON join_invites(invite_link);

CREATE TABLE IF NOT EXISTS settings (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,
  encrypted INTEGER NOT NULL DEFAULT 0,
  updated_at TEXT NOT NULL,
  updated_by TEXT REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS squad_products (
  id TEXT PRIMARY KEY,
  remna_squad_uuid TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  price_txb_minor INTEGER NOT NULL DEFAULT 0 CHECK (price_txb_minor >= 0),
  visible INTEGER NOT NULL DEFAULT 0,
  upstream_present INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS combos (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  price_txb_minor INTEGER NOT NULL CHECK (price_txb_minor >= 0),
  validity_days INTEGER NOT NULL CHECK (validity_days BETWEEN 1 AND 3650),
  traffic_limit_bytes INTEGER NOT NULL CHECK (traffic_limit_bytes > 0),
  reset_strategy TEXT NOT NULL CHECK (reset_strategy IN ('DAY','WEEK','MONTH')),
  active INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS combo_squads (
  combo_id TEXT NOT NULL REFERENCES combos(id) ON DELETE CASCADE,
  squad_product_id TEXT NOT NULL REFERENCES squad_products(id),
  PRIMARY KEY (combo_id, squad_product_id)
);

CREATE TABLE IF NOT EXISTS balances (
  user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  txb_minor INTEGER NOT NULL DEFAULT 0,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS ledger_entries (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id),
  delta_txb_minor INTEGER NOT NULL,
  balance_after INTEGER NOT NULL,
  kind TEXT NOT NULL,
  reference_id TEXT NOT NULL,
  note TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  UNIQUE(kind, reference_id)
);
CREATE INDEX IF NOT EXISTS ledger_user_created_idx ON ledger_entries(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS purchases (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id),
  combo_id TEXT NOT NULL REFERENCES combos(id),
  price_txb_minor INTEGER NOT NULL CHECK (price_txb_minor >= 0),
  valid_from TEXT NOT NULL,
  valid_until TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('activating','active','queued','expired','cancelled','failed')),
  traffic_limit_bytes INTEGER NOT NULL,
  reset_strategy TEXT NOT NULL,
  catalog_snapshot TEXT NOT NULL,
  traffic_reset_started_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS purchases_user_status_idx ON purchases(user_id, status, valid_until);

CREATE TABLE IF NOT EXISTS purchase_squads (
  purchase_id TEXT NOT NULL REFERENCES purchases(id) ON DELETE CASCADE,
  remna_squad_uuid TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('included','addon')),
  price_txb_minor INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (purchase_id, remna_squad_uuid)
);

CREATE TABLE IF NOT EXISTS payment_orders (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id),
  provider TEXT NOT NULL CHECK (provider IN ('ezpay','bepusdt','stars')),
  status TEXT NOT NULL CHECK (status IN ('creating','pending','paid','expired','failed','refunded')),
  txb_minor INTEGER NOT NULL CHECK (txb_minor > 0),
  payable_amount TEXT NOT NULL,
  payable_currency TEXT NOT NULL,
  rate_snapshot TEXT NOT NULL,
  provider_trade_id TEXT UNIQUE,
  provider_charge_id TEXT UNIQUE,
  payment_url TEXT,
  qr_payload TEXT,
  provider_payload TEXT NOT NULL DEFAULT '{}',
  expires_at TEXT NOT NULL,
  paid_at TEXT,
  refunded_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS payment_orders_user_created_idx ON payment_orders(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS refunds (
  id TEXT PRIMARY KEY,
  payment_order_id TEXT NOT NULL UNIQUE REFERENCES payment_orders(id),
  actor_user_id TEXT REFERENCES users(id),
  txb_minor INTEGER NOT NULL CHECK (txb_minor > 0),
  reason TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('completed')),
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS refunds_created_idx ON refunds(created_at DESC);

CREATE TABLE IF NOT EXISTS webhook_events (
  id TEXT PRIMARY KEY,
  provider TEXT NOT NULL,
  dedupe_key TEXT NOT NULL,
  order_id TEXT NOT NULL REFERENCES payment_orders(id),
  payload_hash TEXT NOT NULL,
  received_at TEXT NOT NULL,
  processed_at TEXT,
  UNIQUE(provider, dedupe_key)
);

CREATE TABLE IF NOT EXISTS outbox_jobs (
  id TEXT PRIMARY KEY,
  kind TEXT NOT NULL,
  aggregate_id TEXT NOT NULL,
  payload TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','processing','done','failed')),
  attempts INTEGER NOT NULL DEFAULT 0,
  available_at TEXT NOT NULL,
  last_error TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS outbox_available_idx ON outbox_jobs(status, available_at);

CREATE TABLE IF NOT EXISTS audit_events (
  id TEXT PRIMARY KEY,
  actor_user_id TEXT REFERENCES users(id),
  action TEXT NOT NULL,
  target_type TEXT NOT NULL,
  target_id TEXT NOT NULL,
  detail TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS audit_created_idx ON audit_events(created_at DESC);

CREATE TABLE IF NOT EXISTS backup_runs (
  id TEXT PRIMARY KEY,
  path TEXT NOT NULL,
  size_bytes INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL CHECK (status IN ('running','complete','failed')),
  error TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  completed_at TEXT
);
