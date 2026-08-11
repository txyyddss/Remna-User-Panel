# Layout components

- `AppShell.vue` owns the fullscreen safe-area shell, member navigation, content focus, and Telegram back navigation. Mobile primary navigation uses explicit Vue Router button actions without a native `href`, preventing Telegram WebView from treating internal navigation as a document load.
- Route focus restoration and Telegram BackButton callbacks are guarded against WebView teardown and rejected promises.
- `LanguageControl.vue` provides the compact language selector and its accessible locale popover; the mobile control is positioned independently of the three primary bottom-navigation items.
- `AppShell.test.ts` verifies focus restoration and Telegram BackButton behavior.
