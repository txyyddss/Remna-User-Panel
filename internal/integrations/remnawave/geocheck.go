package remnawave

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// GeocheckImage is the SVG image returned by a completed node geocheck.
type GeocheckImage struct {
	Format    string
	MediaType string
	Encoding  string
	Data      string
}

// NodeGeocheck is one asynchronous geocheck job result.
type NodeGeocheck struct {
	Completed bool
	Failed    bool
	Success   bool
	NodeUUID  string
	Image     *GeocheckImage
}

// RequestNodeGeocheck starts a documented asynchronous geocheck for one node.
func (c *Client) RequestNodeGeocheck(ctx context.Context, nodeUUID string) (string, error) {
	if _, err := uuid.Parse(strings.TrimSpace(nodeUUID)); err != nil {
		return "", errors.New("invalid Remnawave node UUID")
	}
	var envelope struct {
		Response struct {
			JobID string `json:"jobId"`
		} `json:"response"`
	}
	if err := c.do(ctx, http.MethodPost, "/api/connections/geocheck/"+url.PathEscape(nodeUUID), nil, struct{}{}, &envelope); err != nil {
		return "", err
	}
	if strings.TrimSpace(envelope.Response.JobID) == "" {
		return "", errors.New("remnawave geocheck returned an empty job id")
	}
	return envelope.Response.JobID, nil
}

// NodeGeocheckResult polls a documented node geocheck job.
func (c *Client) NodeGeocheckResult(ctx context.Context, jobID string) (NodeGeocheck, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return NodeGeocheck{}, errors.New("remnawave geocheck job id is empty")
	}
	var envelope struct {
		Response struct {
			Completed bool `json:"isCompleted"`
			Failed    bool `json:"isFailed"`
			Result    *struct {
				Success  bool   `json:"success"`
				NodeUUID string `json:"nodeUuid"`
				Image    *struct {
					Format    string `json:"format"`
					MediaType string `json:"media_type"`
					Encoding  string `json:"encoding"`
					Data      string `json:"data"`
				} `json:"image"`
			} `json:"result"`
		} `json:"response"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/connections/geocheck/"+url.PathEscape(jobID), nil, nil, &envelope); err != nil {
		return NodeGeocheck{}, err
	}
	result := NodeGeocheck{Completed: envelope.Response.Completed, Failed: envelope.Response.Failed}
	if envelope.Response.Result == nil {
		return result, nil
	}
	result.Success = envelope.Response.Result.Success
	result.NodeUUID = envelope.Response.Result.NodeUUID
	if image := envelope.Response.Result.Image; image != nil {
		result.Image = &GeocheckImage{Format: image.Format, MediaType: image.MediaType, Encoding: image.Encoding, Data: image.Data}
	}
	if result.Completed && !result.Failed && !result.Success {
		result.Failed = true
	}
	return result, nil
}
