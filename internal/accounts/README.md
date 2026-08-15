# Accounts package

Trusted Telegram authentication and resumable account onboarding live here. Network and persistence dependencies are expressed as package-level interfaces so the application layer can supply adapters.

## Files

- `service.go` defines service contracts, authentication, session lookup, and service construction.
- `membership.go` creates signed Telegram join invites, checks membership, and handles join requests.
- `onboarding.go` reserves usernames and reconciles agreement acceptance with
  Remnawave, preferring a persisted upstream identity for returning users.
- `service_test.go` covers authentication and session behavior.
- `membership_test.go` covers invite and membership flows.
- `onboarding_test.go` covers username reservation and agreement reconciliation.
- `service_test_helpers_test.go` contains package-wide repository and upstream test doubles.
- `README.md` documents the package layout.
