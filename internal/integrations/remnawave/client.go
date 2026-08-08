package remnawave

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const maxResponseBytes = 4 << 20

// HTTPDoer is implemented by *http.Client and permits transport injection in tests.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// APIError is a sanitized Remnawave API error.
type APIError struct {
	HTTPStatus int
	ErrorCode  string
	Message    string
	Issues     []ValidationIssue
}

// Error implements error.
func (e *APIError) Error() string {
	if e == nil {
		return "remnawave API error"
	}
	if e.ErrorCode != "" {
		return fmt.Sprintf("remnawave API failed (http=%d code=%s): %s", e.HTTPStatus, e.ErrorCode, e.Message)
	}
	return fmt.Sprintf("remnawave API failed (http=%d): %s", e.HTTPStatus, e.Message)
}

// IsErrorCode reports whether err is a Remnawave API error with code, such as A019.
func IsErrorCode(err error, code string) bool {
	var apiError *APIError
	return errors.As(err, &apiError) && apiError.ErrorCode == code
}

// IsNotFound reports whether Remnawave returned HTTP 404.
func IsNotFound(err error) bool {
	var apiError *APIError
	return errors.As(err, &apiError) && apiError.HTTPStatus == http.StatusNotFound
}

// Client calls the protected Remnawave v3.2.1 API using bearer authentication.
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
			return errors.New("remnawave HTTP client is nil")
		}
		c.httpClient = client
		return nil
	}
}

// NewClient creates a Remnawave client from the panel origin and API bearer token.
func NewClient(rawBaseURL, token string, options ...Option) (*Client, error) {
	baseURL, err := parseBaseURL(rawBaseURL)
	if err != nil {
		return nil, fmt.Errorf("remnawave base URL: %w", err)
	}
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("remnawave bearer token is empty")
	}
	c := &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout:       20 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(c); err != nil {
			return nil, fmt.Errorf("configure remnawave client: %w", err)
		}
	}
	return c, nil
}

// CreateUser creates a Remnawave user.
func (c *Client) CreateUser(ctx context.Context, input CreateUserRequest) (*User, error) {
	if strings.TrimSpace(input.Username) == "" || input.ExpireAt.IsZero() || input.TelegramID <= 0 || input.TrafficLimitBytes < 0 {
		return nil, errors.New("remnawave create user requires username, expiration, positive Telegram id, and non-negative traffic")
	}
	if input.Status == "" {
		input.Status = UserStatusActive
	}
	if input.TrafficLimitStrategy == "" {
		input.TrafficLimitStrategy = TrafficNoReset
	}
	if input.ActiveInternalSquads == nil {
		input.ActiveInternalSquads = []string{}
	}
	var envelope struct {
		Response User `json:"response"`
	}
	if err := c.do(ctx, http.MethodPost, "/api/users", nil, input, &envelope); err != nil {
		return nil, err
	}
	return &envelope.Response, nil
}

// UpdateUser patches a user identified by exactly one of ID or Username.
func (c *Client) UpdateUser(ctx context.Context, input UpdateUserRequest) (*User, error) {
	payload, err := updatePayload(input)
	if err != nil {
		return nil, err
	}
	var envelope struct {
		Response User `json:"response"`
	}
	if err := c.do(ctx, http.MethodPatch, "/api/users", nil, payload, &envelope); err != nil {
		return nil, err
	}
	return &envelope.Response, nil
}

// GetUserByID retrieves a user by Remnawave numeric ID.
func (c *Client) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	if userID <= 0 {
		return nil, errors.New("remnawave user id must be positive")
	}
	return c.getUser(ctx, "/api/users/"+strconv.FormatInt(userID, 10))
}

// GetUserByUsername retrieves a user by exact username.
func (c *Client) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("remnawave username is empty")
	}
	return c.getUser(ctx, "/api/users/by-username/"+url.PathEscape(username))
}

// FindUserByUsername retrieves an exact username and maps HTTP 404 to exists=false.
func (c *Client) FindUserByUsername(ctx context.Context, username string) (*User, bool, error) {
	user, err := c.GetUserByUsername(ctx, username)
	if IsNotFound(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return user, true, nil
}

// IsDuplicateError reports the authoritative A019 username race error.
func (c *Client) IsDuplicateError(err error) bool {
	return IsErrorCode(err, "A019")
}

// ResolveUser resolves one ID, short UUID, or username to Remnawave's canonical identity.
func (c *Client) ResolveUser(ctx context.Context, selector UserSelector) (*UserSelector, error) {
	set := 0
	if selector.ID > 0 {
		set++
	}
	if selector.ShortUUID != "" {
		set++
	}
	if selector.Username != "" {
		set++
	}
	if set != 1 {
		return nil, errors.New("remnawave resolve requires exactly one user selector")
	}
	var envelope struct {
		Response UserSelector `json:"response"`
	}
	if err := c.do(ctx, http.MethodPost, "/api/users/resolve", nil, selector, &envelope); err != nil {
		return nil, err
	}
	return &envelope.Response, nil
}

// ListUsers retrieves one page of users using the documented start/size parameters.
func (c *Client) ListUsers(ctx context.Context, start, size int) ([]User, int, error) {
	if start < 0 || size < 1 || size > 1000 {
		return nil, 0, errors.New("remnawave user page requires start >= 0 and size in 1..1000")
	}
	query := url.Values{"start": {strconv.Itoa(start)}, "size": {strconv.Itoa(size)}}
	var envelope struct {
		Response struct {
			Users []User `json:"users"`
			Total int    `json:"total"`
		} `json:"response"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/users", query, nil, &envelope); err != nil {
		return nil, 0, err
	}
	return envelope.Response.Users, envelope.Response.Total, nil
}

// FindUserByTelegramID scans documented user pages for an exact Telegram ID match.
// It returns (nil, nil) when no user is found.
func (c *Client) FindUserByTelegramID(ctx context.Context, telegramID int64) (*User, error) {
	if telegramID <= 0 {
		return nil, errors.New("remnawave Telegram id must be positive")
	}
	const pageSize = 1000
	for start := 0; ; start += pageSize {
		users, total, err := c.ListUsers(ctx, start, pageSize)
		if err != nil {
			return nil, err
		}
		for i := range users {
			if users[i].TelegramID != nil && *users[i].TelegramID == telegramID {
				return &users[i], nil
			}
		}
		if len(users) == 0 || start+len(users) >= total {
			return nil, nil
		}
	}
}

// RevokeSubscription rotates a user's subscription credentials.
func (c *Client) RevokeSubscription(ctx context.Context, userID int64, revokeOnlyPasswords bool) (*User, error) {
	if userID <= 0 {
		return nil, errors.New("remnawave user id must be positive")
	}
	payload := struct {
		RevokeOnlyPasswords bool `json:"revokeOnlyPasswords"`
	}{RevokeOnlyPasswords: revokeOnlyPasswords}
	var envelope struct {
		Response User `json:"response"`
	}
	path := "/api/users/" + strconv.FormatInt(userID, 10) + "/actions/revoke"
	if err := c.do(ctx, http.MethodPost, path, nil, payload, &envelope); err != nil {
		return nil, err
	}
	return &envelope.Response, nil
}

// ResetTraffic resets a user's used traffic counter.
func (c *Client) ResetTraffic(ctx context.Context, userID int64) (*User, error) {
	if userID <= 0 {
		return nil, errors.New("remnawave user id must be positive")
	}
	var envelope struct {
		Response User `json:"response"`
	}
	path := "/api/users/" + strconv.FormatInt(userID, 10) + "/actions/reset-traffic"
	if err := c.do(ctx, http.MethodPost, path, nil, nil, &envelope); err != nil {
		return nil, err
	}
	return &envelope.Response, nil
}

// GetSubscription retrieves a user's protected subscription data.
func (c *Client) GetSubscription(ctx context.Context, userID int64) (*Subscription, error) {
	if userID <= 0 {
		return nil, errors.New("remnawave user id must be positive")
	}
	var envelope struct {
		Response Subscription `json:"response"`
	}
	path := "/api/subscriptions/by-id/" + strconv.FormatInt(userID, 10)
	if err := c.do(ctx, http.MethodGet, path, nil, nil, &envelope); err != nil {
		return nil, err
	}
	return &envelope.Response, nil
}

// GetUserStats retrieves a user's usage grouped by node for an inclusive date range.
func (c *Client) GetUserStats(ctx context.Context, userID int64, start, end time.Time, topNodesLimit int) (*UserStats, error) {
	if userID <= 0 || start.IsZero() || end.IsZero() {
		return nil, errors.New("remnawave stats require a positive user id and date range")
	}
	if start.After(end) {
		return nil, errors.New("remnawave stats start date is after end date")
	}
	if topNodesLimit <= 0 {
		topNodesLimit = 20
	}
	query := url.Values{
		"start":         {start.UTC().Format(time.DateOnly)},
		"end":           {end.UTC().Format(time.DateOnly)},
		"topNodesLimit": {strconv.Itoa(topNodesLimit)},
	}
	var envelope struct {
		Response UserStats `json:"response"`
	}
	path := "/api/bandwidth-stats/users/" + strconv.FormatInt(userID, 10)
	if err := c.do(ctx, http.MethodGet, path, query, nil, &envelope); err != nil {
		return nil, err
	}
	return &envelope.Response, nil
}

// ListInternalSquads returns all Remnawave internal squads available for import.
func (c *Client) ListInternalSquads(ctx context.Context) ([]InternalSquad, error) {
	var envelope struct {
		Response struct {
			Total          int             `json:"total"`
			InternalSquads []InternalSquad `json:"internalSquads"`
		} `json:"response"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/internal-squads", nil, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Response.InternalSquads, nil
}

func (c *Client) getUser(ctx context.Context, endpoint string) (*User, error) {
	var envelope struct {
		Response User `json:"response"`
	}
	if err := c.do(ctx, http.MethodGet, endpoint, nil, nil, &envelope); err != nil {
		return nil, err
	}
	return &envelope.Response, nil
}

func updatePayload(input UpdateUserRequest) (map[string]any, error) {
	if (input.ID > 0) == (strings.TrimSpace(input.Username) != "") {
		return nil, errors.New("remnawave update requires exactly one of id or username")
	}
	payload := make(map[string]any)
	if input.ID > 0 {
		payload["id"] = input.ID
	} else {
		payload["username"] = input.Username
	}
	changes := 0
	if input.Status != nil {
		payload["status"] = *input.Status
		changes++
	}
	if input.TrafficLimitBytes != nil {
		if *input.TrafficLimitBytes < 0 {
			return nil, errors.New("remnawave traffic limit must be non-negative")
		}
		payload["trafficLimitBytes"] = *input.TrafficLimitBytes
		changes++
	}
	if input.TrafficLimitStrategy != nil {
		payload["trafficLimitStrategy"] = *input.TrafficLimitStrategy
		changes++
	}
	if input.ExpireAt != nil {
		payload["expireAt"] = *input.ExpireAt
		changes++
	}
	if input.Description != nil {
		payload["description"] = *input.Description
		changes++
	}
	if input.TelegramID != nil {
		payload["telegramId"] = *input.TelegramID
		changes++
	}
	if input.ActiveInternalSquads != nil {
		squads := *input.ActiveInternalSquads
		if squads == nil {
			squads = []string{}
		}
		payload["activeInternalSquads"] = squads
		changes++
	}
	if input.ClearExternalSquad {
		payload["externalSquadUuid"] = nil
		changes++
	} else if input.ExternalSquadUUID != nil {
		payload["externalSquadUuid"] = *input.ExternalSquadUUID
		changes++
	}
	if changes == 0 {
		return nil, errors.New("remnawave update contains no changes")
	}
	return payload, nil
}

func (c *Client) do(ctx context.Context, method, endpoint string, query url.Values, input, output any) error {
	if ctx == nil {
		return errors.New("remnawave request context is nil")
	}
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("remnawave encode request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}
	target := *c.baseURL
	target.Path = strings.TrimRight(target.Path, "/") + endpoint
	if query != nil {
		target.RawQuery = query.Encode()
	}
	request, err := http.NewRequestWithContext(ctx, method, target.String(), body)
	if err != nil {
		return fmt.Errorf("remnawave create request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+c.token)
	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("remnawave request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil {
		return fmt.Errorf("remnawave read response: %w", err)
	}
	if len(responseBody) > maxResponseBytes {
		return fmt.Errorf("remnawave response exceeds %d bytes", maxResponseBytes)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return decodeAPIError(response.StatusCode, responseBody)
	}
	if output == nil || response.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.Unmarshal(responseBody, output); err != nil {
		return fmt.Errorf("remnawave decode response (http=%d): %w", response.StatusCode, err)
	}
	return nil
}

func decodeAPIError(status int, body []byte) error {
	var payload struct {
		ErrorCode  string            `json:"errorCode"`
		Message    string            `json:"message"`
		StatusCode int               `json:"statusCode"`
		Errors     []ValidationIssue `json:"errors"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return &APIError{HTTPStatus: status, Message: http.StatusText(status)}
	}
	if payload.Message == "" {
		payload.Message = http.StatusText(status)
	}
	return &APIError{HTTPStatus: status, ErrorCode: payload.ErrorCode, Message: payload.Message, Issues: payload.Errors}
}

func parseBaseURL(rawURL string) (*url.URL, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return nil, errors.New("URL must use http or https")
	}
	if u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, errors.New("URL must be absolute and contain no credentials, query, or fragment")
	}
	return u, nil
}
