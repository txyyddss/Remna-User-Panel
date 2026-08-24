ALTER TABLE abuse_policy ADD COLUMN warning_cooldown_minutes INTEGER NOT NULL DEFAULT 30 CHECK (warning_cooldown_minutes BETWEEN 0 AND 525600);
