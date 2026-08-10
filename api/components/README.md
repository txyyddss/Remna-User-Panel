# Component shards

- `security-schemes.yaml` defines the session cookie and three request-signing security schemes.
- `parameters.yaml` contains reusable path, query, and idempotency parameters.
- `responses.yaml` contains reusable JSON error responses.
- `schemas/` contains the bounded schema shards and its definition index.

Internal component references point through `../openapi.yaml` so the root registry remains the single public contract namespace.
