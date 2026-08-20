# Telegram formatting

- `markdown.go` escapes dynamic MarkdownV2 values and applies Telegram's
  4096-rune message bound without leaving an incomplete escape.
- `markdown_test.go` covers reserved characters and safe truncation.
