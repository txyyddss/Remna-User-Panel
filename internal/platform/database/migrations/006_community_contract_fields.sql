ALTER TABLE activity_games ADD COLUMN description TEXT NOT NULL DEFAULT '';
ALTER TABLE activity_lucky_draws ADD COLUMN description TEXT NOT NULL DEFAULT '';
ALTER TABLE questionnaires ADD COLUMN closes_at TEXT;

CREATE INDEX questionnaires_active_closes_idx ON questionnaires(status, closes_at);
