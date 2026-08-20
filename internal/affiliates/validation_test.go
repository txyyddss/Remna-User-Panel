package affiliates

import "testing"

func TestValidateTiers(t *testing.T) {
	valid := []Tier{
		{Name: "Starter", Threshold: 0, Enabled: true, Reward: Reward{Kind: "none"}},
		{Name: "Partner", Threshold: 3, Enabled: true, CommissionEnabled: true, CommissionBPS: 725, Reward: Reward{Kind: "txb", TXBMinor: 500}},
	}
	if err := ValidateTiers(valid); err != nil {
		t.Fatalf("ValidateTiers() error = %v", err)
	}
	invalid := [][]Tier{
		{},
		{{Name: "Late", Threshold: 1, Enabled: true, Reward: Reward{Kind: "none"}}},
		{{Name: "Starter", Threshold: 0, Enabled: true, Reward: Reward{Kind: "txb", TXBMinor: 1}}},
		{{Name: "Starter", Threshold: 0, Enabled: true, CommissionBPS: 1, Reward: Reward{Kind: "none"}}},
	}
	for index, tiers := range invalid {
		if err := ValidateTiers(tiers); err == nil {
			t.Errorf("ValidateTiers(invalid[%d]) returned nil", index)
		}
	}
}

func TestNormalizeLocale(t *testing.T) {
	for input, expected := range map[string]string{"zh-hans": LocaleChinese, "zh_CN": LocaleChinese, "en-US": LocaleEnglish, "": LocaleEnglish} {
		if actual := NormalizeLocale(input); actual != expected {
			t.Errorf("NormalizeLocale(%q) = %q, want %q", input, actual, expected)
		}
	}
}
