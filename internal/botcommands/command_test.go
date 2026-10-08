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

func TestParseKeepsMyComboAsOnlyComboCommand(t *testing.T) {
	t.Parallel()
	myCombo, slash := Parse("/mycombo@txcarpool_bot")
	if !slash || !myCombo.Known || myCombo.Name != MyCombo {
		t.Fatalf("Parse(/mycombo) = (%+v, %t)", myCombo, slash)
	}
	combo, slash := Parse("/combo")
	if !slash || combo.Known || combo.Name != "combo" {
		t.Fatalf("Parse(/combo) = (%+v, %t)", combo, slash)
	}
}

func TestParseRecognizesPMAdminCommands(t *testing.T) {
	t.Parallel()
	for _, name := range []Name{Refund, AddTXB, DeductTXB} {
		command, slash := Parse("/" + string(name) + "@txcarpool_bot 25.50")
		if !slash || !command.Known || command.Name != name || len(command.Args) != 1 || command.Args[0] != "25.50" {
			t.Fatalf("Parse(/%s) = (%+v, %t)", name, command, slash)
		}
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
	if len([]rune(result)) > MessageLimit || !strings.HasSuffix(result, markdownV2Ellipsis) {
		t.Fatalf("bounded result has %d runes", len([]rune(result)))
	}
}
