package payments

import (
	"net/url"
	"strings"
	"testing"

	"remna-user-panel/internal/config"
)

func TestEZPayWebhookVerification(t *testing.T) {
	params := map[string]string{
		"pid":          "1001",
		"type":         "alipay",
		"out_trade_no": "po_test",
		"trade_no":     "ez_test",
		"trade_status": "TRADE_SUCCESS",
		"money":        "10",
	}
	params["sign"] = ezpaySign(params, "secret")
	body := url.Values{}
	for key, value := range params {
		body.Set(key, value)
	}

	result, err := handleEZPayWebhook(config.EZPaySettings{Key: "secret"}, []byte(body.Encode()))
	if err != nil {
		t.Fatalf("handleEZPayWebhook returned error: %v", err)
	}
	if !result.Paid || result.OrderID != "po_test" || result.ProviderPaymentID != "ez_test" {
		t.Fatalf("unexpected webhook result: %#v", result)
	}
}

func TestBEPUSDTWebhookVerification(t *testing.T) {
	body := `{"trade_id":"bp_test","order_id":"po_test","amount":10,"actual_amount":1.23,"token":"wallet","block_transaction_id":"tx","status":2`
	signature := bepusdtSign(map[string]string{
		"trade_id":             "bp_test",
		"order_id":             "po_test",
		"amount":               "10",
		"actual_amount":        "1.23",
		"token":                "wallet",
		"block_transaction_id": "tx",
		"status":               "2",
	}, "secret")
	body += `,"signature":"` + signature + `"}`

	result, err := handleBEPUSDTWebhook(config.BEPUSDTSettings{Token: "secret"}, []byte(body))
	if err != nil {
		t.Fatalf("handleBEPUSDTWebhook returned error: %v", err)
	}
	if !result.Paid || result.OrderID != "po_test" || result.ProviderPaymentID != "bp_test" {
		t.Fatalf("unexpected webhook result: %#v", result)
	}
}

func TestValidateProviderConfig(t *testing.T) {
	err := validateProviderConfig(Config{WebhookBaseURL: "https://example.com"}, ProviderEZPay)
	if err == nil || !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected disabled error, got %v", err)
	}

	err = validateProviderConfig(Config{
		WebhookBaseURL: "https://example.com",
		EZPay:          config.EZPaySettings{Enabled: true, BaseURL: "https://pay.example.com", PID: 10, Key: "secret", ReturnURL: "https://app.example.com/"},
	}, ProviderEZPay)
	if err != nil {
		t.Fatalf("expected configured EZPay to pass, got %v", err)
	}
}
