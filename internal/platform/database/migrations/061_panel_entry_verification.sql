-- Only Telegram identity and the verified flag persist before onboarding.
CREATE TABLE panel_entry_verification (
    telegram_id INTEGER PRIMARY KEY REFERENCES users(telegram_id) ON DELETE CASCADE,
    checked INTEGER NOT NULL DEFAULT 0 CHECK (checked IN (0,1))
);

-- Collision recovery clears agreement fields only after a completed account is
-- refunded. Preserve that trusted identity until account setup completes again.
INSERT INTO panel_entry_verification(telegram_id,checked)
SELECT telegram_id,1 FROM users WHERE recovery_reason='remnawave_username_conflict';
