package botcommands

import (
	"strings"
	"testing"
)

func TestParseExcludesUnknownSlashCommands(t *testing.T) {
	t.Parallel()
	command, slash := Parse("/unknown@txcarpool_bot value")
	if !slash || command.Known || command.Name != "unknown" || len(command.Args) != 1 {
		t.Fatalf("Parse() = (%+v, %t)", command, slash)
	}
}

func TestLanguageForChineseVariants(t *testing.T) {
	t.Parallel()
	for _, code := range []string{"zh", "zh-CN", "zh-hans"} {
		if got := LanguageFor(code); got != Chinese {
			t.Fatalf("LanguageFor(%q) = %q", code, got)
		}
	}
}

func TestLimitBoundsTelegramReply(t *testing.T) {
	t.Parallel()
	result := Limit(strings.Repeat("界", MessageLimit+100))
	if len([]rune(result)) > MessageLimit || !strings.HasSuffix(result, "...") {
		t.Fatalf("bounded result has %d runes", len([]rune(result)))
	}
}
