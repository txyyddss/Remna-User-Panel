CREATE TABLE IF NOT EXISTS emby_accounts (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  base_username TEXT NOT NULL,
  remote_user_id TEXT UNIQUE,
  remote_username TEXT,
  candidate_username TEXT UNIQUE,
  status TEXT NOT NULL CHECK (status IN ('queued','provisioning','active','failed')),
  setup_price_txb_minor INTEGER NOT NULL CHECK (setup_price_txb_minor >= 0),
  setup_attempt INTEGER NOT NULL DEFAULT 1 CHECK (setup_attempt >= 1),
  password_ciphertext TEXT NOT NULL DEFAULT '',
  password_context TEXT NOT NULL DEFAULT '',
  max_parental_rating INTEGER,
  create_attempted INTEGER NOT NULL DEFAULT 0 CHECK (create_attempted IN (0,1)),
  last_error TEXT NOT NULL DEFAULT '',
  provisioned_at TEXT,
  refunded_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS emby_accounts_status_idx ON emby_accounts(status, updated_at);

CREATE TABLE IF NOT EXISTS emby_account_folders (
  account_id TEXT NOT NULL REFERENCES emby_accounts(id) ON DELETE CASCADE,
  folder_id TEXT NOT NULL,
  PRIMARY KEY (account_id, folder_id)
);
