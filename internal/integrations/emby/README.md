# Emby integration

- `client.go` implements the typed HTTP adapter and maps Emby payloads into the domain model.
- `validation.go` validates the configured server URL and Emby GUID identifiers.
- `client_test.go` verifies request paths, authentication, payloads, response mapping, and error handling.
- `boundary_validation_test.go` rejects malformed user identifiers returned by list and create operations.
- `redirect_test.go` verifies provider redirects cannot receive the API token or password body.
- `doc.go` documents the package boundary.
