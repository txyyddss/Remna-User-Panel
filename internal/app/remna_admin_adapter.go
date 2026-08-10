package app

import (
	"context"

	"github.com/txyyddss/Remna-User-Panel/internal/admin"
	"github.com/txyyddss/Remna-User-Panel/internal/catalog"
	"github.com/txyyddss/Remna-User-Panel/internal/integrations/remnawave"
)

func (a remnaAdapter) ListInternalSquads(ctx context.Context) ([]admin.UpstreamSquad, error) {
	squads, err := a.listSquads(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]admin.UpstreamSquad, 0, len(squads))
	for _, squad := range squads {
		result = append(result, admin.UpstreamSquad{UUID: squad.UUID, Name: squad.Name})
	}
	return result, nil
}

func (a remnaAdapter) ListCatalogSquads(ctx context.Context) ([]catalog.RemoteSquad, error) {
	squads, err := a.listSquads(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]catalog.RemoteSquad, 0, len(squads))
	for _, squad := range squads {
		result = append(result, catalog.RemoteSquad{UUID: squad.UUID, Name: squad.Name})
	}
	return result, nil
}

func (a remnaAdapter) listSquads(ctx context.Context) ([]remnawave.InternalSquad, error) {
	return remnaCall(ctx, a, func(callCtx context.Context, client remnaClient) ([]remnawave.InternalSquad, error) {
		return client.ListInternalSquads(callCtx)
	})
}

func (a remnaAdapter) ListNodes(ctx context.Context) ([]admin.UpstreamNode, error) {
	nodes, err := remnaCall(ctx, a, func(callCtx context.Context, client remnaClient) ([]remnawave.Node, error) {
		return client.ListNodes(callCtx)
	})
	if err != nil {
		return nil, err
	}
	result := make([]admin.UpstreamNode, 0, len(nodes))
	for _, node := range nodes {
		inbounds := make([]string, 0, len(node.ConfigProfile.ActiveInbounds))
		for _, inbound := range node.ConfigProfile.ActiveInbounds {
			inbounds = append(inbounds, inbound.UUID)
		}
		result = append(result, admin.UpstreamNode{
			UUID: node.UUID, Name: node.Name, CountryCode: node.CountryCode,
			ConsumptionMultiplier: node.ConsumptionMultiplier,
			ActiveInboundUUIDs:    inbounds, Disabled: node.IsDisabled,
		})
	}
	return result, nil
}

func (a remnaAdapter) AccessibleNodeUUIDs(ctx context.Context, squadUUID string) ([]string, error) {
	nodes, err := remnaCall(ctx, a, func(callCtx context.Context, client remnaClient) ([]remnawave.AccessibleNode, error) {
		return client.InternalSquadAccessibleNodes(callCtx, squadUUID)
	})
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, node.UUID)
	}
	return result, nil
}

func (a remnaAdapter) UpdateInternalSquadInbounds(ctx context.Context, squadUUID string, inbounds []string) error {
	input := append([]string(nil), inbounds...)
	return remnaExecute(ctx, a, func(callCtx context.Context, client remnaClient) error {
		_, err := client.UpdateInternalSquadInbounds(callCtx, squadUUID, input)
		return err
	})
}

var _ admin.SquadImporter = remnaAdapter{}
var _ admin.SquadNodeManager = remnaAdapter{}
