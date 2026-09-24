# Telegram formatting

- `markdown.go` escapes dynamic MarkdownV2 values and applies Telegram's
  4096-rune bound at complete escapes, bold spans, and mention links.
- `markdown_test.go` covers reserved characters and markup-safe truncation.
