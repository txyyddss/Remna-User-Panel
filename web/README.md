# TX Carpool web client

Vue 3.5 Telegram Mini App for onboarding, subscription management, TXB billing, catalog purchases, and administration. The interface is English-only, mobile-first, and locked to the premium dark TX visual system.

## Commands

```sh
npm install
npm run generate:api
npm run dev
npm run lint
npm run typecheck
npm run test
npm run build
```

`npm run generate:api` reads `../api/openapi.yaml` and writes `src/api/generated.ts`. The typed runtime client uses those generated request contracts and adds focused view models in `src/api/types.ts`.

The production build writes to `../internal/webui/dist`. The Go server embeds that directory into the single application binary.

## Module specification

### Application platform

- `src/main.ts` initializes Telegram, Pinia, Vue Router, and global styles.
- `src/router` lazy-loads route views and enforces onboarding and admin access before navigation.
- `src/stores/session.ts` is the only global store. It exchanges trusted Telegram `initData` for the HttpOnly session and exposes current user state.
- `src/api` contains the same-origin JSON client, generated OpenAPI types, and focused application-facing types.
- `src/utils` contains pure formatting and Telegram bridge helpers. Prices displayed from the catalog always use server-formatted money values.

### Layout and shared UI

- `src/components/layout/AppShell.vue` provides the desktop navigation rail, mobile header, Telegram safe areas, and bottom navigation.
- `src/components/common` provides status, notice, loading, and confirmation primitives. Dialogs use Reka UI and render through the shared overlay target.
- `src/styles/main.css` defines the graphite, charcoal, off-white, and muted-mint design tokens, the 16px panel and 12px control radius rules, responsive layouts, focus states, and reduced-motion fallbacks.

### Onboarding

- `src/composables/useIntroSequence.ts` owns the three exact 900ms welcome messages and skip behavior.
- `src/composables/useOnboarding.ts` owns resumable server state, identity-bound invite links, canonical membership checks, username reservation, and agreement acceptance.
- `src/components/onboarding` contains one focused panel for each step. New members never proceed until both Telegram memberships and the username have been accepted by the server.

### Member dashboard

- `src/composables/useDashboard.ts` loads the account projection, derives the usage ratio, handles cached Remnawave warnings, and rotates subscription credentials.
- `src/components/dashboard` displays the TXB balance, active and queued purchases, usage, top nodes, bearer-safe subscription controls, and visible future-product routes.

### Catalog and purchasing

- `src/composables/useCatalog.ts` owns catalog loading, plan and add-on selection, checkout state, server-authoritative purchasing, and insufficient-balance recovery.
- `src/components/catalog` displays asymmetric combo choices, squad add-ons, and a Reka UI review sheet. The client never submits a calculated purchase total.

### Balance and payments

- `src/composables/useBalance.ts` loads the current ledger-derived balance, server-configured provider availability, and immutable activity.
- `src/composables/usePaymentOrder.ts` creates EZPay, BEPusdt, or Stars orders, generates QR images locally, opens Telegram invoices, and polls durable server state. Redirects and invoice callbacks never credit the UI by themselves.
- `src/components/billing` provides the balance page, payment sheet, exact payable amount, QR view, external link, and recent ledger list.

### Administration

- `src/composables/useAdminSection.ts` provides consistent loading, mutation, error, refresh, and haptic behavior for audited admin resources.
- `src/components/admin/AdminShell.vue` exposes responsive routes for settings, catalog, users, entitlements, payments, backups, and audit events.
- Admin panels support masked setting updates, combo CRUD, editable imported squad merchandising, user balance adjustments, reason-bound entitlement cancellation, provider refunds with immutable reversal records, on-demand backups, failed-job retries, and append-only audit inspection.

### Future products

- `src/components/placeholders` supplies polished, linked placeholders for Games, Questionnaire, and Emby. No Emby integration or schema is included in v1.

## Tests

Vitest covers authentication route decisions, admin authorization, the resumable welcome sequence, integer TXB input, server-formatted catalog values, stale Remnawave data, authoritative payment polling, and reason-bound admin mutation contracts. Tests run in Happy DOM with Telegram and clipboard boundaries mocked.
