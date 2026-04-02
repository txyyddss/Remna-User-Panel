package ezpay

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Client is the EZPay (MotionPay) API client
type Client struct {
	baseURL    string
	pid        int
	key        string
	notifyURL  string
	returnURL  string
	httpClient *http.Client
}

// NewClient creates a new EZPay client
func NewClient(baseURL string, pid int, key, notifyURL, returnURL string) *Client {
	return &Client{
		baseURL:   strings.TrimRight(baseURL, "/"),
		pid:       pid,
		key:       key,
		notifyURL: notifyURL,
		returnURL: returnURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// PaymentResponse is the EZPay API payment response
type PaymentResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	TradeNo   string `json:"trade_no"`
	PayURL    string `json:"payurl"`
	QRCode    string `json:"qrcode"`
	URLScheme string `json:"urlscheme"`
}

// CreatePayment creates a payment via the API interface (mapi.php)
func (c *Client) CreatePayment(outTradeNo, payType, name, money, clientIP string) (*PaymentResponse, error) {
	params := map[string]string{
		"pid":          fmt.Sprintf("%d", c.pid),
		"type":         payType,
		"out_trade_no": outTradeNo,
		"notify_url":   c.notifyURL,
		"return_url":   c.returnURL,
		"name":         name,
		"money":        money,
		"clientip":     clientIP,
		"sign_type":    "MD5",
	}
	params["sign"] = c.generateSign(params)

	form := url.Values{}
	for k, v := range params {
		form.Set(k, v)
	}

	resp, err := c.httpClient.PostForm(c.baseURL+"/mapi.php", form)
	if err != nil {
		return nil, fmt.Errorf("post payment: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result PaymentResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w (body: %s)", err, string(body))
	}

	if result.Code != 1 {
		return &result, fmt.Errorf("payment failed: %s", result.Msg)
	}
	return &result, nil
}

// GetPaymentURL generates a redirect URL for page payment (submit.php)
func (c *Client) GetPaymentURL(outTradeNo, payType, name, money string) string {
	params := map[string]string{
		"pid":          fmt.Sprintf("%d", c.pid),
		"type":         payType,
		"out_trade_no": outTradeNo,
		"notify_url":   c.notifyURL,
		"return_url":   c.returnURL,
		"name":         name,
		"money":        money,
		"sign_type":    "MD5",
	}
	params["sign"] = c.generateSign(params)

	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return c.baseURL + "/submit.php?" + values.Encode()
}

// VerifyCallback verifies a payment callback signature
func (c *Client) VerifyCallback(params map[string]string) bool {
	receivedSign := params["sign"]
	if receivedSign == "" {
		return false
	}

	verifyParams := make(map[string]string)
	for k, v := range params {
		if k != "sign" && k != "sign_type" {
			verifyParams[k] = v
		}
	}
	return c.generateSign(verifyParams) == receivedSign
}

// QueryOrder queries a single order
func (c *Client) QueryOrder(outTradeNo string) ([]byte, error) {
	params := map[string]string{
		"act":          "order",
		"pid":          fmt.Sprintf("%d", c.pid),
		"out_trade_no": outTradeNo,
		"sign_type":    "MD5",
	}
	params["sign"] = c.generateSign(params)

	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	resp, err := c.httpClient.Get(c.baseURL + "/api.php?" + values.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// generateSign generates the MD5 signature per EZPay spec:
// 1. Sort params by key ASCII; exclude sign, sign_type, empty and "0" values
// 2. Join as key=value with &
// 3. Append merchant key directly (no &) and MD5 hash
func (c *Client) generateSign(params map[string]string) string {
	var keys []string
	for k, v := range params {
		if k == "sign" || k == "sign_type" || v == "" || v == "0" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	str := strings.Join(parts, "&") + c.key
	hash := md5.Sum([]byte(str))
	return fmt.Sprintf("%x", hash)
}
