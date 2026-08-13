# Remnawave integration
- `types_part2.go` continues the focused implementation from its original package module.
- `users_part2.go` continues the focused implementation from its original package module.

This package implements the Remnawave v3.2.1 operations used by TX Carpool. Endpoint and response shapes follow `reference/Upstream/Remnawave/api.json`; the shared transport applies bearer authentication and sanitizes provider errors.

- `client.go` owns client construction, authenticated HTTP transport, bounded response decoding, and sanitized API errors.
- `users.go` implements user creation and identity lookup; `users_part2.go` implements documented stream pagination, update, credential revocation, and traffic reset operations.
- `subscriptions.go` implements protected subscription retrieval and the documented `GET /api/bandwidth-stats/users/{userId}` date-bounded usage statistics.
- `squads.go` implements internal-squad, node, accessible-node, and inbound assignment operations.
- `types.go` defines the provider request, response, status, traffic, subscription, node, and squad contracts.
- `user_identity.go` binds user responses to the exact requested identifier before callers can use bearer data or issue mutations.
- `user_validation.go` strictly validates every reference-required user and nested traffic field during JSON decoding.
- `doc.go` supplies the package documentation and supported upstream version.
- `client_test.go` verifies endpoint, method, payload, bearer, error-code, and lookup contracts.
- `redirect_test.go` verifies provider redirects cannot receive the bearer credential.
- `squad_identity_test.go` rejects a squad update response whose UUID differs from the requested squad.
- `user_validation_test.go` verifies strict required-field response validation and provides canonical user fixtures.
- `user_identity_test.go` rejects mismatched identities from every user-returning operation.
- `README.md` documents the package layout and upstream contract boundary.
- `users_part2.go` contains paged user lookup, mutation payload, and action helpers.
