# Purchase rollover

- `service.go` coordinates queued purchase rollover, remote traffic quiescence, cadence-aware daily usage capture, and aggregate-only local finalization.
- `service_test.go` verifies operation ordering and safe finalization when local or remote identities are missing.
