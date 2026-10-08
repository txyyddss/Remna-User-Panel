-- Boost changes and redelivered updates must not publish a second appreciation.
CREATE TABLE telegram_boost_receipts (
    chat_id INTEGER NOT NULL,
    boost_id TEXT NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (chat_id, boost_id)
);
