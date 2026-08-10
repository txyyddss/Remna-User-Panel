package remnawave

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// GetSubscription retrieves a user's protected subscription data.
func (c *Client) GetSubscription(ctx context.Context, userID int64) (*Subscription, error) {
	if userID <= 0 {
		return nil, errors.New("remnawave user id must be positive")
	}
	var envelope struct {
		Response Subscription `json:"response"`
	}
	path := "/api/subscriptions/by-id/" + strconv.FormatInt(userID, 10)
	if err := c.do(ctx, http.MethodGet, path, nil, nil, &envelope); err != nil {
		return nil, err
	}
	return &envelope.Response, nil
}

// GetUserStats retrieves a user's usage grouped by node for an inclusive date range.
func (c *Client) GetUserStats(ctx context.Context, userID int64, start, end time.Time, topNodesLimit int) (*UserStats, error) {
	if userID <= 0 || start.IsZero() || end.IsZero() {
		return nil, errors.New("remnawave stats require a positive user id and date range")
	}
	if start.After(end) {
		return nil, errors.New("remnawave stats start date is after end date")
	}
	if topNodesLimit <= 0 {
		topNodesLimit = 20
	}
	query := url.Values{
		"start":         {start.UTC().Format(time.DateOnly)},
		"end":           {end.UTC().Format(time.DateOnly)},
		"topNodesLimit": {strconv.Itoa(topNodesLimit)},
	}
	var envelope struct {
		Response UserStats `json:"response"`
	}
	path := "/api/bandwidth-stats/users/" + strconv.FormatInt(userID, 10)
	if err := c.do(ctx, http.MethodGet, path, query, nil, &envelope); err != nil {
		return nil, err
	}
	return &envelope.Response, nil
}
