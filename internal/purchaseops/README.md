# Member purchase operations

- `types.go` owns reset/refund contracts and stable reason codes.
- `eligibility.go` applies immutable reset pricing and refund eligibility rules.
- `service.go` creates idempotent durable operations without provider calls in transactions.
- `worker.go` owns shared durable phases and receipt completion.
- `reset_worker.go` reconciles reset timestamps and requests once-only compensation.
- `refund_worker.go` quiesces, rechecks usage, restores conflicts, and commits refunds.
- `eligibility_test.go` covers cadence rounding and the strict refund boundary.

Provider calls are supplied by the application adapter and always traverse the shared Remnawave queue. Tests are added for CI and are not run locally.
