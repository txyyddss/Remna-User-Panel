# Emby package
- `types_part2.go` continues the focused implementation from its original package module.

Priced Emby setup, durable provisioning, policy hardening, preference updates, and password changes live here.

## Files

- `types.go` defines safe domain types, dependency interfaces, secret adapters, and policy-hardening helpers.
- `service.go` constructs the service and handles account lookup, options, retries, and setup queueing.
- `provisioning.go` dispatches and performs durable provisioning jobs.
- `provisioning_reconcile.go` resolves ambiguous creates, selects collision-safe usernames, validates preferences, and classifies provisioning failures.
- `preferences.go` applies linked-account preferences and password changes.
- `helpers.go` contains username, preference normalization, password-memory, and redacted-error helpers.
- `service_test.go` covers policy hardening, setup, linked updates, and secret adapters.
- `provisioning_test.go` covers provisioning failure, recovery, reconciliation, and validation branches.
- `service_test_helpers_test.go` contains shared repository, remote, price, secret, and cipher doubles.
- `README.md` documents the package layout.
- `types_part2.go` contains policy hardening and policy comparison helpers.
