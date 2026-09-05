package catalog

import (
	"context"
	"errors"
	"testing"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

func TestAutomaticRenewalRetainsRepricedHiddenSquads(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name       string
		present    bool
		balance    int64
		wantReason string
	}{
		{name: "hidden but live", present: true, balance: 1_700},
		{name: "removed upstream", balance: 1_700, wantReason: database.AutoRenewalReasonPaidAddonUnavailable},
		{name: "insufficient at current price", present: true, balance: 1_300, wantReason: database.AutoRenewalReasonInsufficientBalance},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			plan := eligibleAutoRenewalPlan()
			plan.Purchase.PriceTXBMinor = 1_300
			plan.Addons = []model.SquadProduct{{ID: "retained", RemnaSquadUUID: "retained", PriceTXBMinor: 700}}
			plan.GrossMinor, plan.NetMinor = 1_700, 1_700
			repository := dueAutoRenewalRepositoryForPlan(plan)
			repository.balance = model.TXBMoney(test.balance)
			repository.addons = []model.SquadProduct{{ID: "retained", RemnaSquadUUID: "retained", Visible: false}}
			remote := renewalTestRemote()
			if test.present {
				remote.squads = append(remote.squads, RemoteSquad{UUID: "retained", Name: "Retained"})
				remote.accessible["retained"] = []string{"node-1"}
			}
			service := newCatalogServiceForTest(repository, remote)
			user := model.User{ID: plan.Purchase.UserID}
			status, err := service.AutomaticRenewal(ctx, user, plan.Purchase.ID)
			if err != nil || status.NetPrice.MinorInt64() != 1_700 || status.CanEnable != (test.wantReason == "") {
				t.Fatalf("AutomaticRenewal() = (%+v, %v)", status, err)
			}
			if test.wantReason != "" && (status.IneligibleReason == nil || *status.IneligibleReason != test.wantReason) {
				t.Fatalf("renewal reason = %v, want %s", status.IneligibleReason, test.wantReason)
			}
			updated, err := service.SetAutomaticRenewal(ctx, user, plan.Purchase.ID, true)
			if test.wantReason == "" {
				if err != nil || !updated.Enabled || !repository.setEnabled || updated.NetPrice != status.NetPrice {
					t.Fatalf("SetAutomaticRenewal() = (%+v, %v)", updated, err)
				}
			} else if !errors.Is(err, ErrAutoRenewalIneligible) || repository.setEnabled {
				t.Fatalf("SetAutomaticRenewal(ineligible) = %v, enabled=%t", err, repository.setEnabled)
			}
			if err := service.ProcessDueAutoRenewals(ctx, service.now()); err != nil {
				t.Fatal(err)
			}
			if test.wantReason == "" {
				if len(repository.commitIDs) != 1 || len(repository.failed) != 0 {
					t.Fatalf("renewal commits=%v failures=%v", repository.commitIDs, repository.failed)
				}
			} else if len(repository.commitIDs) != 0 || len(repository.failed) != 1 || repository.failed[0].reason != test.wantReason {
				t.Fatalf("blocked renewal commits=%v failures=%v", repository.commitIDs, repository.failed)
			}
			public, err := service.Catalog(ctx)
			if err != nil || len(public.Addons) != 0 {
				t.Fatalf("public catalog exposed hidden add-on: (%+v, %v)", public, err)
			}
		})
	}
}

func TestRenewalNodesIncludeUnlistedOwnedSquads(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repository := renewalTestRepository()
	repository.quote.AddonSquadUUIDs = []string{"retained"}
	remote := renewalTestRemote()
	remote.squads = append(remote.squads, RemoteSquad{UUID: "retained", Name: "Retained"})
	remote.accessible = map[string][]string{"retained": {"node-1"}}
	service := newCatalogServiceForTest(repository, remote)
	user := model.User{ID: "user", OnboardingState: "complete"}
	quote, err := service.RenewalQuote(ctx, user, "purchase", 1)
	if err != nil || len(quote.AccessibleNodes) != 1 {
		t.Fatalf("RenewalQuote(owned, unlisted) = (%+v, %v)", quote, err)
	}
	remote.squads = remote.squads[:1]
	remote.accessible["core-squad"] = []string{"node-1"}
	if _, err := service.RenewalQuote(ctx, user, "purchase", 1); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("RenewalQuote(missing add-on) = %v, want ErrNotFound", err)
	}
}
