# Admin package

Audited administrative operations for catalog data, settings, balances, refunds, purchases, backups, and durable jobs live here.

## Files

- `service.go` defines admin-facing dependency contracts, service construction, backup deletion, and settings forwarding.
- `catalog.go` manages combos, squad products, upstream squad imports, and squad-node assignments.
- `finance.go` manages balance adjustments, deductions, refunds, terminal-payment courtesy credits, entitlement cancellation, backups, job retries, and audit recording.
- `settings.go` defines editable settings, encrypted storage access, safe listing, and readiness checks.
- `setting_validators.go` contains the individual setting-value validators.
- `service_test.go` covers settings forwarding and catalog operations.
- `finance_test.go` covers balance, refund, cancellation, backup, and retry operations.
- `service_test_helpers_test.go` contains shared admin service test doubles.
- `settings_test.go` covers secret handling, safe listing, readiness, and validators.
- `catalog_test.go` covers live squad-node normalization, node updates, and optional setting validators.
- `README.md` documents the package layout.
