# Telegram bot commands

This package owns Telegram slash-command parsing, localized reply copy, and
MarkdownV2-safe formatting. HTTP webhook code supplies authenticated domain data;
the package never calls persistence or providers directly.

- `command.go` parses every slash command before group analytics can record it.
- `locale.go` selects English or Simplified Chinese and owns command copy.
- `format.go` formats command status cards, balance, subscription, combo,
  check-in, and deduction replies within Telegram's message limit.
- `markdown.go` escapes dynamic content and truncates replies without leaving
  incomplete MarkdownV2 escape sequences.
- `command_test.go` covers suffix parsing, unknown-command exclusion, locale
  selection, and bounded output.
- `format_test.go` covers every localized command-card state, MarkdownV2
  escaping, and truncation at an escape boundary.
