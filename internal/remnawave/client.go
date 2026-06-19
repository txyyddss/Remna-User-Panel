// Package remnawave contains the small Remnawave Panel API client used by the
// Web App and admin handlers.
package remnawave

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"remna-user-panel/internal/config"
	appsettings "remna-user-panel/internal/settings"
)

var ErrNotConfigured = errors.New("remnawave_not_configured")
var ErrNotFound = errors.New("remnawave_not_found")

type APIError struct {
	StatusCode int
	ErrorCode  string
	Message    string
}

func (e APIError) Error() string {
	if e.ErrorCode != "" {
		return e.ErrorCode
	}
	if e.Message != "" {
		return e.Message
	}
	if e.StatusCode > 0 {
		return fmt.Sprintf("remnawave_status_%d", e.StatusCode)
	}
	return "remnawave_request_failed"
}

// EffectiveConfig is the runtime Remnawave integration configuration after
// app_settings overrides are applied.
type EffectiveConfig struct {
	BaseURL               string
	APIKey                string
	TotalTimeout          time.Duration
	UserTrafficLimitGB    float64
	UserTrafficStrategy   string
	UserSquadUUIDs        []string
	UserExternalSquadUUID string
}

// Client talks to the Remnawave Panel REST API.
type Client struct {
	settings config.Settings
	store    appsettings.Store
	http     *http.Client
}

func NewClient(settings config.Settings, store appsettings.Store) *Client {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          10,
		IdleConnTimeout:       60 * time.Second,
	}
	return &Client{
		settings: settings,
		store:    store,
		http: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}
}

func (c *Client) EffectiveConfig(ctx context.Context) EffectiveConfig {
	cfg := EffectiveConfig{
		BaseURL:               strings.TrimRight(c.store.String(ctx, "PANEL_API_URL", c.settings.PanelAPIURL), "/"),
		APIKey:                c.store.String(ctx, "PANEL_API_KEY", c.settings.PanelAPIKey),
		TotalTimeout:          secondsSetting(c.store.Float(ctx, "PANEL_API_TOTAL_TIMEOUT_SECONDS", c.settings.PanelAPITotalTimeout.Seconds()), 25*time.Second),
		UserTrafficLimitGB:    c.store.Float(ctx, "USER_TRAFFIC_LIMIT_GB", c.settings.UserTrafficLimitGB),
		UserTrafficStrategy:   normalizeTrafficStrategy(c.store.String(ctx, "USER_TRAFFIC_STRATEGY", c.settings.UserTrafficStrategy)),
		UserSquadUUIDs:        splitList(c.store.String(ctx, "USER_SQUAD_UUIDS", strings.Join(c.settings.UserSquadUUIDs, ","))),
		UserExternalSquadUUID: strings.TrimSpace(c.store.String(ctx, "USER_EXTERNAL_SQUAD_UUID", c.settings.UserExternalSquadUUID)),
	}
	return cfg
}

func (c *Client) Configured(ctx context.Context) bool {
	cfg := c.EffectiveConfig(ctx)
	return cfg.BaseURL != "" && strings.TrimSpace(cfg.APIKey) != ""
}

func (c *Client) GetUserByUUID(ctx context.Context, uuid string) (map[string]any, bool, error) {
	var out map[string]any
	err := c.request(ctx, http.MethodGet, "/users/"+url.PathEscape(strings.TrimSpace(uuid)), nil, nil, &out)
	if errors.Is(err, ErrNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return out, true, nil
}

func (c *Client) GetUsersByTelegramID(ctx context.Context, telegramID int64) ([]map[string]any, error) {
	var out []map[string]any
	err := c.request(ctx, http.MethodGet, "/users/by-telegram-id/"+strconv.FormatInt(telegramID, 10), nil, nil, &out)
	if errors.Is(err, ErrNotFound) {
		return []map[string]any{}, nil
	}
	return out, err
}

func (c *Client) GetUsersByEmail(ctx context.Context, email string) ([]map[string]any, error) {
	var out []map[string]any
	err := c.request(ctx, http.MethodGet, "/users/by-email/"+url.PathEscape(strings.TrimSpace(email)), nil, nil, &out)
	if errors.Is(err, ErrNotFound) {
		return []map[string]any{}, nil
	}
	return out, err
}

func (c *Client) CreateUser(ctx context.Context, payload map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.request(ctx, http.MethodPost, "/users", nil, payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) UpdateUser(ctx context.Context, payload map[string]any) (map[string]any, error) {
	var out map[string]any
	if err := c.request(ctx, http.MethodPatch, "/users", nil, payload, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) SetUserEnabled(ctx context.Context, uuid string, enabled bool) error {
	action := "disable"
	if enabled {
		action = "enable"
	}
	return c.request(ctx, http.MethodPost, "/users/"+url.PathEscape(uuid)+"/actions/"+action, nil, nil, nil)
}

func (c *Client) ResetUserTraffic(ctx context.Context, uuid string) error {
	return c.request(ctx, http.MethodPost, "/users/"+url.PathEscape(strings.TrimSpace(uuid))+"/actions/reset-traffic", nil, nil, nil)
}

func (c *Client) RevokeUserSubscription(ctx context.Context, uuid string) error {
	return c.request(ctx, http.MethodPost, "/users/"+url.PathEscape(strings.TrimSpace(uuid))+"/actions/revoke", nil, nil, nil)
}

func (c *Client) DeleteUser(ctx context.Context, uuid string) error {
	err := c.request(ctx, http.MethodDelete, "/users/"+url.PathEscape(strings.TrimSpace(uuid)), nil, nil, nil)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	return err
}

func (c *Client) GetUserDevices(ctx context.Context, userUUID string) (map[string]any, error) {
	var out map[string]any
	err := c.request(ctx, http.MethodGet, "/hwid/devices/"+url.PathEscape(strings.TrimSpace(userUUID)), nil, nil, &out)
	if errors.Is(err, ErrNotFound) {
		return map[string]any{"total": 0, "devices": []any{}}, nil
	}
	return out, err
}

func (c *Client) DisconnectDevice(ctx context.Context, userUUID string, hwid string) error {
	return c.request(ctx, http.MethodPost, "/hwid/devices/delete", nil, map[string]any{
		"userUuid": userUUID,
		"hwid":     hwid,
	}, nil)
}

func (c *Client) GetInternalSquads(ctx context.Context) ([]map[string]any, error) {
	var out any
	if err := c.request(ctx, http.MethodGet, "/internal-squads", nil, nil, &out); err != nil {
		return nil, err
	}
	return mapsFromAny(out), nil
}

func (c *Client) GetInternalSquadAccessibleNodes(ctx context.Context, squadUUID string) ([]map[string]any, error) {
	var out any
	err := c.request(ctx, http.MethodGet, "/internal-squads/"+url.PathEscape(strings.TrimSpace(squadUUID))+"/accessible-nodes", nil, nil, &out)
	if errors.Is(err, ErrNotFound) {
		return []map[string]any{}, nil
	}
	if err != nil {
		return nil, err
	}
	return mapsFromAny(out), nil
}

func (c *Client) AddUsersToInternalSquad(ctx context.Context, squadUUID string, userUUIDs []string) error {
	return c.request(ctx, http.MethodPost, "/internal-squads/"+url.PathEscape(strings.TrimSpace(squadUUID))+"/bulk-actions/add-users", nil, map[string]any{
		"users":     cleanStrings(userUUIDs),
		"userUuids": cleanStrings(userUUIDs),
	}, nil)
}

func (c *Client) RemoveUsersFromInternalSquad(ctx context.Context, squadUUID string, userUUIDs []string) error {
	return c.request(ctx, http.MethodDelete, "/internal-squads/"+url.PathEscape(strings.TrimSpace(squadUUID))+"/bulk-actions/remove-users", nil, map[string]any{
		"users":     cleanStrings(userUUIDs),
		"userUuids": cleanStrings(userUUIDs),
	}, nil)
}

func (c *Client) GetSystemStats(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	err := c.request(ctx, http.MethodGet, "/system/stats", nil, nil, &out)
	return out, err
}

func (c *Client) GetBandwidthStats(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	err := c.request(ctx, http.MethodGet, "/system/stats/bandwidth", nil, nil, &out)
	return out, err
}

func (c *Client) GetNodesStats(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	err := c.request(ctx, http.MethodGet, "/system/stats/nodes", nil, nil, &out)
	return out, err
}

func (c *Client) GetNodesBandwidthStats(ctx context.Context, query url.Values) (map[string]any, error) {
	var out map[string]any
	err := c.request(ctx, http.MethodGet, "/bandwidth-stats/nodes", query, nil, &out)
	return out, err
}

func (c *Client) GetUserBandwidthStats(ctx context.Context, userUUID string, query url.Values) (map[string]any, error) {
	var out map[string]any
	err := c.request(ctx, http.MethodGet, "/bandwidth-stats/users/"+url.PathEscape(strings.TrimSpace(userUUID)), query, nil, &out)
	return out, err
}

func (c *Client) GetSubscriptionPageConfigs(ctx context.Context) ([]map[string]any, error) {
	var out any
	if err := c.request(ctx, http.MethodGet, "/subscription-page-configs", nil, nil, &out); err != nil {
		return nil, err
	}
	return mapsFromAny(out), nil
}

func (c *Client) GetSubscriptionPageConfigByUUID(ctx context.Context, uuid string) (map[string]any, bool, error) {
	var out map[string]any
	err := c.request(ctx, http.MethodGet, "/subscription-page-configs/"+url.PathEscape(strings.TrimSpace(uuid)), nil, nil, &out)
	if errors.Is(err, ErrNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return out, true, nil
}

func (c *Client) request(ctx context.Context, method string, endpoint string, query url.Values, payload any, out any) error {
	cfg := c.EffectiveConfig(ctx)
	if cfg.BaseURL == "" || cfg.APIKey == "" {
		return ErrNotConfigured
	}
	if cfg.TotalTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.TotalTimeout)
		defer cancel()
	}
	requestURL := apiBaseURL(cfg.BaseURL) + "/" + strings.TrimLeft(endpoint, "/")
	if len(query) > 0 {
		requestURL += "?" + query.Encode()
	}

	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	slog.Debug("remnawave api request", "method", method, "url", requestURL)
	resp, err := c.http.Do(req)
	if err != nil {
		slog.Error("remnawave api request failed", "method", method, "url", requestURL, "error", err)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := parseAPIError(resp.StatusCode, responseBody)
		slog.Error("remnawave api error response", "method", method, "url", requestURL, "status", resp.StatusCode, "error_code", apiErr.ErrorCode, "message", apiErr.Message)
		if resp.StatusCode == http.StatusNotFound || apiErr.ErrorCode == "A040" || apiErr.ErrorCode == "A062" {
			return ErrNotFound
		}
		return apiErr
	}
	if out == nil {
		return nil
	}
	return decodeEnvelope(responseBody, out)
}

func apiBaseURL(base string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if strings.HasSuffix(strings.ToLower(base), "/api") {
		return base
	}
	return base + "/api"
}

func decodeEnvelope(body []byte, out any) error {
	var envelope struct {
		Response json.RawMessage `json:"response"`
	}
	if json.Unmarshal(body, &envelope) == nil && len(envelope.Response) > 0 && string(envelope.Response) != "null" {
		return json.Unmarshal(envelope.Response, out)
	}
	return json.Unmarshal(body, out)
}

func parseAPIError(status int, body []byte) APIError {
	var payload map[string]any
	_ = json.Unmarshal(body, &payload)
	message := ""
	if value, ok := payload["message"]; ok {
		message = fmt.Sprint(value)
	}
	code := ""
	if value, ok := payload["errorCode"]; ok {
		code = fmt.Sprint(value)
	}
	if code == "" {
		if value, ok := payload["code"]; ok {
			code = fmt.Sprint(value)
		}
	}
	return APIError{StatusCode: status, ErrorCode: code, Message: message}
}

func normalizeTrafficStrategy(raw string) string {
	value := strings.ToUpper(strings.TrimSpace(raw))
	switch value {
	case "DAY", "WEEK", "MONTH", "MONTH_ROLLING":
		return value
	default:
		return "NO_RESET"
	}
}

func secondsSetting(value float64, fallback time.Duration) time.Duration {
	if value <= 0 {
		return fallback
	}
	return time.Duration(value * float64(time.Second))
}

func optionalIntFromJSON(raw json.RawMessage) *int {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var number float64
	if json.Unmarshal(raw, &number) == nil {
		value := int(number)
		return &value
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		text = strings.TrimSpace(text)
		if text == "" {
			return nil
		}
		if value, err := strconv.Atoi(text); err == nil {
			return &value
		}
	}
	return nil
}

func splitList(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		if value := strings.TrimSpace(field); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func cleanStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func mapsFromAny(value any) []map[string]any {
	switch typed := value.(type) {
	case []map[string]any:
		return typed
	case []any:
		result := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if mapped, ok := item.(map[string]any); ok {
				result = append(result, mapped)
			}
		}
		return result
	case map[string]any:
		for _, key := range []string{"internalSquads", "squads", "items", "data"} {
			if nested, ok := typed[key]; ok {
				return mapsFromAny(nested)
			}
		}
	}
	return []map[string]any{}
}
