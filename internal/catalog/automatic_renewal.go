package catalog

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

var ErrAutoRenewalEnabled = errors.New("automatic renewal is enabled")
var ErrAutoRenewalIneligible = errors.New("automatic renewal is currently ineligible")

type automaticRenewalRepository interface {
	AutoRenewalPlan(context.Context, string, string, time.Time) (database.AutoRenewalPlan, error)
	SetAutoRenewal(context.Context, string, string, bool, time.Time) error
	DueAutoRenewals(context.Context, time.Time) ([]database.DueAutoRenewal, error)
	CommitAutoRenewal(context.Context, string, time.Time) (model.Purchase, error)
	MarkAutoRenewalFailed(context.Context, string, string, time.Time) error
	HasEnabledAutoRenewal(context.Context, string, time.Time) (bool, error)
}

type automaticRenewalFailureRepository interface {
	AutoRenewalFailure(context.Context, string) (*model.AutoRenewalFailure, error)
}

// AutomaticRenewal returns a member's authoritative current next-cycle quote.
func (s *Service) AutomaticRenewal(ctx context.Context, user model.User, purchaseID string) (model.AutoRenewal, error) {
	return s.automaticRenewalAt(ctx, user, purchaseID, s.now().UTC())
}

func (s *Service) automaticRenewalAt(ctx context.Context, user model.User, purchaseID string, now time.Time) (model.AutoRenewal, error) {
	repository, ok := s.repository.(automaticRenewalRepository)
	if !ok || strings.TrimSpace(user.ID) == "" || strings.TrimSpace(purchaseID) == "" {
		return model.AutoRenewal{}, database.ErrNotFound
	}
	plan, err := repository.AutoRenewalPlan(ctx, user.ID, purchaseID, now)
	if err != nil {
		return model.AutoRenewal{}, err
	}
	result := autoRenewalFromPlan(plan)
	if plan.IneligibleReason == "" {
		result.IneligibleReason, err = s.autoRenewalLiveReason(ctx, plan)
		if err != nil {
			return model.AutoRenewal{}, err
		}
	}
	result.CanEnable = result.IneligibleReason == nil
	return result, nil
}

// SetAutomaticRenewal saves an eligible enablement choice or always permits disabling.
func (s *Service) SetAutomaticRenewal(ctx context.Context, user model.User, purchaseID string, enabled bool) (model.AutoRenewal, error) {
	repository, ok := s.repository.(automaticRenewalRepository)
	if !ok {
		return model.AutoRenewal{}, database.ErrNotFound
	}
	if !enabled {
		plan, err := repository.AutoRenewalPlan(ctx, user.ID, purchaseID, s.now().UTC())
		if err != nil {
			return model.AutoRenewal{}, err
		}
		if err := repository.SetAutoRenewal(ctx, user.ID, purchaseID, false, s.now().UTC()); err != nil {
			return model.AutoRenewal{}, err
		}
		status := autoRenewalFromPlan(plan)
		status.Enabled = false
		status.CanEnable = status.IneligibleReason == nil
		return status, nil
	}
	status, err := s.AutomaticRenewal(ctx, user, purchaseID)
	if err != nil {
		return model.AutoRenewal{}, err
	}
	if enabled && !status.CanEnable {
		return status, ErrAutoRenewalIneligible
	}
	if err := repository.SetAutoRenewal(ctx, user.ID, purchaseID, true, s.now().UTC()); err != nil {
		return model.AutoRenewal{}, err
	}
	status.Enabled = true
	return status, nil
}

// ProcessDueAutoRenewals revalidates due terms through the queued provider before expiry work.
func (s *Service) ProcessDueAutoRenewals(ctx context.Context, now time.Time) error {
	repository, ok := s.repository.(automaticRenewalRepository)
	if !ok {
		return nil
	}
	candidates, err := repository.DueAutoRenewals(ctx, now.UTC())
	if err != nil {
		return err
	}
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return err
		}
		status, statusErr := s.automaticRenewalAt(ctx, model.User{ID: candidate.UserID}, candidate.PurchaseID, now.UTC())
		if statusErr == nil && status.CanEnable {
			if _, err := repository.CommitAutoRenewal(ctx, candidate.PurchaseID, now.UTC()); err == nil {
				continue
			} else {
				statusErr = err
			}
		}
		reason := autoRenewalFailureReason(status, statusErr)
		if err := repository.MarkAutoRenewalFailed(ctx, candidate.PurchaseID, reason, now.UTC()); err != nil {
			return err
		}
	}
	return nil
}

// CatalogForUser returns the displayable catalog for an onboarded member.
// Checkout operations enforce the automatic-renewal restriction separately so
// dashboard metadata such as squad names and node multipliers remains visible.
func (s *Service) CatalogForUser(ctx context.Context, _ model.User) (model.Catalog, error) {
	return s.Catalog(ctx)
}

func (s *Service) ensureCatalogAvailable(ctx context.Context, userID string) error {
	repository, ok := s.repository.(automaticRenewalRepository)
	if !ok {
		return nil
	}
	enabled, err := repository.HasEnabledAutoRenewal(ctx, userID, s.now().UTC())
	if err != nil {
		return err
	}
	if enabled {
		return ErrAutoRenewalEnabled
	}
	return nil
}

func autoRenewalFromPlan(plan database.AutoRenewalPlan) model.AutoRenewal {
	result := model.AutoRenewal{PurchaseID: plan.Purchase.ID, Enabled: plan.Purchase.AutoRenewEnabled,
		GrossPrice: model.TXBMoney(plan.GrossMinor), Discount: model.TXBMoney(plan.DiscountMinor), NetPrice: model.TXBMoney(plan.NetMinor),
		ScheduledAt: plan.ScheduledAt, NextCycleEndsAt: plan.NextCycleEndsAt}
	if plan.IneligibleReason != "" {
		reason := plan.IneligibleReason
		result.IneligibleReason = &reason
	}
	return result
}

func (s *Service) autoRenewalLiveReason(ctx context.Context, plan database.AutoRenewalPlan) (*string, error) {
	addonIDs := make([]string, 0, len(plan.Addons))
	for _, addon := range plan.Addons {
		addonIDs = append(addonIDs, addon.RemnaSquadUUID)
	}
	catalog, err := s.Catalog(ctx)
	if err != nil {
		return nil, err
	}
	nodes := quoteAccessibleNodes(catalog, plan.Combo.ID, addonIDs)
	if len(nodes) == 0 {
		return autoRenewalReason(database.AutoRenewalReasonNoAccessibleNodes), nil
	}
	balance, err := s.repository.Balance(ctx, plan.Purchase.UserID)
	if err != nil {
		return nil, err
	}
	if balance.MinorInt64() < plan.NetMinor {
		return autoRenewalReason(database.AutoRenewalReasonInsufficientBalance), nil
	}
	return nil, nil
}

func autoRenewalFailureReason(status model.AutoRenewal, err error) string {
	if status.IneligibleReason != nil {
		return *status.IneligibleReason
	}
	if errors.Is(err, database.ErrInsufficientBalance) {
		return database.AutoRenewalReasonInsufficientBalance
	}
	return database.AutoRenewalReasonUnavailable
}

func autoRenewalReason(value string) *string { return &value }
