# Vue web UI module

## Composition and ownership

The web module is a mobile-first Telegram Mini App built with Vue 3 Composition API and `<script setup lang="ts">`. Route views are composition surfaces; typed feature composables own remote state, polling, and mutations; reusable components own presentation and local interaction state.

The main boundaries are:

- `AppShell`: safe areas, session bootstrap, desktop/mobile navigation, skip link, Telegram BackButton, and focus restoration after navigation.
- Member features: `ActivityPage`, `CouponWalletPanel`, `QuestionnairePage`, `EmbyPage`, `CatalogCheckout`, and `BalancePaymentSheet`.
- Shared controls: `TxbAmountField` converts human major-unit input to integer hundredths, `SwitchField` wraps accessible Reka switches, and `MarkdownContent` renders sanitized Markdown with raw HTML disabled.
- `AdminShell`: lazy section composition grouped as Commerce, Community, Accounts, and System while preserving every `/admin/:section` URL.
- Workflow surfaces: questionnaire CSV import and database record editing use mobile drawers/dialogs with explicit review states rather than exposing raw SQL or unreviewed mutations.

`src/api/generated.ts` is generated from `api/openapi.yaml`. `src/api/features.ts` is a narrow handwritten transport adapter for the new feature screens; it uses the same error envelope, credentials, decimal-string conventions, and canonical server routes.

Localization is embedded from `locales/en.json` and `locales/zh-CN.json` through `src/i18n/generated.ts`. Vite validates identical nested leaf-key structures before a build; runtime `t()` supports placeholder interpolation, persisted switching, and Telegram-language defaulting. Known API error codes map to locale keys before generic fallbacks. The language selector is available before authentication and inside the authenticated shell.

## Routes and state behavior

The canonical member routes are `/home`, `/catalog`, `/balance`, `/activity`, `/questionnaire`, and `/emby`. `/games` redirects to `/activity`. The three expanded product areas are real data-backed routes; they no longer share a placeholder view.

Activity shows daily check-in, group-message reward progress/claimed state, enabled betting games, lucky draws, recent outcomes, exact stakes/fees, and resulting balance. It does not use streak pressure, near-miss animation, celebratory loops, or other manipulative gambling cues. Every financial action is disabled while pending and displays the server result.

Coupon wallet redemption is explicit. Catalog checkout displays at most one selected eligible grant, uses the server quote, prunes add-on squads newly included by a changed combo, and disables included squads with an `Included` label. Combo and squad descriptions pass through `MarkdownContent`; admin editing provides the same renderer as a preview.

Questionnaire participation retrieves the same durable validation code on repeat visits. The administrator import flow progresses through upload, header/sample review, validation-column selection, match analysis, explicit settlement, and background-status polling. Empty, malformed, duplicate, unknown, failed, and completed states remain visible without losing the selected import.

Emby setup collects a write-only password, parental rating, and libraries before debit and shows the exact TXB setup price. Linked accounts expose only approved password and preference controls; raw policy fields are never presented. Failed retryable provisioning shows a bounded retry action.

The payment sheet selects a canonical method ID in two stages: provider, then rail. It separately renders provider URL, QR payload, receiving address, actual crypto amount, and currency when supplied. Pending orders can be cancelled; cancellation stops polling, while the UI still accepts a later authoritative `paid` projection and refreshes the balance exactly once.

## Administration

All TXB form fields accept major-unit decimal strings. For example, entering `150` sends `15000` minor units. Rates, percentages, and multipliers retain their documented server units.

The database editor lists allowlisted application tables, cursor-paginates records, and renders typed null, boolean, numeric, text, and blob controls. A write first requests a server diff/review hash, then requires the optimistic record hash, reason, and typed `EDIT <table>` confirmation. The UI warns that direct edits bypass domain synchronization hooks.

Backup downloads use an authenticated binary response. Restore requires `RESTORE <filename>`, describes the automatic rescue backup, submits a staged restore, polls its operation, and enters reconnect/reauthentication state after the server begins its graceful restart.

Encrypted settings render as masked values and only permit write-only replacement. Financial and destructive actions retain visible text labels even when icons are present.

## Visual and accessibility contract

The visual system preserves the existing premium-dark graphite/mint identity: graphite canvas, layered charcoal surfaces, off-white text, one muted mint accent, 16 px panels, and 12 px controls. Member density remains lower than administrator density, copy is task-oriented, and motion is restrained.

Interactive targets are at least 44 px. Keyboard focus is visible and restored to the main landmark after route changes. The skip link, semantic fieldsets, status live regions, Reka dialog focus trapping, Telegram BackButton integration, safe-area variables, and 320 px layouts are part of the acceptance contract. Faint text meets usable contrast, and `prefers-reduced-motion` disables nonessential transitions.

## Build, failure behavior, and verification

`npm run generate:api` regenerates the OpenAPI types. `npm run typecheck`, `npm run lint`, `npm test -- --run`, and `npm run build` are the local verification sequence. Production assets are written to `internal/webui/dist` for Go embedding.

Feature requests abort or stop polling on unmount, route change, sheet close, or cancellation. Standard loading, empty, stale, validation, upstream-failure, retry, cancelled, and late-settlement states are rendered rather than represented by blank panels. Telegram bootstrap waits for delayed WebApp initialization and accepts the standard WebApp data source; an already-detected WebApp context reports a loading state rather than the misleading outside-Telegram message. A 401 clears session state and restarts Telegram authentication; no plaintext password, callback capability, subscription URL, or encrypted setting is logged by the client.
