// Package catalog implements catalog browsing, TXB purchases, and dashboard composition.
package catalog

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

// Repository is the local catalog and entitlement persistence contract.
type Repository interface {
	ListCombos(context.Context, bool) ([]model.Combo, error)
	ListSquadProducts(context.Context, bool) ([]model.SquadProduct, error)
	CreatePurchase(context.Context, database.PurchaseInput, time.Time) (model.Purchase, error)
	ListPurchases(context.Context, string) ([]model.Purchase, error)
	Balance(context.Context, string) (model.Money, error)
	ActiveAndQueuedPurchases(context.Context, string, time.Time) (*model.Purchase, *model.Purchase, error)
	UpdateSubscriptionURL(context.Context, string, string) error
}

// RemoteDashboard is the safe, normalized Remnawave dashboard response.
type RemoteDashboard struct {
	Statistics      model.Statistics
	SubscriptionURL string
}

// RemnawaveClient fetches user state and revokes subscription credentials.
type RemnawaveClient interface {
	Dashboard(context.Context, string) (RemoteDashboard, error)
	RevokeSubscription(context.Context, string) (string, error)
}

// RemoteSquad is the live identity portion of a Remnawave internal squad.
type RemoteSquad struct {
	UUID string
	Name string
}

type remoteSquadLister interface {
	ListCatalogSquads(context.Context) ([]RemoteSquad, error)
}

type quoteRepository interface {
	QuotePurchase(context.Context, database.PurchaseInput, time.Time) (model.PurchaseQuote, error)
}

type cachedDashboard struct {
	value     RemoteDashboard
	fetchedAt time.Time
}

// Service owns price selection and short-lived statistics caching.
type Service struct {
	repository Repository
	remnawave  RemnawaveClient
	cacheTTL   time.Duration
	now        func() time.Time
	cacheMu    sync.RWMutex
	cache      map[string]cachedDashboard
}

// NewService creates a catalog service.
func NewService(repository Repository, remnawave RemnawaveClient, cacheTTL time.Duration) *Service {
	return &Service{repository: repository, remnawave: remnawave, cacheTTL: cacheTTL, now: time.Now, cache: make(map[string]cachedDashboard)}
}

// Catalog overlays local merchandising on the live Remnawave squad list. A
// combo with a stale included UUID is omitted until the upstream identity is
// restored, preventing checkout against a locally cached catalog.
func (s *Service) Catalog(ctx context.Context) (model.Catalog, error) {
	combos, err := s.repository.ListCombos(ctx, true)
	if err != nil {
		return model.Catalog{}, err
	}
	addons, err := s.repository.ListSquadProducts(ctx, true)
	if err != nil {
		return model.Catalog{}, err
	}
	return s.hydrateLiveCatalog(ctx, combos, addons)
}

func (s *Service) hydrateLiveCatalog(ctx context.Context, combos []model.Combo, overrides []model.SquadProduct) (model.Catalog, error) {
	lister, ok := s.remnawave.(remoteSquadLister)
	if !ok {
		return model.Catalog{Combos: combos, Addons: overrides}, nil
	}
	remote, err := lister.ListCatalogSquads(ctx)
	if err != nil {
		return model.Catalog{}, err
	}
	live := make(map[string]string, len(remote))
	for _, squad := range remote {
		live[strings.TrimSpace(squad.UUID)] = strings.TrimSpace(squad.Name)
	}
	overrideByUUID := make(map[string]model.SquadProduct, len(overrides))
	addons := make([]model.SquadProduct, 0, len(overrides))
	for _, override := range overrides {
		name, present := live[override.RemnaSquadUUID]
		if !present {
			continue
		}
		override.Name = name
		override.UpstreamPresent = true
		overrideByUUID[override.RemnaSquadUUID] = override
		if override.Visible {
			addons = append(addons, override)
		}
	}
	liveCombos := make([]model.Combo, 0, len(combos))
	for _, combo := range combos {
		included := make([]model.SquadProduct, 0, len(combo.IncludedSquads))
		valid := true
		for _, placeholder := range combo.IncludedSquads {
			name, present := live[placeholder.RemnaSquadUUID]
			if !present {
				valid = false
				break
			}
			product, hasOverride := overrideByUUID[placeholder.RemnaSquadUUID]
			if !hasOverride {
				product = model.SquadProduct{ID: placeholder.RemnaSquadUUID, RemnaSquadUUID: placeholder.RemnaSquadUUID, Visible: true}
			}
			product.Name = name
			product.UpstreamPresent = true
			included = append(included, product)
		}
		if valid {
			combo.IncludedSquads = included
			liveCombos = append(liveCombos, combo)
		}
	}
	return model.Catalog{Combos: liveCombos, Addons: addons}, nil
}

// Purchase delegates all pricing and balance work to one SQLite transaction.
func (s *Service) Purchase(ctx context.Context, user model.User, comboID string, addonIDs []string, idempotencyKey string) (model.Purchase, error) {
	return s.PurchaseWithCoupon(ctx, user, comboID, addonIDs, "", idempotencyKey)
}

// PurchaseWithCoupon applies at most one explicitly selected wallet grant.
func (s *Service) PurchaseWithCoupon(ctx context.Context, user model.User, comboID string, addonIDs []string, couponGrantID, idempotencyKey string) (model.Purchase, error) {
	if user.OnboardingState != "complete" {
		return model.Purchase{}, errors.New("onboarding is incomplete")
	}
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if comboID == "" || len(addonIDs) > 100 || idempotencyKey == "" || len(idempotencyKey) > 128 {
		return model.Purchase{}, errors.New("invalid purchase selection")
	}
	if err := s.validateLiveSelection(ctx, comboID, addonIDs); err != nil {
		return model.Purchase{}, err
	}
	return s.repository.CreatePurchase(ctx, database.PurchaseInput{UserID: user.ID, ComboID: comboID, AddonSquadIDs: addonIDs,
		CouponGrantID: couponGrantID, IdempotencyKey: idempotencyKey}, s.now().UTC())
}

// Quote returns the authoritative price and entitlement start time without
// mutating the user's balance, coupon wallet, or purchases.
func (s *Service) Quote(ctx context.Context, user model.User, comboID string, addonIDs []string, couponGrantID string) (model.PurchaseQuote, error) {
	if user.OnboardingState != "complete" {
		return model.PurchaseQuote{}, errors.New("onboarding is incomplete")
	}
	if comboID == "" || len(addonIDs) > 100 {
		return model.PurchaseQuote{}, errors.New("invalid purchase selection")
	}
	if err := s.validateLiveSelection(ctx, comboID, addonIDs); err != nil {
		return model.PurchaseQuote{}, err
	}
	repository, ok := s.repository.(quoteRepository)
	if !ok {
		return model.PurchaseQuote{}, errors.New("purchase quotes are unavailable")
	}
	return repository.QuotePurchase(ctx, database.PurchaseInput{UserID: user.ID, ComboID: comboID, AddonSquadIDs: addonIDs, CouponGrantID: couponGrantID}, s.now().UTC())
}

func (s *Service) validateLiveSelection(ctx context.Context, comboID string, addonIDs []string) error {
	if _, ok := s.remnawave.(remoteSquadLister); !ok {
		return nil
	}
	catalog, err := s.Catalog(ctx)
	if err != nil {
		return err
	}
	comboFound := false
	for _, combo := range catalog.Combos {
		if combo.ID == comboID {
			comboFound = true
			break
		}
	}
	if !comboFound {
		return database.ErrNotFound
	}
	visible := make(map[string]struct{}, len(catalog.Addons))
	for _, addon := range catalog.Addons {
		visible[addon.RemnaSquadUUID] = struct{}{}
	}
	for _, addonID := range addonIDs {
		if _, exists := visible[strings.TrimSpace(addonID)]; !exists {
			return database.ErrNotFound
		}
	}
	return nil
}

// Purchases returns the account's immutable purchase history.
func (s *Service) Purchases(ctx context.Context, userID string) ([]model.Purchase, error) {
	return s.repository.ListPurchases(ctx, userID)
}

// Dashboard prefers fresh upstream data and falls back to a clearly marked cached snapshot.
func (s *Service) Dashboard(ctx context.Context, user model.User) (model.Dashboard, error) {
	now := s.now().UTC()
	balance, err := s.repository.Balance(ctx, user.ID)
	if err != nil {
		return model.Dashboard{}, err
	}
	active, queued, err := s.repository.ActiveAndQueuedPurchases(ctx, user.ID, now)
	if err != nil {
		return model.Dashboard{}, err
	}
	dashboard := model.Dashboard{User: user, Balance: balance, ActivePurchase: active, QueuedPurchase: queued, SubscriptionURL: user.RemnaSubscriptionURL, FetchedAt: now}
	if user.RemnaUserID == nil {
		return dashboard, nil
	}
	s.cacheMu.RLock()
	cached, exists := s.cache[user.ID]
	s.cacheMu.RUnlock()
	if exists && now.Sub(cached.fetchedAt) <= s.cacheTTL {
		dashboard.Statistics = &cached.value.Statistics
		dashboard.SubscriptionURL = &cached.value.SubscriptionURL
		dashboard.FetchedAt = cached.fetchedAt
		return dashboard, nil
	}
	remote, remoteErr := s.remnawave.Dashboard(ctx, *user.RemnaUserID)
	if remoteErr == nil {
		dashboard.Statistics = &remote.Statistics
		dashboard.SubscriptionURL = &remote.SubscriptionURL
		s.cacheMu.Lock()
		s.cache[user.ID] = cachedDashboard{value: remote, fetchedAt: now}
		s.cacheMu.Unlock()
		return dashboard, nil
	}
	if !exists {
		dashboard.StatisticsStale = true
		dashboard.StatisticsWarning = "Live usage is temporarily unavailable. Local balance and entitlement data are still current."
		return dashboard, nil
	}
	dashboard.Statistics = &cached.value.Statistics
	dashboard.SubscriptionURL = &cached.value.SubscriptionURL
	dashboard.StatisticsStale = true
	dashboard.StatisticsWarning = "Live usage is unavailable. Showing the latest cached data."
	dashboard.FetchedAt = cached.fetchedAt
	return dashboard, nil
}

// RevokeSubscription rotates the bearer URL and immediately updates the local secret cache.
func (s *Service) RevokeSubscription(ctx context.Context, user model.User) (string, error) {
	if user.RemnaUserID == nil {
		return "", database.ErrNotFound
	}
	url, err := s.remnawave.RevokeSubscription(ctx, *user.RemnaUserID)
	if err != nil {
		return "", err
	}
	if err := s.repository.UpdateSubscriptionURL(ctx, user.ID, url); err != nil {
		return "", err
	}
	s.cacheMu.Lock()
	delete(s.cache, user.ID)
	s.cacheMu.Unlock()
	return url, nil
}
