# Validation package

This package is the shared transport validation boundary.

- `strings.go` defines UTF-8, length, control-character, and regular-expression checks without logging rejected values.
- `request.go` validates HTTP methods, escaped paths, decoded query names and values, and header names and values.
- `json.go` validates every JSON key and string before the strict destination decoder runs. It permits printable Unicode, passwords, Telegram `initData`, multiline text, and CSV content.
- `validation_test.go` exercises accepted international text and rejected malformed input with table-driven cases.

Format-specific handlers remain responsible for domain grammars such as UUIDs, money, dates, and CSV column rules.
