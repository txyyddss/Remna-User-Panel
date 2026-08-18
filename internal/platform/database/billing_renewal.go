package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/txyyddss/Remna-User-Panel/internal/coupons"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
	"strings"
	"time"
)

// RenewalInput describes one contiguous renewal batch.
type RenewalInput struct {
	UserID, PurchaseID, IdempotencyKey string
	TermCount                          int
}

func (s *Store) RenewalQuote(ctx context.Context, userID, purchaseID string, termCount int, now time.Time) (model.RenewalQuote, error) {
	if termCount < 1 || termCount > 6 {
		return model.RenewalQuote{}, ErrConflict
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return model.RenewalQuote{}, err
	}
	defer func() { _ = tx.Rollback() }()
	quote, _, _, err := renewalQuoteTx(ctx, tx, userID, purchaseID, termCount, now.UTC())
	return quote, err
}

func renewalQuoteTx(ctx context.Context, tx *sql.Tx, userID, purchaseID string, termCount int, now time.Time) (model.RenewalQuote, model.Combo, []model.SquadProduct, error) {
	purchase, err := scanPurchase(tx.QueryRowContext(ctx, purchaseSelect+` WHERE purchases.id=? AND purchases.user_id=?`, purchaseID, userID))
	if err != nil {
		return model.RenewalQuote{}, model.Combo{}, nil, err
	}
	if purchase.Status != "active" && purchase.Status != "activating" && purchase.Status != "queued" {
		return model.RenewalQuote{}, model.Combo{}, nil, ErrConflict
	}
	combo, err := comboByIDTx(ctx, tx, purchase.ComboID, true)
	if err != nil {
		return model.RenewalQuote{}, model.Combo{}, nil, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT remna_squad_uuid FROM purchase_addons WHERE purchase_id=? ORDER BY remna_squad_uuid`, purchaseID)
	if err != nil {
		return model.RenewalQuote{}, model.Combo{}, nil, err
	}
	addons := make([]model.SquadProduct, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return model.RenewalQuote{}, model.Combo{}, nil, err
		}
		product, loadErr := squadByIDTx(ctx, tx, id, false)
		if loadErr != nil {
			_ = rows.Close()
			return model.RenewalQuote{}, model.Combo{}, nil, loadErr
		}
		addons = append(addons, product)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return model.RenewalQuote{}, model.Combo{}, nil, err
	}
	_ = rows.Close()
	// Included squads reserve capacity only on the initial purchase. A renewal
	// rechecks its paid add-ons without displacing the current member.
	selected := make([]string, 0, len(addons))
	for _, squad := range addons {
		selected = append(selected, squad.RemnaSquadUUID)
	}
	if err := checkSquadStockTx(ctx, tx, selected, userID); err != nil {
		return model.RenewalQuote{}, model.Combo{}, nil, err
	}
	start, err := nextPurchaseStartTx(ctx, tx, userID, now)
	if err != nil {
		return model.RenewalQuote{}, model.Combo{}, nil, err
	}
	grossPerTerm := combo.PriceTXBMinor
	addonIDs := make([]string, 0, len(addons))
	for _, addon := range addons {
		grossPerTerm += addon.PriceTXBMinor
		addonIDs = append(addonIDs, addon.RemnaSquadUUID)
	}
	discount := coupons.Discount{GrossMinor: grossPerTerm, NetMinor: grossPerTerm}
	if purchase.RecurringDiscountAttached {
		if purchase.CouponGrantID == nil {
			return model.RenewalQuote{}, model.Combo{}, nil, ErrConflict
		}
		discount, err = quoteRecurringRenewalCouponTx(ctx, tx, userID, *purchase.CouponGrantID, grossPerTerm)
		if err != nil {
			return model.RenewalQuote{}, model.Combo{}, nil, err
		}
	}
	var couponGrantID *string
	if discount.Recurring && discount.GrantID != "" {
		couponGrantID = &discount.GrantID
	}
	return model.RenewalQuote{PurchaseID: purchaseID, ComboID: combo.ID, TermCount: termCount, GrossPrice: model.TXBMoney(grossPerTerm), Discount: model.TXBMoney(discount.DiscountMinor), PricePerTerm: model.TXBMoney(discount.NetMinor), TotalPrice: model.TXBMoney(discount.NetMinor * int64(termCount)), CouponGrantID: couponGrantID, EffectiveAt: start,
		ExpiresAt: start.AddDate(0, 0, combo.ValidityDays*termCount), AddonSquadUUIDs: addonIDs}, combo, addons, nil
}

// Renew creates all terms and one ledger debit in one SQLite transaction.
func (s *Store) Renew(ctx context.Context, input RenewalInput, now time.Time) (model.RenewalBatch, error) {
	if input.TermCount < 1 || input.TermCount > 6 || strings.TrimSpace(input.IdempotencyKey) == "" {
		return model.RenewalBatch{}, ErrConflict
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return model.RenewalBatch{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var existingID, existingFingerprint string
	fingerprint := fmt.Sprintf("%s:%d", input.PurchaseID, input.TermCount)
	err = tx.QueryRowContext(ctx, `SELECT id,request_fingerprint FROM renewal_batches WHERE user_id=? AND idempotency_key=?`, input.UserID, input.IdempotencyKey).Scan(&existingID, &existingFingerprint)
	if err == nil {
		if existingID == "" || existingFingerprint != fingerprint {
			return model.RenewalBatch{}, ErrConflict
		}
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return model.RenewalBatch{}, rollbackErr
		}
		return s.loadRenewalBatch(ctx, existingID)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return model.RenewalBatch{}, err
	}
	quote, combo, addons, err := renewalQuoteTx(ctx, tx, input.UserID, input.PurchaseID, input.TermCount, now.UTC())
	if err != nil {
		return model.RenewalBatch{}, err
	}
	batchID, err := ids.New()
	if err != nil {
		return model.RenewalBatch{}, err
	}
	newBalance, err := debitBalanceTx(ctx, tx, input.UserID, quote.TotalPrice.MinorInt64(), now.UTC())
	if err != nil {
		return model.RenewalBatch{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO renewal_batches(id,user_id,source_purchase_id,idempotency_key,request_fingerprint,term_count,charged_txb_minor,created_at) VALUES(?,?,?,?,?,?,?,?)`, batchID, input.UserID, input.PurchaseID, input.IdempotencyKey, fingerprint, input.TermCount, quote.TotalPrice.MinorInt64(), stamp(now)); err != nil {
		return model.RenewalBatch{}, err
	}
	start := quote.EffectiveAt
	idsForBatch := make([]string, 0, input.TermCount)
	for index := 0; index < input.TermCount; index++ {
		purchaseID, err := ids.New()
		if err != nil {
			return model.RenewalBatch{}, err
		}
		from := start.AddDate(0, 0, combo.ValidityDays*index)
		until := from.AddDate(0, 0, combo.ValidityDays)
		status := "queued"
		if !from.After(now) {
			status = "activating"
		}
		requestKey := any(nil)
		if index == 0 {
			requestKey = input.IdempotencyKey
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO purchases(id,user_id,combo_id,charged_txb_minor,valid_from,valid_until,status,coupon_grant_id,gross_price_txb_minor,core_gross_txb_minor,coupon_discount_txb_minor,recurring_discount_attached,idempotency_key,request_fingerprint,renewal_batch_id,renewal_index,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, purchaseID, input.UserID, combo.ID, quote.PricePerTerm.MinorInt64(), stamp(from), stamp(until), status, quote.CouponGrantID, quote.GrossPrice.MinorInt64(), combo.PriceTXBMinor, quote.Discount.MinorInt64(), boolInt(quote.CouponGrantID != nil), requestKey, fingerprint, batchID, index, stamp(now), stamp(now)); err != nil {
			return model.RenewalBatch{}, err
		}
		for _, addon := range addons {
			if _, err := tx.ExecContext(ctx, `INSERT INTO purchase_addons(purchase_id,remna_squad_uuid,charged_txb_minor) VALUES(?,?,?)`, purchaseID, addon.RemnaSquadUUID, addon.PriceTXBMinor); err != nil {
				return model.RenewalBatch{}, err
			}
		}
		if err := enqueuePurchaseTransitionTx(ctx, tx, purchaseID, status, from, now); err != nil {
			return model.RenewalBatch{}, err
		}
		idsForBatch = append(idsForBatch, purchaseID)
	}
	if _, err := insertLedgerTx(ctx, tx, input.UserID, -quote.TotalPrice.MinorInt64(), newBalance, "purchase_debit", batchID, "renewal batch", now); err != nil {
		return model.RenewalBatch{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.RenewalBatch{}, err
	}
	return s.loadRenewalBatch(ctx, batchID)
}
