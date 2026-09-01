# Layout components

- `AppShell.vue` owns the fullscreen safe-area shell, content focus, Telegram back navigation, and the resizable Nuxt UI `UDashboardSidebar` desktop shell. Phones keep their centered fixed bottom navigation, while desktop exposes the complete member and administrator route hierarchy without a fold control or product wordmark.
- Route focus restoration and Telegram BackButton callbacks are guarded against WebView teardown and rejected promises. The native BackButton uses a shared owner stack so an open payment sheet closes before route history changes.
- `LanguageControl.vue` provides the compact `ULocaleSelect` language control. Home exposes it on mobile, while the desktop rail, auth, and onboarding keep localized entry points in their own context.
- `navigation.ts` owns the localized mobile and hierarchical desktop sidebar item definitions, including the administrator-only section list.
- `AppShell.test.ts` verifies focus restoration, Telegram BackButton behavior, and the admin mobile navigation entry.

Route and native Back actions use soft navigation feedback; locale feedback is emitted only after the locale actually changes.
