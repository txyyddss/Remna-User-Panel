package remnawave

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

// ListInternalSquads returns all Remnawave internal squads available for import.
func (c *Client) ListInternalSquads(ctx context.Context) ([]InternalSquad, error) {
	var envelope struct {
		Response struct {
			Total          int             `json:"total"`
			InternalSquads []InternalSquad `json:"internalSquads"`
		} `json:"response"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/internal-squads", nil, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Response.InternalSquads, nil
}

// ListNodes returns the documented node collection used for administrator
// squad assignment. The caller decides which active inbound UUIDs to union.
func (c *Client) ListNodes(ctx context.Context) ([]Node, error) {
	var envelope struct {
		Response []Node `json:"response"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/nodes", nil, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Response, nil
}

// InternalSquadAccessibleNodes re-fetches Remnawave's actual accessibility,
// including extra nodes made reachable by shared inbound UUIDs.
func (c *Client) InternalSquadAccessibleNodes(ctx context.Context, squadUUID string) ([]AccessibleNode, error) {
	if _, err := uuid.Parse(squadUUID); err != nil {
		return nil, errors.New("invalid Remnawave squad UUID")
	}
	var envelope struct {
		Response struct {
			AccessibleNodes []AccessibleNode `json:"accessibleNodes"`
		} `json:"response"`
	}
	if err := c.do(ctx, http.MethodGet, "/api/internal-squads/"+url.PathEscape(squadUUID)+"/accessible-nodes", nil, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Response.AccessibleNodes, nil
}

// UpdateInternalSquadInbounds applies only the documented UUID and inbounds
// fields, leaving the squad name untouched.
func (c *Client) UpdateInternalSquadInbounds(ctx context.Context, squadUUID string, inbounds []string) (*InternalSquad, error) {
	if _, err := uuid.Parse(squadUUID); err != nil {
		return nil, errors.New("invalid Remnawave squad UUID")
	}
	if inbounds == nil {
		inbounds = []string{}
	}
	for _, inboundUUID := range inbounds {
		if _, err := uuid.Parse(inboundUUID); err != nil {
			return nil, errors.New("invalid Remnawave inbound UUID")
		}
	}
	var envelope struct {
		Response InternalSquad `json:"response"`
	}
	if err := c.do(ctx, http.MethodPatch, "/api/internal-squads", nil, map[string]any{"uuid": squadUUID, "inbounds": inbounds}, &envelope); err != nil {
		return nil, err
	}
	if envelope.Response.UUID != squadUUID {
		return nil, errors.New("Remnawave squad response identity does not match request")
	}
	return &envelope.Response, nil
}
