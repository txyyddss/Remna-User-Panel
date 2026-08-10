package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/coupons"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
)

// PurchaseInput contains the server-selected catalog IDs for a checkout.
type PurchaseInput struct {
	UserID         string
	ComboID        string
	AddonSquadIDs  []string
	CouponGrantID  string
	IdempotencyKey string
}

// CreatePurchase atomically revalidates live pricing, debits TXB, stores only
// transaction facts and selected paid add-ons, and queues Remnawave sync.
func (s *Store) CreatePurchase(ctx context.Context, input PurchaseInput, now time.Time) (model.Purchase, error) {
	fingerprint, err := normalizeAndFingerprintPurchase(&input)
	if err != nil {
		return model.Purchase{}, err
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	now = now.UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.Purchase{}, fmt.Errorf("begin purchase: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var existingID, existingFingerprint string
	loadErr := tx.QueryRowContext(ctx, `SELECT id,request_fingerprint FROM purchases WHERE user_id=? AND idempotency_key=?`, input.UserID, input.IdempotencyKey).Scan(&existingID, &existingFingerprint)
	if loadErr == nil {
		if existingFingerprint != fingerprint {
			return model.Purchase{}, ErrConflict
		}
		// End the read transaction before loading the full aggregate (including
		// purchase_squads) through the normal repository query path.
		if err := tx.Rollback(); err != nil {
			return model.Purchase{}, fmt.Errorf("close purchase replay transaction: %w", err)
		}
		return s.PurchaseByID(ctx, existingID)
	}
	if !errors.Is(loadErr, sql.ErrNoRows) {
		return model.Purchase{}, fmt.Errorf("load purchase idempotency key: %w", loadErr)
	}

	quote, combo, addonRows, err := quotePurchaseTx(ctx, tx, input, now)
	if err != nil {
		return model.Purchase{}, err
	}
	validFrom := quote.EffectiveAt
	validUntil := quote.ExpiresAt
	status := "queued"
	if !validFrom.After(now) {
		status = "activating"
	}

	purchaseID, err := ids.New()
	if err != nil {
		return model.Purchase{}, err
	}
	discount := coupons.Discount{GrossMinor: quote.GrossPriceTXBMinor, NetMinor: quote.GrossPriceTXBMinor}
	if input.CouponGrantID != "" {
		pricedAddonIDs := make([]string, 0, len(addonRows))
		for _, addon := range addonRows {
			pricedAddonIDs = append(pricedAddonIDs, addon.RemnaSquadUUID)
		}
		discount, err = applyPurchaseCouponTx(ctx, tx, coupons.PurchaseContext{
			UserID: input.UserID, GrantID: input.CouponGrantID, ComboID: combo.ID,
			AddonSquadIDs: pricedAddonIDs, GrossPriceMinor: quote.GrossPriceTXBMinor,
		}, purchaseID, now)
		if err != nil {
			return model.Purchase{}, err
		}
	}
	netPrice := discount.NetMinor
	newBalance, err := debitBalanceTx(ctx, tx, input.UserID, netPrice, now)
	if err != nil {
		return model.Purchase{}, err
	}
	var couponGrantID any
	if input.CouponGrantID != "" {
		couponGrantID = input.CouponGrantID
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO purchases(id,user_id,combo_id,charged_txb_minor,valid_from,valid_until,status,coupon_grant_id,gross_price_txb_minor,coupon_discount_txb_minor,idempotency_key,request_fingerprint,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, purchaseID, input.UserID, combo.ID, netPrice, stamp(validFrom), stamp(validUntil), status,
		couponGrantID, quote.GrossPriceTXBMinor, discount.DiscountMinor, input.IdempotencyKey, fingerprint, stamp(now), stamp(now))
	if err != nil {
		return model.Purchase{}, fmt.Errorf("insert purchase: %w", err)
	}
	if status == "activating" {
		if err := applyPendingExtensionsToActivationTx(ctx, tx, purchaseID, now); err != nil {
			return model.Purchase{}, fmt.Errorf("apply pending subscription extensions: %w", err)
		}
	}
	for _, product := range addonRows {
		if _, err := tx.ExecContext(ctx, `INSERT INTO purchase_addons(purchase_id,remna_squad_uuid,charged_txb_minor) VALUES(?,?,?)`, purchaseID, product.RemnaSquadUUID, product.PriceTXBMinor); err != nil {
			return model.Purchase{}, fmt.Errorf("snapshot add-on squad: %w", err)
		}
	}
	if _, err := insertLedgerTx(ctx, tx, input.UserID, -netPrice, newBalance, "purchase_debit", purchaseID, combo.Name, now); err != nil {
		return model.Purchase{}, err
	}
	if status == "activating" {
		if err := insertOutboxTx(ctx, tx, "remna_apply_entitlement", `{"purchaseId":"`+purchaseID+`"}`, now, now); err != nil {
			return model.Purchase{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return model.Purchase{}, fmt.Errorf("commit purchase: %w", err)
	}
	return s.PurchaseByID(ctx, purchaseID)
}

// QuotePurchase performs the same catalog, coupon, and effective-date checks as
// checkout without changing balances, coupon grants, or entitlement state.
func (s *Store) QuotePurchase(ctx context.Context, input PurchaseInput, now time.Time) (model.PurchaseQuote, error) {
	input.IdempotencyKey = "quote"
	if _, err := normalizeAndFingerprintPurchase(&input); err != nil {
		return model.PurchaseQuote{}, err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return model.PurchaseQuote{}, fmt.Errorf("begin purchase quote: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	quote, _, _, err := quotePurchaseTx(ctx, tx, input, now.UTC())
	return quote, err
}

func quotePurchaseTx(ctx context.Context, tx *sql.Tx, input PurchaseInput, now time.Time) (model.PurchaseQuote, model.Combo, []model.SquadProduct, error) {
	combo, err := comboByIDTx(ctx, tx, input.ComboID, true)
	if err != nil {
		return model.PurchaseQuote{}, model.Combo{}, nil, err
	}
	includedUUIDs := make(map[string]struct{}, len(combo.IncludedSquads))
	for _, product := range combo.IncludedSquads {
		includedUUIDs[product.RemnaSquadUUID] = struct{}{}
	}
	addonRows := make([]model.SquadProduct, 0, len(input.AddonSquadIDs))
	addonPrice := int64(0)
	for _, squadUUID := range uniqueSorted(input.AddonSquadIDs) {
		product, loadErr := squadByIDTx(ctx, tx, squadUUID, true)
		if loadErr != nil {
			return model.PurchaseQuote{}, model.Combo{}, nil, loadErr
		}
		if _, included := includedUUIDs[product.RemnaSquadUUID]; included {
			continue
		}
		includedUUIDs[product.RemnaSquadUUID] = struct{}{}
		addonRows = append(addonRows, product)
		addonPrice += product.PriceTXBMinor
	}
	grossMinor := combo.PriceTXBMinor + addonPrice
	validFrom, err := nextPurchaseStartTx(ctx, tx, input.UserID, now)
	if err != nil {
		return model.PurchaseQuote{}, model.Combo{}, nil, err
	}
	discount := coupons.Discount{GrossMinor: grossMinor, NetMinor: grossMinor}
	addonUUIDs := make([]string, 0, len(addonRows))
	for _, addon := range addonRows {
		addonUUIDs = append(addonUUIDs, addon.RemnaSquadUUID)
	}
	if input.CouponGrantID != "" {
		discount, err = quoteCouponGrantTx(ctx, tx, coupons.PurchaseContext{
			UserID: input.UserID, GrantID: input.CouponGrantID, ComboID: combo.ID,
			AddonSquadIDs: addonUUIDs, GrossPriceMinor: grossMinor,
		}, now)
		if err != nil {
			return model.PurchaseQuote{}, model.Combo{}, nil, err
		}
	}
	return model.PurchaseQuote{
		ComboID: combo.ID, ComboName: combo.Name,
		GrossPriceTXBMinor: grossMinor, DiscountTXBMinor: discount.DiscountMinor, NetPriceTXBMinor: discount.NetMinor,
		GrossPrice: model.TXBMoney(grossMinor), Discount: model.TXBMoney(discount.DiscountMinor), NetPrice: model.TXBMoney(discount.NetMinor),
		EffectiveAt: validFrom, ExpiresAt: validFrom.AddDate(0, 0, combo.ValidityDays), Queued: validFrom.After(now), AddonSquadUUIDs: addonUUIDs,
	}, combo, addonRows, nil
}

func nextPurchaseStartTx(ctx context.Context, tx *sql.Tx, userID string, now time.Time) (time.Time, error) {
	validFrom := now
	var latestEnd string
	err := tx.QueryRowContext(ctx, `SELECT valid_until FROM purchases WHERE user_id=? AND status IN ('activating','active','queued') ORDER BY valid_until DESC LIMIT 1`, userID).Scan(&latestEnd)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, fmt.Errorf("find current entitlement: %w", err)
	}
	if err == nil {
		parsed, parseErr := parseStamp(latestEnd)
		if parseErr != nil {
			return time.Time{}, fmt.Errorf("parse current entitlement: %w", parseErr)
		}
		if parsed.After(validFrom) {
			validFrom = parsed
		}
	}
	return validFrom, nil
}

func normalizeAndFingerprintPurchase(input *PurchaseInput) (string, error) {
	if input == nil {
		return "", errors.New("purchase input is required")
	}
	input.UserID = strings.TrimSpace(input.UserID)
	input.ComboID = strings.TrimSpace(input.ComboID)
	input.CouponGrantID = strings.TrimSpace(input.CouponGrantID)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	if input.UserID == "" || input.ComboID == "" || input.IdempotencyKey == "" || len(input.IdempotencyKey) > 128 {
		return "", errors.New("purchase user, combo, and idempotency key are required")
	}
	normalizedAddons := make([]string, len(input.AddonSquadIDs))
	for index := range input.AddonSquadIDs {
		normalizedAddons[index] = strings.TrimSpace(input.AddonSquadIDs[index])
	}
	input.AddonSquadIDs = uniqueSorted(normalizedAddons)
	payload, err := json.Marshal(struct {
		Version       int      `json:"version"`
		ComboID       string   `json:"comboId"`
		AddonSquadIDs []string `json:"addonSquadIds"`
		CouponGrantID string   `json:"couponGrantId"`
	}{Version: 1, ComboID: input.ComboID, AddonSquadIDs: input.AddonSquadIDs, CouponGrantID: input.CouponGrantID})
	if err != nil {
		return "", fmt.Errorf("encode purchase request fingerprint: %w", err)
	}
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:]), nil
}

func comboByIDTx(ctx context.Context, tx *sql.Tx, id string, activeOnly bool) (model.Combo, error) {
	query := comboSelect + ` WHERE id=?`
	if activeOnly {
		query += ` AND active=1`
	}
	return scanCombo(tx.QueryRowContext(ctx, query, id))
}

func squadByIDTx(ctx context.Context, tx *sql.Tx, id string, requireVisible bool) (model.SquadProduct, error) {
	query := squadSelect + ` WHERE remna_squad_uuid=?`
	if requireVisible {
		query += ` AND visible=1`
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
	defer func() { _ = tx.Rollback() }()
	balance, err := adjustBalanceTx(ctx, tx, userID, delta, now)
	if err != nil {
		return model.LedgerEntry{}, fmt.Errorf("adjust balance: %w", err)
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

// DeductBalance appends an immutable administrator-authored debit without allowing debt.
func (s *Store) DeductBalance(ctx context.Context, userID string, amount int64, referenceID, note string, now time.Time) (model.LedgerEntry, error) {
	if amount <= 0 {
		return model.LedgerEntry{}, ErrInsufficientBalance
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.LedgerEntry{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var existingID string
	if err := tx.QueryRowContext(ctx, `SELECT id FROM ledger_entries WHERE kind='telegram_deduct' AND reference_id=? LIMIT 1`, referenceID).Scan(&existingID); err == nil {
		_ = tx.Rollback()
		return s.LedgerEntryByID(ctx, existingID)
	} else if !errors.Is(err, sql.ErrNoRows) {
		return model.LedgerEntry{}, err
	}
	balance, err := changeBalanceTx(ctx, tx, userID, -amount, now.UTC())
	if err != nil {
		return model.LedgerEntry{}, fmt.Errorf("deduct balance: %w", err)
	}
	entryID, err := insertLedgerTx(ctx, tx, userID, -amount, balance, "telegram_deduct", referenceID, note, now.UTC())
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
	defer func() { _ = rows.Close() }()
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

// PurchaseByID returns transaction facts combined with the combo's current
// live configuration, as required by the mutable-catalog contract.
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
	defer func() { _ = rows.Close() }()
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
	defer func() { _ = rows.Close() }()
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
	defer func() { _ = tx.Rollback() }()
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
	balance, err := adjustBalanceTx(ctx, tx, purchase.UserID, purchase.PriceTXBMinor, now)
	if err != nil {
		return model.Purchase{}, fmt.Errorf("refund cancelled purchase: %w", err)
	}
	if _, err := insertLedgerTx(ctx, tx, purchase.UserID, purchase.PriceTXBMinor, balance, "admin_entitlement_cancellation", purchase.ID, reason, now); err != nil {
		return model.Purchase{}, err
	}
	if previousStatus != "queued" {
		if err := insertOutboxTx(ctx, tx, "remna_sync_user", `{"userId":"`+purchase.UserID+`"}`, now, now); err != nil {
			return model.Purchase{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return model.Purchase{}, err
	}
	return s.PurchaseByID(ctx, purchase.ID)
}

const purchaseSelect = `SELECT purchases.id,purchases.user_id,purchases.combo_id,combos.name,purchases.charged_txb_minor,purchases.valid_from,purchases.valid_until,
	purchases.status,combos.traffic_limit_bytes,combos.reset_strategy,purchases.coupon_grant_id,COALESCE(purchases.gross_price_txb_minor,purchases.charged_txb_minor),purchases.coupon_discount_txb_minor,
	combos.rollover_min_remaining_bps,combos.rollover_max_txb_minor,purchases.created_at,purchases.updated_at FROM purchases JOIN combos ON combos.id=purchases.combo_id`

func scanPurchase(row rowScanner) (model.Purchase, error) {
	var purchase model.Purchase
	var validFrom, validUntil, created, updated string
	var couponGrantID sql.NullString
	if err := row.Scan(&purchase.ID, &purchase.UserID, &purchase.ComboID, &purchase.ComboName, &purchase.PriceTXBMinor,
		&validFrom, &validUntil, &purchase.Status, &purchase.TrafficLimitBytes, &purchase.ResetStrategy, &couponGrantID, &purchase.GrossPriceTXBMinor,
		&purchase.CouponDiscountTXBMinor, &purchase.RolloverMinRemainingBPS, &purchase.RolloverMaxTXBMinor, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Purchase{}, ErrNotFound
		}
		return model.Purchase{}, fmt.Errorf("scan purchase: %w", err)
	}
	purchase.Price = model.TXBMoney(purchase.PriceTXBMinor)
	purchase.GrossPrice = model.TXBMoney(purchase.GrossPriceTXBMinor)
	purchase.CouponDiscount = model.TXBMoney(purchase.CouponDiscountTXBMinor)
	purchase.CouponGrantID = nullableString(couponGrantID)
	purchase.RolloverMax = model.TXBMoney(purchase.RolloverMaxTXBMinor)
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
	rows, err := s.db.QueryContext(ctx, `SELECT value AS remna_squad_uuid
		FROM purchases JOIN combos ON combos.id=purchases.combo_id, json_each(combos.included_squad_uuids)
		WHERE purchases.id=?
		UNION
		SELECT remna_squad_uuid FROM purchase_addons WHERE purchase_id=?
		ORDER BY remna_squad_uuid`, purchaseID, purchaseID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
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
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
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
	if order.RateDirection == "" {
		order.RateDirection = "currency_per_txb"
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.PaymentOrder{}, fmt.Errorf("begin payment order retention: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `UPDATE payment_orders SET status='expired',updated_at=? WHERE status IN ('creating','pending') AND cancelled_at IS NULL AND expires_at<=?`, stamp(now), stamp(now)); err != nil {
		return model.PaymentOrder{}, fmt.Errorf("expire stale payment orders before insert: %w", err)
	}
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM payment_orders`).Scan(&count); err != nil {
		return model.PaymentOrder{}, fmt.Errorf("count payment orders: %w", err)
	}
	pruneCount := count - 199
	if pruneCount > 0 {
		rows, err := tx.QueryContext(ctx, `SELECT id FROM payment_orders
			WHERE status IN ('paid','expired','failed','refunded') OR (cancelled_at IS NOT NULL AND expires_at<=?)
			ORDER BY created_at,id LIMIT ?`, stamp(now), pruneCount)
		if err != nil {
			return model.PaymentOrder{}, fmt.Errorf("select prunable payment orders: %w", err)
		}
		idsToDelete := make([]string, 0, pruneCount)
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				_ = rows.Close()
				return model.PaymentOrder{}, fmt.Errorf("scan prunable payment order: %w", err)
			}
			idsToDelete = append(idsToDelete, id)
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return model.PaymentOrder{}, fmt.Errorf("iterate prunable payment orders: %w", err)
		}
		_ = rows.Close()
		if len(idsToDelete) != pruneCount {
			return model.PaymentOrder{}, ErrPaymentCapacity
		}
		for _, id := range idsToDelete {
			if _, err := tx.ExecContext(ctx, `DELETE FROM webhook_events WHERE order_id=?`, id); err != nil {
				return model.PaymentOrder{}, fmt.Errorf("prune payment webhook events: %w", err)
			}
			if _, err := tx.ExecContext(ctx, `DELETE FROM refunds WHERE payment_order_id=?`, id); err != nil {
				return model.PaymentOrder{}, fmt.Errorf("prune payment refunds: %w", err)
			}
			if _, err := tx.ExecContext(ctx, `DELETE FROM payment_orders WHERE id=?`, id); err != nil {
				return model.PaymentOrder{}, fmt.Errorf("prune payment order: %w", err)
			}
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO payment_orders(id,user_id,provider,method_id,provider_rail,status,txb_minor,payable_amount,payable_currency,rate_snapshot,rate_direction,provider_payload,expires_at,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, order.ID, order.UserID, order.Provider, order.MethodID, order.ProviderRail, order.Status, order.TXBMinor, order.PayableAmount,
		order.PayableCurrency, order.RateSnapshot, order.RateDirection, order.ProviderPayload, stamp(order.ExpiresAt), stamp(now), stamp(now))
	if err != nil {
		return model.PaymentOrder{}, fmt.Errorf("create payment order: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return model.PaymentOrder{}, fmt.Errorf("commit payment order: %w", err)
	}
	return s.PaymentOrderByID(ctx, order.ID)
}

// UpdatePaymentCheckout stores the provider response without changing the requested TXB amount.
func (s *Store) UpdatePaymentCheckoutDetails(ctx context.Context, orderID string, tradeID, paymentURL, qrPayload, receivingAddress, actualCryptoAmount, actualCryptoCurrency *string, payableAmount, payableCurrency, providerPayload string, expiresAt time.Time) (model.PaymentOrder, error) {
	_, err := s.db.ExecContext(ctx, `UPDATE payment_orders SET status='pending',provider_trade_id=?,payment_url=?,qr_payload=?,receiving_address=?,actual_crypto_amount=?,actual_crypto_currency=?,payable_amount=?,payable_currency=?,provider_payload=?,expires_at=?,updated_at=? WHERE id=? AND status='creating'`,
		tradeID, paymentURL, qrPayload, receivingAddress, actualCryptoAmount, actualCryptoCurrency, payableAmount, payableCurrency, providerPayload, stamp(expiresAt), stamp(time.Now().UTC()), orderID)
	if err != nil {
		return model.PaymentOrder{}, fmt.Errorf("update payment checkout: %w", err)
	}
	return s.PaymentOrderByID(ctx, orderID)
}

// UpdatePaymentCheckout preserves the original repository contract for
// adapters that do not return a separate receiving address.
func (s *Store) UpdatePaymentCheckout(ctx context.Context, orderID string, tradeID, paymentURL, qrPayload *string, payableAmount, payableCurrency, providerPayload string, expiresAt time.Time) (model.PaymentOrder, error) {
	return s.UpdatePaymentCheckoutDetails(ctx, orderID, tradeID, paymentURL, qrPayload, nil, nil, nil, payableAmount, payableCurrency, providerPayload, expiresAt)
}

// FailPaymentOrder records a sanitized provider creation failure.
func (s *Store) FailPaymentOrder(ctx context.Context, orderID, providerPayload string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE payment_orders SET status='failed',provider_payload=?,updated_at=? WHERE id=? AND status='creating'`, providerPayload, stamp(time.Now().UTC()), orderID)
	return err
}

// ExpirePaymentOrder records an authoritative provider timeout without affecting balance.
func (s *Store) ExpirePaymentOrder(ctx context.Context, orderID, provider string, now time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE payment_orders SET status='expired',updated_at=? WHERE id=? AND provider=? AND status IN ('creating','pending') AND cancelled_at IS NULL`, stamp(now), orderID, provider)
	if err != nil {
		return fmt.Errorf("expire payment order: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		order, loadErr := s.PaymentOrderByID(ctx, orderID)
		if loadErr != nil {
			return loadErr
		}
		if order.Status != "paid" && order.Status != "refunded" && order.Status != "expired" && order.Status != "cancelled" {
			return ErrConflict
		}
	}
	return nil
}

// ExpireStalePaymentOrders closes locally expired attempts without crediting them.
func (s *Store) ExpireStalePaymentOrders(ctx context.Context, now time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE payment_orders SET status='expired',updated_at=? WHERE status IN ('creating','pending') AND cancelled_at IS NULL AND expires_at<=?`, stamp(now), stamp(now))
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
	defer func() { _ = rows.Close() }()
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

const paymentSelect = `SELECT id,user_id,provider,method_id,provider_rail,status,txb_minor,payable_amount,payable_currency,rate_snapshot,rate_direction,provider_trade_id,provider_charge_id,payment_url,qr_payload,receiving_address,actual_crypto_amount,actual_crypto_currency,provider_payload,expires_at,paid_at,refunded_at,cancelled_at,cancel_reason,provider_cancel_status,created_at,updated_at FROM payment_orders`

func scanPaymentOrder(row rowScanner) (model.PaymentOrder, error) {
	var order model.PaymentOrder
	var tradeID, chargeID, paymentURL, qr, receivingAddress, actualCryptoAmount, actualCryptoCurrency, paid, refunded, cancelled sql.NullString
	var methodID, providerRail, rateDirection, cancelReason, providerCancelStatus sql.NullString
	var expires, created, updated string
	if err := row.Scan(&order.ID, &order.UserID, &order.Provider, &methodID, &providerRail, &order.Status, &order.TXBMinor, &order.PayableAmount,
		&order.PayableCurrency, &order.RateSnapshot, &rateDirection, &tradeID, &chargeID, &paymentURL, &qr, &receivingAddress, &actualCryptoAmount, &actualCryptoCurrency, &order.ProviderPayload,
		&expires, &paid, &refunded, &cancelled, &cancelReason, &providerCancelStatus, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PaymentOrder{}, ErrNotFound
		}
		return model.PaymentOrder{}, fmt.Errorf("scan payment order: %w", err)
	}
	order.TXB = model.TXBMoney(order.TXBMinor)
	order.MethodID = methodID.String
	order.ProviderRail = providerRail.String
	order.RateDirection = rateDirection.String
	order.ProviderTradeID = nullableString(tradeID)
	order.ProviderChargeID = nullableString(chargeID)
	order.PaymentURL = nullableString(paymentURL)
	order.QRPayload = nullableString(qr)
	order.ReceivingAddress = nullableString(receivingAddress)
	order.ActualCryptoAmount = nullableString(actualCryptoAmount)
	order.ActualCryptoCurrency = nullableString(actualCryptoCurrency)
	order.CancelReason = cancelReason.String
	order.ProviderCancelStatus = providerCancelStatus.String
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
	if cancelled.Valid {
		value, err := parseStamp(cancelled.String)
		if err != nil {
			return model.PaymentOrder{}, err
		}
		order.CancelledAt = &value
		if order.Status == "creating" || order.Status == "pending" {
			order.Status = "cancelled"
		}
	}
	if order.CreatedAt, err = parseStamp(created); err != nil {
		return model.PaymentOrder{}, err
	}
	order.UpdatedAt, err = parseStamp(updated)
	return order, err
}

// CancelPaymentOrder marks a user-owned unpaid attempt cancelled. It deliberately
// leaves the provider status payable so an authoritative late paid callback can
// still settle and credit the order exactly once.
func (s *Store) CancelPaymentOrder(ctx context.Context, orderID, userID, reason string, now time.Time) (model.PaymentOrder, bool, error) {
	result, err := s.db.ExecContext(ctx, `UPDATE payment_orders SET cancelled_at=?,cancel_reason=?,updated_at=?
		WHERE id=? AND user_id=? AND status IN ('creating','pending') AND paid_at IS NULL AND cancelled_at IS NULL`,
		stamp(now), reason, stamp(now), orderID, userID)
	if err != nil {
		return model.PaymentOrder{}, false, fmt.Errorf("cancel payment order: %w", err)
	}
	order, loadErr := s.PaymentOrderForUser(ctx, orderID, userID)
	if loadErr != nil {
		return model.PaymentOrder{}, false, loadErr
	}
	affected, _ := result.RowsAffected()
	if affected == 0 && order.Status != "cancelled" && order.Status != "paid" && order.Status != "refunded" {
		return model.PaymentOrder{}, false, ErrConflict
	}
	return order, affected == 1, nil
}

// SetPaymentProviderCancellation records the redacted result of a best-effort
// provider cancellation without changing the authoritative settlement state.
func (s *Store) SetPaymentProviderCancellation(ctx context.Context, orderID, status string, now time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE payment_orders SET provider_cancel_status=?,updated_at=? WHERE id=?`, status, stamp(now), orderID)
	if err != nil {
		return fmt.Errorf("record provider cancellation: %w", err)
	}
	return nil
}

// SettlePayment records one authoritative provider event and credits the exact requested TXB amount once.
func (s *Store) SettlePayment(ctx context.Context, provider, dedupeKey, payloadHash, orderID, tradeID, chargeID string, now time.Time) (model.PaymentOrder, bool, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.PaymentOrder{}, false, err
	}
	defer func() { _ = tx.Rollback() }()
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
	if order.ProviderTradeID != nil && tradeID != "" && *order.ProviderTradeID != tradeID {
		return model.PaymentOrder{}, false, ErrConflict
	}
	if order.ProviderChargeID != nil && chargeID != "" && *order.ProviderChargeID != chargeID {
		return model.PaymentOrder{}, false, ErrConflict
	}
	if order.Status == "paid" || order.Status == "refunded" {
		return order, false, nil
	}
	if order.Status != "pending" && order.Status != "creating" && order.Status != "expired" && order.Status != "cancelled" {
		return model.PaymentOrder{}, false, ErrConflict
	}
	if _, err := tx.ExecContext(ctx, `UPDATE payment_orders SET status='paid',provider_trade_id=COALESCE(NULLIF(?,''),provider_trade_id),provider_charge_id=NULLIF(?,''),paid_at=?,updated_at=? WHERE id=?`, tradeID, chargeID, stamp(now), stamp(now), order.ID); err != nil {
		return model.PaymentOrder{}, false, fmt.Errorf("mark payment paid: %w", err)
	}
	balance, err := adjustBalanceTx(ctx, tx, order.UserID, order.TXBMinor, now)
	if err != nil {
		return model.PaymentOrder{}, false, fmt.Errorf("credit payment: %w", err)
	}
	if _, err := insertLedgerTx(ctx, tx, order.UserID, order.TXBMinor, balance, "payment_credit", order.ID, provider+" payment", now); err != nil {
		return model.PaymentOrder{}, false, err
	}
	// The terminal order state plus provider trade/charge identifiers are the
	// durable replay guard. The webhook row is only a transient concurrency
	// claim and is removed in the same successful settlement transaction.
	if _, err := tx.ExecContext(ctx, `DELETE FROM webhook_events WHERE id=?`, eventID); err != nil {
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
	defer func() { _ = tx.Rollback() }()
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
	balance, err := adjustBalanceTx(ctx, tx, order.UserID, -order.TXBMinor, now)
	if err != nil {
		return model.PaymentOrder{}, fmt.Errorf("reverse payment balance: %w", err)
	}
	if _, err := insertLedgerTx(ctx, tx, order.UserID, -order.TXBMinor, balance, "payment_reversal", order.ID, reason, now); err != nil {
		return model.PaymentOrder{}, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT id,charged_txb_minor,status FROM purchases WHERE user_id=? AND status IN ('queued','activating','active') ORDER BY CASE status WHEN 'queued' THEN 0 ELSE 1 END,created_at DESC`, order.UserID)
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
			_ = rows.Close()
			return model.PaymentOrder{}, err
		}
		cancellations = append(cancellations, item)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return model.PaymentOrder{}, err
	}
	_ = rows.Close()
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
			if err := insertOutboxTx(ctx, tx, "remna_sync_user", `{"userId":"`+order.UserID+`"}`, now, now); err != nil {
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
	defer func() { _ = rows.Close() }()
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
