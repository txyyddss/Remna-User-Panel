package database

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPaymentCallbackTombstoneSurvivesPaymentCompaction(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 29101)
	settledAt := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	order := createTestPaymentOrder(t, store, user.ID, "ezpay", 500, settledAt)
	if _, applied, err := store.SettlePayment(ctx, "ezpay", "callback-1", "payload-1",
		order.ID, "trade-1", "charge-1", settledAt); err != nil || !applied {
		t.Fatalf("SettlePayment() = (applied %t, error %v)", applied, err)
	}

	cleanupAt := settledAt.Add(9 * 24 * time.Hour)
	if _, err := store.CompactAndPrune(ctx, cleanupAt.Add(-7*24*time.Hour), cleanupAt.Add(-24*time.Hour), cleanupAt); err != nil {
		t.Fatalf("CompactAndPrune(): %v", err)
	}
	if _, err := store.PaymentOrderByID(ctx, order.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("PaymentOrderByID() error = %v, want ErrNotFound", err)
	}
	replayed, err := store.PaymentCallbackReplay(ctx, "ezpay", "callback-1", order.ID)
	if err != nil || !replayed {
		t.Fatalf("PaymentCallbackReplay() = (%t, %v), want (true, nil)", replayed, err)
	}
	if _, err := store.PaymentCallbackReplay(ctx, "ezpay", "callback-1", "different-order"); !errors.Is(err, ErrConflict) {
		t.Fatalf("cross-order replay error = %v, want ErrConflict", err)
	}
	assertRowCount(t, store, "payment_callback_tombstones", 1)
}
