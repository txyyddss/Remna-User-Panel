package botcommands

import (
	"strings"
	"testing"
)

func TestLocalizedBoostRefusals(t *testing.T) {
	for _, language := range []Language{English, Chinese} {
		copy := Text(language)
		required := FormatGroupBoostRequired(copy)
		unavailable := FormatGroupBoostUnavailable(copy)
		if copy.GroupBoostRequired == "" || copy.GroupBoostUnavailable == "" || !strings.Contains(required, escapeMarkdownV2(copy.GroupBoostRequired)) || !strings.Contains(unavailable, escapeMarkdownV2(copy.GroupBoostUnavailable)) {
			t.Fatalf("missing safe boost copy for %s", language)
		}
		if required == unavailable {
			t.Fatal("provider failure was represented as ineligibility")
		}
	}
}
