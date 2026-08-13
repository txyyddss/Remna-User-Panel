package remnawave

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

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

// FindUserByTelegramID uses the documented stream filter for an exact identity lookup.
// It returns (nil, nil) when no user is found.

func (c *Client) FindUserByTelegramID(ctx context.Context, telegramID int64) (*User, error) {
	if telegramID <= 0 {
		return nil, errors.New("remnawave Telegram id must be positive")
	}
	const pageSize = 1000
	var cursor string
	for {
		users, nextCursor, hasMore, err := c.listUsersStream(ctx, cursor, pageSize, &telegramID)
		if err != nil {
			return nil, err
		}
		if len(users) > 1 {
			return nil, errors.New("remnawave returned multiple users for one Telegram id")
		}
		if len(users) == 1 {
			if users[0].TelegramID == nil || *users[0].TelegramID != telegramID {
				return nil, errors.New("remnawave Telegram stream returned a mismatched identity")
			}
			return &users[0], nil
		}
		if !hasMore || nextCursor == nil || *nextCursor == "" || *nextCursor == cursor {
			return nil, nil
		}
		cursor = *nextCursor
	}
}

// RevokeSubscription rotates a user's subscription credentials.
