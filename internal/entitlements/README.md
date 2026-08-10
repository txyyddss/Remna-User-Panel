# Entitlements package

The durable entitlement worker translates outbox jobs and desired purchase state into idempotent Remnawave mutations.

## Files

- `worker.go` defines dependencies and implements job draining, dispatch, retry, expiration, and entitlement synchronization.
- `worker_test.go` covers draining, retry behavior, traffic-reset resumption, and infrastructure errors.
- `worker_process_test.go` covers individual processing branches, expiration, desired-state synchronization, and unknown jobs.
- `worker_test_helpers_test.go` contains shared repository and Remnawave test doubles.
- `README.md` documents the package layout.
