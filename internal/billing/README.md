# Billing package
- `events_part2.go` continues the focused implementation from its original package module.

Payment-method discovery, exact decimal arithmetic, checkout creation, provider event validation, and settlement live here.

## Files

- `service.go` defines settings/repository contracts, service construction, and checkout creation.
- `gateway_contracts.go` defines the server-owned provider checkout/event contracts and gateway interfaces.
- `payment_methods.go` builds the configured rail list, including coupon funding and independently selectable provider-account profiles.
- `events.go` validates and authorizes provider events, settles orders, exposes the narrow signed-return receipt projection, and handles cancellation.
- `payment_profile_runtime.go` gates channels from stable-ID profiles and derives each BEPusdt callback capability from the selected account credential.
- `methods.go` defines canonical payment method identifiers and enabled-rail ordering.
- `decimal.go` parses and computes exact TXB and provider-currency amounts without floating point.
- `service_test.go` covers checkout creation and provider failure paths.
- `events_test.go` covers event validation, authorization, settlement, and BEPUSDT compatibility.
- `service_test_helpers_test.go` contains shared billing repository, settings, and gateway doubles.
- `methods_test.go` covers documented method identifiers and rail ordering.
- `decimal_test.go` covers decimal parsing, equivalence, and rounding.
- `methods_service_test.go` covers configured method availability, cancellation behavior, and BEPusdt callback capabilities.
- `payment_profile_runtime_test.go` covers account-scoped BEPusdt callback capabilities.
- `README.md` documents the package layout.
