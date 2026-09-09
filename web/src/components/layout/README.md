# Layout components

- `AppShell.vue` is the single route-motion entrance. It uses keyed `AnimatePresence` with 14px local offsets, no initial hydration animation, and a reduced-motion opacity-only path.
- `MobileNavigation.vue` and `AdminSectionNavigation.vue` own shared `layoutId` active markers and short icon/selection feedback. Nuxt UI remains the control system for desktop navigation and buttons.
- `usePageTransition.ts` computes forward/backward direction from navigation order and tracks the wide viewport axis without blocking router resolution on animation callbacks.

- `AppShell.vue` owns the fullscreen safe-area shell, route-change scroll reset, content focus, Telegram back navigation, and the resizable Nuxt UI `UDashboardSidebar` desktop shell. Phones keep their centered fixed bottom navigation, while desktop exposes the complete member and administrator route hierarchy without a fold control or product wordmark; its Home parent is a real route link while its quick actions remain expanded children.
- Route focus restoration and Telegram BackButton callbacks are guarded against WebView teardown and rejected promises. The native BackButton uses a shared owner stack so an open payment sheet closes before route history changes.
- `LanguageControl.vue` provides the compact `ULocaleSelect` language control. Home exposes it on mobile, while the desktop rail, auth, and onboarding keep localized entry points in their own context.
- `MobileNavigation.vue` maps confirmed routes to Nuxt UI bottom tabs with localized labels, keyboard navigation, safe-area spacing, and the administrator entry.
- `SidebarMember.vue` displays the current Telegram user's photo with `UAvatar`, a localized greeting, and a truncated Telegram username. Photos are display-only SDK data, matched to the authenticated Telegram ID, and are never persisted. Missing/private photos use the avatar's initials fallback. Desktop omits the duplicate fullscreen greeting; the resizable sidebar starts at its compact 13rem minimum.
- `navigation.ts` owns the localized mobile and hierarchical desktop sidebar item definitions, including the administrator-only section list.
- `usePageTransition.ts` coordinates Vue `Transition` and CSS animations with router completion, including Telegram WebViews without the View Transitions API. Navigation order determines forward/backward slides; rapid navigation waits for the current slide. Outgoing panels become inert and keep their scroll position, while the incoming panel receives focus. Query-only updates do not slide; reduced motion completes immediately.
- `AppShell.test.ts` verifies focus restoration, Telegram BackButton behavior, and the admin mobile navigation entry.

Route and native Back actions use soft navigation feedback; locale feedback is emitted only after the locale actually changes.
