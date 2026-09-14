ALTER TABLE users ADD COLUMN display_currency TEXT NOT NULL DEFAULT 'TXB'
  CHECK (display_currency IN ('TXB','CNY','USD'));
