# Billing components

- `BalancePaymentSheet.vue` drives coupon funding and the provider order lifecycle from Home, shows the redeemed TXB gain, exposes Telegram native back navigation, and handles failed/expired-order reissue.
- `BalancePaymentConfiguration.vue` is the billing flow entrypoint. It keeps provider selection, channel selection, coupon redemption, and order actions coordinated without changing the payment API contract.
- `PaymentProviderStep.vue` renders the first Add TXB step, choosing one stable provider-account profile or the coupon funding source.
- `PaymentChannelStep.vue` renders the second Add TXB step, choosing a channel, entering the TXB amount, and starting the provider order.
- `paymentOptions.ts` owns the typed provider and channel option contracts shared by the two step components.
- `PaymentReceiptDetails.vue` renders the browser-safe provider-return receipt with localized amount, payment ID, provider, channel, status, and timing metadata using Nuxt UI surfaces.
- `BalancePaymentSheet.test.ts` verifies provider labels stay distinct from provider-owned payment channels, such as EZPay and Alipay, and verifies the two-step transition.
