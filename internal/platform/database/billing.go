package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
)

// PurchaseInput contains the server-selected catalog IDs for a checkout.
type PurchaseInput struct {
	UserID        string
	ComboID       string
	AddonSquadIDs []string
}

// CreatePurchase atomically snapshots the catalog, debits TXB, and queues Remnawave synchronization.
func (s *Store) CreatePurchase(ctx context.Context, input PurchaseInput, now time.Time) (model.Purchase, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	now = now.UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Purchase{}, fmt.Errorf("begin purchase: %w", err)
	}
	defer tx.Rollback()

	combo, err := comboByIDTx(ctx, tx, input.ComboID, true)
	if err != nil {
		return model.Purchase{}, err
	}
	included, err := comboSquadsTx(ctx, tx, combo.ID)
	if err != nil {
		return model.Purchase{}, err
	}
	combo.IncludedSquads = included
	includedUUIDs := make(map[string]struct{}, len(included))
	squadUUIDs := make([]string, 0, len(included)+len(input.AddonSquadIDs))
	for _, product := range included {
		includedUUIDs[product.RemnaSquadUUID] = struct{}{}
		squadUUIDs = append(squadUUIDs, product.RemnaSquadUUID)
	}
	addonPrice := int64(0)
	addonRows := make([]model.SquadProduct, 0, len(input.AddonSquadIDs))
	for _, productID := range uniqueSorted(input.AddonSquadIDs) {
		product, err := squadByIDTx(ctx, tx, productID, true)
		if err != nil {
			return model.Purchase{}, err
		}
		if _, alreadyIncluded := includedUUIDs[product.RemnaSquadUUID]; alreadyIncluded {
			continue
		}
		includedUUIDs[product.RemnaSquadUUID] = struct{}{}
		squadUUIDs = append(squadUUIDs, product.RemnaSquadUUID)
		addonRows = append(addonRows, product)
		addonPrice += product.PriceTXBMinor
	}
	squadUUIDs = uniqueSorted(squadUUIDs)
	totalPrice := combo.PriceTXBMinor + addonPrice

	validFrom := now
	var latestEnd string
	err = tx.QueryRowContext(ctx, `SELECT valid_until FROM purchases WHERE user_id=? AND status IN ('activating','active','queued') ORDER BY valid_until DESC LIMIT 1`, input.UserID).Scan(&latestEnd)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return model.Purchase{}, fmt.Errorf("find current entitlement: %w", err)
	}
	if err == nil {
		parsed, parseErr := parseStamp(latestEnd)
		if parseErr != nil {
			return model.Purchase{}, fmt.Errorf("parse current entitlement: %w", parseErr)
		}
		if parsed.After(validFrom) {
			validFrom = parsed
		}
	}
	validUntil := validFrom.AddDate(0, 0, combo.ValidityDays)
	status := "queued"
	if !validFrom.After(now) {
		status = "activating"
	}

	newBalance, err := debitBalanceTx(ctx, tx, input.UserID, totalPrice, now)
	if err != nil {
		return model.Purchase{}, err
	}
	purchaseID, err := ids.New()
	if err != nil {
		return model.Purchase{}, err
	}
	snapshotBytes, err := json.Marshal(struct {
		Combo  model.Combo          `json:"combo"`
		Addons []model.SquadProduct `json:"addons"`
	}{Combo: combo, Addons: addonRows})
	if err != nil {
		return model.Purchase{}, fmt.Errorf("encode catalog snapshot: %w", err)
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO purchases(id,user_id,combo_id,price_txb_minor,valid_from,valid_until,status,traffic_limit_bytes,reset_strategy,catalog_snapshot,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, purchaseID, input.UserID, combo.ID, totalPrice, stamp(validFrom), stamp(validUntil), status,
		combo.TrafficLimitBytes, combo.ResetStrategy, string(snapshotBytes), stamp(now), stamp(now))
	if err != nil {
		return model.Purchase{}, fmt.Errorf("insert purchase: %w", err)
	}
	for _, product := range included {
		if _, err := tx.ExecContext(ctx, `INSERT INTO purchase_squads(purchase_id,remna_squad_uuid,kind,price_txb_minor) VALUES(?,?,?,?)`, purchaseID, product.RemnaSquadUUID, "included", 0); err != nil {
			return model.Purchase{}, fmt.Errorf("snapshot included squad: %w", err)
		}
	}
	for _, product := range addonRows {
		if _, err := tx.ExecContext(ctx, `INSERT INTO purchase_squads(purchase_id,remna_squad_uuid,kind,price_txb_minor) VALUES(?,?,?,?)`, purchaseID, product.RemnaSquadUUID, "addon", product.PriceTXBMinor); err != nil {
			return model.Purchase{}, fmt.Errorf("snapshot add-on squad: %w", err)
		}
	}
	if _, err := insertLedgerTx(ctx, tx, input.UserID, -totalPrice, newBalance, "purchase_debit", purchaseID, combo.Name, now); err != nil {
		return model.Purchase{}, err
	}
	if status == "activating" {
		if err := insertOutboxTx(ctx, tx, "remna_apply_entitlement", purchaseID, `{"purchaseId":"`+purchaseID+`"}`, now, now); err != nil {
			return model.Purchase{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return model.Purchase{}, fmt.Errorf("commit purchase: %w", err)
	}
	return s.PurchaseByID(ctx, purchaseID)
}

func comboByIDTx(ctx context.Context, tx *sql.Tx, id string, activeOnly bool) (model.Combo, error) {
	query := comboSelect + ` WHERE id=?`
	if activeOnly {
		query += ` AND active=1`
	}
	return scanCombo(tx.QueryRowContext(ctx, query, id))
}

func comboSquadsTx(ctx context.Context, tx *sql.Tx, comboID string) ([]model.SquadProduct, error) {
	rows, err := tx.QueryContext(ctx, squadSelect+` JOIN combo_squads cs ON cs.squad_product_id=squad_products.id WHERE cs.combo_id=? ORDER BY name,id`, comboID)
	if err != nil {
		return nil, fmt.Errorf("list combo squads: %w", err)
	}
	defer rows.Close()
	products := make([]model.SquadProduct, 0)
	for rows.Next() {
		product, err := scanSquad(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, rows.Err()
}

func squadByIDTx(ctx context.Context, tx *sql.Tx, id string, requireVisible bool) (model.SquadProduct, error) {
	query := squadSelect + ` WHERE id=?`
	if requireVisible {
		query += ` AND visible=1 AND upstream_present=1`
	}
	return scanSquad(tx.QueryRowContext(ctx, query, id))
}

func debitBalanceTx(ctx context.Context, tx *sql.Tx, userID string, amount int64, now time.Time) (int64, error) {
	if amount < 0 {
		return 0, errors.New("negative debit")
	}
	var balance int64
	err := tx.QueryRowContext(ctx, `UPDATE balances SET txb_minor=txb_minor-?,updated_at=? WHERE user_id=? AND txb_minor>=? RETURNING txb_minor`, amount, stamp(now), userID, amount).Scan(&balance)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrInsufficientBalance
	}
	if err != nil {
		return 0, fmt.Errorf("debit balance: %w", err)
	}
	return balance, nil
}

// Balance returns the current TXB amount.
func (s *Store) Balance(ctx context.Context, userID string) (model.Money, error) {
	var balance int64
	if err := s.db.QueryRowContext(ctx, `SELECT txb_minor FROM balances WHERE user_id=?`, userID).Scan(&balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.TXBMoney(0), nil
		}
		return model.Money{}, fmt.Errorf("read balance: %w", err)
	}
	return model.TXBMoney(balance), nil
}

// AdjustBalance appends an immutable administrator-authored ledger entry.
func (s *Store) AdjustBalance(ctx context.Context, userID string, delta int64, referenceID, note string, now time.Time) (model.LedgerEntry, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.LedgerEntry{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT INTO balances(user_id,txb_minor,updated_at) VALUES(?,?,?) ON CONFLICT(user_id) DO UPDATE SET txb_minor=balances.txb_minor+excluded.txb_minor,updated_at=excluded.updated_at`, userID, delta, stamp(now)); err != nil {
		return model.LedgerEntry{}, fmt.Errorf("adjust balance: %w", err)
	}
	var balance int64
	if err := tx.QueryRowContext(ctx, `SELECT txb_minor FROM balances WHERE user_id=?`, userID).Scan(&balance); err != nil {
		return model.LedgerEntry{}, err
	}
	entryID, err := insertLedgerTx(ctx, tx, userID, delta, balance, "admin_adjustment", referenceID, note, now)
	if err != nil {
		return model.LedgerEntry{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.LedgerEntry{}, err
	}
	return s.LedgerEntryByID(ctx, entryID)
}

func insertLedgerTx(ctx context.Context, tx *sql.Tx, userID string, delta, balance int64, kind, referenceID, note string, now time.Time) (string, error) {
	id, err := ids.New()
	if err != nil {
		return "", err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO ledger_entries(id,user_id,delta_txb_minor,balance_after,kind,reference_id,note,created_at) VALUES(?,?,?,?,?,?,?,?)`, id, userID, delta, balance, kind, referenceID, note, stamp(now))
	if err != nil {
		return "", fmt.Errorf("append ledger: %w", err)
	}
	return id, nil
}

// LedgerEntryByID returns one immutable ledger row.
func (s *Store) LedgerEntryByID(ctx context.Context, id string) (model.LedgerEntry, error) {
	return scanLedger(s.db.QueryRowContext(ctx, ledgerSelect+` WHERE id=?`, id))
}

// ListLedger returns newest entries first.
func (s *Store) ListLedger(ctx context.Context, userID string, limit int) ([]model.LedgerEntry, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, ledgerSelect+` WHERE user_id=? ORDER BY created_at DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list ledger: %w", err)
	}
	defer rows.Close()
	entries := make([]model.LedgerEntry, 0)
	for rows.Next() {
		entry, err := scanLedger(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

const ledgerSelect = `SELECT id,delta_txb_minor,balance_after,kind,reference_id,note,created_at FROM ledger_entries`

func scanLedger(row rowScanner) (model.LedgerEntry, error) {
	var entry model.LedgerEntry
	var created string
	if err := row.Scan(&entry.ID, &entry.DeltaTXBMinor, &entry.BalanceAfterRaw, &entry.Kind, &entry.ReferenceID, &entry.Note, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.LedgerEntry{}, ErrNotFound
		}
		return model.LedgerEntry{}, err
	}
	entry.Delta = model.TXBMoney(entry.DeltaTXBMinor)
	entry.BalanceAfter = model.TXBMoney(entry.BalanceAfterRaw)
	var err error
	entry.CreatedAt, err = parseStamp(created)
	return entry, err
}

// PurchaseByID returns a catalog-snapshotted purchase.
func (s *Store) PurchaseByID(ctx context.Context, id string) (model.Purchase, error) {
	purchase, err := scanPurchase(s.db.QueryRowContext(ctx, purchaseSelect+` WHERE purchases.id=?`, id))
	if err != nil {
		return model.Purchase{}, err
	}
	purchase.SquadUUIDs, err = s.purchaseSquads(ctx, purchase.ID)
	return purchase, err
}

// ListPurchases returns a user's purchase history.
func (s *Store) ListPurchases(ctx context.Context, userID string) ([]model.Purchase, error) {
	rows, err := s.db.QueryContext(ctx, purchaseSelect+` WHERE purchases.user_id=? ORDER BY purchases.valid_from DESC,purchases.created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list purchases: %w", err)
	}
	defer rows.Close()
	purchases := make([]model.Purchase, 0)
	for rows.Next() {
		purchase, err := scanPurchase(rows)
		if err != nil {
			return nil, err
		}
		purchases = append(purchases, purchase)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for index := range purchases {
		purchases[index].SquadUUIDs, err = s.purchaseSquads(ctx, purchases[index].ID)
		if err != nil {
			return nil, err
		}
	}
	return purchases, nil
}

// ListAllPurchases returns recent entitlement records for the administrative view.
func (s *Store) ListAllPurchases(ctx context.Context, limit int) ([]model.Purchase, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, purchaseSelect+` ORDER BY purchases.created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	purchases := make([]model.Purchase, 0)
	for rows.Next() {
		purchase, err := scanPurchase(rows)
		if err != nil {
			return nil, err
		}
		purchases = append(purchases, purchase)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for index := range purchases {
		purchases[index].SquadUUIDs, err = s.purchaseSquads(ctx, purchases[index].ID)
		if err != nil {
			return nil, err
		}
	}
	return purchases, nil
}

// CancelPurchase credits its snapshotted price and schedules entitlement replacement when needed.
func (s *Store) CancelPurchase(ctx context.Context, purchaseID, reason string, now time.Time) (model.Purchase, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Purchase{}, err
	}
	defer tx.Rollback()
	purchase, err := scanPurchase(tx.QueryRowContext(ctx, purchaseSelect+` WHERE purchases.id=?`, purchaseID))
	if err != nil {
		return model.Purchase{}, err
	}
	if purchase.Status == "cancelled" {
		return purchase, nil
	}
	if purchase.Status == "expired" || purchase.Status == "failed" {
		return model.Purchase{}, ErrConflict
	}
	previousStatus := purchase.Status
	if _, err := tx.ExecContext(ctx, `UPDATE purchases SET status='cancelled',updated_at=? WHERE id=?`, stamp(now), purchase.ID); err != nil {
		return model.Purchase{}, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE balances SET txb_minor=txb_minor+?,updated_at=? WHERE user_id=?`, purchase.PriceTXBMinor, stamp(now), purchase.UserID); err != nil {
		return model.Purchase{}, err
	}
	var balance int64
	if err := tx.QueryRowContext(ctx, `SELECT txb_minor FROM balances WHERE user_id=?`, purchase.UserID).Scan(&balance); err != nil {
		return model.Purchase{}, err
	}
	if _, err := insertLedgerTx(ctx, tx, purchase.UserID, purchase.PriceTXBMinor, balance, "admin_entitlement_cancellation", purchase.ID, reason, now); err != nil {
		return model.Purchase{}, err
	}
	if previousStatus != "queued" {
		if err := insertOutboxTx(ctx, tx, "remna_sync_user", purchase.UserID, `{"userId":"`+purchase.UserID+`"}`, now, now); err != nil {
			return model.Purchase{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return model.Purchase{}, err
	}
	return s.PurchaseByID(ctx, purchase.ID)
}

const purchaseSelect = `SELECT purchases.id,purchases.user_id,purchases.combo_id,combos.name,purchases.price_txb_minor,purchases.valid_from,purchases.valid_until,
	purchases.status,purchases.traffic_limit_bytes,purchases.reset_strategy,purchases.created_at,purchases.updated_at FROM purchases JOIN combos ON combos.id=purchases.combo_id`

func scanPurchase(row rowScanner) (model.Purchase, error) {
	var purchase model.Purchase
	var validFrom, validUntil, created, updated string
	if err := row.Scan(&purchase.ID, &purchase.UserID, &purchase.ComboID, &purchase.ComboName, &purchase.PriceTXBMinor,
		&validFrom, &validUntil, &purchase.Status, &purchase.TrafficLimitBytes, &purchase.ResetStrategy, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Purchase{}, ErrNotFound
		}
		return model.Purchase{}, fmt.Errorf("scan purchase: %w", err)
	}
	purchase.Price = model.TXBMoney(purchase.PriceTXBMinor)
	purchase.TrafficLimit = strconv.FormatInt(purchase.TrafficLimitBytes, 10)
	var err error
	if purchase.ValidFrom, err = parseStamp(validFrom); err != nil {
		return model.Purchase{}, err
	}
	if purchase.ValidUntil, err = parseStamp(validUntil); err != nil {
		return model.Purchase{}, err
	}
	if purchase.CreatedAt, err = parseStamp(created); err != nil {
		return model.Purchase{}, err
	}
	purchase.UpdatedAt, err = parseStamp(updated)
	return purchase, err
}

func (s *Store) purchaseSquads(ctx context.Context, purchaseID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT remna_squad_uuid FROM purchase_squads WHERE purchase_id=? ORDER BY remna_squad_uuid`, purchaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]string, 0)
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

// ActiveAndQueuedPurchases returns the current home-screen entitlement summary.
func (s *Store) ActiveAndQueuedPurchases(ctx context.Context, userID string, now time.Time) (*model.Purchase, *model.Purchase, error) {
	purchases, err := s.ListPurchases(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	var active, queued *model.Purchase
	for index := range purchases {
		purchase := purchases[index]
		if (purchase.Status == "active" || purchase.Status == "activating") && !purchase.ValidUntil.Before(now) && active == nil {
			copy := purchase
			active = &copy
		}
		if purchase.Status == "queued" && queued == nil {
			copy := purchase
			queued = &copy
		}
	}
	return active, queued, nil
}

// CreatePaymentOrder persists an attempt before contacting a provider, so immediate callbacks are safe.
func (s *Store) CreatePaymentOrder(ctx context.Context, order model.PaymentOrder) (model.PaymentOrder, error) {
	if order.ID == "" {
		var err error
		order.ID, err = ids.New()
		if err != nil {
			return model.PaymentOrder{}, err
		}
	}
	now := time.Now().UTC()
	if order.Status == "" {
		order.Status = "creating"
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO payment_orders(id,user_id,provider,status,txb_minor,payable_amount,payable_currency,rate_snapshot,provider_payload,expires_at,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, order.ID, order.UserID, order.Provider, order.Status, order.TXBMinor, order.PayableAmount,
		order.PayableCurrency, order.RateSnapshot, order.ProviderPayload, stamp(order.ExpiresAt), stamp(now), stamp(now))
	if err != nil {
		return model.PaymentOrder{}, fmt.Errorf("create payment order: %w", err)
	}
	return s.PaymentOrderByID(ctx, order.ID)
}

// UpdatePaymentCheckout stores the provider response without changing the requested TXB amount.
func (s *Store) UpdatePaymentCheckout(ctx context.Context, orderID string, tradeID, paymentURL, qrPayload *string, payableAmount, payableCurrency, providerPayload string, expiresAt time.Time) (model.PaymentOrder, error) {
	_, err := s.db.ExecContext(ctx, `UPDATE payment_orders SET status='pending',provider_trade_id=?,payment_url=?,qr_payload=?,payable_amount=?,payable_currency=?,provider_payload=?,expires_at=?,updated_at=? WHERE id=? AND status='creating'`,
		tradeID, paymentURL, qrPayload, payableAmount, payableCurrency, providerPayload, stamp(expiresAt), stamp(time.Now().UTC()), orderID)
	if err != nil {
		return model.PaymentOrder{}, fmt.Errorf("update payment checkout: %w", err)
	}
	return s.PaymentOrderByID(ctx, orderID)
}

// FailPaymentOrder records a sanitized provider creation failure.
func (s *Store) FailPaymentOrder(ctx context.Context, orderID, providerPayload string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE payment_orders SET status='failed',provider_payload=?,updated_at=? WHERE id=? AND status='creating'`, providerPayload, stamp(time.Now().UTC()), orderID)
	return err
}

// ExpirePaymentOrder records an authoritative provider timeout without affecting balance.
func (s *Store) ExpirePaymentOrder(ctx context.Context, orderID, provider string, now time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE payment_orders SET status='expired',updated_at=? WHERE id=? AND provider=? AND status IN ('creating','pending')`, stamp(now), orderID, provider)
	if err != nil {
		return fmt.Errorf("expire payment order: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		order, loadErr := s.PaymentOrderByID(ctx, orderID)
		if loadErr != nil {
			return loadErr
		}
		if order.Status != "paid" && order.Status != "refunded" && order.Status != "expired" {
			return ErrConflict
		}
	}
	return nil
}

// ExpireStalePaymentOrders closes locally expired attempts without crediting them.
func (s *Store) ExpireStalePaymentOrders(ctx context.Context, now time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE payment_orders SET status='expired',updated_at=? WHERE status IN ('creating','pending') AND expires_at<=?`, stamp(now), stamp(now))
	if err != nil {
		return fmt.Errorf("expire stale payment orders: %w", err)
	}
	return nil
}

// PaymentOrderByID loads one payment order.
func (s *Store) PaymentOrderByID(ctx context.Context, id string) (model.PaymentOrder, error) {
	return scanPaymentOrder(s.db.QueryRowContext(ctx, paymentSelect+` WHERE id=?`, id))
}

// PaymentOrderForUser prevents order-ID enumeration across accounts.
func (s *Store) PaymentOrderForUser(ctx context.Context, id, userID string) (model.PaymentOrder, error) {
	return scanPaymentOrder(s.db.QueryRowContext(ctx, paymentSelect+` WHERE id=? AND user_id=?`, id, userID))
}

// ListPaymentOrders returns recent payment attempts.
func (s *Store) ListPaymentOrders(ctx context.Context, userID string, limit int) ([]model.PaymentOrder, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	query := paymentSelect
	args := []any{}
	if userID != "" {
		query += ` WHERE user_id=?`
		args = append(args, userID)
	}
	query += ` ORDER BY created_at DESC LIMIT ?`
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orders := make([]model.PaymentOrder, 0)
	for rows.Next() {
		order, err := scanPaymentOrder(rows)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

const paymentSelect = `SELECT id,user_id,provider,status,txb_minor,payable_amount,payable_currency,rate_snapshot,provider_trade_id,provider_charge_id,payment_url,qr_payload,provider_payload,expires_at,paid_at,refunded_at,created_at,updated_at FROM payment_orders`

func scanPaymentOrder(row rowScanner) (model.PaymentOrder, error) {
	var order model.PaymentOrder
	var tradeID, chargeID, paymentURL, qr, paid, refunded sql.NullString
	var expires, created, updated string
	if err := row.Scan(&order.ID, &order.UserID, &order.Provider, &order.Status, &order.TXBMinor, &order.PayableAmount,
		&order.PayableCurrency, &order.RateSnapshot, &tradeID, &chargeID, &paymentURL, &qr, &order.ProviderPayload,
		&expires, &paid, &refunded, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PaymentOrder{}, ErrNotFound
		}
		return model.PaymentOrder{}, fmt.Errorf("scan payment order: %w", err)
	}
	order.TXB = model.TXBMoney(order.TXBMinor)
	order.ProviderTradeID = nullableString(tradeID)
	order.ProviderChargeID = nullableString(chargeID)
	order.PaymentURL = nullableString(paymentURL)
	order.QRPayload = nullableString(qr)
	var err error
	if order.ExpiresAt, err = parseStamp(expires); err != nil {
		return model.PaymentOrder{}, err
	}
	if paid.Valid {
		value, err := parseStamp(paid.String)
		if err != nil {
			return model.PaymentOrder{}, err
		}
		order.PaidAt = &value
	}
	if refunded.Valid {
		value, err := parseStamp(refunded.String)
		if err != nil {
			return model.PaymentOrder{}, err
		}
		order.RefundedAt = &value
	}
	if order.CreatedAt, err = parseStamp(created); err != nil {
		return model.PaymentOrder{}, err
	}
	order.UpdatedAt, err = parseStamp(updated)
	return order, err
}

// SettlePayment records one authoritative provider event and credits the exact requested TXB amount once.
func (s *Store) SettlePayment(ctx context.Context, provider, dedupeKey, payloadHash, orderID, tradeID, chargeID string, now time.Time) (model.PaymentOrder, bool, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.PaymentOrder{}, false, err
	}
	defer tx.Rollback()
	var existingOrderID string
	err = tx.QueryRowContext(ctx, `SELECT order_id FROM webhook_events WHERE provider=? AND dedupe_key=?`, provider, dedupeKey).Scan(&existingOrderID)
	if err == nil {
		if existingOrderID != orderID {
			return model.PaymentOrder{}, false, ErrConflict
		}
		order, loadErr := paymentOrderTx(ctx, tx, orderID)
		if loadErr == nil && order.ProviderTradeID != nil && tradeID != "" && *order.ProviderTradeID != tradeID {
			return model.PaymentOrder{}, false, ErrConflict
		}
		return order, false, loadErr
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return model.PaymentOrder{}, false, err
	}
	eventID, err := ids.New()
	if err != nil {
		return model.PaymentOrder{}, false, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO webhook_events(id,provider,dedupe_key,order_id,payload_hash,received_at) VALUES(?,?,?,?,?,?)`, eventID, provider, dedupeKey, orderID, payloadHash, stamp(now)); err != nil {
		return model.PaymentOrder{}, false, err
	}
	order, err := paymentOrderTx(ctx, tx, orderID)
	if err != nil {
		return model.PaymentOrder{}, false, err
	}
	if order.Provider != provider {
		return model.PaymentOrder{}, false, ErrConflict
	}
	if order.Status == "paid" || order.Status == "refunded" {
		if order.ProviderTradeID != nil && tradeID != "" && *order.ProviderTradeID != tradeID {
			return model.PaymentOrder{}, false, ErrConflict
		}
		return order, false, nil
	}
	if order.Status != "pending" && order.Status != "creating" && order.Status != "expired" {
		return model.PaymentOrder{}, false, ErrConflict
	}
	if _, err := tx.ExecContext(ctx, `UPDATE payment_orders SET status='paid',provider_trade_id=COALESCE(NULLIF(?,''),provider_trade_id),provider_charge_id=NULLIF(?,''),paid_at=?,updated_at=? WHERE id=?`, tradeID, chargeID, stamp(now), stamp(now), order.ID); err != nil {
		return model.PaymentOrder{}, false, fmt.Errorf("mark payment paid: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO balances(user_id,txb_minor,updated_at) VALUES(?,?,?) ON CONFLICT(user_id) DO UPDATE SET txb_minor=balances.txb_minor+excluded.txb_minor,updated_at=excluded.updated_at`, order.UserID, order.TXBMinor, stamp(now)); err != nil {
		return model.PaymentOrder{}, false, err
	}
	var balance int64
	if err := tx.QueryRowContext(ctx, `SELECT txb_minor FROM balances WHERE user_id=?`, order.UserID).Scan(&balance); err != nil {
		return model.PaymentOrder{}, false, err
	}
	if _, err := insertLedgerTx(ctx, tx, order.UserID, order.TXBMinor, balance, "payment_credit", order.ID, provider+" payment", now); err != nil {
		return model.PaymentOrder{}, false, err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE webhook_events SET processed_at=? WHERE id=?`, stamp(now), eventID); err != nil {
		return model.PaymentOrder{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return model.PaymentOrder{}, false, err
	}
	settled, err := s.PaymentOrderByID(ctx, order.ID)
	return settled, true, err
}

func paymentOrderTx(ctx context.Context, tx *sql.Tx, id string) (model.PaymentOrder, error) {
	return scanPaymentOrder(tx.QueryRowContext(ctx, paymentSelect+` WHERE id=?`, id))
}

// RefundPayment appends a reversal and unwinds entitlements until the account is solvent or no purchases remain.
func (s *Store) RefundPayment(ctx context.Context, actorID *string, orderID, reason string, now time.Time) (model.PaymentOrder, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.PaymentOrder{}, err
	}
	defer tx.Rollback()
	order, err := paymentOrderTx(ctx, tx, orderID)
	if err != nil {
		return model.PaymentOrder{}, err
	}
	if order.Status == "refunded" {
		return order, nil
	}
	if order.Status != "paid" {
		return model.PaymentOrder{}, ErrConflict
	}
	refundID, err := ids.New()
	if err != nil {
		return model.PaymentOrder{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO refunds(id,payment_order_id,actor_user_id,txb_minor,reason,status,created_at) VALUES(?,?,?,?,?,'completed',?)`, refundID, order.ID, actorID, order.TXBMinor, reason, stamp(now)); err != nil {
		return model.PaymentOrder{}, fmt.Errorf("record refund: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE balances SET txb_minor=txb_minor-?,updated_at=? WHERE user_id=?`, order.TXBMinor, stamp(now), order.UserID); err != nil {
		return model.PaymentOrder{}, err
	}
	var balance int64
	if err := tx.QueryRowContext(ctx, `SELECT txb_minor FROM balances WHERE user_id=?`, order.UserID).Scan(&balance); err != nil {
		return model.PaymentOrder{}, err
	}
	if _, err := insertLedgerTx(ctx, tx, order.UserID, -order.TXBMinor, balance, "payment_reversal", order.ID, reason, now); err != nil {
		return model.PaymentOrder{}, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id,price_txb_minor,status FROM purchases WHERE user_id=? AND status IN ('queued','activating','active') ORDER BY CASE status WHEN 'queued' THEN 0 ELSE 1 END,created_at DESC`, order.UserID)
	if err != nil {
		return model.PaymentOrder{}, err
	}
	type cancellation struct {
		id, status string
		price      int64
	}
	cancellations := make([]cancellation, 0)
	for rows.Next() {
		var item cancellation
		if err := rows.Scan(&item.id, &item.price, &item.status); err != nil {
			rows.Close()
			return model.PaymentOrder{}, err
		}
		cancellations = append(cancellations, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return model.PaymentOrder{}, err
	}
	rows.Close()
	for _, item := range cancellations {
		if balance >= 0 {
			break
		}
		if _, err := tx.ExecContext(ctx, `UPDATE purchases SET status='cancelled',updated_at=? WHERE id=?`, stamp(now), item.id); err != nil {
			return model.PaymentOrder{}, err
		}
		balance += item.price
		if _, err := tx.ExecContext(ctx, `UPDATE balances SET txb_minor=?,updated_at=? WHERE user_id=?`, balance, stamp(now), order.UserID); err != nil {
			return model.PaymentOrder{}, err
		}
		if _, err := insertLedgerTx(ctx, tx, order.UserID, item.price, balance, "purchase_cancellation", item.id, "cancelled after payment refund", now); err != nil {
			return model.PaymentOrder{}, err
		}
		if item.status != "queued" {
			if err := insertOutboxTx(ctx, tx, "remna_sync_user", order.UserID, `{"userId":"`+order.UserID+`"}`, now, now); err != nil {
				return model.PaymentOrder{}, err
			}
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE payment_orders SET status='refunded',refunded_at=?,updated_at=? WHERE id=?`, stamp(now), stamp(now), order.ID); err != nil {
		return model.PaymentOrder{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.PaymentOrder{}, err
	}
	return s.PaymentOrderByID(ctx, order.ID)
}

// ListRefunds returns immutable refund records newest first.
func (s *Store) ListRefunds(ctx context.Context, limit int) ([]model.Refund, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,payment_order_id,actor_user_id,txb_minor,reason,status,created_at FROM refunds ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	refunds := make([]model.Refund, 0)
	for rows.Next() {
		var refund model.Refund
		var txbMinor int64
		var created string
		var actor sql.NullString
		if err := rows.Scan(&refund.ID, &refund.PaymentOrderID, &actor, &txbMinor, &refund.Reason, &refund.Status, &created); err != nil {
			return nil, err
		}
		refund.ActorUserID = nullableString(actor)
		refund.TXB = model.TXBMoney(txbMinor)
		refund.CreatedAt, err = parseStamp(created)
		if err != nil {
			return nil, err
		}
		refunds = append(refunds, refund)
	}
	return refunds, rows.Err()
}
