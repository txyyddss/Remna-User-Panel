-- Telegram rejected these messages before delivery. Other failed sends may
-- have reached the recipient and remain available for deliberate admin retry.
UPDATE outbox_jobs
SET status='pending', attempts=0, last_error='',
    available_at=strftime('%Y-%m-%dT%H:%M:%fZ','now'),
    updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now')
WHERE status='failed'
  AND kind IN (
    'telegram_user_notification',
    'telegram_affiliate_success',
    'telegram_affiliate_tier_upgrade',
    'telegram_payment_success_announcement'
  )
  AND (
    lower(last_error) LIKE '%can''t parse entities%'
    OR lower(last_error) LIKE '%reserved and must be escaped%'
    OR lower(last_error) LIKE '%can''t find end of the entity%'
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
