package database

import (
	"context"
	"fmt"
	"math"
	"sync"
	"testing"
	"time"
)

func TestAdjustBalanceConcurrentWritesRemainConsistent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 10003)

	const (
		writers = 32
		delta   = int64(25)
	)
	start := make(chan struct{})
	errorsCh := make(chan error, writers)
	var wait sync.WaitGroup
	for index := 0; index < writers; index++ {
		index := index
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			_, err := store.AdjustBalance(ctx, user.ID, delta, fmt.Sprintf("concurrent-%02d", index), "concurrent test", time.Now())
			errorsCh <- err
		}()
	}
	close(start)
	wait.Wait()
	close(errorsCh)
	for err := range errorsCh {
		if err != nil {
			t.Fatalf("concurrent AdjustBalance(): %v", err)
		}
	}

	balance, err := store.Balance(ctx, user.ID)
	if err != nil {
		t.Fatalf("Balance(): %v", err)
	}
	if balance.Minor != "800" {
		t.Fatalf("balance = %s, want 800", balance.Minor)
	}
	entries, err := store.ListLedger(ctx, user.ID, 100)
	if err != nil {
		t.Fatalf("ListLedger(): %v", err)
	}
	if len(entries) != writers {
		t.Fatalf("ledger entry count = %d, want %d", len(entries), writers)
	}

	if _, err := store.AdjustBalance(ctx, user.ID, 999, "concurrent-00", "duplicate reference", time.Now()); err == nil {
		t.Fatal("duplicate ledger reference unexpectedly succeeded")
	}
	balance, err = store.Balance(ctx, user.ID)
	if err != nil {
		t.Fatalf("Balance() after duplicate: %v", err)
	}
	if balance.Minor != "800" {
		t.Fatalf("balance after rolled-back duplicate = %s, want 800", balance.Minor)
	}
}

func TestAdjustBalanceOverflowRollsBack(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 10008)
	now := time.Date(2026, time.August, 8, 9, 0, 0, 0, time.UTC)
	if _, err := store.AdjustBalance(ctx, user.ID, math.MaxInt64, "balance-max", "test", now); err != nil {
		t.Fatalf("AdjustBalance(max): %v", err)
	}
	if _, err := store.AdjustBalance(ctx, user.ID, 1, "balance-overflow", "test", now.Add(time.Second)); err == nil {
		t.Fatal("AdjustBalance(overflow) unexpectedly succeeded")
	}
	balance, err := store.Balance(ctx, user.ID)
	if err != nil {
		t.Fatalf("Balance(): %v", err)
	}
	if balance.Minor != "9223372036854775807" {
		t.Fatalf("balance = %s, want MaxInt64", balance.Minor)
	}
	entries, err := store.ListLedger(ctx, user.ID, 10)
	if err != nil {
		t.Fatalf("ListLedger(): %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("ledger count = %d, want 1", len(entries))
	}
}
