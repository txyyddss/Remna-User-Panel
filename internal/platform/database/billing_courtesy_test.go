package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

func TestCourtesyCreditPreservesTerminalPaymentState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	member := createTestUser(t, store, 10_007)
	admin := createTestUser(t, store, 10_008)
	now := time.Date(2026, time.August, 11, 10, 0, 0, 0, time.UTC)
	expired := createTestPaymentOrder(t, store, member.ID, "ezpay", 2_500, now)
	if err := store.ExpirePaymentOrder(ctx, expired.ID, "ezpay", now); err != nil {
		t.Fatalf("ExpirePaymentOrder(): %v", err)
	}
	credit, err := store.CourtesyCreditPayment(ctx, admin.ID, expired.ID, "provider support confirmed failure", now)
	if err != nil || credit.TXB.Minor != "2500" || credit.Replayed {
		t.Fatalf("CourtesyCreditPayment() = (%+v, %v)", credit, err)
	}
	replayed, err := store.CourtesyCreditPayment(ctx, admin.ID, expired.ID, "duplicate request", now.Add(time.Minute))
	if err != nil || !replayed.Replayed || replayed.ID != credit.ID {
		t.Fatalf("CourtesyCreditPayment(replay) = (%+v, %v)", replayed, err)
	}
	order, err := store.PaymentOrderByID(ctx, expired.ID)
	if err != nil || order.Status != "expired" || order.PaidAt != nil {
		t.Fatalf("PaymentOrderByID() = (%+v, %v), want unchanged expired order", order, err)
	}
	if _, _, err := store.SettlePayment(ctx, "ezpay", "late-event", "late-payload", expired.ID, "late-trade", "", now.Add(2*time.Minute)); !errors.Is(err, ErrConflict) {
		t.Fatalf("SettlePayment(courtesy credited) = %v, want conflict", err)
	}
	balance, err := store.Balance(ctx, member.ID)
	if err != nil || balance.Minor != "2500" {
		t.Fatalf("Balance() = (%+v, %v), want 2500", balance, err)
	}
	entries, err := store.ListLedger(ctx, member.ID, 10)
	if err != nil || countLedgerKind(entries, "payment_courtesy_credit") != 1 {
		t.Fatalf("ListLedger() = (%+v, %v), want one courtesy credit", entries, err)
	}
	events, err := store.ListAuditEvents(ctx, 10)
	if err != nil || len(events) != 1 || events[0].Action != "payment.courtesy_credit" {
		t.Fatalf("ListAuditEvents() = (%+v, %v), want courtesy audit", events, err)
	}

	failed, err := store.CreatePaymentOrder(ctx, model.PaymentOrder{UserID: member.ID, Provider: "ezpay", Status: "creating", TXBMinor: 300,
		PayableAmount: "3.00", PayableCurrency: "CNY", RateSnapshot: "1", ExpiresAt: now.Add(time.Hour)})
	if err != nil {
		t.Fatalf("CreatePaymentOrder(failed): %v", err)
	}
	if err := store.FailPaymentOrder(ctx, failed.ID); err != nil {
		t.Fatalf("FailPaymentOrder(): %v", err)
	}
	if _, err := store.CourtesyCreditPayment(ctx, admin.ID, failed.ID, "checkout was unavailable", now); err != nil {
		t.Fatalf("CourtesyCreditPayment(failed): %v", err)
	}
}
