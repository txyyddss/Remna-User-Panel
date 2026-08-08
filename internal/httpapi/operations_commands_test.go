package httpapi

import "testing"

func TestTelegramDeductCommandParsesHumanMajorAmounts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		text   string
		amount string
		ok     bool
	}{
		{name: "plain command", text: "/deduct 12.50", amount: "12.50", ok: true},
		{name: "bot suffix", text: "/deduct@txcarpool_bot 1", amount: "1", ok: true},
		{name: "zero rejected", text: "/deduct 0", ok: false},
		{name: "negative rejected", text: "/deduct -1", ok: false},
		{name: "too precise rejected", text: "/deduct 1.001", ok: false},
		{name: "extra text rejected", text: "/deduct 1 note", ok: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			amount, ok := telegramDeductCommand(test.text)
			if ok != test.ok || (ok && amount != test.amount) {
				t.Fatalf("telegramDeductCommand(%q) = (%s, %t), want (%s, %t)", test.text, amount, ok, test.amount, test.ok)
			}
		})
	}
}
