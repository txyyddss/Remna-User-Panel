PRAGMA foreign_keys = ON;

ALTER TABLE users ADD COLUMN auto_traffic_reset_enabled INTEGER NOT NULL DEFAULT 0
  CHECK (auto_traffic_reset_enabled IN (0,1));
