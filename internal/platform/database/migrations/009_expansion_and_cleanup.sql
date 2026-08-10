PRAGMA foreign_keys = ON;

-- Onboarding content is versioned independently so welcome-copy changes do not
-- force agreement re-consent. Durations are derived at read time.
ALTER TABLE users ADD COLUMN accepted_agreement_revision INTEGER NOT NULL DEFAULT 0;
UPDATE users SET accepted_agreement_revision=1 WHERE policy_accepted_at IS NOT NULL;

CREATE TABLE onboarding_content (
  kind TEXT PRIMARY KEY CHECK (kind IN ('welcome','agreements')),
  draft_json TEXT NOT NULL CHECK (json_valid(draft_json)),
  published_json TEXT NOT NULL CHECK (json_valid(published_json)),
  draft_revision INTEGER NOT NULL DEFAULT 1 CHECK (draft_revision >= 1),
  published_revision INTEGER NOT NULL DEFAULT 1 CHECK (published_revision >= 1),
  updated_at TEXT NOT NULL,
  published_at TEXT NOT NULL,
  updated_by TEXT REFERENCES users(id)
);

INSERT INTO onboarding_content(kind,draft_json,published_json,draft_revision,published_revision,updated_at,published_at)
VALUES
  ('welcome',
   '{"en":[{"id":"hello","text":"Hi, how are you?"},{"id":"welcome","text":"Welcome to TX Carpool"},{"id":"setup","text":"It only takes a few seconds to complete setup."}],"zh-CN":[{"id":"hello","text":"\u4f60\u597d\uff0c\u6700\u8fd1\u600e\u4e48\u6837\uff1f"},{"id":"welcome","text":"\u6b22\u8fce\u6765\u5230 TX Carpool"},{"id":"setup","text":"\u53ea\u9700\u51e0\u79d2\u5373\u53ef\u5b8c\u6210\u8bbe\u7f6e\u3002"}]}',
   '{"en":[{"id":"hello","text":"Hi, how are you?"},{"id":"welcome","text":"Welcome to TX Carpool"},{"id":"setup","text":"It only takes a few seconds to complete setup."}],"zh-CN":[{"id":"hello","text":"\u4f60\u597d\uff0c\u6700\u8fd1\u600e\u4e48\u6837\uff1f"},{"id":"welcome","text":"\u6b22\u8fce\u6765\u5230 TX Carpool"},{"id":"setup","text":"\u53ea\u9700\u51e0\u79d2\u5373\u53ef\u5b8c\u6210\u8bbe\u7f6e\u3002"}]}',
   1,1,strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  ('agreements',
   '{"en":[{"id":"private-link","icon":"link-break","title":"Keep your link private","body":"Never post, forward, sell, or share your subscription URL with another person."}],"zh-CN":[{"id":"private-link","icon":"link-break","title":"\u8bf7\u59a5\u5584\u4fdd\u7ba1\u8ba2\u9605\u94fe\u63a5","body":"\u8bf7\u52ff\u53d1\u5e03\u3001\u8f6c\u53d1\u3001\u51fa\u552e\u8ba2\u9605\u94fe\u63a5\uff0c\u6216\u4e0e\u4ed6\u4eba\u5171\u4eab\u3002"}]}',
   '{"en":[{"id":"private-link","icon":"link-break","title":"Keep your link private","body":"Never post, forward, sell, or share your subscription URL with another person."}],"zh-CN":[{"id":"private-link","icon":"link-break","title":"\u8bf7\u59a5\u5584\u4fdd\u7ba1\u8ba2\u9605\u94fe\u63a5","body":"\u8bf7\u52ff\u53d1\u5e03\u3001\u8f6c\u53d1\u3001\u51fa\u552e\u8ba2\u9605\u94fe\u63a5\uff0c\u6216\u4e0e\u4ed6\u4eba\u5171\u4eab\u3002"}]}',
   1,1,strftime('%Y-%m-%dT%H:%M:%fZ','now'),strftime('%Y-%m-%dT%H:%M:%fZ','now'));

-- Keep only administrator-authored merchandising overrides. Remnawave owns
-- squad identity, name, availability, nodes, flags, and multipliers.
CREATE TABLE squad_product_overrides (
  remna_squad_uuid TEXT PRIMARY KEY,
  description TEXT NOT NULL DEFAULT '',
  price_txb_minor INTEGER NOT NULL DEFAULT 0 CHECK (price_txb_minor >= 0),
  visible INTEGER NOT NULL DEFAULT 0 CHECK (visible IN (0,1)),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

INSERT INTO squad_product_overrides(remna_squad_uuid,description,price_txb_minor,visible,created_at,updated_at)
SELECT remna_squad_uuid,description,price_txb_minor,visible,created_at,updated_at
FROM squad_products
WHERE description<>'' OR price_txb_minor<>0 OR visible<>0;

ALTER TABLE combos ADD COLUMN included_squad_uuids TEXT NOT NULL DEFAULT '[]'
  CHECK (json_valid(included_squad_uuids) AND json_type(included_squad_uuids)='array');

UPDATE combos
SET included_squad_uuids=COALESCE((
  SELECT json_group_array(remna_squad_uuid)
  FROM (
    SELECT sp.remna_squad_uuid AS remna_squad_uuid
    FROM combo_squads cs
    JOIN squad_products sp ON sp.id=cs.squad_product_id
    WHERE cs.combo_id=combos.id
    ORDER BY sp.remna_squad_uuid
  )
),'[]');

UPDATE coupon_definitions AS coupon
SET eligible_squad_ids=COALESCE((
  SELECT json_group_array(sp.remna_squad_uuid)
  FROM json_each(coupon.eligible_squad_ids) selected
  JOIN squad_products sp ON sp.id=selected.value
),'[]');

DROP TABLE combo_squads;
DROP TABLE squad_products;

-- A purchase keeps transaction state and a live combo reference. Included
-- squads and entitlement configuration are resolved from the combo at use time.
ALTER TABLE purchases RENAME COLUMN price_txb_minor TO charged_txb_minor;
ALTER TABLE purchases DROP COLUMN catalog_snapshot;
ALTER TABLE purchases DROP COLUMN traffic_limit_bytes;
ALTER TABLE purchases DROP COLUMN reset_strategy;
ALTER TABLE purchases DROP COLUMN rollover_min_remaining_bps;
ALTER TABLE purchases DROP COLUMN rollover_max_txb_minor;

CREATE TABLE purchase_addons (
  purchase_id TEXT NOT NULL REFERENCES purchases(id) ON DELETE CASCADE,
  remna_squad_uuid TEXT NOT NULL,
  charged_txb_minor INTEGER NOT NULL DEFAULT 0 CHECK (charged_txb_minor >= 0),
  PRIMARY KEY (purchase_id, remna_squad_uuid)
);

INSERT INTO purchase_addons(purchase_id,remna_squad_uuid,charged_txb_minor)
SELECT purchase_id,remna_squad_uuid,price_txb_minor
FROM purchase_squads
WHERE kind='addon';
DROP TABLE purchase_squads;
CREATE INDEX purchase_addons_squad_idx ON purchase_addons(remna_squad_uuid,purchase_id);
CREATE INDEX purchases_combo_created_idx ON purchases(combo_id,created_at);

-- Durable jobs identify targets through canonical typed payloads only.
CREATE TABLE outbox_jobs_v2 (
  id TEXT PRIMARY KEY,
  kind TEXT NOT NULL,
  payload TEXT NOT NULL CHECK (json_valid(payload)),
  status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','processing','done','failed')),
  attempts INTEGER NOT NULL DEFAULT 0,
  available_at TEXT NOT NULL,
  last_error TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

INSERT INTO outbox_jobs_v2(id,kind,payload,status,attempts,available_at,last_error,created_at,updated_at)
SELECT id,kind,json(payload),status,attempts,available_at,last_error,created_at,updated_at FROM outbox_jobs;

DELETE FROM outbox_jobs_v2
WHERE status IN ('pending','processing')
  AND EXISTS (
    SELECT 1 FROM outbox_jobs_v2 earlier
    WHERE earlier.kind=outbox_jobs_v2.kind
      AND earlier.payload=outbox_jobs_v2.payload
      AND earlier.status IN ('pending','processing')
      AND (earlier.created_at<outbox_jobs_v2.created_at OR (earlier.created_at=outbox_jobs_v2.created_at AND earlier.id<outbox_jobs_v2.id))
  );

DROP TABLE outbox_jobs;
ALTER TABLE outbox_jobs_v2 RENAME TO outbox_jobs;
CREATE INDEX outbox_available_idx ON outbox_jobs(status,available_at);
CREATE UNIQUE INDEX outbox_active_payload_idx ON outbox_jobs(kind,payload)
  WHERE status IN ('pending','processing');

DROP TABLE join_invites;

-- Existing accounts intentionally start with no disabled folders: all
-- selectable Emby libraries become enabled on the next verified sync.
DROP TABLE emby_account_folders;
CREATE TABLE emby_account_disabled_folders (
  account_id TEXT NOT NULL REFERENCES emby_accounts(id) ON DELETE CASCADE,
  folder_id TEXT NOT NULL,
  PRIMARY KEY (account_id,folder_id)
);
ALTER TABLE emby_accounts ADD COLUMN pending_preferences_json TEXT NOT NULL DEFAULT '{}'
  CHECK (json_valid(pending_preferences_json));

INSERT INTO settings(key,value,encrypted,updated_at)
SELECT 'activity.daily_reward_min_txb',value,0,updated_at
FROM settings WHERE key='activity.daily_reward_txb'
ON CONFLICT(key) DO NOTHING;
INSERT INTO settings(key,value,encrypted,updated_at)
SELECT 'activity.daily_reward_max_txb',value,0,updated_at
FROM settings WHERE key='activity.daily_reward_txb'
ON CONFLICT(key) DO NOTHING;
DELETE FROM settings WHERE key='activity.daily_reward_txb';

CREATE INDEX activity_bets_game_created_idx ON activity_bets(game_id,created_at);
CREATE INDEX activity_draw_results_draw_created_idx ON activity_draw_results(draw_id,created_at);
CREATE INDEX questionnaire_participants_questionnaire_awarded_idx
  ON questionnaire_participants(questionnaire_id,awarded_at);
