# Billing components

- `BalancePaymentSheet.vue` drives coupon funding and the provider order lifecycle from Home, shows the redeemed TXB gain, exposes Telegram native back navigation, and handles failed/expired-order reissue.
- `BalancePaymentConfiguration.vue` presents the localized amount, provider-account tiles, channel, and coupon-funding selection step with Nuxt UI controls; its coupon input and action remain separate on narrow screens and provider labels come from each stable-ID profile.
- `PaymentReceiptDetails.vue` renders the browser-safe provider-return receipt with localized amount, payment ID, provider, channel, status, and timing metadata using Nuxt UI surfaces.
- `BalancePaymentSheet.test.ts` verifies provider labels stay distinct from provider-owned payment channels, such as EZPay and Alipay.
