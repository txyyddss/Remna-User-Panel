package admin

import (
	"context"
	"errors"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	"sort"
	"strings"
)

// SaveCombo validates a complete catalog item.
func (s *Service) SaveCombo(ctx context.Context, actorID string, input database.ComboInput) (model.Combo, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || input.PriceTXBMinor < 0 || input.PriceTXBMinor > 1_000_000_000_000 || input.ValidityDays < 1 || input.ValidityDays > 3650 || input.TrafficLimitBytes <= 0 ||
		input.RolloverMinRemainingBPS < 0 || input.RolloverMinRemainingBPS > 10_000 || input.RolloverMaxTXBMinor < 0 || input.RolloverMaxTXBMinor > 1_000_000_000_000 {
		return model.Combo{}, errors.New("invalid combo fields")
	}
	if input.ResetStrategy != "DAY" && input.ResetStrategy != "WEEK" && input.ResetStrategy != "MONTH" && input.ResetStrategy != "MONTH_ROLLING" {
		return model.Combo{}, errors.New("invalid reset strategy")
	}
	live, err := s.liveSquads(ctx)
	if err != nil {
		return model.Combo{}, err
	}
	for _, squadUUID := range input.SquadProductIDs {
		if _, exists := live[strings.TrimSpace(squadUUID)]; !exists {
			return model.Combo{}, database.ErrNotFound
		}
	}
	combo, err := s.repository.SaveCombo(ctx, input)
	if err != nil {
		return model.Combo{}, err
	}
	if err := s.audit(ctx, actorID, "combo.save", "combo", combo.ID, map[string]any{"name": combo.Name}); err != nil {
		return model.Combo{}, err
	}
	return combo, nil
}

// DeleteCombo hides the selected combo.
func (s *Service) DeleteCombo(ctx context.Context, actorID, comboID string) error {
	if err := s.repository.DeleteCombo(ctx, comboID); err != nil {
		return err
	}
	return s.audit(ctx, actorID, "combo.hide", "combo", comboID, nil)
}

// SaveSquadProduct validates local merchandising data.
func (s *Service) SaveSquadProduct(ctx context.Context, actorID string, input database.SquadProductInput) (model.SquadProduct, error) {
	input.RemnaSquadUUID = strings.TrimSpace(input.RemnaSquadUUID)
	input.Name = strings.TrimSpace(input.Name)
	if input.RemnaSquadUUID == "" || input.PriceTXBMinor < 0 || input.PriceTXBMinor > 1_000_000_000_000 || (input.StockLimit != nil && (*input.StockLimit < 0 || *input.StockLimit > 1_000_000_000)) {
		return model.SquadProduct{}, errors.New("invalid squad product")
	}
	live, err := s.liveSquads(ctx)
	if err != nil {
		return model.SquadProduct{}, err
	}
	upstreamName, exists := live[input.RemnaSquadUUID]
	if !exists {
		return model.SquadProduct{}, database.ErrNotFound
	}
	input.ID = input.RemnaSquadUUID
	input.Name = upstreamName
	input.UpstreamPresent = true
	product, err := s.repository.SaveSquadProduct(ctx, input)
	if err != nil {
		return model.SquadProduct{}, err
	}
	product.Name = upstreamName
	product.UpstreamPresent = true
	if err := s.audit(ctx, actorID, "squad_product.save", "squad_product", product.ID, map[string]any{"name": upstreamName}); err != nil {
		return model.SquadProduct{}, err
	}
	return product, nil
}

// ImportSquads now returns a live upstream list overlaid with the sparse local
// overrides. The retained import call is an audited compatibility endpoint and
// no longer stores a second catalog identity table.
func (s *Service) ImportSquads(ctx context.Context, actorID string) ([]model.SquadProduct, error) {
	upstream, err := s.importer.ListInternalSquads(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.audit(ctx, actorID, "squads.import", "catalog", "squads", map[string]any{"count": len(upstream)}); err != nil {
		return nil, err
	}
	return s.Squads(ctx)
}

// Squads builds the admin catalog from live Remnawave identities plus sparse
// local merchandising overrides.
func (s *Service) Squads(ctx context.Context) ([]model.SquadProduct, error) {
	upstream, err := s.importer.ListInternalSquads(ctx)
	if err != nil {
		return nil, err
	}
	overrides, err := s.repository.ListSquadProducts(ctx, false)
	if err != nil {
		return nil, err
	}
	overrideByUUID := make(map[string]model.SquadProduct, len(overrides))
	for _, override := range overrides {
		overrideByUUID[override.RemnaSquadUUID] = override
	}
	products := make([]model.SquadProduct, 0, len(upstream))
	for _, squad := range upstream {
		product, exists := overrideByUUID[squad.UUID]
		if !exists {
			product = model.SquadProduct{ID: squad.UUID, RemnaSquadUUID: squad.UUID, Visible: false, Price: model.TXBMoney(0)}
		}
		product.ID = squad.UUID
		product.RemnaSquadUUID = squad.UUID
		product.Name = squad.Name
		product.UpstreamPresent = true
		products = append(products, product)
	}
	return products, nil
}

func (s *Service) liveSquads(ctx context.Context) (map[string]string, error) {
	upstream, err := s.importer.ListInternalSquads(ctx)
	if err != nil {
		return nil, err
	}
	live := make(map[string]string, len(upstream))
	for _, squad := range upstream {
		live[strings.TrimSpace(squad.UUID)] = strings.TrimSpace(squad.Name)
	}
	return live, nil
}

// SquadNodes returns the selectable nodes together with Remnawave's actual
// current accessibility for the squad.
func (s *Service) SquadNodes(ctx context.Context, squadUUID string) ([]model.RemnaNode, error) {
	manager, ok := s.importer.(SquadNodeManager)
	if !ok {
		return nil, errors.New("Remnawave node management is unavailable")
	}
	live, err := s.liveSquads(ctx)
	if err != nil {
		return nil, err
	}
	if _, exists := live[strings.TrimSpace(squadUUID)]; !exists {
		return nil, database.ErrNotFound
	}
	nodes, err := manager.ListNodes(ctx)
	if err != nil {
		return nil, err
	}
	accessibleUUIDs, err := manager.AccessibleNodeUUIDs(ctx, squadUUID)
	if err != nil {
		return nil, err
	}
	accessible := make(map[string]struct{}, len(accessibleUUIDs))
	for _, nodeUUID := range accessibleUUIDs {
		accessible[nodeUUID] = struct{}{}
	}
	result := make([]model.RemnaNode, 0, len(nodes))
	for _, node := range nodes {
		_, isAccessible := accessible[node.UUID]
		result = append(result, model.RemnaNode{UUID: node.UUID, Name: node.Name, CountryCode: normalizedCountryCode(node.CountryCode),
			ConsumptionMultiplier: node.ConsumptionMultiplier, ActiveInboundUUIDs: append([]string(nil), node.ActiveInboundUUIDs...), Accessible: isAccessible})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name == result[j].Name {
			return result[i].UUID < result[j].UUID
		}
		return result[i].Name < result[j].Name
	})
	return result, nil
}

// UpdateSquadNodes validates node UUIDs, unions their current active inbounds,
// patches Remnawave, then re-fetches accessibility instead of assuming the
// selected node set equals the resulting state.
func (s *Service) UpdateSquadNodes(ctx context.Context, actorID, squadUUID string, selectedNodeUUIDs []string) ([]model.RemnaNode, error) {
	manager, ok := s.importer.(SquadNodeManager)
	if !ok {
		return nil, errors.New("Remnawave node management is unavailable")
	}
	live, err := s.liveSquads(ctx)
	if err != nil {
		return nil, err
	}
	if _, exists := live[strings.TrimSpace(squadUUID)]; !exists {
		return nil, database.ErrNotFound
	}
	nodes, err := manager.ListNodes(ctx)
	if err != nil {
		return nil, err
	}
	byUUID := make(map[string]UpstreamNode, len(nodes))
	for _, node := range nodes {
		byUUID[node.UUID] = node
	}
	inboundSet := make(map[string]struct{})
	selectedSet := make(map[string]struct{}, len(selectedNodeUUIDs))
	for _, rawUUID := range selectedNodeUUIDs {
		nodeUUID := strings.TrimSpace(rawUUID)
		if _, duplicate := selectedSet[nodeUUID]; duplicate {
			continue
		}
		node, exists := byUUID[nodeUUID]
		if !exists || node.Disabled {
			return nil, database.ErrNotFound
		}
		selectedSet[nodeUUID] = struct{}{}
		for _, inboundUUID := range node.ActiveInboundUUIDs {
			inboundSet[inboundUUID] = struct{}{}
		}
	}
	inbounds := make([]string, 0, len(inboundSet))
	for inboundUUID := range inboundSet {
		inbounds = append(inbounds, inboundUUID)
	}
	sort.Strings(inbounds)
	if err := manager.UpdateInternalSquadInbounds(ctx, squadUUID, inbounds); err != nil {
		return nil, err
	}
	if err := s.audit(ctx, actorID, "squad.nodes.update", "squad_product", squadUUID, map[string]any{"selectedNodeUuids": selectedNodeUUIDs}); err != nil {
		return nil, err
	}
	return s.SquadNodes(ctx, squadUUID)
}

func normalizedCountryCode(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if len(value) != 2 || value[0] < 'A' || value[0] > 'Z' || value[1] < 'A' || value[1] > 'Z' {
		return ""
	}
	return value
}
