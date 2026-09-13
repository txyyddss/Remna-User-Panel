# Coupon components

`CouponGrantList.vue` owns stable-ID row presence and layout reflow. Coupon
feedback remains localized and the discard action keeps its existing durable
mutation boundary.

- `CouponWalletPanel.vue` redeems coupon codes and confirms server-backed grant discards.
- `CouponGrantList.vue` renders wallet grants and emits an explicit discard request.

Redemption uses action feedback, opening discard review is soft, and final discard confirmation is heavy.

`CouponWalletPanel.vue` owns local wallet-state presence while `CouponGrantList.vue` owns coupon-row layout changes.
