package payments

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"remna-user-panel/internal/config"
)

type bepusdtCreateRequest struct {
	OrderID     string      `json:"order_id"`
	Amount      json.Number `json:"amount"`
	NotifyURL   string      `json:"notify_url"`
	RedirectURL string      `json:"redirect_url"`
	Signature   string      `json:"signature"`
	TradeType   string      `json:"trade_type,omitempty"`
	Fiat        string      `json:"fiat,omitempty"`
	Name        string      `json:"name,omitempty"`
	Timeout     int         `json:"timeout,omitempty"`
}

type bepusdtTransactionResponse struct {
	StatusCode int                    `json:"status_code"`
	Message    string                 `json:"message"`
	Data       bepusdtTransactionData `json:"data"`
}

type bepusdtTransactionData struct {
	Fiat           string `json:"fiat"`
	TradeID        string `json:"trade_id"`
	OrderID        string `json:"order_id"`
	Amount         string `json:"amount"`
	ActualAmount   string `json:"actual_amount"`
	Token          string `json:"token"`
	ExpirationTime int    `json:"expiration_time"`
	Status         int    `json:"status"`
	PaymentURL     string `json:"payment_url"`
}

type bepusdtCallbackData struct {
	TradeID            string  `json:"trade_id"`
	OrderID            string  `json:"order_id"`
	Amount             float64 `json:"amount"`
	ActualAmount       float64 `json:"actual_amount"`
	Token              string  `json:"token"`
	BlockTransactionID string  `json:"block_transaction_id"`
	Signature          string  `json:"signature"`
	Status             int     `json:"status"`
}

func createBEPUSDTPayment(ctx context.Context, client *http.Client, cfg config.BEPUSDTSettings, notifyURL string, req providerPaymentRequest, tradeType string) (providerPaymentResponse, error) {
	if strings.TrimSpace(tradeType) == "" {
		tradeType = "usdt.polygon"
	}
	payload := bepusdtCreateRequest{
		OrderID:     req.OrderID,
		Amount:      json.Number(floatString(req.Amount)),
		NotifyURL:   notifyURL,
		RedirectURL: cfg.ReturnURL,
		TradeType:   tradeType,
		Fiat:        req.Currency,
		Name:        req.Description,
		Timeout:     600,
	}
	payload.Signature = bepusdtSign(map[string]string{
		"order_id":     payload.OrderID,
		"amount":       payload.Amount.String(),
		"notify_url":   payload.NotifyURL,
		"redirect_url": payload.RedirectURL,
		"trade_type":   payload.TradeType,
		"fiat":         payload.Fiat,
		"name":         payload.Name,
		"timeout":      strconv.Itoa(payload.Timeout),
	}, cfg.Token)

	body, err := json.Marshal(payload)
	if err != nil {
		return providerPaymentResponse{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(cfg.BaseURL, "/")+"/api/v1/order/create-transaction", bytes.NewReader(body))
	if err != nil {
		return providerPaymentResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(httpReq)
	if err != nil {
		return providerPaymentResponse{}, fmt.Errorf("bepusdt create payment: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return providerPaymentResponse{}, fmt.Errorf("read bepusdt response: %w", err)
	}
	var result bepusdtTransactionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return providerPaymentResponse{}, fmt.Errorf("parse bepusdt response: %w", err)
	}
	if result.StatusCode != 200 {
		return providerPaymentResponse{}, fmt.Errorf("bepusdt payment failed: %s", result.Message)
	}
	network := ""
	if _, tail, ok := strings.Cut(tradeType, "."); ok {
		network = strings.ToUpper(tail)
	}
	qrContent := result.Data.PaymentURL
	if qrContent == "" {
		qrContent = result.Data.Token
	}
	displayAmount := result.Data.ActualAmount
	if displayAmount == "" {
		displayAmount = result.Data.Amount
	}
	return providerPaymentResponse{
		ProviderPaymentID: result.Data.TradeID,
		PaymentURL:        result.Data.PaymentURL,
		QRContent:         qrContent,
		DisplayAmount:     displayAmount,
		DisplayCurrency:   "USDT",
		PaymentAddress:    result.Data.Token,
		Network:           network,
		ExpiresInSeconds:  result.Data.ExpirationTime,
	}, nil
}

func handleBEPUSDTWebhook(cfg config.BEPUSDTSettings, body []byte) (webhookResult, error) {
	var data bepusdtCallbackData
	if err := json.Unmarshal(body, &data); err != nil {
		return webhookResult{}, fmt.Errorf("parse bepusdt webhook: %w", err)
	}
	expected := bepusdtSign(map[string]string{
		"trade_id":             data.TradeID,
		"order_id":             data.OrderID,
		"amount":               formatCallbackNumber(data.Amount),
		"actual_amount":        formatCallbackNumber(data.ActualAmount),
		"token":                data.Token,
		"block_transaction_id": data.BlockTransactionID,
		"status":               strconv.Itoa(data.Status),
	}, cfg.Token)
	if data.Signature == "" || subtle.ConstantTimeCompare([]byte(expected), []byte(data.Signature)) != 1 {
		return webhookResult{}, fmt.Errorf("invalid bepusdt signature")
	}
	var raw map[string]any
	if err := json.Unmarshal(body, &raw); err != nil {
		raw = map[string]any{}
	}
	return webhookResult{
		OrderID:           data.OrderID,
		ProviderPaymentID: data.TradeID,
		Paid:              data.Status == 2,
		Raw:               raw,
	}, nil
}

func bepusdtSign(params map[string]string, token string) string {
	keys := make([]string, 0, len(params))
	for name, value := range params {
		if name == "signature" || value == "" {
			continue
		}
		keys = append(keys, name)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, name := range keys {
		parts = append(parts, name+"="+params[name])
	}
	hash := md5.Sum([]byte(strings.Join(parts, "&") + token))
	return fmt.Sprintf("%x", hash)
}

func formatCallbackNumber(amount float64) string {
	return strconv.FormatFloat(amount, 'f', -1, 64)
}
