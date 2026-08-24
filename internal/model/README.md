# Shared models

- `purchase_addons.go` defines the server-priced active-term squad-addition quote payload.

- `model.go` documents the dependency-free shared model package.
- `identity.go` defines Telegram identity and local user representations, including the private account-wide automatic-reset flag.
- `commerce.go` defines catalog products with required live accessible-node groups, typed squad-profile projections, purchase, entitlement, legacy renewal, quote, rollover, and money payloads.
- `commerce_foundation.go` defines the global Add TXB bounds returned to members and administrators.
- `commerce_renewals.go` retains legacy internal renewal transport types separately from the automatic-renewal contract.
- `automatic_renewal.go` defines the owner-facing automatic-renewal state, current one-cycle price, and dashboard failure notice.
- `finance.go` defines ledger, payment, masked payment-profile, refund, and terminal-payment courtesy-credit payloads plus canonical TXB formatting.
- `analytics.go` defines member traffic, date-bounded per-node usage, and administrator statistics payloads.
- `operations.go` defines backup, audit, outbox, and safe setting views.
- `provider_operations.go` defines durable provider-operation receipts and item states shared across workflows.
- `statistics_dashboard.go` defines the persisted member-facing statistics dashboard projections.
- `traffic_reset_automation.go` defines the public account-wide automatic-reset preference projection.
