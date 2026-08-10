# Purchase rollover

- `service.go` coordinates queued purchase rollover, remote traffic quiescence, usage capture, and local finalization.
- `service_test.go` verifies operation ordering and safe finalization when local or remote identities are missing.
