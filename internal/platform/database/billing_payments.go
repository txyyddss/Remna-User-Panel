package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
	"time"
)

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
	if _, err := tx.ExecContext(ctx, `UPDATE payment_orders SET status='expired',provider_payload='{}',payment_url=NULL,
		qr_payload=CASE WHEN receiving_address IS NOT NULL THEN NULL ELSE qr_payload END,updated_at=?
		WHERE status IN ('creating','pending') AND cancelled_at IS NULL AND expires_at<=?`, stamp(now), stamp(now)); err != nil {
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
		VALUES(?,?,?,?,?,?,?,?,?,?,?,'{}',?,?,?)`, order.ID, order.UserID, order.Provider, order.MethodID, order.ProviderRail, order.Status, order.TXBMinor, order.PayableAmount,
		order.PayableCurrency, order.RateSnapshot, order.RateDirection, stamp(order.ExpiresAt), stamp(now), stamp(now))
	if err != nil {
		return model.PaymentOrder{}, fmt.Errorf("create payment order: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return model.PaymentOrder{}, fmt.Errorf("commit payment order: %w", err)
	}
	return s.PaymentOrderByID(ctx, order.ID)
}

// UpdatePaymentCheckout stores the provider response without changing the requested TXB amount.
func (s *Store) UpdatePaymentCheckoutDetails(ctx context.Context, orderID string, tradeID, paymentURL, qrPayload, receivingAddress, actualCryptoAmount, actualCryptoCurrency *string, payableAmount, payableCurrency string, expiresAt time.Time) (model.PaymentOrder, error) {
	_, err := s.db.ExecContext(ctx, `UPDATE payment_orders SET status='pending',provider_trade_id=?,payment_url=?,qr_payload=?,receiving_address=?,actual_crypto_amount=?,actual_crypto_currency=?,payable_amount=?,payable_currency=?,provider_payload='{}',expires_at=?,updated_at=? WHERE id=? AND status='creating'`,
		tradeID, paymentURL, qrPayload, receivingAddress, actualCryptoAmount, actualCryptoCurrency, payableAmount, payableCurrency, stamp(expiresAt), stamp(time.Now().UTC()), orderID)
	if err != nil {
		return model.PaymentOrder{}, fmt.Errorf("update payment checkout: %w", err)
	}
	return s.PaymentOrderByID(ctx, orderID)
}

// UpdatePaymentCheckout preserves the original repository contract for
// adapters that do not return a separate receiving address.
func (s *Store) UpdatePaymentCheckout(ctx context.Context, orderID string, tradeID, paymentURL, qrPayload *string, payableAmount, payableCurrency string, expiresAt time.Time) (model.PaymentOrder, error) {
	return s.UpdatePaymentCheckoutDetails(ctx, orderID, tradeID, paymentURL, qrPayload, nil, nil, nil, payableAmount, payableCurrency, expiresAt)
}

// FailPaymentOrder records a provider creation failure without retaining provider data.
func (s *Store) FailPaymentOrder(ctx context.Context, orderID string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE payment_orders SET status='failed',provider_payload='{}',payment_url=NULL,qr_payload=NULL,updated_at=? WHERE id=? AND status='creating'`, stamp(time.Now().UTC()), orderID)
	return err
}

// ExpirePaymentOrder records an authoritative provider timeout without affecting balance.
func (s *Store) ExpirePaymentOrder(ctx context.Context, orderID, provider string, now time.Time) error {
	result, err := s.db.ExecContext(ctx, `UPDATE payment_orders SET status='expired',provider_payload='{}',payment_url=NULL,
		qr_payload=CASE WHEN receiving_address IS NOT NULL THEN NULL ELSE qr_payload END,updated_at=?
		WHERE id=? AND provider=? AND status IN ('creating','pending') AND cancelled_at IS NULL`, stamp(now), orderID, provider)
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
	_, err := s.db.ExecContext(ctx, `UPDATE payment_orders SET status='expired',provider_payload='{}',payment_url=NULL,
		qr_payload=CASE WHEN receiving_address IS NOT NULL THEN NULL ELSE qr_payload END,updated_at=?
		WHERE status IN ('creating','pending') AND cancelled_at IS NULL AND expires_at<=?`, stamp(now), stamp(now))
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

const paymentSelect = `SELECT id,user_id,provider,method_id,provider_rail,status,txb_minor,payable_amount,payable_currency,rate_snapshot,rate_direction,provider_trade_id,provider_charge_id,payment_url,qr_payload,receiving_address,actual_crypto_amount,actual_crypto_currency,expires_at,paid_at,refunded_at,cancelled_at,cancel_reason,provider_cancel_status,created_at,updated_at FROM payment_orders`

func scanPaymentOrder(row rowScanner) (model.PaymentOrder, error) {
	var order model.PaymentOrder
	var tradeID, chargeID, paymentURL, qr, receivingAddress, actualCryptoAmount, actualCryptoCurrency, paid, refunded, cancelled sql.NullString
	var methodID, providerRail, rateDirection, cancelReason, providerCancelStatus sql.NullString
	var expires, created, updated string
	if err := row.Scan(&order.ID, &order.UserID, &order.Provider, &methodID, &providerRail, &order.Status, &order.TXBMinor, &order.PayableAmount,
		&order.PayableCurrency, &order.RateSnapshot, &rateDirection, &tradeID, &chargeID, &paymentURL, &qr, &receivingAddress, &actualCryptoAmount, &actualCryptoCurrency,
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
