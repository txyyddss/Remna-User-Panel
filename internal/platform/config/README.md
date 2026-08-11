# Runtime configuration

- `config.go` loads environment-backed server, database, Telegram, provider, and security configuration with validated defaults.
- `ADMIN_TELEGRAM_ID` accepts one or more comma-separated positive Telegram user IDs; duplicates are ignored.
- `config_test.go` verifies canonical public-origin validation and complete valid environment loading.
