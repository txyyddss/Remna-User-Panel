# TX Carpool OpenAPI contract

- `openapi.yaml` is the public OpenAPI 3.1 entry point. It registers all API paths, reusable components, and the authenticated browser request-signing requirement.
- `paths/` contains one bounded Path Item Object per API URL and an index mapping filenames to routes.
- `components/security-schemes.yaml` defines the session cookie plus timestamp, nonce, and HMAC signature schemes.
- `components/parameters.yaml` contains reusable path, query, and idempotency parameters.
- `components/responses.yaml` contains reusable API error responses.
- `components/schemas/` contains bounded definitions grouped into shards with a schema index.
- Squad profile schemas in `schemas-02.yaml` describe local Broadband, China Optimized, and International Network metadata; they do not alter the Remnawave upstream contract.

Generate TypeScript types from the root entry point with `npm run generate:api` in `web/`. Authenticated operations inherit all four AND-combined security schemes. Telegram authentication, provider callbacks, payment returns, operational probes, and static delivery retain their explicit unsigned protocols.
