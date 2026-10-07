PRAGMA foreign_keys = ON;

-- Store only explicit disabled references; purchased squads remain authoritative.
CREATE TABLE user_disabled_squads (
 user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 squad_uuid TEXT NOT NULL,
 PRIMARY KEY(user_id,squad_uuid)
);
