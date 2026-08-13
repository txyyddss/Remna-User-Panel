# Layout components

- `AppShell.vue` owns the fullscreen safe-area shell, member navigation, content focus, and Telegram back navigation. The shared shell reserves a minimum mobile top boundary and combines Telegram/system top insets so native header controls cannot cover page content. Mobile primary navigation uses explicit Vue Router button actions without a native `href`, preventing Telegram WebView from treating internal navigation as a document load. The member navigation contains only route destinations; locale selection lives in the Home footer.
- Route focus restoration and Telegram BackButton callbacks are guarded against WebView teardown and rejected promises. The native BackButton uses a shared owner stack so an open payment sheet closes before route history changes.
- `LanguageControl.vue` provides the compact language selector and its accessible locale popover. Home exposes it in the page footer, while auth and onboarding keep their own localized entry points.
- `AppShell.test.ts` verifies focus restoration, Telegram BackButton behavior, and the admin mobile navigation entry.
