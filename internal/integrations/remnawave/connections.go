package remnawave

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// RequestUserConnections starts Remnawave's asynchronous connection scan.
func (c *Client) RequestUserConnections(ctx context.Context, userID int64) (string, error) {
	if userID <= 0 {
		return "", errors.New("remnawave user id must be positive")
	}
	var envelope struct {
		Response struct {
			JobID string `json:"jobId"`
		} `json:"response"`
	}
	path := "/api/connections/by-user/" + strconv.FormatInt(userID, 10)
	if err := c.do(ctx, http.MethodPost, path, nil, nil, &envelope); err != nil {
		return "", err
	}
	if strings.TrimSpace(envelope.Response.JobID) == "" {
		return "", errors.New("remnawave connection scan returned an empty job id")
	}
	return envelope.Response.JobID, nil
}

// UserConnections polls one previously-created connection job.
func (c *Client) UserConnections(ctx context.Context, jobID string) (ConnectionScan, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return ConnectionScan{}, errors.New("remnawave connection job id is empty")
	}
	var envelope struct {
		Response struct {
			Completed bool `json:"isCompleted"`
			Failed    bool `json:"isFailed"`
			Progress  struct {
				Percent float64 `json:"percent"`
			} `json:"progress"`
			Result *struct {
				Success bool             `json:"success"`
				Nodes   []ConnectionNode `json:"nodes"`
			} `json:"result"`
		} `json:"response"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/connections/by-user/"+url.PathEscape(jobID), nil, nil, &envelope); err != nil {
		return ConnectionScan{}, err
	}
	result := ConnectionScan{Completed: envelope.Response.Completed, Failed: envelope.Response.Failed, Progress: envelope.Response.Progress.Percent}
	if envelope.Response.Result != nil {
		result.Nodes = envelope.Response.Result.Nodes
		if result.Completed && !envelope.Response.Result.Success && len(result.Nodes) == 0 {
			result.Failed = true
		}
	}
	return result, validateConnectionScan(result)
}

// DropConnectionByIP drops one selected IP on one selected node.
func (c *Client) DropConnectionByIP(ctx context.Context, ip, nodeUUID string) error {
	if net.ParseIP(strings.TrimSpace(ip)) == nil {
		return errors.New("invalid connection IP address")
	}
	if _, err := uuid.Parse(nodeUUID); err != nil {
		return errors.New("invalid connection node UUID")
	}
	payload := map[string]any{
		"dropBy":      map[string]any{"by": "ipAddresses", "ipAddresses": []string{ip}},
		"targetNodes": map[string]any{"target": "specificNodes", "nodeUuids": []string{nodeUUID}},
	}
	return c.do(ctx, http.MethodPost, "/api/connections/drop", nil, payload, nil)
}

func validateConnectionScan(scan ConnectionScan) error {
	if scan.Progress < 0 || scan.Progress > 100 {
		return errors.New("remnawave connection progress is outside 0..100")
	}
	for _, node := range scan.Nodes {
		if _, err := uuid.Parse(node.UUID); err != nil {
			return errors.New("remnawave connection result has an invalid node UUID")
		}
		for _, item := range node.IPs {
			if net.ParseIP(item.IP) == nil || item.LastSeen.IsZero() {
				return errors.New("remnawave connection result has an invalid IP observation")
			}
		}
	}
	return nil
}
