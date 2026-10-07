package purchaseops

import (
	"context"
	"sort"
)

// ComboControls loads the member's current switches without mirroring upstream squads.
func (s *Service) ComboControls(ctx context.Context, userID string) (ComboControls, error) {
	repository, err := s.controlsRepository()
	if err != nil {
		return ComboControls{}, err
	}
	now := s.now().UTC()
	active, queued, err := repository.ActiveAndQueuedPurchases(ctx, userID, now)
	result := ComboControls{Active: active, Queued: queued, Squads: []SquadControl{}}
	if err != nil {
		return result, err
	}
	result.Operation, err = repository.LatestComboControl(ctx, userID)
	if err != nil || active == nil {
		return result, err
	}
	conflict, err := repository.ComboControlConflict(ctx, userID)
	if err != nil {
		return result, err
	}
	result.Mutable = active.Status == "active" && !now.Before(active.ValidFrom) && now.Before(active.ValidUntil) && !conflict
	disabled, err := repository.UserDisabledSquads(ctx, userID)
	if err != nil {
		return result, err
	}
	names := map[string]string{}
	if source, ok := s.remote.(squadNameSource); ok {
		names, err = source.MemberSquadNames(ctx)
		if err != nil {
			return result, err
		}
	}
	enabled := EnabledSquads(active.SquadUUIDs, disabled)
	enabledSet := make(map[string]bool, len(enabled))
	for _, uuid := range enabled {
		enabledSet[uuid] = true
	}
	for _, uuid := range active.SquadUUIDs {
		name := names[uuid]
		if name == "" {
			name = uuid
		}
		result.Squads = append(result.Squads, SquadControl{UUID: uuid, Name: name, Enabled: enabledSet[uuid], CanDisable: result.Mutable && enabledSet[uuid] && len(enabled) > 1})
	}
	return result, nil
}

// EnabledSquads intersects account preferences with ownership, retaining one squad.
func EnabledSquads(owned, disabled []string) []string {
	blocked := make(map[string]bool, len(disabled))
	for _, uuid := range disabled {
		blocked[uuid] = true
	}
	ordered := append([]string{}, owned...)
	sort.Strings(ordered)
	enabled := make([]string, 0, len(ordered))
	for _, uuid := range ordered {
		if !blocked[uuid] {
			enabled = append(enabled, uuid)
		}
	}
	if len(enabled) == 0 && len(ordered) > 0 {
		enabled = append(enabled, ordered[0])
	}
	return enabled
}
