# User notifications

- `facts.go` owns stable immutable payload fact keys.
- `copy.go` owns matching English and Simplified Chinese notification copy.
- `format.go`, `format_fields.go`, and `format_admin.go` render concise,
  escaped MarkdownV2 cards with one title emoji and ordered detail rows.
- Immediate combo activation, successful Add TXB settlement, and provider refunds
  enqueue private receipts after their authoritative state changes. Refund cards
  show TXB deducted and any cancelled combos; they wait for user sync when needed.
- Expiry reminders say the subscription is expiring soon because scans can run at
  any point in the 48-hour window. Renewal charges display a positive amount,
  and reset refund reasons use localized explanations.
- `format_compensation.go` renders provider-gated outage compensation cards with
  the frozen node, squads, observation interval, applied duration, expiry change,
  review reason, and capped-calculation indicator.
- `worker.go` delivers durable private-chat jobs through the queued Telegram sender.
- `scanner.go` evaluates the 48-hour reminder, strict 90% notice boundary, and strict-above-95% automatic-reset handoff.
- Automatic-reset success is provider-gated; insufficient balance disables the preference in the reset transaction, while definitive failure produces a localized refund notice.
- `format_test.go`, `automatic_reset_format_test.go`, `scanner_test.go`, and `worker_test.go` cover formatting,
  localized reset details, threshold behavior, and retryable delivery failures.
- `format_compensation_test.go` covers localized compensation details and MarkdownV2 escaping.
- `format_receipts_test.go` covers both locales for new receipts and reset failures.
