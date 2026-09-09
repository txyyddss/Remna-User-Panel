# Layout components

- `AppShell.vue` owns the fullscreen safe-area shell, route-change scroll reset, content focus, Telegram back navigation, and the resizable Nuxt UI `UDashboardSidebar` desktop shell. Phones keep their centered fixed bottom navigation, while desktop exposes the complete member and administrator route hierarchy without a fold control or product wordmark; its Home parent is a real route link while its quick actions remain expanded children.
- Route focus restoration and Telegram BackButton callbacks are guarded against WebView teardown and rejected promises. The native BackButton uses a shared owner stack so an open payment sheet closes before route history changes.
- `LanguageControl.vue` provides the compact `ULocaleSelect` language control. Home exposes it on mobile, while the desktop rail, auth, and onboarding keep localized entry points in their own context.
- `MobileNavigation.vue` maps confirmed routes to Nuxt UI bottom tabs with localized labels, keyboard navigation, safe-area spacing, and the administrator entry.
- `SidebarMember.vue` displays the current Telegram user's photo with `UAvatar`, a localized greeting, and a truncated Telegram username. Photos are display-only SDK data, matched to the authenticated Telegram ID, and are never persisted. Missing/private photos use the avatar's initials fallback. Desktop omits the duplicate fullscreen greeting; the resizable sidebar starts at its compact 13rem minimum.
- `navigation.ts` owns the localized mobile and hierarchical desktop sidebar item definitions, including the administrator-only section list.
- `usePageTransition.ts` coordinates native viewport snapshots with router completion. Navigation order determines forward/backward slides; rapid navigation waits for the current slide, and teardown/failures release pending snapshots. Query-only updates, reduced motion, and browsers without View Transitions navigate immediately.
- `AppShell.test.ts` verifies focus restoration, Telegram BackButton behavior, and the admin mobile navigation entry.

Route and native Back actions use soft navigation feedback; locale feedback is emitted only after the locale actually changes.
