CREATE TABLE activity_group_message_windows (
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  local_date TEXT NOT NULL,
  message_count INTEGER NOT NULL DEFAULT 0 CHECK (message_count >= 0),
  rewarded_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  PRIMARY KEY (user_id, local_date)
);

CREATE INDEX activity_group_message_windows_date_idx
  ON activity_group_message_windows(local_date, updated_at);

CREATE TABLE activity_group_message_events (
  chat_id INTEGER NOT NULL,
  message_id INTEGER NOT NULL,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  local_date TEXT NOT NULL,
  counted INTEGER NOT NULL DEFAULT 0 CHECK (counted IN (0,1)),
  created_at TEXT NOT NULL,
  PRIMARY KEY (chat_id, message_id)
);

CREATE INDEX activity_group_message_events_created_idx
  ON activity_group_message_events(created_at);
