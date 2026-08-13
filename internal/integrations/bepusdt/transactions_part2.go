package bepusdt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (c *Client) CancelTransaction(ctx context.Context, tradeID string) error {
	tradeID = strings.TrimSpace(tradeID)
	if tradeID == "" {
		return errors.New("bepusdt trade id is empty")
	}
	encoded, err := json.Marshal(map[string]string{
		"trade_id": tradeID, "signature": Sign(map[string]string{"trade_id": tradeID}, c.token),
	})
	if err != nil {
		return fmt.Errorf("bepusdt encode cancellation: %w", err)
	}
	target := *c.baseURL
	target.Path = strings.TrimRight(target.Path, "/") + "/api/v1/order/cancel-transaction"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), bytes.NewReader(encoded))
	if err != nil {
		return fmt.Errorf("bepusdt create cancellation request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("bepusdt cancellation request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil {
		return fmt.Errorf("bepusdt read cancellation response: %w", err)
	}
	if len(body) > maxResponseBytes {
		return fmt.Errorf("bepusdt response exceeds %d bytes", maxResponseBytes)
	}
	var envelope struct {
		StatusCode flexibleInt `json:"status_code"`
		Message    string      `json:"message"`
		RequestID  string      `json:"request_id"`
		Data       struct {
			TradeID string `json:"trade_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("bepusdt decode cancellation response (http=%d): %w", response.StatusCode, err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices || int(envelope.StatusCode) != http.StatusOK {
		return &APIError{HTTPStatus: response.StatusCode, StatusCode: int(envelope.StatusCode), Message: envelope.Message, RequestID: envelope.RequestID}
	}
	if envelope.Data.TradeID != tradeID {
		return errors.New("bepusdt cancellation response trade id mismatch")
	}
	return nil
}
