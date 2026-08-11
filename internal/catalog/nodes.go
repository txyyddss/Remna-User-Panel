package catalog

import (
	"context"
	"sort"
	"strings"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

// RemoteNode is the non-sensitive node data needed to preview a purchase.
type RemoteNode struct {
	UUID                  string
	Name                  string
	CountryCode           string
	ConsumptionMultiplier float64
	ActiveInboundUUIDs    []string
	Disabled              bool
}

type remoteNodeLister interface {
	ListCatalogNodes(context.Context) ([]RemoteNode, error)
	AccessibleCatalogNodeUUIDs(context.Context, string) ([]string, error)
}

func (s *Service) quoteAccessibleNodes(ctx context.Context, comboID string, addonIDs []string) ([]model.RemnaNode, error) {
	provider, ok := s.remnawave.(remoteNodeLister)
	if !ok {
		return []model.RemnaNode{}, nil
	}
	catalog, err := s.Catalog(ctx)
	if err != nil {
		return nil, err
	}
	squadUUIDs := selectedSquadUUIDs(catalog, comboID, addonIDs)
	accessible := make(map[string]struct{})
	for _, squadUUID := range squadUUIDs {
		nodes, lookupErr := provider.AccessibleCatalogNodeUUIDs(ctx, squadUUID)
		if lookupErr != nil {
			return nil, lookupErr
		}
		for _, node := range nodes {
			accessible[node] = struct{}{}
		}
	}
	nodes, err := provider.ListCatalogNodes(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]model.RemnaNode, 0, len(accessible))
	for _, node := range nodes {
		if node.Disabled {
			continue
		}
		if _, ok := accessible[node.UUID]; !ok {
			continue
		}
		result = append(result, model.RemnaNode{UUID: node.UUID, Name: node.Name, CountryCode: node.CountryCode,
			ConsumptionMultiplier: node.ConsumptionMultiplier, ActiveInboundUUIDs: node.ActiveInboundUUIDs, Accessible: true})
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].CountryCode == result[right].CountryCode {
			return result[left].Name < result[right].Name
		}
		return result[left].CountryCode < result[right].CountryCode
	})
	return result, nil
}

func selectedSquadUUIDs(catalog model.Catalog, comboID string, addonIDs []string) []string {
	selected := make(map[string]struct{})
	for _, combo := range catalog.Combos {
		if combo.ID != comboID {
			continue
		}
		for _, squad := range combo.IncludedSquads {
			selected[squad.RemnaSquadUUID] = struct{}{}
		}
		break
	}
	for _, addonID := range addonIDs {
		trimmed := strings.TrimSpace(addonID)
		for _, addon := range catalog.Addons {
			if addon.ID == trimmed || addon.RemnaSquadUUID == trimmed {
				selected[addon.RemnaSquadUUID] = struct{}{}
				break
			}
		}
	}
	result := make([]string, 0, len(selected))
	for squadUUID := range selected {
		result = append(result, squadUUID)
	}
	sort.Strings(result)
	return result
}
