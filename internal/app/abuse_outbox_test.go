package app

import "testing"

func TestEscapeMarkdownEscapesTelegramV2Syntax(t *testing.T) {
	if got, want := escapeMarkdown("a_*[]()!"), "a\\_\\*\\[\\]\\(\\)\\!"; got != want {
		t.Fatalf("escaped = %q, want %q", got, want)
	}
}
