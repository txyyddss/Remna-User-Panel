package backup

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type restoreRequest struct {
	BackupID       string
	ActorID        string
	IdempotencyKey string
	Reason         string
	Confirmation   string
	Fingerprint    string
}

func normalizeRestoreRequest(backupID, actorID, key, reason, confirmation string) (restoreRequest, error) {
	request := restoreRequest{
		BackupID:       strings.TrimSpace(backupID),
		ActorID:        strings.TrimSpace(actorID),
		IdempotencyKey: strings.TrimSpace(key),
		Reason:         strings.TrimSpace(reason),
		Confirmation:   strings.TrimSpace(confirmation),
	}
	if len(request.IdempotencyKey) < 1 || len(request.IdempotencyKey) > 128 {
		return restoreRequest{}, fmt.Errorf("%w: idempotency key must contain 1 to 128 characters", ErrRestoreConflict)
	}
	if len(request.Reason) < 4 || len(request.Reason) > 500 {
		return restoreRequest{}, fmt.Errorf("%w: restore reason must contain 4 to 500 characters", ErrRestoreConflict)
	}
	payload, err := json.Marshal(struct {
		BackupID     string `json:"backupId"`
		Reason       string `json:"reason"`
		Confirmation string `json:"confirmation"`
	}{request.BackupID, request.Reason, request.Confirmation})
	if err != nil {
		return restoreRequest{}, fmt.Errorf("fingerprint restore request: %w", err)
	}
	digest := sha256.Sum256(payload)
	request.Fingerprint = hex.EncodeToString(digest[:])
	return request, nil
}

func (s *Service) restoreReplay(ctx context.Context, request restoreRequest) (RestoreJob, bool, error) {
	job, err := s.restoreByRequestKey(ctx, request.ActorID, request.IdempotencyKey)
	if errors.Is(err, ErrBackupNotFound) {
		return RestoreJob{}, false, nil
	}
	if err != nil {
		return RestoreJob{}, false, err
	}
	if job.RequestFingerprint != request.Fingerprint {
		return RestoreJob{}, false, fmt.Errorf("%w: idempotency key is already associated with another restore request", ErrRestoreConflict)
	}
	return job, true, nil
}

func validRestoreReplayIdentity(marker restoreMarker) bool {
	if marker.RequestActorID != strings.TrimSpace(marker.RequestActorID) ||
		marker.IdempotencyKey != strings.TrimSpace(marker.IdempotencyKey) ||
		len(marker.IdempotencyKey) < 1 || len(marker.IdempotencyKey) > 128 || len(marker.RequestFingerprint) != 64 {
		return false
	}
	_, err := hex.DecodeString(marker.RequestFingerprint)
	return err == nil
}
