package telegram

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxResponseBytes = 2 << 20

// WebhookSecretHeader is the header Telegram uses for configured webhook secrets.
const WebhookSecretHeader = "X-Telegram-Bot-Api-Secret-Token"

// HTTPDoer is implemented by *http.Client and permits transport injection in tests.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// APIError is a sanitized Telegram Bot API failure. It never contains the bot token.
type APIError struct {
	Method      string
	HTTPStatus  int
	ErrorCode   int
	Description string
	RetryAfter  int
}

// Error implements error.
func (e *APIError) Error() string {
	if e == nil {
		return "telegram API error"
	}
	return fmt.Sprintf("telegram %s failed (http=%d api=%d): %s", e.Method, e.HTTPStatus, e.ErrorCode, e.Description)
}

// Client calls the Telegram Bot API.
type Client struct {
	token      string
	baseURL    *url.URL
	httpClient HTTPDoer
}

// Option configures a Client.
type Option func(*Client) error

// WithBaseURL overrides the Telegram API origin, primarily for a local Bot API server or tests.
func WithBaseURL(rawURL string) Option {
	return func(c *Client) error {
		u, err := parseBaseURL(rawURL)
		if err != nil {
			return err
		}
		c.baseURL = u
		return nil
	}
}

// WithHTTPClient installs a custom HTTP transport. TLS policy remains the transport's responsibility.
func WithHTTPClient(client HTTPDoer) Option {
	return func(c *Client) error {
		if client == nil {
			return errors.New("telegram HTTP client is nil")
		}
		c.httpClient = client
		return nil
	}
}

// NewClient creates a Telegram Bot API client using certificate-verified HTTPS defaults.
func NewClient(token string, options ...Option) (*Client, error) {
	if !validBotToken(token) {
		return nil, errors.New("telegram bot token is empty or malformed")
	}
	baseURL, _ := url.Parse("https://api.telegram.org")
	c := &Client{token: token, baseURL: baseURL, httpClient: &http.Client{
		Timeout:       20 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(c); err != nil {
			return nil, fmt.Errorf("configure telegram client: %w", err)
		}
	}
	return c, nil
}

// WebhookConfig describes Telegram webhook setup.
type WebhookConfig struct {
	URL                string   `json:"url"`
	SecretToken        string   `json:"secret_token,omitempty"`
	AllowedUpdates     []string `json:"allowed_updates,omitempty"`
	DropPendingUpdates bool     `json:"drop_pending_updates,omitempty"`
	MaxConnections     int      `json:"max_connections,omitempty"`
}

// DefaultAllowedUpdates returns the update kinds required by onboarding and Stars payments.
func DefaultAllowedUpdates() []string {
	return []string{"message", "chat_member", "chat_join_request", "pre_checkout_query"}
}

// VerifyWebhookSecret compares a received webhook header with the configured secret in constant time.
func VerifyWebhookSecret(provided, expected string) bool {
	if provided == "" || expected == "" || len(provided) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

// SetWebhook configures Telegram webhook delivery.
func (c *Client) SetWebhook(ctx context.Context, config WebhookConfig) error {
	if err := validateHTTPSURL(config.URL); err != nil {
		return fmt.Errorf("telegram webhook URL: %w", err)
	}
	if config.SecretToken != "" && !validWebhookSecret(config.SecretToken) {
		return errors.New("telegram webhook secret must contain 1-256 ASCII letters, digits, underscores, or hyphens")
	}
	if config.MaxConnections < 0 || config.MaxConnections > 100 {
		return errors.New("telegram webhook max connections must be in 1..100 when set")
	}
	var result bool
	if err := c.call(ctx, "setWebhook", config, &result); err != nil {
		return err
	}
	if !result {
		return errors.New("telegram setWebhook returned false")
	}
	return nil
}

// SetChatMenuButton configures the default private-chat button to open the Mini App.
func (c *Client) SetChatMenuButton(ctx context.Context, text, webAppURL string) error {
	if strings.TrimSpace(text) == "" {
		return errors.New("telegram menu button text is empty")
	}
	if err := validateHTTPSURL(webAppURL); err != nil {
		return fmt.Errorf("telegram menu Web App URL: %w", err)
	}
	payload := struct {
		MenuButton struct {
			Type   string `json:"type"`
			Text   string `json:"text"`
			WebApp struct {
				URL string `json:"url"`
			} `json:"web_app"`
		} `json:"menu_button"`
	}{}
	payload.MenuButton.Type = "web_app"
	payload.MenuButton.Text = text
	payload.MenuButton.WebApp.URL = webAppURL
	var result bool
	if err := c.call(ctx, "setChatMenuButton", payload, &result); err != nil {
		return err
	}
	if !result {
		return errors.New("telegram setChatMenuButton returned false")
	}
	return nil
}

// CreateJoinRequestInvite creates a revocable invite which requires administrator approval.
func (c *Client) CreateJoinRequestInvite(ctx context.Context, chatID, name string, expireAt time.Time) (*ChatInviteLink, error) {
	if strings.TrimSpace(chatID) == "" {
		return nil, errors.New("telegram chat id is empty")
	}
	if expireAt.IsZero() {
		return nil, errors.New("telegram invite expiration is empty")
	}
	if len([]rune(name)) > 32 {
		return nil, errors.New("telegram invite name exceeds 32 characters")
	}
	payload := struct {
		ChatID             string `json:"chat_id"`
		Name               string `json:"name,omitempty"`
		ExpireDate         int64  `json:"expire_date"`
		CreatesJoinRequest bool   `json:"creates_join_request"`
	}{ChatID: chatID, Name: name, ExpireDate: expireAt.Unix(), CreatesJoinRequest: true}
	var result ChatInviteLink
	if err := c.call(ctx, "createChatInviteLink", payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ApproveJoinRequest approves a single Telegram user for a chat.
func (c *Client) ApproveJoinRequest(ctx context.Context, chatID string, userID int64) error {
	return c.booleanCall(ctx, "approveChatJoinRequest", memberRequest{ChatID: chatID, UserID: userID})
}

// RevokeInviteLink revokes a bot-created chat invite link.
func (c *Client) RevokeInviteLink(ctx context.Context, chatID, inviteLink string) (*ChatInviteLink, error) {
	if strings.TrimSpace(chatID) == "" || strings.TrimSpace(inviteLink) == "" {
		return nil, errors.New("telegram chat id and invite link are required")
	}
	payload := struct {
		ChatID     string `json:"chat_id"`
		InviteLink string `json:"invite_link"`
	}{ChatID: chatID, InviteLink: inviteLink}
	var result ChatInviteLink
	if err := c.call(ctx, "revokeChatInviteLink", payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetChatMember fetches canonical membership state for a user.
func (c *Client) GetChatMember(ctx context.Context, chatID string, userID int64) (*ChatMember, error) {
	if err := validateMemberRequest(chatID, userID); err != nil {
		return nil, err
	}
	var result ChatMember
	if err := c.call(ctx, "getChatMember", memberRequest{ChatID: chatID, UserID: userID}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

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

type memberRequest struct {
	ChatID string `json:"chat_id"`
	UserID int64  `json:"user_id"`
}

func (c *Client) booleanCall(ctx context.Context, method string, payload any) error {
	if request, ok := payload.(memberRequest); ok {
		if err := validateMemberRequest(request.ChatID, request.UserID); err != nil {
			return err
		}
	}
	var result bool
	if err := c.call(ctx, method, payload, &result); err != nil {
		return err
	}
	if !result {
		return fmt.Errorf("telegram %s returned false", method)
	}
	return nil
}

func validateMemberRequest(chatID string, userID int64) error {
	if strings.TrimSpace(chatID) == "" || userID <= 0 {
		return errors.New("telegram chat id and positive user id are required")
	}
	return nil
}

func (c *Client) call(ctx context.Context, method string, payload, result any) error {
	if ctx == nil {
		return fmt.Errorf("telegram %s: context is nil", method)
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("telegram %s encode request: %w", method, err)
	}
	endpoint := *c.baseURL
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/bot" + c.token + "/" + method
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram %s create request: %w", method, err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(request)
	if err != nil {
		var urlError *url.Error
		if errors.As(err, &urlError) {
			return fmt.Errorf("telegram %s request: %w", method, urlError.Err)
		}
		return fmt.Errorf("telegram %s request: %w", method, err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil {
		return fmt.Errorf("telegram %s read response: %w", method, err)
	}
	if len(responseBody) > maxResponseBytes {
		return fmt.Errorf("telegram %s response exceeds %d bytes", method, maxResponseBytes)
	}

	var envelope struct {
		OK          bool            `json:"ok"`
		Result      json.RawMessage `json:"result"`
		ErrorCode   int             `json:"error_code"`
		Description string          `json:"description"`
		Parameters  struct {
			RetryAfter int `json:"retry_after"`
		} `json:"parameters"`
	}
	if err := json.Unmarshal(responseBody, &envelope); err != nil {
		return fmt.Errorf("telegram %s decode response (http=%d): %w", method, response.StatusCode, err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices || !envelope.OK {
		return &APIError{Method: method, HTTPStatus: response.StatusCode, ErrorCode: envelope.ErrorCode, Description: envelope.Description, RetryAfter: envelope.Parameters.RetryAfter}
	}
	if result == nil {
		return nil
	}
	if len(envelope.Result) == 0 || string(envelope.Result) == "null" {
		return fmt.Errorf("telegram %s response has no result", method)
	}
	if err := json.Unmarshal(envelope.Result, result); err != nil {
		return fmt.Errorf("telegram %s decode result: %w", method, err)
	}
	return nil
}

func parseBaseURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return nil, errors.New("base URL must use http or https")
	}
	if u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("base URL must be an absolute URL without credentials, query, or fragment")
	}
	return u, nil
}

func validateHTTPSURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if u.Scheme != "https" || u.Host == "" || u.User != nil {
		return errors.New("URL must be an absolute HTTPS URL without credentials")
	}
	return nil
}

func validWebhookSecret(secret string) bool {
	if len(secret) == 0 || len(secret) > 256 {
		return false
	}
	for _, character := range []byte(secret) {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '_' || character == '-' {
			continue
		}
		return false
	}
	return true
}

func validBotToken(token string) bool {
	if token == "" {
		return false
	}
	colon := strings.IndexByte(token, ':')
	if colon < 1 || colon == len(token)-1 {
		return false
	}
	for index, character := range []byte(token) {
		if index < colon {
			if character < '0' || character > '9' {
				return false
			}
			continue
		}
		if index == colon {
			continue
		}
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '_' || character == '-' {
			continue
		}
		return false
	}
	return true
}
