package emby

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	domain "github.com/txyyddss/Remna-User-Panel/internal/emby"
)

const maxResponseBytes = 4 << 20

// APIError is a non-success response from Emby.
type APIError struct {
	HTTPStatus int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Emby API returned HTTP %d: %s", e.HTTPStatus, e.Message)
}

// Option configures an Emby client.
type Option func(*Client)

// WithHTTPClient replaces the bounded default HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		if client != nil {
			c.httpClient = client
		}
	}
}

// Client calls the exact Emby endpoints required by the account saga.
type Client struct {
	baseURL    *url.URL
	token      string
	httpClient *http.Client
}

// NewClient validates configuration and creates an Emby API client.
func NewClient(rawBaseURL, token string, options ...Option) (*Client, error) {
	baseURL, err := parseBaseURL(rawBaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse Emby base URL: %w", err)
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("Emby API token is empty")
	}
	client := &Client{baseURL: baseURL, token: token, httpClient: &http.Client{Timeout: 15 * time.Second}}
	for _, option := range options {
		option(client)
	}
	return client, nil
}

// ListUsers returns all Emby users from GET /Users.
func (c *Client) ListUsers(ctx context.Context) ([]domain.RemoteUser, error) {
	var response []userDTO
	if err := c.do(ctx, http.MethodGet, "/Users", nil, &response); err != nil {
		return nil, err
	}
	users := make([]domain.RemoteUser, 0, len(response))
	for _, user := range response {
		users = append(users, mapUser(user))
	}
	return users, nil
}

// FindUserByName performs the documented list call and matches Name exactly.
func (c *Client) FindUserByName(ctx context.Context, name string) (domain.RemoteUser, bool, error) {
	if strings.TrimSpace(name) == "" {
		return domain.RemoteUser{}, false, errors.New("Emby user name is empty")
	}
	users, err := c.ListUsers(ctx)
	if err != nil {
		return domain.RemoteUser{}, false, err
	}
	for _, user := range users {
		if user.Name == name {
			return user, true, nil
		}
	}
	return domain.RemoteUser{}, false, nil
}

// CreateUser creates a named Emby user through POST /Users/New.
func (c *Client) CreateUser(ctx context.Context, name string) (domain.RemoteUser, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return domain.RemoteUser{}, errors.New("Emby user name is empty")
	}
	var response userDTO
	if err := c.do(ctx, http.MethodPost, "/Users/New", createUserByName{Name: name}, &response); err != nil {
		return domain.RemoteUser{}, err
	}
	user := mapUser(response)
	if user.ID == "" || user.Name != name {
		return domain.RemoteUser{}, errors.New("Emby create response has an unexpected identity")
	}
	return user, nil
}

// GetUser retrieves one exact Emby user through GET /Users/{Id}.
func (c *Client) GetUser(ctx context.Context, id string) (domain.RemoteUser, error) {
	if err := validateID(id); err != nil {
		return domain.RemoteUser{}, err
	}
	var response userDTO
	if err := c.do(ctx, http.MethodGet, "/Users/"+id, nil, &response); err != nil {
		return domain.RemoteUser{}, err
	}
	if response.Policy == nil {
		return domain.RemoteUser{}, errors.New("Emby user response is missing Policy")
	}
	user := mapUser(response)
	if user.ID == "" {
		return domain.RemoteUser{}, errors.New("Emby user response is missing Id")
	}
	return user, nil
}

// SetPassword posts the complete documented UpdateUserPassword body.
func (c *Client) SetPassword(ctx context.Context, id string, currentPassword, newPassword []byte) error {
	if err := validateID(id); err != nil {
		return err
	}
	if len(newPassword) == 0 {
		return errors.New("new Emby password is empty")
	}
	payload := updateUserPassword{
		ID: id, CurrentPassword: string(currentPassword), NewPassword: string(newPassword), ResetPassword: false,
	}
	return c.do(ctx, http.MethodPost, "/Users/"+id+"/Password", payload, nil)
}

// UpdatePolicy posts the complete fetched-and-overlaid Users.UserPolicy.
func (c *Client) UpdatePolicy(ctx context.Context, id string, policy domain.Policy) error {
	if err := validateID(id); err != nil {
		return err
	}
	if policy == nil {
		return errors.New("Emby policy is nil")
	}
	return c.do(ctx, http.MethodPost, "/Users/"+id+"/Policy", policy, nil)
}

// ListSelectableFolders returns GET /Library/SelectableMediaFolders.
func (c *Client) ListSelectableFolders(ctx context.Context) ([]domain.Folder, error) {
	var response []mediaFolder
	if err := c.do(ctx, http.MethodGet, "/Library/SelectableMediaFolders", nil, &response); err != nil {
		return nil, err
	}
	folders := make([]domain.Folder, 0, len(response))
	for _, folder := range response {
		if folder.ID != "" {
			folders = append(folders, domain.Folder{ID: folder.ID, Name: folder.Name})
		}
	}
	return folders, nil
}

// ListParentalRatings returns GET /Localization/ParentalRatings.
func (c *Client) ListParentalRatings(ctx context.Context) ([]domain.ParentalRating, error) {
	var response []parentalRating
	if err := c.do(ctx, http.MethodGet, "/Localization/ParentalRatings", nil, &response); err != nil {
		return nil, err
	}
	ratings := make([]domain.ParentalRating, 0, len(response))
	for _, rating := range response {
		ratings = append(ratings, domain.ParentalRating{Name: rating.Name, Value: rating.Value})
	}
	return ratings, nil
}

// IsNotFound reports an authoritative Emby HTTP 404.
func (c *Client) IsNotFound(err error) bool {
	var apiError *APIError
	return errors.As(err, &apiError) && apiError.HTTPStatus == http.StatusNotFound
}

// IsTerminal classifies non-retryable client-side Emby responses.
func (c *Client) IsTerminal(err error) bool {
	var apiError *APIError
	if !errors.As(err, &apiError) {
		return false
	}
	return apiError.HTTPStatus >= 400 && apiError.HTTPStatus < 500 &&
		apiError.HTTPStatus != http.StatusRequestTimeout && apiError.HTTPStatus != http.StatusTooManyRequests
}

type userDTO struct {
	Name   string        `json:"Name"`
	ID     string        `json:"Id"`
	Policy domain.Policy `json:"Policy"`
}

type createUserByName struct {
	Name string `json:"Name"`
}

type updateUserPassword struct {
	ID              string `json:"Id"`
	CurrentPassword string `json:"CurrentPw"`
	NewPassword     string `json:"NewPw"`
	ResetPassword   bool   `json:"ResetPassword"`
}

type mediaFolder struct {
	Name string `json:"Name"`
	ID   string `json:"Id"`
}

type parentalRating struct {
	Name  string `json:"Name"`
	Value int32  `json:"Value"`
}

func mapUser(user userDTO) domain.RemoteUser {
	if user.Policy == nil {
		user.Policy = make(domain.Policy)
	}
	return domain.RemoteUser{ID: user.ID, Name: user.Name, Policy: user.Policy}
}

func (c *Client) do(ctx context.Context, method, endpoint string, input, output any) error {
	if ctx == nil {
		return errors.New("Emby request context is nil")
	}
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return fmt.Errorf("encode Emby request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}
	target := *c.baseURL
	target.Path = strings.TrimRight(target.Path, "/") + endpoint
	request, err := http.NewRequestWithContext(ctx, method, target.String(), body)
	if err != nil {
		return fmt.Errorf("create Emby request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("X-Emby-Token", c.token)
	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("perform Emby request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil {
		return fmt.Errorf("read Emby response: %w", err)
	}
	if len(responseBody) > maxResponseBytes {
		return fmt.Errorf("Emby response exceeds %d bytes", maxResponseBytes)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return decodeAPIError(response.StatusCode, responseBody)
	}
	if output == nil || response.StatusCode == http.StatusNoContent || len(bytes.TrimSpace(responseBody)) == 0 {
		return nil
	}
	if err := json.Unmarshal(responseBody, output); err != nil {
		return fmt.Errorf("decode Emby response: %w", err)
	}
	return nil
}

func decodeAPIError(status int, _ []byte) error {
	// Do not retain provider error bodies: a password validation response could
	// echo secret material and later be persisted in outbox/account diagnostics.
	return &APIError{HTTPStatus: status, Message: http.StatusText(status)}
}

func parseBaseURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, err
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return nil, errors.New("URL must use http or https")
	}
	if parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, errors.New("URL must be absolute and contain no credentials, query, or fragment")
	}
	return parsed, nil
}

func validateID(id string) error {
	if strings.TrimSpace(id) == "" || strings.ContainsAny(id, `/\\?#`) {
		return errors.New("invalid Emby user Id")
	}
	return nil
}

var _ domain.Remote = (*Client)(nil)
