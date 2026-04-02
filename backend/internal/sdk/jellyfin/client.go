package jellyfin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the Jellyfin API client
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new Jellyfin API client
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
	req.Header.Set("Authorization", fmt.Sprintf("MediaBrowser Token=\"%s\"", c.token))
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
		return respBody, fmt.Errorf("Jellyfin API error %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

// --- User Management ---

type JellyfinUser struct {
	ID                 string `json:"Id"`
	Name               string `json:"Name"`
	HasPassword        bool   `json:"HasPassword"`
	HasConfiguredPassword bool `json:"HasConfiguredPassword"`
	EnableAutoLogin    bool   `json:"EnableAutoLogin"`
}

type CreateUserRequest struct {
	Name     string `json:"Name"`
	Password string `json:"Password"`
}

type UserPolicy struct {
	IsAdministrator               bool   `json:"IsAdministrator"`
	IsDisabled                    bool   `json:"IsDisabled"`
	MaxParentalRating             *int   `json:"MaxParentalRating,omitempty"`
	EnableContentDeletion         bool   `json:"EnableContentDeletion"`
	EnableContentDownloading      bool   `json:"EnableContentDownloading"`
	EnableMediaPlayback           bool   `json:"EnableMediaPlayback"`
	EnableAudioPlaybackTranscoding bool  `json:"EnableAudioPlaybackTranscoding"`
	EnableVideoPlaybackTranscoding bool  `json:"EnableVideoPlaybackTranscoding"`
	EnablePlaybackRemuxing        bool   `json:"EnablePlaybackRemuxing"`
	EnableMediaConversion         bool   `json:"EnableMediaConversion"`
	EnableRemoteAccess            bool   `json:"EnableRemoteAccess"`
	SimultaneousStreamLimit       int    `json:"SimultaneousStreamLimit"`
	AuthenticationProviderId      string `json:"AuthenticationProviderId"`
	PasswordResetProviderId       string `json:"PasswordResetProviderId"`
}

// CreateUser creates a new Jellyfin user with transcoding disabled
func (c *Client) CreateUser(name, password string) (*JellyfinUser, error) {
	data, err := c.do("POST", "/Users/New", CreateUserRequest{
		Name:     name,
		Password: password,
	})
	if err != nil {
		return nil, err
	}
	var user JellyfinUser
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}

	// Disable transcoding immediately
	rating := 0
	policy := UserPolicy{
		MaxParentalRating:              &rating,
		EnableMediaPlayback:            true,
		EnableAudioPlaybackTranscoding: false,
		EnableVideoPlaybackTranscoding: false,
		EnablePlaybackRemuxing:         false,
		EnableMediaConversion:          false,
		EnableRemoteAccess:             true,
		EnableContentDownloading:       false,
		EnableContentDeletion:          false,
		AuthenticationProviderId:       "Jellyfin.Server.Implementations.Users.DefaultAuthenticationProvider",
		PasswordResetProviderId:        "Jellyfin.Server.Implementations.Users.DefaultPasswordResetProvider",
	}
	if _, err := c.do("POST", "/Users/"+user.ID+"/Policy", policy); err != nil {
		return &user, fmt.Errorf("set policy: %w", err)
	}

	return &user, nil
}

// DeleteUser deletes a Jellyfin user
func (c *Client) DeleteUser(userID string) error {
	_, err := c.do("DELETE", "/Users/"+userID, nil)
	return err
}

// UpdatePassword changes a user's password
func (c *Client) UpdatePassword(userID, currentPass, newPass string) error {
	body := map[string]string{
		"CurrentPw": currentPass,
		"NewPw":     newPass,
	}
	_, err := c.do("POST", "/Users/"+userID+"/Password", body)
	return err
}

// UpdateParentalRating updates a user's max parental rating (0~22)
func (c *Client) UpdateParentalRating(userID string, rating int) error {
	if rating < 0 {
		rating = 0
	}
	if rating > 22 {
		rating = 22
	}
	policy := map[string]interface{}{
		"MaxParentalRating": rating,
	}
	_, err := c.do("POST", "/Users/"+userID+"/Policy", policy)
	return err
}

// --- Quick Connect ---

// AuthorizeQuickConnect authorizes a Quick Connect code
func (c *Client) AuthorizeQuickConnect(code string) error {
	_, err := c.do("POST", fmt.Sprintf("/QuickConnect/Authorize?Code=%s", code), nil)
	return err
}

// --- Devices ---

type DeviceInfo struct {
	ID             string `json:"Id"`
	Name           string `json:"Name"`
	AppName        string `json:"AppName"`
	AppVersion     string `json:"AppVersion"`
	LastUserName   string `json:"LastUserName"`
	LastUserID     string `json:"LastUserId"`
	DateLastActivity string `json:"DateLastActivity"`
}

type DevicesResult struct {
	Items      []DeviceInfo `json:"Items"`
	TotalCount int          `json:"TotalRecordCount"`
}

// GetDevices gets devices for a user
func (c *Client) GetDevices(userID string) (*DevicesResult, error) {
	data, err := c.do("GET", "/Devices?UserId="+userID, nil)
	if err != nil {
		return nil, err
	}
	var result DevicesResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetUser gets a user by ID
func (c *Client) GetUser(userID string) (*JellyfinUser, error) {
	data, err := c.do("GET", "/Users/"+userID, nil)
	if err != nil {
		return nil, err
	}
	var user JellyfinUser
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}
	return &user, nil
}
