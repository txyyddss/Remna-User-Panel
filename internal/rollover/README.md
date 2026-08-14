# Purchase rollover
- `service_part2.go` continues the focused implementation from its original package module.

- `service.go` coordinates queued purchase rollover, remote traffic quiescence, cadence-aware daily usage capture, and aggregate-only local finalization.
- `cadence.go` owns reset-window boundaries, including calendar-month and fixed rolling-month arithmetic.
- `service_test.go` verifies operation ordering and safe finalization when local or remote identities are missing.
- `service_part2.go` contains cadence and aggregate usage calculations. Daily
  usage buckets are clipped to both the term range and each reset period, so a
  non-midnight DAY or WEEK boundary cannot count one date in two periods.
- The newest reset period uses Remnawave's authoritative current used-traffic
  counter when available; earlier periods continue to use bounded daily buckets.
  `MONTH_ROLLING` advances by a fixed 30-day window, while `MONTH` follows
  calendar-month boundaries.
- `projection.go` shares the cadence evaluator for live current-term/reset-period projections, integer-safe rollover caps, maximum thresholds, and net-paid savings percentages.
- `projection_test.go` covers net-paid savings, latest reset selection, strict thresholds, and unreachable maximum states.
- `projection_current_counter_test.go` covers authoritative current usage,
  daily fallback, zero usage, and all supported reset cadences.
