# Telegram bot commands

This package owns Telegram slash-command parsing, localized reply copy, and
plain-text formatting. HTTP webhook code supplies authenticated domain data;
the package never calls persistence or providers directly.

- `command.go` parses every slash command before group analytics can record it.
- `locale.go` selects English or Simplified Chinese and owns command copy.
- `format.go` formats balance, subscription, combo, and check-in replies within
  Telegram's message limit.
- `command_test.go` covers suffix parsing, unknown-command exclusion, locale
  selection, and bounded output.

