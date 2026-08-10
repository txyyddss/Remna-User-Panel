package entitlements

import (
	"context"
	"errors"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"testing"
	"time"
)

func TestWorkerProcessApplyBranches(t *testing.T) {
	t.Parallel()

	remoteID := "remote-1"
	testError := errors.New("test failure")
	future := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		repository *entitlementRepository
		remote     *entitlementRemnawave
		wantError  bool
		wantApply  bool
	}{
		{name: "cancelled purchase", repository: &entitlementRepository{purchase: model.Purchase{Status: "cancelled"}}, remote: &entitlementRemnawave{}},
		{name: "expired purchase", repository: &entitlementRepository{purchase: model.Purchase{Status: "expired"}}, remote: &entitlementRemnawave{}},
		{name: "purchase lookup", repository: &entitlementRepository{purchaseErr: testError}, remote: &entitlementRemnawave{}, wantError: true},
		{name: "user lookup", repository: &entitlementRepository{purchase: model.Purchase{ID: "purchase", ValidUntil: future}, userForPurchaseErr: testError}, remote: &entitlementRemnawave{}, wantError: true},
		{name: "missing remote identity", repository: &entitlementRepository{purchase: model.Purchase{ID: "purchase", ValidUntil: future}, user: model.User{}}, remote: &entitlementRemnawave{}, wantError: true},
		{name: "reset phase lookup", repository: &entitlementRepository{purchase: model.Purchase{ID: "purchase", ValidUntil: future}, user: model.User{RemnaUserID: &remoteID}, resetPhaseErr: testError}, remote: &entitlementRemnawave{}, wantError: true},
		{name: "quiesce", repository: &entitlementRepository{purchase: model.Purchase{ID: "purchase", ValidUntil: future}, user: model.User{RemnaUserID: &remoteID}}, remote: &entitlementRemnawave{removeErr: testError}, wantError: true},
		{name: "reset traffic", repository: &entitlementRepository{purchase: model.Purchase{ID: "purchase", ValidUntil: future}, user: model.User{RemnaUserID: &remoteID}, resetPhase: "quiesced"}, remote: &entitlementRemnawave{resetErr: testError}, wantError: true},
		{name: "unknown reset phase", repository: &entitlementRepository{purchase: model.Purchase{ID: "purchase", ValidUntil: future}, user: model.User{RemnaUserID: &remoteID}, resetPhase: "mystery"}, remote: &entitlementRemnawave{}, wantError: true},
		{name: "remote apply", repository: &entitlementRepository{purchase: model.Purchase{ID: "purchase", ValidUntil: future}, user: model.User{RemnaUserID: &remoteID}}, remote: &entitlementRemnawave{applyErr: testError}, wantError: true, wantApply: true},
		{name: "mark result", repository: &entitlementRepository{purchase: model.Purchase{ID: "purchase", ValidUntil: future}, user: model.User{RemnaUserID: &remoteID}, markErr: testError}, remote: &entitlementRemnawave{}, wantError: true, wantApply: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := newEntitlementWorkerForTest(test.repository, test.remote).process(context.Background(), model.OutboxJob{Kind: "remna_apply_entitlement", Payload: `{"purchaseId":"purchase"}`})
			if test.wantError && err == nil {
				t.Fatal("process() unexpectedly succeeded")
			}
			if !test.wantError && err != nil {
				t.Fatalf("process(): %v", err)
			}
			if got := test.remote.applyCalls > 0; got != test.wantApply {
				t.Fatalf("ApplyEntitlement called = %t, want %t", got, test.wantApply)
			}
		})
	}
}

func TestWorkerProcessExpiresDuePurchase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		expireErr error
		wantError bool
	}{
		{name: "success"},
		{name: "repository error", expireErr: errors.New("expire failure"), wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := &entitlementRepository{purchase: model.Purchase{ID: "purchase", Status: "active", ValidUntil: time.Date(2026, 8, 7, 11, 0, 0, 0, time.UTC)}, expireErr: test.expireErr}
			err := newEntitlementWorkerForTest(repository, &entitlementRemnawave{}).process(context.Background(), model.OutboxJob{Kind: "remna_apply_entitlement", Payload: `{"purchaseId":"purchase"}`})
			if test.wantError && err == nil {
				t.Fatal("process() unexpectedly succeeded")
			}
			if !test.wantError && err != nil {
				t.Fatalf("process(): %v", err)
			}
			if repository.expiredPurchaseID != "purchase" {
				t.Fatalf("expired purchase ID = %q", repository.expiredPurchaseID)
			}
		})
	}
}

func TestWorkerProcessSynchronizesDesiredOrEmptyState(t *testing.T) {
	t.Parallel()

	remoteID := "remote-1"
	desired := model.Purchase{TrafficLimitBytes: 4321, ResetStrategy: "WEEK", SquadUUIDs: []string{"squad"}}
	testError := errors.New("test failure")
	tests := []struct {
		name       string
		repository *entitlementRepository
		remote     *entitlementRemnawave
		wantError  bool
		wantApply  bool
		wantRemove bool
	}{
		{name: "user lookup", repository: &entitlementRepository{userByIDErr: testError}, remote: &entitlementRemnawave{}, wantError: true},
		{name: "no remote identity", repository: &entitlementRepository{user: model.User{}}, remote: &entitlementRemnawave{}},
		{name: "desired lookup", repository: &entitlementRepository{user: model.User{RemnaUserID: &remoteID}, desiredErr: testError}, remote: &entitlementRemnawave{}, wantError: true},
		{name: "remove empty", repository: &entitlementRepository{user: model.User{RemnaUserID: &remoteID}}, remote: &entitlementRemnawave{}, wantRemove: true},
		{name: "remove error", repository: &entitlementRepository{user: model.User{RemnaUserID: &remoteID}}, remote: &entitlementRemnawave{removeErr: testError}, wantRemove: true, wantError: true},
		{name: "apply desired", repository: &entitlementRepository{user: model.User{RemnaUserID: &remoteID}, desired: &desired}, remote: &entitlementRemnawave{}, wantApply: true},
		{name: "apply desired error", repository: &entitlementRepository{user: model.User{RemnaUserID: &remoteID}, desired: &desired}, remote: &entitlementRemnawave{applyErr: testError}, wantApply: true, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := newEntitlementWorkerForTest(test.repository, test.remote).process(context.Background(), model.OutboxJob{Kind: "remna_sync_user", Payload: `{"userId":"user-1"}`})
			if test.wantError && err == nil {
				t.Fatal("process() unexpectedly succeeded")
			}
			if !test.wantError && err != nil {
				t.Fatalf("process(): %v", err)
			}
			if got := test.remote.applyCalls > 0; got != test.wantApply {
				t.Fatalf("ApplyEntitlement called = %t, want %t", got, test.wantApply)
			}
			if got := test.remote.removeCalls > 0; got != test.wantRemove {
				t.Fatalf("RemoveEntitlement called = %t, want %t", got, test.wantRemove)
			}
			if test.wantApply && test.remote.resetCalls != 0 {
				t.Fatal("sync-user apply unexpectedly reset traffic")
			}
		})
	}
}

func TestWorkerProcessRejectsUnknownJob(t *testing.T) {
	t.Parallel()

	err := newEntitlementWorkerForTest(&entitlementRepository{}, &entitlementRemnawave{}).process(context.Background(), model.OutboxJob{Kind: "mystery"})
	if err == nil {
		t.Fatal("process() unexpectedly accepted unknown job")
	}
}
