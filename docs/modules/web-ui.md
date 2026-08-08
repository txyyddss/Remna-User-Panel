# Vue web UI module

## Composition and ownership

The web module is a mobile-first Telegram Mini App built with Vue 3 Composition API and `<script setup lang="ts">`. Route views are composition surfaces. API state, polling, Telegram bridge effects, and business interactions live in typed feature composables; presentational components receive readonly props and emit typed events.

Primary boundaries are:

- `AppShell`: safe areas, session bootstrap, navigation, global notices, and route outlet.
- `OnboardingFlow`: timed introduction and resumable membership, username, and agreement steps.
- `UserHome`: balance, entitlement/renewal, usage, top nodes, subscription copy/open/revoke actions.
- `CatalogCheckout`: server catalog selection and purchase submission; browser totals are explanatory only.
- `BalancePaymentSheet`: top-up amount, provider choice, exact payable/QR/link, and durable order polling.
- `AdminShell`: guarded configuration and domain-operation views.

Pinia stores only session-wide identity/auth state and small navigation concerns. Feature composables own remote state. Pure formatting functions stay utilities. Props flow down, events flow up, and server responses remain the only source of payment, balance, price, and entitlement truth.

## Navigation and behavior

Unauthenticated startup posts raw `Telegram.WebApp.initData`; browser mock identity is not supported in production. Route guards send incomplete users to their persisted onboarding step, complete users to Home, and the exact administrator to admin setup when requested. Games, Questionnaire, and Emby are accessible polished “Coming soon” routes without placeholder network clients.

New users see “Hi, how are you?”, “Welcome to TX Carpool”, and “Just take you several seconds to complete” for 900ms each with a skip control. One primary action is shown per onboarding step. Joining links open through Telegram; “Already joined” calls the server's canonical check. Username feedback renders server field/conflict errors and never assumes a client preflight reserved the value.

The payment sheet renders the server's exact payable amount and QR payload and uses `Telegram.WebApp.openInvoice` for Stars. It polls with cancellation/backoff while visible and closes only when the GET order response is `paid`. Redirect, invoice-close, and app-resume signals trigger a refetch, not optimistic credit.

Subscription revoke requires confirmation. URLs are copied/opened only through explicit user actions and are never placed in analytics or error telemetry. Stale Remnawave statistics remain visible with an inline warning while independent dashboard sections continue working.

An incomplete administrator enters the admin dashboard by default: root and user-product routes redirect to `/admin/settings`, while admin sections remain available. The dashboard provides an explicit entry to the standard signup flow. This does not provision a Remnawave identity until signup completes; a completed administrator can use both admin and user-side interfaces.

## Visual and accessibility contract

The visual system is English-only premium dark: graphite canvas, layered charcoal surfaces, off-white text, and one muted mint accent. Panels use 16px radii, controls 12px, and pills only communicate semantic status. Telegram top/bottom safe-area variables are honored. Interactive targets are at least 44px, focus is visible, color contrast meets WCAG AA, and dialogs use accessible Reka UI focus/keyboard behavior.

Motion lasts roughly 180–240ms except the specified 900ms intro cadence. `prefers-reduced-motion` disables nonessential transitions and turns staged onboarding into immediate readable state. Loading, empty, failure, stale, offline, and success states are designed rather than represented by blank panels.

## Build, failure behavior, and verification

`web` uses npm. `npm run generate:api` creates `src/api/generated.ts` from `../api/openapi.yaml`; generated code is not hand-edited. `npm run build` first type-checks and writes production assets to `../internal/webui/dist` for Go embedding.

The API client sends same-origin credentials, parses the standard error envelope, and attaches a request ID to operator-facing diagnostics. A 401 clears session state and reauthenticates from current Telegram init data; onboarding conflict responses route to the appropriate safe surface. Pollers and requests abort on unmount, route change, or sheet close.

Vitest and Vue Test Utils cover auth/onboarding guards, intro skip/reduced motion, resume at each onboarding state, server-owned catalog totals, purchase conflict/debt states, payment polling and cleanup, callback-return refetch, stale statistics, subscription confirmation, admin authorization, placeholders, and loading/empty/error accessibility. ESLint, `vue-tsc --noEmit`, dependency audit, and production build are release gates.
