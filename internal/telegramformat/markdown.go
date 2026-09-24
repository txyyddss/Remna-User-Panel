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

// Limit truncates at a complete MarkdownV2 token and closes an open bold span.
// Project messages use escaped text, bold labels, and Telegram mention links.
func Limit(value string) string {
	if utf8.RuneCountInString(value) <= MessageLimit {
		return value
	}
	runes := []rune(value)
	max := MessageLimit - utf8.RuneCountInString(ellipsis)
	last, closeBold := 0, false
	bold, boldHasText, linkStage := false, false, 0
	for index := 0; index < len(runes) && index < max; index++ {
		character := runes[index]
		if character == '\\' && index+1 < len(runes) {
			index++ // Keep an escaped control character and its slash together.
			if bold {
				boldHasText = true
			}
		} else {
			switch {
			case linkStage == 1 && character == ']':
				linkStage = 2
			case linkStage == 2 && character == '(':
				linkStage = 3
			case linkStage == 3 && character == ')':
				linkStage = 0
			case linkStage == 0 && character == '[':
				linkStage = 1
			case linkStage == 0 && character == '*':
				bold = !bold
				boldHasText = false
			default:
				if bold {
					boldHasText = true
				}
				if linkStage == 2 {
					linkStage = 1 // A closing bracket without a URL was plain text.
				}
			}
		}
		end := index + 1
		if linkStage == 0 && (!bold || boldHasText) && end+boolRune(bold) <= max {
			last, closeBold = end, bold
		}
	}
	if last == 0 {
		return ellipsis
	}
	truncated := strings.TrimSpace(string(runes[:last]))
	if closeBold {
		truncated += "*"
	}
	return truncated + ellipsis
}

func boolRune(value bool) int {
	if value {
		return 1
	}
	return 0
}

func trailingBackslashes(value []rune) int {
	count := 0
	for index := len(value) - 1; index >= 0 && value[index] == '\\'; index-- {
		count++
	}
	return count
}
