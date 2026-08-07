package ezpay

import (
	"errors"
	"net/url"
	"strings"
	"testing"
)

func TestSignLocalSDKFixture(t *testing.T) {
	t.Parallel()
	values := url.Values{
		"pid":          {"1000"},
		"type":         {"alipay"},
		"notify_url":   {"https://app.example/api/v1/webhooks/ezpay"},
		"return_url":   {"https://app.example/payments/return"},
		"out_trade_no": {"order-1"},
		"name":         {"TXB"},
		"money":        {"12.34"},
		"sign":         {"ignored"},
		"sign_type":    {"MD5"},
		"empty":        {""},
	}
	const want = "02f759754b34efde8f1c88c5c0ede4ef"
	if got := Sign(values, "SECRET"); got != want {
		t.Fatalf("Sign() = %q, want %q", got, want)
	}
}

func TestCheckoutURL(t *testing.T) {
	t.Parallel()
	client, err := NewClient("https://pay.example/gateway", "1000", "SECRET")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	checkout, err := client.CheckoutURL(CheckoutRequest{
		Type: PaymentAlipay, NotifyURL: "https://app.example/api/v1/webhooks/ezpay",
		ReturnURL: "https://app.example/payments/return", OutTradeNo: "order-1", Name: "TXB", Money: "12.34",
	})
	if err != nil {
		t.Fatalf("CheckoutURL() error = %v", err)
	}
	parsed, err := url.Parse(checkout)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if parsed.Path != "/gateway/submit.php" || parsed.Query().Get("sign_type") != "MD5" {
		t.Fatalf("CheckoutURL() = %q", checkout)
	}
	if got := parsed.Query().Get("sign"); got != "02f759754b34efde8f1c88c5c0ede4ef" {
		t.Fatalf("checkout sign = %q", got)
	}
}

func TestParseNotification(t *testing.T) {
	t.Parallel()
	client, err := NewClient("https://pay.example", "1000", "SECRET")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	base := url.Values{
		"pid": {"1000"}, "out_trade_no": {"order-1"}, "trade_no": {"gateway-1"},
		"trade_status": {"TRADE_SUCCESS"}, "type": {"alipay"}, "money": {"12.34"},
	}
	base.Set("sign", Sign(base, "SECRET"))
	base.Set("sign_type", "MD5")

	tests := []struct {
		name    string
		values  url.Values
		wantErr error
	}{
		{name: "valid", values: cloneValues(base)},
		{name: "tampered", values: withEZValue(base, "money", "13.34"), wantErr: ErrInvalidSignature},
		{name: "duplicate", values: withEZDuplicate(base, "money", "12.34"), wantErr: errors.New("duplicate")},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			notification, err := client.ParseNotification(test.values)
			if test.wantErr != nil {
				if err == nil {
					t.Fatal("ParseNotification() error = nil")
				}
				if errors.Is(test.wantErr, ErrInvalidSignature) && !errors.Is(err, ErrInvalidSignature) {
					t.Fatalf("ParseNotification() error = %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseNotification() error = %v", err)
			}
			if !notification.Successful() || notification.OutTradeNo != "order-1" || notification.Money != "12.34" {
				t.Fatalf("notification = %#v", notification)
			}
		})
	}
}

func TestCheckoutValidation(t *testing.T) {
	t.Parallel()
	client, err := NewClient("https://pay.example", "1000", "SECRET")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	tests := []struct {
		name  string
		money string
	}{
		{name: "empty"},
		{name: "zero", money: "0"},
		{name: "negative", money: "-1"},
		{name: "exponent", money: "1e3"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := client.CheckoutURL(CheckoutRequest{
				Type: PaymentAlipay, NotifyURL: "https://app.example/notify", ReturnURL: "https://app.example/return",
				OutTradeNo: "order", Name: "TXB", Money: test.money,
			})
			if err == nil || strings.Contains(err.Error(), "SECRET") {
				t.Fatalf("CheckoutURL() error = %v", err)
			}
		})
	}
}

func cloneValues(values url.Values) url.Values {
	clone := make(url.Values, len(values))
	for key, entries := range values {
		clone[key] = append([]string(nil), entries...)
	}
	return clone
}

func withEZValue(values url.Values, key, value string) url.Values {
	clone := cloneValues(values)
	clone.Set(key, value)
	return clone
}

func withEZDuplicate(values url.Values, key, value string) url.Values {
	clone := cloneValues(values)
	clone[key] = append(clone[key], value)
	return clone
}
