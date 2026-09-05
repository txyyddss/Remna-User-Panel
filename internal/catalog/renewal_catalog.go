package catalog

import (
	"context"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

// renewalCatalog hydrates only an owned renewal selection. Retained add-ons
// bypass storefront visibility, while every squad must still exist upstream.
// The provider adapter keeps all live reads behind the existing queue.
func (s *Service) renewalCatalog(ctx context.Context, comboID string, addonIDs []string) (model.Catalog, string, error) {
	combos, err := s.repository.ListCombos(ctx, true)
	if err != nil {
		return model.Catalog{}, "", err
	}
	selectedCombos := make([]model.Combo, 0, 1)
	for _, combo := range combos {
		if combo.ID == comboID {
			selectedCombos = append(selectedCombos, combo)
			break
		}
	}
	addons := make([]model.SquadProduct, 0, len(addonIDs))
	for _, id := range addonIDs {
		addons = append(addons, model.SquadProduct{ID: id, RemnaSquadUUID: id, Visible: true})
	}
	catalog, err := s.hydrateLiveCatalog(ctx, selectedCombos, addons)
	if err != nil {
		return model.Catalog{}, "", err
	}
	if len(catalog.Combos) == 0 {
		return catalog, database.AutoRenewalReasonComboUnavailable, nil
	}
	liveAddons := make(map[string]bool, len(catalog.Addons))
	for _, addon := range catalog.Addons {
		liveAddons[addon.RemnaSquadUUID] = true
	}
	for _, id := range addonIDs {
		if !liveAddons[id] {
			return catalog, database.AutoRenewalReasonPaidAddonUnavailable, nil
		}
	}
	return catalog, "", nil
}
