# Frontend API contracts

- `activity.ts` defines activity and reward responses.
- `community.ts` defines coupon and questionnaire responses.
- `commerce.ts` defines Emby and payment responses, including discovered crypto currency/network metadata and pending crypto orders.
- `admin.ts` defines database, statistics, onboarding, and restore responses.
- `affiliates.ts` defines member referral projections and versioned tier editor contracts.
- `compensation.ts` defines generated aliases for outage configuration, events, pages, and review writes.
- `abuse.ts` exports the privacy-safe detector contract aliases.

Endpoint methods live in `../features.ts`; generated core schemas live in
`../generated.ts` and are regenerated from the application OpenAPI document.
