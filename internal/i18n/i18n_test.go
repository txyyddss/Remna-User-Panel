package i18n

import "testing"

func TestTranslateFallsBackToChinese(t *testing.T) {
	catalog := &Catalog{
		defaultLang: "zh",
		messages: map[string]map[string]string{
			"zh": {"hello": "你好"},
			"en": {"hello": "Hello"},
		},
	}
	if got := catalog.Translate("de", "hello"); got != "你好" {
		t.Fatalf("Translate() = %q, want Chinese fallback", got)
	}
}
