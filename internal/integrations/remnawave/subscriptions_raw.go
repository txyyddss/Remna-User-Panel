package remnawave

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

// RawSubscription contains credential-bearing client configurations. It must
// remain server-side and must never be logged or persisted.
type RawSubscription struct {
	User struct {
		ID        int64      `json:"id"`
		ShortUUID string     `json:"shortUuid"`
		Status    UserStatus `json:"status"`
	} `json:"user"`
	ConvertedUserInfo struct {
		HWIDCheckup *struct {
			SubscriptionAllowed bool `json:"subscriptionAllowed"`
		} `json:"hwidCheckup"`
	} `json:"convertedUserInfo"`
	ResolvedProxyConfigs []json.RawMessage `json:"resolvedProxyConfigs"`
}

// GetRawSubscription retrieves the protected, resolved subscription, including
// hidden hosts and excluding disabled hosts according to the documented default.
func (c *Client) GetRawSubscription(ctx context.Context, shortUUID string) (*RawSubscription, error) {
	if strings.TrimSpace(shortUUID) == "" {
		return nil, errors.New("Remnawave short UUID is required")
	}
	var envelope struct {
		Response RawSubscription `json:"response"`
	}
	path := "/api/subscriptions/by-short-uuid/" + url.PathEscape(shortUUID) + "/raw"
	if err := c.do(ctx, http.MethodGet, path, nil, nil, &envelope); err != nil {
		return nil, err
	}
	if envelope.Response.User.ID <= 0 || envelope.Response.User.ShortUUID != shortUUID || envelope.Response.ResolvedProxyConfigs == nil {
		return nil, errors.New("Remnawave raw subscription identity or configuration collection is invalid")
	}
	return &envelope.Response, nil
}
