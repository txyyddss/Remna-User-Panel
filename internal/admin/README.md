# Admin package
- `settings_part2.go` continues the focused implementation from its original package module.

Audited administrative operations for catalog data, settings, balances, refunds, purchases, backups, and durable jobs live here.

## Files

- `service.go` defines admin-facing dependency contracts, service construction, backup deletion, and settings forwarding.
- `catalog.go` manages combos, typed squad profiles, and upstream squad imports; Remnawave remains authoritative for node accessibility.
- `finance.go` manages balance adjustments, deductions, refunds, terminal-payment courtesy credits, entitlement cancellation, backups, job retries, and audit recording.
- `settings.go` defines editable settings, encrypted storage access, masked payment profiles, safe listing, and readiness checks.
- `payment_profiles.go` validates, encrypts, and resolves multiple provider-account profiles with independently enabled channels.
- `payment_profile_readiness.go` maps enabled provider profiles to stable readiness diagnostics.
- `setting_validators.go` contains the individual setting-value validators.
- `service_test.go` covers settings forwarding and catalog operations.
- `finance_test.go` covers balance, refund, cancellation, backup, and retry operations.
- `service_test_helpers_test.go` contains shared admin service test doubles.
- `settings_test.go` covers secret handling, safe listing, readiness, and validators.
- `payment_profiles_test.go` covers provider profile validation, encryption, masking, and readiness behavior.
- `zero_coverage_test.go` covers backup deletion, balance deductions, stable-ID payment-profile lookups and runtimes, legacy EZPay validation, and activity reward boundaries.
- `catalog_test.go` covers optional setting validators.
- `README.md` documents the package layout.

Provider profiles have stable IDs, separate endpoint, credential, merchant ID,
provider name, and acknowledgement values. Channel selection only controls
which rails are offered for that account. Balance adjustments create unique
ledger references so repeated reasons remain separate audited operations.

- `settings_part2.go` contains the remaining settings validation and persistence helpers.
