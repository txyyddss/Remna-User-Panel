package payments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
)

func createTelegramStarsInvoice(ctx context.Context, client *http.Client, botToken string, req providerPaymentRequest) (providerPaymentResponse, error) {
	amount := int(math.Round(req.Amount))
	if amount <= 0 {
		return providerPaymentResponse{}, fmt.Errorf("invalid_stars_amount")
	}
	title := strings.TrimSpace(req.Description)
	if title == "" {
		title = "Subscription"
	}
	if len([]rune(title)) > 32 {
		title = string([]rune(title)[:32])
	}
	description := strings.TrimSpace(req.Description)
	if description == "" {
		description = title
	}
	if len([]rune(description)) > 255 {
		description = string([]rune(description)[:255])
	}
	payload := map[string]any{"title": title, "description": description, "payload": req.OrderID, "currency": "XTR", "prices": []map[string]any{{"label": title, "amount": amount}}}
	body, _ := json.Marshal(payload)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.telegram.org/bot"+strings.TrimSpace(botToken)+"/createInvoiceLink", bytes.NewReader(body))
	if err != nil {
		return providerPaymentResponse{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return providerPaymentResponse{}, fmt.Errorf("telegram_create_invoice: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	responseBody, _ := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	var result struct {
		OK          bool   `json:"ok"`
		Result      string `json:"result"`
		Description string `json:"description"`
	}
	if json.Unmarshal(responseBody, &result) != nil || !result.OK || result.Result == "" {
		return providerPaymentResponse{}, fmt.Errorf("telegram_create_invoice_failed: %s", result.Description)
	}
	return providerPaymentResponse{ProviderPaymentID: req.OrderID, PaymentURL: result.Result, DisplayAmount: fmt.Sprintf("%d", amount), DisplayCurrency: "XTR"}, nil
}
