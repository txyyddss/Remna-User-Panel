# Frontend API

- `client.ts` exposes typed member and administrator API operations.
- `client.test.ts` covers client request behavior.
- `features.ts` exposes feature-specific endpoints and contract types.
- `features.test.ts` verifies feature request construction.
- `generated.ts` contains the compact generated OpenAPI contract.
- `http.ts` provides authenticated HTTP and error handling primitives.
- `request-signing.ts` signs mutation requests when required.
- `request-signing.test.ts` covers signing and nonce behavior.
- `types.ts` exports stable aliases over generated schema types.

