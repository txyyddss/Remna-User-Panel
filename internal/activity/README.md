# Activity domain

- `activity.go` defines games, betting, and daily check-in types and validation.
- `reward.go` defines typed draw effects, validation, and persistence-facing payloads.
- `draws.go` defines instant probabilities, raffle stock, outcomes, and group rewards.
- `draw_random.go` validates quantized ranges and samples uniform, truncated Gaussian, and power-law distributions with the injected secure integer source.
- `raffles.go` defines group-entry receipts and settlement summaries; `service.go` exposes publication, entry, and settlement to HTTP and the durable worker.
- `service.go` validates and coordinates member and administrator activity workflows.
- `random.go` provides the cryptographically secure random source.
- `activity_test.go` verifies validation, idempotency, randomness, and service behavior.
