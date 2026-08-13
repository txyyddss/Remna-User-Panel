# SQLite store
- `activity_draws_part2.go` continues the focused implementation from its original package module.
- `activity_extensions_part2.go` continues the focused implementation from its original package module.
- `activity_history_part2.go` continues the focused implementation from its original package module.
- `activity_part2.go` continues the focused implementation from its original package module.
- `billing_payments_part2.go` continues the focused implementation from its original package module.
- `coupon_purchase_part2.go` continues the focused implementation from its original package module.
- `coupons_part2.go` continues the focused implementation from its original package module.
- `coupons_part3.go` continues the focused implementation from its original package module.
- `emby_provisioning_part2.go` continues the focused implementation from its original package module.
- `outbox_jobs_part2.go` continues the focused implementation from its original package module.
- `questionnaire_queries_part2.go` continues the focused implementation from its original package module.
- `questionnaires_part2.go` continues the focused implementation from its original package module.
- `rollover_part2.go` continues the focused implementation from its original package module.
- `statistics_part2.go` continues the focused implementation from its original package module.
- `store_part2.go` continues the focused implementation from its original package module.

This package owns the authoritative SQLite connection, migrations, transactional
repositories, durable outbox records, and restore validation. Domain services
depend on the `Store` methods; provider network calls do not belong here.

The persistence implementation is split by domain operation. The `_part2.go` files contain continuation methods for the corresponding bounded repository module.

- `activity_draws_part2.go`, `activity_extensions_part2.go`, and `activity_part2.go` continue activity persistence.
- `billing_payments_part2.go`, `coupons_part2.go`, and `coupons_part3.go` continue billing and coupon persistence.
- `emby_provisioning_part2.go`, `outbox_jobs_part2.go`, `questionnaires_part2.go`, `statistics_part2.go`, and `store_part2.go` continue their focused repositories.

## Production files

- `database.go` — opens SQLite, applies embedded migrations, checkpoints WAL,
  and exposes migration versions.
- `store.go` — core store type, user/session persistence, membership, username,
  and Remnawave recovery state.
- `settings.go` — encrypted-or-plain application setting records.
- `onboarding.go` — versioned onboarding content, agreement contracts, and
  onboarding completion.
- `catalog.go` — combo inputs and transactional combo writes.
- `catalog_queries.go` — combo reads, scans, and normalized squad UUID lists.
- `catalog_squads.go` — sparse local merchandising overrides for upstream squads.
- `billing.go` — purchase normalization, quotes, creation, pricing, and debit helpers.
- `billing_renewal.go` — renewal quotes, contiguous batch creation, and idempotent replay.
- `billing_purchase_helpers.go` — stock reservation checks, purchase fingerprints, catalog row loaders, and balance debit helpers.
- `billing_ledger.go` — balances, audited adjustments, deductions, and ledger reads.
- `ledger_page.go` — stable cursor-based ledger pagination.
- `billing_purchases.go` — purchase reads, cancellation, squad hydration, and
  active/queued selection.
- `billing_payments.go` — payment-order creation, checkout updates, expiry, reads,
  and row scanning.
- `billing_purchase_cancellation.go` — owner-scoped queued cancellation with
  atomic status transition, TXB refund, and immutable ledger entry.
- `billing_payment_settlement.go` — customer cancellation and idempotent provider
  settlement transitions.
- `billing_courtesy.go` — atomic terminal-payment courtesy credits with linked
  immutable ledger and audit records.
- `billing_refunds.go` — transactional refunds, compensating ledger entries, and
  refund history.
- `balance_transactions.go` — shared checked balance mutation helpers used inside
  larger transactions.
- `coupons.go` — coupon definitions, direct redemption, grants, wallet reads, and
  durable member soft-discard records.
- `coupon_purchase.go` — purchase discount quoting, grant consumption, limits,
  and coupon scans.
- `coupon_records.go` — grant/redemption lookup and scan helpers.
- `activity.go` — game configuration, bets, and daily check-ins.
- `activity_draws.go` — lucky-draw configuration, listing, and atomic play results.
- `activity_history.go` — combined activity history and group-message rewards.
- `activity_queries.go` — game, bet, check-in, draw, and result scanners.
- `activity_extensions.go` — durable subscription-extension credits and activation
  application.
- `questionnaires.go` — questionnaire definitions, active selection, participants,
  and participation history.
- `questionnaire_imports.go` — CSV import creation and analysis.
- `questionnaire_settlement.go` — durable settlement queueing and award application.
- `questionnaire_queries.go` — questionnaire/import state and scan helpers.
- `emby.go` — Emby setup debit/outbox creation and account reads.
- `emby_provisioning.go` — retryable provisioning-saga state transitions, refund,
  preferences, and account touch operations.
- `emby_queries.go` — provisioning scans, preference hydration, folder replacement,
  and provider-error normalization.
- `operations.go` — durable job insertion, entitlement-transition enqueueing, and
  purchase expiry.
- `expansion_backup.go` — expansion-state backup and restore helpers.
- `outbox_jobs.go` — outbox claim, completion, retry, deletion, recovery, listing,
  scanning, and persisted-error sanitization.
- `purchase_sync.go` — entitlement synchronization and traffic-reset phase state.
- `administration_records.go` — audit events, administrator user lists, and backup
  run records.
- `rollover.go` — durable rollover processing and finalization.
- `payment_profiles.go` — provider-account profile masking and encrypted credential persistence with per-account channel lists.
- `retention.go` — bounded cleanup of aged operational records.
- `statistics.go` — catalog and activity administrator statistics.
- `destructive.go` — audited feature deletion transactions.
- `restore.go` — restore snapshot preparation and high-level validation.
- `restore_schema.go` — canonical schema-shape types and database introspection.
- `restore_schema_tables.go` — per-table columns, foreign keys, indexes, objects,
  and schema SQL normalization.

## Test files

- `store_test.go` — core users, catalog guards, and store behavior.
- `billing_purchase_test.go` — purchase creation, quoting, and idempotency.
- `billing_balance_test.go` — concurrent and bounded balance mutations.
- `billing_payment_test.go` — payment settlement and deduplication.
- `billing_refund_test.go` — refund, cancellation, and debt behavior.
- `billing_test_helpers_test.go` — shared billing fixtures and payment constructors.
- `activity_bet_test.go` — atomic bet outcomes and replay.
- `activity_daily_test.go` — daily check-in reward boundaries.
- `activity_draw_test.go` — lucky-draw prizes, extensions, and replay.
- `activity_group_reward_test.go` — group-message counting and rewards.
- `activity_test_helpers_test.go` — shared activity random source.
- `coupon_features_test.go` — coupon redemption and purchase discounts.
- `questionnaire_features_test.go` — questionnaire CSV analysis and settlement.
- `emby_test.go` — Emby setup, provisioning, retry, and refund atomicity.
- `rollover_test.go` — rollover transitions and calculations.
- `retention_test.go` — retained operational-record cleanup behavior.
- `restore_test.go` — restore validation and schema compatibility.
- `ledger_page_test.go` — cursor pagination order and validation.

`payment_profiles_test.go` covers independent provider-account persistence and stable-ID lookup.

The `migrations/` directory contains the ordered embedded schema history. New
schema changes must be additive migrations; deployed migration files are immutable.

Migration `015_payment_provider_profiles.sql` consolidates the temporary
per-rail profile table into one row per provider. Migration
`016_multiple_payment_profiles.sql` removes the provider-wide uniqueness
constraint so additional independent accounts can be stored. Legacy settings
remain available for decryption fallback, while new writes use the stable
profile record and preserve masked credentials.

`billing_courtesy_test.go` covers terminal-payment courtesy-credit atomicity,
idempotent replay, and late-provider-callback blocking.

`renewal_coupon.go` provides read-only recurring coupon reuse for renewal pricing
without coupon-use writes. `renewal_batch.go` projects renewal batches and their
purchase records.
