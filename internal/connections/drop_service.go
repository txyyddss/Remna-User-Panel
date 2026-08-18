package connections

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/providerops"
)

// SecretBox seals short-lived capabilities for the durable outbox only.
type SecretBox interface {
	Seal(context string, plaintext []byte) (string, error)
	Open(context string, ciphertext string) ([]byte, error)
}

// DropRepository persists hashed targets and owner-scoped operation receipts.
type DropRepository interface {
	CreateProviderOperation(context.Context, providerops.CreateInput, time.Time) (providerops.Operation, bool, error)
	ProviderOperationForActorKey(context.Context, string, string, string, string) (model.OperationReceipt, bool, error)
	ConnectionScanForUser(context.Context, string, string) (providerops.ConnectionScan, error)
}

// DropService verifies a capability and queues one selected-IP unlink.
type DropService struct {
	repository DropRepository
	signer     *Signer
	secrets    SecretBox
	now        func() time.Time
}

// NewDropService creates the durable connection-drop command service.
func NewDropService(repository DropRepository, signer *Signer, secrets SecretBox) *DropService {
	return &DropService{repository: repository, signer: signer, secrets: secrets, now: time.Now}
}

// Drop creates or replays one durable connection unlink operation.
func (s *DropService) Drop(ctx context.Context, userID, handle, idempotencyKey string) (model.OperationReceipt, error) {
	handle = strings.TrimSpace(handle)
	fingerprint := dropFingerprint(handle)
	if receipt, found, err := s.repository.ProviderOperationForActorKey(ctx, userID, DropOperationKind, idempotencyKey, fingerprint); found || err != nil {
		return receipt, err
	}
	now := s.now().UTC()
	claims, err := s.signer.Verify(handle, userID, now)
	if err != nil {
		return model.OperationReceipt{}, err
	}
	scan, err := s.repository.ConnectionScanForUser(ctx, claims.ScanID, userID)
	if err != nil || scan.Status != providerops.StatusSucceeded || !now.Before(scan.ExpiresAt) {
		return model.OperationReceipt{}, ErrScanNotFound
	}
	sealed, err := s.secrets.Seal(dropSecretContext(userID, fingerprint), []byte(handle))
	if err != nil {
		return model.OperationReceipt{}, err
	}
	operation, _, err := s.repository.CreateProviderOperation(ctx, providerops.CreateInput{
		ActorUserID: userID, OwnerUserID: userID, Kind: DropOperationKind, IdempotencyKey: strings.TrimSpace(idempotencyKey),
		RequestFingerprint: fingerprint, SealedTarget: sealed,
		Items: []providerops.ItemInput{{Key: "connection", TargetType: "connection_handle_sha256", TargetID: hash(handle)}},
	}, now)
	return operation.Receipt, err
}

func dropFingerprint(handle string) string { return hash("connection-drop:v1:" + handle) }

func hash(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func dropSecretContext(userID, fingerprint string) string {
	return "connection-drop:" + userID + ":" + fingerprint
}
