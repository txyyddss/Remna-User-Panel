package database

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"
)

// MarkPMDeliveryRead records a user's explicit acknowledgment for one outbound delivery.
func (s *Store) MarkPMDeliveryRead(ctx context.Context, operationID string, telegramID, chatID, messageID, updateID int64, now time.Time) (bool, error) {
	if operationID == "" || telegramID <= 0 || chatID != telegramID || messageID <= 0 || updateID <= 0 {
		return false, ErrConflict
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()
	var ownerTelegramID int64
	var resultMessage string
	err = tx.QueryRowContext(ctx, `SELECT u.telegram_id,r.provider_reference
		FROM provider_operations o
		JOIN provider_operation_items t ON t.operation_id=o.id AND t.item_key='topic' AND t.target_type='pm_conversation'
		JOIN telegram_pm_conversations c ON c.id=t.target_id
		JOIN users u ON u.id=c.user_id
		JOIN provider_operation_items r ON r.operation_id=o.id AND r.item_key='relay'
		WHERE o.id=? AND o.kind='telegram_pm_relay' AND o.status='succeeded'
		AND r.status='succeeded' AND r.target_type='pm_outbound_message'`, operationID).Scan(&ownerTelegramID, &resultMessage)
	if errors.Is(err, sql.ErrNoRows) || err == nil && (ownerTelegramID != telegramID || resultMessage != strconv.FormatInt(messageID, 10)) {
		return false, ErrNotFound
	}
	if err != nil {
		return false, err
	}
	added, err := recordPMUpdateTx(ctx, tx, updateID, now)
	if err != nil {
		return false, err
	}
	if !added {
		if err := tx.Commit(); err != nil {
			return false, err
		}
		return false, nil
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO telegram_pm_read_status(operation_id,read_at,source)
		VALUES(?,?,'explicit') ON CONFLICT(operation_id) DO UPDATE SET read_at=excluded.read_at,source='explicit'`, operationID, stamp(now))
	if err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

func markPriorPMDeliveriesReadTx(ctx context.Context, tx *sql.Tx, conversationID string, replyToID int64, occurredAt time.Time) error {
	rows, err := tx.QueryContext(ctx, `SELECT o.id,r.provider_reference FROM provider_operations o
		JOIN provider_operation_items t ON t.operation_id=o.id AND t.item_key='topic' AND t.target_type='pm_conversation'
		JOIN provider_operation_items r ON r.operation_id=o.id AND r.item_key='relay'
		WHERE o.kind='telegram_pm_relay' AND o.status='succeeded' AND t.target_id=?
		AND (? > 0 OR o.created_at<=?) AND r.status='succeeded' AND r.target_type='pm_outbound_message'
		AND (?=0 OR r.provider_reference=?)
		AND NOT EXISTS(SELECT 1 FROM telegram_pm_read_status pm_read WHERE pm_read.operation_id=o.id)
		ORDER BY o.created_at`, conversationID, replyToID, stamp(occurredAt), replyToID, strconv.FormatInt(replyToID, 10))
	if err != nil {
		return err
	}
	type deliveryRef struct{ operationID string }
	deliveries := make([]deliveryRef, 0)
	for rows.Next() {
		var operationID, providerReference string
		if err := rows.Scan(&operationID, &providerReference); err != nil {
			_ = rows.Close()
			return err
		}
		if replyToID > 0 && providerReference != strconv.FormatInt(replyToID, 10) {
			continue
		}
		deliveries = append(deliveries, deliveryRef{operationID: operationID})
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, delivery := range deliveries {
		if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO telegram_pm_read_status(operation_id,read_at,source)
			VALUES(?,?,'reply')`, delivery.operationID, stamp(occurredAt)); err != nil {
			return err
		}
	}
	return nil
}
