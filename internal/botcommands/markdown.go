package botcommands

import (
	"strings"
	"unicode/utf8"
)

const markdownV2Ellipsis = `\.\.\.`

var markdownV2Escaper = strings.NewReplacer(
	"\\", "\\\\", "_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]", "(", "\\(", ")", "\\)",
	"~", "\\~", "`", "\\`", ">", "\\>", "#", "\\#", "+", "\\+", "-", "\\-", "=", "\\=", "|", "\\|", "{", "\\{", "}", "\\}", ".", "\\.", "!", "\\!",
)

func escapeMarkdownV2(value string) string {
	return markdownV2Escaper.Replace(value)
}

func limitMarkdownV2(value string) string {
	if utf8.RuneCountInString(value) <= MessageLimit {
		return value
	}
	limit := MessageLimit - utf8.RuneCountInString(markdownV2Ellipsis)
	runes := []rune(value)
	truncated := runes[:limit]
	for trailingBackslashes(truncated)%2 == 1 {
		truncated = truncated[:len(truncated)-1]
	}
	return strings.TrimSpace(string(truncated)) + markdownV2Ellipsis
}

func trailingBackslashes(value []rune) int {
	count := 0
	for index := len(value) - 1; index >= 0 && value[index] == '\\'; index-- {
		count++
	}
	return count
}
