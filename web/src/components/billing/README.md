# Billing components

- `BalancePaymentSheet.vue` drives coupon funding and the provider order lifecycle from Home, including a verified failed/expired-order reissue that creates a replacement order.
- `BalancePaymentConfiguration.vue` presents the localized amount, provider, channel, and coupon-funding selection step; its coupon input and action remain separate on narrow screens and provider labels come from the provider profile.
- `BalancePaymentSheet.test.ts` verifies provider labels stay distinct from provider-owned payment channels, such as EZPay and Alipay.
