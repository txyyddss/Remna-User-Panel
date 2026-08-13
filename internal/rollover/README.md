# Purchase rollover
- `service_part2.go` continues the focused implementation from its original package module.

- `service.go` coordinates queued purchase rollover, remote traffic quiescence, cadence-aware daily usage capture, and aggregate-only local finalization.
- `service_test.go` verifies operation ordering and safe finalization when local or remote identities are missing.
- `service_part2.go` contains cadence and aggregate usage calculations.
- `projection.go` shares the cadence evaluator for live current-term/reset-period projections, integer-safe rollover caps, maximum thresholds, and net-paid savings percentages.
- `projection_test.go` covers net-paid savings, latest reset selection, strict thresholds, and unreachable maximum states.
