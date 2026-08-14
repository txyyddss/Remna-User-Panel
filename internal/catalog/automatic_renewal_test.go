package catalog

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

func TestCatalogOperationsBlockWhileAutomaticRenewalIsEnabled(t *testing.T) {
	t.Parallel()
	repository := &automaticRenewalCatalogRepository{catalogRepository: &catalogRepository{}, enabled: true}
	service := newAutomaticRenewalCatalogService(repository)
	user := model.User{ID: "member", OnboardingState: "complete"}
	if _, err := service.CatalogForUser(context.Background(), user); !errors.Is(err, ErrAutoRenewalEnabled) {
		t.Fatalf("CatalogForUser() error = %v, want ErrAutoRenewalEnabled", err)
	}
	if _, err := service.Quote(context.Background(), user, "combo", nil, ""); !errors.Is(err, ErrAutoRenewalEnabled) {
		t.Fatalf("Quote() error = %v, want ErrAutoRenewalEnabled", err)
	}
	if _, err := service.Purchase(context.Background(), user, "combo", nil, "purchase-key"); !errors.Is(err, ErrAutoRenewalEnabled) {
		t.Fatalf("Purchase() error = %v, want ErrAutoRenewalEnabled", err)
	}
}

func TestAutomaticRenewalToggleRejectsIneligibleEnablement(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	plan := database.AutoRenewalPlan{Purchase: model.Purchase{ID: "purchase", UserID: "member"}, Combo: model.Combo{ID: "combo", ValidityDays: 30,
		IncludedSquads: []model.SquadProduct{{RemnaSquadUUID: "included"}}}, GrossMinor: 100, NetMinor: 100,
		ScheduledAt: now.Add(30 * 24 * time.Hour), NextCycleEndsAt: now.Add(60 * 24 * time.Hour)}
	repository := &automaticRenewalCatalogRepository{catalogRepository: &catalogRepository{
		combos: []model.Combo{plan.Combo}, balance: model.TXBMoney(99),
	}, plan: plan}
	service := newAutomaticRenewalCatalogService(repository)
	service.now = func() time.Time { return now }
	user := model.User{ID: "member"}
	status, err := service.AutomaticRenewal(context.Background(), user, plan.Purchase.ID)
	if err != nil || status.CanEnable || status.IneligibleReason == nil || *status.IneligibleReason != database.AutoRenewalReasonInsufficientBalance {
		t.Fatalf("AutomaticRenewal() = (%+v, %v)", status, err)
	}
	if _, err := service.SetAutomaticRenewal(context.Background(), user, plan.Purchase.ID, true); !errors.Is(err, ErrAutoRenewalIneligible) {
		t.Fatalf("SetAutomaticRenewal(ineligible) = %v, want ErrAutoRenewalIneligible", err)
	}
	repository.balance = model.TXBMoney(100)
	updated, err := service.SetAutomaticRenewal(context.Background(), user, plan.Purchase.ID, true)
	if err != nil || !updated.Enabled || !repository.setEnabled {
		t.Fatalf("SetAutomaticRenewal(eligible) = (%+v, %v), stored=%t", updated, err, repository.setEnabled)
	}
}

type automaticRenewalCatalogRepository struct {
	*catalogRepository
	plan       database.AutoRenewalPlan
	enabled    bool
	setEnabled bool
}

func (r *automaticRenewalCatalogRepository) AutoRenewalPlan(context.Context, string, string, time.Time) (database.AutoRenewalPlan, error) {
	return r.plan, nil
}
func (r *automaticRenewalCatalogRepository) SetAutoRenewal(_ context.Context, _ string, _ string, enabled bool, _ time.Time) error {
	r.setEnabled = enabled
	return nil
}
func (r *automaticRenewalCatalogRepository) DueAutoRenewals(context.Context, time.Time) ([]database.DueAutoRenewal, error) {
	return nil, nil
}
func (r *automaticRenewalCatalogRepository) CommitAutoRenewal(context.Context, string, time.Time) (model.Purchase, error) {
	return model.Purchase{}, nil
}
func (r *automaticRenewalCatalogRepository) MarkAutoRenewalFailed(context.Context, string, string, time.Time) error {
	return nil
}
func (r *automaticRenewalCatalogRepository) HasEnabledAutoRenewal(context.Context, string, time.Time) (bool, error) {
	return r.enabled, nil
}

func newAutomaticRenewalCatalogService(repository *automaticRenewalCatalogRepository) *Service {
	return newCatalogServiceForTest(repository, &catalogPurchaseRemote{catalogRemnawave: &catalogRemnawave{}})
}
