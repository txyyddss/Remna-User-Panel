# Durable outbox worker

- `worker.go` registers typed handlers and drains claimed durable jobs with bounded retries and persisted outcomes.
- `worker_test.go` verifies dispatch, unknown-kind handling, failure persistence, and registration validation.
