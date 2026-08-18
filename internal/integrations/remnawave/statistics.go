package remnawave

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// GetStatsDigest returns Remnawave's half-open range digest.
func (c *Client) GetStatsDigest(ctx context.Context, start, end time.Time) (StatsDigest, error) {
	if start.IsZero() || !end.After(start) {
		return StatsDigest{}, errors.New("remnawave digest requires an increasing range")
	}
	query := url.Values{"start": {start.UTC().Format(time.RFC3339)}, "end": {end.UTC().Format(time.RFC3339)}}
	var envelope struct {
		Response struct {
			Users struct {
				Created int64 `json:"createdCount"`
				Expired int64 `json:"expiredCount"`
			} `json:"users"`
			Traffic struct {
				Total string `json:"totalBytes"`
			} `json:"traffic"`
		} `json:"response"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/system/stats/digest", query, nil, &envelope); err != nil {
		return StatsDigest{}, err
	}
	traffic, err := strconv.ParseInt(envelope.Response.Traffic.Total, 10, 64)
	if err != nil || traffic < 0 {
		return StatsDigest{}, errors.New("remnawave digest returned invalid traffic")
	}
	return StatsDigest{CreatedUsers: envelope.Response.Users.Created, ExpiredUsers: envelope.Response.Users.Expired, TrafficBytes: traffic}, nil
}

// GetNodesUsage returns raw per-node series for an inclusive date range.
func (c *Client) GetNodesUsage(ctx context.Context, start, end time.Time) (NodesUsage, error) {
	if start.IsZero() || end.Before(start) {
		return NodesUsage{}, errors.New("remnawave node usage requires a valid range")
	}
	query := url.Values{"start": {start.UTC().Format(time.DateOnly)}, "end": {end.UTC().Format(time.DateOnly)}, "topNodesLimit": {"1000"}}
	var envelope struct {
		Response NodesUsage `json:"response"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/bandwidth-stats/nodes", query, nil, &envelope); err != nil {
		return NodesUsage{}, err
	}
	return envelope.Response, nil
}

// GetNodesMetrics returns the on-demand online-user counters.
func (c *Client) GetNodesMetrics(ctx context.Context) ([]NodeMetric, error) {
	var envelope struct {
		Response struct {
			Nodes []NodeMetric `json:"nodes"`
		} `json:"response"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/system/nodes/metrics", nil, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Response.Nodes, nil
}
