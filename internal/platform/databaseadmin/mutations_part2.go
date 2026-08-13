package databaseadmin

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

func mutationDigest(actor string, request MutationRequest) (string, error) {
	request.ReviewHash = ""
	request.Confirmation = ""
	payload := struct {
		Actor   string          `json:"actor"`
		Request MutationRequest `json:"request"`
	}{Actor: actor, Request: request}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode database mutation review: %w", err)
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}

func (s *Service) checkReview(ctx context.Context, actor string, request MutationRequest, digest string) error {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM database_admin_reviews
		WHERE id=? AND actor_user_id=? AND action=? AND table_name=? AND request_hash=? AND consumed_at IS NULL AND expires_at>=?`,
		request.ReviewHash, actor, request.Action, request.Table, digest, s.now().UTC().Format(time.RFC3339Nano)).Scan(&count)
	if err != nil {
		return fmt.Errorf("verify database review: %w", err)
	}
	if count != 1 {
		return ErrReviewConflict
	}
	return nil
}

func consumeReview(ctx context.Context, tx *sql.Tx, actor string, request MutationRequest, digest string, now time.Time) error {
	result, err := tx.ExecContext(ctx, `UPDATE database_admin_reviews SET consumed_at=?
		WHERE id=? AND actor_user_id=? AND action=? AND table_name=? AND request_hash=? AND consumed_at IS NULL AND expires_at>=?`,
		now.Format(time.RFC3339Nano), request.ReviewHash, actor, request.Action, request.Table, digest, now.Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("consume database review: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count consumed database review: %w", err)
	}
	if affected != 1 {
		return ErrReviewConflict
	}
	return nil
}

