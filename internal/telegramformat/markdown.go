// Package telegramformat owns shared Telegram MarkdownV2 safety helpers.
package telegramformat

import (
	"strings"
	"unicode/utf8"
)

// MessageLimit is Telegram's maximum sendMessage text length in runes.
const MessageLimit = 4096

const ellipsis = `\.\.\.`

var escaper = strings.NewReplacer(
	"\\", "\\\\", "_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(", "\\(", ")", "\\)",
	"~", "\\~", "`", "\\`", ">", "\\>", "#", "\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|", "\\|",
	"{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!",
)

// Escape replaces every MarkdownV2 control character in dynamic text.
func Escape(value string) string {
	return escaper.Replace(value)
}

// Limit truncates safely without leaving an unmatched escape character.
func Limit(value string) string {
	if utf8.RuneCountInString(value) <= MessageLimit {
		return value
	}
	runes := []rune(value)
	limit := MessageLimit - utf8.RuneCountInString(ellipsis)
	truncated := runes[:limit]
	for trailingBackslashes(truncated)%2 == 1 {
		truncated = truncated[:len(truncated)-1]
	}
	return strings.TrimSpace(string(truncated)) + ellipsis
}

func trailingBackslashes(value []rune) int {
	count := 0
	for index := len(value) - 1; index >= 0 && value[index] == '\\'; index-- {
		count++
	}
	return count
}
