package compensation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
)

type Repository interface {
	CompensationConfig(context.Context) (Config, error)
	UpdateCompensationConfig(context.Context, string, ConfigUpdate, time.Time) (Config, error)
	RecordCompensationObservation(context.Context, Config, []Node, time.Time) error
	ListCompensationEvents(context.Context, string, string, int) (EventPage, error)
	ApproveCompensationEvent(context.Context, ReviewInput, string, time.Time) (Event, error)
	DismissCompensationEvent(context.Context, ReviewInput, string, time.Time) (Event, error)
}

type Provider interface {
	CompensationNodes(context.Context) ([]Node, error)
	CompensationSquads(context.Context) ([]Squad, error)
}

type Service struct {
	repository Repository
	provider   Provider
}

func NewService(repository Repository, provider Provider) *Service {
	return &Service{repository: repository, provider: provider}
}

func (s *Service) Config(ctx context.Context) (Config, error) {
	return s.repository.CompensationConfig(ctx)
}

func (s *Service) UpdateConfig(ctx context.Context, actorID string, input ConfigUpdate, now time.Time) (Config, error) {
	if input.Revision < 0 || !validOptional(input.ThresholdMinutes, 1, MaxMinutes) ||
		!validOptional(input.MultiplierBPS, MinMultiplierBPS, MaxMultiplierBPS) ||
		(input.Enabled && (input.ThresholdMinutes == nil || input.MultiplierBPS == nil)) {
		return Config{}, ErrInvalid
	}
	return s.repository.UpdateCompensationConfig(ctx, actorID, input, now.UTC())
}

func (s *Service) Events(ctx context.Context, status, cursor string, limit int) (EventPage, error) {
	if limit < 1 || limit > 100 || !validStatusFilter(status) {
		return EventPage{}, ErrInvalid
	}
	return s.repository.ListCompensationEvents(ctx, status, cursor, limit)
}

func (s *Service) Approve(ctx context.Context, input ReviewInput, now time.Time) (Event, error) {
	if err := validateReview(input, true); err != nil {
		return Event{}, err
	}
	return s.repository.ApproveCompensationEvent(ctx, input, reviewFingerprint("approve", input), now.UTC())
}

func (s *Service) Dismiss(ctx context.Context, input ReviewInput, now time.Time) (Event, error) {
	if err := validateReview(input, false); err != nil {
		return Event{}, err
	}
	return s.repository.DismissCompensationEvent(ctx, input, reviewFingerprint("dismiss", input), now.UTC())
}

func validateReview(input ReviewInput, approve bool) error {
	if strings.TrimSpace(input.EventID) == "" || strings.TrimSpace(input.ActorUserID) == "" ||
		strings.TrimSpace(input.IdempotencyKey) == "" || strings.TrimSpace(input.Reason) == "" || input.Revision < 0 {
		return ErrInvalid
	}
	if approve && (input.ExtensionMinutes < 1 || input.ExtensionMinutes > MaxMinutes) {
		return ErrInvalid
	}
	return nil
}

func reviewFingerprint(action string, input ReviewInput) string {
	payload, _ := json.Marshal([]any{action, input.EventID, input.Revision, input.ExtensionMinutes, strings.TrimSpace(input.Reason)})
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func validOptional(value *int, min, max int) bool {
	return value == nil || (*value >= min && *value <= max)
}

func validStatusFilter(status string) bool {
	switch status {
	case "", "observing", "pending_review", "queued", "dismissed", "ineligible":
		return true
	default:
		return false
	}
}
