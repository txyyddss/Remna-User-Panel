# Vue web UI module

## Composition and ownership

The web module is a mobile-first Telegram Mini App built with Vue 3 Composition API and `<script setup lang="ts">`. Route views are composition surfaces; typed feature composables own remote state, polling, and mutations; reusable components own presentation and local interaction state.

The main boundaries are:

- `AppShell`: safe areas, session bootstrap, desktop/mobile navigation, skip link, Telegram BackButton, and guarded focus restoration after navigation.
- Member features: `ActivityPage`, `CouponWalletPanel`, `QuestionnairePage`, `EmbyPage`, `CatalogCheckout`, and `BalancePaymentSheet`.
- Shared controls: `TxbAmountField` converts human major-unit input to integer hundredths, Nuxt UI owns accessible switches/forms/overlays/tables, and `MarkdownContent` renders sanitized Markdown with raw HTML disabled.
- `AdminShell`: compact responsive navigation for Commerce, Community, Onboarding, Users, Emby, and System operations. Every registered backend administration domain has a live UI route.
- Workflow surfaces: questionnaire CSV import and database record editing use mobile drawers/dialogs with explicit review states rather than exposing raw SQL or unreviewed mutations.

`src/api/generated.ts` is generated from the split `api/openapi.yaml` contract. `src/api/http.ts` is the only HTTP transport and `src/api/features.ts` is a narrow feature adapter. Once Telegram authentication supplies the companion key, every protected request signs its uppercase method, escaped path and raw query, exact timestamp/nonce, and SHA-256 body hash. JSON, multipart, binary download, and empty-body calls all share that implementation.

Localization is embedded from matching domain shards under `locales/en/` and `locales/zh-CN/` through `src/i18n/generated.ts`. Vite validates identical nested leaf keys and placeholders before a build; runtime `t()` supports interpolation, persisted switching, and Telegram-language defaulting. API and local failure states resolve to locale keys rather than displaying raw transport text. The language selector is available before authentication and inside the authenticated shell.

## Routes and state behavior

The canonical member routes are `/home`, `/catalog`, `/activity`, `/questionnaire`, `/emby`, and `/payment-result`. `/balance` is retained only as a compatibility redirect to Home's funding sheet, and `/games` redirects to `/activity`. Telegram route mode is selected before Vue Router construction from the native `TelegramWebviewProxy` bridge, launch parameters, WebApp data, platform, or Telegram user-agent signals; Telegram uses in-memory route history and strips launch-only parameters from the initial route, while regular browsers use HTML5 history. Failed lazy route imports receive one guarded reload attempt, while the server revalidates `index.html` and gives hashed assets immutable caching. The provider-return route is browser-public only for the signed return landing flow: Telegram verifies the owner-scoped durable order, while a regular browser receives only the capability-limited status projection. Only failed or expired Telegram orders may hand their ID to Home, which refetches it before presenting a replacement payment. The three expanded product areas are real data-backed routes; they no longer share a placeholder view.

Home's traffic detail opens on demand, defaults to the latest seven inclusive UTC dates, lets the member choose a start and end date within 31 dates, and displays Remnawave's top 20 returned nodes with daily byte totals. The date-range request is independent of the cached home summary and preserves an explicit upstream-unavailable state.

Activity shows daily check-in, group-message reward progress/claimed state, enabled betting games, lucky draws, recent outcomes, exact stakes/fees, and resulting balance. It does not use streak pressure, near-miss animation, celebratory loops, or other manipulative gambling cues. Every financial action is disabled while pending and displays the server result.

Coupon wallet redemption is explicit. Catalog checkout displays at most one selected eligible grant, uses the non-mutating server quote, shows the authoritative effective date before confirmation, prunes add-on squads newly included by a changed combo, and disables included squads with an `Included` label. Combo and squad descriptions pass through `MarkdownContent`; the safe `[text]{color=accent size=lg}` directive maps only allowlisted colors and sizes to CSS classes with raw HTML disabled. Admin editing provides toolbar controls and the same renderer as a live preview.

Questionnaire participation retrieves the same durable validation code on repeat visits. The administrator import flow progresses through upload, header/sample review, validation-column selection, match analysis, explicit settlement, and background-status polling. Closing and destructive deletion are separate actions. Activity result dialogs are focus-trapped, game icons come from the allowlisted external Iconify Phosphor set, and statistics expose accessible table data alongside daily/weekly graphs.

Emby setup collects a write-only password, parental rating, and libraries before debit and shows the exact TXB setup price. Linked accounts expose only approved password and preference controls; raw policy fields are never presented. Failed retryable provisioning shows a bounded retry action.

The payment sheet selects a canonical method ID in two stages: provider, then rail. It separately renders provider URL, QR payload, receiving address, actual crypto amount, and currency when supplied. Pending orders can be cancelled; cancellation stops polling, while the UI still accepts a later authoritative `paid` projection and refreshes the balance exactly once.

## Administration

All TXB form fields accept major-unit decimal strings. For example, entering `150` sends `15000` minor units. Rates, percentages, and multipliers retain their documented server units.

The database editor lists allowlisted application tables, debounces broad search, supports up to five typed filters, cursor-paginates fingerprinted queries, and renders typed null, boolean, numeric, text, and blob controls. A write first requests a server diff/review hash, then requires the optimistic record hash, reason, and typed `EDIT <table>` confirmation. The UI warns that direct edits bypass domain synchronization hooks.

Backup downloads use an authenticated binary response. Restore requires `RESTORE <filename>`, describes the automatic rescue backup, submits a staged restore, polls its operation, and enters reconnect/reauthentication state after the server begins its graceful restart.

Encrypted settings render as masked values and only permit write-only replacement. Financial and destructive actions retain visible text labels even when icons are present.

## Visual and accessibility contract

The visual system preserves the existing premium-dark graphite/mint identity: graphite canvas, layered charcoal surfaces, off-white text, one muted mint accent, 16 px panels, and 12 px controls. Member density remains lower than administrator density, copy is task-oriented, and motion is restrained.

Interactive targets are at least 44 px. Keyboard focus is visible and restored to the main landmark after route changes. The skip link, semantic fieldsets, status live regions, Nuxt UI overlay focus management, Telegram BackButton/MainButton integration, safe-area variables, and 320 px layouts are part of the acceptance contract. Faint text meets usable contrast, and `prefers-reduced-motion` disables nonessential transitions.

## Build, failure behavior, and verification

`npm run generate:api` regenerates the OpenAPI types. `npm run audit:structure` enforces line limits, README inventories, locale parity, the no-local-icon/native-control policy, and Vite client-bundle coverage for production-source icons. `npm run check` then lints, type-checks, runs Vitest, and builds. Production assets are written to `internal/webui/dist` for Go embedding.

Feature requests abort or stop polling on unmount, route change, sheet close, or cancellation. Standard loading, empty, stale, validation, upstream-failure, retry, cancelled, and late-settlement states are rendered rather than represented by blank panels. Telegram bootstrap initializes before router construction, waits for delayed WebApp initialization for authentication, and accepts the standard WebApp data source; an already-detected WebApp context reports a loading state rather than the misleading outside-Telegram message. Internal controls use Vue Router actions rather than native anchor navigation. A 401 clears session state and restarts Telegram authentication; no plaintext password, callback capability, subscription URL, or encrypted setting is logged by the client.
