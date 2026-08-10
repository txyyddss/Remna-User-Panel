package telegram

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const maxResponseBytes = 2 << 20

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
