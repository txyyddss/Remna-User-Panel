package statistics

import "github.com/txyyddss/Remna-User-Panel/internal/model"

func (s *Service) cachedSquadNames() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]string, len(s.snapshot.Remote.SquadNames))
	for uuid, name := range s.snapshot.Remote.SquadNames {
		result[uuid] = name
	}
	return result
}

func applySquadNames(statistics *model.DatabaseStatistics, names map[string]string) {
	if statistics == nil || len(names) == 0 {
		return
	}
	for groupIndex := range statistics.SquadByCombo {
		for segmentIndex := range statistics.SquadByCombo[groupIndex].Segments {
			segment := &statistics.SquadByCombo[groupIndex].Segments[segmentIndex]
			if name := names[segment.ID]; name != "" {
				segment.Label = name
			}
		}
	}
	for groupIndex := range statistics.ComboBySquad {
		group := &statistics.ComboBySquad[groupIndex]
		if name := names[group.ID]; name != "" {
			group.Label = name
		}
	}
}
