package admin

import (
	"context"
	"errors"
	"strings"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

// PreviewBulkExtension counts current active targets and all queued successors.
func (s *UserWorkflows) PreviewBulkExtension(ctx context.Context, request BulkExtension) (database.AdminBulkExtensionPreview, error) {
	filter, err := s.normalizeBulkFilter(ctx, request)
	if err != nil {
		return database.AdminBulkExtensionPreview{}, err
	}
	if request.Days < 1 || request.Days > 3650 {
		return database.AdminBulkExtensionPreview{}, errors.New("extension days must be between 1 and 3650")
	}
	return s.repository.PreviewAdminBulkExtension(ctx, filter, s.now().UTC())
}

// CreateBulkExtension shifts each deduplicated active user exactly once.
func (s *UserWorkflows) CreateBulkExtension(ctx context.Context, actorID, key string, request BulkExtension) (model.OperationReceipt, error) {
	request.Reason = strings.TrimSpace(request.Reason)
	if !validCommand(actorID, key, request.Reason) || request.Days < 1 || request.Days > 3650 {
		return model.OperationReceipt{}, errors.New("invalid bulk extension")
	}
	filter, err := s.normalizeBulkFilter(ctx, request)
	if err != nil {
		return model.OperationReceipt{}, err
	}
	fingerprint, err := commandFingerprint(struct {
		Filter database.AdminBulkExtensionFilter
		Days   int
		Reason string
	}{filter, request.Days, request.Reason})
	if err != nil {
		return model.OperationReceipt{}, err
	}
	return s.repository.CreateAdminBulkExtension(ctx, database.AdminBulkExtensionInput{
		ActorUserID: actorID, IdempotencyKey: strings.TrimSpace(key), RequestFingerprint: fingerprint,
		Reason: request.Reason, Filter: filter, Days: request.Days,
	}, s.now().UTC())
}

func (s *UserWorkflows) normalizeBulkFilter(ctx context.Context, request BulkExtension) (database.AdminBulkExtensionFilter, error) {
	comboIDs, err := normalizeIDs(request.ComboIDs)
	if err != nil {
		return database.AdminBulkExtensionFilter{}, err
	}
	addons, err := normalizeUUIDs(request.AddonSquadUUIDs)
	if err != nil {
		return database.AdminBulkExtensionFilter{}, err
	}
	if len(comboIDs) == 0 && len(addons) == 0 {
		return database.AdminBulkExtensionFilter{}, errors.New("at least one bulk extension filter is required")
	}
	for _, comboID := range comboIDs {
		if _, err := s.repository.ComboByID(ctx, comboID, false); err != nil {
			return database.AdminBulkExtensionFilter{}, err
		}
	}
	return database.AdminBulkExtensionFilter{ComboIDs: comboIDs, AddonSquadUUIDs: addons}, nil
}
