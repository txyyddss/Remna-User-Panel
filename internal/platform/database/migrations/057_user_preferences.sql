CREATE TABLE user_preferences (
  user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  notify_combos INTEGER NOT NULL DEFAULT 1 CHECK (notify_combos IN (0,1)),
  notify_traffic INTEGER NOT NULL DEFAULT 1 CHECK (notify_traffic IN (0,1)),
  notify_money INTEGER NOT NULL DEFAULT 1 CHECK (notify_money IN (0,1)),
  notify_activity INTEGER NOT NULL DEFAULT 1 CHECK (notify_activity IN (0,1)),
  notify_account INTEGER NOT NULL DEFAULT 1 CHECK (notify_account IN (0,1)),
  show_referral_username INTEGER NOT NULL DEFAULT 1 CHECK (show_referral_username IN (0,1)),
  show_around_tx INTEGER NOT NULL DEFAULT 0 CHECK (show_around_tx IN (0,1)),
  show_activity INTEGER NOT NULL DEFAULT 0 CHECK (show_activity IN (0,1)),
  include_node_prices INTEGER NOT NULL DEFAULT 1 CHECK (include_node_prices IN (0,1))
);

-- Clear optional entrances on loss of access and before a later purchase can
-- restore access. Use each mutation's application timestamp, not the DB clock.
CREATE TRIGGER preference_visibility_before_purchase
BEFORE INSERT ON purchases
BEGIN
  UPDATE user_preferences SET show_around_tx=0,show_activity=0
  WHERE user_id=NEW.user_id AND NOT EXISTS (
    SELECT 1 FROM purchases WHERE user_id=NEW.user_id AND status='active'
      AND julianday(valid_from)<=julianday(NEW.created_at)
      AND julianday(valid_until)>julianday(NEW.created_at)
  );
END;

CREATE TRIGGER preference_visibility_before_transition
BEFORE UPDATE OF status,valid_from,valid_until ON purchases
BEGIN
  UPDATE user_preferences SET show_around_tx=0,show_activity=0
  WHERE user_id=NEW.user_id AND NOT EXISTS (
    SELECT 1 FROM purchases WHERE user_id=NEW.user_id AND status='active'
      AND julianday(valid_from)<=julianday(NEW.updated_at)
      AND julianday(valid_until)>julianday(NEW.updated_at)
  );
END;

CREATE TRIGGER preference_visibility_after_transition
AFTER UPDATE OF status,valid_from,valid_until ON purchases
BEGIN
  UPDATE user_preferences SET show_around_tx=0,show_activity=0
  WHERE user_id=NEW.user_id AND NOT EXISTS (
    SELECT 1 FROM purchases WHERE user_id=NEW.user_id AND status='active'
      AND julianday(valid_from)<=julianday(NEW.updated_at)
      AND julianday(valid_until)>julianday(NEW.updated_at)
  );
END;
