# Secret vault

- `vault.go` encrypts and decrypts setting-bound secrets using versioned authenticated encryption.
- `vault_test.go` verifies round trips, setting-name binding, key validation, and malformed ciphertext rejection.
