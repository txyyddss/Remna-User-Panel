# Coupons

- `coupons.go` defines coupon/grant types, validation, normalization, eligibility, and fixed-point discount math.
- `service.go` coordinates definitions, redemptions, wallet reads, member soft-discard, and purchase quotes.

Wallet deletion writes a durable discard record rather than removing a grant. Its
redemption and purchase foreign keys remain available for history, while quotes
and checkout reject discarded grants. Administrator definition deletion remains
deactivation, so existing history is preserved there as well.
- `coupons_test.go` verifies coupon validation, arithmetic, and application behavior.
