package bepusdt

import (
	"crypto/md5" // #nosec G501 -- MD5 is mandated by the BEPusdt wire protocol.
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ErrInvalidSignature means a BEPusdt callback did not authenticate.
var ErrInvalidSignature = errors.New("bepusdt signature is invalid")

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
