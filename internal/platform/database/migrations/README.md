# Database migrations

Migration files are embedded and applied in lexical order by `database.go`.
Never edit a migration that may already have been deployed; add the next numbered
file instead.

- `001_initial.sql` — core users, sessions, catalog, purchases, ledger, settings,
  outbox, audit, backup, and onboarding schema.
- `002_activity_coupons_questionnaires.sql` — community activity, coupon, and
  questionnaire aggregates.
- `003_emby.sql` — Emby account and provisioning saga tables.
- `004_platform_payments_rollover_restore.sql` — expanded payment, rollover, and
  restore lifecycle records.
- `005_database_admin_reviews.sql` — database administration review support.
- `006_community_contract_fields.sql` — additional community contract fields.
- `007_purchase_idempotency.sql` — purchase request fingerprinting and replay keys.
- `008_group_message_rewards.sql` — Telegram group-message reward tracking.
- `009_expansion_and_cleanup.sql` — current expansion, normalization, and obsolete
  snapshot cleanup.
- `010_clear_subscription_cache.sql` — clears legacy Remnawave subscription bearer
  values while retaining the nullable column for backup/restore schema compatibility.
- `011_minimize_payment_payloads.sql` — clears terminal provider display payloads
  that are not required for settlement.
- `012_minimize_questionnaire_imports.sql` — removes settled CSV payloads and
  aligns exhausted questionnaire jobs with the retry lifecycle.
- `013_coupon_discards_and_courtesy_credits.sql` — adds durable wallet-discard
  evidence and exactly-once terminal-payment courtesy-credit records.
- `014_goal_completion.sql` adds optional squad stock, renewal batches, bounded
  rollover aggregates, and the temporary per-rail payment profile seed.
- `015_payment_provider_profiles.sql` consolidates each provider into one
  profile row with independently enabled channels and removes the temporary
  per-rail table.
- `016_multiple_payment_profiles.sql` allows multiple independent EZPay and
  BEPusdt accounts while preserving the existing provider profile rows.
- `017_squad_profiles.sql` adds nullable normalized local metadata for the
  three customer-facing internal-squad profile types without duplicating the
  generated description.
- `018_automatic_renewal.sql` adds opt-in automatic-renewal state, a durable
  due-cycle failure notice, recurring-discount attachment, and a unique
  source-successor link while leaving legacy renewal records intact.
