-- Earlier recovery sent fully onboarded members back to agreement and cleared
-- acceptance time after a confirmed provider 404. Keep their prior revision.
UPDATE users
SET onboarding_state='complete',
    policy_accepted_at=COALESCE(policy_accepted_at, updated_at),
    recovery_reason='',
    updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now')
WHERE onboarding_state='agreement'
  AND recovery_reason='remnawave_user_missing'
  AND username IS NOT NULL;
