# Layout components

- `AppShell.vue` owns the fullscreen safe-area shell, content focus, and Telegram back navigation. Every viewport uses the same centered phone-width content column and fixed bottom navigation; explicit Vue Router button actions avoid native document loads. The shell combines Telegram/system insets, and fullscreen greetings prefer the app-managed username. Locale selection lives in the Home footer.
- Route focus restoration and Telegram BackButton callbacks are guarded against WebView teardown and rejected promises. The native BackButton uses a shared owner stack so an open payment sheet closes before route history changes.
- `LanguageControl.vue` provides the compact language selector and its accessible locale popover. Home exposes it in the page footer, while auth and onboarding keep their own localized entry points.
- `AppShell.test.ts` verifies focus restoration, Telegram BackButton behavior, and the admin mobile navigation entry.

Route and native Back actions use soft navigation feedback; locale feedback is emitted only after the locale actually changes.
