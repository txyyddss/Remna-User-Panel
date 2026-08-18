CREATE TABLE activity_group_message_raw_events (
  chat_id INTEGER NOT NULL,
  message_id INTEGER NOT NULL,
  local_date TEXT NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY (chat_id,message_id)
);

CREATE INDEX activity_group_message_raw_events_created_idx
  ON activity_group_message_raw_events(created_at);

INSERT OR IGNORE INTO activity_group_message_raw_events(chat_id,message_id,local_date,created_at)
SELECT chat_id,message_id,local_date,created_at
FROM activity_group_message_events;
