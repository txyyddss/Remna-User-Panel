# Telegram integration

This package contains the Telegram Bot API and Mini App authentication surface used by TX Carpool. Bot tokens remain confined to request URLs inside the transport, while returned and transport errors are sanitized before propagation.

- `client.go` defines the sanitized API error, client options, injected transport, and validated client construction.
- `transport.go` performs bounded Bot API calls, decodes envelopes, sanitizes failures, and validates URLs and secrets.
- `webhooks.go` configures webhooks and the Mini App menu button and verifies webhook secrets in constant time.
- `bot.go` registers localized command menus and sends bounded plain-text or MarkdownV2 replies after decoding Telegram's message result.
- `memberships.go` manages join-request invites, approvals, revocation, and canonical chat membership lookup.
- `stars.go` creates Stars invoice links, answers pre-checkout queries, lists transactions, and requests refunds.
- `types.go` defines Telegram user, chat, update, membership, invite, payment, and Stars wire contracts.
- `initdata.go` verifies signed Telegram Mini App init data with freshness and replay-resistant field handling.
- `doc.go` supplies the package documentation.
- `client_test.go` covers Bot API wire methods, sanitized errors, update decoding, and webhook verification.
- `bot_test.go` verifies MarkdownV2 send-message payloads and successful Message result decoding.
- `redirect_test.go` verifies provider redirects cannot receive the bot token or request body.
- `initdata_test.go` covers init-data signatures, time bounds, token validation, and optional Telegram signature fields.
- `README.md` documents the package layout and security boundary.
