package catalog

import (
	"context"
	"strings"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

// RemoteSquad is the live identity portion of a Remnawave internal squad.
type RemoteSquad struct {
	UUID string
	Name string
}

type remoteSquadLister interface {
	ListCatalogSquads(context.Context) ([]RemoteSquad, error)
}

// Catalog overlays local merchandising on the live Remnawave squad list. A
// combo with a stale included UUID is omitted until the upstream identity is
// restored, preventing checkout against a locally cached catalog.
func (s *Service) Catalog(ctx context.Context) (model.Catalog, error) {
	combos, err := s.repository.ListCombos(ctx, true)
	if err != nil {
		return model.Catalog{}, err
	}
	addons, err := s.repository.ListSquadProducts(ctx, true)
	if err != nil {
		return model.Catalog{}, err
	}
	return s.hydrateLiveCatalog(ctx, combos, addons)
}

func (s *Service) hydrateLiveCatalog(ctx context.Context, combos []model.Combo, overrides []model.SquadProduct) (model.Catalog, error) {
	lister, ok := s.remnawave.(remoteSquadLister)
	if !ok {
		return model.Catalog{Combos: combos, Addons: overrides, Nodes: []model.CatalogNode{}}, nil
	}
	remote, err := lister.ListCatalogSquads(ctx)
	if err != nil {
		return model.Catalog{}, err
	}
	live := make(map[string]string, len(remote))
	for _, squad := range remote {
		live[strings.TrimSpace(squad.UUID)] = strings.TrimSpace(squad.Name)
	}
	overrideByUUID := make(map[string]model.SquadProduct, len(overrides))
	addons := make([]model.SquadProduct, 0, len(overrides))
	for _, override := range overrides {
		name, present := live[override.RemnaSquadUUID]
		if !present {
			continue
		}
		override.Name = name
		override.UpstreamPresent = true
		overrideByUUID[override.RemnaSquadUUID] = override
		if override.Visible {
			addons = append(addons, override)
		}
	}
	liveCombos := make([]model.Combo, 0, len(combos))
	for _, combo := range combos {
		included := make([]model.SquadProduct, 0, len(combo.IncludedSquads))
		valid := true
		for _, placeholder := range combo.IncludedSquads {
			name, present := live[placeholder.RemnaSquadUUID]
			if !present {
				valid = false
				break
			}
			product, hasOverride := overrideByUUID[placeholder.RemnaSquadUUID]
			if !hasOverride {
				product = model.SquadProduct{ID: placeholder.RemnaSquadUUID, RemnaSquadUUID: placeholder.RemnaSquadUUID, Visible: true}
			}
			product.Name = name
			product.UpstreamPresent = true
			included = append(included, product)
		}
		if valid {
			combo.IncludedSquads = included
			liveCombos = append(liveCombos, combo)
		}
	}
	catalogNodes := make([]model.CatalogNode, 0)
	if nodeLister, ok := s.remnawave.(remoteNodeLister); ok {
		if remoteNodes, nodeErr := nodeLister.ListCatalogNodes(ctx); nodeErr == nil {
			catalogNodes = projectCatalogNodes(remoteNodes)
		}
	}
	return model.Catalog{Combos: liveCombos, Addons: addons, Nodes: catalogNodes}, nil
}
