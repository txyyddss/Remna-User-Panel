PRAGMA foreign_keys = ON;

CREATE TABLE user_notification_events (
  event_key TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  source_kind TEXT NOT NULL,
  source_id TEXT NOT NULL,
  kind TEXT NOT NULL,
  payload_json TEXT NOT NULL CHECK (json_valid(payload_json)),
  gate_key TEXT NOT NULL DEFAULT '',
  queued_at TEXT,
  created_at TEXT NOT NULL,
  CHECK (gate_key <> '' OR queued_at IS NOT NULL)
);

CREATE INDEX user_notification_events_gate_idx
  ON user_notification_events(gate_key,queued_at)
  WHERE queued_at IS NULL;

CREATE INDEX user_notification_events_source_idx
  ON user_notification_events(source_kind,source_id);

CREATE TRIGGER user_notification_events_immutable
BEFORE UPDATE OF user_id,source_kind,source_id,kind,payload_json,gate_key,created_at
ON user_notification_events
BEGIN
  SELECT RAISE(ABORT,'user notification payloads are immutable');
END;
