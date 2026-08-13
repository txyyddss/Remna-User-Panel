# Shared models

- `model.go` documents the dependency-free shared model package.
- `identity.go` defines Telegram identity and local user representations.
- `commerce.go` defines catalog, typed squad-profile projections, purchase, entitlement, renewal, quote, rollover, and money payloads.
- `finance.go` defines ledger, payment, masked payment-profile, refund, and terminal-payment courtesy-credit payloads plus canonical TXB formatting.
- `analytics.go` defines member traffic, date-bounded per-node usage, and administrator statistics payloads.
- `operations.go` defines backup, audit, outbox, and safe setting views.
