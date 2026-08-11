package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/txyyddss/Remna-User-Panel/internal/coupons"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
	"strings"
	"time"
)

const couponSelect = `SELECT id,code,name,kind,discount_mode,value_minor_or_bps,percent_cap_minor,
	eligible_combo_ids,eligible_squad_ids,expires_at,global_use_limit,per_user_use_limit,active,created_at,updated_at,
	(SELECT COUNT(*) FROM coupon_uses WHERE coupon_id=coupon_definitions.id) +
	(SELECT COUNT(*) FROM coupon_redemptions WHERE coupon_id=coupon_definitions.id AND balance_delta_minor<>0)
	FROM coupon_definitions`

// SaveCoupon creates or updates a canonical coupon definition.
func (s *Store) SaveCoupon(ctx context.Context, input coupons.CouponInput, now time.Time) (coupons.Coupon, error) {
	normalized, err := input.Normalize()
	if err != nil {
		return coupons.Coupon{}, err
	}
	now = now.UTC()
	comboJSON, err := json.Marshal(normalized.EligibleComboIDs)
	if err != nil {
		return coupons.Coupon{}, fmt.Errorf("encode coupon combo eligibility: %w", err)
	}
	squadJSON, err := json.Marshal(normalized.EligibleSquadIDs)
	if err != nil {
		return coupons.Coupon{}, fmt.Errorf("encode coupon squad eligibility: %w", err)
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if normalized.ID == "" {
		normalized.ID, err = ids.New()
		if err != nil {
			return coupons.Coupon{}, err
		}
		_, err = s.db.ExecContext(ctx, `INSERT INTO coupon_definitions(id,code,name,kind,discount_mode,value_minor_or_bps,percent_cap_minor,
			eligible_combo_ids,eligible_squad_ids,expires_at,global_use_limit,per_user_use_limit,active,created_at,updated_at)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, normalized.ID, normalized.Code, normalized.Name, normalized.Kind, normalized.DiscountMode,
			normalized.ValueMinorOrBPS, nullableInt64(normalized.PercentCapMinor), string(comboJSON), string(squadJSON), nullableStamp(normalized.ExpiresAt),
			nullableInt64(normalized.GlobalUseLimit), nullableInt64(normalized.PerUserUseLimit), boolInt(normalized.Active), stamp(now), stamp(now))
	} else {
		var result sql.Result
		result, err = s.db.ExecContext(ctx, `UPDATE coupon_definitions SET code=?,name=?,kind=?,discount_mode=?,value_minor_or_bps=?,percent_cap_minor=?,
			eligible_combo_ids=?,eligible_squad_ids=?,expires_at=?,global_use_limit=?,per_user_use_limit=?,active=?,updated_at=? WHERE id=?`,
			normalized.Code, normalized.Name, normalized.Kind, normalized.DiscountMode, normalized.ValueMinorOrBPS, nullableInt64(normalized.PercentCapMinor),
			string(comboJSON), string(squadJSON), nullableStamp(normalized.ExpiresAt), nullableInt64(normalized.GlobalUseLimit), nullableInt64(normalized.PerUserUseLimit),
			boolInt(normalized.Active), stamp(now), normalized.ID)
		if err == nil {
			if affected, rowsErr := result.RowsAffected(); rowsErr != nil {
				return coupons.Coupon{}, fmt.Errorf("inspect coupon update: %w", rowsErr)
			} else if affected == 0 {
				return coupons.Coupon{}, ErrNotFound
			}
		}
	}
	if err != nil {
		if isUniqueConstraint(err) {
			return coupons.Coupon{}, ErrConflict
		}
		return coupons.Coupon{}, fmt.Errorf("save coupon: %w", err)
	}
	return couponByID(ctx, s.db, normalized.ID)
}

// ListCoupons returns coupon definitions newest first.
func (s *Store) ListCoupons(ctx context.Context, activeOnly bool) ([]coupons.Coupon, error) {
	query := couponSelect
	args := make([]any, 0, 1)
	if activeOnly {
		query += ` WHERE active=1 AND (expires_at IS NULL OR expires_at>?)`
		args = append(args, stamp(time.Now().UTC()))
	}
	query += ` ORDER BY created_at DESC,id DESC`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list coupons: %w", err)
	}
	defer func() { _ = rows.Close() }()
	result := make([]coupons.Coupon, 0)
	for rows.Next() {
		coupon, scanErr := scanCoupon(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, coupon)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate coupons: %w", err)
	}
	return result, nil
}

// RedeemCoupon atomically creates a wallet grant or applies an immediate balance effect.
func (s *Store) RedeemCoupon(ctx context.Context, userID, code, idempotencyKey string, now time.Time) (coupons.RedemptionResult, error) {
	canonical, err := coupons.CanonicalCode(code)
	if err != nil {
		return coupons.RedemptionResult{}, err
	}
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(idempotencyKey) == "" || len(idempotencyKey) > 128 {
		return coupons.RedemptionResult{}, coupons.ErrInvalidInput
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	now = now.UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return coupons.RedemptionResult{}, fmt.Errorf("begin coupon redemption: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if existing, loadErr := redemptionByKeyTx(ctx, tx, userID, idempotencyKey); loadErr == nil {
		existing.Replayed = true
		return existing, nil
	} else if !errors.Is(loadErr, ErrNotFound) {
		return coupons.RedemptionResult{}, loadErr
	}
	coupon, err := couponByCodeTx(ctx, tx, canonical)
	if err != nil {
		return coupons.RedemptionResult{}, err
	}
	if err := couponAvailable(coupon, now); err != nil {
		return coupons.RedemptionResult{}, err
	}
	redemptionID, err := ids.New()
	if err != nil {
		return coupons.RedemptionResult{}, err
	}
	balance, err := balanceTx(ctx, tx, userID)
	if err != nil {
		return coupons.RedemptionResult{}, err
	}
	var grantID *string
	delta := int64(0)
	switch coupon.Kind {
	case coupons.KindPurchaseRecurring, coupons.KindPurchaseOnce:
		grant, grantErr := grantCouponTx(ctx, tx, userID, coupon, "code", coupon.Code, now)
		if grantErr != nil {
			return coupons.RedemptionResult{}, grantErr
		}
		grantID = &grant.ID
	case coupons.KindBalanceAdd, coupons.KindBalanceMultiply:
		if err := ensureCouponUseAvailableTx(ctx, tx, coupon, userID); err != nil {
			return coupons.RedemptionResult{}, err
		}
		if coupon.Kind == coupons.KindBalanceAdd {
			delta = coupon.ValueMinorOrBPS
		} else {
			delta, err = coupons.CalculateBalanceMultiplyCredit(balance, coupon.ValueMinorOrBPS)
			if err != nil {
				return coupons.RedemptionResult{}, err
			}
		}
		balance, err = changeBalanceTx(ctx, tx, userID, delta, now)
		if err != nil {
			return coupons.RedemptionResult{}, err
		}
		ledgerKind := "coupon_balance_add"
		if coupon.Kind == coupons.KindBalanceMultiply {
			ledgerKind = "coupon_balance_multiply"
		}
		if _, err := insertLedgerTx(ctx, tx, userID, delta, balance, ledgerKind, redemptionID, coupon.Code, now); err != nil {
			return coupons.RedemptionResult{}, err
		}
	default:
		return coupons.RedemptionResult{}, coupons.ErrInvalidInput
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO coupon_redemptions(id,coupon_id,grant_id,user_id,balance_delta_minor,balance_after_minor,idempotency_key,created_at)
		VALUES(?,?,?,?,?,?,?,?)`, redemptionID, coupon.ID, nullableStringValue(grantID), userID, delta, balance, idempotencyKey, stamp(now))
	if err != nil {
		return coupons.RedemptionResult{}, fmt.Errorf("record coupon redemption: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return coupons.RedemptionResult{}, fmt.Errorf("commit coupon redemption: %w", err)
	}
	return s.redemptionByID(ctx, redemptionID)
}

// GrantCoupon idempotently puts a purchase-discount coupon in a member wallet.
func (s *Store) GrantCoupon(ctx context.Context, userID, couponID, sourceType, sourceID string, now time.Time) (coupons.Grant, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(couponID) == "" || strings.TrimSpace(sourceType) == "" || strings.TrimSpace(sourceID) == "" {
		return coupons.Grant{}, coupons.ErrInvalidInput
	}
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return coupons.Grant{}, err
	}
	defer func() { _ = tx.Rollback() }()
	coupon, err := couponByID(ctx, tx, couponID)
	if err != nil {
		return coupons.Grant{}, err
	}
	if err := couponAvailable(coupon, now); err != nil {
		return coupons.Grant{}, err
	}
	grant, err := grantCouponTx(ctx, tx, userID, coupon, sourceType, sourceID, now.UTC())
	if err != nil {
		return coupons.Grant{}, err
	}
	if err := tx.Commit(); err != nil {
		return coupons.Grant{}, err
	}
	return grant, nil
}

// ListCouponGrants returns currently usable purchase-discount grants.
func (s *Store) ListCouponGrants(ctx context.Context, userID string, now time.Time) ([]coupons.Grant, error) {
	if _, err := s.db.ExecContext(ctx, `UPDATE coupon_grants SET status='expired' WHERE user_id=? AND status='active'
		AND coupon_id IN (SELECT id FROM coupon_definitions WHERE expires_at IS NOT NULL AND expires_at<=?)
		AND NOT EXISTS (SELECT 1 FROM coupon_grant_discards WHERE coupon_grant_discards.grant_id=coupon_grants.id)`, userID, stamp(now)); err != nil {
		return nil, fmt.Errorf("expire coupon grants: %w", err)
	}
	rows, err := s.db.QueryContext(ctx, grantSelect+` WHERE coupon_grants.user_id=? AND coupon_grants.status='active'
		AND coupon_definitions.active=1 AND (coupon_definitions.expires_at IS NULL OR coupon_definitions.expires_at>?)
		AND NOT EXISTS (SELECT 1 FROM coupon_grant_discards WHERE coupon_grant_discards.grant_id=coupon_grants.id)
		ORDER BY coupon_grants.created_at DESC`, userID, stamp(now))
	if err != nil {
		return nil, fmt.Errorf("list coupon grants: %w", err)
	}
	defer func() { _ = rows.Close() }()
	result := make([]coupons.Grant, 0)
	for rows.Next() {
		grant, scanErr := scanGrant(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, grant)
	}
	return result, rows.Err()
}

// DiscardCouponGrant hides one active grant while preserving its immutable
// redemption and purchase links. Replaying a member's own discard is a no-op.
func (s *Store) DiscardCouponGrant(ctx context.Context, userID, grantID string, now time.Time) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin coupon discard: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	var existingGrantID string
	err = tx.QueryRowContext(ctx, `SELECT grant_id FROM coupon_grant_discards WHERE grant_id=? AND user_id=?`, grantID, userID).Scan(&existingGrantID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("load coupon discard: %w", err)
	}
	var status string
	err = tx.QueryRowContext(ctx, `SELECT status FROM coupon_grants WHERE id=? AND user_id=?`, grantID, userID).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("load coupon grant for discard: %w", err)
	}
	if status != "active" {
		return ErrConflict
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO coupon_grant_discards(grant_id,user_id,discarded_at) VALUES(?,?,?)`, grantID, userID, stamp(now.UTC())); err != nil {
		if isUniqueConstraint(err) {
			return nil
		}
		return fmt.Errorf("record coupon discard: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit coupon discard: %w", err)
	}
	return nil
}
