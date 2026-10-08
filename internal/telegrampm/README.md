# Qualified Telegram private messaging

This module routes ordinary bot messages to a configured forum and lets configured
panel admins reply from its per-user topic. It reuses the provider-operation
dispatcher, durable outbox and bounded Telegram executor. Bodies, entities,
captions and files remain in Telegram; `copyMessage` receives only references.

- `service.go` checks first panel-entry evidence, settings and user moderation,
  then queues incoming/outgoing references. Commands and payment service messages
  remain on their existing paths. Configured admin webhook identities may refresh
  their existing account metadata; this never initializes panel-entry evidence.
- `callbacks.go` authorizes panel-admin profile refresh/moderation and recipient
  read acknowledgments against the exact conversation and message references.
- `copy.go` owns localized PM guidance and topic names; `profile_card.go` renders
  account facts, refresh/moderation controls, and the Main Mini App deep link.
  The injected profile reader builds aggregates outside Telegram's queued request callback.
- `worker.go` sequences durable topic/profile/relay/repair phases and revalidates
  configuration and identity inside each queued provider callback.
- `worker_interfaces.go` defines injected repositories and queued sender factories.
- `worker_topic.go` claims creation atomically and persists the returned topic ID
  before publishing a profile card. A restarted ambiguous creation needs review.
- `worker_profile.go` edits known cards idempotently from current account and usage
  projections; append-only card ambiguity persists separately from delivery receipts.
- `worker_relay.go` preserves formatting/captions via message references, applies
  mute state at execution time, and adds a read-ack callback to outbound copies.
  Interrupted copies are never replayed automatically.
- `worker_repair.go` probes explicitly supplied topic/card references; returned
  Telegram chat/thread identity must match before binding. Only unsent waiters resume.
- `worker_errors.go` distinguishes preflight failures, definitive rejections and
  ambiguity. Only definitive missing/deleted-topic responses permit recreation.
  Telegram429 follows the existing bounded outbox retry policy. No second queue exists.

Topic/card references and normalized user flags are stored by the database module.
The verified-entry table is owned by accounts. Moderation is independent of paid
entitlements; PM review must not block combo controls. Admin repair is available
after inspecting Telegram because the Bot API provides no arbitrary message/topic
lookup for recovering an ambiguous create response. Keyword filters, auto replies,
backup destinations and edit synchronization are outside this module.

Selected topic locking/profile/moderation behavior was adapted from the innovation
3.0.5 version of TGbot-D1 using Go and the panel's identity/queue boundaries:
https://github.com/moistrr/TGbot-D1/blob/main/TGbot-D1%E5%88%9B%E6%96%B0%E7%89%88%E6%9C%AC3.0.5.js

Wire contracts: https://core.telegram.org/bots/api

Read state is reference-only: a member can acknowledge one delivered copy explicitly,
or an inbound PM infers earlier successful copies as read. The admin user profile
shows the source and timestamp. Neither path stores message content or claims a
Telegram-native read receipt. Admin `/refund`, `/addtxb`, and `/deducttxb` commands
are accepted only inside that member's PM topic and use existing refund, ledger,
audit, and provider-operation services.

The topic profile's `startapp` button uses the verified bot username and target
user ID. Configure the bot's Main Mini App URL in BotFather; the client consumes
the start parameter only as a route hint, then applies the existing admin session
guard before loading the target profile.

Hosted regressions: `service_test.go` covers entry, supported media, commands,
forum replies and callback identity/card scope. `worker_test_helpers_test.go`
supplies reference-only repositories/senders. `worker_delivery_test.go` covers
execution-time moderation, success replay and definitive topic recovery;
`worker_recovery_test.go` covers ambiguous phases, safe rate-limit retries and
wrong-thread repair rejection. `profile_card_test.go` covers onboarding hiding,
optional coupon details, traffic, rollover, and the Mini App deep link. No local
automated suites are required or run.
