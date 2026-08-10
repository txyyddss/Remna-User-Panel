-- Remnawave subscription URLs are bearer credentials and must not persist.
-- Keep the nullable column for backup/restore schema compatibility.
UPDATE users
SET remna_subscription_url = NULL
WHERE remna_subscription_url IS NOT NULL;
