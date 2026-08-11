# Billing components

- `BalancePaymentSheet.vue` drives the provider order lifecycle from Home, including a verified failed/expired-order reissue that creates a replacement order.
- `BalancePaymentSheet.test.ts` verifies provider labels stay distinct from provider-owned payment channels, such as EZPay and Alipay.
