PRAGMA foreign_keys = ON;

CREATE TABLE admin_temporary_bans (
  user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  actor_user_id TEXT NOT NULL REFERENCES users(id), reason TEXT NOT NULL,
  expires_at TEXT NOT NULL, ban_operation_id TEXT NOT NULL REFERENCES provider_operations(id),
  unban_operation_id TEXT REFERENCES provider_operations(id), unban_reason TEXT,
  restored_at TEXT, created_at TEXT NOT NULL, updated_at TEXT NOT NULL
);
CREATE INDEX admin_temporary_bans_due ON admin_temporary_bans(expires_at) WHERE restored_at IS NULL;
