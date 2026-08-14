package billing

import (
	"reflect"
	"testing"
)

func TestPaymentChannelsReturnsStableProviderOrderAndCopy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		provider string
		want     []string
	}{
		{provider: "ezpay", want: []string{"alipay", "wxpay", "qqpay", "bank", "jdpay"}},
		{provider: "bepusdt", want: []string{"usdt.trc20", "usdt.erc20", "usdt.polygon", "usdt.bep20", "usdt.aptos", "usdt.solana", "usdt.xlayer", "usdt.arbitrum", "usdt.plasma", "usdt.ton"}},
		{provider: "unknown", want: nil},
	}
	for _, test := range tests {
		t.Run(test.provider, func(t *testing.T) {
			got := PaymentChannels(test.provider)
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("PaymentChannels(%q) = %v, want %v", test.provider, got, test.want)
			}
			if len(got) > 0 {
				got[0] = "mutated"
				if PaymentChannels(test.provider)[0] == "mutated" {
					t.Fatal("PaymentChannels() returned its internal order slice")
				}
			}
		})
	}
}

func TestValidatePaymentChannels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider string
		channels []string
		wantErr  bool
	}{
		{name: "EZPay trims and normalizes", provider: "ezpay", channels: []string{" ALIPAY ", "wxpay"}},
		{name: "BEPusdt accepts a supported rail", provider: "bepusdt", channels: []string{"usdt.trc20"}},
		{name: "unsupported provider", provider: "other", channels: []string{"rail"}, wantErr: true},
		{name: "unsupported channel", provider: "ezpay", channels: []string{"crypto"}, wantErr: true},
		{name: "duplicate channel", provider: "ezpay", channels: []string{"alipay", " ALIPAY "}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := ValidatePaymentChannels(test.provider, test.channels); (err != nil) != test.wantErr {
				t.Fatalf("ValidatePaymentChannels(%q, %v) error = %v, want error %t", test.provider, test.channels, err, test.wantErr)
			}
		})
	}
}

func TestValidateProviderMethodLists(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		validate func(string) error
		valid    string
		invalid  []string
	}{
		{name: "EZPay", validate: ValidateEZPayMethods, valid: " wxpay, alipay ", invalid: []string{"", "unknown", "alipay,alipay"}},
		{name: "BEPusdt", validate: ValidateBEPusdtMethods, valid: "usdt.trc20,usdt.ton", invalid: []string{"", "usdt.unknown", "usdt.ton,usdt.ton"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := test.validate(test.valid); err != nil {
				t.Fatalf("validate(%q): %v", test.valid, err)
			}
			for _, value := range test.invalid {
				if err := test.validate(value); err == nil {
					t.Errorf("validate(%q) unexpectedly succeeded", value)
				}
			}
		})
	}
}
