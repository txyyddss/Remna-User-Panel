package remnawave

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (c *Client) listUsersStream(ctx context.Context, cursor string, size int, telegramID *int64) ([]User, *string, bool, error) {
	if size < 1 || size > 1000 {
		return nil, nil, false, errors.New("remnawave user stream requires size in 1..1000")
	}
	query := url.Values{"size": {strconv.Itoa(size)}}
	if cursor != "" {
		query.Set("cursor", cursor)
	}
	if telegramID != nil {
		if *telegramID <= 0 {
			return nil, nil, false, errors.New("remnawave Telegram id must be positive")
		}
		query.Set("telegramId", strconv.FormatInt(*telegramID, 10))
	}
	var envelope struct {
		Response struct {
			Users      []User  `json:"users"`
			NextCursor *string `json:"nextCursor"`
			HasMore    bool    `json:"hasMore"`
		} `json:"response"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/users/stream", query, nil, &envelope); err != nil {
		return nil, nil, false, err
	}
	return envelope.Response.Users, envelope.Response.NextCursor, envelope.Response.HasMore, nil
}

// ListUsersStream returns one documented cursor page without a Telegram filter.
func (c *Client) ListUsersStream(ctx context.Context, cursor string, size int) ([]User, *string, bool, error) {
	return c.listUsersStream(ctx, cursor, size, nil)
}

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
