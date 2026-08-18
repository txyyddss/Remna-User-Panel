package remnawave

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// ListHosts returns the fields needed for scheduled multiplier remark upkeep.
func (c *Client) ListHosts(ctx context.Context) ([]Host, error) {
	var envelope struct {
		Response []Host `json:"response"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/hosts", nil, nil, &envelope); err != nil {
		return nil, err
	}
	for _, host := range envelope.Response {
		if _, err := uuid.Parse(host.UUID); err != nil {
			return nil, errors.New("remnawave host response contains an invalid UUID")
		}
	}
	return envelope.Response, nil
}

// UpdateHostRemark patches only the documented UUID and changed remark.
func (c *Client) UpdateHostRemark(ctx context.Context, hostUUID, remark string) error {
	if _, err := uuid.Parse(hostUUID); err != nil {
		return errors.New("invalid Remnawave host UUID")
	}
	if strings.TrimSpace(remark) == "" || len([]rune(remark)) > 100 {
		return errors.New("invalid Remnawave host remark")
	}
	var envelope struct {
		Response Host `json:"response"`
	}
	if err := c.do(ctx, http.MethodPatch, "/api/hosts", nil, map[string]string{"uuid": hostUUID, "remark": remark}, &envelope); err != nil {
		return err
	}
	if envelope.Response.UUID != hostUUID {
		return errors.New("Remnawave host response identity does not match request")
	}
	return nil
}
