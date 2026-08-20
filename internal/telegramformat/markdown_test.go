package telegramformat

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestEscapeAndLimit(t *testing.T) {
	t.Parallel()
	if got := Escape(`a_[b]!`); got != `a\_\[b\]\!` {
		t.Fatalf("Escape() = %q", got)
	}
	got := Limit(strings.Repeat("a", MessageLimit+10) + `\`)
	if utf8.RuneCountInString(got) > MessageLimit || strings.HasSuffix(strings.TrimSuffix(got, ellipsis), `\`) {
		t.Fatalf("Limit() produced unsafe output of %d runes", utf8.RuneCountInString(got))
	}
}
