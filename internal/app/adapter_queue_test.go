package app

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/accounts"
	domain "github.com/txyyddss/Remna-User-Panel/internal/emby"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/upstreamqueue"
)

func TestRemnawaveAdapterMethodsEnterQueueBeforeClientCreation(t *testing.T) {
	queue, err := upstreamqueue.New(upstreamqueue.Config{Name: "remna-test", Capacity: 1})
	if err != nil {
		t.Fatal(err)
	}
	var factoryCalls atomic.Int32
	adapter := remnaAdapter{
		queue: queue,
		clientFactory: func(context.Context) (remnaClient, error) {
			factoryCalls.Add(1)
			return nil, errors.New("client factory must not run")
		},
	}
	createInput := accounts.RemoteCreateUser{
		Username: "tester", Status: "ACTIVE", ExpireAt: time.Now().Add(time.Hour),
		TelegramID: 42, TrafficLimitStrategy: "NO_RESET",
	}
	tests := []struct {
		name string
		call func() error
	}{
		{name: "find username", call: func() error { _, _, err := adapter.FindUserByUsername(context.Background(), "tester"); return err }},
		{name: "find Telegram id", call: func() error { _, _, err := adapter.FindUserByTelegramID(context.Background(), 42); return err }},
		{name: "find remote id", call: func() error { _, _, err := adapter.FindUserByID(context.Background(), "1"); return err }},
		{name: "create user", call: func() error { _, err := adapter.CreateUser(context.Background(), createInput); return err }},
		{name: "dashboard", call: func() error { _, err := adapter.Dashboard(context.Background(), "1"); return err }},
		{name: "revoke subscription", call: func() error { _, err := adapter.RevokeSubscription(context.Background(), "1"); return err }},
		{name: "apply entitlement", call: func() error {
			return adapter.ApplyEntitlement(context.Background(), "1", 10, "NO_RESET", []string{"squad"})
		}},
		{name: "reset traffic", call: func() error { return adapter.ResetTraffic(context.Background(), "1") }},
		{name: "remove entitlement", call: func() error { return adapter.RemoveEntitlement(context.Background(), "1") }},
		{name: "quiesce rollover", call: func() error { return adapter.QuiesceForRollover(context.Background(), "1") }},
		{name: "read rollover traffic", call: func() error { _, _, err := adapter.TrafficForRollover(context.Background(), "1"); return err }},
		{name: "list admin squads", call: func() error { _, err := adapter.ListInternalSquads(context.Background()); return err }},
		{name: "list catalog squads", call: func() error { _, err := adapter.ListCatalogSquads(context.Background()); return err }},
		{name: "list nodes", call: func() error { _, err := adapter.ListNodes(context.Background()); return err }},
		{name: "accessible nodes", call: func() error { _, err := adapter.AccessibleNodeUUIDs(context.Background(), "squad"); return err }},
		{name: "update squad inbounds", call: func() error {
			return adapter.UpdateInternalSquadInbounds(context.Background(), "squad", []string{"inbound"})
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			factoryCalls.Store(0)
			if err := test.call(); !errors.Is(err, upstreamqueue.ErrNotRunning) {
				t.Fatalf("adapter error = %v, want ErrNotRunning", err)
			}
			if calls := factoryCalls.Load(); calls != 0 {
				t.Fatalf("client factory calls = %d, want 0", calls)
			}
		})
	}
}

func TestEmbyAdapterMethodsEnterQueueBeforeClientCreation(t *testing.T) {
	queue, err := upstreamqueue.New(upstreamqueue.Config{Name: "emby-test", Capacity: 1})
	if err != nil {
		t.Fatal(err)
	}
	var factoryCalls atomic.Int32
	adapter := embyAdapter{
		queue: queue,
		clientFactory: func(context.Context) (embyClient, error) {
			factoryCalls.Add(1)
			return nil, errors.New("client factory must not run")
		},
	}
	tests := []struct {
		name string
		call func() error
	}{
		{name: "find user", call: func() error { _, _, err := adapter.FindUserByName(context.Background(), "tester"); return err }},
		{name: "create user", call: func() error { _, err := adapter.CreateUser(context.Background(), "tester"); return err }},
		{name: "get user", call: func() error { _, err := adapter.GetUser(context.Background(), "user"); return err }},
		{name: "set password", call: func() error { return adapter.SetPassword(context.Background(), "user", []byte("old"), []byte("new")) }},
		{name: "update policy", call: func() error { return adapter.UpdatePolicy(context.Background(), "user", domain.Policy{}) }},
		{name: "list folders", call: func() error { _, err := adapter.ListSelectableFolders(context.Background()); return err }},
		{name: "list ratings", call: func() error { _, err := adapter.ListParentalRatings(context.Background()); return err }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			factoryCalls.Store(0)
			if err := test.call(); !errors.Is(err, upstreamqueue.ErrNotRunning) {
				t.Fatalf("adapter error = %v, want ErrNotRunning", err)
			}
			if calls := factoryCalls.Load(); calls != 0 {
				t.Fatalf("client factory calls = %d, want 0", calls)
			}
		})
	}
}
