# User notifications

- `facts.go` owns stable immutable payload fact keys.
- `copy.go` owns matching English and Simplified Chinese notification copy.
- `format.go`, `format_fields.go`, and `format_admin.go` render concise,
  escaped MarkdownV2 cards with one title emoji and ordered detail rows.
- `worker.go` delivers durable private-chat jobs through the queued Telegram sender.
- `scanner.go` evaluates the 48-hour reminder, strict 90% notice boundary, and strict-above-99% automatic-reset handoff.
- Automatic-reset success is provider-gated; insufficient balance disables the preference in the reset transaction, while definitive failure produces a localized refund notice.
- `format_test.go`, `automatic_reset_format_test.go`, `scanner_test.go`, and `worker_test.go` cover formatting,
  localized reset details, threshold behavior, and retryable delivery failures.
