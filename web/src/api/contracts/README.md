# Frontend API contracts

- `responseSchemas.json` contains generated response shapes from OpenAPI. `responses.ts` lazily builds Zod validators for cached and refreshed reads, accepts new optional fields, and rejects malformed required data before it reaches Vue state. `responses.test.ts` covers additive changes and malformed successful responses.

- `activity.ts` defines activity and reward responses.
- `community.ts` defines coupon and questionnaire responses.
- `commerce.ts` defines Emby and payment responses, including discovered crypto currency/network metadata and pending crypto orders.
- `admin.ts` defines database, statistics, onboarding, and restore responses.
- `affiliates.ts` defines member referral projections and versioned tier editor contracts.
- `compensation.ts` defines generated aliases for outage configuration, events, pages, and review writes.
- `abuse.ts` exports the privacy-safe detector contract aliases.

Endpoint methods live in `../features.ts`; generated core schemas live in
`../generated.ts` and are regenerated from the application OpenAPI document.
