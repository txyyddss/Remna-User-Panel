# Admin package
- `settings_part2.go` continues the focused implementation from its original package module.

Audited administrative operations for catalog data, settings, balances, refunds, purchases, backups, and durable jobs live here.

## Files

- `service.go` defines admin-facing dependency contracts, service construction, backup deletion, and settings forwarding.
- `catalog.go` manages combos, typed squad profiles, and upstream squad imports; Remnawave remains authoritative for node accessibility.
- `finance.go` validates balance adjustments, deductions, refunds,
  terminal-payment courtesy credits, and entitlement cancellations; production
  persistence commits their audit and user notification atomically.
- `settings.go` defines editable settings, encrypted storage access, masked payment profiles, safe listing, and readiness checks.
- `payment_profiles.go` validates, encrypts, and resolves multiple provider-account profiles with independently enabled channels.
- `payment_profile_readiness.go` maps enabled provider profiles to stable readiness diagnostics.
- `setting_validators.go` contains the individual setting-value validators.
- `mutation_operations.go` creates idempotent administrator outbox-retry and payment-refund receipts.
- `mutation_worker.go` dispatches administrator mutation commands from the shared provider-operation lane.
- `mutation_lifecycle.go` marks administrator operation items and receipts through one bounded lifecycle.
- `refund_worker.go` reconciles Telegram Stars outcomes before applying one local payment refund.
- `user_profiles.go` builds the user aggregate from identity, balance, active and queued entitlements, Emby, payments, refunds, and open operations.
- `user_command_types.go` defines the complete entitlement, combo replacement, and bulk-extension command inputs.
- `user_command_validation.go` normalizes IDs and squads and fingerprints idempotent administrator commands.
- `user_entitlement_commands.go` validates full edits, exactly-once refunds, and no-charge combo replacements.
- `user_bulk_commands.go` validates inclusive-OR bulk previews and durable extension jobs.
- `user_operation_resolution.go` validates idempotent audited resolutions for ambiguous provider outcomes without dispatching another provider call.
- `user_operation_worker.go` applies exact local entitlement state through the
  shared provider-operation dispatcher and releases user messages only after a
  real successful provider item.
- `service_test.go` covers settings forwarding and catalog operations.
- `finance_test.go` covers balance, refund, cancellation, backup, and retry operations.
- `service_test_helpers_test.go` contains shared admin service test doubles.
- `settings_test.go` covers secret handling, safe listing, readiness, and validators.
- `payment_profiles_test.go` covers provider profile validation, encryption, masking, and readiness behavior.
- `zero_coverage_test.go` covers backup deletion, balance deductions, stable-ID payment-profile lookups and runtimes, legacy EZPay validation, and activity reward boundaries.
- `catalog_test.go` covers optional setting validators.
- `user_profiles_test.go` covers aggregate selection, ordering, and synchronization projection.
- `user_profiles_test_helpers_test.go` contains focused aggregate-profile repository doubles.
- `user_operation_worker_test.go` covers ambiguous provider outcomes and partial bulk completion.
- `user_operation_worker_test_helpers_test.go` contains reusable provider-operation worker fixtures.
- `README.md` documents the package layout.

Provider profiles have stable IDs, separate endpoint, credential, merchant ID,
provider name, and acknowledgement values. Channel selection only controls
which rails are offered for that account. Balance adjustments create unique
ledger references so repeated reasons remain separate audited operations.

- `settings_part2.go` contains the remaining settings validation and persistence helpers.
