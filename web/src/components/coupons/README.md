# Coupon components

- `CouponWalletPanel.vue` redeems coupon codes and confirms server-backed grant discards.
- `CouponGrantList.vue` renders wallet grants and emits an explicit discard request.

Redemption uses action feedback, opening discard review is soft, and final discard confirmation is heavy.
