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

- `app.go` defines the application container and composes configuration, persistence, domain services, adapters, workers, and HTTP delivery.
- `application_lifecycle.go` starts and stops provider queues, the scheduler, HTTP serving, and database resources in dependency order.
- `bootstrap_settings.go` validates and persists the encrypted provider settings required at first startup.
- `provider_queues.go` configures, starts, and shuts down the independent Remnawave and Emby admission queues.
- `adapters.go` implements Telegram identity/membership and payment-provider bridges used by domain services.
- `remna_adapter.go` owns queued Remnawave client creation, shared call helpers, user-ID validation, and domain mapping.
- `remna_accounts_adapter.go` implements queued Remnawave account lookup and creation operations.
- `remna_catalog_adapter.go` implements queued dashboard reads and subscription revocation for catalog workflows.
- `remna_usage_adapter.go` implements queued date-bounded per-node usage reads for the member dashboard.
- `remna_admin_adapter.go` implements queued live squad and node projections for catalog composition; it exposes no node-assignment mutation.
- `remna_entitlements_adapter.go` implements queued entitlement, traffic reset, removal, and rollover operations.
- `emby_adapter.go` implements queued Emby client creation, account operations, policy updates, and metadata lookups.
- `scheduler.go` runs automatic due-renewal revalidation before entitlement transitions, plus recurring outbox, rollover, backup, and maintenance work until application cancellation.
- `telegram_scheduler.go` configures Telegram delivery and reconciles Telegram Stars transactions.
- `app_test.go` verifies normalized Telegram Stars transaction directions.
- `adapter_queue_test.go` verifies Remnawave and Emby adapter calls enter their queue before client construction.
- `remna_entitlements_adapter_test.go` verifies active expiry propagation and the disabled-user far-future expiry.
- `provider_queues_test.go` verifies provider queue configuration and independent worker operation.
- `adapters_part2.go` contains the remaining payment-provider verification and callback mapping helpers.
