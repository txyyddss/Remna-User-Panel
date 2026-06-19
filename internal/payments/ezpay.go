package payments

import (
	"context"
	"crypto/md5"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"remna-user-panel/internal/config"
)

type ezpayPaymentResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	TradeNo   string `json:"trade_no"`
	PayURL    string `json:"payurl"`
	QRCode    string `json:"qrcode"`
	URLScheme string `json:"urlscheme"`
}

func createEZPayPayment(ctx context.Context, client *http.Client, cfg config.EZPaySettings, notifyURL string, req providerPaymentRequest, payType string) (providerPaymentResponse, error) {
	params := map[string]string{
		"pid":          fmt.Sprintf("%d", cfg.PID),
		"type":         payType,
		"out_trade_no": req.OrderID,
		"notify_url":   notifyURL,
		"return_url":   cfg.ReturnURL,
		"name":         req.Description,
		"money":        floatString(req.Amount),
		"clientip":     req.ClientIP,
		"sign_type":    "MD5",
	}
	params["sign"] = ezpaySign(params, cfg.Key)

	form := url.Values{}
	for key, value := range params {
		form.Set(key, value)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(cfg.BaseURL, "/")+"/mapi.php", strings.NewReader(form.Encode()))
	if err != nil {
		return providerPaymentResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(httpReq)
	if err != nil {
		return providerPaymentResponse{}, fmt.Errorf("ezpay create payment: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return providerPaymentResponse{}, fmt.Errorf("read ezpay response: %w", err)
	}
	var result ezpayPaymentResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return providerPaymentResponse{}, fmt.Errorf("parse ezpay response: %w", err)
	}
	if result.Code != 1 {
		return providerPaymentResponse{}, fmt.Errorf("ezpay payment failed: %s", result.Msg)
	}
	paymentURL := result.PayURL
	if paymentURL == "" {
		paymentURL = result.QRCode
	}
	if paymentURL == "" {
		paymentURL = ezpaySubmitURL(cfg, notifyURL, req, payType)
	}
	return providerPaymentResponse{
		ProviderPaymentID: result.TradeNo,
		PaymentURL:        paymentURL,
		QRContent:         result.QRCode,
		DisplayAmount:     floatString(req.Amount),
		DisplayCurrency:   req.Currency,
		URLScheme:         result.URLScheme,
	}, nil
}

func ezpaySubmitURL(cfg config.EZPaySettings, notifyURL string, req providerPaymentRequest, payType string) string {
	params := map[string]string{
		"pid":          fmt.Sprintf("%d", cfg.PID),
		"type":         payType,
		"out_trade_no": req.OrderID,
		"notify_url":   notifyURL,
		"return_url":   cfg.ReturnURL,
		"name":         req.Description,
		"money":        floatString(req.Amount),
		"sign_type":    "MD5",
	}
	params["sign"] = ezpaySign(params, cfg.Key)
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	return strings.TrimRight(cfg.BaseURL, "/") + "/submit.php?" + values.Encode()
}

func handleEZPayWebhook(cfg config.EZPaySettings, body []byte) (webhookResult, error) {
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return webhookResult{}, fmt.Errorf("parse ezpay webhook: %w", err)
	}
	params := map[string]string{}
	raw := map[string]any{}
	for key, bucket := range values {
		if len(bucket) == 0 {
			continue
		}
		params[key] = bucket[0]
		raw[key] = bucket[0]
	}
	got := params["sign"]
	if got == "" {
		return webhookResult{}, fmt.Errorf("missing ezpay signature")
	}
	expected := ezpaySign(params, cfg.Key)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(got)) != 1 {
		return webhookResult{}, fmt.Errorf("invalid ezpay signature")
	}
	return webhookResult{
		OrderID:           params["out_trade_no"],
		ProviderPaymentID: params["trade_no"],
		Paid:              params["trade_status"] == "TRADE_SUCCESS",
		Raw:               raw,
	}, nil
}

func ezpaySign(params map[string]string, key string) string {
	keys := make([]string, 0, len(params))
	for name, value := range params {
		if name == "sign" || name == "sign_type" || value == "" || value == "0" {
			continue
		}
		keys = append(keys, name)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, name := range keys {
		parts = append(parts, name+"="+params[name])
	}
	hash := md5.Sum([]byte(strings.Join(parts, "&") + key))
	return fmt.Sprintf("%x", hash)
}
