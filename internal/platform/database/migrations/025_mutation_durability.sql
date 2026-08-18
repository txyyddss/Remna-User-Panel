PRAGMA foreign_keys = ON;

CREATE TABLE emby_accounts_v2 (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  base_username TEXT NOT NULL,
  remote_user_id TEXT UNIQUE,
  remote_username TEXT,
  candidate_username TEXT UNIQUE,
  status TEXT NOT NULL
    CHECK (status IN ('queued','provisioning','active','failed','pending_review')),
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
  updated_at TEXT NOT NULL,
  pending_preferences_json TEXT NOT NULL DEFAULT '{}'
    CHECK (json_valid(pending_preferences_json))
);

INSERT INTO emby_accounts_v2
SELECT id,user_id,base_username,remote_user_id,remote_username,candidate_username,status,
  setup_price_txb_minor,setup_attempt,password_ciphertext,password_context,max_parental_rating,
  create_attempted,last_error,provisioned_at,refunded_at,created_at,updated_at,pending_preferences_json
FROM emby_accounts;

CREATE TABLE emby_account_disabled_folders_v2 (
  account_id TEXT NOT NULL REFERENCES emby_accounts_v2(id) ON DELETE CASCADE,
  folder_id TEXT NOT NULL,
  PRIMARY KEY (account_id,folder_id)
);
INSERT INTO emby_account_disabled_folders_v2
SELECT account_id,folder_id FROM emby_account_disabled_folders;

DROP TABLE emby_account_disabled_folders;
DROP TABLE emby_accounts;
ALTER TABLE emby_accounts_v2 RENAME TO emby_accounts;
ALTER TABLE emby_account_disabled_folders_v2 RENAME TO emby_account_disabled_folders;
CREATE INDEX emby_accounts_status_idx ON emby_accounts(status,updated_at);

CREATE UNIQUE INDEX emby_setup_open_operation_idx
  ON provider_operations(owner_user_id,kind)
  WHERE kind='emby_setup' AND status IN ('queued','processing','pending_review');
