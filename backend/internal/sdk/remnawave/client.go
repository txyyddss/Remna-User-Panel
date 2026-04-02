package remnawave

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the Remnawave API client
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new Remnawave API client
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) do(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return respBody, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

// APIResponse is the generic Remnawave response wrapper
type APIResponse struct {
	Response json.RawMessage `json:"response"`
}

// --- User types ---

type UserTrafficData struct {
	UsedTrafficBytes         int64      `json:"usedTrafficBytes"`
	LifetimeUsedTrafficBytes int64      `json:"lifetimeUsedTrafficBytes"`
	OnlineAt                 *time.Time `json:"onlineAt"`
}

type UserData struct {
	UUID                     string     `json:"uuid"`
	ShortUUID                string     `json:"shortUuid"`
	Username                 string     `json:"username"`
	Status                   string     `json:"status"`
	TrafficLimitBytes        int64      `json:"trafficLimitBytes"`
	TrafficLimitStrategy     string     `json:"trafficLimitStrategy"`
	UsedTrafficBytes         int64      `json:"usedTrafficBytes"`
	LifetimeUsedTrafficBytes int64      `json:"lifetimeUsedTrafficBytes"`
	ExpireAt                 time.Time  `json:"expireAt"`
	CreatedAt                time.Time  `json:"createdAt"`
	LastTrafficResetAt       *time.Time `json:"lastTrafficResetAt"`
	TelegramID               *int64     `json:"telegramId"`
	Email                    string     `json:"email"`
	Description              string     `json:"description"`
	Tag                      string     `json:"tag"`
	HwidDeviceLimit          int        `json:"hwidDeviceLimit"`
	SubscriptionURL          string     `json:"subscriptionUrl"`
	OnlineAt                 *time.Time `json:"onlineAt"`
	SubLastUserAgent         string     `json:"subLastUserAgent"`
	SubRevokedAt             *time.Time `json:"subRevokedAt"`
	ActiveInternalSquads     []Squad    `json:"activeInternalSquads"`
	ExternalSquadUUID        string     `json:"externalSquadUuid"`
	UserTraffic              UserTrafficData `json:"userTraffic"`
}

func (u *UserData) UnmarshalJSON(data []byte) error {
	type alias UserData
	var raw struct {
		alias
		ActiveInternalSquads json.RawMessage `json:"activeInternalSquads"`
		UserTraffic          *UserTrafficData `json:"userTraffic"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*u = UserData(raw.alias)

	if raw.UserTraffic != nil {
		u.UserTraffic = *raw.UserTraffic
		if u.UsedTrafficBytes == 0 {
			u.UsedTrafficBytes = raw.UserTraffic.UsedTrafficBytes
		}
		if u.LifetimeUsedTrafficBytes == 0 {
			u.LifetimeUsedTrafficBytes = raw.UserTraffic.LifetimeUsedTrafficBytes
		}
		if u.OnlineAt == nil {
			u.OnlineAt = raw.UserTraffic.OnlineAt
		}
	}

	if len(raw.ActiveInternalSquads) == 0 || string(raw.ActiveInternalSquads) == "null" {
		return nil
	}

	if err := json.Unmarshal(raw.ActiveInternalSquads, &u.ActiveInternalSquads); err == nil {
		return nil
	}

	var squadIDs []string
	if err := json.Unmarshal(raw.ActiveInternalSquads, &squadIDs); err == nil {
		u.ActiveInternalSquads = make([]Squad, 0, len(squadIDs))
		for _, squadID := range squadIDs {
			u.ActiveInternalSquads = append(u.ActiveInternalSquads, Squad{UUID: squadID})
		}
		return nil
	}

	var squadID string
	if err := json.Unmarshal(raw.ActiveInternalSquads, &squadID); err == nil && squadID != "" {
		u.ActiveInternalSquads = []Squad{{UUID: squadID}}
		return nil
	}

	var singleSquad Squad
	if err := json.Unmarshal(raw.ActiveInternalSquads, &singleSquad); err == nil && singleSquad.UUID != "" {
		u.ActiveInternalSquads = []Squad{singleSquad}
		return nil
	}

	return nil
}

type CreateUserRequest struct {
	Username             string   `json:"username"`
	Status               string   `json:"status,omitempty"`
	TrafficLimitBytes    int64    `json:"trafficLimitBytes,omitempty"`
	TrafficLimitStrategy string   `json:"trafficLimitStrategy,omitempty"`
	ExpireAt             string   `json:"expireAt"`
	TelegramID           int64    `json:"telegramId,omitempty"`
	Description          string   `json:"description,omitempty"`
	Tag                  string   `json:"tag,omitempty"`
	HwidDeviceLimit      int      `json:"hwidDeviceLimit,omitempty"`
	ActiveInternalSquads []string `json:"activeInternalSquads,omitempty"`
	ExternalSquadUUID    string   `json:"externalSquadUuid,omitempty"`
}

type UpdateUserRequest struct {
	UUID                 string   `json:"uuid,omitempty"`
	Username             string   `json:"username,omitempty"`
	Status               string   `json:"status,omitempty"`
	TrafficLimitBytes    int64    `json:"trafficLimitBytes,omitempty"`
	TrafficLimitStrategy string   `json:"trafficLimitStrategy,omitempty"`
	ExpireAt             string   `json:"expireAt,omitempty"`
	TelegramID           int64    `json:"telegramId,omitempty"`
	Description          string   `json:"description,omitempty"`
	Tag                  string   `json:"tag,omitempty"`
	HwidDeviceLimit      int      `json:"hwidDeviceLimit,omitempty"`
	ActiveInternalSquads []string `json:"activeInternalSquads,omitempty"`
	ExternalSquadUUID    string   `json:"externalSquadUuid,omitempty"`
}

func decodeUsersResponse(data []byte) ([]UserData, error) {
	var direct struct {
		Response []UserData `json:"response"`
	}
	if err := json.Unmarshal(data, &direct); err == nil && direct.Response != nil {
		return direct.Response, nil
	}

	var wrapped struct {
		Response struct {
			Users []UserData `json:"users"`
		} `json:"response"`
	}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return nil, err
	}
	return wrapped.Response.Users, nil
}

func decodeUserResponse(data []byte) (*UserData, error) {
	var direct struct {
		Response UserData `json:"response"`
	}
	if err := json.Unmarshal(data, &direct); err == nil && direct.Response.UUID != "" {
		return &direct.Response, nil
	}

	var wrapped struct {
		Response struct {
			User UserData `json:"user"`
		} `json:"response"`
	}
	if err := json.Unmarshal(data, &wrapped); err == nil && wrapped.Response.User.UUID != "" {
		return &wrapped.Response.User, nil
	}

	var user UserData
	if err := json.Unmarshal(data, &user); err == nil && user.UUID != "" {
		return &user, nil
	}

	return nil, fmt.Errorf("unexpected Remnawave user response shape")
}

func decodeSquadsResponse(data []byte, key string) ([]Squad, error) {
	var direct struct {
		Response []Squad `json:"response"`
	}
	if err := json.Unmarshal(data, &direct); err == nil && direct.Response != nil {
		return direct.Response, nil
	}

	var wrapped struct {
		Response map[string]json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return nil, err
	}

	raw, ok := wrapped.Response[key]
	if !ok || len(raw) == 0 {
		return []Squad{}, nil
	}

	var squads []Squad
	if err := json.Unmarshal(raw, &squads); err != nil {
		return nil, err
	}
	return squads, nil
}

// CreateUser creates a new user in Remnawave
func (c *Client) CreateUser(req CreateUserRequest) (*UserData, error) {
	data, err := c.do("POST", "/api/users", req)
	if err != nil {
		return nil, err
	}
	return decodeUserResponse(data)
}

// UpdateUser updates a user
func (c *Client) UpdateUser(req UpdateUserRequest) (*UserData, error) {
	data, err := c.do("PATCH", "/api/users", req)
	if err != nil {
		return nil, err
	}
	return decodeUserResponse(data)
}

// GetUserByUUID gets a user by UUID
func (c *Client) GetUserByUUID(uuid string) (*UserData, error) {
	data, err := c.do("GET", "/api/users/"+uuid, nil)
	if err != nil {
		return nil, err
	}
	return decodeUserResponse(data)
}

// GetUserByTelegramID gets users by Telegram ID
func (c *Client) GetUserByTelegramID(telegramID int64) ([]UserData, error) {
	data, err := c.do("GET", fmt.Sprintf("/api/users/by-telegram-id/%d", telegramID), nil)
	if err != nil {
		return nil, err
	}
	return decodeUsersResponse(data)
}

// GetUserByShortUUID gets a user by short UUID
func (c *Client) GetUserByShortUUID(shortUUID string) (*UserData, error) {
	data, err := c.do("GET", "/api/users/by-short-uuid/"+shortUUID, nil)
	if err != nil {
		return nil, err
	}
	return decodeUserResponse(data)
}

// DeleteUser deletes a user by UUID
func (c *Client) DeleteUser(uuid string) error {
	_, err := c.do("DELETE", "/api/users/"+uuid, nil)
	return err
}

// EnableUser enables a user
func (c *Client) EnableUser(uuid string) error {
	_, err := c.do("POST", "/api/users/"+uuid+"/actions/enable", nil)
	return err
}

// DisableUser disables a user
func (c *Client) DisableUser(uuid string) error {
	_, err := c.do("POST", "/api/users/"+uuid+"/actions/disable", nil)
	return err
}

// RevokeUser revokes a user's subscription
func (c *Client) RevokeUser(uuid string) error {
	_, err := c.do("POST", "/api/users/"+uuid+"/actions/revoke", nil)
	return err
}

// ResetTraffic resets a user's traffic
func (c *Client) ResetTraffic(uuid string) error {
	_, err := c.do("POST", "/api/users/"+uuid+"/actions/reset-traffic", nil)
	return err
}

// --- Squads ---

type Squad struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

// GetInternalSquads gets all internal squads
func (c *Client) GetInternalSquads() ([]Squad, error) {
	data, err := c.do("GET", "/api/internal-squads", nil)
	if err != nil {
		return nil, err
	}
	return decodeSquadsResponse(data, "internalSquads")
}

// GetExternalSquads gets all external squads
func (c *Client) GetExternalSquads() ([]Squad, error) {
	data, err := c.do("GET", "/api/external-squads", nil)
	if err != nil {
		return nil, err
	}
	return decodeSquadsResponse(data, "externalSquads")
}

// --- Bandwidth Stats ---

type BandwidthUsage struct {
	UserUUID    string `json:"userUuid"`
	NodeUUID    string `json:"nodeUuid"`
	NodeName    string `json:"nodeName"`
	CountryCode string `json:"countryCode"`
	Total       int64  `json:"total"`
	Date        string `json:"date"`
}

// GetUserBandwidthStats gets user bandwidth usage (legacy endpoint for per-node breakdown)
func (c *Client) GetUserBandwidthStats(userUUID, start, end string) ([]BandwidthUsage, error) {
	path := fmt.Sprintf("/api/bandwidth-stats/users/%s/legacy?start=%s&end=%s", userUUID, start, end)
	data, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Response []BandwidthUsage `json:"response"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Response, nil
}

// --- HWID Devices ---

type HWIDDevice struct {
	HWID        string `json:"hwid"`
	UserUUID    string `json:"userUuid"`
	Platform    string `json:"platform"`
	OsVersion   string `json:"osVersion"`
	DeviceModel string `json:"deviceModel"`
	UserAgent   string `json:"userAgent"`
}

// GetUserHWIDDevices gets a user's HWID devices
func (c *Client) GetUserHWIDDevices(userUUID string) ([]HWIDDevice, error) {
	data, err := c.do("GET", "/api/hwid/devices/"+userUUID, nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Response []HWIDDevice `json:"response"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Response, nil
}

// --- IP Management ---

type FetchIPsResponse struct {
	JobID string `json:"jobId"`
}

type IPResult struct {
	IPs []string `json:"ips"`
}

// FetchUserIPs requests IP list for a user (returns job ID)
func (c *Client) FetchUserIPs(userUUID string) (string, error) {
	data, err := c.do("POST", "/api/ip-control/fetch-ips/"+userUUID, nil)
	if err != nil {
		return "", err
	}
	var resp struct {
		Response FetchIPsResponse `json:"response"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	return resp.Response.JobID, nil
}

// GetFetchIPsResult gets the IP list result by job ID
func (c *Client) GetFetchIPsResult(jobID string) (json.RawMessage, error) {
	data, err := c.do("GET", "/api/ip-control/fetch-ips/result/"+jobID, nil)
	if err != nil {
		return nil, err
	}
	var resp APIResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Response, nil
}

// DropConnections drops connections for a user
func (c *Client) DropConnections(userUUIDs []string) error {
	body := map[string]interface{}{
		"dropBy": map[string]interface{}{
			"userUuids": userUUIDs,
		},
		"targetNodes": map[string]interface{}{
			"allNodes": true,
		},
	}
	_, err := c.do("POST", "/api/ip-control/drop-connections", body)
	return err
}

// --- Subscription History ---

type SubRequestHistory struct {
	ID        int    `json:"id"`
	Path      string `json:"path"`
	UserAgent string `json:"userAgent"`
	IP        string `json:"ip"`
	CreatedAt string `json:"createdAt"`
}

// GetUserSubHistory gets subscription request history
func (c *Client) GetUserSubHistory(userUUID string) ([]SubRequestHistory, error) {
	data, err := c.do("GET", "/api/users/"+userUUID+"/subscription-request-history", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Response []SubRequestHistory `json:"response"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Response, nil
}

// GetUserAccessibleNodes gets user's accessible nodes
func (c *Client) GetUserAccessibleNodes(userUUID string) (json.RawMessage, error) {
	data, err := c.do("GET", "/api/users/"+userUUID+"/accessible-nodes", nil)
	if err != nil {
		return nil, err
	}
	var resp APIResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Response, nil
}
