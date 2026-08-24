# Database migrations

Migration files are embedded and applied in lexical order by `database.go`.

- `030_affiliates.sql` adds referral eligibility, immutable tier versions, settlements, and awards.
- `031_user_notifications.sql` adds durable semantic notification deduplication
  and provider-success release gates.
- `032_automatic_traffic_reset.sql` adds the account-wide opt-in preference
  used by the five-minute traffic scanner.
- `033_node_compensation.sql` adds nullable revisioned policy, persistent node state, immutable outage snapshots, frozen recipients, and reviewed operation linkage.
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
- `019_rollover_activation_codes.sql` removes the persisted rollover cap and
  adds bcrypt-backed squad activation metadata.
- `020_release1_commerce.sql` converts monthly cadence to `MONTH_ROLLING`, snapshots immutable core price, and adds billing limits, provider receipts, connection scans, upload metadata, and replay facts.
- `021_admin_workflows.sql` adds purchase-level entitlement overrides and the durable backup-upload publication saga.
- `022_release1_rollups.sql` adds cleanup leases, compact activity/payment/purchase rollups, and timestamped statistics partitions.
- `023_payment_callback_tombstones.sql` keeps compact provider-callback replay
  evidence after terminal payment detail is pruned.
- `024_operation_durability.sql` allows review-required connection scans and
  preserves actor-scoped staged-restore replay identity across a database swap.
- `025_mutation_durability.sql` adds review-required Emby state and prevents concurrent open setup commands.
- `026_group_message_facts.sql` stores deduplicated raw configured-group non-command message facts for cumulative statistics.
- `027_host_operation_guard.sql` prevents a new host remark mutation while the same host still has an unresolved operation.
- `028_connection_ip_blocks.sql` stores only active encrypted node-scoped IP
  blocks and their scheduled three-day cleanup references.
- `029_security_and_history_indexes.sql` bounds each user to one current session
  and indexes stable administrative inventories plus questionnaire history.
- `034_abuse_qps_detector.sql` adds encrypted node credentials, privacy-safe QPS samples, detector state, incidents, delivery records, and temporary-ban restoration facts.
- `035_abuse_warning_cooldown.sql` adds the revisioned administrator warning-record cooldown.
- `036_purchase_addon_adjustments.sql` records idempotent, prorated active-term squad additions without duplicating user or provider data.
