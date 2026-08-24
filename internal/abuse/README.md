# Abuse detector

- `types.go` defines the safe persisted and transport-facing detector values.
- `parser.go` accepts only direct Xray domain connection accepts and fingerprints lines without retaining them.
- `service.go` authenticates node reports, aggregates samples, evaluates grace-delayed streaks, and queues incidents.
- `admin.go` exposes validated member and administrator detector operations.
- `node_keys.go` provisions, encrypts, copies, and rotates node agent keys.
- `repository.go` declares the persistence boundary used by the detector.
- `parser_test.go` covers accepted direct domains and rejection of IP, proxy, and error records.
