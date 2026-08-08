package database

import (
	"context"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
)

func TestCreatePurchaseSnapshotsAndRenewsAtCurrentTermEnd(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 10001)
	included := saveTestSquad(t, store, "included-squad", 900, true)
	addon := saveTestSquad(t, store, "addon-squad", 300, true)
	combo := saveTestCombo(t, store, "monthly", 1_000, 30, included.ID)
	if _, err := store.AdjustBalance(ctx, user.ID, 5_000, "seed-credit", "test credit", time.Now()); err != nil {
		t.Fatalf("AdjustBalance(): %v", err)
	}

	now := time.Date(2026, time.August, 7, 12, 0, 0, 0, time.UTC)
	first, err := store.CreatePurchase(ctx, PurchaseInput{
		UserID:        user.ID,
		ComboID:       combo.ID,
		AddonSquadIDs: []string{addon.ID, addon.ID},
		IdempotencyKey: "purchase-snapshot-first",
	}, now)
	if err != nil {
		t.Fatalf("CreatePurchase(first): %v", err)
	}
	if first.Status != "activating" {
		t.Fatalf("first status = %q, want activating", first.Status)
	}
	if !first.ValidFrom.Equal(now) {
		t.Fatalf("first validFrom = %s, want %s", first.ValidFrom, now)
	}
	if want := now.AddDate(0, 0, 30); !first.ValidUntil.Equal(want) {
		t.Fatalf("first validUntil = %s, want %s", first.ValidUntil, want)
	}
	if first.PriceTXBMinor != 1_300 {
		t.Fatalf("first price = %d, want 1300", first.PriceTXBMinor)
	}
	if got, want := first.SquadUUIDs, []string{"addon-squad", "included-squad"}; !equalStrings(got, want) {
		t.Fatalf("first squads = %v, want %v", got, want)
	}

	second, err := store.CreatePurchase(ctx, PurchaseInput{
		UserID:        user.ID,
		ComboID:       combo.ID,
		AddonSquadIDs: []string{addon.ID},
		IdempotencyKey: "purchase-snapshot-second",
	}, now.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("CreatePurchase(renewal): %v", err)
	}
	if second.Status != "queued" {
		t.Fatalf("renewal status = %q, want queued", second.Status)
	}
	if !second.ValidFrom.Equal(first.ValidUntil) {
		t.Fatalf("renewal validFrom = %s, want prior validUntil %s", second.ValidFrom, first.ValidUntil)
	}
	if want := first.ValidUntil.AddDate(0, 0, 30); !second.ValidUntil.Equal(want) {
		t.Fatalf("renewal validUntil = %s, want %s", second.ValidUntil, want)
	}

	balance, err := store.Balance(ctx, user.ID)
	if err != nil {
		t.Fatalf("Balance(): %v", err)
	}
	if balance.Minor != "2400" {
		t.Fatalf("balance = %s, want 2400", balance.Minor)
	}
	entries, err := store.ListLedger(ctx, user.ID, 20)
	if err != nil {
		t.Fatalf("ListLedger(): %v", err)
	}
	if got := countLedgerKind(entries, "purchase_debit"); got != 2 {
		t.Fatalf("purchase debit count = %d, want 2", got)
	}
	var outboxCount int
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM outbox_jobs WHERE kind='remna_apply_entitlement'`).Scan(&outboxCount); err != nil {
		t.Fatalf("count entitlement outbox: %v", err)
	}
	if outboxCount != 1 {
		t.Fatalf("entitlement outbox count = %d, want 1 before renewal becomes due", outboxCount)
	}
}

func TestCreatePurchaseInsufficientBalanceIsAtomic(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 10002)
	combo := saveTestCombo(t, store, "too-expensive", 1_000, 30)
	if _, err := store.AdjustBalance(ctx, user.ID, 999, "almost-enough", "test credit", time.Now()); err != nil {
		t.Fatalf("AdjustBalance(): %v", err)
	}

	_, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, IdempotencyKey: "purchase-insufficient"}, time.Now())
	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("CreatePurchase() error = %v, want ErrInsufficientBalance", err)
	}

	balance, err := store.Balance(ctx, user.ID)
	if err != nil {
		t.Fatalf("Balance(): %v", err)
	}
	if balance.Minor != "999" {
		t.Fatalf("balance = %s, want 999", balance.Minor)
	}
	assertRowCount(t, store, "purchases", 0)
	assertRowCount(t, store, "outbox_jobs", 0)
	var purchaseDebits int
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM ledger_entries WHERE kind='purchase_debit'`).Scan(&purchaseDebits); err != nil {
		t.Fatalf("count purchase debits: %v", err)
	}
	if purchaseDebits != 0 {
		t.Fatalf("purchase debit count = %d, want 0", purchaseDebits)
	}
}

func TestCreatePurchaseIdempotencyReplaysAndRejectsFingerprintReuse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 10004)
	addon := saveTestSquad(t, store, "idempotent-addon", 300, true)
	combo := saveTestCombo(t, store, "idempotent-combo", 1_000, 30)
	otherCombo := saveTestCombo(t, store, "different-combo", 500, 30)
	if _, err := store.AdjustBalance(ctx, user.ID, 5_000, "idempotency-seed", "seed", time.Now()); err != nil {
		t.Fatal(err)
	}
	input := PurchaseInput{UserID: user.ID, ComboID: combo.ID, AddonSquadIDs: []string{" " + addon.ID + " ", addon.ID}, IdempotencyKey: " purchase-attempt "}

	const callers = 12
	start := make(chan struct{})
	results := make(chan model.Purchase, callers)
	errorsCh := make(chan error, callers)
	for range callers {
		go func() {
			<-start
			purchase, err := store.CreatePurchase(ctx, input, time.Now())
			results <- purchase
			errorsCh <- err
		}()
	}
	close(start)
	var purchaseID string
	for range callers {
		purchase := <-results
		if err := <-errorsCh; err != nil {
			t.Fatalf("CreatePurchase(retry): %v", err)
		}
		if purchaseID == "" {
			purchaseID = purchase.ID
		} else if purchase.ID != purchaseID {
			t.Fatalf("replay purchase ID = %q, want %q", purchase.ID, purchaseID)
		}
	}
	conflict := input
	conflict.ComboID = otherCombo.ID
	if _, err := store.CreatePurchase(ctx, conflict, time.Now()); !errors.Is(err, ErrConflict) {
		t.Fatalf("CreatePurchase(fingerprint conflict) error = %v, want ErrConflict", err)
	}
	if _, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID}, time.Now()); err == nil {
		t.Fatal("CreatePurchase() accepted a missing idempotency key")
	}
	var purchases, debits, jobs int
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM purchases WHERE user_id=?`, user.ID).Scan(&purchases); err != nil {
		t.Fatal(err)
	}
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM ledger_entries WHERE user_id=? AND kind='purchase_debit'`, user.ID).Scan(&debits); err != nil {
		t.Fatal(err)
	}
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM outbox_jobs WHERE kind='remna_apply_entitlement' AND aggregate_id=?`, purchaseID).Scan(&jobs); err != nil {
		t.Fatal(err)
	}
	if purchases != 1 || debits != 1 || jobs != 1 {
		t.Fatalf("idempotent effects purchases=%d debits=%d jobs=%d, want 1/1/1", purchases, debits, jobs)
	}
	balance, err := store.Balance(ctx, user.ID)
	if err != nil || balance.Minor != "3700" {
		t.Fatalf("idempotent balance = (%+v, %v), want 3700", balance, err)
	}
}

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

func TestSettlePaymentCreditsExactlyOnce(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 10004)
	now := time.Date(2026, time.August, 7, 13, 0, 0, 0, time.UTC)
	order := createTestPaymentOrder(t, store, user.ID, "ezpay", 2_500, now)

	paid, applied, err := store.SettlePayment(ctx, "ezpay", "event-1", "payload-a", order.ID, "trade-1", "charge-1", now)
	if err != nil {
		t.Fatalf("SettlePayment(first): %v", err)
	}
	if !applied || paid.Status != "paid" {
		t.Fatalf("first settlement = (status %q, applied %t), want (paid, true)", paid.Status, applied)
	}

	replayed, applied, err := store.SettlePayment(ctx, "ezpay", "event-1", "payload-a", order.ID, "trade-1", "charge-1", now.Add(time.Minute))
	if err != nil {
		t.Fatalf("SettlePayment(replay): %v", err)
	}
	if applied || replayed.Status != "paid" {
		t.Fatalf("replay = (status %q, applied %t), want (paid, false)", replayed.Status, applied)
	}

	if replayed, applied, err = store.SettlePayment(ctx, "ezpay", "event-1", "different-payload", order.ID, "trade-1", "charge-1", now.Add(2*time.Minute)); err != nil || applied || replayed.Status != "paid" {
		t.Fatalf("metadata-varied replay = (status %q, applied %t, error %v), want (paid, false, nil)", replayed.Status, applied, err)
	}
	otherOrder := createTestPaymentOrder(t, store, user.ID, "ezpay", 2_500, now.Add(3*time.Minute))
	if _, _, err := store.SettlePayment(ctx, "ezpay", "event-1", "payload-a", otherOrder.ID, "trade-1", "charge-1", now.Add(4*time.Minute)); !errors.Is(err, ErrConflict) {
		t.Fatalf("cross-order replay error = %v, want ErrConflict", err)
	}
	balance, err := store.Balance(ctx, user.ID)
	if err != nil {
		t.Fatalf("Balance(): %v", err)
	}
	if balance.Minor != "2500" {
		t.Fatalf("balance = %s, want 2500", balance.Minor)
	}
	entries, err := store.ListLedger(ctx, user.ID, 20)
	if err != nil {
		t.Fatalf("ListLedger(): %v", err)
	}
	if got := countLedgerKind(entries, "payment_credit"); got != 1 {
		t.Fatalf("payment credit count = %d, want 1", got)
	}
	assertRowCount(t, store, "webhook_events", 1)
}

func TestSettlePaymentOverflowRollsBackWebhookAndCredit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 10009)
	now := time.Date(2026, time.August, 8, 10, 0, 0, 0, time.UTC)
	order := createTestPaymentOrder(t, store, user.ID, "ezpay", 1, now)
	if _, err := store.AdjustBalance(ctx, user.ID, math.MaxInt64, "payment-overflow-max", "test", now); err != nil {
		t.Fatalf("AdjustBalance(max): %v", err)
	}
	if _, _, err := store.SettlePayment(ctx, "ezpay", "overflow-event", "overflow-payload", order.ID, "overflow-trade", "", now.Add(time.Second)); err == nil {
		t.Fatal("SettlePayment(overflow) unexpectedly succeeded")
	}
	current, err := store.PaymentOrderByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("PaymentOrderByID(): %v", err)
	}
	if current.Status != "pending" {
		t.Fatalf("order status = %q, want pending", current.Status)
	}
	assertRowCount(t, store, "webhook_events", 0)
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
	if got := countLedgerKind(entries, "payment_credit"); got != 0 {
		t.Fatalf("payment credit count = %d, want 0", got)
	}
}

func TestRefundPaymentUnderflowRollsBack(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 10010)
	now := time.Date(2026, time.August, 8, 11, 0, 0, 0, time.UTC)
	order := createTestPaymentOrder(t, store, user.ID, "ezpay", 1, now)
	if _, applied, err := store.SettlePayment(ctx, "ezpay", "underflow-paid", "paid", order.ID, "underflow-trade", "", now); err != nil || !applied {
		t.Fatalf("SettlePayment() = (applied %t, err %v)", applied, err)
	}
	if _, err := store.AdjustBalance(ctx, user.ID, math.MinInt64, "refund-underflow-a", "test", now.Add(time.Second)); err != nil {
		t.Fatalf("AdjustBalance(MinInt64): %v", err)
	}
	if _, err := store.AdjustBalance(ctx, user.ID, -1, "refund-underflow-b", "test", now.Add(2*time.Second)); err != nil {
		t.Fatalf("AdjustBalance(-1): %v", err)
	}
	actorID := user.ID
	if _, err := store.RefundPayment(ctx, &actorID, order.ID, "underflow", now.Add(3*time.Second)); err == nil {
		t.Fatal("RefundPayment(underflow) unexpectedly succeeded")
	}
	current, err := store.PaymentOrderByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("PaymentOrderByID(): %v", err)
	}
	if current.Status != "paid" {
		t.Fatalf("order status = %q, want paid", current.Status)
	}
	assertRowCount(t, store, "refunds", 0)
	balance, err := store.Balance(ctx, user.ID)
	if err != nil {
		t.Fatalf("Balance(): %v", err)
	}
	if balance.Minor != "-9223372036854775808" {
		t.Fatalf("balance = %s, want MinInt64", balance.Minor)
	}
	entries, err := store.ListLedger(ctx, user.ID, 10)
	if err != nil {
		t.Fatalf("ListLedger(): %v", err)
	}
	if got := countLedgerKind(entries, "payment_reversal"); got != 0 {
		t.Fatalf("payment reversal count = %d, want 0", got)
	}
}

func TestRefundCancelsQueuedBeforeActiveAndIsIdempotent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := newTestStore(t)
	user := createTestUser(t, store, 10005)
	combo := saveTestCombo(t, store, "refund-plan", 6_000, 30)
	base := time.Date(2026, time.August, 7, 14, 0, 0, 0, time.UTC)
	order := createTestPaymentOrder(t, store, user.ID, "bepusdt", 10_000, base)
	if _, applied, err := store.SettlePayment(ctx, "bepusdt", "paid-event", "paid-hash", order.ID, "trade-refund", "", base); err != nil || !applied {
		t.Fatalf("SettlePayment() = (applied %t, err %v), want (true, nil)", applied, err)
	}
	if _, err := store.AdjustBalance(ctx, user.ID, 5_000, "preexisting-credit", "test credit", base); err != nil {
		t.Fatalf("AdjustBalance(): %v", err)
	}
	active, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, IdempotencyKey: "purchase-cancel-active"}, base.Add(time.Minute))
	if err != nil {
		t.Fatalf("CreatePurchase(active): %v", err)
	}
	queued, err := store.CreatePurchase(ctx, PurchaseInput{UserID: user.ID, ComboID: combo.ID, IdempotencyKey: "purchase-cancel-queued"}, base.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("CreatePurchase(queued): %v", err)
	}
	if active.Status != "activating" || queued.Status != "queued" {
		t.Fatalf("purchase statuses = %q, %q, want activating, queued", active.Status, queued.Status)
	}

	actorID := user.ID
	refunded, err := store.RefundPayment(ctx, &actorID, order.ID, "provider reversal", base.Add(3*time.Minute))
	if err != nil {
		t.Fatalf("RefundPayment(): %v", err)
	}
	if refunded.Status != "refunded" || refunded.RefundedAt == nil {
		t.Fatalf("refunded order = (status %q, refundedAt %v)", refunded.Status, refunded.RefundedAt)
	}

	purchases, err := store.ListPurchases(ctx, user.ID)
	if err != nil {
		t.Fatalf("ListPurchases(): %v", err)
	}
	if len(purchases) != 2 {
		t.Fatalf("purchase count = %d, want 2", len(purchases))
	}
	for _, purchase := range purchases {
		if purchase.Status != "cancelled" {
			t.Fatalf("purchase %s status = %q, want cancelled", purchase.ID, purchase.Status)
		}
	}
	balance, err := store.Balance(ctx, user.ID)
	if err != nil {
		t.Fatalf("Balance(): %v", err)
	}
	if balance.Minor != "5000" {
		t.Fatalf("balance = %s, want original non-payment credit 5000", balance.Minor)
	}
	entries, err := store.ListLedger(ctx, user.ID, 20)
	if err != nil {
		t.Fatalf("ListLedger(): %v", err)
	}
	wantKinds := map[string]int{
		"admin_adjustment":      1,
		"payment_credit":        1,
		"payment_reversal":      1,
		"purchase_cancellation": 2,
		"purchase_debit":        2,
	}
	for kind, want := range wantKinds {
		if got := countLedgerKind(entries, kind); got != want {
			t.Fatalf("%s ledger count = %d, want %d", kind, got, want)
		}
	}
	var syncJobs int
	if err := store.DB().QueryRowContext(ctx, `SELECT COUNT(*) FROM outbox_jobs WHERE kind='remna_sync_user' AND aggregate_id=?`, user.ID).Scan(&syncJobs); err != nil {
		t.Fatalf("count refund sync jobs: %v", err)
	}
	if syncJobs != 1 {
		t.Fatalf("refund sync job count = %d, want 1 for the active entitlement only", syncJobs)
	}
	assertRowCount(t, store, "refunds", 1)

	if _, err := store.RefundPayment(ctx, &actorID, order.ID, "duplicate request", base.Add(4*time.Minute)); err != nil {
		t.Fatalf("RefundPayment(replay): %v", err)
	}
	assertRowCount(t, store, "refunds", 1)
	entriesAfterReplay, err := store.ListLedger(ctx, user.ID, 20)
	if err != nil {
		t.Fatalf("ListLedger() after replay: %v", err)
	}
	if len(entriesAfterReplay) != len(entries) {
		t.Fatalf("ledger count after refund replay = %d, want %d", len(entriesAfterReplay), len(entries))
	}
}

func newTestStore(t *testing.T) *Store {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "tx-carpool-test.db"))
	if err != nil {
		t.Fatalf("Open(): %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})
	return NewStore(db)
}

func createTestUser(t *testing.T, store *Store, telegramID int64) model.User {
	t.Helper()

	user, created, err := store.UpsertTelegramUser(context.Background(), model.TelegramProfile{
		ID:        telegramID,
		FirstName: "Test",
		Username:  fmt.Sprintf("telegram%d", telegramID),
	}, false)
	if err != nil {
		t.Fatalf("UpsertTelegramUser(): %v", err)
	}
	if !created {
		t.Fatal("test user was not created")
	}
	return user
}

func saveTestSquad(t *testing.T, store *Store, uuid string, price int64, visible bool) model.SquadProduct {
	t.Helper()

	imported, err := store.ImportSquad(context.Background(), uuid, uuid)
	if err != nil {
		t.Fatalf("ImportSquad(%q): %v", uuid, err)
	}
	product, err := store.SaveSquadProduct(context.Background(), SquadProductInput{
		ID:              imported.ID,
		RemnaSquadUUID:  uuid,
		Name:            uuid,
		Description:     "test squad",
		PriceTXBMinor:   price,
		Visible:         visible,
		UpstreamPresent: true,
	})
	if err != nil {
		t.Fatalf("SaveSquadProduct(%q): %v", uuid, err)
	}
	return product
}

func saveTestCombo(t *testing.T, store *Store, name string, price int64, validityDays int, squadIDs ...string) model.Combo {
	t.Helper()

	combo, err := store.SaveCombo(context.Background(), ComboInput{
		Name:              name,
		Description:       "test combo",
		PriceTXBMinor:     price,
		ValidityDays:      validityDays,
		TrafficLimitBytes: 100 * 1024 * 1024,
		ResetStrategy:     "MONTH",
		Active:            true,
		SquadProductIDs:   squadIDs,
	})
	if err != nil {
		t.Fatalf("SaveCombo(%q): %v", name, err)
	}
	return combo
}

func createTestPaymentOrder(t *testing.T, store *Store, userID, provider string, txbMinor int64, now time.Time) model.PaymentOrder {
	t.Helper()

	currency := "CNY"
	switch provider {
	case "bepusdt":
		currency = "USD"
	case "stars":
		currency = "XTR"
	}
	order, err := store.CreatePaymentOrder(context.Background(), model.PaymentOrder{
		UserID:          userID,
		Provider:        provider,
		Status:          "pending",
		TXBMinor:        txbMinor,
		PayableAmount:   "10.00",
		PayableCurrency: currency,
		RateSnapshot:    "1",
		ProviderPayload: "{}",
		ExpiresAt:       now.Add(30 * time.Minute),
	})
	if err != nil {
		t.Fatalf("CreatePaymentOrder(): %v", err)
	}
	return order
}

func countLedgerKind(entries []model.LedgerEntry, kind string) int {
	count := 0
	for _, entry := range entries {
		if entry.Kind == kind {
			count++
		}
	}
	return count
}

func assertRowCount(t *testing.T, store *Store, table string, want int) {
	t.Helper()

	var query string
	switch table {
	case "outbox_jobs":
		query = `SELECT COUNT(*) FROM outbox_jobs`
	case "purchases":
		query = `SELECT COUNT(*) FROM purchases`
	case "refunds":
		query = `SELECT COUNT(*) FROM refunds`
	case "webhook_events":
		query = `SELECT COUNT(*) FROM webhook_events`
	default:
		t.Fatalf("assertRowCount called with unsupported table %q", table)
	}
	var got int
	if err := store.DB().QueryRowContext(context.Background(), query).Scan(&got); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	if got != want {
		t.Fatalf("%s row count = %d, want %d", table, got, want)
	}
}

func equalStrings(left, right []string) bool {
	leftCopy := append([]string(nil), left...)
	rightCopy := append([]string(nil), right...)
	sort.Strings(leftCopy)
	sort.Strings(rightCopy)
	if len(leftCopy) != len(rightCopy) {
		return false
	}
	for index := range leftCopy {
		if leftCopy[index] != rightCopy[index] {
			return false
		}
	}
	return true
}
