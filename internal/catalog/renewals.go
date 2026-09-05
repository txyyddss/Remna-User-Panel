package catalog

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

type renewalRepository interface {
	RenewalQuote(context.Context, string, string, int, time.Time) (model.RenewalQuote, error)
	Renew(context.Context, database.RenewalInput, time.Time) (model.RenewalBatch, error)
}

// RenewalQuote previews a contiguous renewal using current prices and any
// valid recurring coupon attached to the source purchase.
func (s *Service) RenewalQuote(ctx context.Context, user model.User, purchaseID string, termCount int) (model.RenewalQuote, error) {
	if user.OnboardingState != "complete" || strings.TrimSpace(purchaseID) == "" || termCount < 1 || termCount > 6 {
		return model.RenewalQuote{}, errors.New("invalid renewal selection")
	}
	repository, ok := s.repository.(renewalRepository)
	if !ok {
		return model.RenewalQuote{}, errors.New("renewal is unavailable")
	}
	quote, err := repository.RenewalQuote(ctx, user.ID, purchaseID, termCount, s.now().UTC())
	if err != nil {
		return model.RenewalQuote{}, err
	}
	catalog, reason, err := s.renewalCatalog(ctx, quote.ComboID, quote.AddonSquadUUIDs)
	if err != nil {
		return model.RenewalQuote{}, err
	}
	if reason != "" {
		return model.RenewalQuote{}, database.ErrNotFound
	}
	quote.AccessibleNodes = quoteAccessibleNodes(catalog, quote.ComboID, quote.AddonSquadUUIDs)
	if len(quote.AccessibleNodes) == 0 {
		return model.RenewalQuote{}, ErrNoAccessibleNodes
	}
	return quote, nil
}

// Renew commits one atomic debit for 1-6 contiguous current-ride terms.
func (s *Service) Renew(ctx context.Context, user model.User, purchaseID string, termCount int, idempotencyKey string) (model.RenewalBatch, error) {
	if user.OnboardingState != "complete" || strings.TrimSpace(purchaseID) == "" || termCount < 1 || termCount > 6 {
		return model.RenewalBatch{}, errors.New("invalid renewal selection")
	}
	if strings.TrimSpace(idempotencyKey) == "" || len(idempotencyKey) > 128 {
		return model.RenewalBatch{}, errors.New("invalid idempotency key")
	}
	if _, err := s.RenewalQuote(ctx, user, purchaseID, termCount); err != nil {
		return model.RenewalBatch{}, err
	}
	repository, ok := s.repository.(renewalRepository)
	if !ok {
		return model.RenewalBatch{}, errors.New("renewal is unavailable")
	}
	return repository.Renew(ctx, database.RenewalInput{UserID: user.ID, PurchaseID: purchaseID, TermCount: termCount, IdempotencyKey: idempotencyKey}, s.now().UTC())
}
