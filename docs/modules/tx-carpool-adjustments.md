# TX Carpool UI adjustment module

## Ownership

This adjustment module owns the two requested member-facing changes: the Add
TXB payment selection flow and daily check-in result feedback. Billing remains
the owner of payment method identity, order creation, and settlement. Activity
remains the owner of the authoritative check-in result and balance update.

## Add TXB flow

The localized billing flow has two visible steps for external payment methods:

1. `PaymentProviderStep` chooses a stable provider-account profile.
2. `PaymentChannelStep` chooses one enabled channel, accepts the TXB amount,
   and starts the existing `methodId`-based payment order.

The browser sends the same provider-owned method ID as before. No provider
credentials, channel mapping, or payment API contract is duplicated in Vue.
Coupon redemption remains on the provider step because it is a balance funding
source without an external payment channel.

## Daily check-in result

The activity result dialog reads the server-returned `txb_delta` reward from a
successful check-in and renders the formatted TXB amount with a localized
`activity.checkInReward` label. The post-action balance remains visible below
it. A check-in with no reward does not render a misleading amount.

## Localization and verification

All new visible copy is present in both English and Simplified Chinese locale
shards. The billing and activity README maps list every new implementation
file. Verification is static: structure audit, lint, typecheck, diff review,
and `git diff --check`; local tests are intentionally not run for this task.
