# Utilities

- `format.ts` formats money, dates, bytes, and TXB input.
- `dom.ts` provides WebView-safe focus restoration.
- `dom.test.ts` covers focus restoration behavior.
- `format.test.ts` covers formatting and money conversions.
- `browserCompatibility.ts` checks baseline browser capabilities, supplies secure UUID entropy, and installs narrow WebView constructor fallbacks before the app bundle mounts.
- `browserCompatibility.test.ts` covers constructor, UUID, and fail-closed entropy behavior.
- `telegram.ts` wraps Telegram Mini App integration helpers.
- `telegramContext.ts` classifies Telegram launch markers, platforms, and user agents before route construction.
- `telegram.test.ts` covers Telegram capability behavior.
- `validation.ts` contains shared input validation helpers.
- `validation.test.ts` covers validation boundaries.
