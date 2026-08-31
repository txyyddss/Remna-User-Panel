UPDATE users
SET onboarding_state=CASE WHEN username IS NULL THEN 'username' ELSE 'agreement' END
WHERE onboarding_state='membership';
