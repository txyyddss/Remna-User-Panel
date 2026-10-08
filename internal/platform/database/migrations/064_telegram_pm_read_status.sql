-- Store only per-delivery read evidence; Telegram message bodies remain upstream.
CREATE TABLE telegram_pm_read_status (
    operation_id TEXT PRIMARY KEY REFERENCES provider_operations(id) ON DELETE CASCADE,
    read_at TEXT NOT NULL,
    source TEXT NOT NULL CHECK (source IN ('explicit','reply'))
);
