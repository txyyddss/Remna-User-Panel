package remnawave

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

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
	if err := requireCreatedUserIdentity(&envelope.Response, input); err != nil {
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
	if err := requireUpdatedUserIdentity(&envelope.Response, input); err != nil {
		return nil, err
	}
	return &envelope.Response, nil
}

// GetUserByID retrieves a user by Remnawave numeric ID.
func (c *Client) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	if userID <= 0 {
		return nil, errors.New("remnawave user id must be positive")
	}
	user, err := c.getUser(ctx, "/api/users/"+strconv.FormatInt(userID, 10))
	if err != nil {
		return nil, err
	}
	if err := requireUserID(user, userID); err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByUsername retrieves a user by exact username.
func (c *Client) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	if strings.TrimSpace(username) == "" {
		return nil, errors.New("remnawave username is empty")
	}
	user, err := c.getUser(ctx, "/api/users/by-username/"+url.PathEscape(username))
	if err != nil {
		return nil, err
	}
	if err := requireUsername(user, username); err != nil {
		return nil, err
	}
	return user, nil
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
	if err := requireResolvedUserIdentity(&envelope.Response, selector); err != nil {
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
	if err := requireUserID(&envelope.Response, userID); err != nil {
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
	if err := requireUserID(&envelope.Response, userID); err != nil {
		return nil, err
	}
	return &envelope.Response, nil
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
