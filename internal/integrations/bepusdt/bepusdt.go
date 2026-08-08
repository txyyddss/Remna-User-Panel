package bepusdt

import (
	"bytes"
	"context"
	"crypto/md5" // #nosec G501 -- MD5 is mandated by the BEPusdt wire protocol.
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const maxResponseBytes = 2 << 20

var (
	// ErrInvalidSignature means a BEPusdt callback did not authenticate.
	ErrInvalidSignature = errors.New("bepusdt signature is invalid")
	decimalPattern      = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?$`)
)

// HTTPDoer is implemented by *http.Client and permits transport injection in tests.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// APIError describes a sanitized BEPusdt HTTP or application error.
type APIError struct {
	HTTPStatus int
	StatusCode int
	Message    string
	RequestID  string
}

// Error implements error.
func (e *APIError) Error() string {
	if e == nil {
		return "bepusdt API error"
	}
	return fmt.Sprintf("bepusdt API failed (http=%d api=%d): %s", e.HTTPStatus, e.StatusCode, e.Message)
}

// CreateTransactionRequest is the documented /api/v1/order/create-transaction request.
// Amount is an exact base-10 fiat amount and is emitted as a JSON number.
type CreateTransactionRequest struct {
	OrderID        string
	Amount         string
	Fiat           string
	TradeType      string
	Address        string
	Name           string
	NotifyURL      string
	RedirectURL    string
	TimeoutSeconds int64
	Rate           string
}

// Transaction is the payment instruction returned by BEPusdt.
type Transaction struct {
	Fiat           string
	TradeID        string
	OrderID        string
	Amount         string
	ActualAmount   string
	Status         int
	Token          string
	ExpirationTime int64
	PaymentURL     string
}

// ExpiresAt converts BEPusdt's expiration_time duration (seconds) into an instant.
func (t Transaction) ExpiresAt(createdAt time.Time) time.Time {
	return createdAt.Add(time.Duration(t.ExpirationTime) * time.Second)
}

// Webhook is a parsed BEPusdt order notification. Amount strings retain exact decimals.
type Webhook struct {
	TradeID            string
	OrderID            string
	Amount             string
	ActualAmount       string
	Token              string
	BlockTransactionID string
	Status             int
	CreatedAt          string
	ExpiredAt          string
	Signature          string
}

// Paid reports whether BEPusdt marked the order paid.
func (w Webhook) Paid() bool {
	return w.Status == 2
}

// Client calls a BEPusdt server.
type Client struct {
	baseURL    *url.URL
	token      string
	httpClient HTTPDoer
}

// Option configures a Client.
type Option func(*Client) error

// WithHTTPClient installs a custom HTTP transport. Production transports must retain TLS verification.
func WithHTTPClient(client HTTPDoer) Option {
	return func(c *Client) error {
		if client == nil {
			return errors.New("bepusdt HTTP client is nil")
		}
		c.httpClient = client
		return nil
	}
}

// NewClient creates a BEPusdt API client.
func NewClient(rawBaseURL, token string, options ...Option) (*Client, error) {
	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, fmt.Errorf("bepusdt parse base URL: %w", err)
	}
	if baseURL.Scheme != "https" && baseURL.Scheme != "http" {
		return nil, errors.New("bepusdt base URL must use http or https")
	}
	if baseURL.Host == "" || baseURL.User != nil || baseURL.RawQuery != "" || baseURL.Fragment != "" {
		return nil, errors.New("bepusdt base URL must be absolute and contain no credentials, query, or fragment")
	}
	if token == "" {
		return nil, errors.New("bepusdt API token is empty")
	}
	c := &Client{baseURL: baseURL, token: token, httpClient: &http.Client{
		Timeout:       20 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(c); err != nil {
			return nil, fmt.Errorf("configure bepusdt client: %w", err)
		}
	}
	return c, nil
}

// CreateTransaction creates a fixed-fiat BEPusdt transaction.
func (c *Client) CreateTransaction(ctx context.Context, input CreateTransactionRequest) (*Transaction, error) {
	if strings.TrimSpace(input.OrderID) == "" {
		return nil, errors.New("bepusdt order id is empty")
	}
	if err := validatePositiveDecimal(input.Amount); err != nil {
		return nil, fmt.Errorf("bepusdt amount: %w", err)
	}
	if input.Fiat == "" {
		input.Fiat = "USD"
	}
	if input.TradeType == "" {
		input.TradeType = "usdt.trc20"
	}
	if err := validateCallbackURL(input.NotifyURL); err != nil {
		return nil, fmt.Errorf("bepusdt notify URL: %w", err)
	}
	if err := validateCallbackURL(input.RedirectURL); err != nil {
		return nil, fmt.Errorf("bepusdt redirect URL: %w", err)
	}
	if input.TimeoutSeconds != 0 && input.TimeoutSeconds < 120 {
		return nil, errors.New("bepusdt timeout must be at least 120 seconds")
	}

	values := map[string]string{
		"order_id":     input.OrderID,
		"amount":       input.Amount,
		"fiat":         input.Fiat,
		"trade_type":   input.TradeType,
		"notify_url":   input.NotifyURL,
		"redirect_url": input.RedirectURL,
		"address":      input.Address,
		"name":         input.Name,
		"rate":         input.Rate,
	}
	if input.TimeoutSeconds > 0 {
		values["timeout"] = strconv.FormatInt(input.TimeoutSeconds, 10)
	}
	wire := map[string]any{
		"order_id":     input.OrderID,
		"amount":       json.Number(input.Amount),
		"fiat":         input.Fiat,
		"trade_type":   input.TradeType,
		"notify_url":   input.NotifyURL,
		"redirect_url": input.RedirectURL,
		"signature":    Sign(values, c.token),
	}
	for _, key := range []string{"address", "name", "rate"} {
		if values[key] != "" {
			wire[key] = values[key]
		}
	}
	if input.TimeoutSeconds > 0 {
		wire["timeout"] = input.TimeoutSeconds
	}

	encoded, err := json.Marshal(wire)
	if err != nil {
		return nil, fmt.Errorf("bepusdt encode request: %w", err)
	}
	target := *c.baseURL
	target.Path = strings.TrimRight(target.Path, "/") + "/api/v1/order/create-transaction"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), bytes.NewReader(encoded))
	if err != nil {
		return nil, fmt.Errorf("bepusdt create request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("bepusdt request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("bepusdt read response: %w", err)
	}
	if len(responseBody) > maxResponseBytes {
		return nil, fmt.Errorf("bepusdt response exceeds %d bytes", maxResponseBytes)
	}

	var envelope struct {
		StatusCode flexibleInt `json:"status_code"`
		Message    string      `json:"message"`
		RequestID  string      `json:"request_id"`
		Data       struct {
			Fiat           string         `json:"fiat"`
			TradeID        string         `json:"trade_id"`
			OrderID        string         `json:"order_id"`
			Amount         flexibleString `json:"amount"`
			ActualAmount   flexibleString `json:"actual_amount"`
			Status         flexibleInt    `json:"status"`
			Token          string         `json:"token"`
			ExpirationTime flexibleInt    `json:"expiration_time"`
			PaymentURL     string         `json:"payment_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(responseBody, &envelope); err != nil {
		return nil, fmt.Errorf("bepusdt decode response (http=%d): %w", response.StatusCode, err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices || int(envelope.StatusCode) != http.StatusOK {
		return nil, &APIError{HTTPStatus: response.StatusCode, StatusCode: int(envelope.StatusCode), Message: envelope.Message, RequestID: envelope.RequestID}
	}
	transaction := &Transaction{
		Fiat: envelope.Data.Fiat, TradeID: envelope.Data.TradeID, OrderID: envelope.Data.OrderID,
		Amount: string(envelope.Data.Amount), ActualAmount: string(envelope.Data.ActualAmount),
		Status: int(envelope.Data.Status), Token: envelope.Data.Token,
		ExpirationTime: int64(envelope.Data.ExpirationTime), PaymentURL: envelope.Data.PaymentURL,
	}
	if transaction.TradeID == "" || transaction.OrderID == "" || transaction.Token == "" || transaction.PaymentURL == "" {
		return nil, errors.New("bepusdt success response is missing payment data")
	}
	if transaction.OrderID != input.OrderID || !strings.EqualFold(transaction.Fiat, input.Fiat) || !decimalEquivalent(transaction.Amount, input.Amount) {
		return nil, errors.New("bepusdt success response does not match the requested order")
	}
	if err := validatePositiveDecimal(transaction.ActualAmount); err != nil {
		return nil, fmt.Errorf("bepusdt success response actual amount: %w", err)
	}
	if transaction.ExpirationTime <= 0 {
		return nil, errors.New("bepusdt success response expiration is invalid")
	}
	if err := validateHTTPSPaymentURL(transaction.PaymentURL); err != nil {
		return nil, fmt.Errorf("bepusdt success response payment URL: %w", err)
	}
	return transaction, nil
}

// CancelTransaction calls BEPusdt's signed cancellation endpoint for a direct
// transaction. A paid notification may race with this operation and remains
// authoritative to the caller.
func (c *Client) CancelTransaction(ctx context.Context, tradeID string) error {
	tradeID = strings.TrimSpace(tradeID)
	if tradeID == "" {
		return errors.New("bepusdt trade id is empty")
	}
	encoded, err := json.Marshal(map[string]string{
		"trade_id":  tradeID,
		"signature": Sign(map[string]string{"trade_id": tradeID}, c.token),
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

// ParseWebhook parses a BEPusdt JSON callback while accepting string or numeric scalar fields.
// Duplicate keys and nested values are rejected.
func ParseWebhook(body []byte) (*Webhook, map[string]string, error) {
	raw, err := decodeUniqueObject(body)
	if err != nil {
		return nil, nil, err
	}
	values := make(map[string]string, len(raw))
	for key, value := range raw {
		scalar, err := scalarString(value)
		if err != nil {
			return nil, nil, fmt.Errorf("bepusdt webhook field %q: %w", key, err)
		}
		values[key] = scalar
	}
	status, err := strconv.Atoi(values["status"])
	if err != nil {
		return nil, nil, errors.New("bepusdt webhook status is invalid")
	}
	webhook := &Webhook{
		TradeID: values["trade_id"], OrderID: values["order_id"], Amount: values["amount"],
		ActualAmount: values["actual_amount"], Token: values["token"],
		BlockTransactionID: values["block_transaction_id"], Status: status,
		CreatedAt: values["created_at"], ExpiredAt: values["expired_at"], Signature: values["signature"],
	}
	if webhook.TradeID == "" || webhook.OrderID == "" || webhook.Amount == "" || webhook.ActualAmount == "" || webhook.Signature == "" {
		return nil, nil, errors.New("bepusdt webhook is missing required fields")
	}
	if err := validateDecimal(webhook.Amount); err != nil {
		return nil, nil, fmt.Errorf("bepusdt webhook amount: %w", err)
	}
	if err := validateDecimal(webhook.ActualAmount); err != nil {
		return nil, nil, fmt.Errorf("bepusdt webhook actual amount: %w", err)
	}
	return webhook, values, nil
}

// ParseUnsignedWebhook parses the v1.19 direct-transaction notification shape.
// Authentication is intentionally left to the per-order callback URL capability.
func ParseUnsignedWebhook(body []byte) (*Webhook, error) {
	raw, err := decodeUniqueObject(body)
	if err != nil {
		return nil, err
	}
	values := make(map[string]string, len(raw))
	for key, value := range raw {
		scalar, scalarErr := scalarString(value)
		if scalarErr != nil {
			return nil, fmt.Errorf("bepusdt webhook field %q: %w", key, scalarErr)
		}
		values[key] = scalar
	}
	if values["signature"] != "" {
		return nil, errors.New("signed callback must use signature verification")
	}
	status, err := strconv.Atoi(values["status"])
	if err != nil {
		return nil, errors.New("bepusdt webhook status is invalid")
	}
	webhook := &Webhook{
		TradeID: values["trade_id"], OrderID: values["order_id"], Amount: values["amount"],
		ActualAmount: values["actual_amount"], Token: values["token"],
		BlockTransactionID: values["block_transaction_id"], Status: status,
		CreatedAt: values["created_at"], ExpiredAt: values["expired_at"],
	}
	if webhook.OrderID == "" || webhook.Amount == "" || webhook.ActualAmount == "" || webhook.Token == "" {
		return nil, errors.New("bepusdt webhook is missing required fields")
	}
	if err := validateDecimal(webhook.Amount); err != nil {
		return nil, fmt.Errorf("bepusdt webhook amount: %w", err)
	}
	if err := validateDecimal(webhook.ActualAmount); err != nil {
		return nil, fmt.Errorf("bepusdt webhook actual amount: %w", err)
	}
	return webhook, nil
}

// VerifyWebhook verifies a parsed webhook's complete scalar field set.
func VerifyWebhook(values map[string]string, token string) error {
	provided := strings.ToLower(values["signature"])
	if len(provided) != md5.Size*2 {
		return ErrInvalidSignature
	}
	expected := Sign(values, token)
	if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
		return ErrInvalidSignature
	}
	return nil
}

// ParseAndVerifyWebhook authenticates and parses a BEPusdt callback in one step.
func (c *Client) ParseAndVerifyWebhook(body []byte) (*Webhook, error) {
	webhook, values, err := ParseWebhook(body)
	if err != nil {
		return nil, err
	}
	if err := VerifyWebhook(values, c.token); err != nil {
		return nil, err
	}
	return webhook, nil
}

// Sign implements BEPusdt's sorted non-empty key=value MD5 suffix signature.
func Sign(values map[string]string, token string) string {
	keys := make([]string, 0, len(values))
	for key, value := range values {
		if key == "signature" || value == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+values[key])
	}
	sum := md5.Sum([]byte(strings.Join(parts, "&") + token)) // #nosec G401 -- required for protocol compatibility.
	return hex.EncodeToString(sum[:])
}

type flexibleString string

func (s *flexibleString) UnmarshalJSON(data []byte) error {
	value, err := scalarString(data)
	if err != nil {
		return err
	}
	*s = flexibleString(value)
	return nil
}

type flexibleInt int64

func (i *flexibleInt) UnmarshalJSON(data []byte) error {
	value, err := scalarString(data)
	if err != nil {
		return err
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("parse integer: %w", err)
	}
	*i = flexibleInt(parsed)
	return nil
}

func scalarString(data []byte) (string, error) {
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) {
		return "", nil
	}
	if len(trimmed) > 0 && trimmed[0] == '"' {
		var value string
		if err := json.Unmarshal(trimmed, &value); err != nil {
			return "", err
		}
		return value, nil
	}
	var number json.Number
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.UseNumber()
	if err := decoder.Decode(&number); err == nil {
		if _, ok := new(big.Rat).SetString(number.String()); !ok {
			return "", errors.New("invalid numeric value")
		}
		return number.String(), nil
	}
	if bytes.Equal(trimmed, []byte("true")) || bytes.Equal(trimmed, []byte("false")) {
		return string(trimmed), nil
	}
	return "", errors.New("value must be a JSON scalar")
}

func decodeUniqueObject(body []byte) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	opening, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("bepusdt decode webhook: %w", err)
	}
	if delimiter, ok := opening.(json.Delim); !ok || delimiter != '{' {
		return nil, errors.New("bepusdt webhook must be a JSON object")
	}
	result := make(map[string]json.RawMessage)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, fmt.Errorf("bepusdt decode webhook key: %w", err)
		}
		key, ok := token.(string)
		if !ok {
			return nil, errors.New("bepusdt webhook contains a non-string key")
		}
		if _, exists := result[key]; exists {
			return nil, fmt.Errorf("bepusdt webhook contains duplicate field %q", key)
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, fmt.Errorf("bepusdt decode webhook field %q: %w", key, err)
		}
		result[key] = value
	}
	if _, err := decoder.Token(); err != nil {
		return nil, fmt.Errorf("bepusdt decode webhook end: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, errors.New("bepusdt webhook contains trailing JSON")
		}
		return nil, fmt.Errorf("bepusdt decode trailing data: %w", err)
	}
	return result, nil
}

func validateCallbackURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.User != nil {
		return errors.New("URL must be absolute http(s) without credentials")
	}
	return nil
}

func validateHTTPSPaymentURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if u.Scheme != "https" || u.Host == "" || u.User != nil {
		return errors.New("URL must be absolute HTTPS without credentials")
	}
	return nil
}

func validatePositiveDecimal(value string) error {
	if err := validateDecimal(value); err != nil {
		return err
	}
	parsed, _ := new(big.Rat).SetString(value)
	if parsed.Sign() <= 0 {
		return errors.New("value must be positive")
	}
	return nil
}

func validateDecimal(value string) error {
	if len(value) > 64 {
		return errors.New("value is too long")
	}
	if !decimalPattern.MatchString(value) {
		return errors.New("value must be a base-10 decimal")
	}
	if _, ok := new(big.Rat).SetString(value); !ok {
		return errors.New("value is outside supported decimal syntax")
	}
	return nil
}

func decimalEquivalent(left, right string) bool {
	leftValue, leftOK := new(big.Rat).SetString(left)
	rightValue, rightOK := new(big.Rat).SetString(right)
	return leftOK && rightOK && leftValue.Cmp(rightValue) == 0
}
