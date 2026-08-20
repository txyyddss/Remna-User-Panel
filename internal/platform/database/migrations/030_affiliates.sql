ALTER TABLE users ADD COLUMN new_user INTEGER NOT NULL DEFAULT 0 CHECK (new_user IN (0,1));
ALTER TABLE users ADD COLUMN inviter_id INTEGER REFERENCES users(telegram_id);
ALTER TABLE users ADD COLUMN notification_locale TEXT NOT NULL DEFAULT 'en' CHECK (notification_locale IN ('en','zh-CN'));

CREATE INDEX users_inviter_created_idx ON users(inviter_id,created_at DESC);

CREATE TABLE affiliate_config_versions (
  id TEXT PRIMARY KEY,
  version INTEGER NOT NULL UNIQUE CHECK (version > 0),
  created_by TEXT REFERENCES users(id),
  created_at TEXT NOT NULL
);

CREATE TABLE affiliate_tiers (
  id TEXT PRIMARY KEY,
  config_id TEXT NOT NULL REFERENCES affiliate_config_versions(id),
  position INTEGER NOT NULL CHECK (position >= 0),
  name TEXT NOT NULL,
  success_threshold INTEGER NOT NULL CHECK (success_threshold >= 0),
  enabled INTEGER NOT NULL CHECK (enabled IN (0,1)),
  commission_enabled INTEGER NOT NULL CHECK (commission_enabled IN (0,1)),
  commission_bps INTEGER NOT NULL CHECK (commission_bps BETWEEN 0 AND 10000),
  reward_kind TEXT NOT NULL CHECK (reward_kind IN ('none','coupon','txb','subscription_extension')),
  reward_coupon_id TEXT REFERENCES coupon_definitions(id),
  reward_txb_minor INTEGER CHECK (reward_txb_minor > 0),
  reward_extension_days INTEGER CHECK (reward_extension_days BETWEEN 1 AND 3650),
  CHECK (
    (reward_kind='none' AND reward_coupon_id IS NULL AND reward_txb_minor IS NULL AND reward_extension_days IS NULL) OR
    (reward_kind='coupon' AND reward_coupon_id IS NOT NULL AND reward_txb_minor IS NULL AND reward_extension_days IS NULL) OR
    (reward_kind='txb' AND reward_coupon_id IS NULL AND reward_txb_minor IS NOT NULL AND reward_extension_days IS NULL) OR
    (reward_kind='subscription_extension' AND reward_coupon_id IS NULL AND reward_txb_minor IS NULL AND reward_extension_days IS NOT NULL)
  ),
  UNIQUE(config_id,position)
);
CREATE UNIQUE INDEX affiliate_tiers_enabled_threshold_idx ON affiliate_tiers(config_id,success_threshold) WHERE enabled=1;

CREATE TABLE affiliate_config_current (
  singleton INTEGER PRIMARY KEY CHECK (singleton=1),
  config_id TEXT NOT NULL REFERENCES affiliate_config_versions(id)
);

INSERT INTO affiliate_config_versions(id,version,created_at)
VALUES('affiliate-config-1',1,strftime('%Y-%m-%dT%H:%M:%fZ','now'));
INSERT INTO affiliate_tiers(id,config_id,position,name,success_threshold,enabled,commission_enabled,commission_bps,reward_kind)
VALUES('affiliate-tier-1','affiliate-config-1',0,'Starter',0,1,0,0,'none');
INSERT INTO affiliate_config_current(singleton,config_id) VALUES(1,'affiliate-config-1');

CREATE TABLE affiliate_settlements (
  id TEXT PRIMARY KEY,
  invited_user_id TEXT NOT NULL UNIQUE REFERENCES users(id),
  inviter_user_id TEXT NOT NULL REFERENCES users(id),
  payment_order_id TEXT NOT NULL UNIQUE REFERENCES payment_orders(id),
  config_id TEXT NOT NULL REFERENCES affiliate_config_versions(id),
  tier_id TEXT NOT NULL REFERENCES affiliate_tiers(id),
  tier_name TEXT NOT NULL,
  commission_bps INTEGER NOT NULL CHECK (commission_bps BETWEEN 0 AND 10000),
  topup_txb_minor INTEGER NOT NULL CHECK (topup_txb_minor > 0),
  commission_txb_minor INTEGER NOT NULL CHECK (commission_txb_minor >= 0),
  settled_at TEXT NOT NULL
);
CREATE INDEX affiliate_settlements_inviter_idx ON affiliate_settlements(inviter_user_id,settled_at DESC,id DESC);

CREATE TABLE affiliate_tier_awards (
  id TEXT PRIMARY KEY,
  inviter_user_id TEXT NOT NULL REFERENCES users(id),
  tier_id TEXT NOT NULL REFERENCES affiliate_tiers(id),
  settlement_id TEXT NOT NULL UNIQUE REFERENCES affiliate_settlements(id),
  tier_name TEXT NOT NULL,
  reward_kind TEXT NOT NULL CHECK (reward_kind IN ('none','coupon','txb','subscription_extension')),
  reward_description TEXT NOT NULL,
  reward_coupon_id TEXT REFERENCES coupon_definitions(id),
  reward_txb_minor INTEGER,
  reward_extension_days INTEGER,
  awarded_at TEXT NOT NULL,
  CHECK (
    (reward_kind='none' AND reward_coupon_id IS NULL AND reward_txb_minor IS NULL AND reward_extension_days IS NULL) OR
    (reward_kind='coupon' AND reward_coupon_id IS NOT NULL AND reward_txb_minor IS NULL AND reward_extension_days IS NULL) OR
    (reward_kind='txb' AND reward_coupon_id IS NULL AND reward_txb_minor IS NOT NULL AND reward_extension_days IS NULL) OR
    (reward_kind='subscription_extension' AND reward_coupon_id IS NULL AND reward_txb_minor IS NULL AND reward_extension_days IS NOT NULL)
  ),
  UNIQUE(inviter_user_id,tier_id)
);

CREATE TRIGGER affiliate_config_versions_no_update BEFORE UPDATE ON affiliate_config_versions BEGIN SELECT RAISE(ABORT,'affiliate config history is immutable'); END;
CREATE TRIGGER affiliate_config_versions_no_delete BEFORE DELETE ON affiliate_config_versions BEGIN SELECT RAISE(ABORT,'affiliate config history is immutable'); END;
CREATE TRIGGER affiliate_tiers_no_update BEFORE UPDATE ON affiliate_tiers BEGIN SELECT RAISE(ABORT,'affiliate tier history is immutable'); END;
CREATE TRIGGER affiliate_tiers_no_delete BEFORE DELETE ON affiliate_tiers BEGIN SELECT RAISE(ABORT,'affiliate tier history is immutable'); END;
