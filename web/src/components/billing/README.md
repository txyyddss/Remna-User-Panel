# Billing components

- `BalancePaymentSheet.vue` drives coupon funding and the provider order lifecycle from Home, including a verified failed/expired-order reissue that creates a replacement order.
- `BalancePaymentConfiguration.vue` presents the localized amount, rail, channel, and coupon-funding selection step.
- `BalancePaymentSheet.test.ts` verifies provider labels stay distinct from provider-owned payment channels, such as EZPay and Alipay.
