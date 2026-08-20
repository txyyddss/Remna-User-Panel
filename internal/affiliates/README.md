# Affiliates domain

This package defines referral, immutable tier configuration, reward, member-page,
and administrator contracts. Persistence owns atomic settlement and projection
queries; this package owns portable validation and locale normalization.

## Files

- `types.go` defines configuration, rewards, money, and bot identity.
- `projections.go` defines member and administrator API projections.
- `validation.go` enforces tier ordering, monetary bounds, and reward unions.
- `bot_identity.go` caches validated queued `getMe` results with stale fallback.
- `notifications.go` formats and sends localized durable MarkdownV2 jobs.
- `service.go` coordinates member, administrator, referral, and discovery operations.
- `validation_test.go` covers locale normalization and tier invariants without provider calls.
- `bot_identity_test.go` covers 24-hour caching and stale-on-transient-failure behavior.
- `notifications_test.go` covers dynamic MarkdownV2 escaping and fixed-point amount copy.
