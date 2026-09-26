PRAGMA foreign_keys = ON;

-- Keep settled financial evidence while retiring every legacy draw definition.
CREATE TABLE activity_draw_results_v2 (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  draw_id TEXT NOT NULL,
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
INSERT INTO activity_draw_results_v2 SELECT * FROM activity_draw_results;
DROP TABLE activity_draw_results;
ALTER TABLE activity_draw_results_v2 RENAME TO activity_draw_results;
CREATE INDEX activity_draw_results_user_created_idx ON activity_draw_results(user_id, created_at DESC);
CREATE INDEX activity_draw_results_draw_created_idx ON activity_draw_results(draw_id, created_at);

DROP TABLE activity_lucky_prizes;
DROP TABLE activity_lucky_draws;
CREATE TABLE activity_lucky_draws (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  kind TEXT NOT NULL CHECK (kind IN ('instant','raffle')),
  status TEXT NOT NULL CHECK (status IN ('draft','publishing','open','settling','completed','cancelled')),
  fee_minor INTEGER NOT NULL CHECK (fee_minor > 0),
  expected_participation INTEGER CHECK (expected_participation IS NULL OR expected_participation > 0),
  threshold INTEGER CHECK (threshold IS NULL OR threshold > 0),
  keyword TEXT NOT NULL DEFAULT '',
  command TEXT NOT NULL DEFAULT '',
  group_chat_id INTEGER,
  announcement_message_id INTEGER,
  revision INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  CHECK ((kind='instant' AND expected_participation IS NOT NULL AND threshold IS NULL)
      OR (kind='raffle' AND threshold IS NOT NULL AND expected_participation IS NULL))
);
CREATE UNIQUE INDEX activity_raffle_open_keyword_idx ON activity_lucky_draws(lower(keyword))
  WHERE kind='raffle' AND status IN ('publishing','open') AND keyword<>'';
CREATE UNIQUE INDEX activity_raffle_open_command_idx ON activity_lucky_draws(lower(command))
  WHERE kind='raffle' AND status IN ('publishing','open') AND command<>'';

CREATE TABLE activity_lucky_prizes (
  id TEXT PRIMARY KEY,
  draw_id TEXT NOT NULL REFERENCES activity_lucky_draws(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  position INTEGER NOT NULL CHECK (position >= 0),
  probability_bps INTEGER CHECK (probability_bps IS NULL OR probability_bps BETWEEN 1 AND 10000),
  stock INTEGER CHECK (stock IS NULL OR stock > 0),
  reward_payload TEXT NOT NULL CHECK (json_valid(reward_payload)),
  UNIQUE(draw_id, position),
  CHECK ((probability_bps IS NOT NULL AND stock IS NULL) OR (probability_bps IS NULL AND stock IS NOT NULL))
);
CREATE INDEX activity_prizes_draw_idx ON activity_lucky_prizes(draw_id, position);

CREATE TABLE activity_raffle_tickets (
  id TEXT PRIMARY KEY,
  draw_id TEXT NOT NULL REFERENCES activity_lucky_draws(id),
  user_id TEXT NOT NULL REFERENCES users(id),
  chat_id INTEGER NOT NULL,
  message_id INTEGER NOT NULL,
  fee_minor INTEGER NOT NULL CHECK (fee_minor > 0),
  reserve_minor INTEGER NOT NULL DEFAULT 0 CHECK (reserve_minor >= 0),
  configuration_revision INTEGER NOT NULL CHECK (configuration_revision > 0),
  status TEXT NOT NULL CHECK (status IN ('active','refunded','settled')),
  prize_id TEXT,
  result_id TEXT REFERENCES activity_draw_results(id),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(chat_id, message_id)
);
CREATE INDEX activity_raffle_tickets_draw_status_idx ON activity_raffle_tickets(draw_id,status,created_at,id);
CREATE TABLE activity_draw_revisions (
  id TEXT PRIMARY KEY,
  draw_id TEXT NOT NULL REFERENCES activity_lucky_draws(id) ON DELETE CASCADE,
  revision INTEGER NOT NULL,
  snapshot_json TEXT NOT NULL CHECK (json_valid(snapshot_json)),
  created_at TEXT NOT NULL,
  UNIQUE(draw_id,revision)
);

ALTER TABLE coupon_definitions ADD COLUMN admin_visible INTEGER NOT NULL DEFAULT 1 CHECK (admin_visible IN (0,1));
ALTER TABLE purchases ADD COLUMN reward_renewal_price_minor INTEGER CHECK (reward_renewal_price_minor IS NULL OR reward_renewal_price_minor >= 0);
ALTER TABLE purchases ADD COLUMN reward_rollover_min_remaining_bps INTEGER CHECK (reward_rollover_min_remaining_bps IS NULL OR reward_rollover_min_remaining_bps BETWEEN 0 AND 10000);
ALTER TABLE purchases ADD COLUMN reward_traffic_renewal INTEGER NOT NULL DEFAULT 1 CHECK (reward_traffic_renewal IN (0,1));
ALTER TABLE purchases ADD COLUMN reward_renewal_traffic_limit_bytes INTEGER
  CHECK (reward_renewal_traffic_limit_bytes IS NULL OR reward_renewal_traffic_limit_bytes > 0);
ALTER TABLE activity_extension_credits ADD COLUMN hours INTEGER CHECK (hours IS NULL OR hours > 0);
CREATE TABLE activity_reward_squad_access (
  purchase_id TEXT NOT NULL REFERENCES purchases(id) ON DELETE CASCADE,
  squad_uuid TEXT NOT NULL,
  source_result_id TEXT NOT NULL REFERENCES activity_draw_results(id),
  PRIMARY KEY(purchase_id,squad_uuid,source_result_id)
);
CREATE TABLE activity_draw_delivery (
  delivery_key TEXT PRIMARY KEY,
  chat_id INTEGER NOT NULL,
  message_id INTEGER NOT NULL,
  created_at TEXT NOT NULL
);
