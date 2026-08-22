package billing

import "testing"

func TestDocumentedPaymentMethodIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		methodID string
		provider string
		rail     string
	}{
		{methodID: "ezpay:alipay", provider: "ezpay", rail: "alipay"},
		{methodID: "ezpay:wxpay", provider: "ezpay", rail: "wxpay"},
		{methodID: "ezpay:qqpay", provider: "ezpay", rail: "qqpay"},
		{methodID: "ezpay:bank", provider: "ezpay", rail: "bank"},
		{methodID: "ezpay:jdpay", provider: "ezpay", rail: "jdpay"},
		{methodID: "bepusdt:usdt.trc20", provider: "bepusdt", rail: "usdt.trc20"},
		{methodID: "bepusdt:usdt.erc20", provider: "bepusdt", rail: "usdt.erc20"},
		{methodID: "bepusdt:usdt.polygon", provider: "bepusdt", rail: "usdt.polygon"},
		{methodID: "bepusdt:usdt.bep20", provider: "bepusdt", rail: "usdt.bep20"},
		{methodID: "bepusdt:usdt.aptos", provider: "bepusdt", rail: "usdt.aptos"},
		{methodID: "bepusdt:usdt.solana", provider: "bepusdt", rail: "usdt.solana"},
		{methodID: "bepusdt:usdt.xlayer", provider: "bepusdt", rail: "usdt.xlayer"},
		{methodID: "bepusdt:usdt.arbitrum", provider: "bepusdt", rail: "usdt.arbitrum"},
		{methodID: "bepusdt:usdt.plasma", provider: "bepusdt", rail: "usdt.plasma"},
		{methodID: "bepusdt:usdt.ton", provider: "bepusdt", rail: "usdt.ton"},
		{methodID: "bepusdt:usdc.trc20", provider: "bepusdt", rail: "usdc.trc20"},
		{methodID: "bepusdt:usdc.erc20", provider: "bepusdt", rail: "usdc.erc20"},
		{methodID: "bepusdt:usdc.polygon", provider: "bepusdt", rail: "usdc.polygon"},
		{methodID: "bepusdt:usdc.bep20", provider: "bepusdt", rail: "usdc.bep20"},
		{methodID: "bepusdt:usdc.aptos", provider: "bepusdt", rail: "usdc.aptos"},
		{methodID: "bepusdt:usdc.solana", provider: "bepusdt", rail: "usdc.solana"},
		{methodID: "bepusdt:usdc.xlayer", provider: "bepusdt", rail: "usdc.xlayer"},
		{methodID: "bepusdt:usdc.arbitrum", provider: "bepusdt", rail: "usdc.arbitrum"},
		{methodID: "bepusdt:usdc.base", provider: "bepusdt", rail: "usdc.base"},
		{methodID: "stars", provider: "stars"},
	}
	for _, test := range tests {
		t.Run(test.methodID, func(t *testing.T) {
			t.Parallel()
			provider, rail, err := ParseMethodID(test.methodID)
			if err != nil || provider != test.provider || rail != test.rail {
				t.Fatalf("ParseMethodID(%q) = %q, %q, %v", test.methodID, provider, rail, err)
			}
			if !CanonicalMethodID(test.methodID) {
				t.Fatalf("CanonicalMethodID(%q) = false", test.methodID)
			}
		})
	}
}

func TestPaymentMethodIDRejectsUnknownAndMalformedRails(t *testing.T) {
	t.Parallel()

	for _, methodID := range []string{"", "ezpay:", "unknown:rail", "ezpay:crypto", "bepusdt:usdc.unknown", "stars:default"} {
		methodID := methodID
		t.Run(methodID, func(t *testing.T) {
			t.Parallel()
			if _, _, err := ParseMethodID(methodID); err == nil {
				t.Fatalf("ParseMethodID(%q) unexpectedly succeeded", methodID)
			}
		})
	}
}

func TestProfileScopedPaymentMethodID(t *testing.T) {
	t.Parallel()

	provider, profileID, rail, err := ParseMethodSelection("ezpay:backup-account:alipay")
	if err != nil || provider != "ezpay" || profileID != "backup-account" || rail != "alipay" {
		t.Fatalf("ParseMethodSelection() = %q, %q, %q, %v", provider, profileID, rail, err)
	}
	if !CanonicalMethodID("ezpay:backup-account:alipay") {
		t.Fatal("CanonicalMethodID(profile-scoped) = false")
	}
}

func TestEnabledPaymentRailsPreserveAdministratorOrder(t *testing.T) {
	t.Parallel()

	want := []string{"jdpay", "alipay", "bank"}
	got, err := parseEnabledRails("jdpay,alipay,bank", ezpayRails)
	if err != nil || len(got) != len(want) {
		t.Fatalf("parseEnabledRails() = %#v, %v", got, err)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("rail[%d] = %q, want %q", index, got[index], want[index])
		}
	}
	if _, err := parseEnabledRails("alipay,alipay", ezpayRails); err == nil {
		t.Fatal("parseEnabledRails() accepted a duplicate rail")
	}
}
