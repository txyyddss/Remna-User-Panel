package currencydisplay

import (
	"strconv"
	"testing"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

func TestFormatMoneyUsesFixedPointFlooring(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, minor, rate, want string
		target                  Currency
	}{
		{name: "CNY exact", minor: "1250", rate: "10", target: CNY, want: "1.25 CNY"},
		{name: "USD fractional floor", minor: "100", rate: "3", target: USD, want: "0.33 USD"},
		{name: "negative truncates toward zero", minor: "-100", rate: "3", target: USD, want: "-0.33 USD"},
		{name: "decimal rate", minor: "1250", rate: "2.5", target: CNY, want: "5.00 CNY"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			money, err := FormatMoney(model.TXBMoney(mustMinor(t, test.minor)), test.target, test.rate)
			if err != nil || money.Display != test.want {
				t.Fatalf("FormatMoney() = (%q, %v), want (%q, nil)", money.Display, err, test.want)
			}
		})
	}
}

func TestFormatMoneyRejectsInvalidRate(t *testing.T) {
	t.Parallel()
	if _, err := FormatMoney(model.TXBMoney(100), CNY, "0"); err == nil {
		t.Fatal("FormatMoney() accepted a zero rate")
	}
}

func TestValidRate(t *testing.T) {
	t.Parallel()
	if !ValidRate("2.50") || ValidRate("0") || ValidRate("invalid") {
		t.Fatal("ValidRate() accepted an invalid fixed-decimal rate")
	}
}

func mustMinor(t *testing.T, value string) int64 {
	t.Helper()
	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
