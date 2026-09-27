-- Entity parse errors mean Telegram rejected the send or edit before delivery.
-- Raffle settlement is transactional and can be retried after a code fix.
UPDATE outbox_jobs
SET status='pending', attempts=0, last_error='',
    available_at=strftime('%Y-%m-%dT%H:%M:%fZ','now'),
    updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now')
WHERE status='failed'
  AND (
    kind='draw_raffle_settle'
    OR (kind IN ('telegram_user_notification','draw_telegram_update') AND (
      lower(last_error) LIKE '%can''t parse entities%'
      OR lower(last_error) LIKE '%reserved and must be escaped%'
      OR lower(last_error) LIKE '%can''t find end of the entity%'
    ))
    OR (kind='draw_telegram_update' AND (
      lower(last_error) LIKE '%message is not modified%'
      OR lower(last_error) LIKE '%raffle announcement exceeds telegram limit%'
    ))
  )
  AND id = (
    SELECT MIN(candidate.id) FROM outbox_jobs candidate
    WHERE candidate.kind=outbox_jobs.kind AND candidate.payload=outbox_jobs.payload
      AND candidate.status='failed'
  )
  AND NOT EXISTS (
    SELECT 1 FROM outbox_jobs active
    WHERE active.kind=outbox_jobs.kind AND active.payload=outbox_jobs.payload
      AND active.status IN ('pending','processing')
  );
