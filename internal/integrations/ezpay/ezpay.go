package ezpay

import (
	"crypto/md5" // #nosec G501 -- MD5 is mandated by the EZPay wire protocol.
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var (
	// ErrInvalidSignature means an EZPay callback did not authenticate.
	ErrInvalidSignature = errors.New("ezpay signature is invalid")
	decimalPattern      = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]+)?$`)
)

// PaymentType identifies an EZPay payment rail.
type PaymentType string

const (
	// PaymentAlipay selects Alipay.
	PaymentAlipay PaymentType = "alipay"
	// PaymentWeChat selects WeChat Pay.
	PaymentWeChat PaymentType = "wxpay"
	// PaymentQQ selects QQ Wallet.
	PaymentQQ PaymentType = "qqpay"
	// PaymentBank selects UnionPay/cloud flash payment.
	PaymentBank PaymentType = "bank"
	// PaymentJD selects JD Pay.
	PaymentJD PaymentType = "jdpay"
)

// CheckoutRequest contains the signed submit.php parameters used by the supplied SDK.
type CheckoutRequest struct {
	Type       PaymentType
	NotifyURL  string
	ReturnURL  string
	OutTradeNo string
	Name       string
	Money      string
}

// Notification is a verified EZPay GET notification.
type Notification struct {
	MerchantID  string
	OutTradeNo  string
	TradeNo     string
	TradeStatus string
	Type        PaymentType
	Money       string
}

// Successful reports whether EZPay marked the transaction paid.
func (n Notification) Successful() bool {
	return n.TradeStatus == "TRADE_SUCCESS"
}

// Client owns an EZPay merchant's checkout URL and signing key.
type Client struct {
	baseURL    *url.URL
	merchantID string
	key        string
}

// NewClient creates an EZPay checkout/signature client.
func NewClient(rawBaseURL, merchantID, key string) (*Client, error) {
	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, fmt.Errorf("ezpay parse base URL: %w", err)
	}
	if baseURL.Scheme != "https" && baseURL.Scheme != "http" {
		return nil, errors.New("ezpay base URL must use http or https")
	}
	if baseURL.Host == "" || baseURL.User != nil || baseURL.RawQuery != "" || baseURL.Fragment != "" {
		return nil, errors.New("ezpay base URL must be absolute and contain no credentials, query, or fragment")
	}
	if strings.TrimSpace(merchantID) == "" || key == "" {
		return nil, errors.New("ezpay merchant id and key are required")
	}
	return &Client{baseURL: baseURL, merchantID: merchantID, key: key}, nil
}

// CheckoutURL returns the signed submit.php URL for a payment order.
func (c *Client) CheckoutURL(input CheckoutRequest) (string, error) {
	if strings.TrimSpace(string(input.Type)) == "" || strings.TrimSpace(input.OutTradeNo) == "" || strings.TrimSpace(input.Name) == "" {
		return "", errors.New("ezpay checkout requires payment type, order number, and name")
	}
	if err := validateCallbackURL(input.NotifyURL); err != nil {
		return "", fmt.Errorf("ezpay notify URL: %w", err)
	}
	if err := validateCallbackURL(input.ReturnURL); err != nil {
		return "", fmt.Errorf("ezpay return URL: %w", err)
	}
	if err := validatePositiveDecimal(input.Money); err != nil {
		return "", fmt.Errorf("ezpay money: %w", err)
	}
	values := url.Values{
		"pid":          {c.merchantID},
		"type":         {string(input.Type)},
		"notify_url":   {input.NotifyURL},
		"return_url":   {input.ReturnURL},
		"out_trade_no": {input.OutTradeNo},
		"name":         {input.Name},
		"money":        {input.Money},
	}
	values.Set("sign", Sign(values, c.key))
	values.Set("sign_type", "MD5")
	target := *c.baseURL
	target.Path = strings.TrimRight(target.Path, "/") + "/submit.php"
	target.RawQuery = values.Encode()
	return target.String(), nil
}

// Verify authenticates a callback using the supplied SDK's sorted MD5 suffix signature.
func (c *Client) Verify(values url.Values) error {
	if len(values) == 0 {
		return errors.New("ezpay callback is empty")
	}
	for key, entries := range values {
		if len(entries) != 1 {
			return fmt.Errorf("ezpay callback contains duplicate field %q", key)
		}
	}
	provided := strings.ToLower(values.Get("sign"))
	if len(provided) != md5.Size*2 {
		return ErrInvalidSignature
	}
	expected := Sign(values, c.key)
	if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
		return ErrInvalidSignature
	}
	return nil
}

// ParseNotification verifies and parses a GET callback. Unknown signed fields are accepted.
func (c *Client) ParseNotification(values url.Values) (*Notification, error) {
	if err := c.Verify(values); err != nil {
		return nil, err
	}
	notification := &Notification{
		MerchantID:  values.Get("pid"),
		OutTradeNo:  values.Get("out_trade_no"),
		TradeNo:     values.Get("trade_no"),
		TradeStatus: values.Get("trade_status"),
		Type:        PaymentType(values.Get("type")),
		Money:       values.Get("money"),
	}
	if notification.OutTradeNo == "" || notification.TradeNo == "" || notification.TradeStatus == "" || notification.Type == "" {
		return nil, errors.New("ezpay callback is missing required transaction fields")
	}
	if notification.MerchantID != "" && notification.MerchantID != c.merchantID {
		return nil, errors.New("ezpay callback merchant id does not match")
	}
	if err := validatePositiveDecimal(notification.Money); err != nil {
		return nil, fmt.Errorf("ezpay callback money: %w", err)
	}
	return notification, nil
}

// Sign implements the supplied EZPay SDK algorithm: sorted non-empty key=value pairs,
// excluding sign and sign_type, followed immediately by the merchant key and lower-case MD5.
func Sign(values url.Values, key string) string {
	keys := make([]string, 0, len(values))
	for name := range values {
		if name == "sign" || name == "sign_type" || values.Get(name) == "" {
			continue
		}
		keys = append(keys, name)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, name := range keys {
		parts = append(parts, name+"="+values.Get(name))
	}
	sum := md5.Sum([]byte(strings.Join(parts, "&") + key)) // #nosec G401 -- required for protocol compatibility.
	return hex.EncodeToString(sum[:])
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

func validatePositiveDecimal(value string) error {
	if len(value) > 64 {
		return errors.New("value is too long")
	}
	if !decimalPattern.MatchString(value) {
		return errors.New("value must be an unsigned base-10 decimal")
	}
	parsed, ok := new(big.Rat).SetString(value)
	if !ok || parsed.Sign() <= 0 {
		return errors.New("value must be positive")
	}
	return nil
}
