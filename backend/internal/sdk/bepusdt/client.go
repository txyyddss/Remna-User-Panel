package bepusdt

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// Client is the BEPusdt API client
type Client struct {
	baseURL    string
	token      string
	notifyURL  string
	redirectURL string
	httpClient *http.Client
}

// NewClient creates a new BEPusdt client
func NewClient(baseURL, token, notifyURL, redirectURL string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		notifyURL:  notifyURL,
		redirectURL: redirectURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// --- Types ---

type CreateTransactionRequest struct {
	OrderID     string  `json:"order_id"`
	Amount      float64 `json:"amount"`
	NotifyURL   string  `json:"notify_url"`
	RedirectURL string  `json:"redirect_url"`
	Signature   string  `json:"signature"`
	TradeType   string  `json:"trade_type,omitempty"`
	Fiat        string  `json:"fiat,omitempty"`
	Name        string  `json:"name,omitempty"`
	Timeout     int     `json:"timeout,omitempty"`
}

type TransactionResponse struct {
	StatusCode int              `json:"status_code"`
	Message    string           `json:"message"`
	Data       TransactionData  `json:"data"`
}

type TransactionData struct {
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

type CreateOrderRequest struct {
	OrderID     string  `json:"order_id"`
	Amount      float64 `json:"amount"`
	NotifyURL   string  `json:"notify_url"`
	RedirectURL string  `json:"redirect_url"`
	Signature   string  `json:"signature"`
	Currencies  string  `json:"currencies,omitempty"`
	Fiat        string  `json:"fiat,omitempty"`
	Name        string  `json:"name,omitempty"`
	Timeout     int     `json:"timeout,omitempty"`
}

type CallbackData struct {
	TradeID            string  `json:"trade_id"`
	OrderID            string  `json:"order_id"`
	Amount             float64 `json:"amount"`
	ActualAmount       float64 `json:"actual_amount"`
	Token              string  `json:"token"`
	BlockTransactionID string  `json:"block_transaction_id"`
	Signature          string  `json:"signature"`
	Status             int     `json:"status"`
}

// CreateTransaction creates a USDT payment transaction
func (c *Client) CreateTransaction(orderID string, amount float64, name string) (*TransactionResponse, error) {
	params := map[string]string{
		"order_id":     orderID,
		"amount":       fmt.Sprintf("%.2f", amount),
		"notify_url":   c.notifyURL,
		"redirect_url": c.redirectURL,
	}
	signature := c.generateSignature(params)

	req := CreateTransactionRequest{
		OrderID:     orderID,
		Amount:      amount,
		NotifyURL:   c.notifyURL,
		RedirectURL: c.redirectURL,
		Signature:   signature,
		TradeType:   "usdt.trc20",
		Fiat:        "CNY",
		Name:        name,
		Timeout:     600,
	}

	data, err := c.doPost("/api/v1/order/create-transaction", req)
	if err != nil {
		return nil, err
	}

	var resp TransactionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return &resp, fmt.Errorf("BEPusdt error: %s", resp.Message)
	}
	return &resp, nil
}

// CreateOrder creates an order with multiple payment method options
func (c *Client) CreateOrder(orderID string, amount float64, name string) (*TransactionResponse, error) {
	params := map[string]string{
		"order_id":     orderID,
		"amount":       fmt.Sprintf("%.2f", amount),
		"notify_url":   c.notifyURL,
		"redirect_url": c.redirectURL,
	}
	signature := c.generateSignature(params)

	req := CreateOrderRequest{
		OrderID:     orderID,
		Amount:      amount,
		NotifyURL:   c.notifyURL,
		RedirectURL: c.redirectURL,
		Signature:   signature,
		Fiat:        "CNY",
		Name:        name,
		Timeout:     600,
	}

	data, err := c.doPost("/api/v1/order/create-order", req)
	if err != nil {
		return nil, err
	}

	var resp TransactionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return &resp, fmt.Errorf("BEPusdt error: %s", resp.Message)
	}
	return &resp, nil
}

// CancelTransaction cancels a transaction by trade ID
func (c *Client) CancelTransaction(tradeID string) error {
	params := map[string]string{
		"trade_id": tradeID,
	}
	signature := c.generateSignature(params)

	body := map[string]string{
		"trade_id":  tradeID,
		"signature": signature,
	}

	data, err := c.doPost("/api/v1/order/cancel-transaction", body)
	if err != nil {
		return err
	}

	var resp struct {
		StatusCode int    `json:"status_code"`
		Message    string `json:"message"`
	}
	json.Unmarshal(data, &resp)
	if resp.StatusCode != 200 {
		return fmt.Errorf("cancel failed: %s", resp.Message)
	}
	return nil
}

// VerifyCallback verifies a payment callback signature
func (c *Client) VerifyCallback(data *CallbackData) bool {
	params := map[string]string{
		"trade_id":             data.TradeID,
		"order_id":             data.OrderID,
		"amount":               fmt.Sprintf("%v", data.Amount),
		"actual_amount":        fmt.Sprintf("%v", data.ActualAmount),
		"token":                data.Token,
		"block_transaction_id": data.BlockTransactionID,
		"status":               fmt.Sprintf("%d", data.Status),
	}
	expected := c.generateSignature(params)
	return expected == data.Signature
}

// generateSignature generates the MD5 signature per BEPusdt spec:
// 1. Filter non-empty, non-signature params
// 2. Sort by key ASCII
// 3. Join as key=value with &
// 4. Append API token directly (no &) and MD5 hash (lowercase)
func (c *Client) generateSignature(params map[string]string) string {
	var keys []string
	for k, v := range params {
		if k == "signature" || v == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	str := strings.Join(parts, "&") + c.token
	hash := md5.Sum([]byte(str))
	return fmt.Sprintf("%x", hash)
}

func (c *Client) doPost(path string, body interface{}) ([]byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
