# Router

- `index.ts` declares routes and installs the session guard.
- `history.ts` selects WebView-safe in-memory history or browser history.
- `recovery.ts` performs one-shot recovery for stale lazy route chunks.
- `guards.ts` resolves onboarding, administrator, and member access.
- `guards.test.ts` covers protected-route decisions.
- `history.test.ts` covers Telegram and browser history selection.
- `recovery.test.ts` covers one-shot lazy-route recovery.

Telegram sessions use Vue Router's in-memory history so internal links do not
reload the Mini App at a new WebView URL and lose Telegram's launch context.
Regular browser sessions use HTML5 history and keep shareable route URLs.
If a cached WebView bundle cannot load a lazy route chunk, the router retries
once after reloading and preserves the intended route in session storage.

`/payment-result` is an immersive provider-return landing route. Telegram opens it with owner-scoped durable order polling; a provider redirect opened in a regular browser uses the short-lived capability-return status endpoint and never boots the protected app shell. Failed or expired Telegram orders hand their ID to Home for an owner-scoped reissue fetch.

`/balance` remains a compatibility redirect to `/home?topUp=1`; Home owns the real funding sheet.
