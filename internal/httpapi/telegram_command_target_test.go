package httpapi

import "testing"

func TestNormalizeTelegramCommand(t *testing.T) {
	tests := []struct {
		name, text, bot, want string
		accepted              bool
	}{
		{name: "plain raffle command", text: "/raffle", bot: "TxCarpoolBot", want: "/raffle", accepted: true},
		{name: "addressed raffle command", text: "/raffle@TxCarpoolBot", bot: "TxCarpoolBot", want: "/raffle", accepted: true},
		{name: "case insensitive target", text: "/raffle@txcarpoolbot", bot: "TxCarpoolBot", want: "/raffle", accepted: true},
		{name: "built-in arguments", text: "/start@TxCarpoolBot 123", bot: "TxCarpoolBot", want: "/start 123", accepted: true},
		{name: "other bot", text: "/raffle@OtherBot", bot: "TxCarpoolBot"},
		{name: "unknown own identity", text: "/raffle@TxCarpoolBot"},
		{name: "keyword", text: "raffle keyword", bot: "TxCarpoolBot", want: "raffle keyword", accepted: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, accepted := normalizeTelegramCommand(test.text, test.bot)
			if got != test.want || accepted != test.accepted {
				t.Fatalf("normalizeTelegramCommand(%q) = (%q, %t), want (%q, %t)", test.text, got, accepted, test.want, test.accepted)
			}
		})
	}
}
