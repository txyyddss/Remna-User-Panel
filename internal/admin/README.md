# Admin package

Audited administrative operations for catalog data, settings, balances, refunds, purchases, backups, and durable jobs live here.

## Files

- `service.go` defines admin-facing dependency contracts, service construction, backup deletion, and settings forwarding.
- `catalog.go` manages combos, squad products, upstream squad imports, and squad-node assignments.
- `finance.go` manages balance adjustments, deductions, refunds, terminal-payment courtesy credits, entitlement cancellation, backups, job retries, and audit recording.
- `settings.go` defines editable settings, encrypted storage access, masked payment profiles, safe listing, and readiness checks.
- `payment_profiles.go` validates, encrypts, and resolves one provider-level payment profile with independently enabled channels.
- `payment_profile_readiness.go` maps enabled provider profiles to stable readiness diagnostics.
- `setting_validators.go` contains the individual setting-value validators.
- `service_test.go` covers settings forwarding and catalog operations.
- `finance_test.go` covers balance, refund, cancellation, backup, and retry operations.
- `service_test_helpers_test.go` contains shared admin service test doubles.
- `settings_test.go` covers secret handling, safe listing, readiness, and validators.
- `catalog_test.go` covers live squad-node normalization, node updates, and optional setting validators.
- `README.md` documents the package layout.

Provider profiles share one endpoint, credential, merchant ID, provider name,
and acknowledgement per provider. Channel selection only controls which
provider-owned rails are offered. Balance adjustments create unique ledger
references so repeated reasons remain separate audited operations.
