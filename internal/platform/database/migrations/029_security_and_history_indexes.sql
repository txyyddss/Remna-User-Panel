DELETE FROM sessions
WHERE rowid NOT IN (SELECT MAX(rowid) FROM sessions GROUP BY user_id);

CREATE UNIQUE INDEX IF NOT EXISTS sessions_user_id_unique_idx ON sessions(user_id);
CREATE INDEX IF NOT EXISTS users_admin_page_idx ON users(created_at DESC,id DESC);
CREATE INDEX IF NOT EXISTS purchases_admin_page_idx ON purchases(created_at DESC,id DESC);
CREATE INDEX IF NOT EXISTS purchases_admin_status_page_idx ON purchases(status,created_at DESC,id DESC);
CREATE INDEX IF NOT EXISTS payment_orders_admin_page_idx ON payment_orders(created_at DESC,id DESC);
CREATE INDEX IF NOT EXISTS payment_orders_admin_status_page_idx ON payment_orders(status,created_at DESC,id DESC);
CREATE INDEX IF NOT EXISTS refunds_admin_status_page_idx ON refunds(status,created_at DESC,id DESC);
CREATE INDEX IF NOT EXISTS outbox_jobs_admin_page_idx ON outbox_jobs(created_at DESC,id DESC);
CREATE INDEX IF NOT EXISTS outbox_jobs_admin_status_page_idx ON outbox_jobs(status,created_at DESC,id DESC);
CREATE INDEX IF NOT EXISTS questionnaire_participants_user_history_idx
  ON questionnaire_participants(user_id,created_at DESC,id DESC);
