# User notifications

- `facts.go` owns stable immutable payload fact keys.
- `copy.go` owns matching English and Simplified Chinese notification copy.
- `format.go`, `format_fields.go`, and `format_admin.go` render concise,
  escaped MarkdownV2 cards with one title emoji and ordered detail rows.
- `worker.go` delivers durable private-chat jobs through the queued Telegram sender.
- `scanner.go` evaluates the 48-hour reminder and strict 90% traffic boundary.
- `format_test.go`, `scanner_test.go`, and `worker_test.go` cover formatting,
  threshold behavior, and retryable delivery failures.
