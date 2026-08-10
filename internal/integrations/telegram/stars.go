package telegram

import (
	"context"
	"errors"
	"strings"
)

// StarsInvoiceRequest describes a one-time Telegram Stars invoice.
type StarsInvoiceRequest struct {
	Title       string
	Description string
	Payload     string
	Label       string
	Amount      int64
}

// CreateStarsInvoiceLink creates a one-line XTR invoice link.
func (c *Client) CreateStarsInvoiceLink(ctx context.Context, invoice StarsInvoiceRequest) (string, error) {
	if strings.TrimSpace(invoice.Title) == "" || len([]rune(invoice.Title)) > 32 {
		return "", errors.New("telegram invoice title must contain 1-32 characters")
	}
	if strings.TrimSpace(invoice.Description) == "" || len([]rune(invoice.Description)) > 255 {
		return "", errors.New("telegram invoice description must contain 1-255 characters")
	}
	if invoice.Payload == "" || len([]byte(invoice.Payload)) > 128 {
		return "", errors.New("telegram invoice payload must contain 1-128 bytes")
	}
	if strings.TrimSpace(invoice.Label) == "" || invoice.Amount <= 0 {
		return "", errors.New("telegram invoice label and positive amount are required")
	}
	payload := struct {
		Title         string         `json:"title"`
		Description   string         `json:"description"`
		Payload       string         `json:"payload"`
		ProviderToken string         `json:"provider_token"`
		Currency      string         `json:"currency"`
		Prices        []LabeledPrice `json:"prices"`
	}{
		Title: invoice.Title, Description: invoice.Description, Payload: invoice.Payload,
		ProviderToken: "", Currency: "XTR", Prices: []LabeledPrice{{Label: invoice.Label, Amount: invoice.Amount}},
	}
	var result string
	if err := c.call(ctx, "createInvoiceLink", payload, &result); err != nil {
		return "", err
	}
	if result == "" {
		return "", errors.New("telegram createInvoiceLink returned an empty URL")
	}
	return result, nil
}

// AnswerPreCheckoutQuery accepts or rejects Telegram's final payment check.
func (c *Client) AnswerPreCheckoutQuery(ctx context.Context, queryID string, approved bool, errorMessage string) error {
	if strings.TrimSpace(queryID) == "" {
		return errors.New("telegram pre-checkout query id is empty")
	}
	payload := struct {
		ID           string `json:"pre_checkout_query_id"`
		OK           bool   `json:"ok"`
		ErrorMessage string `json:"error_message,omitempty"`
	}{ID: queryID, OK: approved, ErrorMessage: errorMessage}
	if !approved && strings.TrimSpace(errorMessage) == "" {
		return errors.New("telegram rejected pre-checkout query requires an error message")
	}
	return c.booleanCall(ctx, "answerPreCheckoutQuery", payload)
}

// GetStarTransactions returns Telegram Stars transactions in chronological order.
func (c *Client) GetStarTransactions(ctx context.Context, offset, limit int) ([]StarTransaction, error) {
	if offset < 0 {
		return nil, errors.New("telegram Stars offset must be non-negative")
	}
	if limit == 0 {
		limit = 100
	}
	if limit < 1 || limit > 100 {
		return nil, errors.New("telegram Stars limit must be in 1..100")
	}
	payload := struct {
		Offset int `json:"offset,omitempty"`
		Limit  int `json:"limit"`
	}{Offset: offset, Limit: limit}
	var result struct {
		Transactions []StarTransaction `json:"transactions"`
	}
	if err := c.call(ctx, "getStarTransactions", payload, &result); err != nil {
		return nil, err
	}
	return result.Transactions, nil
}

// RefundStarPayment asks Telegram to refund a settled Stars payment.
func (c *Client) RefundStarPayment(ctx context.Context, userID int64, chargeID string) error {
	if userID <= 0 || strings.TrimSpace(chargeID) == "" {
		return errors.New("telegram Stars refund requires a positive user id and charge id")
	}
	payload := struct {
		UserID   int64  `json:"user_id"`
		ChargeID string `json:"telegram_payment_charge_id"`
	}{UserID: userID, ChargeID: chargeID}
	return c.booleanCall(ctx, "refundStarPayment", payload)
}
