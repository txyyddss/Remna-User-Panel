# Shared models

- `model.go` documents the dependency-free shared model package.
- `identity.go` defines Telegram identity and local user representations.
- `commerce.go` defines catalog, typed squad-profile projections, purchase, entitlement, legacy renewal, quote, rollover, and money payloads.
- `commerce_renewals.go` retains legacy internal renewal transport types separately from the automatic-renewal contract.
- `automatic_renewal.go` defines the owner-facing automatic-renewal state, current one-cycle price, and dashboard failure notice.
- `finance.go` defines ledger, payment, masked payment-profile, refund, and terminal-payment courtesy-credit payloads plus canonical TXB formatting.
- `analytics.go` defines member traffic, date-bounded per-node usage, and administrator statistics payloads.
- `operations.go` defines backup, audit, outbox, and safe setting views.
