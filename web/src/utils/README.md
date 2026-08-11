# Utilities

- `format.ts` formats money, dates, bytes, and TXB input.
- `format.test.ts` covers formatting and money conversions.
- `browserCompatibility.ts` checks baseline browser capabilities, supplies secure UUID entropy, and installs narrow WebView constructor fallbacks before the app bundle mounts.
- `browserCompatibility.test.ts` covers constructor, UUID, and fail-closed entropy behavior.
- `telegram.ts` wraps Telegram Mini App integration helpers.
- `telegram.test.ts` covers Telegram capability behavior.
- `validation.ts` contains shared input validation helpers.
- `validation.test.ts` covers validation boundaries.
