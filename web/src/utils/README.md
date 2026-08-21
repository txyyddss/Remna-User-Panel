# Utilities

- `format.ts` formats money, dates, bytes, and TXB input.
- `dom.ts` provides WebView-safe focus restoration.
- `dom.test.ts` covers focus restoration behavior.
- `format.test.ts` covers formatting and money conversions.
- `browserCompatibility.ts` checks baseline browser capabilities, supplies secure UUID entropy, and installs narrow WebView constructor fallbacks before the app bundle mounts.
- `browserCompatibility.test.ts` covers constructor, UUID, and fail-closed entropy behavior.
- `bootstrapFallback.ts` renders the locale-owned recovery action when the app
  cannot load far enough for Vue's render boundary to mount.
- `coupons.ts` centralizes client-side coupon visibility checks for expired or exhausted grants.
- `coupons.test.ts` covers coupon visibility boundaries.
- `telegram.ts` wraps Telegram Mini App integration helpers.
- `telegramFullscreen.ts` owns fullscreen request state, event synchronization, and safe-area reserve behavior.
- `telegramHaptics.ts` contains delegated light/heavy click feedback, `data-haptic="selection"` semantic feedback, and distinct payment/bet outcomes re-exported by `telegram.ts`.
- `telegramLinks.ts` opens Telegram-native links with a browser fallback.
- `telegramContext.ts` classifies the native `TelegramWebviewProxy` bridge, launch markers, SDK-captured `WebView.initParams`, platforms, and user agents before route construction.
- `telegram.test.ts` covers Telegram capability behavior and delegated impact/selection disposal.
- `validation.ts` contains shared input validation helpers.
- `validation.test.ts` covers validation boundaries.
- `latestRequest.ts` supplies generation guards so disposed or superseded async
  work cannot overwrite current Vue state or restart timers.
