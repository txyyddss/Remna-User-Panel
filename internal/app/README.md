# Application composition
- `adapters_part2.go` continues the focused implementation from its original package module.

The `app` package is the process composition root. It opens SQLite, constructs
domain services and provider adapters, registers durable outbox handlers, and
owns HTTP, scheduler, backup, and provider-queue lifecycles.

Remnawave and Emby clients are intentionally created only inside their adapter
queue callbacks. Reads and writes therefore share bounded per-provider pacing;
durable outbox handlers keep their existing retry semantics and use the same
adapters for the final network attempt.

`Application.Run` starts provider workers before accepting HTTP traffic and
waits for the scheduler and provider workers to finish before returning.
`Application.Close` checkpoints and closes SQLite only after that runtime has
stopped.

- `app.go` composes configuration, persistence, domain services, adapters, workers, and HTTP delivery.
- `application.go` defines the process-owned application resource container.
- `member_workflows.go` composes member connection/reset/refund services and registers legacy drop plus block/unblock operation workers on one dispatcher.
- `outbox_composition.go` registers core jobs, user/payment/affiliate Telegram notifications, the scheduled IP-unblock backstop, and the shared provider-operation dispatcher.
- `admin_user_outbox.go` schedules due manual temporary-ban restoration through the shared provider-operation lane.
- `payment_operations.go` registers durable payment create and cancellation handlers on that dispatcher.
- `payment_profile_manager.go` probes BEPUSDT profiles at startup and after saves, publishes process-local discovered rails, and reports disabled profiles to every Telegram administrator.
- `mutation_operation_composition.go` registers subscription, Emby, questionnaire, retry, and refund command handlers.
- `payment_refund_adapter.go` reconciles ambiguous Telegram Stars refunds from authoritative transaction history.
- `application_lifecycle.go` starts and stops provider queues, the scheduler, HTTP serving, and database resources in dependency order.
- Startup upload reconciliation receives an independent ten-minute context so a large verified candidate is not truncated by the ordinary bootstrap deadline.
- `bootstrap_settings.go` validates and persists the encrypted provider settings required at first startup.
- `provider_queues.go` configures, starts, and shuts down the independent Remnawave and Emby admission queues.
- `adapters.go` implements Telegram identity/membership and payment-provider bridges used by domain services.
- `remna_adapter.go` owns queued Remnawave client creation, shared call helpers, user-ID validation, and domain mapping.
- `remna_admin_user_adapter.go` verifies replacement identities and queues documented account disable and enable actions.
- `remna_accounts_adapter.go` implements queued Remnawave account lookup and creation operations.
- `remna_catalog_adapter.go` implements queued dashboard reads and subscription revocation for catalog workflows.
- `remna_usage_adapter.go` implements queued date-bounded per-node usage reads for the member dashboard.
- `remna_admin_adapter.go` implements queued live squad and node projections for catalog composition; it exposes no node-assignment mutation.
- `node_multiplier_cache.go` owns the copied five-minute node-multiplier cache shared by queued Remnawave projections and rollover usage mapping.
- `node_multiplier_cache_test.go` covers cache expiry at the five-minute boundary.
- `remna_entitlements_adapter.go` implements queued entitlement, traffic reset, removal, and rollover operations.
- `remna_notifications_adapter.go` maps documented queued user-stream pages to
  the narrow traffic-threshold projection.
- `remna_member_operations_adapter.go` implements queued connection scans, plugin block/unblock execution, disconnect reconciliation, usage, quiesce, and restore calls.
- `remna_statistics_adapter.go` implements queue-backed Remnawave digest, node,
  traffic, Geocheck, and host operations for product statistics; documented fractional
  live counts and byte rates are rounded into the integer public contract.
- `remna_compensation_adapter.go` maps complete queue-backed node and squad snapshots without persisting authoritative assignments.
- `remna_abuse_adapter.go` supplies queue-backed node identity to detector key provisioning.
- `remna_abuse_actions.go` keeps every detector-triggered Remnawave call in the provider queue.
- `abuse_ip_ban.go` persists and resumes asynchronous Remnawave connection scans before applying an IP ban.
- `abuse_outbox.go` consumes durable detector punishment, restoration, and Telegram delivery jobs.
- `abuse_scheduler.go` performs startup catch-up and UTC-aligned `:00`/`:30` durable abuse processing.
- `abuse_outbox_test.go` covers MarkdownV2 escaping for detector notifications.
- `remna_statistics_adapter_test.go` covers bounded rounding of Remnawave live numeric fields.
- `emby_adapter.go` implements queued Emby client creation, account operations, policy updates, and metadata lookups.
- `scheduler.go` runs prompt temporary-ban restoration, automatic due-renewal revalidation before entitlement transitions, plus recurring outbox, rollover, backup, and maintenance work until application cancellation.
- `notification_scheduler.go` runs the bounded startup and five-minute reminder
  and traffic scan independently from the outbox drain.
- `statistics_scheduler.go` schedules the 30-minute statistics refresh, startup
  and six-hour node Geocheck cache work, node-compensation observations, and queued host-multiplier reconciliation.
- `statistics_setup.go` composes the statistics service and its provider adapter.
- The scheduler delegates the daily backup-gated cleanup order to
  `internal/maintenance`; the durable local-date lease prevents overlapping
  runs across process instances.
- `telegram_scheduler.go` configures Telegram delivery and reconciles Telegram Stars transactions.
- `telegram_queue_adapter.go` queues bot identity, membership, setup, and message operations.
- `telegram_queue_payments.go` queues Telegram Stars invoice, query, history, and refund operations.
- `telegram_command_registration.go` installs localized private-chat and configured-group command menus.
- `app_test.go` verifies normalized Telegram Stars transaction directions.
- `adapter_queue_test.go` verifies Remnawave and Emby adapter calls enter their queue before client construction.
- `remna_entitlements_adapter_test.go` verifies active expiry propagation and the disabled-user far-future expiry.
- `remna_entitlements_rollover_test.go` verifies that rollover usage keeps Remnawave's inclusive final date.
- `provider_queues_test.go` verifies provider queue configuration and independent worker operation.
- `adapters_part2.go` contains the remaining payment-provider verification and callback mapping helpers.
