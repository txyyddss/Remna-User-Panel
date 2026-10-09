package remnawave

import (
	"context"
	"errors"
	"net/http"
	"strconv"
)

// ConnectionKeys is an ephemeral set of credential-bearing native import links.
type ConnectionKeys struct {
	EnabledKeys  []string `json:"enabledKeys"`
	HiddenKeys   []string `json:"hiddenKeys"`
	DisabledKeys []string `json:"disabledKeys"`
}

// GetConnectionKeys retrieves upstream-native links for precisely one user.
// Callers must authorize ownership and must not log, persist, or cache these keys.
func (c *Client) GetConnectionKeys(ctx context.Context, userID int64) (*ConnectionKeys, error) {
	if userID <= 0 {
		return nil, errors.New("Remnawave user ID must be positive")
	}
	var envelope struct {
		Response ConnectionKeys `json:"response"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/subscriptions/connection-keys/"+strconv.FormatInt(userID, 10), nil, nil, &envelope); err != nil {
		return nil, err
	}
	if envelope.Response.EnabledKeys == nil || envelope.Response.HiddenKeys == nil || envelope.Response.DisabledKeys == nil {
		return nil, errors.New("Remnawave connection key collections are missing")
	}
	return &envelope.Response, nil
}
