# BEPusdt integration

This package implements the narrow BEPusdt transaction and callback surface used by TX Carpool. Provider-mandated MD5 signatures are isolated to the callback protocol, compared in constant time, and never placed in errors.

- `client.go` defines the sanitized API error, injectable HTTP transport, client options, and validated client construction.
- `transactions.go` validates, signs, sends, and verifies transaction creation and cancellation wire contracts.
- `webhooks.go` parses signed and capability-authenticated callbacks and implements the provider signature algorithm.
- `wire.go` decodes BEPusdt's mixed JSON scalar representations and rejects duplicate or nested callback fields.
- `validation.go` validates callback/payment URLs and exact base-10 decimal values.
- `doc.go` supplies the package documentation.
- `bepusdt_test.go` covers signing fixtures, transaction requests, callback authentication, and response validation.
- `diagnostic_test.go` contains the explicitly opt-in, redacted live create/cancel diagnostic.
- `redirect_test.go` verifies provider redirects cannot receive signed request bodies.
- `README.md` documents the package layout and security boundary.
