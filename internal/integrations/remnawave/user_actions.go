package remnawave

import (
	"context"
	"errors"
	"net/http"
	"strconv"
)

// DisableUser uses Remnawave's documented disable action for one numeric user.
func (c *Client) DisableUser(ctx context.Context, userID int64) error {
	return c.userAction(ctx, userID, "disable")
}

// EnableUser uses Remnawave's documented enable action for one numeric user.
func (c *Client) EnableUser(ctx context.Context, userID int64) error {
	return c.userAction(ctx, userID, "enable")
}

func (c *Client) userAction(ctx context.Context, userID int64, action string) error {
	if userID <= 0 {
		return errors.New("remnawave user id must be positive")
	}
	return c.do(ctx, http.MethodPost, "/api/users/"+strconv.FormatInt(userID, 10)+"/actions/"+action, nil, nil, nil)
}
