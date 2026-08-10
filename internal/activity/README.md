# Activity domain

- `activity.go` defines games, rewards, betting, and daily check-in types and validation.
- `draws.go` defines lucky draws, prize outcomes, extension credits, history, and group rewards.
- `service.go` validates and coordinates member and administrator activity workflows.
- `random.go` provides the cryptographically secure random source.
- `activity_test.go` verifies validation, idempotency, randomness, and service behavior.
