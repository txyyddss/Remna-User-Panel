package onboarding

import (
	"unicode"
)

func MessageDurationMS(text string) int {
	cjk := 0
	latinWords := 0
	inLatinWord := false
	for _, character := range text {
		if isCJK(character) {
			cjk++
			if inLatinWord {
				latinWords++
				inLatinWord = false
			}
			continue
		}
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			inLatinWord = true
		} else if inLatinWord {
			latinWords++
			inLatinWord = false
		}
	}
	if inLatinWord {
		latinWords++
	}
	milliseconds := 600 + latinWords*60_000/220 + cjk*60_000/300
	if milliseconds < 1800 {
		return 1800
	}
	if milliseconds > 12000 {
		return 12000
	}
	return milliseconds
}

func isCJK(value rune) bool {
	return unicode.In(value, unicode.Han, unicode.Hiragana, unicode.Katakana, unicode.Hangul)
}

