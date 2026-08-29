# Layout components

- `AppShell.vue` owns the fullscreen safe-area shell, content focus, and Telegram back navigation. Phones use a centered content column with fixed bottom navigation; desktop viewports expand route-owned widths beside a sticky navigation rail that carries the app link, signed-in member context, and locale control. The shell combines Telegram/system insets, and fullscreen greetings prefer the app-managed username.
- Route focus restoration and Telegram BackButton callbacks are guarded against WebView teardown and rejected promises. The native BackButton uses a shared owner stack so an open payment sheet closes before route history changes.
- `LanguageControl.vue` provides the compact language selector and its accessible locale popover. Home exposes it on mobile, while the desktop rail, auth, and onboarding keep localized entry points in their own context.
- `AppShell.test.ts` verifies focus restoration, Telegram BackButton behavior, and the admin mobile navigation entry.

Route and native Back actions use soft navigation feedback; locale feedback is emitted only after the locale actually changes.
